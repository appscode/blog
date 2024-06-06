---
title: Introducing KubeStash v2024.6.4
date: "2024-06-05"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- disaster-recovery
- kubernetes
- kubestash
- manifest-backup
- manifest-restore
- restore
---

We are pleased to announce the release of [KubeStash v2024.6.4](https://kubestash.com/docs/v2024.6.4/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.6.4/README.md). In this post, we'll highlight the key updates.

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### Workload Manifest Backup & Restore

We've introduced functionality to backup and restore workload (`Deployment`/`StatefulSet`/`DaemonSet`) manifests within the `workload-addon`. The new `mainfest-backup` and `manifest-restore` tasks of the `workload-addon` enable backing up and restoring of workload manifests, including their associated volumes, service account (if used), and service (in case of `StatefulSet`).

Here is an example of `BackupConfiguration` that takes backup of a `StatefulSet` manifests:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-backup
  namespace: demo
spec:
  target:
    apiGroup: apps
    kind: StatefulSet
    name: sample-sts
    namespace: demo
  backends:
  - name: s3-storage
    storageRef:
      namespace: demo
      name: s3-storage
    retentionPolicy:
      name: demo-retention
      namespace: demo
  sessions:
  - name: workload-backup
    sessionHistoryLimit: 3
    scheduler: 
      schedule: "*/5 * * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: demo-storage
      backend: s3-storage
      directory: /sts
      encryptionSecret:
        name: encry-secret
        namespace: demo
    addon:
      name: workload-addon
      tasks:
      - name: manifest-backup
    retryConfig:
      maxRetry: 2
      delay: 1m
```

Here is the example of `RestoreSession` that restores the `StatefulSet` manifests:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-restore
  namespace: demo
spec:
  manifestOptions:
    restoreNamespace: dev
  dataSource:
    repository: demo-storage
    snapshot: latest
    encryptionSecret:
      name: encry-secret
      namespace: demo
  addon:
    name: workload-addon
    tasks:
    - name: manifest-restore
```

Here, we can configure in which namespace we want to restore our workload by providing the namespace in `spec.manifestOptions.restoreNamespace`.

#### Unified Manifest and Data Recovery

Now you can restore the manifests and data of workloads or databases (`MySQL`, `MariaDB`, `MongoDB`, `PostgreSQL`) by creating just one `RestoreSession`. KubeStash will deploy the workload or database in a cluster from the backed-up manifests and then restore data into it from the backed-up data. For this, you need to configure the manifest backup and data backup in the same session of a `BackupConfiguration`. This enables KubeStash to restore the manifests and data using a single `Snapshot`.

Here is an example of a `BackupConfiguration` for backing up both the manifests and data of a `StatefulSet` in the same session:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-backup
  namespace: demo
spec:
  target:
    apiGroup: apps
    kind: StatefulSet
    name: sample-sts
    namespace: demo
  backends:
  - name: s3-storage
    storageRef:
      namespace: demo
      name: s3-storage
    retentionPolicy:
      name: demo-retention
      namespace: demo
  sessions:
  - name: workload-backup
    sessionHistoryLimit: 3
    scheduler:
      schedule: "*/5 * * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: demo-storage
      backend: s3-storage
      directory: /sts
      encryptionSecret:
        name: encry-secret
        namespace: demo
    addon:
      name: workload-addon
      tasks:
      - name: manifest-backup
      - name: logical-backup
        params:
          paths: /source/data
          exclude: /source/data/tmp
    retryConfig:
      maxRetry: 2
      delay: 1m
```

Here's an example of a `RestoreSession` that restores `StatefulSet` manifests and then restores its data:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-restore
  namespace: demo
spec:
  manifestOptions:
    restoreNamespace: dev
  dataSource:
    repository: demo-storage
    snapshot: latest
    encryptionSecret:
      name: encry-secret
      namespace: demo
  addon:
    name: workload-addon
    tasks:
    - name: manifest-restore
    - name: logical-backup-restore
```

In this case, a helper `RestoreSession` will be created in the same namespace of the applied `RestoreSession` to restore the data of the workload.

Here is an example of a `BackupConfiguration` for backing up both the manifests and data (dump) of a `MySQL` database in the same session:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: mysql-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MySQL
    namespace: demo
    name: mysql-backup
  sessions:
  - name: frequent-backup
    sessionHistoryLimit: 3
    scheduler:
      schedule: "*/5 * * * *"
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: mysql-storage
      directory: /mysql
      encryptionSecret:
        name: encry-secret
        namespace: demo
    addon:
      name: mysql-addon
      tasks:
      - name: logical-backup
      - name: manifest-backup
    retryConfig:
      maxRetry: 2
      delay: 1m
```

Here's an example of a `RestoreSession` that restores `MySQL` manifests and then restores its data:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: mysql-restore
  namespace: demo
spec:
  manifestOptions:
    restoreNamespace: dev
    mySQL:
      db: true
  dataSource:
    snapshot: latest
    repository: mysql-storage
    encryptionSecret:
      name: encry-secret
      namespace: demo
  addon:
    name: mysql-addon
    tasks:
    - name: logical-backup-restore
    - name: manifest-restore
```

 > Ensure that you have set the `spec.manifestOptions.mySQL.db` to `true`, as the restoration of the `MySQL` object manifest relies on this field.

### Improvements & Bug Fixes

- In this release, we've addressed an issue with `MongoDB` restoration. Previously, even when specifying components in the `RestoreSession`, all components present in the `Snapshot` were restored. Now, if the components are provided in the `RestoreSession` that will be restored, otherwise all the components in `Snapshot` will be restored.
- In this release, we've resolved an RBAC issue where no `RoleBinding` was created in the manifest restore namespace.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).