---
title: Announcing KubeDB v2024.12.18
date: "2024-12-23"
weight: 16
authors:
- Pritam Das
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

KubeDB **v2024.12.18** is now available! This latest release brings significant performance enhancements, improved reliability, and new features to database management experience on Kubernetes. Here are some of the key features to mention -
- **OpsRequest Support**: New `OpsRequest` features have been added for `Memcached`, `Microsoft SQL Server`, `MySQL`, offering greater flexibility for managing database administrative tasks. Moreover, a new `OpsRequest` feature named `ReplicationModeTransformation` has been introduced in this release.
- **Recommendation Engine**: Recommendation support for `KubeDB` managed kafka has been added.
- **New Version Support**: New versions have been added for `Druid`, `Elasticsearch`, `FerretDB`, `Kafka`, `MariaDB`, `Memcached`, `Microsoft SQL Server`, `MySQL`, `RabbitMQ`, `Redis`, `Solr`, `Singlestore`.
- **Archiver**: Archiver and point-in-time-recovery within `KubeDB` has been enhanced for `MongoDBArchiver`, `MariaDBArchiver`, `MySQLArchiver`, `PostgresArchiver`, `MSSQLServerArchiver`.

## Archiver
Archiver support has been enhanced for `MongoDBArchiver`, `MariaDBArchiver`, `MySQLArchiver`, `PostgresArchiver` and `MSSQLServerArchiver`. We have replaced the field `spec.walBackup` in spec with `spec.logBackup` with some enhancement. Besides two existing field `RuntimeSettings` and `ConfigSecret`, two more fields have been added and those are:

**SuccessfulLogHistoryLimit**: It defines the number of successful Logs backup status that the incremental snapshot will retain. The default value is 5.

**FailedLogHistoryLimit**: It defines the number of failed Logs backup that the incremental snapshot will retain for debugging purposes. The default value is 5.

You can find full spec [here](https://github.com/kubedb/apimachinery/blob/master/apis/archiver/v1alpha1/types.go#L74C1-L92C2).

We update start time and end time of continuous oplog/wal/log/binlog push in incremental snapshot status. From now on new fields in snapshot’s status field have been introduced.  So the maximum `successfulLogHistoryLimit` successful oplog/wal/log/binlog push and maximum `failedLogHistoryLimit` failed oplog/wal/log/binlog push information will be stored in our incremental snapshot’s status.

Here is a sample YAML for `MongoDBArchiver`. Changes in the field `.spec.logBackup` will be same for other archivers as well.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MongoDBArchiver
metadata:
  name: arch
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: "Same"
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: rp
    namespace: demo
  encryptionSecret:
    name: encry-secret
    namespace: demo
  logBackup:
    successfulLogHistoryLimit: 10
    failedLogHistoryLimit: 10
  fullBackup:
    driver: VolumeSnapshotter
    task:
      params:
        volumeSnapshotClassName: longhorn-backup-vsc   
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/50 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
```

## Recommendation Engine

In this Release, we are introducing `Recommendation` support for KubeDB managed Kafka instances. KubeDB ops-manager generates three types of recommendations for Kafka - Version Update Recommendation, TLS Certificates Rotation Recommendation and Authentication Secret Rotation Recommendation. Authentication Secret rotation is a new type of recommendation supported by KubeDB. It recommends to rotate authentication secret of a particular db if only one month remaining for rotating or two third of it's lifespan has been completed. 
Here's a sample rotate authSecret recommendation for a Kafka instance.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2024-12-19T18:40:02Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kafka-prod-tls
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-auth
  name: kafka-prod-tls-x-kafka-x-rotate-auth-4ikl43
  namespace: demo
  resourceVersion: "551880"
  uid: 450e1000-6db8-430b-9931-1218c0fa9f01
spec:
  backoffLimit: 5
  deadline: "2024-12-19T18:22:03Z"
  description: Recommending AuthSecret rotation,kafka-prod-tls-auth AuthSecret
    needs to be rotated before 2024-12-19 18:32:03 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: KafkaOpsRequest
    metadata:
      name: rotate-auth
      namespace: demo
    spec:
      databaseRef:
        name: kafka-prod-tls
      type: RotateAuth
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Kafka
    name: kafka-prod-tls
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```

## Druid

### New version
Support for `Druid` latest version `31.0.0` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid
  namespace: demo
spec:
  deepStorage:
    configSecret:
      name: deep-storage-config
    type: s3
  topology:
    routers:
      replicas: 1
  version: 31.0.0
```

## Elasticsearch

### New Version
We have added support for `Elasticsearch` versions `xpack-7.17.25`, `xpack-8.15.4`, `xpack-8.16.0` and `opensearch-1.3.19`.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  storageType: Durable
  enableSSL: true
  topology:
    data:
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    ingest:
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    master:
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  version: xpack-8.16.0
```

## Opensearch

### Bug Fix
The previous release added rotateAuth `OpsRequest` for KubeDB managed `Opensearch` clusters. The authentication secret was getting rotated successfully upon applying the `OpsRequest` manifest. However, the `Opensearch` configuration must be reloaded if its security feature is modified or updated. Hence, the Database was not ready internally. In this release, we have fixed this issue. RotateAuth OpsRequest for `Opensearch` ensures that database authentication secrets are rotated, configurations are reloaded to the cluster, and a new pair of credentials are accepted by `Opensearch` clients.

Here is the sample YAML for Opensearch RotateAuth OpsRequest.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: roatate-os
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: os-cluster
  type: RotateAuth
```

If the secret is referenced, the operator will update the `.spec.authSecret.name` with the new secret name. Here is the yaml of new Secret:
```yaml
apiVersion: v1
data:
  password: T0thOSlfKmo1TUgyd2ZzQg==
  username: YWRtaW4=
kind: Secret
metadata:
  name: new-auth
  namespace: demo
type: Opaque
```


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: roatate-os
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: os-cluster
  type: RotateAuth
  authentication:
    secretRef:
      name: new-auth
```

## Kafka

We have deprecated some versions of `Kafka` and added new versions in this release. Deprecated versions either contain major bugs or they have latest fix patches which are recommended to use.

**Deprecated Versions**: `3.3.2`,`3.4.1`,`3.5.1` and `3.6.0`.

**Added Versions**: `3.7.2`, `3.8.1` and `3.9.0`.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka
  namespace: demo
spec:
  deletionPolicy: Delete
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 3.9.0
```

### Kafka ConnectCluster

We have deprecated some connector versions of `postgres`, `mysql`, `mongodb` and added new versions in this release.

**Deprecated Versions**: `mongodb-1.11.0`, `mysql-2.4.2.final` and `postgres-2.4.2.final`.

**Added Versions**: `mongodb-1.13.1`, `mongodb-1.14.1`, `postgres-2.7.4.final`, `postgres-3.0.5.final`, `mysql-2.7.4.final`, `mysql-3.0.5.final`, `jdbc-2.7.4.final` and `jdbc-3.0.5.final`.

Here is a sample YAML for `Kafka ConnectCluster` with some new versions of connectors.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connect-cluster
  namespace: demo
spec:
  version: 3.9.0
  replicas: 2
  connectorPlugins:
  - mongodb-1.14.1
  - mysql-3.0.5.final
  - postgres-3.0.5.final
  - jdbc-3.0.5.final
  kafkaRef:
    name: kafka-prod
    namespace: demo
  deletionPolicy: WipeOut
```

## FerretDB

### New Version
Support for `FerretDB` latest version `1.24.0` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb
  namespace: demo
spec:
  authSecret:
    externallyManaged: false
  backend:
    externallyManaged: false
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  version: 1.24.0
```


## MariaDB

### New version
Support for `MariaDB` latest version `11.6.2` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-demo
  namespace: demo
spec:
  deletionPolicy: Delete
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 11.6.2
```


## Memcached

### New version
Support for `Memcached` latest version `1.6.33` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-demo
  namespace: demo
spec:
  deletionPolicy: Delete
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
  replicas: 3
  version: 1.6.33
```


### OpsRequest

`RotateAuth` OpsRequest for `Memcached` has been added. If a user wants to update the authentication credentials for a particular database, they can create an `OpsRequest` of type `RotateAuth` with or without referencing an authentication secret.
If the secret is not referenced, the ops-manager operator will create a new credential and update the current secret. Here is the Yaml for rotating authentication credentials for a `Memcached` cluster using `MemcachedOpsRequest`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: memcd-quickstart
```

If the secret is referenced, the operator will update the `.spec.authSecret.name` with the new secret name. Here is the yaml of new Secret:

```yaml
apiVersion: v1
data:
 authData: dXNlcjpwYXNzCg==
kind: Secret
metadata:
   name: mc-new-auth
   namespace: demo
type: Opaque
```


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: memcd-quickstart
  authentication:
    secretRef:
      name: mc-new-auth
```

Finally, the operator will update the database cluster with the new credential and the old credentials will be stored in the secret with keys username.prev and password.prev.
We have added a field `.spec.authSecret.activeFrom` to the db yaml which refers to the timestamp of the credential is active from. We also add an annotation `kubedb.com/auth-active-from` in currently using auth secret which refer to the active from time of this secret.

## Microsoft SQL Server

### New Feature: Rotate Authentication Credentials (`RotateAuth`)

A new `OpsRequest` has been introduced to simplify updating authentication credentials for SQL Server deployments. If users want to update the authentication credentials, they can create an `OpsRequest` of type `RotateAuth` with or without referencing an authentication secret.


#### Rotate Authentication Without Referencing a Secret

If no secret is referenced, the `ops-manager` operator will generate new credentials and updates the existing secret with the new credentials.

**Example YAML:**
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mssql-ag-cluster
  timeout: 5m
  apply: IfReady
```

#### Rotate Authentication With a Referenced Secret

If a secret is referenced, the operator will update the `.spec.authSecret.name` field with the new secret name. Archives the old credentials in the existing secret under the keys `username.prev` and `password.prev`.


**New Secret Example:**
```yaml
apiVersion: v1
data:
  password: dDh5Nmh1WTdoamczR3NZZQ==
  username: c2E=
kind: Secret
metadata:
  name: mssql-ag-cluster-new-auth
  namespace: demo
type: kubernetes.io/basic-auth
```
**Example YAML with Secret Reference:**
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: mssql-ag-cluster
  authentication:
    secretRef:
      name: mssql-ag-cluster-new-auth
```

**Enhancements for Credential Activation Tracking**
- Added `.spec.authSecret.activeFrom` in the database YAML to specify the activation timestamp of new credentials.
- Added `kubedb.com/auth-active-from` annotation in the current authentication secret, indicating when it became active.

  
### New version Support
Added support for SQL Server version `2022-CU16-ubuntu-22.04`, providing compatibility with the latest SQL Server features and updates.


## MySQL
In this release, we added support for new `MySQL` version, improved `MySQL` continuous archiving and `PITR` within `KubeDB`, and `MySQL` replication mode change(remote replica to group replication) ops-request.

### New Version
Support for MySQL version `8.4.3`, `9.0.1`, `9.1.0` has been added in the new release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-demo
  namespace: demo
spec:
  deletionPolicy: Delete
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 9.1.0
```


### Archiver
In this release, we’ve introduced several enhancements to improve MySQL continuous archiving and point-in-time recovery within KubeDB. Below is an overview of the key updates:

***Restic Driver for Base Backup Support***:
We now support the `Restic` driver for MySQL continuous archiving and recovery. Previously, only the `VolumeSnapshotter` driver was available.
To use the Restic driver, configure the `MySQLArchiver` Custom Resource (CR) by setting `.spec.fullBackup.Driver` to "Restic"

***Replication Strategies for MySQL Archiver Restore***:
A new replication strategy feature has been introduced, supporting four distinct approaches to restoring MySQL replicas. The available strategies are as follows:

- ***none***: Each `MySQL` replica independently restores the base backup and binlog files. After completing the restore process, the replicas individually join the replication group.
- ***sync***: The base backup and binlog files are restored exclusively on pod-0. Other replicas then synchronize their data by leveraging the MySQL clone plugin to replicate from pod-0.
- ***fscopy***: The base backup and binlog files are restored on pod-0. The data is then copied from pod-0's data directory to the data directories of other replicas using file system copy. Once the data transfer is complete, the group replication process begins.
- ***clone***: Each `MySQL` replica independently restores the base backup and binlog files. After completing the restore process, the replicas individually join the replication group. The clone strategy works for only VolumeSnapshotter driver.

Please note that `fscopy` does not support cross-zone operations.

Here is a sample YAML configuration for setting up a `MySQLArchiver` in KubeDB:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MySQLArchiver
metadata:
  name: mysqlarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: rp
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "Restic"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "storage"
      namespace: "demo"
```

Here’s the sample YAML configuration for restoring MySQL from a backup using the new features:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: restore-mysql
  namespace: demo
spec:
  init:
    archiver:
      replicationStrategy: sync
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mysql-full
        namespace: demo
      recoveryTimestamp: "2024-12-02T06:38:42Z"
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### OpsRequest
In this release, we have added the MySQL replication mode change, which is currently supported for remote replication to group replication and version `8.4.2` or above. Here is the YAML file of ops-request.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: mysql-replication-mode-change
  namespace: demo
spec:
  type: ReplicationModeTransformation
  databaseRef:
    name: mysql-remote-replica
  replicationModeTransformation:
    mode: Multi-Primary
    requireSSL: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  timeout: 10m
  apply: Always
```

Here you can mention the mode of group replication single or Multi primary, requireSSL and issuerRef for TLS secure connection on group replication mode.

## Postgres

### Point-in-Time Recovery (PITR) Enhancements:
Our PITR algorithm has been significantly improved. We now support the latest point-in-time recovery for PostgreSQL.
In order to find latest point  in time a user can recover, they need to follow the below the approach.
- Run `kubectl get snapshots.storage.kubestash.com -n <ns> <db-name>-incremental-snapshot`
    ```yaml
    status:
      components:
        wal:
          logStats:
            end: "2024-12-19T15:36:50.432136Z"
            lsn: 0/1C000468
    ```
- Use the time mentioned in `components.wal.logStats.end` as your recoveryTimeStamp.
### Bug Fixes and Improvements:
#### PostgreSQL Failover:
- Improved failover algorithms ensure almost instant failover to a healthy standby pod.
- Fixed several bugs that prevented failover when the old primary was not available in the cluster.
#### PostgreSQL Version Upgrade:
- Resolved timing issues in the PostgreSQL version upgrade ops-request.

## RabbitMQ

### New Version
This release adds support for RabbitMQ version `4.0.4`. Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq
  namespace: demo
spec:
  version: "4.0.4"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

## Redis

### New Version
Support for `Redis` latest versions `6.2.16`, `7.2.6` and `7.4.1` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: rd
  namespace: demo
spec:
  version: 7.4.1
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

## Singlestore

### New Version
Support for `Singlestore` latest versions `8.7.21` and `8.9.3` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.9.3"
  topology:
    aggregator:
      replicas: 2
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
  deletionPolicy: WipeOut
```

## Solr

### New Version
Support for `Solr` latest versions `8.11.4`, `9.7.0` has been added in this release.

Here is a sample YAML file to try out the latest version.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.7.0
  zookeeperRef:
    name: zk-com
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        storageClassName: standard
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 2
      storage:
        storageClassName: standard
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      replicas: 2
      storage:
        storageClassName: standard
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
