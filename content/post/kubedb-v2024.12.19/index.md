---
title: Announcing KubeDB v2024.12.19
date: "2024-12-19"
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
- redis
- restore
- s3
- security
- singlestore
- solr
- tls
- zookeeper
---

KubeDB **v2024.12.19** is now available! This latest release brings significant performance enhancements, improved reliability, and new features to database management experience on Kubernetes. Here are some of the key features to mention
- **OpsRequest Support**: New `OpsRequest` features have been added for `Memcached`, `MySQL`, offering greater flexibility for managing database administrative tasks. Moreover, a new `OpsRequest` feature named `ReplicationModeTransformation` has been introduced in this release.
- **Recommendation Engine**: Recommendation support for `KubeDB` managed kafka has been added.
- **New Version Support**: New versions have been added for `Elasticsearch`, `MySQL`, `Redis`, `Solr`, `Singlestore`.
- **Archiver**: Archiver and point-in-time-recovery within `KubeDB` has been enhanced for `MongoDBArchiver`, `MariaDBArchiver`, `MySQLArchiver`, `PostgressArchiver`, `MSSQLArchiver`.

## Archiver
Archiver support has been enhanced for `MongoDBArchiver`, `MariaDBArchiver`, `MySQLArchiver`, `PostgressArchiver` and `MSSQLArchiver`. We have replaced the field `spec.walBackup` in spec with `spec.logBackup` with some enhancement. Besides two existing field `RuntimeSettings` and `ConfigSecret`, two more fields have been added and those are:

**SuccessfulLogHistoryLimit**: `SuccessfulLogHistoryLimit` defines the number of successful Logs backup status that the incremental snapshot will retain. It's default value is 5.

**FailedLogHistoryLimit**: FailedLogHistoryLimit defines the number of failed Logs backup that the incremental snapshot will retain for debugging purposes. It's default value is 5.

You can find full spec [here](https://github.com/kubedb/apimachinery/blob/master/apis/archiver/v1alpha1/types.go#L74C1-L92C2)

So Incremental snapshots status field will store successful and failed log history according to these extra two fields. This will make easier our debugging process.

Here is a sample yaml for `MongoDBArchiver`. Changes in the field `.spec.logBackup` will be same for other archivers as well.
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

## Druid

### New version
We have added support for `Druid` versions `31.0.0`.
Here is yaml for druid version `31.0.0`

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

### New Versions
We have added support for `Elasticsearch` versions `xpack-7.17.25`, `xpack-8.15.4`, `xpack-8.16.0` and `opensearch-1.3.19`.
Here is yaml for elasticsearch version `xpack-8.16.0`

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

### New Versions
We have added support for `Elasticsearch` version `opensearch-1.3.19`.
Here is yaml for elasticsearch version `opensearch-1.3.19`

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
  version: opensearch-1.3.19
```

### Bug Fix
We have fixed `RotateAuth` OpsRequest issue for `Opensearch` in this release.
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

If the secret is referenced, the operator will update the `.spec.authSecret.name` with the new secret name. Here is the Yaml:
New Secret:
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

## FerretDB

### New Versions
We have added support for `FerretDB` version `1.24.0`.
Here is a simple yaml for `FerretDB` version `1.24.0`

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


## Kafka

### New Versions
We have added support for `Kafka` versions `3.7.2`, `3.8.1`and `3.9.0`.
Here is a simple yaml for `Kafka` version `3.9.0`

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

## Memcached

### Memcached OpsRequest

`RotateAuth` OpsRequest for `Memcached` has been added. If a user wants to update the authentication credentials for a particular database, they can create an `OpsRequest` of type `RotateAuth` with or without referencing an authentication secret.
If the secret is not referenced, the ops-manager operator will create a new credential and update the current secret. Here is the Yaml for rotating authentication credentials for a Memcached cluster using MemcachedOpsRequest.

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

If the secret is referenced, the operator will update the .spec.authSecret.name with the new secret name. Here is the Yaml:
New Secret:

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
We have added a field `.spec.authSecret.activeFrom` to the db yaml which refers to the timestamp of the credential is active from. We also add an annotation kubedb.com/auth-active-from in currently using auth secret which refer to the active from time of this secret.



## MySQL
In this release, we added support for new `MySQL` version, improved `MySQL` continuous archiving and `PITR` within `KubeDB`, and `MySQL` replication mode change(remote replica to group replication) ops-request.

### MySQL New Version
Support for MySQL version `9.0.1` has been added in the new release.

### MySQL Archiver
We now support the `Restic` driver for MySQL continuous archiving and recovery. Previously, only the `VolumeSnapshotter` driver was available.
To use the `Restic` driver, configure the `MySQLArchiver` Custom Resource (CR) by setting `.spec.fullBackup.Driver` to "Restic"

***Restic Driver for Base Backup Support***:
We now support the `Restic` driver for MySQL continuous archiving and recovery. Previously, only the `VolumeSnapshotter` driver was available.
To use the Restic driver, configure the `MySQLArchiver` Custom Resource (CR) by setting .spec.fullBackup.Driver to "Restic"

***Replication Strategies for MySQL Archiver Restore***:
A new replication strategy feature has been introduced, supporting four distinct approaches to restoring MySQL replicas. The available strategies are as follows:

- ***none***: Each `MySQL` replica independently restores the base backup and binlog files. After completing the restore process, the replicas individually join the replication group.
- ***sync***: The base backup and binlog files are restored exclusively on pod-0. Other replicas then synchronize their data by leveraging the MySQL clone plugin to replicate from pod-0.
- ***fscopy***: The base backup and binlog files are restored on pod-0. The data is then copied from pod-0's data directory to the data directories of other replicas using file system copy. Once the data transfer is complete, the group replication process begins.
- ***clone***: Each `MySQL` replica independently restores the base backup and binlog files. After completing the restore process, the replicas individually join the replication group. The clone strategy works for only VolumeSnapshotter driver.

Please note that `fscopy` does not support cross-zone operations.

Below is a sample YAML configuration for setting up a `MySQLArchiver` in KubeDB:
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

Here’s a sample YAML configuration for restoring MySQL from a backup using the new features:
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

### MySQL OpsRequest
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

## Redis

### New Versions
This release adds support for redis version `6.2.16`, `7.2.6` and `7.4.1`. Here is the sample YAML file of the Redis cluster using the 7.4.1 version

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

### New Versions
In this release, we have added support for `SingleStore` versions `8.7.21` and `8.9.3`. Here is the sample YAML file of the SingleStore cluster using the 8.9.3 version,

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

### New Versions
We have added new `Solr` versions `8.11.4`, `9.7.0` within `KubeDB`.
Here is a yaml of solr cluster for version `9.7.0`.

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

## Recommendation Engine

In this Release, we are introducing `Recommendation` support for KubeDB managed Kafka instances. KubeDB ops-manager generates three types of recommendations for Kafka - Version Update Recommendation, TLS Certificates Rotation Recommendation,n and Authentication Secret Rotation Recommendation.
