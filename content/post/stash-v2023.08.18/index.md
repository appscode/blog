---
title: Introducing Stash v2023.08.18
date: "2023-08-22"
weight: 10
authors:
- Hossain Mahmud
tags:
- backup
- kubernetes
- mongodb
- mysql
- restore
- stash
---

We are announcing Stash `v2023.08.18` which includes a few bug fixes and Improvments. You can find the complete changelog [here](https://github.com/stashed/CHANGELOG/blob/master/releases/v2023.08.18/README.md). In this post, we are going to highlight the changes.

### Bug Fixes
- Resolved a bug causing MySQL `5.x.x` restoration failure in group replication mode. The issue occurred while restoring databases with *MyISAM* storage engine tables. To address this, we now exclude system databases from the backup process.
  ([#198](https://github.com/stashed/mysql/pull/716))
- Resolved a bug that was causing issues with shard backups in MongoDB versions `3.x.x`, `4.x.x`, and `6.x.x`. In the last release, some accidental changes were made. This new release fixes everything and makes sure things work correctly.
- Resolved the issue of *`_mergeAuthzCollections.tempRolesCollection` missing as a required field* during restoration, specifically for early versions like `5.0.2` and `5.0.3`. We've now split the addon into two versions: one for early versions (`5.0.2`, `5.0.3`) and another for later versions (`5.0.15`).

### Improvements

We've enhanced the way you manage licenses in Stash Helm Charts. Instead of providing the license directly as a value, you can now reference it by its secret name. The license itself should be placed under the key `key.txt` withing the secret. ([#651](https://github.com/kubedb/installer/pull/651))

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/v2023.03.20/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/v2023.03.20/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
