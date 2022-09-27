---
title: Introducing Stash v2022.09.29
date: 2022-09-29
weight: 10
authors:
  - Emruz Hossain
tags:
  - kubernetes
  - stash
  - backup
  - restore
---

We are very excited to announce Stash `v2022.09.29`. It comes with a few new features, bug fixes, and major improvements to the codebase. We have also removed support for a few unused features. In this post, we are going to highlight the most significant changes.

You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.09.29/README.md).

### New Features

Here, are the major features that have been added in this release.

##### Support Retry for Failed Backup

In this release, we have added retry support for failed backup.You can now add a `retryConfig` field in your `BackupConfiguration`, `BackupBatch`, and `BackupBlueprint` spec.

This is what `retryConfig` will look like:

```yaml
  retryConfig:
    maxRetry: 3
    delay: 10m
```

Here,

- `retryConfig.maxRetry` specifies the maximum number of times Stash should retry if the respective backup fails.
- `retryConfig.delay` specifies the amount of time Stash should wait after a failed backup before retrying it.

For more details, please check [here](https://stash.run/docs/latest/concepts/crds/backupconfiguration/#specretryconfig).

##### Multiple BackupBlueprint support in auto-backup

You can now provide multiple BackupBlueprint names separated by a comma (`,`) in the Stash auto-backup annotation. Stash will create the respective BackupConfiguration and Repository for each BackupBlueprint. This will be particularly helpful when you want to keep backed-up data of different schedules into different backends.

Here, is a sample auto-backup annotation that shows passing multiple BackupBlueprint names:

```yaml
stash.appscode.com/backup-blueprint: daily-gcs-backup,weekly-s3-backup
```

##### Support Kubernetes 1.25.x

We have upgraded underlying Kubernetes libraries to v1.25.1. Stash should now work flawlessly in Kubernetes 1.25.x clusters.

##### Upgrade CronJob APIs to v1

We have upgraded CronJob APIs from `batch/v1beta1` to `batch/v1`. For older clusters where `batch/v1` CronJob is not available, Stash automatically uses `batch/v1beta1` CronJob.

##### Upgrade VolumeSnapshot APIs to v1

We have also upgraded VolumeSnapshot APIs to `v1` version. The VolumeSnapshot features should now work in the newer clusters without any issues.

##### Support acquiring license from license-proxyserver

We have added logic to acquire the license from a proxy license server running inside the cluster. This will let us introduce automatic license renewal in the future.

##### Add Redis 7.0.5 addon

We have added an addon for Redis 7.0.5. Now, you should be able to back up your Redis 7.x.x server using this addon.

### Bug Fixes

- Fixed defaulting issue in hook ([#179](https://github.com/stashed/apimachinery/pull/179)).
- Fixed namespace defaulting in restore target ([#180](https://github.com/stashed/apimachinery/pull/180)).
- Fixed RestoreSession phase calculation ([#187](https://github.com/stashed/apimachinery/pull/187)).
- Fixed backup/restore metrics were not published properly ([#178](https://github.com/stashed/apimachinery/pull/178)).
- Fixed VolumeSnapshot not taking snapshots of non-claimed volumes for StatefulSet.

### Improvements

We have also made some major quality-of-life improvements to the product. Here are the notable changes.

##### Refactored Codebase

We have refactored the entire codebase. This refactoring has uncovered some hidden bugs. The codebase is now organized, modular, flexible, and easy to understand. This will help us to add new features or fix bugs rapidly.

#### Structured Logging

We have now used structural logging throughout the codebase. Now, every log line has proper context. The logs are now much cleaner and more friendly for log analytics tools. We have also cleaned up the unnecessary logs. Stash should no longer flood your logs with unnecessary information.

### Deprecation/Removal

We have removed the following feature in this release.

##### Remove support for `ReplicaSet` and `ReplicationController`

Since the `ReplicaSet` and `ReplicationController` are no longer used independently to deploy an application, we have removed their backup support. You can backup their data using their parent Deployment.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.09.29/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.09.29/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
