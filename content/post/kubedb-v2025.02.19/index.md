---
title: Announcing KubeDB v2025.02.19
date: "2025-02-20"
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

KubeDB **v2025.02.19** is now available! This latest release brings significant performance enhancements, improved reliability, and new features to database management experience on Kubernetes. 
- **OpsRequest Support**: New `OpsRequest` support have been added for  `Pgbouncer`, `Pgpool` and `Postgres`, offering greater flexibility for managing database administrative tasks.
- **New Version Support**: New versions have been added for `Pgbouncer` and `PerconaXtraDB`.

## MongoDB

In this release we fixed the permission issue of Point in Time Recovery with MongoDBArchiver for Shard cluster.

## PerconaXtraDB

### New Version
In this release new version 8.0.40 and 8.4.3 added.

## Pgbouncer

### SecurityContext

In this release we fixed the security-context issue. You can deploy **pgbouncer** in kubedb using this yaml:
```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  generation: 3
  name: pb
  namespace: demo
spec:
  database:
    databaseName: postgres
    databaseRef:
      name: ha-postgres
      namespace: demo
  podTemplate:
    spec:
      containers:
      - name: pgbouncer
        securityContext:
          runAsGroup: 0
          runAsNonRoot: true
          runAsUser: 1000670000
          seccompProfile:
            type: RuntimeDefault
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 1000670000
      serviceAccountName: pb
  replicas: 3
  sslMode: disable
  version: 1.23.1
```
### ReconfigureTLS

To configure TLS with an opsrequest in **pgbouncer** we have added **ReconfigureTLS**. To add TLS in **pgbouncer** you can simply deploy a yaml.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pb
  tls:
    sslMode: verify-full
    clientAuthMode: md5
    issuerRef:
      name: pb-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - pgbouncer
          organizationalUnits:
            - client
  apply: Always
```

### RotateAuth
To modify the `admin_user` in **pgbouncer** you can use **RotateAuth**. This **OpsRequest** will update the admin user name or password.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: change-auth-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: pb
  authentication:
    secretRef:
      name: new-authsecret
```


### New version

**PgBouncer** version 1.24.0 is now available in kubedb. To deploy **PgBouncer** with version 1.24.0 you can simply use this yaml:


```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb
  namespace: demo
spec:
  version: "1.24.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
```

Or, if you have a pgbouncer instance running, you can use UpdateVersion opsrequest to change the version.


## Pgpool
In this Release we have added a **PgpoolOpsRequest**

### RotateAuth
To update the pcp user in **pgpool** you can use **RotateAuth**. This Opsrequest will update the pcp user name or password.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: change-auth-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: pp
  authentication:
    secretRef:
      name: new-authsecret
```

### bug-fix
- Fixes RemoveCustomConfig and configuration merging order.

### Feature Improvements 
- Introduced Pgpool reload instead of pod restart while performing reconfiguration.


## Postgres

## PostgresOpsRequest
In this Release we have added 3 new **PostgresOpsRequest**s

### ReconnectStandby

If your database is in **Critical** Phase, then applying this OpsRequest will bring your database in **Ready** state. It will try to make your database ready by following steps:
- Try Restarting the standby databases.
- If not ready, do some internal processing and take base backup from primary.
- Restart again so that standby can join with the primary.

A sample yaml for `ReconnectStandby` `PostgresOpsRequest` is Given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: reconnect-standby
  namespace: demo
spec:
  apply: Always
  type: ReconnectStandby
  databaseRef:
    name: ha-postgres

```

Here you just need to provide the `.spec.type: ReconnectStandby` and the database reference. Also you need to set `.spec.apply` field to **Always**.

> Note: As you can see we take base backup from primary in order to make the standby replica ready, we can’t apply this OpsRequest when database is in **Not Ready** state. Also If your database size is huge, let's say more than 50GB, we suggest taking either `dump-based` or `file system` level backup before applying this.
 
### ForceFailOver

For kubedb managed **Postgres**, We try to guarantee no data loss. So, if a scenario arise, where a primary replica got disconnected (maybe node crashed or pod unschedule able) from the cluster and it has the data that standby replicas don’t have, we will not do that failover automatically as this will clearly result in data loss. However from your end if you value the availability most and you can incur few data loss, then you can run a **ForceFailOver** **PostgresOpsRequest** to promote one of the standby’s as primary.

A sample yaml for this `PostgresOpsRequest` is given below considering you had 3 replicas and replica 0 was the primary:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: force-failover
  namespace: demo
spec:
  apply: Always
  type: ForceFailOver
  databaseRef:
    name: ha-postgres
  forceFailOver:
    candidates:
      - ha-postgres-1
      - ha-postgres-2
```

Here `.spec.apply` field has to be **Always**, `.spec.type` to **ForceFailOver** and `.spec.forceFailOver.candidates` will be the list of primary candidates. If you provide more than one candidate, then we will choose the best candidate and make it primary.

### SetRaftKeyPair
We use **Raft Consensus Algorithm** implemented by [etcd.io](https://github.com/etcd-io/raft) which has a Key Value Store. Using **SetRaftKeyPair**, we can set/unset any Key-Value pair inside the Store.

A sample yaml for **SetRaftKeyPair** **PostgresOpsRequest** is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: raft-key-pair
  namespace: demo
spec:
  apply: Always
  type: RaftKeyPair
  databaseRef:
    name: ha-postgres
  setRaftKeyPair:
    keyPair:
      key: "value"

```

## Bug fix 
- WAL accumulating on the standby instance has been fixed.
## Improvements 
- Don’t allow failover if previous primary is already running


## Solr

Internal zookeeper has been configured for solr. Now, we don’t need to mention zookeeper reference to deploy solr.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  solrModules:
  - s3-repository
  - gcs-repository
  - prometheus-exporter
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
