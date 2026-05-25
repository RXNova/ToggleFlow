# ToggleFlow

A self-hostable feature flag and configuration management system. Run it as a single Docker container — no external dependencies.

---

## What it does

- Create and manage feature flags across multiple projects and environments
- Target specific users with attribute-based rules (email, country, plan, etc.)
- Gradually roll out features to a percentage of users with consistent hashing
- Multivariate flags — boolean, string, number, or JSON variations
- Real-time flag updates pushed to connected SDKs via SSE
- Full audit log of every flag change

---

## Self-hosting

```bash
docker run -p 8080:8080 \
  -v ./data:/data \
  -e ADMIN_TOKEN=your-secret-token \
  toggleflow:latest
```

Dashboard: `http://localhost:8080`  
API: `http://localhost:8080/api`  
SDK stream: `http://localhost:8080/sdk/stream?sdk_key=<key>`

Data is persisted to `/data/flags.db` — mount a volume to keep it across restarts.

---

## Tech Stack

### Backend

#### Go
The entire backend compiles to a single binary. The binary embeds the built Vue frontend assets via `go:embed`, so one executable serves both the API and the dashboard. Goroutines make it cheap to hold thousands of concurrent SSE connections open — each connected SDK client gets a goroutine, and Go schedules them efficiently without the overhead of OS threads.

#### Fiber v2
HTTP framework built on `fasthttp` instead of the standard `net/http`. This makes it 2–3x faster for high-throughput endpoints like `/sdk/evaluate`, which gets called on every flag check in client applications. The API is Express-like — easy to pick up and extend. Fiber ships built-in middleware for CORS, rate limiting, request logging, and gzip compression, so no extra packages are needed for those.

#### Bun ORM
Database access layer with SQLite support. Define Go structs, get type-safe queries — no SQL files to maintain and no code generation step. Bun handles migrations via struct tags and is significantly faster than GORM while being much easier to use than raw `database/sql` or `sqlc`. Adding a new table or query is a struct change plus one method call.

```go
// Fetch all flags for an environment
var flags []Flag
err := db.NewSelect().
    Model(&flags).
    Where("environment_id = ?", envID).
    Where("deleted_at IS NULL").
    OrderExpr("created_at DESC").
    Scan(ctx)
```

#### SQLite (WAL mode)
Single-file database — no separate process, no connection string, no credentials. Write-Ahead Logging (WAL) mode enables concurrent reads without blocking writes, which matters because flag evaluation is read-heavy. The database file lives at `/data/flags.db` and is the only thing that needs to be persisted.

This handles thousands of read requests per second comfortably for a self-hosted deployment. If you need horizontal scaling or have very high write volume, swap in Postgres by changing the Bun driver — the rest of the code stays the same.

#### In-process SSE Broker
Flag changes are broadcast to connected SDK clients using a Go `sync.Map` of channels — one channel per connected client. When a flag is updated, the API handler writes the change to SQLite, then publishes to the broker, which fans out to all connected clients instantly. No Redis, no message queue — just Go channels.

```
Flag update request
        │
        ▼
  SQLite write
        │
        ▼
  Broker.Publish()
        │
   ┌────┼────┐
   ▼    ▼    ▼
SDK1  SDK2  SDK3   ← each holding an SSE connection
```

---

### Frontend

#### Vue 3
Component framework for the dashboard. The Composition API keeps component logic clean and reusable. The built output is static files that Go embeds directly into the binary — no Node.js or separate web server needed in production.

#### Vite
Build tool and dev server. Hot module replacement makes frontend development fast. In development, Vite proxies API requests to the Go server so you run both processes and get instant feedback on UI changes without rebuilding the binary.

#### Tanstack Query (Vue Query)
Manages all server state — fetching, caching, refetching, and mutations. Replaces the need for Pinia stores for API data. Loading states, error states, and cache invalidation are handled automatically. After a flag update mutation, Tanstack Query invalidates the flags query and refetches — no manual store updates needed.

```ts
// Fetch flags — loading, error, and caching handled automatically
const { data: flags, isLoading } = useQuery({
  queryKey: ['flags', projectId, envId],
  queryFn: () => api.getFlags(projectId, envId)
})

// Update a flag — cache invalidated automatically on success
const { mutate: toggleFlag } = useMutation({
  mutationFn: (key: string) => api.toggleFlag(projectId, key),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['flags'] })
})
```

#### Pinia
Used only for UI state that doesn't come from the server — the currently selected project, sidebar open/closed state, active environment. Keeping it separate from server state (Tanstack Query) avoids the complexity of syncing cache with a store.

#### shadcn-vue + Tailwind CSS
shadcn-vue provides accessible, unstyled components (dialogs, dropdowns, tables, toggles) built on Radix Vue primitives. Tailwind handles all styling. Components live in your codebase, not in a node_modules dependency, so they're easy to modify.

#### Native EventSource API
SSE client built into every browser — no package needed. SDK keys are passed as a query parameter, which is how LaunchDarkly's own SDK stream authentication works. SDK keys are read-only and scoped to a single environment, so query param auth is appropriate.

```ts
const stream = new EventSource(`/sdk/stream?sdk_key=${sdkKey}`)
stream.onmessage = (event) => {
  const patch = JSON.parse(event.data)
  applyFlagPatch(patch)
}
```

---

### Container

Multi-stage Docker build — the final image contains only the compiled Go binary and an Alpine base. No Node.js, no Go toolchain, no build artifacts.

```dockerfile
# Stage 1 — build Vue dashboard
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/ .
RUN npm ci && npm run build

# Stage 2 — build Go binary with embedded frontend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/ .
COPY --from=frontend /app/dist ./internal/ui/dist
RUN go build -o toggleflow ./cmd/server

# Stage 3 — final image (~20MB)
FROM alpine:3.19
COPY --from=backend /app/toggleflow /toggleflow
EXPOSE 8080
ENTRYPOINT ["/toggleflow"]
```

---

## Project Structure

```
toggle-flow/
├── backend/
│   ├── cmd/
│   │   └── server/             # main entry point
│   ├── internal/
│   │   ├── api/                # Fiber route handlers
│   │   │   ├── flags.go
│   │   │   ├── environments.go
│   │   │   └── audit.go
│   │   ├── db/                 # Bun models and queries
│   │   │   ├── models.go
│   │   │   └── migrations.go
│   │   ├── eval/               # flag evaluation engine
│   │   │   ├── engine.go       # targeting rules, % rollout
│   │   │   └── hash.go         # consistent hashing
│   │   ├── stream/             # SSE broker
│   │   │   └── broker.go
│   │   └── ui/                 # go:embed for Vue dist
│   │       └── embed.go
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── views/              # page components
│   │   │   ├── FlagsView.vue
│   │   │   ├── AuditView.vue
│   │   │   └── SettingsView.vue
│   │   ├── components/         # reusable UI components
│   │   ├── stores/             # Pinia — UI state only
│   │   ├── api/                # API client functions
│   │   └── composables/        # Vue composables
│   ├── package.json
│   └── vite.config.ts
└── Dockerfile
```

---

## API Reference

### Flag Management
```
POST   /api/projects/:pid/flags              create flag
GET    /api/projects/:pid/flags              list flags
GET    /api/projects/:pid/flags/:key         get flag
PATCH  /api/projects/:pid/flags/:key         update flag
DELETE /api/projects/:pid/flags/:key         delete flag
```

### Environments
```
POST   /api/projects/:pid/environments       create environment
GET    /api/projects/:pid/environments       list environments
```

### Audit Log
```
GET    /api/projects/:pid/audit              paginated audit log
```

### SDK Endpoints
```
GET    /sdk/flags?sdk_key=<key>              fetch all flag configs (polling)
POST   /sdk/evaluate                         evaluate a flag for a user context
GET    /sdk/stream?sdk_key=<key>             SSE stream for real-time updates
```

---

## Flag Evaluation Order

Every flag evaluation follows this order:

1. Flag disabled → return default variation
2. Walk targeting rules top-to-bottom, first match wins
   - Rule conditions are AND-ed together
   - Supported operators: `in`, `notIn`, `contains`, `startsWith`, `endsWith`, `gt`, `lt`, `gte`, `lte`, `semVerGt`, `semVerLt`
3. Percentage rollout rule → `hash(userKey + flagKey) % 100`
4. No rule matched → return default variation

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `ADMIN_TOKEN` | — | Required. Token for dashboard and API access |
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `/data/flags.db` | SQLite file path |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

---

## Development

**Requirements:** Go 1.23+, Node 20+

```bash
# Terminal 1 — Go backend with hot reload
cd backend
air

# Terminal 2 — Vue frontend with HMR
cd frontend
npm install
npm run dev
```

Vite proxies `/api` and `/sdk` to `localhost:8080` during development.

---

## Roadmap

- [ ] User segments — reusable groups for targeting rules
- [ ] Webhooks — notify external services on flag changes
- [ ] Scheduled flag changes — enable/disable at a set time
- [ ] Postgres support — for high-write or multi-replica deployments
- [ ] Official SDKs — JavaScript, Go, Python
