---
title: Announcing KubeDB v2026.2.26
date: "2026-02-26"
weight: 15
authors:
- Saurov Chandra Biswas
tags:
- ai
- cloud-native
- configuration
- database
- elasticsearch
- gitops
- hanadb
- kubedb
- kubernetes
- mariadb
- migration
- milvus
- mysql
- neo4j
- oracle
- postgres
- qdrant
- redis
- vector-database
- weaviate
---

KubeDB **v2026.2.26** focuses on **operational efficiency, GitOps safety, high-availability improvements, and advanced recovery capabilities** across the entire database ecosystem. This release enhances OpsRequest execution speed, improves TLS and configuration automation, introduces PostgreSQL read replicas and migration tooling, expands monitoring support, and strengthens distributed database capabilities.

From enterprise databases like **Oracle and HanaDB** to modern vector and graph databases like **Milvus, Qdrant, and Neo4j**, this release significantly improves reliability, observability, and production readiness.

---

## Key Highlights

* Faster OpsRequest execution with reduced waiting time
* Smarter GitOps-driven configuration & TLS reconfiguration
* PostgreSQL Read Replica support
* PostgreSQL Migration Tool for seamless migrations with minimal downtime
* WAL-log retention automation for Postgres
* PITR restore to the same database
* Distributed MariaDB PITR support
* Native monitoring for Neo4j, Qdrant, and Oracle
* Milvus Distributed Mode support
* HanaDB System Replication clustering
* New PostgreSQL, Milvus versions

---

## GitOps Improvements

This release improves GitOps workflows by making configuration updates, TLS reconfiguration, and monitoring-related changes more reliable across all supported databases.

### Configuration Updates:
Fixed generation and deduplication of `Reconfigure OpsRequests` for `inline configuration` changes across all databases.

### TLS Reconfiguration Enhancements:
Extended TLS reconfiguration to correctly detect database-specific field changes:
- MySQL / MariaDB: Changes to `RequireSSL` now trigger TLS reconfigure OpsRequests, and `RequireSSL` is included in TLS ops.
- PostgreSQL: Changes to `SSLMode` now trigger TLS reconfigure OpsRequests.
- MSSQL: Changes to `ClientTLS` now trigger TLS reconfigure OpsRequests.

### Restart Ops:
Restart OpsRequests are now automatically triggered when monitor specifications are removed.

This update makes GitOps-driven database management safer, more predictable, and operationally reliable.


## PostgreSQL

### Read Replica Support

We've added support for **read replicas** in our Postgres setup.

This feature is inspired by how companies like OpenAI handle huge amounts of read traffic (as explained in their recent [PostgreSQL scaling blog](https://openai.com/index/scaling-postgresql/)).

#### The simple idea

In a normal setup:
- The primary database handles both writes (changes) and reads.
- Standby replicas are also ready to become primary during failover — so they usually get the same strong CPU and memory as the primary.
- This works great for high availability, but it's expensive when you have **a lot more reads than writes**.

When read traffic grows a lot, you often want:
- A few strong replicas (2–3) that stay ready for failover (same resources as primary).
- But **many more** replicas just for reading — these don't need to handle writes or be failover candidates, so they can run with much less CPU and memory.

#### Benefits of read replicas (the new feature)

- You can create separate groups of read-only replicas.
- Each group can have its own CPU and memory settings — usually much smaller than the primary.
- This saves a lot of money (no need to give every read replica full production resources).
- You can send different kinds of read queries to different replica groups (e.g., one group for reporting, another for user-facing apps).
- Failover still works normally using your small number of full-strength standbys ensuring high availability.

In short: Keep a few powerful replicas for safety/failover, and add lots of lightweight replicas just for fast, cheap reads — exactly like OpenAI does to serve massive traffic efficiently.


```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  readReplicas: 
    - name: reporting  # group 1, it has a service named ha-postgres-rr-reporting
      replicas: 2
    - name: user-facing # group 2, has a service named ha-postgres-rr-user-facing
      replicas: 3
    - name: analytics # group 3, has a service named ha-postgres-rr-analytics
      replicas: 5
  ...
```

You can direct traffic to any of the group you want using the corresponding service name (e.g., `ha-postgres-rr-reporting` for the reporting group). Each group can have its own resource settings, so you can optimize costs based on the read workload of each group.

you can also set below fields in the `.spec.readReplicas[i]`


```yaml
# ────────────────────────────────────────────────
  # 3. Resources (sidecar or member container resources)
  resources:
    limits:
      cpu: "800m"
      memory: "2Gi"
    requests:
      cpu: "400m"
      memory: "1Gi"

  # ────────────────────────────────────────────────
  # 4. NodeSelector
  nodeSelector:
    kubernetes.io/hostname: worker-node-05
    disktype: ssd
    region: ap-south-1

  # ────────────────────────────────────────────────
  # 5. Tolerations
  tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "postgres"
      effect: "NoSchedule"
    - key: "node.kubernetes.io/unreachable"
      operator: "Exists"
      effect: "NoExecute"
      tolerationSeconds: 300

  # ────────────────────────────────────────────────
  # 6. StorageType
  storageType: Durable              # or: Ephemeral

  # ────────────────────────────────────────────────
  # 7. Storage (full PVC spec)
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
    storageClassName: gp3             # or "standard", "longhorn", etc.
    selector:
      matchLabels:
        purpose: fast-storage
    volumeMode: Filesystem

  # ────────────────────────────────────────────────
  # 8. PodPlacementPolicy (reference to a PlacementPolicy CR)
  podPlacementPolicy:
    name: postgres-preferred     

  # ────────────────────────────────────────────────
  # 9. ServiceTemplate for modifying the read replica service
```

#### Scaling Read Replicas

Horizontal and vertical scaling of read replicas is supported using `PostgresOpsRequest`. You can increase or decrease the number of replicas in each group, and adjust their resource settings as needed. This allows you to easily adapt to changing read workloads without affecting the primary or standby replicas.

**Vertical Scaling**: You can update the CPU and memory resources allocated to the read replicas to handle increased read traffic or reduce costs during low traffic periods.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-vertical
  namespace: demo
spec:
  apply: Always
  type: VerticalScaling
  databaseRef:
    name: ha-postgres
  verticalScaling:
    readReplicas:
    - name: reporting
      postgres:
        resources:
          requests:
            memory: 512Mi
            cpu: 100m
```

Here, you need to specify the name of the read replica group you want to scale (e.g., `reporting` in this example, multiple read replica group can be scaled in a single OpsRequest) and the new resource settings for that group.

**Horizontal Scaling**: You can increase or decrease the number of read replicas in each group to handle changes in read traffic.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-horizontal-slot
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ha-postgres
  horizontalScaling:
    readReplicas:
      - name: reporting
        replicas: 8
      - name: user-facing
        replicas: 2
```

In this example, the `reporting` group is scaled up to 8 replicas, while the `user-facing` group is scaled down to 2 replicas. You can adjust the number of replicas in each group based on your read workload requirements.

### PostgreSQL Migration Tool

We are excited to introduce our new PostgreSQL migration tool, designed to simplify and streamline database migrations to KubeDB-managed PostgreSQL instances. This tool enables seamless migration of PostgreSQL databases from any source environment—including Amazon RDS, CloudNativePG (CNPG), Zalando PostgreSQL Operator, Bitnami Helm charts, and self-hosted PostgreSQL instances—to KubeDB-managed PostgreSQL clusters with **minimal downtime**.

The migration tool performs live migrations while the source database remains operational, ensuring continuous availability throughout the process. A brief downtime occurs only during the final switchover when application endpoints are redirected to the target database. The tool operates in three distinct phases:

**Migration Phases:**

1. **Schema Migration** - Transfers the complete database schema (tables, indexes, constraints, etc.) from the source to the target KubeDB PostgreSQL instance.

2. **Snapshot Migration** - Performs an initial bulk data transfer, copying all existing data from the source database to the target.

3. **Streaming (CDC)** - Continuously replicates ongoing changes from the source database to the target using Change Data Capture (CDC), ensuring the target stays synchronized with the source in real-time.

**Prerequisite:**
Source postgres `wal_level` should be `logical` and the provided user should have privilege to perform logical replication.

**Apply:**
First deploy a `kubedb` managed `postgresql` by [quickstart](https://kubedb.com/docs/v2026.1.19/guides/postgres/quickstart/quickstart/) guide.

Then apply the following `Migrator` CR:

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: postgres-migrate
  namespace: migration
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    postgres:
      connectionInfo:
        url: "postgresql://<username>:<password>@<host>:<port>/<dbname>"
        maxConnections: 100
      pgDump:
        schemaOnly: true
      logicalReplication:
        copyData: true
        publication:
          name: "pub"
          mode: "default"
        subscription:
          name: "sub"
  target:
    postgres:
      connectionInfo:
        dbName: test
        appBinding:
          name: target-postgres
          namespace: migration
        maxConnections: 100
```

> You can provide the connection information in two ways: 'connectionInfo.url' expect a postgres connection string and `connectionInfo.appBinding` expect the `AppBinding` in `KubeDB` ecosystem. For external postgres `AppBinding` should be created manually and for `kubedb` managed postgres `AppBinding` is automatically created.


### PITR Restore to the same database

From this release, you can PITR to the same database. Contact to the kubedb team for doing pitr on the same database.


### New Postgres WAL-Log Retention Feature for PITR

We’re excited to announce a new WAL-log retention management feature in our Postgres operator, designed to give users more control over their log storage.


#### Automatic WAL-log Deletion

Previous WAL-log files can now be automatically deleted based on user-defined retention policies. This process runs through the sidekick pod.

#### Flexible Retention Policies
In `postgresArchiver.spec.logBackup`, you can now define:

- logRetentionHistoryLimit: Number of retention stats to keep (e.g., 5)

- retentionPeriod: Duration after which old WAL files are deleted (e.g., 10d), default is 1 year

- retentionSchedule: Cron-style schedule for running the deletion process (e.g., "*/15 * * * *"), default is every 3 months

The system maintains the last logRetentionHistoryLimit retention stats in your incremental snapshots. WAL files older than the retentionPeriod will be automatically removed.


If you change any of these retention fields, a sidekick pod restart is required for the new configuration to take effect.

Example archiver yaml:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: postgresarchiver-sample
  namespace: aws-demo12
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
         kubernetes.io/metadata.name: aws-demo12
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: postgres-retention-policy
    namespace: aws-demo12
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "aws-demo12"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/3 * * * *"
    sessionHistoryLimit: 2
  logBackup:
    logRetentionHistoryLimit: 5
    retentionPeriod: 10d
    retentionSchedule: "*/1 * * * *"
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/3 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "s3-storage"
      namespace: "backupstorage"

```


---

## Redis

In this release we fixed the RedisOpsRequest issue where it failed while restarting pods and
`Reconfigure-TLS` failing issue for sentinel mode.


---

## MariaDB

In this release, we have introduced support for PITR (Point-in-Time Recovery) backup and restore for distributed databases.
The distributed MariaDB archiver now supports both **VolumeSnapshotter**-based and **Restic** physical backups.
To use VolumeSnapshot as the base backup method, you must:

- Install the Volume Snapshotter controller on your spoke cluster, and
- Ensure a suitable VolumeSnapshotClass already exists on the spoke cluster.

This enables reliable point-in-time recovery capabilities across distributed MariaDB deployments.

---

## MongoDB

In this release we started to use the MaxConnectionIdleTime(30 seconds) & MaxPoolSize(20) for the mongodb client calls, to limit the total number of connections created by the kubedb provisioner for health-check purposes.  This avoids unnecessary memory consumption for the idle connections eventually.

---
## MySQL

### New MySQL bin-Log Retention Feature

We’re excited to announce a new bin-log retention management feature in our MySQL operator, designed to give users more control over their binlog storage.


#### Automatic bin-log Deletion
Previous bin-log files can now be automatically deleted based on user-defined retention policies. This process runs through the sidekick pod.

#### Flexible Retention Policies
In `mysqlArchiver.spec.logBackup`, you can now define:

- logRetentionHistoryLimit: Number of retention stats to keep (e.g., 5)

- retentionPeriod: Duration after which old bin-log files are deleted (e.g., 10d), default is 1 year

- retentionSchedule: Cron-style schedule for running the deletion process (e.g., "*/15 * * * *"), default is every 3 months

The system maintains the last logRetentionHistoryLimit retention stats in your incremental snapshots. binlog files older than the retentionPeriod will be automatically removed.


If you change any of these retention fields, a sidekick pod restart is required for the new configuration to take effect.


```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MySQLArchiver
metadata:
  name: mysqlarchiver-sample
  namespace: aws-demo11
spec:
  backupStorage:
    ref:
      name: s3-storage
      namespace: aws-demo11
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: aws-demo11
    selector:
      matchLabels:
        archiver: "true"
  fullBackup:
    driver: Restic
    scheduler:
      schedule: "*/3 * * * *"
      failedJobsHistoryLimit: 1
      successfulJobsHistoryLimit: 1
    sessionHistoryLimit: 2
    containerRuntimeSettings:
      securityContext:
        runAsUser: 0
        runAsGroup: 0
        runAsNonRoot: false
        allowPrivilegeEscalation: true
        privileged: true
  logBackup:
    logRetentionHistoryLimit: 5
    retentionPeriod: 10d
    retentionSchedule: "*/1 * * * *"
  manifestBackup:
    scheduler:
      schedule: "*/3 * * * *"
      failedJobsHistoryLimit: 1
      successfulJobsHistoryLimit: 1
    sessionHistoryLimit: 2
  encryptionSecret:
    name: encrypt-secret
    namespace: aws-demo11
  retentionPolicy:
    name: mysql-retention-policy
    namespace: aws-demo11


```


---


## Oracle

This release introduces native monitoring for Oracle with Prometheus and Grafana.

Metrics are collected using the free, public **Oracle AI Database Metrics Exporter**, which gathers standard Oracle database metrics and supports custom metrics collection for specific needs. The metrics are visualized in Grafana through flexible dashboards, enabling users to monitor database health and performance.

### Monitoring Configuration:
Include monitor configuration while deploying the Oracle database to enable the monitoring feature.

Example: Enable monitoring for Oracle standalone

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle
  namespace: exporter
spec:
  configuration:
    secretName: test-config
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9161
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```


---

## Neo4j

This release introduces three major capabilities for Neo4j managed by KubeDB on Kubernetes: native monitoring support via Prometheus, TLS/SSL encryption for secure communication, and OpsRequest lifecycle management including Restart and ReconfigureTLS operations.

### Feature 1: Neo4j Monitoring Support
KubeDB now supports native monitoring for Neo4j deployments using Prometheus and Grafana. Metrics are exposed directly from Neo4j and seamlessly integrated with the KubeDB monitoring stack.
Yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  version: "2025.11.2"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

### Feature 2: Neo4j TLS Support
KubeDB now supports TLS configuration for Neo4j deployments in Kubernetes, enabling encrypted communication for client and intra-cluster traffic. TLS can be configured for Bolt, HTTPS, and cluster communication by defining the appropriate SSL policies as described in the Neo4j configuration. Certificates can be provisioned using cert-manager and are mounted into the Neo4j Pods according to the configured SSL policy.

Yaml
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-tls
  namespace: demo
spec:
  version: 2025.11.2
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: neo4j-ca-issuer
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

### Feature 3: OpsRequest Support

KubeDB introduces OpsRequest support for Neo4j, enabling controlled, operator-managed lifecycle operations. In this release, two OpsRequest types are supported: Restart and ReconfigureTLS.

#### OpsRequest: ReconfigureTLS

The ReconfigureTLS OpsRequest allows administrators to add, update, rotate, or remove TLS configuration on a running Neo4j instance without downtime-inducing manual intervention. This operation is fully managed by the KubeDB operator.

**Add TLS**
Enables TLS on a Neo4j instance that was previously running without encryption.

**Update TLS**
Replaces the existing TLS issuer or certificates with new ones.

**Rotate Certificates**
Forces immediate re-issuance of all TLS certificates from the existing issuer.

**Remove TLS**
Disables TLS and reverts Neo4j to plain-text connections.

Here is sample yaml to **rotate-certificates**
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: neo4j-tls
  tls:
    rotateCertificates: true
    bolt:
      mode: mTLS
```


**OpsRequest: Restart**
The Restart OpsRequest triggers a controlled rolling restart of all Neo4j pods. This is useful after manual configuration changes, certificate updates, or when recovering from transient pod failures.

Sample yaml
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: neo4j-tls
  timeout: 5m
  apply: Always
```



---

## Qdrant

This release introduces native monitoring support for Qdrant using Prometheus and Grafana.

It leverages Qdrant’s built-in metrics endpoint to collect detailed performance and operational data. These metrics are integrated with Prometheus and visualized through customizable Grafana dashboards, allowing users to effectively monitor database health, performance, and resource utilization.

To enable monitoring, include the monitoring configuration during database deployment.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
spec:
  version: 1.16.2
  mode: Distributed
  replicas: 3
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 6333
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi

  deletionPolicy: WipeOut
```


---

## Milvus

We are excited to announce that KubeDB now supports Milvus in **Distributed Mode**, enabling scalable, production-ready Milvus deployments on Kubernetes. Two additional versions (2.6.9, 2.6.11) have also been added.

Distributed Milvus Deployment

The following manifest provisions a distributed Milvus instance.

```yaml

apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: kubedb
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
      datanode:
        replicas: 2
      proxy:
        replicas: 2
      querynode:
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: milvus
                resources:
                  requests:
                    cpu: "156m"
                    memory: "512Mi"
                  limits:
                    memory: "1Gi"
      streamingnode:
        replicas: 3
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 10Gi

```


---

### SAP HanaDB 

KubeDB supports **clustered HanaDB deployments** using **system replication** for high availability.
When clustering is enabled default modes are:
- **Replication Mode (default)**: `sync`
- **Operation Mode (default)**: `logreplay`

**Clustered Deployment (System Replication)**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hana-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 3
  storageType: Durable
  topology:
    mode: SystemReplication
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
```

This creates a 3-node HanaDB cluster with:

- System Replication enabled
- Default sync replication
- Default logreplay operation mode
- Persistent storage

**Custom Replication Settings**

You can override the defaults:

```yaml
topology:
  mode: SystemReplication
  systemReplication:
    replicationMode: fullsync
    operationMode: logreplay_readaccess
```

**Available Replication Modes**

- sync (default)
- syncmem
- async
- fullsync


**Available Operation Modes**
- logreplay (default)
- logreplay_readaccess
- Delta_datashipping

**Custom Database Configuration**
Inline configuration is supported:
```yaml
configuration:
  inline:
    global.ini: |-
      [memorymanager]
      global_allocation_limit = 8589934592
```
This example sets a custom global memory allocation limit.

---

## New Versions

- PostgreSQL: 18.2, 17.8, 16.12, 15.16, 14.21
- Milvus: 2.6.7, 2.6.9, 2.6.11
- Kafka: 4.2.0
- Qdrant: 1.17.0

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
---
