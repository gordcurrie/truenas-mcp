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
| `list_datasets` | List ZFS datasets and zvols, optionally filtered by pool | `pool` (string, optional) |
| `get_dataset` | Get detailed information about a specific ZFS dataset or zvol by its full path | `id` (string, e.g. `Storage/backups`) |
| `create_dataset` | Create a new ZFS dataset or zvol | `name` (string, required); `type`, `compression`, `comments`, `quota` (optional) |

### Virtual Machines

| Tool | Description | Parameters |
|---|---|---|
| `list_vms` | List all VMs and their state (RUNNING/STOPPED), CPU, and memory | _(none)_ |
| `get_vm` | Get detailed information about a specific VM | `id` (int) |
| `start_vm` | Start a VM; returns async job ID immediately | `id` (int) |
| `stop_vm` | Stop a VM; set `force=true` to forcibly terminate | `id` (int); `force` (bool, optional) |
| `restart_vm` | Restart a VM; returns async job ID immediately | `id` (int) |
| `create_vm` | Create a new VM; returns the created VM | `name`, `memory` (required); `vcpus`, `bootloader`, `autostart`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` (optional) |
| `update_vm` | Update an existing VM configuration; omitted fields are unchanged | `id` (required); any subset of `name`, `memory`, `vcpus`, `bootloader`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` |

### Docker Containers

| Tool | Description | Parameters |
|---|---|---|
| `list_containers` | List all Docker containers managed by TrueNAS SCALE | _(none)_ |
| `get_container` | Get detailed information about a specific container by its app name | `name` (string) |
| `start_container` | Start a container; returns async job ID immediately | `name` (string) |
| `stop_container` | Stop a running container; returns async job ID immediately | `name` (string) |
| `restart_container` | Restart a container; returns async job ID immediately | `name` (string) |
| `list_images` | List all Docker images stored on the TrueNAS SCALE system | _(none)_ |

### Destructive (requires `TRUENAS_ALLOW_DESTRUCTIVE=true`)

| Tool | Description | Parameters |
|---|---|---|
| `delete_vm` | Permanently delete a stopped VM | `id` (int); `confirmed: true` (required) |

## Development

```bash
make install-tools   # install gofumpt, gosec, govulncheck, golangci-lint
make check           # run all quality gates (fmt, vet, lint, sec, vulncheck, test, build)
make build           # build binary to bin/truenas-mcp
make test            # run tests with race detector
```
