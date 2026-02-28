# truenas-mcp

An MCP server that exposes [TrueNAS SCALE](https://www.truenas.com/truenas-scale/) operations as MCP tools, written in Go.

## Setup

### Prerequisites

- Go 1.26+
- A TrueNAS SCALE instance
- An API key (create one in TrueNAS UI under **Credentials → API Keys**)

### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TRUENAS_API_URL` | yes | e.g. `https://truenas.local/api/v2.0` |
| `TRUENAS_API_KEY` | yes | API key from TrueNAS UI |
| `TRUENAS_INSECURE` | no | `true` to skip TLS verification (self-signed certs) |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | `true` to enable destructive tools (default: disabled) |

### Running

```bash
# stdio (default — for use with MCP clients like Claude Desktop / VS Code Copilot)
TRUENAS_API_URL=https://truenas.local/api/v2.0 \
TRUENAS_API_KEY=your-api-key \
TRUENAS_INSECURE=true \
./bin/truenas-mcp

# HTTP transport
./bin/truenas-mcp --transport http --addr localhost:8081
```

### VS Code / Claude Desktop config

```json
{
  "mcpServers": {
    "truenas": {
      "command": "/path/to/truenas-mcp",
      "env": {
        "TRUENAS_API_URL": "https://truenas.local/api/v2.0",
        "TRUENAS_API_KEY": "your-api-key",
        "TRUENAS_INSECURE": "true"
      }
    }
  }
}
```

## Tools

### System

| Tool | Description | Parameters |
|---|---|---|
| `get_system_info` | Get TrueNAS SCALE system info: version, hostname, CPU, memory, uptime, load | _(none)_ |

### Storage

| Tool | Description | Parameters |
|---|---|---|
| `list_pools` | List all ZFS storage pools and their status, size, and health | _(none)_ |
| `get_pool` | Get detailed information about a specific ZFS pool | `id` (int) |

## Development

```bash
make install-tools   # install gofumpt, gosec, govulncheck, golangci-lint
make check           # run all quality gates (fmt, vet, lint, sec, vulncheck, test, build)
make build           # build binary to bin/truenas-mcp
make test            # run tests with race detector
```
