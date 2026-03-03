---
title: Introducing KubeStash v2026.2.26
date: "2026-02-26"
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

We are pleased to announce the release of [KubeStash v2026.2.26](https://kubestash.com/docs/v2026.2.26/setup/), bringing significant architectural improvements, smarter cloud integration, and critical reliability fixes. This release focuses on making backup operations more observable, the restic layer more modular, and cloud-native workflows across AWS and GCP smoother than ever. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.2.26/README.md). In this post, we'll walk through the highlights.

---

### Quick highlights
- Restic package extracted into a standalone module with a cleaner, more robust design.
- Cloud and bucket annotation propagation now works seamlessly for both AWS (S3/IRSA) and GCP (GCS/Workload Identity).
- Feature reporting via site info for better operational insights.
- Fixed retention policy pods getting stuck in a pending state.

---

### What's New

#### Standalone Restic Package [PR](https://github.com/kubestash/apimachinery/pull/198)

The restic integration layer has been refactored into a fully standalone package with an improved design:

* **Independent & testable** — The restic package can now be used and tested in isolation, without pulling in the entire apimachinery dependency tree.
* **Cleaner interfaces** — Internal APIs have been redesigned for better separation of concerns.
* **Upstream alignment** — The final release adopts `gomodules.xyz/restic v0.2.0` ([PR](https://github.com/kubestash/apimachinery/pull/210)) as the upstream library.

All dependent components — `cli`, `kubedump`, `manifest`, `pvc`, `workload`, `kubestash`, and all the `KubeDB` managed database's backup `plugin` — have been updated to use the new standalone package.

#### Cloud & Bucket Annotation Propagation [PR](https://github.com/kubestash/kubestash/pull/330)

Running KubeStash on managed Kubernetes with cloud-native identity (AWS IRSA, GCP Workload Identity) just got easier.

KubeStash now automatically propagates cloud and bucket annotations to backup/restore Jobs and their ServiceAccounts. Here's what changed:

* **S3 & GCS bucket annotations** are now generated from BackupStorage references and attached to Job ServiceAccounts.
* **GCP IAM annotations** are propagated for GCP Workload Identity setups ([PR](https://github.com/kubestash/apimachinery/pull/209)).

This means credential and access propagation in cloud-native backup workflows is now automatic — no more manually wiring annotations on every ServiceAccount.

---

### Improvements and Bug Fixes

#### RetentionPolicy Pod Stuck in Pending State [PR](https://github.com/kubestash/kubestash/pull/332)

Previously, retention policy pods could get stuck in a `Pending` state indefinitely — silently blocking subsequent backup schedules. The root cause was that the controller had no way to detect *why* a pod wasn't scheduling.

This release adds proper type and reason detection for pod pending states ([PR](https://github.com/kubestash/apimachinery/pull/200)). The controller now identifies the cause, updates the `BackupSession` appropriately, and makes the current `RetentionPolicy` job fail so that next scheduled backup can proceed.

```bash
$ kubectl get backupsession.core.kubestash.com <backupsession-name> -n <namespace> -oyaml

apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
...
status:
  conditions:
  - lastTransitionTime: "2026-02-06T07:03:42Z"
    message: '0/1 nodes are available: 1 node(s) didn''t match Pod''s node affinity/selector.
      preemption: 0/1 nodes are available: 1 Preemption is not helpful for scheduling.'
    reason: PodHasBeenInPendingStateForLongerThanExpected
    status: Unknown
    type: PodHasBeenInPendingStateForLongerThanExpected
  phase: Pending
```

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.2.26/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.2.26/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).