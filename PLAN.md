# TrueNAS SCALE MCP Server — Project Plan

## Summary

Build an MCP server in Go that exposes TrueNAS SCALE operations as MCP tools.
Uses the official `modelcontextprotocol/go-sdk`, a custom TrueNAS REST API HTTP client
(no third-party TrueNAS library), API key auth, and strict linting enforced via
`golangci-lint`, `gosec`, `govulncheck`, and `go fix`.

**Primary motivation**: Allow AI agents to configure TrueNAS SCALE as part of larger
cross-product workflows — the first target being provisioning a Proxmox Backup Server (PBS)
VM on TrueNAS SCALE automatically from the Proxmox MCP.

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Go SDK, same as proxmox-mcp |
| TrueNAS client | Custom `net/http` wrapper | Full control, no external dep, easy to extend |
| Auth | API key (`Authorization: Bearer <key>`) | Stateless, no session tokens, matches TrueNAS REST auth |
| Transports | Both stdio + HTTP (flag-selected) | stdio for local clients, HTTP for remote/shared |
| Linter | `golangci-lint` + `gosec` + `govulncheck` | Security-first, idiomatic Go — same as proxmox-mcp |
| Formatter | `gofumpt` (stricter than `gofmt`) | Consistent style |
| TLS | `TRUENAS_INSECURE=true` opt-in for skip-verify | TrueNAS commonly uses self-signed certs on LAN |
| Error handling | Always wrap with `fmt.Errorf("doing X: %w", err)` | Idiomatic, stack-traceable |
| Global state | None — client injected, no `init()` | Testable, explicit |

## TrueNAS SCALE API Notes

- Base URL: `https://<host>/api/v2.0`
- All endpoints return JSON directly (no `{"data": ...}` wrapper, unlike Proxmox)
- Auth header: `Authorization: Bearer <api-key>`
- Long-running operations return a job ID (`{"id": 123}`) — poll `/core/get_jobs?id=<id>`
  for completion, same pattern as Proxmox UPIDs
- API docs available at `https://<host>/api/docs/` on any TrueNAS SCALE instance
- API key is created in the TrueNAS UI under Credentials → API Keys

## Project Structure

```
truenas_mcp/
├── cmd/
│   └── truenas-mcp/
│       └── main.go               # entrypoint, CLI flags, transport selection
├── internal/
│   └── truenas/
│       ├── client.go             # custom HTTP client (auth, TLS, base URL)
│       ├── client_test.go        # httptest-based unit tests
│       ├── types.go              # shared TrueNAS response/request structs
│       ├── types_test.go
│       ├── system.go             # system info API calls
│       ├── system_test.go
│       ├── pool.go               # storage pool API calls
│       ├── pool_test.go
│       ├── dataset.go            # dataset/zvol API calls
│       ├── dataset_test.go
│       ├── snapshot.go           # ZFS snapshot API calls
│       ├── snapshot_test.go
│       ├── vm.go                 # VM API calls
│       ├── vm_test.go
│       ├── app.go                # TrueNAS Apps (Docker/catalog) API calls
│       ├── app_test.go
│       └── jobs.go               # async job polling (/core/get_jobs)
│           jobs_test.go
├── tools/
│   ├── register.go               # RegisterAll(cfg Config) wires all tools onto the MCP server
│   ├── helpers.go                # shared tool helpers
│   ├── helpers_test.go
│   ├── system.go                 # system info MCP tools
│   ├── pool.go                   # storage pool MCP tools
│   ├── dataset.go                # dataset MCP tools
│   ├── snapshot.go               # snapshot MCP tools
│   ├── vm.go                     # VM MCP tools
│   ├── app.go                    # TrueNAS Apps MCP tools
│   └── destructive.go            # delete_dataset, delete_vm, delete_snapshot (opt-in)
├── .golangci.yml                 # linter config (copy from proxmox-mcp, same rules)
├── Makefile                      # quality gate targets (same targets as proxmox-mcp)
├── go.mod
├── go.sum
├── README.md
└── PLAN.md                       # this file
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TRUENAS_API_URL` | yes | Base URL, e.g. `https://truenas.local/api/v2.0` |
| `TRUENAS_API_KEY` | yes | API key created in TrueNAS UI under Credentials → API Keys |
| `TRUENAS_INSECURE` | no | Set `true` to skip TLS verification (self-signed certs) |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | Set `true` to register `delete_dataset`, `delete_vm`, `delete_snapshot` (default: disabled) |

## CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | `stdio` or `http` |
| `--addr` | `localhost:8081` | Listen address when `--transport=http` |

## Makefile Targets

| Target | What it does |
|---|---|
| `make fix` | `go fix ./...` |
| `make fmt` | `gofumpt -w .` |
| `make vet` | `go vet ./...` |
| `make lint` | `golangci-lint run ./...` |
| `make sec` | `gosec ./...` |
| `make vulncheck` | `govulncheck ./...` |
| `make test` | `go test -race -count=1 ./...` |
| `make build` | `go build ./cmd/truenas-mcp/` |
| `make check` | runs all of the above in order |

## Implementation Phases

---

### Phase 1 — Foundation (bootstrapping + system info)

**Goal**: Working binary, client, auth, and basic read-only tools.

**Tasks**:
- [ ] `go mod init`, add MCP SDK dependency
- [ ] Copy `.golangci.yml` and `Makefile` from proxmox-mcp (same rules)
- [ ] Implement `internal/truenas/client.go` — HTTP client, API key auth, TLS opt-out, context timeouts
- [ ] Implement `internal/truenas/jobs.go` — poll `/core/get_jobs` until job completes
- [ ] Implement `internal/truenas/system.go` — `/system/info`, `/system/version`
- [ ] Implement `tools/system.go` — `get_system_info` tool
- [ ] Implement `cmd/truenas-mcp/main.go` — flags, env, wire-up, stdio + HTTP transports
- [ ] `make check` passes

**Tools delivered** (~2):
`get_system_info`, `get_system_version`

---

### Phase 2 — Storage: Pools & Datasets

**Goal**: Read and create ZFS datasets — the core of the PBS provisioning workflow.

**Tasks**:
- [x] `internal/truenas/pool.go` — list pools, get pool status
- [x] `internal/truenas/dataset.go` — list datasets, get dataset, create dataset, set dataset properties (quota, compression, etc.)
- [x] `tools/pool.go` — `list_pools`, `get_pool`
- [x] `tools/dataset.go` — `list_datasets`, `get_dataset`, `create_dataset`
- [x] Update README

**Tools delivered** (~5):
`list_pools`, `get_pool`, `list_datasets`, `get_dataset`, `create_dataset`

---

### Phase 3 — VMs

**Goal**: Create, configure, start, stop, and delete VMs — needed for spinning up a PBS VM.

**Tasks**:
- [x] `internal/truenas/jobs.go` — async job type + PollJob helper
- [x] `internal/truenas/vm.go` (3a) — list VMs, get VM, start/stop/restart VM
- [x] `internal/truenas/vm.go` (3b) — create VM, update VM, delete VM
- [x] `tools/vm.go` (3a) — `list_vms`, `get_vm`, `start_vm`, `stop_vm`, `restart_vm`
- [x] `tools/vm.go` (3b) — `create_vm`, `update_vm`
- [x] `tools/destructive.go` — opt-in `delete_vm`
- [x] Update README (3a + 3b)

**Tools delivered** (~8):
`list_vms`, `get_vm`, `create_vm`, `update_vm`, `start_vm`, `stop_vm`, `restart_vm`, `delete_vm` (destructive)

---

### Phase 4 — TrueNAS Apps (Docker Containers)

**Goal**: Manage TrueNAS Apps — catalog-based and custom Docker Compose applications.

TrueNAS SCALE 25.10 removed the legacy `/container/container` service. Apps are now
managed exclusively via the `/app` endpoint. Catalog apps install from the TrueNAS app
catalog (`catalog_app` + `train` + `version`); custom apps accept a raw Docker Compose
config (`custom_compose_config_string`).

#### Phase 4a — Read + Lifecycle (✅ PR #5)

> **Note**: files and tool names originally used `container` terminology; renamed to `app`
> in Phase 6 once it became clear the `/app` endpoint is the apps API, not a container API.
> The experimental container UI in TrueNAS 25.10 will get its own endpoints in a future release.

**Tasks**:
- [x] `internal/truenas/container.go` → `app.go` — list apps, get app,
  start/stop/restart app; list images
- [x] `internal/truenas/container_test.go` → `app_test.go`
- [x] `tools/container.go` → `app.go` — `list_apps`, `get_app`,
  `start_app`, `stop_app`, `restart_app`, `list_images`
- [x] Update README

**Tools delivered (6)**:
`list_apps`, `get_app`, `start_app`, `stop_app`, `restart_app`, `list_images`

#### Phase 4b — Create + Delete (✅ PR #6)

**Tasks**:
- [x] `internal/truenas/app.go` — `CreateApp` (catalog + custom compose),
  `DeleteApp`
- [x] `internal/truenas/app_test.go`
- [x] `tools/app.go` — `install_app` (catalog), `install_custom_app`
- [x] `tools/destructive.go` — opt-in `delete_app`
- [x] Update README

**Tools delivered (3)**:
`install_app`, `install_custom_app`, `delete_app` (destructive)

---

### Phase 5 — ZFS Snapshots (✅ PR #7)

**Goal**: Create, list, roll back, and delete snapshots — useful for pre-backup snapshotting.

**Tasks**:
- [x] `internal/truenas/snapshot.go` — list, create, rollback, delete
- [x] `internal/truenas/snapshot_test.go`
- [x] `tools/snapshot.go` — `list_snapshots`, `get_snapshot`, `create_snapshot`, `rollback_snapshot`
- [x] `tools/destructive.go` — opt-in `delete_snapshot`
- [x] Update README

**Tools delivered (5)**:
`list_snapshots`, `get_snapshot`, `create_snapshot`, `rollback_snapshot`, `delete_snapshot` (destructive)

---

### Phase 6 — Apps Rename + Upgrade/Rollback

**Goal**: Rename all `container` terminology to `app` (the `/app` endpoint is the TrueNAS
apps API — the experimental container UI will get its own endpoints in a future TrueNAS
release). Add the genuinely new operations not covered in Phase 4: upgrade and rollback.

**Tasks**:
- [x] `git mv internal/truenas/container.go internal/truenas/app.go` — rename file
- [x] `git mv internal/truenas/container_test.go internal/truenas/app_test.go` — rename file
- [x] `git mv tools/container.go tools/app.go` — rename file
- [x] Rename all types/functions: `Container`→`App`, `ListContainers`→`ListApps`,
  `GetContainer`→`GetApp`, `StartContainer`→`StartApp`, `StopContainer`→`StopApp`,
  `RestartContainer`→`RestartApp`, `CreateContainerParams`→`CreateAppParams`,
  `CreateContainer`→`CreateApp`, `DeleteContainer`→`DeleteApp`
- [x] Rename all MCP tool names: `list_containers`→`list_apps`, `get_container`→`get_app`,
  `start_container`→`start_app`, `stop_container`→`stop_app`,
  `restart_container`→`restart_app`, `create_container`→`install_app`,
  `create_custom_container`→`install_custom_app`, `delete_container`→`delete_app`
- [x] `internal/truenas/app.go` — add `UpgradeApp`, `RollbackApp`, `UpgradeSummary`
- [x] `tools/app.go` — add `upgrade_app`, `rollback_app`, `upgrade_summary`
- [x] Update `tools/register.go`: `registerContainerTools` → `registerAppTools`
- [x] Update README
- [x] Update PLAN.md

**Tools delivered** (~3 net new, 8 renamed):
`upgrade_app`, `rollback_app`, `upgrade_summary`

---

### Phase 7 — PBS Provisioning Workflow ✅

**Goal**: End-to-end automated PBS setup driven by an AI agent across both MCPs.

This isn't a new tool — it's a validation that the existing tools compose correctly to
accomplish the full workflow:

1. `create_dataset` (truenas-mcp) — create `tank/proxmox-backups`
2. `create_vm` or `install_app` (truenas-mcp) — Debian 12 VM or lightweight app
3. `start_vm` / `start_app` (truenas-mcp) — boot the instance
4. *(manual step)* — install PBS inside the guest via SSH / console
5. `add_storage` (proxmox-mcp — future tool) — register PBS as backup target
6. `create_backup_job` (proxmox-mcp — future tool) — schedule all VMs nightly

**Tasks**:
- [x] Document this workflow in README as a worked example

---

## Tool Count by Phase

| Phase | New Tools | Running Total |
|---|---|---|
| 1 — Foundation | 2 | 2 |
| 2 — Pools & Datasets | 5 | 7 |
| 3 — VMs | 8 | 15 |
| 4 — Docker Containers | 9 | 24 |
| 5 — Snapshots | 5 | 29 |
| 6 — Apps Rename + Upgrade/Rollback | 3 | 32 |
| 7 — PBS Workflow | 0 (validation) | 32 |
| 8 — CI & Releases | 0 | 32 |

---

## Phase 8 — CI & Releases (both repos)

Prerequisite: all feature phases merged and repos made public.

### Goals

Make both `truenas-mcp` and `proxmox-mcp` consumable by non-Go users with zero build
tooling — just download a binary and configure `mcp.json`.

### GitHub Actions Workflows (per repo)

**`.github/workflows/ci.yml`** — runs on every push and PR to `main`:
- `make check` (fix, fmt, vet, lint, sec, vulncheck, test -race, build)

**`.github/workflows/release.yml`** — runs on `v*` tag push:
- Cross-compile for: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Upload binaries to GitHub Release (created automatically by the workflow)
- Binary naming convention: `<name>_<os>_<arch>[.exe]`

### README Updates (per repo)

- Add **Installation** section: download binary from Releases page
- Add **VS Code `mcp.json`** usage snippet with env var instructions
- Add **Building from source** section for Go users

### Definition of Done

- [x] `ci.yml` passes on `main` for both repos
- [x] `release.yml` produces binaries on a `v*` tag for both repos
- [ ] README installation section complete for both repos

---

## Phase 9 — WebSocket API Migration (target: before v26.04)

**Background**: TrueNAS 25.10 deprecated the REST API (`/api/v2.0`). It will be removed
in v26.04 (estimated mid-2026). The replacement is JSON-RPC 2.0 over WebSocket at
`wss://{host}/api/current`. Confirmed live on 2026-02-28 against TrueNAS 25.10.2.1.

**Reference**: https://nas.1og.me/api/docs/current/jsonrpc.html

### Impact

Only `internal/truenas/client.go` needs a full rewrite. All other files
  (`app.go`, `vm.go`, `pool.go`, `dataset.go`, `system.go`, `jobs.go`) keep their
method signatures and logic — only the underlying transport changes.

Method names map 1-to-1 from REST path to JSON-RPC method, e.g.:
- `GET /app` → `app.query`
- `POST /app` → `app.create`
- `POST /app/start` (body `"name"`) → `app.start` (params `["name"]`)
- `DELETE /app/id/{name}` → `app.delete` (params `["name"]`)
- `GET /pool/dataset` → `pool.dataset.query`
- `POST /system/info` → `system.info`

### Wire format

Every call is a single WebSocket message:

```json
{"jsonrpc": "2.0", "id": 1, "method": "app.query", "params": []}
```

Response:

```json
{"jsonrpc": "2.0", "id": 1, "result": [...]}
```

Auth is done once after connect via:

```json
{"jsonrpc": "2.0", "id": 1, "method": "auth.login_with_api_key", "params": ["<api-key>"]}
```

### Design decisions

| Decision | Choice | Rationale |
|---|---|---|
| WebSocket library | `nhooyr.io/websocket` | Context-aware, no gorilla dependency, actively maintained |
| Multiplexing | `sync.Map` of pending `chan response` keyed by request ID | WebSocket is async; responses arrive out of order |
| Request ID | Atomic `uint64` counter | Simple, unique, no coordination needed |
| Connection lifecycle | Persistent single connection with reconnect | Avoids per-call handshake overhead |
| Job polling | Unchanged — `core.get_jobs` still works, same int job ID | No change needed in `jobs.go` |
| Insecure TLS | Same `TRUENAS_INSECURE=true` env var, passed to `nhooyr.io/websocket` dial options | Consistent with current behaviour |

### Tasks

- [ ] Add `nhooyr.io/websocket` dependency (`go get`, run `make vulncheck`)
- [ ] Rewrite `internal/truenas/client.go`:
  - `Dial(ctx, url, apiKey, insecure)` — connects WebSocket, authenticates, starts read loop
  - `call(ctx, method string, params, result any) error` — marshals request, registers pending channel, waits for response
  - read loop goroutine — decodes incoming messages, routes to pending channels
  - graceful `Close()` method
- [ ] Update `client_test.go` — replace `httptest.Server` with a local WebSocket test server
- [ ] Audit each method in `container.go`, `vm.go`, `pool.go`, `dataset.go`, `system.go`:
  - Replace `c.get(ctx, "/path", &result)` → `c.call(ctx, "service.method", params, &result)`
  - Replace `c.post(ctx, "/path", &result)` → `c.call(ctx, "service.method", []any{...}, &result)`
  - Etc. for `postWithBody`, `put`, `delete`
- [ ] Update all `*_test.go` files to use WebSocket test server
- [ ] `make check` passes
- [ ] Smoke-test against live TrueNAS; confirm no deprecation warnings

### Definition of Done

- [ ] No more REST deprecation warnings from TrueNAS UI
- [ ] All existing tools work identically from the MCP client's perspective
- [ ] `make check` passes
- [ ] Committed and pushed before TrueNAS 26.04 is released

---

## Security Rules (same as proxmox-mcp)

- No credentials in source — env vars only
- `TRUENAS_INSECURE=true` is the only way to skip TLS — off by default
- `TRUENAS_ALLOW_DESTRUCTIVE=true` required to register delete tools
- All `//nolint:` directives require explicit user approval + inline justification
- `gosec` warnings treated as bugs
