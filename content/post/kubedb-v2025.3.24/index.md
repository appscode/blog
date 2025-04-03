---
title: Announcing KubeDB v2025.3.24
date: "2025-03-27"
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
- **Operator Sharding Support** has been added to this release.
- **Virtual Secret** Support has been added.

## Operator-Shard-Manager
Detailed blog on Operator Shard Manger has been written [here](https://appscode.com/blog/post/operator-shard-manager-v2025.3.14/).
 
## Virtual Secrets
In this release, Virtual Secrets support has been integrated into KubeDB, with initial support for PostgreSQL. Virtual Secrets allows you to store database auth secret data securely in an central external secret manager while maintaining the same functionality as native Kubernetes secrets for a user point of view.
- Learn more about Virtual Secrets [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/).
- Follow the step-by-step guide to use Virtual Secrets with KubeDB [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/#use-virtual-secrets-with-kubedb).

## Elasticsearch

### New Version
In this release we have added `7.17.27-xpack`,`8.17.1-xpack`,`2.19.0-opensearch` new `ElasticsearchVersion`.

## FerretDB
We are thrilled to announce that from this release KubeDB supports general availability (GA) of FerretDB v2.0, a groundbreaking release that delivers a high-performance, fully open-source alternative to MongoDB, ready for production workloads. Version 2.0 introduces Over 20x faster performance powered by Microsoft DocumentDB, Replication support for high-availability, Vector search support for AI-driven use cases and many more.

We remove `spec.podTemplate` and `spec.replicas` sections from KubeDB FerretDB. Add `spec.server.primary` and `spec.server.secondary` field to provide information about primary and secondary servers about their replica count and podTemplate specification.

We also removed the `spec.backend` part from KubeDB FerretDB. We are no longer supporting externally managed postgres backend. The whole backend stuff will also be maintained by KubeDB. In `FerretDBVersion` we introduce a field `spec.postgres` which will hold the information about which postgres backend will be used for this FerretDB version.

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
In this release we have added `13.20`, `14.17`, `15.12`, `16.8`, and `17.4` new `PostgresVersion`.

### Improvements
In this release we have updated the raft library version that we were using for leader election to select the Postgres cluster primary

### Bug fix
We have fixed a bug that prevented the standby from joining back to the cluster.

## Solr

### New Version
In this release we have added `9.8.0` new `SolrVersion`.


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
