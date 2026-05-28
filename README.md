# ToggleFlow

A self-hostable feature flag and configuration management system. Runs as a single Docker container with no external dependencies — drop it in front of any stack and start shipping safer.

---

## Features

- **Multivariate flags** — boolean, string, number, and JSON variation types
- **Per-environment state** — independent enabled/disabled and default variation per environment
- **Targeting rules** — serve a variation based on user attributes (email, country, plan, etc.)
- **Percentage rollouts** — consistent hash-based bucketing ensures the same user always gets the same variation
- **Segments** — reusable user groups referenced across multiple flag rules
- **Real-time updates** — flag changes pushed to connected SDK clients instantly over SSE
- **Audit log** — every flag, environment, and user change is recorded with actor and before/after values
- **User management** — invite users by email, role-based access (superuser / admin / owner / editor / viewer)
- **Project members** — scope users to specific projects
- **SDK keys** — per-environment read-only keys for SDK authentication
- **API keys** — project-scoped keys for programmatic flag management

---

## Quick Start

```bash
docker run -p 8080:8080 \
  -v ./data:/data \
  -e ADMIN_TOKEN=your-secret-token \
  ghcr.io/toggleflow/toggleflow:latest
```

Open `http://localhost:8080` to access the dashboard.

Data is persisted to `/data/flags.db`. Mount a volume to keep it across container restarts.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `ADMIN_TOKEN` | — | **Required.** Initial superuser token for first-time setup |
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `/data/flags.db` | SQLite file path |

---

## SDK Endpoints

These endpoints are authenticated by SDK key — passed as a query parameter for SSE compatibility.

```
GET   /sdk/flags?sdk_key=<key>      Fetch all flag configs for an environment
POST  /sdk/evaluate                  Evaluate a single flag for a user context
GET   /sdk/stream?sdk_key=<key>     SSE stream — receive flag changes in real time
```

### Evaluating a flag

```bash
curl -X POST https://your-host/sdk/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "sdk_key": "sdk_...",
    "flag_key": "dark-mode",
    "user_key": "user-123",
    "context": {
      "email": "user@example.com",
      "plan": "pro",
      "country": "US"
    }
  }'
```

```json
{
  "flag_key": "dark-mode",
  "enabled": true,
  "variation_index": 1,
  "variation_value": true
}
```

### Streaming updates

```js
const stream = new EventSource(`/sdk/stream?sdk_key=${sdkKey}`)
stream.onmessage = (e) => {
  const { type, flag, env } = JSON.parse(e.data)
  // type: "flag.updated" | "flag.deleted" | "flag.created" | "connected"
  refetchFlags()
}
```

---

## Flag Evaluation Order

Every evaluation follows this sequence:

```
1. Flag disabled?
   → return defaultVariation

2. Walk targeting rules top-to-bottom (first match wins)
   → conditions are AND-ed
   → operators: in, notIn, equals, contains, startsWith, endsWith, gt, gte, lt, lte

3. Matched rule has a rollout?
   → hash(flagKey + "." + userKey) % 100
   → bucket into variation by cumulative weights

4. No rule matched
   → return defaultVariation
```

The same evaluation logic runs inside SDK clients — no round-trip needed per flag check.

---

## API Reference

All management endpoints require a `Bearer <token>` header (JWT from login, or a project-scoped API key prefixed `tfk_`).

### Auth
```
POST  /api/auth/login
POST  /api/auth/activate
POST  /api/auth/reset
GET   /api/auth/invite/:uuid
GET   /api/auth/reset/:uuid
GET   /api/auth/me
PATCH /api/auth/profile
```

### Users (superuser / admin only)
```
GET    /api/users
POST   /api/users
PATCH  /api/users/:id
DELETE /api/users/:id
POST   /api/users/:id/reinvite
POST   /api/users/:id/reset-link
GET    /api/users/:id/audit
```

### Projects
```
GET    /api/projects
POST   /api/projects
PATCH  /api/projects/:id
DELETE /api/projects/:id
```

### Environments
```
GET    /api/projects/:pid/environments
POST   /api/projects/:pid/environments
PATCH  /api/projects/:pid/environments/:eid
DELETE /api/projects/:pid/environments/:eid
```

### Flags
```
GET    /api/projects/:pid/flags
GET    /api/projects/:pid/flags/:key
POST   /api/projects/:pid/flags
PATCH  /api/projects/:pid/flags/:key
PATCH  /api/projects/:pid/flags/:key/env
PUT    /api/projects/:pid/flags/:key/rules
DELETE /api/projects/:pid/flags/:key
```

### Segments
```
GET    /api/projects/:pid/segments
POST   /api/projects/:pid/segments
PATCH  /api/projects/:pid/segments/:sid
DELETE /api/projects/:pid/segments/:sid
```

### Members
```
GET    /api/projects/:pid/members
POST   /api/projects/:pid/members
DELETE /api/projects/:pid/members/:uid
```

### SDK Keys
```
GET    /api/projects/:pid/environments/:eid/sdk-keys
POST   /api/projects/:pid/environments/:eid/sdk-keys
DELETE /api/projects/:pid/environments/:eid/sdk-keys/:kid
```

### API Keys
```
GET    /api/projects/:pid/api-keys
POST   /api/projects/:pid/api-keys
DELETE /api/projects/:pid/api-keys/:kid
```

### Audit Log
```
GET    /api/projects/:pid/audit
```

---

## Tech Stack

### Backend

**Go 1.23** — compiles to a single static binary. The binary embeds the built Vue dashboard via `go:embed`, so one executable serves both the API and the UI. Goroutines make holding thousands of concurrent SSE connections cheap.

**Fiber v2** — HTTP framework built on `fasthttp`. Chosen for performance on high-throughput SDK endpoints like `/sdk/flags` and `/sdk/evaluate`.

**Bun ORM** — type-safe queries against SQLite. Migrations run automatically on startup — no migration CLI, no manual steps.

**SQLite (WAL mode)** — single-file embedded database. WAL mode allows concurrent reads without blocking writes. The entire database is one file at `/data/flags.db`.

**In-memory flag cache** — flag configs are cached in memory as pre-marshalled JSON bytes, keyed by `(projectID, environmentID)`. The cache is busted explicitly after every write. SDK read requests are served entirely from memory on cache hit with no DB queries.

**ETag support** — `GET /sdk/flags` returns an `ETag` header. Clients that send `If-None-Match` get a `304 Not Modified` when flags haven't changed, saving bandwidth on every polling cycle.

**SSE broker (sharded by project)** — each project has its own subscriber set. A flag change in project A only fans out to project A's connected clients — not all clients globally.

### Frontend

**Vue 3** — dashboard UI with Composition API and `<script setup>`.

**Tanstack Query** — all server state (flags, environments, audit entries) is managed through `useQuery` / `useMutation`. Cache invalidation is automatic on mutation success.

**Pinia** — used only for client-side UI state (selected project, active environment, sidebar state).

**shadcn-vue + Tailwind CSS v4** — accessible component primitives with full Tailwind styling. Components live in the codebase, not in `node_modules`.

### Infrastructure

**Multi-stage Docker build** — Stage 1 builds the Vue dashboard, Stage 2 compiles the Go binary with embedded frontend assets, Stage 3 produces a final Alpine image (~20MB) with no Node.js or Go toolchain.

---

## Development

**Requirements:** Go 1.23+, Node 20+, [Air](https://github.com/air-verse/air) (Go hot reload)

```bash
# Terminal 1 — backend with hot reload
cd backend
air

# Terminal 2 — frontend with HMR
cd frontend
pnpm install
pnpm dev
```

Vite proxies `/api` and `/sdk` to `localhost:8080` during development, so the frontend and backend work together without any extra configuration.

---

## SDKs

| Language | Repository | Package |
|---|---|---|
| JavaScript / TypeScript | [toggle-flow/js-sdk](https://github.com/toggle-flow/js-sdk) | `npm install @toggleflow/js-sdk` |
| Python | [toggle-flow/python-sdk](https://github.com/toggle-flow/python-sdk) | `pip install toggleflow` |
| Go | [toggle-flow/go-sdk](https://github.com/toggle-flow/go-sdk) | `go get github.com/toggle-flow/go-sdk` |

---

## Roadmap

- [ ] Webhooks — notify external services on flag changes
- [ ] Scheduled flag changes — enable / disable at a set time
- [ ] Postgres driver — for high-write or multi-replica deployments
