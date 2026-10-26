package codeseam

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Resources func(context.Context, *service.APIKey) (*service.CodeSeamResource, error)

// Runs after native auth and before every compatible endpoint. Auth caches do
// not extend a dedicated service's lifetime: reload the native key and group.
func GatewayGuard(keys KeyStore, subs Subscriptions, resources Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		// The guard is installed on the shared gateway. A nil resolver is used by
		// lightweight route tests and by partially wired startup; ordinary native
		// routing must continue in that case.
		if resources == nil {
			return
		}
		cached, ok := middleware.GetAPIKeyFromContext(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// The auth cache intentionally omits the binding description. Rehydrate
		// through the native service so revocation and relation drift are observed
		// on every request. A nil service is only used by lightweight route tests.
		if keys == nil {
			return
		}
		if nativeKeys, ok := keys.(*service.APIKeyService); ok && nativeKeys == nil {
			return
		}
		k, err := keys.GetByID(c.Request.Context(), cached.ID)
		if err != nil || k == nil {
			denyRoute(c, "SERVICE_UNAVAILABLE")
			return
		}
		r, err := resources(c.Request.Context(), k)
		if err != nil {
			denyRoute(c, err.Error())
			return
		}
		if r == nil {
			return
		}
		v := view(c.Request.Context(), k, subs)
		if v.State != "ACTIVE" {
			denyRoute(c, v.State)
			return
		}
		// P2 only supports native Anthropic endpoints, never alternative dispatch.
		path := strings.TrimSuffix(c.Request.URL.Path, "/")
		for _, prefix := range []string{"/api/v3", "/v3", "/v1"} {
			if strings.HasPrefix(path, prefix) {
				path = strings.TrimPrefix(path, prefix)
				break
			}
		}
		switch {
		case path == "/messages", path == "/messages/count_tokens", path == "/models", path == "/usage", path == "/sub2api/billing":
		case strings.HasPrefix(path, "/models/"):
		default:
			denyRoute(c, "DEDICATED_ENDPOINT_UNSUPPORTED")
			return
		}
		c.Set(string(middleware.ContextKeyAPIKey), k)
		c.Request = c.Request.WithContext(service.WithCodeSeamRoute(c.Request.Context(), k))
	}
}

func denyRoute(c *gin.Context, reason string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"type": "error", "error": gin.H{"type": "permission_error", "message": "CodeSeam: " + reason}})
}
