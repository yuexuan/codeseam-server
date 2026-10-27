# CodeSeam Windows MVP delivery

The existing Go service serves the independently built Portal at `/portal/` and the configured installer at `/downloads/codeseam-windows.exe`. Auth, authorization, subscriptions and usage remain native Sub2API data. Operations routes retain their existing behavior.

Build the client with its `scripts/build-windows.ps1`. In `codeseam/portal`, install the frozen pnpm dependencies, then run `pwsh -File codeseam/deploy/build-delivery.ps1` from the server root. The script builds the Portal with a same-origin download link and copies the installer into ignored `codeseam/deploy/dist/`. It does not embed credentials or create any business resource.

On the existing server, set:

```text
CODESEAM_PORTAL_DIR=<absolute path to codeseam/portal/dist>
CODESEAM_WINDOWS_INSTALLER=<absolute path to CodeSeam-windows-x64-setup.exe>
```

Both paths must contain deployment-controlled public files only. Missing installer files return 404; directory listings and escaping the Portal root are rejected. Preserve these paths when moving the service. The existing HTTPS ingress should forward `/portal/`, `/downloads/` and `/api/v1/` to this same Go process. A same-origin Portal needs no new CORS policy or separate Vite process. `start-p1.ps1` selects these paths for local loopback QA only; it is not evidence of a public HTTPS deployment.

For local repeatable verification: the existing P2 QA fixture must already exist. `p5-loopback-relay.mjs` starts an authenticated TCP/HTTP relay, observes real shared-host egress, and temporarily updates only that QA binding. It synthesizes neither IP responses nor model output. `--restore` restores the saved binding. The relay is not a dedicated public proxy or proof of P3 protection. `verify-delivery.mjs` checks production assets, installer hash, actual current-user data and revoke/restore behavior; an installed active client is required. Secrets stay in ignored `.local/`, never in the report.

The single P5 task and redacted evidence are in `codeseam/delivery/`; P5 remains non-PASS until the documented release-signing requirement and remaining P5 acceptance gaps are closed. Missing Anthropic/dedicated-egress/P3 evidence is tracked separately and is not used as an additional P5 gate.
