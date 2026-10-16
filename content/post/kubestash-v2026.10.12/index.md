---
title: Introducing KubeStash v2026.10.12
date: "2026-10-16"
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
- milvus
- neo4j
- oracle
- pitr
- restic
- restore
- solr
- vm-backup
---

We are pleased to announce the release of [KubeStash v2026.10.12](https://kubestash.com/docs/v2026.10.12/setup/). This release brings **backup and restore for KubeVirt VMs**, **Oracle physical backup**, a **Milvus full backup and archiver**, **DocumentDB** and **Solr** backup, Neo4j **composite database** backup and **continuous archiving**, and the new **KubeStash UI server**. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.10.12/README.md). In this post, we'll highlight the key changes.

---

### Quick highlights
- Added **KubeVirt VM backup & restore** with two modes for running VMs: `cbt` (changed block tracking) and `vmsnapshot` (CSI `VolumeSnapshot`s).
- Added **Oracle physical backup** with RMAN, streamed straight to S3.
- Added **Neo4j composite database & alias backup** and **continuous archiving** with manifest backup.
- Added **Milvus full backup** and a **`MilvusArchiver`** for continuous archiving and PITR. Snapshots now record `milvusStats`.
- Added **DocumentDB** backup & restore.
- Added **Solr** backup & restore to S3 and GCS.
- Introduced the **KubeStash UI server** and the `ui.kubestash.com/v1alpha1` API.
- Added **`--new-password-stdin`** and **`--new-password`** to `kubectl kubestash password add` and `update`.
- Exported the cloud storage constants and added a `SnapshotsDirectory` constant in `apimachinery`.
- Built with **Go 1.27**, tested on Kubernetes up to **1.37**, **debian:13** base images, restic bumped to **v0.5.1**.

---

### What's New

The examples assume the namespace `demo`, an S3 `BackupStorage` named `s3-storage`, a `RetentionPolicy` named `demo-retention` and an encryption secret named `encrypt-secret`.

#### KubeVirt VM Backup & Restore [PR](https://github.com/kubestash/installer/pull/366)

KubeStash can now back up and restore **running** [KubeVirt](https://kubevirt.io/) virtual machines. The new `kubevirt` addon captures the VM manifest and every disk, encrypted and deduplicated, in any KubeStash backend. The `backupMode` parameter decides how the disks of a running VM are read:

| Mode | How the disks are read | What you need |
|------|------------------------|---------------|
| `vmsnapshot` (default) | KubeVirt creates a `VirtualMachineSnapshot`, and the plugin backs up the CSI `VolumeSnapshot` of each disk. | A snapshot-capable CSI driver and a `VolumeSnapshotClass`. |
| `cbt` | KubeVirt **changed block tracking** and its pull-mode export read the disks directly. | Any StorageClass. KubeVirt v1.8 or newer. |

Enable the KubeVirt feature gates first: `Snapshot` for `vmsnapshot`, and `VMExport`, `UtilityVolumes` and `IncrementalBackup` for `cbt`. For `cbt`, label the `VirtualMachine` itself with `cbt: enabled` (it must match a `changedBlockTrackingLabelSelectors` entry in the KubeVirt CR) and use `masquerade` networking, so the pull-mode export can reach the VM.

The target of the `BackupConfiguration` is the `VirtualMachine`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: cbt-vm-backup
  namespace: demo
spec:
  target:
    apiGroup: kubevirt.io
    kind: VirtualMachine
    name: cbt-vm
    namespace: demo
  backends:
    - name: s3-backend
      storageRef:
        name: s3-storage
        namespace: demo
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: full
      scheduler:
        schedule: "0 */6 * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: cbt-vm-repo
          backend: s3-backend
          directory: /cbt-vm
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: kubevirt
        tasks:
          - name: logical-backup
            params:
              backupMode: cbt
              ## Every run reads each whole disk. Required in this release.
              forceFull: "true"
              ## KubeVirt needs a scratch PVC for each CBT run.
              cbtScratchStorageClass: local-path
              snapshotReadyTimeout: 10m
```

For `vmsnapshot`, only the task parameters change:

```yaml
          - name: logical-backup
            params:
              backupMode: vmsnapshot
              snapshotReadyTimeout: 10m
```

To restore, delete the source VM with its `DataVolume`s and PVCs, then create a `RestoreSession`. There is deliberately **no `spec.target`**: the VM does not exist at restore time, so the plugin reads its identity from the backed-up manifest and recreates the disks, then the VM:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: cbt-vm-restore
  namespace: demo
spec:
  dataSource:
    repository: cbt-vm-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: kubevirt
    tasks:
      - name: logical-restore
```

> `cbt` backups are full backups for now (`forceFull: "true"` is required), so incremental CBT backups are not available yet.

---

#### Oracle Physical Backup with RMAN [PR](https://github.com/kubestash/apimachinery/pull/238)

The new `oracle-addon` takes **full physical backups of KubeDB-managed Oracle databases**. It runs RMAN and streams the backup pieces directly to S3-compatible storage through the Oracle Secure Backup Cloud Module (OSBWS). The OSBWS wallet is set up for you from your `BackupStorage`. Each run covers the database, the archived redo logs, the control file and the SPFILE under a single RMAN tag, and the RMAN tag, DBID and until-SCN are recorded on the `Snapshot`. It works with Standalone and DataGuard deployments.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: oracle-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Oracle
    name: sa-test
    namespace: demo
  backends:
    - name: s3-backend
      storageRef:
        name: s3-storage
        namespace: demo
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/30 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: oracle-s3-repo
          backend: s3-backend
          directory: /oracle
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: oracle-addon
        tasks:
          - name: physical-backup
            ## params:
            ##   targetPod: sa-test-0     # optional: run RMAN on a specific Oracle pod
```

A restore stops at the boundary recorded on the snapshot by default. For point-in-time recovery, set `untilTime` (UTC). `untilSCN` and `recoverToLatest` are the other recovery targets.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: oracle-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Oracle
    namespace: demo
    name: sa-test-restored
  dataSource:
    repository: oracle-s3-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: oracle-addon
    tasks:
      - name: physical-backup-restore
        params:
          untilTime: "2026-10-10 09:30:00"   # optional point-in-time target (UTC)
```

> Only S3-compatible storage is supported, with one repository per session.

---

#### Neo4j: Composite Databases and Continuous Archiving [PR](https://github.com/kubestash/apimachinery/pull/246)

**Composite databases and aliases.** The `neo4j-addon` logical backup now understands Neo4j composite databases. A backup contains the physical databases **plus** the metadata for composite databases and for local and remote aliases. To back up only some of them, set `databases: "media"` in the `logical-backup` task params. KubeStash also selects the physical databases the aliases depend on.

On restore, KubeStash recreates the physical databases first, then the composite databases and the aliases. The password of a stored-native remote alias is deliberately **not** in the backup. Supply it at restore time through a Secret whose `credentials.yaml` key maps each alias name to its `password`, and reference that Secret with `remoteAliasCredentialsSecret`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: source-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-composite-repo
    snapshot: latest
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0"                        # Neo4j pod that is seeded
          remoteAliasCredentialsSecret: neo4j-remote-alias-credentials
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0                      # PVC of the seed pod
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474                                           # the neo4j user
```

**Continuous archiving.** KubeDB's `Neo4jArchiver` combines full backups, a **manifest backup** of the Kubernetes objects around the database, and differential backups every few minutes that carry the transaction logs. A Neo4j database joins the archiver with the label `archiver: neo4j-pitr` and `spec.archiver.ref`. See the full [continuous archiving guide](https://github.com/kubedb/docs/blob/master/docs/guides/neo4j/pitr/archiver.md).

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: Neo4jArchiver
metadata:
  name: neo4j-pitr-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: neo4j-pitr
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
    subDir: /neo4j-backup
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: Neo4jAdmin
    scheduler:
      schedule: "0 1 * * *"
    sessionHistoryLimit: 3
    timeout: 1h
  differentialBackup:
    backupInterval: 5m
    successfulLogHistoryLimit: 5
    failedLogHistoryLimit: 5
  manifestBackup:
    scheduler:
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 3
    timeout: 15m
  deletionPolicy: Delete
```

To recover to a point in time, create a new Neo4j with `spec.init.archiver`. KubeDB creates the restore sessions for you:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-pitr-restored
  namespace: demo
spec:
  version: "2026.06.0"
  replicas: 1
  init:
    archiver:
      fullDBRepository:
        name: neo4j-pitr-source-full
        namespace: demo
      manifestRepository:
        name: neo4j-pitr-source-manifest
        namespace: demo
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      recoveryTimestamp: "2026-10-10T13:32:30Z"   # a whole-second UTC time inside the archived window
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

---

#### Milvus: Full Backup and Continuous Archiver [PR](https://github.com/kubestash/apimachinery/pull/263)

The new `milvus-addon` takes **full backups** of [Milvus](https://milvus.io/). `physical-backup` captures the etcd metadata at a single revision together with the bucket objects, with writes quiesced for a consistent view. `logical-backup` backs up chosen collections through the `collections` parameter. Each `Snapshot` now records `milvusStats`: the etcd `metaRevision`, the `tso`, the `fenceTime` (the point the full backup restores to, and the base of PITR), how Milvus was `quiesce`d, and the `objectCount` and `bytesCopied`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: milvus-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Milvus
    namespace: demo
    name: milvus-pitr
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: full-backup
      scheduler:
        schedule: "0 2 * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: milvus-s3-repo
          backend: s3-backend
          directory: /milvus
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: milvus-addon
        tasks:
          - name: physical-backup
            params:
              quiesce: DenyWrites     # block writes while the metadata is captured
              fenceSettle: 10s        # let Milvus pick up the write fence
              fenceTimeout: 60s       # writes never stay denied longer than this
              parallelism: "16"       # parallel object copies
```

For **continuous archiving and point-in-time recovery**, the new `MilvusArchiver` takes periodic full backups, records every etcd change and every new object version in a log, and backs up the manifests. A Milvus joins it with the label `archiver: milvus-pitr` and `spec.archiver.ref`:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MilvusArchiver
metadata:
  name: milvus-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: milvus-pitr
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
    subDir: /milvus-backup
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: Restic
    task:
      params:
        quiesce: DenyWrites
        fenceSettle: 10s
        fenceTimeout: 60s
    scheduler:
      schedule: "0 1 * * *"
    sessionHistoryLimit: 3
    timeout: 1h
  logBackup:                          # without this, only full backups are taken
    retentionPeriod: 30d              # how long archived change logs are kept
    retentionSchedule: "0 0 1 * *"    # when old logs are pruned
    successfulLogHistoryLimit: 5
    failedLogHistoryLimit: 5
  manifestBackup:
    scheduler:
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 3
    timeout: 15m
  deletionPolicy: Delete
```

Recover into a new Milvus with `spec.init.archiver`. Give it its own **empty bucket** and the same `rootPath` as the source:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-restored
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-restore-storage     # points at an empty bucket
  init:
    archiver:
      recoveryTimestamp: "2026-10-10T16:22:29Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: milvus-pitr-full
        namespace: demo
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: Delete
```

---

#### DocumentDB Backup & Restore [PR](https://github.com/kubedb/installer/pull/2431)

DocumentDB runs on a PostgreSQL engine, so the new `documentdb-addon` protects it with the proven PostgreSQL tools: `logical-backup` (`pg_dumpall` or `pg_dump`), `physical-backup` (`pg_basebackup`) and `volume-snapshot` (CSI `VolumeSnapshot`). The backup plugins connect through a `<db>-admin` `AppBinding` that the operator publishes for you.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: documentdb-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: DocumentDB
    namespace: demo
    name: sample-documentdb
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
        schedule: "*/30 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: documentdb-s3-repo
          backend: s3-backend
          directory: /documentdb
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: documentdb-addon
        tasks:
          - name: logical-backup
            params:
              backupCmd: pg_dumpall    # or pg_dump
```

To restore into a new DocumentDB, create it with `init.waitForInitialRestore: true`, so it waits for the restore before becoming ready:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: documentdb-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: DocumentDB
    namespace: demo
    name: restored-documentdb
  dataSource:
    repository: documentdb-s3-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: documentdb-addon
    tasks:
      - name: logical-backup-restore
```

---

#### Solr Backup & Restore [PR](https://github.com/kubedb/installer/pull/2440)

KubeDB-managed [Apache Solr](https://solr.apache.org/) can now be backed up to **S3** and **GCS** with the new `solr-addon`. By default it skips internal collections (names starting with `.`, `apm-agent-configuration` or `kubedb-system`); change that with the `args` parameter.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: solr-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Solr
    namespace: demo
    name: solr-cluster
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: solr-frequent-backup
      scheduler:
        schedule: "*/30 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: solr-s3-repo
          backend: s3-backend
          directory: cluster
      addon:
        name: solr-addon
        tasks:
          - name: logical-backup
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: solr-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Solr
    namespace: demo
    name: solr-restored
  dataSource:
    snapshot: latest
    repository: solr-s3-repo
  addon:
    name: solr-addon
    tasks:
      - name: logical-backup-restore
```

---

#### KubeStash UI Server [PR](https://github.com/kubestash/installer/pull/386)

How do you know your data is protected? The new **KubeStash UI server** is an aggregated API server that computes the answer and serves it under **`ui.kubestash.com/v1alpha1`**, ready for dashboards and other tools. It ships as the `kubestash-ui-server` chart, and the `kubestash-opscenter` chart bundles it with `kubestash-metrics` and the user roles:

```bash
$ helm install kubestash-opscenter oci://ghcr.io/appscode-charts/kubestash-opscenter \
        --version v2026.10.12 \
        --namespace stash --create-namespace \
        --set-file global.license=/path/to/the/license.txt
```

It serves four resources:

| Resource | What it tells you |
|----------|-------------------|
| `BackupOverview` | The computed state of one `BackupConfiguration`: schedules, last runs, repositories, backend reachability, the PITR window, the achieved **RPO** and measured **RTO**, and a recovery timeline. |
| `DatabaseBackup` | The backup state of one KubeDB database, **including databases that no `BackupConfiguration` targets** (`Unprotected`). |
| `BackupCoverage` | One cluster-wide object, `cluster`, counting databases that are protected, failing, PITR-enabled or breaching their RPO. |
| `RestoreRequest` | A create-only request that renders and creates the objects needed to restore from a repository. |

Each backup is reported as `Healthy`, `Paused`, `NotReady`, `StorageUnreachable`, `LastBackupFailed`, `NoBackupYet` or `Unprotected`. Set the RPO a `BackupConfiguration` must meet with an annotation, and the UI server reports whether it is breached:

```bash
$ kubectl annotate backupconfiguration sample-postgres-backup -n demo \
    ui.kubestash.com/rpo-target=1h
```

A `RestoreRequest` of type `Clone` restores an application into a **new object in a separate namespace** and leaves the source untouched. `deleteAfter` garbage-collects the clone, which makes recovery drills easy. Use `pitr: {targetTime: ...}` instead of `snapshot` to recover to a point in time, or `--dry-run=server` to see the rendered objects in `status.rendered` without creating anything.

```yaml
apiVersion: ui.kubestash.com/v1alpha1
kind: RestoreRequest
metadata:
  name: postgres-drill
  namespace: demo
spec:
  type: Clone
  source:
    repository: s3-postgres-repo
    snapshot: latest             # a Snapshot name, or "latest"
  clone:
    namespace: demo-drill        # created when missing
    name: postgres-drill
    deleteAfter: 24h
```

---

#### CLI: Update the Repository Password from stdin [PR](https://github.com/kubestash/cli/pull/108)

`kubectl kubestash password add` and `update` used to accept the new restic password only from a file. They now also accept **`--new-password-stdin`**, which is recommended for scripts, and `--new-password`, which exposes the password in your shell history and process list. Exactly one of `--new-password-stdin`, `--new-password` and `--new-password-file` must be given.

```bash
$ printf '%s' "$NEW_PASSWORD" | kubectl kubestash password update s3-postgres-repo \
    -n demo --new-password-stdin
```

Use `printf '%s'` rather than `echo`, because `echo` appends a newline that would become part of the password.

---

### Improvements
- Exported the cloud storage constants in `apimachinery`'s `pkg/blob` (`GCSPrefix`, `AzurePrefix`, `CredentialsDir`, `CACertData` and others) so plugins can reuse them [PR](https://github.com/kubestash/apimachinery/pull/256).
- Added the `SnapshotsDirectory = "snapshots"` constant next to the existing directory constants [PR](https://github.com/kubestash/apimachinery/pull/254).
- All components are built with **Go 1.27** and use **Debian 13** base images, and the installer's CI matrix now covers Kubernetes **v1.31, v1.33, v1.35 and v1.37** [PR](https://github.com/kubestash/apimachinery/pull/267).
- Bumped `gomodules.xyz/restic` to **v0.5.1**. Your existing `BackupConfiguration`, `Repository` and `BackupStorage` resources work without changes [PR](https://github.com/kubestash/apimachinery/pull/253).

---

### Component Versions

The following components have been released as part of KubeStash v2026.10.12:

| Component | Version |
|-----------|---------|
| kubestash/apimachinery | [v0.30.0](https://github.com/kubestash/apimachinery/releases/tag/v0.30.0) |
| kubestash/kubestash | [v0.30.0](https://github.com/kubestash/kubestash/releases/tag/v0.30.0) |
| kubestash/cli | [v0.29.0](https://github.com/kubestash/cli/releases/tag/v0.29.0) |
| kubestash/kubedump | [v0.29.0](https://github.com/kubestash/kubedump/releases/tag/v0.29.0) |
| kubestash/pvc | [v0.29.0](https://github.com/kubestash/pvc/releases/tag/v0.29.0) |
| kubestash/workload | [v0.29.0](https://github.com/kubestash/workload/releases/tag/v0.29.0) |
| kubestash/manifest | [v0.22.0](https://github.com/kubestash/manifest/releases/tag/v0.22.0) |
| kubestash/volume-snapshotter | [v0.29.0](https://github.com/kubestash/volume-snapshotter/releases/tag/v0.29.0) |
| kubestash/vault | [v0.4.0](https://github.com/kubestash/vault/releases/tag/v0.4.0) |
| kubestash/kubevirt | [v0.2.0](https://github.com/kubestash/kubevirt/releases/tag/v0.2.0) |
| kubestash/ui-server | [v0.2.0](https://github.com/kubestash/ui-server/releases/tag/v0.2.0) |
| kubestash/installer | [v2026.10.12](https://github.com/kubestash/installer/releases/tag/v2026.10.12) |
| kubestash/docs | [v2026.10.12](https://github.com/kubestash/docs/releases/tag/v2026.10.12) |

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.10.12/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.10.12/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
