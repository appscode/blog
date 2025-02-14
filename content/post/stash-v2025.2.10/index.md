---
title: Introducing Stash v2025.2.10
date: "2025-02-10"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- restore
- stash
---

We are pleased to announce the release of [Stash v2025.2.10](https://stash.run/docs/v2025.2.10/setup/), packed with major improvement. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2025.2.10/README.md). In this post, we'll highlight the changes done in this release.

### Upgrade Restic Version to `0.17.3`

With this release, we have upgraded Restic from version `0.13.1` to `0.17.3`, introducing support for repository compression. To take advantage of this feature, the repository must be migrated to version `v2`. By default, if no migration is performed, the repository remains in version `v1`.

To simplify the migration process, we have introduced two new Stash CLI commands: [migrate](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#migrate-repository) and [prune](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#prune). These commands help transition repositories seamlessly while ensuring data integrity and efficiency.

Steps to upgrade a Restic repository to `v2`:

1. **Pause the Corresponding Backup:** Before upgrading, pause the backup associated with the repository. You can do this using the [pause backup](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#pause-backup) command provided by the Stash kubectl plugin.
2. **Run the Migration Command:** Next, execute the [migrate](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#migrate-repository) command provided by the Stash kubectl plugin. This command will first check the repository’s integrity and then upgrade its format to version 2. Note that if any issues are found during the integrity check, they must be resolved before the migration can proceed.
3. **Run the Prune Command:** After a successful migration, use the [prune](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#prune) command provided by Stash kubectl plugin to compress the repository metadata. If you want to limit the amount of data rewritten in a single operation, use the `--max-repack-size` flag with the `prune` command.
4. **Resume the Corresponding Backup:** Now resume the backup associated with the repository. You can do this using the [resume backup](https://stash.run/docs/v2025.2.10/guides/cli/kubectl-plugin/#resume-backup) command provided by the Stash kubectl plugin.

Keep in mind that the contents of files already stored in the repository will not be rewritten during the upgrade. Only data from new backups will be compressed. Over time, more and more of the repository will be automatically compressed as new backups are added.

> Upgrading repository format version will take some time depending on the repository size. Repository issues must be corrected before upgrading. It is recommended to contact with Stash team before upgrading repository.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
