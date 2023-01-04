---
title: Introducing Stash v2023.01.05
date: "2023-01-05"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- kubernetes
- restore
- stash
---

We are announcing Stash `v2023.01.05` which includes a bug fix and few improvements. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.01.05/README.md). In this post, we are going to highlight the changes.

### Bug Fixes

- Fixed a bug for which the temporary volume created by Stash was not updating correctly ([1498](https://github.com/stashed/stash/pull/1498)).
- Fixed a bug that was causing null pointer exception while checking for the license. ([1497](https://github.com/stashed/stash/pull/1497))

### Improvments

- We have added `stash` prefix to the temporary volume to prevent overwriting to existing volume with the name `tmp`([196](https://github.com/stashed/apimachinery/pull/196)). We have also added this prefix to the cache directory.

- We have added `stash` prefix in the output directory for addons to match with the mount path of the temporary directory. ([289](https://github.com/stashed/installer/pull/289)).



## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.06.21/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.06.21/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
