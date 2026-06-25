---
title: Introducing KubeStash v2026.6.19
date: "2026-06-19"
weight: 10
authors:
- Arnab Baishnab Nipun
tags:
- backup
- backup-verification
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of [KubeStash v2026.6.19](https://kubestash.com/docs/v2026.6.19/setup/). This release focuses on observability and reliability, bringing real-time backup & restore progress streaming, first-class ClickHouse support, and a set of fixes that make `BackupSession` and `RestoreSession` lifecycle management more predictable. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.6.19/README.md). In this post, we'll highlight the major changes.

---

### Quick highlights
- Added real-time backup & restore progress streaming, exposing per-component progress (files processed, bytes transferred, session completion) in the `Snapshot` and `RestoreSession` status.
- Introduced a unified stream reader for backup download & upload, powering the progress reporting pipeline.
- Added first-class ClickHouse backup & restore support via a new manifest/addon.
- `BackupSession` is now skipped when another session for the same `BackupConfiguration` is still running, preventing overlapping runs.
- Fixed a restic repository integrity issue that could leave repositories in an inconsistent state.
- Snapshot/RestoreSession now correctly transition to `Failed` when any single component fails.
- Added WAL backup support for Azure credential-less mode and updated the Neo4j driver list.
- Improved installation: Cilium network policy support, offline (air-gapped) installation, and ArgoCD/FluxCD/OpenShift guides.

---

### What's New

#### Backup & Restore Progress Streaming

Knowing how far along a backup or restore is has historically required inspecting pod logs or guessing from restic output. With v2026.6.19, KubeStash now reports **real-time progress** directly in the Kubernetes API surface, so you can monitor backup & restore activity with standard tools like `kubectl`.

Two new structures—`BackupProgress` and `RestoreProgress`—have been added to the `Snapshot` and `RestoreSession` status respectively. Each component reports fields such as `secondsElapsed`, `percentDone`, `totalFiles`, `filesDone`, the amount of data done (`backupDone`/`restoreDone`, human-readable), the `total` to transfer, and a `speed` figure. This data is produced by a dedicated **stream reader** introduced in this release for both download and upload paths, which every backup driver (workload, PVC, kubedump, and the manifest addons) now uses to emit progress events back to the operator.

**Example: inspecting snapshot progress**

```bash
$ kubectl get -n demo snapshot.storage.kubestash.com <snapshot-name> -ojsonpath='{.status.components.dump.resticStats[0].progress}' | jq
{
  "backupDone": "3.073GB",
  "percentDone": "93.94%",
  "secondsElapsed": 100,
  "total": "3.272GB",
  "totalFiles": 7,
  "speed": "32MB/s"
}
```

**Example: Snapshot component status**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
# ...
status:
  components:
    dump:
      driver: Restic
      path: repository/v1/frequent-backup/dump
      phase: Running
      resticStats:
        - progress:
            backupDone: "3.073GB"
            percentDone: "93.94%"
            secondsElapsed: 100
            total: "3.272GB"
            totalFiles: 7
            filesDone: 6
            speed: "32MB/s"
```

**Example: RestoreSession component progress**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RestoreSession
# ...
status:
  phase: Running
  components:
    dump:
      phase: Running
      progress:
        restoreDone: "1.5Gi"
        percentDone: "50.00%"
        secondsElapsed: 45
        total: "3Gi"
        totalFiles: 1200
        filesDone: 612
        speed: "34MB/s"
```

Because progress is published as structured status, it integrates cleanly with dashboards, alerts, and GitOps workflows without scraping logs.

PR Links:
- https://github.com/kubestash/apimachinery/pull/232
- https://github.com/kubestash/apimachinery/pull/239
- https://github.com/kubestash/kubedump/pull/95
- https://github.com/kubestash/pvc/pull/88
- https://github.com/kubestash/workload/pull/103
- https://github.com/kubestash/manifest/pull/71

---

#### ClickHouse Support

KubeStash now ships a dedicated addon manifest for [ClickHouse](https://clickhouse.com/), enabling native backup & restore of ClickHouse databases using the ClickHouse `BACKUP`/`RESTORE` commands. Combined with the new progress streaming, you can now track ClickHouse backup completion directly from the `Snapshot` status.

```yaml
apiVersion: addons.kubestash.com/v1alpha1
kind: Function
metadata:
  name: clickhouse-backup
spec:
  image: ghcr.io/kubestash/clickhouse:latest
  driverImage: ghcr.io/kubestash/kubedump:v0.27.0
```

This rounds out KubeStash's growing list of supported databases and makes ClickHouse a first-class citizen in your disaster-recovery strategy.

PR Link: https://github.com/kubestash/apimachinery/pull/233

---

#### Skip BackupSession When Another Is Running

Previously, if a backup run took longer than the schedule interval, a new `BackupSession` could be created while the previous one was still in progress—competing for the same restic repository locks and potentially corrupting state. KubeStash now detects that a session is already running for a given `BackupConfiguration` and **skips** the newly invoked `BackupSession`, marking it appropriately instead of launching a duplicate job.

This makes backup scheduling far more resilient to long-running backups and slow repositories.

PR Link: https://github.com/kubestash/kubestash/pull/372

---

#### WAL Backup Support for Azure Credential-less Mode

Following the Azure credential-less mode introduced in v2026.4.27, this release extends it to support **WAL (Write-Ahead Log) backups**. Continuous WAL archiving for databases like PostgreSQL can now leverage the same Azure managed identity pipeline, so WAL backups no longer require storing long-lived storage credentials.

PR Link: https://github.com/kubestash/apimachinery/pull/229

---

#### Neo4j Driver List Update

The Neo4j addon driver list has been refreshed to align with the latest Neo4j versions, ensuring compatibility with current Neo4j deployments.

PR Link: https://github.com/kubestash/apimachinery/pull/231

---

## Bug Fixes and Improvements

### Fail Fast on Component Failure

Previously, a failure in one component of a multi-component backup or restore could leave the overall `Snapshot`/`RestoreSession` in an ambiguous state. Now, if any single component fails, the parent `Snapshot` or `RestoreSession` transitions to `Failed` with an explanatory condition, giving you a clear, terminal signal to act on.

PR Link: https://github.com/kubestash/apimachinery/pull/243

---

### Restic Repository Integrity Fix

This release fixes a restic repository integrity issue that could, under certain conditions, leave a repository in an inconsistent state. The repository initialization and locking paths have been hardened so that concurrent operations and interrupted runs no longer corrupt the index.

PR Link: https://github.com/kubestash/kubestash/pull/369

---

### RBAC Permissions for Vault

The KubeStash operator now ships the RBAC permissions required to interact with Vault when Vault is used as an encryption/credentials backend, removing the need to manually patch role bindings.

PR Link: https://github.com/kubestash/kubestash/pull/356

---

## Installation Improvements

This release also brings a number of installation and operational enhancements:

- **Cilium network policies**: The installer now supports generating [Cilium](https://cilium.io/) network policies alongside the default Kubernetes network policies, enforced via a new `global.networkPolicy` configuration. PR: https://github.com/kubestash/installer/pull/348
- **Offline (air-gapped) installation**: A complete air-gapped installation guide lets you deploy KubeStash in restricted environments with no internet access. Docs: https://github.com/kubestash/docs/pull/87
- **GitOps & platform guides**: New step-by-step installation guides for ArgoCD, FluxCD, and OpenShift. Docs: https://github.com/kubestash/docs/pull/79

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.6.19/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.6.19/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
