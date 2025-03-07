---
title: Announcing KubeDB v2023.12.11
date: "2023-12-11"
weight: 14
authors:
- Arnob kumar saha
tags:
- archiver
- cloud-native
- crossplane
- dashboard
- database
- day-2-operations
- elasticsearch
- git-sync
- kafka
- kubedb
- kubedb-cli
- kubernetes
- kubestash
- mariadb
- mongodb
- mysql
- non-root
- opensearch
- percona
- percona-xtradb
- pgbouncer
- postgresql
- prometheus
- proxysql
- redis
- security
- walg
---

We are pleased to announce the release of [KubeDB v2023.12.11](https://kubedb.com/docs/v2023.12.11/setup/). This release contains some major features like archiver, using non-root users, git-sync, crossplane support etc. This post lists all the major changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.12.11/README.md). Let’s see the changes done in this release.


## Non-root user
In our prior releases, all the containers (db container, init-docker, metrics-exporter, other sidecars) were run with the root user. This was a big security concern for some of our users. It is also very important in cases like preventing Privilege escalations, restrict the behavior of pods, restrict certain kernel-level operations etc. We have focused on this issue in this release, & made all of our docker images run as non-root user.

We also enforce it from the kubernetes perspective & set the `securityContext` by default. So that the containers abide by the rules of `restricted` [PodSecurityStandards](https://kubernetes.io/docs/concepts/security/pod-security-standards/). This change is common for all of our supported databases. One exception is PostgreSQL. PostgresQL uses non-root `postgres` user but still uses additional capabilities. This will be further investigates in future releases.


## MongoDB

### MongoDB Archiver

This feature supports continuous archiving of a MongoDB database by syncing oplog to an object storage. You can also do point-in-time recovery (PITR) restoration of the database at any point. 
To use this feature, You need [KubeStash](https://kubedb.com/docs) installed in your cluster. KubeStash (aka Stash 2.0) is a ground up rewrite of [Stash](https://stash.run/docs/v2023.10.9/welcome/) with various improvements planned. KubeStash works with any existing KubeDB or Stash license key. To use continuous archiving feature, We have introduced a CRD also in KubeDB side, named `MongoDBArchiver`.


Here is all the details of using [MongoDB Archiver](https://kubedb.com/docs/v2023.12.11/guides/mongodb/pitr/pitr/).
In short, You need to create a
- `BackupStorage` which refers a cloud storage backend (like s3, gcs etc.) you prefer.
- `RetentionPolicy` allows you to set how long you'd like to retain the backed up data.
- `encryption-secret` which will be used for encryption before uploading the backed-up data into cloud.
- `VolumeSnapshotClass` which holds the csi-driver information which is responsible for taking VolumeSnapshots. This is vendor specific.
- `MongoDBArchiver` which holds all of these metadata information. 

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MongoDBArchiver
metadata:
  name: mongodbarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: "Same"
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: mongodb-retention-policy
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: VolumeSnapshotter
    task:
      params:
        volumeSnapshotClassName: gke-vsc 
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/50 * * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/2 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: gcs-storage
      namespace: demo     
```

Now after creating this archiver CR, if we create a MongoDB with `archiver: "true"` label, in the same namespace (as per the double-optin configured in `.spec.databases` field), The KubeDB operator will start doing 3 separate things:
- Create 2 `Repository` with convention `<db-name>-full` & `<db-name>-manifest`.
- Take full back-up in every 50 minute (`.spec.fullBackup.scheduler`) to `<db-name>-full` repository.
- Take manifest back-up in every 2 minute (`.spec.manifestBackup.scheduler`) to `<db-name>-manifest`.
- Start syncing mongodb oplogs to `<db-name>-full` in a directory named `oplog`. 


For point-in-time-recovery, all you need is to set the repository names & set a recoveryTimestamp in `mongodb.spec.init.archiver` section.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs-restored
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "rs"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "2023-12-13T09:35:30Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mg-rs-full
        namespace: demo
      manifestRepository:
        name: mg-rs-manifest
        namespace: demo
  terminationPolicy: WipeOut
```

KubeDB Operator will create a PVC with the VolumeSnapshot reference of the last full-backup (which is before this referred timestamp). And then apply the oplogs from time interval "Last VolumeSnapshot time" to "recoveryTimestamp".

### Git Sync
We have added a new feature now you can initialize mongodb from the public/private git repository.
Here’s a quick example of how to configure it. Here we are going to create a mongodb replicaset with some initial data from  [git-sync-demo](https://github.com/kubedb/git-sync-demo.git) repo.

**From Public Registry:**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
 name: rs
 namespace: demo
spec:
 init:
   script:
     scriptPath: "current"
     git:
       args:
       - --repo=https://github.com/kubedb/git-sync-demo.git
       - --depth=1
       - --period=60s
       - --link=current
       - --root=/git
       # terminate after successful sync
       - --one-time 
 version: "4.4.26"
 replicas: 3
 replicaset:
   name: rs0
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
*From Private Registry:***

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
 name: rs
 namespace: demo
spec:
 init:
   script:
     scriptPath: "current"
     git:
       args:
       # use --ssh for private repository
       - --ssh
       - --repo=git@github.com:kubedb/git-sync-demo.git
       - --depth=1
       - --period=60s
       - --link=current
       - --root=/git
       # terminate after successful sync
       - --one-time
       authSecret:
         name: git-creds
       # run as git sync user 
       securityContext:
         runAsUser: 65533  
 podTemplate:
   spec:
     # permission for reading ssh key
     securityContext:
      fsGroup: 65533
 version: "4.4.26"
 replicas: 3
 replicaset:
   name: "rs0"
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

This example refers to initialization from a private git repository
`.spec.init.git.args` represents the arguments required to represent the git repository and its actions. You can find details at [git_syc_docs](https://github.com/kubernetes/git-sync/blob/master/README.md)

`.spec.init.git.authSecret` holds  the necessary information to pull from the private repository
You have to provide a secret with the `id_rsa` and `githubkwonhosts`
You can find detailed information at [git_sync_docs](https://github.com/kubernetes/git-sync/blob/master/docs/ssh.md).
If you are using different authentication mechanism for your git repository, please consult the documentation for [git-sync](https://github.com/kubernetes/git-sync/tree/master/docs) project.

`.spec.init.git.securityContext.runAsUser`  the init container git_sync run with user `65533`.

`.spec.podTemplate.Spec.securityContext.fsGroup` In order to read the ssh key the fsGroup also should be `65533`.

```bash
ssh-keyscan $YOUR_GIT_HOST > /tmp/known_hosts
kubectl create secret generic -n demo git-creds \
   --from-file=ssh=$HOME/.ssh/id_rsa \
   --from-file=known_hosts=/tmp/known_hosts
```


### Version support

We have added some new version support: `4.2.24`, `4.4.26`, `5.0.23`, `6.0.12`. We also made all the older patch versions of these added images deprecated in this release. The docker images for the new versions are built to minimize CVEs if not remove completely.

So, Please apply a MongoDBOpsRequest to update your database in latest patch versions supported.  For example, if the current db version is `4.4.6`, the latest patch version is `4.4.26`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: versionUpd
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mg-rs
  updateVersion:
    targetVersion: 4.4.26
  apply: Always
```


## Postgres
Now you can continuously archive and recover point-in-time  using KubeDB Managed PostgreSQL. Please follow the documentation to try out the new features.

### Arbiter
We use raft consensus algorithm for choosing primary node. Raft uses Quorum based voting system. So if we have an even number of replicas(nodes), there is a high chance of split vote. So it is recommended by raft  to use an odd number of nodes. But many users only wants 2 replicas, a primary node for write/read operation and an extra node for standby/read query purposes. They do not want to have more nodes to keep the cost down. So we are introducing an extra node in the cluster which will solve this issue. This node is called an Arbiter node.

An arbiter node will have a separate Statefulset and a pvc with bare minimum storage(2GB but configurable, however 2GB is enough if your cluster does not have many replicas). It will have a single pod which runs a single container inside, that only votes for the leader election but does not store any database related data.

So if you deploy a Postgres database with an even number of nodes, Arbiter node will be deployed with it automatically.

Postgres Arbiter Ops Request Support
Added support for volume expansion and vertical scaling of Arbiter node.

Primary failover in ops request
Before postgres restart ops request, we do a failover upon the restart of primary pod.


### Postgres Archiver

This feature supports continuous archiving of a Postgres database. You can also do point-in-time recovery (PITR) restoration of the database at any point.

To use this feature, You need [KubeStash](https://kubedb.com/docs) installed in your cluster. KubeStash (aka Stash 2.0) is a ground up rewrite of [Stash](https://stash.run/docs/v2023.10.9/welcome/) with various improvements planned. KubeStash works with any existing KubeDB or Stash license key. To use continuous archiving feature, We have introduced a CRD also in KubeDB side, named `PostgresArchiver`.

Here is all the details of using [PostgresArchiver](https://kubedb.com/docs/v2023.12.11/guides/postgres/pitr/archiver/).
In short, You need to create an
- `BackupStorage` which refers a cloud storage backend (like s3, gcs etc.) you prefer.
- `RetentionPolicy` allows you to set how long you'd like to retain the backup data.
- encryption-secret which will be used for encryption before uploading the backed-up data into cloud.
- `VolumeSnapshotClass` which holds the csi-driver information which is responsible for taking VolumeSnapshots. This is vendor specific.
- `PostgresArchiver` which holds all of these metadata information.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: postgresarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
         kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: postgres-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "standard-csi"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "30 3 * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "30 3 * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "s3-storage"
      namespace: "demo"
```


Now after creating this archiver CR, if we create a Postgres with `archiver: "true"` label, in the same namespace (as per the double-optin configured in `.spec.databases` field), The KubeDB operator will start doing 3 separate things:
- Create 2 `Repository` with convention `<db-name>-full` & `<db-name>-manifest`.
- Take full back-up in every 50 minute (`.spec.fullBackup.scheduler`) to `<db-name>-full` repository.
- Take manifest back-up in every 2 minute (`.spec.manifestBackup.scheduler`) to `<db-name>-manifest`.
- Start syncing postgres wal files to `<db-name>-full` in a directory named `oplog`.


For point-in-time-recovery, all you need is to set the repository names & set a recoveryTimestamp in `postgres.spec.init.archiver` section.

Here is an example of `Postgres` CR for point-in-time-recovery.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: restore-pg
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: demo-pg-repository
        namespace: demo
      manifestRepository:
        name: demo-pg-manifest
        namespace: demo
      recoveryTimestamp: "2023-12-12T13:43:41.300216Z"
  version: "13.6"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

### Version support

We have added support for postgres 16.1 in this release

## Elasticsearch

Elasticsearch uses a `mmapfs` directory by default to store its indices. The default operating system limits on mmap counts is likely to be too low, which may result in out of memory exceptions. In order to bootstrap Elasticsearch successfully, it is necessary to increase the limits by running the following command as `root`:

```bash
sysctl -w vm.max_map_count=262144
```

From this release KubeDB ensures that all database pods will be running as non-root user. But, a single init container runs as root in privileged mode to increase `vm.max_map_count` in kernel settings. We call this `sysctl-init` container. We are continuing with that as default. However, If it is not possible to run a container as root and in privileged mode it is advisable to set `spec.kernelSettings.disableDefaults` to `true` prior to apply Elasticsearch custom resource. In this case you pre-setup `vm.max_map_count` value for your kubernetes nodes. You can also use kubedb `prepare-cluster` helm chart to do this easily.

```
helm upgrade -i prepare-cluster \
  oci://ghcr.io/appscode-charts/prepare-cluster \
  --version v2023.12.11 \
  -n kube-system \
  --set node.sysctls[0].name=vm.max_map_count \
  --set node.sysctls[0].value=262144
```

Here's a sample yaml for deploying elasticsearch cluster that you can deploy ensuring the privileged init container doesn't run before the elasticsearch containers. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-quickstart
  namespace: demo
spec:
  version: xpack-8.11.1
  enableSSL: true
  replicas: 3
  storageType: Durable
  kernelSettings:
    disableDefaults: true
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  terminationPolicy: DoNotTerminate
```

We have worked on providing CVE free images for Elasticsearch. Most of the high and critical vulnerabilities have been removed. 

### Version support

In this release, support for Elasticsearch version `xpack-8.11.1` and Opensearch version `opensearch-2.11.1` have been added.
All the KubeDB supported ElasticSearch and OpenSearch versions have been upgraded to latest patches as they are more stable and includes fewer CVEs. Earlier supported versions with older patches have been marked deprecated. Versions that used either SearchGuard or OpenDistro as security plugins, have also been marked deprecated.

Currently supported Elasticsearch versions are: `xpack-8.11.1`, `xpack-8.8.2`, `xpack-8.6.2`, `xpack-8.5.3`, `xpack-8.2.3`, `xpack-7.17.15`, `xpack-7.16.3`, `xpack-7.14.2`, `xpack-7.13.4`, `xpack-6.8.23`.

And These are the currently supported OpenSearch versions. `opensearch-2.11.1`, `opensearch-2.8.0`, `opensearch-2.5.0`, `opensearch-2.0.1`, `opensearch-1.3.13`, `opensearch-1.2.4`, `opensearch-1.1.0`.

## Kafka

KubeDB managed Apache Kafka went through some major updates and vulnerability fixes in this release. Kafka now runs on `Java 17` instead of `Java 11`. A single headless service is now provisioned by the operator for each kafka cluster. Kafka now Bootstraps with listeners and advertised listeners for brokers, controllers and localhost. User provided listeners/advertised listeners will be simply appended to the default lists.

### Version support
In this release, support for Kafka version `3.6.0` have been added. Here's a sample yaml to deploy a simple 3 broker, 3 controller Apache kafka cluster.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.6.0
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
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  terminationPolicy: DoNotTerminate
```

## MariaDB

### Version support
In this release new version supports have been added. Currently available versions are 10.10.2, 10.10.7, 10.11.2, 10.11.6, 10.4.31, 10.4.32, 10.5.23, 10.6.16, 10.6.4, 11.0.4 and 11.1.3. Version 10.5.8 and 10.4.17 are deprecated from this release.

We have also built these images ourselves so that the non-root mysql use is used by default. This has also reduced the CVEs found in these images.

## MySQL
### Version support

In this release new version supports have been added. Currently available versions are 8.2.0-oracle, 8.1.0-oracle, 8.0.35-debian, 8.0.35-oracle, 8.0.32-oracle, 8.0.31-oracle, 5.7.44-oracle, 5.7.41-oracle. All the other versions have been made deprecated.

We have also built these images ourselves so that the non-root mysql use is used by default. This has also reduced the CVEs found in these images.

## Redis
### Version support

In this release two new version supports have been added namely 7.2.3 & 6.2.14 . 

## Crossplane

KubeDB is now a [Crossplane](https://www.crossplane.io/) distribution for Hyper Clouds. Crossplane connects your Kubernetes cluster to external, non-Kubernetes resources, and allows platform teams to build custom Kubernetes APIs to consume those resources. We have introduced providers for AWS, Azure and GCP.

You need [crossplane](https://docs.crossplane.io/v1.14/) already installed in your cluster. This will allow KubeDB users to provision and manage Cloud provider managed databases in a Kubernetes native way.

## Provider-AWS
### Installation
Install the aws provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-aws \
  oci://ghcr.io/appscode-charts/kubedb-provider-aws \
  --version=v2023.12.11 \
  -n crossplane-system --create-namespace
```

The command deploys a KubeDB AWS provider on the Kubernetes cluster in the default configuration. This will install CRDs representing AWS database services. These CRDs allow you to create AWS database resources inside Kubernetes.

### Setup Provider Config

Create a text file containing the AWS account aws_access_key_id and aws_secret_access_key with the following command.

```bash
echo '
[default]
aws_access_key_id = <your_access_key>
aws_secret_access_key = <your_secret_access_key>
' > aws-credentials.txt
```

Create a Kubernetes secret with the AWS credentials.

```bash
kubectl create secret  generic aws-secret -n crossplane-system --from-file=creds=./aws-credentials.txt
```

Create the ProviderConfig with the following yaml file

```yaml
apiVersion: aws.kubedb.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: aws-secret
      key: creds
      namespace: crossplane-system
```
### Create DynamoDB

```yaml
apiVersion: dynamodb.aws.kubedb.com/v1alpha1
kind: Table
metadata:
  annotations:
    meta.kubedb.com/example-id: dynamodb/v1alpha1/table
  labels:
    testing.kubedb.com/example-name: mytable
  name: mytable
spec:
  forProvider:
    attribute:
    - name: exampleHashKey
      type: S
    hashKey: exampleHashKey
    region: us-east-2
    streamEnabled: true
    streamViewType: NEW_AND_OLD_IMAGES
    billingMode: PAY_PER_REQUEST
```

### Provider AWS also supports

- DocumentDB
- Elasticache
- RDS
  - Mariadb
  - Mysql
  - Postgres

## Provider-Azure
### Installation

Install the Azure provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-azure \
  oci://ghcr.io/appscode-charts/kubedb-provider-azure \
  --version=v2023.12.11 \
  -n crossplane-system --create-namespace
```

The command deploys a KubeDB Azure provider on the Kubernetes cluster in the default configuration. This will install CRDs representing Azure database services. These CRDs allow you to create Azure database resources inside Kubernetes.
### Setup Provider Config
Create an Azure service principal and save it as azure-credentials.json

```bash
az ad sp create-for-rbac --sdk-auth --role Owner --scopes <your-subscription-id>
```

Create a Kubernetes secret with the Azure credentials.

```bash
kubectl create secret generic azure-secret -n crossplane-system --from-file=creds=./azure-credentials.json
```

Create the ProviderConfig with the following yaml file
```yaml
apiVersion: azure.kubedb.com/v1beta1
metadata:
  name: default
kind: ProviderConfig
spec:
  credentials:
    source: Secret
    secretRef:
      name: azure-secret
      key: creds
      namespace: crossplane-system
```

### Create CosmosDB Cassandra Account

```yaml
apiVersion: cosmosdb.azure.kubedb.com/v1alpha1
kind: Account
metadata:
  annotations:
    meta.kubedb.com/example-id: cosmosdb/v1alpha1/cassandratable
  labels:
    testing.kubedb.com/example-name: cassandra
  name: example-cosmosdb-cassandra
spec:
  forProvider:
    capabilities:
      - name: EnableCassandra
    consistencyPolicy:
      - consistencyLevel: Strong
    geoLocation:
      - failoverPriority: 0
        location: West Europe
    location: West Europe
    offerType: Standard
    resourceGroupNameRef:
      name: db
---
apiVersion: azure.kubedb.com/v1alpha1
kind: ResourceGroup
metadata:
  annotations:
    meta.kubedb.com/example-id: cosmosdb/v1alpha1/cosmosdb
  labels:
    testing.kubedb.com/example-name: db
  name: db
spec:
  forProvider:
    location: West Europe
```

### Provider Azure also supports

- MySQL
- MSSQL
- CosmosDB
  - GremlinGraph
  - Mongo
  - SQL
  - Table

## Provider-GCP
### Installation

Install the GCP provider into Kubernetes cluster with helm chart.

```bash
helm upgrade -i kubedb-provider-gcp \
  oci://ghcr.io/appscode-charts/kubedb-provider-gcp \
  --version=v2023.12.11 \
  -n crossplane-system --create-namespace
```

The command deploys a KubeDB GCP provider on the Kubernetes cluster in the default configuration. This will install CRDs representing GCP database services. These CRDs allow you to create GCP database resources inside Kubernetes.

### Setup Provider Config

Generate a GCP service account JSON file and save it as appscode-testing.json.

Create a Kubernetes secret with the GCP credentials.

```bash
kubectl create secret generic gcp-secret -n crossplane-system --from-file=creds=./appscode-testing.json
```

Create the ProviderConfig with the following yaml file

```yaml
apiVersion: gcp.kubedb.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  projectID: <PROJECT_ID>
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: gcp-secret
      key: creds
```

### Create SQL Database Instance

```yaml
apiVersion: sql.gcp.kubedb.com/v1alpha1
kind: DatabaseInstance
metadata:
  annotations:
    meta.kubedb.com/example-id: sql/v1alpha1/databaseinstance
  labels:
    testing.kubedb.com/example-name: example_instance
  name: example-instance
spec:
  forProvider:
    region: "us-central1"
    databaseVersion: "MYSQL_5_7"
    settings:
      - tier: "db-f1-micro"
        diskSize: 20
    deletionProtection: false
  writeConnectionSecretToRef:
    name: example-sql-db-instance-secret
    namespace: crossplane-system
```

### Provider GCP also supports
- Redis
- Spanner

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.12.11/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.12.11/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
