---
title: Introducing Stash v2025.7.31
date: "2025-07-31"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- restic
- restore
- stash
- unlock
---

We are pleased to announce the release of [Stash v2025.7.31](https://stash.run/docs/v2025.7.31/setup/), packed with major improvement. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2025.7.31/README.md). In this post, we'll highlight the changes done in this release.

### Automatic `Restic` Unlocking — No More Manual Hassles

We’ve added a thoughtful little feature in this release that quietly takes care of something annoying: **locked Restic repositories** after a backup pod vanishes.

Sometimes, Kubernetes deletes backup pods mid-backup (maybe the node went down, resources ran out, or the autoscaler kicked in). When that happens, the `Restic` repository can get stuck in a locked state, blocking future backups until someone manually unlocks it. Not ideal.

But now, **Stash will automatically unlock the `Restic` repo** if it detects that the `BackupSession` failed because the pod disappeared. No more manual commands. No more wondering why your next backup won't start.

#### What’s better now:

1. **Auto-Unlock Magic** — If the `Restic` repo get locked, Stash will notice and unlock the repo for you.
2. **Smoother Experience** — Less manual cleanup, less friction. Backups just keep working.
3. **Less Downtime** — No waiting around or debugging why your backups are stuck.

---

### Better Handling of Deleted Backup Pods

Alongside the unlock improvement, we’ve also fixed a pesky issue with how `BackupSession` statuses were handled.

Previously, if a backup pod disappeared mid-backup (common in dynamic Kubernetes environments), the `BackupSession` might hang forever in a `Running` state, even if the backup never finished. That’s now fixed.

From this release on, **Stash will properly mark such `BackupSessions` as `Failed`**, so you have clearer visibility into what went wrong and when.


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).


### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).

