---
title: Comparison between Stash and Stash 2.0 (aka KubeStash)
date: "2024-04-08"
weight: 10
authors:
- Md Ishtiaq Islam
tags:
- backup
- disaster-recovery
- kubernetes
- kubestash
- restore
- stash
---

### Overview

[Stash](https://stash.run/), a cloud-native data backup and recovery solution tailored for Kubernetes workloads, has been in operation for several years. However, as our interactions with an expanding base of 
enterprise customers have deepened, we've uncovered interesting use cases that the current Stash APIs cannot fully accommodate. To meet these challenges, we've introduced [Stash 2.0 (aka KubeStash)](https://kubestash.com/) APIs. These 
new APIs are designed to augment Stash, enhancing its capabilities, fortifying its robustness, and significantly expanding its extensibility.

With the introduction of Stash 2.0 (aka KubeStash) APIs, some customers have expressed confusion between Stash and KubeStash. Therefore, we are writing this post to address the confusion and provide a clear comparison between the two.

To avoid confusion between Stash and Stash 2.0 due to their similar names, we have decided to use KubeStash instead of Stash 2.0 in this blog post.

### Feature Comparison

Many noticeable changes have been made in KubeStash compared to Stash. Let's discuss them below:

- The declarative APIs have undergone drastic changes. You may find that some Custom Resource Definitions (CRDs) have similar names, but their use cases have changed in most instances. To learn about the KubeStash declarative api visit [HERE](https://kubestash.com/docs/v2024.3.16/concepts/#declarative-api).
- In Stash, you need to create a [BackupConfiguration](https://stash.run/docs/v2024.4.8/concepts/crds/backupconfiguration/) and a [Repository](https://stash.run/docs/v2024.4.8/concepts/crds/repository/) for each application you want to backup. If you need multiple schedules or backends for one application, you'll have to create several BackupConfigurations and Repositories accordingly. However, in KubeStash, you only need to create one [BackupConfiguration](https://kubestash.com/docs/v2024.3.16/concepts/crds/backupconfiguration/), which allows you to specify multiple backends and schedules as needed. 
- To perform backup and restore operations Stash injects a [sidecar or init-container](https://stash.run/docs/v2024.4.8/guides/workloads/overview/) to the target if it is workload (i.e. `Deployment`, `DaemonSet`, `StatefulSet` etc.). This causes the workload to restart, which is generally not preferred by users in most cases. To mitigate this issue, KubeStash introduces the [Job Model](https://kubestash.com/docs/v2024.3.16/guides/workloads/overview/) for performing backup and restore operations, eliminating the need for injecting sidecar/init containers. 
- In Stash, you need to specify the backend information (where the backed up data will be stored) in each Repository. This means repeating the backend information in every Repository, even if it's the same backend. To simplify this, KubeStash introduces the [BackupStorage](https://kubestash.com/docs/v2024.3.16/concepts/crds/backupstorage/) CRD. It holds backend information for a specific backend and can be referenced in multiple Repositories. This allows users to reuse the same backend within a namespace or across the cluster.
- KubeStash can sync your Repositories between your storage backend and your cluster. Now, when you create a `BackupStorage` object with respective backend information, KubeStash will automatically create the respective Repositories object that were backed up in that storage. This makes restoring into a new cluster much more convenient as you no longer need to create the respective repository manually.
- In KubeStash, we've introduced a new CRD named [Snapshot](https://kubestash.com/docs/v2024.3.16/concepts/crds/snapshot/) which represents the state of a backup run for one or more components of an application. For every `BackupSession`, it creates `Snapshot` CRs. If a `BackupSession` involves multiple repositories, a `Snapshot` is created for each repository. This makes the task easier to restore specific snapshot and components of an application. Restoring a specific snapshot in Stash was more complex and tiresome process.
- In KubeStash, we've introduced a new CRD named [RetentionPolicy](https://kubestash.com/docs/v2024.3.16/concepts/crds/retentionpolicy/) that allows you to set how long you’d like to retain the backup data. It offers more control over snapshot cleanup policies compared to Stash. It is also reusable across namespaces. In contrast, in Stash, you need to provide the retention policy separately in every BackupConfiguration.
- In KubeStash, we've introduced a new CRD named [HookTemplate](https://kubestash.com/docs/v2024.3.16/concepts/crds/hooktemplate/) which specifies a template for an action that will be executed before or/and after backup/restore process. It provides better control over the backup process by letting you use custom variable in the `BackupBlueprint`. It is also reusable across namespaces. In contrast, in Stash, you need to specify the hook separately in every `BackupConfiguration`/`RestoreSession`.
- KubeStash supports Point-In-Time Recovery (PITR). This feature is not available in Stash.

These are some key features that have been improved or added in KubeStash. Additionally, there are numerous other features that have been enhanced or incorporated into KubeStash. To learn more visit [KubeStash docs](https://kubestash.com/docs/v2024.3.16/welcome/).

### Migrate from Stash to KubeStash

If you are already using Stash and want to migrate to KubeStash, then follow the following steps. 
- [Install KubeStash](https://kubestash.com/docs/v2024.3.16/setup/install/kubestash/) and run Stash and KubeStash backup simultaneously for all target applications.
- [Uninstall Stash](https://stash.run/docs/v2024.4.8/setup/uninstall/stash/) once the KubeStash backup for all target applications has run to the length of the retention policy.
- Clean up the backend data that was taken backup using Stash.

Now, you are ready to go with only KubeStash.

If you are not using Stash, then just [install KubeStash](https://kubestash.com/docs/v2024.3.16/setup/install/kubestash/).

> It is recommended to use `KubeStash` if you are looking for a cloud-native data backup and recovery solution for Kubernetes workloads.

### Upcoming Features in KubeStash

The following features are planned to incorporate into KubeStash in future releases.
- Support Backup Verification which will support automatic verification of backed up data, verification of subset of the sessions, verification of after every few backups, triggering backup verification manually, and application specific verification logic.
- Support for creating backups across Multiple Backends for same session.
- Support for multi-level restore. For example, deploy the application first, then restore data into it.
- Implement new addons according to the further requirements.

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
