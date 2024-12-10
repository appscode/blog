---
title: Introducing KubeStash v2024.12.9
date: "2024-12-09"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- backup-verification
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of [KubeStash v2024.12.9](https://kubestash.com/docs/v2024.12.9/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.12.9/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### KubeDB Managed Database Backup Verification

We've introduced functionality to verify KubeDB managed database backup. You can configure following types of backup verification strategies:

- **RestoreOnly :** KubeStash operator will initiate a `RestoreSession` using the addon information specified in `BackupVerifier`. The verification of the backup will rely on the status of the `RestoreSession` phase; if the restore completes successfully, the backup is considered verified.
- **Query :** At first, KubeStash operator will initiate a `RestoreSession` and after successful restore, it will create a verifier job to run the queries provided in `BackupVerifier`.
- **Script :** At first, KubeStash operator will initiate a `RestoreSession` and after successful restore, it will create a verifier job to run the script provided in `BackupVerifier`.

Here we will give an example of query verification strategy for KubeDB managed `MySQL`.

At first, we need to check if the verification function is present in our cluster using the following command:

```bash
$ kubectl get functions | grep -i `kubedbverifier`
kubedbverifier            19h
```

To verify the backup of a `MySQL` database, a separate `MySQL` instance is required to run the verification process for the original database. Here is an example of a `MySQL` instance for verification:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: verify
spec:
  version: "8.0.35"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
```

It should have a similar configuration to the original backup application (e.g., the database version should match), as we are going to restore the original application's backup data into it.

Here is the example of `BackupVerifier` that verifies the `MySQL`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupVerifier
metadata:
  name: mysql-query-verifier
  namespace: demo
spec:
  function: kubedbverifier
  restoreOption:
    target:
      apiGroup: kubedb.com
      kind: MySQL
      name: sample-mysql
      namespace: verify
    addonInfo:
      name: mysql-addon
      tasks:
        - name: logical-backup-restore
  scheduler:
    schedule: "*/5 * * * *"
    jobTemplate:
      backoffLimit: 1
  type: Query
  query:
    mySQL:
      - database: auth
        table: user
        rowCount:
          operator: Equal
          value: 100
      - database: store
        table: order
  sessionHistoryLimit: 2
```

Here is the `BackupConfiguration` configured for the original database, which refers to the `BackupVerifier`. 

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MySQL
    name: sample-mysql
    namespace: demo
  backends:
  - name: gcs-storage
    storageRef:
      namespace: demo
      name: gcs-storage
    retentionPolicy:
      name: demo-retention
      namespace: demo
  sessions:
  - name: freq-backup
    sessionHistoryLimit: 3
    scheduler:
      schedule: "*/5 * * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: demo-storage
      backend: gcs-storage
      backupVerifier: # updated backupVerifier
        name: mysql-query-verifier
        namespace: demo
      directory: /dep/mysql
      encryptionSecret:
        name: encryption-secret
        namespace: demo
    addon:
      name: mysql-addon
      tasks:
      - name: logical-backup
    retryConfig:
      maxRetry: 2
      delay: 1m
```

Here, `.spec.sessions[*].repositories[*].backupVerifier` contains the `BackupVerifier` reference. 

#### Templating for Exec Hook Commands

We've added support for [Go template](https://pkg.go.dev/text/template) in exec type as well. Previously it was only limited to httpPost hook. Here is an example of configuring template in exec hook:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: HookTemplate
metadata:
  name: sample-hook
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  action:
    exec:
      command:
      - curl
      - -X
      - POST
      - -H
      - "Content-Type: application/json"
      - -d
      - `{"text":"Name: {{ .Name }} Namespace: {{.Namespace}} Phase: {{.Status.Phase}}"}`
      - https://slack-webhook.url/kubestash
  executor:
    type: Pod
    pod:
      selector: name=test-app, test=hook
```


### Improvements & Bug Fixes

- We’ve fixed an issue with external database backups where the database version was not resolved during the backup process. This has now been fixed in this release.
- We have also fixed an issue where the addon job template specified in the `BackupConfiguration` was not applied. It has been fixed in this release.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2024.6.4/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
