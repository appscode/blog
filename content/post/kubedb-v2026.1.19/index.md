---
title: Announcing KubeDB v2026.1.19
date: "2026-1-19"
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

KubeDB **v2026.1.19** is a major milestone focused on **unified configuration management**, **GitOps safety**, **operational reliability**, and **massive database ecosystem expansion**. This release introduces a redesigned configuration/reconfiguration framework across *all* databases, smarter OpsRequest execution, and first-class support for several new **vector, graph, and enterprise databases**.

Alongside core stability improvements, this release brings **PostgreSQL sharded OpsRequests**, **Postgresql auto-tuning with pgtune**, **spec-driven TLS for Oracle**, and **new database engines** like **Neo4j, Qdrant, Milvus, Weaviate, and SAP HanaDB**.

---

## Key Highlights

* **Unified Configuration & Reconfiguration API for all databases**
* **Merged & robust Reconfigure OpsRequest execution**
* **PostgreSQL OpsRequest sharding and auto-tuning**
* **New database support: Neo4j, Qdrant, Milvus, Weaviate, HanaDB**
* **Spec-driven TLS for Oracle**
* **Multiple new database versions across MariaDB, MySQL, Redis, MSSQL, ElasticSearch**

---

## Common Improvements

### Unified Configuration & Reconfiguration Framework

In this release, we have **redesigned and generalized the Configure / Reconfigure workflow** across all supported databases. The update makes configuration management consistent, GitOps-friendly, and safer for shared resources.

### New Unified Configuration API

The legacy `spec.configSecret` has been replaced with a **flexible configuration model**:


```yaml
spec:
  configuration:
    secretName: custom-config
    inline:
      key: value
    tuning:
    acl:
```

#### Why this matters

* Configuration sources are now **explicitly visible** in the database CR
* Inline configs can be applied during **provisioning and reconfiguration**
* Eliminates hidden operator mutations, simplifying GitOps workflows
* Old fields are **automatically migrated** and deprecated

---

### User Secrets Are Now Read-Only 

Operators no longer mutate user-provided custom config secrets.

**New behavior**

* User secrets are **read-only**
* Operator creates an internal secret:

  ```
  <db-name>-<db-cr-uid-last-6>
  ```
* Internal secret stores:

    * Generated configs
    * Inline configs
    * Tuning outputs

This dramatically improves **GitOps safety and auditability**.

---

### Generalized ReconfigureOpsRequest

The `ReconfigureOpsRequest` now supports **safe, composable configuration changes**.

```yaml
spec:
  configuration:
    removeCustomConfig: true
    configSecret: new-secret
    applyConfig:
      key: value
```

#### Behavior Summary

* `removeCustomConfig: true` → removes all previous custom config
* `configSecret` → applies a new secret-based config, replaces the previous config secret, but keeps the existing Inline
* `applyConfig` → apply inline configuration, merged with existing Inline configuration (if any)
---

> Note: If the same configuration exists in both Secret and Inline, Inline takes priority

### Restart-Aware Reconfiguration

```yaml
spec:
  configuration:
    restart: auto | true | false
```

* `auto` (default): restart only if required (determined by ops manager operator)
* `false`: no restart
* `true`: always restart

This significantly reduces **unnecessary downtime**.

---

## Ops Manager Improvements

### Merged Reconfigure OpsRequests

Multiple pending reconfiguration ops requests meant for a single database are now **automatically merged** into one.

**Benefits**

* Fewer reconciliations
* Fewer restarts
* Faster convergence

**Behavior**

* All pending Reconfigure OpsRequests for the same database are aggregated into one.
* Configurations from pending requests are merged one by one in creation-time order, and for any overlapping configs, the value from the latest request takes precedence.
* A new OpsRequest (e.g., <db-name>-rcfg-merged-<timestamp>) is created with the combined configuration.
* Original pending requests are marked Skipped with reason `ConfigurationMerged`. The merged request records its sources via the `MergedFromOps` condition.
* Ongoing (Progressing) OpsRequests are not affected.

---

### Robust OpsRequest Execution

This release improves the reliability and safety of OpsRequest execution by fixing several issues related to stale state, retries, and cancellation.

**Enhancements:**

* Added configurable timeouts via spec.timeout with a dynamic default (2 × pod count minutes).
* Implemented context-based graceful cancellation when OpsRequests are deleted.
* Ensured concurrency safety using per-operation locking and proper resource cleanup.

**Bug Fixes:**

* Fixed unnecessary database restarts caused by stale status conditions from the informer cache.
* Resolved OpsRequests stopping mid-execution due to missing requeue logic.
* Prevented infinite execution and memory leaks when OpsRequests are deleted during execution.
* Avoid duplicate create/delete attempts for already-completed resources.


---

## GitOps Operator Fix

Fixed an issue where **Postgres arbiter storage changes** incorrectly triggered a `VerticalScaling` OpsRequest.
These updates now correctly create a **VolumeExpansion OpsRequest**.

---

## PostgreSQL

### Sharded OpsRequest Support

We are pleased to announce support for PostgreSQL ops-request sharding. In our earlier [release](https://appscode.com/blog/post/operator-shard-manager-v2025.3.14/), we had introduced support for sharding our database resources. From this release, you will be able to create `PostgresOpsRequest` that will be handled by sharded `OpsManager` pods.
The process and benefits are similar that has been discussed in the above blog post. You just need to apply a different `ShardConfiguration` object with the given below:


```yaml
apiVersion: operator.k8s.appscode.com/v1alpha1
kind: ShardConfiguration
metadata:
  name: kubedb-ops-manager
spec:
  controllers:
  - apiGroup: apps
    kind: StatefulSet
    name: kubedb-kubedb-ops-manager # Statefulset for controlling lifecycle of ops requests
    namespace: kubedb
  resources:
  - apiGroup: kubedb.com
  - apiGroup: elasticsearch.kubedb.com
  - apiGroup: kafka.kubedb.com
  - apiGroup: postgres.kubedb.com
  - apiGroup: ops.kubedb.com
    shardKey: ".spec.databaseRef.name"
    useCooperativeShardMigration: true
```
Note, we have introduced a new field `.spec.resources[*].shardKey`, this key is used to assign your ops-request to a sharded pod of Ops-Manager.
If you do not provide any shardKey, then we will use the ops object name as ShardKey by default. We want each of our ops-request that is meant for a single database to go on a particular shard, that's what is being done here using `.spec.resources[*].shardKey`. As all ops-requests meant for a particular database will have the same `.spec.databaseRef.name`, so they will be scheduled on the same shard.

---

### Auto Configuration Tuning with pgtune

Postgres now supports **automatic tuning** powered by [pgtune](https://github.com/gregs1104/pgtune).

```yaml
spec:
  configuration:
    tuning:
      maxConnections: 200
      profile: web
      storageType: ssd
```

Some example of auto tuning parameters are given as sample:

```bash

/* Database specs */
// DB Version: 17
// OS Type: linux
// DB Type: web
// Total Memory (RAM): 4 GB
// CPUs num: 2
// Connections num: 98
// Data Storage: ssd

/* Tuned parameters */
max_connections = 98
shared_buffers = 1GB
effective_cache_size = 3GB
maintenance_work_mem = 256MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 9892kB
huge_pages = off
min_wal_size = 1GB
max_wal_size = 4GB
```

```bash
/* Database specs */

// DB Version: 17
// OS Type: linux
// DB Type: web
// Total Memory (RAM): 6 GB
// CPUs num: 4
// Connections num: 200
// Data Storage: ssd

/* Tuned parameters */

max_connections = 200
shared_buffers = 1536MB
effective_cache_size = 4608MB
maintenance_work_mem = 384MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 7710kB
huge_pages = off
min_wal_size = 1GB
max_wal_size = 4GB
max_worker_processes = 4
max_parallel_workers_per_gather = 2
max_parallel_workers = 4
max_parallel_maintenance_workers = 2
```

See the Api Enhancements section about how to set up auto-tuning.


### Api Enhancements

Previously we used the  `.spec.configSecret` field to provide custom configuration to our postgres databases. Now we have introduced a new field `.spec.configuration`. Using this field, now you can provide configuration in 3 different ways.


#### Using configSecret

```yaml

spec:
  configuration:
    secretName: pg-conf  // previously this secret was given vai .spec.configSecret field
```

#### Using inline configuration
```yaml
spec:
  configuration:
    inline:
      user.conf: |
        max_connections=135     // you can provide multiple configurations like this
        shared_buffers=256MB
```
**Note** In this way, `user.conf` key is fixed, you can’t use any other key life abc.conf, xyz.apply etc.

#### Auto tuning

```yaml
spec:
  configuration:
    tuning:
      maxConnections: 200
      profile: web
      storageType: ssd

```

If you want auto-tuning of your postgresql resource, then you can use this specification. Just make sure to change your `maxConnections`, `profile` and storageType.

For now, available options for profile is:

```bash
web // web optimizes for web applications with many simple queries
oltp // oltp optimizes for OLTP workloads with many short transactions
dw // dw optimizes for data warehousing with complex analytical queries
mixed // mixed optimizes for mixed workloads
desktop // desktop optimizes for desktop or development environments
```
For storageType supported values are `ssd`, `hdd`, and `san`.  

**We will extract `CPU`, `Memory`, and `Version` from your database Custom Resource.**

> **Note**: If you use all three options in `.spec.configuration` then `.spec.configuration.inline` > `.spce.configuration.secretName` > `.spec.configuration.tuning` will be the priority order of the parameters. For example, if you set max_connections parameter in all three of `.spec.configuration`, then the value set in `.spec.configuration.inline` will take precedence.
If you run `show max_connections;`, you will see the value that was set in `.spec.configuration.inline`.

---

### Force Failover Support

```yaml
spec:
  replication:
    forceFailoverAcceptingDataLossAfter: 5m
```

Users can now set `.spec.replication.forceFailoverAcceptingDataLossAfter`. ForceFailoverAcceptingDataLossAfter is the maximum time to wait before running a force failover process. This is helpful for a scenario where the old primary is not available and it has the most updated wal lsn. Doing force failover may or may not end up losing data depending on any write transaction in the range lagged lsn between the new primary and the old primary.

---

### Bug fix
* We introduced grpc server from kubedb v2025.7.31, so upgrading from a later version was causing an issue. After upgrading, if you restart any one of your postgres pods, then postgres would go in critical state because the other pods do not have grpc server. We fixed that issue in this release.
* Fix an issue with adding tls in the database.
* Fix some cases where pg_rewind was necessary, but our coordinator was skipping pg_rewind.

---
### Improvements

* Reduced raft load using peer-to-peer gRPC
* Smart basebackup with disk-space awareness
* Automatic rollback on backup failure
* Intelligent backup strategy
    - Uses `/var/pv/data.bc` if >50% space available
    - Falls back to `/tmp/var/pv/data` for low disk space scenarios
* Continuous mount health checks

Split-Brain Protection has been added in this release.
Continuous mount check has been added in case of node failure. A evict will be performed in case of mount failure by kubelet or any chaos related scenarios.


---

## Redis

In this release we have updated the API of redis where we moved the ACL Spec from `spec.acl` to `spec.configuration.acl`. As an example you can have a look in this yaml:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-instance
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  deletionPolicy: WipeOut
  configuration:
    acl:
      secretRef:
        name: old-acl-secret         # Secret that holds passwords referenced by variables like ${k1}
      rules:
        - userName1 ${k1} allkeys +@string +@set -SADD
        - userName2 ${k2} allkeys +@string +@set -SADD
        - userName3 ${k3} allkeys +@string +@set -SADD
        - userName4 ${k4} allkeys +@string +@set -SADD
```


---

## ProxySQL

The configuration API has been updated to improve clarity and maintainability.

### Previous API

```yaml
spec:
  initConfig:
    mysqlUsers:
    mysqlQueryRules:
    mysqlVariables:
    adminVariables:
  configSecret:
    name: proxysql-custom-config
```

### New API

```yaml
 configuration:
    init:
      inline: 
        mysqlUsers:
        mysqlQueryRules:
        mysqlVariables:
        adminVariables:
      secretName: proxysql-custom-config
```

### Key Points

* If the same configuration (e.g., MySQLUsers) exists in both inline and secret, the inline configuration overrides the secret configuration (inline > secret).
* Old fields are deprecated; if set, the operator automatically copies their values to the new API fields.


## MariaDB

In this release, MariaDB has resolved the physical base backup version incompatibility issue that previously affected point-in-time recovery (PITR) backup and restore operations.

---
## Pgpool

In this release we have removed `spec.configSecret` & `spec.initConfig` from the pgpool api.
Instead you can use `spec.configuration`. See the common changes for more details.

---

## Oracle

### Spec-Driven TLS Enablement

TLS is now **fully spec-driven** via `tcpsConfig`.

```yaml
spec:
  tcpsConfig:
    tls:
      issuerRef:
        name: oracle-ca-issuer
```

* No manual wallet setup
* Cert-manager integration
* Only Server-side TLS (TCPS on port 2484)

---

## New Database Engines

### Neo4j (NEW)

We’re excited to introduce support for Neo4j, the world’s leading graph database management system designed to harness the power of connected data. Neo4j offers native graph storage, full ACID compliance, and the expressive Cypher query language, making it ideal for knowledge graphs, fraud detection, and real-time recommendation engines. Key features include:

* Cluster Mode Provisioning: Deploy Neo4j Autonomous Clusters with ease using KubeDB.
* Custom Configuration: Support for custom configurations via Kubernetes secrets to fine-tune your graph engine.
* Authentication: Enhanced security with built-in authentication.
* 
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j
  namespace: demo
spec:
  replicas: 3
  podTemplate:
    spec:
      securityContext:
        fsGroup: 7474
        fsGroupChangePolicy: Always
        runAsGroup: 7474
        runAsNonRoot: true
        runAsUser: 7474
      terminationGracePeriodSeconds: 3600
      containers:
        - name: neo4j
          resources:
            limits:
              cpu: 500m
              memory: 2Gi
            requests:
              cpu: 500m
              memory: 2Gi
  version: "2025.10.1"
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```
Supported Version: 2025.10.1

---

### Qdrant (NEW)

This release brings support for provisioning Qdrant.
Qdrant is a powerful, open-source vector database designed for high-performance similarity search and AI-driven applications. Built to handle embeddings at scale, Qdrant enables fast and accurate nearest-neighbor search, making it an ideal choice for use cases like semantic search, recommendation systems, RAG pipelines, and anomaly detection. With its efficient indexing, filtering capabilities, and real-time updates, Qdrant delivers low-latency search even across millions of vectors.

Supports include:

* Distributed mode
* TLS support
* Custom configuration

Here’s a sample manifest to provision Qdrant.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: 1.16.2
  replicas: 2
  mode: Distributed
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

---

#### Custom Configuration for Qdrant
Store your custom config.yaml in a Secret and reference it from the Qdrant manifest. Alternatively, you can define the configuration directly in the manifest using inline:
```yaml
spec:
  configuration:
    secretName: custom-config
    inline:
      config.yaml: |
        log_level: INFO
 ```

---
#### Enabling TLS for Qdrant

To secure communication between Qdrant replicas and external clients, you can enable TLS by configuring the TLS section in the Qdrant manifest. KubeDB integrates with Kubernetes Secrets to manage certificates and keys securely, ensuring encrypted traffic for both client and inter-node communication.
```yaml
spec:
  tls:
    issuerRef:
      name: qdrant-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    p2p: true
    client: true
```

Supported versions: **1.16.2, 1.15.4**

---

### Milvus (NEW)

We are excited to announce that KubeDB now supports Milvus, an open-source vector database optimized for similarity search and AI applications. With KubeDB integration, users can deploy Milvus on Kubernetes seamlessly, using persistent storage, object storage, authentication, and configurable pod templates for production-ready deployments.

Key features include:

* Standalone mode
* External object storage (MinIO)
* Authentication & custom config

The following manifest provisions a standalone Milvus instance
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-standalone
  namespace: milvus-standalone
spec:
  version: "2.6.7"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: "minio-secret"
  storage:
    storageClassName: standard
```

---

#### Object storage integration

Supports externally managed object storage to store vector data securely and efficiently.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-secret
  namespace: milvus-standalone
type: Opaque
stringData:
  address: "milvus-minio:9000"
  accessKeyId: "minioadmin"
  secretAccessKey: "minioadmin"
```

#### Secure authentication with optional auth secret

Enables authentication using either internally managed secrets or externally managed secrets.
```yaml
spec:
  disableSecurity: false
  authSecret:
    Name: milvus-auth
```
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-auth
  namespace: milvus-standalone
type: kubernetes.io/basic-auth
stringData:
  username: "root"
  password: "Milvus"
```

#### Custom Configuration

Users can provide a custom Milvus configuration either via a Secret or inline in the manifest. This keeps the configuration flexible and easy to manage.

```yaml
spec:
  configuration:
    secretName: milvus-standalone-custom-config   
    inline:
      milvus.yaml: |                              
        queryNode:
          gracefulTime: 20
```


Supported version: **2.6.7**

---

### Weaviate (NEW)

KubeDB now supports Weaviate, an open-source vector database for storing and querying data using machine-learning-powered vector embeddings. Weaviate enables efficient vector similarity and hybrid (vector + keyword) search, making it ideal for AI, NLP, recommendation systems, and knowledge graph applications.
With KubeDB, users can seamlessly deploy Weaviate on Kubernetes with persistent storage, authentication, and configurable pod templates for production-ready workloads.

Features include:

* Persistent storage
* Custom config
* Health checks

```yaml

apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  configuration:
      secretName: weaviate-user-config
      inline:
          conf.yaml: |- 
              query_defaults:
	       limit: 1000
  podTemplate:
    spec:
      containers:
        - name: weaviate
          securityContext:
            runAsNonRoot: false
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              memory: 1Gi

  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
```


Supported version: **1.33.1**

---

### SAP HanaDB (NEW)

KubeDB now supports SAP HANA Database (HanaDB), an enterprise-grade in-memory database platform designed for real-time analytics and transactional processing. HanaDB combines row-based, column-based, and object-based database technologies to deliver exceptional performance for complex queries and high-volume data processing. With KubeDB integration, users can deploy HanaDB on Kubernetes with persistent storage, authentication, and health monitoring for production workloads.

Supported features:

* Standalone deployment
* Durable storage
* Authentication

The following manifest provisions a standalone HanaDB instance with durable storage, authentication, and health checks:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hana-standalone
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: standard
```

HanaDB requires authentication credentials for the SYSTEM user. You can either provide an externally managed secret or let KubeDB generate one automatically. Here's an example of an externally managed authentication secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hana-cluster-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: SYSTEM
  password: HanaCluster123!
```
When using an externally managed secret, set authSecret.externallyManaged: true in the HanaDB spec. For automatic credential generation, set it to false and KubeDB will create and manage the authentication secret for you.


Supported version: **2.0.82**

---

## Version Updates

* **MariaDB**: 11.8.5, 12.1.2
* **MySQL**: 9.4.0
* **Microsoft SQL Server**: 2025-RTM-ubuntu-22.04, 2022-CU22-ubuntu-22.04
* **ElasticSearch**:

    * xpack-9.2.3
    * xpack-9.1.9
    * xpack-9.0.8
    * xpack-8.19.9
    * xpack-8.18.8
    * xpack-8.17.10

---

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
---
