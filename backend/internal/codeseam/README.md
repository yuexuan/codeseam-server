# CodeSeam P1 adapter

`client.go` uses existing APIKeyService and SubscriptionService with existing JWT, audit and panel-rate-limit middleware. No user, order, subscription, accounting or binding tables are added. `internal/server/router.go` is the only upstream Go integration change.

- `GET /api/v1/codeseam/service/current`: `data.services[]` contains owned Key ID/name, authorization state, effective expiry and native Key quota/usage; never Key material. ACTIVE describes authorization, not Account/Proxy health or exclusive routing.
- `POST /api/v1/codeseam/client/activate`, body `{"service_id": <owned key ID>}`: checks JWT ownership, native Key/user/group status, expiry/quota/balance or subscription validity/limits; returns `data.service` and `data.api_key` to the desktop native service. No client-selected user/account/proxy identity is trusted. Both endpoints are no-store.

Actual gateway calls still use original Sub2API `/v1/messages` authentication, scheduling and metering. DeepSeek P1 uses a native Anthropic API-key Account with its base URL and model mapping. No custom DeepSeek forwarding path was added. Network provisioning and strict exclusive routing remain P2.
