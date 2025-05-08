---
title: Announcing KubeDB v2025.4.30
date: "2025-05-05"
weight: 16
authors:
- Arnob Kumar Saha
tags:
- alert
- archiver
- backup
- cassandra
- catalog-manager
- clickhouse
- cloud-native
- dashboard
- database
- dns
- grafana
- kafka
- kubedb
- kubernetes
- kubestash
- memcached
- mongodb
- mssqlserver
- network
- percona-xtradb
- pgbouncer
- postgres
- postgresql
- prometheus
- proxysql
- redis
- restore
- s3
- security
- singlestore
- solr
- srv
- tls
- valkey
- voyager-gateway
- zookeeper
---

KubeDB **v2025.4.30** is here, delivering enhanced performance, expanded database support, and streamlined management for Kubernetes-based database deployments. This release introduces new features, improved reliability, and broader GitOps integration, making database operations more efficient and production-ready.

## Key Changes
- **New Database Support**: Added support for **Apache Ignite**, a powerful in-memory computing platform for high-performance, low-latency workloads.
- **Expanded GitOps Support**: Extended GitOps capabilities to include Redis, Redis Sentinel, and MariaDB, alongside existing support for PostgreSQL, MongoDB, and Kafka.
- **Virtual Secrets Enhancements**: Improved integration with external secret managers, including a bug fix for PostgreSQL Virtual Secrets deletion.
- **New Versions**: Added support for new versions of various DBs.

## Cassandra
KubeDB now supports backup and restore for **Cassandra** databases using KubeStash and the Medusa plugin. This enables reliable data protection with cloud storage backends (e.g., S3, GCS).

### Backup and Restore Workflow
- **BackupStorage**: Specifies the cloud storage backend.
- **RetentionPolicy**: Defines how long backup data is retained.
- **Secrets**: Stores backend access credentials.
- **BackupConfiguration**: Configures the target database, backend, and addon.
- **RestoreSession**: Restores data from a specified snapshot.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: cass-backup
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
  backends:
    - name: s3-backend
      storageRef:
        namespace: default
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: default
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-cassandra-repo
          backend: s3-backend
          directory: /cassandra
      addon:
        name: cassandra-addon
        tasks:
          - name: logical-backup
        jobTemplate:
          spec:
            serviceAccountName: cluster-resource-reader
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-cas
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
  dataSource:
    repository: s3-cassandra-repo
    snapshot: s3-cassandra-repo-cass-backup-frequent-backup-1744202600
  addon:
    name: cassandra-addon
    tasks:
      - name: logical-backup-restore
    jobTemplate:
      spec:
        serviceAccountName: cluster-resource-reader
```

## Apache Ignite
We’re excited to introduce support for **Apache Ignite**, an in-memory computing platform designed for high-speed transactions, real-time analytics, and hybrid transactional/analytical processing (HTAP). Ignite offers distributed caching, durable storage, distributed SQL queries, and co-located processing for low-latency, high-throughput workloads. Key features include:
- **Cluster Mode Provisioning**: Deploy Ignite clusters with ease using KubeDB.
- **Custom Configuration**: Support for custom configurations via Kubernetes secrets.
- **Authentication**: Enhanced security with authentication support.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "2.17.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
**Supported Version**: 2.17.0

## Microsoft SQL Server
This release introduces support for **active and passive secondary replicas** in **Microsoft SQL Server** Availability Groups, enabling cost-efficient deployments by supporting passive replicas that avoid licensing costs.

### Active/Passive Secondary Replicas
- **secondaryAccessMode**: A new field in the `MSSQLServer` CRD under `spec.topology.availabilityGroup` allows control over secondary replica connection modes:
  - **Passive**: No client connections (default, ideal for DR or failover without licensing costs).
  - **ReadOnly**: Accepts read-intent connections only.
  - **All**: Allows all read-only connections.

```yaml
spec:
  topology:
    availabilityGroup:
      secondaryAccessMode: Passive | ReadOnly | All
```

**T-SQL Mapping**:
- **Passive**: `SECONDARY_ROLE (ALLOW_CONNECTIONS = NO)`
- **ReadOnly**: `SECONDARY_ROLE (ALLOW_CONNECTIONS = READ_ONLY)`
- **All**: `SECONDARY_ROLE (ALLOW_CONNECTIONS = ALL)`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ag-passive
  namespace: demo
spec:
  version: "2022-cu16"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup: 
      databases:
        - agdb1
        - agdb2
      secondaryAccess: "Passive"
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
      - name: mssql
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: Evaluation
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## MongoDB
### `mongodb+srv` style connection url support
From this release, users can connect to the mongodb using `mongosh "mongodb+srv://<user>:<pass>@<host>/<database>?tls=true&tlsCAFile=./db-ca.crt"` connection uri. We have added `db.spec.replicaSet.horizons` field to specify the server dns & pod CNAME records. Here is a sample:

```yaml
...
  replicaSet:
    horizons:
      dns: kubedb.cloud
      pods:
      - mongo-0.kubedb.cloud
      - mongo-1.kubedb.cloud
      - mongo-2.kubedb.cloud
```

To achieve this feature, there are multiple other operators needed. Specially the [voyager-gateway](https://github.com/voyagermesh/installer/tree/v2025.4.30/charts/voyager-gateway) & [catalog-manager](https://github.com/appscode-cloud/installer/tree/release-v2025.4.30/charts/catalog-manager). Catalog-manager will act as the co-ordinator here among these various components. Attaching a [gist](https://gist.github.com/ArnobKumarSaha/affc2d5113442e19e516111c41338c51) here which describes the steps & commands. 

## MySQL 
### Configuration Improvements
This release enhances **MySQL** configuration flexibility by allowing customization of key system variables:
- **innodb_buffer_pool_size**
- **group_replication_message_cache_size**
- **binlog_expire_logs_seconds** (default: 7 days)

### Default Values Based on System RAM
- **RAM ≤ 1 GiB**:
  - `innodb_buffer_pool_size`: 128 MiB
  - `group_replication_message_cache_size`: 128 MiB
- **RAM ≤ 2 GiB**:
  - `innodb_buffer_pool_size`: 256 MiB
  - `group_replication_message_cache_size`: 256 MiB
- **RAM ≤ 4 GiB**:
  - `innodb_buffer_pool_size`: RAM - 1024 MiB
  - `group_replication_message_cache_size`: 256 MiB
- **RAM > 4 GiB**:
  - `innodb_buffer_pool_size`: 60% of RAM
  - `group_replication_message_cache_size`: (RAM × 0.25 - 256 MiB) × 0.40

## PostgreSQL
### Virtual Secrets Bug Fix
- **Bug Fix**: Resolved an issue with PostgreSQL Virtual Secrets where data deletion from the external Secret Manager was not handled correctly.


## ProxySQL
KubeDB now supports **PerconaXtraDB Galera Clusters** with **ProxySQL**, providing robust high-availability, query routing, and load balancing.

### Galera Cluster Support
- **Galera Integration**: ProxySQL manages traffic across PerconaXtraDB Galera nodes.
- **Automatic Hostgroup Assignment**: Dynamically assigns nodes to read/write hostgroups.
- **Health Checks**: Automatically removes faulty nodes from routing with graceful failover.

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: xtradb-galera
  namespace: demo
spec:
  version: "8.4.3"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: xtradb-proxy
  namespace: demo
spec:
  version: "2.6.3-debian"
  replicas: 3
  syncUsers: false
  backend:
    name: xtradb-galera
  deletionPolicy: WipeOut
```

**ProxySQL Galera Hostgroup Mapping**:
```sql
SELECT * FROM mysql_galera_hostgroups;
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind | comment |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| 2                | 4                       | 3                | 1                 | 1      | 1           | 1                     | 0                       |         |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
```

**Registered Nodes in runtime_mysql_servers**:
```sql
SELECT * FROM runtime_mysql_servers;
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                    | port | gtid_port | status  | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
```

- **New Version**: Added support for **ProxySQLVersion 2.7.3**.

## Redis and Valkey
### Redis Versioning
A new field in the `RedisVersion` CRD distinguishes **Redis** from **Valkey**, ensuring compatibility and clarity.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: RedisVersion
metadata:
  name: <>
spec:
  # all other fields
  distribution: Official # Possible values: Official, Valkey
```

### Valkey Support
KubeDB now supports **Valkey**, a high-performance Redis fork, with new images for versions **8.1.1**, **8.0.3**, **7.2.9**, and **7.2.5**. Valkey is provisioned under the `Redis` CRD for consistency.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: RedisVersion
metadata:
  name: 7.2.9
spec:
  coordinator:
    image: ghcr.io/kubedb/redis-coordinator:v0.32.0
  db:
    image: ghcr.io/appscode-images/valkey:7.2.9
  distribution: Valkey
  exporter:
    image: ghcr.io/kubedb/redis_exporter:1.66.0
  initContainer:
    image: ghcr.io/kubedb/redis-init:0.10.0
  podSecurityPolicies:
    databasePolicyName: redis-db
  securityContext:
    runAsUser: 999
  updateConstraints:
    allowlist:
      - 7.2.9
  version: 7.2.9
```

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: valkey
  namespace: demo
spec:
  version: 8.1.1
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: Halt
```


## GitOps Enhancements
GitOps support now extends to **Redis**, **Redis Sentinel**, and **MariaDB**, in addition to PostgreSQL, MongoDB, and Kafka. The `gitops.kubedb.com/v1alpha1` CRD enables seamless integration with GitOps tools like ArgoCD and Flux, automating database provisioning and reconciliation.

### GitOps Features
- **Unified CRD**: The `gitops.kubedb.com/v1alpha1` CRD mirrors core KubeDB CRs, enabling consistent database management.
- **Automated Provisioning**: GitOps pipelines create corresponding KubeDB CRs to provision databases based on Git-defined states.
- **Smart Reconciliation**: The GitOps controller detects and resolves discrepancies in database configurations, replicas, or versions.

We have also changed the way to enable gitops feature. Pass `--set kubedb-crd-manager.installGitOpsCRDs=true` in the kubedb installation process to enable this.

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).