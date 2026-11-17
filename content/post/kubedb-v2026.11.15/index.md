---
title: Announcing KubeDB v2026.11.15
date: "2026-11-15"
weight: 15
authors:
- SK Ali Arman
tags:
- archiver
- backup
- clickhouse
- cloud-native
- database
- disaster-recovery
- documentdb
- elasticsearch
- kafka
- kubedb
- kubernetes
- mariadb
- milvus
- mssqlserver
- mysql
- neo4j
- pitr
- postgres
- reconfigure
- tls
---

KubeDB **v2026.11.15** extends **KubeDB Branching** to Kafka, MSSQLServer and Neo4j, and brings **dynamic configuration** to MySQL, MariaDB, Elasticsearch and Neo4j, so most configuration changes no longer need a pod restart.

This release also adds **automatic recovery of lost ClickHouse replicas**, **in-place point-in-time restore** for MySQL and MariaDB, **point-in-time recovery for DocumentDB**, **TLS for Milvus object storage**, and a set of security and reliability improvements for Postgres and its coordinator.

---

## Key Highlights

* **Branching for Kafka, MSSQLServer and Neo4j** — create an independent copy of a database for testing and development
* **Dynamic Configuration** — MySQL, MariaDB, Elasticsearch and Neo4j apply configuration changes without restarting pods when possible, controlled by `spec.configuration.restart`
* **ClickHouse Replica Disaster Recovery** — a replica whose data volume is lost is rebuilt automatically from a healthy replica of the same shard
* **MySQL & MariaDB In-Place Archiver Restore** — restore an existing database to a point in time with the new `ArchiverRestore` ops request
* **DocumentDB PITR** — continuous archiving and point-in-time recovery via `DocumentDBArchiver`
* **Milvus** — TLS for object storage with a private CA, and backup with PITR for Milvus `3.0.0`
* **Postgres** — opt-in service account token hardening, token auth and TLS for the coordinator API, and smarter reconfigure restarts

---

## KubeDB Branching

[KubeDB Branching](https://kubedb.com/docs/v2026.11.15/operatormanual/branching/) now supports **Kafka, MSSQLServer, and Neo4j**, in addition to the databases supported earlier. A branch is an independent copy of your database that you can use for testing and development without affecting the source.

**Prerequisite:** The source storage class must support CSI snapshots.

The following is an example for Kafka:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: prod-kf-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: prod-kf
    namespace: demo
  target:
    namespace: demo
    name: prod-kf-dev
    storageClassName: ceph-rbd
  deletionPolicy: Delete
```

---

## Dynamic Configuration

**MySQL, MariaDB, Elasticsearch, and Neo4j** now support dynamic configuration. Configuration changes can be applied without restarting database pods. `spec.configuration.restart` controls the restart behavior:

- `auto` (default): KubeDB checks which parameters you changed. If all of them can be applied on a running database it applies them with no restart. It restarts the pods only if at least one changed parameter needs a restart.
- `false`: Never restart. KubeDB applies the changes to the running database. If any changed parameter is static the OpsRequest fails.
- `true`: Always restart the pods after applying the changes, even if every change could have been applied dynamically.

The following is an example for MySQL:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: mysql-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mysql
  configuration:
    restart: "false"
    applyConfig:
      my.cnf: |
        [mysqld]
        max_connections=300
```

---

## ClickHouse

### Disaster Recovery for Shard Replicas

If a ClickHouse replica loses its data volume (for example, the PVC is deleted, or the node and its local disk are gone), the pod comes back empty. Before this release it rejoined the cluster with no tables, and the shard stayed degraded until someone rebuilt the schema by hand.

KubeDB now recovers such a replica automatically:

- **Detection:** the operator records the UID of every data PVC. A PVC that comes back with the same name but a new UID is treated as a lost replica. Scale-up PVCs are not mistaken for lost replicas.
- **Donor selection:** once the new pod is running and ClickHouse has finished loading its tables, KubeDB picks a healthy replica of the same shard as the donor. The replica being rebuilt is never used as a donor.
- **Schema replay:** KubeDB copies the donor's databases, tables and materialized views onto the empty replica, in dependency order, and keeps each table's UUID so the replica joins the existing replication group. `ReplicatedMergeTree` tables then fetch their data from the other replicas.
- **Keeper cleanup:** stale replica registrations left in ClickHouse Keeper by the old pod, including those of `Replicated` databases, are removed first, so the new replica can register without `REPLICA_ALREADY_EXISTS` errors.

Recovery runs only for clusters that can rebuild from a sibling: cluster topology, durable storage, and at least **two replicas per shard**. Objects whose DDL contains a credential (any `PASSWORD` clause) cannot be read back from ClickHouse, so they are skipped and named in the operator log; recreate them manually.

The same Keeper cleanup now runs during horizontal scaling, which fixes a possible data loss during shard scaling and lets a reused pod ordinal register again cleanly.

---

## MySQL and MariaDB

### In-Place Archiver Restore

MySQL and MariaDB can now restore an existing database to a point in time **in place**, using the new `ArchiverRestore` ops request.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: mysql-inplace-restore
  namespace: demo
spec:
  type: ArchiverRestore
  databaseRef:
    name: mysql
  # Required: the database is not Ready while its volumes are being restored.
  apply: Always
  # Size this to your database: base backup restore, binlog replay and re-seeding the other members.
  timeout: 30m
  archiver:
    recoveryTimestamp: "2026-11-15T06:38:42Z"
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
    fullDBRepository:
      name: mysql-full
      namespace: demo
    replicationStrategy: sync
    # true (the default) keeps the pre-restore PersistentVolumes after a successful restore.
    retainPV: true
```

The request suspends archiving, keeps the old volumes (when `retainPV` is `true`), wipes the data volumes, restores the base backup, and replays binlogs up to `recoveryTimestamp`. Each step is recorded as a condition on the ops request.

**Note:** archiving stays disabled after the restore, because the restore forks the binlog history. Re-enable it once you have checked the restored data. MariaDB uses the same request type through `MariaDBOpsRequest`.

---

## Postgres

### Operator

- Opt-in service account token hardening, plus token auth and TLS for the coordinator API.
- `Reconfigure` with restart `auto` decides on its own whether a restart is needed, and checks that the change was reloaded through PostgreSQL.
- Major version upgrades set a Raft key for their duration.
- On OpenShift, Postgres UID, GID and fsGroup default to the namespace's SCC ranges.
- Issuer events that arrive before the Postgres watchers start are ignored.
- `run.sh` is found via `/proc` on RHEL-family images.
- PITR restores work when the archiver repository is in another namespace.

### pg-coordinator

- The coordinator API has token auth and opt-in TLS; `pprof` is gated.
- An override file can skip the re-seed free-space check.
- A crashed standby restarts in standby mode, not single-user crash recovery.
- A re-seed during a major upgrade no longer keeps the old data.

---

## DocumentDB

### Point-in-Time Recovery

DocumentDB now supports continuous archiving and **Point-in-Time Recovery (PITR)**. KubeDB takes a full base backup with KubeStash and streams WAL to your backup storage, so you can restore the database to any second within the retention window.

**Prerequisite:** A KubeStash `BackupStorage`, a `RetentionPolicy` and an encryption `Secret` must exist.

Create a `DocumentDBArchiver`. It selects databases by label and namespace:

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
      schedule: "0 0 * * *"
  manifestBackup:
    scheduler:
      schedule: "0 0 * * *"
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

Point the DocumentDB at the archiver. `archive_timeout` forces a WAL switch every 60 seconds, so recent writes reach the backup storage even when traffic is low:

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

To restore, create a new DocumentDB with `spec.init.archiver`. Set `recoveryTimestamp` (RFC3339) to the time you want to recover to:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb-pitr-restored
  namespace: demo
spec:
  version: pg17-0.109.0
  replicas: 1
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "2026-11-15T11:10:22Z"
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

---

## Milvus

### TLS for Object Storage

Milvus can now connect to an HTTPS object storage (MinIO / S3-compatible) that uses a private CA. Set `spec.objectStorage.tls.caSecret` to the Secret that holds the CA certificate. KubeDB mounts the CA into Milvus and turns on SSL automatically, so you don't need `useSSL` in the storage secret.

The following is an example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mv-tls-storage
  namespace: demo
stringData:
  address: minio-tls.demo.svc.cluster.local
  port: "9000"
  accessKeyID: minio
  secretAccessKey: minio123
  bucketName: tls-bucket
  rootPath: mv-tls
---
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: mv-tls
  namespace: demo
spec:
  version: "2.6.15"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: mv-tls-storage
    tls:
      caSecret:
        name: minio-tls-ca
        key: ca.crt
  metaStorage:
    size: 1
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Without `spec.objectStorage.tls`, Milvus does not trust the private CA and fails with `x509: certificate signed by unknown authority`.

### Other Improvements

- Backup and point-in-time recovery now support Milvus `3.0.0`.

---

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
