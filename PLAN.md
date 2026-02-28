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
- [ ] `internal/truenas/vm.go` (3b) — create VM, update VM
- [x] `tools/vm.go` (3a) — `list_vms`, `get_vm`, `start_vm`, `stop_vm`, `restart_vm`
- [ ] `tools/vm.go` (3b) — `create_vm`, `update_vm`
- [ ] `tools/destructive.go` — opt-in `delete_vm`
- [x] Update README (3a)

**Tools delivered** (~8):
`list_vms`, `get_vm`, `create_vm`, `update_vm`, `start_vm`, `stop_vm`, `restart_vm`, `delete_vm` (destructive)

---

### Phase 4 — ZFS Snapshots

**Goal**: Create, list, roll back, and delete snapshots — useful for pre-backup snapshotting.

**Tasks**:
- [ ] `internal/truenas/snapshot.go` — list, create, rollback, delete
- [ ] `tools/snapshot.go` — `list_snapshots`, `create_snapshot`, `rollback_snapshot`
- [ ] `tools/destructive.go` — opt-in `delete_snapshot`
- [ ] Update README

**Tools delivered** (~4):
`list_snapshots`, `create_snapshot`, `rollback_snapshot`, `delete_snapshot` (destructive)

---

### Phase 5 — TrueNAS Apps

**Goal**: Manage catalog apps (Helm-based) — useful for deploying containerised workloads
like PBS if a native app becomes available.

**Tasks**:
- [ ] `internal/truenas/app.go` — list apps, get app, install app, start/stop app, delete app
- [ ] `tools/app.go` — `list_apps`, `get_app`, `install_app`, `start_app`, `stop_app`
- [ ] `tools/destructive.go` — opt-in `delete_app`
- [ ] Update README

**Tools delivered** (~6):
`list_apps`, `get_app`, `install_app`, `start_app`, `stop_app`, `delete_app` (destructive)

---

### Phase 6 — PBS Provisioning Workflow (the target use case)

**Goal**: End-to-end automated PBS setup driven by an AI agent across both MCPs.

This isn't a new tool — it's a validation that the existing tools compose correctly to
accomplish the full workflow:

1. `create_dataset` (truenas-mcp) — create `tank/proxmox-backups`
2. `create_vm` (truenas-mcp) — Debian 12 VM, 2 vCPU, 4GB RAM, 32GB disk, NIC on LAN
3. `start_vm` (truenas-mcp) — boot the VM
4. *(manual step)* — install PBS inside the guest via SSH / console
5. `add_storage` (proxmox-mcp — future tool) — register PBS as backup target
6. `create_backup_job` (proxmox-mcp — future tool) — schedule all VMs nightly

Document this workflow in README as a worked example.

---

## Tool Count by Phase

| Phase | New Tools | Running Total |
|---|---|---|
| 1 — Foundation | 2 | 2 |
| 2 — Pools & Datasets | 5 | 7 |
| 3 — VMs | 8 | 15 |
| 4 — Snapshots | 4 | 19 |
| 5 — Apps | 6 | 25 |
| 6 — PBS Workflow | 0 (validation) | 25 |

---

## Security Rules (same as proxmox-mcp)

- No credentials in source — env vars only
- `TRUENAS_INSECURE=true` is the only way to skip TLS — off by default
- `TRUENAS_ALLOW_DESTRUCTIVE=true` required to register delete tools
- All `//nolint:` directives require explicit user approval + inline justification
- `gosec` warnings treated as bugs
