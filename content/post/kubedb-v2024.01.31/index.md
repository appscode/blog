---
title: Announcing KubeDB v2024.01.31
date: "2024-02-06"
weight: 14
authors:
- Raihan Khan
tags:
- cloud-native
- dashboard
- database
- druid
- elasticsearch
- ferretdb
- gcs
- git-sync
- kafka
- kafka-connect
- kibana
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- pgpool
- postgres
- postgresql
- prometheus
- rabbitmq
- redis
- s3
- security
- singlestore
- solr
- walg
---

We are pleased to announce the release of [KubeDB v2024.01.31](https://kubedb.com/docs/v2024.01.31/setup/). This release is primarily focused on extending support for new databases including `Solr`, `Singlestore`, `Druid`, `RabbitMQ`, `FerretDB` and `ZooKeeper`. Besides provisioning, Customizable health checker, Custom template for services, Custom template for pods & containers, Authentication, Multiple Termination strategies, Default Container Security includes this new database supports. This release also brings support for `ConnectCluster` for Kafka and `PgPool` for Postgres. A mojor API change for `ElasticsearchDashboard` has been added to this release as well. This post lists all the major changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.01.31/README.md). Let’s see the changes done in this release.


# Kafka:
In this release, we have improved Kafka Health Checker. The health checker now ensures all kafka brokers are connected in a cluster, messages can be concurrently published and consumed via clients.
Earlier, KubeDB managed Kafka had only two types of termination policy supports - `WipeOut` and `DoNotTerminate`. This release is bringing support for two more - `Halt` and `Delete`, providing the privilege of keeping PVCs and Secrets on cluster deletion.

## Kafka Connect Cluster

This release introduces `ConnectCluster` for **Kafka**, an awesome tool for building reliable and scalable data pipelines. Its pluggable design makes it possible to build powerful pipelines without writing a single line of code. The initial release of Kafka Connect Cluster is bringing support for Provisioning, Custom-Configuration via kubernetes secret and TLS. You can now enable source/sink connector plugins for `GCS`, `S3`, `MongoDB`, `Postgres` and `MySQL`. This release also introduces support for `Connector` CRD. This enables to easily connect source/sink clients to the `ConnectCluster`. 

Let's assume you have a Kafka cluster `kafka-prod`, provisioned using KubeDB is deployed in a namespace called `demo`. You can now provision a `ConnectCluster` using the following yaml.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connect-cluster
  namespace: demo
spec:
  connectorPlugins:
  - gcs-0.13.0
  - mongodb-1.11.0
  - mysql-2.4.2.final
  - postgres-2.4.2.final
  - s3-2.15.0
  kafkaRef:
    name: kafka-prod
    namespace: demo
  replicas: 2
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: connect-ca-issuer
  terminationPolicy: WipeOut
  version: 3.6.0
```

Once ConnectCluster is ready, you can connect any of supported connector via a `Connector` custom resource. In order to Connect a MongoDB database to the `ConnectCluster` apply the following yaml. Provide a kubernetes secret reference in `spec.configSecret` field which should be containing all the necessary connection information to the sink/source client.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: Connector
metadata:
  name: mongo-source
  namespace: demo
spec:
  configSecret:
    name: mongo-source
  connectClusterRef:
    name: connect-cluster
    namespace: demo
  terminationPolicy: WipeOut
```

#### supported versions
`3.3.2`, `3.4.1`, `3.5.1`, `3.6.0`

# Zookeeper

With This release, KubeDB is bringing support for **Apache ZooKeeper**, a pivotal open-source distributed coordination service for managing and synchronizing distributed systems. It offers essential primitives like configuration management, distributed locking, leader election, and naming services. Utilizing a quorum-based approach, ZooKeeper ensures high availability and fault tolerance. Widely applied in systems like Apache Hadoop, Solr, Druid and Kafka, ZooKeeper streamlines development by providing a consistent foundation for coordination tasks. This release of Zookeeper is bringing support for Provisioning and authentication via kubernetes secret.

Apply the following YAML to deploy a ZooKeeper cluster with 3 replicas. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-cluster
  namespace: demo
spec:
  version: "3.9.1"
  replicas: 3
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: "Halt"
```

#### supported versions
`3.7.2`, `3.8.3` and `3.9.1`

# Solr

This release includes support for **Solr**, an open-source enterprise-search platform, written in Java. Its major features include full-text search, hit highlighting, faceted search, real-time indexing, dynamic clustering, database integration, NoSQL features and rich document handling. This release of Solr is coming with support for Provisioning SolrCloud cluster, Embedded Admin UI, Custom Configuration via kubernetes secret. You can deploy Solr either in combined mode or the production preferred topology mode.

Here's a sample yaml for provisioning Solr using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  terminationPolicy: Delete
  zookeeperRef:
    name: zk-cluster
  topology:
    Overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 3Gi
        storageClassName: standard
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 3Gi
        storageClassName: standard
    Coordinator:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 3Gi
        storageClassName: standard
```

#### supported versions
`9.4.1` and `8.11.2`

# Singlestore:
This release introduces support for **SingleStore**, a distributed SQL database for real-time analytics, transactional workloads, and operational applications. With its in-memory processing and scalable architecture, SingleStore enables organizations to achieve high-performance and low-latency data processing across diverse data sets, making it ideal for modern data-intensive applications and analytical workflows. This release is bring support for Provisioning Singlestore in both standalone and cluster mode, Custom Configuration using kubernetes secret and Initialization using script.

You can provision Singlestore cluster in production ideal cluster mode using the following YAML.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: singlestore-sample
  namespace: demo
spec:
  version: "8.1.32"
  topology:
    aggregator:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  terminationPolicy: WipeOut
```

#### supported versions
`3.1.32`

# Pgpool

This release also introduces support for **Pgpool**, an advanced connection pooling and load balancing solution for PostgreSQL databases. It serves as an intermediary between client applications and PostgreSQL database servers, providing features such as connection pooling, load balancing, query caching, and high availability. This release includes support for Provisioning, Custom Configuration using kubernetes secret and Postgres users synchronization.

Deploy Pgpool with a postgres backend using the following YAML.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-test
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 5
  backend:
    name: ha-postgres
  terminationPolicy: WipeOut
  syncUsers: true
  initConfig:
    pgpoolConfig:
      log_statement : on
      log_per_node_statement : on
      sr_check_period : 0
      health_check_period : 0
      backend_clustering_mode : 'streaming_replication'
      num_init_children : 5
      max_pool : 150
      child_life_time : 300
      child_max_connections : 0
      connection_life_time : 0
      client_idle_limit : 0
      connection_cache : on
      load_balance_mode : on
      ssl : on
      failover_on_backend_error : off
      log_min_messages : warning
      statement_level_load_balance: on
      memory_cache_enabled: on
```

#### supported versions
`4.4.5` and `4.5.0`

# Druid

This release is introducing support for **Apache Druid**, a real-time analytics database designed for fast slice-and-dice analytics ("OLAP" queries) on large data sets. Most often, Druid powers use cases where real-time ingestion, fast query performance, and high uptime are important. This release includes support for Provisioning Druid cluster, Custom Configuration using kubernetes secret, Management UI, External Dependency Management: Configuring External Dependencies i.e. MetadataStorage (MySQL/PostgreSQL), Deep Storage, ZooKeeper directly from YAML.

Here's a sample YAML to provision Druid cluster which uses mysql as metadata storage, s3 as deepstorage and zookeeper for coordination. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-sample
  namespace: druid
spec:
  version: 28.0.1
  storageType: Ephemeral
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  metadataStorage:
    name: mysql-cluster
    namespace: druid
    createTables: true
  zooKeeper:
    name: zk-cluster
    namespace: druid
  topology:
    coordinators:
      replicas: 1
    overlords:
      replicas: 1
    brokers:
      replicas: 1
    historicals:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    middleManagers:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    routers:
      replicas: 1
```

#### supported versions
`28.0.1` and `25.0.0`

# FerretDB

This release extends to supporting **FerretDB**, an open-source proxy that translates MongoDB wire protocol queries to SQL, with PostgreSQL or SQLite as the database engine. FerretDB was founded to become the true open-source alternative to MongoDB, making it the go-to choice for most MongoDB users looking for an open-source alternative to MongoDB. With FerretDB, users can run the same MongoDB protocol queries without needing to learn a new language or command.

Currently KubeDB only supports Postgres backend as database engine for FerretDB. Users can use its own Postgres or let KubeDB create and manage backend engine with KubeDB native Postgres. This release includes support for provisioning and monitoring with prometheus and grafana dashboards.

Here's a sample manifest to provision FerretDB with User provided backend

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb-sample
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  authSecret:
    externallyManaged: true
    name: ha-postgres-auth
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 600Mi  
  backend:
    externallyManaged: true
    postgres:
      url: postgres://ha-postgres.demo.svc.cluster.local:5432/
  terminationPolicy: WipeOut
```

Here's another sample manifest using KubeDB provided and managed backend

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferredb-sample-2
  namespace: demo
spec:
  version: "1.18.0"
  terminationPolicy: WipeOut
  replicas: 1
  sslMode: disabled  
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  backend:
    externallyManaged: false
  authSecret:
    externallyManaged: false
```

#### supported versions
`1.18.0`

# RabbitMQ

This release also brings support for provisioning **RabbitMQ**, a popular message queue tool that connects applications to exchange messages, facilitating scalability and stability for a competitive edge. Support for provisioning, Custom configuration via kubernetes secrets and preinstalled management UI has been added in this release.

Here's a sample manifest to provision RabbitMQ.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbit-dev
  namespace: demo
spec:
  version: "3.12.12"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
```

#### supported versions
`3.12.12`

# Elasticsearch

ElasticsearchDashboard, a custom resource is used to provision Kibana for Elasticsearch cluster and Opensearch-Dashboards for Opensearch cluster has received a major change. The `apiVersion` for this CR has been updated to  `elasticsearch.kubedb.com/v1alpha1`. Here's a sample manifest to provision ElasticsearchDashboard.

```yaml
apiVersion: elasticsearch.kubedb.com/v1alpha1
kind: ElasticsearchDashboard
metadata:
  name: es-cluster-dashboard
  namespace: demo
spec:
  enableSSL: true
  databaseRef:
    name: es
  terminationPolicy: WipeOut
```

# Postgres

### Grafana Alerts Dashboard
In this release, we bring support for Postgres grafana summary dashboard with alerts for a specific Postgres instance.

In order to use the alert dashboards, You need to do the following:

* First you need to have kubedb operator and a postgres instance running on you cluster along with prometheus operator.

* Then create a file named overwrite.yaml having following values. Make sure you change the values matching your needs.

    - For prometheus operator
        ```yaml
        resources:
          - postgres

        dashboard:
          folderID: 0
          overwrite: true
          alerts: true 
          replacements: []

        grafana:
          version: “provide your grafana version”
          url: "provide your grafana service dns" # example: http://grafana.monitoring.svc:80 
          apikey: "provide grafana apikey"

        app:
          name: “provide your postgres instance name”
          namespace: “provide the namespace”
        ```
    - For prometheus builtin
       ```yaml
        resources:
          - postgres

        dashboard:
          folderID: 0
          overwrite: true
          alerts: true 
        replacements:
          job=\"kube-state-metrics\": job=\"kubernetes-service-endpoints\"
          job=\"kubelet\": job=\"kubernetes-nodes-cadvisor\"
          job=\"$app-stats\": job=\"kubedb-databases\"

        grafana:
          version: “provide your grafana version”
          url: "provide your grafana service dns" # example: http://grafana.monitoring.svc:80 
          apikey: "provide grafana apikey"

        app:
          name: “provide your postgres instance name”
          namespace: “provide the namespace”
        ```    

* now follow the steps to install grafana dashboards:
    ```bash
    $ helm repo add appscode https://charts.appscode.com/stable/
    $ helm repo update
    $ helm search repo appscode/kubedb-grafana-dashboards --version=v2024.1.31
    $ helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards -n kubeops --create-namespace --version=v2024.1.31 
    -f  /path/to/the/overwrite.yaml
    ```

### Bug fix and improvements
* Upgrade ops request bugfix has been implemented. Now you can upgrade your postgres instance using postgres version upgrade ops request.
    ```yaml
    apiVersion: ops.kubedb.com/v1alpha1
    kind: PostgresOpsRequest
    metadata:
    name: update-version
    namespace: demo
    spec:
    type: UpdateVersion
    updateVersion:
        targetVersion: "16.1"
    databaseRef: 
        name: ha-postgres
    ```

# Point in Time Recovery:


We have updated the backup directory structure for point-in-time recovery, providing a more organized layout. Going forward, the directories will follow the following standardized structure:

- **Full Backup:**
<SUB-DIRECTORY>/<DATABASE_NAMESPACE>/<DATABASE_NAME>/full-backup
- **Manifest Backup:**
<SUB-DIRECTORY>/<DATABASE_NAMESPACE>/<DATABASE_NAME>/manifest-backup
- **Binlog Backup (MySQL/MariaDB):**
<SUB-DIRECTORY>/<DATABASE_NAMESPACE>/<DATABASE_NAME>/binlog-backup
- **Oplog Backup (MongoDB):**
<SUB-DIRECTORY>/<DATABASE_NAMESPACE>/<DATABASE_NAME>/oplog-backup
- **WAL Backup (PostgreSQL):**
<SUB-DIRECTORY>/<DATABASE_NAMESPACE>/<DATABASE_NAME>/wal-backup

We've introduced TLS support for MySQL point-in-time recovery. Additionally, we now support using NFS, GCS, and Azure Storage as backend options for MySQL point-in-time recovery.


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2024.01.31/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2024.01.31/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).