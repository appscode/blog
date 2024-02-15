---
title: Announcing KubeDB v2024.2.14
date: "2024-02-15"
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

We are pleased to announce the release of [KubeDB v2024.2.14](https://kubedb.com/docs/v2024.2.14/setup/). This release is primarily focused on extending support for new databases including `Solr`, `Singlestore`, `Druid`, `RabbitMQ`, `FerretDB` and `ZooKeeper`. Besides provisioning, customizable health checker, custom template for services, custom template for pods & containers, authentication, multiple termination strategies, default container security are included along with these new database supports. This release also brings support for `ConnectCluster` for Kafka and `Pgpool` for Postgres. A breaking update in API for `ElasticsearchDashboard` has been added to this release as well. `Grafana summary dashboard with alerts` for Postgres and `Point in Time Recovery` for archiver supported databases are further additions in this release. All the KubeDB managed database images has been custom-built ensuring less CVEs. This post lists all the major changes done in this release since the last release.  Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.2.14/README.md) and [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.1.31/README.md). Let’s see the changes done in this release.

## Introducing FeatureGate for KubeDB installation

With this release, KubeDB installation is getting enhanced with FeatureGate support. Now you can manage the CRDs while installing KubeDB via helm chart. Earlier, all the supported CRDs were installed. With FeatureGate, now it's possible to enable or disable certain database CRDs while installing KubeDB.

Let's say you want to provision Druid, use MySQL or PostgreSQL as metadata storage and ZooKeeper for Coordination. You want to install just those required CRDs on your cluster. Install KubeDB using the following command:

```bash
    helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
        --version v2024.2.14 \
        --namespace kubedb --create-namespace \
        --set-file global.license=/path/to/the/license.txt \
        --set kubedb-provisioner.operator.securityContext.seccompProfile.type=RuntimeDefault \
        --set kubedb-webhook-server.server.securityContext.seccompProfile.type=RuntimeDefault \
        --set kubedb-ops-manager.operator.securityContext.seccompProfile.type=RuntimeDefault \
        --set kubedb-autoscaler.operator.securityContext.seccompProfile.type=RuntimeDefault \
        --set global.featureGates.Druid=true --set global.featureGates.ZooKeeper=true \
        --set global.featureGates.MySQL=true -set global.featureGates.PostgreSQL=true 
```

## New Database Supports

We are introducing new database supports for KubeDB. Now, you can provision databases like `Solr`, `Singlestore`, `Druid`, `RabbitMQ`, `FerretDB` and `ZooKeeper`. `Pgpool` for Postgres and `ConnectCluster` for Kafka are also two new additions coming out with this release. Check out the features coming out in this release for these newly added databases.

## Druid

This release is introducing support for **Apache Druid**, a real-time analytics database designed for fast slice-and-dice analytics ("OLAP" queries) on large data sets. Most often, Druid powers use cases where real-time ingestion, fast query performance, and high uptime are important. This release includes support for Provisioning Druid cluster, Custom Configuration using kubernetes secret, Management UI, External Dependency Management i.e. Configuring Metadata Storage (MySQL/PostgreSQL), Deep Storage, ZooKeeper directly from YAML.

Here's a sample YAML to provision Druid cluster which uses mysql as metadata storage, S3 as Deepstorage and zookeeper for coordination. A ZooKeeper instance `zk-cluster` needs and MySQL instance `mysql-cluster` need to be provisioned prior to deploying the following manifest.

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
`25.0.0` and `28.0.1`

## FerretDB

This release extends to supporting **FerretDB**, an open-source proxy that translates MongoDB wire protocol queries to SQL, with PostgreSQL or SQLite as the database engine. FerretDB was founded to become the true open-source alternative to MongoDB, making it the go-to choice for most MongoDB users looking for an open-source alternative to MongoDB. With FerretDB, users can run the same MongoDB protocol queries without needing to learn a new language or command.

Currently, KubeDB supports Postgres backend as database engine for FerretDB. Users can use its own Postgres or let KubeDB create and manage backend engine with KubeDB native Postgres. This release includes support for provisioning and monitoring with prometheus and grafana dashboards.

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
  terminationPolicy: WipeOut
```

#### supported versions
`1.18.0`

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

## Pgpool

This release also introduces support for **Pgpool**, an advanced connection pooling and load balancing solution for PostgreSQL databases. It serves as an intermediary between client applications and PostgreSQL database servers, providing features such as connection pooling, load balancing, query caching, and high availability. This release includes support for Provisioning, Custom configuration provided in `.spec.initConfig.pgpoolConfig` field and Postgres users synchronization.

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

## RabbitMQ

This release also brings support for provisioning **RabbitMQ**, a popular message queue tool that connects applications to exchange messages, facilitating scalability and stability for a competitive edge. Support for provisioning, Custom configuration via kubernetes secrets and preinstalled management UI has been added in this release. The RabbitMQ Management UI is a user-friendly interface that let you monitor and handle your RabbitMQ server from a web browser. Among other things queues, connections, channels, exchanges, users and user permissions can be handled - created, deleted and listed in the browser. You can monitor message rates and send/receive messages manually. Monitoring RabbitMQ metrics via Grafana Dashboards has also been added in this release. You can use our prebuilt Grafana Dashboards from [here](https://github.com/appscode/grafana-dashboards/tree/master/rabbitmq).

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
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  storageType: Durable
  terminationPolicy: WipeOut
```

#### supported versions
`3.12.12`

## Singlestore
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

## Solr

This release includes support for **Solr**, an open-source enterprise-search platform, written in Java. Its major features include full-text search, hit highlighting, faceted search, real-time indexing, dynamic clustering, database integration, NoSQL features and rich document handling. This release of Solr is coming with support for Provisioning SolrCloud cluster, Embedded Admin UI, Custom Configuration via kubernetes secret. You can deploy Solr either in combined mode or the production preferred topology mode.

Here's a sample yaml for provisioning Solr using KubeDB. A ZooKeeper instance `zk-cluster` should be provisioned in the same namespace before deploying.

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
`8.11.2` and `9.4.1`

## Zookeeper

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
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: "Halt"
```

#### supported versions
`3.7.2`, `3.8.3` and `3.9.1`

## Features and Improvements

This release also contains some fixes, improvements and new features for existing database supports. Following are those in details.

## Elasticsearch

ElasticsearchDashboard, a custom resource used to provision Kibana for Elasticsearch cluster and Opensearch-Dashboards for Opensearch cluster has received a major change. Through this release we are bringing a breaking change by updating the `apiVersion` for this CR which has been updated to  `elasticsearch.kubedb.com/v1alpha1`.

Here's a sample manifest to provision ElasticsearchDashboard. An Elasticsearch instance `es-cluster` is expected to be provisioned prior to deploying the following manifest.

```yaml
apiVersion: elasticsearch.kubedb.com/v1alpha1
kind: ElasticsearchDashboard
metadata:
  name: es-cluster-dashboard
  namespace: demo
spec:
  enableSSL: true
  databaseRef:
    name: es-cluster
  terminationPolicy: WipeOut
```

## Kafka

In this release, we have improved Kafka Health Checker. The health checker now ensures all kafka brokers are connected in a cluster, messages can be concurrently published and consumed via clients.
Earlier, KubeDB managed Kafka had only two types of termination policy supports - `WipeOut` and `DoNotTerminate`. This release is bringing support for two more - `Halt` and `Delete`, providing the privilege of keeping PVCs and Secrets on cluster deletion.

## Redis
We have added support for initialization script in Redis. You can provide a `bash` or `lua` script through configmap and The script will be run at the start.

Here is a sample configmap which has a `bash` script that creates two users in Redis nodes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-init-script
  namespace: demo
data:
  script.sh: |

    user1=$(cat auth/1/username)
    pass1=$(cat auth/1/password)
    user2=$(cat auth/2/username)
    pass2=$(cat auth/2/password)


    redis-cli ACL SETUSER "${user1}" on ">${pass1}" +@read ~* 
    redis-cli ACL SETUSER "${user2}" on ">${pass2}"
```
First create two basic-auth type secrets which will be mounted on `redis` pod and will be used by the script.
```bash
kubectl create secret -n demo generic rd-auth1 \
          --from-literal=username=alice \
          --from-literal=password='1234'
kubectl create secret -n demo generic rd-auth1 \
          --from-literal=username=bob \
          --from-literal=password='5678'

```
Then deploy redis with `init` configured. Here's a sample Redis manifest that runs this script at the start.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: standalone-redis
  namespace: demo
spec:
  version: 7.2.3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      projected:
        sources:
        - secret:
            name: rd-auth1
            items:
              - key: username
                path: auth/1/username
              - key: password
                path: auth/1/password
        - secret:
            name: rd-auth2
            items:
              - key: username
                path: auth/2/username
              - key: password
                path: auth/2/password
        - configMap:
            name: redis-init-script
  terminationPolicy: WipeOut
```
After successful deployment two acl users will be created in redis nodes named `alice` and `bob` as per referred secrets.

## Postgres

We have fixed a bug in VersionUpdate OpsRequest for Postgres, which was causing issues on updating major versions.

### Grafana Alerts Dashboard
In this release, we bring support for Postgres grafana summary dashboard with alerts for a specific Postgres instance.

To use the alert dashboards :

* First you need to have KubeDB operator and a Postgres instance running on your cluster along with Prometheus operator or Builtin Prometheus supported by Kubedb.

* Then create a file named overwrite.yaml having the following values. Make sure you change the values matching your needs.

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
    $ helm search repo appscode/kubedb-grafana-dashboards --version=v2024.2.14
    $ helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards -n kubeops --create-namespace --version=v2024.1.31 
    -f  /path/to/the/overwrite.yaml
    ```

## Point in Time Recovery

We have updated the backup directory structure for point-in-time recovery, providing a more organized layout. These changes are applicable to all of our archiver supported databases. More specifically : `MongoDB`, `MySQL` & `Postgres`. Going forward, the directories will follow the following standardized structure:

- **Full Backup:** `<sub-directory>/<database_namespace>/<database_name>/full-backup`
- **Manifest Backup:** `<sub-directory>/<database_namespace>/<database_name>/manifest-backup`
- **Binlog Backup (MySQL/MariaDB):** `<sub-directory>/<database_namespace>/<database_name>/binlog-backup`
- **Oplog Backup (MongoDB):** `<sub-directory>/<database_namespace>/<database_name>/oplog-backup`
- **WAL Backup (PostgreSQL):** `<sub-directory>/<database_namespace>/<database_name>/wal-backup`


We've added TLS support and included support for NFS, GCS, and Azure Storage as backend options for MySQL point-in-time recovery.

> **Breaking changes**: We have made the volumeExpansion mode required. `ops.volumeExpansion.mode` is not being defaulted now. Similar changes are also done for `<scaler>.spec.storage.<node_type>`.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2024.2.14/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2024.2.14/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
