---
title: Announcing KubeDB v2025.3.24
date: "2025.3.24"
weight: 16
authors:
- Saurov Chandra Biswas
tags:
- alert
- archiver
- autoscaler
- backup
- cassandra
- clickhouse
- cloud-native
- dashboard
- database
- druid
- grafana
- kafka
- kubedb
- kubernetes
- kubestash
- memcached
- mongodb
- mssqlserver
- network
- networkpolicy
- pgbouncer
- pgpool
- postgres
- postgresql
- prometheus
- rabbitmq
- recommendation
- redis
- restore
- s3
- security
- singlestore
- solr
- tls
- zookeeper
---

KubeDB **v2025.3.24** is now available! This latest release brings significant performance enhancements, improved reliability, and new features to the database management experience on Kubernetes. 
- We have switched to using StatefulSet for operators and removed all the APIServices for webhooks.
- **New Version Support**: New versions has been added for `Elasticsearch`, `FerretDB`, `Postgres` and `Solr`.
- **Operator Sharding**, **Virtual Secret**, and **GitOps** support  has been added to this release.

## GitOps Support

We’re thrilled to unveil GitOps support for database management across multiple databases, including `PostgreSQL`, `MongoDB`, and `Kafka`, with plans to expand further. This release introduces a new Custom Resource Definition (CRD) under the `gitops.kubedb.com/v1alpha1` API group, mirroring the familiar kinds and specs of existing database CRs while enabling seamless GitOps workflows. Here’s what this update delivers.

### Feature Overview

- **Dedicated GitOps CRD**: A new CRD under `gitops.kubedb.com/v1alpha1` replicates the same CR kinds (e.g., `PostgreSQL`, `MongoDB`, `Kafka`) and specs as those in the core KubeDB API group, purpose-built for GitOps-driven management.
- **Automated Database Provisioning**: When a CR from `gitops.kubedb.com/v1alpha1` is applied through a GitOps pipeline (e.g., ArgoCD, Flux), the GitOps Operator generates a matching database CR in the core KubeDB API group, provisioning the database and aligning it with your Git-defined desired state.
- **Intelligent Reconciliation**: The GitOps controller continuously monitors for differences between the desired state in `gitops.kubedb.com/v1alpha1` CRs (e.g., `spec.version`, `resources`, or `configurations`) and the actual database state, automatically triggering ops requests to reconcile them.

### Supported Ops Requests

The GitOps controller streamlines database lifecycle management with these automated ops requests:

- **Version Updates**: Generates an UpdateVersion ops request to sync the database with the specified version.
- **Resource Adjustments**: Triggers a resource update ops request to modify compute or storage as required.
- **Horizontal Scaling**: Initiates a scaling ops request to adjust the number of replicas, enabling scale-up or scale-down based on the desired state.
- **TLS Configuration**: Creates an ops request to add, update, or remove TLS settings.
- **Authentication Changes**: Launches an ops request to apply or modify authentication configurations.
- **Configuration Adjustments**: Generates an ops request to update database configurations, ensuring alignment with the desired state.

This automation ensures your databases remain fully aligned with the desired state in Git, minimizing manual effort and enhancing operational reliability and scalability.

Sample Yaml:
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  version: "16.6"
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## Operator-Shard-Manager
The Operator Shard Manager is designed using Consistent Hashing with bounded loads algorithm to efficiently distribute workloads among multiple controller instances.

Detailed blog on Operator Shard Manger has been written [here](https://appscode.com/blog/post/operator-shard-manager-v2025.3.14/).
 
## Virtual Secrets
In this release, Virtual Secrets support has been integrated into KubeDB, with initial support for PostgreSQL. Virtual Secrets allows you to store database auth secret data securely in an external secret manager while maintaining the same functionality as native Kubernetes secrets for a user point of view.
- Learn more about Virtual Secrets [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/).
- Follow the step-by-step guide to use Virtual Secrets with KubeDB [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/#use-virtual-secrets-with-kubedb).

## IRSA Annotation Key

Added IRSA annotation key to the `ServiceAccount` to enable credential-less mode for archiver-enabled databases, including `MariaDB`, `MongoDB`, `MSSQLServer`, `MySQL`, and `Postgres`.
## Elasticsearch

### New Version
In this release, we have added the following new `ElasticsearchVersion`s: `7.17.27-xpack`, `8.17.1-xpack`, and `2.19.0-opensearch`.

## FerretDB
We are thrilled to announce that from this release KubeDB supports general availability (GA) of FerretDB v2.0, a groundbreaking release that delivers a high-performance, fully open-source alternative to MongoDB, ready for production workloads. Version 2.0 introduces Over 20x faster performance powered by Microsoft DocumentDB, Replication support for high-availability, Vector search support for AI-driven use cases and many more.

We remove `spec.podTemplate` and `spec.replicas` sections from KubeDB FerretDB. Add `spec.server.primary` and `spec.server.secondary` field to provide information about primary and secondary servers about their replica count and podTemplate specification.

We also removed the `spec.backend` part from KubeDB FerretDB. We are no longer supporting externally managed postgres backend. The whole backend stuff will also be maintained by KubeDB. In `FerretDBVersion`, we introduce a field `spec.postgres` which will hold the information about which postgres backend will be used for this FerretDB version.

Here is the yaml example,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: fr
  namespace: demo
spec:
  version: "2.0.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  deletionPolicy: Delete
  server:
    primary:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: ferretdb
              resources:
                requests:
                  cpu: "200m"
                  memory: "300Mi"
                limits:
                  cpu: "200m"
                  memory: "300Mi"    
    secondary:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: ferretdb
              resources:
                requests:
                  cpu: "200m"
                  memory: "300Mi"
                limits:
                  cpu: "200m"
                  memory: "300Mi"
```



## MariaDB

In this release, we have introduced a new MariaDB topology mode called MariaDBReplication, which implements MariaDB's standard replication with a Master/Slave architecture. We leverage the MaxScale Proxy Server to manage automatic failover seamlessly.
KubeDB now supports two topology modes for MariaDB: Galera and MariaDBReplication.
Here is a sample yaml for MariaDBReplication topology mode-


```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: md-max
  namespace: demo
spec:
  version: 10.6.16
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      enableUI: true
      replicas: 3
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 100Mi
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  deletionPolicy: Delete
```


## Postgres

### New Version
In this release, we have added `13.20`, `14.17`, `15.12`, `16.8`, and `17.4` new `PostgresVersion`.

### Improvements
In this release, we have updated the raft library version that we were using for leader election to select the Postgres cluster primary

### Bug fix
We have fixed a bug that prevented the standby from joining back to the cluster.


## ProxySQL
We’re excited to announce that KubeDB ProxySQL now supports MariaDB Galera Clusters. This addition extends ProxySQL's capabilities beyond standard MySQL setups, enabling robust high-availability query routing and load balancing for Galera-based deployments.

### Key Highlights

- Galera Cluster Integration: ProxySQL can now be seamlessly deployed alongside MariaDB Galera clusters to manage traffic routing, load balancing, and failover handling across multiple database nodes.
- Automatic Hostgroup Assignment: ProxySQL detects read/write roles automatically and assigns nodes to appropriate hostgroups.
- Built-in Health Checks: Faulty nodes are automatically removed from the routing table, with graceful failover handling that keeps workloads running smoothly.

This release makes it easier to run production-grade, fault-tolerant MariaDB Galera clusters with intelligent traffic management and enhanced observability.

#### Sample YAMLs for Galera Setup with ProxySQL
Deploy MariaDB Galera with KubeDB:
```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-galera-cluster
  namespace: demo
spec:
  version: "10.6.16"
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

Deploy ProxySQL for Galera Cluster:
```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: mariadb-galera-proxy
  namespace: demo
spec:
  version: "2.6.3-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: mariadb-galera-cluster
  deletionPolicy: WipeOut
```

Internal ProxySQL Configuration for MariaDB Galera:   
Galera Hostgroup Mapping:
```sql
SELECT * FROM mysql_galera_hostgroups;
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind | comment |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| 2                | 4                       | 3                | 1                 | 1      | 1           | 1                     | 0                       |         |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
```
   
Registered Nodes in runtime_mysql_servers
```sql
SELECT * FROM runtime_mysql_servers;

+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                      | port | gtid_port | status  | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mariadb-galera-cluster-0.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-cluster-1.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-cluster-2.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-cluster-0.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-cluster-1.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-cluster-2.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | mariadb-galera-cluster-0.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | mariadb-galera-cluster-1.mariadb-galera-cluster-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
```

### MySQL Configuration Enhancement: From Service DNS to Pod DNS
In this release, we’ve improved ProxySQL’s integration with MySQL by switching from Kubernetes Service DNS names to individual Pod DNS names (via headless services).   

#### Previous Setup (Service-based DNS):
```sql
SELECT * FROM mysql_servers;
+--------------+-------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                      | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
| 2            | mysql-server.demo.svc         | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-standby.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
```
This configuration routed traffic to the Kubernetes service endpoint, which load-balanced connections across all pods, limiting observability and fine-grained control.

#### Current Setup (Pod-level DNS): 
```sql
SELECT * FROM mysql_servers;
+--------------+-------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                  | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
| 2            | mysql-server-0.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-0.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-1.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-2.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
```

####  Benefits of Pod-Level Configuration
- Direct Routing to Pods – Enables custom read/write distribution and replica lag handling.

- Fine-Grained Failover – Traffic can be redirected on a per-pod basis in case of failures.

- Improved Query Routing – Based on hostgroup roles, replication metrics, or pod health.

- Better Observability – Monitor status per pod (e.g., ONLINE, SHUNNED, OFFLINE_SOFT).

- Higher Resilience – Reduces reliance on service-level load balancing.



## Solr

### New Version
In this release, we have added `9.8.0` new `SolrVersion`.


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
