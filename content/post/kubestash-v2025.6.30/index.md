---
title: Introducing KubeStash v2025.6.30
date: "2025-06-30"
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

We are pleased to announce the release of [KubeStash v2025.6.30](https://kubestash.com/docs/v2025.6.30/setup/), packed with new features. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.6.30/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.


# Backup and Restore Manifests of All Kubernetes Resources

KubeStash now supports backing up and restoring **manifests of all cluster resources**, with fine-grained filtering options for precise control over what gets backed up or restored.

---

### Backup: 

You can now filter resources more specifically during backup using **include/exclude flags** and **label selectors**.

> - By default, `Include` flags are set to `*`, meaning all resources are included.  
> - `Exclude` flags and label selectors are **empty by default**.

#### Example `BackupConfiguration`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
---
addon:
  name: kubedump-addon
  tasks:
    - name: manifest-backup
      params:
        IncludeClusterResources: "true"
        IncludeNamespaces: "demo,kubedb,kubestash"
        ExcludeNamespaces: "default,kubevault"
        IncludeResources: "secrets,configmaps,deployments"
        ExcludeResources: "persistentvolumeclaims,persistentvolumes"
        ANDedLabelSelectors: "app:my-app"
        ORedLabelSelectors: "app1:my-app1,app2:my-app2"
  jobTemplate:
    spec:
      serviceAccountName: cluster-resource-reader-writter
```
### Restore: 

You have advanced control over how manifests are restored into the cluster.

> - `RestorePVs`: Whether to restore PersistentVolumes.  
> - `OverrideResources`: Whether to overwrite existing resources.  
> - `StorageClassMappings`: Map old storage classes to new ones (e.g., `old1=new1,old2=new2`).

#### Example `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
---
addon:
  name: kubedump-addon
  tasks:
    - name: manifest-restore
      params:
        IncludeClusterResources: "true"
        IncludeNamespaces: "*"
        ExcludeNamespaces: "kube-system"
        IncludeResources: "*"
        ExcludeResources: "deployments"
        ANDedLabelSelectors: "app:my-app"
        ORedLabelSelectors: "app1:my-app1,app2:my-app2"
        RestorePVs: "true"
        OverrideResources: "true"
        StorageClassMappings: "longhorn=openebs-hostpath"
  jobTemplate:
    spec:
      serviceAccountName: cluster-resource-reader-writter
```

> Note that include/exclude flags and label selectors are common for both backup and restore. 

# Introducing TaskQueue in KubeStash

We're excited to introduce a new feature in this release: **TaskQueue**.  
This feature can be enabled during the installation or upgrade of **KubeStash**.

---

## What is TaskQueue?

**TaskQueue** acts as a centralized controller that manages the execution of `BackupSessions` based on a defined **maximum concurrency limit**. It queues incoming `BackupSessions` and ensures they are processed in order—either one at a time or concurrently up to the specified limit.

---

## Why is TaskQueue Important?

In environments where multiple `BackupConfigurations` share the same schedule, all corresponding `BackupSessions` may be triggered simultaneously. Without **TaskQueue**, this can overwhelm the system, leading to **resource contention** and **backup failures**.

By enforcing a controlled execution flow, **TaskQueue** ensures:

-  The number of active `BackupSessions` never exceeds the defined concurrency limit.
-  BackupSessions are processed in order—queued and executed one after another or concurrently, as allowed.
-  Prevents simultaneous creation of backup jobs, reducing resource spikes.
-  Efficient utilization of cluster resources based on the concurrency limit.
-  Spreads backup execution over time, optimizing overall cluster resource usage.
-  Increases reliability by minimizing failures caused by resource exhaustion.

---

## How to Enable TaskQueue?

You can enable TaskQueue during installation or upgrade of **KubeStash** using the `--enable-task-queue` flag.

### Example:

```bash
helm upgrade kubestash oci://ghcr.io/appscode-charts/kubestash \
  --install \
  --set global.taskQueue.enabled=true \
  --set global.taskQueue.maxConcurrentSessions=<max_concurrent_sessions>


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

