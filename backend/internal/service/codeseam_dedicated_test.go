package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type csAccounts struct {
	AccountRepository
	a   *Account
	all []Account
}

func (r csAccounts) CodeSeamProxyExclusive(_ context.Context, proxyID, accountID int64) (bool, error) {
	for _, a := range r.all {
		if a.ID != accountID && a.ProxyID != nil && *a.ProxyID == proxyID {
			return false, nil
		}
	}
	return true, nil
}

func (r csAccounts) GetByID(context.Context, int64) (*Account, error) { return r.a, nil }
func (r csAccounts) ListAllWithFilters(context.Context, string, string, string, string, int64, string) ([]Account, error) {
	return r.all, nil
}

type csGroups struct {
	GroupRepository
	g   *Group
	ids []int64
}

func (r csGroups) GetByID(context.Context, int64) (*Group, error) { return r.g, nil }
func (r csGroups) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	return r.ids, nil
}

func csFixture() (*GatewayService, *APIKey, *Group, *Account) {
	b, _ := json.Marshal(CodeSeamBinding{UserID: 1, ServiceID: 2, AccountID: 3, ProxyID: 4, ExpectedEgress: "192.0.2.1", Revision: "1"})
	g := &Group{ID: 5, Description: CodeSeamBindingPrefix + string(b), Status: StatusActive, Platform: PlatformAnthropic, IsExclusive: true, SubscriptionType: SubscriptionTypeSubscription}
	pid := int64(4)
	a := &Account{ID: 3, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true, GroupIDs: []int64{5}, ProxyID: &pid, Extra: map[string]any{"codeseam_service_id": "2"}, Proxy: &Proxy{ID: 4, Status: StatusActive, Protocol: "http", Host: "localhost", Port: 8080, FallbackMode: FallbackModeNone}}
	s := &GatewayService{groupRepo: csGroups{g: g, ids: []int64{3}}, accountRepo: csAccounts{a: a}}
	return s, &APIKey{ID: 2, UserID: 1, GroupID: &g.ID, Status: StatusActive, User: &User{ID: 1, Status: StatusActive}}, g, a
}

func TestCodeSeamDedicatedRejectsDrift(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*GatewayService, *APIKey, *Group, *Account)
	}{
		{"foreign owner", func(s *GatewayService, k *APIKey, g *Group, a *Account) { k.UserID = 9 }},
		{"foreign key", func(s *GatewayService, k *APIKey, g *Group, a *Account) { k.ID = 9 }},
		{"fallback group", func(s *GatewayService, k *APIKey, g *Group, a *Account) { g.FallbackGroupID = &g.ID }},
		{"invalid request fallback", func(s *GatewayService, k *APIKey, g *Group, a *Account) { g.FallbackGroupIDOnInvalidRequest = &g.ID }},
		{"model priority", func(s *GatewayService, k *APIKey, g *Group, a *Account) { g.ModelRoutingEnabled = true }},
		{"extra candidate", func(s *GatewayService, k *APIKey, g *Group, a *Account) {
			s.groupRepo = csGroups{g: g, ids: []int64{3, 9}}
		}},
		{"shared account", func(s *GatewayService, k *APIKey, g *Group, a *Account) { a.GroupIDs = append(a.GroupIDs, 9) }},
		{"disabled account", func(s *GatewayService, k *APIKey, g *Group, a *Account) { a.Status = "disabled" }},
		{"expired account", func(s *GatewayService, k *APIKey, g *Group, a *Account) {
			v := time.Now().Add(-time.Hour)
			a.ExpiresAt = &v
		}},
		{"missing proxy", func(s *GatewayService, k *APIKey, g *Group, a *Account) { a.Proxy = nil }},
		{"direct fallback", func(s *GatewayService, k *APIKey, g *Group, a *Account) { a.Proxy.FallbackMode = FallbackModeDirect }},
		{"expired proxy", func(s *GatewayService, k *APIKey, g *Group, a *Account) {
			v := time.Now().Add(-time.Hour)
			a.Proxy.ExpiresAt = &v
		}},
		{"shared proxy", func(s *GatewayService, k *APIKey, g *Group, a *Account) {
			s.accountRepo = csAccounts{a: a, all: []Account{{ID: 9, ProxyID: a.ProxyID}}}
		}},
		{"malformed binding", func(s *GatewayService, k *APIKey, g *Group, a *Account) { g.Description = "codeseam:v1:{}" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, k, g, a := csFixture()
			tc.mutate(s, k, g, a)
			if _, err := s.CodeSeamResource(context.Background(), k); err == nil {
				t.Fatal("unsafe route accepted")
			}
		})
	}
}

func TestCodeSeamDedicatedRehydratesAndRejectsStaleSelection(t *testing.T) {
	s, k, _, a := csFixture()
	ctx := WithCodeSeamRoute(context.Background(), k)
	got, err := s.codeSeamForwardAccount(ctx, &Account{ID: 3})
	if err != nil || got != a {
		t.Fatal("live account not used", err)
	}
	if _, err = s.codeSeamForwardAccount(ctx, &Account{ID: 9}); err == nil {
		t.Fatal("stale/foreign account accepted")
	}
	if _, err = s.codeSeamForwardAccount(context.Background(), a); err == nil {
		t.Fatal("dedicated account used without dedicated auth")
	}
}
