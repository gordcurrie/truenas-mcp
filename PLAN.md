# TrueNAS SCALE MCP Server — Project Plan

## Summary

Go MCP server exposing TrueNAS SCALE operations as tools. Uses `modelcontextprotocol/go-sdk`,
a custom `net/http` REST client (no third-party TrueNAS library), API key auth, and strict
linting via `golangci-lint`/`gosec`/`govulncheck`. Primary goal: allow AI agents to configure
TrueNAS SCALE as part of the Proxmox backup workflow.

## Plan Status

| Phase | Status |
|---|---|
| 1–8 (Foundation through CI & Releases) | ✅ Complete — 32 tools shipped |
| PR — VM Device Management | ✅ Complete — 35 tools shipped |
| PR — Safety & Quality Fixes | 🔜 Immediate next (no new tools) |
| 9 — WebSocket API Migration | ⏳ Before TrueNAS v26.04 (~mid-2026) |
| 10 — NFS Share Management | 🔜 After safety PR (Proxmox backup workflow) |

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Go SDK, same as proxmox-mcp |
| TrueNAS client | Custom `net/http` wrapper | Full control, no external dep |
| Auth | API key (`Authorization: Bearer <key>`) | Stateless, no session tokens |
| Transports | stdio (default) + HTTP (`--transport=http`) | stdio for local clients, HTTP for remote |
| Formatter | `gofumpt` | Stricter than `gofmt` |
| TLS | Verify on by default; `TRUENAS_INSECURE=true` to skip | TrueNAS commonly uses self-signed certs |
| Destructive tools | Opt-in via `TRUENAS_ALLOW_DESTRUCTIVE=true` | Safe by default |

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TRUENAS_API_URL` | yes | e.g. `https://truenas.local/api/v2.0` |
| `TRUENAS_API_KEY` | yes | API key from TrueNAS UI → Credentials → API Keys |
| `TRUENAS_INSECURE` | no | `true` to skip TLS verification |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | `true` to register destructive tools |

## CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | `stdio` or `http` |
| `--addr` | `localhost:8081` | Listen address when `--transport=http` |

## Makefile Targets

`make fix` → `make fmt` → `make vet` → `make lint` → `make sec` → `make vulncheck` →
`make test` → `make build` — all run by `make check`.

## TrueNAS API Notes

- Base URL: `https://<host>/api/v2.0`
- Responses are direct JSON — no `{"data": ...}` envelope (unlike Proxmox)
- Auth header: `Authorization: Bearer <api-key>`
- Long-running operations return a job ID — poll `/core/get_jobs?id=<id>` for completion
- API docs: `https://<host>/api/docs/` on any TrueNAS SCALE instance

---

## PR — VM Device Management ✅

**Goal**: Allow agents to attach and manage hardware devices on TrueNAS SCALE VMs — disks,
CDROMs (ISO files), NICs, and displays — so a VM can be fully configured via MCP without
using the TrueNAS web UI.

### New tools (3)

| Tool | API endpoint | Description |
|---|---|---|
| `list_vm_devices` | `GET /vm/device` | List all devices on a VM, filtered by VM ID |
| `add_vm_device` | `POST /vm/device` | Attach a DISK, CDROM, NIC, or DISPLAY device |
| `delete_vm_device` | `DELETE /vm/device/id/{id}` | Remove a device by its device ID |

### Note — `install_custom_app` visibility

`install_custom_app` is registered in the server and works correctly. If it does not appear
in Copilot's available tools, do a **full VS Code restart** (not just MCP server restart).
Copilot caches the tool list per-connection; a stale connection from before the tool was
added to the binary will not see it until the cache is cleared.

**Tool count:** 32 + 3 = **35 tools**

---

## PR — Safety & Quality Fixes

**Goal**: Address critical safety, correctness, and maintainability issues identified in the
MCP expert review before any new feature work begins. No new tools are added; no public
interfaces change.

> **Why 6/10?** The architecture and Go idioms are genuinely strong. The score reflects
> one real data-safety gap (`rollback_snapshot` can destroy datasets through an agent with
> no confirmation), a timer leak in the job poller, misleading security-audit comments that
> undermine the `//nolint` review trail, and missing input validation on integer IDs. These
> are not cosmetic — they are the difference between a server you can trust to run
> unsupervised and one you cannot. Fix these and the baseline comfortably reaches 8+.

---

### Critical — Must Fix

#### C1: `rollback_snapshot` is missing the destructive gate

`rollback_snapshot` destroys all ZFS data written after the target snapshot point. It is at
least as destructive as `delete_snapshot`, which already requires all three safety layers.
Currently `rollback_snapshot` sits in `registerSnapshotTools`, has no `confirmed bool`
guard, no `DestructiveHint`, and is registered even when `TRUENAS_ALLOW_DESTRUCTIVE=false`.
An LLM agent can call it silently in a default configuration.

**Tasks:**
- [ ] Move `rollback_snapshot` out of `tools/snapshot.go` into `tools/destructive.go`
- [ ] Add `Confirmed bool` field (`jsonschema:"required,..."`) to `rollbackSnapshotInput`
- [ ] Return early with an error when `!p.Confirmed`
- [ ] Add `DestructiveHint: &destructiveHint` to the tool annotations
- [ ] Update README destructive table and PLAN.md tool count if reclassification changes the numbers

#### C2: Stale Proxmox copy-paste in `client.go` comments

Three places in `internal/truenas/client.go` reference the Proxmox project this was forked
from. This directly undermines the audit trail for `//nolint` directives — a reviewer
checking the `gosec` annotation will see the wrong env var name and reasonably conclude the
justification was not reviewed for this codebase.

**Tasks:**
- [ ] `//nolint:gosec` on `InsecureSkipVerify`: change `PROXMOX_INSECURE` → `TRUENAS_INSECURE`
- [ ] `//nolint:gosec` on `httpClient.Do`: change `PROXMOX_API_URL` → `TRUENAS_API_URL`
- [ ] `post()` doc comment: remove the sentence referencing the Proxmox `{"data": ...}` envelope

#### C3: Timer leak in `PollJob`

`time.After` in a loop allocates a new timer every iteration. The timer has a 2-second fuse
and is not garbage-collected until it fires. If the loop exits early (context cancelled, job
succeeded, job failed), the current-iteration timer leaks until it fires. Under concurrent
job polling this accumulates.

**Tasks:**
- [ ] Replace `time.After(jobPollInterval)` with a `time.NewTimer` created once before the
  loop, stopped via `defer timer.Stop()`, and reset at the top of each iteration using
  `timer.Reset(jobPollInterval)` (drain the channel first if needed to avoid stale tick)
- [ ] Update `jobs_test.go` if the timing behaviour changes

#### C4: Zero-value integer IDs are not validated

`get_vm`, `get_pool`, `list_vm_devices`, and `add_vm_device` accept `id int`. Go's zero
value is `0`; if an LLM omits the field or passes `null`, the server silently issues a real
HTTP request to `/vm/id/0`. No TrueNAS resource ever has ID 0.

**Tasks:**
- [ ] Add `if p.ID <= 0 { return nil, nil, errors.New("...: id must be a positive integer") }`
  at the top of each affected handler in `tools/vm.go` and `tools/pool.go`
- [ ] Apply the same guard in `tools/destructive.go` for `deleteVMInput` and `deleteVMDeviceInput`

---

### Recommended — High Value

#### R1: `list_snapshots` fetches all snapshots then filters in Go

A large NAS instance may have thousands of snapshots. The full list is transferred over the
wire and decoded before the dataset filter is applied. TrueNAS v2 supports server-side query
filters.

**Tasks:**
- [ ] When `dataset != ""`, request
  `GET /pool/snapshot?query-filters=[["dataset","=","<dataset>"]]` instead of fetching all
- [ ] Update `ListSnapshots` in `internal/truenas/snapshot.go`
- [ ] Update `snapshot_test.go`

#### R2: HTTP graceful shutdown has no timeout

`httpServer.Shutdown(context.Background())` in `cmd/truenas-mcp/main.go` waits forever for
active connections to drain if a client holds a streaming connection open.

**Tasks:**
- [ ] Replace `context.Background()` with `context.WithTimeout(context.Background(), 10*time.Second)`
- [ ] `defer cancel()` the returned cancel function

#### R3: `update_vm` memory validation is effectively a no-op

The guard `params.Memory < 0` never triggers in practice because `Memory int` with
`omitempty` means 0 (the zero value) is never serialised. The validation looks correct but
does not protect against the actual edge case.

**Tasks:**
- [ ] The correct rule is: if Memory is provided it must be > 0; since 0 means "unchanged"
  (omitted by `omitempty`), only negative values need to be rejected — add a comment
  explaining this clearly so the next reader does not remove it as dead code
- [ ] Confirm the positive guard in `CreateVM` is intentional and add a matching comment

#### R4: Required string inputs in tool handlers have no blank-check

Tools such as `get_app`, `get_dataset`, `create_snapshot`, `get_snapshot`, and
`list_directory` accept required string fields with no empty-string guard. Blank input
produces a confusing 404 or opaque API error from TrueNAS rather than a clear MCP tool error.

**Tasks:**
- [ ] Add `if p.Name == "" { return nil, nil, errors.New("...: name must not be empty") }`
  (or equivalent for `id`, `path`, `dataset`) at the top of each affected tool handler
- [ ] Add one table-driven test case per handler that passes the zero-value and expects an error

#### R5: Verify `jsonschema` tag format produces correct schema

Tags like `` jsonschema:"required,Numeric VM ID" `` mix a constraint keyword and a
natural-language description in a single unkeyed string. Confirm what `google/jsonschema-go`
actually emits and whether `required` is honoured as a constraint or treated as description text.

**Tasks:**
- [ ] Write a short test that marshals the JSON schema for a typed input struct with a
  `required` int field and asserts the output contains `"required"` as a schema keyword
- [ ] If not emitted correctly: update affected tags to the format the library supports
- [ ] If emitted correctly: add a comment in `copilot-instructions.md` confirming the format

---

### Minor — Clean Up

#### M1: HTTP server factory comment

The `func(*http.Request) *mcp.Server { return server }` factory in `main.go` always returns
the same instance. Add a one-line comment stating this is intentional for a stateless tool
server without per-session resources, so a future maintainer does not "fix" it.

#### M2: `add_vm_device` dtype is not validated before the API call

`dtype` is passed as a free string and cast to `VMDeviceType`. An invalid value like `"USB"`
is forwarded to TrueNAS, which returns an opaque 422. Validate against the known constants
(`DISK`, `CDROM`, `NIC`, `DISPLAY`, `RAW`) in the tool handler before calling the client.

---

### Full Checklist

- [x] C1 — `rollback_snapshot` moved to destructive, `confirmed` gate added
- [x] C2 — Proxmox copy-paste comments corrected in `client.go`
- [x] C3 — `PollJob` timer leak fixed with `time.NewTimer`
- [x] C4 — Zero-value ID guards added in VM and pool tool handlers
- [x] R1 — `list_snapshots` uses server-side query filter when dataset is specified
- [x] R2 — HTTP shutdown uses a 10 s timeout context
- [x] R3 — `update_vm` memory validation clarified with comments
- [x] R4 — Required string inputs validated in tool handlers
- [x] R5 — `jsonschema` tag format verified: entire value = description; required is inferred from absence of `omitempty`; all redundant `required,` prefixes stripped from descriptions
- [x] M1 — HTTP factory comment added
- [x] M2 — `add_vm_device` dtype validated against known constants
- [x] `make check` passes
- [x] README and PLAN.md tool counts updated if reclassification occurs

---

## Phase 9 — WebSocket API Migration (before TrueNAS v26.04)

**Background**: TrueNAS 25.10 deprecated `/api/v2.0`. It will be removed in v26.04
(~mid-2026). The replacement is JSON-RPC 2.0 over WebSocket at `wss://{host}/api/current`.

Only `internal/truenas/client.go` needs a full rewrite. All other files keep their method
signatures — only the transport changes.

### Wire format

```json
// Auth (first call after connect)
{"jsonrpc": "2.0", "id": 1, "method": "auth.login_with_api_key", "params": ["<key>"]}

// Request
{"jsonrpc": "2.0", "id": 2, "method": "app.query", "params": []}

// Response
{"jsonrpc": "2.0", "id": 2, "result": [...]}
```

REST → JSON-RPC method mapping: `GET /app` → `app.query`, `POST /app` → `app.create`,
`DELETE /app/id/{name}` → `app.delete` (params `["name"]`), `GET /pool/dataset` →
`pool.dataset.query`, `POST /system/info` → `system.info`.

### Design decisions

| Decision | Choice | Rationale |
|---|---|---|
| WebSocket library | `nhooyr.io/websocket` | Context-aware, no gorilla dep, actively maintained |
| Multiplexing | `sync.Map` of pending `chan response` keyed by atomic uint64 ID | Responses arrive out of order |
| Connection lifecycle | Persistent single connection with reconnect | Avoids per-call handshake overhead |
| Job polling | Unchanged — `core.get_jobs` still works | No change needed in `jobs.go` |

### Tasks

- [ ] `go get nhooyr.io/websocket`, run `make vulncheck`
- [ ] Rewrite `internal/truenas/client.go` — `Dial`, `call`, read loop goroutine, `Close`
- [ ] Update `client_test.go` with a local WebSocket test server
- [ ] Audit each method in `app.go`, `vm.go`, `pool.go`, `dataset.go`, `system.go` —
  replace REST calls with `call(ctx, "service.method", params, &result)`
- [ ] Update all `*_test.go` files
- [ ] `make check` passes
- [ ] Smoke-test against live TrueNAS 25.10+; confirm no deprecation warnings

---

## Phase 10 — NFS Share Management (target: next)

**Goal**: Allow agents to configure NFS exports on TrueNAS — needed as the NFS fallback
backup target for Proxmox, and a useful complement to the PBS Docker Compose approach.

### Backup workflows enabled (with proxmox-mcp Phase 7)

**Option A — PBS Docker (preferred)**:
1. `create_dataset` *(exists)* — create the PBS datastore path (e.g. `tank/pbs-datastore`)
2. `install_custom_app` *(exists)* — deploy PBS Docker Compose on TrueNAS, mounting the dataset
3. proxmox-mcp `add_storage` *(Phase 7)* — register `type=pbs` pointing at TrueNAS IP:8007
4. proxmox-mcp `create_backup` *(exists)* — run backups

**Option B — NFS**:
1. `create_dataset` *(exists)* — create backup dataset
2. `create_nfs_share` *(this phase)* — export it, restricted to Proxmox node IPs
3. proxmox-mcp `add_storage` *(Phase 7)* — register `type=nfs` pointing at TrueNAS

### PR — NFS shares (4 new tools)

New file `internal/truenas/nfs.go`. New file `tools/nfs.go`. `RegisterAll` gains
`registerNFSTools`.

TrueNAS API: `GET /sharing/nfs`, `GET /sharing/nfs/id/{id}`, `POST /sharing/nfs`,
`DELETE /sharing/nfs/id/{id}`.

| Tool | API endpoint | Params |
|---|---|---|
| `list_nfs_shares` | `GET /sharing/nfs` | — |
| `get_nfs_share` | `GET /sharing/nfs/id/{id}` | `id` (int) |
| `create_nfs_share` | `POST /sharing/nfs` | `path` (required), `comment` (optional), `hosts` (optional list of allowed IPs/CIDRs), `maproot_user` (optional), `maproot_group` (optional), `readonly` (optional bool) |
| `delete_nfs_share` | `DELETE /sharing/nfs/id/{id}` | `id` (int), `confirmed: true` — destructive opt-in |

`list_nfs_shares` and `get_nfs_share` get `ReadOnlyHint: true`. `delete_nfs_share` follows
the 3-layer safety pattern (`TRUENAS_ALLOW_DESTRUCTIVE` + `confirmed: true` + `DestructiveHint`).

Tests: success + notFound for list/get; success + apiError for create; success + notFound
for delete (8 new tests).

**Phase 10 target tool count:** 32 + 4 = **36 tools** (34 always-on + up to 4 destructive opt-in).

---

## Security Rules

- No credentials in source — env vars only
- `TRUENAS_INSECURE=true` is the only way to skip TLS — off by default
- `TRUENAS_ALLOW_DESTRUCTIVE=true` required to register delete tools
- All `//nolint:` directives require explicit user approval + inline justification
- `gosec` warnings treated as bugs
