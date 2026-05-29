# ToggleFlow — Architecture Decisions

Every significant design decision made while building ToggleFlow, with the reasoning behind each choice and the trade-offs considered.

---

## Table of Contents

1. [Single binary deployment](#1-single-binary-deployment)
2. [SQLite over Postgres](#2-sqlite-over-postgres)
3. [SQLite WAL mode](#3-sqlite-wal-mode)
4. [Fiber v2 as the HTTP framework](#4-fiber-v2-as-the-http-framework)
5. [Bun ORM with startup migrations](#5-bun-orm-with-startup-migrations)
6. [In-memory flag cache](#6-in-memory-flag-cache)
7. [ETag support on SDK flag endpoint](#7-etag-support-on-sdk-flag-endpoint)
8. [Server-Sent Events over WebSockets](#8-server-sent-events-over-websockets)
9. [SSE broker sharded by project](#9-sse-broker-sharded-by-project)
10. [Async audit log writes](#10-async-audit-log-writes)
11. [Consistent hash rollouts](#11-consistent-hash-rollouts)
12. [First-match-wins rule evaluation](#12-first-match-wins-rule-evaluation)
13. [Variations and rules stored as JSON columns](#13-variations-and-rules-stored-as-json-columns)
14. [Per-environment flag state](#14-per-environment-flag-state)
15. [SDK keys vs API keys](#15-sdk-keys-vs-api-keys)
16. [Role hierarchy — five levels](#16-role-hierarchy--five-levels)
17. [Project membership scoping](#17-project-membership-scoping)
18. [Multi-stage Docker build](#18-multi-stage-docker-build)
19. [Vue 3 with Tanstack Query for server state](#19-vue-3-with-tanstack-query-for-server-state)
20. [Pinia for UI state only](#20-pinia-for-ui-state-only)
21. [shadcn-vue — owned components](#21-shadcn-vue--owned-components)
22. [fly-demo branch for seed data](#22-fly-demo-branch-for-seed-data)

---

## 1. Single binary deployment

**Decision:** The Go backend embeds the built Vue dashboard using `go:embed`. One binary serves both the REST API and the frontend.

**Why:** The goal from day one was zero-dependency self-hosting. A single `docker run` command with a mounted volume is all anyone needs to get a running instance. Shipping two separate processes (a Node static file server plus a Go API) would require a reverse proxy, coordinated startup, and more surface area for configuration errors.

**How:** The Vite build outputs to `frontend/dist`. The Dockerfile copies that directory into `backend/internal/ui/dist` before compiling the Go binary. A `go:embed` directive in `internal/ui/embed.go` bakes the files into the binary at compile time. The Fiber app serves the SPA from that embedded filesystem.

**Trade-off:** Every frontend change requires a full Go rebuild. Acceptable for a deployable product; during development Vite's dev server handles the frontend independently and proxies `/api` to the Go process.

---

## 2. SQLite over Postgres

**Decision:** SQLite is the only database. There is no Postgres driver, no connection string, no external database process.

**Why:** The target user is a developer or small team self-hosting a feature flag service. Requiring them to provision and maintain a Postgres instance adds significant operational overhead for a tool that, for most teams, handles maybe a few hundred flag evaluations per second at peak. SQLite is a file on disk. The entire database is one file at `/data/flags.db`. Backups are a file copy. Restores are a file copy.

**How:** The SQLite file is mounted as a Docker volume at `/data`. The Go binary opens it with Bun's SQLite driver on startup.

**Trade-off:** SQLite has a single-writer constraint — only one write at a time. For a feature flag service this is fine: writes (flag changes, rule saves, audit entries) are infrequent relative to reads. The SDK read path is served entirely from an in-memory cache with zero DB queries on cache hit, so write serialisation never affects SDK latency. Postgres remains a roadmap item for teams with high-write or multi-replica requirements.

---

## 3. SQLite WAL mode

**Decision:** The database is opened with WAL (Write-Ahead Logging) mode enabled.

**Why:** SQLite's default journal mode (DELETE) takes an exclusive lock during writes, which blocks all concurrent reads. WAL allows readers and writers to proceed concurrently — readers see a consistent snapshot and are never blocked by a writer.

**How:** Bun's SQLite driver exposes a `_journal_mode=WAL` connection parameter, set in the `db.Connect()` function. This is a one-time pragma; once set it persists in the database file.

**Trade-off:** WAL produces two extra files (`.wal` and `.shm`) alongside the main database file. They are automatically checkpointed and are safe to include in a volume backup as long as you copy all three files together.

---

## 4. Fiber v2 as the HTTP framework

**Decision:** Fiber v2 is used as the HTTP framework instead of the standard library `net/http` or alternatives like Chi, Echo, or Gin.

**Why:** The two hottest paths in ToggleFlow are `GET /sdk/flags` (SDK polling) and `POST /sdk/evaluate` (per-request flag evaluation). These can receive thousands of requests per second from SDK clients. Fiber is built on `fasthttp`, which avoids allocating a new `http.Request` object per connection by pooling them — significantly reducing GC pressure at high throughput. The routing and middleware API is ergonomic and close to Express, which maps well conceptually to NestJS controllers.

**How:** Route handlers receive a `*fiber.Ctx` instead of `(http.ResponseWriter, *http.Request)`. Fiber handles graceful shutdown, TLS, and middleware chaining natively.

**Trade-off:** `fasthttp` reuses request context objects from a pool, which means you cannot hold a reference to `fiber.Ctx` after the handler returns. This is why audit log goroutines extract `actor := h.actorName(c)` synchronously before launching the goroutine — the context object may be recycled by the time the goroutine runs.

---

## 5. Bun ORM with startup migrations

**Decision:** Bun is used as the ORM. Migrations run automatically on every startup; there is no migration CLI.

**Why:** Bun is the most ergonomic Go ORM for SQLite — it supports type-safe queries, struct tags for column mapping, and has a simple migration system. The no-CLI migration approach means deployment is a single step: start the container, migrations run, the app is ready. There is no "run migrations first, then deploy" coordination problem.

**How:** `db.Migrate(database)` is called in `main.go` before the HTTP server starts. Bun tracks applied migrations in a `bun_migrations` table and only runs new ones.

**Trade-off:** Auto-migrations on startup require that migrations are always backwards-compatible (additive only). Destructive migrations — dropping a column, renaming a table — need to be handled carefully. For an early-stage product this is an acceptable constraint; it prevents the entire class of "forgot to run the migration" deployment failures.

---

## 6. In-memory flag cache

**Decision:** Flag configurations are cached in memory as pre-marshalled JSON bytes, keyed by `(projectID, environmentID)`. SDK read requests are served from this cache with zero database queries on a cache hit.

**Why:** SDK clients — especially in serverless environments — may call `/sdk/flags` or `/sdk/evaluate` on every request. If each SDK call hit the database, SQLite's single-writer serialisation would become a bottleneck and latency would be unpredictable under load. By caching pre-serialised JSON, the hot path is a map lookup and a `c.Send(bytes)` call — no unmarshalling, no query planning, no disk I/O.

**How:** The cache is a `sync.Map` keyed by `(projectID, environmentID)` storing `[]byte`. It is populated on first read (lazy) and explicitly invalidated after every write that changes flag or environment state. There is no TTL — the cache is only busted by explicit writes, so it never goes stale while the process is running.

**Trade-off:** The cache lives in a single process. If ToggleFlow were horizontally scaled across multiple machines each instance would have its own cache, and a flag change on one instance would not invalidate the cache on others. This is a known limitation of the single-container model and is consistent with the SQLite single-node constraint. Postgres + a shared invalidation mechanism would solve this for multi-replica deployments.

---

## 7. ETag support on SDK flag endpoint

**Decision:** `GET /sdk/flags` returns an `ETag` header (a SHA-256 hash of the response body). SDK clients that send `If-None-Match` receive a `304 Not Modified` when flags have not changed, with no body.

**Why:** SDK clients that poll for flag updates (rather than using SSE) would otherwise re-download the full flag payload on every poll cycle even when nothing changed. For a project with many flags and a fast poll interval this wastes bandwidth and CPU on both sides. With ETags, a poll cycle where nothing changed costs one network round-trip and two header comparisons — no JSON parsing, no cache busting.

**How:** The ETag is computed as `sha256(cachedBytes)` immediately before sending. Because the cache stores pre-serialised bytes, the hash computation is cheap. The `If-None-Match` check compares the request header to the computed ETag string; on match, Fiber returns `304` with no body.

**Trade-off:** ETags are per-instance — a load balancer distributing requests across multiple ToggleFlow instances would cause spurious cache misses if different instances have different in-memory caches. Again, consistent with the single-instance model.

---

## 8. Server-Sent Events over WebSockets

**Decision:** Real-time flag updates are pushed to SDK clients over SSE (`GET /sdk/stream`) rather than WebSockets.

**Why:** SSE is unidirectional — the server pushes events, the client only listens. Flag updates are exactly this pattern: the SDK never needs to send data back on the stream. WebSockets are bidirectional and carry overhead (framing protocol, ping/pong, upgrade negotiation) that buys nothing here. SSE runs over plain HTTP/1.1, works transparently through most proxies and CDNs, and is supported natively by every browser via `EventSource` with automatic reconnection built in.

**How:** The Fiber handler writes `text/event-stream` headers and flushes `data: {...}\n\n` frames on every flag change event received from the broker. Go goroutines make holding thousands of open SSE connections cheap — each connection is a goroutine blocked on a channel receive, costing roughly 2–4 KB of stack.

**Trade-off:** SSE connections are long-lived HTTP requests. Some proxies (notably nginx with default config) will buffer or close them. The Fly.io deployment sets `force_https = true` and the health check uses a short-lived GET, so SSE connections work without special proxy config. Customers self-hosting behind nginx need to set `proxy_buffering off` and `proxy_read_timeout` appropriately.

---

## 9. SSE broker sharded by project

**Decision:** The in-process SSE broker maintains a separate subscriber set per project (`map[int64]map[chan Event]struct{}`). Publishing a flag change only fans out to subscribers of that project.

**Why:** Without sharding, every flag change — regardless of which project it belongs to — would iterate over every connected SSE client globally. In a multi-tenant deployment with many projects and many connected SDKs this becomes O(all_clients_globally) per publish. With project sharding it is O(clients_in_that_project), which is always smaller and usually much smaller.

**How:** `broker.Subscribe(projectID)` creates or finds the subscriber set for that project and adds the channel. `broker.Publish(event)` reads `event.ProjectID` and only iterates `subs[event.ProjectID]`. Unsubscribe removes the channel from that project's set and cleans up the outer map entry if it becomes empty.

**Trade-off:** The broker holds a `sync.RWMutex`. Subscribe and Unsubscribe take a write lock; Publish takes a read lock (since it only reads the map and sends on channels). This means concurrent publishes to different projects are non-blocking relative to each other, which is the right behaviour.

---

## 10. Async audit log writes

**Decision:** Audit log entries are written in a goroutine (`go h.writeAudit(...)`) rather than blocking the request response.

**Why:** Audit writes are fire-and-forget — the client does not need to wait for the audit entry to be persisted before receiving the API response. Making the client wait adds SQLite write latency (typically 1–5ms) to every flag toggle, rule save, and user action. Over time this adds up and makes the dashboard feel sluggish.

**How:** Before launching the goroutine, the actor name is resolved synchronously from the Fiber context (`actor := h.actorName(c)`). This is critical — Fiber recycles `*fiber.Ctx` objects from a pool after the handler returns, so any access to the context inside the goroutine would be a data race. The goroutine only receives plain `string` and `int64` values that are safe to use after the handler returns.

**Trade-off:** If the process crashes between a flag toggle and the audit write completing, the audit entry is lost. This is an acceptable trade-off for an audit log — the flag state itself was persisted synchronously, which is what matters for correctness. The audit log is best-effort observability, not a transaction log.

---

## 11. Consistent hash rollouts

**Decision:** Percentage rollouts use `sha256(flagKey + "." + userKey) % 100` to assign users to buckets. The same algorithm is implemented identically in the Go backend, the JavaScript SDK, the Python SDK, and the Go SDK.

**Why:** Rollouts must be consistent — the same user must always get the same variation for a given flag. A random assignment would re-roll on every evaluation, giving users a different experience each time. By hashing the flag key and user key together we get a deterministic, uniformly distributed bucket number that requires no persistent storage of "which variation did user X get for flag Y".

**How:** Take the first 4 bytes of the SHA-256 digest as a big-endian uint32, then mod 100. The flag key is included in the hash input so that a user bucketed into the 20% cohort for one flag is not necessarily in the 20% cohort for a different flag — each flag gets an independent, uncorrelated distribution.

**Trade-off:** The bucket assignment is immutable for a given `(flagKey, userKey)` pair. If you want to "re-roll" a rollout you must change the flag key, which creates a new flag. This is intentional — it prevents accidental re-bucketing of users who have already been assigned to a variation.

---

## 12. First-match-wins rule evaluation

**Decision:** Targeting rules are evaluated top-to-bottom and the first matching rule wins. Subsequent rules are not evaluated.

**Why:** This is the industry-standard evaluation model (LaunchDarkly, Split, Unleash all use it). It gives flag authors a clear mental model: put the most specific rules at the top. A "VIP users" rule at the top will catch those users before a broader "all beta users" rule below it. If rules were evaluated independently and could conflict, the resolution would be ambiguous.

**How:** The evaluation engine in `internal/eval/engine.go` iterates rules in order, evaluates each rule's conditions (AND-ed), and returns immediately on the first match. If no rule matches, the flag's `defaultVariation` is served.

**Trade-off:** Rule ordering matters, which means reordering rules changes behaviour. This is a feature, not a bug — it gives operators fine-grained control. The UI presents rules in an ordered list with drag-to-reorder to make the ordering explicit.

---

## 13. Variations and rules stored as JSON columns

**Decision:** Flag variations and targeting rules are stored as JSON strings in SQLite columns rather than as normalised relational rows.

**Why:** Variations and rules are always read and written as a complete unit — you never need to query "give me only the third condition of the second rule of this flag". Storing them as JSON means a flag fetch is a single row read with no joins. It also means the schema for rules is flexible — adding a new operator or condition type requires no migration, only a code change in the evaluation engine.

**How:** Variations are a `[]map[string]any` serialised to a JSON string column. Rules follow the same pattern. Bun handles the marshalling/unmarshalling via struct tags.

**Trade-off:** You lose the ability to query inside rules using SQL predicates (e.g. "find all flags that use segment X"). We work around this by loading all flags for a project at once and filtering in Go, which is fast enough given typical project sizes (tens to low hundreds of flags).

---

## 14. Per-environment flag state

**Decision:** Each flag has a separate `FlagEnvironment` row per environment, storing `enabled`, `defaultVariation`, and `rules` independently. Changing a flag in staging does not touch production.

**Why:** This is the core value proposition of a feature flag service. Developers need to be able to test a flag in development and staging without affecting production users. The flag's metadata (name, key, variations) is shared across environments; the operational state (on/off, which variation is default, what rules apply) is environment-specific.

**How:** The `flag_environments` table has a composite primary key of `(flag_id, environment_id)`. The SDK authenticates with an environment-scoped SDK key, so it only ever sees the state for its own environment.

**Trade-off:** Creating a new flag creates a `FlagEnvironment` row for every environment in the project. This is intentional — a flag that does not exist in an environment is ambiguous (is it off? does the SDK crash?). Having a row in every environment with `enabled = false` as the default makes the state explicit and unambiguous.

---

## 15. SDK keys vs API keys

**Decision:** Two separate key types exist. SDK keys are environment-scoped and read-only. API keys are project-scoped and have write access.

**Why:** SDK keys are embedded in client-side or edge code — mobile apps, browser bundles, CDN workers. They will be exposed. They must not be able to modify flags. API keys are used by CI/CD pipelines and server-side automation — they need write access but should be scoped to a single project, not the entire instance.

**How:** SDK keys are prefixed `sdk_` and are looked up against a `sha256` hash stored in the `sdk_keys` table (the raw key is never stored). API keys are prefixed `tfk_` and follow the same hash storage pattern. The authentication middleware checks the prefix to determine which table to query and what permissions to grant.

**Trade-off:** Key rotation requires issuing a new key and updating all consumers. There is currently no key expiry — keys are valid until explicitly deleted. Expiry is a roadmap item.

---

## 16. Role hierarchy — five levels

**Decision:** Five roles exist in strict order: `superuser > admin > owner > editor > viewer`.

**Why:** Different users need different scopes of access. A viewer embedded in a monitoring dashboard should not be able to toggle flags. An editor working on flags should not be able to delete environments. An owner managing their project should not be able to delete other users. Superusers (instance administrators) should be able to do everything. Five levels maps cleanly to these real use cases without over-complicating the permission model.

| Role | Can do |
|---|---|
| Viewer | Read flags, audit log, environments |
| Editor | Everything Viewer can do, plus create/edit/delete flags and segments |
| Owner | Everything Editor can do, plus manage environments, SDK keys, API keys, and project members |
| Admin | Everything Owner can do, plus invite and manage users across the instance |
| Superuser | Everything Admin can do, plus delete users and generate password reset links |

**How:** Roles are stored as strings. The middleware `auth.RequireRole(minRole)` checks that the authenticated user's role is at or above the required level using an integer comparison against a `roleRank` map.

**Trade-off:** Roles are instance-wide, not per-project. A user who is an Editor on one project is an Editor on all projects they are a member of. Per-project roles would be more flexible but significantly more complex to implement and reason about for small teams.

---

## 17. Project membership scoping

**Decision:** Users must be explicitly added to a project as members to access its flags, environments, and audit log.

**Why:** ToggleFlow is multi-tenant at the project level. An organisation might have separate projects for different products or teams. A user working on Project A should not automatically see Project B's flags. Explicit membership is the safest default.

**How:** The `project_members` join table maps `(project_id, user_id)`. Every handler that accesses project-scoped resources checks membership before proceeding. Admins and Superusers can see all projects without being explicit members.

**Trade-off:** Adding a new team member to ToggleFlow requires two steps: create the user account (Admin action) and add them to the relevant project (Owner action). This is by design — a new user should not automatically have access to all projects.

---

## 18. Multi-stage Docker build

**Decision:** The Dockerfile has three stages: Node builds the Vue dashboard, Go compiles the binary with embedded frontend, Alpine produces the final image.

**Why:** A single-stage build that includes both Node and Go toolchains would produce an image of several hundred megabytes. The final image only needs the compiled Go binary and the Alpine runtime — no Node, no Go toolchain, no source code. The three-stage approach produces a ~14 MB image.

**How:**
- Stage 1 (`node:20-alpine`): `pnpm install && pnpm build` → outputs `dist/`
- Stage 2 (`golang:1.25-alpine`): copies `dist/` into the backend source tree, runs `go build` — the `go:embed` directive bakes `dist/` into the binary
- Stage 3 (`alpine:3.19`): copies only the compiled binary from Stage 2

**Trade-off:** Docker layer caching is critical for build speed. The `COPY go.mod go.sum` + `RUN go mod download` step is placed before `COPY backend/` so that the module download layer is only invalidated when dependencies change, not on every source change. Same pattern on the frontend: `COPY package.json pnpm-lock.yaml` before `COPY frontend/`.

---

## 19. Vue 3 with Tanstack Query for server state

**Decision:** All data fetched from the API is managed by Tanstack Query (`useQuery` / `useMutation`). No manual `fetch` calls, no Pinia state for API data.

**Why:** Server state has different characteristics than UI state — it lives on the server, can go stale, needs refetching, and needs to be invalidated after mutations. Tanstack Query handles all of this: automatic background refetching, cache invalidation on mutation, loading and error states, deduplication of parallel requests. Writing this manually with `fetch` + Pinia would require reimplementing most of Tanstack Query and doing it worse.

**How:** Each resource (flags, environments, audit entries) has a `useQuery` composable in `src/api/`. Mutations call `queryClient.invalidateQueries` on success to trigger automatic refetch of affected data.

**Trade-off:** Tanstack Query is an additional dependency (~13 KB gzipped). Worth it for the correctness and developer experience it provides.

---

## 20. Pinia for UI state only

**Decision:** Pinia stores are only used for client-side UI state — the selected project, the active environment, sidebar open/closed. Never for API data.

**Why:** Mixing server state and UI state in Pinia causes the classic problem of stale data — the Pinia store holds a snapshot of flags that was fetched at some point but may no longer reflect server state. Tanstack Query owns server state with its own cache and invalidation logic. Pinia owns purely local state that has no server representation.

**How:** `useProjectStore()` stores `selectedProjectId` and `selectedEnvironmentId`. These IDs are used as parameters in Tanstack Query's `useQuery` calls but are never combined with API response data in the Pinia store.

**Trade-off:** Developers coming from NgRx or Redux may instinctively reach for the Pinia store for everything. The discipline of "Pinia = UI, Tanstack Query = server" needs to be maintained as the codebase grows.

---

## 21. shadcn-vue — owned components

**Decision:** UI components (Button, Dialog, Table, Badge, etc.) come from shadcn-vue, but they live in the project's source code (`src/components/ui/`) rather than in `node_modules`.

**Why:** shadcn-vue is not a traditional component library — it is a collection of component source files you own. This means components can be modified freely without forking a dependency or waiting for upstream changes. It also means the exact component code is auditable and testable. There is no `shadcn-vue` in `package.json`.

**How:** Components were generated by the shadcn-vue CLI into `src/components/ui/` and are imported like any other local component. They use Radix Vue for accessible primitives and Tailwind for styling.

**Trade-off:** Updating to a new shadcn-vue version means manually re-running the CLI or cherry-picking changes. There is no `npm update` for components. This is the explicit trade-off shadcn-vue makes — you own the code, you own the updates.

---

## 22. fly-demo branch for seed data

**Decision:** The seed data function (`db.Seed`) and its invocation in `main.go` live on the `fly-demo` branch, not on `main`.

**Why:** The seed function creates demo users with a hardcoded password and populates the database with fictional data. This code has no place in production deployments and pollutes `main` with deployment-specific concerns. Keeping it on a separate branch makes it clear that the demo environment is a special deployment of a specific branch, not the canonical codebase.

**How:** `fly-demo` branches off `main` after the last production commit. Fly.io is configured to deploy from the `fly-demo` branch. `main` remains clean. Any production-ready changes go to `main` first, then are merged into `fly-demo` to keep the demo environment up to date.

**Trade-off:** The demo deployment is always one merge behind `main`. This is acceptable — the demo exists to show features, not to track the bleeding edge of development.
