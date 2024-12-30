---
title: Introducing KubeStash (aka Stash 2.0)
date: "2024-12-30"
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
enterprise customers have deepened, we've uncovered interesting use cases that the current Stash APIs cannot fully accommodate. To meet these challenges, we've introduced [KubeStash (aka Stash 2.0)](https://kubestash.com/). These 
new APIs are designed to augment Stash, enhancing its capabilities, fortifying its robustness, and significantly expanding its extensibility.

To avoid confusion between Stash and KubeStash (aka Stash 2.0) due to their similar names, we use KubeStash to refer to Stash 2.0.

### Feature Comparison

Many noticeable changes have been made in KubeStash compared to Stash. The declarative APIs have undergone drastic changes. The api groups for KubeStash live under `kubestash.com` domain. You may find that some Custom Resource Definitions (CRDs) have similar names, but their use cases have changed in most instances. To learn about the KubeStash declarative api visit [here](https://kubestash.com/docs/latest/concepts/#declarative-api). Let's discuss the comparison in below table:

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

Stash is hte most widely deployed backup/restore solution offered by AppsCode. We intend to offer support for it for 3 more years (till Dec 31, 2027). We plan to perform on going releases to go with the Kubernetes release and bug fixes as necessary. We also plan to update restic version used by Stash.

Given the api groups are different for CRDs, the recommended migration process from Stash 1.0 to KubeStash is to install both of them side-by-side. Once you have sufficient backup data using KubeStash, remove the backup data from Stash and uninstall it from the cluster. So, if you are already using Stash and want to migrate to KubeStash, then follow the following steps:

- [Install KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) and run Stash and KubeStash backup simultaneously for all target applications.
- [Uninstall Stash](https://stash.run/docs/latest/setup/uninstall/stash/) once the KubeStash backup for all target applications has run to the length of the retention policy.
- Clean up the backend data that was taken backup using Stash.

Now, you are ready to go with only KubeStash. This might not be ok for those who have very large existing backups using Stash. Please reach out to support@appscode.com or directly if that is your use-case.

If you are not using Stash, then just [install KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/).

> It is recommended to use `KubeStash` for new projects and cluster, if possible.

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).
