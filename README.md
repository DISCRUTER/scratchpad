# ScratchPad

A server-rendered Go web application for creating and sharing self-destructing text snippets. Sign up, log in, publish a snippet with a title and content, and choose how long it lives (1 day, 1 week, or 1 year) before it disappears.

## Features

- Create and view text snippets with an expiry date (1 day, 1 week, or 1 year)
- Home page lists the 10 most recent snippets that haven't expired
- User authentication: signup, login, and logout
- Passwords hashed with bcrypt (cost 12)
- Session-based authentication (sessions stored in PostgreSQL)
- CSRF protection on all POST forms
- HTTPS/TLS with a hardened TLS 1.3 configuration
- Security headers (CSP, nosniff, frame protection, etc.)
- Flash messages on form submission
- JSON structured logging via `log/slog`
- Prometheus metrics exposed on a separate port with bounded-cardinality labels
- Static assets and templates embedded into the binary via `embed`
- Schema managed by [golang-migrate](https://github.com/golang-migrate/migrate); applied automatically in the container stack

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP server | Standard library `net/http` (Go 1.22 method + path routing) |
| Database | PostgreSQL 17 (`lib/pq`) |
| Sessions | `alexedwards/scs/v2` with `postgresstore` |
| Metrics | `prometheus/client_golang` on a dedicated metrics server |
| CSRF protection | `justinas/nosurf` |
| Middleware chaining | `justinas/alice` |
| Form decoding | `go-playground/form/v4` |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Templating | `html/template` with embedded filesystem |
| Deployment | Podman Compose (Docker Compose compatible) |
| Observability | Prometheus + Grafana (+ pgAdmin for the database) |

## Getting Started

### Prerequisites

- Go 1.26+
- Podman with podman-compose (or any Docker Compose v2 compatible tool)
- PostgreSQL 17 (via the compose stack, or locally for bare-metal development)

### Option A — Full containerized stack (recommended)

```sh
podman-compose up -d --build
```

This builds the app image plus a small migrator image, applies all migrations once the database is healthy, then starts everything else:

| Service | URL | Credentials |
|---|---|---|
| ScratchPad (HTTPS) | https://localhost:8080 | sign up in the app |
| Grafana | http://localhost:3000 | `admin` / `password` |
| Prometheus | http://localhost:9090 | n/a |
| pgAdmin | http://localhost:5050 | `admin@scratchpad.com` / `password` |
| PostgreSQL | internal only (`db:5432` on `scratchpad_net`) | `scratchpad` / `password`, db `scratchpad` |

> The credentials above are development dummies hardcoded in `docker-compose.yml`. Change them before exposing this stack anywhere.

The app container mounts `./tls` read-only and expects `cert.pem`/`key.pem` inside — generate them first (next section) if you don't have them yet.

### Option B — Local development (bare metal)

**1. Generate TLS certificates**

The server only speaks HTTPS and expects certificates at `./tls/cert.pem` and `./tls/key.pem` (both are gitignored):

```sh
mkdir -p tls
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout tls/key.pem -out tls/cert.pem -days 365 \
  -subj "/CN=localhost"
```

**2. Start PostgreSQL**

Either use just the database from the compose stack:

```sh
make compose-up          # or: podman-compose up -d db
```

or point at any existing PostgreSQL instance via `DSN`.

**3. Apply migrations**

```sh
make migrate-up          # uses ~/go/bin/migrate if not on PATH
# rollback one step: make migrate-down
```

Raw CLI equivalent:

```sh
migrate -path=./migrations \
  -database="postgres://scratchpad:password@localhost/scratchpad?sslmode=disable" up
```

**4. Run the app**

```sh
make run
```

Then open https://localhost:8080 (accept the self-signed certificate warning). Metrics are served separately on http://localhost:8081/metrics.

## Configuration

Every setting is a CLI flag with an environment-variable fallback:

| Setting | Flag | Environment variable | Default |
|---|---|---|---|
| App listen port | `-port` | `PORT` | `8080` |
| Metrics listen port | `-metrics_port` | `METRICS_PORT` | `8081` |
| PostgreSQL DSN | `-dsn` | `DSN` | `postgres://scratchpad:password@localhost/scratchpad?sslmode=disable` |
| Max open DB connections | `-db-max-open-conns` | – | `25` |
| Max idle DB connections | `-db-max-idle-conns` | – | `25` |
| Max idle connection time | `-db-max-idle-time` | – | `15m` |

> **Note:** the app serves over HTTPS only. Sessions and CSRF cookies are marked `Secure`, so the site won't function over plain HTTP.

## Routes

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/` | Public | Home page with the 10 latest pads |
| GET | `/pads/view/{id}` | Public | View a single pad |
| GET | `/pads/create` | Required | Show the create-snippet form |
| POST | `/pads/create` | Required | Create a snippet |
| GET | `/user/signup` | Public | Show the signup form |
| POST | `/user/signup` | Public | Create an account |
| GET | `/user/login` | Public | Show the login form |
| POST | `/user/login` | Public | Log in |
| POST | `/user/logout` | Required | Log out |
| GET | `/ping` | Public | Health check, responds `OK` |
| GET | `/static/` | Public | Static assets |
| GET | `/metrics` | Internal | Prometheus scrape endpoint (separate port) |

## Metrics

Metrics live on their own listener (default `:8081`) behind a dedicated Prometheus registry, so scrape traffic never touches the application mux.

| Metric | Type | Labels |
|---|---|---|
| `scratchpad_http_request_total` | counter | `method`, `endpoint`, `status_code` |
| `scratchpad_http_request_duration_seconds` | histogram | `method`, `endpoint` |
| `scratchpad_http_active_requests` | gauge | – |

The `endpoint` label is the matched **route pattern** (e.g. `GET /pads/view/{id}`), never the raw request path — parameterized URLs can't blow up label cardinality. Requests that don't match any route are labeled `unmatched`. Panics recovered by middleware still produce a recorded `500` series, and the active-request gauge is decremented even on panic paths.

Grafana's Prometheus datasource is provisioned automatically from `grafana/datasources.yaml`; the scrape config lives in `prometheus/prometheus.yml`.

## Database & Migrations

Three tracked migrations create the schema: `pads`, `users` (unique email constraint), and `sessions` (token/data/expiry used by scs's `postgresstore`).

```sh
make migrate-up     # apply all pending
make migrate-down   # roll back the latest
DSN="postgres://user:pass@host/dbname?sslmode=disable" make migrate-up   # custom target
```

In the container stack, the one-shot `migrate` service runs automatically before the app starts (`depends_on: service_completed_successfully`), so `podman-compose down && podman-compose up` re-applies pending migrations safely.

## Tests

```sh
make test         # verbose go test ./...
make test-race    # race detector, no caching (same as CI)
```

- `cmd/web` — handler, middleware, template, and metrics tests run fully offline against model mocks (`httptest` + embedded templates).
- `internal/models` — data-layer integration tests that connect to PostgreSQL at the default local DSN. Their fixtures **reset** the `users`, `pads`, and `sessions` tables of the target database, so run them against a disposable database.
- CI (`.github/workflows/ci.yml`) builds, vets, and runs the whole suite with `-race` against a `postgres:17` service container.

## Project Structure

```
cmd/web/              Web application entrypoint: main, routes, handlers,
                      middleware, helpers, templates, context keys
internal/metrics/     Prometheus collectors, registry wiring, /metrics handler
internal/models/      Data access layer for pads and users, sentinel errors,
                      and mocks used by offline tests
internal/validator/   Form validation helpers (embedded in form structs)
internal/assert/      Tiny generic test assertion helpers
migrations/           golang-migrate SQL migrations (up/down pairs)
prometheus/           Scrape config mounted into the Prometheus container
grafana/              Datasource provisioning mounted into Grafana
ui/                   Embedded HTML templates, static CSS, and assets (embed.FS)
Dockerfile            Multi-stage build for the web binary
Containerfile.migrate Builds the pinned golang-migrate CLI image
docker-compose.yml    Full stack: db, migrate, scratchpad, prometheus, grafana, pgadmin
Makefile              Common dev tasks
tls/                  TLS certificates (gitignored)
```

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Build the binary to `./build/web` |
| `make fmt` | Format all Go source files |
| `make run` | Build and run the server |
| `make test` | Run the test suite verbosely |
| `make test-race` | Run the suite with the race detector (`-race -count=1`) |
| `make migrate-up` | Apply pending migrations (honors `DSN=` override) |
| `make migrate-down` | Roll back the most recent migration |
| `make compose-up` | Build and start the full container stack |
| `make compose-down` | Tear down the container stack |
