---
title: Introducing Stash v2024.02.13
date: "2023-02-14"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- elasticsearch
- opensearch
- dashboard-backup
- kibana
- restore
- stash
---

We are pleased to announce the relase of [Stash v2024.02.13](https://stash.run/docs/v2024.2.13/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2024.2.13/README.md).
In this post, we'll highlight the key updates.

### New Features

1. You can now backup and restore your `Elasticsearch` dashboard along with `Elasticsearch` database using Stash. All you have to do is to set the value of `ENABLE_DASHBOARD_BACKUP` (during backup) or `ENABLE_DASHBOARD_RESTORE` (during restore) parameter to `true` in respective `BackupConfiguration` or `RestoreSession`.

   Here is an example of the corresponding `BackupConfiguration` where we have enabled dashboard backup:
   ```yaml
   apiVersion: stash.appscode.com/v1beta1
   kind: BackupConfiguration
   metadata:
      name: sample-backup
      namespace: demo
   spec:
      schedule: "*/5 * * * *"
      task:
         name: elasticsearch-backup-7.14.0
         params:
         - name: ENABLE_DASHBOARD_BACKUP
           value: "true"
      repository:
        name: gcs-repo
      target:
         ref:
            apiVersion: appcatalog.appscode.com/v1alpha1
            kind: AppBinding
            name: es
            namespace: demo
      interimVolumeTemplate:
         metadata:
          name: sample-es-backup-tmp-storage
         spec:
            accessModes: [ "ReadWriteOnce" ]
            resources:
              requests:
                storage: 1Gi
      runtimeSettings:
         pod:
          securityContext:
            fsGroup: 65534
      retentionPolicy:
         name: keep-last-5
         keepLast: 5
         prune: true
   ```
   
   Here is an example of the corresponding `RestoreSession` where we have enabled dashboard restore: 
   ```yaml
   apiVersion: stash.appscode.com/v1beta1
   kind: RestoreSession
   metadata:
     name: sample-restore
     namespace: demo
   spec:
     task:
       name: elasticsearch-restore-7.14.0
       params:
       - name: ENABLE_DASHBOARD_RESTORE
         value: "true"
     repository:
       name: gcs-repo
     target:
       ref:
         apiVersion: appcatalog.appscode.com/v1alpha1
         kind: AppBinding
         name: es
         namespace: demo
     interimVolumeTemplate:
       metadata:
         name: sample-es-restore-tmp-storage
       spec:
         accessModes: [ "ReadWriteOnce" ]
         resources:
         requests:
         storage: 1Gi
     runtimeSettings:
       pod:
         securityContext:
           fsGroup: 65534
     rules:
     - snapshots: [latest]
   ```
   
   You can find the available `Elasticsearch` addon versions [HERE](https://stash.run/docs/v2024.2.13/addons/elasticsearch/#available-elasticsearch-addon-versions). Any addon with matching major version with the database version should be able to take backup of that database. For example, Elasticsearch addon with version `7.x.x` should be able to take backup of any Elasticsearch of `7.x.x` series. Here we have added dashboard backup and restore support for Elasticsearch of `7.x.x` and `8.x.x` series.
   
2. We’ve added support for disabling TLS certificate verification for S3 storage backend ([#219](https://github.com/stashed/apimachinery/pull/219)). We’ve introduced a new field insecureTLS in `Repository`’s `spec.backend.s3` section. When this field is set to `true`, it disables TLS certificate verification. By default, this value is set to `false`. It is important to note that this option should only be utilized for testing purposes or in combination with `VerifyConnection` or `VerifyPeerCertificate`.

   Below is an example demonstrating the usage of disabling TLS certificate verification for a TLS-secured S3 storage backend:
   ```yaml
   apiVersion: stash.appscode.com/v1alpha1
   kind: Repository
   metadata:
     name: s3-repo
     namespace: demo
   spec:
     backend:
       s3:
         endpoint: s3.amazonaws.com
         bucket: stash-demo
         region: us-west-1
         prefix: /backup/demo
         insecureTLS: true
       storageSecretName: s3-secret
   ```

### Improvements & bug fixes
- Previously the addon task parameters in `BackupConfiguration` always replaced by the parameters provided in the tasks. Now we fixed the issue by upserting the parameters that are provided in the `BackupConfiguration`. 


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
