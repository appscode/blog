---
title: Introducing KubeStash v2025.10.17
date: "2025-10-17"
weight: 10
authors:
- Arnab Baishnab Nipun
tags:
- backup
- backup-verification
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of [KubeStash v2025.10.17](https://kubestash.com/docs/v2025.10.17/setup/), featuring enhanced documentation and stability improvements.

You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.10.17/README.md). In this post, we'll walk you through the highlights of this release.

---

### Documentation Updates: Manifest-based Backup & Restore of Cluster Resources

This release includes comprehensive documentation for manifest-based backup and restore operations, helping you protect and recover your cluster resources with greater flexibility.

- [Overview](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/overview/)
- [Configure Storage and RBAC](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/configure-storage-and-rbac/)
- [Backup Filtering Options](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/backup-filtering-options/)
- [Restore Filtering Options](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/restore-filtering-options/)
- [Filtering Demonstration](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/filtering-demonstration/)
- [Full Cluster Backup & Restore](https://kubestash.com/docs/v2025.10.17/guides/cluster-resources/full-cluster-backup-and-restore/)

---

### Improvements and Bug Fixes

In managed Kubernetes clusters, it’s common for certain **resources or API groups to be restricted** — meaning you might not have permission to create or restore them. In earlier versions, this could lead to unnecessary errors or confusion during restore operations.

Starting from this release, **KubeStash now automatically excludes these resources** from backup and restore workflows, ensuring a smoother experience and better compatibility with managed environments.

#### What’s New

- #### Automatic Resource Exclusion
  KubeStash now **skips any resources that do not support the `create` verb**, as these cannot be restored in new cluster.
  
  This optimization ensures backups include only restorable resources, reducing restore-time errors and improving reliability. 
  

- #### Default Exclusions for System-Managed Resources
   By default, the following resources are now excluded from backup and restore operations:
   - `nodes` 
   -  `endpointslices.discovery.k8s.io`

  These objects are dynamically managed by Kubernetes and do not need to be backed up.

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.10.17/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.10.17/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

