---
title: Announcing KubeDB v2024.6.4
date: "2024-06-04"
weight: 14
authors:
- Obaydullah
tags:
- alert
- archiver
- backup
- clickhouse
- cloud-native
- dashboard
- database
- druid
- grafana
- kafka
- kubedb
- kubernetes
- kubestash
- memcached
- mongodb
- mssqlserver
- pgbouncer
- pgpool
- postgres
- postgresql
- prometheus
- rabbitmq
- redis
- restore
- s3
- schema-registry
- security
- singlestore
- solr
- tls
- zookeeper
---

We are pleased to announce the release of [KubeDB v2024.6.4](https://kubedb.com/docs/v2024.6.4/setup/). This release includes features like (1) OpsRequest support for Druid, Memcached, Pgpool, RabbitMQ and Singlestore. (2) Autoscaling support for Druid, Pgpool and Singlestore. (3) PDB support for Singlestore, Pgpool, ClickHouse and Zookeeper. (4) initial release of ClickHouse and Kafka Schema Registry support (5) Multi user support for PgBouncer. (6) TLS support for Microsoft SQL Server. This post lists all the major changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.6.4/README.md). Now, you can proceed to detail the specific features and updates included in the release.
## ClickHouse
We are thrilled to announce that KubeDB now supports ClickHouse, an open-source column-oriented DBMS (columnar database management system) for online analytical processing (OLAP) that allows users to generate analytical reports using SQL queries in real-time.
ClickHouse works `100-1000x` faster than traditional database management systems, and processes hundreds of millions to over a billion rows and tens of gigabytes of data per server per second. With a widespread user base around the globe, the technology has received praise for its reliability, ease of use, and fault tolerance.

Here's a sample manifest to provision ClickHouse.

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
        storage: 2Gi
  deletionPolicy: WipeOut
```
Here's a sample manifest to provision ClickHouse in clusterTopology mode.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch-cluster
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      node:
        host: clickhouse-keeper.click-keeper
        port: 2181
    cluster:
    - name: click-cluster
      shards: 2
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
  deletionPolicy: WipeOut
```
**New Version support**: `24.4.1`

> Note: To get clickhouse keeper server host and port, You need to setup [clickhouse-keeper](https://clickhouse.com/docs/en/guides/sre/keeper/clickhouse-keeper) server manually. 

## Druid
In this release, Druid API has been updated. Now, Druid can be installed with a simpler YAML. Consequently, users do not need to mention the required nodes (i.e. `coordinators`, `brokers`, `middleManager`, `historicals`) anymore and the KubeDB operator will handle those and deploy the mandatory nodes with the default configurations. You can find the sample YAML below:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  metadataStorage:
    name: mysql
    namespace: demo
    createTables: true
  zookeeperRef:
    name: zookeeper
    namespace: demo
  topology: {}
```

### OpsRequest
In this release, support for Druid Ops Request has been integrated. Druid Ops Request provides a declarative configuration for the Druid administrative operations like database restart, vertical scaling, volume expansion, etc. in a Kubernetes native way.

#### Restart
Restart ops request is used to perform a smart restart of the Druid cluster. An example YAML is provided below:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: druid
```

#### Vertical Scaling:
Vertical Scaling allows you to vertically scale the Druid nodes (ie. pods). The necessary information required for vertical scaling, must be provided in the `spec.verticalScaling` field.
An example yaml is provided below:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name:  dops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: druid
  verticalScaling:
    middleManagers:
      resources:
        limits:
          memory: 2560Mi
        requests:
          cpu: 900m
          memory: 2560Mi
    coordinators:
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 600m
          memory: 1Gi
```

#### Volume Expansion:
Volume Expansion is used to expand the storage of the Druid nodes (ie. pods). The necessary information required for volume expansion, must be provided in `spec.volumeExpansion` field. Example YAML:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: dops-vol-exp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: druid
  volumeExpansion:
    mode: "Online"
    middleManagers: 4Gi
    historicals: 4Gi
```

### Autoscaler
Support for Druid Compute Autoscaling for all druid nodes (i.e. pods) and Storage Autoscaling for druid data nodes (i.e. `historicals` & `middleManagers` pod) has also been added. To enable autoscaling with a particular specification users need to install a Custom Resource Object of Kind `DruidAutoscaler`. DruidAutoscaler is a Kubernetes Custom Resource Definitions (CRD). It provides a declarative configuration for autoscaling Druid compute resources and storage of database components in a Kubernetes native way.  Some sample DruidAutoscaler CRs for autoscaling different components of database is given below:
```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DruidAutoscaler
metadata:
  name: druid-as
  namespace: demo
spec:
  databaseRef:
    name: druid  
  compute:
    middleManagers:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 600m
        memory: 3Gi
      maxAllowed:
        cpu: 1
        memory: 5Gi
      resourceDiffPercentage: 20
      controlledResources: [ "cpu", "memory" ]
  storage:
    historicals:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 70
      scalingThreshold: 50
```
## Elasticsearch
**New Version support**: `xpack-8.13.4`(Elasticsearch), `opensearch-2.14.0`(Opensearch)
Elasticsearch yaml for xpack-8.13.4:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  storageType: Durable
  version: xpack-8.13.4
  enableSSL: true
  topology:
    data:
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    ingest:
      replicas: 1
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    master:
      replicas: 1
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

## Kafka Schema Registry
This release introduces Schema Registry for Kafka, an awesome tool that provides a centralized repository and validating schemas for kafka topic messages and for serialization and deserialization of the data.It plays a critical role in ensuring that data formats are consistent and compatible over time, especially in environments where multiple producers and consumers interact with Kafka.
The initial release of Schema Registry is bringing support for Provisioning. You can now enable schema registry for Avro, Protobuf, JSON etc. You can also use this schema registry with Kafka Connect Cluster source/sink connector to serialize and deserialize data.

You can run Schema Registry with `In-memory` and `KafkaSQL` as storage backend in this release.

Let’s assume you have a Kafka cluster `kafka-prod`, provisioned using KubeDB is deployed in a namespace called demo. You can now provision a SchemaRegistry using the following yaml.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: SchemaRegistry
metadata:
  name: schemaregistry
  namespace: demo
spec:
  version: 2.5.11.final
  replicas: 2
  kafkaRef:
    name: kafka-prod
    namespace: demo
  deletionPolicy: WipeOut
```
**New Version support**: `2.5.11.final`

> Note: To run Schema Registry as `In-memory`, you just need to remove `kafkaRef` field from the above yaml.

## Microsoft SQL Server

In this release, we are introducing TLS support for Microsoft SQL Server. By implementing TLS support, Microsoft SQL Server enhances the security of client-to-server encrypted communication. 

With TLS enabled, client applications can securely connect to the Microsoft SQL Server cluster, ensuring that data transmitted between clients and servers remains encrypted and protected from unauthorized access or tampering. This encryption adds an extra layer of security, essential for sensitive data environments where confidentiality and integrity are paramount.

To configure TLS/SSL in Microsoft SQL Server, KubeDB utilizes the cert-manager to issue certificates. So, first, you have to ensure the cluster has the cert-manager installed. To install cert-manager in your cluster, follow the steps [here](https://cert-manager.io/docs/installation/).

To issue a certificate, the following Custom Resource (CR) of cert-manager is used:

**Issuer/ClusterIssuer**: Issuers and ClusterIssuers represent certificate authorities (CAs) that can generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer in a ready condition to attempt to serve the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

**Certificate**: The cert-manager has the concept of Certificates that define the desired `x509` certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/usage/certificate/).

Here’s a sample YAML for TLS-enabled Microsoft SQL Server:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-standalone-tls
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - kubedb
        emailAddresses:
          - abc@appscode.com
      - alias: server
        subject:
          organizations:
            - kubedb
        emailAddresses:
          - abc@appscode.com
    clientTLS: true
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
The users must specify the `spec.tls.issuerRef` field. If user set `spec.tls.clientTLS: true`  then tls enabled SQL Server will be provisioned. The user have to install [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate issuers (/etc/ssl/certs/ca-certificates.crt).



If `tls.clientTLS: false` is specified then tls will not be enabled for SQL Server but the Issuer will be used to configure tls enabled wal-g proxy-server which is required for SQL Server backup restore.
KubeDB uses the issuer or clusterIssuer referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using Issuer/ClusterIssuers specification. These certificate secrets includes `ca.crt`, `tls.crt` and `tls.key` etc. and are used to configure Microsoft SQL Server


## MongoDB
### MongoDBArchiver Shard Support: 
We are pleased to announce that this release includes support for the `MongoDBArchiver` in Sharded MongoDB Cluster environments. This significant enhancement enables Point-in-Time Recovery (PITR) for the Sharded MongoDB Cluster managed by KubeDB, providing the capability to restore data to any specific point in time following a disaster. This constitutes a major feature addition that will greatly benefit users by improving disaster recovery processes and minimizing potential data loss.
### PVCs Backup for Shard: 
We have introduced support for Sharded MongoDB Cluster in the `mongodb-csi-snapshotter` plugin.This enhancement allows users to back up Persistent Volume Claims (PVCs) of their KubeDB-managed Sharded MongoDB Cluster, thereby ensuring greater data protection and ease of recovery.
### Bug Fix: 
Specific components restoression provided in KubeStash Restoression wasn’t working properly. This bug has been fixed in this release.

## Memcached
### Custom Configuration
This release introduces custom configuration for Memcached. By using custom configuration file, you can use KubeDB to run Memcached with custom configuration.
The necessary information required for custom configuration is memcached.conf file which is the Memcached configuration file containing the custom configurations. For custom configuration, you can use YAML like this:

```yaml
apiVersion: v1 
stringData: 
  memcached.conf: | 
    -m 32
    -c 500
kind: Secret 
metadata: 
  name: mc-configuration 
  namespace: demo 
  resourceVersion: "4505"
```

In the above YAML, `-m` is max memory limit to use for object storage & `-c` is max simultaneous connections.

To apply this custom configuration, the Memcached YAML will be like:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: memcached
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  configSecret:
    name: mc-configuration
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
  terminationPolicy: WipeOut
```

### OpsRequest
Memcached Ops Request support has been introduced through this release. Ops Request for Restart, Vertical Scaling, and Reconfiguration have been added.

#### Restart
Restart ops request is used to perform a smart restart of the Memcached. An example YAML is provided below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-new
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: memcd-quickstart
```

#### Vertical Scaling
Vertical Scaling allows you to vertically scale the Memcached nodes (ie. pods). The necessary information required for vertical scaling, must be provided in the `spec.verticalScaling` field.
An example YAML is provided below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcached-v
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: memcd-quickstart
  verticalScaling:
    memcached:
      resources:
        requests:
          memory: "700Mi"
          cpu: "700m"
        limits:
          memory: "700Mi"
          cpu: "700m"
```

#### Reconfiguration
Reconfiguration allows you to update the configuration through a new secret or apply a config. Users can also remove the custom config using RemoveCustomConfig. The `spec.configuration` field needs to contain the data required for reconfiguration. An example yaml is provided below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: memcd-quickstart
  configuration:
    applyConfig:
      memcached.conf: |
        -m 128
        -c 50
```
## PgBouncer
### Multiple user support:
In this release an user can provide multiple postgres users to connect with pgbouncer. User just need to create secrets which contain `username` & `password`. To apply those secrets into pgbouncer pods the user needs to add some specific labels. An example of secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: quick-postgres
  namespace: <appbinding namespace>
  labels:
    app.kubernetes.io/instance: <database-name>
    app.kubernetes.io/name: postgreses.kubedb.com
stringData:
  password: "<password>"
  username: "<username>"
```
In previous versions if a user made any changes on the secret, it doesn’t reflect on a running pgbouncer pod. But, now if any secret with those specific labels creates/update/delete, it will reflect on running pgbouncer pods via reloading the pgbouncer configuration.

### One database per pgbouncer resource:
Previously there were multiple postgres database servers which the pgbouncer can connect to. But, there was a conflict between them with the same username with different password. To solve this we removed the feature of multiple database servers and made this to connect only one postgres database server.

### HealthCheck:
Health check is configured in this release. Now it can do write check and it can check every pgbouncer pod if it is healthy or not.

## Pgpool
### OpsRequest
In this release, we have introduced support for Pgpool Ops Requests. Current Ops Request supports for Pgpool are: Restart, Vertical Scaling, and Reconfigure.
#### Restart
Restart ops request is used to perform a smart restart to Pgpool pods. An example YAML is provided below:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: pgpool
```

#### Vertical Scaling
Vertical Scaling allows to vertically scale pgpool pods. The necessary information for vertical scaling must be provided in the `spec.verticalScaling.node` field. Additionally it can also take `spec.verticalScaling.nodeSelectionPolicy` and  `spec.verticalScaling.topology` fields. An example YAML is provided below:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pgpool
  verticalScaling:
    node:
      resources:
        requests:
          memory: "1200Mi"
          cpu: "1"
        limits:
          memory: "1800Mi"
          cpu: "2"
```

#### Reconfigure
Reconfigure allows you to reconfigure Pgpool with new configuration via `spec.configuration.configSecret` or `spec.configuration.applyConfig` can be used if you want to apply changes or add some new configuration in addition to currently used configuration. Also, you can remove currently used configuration and use the default configuration by using `spec.configuration.removeCustomConfig`. An example YAML is provided below:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pgpool
  configuration:
    applyConfig:
      pgpool.conf: |-
        memory_cache_enabled = on
```

### AutoScaler
In this release, we are introducing PgpoolAutoscaler, a Kubernetes Custom Resource Definition (CRD) that supports auto scaling for Pgpool. This allows you to configure auto scaling for Pgpool based on cpu, memory and nodeTopology. An example YAML is provided below:
```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PgpoolAutoscaler
metadata:
  name: pgpool-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: pgpool
  compute:
    pgpool:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 3
        memory: 2Gi
      maxAllowed:
        cpu: 5
        memory: 5Gi
      resourceDiffPercentage: 5
      controlledResources: ["cpu", "memory"]
```
### Custom configuration via config secret
Now we can create a secret with `pgpool.conf` as the key and  refer this secret to Pgpool for use. KubeDB operator will validate this secret and merge with default configuration and use it for Pgpool. Just give the reference for the config secret to `spec.configSecret` and the KubeDB operator will do the rest for you. We can also give custom configuration via `spec.initConfig` as previous versions.
### Pod Disruption Budget (PDB)
We are now automatically creating a Pod Disruption Budgets (PDB) for Pgpool upon creating PetSet for it. A PDB helps ensure the availability of Pgpool by limiting the number of pods that can be down simultaneously due to voluntary disruptions (e.g., maintenance or upgrades).
### New service port for pcp user
Now you can use pcp users through the primary service to do administrative level tasks for Pgpool. By default the port is `9595`.

## Postgres
In this release, we have added a PostgreSQL extension for the Apache AGE Graph Database in PostgreSQL version 15. This extension will be supported on Linux Alpine and Debian-based PostgreSQL images. 
Please refer to these links for Apache AGE extension trial. 

Alpine: https://github.com/kubedb/postgres-docker/tree/release-15.5-alpine-age

Debian: https://github.com/kubedb/postgres-docker/tree/release-15.5-bookworm-age

Additionally, Postgres Remote Replica will be supported for PostgreSQL major versions 13 and 14. As a reminder, we already have support for major PostgresSQL versions 15 and 16.
Please refer to this [link](https://kubedb.com/docs/v2024.4.27/guides/postgres/remote-replica/remotereplica/) to know more about Postgres Remote Replica concepts and example.

## RabbitMQ
### OpsRequest

This release is going to introduce more OpsRequest for RabbitMQ clusters. The last release included RabbitMQ OpsRequests for Restart, Vertical Scaling, and Volume Expansion. This release brings support for Horizontal Scaling, Update Version, Reconfigurations, and ReconfigureTLS. Here’s a sample YAML for Upgrading RabbitMQ `v3.12.12` cluster named **rabbitmq** in demo namespace to `v3.13.2` - 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rabbitmq-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: rabbitmq
  updateVersion:
    targetVersion: 3.13.2
```

Here’s another YAML for horizontally scaling up a RabbitMQ cluster.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rabbitmq-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: rabbitmq
  horizontalScaling:
    node: 5
```

#### TLS support: 
Now you can deploy RabbitMQ clusters with TLS enabled. This will let publishers and consumers communicate with TLS (SSL) listener on the `5671` port via `AMQP` protocol. RabbitMQ peers will also communicate with TLS-encrypted messages. KubeDB only provides TLS support via cert-manager issued certificates. So, you need to have cert-manager installed first. Create either a Issuer or ClusterIssuer representing certificate authorities (CAs) that can generate signed certificates by honoring certificate signing requests. Here’s a sample YAML of RabbitMQ cluster with enabled TLS. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: rabbitmq-ca-issuer
  storageType: Durable
  deletionPolicy: Delete
```

**New Version support**: `3.13.2`.

## SingleStore

### OpsRequest
In this release, we have introduced support for SingleStore Ops Requests. Initially, Ops Requests for Restart, Vertical Scaling, Volume Expansion, and Reconfiguration have been added for both clustering and standalone modes.

#### Vertical Scaling
Vertical Scaling allows you to vertically scale the SingleStore nodes (i.e., pods). The necessary information for vertical scaling must be provided in the `spec.verticalScaling.(aggregator/leaf/node/coordinator)` field.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-vscale
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sdb-sample
  verticalScaling:
    aggregator:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"

```
#### Volume Expansion
Volume Expansion allows you to expand the storage of the SingleStore nodes (i.e., pods). The necessary information for volume expansion must be provided in the `spec.volumeExpansion.(aggregator/leaf/node)` field.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-volume-ops
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: sdb-sample
  volumeExpansion:
    mode: "Offline"
    aggregator: 10Gi
    leaf: 20Gi
```
#### Reconfiguration
Reconfiguration allows you to update the configuration through a new secret or apply a config. Users can also remove the custom config using RemoveCustomConfig. The necessary information for reconfiguration must be provided in the `spec.configuration.(aggregator/leaf/node)` field.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Configuration
  databaseRef:
    name: sdb-standalone
  configuration:
    node:
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 350
```
### Autoscaler

In this release, we are also introducing the SinglestoreAutoscaler, a Kubernetes Custom Resource Definition (CRD) that supports autoscaling for SingleStore. This CRD allows you to configure autoscaling for SingleStore compute resources and storage in a declarative, Kubernetes-native manner.

Deploying the SingleStore Autoscaler

To deploy an Autoscaler for a KubeDB-managed SingleStore cluster, you can use the following YAML configuration:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SinglestoreAutoscaler
metadata:
  name: sdb-auto
  namespace: demo
spec:
  databaseRef:
    name: sdb-sample
  storage:
    aggregator:
      trigger: "On"
      usageThreshold: 40
      scalingThreshold: 50
      expansionMode: "Offline"
      upperBound: "100Gi"
  compute:
    leaf:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 900m
        memory: 3000Mi
      maxAllowed:
        cpu: 2000m
        memory: 6Gi
      controlledResources: ["cpu", "memory"]
      resourceDiffPercentage: 10
```

### Pod Disruption Budget (PDB)

In this release, we have added support for Pod Disruption Budgets (PDB) for SingleStore. A PDB helps ensure the availability of your SingleStore application by limiting the number of pods that can be down simultaneously due to voluntary disruptions (e.g., maintenance or upgrades).

## ZooKeeper

### Pod Disruption Budget (PDB)

In this release, we have added support for Pod Disruption Budgets (PDB) for ZooKeeper. A PDB helps ensure the availability of your ZooKeeper application by limiting the number of pods that can be down simultaneously due to voluntary disruptions (e.g., maintenance or upgrades).


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2024.6.4/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2024.6.4/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
