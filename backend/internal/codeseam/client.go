// Package codeseam is a thin desktop adapter over Sub2API's existing resources.
package codeseam

import (
	"context"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type KeyStore interface {
	List(context.Context, int64, pagination.PaginationParams, service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error)
	GetByID(context.Context, int64) (*service.APIKey, error)
}
type Subscriptions interface {
	GetActiveSubscription(context.Context, int64, int64) (*service.UserSubscription, error)
	ValidateAndCheckLimits(*service.UserSubscription, *service.Group) (bool, error)
}

type ServiceView struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	State     string     `json:"state"`
	ExpiresAt *time.Time `json:"expires_at"`
	QuotaUsed float64    `json:"quota_used"`
	Quota     float64    `json:"quota"`
}

// State describes authorization only. Gateway scheduling/health remains authoritative.
func view(ctx context.Context, k *service.APIKey, subs Subscriptions) ServiceView {
	v := ServiceView{ID: k.ID, Name: k.Name, State: "ACTIVE", ExpiresAt: k.ExpiresAt, Quota: k.Quota, QuotaUsed: k.QuotaUsed}
	switch {
	case !k.IsActive():
		v.State = "DISABLED"
	case k.IsExpired():
		v.State = "EXPIRED"
	case k.IsQuotaExhausted():
		v.State = "QUOTA_EXHAUSTED"
	case k.Group == nil || k.Group.Platform != service.PlatformAnthropic:
		v.State = "NOT_SUPPORTED"
	case k.Group.Status != service.StatusActive:
		v.State = "DISABLED"
	case k.User == nil || !k.User.IsActive():
		v.State = "DISABLED"
	default:
		if k.Group.IsSubscriptionType() {
			sub, err := subs.GetActiveSubscription(ctx, k.UserID, k.Group.ID)
			if err != nil || sub == nil {
				v.State = "NOT_PROVISIONED"
				break
			}
			if v.ExpiresAt == nil || sub.ExpiresAt.Before(*v.ExpiresAt) {
				v.ExpiresAt = &sub.ExpiresAt
			}
			if !sub.IsActive() || time.Now().Before(sub.StartsAt) {
				v.State = "EXPIRED"
				break
			}
			if _, err := subs.ValidateAndCheckLimits(sub, k.Group); err != nil {
				v.State = "QUOTA_EXHAUSTED"
			}
		} else if k.User.Balance <= 0 {
			v.State = "QUOTA_EXHAUSTED"
		}
	}
	return v
}

// Register runs inside the existing authenticated, audited and rate-limited group.
// No user/account/proxy IDs supplied by the desktop are trusted.
func Register(r *gin.RouterGroup, keys KeyStore, subs Subscriptions, userID func(*gin.Context) int64) {
	r.GET("/service/current", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		items := make([]ServiceView, 0)
		for page := 1; ; page++ {
			rows, p, err := keys.List(c.Request.Context(), userID(c), pagination.PaginationParams{Page: page, PageSize: 100}, service.APIKeyListFilters{})
			if err != nil {
				response.ErrorFrom(c, err)
				return
			}
			for i := range rows {
				// GetByID hydrates the same user/group relations used at activation.
				k, err := keys.GetByID(c.Request.Context(), rows[i].ID)
				if err != nil {
					response.ErrorFrom(c, err)
					return
				}
				if k.UserID == userID(c) {
					items = append(items, view(c.Request.Context(), k, subs))
				}
			}
			if p == nil || page >= p.Pages {
				break
			}
		}
		response.Success(c, gin.H{"services": items, "scope": "P1 authorization; upstream health checked by Gateway"})
	})
	r.POST("/client/activate", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		var req struct {
			ServiceID int64 `json:"service_id" binding:"required,gt=0"`
		}
		if c.ShouldBindJSON(&req) != nil {
			response.BadRequest(c, "请选择本人服务")
			return
		}
		k, err := keys.GetByID(c.Request.Context(), req.ServiceID)
		if err != nil || k == nil || k.UserID != userID(c) {
			response.Error(c, http.StatusNotFound, "服务不存在")
			return
		}
		v := view(c.Request.Context(), k, subs)
		if v.State != "ACTIVE" {
			response.ErrorWithDetails(c, http.StatusForbidden, "服务不可激活："+v.State, v.State, nil)
			return
		}
		response.Success(c, gin.H{"service": v, "api_key": k.Key})
	})
}
