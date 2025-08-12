---
title: Announcing KubeDB v2025.7.31
date: "2025-08-12"
weight: 15
authors:
- Arnob Kumar Saha
tags:
- autoscaler
- backup
- cassandra
- clickhouse
- cloud-native
- database
- distributed
- druid
- hazelcast
- ignite
- kubedb
- kubernetes
- mariadb
- mongodb
- mssqlserver
- mysql
- postgres
- recommendation
- redis
- restore
- security
- tls
---

KubeDB **v2025.7.31** introduces new features like autoscaling, TLS reconfiguration, authentication rotation, distributed database support, and version updates across various databases. This release enhances operational efficiency, security, and scalability for managing databases in Kubernetes environments.

## Key Changes
- **Autoscaling Support**: Added autoscaling for **Cassandra** and **Hazelcast** clusters.
- **TLS and Authentication Enhancements**: Introduced ReconfigureTLS and RotateAuth OpsRequests for **Ignite**, **Cassandra** etc databases.
- **Distributed Database Support**: New distributed deployment capabilities for **MariaDB** and **Postgres** across multiple clusters using OCM and KubeSlice.
- **Initialization and Configuration**: Added script-based initialization for **Microsoft SQL Server** and custom configuration for **MariaDB MaxScale**.
- **Version Updates and Recommendations**: Support for new versions in **ClickHouse** and same-version update recommendations in the engine.

## Cassandra
This release introduces `RotateAuth` OpsRequest and the AutoScaling feature for Cassandra.

### RotateAuth
To modify the credential in Cassandra, you can use RotateAuth. This IgniteOpsRequest will update the credential.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: cassandra-prod
  timeout: 5m
  apply: IfReady
```

### Autoscaler
```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: CassandraAutoscaler
metadata:
  name: cassandra-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: cassandra-prod
  compute:
    cassandra:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.2Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: CassandraAutoscaler
metadata:
  name: cassandra-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: cassandra-autoscale
  storage:
    cassandra:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 100
      expansionMode: "Offline"
```

## ClickHouse
Add supports for various opsRequest in this release.

### Horizontal Scaling
Scale ClickHouse nodes horizontally using the following OpsRequest:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-scale-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ch
  horizontalScaling:
    cluster:
      - clusterName: appscode-cluster
        replicas: 3
```

### Volume Expansion
Expand Storage for ClickHouse nodes:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-offline-volume-expansion
  namespace: demo
spec:
  apply: "IfReady"
  type: VolumeExpansion
  databaseRef:
    name: ch
  volumeExpansion:
    mode: "Offline"
    node: 2Gi
```

### Reconfigure
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ch
  configuration:
    configSecret:
      name: ch-custom-config
```

### Version Update
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: ch
  updateVersion:
    targetVersion: 25.7.1
```

### Version Support
Support Added for ClickHouse version 25.7.1

## Hazelcast
This release introduces the HazelcastAutoscaler — a Kubernetes Custom Resource Definition (CRD) — that enables automatic compute (CPU/memory) and storage autoscaling for Hazelcast clusters.
Here’s a sample manifest to deploy HazelcastAutoscaler for a KubeDB-managed Hazelcast cluster:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: HazelcastAutoscaler
metadata:
  name: hz-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-sample
  opsRequestOptions:
      timeout: 5m
      apply: IfReady
  compute:
    hazelcast:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 1
      minAllowed:
        cpu: 550m
        memory: 1.6Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
  storage:
    hazelcast:
      trigger: "On"
      expansionMode: "Online"
      usageThreshold: 1
      scalingThreshold: 50
```

## Ignite
This release introduces ReconfigureTLS, RotateAuth, VersionUpdate OpsRequests for Ignite.
### ReconfigureTLS
To configure TLS in Ignite using a IgniteOpsRequest, we have introduced ReconfigureTLS. You can enable TLS in Ignite by simply deploying a YAML configuration:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ignite-quickstart
  tls:
    issuerRef:
      name: ig-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
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
  timeout: 5m
  apply: IfReady
```


### RotateAuth
To modify the credential in Ignite you can use RotateAuth. This IgniteOpsRequest will update the credential.


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-auth-rotate
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: ignite-quickstart
```

### Version Update
In order to update the version of Ignite, we have to create a IgniteOpsRequest CR with the desired version that is supported by KubeDB.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-update-version
  namespace: demo
spec:
  databaseRef:
    name: ignite-quickstart
  type: UpdateVersion
  updateVersion:
    targetVersion: 2.17.0
```

## MariaDB
We are thrilled to announce the addition of distributed MariaDB support in this release. You can now deploy MariaDB across multiple clusters with ease. We leverage OCM to streamline deployment management across clusters and utilize KubeSlice to ensure seamless network connectivity among pods on different clusters.
You must create a `PodPlacementPolicy` to specify which pods are scheduled on particular clusters and reference it in the database YAML using spec.podTemplate.spec.podPlacementPolicy.name. Additionally, set the spec.distributed field to true.
This release also includes `custom configuration support` for MariaDB MaxScale server. Users can now provide a custom configuration file via config-secret. This custom configuration will overwrite the existing default one.

PlacementPolicy yaml:
```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb
spec:
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  ocm:
    distributionRules:
      - clusterName: kubeslice-worker-1
        replicas:
          - 0
          - 2
      - clusterName: kubeslice-worker-2
        replicas:
          - 1
    sliceName: demo-slice
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

Distributed MariaDB sample yaml:
```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      configSecret:
        name: maxscale-configuration
      enableUI: true
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-mariadb
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
For more details see the [KubeDB docs](https://kubedb.com/docs/v2025.7.31/guides/mariadb/distributed/overview/)


### MaxScale Server Custom Config yamls:
Lets create a secret that contains the custom MaxScale configuration
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: maxscale-configuration
  namespace: demo
type: Opaque
stringData:
  mx-config.cnf: |
    [maxscale]
    threads=auto
    log_info=true
```

Now, create a MariaDB CR specifying spec.topology.maxscale.configSecret field.
MariaDB yaml:
```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      configSecret:
        name: maxscale-configuration
      enableUI: true
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

## Microsoft SQL Server
This release introduces a new feature for database initialization using scripts for both standalone and Availability Group deployments. This allows users to bootstrap their SQL Server instances with predefined schemas, data, users, etc., declaratively at creation time.
### Initialization Using Script
You can now initialize your MSSQLServer instances using .sql, .sh, or .sql.gz files provided via a Kubernetes volume source, such as a ConfigMap volume. KubeDB will automatically execute these scripts. This is ideal for setting up initial databases, creating tables, seeding data, or configuring users and permissions.
The scripts are specified in the spec.init.script field of the MSSQLServer CR. Scripts within the specified source will be executed alphabetically.

Example: `Initializing a Standalone SQL Server`
1) Create a ConfigMap containing your initialization SQL script:
```bash
kubectl create configmap -n demo mssql-init-scripts --from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mssqlserver-init-scripts/raw/master/init.sql)"
```

2) Reference this ConfigMap in your MSSQLServer manifest:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-init
  namespace: demo
spec:
  # ... other fields like version, storage, tls, podTemplate ...
  init:
    script:
      configMap:
        name: mssql-init-scripts
```
This feature is also fully compatible with Availability Group deployments. When used with an AG, the initialization script is executed on the primary replica, and the created databases and data are then automatically replicated to all secondary replicas.
For a detailed guide on using this feature, please check out our [documentation](https://kubedb.com/docs/v2025.7.31/guides/mssqlserver/initialization/).

## MySQL
This release introduces RotateAuth OpsRequest for MySQL database. If a user wants to update the authentication credentials for a particular database, they can create an OpsRequest of type RotateAuth with or without referencing an authentication secret.
Rotate Authentication Without Referencing a Secret
If the secret is not referenced, the ops-manager operator will create new credentials and update the existing secret with the new credentials, keeping previous credentials under the keys username.prev and password.prev.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-mysql
```

### Rotate Authentication With a Referenced Secret
If a secret is referenced, the operator will update the .spec.authSecret.name field with the new secret name. Archives the old credentials in the newly created secret under the keys username.prev and password.prev. However, username change is not supported. So previous username must be used in the referenced secret.
New Secret Example:

```yaml
apiVersion: v1
data:
  password: bXlQYXNzd29yZA==
  username: cm9vdA==
kind: Secret
metadata:
  name: my-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: rotate-auth-external
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-mysql
  authentication:
    secretRef:
      name: my-auth
```
Note: In Remote Replica mode if the source authentication is changed a separate RotateAuth OpsRequest with the updated credentials must be applied.

## Postgres

We are thrilled to announce the addition of distributed Postgres support in this release. You can now deploy Postgres across multiple clusters with ease. We leverage OCM to streamline deployment management across clusters and utilize KubeSlice to ensure seamless network connectivity among pods on different clusters.
You must create a PodPlacementPolicy to specify which pods are scheduled on particular clusters and reference it in the database YAML using spec.podTemplate.spec.podPlacementPolicy.name. Additionally, set the spec.distributed field to true.
Sample yamls

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-postgres
spec:
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  ocm:
    distributionRules:
      - clusterName: kubeslice-worker-1
        replicas:
          - 0
          - 2
          - 4
      - clusterName: kubeslice-worker-2
        replicas:
          - 1
          - 3
          - 5
    sliceName: demo-slice
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: monitor-postgres
  namespace: demo
spec:
  clientAuthMode: md5
  deletionPolicy: WipeOut
  distributed: true
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  leaderElection:
    electionTick: 10
    heartbeatTick: 1
    maximumLagBeforeFailover: 67108864
    period: 300ms
    transferLeadershipInterval: 1s
    transferLeadershipTimeout: 1m0s
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  podTemplate:
    spec:
      containers:
      - name: postgres
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
      podPlacementPolicy:
        name: distributed-postgres
  replicas: 3
  sslMode: disable
  standbyMode: Hot
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  version: "16.8"
```

## Recommendation Engine
### Same-Version-Update Recommendation
A Same-Version Update Recommendation is created when the database version image looks the same as before, but there is a difference in either the image tag or the image digest (the unique ID of the image content).
Same-Version-Update recommendations now appear alongside traditional version updates with Deadline.
Sample YAML:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 28.0.1
  creationTimestamp: "2025-08-04T10:46:29Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: dr
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
  name: dr-x-druid-x-update-same-version-qpc8iz
  namespace: demo
  resourceVersion: "1986825"
  uid: eff538e5-fb4b-4dcd-a617-bd929974384c
spec:
  backoffLimit: 10
  deadline: "2025-08-04T10:47:29Z"
  description: |-
    Same version update is required. Update details:
    1: index.docker.io/frsarker/druid:latest@sha256:3471ff921519207c82390145d03ca72c4703e4f9e94735acc870ff09f627c656 => sha256:14461a562cc3e5b7f788a0165121ff21cda0de361bd8fba94e4fb30b6ca599db
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: DruidOpsRequest
    metadata:
      name: update-same-version
      namespace: demo
    spec:
      databaseRef:
        name: dr
      type: UpdateVersion
      updateVersion:
        targetVersion: 28.0.1
    status: {}
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Druid
    name: dr
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

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).