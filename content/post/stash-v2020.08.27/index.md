---
title: Stash - Introducing Stash Enterprise Edition
date: 2020-08-27
weight: 15
authors:
  - Emruz Hossain
tags:
  - kubernetes
  - stash
  - backup
---

We are very excited to announce Stash `v2020.08.27`. Yes, we have changed the versioning scheme. We will explain it later in this post. This version introduces two different editions of Stash named `Stash Community Edition` and `Stash Enterprise Edition`. It also introduces `RestoreBatch` CRD for restoring data backed up using `BackupBatch`.

Install Stash in your cluster by following the setup guide from [here](https://stash.run/docs/latest/setup/).

## What's New

The following major changes have been introduced in this version.

### New Versioning Scheme

We initially started Stash as a single Github repository. Back then, Stash was simple. Over time, Stash has grown bigger and complex. Now, Stash has its own Github organization and many sub-projects. We have split the original repo into multiple repositories for better code management. We have also made the sub-projects versions independent of the operator version. This helps us to work with the sub-projects easily.

The old versioning scheme referred to the operator version only. Furthermore, it was hard to co-relate which sub-project version will work with which operator version. This has lead us to adopt the new versioning scheme. The new version will follow `vYYYY.MM.DD` format. It may have an optional pre-release tag. Now, the version will refer to the entire Stash project version instead of the operator version only. The individual components may have their own versioning independent of the project version. The component versions and their changelog for a release can be found [here](https://github.com/stashed/CHANGELOG/tree/master/releases).

### Introducing Stash Enterprise Edition

Building a production-grade tool for a cloud-native technology like Kubernetes is not free of cost. In the beginning, Stash was simple and required less maintenance. Over time, it has grown large and complex. Now, it demands a dedicated team behind it. A skilled Kubernetes team is not cheap. The test infrastructure required for it is not cheap either. So, in order to ensure a sustainable future for Stash, we are introducing an enterprise edition. From now, we will be offering Stash in two different flavors. The open-sourced `Community Edition` with basic features set will serve the non-commercial users under [PolyForm-Noncommercial-1.0.0](https://github.com/stashed/stash/blob/master/LICENSE.md) license. The open core `Enterprise Edition` with full feature set will serve the commercial users.

If you are upgrading from `0.9.x` to `v2020.08.27`, please note that following features are only available in Enterprise edition:

- **Auto-Backup:** Auto-backup is now an enterprise feature. You won't be able to setup any new backup using auto-backup. However, your existing auto-backup resources should keep functioning.
- **Batch Backup:** Batch backup and restore is also now an enterprise feature. You won't be able to create any new backup using batch-backup. However, your existing backup should continue to work and you would be able to restore the data that were backed up using BatchBackup.
- **Local Backend:** Local backend now is an enterprise feature. If you are using any Kubernetes volume (i.e. NFS, PVC, HostPath, etc.) as backend, you won't be able to create any new backup using those backends. However, your existing backup that uses sidecar model should keep functioning. You have to use the Enterprise edition to restore from the backed up data.

You can find a full feature comparison between the two versions in [here](https://stash.run/docs/latest/concepts/what-is-stash/overview/). If you are interested in purchasing Enterprise license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/30min).

### Introducing RestoreBatch

In Stash `v0.9.x`, we had introduced a `BackupBatch` CRD that supports taking backup of multiple co-related applications together. However, users had to restore the applications individually. So, it was the next logical step to introduce a similar CRD to restore. We are happy to introduce `RestoreBatch` CRD in this release. The `RestoreBatch` CRD will help the users to restore the applications together that were backed up using a `BackupBatch`.

Furthermore, we have introduced `executionOrder` filed in both `BackupBatch` and `RestoreBatch`. Now, the users will be able to control the backup or restore order of the applications.

- See detailed specification of the `RestoreBatch` from [here](https://stash.run/docs/latest/concepts/crds/restoresession/).
- See how the restore process works for a `RestoreBatch` from [here](https://stash.run/docs/latest/guides/latest/batch-backup/overview/).
- See a `RestoreBatch` in action from [here](https://stash.run/docs/latest/guides/latest/batch-backup/batch-backup/).

### File Filtering During Backup and Restore

This release adds support for filtering files during backup and restore. Now, you can take backup of a subset of files that matches some particular patterns as well as you can restore only a subset of files from the backed up data.

- See how to use the `exclude` field to ignore files during backup that matches some patterns from [here](https://stash.run/docs/latest/concepts/crds/backupconfiguration/).
- See how to use `include` field to restore only the files that match some patterns or `exclude` field to ignore files that match some patterns during restore from [here](https://stash.run/docs/latest/concepts/crds/restoresession/).

### More Auto-Backup Annotations

We have introduced two new annotations for auto-backup in this release. Now, you can overwrite the schedule specified in the respective `BackupBlueprint` for an application by adding ` stash.appscode.com/schedule: <Cron Expression>` annotation. If you don't specify this annotation, Stash will use the schedule from the respective `BackupBlueprint`.

You can also pass parameters to the respective `Task` via `params.stash.appscode.com/<key>: <value>` annotation. You can pass multiple parameters to the `Task` by adding multiple annotations. For example, you can pass multiple parameters as below,

```yaml
params.stash.appscode.com/key1: value1
params.stash.appscode.com/key2: value2,value3
params.stash.appscode.com/key3: ab=123,bc=234
```

We have also fixed some critical bugs regarding `BatchBackup` and database backup. A full changelog of this release can be found [here](https://github.com/stashed/CHANGELOG/tree/master/releases/v2020.08.27).

Please try this release and give us your valuable feedback.
