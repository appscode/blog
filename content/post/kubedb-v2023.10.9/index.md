---
title: Announcing KubeDB v2023.10.9
date: "2023-10-13"
weight: 14
authors:
- Mehedi Hasan
tags:
- cloud-native
- dashboard
- database
- day-2-operations
- elasticsearch
- kafka
- kubedb
- kubedb-cli
- kubernetes
- mariadb
- mongodb
- mysql
- opensearch
- percona
- percona-xtradb
- pgbouncer
- postgresql
- prometheus
- proxysql
- redis
- security
---

We are pleased to announce the release of [KubeDB v2023.10.9](https://kubedb.com/docs/v2023.10.9/setup/). This post lists all the major changes done in this release since the last release. The release includes -
- **Remote Replica for PostgreSQL & MySQL** ⇒ One of the major feature of this release , Now you can replicate PostgreSQL and MySQL across cluster using remote replica. 
- **OpenSearch hot-warm-cold cluster** ⇒ resource optimization using different hardware profiles
- **Kafka OpsRequest** ⇒ Day2 operations for KubeDB managed Kafka
- **CLI** ⇒ Generate Remote Replica Config
Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.10.9/README.md). Let’s see what are the database specific changes coming with this release.

## Postgres:
**Support for Remote Replica**: In this release, we have added support for Remote Replica. Now you can replicate your PostgreSQL databases `in and across` cluster. Remote Replica
- can be used to scale of read-intensive workloads
- can be used as workaround for your  BI and analytical workloads
- can be geo-replicated across cluster

**Follow the [Remote replica](https://kubedb.com/docs/v2023.10.9/guides/postgres/remote-replica/remotereplica/) documentation for details.**

## MySQL:
**Supports for Remote Replica** We have also added support for Remote Replica for Mysql as well. Now you can replicate your MySQL databases `in and across` cluster. Remote Replica
- can be used to scale of read-intensive workloads
- can be used as workaround for your  BI and analytical workloads
- can be geo-replicated across cluster

**Follow the [Remote replica](https://kubedb.com/docs/v2023.10.9/guides/mysql/clustering/remote-replica/) documentation for details.**

## Kafka:
From this release, Kafka will be using a separate advertised listener for listening to clients to establish connections with the correct network interfaces and addresses that are accessible from their environment. We are calling this **External** listener. This release also brings support for `spec.serviceTemplates` in Kafka API which is an optional field that can be used to provide template for the primary service created by KubeDB operator for Kafka.

This release includes support for `KafkaOpsrequest`, a Kubernetes Custom Resource Definitions (CRD). It provides a declarative configuration for the Apache Kafka administrative operations like database version update, horizontal scaling, vertical scaling, reconfiguration etc. in a Kubernetes native way. Let’s assume we have a KubeDB managed kafka cluster running with 2 brokers and 2 controllers. Here’s a sample yaml for scaling up kafka cluster horizontally:

``` yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: horizontal-scaling-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: kafka-prod
  horizontalScaling:
    topology:
      broker: 3
      controller: 3

```

## Elasticsearch/OpenSearch:

The **hot-warm-cold** architecture aims for resource optimization using different hardware profiles for different phases of data life.  KubeDB introduces dedicated hot, warm and cold nodes for opensearch in this release. Configure each types of nodes with dedicated attributes and configurations. Support for `data_content`, `ml`, `transform` and `frozen` nodes is also coming in this release. Here’s a sample yaml for deploying a simple Hot Warm Cold cluster.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: opensearch-hwc
  namespace: demo
spec:
  version: opensearch-2.8.0
  enableSSL: true
  storageType: Durable
  topology:
    master:
      suffix: master
      replicas: 2
      storage:
        storageClassName: "linode-block-storage" 
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      suffix: client
      replicas: 2
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    dataHot:
      replicas: 3
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        requests:
          cpu: 1.5
          memory: 2Gi
        limits:
          cpu: 2
          memory: 3Gi
    dataWarm:
      replicas: 2
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        limits:
          cpu: 1
          memory: 2Gi
    dataCold:
      replicas: 1
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        requests:
          cpu: .5
          memory: 1Gi
        limits:
          cpu: .5
          memory: 1.5Gi
```
**BugFix:** Fix for `Opensearch v2 failing to be reconfigured using ElasticsearchOpsRequest` and `ReconfigureTLS opsrequest failing to transform operator managed TLS certificates to cert-managed managed certificates` have been addressed in this release.

**New Version support:** Support for opensearch v1.13.3 has been added in this release.

## MongoDB:
**Improvement:** Health check for unusual secondary lock. In the health checker, We now check if some mongodb secondaries are locked and synced behind the Primary.   If these types of lock are found, that are not done by stash addons, We set the ReplicaReady condition to false, in that case. 

## KubeDB ClI:
We have added a new commands in KubeDB cli to help you generate `remote replica config` for KubeDB managed databases. You can follow the [instruction](https://kubedb.com/docs/v2023.08.18/setup/install/kubectl_plugin/) to download/update the CLI

```bash
kubectl dba remote-config <sub-command> <db-kind>  -n <ns> <db-name> -d <dns> -u<user_to_gen> -p<pass>

#Examples
kubectl dba remote-config postgres -n demo demo-pg -uremote -ppass -d 172.104.37.147 
kubectl dba remote-config mysql -n demo demo-mysql -uremote -ppass -d 172.104.37.147 

```

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.10.9/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.10.9/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
