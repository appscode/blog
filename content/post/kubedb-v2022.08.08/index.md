---
title: Announcing KubeDB v2022.08.08
date: "2022-08-08"
weight: 25
authors:
- Md Kamol Hasan
tags:
- cloud-native
- dashboard
- database
- elasticsearch
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
- postgresql
- redis
- schema-manager
---

We are pleased to announce the release of [KubeDB v2022.08.08](https://kubedb.com/docs/v2022.08.08/setup/). This post lists all the major changes done in this release since the last release.
This release offers some major features like **Recommendation Engine**, **Configurable Health-Checker**, **Custom Volumes & VolumeMounts**, **Redesigned PerconaXtraDB Operator**, **MongoDB In-Memory AutoScaler**, **ProxySQL Declarative Configuration**, **ProxySQL User Sync**, **Elasticsearch V8 Ops-Request Support**, etc.
It also contains various improvements and bug fixes. You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.08.08/README.md).

## KubeDB Recommendation Engine

In this release, we have introduced a recommendation engine to generate recommendations for database resources.
Recommendation-engine runs inside the `kubedb-ops-manager` pod. Currently, it will generate **Rotate TLS** and **Version Upgrade** ops-request recommendations depending on the TLS certificate expiry date and available updated database versions. We have already introduced Supervisor to execute the recommendation in a user-defined maintenance window. To install the Supervisor helm chart, please visit [here](https://github.com/kubeops/installer/tree/master/charts/supervisor).

To know more about Recommendation-engine and Supervisor, please watch the following videos where we discussed briefly and showed a demo on how to execute recommendations.

- [KubeDB AutoOps: Automate Day 2 Life-cycle Management for Databases on Kubernetes](https://blog.byte.builders/post/kubedb-autoops-webinar-2022.04.13/)
- [KubeDB AutoOps Playlist](https://www.youtube.com/playlist?list=PLoiT1Gv2KR1gszfJIwCcC4Kj-JcE4uyR_)

To see the generated recommendations:

```bash
$ kubectl get recommendations.supervisor.appscode.com -A
```

### Recommendation: Rotate TLS

By default, the recommendation engine will generate a Rotate TLS recommendation for the TLS secured database before one month of the TLS certificate expiry date. But if the TLS certificate lifespan is less than one month, then it will generate the recommendation before half of the TLS certificate lifespan. Also while installing KubeDB, users can specify custom flags to configure the recommendation-engine to create Rotate TLS recommendations before a specific time of TLS certificate expiry date.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  labels:
    app.kubernetes.io/instance: mongo-sh-tls
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-tls
  name: mongo-sh-tls-x-mongodb-x-rotate-tls-w6irof
  namespace: demo
spec:
  description: TLS Certificate is going to be expire on 2022-11-03 08:44:08 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: MongoDBOpsRequest
    metadata:
      name: rotate-tls
      namespace: demo
    spec:
      databaseRef:
        name: mongo-sh-tls
      tls:
        rotateCertificates: true
      type: ReconfigureTLS
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: MongoDB
    name: mongo-sh-tls
```

```bash
$ helm install kubedb appscode/kubedb \
    --version v2022.08.08 \
    --namespace kubedb --create-namespace \
    --set kubedb-provisioner.enabled=true \
    --set kubedb-ops-manager.enabled=true \
    --set kubedb-autoscaler.enabled=true \
    --set kubedb-dashboard.enabled=ture \
    --set kubedb-schema-manager.enabled=true \
    --set kubedb-ops-manager.recommendationEngine.genRotateTLSRecommendationBeforeExpiryMonth=2 \
    --set-file global.license=/path/to/license/file
```

With the above installation, the recommendation engine will generate recommendations before two months of the TLS certificate expiry date. To know more about the recommendation engine flags please visit [here](https://github.com/kubedb/installer/tree/master/charts/kubedb-ops-manager).

### Recommendation: Version Upgrade

The recommendation engine will generate major/minor and patch version upgrade recommendations depending on the available new database versions. It also recommends the same version upgrade if the database/sidecar image version is out of date according to the catalog version. Even if the image digest is mismatched with the image registry(in the case of a retagged image), it will also generate the same version upgrade recommendation. When a user approves any recommendation, the Supervisor will create an Ops Request to execute the recommendation according to the specified maintenance window.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 4.2.3
  labels:
    app.kubernetes.io/instance: mg-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-upgrade
  name: mg-test-x-mongodb-x-update-version-sqa7dc
  namespace: demo
spec:
  description: Latest Major/Minor version is available. Recommending version upgrade
    from 4.2.3 to 4.4.6
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: MongoDBOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: mg-test
      type: Upgrade
      upgrade:
        targetVersion: 4.4.6
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
    kind: MongoDB
    name: mg-test
```

## Health Checker

In this release, we've improved KubeDB database health checks. We've added a new field called `healthChecker` under spec. It controls the behavior of health checks. It has the following fields:

- `spec.healthChecker.periodSeconds` : Specifies the interval between each health check iteration.
- `spec.healthChecker.timeoutSeconds` : Specifies the timeout for each health check iteration.
- `spec.healthChecker.failureThreshold` : Specifies the number of consecutive failures to mark the database as NotReady.
- `spec.healthChecker.disableWriteCheck` : KubeDB does a database write a check by default, but if you want to disable it, you can use this field.

Example YAML:

```yaml
spec:
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
    disableWriteCheck: true
```

## Custom Volumes & VolumeMounts

With this new release, users will be able to add additional volumes and volumeMounts to the database container. It will provide more flexibility to users.

```yaml
 podTemplate:
   spec:
     volumeMounts:
     - name: my-volume
       mountPath: /usr/share/custom-config
     volumes:
     - name: my-volume
       secret:
         secretName: custom-config
```

## OpsRequests
There are two changes on all the kubeDB opsRequest Types.

### Apply
We have added the `apply` field to OpsRequest CRD to control the execution of the opsRequest depending on the Database state. Its supported values are `IfReady` & `Always`.
Use `IfReady` if you want to process the opsReq only when the database is Ready.  And use `Always` if you want to process the execution of opsReq irrespective of the Database state.


### Skip
We have also added opsRequest status named `Skipped`, & redesigned the execution order of OpsRequest CROs. The idea behind skipping is something like this :

1. If there are multiple opsReqs of the same ops-request type (like 3 'VerticalScaling' ) in the Pending state,
   We don't want to reconcile them one by one. Because only reconciling the last one is enough, & that is the user's desired spec.
   we are setting opsReq Phase `Skipped` in this situation, for all, except the last one.

2. If there are multiple opsReqs of different ops-request type (like 3 'VerticalScaling', 2 `Upgrade`, 2 `Reconfigure`) in the Pending state,
   After skipping the previous step, there will be exactly one opsReq of each type in Pending
   And now, as they are different types, We want to reconcile the oldest one first.


## Image Digest

Now KubeDB operators use docker image with digest value (ie. `elasticsearch:7.12.0@sha256:383e9fb572f3ca2fdef5ba2edb0dae2c467736af96aba2c193722aa0c08ca7ec` ) while provisioning databases. For a given docker image, if there is any vulnerability-fix pushed in the docker registry with the same tag, the running pods will have no idea about the change. Adding the digest helps the KubeDB recommendation engine to verify whether the current docker images are up to date or not. If not, the engine will generate recommendations for Restart/Upgrade ops requests to update the database pods with the latest images.



## MongoDB

### AutoScaler

To autoscale the InMemory databases, A field named `inMemoryStorage` has been added. There are two fields inside it: `usageThresholdPercentage`, `scalingFactorPercentage`.

We have also added a field named `opsRequestOptions` on MongoDBAutoscaler. It has 3 fields inside. `readinessCriteria` to specify the `oplogMaxLagSeconds` which defines the maximum lagging time among the replicas, & `objectsCountDiffPercentage` which defines the maximum objectsCount difference among replicas.

We have two more fields : `timeout` & `apply`. If any step in opsRequest execution doesn’t finish with a specified timeout, the corresponding opsRequest will result in `Failed`. The `apply` field is the same as discussed in the above OpsRequest section.  A sample MongoDBAutoscaler YAML is given below :

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as-rs
  namespace: demo
spec:
  opsRequestOptions:
    readinessCriteria:
      oplogMaxLagSeconds: 10
      objectsCountDiffPercentage: 10
    timeout: 15
    apply: "IfReady"
  databaseRef:
    name: mg-rs
  compute:
    replicaSet:
      trigger: "On"
      podLifeTimeThreshold: 3m
      minAllowed:
        cpu: 600m
        memory: 250Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      inMemoryStorage:
        usageThresholdPercentage: 70
        scalingFactorPercentage: 50
```

### Fixes

- Fixed the Connection leak issue when ping fails.

## PerconaXtraDB

We have redesigned the PerconaXtraDB Cluster Operator of KubeDB. Now It has the latest features of KubeDB including Galera clustering, failure recovery, custom configuration, TLS support, and monitoring using Prometheus. On this release, we are providing support for Percona XtraDB Cluster version 8.0.26. We have added a SystemUserSecrets field on PerconaXtraDB where custom MySQL users can be added for monitoring and clustering.

Sample  YAML for the  PerconaXtraDB Cluster:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  systemUserSecrets:
    monitorUserSecret:
      name: db-monitor-user
    replicationUserSecret:
      name: db-replication-user
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  tls:
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
  requireSSL: true
  terminationPolicy: WipeOut
```

## MariaDB

**Primary Component Selector:** From this release, No request will be routed to a MariaDB Cluster node that is not part of the primary component. So, A node only receives requests via database service if the node is a member of the primary component of the Galera Cluster. This feature will ensure that requests to database cluster nodes are successful.

**Custom Auth Secret:** The bug related to using user-provided custom auth secret is fixed. Validation of a custom auth secret is also added in this release.

## MySQL

**OpsRequest:** All the MySQL ops requests have been reworked . We have also added support for  for `InnoDBCluster` , `SemiSync` and `ReadReplica`. Now you can perform MySQLOpsRequest for all the available topologies.

**Custom Auth Secret:** The bug related to using user-provided custom auth secret is fixed. Validation of a custom auth secret is also added in this realease.

## Elasticsearch

The re-constructed Elasticsearch health checker provides you with continuous read/write accessibility checking in the database. When the disk space reaches watermark limits, Elasticsearch has a protective function that locks the indices, stopping new data from being written to them. This is to stop Elasticsearch from using any further disk causing the disk to become exhausted. While the indices are locked, new data is not indexed and thus is not searchable. The read/write healthChecker makes sure that database accessibility is reflected in the `Elasticsearch` CRO status. The R/W accessibility checking can be halted by setting `.spec.healthCheck.disableWriteCheck` to `true`.

**OpsRequest:** `OpsRequest` support has been added for Elasticsearch V8. Now you can perform `Upgrade`, `Restart`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion` and `ReconfigureTLS` in your KubeDB managed `Elasticsearch V8` cluster with `ElasticsearchOpsRequest` CRO. Upgrading the Elasticsearch cluster version will automatically trigger an upgrade for `ElasticsearchDashboard` CRO resulting in a compatible version upgrade for KubeDB managed `KIbana V8` deployed in your cluster. OpsRequest support for `Opensearch V1` has also been re-constructed to function more natively.

## ProxySQL

**New Version Support:**  KubeDB ProxySQL is now supporting proxysql-2.3.2-centos in the latest release.

**Declarative Configuration:** You can now set up the ProxySQL bootstrap configuration with declarative YAML. You can mention all the configurations regarding `mysql_users`, `mysql_query_rules`, `mysql_variables` and `admin_variables` through the YAML or config secret  and KubeDB ProxySQL will generate a configuration file based on these and bootstrap the server with that. With this feature ProxySQL admin does not need to set up anything from the command line.

**User synchronization with backend:** `syncUser` field is newly introduced to KubeDB ProxySQL. When enabled it syncs all the necessary information of the user with the backend server. Any updates in the backend will automatically reflect in the ProxySQL server if needed. With this feature, ProxySQL admin does not need to enter users manually in the ProxySQL server or update any information like password. Even in the case when a user is deleted from the backend, the operator will auto remove it from ProxySQL server.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-mysql
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  mode: GroupReplication
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: mysql-server
    replicas: 3
  syncUsers: false
  initConfig:
    mysqlUsers: 
      - username: reader
        active: 1
        default_hostgroup: 2
    adminVariables: 
      refresh_interval: 2050
      cluster_mysql_servers_save_to_disk: false
    mysqlVariables:
      max_connections: 1024
      default_schema: "information_schema"
    mysqlQueryRules:
      - rule_id: 1
        active: 1
        match_digest: "^SELECT.*FOR UPDATE$"
        destination_hostgroup: 2
        apply: 1
      - rule_id: 2
        active: 1
        match_digest: "^SELECT"
        destination_hostgroup: 3
        apply: 1
    terminationPolicy: WipeOut
```

## PostgreSQL

**Removal of PG-Coordinator Container for Standalone Mode:** Now kubeDB Postgres standalone doesn't require running the pg-coordinator sidecar container with it. In this release, We have removed all the pg-coordinator sidecar container dependencies.

## Redis

**New Version Support:** We are excited to announce KubeDB now supports the newly released Redis Version `7.0.4`. You can deploy and manage your Redis Version `7.0.4` Cluster using KubeDB. Along with Version `7.0.4`, we also added support for two new versions; Redis `5.0.14` and Redis `6.2.7`.

**Ops Requests:** KubeDB Ops Requests have been improved for Standalone Mode and Cluster Mode in this release. You can perform `Upgrade`, `Reconfigure`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion` and `ReconfigureTLS` in your KubeDB managed `Redis` cluster with `RedisOpsRequest`

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.08.08/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
