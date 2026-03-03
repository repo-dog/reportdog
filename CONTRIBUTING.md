# Contributing to ReportDog

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

1. **Prerequisites:** Go 1.22+, Node.js 20+, Docker & Docker Compose
2. **Clone the repo:**
   ```bash
   git clone https://github.com/repo-dog/reportdog.git
   cd reportdog
   ```
3. **Start the database:**
   ```bash
   docker compose up postgres -d
   ```
4. **Run the backend:**
   ```bash
   cd backend
   go mod download
   DB_HOST=localhost DB_PORT=5432 DB_USER=reportdog DB_PASSWORD=reportdog DB_NAME=reportdog go run ./cmd/server
   ```
5. **Run the frontend:**
   ```bash
   cd frontend
   npm install
   VITE_API_BASE_URL=http://localhost:8080 npm run dev
   ```

## Making Changes

1. Fork the repository and create a feature branch from `main`.
2. Make your changes with clear, descriptive commits.
3. Ensure the backend builds cleanly: `cd backend && go build ./... && go vet ./...`
4. Ensure the frontend builds cleanly: `cd frontend && npm run lint && npx tsc --noEmit && npm run build`
5. Open a pull request against `main`.

## Project Structure

```
reportdog/
├── backend/
│   ├── cmd/server/         # Entrypoint
│   ├── internal/
│   │   ├── db/             # Database connection & migrations
│   │   ├── handlers/       # HTTP handlers
│   │   ├── models/         # GORM models
│   │   ├── router/         # Route definitions
│   │   └── services/       # Business logic
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/            # Axios API client
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components
│   │   ├── theme.ts        # MUI theme
│   │   └── types.ts        # TypeScript interfaces
│   ├── Dockerfile
│   └── package.json
├── scripts/                # Helper scripts
├── docker-compose.yml
└── README.md
```

## Code Style

- **Go:** Standard `gofmt` formatting, `go vet` must pass.
- **TypeScript:** ESLint configuration in the frontend; run `npm run lint`.
- Keep functions small and focused.
- Add comments for non-obvious logic.

## Reporting Issues

- Use [GitHub Issues](https://github.com/repo-dog/reportdog/issues) to report bugs or request features.
- Include steps to reproduce for bugs.
- Check existing issues before opening a new one.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
