---
title: Introducing Stash v2022.06.27
date: 2022-06-27
weight: 10
authors:
- Hossain Mahmud
tags:
- kubernetes
- stash
- backup
- restore
---

We are announcing Stash `v2022.06.27` which include a critical bug fix for PostgreSQL and MongoDB addon. This release fixes a bug that was causing the license checker to fail in the backup and restore jobs for PostgreSQL ([#1077](https://github.com/stashed/postgres/pull/1077)) and MongoDB ([#1593](https://github.com/stashed/mongodb/pull/1593)) databases. This bug was introduced in [Stash v2022.06.21](https://blog.byte.builders/post/stash-v2022.06.21/) release. So, if you deployed Stash v2022.06.21 previously, you are recommended to upgrade to today's release.

You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.06.27/README.md).

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.06.21/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.06.21/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
