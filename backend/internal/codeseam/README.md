# CodeSeam P2 adapter

`client.go` uses existing APIKeyService and SubscriptionService with existing JWT, audit and panel-rate-limit middleware. No user, order, subscription, accounting or binding tables are added. `internal/server/router.go` is the only upstream Go integration change.

- `GET /api/v1/codeseam/service/current`: `data.services[]` contains owned Key ID/name, authorization state, effective expiry and native Key quota/usage; never Key material. ACTIVE describes authorization plus a valid native dedicated resource binding.
- `POST /api/v1/codeseam/client/activate`, body `{"service_id": <owned key ID>}`: checks JWT ownership, native Key/user/group status, expiry/quota/balance or subscription validity/limits, and returns the resource revision with the managed Key to the desktop native service. No client-selected user/account/proxy identity is trusted.
- `POST /api/v1/codeseam/client/network-config`, body `{"service_id": <owned key ID>}`: returns the assigned native proxy configuration and expected egress to the Rust service only. React receives only a redacted network summary; all three endpoints are `no-store`.

Dedicated bindings are encoded as `codeseam:v1:<json>` in the native exclusive Anthropic subscription Group description. The binding points to one user, API Key, independent Account and Proxy, plus an operator-verified expected egress. Gateway middleware rehydrates native resources on every request, rejects fallback groups, model routing, shared accounts/proxies, invalid subscriptions and unsupported endpoints, and the forwarding/scheduling layer rechecks the selected Account immediately before upstream use.

Actual gateway calls still use original Sub2API `/v1/messages` authentication, scheduling and metering. DeepSeek remains the temporary development upstream through a native Anthropic API-key Account with its base URL and model mapping; no custom DeepSeek forwarding path was added. The P2 evidence uses authenticated loopback CONNECT fixtures and explicitly does not certify a dedicated public IP, non-API process traffic, DNS/IPv6/WebRTC leak protection or a kill-switch; those remain P3.
