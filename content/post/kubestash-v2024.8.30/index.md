---
title: Introducing KubeStash v2024.8.30
date: "2024-08-30"
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
- mssqlserver
- restore
---

We are very excited to announce the release of [KubeStash v2024.8.30](https://kubestash.com/docs/v2024.8.30/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.8.30/README.md). 

In this post, we'll highlight the key updates for both the `v2024.8.14` and `v2024.8.30` releases. Since the blog post for our `v2024.8.14` release has been delayed, we will be combining the updates for both the releases into a single blog post. This will allow us to cover all the key updates and improvements in one comprehensive overview.  

### New Features

Here, we are going to highlight the new features that have been introduced in `v2024.8.14` release.

#### KubeDB Managed MSSQLServer Database Backup & Restore

We've introduced functionality to backup and restore [KubeDB managed MSSQLServer](https://kubedb.com/docs/v2024.8.21/guides/mssqlserver/) database.

Here is an example of `BackupConfiguration` that takes backup of a `MSSQLServer`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: mssql-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: mssqlserver-ag
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
    sessionHistoryLimit: 1
    scheduler:
      schedule: "*/5 * * * *"
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: gcs-mssql-repo
      backend: gcs-backend
      directory: /mssql
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
    addon:
      name: mssqlserver-addon
      jobTemplate:
        spec:
          securityContext:
            runAsUser: 0
      tasks:
      - name: logical-backup
        params:
          databases: agdb1, agdb2
```

In the `spec.sessions[*].addon.tasks[*].params` section, you can specify the `databases` parameter. This parameter should contain a comma-separated list of the database names you want to backup.

> Note: To run the backup job, you must use `root` user privilege because `WAL-G`, the tool we use for backups, requires this permission to function properly.


Here is the example of `RestoreSession` that restores the `MSSQLServer`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: mssql-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: restored-mssqlserver-ag
    namespace: demo
  dataSource:
    snapshot: latest
    repository: gcs-mssql-repo
  addon:
    name: mssqlserver-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
    tasks:
    - name: logical-backup-restore
      params:
        databases: agdb1
```

In the `spec.addon.tasks[*].params` section, you can specify the `databases` parameter. This parameter should contain a comma-separated list of the database names you want to restore.

> Note: To run the restore job, you must use `root` user privilege because `WAL-G`, the tool we use for restore, requires this permission to function properly.

#### KubeDB Managed MSSQLServer Manifest Backup & Restore

KubeStash now supports manifest backup and restore for [KubeDB managed MSSQLServer](https://kubedb.com/docs/v2024.8.21/guides/mssqlserver/).

Here is an example of a `BackupConfiguration` for `MSSQLServer` manifest backup:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: mssql-manifest-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: mssqlserver-ag
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
    sessionHistoryLimit: 1
    scheduler:
      schedule: "*/2 * * * *"
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: gcs-mssql-manifest-repo
      backend: gcs-backend
      directory: /mssql-manifest
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
    addon:
      name: mssqlserver-addon
      tasks:
      - name: manifest-backup
```

Here's an example of a `RestoreSession` that restores the manifest of `MSSQLServer`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-restoresession
  namespace: demo
spec:
  manifestOptions:
    restoreNamespace: demo
    msSQLServer:
      db: true
      dbName: new-mssqlserver
      authSecret: true
      authSecretName: new-authsecret-mssqlserver
      internalAuthIssuerRef:
        apiGroup: cert-manager.io
        kind: Issuer
        name: mssqlserver-issuer
      tlsIssuerRef:
        apiGroup: "cert-manager.io"
        kind: Issuer
        name: mssqlserver-issuer
  dataSource:
    snapshot: latest
    repository: gcs-mssql-manifest-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: mssqlserver-addon
    tasks:
    - name: manifest-restore
```

Here,
- `spec.manifestOptions.msSQLServer.internalAuthIssuerRef` refers to an issuer that must be created before applying the `RestoreSession` for an `AvailabilityGroup` database. This issuer is used by the restored MSSQLServer instance to authenticate internal communications.

- `spec.manifestOptions.msSQLServer.tlsIssuerRef` refers to an issuer that must be created before applying `RestoreSession`. The restored MSSQLServer instance will use this issuer to configure TLS for secure communication.

### Improvements

In the `v2024.8.30` release, we've added support to trigger an instant backup as soon as the `BackupConfiguration` is ready. This ensures that a backup is created immediately upon the `BackupConfiguration`'s creation, without having to wait for the next scheduled CronJob.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
