# Recapo

Order management system for Jastipers (Indonesian cross-border social media sellers).

## Services

| Service | Description | Port |
|---|---|---|
| arion | Go REST API backend | 4000 |
| oncius | Next.js 14 frontend | 3000 |
| postgres | PostgreSQL 15 database | 5432 |

## Project Structure

```
recapo/
├── arion/           # Go REST API backend
├── oncius/          # Next.js 14 frontend
├── monitoring/      # Prometheus, Grafana, Alloy configs
├── docker-compose.yml
└── render.yaml      # Render deployment blueprint
```

## Quick Start

### Prerequisites
- Go 1.20+
- Node.js 18+
- Docker

### 1. Start the database
```bash
docker compose up -d postgres
```

### 2. Start the backend
```bash
cd arion
cp .env.example .env.local  # fill in values
go run .
```

### 3. Start the frontend
```bash
cd oncius
cp .env.example .env.local  # fill in values
npm install
npm run dev
```

## Docker Compose

```bash
# Start all services
docker compose up -d

# Start only the database
docker compose up -d postgres
```

## Monitoring

### Local (Prometheus + Grafana)
```bash
docker compose up -d prometheus grafana
```
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin / recapo_grafana)

### Production (Grafana Alloy → Grafana Cloud)

Alloy scrapes arion and pushes metrics to Grafana Cloud. Required env vars:
```
GRAFANA_REMOTE_WRITE_URL=
GRAFANA_USER_ID=
GRAFANA_API_KEY=
ARION_URL=
```

Test locally:
```bash
docker compose --env-file .env.production --profile prod up alloy-prod
```

## Deployment

Deployed on Render via `render.yaml` blueprint. Services:
- `recapo-backend` — arion
- `recapo-frontend` — oncius
- `recapo-db` — PostgreSQL (free tier, expires after 90 days)
- `recapo-alloy` — Grafana Alloy (pushes metrics to Grafana Cloud)

Secrets set manually in Render dashboard — see comments in `render.yaml`.

## Logs

```bash
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs --tail=100 backend
```
