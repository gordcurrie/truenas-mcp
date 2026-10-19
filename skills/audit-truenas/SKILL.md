---
name: audit-truenas
description: >
  Health review of a TrueNAS SCALE system using the truenas-mcp tools. Use this
  skill when asked to audit, review, check, or assess the health of a TrueNAS
  NAS — including pool status and scrubs, pool and dataset capacity, snapshot
  coverage, app state and upgrade backlog, and VM state. Produces a
  prioritised findings report.
license: MIT
---

# Skill: Audit TrueNAS Health (via truenas-mcp)

This skill performs a structured, **read-only** health review of a TrueNAS
SCALE system using the tools exposed by
[gordcurrie/truenas-mcp](https://github.com/gordcurrie/truenas-mcp). Work
through each section below in order, collect findings, then produce a
**Prioritised Findings Report** at the end.

Do not change anything during the audit. If a finding suggests a fix (take a
snapshot, upgrade an app, start a VM), list it as a recommendation. Act on it
only if the user asks afterwards, using the `truenas` skill and its approval
rules.

---

## Before you start

Confirm the truenas-mcp server is configured and the following tools are
available: `get_system_info`, `list_pools`, `list_datasets`, `list_snapshots`,
`list_apps`, `upgrade_summary`, `list_vms`, `list_vm_devices`.

### Complete lists

The list tools return a plain JSON array with no total count. Call them
without `limit` to get every item. If you do pass a `limit`, keep calling with
`offset` increased by the number of items returned until a page comes back
shorter than `limit`. Collect everything before analysing.

---

## Section 1 — System

1. Call `get_system_info`. Note `hostname`, `version`, `physmem`, `cores`,
   `loadavg` and `uptime_seconds`.

**Findings to raise:**
- `[MEDIUM]` 15-minute load average (`loadavg[2]`) higher than `cores` —
  sustained CPU saturation.
- `[LOW]` Uptime under one day when the user did not expect a reboot — check
  for an unexpected restart.
- `[INFO]` Version, hostname, memory and uptime for the record. truenas-mcp
  cannot check whether a TrueNAS update is available; add that to the manual
  review list.

---

## Section 2 — Pool health

1. Call `list_pools`. For each pool note `name`, `status`, `healthy`,
   `warning`, `status_code`, `status_detail`, `fragmentation` and `scan`
   (`function`, `state`, `errors`).

**Check for:**
- `status` other than `ONLINE` (for example `DEGRADED`, `FAULTED`,
  `UNAVAIL`, `OFFLINE`, `REMOVED`), or `healthy: false` — flag `[CRITICAL]`.
  Quote `status_code` and `status_detail`.
- `warning: true` while `ONLINE` — flag `[HIGH]`.
- `scan.errors` greater than `0` — flag `[HIGH]`; the last scrub or resilver
  found errors.
- `scan.function` is `RESILVER` and `scan.state` is `SCANNING` — flag
  `[HIGH]`; a disk is being rebuilt and redundancy is reduced until it ends.
- `scan.state` is `CANCELED`, or no scan has run — flag `[MEDIUM]`.
- `fragmentation` above 50 — flag `[LOW]`.

The tools do not return the time of the last scrub or disk SMART results. Add
both to the manual review list.

---

## Section 3 — Capacity

1. From the Section 2 data, compute used % for each pool as
   `allocated / size` (both in bytes).
2. Call `list_datasets`. Each dataset appears at the top level **and** inside
   its parent's `children`: de-duplicate by `id`. For each dataset note `id`,
   `type`, `used`, `available` and `quota`. `used`, `available` and `quota`
   are human-readable strings such as `"1.52 TiB"`; `quota` is `null` when
   unset.

**Check for:**
- Pool used % at or above 90 — flag `[CRITICAL]`; ZFS performance drops
  sharply and writes can fail.
- Pool used % at or above 80 — flag `[HIGH]`.
- A dataset whose `available` is less than about 10 % of its `used`, or
  whose `available` is much lower than its pool's free space (a quota or
  reservation is close to full) — flag `[MEDIUM]`. Name the dataset and both
  values.
- `[INFO]` The five largest datasets by `used`.

---

## Section 4 — Snapshot coverage

1. Build the list of datasets that matter: every top-level dataset under each
   pool plus any dataset used by an app or VM (for example
   `Storage/immich`, or the zvols named in VM DISK `path` attributes from
   Section 6). Skip the pool's root dataset if it holds no data of its own.
2. For each of these datasets, call `list_snapshots` with `dataset`.
   Snapshots are not sorted and have no timestamp. Find the newest by the
   largest `createtxg` (compare as a number). Read the date from
   `snapshot_name` when it contains one (automatic snapshots look like
   `auto-2026-10-17_00-00 - Daily`).

**Check for:**
- A dataset with **no snapshots** — flag `[HIGH]`; no point-in-time recovery.
- A dataset whose newest snapshot is more than 7 days old, judged from its
  name — flag `[MEDIUM]`; periodic snapshot tasks may have stopped.
- Manual snapshots (names without the `auto-` prefix) older than 90 days,
  judged from the name or from a much lower `createtxg` than the newest —
  flag `[LOW]`; forgotten snapshots hold on to deleted data and space.

If no date can be read from any snapshot name, report counts only and say so.
Snapshot task schedules and retention are not visible through truenas-mcp;
add them to the manual review list.

---

## Section 5 — Apps

1. Call `list_apps`. Note each app's `name` and `state`.
2. Call `upgrade_summary` for **every** app. Note `upgrade_available`,
   `latest_version` and `latest_human_version`.

**Check for:**
- `state` is `CRASHED` — flag `[HIGH]`.
- `state` is `STOPPED` — flag `[LOW]`; confirm it is meant to be off.
- `state` is `DEPLOYING` — flag `[INFO]`; re-check later if it stays there.
- `upgrade_available: true` — flag `[MEDIUM]` per app (app updates often
  carry security fixes). List the app and the target version.
- An app whose data dataset has no recent snapshot (Section 4) and has an
  upgrade pending — flag `[MEDIUM]`; snapshot before upgrading.

---

## Section 6 — Virtual machines

1. Call `list_vms`. Note each VM's `id`, `name`, `status.state`,
   `autostart`, `vcpus` and `memory` (MiB).
2. Call `list_vm_devices` for every VM. Each device's type is in `dtype`.
   Note DISK `path` values (zvols for Section 4) and
   DISPLAY devices. Do not copy DISPLAY `password` values into the report.

**Check for:**
- `autostart: true` but `status.state` is `STOPPED` — flag `[HIGH]`; the VM
  is meant to be running.
- `status.state` is `STOPPED` and `autostart: false` — flag `[INFO]`.
- A DISPLAY device with `bind` not `127.0.0.1` — flag `[MEDIUM]`; the VM
  console is reachable from the network.
- A DISK whose zvol has no snapshots (Section 4) — flag `[MEDIUM]`.
- Total `memory` of running VMs above about 50 % of `physmem` — flag `[LOW]`;
  less RAM is left for the ZFS cache (ARC) and apps.

---

## Section 7 — Limits of truenas-mcp

Some health data is not available through these tools. List these for manual
review in the TrueNAS web UI:

| Data | Where to find it |
|---|---|
| Disk SMART results and temperatures | Storage → Disks |
| Last scrub date and scrub schedule | Storage → pool → Scrub; Data Protection → Scrub Tasks |
| Periodic snapshot tasks and retention | Data Protection → Periodic Snapshot Tasks |
| Replication and cloud sync tasks | Data Protection |
| Active alerts | Bell icon in the top bar |
| TrueNAS system updates | System → Update |

---

## Output — Prioritised Findings Report

After completing all sections, produce a report in this format:

```
# TrueNAS Health Audit — <hostname> — <date>

## Summary
- TrueNAS version: <version>
- Pools: N (M healthy)
- Datasets reviewed: N; snapshots found: M
- Apps: N (M with upgrades); VMs: N (M running)
- Total findings: X critical, Y high, Z medium, W low, V info

## Findings

### [CRITICAL] <title>
**Section:** <N>
**Detail:** <what was found, with pool, dataset, app or VM names and values>
**Recommendation:** <what to do>

### [HIGH] <title>
...

### [MEDIUM] <title>
...

### [LOW] <title>
...

### [INFO] <title>
...

## Items requiring manual review in the TrueNAS UI
- [ ] Disk SMART results
- [ ] Last scrub date and schedule
- [ ] Periodic snapshot tasks and retention
- [ ] Replication / cloud sync tasks
- [ ] Active alerts
- [ ] Available TrueNAS updates
```

Severity definitions:
- **CRITICAL** — data at risk now; act immediately
- **HIGH** — significant risk; fix before the next audit
- **MEDIUM** — meaningful risk; fix within a reasonable window
- **LOW** — hygiene issue; fix when convenient
- **INFO** — noted for the record; no action required
