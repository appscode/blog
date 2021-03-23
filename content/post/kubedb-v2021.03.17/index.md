---
title: KubeDB v2021.03.17- Introducing MariaDB and Re-designed PostgreSQL operator
date: 2021-03-18
weight: 18
authors:
  - Md Kamol Hasan
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - mariadb
  - memcached
  - mongodb
  - mysql
  - postgresql
  - redis
  - kubedb
---

We are pleased to announce [KubeDB v2021.03.17](https://kubedb.com/docs/v2021.03.17/setup/). This post lists all the major changes done in this release since `v2021.01.26`.  This release offers **MariaDB** support with Galera Clustering, Backup and Recovery, TLS and many more features. It also contains the fix for the **PostgreSQL data loss issue**. The  support for the PostgreSQL TLS configuration and the support for official **TimescaleDB** images has also been added to this release.

## MariaDB

We are very excited to announce that MariaDB support has been added to KubeDB. It offers a number of cool features including Galera Clustering, TLS/SSL encryption, Backup, and Restore Custom Configuration and Monitoring.  

- **Multi-Master Support for clustering:** KubeDB operator supports Galera Cluster in MariaDB which offers a virtually synchronous multi-master cluster. That means reading and writing from any cluster node is possible.
- **Backup and Restore Database:** MariaDB supports backup and restores using Stash. For more information, visit [here](https://kubedb.com/docs/v2021.03.17/guides/mariadb/backup/overview/).
- **TLS Support:** To add an extra layer of security, KubeDB can automate provisioning SSL certificates to encrypt client/server communication. This is an enterprise feature.
- **Custom Configuration for MariaDB:** KubeDB provides custom configuration for MariaDB databases using a configuration file or using a pod template.
- **Monitoring via Prometheus:** Monitoring support has been added using a built-in Prometheus scraper and Prometheus operator.
- **Supported MariaDB version:** MariaDB version `10.5.8` and `10.4.17` are supported in this release.

## PostgreSQL

PostgreSQL operator has been updated to support official images provided by  2 distributions:

- Official PostgreSQL
- TimescaleDB

**Postgres Data Loss Issue Resolved:** In the new release, we have finally shifted from Kubernetes leader election to directly using raft protocol to elect new primary instance. For every pod, there will be a sidecar running along with a PostgreSQL container. This sidecar will make sure that among all active pods, the pod with the highest WAL position will be the leader or primary. And other replicas will try to sync with the primary. This addresses a long standing issue where a flaky connection to the Kubernetes control plane will result in loss of leader and trigger a sequence of events that might eventually lead to data loss. In this release, we have avoided the scenario where replicas have always taken base backup from the master whenever a failover happens. In case of failover, the redesigned operator will make sure that the current master is in the Ready status if replicas need to take base backup.

Updated leaderElection field:

```yaml
leaderElection:
    # 10*period = 10*100ms = 1000ms
    electionTick: 10 
    # 1*period = 1*100ms = 100ms
    heartbeatTick: 1 
    period: 100ms 
   # 33554432 = 32MB if the replica is lagging from master more than 32MB in wal position, replica is going to take base backup from master.
   maximumLagBeforeFailover: 33554432  
```

- **Support for latest versions:** KubeDB now supports PostgreSQL version `13.2`, `12.6`, `11.11`, `10.16`, and `9.6.21`.
- **TLS support for PostgreSQL cluster:** In this release, we have added support for encrypting client/server communications for increased security.
- **Client authentication mode:** We have introduced `SCRAM-SHA-256` authentication support to verify the user's password and client certificate authentication method when a PostgreSQL cluster has enabled SSL.
- **Wal-g support has been dropped:** We have dropped continuous archiving with wal-g for now. We are planning to bring back this feature via Stash (the backup solution) in the future.

## MySQL

- **Introduce new MySQL versions:** `8.0.23` and `5.7.33`.
- Various improvements made to the on-start script in MySQL docker image for fixing restart problems.
- Introduce clone plugin from version `8.0.20-v1` to `8.0.23` for new joiner(MySQL instance) into a group replicated cluster. The clone plugin permits cloning data from other MySQL server instances at the time of joining a group. This process will be automatically triggered when the joiner gets a valid donor and the primary member's data will be greater than or equal to `128MB`.
- By default set various environment variables for fixing memory leak issues and getting better performance. These settings are recommended and users can overwrite these by passing environment variables or using the pod template args. Recommended default settings are:
  - Reserved `256MB` memory for performance schema and other processes.
  - Then allocate `75%` of the available memory for InnoDB buffer pool size.
  - Then allocate the rest of the memory for group replication cache size.
- Improve MySQL docs: QuickStart, Concept, Clustering,  Upgrading, Scaling, TLS configuration, etc.
- `MySQLOpsRequest` spec is now immutable i.e. the spec can't be changed once the Ops Request has been created.
- Only One `MySQLOpsRequest` can be in the `Progressing` phase at a time.
- Added validation and mutation webhook for `MySQLOpsRequest`.
- Added support for using different cert-manager `Issuer`/`ClusterIssuer` for different certificates.

## MongoDB

- `MongoDBOpsRequest` spec is now immutable i.e. the spec can't be changed once the Ops Request has been created.
- Only One `MongoDBOpsRequest` can be in the `Progressing` phase at a time.
- Added validation webhook for `MongoDBOpsRequest`.
- Added support for using different cert-manager `Issuer`/`ClusterIssuer` for different certificates. This is necessary to use HashiCorp Vault to provision TLS certificates.
- Added **IPv6** support in MongoDB.
- Improved various `MongoDB` and `MongoDBOpsRequest` docs.

## Elasticsearch

- Replaced `busybox` images with `toybox`/`alpine` images for the init-containers.
- Added support for new Elasticsearch versions: `searchguard-7.3.2`, `searchguard-7.9.3`, `opendistro-1.10.1`, `opendistro-1.11.0`, `opendistro-1.12.0`.
- User provided custom config files are mounted on the `"$ES_CONFIG_DIR/custom-config/"` node directory.
- Update ENV while upgrading Elasticsearch version from `V6` to `V7`: like `discovery.zen.minimum_master_nodes` will become `cluster.initial_master_nodes`.
- Introduce a new ElasticsearchVersion naming convention: `{Distribution Name}-{Application Version}-{Modification Tag}`. Samples: `xpack-7.9.1-v1`,  `searchguard-7.9.3`,  `opendistro-1.12.0`, etc.
- Allow users to provide custom passwords for the default internal database users like `kibanaserver`, `readall`, etc.
- **Improve Elasticsearch docs:** Quickstart, Concepts, Clustering, Custom Configuration, Auto Scaling.
- Added support for **IPv6** in Elasticsearch.
- `ElasticsearchOpsRequest` spec is now immutable i.e. the spec can't be changed once the Ops Request has been created.
- Only One `ElasticsearchOpsRequest` can be in the `Progressing` phase at a time.
- Added validation webhook for ElasticsearchOpsRequest.
- Added support for using different cert-manager `Issuer`/`ClusterIssuer` for different certificates.

## Redis

- Added TLS reconfigure, Restart OpsRequests support for Redis.
- `RedisOpsRequest` spec is now immutable i.e. the spec can't be changed once the Ops Request has been created.
- Only One `RedisOpsRequest` can be in the `Progressing` phase at a time.
- Added validation webhook for `RedisOpsRequest`.
- Added support for using different cert-manager `Issuer`/`ClusterIssuer` for different certificates.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.03.17/setup).
- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.03.17/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
