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
| `start_vm` | Start a VM; returns async job ID immediately | `id` (int) |
| `stop_vm` | Stop a VM; set `force=true` to forcibly terminate | `id` (int); `force` (bool, optional) |
| `restart_vm` | Restart a VM; returns async job ID immediately | `id` (int) |
| `create_vm` | Create a new VM; returns the created VM | `name`, `memory` (required); `vcpus`, `bootloader`, `autostart`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` (optional) |
| `update_vm` | Update an existing VM configuration; omitted fields are unchanged | `id` (required); any subset of `name`, `memory`, `vcpus`, `bootloader`, `cores`, `threads`, `cpu_mode`, `cpu_model`, `shutdown_timeout`, `description` |
| `list_vm_devices` | List all hardware devices attached to a VM (disks, CDROMs, NICs, displays) | `id` (int) |
| `add_vm_device` | Attach a hardware device to a VM (DISK, CDROM, NIC, or DISPLAY) | `vm_id` (int), `dtype` (string), `attributes` (object); `order` (optional) |

### Apps

| Tool | Description | Parameters |
|---|---|---|
| `list_apps` | List all apps managed by TrueNAS SCALE | `limit`, `offset` (int, optional — server-side pagination) |
| `get_app` | Get detailed information about a specific app by its app name | `name` (string) |
| `start_app` | Start an app; returns async job ID immediately | `name` (string) |
| `stop_app` | Stop a running app; returns async job ID immediately | `name` (string) |
| `restart_app` | Restart an app; returns async job ID immediately | `name` (string) |
| `list_images` | List all Docker images stored on the TrueNAS SCALE system | _(none)_ |
| `install_app` | Install a catalog app from the TrueNAS app catalog; returns async job ID | `app_name` (string); `catalog_app` (string); `train` (string, default: stable); `version` (string, default: latest) |
| `install_custom_app` | Install a custom Docker Compose app; returns async job ID | `app_name` (string); `custom_compose_config_string` (string, YAML) |
| `upgrade_app` | Upgrade an app to a newer version; returns async job ID immediately | `name` (string); `version` (string, optional — omit for latest) |
| `upgrade_summary` | Get upgrade availability and changelog for an app | `name` (string) |
| `rollback_app` | Roll an app back to a previous version; returns async job ID immediately | `name` (string); `version` (string, required) |

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

Make it executable and place it on your `PATH` (substitute the filename for your platform):

```bash
chmod +x <binary-name>
mv <binary-name> /usr/local/bin/truenas-mcp
```

> Windows users: rename the `.exe` and add it to a directory on your `%PATH%`.

### Build from source

Requires Go 1.26+. You will also need a TrueNAS SCALE instance and an API key — create one in the TrueNAS UI under **Credentials → API Keys**.

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
| `TRUENAS_API_URL` | yes | e.g. `https://truenas.local/api/v2.0` |
| `TRUENAS_API_KEY` | yes | API key from TrueNAS UI |
| `TRUENAS_INSECURE` | no | `true` to skip TLS verification (self-signed certs) |
| `TRUENAS_ALLOW_DESTRUCTIVE` | no | `true` to enable destructive tools (default: disabled) |

## MCP Client Setup

### Running

```bash
# stdio (default — for use with local MCP clients)
TRUENAS_API_URL=https://truenas.local/api/v2.0 \
TRUENAS_API_KEY=your-api-key \
./truenas-mcp

# HTTP transport (for remote or shared deployments)
./truenas-mcp --transport http --addr localhost:8081
```

### VS Code Copilot

Create `.vscode/mcp.json` in your workspace (already gitignored):

```json
{
  "servers": {
    "truenas": {
      "type": "stdio",
      "command": "/path/to/truenas-mcp",
      "env": {
        "TRUENAS_API_URL": "https://truenas.local/api/v2.0",
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
        "TRUENAS_API_URL": "https://truenas.local/api/v2.0",
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
        "TRUENAS_API_URL": "https://truenas.local/api/v2.0",
        "TRUENAS_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Worked Example — Proxmox Backup Server on TrueNAS

This example shows how an AI agent can use `truenas-mcp` (plus the companion
[proxmox-mcp](https://github.com/gordcurrie/proxmox-mcp)) to fully automate setting up
[Proxmox Backup Server](https://www.proxmox.com/en/proxmox-backup-server) on TrueNAS, then
register it with a Proxmox VE cluster and schedule nightly backups.

### Step 1 — Create a ZFS dataset for backup storage (truenas-mcp)

```
create_dataset
  name: "tank/proxmox-backups"
  comments: "Proxmox Backup Server datastore"
```

### Step 2 — Create a VM to run PBS (truenas-mcp)

```
create_vm
  name: "pbs"
  memory: 4096
  vcpus: 2
  bootloader: "UEFI"
  autostart: true
  description: "Proxmox Backup Server"
```

### Step 3 — Start the VM (truenas-mcp)

```
start_vm
  id: <vm_id returned by create_vm>
```

### Step 4 — Install PBS inside the VM _(manual)_

Attach a Debian 12 ISO, install the OS, then follow the
[PBS installation guide](https://pbs.proxmox.com/docs/installation.html) to install and
configure PBS. Add the `tank/proxmox-backups` dataset as a PBS datastore via NFS or
direct mount.

### Step 5 — Register PBS with Proxmox VE _(proxmox-mcp — future tool)_

```
add_storage          # not yet implemented in proxmox-mcp
  type: "pbs"
  server: "<PBS VM IP>"
  datastore: "proxmox-backups"
```

### Step 6 — Schedule nightly backups _(proxmox-mcp — future tool)_

```
create_backup_job    # not yet implemented in proxmox-mcp
  storage: "pbs"
  schedule: "daily"
  all_vms: true
```

Steps 1–4 are fully automated today. Steps 5–6 require future proxmox-mcp tools.

---

## Development

```bash
make install-tools   # install golangci-lint, gosec, govulncheck, gofumpt
make check           # full quality gate: fix, fmt, vet, lint, sec, vulncheck, test, build
make test            # tests only (with race detector)
make build           # build only → bin/truenas-mcp
make clean           # remove bin/truenas-mcp
```
