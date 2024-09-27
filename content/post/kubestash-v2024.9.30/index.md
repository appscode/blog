---
title: Introducing KubeStash v2024.9.30
date: "2024-09-30"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- disaster-recovery
- druid
- kubernetes
- kubestash
- manifest-backup
- manifest-restore
- singlestore
- restore
---

We are very excited to announce the release of [KubeStash v2024.9.30](https://kubestash.com/docs/v2024.9.30/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.9.30/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### KubeDB Managed Druid Database Backup & Restore

We've introduced functionality to backup and restore [KubeDB managed Druid](https://kubedb.com/docs/v2024.9.30/guides/druid/) database. You can take both Logical and Manifest backup of druid clusters separately or both together which we call Application Level backup.

Here is an example of `BackupConfiguration` that takes Application level backup of a `Druid`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-druid-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Druid
    name: sample-druid
    namespace: demo
  backends:
  - name: gcs-backend
    storageRef:
      name: gcs-storage
      namespace: demo
    retentionPolicy:
      name: demo-retention
      namespace: demo
  sessions:
  - name: frequent-backup
    scheduler:
      schedule: "*/5 * * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: gcs-druid-repo
      backend: gcs-backend
      directory: /druid
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
    addon:
      name: druid-addon
      tasks:
      - name: manifest-backup
      - name: mysql-metadata-storage-backup
```

Here, `mysql-metadata-storage-backup` needs to be replaced with `postgres-metadata-storage-backup` if `PostgreSQL` is used as metadata storage.

Here is the example of `RestoreSession` that restores the `Druid`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-druid
  namespace: demo
spec:
  manifestOptions:
    druid:
      restoreNamespace: dev
      dbName: restored-druid
  dataSource:
    repository: gcs-druid-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: druid-addon
    tasks:
    - name: mysql-metadata-storage-restore
    - name: manifest-restore
```

#### KubeDB Managed SingleStore Manifest Backup & Restore

KubeStash now supports manifest backup and restore for [KubeDB managed SingleStore](https://kubedb.com/docs/v2024.9.30/guides/singlestore/).

Here is an example of a `BackupConfiguration` for `SingleStore` manifest backup:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-singlestore-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Singlestore
    name: sample-singlestore
    namespace: demo
  backends:
  - name: gcs-backend
    storageRef:
      namespace: demo
      name: gcs-storage
    retentionPolicy:
      name: demo-retention
      namespace: demo
  sessions:
  - name: frequent-backup
    scheduler:
      schedule: "*/5 * * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: gcs-singlestore-repo
      backend: gcs-backend
      directory: /ss
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
    addon:
      name: singlestore-addon
      tasks:
      - name: manifest-backup
```

Here's an example of a `RestoreSession` that restores the manifest of `SingleStore`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-singlestore-restore
  namespace: demo
spec:
  manifestOptions:
    singlestore:
      restoreNamespace: dev
  dataSource:
    repository: gcs-singlestore-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: singlestore-addon
    tasks:
    - name: manifest-restore
```

#### Support for NetworkPolicy
We've added support for [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) in the release. Now users can pass `--set global.networkPolicy.enabled=true` while installing KubeStash. The required Network policies for operator will be created as part of the release process.

### Improvements & Bug Fixes

#### Updated Declarative API 
We’ve updated the declarative API for `BackupConfiguration` and `RestoreSession` in KubeStash. The `spec.sessions[*].timeout` field has been removed from `BackupConfiguration` and replaced with `spec.sessions[*].backupTimeout`, which sets the maximum duration for a backup. If the backup doesn’t finish within this time, it will be marked as failed. By default, no backup timeout is set.

Similarly, the `spec.timeout` field has been removed from `RestoreSession` and replaced with `spec.restoreTimeout`. This defines how long KubeStash should wait for a restore to complete before marking it as failed.

#### Handle Backup Or Restore Timeout
We’ve fixed a bug where the backup/restore job could remain active even after the deadline was exceeded (based on the timeout) with the `BackupSession` marked as failed. Now, KubeStash will wrap timeout with backup commands so the backup/restore job(s) will update `Snapshot` status with deadline exceeded error message and mark it as failed.

#### Handle Unexpected failures for Backup Or Restore Containers
We’ve fixed a bug that caused the `BackupSession` to remain in the running phase when the backup/restore container failed unexpectedly with an error (i.e. OOMKill).


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).