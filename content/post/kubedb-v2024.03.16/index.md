---
title: Announcing KubeDB v2024.3.16
date: "2024-03-27"
weight: 14
authors:
- Raihan Khan
tags:
- alert
- cloud-native
- dashboard
- database
- druid
- elasticsearch
- ferretdb
- grafana
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
---

We are pleased to announce the release of [KubeDB v2024.3.16](https://kubedb.com/docs/v2024.3.16/setup/). This release is primarily focused on adding monitoring support for `Kafka`, `RabbitMQ`, `Zookeeper`, `SingleStore` and `Pgpool`. Besides monitoring support additions of new database versions and a few bug fixes are coming out in this release. Postgres is getting replication slot support through this release as well. We have also focused on updating some of the metrics exporter sidecar images resulting in less CVEs and more stability. This post lists all the major changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.3.16/README.md). Let’s see the changes done in this release.



## Postgres
In this release we have introduced replication slot support for Postgres Database. Our replication slot support can tolerate failover. Below is a sample yaml that you need to use for using replication slot:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: bu-postgres
  namespace: demo
spec:
  replicas: 1
  storageType: Durable
  replication:
    walLimitPolicy: "ReplicationSlot"
    maxSlotWALKeepSize: -1
  terminationPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  version: "16.1"

```
Here for walLimitPolicy, supported values are `WALKeepSegment`, `WALKeepSize`, `ReplicationSlot`.

If your used postgres major version is below `13`, then you’ll have to use:
```yaml
  replication:
    walLimitPolicy: "WALKeepSegment"
    walKeepSegment: 64
```

If your used postgres major version is more than `12`, then you can use either of them:
```yaml
  replication:
    walLimitPolicy: "WALKeepSize"
    walKeepSize: 1024
```
```yaml
  replication:
    walLimitPolicy: "ReplicationSlot"
    maxSlotWALKeepSize: -1
```

Note `walKeepSegment`, `walKeepSize`, `maxSlotWALKeepSize` resembles `wal_keep_segment`, `wal_keep_size` and `max_slot_wal_keep_size` of `postgresql.conf` file.


## Redis

Redis metrics exporter images has been updated to the latest version `v1.58.0`. The latest version contains several CVE fixes.

## Elasticsearch

Metrics exporter images for Elasticsearch has been updated from `v1.3.0` to `v1.7.0`. The updated image contains less CVE and have a few bug fixes. These changes are applicable to both the elasticsearch versions with xpack plugins and OpenSearch. 

## Zookeeper

This release includes Grafana dashboards for easier monitoring of KubeDB managed ZooKeeper. The grafana dashboard shows several ZooKeeper specific data, status and diagram of memory and cpu consumption. You can check the dashboard to see the overall health of your zookeeper servers easily. As usual KubeDB provided Grafana dashboards comes in a bundle of three - Summary dashboard for overall monitoring of the database cluster, Database dashboard for database insights and Pod dashboard for monitoring database pods specifically. Here's a preview of the summary dashboard for Zookeeper.

![zookeeper-dashboard](images/kubedb-zookeeper-summary.png)

A step-by-step guide to monitoring is given here ( https://github.com/appscode/grafana-dashboards/tree/master/zookeeper)
We have also added configurable alerting support for KubeDB ZooKeeper. You can configure Alert-manager to get notified when a metrics of zookeeper servers exceeds a given threshold.

To learn more, have a look here ( https://github.com/appscode/alerts/tree/master/charts )


## RabbitMQ

This release includes Grafana dashboards for easier monitoring of KubeDB managed RabbitMQ. The grafana dashboards shows several RabbitMQ specific data, status and diagram of memory and cpu consumption. You can check the dashboard to see the overall health of your RabbitMQ servers easily. As usual KubeDB provided Grafana dashboards comes in a bundle of three - Summary dashboard for overall monitoring of the database cluster, Database dashboard for database insights and Pod dashboard for monitoring database pods specifically.
A step-by-step guide to monitoring is given here ( https://github.com/appscode/grafana-dashboards/tree/master/rabbitmq)
We have also added configurable alerting support for KubeDB RabbitMQ. You can configure Alert-manager to get notified when a metrics of rabbitmq servers exceeds a given threshold.

To learn more, have a look here ( https://github.com/appscode/alerts/tree/master/charts )


## Kafka

This release includes Grafana dashboards for easier monitoring of KubeDB managed Kafka. The grafana dashboard shows several Kafka broker and cluster specific data, status and diagram of memory and cpu consumption. You can check the dashboard to see the overall health of your Kafka brokers easily. As usual KubeDB provided Grafana dashboards comes in a bundle of three - Summary dashboard for overall monitoring of the database cluster, Database dashboard for database insights and Pod dashboard for monitoring database pods specifically.

A step-by-step guide to monitoring is given here ( https://github.com/appscode/grafana-dashboards/tree/master/kafka).

In this release, KubeDB is introducing support for Kafka Connect Cluster monitoring using Grafana dashboards. These dashboards give you clear visuals on how your clusters are doing, including things like connectors details, worker details, tasks details, performance metrics and resource usage. It's a simple way to make sure everything is running smoothly. Here's a sneak preview of the summary dashboard for connect cluster.

![connect-cluster dashboard](images/kubedb-kafka-connectcluster-summary.png)

To get started with monitoring, we've prepared a step-by-step guide available at: (https://github.com/appscode/grafana-dashboards/tree/master/connectcluster).

We have also added configurable alerting support for KubeDB Kafka. You can configure Alert-manager to get notified when a metrics of kafka brokers exceeds a given threshold.

To learn more, have a look here ( https://github.com/appscode/alerts/tree/master/charts).

### Autoscaler

This release includes support for `KafkaAutoscaler`, a Kubernetes Custom Resource Definitions (CRD). It provides a declarative configuration for autoscaling `Kafka` compute resources and storage of database components in a Kubernetes native way. Let’s assume we have a KubeDB managed kafka cluster running in topology mode named `kafka-prod`. Here’s a sample yaml for autoscaling `Kafka` compute resources.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kafka-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: kafka-prod
  compute:
    broker:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
    controller:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
```

#### New Version Supports
In this release, we have extended support to include two new versions each for both Kafka and Kafka Connect Cluster.
- `3.6.1`
- `3.5.2`
Note: We have deprecated versions `3.4.0` and `3.3.0` from this release. These versions are unstable and will not be maintained in the upcoming releases. We recommend using versions `3.4.1` and `3.3.2`. 


## SingleStore

This release introduces an enhanced monitoring feature for KubeDB-managed SingleStore deployments by integrating the Grafana dashboard. This dashboard offers comprehensive insights into various SingleStore-specific metrics, including data status and visualizations of memory and CPU consumption. With this dashboard, users can effortlessly assess the overall health and performance of their SingleStore clusters, enabling more informed decision-making and efficient management of resources. Here's a preview of the summary dashboard.
![Singlestore dashboard](images/kubedb-singlestore-summary.png)

Here is the step-by-step guideline (https://github.com/appscode/grafana-dashboards/tree/master/singlestore)
We have added configurable alerting support for KubeDB SingleStore. Users can configure Alertmanager to receive notifications when a metric of SingleStore exceeds a given threshold.

To learn more, have a look here ( https://github.com/appscode/alerts/tree/master/charts )

### Backup and Restore

This release introduces support for comprehensive disaster recovery with Stash 2.0, also known as Kubestash. Shortly Kubestash offered by AppsCode, provides a cloud-native backup and recovery solution for Kubernetes workloads, streamlining operations through its operator-driven approach. It facilitates the backup of volumes, databases, and custom workloads via addons, leveraging restic or Kubernetes CSI Driver VolumeSnapshotter functionality. For SingleStore, creating backups involves configuring resources like BackupStorage (for cloud storage backend), RetentionPolicy (for backup data retention settings), Secret (for storing restic password), BackupConfiguration (specifying backup task details), and RestoreSession(specifying restore task details).

### SingleStore Studio (UI)

This release also introduces integrated SingleStore Studio (UI) with SingleStore. To connect to SingleStore Studio need to hit `8081` port of the primary service for your database. Here's a preview of the Studio UI.

![singlestore ui](images/singlestore-ui.png)

#### New Version Support
This release adds support for SingleStore `v8.5.7`.

## Pgpool
In this latest release, KubeDB now supports monitoring for Pgpool which includes Grafana dashboards tailored specifically for monitoring KubeDB-managed Pgpool instances. These dashboards provide comprehensive insights into various Pgpool-specific metrics, statuses, as well as visual representations of memory and CPU consumption. 

To get started with monitoring, we've prepared a step-by-step guide available at: (https://github.com/appscode/grafana-dashboards/tree/master/pgpool).

Additionally, we've introduced configurable alerting support for KubeDB Pgpool. Now, you can easily set up alerts to receive notifications based on customizable alert rules. Here's a sneak preview fo the summary dashboard when alert is enabled.

![pgpool dashboard](images/kubedb-pgpool-summary.png)

For more details and to explore these new alert capabilities further, please visit: (https://github.com/appscode/alerts/tree/master/charts).


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2024.3.16/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2024.3.16/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
