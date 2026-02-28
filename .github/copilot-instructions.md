# Copilot Instructions — truenas-mcp

This is a Go MCP server that exposes TrueNAS SCALE operations as MCP tools.
See `PLAN.md` for the full project plan, decisions, and implementation order.

## Repository

- **GitHub**: `git@github.com:gordcurrie/truenas-mcp.git`
- **Owner**: `gordcurrie` | **Repo**: `truenas-mcp`
- Use owner `gordcurrie` and repo `truenas-mcp` for all GitHub MCP tool calls (PRs, issues, etc.)

## Stack

- **Language**: Go (latest stable)
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk/mcp`
- **TrueNAS client**: custom `net/http` wrapper in `internal/truenas/` — no third-party TrueNAS library
- **Auth**: TrueNAS API keys only (`Authorization: Bearer <api-key>`)
- **Transports**: stdio (default) and HTTP (streamable), selected via `--transport` flag

## Code Style — Idiomatic Go

- Wrap errors: `fmt.Errorf("doing X: %w", err)` — never discard errors
- Sentinel errors: `var ErrNotFound = errors.New("resource not found")`
- `context.Context` is always the first parameter on any function that does I/O
- Pointer receivers on `Client` and mutable types; value receivers on small read-only structs
- No `init()` functions anywhere — explicit initialization only
- No global mutable state — inject dependencies
- All exported types and functions must have doc comments
- `json` tags on all API types; `jsonschema` tags on MCP tool input structs
- Table-driven tests using `t.Run` subtests
- Use `gofumpt` formatting (stricter than `gofmt`)

## Documentation — Required for Every PR

Every PR that adds or changes tools **must** update both files before committing:

1. **`README.md`** — add each new tool to the appropriate table (System, Storage, VMs,
   Snapshots, Apps, or Destructive). Include the tool name, a one-line description, and its
   parameters. Destructive tools go in the Destructive table.

2. **`PLAN.md`** — mark the PR's phase tasks as completed (add ✅) and update the running
   tool count at the bottom of the phase section.

Do not skip these updates under any circumstances — documentation is part of the definition
of done for every PR, the same as passing `make check`.

## PR Size

Keep PRs under **500 lines of code** (excluding tests and docs). If a phase is large, split
it into multiple PRs — e.g. pools first, then datasets.

## Git Commits

Always write multi-line commit messages via a temp file to avoid shell quoting issues:

```bash
python3 -c "open('/tmp/msg.txt','w').write('''subject line\n\nbody line 1\nbody line 2\n''')"
git add . && git commit -F /tmp/msg.txt
```

Never pass multi-line messages with `-m` — the shell mangles them.

## Git Push Rules

- Use plain `git push` for normal commits (new commits on top of the branch).
- Use `git push --force-with-lease` **only** when history has been rewritten (rebase, amend, etc.).
- **Never** use `git push --force` under any circumstances without explicit user permission.

## Quality Gates — `make check` must pass before every commit

```
make fix        # go fix ./...
make fmt        # gofumpt -w .
make vet        # go vet ./...
make lint       # golangci-lint run ./...
make sec        # gosec ./...
make vulncheck  # govulncheck ./...
make test       # go test -race -count=1 ./...
make build      # go build ./cmd/truenas-mcp/
```

## Linting

Config is in `.golangci.yml`. Key linters: `gosec`, `govet`, `staticcheck`, `errcheck`,
`bodyclose`, `noctx`, `gofumpt`, `revive`, `gocritic`, `unparam`, `unconvert`.

When `InsecureSkipVerify` is needed (TrueNAS self-signed TLS), annotate with:
```go
//nolint:gosec // G402: InsecureSkipVerify is only set when TRUENAS_INSECURE=true,
// which the user must explicitly opt into. Default is secure (verify enabled).
```

## Package Layout

- `internal/truenas/` — TrueNAS API client and types. No MCP imports here.
- `tools/` — MCP tool registration. Imports both `mcp` and `internal/truenas`.
- `cmd/truenas-mcp/` — Entrypoint only. Reads env/flags, wires things together, runs the server.

## TrueNAS API Conventions

- Base URL: `https://<host>/api/v2.0`
- Responses are **direct JSON** — no `{"data": ...}` envelope (unlike Proxmox).
- Long-running operations return a job object with an `id` field. Poll
  `GET /core/get_jobs?id=<id>` until `state` is `"SUCCESS"` or `"FAILED"` via `jobs.go`.
- Lifecycle tools (`start_vm`, `stop_vm`, etc.) return job info immediately (non-blocking).
- API docs available at `https://<host>/api/docs/` on any TrueNAS SCALE instance.

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TRUENAS_API_URL` | yes | e.g. `https://truenas.local/api/v2.0` |
| `TRUENAS_API_KEY` | yes | API key from TrueNAS UI → Credentials → API Keys |
| `TRUENAS_INSECURE` | no | `true` to skip TLS verification (self-signed certs) |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | `true` to register delete tools (default: disabled) |

## Security Rules — Non-Negotiable

### Never Commit Secrets
- No API keys, passwords, or credentials anywhere in source files
- No hardcoded IPs, hostnames, or environment-specific values
- Secrets come from environment variables only — never from flags, config files, or defaults
- If a test needs credentials, use a clearly fake placeholder (e.g. `test-api-key`) and
  document that it is a test fixture, not a real value
- `.env` files are for local development only — always in `.gitignore`, never committed
- If you spot a secret in code or history, flag it immediately before doing anything else

### Security-Conscious Development
- Prefer the most restrictive option by default (TLS verification is ON unless explicitly
  disabled via `TRUENAS_INSECURE=true`)
- Validate and sanitize all inputs from MCP tool arguments before passing to the TrueNAS API
- Use `context.Context` with timeouts on all outbound HTTP calls — never hang indefinitely
- Keep dependencies minimal; every new dependency is a potential vulnerability surface
- Run `make vulncheck` after adding or updating any dependency
- Treat `gosec` warnings as bugs, not style suggestions

### Skipping Lint or Security Checks
- **Always ask the user before adding any `//nolint:` directive**
- When a `//nolint:` is genuinely required (e.g. the known `InsecureSkipVerify` opt-in),
  it must include the specific rule ID and a plain-English explanation of why it is safe:
  ```go
  //nolint:gosec // G402: InsecureSkipVerify is only set when TRUENAS_INSECURE=true,
  // which the user must explicitly opt into. Default is secure (verify enabled).
  ```
- Never suppress an entire linter for a file (`//nolint:gosec` at file scope) — always
  target the narrowest possible scope (single line)
- Never disable `errcheck` — if an error truly cannot be handled, document why with a comment
