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
| `feat/submissions` | #40 | Submission store + retrieval handlers + submission UI |
| `feat/sandbox` | #41 | Goroutine sandbox worker + cross-platform executor |
| `feat/enrollment-gate` | #42 | Enrollment enforcement for exercises, test cases, submissions |
| `feat/dashboard` | #43 | Student dashboard (stat cards + recent submissions) |
| `feat/course-stats` | #44 | Professor course report (staff, student roster, solve rates) |
| `fix/code-review-bugs` | #44 | 8 code-review findings fixed (DB errors, enrollment logic, etc.) |
| `feat/gamified-frontend` | #45 | Gamified UI: XP bars, streaks, learning path nodes, role-adaptive dashboard, split-pane exercise editor |
| `fix/security-vulnerabilities` | #46 #48 | 8 security fixes: role whitelist, CORS allowlist, IDOR on test cases, course ownership, body size cap, sandbox limits |
| `feWork` | #47 | Register page layout fix |
| `feat/admin-panel` | #50 | Admin UI: user list, role/active toggle, course overview, platform stats |
| `fix/minor-changes` | #51 | UI fixes, role flows, minor corrections |
| `feature/fix-TA-role` | #52 | TA can create/edit exercises in courses they're enrolled in |
| `feature/quizz` | #53 | Quiz exercise type: multiple-choice questions with options, correct answers, allow-multiple flag |
| `feature/UI` | #54 | Dark mode, UI repolish, leaderboard tab on course page |

### In progress / pending

| Branch | Status | Notes |
|---|---|---|
| `feat/ta-enrollment-controls` | Local only — not pushed | Remove TA from register, enforce course-level TA access, professor promote/demote UI in members table |
| `fix/code-review-findings` | Local only — not pushed | `isEnrolledAsTA` fix, `MaxBytesReader` on exercise handlers, unpublished visibility fixes |

### Upcoming features

| Branch (suggested) | Feature | Description |
|---|---|---|
| `feat/notifications` | Real-time status updates | Poll or SSE so the submission result banner auto-refreshes when sandbox finishes. |
| `feat/plagiarism` | Similarity check | Flag suspicious submissions within a course. |

---

## Repository layout

```
backend/              Go HTTP API (stdlib net/http, port 8080)
  cmd/server/         main.go — entry point, route registration
  internal/
    auth/             JWT, bcrypt, middleware, role constants
    courses/          Course CRUD + enrollment + members + leaderboard handlers
    exercises/        Exercise CRUD + test case handlers (coding + quiz types)
    submissions/      Submission store + retrieval handlers
    dashboard/        Student/professor dashboard stats handler
    coursestats/      Professor course report handler
    admin/            Admin panel: user list, role/active toggle, course overview, platform stats
    sandbox/          Goroutine sandbox worker + cross-platform code executor
    db/               pgxpool singleton (db.Pool)
    httputil/         WriteJSON, Error, CORS middleware
frontend/             SvelteKit 2 + Svelte 5 + TypeScript (port 5173 dev)
  src/
    lib/
      api/            Typed fetch wrappers
        client.ts     apiFetch<T> — attaches Bearer token, throws ApiError
        auth.ts       login(), register()
        courses.ts    listCourses, getCourse, createCourse, updateCourse,
                      deleteCourse, enrollCourse, getCourseMembers,
                      getCourseLeaderboard, updateMemberRole (feat/ta branch)
        exercises.ts  listExercises, getExercise, createExercise, updateExercise,
                      deleteExercise, listTestCases, createTestCase, deleteTestCase
        submissions.ts submitCode, listSubmissions, getSubmission
        dashboard.ts  getDashboard
        coursestats.ts getCourseStats
        admin.ts      getAdminStats, listAdminUsers, listAdminCourses,
                      updateUserRole, toggleUserActive
      components/
        ExerciseForm.svelte      Reusable create/edit exercise form (coding + quiz modes)
        DifficultyDot.svelte     Colored dot for easy/medium/hard difficulty
        XpBar.svelte             Level + XP progress bar (reads userStore)
        StatChip.svelte          Colored stat card with icon (enrolled/solved/etc.)
        ProgressRing.svelte      SVG circular progress ring with label
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
        register/     Register form — student / professor only (TA removed in feat/ta branch)
      admin/
        +page.svelte  Admin panel: platform stats, user management (role + active toggle), course overview
      dashboard/
        +page.svelte  Role-adaptive: student (XP bar + stats + learning path preview)
                      vs professor (course management) vs admin/TA
      courses/
        +page.svelte              Course grid with colored strips, filter tabs (All/Enrolled/Available)
        [id]/
          +page.svelte            Course hero + progress ring + learning path nodes
                                  Members tab (professor/TA/admin) + Leaderboard tab
                                  Promote/Demote buttons for professors (feat/ta branch)
          stats/
            +page.svelte          Professor course report (enrollment, solve rates)
          exercises/new/
            +page.svelte          Create exercise (professor/admin/enrolled-TA)
      exercises/
        [id]/
          +page.svelte            Split-pane: description left, dark code editor right
                                  Quiz mode: radio/checkbox answer UI
                                  Submit+poll flow, XP float animation, test dot grid
          edit/
            +page.svelte          Edit exercise (professor/admin/enrolled-TA)
      submissions/
        [id]/
          +page.svelte            Submission detail: status, score, per-test-case results
      sandbox/
        +page.svelte  Split-pane scratchpad: notes left, dark code editor right
                      Language starters, localStorage persistence, no backend needed
infra/migrations/     Numbered SQL migration files (applied by Docker on first start)
  001–005             Core schema (users, courses, enrollments, exercises, submissions)
  006                 Add exercise_type enum + quiz_options JSONB + quiz_correct SMALLINT
  007                 Convert quiz_correct to JSONB array, add quiz_allow_multiple BOOLEAN
docker-compose.yml    PostgreSQL 16-alpine; auto-applies migrations on first start
.env                  Local secrets — never committed (see .env.example)
CLAUDE.md             This file
```

---

## Security

### Fixed vulnerabilities (merged to main)

| Severity | File | What was fixed |
|---|---|---|
| **Critical** | `auth/handler.go` | Role whitelist on register — only `student`, `teaching_assistant`, `professor` allowed; `admin` is rejected. |
| **High** | `httputil/cors.go` | CORS origin allowlist — only origins in `CORS_ORIGIN` env var (default `http://localhost:5173`) receive headers. |
| **High** | `exercises/handler.go` | IDOR on test cases — handlers verify caller owns the parent exercise before mutating. |
| **High** | `exercises/handler.go` | Course ownership check on exercise creation — professor must own the course. |
| **Medium** | `submissions/handler.go` | Submission body capped at 512 KB via `http.MaxBytesReader`. |
| **Medium** | `sandbox/executor.go` | `timeLimitMs` ceiling of 15 000 ms. |
| **Medium** | `sandbox/executor.go` | Temp source files created with mode `0600`. |

### Remaining / won't-fix at application layer

| Issue | Reason |
|---|---|
| Sandbox process/network/memory isolation | Requires Docker-in-Docker or seccomp — not addressable in Go application code. |
| JWT stored in `localStorage` (XSS risk) | Acceptable risk for current threat model. |
| No rate limiting on auth endpoints | Planned for `feat/rate-limiting`. |
| No token revocation on role change | Requires a token blocklist (Redis). |

### Adding a new endpoint — security checklist

1. Is it behind `auth.Middleware`? (all non-public routes must be)
2. Does it call `auth.RequireRole` if only certain roles should access it?
3. If it mutates a resource owned by a specific user, does it verify `owner == userID || role == admin`?
4. Does it call `http.MaxBytesReader` before decoding a request body with user-supplied text?
5. Are all DB queries using parameterized `$1`-style placeholders (never string concatenation)?

---

## Environment setup

The `.env` file lives at the repo root. **Do not rename or move it** — the backend loads it automatically via `godotenv.Load("../.env")` (relative to `backend/`).

```powershell
Copy-Item .env.example .env
docker compose up -d
```

To wipe and re-apply migrations from scratch:
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

### Creating an admin user

The register page only allows student and professor roles. Create an admin via the API directly:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/auth/register -Method Post `
  -ContentType "application/json" `
  -Body '{"email":"admin@example.com","password":"yourpassword","first_name":"Admin","last_name":"User","role":"admin"}'
```

Once logged in as admin, use the Admin panel (`/admin`) to manage all users and promote roles.

---

## Known issues and troubleshooting

### `bind: Only one usage of each socket address` (port 8080 already in use)

```powershell
Stop-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess -Force
```

### `fatal: DATABASE_URL environment variable is not set`

- `.env` does not exist — copy from `.env.example`
- You are running `go run` from a directory other than `backend/`
- Docker is not running — start with `docker compose up -d`

### `db: retrying... (attempt N)` — PostgreSQL not ready

```powershell
docker ps                     # is Docker running?
docker compose ps             # is the db container healthy?
docker compose restart db     # restart the container
```

### Frontend shows blank page or fails to load

- Backend must be running: `curl http://localhost:8080/health` → `ok`
- The frontend is a pure SPA (`ssr = false`) — all data fetching is client-side

### 500 on `GET /courses/{id}/exercises`

The `exercises` table SELECT includes `quiz_allow_multiple` (added in migration 007). If the DB volume predates migration 007, every exercise query fails. Fix by recreating the volume:
```powershell
docker compose down -v; docker compose up -d
```

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

| Endpoint | Method | Roles | Notes |
|---|---|---|---|
| `/courses` | GET | any authenticated | |
| `/courses` | POST | professor, admin | |
| `/courses/{id}` | GET | any authenticated | |
| `/courses/{id}` | PUT | professor (owner), admin | |
| `/courses/{id}` | DELETE | professor (owner), admin | |
| `/courses/{id}/enroll` | POST | student, teaching_assistant | TA defaults to enrollment_role=teaching_assistant |
| `/courses/{id}/members` | GET | professor, teaching_assistant, admin | Returns enrollment role (not global role) |
| `/courses/{id}/leaderboard` | GET | any authenticated | Ranks students by accepted exercise count |
| `/courses/{id}/members/{userId}/role` | PATCH | professor (owner), admin | Promote/demote enrollment role — in `feat/ta-enrollment-controls` |

### Exercises package (`internal/exercises`)

| Endpoint | Method | Roles | Notes |
|---|---|---|---|
| `/courses/{id}/exercises` | GET | any authenticated | |
| `/courses/{id}/exercises` | POST | professor, teaching_assistant, admin | TA write-checked by `canModifyInCourse` |
| `/exercises/{id}` | GET | any authenticated | |
| `/exercises/{id}` | PUT | professor, teaching_assistant, admin | TA write-checked by `canModifyInCourse` |
| `/exercises/{id}` | DELETE | professor, teaching_assistant, admin | TA write-checked by `canModifyInCourse` |
| `/exercises/{id}/test-cases` | GET | any authenticated | Hidden cases stripped for students |
| `/exercises/{id}/test-cases` | POST | professor, teaching_assistant, admin | |
| `/test-cases/{id}` | PUT | professor, teaching_assistant, admin | |
| `/test-cases/{id}` | DELETE | professor, teaching_assistant, admin | |

**TA write access:** `canModifyInCourse` in `exercises/handler.go` checks:
- Admin → always allowed
- Professor → must be the resource owner (`ownerID == userID`)
- TA → must be enrolled in that course with `enrollment_role = 'teaching_assistant'` (via `isEnrolledAsTA`)

**Exercise types:** `exercise_type` is `'coding'` or `'quiz'`.
- Coding: requires `language` field; has `template_code`, `time_limit_ms`, `memory_limit_kb`, test cases
- Quiz: has `quiz_options []string`, `quiz_correct []int`, `quiz_allow_multiple bool`; no test cases

### Submissions package (`internal/submissions`)

| Endpoint | Method | Roles | Notes |
|---|---|---|---|
| `/exercises/{id}/submit` | POST | any authenticated | Validates exercise exists + published; stores with `pending` status |
| `/exercises/{id}/submissions` | GET | any authenticated | Students see own only (LIMIT 50); privileged see all (LIMIT 200) |
| `/submissions/{id}` | GET | any authenticated | Students 404 on others' submissions; returns submission + `results[]` array |

### Admin package (`internal/admin`)

| Endpoint | Method | Notes |
|---|---|---|
| `/admin/stats` | GET | Platform totals: users, courses, exercises, submissions, active users, published courses |
| `/admin/users` | GET | All users with email, role, is_active, created_at |
| `/admin/users/{id}/role` | PATCH | Change any user's global role |
| `/admin/users/{id}/active` | PATCH | Toggle is_active (disables login for deactivated users) |
| `/admin/courses` | GET | All courses with creator name, enrollment count, exercise count |

### Dashboard package (`internal/dashboard`)

`GET /dashboard` — role-adaptive stats response. Students get XP/streak/solved counts. Professors get course/enrollment counts.

### Course stats package (`internal/coursestats`)

`GET /courses/{id}/stats` — professor/TA/admin only. Returns per-student solve rates, exercise difficulty breakdown, enrollment list with scores.

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
**Theme:** dark mode supported via CSS variables; toggled in `+layout.svelte`

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
| `users` | `id UUID PK`, `email UNIQUE`, `password_hash`, `first_name`, `last_name`, `role user_role`, `is_active BOOL` |
| `courses` | `id UUID PK`, `created_by → users.id`, `title`, `description`, `is_published BOOL` |
| `course_enrollments` | `(user_id, course_id) PK`, `role enrollment_role` |
| `exercises` | `id UUID PK`, `course_id → courses`, `created_by → users`, `title`, `description`, `instructions`, `difficulty difficulty_level`, `language prog_language` (nullable), `template_code`, `time_limit_ms INT`, `memory_limit_kb INT`, `is_published BOOL`, `exercise_type exercise_type`, `quiz_options JSONB`, `quiz_correct JSONB`, `quiz_allow_multiple BOOL` |
| `test_cases` | `id UUID PK`, `exercise_id → exercises`, `input TEXT`, `expected_output TEXT`, `is_hidden BOOL`, `ordinal INT` |
| `submissions` | `id UUID PK`, `exercise_id → exercises`, `user_id → users`, `code TEXT`, `language`, `status submission_status DEFAULT 'pending'`, `score FLOAT DEFAULT 0`, `stderr TEXT` |
| `submission_results` | `id UUID PK`, `submission_id → submissions`, `test_case_id → test_cases`, `status`, `actual_output TEXT`, `runtime_ms INT`, `memory_kb INT` |

**Enums:**
- `user_role`: `student | teaching_assistant | professor | admin`
- `enrollment_role`: `student | teaching_assistant`
- `exercise_type`: `coding | quiz`
- `difficulty_level`: `easy | medium | hard`
- `prog_language`: `python | go | java | c | cpp | javascript` (language is NULL for quiz exercises)
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
