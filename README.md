# User Management API (Go)

A small backend application with JWT authentication, user management endpoints, and a simple web dashboard.

## Features

- JWT sign-in with email/password
- Protected user endpoints (`GET /users`, `GET /users/{id}`, `PUT /users/{id}`)
- Optional PostgreSQL persistence (falls back to in-memory store)
- Input validation with JSON error responses
- Compact web UI served from `/`

## Requirements

- Go 1.24+
- Docker + Docker Compose (optional, for containerized run)

## Quick Start (Local)

```bash
go run ./cmd/server
```

Open:

- http://localhost:8080/ (web UI)
- http://localhost:8080/health

## Quick Start (Docker)

```bash
docker compose up --build -d
```

Open:

- http://localhost:8081/ (web UI)
- http://localhost:8081/health

Stop:

```bash
docker compose down
```

Stop and remove Postgres volume:

```bash
docker compose down -v
```

## Configuration

Environment variables:

- `PORT` (default: `8080`)
- `JWT_SECRET` (default: `development-secret`)
- `JWT_ISSUER` (default: `interviewsw`)
- `DATABASE_URL` (optional; when set, PostgreSQL store is used)

Example PostgreSQL run without Docker app container:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/interviewsw?sslmode=disable"
go run ./cmd/server
```

Notes:

- If `DATABASE_URL` is not set, the app uses in-memory storage.
- On PostgreSQL startup, schema is ensured and seed data is inserted only when `users` is empty.

## Seed Users

- `alice@example.com` / `Password123`
- `bob@example.com` / `Password123`

## API

### Public Endpoints

- `GET /health`
- `POST /auth/sign-in`

### Authenticated Endpoints

All `/users*` endpoints require:

- Header: `Authorization: Bearer <jwt>`

Endpoints:

- `GET /users` (optional query: `email`)
- `GET /users/{id}`
- `PUT /users/{id}`

### Example: Sign In

```bash
curl -X POST http://localhost:8080/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Password123"}'
```

Response:

```json
{
  "token": "<jwt>"
}
```

### Example: List Users

```bash
curl -H "Authorization: Bearer <jwt>" \
  "http://localhost:8080/users?email=alice"
```

### Example: Get User

```bash
curl -H "Authorization: Bearer <jwt>" \
  http://localhost:8080/users/1
```

### Example: Update User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Cooper"}'
```

## Web UI

The UI supports:

- Sign in and token session handling
- List/search users by email
- Load a user by ID
- Update a user's name
- Inspect JSON responses

## Run Tests

```bash
go test ./...
```

## Troubleshooting

### Port already in use

- Local run fails on `:8080`: set a different port, for example:

```bash
PORT=8090 go run ./cmd/server
```

- Docker run fails on `:8081`: change host mapping in `docker-compose.yml` from `8081:8080` to another free host port.

### PostgreSQL connection issues

- For local app + local DB, verify `DATABASE_URL` points to a reachable database.
- If running with Docker Compose, start services with:

```bash
docker compose up --build -d
```

- Check container status/logs:

```bash
docker compose ps
docker compose logs app postgres
```

### Unauthorized on protected endpoints

- Call `POST /auth/sign-in` first and pass the returned token as:

```bash
Authorization: Bearer <jwt>
```

- Ensure there are no extra spaces or missing `Bearer` prefix.
