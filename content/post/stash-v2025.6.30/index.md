---
title: Introducing Stash v2025.6.30
date: "2025-06-30"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- pendingtask
- restore
- stash
- taskqueue
---

We are pleased to announce the release of [Stash v2025.6.30](https://stash.run/docs/v2025.6.30/setup/), packed with major improvement. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2025.6.30/README.md). In this post, we'll highlight the changes done in this release.

### New Feature: `TaskQueue`

We're excited to introduce a new feature in this release. This feature can be enabled during the installation or upgrade of `Stash`.

`TaskQueue` acts as a centralized controller that manages the execution of `BackupSessions` based on a defined **Maximum Concurrency Limit**. It queues incoming `BackupSessions` and ensures they are processed one at a time (or concurrently, up to the configured limit), maintaining their original order of arrival.

#### Why is `TaskQueue` important?

In environments where multiple `BackupConfigurations` share the same schedule, all corresponding `BackupSessions` may be triggered simultaneously. Without `TaskQueue`, this can overwhelm the system, leading to resource contention and backup failures.

By enforcing a controlled execution flow, `TaskQueue` ensures:

* Ensures the number of active `BackupSessions` never exceeds the defined concurrency limit.
* It maintains a `Queue` and processes `BackupSessions` in order, one after another or up to the allowed concurrency.
* Prevents simultaneous backup job creation, resources are used efficiently, reducing resource spikes.
* Cluster System resources are used efficiently, based on the concurrency limit.
* Optimizes resource usage across the cluster by distributing backup execution over time.
* Backups have a higher chance of completing successfully without overloading the cluster.
* Increases the reliability of backups by minimizing the risk of failure due to resource exhaustion.

#### How to Enable `TaskQueue`?

To enable `TaskQueue`, you can use the `--enable-task-queue` flag during the installation or upgrade of Stash. Here is an example:

```bash
$ helm install stash oci://ghcr.io/appscode-charts/stash \
  --version v2025.6.30 \
  --namespace stash --create-namespace \
  --set features.enterprise=true \
  --set global.taskQueue.enabled=true \
  --set global.taskQueue.maxConcurrentSessions=<max_concurrent_sessions> \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Here, 
- `global.taskQueue.enabled` is set to `true` to enable the `TaskQueue` feature.
- `global.taskQueue.maxConcurrentSessions` is set to define the maximum number of concurrent `BackupSessions` that can be executed at a time.

> Note: The `TaskQueue` feature is available **only in the Enterprise Edition** of Stash. To use this feature, make sure you install the Enterprise Edition.
> Installation instructions can be found [here](https://stash.run/docs/latest/setup/).

#### How Stash utilize `TaskQueue`?

When this feature enabled, `Stash` uses a separate controller called `TaskQueueController`,

**It works as follows:**

1. Instead of triggering `BackupSessions` directly, the `Stash` operator creates a resource called `PendingTask` for each `BackupSession`.
2. Each PendingTask contains the actual `BackupSession` and is monitored by the `TaskQueueController`.
3. The `TaskQueueController` processes these `PendingTask` resources based on the defined maximum concurrency limit.

**Example of `TaskQueue` YAML**

```yaml
apiVersion: batch.k8s.appscode.com/v1alpha1
kind: TaskQueue
metadata:
  name: appscode-stash-task-queue
spec:
  maxConcurrentTasks: 10
  tasks:
  - rules:
      failed: has(self.status.phase) && self.status.phase == 'Failed'
      inProgress: has(self.status.phase) && self.status.phase == 'Running'
      success: has(self.status.phase) && self.status.phase == 'Succeeded'
    type:
      group: stash.appscode.com
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
      - apiVersion: stash.appscode.com/v1beta1
        blockOwnerDeletion: true
        controller: true
        kind: BackupConfiguration
        name: s3-pvc-backup-2
        uid: c00376b7-1baf-4b7a-98df-c55f848e936c
    spec:
      invoker:
        apiGroup: stash.appscode.com
        kind: BackupConfiguration
        name: s3-pvc-backup
    status: {}
  taskType:
    group: stash.appscode.com
    kind: BackupSession
status:
  taskQueueName: appscode-stash-task-queue
```

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
