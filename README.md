<p align="center">
  <img src="docs/logo.svg" alt="ReportDog Logo" width="120" />
</p>

<h1 align="center">ReportDog</h1>

<p align="center">
  <strong>Your faithful companion for test results.</strong><br/>
  Collect, visualize, and track test reports — so no failure slips through the cracks.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#what-it-does">What It Does</a> •
  <a href="#api-reference">API</a> •
  <a href="#configuration">Config</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/akhilbojedla/reportdog/actions/workflows/ci.yml"><img src="https://github.com/akhilbojedla/reportdog/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/akhilbojedla/reportdog/releases/latest"><img src="https://img.shields.io/github/v/release/akhilbojedla/reportdog" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/akhilbojedla/reportdog" alt="License"></a>
  <a href="https://hub.docker.com/r/akhilbojedla/reportdog"><img src="https://img.shields.io/docker/pulls/akhilbojedla/reportdog" alt="Docker Pulls"></a>
</p>

---

<p align="center">
  <img src="docs/reportdog-demo.gif" alt="ReportDog Demo" width="800" />
</p>

---

ReportDog is a self-hosted dashboard for your test results. Point your CI pipeline at it, and it will store every JUnit XML report you throw at it — giving you a single place to see what's passing, what's failing, and what's been flaky over time.

## What It Does

**A dashboard that tells the story at a glance** — Total reports, pass rates, failure trends, and your most recent runs — all on one screen. No digging through CI logs.

**Browse and search every report** — Find any test run by name, status, or tags. Filter down to just the failures, or search across hundreds of reports in seconds.

**Track tests across runs** — See how a specific test pipeline performs over time with execution history. Drill into an individual test case to spot intermittent failures that might otherwise go unnoticed.

**Tag and organize** — Attach key-value tags to reports (branch, environment, team, sprint — whatever makes sense for you). Click any tag to instantly filter reports. Tags autocomplete as you type, so you stay consistent without extra effort.

**Upload from anywhere** — Send reports from your CI pipeline via a simple API call, or upload them manually through the UI when you need to. Manual upload can be turned off if you want a purely API-driven workflow.

**Dark mode included** — A warm, amber-toned interface that looks good in both light and dark mode.

**Ships as a single Docker image** — One container runs both the UI and the API. Just add PostgreSQL and you're done.

## Architecture

ReportDog ships as a **single Docker image**. The Go backend serves both the API and the frontend static files.

```
┌──────────────────────────────┐       ┌────────────┐
│         ReportDog             │       │            │
│  ┌──────────┐ ┌───────────┐  │       │ PostgreSQL │
│  │ Frontend  │ │  API      │  │──────▶│   :5432    │
│  │ (static)  │ │  /api/v1  │  │       │            │
│  └──────────┘ └───────────┘  │       └────────────┘
│           :8080               │
└──────────────────────────────┘
```

| Component  | Tech                                          |
|------------|-----------------------------------------------|
| Frontend   | React 19, TypeScript, Vite, MUI v5, Recharts  |
| Backend    | Go, Gin, GORM                                 |
| Database   | PostgreSQL 16                                  |
| Infra      | Docker, Docker Compose                         |

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/) v2+

### Run with Docker Compose

```bash
git clone https://github.com/akhilbojedla/reportdog.git
cd reportdog
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

### Run with Pre-built Images

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: reportdog
      POSTGRES_PASSWORD: reportdog
      POSTGRES_DB: reportdog
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U reportdog"]
      interval: 5s
      timeout: 5s
      retries: 5

  reportdog:
    image: akhilbojedla/reportdog:latest
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_USER: reportdog
      DB_PASSWORD: reportdog
      DB_NAME: reportdog
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  pgdata:
```

```bash
docker compose up
```

## Development

### Backend

```bash
cd backend
go mod download
DB_HOST=localhost DB_PORT=5432 DB_USER=reportdog DB_PASSWORD=reportdog DB_NAME=reportdog go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```

The dev server runs on [http://localhost:5173](http://localhost:5173).

## Configuration

### Backend Environment Variables

| Variable            | Default       | Description                         |
|---------------------|---------------|-------------------------------------|
| `PORT`              | `8080`        | Server listen port                  |
| `DB_HOST`           | `localhost`   | PostgreSQL host                     |
| `DB_PORT`           | `5432`        | PostgreSQL port                     |
| `DB_USER`           | `reportdog`   | PostgreSQL user                     |
| `DB_PASSWORD`       | `reportdog`   | PostgreSQL password                 |
| `DB_NAME`           | `reportdog`   | PostgreSQL database name            |
| `CORS_ALLOW_ORIGIN`      | `*`          | Allowed CORS origin (for dev)       |
| `PUBLIC_DIR`             | `./public`   | Path to frontend static files       |
| `DISABLE_MANUAL_UPLOAD`  | _(unset)_    | Set to `true` to disable UI upload  |

### Frontend Environment Variables (dev only)

| Variable             | Default                   | Description                                  |
|----------------------|---------------------------|----------------------------------------------|
| `VITE_API_BASE_URL`  | _(empty — same origin)_   | Backend API base URL (set for separate dev)  |

## API Reference

All endpoints are prefixed with `/api/v1`.

### Reports

| Method   | Path                                            | Description                         |
|----------|-------------------------------------------------|-------------------------------------|
| `POST`   | `/reports/upload`                               | Upload a JUnit XML file             |
| `POST`   | `/reports/ingest`                               | Ingest raw XML body                 |
| `GET`    | `/reports`                                      | List reports (with search/filter)   |
| `GET`    | `/reports/:id`                                  | Get report details                  |
| `GET`    | `/reports/:id/raw`                              | Download original XML               |

### Tags

| Method   | Path                                            | Description                         |
|----------|-------------------------------------------------|-------------------------------------|
| `POST`   | `/reports/:id/tags`                             | Add tags to a report                |
| `DELETE` | `/reports/:id/tags`                             | Remove a tag (body: `{key, value}`) |
| `GET`    | `/tags`                                         | List all tags with counts           |
| `GET`    | `/tags/keys`                                    | List known tag keys (autocomplete)  |

### Executions & Tests

| Method   | Path                                            | Description                         |
|----------|-------------------------------------------------|-------------------------------------|
| `GET`    | `/executions/:name/reports`                     | Execution run history               |
| `GET`    | `/executions/:name/tests/:testName/history`     | Individual test history             |

### Stats & Metadata

| Method   | Path                                            | Description                         |
|----------|-------------------------------------------------|-------------------------------------|
| `GET`    | `/stats`                                        | Dashboard summary stats             |
| `GET`    | `/execution-names`                              | List all execution names            |

### Health Check

| Method | Path      | Description    |
|--------|-----------|----------------|
| `GET`  | `/health` | Liveness check |

### Upload Examples

**File upload:**

```bash
curl -X POST http://localhost:8080/api/v1/reports/upload \
  -F "file=@results.xml" \
  -H "X-Execution-Name: my-pipeline" \
  -H "X-Report-Name: unit-tests" \
  -H 'X-Tags: [{"key":"branch","value":"main"},{"key":"env","value":"ci"}]'
```

**Raw XML ingest:**

```bash
curl -X POST http://localhost:8080/api/v1/reports/ingest \
  -H "Content-Type: application/xml" \
  -H "X-Execution-Name: my-pipeline" \
  -H "X-Report-Name: integration-tests" \
  -d @results.xml
```

## Test Data Generation

A helper script generates sample reports for testing:

```bash
pip install requests
python scripts/generate_test_data.py
```

This creates ~500 diverse test reports with varied pipelines, tags, and failure rates.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) — free to use, modify, and distribute.
