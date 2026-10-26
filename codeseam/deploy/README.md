# P1 Windows local main-service deployment

Run `pwsh -File codeseam/deploy/start-p1.ps1 -Build` from the server root. It starts only this project's PostgreSQL 16 and Redis 7 dependencies, builds the current Fork with the original embedded operations UI, and starts the Windows main server at `http://127.0.0.1:18081`. Dependencies bind to loopback ports 15432/16379. It uses standard billing mode and the original schema/migrations; it is not a mock or an additional backend.

The ignored `.local/` directory holds database data, local test credentials, configuration, PID and logs. The script restricts its ACL to the current user/SYSTEM and supplies the Go toolchain's zoneinfo archive for Windows. Repeated starts refuse an occupied server port. To restart, inspect `.local/server.pid` and its executable path before stopping that CodeSeam process. Logs rotate on restart.

For a fresh local fixture, `provision-p1.ps1 -UpstreamFile <private JSON>` uses only native registration/admin/key APIs. JSON fields: `base_url`, `api_key`, `model`. Pass `-AcceptLocalCompliance` only with the operator's authorization of the upstream admin acknowledgement. Existing test resources stop provisioning instead of duplicating them. Real upstream credentials must never be committed.

2026-10-26 P1 used DeepSeek's official Anthropic-compatible URL and `deepseek-flash`, via a native Anthropic API-key Account and `model_mapping`. Test credit/Key quota are fixture values, not an upstream balance report. Real Anthropic and dedicated proxy/egress acceptance remain non-PASS. This is local development, not a public deployment.
