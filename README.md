# User Management API (Golang)

Backend web application with JWT-based authentication and user management APIs.

## Features

- Sign in with email and password
- Get user by ID
- Update user
- List users with optional email search
- Modern compact web UI dashboard
- JWT-protected user endpoints
- Input validation and meaningful JSON error responses

## Run

```bash
go run ./cmd/server
```

Then open:

- `http://localhost:8080/` (web UI)

## Run with Docker (recommended)

```bash
docker compose up --build -d
```

Then open:

- `http://localhost:8081/`

Stop stack:

```bash
docker compose down
```

Stop and remove DB volume:

```bash
docker compose down -v
```

Environment variables:

- `PORT` (default: `8080`)
- `JWT_SECRET` (default: `development-secret`)
- `JWT_ISSUER` (default: `interviewsw`)
- `DATABASE_URL` (optional, enables PostgreSQL persistence)

## PostgreSQL (optional)

If `DATABASE_URL` is set, the app uses PostgreSQL; otherwise it runs with in-memory storage.

Example:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/interviewsw?sslmode=disable"
go run ./cmd/server
```

On startup, schema is created automatically and seed users are inserted only when the `users` table is empty.

## API

## Public testing

The following endpoints are public and do not require a JWT token:

- `GET /health`
- `POST /auth/sign-in`

Quick test (local run on `:8080`):

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Password123"}'
```

Quick test (docker run on `:8081`):

```bash
curl http://localhost:8081/health

curl -X POST http://localhost:8081/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Password123"}'
```

### 1) Sign In

- `POST /auth/sign-in`

Request body:

```json
{
  "email": "alice@example.com",
  "password": "Password123"
}
```

Success response:

```json
{
  "token": "<jwt>"
}
```

### 2) Get User

- `GET /users/{id}`
- Header: `Authorization: Bearer <jwt>`

### 3) Update User

- `PUT /users/{id}`
- Header: `Authorization: Bearer <jwt>`

Request body:

```json
{
  "name": "Alice Cooper"
}
```

### 4) List Users

- `GET /users`
- Optional query: `email`
- Header: `Authorization: Bearer <jwt>`

Example:

```bash
curl -H "Authorization: Bearer <jwt>" "http://localhost:8080/users?email=alice"   # local run
curl -H "Authorization: Bearer <jwt>" "http://localhost:8081/users?email=alice"   # docker run
```

## Web UI

The UI provides:

- Sign in and token session state
- Search + list users
- Load/get a user by ID
- Update user name
- JSON response panel for quick debugging

## Seeded users

- `alice@example.com` / `Password123`
- `bob@example.com` / `Password123`
