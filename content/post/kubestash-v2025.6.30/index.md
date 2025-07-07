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

We’re excited to introduce some new features in this release. Some features can be enabled during the `installation` or `upgrade` of KubeStash & some will comes default, we’ll discuss each of this feature below.

---

# Enhanced Filtering in Backup Task

The `manifest-backup` task in the `kubedump-addon` now supports fine-grained filtering, providing precise control over which Kubernetes resources are included in a backup.

This feature helps you optimize storage usage, reduce restore noise, and back up only the components that matter most to your application.

---

### Newly Introduced Parameters

``` yaml 
- `ANDedLabelSelectors`
  - Usage: A set of labels, all of which need to be matched to filter the resources (comma-separated, e.g., `key1:value1,key2:value2`)
  - Default: ""
  - Required: false

- `ORedLabelSelectors`
  - Usage: A set of labels, at least one of which need to be matched to filter the resources (comma-separated, e.g., `key1:value1,key2:value2`)
  - Default: ""
  - Required: false

- `IncludeClusterResources`
  - Usage: Specify whether to restore cluster scoped resources
  - Default: "false"
  - Required: false

- `IncludeNamespaces`
  - Usage: Namespaces to include in backup (comma-separated, e.g., `demo,kubedb,kubestash`)
  - Default: "*"
  - Required: false

- `ExcludeNamespaces`
  - Usage: Namespaces to exclude from backup (comma-separated, e.g., `default,kube-system`)
  - Default: ""
  - Required: false

- `IncludeResources`
  - Usage: Resource types to include in backup (comma-separated, e.g., `pods,deployments`)
  - Default: "*"
  - Required: false

- `ExcludeResources`
  - Usage: Resource types to exclude from backup (comma-separated, e.g., `secrets,configmaps`)
  - Default: ""
  - Required: false
```

---

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
        IncludeNamespaces: "demo,kubedb,kubestash"
        IncludeResources: "secrets,configmaps,deployments"
        ORedLabelSelectors: "environment:prod,tier:db"
  jobTemplate:
    spec:
      serviceAccountName: cluster-resource-reader-writter
```

---

# Introducing Manifest Restore Support in `kubedump-addon`

We're excited to introduce the `manifest-restore` feature in the `kubedump-addon`, which brings full manifest-based restore support to KubeStash.

---

## Why `manifest-restore`?

Disaster can strike at any time — whether due to accidental deletion or infrastructure failure. The new `manifest-restore` task can **bring your Kubernetes cluster back to its previous state**, using the manifests captured in the backup snapshots.

---

# Restore Task

The newly introduced `manifest-restore` task in the `kubedump-addon` brings powerful restore capabilities to KubeStash. It allows you to restore previously backed-up Kubernetes manifests and apply them with fine-grained control over which resources to restore.

This feature is especially valuable in disaster recovery scenarios, where restoring cluster state accurately and efficiently is critical.

---

### Supported Parameters

``` yaml
- `OverrideResources`
  - Usage: Specify whether to override resources while restoring
  - Default: "false"
  - Required: false

- `RestorePVs`
  - Usage: Specify whether to restore PersistentVolumes
  - Default: "false"
  - Required: false

- `StorageClassMappings`
  - Usage: Mapping of old to new storage classes (e.g., `old1=new1,old2=new2`)
  - Default: ""
  - Required: false

```  

--- 

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
        IncludeNamespaces: "*"
        ExcludeNamespaces: "kube-system,default"
        RestorePVs: "true"
        StorageClassMappings: "longhorn=openebs-hostpath"
  jobTemplate:
    spec:
      serviceAccountName: cluster-resource-reader-writter
```
> Note that include/exclude flags and label selectors are common for both backup and restore tasks. 

--- 

# Introducing TaskQueue in KubeStash

We're excited to introduce another new feature in this release: **TaskQueue**.  
This feature can be enabled during the installation or upgrade of **KubeStash**.

---

## What is TaskQueue?

**TaskQueue** acts as a centralized controller that manages the execution of `BackupSessions` based on a defined **maximum concurrency limit**. It queues incoming `BackupSessions` and ensures they are processed in order—either one at a time or concurrently up to the specified limit.

---

## Why is TaskQueue Important?

In environments where multiple `BackupConfigurations` share the same schedule, all corresponding `BackupSessions` may be triggered simultaneously. Without **TaskQueue**, this can overwhelm the system, leading to **resource contention** and **backup failures**.

By enforcing a controlled execution flow, **TaskQueue** ensures:

- The number of active `BackupSessions` never exceeds the defined concurrency limit.
- `BackupSessions` are processed in order—queued and executed one after another or concurrently, as allowed.
- Prevents simultaneous creation of backup jobs, reducing resource spikes.
- Efficient utilization of cluster resources based on the concurrency limit.
- Spreads backup execution over time, optimizing overall cluster resource usage.
- Increases reliability by minimizing failures caused by resource exhaustion.

---

## How to Enable TaskQueue?

You can enable TaskQueue during installation or upgrade of **KubeStash** using the following Helm flags:

### Example:

```bash
helm upgrade kubestash oci://ghcr.io/appscode-charts/kubestash \
  --install \
  --create-namespace \
  --namespace kubestash \
  --set global.taskQueue.enabled=true \
  --set global.taskQueue.maxConcurrentSessions=<max_concurrent_sessions>
```
Here, 
- `global.taskQueue.enabled` is set to `true` to enable the `TaskQueue` feature.
- `global.taskQueue.maxConcurrentSessions` is set to define the maximum number of concurrent `BackupSessions` that can be executed at a time.

---

## How KubeStash Utilizes TaskQueue

Once enabled, **KubeStash** uses a dedicated controller called the **TaskQueueController**.

### It works as follows:

1. Instead of triggering `BackupSessions` directly, the `KubeStash` operator creates a resource called `PendingTask` for each `BackupSession`.
2. Each PendingTask contains the actual `BackupSession` and is monitored by the `TaskQueueController`.
3. The `TaskQueueController` processes these `PendingTask` resources based on the defined maximum concurrency limit.

**Example of `TaskQueue` YAML**

```yaml
apiVersion: batch.k8s.appscode.com/v1alpha1
kind: TaskQueue
metadata:
  name: appscode-kubestash-task-queue
spec:
  maxConcurrentTasks: 10
  tasks:
  - rules:
      failed: has(self.status.phase) && self.status.phase == 'Failed'
      inProgress: has(self.status.phase) && self.status.phase == 'Running'
      success: has(self.status.phase) && self.status.phase == 'Succeeded'
    type:
      group: core.kubestash.appscode.com
      kind: BackupSession
```
Here,
- `maxConcurrentTasks` is set to `10`, meaning a maximum of 10 `BackupSessions` can be executed concurrently.
- If you need to reconfigure the `TaskQueue` after enabling it, you can modify the `maxConcurrentTasks` value according to your cluster's capacity.

**Example of `PendingTask` YAML**

```yaml
apiVersion: batch.k8s.appscode.com/v1alpha1
kind: PendingTask
metadata:
  name: backupconfiguration-demo-s3-pvc-backup
spec:
  resource:
    metadata:
      name: s3-pvc-backup-2-1751630401
      namespace: demo
      ownerReferences:
      - apiVersion: core.kubestash.appscode.com/v1alpha1
        blockOwnerDeletion: true
        controller: true
        kind: BackupConfiguration
        name: s3-pvc-backup-2
        uid: c00376b7-1baf-4b7a-98df-c55f848e936c
    spec:
      invoker:
        apiGroup: core.kubestash.appscode.com
        kind: BackupConfiguration
        name: s3-pvc-backup
    status: {}
  taskType:
    group: core.kubestash.appscode.com
    kind: BackupSession
status:
  taskQueueName: appscode-kubestash-task-queue
```


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

