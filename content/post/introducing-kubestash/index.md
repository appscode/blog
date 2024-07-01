---
title: Introducing KubeStash (aka Stash 2.0)
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

Many noticeable changes have been made in KubeStash compared to Stash. The declarative APIs have undergone drastic changes. You may find that some Custom Resource Definitions (CRDs) have similar names, but their use cases have changed in most instances. To learn about the KubeStash declarative api visit [HERE](https://kubestash.com/docs/v2024.3.16/concepts/#declarative-api). Let's discuss the comparison in below table:

| Description                                                                                                                                                       | Stash    | KubeStash (aka Stash 2.0) |
|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------------------------|
| **Backup & Restore Kubernetes Volumes:** You can backup & restore Kubernetes volumes mounted inside workloads as well as the standalone PVCs.                     | &#10003; | &#10003;                  |
| **Backup & Restore Databases running inside Kubernetes:** You can backup & restore the databases running inside Kubernetes cluster.                               | &#10003; | &#10003;                  |
| **Point-In-Time Recovery:** You can provide a timestamp during databases restore up-to where you want to restore.                                                 | &#10007; | &#10003;                  |
| **Multiple Schedules:** You can specify multiple schedules for a single backup process.                                                                           | &#10007; | &#10003;                  |
| **Multiple Repository:** You can provide multiple `Repository` reference in a single `BackupConfiguration`                                                        | &#10007; | &#10003;                  |
| **Reusable Backend:** You can store storage info in a `BackupStorage` object and refer it into multiple Repositories.                                             | &#10007; | &#10003;                  |
| **Automatic Repository synchronization:** The system can sync your Repositories between your storage backend and your cluster.                                    | &#10007; | &#10003;                  |
| **Custom Variable in Auto-backup:** You can use custom variable in the `BackupBlueprint` to gain better control over backup process.                              | &#10007; | &#10003;                  |
| **Reusable Hook:** You can provide `HookTemplate` that let you re-use the hooks across different backup process.                                                  | &#10007; | &#10003;                  |
| **Reusable Retention Policy** You can provide `RetentionPolicy` that let you re-use the retention policies for your cluster and across different backup processes | &#10007; | &#10003;                  |

### Improvements

- **Repository Structure:** We have overhauled Repository structure in this new API to support multiple backup driver. Now, your backed up data, application manifest can live into the same repository. Each repository stores a meta file that holds the respective snapshots info.
- **Snapshot Structure:** Previous, different component of your application required different snapshots. As a result, a single backup session results multiple snapshot which was difficult to work with if you wanted to restore an earlier version of your data rather than the latest one. We have overhauled the Snapshot structure too. Now, a single Snapshot refer to a logical state of your application. It can contain multiple components of your application as well as your application manifest files.
- **Workload Addon Mechanism:** To perform backup and restore operations Stash injects a [sidecar or init-container](https://stash.run/docs/v2024.4.8/guides/workloads/overview/) to the target if it is workload (i.e. `Deployment`, `DaemonSet`, `StatefulSet` etc.). This causes the workload to restart, which is generally not preferred by users in most cases. To mitigate this issue, KubeStash introduces the [Job Model](https://kubestash.com/docs/v2024.3.16/guides/workloads/overview/) for performing backup and restore operations, eliminating the need for injecting sidecar/init containers. 

While the above highlights some of the key features and improvements of KubeStash, explore the KubeStash [documentation](https://kubestash.com/docs/v2024.3.16/welcome/) for a comprehensive view of all the new features.

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
- **Backup Verification:** No backup is admissible until you can successfully restore your application from it. So, it is necessary to ensure that your application is recoverable from the backed up data. Support for automatic verification of your backed up data will be added where it automatically spins up a temporary instance of your application, restore data into it, run some checks, and then removes the temporary instance.
- **Application Level Backup:** If you deploy, your application using any package manager like Helm or an operator, backing up only the application manifest and the data is not enough. You can not just re-create your application from the backed up manifest. Instead, your restored application should be managed by the same package manager or operator you deployed originally. Support for taking backup of the relative resources based on application manager will be added so that it can restore in the same way as you deployed originally.
- **New Addons:** Support backup and restore for new databases (i.e. MS SQL, RabbitMQ, etc.) will be added.

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
