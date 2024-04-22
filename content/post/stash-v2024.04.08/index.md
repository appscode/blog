---
title: Introducing Stash v2024.4.8
date: "2024-04-09"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- disaster-recovery
- kubedump
- kubernetes
- postgres
- restore
- stash
- tls
---

We are pleased to announce the release of [Stash v2024.4.8](https://stash.run/docs/v2024.4.8/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2024.4.8/README.md).  In this post, we'll highlight the changes done in this release.

### New Features

1. With this new feature in `Kubedump`, you can now specify the types of resources you want to ignore during backup by defining their `GroupKind`. GroupKind refers to the category and type of Kubernetes resource. For example, `Deployment` is a kind in `apps` group. All you have to do is to provide the value of `ignoreGroupKinds` parameter in respective `BackupConfiguration`. When specifying which GroupKinds to ignore, you'll use the `<kind.group>` format. For instance, if you want to exclude Deployments from your backup, you would provide it like this: `Deployment.apps`. You can specify multiple GroupKinds to ignore by separating them with commas. For instance, if you want to ignore DaemonSets and Snapshots, you would list them like this: `DaemonSet.apps,Snapshot.repositories.stash.appscode.com`.

   Here is an example of the corresponding `BackupConfiguration` where we have ignored `Snapshot` resources:
   ```yaml
   apiVersion: stash.appscode.com/v1beta1
   kind: BackupConfiguration
   metadata:
     name: demo-ns-backup
     namespace: demo
   spec:
     schedule: "*/5 * * * *"
     task:
       name: kubedump-backup-0.1.0
       params:
       - name: ignoreGroupKinds
         value: Snapshot.repositories.stash.appscode.com
     repository:
       name: s3-repo
     target:
       ref:
         apiVersion: v1
         kind: Namespace
         name: demo
     runtimeSettings:
       pod:
         serviceAccountName: cluster-resource-reader
     retentionPolicy:
       name: keep-last-5
       keepLast: 5
       prune: true
   ```

2. We’ve added support for setting node labels, tolerations and affinity rules for CRD installer and cleaner job. You can find the values file [HERE](https://github.com/stashed/installer/blob/a5779c26b41c2ded9b0e6fa1372dc517b2eb9f71/charts/stash-enterprise/values.yaml).

3. We have added support for backup and restore of `PostgreSQL` version `16`. You can find the supported `PostgreSQL` addon versions [HERE](https://stash.run/docs/v2024.4.8/addons/postgres/#supported-postgresql-versions).

### Bug fixes

We have fixed a bug for backup and restore of tls enabled `PostgreSQL` instance.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
