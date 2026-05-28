# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

CS Educational Platform — a web app for managing CS courses, programming exercises, and automated code evaluation. Users are students, TAs, professors, and admins.

## Repository layout

```
backend/              Go HTTP API (stdlib net/http, port 8080)
  cmd/server/         main.go — entry point, route registration
  internal/
    auth/             JWT, bcrypt, middleware, role constants
    courses/          Course CRUD + enrollment handlers
    db/               pgxpool singleton (db.Pool)
    httputil/         WriteJSON, Error, CORS middleware
frontend/             SvelteKit 2 + Svelte 5 + TypeScript (port 5173 dev)
  src/
    lib/api/          Typed fetch wrappers: client.ts, auth.ts, courses.ts
    lib/stores/       auth.svelte.ts — Svelte 5 $state auth singleton
    routes/           File-based routing (+page.svelte, +layout.svelte)
infra/migrations/     Numbered SQL migration files (applied in order by Docker)
sandbox/              Code execution engine (not yet implemented)
docker-compose.yml    PostgreSQL 16-alpine; auto-applies migrations on first start
.env                  Local secrets — never committed (see .env.example)
```

## Environment setup

The `.env` file already exists at the repo root. **Do not rename or move it** — the backend loads it automatically via `godotenv.Load("../.env")` (relative to `backend/`).

If you need to recreate it, copy `.env.example`:
```bash
cp .env.example .env
```

Default values in `.env.example` work out of the box with Docker Compose.

Start the database (required before starting the backend):
```bash
docker compose up -d
```

The `initdb.d` mount applies migrations automatically the **first** time the volume is created.  
To wipe and re-apply from scratch: `docker compose down -v && docker compose up -d`

---

## Running the app

Open **three terminals**, each from the repo root:

```bash
# Terminal 1 — database (keep running)
docker compose up -d

# Terminal 2 — backend API
cd backend && go run ./cmd/server

# Terminal 3 — frontend
cd frontend && npm run dev
```

- Backend: http://localhost:8080
- Frontend: http://localhost:5173

---

## Known issues and troubleshooting

### `bind: Only one usage of each socket address` (port 8080 already in use)

This happens when a previous server process was not shut down cleanly (e.g. after Ctrl+C or a crashed terminal session).

**Fix — run this in PowerShell, then retry `go run`:**
```powershell
Stop-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess -Force
```

Or find the PID manually and kill it:
```powershell
netstat -ano | findstr :8080
# note the PID in the last column, then:
Stop-Process -Id <PID> -Force
```

### `fatal: DATABASE_URL environment variable is not set`

The server auto-loads `../.env` (relative to `backend/`) via `godotenv`. This error means either:
- `.env` does not exist — copy it from `.env.example`
- You are running `go run` from a directory other than `backend/` — always run from `backend/`
- Docker is not running — start it with `docker compose up -d`

### `db: retrying... (attempt N)` — PostgreSQL not ready

The DB pool retries 10 times with 2 s gaps. If it exhausts all retries:
- Check Docker is running: `docker ps`
- Check the container is healthy: `docker compose ps`
- Restart the container: `docker compose restart db`

### Frontend shows blank page or fails to load courses

- Confirm the backend is running: `curl http://localhost:8080/health` should return `ok`
- CORS is handled by the `httputil.CORS` middleware — if you see CORS errors in the browser console, make sure you are running the backend (not just the frontend)
- The frontend is a pure SPA (`ssr = false`) — it requires the backend to be up

---

## Backend

**Module:** `cs-educational-platform/backend`  
**Go version:** 1.26.2  
**Key dependencies:** `pgx/v5`, `golang-jwt/jwt/v5`, `golang.org/x/crypto`, `godotenv`

```bash
cd backend

go run ./cmd/server          # run
go build ./cmd/server        # compile binary
go test ./...                # all tests
go test ./internal/auth/...  # single package
go mod tidy                  # after adding/removing dependencies
```

### Request flow

```
HTTP request
  → httputil.CORS (sets Access-Control-* headers, handles OPTIONS preflight)
  → main.go mux  (Go 1.22+ method+path routing, e.g. "POST /auth/login")
  → auth.Middleware (validates Bearer JWT → injects user_id + role into context)
  → auth.RequireRole(...) (optional role guard, must come after Middleware)
  → handler function
  → db.Pool (pgxpool singleton)
```

### Auth package (`internal/auth`)

| File | Responsibility |
|---|---|
| `handler.go` | `POST /auth/register`, `POST /auth/login` → returns `{token, user_id, role, first_name, last_name}` |
| `jwt.go` | HS256 token generation/parsing, 24 h TTL, `JWT_SECRET` cached at startup via `Init()` |
| `middleware.go` | `Middleware` (Bearer validation) · `RequireRole(roles...)` guard · `UserIDFromCtx` / `RoleFromCtx` helpers |
| `password.go` | bcrypt helpers, cost 12 |
| `roles.go` | Role constants: `RoleStudent`, `RoleTeachingAssistant`, `RoleProfessor`, `RoleAdmin` |

Protecting a route:
```go
mux.Handle("POST /courses", auth.Middleware(
    auth.RequireRole(auth.RoleProfessor, auth.RoleAdmin)(
        http.HandlerFunc(courses.CreateHandler),
    ),
))
```

### Courses package (`internal/courses`)

| File | Responsibility |
|---|---|
| `model.go` | `Course`, `Member`, `createRequest`, `updateRequest`, `enrollRequest` types |
| `handler.go` | All 7 course handlers; `courseFields` const; `courseOwner` and `scanCourse` helpers |

**Endpoints:**

| Method | Path | Roles |
|---|---|---|
| GET | `/courses` | any authenticated |
| POST | `/courses` | professor, admin |
| GET | `/courses/{id}` | any authenticated |
| PUT | `/courses/{id}` | professor (owner), admin |
| DELETE | `/courses/{id}` | professor (owner), admin |
| POST | `/courses/{id}/enroll` | student, teaching_assistant |
| GET | `/courses/{id}/members` | professor, teaching_assistant, admin |

### Httputil package (`internal/httputil`)

| Function/File | Purpose |
|---|---|
| `WriteJSON(w, status, v)` | Marshal → set Content-Type → write status + body |
| `Error(w, msg, code)` | Write `{"error": msg}` JSON response |
| `CORS(next)` | Wrap handler with Access-Control headers + OPTIONS preflight |

### Database package (`internal/db`)

`db.Pool` is a `*pgxpool.Pool` initialised in `main.go` via `db.Connect(ctx)`. All packages import and use `db.Pool` directly. The pool retries up to 10 times (2 s apart) to tolerate slow container startup.

### Adding a new feature package

1. Create `backend/internal/<feature>/model.go` and `handler.go`
2. Register routes in `backend/cmd/server/main.go` (copy the courses block as a template)
3. Wrap with `auth.Middleware` and `auth.RequireRole` as appropriate
4. Add any new tables as a new numbered migration file in `infra/migrations/`
5. Run `docker compose down -v && docker compose up -d` to apply new migrations

---

## Frontend

**Stack:** SvelteKit 2 · Svelte 5 (runes) · TypeScript · Vite 6  
**SSR:** disabled (`ssr = false` in `+layout.ts`) — pure SPA, all rendering in the browser

```bash
cd frontend

npm install          # first time only
npm run dev          # dev server → http://localhost:5173
npm run build        # production build
npm run check        # TypeScript + Svelte type-check (run before committing)
```

### Frontend structure

```
src/
  lib/
    api/
      client.ts      # apiFetch<T>(path, options) — attaches Bearer token, throws ApiError
      auth.ts        # login(), register()
      courses.ts     # listCourses(), getCourse(), createCourse(), updateCourse(),
                     # deleteCourse(), enrollCourse(), getCourseMembers()
    stores/
      auth.svelte.ts # AuthStore class with $state runes — persists token/user to localStorage
  routes/
    +layout.ts       # export const ssr = false
    +layout.svelte   # App shell: sticky nav, auth guard (redirects to /auth/login if not logged in)
    +page.svelte     # Root redirect → /courses or /auth/login
    auth/
      login/         # Login form
      register/      # Register form (includes role selector)
    courses/
      +page.svelte   # Course grid; role-aware: enroll (student/TA), create/delete (prof/admin)
      [id]/
        +page.svelte # Course detail + members table (professor/TA/admin only)
```

### Svelte 5 patterns used

```ts
// Reactive state in a .svelte.ts class (shared singleton)
class AuthStore {
  token = $state<string | null>(null);
  user  = $state<User | null>(null);
  get isLoggedIn() { return this.token !== null; }
}
export const auth = new AuthStore();

// In components
let loading = $state(false);
let courses = $state<Course[]>([]);
const canManage = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');
```

### Auth flow

1. User submits login/register form → API call → receives `{ token, user_id, role, first_name, last_name }`
2. `auth.set(token, user)` stores both in `$state` and `localStorage`
3. `+layout.svelte` calls `auth.init()` on mount to restore session from `localStorage`
4. `apiFetch` reads `localStorage.getItem('token')` and attaches it as `Authorization: Bearer <token>`
5. Logout: `auth.logout()` clears state + localStorage, redirects to `/auth/login`

---

## Database schema

Migrations in `infra/migrations/` must be numbered sequentially (`001_`, `002_`, …).

| Table | Key columns / relationships |
|---|---|
| `users` | `id UUID PK`, `email UNIQUE`, `password_hash`, `first_name`, `last_name`, `role user_role` |
| `courses` | `id UUID PK`, `created_by → users.id`, `is_published BOOL` |
| `course_enrollments` | `(user_id, course_id) PK`, `role enrollment_role` |
| `exercises` | `course_id → courses`, `created_by → users`, `time_limit_ms`, `memory_limit_kb`, `difficulty difficulty_level`, `language prog_language` |
| `test_cases` | `exercise_id → exercises`, `is_hidden BOOL` |
| `submissions` | `exercise_id + user_id`, `status submission_status` |
| `submission_results` | per-test-case breakdown per submission |

**Enums:**
- `user_role`: `student | teaching_assistant | professor | admin`
- `enrollment_role`: subset of above
- `difficulty_level`: (see migration 004)
- `prog_language`: `python | go | java | c | cpp | javascript`
- `submission_status`: `pending | running | accepted | wrong_answer | runtime_error | time_limit | memory_limit | compile_error`

`set_updated_at()` trigger is applied to `users`, `courses`, `exercises`.

---

## Git workflow

`main` is always stable. One branch per feature:

```bash
git checkout main && git pull origin main
git checkout -b feat/<name>        # or fix/<name>
# … commit small and often …
git push origin feat/<name>
# open PR on GitHub → review → squash-merge → delete remote branch

# After merge, clean up locally:
git checkout main && git pull origin main
git branch -d feat/<name>
```

Commit message format: `type: short description`  
Types: `feat` | `fix` | `chore` | `docs` | `refactor` | `test`
