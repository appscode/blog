---
title: Announcing KubeDB v2022.05.24
date: 2022-05-24
weight: 25
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
  - elasticsearch-dashboard
  - kibana
  - schema-manager
---

We are pleased to announce the release of [KubeDB v2022.05.24](https://kubedb.com/docs/v2022.05.24/setup/). This post lists all the major changes done in this release since the last release. This release offers some major features like **MySQL Semi-Synchronous Replication**, **MongoDB Arbiter**, **PGBouncer**, **MariaDB Schema Manager**, **ProxySQL**, **Redigned Redis**, **Elasticsearch V8**, **OpenSearch Dashboard**, etc. It also contains various improvments and bug fixes. You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.05.24/README.md).

## MySQL

We have added support for `semi-synchronous` replication. Now you will be able to provision, take backups, and establish a TLS encrypted connection in a  cluster of semi-synchronous replication.

Sample MySQL YAML:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
 name: mysql
 namespace: demo
spec:
 version: "8.0.29"
 replicas: 3
 topology:
   mode: SemiSync
   semiSync:
     sourceWaitForReplicaCount: 1
     sourceTimeout: 24h
     errantTransactionRecoveryPolicy: PseudoTransaction
 storageType: Durable
 storage:
   storageClassName: "standard"
   accessModes:
     - ReadWriteOnce
   resources:
     requests:
       storage: 10Gi
 terminationPolicy: WipeOut
```

**What's New in MYSQL CRO?**

In the `spec.topology` you refer to `SemiSync` Mode and in the `spec.topolgy.semiSync` you can refer to how many replicas, the source will wait for an acknowledgment before committing, how long the source will wait for replicas before going back to asynchronous replication, and the errant transaction recovery policy. We support two recovery policies for errant transactions that is PseudoTranscation and Clone. 

```yaml
spec:
 topology:
   mode: SemiSync
   semiSync:
     sourceWaitForReplicaCount: 1
     sourceTimeout: 24h
     errantTransactionRecoveryPolicy: PseudoTransaction
```

**MySQL 8.0.29:** MySQL newer version support is added. You use Stand Alone, Group Replication, Semi-sync, and Read Replica features using KubeDB.

**Update Replication User Password:** The replication user Password was updated to `MYSQL_ROOT_PASSWORD` . For the older version if you upgrade to `8.0.29` or any upper version it will be updated.

**Improvement and Bug Fixes:** Several improvements and bug fixes are done in the operator regarding patch and health checker.

## ProxySQL

The KubeDB ProxySQL is now supporting `ProxySQL-2.3.2` in the latest release.

**Backend Version support:** KubeDB ProxySQL allows MySQL-5 and MySQL-8 as backend from the latest release.

**Custom configuration:** You can now bootstrap KubeDB ProxySQL with your own custom configuration file.  

**Clustering:** From this release, we have added the clustering feature for ProxySQL. You can create a ProxySQL cluster with 3 or more nodes. Also, failover recovery for clusters has been implemented for both partial and complete cluster failure.

**TLS Configuration:** KubeDB now allows you to establish TLS secured connections for both ProxySQL backend and ProxySQL frontend. Cert-manager can be used for issuing certificates.

**Monitoring:** All the monitoring essentials are available for KubeDB ProxySQL.

## MongoDB

We have added MongoDB `Arbiter` support on replica-set and sharded databases. Arbiter is a special type of node which does not have a copy of the data set and cannot become a primary. It is a `priority-0` member that always contains exactly `1 election vote`. Thus it allows replica sets to have an uneven number of voting members without the overhead of an additional member that replicates data. So now you can provision Kubernetes-integrated MongoDB even if your cost constraints prohibit adding another secondary.

Sample MongoDB YAML:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb
  namespace: db
spec:
  allowedSchemas:
    namespaces:
      from: All
  version: "4.4.6"
  replicaSet:
    name: "replicaset"
  replicas: 2
  arbiter:
    podTemplate:
      spec:
        resources:
          requests:
          cpu: "500m"
          memory: "400Mi"
```

## PostgreSQL

IN this release, the Pg-Coordinator container has been improved. Now it exposes Prometheus metrics to monitor the Primary status of the pod. It also comes with various bug fixes.

## PGBouncer

In this release, KubeDB-managed PGBouncer is now supporting the latest pgBouncer version `1.17.0`. The new release also supports a `TLS secured` connection for both client and server(PostgreSQL). It also enables support for Prometheus `monitoring` for pgBouncer. To expand the compatibility with third-party software, we are creating an `AppBinding` with connection credentials so that any third-party software can communicate seamlessly with the pgBouncer.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: pg
spec:
  version: "1.17.0"
  replicas: 3
  databases:
  - alias: "tmpdb"
    databaseName: "tmpdb"
    databaseRef:
      name: "postgres-demo"
      namespace: pg
  connectionPool:
    port: 5432
    defaultPoolSize: 20
    poolMode: session
    minPoolSize: 0
    maxClientConnections: 20
    reservePoolSize: 5
    maxDBConnections: 0
    maxUserConnections: 0
    statsPeriodSeconds: 60
    authType: md5
    adminUsers:
    - myuser
    authUser: myuser
  userListSecretRef:
    name: db-user-pass
  terminationPolicy: WipeOut
 ```

## Redis

We have redesigned the KubeDB Redis operator. We have distributed responsibility across different components so that there is no single point of failure. The primary task of the KubeDB Redis operator is to provision the cluster according to the provided configuration. After the KubeDB Redis operator successfully provisions the cluster, it no longer depends on the operator. The pods of the cluster can rejoin the cluster by themselves in case of node failure or network issues. We have fixed some minor bugs in KubeDB Redis Operator.

## MariaDB

For KubeDB MariaDB, better health checking has been added. KubeDB MariaDB CR now supports more accurate status conditions.

## Schema Manager

We have added `Multi-Tenancy` support for `MariaDB` in this latest release. With schema-manager now you can create a database, initialize a database using a script, restore a snapshot  for MariaDB as well. So, KubeDB now supports managing schemas with declarative YAML for MariaDB.

Sample MariaDBDatabase YAML:

```yaml
apiVersion: schema.kubedb.com/v1alpha1
kind: MariaDBDatabase
metadata:
 name: sample-mariadb-schema
 namespace: demo
spec:
 database:
   serverRef:
     name: mariadb-server
     namespace: dev
   config:
     name: demo
     characterSet: utf8
     encryption: disable
     readOnly: 0
 vaultRef:
   name: vault
   namespace: dev
 accessPolicy:
   subjects:
     - kind: ServiceAccount
       name: "tester"
       namespace: "demo"
   defaultTTL: "10m"
 deletionPolicy: "Delete"
```

The given YAML will create a MariaDB database schema in the cluster which will eventually end up creating a database demo inside  mariadb-server which is a KubeDB MariaDB server. The demo database will be created with characterset `utf8` and `encryption disabled`  as it is mentioned in the `spec.database.configuration` section. The operator will create a user and its credentials using the `KubeVault` and bind the permission under the subject which is mentioned in the `spec.accessPolicy.subjects` section. The credential and permission will be valid for 10 minutes. The time duration is  mentioned in the `spec.accessPolicy.defaultTTL` section. After that, the schema will be expired.

## Elasticsearch

In this release, We are happy to introduce the support for `Elasticsearch V8`. Now, you can provision and manage Elasticsearch V8 with KubeDB. New supported ElasticsearchVersion CROs are `xpack-8.2.0`, `kubedb-xpack-8.2.0`, and `opensearch-1.3.2`. The `kubedb-xpack-8.2.0` packed with  `elasticsearch:8.2.0` with pre-installed plugins; `repository-s3`, `repository-azure`, `repository-hdfs`, and `repository-gcs`.

**Configure Built-In Users:** KubeDB Elasticsearch CR now allows users to configure and manage Elasticsearch built-in users. Supported users are `elastic` , `kibana_system`, `logstash_system`, `beats system`, `apm_system`, and `remote_monitoring_user`. Passwords can also be customized via user-provided k8s secrets. Otherwise, the operator will create randomly generated passwords for those supported users. The k8s secrets with the user credentials can be found on the same namespace of the Elasticsearch CRO.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  version: xpack-8.2.0
  enableSSL: true
  replicas: 3
  storage:
    storageClassName: "standard"
    resources:
      requests:
        storage: 1Gi
  internalUsers:
      kibana_system:
        secretName: kibana-system-custom-cred
```

## Dashboard

In this release, we are expanding our support for dashboards. Now KubeDB supports `opensearch-dashboard` with ElasticsearchVersion CRO `opensearch-1.1.0`, `opensearch-1.2.0`, and `opensearch-1.3.2`. It also supports kibana  with ElasticsearchVersion CRO `xpack-8.2.0`, `xpack-7.17.3`, `xpack-7.16.2`, `xpack-7.13.2`, `xpack-7.12.0`, `xpack-7.9.1`, and `xpack-6.8.22`.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.05.24/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.05.24/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
