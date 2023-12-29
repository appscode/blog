---
title: Announcing KubeDB v2022.02.22
date: "2022-02-22"
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
- postgresql
- redis
- schema-manager
---

We are pleased to announce the release of [KubeDB v2022.02.22](https://kubedb.com/docs/v2022.02.22/setup/). This post lists all the major changes done in this release since the last release. This release offers support for the **Schema Manager for multi-tenancy**, **MySQL read replica**, **ElasticsearchDashboard (Kibana)**, Elasticsearch configurable JVM heap, MongoDB reprovision opsRequest, MongoDB configurable ephemeral storage, MongoDB JS file support in reconfigure opsRequest, MariaDB storage and compute autoscaling, MariaDB offline volume expansion, MariaDB reconfigure opsRequest, Postgres offline volume expansion, Redis disable authentication, etc. You can find the detailed change logs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.02.22/README.md).

## Schema Manager

We have added Multi-Tenancy support for MySQL, MongoDB, and PostgreSQL. KubeDB Multi-Tenancy support allows multiple users logically isolated in the same shared database server. It uses KubeVault to generate and maintain user credentials for databases. With schema-manager, users can create and initialize a database using a script. Users can also restore databases using Stash. So, KubeDB now supports managing schemas with declarative `YAML`.

- **Sample:** The following YAML will create a mysqldatabase schema in the cluster which will eventually end up creating a database `demo` inside  `mysql-server` which is a kubedb mysql server . The `demo` database will be created with character set utf8, encryption disabled and `readOnly` false,  as it is mentioned in the `spec.database.configuration` section. The operator will create a user and its credentials using the KubeVault and bind the permission under the subject which is mentioned in the `spec.accessPolicy.subjects` section. The credential and permission will be valid for 10 minutes. The time duration is  mentioned in the `spec.accessPolicy.defaultTTL` section. After that the schema will expire.

  ```yaml
  apiVersion: schema.kubedb.com/v1alpha1
  kind: MySQLDatabase
  metadata:
    name: sample-mysql-schema
    namespace: demo
  spec:
    database:
      serverRef:
        name: mysql-server
        namespace: dev
      config: 
        name: demo
        characterSet: utf8
        encryption: disable
        readOnly: 0
    vaultRef:
      name: vault
      namespace: dev
    accessPolicy:
      subjects:
        - kind: ServiceAccount
          name: "tester"
          namespace: "demo"
      defaultTTL: "10m"
    deletionPolicy: "Delete"
  ```

## Dashboard

We are very excited to announce that Dashboard support has been added to KubeDB. In this release, we have added support for Kibana that lets you visualize and manage KubeDB managed Elasticsearch. It offers users to deploy production-ready Kibana instances with features like TLS/SSL encryption, custom configuration, health-checker etc. Currently the Dashboard supports elasticsearchVersion `xpack-7.14.0`, more are coming soon.

- Sample Elasticsearch:

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: Elasticsearch
  metadata:
    name: es-standalone
    namespace: demo
  spec:
    version: xpack-7.14.0
    enableSSL: true
    replicas: 1
    storageType: Durable
    storage:
      storageClassName: "standard"
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
    terminationPolicy: DoNotTerminate
  ```

- Sample ElasticsearchDashboard:

  ```yaml
  apiVersion: dashboard.kubedb.com/v1alpha1
  kind: ElasticsearchDashboard
  metadata:
    name: ed-sample
    namespace: demo # Have to be in the same namespace as Elasticsearch instance
  spec:
    enableSSL: true
    databaseRef:
      name: es-standalone # pointing to the sample Elasticsearch instance
  ```

## Elasticsearch

- **New ElasticsearchVersion:** `kubedb-searchguard-5.6.16`.
- **Configurable JVM Heap:** Now JVM heap size can be defined in terms of percentage of memory. It can be global or node role-specific. The node-level setting has higher precedence than the global one. If not set, defaults to **50%** of the memory limit.

    ```yaml
    apiVersion: kubedb.com/v1alpha2
    kind: Elasticsearch
    metadata:
      name: sample-es
      namespace: demo
    spec:
      heapSizePercentage: 50 # applied to all nodes
      version: kubedb-searchguard-5.6.16
      storageType: Durable
      terminationPolicy: WipeOut
      topology:
        master:
          replicas: 1
          heapSizePercentage: 40 # -xms -xmx will be 40% of memory limit
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 100Mi
        data:
          heapSizePercentage: 70 # -xms -xmx will be 70% of memory limit
          replicas: 1
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 200Mi
        ingest:
        # heapSizePercentage is missing, will use the global one, -xms -xmx will be 50% of memory limit
          suffix: ingest
          replicas: 1
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 150Mi
    ```

## MongoDB

- **New Version Support:** In this release, We are happy to introduce the support for Percona MongoDB `4.4.10`.
- **Configure Ephemeral Storage:** We’ve added `ephermeralStorage` field to configure Ephemeral Storage. For example, if you want to configure the ephemeral storage of a `replicaSet`, you can configure it using the following YAML:

  ```yaml
  spec:
    ephemeralStorage:
      sizeLimit: "1Gi"
      medium: Memory
  ```

- **Reprovision Ops Request:** We have introduced a new ops request `Reprovision`. Using this ops request, you can newly provision a database, if there is any failure. Please note that, if you are using `Ephemeral` or `InMemory` database, your data will be lost. Sample reprovision YAML:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: MongoDBOpsRequest
  metadata:
    name: mops-reprovision-rs
    namespace: demo
  spec:
    type: Reprovision
    databaseRef:
      name: mg-rs
  ```

- **JS File Configuration:** From this release, you can provide a JS file in the configSecret, which will be applied on each reconciliation. Below is an example of the `configSecret` which includes the JS file configuration.

  ```yaml
  apiVersion: v1
  kind: Secret
  type: Opaque
  metadata:
    name: config
    namespace: demo
  stringData:
    configuration.js: |
        print("Hello World")
  ```

- **Reconfigure Js File Configuration:** You can also reconfigure the `configuration.js` using the Reconfigure ops request. Below is an example of the Reconfigure ops request:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: MongoDBOpsRequest
  metadata:
    name: mops-reconfiugre
    namespace: demo
  spec:
    type: Reconfigure
    databaseRef:
      name: mg-rs
    configuration:
      replicaSet:
        applyConfig:
          configuration.js: |
            print("hello world!")
  ```

- **Bug Fixes:** We’ve fixed various minor bugs in MongoDB Autoscaler.

## PostgreSQL

- **Offline Volume Expansion:** Earlier we had support for online volume expansion. In this scenario, It wasn’t necessary to restart pods to apply expansion changes in volume.  But this kind of expansion is not supported by some of the StorageClass. That's why we have introduced volume expansion offline which is going to be supported by those StorageClass types.

- **Bug Fixes:** Fixed minor Issues Regarding custom authentication. Removed invalid characters from Auth Password which are not supported by postgres client engine.

## **MariaDB**

- **Autoscaling:** Added autoscaling support for MariaDB cluster which can scale different resources like storage, memory, CPU automatically. Sample autoscaling yaml:

  ```yaml
  apiVersion: autoscaling.kubedb.com/v1alpha1
  kind: MariaDBAutoscaler
  metadata:
    name: md-autoscaler
    namespace: demo
  spec:
    databaseRef:
      name: sample-mariadb
    compute:
      mariadb:
        trigger: "On"
        podLifeTimeThreshold: 5m
        minAllowed:
          cpu: 250m
          memory: 350Mi
        maxAllowed:
          cpu: 1
          memory: 1Gi
        controlledResources: ["cpu", "memory"]
    storage:
      mariadb:
        trigger: "On"
        usageThreshold: 60
        scalingThreshold: 50
        expansionMode: "Online"
  ```

- **Offline Volume Expansion:** Now KubeDB support storage volume expansion with both online and offline modes. StorageClasses with offline mode do not allow volume expansion while the PVC/PV is being used by any pod.

- **Reconfigure MariaDB**: MariaDB Reconfigure opsRequest now supports applyConfig that can be used to apply changes on existing config files or creates a new config file if not it doesn’t exist.

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: MariaDBOpsRequest
  metadata:
    name: mdops-reconfigure-apply-config
    namespace: demo
  spec:
    type: Reconfigure
    databaseRef:
      name: sample-mariadb
    configuration:   
      applyConfig:
        new-custom-config.cnf: |
          [mysqld]
          max_connections = 200
  ```

- **Bug Fixes and Improvements:**  Fixed appending of custom config directory in `my.cnf` more than once, on mysql process restart. Also we’ve fixed the certificate and secret cleanup issue of MariaDB ReconfigureTLS opsRequest.

## MySQL

- **MySQL Read Replica:** We have added support for MySQL Read Replica which will allow you to create an asynchronous replication of the primary database server. That may be used in offloading your read request, analytics traffic from the  primary instance, etc. To allow `Read Replica` to read from  the source we had added a new field, `allowedReadReplicas` that will ensure you which replica instance and from what namespace it would be allowed to read from the source. By default it will allow read-replicas from the same namespace. You can see the demo [here](https://youtu.be/egzPGc6Yk_A).

- **Source Sample:**

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MySQL
  metadata:
    name: mysql-source
    namespace: demo
  spec:
    version: "8.0.27"
    replicas: 3
    topology:
      mode: GroupReplication
    allowedReadReplicas:
      namespaces:
        from: Selector
        selector:
          matchLabels:
            kubernetes.io/metadata.type: readReplica
      selector:
        matchLabels:
          kubedb.com/instance_name: ReadReplica
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

- **Read-Replica Sample:** For creating a read replica, you need to mention the topology `ReadReplica` and  the `sourceRef` . Here, we are referring to a group replicated cluster, `mysql-source` as a source. You can create a read replica from a standalone instance or replicated cluster.

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: MySQL
  metadata:
    name: mysql-read
    namespace: demo
    labels:
      kubedb.com/instance_name: ReadReplica
  spec:
    version: "8.0.27"
    topology:
      mode: ReadReplica
      readReplica:
        sourceRef:
            name: mysql-source
            namespace: demo
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

- **Bug fixes and Improvement:** It contains various bug fixes and codebase improvement in reconfigureTLS ops request.

## Redis

- **Disable Authentication for Redis Cluster:** In some cases, users want to configure their Redis cluster without enabling Authentication to avoid complexity regarding authentication as it’s not necessary in some use cases. So, We have Added Support to Disable Authentication for Redis Cluster.

  ```yaml
  apiVersion: kubedb.com/v1alpha2
  kind: RedisSentinel
  metadata:
    name: sen
    namespace: demo
  spec:
    version: 6.2.5
    disableAuth: true
    storageType: Durable
    storage:
      resources:
        requests:
          storage: 1Gi
      storageClassName: "standard"
      accessModes:
      - ReadWriteOnce
    terminationPolicy: WipeOut
  ```

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.02.22/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.02.22/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
