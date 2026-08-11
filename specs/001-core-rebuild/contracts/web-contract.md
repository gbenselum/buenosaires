# Web Interface Contract

**Feature**: [spec.md](../spec.md) | **File**: contracts/web-contract.md

## Routes

| Route | Method | Purpose | Success | Errors |
|-------|--------|---------|---------|--------|
| `/` | GET | List `.log` files in the log directory | 200 HTML gallery | 500 (read failure); 200 empty list when log dir missing |
| `/logs/{name}` | GET | Render one log + asset JSON accordion | 200 HTML | 400 (unsafe name), 404 (missing), 500 |
| `/static/patternfly.min.css` | GET | Embedded stylesheet (go:embed) | 200 CSS | 500 |
| `/plugins/shell/assets/{file}` | GET | Asset JSON by basename | 200 JSON | 400 (unsafe name), 404 |

## Security Requirements (constitution §II — non-negotiable)

- Every response carries: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`.
- `{name}`/`{file}` MUST be a single path segment: reject empty, `..`, `/`, `\`.
- Resolved path MUST be verified to stay inside the configured log directory (defense in depth).
- No auth/TLS in v1 (single-host assumption per spec); bind address is operator-controlled.

## Runtime Behavior

- Server errors are logged, never fatal to the monitor (constitution §Security & Runtime Constraints).
- HTTP server timeouts: Read 5s, Write 10s, Idle 120s.
- The embedded `patternfly.min.css` is build-critical — removing it breaks the build.

## HTML Requirements

- Log list: PatternFly card gallery, links to `/logs/<name>`.
- Log view: `<pre>` content, asset accordion fetching `/plugins/shell/assets/<encodeURIComponent(basename)>.json` client-side.
- Asset metadata displayed as pretty-printed JSON; missing asset ⇒ "No asset metadata available."
