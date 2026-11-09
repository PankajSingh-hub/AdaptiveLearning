# Setup

## Prerequisites

Go (see `go` line in `backend/go.mod`), Node 20+, MySQL 8, optionally Docker.

## Local

1. `mysql -u root -p < database/schema.sql`
2. `cd backend && cp .env.example .env` → set `DB_PASSWORD`, `JWT_SECRET` (`openssl rand -hex 32`), keep `AUTO_MIGRATE=false`
3. `go run ./cmd/server` — seeds demo users, 31 questions, 2 assessments on first start
4. `cd frontend && cp .env.example .env && npm install && npm run dev`

## Docker

```bash
cp .env.example .env     # root-level file
docker compose up --build
```

App: http://localhost:5173 · API: http://localhost:8080 · Docs: http://localhost:8080/swagger.
Reset the database: `docker compose down -v`.

## Tests

```bash
cd backend
go test ./internal/service/                      # unit tests
mysql -u root -p -e "CREATE DATABASE adaptive_assessment_test;"
TEST_DB_DSN='root:PASSWORD@tcp(localhost:3306)/adaptive_assessment_test?charset=utf8mb4&parseTime=True&loc=Local' \
  go test ./internal/service/ -v                 # + integration tests
```

The test database name must contain "test"; its tables are dropped and recreated.

## Deployment (untested)

Frontend → Vercel/Netlify (set `VITE_API_URL`). Backend → Render/Railway/AWS (set env vars, `ALLOWED_ORIGIN` to the frontend URL). Database → any MySQL-compatible cloud service; run `schema.sql` once. Change the seeded demo passwords before exposing anything publicly. None of this has been tested.
