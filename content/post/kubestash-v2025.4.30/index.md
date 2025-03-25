---
title: Introducing KubeStash v2025.4.30
date: "2025-04-30"
weight: 10
authors:
  - Md Anisur Rahman
tags:
  - backup
  - backup-verification
  - disaster-recovery
  - kubernetes
  - kubestash
  - restore
---

We are pleased to announce the release of [KubeStash v2025.4.30](https://kubestash.com/docs/v2025.4.30/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.4.30/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### KubeDB managed Cassandra Backup/Restore

KubeDB now supports backup and restore for Cassandra databases using KubeStash and the Medusa plugin. This enables reliable data protection with cloud storage backends (e.g., S3, GCS).

**Backup and Restore Workflow**

- BackupStorage: Specifies the cloud storage backend.
- RetentionPolicy: Defines how long backup data is retained.
- Secrets: Stores backend access credentials.
- BackupConfiguration: Configures the target database, backend, and addon.
- RestoreSession: Restores data from a specified snapshot.

In below given the example of a `BackupConfiguration` for a `Cassandra` database

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: cass-backup
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
  backends:
    - name: s3-backend
      storageRef:
        namespace: default
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: default
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-cassandra-repo
          backend: s3-backend
          directory: /cassandra
      addon:
        name: cassandra-addon
        tasks:
          - name: logical-backup
        jobTemplate:
          spec:
            serviceAccountName: cluster-resource-reader
```

And for `Cassandra` database below is the example of a `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-cas
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
  dataSource:
    repository: s3-cassandra-repo
    snapshot: s3-cassandra-repo-cass-backup-frequent-backup-1744202600
  addon:
    name: cassandra-addon
    tasks:
      - name: logical-backup-restore
    jobTemplate:
      spec:
        serviceAccountName: cluster-resource-reader
```

### Introduce Workload Manifest Backup/Restore along with RBAC resources

KubeStash now supports backup and restore of Kubernetes `workload manifests`, including `RBAC` resources. This feature allows users to back up and restore not only the data but also the entire configuration of their workloads, ensuring a complete disaster recovery solution.

Here is an example of `BackupConfiguration`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-backupconfig
  namespace: demo
spec:
  target:
    apiGroup: apps
    kind: StatefulSet
    name: my-sts
    namespace: demo
  sessions:
    - name: workload-backup
      addon:
        name: workload-addon
        tasks:
          - name: manifest-backup
            params:
              includeRBACResources: "true"
        jobTemplate:
          spec:
           serviceAccountName: cluster-resource-reader
```

Here is an example of `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-restoresession
  namespace: demo
spec:
  manifestOptions:
    workload:
      restoreNamespace: demo
  addon:
    name: workload-addon
    tasks:
      - name: manifest-restore
        params:
          includeRBACResources: "true"
          overrideResources: "true"
```

Note that you've to set `includeRBACResources` flag to `true` if you want to backup and restore `RBAC` resources like Role, RoleBinding, ClusterRole, ClusterRoleBinding which are being used by your `workload`. Also you've to set `overrideResources` flag to `true` if you want to override the resources while restoring. Both of the flags are by default set to `false`.

### Improvements & Bug Fixes


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

