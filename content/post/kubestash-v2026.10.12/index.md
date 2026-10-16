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

We are pleased to announce the release of [KubeStash v2026.10.12](https://kubestash.com/docs/v2026.10.12/setup/). This release brings **backup and restore for KubeVirt virtual machines** with two hot-backup modes, **physical backup for Oracle**, a **Milvus full backup and continuous archiver**, and backup support for **DocumentDB** and **Solr**. Neo4j gains **composite database backup** and **continuous archiving**. The release also introduces the **KubeStash UI server**, a new `ui.kubestash.com` API that tells you at a glance whether your data is protected, and a CLI that can now take a new repository password from **stdin**. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.10.12/README.md). In this post, we'll highlight the key changes.

---

### Quick highlights
- Added **KubeVirt VM backup & restore** with two modes for running VMs: `cbt` (KubeVirt changed block tracking) and `vmsnapshot` (`VirtualMachineSnapshot` backed by CSI `VolumeSnapshot`s).
- Added **Oracle physical backup** using RMAN, streamed straight to S3-compatible storage.
- Added **Neo4j composite database & alias backup** and **Neo4j continuous archiving** with manifest backup.
- Added **Milvus full backup** (physical and logical) and a **`MilvusArchiver`** for continuous archiving and point-in-time recovery. Snapshots now record `milvusStats`.
- Added **DocumentDB backup & restore** with logical, physical and volume-snapshot tasks.
- Added **Solr backup & restore** to S3 and GCS.
- Introduced the **KubeStash UI server**, the `ui.kubestash.com/v1alpha1` API group and the `kubestash-ui-server` and `kubestash-opscenter` Helm charts.
- Added **`--new-password-stdin`** and **`--new-password`** to `kubectl kubestash password add` and `update`.
- Exported the cloud storage constants in `apimachinery` and added a `SnapshotsDirectory` constant.
- Moved the build to **Go 1.27**, the test matrix up to **Kubernetes 1.37**, the base images to **debian:13**, and bumped restic to **v0.5.1**.

---

### What's New

The examples in this post use the namespace `demo` and share one storage setup: an S3 `BackupStorage` named `s3-storage`, a `RetentionPolicy` named `demo-retention`, and an encryption secret named `encrypt-secret`. If you already have these, you can skip this block. Replace the placeholders with your own values.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-secret
  namespace: demo
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: "<your-access-key-id>"
  AWS_SECRET_ACCESS_KEY: "<your-secret-access-key>"
---
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: <your-bucket-name>
      endpoint: https://s3.amazonaws.com
      region: us-east-1
      prefix: demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: false
  deletionPolicy: Delete
---
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: demo-retention
  namespace: demo
spec:
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  failedSnapshots:
    last: 2
  usagePolicy:
    allowedNamespaces:
      from: All
---
apiVersion: v1
kind: Secret
metadata:
  name: encrypt-secret
  namespace: demo
type: Opaque
stringData:
  RESTIC_PASSWORD: "<your-restic-repository-password>"
```

---

#### KubeVirt VM Backup & Restore [PR](https://github.com/kubestash/installer/pull/366)

[KubeVirt](https://kubevirt.io/) lets you run virtual machines next to your containers. With v2026.10.12, KubeStash can back up and restore **running** KubeVirt VMs. The new `kubevirt` addon (plugin `v0.2.0`) ships in the KubeStash catalog and captures the VM manifest together with every disk, then stores them encrypted and deduplicated in any KubeStash backend.

How the disks of a running VM are read is decided by the `backupMode` parameter:

| Mode | How the disks are read | What you need |
|------|------------------------|---------------|
| `vmsnapshot` (default) | KubeVirt creates a `VirtualMachineSnapshot`. The plugin backs up the resulting CSI `VolumeSnapshot` of each disk through a temporary PVC, then cleans everything up. | A snapshot-capable CSI driver and a `VolumeSnapshotClass`. |
| `cbt` | KubeVirt **changed block tracking** and its pull-mode export read the disks directly. No CSI snapshot is involved. | Any StorageClass. KubeVirt v1.8 or newer. |

The parameter also accepts `auto`. As a rule of thumb, pick `cbt` when your StorageClass cannot take CSI snapshots (such as `local-path`), and `vmsnapshot` when it can.

**Step 1: Enable the KubeVirt features**

Both modes need a few KubeVirt feature gates. `Snapshot` serves `vmsnapshot`; `VMExport`, `UtilityVolumes` and `IncrementalBackup` serve `cbt`. CBT is switched on per VM through a label selector. Apply this as a **patch**, never as a full `kubectl apply`, so the rest of your KubeVirt CR is left untouched:

```yaml
# kubectl -n kubevirt patch kubevirt kubevirt --type merge --patch-file kubevirt-feature-gates.patch.yaml
spec:
  configuration:
    developerConfiguration:
      featureGates:
        - Snapshot
        - VMExport
        - UtilityVolumes
        - IncrementalBackup
    changedBlockTrackingLabelSelectors:
      virtualMachineLabelSelector:
        matchLabels:
          cbt: enabled
```

> `featureGates` is a list, so a merge patch **replaces** it. If your KubeVirt CR already enables other gates, include them here too.

**Step 2: A VM that is ready for CBT**

Two details matter for CBT. The `cbt: enabled` label must sit on the `VirtualMachine` object itself, not on `spec.template`, and KubeVirt only applies it when the VM starts. The VM also needs `masquerade` networking, because the pull-mode export is served by `virt-launcher` dialling out to the export Service.

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cbt-vm
  namespace: demo
  labels:
    cbt: enabled                 # matches changedBlockTrackingLabelSelectors above
spec:
  runStrategy: Always
  dataVolumeTemplates:
  - metadata:
      name: cbt-vm-root
    spec:
      source:
        registry:
          url: docker://quay.io/kubevirt/cirros-container-disk-demo:latest
      storage:
        accessModes: [ReadWriteOnce]
        volumeMode: Filesystem
        storageClassName: local-path     # CBT reads through KubeVirt, so no CSI snapshot is needed
        resources:
          requests:
            storage: 2Gi
  - metadata:
      name: cbt-vm-data
    spec:
      source:
        blank: {}
      storage:
        accessModes: [ReadWriteOnce]
        volumeMode: Filesystem
        storageClassName: local-path
        resources:
          requests:
            storage: 1Gi
  template:
    metadata:
      labels:
        kubevirt.io/vm: cbt-vm
    spec:
      domain:
        cpu:
          cores: 1
        resources:
          requests:
            memory: 256Mi
        devices:
          interfaces:
          - name: default
            masquerade: {}               # required for the CBT export
          disks:
          - name: rootdisk
            disk:
              bus: virtio
          - name: datadisk
            disk:
              bus: virtio
      networks:
      - name: default
        pod: {}
      terminationGracePeriodSeconds: 30
      volumes:
      - name: rootdisk
        dataVolume:
          name: cbt-vm-root
      - name: datadisk
        dataVolume:
          name: cbt-vm-data
```

**Step 3: Back up in `cbt` mode**

The target of the `BackupConfiguration` is the `VirtualMachine` itself. The `kubevirt` addon's `logical-backup` task does the work, and `backupMode` selects the mechanism:

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
              ## Every run reads each whole disk. Required in this release so the
              ## restore can read the data back from the repository.
              forceFull: "true"
              ## KubeVirt needs a scratch PVC for each CBT run.
              cbtScratchStorageClass: local-path
              snapshotReadyTimeout: 10m
```

**Step 4: Back up in `vmsnapshot` mode**

For snapshot mode your VM disks must live on a CSI driver that can snapshot, with a matching `VolumeSnapshotClass`. The example below uses Longhorn; any CSI driver with snapshot support works the same way.

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: longhorn-snapshot-vsc
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
driver: driver.longhorn.io
deletionPolicy: Delete
parameters:
  type: snap                  # an in-cluster Longhorn snapshot, which is what KubeVirt needs
```

The `BackupConfiguration` is the same shape. Only the target VM and the task parameters change. Notice that `vmsnapshot` needs no `forceFull` and no scratch StorageClass:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: vmsnap-vm-backup
  namespace: demo
spec:
  target:
    apiGroup: kubevirt.io
    kind: VirtualMachine
    name: vmsnap-vm
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
        - name: vmsnap-vm-repo
          backend: s3-backend
          directory: /vmsnap-vm
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: kubevirt
        tasks:
          - name: logical-backup
            params:
              backupMode: vmsnapshot
              snapshotReadyTimeout: 10m
```

For each run, KubeVirt cuts one `VolumeSnapshot` per disk, the plugin clones a temporary PVC from each, backs those up, and then deletes the temporary PVCs, the `VirtualMachineSnapshot` and the worker Jobs.

**Trigger a backup on demand**

Besides the schedule, you can start a backup at any time with a `BackupSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: cbt-vm-manual-1
  namespace: demo
spec:
  invoker:
    apiGroup: core.kubestash.com
    kind: BackupConfiguration
    name: cbt-vm-backup
  session: full
```

Check the result. Use the fully qualified `Snapshot` name, because other projects (such as Longhorn) also define a `snapshots` resource:

```bash
$ kubectl get backupsession -n demo
$ kubectl get snapshots.storage.kubestash.com -n demo
```

A successful `Snapshot` contains a `manifest` component, which holds the VM definition, and one volume component for each disk. A CBT snapshot also carries the `kubevirt.kubestash.com/raw-disk-image` label, which tells the restore to write raw disk images. You do not need to configure anything for that.

**Restore the VM**

To restore in place, make sure the source VM, its `DataVolume`s and its PVCs are gone, then create a `RestoreSession`. There is deliberately **no `spec.target`**: the VM does not exist at restore time, so the plugin reads its identity from the backed-up manifest and recreates the disks first, then the VM and its `DataVolume`s.

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

The addon also exposes `volume-backup` / `volume-restore` and `manifest-backup` / `manifest-restore` tasks if you want to handle the disks and the VM definition separately.

**Things to know in this release**
- `cbt` backups are full backups. `forceFull: "true"` is required for now, so incremental CBT backups are not available yet.
- Restoring a VM into a **different namespace** is not supported in plugin `v0.2.0`. Restore in place instead.
- A VM without the QEMU guest agent is backed up **crash-consistent**, as if the power had been cut. The backup log reports this with a `NoGuestAgent` warning. Install the QEMU guest agent in the guest if you need it quiesced during the backup.
- Deleting the VM while its `BackupConfiguration` is still scheduled makes the next run fail. Pause the `BackupConfiguration` with `spec.paused: true` first.

---

#### Oracle Physical Backup with RMAN [PR](https://github.com/kubestash/apimachinery/pull/238)

KubeStash can now take **full physical backups of KubeDB-managed Oracle databases** with the new `oracle-addon`. The backup runs Oracle RMAN and streams the backup pieces directly to S3-compatible object storage through the Oracle Secure Backup Cloud Module (OSBWS, the `SBT_TAPE` channel).

Highlights:
- **The OSBWS wallet is configured for you.** The plugin builds the wallet config secret from your `BackupStorage` and triggers a `Reconfigure` `OracleOpsRequest` to set the wallet up on the database. Later backups reuse the wallet and refresh it only when the backup storage changes.
- **One consistent backup set.** Each run covers the database, the archived redo logs, the current control file and the SPFILE, all under a single RMAN tag (`KS-...`).
- **Recovery metadata on the `Snapshot`.** The RMAN tag, DBID, until-SCN, incarnation, and the control file and SPFILE handles are recorded, so a restore knows exactly where to stop.
- Works with **Standalone** and **DataGuard** Oracle deployments. Use the optional `targetPod` parameter to run RMAN on a specific pod.

**BackupStorage**

The credential secret must use the keys `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oracle-s3-secret
  namespace: demo
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: "<your-access-key-id>"
  AWS_SECRET_ACCESS_KEY: "<your-secret-access-key>"
---
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: oracle-s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: s3.amazonaws.com
      region: us-east-1
      bucket: oracle-backup
      secretName: oracle-s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: false
  deletionPolicy: WipeOut
```

**BackupConfiguration for Oracle**

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
        name: oracle-s3-storage
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

**RestoreSession for Oracle**

The `physical-backup-restore` task restores the database and, by default, stops recovery exactly at the boundary recorded on the snapshot, so no redo written after the backup is applied. For point-in-time recovery, set `untilTime` (UTC, `YYYY-MM-DD HH24:MI:SS`). RMAN stops *before* that instant, so leave a few seconds of margin. `untilSCN` and `recoverToLatest` are the other recovery targets, and the three are mutually exclusive.

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

**Limitations**
- Only S3-compatible `BackupStorage` is supported.
- Use **exactly one repository per session**. Multi-repository backup sessions are not supported.
- The S3 `prefix` is **not** applied to RMAN data, because the SBT module writes backup pieces at the root of the bucket. The prefix applies only to KubeStash metadata, so do not rely on it for bucket lifecycle rules or manual cleanup.
- Only full backups are supported in this release.

---

#### Neo4j: Composite Databases and Continuous Archiving [PR](https://github.com/kubestash/apimachinery/pull/246)

Two Neo4j capabilities land in this release.

**Backup and restore of composite databases and aliases**

A Neo4j composite database provides one query context over several constituent databases. The `neo4j-addon` logical backup now understands them (see the full [composite database guide](https://github.com/kubedb/docs/blob/master/docs/guides/neo4j/backup/kubestash/logical/composite-database/index.md)). KubeStash discovers the catalog through Neo4j's `system` database, so a backup contains the physical database artifacts **plus** the metadata for composite databases, composite aliases and standalone (local and remote) aliases.

Imagine a `source-neo4j` that owns a `movies` database and a `media` composite database. `media.movies` is a local alias, and `media.reviews` is a remote alias pointing at a `reviews` database on a different Neo4j deployment:

```cypher
CREATE DATABASE movies WAIT;
CREATE COMPOSITE DATABASE media WAIT;
CREATE ALIAS media.movies FOR DATABASE movies;
```

The backup configuration looks like any other logical backup. By default the `*` selector backs up every visible physical database and captures all composite and standalone alias definitions:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: source-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: source-neo4j
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
        - name: s3-neo4j-composite-repo
          backend: s3-backend
          directory: /neo4j-composite
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
```

To back up only the `media` composite database, add `databases: "media"` to the task `params`. KubeStash then also selects `movies`, because the local alias `media.movies` depends on it. Do not exclude a physical database that a selected local alias needs.

On restore, KubeStash works in dependency order: it restores the physical databases, waits for local alias targets to come online, creates the composite databases, restores the aliases and finally verifies the reconstructed catalog. Preflight checks run before any destructive step, so a conflicting database or alias fails fast unless overwrite is enabled.

One thing is deliberately **not** in the backup: the password of a stored-native remote alias. You supply it at restore time through a Secret whose `credentials.yaml` key maps each alias name to its password:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: neo4j-remote-alias-credentials
  namespace: demo
type: Opaque
stringData:
  credentials.yaml: |
    media.reviews:
      password: "<remote-database-password>"
    reviews-remote:
      password: "<remote-database-password>"
```

Reference that Secret from the `RestoreSession` with `remoteAliasCredentialsSecret`:

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

**Continuous archiving with manifest backup**

For continuous protection, KubeDB's `Neo4jArchiver` combines KubeStash full backups, a **manifest backup** of the Kubernetes objects around the database, and differential backups every few minutes that carry the transaction logs. A restore can then recover the graph to a chosen point in time (see the full [continuous archiving guide](https://github.com/kubedb/docs/blob/master/docs/guides/neo4j/pitr/archiver.md)). One archiver serves every Neo4j database that selects it:

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

A Neo4j database opts in with the matching label and an explicit archiver reference:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-pitr-source
  namespace: demo
  labels:
    archiver: neo4j-pitr
spec:
  version: "2026.06.0"
  replicas: 3
  archiver:
    ref:
      name: neo4j-pitr-archiver
      namespace: demo
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

To recover, create a new Neo4j with `spec.init.archiver`. KubeDB creates the restore sessions for you and replays the differential chain up to `recoveryTimestamp`:

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

A `recoveryTimestamp` newer than the latest archived transaction does not prove point-in-time recovery: the restore falls back to the latest available state. Pick a time that sits between two archived backups.

---

#### Milvus: Full Backup and Continuous Archiver [PR](https://github.com/kubestash/apimachinery/pull/263)

[Milvus](https://milvus.io/) is a vector database used for semantic search and RAG pipelines. KubeStash v2026.10.12 adds the `milvus-addon` with **full backups in two flavors**, plus a **`MilvusArchiver`** for continuous archiving and point-in-time recovery (PITR).

| Task | Driver | What it does |
|------|--------|--------------|
| `physical-backup` | `Restic` | Captures the etcd metadata at a single revision together with the bucket objects, fenced so writes are quiesced for a consistent view. |
| `logical-backup` | `Milvus` | Backs up collections. Use the `collections` parameter to pick them. |
| `logical-backup-restore` | `Milvus` | Restores collections, appending `restoreSuffix` (default `_restored`) to their names. |
| `physical-backup-restore` | `Restic` | Restores a physical backup and verifies it (`verify`: `sample`, `full` or `none`). |

The backup tells you precisely what it captured. The `Snapshot` now carries `milvusStats`:

| Field | Meaning |
|-------|---------|
| `metaRevision` | The etcd revision at which the metadata was captured. |
| `tso` | The Milvus timestamp high-water mark at that revision. |
| `fenceTime` | The wall-clock time of the metadata capture. This is the point in time the full backup restores to, and the base of PITR. |
| `quiesce` | How Milvus was quiesced: `DenyWrites`, `Flush` or `None`. |
| `objectCount`, `bytesCopied` | How many object versions and bytes this backup archived. |

**A Milvus to protect**

Milvus keeps its durable state in the meta etcd and in the object storage bucket (the Woodpecker write-ahead log lives in the same bucket). First point Milvus at its object storage:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-storage
  namespace: demo
type: Opaque
stringData:
  address: "<object-storage-host>"
  accessKeyID: "<access-key-id>"
  secretAccessKey: "<secret-access-key>"
  bucketName: "<milvus-bucket>"
  rootPath: "files"
  port: "9000"
```

The database opts in to the archiver with a label and a reference (a double opt-in, like the other archivers):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-pitr
  namespace: demo
  labels:
    archiver: milvus-pitr
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-storage
  archiver:
    ref:
      name: milvus-archiver
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

> A Standalone Milvus created by an older operator keeps its write-ahead log in RocksMQ on the data PVC. It cannot be archived. Migrate it with a logical backup restored into a new Milvus.

**Full backup with a `BackupConfiguration`**

A scheduled physical (full) backup of the Milvus above:

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

To back up only some collections instead, swap the task for the logical one:

```yaml
      addon:
        name: milvus-addon
        tasks:
          - name: logical-backup
            params:
              collections: "products,embeddings"   # empty means all collections
```

A logical restore recreates the collections next to the originals, so nothing is overwritten:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: milvus-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Milvus
    namespace: demo
    name: milvus-pitr
  dataSource:
    repository: milvus-s3-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: milvus-addon
    tasks:
      - name: logical-backup-restore
        params:
          restoreSuffix: _restored    # "products" comes back as "products_restored"
```

**Continuous archiving and PITR**

The `MilvusArchiver` takes periodic full backups with the `Restic` driver, continuously records every etcd change and every new object version in a log, and also backs up the manifests. Together they let you restore to any point inside the recorded window:

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

Recover into a brand new Milvus with `spec.init.archiver`. Give the new Milvus its own **empty bucket** and the same `rootPath` as the source:

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

DocumentDB speaks the MongoDB wire protocol on top of a PostgreSQL engine. Because that engine is PostgreSQL, KubeStash can protect it with the same proven tools. The new `documentdb-addon` provides three backup tasks:

| Task | Driver | What it does |
|------|--------|--------------|
| `logical-backup` | `Restic` | Dumps the database with `pg_dumpall` (or `pg_dump`). |
| `physical-backup` | `Restic` | Takes a physical base backup with `pg_basebackup`. |
| `volume-snapshot` | `VolumeSnapshotter` | Takes a CSI `VolumeSnapshot` of the data volume. |

The restore tasks are `logical-backup-restore` and `physical-backup-restore`, and every backup can be paired with a `manifest-backup` of the surrounding Kubernetes objects. The backup plugins connect to the backend PostgreSQL through a `<db>-admin` `AppBinding` that the operator publishes for you.

**A DocumentDB to protect**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: sample-documentdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

**BackupConfiguration for a logical backup**

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

For a physical backup, change the task and its command:

```yaml
      addon:
        name: documentdb-addon
        tasks:
          - name: physical-backup
            params:
              backupCmd: pg_basebackup
```

For CSI volume snapshots, use the `volume-snapshot` task and name your `VolumeSnapshotClass` through the `volumeSnapshotClassName` parameter.

**RestoreSession**

Create the target DocumentDB with `init.waitForInitialRestore: true` so it waits for the restore before becoming ready:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: restored-documentdb
  namespace: demo
spec:
  init:
    waitForInitialRestore: true
  version: "pg17-0.109.0"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

Then restore the latest snapshot into it:

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

KubeDB-managed [Apache Solr](https://solr.apache.org/) can now be backed up and restored with KubeStash. The new `solr-addon` gives you reliable data protection with cloud storage backends such as **S3** and **GCS**. By default it skips internal collections (names starting with `.`, `apm-agent-configuration` or `kubedb-system`). Narrow or widen that with the `args` parameter.

**BackupConfiguration for Solr**

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
      sessionHistoryLimit: 10
      scheduler:
        schedule: "*/30 * * * *"
        successfulJobsHistoryLimit: 1
        failedJobsHistoryLimit: 1
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

**RestoreSession for Solr**

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
    repository: solr-s3-repo
    snapshot: latest
  addon:
    name: solr-addon
    tasks:
      - name: logical-backup-restore
```

---

#### KubeStash UI Server [PR](https://github.com/kubestash/installer/pull/386)

How do you know your data is actually protected? Before this release, answering that meant stitching together `BackupConfiguration`s, `Repository`s, `Snapshot`s and `BackupSession`s by hand. v2026.10.12 introduces the **KubeStash UI server**, an aggregated API server that computes the answer for you and serves it under the new **`ui.kubestash.com/v1alpha1`** API group. Dashboards and any other tool can consume it as a ready-made view of backup health.

It ships as two Helm charts. `kubestash-ui-server` installs the API server, and `kubestash-opscenter` bundles it with `kubestash-metrics` and the user roles:

```bash
$ helm install kubestash-opscenter oci://ghcr.io/appscode-charts/kubestash-opscenter \
        --version v2026.10.12 \
        --namespace stash --create-namespace \
        --set-file global.license=/path/to/the/license.txt
```

Once installed, the API server registers itself with the Kubernetes API aggregation layer:

```bash
$ kubectl get apiservice v1alpha1.ui.kubestash.com
```

It serves four resources:

| Resource | Scope | What it tells you |
|----------|-------|-------------------|
| `BackupOverview` | namespaced | The computed state of one `BackupConfiguration`: sessions and schedules, last runs, repositories and their integrity, backends and reachability, total size, the continuous-log (PITR) window, and the achieved **RPO** and measured **RTO**. It also lists every recovery point inside the RPO window as a timeline. It shares its name with the `BackupConfiguration`. |
| `DatabaseBackup` | namespaced | The backup state of one KubeDB database, **whether or not any `BackupConfiguration` targets it**. Unprotected databases show up as `Unprotected`. Named `<name>~<Kind>.<group>`, for example `pg~Postgres.kubedb.com`. |
| `BackupCoverage` | cluster | One object named `cluster` that counts databases, how many are protected, failing, PITR-enabled or breaching their RPO, plus the cluster manifest backups. |
| `RestoreRequest` | namespaced | A create-only request that renders and creates the objects needed to restore from a repository. |

Each backup gets one of these health values: `Healthy`, `Paused`, `NotReady`, `StorageUnreachable`, `LastBackupFailed`, `NoBackupYet` or `Unprotected`. The backup methods are reported as `Logical`, `Manifest`, `Physical`, `VolumeSnapshot` or `ContinuousLog`.

```bash
# the whole cluster at a glance
$ kubectl get backupcoverages.ui.kubestash.com cluster -o yaml

# every database, protected or not
$ kubectl get databasebackups.ui.kubestash.com -A

# the details behind one BackupConfiguration
$ kubectl get backupoverviews.ui.kubestash.com -n demo
```

The chart installs a `kubestash:ui:viewer` `ClusterRole` that grants `get` and `list` on the `ui.kubestash.com` resources and is aggregated into the built-in `view` role. Creating a `RestoreRequest` needs additional permissions.

**Set an RPO target**

Annotate a `BackupConfiguration` with the RPO it must meet. The UI server then reports whether the objective is currently breached:

```bash
$ kubectl annotate backupconfiguration sample-postgres-backup -n demo \
    ui.kubestash.com/rpo-target=1h
```

Or declare it in the manifest. The value is any Go duration, such as `15m` or `2h`:

```yaml
metadata:
  name: sample-postgres-backup
  namespace: demo
  annotations:
    ui.kubestash.com/rpo-target: "1h"
```

**Restore from a request**

A `RestoreRequest` of type `Clone` restores an application into a **new object in a separate namespace** and leaves the source untouched. The optional `deleteAfter` garbage-collects the clone once it has existed that long, which makes disposable recovery drills easy:

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
  timeout: 30m
```

To recover to a point inside a continuous-log window, replace `snapshot` with `pitr: {targetTime: "2026-10-10T13:32:30Z"}`. `snapshot` and `pitr` are mutually exclusive. A request of type `Manifest` re-applies backed-up Kubernetes manifests instead, and you can narrow it by namespace, resource and label selectors.

A `RestoreRequest` is never persisted. With a server-side dry run, nothing is created and the rendered objects come back in `status.rendered`, so you can review exactly what a restore would do:

```bash
$ kubectl create -f restorerequest.yaml --dry-run=server -o yaml
```

---

#### CLI: Update the Repository Password from stdin [PR](https://github.com/kubestash/cli/pull/108)

KubeStash encrypts every repository with a restic password, and the `kubectl kubestash password` command (alias `pw`) manages those keys. Until now `password add` and `password update` only accepted a **file** (`--new-password-file`), which meant writing a secret to disk first. This release adds two more ways to supply the new password:

| Flag | Use it when |
|------|-------------|
| `--new-password-stdin` | You pipe the password in, for example from a password manager or a secrets CLI. Recommended for scripts. |
| `--new-password` | You are experimenting. The password lands in your shell history and the process list, so prefer the other two. |
| `--new-password-file` | The password is already in a file. This keeps working as before. |

Exactly one of the three must be given, and they are mutually exclusive. A temporary file is created for the operation and removed when the command finishes.

Update the password of the repository `s3-postgres-repo`:

```bash
$ printf '%s' "$NEW_PASSWORD" | kubectl kubestash password update s3-postgres-repo \
    -n demo --new-password-stdin
```

Add a second key to a repository instead of replacing the existing one:

```bash
$ printf '%s' "$NEW_PASSWORD" | kubectl kubestash password add s3-postgres-repo \
    -n demo --new-password-stdin --user backup-admin
```

Use `printf '%s'` rather than `echo`: `echo` appends a trailing newline, which would become part of the password. Use `--paths` to limit the change to specific component repositories.

---

### Improvements

#### Exported Cloud Constants [PR](https://github.com/kubestash/apimachinery/pull/256)

The cloud storage constants in `apimachinery`'s `pkg/blob` package are now exported. Values such as `GCSPrefix`, `AzurePrefix`, `LocalPrefix`, `CredentialsDir`, `AzureStorageAccount`, `AzureStorageKey`, `AzureAccountKey`, `GoogleServiceAccountJSONKey`, `GoogleApplicationCredentials` and `CACertData` can be reused by plugins and other tools instead of being redefined.

#### `SnapshotsDirectory` Constant [PR](https://github.com/kubestash/apimachinery/pull/254)

`apimachinery` now defines `SnapshotsDirectory = "snapshots"`, next to the existing directory constants such as `DirRepository`, so components share one definition of the `snapshots` directory name instead of repeating the string.

#### Go 1.27, Kubernetes 1.37 and debian:13 [PR](https://github.com/kubestash/apimachinery/pull/267)

All components are now built with **Go 1.27**. Their container base images move to **Debian 13**, and the installer's CI matrix now runs against Kubernetes **v1.31, v1.33, v1.35 and v1.37**.

#### restic v0.5.1 [PR](https://github.com/kubestash/apimachinery/pull/253)

The restic dependency (`gomodules.xyz/restic`) is bumped to **v0.5.1**. This is transparent to users: your existing `BackupConfiguration`, `Repository` and `BackupStorage` resources continue to work without any changes.

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
