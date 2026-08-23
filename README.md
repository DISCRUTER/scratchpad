# ScratchPad

A server-rendered Go web application for creating and sharing self-destructing text snippets. Sign up, log in, publish a snippet with a title and content, and choose how long it lives (1 day, 1 week, or 1 year) before it disappears.

## Features

- Create and view text snippets with an expiry date (1 day, 1 week, or 1 year)
- Home page lists the 10 most recent snippets that haven't expired
- User authentication: signup, login, and logout
- Passwords hashed with bcrypt (cost 12)
- Session-based authentication (sessions stored in MySQL)
- CSRF protection on all POST forms
- HTTPS/TLS with a hardened TLS 1.3 configuration
- Security headers (CSP, nosniff, frame protection, etc.)
- Flash messages on form submission
- JSON structured logging via `log/slog`
- Static assets and templates embedded into the binary via `embed`

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP server | Standard library `net/http` (Go 1.22 method + path routing) |
| Database | MySQL 8+ |
| Sessions | `alexedwards/scs/v2` with `mysqlstore` |
| CSRF protection | `justinas/nosurf` |
| Middleware chaining | `justinas/alice` |
| Form decoding | `go-playground/form/v4` |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Templating | `html/template` with embedded filesystem |
| Deployment | Docker / docker-compose |

## Getting Started (Local Development)

### Prerequisites

- Go 1.26+
- MySQL 8+ (or Docker)

### 1. Set up the database

The fastest way is to use the included compose file, which creates a `scratchpad` database and applies `db/schema.sql` automatically:

```sh
docker compose up -d db
```

Alternatively, create a database and import the schema manually:

```sh
mysql -u root -p -e "CREATE DATABASE scratchpad CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
mysql -u root -p scratchpad < db/schema.sql
```

### 2. Generate TLS certificates

The server starts with `ListenAndServeTLS` and expects certificates at `./tls/cert.pem` and `./tls/key.pem` (both are gitignored). For local development, generate self-signed certs:

```sh
mkdir -p tls
openssl req -x509 -newkey rsa:2048 -nodes -keyout tls/key.pem -out tls/cert.pem -days 365 -subj "/CN=localhost"
```

### 3. Run the app

```sh
make run
```

This builds the binary and starts the server. By default it listens on `:4000` and connects to MySQL with `web:pass@tcp(localhost:3306)/scratchpad`. Both can be overridden:

| Setting | Flag | Environment variable | Default |
|---|---|---|---|
| Listen address | `-addr` | `ADDR` | `:4000` |
| MySQL DSN | `-dsn` | `DSN` | `web:pass@tcp(localhost:3306)/scratchpad?parseTime=true` |

```sh
make build && ADDR=:8080 ./build/web
```

Open `https://localhost:4000` in your browser (accept the self-signed cert warning).

> **Note:** the app serves over HTTPS only. Sessions and CSRF cookies are marked `Secure`, so the site won't function over plain HTTP.

## Migration Commands

>[!IMPORTANT]
> The migration is performed using goland-migration cli tool and therfore, is adivsed to install it before running migrations.

```sh
# Migrate Up
~/go/bin/migrate -path=./migrations -database=postgres://scratchpad:password@localhost/scratchpad?sslmode=disable up
# Migrate Down
 ~/go/bin/migrate -path=./migrations -database=postgres://scratchpad:password@localhost/scratchpad?sslmode=disable down
```

## Docker Deployment

Copy `.env.example` to `.env` and set the database credentials, then start the stack:

```sh
cp .env.example .env
# edit .env, e.g. ADDR=:8080, DB_ROOT_PASSWORD, DB_PASSWORD
docker compose up -d --build
```

The `web` container serves the app and the `db` container runs MySQL with a healthcheck; the web service only starts once the database is healthy.

> **Note:** the app defaults to port `4000`, but `docker-compose.yml` maps host port `8080`. Set `ADDR=:8080` in `.env` (or the compose `environment` block) so the container listens on the mapped port.

## Project Structure

```
cmd/web/              Web application entrypoint (main, routes, handlers, middleware, helpers, templates)
internal/models/      Data access layer for pads and users + mocks used in tests
internal/validator/   Form validation helpers
internal/assert/      Test assertion helpers
db/schema.sql         Database schema
ui/                   Embedded HTML templates, static CSS, and assets (via embed.FS)
build/                Compiled binary output (gitignored)
tls/                  TLS certificates (gitignored)
Dockerfile            Multi-stage build for the web binary
docker-compose.yml    MySQL + web services
Makefile              Common dev tasks (build, fmt, run, test)
```

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

## Tests

```sh
make test
```

Runs the full test suite (`go test -v ./...`), covering handlers, middleware, templates, and the data layer. Data-layer tests run against MySQL and use the fixtures in `internal/models/testdata/`.

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Build the binary to `./build/web` |
| `make fmt` | Format all Go source files |
| `make run` | Build and run the server |
| `make test` | Run the test suite verbosely |
