# Production Readiness / Architecture Review — GitStore

## Repo Overview

This monorepo contains 4 major deliverables:

- **Go backend server**: `gitClone/cmd/gitserver/main.go` + HTTP transport layer in `gitClone/internal/transport/http/`
- **GitClone engine (Go)**: CLI + core/storage under `gitClone/internal/`
- **Storage engine (Go)**: log-structured KV store in `gitDb/` (imported by `gitClone` as module `GitDb`)
- **Frontend (React/Vite)**: `Client/` (Vite + React + TS + Tailwind), calling backend via `/api/*`
- **Node CLI**: `cli/` (TypeScript). Supports file ops + limited Git operations for standard `.git` repos; GitStore repos are blocked without API.

Notable repo observations:

- There is **no CI pipeline** (no `.github/` workflows) and **no Docker** artifacts (no `Dockerfile` / `docker-compose.yml`).
- The repo currently contains **runtime data checked into source**: `gitClone/data/` includes `db/log` and multiple repos in `data/repos/*`. This is not production-safe for build artifacts (should be a volume/external persistent store, not committed).

## Current Architecture Map

### Backend (Go server)

- Entry point: `gitClone/cmd/gitserver/main.go`
  - Env vars: `PORT`, `GITSTORE_REPO_BASE`, `GITSTORE_DB_PATH`
  - Defaults: `./data/repos` and `./data/db`
  - Uses `http.ListenAndServe` directly (no server timeouts, no graceful shutdown).
- HTTP routing: `gitClone/internal/transport/http/router.go`
  - Uses `http.NewServeMux` with `/api/repos` and `/api/repos/` handler.
  - Adds a permissive CORS middleware with `Access-Control-Allow-Origin: *`.
- HTTP handlers: `gitClone/internal/transport/http/handlers_*.go`
  - Repo CRUD: `handlers_repos.go`
  - Branches: `handlers_branches.go`
  - Commits/push: `handlers_commits.go`
  - Merge: `handlers_merge.go`
  - Files/add: `handlers_files.go`
  - Issues: `handlers_issues.go`
- Server composition: `gitClone/internal/transport/http/server.go`
  - Constructs service layer objects from `internal/app/*`.

### App / Service layer (Go)

- Repo path resolution helper: `gitClone/internal/app/repos/resolver.go`
  - Resolves `repoBase + repoID` and checks existence + `.gitclone` directory.
  - **Does not enforce “stay within repoBase”** beyond `filepath.Join` + `Abs`.
- Branch service: `gitClone/internal/app/branches/service.go`
  - Uses per-repo KV store `gitClone/internal/infra/storage/repo_store.go`.
  - Updates global metadata in `gitClone/internal/metadata/store.go`.
- Commit service: `gitClone/internal/app/commits/service.go`
  - Implements `ListCommits`, `CreateCommit`, `PushCommits`.
  - UI semantics: **commit list reads from remote refs** (`refs/remotes/origin/<branch>`) not local refs (`refs/heads/<branch>`).
- Files service: `gitClone/internal/app/files/service.go`
  - Staging uses RepoStore + index in the per-repo DB, but also uses `os.Chdir` (see risks).
  - Write file uses `filepath.Join(repoPath, filePath)` without sandboxing (see security risks).

### GitClone “core” and “commands”

- CLI entry: `gitClone/cmd/gitclone/main.go` + `gitClone/internal/commands/*.go`
  - Commands rely heavily on `os.Getwd()` and operate on the **process CWD**.
  - This becomes relevant because the HTTP server calls `commands.*` after `os.Chdir(...)` in request handlers.

### Storage layout (on disk)

- Default repo base: `gitClone/cmd/gitserver/main.go` → `./data/repos`
  - Each repo is a directory containing `.gitclone/` (`gitClone/internal/storage/filesystem.go`)
  - Per-repo DB directory is `.gitclone/db/` (`gitClone/internal/infra/storage/repo_store.go`)
- Default metadata DB path: `gitClone/cmd/gitserver/main.go` → `./data/db`
  - Global GitDb log file at `${GITSTORE_DB_PATH}/log`
  - Keys like:
    - `repos:index`
    - `repo:<id>` (metadata JSON)
    - `repo:<id>:issues` (issues JSON, via `gitClone/internal/transport/http/server.go`)

### gitDb (log-structured KV store)

- Implementation: `gitDb/db.go`, `gitDb/record.go`, `gitDb/select.go`
  - In-memory `[]byte log` is loaded in full on Open.
  - Index rebuild on Open scans the entire in-memory log: `db.rebuildIndex()` in `gitDb/db.go`.
  - `Put` appends to in-memory log and also appends to `${path}/log`, then `Sync()`s.
  - No locking; not safe for concurrent use from multiple goroutines/handles (see reliability risks).

### Frontend (Client)

- Vite config: `Client/vite.config.js`
  - Dev proxy hardcoded to `http://localhost:8080` for `/api`.
- API client: `Client/src/lib/api.ts`
  - Uses `VITE_API_URL` if set, otherwise relative calls (Vite proxy in dev).
  - **No auth headers** included; backend also accepts all requests.
- Firebase auth: `Client/src/firebase.ts` + used in `Client/src/context/GitContext.tsx`
  - Auth is **frontend-only**: Firebase user is used to fill author fields but **no token is sent to backend**.

### Node CLI (`cli/`)

- Entry: `cli/src/cli/index.ts`
- File ops: `cli/src/cli/commands/file.ts` uses `join(repoPath, filePath)` with user input.
- Git ops: `cli/src/cli/commands/git.ts` uses `simple-git`
  - For `gitstore` repos it checks `isApiAvailable()` but then errors out (`cli/src/cli/utils/git-handler.ts`).

## Findings (Good / Risks / Blockers)

### Good architecture (keep)

- **Clear modular separation**:
  - `gitDb/` is isolated as a Go module with its own durability tests (`gitDb/gitdb_durability_test.go`).
  - `gitClone/internal/app/*` expresses a service layer separate from HTTP (`gitClone/internal/transport/http/*`).
  - Per-repo store abstraction exists (`gitClone/internal/infra/storage/repo_store.go`).
- **Append-only persistence model** in `gitDb/db.go`:
  - `Put()` appends and fsyncs, which is the right direction for crash safety.
  - A regression suite exists for multi-handle truncation bugs (`gitDb/gitdb_durability_test.go` + `gitClone/internal/app/commits/service_remote_ref_visibility_test.go`).
- **Explicit “remote ref” model**:
  - Backend commit list reading from `refs/remotes/origin/<branch>` (`gitClone/internal/app/commits/service.go`) creates a clear UX distinction: “visible after push”.
  - Frontend acknowledges this invariant and pushes after merge (`Client/src/context/GitContext.tsx`).

### Risky architecture (likely prod bugs)

- **Process-global `os.Chdir` used inside request flow** (critical concurrency hazard):
  - `gitClone/internal/transport/http/handlers_repos.go` changes process CWD during `POST /api/repos` before calling `commands.Init`.
  - `gitClone/internal/transport/http/handlers_merge.go` changes process CWD before calling `commands.Merge`.
  - `gitClone/internal/app/files/service.go` changes process CWD during staging (`StageFilesWithInfo`).
  - In a real HTTP server, requests are concurrent. Any handler changing CWD will race other handlers and can cause:
    - operations applied to the wrong repo,
    - commits/merges reading/writing the wrong `.gitclone/db`,
    - nondeterministic corruption.
  - This is a **deployment blocker** (see below).
- **“Atomic” WriteBatch is not actually atomic**:
  - `gitClone/internal/infra/storage/write_batch.go` writes a `_tx/*` marker, then writes each key, then overwrites marker as committed.
  - Readers are not transaction-aware; partial writes are visible immediately.
  - Recovery (`RecoverTransactions`) **does not roll back** partial writes; it only marks tx keys as “batch_recovered”.
  - The tx key is not unique: `txMarkerKey := fmt.Sprintf("_tx/%d", len(wb.writes))` (comment claims timestamp+count, but code uses only count). Concurrent batches with same number of writes will collide.
  - Result: under crash or concurrency, **state can become inconsistent** even if “Commit()” returns nil.
- **Repo ID path validation is inconsistent / incomplete**:
  - Create repo validates name against `/`, `\`, and `..` (`gitClone/internal/transport/http/handlers_repos.go`).
  - But most other endpoints rely on `repos.ResolveRepoPath` (`gitClone/internal/app/repos/resolver.go`), which does **not** block `..` or enforce that `absPath` stays within `repoBase`.
  - Even if routing splits on `/`, encoded traversal (`..`, `%2e%2e`) is a realistic concern in prod.
- **gitDb has no concurrency guarantees**:
  - `gitDb/Index` is a plain map without mutex (`gitDb/index.go`).
  - `DB.Put` and `DB.Get` are not synchronized; concurrent reads/writes can race and corrupt in-memory state.
  - Multiple DB handles to the same log file can interleave writes (no file locking). Append-only writes are not sufficient if multiple processes/threads write concurrently.
- **Unbounded log growth and O(N) startup**:
  - `gitDb.Open` reads the entire log into memory and rebuilds the index by scanning all records (`gitDb/db.go`).
  - There is no compaction/vacuum; “delete” is modeled by writing empty values (e.g. index clear writes empty entries in `gitClone/internal/storage/repo_store_wrappers.go`).
  - For long-running prod usage (days/weeks), this will eventually become a memory and startup-time issue.

### Deployment blockers (must fix before real prod)

1. **Global working directory mutation** (`os.Chdir`) in request paths:
   - Files: `gitClone/internal/transport/http/handlers_repos.go`, `gitClone/internal/transport/http/handlers_merge.go`, `gitClone/internal/app/files/service.go`.
   - Blocker reason: concurrency → incorrect repo operations / data corruption.
2. **No authentication / authorization on backend**:
   - CORS is `*` and no auth checks exist in `gitClone/internal/transport/http/*`.
   - Frontend uses Firebase auth but does not send an ID token (`Client/src/context/GitContext.tsx`, `Client/src/lib/api.ts`).
   - In prod, anyone with network access can create repos, write files, push, merge, create issues.
3. **No operational hardening for HTTP server**:
   - `http.ListenAndServe` with default server: no read/write/idle timeouts (`gitClone/cmd/gitserver/main.go`).
   - No graceful shutdown (signals).
4. **No defined production build/deploy pipeline**:
   - No Dockerfiles, no CI workflows, no release artifacts configuration.
5. **Data is committed in-repo**:
   - `gitClone/data/db/log` and `gitClone/data/repos/*` should not be in the deploy artifact. Needs clear “volume” story.

### Security issues

- **Path traversal in server-side file writes**:
  - `gitClone/internal/app/files/service.go` → `fullPath := filepath.Join(repoPath, filePath)` with user-provided `filePath`.
  - `filepath.Join` permits `../` to escape the repo directory unless explicitly cleaned and validated.
  - Endpoint: `POST /api/repos/:id/files` (`gitClone/internal/transport/http/handlers_files.go`).
- **Path traversal in Node CLI file ops**:
  - `cli/src/cli/commands/file.ts` → `join(absoluteRepoPath, filePath)` with user-provided `filePath`.
  - This can write outside the repo directory (e.g. `../../.ssh/config`) if user runs CLI locally.
- **Repo traversal / repoBase escape risk**:
  - `gitClone/internal/app/repos/resolver.go` does not enforce “within repoBase”.
- **CORS `*` with mutation endpoints**:
  - `gitClone/internal/transport/http/router.go` allows any origin and allows POST/PUT/DELETE.
  - If this server is reachable by browsers, it is vulnerable to cross-site request abuse unless authentication is added.

### Reliability issues (durability, atomicity, concurrency)

- **Multi-request correctness is not guaranteed**:
  - `os.Chdir` makes request processing non-thread-safe.
- **gitDb is not safe under concurrent writers**:
  - No file locking; potential log corruption and in-memory races.
- **WriteBatch transaction markers do not provide atomic semantics**:
  - Marker collisions and partial-write visibility.
- **Mixed storage responsibilities**:
  - Global metadata DB stores both repo registry and issues (`repo:<id>:issues` in `gitClone/internal/transport/http/server.go`).
  - Per-repo DB stores refs/objects/index.
  - There is no explicit cross-store transaction boundary; operations can partially succeed (repo created on disk but metadata missing, etc.).
- **Operational scalability**:
  - Full log replay on every open can become expensive (`gitDb/db.go`).
  - Per-request open/close patterns (RepoStore opened in every service call) increases churn.

### Observability gaps

- **No request IDs / structured logs**:
  - Logs are plain `log.Printf` across handlers and services.
  - No correlation IDs between HTTP request and per-repo operations.
- **No metrics**:
  - No latency/error/throughput metrics; no storage size metrics; no compaction signals.
- **No tracing**:
  - No distributed tracing hooks.
- **Health endpoints**:
  - No `/healthz` or `/readyz` endpoints exist.

## Proposed Target Architecture

Goal: make the system safe for long-running production use, with clear boundaries:

### Layering model

- **Transport layer** (`internal/transport/http`):
  - Only parses/validates input, auth, response mapping.
  - Must never mutate global process state (`os.Chdir`).
- **Application layer** (`internal/app/*`):
  - Orchestrates use-cases with explicit dependencies:
    - `RepoRegistry` (global metadata store)
    - `RepoStoreFactory` (per-repo store open/close)
    - `Clock`, `Logger`
  - Implements invariants like “commits visible after push”.
- **Domain layer** (`internal/domain/*`):
  - Pure structs + rules for refs/commits/trees/index.
  - No IO, no JSON, no time.Now directly.
- **Infrastructure layer** (`internal/infra/*`):
  - `gitdb` implementation + file system + per-repo store + locking primitives.

### RepoStore & Global metadata relationship (prod-safe)

- **Global metadata store** should be authoritative for listing repos, but must support:
  - create/update/delete semantics with idempotency,
  - a repo “state” field: `active|missing|corrupt|migrating`,
  - migrations/versioning for schema changes (even if minimal).
- **Per-repo RepoStore** should be responsible for:
  - refs (`refs/heads/*`, `refs/remotes/*`),
  - objects (`objects/<id>`),
  - index (`index/entries/*`),
  - HEAD (`meta/HEAD`),
  - and should expose a transactional API whose atomicity is real (see plan).

### Modeling gitclone index/refs/objects

Recommended stable model:

- **Refs**:
  - `refs/heads/<branch>`: local tip
  - `refs/remotes/<remote>/<branch>`: remote tip (pushed view)
  - `HEAD`: points to current branch ref
- **Commit object**:
  - explicit `TreeID` (not overloaded with commit ID)
  - `Parents[]` array for merge support instead of `Parent`/`Parent2`
- **Tree objects**:
  - content-addressed objects (hash) preferred; numeric IDs acceptable but must be consistent.
- **Index**:
  - should support add/remove and should not require “scan entire log” for every read.

### Remote refs + “commit syns efter push”

Current behavior:
- `ListCommits` shows commits from `refs/remotes/origin/<branch>` (`gitClone/internal/app/commits/service.go`).

Target behavior options (pick one and enforce consistently):

1. **Explicit pushed view** (keep current):
   - `GET /commits` returns pushed commits only.
   - Add an explicit endpoint for local commits: `GET /commits?view=local`.
   - UI shows “local changes not pushed” clearly.
2. **Local view by default** (more git-like):
   - `GET /commits` reads `refs/heads/<branch>`.
   - A separate endpoint returns remote/pushed state.

Regardless of choice, define invariants and tests around:
- after `POST /commit`, local ref moves,
- after `POST /push`, remote ref matches local ref,
- after `POST /merge`, local ref moves and remote ref only moves after push.

## Proposed Folder Structure

Rough production-oriented structure (200–300 lines/file guideline; split by responsibility):

- `gitClone/`
  - `cmd/`
    - `gitserver/` (entry + config + shutdown)
    - `gitclone/` (CLI)
  - `internal/`
    - `transport/http/`
      - `router.go`
      - `middleware/` (cors, auth, request-id, logging)
      - `handlers/` (repos, branches, commits, merge, files, issues)
      - `dto/` (request/response types)
    - `app/`
      - `repos/` (use-cases)
      - `branches/`
      - `commits/`
      - `files/`
      - `issues/`
    - `domain/`
      - `repo.go`, `ref.go`, `commit.go`, `tree.go`, `index.go`
    - `infra/`
      - `repostore/` (open/close + locking + transactions)
      - `metadata/` (repo registry)
      - `gitdb/` (adapter)
      - `fs/` (filesystem helpers)

## Staged Refactor Plan

> This is a step-by-step plan. Each step has limited scope, explicit acceptance criteria, and test additions. No step introduces new product features; it focuses on correctness/safety.

### Step 0 — Production “guardrails” (no behavior change intended)

- **Scope**
  - Add health endpoints and server timeouts, plus graceful shutdown wiring.
  - Introduce request IDs and structured logging format (even if still `log.Printf`).
- **Files touched**
  - `gitClone/cmd/gitserver/main.go`
  - `gitClone/internal/transport/http/router.go`
  - new: `gitClone/internal/transport/http/middleware/*`
  - new: `gitClone/internal/transport/http/handlers/health.go`
- **Acceptance criteria**
  - `GET /healthz` returns 200.
  - Server has non-zero ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout.
  - Shutdown on SIGTERM finishes in bounded time.
- **Tests**
  - HTTP handler tests for health endpoints.

### Step 1 — Remove `os.Chdir` from request paths (deployment blocker)

- **Scope**
  - Eliminate all uses of process-global CWD changes in server codepaths.
  - Replace `internal/commands/*` usage in server with service-layer methods that accept explicit repo paths/store handles.
- **Files touched**
  - `gitClone/internal/transport/http/handlers_repos.go` (create repo)
  - `gitClone/internal/transport/http/handlers_merge.go` (merge)
  - `gitClone/internal/app/files/service.go` (staging)
  - Potentially: `gitClone/internal/commands/*` (server should stop calling these)
- **Acceptance criteria**
  - Concurrent requests to different repos do not interfere (add a concurrency test).
  - No server code uses `os.Chdir`.
- **Tests**
  - New test: parallel `POST /commit` or `POST /merge` on two repos concurrently and validate each repo’s refs changed correctly.

### Step 2 — Fix Repo ID and file-path sandboxing (security)

- **Scope**
  - Centralize repoID validation for all endpoints (reject `..`, path separators, percent-encoded traversal).
  - Sandbox file paths for `POST /files` to prevent escaping repo root.
- **Files touched**
  - `gitClone/internal/app/repos/resolver.go`
  - `gitClone/internal/transport/http/handlers_*` (ensure they call the same validation)
  - `gitClone/internal/app/files/service.go`
- **Acceptance criteria**
  - `POST /api/repos/<id>/files` with `../` paths fails with 400.
  - Repo ID traversal attempts fail with 400/404 consistently.
- **Tests**
  - Add table-driven tests for invalid repo IDs and invalid file paths.

### Step 3 — Make gitDb safe for concurrent use (reliability)

- **Scope**
  - Add locking around DB operations and/or ensure single-writer semantics per DB path.
  - Add file locking if multiple processes are expected.
- **Files touched**
  - `gitDb/db.go`, `gitDb/index.go`
  - `gitClone/internal/infra/storage/repo_store.go` (ensure correct locking strategy)
- **Acceptance criteria**
  - `go test -race ./...` passes for gitDb + gitClone under representative concurrency tests.
  - No corrupted log under parallel puts.
- **Tests**
  - Multi-goroutine put/get tests.
  - Multi-handle append correctness tests (extend `gitDb/gitdb_durability_test.go`).

### Step 4 — Replace WriteBatch with real atomic semantics

- **Scope**
  - Either:
    - implement MVCC-style “staged writes” + commit marker and have readers respect commit markers, or
    - implement single-record “batch” append that is atomically decoded and applied, or
    - move to a known-good embedded store (BoltDB/Badger) behind an interface.
- **Files touched**
  - `gitClone/internal/infra/storage/write_batch.go`
  - `gitClone/internal/storage/repo_store_wrappers.go` (batch usage)
- **Acceptance criteria**
  - Crash simulation: partial batch never becomes visible after restart.
  - No tx marker collisions.
- **Tests**
  - Fault-injection style tests: simulate crash between writes (requires hooks).

### Step 5 — Compaction / bounded growth

- **Scope**
  - Add log compaction to gitDb and/or periodic snapshotting.
  - Define retention and vacuum cadence.
- **Files touched**
  - `gitDb/*`
  - `gitClone/internal/metadata/store.go` (if schema changes)
- **Acceptance criteria**
  - Startup time bounded under large datasets.
  - Storage usage does not grow without bound for typical workflows.
- **Tests**
  - Create many records, compact, verify correctness and reduced size.

## Deployment Checklist

### Configuration / env vars

- **Backend**
  - `PORT` (default 8080)
  - `GITSTORE_REPO_BASE` (must be a persistent volume path in prod)
  - `GITSTORE_DB_PATH` (must be a persistent volume path in prod)
- **Frontend**
  - `VITE_API_URL` (prod should point to API base; dev can use proxy)
  - `VITE_FIREBASE_*` variables (`Client/src/firebase.ts`)
- **CLI**
  - `GITSTORE_API_URL` (defaults to `http://localhost:8080`) in `cli/src/cli/utils/gitstore-api.ts`

### Build commands

- **Backend**
  - `cd gitClone && go build -o gitserver ./cmd/gitserver`
- **Client**
  - `cd Client && npm ci && npm run build`
- **CLI**
  - `cd cli && npm ci && npm run build`

### Data persistence (volumes)

- Persist these paths:
  - `${GITSTORE_REPO_BASE}` (repo directories + `.gitclone/db/log`)
  - `${GITSTORE_DB_PATH}` (global metadata GitDb log)
- Ensure the deploy artifact does **not** ship with `gitClone/data/` contents from source control.

### Networking / CORS

- Restrict allowed origins in prod (do not use `*`) — `gitClone/internal/transport/http/router.go`.
- Plan for auth headers (`Authorization`) when backend starts verifying Firebase ID tokens.

### Health checks

- Add and use:
  - **liveness**: `/healthz` (process up)
  - **readiness**: `/readyz` (can open metadata DB, can access repo base)

### Operational requirements

- Run backend with:
  - timeouts,
  - graceful shutdown,
  - structured logs,
  - rotation / log aggregation.
- Add alerts on:
  - error rates,
  - storage log growth rate,
  - startup/rebuild times for gitDb,
  - disk full conditions.


