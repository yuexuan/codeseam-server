package codeseam

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Delivery serves the independently built Portal before the operations SPA.
// Empty deployment paths leave the existing server behavior unchanged.
func Delivery() gin.HandlerFunc {
	portal := os.Getenv("CODESEAM_PORTAL_DIR")
	installer := os.Getenv("CODESEAM_WINDOWS_INSTALLER")
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == "/downloads/codeseam-windows.exe" {
			if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				c.AbortWithStatus(405)
				return
			}
			if info, err := os.Stat(installer); err != nil || !info.Mode().IsRegular() {
				c.AbortWithStatus(404)
				return
			}
			c.Header("Content-Disposition", "attachment; filename=CodeSeam-windows-x64-setup.exe")
			c.Header("Cache-Control", "no-cache")
			c.File(installer)
			c.Abort()
			return
		}
		if portal == "" || (p != "/portal" && !strings.HasPrefix(p, "/portal/")) {
			c.Next()
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatus(405)
			return
		}
		if p == "/portal" {
			c.Redirect(http.StatusTemporaryRedirect, "/portal/")
			c.Abort()
			return
		}
		rel := strings.TrimPrefix(p, "/portal/")
		// The portal uses a single page, with no history-router fallback required.
		if rel == "" {
			rel = "index.html"
		}
		if !filepath.IsLocal(filepath.FromSlash(rel)) || strings.Contains(rel, "\\") {
			c.AbortWithStatus(404)
			return
		}
		target := filepath.Join(portal, filepath.FromSlash(rel))
		if info, err := os.Stat(target); err != nil || !info.Mode().IsRegular() {
			c.AbortWithStatus(404)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.File(target)
		c.Abort()
	}
}
