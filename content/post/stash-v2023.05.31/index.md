---
title: Introducing Stash v2023.05.31
date: "2023-05-31"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- elasticsearch
- kubernetes
- mongodb
- redis
- restore
- stash
---

We are announcing Stash v2023.05.31 which includes various bug fixes and enhancements. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.05.31/README.md). In this post, we are going to highlight the key changes.

### Bug Fixes

- Fixed a bug that was causing the BackupSession to remain in the `Running` phase, even after a MongoDB shard backup failure. [#1806](https://github.com/stashed/mongodb/pull/1806)
- Fixed a bug that was causing the failure of shard backup for `MongoDB 5.0.3` due to an authorization error.
- Fixed a bug that resulted in backup failures for MongoDB when TLS is enabled. [#1805](https://github.com/stashed/mongodb/pull/1805)
- Fixed a bug that was causing failures in the creation of backup and restore jobs.
- Fixed a bug that was causing the `get snapshots` command to fail for the local backend of NFS.
- Fixed a bug that was causing the `Skipped` BackupSessions to not be cleaned up. [#1525](https://github.com/stashed/stash/pull/1525)
- Fixed a bug that was that was causing an incorrect name format for the Cronjob created for scheduled backup.[#1525](https://github.com/stashed/stash/pull/1525)

### Enhancements

#### Automatic Storage Size Calculation for Elasticsearch Backup and Restore
We have introduced a new feature to simplify the Elasticsearch backup and restore process by automatically calculating the storage limit for the interim PVC. Now, you have the option to leave the resources field empty, and Stash will handle the calculation for you. This is support for Elasticsearch 7.15+ and Opensearch v1.x and 2.x. Stash wil calculate the PVC size by summing up the size of the ES indices and adding a 20% buffer for safety. The minimum interim PVC size will be 1 GiB. The PVC size will be stored as part of the snapshot and will be used by the restore process automatically if no storage limit is set in a RestoreSession object. Below is an example of how to take advantage of this feature:

```yaml
interimVolumeTemplate:
    metadata:
      name: sample-es-backup-tmp-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      # resources:
      # requests:
      # storage: 1Gi
```

With this improvement, you no longer need to manually specify the storage size, as Stash will dynamically determine it for you.

#### Backup and Restore TLS-enabled Redis clusters
In this latest release, we have introduced support for the backup and restoration of TLS enabled Redis clusters. [#172](https://github.com/stashed/redis/pull/172)

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.03.20/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.03.20/setup/upgrade/).
- If you want to upgrade the Stash kubectl Plugin, please follow the instruction from [here](https://stash.run/docs/v2023.03.20/setup/install/kubectl-plugin/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
