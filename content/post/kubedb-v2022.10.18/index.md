---
title: Announcing KubeDB v2022.10.18
date: "2022-10-18"
weight: 25
authors:
- Md. Alif Biswas
tags:
- autoscaler
- cloud-native
- database
- elasticsearch
- gitops
- kibana
- kubedb
- kubernetes
- mariadb
- memcached
- mongodb
- mysql
- opensearch
- percona
- percona-xtradb
- pgbouncer
- postgresql
- proxysql
- redis
- schema-manager
---

We are pleased to announce the release of [KubeDB v2022.10.18](https://kubedb.com/docs/v2022.10.18/setup/). This post lists all the major changes done in this release since the last release. `Kubernetes version 1.25` is supported in KubeDB from this release. This release offers some major features like **Externally Managed AuthSecret**, **Percona XtraDB OpRequests**, **ProxySQL OpsRequests**, **Redis Sentinel OpsRequests**, **PostgreSQL Logical Replication**, **MongoDB Hidden Member Support**, **Autoscaler Support for MySQL, Postgres, Redis, Redis Sentinel & PerconaXtraDB** etc.
It also contains various improvements and bug fixes. You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.10.18/README.md).

## Externally Managed AuthSecret

In this release, we have introduced a new boolean field `externallyManaged` in the CRD of every database we support. If enabled, the operator will wait for the custom auth secret given in the `authSecret` field. It allows users to provide auth secrets via [Bitnami sealed secrets](https://github.com/bitnami-labs/sealed-secrets) or [external secrets](https://github.com/external-secrets/external-secrets) for GitOps. Here is an example for MariaDB using `Externally Managed Secrret`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
   name: sample-mariadb
   namespace: demo
spec:
   authSecret:
      name: custom-auth
      externallyManaged: true
   version: "10.5.8"
   ...
   terminationPolicy: WipeOut
```

## KubeDB Autoscaler

We have added Autoscaler support for MySQL, Postgres, Redis, RedisSentinel & PerconaXtraDB in this release.  
We support `compute` (to autoscale CPU & memory resources) & `storage` (to autoscale persistent storage resources) autoscaling currently.
The structure of those autoscaler yamls is similar for all databases, except, RedisSentinelAutoscaler doesn't support storage autoscaling.
`compute` utilizes the `VerticalScale` OpsRequest & `storage` utilizes the `VolumeExpansion` OpsRequest internally while autoscaling.
There are two more common fields in the CRDs spec :
1) `spec.database` to refer to the actual database server.
2) `spec.opsRequestOptions` to control the behavior of the ops-manager operator. It has two fields `apply` and `timeout`.

The supported values of `spec.opsRequestOptions.apply` is `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the database is `Ready`. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.
`spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest(in seconds). If a step doesn't finish within the specified timeout, the ops request will result in failure.

NB: We have also added Arbiter & HiddenNode support in MongoDB Autoscaler. We refer to the hidden-node corresponding section in MongoDB below, for that.

Here is an example of the RedisAutoscaler object, where we want to autoscale the CPU & memory resources of the Redis pods.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisAutoscaler
metadata:
  name: redis-auto
  namespace: demo
spec:
  databaseRef:
    name: redis-db
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
  compute:
    cluster:
      trigger: "On" # whether compute autoscaler is enabled
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits" # Limits will also be calculated in recommendation
      minAllowed: # Specifies the minimum amount of resources that will be recommended.
        cpu: 600m
        memory: 600Mi
      maxAllowed: # Specifies the maximum amount of resources that will be recommended.
        cpu: 1
        memory: 1Gi
      podLifeTimeThreshold: 10m  # Specifies the minimum lifetime of a pod before update. (OOM is an exception in this case)
      resourceDiffPercentage: 20 # if the diff between current & recommended resource is less than ResourceDiffPercentage, autoscaler Operator will ignore the updating.
```

Example of a PerconaXtraDBAutoscaler object, where we want to autoscale the storage resources.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PerconaXtraDBAutoscaler
metadata:
  name: px-auto
  namespace: demo
spec:
  databaseRef:
    name: pxc
  opsRequestOptions:
    apply: Always
    timeout: 10m
  storage:
    perconaxtradb:
      trigger: "On"
      usageThreshold: 60 # If PVC usage percentage is less than the usageThreshold, we don't want to scale the storage.
      scalingThreshold: 50 # If the usageThreshold criteria is satisfied, we want to scale PVC storage by the specified scalingThreshold.
      expansionMode: "Offline" # possible values : `Offline` & `Online`.
```


## MongoDB

### Hidden Node Support

We introduce MongoDB hidden members in this latest release. A hidden member in mongoDB is a part of replica-set.
It maintains the copy of the primary data set, but remains invisible to client applications. They are good for workloads with different usage patterns.
Like, You are using inMemory databases for your primary & electable secondaries (to improve latency issue), but at the same time, you also want your data not to be lost in time of pod restart, then Hidden member is the solution for you. 

Here is a sample hidden member enabled mongoDB sharded cluster example:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongo-sh
  namespace: demo
spec:
  version: "4.4.6"
  shardTopology:
    configServer:
      replicas: 3
      storage:
        resources:
          requests:
            storage: 5Gi
        storageClassName: "linode-block-storage"
    mongos:
      replicas: 2
    shard:
      replicas: 2
      shards: 2
      storage:
        resources:
          requests:
            storage: 5Gi
        storageClassName: "linode-block-storage"
  terminationPolicy: WipeOut
  arbiter: {}
  hidden:
    podTemplate:
      spec:
        resources:
          requests:
            cpu: "500m"
            memory: "500Mi"
    replicas: 2
    storage:
      storageClassName: "linode-block-storage"
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 15Gi
```

### OpsRequests for Hidden Node

We have also added all the possible opsRequests for MongoDB Hidden node. Like `VerticalScaling`, `HorizontalScaling`, `Reconfigure` & `VolumeExpansion`. 

Here are some sample given for those opsRequests :

#### Volume Expansion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: mg-sh
  volumeExpansion:
    shard: 12Gi
    hidden: 20Gi
    mode: "Online"
```

#### Version Scaling

```yaml
spec:
  type: VerticalScaling
  ...
  verticalScaling:
    ...
    hidden:
      requests:
        memory: "600Mi"
        cpu: "600m"
```

#### Horizontal Scaling

```yaml
spec:
  type: HorizontalScaling
  ...
  horizontalScaling:
    ...
    hidden:
      replicas: 1
```

#### Reconfigure

```yaml
spec:
  type: Reconfigure
  ...
  configuration:
    ...
    hidden:
      configSecret:
        name: log-conf <the secret nam with your hidden-node configurations>
```

### Autoscaler features

We have added compute autoscaler support for MongoDB Arbiter & HiddenNode and Storage autoscaler support for Hidden Node.

Here are the truncated examples:

#### Compute autoscaling on Arbiter

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
...
spec:
  ...
  compute:
    arbiter:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 540m
        memory: 256Mi
```

#### Storage autoscaling on Hidden members

```yaml
spec:
  ...
  storage:
    hidden:
      trigger: "On"
      usageThreshold: 30
      scalingThreshold: 40
      expansionMode: "Online"
```

## MySQL

### New Version Support

We have added the most recent MySQL version. Now you can create a Standalone , Group Replication, Innodb Cluster, Read Replica and Semi-sync Cluster with `MySQL 8.0.31`.
You can get started with `mysql:8.0.31` by applying the yaml below:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql
  namespace: demo
spec:
  version: "8.0.31"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  terminationPolicy: WipeOut
```

### BugFix and Improvement

Several bugs in mysql ops request including mysql version upgrading and reconfigure TLS and Custom configuration are fixed.
OpsRequest for innodb cluster were improved. And in this release innodbCluster is more stable in provisioning and failover.
Now kubedb managed MySQL instance can automatically recover when some group member has extra transaction on it. It will clone from the Primary instance to recover.

## MariaDB

Along with externally managed auth secret support, We have fixed bugs and made improvements in several places of MariaDB Provisioner, OpsRequests and Autoscaler.

## Percona XtraDB Cluster

### OpsRequests (Day-2 Operations)

In this release we have introduced the PerconaXtraDB OpsRequests that simplify the day-2 operations of Percona XtraDB Cluster. 
Here is the list of supported Percona XtraDB OpsRequests:

#### Horizontal Scaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: hor-scale-up-px
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: sample-pxc
  horizontalScaling:
    member: 5
  timeout: 5s
```

#### Vertical Scaling

```yaml
spec:
  type: VerticalScaling  
  ...
  verticalScaling:
    perconaxtradb:
      requests:
        memory: "600Mi"
        cpu: "0.1"
      limits:
        memory: "600Mi"
        cpu: "0.1"
```

#### Version Update

```yaml
spec:
  type: Upgrade  
  ...
  upgrade:   
    targetVersion: "8.0.28"
```

#### Reconfigure

```yaml
spec:
  type: Reconfigure
  ...
  configuration:   
    configSecret:
      name: px-config
```

#### Reconfigure TLS

```yaml
spec:
  type: ReconfigureTLS
  ...
  tls:
    requireSSL: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: px-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

#### Volume Expansion

```yaml
spec:
  type: VolumeExpansion
  ...
  volumeExpansion:
    mode: Offline
    perconaxtradb: 2Gi
```

### Grafana Dashboards

Grafana dashboards, for better monitoring of KubeDB Provisioned Percona XtraDB Cluster, are added in this release. Here is the list of dashboards supported on
KubeDB provisioned Percona XtraDB Cluster:
- Summary dashboard shows overall summary of a Percona XtraDB Cluster instance.
- Pod dashboard shows individual pod-level information.
- Database dashboard shows Percona XtraDB Clusters internal metrics for an instance.
- Galera Cluster dashboard shows Percona XtraDB replication related metrics.

### Externally Managed System User Secret

Along with the externally managed auth secret, We are supporting externally managed system user secrets for monitor and replication user. We have fixed some bugs on the initialization scripts of KubeDB provisioned Percona XtraDB Cluster.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  systemUserSecrets:
    replicationUserSecret:
      name: custom-replication-auth
      externallyManaged: true
    monitorUserSecret:
      name: custom-monitor-auth
      externallyManaged: true
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  ...
  terminationPolicy: WipeOut
```

### Alerting

We have added configurable alerting support for Percona XtraDB Cluster.

### New Version Support

Percona XtraDB Cluster version `8.0.28` is now supported on KubeDB from this release.

## PostgreSQL Logical Replication

In this release, we have introduced Logical Replication for PostgreSQL. Here we have brought two new CRDs `Publisher` and `Subscriber`. `Publisher` CRD will create and maintain publication in the database and `Subscriber` CRD will create and maintain subscription for both `KubeDB` managed `Publisher` and externally managed publication. It also implements `double-optin` to restrict unauthorized access of `Subscriber` to the `Publisher`.

Sample yaml for `Publisher`:

```yaml
apiVersion: postgres.kubedb.com/v1alpha1
kind: Publisher
metadata:
  name: publisher-sample
  namespace: demo
spec:
  name: my_pub
  serverRef:
    name: publisher-db
  databaseName: pub
  allowedSubscribers:
    namespaces:
      from: Selector
      selector:
        matchLabels:
         kubernetes.io/metadata.name: test
    selector:
      matchLabels:
        replication.kubedb.com/instance_name: subscriber  
  tables:
    - table_1
  parameters:
    operations:
      - insert
      - update
      - delete
      - truncate
  deletionPolicy: Delete

```

Sample yaml for `Subscriber`:

```yaml
apiVersion: postgres.kubedb.com/v1alpha1
kind: Subscriber
metadata:
  name: subscriber-sample
  namespace: test
  labels:
    replication.kubedb.com/instance_name: subscriber
spec:
  name: mysub
  serverRef:
    name: subscriber-db
  databaseName: sub
  parameters:
    tableCreationPolicy: IfNotPresent
    createSlot: true
    enabled: true
    connect: true
  publisher: 
    managed:
      namespace: demo
      refs:
        - name: publisher-sample
  deletionPolicy: Delete
```

## Redis

### Redis Ops Requests in Sentinel Mode

- Horizontal and Vertical Scaling
- Version Update
- Reconfigure
- Reconfigure TLS
- Volume Expansion
- Replace Sentinel

To run Redis Database Cluster and Redis Sentinel Cluster in TLS-secured mode, both Redis and Sentinel need to be configured using the same `Issuer` or `ClusterIssuer` provided by `cert-manager`.

So Adding TLS or Removing TLS in Redis Sentinel Mode should be done carefully as the Redis Sentinel Cluster needs to be replaced as well. Because the existing Redis Sentinel Cluster can not monitor the Redis Database Cluster anymore.
So a new Redis Sentinel Cluster is needed. Here is a sample YAML to add TLS in Redis Database Cluster running in Sentinel Mode.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: redis-ops-tls-add
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-redis
  tls:
    sentinel:
      ref:
        name: sentinel-tls
        namespace: test
      removeUnusedSentinel: true
    issuerRef:
      apiGroup: cert-manager.io
      name: redis-ca-issuer
      kind: ClusterIssuer
```

In the TLS specification, users need to tell a new sentinel reference because if Redis Database Cluster is TLS enabled, the old Sentinel can not monitor anymore. So, KubeDB will look for a Redis Sentinel Cluster which is TLS enabled. If the user did not create such a Redis Sentinel Cluster, KubeDB would create one using the issuer given in the TLS specification.

In this release, we are introducing another Ops Request which enables users to replace the Redis Sentinel Cluster of the Redis Database Cluster. Users need to provide a sentinel reference that replaces the old Sentinel. The user needs to create a Redis Sentinel Cluster beforehand and provide a reference to that to perform the `ReplaceSentinel` Ops Request. 
A sample ReplaceSentinel Ops Requests YAML is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: replace-sentinel
  namespace: demo
spec:
  type: ReplaceSentinel
  databaseRef:
    name: sample-redis
  sentinel:
    ref:
      name: new-sentinel
      namespace: demo
    removeUnusedSentinel: true

```

If `removeUnusedSentinel` is set to true, KubeDB checks if the old sentinel is monitoring any other Redis Database Cluster after the replacement. If no, KubeDB deletes the old Redis Sentinel Cluster.

### Redis Sentinel Ops Request
- Horizontal and Vertical Scaling
- Version Update
- Reconfigure TLS
  - Rotate Certificates
  - Update Certificates

Here are some sample YAMLs given for those OpsRequests

#### Horizontal Scaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: sentinel-horizontal-scaling
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sentinel-tls
  horizontalScaling:
    replicas: 5
```

#### Vertical Scaling

```yaml
spec:
  type: VerticalScaling  
  ...
  verticalScaling:
    redissentinel:
      requests:
        memory: "600Mi"
        cpu: "0.1"
      limits:
        memory: "600Mi"
        cpu: "0.1"
```

#### Reconfigure TLS

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: tls-rotate-redis-sentinel
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sentinel-tls
  tls:
    rotateCertificates: true
```

#### Version Update

```yaml
spec:
  type: UpdateVersion  
  ...
  upgrade:   
    targetVersion: "7.0.5"
```

### New Version Support
KubeDB Redis is now supporting the latest Redis version 7.0.5.

## ProxySQL

### External MySQL

KubeDB ProxySQL is now supporting externally managed MySQL servers as backend. The user needs to create an appbinding with necessary information of that backend like following.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: 
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTi... #truncated
    service:
      name: mysql-external
      path: /
      port: 3306
      scheme: mysql
    url: tcp(mysql-external.demo.svc:3306)/
  secret:
    name: backend-auth
  tlsSecret:
    name: mysql-external-client-cert 
  type: mysql
  version: 5.7.36
```

* **spec.clientConfig** contains all the necessary information for establishing connection . It is a required field .
* **spec.secret.name** contains a secret name, which has key `username` and `password` and corresponding data in it. The credential provided with this secret can be root. If not , the credential must be privileged with `SELECT` and `INSERT` in `sys.*` and `mysql.*`. This is a required field.
* **spec.tlsSecret.name** contains a secret name, which must include key `ca.crt` , `tls.crt` and `tls.key` . It is necessary for tls secured connection between ProxySQL and MySQL backend. It is an optional field.
* **version** contains the MySQL server version. It is a required field.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-auth
  namespace: demo
type: Opaque
data:
  username: bW9ub3g= # must be privileged with SELECT and INSERT on sys.* and mysql.*
  password: bW9ub3hwYXNz 
```

With the below Yaml one can configure a KubeDB ProxySQL for the above-mentioned external MySQL. Just need to mention the appbinding name in the `spec.backend.name` section and the operator will do the rest.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  authSecret:
    name: clusterauth
    externallyManaged: true
  version: "2.3.2-debian"
  replicas: 3
  backend:
    name: mysql-external
  syncUsers: false
  initConfig:
    mysqlUsers: 
      - username: test
        active: 1
        default_hostgroup: 2
    adminVariables: 
      restapi_enabled: true
      restapi_port: 6070
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: proxysql-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  terminationPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
```

Lastly, there is a sample externally-managed auth secret. This secret is mentioned in the above ProxySQL yaml in the `spec.authSecret.name` section. This field is optional if absent the operator will create one itself and patch the name in the yaml. In either case, this auth secret will be used in KubeDB ProxySQL to configure `admin-cluster_username` and `admin-cluster_password` variables. The secret must contain `username` and `password` key and corresponding data. Also the `password` must be alpha-numeric.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: clusterauth
  namespace: demo
type: Opaque
data:
  username: eHVzcmV0
  password: eHVnbzEyc2xvdmlh # must be alpha-numeric
```

### Ops-requests

#### Update Version

with `UpdateVersion` ops-request user can update ProxySQL version .

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: up
  namespace: demo
spec:
  type: UpdateVersion
  proxyRef:
    name: proxy-server
  upgrade:
    targetVersion: "2.4.4-debian"
```

Currently available versions : `2.3.2-debian`, `2.3.2-centos`, `2.4.4-debian`, `2.4.4-centos`

#### Horizontal Scaling

* Scale Up

```yaml 
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: scale-up
  namespace: demo
spec:
  type: HorizontalScaling  
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 5
```

* Scale Down

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: scale-down
  namespace: demo
spec:
  type: HorizontalScaling  
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 3
```

#### Reconfigure

* Reconfigure User
    * Add User \
      Add user to `mysql_users` table.

    ```yaml
    ...
    spec:
      type: Reconfigure  
      ...
      configuration:
        mysqlUsers:
          users: 
          - username: A_Add
            active: 1
            default_hostgroup: 2  
          - username: B_Add
            active: 1
            default_hostgroup: 3
          reqType: add
    ```

    * Delete User \
      Remove users from `mysql_users` table.

    ```yaml
    ...
    configuration:
        mysqlUsers:
          users: 
          - username: C_Del
          - username: D_Del
        reqType: delete
    ```

    * Update User Configuration \
      Update user configuration in `mysql_users` table

    ```yaml
    ...
    configuration:
        mysqlUsers:
            users: 
            - username: E_Upd
                active: 0
                default_hostgroup: 3
            - username: F_Upd
                active: 1
                default_hostgroup: 3
            reqType: update
    ```

* Reconfigure MySQL Query Rules \
  User can add, update and delete into mysql_query_rules table with this ops-request. Here we have provided a sample for update only. The other two are similar as the user reconfig ops requst .

```yaml
  ...
  configuration:
    mysqlQueryRules:
      rules: 
        - rule_id: 2
          active: 0
      reqType: update
```

* Reconfigure Global Variables \
  Update global variables in ProxySQL server.

```yaml
  ...
  configuration:
    adminVariables:
      refresh_interval: 2055
      cluster_check_interval_ms: 205
    mysqlVariables:
      max_transaction_time: 1540000
      max_stmts_per_connection: 19
```

#### Reconfigure TLS

Reconfigure all the TLS configuration with ReconfigureTLS ops-request. This is related to only proxysql frontend tls.

* Add TLS

```yaml
...
spec:
  type: ReconfigureTLS
  ...
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: my-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses: 
        - "tasdid@appscode.com"
```

* Update TLS

```yaml
  ...
  tls:
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "md.alifbiswas@gmail.com"
      certificates:
    - alias: client
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "md.alifbiswas@gmail.com"
```

* Rotate TLS

```yaml
  ...
  tls:
    rotateCertificates: true
```

* Remove TLS

```yaml
  ...
  tls:
    remove: true
```

#### Restart

Finally if the user wants to restart all the ProxySQL pods , this ops-request will be helpful.

```yaml
...
spec:
  type: Restart  
  proxyRef:
    name: proxy-server
```

## Support for Private Registry

In this release, we have fixed a bug related to using private Docker registry to detect Docker image dogest. This bug affects private registries using custom domain with self-signed tls certificates.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.10.18/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.10.18/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
