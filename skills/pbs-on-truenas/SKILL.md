---
name: pbs-on-truenas
description: >
  Set up Proxmox Backup Server (PBS) as a VM on TrueNAS SCALE and register it
  with a Proxmox VE cluster, using truenas-mcp and the companion proxmox-mcp
  tools. Use this skill when asked to host Proxmox backups on TrueNAS, create a
  PBS VM on TrueNAS, or connect a TrueNAS-hosted PBS to Proxmox VE.
license: MIT
---

# Skill: Proxmox Backup Server on TrueNAS

This skill walks through running
[Proxmox Backup Server](https://www.proxmox.com/en/proxmox-backup-server) in a
TrueNAS SCALE VM with
[gordcurrie/truenas-mcp](https://github.com/gordcurrie/truenas-mcp), then
registering it with Proxmox VE with
[gordcurrie/proxmox-mcp](https://github.com/gordcurrie/proxmox-mcp).

Follow the approval rules in the `truenas` skill: show each change and get
explicit approval before every create, start or add call.

---

## Before you start

- truenas-mcp must be connected (`list_pools`, `create_dataset`, `create_vm`,
  `add_vm_device`, `start_vm`).
- Steps 6–7 also need proxmox-mcp (`add_storage`, `create_backup`). Without
  it, stop after step 5 and give the user the manual steps.
- Ask the user for: the pool to use, the datastore size, VM memory and vCPUs,
  the bridge or interface for the VM NIC, and the path of a PBS installer ISO
  already on the NAS. Check for existing VMs and datasets with the same names
  first (`list_vms`, `list_datasets`).

---

## Step 1 — Check free space (truenas-mcp)

`list_pools`. Confirm the pool has enough `free` bytes for the datastore zvol
plus headroom; keep the pool below 80 % used after the zvol is created.

## Step 2 — Create a zvol for the datastore (truenas-mcp)

```
create_dataset
  name: "Storage/pbs-disk"
  type: "VOLUME"
  volsize: <size in bytes, e.g. 214748364800 for 200 GiB>
  comments: "Proxmox Backup Server disk"
```

The PBS installer puts its OS on this disk, and the datastore can live on the
same disk. For a separate datastore disk, create a second zvol and attach it
as another DISK in step 4.

## Step 3 — Create the VM (truenas-mcp)

```
create_vm
  name: "pbs"            # alphanumeric only — no hyphens or underscores
  memory: 4096           # MiB
  vcpus: 2
  bootloader: "UEFI"
  autostart: true
  description: "Proxmox Backup Server"
```

Note the `id` in the result.

## Step 4 — Attach devices (truenas-mcp)

Use `list_interfaces` to find the bridge or interface name and
`list_directory` to find the ISO. Then call `add_vm_device` once per device.
Do not put `dtype` inside `attributes`.

```
add_vm_device  vm_id: <id>  dtype: "DISK"
  attributes: {"path": "/dev/zvol/Storage/pbs-disk", "type": "VIRTIO"}
  order: 2

add_vm_device  vm_id: <id>  dtype: "CDROM"
  attributes: {"path": "/mnt/Storage/isos/proxmox-backup-server.iso"}
  order: 1          # boot the installer first

add_vm_device  vm_id: <id>  dtype: "NIC"
  attributes: {"type": "VIRTIO", "nic_attach": "br0"}

add_vm_device  vm_id: <id>  dtype: "DISPLAY"
  attributes: {"type": "SPICE", "web": true}
```

Check the result with `list_vm_devices`.

## Step 5 — Start the VM and install PBS (truenas-mcp, then manual)

`start_vm` with the VM `id`, then confirm with `get_vm` that `status.state`
is `RUNNING`.

The user then installs PBS from the console (TrueNAS UI → Virtualization →
`pbs` → Display), following the
[PBS installation guide](https://pbs.proxmox.com/docs/installation.html):
create a datastore (for example `proxmox-backups`) on the data disk, create a
backup user or API token, and note the server's IP address and TLS
certificate fingerprint. These steps cannot be automated with these tools.

After installation, remove the CDROM so the VM boots from its disk. Use `delete_vm_device`. This is
a destructive tool, only available with `TRUENAS_ALLOW_DESTRUCTIVE=true`. If
it is not available, the user can remove the CDROM in the TrueNAS UI.

## Step 6 — Register PBS with Proxmox VE (proxmox-mcp)

```
add_storage
  storage: "pbs"
  type: "pbs"
  server: "<PBS VM IP>"
  datastore: "proxmox-backups"
  username: "backup@pbs"
  password: "<password or API token secret>"
  fingerprint: "<PBS certificate fingerprint>"
  content: "backup"
```

Ask the user for the password or token secret; do not echo it back. Confirm
with `get_storage` `storage: "pbs"`.

## Step 7 — Test a backup (proxmox-mcp)

```
create_backup
  node: "<node name>"
  vmid: <a small VM or container>
  storage: "pbs"
```

Poll the returned task with `get_task_status` until it finishes.

proxmox-mcp cannot create a **scheduled** backup job yet. Tell the user to add
one in the Proxmox VE UI: Datacenter → Backup → Add, with storage `pbs`.

---

## Afterwards

- Offer to snapshot the zvol (`create_snapshot` `dataset: "Storage/pbs-disk"`)
  now that PBS is set up, as a restore point for the VM itself.
- Run the `audit-truenas` skill to confirm the pool still has headroom.
