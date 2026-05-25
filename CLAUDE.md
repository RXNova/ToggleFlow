# ToggleFlow — Claude Instructions

## About the Developer

- Experienced with **Angular** and **NestJS** — use these as reference points when explaining concepts
- First time working with **Go** and **Vue 3**
- Always explain Go and Vue concepts by drawing parallels to Angular/NestJS equivalents
- Never assume familiarity with Go idioms, Vue patterns, or their ecosystems

### Mental model mappings to use

| Angular / NestJS | Go / Vue equivalent |
|---|---|
| NestJS module | Go package |
| NestJS service | Go struct with methods |
| NestJS controller | Fiber route handler |
| NestJS middleware | Fiber middleware |
| NestJS DTO + class-validator | Go struct with validation tags |
| NestJS TypeORM entity | Bun model struct |
| Angular component | Vue Single File Component (`.vue`) |
| Angular service + RxJS | Tanstack Query composable |
| Angular `@Input()` / `@Output()` | Vue `defineProps()` / `defineEmits()` |
| Angular `ngOnInit` | Vue `onMounted` |
| Angular Router | Vue Router |
| NgRx / Angular signals | Pinia store |
| Angular `HttpClient` | Tanstack Query `useQuery` / `useMutation` |
| RxJS Observable | Vue `ref` / `computed` |
| `*ngFor` | `v-for` |
| `*ngIf` | `v-if` |
| `[(ngModel)]` | `v-model` |
| `{{ interpolation }}` | `{{ interpolation }}` (same) |
| `[property]` binding | `:prop` (v-bind shorthand) |
| `(event)` binding | `@event` (v-on shorthand) |

---

## How Claude Should Behave

### Always ask before implementing
Before writing any code, present options with trade-offs and wait for a decision. Never assume and proceed. Example format:

> I can approach this two ways:
> - **Option A** — [what it is, Angular/NestJS parallel, trade-off]
> - **Option B** — [what it is, Angular/NestJS parallel, trade-off]
>
> Which do you prefer?

### Explain every Go and Vue concept used
When introducing a Go or Vue pattern for the first time, add a short explanation with a NestJS/Angular equivalent. Example:

> `defer` in Go runs a function when the current function returns — similar to `finally` in a try/catch block in TypeScript.

### Never write code silently
Always explain what you are about to write before writing it. After writing, explain what each significant part does.

### Flag unfamiliar patterns proactively
If you are about to use a Go or Vue pattern that has no Angular/NestJS equivalent, call it out explicitly and explain it from first principles.

---

## Project Overview

**ToggleFlow** is a self-hostable feature flag and configuration management system — a LaunchDarkly clone companies can run as a single Docker container.

### Core features (MVP)
- Feature flags with boolean, string, number, and JSON variations
- Per-environment flag state (production / staging / development)
- Targeting rules — serve a variation based on user attributes
- Percentage rollouts — consistent hash rollout to X% of users
- Real-time flag updates pushed to SDKs via SSE
- Audit log of every flag change
- Single Docker container, no external dependencies

---

## Tech Stack

### Backend

| Technology | Version | Role |
|---|---|---|
| Go | 1.23 | Language |
| Fiber v2 | latest | HTTP framework |
| Bun ORM | latest | Database ORM + migrations |
| SQLite (WAL mode) | — | Embedded database |
| Native SSE | — | Real-time flag push to SDKs |

### Frontend

| Technology | Version | Role |
|---|---|---|
| Vue 3 | latest | UI framework |
| Vite | latest | Build tool + dev server |
| Tanstack Query | latest | Server state (API data) |
| Pinia | latest | UI state only |
| shadcn-vue | latest | Component library (owned, not dependency) |
| Tailwind CSS | v4 | Styling |
| Vue Router | latest | Client-side routing |
| Native EventSource | browser built-in | SSE client |

### Infrastructure
- Single Docker container (multi-stage build)
- SQLite file mounted as a volume at `/data/flags.db`
- Go binary embeds built Vue assets via `go:embed`

---

## Project Structure

```
toggle-flow/
├── backend/
│   ├── cmd/
│   │   └── server/             # main entry point (like main.ts in NestJS)
│   ├── internal/
│   │   ├── api/                # Fiber route handlers (like NestJS controllers)
│   │   │   ├── flags.go
│   │   │   ├── environments.go
│   │   │   └── audit.go
│   │   ├── db/                 # Bun models and queries (like TypeORM entities)
│   │   │   ├── models.go
│   │   │   └── migrations.go
│   │   ├── eval/               # Flag evaluation engine (pure business logic)
│   │   │   ├── engine.go
│   │   │   └── hash.go
│   │   ├── stream/             # SSE broker (in-process pub/sub)
│   │   │   └── broker.go
│   │   └── ui/                 # go:embed for Vue dist
│   │       └── embed.go
│   ├── go.mod                  # like package.json
│   └── go.sum                  # like package-lock.json
├── frontend/
│   ├── src/
│   │   ├── views/              # page-level components (like Angular page components)
│   │   ├── components/         # reusable components
│   │   │   └── ui/             # shadcn-vue components (owned source)
│   │   ├── stores/             # Pinia stores (UI state only)
│   │   ├── api/                # API client functions
│   │   └── composables/        # reusable Vue logic (like Angular services)
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile
├── CLAUDE.md
└── README.md
```

---

## Data Model

```
Project
  └── Environment  (has one SDK key, scoped read-only)
        └── Flag
              ├── key           string, unique per project
              ├── name          string
              ├── description   string
              ├── variations    [{value: any, type: bool|string|number|json}]
              ├── defaultVariation  index into variations
              ├── enabled       bool, per environment
              ├── rules         []TargetingRule (ordered, first-match wins)
              └── rollout       []VariationWeight (percentage split)

TargetingRule
  ├── conditions  []Condition (AND-ed)
  └── serve       variation index OR rollout

Condition
  ├── attribute   string  (e.g. "email", "country", "plan")
  ├── operator    enum    (in, notIn, contains, startsWith, gt, lt, ...)
  └── values      []any

AuditEntry
  ├── actor       string
  ├── action      string
  ├── resource    string
  ├── oldValue    json
  ├── newValue    json
  └── timestamp   time
```

---

## API Design

### Flag Management
```
POST   /api/projects/:pid/flags
GET    /api/projects/:pid/flags
GET    /api/projects/:pid/flags/:key
PATCH  /api/projects/:pid/flags/:key
DELETE /api/projects/:pid/flags/:key
```

### Environments
```
POST   /api/projects/:pid/environments
GET    /api/projects/:pid/environments
```

### Audit Log
```
GET    /api/projects/:pid/audit
```

### SDK Endpoints (authenticated by sdk_key query param)
```
GET    /sdk/flags?sdk_key=<key>        all flag configs for polling
POST   /sdk/evaluate                   evaluate one flag for a user context
GET    /sdk/stream?sdk_key=<key>       SSE stream for real-time updates
```

---

## Flag Evaluation Order

```
1. Flag disabled?
   → return defaultVariation

2. Walk targeting rules (top to bottom, first match wins)
   → conditions are AND-ed
   → if matched: serve variation OR run % rollout

3. Percentage rollout?
   → consistent hash(userKey + flagKey) % 100
   → bucket into variation by cumulative weights

4. No rule matched
   → return defaultVariation
```

---

## Environment Variables

| Variable | Default | Required | Description |
|---|---|---|---|
| `ADMIN_TOKEN` | — | Yes | Auth token for dashboard + API |
| `PORT` | `8080` | No | HTTP listen port |
| `DB_PATH` | `/data/flags.db` | No | SQLite file path |
| `LOG_LEVEL` | `info` | No | debug / info / warn / error |

---

## Development Setup

```bash
# Backend — hot reload via Air (like ts-node-dev in NestJS)
cd backend && air

# Frontend — HMR via Vite (like ng serve in Angular)
cd frontend && npm install && npm run dev
```

Vite proxies `/api` and `/sdk` to `localhost:8080` in development.

---

## Docker Build

```bash
docker build -t toggleflow .
docker run -p 8080:8080 -v ./data:/data -e ADMIN_TOKEN=secret toggleflow
```

---

## Coding Conventions

### Go
- Use `internal/` for all packages not meant to be imported externally
- Errors are returned as values, not thrown — always check them
- No global state — pass dependencies explicitly via struct constructors
- Keep handlers thin — business logic lives in services, not in the API layer

### Vue
- Use `<script setup>` syntax in all components (most concise, like Angular standalone components)
- Tanstack Query for anything that comes from the API
- Pinia only for client-side UI state (selected project, sidebar state, etc.)
- One component per file, colocated styles via Tailwind classes only

### General
- No premature abstraction — solve the problem in front of you
- No comments explaining what the code does — only write a comment when the why is non-obvious
- Ask before adding a new dependency

### Git Commits
- Write commit messages in plain human style — short, lowercase, no ticket refs, no AI phrasing
- Examples: `add flag toggle endpoint`, `fix env slug uniqueness`, `wire up sse broker`
- Never add `Co-Authored-By` or any Claude attribution to commits
- Never mention AI, Claude, or code generation in any commit message
