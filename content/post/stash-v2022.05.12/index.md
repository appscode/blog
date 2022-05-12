---
title: Introducing Stash v2022.05.12
date: 2022-05-12
weight: 10
authors:
  - Piyush Kanti Das
tags:
  - kubernetes
  - stash
  - backup
  - restore
  - cross-namespace
  - kubedump
  - grafana
  - cli
  - docs
---

We are very excited to announce Stash `v2022.05.12`.  In this release, we have added exciting new features and improvements. We have squashed a few bugs as well. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.05.12/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

### Introducing KubeDump Add-on

In this release, we have added Kubedump to our Stash add-ons family. Now, you can backup the YAMLs of your cluster resources using Stash. This add-on lets you backup the YAMLs of an application along with its dependants, all the resources of a particular namespace, and all the resources of the entire cluster.

To know more details about taking backup using KubeDump, please follow the guides from below:
- [Cluster Backup](https://stash.run/docs/latest/addons/kubedump/cluster/)
- [Namespace Backup](https://stash.run/docs/latest/addons/kubedump/namespace/)
- [Application Manifest Backup](https://stash.run/docs/v2022.05.12/addons/kubedump/application/)

### Support cross-namespace Target reference

We are introducing support for cross-namespace Target reference. Now, you can refer to a Target in BackupConfiguration and RestoreSession from different namespaces. So, you can now use a dedicated namespace for keeping Stash resources isolated from your applications. It lets you manage the backup and restore of your applications across all namespaces from the dedicated namespace.

However, these features currently work for database backup only. For workload backup, you must create respective BackupConfiguration or RestoreSession in the application namespace.

To know more details on how you can manage backup and restore from a dedicated namespace, please check the documentation from [here](https://stash.run/docs/latest/guides/managed-backup/dedicated-backup-namespace/). 

### Add support for `topologySpreadConstraints` to RuntimeSettings

You can now pass topologySpreadConstraints to BackupConfiguration or RestoreSession through `spec.runtimesettings.pod` section.

### Use restic 0.13.1

We have also updated the underlying restic version to 0.13.1. It fixes some underlying bugs of restic. For more details, please follow the link [here](https://github.com/restic/restic/releases/tag/v0.13.1).

## Bug Fixes and Improvements

We have refactored the codebase in this release for better maintainability and resiliency. Also, we have fixed some bugs and upgraded our documentation. Here are a few notable changes,

- Fix: VolumeSnapshotter backup stuck in NotReady state [Fix #1439](https://github.com/stashed/stash/pull/1439)
- Fix: RBAC resources names created by Stash to avoid conflicts [Fix #1438](https://github.com/stashed/stash/pull/1437)
- Improvement: Utilize user-provided RBAC permissions in backup and restore jobs for cross-namespace-target support [Improvement #1441](https://github.com/stashed/stash/pull/1441)

### Documentation Improvements

Here are the improvements we made on the documentation side.
- Added documentation for sending backup notification into [Slack incoming webhook](https://stash.run/docs/latest/guides/hooks/slack-notification/)
- Added [Managed Backup](https://stash.run/docs/latest/guides/managed-backup/dedicated-backup-namespace/) section that demonstrates different use-cases of cross-namespace-target and cross-namespace-repository
- Added documentation for defining [cluster wide backup policy using Auto-Backup](https://stash.run/docs/latest/guides/managed-backup/dedicated-backup-namespace-auto-backup/)
- Added documentation for BackupConfiguration's [NotReady](https://stash.run/docs/latest/guides/troubleshooting/how-to-troubleshoot/#backupconfiguration-notready) and RestoreSession's [Pending](https://stash.run/docs/latest/guides/troubleshooting/how-to-troubleshoot/#restore-pending) state in the Troubleshooting guides
- Added documentation for [Workload Identity Support](https://stash.run/docs/latest/guides/platforms/gke/) in the Platform guides
- Updated [Volume Snapshot](https://stash.run/docs/latest/guides/volumesnapshot/overview/) documentation
- Updated [FAQ](https://stash.run/docs/latest/faq/) section

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.05.12/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.05.12/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
