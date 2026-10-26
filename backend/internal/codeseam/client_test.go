package codeseam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type testKeys struct{ key *service.APIKey }

func (s testKeys) List(_ context.Context, uid int64, _ pagination.PaginationParams, _ service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	if s.key.UserID != uid {
		return nil, &pagination.PaginationResult{Pages: 1}, nil
	}
	return []service.APIKey{*s.key}, &pagination.PaginationResult{Pages: 1}, nil
}
func (s testKeys) GetByID(context.Context, int64) (*service.APIKey, error) { return s.key, nil }
func testKey() *service.APIKey {
	return &service.APIKey{ID: 10, UserID: 1, Key: "secret-must-not-appear-in-view", Name: "P1", Status: service.StatusActive, User: &service.User{ID: 1, Status: service.StatusActive, Balance: 10}, Group: &service.Group{ID: 1, Platform: service.PlatformAnthropic, Status: service.StatusActive}}
}
func request(k *service.APIKey, uid int64, method, path, body string) *httptest.ResponseRecorder {
	r := gin.New()
	Register(r.Group(""), testKeys{k}, nil, func(*gin.Context) int64 { return uid })
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestViewNeverIncludesSecretAndActivationChecksOwner(t *testing.T) {
	k := testKey()
	w := request(k, 1, http.MethodGet, "/service/current", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), k.Key) {
		t.Fatal("service view leaked or failed")
	}
	w = request(k, 2, http.MethodPost, "/client/activate", `{"service_id":10,"user_id":1}`)
	if w.Code != 404 || strings.Contains(w.Body.String(), k.Key) {
		t.Fatal("foreign activation allowed")
	}
	w = request(k, 1, http.MethodPost, "/client/activate", `{"service_id":10}`)
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if w.Code != 200 || !strings.Contains(w.Body.String(), k.Key) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("owner activation failed")
	}
}
func TestInvalidAuthorizationCannotActivate(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	for _, mutate := range []func(*service.APIKey){
		func(k *service.APIKey) { k.Status = "inactive" }, func(k *service.APIKey) { k.ExpiresAt = &past },
		func(k *service.APIKey) { k.Quota = 1; k.QuotaUsed = 1 }, func(k *service.APIKey) { k.Group.Status = "inactive" },
		func(k *service.APIKey) { k.User.Status = "disabled" }, func(k *service.APIKey) { k.User.Balance = 0 },
		func(k *service.APIKey) { k.Group = nil },
	} {
		k := testKey()
		mutate(k)
		w := request(k, 1, http.MethodPost, "/client/activate", `{"service_id":10}`)
		if w.Code != 403 || strings.Contains(w.Body.String(), k.Key) {
			t.Fatal("invalid authorization activated")
		}
	}
}
