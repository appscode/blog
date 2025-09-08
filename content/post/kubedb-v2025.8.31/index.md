---
title: Announcing KubeDB v2025.8.31
date: "2025-09-08"
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
- gitsync
- hazelcast
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- pgbouncer
- pgpool
- postgres
- recommendation
- redis
- restore
- security
- tls
---

KubeDB **v2025.8.31** introduces new features like storage migration, TLS reconfiguration, authentication rotation, Git-Sync initialization support, and version updates across various databases. This release enhances operational efficiency, security, and flexibility for managing databases in Kubernetes environments.

## Key Changes
- **Storage Migration Support**: Added storage migration for **Postgres** and **MySQL** to change StorageClass seamlessly.
- **TLS and Authentication Enhancements**: Introduced ReconfigureTLS and RotateAuth OpsRequests for **ClickHouse** and **Hazelcast**.
- **Git-Sync Initialization**: New support for initializing databases from public/private Git repositories in **MariaDB**, **Redis**, **Pgpool**, and **Pgbouncer**.
- **Recommendation Engine**: Added recommendations for **Pgpool**, **Hazelcast**, and **Pgbouncer**, including version updates, TLS rotation, and auth secret rotation.
- **Version Updates**: New versions for **Postgres** (17.6, 16.10, 15.14, 14.19, 13.22).
- **Distributed Database Enhancements**: OpsRequest support for TLS reconfiguration, version upgrades, volume expansion, and auth rotation in distributed **MariaDB** Galera Cluster mode, with fixed full cluster disaster recovery.

## ClickHouse

This release introduces ReconfigureTLS, RotateAuth OpsRequests for ClickHouse

### ReconfigureTLS

To configure TLS in Ignite using a ClickHouseOpsRequest, we have introduced ReconfigureTLS. You can enable TLS in ClickHouse by simply deploying a YAML configuration:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    sslVerificationMode: "relaxed"
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: clickhouse-ca-issuer
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
  apply: IfReady
```

### RotateAuth
To modify the credentials in ClickHouse, you can use RotateAuth. This ClickHouseOpsRequest will update the credential.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: clickhouse-prod
  timeout: 5m
  apply: IfReady
```

## Hazelcast
### Recommendation Engine
KubeDB now supports recommendations for Hazelcast, including Version Update, TLS Certificate Rotation, and Authentication Secret Rotation. Recommendations are generated if .spec.authSecret.rotateAfter is set, based on:

AuthSecret lifespan > 1 month with < 1 month remaining.
AuthSecret lifespan < 1 month with < 1/3 lifespan remaining.
Example recommendation:
```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 5.5.2
  creationTimestamp: "2025-09-01T11:07:26Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: hazelcast
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
  name: hazelcast-x-hazelcast-x-update-version-yln1o1
  namespace: demo
  resourceVersion: "8102465"
  uid: 8fa1d430-a94d-4bed-819f-216ea659b7c3
spec:
  backoffLimit: 10
  description: Latest patch version is available. Recommending version Update from
    5.5.2 to 5.5.6.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: HazelcastOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: hazelcast
      type: UpdateVersion
      updateVersion:
        targetVersion: 5.5.6
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
    kind: Hazelcast
    name: hazelcast
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
This release introduces multiple OpsRequests for Hazelcast, including Rotate-Authentication,
Reconfigure-TLS
To configure TLS in Hazelcast using a HazelcastOpsRequest, we have introduced ReconfigureTLS. You can enable TLS in Hazelast by simply deploying a YAML configuration:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz
  tls:
    issuerRef:
      name: hz-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - hazelcast
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

### Rotate-Authentication
To modify the credential in Hazelcast you can use RotateAuth. This HazelcastOpsRequest will update the credential.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: hz
  apply: IfReady
```

## MariaDB
In this release, Ops-Request now supports TLS reconfiguration, version upgrades, volume expansion, and authentication rotation for distributed MariaDB in Galera Cluster mode. Additionally, full cluster disaster recovery support has been fixed.

### Git-Sync
We have added a new feature now you can initialize MariaDB from the public/private git repository. Here’s a quick example of how to configure it.

#### From Public Repository

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       - --repo=https://github.com/kubedb/mysql-init-scripts
       - --depth=1
       - --period=60s
       - --link=current
       - --root=/my-path
       # terminate after successful sync
       - --one-time 
  version: "10.5.23"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

#### From Private Repository

Create a kubernetes secret containing your SSH keys:

```bash
ssh-keyscan $YOUR_GIT_HOST > /tmp/known_hosts
kubectl create secret generic -n demo git-creds \
   --from-file=ssh=$HOME/.ssh/id_rsa \
   --from-file=known_hosts=/tmp/known_hosts
```
Apply the following yaml:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       # update with your private repository    
       - --repo=git@github.com:refat75/mysql-init-scripts.git
       - --depth=1
       - --period=60s
       - --link=current
       - --root=/my-path
       # terminate after successful sync
       - --one-time 
       authSecret:
         name: git-creds
       # run as git sync user 
       securityContext:
         runAsUser: 65533
  version: "10.5.23"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

This example refers to initialization from a private git repository `.spec.init.git.args` represents the arguments required to represent the git repository and its actions. You can find details at [git_syc_docs](https://github.com/kubernetes/git-sync/blob/master/README.md)

`.spec.init.git.authSecret` holds the necessary information to pull from the private repository You have to provide a secret with the `id_rsa` and `githubkwonhosts`. You can find detailed information at [git_sync_docs](https://github.com/kubernetes/git-sync/blob/master/docs/ssh.md).


If you are using a different authentication mechanism for your git repository, please consult the documentation for [git-sync](https://github.com/kubernetes/git-sync/tree/master/docs) project.

## MySQL

In this release, a new OpsRequest called `StorageMigration` has been introduced. It allows you to change your database’s `StorageClass` seamlessly. 

First deploy a standalone MySQL:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.35"
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 16Gi
  deletionPolicy: WipeOut
```
Then apply the following `StorageMigration` ops-request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample-mysql
  migration:
    storageClassName: local-path
    oldPVReclaimPolicy: Retain
```
> Note: If you don’t want to keep the old PersistentVolume set `spec.migration.oldPVReclaimPolicy` to `Delete`

## Pgbouncer
KubeDB now supports Recommandation for PgBouncer. For more details follow [kubedb docs](https://kubedb.com/docs/v2025.8.31/guides/pgbouncer/)

### Git-Sync

We have added a new feature now you can initialize Redis from the public/private git repository. Here’s a quick example of how to configure it.

#### From Public Repository

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-demo
  namespace: demo
spec:
  version: "1.24.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "postgres-demo"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  init:
    script:
      scriptPath: "sync-test"
      git:
        args:
        - --repo=<desired repo>
        - --depth=1
        - --add-user=true
        - --period=60s
        - --one-time
        securityContext:
          runAsUser: 999
```
#### From Private Repository
```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-demo
  namespace: demo
spec:
  version: "1.24.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "postgres-demo"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  init:
  script:
    scriptPath: "current"
    git:
      args:
      # use --ssh for private repository
      - --ssh
      - --repo=<desired repo>
      - --depth=1
      - --period=60s
      - --link=current
      - --root=/init-script-from-git
      # terminate after successful sync
      - --one-time
      authSecret:
        name: git-creds
      # run as git sync user
        securityContext:
          runAsUser: 999
```
This example refers to initialization from a private git repository .spec.init.git.args represents the arguments required to represent the git repository and its actions. You can find details at git_syc_docs

.spec.init.git.authSecret holds the necessary information to pull from the private repository You have to provide a secret with the id_rsa and githubkwonhosts You can find detailed information at git_sync_docs . If you are using a different authentication mechanism for your git repository, please consult the documentation for the git-sync project.
.spec.init.git.securityContext.runAsUser the init container git_sync runs with user 999.
.spec.podTemplate.Spec.securityContext.fsGroup In order to read the ssh key the fsGroup also should be 999.

## Pgpool

KubeDB now supports Recommandation for Pgpool. For more details follow [kubedb docs](https://kubedb.com/docs/v2025.8.31/guides/pgpool/)

### Git-Sync

We have added a new feature now you can initialize Redis from the public/private git repository. Here’s a quick example of how to configure it.

#### From Public Repository

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-demo
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: <postgres-ref>
    namespace: demo
  initConfig:
    pgpoolConfig:
      num_init_children : 6
      max_pool : 65
      child_life_time : 400
  deletionPolicy: WipeOut
  init:
    script:
      scriptPath: "sync-test"
      git:
        args:
        - --repo=<desired repo>
        - --depth=1
        - --add-user=true
        - --period=60s
        - --one-time
        securityContext:
          runAsUser: 999
```
#### From Private Repository
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-demo
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: <postgres-ref>
    namespace: demo
  initConfig:
    pgpoolConfig:
      num_init_children : 6
      max_pool : 65
      child_life_time : 400
  deletionPolicy: WipeOut
  init:
  script:
    scriptPath: "current"
    git:
      args:
      # use --ssh for private repository
      - --ssh
      - --repo=<desired repo>
      - --depth=1
      - --period=60s
      - --link=current
      - --root=/init-script-from-git
      # terminate after successful sync
      - --one-time
      authSecret:
        name: git-creds
      # run as git sync user
        securityContext:
          runAsUser: 999
```
This example refers to initialization from a private git repository .spec.init.git.args represents the arguments required to represent the git repository and its actions. You can find details at git_syc_docs

.spec.init.git.authSecret holds the necessary information to pull from the private repository You have to provide a secret with the id_rsa and githubkwonhosts You can find detailed information at git_sync_docs . If you are using a different authentication mechanism for your git repository, please consult the documentation for the git-sync project.
.spec.init.git.securityContext.runAsUser the init container git_sync runs with user 999.
.spec.podTemplate.Spec.securityContext.fsGroup In order to read the ssh key the fsGroup also should be 999.

## Postgres
A new OpsRequest called `StorageMigration` has been introduced. It allows you to change your database’s `StorageClass` seamlessly. 

First deploy a standalone PostgreSQL:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres
  namespace: demo
spec:
  version: "13.13"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
Then apply the following `StorageMigration` ops-request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample-postgres
  migration:
    storageClassName: local-path
    oldPVReclaimPolicy: Retain
```
> Note: If you don’t want to keep the old PersistentVolume set `spec.migration.oldPVReclaimPolicy` to `Delete`

## Redis
### Git-Sync

We have added a new feature now you can initialize Redis from the public/private git repository. Here’s a quick example of how to configure it. 

### From Public Repository

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-demo
  namespace: demo
spec:
  version: 7.2.4
  mode: Cluster
  init:
    script:
      scriptPath: "sync-test"
      git:
        args:
        - --repo=<desired git repository>
        - --depth=1
        - --add-user=true
        - --period=60s
        - --one-time
        securityContext:
          runAsUser: 999
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  deletionPolicy: WipeOut
```
### From Private Repository
```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-demo
  namespace: demo
spec:
  version: 7.2.4
  mode: Cluster
  init:
  script:
    scriptPath: "current"
    git:
      args:
      # use --ssh for private repository
      - --ssh
      - --repo=<desired repo>
      - --depth=1
      - --period=60s
      - --link=current
      - --root=/init-script-from-git
      # terminate after successful sync
      - --one-time
      authSecret:
        name: git-creds
      # run as git sync user
        securityContext:
          runAsUser: 999
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  deletionPolicy: WipeOut
```
This example refers to initialization from a private git repository .spec.init.git.args represents the arguments required to represent the git repository and its actions. You can find details at git_syc_docs

.spec.init.git.authSecret holds the necessary information to pull from the private repository You have to provide a secret with the id_rsa and githubkwonhosts You can find detailed information at git_sync_docs . If you are using a different authentication mechanism for your git repository, please consult the documentation for the git-sync project.
.spec.init.git.securityContext.runAsUser the init container git_sync run with user 999.
.spec.podTemplate.Spec.securityContext.fsGroup In order to read the ssh key the fsGroup also should be 999.

## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).