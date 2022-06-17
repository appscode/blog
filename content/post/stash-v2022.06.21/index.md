---
title: Introducing Stash v2022.06.21
date: 2022-06-21
weight: 10
authors:
  - Piyush Kanti Das
tags:
  - kubernetes
  - stash
  - backup
  - restore
  - grafana
  - timeout
  - hook
---

We are very excited to announce Stash `v2022.06.21`.  In this release, we have added a few new features and improvements. We have squashed a few bugs as well. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.06.21/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

- **Add timeout for backup and restore**: In this release, we have added timeout for backup and restore. Now, you can mention a timeout in the `spec.task` field of BackupSession or RestoreSession. BackupSession/RestoreSession will be considered as `Failed` if the backup/restore does not complete within this time limit. By default, Stash don’t set any timeout for the backup/restore.

- **Add Hook and timeOut support in Auto-Backup**: We have added Hooks and timeout support in Auto-Backup as well. Now, you can use pre-backup or post-backup Hooks from BackupBlueprint similar to BackupConfiguration.

## Bug Fixes and Improvements

We have also fixed few bugs and made some improvements. Here are the notable changes,

- Fix: Fixed a bug by checking if the restore output is nil or not. This nil restore output put the Restore job to CrashloopBackOfff state in some workload restore process [#1448](https://github.com/stashed/stash/pull/1448).
- Fix: Fix RBAC permissions for Backup sidecar and Restore init-container [#1452](https://github.com/stashed/stash/pull/1452).
- Improvement: Update metrics labels for Stash Graphana dashboards [#170](https://github.com/stashed/apimachinery/pull/170).
- Improvement: Clean up dependencies in Stash [#1451](https://github.com/stashed/stash/pull/1451).
- Improvement: Update toolchain to Kubernetes 1.24 [#1450](https://github.com/stashed/stash/pull/1450).
- Improvement: Use Restic 0.13.1 in Stash Addons

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.06.21/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.06.21/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
