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
- oracle
- pitr
- postgres
- reconfigure
- tls
---

KubeDB **v2026.11.15** extends **KubeDB Branching** to Kafka, MSSQLServer and Neo4j, and brings **dynamic configuration** to Postgres, MySQL, MariaDB, Elasticsearch and Neo4j, so most configuration changes no longer need a pod restart.

This release also adds **automatic recovery of lost ClickHouse replicas**, **in-place point-in-time restore** for MySQL and MariaDB, **point-in-time recovery for DocumentDB**, **TLS for Milvus object storage**, **Oracle backup to S3-compatible storage**, and a set of security and reliability improvements for Postgres and its coordinator.

---

## Key Highlights

* **Branching for Kafka, MSSQLServer and Neo4j** — create an independent copy of a database for testing and development
* **Dynamic Configuration** — Postgres, MySQL, MariaDB, Elasticsearch and Neo4j apply configuration changes without restarting pods when possible
* **ClickHouse Replica Disaster Recovery** — a replica whose data volume is lost is rebuilt automatically from a healthy replica of the same shard
* **MySQL & MariaDB In-Place Archiver Restore** — restore an existing database to a point in time with the new `ArchiverRestore` ops request
* **DocumentDB PITR** — continuous archiving and point-in-time recovery via `DocumentDBArchiver`
* **Milvus** — TLS for object storage with a private CA, and backup with PITR for Milvus `3.0.0`
* **Oracle Backup to S3-Compatible Storage** — RMAN backups now work with S3-compatible object stores, over HTTPS or plain HTTP
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

**Postgres, MySQL, MariaDB, Elasticsearch, and Neo4j** now support dynamic configuration. Configuration changes can be applied without restarting database pods. `spec.configuration.restart` controls the restart behavior:

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

- **Detection:** the operator records the UID of every data PVC. A PVC that comes back with the same name but a new UID is treated as a lost replica. New PVCs created by a scale-up are not treated as lost replicas.
- **Donor selection:** once the new pod is running and ClickHouse has finished loading its tables, KubeDB picks a healthy replica of the same shard as the donor. The replica being rebuilt is never used as a donor.
- **Schema replay:** KubeDB copies the donor's databases, tables and materialized views onto the empty replica, in dependency order, and keeps each table's UUID so the replica joins the existing replication group. `ReplicatedMergeTree` tables then fetch their data from the other replicas.
- **Keeper cleanup:** stale replica registrations left in ClickHouse Keeper by the old pod, including those of `Replicated` databases, are removed first, so the new replica can register without `REPLICA_ALREADY_EXISTS` errors.

Recovery runs only for clusters that can rebuild from a sibling: cluster topology, durable storage, and at least **two replicas per shard**. Objects whose DDL contains a credential (any `PASSWORD` clause) cannot be read back from ClickHouse, so they are skipped and named in the operator log; recreate them manually.

The same Keeper cleanup now runs during horizontal scaling, which prevents data loss during shard scaling and lets a reused pod ordinal register again cleanly.

---

## MySQL

### In-Place Archiver Restore

MySQL can now restore an existing database to a point in time **in place**, using the new `ArchiverRestore` ops request.

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

**Note:** archiving stays disabled after the restore, because the restore forks the binlog history. Re-enable it once you have checked the restored data.

---

## MariaDB

### In-Place Archiver Restore

MariaDB can now restore an existing database to a point in time **in place**, using the new `ArchiverRestore` ops request.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mariadb-inplace-restore
  namespace: demo
spec:
  type: ArchiverRestore
  timeout: 20m
  databaseRef:
    name: mariadb
  archiver:
    recoveryTimestamp: "2026-11-15T06:38:42Z"
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
    fullDBRepository:
      name: mariadb-full
      namespace: demo
    manifestRepository:
      name: mariadb-manifest
      namespace: demo
    # true (the default) keeps the pre-restore PersistentVolumes after a successful restore.
    retainPV: true
```

As with MySQL, the request suspends archiving, keeps the old volumes (when `retainPV` is `true`), wipes the data volumes, restores the base backup, and replays binlogs up to `recoveryTimestamp`. Archiving stays disabled after the restore; re-enable it once you have checked the restored data.

### Other Improvements

- **Primary service template:** in replication mode, the `primary` service template from `spec.serviceTemplates` is applied to the MaxScale Service by default. Set the `kubedb.com/primary-service-template` annotation on the MariaDB to `mariadb` to apply it to the MariaDB primary Service instead, or `both` to apply it to both.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
  annotations:
    kubedb.com/primary-service-template: mariadb # maxscale (default) | mariadb | both
```

---

## Postgres

### Operator

- **Security:** opt-in service account token hardening, plus token authentication and TLS for the coordinator API.
- **Reconfigure:** with `restart: auto`, KubeDB decides whether a restart is needed and verifies that PostgreSQL has reloaded the change.
- **Major version upgrade:** the upgrade now holds a Raft key while it runs.
- **OpenShift:** the Postgres UID, GID and fsGroup now default to the namespace's SCC ranges.
- **Startup:** Issuer events received before the Postgres watchers start are now ignored.
- **RHEL-based images:** `run.sh` is now located through `/proc`, so these images start correctly.

### pg-coordinator

- **API security:** the coordinator API now supports token authentication and opt-in TLS, and the `pprof` endpoint is gated.
- **Re-seed:** an override file can now skip the free-space check before a re-seed.
- **Crash recovery:** a crashed standby now restarts in standby mode instead of single-user crash recovery.
- **Major upgrade:** a re-seed during a major version upgrade no longer keeps the old data.

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

Milvus can now connect to an HTTPS S3-compatible object storage that uses a private CA. Set `spec.objectStorage.tls.caSecret` to the Secret that holds the CA certificate. KubeDB mounts the CA into Milvus and turns on SSL automatically, so you don't need `useSSL` in the storage secret.

The following is an example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mv-tls-storage
  namespace: demo
type: Opaque
stringData:
  address: object-storage.storage.svc.cluster.local # HTTPS endpoint of the object storage
  port: "443"
  accessKeyID: <access-key-id>
  secretAccessKey: <secret-access-key>
  bucketName: milvus-data
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
        name: object-storage-ca
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

## Oracle

### Backup to S3-Compatible Storage

Oracle RMAN backup, added in the previous release for AWS S3, now also works with **S3-compatible object stores** (for example Ceph RGW or SeaweedFS), over both HTTPS and plain HTTP.

- **Endpoint detection:** KubeDB builds the OSBWS wallet from the `BackupStorage` endpoint. AWS and HTTPS endpoints use the Oracle installer, which imports the endpoint's TLS certificate. Plain HTTP endpoints (for example an in-cluster object store) get a credential-only wallet instead.
- **Addressing style:** path-style addressing (`endpoint/bucket`), which most S3-compatible stores need, is now the default. Virtual-hosted style (`bucket.endpoint`) is kept for AWS S3 endpoints, so existing AWS setups behave as before.
- **Override:** set the `oracle.kubedb.com/osbws-virtual-host` annotation on the `BackupStorage` to `"true"` or `"false"` to force one style.

The following is a `BackupStorage` for an in-cluster S3-compatible store. The rest of the backup setup is the same as for AWS S3:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-compatible-storage
  namespace: demo
  # Optional: force the S3 addressing style instead of detecting it from the endpoint.
  # annotations:
  #   oracle.kubedb.com/osbws-virtual-host: "false"
spec:
  storage:
    provider: s3
    s3:
      endpoint: http://object-store.storage.svc.cluster.local:8080
      region: us-east-1
      bucket: oracle-backup
      secretName: s3-compatible-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: WipeOut
```

The credential secret uses the same keys as for AWS: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

---

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
