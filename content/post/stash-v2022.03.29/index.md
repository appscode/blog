---
title: Introducing Stash v2022.03.29
date: 2022-03-30
weight: 15
authors:
  - Piyush Kanti Das
tags:
  - kubernetes
  - stash
  - backup
  - restore
  - cross-namespace
  - grafana
  - cli
  - docs
---

We are very excited to announce Stash `v2022.03.29`.  In this release, we have made a bunch of enhancements and improvements for Stash. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.03.29/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

### Support templating in hooks

Stash now supports templating in backup and restore hooks. This is particularly useful when you want to send different messages in Slack notifications based on backup/restore status.

You can check the hook templating in action in this [video](https://youtu.be/MREdcm9S8Xg).

### Support credential-less backup

We have added support for credential-less backup in this release. Now, you can use your vendor-provided authentication mechanism for Kubernetes like GKE workload identity, AWS IRSA or kube2iam, etc. with Stash.

We have introduced two new fields named `spec.runtimesettings.serviceAccountAnnotations ` , and `spec.runtimesettings.podAnnotations ` to support this feature. Stash can upload its backup data or download the data from the backend using these Annotations. You need to use these Annotations in BackupConfiguration or RestoreSession to get the backup/restore done.
Restic password is still required in your Secret to take backup or restore using Stash. However, you can use the secret keys in the Secret like our old method as well. Both of the authentication methods are now supported in Stash.


### Better Auto-backup Experience

You can now keep your Repository and storage Secret separate from the application namespaces. This lets you configure a backup policy for your users without exposing their storage credentials to them. For more details, please check the auto-backup documentation from [here](https://stash.run/docs/latest/guides/auto-backup/overview/).

## Bug Fixes and Improvements

We have refactored our implementations in this release for better maintainability and resiliency. Also, we have fixed some bugs and upgraded our documentation.

### Bug fixes

We have resolved a few bugs in this new release which should improve Stash’s user experience more than before. Some of the notable fixes are:
- Fixed the issue Cronjob getting skipped with Kubernetes version prior to 1.20
- Fixed the issue of the  BackupConfiguration Phase not being updated accordingly
- Fixed an RBAC issue of Restore init container

### Documentation Improvements

Here are the improvements we made on the documentation side.
- Updated documentation homepage to serve as a quick tour
- Added requirements section in installation docs
- Added FAQ section
- Added documentation for filtering files during backup/restore
- Improved troubleshooting guide
- Added guide for cross-cluster backup and restore
- Improved CLI docs
- Improved concept docs

### Use Go 1.18 and restic 0.13.0

All of the Stash components are now updated with Go 1.18. We have also updated the underlying restic version to 0.13.0.


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.03.29/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.03.29/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
