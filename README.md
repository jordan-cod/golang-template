<p align="center">
  <img src="https://raw.githubusercontent.com/golang-samples/gopher-vector/master/gopher.png" alt="Gopher" width="200"/>
</p>

# Go Service Template

A lean, production-ready starting point for HTTP services in Go. Standard library first: no web framework, no logging library, no config library. Clone it, rename the module, and start adding handlers.

## Features

- **HTTP server** on `net/http` with Go 1.22+ `ServeMux` patterns (`GET /path/{id}`), read/write/idle timeouts
- **Graceful shutdown** on `SIGINT`/`SIGTERM`: readiness flips to 503, in-flight requests drain, bounded by `SHUTDOWN_TIMEOUT`
- **Probes**: `GET /healthz` (liveness) and `GET /readyz` (readiness)
- **Structured logging** with `log/slog`: JSON in production, text in development
- **Middleware**: request ID (`X-Request-ID`, propagated or generated), request logging, panic recovery
- **Config from environment** with defaults and validation that reports every problem at once
- **Hot reload** with [air](https://github.com/air-verse/air), locally or in Docker Compose
- **Docker**: multi-stage build, static binary on `distroless/static:nonroot`, plus a `dev` stage for Compose
- **Linting** with `golangci-lint` v2
- **Tests** with `httptest`, race detector and coverage
- **CI** on GitHub Actions: lint, test, build and Docker build, actions pinned to commit SHAs, read-only token
- **Dependabot** for Go modules, GitHub Actions and Docker base images

## Project structure

```
.
├── cmd/app/main.go            entrypoint: config, logger, signals, run
├── internal/
│   ├── config/                env parsing, defaults, validation
│   ├── logger/                slog setup (JSON or text)
│   ├── middleware/            request ID, logging, recover, Chain
│   └── server/                http.Server lifecycle, routes, handlers
├── .air.toml                  hot reload config
├── .github/
│   ├── dependabot.yml
│   └── workflows/ci.yml
├── .golangci.yml
├── Dockerfile                 base / dev / build / prod stages
├── docker-compose.yml         dev stack with hot reload
├── Makefile
└── .env.example
```

## Quickstart

Requirements: Go 1.27+. Docker is optional.

```sh
cp .env.example .env
set -a; . ./.env; set +a
make dev
```

```sh
curl -i localhost:8080/healthz
curl -i localhost:8080/readyz
```

Or with Docker, no local Go needed (reads `.env` if present, publishes on `HOST_PORT`, default `8080`):

```sh
make compose-up
```

Production image:

```sh
make docker-build
make docker-run
```

## Make targets

Run `make` or `make help` to list them.

| Target | Description |
|---|---|
| `dev` | Run with hot reload (installs air if missing) |
| `run` | Run without hot reload |
| `build` | Build a static binary into `bin/` |
| `test` | Run tests with `-race` and coverage (`coverage.out`) |
| `cover` | Open the HTML coverage report |
| `lint` | Run golangci-lint (installs it if missing) |
| `fmt` | Format code with gofmt and goimports |
| `tidy` | `go mod tidy` and `go mod verify` |
| `docker-build` | Build the production image |
| `docker-run` | Build and run the production image |
| `compose-up` / `compose-down` | Start / stop the dev stack |
| `clean` | Remove build artifacts |

## Configuration

All configuration comes from environment variables. The app does not read `.env` by itself: export it in your shell, or let Docker Compose load it.

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` (text logs) or `production` (JSON logs) |
| `PORT` | `8080` | Port to listen on |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `HTTP_READ_TIMEOUT` | `10s` | Max time to read a request, headers included |
| `HTTP_WRITE_TIMEOUT` | `15s` | Max time to write a response |
| `HTTP_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout |
| `SHUTDOWN_TIMEOUT` | `15s` | Max time to drain connections on shutdown |

Durations use Go syntax (`500ms`, `10s`, `1m`). Invalid values stop the app at startup with a message listing every invalid variable.

## Adding a route

Register handlers in `internal/server/routes.go`:

```go
mux.HandleFunc("GET /users/{id}", users.get)
```

Read path values with `r.PathValue("id")`. Every route goes through the request ID, logging and recover middleware.

## Using this template

1. Click **Use this template** on GitHub, or clone and drop the history:
   ```sh
   git clone https://github.com/jordan-cod/golang-template my-service
   cd my-service && rm -rf .git && git init
   ```
2. Rename the module everywhere:
   ```sh
   go mod edit -module github.com/you/my-service
   grep -rl 'github.com/jordan-cod/golang-template' --include='*.go' --include='.golangci.yml' . \
     | xargs sed -i.bak 's#github.com/jordan-cod/golang-template#github.com/you/my-service#g'
   find . -name '*.bak' -delete
   ```
3. Optionally rename `cmd/app` and update `APP` in the `Makefile` and the paths in `.air.toml` and `Dockerfile`.
4. Set `IMAGE` in the `Makefile`, then run `make lint test build`.

