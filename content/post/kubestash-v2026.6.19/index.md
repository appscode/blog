---
title: Introducing KubeStash v2026.6.19
date: "2026-06-19"
weight: 10
authors:
- Md Anisur Rahman
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
- Added first-class ClickHouse backup & restore support via a new manifest addon.
- `BackupSession` is now skipped when another session for the same `BackupConfiguration` is still running, preventing overlapping runs.
- Fixed a restic repository integrity issue that could leave repositories in an inconsistent state.
- `Snapshot`/`RestoreSession` now correctly transitions to `Failed` when any single component fails.
- Added WAL backup support for Azure credential-less mode and introduced Neo4j backup & restore support with updated driver compatibility.
- Improved installation: Cilium network policy support, offline (air-gapped) installation, and ArgoCD/FluxCD/OpenShift guides.

---

### What's New

#### Backup & Restore Progress Streaming [PR](https://github.com/kubestash/apimachinery/pull/232)

Knowing how far along a backup or restore is has historically required inspecting pod logs or guessing from restic output. With v2026.6.19, KubeStash now reports **real-time progress** directly in the Kubernetes API surface, so you can monitor backup & restore activity with standard tools like `kubectl`.

Two new structures—`BackupProgress` and `RestoreProgress`—have been added to the `Snapshot` and `RestoreSession` status respectively. Each component reports fields such as `secondsElapsed`, `percentDone`, `totalFiles`, `filesDone`, the amount of data transferred (`backupDone`/`restoreDone`, in human-readable form), `total` size, and a live `speed` figure. This data is produced by a dedicated **stream reader** introduced in this release for both download and upload paths, which every backup driver (workload, PVC, kubedump, and the manifest addons) now uses to emit progress events back to the operator.

**Example: inspecting snapshot progress in real time**

```bash
$ kubectl get -n demo snapshot.storage.kubestash.com <snapshot-name> \
    -ojsonpath='{.status.components.dump.resticStats[0].progress}' | jq
{
  "backupDone": "3.073GB",
  "percentDone": "93.94%",
  "secondsElapsed": 100,
  "total": "3.272GB",
  "totalFiles": 7,
  "filesDone": 6,
  "speed": "32MB/s"
}
```

**Example: Snapshot component status during an active backup**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  name: my-backup-config-frequent-backup-1750000000
  namespace: demo
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

**Example: RestoreSession component progress during an active restore**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: my-restore
  namespace: demo
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

Because progress is published as structured status fields, it integrates cleanly with dashboards, alerts, and GitOps workflows—no log scraping required.

Related PRs: [stream reader](https://github.com/kubestash/apimachinery/pull/239), [kubedump](https://github.com/kubestash/kubedump/pull/95), [pvc](https://github.com/kubestash/pvc/pull/88), [workload](https://github.com/kubestash/workload/pull/103), [manifest](https://github.com/kubestash/manifest/pull/71)

---

#### ClickHouse Support [PR](https://github.com/kubestash/apimachinery/pull/233)

KubeStash now ships a dedicated addon manifest for [ClickHouse](https://clickhouse.com/), enabling native backup & restore of ClickHouse databases. This makes ClickHouse a first-class citizen in your disaster-recovery strategy alongside the databases KubeStash already supports.

The ClickHouse addon uses ClickHouse's native `BACKUP`/`RESTORE` SQL commands to transfer data to and from the configured storage backend. Combined with the progress streaming introduced in this release, you can track ClickHouse backup completion directly from the `Snapshot` status.

**BackupConfiguration for ClickHouse**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: clickhouse-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: sample-clickhouse
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      schedule: "*/30 * * * *"
      addon:
        name: clickhouse-addon
        tasks:
          - name: logical-backup
```

**Restoring a ClickHouse database from a snapshot**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: clickhouse-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: sample-clickhouse
  dataSource:
    repository: s3-repo
    snapshot: latest
    encryptionSecret:
      name: encryption-secret
      namespace: demo
  addon:
    name: clickhouse-addon
    tasks:
      - name: logical-backup-restore
```

The underlying `Function` that drives the addon:

```yaml
apiVersion: addons.kubestash.com/v1alpha1
kind: Function
metadata:
  name: clickhouse-backup
spec:
  image: ghcr.io/kubestash/clickhouse:latest
  driverImage: ghcr.io/kubestash/kubedump:v0.27.0
```

---

#### Skip BackupSession When Another Is Running [PR](https://github.com/kubestash/kubestash/pull/372)

Previously, if a backup run took longer than the schedule interval, a new `BackupSession` could be created while the previous one was still in progress—competing for the same restic repository locks and potentially corrupting state. KubeStash now detects that a session is already running for a given `BackupConfiguration` and **skips** the newly invoked `BackupSession`, marking it with a clear status instead of launching a duplicate job.

When a `BackupSession` is skipped, its status reflects a `Skipped` phase with an explanatory condition:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: my-backup-config-frequent-backup-1750000200
  namespace: demo
status:
  phase: Skipped
  conditions:
    - type: BackupSkipped
      status: "True"
      reason: AnotherBackupSessionRunning
      message: >-
        Skipped this BackupSession because another BackupSession
        "my-backup-config-frequent-backup-1750000000" is currently
        running for the same BackupConfiguration.
      lastTransitionTime: "2026-06-19T10:10:00Z"
```

This makes backup scheduling far more resilient to long-running backups and slow repositories, and eliminates the risk of repository lock contention.

---

#### WAL Backup Support for Azure Credential-less Mode [PR](https://github.com/kubestash/apimachinery/pull/229)

Following the Azure credential-less mode introduced in [v2026.4.27](https://appscode.com/blog/post/kubestash-v2026.4.27/), this release extends it to support **WAL (Write-Ahead Log) backups**. Continuous WAL archiving for databases like PostgreSQL can now leverage the same Azure managed identity pipeline, so WAL backups no longer require storing long-lived storage credentials.

A `BackupStorage` without any `secretName` is sufficient—KubeStash resolves Azure Blob Storage credentials through the managed identity at runtime:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: azure-storage
  namespace: demo
spec:
  storage:
    provider: azure
    azure:
      storageAccount: myaccount
      container: mycontainer
      prefix: wal-backup
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: WipeOut
```

With the above storage configured, a continuous WAL-archiving `BackupConfiguration` for PostgreSQL uses it transparently:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: postgres-wal-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
  backends:
    - name: azure-backend
      storageRef:
        namespace: demo
        name: azure-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: continuous-wal
      addon:
        name: postgres-addon
        tasks:
          - name: wal-backup
```

No `secretName` or static credentials needed—Azure Workload Identity handles authentication end-to-end.

---

#### Neo4j Backup & Restore Support [PR](https://github.com/kubedb/installer/pull/2349) [PR](https://github.com/kubestash/apimachinery/pull/231)

KubeStash now supports backup and restore for [Neo4j](https://neo4j.com/) databases. The Neo4j addon uses Neo4j's native `neo4j-admin database dump` and `neo4j-admin database load` commands, and the driver list has been refreshed to cover the latest Neo4j versions.

**Deploy a Neo4j cluster to back up**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-backup
  namespace: demo
spec:
  version: 2025.11.2
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageType: Durable
  deletionPolicy: WipeOut
```

**BackupConfiguration for Neo4j**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: neo4j-backup-config
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: neo4j-backup
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /backup
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
```

**Restoring Neo4j from a snapshot**

The `RestoreSession` requires a `seedServerName` parameter to identify which Neo4j cluster member will seed the restore, and mounts the target pod's data volume directly so the restore job can write into place:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-neo4j
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: neo4j-restore
  dataSource:
    repository: s3-neo4j-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "neo4j-restore-0"
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-neo4j-restore-0
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

---

#### GitOps & Platform Installation Guides [PR](https://github.com/kubestash/docs/pull/79)

New step-by-step installation guides for **ArgoCD**, **FluxCD**, and **OpenShift** make it straightforward to deploy and manage KubeStash through your existing GitOps workflows and on Red Hat OpenShift clusters.

---

### Improvements and Bug Fixes

#### Fail Fast on Component Failure [PR](https://github.com/kubestash/apimachinery/pull/243)

Previously, a failure in one component of a multi-component backup or restore could leave the overall `Snapshot`/`RestoreSession` in an ambiguous state—neither `Succeeded` nor explicitly `Failed`. Now, if any single component fails, the parent `Snapshot` or `RestoreSession` immediately transitions to `Failed` with an explanatory condition, giving you a clear, terminal signal to act on.

**Example: Snapshot with one failed component**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  name: my-backup-config-frequent-backup-1750000000
  namespace: demo
status:
  phase: Failed
  conditions:
    - type: AllComponentsCompleted
      status: "False"
      reason: ComponentFailed
      message: "Component 'dump' failed: exit status 1"
      lastTransitionTime: "2026-06-19T11:00:00Z"
  components:
    dump:
      driver: Restic
      phase: Failed
      error: "exit status 1: unable to open repository: pack file is missing"
    schema:
      driver: Restic
      phase: Succeeded
```

This eliminates ambiguous "Running forever" situations and makes alerting on partial backup failures straightforward.

---

#### Restic Repository Integrity Fix [PR](https://github.com/kubestash/kubestash/pull/369)

This release fixes a restic repository integrity issue that could, under certain conditions, leave a repository in an inconsistent state. The repository initialization and locking paths have been hardened so that concurrent operations and interrupted runs no longer corrupt the index.

---

#### RBAC Permissions for Vault [PR](https://github.com/kubestash/kubestash/pull/356)

The KubeStash operator now ships the RBAC permissions required to interact with Vault when Vault is used as an encryption/credentials backend. Previously these had to be patched in manually; they are now included out of the box.

---

### Installation Improvements

This release brings several installation and operational enhancements:

#### Cilium Network Policy Support [PR](https://github.com/kubestash/installer/pull/348)

The installer now supports generating [Cilium](https://cilium.io/) network policies alongside the default Kubernetes network policies. Enable them during installation by setting `global.networkPolicy.type` to `cilium` in your Helm values:

```yaml
global:
  networkPolicy:
    enabled: true
    type: cilium
```

Full configuration options are documented [here](https://github.com/kubestash/docs/pull/88).

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.6.19/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.6.19/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
