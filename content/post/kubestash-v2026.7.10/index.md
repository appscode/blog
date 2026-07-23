---
title: Introducing KubeStash v2026.7.10
date: "2026-07-16"
weight: 10
authors:
- Pulok Saha
tags:
- backup
- backup-verification
- clickhouse
- disaster-recovery
- kubernetes
- kubestash
- neo4j
- pitr
- qdrant
- restic
- restore
---

We are pleased to announce the release of [KubeStash v2026.7.10](https://kubestash.com/docs/v2026.7.10/setup/). This release delivers a major restic upgrade, new database backup support for **Qdrant**, and powerful new capabilities for ClickHouse including an **archiver with point-in-time recovery (PITR)**. It also brings a new validation checker for local backend PVCs, and important fixes for post-restore hooks and backup inspection. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.7.10/README.md). In this post, we'll highlight the key changes.

---

### Quick highlights
- Upgraded the restic dependency to **v0.5.0**, bringing the latest upstream features, performance improvements, and security patches.
- Added **Qdrant backup & restore** support—now you can protect your vector database collections with the same KubeStash workflow.
- Introduced a **ClickHouse Archiver** with **incremental point-in-time recovery (PITR)**, enabling continuous archiving and precise time-based restores.
- Added a new **LocalBackendPVC checker** that validates PVC configurations when using a local storage backend, catching misconfigurations before a backup runs.
- Added **`excludeDatabases`** parameter to Neo4j backup & restore, giving you fine-grained control over which databases to include or skip.
- Fixed Neo4j backup inspection output to **preserve explicit `false` boolean values**, eliminating ambiguity in backup integrity reports.
- Corrected the **post-restore `OnSuccess` execution policy** check so that post-restore hooks fire reliably after a successful restore.
- Hook pod selectors are now parsed with **`labels.Parse`**, ensuring selectors with complex label expressions (e.g., `key in (v1, v2)`) work correctly.

---

### What's New

#### Restic v0.5.0 [PR](https://github.com/kubestash/apimachinery/pull/248)

KubeStash uses [restic](https://restic.net/) as its core backup engine for fast, secure, and deduplicated backups. This release bumps the restic dependency from the previous version to **v0.5.0**, incorporating all upstream improvements from the restic project.

The restic v0.5.0 upgrade brings several benefits to your KubeStash backups:

- **Improved performance**: Faster pack file handling and index operations during backup and restore.
- **Enhanced repository integrity**: Upstream fixes for edge cases that could, in rare situations, affect repository consistency.
- **Better compression and deduplication**: Smarter chunking algorithms reduce storage footprint and transfer time.
- **Security patches**: All upstream security fixes from restic are now included.

This upgrade is transparent to end users—your existing `BackupConfiguration`, `Repository`, and `BackupStorage` resources continue to work without any changes. The updated restic binary is baked into the KubeStash operator and sidecar images.

**Verifying your restic version after upgrade**

Once you upgrade to KubeStash v2026.7.10, any newly created backup job will use restic v0.5.0. You can confirm the version by inspecting the operator pod:

```bash
$ kubectl logs -n kubestash deployment/kubestash-operator \
    | grep "restic"
restic 0.5.0 compiled with go1.23 on linux/amd64
```

No changes are required to your existing `BackupConfiguration`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: daily-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: daily-snapshot
      schedule: "0 2 * * *"
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
```

The same configuration now leverages restic v0.5.0 under the hood—faster, more secure, and more reliable.

---

#### LocalBackendPVC Checker [PR](https://github.com/kubestash/apimachinery/pull/247)

When backing up to a **local backend** (e.g., a PersistentVolumeClaim mounted directly into the cluster), it's critical that the PVC is properly configured and accessible. A misconfigured PVC can lead to silent backup failures or data loss.

This release introduces a **LocalBackendPVC checker**—a validation webhook that inspects your `BackupStorage` configuration when the storage provider is set to a local backend and verifies that:

- The referenced PVC exists in the expected namespace.
- The PVC has the required access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.) to support backup operations.
- The PVC is in a `Bound` state and ready to accept data.
- Sufficient capacity is available for the expected backup volume.

The checker runs at admission time, so a misconfigured `BackupStorage` is rejected immediately with a clear error message—no more discovering PVC issues mid-backup.

**Example: BackupStorage with a local backend**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: local-nfs-storage
  namespace: demo
spec:
  storage:
    provider: local
    local:
      mountPath: /backup/data
      persistentVolumeClaim:
        claimName: nfs-backup-pvc
        ## The LocalBackendPVC checker validates this claim
        ## before the BackupStorage is admitted.
  usagePolicy:
    allowedNamespaces:
      from: All
  default: false
  deletionPolicy: WipeOut
```

If the PVC `nfs-backup-pvc` does not exist or is not bound, the `BackupStorage` creation is rejected:

```
Error: admission webhook "backupstorage.validator.kubestash.com" denied
the request: local backend PVC "nfs-backup-pvc" in namespace "demo" is
not bound (phase: Pending). Please verify the PVC is correctly provisioned.
```

This proactive validation saves troubleshooting time and prevents backups from silently failing due to storage misconfiguration.

---

#### Qdrant Backup & Restore [PR](https://github.com/kubedb/installer/pull/2359)

[Qdrant](https://qdrant.tech/) is a high-performance vector database powering AI and machine learning workloads—semantic search, recommendation engines, and RAG (Retrieval-Augmented Generation) pipelines. Losing vector embeddings means losing the intelligence layer of your application.

With v2026.7.10, KubeStash introduces **first-class backup & restore support for Qdrant**. The new `qdrant-addon` lets you back up individual collections or the entire Qdrant deployment to any KubeStash-supported backend (S3, GCS, Azure Blob, or local PVC).

**Under the hood**: The `qdrant-restic-plugin` streams collection snapshots from each Qdrant node directly to restic, with configurable parallelism (`maxParallelDump`) to balance throughput against cluster load.

**Qdrant addon definition**

```yaml
apiVersion: addons.kubestash.com/v1alpha1
kind: Addon
metadata:
  name: qdrant-addon
spec:
  backupTasks:
  - name: logical-backup
    function: qdrant-backup
    driver: Restic
    executor: Job
    singleton: true
    parameters:
    - name: collections
      usage: Comma-separated list of collections to backup. If not specified, all collections will be backed up.
      required: false
    - name: maxParallelDump
      usage: Maximum number of concurrent dump workers across all Qdrant nodes. Each worker streams one collection snapshot from a node to restic.
      required: false
      default: "10"
  restoreTasks:
  - name: logical-backup-restore
    function: qdrant-restore
    driver: Restic
    executor: Job
    singleton: true
    parameters:
    - name: collections
      usage: Comma-separated list of collections to restore. If not provided, all collections will be restored.
      required: false
```

**Backup functions**

```yaml
---
apiVersion: addons.kubestash.com/v1alpha1
kind: Function
metadata:
  name: qdrant-backup
spec:
  args:
  - backup
  - --namespace=${namespace:=default}
  - --backupsession=${backupSession:=}
  - --wait-timeout=${waitTimeout:=300}
  image: ghcr.io/kubedb/qdrant-restic-plugin:v0.1.0
---
apiVersion: addons.kubestash.com/v1alpha1
kind: Function
metadata:
  name: qdrant-restore
spec:
  args:
  - restore
  - --namespace=${namespace:=default}
  - --restoresession=${restoreSession:=}
  - --snapshot=${snapshot:=}
  image: ghcr.io/kubedb/qdrant-restic-plugin:v0.1.0
```

**BackupConfiguration for Qdrant**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: qdrant-backup-config
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Qdrant
    namespace: demo
    name: sample-qdrant
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
        name: qdrant-addon
        tasks:
          - name: logical-backup
            params:
              collections: "products,users,embeddings"
              maxParallelDump: "5"
```

**RestoreSession for Qdrant**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: qdrant-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Qdrant
    namespace: demo
    name: sample-qdrant
  dataSource:
    repository: s3-qdrant-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: qdrant-addon
    tasks:
      - name: logical-backup-restore
        params:
          collections: "products,users,embeddings"
```

With Qdrant backup support, your vector databases are now protected by the same battle-tested KubeStash pipeline—encrypted, scheduled, and ready to restore when you need them.

---

#### ClickHouse Archiver & Incremental PITR Restore [PR](https://github.com/kubedb/installer/pull/2369)

[ClickHouse](https://clickhouse.com/) support was introduced in the previous release. v2026.7.10 takes it further with two major additions: the **ClickHouse Archiver** and **incremental point-in-time recovery (PITR)**.

**ClickHouse Archiver**

The Archiver is KubeDB's continuous backup subsystem for ClickHouse. Unlike periodic full backups, the archiver captures WAL (Write-Ahead Log) segments as they are written, enabling near-zero RPO (Recovery Point Objective). The archiver is configured declaratively on the `ClickHouseVersion` catalog resource and applies across ClickHouse versions **24.4.1**, **25.7.1**, **25.12.3**, and **26.2.6**.

Each supported ClickHouse version now carries an `archiver` block that maps the full backup, manifest backup, and incremental restore tasks to the corresponding addon functions:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ClickHouseVersion
metadata:
  name: 26.2.6
spec:
  archiver:
    addon:
      name: clickhouse-addon
      tasks:
        fullBackup:
          name: logical-backup
        fullBackupRestore:
          name: logical-backup-restore
        incBackupRestore:
          name: incremental-restore
        manifestBackup:
          name: manifest-backup
        manifestRestore:
          name: manifest-restore
    walg:
      image: ghcr.io/kubedb/clickhouse-backup-plugin:v0.3.0
  db:
    image: docker.io/clickhouse/clickhouse-server:26.2.6
```

**Incremental PITR Restore**

The new `clickhouse-incremental-restore` function enables **point-in-time recovery**—you can restore a ClickHouse database to any specific timestamp, down to the second. This is critical for recovering from logical errors (e.g., a bad `DELETE` or `DROP TABLE`) without losing the data changes that happened before the mistake.

The incremental restore function accepts a `pitrTime` parameter specifying the exact recovery timestamp:

```yaml
apiVersion: addons.kubestash.com/v1alpha1
kind: Function
metadata:
  name: clickhouse-incremental-restore
spec:
  args:
    - pitr-restore
    - --namespace=${namespace:=default}
    - --restoresession=${restoreSession:=}
    - --wait-timeout=${waitTimeout:=300}
    - --snapshot=${snapshot:=}
    - --pitr-time=${pitrTime:=}
  image: ghcr.io/kubedb/clickhouse-backup-plugin:v0.3.0
```

**RestoreSession with PITR**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: clickhouse-pitr-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: sample-clickhouse
  dataSource:
    repository: s3-clickhouse-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: clickhouse-addon
    tasks:
      - name: incremental-restore
        params:
          pitrTime: "2026-07-15T14:30:00Z"
          enableCache: "true"
          scratchDir: /kubestash-tmp
```

**ClickHouse Addon — full task list**

The updated `clickhouse-addon` now exposes four distinct tasks:

| Task | Purpose |
|------|---------|
| `logical-backup` | Full ClickHouse backup via `clickhouse-backup` |
| `logical-backup-restore` | Full restore from a snapshot |
| `incremental-restore` | Point-in-time recovery using WAL archives |
| `manifest-backup` / `manifest-restore` | Backup & restore of KubeDB manifest resources |

The architecture: `logical-backup` creates a full baseline; the archiver continuously ships WAL segments; when you need to recover, `incremental-restore` replays the WAL up to the `pitrTime` you specify. The result is a ClickHouse database restored to the exact state it was in at that timestamp.

---

#### Neo4j: Exclude Databases from Backup & Restore [PR](https://github.com/kubedb/installer/pull/2373)

Large Neo4j deployments often host multiple databases within the same cluster, and not all of them need to be backed up. The `system` database, for example, contains cluster metadata that is typically recreated during restore rather than recovered from a backup.

The Neo4j addon now supports an **`excludeDatabases`** parameter on both backup and restore tasks. You can specify a comma-separated list of database names to skip, giving you fine-grained control over the backup scope.

**BackupConfiguration excluding specific Neo4j databases**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: neo4j-selective-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: neo4j-prod
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: daily-backup
      schedule: "0 2 * * *"
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
            params:
              excludeDatabases: "system,analytics"
              neo4jAdminArgs: "--compress=true,--expand-commands,--remote-address-resolution=true,--type=FULL"
```

The `excludeDatabases` parameter defaults to `"system"` for backup (skipping the system database) and `"system,neo4j"` for restore (skipping both the system database and the default `neo4j` database). This ensures that restores don't inadvertently overwrite cluster-managed databases while giving you the flexibility to override the defaults when needed.

**RestoreSession with database filtering**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: neo4j-selective-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: neo4j-prod
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
          excludeDatabases: "system,neo4j"
          seedServerName: "neo4j-prod-0"
```

---

### Improvements and Bug Fixes

#### Neo4jStats: Preserve Explicit `false` Boolean Values [PR](https://github.com/kubestash/apimachinery/pull/245)

When inspecting a Neo4j backup, KubeStash emits a `Neo4jStats` structure containing fields like `Full`, `Compressed`, and `Recovered` that report the state of the backup. Due to how Go's `omitempty` JSON tag interacts with the `bool` zero value, an explicitly `false` value was previously **dropped** from the JSON output during serialization.

This meant you could not distinguish between:
- A field that was **unset** (no data available).
- A field that was **explicitly `false`** (e.g., the backup was *not* compressed).

The fix changes these fields from `bool` to `*bool` (nullable boolean). Now, when a Neo4j backup reports `"Compressed": false`, that `false` is **preserved** in the API output, giving you an unambiguous view of your backup's integrity.

**Before (false values were dropped)**

```json
{
  "neo4jStats": {
    "Full": true
  }
}
```

Notice `Compressed` and `Recovered` are absent—were they `false` or just not reported? Impossible to tell.

**After (false values are preserved)**

```json
{
  "neo4jStats": {
    "Full": true,
    "Compressed": false,
    "Recovered": false
  }
}
```

Every field is now present, making backup inspection reports fully deterministic.

---

#### Fix Post-Restore `OnSuccess` Execution Policy [Commit](https://github.com/kubestash/kubestash/commit/95bc2d66)

KubeStash supports **post-restore hooks**—user-defined actions (e.g., sending a notification, running a schema migration, or executing a health check) that run after a restore completes. The hook's `executionPolicy` field controls when the hook is triggered:

- `Always`: Run the hook regardless of restore outcome.
- `OnSuccess`: Run the hook only when the restore succeeds.
- `OnFailure`: Run the hook only when the restore fails.

In prior versions, the `OnSuccess` execution policy was not being properly evaluated after a restore. A post-restore hook configured with `executionPolicy: OnSuccess` could be **skipped even when the restore succeeded**, or conversely, could fire when it shouldn't have.

This release corrects the policy check so that `OnSuccess` hooks fire **exactly when they should**—after a successful restore and only after a successful restore.

**Example: Post-restore hook with OnSuccess policy**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
  dataSource:
    repository: s3-postgres-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    tasks:
      - name: logical-backup-restore
  hooks:
    postRestore:
      - name: run-health-check
        executionPolicy: OnSuccess   ## Now reliably fires after a successful restore
        hookTemplate:
          podSpec:
            containers:
              - name: health-check
                image: postgres:16
                command:
                  - pg_isready
                  - -h
                  - sample-postgres.demo.svc
                  - -U
                  - postgres
            restartPolicy: Never
```

---

#### Parse Hook Pod Selector with `labels.Parse` [Commit](https://github.com/kubestash/kubestash/commit/e561d6c6)

Hooks in KubeStash can target specific pods using a **pod selector**—a label selector that determines which pod the hook job should run against. In previous versions, the pod selector was parsed with a basic string parser that did not support the full Kubernetes label selector syntax, such as set-based expressions (`key in (v1, v2)`, `key notin (v3)`, `key exists`).

This release updates the selector parsing to use the standard [`labels.Parse`](https://pkg.go.dev/k8s.io/apimachinery/pkg/labels#Parse) function from `k8s.io/apimachinery`, which correctly handles all valid Kubernetes label selector expressions.

**Example: Hook with set-based pod selector**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: postgres-with-hook
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: daily-backup
      schedule: "0 2 * * *"
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
  hooks:
    preBackup:
      - name: notify-slack
        executionPolicy: Always
        hookTemplate:
          podSpec:
            containers:
              - name: slack-notifier
                image: curlimages/curl:latest
                command:
                  - curl
                  - -X
                  - POST
                  - https://hooks.slack.com/services/xxx
                env:
                  - name: BACKUP_NAME
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.name
            restartPolicy: Never
    postBackup:
      - name: pg-verify
        executionPolicy: OnSuccess
        podSelector:
          matchExpressions:
            - key: kubedb.com/role
              operator: In
              values:
                - primary
                - standby
        hookTemplate:
          podSpec:
            containers:
              - name: verify
                image: postgres:16
                command:
                  - pg_verifybackup
                  - /backup/data
            restartPolicy: Never
```

With the `labels.Parse` fix, set-based selectors like `matchExpressions` with the `In`, `NotIn`, and `Exists` operators now work correctly, giving you precise control over which pods receive hook execution.

---

#### Installer: Document Regenerating Certified Charts [Commit](https://github.com/kubestash/installer/commit/380d4ae)

The installer repository now includes documentation in `AGENTS.md` explaining how to regenerate the `-certified` Helm charts. This is relevant for users who maintain private chart registries or need to produce customized, signed chart variants of KubeStash for air-gapped or compliance-regulated environments.

---

### Component Versions

The following components have been released as part of KubeStash v2026.7.10:

| Component | Version |
|-----------|---------|
| kubestash/apimachinery | [v0.29.0](https://github.com/kubestash/apimachinery/releases/tag/v0.29.0) |
| kubestash/kubestash | [v0.29.0](https://github.com/kubestash/kubestash/releases/tag/v0.29.0) |
| kubestash/cli | [v0.28.0](https://github.com/kubestash/cli/releases/tag/v0.28.0) |
| kubestash/kubedump | [v0.28.0](https://github.com/kubestash/kubedump/releases/tag/v0.28.0) |
| kubestash/pvc | [v0.28.0](https://github.com/kubestash/pvc/releases/tag/v0.28.0) |
| kubestash/workload | [v0.28.0](https://github.com/kubestash/workload/releases/tag/v0.28.0) |
| kubestash/manifest | [v0.21.0](https://github.com/kubestash/manifest/releases/tag/v0.21.0) |
| kubestash/volume-snapshotter | [v0.28.0](https://github.com/kubestash/volume-snapshotter/releases/tag/v0.28.0) |
| kubestash/vault | [v0.3.0](https://github.com/kubestash/vault/releases/tag/v0.3.0) |
| kubestash/installer | [v2026.7.10](https://github.com/kubestash/installer/releases/tag/v2026.7.10) |
| kubestash/docs | [v2026.7.10](https://github.com/kubestash/docs/releases/tag/v2026.7.10) |

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.7.10/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.7.10/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
