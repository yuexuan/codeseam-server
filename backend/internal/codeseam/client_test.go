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

func TestNetworkConfigurationIsOwnerOnlyAndNoStore(t *testing.T) {
	k := testKey()
	resource := &service.CodeSeamResource{Binding: service.CodeSeamBinding{Revision: "1", ExpectedEgress: "192.0.2.1"}, Account: &service.Account{}, Proxy: &service.Proxy{Protocol: "http", Host: "localhost", Port: 8888, Username: "proxy-user", Password: "proxy-secret"}}
	for _, uid := range []int64{1, 2} {
		r := gin.New()
		Register(r.Group(""), testKeys{k}, nil, func(*gin.Context) int64 { return uid }, func(context.Context, *service.APIKey) (*service.CodeSeamResource, error) { return resource, nil })
		req := httptest.NewRequest("POST", "/client/network-config", strings.NewReader(`{"service_id":10,"user_id":1}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("configuration may be cached")
		}
		if uid == 1 && (w.Code != 200 || !strings.Contains(w.Body.String(), "proxy-secret") || strings.Contains(w.Body.String(), k.Key)) {
			t.Fatal("wrong owner configuration")
		}
		if uid == 2 && (w.Code != 404 || strings.Contains(w.Body.String(), "proxy-secret")) {
			t.Fatal("foreign configuration disclosed")
		}
	}
}

type csSubscriptions struct{ sub *service.UserSubscription }

func (s csSubscriptions) GetActiveSubscription(context.Context, int64, int64) (*service.UserSubscription, error) {
	return &service.UserSubscription{Status: service.SubscriptionStatusActive, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s csSubscriptions) GetActiveSubscriptionFresh(context.Context, int64, int64) (*service.UserSubscription, error) {
	return s.sub, nil
}
func (s csSubscriptions) ValidateAndCheckLimits(*service.UserSubscription, *service.Group) (bool, error) {
	return false, nil
}

func TestDedicatedSubscriptionStateUsesFreshRecord(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     string
		start, end time.Time
		want       string
	}{
		{"active", service.SubscriptionStatusActive, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), "ACTIVE"},
		{"future", service.SubscriptionStatusActive, time.Now().Add(time.Hour), time.Now().Add(2 * time.Hour), "NOT_STARTED"},
		{"expired", service.SubscriptionStatusActive, time.Now().Add(-2 * time.Hour), time.Now().Add(-time.Hour), "EXPIRED"},
		{"suspended", service.SubscriptionStatusSuspended, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), "SUSPENDED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			k := testKey()
			k.Group.SubscriptionType = service.SubscriptionTypeSubscription
			s := csSubscriptions{sub: &service.UserSubscription{Status: tc.status, StartsAt: tc.start, ExpiresAt: tc.end}}
			v := view(context.Background(), k, s)
			if v.State != tc.want {
				t.Fatalf("got %s want %s", v.State, tc.want)
			}
		})
	}
}
