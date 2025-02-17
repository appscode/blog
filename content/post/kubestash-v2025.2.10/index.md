---
title: Introducing KubeStash v2025.2.10
date: "2025-02-10"
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

We are pleased to announce the release of [KubeStash v2025.2.10](https://kubestash.com/docs/v2025.2.10/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.2.10/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### KubeDB Archiver Backup/Restore

We've introduced functionality to back up and restore the KubeDB `Archiver` resource. By default, KubeStash backs up the `Archiver` YAML during the manifest backup of any KubeDB-managed databases.

Below is an example of a `RestoreSession` for restoring an `Archiver` YAML for `MySQL` database:

```yaml
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mysql-full
        namespace: demo
      recoveryTimestamp: "2025-03-11T14:26:01Z"
      manifestOptions:
        archiver: true
        archiverRef:
          namespace: kubedb
          name:  mysqlarchiver-sample
---
```

> Note: By default, we don't restore `Archiver`. If you want to restore set the archiver field `true`.

Here,
- `spec.init.archiver.manifestOptions` specifies that the Archiver YAML should be restored during manifest restoration.
- `spec.init.archiver.manifestOptions.archiverRef` defines the target namespace and name for restoring the Archiver.

#### Single Dump, Multiple Backup Destinations

We've introduced a powerful new feature for database backups. `Multi-writer` support for the `restic` driver. Now, instead of executing a separate `dump` for each backend, we dump the database once and simultaneously stream the data to multiple backup destinations. This improves efficiency and reduces redundancy in the backup process.

Here’s an example of a `BackupConfiguration` for a `MySQL` database with multiple backup destinations,

```yaml
---
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        ---
      repositories:
        - name: gcs-mysql-repo
          backend: gcs-backend
          directory: /mysql-gcs
          encryptionSecret:
            ---
        - name: s3-mysql-repo
          backend: s3-backend
          directory: /mysql-s3
          encryptionSecret:
            ---
      addon:
        name: mysql-addon
        tasks:
          - name: logical-backup
```
- `spec.backends` section configures multiple backends, one for S3 and another for GCS.
- `spec.sessions[0].repositories` section references multiple backends, ensuring that backup data is stored in both destinations.

#### AWS Pod Identity Support

We've added support for `credential-less` backup and restore through AWS Pod Identity. This eliminates the need to manage AWS credentials manually.

Follow the AWS documentation to set up an EKS cluster with `pod-identity`:
- [Pod Identities Overview](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html)
- [How Pod Identity Works](https://docs.aws.amazon.com/eks/latest/userguide/pod-id-how-it-works.html)
- [Setting Up the Pod Identity Agent](https://docs.aws.amazon.com/eks/latest/userguide/pod-id-agent-setup.html)

### Improvements & Bug Fixes

#### Migrated S3 SDK from aws-sdk-v1 to aws-sdk-v2

We have upgraded our S3 backend codebase from `aws-sdk-v1` to `aws-sdk-v2`, as AWS is deprecating updates for `aws-sdk-v1`. To cooperate with `aws-sdk-v2`, S3 endpoints now require:

- `https://` prefix for public S3 providers.
- `http://` prefix for self-managed or S3-compatible storage.

Below is an updated example of a BackupStorage configuration using the new S3 endpoint format:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: appscode-testing
      region: us-west-2
      endpoint: https://s3.us-west-2.amazonaws.com
      secretName: aws-s3-secret
      prefix: appscode-qa
  usagePolicy:
    allowedNamespaces:
      from: All
  default: false
  deletionPolicy: WipeOut
```

Here,
- `spec.storage.s3.endpoint` now explicitly includes the `https://` prefix for AWS S3 endpoints.

#### New Deletion Policy for BackupConfiguration

We've introduced a new deletion policy, `Retain`, for BackupConfiguration. This allows users to delete a `BackupConfiguration` while retaining the associated snapshots and repository.

Below is an example of a `BackupConfiguration` with the deletion policy set to `Retain`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
---
spec:
  sessions:
  - name: frequent-backup
    repositories: 
    - name: pvc-backup
      backend: gcs-storage
      directory: /pvc-backup
      encryptionSecret:
       name: encrypt-secret 
       namespace: demo
      deletionPolicy: Retain
---
```

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).