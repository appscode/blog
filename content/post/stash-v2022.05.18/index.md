---
title: Introducing Stash v2022.05.12
date: 2022-05-18
weight: 10
authors:
  - Piyush Kanti Das
tags:
  - kubernetes
  - stash
  - backup
  - restore
  - docs
  - nats
  - elasticsearch
---

We are very excited to announce Stash `v2022.05.18`.  In this release, we have added exciting new features and improvements. We have squashed a few bugs as well. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2022.05.18/README.md). We are going to highlight the major changes in this post.

## New Features

Here, we are going to highlight the new features that have been introduced in this release.

### Add support for Elasticsearch 8.2.0

In this release, we have added Elasticsearch 8.2.0 support to our Stash Elasticsearch add-on.


### Add Nat's Account Backup Support

We have added Account Backup Support to our Nat's Addon.

## Bug Fixes and Improvements

We have fixed some bugs and did some upgrade. Here are a few notable changes,

- Fix: ImagePullSecrets not passing to the backup job properly [Fix #1445](https://github.com/stashed/stash/pull/1445)
- Improvement: Updated BackupConfiguration webhook to make the target immutable [Improvement #1444](https://github.com/stashed/stash/pull/1444)


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2022.05.18/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2022.05.18/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
