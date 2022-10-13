---
title: Announcing KubeDB & Stash v2021.11.24
date: 2021-11-30
weight: 25
authors:
- Tamal Saha
aliases:
- /post/kubedb.v2021.11.24/
tags:
- cloud-native
- kubernetes
- database
- elasticsearch
- mariadb
- memcached
- mongodb
- mysql
- postgresql
- redis
- kubedb
---

We are pleased to announce the release of KubeDB and Stash `v2021.11.24`. This post lists all the major changes done in this release since `v2021.09.30`. The headline features of this release are `OpenSearch` support, `InnoDB Cluster` support for MySQL, support for PostgreSQL version `14.1` and `PostGIS`.

## General API Improvements

- **Custom Labels/Annotations Support**: Now you can provide custom labels/annotations to the pods, pod’s controller (ie. StatefulSets), and services for any supported databases. Labels applied to the KubeDB custom resources will be passed down to all offshoots including the database pods. But you can also set labels to database pods or controllers (StatefulSets) via the PodTemplate and to the services via the ServiceTemplates.

## Grafana Dashboards

We now have Grafana Dashboards for all KubeDB and Stash resource types. Please reach out to us if you are an Enterprise customer on how to setup these dashboards in your clusters.

## ElasticSearch

- **OpenSearch**:  KubeDB supports `OpenSearch` version `1.1.0`. Now you can deploy and manage the OpenSearch cluster in the Kubernetes native way.
- **Custom Labels/Annotations Support**: Now you can provide custom labels/annotations to the pods, pod’s controller (ie. StatefulSets), and services.

  **Sample Elasticsearch Yaml:**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: Elasticsearch
  metadata:
  name: es
  namespace: demo
  spec:
  serviceTemplates:
  - alias: primary
    metadata:
      labels:
        elasticsearch.com/custom-svc-label: set
      annotations:
        passTo: service
  podTemplate:
    metadata:
      labels:
        elasticsearch.com/custom-pod-label: set
      annotations:
        pass-to: pods
    controller:
      labels:
        elasticsearch.com/custom-sts-label: set
      annotations:
        pass-to: statefulset
  version: opensearch-1.1.0
  storageType: Durable
  terminationPolicy: WipeOut
  monitor:
    agent: prometheus.io
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  ```

- **Exporter**: Elasticsearch exporter images are upgraded to `v1.3.0`.

  **Sample Elasticsearch Yaml:**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: Elasticsearch
  metadata:
  name: es
  namespace: demo
  spec:
  serviceTemplates:
  - alias: primary
    metadata:
      labels:
        elasticsearch.com/custom-svc-label: set
      annotations:
        passTo: service
  podTemplate:
    metadata:
      labels:
        elasticsearch.com/custom-pod-label: set
      annotations:
        pass-to: pods
    controller:
      labels:
        elasticsearch.com/custom-sts-label: set
      annotations:
        pass-to: statefulset
  version: opensearch-1.1.0
  storageType: Durable
  terminationPolicy: WipeOut
  monitor:
    agent: prometheus.io
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  ```

- **Reconfigure Ops Request**: Now users can apply/change configuration to a running Elasticsearch cluster without downtime. While reconfiguring, the operator also executes `security-admin.sh` script to reload the security index.

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
  name: reconfigure-es-apply-config
  namespace: demo
  spec:
  type: Reconfigure
  configuration:
    applyConfig:
      elasticsearch.yml: |
        thread_pool:
          search:
            size: 35
            queue_size: 500
            min_queue_size: 10
            max_queue_size: 1000
            auto_queue_frame_size: 2000
            target_response_time: 1s
  databaseRef:
    name: sample-es
  ```

- Prometheus exporter sidecar collects metrics from all types of Elasticsearch nodes.

## MariaDB

- Exporter: MySQL/MariaDB exporter is up to date with the latest `mysqld-exporter` release `v0.13.0`.
- Custom Labels/Annotations Support: Now you can provide custom labels/annotations to the pods, pod’s controller (ie. StatefulSets), and services.

  **Sample MariaDB Yaml**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MariaDB
  metadata:
    name: mariadb
    namespace: demo
  spec:
    serviceTemplates:
      - alias: primary
        metadata:
          labels:
            kubedb.com/svc: primary
            svc-label: primary
          annotations:
            passTo: service
      - alias: stats
        metadata:
          labels:
            kubedb.com/svc: stats
            svc-label: stats
          annotations:
            passTo: stats-service
    podTemplate:
        metadata:
          labels:
            pass-to: pod
          annotations:
            annotate-to: pod
        controller:
          labels:
            controler-label : custom-label
          annotations:
            controller-annotations : custom-annotation
    version: "10.6.4"
    replicas: 3
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

## MongoDB

- **Replicaset Configuration**: In this release, we've added support for configuring replicaSet using custom configSecret. Previously, we only supported the file-based configuration i.e. `mongod.conf` file. Now, you can use a key `replicaset.json` to configure the replicaSet configuration. Also, you can reconfigure it by using the `Reconfiguration` ops request. Below is an example of the configSecret which includes both replicaset configuration and `mongod.conf` configuration file.

  ```yaml
  apiVersion: v1
  kind: Secret
  type: Opaque
  metadata:
    name: custom-config
    namespace: demo
  stringData:
    mongod.conf: |
      net:
          maxIncomingConnections: 50000
    replicaset.json: |
      {
        "settings" : {
          "electionTimeoutMillis" : 5000
        }
      }
  ```

- **Custom Labels/Annotations Support**: Now, you can provide custom labels and annotations to the pods, statefulsets, and services associated with the database. Below is an example MongoDB replicaSet yaml, that shows how to provide the custom labels for pods, statefulsets, and services.

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MongoDB
  metadata:
    name: mg-sh
    namespace: demo
  spec:
    version: "4.4.6"
    shardTopology:
      configServer:
        replicas: 3
        storage:
          resources:
            requests:
              storage: 1Gi
        podTemplate:
          controller:
            labels:
              sts-label: "config"
          metadata:
            labels:
              pod-label: "config"
      mongos:
        replicas: 1
        podTemplate:
          controller:
            labels:
              sts-label: "mongos"
          metadata:
            labels:
              pod-label: "mongos"
      shard:
        replicas: 3
        shards: 2
        storage:
          resources:
            requests:
              storage: 1Gi
        podTemplate:
          controller:
            labels:
              sts-label: "shard"
          metadata:
            labels:
              pod-label: "shard"
    storageType: Durable
    monitor:
      agent: prometheus.io/operator
      prometheus:
        serviceMonitor:
          labels:
            release: prometheus
          interval: 10s
    serviceTemplates:
    - alias: primary
      metadata:
        labels:
          svc-label: primary
        annotations:
          svc-annotations: primary
    - alias: stats
      metadata:
        labels:
          svc-label: stats 
        annotations:
          svc-annotations: stats    
    terminationPolicy: WipeOut
  ```

- **Bug Fix**: Previously, MongoDB compute autoscaler would generate recommendations too frequently for sharded clusters. In this release, we've fixed that bug. Now, MongoDB compute autoscaler generates recommendations as expected. Also, we've fixed some other minor bugs in KubeDB MongoDB operators.

## MySQL

If you are a current user of KubeDB, then you will need to run `MySQLOpsRequest` to upgrade the database versions as we have changed the coordinator sidecar and Prometheus exporter sidecar. If you are using the following versions, then please upgrade accordingly. You can find the currently supported versoins [here](https://github.com/kubedb/installer/blob/v2021.11.24/catalog/active_versions.json#L89-L94).

| Current Version | Mode | Upgraded Version  	|
|---	|---	|---	|
| 5.7.x | Standalone/Group replication | 5.7.36  	|
| 8.0.3 | Standalone | 8.0.3-v4   	|
| 8.0.3 | Group Replication | 8.0.17   	|
| 8.0.x | Standalone / Group Replication | 8.0.27   	|

- If you would like to upgrade to MySQL 5.7.x to 8.x, please first upgrade from 5.7.x to 5.7.36 and then upgrade from 5.7.36 to 8.0.27. This 2 step process is required to ensure that TLS communication works among the MySQL nodes during the upgrade process.
- If you are using 8.0.3 (Non-GA) Group Replicated cluster, now you will be able to upgrade to the 8.0.17 version and then upgrade to the latest 8.0.27. . This 2 step process is required to ensure that TLS communication works among the MySQL nodes during the upgrade process. We recommend upgrading to the 8.0.27 version because it has support for the clone plugin.
- If you are using 8.0.3 (Non-GA) Standalone cluster, then we are not able to automatically upgrade to a GA version of MySQL 8 instance. Because during the upgrade process, we need to empty the PVC so that the data files can be written in the correct format for MySQL 8 GA versions. You have 2 choices here currently:
  - Upgrade to `8.0.3-v4` which will use the updated sidecar and init containers.
  - Take a backup and restore into a blank MySQL 8.0.27 cluster.
- If you are starting with a new MySQL 8 instance, we recommend using MySQL 8.0.27 . We recommend against using 8.0.17 for new instances. 8.0.17 support is added so that 8.0.3 clusters can be upgraded to 8.0.27 eventually.
- To run the upgrade operation you can use an ops request like below. We always recommend taking a full database backup before running upgrade commands.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myopsreq-group
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: my-group
  upgrade:
    targetVersion: "8.0.20"
```

If you have any questions, please reach out via our existing chat channels or support@appscode.com email.

- **Memory Management**: Improved MySQL server memory management to avoid OOM errors for small, medium and large instances. Now, the operator will calculate default values for `innodb_buffer_pool_size` and `group_replication_message_cache _size` for MySQL instances with memory size 1G, 1G < Memory < 4G and > 4G.
- **Failover** : We have replaced the old replication mode detector sidecar with a new mysql-coordinator sidecar. This sidecar can recover from a single node failover, multiple node failover and complete cluster failover ensuring the MySQL node with the highest LSN is always the new primary after recovery.
- **Process Restart**: Added support for mysqld process restart on MySQL pod restart. If the mysqld process stops or is killed by OOM killer KubeDB will restart the mysqld process.
- **Custom Labels/Annotations support**: Now you can provide custom labels/ annotations to the pod and the pod’s controller (ie. StatefulSets) and services.

  **Sample MySQL Yaml**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MySQL
  metadata:
    name: m1
    namespace: demo
    labels:
      passMe: ToAllOffshootsIncludingPods
  spec:
    version: "8.0.27"
    topology:
      mode: GroupReplication
      group:
        name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
    authSecret:
      name: m1-auth
    storageType: "Durable"
    storage:
      storageClassName: "standard"
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
    init:
      script:
        configMap:
          name: mg-init-script
    monitor:
      agent: prometheus.io/operator
      prometheus:
        serviceMonitor:
          labels:
            app: kubedb
          interval: 10s
    requireSSL: true
    tls:
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
    configSecret:
      name: my-custom-config
    podTemplate:
      metadata:
        labels:
          passMe: ToDatabasePod
        annotations:
          passMe: ToDatabasePod
      controller:
        labels:
          passMe: ToStatefulSet
        annotations:
          passMe: ToStatefulSet
      spec:
        serviceAccountName: my-service-account
        schedulerName: my-scheduler
        nodeSelector:
          disktype: ssd
        imagePullSecrets:
        - name: myregistrykey
        args:
        - --character-set-server=utf8mb4
        env:
        - name: MYSQL_DATABASE
          value: myDB
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
    serviceTemplates:
    - alias: primary
      metadata:
        labels:
          passMe: ToService
        annotations:
          passMe: ToService
      spec:
        type: NodePort
        ports:
        - name:  http
          port:  9200
  ```

- **Exporter** : MySQL/MariaDB exporter has been replaced with a custom KubeDB managed exporter fork which exports necessary metrics for Group replication.
- **InnoDB Cluster**: We have added support for [MySQL InnoDB Cluster](https://dev.mysql.com/doc/refman/8.0/en/mysql-innodb-cluster-introduction.html) which will enhance your experience with mysql group replication. We’ve also added support for MySQL Router for InnoDB Cluster which acts  as a load balancer. InnoDB cluster itself supports the MySQL shell for which enables you to work with MySQL AdminAPI. `mysqlsh` allows you to work with InnoDB Cluster, InnoDB ClusterSet, and InnoDB ReplicaSet. Please consider InnoDB cluster support beta quality. There are a few known issues that we are looking to address in the next release. These issues include if multiple replicas try to clone data at the same time clone operator might fail and TLS reconfiguration support.

  **Sample MySQL InnoDB Yaml**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MySQL
  metadata:
    name: "innodb"
    namespace: demo
  spec:
    version: "8.0.27-innodb"
    replicas: 3
    topology:
      mode: InnoDBCluster
      innoDBCluster:
        router:
          replicas: 1
    storageType: Durable
    storage:
      storageClassName: "standard"
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
    terminationPolicy: WipeOut
  ```

## PostgreSQL

- Add support for PostgreSQL new Versions: `14.1`, `13.5`, `12.9`, `11.14`, `10.19`, `9.6.24`.
- Added support for PostGIS distribution.
- Added support for TimescaleDB 14.1 release.
- **Process Restart**: Added support for postgres process restart on Postgres pod restart. If the postgres process stops or is killed by OOM killer KubeDB will restart the postgres process.
- Few fixes regarding pg-coordinator which will improve the postgres performance and stability.

## Redis

Fixed The custom-config support for redis.

## Introducing Stash `v2021.11.24`

We are very excited to announce Stash `v2021.11.24`. In this release, we have made a few enhancements for Stash. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2021.11.24/README.md). Here, we are going to highlight the major changes.

### Added Elasticsearch 7.14.0 support

We have added support for Elasticsearch 7.14.0 in Stash. Now, you can backup and restore your Elasticsearch 7.x.x and the equivalent OpenSearch versions using this Stash addon.

### Apply runtime settings to the CronJob properly

Previously, Stash only supported applying runtime settings to the backup sidecar/Job. It did not pass those runtime settings to the respective backup triggering CronJob properly. As a result, some users were not able to force the CronJob to run on a particular node. Now, we pass all the pod level runtime settings to the CronJob. This will allow you to configure nodeSelector, securityContext etc. for the CronJob.

## Remove Google Analytics

Previously, we were using Google Analytics to count active users of Stash and KubeDB. In this release, we have removed the Google Analytics and replaced it with our in-house open source [auditor](https://github.com/bytebuilders/audit) library. This helps us to ensure that our customer's information is not shared with any third-party organization. The new analytics solution sends some basic information such as Product version, Kubernetes version and provider, Number of nodes in a cluster info etc. These information is used count number of active users and auditing billing information.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
