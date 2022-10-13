---
title: Introducing Stash v2022.07.09
date: "2022-07-09"
weight: 10
authors:
- Piyush Kanti Das
tags:
- backup
- hooks
- kubernetes
- restore
- stash
---

We are announcing Stash `v2022.07.09` which introduces a new feature and a few improvements.

You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.07.09/README.md).

We are going to highlight the major changes in this post.

## New Features

Here is the new feature of this release:

- **Add Hook ExecutionPolicy**: We have added `executionPolicy` to our PostBackupHook and PostRestoreHook. You can specify any one of these three executionPolicy - 'Always', 'OnFailure', 'OnSuccess' along with the Hooks. The default executionPolicy is 'Always'. If you use 'OnSuccess' executionPolicy, the Hooks will be executed only after a successful backup or restore. On the other hand, 'onFailure' will execute the Hooks after failed backup or restore only.

## Improvements

We have also made some improvements. Here are the notable changes,

- Ensured that the backupsession will be marked as completed after executing all the steps [#176](https://github.com/stashed/apimachinery/pull/176).
- Added RBAC permissions for finalizer [#1458](https://github.com/stashed/stash/pull/1458).
- Custom labels will be passed to Volume Snapshotter Job properly [#1461](https://github.com/stashed/stash/pull/1461).

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.07.09/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.07.09/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
