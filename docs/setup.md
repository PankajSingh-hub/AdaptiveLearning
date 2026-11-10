# Setup

## Prerequisites

- Go (see `go` line in `backend/go.mod`)
- Node.js 20+
- MySQL 8.0 (for local setup)
- Docker and Docker Compose (for Docker setup)

## Local Setup

1. Make sure MySQL 8.0 is running.

2. Create the database:

   ```bash
   mysql -u root -p -e "CREATE DATABASE adaptive_assessment CHARACTER SET utf8mb4;"
   ```

3. Configure the backend environment:

   ```bash
   cd backend
   cp .env.example .env
   ```

   Set the following values in `backend/.env`:
   - `DB_PASSWORD` — your MySQL password
   - `JWT_SECRET` — generate using `openssl rand -hex 32`
   - `RUN_MIGRATIONS=true`

4. Start the backend:

   ```bash
   go run ./cmd/server
   ```

   On startup, the backend applies pending database migrations and seeds demo users, 31 questions, and 2 assessments.

5. Open another terminal and start the frontend:

   ```bash
   cd frontend
   cp .env.example .env
   npm install
   npm run dev
   ```

## Docker Setup

From the project root:

```bash
cp .env.example .env
docker compose up -d --build
```

Docker Compose waits for MySQL's healthcheck before starting the backend. The backend applies pending migrations and runs the configured seed process.

- App: http://localhost:5173
- API: http://localhost:8080
- API documentation: http://localhost:8080/swagger

To stop the application:

```bash
docker compose down
```

To reset the Docker database and delete its stored data:

```bash
docker compose down -v
docker compose up -d --build
```

**Warning:** `docker compose down -v` permanently deletes the Docker database volume and its data.

## Tests

```bash
cd backend

# Unit tests
go test ./internal/service/

# Create a test database
mysql -u root -p -e "CREATE DATABASE adaptive_assessment_test;"
```

Run integration tests:

```bash
TEST_DB_DSN='root:PASSWORD@tcp(localhost:3306)/adaptive_assessment_test?charset=utf8mb4&parseTime=True&loc=Local' \
go test ./internal/service/ -v
```

Replace `PASSWORD` with your MySQL password.

The test database name must contain `test`; its tables are dropped and recreated.

## Deployment (Untested)

- Frontend: Vercel or Netlify (configure `VITE_API_URL`).
- Backend: Render, Railway, or AWS (configure the required environment variables).
- Database: Any compatible MySQL cloud service.

Configure `ALLOWED_ORIGIN` with the deployed frontend URL and enable database migrations during backend startup using `RUN_MIGRATIONS=true`.

Change all seeded demo passwords and use a strong `JWT_SECRET` before exposing the application publicly.

**Note:** Deployment instructions have not been tested.
