---
title: Introducing Stash v2023.05.31
date: "2023-05-31"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- restore
- kubernetes
- mongodb
- redis
- elasticsearch
- stash
---

We are announcing Stash v2023.05.31 which includes a few bug fixes and enhancements. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.05.31/README.md). In this post, we are going to highlight the changes.

### Bug Fixes

- Fixed a bug that was causing the BackupSession to remain in the `Running` phase, even after a MongoDB shard backup failure. [#1806](https://github.com/stashed/mongodb/pull/1806)
- Fixed a bug that was causing the failure of shard backup for `MongoDB 5.0.3`  due to an authorization error.
- Fixed a bug that resulted in backup failures for MongoDB when TLS is enabled. [#1805](https://github.com/stashed/mongodb/pull/1805)
- Fixed a bug that was causing failures in the creation of backup and restore jobs.
- Fixed a bug that was causing the `get snapshots` command to fail for the local backend of NFS.
- Fixed a bug that was causing the `Skipped` BackupSessions to not be cleaned up. [#1525](https://github.com/stashed/stash/pull/1525)
- Fixed a bug that was that was causing an incorrect name format for the Cronjob created for scheduled backup.[#1525](https://github.com/stashed/stash/pull/1525)

### Enhancements

#### Automatic Storage Size Calculation for Elasticsearch Backup and Restore
We have introduced a new feature to enhance the convenience of Elasticsearch backup and restore by enabling automatic calculation of the storage limit for the interim volume template. Now, you have the option to leave the resources field empty, and Stash will handle the calculation process for you. Here's an example of how to utilize this functionality:

```yaml
interimVolumeTemplate:
    metadata:
      name: sample-es-backup-tmp-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      # resources:
      # requests:
      # storage: 1Gii
```
With this improvement, you no longer need to manually specify the storage size, as Stash will dynamically determine it for you.

#### Backup and Restore TLS-enabled Redis clusters
In this latest release, we have introduced support for the backup and restoration of Redis clusters with TLS encryption enabled. [#172](https://github.com/stashed/redis/pull/172)

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.03.20/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.03.20/setup/upgrade/).
- If you want to upgrade the Stash kubectl Plugin, please follow the instruction from [here](https://stash.run/docs/v2023.03.20/setup/install/kubectl-plugin/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
