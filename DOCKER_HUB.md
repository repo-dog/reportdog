# ReportDog

**Your faithful companion for test results.**  
Collect, visualize, and track JUnit XML reports — so no failure slips through the cracks.

![ReportDog Demo](https://raw.githubusercontent.com/repo-dog/reportdog/main/docs/reportdog-demo.gif)

---

## Quick Start

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: reportdog
      POSTGRES_PASSWORD: reportdog
      POSTGRES_DB: reportdog
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

Open **http://localhost:8080** in your browser.

---

## Send Reports from CI

**GitHub Actions / GitLab CI / Jenkins — any tool that can run `curl`:**

```bash
curl -X POST http://your-reportdog-host/api/v1/reports/ingest \
  -H "Content-Type: application/xml" \
  -H "X-Execution-Name: my-pipeline" \
  -H "X-Tags: branch:main,env:ci" \
  -d @results.xml
```

Or as a multipart upload:

```bash
curl -X POST http://your-reportdog-host/api/v1/reports/upload \
  -F "file=@results.xml" \
  -F "execution_name=my-pipeline" \
  -F "tags=branch:main,env:ci"
```

---

## Configuration

| Variable               | Default      | Description                                          |
|------------------------|--------------|------------------------------------------------------|
| `DB_HOST`              | `localhost`  | PostgreSQL host                                      |
| `DB_PORT`              | `5432`       | PostgreSQL port                                      |
| `DB_USER`              | `reportdog`  | PostgreSQL user                                      |
| `DB_PASSWORD`          | `reportdog`  | PostgreSQL password                                  |
| `DB_NAME`              | `reportdog`  | PostgreSQL database name                             |
| `PORT`                 | `8080`       | Server listen port                                   |
| `DISABLE_MANUAL_UPLOAD`| _(unset)_    | Set to `true` to disable the UI upload form          |
| `AUTO_MIGRATE`         | `true`       | Set to `false` to manage database migrations manually |
| `MIGRATIONS_DIR`       | `migrations` | Path to `.sql` migration files inside the container  |

---

## Full Documentation

→ **[github.com/akhilbojedla/reportdog](https://github.com/akhilbojedla/reportdog)**

Includes the complete CI integration guide, manual migration instructions, development setup, and API reference.
