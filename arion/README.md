# Arion

Go REST API backend for Recapo. Runs on port 4000.

## Architecture

Clean layered architecture: **Handler → Service → Store**

- `handler/` — HTTP handlers (Gorilla mux)
- `service/` — Business logic
- `store/` — PostgreSQL queries via `lib/pq`
- `model/model.go` — Shared data structs
- `common/` — Config, database, logger, middleware, response helpers, i18n

## Development

```bash
cp .env.example .env.local  # fill in values
go run .
```

Server runs on http://localhost:4000

## Testing

```bash
go test ./...                          # all tests
go test ./handler/...                  # specific package
go test -run TestName ./handler/...    # single test
```

## Swagger

```bash
swag init -g main.go   # regenerate docs after changing handler godoc comments
```

Swagger UI: http://localhost:4000/swagger/

## Docker

```bash
docker build -t arion .
docker run -p 4000:4000 arion
```

## Environment Variables

Copy `.env.example` to `.env.local` and fill in:

| Variable | Description |
|---|---|
| `ENV` | `development` or `production` |
| `PORT` | Server port (default 4000) |
| `DB_*` | PostgreSQL connection details |
| `SECRET_KEY` | JWT signing key |
| `SENTRY_DSN` | Sentry error tracking (optional) |
| `MIDTRANS_SERVER_KEY` | Midtrans payment gateway |
| `RESEND_API_KEY` | Resend email service |
| `R2_*` | Cloudflare R2 object storage (optional, falls back to local filesystem) |
| `GITHUB_TOKEN` | GitHub API for feedback issues (optional) |

## Metrics

Prometheus metrics exposed at `/metrics`. Scraped by Grafana Alloy in production.
