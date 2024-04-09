---
title: Comparing Stash and Stash 2.0 (aka KubeStash)
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
- In Stash, you need to create a [BackupConfiguration](https://stash.run/docs/v2024.4.8/concepts/crds/backupconfiguration/) object and a [Repository](https://stash.run/docs/v2024.4.8/concepts/crds/repository/) object for each target application to take backup. To provide different schedule or multiple repository for a target application, you need to create more `BackupConfiguration` and `Repository` for a target application according to the requirement. In KubeStash, you need to create only [BackupConfiguration](https://kubestash.com/docs/v2024.3.16/concepts/crds/backupconfiguration/) object for a target application. In the `BackupConfigution`, you can configure multiple sessions (multiple schedule) and multiple repositories for a target application as per requirement. 
- To perform backup and restore process Stash injects a [sidecar or init-container](https://stash.run/docs/v2024.4.8/guides/workloads/overview/) to the target if it is workload (i.e. `Deployment`, `DaemonSet`, `StatefulSet` etc.). This causes the workload to restart, which is generally not preferred by users in most cases. To mitigate this issue, KubeStash introduces [Job Model](https://kubestash.com/docs/v2024.3.16/guides/workloads/overview/) to perform backup and restore process. 
- In Stash, user has to provide the backend (where the backed up data will be stored) specification in `Repository` CR. Since, `Repository` CR has one-to-one mapping with the target application, user has to repeat the backend information in every `Repository` CR they create in the cluster. To solve this issue, we've introduced [BackupStorage](https://kubestash.com/docs/v2024.3.16/concepts/crds/backupstorage/) CR in KubeStash and moved the backend information into it. This will allow the users re-use the backend within a namespace or within the cluster.
- In KubeStash, the Repositories and Snapshots linked to a `BackupStorage` is uploaded in the backend as metadata. When a `BackupStorage` is created with existing backup data, KubeStash automatically synchronizes the Repositories and Snapshots linked to it from the backend. There is no such feature in Stash.
- In KubeStash, we've introduced a new CRD named [Snapshot](https://kubestash.com/docs/v2024.3.16/concepts/crds/snapshot/) which represents the state of a backup run for one or more components of an application. For every `BackupSession`, it creates `Snapshot` CRs. If a `BackupSession` involves multiple repositories, a `Snapshot` is created for each repository. This makes the task easier to restore specific snapshot and components of an application. To restore specific snapshot in Stash was more complex and tiresome process.
- In KubeStash, we've introduced a new CRD named [RetentionPolicy](https://kubestash.com/docs/v2024.3.16/concepts/crds/retentionpolicy/) which specifies how the old Snapshots should be cleaned up. This is re-usable across namespaces. By this, users are able to specify that they want to remove backup data that are older than particular duration. In Stash, there is no option for re-usable retention policy.
- In KubeStash, we've introduced a new CRD named [HookTemplate](https://kubestash.com/docs/v2024.3.16/concepts/crds/hooktemplate/) which basically specifies a template for an action that will be executed before or/and after backup/restore process. This is re-usable across namespaces. In Stash, there is no option for re-usable hook.
- KubeStash supports Point-In-Time Recovery (PITR). This feature is not available in Stash.

These are some key features that have been improved or added in KubeStash. Additionally, there are numerous other features that have been enhanced or incorporated into KubeStash. To learn more visit [KubeStash docs](https://kubestash.com/docs/v2024.3.16/welcome/).

### Migrate from Stash to KubeStash

If you are already using Stash and want to migrate to KubeStash, then follow the following steps. 
- [Install KubeStash](https://kubestash.com/docs/v2024.3.16/setup/install/kubestash/) and run Stash and KubeStash backup simultaneously for all target applications.
- [Uninstall Stash](https://stash.run/docs/v2024.4.8/setup/uninstall/stash/) once the KubeStash backup for all target applications has run to the length of the retention policy.
- Clean up the backend data that was taken backup using Stash.

Now, you are ready to go with only KubeStash.

If you are not using Stash, then just [install KubeStash](https://kubestash.com/docs/v2024.3.16/setup/install/kubestash/).

> It is recommended to use KubeStash over Stash.

### Upcoming Features in KubeStash

The following features are planned to incorporate into KubeStash in future releases.
- Backup Verification
- Support for creating backups across Multiple Backends
- Support for multi-level restore
- Implement new addons according to the users requirements

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
