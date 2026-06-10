# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

CS Educational Platform — a web app for managing CS courses, programming exercises, and automated code evaluation. Users are students, TAs, professors, and admins.

---

## Progress tracker

### Completed and merged to `main`

| Branch | PR | What was built |
|---|---|---|
| `feat/db-migrations` | #34 | PostgreSQL schema, Docker Compose, migration files |
| `feat/auth` | #35 | JWT auth, bcrypt, register/login endpoints, role middleware |
| `feat/course-api` | #36 | Course CRUD, enrollment, members endpoints |
| `feat/frontend-auth` | #37 | SvelteKit SPA, auth pages, courses UI, CORS middleware |
| `feat/exercise-api` | #38 | Exercise CRUD + test case management endpoints |
| `feat/frontend-exercises` | #39 | Exercises list/detail/create/edit UI in frontend |

### In progress (current branch: `feat/submissions`)

<<<<<<< Updated upstream
**Status: code complete, smoke-tested, NOT YET COMMITTED**

Files changed vs `main`:
- `backend/internal/submissions/model.go` — NEW
- `backend/internal/submissions/handler.go` — NEW (SubmitHandler, ListHandler, GetHandler)
- `backend/cmd/server/main.go` — 3 new submission routes registered
- `frontend/src/lib/api/submissions.ts` — NEW (submitCode, listSubmissions, getSubmission)
- `frontend/src/routes/exercises/[id]/+page.svelte` — added submission panel (code editor, result banner, history)

**To commit and push:**
```powershell
cd D:\Proiect_Colectiv2026\cs-educational-platform
git add backend/internal/submissions/ backend/cmd/server/main.go frontend/src/lib/api/submissions.ts frontend/src/routes/exercises/[id]/+page.svelte
git commit -m "feat: add code submission endpoints and submission UI on exercise page"
git push origin feat/submissions
```
Then open a PR on GitHub (`feat/submissions → main`) and merge it.
=======
| Branch | Status |
|---|---|
| `fix/security-vulnerabilities` | **Open — see Security section below** |
| `feat/gamified-frontend` | **Local (uncommitted)** — full gamified UI: XP bars, streaks, learning path nodes, role-adaptive dashboard, split-pane exercise editor |
>>>>>>> Stashed changes

### Upcoming features (in rough priority order)

| Branch (suggested) | Feature | Description |
|---|---|---|
| `feat/sandbox` | **Code execution engine** | Pick up `pending` submissions, run code in isolated Docker container, compare output against test cases, update status (`accepted` / `wrong_answer` / `runtime_error` / etc.) and per-test `submission_results` rows. This is the core value-add of the platform. |
| `feat/dashboard` | Student dashboard | Personal stats: submissions count, acceptance rate, exercises attempted, recent activity. |
| `feat/leaderboard` | Course leaderboard | Rank students in a course by score / accepted exercises. |
| `feat/notifications` | Real-time status updates | Poll or SSE so the submission result banner auto-refreshes when sandbox finishes (instead of showing "pending" forever). |
| `feat/plagiarism` | Similarity check | Flag suspicious submissions within a course. |
| `feat/admin-panel` | Admin management UI | User list, role changes, course overview. |

---

## Repository layout

```
backend/              Go HTTP API (stdlib net/http, port 8080)
  cmd/server/         main.go — entry point, route registration
  internal/
    auth/             JWT, bcrypt, middleware, role constants
    courses/          Course CRUD + enrollment handlers
    exercises/        Exercise CRUD + test case handlers
    submissions/      Submission store + retrieval handlers
    db/               pgxpool singleton (db.Pool)
    httputil/         WriteJSON, Error, CORS middleware
frontend/             SvelteKit 2 + Svelte 5 + TypeScript (port 5173 dev)
  src/
    lib/
      api/            Typed fetch wrappers
        client.ts     apiFetch<T> — attaches Bearer token, throws ApiError
        auth.ts       login(), register()
        courses.ts    listCourses, getCourse, createCourse, updateCourse,
                      deleteCourse, enrollCourse, getCourseMembers
        exercises.ts  listExercises, getExercise, createExercise, updateExercise,
                      deleteExercise, listTestCases, createTestCase, deleteTestCase
        submissions.ts submitCode, listSubmissions, getSubmission
      components/
        ExerciseForm.svelte   Reusable create/edit exercise form
        DifficultyDot.svelte  Colored dot for easy/medium/hard difficulty
        XpBar.svelte          Level + XP progress bar (reads userStore)
        StatChip.svelte       Colored stat card with icon (enrolled/solved/etc.)
        ProgressRing.svelte   SVG circular progress ring with label
        LearningPathNode.svelte  done/current/locked exercise node with pulse animation
      stores/
        auth.svelte.ts        Svelte 5 $state auth singleton (localStorage-backed)
        userStore.svelte.ts   Gamification store — XP, streak, level, solvedExercises
                              Persisted to localStorage key `cp_gamification`
                              Level = floor(xp/500)+1; 100 XP per easy/medium, 200 XP hard
    routes/           File-based SvelteKit routing
      +layout.ts      export const ssr = false (pure SPA)
      +layout.svelte  App shell: nav with XP chip + streak badge + avatar, auth guard
      +page.svelte    Root redirect → /dashboard or /auth/login
      auth/
        login/        Login form
        register/     Register form (student / TA / professor role selector)
      dashboard/
        +page.svelte  Role-adaptive: student (XP bar + stats + learning path preview)
                      vs professor (course management) vs admin/TA
      courses/
        +page.svelte              Course grid with colored strips, filter tabs (All/Enrolled/Available)
        [id]/
          +page.svelte            Course hero + progress ring + learning path nodes + members table
          exercises/new/
            +page.svelte          Create exercise (professor/admin only)
      exercises/
        [id]/
          +page.svelte            Split-pane: description left, dark code editor right
                                  Submit+poll flow, XP float animation, test dot grid
          edit/
            +page.svelte          Edit exercise (professor/admin only)
      sandbox/
        +page.svelte  Split-pane scratchpad: notes left, dark code editor right
                      Language starters, localStorage persistence, no backend needed
infra/migrations/     Numbered SQL migration files (applied by Docker on first start)
sandbox/              Code execution engine (not yet implemented)
docker-compose.yml    PostgreSQL 16-alpine; auto-applies migrations on first start
.env                  Local secrets — never committed (see .env.example)
CLAUDE.md             This file
```

---

## Environment setup

The `.env` file lives at the repo root. **Do not rename or move it** — the backend loads it automatically via `godotenv.Load("../.env")` (relative to `backend/`).

To recreate it:
```powershell
Copy-Item .env.example .env
```

Default values in `.env.example` work out of the box with Docker Compose.

Start the database (required before starting the backend):
```powershell
docker compose up -d
```

The `initdb.d` mount applies migrations automatically the **first** time the volume is created.  
To wipe and re-apply from scratch:
```powershell
docker compose down -v; docker compose up -d
```

---

## Running the app

Open **three terminals**, each from the repo root:

```powershell
# Terminal 1 — database (keep running)
docker compose up -d

# Terminal 2 — backend API
cd backend; go run ./cmd/server

# Terminal 3 — frontend
cd frontend; npm run dev
```

- Backend: http://localhost:8080
- Frontend: http://localhost:5173

### Registering an admin user (PowerShell)

The register page only exposes student/TA/professor roles. To create an admin via the API:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/auth/register -Method Post `
  -ContentType "application/json" `
  -Body '{"email":"admin@test.com","password":"secret123","first_name":"Admin","last_name":"User","role":"admin"}'
```

---

## Known issues and troubleshooting

### `bind: Only one usage of each socket address` (port 8080 already in use)

A previous server process was not shut down cleanly.

```powershell
Stop-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess -Force
```

Or find the PID manually:
```powershell
netstat -ano | findstr :8080
Stop-Process -Id <PID> -Force
```

### `fatal: DATABASE_URL environment variable is not set`

The server auto-loads `../.env` (relative to `backend/`) via `godotenv`. This error means:
- `.env` does not exist — copy from `.env.example`
- You are running `go run` from a directory other than `backend/`
- Docker is not running — start with `docker compose up -d`

### `db: retrying... (attempt N)` — PostgreSQL not ready

The DB pool retries 10 × 2 s. If it exhausts retries:
```powershell
docker ps                     # is Docker running?
docker compose ps             # is the db container healthy?
docker compose restart db     # restart the container
```

### Frontend shows blank page or fails to load

- Backend must be running: `curl http://localhost:8080/health` → `ok`
- CORS is handled by `httputil.CORS` middleware — no proxy needed
- The frontend is a pure SPA (`ssr = false`) — all data fetching is client-side

### Submission stays "Pending" forever

Expected — the sandbox (feat/sandbox) has not been built yet. Submissions are stored with status `pending` and will remain there until the execution engine is implemented.

---

## Backend

**Module:** `cs-educational-platform/backend`  
**Go version:** 1.26.2  
**Key dependencies:** `pgx/v5`, `golang-jwt/jwt/v5`, `golang.org/x/crypto`, `godotenv`

```powershell
cd backend

go run ./cmd/server          # run
go build ./cmd/server        # compile binary
go test ./...                # all tests
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

| Endpoint | Method | Roles |
|---|---|---|
| `/courses` | GET | any authenticated |
| `/courses` | POST | professor, admin |
| `/courses/{id}` | GET | any authenticated |
| `/courses/{id}` | PUT | professor (owner), admin |
| `/courses/{id}` | DELETE | professor (owner), admin |
| `/courses/{id}/enroll` | POST | student, teaching_assistant |
| `/courses/{id}/members` | GET | professor, teaching_assistant, admin |

### Exercises package (`internal/exercises`)

| Endpoint | Method | Roles |
|---|---|---|
| `/courses/{id}/exercises` | GET | any authenticated |
| `/courses/{id}/exercises` | POST | professor, admin |
| `/exercises/{id}` | GET | any authenticated |
| `/exercises/{id}` | PUT | professor, admin |
| `/exercises/{id}` | DELETE | professor, admin |
| `/exercises/{id}/test-cases` | GET | any authenticated (hidden cases stripped for students) |
| `/exercises/{id}/test-cases` | POST | professor, admin |
| `/test-cases/{id}` | PUT | professor, admin |
| `/test-cases/{id}` | DELETE | professor, admin |

Hidden test case handling: students receive `HiddenTestCase` (no `input`/`expected_output`); professors/TAs/admins receive full `TestCase`.

### Submissions package (`internal/submissions`)

| Endpoint | Method | Roles | Notes |
|---|---|---|---|
| `/exercises/{id}/submit` | POST | any authenticated | Validates exercise exists + published; stores with `pending` status |
| `/exercises/{id}/submissions` | GET | any authenticated | Students see own only (LIMIT 50); privileged see all (LIMIT 200) |
| `/submissions/{id}` | GET | any authenticated | Students 404 on others' submissions; returns submission + `results[]` array |

### Httputil package (`internal/httputil`)

| Function/File | Purpose |
|---|---|
| `WriteJSON(w, status, v)` | Marshal → set Content-Type → write status + body |
| `Error(w, msg, code)` | Write `{"error": msg}` JSON response |
| `CORS(next)` | Wrap handler with Access-Control headers + OPTIONS preflight |

### Database package (`internal/db`)

`db.Pool` is a `*pgxpool.Pool` initialised in `main.go` via `db.Connect(ctx)`. All packages import and use `db.Pool` directly. The pool retries up to 10 times (2 s apart) to tolerate slow container startup.

### Coding patterns

**COALESCE for partial updates** (only update fields that are provided):
```go
type updateRequest struct {
    Title *string `json:"title"` // pointer = optional
}
// In SQL:
// title = COALESCE($1, title)
```

**scanX helper** — keeps column order in one place:
```go
func scanCourse(c *Course, scan func(...any) error) error {
    return scan(&c.ID, &c.Title, /* ... */)
}
// Usage: scanCourse(&c, row.Scan)
```

**ErrNoRows → 404**:
```go
if errors.Is(err, pgx.ErrNoRows) {
    httputil.Error(w, "not found", http.StatusNotFound)
    return
}
```

**Delete 404 via RowsAffected**:
```go
tag, err := db.Pool.Exec(ctx, `DELETE FROM ... WHERE id = $1`, id)
if tag.RowsAffected() == 0 {
    httputil.Error(w, "not found", http.StatusNotFound)
    return
}
```

### Adding a new feature package

1. Create `backend/internal/<feature>/model.go` and `handler.go`
2. Register routes in `backend/cmd/server/main.go` (copy the submissions block as a template)
3. Wrap with `auth.Middleware` and `auth.RequireRole` as appropriate
4. Add any new tables as a new numbered migration file in `infra/migrations/`
5. Run `docker compose down -v && docker compose up -d` to apply new migrations

---

## Frontend

**Stack:** SvelteKit 2 · Svelte 5 (runes) · TypeScript · Vite 6  
**SSR:** disabled (`ssr = false` in `+layout.ts`) — pure SPA, all rendering in the browser

```powershell
cd frontend

npm install          # first time only
npm run dev          # dev server → http://localhost:5173
npm run build        # production build
npm run check        # TypeScript + Svelte type-check (run before committing)
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
let items   = $state<Item[]>([]);
const canManage = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');

// untrack() to silence state_referenced_locally warning when initialising
// $state from a prop (ExerciseForm pattern):
import { untrack } from 'svelte';
let title = $state(untrack(() => initial.title ?? ''));
```

### Auth flow

1. User submits login/register form → API call → receives `{ token, user_id, role, first_name, last_name }`
2. `auth.set(token, user)` stores both in `$state` and `localStorage`
3. `+layout.svelte` calls `auth.init()` on mount to restore session from `localStorage`
4. `apiFetch` reads `localStorage.getItem('token')` and attaches `Authorization: Bearer <token>`
5. Logout: `auth.logout()` clears state + localStorage, redirects to `/auth/login`

### API client pattern

```ts
// lib/api/client.ts
export class ApiError extends Error {
  constructor(public readonly status: number, message: string) { super(message); }
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token');
  const res = await fetch(`http://localhost:8080${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json',
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
                ...options.headers },
  });
  if (!res.ok) { const e = await res.json().catch(() => ({})); throw new ApiError(res.status, e.error ?? 'request failed'); }
  if (res.status === 204) return undefined as T;
  return res.json();
}
```

---

## Database schema

Migrations in `infra/migrations/` must be numbered sequentially (`001_`, `002_`, …).

| Table | Key columns |
|---|---|
| `users` | `id UUID PK`, `email UNIQUE`, `password_hash`, `first_name`, `last_name`, `role user_role` |
| `courses` | `id UUID PK`, `created_by → users.id`, `title`, `description`, `is_published BOOL` |
| `course_enrollments` | `(user_id, course_id) PK`, `role enrollment_role` |
| `exercises` | `id UUID PK`, `course_id → courses`, `created_by → users`, `title`, `description`, `instructions`, `difficulty difficulty_level`, `language prog_language`, `template_code`, `time_limit_ms INT`, `memory_limit_kb INT`, `is_published BOOL` |
| `test_cases` | `id UUID PK`, `exercise_id → exercises`, `input TEXT`, `expected_output TEXT`, `is_hidden BOOL`, `ordinal INT` |
| `submissions` | `id UUID PK`, `exercise_id → exercises`, `user_id → users`, `code TEXT`, `language`, `status submission_status DEFAULT 'pending'`, `score FLOAT DEFAULT 0`, `stderr TEXT` |
| `submission_results` | `id UUID PK`, `submission_id → submissions`, `test_case_id → test_cases`, `status`, `actual_output TEXT`, `runtime_ms INT`, `memory_kb INT` |

**Enums:**
- `user_role`: `student | teaching_assistant | professor | admin`
- `difficulty_level`: `easy | medium | hard`
- `prog_language`: `python | go | java | c | cpp | javascript`
- `submission_status`: `pending | running | accepted | wrong_answer | runtime_error | time_limit | memory_limit | compile_error`

`set_updated_at()` trigger is applied to `users`, `courses`, `exercises`.

---

## Git workflow

`main` is always stable. One branch per feature:

```powershell
git checkout main; git pull origin main
git checkout -b feat/<name>        # or fix/<name>
# … implement → npm run check → go build ./... → commit …
git push origin feat/<name>
# open PR on GitHub → review → squash-merge → delete remote branch

# After merge, clean up locally:
git checkout main; git pull origin main
git branch -d feat/<name>
```

Commit message format: `type: short description`  
Types: `feat` | `fix` | `chore` | `docs` | `refactor` | `test`

**Pre-commit checklist:**
1. `cd backend && go build ./...` — must compile with 0 errors
2. `cd frontend && npm run check` — must report 0 errors, 0 warnings
3. Run a quick smoke test of the new endpoints (curl / Invoke-RestMethod)
4. Stage only the files changed for this feature (avoid `git add .`)
