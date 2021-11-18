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


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
