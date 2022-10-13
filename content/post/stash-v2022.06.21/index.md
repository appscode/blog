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

We are very excited to announce Stash `v2022.06.21`. This release adds support for Kubernetes `v1.24.x`. We have also introduces a few new features and improvements. We have squashed a few bugs as well. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.06.21/README.md).

We are going to highlight the major changes in this post.

## New Features

Here are the highlighting features of this release:

- **Support Kubernetes `v1.24.x`:** We have upgraded the underlying libraries to support Kubernetes `v1.24.x`. Now, Stash should work flawlessly with the latest Kubernetes version.
- **Add timeout for backup and restore**: You can now specify a timeout for your backup / restore process using `spec.timeOut` field of BackupConfiguration or RestoreSession. Respective backup/restore process will be considered as `Failed` if it does not complete within this time limit. This will ensure that backup process won't get stuck by some kind of network issue. By default, Stash don’t set any timeout for the backup/restore.

- **Add Hook and timeOut support in Auto-Backup**: We have added Hooks and timeout support in Auto-Backup as well. Now, you can specify a backup hook in your BackupBlueprint object.

## Bug Fixes and Improvements

We have also fixed few bugs and made some improvements. Here are the notable changes,

- Fixed a bug that put the restore job to CrashloopBackOfff state in some workload restore process [#1448](https://github.com/stashed/stash/pull/1448).
- Fix RBAC permissions for backup sidecar and restore init-container [#1452](https://github.com/stashed/stash/pull/1452).
- Update metrics labels for Stash Grafana dashboards [#170](https://github.com/stashed/apimachinery/pull/170).
- Updated [restic](https://github.com/restic/restic) version to 0.13.1 in Stash Addons.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.06.21/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.06.21/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
