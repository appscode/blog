---
title: Announcing KubeDB v2025.10.27
date: "2025-10-27"
weight: 15
authors:
- Arnob Kumar Saha
tags:
- autoscaler
- backup
- clickhouse
- cloud-native
- database
- distributed
- kafka
- kubedb
- kubernetes
- mariadb
- postgres
- recommendation
- redis
- restore
- security
- tls
- valkey
---

KubeDB **v2025.10.27** introduces enhancements like rack awareness for Kafka, distributed autoscaling and advanced backup/restore for MariaDB, health check improvements for Postgres, ACL support for Redis/Valkey, and autoscaling with recommendations for ClickHouse. This release focuses on improving fault tolerance, security, scalability, and recovery capabilities for databases in Kubernetes.

## Key Changes
- **Rack Awareness for Kafka**: Added support for rack-aware replica placement to enhance fault tolerance.
- **Distributed MariaDB Enhancements**: Introduced autoscaling support and KubeStash (Stash 2.0) backup/restore, with Restic driver for continuous archiving and new replication strategies for PITR.
- **Postgres Improvements**: Updated health checks to avoid unnecessary LSN advancement and fixed standby join issues.
- **Redis/Valkey ACL**: Added Access Control List (ACL) for fine-grained user permissions, plus new Redis version 8.2.2.
- **ClickHouse Features**: Introduced autoscaling for compute and storage, along with recommendation engine for version updates, TLS, and auth rotations.

## ClickHouse

This release introduces Recommendations and the AutoScaling feature for ClickHouse.

### AutoScaling

This release introduces the ClickHouseAutoscaler — a Kubernetes Custom Resource Definition (CRD) — that enables automatic compute (CPU/memory) and storage autoscaling for ClickHouse. Here’s a sample manifest to deploy ClickHouseAutoscaler for a KubeDB-managed ClickHouse:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ClickHouseAutoscaler
metadata:
  name: ch-compute-autoscale
  namespace: demo
spec:
  databaseRef:
    name: clickhouse-prod
  compute:
    clickhouse:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

### Recommendation Engine

KubeDB now supports recommendations for ClickHouse, including Version Update, TLS Certificate Rotation, and Authentication Secret Rotation. Recommendations are generated if .spec.authSecret.rotateAfter is set, based on:

AuthSecret lifespan > 1 month with < 1 month remaining. AuthSecret lifespan < 1 month with < 1/3 lifespan remaining. Example recommendation:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 24.4.1
  creationTimestamp: "2025-10-20T09:40:21Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: ch
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
    kubedb.com/version-update-recommendation-type: major-minor
  name: ch-x-clickhouse-x-update-version-lr21eg
  namespace: demo
  resourceVersion: "192088"
  uid: 286152eb-ba5c-45d8-bc54-6ef0a7255362
spec:
  backoffLimit: 10
  description: Latest Major/Minor version is available. Recommending version Update
    from 24.4.1 to 25.7.1.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ClickHouseOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: ch
      type: UpdateVersion
      updateVersion:
        targetVersion: 25.7.1
    status: {}
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    name: ch
  vulnerabilityReport:
    message: no matches for kind "ImageScanReport" in version "scanner.appscode.com/v1alpha1"
    status: Failure
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```

## Kafka

This release introduces rack awareness support for Kafka. A new `brokerRack` field has been added to the Kafka CRD to enable rack-aware replica placement using a specified Kubernetes topology key, such as `topology.kubernetes.io/zone`. When enabled, a default `replica.selector.class.name` configuration is automatically applied to distribute replicas across different racks or zones for improved fault tolerance and high availability.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 4.0.0
  brokerRack:
    topologyKey: topology.kubernetes.io/zone
  ……
  …..
  …
```

## MariaDB

### Distributed MariaDB Autoscaler
In this release, we have introduced support for the Distributed MariaDB Autoscaler and KubeStash (also known as Stash 2.0) backup and restore functionalities.

To enable autoscaling, the metrics server and monitoring agent (Prometheus) must be installed on all clusters where Distributed MariaDB pods are running. You can provide the monitoring agent details via the PlacementPolicy CR.

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb
spec:
  clusterSpreadConstraint:
    distributionRules:
    - clusterName: demo-controller
      monitoring:
        prometheus:
          url: http://prometheus-operated.monitoring.svc.cluster.local:9090
      replicaIndices:
      - 2
    - clusterName: demo-worker
      monitoring:
        prometheus:
          url: http://prometheus-operated.monitoring.svc.cluster.local:9090
      replicaIndices:
      - 0
      - 1
      - 3
    slice:
      projectNamespace: kubeslice-demo-distributed-mariadb
      sliceName: demo-slice
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

We have implemented several enhancements to bolster continuous archiving and point-in-time recovery for MariaDB in KubeDB. Here's an overview of the key updates:

### Restic Driver for Base Backup Support

We now offer support for the Restic driver in MariaDB continuous archiving and recovery operations. Previously, only the VolumeSnapshotter driver was supported.

To utilize the Restic driver, configure the MariaDBArchiver Custom Resource (CR) by setting `.spec.fullBackup.Driver` to "Restic".

### Replication Strategies for MariaDB Archiver Restore

We have introduced a new replication strategy feature that supports two distinct methods for restoring MariaDB replicas. The available strategies are outlined below:

***none***: Each MariaDB replica restores the base backup and binlog files independently. Once the restore process is complete, the replicas join the cluster individually.

***sync***: The base backup and binlog files are restored solely on pod-0. The other replicas then synchronize their data using MariaDB’s native replication mechanism from pod-0.

Two more methods will be added in upcoming releases. Below is a sample YAML configuration for setting up a MariaDBArchiver in KubeDB:

Note: You must set RunAsUser to the database user ID (999) in the JobTemplate.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MariaDBArchiver
metadata:
  name: mariadbarchiver-sample
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
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 999
          runAsGroup: 0
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

Here’s a sample YAML configuration for restoring MariaDB from a backup using the new features:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: restore-mariadb
  namespace: demo
spec:
  init:
    archiver:
      replicationStrategy: sync
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: md-full
        namespace: demo
      recoveryTimestamp: "2026-10-01T06:33:02Z"
  version: "11.6.2"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## Postgres

### Health Check

The way we do the write check for postgres databases has been changed. Previously we used to create a `kubedb_write_check table` for checking if we can write in the database in order to mark the database as healthy. From now on, we will try to run `BEGIN READ WRITE; ROLLBACK;`.

As a result of doing this, LSN will not be advanced unnecessarily. Note this might throw `HighRollBackAlert` in your `postgres` database in case you had set a lower threshold.

### Bug fix

We have fixed a bug which was not letting a standby join with primary saying `grpc call to pg_controldata failed` with some error.

## Redis/Valkey

### ACL (Access Control List)

In this Release, we have added an Access Control List (ACL) for Redis.

Initially, you can deploy Redis/Valkey with user-associated password and rules.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: vk
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  deletionPolicy: WipeOut
  acl:
    secretRef:
      name: acl-secret
    rules:
      - app1 ${k1} allkeys +@string +@set -SADD
      - app2 ${k2} allkeys +@string +@set -SADD
```

Deploy the secret associated with `acl.rules`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: acl-secret
  namespace: demo
type: Opaque
stringData:
  k1: "pass1"
  k2: "pass2"
```

You can check the `acl list` now.

To add/update/delete ACL, you can use RedisOpsRequest. As an example:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: vk
  configuration:
    auth:
      syncACL:
        - app1 ${k1} +get ~mykeys:*
        - app10 ${k10} +get ~mykeys:*
      deleteUsers:
        - app2
      secretRef:
        name: <new/old secret name which have all the key list>
```

### New version support
Redis version `8.2.2` is now available in KubeDB.

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).