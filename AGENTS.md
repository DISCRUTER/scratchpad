# AGENTS.md

Context for AI coding agents working in this repository. Read this before making changes; it documents architecture, conventions, and invariants that are easy to break silently.

## What this is

ScratchPad — a server-rendered Go web app for sharing self-destructing text snippets. Standard-library-only HTTP layer (`net/http` ServeMux with Go 1.22+ method/pattern routing), PostgreSQL persistence, session auth with CSRF, embedded templates, Prometheus metrics on a separate listener, and a Podman Compose stack wiring app + db + migrations + Prometheus + Grafana + pgAdmin.

## Commands

```sh
make build            # go build -o ./build/web ./cmd/web   (main package lives in cmd/web)
make run              # build + run; reads PORT / METRICS_PORT / DSN env vars
make test             # go test -v ./...
make test-race        # go test -race -count=1 ./...        (what CI runs)
go vet ./...          # must stay clean
gofmt -l cmd internal # must output nothing
make migrate-up       # golang-migrate up (DSN=... to override target)
make compose-up       # podman-compose up -d --build
```

- `cmd/web` tests are fully offline (model mocks + httptest).
- `internal/models` tests are integration tests: they need a live PostgreSQL at `postgres://scratchpad:password@localhost/scratchpad?sslmode=disable` and **reset** the users/pads/sessions tables via `testdata/setup.sql`/`teardown.sql` (idempotent `DROP TABLE IF EXISTS` first).

## Architecture

### Request pipeline — order matters

```
trackMetrics → recoverPanic → logRequest → commonHeaders → mux
mux route chains:
  dynamic   = sessionManager.LoadAndSave → noSurf(CSRF) → authenticate
  protected = dynamic → requireAuthentication
```

Invariants:

1. `trackMetrics` MUST wrap `recoverPanic` (be outermost of the two). It records status via a wrapping `statusRecorder`; if `recoverPanic` were outside it, panic responses would be recorded as 200.
2. Recording happens in a `defer` inside `trackMetrics`. Do not move it after `next.ServeHTTP` — panics skip it and leak the `ActiveRequests` gauge upward forever.
3. Nothing between `trackMetrics` and `mux.ServeHTTP` may clone the request (e.g. `r.WithContext`). The endpoint label comes from `r.Pattern`, which `ServeMux.ServeHTTP` writes onto the same `*http.Request` during dispatch. A shallow copy between them means `trackMetrics` reads a stale request where `Pattern` stays empty.

### Routing

Go 1.22+ ServeMux patterns (`"GET /pads/view/{id}"`, `"GET /{$}"`). Route table lives in `routes.go`; handlers in `handlers.go`; `ping` is a plain function used by health checks/tests.

### Key files

| File | Responsibility |
|---|---|
| `cmd/web/main.go` | Flags/env config, TLS server, separate metrics server goroutine, DB pool setup |
| `cmd/web/routes.go` | Mux registration, alice chains, `nueturedFileSystem` (blocks directory listing) |
| `cmd/web/middleware.go` | Headers, logging, panic recovery, auth gate, CSRF, `trackMetrics` + `statusRecorder` |
| `cmd/web/handlers.go` | All HTTP handlers |
| `cmd/web/helper.go` | `render` (buffered template exec), `serverError`/`clientError`, `newTemplateData`, `decodePostForm` |
| `cmd/web/templates.go` | Template cache built from `ui.Files` embed.FS with funcmap (`humanDate`) |
| `cmd/web/context.go` | `contextKey` string type + `isAuthenticatedContextKey` |
| `internal/metrics/metrics.go` | Collector definitions; `GetMetrics()` returns custom-registry-backed Metrics + mux |
| `internal/models/` | `PadsModel`/`UserModel` implementations, interfaces, sentinel errors, `mocks/` for tests |
| `internal/validator/` | Embeddable `Validator` struct + field checks |
| `internal/assert/` | Generic `Equal`, `StringContains`, `NilError` helpers |

## Conventions

- **Comments**: short doc comments above functions/types, matching existing style.
- **Errors**: handlers call `app.serverError(w, r, err)` (500 + log) or `app.clientError(w, status)`; sentinel errors (`models.ErrNoRecord`, `ErrDuplicateEmail`, `ErrInvalidCredentials`) are matched with `errors.Is`.
- **Validation**: form structs embed `validator.Validator`; checks via `form.CheckField(validator.NotBlank(...), "field", "msg")`; re-render template with 422 on failure.
- **Rendering**: always through `app.render` (executes into a buffer before writing, so template errors become 500s instead of half-written pages).
- **CSRF/session**: all POSTs require a valid token (nosurf); flash messages via `sessionManager.Put/PopString(ctx, "flash")`; auth state stored as `authenticatedUserID` int in the session, mirrored into context by `authenticate` middleware.
- **`decodePostForm` deliberately panics** on `form.InvalidDecoderError` — that's caught by `recoverPanic`. Don't "fix" it into a returned error without changing the handler flow.
- No external router, DI, or ORM by design; prefer stdlib additions.

## Metrics subsystem — read before touching

- `GetMetrics()` creates a **private registry** (not `prometheus.DefaultRegisterer`) and returns collectors + a dedicated `/metrics` mux served on its own port (default 8081). This isolates scrape traffic from the app mux.
- Collectors: `scratchpad_http_request_total` (counter, labels method/endpoint/status_code), `scratchpad_http_request_duration_seconds` (histogram, method/endpoint), `scratchpad_http_active_requests` (gauge).
- **Cardinality rule**: the `endpoint` label is only ever the bounded route pattern (`r.Pattern`, e.g. `"GET /pads/view/{id}"`) or the literal `"unmatched"` fallback. Never use `r.URL.Path` or any per-request identifier as a label.
- `r.Response.StatusCode` does NOT exist for server requests (`Request.Response` is nil there). Status codes come from `statusRecorder` (first `WriteHeader` wins; implicit 200 on bare `Write`; implements `Unwrap()` for `http.ResponseController` passthrough).
- `main.go` intentionally exits the process if the metrics listener fails (fail-fast, mirrors main server behavior).
- Tests (`cmd/web/metrics_test.go`) build their own registry + scrape exposition text via `promhttp` rather than using `prometheus/testutil` — importing testutil triggers `go mod tidy` churn. Assertion strings rely on prometheus text encoding sorting labels alphabetically; keep label sets stable.

## Database

- Driver `lib/pq`; DSN format `postgres://user:pass@host/db?sslmode=disable` (flag `-dsn`, env `DSN`).
- Pool tuned via flags (max open/idle conns, idle time); startup pings with a 5s timeout context.
- Migrations in `migrations/` (`000001_create_pads_table`, `000002_create_users_table`, `000003_create_session_table`), applied with golang-migrate. In compose, the one-shot `migrate` service gates app start via `depends_on: condition: service_completed_successfully`.

## Config & TLS

| Flag | Env | Default |
|---|---|---|
| `-port` | `PORT` | `8080` |
| `-metrics_port` | `METRICS_PORT` | `8081` |
| `-dsn` | `DSN` | `postgres://scratchpad:password@localhost/scratchpad?sslmode=disable` |
| `-db-max-open-conns` / `-db-max-idle-conns` / `-db-max-idle-time` | – | 25 / 25 / 15m |

HTTPS only (TLS 1.3 min, hardened curve prefs); certs expected at `./tls/cert.pem` and `./tls/key.pem` relative to working directory (gitignored, self-signed locally). Session + CSRF cookies are `Secure`, so plain HTTP will not work.

## Containers

- `Dockerfile`: multi-stage `docker.io/library/golang:1.26-alpine` → static binary (`CGO_ENABLED=0`, `./cmd/web`), non-root runtime user, `EXPOSE 8080 8081`. TLS certs are mounted, never baked.
- `Containerfile.migrate`: builds pinned golang-migrate CLI (`ARG MIGRATE_VERSION`, postgres tag) from the local golang image — chosen because `migrate/migrate` isn't pulled locally.
- `docker-compose.yml` services: `db` (postgres:17, healthcheck), `migrate` (one-shot), `scratchpad` (service name matches the `prometheus.yml` target `scratchpad:8081` — rename both together), `prometheus`, `grafana` (datasource provisioned from `grafana/datasources.yaml`), `pgadmin`.
- Dev dummy credentials everywhere: db `scratchpad/password/scratchpad`, Grafana `admin/password`, pgAdmin `admin@scratchpad.com/password`. Not production-safe by intent.
- Image tags are pinned to what's available on the local machine (`podman images`); bumping tags may require pulls.

## Gotchas

- `os.Exit` in `main` skips deferred calls (e.g. `db.Close`) — acceptable here, but don't add cleanup that depends on defers running at shutdown.
- `Request.SetPattern` doesn't exist in this toolchain; set/read the exported `Pattern` field directly.
- `WithContext` never propagates values back up to parent requests — middleware below can't hand data to middleware above via context injection.
- CI (`.github/workflows/ci.yml`) runs build + vet + `go test -race -count=1 -v ./...` with a `postgres:17` service; keep the suite offline-capable except `internal/models`, which uses that service DB.
- README.md is user-facing docs; AGENTS.md (this file) is agent-facing. Update both when behavior changes.
