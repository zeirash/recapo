# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Recapo is an order management system for Jastipers (Indonesian cross-border social media sellers). It consists of two services:
- **arion** — Go REST API backend (port 4000)
- **oncius** — Next.js 14 frontend (port 3000)

## Development Commands

### Backend (arion)
```bash
cd arion
go run main.go          # Start dev server on :4000
go test ./...           # Run all tests
go test ./handler/...   # Run tests in a specific package
go test -run TestName ./handler/...  # Run a single test
swag init -g main.go    # Regenerate Swagger docs
```

### Frontend (oncius)
```bash
cd oncius
npm run dev             # Start dev server on :3000
npm run build           # Production build
npm run lint            # ESLint
npm run type-check      # TypeScript check without emit
```

### Infrastructure
```bash
docker compose up -d postgres   # Start PostgreSQL only
docker compose up -d            # Start all services
```

## Backend Architecture (arion)

Clean layered architecture: **Handler → Service → Store**

- `handler/` — HTTP handlers using Gorilla mux; each file corresponds to a domain (order, customer, product, user, shop, authentication)
- `service/` — Business logic; each handler package has a corresponding service with interface-based design
- `store/` — Database queries against PostgreSQL via `lib/pq`
- `model/model.go` — All shared data structs
- `common/` — Cross-cutting concerns: config (env vars), database connection, logger (logrus), middleware (JWT auth, recovery), response helpers, i18n

**Handler initialization pattern:** Each handler file initializes its service via `Init()` and exposes a `Set*Service()` setter for test injection.

**Middleware:** Routes use `middleware.ChainMiddleware(middleware.Authentication, ...)` for protected endpoints. Unauthenticated routes use `r.HandleFunc` directly.

**API response envelope** (all endpoints):
```json
{ "success": bool, "data": <payload>, "code": string, "message": string }
```

**Testing:** All three layers have `*_test.go` files. Uses `DATA-DOG/go-sqlmock` for store tests and `golang/mock` for service/handler tests. Mocks live in `mock/`.

**Backend environment variables** (copy from `.env.example`):
```
ENV, SERVICE_NAME, SERVICE_PORT, DB_NAME, DB_HOST, DB_USERNAME, DB_PASSWORD, DB_PORT, SECRET_KEY
```

## Frontend Architecture (oncius)

Next.js 14 App Router application.

- `src/app/` — Page segments: `orders/`, `customers/`, `products/`, `purchase/`, `temp_orders/`, `share/`, `dashboard/`, `login/`, `register/`
- `src/components/` — Shared UI components; Layout wrapper in `Layout/`
- `src/hooks/` — `useAuth.ts` (JWT auth state), `useLocale.ts` (i18n)
- `src/utils/api.ts` — Centralized API client
- `src/types/` — TypeScript type definitions
- `src/providers/` — React context providers

**Styling:** MUI (Material UI) + Emotion + SASS. No CSS modules.

**Data fetching:** React Query 3.x (`@tanstack/react-query` not used; it's `react-query` v3).

**i18n:** `next-intl` with config in `src/i18n.ts` and `src/middleware.ts`.

**API proxy:** `next.config.js` rewrites `/api/:path*` to the backend (`API_BASE_URL` env var, defaults to `http://localhost:3000`).

**Path aliases:** `@/*` maps to `src/*`.

## Database

Single migration file: `arion/migrations/000_ddl_all_tables.sql`. Apply manually before first run.

Database name: `recapo_master` (PostgreSQL 15).

## Workflow Rules

- Always update unit tests when changing `arion/store`, `arion/service`, or `arion/handler`

## Use the mui-mcp server to answer any MUI questions --

- 1. call the "useMuiDocs" tool to fetch the docs of the package relevant in the question
- 2. call the "fetchDocs" tool to fetch any additional docs if needed using ONLY the URLs present in the returned content.
- 3. repeat steps 1-2 until you have fetched all relevant docs for the given question
- 4. use the fetched content to answer the question

## Use lucide-react for icon library
- Always use icons from lucide-react for all icons
