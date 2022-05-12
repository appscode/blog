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

We are very excited to announce Stash `v2022.05.12`.  In this release, we have made a bunch of enhancements and improvements for Stash. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.05.12/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

### Introducing Kubedump Add-on

In this release, we have added Kubedump to our Stash add-ons family. Now, you can backup all of the applied manifest files in your cluster using Stash. This addon provides you with a wide range of flexibilities. For example, you can take backup of particular manifest files, manifests of a specific namespace, or manifests of a whole cluster at once. 

### Support cross-namespace Target reference

We are introducing support for cross-namespace Target reference. Now, you can refer to a Target in BackupConfiguration and RestoreSession from different namespaces. So, you can now easily take backup and restore into different application namespaces by keeping the Stash resources in an isolated namespace. For now, this feature is not compatible with Kubernetes Workload's volumes.
To ensure better security, initially, the Stash operator does not provide RBAC permissions to BackupConfigurations of RestoreSessions for getting Targets from different namespaces. You need to provide necessary permissions to BackupConfigurations and RestoreSessions through a ServiceAccount. 

To know more details on how you can manage backup and restore from a dedicated namespace, please check the documentation from [here](https://stash.run/docs/latest/guides/managed-backup/dedicated-backup-namespace/). 

### Add support for topologySpreadConstraints to RuntimeSettings

You can now pass topologySpreadConstraints to BackupConfiguration or RestoreSession through `spec.runtimesettings.pod` section.

## Bug Fixes and Improvements

We have refactored our implementations in this release for better maintainability and resiliency. Also, we have fixed some bugs and upgraded our documentation.

### Bug fixes

We have resolved a few bugs in this new release which should improve Stash’s user experience more than before. Some of the notable fixes are:
- Fixed VolumeSnapshotter backup issues
- Fixed implementation to pass user-provided RBAC permissions to backup and restore jobs properly
- Fixed naming of some resources created by Stash to avoid conflicts
- Updated Auto-Backup implementation to support cross-namespace target

### Documentation Improvements

Here are the improvements we made on the documentation side.
- Added `Managed Backup` section
- Added documentation on managing backup from a dedicated namespace
- Added documentation for Automatic Backup in a dedicated namespace
- Added documentation for filtering files during backup/restore
- Improved troubleshooting guide
- Improved webhooks documentation
- Updated Platform documentation for Workload Identity Support
- Improved FAQ docs

### Use restic 0.13.1

We have also updated the underlying restic version to 0.13.1.


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.05.12/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.05.12/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
