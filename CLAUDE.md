# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

CS Educational Platform — a web app for managing CS courses, programming exercises, and automated code evaluation. Users are students, TAs, professors, and admins.

## Repository layout

```
backend/          Go HTTP API (stdlib net/http, port 8080)
frontend/         SvelteKit + TypeScript UI (port 5173 dev)
infra/migrations/ Numbered SQL migration files (applied in order)
sandbox/          Code execution engine (not yet implemented)
docker-compose.yml  Spins up PostgreSQL 16; auto-applies migrations on first start
```

## Environment setup

Copy `.env.example` to `.env` and fill in values before running anything:

```
DATABASE_URL=postgres://csplatform:csplatform_password@localhost:5432/csplatform_db?sslmode=disable
JWT_SECRET=<long random string>
```

Start the database:
```bash
docker compose up -d
```

The `initdb.d` mount means migrations run automatically the **first** time the volume is created. To re-apply from scratch, destroy the volume first: `docker compose down -v`.

## Backend

**Module:** `cs-educational-platform/backend`  
**Go version:** 1.26.2  
**Key dependencies:** `pgx/v5` (Postgres), `golang-jwt/jwt/v5`, `golang.org/x/crypto` (bcrypt)

```bash
# Run
cd backend && go run ./cmd/server

# Build
cd backend && go build ./cmd/server

# Test
cd backend && go test ./...

# Single package test
cd backend && go test ./internal/auth/...

# Tidy modules after adding a dependency
cd backend && go mod tidy
```

The server exits on startup if `DATABASE_URL` or `JWT_SECRET` are unset.

## Frontend

```bash
cd frontend && npm install   # first time only
cd frontend && npm run dev   # dev server on http://localhost:5173
cd frontend && npm run build
cd frontend && npm run check # TypeScript + Svelte type-check
```

## Architecture

### Request flow (backend)

```
HTTP request
  → main.go mux (method+path routing, e.g. "POST /auth/login")
  → auth.Middleware (validates Bearer JWT, injects user_id + role into context)
  → auth.RequireRole(...) (optional role guard)
  → handler function
  → db.Pool (pgxpool, shared singleton in internal/db)
```

### Auth package (`internal/auth`)

| File | Responsibility |
|---|---|
| `handler.go` | `POST /auth/register`, `POST /auth/login` — return JWT + user profile |
| `jwt.go` | HS256 token generation/parsing, 24 h TTL, reads `JWT_SECRET` from env |
| `middleware.go` | `Middleware` (Bearer validation) + `RequireRole(roles...)` guard + `UserIDFromCtx` / `RoleFromCtx` helpers |
| `password.go` | bcrypt helpers at cost 12 |

Protecting a route:
```go
mux.Handle("POST /courses", auth.Middleware(
    auth.RequireRole("professor", "admin")(http.HandlerFunc(courses.CreateHandler)),
))
```

### Database package (`internal/db`)

`db.Pool` is a `*pgxpool.Pool` initialised in `main.go` via `db.Connect(ctx)`. All packages import and use `db.Pool` directly. The pool retries up to 10 times (2 s apart) to tolerate slow container startup.

### Database schema

Migrations in `infra/migrations/` must be numbered sequentially (`001_`, `002_`, …). Current tables:

| Table | Key relationships |
|---|---|
| `users` | root entity; `user_role` enum: student / teaching_assistant / professor / admin |
| `courses` | `created_by → users.id` |
| `course_enrollments` | `(user_id, course_id)` PK; `enrollment_role` enum |
| `exercises` | `course_id → courses`, `created_by → users`; has `time_limit_ms`, `memory_limit_kb` |
| `test_cases` | `exercise_id → exercises`; `is_hidden` hides from students |
| `submissions` | `exercise_id + user_id`; `submission_status` enum tracks sandbox result |
| `submission_results` | per-test-case breakdown of each submission |

`set_updated_at()` trigger function is applied to `users`, `courses`, and `exercises`.

### Adding a new feature package

1. Create `backend/internal/<feature>/handler.go` (and other files as needed).
2. Register routes in `backend/cmd/server/main.go`.
3. Wrap with `auth.Middleware` and `auth.RequireRole` as appropriate.
4. Add any new tables as a new numbered migration file in `infra/migrations/`.

### Frontend (`frontend/src`)

- `lib/api/` — backend API client functions (fetch wrappers)
- `lib/components/` — reusable Svelte components
- `routes/` — SvelteKit file-based routing (`+page.svelte`, `+layout.svelte`, etc.)

## Git workflow

`main` is always stable. One branch per feature:

```bash
git checkout main && git pull origin main
git checkout -b feat/<name>        # or fix/<name>
# … commit small and often …
git push origin feat/<name>
# open PR → review → merge → delete branch
```

Commit message format: `type: short description`  
Types: `feat` | `fix` | `chore` | `docs` | `refactor` | `test`
