---
title: Introducing KubeStash v2026.11.15
date: "2026-11-17"
weight: 10
authors:
- Pulok Saha
tags:
- backup
- cbt
- disaster-recovery
- documentdb
- kubernetes
- kubestash
- kubevirt
- pitr
- restic
- restore
- vm-backup
---

We are pleased to announce the release of [KubeStash v2026.11.15](https://kubestash.com/docs/v2026.11.15/setup/). This release rolls up the changes from [v2026.11.10](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.11.10/README.md) and brings a **resident KubeVirt CBT archiver with point-in-time restore**, **PVC support for `kubectl kubestash download`**, **DocumentDB continuous archiving with PITR**, and a fix that stops a repository's `lastBackupTime` from moving backwards, along with updated dependencies. You can check out the full changelogs for [v2026.11.15](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.11.15/README.md) and [v2026.11.10](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.11.10/README.md). In this post, we'll highlight the key changes.

---

### Quick highlights
- Added the **`KubeVirtArchiver`** and **`KubeVirtArchiverRestore`** APIs: continuous, incremental VM backup using KubeVirt **changed block tracking (CBT)** with Restic, and restore to the latest or to a **point in time**.
- Added **PVC as storage** support to `kubectl kubestash download`, so snapshots in a local PVC backend can be downloaded.
- Added **DocumentDB continuous archiving** with a `DocumentDBArchiver` (full, WAL and manifest backups) and **point-in-time recovery**.
- Fixed `Repository` **`lastBackupTime`** moving backwards when snapshots are resynced.
- Added the RBAC needed for **Solr restore** through `OpsRequest`.
- Updated `apimachinery` dependencies for **Kubernetes 1.34** and cleaned up unused dependencies.

---

### What's New

The examples assume the namespace `demo`, an S3 `BackupStorage` named `s3-storage` and an encryption secret named `encrypt-secret`.

#### KubeVirt: Continuous CBT Backup and Point-in-Time Restore

KubeStash already backs up KubeVirt VMs on a schedule. The new **`KubeVirtArchiver`** makes this continuous. A resident job per VM uses KubeVirt's **native CBT pull-mode export** and streams each disk's map-and-extent artifact **straight into Restic**. That gives you:

- **One full backup, then native incrementals.** After the first full, every run carries only the dirty extents (plus map operations such as `zero`/`hole`), not a full raw disk.
- **No VM-sized staging PVC.** The only temporary volume is KubeVirt's per-export scratch PVC, which the archiver deletes after each iteration.
- **An ordered checkpoint lineage per disk**, recorded on the `Snapshot` (`status.components.<disk>.resticCBTStats`): each entry has a `kind` (`full` or `incremental`), its `checkpoint`, and the `parentCheckpoint` it extends.
- **Safe failure handling.** A partial checkpoint is never published. If a run fails, the next one re-anchors with a full backup instead of extending a broken chain.

Prepare the VM first. Label it so it matches a `changedBlockTrackingLabelSelectors` entry in your KubeVirt CR (here `cbt: enabled`), plus a label the archiver selects on, and use `masquerade` networking so the pull-mode export can reach the VM:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: restic-cbt-vm
  namespace: demo
  labels:
    cbt: enabled                         # matches KubeVirt's CBT label selector
    kubestash.com/cbt-archive: enabled   # selected by the KubeVirtArchiver
spec:
  runStrategy: Always
  template:
    spec:
      domain:
        devices:
          interfaces:
            - name: default
              masquerade: {}             # pull-mode export traffic returns via the pod network
      networks:
        - name: default
          pod: {}
  # disks, DataVolumes and volumes omitted for brevity
```

Then create the archiver:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: KubeVirtArchiver
metadata:
  name: restic-cbt-archiver
  namespace: demo
spec:
  virtualMachines:
    namespaces:
      from: Same
    selector:
      matchLabels:
        kubestash.com/cbt-archive: enabled
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
    subDir: /cbt-archiver
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: Restic
    scheduler:
      schedule: "0 1 * * *"       # a fresh full every night
      jobTemplate: {}             # required by the API, even when empty
  cbtBackup:
    interval: 10m                 # incremental cadence, and so your RPO
    scratchStorageClass: local-path
    # skipQuiesce: true           # only for guests without a QEMU guest agent
    retentionPeriod: 7d
    retentionSchedule: "0 0 * * *"
    successfulHistoryLimit: 5
    failedHistoryLimit: 5
```

Watch the lineage build up. Every disk should show a `full` entry followed by `incremental` entries whose `parentCheckpoint` points at the previous checkpoint:

```bash
$ kubectl -n demo get snapshots.storage.kubestash.com
$ kubectl -n demo get snapshot <snapshot-name> -o json | jq '
    .status.components
    | to_entries[]
    | {component: .key, checkpoints: [.value.resticCBTStats[]
        | {kind, checkpoint, parentCheckpoint}]}'
```

To recover, create a **`KubeVirtArchiverRestore`**. It restores the newest complete multi-disk checkpoint, or, with `recoveryPoint.targetTime`, the newest complete checkpoint at or before that UTC time. Each checkpoint holds every disk plus an immutable copy of the VM manifest, so the restored VM is consistent. The target VM must not already exist: the restore refuses to overwrite one, and your source VM is never touched.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: KubeVirtArchiverRestore
metadata:
  name: restic-cbt-vm-restore
  namespace: demo
spec:
  archiverRef:
    name: restic-cbt-archiver
  source:
    name: restic-cbt-vm
  target:
    name: restic-cbt-vm-restored
  recoveryPoint:
    targetTime: "2026-11-10T07:47:39Z"   # optional; omit to restore the latest checkpoint
```

> A point-in-time target resolves to the newest *complete checkpoint* at or before the requested time, so your recovery granularity equals the `cbtBackup.interval`. The restore can also be run from a different namespace than the source.

---

#### CLI: Download Snapshots from a PVC Backend

Local backends that are backed by a PersistentVolumeClaim are common in air-gapped and on-prem clusters, but `kubectl kubestash download` used to fail on them with `unsupported type of local backend provided`. It now supports a PVC as the storage backend. For example, with this `BackupStorage`:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: local-storage-with-pvc
  namespace: demo
spec:
  default: true
  deletionPolicy: WipeOut
  runtimeSettings:
    pod:
      securityContext:
        fsGroup: 65534
  storage:
    provider: local
    local:
      mountPath: /safe/data
      persistentVolumeClaim:
        claimName: backup-pvc
  usagePolicy:
    allowedNamespaces:
      from: Same
```

you can pull any snapshot to your workstation:

```bash
$ kubectl kubestash download <snapshot-name> -n demo --destination=/tmp/snapshot-data
```

---

#### DocumentDB: Manifest Backup and Continuous Archiving (PITR)

KubeStash can now archive a KubeDB-managed **DocumentDB** continuously. The new `DocumentDBArchiver` combines three things:

- **Full backups** with Restic, taken on a schedule.
- **WAL log backups**, shipped continuously so you can recover to any second inside the archived window.
- **Manifest backups** of the Kubernetes objects around the database (the new manifest backup support in `apimachinery`), kept in their own repository.

Create the archiver. A DocumentDB joins it through a label selector:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: DocumentDBArchiver
metadata:
  name: documentdb-archiver
  namespace: demo
spec:
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: enabled
  fullBackup:
    driver: Restic
    scheduler:
      schedule: "0 1 * * *"           # nightly full backup
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
  manifestBackup:
    scheduler:
      schedule: "*/30 * * * *"        # manifests every 30 minutes
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
  logBackup:
    runtimeSettings: {}
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  retentionPolicy:
    name: demo-retention
    namespace: demo
```

Then label the DocumentDB and point `spec.archiver.ref` at the archiver. WAL is only shipped when a segment fills (16 MB) or when `archive_timeout` fires, so set `archive_timeout` to bound how much data you can lose. Here it is 60 seconds:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb-pitr
  namespace: demo
  labels:
    archiver: enabled
spec:
  version: pg17-0.109.0
  replicas: 1
  storageType: Durable
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 1Gi
  configuration:
    inline:
      user.conf: |
        archive_timeout = 60
  archiver:
    ref:
      name: documentdb-archiver
      namespace: demo
  deletionPolicy: WipeOut
```

To recover, create a **new** DocumentDB with `spec.init.archiver`. Recovery is never done in place. `recoveryTimestamp` is an RFC3339 UTC time with second precision. `manifestRepository` is optional; when set, the restore waits for a manifest snapshot before it starts.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb-pitr-restored
  namespace: demo
spec:
  version: pg17-0.109.0
  replicas: 1
  storageType: Durable
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "2026-11-15T11:10:22Z"   # a UTC time inside the archived window
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: dcdb-pitr-full
        namespace: demo
      manifestRepository:
        name: dcdb-pitr-manifest
        namespace: demo
  deletionPolicy: WipeOut
```

> In our test we inserted a batch of rows, noted the time, inserted a second batch, and restored to a timestamp between the two. The restored database contained only the first batch.

---

### Bug Fixes and Improvements
- **`lastBackupTime` no longer goes backwards.** When snapshots were resynced, the `Repository`'s `lastBackupTime` could be set to an older snapshot's time. That made a healthy repository look stale and could trigger false RPO alerts. It now only moves forward.
- **Solr restore permissions.** The operator now carries the `OpsRequest` permission that Solr restore needs.
- **Kubernetes 1.34 libraries.** `apimachinery` v0.32.0 moves its dependencies to the Kubernetes 1.34 line, and every component is rebuilt on it.
- **Leaner dependencies.** A "Clean up deps" change removes unused modules from `apimachinery`, which every plugin inherits.
- **Added constants** in `apimachinery` for the KubeVirt CBT feature.
- **CI.** All repositories now pull the linter image with `docker run --pull=always`.

---

### Upgrading

No API changes are required for existing `BackupConfiguration`, `BackupStorage`, `Repository` and `RestoreSession` resources. Upgrade the operator with Helm, then run one on-demand backup to confirm everything works:

```bash
$ helm upgrade -i kubestash oci://ghcr.io/appscode-charts/kubestash \
        --version v2026.11.15 \
        --namespace stash --create-namespace \
        --set-file global.license=/path/to/the/license.txt
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: post-upgrade-check
  namespace: demo
spec:
  invoker:
    apiGroup: core.kubestash.com
    kind: BackupConfiguration
    name: sample-backup      # replace with your BackupConfiguration
  session: frequent-backup   # replace with your session name
```

---

### Component Versions

The following components have been released as part of KubeStash v2026.11.15:

| Component | Version |
|-----------|---------|
| kubestash/apimachinery | [v0.32.0](https://github.com/kubestash/apimachinery/releases/tag/v0.32.0) |
| kubestash/kubestash | [v0.32.0](https://github.com/kubestash/kubestash/releases/tag/v0.32.0) |
| kubestash/cli | [v0.31.0](https://github.com/kubestash/cli/releases/tag/v0.31.0) |
| kubestash/kubedump | [v0.31.0](https://github.com/kubestash/kubedump/releases/tag/v0.31.0) |
| kubestash/pvc | [v0.31.0](https://github.com/kubestash/pvc/releases/tag/v0.31.0) |
| kubestash/workload | [v0.31.0](https://github.com/kubestash/workload/releases/tag/v0.31.0) |
| kubestash/manifest | [v0.24.0](https://github.com/kubestash/manifest/releases/tag/v0.24.0) |
| kubestash/volume-snapshotter | [v0.31.0](https://github.com/kubestash/volume-snapshotter/releases/tag/v0.31.0) |
| kubestash/vault | [v0.6.0](https://github.com/kubestash/vault/releases/tag/v0.6.0) |
| kubestash/kubevirt | [v0.4.0](https://github.com/kubestash/kubevirt/releases/tag/v0.4.0) |
| kubestash/ui-server | [v0.4.0](https://github.com/kubestash/ui-server/releases/tag/v0.4.0) |
| kubestash/installer | [v2026.11.15](https://github.com/kubestash/installer/releases/tag/v2026.11.15) |
| kubestash/docs | [v2026.11.15](https://github.com/kubestash/docs/releases/tag/v2026.11.15) |

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.11.15/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.11.15/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
