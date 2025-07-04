---
title: Introducing Stash v2025.2.10
date: "2025-06-30"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- cli
- disaster-recovery
- kubernetes
- restore
- stash
---

We are pleased to announce the release of [Stash v2025.6.30](https://stash.run/docs/v2025.6.30/setup/), packed with major improvement. You can check out the full changelog [HERE](https://github.com/stashed/CHANGELOG/blob/master/releases/v2025.6.30/README.md). In this post, we'll highlight the changes done in this release.

### New Features

In this release, we've introduced a new feature `TaskQueue`. It can be enabled while installing or upgrading Stash. The `TaskQueue` feature maintains all triggered `BackupSessions` and executes them one by one based one maximum concurrency limit in your cluster. It's act as a queue, ensuring that `BackupSessions` are executed in the order they are received.

### New Feature: `TaskQueue`

We're excited to introduce a new feature in this release **`TaskQueue`**. This feature can be enabled during the installation or upgrade of Stash.

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
$ helm upgrade stash oci://ghcr.io/appscode-charts/stash \
      --- \
      -- set global.taskQueue.enabled=true \
      -- set global.taskQueue.maxConcurrentSessions=<max_concurrent_sessions> \
      --- 
```

Here, 
- `global.taskQueue.enabled` is set to `true` to enable the `TaskQueue` feature.
- `global.taskQueue.maxConcurrentSessions` is set to define the maximum number of concurrent `BackupSessions` that can be executed at a time.


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [HERE](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [HERE](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
