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

We are pleased to announce the release of [KubeStash v2026.2.26](https://kubestash.com/docs/v2026.2.26/setup/). This release includes improvements across the KubeStash ecosystem with a focus on better modularity around the restic integration, smoother cloud identity workflows (AWS/GCP), and reliability fixes around retention policy scheduling. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.2.26/README.md). In this post, we’ll highlight the major changes.

---

### Quick highlights
- Added backup and restore support for a `MongoDB` instance deployed as a `StatefulSet` within the same Kubernetes cluster.
- Refactored restic integration into a standalone package and updated dependent components accordingly.
- Improved cloud identity workflows by propagating cloud and bucket annotations to backup/restore Jobs and ServiceAccounts.
- Fixed retention policy pods getting stuck in the `Pending` state by surfacing scheduling reasons in status.

---

### What’s New

#### MongoDB Backup & Restore for Non-KubeDB Managed Instances [PR](https://github.com/kubedb/mongodb-restic-plugin/pull/112)

KubeStash can now back up and restore MongoDB instances that are running inside a Statefulset in the same cluster and not managed by KubeDB.

**Sample `AppBinding`**

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mongodb-appbinding
  namespace: demo
spec:
  type: mongodb
  version: "8.2.5"
  clientConfig:
    service:
      name: mongodb
      port: 27017
      scheme: mongodb
  secret:
    name: mongodb-secret
```

**Sample `BackupConfiguration`**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mongo-sts-backup
  namespace: demo
spec:
  target:
    apiGroup: appcatalog.appscode.com
    kind: AppBinding
    name: mongodb-appbinding
    namespace: demo
  sessions:
  - addon:
      name: mongodb-addon
      tasks:
      - name: logical-backup
```

**Backup pod logs (example)**

```bash
$ kubectl logs <backup-pod-name> -n demo -f --all-containers

I0304 05:15:18.949337       1 unlock.go:158] All repositories are ready for new operations.
I0304 05:15:18.949370       1 commands.go:200] Backing up stdin data
[golang-sh]$ mongodump --host mongodb.demo.svc --archive --username=***** --password=***** --authenticationDatabase admin...
2026-03-04T05:15:19.020+0000	writing playground1.users to archive on stdout
2026-03-04T05:15:19.022+0000	writing playground.users to archive on stdout
2026-03-04T05:15:19.024+0000	writing playground2.users to archive on stdout
2026-03-04T05:15:19.025+0000	done dumping playground1.users (10 documents)
2026-03-04T05:15:19.026+0000	done dumping playground2.users (1 document)
2026-03-04T05:15:19.033+0000	writing playground2.products to archive on stdout
2026-03-04T05:15:19.035+0000	done dumping playground.users (3 documents)
2026-03-04T05:15:19.035+0000	done dumping playground2.products (3 documents)
I0304 05:15:25.797832       1 commands.go:425] sh-output: {"message_type":"summary","files_new":1,"files_changed"...}
```

**Sample `RestoreSession`**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-mongodb
  namespace: demo
spec:
  target:
    apiGroup: appcatalog.appscode.com
    kind: AppBinding
    name: mongodb-appbinding
    namespace: demo
  addon:
    name: mongodb-addon
    tasks:
    - name: logical-backup-restore
      params:
        args: --nsInclude=playground1.* --nsInclude=playground2.*
```

**Restore pod logs (example)**

```bash
$ kubectl logs <restore-pod> -n demo -f --all-containers

I0304 05:06:16.955899       1 commands.go:274] Dumping backed up data
2026-03-04T05:06:22.214+0000	reading metadata for playground1.users from archive on stdin
2026-03-04T05:06:22.214+0000	reading metadata for playground2.products from archive on stdin
2026-03-04T05:06:22.214+0000	reading metadata for playground2.users from archive on stdin
2026-03-04T05:06:22.235+0000	restoring playground1.users from archive on stdin
2026-03-04T05:06:22.247+0000	finished restoring playground1.users (10 documents, 0 failures)
2026-03-04T05:06:22.258+0000	restoring playground2.users from archive on stdin
2026-03-04T05:06:22.270+0000	finished restoring playground2.users (1 document, 0 failures)
2026-03-04T05:06:22.289+0000	restoring playground2.products from archive on stdin
2026-03-04T05:06:22.300+0000	finished restoring playground2.products (3 documents, 0 failures)
2026-03-04T05:06:22.512+0000	no indexes to restore for collection playground1.users
2026-03-04T05:06:22.512+0000	no indexes to restore for collection playground2.users
2026-03-04T05:06:22.512+0000	no indexes to restore for collection playground2.products
2026-03-04T05:06:22.512+0000	14 document(s) restored successfully. 0 document(s) failed to restore.
```

> Note: The examples above are intended to illustrate the workflow. You may need to adjust `version`, connection settings, and addon/task parameters based on your MongoDB deployment and your backup requirements.

#### Standalone Restic Package [PR](https://github.com/kubestash/apimachinery/pull/198)

The restic integration layer has been refactored into a standalone package. This helps keep the restic-related logic more modular and makes it easier for multiple components to share the same implementation.

The release also adopts `gomodules.xyz/restic v0.2.0` ([PR](https://github.com/kubestash/apimachinery/pull/210)). Dependent components—`cli`, `kubedump`, `manifest`, `pvc`, `workload`, `kubestash`, and related database plugins—have been updated to use the new standalone package.

#### Cloud & Bucket Annotation Propagation [PR](https://github.com/kubestash/kubestash/pull/330)

KubeStash now propagates cloud and bucket annotations to backup/restore Jobs and their ServiceAccounts. This is intended to reduce manual setup when using cloud-native identity mechanisms such as:

- **AWS IRSA** (S3 access through IAM roles for service accounts)

---

### Improvements and Bug Fixes

#### RetentionPolicy Pod Stuck in Pending State [PR](https://github.com/kubestash/kubestash/pull/332)

Previously, in some cases retention policy pods could remain in a `Pending` state for a long time, which could block subsequent scheduled backups.

This release improves detection of pod pending reasons ([PR](https://github.com/kubestash/apimachinery/pull/200)). The controller now records the relevant scheduling condition in the `BackupSession` status and fails the current retention policy Job when appropriate, so that future schedules can proceed.

```bash
$ kubectl get backupsession.core.kubestash.com <backupsession-name> -n <namespace> -oyaml

apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
...
status:
  conditions:
  - lastTransitionTime: "2026-02-06T07:03:42Z"
    message: '0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector.
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