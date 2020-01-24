---
title: Stash v0.9.0-rc.4 - Introducing Batch Backup and Hooks
date: 2020-01-24
weight: 15
authors:
  - Emruz Hossain
tags:
  - kubernetes
  - stash
  - backup
---

We are very excited to announce Stash `v0.9.0-rc.4` which brings some cool features like batch backup and hooks. We have also added `Percona-XtraDB`  addon. This version also comes with some bug fixes and general improvements.

## What's New

The following new features has been introduced in this version.

### Batch Backup

Sometimes, a single component may not meet the requirement for your application. For example, in order to deploy a WordPress, you will need a Deployment for the WordPress and another Deployment for database to store it’s contents. Now, you may want to backup both of the deployment and database under a single configuration as they are parts of a single application.

In order to support this use case, we are introducing [BackupBatch](https://stash.run/docs/v0.9.0-rc.4/concepts/crds/backupbatch/) CRD. This allows to specify multiple targets under single configuration. For more details, please check these guides:

- [BackupBatch CRD Specification](https://stash.run/docs/v0.9.0-rc.4/concepts/crds/backupbatch/)
- [How Batch Backup Works](https://stash.run/docs/v0.9.0-rc.4/guides/latest/batch-backup/overview/)
- [An Example of Batch Backup](https://stash.run/docs/v0.9.0-rc.4/guides/latest/batch-backup/batch-backup/)

### Hooks

You may need to prepare your application before backup/restore. For, example you may want to pause your application before backup and resume it after backup to ensure the backed up data consistency. You may want to send a notification after a backup is completed. Or, you may want to remove corrupted data before restoring. In order to support those use cases, we are introducing `Hooks` in this version.

Stash hooks allows executing certain action before and after a backup/restore process. You can send HTTP requests to a remote server via `httpGet` or `httpPost` hooks. You can check whether a TCP socket is open using `tcpSocket` hook. You can also execute some commands into your application pod using `exec` hook. For more details, please follow these guides:

- [Types of Hooks and How They Works](https://stash.run/docs/v0.9.0-rc.4/guides/latest/hooks/overview/)
- [Example of Backup and Restore Hooks](https://stash.run/docs/v0.9.0-rc.4/guides/latest/hooks/backup-and-restore-hooks/)
- [Example of Hooks in Batch Backup](https://stash.run/docs/v0.9.0-rc.4/guides/latest/hooks/batch-backup-hooks/)
- [How to Configure Different Types of Hooks](https://stash.run/docs/v0.9.0-rc.4/guides/latest/hooks/configuring-hooks/)

### Percona-XtraDB Addon

We have added a new addon for taking backup `Percona-XtraDB` cluster. Currently, it only supports Percona XtraDB version `5.7`. Please check these guides for more details:

- [What is a Stash Addon](https://stash.run/docs/v0.9.0-rc.4/guides/latest/addons/overview/)
- [How Percona XtraDB Backup Works](https://stash.run/docs/v0.9.0-rc.4/addons/percona-xtradb/overview/)
- [How to Install Percona-XtraDB Addon](https://stash.run/docs/v0.9.0-rc.4/addons/percona-xtradb/setup/install/)
- [Example of How to Backup a Percona-XtraDB Cluster](https://stash.run/docs/v0.9.0-rc.4/addons/percona-xtradb/guides/5.7/clustered/)

## Important Bug Fixes

We have fixes the following noticeable bugs:

- Failed to create CronJob when BackupConfiguration name is same as Database crd name [#1022](https://github.com/stashed/stash/issues/1022)
- Can't install stash [#1032](https://github.com/stashed/stash/issues/1032)
- Backup mysql creates infinite dumps [#1030](https://github.com/stashed/stash/issues/1030)
- BackupSession skipped for paused BackupConfiguration [#997](https://github.com/stashed/stash/issues/997)
- Operator is observing panic() while handling failed CronJob creation for BackupConfiguration [#1019](https://github.com/stashed/stash/issues/1019)
- Stash Addons: Make chart.sh script compatible with helm 3[#984](https://github.com/stashed/stash/issues/984)
- backup error for items [#1018](https://github.com/stashed/stash/issues/1018)
- RBAC permission missing in Chart for PersistentVolumeClaim, exists in installer deploy script [#981](https://github.com/stashed/stash/issues/981)
- Clarify that user can not user `/stash` as `mountPath` for local backend [#945](https://github.com/stashed/stash/issues/945)
- Invalid flag name for PVC restorer job [#956](https://github.com/stashed/stash/issues/956)

Please try this version and let us know if you face any issues. Ping us on [Slack](https://appscode.slack.com) or file an issue [here](https://github.com/stashed/stash/issues).
