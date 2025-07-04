---
title: Announcing KubeDB v2025.6.30
date: "2025-07-02"
weight: 16
authors:
- Arnob Kumar Saha
tags:
- archiver
- backup
- cassandra
- catalog-manager
- clickhouse
- cloud-native
- dashboard
- database
- distributed
- dns
- druid
- grafana
- hazelcast
- ignite
- kafka
- kubedb
- kubernetes
- kubestash
- mongodb
- mssqlserver
- network
- percona-xtradb
- postgres
- prometheus
- recommendation
- redis
- restore
- s3
- security
- tls
- voyager-gateway
- zookeeper
---

KubeDB **v2025.6.30** brings a bunch of new features, enhanced security with TLS support, and expanded operational capabilities for managing databases in Kubernetes. This release focuses on improving scalability, security, and automation for production-grade database deployments.

## Key Changes
- **TLS Support**: Added TLS/SSL support for **Cassandra**, **Ignite**, and **ClickHouse** using cert-manager.
- **New OpsRequests**: Introduced advanced operational requests for **Cassandra**, **Hazelcast**, **Ignite** & **MariaDB** to support scaling, reconfiguration, and more.
- **Distributed Availability Groups (DAGs)**: New support for **Microsoft SQL Server** DAGs for cross-cluster disaster recovery.
- **Druid Recommendations**: Added recommendation engine for **Druid** to manage version updates, TLS certificate rotation, and authentication secret rotation.
- **Kafka Enhancements**: Introduced Kafka version 4.0.0 with a dedicated init container for provisioning and configuration.
- **Redis Hostname Support**: Added hostname configuration for Redis clusters starting from version 7.


## Cassandra
This release introduces **TLS support** for Cassandra, along with new **OpsRequests** for **Horizontal Scaling**, **Volume Expansion**, **Reconfigure**, and **Reconfigure TLS**.

### TLS
Cassandra now supports TLS/SSL via cert-manager. Ensure cert-manager is installed in your cluster ([cert-manager installation guide](https://cert-manager.io/docs/releases/)). Certificates are issued using `Issuer`/`ClusterIssuer` and `Certificate` CRDs.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cass
  namespace: demo
spec:
  version: 5.0.3
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: cassandra-ca-issuer
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

### Horizontal Scaling
Scale Cassandra nodes horizontally using the following OpsRequest:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cassandra-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: cass
  horizontalScaling:
    node: 4
```

### Volume Expansion
Expand storage for Cassandra nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cas-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: cass
  volumeExpansion:
    node: 2Gi
    mode: Online
```

### Reconfigure
Update Cassandra configuration with a custom secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cass-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: cass
  configuration:
    configSecret:
      name: cass-custom-config
  timeout: 5m
  apply: IfReady
```

### Reconfigure TLS
Update TLS configuration for Cassandra:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cas-reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cass
  tls:
    issuerRef:
      name: cassandra-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - cassandra
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

## ClickHouse
### TLS/SSL Support
ClickHouse now supports TLS/SSL via cert-manager ([cert-manager installation guide](https://cert-manager.io/docs/releases/)):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch
  namespace: demo
spec:
  version: 24.4.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  sslVerificationMode: relaxed
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: ch-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  deletionPolicy: WipeOut
```

## Druid
### Recommendation Engine
KubeDB now supports recommendations for Druid, including **Version Update**, **TLS Certificate Rotation**, and **Authentication Secret Rotation**. Recommendations are generated if `.spec.authSecret.rotateAfter` is set, based on:
- AuthSecret lifespan > 1 month with < 1 month remaining.
- AuthSecret lifespan < 1 month with < 1/3 lifespan remaining.

Example recommendation:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2025-07-01T06:31:29Z"
  labels:
    app.kubernetes.io/instance: druid
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-auth
  name: druid-x-druid-x-rotate-auth-mpyxnh
  namespace: demo
spec:
  backoffLimit: 5
  deadline: "2025-07-01T06:41:24Z"
  description: Recommending AuthSecret rotation,druid-auth AuthSecret needs to be rotated before 2025-07-01 06:51:24 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: DruidOpsRequest
    metadata:
      name: rotate-auth
      namespace: demo
    spec:
      databaseRef:
        name: druid
      type: RotateAuth
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Druid
    name: druid
status:
  approvalStatus: Approved
  approvedWindow:
    window: Immediate
  conditions:
  - lastTransitionTime: "2025-07-01T06:34:00Z"
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: "True"
    type: SuccessfullyCreatedOperation
  createdOperationRef:
    name: druid-1751351640-rotate-auth-auto
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: InProgress
  reason: StartedExecutingOperation
```

## FerretDB
### Dashboard
We have added summary dashboards for FerretDB in this release.

## Hazelcast
This release introduces multiple **OpsRequests** for Hazelcast, including **Restart**, **Horizontal Scaling**, **Vertical Scaling**, **Volume Expansion**, **Update Version**, and **Reconfigure**.

### Restart
Restart the Hazelcast cluster:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-restart
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: hazelcast-sample
  type: Restart
```

### Horizontal Scaling
Scale Hazelcast pods:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-scale-up
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-sample
  type: HorizontalScaling
  horizontalScaling:
    hazelcast: 4
```

### Vertical Scaling
Scale Hazelcast pod resources:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-vertical-scaling
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-sample
  type: VerticalScaling
  verticalScaling:
    hazelcast:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

### Volume Expansion
Expand Hazelcast storage:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-volume-expansion
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: hazelcast-sample
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    hazelcast: 5Gi
```

### Update Version
Upgrade Hazelcast version:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-version-upgrade
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-sample
  type: UpdateVersion
  updateVersion:
    targetVersion: 5.5.6
```

### Reconfigure
Reconfigure Hazelcast with a custom config secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: sl-reconfigure-custom-config
  namespace: demo
spec:
  apply: IfReady
  configuration:
    configSecret: 
      name: hazelcast-custom-config
    applyConfig:
      hazelcast.yaml: |-
        hazelcast:
          persistence:
            enabled: true
            validation-timeout-seconds: 2500
            data-load-timeout-seconds: 3000
            auto-remove-stale-data: false
  databaseRef:
    name: hazelcast-sample
  type: Reconfigure
```

## Ignite
This release adds **TLS support** for Ignite and introduces **Restart**, **Horizontal Scaling**, and **Vertical Scaling** OpsRequests.

### TLS
Ignite now supports TLS/SSL via cert-manager ([cert-manager installation guide](https://cert-manager.io/docs/releases/)):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-tls
  namespace: demo
spec:
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: ig-ca-issuer
      kind: Issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
      - alias: client
        subject:
          organizations:
            - kubedb
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  deletionPolicy: WipeOut
  replicas: 3
  version: 2.17.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
```

### Restart
Smart restart for Ignite cluster:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: ignite-tls
```

### Vertical Scaling
Scale Ignite pod resources:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: ignite-tls
  verticalScaling:
    ignite:
      resources:
        requests:
          memory: "3Gi"
          cpu: "3"
        limits:
          memory: "3Gi"
          cpu: "3"
  timeout: 5m
  apply: IfReady
```

### Horizontal Scaling
Scale Ignite replicas:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ignite-quickstart
  horizontalScaling:
    ignite: 5
```

## Kafka
Kafka version **4.0.0** is now supported, with an init container to handle provisioning and configuration, separating these tasks from the main container for improved modularity.

## MariaDB
This release introduces **Vertical Scaling**, **Horizontal Scaling**, and **Volume Expansion** OpsRequests for MariaDB **MaxScale** server, along with archiver backup support for MariaDB Replication topology.

### Vertical Scaling
Scale MaxScale server resources:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: sample-mariadb
  verticalScaling:
    maxscale:
      resources:
        requests:
          memory: "512Mi"
          cpu: "0.3"
        limits:
          memory: "512Mi"
          cpu: "0.3"
```

### Horizontal Scaling
Scale MaxScale server pods:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-mariadb
  horizontalScaling:
    maxscale: true
    member: 3
```

### Volume Expansion
Expand MaxScale storage:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-mariadb
  volumeExpansion:   
    mode: Offline
    maxscale: 100Mi
```

### Archiver fix
Correctly set the securityContext fields in the backupConfiguration & restoreSession on the archiver-enabled mode.


## Microsoft SQL Server
This release introduces **Distributed Availability Groups (DAGs)** for Microsoft SQL Server, enabling disaster recovery across Kubernetes clusters, data centers, or cloud regions.

### Distributed Availability Group (DAG)
DAGs link two Availability Groups (AGs) for robust disaster recovery. KubeDB automates setup using Raft for leader election and failover.

**Primary AG (ag1)**:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ag1
  namespace: demo
spec:
  version: "2022-cu16"
  replicas: 3
  topology:
    mode: DistributedAG
    availabilityGroup:
      databases:
        - agdb
      secondaryAccessMode: "All"
    distributedAG:
      self:
        role: Primary
        url: "10.2.0.236"
      remote:
        name: ag2
        url: "10.2.0.181"
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
      - name: mssql
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: Evaluation
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

**Secondary AG (ag2)**:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ag2
  namespace: demo
spec:
  version: "2022-cu16"
  replicas: 3
  topology:
    mode: DistributedAG
    availabilityGroup:
      secondaryAccessMode: "All"
      loginSecretName: ag1-dbm-login
      masterKeySecretName: ag1-master-key
      endpointCertSecretName: ag1-endpoint-cert
    distributedAG:
      self:
        role: Secondary
        url: "10.2.0.181"
      remote:
        name: ag1
        url: "10.2.0.236"
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
      - name: mssql
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: "Developer"
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

**Failover Steps**:
1. **Synchronize Data** (if primary is online):
   ```sql
   ALTER AVAILABILITY GROUP [dag]
   MODIFY AVAILABILITY GROUP ON
     'ag1' WITH (AVAILABILITY_MODE = SYNCHRONOUS_COMMIT),
     'ag2' WITH (AVAILABILITY_MODE = SYNCHRONOUS_COMMIT);
   ```
2. **Set Primary to Secondary**:
   ```sql
   ALTER AVAILABILITY GROUP [mydag] SET (ROLE = SECONDARY);
   ```
3. **Promote Secondary**:
   ```sql
   ALTER AVAILABILITY GROUP [mydag] FORCE_FAILOVER_ALLOW_DATA_LOSS;
   ```
4. Update `spec.topology.distributedAG.self.role` in both resources.

For detailed setup, see [Distributed Availability Group documentation](https://kubedb.com/docs/latest/guides/mssqlserver/clustering/dag_cluster/).

## MongoDB

### Sharded Cluster fix
If `ipv6 capability` is not enabled in the karnel level, some mongosh commands were failing for `mongos` node. This issue has been sorted out in this release.

### Archiver fix
Correctly set the securityContext fields in the backupConfiguration & restoreSession on the archiver-enabled mode.

### New Version support
A ton of new versions have been added, namely: `5.0.21`,`6.0.24`, `7.0.21`,`8.0.10`, `percona-5.0.29`, `percona-6.0.24`,`percona-7.0.18`, `percona-8.0.8`.

## PostgreSQL
### Horizontal Scaling
Users can now scale from an HA setup to a standalone PostgreSQL instance:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: <your-database-name>
  horizontalScaling:
    replicas: 1
```

> **Note**: Scaling to standalone is not allowed with synchronous replication. Consider replication slots if manually created.

### Archiver fix
Correctly set the securityContext fields in the backupConfiguration & restoreSession on the archiver-enabled mode.


## Redis
Redis now supports hostname configuration for clusters starting from version 7:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-demo
  namespace: demo
spec:
  version: 7.2.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
    announce:
      type: hostname
      shards:
        - endpoints:
          - "www.example.com:10050@10056"
          - "www.example.com:10051@10057"
        - endpoints:
          - "www.example.com:10052@10058"
          - "www.example.com:10053@10059"
        - endpoints:
          - "www.example.com:10054@10060"
          - "www.example.com:10055@10061"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 200M
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
  deletionPolicy: WipeOut
```

### Announce OpsRequest
Convert a Redis cluster from IP to hostname configuration:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: announce-ops
  namespace: demo
spec:
  type: Announce
  databaseRef:
    name: redis-demo
  announce:
    type: hostname
    shards:
      - endpoints:
        - "www.example.com:10050@10056"
        - "www.example.com:10051@10057"
      - endpoints:
        - "www.example.com:10052@10058"
        - "www.example.com:10053@10059"
      - endpoints:
        - "www.example.com:10054@10060"
        - "www.example.com:10055@10061"
```

> **Note**: This OpsRequest may flush existing data. Use backup and restore to prevent data loss.

To achieve this feature, there are multiple other operators needed. Specially the [voyager-gateway](https://github.com/voyagermesh/installer/tree/v2025.6.30/charts/voyager-gateway) & [catalog-manager](https://github.com/appscode-cloud/installer/tree/release-v2025.6.30/charts/catalog-manager). Catalog-manager will act as the co-ordinator here among these various components. You can find detailed documentation [here](https://kubedb.com/docs/latest/guides/redis/external-connections/initialization/).

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).