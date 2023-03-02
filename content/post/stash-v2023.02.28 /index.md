---
title: Introducing Stash v2023.02.28
date: "2023-02-28"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- kubernetes
- restore
- stash
---

We are announcing Stash `v2023.02.28` which includes a few bug fixes. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.02.28/README.md). In this post, we are going to highlight the changes.

### Bug Fixes

- Fixed a bug causing Batch Backups to get stuck in the Running phase when using the `VolumeSnapshotter` driver ([#1500](https://github.com/stashed/stash/pull/1500)).
- Fixed a bug that prevented the autobackup template from fully resolving.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.02.28/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.02.28/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
