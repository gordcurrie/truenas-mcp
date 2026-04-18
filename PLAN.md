# TrueNAS SCALE MCP Server — Project Plan

## Summary

Build an MCP server in Go that exposes TrueNAS SCALE operations as MCP tools.
Uses the official `modelcontextprotocol/go-sdk`, a custom WebSocket JSON-RPC 2.0 client
(no third-party TrueNAS library), API key auth, and strict linting enforced via
`golangci-lint`, `gosec`, `govulncheck`, and `go fix`.

**Primary motivation**: Allow AI agents to configure TrueNAS SCALE as part of larger
cross-product workflows — the first target being provisioning a Proxmox Backup Server (PBS)
VM on TrueNAS SCALE automatically from the Proxmox MCP.

> **TrueNAS 26.0 note**: TrueNAS 26.0.0 dropped the HTTP REST API (`/api/v2.0`) entirely.
> The server now exposes a **JSON-RPC 2.0 over WebSocket** API at `wss://<host>/api/current`.
> The client was migrated in the "WS Migration" phase below.

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Go SDK, same as proxmox-mcp |
| TrueNAS client | Custom WebSocket JSON-RPC 2.0 client | Full control, required for TrueNAS 26.0+ |
| WS library | `github.com/gorilla/websocket` | Battle-tested, MIT, minimal footprint |
| Auth | API key via `auth.login_with_api_key` RPC call | TrueNAS 26.0 WS auth — no Bearer header |
| Transports | Both stdio + HTTP (flag-selected) | stdio for local clients, HTTP for remote/shared |
| Linter | `golangci-lint` + `gosec` + `govulncheck` | Security-first, idiomatic Go — same as proxmox-mcp |
| Formatter | `gofumpt` (stricter than `gofmt`) | Consistent style |
| TLS | `TRUENAS_INSECURE=true` opt-in for skip-verify | TrueNAS commonly uses self-signed certs on LAN |
| Error handling | Always wrap with `fmt.Errorf("doing X: %w", err)` | Idiomatic, stack-traceable |
| Global state | None — client injected, no `init()` | Testable, explicit |

## TrueNAS SCALE API Notes

- **TrueNAS 26.0+**: WebSocket JSON-RPC 2.0 at `wss://<host>/api/current`
- `TRUENAS_API_URL` is the base URL, e.g. `https://nas.example.com` — the client derives
  `wss://nas.example.com/api/current` automatically
- Auth: connect, then call `auth.login_with_api_key` with the API key as the single param
- Request format: `{"jsonrpc":"2.0","id":N,"method":"service.method","params":[...]}`
- Query methods accept `[filters, options]` params — e.g. `[["name","=","tank"]], {"limit":10}]`
- Long-running operations return a job ID — call `core.get_jobs` with `[["id","=",N]]` filter
  to poll until `state` reaches `SUCCESS`, `FAILED`, or `ABORTED`
- API docs: `https://api.truenas.com/v26.0/`
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
| `TRUENAS_API_URL` | yes | Base URL only, e.g. `https://truenas.local` (no path — client derives `wss://…/api/current`) |
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

## Migration Plan — WebSocket JSON-RPC 2.0 (TrueNAS 26.0)

TrueNAS 26.0 dropped the HTTP REST API entirely. The entire transport layer must be replaced
with a WebSocket JSON-RPC 2.0 client. All 25 existing tools must continue to work without
any changes to the `tools/` layer.

---

### PR 1 — New WebSocket client + transport layer

**Goal**: Replace `net/http` transport with a persistent WebSocket JSON-RPC 2.0 connection.

**Tasks**:
- ✅ Add `github.com/gorilla/websocket` to `go.mod`
- ✅ Rewrite `internal/truenas/client.go` — persistent WS connection, `auth.login_with_api_key`,
  per-request channel routing via `pending map[int64]chan rpcResponse`, `readLoop` goroutine, `Close()`
- ✅ Rewrite `internal/truenas/query.go` — replace `buildQueryString` with `buildQueryParams`
  returning a JSON-RPC `[]any{filters, options}` params array; keep `ListOptions` and `validateListOptions`
- ✅ Update `internal/truenas/jobs.go` — replace HTTP GET with `c.call(ctx, "core.get_jobs", ...)`
- ✅ Update `internal/truenas/system.go` — `system.info` RPC
- ✅ Update `internal/truenas/pool.go` — `pool.query`, `pool.get_instance`
- ✅ Update `internal/truenas/dataset.go` — `pool.dataset.query`, `pool.dataset.get_instance`, `pool.dataset.create`
- ✅ Update `internal/truenas/snapshot.go` — `pool.snapshot.query`, `pool.snapshot.get_instance`,
  `pool.snapshot.create`, `pool.snapshot.rollback`, `pool.snapshot.delete`
- ✅ Update `internal/truenas/vm.go` — `vm.query`, `vm.get_instance`, `vm.start`, `vm.stop`,
  `vm.restart`, `vm.create`, `vm.update`, `vm.delete`, `vm.device.query`, `vm.device.create`, `vm.device.delete`
- ✅ Update `internal/truenas/app.go` — `app.query`, `app.get_instance`, `app.start`, `app.stop`,
  `app.redeploy`, `app.create`, `app.delete`, `app.upgrade`, `app.rollback`, `app.upgrade_summary`,
  `app.image.query`
- ✅ Update `internal/truenas/network.go` — `interface.query`
- ✅ Update `internal/truenas/filesystem.go` — `filesystem.listdir`
- ✅ Update `cmd/truenas-mcp/main.go` — derive `wss://<host>/api/current` from base URL,
  call `client.Connect(ctx)` on startup; remove `/api/v2.0` path expectation
- ✅ Rewrite `internal/truenas/client_test.go` — replace `httptest.Server` with a test WS server
- ✅ Update README — `TRUENAS_API_URL` is base URL only (e.g. `https://truenas.local`)
- ✅ `make check` passes

**Key design decisions**:
- Single persistent connection with transparent one-shot reconnect on disconnect
- `NewClient(host, apiKey string, insecure bool) (*Client, error)` — validates configuration eagerly (non-empty host/key, normalises URL); network connection is still lazy until `Connect(ctx)`
- `Connect(ctx context.Context) error` — dials, authenticates, starts `readLoop`
- `call(ctx, method string, params, result any) error` — registers pending channel, sends request, waits
- Error mapping: RPC error code `-32001` → inspect message/errname → `ErrNotFound` or `*APIError`
- `APIError.StatusCode` carries the JSON-RPC error code (not an HTTP status)

---

## Security Rules

- No credentials in source — env vars only
- `TRUENAS_INSECURE=true` is the only way to skip TLS — off by default
- `TRUENAS_ALLOW_DESTRUCTIVE=true` required to register delete tools
- All `//nolint:` directives require explicit user approval + inline justification
- `gosec` warnings treated as bugs
