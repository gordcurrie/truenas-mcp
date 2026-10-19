---
name: truenas
description: >
  Manage and inspect a TrueNAS SCALE system using the truenas-mcp tools. Use
  this skill when asked about TrueNAS pools, ZFS datasets or zvols, disk space,
  snapshots, apps (Docker / catalog apps such as Jellyfin or Immich), app
  upgrades, virtual machines, VM devices, network interfaces, or the TrueNAS
  host filesystem — for example "how much space is left on my NAS", "snapshot
  my photos dataset", "upgrade Jellyfin", "is my VM running", or "attach an ISO
  to the VM".
license: MIT
---

# Skill: TrueNAS SCALE (via truenas-mcp)

This skill tells you how to use the tools exposed by
[gordcurrie/truenas-mcp](https://github.com/gordcurrie/truenas-mcp) safely and
efficiently. For a full health review, use the `audit-truenas` skill instead.

---

## Before you start

The tools are only available if the truenas-mcp server is configured in your
MCP client. If no `list_pools` / `get_system_info` tools are available, stop
and point the user at the setup section of the
[README](https://github.com/gordcurrie/truenas-mcp#installation). Do not try to
reach the TrueNAS API directly with curl unless the user asks you to.

If a call fails with a TLS or certificate error, TrueNAS is most likely using
its default self-signed certificate: the user needs `TRUENAS_INSECURE=true`.
If it fails with `401`/`403`, the API key is wrong or was revoked.

---

## Ground rules

1. **Read before you write.** Look up the current state with a `list_*` or
   `get_*` tool before changing anything, and show the user what will change.
2. **Use real identifiers.** Each resource has its own key. Resolve it from
   the matching `list_*` result. Never guess it.
   - Pools and VMs: numeric `id` (`get_pool`, `get_vm`, `start_vm`, …).
   - VM devices: numeric device `id` from `list_vm_devices`.
   - Datasets and zvols: full path including the pool, e.g. `Storage/immich`.
   - Snapshots: `dataset@name`, e.g. `Storage/immich@before-upgrade`.
   - Apps: the app `name` from `list_apps`.
3. **Every change needs the user's approval.** Get explicit approval for each
   mutation (create, update, start, stop, restart, install, upgrade, rollback,
   add device, delete) before calling the tool. A previous approval does not
   carry over to a different change. Only the destructive tools (see below)
   declare a `confirmed` parameter: set it to `true` only after that approval.
   Do not send `confirmed` to any other tool.
4. **Get complete lists.** `list_pools`, `list_datasets`, `list_snapshots`,
   `list_vms` and `list_apps` return a plain JSON array with no total count.
   Omit `limit` (or pass `0`) to get every item. If you do pass a `limit`,
   keep calling with `offset` increased by the number of items returned until
   a page comes back shorter than `limit`. Do not answer "how many" or "is X
   present" questions from a partial page.
5. **Async jobs.** `start_vm`, `restart_vm`, all app lifecycle tools,
   `install_app`, `install_custom_app`, `upgrade_app` and `rollback_app` return
   `{"job_id": N}` immediately. There is no job-status tool. Check the outcome
   by re-reading the resource (`get_vm` → `status.state`, `get_app` → `state`)
   after a short wait, and tell the user the job was submitted, not finished.

---

## Common tasks

### How much space is left?

1. `list_pools`. Each pool has `size`, `allocated` and `free` in **bytes**,
   plus `status`, `healthy`, `warning` and `fragmentation` (a percentage as a
   string). Compute used % as `allocated / size`.
2. `list_datasets` (optionally `pool`) for per-dataset usage. `used`,
   `available` and `quota` are human-readable strings such as `"1.08 TiB"`;
   `quota` is `null` when unset.
3. `list_datasets` returns each dataset at the top level **and** nested inside
   its parent's `children`. De-duplicate by `id` before you count or sum
   anything.

### Create a dataset or zvol

- `create_dataset` with `name` as the full path (`Storage/backups`). `type` is
  `FILESYSTEM` (default) or `VOLUME`. A `VOLUME` (zvol) needs `volsize` in
  bytes. `quota` is in bytes; `0` means no quota.
- Check with `get_dataset` or `list_datasets` that the name is free first.

### Snapshots

- `list_snapshots` with `dataset` to filter. Results are **not** sorted by
  time and have no timestamp field. `createtxg` (a numeric string) increases
  over time: sort by it numerically to find the newest. Automatic snapshots
  usually carry their date in `snapshot_name`, e.g.
  `auto-2026-10-17_00-00 - Daily`.
- `create_snapshot` takes `dataset` and `name` (the part after `@`). Set
  `recursive: true` to include child datasets.
- **Take a snapshot before risky changes.** Before an app upgrade or rollback,
  a VM reconfiguration, or bulk file changes, offer to snapshot the affected
  dataset(s) first, named e.g. `before-<change>-<YYYY-MM-DD>`.
- Rolling a dataset back to a snapshot (`rollback_snapshot`) permanently
  destroys everything written after it. It is a destructive tool (see below).

### Apps

- `list_apps` returns each app's `name` and `state` (for example `RUNNING`,
  `STOPPED`, `DEPLOYING`, `CRASHED`). `get_app` returns the same for one app.
- `start_app`, `stop_app` and `restart_app` take the app `name`. Stopping or
  restarting an app interrupts its users; say so before asking for approval.
- `list_images` lists the Docker images on the host.

#### Upgrade an app

1. Always call `upgrade_summary` first. It returns `upgrade_available`,
   `latest_version`, `latest_human_version`, `available_versions_for_upgrade`
   and, when TrueNAS has one, `changelog`.
2. If `upgrade_available` is `false`, stop and say so.
3. Show the user the current and target versions and the changelog (or say
   there is none). Offer a snapshot of the app's data dataset(s) first.
4. After approval, call `upgrade_app` with `name` and, if the user picked one,
   `version` (from `available_versions_for_upgrade[].version`, not the human
   version). Omit `version` for the latest.
5. Confirm with `get_app` that the app returns to `RUNNING`.

#### Roll an app back

- `rollback_app` needs `name` and `version`; both are required. Use a
  `version` value the app previously ran. The tools cannot list an app's
  previous versions, so ask the user if you do not know it.
- `rollback_app` reverts the app's version, not its data. If the newer version
  migrated the app's database, the old version may not start. A dataset
  snapshot taken before the upgrade is the only safe way back for the data.

#### Install an app

- `install_app` installs a catalog app: `app_name` (lowercase, hyphens
  allowed, max 40 characters), `catalog_app` (e.g. `jellyfin`), optional
  `train` (default `stable`) and `version` (default `latest`). It cannot set
  app configuration values; the app installs with catalog defaults.
- `install_custom_app` deploys a Docker Compose app from
  `custom_compose_config_string` (YAML). Show the user the full YAML before
  asking for approval.

### Virtual machines

- `list_vms` / `get_vm` return the VM config and `status.state` (`RUNNING` or
  `STOPPED`). `memory` is in MiB.
- `start_vm` and `restart_vm` return a job ID. `stop_vm` does a clean
  shutdown; `force: true` kills the VM like pulling the power cord. Only use
  `force` when the user asks for it or a clean stop has failed.
- `create_vm` needs `name` and `memory` (MiB). VM names must be
  **alphanumeric only**: no hyphens, underscores or spaces. `bootloader` is
  `UEFI` (default), `UEFI_CSM` or `GRUB`. `cpu_mode` is `CUSTOM`,
  `HOST-MODEL` or `HOST-PASSTHROUGH`.
- `update_vm` changes only the fields you send. It cannot clear a field to
  an empty string or `0`. Most changes take effect on the next boot.

#### VM devices

- `list_vm_devices` with the VM `id`. Each device's type is in `dtype`
  (`DISK`, `CDROM`, `NIC`, `DISPLAY`, `RAW`).
- DISPLAY devices include a `password` attribute. Do not repeat it to the user
  unless they ask for it.
- `add_vm_device` takes `vm_id`, `dtype` and `attributes`. Do **not** put
  `dtype` inside `attributes`; the tool rejects that. Examples:
  - DISK: `{"path": "/dev/zvol/Storage/myvm-disk", "type": "VIRTIO"}` — create
    the zvol first with `create_dataset` `type: VOLUME`.
  - CDROM: `{"path": "/mnt/Storage/isos/debian.iso"}` — find the file with
    `list_directory`.
  - NIC: `{"type": "VIRTIO", "nic_attach": "br0"}` — find the bridge or
    interface name with `list_interfaces`.
  - DISPLAY: `{"type": "SPICE", "web": true}`.
- Device changes take effect on the next boot of the VM.

### Host information

- `get_system_info`: version, hostname, `physmem` (bytes), `cores`,
  `loadavg`, `uptime_seconds`.
- `list_interfaces`: physical NICs, bridges, bonds and VLANs with their
  addresses.
- `list_directory` with an absolute `path` (e.g. `/mnt/Storage/isos`).

---

## Destructive tools

These tools are only registered when the server runs with
`TRUENAS_ALLOW_DESTRUCTIVE=true`:

| Tool | Effect |
|---|---|
| `delete_vm` | Permanently delete a VM (stop it first) |
| `delete_app` | Permanently delete an app (stop it first) |
| `delete_snapshot` | Permanently delete a snapshot |
| `delete_vm_device` | Remove a device from a VM |
| `rollback_snapshot` | Roll a dataset back; all later data and, with `recursive`, later snapshots are destroyed |

Each one requires `confirmed: true`. Before calling one, state exactly what
will be lost and get explicit approval. Before `rollback_snapshot`, offer to
take a fresh snapshot of the current state first.

If the user asks for one of these and the tool is missing, explain that it is
disabled by design and how to enable it. Do not look for a workaround, such as
a custom app or VM that deletes data.

There is no tool to delete a dataset or pool. Point the user at the TrueNAS
web UI for that.

---

## Limits

truenas-mcp cannot manage shares (SMB/NFS), users, permissions, replication or
periodic snapshot tasks, disks and SMART tests, alerts, or system updates. It
cannot change app configuration values. If asked for these, say so and suggest
the TrueNAS web UI.
