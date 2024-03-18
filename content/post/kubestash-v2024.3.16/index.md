---
title: Introducing KubeStash v2024.3.16
date: "2024-03-19"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of `KubeStash v2024.3.16`, packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2024.3.16/README.md). In this post, we'll highlight the key updates.

### New Features

Previously, some restore options such as VolumeSnapshots and volumeClaimTemplates were restricted in their functionality. We were not able to restore them to a namespace other than `RestoreSession`'s namespace. This was due to the requirement of creating a `RestoreSession` in the same namespace as the dataSource.

To resolve this issue, we've introduced a new field called `.spec.dataSource.namespace` within the `RestoreSession` (CRD)([#104](https://github.com/kubestash/apimachinery/pull/104)). This field allows you to specify the namespace where the dataSource (i.e., repository and snapshot) is located.

So, with the addition of the `.spec.dataSource.namespace` field, users can now create a `RestoreSession` in a different namespace while referencing the namespace of the dataSource.

> Note, if the `.spec.dataSource.namespace` field is left empty during restoration, the KubeStash operator will consider the DataSource is in the same namespace of the `RestoreSession`.

#### Example:

Let's consider you've previously backed up your workload application to the `backup` namespace, meaning dataSource (repository and snapshot) is located in that namespace.

Now, You might want to restore the PVCs in the `restore` namespace by specifying the volumeClaimTemplates within the `RestoreSession`.

  Here's an example of a `RestoreSession` with the `.spec.dataSource.namespace` field set:

```yaml
 apiVersion: core.kubestash.com/v1alpha1
 kind: RestoreSession
 metadata:
   name: sample-pvcs-restore
   namespace: restore # The namespace where you want to restore the backup data.
 spec:
   spec:
   dataSource:
     namespace: backup # The namespace where your data was previously backed up. The snapshot and repository exist there.
     repository: gcs-repo
     snapshot: latest
     encryptionSecret:
       name: encrypt-secret # some addon may not support encryption
       namespace: backup
   addon:
     name: pvc-addon
     tasks:
       - name: volume-clone
         targetVolumes:
           volumeMounts:
             - name:  restore-data
               mountPath:  /source/data
           volumeClaimTemplates:
             - metadata:
                 name: restore-data
               spec:
                 accessModes: [ "ReadWriteOnce" ]
                 storageClassName: "standard"
                 resources:
                   requests:
                     storage: 2Gi
```

### Improvements & Bug Fixes
- During the backup process of VolumeSnapshots, if the retentionPolicy applier container attempts to clean a volume snapshot that was unintentionally deleted by the user, then the backup will be fail. To address this issue, we ignore the "not found" error.
- During the Backup or Restore process of `KubeDB` managed databases, the KubeStash operator will not initiate backup or restore process until the `AppBinding` is created. `AppBinding` is necessary to resolve the KubeDB managed databases addon version and calculating interim volume size if necessary.
- Fixed issue regarding the `Repository` size update. Now we can find the correct information about the size of backed-up data stored in the `Repository`.


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://github.com/kubestash/installer/blob/master/charts/kubestash-operator/README.md).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).