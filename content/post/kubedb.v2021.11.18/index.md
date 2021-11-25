---
title: Announcing KubeDB v2021.11.18
date: 2021-11-18
weight: 25
authors:
  - Tamal Saha
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

We are pleased to announce the release of KubeDB `v2021.11.18`. This post lists all the major changes done in this release since `v2021.09.30`. The headline features of this release are `OpenSearch` support, `Innodb Cluster` support for MySQL, Support for PostgreSQL version `14.1`, etc.


## ElasticSearch

- **OpenSearch**:  KubeDB supports `OpenSearch` version `1.1.0`. Now you can deploy and manage the OpenSearch cluster in the Kubernetes native way.
- **Exporter**: Elasticsearch exporter images are upgraded to `v1.3.0`.
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

- **Reconfigure Ops Request**: Now users can apply/change configuration to a running Elasticsearch cluster without downtime. While reconfiguring, the operator also executes  `security-admin.sh` script to reload the security index. 

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
- Collect metrics from all type of Elasticsearch node

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

- **Bug Fix**: Previously, MongoDB compute autoscaler was generating recommendations too frequently for the sharded cluster. In this release, we've fixed that bug. Now, MongoDB compute autoscaler generates recommendations as expected. Also, we've fixed some other minor bugs in KubeDB MongoDB.


## MySQL

- **Failover** : Fixed crash recovery  . Now kubedb can recover from a single node failover, multiple node failover and complete cluster failover.
- **Process Restart**: Added support for mysqld process restart on MySQL server failure. If the mysqld process stops or is killed by OOM killer Kubedb can restart the mysqld process.
- **Memory Management**: Improve MySQL server Internal memory management Configuration based On MySQL documentation.We have adjusted `innodb_buffer_pool_size` and `group_replication_message_cache _size` based on requested resources.
**Custom Labels/Annotations support**: Now you can provide custom labels/ annotations to the pod and the pod’s controller (ie. StatefulSets) and services.

  **Sample MySQL Yaml**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MySQL
  metadata:
    name: mysql
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
            controller-label : custom-label
          annotations:
            controller-annotation : custom-annotation
    version: "5.7.36"
    replicas: 3
    topology:
      mode: GroupReplication
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

- **Exporter** : MySQL/MariaDB exporter is up to date with the latest `mysqld-exporter` release `v0.13.0`.
- **Innodb Cluster**: We have added new support for Innodb Cluster which will enhance your experience with mysql group replication.We’ve also added support for MySQL Router for Innodb Cluster which acts  as a load balancer. Innodb cluster itself supports the MySQL shell for which enables you to work with MySQL AdminAPI which allows you to work with InnoDB Cluster, InnoDB ClusterSet, and InnoDB ReplicaSet.

  **Sample MySQL Innodb Yaml**

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
- Few fixes regarding pg-coordinator which will improve the postgres performance.

## Redis
Fix The custom-config support for redis 

## Introducing Stash `v2021.11.24`

We are very excited to announce Stash `v2021.11.24`. In this release, we have made a few enhancements for Stash. You can find the complete changelog here. Here, we are going to highlight the major changes.

### Remove Google Analytics

Previously, we were using Google Analytics to identify the usage pattern of Stash. In this release, we have removed the Google Analytics and replaced it with our in-house open sourced auditor library. This helps us to ensure that our customers' data are not shared with any third-party organization.

At the beginning of operator startup, Stash sends some basic information such as Stash version, Kubernetes version, cluster size, Stash license info etc. You can find what information we send from [here](https://github.com/kmodules/custom-resources/blob/master/apis/auditor/v1alpha1/siteinfo_types.go). We use [this](https://github.com/bytebuilders/audit) library to send this information.

Here, is a sample data that we collect:


```json
{
 "items": [
   {
     "key": "events.s106.user.537.k8s.c45ac3e2-ba76-4dce-879c-c8955ce74f27.product.stash-enterprise.group.YXVkaXRvci5hcHBzY29kZS5jb20=.resource.siteinfos.year.2021.month.11.day.25",
     "values": [
       {
         "resource": {
           "apiVersion": "auditor.appscode.com/v1alpha1",
           "kind": "SiteInfo",
           "kubernetes": {
             "clusterUID": "c45ac3e2-ba76-4dce-879c-c8955ce74f27",
             "controlPlane": {
               "notAfter": "2022-11-25T05:38:43Z",
               "notBefore": "2021-11-25T05:38:43Z"
             },
             "nodeStats": {
               "allocatable": {
                 "cpu": "4",
                 "memory": "16275088Ki"
               },
               "capacity": {
                 "cpu": "4",
                 "memory": "16275088Ki"
               },
               "count": 1
             },
             "version": {
               "buildDate": "2021-05-21T23:01:33Z",
               "compiler": "gc",
               "gitCommit": "5e58841cce77d4bc13713ad2b91fa0d961e69192",
               "gitTreeState": "clean",
               "gitVersion": "v1.21.1",
               "goVersion": "go1.16.4",
               "major": "1",
               "minor": "21",
               "platform": "linux/amd64"
             }
           },
           "metadata": {
             "creationTimestamp": null,
             "name": "5319394310677458003.stash-enterprise,kubedb-ext-stash"
           },
           "product": {
             "licenseID": "5319394310677458003",
             "productName": "stash-enterprise,kubedb-ext-stash",
             "productOwnerName": "appscode",
             "version": {
               "commitHash": "85b094e24a78ec970e00c1e0f0116fb4da74077e",
               "commitTimestamp": "2021-11-24T10:19:26",
               "compiler": "gcc",
               "gitBranch": "HEAD",
               "gitTag": "v0.17.0",
               "goVersion": "go1.17.3",
               "platform": "linux/amd64",
               "version": "v0.17.0",
               "versionStrategy": "tag"
             }
           }
         },
         "resourceID": {
           "group": "auditor.appscode.com",
           "version": "v1alpha1",
           "name": "siteinfos",
           "kind": "SiteInfo",
           "scope": "Cluster"
         },
         "licenseID": "5319394310677458003",
         "version": 184545,
         "timestamp": 1637819145
       }
     ]
   }
 ]
}
```

### Added Elasticsearch 7.14.0 support

We have added support for Elasticsearch 7.14.0 in Stash. Now, you can backup and restore your Elasticsearch 7.x.x and the equivalent OpenSearch versions using this Stash addon.


### Apply runtime settings to the CronJob properly

Previously, Stash only supported applying runtime settings to the backup sidecar/Job. It did not pass those runtime settings to the respective backup triggering CronJob properly. As a result, some users were not able to force the CronJob to run on a particular node. Now, we pass all the pod level runtime settings to the CronJob. This will allow you to configure nodeSelector, securityContext etc. for the CronJob.


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
