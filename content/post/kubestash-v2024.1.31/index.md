---
title: Introducing KubeStash v2024.1.31
date: "2024-02-02"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of `KubeStash v2024.1.31`, packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.1.31/README.md). 
In this post, we'll highlight the key updates.

### New Features

1. We've added support to provide multiple sessions for a single repository. You can now use a single repository across multiple sessions in a `BackupConfiguration`. ([#94](https://github.com/kubestash/apimachinery/pull/94)). 

   Here is an example of a single repository across multiple sessions in a `BackupConfiguration`:
   ```yaml
   apiVersion: core.kubestash.com/v1alpha1
   kind: BackupConfiguration
   metadata:
     name: pvc-backup
     namespace: demo
   spec:
     target:
       apiGroup:
       kind: PersistentVolumeClaim
       name: target-pvc
       namespace: demo
     backends:
     - name: pvc-backend
       storageRef:
         name: gcs-storage
         namespace: demo
       retentionPolicy:
         name: demo-retention
         namespace: demo
     sessions:
     - name: daily-backup
       sessionHistoryLimit: 3
       scheduler:
         schedule: "0 0 * * *"
         jobTemplate:
           backoffLimit: 0
       repositories:
       - name: demo-pvc-gcs
         backend: pvc-backend
         directory: /target-pvc
         encryptionSecret:
           name: encry-secret
           namespace: demo
         deletionPolicy: WipeOut 
       addon:
         name: pvc-addon
         tasks:
         - name: logical-backup
     - name: monthly-backup
       sessionHistoryLimit: 3
       scheduler:
         schedule: "0 0 1 * *"
         jobTemplate:
           backoffLimit: 0
       repositories:
       - name: demo-pvc-gcs
         backend: pvc-backend
         directory: /target-pvc
         encryptionSecret:
           name: encry-secret
           namespace: demo
         deletionPolicy: WipeOut 
       addon:
         name: pvc-addon
         tasks:
         - name: logical-backup
   ```
   For the above configuration you will find two separate directories based on the corresponding session name inside the repository `demo-pvc-gcs`.

2. We've added support for disabling TLS certificate verification for S3 storage backend ([#100](https://github.com/kubestash/apimachinery/pull/100)). We've introduced a new field `insecureTLS` in `BackupStorage`'s `spec.backend.s3` section. When this field is set to `true`, it disables TLS certificate verification. By default, this value is set to `false`. It is important to note that this option should only be utilized for testing purposes or in combination with `VerifyConnection` or `VerifyPeerCertificate`.

   Below is an example demonstrating the usage of disabling TLS certificate verification for a TLS-secured S3 storage backend:
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
         endpoint: 's3.amazonaws.com'
         bucket: kubestash
         prefix: demo
         region: us-east-1
         secretName: s3-secret
         insecureTLS: true
     usagePolicy:
       allowedNamespaces:
         from: All
     default: false
     deletionPolicy: WipeOut
   ```

3. In this release, we've enhanced security by switching from `ClusterRoleBinding` to `RoleBinding` for RBAC in backup and restore jobs. This change aligns with the best practices and enhances access control, reducing potential conflicts with security policies.

   For Local `BackupStorage` (i.e. NFS, Local-host, PVC) the initializer Job uses `ClusterRoleBinding`. It's important to note that this doesn't compromise security policies, as the associated ClusterRole only grants access to resources within the KubeStash `Storage` API group.

4.  We've updated `Addon`'s tasks name ([#48](https://github.com/kubestash/installer/pull/48)). To find out the new tasks name, at first get the corresponding addon and find out proper task name in task's name section.

### Improvements & Bug Fixes
- We have updated the names of the snapshot components used for workload, PVC, and MongoDB sharded database backups. Please be aware that this change is considered breaking.
- Fixed a bug that caused incorrect pruning for failed snapshots when applying retention policies to synchronized snapshots from the backend.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

