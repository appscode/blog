---
title: Introducing Stash v2022.12.11
date: "2022-12-11"
weight: 10
authors:
- Piyush Kanti Das
tags:
- backup
- kubernetes
- postgres
- restore
- stash
- vault
---

We are very excited to announce Stash `v2022.12.11`. It comes with a couple of exciting new features, a few bug fixes, and improvements to the codebase. In this post, we are going to highlight the most significant changes.

You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.12.11/README.md).

### New Features

Here, are the major features that have been added in this release.

#### Introducing Vault addon

In this release, we are introducing [Hashicorp Vault](https://github.com/stashed/vault) addon. You can now backup and recover your Hashicorp Vault using Stash Enterprise.

#### Add Postgres 15.1 addon

We have added support for Postgres 15.1. Now, you should be able to backup/restore your Postgres 15.x.x using this addon.

### Bug Fixes

- Fixed passing PodRuntimeSettings to backup/restore job ([#1489](https://github.com/stashed/stash/pull/1489)).
- Fixed addon info handling from KubeDB AppBinding ([#1487](https://github.com/stashed/stash/pull/1487)).
- Fixed a bug that was causing locked Repository ([#192](https://github.com/stashed/apimachinery/pull/192)).
- Added missing user permission for appcatalog group ([#284](https://github.com/stashed/installer/pull/284)).

### Improvements

- We have updated our dependencies ([#1495](https://github.com/stashed/stash/pull/1495)).
- We have updated our license management ([#1491](https://github.com/stashed/stash/pull/1491)).

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.12.11/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.12.11/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
