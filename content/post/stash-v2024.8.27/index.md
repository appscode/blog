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
- mongodb
- postgresql
- restore
- stash
---

We are pleased to announce the release of [Stash v2024.8.27](https://stash.run/docs/v2024.8.27/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2024.4.8/README.md). In this post, we'll highlight the changes done in this release.

### New Features

In this release, we've introduced new commands to the `stash kubectl Plugin` for improved debugging and management of restic repositories.

**Repository Check Command**

`kubectl stash check` command allows you to verify the integrity and consistency of your restic repository. It's particularly useful for detecting issues caused by faulty storage or unauthorized modifications to your repository files, which could lead to data restoration problems.

We recommend running this check regularly to ensure your backup data remains intact and secure. For more details on how the check command works, refer to the [restic documentation](https://restic.readthedocs.io/en/latest/045_working_with_repos.html#checking-integrity-and-consistency).

**Format:**
```bash
$ kubectl stash check <repository-name> [flags]
```

**Example:**
```bash
 $ kubectl stash check myr-repo --namespace=demo
```

### Improvements & Bug fixes

- We've improved the postBackup hook execution in the Job model for certain scenarios. If the postBackup hook doesn’t run (e.g., due to a timeout or backup disruption, etc.), the Stash operator will handle it instead of the backup job. However, if the preBackup hook is configured but doesn’t execute or fails the postBackup hook will not be run.

- We've fixed a bug where the backup job could remain active even after the deadline was exceeded (based on the Timeout) with the BackupSession marked as failed. Now, Stash will suspend the backup job when the deadline is exceeded.

- We've fixed a bug that caused the BackupSession to remain in the `running` phase even if the backup didn't complete within the deadline or the backup container failed with an error (for job model). 

- We've fixed a bug for backup and restore of externally hosted/managed `Postgres` and `MongoDB` databases.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).


### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).

