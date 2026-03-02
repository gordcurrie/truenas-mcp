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
| PR — Network, Filesystem & Zvol support | ✅ Complete — 37 tools shipped |
| 9 — WebSocket API Migration | ⏳ Before TrueNAS v26.04 (~mid-2026) |
| 10 — NFS Share Management | 🔜 Next (Proxmox backup workflow) |

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

## PR — Network, Filesystem & Zvol support ✅

**Goal**: Add tools needed to configure a PBS VM — find the bridge interface for NIC
attachment, check what ISOs exist in a directory, and create zvol block devices for VM
disks.

### New tools (2)

| Tool | API endpoint | Description |
|---|---|---|
| `list_interfaces` | `GET /interface` | List all network interfaces (bridges, physical, bonds, VLANs) |
| `list_directory` | `POST /filesystem/listdir` | List the contents of a directory on the TrueNAS host |

### Enhanced tools (1)

| Tool | Change |
|---|---|
| `create_dataset` | Added `volsize` parameter (required for `type=VOLUME` zvols) |

**Tool count:** 35 + 2 = **37 tools**

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
