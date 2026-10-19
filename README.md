# truenas-mcp

An MCP server that exposes [TrueNAS SCALE](https://www.truenas.com/truenas-scale/) operations as MCP tools, written in Go.

## Tools

### System

| Tool | Description | Parameters |
|---|---|---|
| `get_system_info` | Get TrueNAS SCALE system info: version, hostname, CPU, memory, uptime, load | _(none)_ |

### Network

| Tool | Description | Parameters |
|---|---|---|
| `list_interfaces` | List all network interfaces (bridges, physical, bonds, VLANs) — useful for finding bridge names when attaching VM NICs | _(none)_ |

### Filesystem

| Tool | Description | Parameters |
|---|---|---|
| `list_directory` | List the contents of a directory on the TrueNAS host filesystem | `path` (string, required) |

### Storage

| Tool | Description | Parameters |
|---|---|---|
| `list_pools` | List all ZFS storage pools and their status, size, and health | `limit`, `offset` (int, optional — server-side pagination) |
| `get_pool` | Get detailed information about a specific ZFS pool | `id` (int) |
| `list_datasets` | List ZFS datasets and zvols, optionally filtered by pool | `pool` (string, optional); `limit`, `offset` (int, optional — server-side pagination) |
| `get_dataset` | Get detailed information about a specific ZFS dataset or zvol by its full path | `id` (string, e.g. `Storage/backups`) |
| `create_dataset` | Create a new ZFS dataset or zvol | `name` (string, required); `type`, `compression`, `comments`, `quota`, `volsize` (optional — `volsize` in bytes is required when `type=VOLUME`) |

### Virtual Machines

| Tool | Description | Parameters |
|---|---|---|
| `list_vms` | List all VMs and their state (RUNNING/STOPPED), CPU, and memory | `limit`, `offset` (int, optional — server-side pagination) |
| `get_vm` | Get detailed information about a specific VM | `id` (int) |
| `start_vm` | Start a VM; waits for the job and returns its result | `id` (int) |
| `stop_vm` | Stop a VM; set `force=true` to forcibly terminate; waits for the job and returns its result | `id` (int); `force` (bool, optional) |
| `restart_vm` | Restart a VM; waits for the job and returns its result | `id` (int) |
| `create_vm` | Create a new VM; returns the created VM | `name`, `memory` (required); `vcpus`, `bootloader`, `autostart`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` (optional) |
| `update_vm` | Update an existing VM configuration; omitted fields are unchanged | `id` (required); any subset of `name`, `memory`, `vcpus`, `bootloader`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` |
| `list_vm_devices` | List all hardware devices attached to a VM (disks, CDROMs, NICs, displays) | `id` (int) |
| `add_vm_device` | Attach a hardware device to a VM (DISK, CDROM, NIC, or DISPLAY) | `vm_id` (int), `dtype` (string), `attributes` (object); `order` (optional) |

### Apps

| Tool | Description | Parameters |
|---|---|---|
| `list_apps` | List all apps managed by TrueNAS SCALE | `limit`, `offset` (int, optional — server-side pagination) |
| `get_app` | Get detailed information about a specific app by its app name | `name` (string) |
| `start_app` | Start an app; waits for the job and returns its result | `name` (string) |
| `stop_app` | Stop a running app; waits for the job and returns its result | `name` (string) |
| `restart_app` | Restart an app; waits for the job and returns its result | `name` (string) |
| `list_images` | List all Docker images stored on the TrueNAS SCALE system | _(none)_ |
| `install_app` | Install a catalog app from the TrueNAS app catalog; waits for the job and returns its result | `app_name` (string); `catalog_app` (string); `train` (string, default: stable); `version` (string, default: latest) |
| `install_custom_app` | Install a custom Docker Compose app; waits for the job and returns its result | `app_name` (string); `custom_compose_config_string` (string, YAML) |
| `upgrade_app` | Upgrade an app to a newer version; waits for the job and returns its result | `name` (string); `version` (string, optional — omit for latest) |
| `upgrade_summary` | Get upgrade availability and changelog for an app | `name` (string) |
| `rollback_app` | Roll an app back to a previous version; waits for the job and returns its result | `name` (string); `version` (string, required) |

Job tools (`start_vm`, `stop_vm`, `restart_vm` and the app lifecycle, install, upgrade and rollback tools) wait up to 10 minutes for the TrueNAS job to finish and return `{"result": ...}`. Older TrueNAS releases return the job ID as the result instead. If the wait times out, the job keeps running on TrueNAS; check `get_vm` or `get_app` before retrying.

### ZFS Snapshots

| Tool | Description | Parameters |
|---|---|---|
| `list_snapshots` | List ZFS snapshots, optionally filtered by dataset path | `dataset` (string, optional); `limit`, `offset` (int, optional — server-side pagination) |
| `get_snapshot` | Get detailed information about a specific snapshot | `id` (string, e.g. `Storage/backups@before-upgrade`) |
| `create_snapshot` | Create a new ZFS snapshot of a dataset | `dataset`, `name` (required); `recursive` (bool, optional) |

### Destructive (requires `TRUENAS_ALLOW_DESTRUCTIVE=true`)

| Tool | Description | Parameters |
|---|---|---|
| `delete_vm` | Permanently delete a stopped VM | `id` (int); `confirmed: true` (required) |
| `delete_app` | Permanently delete a TrueNAS app | `name` (string); `confirmed: true` (required) |
| `delete_snapshot` | Permanently delete a ZFS snapshot | `id` (string, e.g. `Storage/backups@before-upgrade`); `confirmed: true` (required) |
| `delete_vm_device` | Remove a hardware device from a VM by device ID | `id` (int); `confirmed: true` (required) |
| `rollback_snapshot` | Roll a dataset back to a previous snapshot — **all data written after the snapshot is permanently destroyed** | `id` (string); `confirmed: true` (required); `recursive`, `recursive_clones`, `force` (optional) |

### Tool annotations

Every tool advertises [MCP tool annotations](https://modelcontextprotocol.io/specification/2025-06-18/server/tools#tool-annotations) so clients can decide when to ask for confirmation:

| Annotation | Tools |
|---|---|
| `readOnlyHint: true` | All `get_*` and `list_*` tools, plus `upgrade_summary` |
| `destructiveHint: false` | Additive tools: `create_dataset`, `create_snapshot`, `create_vm`, `add_vm_device`, `install_app`, `install_custom_app`, `start_vm`, `start_app` |
| `destructiveHint: true` | Tools that overwrite or interrupt existing state: `update_vm`, `stop_vm`, `restart_vm`, `stop_app`, `restart_app`, `upgrade_app`, `rollback_app`, and every tool in the Destructive table |

## Installation

### Download a pre-built binary

Download the latest release for your platform from the [Releases](https://github.com/gordcurrie/truenas-mcp/releases) page.

| Platform | Binary |
|---|---|
| Linux (amd64) | `truenas-mcp_linux_amd64` |
| Linux (arm64) | `truenas-mcp_linux_arm64` |
| macOS (amd64) | `truenas-mcp_darwin_amd64` |
| macOS (arm64) | `truenas-mcp_darwin_arm64` |
| Windows (amd64) | `truenas-mcp_windows_amd64.exe` |

Each release also includes `checksums.txt` with SHA-256 hashes of every binary. Download it into the same directory and verify before running:

```bash
# Linux
sha256sum --check --ignore-missing checksums.txt
# macOS
shasum -a 256 --check --ignore-missing checksums.txt
```

Windows users can compare the output of `Get-FileHash truenas-mcp_windows_amd64.exe` (PowerShell) against the matching line.

Make it executable and place it on your `PATH` (substitute the filename for your platform):

```bash
chmod +x <binary-name>
mv <binary-name> /usr/local/bin/truenas-mcp
```

> Windows users: rename the `.exe` and add it to a directory on your `%PATH%`.

> macOS users: if Gatekeeper blocks the downloaded binary, clear the quarantine flag with `xattr -d com.apple.quarantine /usr/local/bin/truenas-mcp`.

### Install with Go

Requires Go 1.27.2+. Installs `truenas-mcp` into `$(go env GOPATH)/bin`:

```bash
go install github.com/gordcurrie/truenas-mcp/cmd/truenas-mcp@latest
```

### Build from source

Requires Go 1.27.2+. You will also need a TrueNAS SCALE instance and an API key — create one in the TrueNAS UI under **Credentials → API Keys**.

```bash
git clone https://github.com/gordcurrie/truenas-mcp
cd truenas-mcp
cp .env.example .env   # copy the example env file
$EDITOR .env           # set TRUENAS_* values (see table below)
make build             # binary lands in bin/truenas-mcp
```

## Configuration

All configuration is via environment variables:

| Variable | Required | Description |
|---|---|---|
| `TRUENAS_API_URL` | yes | e.g. `https://truenas.local` |
| `TRUENAS_API_KEY` | yes | API key from TrueNAS UI |
| `TRUENAS_INSECURE` | no | `true` to skip TLS verification (self-signed certs) |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | `true` to enable destructive tools (default: disabled) |

## MCP Client Setup

### Running

```bash
# stdio (default — for use with local MCP clients)
TRUENAS_API_URL=https://truenas.local \
TRUENAS_API_KEY=your-api-key \
./truenas-mcp

# HTTP transport (for remote or shared deployments)
./truenas-mcp --transport http --addr localhost:8081
```

### Claude Code

```bash
claude mcp add truenas --scope user \
  -e TRUENAS_API_URL=https://truenas.local \
  -e TRUENAS_API_KEY=your-api-key \
  -- /path/to/truenas-mcp
```

Run `/mcp` inside Claude Code to confirm the server is connected.

### VS Code Copilot

Create `.vscode/mcp.json` in your workspace (already gitignored):

```json
{
  "servers": {
    "truenas": {
      "type": "stdio",
      "command": "/path/to/truenas-mcp",
      "env": {
        "TRUENAS_API_URL": "https://truenas.local",
        "TRUENAS_API_KEY": "your-api-key"
      }
    }
  }
}
```

Then open the Copilot chat panel, switch to **Agent** mode, and the `truenas` server will appear in the available tools.

### Claude Desktop

Add the server to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "truenas": {
      "command": "/path/to/truenas-mcp",
      "env": {
        "TRUENAS_API_URL": "https://truenas.local",
        "TRUENAS_API_KEY": "your-api-key"
      }
    }
  }
}
```

Restart Claude Desktop after saving the config.

### OpenCode

Add the server to `opencode.json` in your project root (or `~/.config/opencode/opencode.json` for global config):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "truenas": {
      "type": "local",
      "command": ["/path/to/truenas-mcp"],
      "enabled": true,
      "environment": {
        "TRUENAS_API_URL": "https://truenas.local",
        "TRUENAS_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Agent skills

The [`skills/`](skills/) directory contains [Agent Skills](https://agentskills.io) (`SKILL.md` format) that teach an agent how to use these tools well:

| Skill | Use it for |
|---|---|
| [`truenas`](skills/truenas/SKILL.md) | Everyday tasks: pool and dataset usage, snapshots, app installs and upgrades, VM lifecycle and devices — with safe confirmation rules |
| [`audit-truenas`](skills/audit-truenas/SKILL.md) | A read-only health review (pools, capacity, snapshot coverage, app upgrade backlog, VMs) that produces a prioritised findings report |
| [`pbs-on-truenas`](skills/pbs-on-truenas/SKILL.md) | Set up Proxmox Backup Server in a TrueNAS VM and register it with Proxmox VE, together with [proxmox-mcp](https://github.com/gordcurrie/proxmox-mcp) |

Install them for all your projects in Claude Code:

```bash
git clone https://github.com/gordcurrie/truenas-mcp
mkdir -p ~/.claude/skills
cp -R truenas-mcp/skills/* ~/.claude/skills/
```

Other agents that support Agent Skills (GitHub Copilot, Cursor, Codex CLI, Gemini CLI, …) read the same folders; copy them to that agent's skills directory. Inside this repo they are already available to Claude Code and Copilot via `.claude/skills` and `.github/skills`, which link to `skills/`.

## Development

```bash
make install-tools   # install golangci-lint, gosec, govulncheck, gofumpt
make check           # full quality gate: fix, fmt, vet, lint, sec, vulncheck, test, build
make test            # tests only (with race detector)
make build           # build only → bin/truenas-mcp
make clean           # remove bin/truenas-mcp
```
