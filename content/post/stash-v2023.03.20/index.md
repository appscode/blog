---
title: Introducing Stash v2023.03.20
date: "2023-03-20"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- cli
- kubernetes
- restore
- stash
---

We are announcing Stash `v2023.03.20` which includes a few bug fixes. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.03.20/README.md). In this post, we are going to highlight the changes.

### Bug Fixes
- Fixed a bug that was causing backup failures due to the inability to apply the retention policy to a locked repository ([#198](https://github.com/stashed/apimachinery/pull/198)).
- Fixed a bug that was preventing the `unlock` command from working on the local repository ([#179](https://github.com/stashed/cli/pull/179)).
- Fixed a bug that was causing Stash-specific imagePullSecrets to replace the existing imagePullSecrets of the workloads ([#1509](https://github.com/stashed/stash/pull/1509)).

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.03.20/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.03.20/setup/upgrade/).
- If you want to upgrade the Stash kubectl Plugin, please follow the instruction from [here](https://stash.run/docs/v2023.03.20/setup/install/kubectl-plugin/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
