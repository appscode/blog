---
title: Announcing KubeDB v2023.12.28
date: "2023-12-28"
weight: 14
authors:
- Arnob kumar saha
tags:
- autoscaler
- database
- day-2-operations
- elasticsearch
- kafka
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- nodepool
- nodetopology
- percona-xtradb
- postgresql
- prometheus
- proxysql
- redis
- scheduler
- vertical-scaling
---

We are pleased to announce the release of [KubeDB v2023.12.28](https://kubedb.com/docs/v2023.12.28/setup/). This release was mainly focused on improving the kubedb-autoscaler feature. We have also included point-in-time-recovery feature for MySQL. This post lists all the changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.12.28/README.md). Let’s see the changes done in this release.

## Improving KubeDB Autoscaler

Here is an overall workflow of the kubedb-autoscaler to better understand the problem, we solved in this release.

- The autoscaler operator watches the usages of compute resources (cpu, memory) & storage resources, and generates OpsRequest CR to automatically change the resources.
- The ops-manager operator watches the created VerticalOpsRequest for compute resources, update db's statefulsets & evict the db pods following pod disruption budget.
- k8s scheduler sees the updated resource requests in those pods, & finds an appropriate node for scheduling.
- If k8s scheduler can't find an appropriate node, cloud provider's cluster autoscaler (if enabled) scales one of the nodepools to make room for that pod.

This procedure works fine while up-scaling the compute resources. Some nodes from bigger nodepools will be automatically created by the cluster autoscaler whenever some scheduling issues occur.
But this procedure becomes very resource-intensive while down-scaling the compute resources. As the k8s scheduler sees some big nodes are already available for scheduling, it does not choose a smaller node where these down-scaled pods could be easily running.

So to solve this issue, we need a way so that we can forcefully schedule those smaller pods into smaller nodepools.  We have introduced a new CRD, called `NodeTopology` to achieve it.
Here is an example NodeTopology CR:

```yaml
apiVersion: node.k8s.appscode.com/v1alpha1
kind: NodeTopology
metadata:
  name: gke-n1-standard
spec:
  nodeSelectionPolicy: Taint
  topologyKey: nodepool_type
  nodeGroups:
  - allocatable:
      cpu: 940m
      memory: 2.56Gi
    topologyValue: n1-standard-1
  - allocatable:
      cpu: 1930m
      memory: 5.48Gi
    topologyValue: n1-standard-2
  - allocatable:
      cpu: 3920m
      memory: 12.07Gi
    topologyValue: n1-standard-4
  - allocatable:
      cpu: 7910m
      memory: 25.89Gi
    topologyValue: n1-standard-8
  - allocatable:
      cpu: 15890m
      memory: 53.57Gi
    topologyValue: n1-standard-16
  - allocatable:
      cpu: 31850m
      memory: 109.03Gi
    topologyValue: n1-standard-32
  - allocatable:
      cpu: 63770m
      memory: 224.45Gi
    topologyValue: n1-standard-64
```

It is a cluster-scoped resource. It supports two types of nodeSelectionPolicy : `LabelSelector` and `Taint`. Here is the general rule to choose between these two.

If you want to run the database pods in some dedicated nodes, and don't want to allow any other pods to be scheduled there, the `Taint` policy is appropriate for you. For other general cases, use `LabelSelector`.

It is also possible to schedule different types of db pods into different nodepools. Here is an example `MongoDB` CR yaml :
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-database
  namespace: demo
spec:
  version: "4.4.26"
  terminationPolicy: WipeOut
  replicas: 3
  replicaSet:
    name: "rs"
  podTemplate:
    spec:
      nodeSelector:
        app: "kubedb"
        instance: "mongodb"
        component: "mg-database"
      tolerations:
      - key: nodepool_type
        value: n1-standard-2
        effect: NoSchedule
      - key: app
        value: kubedb
        effect: NoSchedule
      - key: instance
        value: mongodb
        effect: NoSchedule
      - key: component
        value: mg-database
        effect: NoSchedule
      resources:
        requests:
          "cpu": "1435m"
          "memory": "4.02Gi"
        limits:
          "cpu": "1930m"
          "memory": "5.48Gi"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

IMPORTANT : The node pool sizes, the starting resource requests, and the auto scaler configuration must be carefully choreographed for optimal behavior.

- Database's initial resource request will be in the mid-point of the extra resource this nodepool provides.  More specifically , It will be, (current nodepool's allocatable - immediate lower nodepool's allocatable)/2 + immediate lower nodepool's allocatable.
- Database's initial resource limit should match the initial nodepool's allocatable resources.
- Autoscaler CR's minAllowed will be database's initial request * 0.9 if this current nodePool is the smallest nodePool. Otherwise, You have to calculate, what could be the initial resource request if this db was provisioned in the smallest nodepool , and then multiply it with 0.9.
- Autoscaler CR's maxAllowed will be biggest nodePool's allocatable resources.

For example, in the above db yaml, requested cpu = (1930m - 940m)/2 + 940m = 1435m. And it's cpu limit is the allocatable cpu of `n1-standard-2` pool, which is 1930m.

In the below autoscaler yaml, minAllowed cpu = (940m / 2) * 0.9 = 423m, memory = (2.56Gi / 2) * 0.9 = 1.15Gi. It's maxAllowed cpu & memory will be biggest nodepool's allocatable. So, 63770m & 224.45Gi respectively.

You can find a list of pre-calculated values in [this spreadsheet](https://docs.google.com/spreadsheets/d/1U7H9rocT03UqLSof9a_YaFObX34WTYIaqU3wKbhzP_0/edit#gid=1036880772). 

Lastly, for autoscaling, all we need is to specify the name of the nodeTopology in the autoscaler yaml.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-database
  namespace: demo
spec:
  databaseRef:
    name: mg-database
  opsRequestOptions:
    timeout: 10m
    apply: IfReady
  compute:
    replicaSet:
#      podLifeTimeThreshold: 15m
#      resourceDiffPercentage: 50
      trigger: "On"
      minAllowed:  # By considering `n1-standard-2` as your smallest db nodepool
        cpu: "1292m"
        memory: "3.62Gi"
#      minAllowed:  # By considering `n1-standard-1` as your smallest nodepool
#        cpu: "423m"
#        memory: "1.15Gi"
      maxAllowed:
        cpu: "63770m"
        memory: "224.45Gi"
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
    nodeTopology:
      name: gke-n1-standard
```

Now, kubedb-autoscaler operator will decide what is the minimum node-configuration for the scaled (up or down) pods to be scheduled. And it will create the `VerticalScale` opsRequest specifying the tolerations so that the pods are scheduled on the desired nodepool.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-jghfjd
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-database
  verticalScaling:
    replicaSet:
      resources:
        requests:
          memory: "8.78Gi"
          cpu: "2925m"
        limits:
          memory: "12.07Gi"
          cpu: "3920m"
      nodeSelectionPolicy: Taint
      topology:
        key: nodepool_type
        value: n1-standard-4
```

## MySQL Archiver

This feature supports continuous archiving of a MySQL database. You can also do point-in-time recovery (PITR) restoration of the database at any point.

To use this feature, You need [KubeStash](https://kubestash.com/) installed in your cluster. KubeStash (aka Stash 2.0) is a ground up rewrite of [Stash](https://stash.run/docs/v2023.10.9/welcome/) with various improvements planned. KubeStash works with any existing KubeDB or Stash license key. To use continuous archiving feature, We have introduced a CRD also in KubeDB side, named `MySQLArchiver`.

Here is all the details of using [MySQLArchiver](https://kubedb.com/docs/v2023.12.11/guides/mysql/pitr/archiver/).
In short, you need to create the following resources:
- `BackupStorage` which refers a cloud storage backend (like s3, gcs etc.) you prefer.
- `RetentionPolicy` allows you to set how long you'd like to retain the backup data.
- `Secret` holds restic password which will be used to encrypt the backup snapshots.
- `VolumeSnapshotClass` which holds the csi-driver information which is responsible for taking VolumeSnapshots. This is vendor specific.
- `MySQLArchiver` which holds all of these metadata information.

NB: All the archiver related yamls are available in this [git repository](https://github.com/kubedb/archiver-demo/tree/master/mysql).

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MySQLArchiver
metadata:
  name: mysqlarchiver-sample
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
    name: mysql-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-snapshot-vsc"
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
      name: "linode-storage"
      namespace: "demo"
```

Now after creating this archiver CR, if we create a MySQL with `archiver: "true"` label, in the same namespace (as per the double-optin configured in `.spec.databases` field), The KubeDB operator will start doing 3 separate things:
- Create 2 `Repository` with convention `<db-name>-full` & `<db-name>-manifest`.
- Take full backup in every day at 3:30 (`.spec.fullBackup.scheduler`) to `<db-name>-full` repository.
- Take manifest backup in every day at 3:30 (`.spec.manifestBackup.scheduler`) to `<db-name>-manifest` repository.
- Start syncing mysql wal files to the directory `.spec.backupStorage.subDir/<db-namespace>/<db-name>/binlog`.


For point-in-time-recovery, all you need is to set the repository names & set a `recoveryTimestamp` in `mysql.spec.init.archiver` section.

Here is an example of `MySQL` CR for point-in-time-recovery.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: restore-mysql
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mysql-repository
        namespace: demo
      recoveryTimestamp: "2023-12-28T17:10:54Z"
  version: "8.2.0"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  terminationPolicy: WipeOut
```


## Postgres Archiver

We supports point in time recovery feature for postgres from KubeDB v2023.12.11 for s3 backends. In this release, we have added some other backends support, namely gcs, azure & local nfs.

To use these backends, you have to configure two things `BackupStorage` & `VolumeSnapshotClass`. 

**Example YAMLs for azure:**

```yaml 
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: azure-storage
  namespace: demo
spec:
  storage:
    provider: azure
    azure:
      storageAccount: storageAccountName
      container: container
      prefix: pg
      secret: azure-secret  # this secret holds the AZURE_ACCOUNT_KEY info
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut 
  
---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: az-vsc
driver: disk.csi.azure.com
deletionPolicy: Delete
```

NB: All the archiver related yamls are available in this [git repository](https://github.com/kubedb/archiver-demo/tree/master/postgres).

**Example YAMLs for GCS:**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: gcs-storage
  namespace: demo
spec:
  storage:
    provider: gcs
    gcs:
      bucket: kubestash-qa
      prefix: pg
      secret: gcs-secret # This secret holds the GOOGLE_PROJECT_ID & GOOGLE_SERVICE_ACCOUNT_JSON_KEY info
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut 

---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: gke-vsc
driver: pd.csi.storage.gke.io
deletionPolicy: Delete
```

**Example YAMLs for NFS:**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: local-storage
  namespace: demo
spec:
  storage:
    provider: local
    local:
      mountPath: /pg/walg
      nfs:
        server: "use the server address here"
        path: "use the shared path here"
  usagePolicy:
    allowedNamespaces:
      from: All
  default: false
  deletionPolicy: WipeOut
  runtimeSettings:
    pod:
      securityContext:
        fsGroup: 70
        runAsUser: 70
```

And lastly, you need to specify them in the `PostgresArchiver` yaml like below :

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: pg-archiver
  namespace: demo
spec:
  pause: false
  retentionPolicy:
    name: postgres-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 70
          runAsGroup: 70
          fsGroup: 70
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-vsc" # "gke-vsc" # "az-vsc"  # Set accordingly
    scheduler:
      successfulJobsHistoryLimit: 2
      failedJobsHistoryLimit: 2
      schedule: "30 3 * * *"
    sessionHistoryLimit: 3
  backupStorage:
    ref:
      name: "s3-storage"
      namespace: "demo"
```

We are setting them in the `spec.fullBackup.task.params.volumeSnapshotClassName` & `spec.backupStorage.ref` fields.


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.12.28/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.12.28/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
