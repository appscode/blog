---
title: Introducing Stash v2024.8.27
date: "2024-08-27"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubedump
- kubernetes
- postgres
- restore
- stash
- tls
---

We are pleased to announce the release of [Stash v2024.8.27](https://stash.run/docs/v2024.8.27/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2024.4.8/README.md).  In this post, we'll highlight the changes done in this release.

### New Features

In this release, we've introduced new commands to the `Stash kubectl Plugin` for improved debugging and management of Restic repositories.

1. `kubectl stash check` command allows you to verify the integrity and consistency of your Restic repository. It's particularly useful for detecting issues caused by faulty storage or unauthorized modifications to your repository files, which could lead to data restoration problems.

We recommend running this check regularly to ensure your backup data remains intact and secure. For more details on how the check command works, refer to the Restic [documentation](https://restic.readthedocs.io/en/latest/045_working_with_repos.html#checking-integrity-and-consistency).

**Format:**
```bash
$ kubectl stash check <repository-name> [flags]
```
**Example:**
```bash
 $ kubectl stash check myr-repo --namespace=demo
```

2. `kubectl stash rebuild-index` command helps you rebuild the index of your Restic repository. In cases where the repository index is damaged, making backups or restores impossible, the rebuild-index command can recreate the index based on the existing pack files in the repository, restoring functionality. For more details on how the rebuild-index command works, refer to the Restic [documentation](https://restic.readthedocs.io/en/v0.13.1/manual_rest.html).

**Format:**
```bash
$ kubectl stash rebuild-index <repository-name> [flags]
```
**Example:**
```bash
$ kubectl stash rebuild-index myr-repo --namespace=demo
```

### Bug fixes

