---
title: Introducing Stash v2023.04.30
date: "2023-05-03"
weight: 10
authors:
- Abdullah Al Shaad
tags:
- backup
- cli
- kubernetes
- redis
- restore
- stash
---

We are announcing Stash `v2023.04.30` which includes redis cluster mode backup and restore. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.04.30/README.md). In this post, we are going to highlight the changes.

### Redis Cluster Backup and Restore

We have added support for backup and restore for Redis Cluster Mode. [#164](https://github.com/stashed/redis/pull/164) .

Our latest release of Stash now supports backup and restore for all Redis modes and versions. Users can now take backups from any Redis mode and restore them to any other mode.

### Using Numeric UID for Docker Images

Previously, we were using user `nobody` in all Stash docker images. This will sometimes result is the following error message if PSP is enabled.

`Error: container has runAsNonRoot and image has non-numeric user (nobody), cannot verify user is non-root`

In this release, we have started using `user 65534` as the default uid in all Stash docker images to address this issue.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.03.20/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.03.20/setup/upgrade/).
- If you want to upgrade the Stash kubectl Plugin, please follow the instruction from [here](https://stash.run/docs/v2023.03.20/setup/install/kubectl-plugin/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
