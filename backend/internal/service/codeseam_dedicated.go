package service

// CodeSeam's only extra relation lives in the native Group description. Native
// subscriptions, keys, accounts, proxies and usage remain the source of truth.
import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const CodeSeamBindingPrefix = "codeseam:v1:"

type CodeSeamBinding struct {
	UserID         int64  `json:"user_id"`
	ServiceID      int64  `json:"service_id"`
	AccountID      int64  `json:"account_id"`
	ProxyID        int64  `json:"proxy_id"`
	ExpectedEgress string `json:"expected_egress"`
	Revision       string `json:"revision"`
}

type CodeSeamResource struct {
	Binding CodeSeamBinding
	Account *Account
	Proxy   *Proxy
}

func CodeSeamParseBinding(g *Group) (*CodeSeamBinding, error) {
	if g == nil || !strings.HasPrefix(g.Description, "codeseam:") {
		return nil, nil
	}
	b := new(CodeSeamBinding)
	if !strings.HasPrefix(g.Description, CodeSeamBindingPrefix) || json.Unmarshal([]byte(strings.TrimPrefix(g.Description, CodeSeamBindingPrefix)), b) != nil || b.UserID <= 0 || b.ServiceID <= 0 || b.AccountID <= 0 || b.ProxyID <= 0 || b.Revision == "" || net.ParseIP(b.ExpectedEgress) == nil {
		return nil, fmt.Errorf("BINDING_INVALID")
	}
	return b, nil
}

// Read the live repositories, never a scheduler/auth cache. A conflicting native
// edit makes the service unavailable instead of widening its candidate pool.
func (s *GatewayService) CodeSeamResource(ctx context.Context, k *APIKey) (*CodeSeamResource, error) {
	if k == nil || k.GroupID == nil {
		return nil, nil
	}
	g, err := s.groupRepo.GetByID(ctx, *k.GroupID)
	if err != nil {
		return nil, fmt.Errorf("RESOURCE_UNAVAILABLE")
	}
	b, err := CodeSeamParseBinding(g)
	if err != nil || b == nil {
		return nil, err
	}
	if b.UserID != k.UserID || b.ServiceID != k.ID {
		return nil, fmt.Errorf("SERVICE_NOT_ASSIGNED")
	}
	if !g.IsActive() || !g.IsExclusive || !g.IsSubscriptionType() || g.Platform != PlatformAnthropic || g.FallbackGroupID != nil || g.FallbackGroupIDOnInvalidRequest != nil || g.ModelRoutingEnabled {
		return nil, fmt.Errorf("ROUTE_CONFLICT")
	}
	ids, err := s.groupRepo.GetAccountIDsByGroupIDs(ctx, []int64{g.ID})
	if err != nil || len(ids) != 1 || ids[0] != b.AccountID {
		return nil, fmt.Errorf("RESOURCE_CONFLICT")
	}
	a, err := s.accountRepo.GetByID(ctx, b.AccountID)
	if err != nil || a == nil || a.Platform != PlatformAnthropic || a.ParentAccountID != nil || len(a.GroupIDs) != 1 || a.GroupIDs[0] != g.ID {
		return nil, fmt.Errorf("RESOURCE_CONFLICT")
	}
	if a.GetExtraString("codeseam_service_id") != strconv.FormatInt(k.ID, 10) {
		return nil, fmt.Errorf("ACCOUNT_NOT_ASSIGNED")
	}
	if !a.IsSchedulable() || (a.ExpiresAt != nil && !a.ExpiresAt.After(time.Now())) {
		return nil, fmt.Errorf("ACCOUNT_UNAVAILABLE")
	}
	p := a.Proxy
	if a.ProxyID == nil || *a.ProxyID != b.ProxyID || a.ProxyFallbackOriginID != nil || p == nil || p.ID != b.ProxyID || !p.IsActive() || p.IsExpired(time.Now()) || (p.FallbackMode != "" && p.FallbackMode != FallbackModeNone) || p.BackupProxyID != nil || p.Host == "" || p.Port < 1 || p.Port > 65535 {
		return nil, fmt.Errorf("PROXY_UNAVAILABLE")
	}
	switch p.Protocol {
	case "http", "https", "socks5", "socks5h":
	default:
		return nil, fmt.Errorf("PROXY_UNSUPPORTED")
	}
	// A proxy assigned to another account is not an exclusive resource.
	checker, ok := s.accountRepo.(interface {
		CodeSeamProxyExclusive(context.Context, int64, int64) (bool, error)
	})
	if !ok {
		return nil, fmt.Errorf("RESOURCE_UNAVAILABLE")
	}
	exclusive, err := checker.CodeSeamProxyExclusive(ctx, p.ID, a.ID)
	if err != nil {
		return nil, fmt.Errorf("RESOURCE_UNAVAILABLE")
	}
	if !exclusive {
		return nil, fmt.Errorf("PROXY_SHARED")
	}
	return &CodeSeamResource{Binding: *b, Account: a, Proxy: p}, nil
}

type codeSeamRouteKey struct{}

func WithCodeSeamRoute(ctx context.Context, k *APIKey) context.Context {
	return context.WithValue(ctx, codeSeamRouteKey{}, k)
}

// Called immediately before forwarding, including count_tokens. Hydrate again
// so a stale sticky/scheduler selection cannot use a previous account or proxy.
func (s *GatewayService) codeSeamForwardAccount(ctx context.Context, selected *Account) (*Account, error) {
	k, ok := ctx.Value(codeSeamRouteKey{}).(*APIKey)
	if !ok {
		if selected != nil && selected.GetExtraString("codeseam_service_id") != "" {
			return nil, fmt.Errorf("DEDICATED_AUTH_REQUIRED")
		}
		return selected, nil
	}
	if !k.IsActive() || k.IsExpired() || k.IsQuotaExhausted() || k.User == nil || !k.User.IsActive() {
		return nil, fmt.Errorf("DEDICATED_AUTH_EXPIRED")
	}
	if s.userSubRepo != nil && k.GroupID != nil {
		sub, err := s.userSubRepo.GetByUserIDAndGroupID(ctx, k.UserID, *k.GroupID)
		if err != nil || sub == nil || !sub.IsActive() || time.Now().Before(sub.StartsAt) {
			return nil, fmt.Errorf("DEDICATED_SUBSCRIPTION_UNAVAILABLE")
		}
	}
	resource, err := s.CodeSeamResource(ctx, k)
	if err != nil || resource == nil {
		return nil, fmt.Errorf("DEDICATED_ROUTE_UNAVAILABLE")
	}
	if selected == nil || selected.ID != resource.Binding.AccountID {
		return nil, fmt.Errorf("DEDICATED_ACCOUNT_MISMATCH")
	}
	return resource.Account, nil
}
