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

We've introduced functionality to backup and restore workload manifests within the `workload-addon`. These tasks encompass backing up and restoring the manifests of the workload, including their associated volumes, service account (if utilized), and service (only the headless service employed in StatefulSet).

Here is an example of `BackupConfiguration` that will take backup of a StatefulSet manifest:

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

Here is the example of `RestoreSession` that will restore the StatefulSet manifest:

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

Here, we can configure in which namespace we want to restore our workload by providing the namespace name in `spec.manifestOptions.restoreNamespace`.

We can also use these manifest tasks in case of application level backup and restore. To do so, we need to configure `BackupConfiguration` in the following manner:

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

Here is the example of `RestoreSession` that will restore the StatefulSet manifest and also its data:

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

#### MySQL, MariaDB, MongoDB & PostgreSQL Application Level Backup & Restore

We've added support application level backup and restore for `MySQL`, `MariaDB`, `MongoDB` and `PostgreSQL`. We'll be backing up both the database manifests and their associated data. When it comes to restoration, the manifests will be restored first, followed by the data.

Here is an example of `BackupConfiguration` for MySQL:

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

Here is an example of `RestoreSession` for MySQL:

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

Make sure you have provided the `spec.manifestOptions.mySQL.db` as the restore of the MySQL manifest depends on this field.

### Improvements & Bug Fixes

- In this release, we've addressed an issue with `MongoDB` restoration. Previously, even when specifying components in the RestoreSession, all components present in the snapshot were restored. Now, if the components are provided in the RestoreSession that will be restored, otherwise all the components in snapshot will be restored.
- In this release, we've resolved an RBAC issue where no `RoleBinding` was created in the manifest restore namespace.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).