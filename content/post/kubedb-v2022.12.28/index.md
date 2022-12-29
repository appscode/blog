---
title: Announcing KubeDB v2022.12.28
date: "2022-12-28"
weight: 25
authors:
- Mehedi Hasan
tags:
- cloud-native
- dashboard
- database
- elasticsearch
- kafka
- kibana
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- opensearch
- percona
- percona-xtradb
- postgresql
- redis
---

We are pleased to announce the release of [KubeDB v2022.12.28](https://kubedb.com/docs/v2022.12.28/setup/). This post lists all the major changes done in this release since the last release. 

The release was mainly focused on updating documentations, writing test and bug fixes and improvements. Aside there was some major features like support for **Kafka,** **ProxySQL: Autoscaler,Monitoring,MariaDB and Percona-XtraDB backend support** . Also, new versions for **ElasticSearch** , **MariaDB**,**PostgreSQL**, **PgBouncer**, **Kibana** are available.

You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.12.28/README.md).

## Kafka
We have added support for Apache Kafka. With this addition, now it is possible to provision Kafka in Kraft mode using KubeDB. Kafka cluster in [kraft mode](https://kafka.apache.org/documentation/#kraft) is provisioned without Zookeeper dependency. It comes with two types of clustering options. One is `Combined` mode clustering where each node acts as both brokers and controllers. The other one is `Topology` mode clustering where each node acts as a dedicated controller or broker. It offers a number of cool features including TLS/SSL encryption, Health Checker etc.

**TLS Support:** To add an extra layer of security, KubeDB can enable TLS/SSL configurations for Kafka. KubeDB uses cert-manager v1 api to provision and manage TLS certificates. Enabling TLS security for Kafka results in provisioning kafka with `SASL_SSL` security protocol to secure the channels that are used to communicate with the kafka servers.which ensures that the communication is encrypted and authenticated using `SASL_PLAIN` mechanism. If TLS is not enabled Kafka is provisioned with `SASL_PLAINTEXT` security protocol by default which sets kafka communications authenticated using `SASL_PLAIN` mechanism without encryption. If security is disabled in kafka yaml, Kafka is provisioned with simple `PLAINTEXT` security protocol where all communication is configured without any authentication or encryption mechanism.

**Health Checker:** KubeDB ensures Kafka health is continuously monitored by checking server response and connectivity status. It also checks for kafka topic creation, publishing messages to a topic and acknowledgement of published messages. Kafka healthChecker also comes with configurable features to control the behavior(interval, timeout, failure threshold etc.) of health Checks.

**Supported version:** Support for Kafka version 3.3.0 is added in this release.

Following is a sample YAML to provision TLS secured kafka cluster with 2 dedicated controller nodes and 3 dedicated broker nodes of version `3.3.0` in demo namespace. Create an Issuer named `kafka-ca-issuer` using cert-manager prior to applying this YAML.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-sample
  namespace: demo
spec:
  version: 3.3.0
  enableSSL : true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: kafka-ca-issuer
      kind: Issuer
  storageType: Durable
  terminationPolicy: Delete
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

## ProxySQL
**MariaDB and Percona-XtraDB backend:**

From this release we have added support for `MariaDB` and `Percona-XtraDB` as ProxySQL backend. Both KubeDB managed and external databases are allowed to be set as backend.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 3
  mode: Galera
  backend:
    name: sample-mariadb
  syncUsers: true
  terminationPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
```

**Vertical Scaling Ops-request :**

We have added vertical scaling ops-request for ProxySQL in this release.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: vertical-scale
  namespace: demo
spec:
  type: VerticalScaling  
  proxyRef:
    name: proxy-server
  verticalScaling:
    proxysql:
      requests:
        memory: "1200Mi"
        cpu: "0"
      limits:
        memory: "1200Mi"
        cpu: "0"
```

**Autoscaler:**
Like other KubeDB managed databases from now you will have autoscaling support for ProxySQL.

We support compute (to autoscale CPU & memory resources) autoscaling currently. The structure of those autoscaler yamls is similar for all databases. Compute utilizes the VerticalScale OpsRequest internally while autoscaling. There are two more common fields in the CRDs spec :

`spec.proxysql` to refer to the actual proxysql server.
`spec.opsRequestOptions` to control the behavior of the ops-manager operator. It has two fields: apply and timeout.

The supported values of `spec.opsRequestOptions.apply` are `IfReady` & `Always`. Use `IfReady` if you want to process the opsReq only when the proxysql server is Ready. And use Always if you want to process the execution of opsReq irrespective of the proxysql server state. spec.opsRequestOptions.timeout specifies the maximum time for each step of the opsRequest(in seconds). If a step doesn’t finish within the specified timeout, the ops request will result in failure.

Here is an example of the ProxySQLAutoscaler object, where we want to autoscale the CPU & memory resources of the ProxySQL pods.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ProxySQLAutoscaler
metadata:
  name: proxy-as-compute
  namespace: demo
spec:
  proxyRef:
    name: proxy-server
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    proxysql:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 250m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

`podLifeTimeThreshold: 5m`   Specifies the minimum lifetime of a pod before update. (OOM is an exception in this case)

`resourceDiffPercentage: 20` if the diff between current & recommended resource is less than ResourceDiffPercentage, autoscaler Operator will ignore the updating.

### Monitoring

For better monitoring of KubeDB Provisioned ProxySQL Grafana dashboards are added in this release. Here's the list of dashboards supported on KubeDB provisioned ProxySQL

- Summary dashboard shows overall summary of a ProxySQL instance.
- Pod dashboard shows individual pod-level information.
- Database dashboard shows ProxySQL internal metrics for an instance.

To learn more about it go to the [link]( https://github.com/appscode/grafana-dashboards/tree/master/proxysql.)

### Alert 

We have added configurable alerting support for ProxySQL.

## PgBouncer

### New Version Support

We have added support for the latest PgBouncer version `1.18.0` in this release.

### New Features: Custom AuthSecret, Custom Configuration

Custom `authSecret` and custom `configSecret` support for PgBouncer is available from this release. Now users can provide custom admin `username` and `password` for PgBouncer and custom configuration for `pgbouncer.ini` file.
We removed `userListSecretRef` from `spec` section and `authSecretRef` from `spec.databases` section. From this release, `Userlist` Secret will be obtained from the provided secret in the Postgres AppBinding `db.spec.databases.databaseref`.

### Custom AuthSecret

#### Sample Custom AuthSecret 

```yaml
apiVersion: v1
stringData:
  password: "12345"
  username: custom
kind: Secret
metadata:
  name: demo-custom
  namespace: demo
type: kubernetes.io/basic-auth
```

#### Sample PgBouncer

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  authSecret:
    name: demo-custom
    externallyManaged: true
  databases:
  - alias: "testdb"
    databaseName: "test"
    databaseRef:
      name: "app1"
      namespace: demo
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
  terminationPolicy: WipeOut
 ```

### Custom Configuration

Sample configuration file `user.conf` to create configSecret `pb-configuration`

```text
defaultPoolSize=20
poolMode=session
```

#### Sample Pgbouncer

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configSecret:
    name: pb-configuration
  databases:
  - alias: "testdb"
    databaseName: "test"
    databaseRef:
      name: "app1"
      namespace: demo
  connectionPool:
    port: 5432
    minPoolSize: 0
    maxClientConnections: 20
    reservePoolSize: 5
    maxDBConnections: 0
    maxUserConnections: 0
    statsPeriodSeconds: 60
    authType: md5
  terminationPolicy: WipeOut

```

## Elasticsearch

### New Version Support

We have added support for the latest [Elasticsearch version 8.5.2](https://www.elastic.co/guide/en/elasticsearch/reference/current/release-notes-8.5.2.html) with xpack authplugin in this release. This version is referred to as ElasticsearchVersion `xpack-8.5.2`.  You can deploy this version as an TLS secured Elasticsearch combined cluster with the following yaml.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  version: xpack-8.5.2
  enableSSL: true
  replicas: 3
  storageType: Durable
  podTemplate:
    spec:
      resources:
        limits:
          memory: 1.5Gi
        requests:
          cpu: 500m
          memory: 1.5Gi
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

## Redis

### New Version Support

We have added two latest Redis versions `6.2.8` and `7.0.6` in this release. To deploy a Redis Standalone instance with version `Redis 7.0.6`, you can apply this yaml

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: 7.0.6
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
``` 

### Fixes And Improvement

- Redis Auth problem in Redis Cluster Mode has been solved. You can deploy Redis in any Redis Mode with or without Auth enabled.

## MongoDB

### Documentation

All the OpsRequest related docs have been updated to reflect the features supported till KubeDB version 2022.12.28. Docs related to Hidden Nodes are also available now.

### Fixes And Improvement

- Issues regarding Upgrade & ReconfigureTLS OpsRequest.

## MariaDB

**New Version:** New MariaDB version 10.10.2 has been added in this release.

## PerconaXtraDB

### Documentation

KubeDB docs have been updated with the latest Percona XtraDB Cluster features

### Fixes And Improvement

- Fixed issue on Reconfigure TLS OpsReq

## MySQL

### Documentation

KubeDB documentations for MySQL `Innodb Cluster` , `Semi Sync` , `Read Replica` , `Autoscaling` are added and updated with the latest release. Existing documentations are polished and improved.

### Fixes And Improvement

- bugs related to reconfigure tls and version upgrading are fixed 

## PostgreSQL

### New Version Support

We have added support for the latest PostgreSQL versions `15.1`, `14.6`, `13.9`, `12.13` in this release.

### Bug Fix and improvement

- Fixed Transfer Leadership issue on raft switchover.
- Fixed issue with Single User mode on raft sidecar.
- Fix issue with pre-conflict error detection in Logical Replication.

## Kibana

### New Version Support

We have added support for the latest [Kibana version 8.5.2](https://www.elastic.co/guide/en/kibana/current/release-notes-8.5.2.html) which is compatible with Elasticsearch 8.5.2. If you have an Elasticsearch cluster with version `xpack-8.5.2` provisioned with KubeDB, you can apply the following yaml to provision TLS secured Kibana 8.5.2 standalone cluster. Just referring to the database in the dashboard yaml is enough as the operator provisions compatible Kibana version with the Elasticsearch cluster.

```yaml
apiVersion: dashboard.kubedb.com/v1alpha1
kind: ElasticsearchDashboard
metadata:
  name: es-cluster-dashboard
  namespace: demo
spec:
  enableSSL: true
  databaseRef:
    name: es-cluster
  terminationPolicy: WipeOut
``` 

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.12.28/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.12.28/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
