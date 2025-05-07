---
title: Introducing KubeStash v2025.4.30
date: "2025-04-30"
weight: 10
authors:
- Md Anisur Rahman
tags:
- backup
- backup-verification
- disaster-recovery
- kubernetes
- kubestash
- restore
---

We are pleased to announce the release of [KubeStash v2025.4.30](https://kubestash.com/docs/v2025.4.30/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.4.30/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### KubeDB managed Cassandra Backup/Restore

KubeDB now supports backup and restore for Cassandra databases using KubeStash and the Medusa plugin. This enables reliable data protection with cloud storage backends (e.g., S3, GCS).

**Backup and Restore Workflow**

- **BackupStorage**: Specifies the cloud storage backend.
- **RetentionPolicy**: Defines how long backup data is retained.
- **Secrets**: Stores backend access credentials.
- **BackupConfiguration**: Configures the target database, backend, and addon.
- **RestoreSession**: Restores data from a specified snapshot.

In below given the example of a `BackupConfiguration` for a `Cassandra` database

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: cass-backup
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
---
      addon:
        name: cassandra-addon
        tasks:
          - name: logical-backup
        jobTemplate:
          spec:
            serviceAccountName: cluster-resource-reader
```

And for `Cassandra` database below is the example of a `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-cas
  namespace: default
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: default
    name: cass-sample
---
  addon:
    name: cassandra-addon
    tasks:
      - name: logical-backup-restore
    jobTemplate:
      spec:
        serviceAccountName: cluster-resource-reader
```

### Introduce Workload Manifest Backup/Restore along with RBAC resources

KubeStash now supports backing up and restoring Kubernetes workload manifests, including associated RBAC resources such as Role, RoleBinding, ClusterRole, and ClusterRoleBinding. This enhancement enables users to capture not just application data but also the complete configuration of workloads for comprehensive disaster recovery.

Here is an example of `BackupConfiguration`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
---
sessions:
  - name: workload-backup
    addon:
      name: workload-addon
      tasks:
        - name: manifest-backup
          params:
            includeRBACResources: "true"
      jobTemplate:
        spec:
          serviceAccountName: cluster-resource-reader
```
Here,
- Set `includeRBACResources: "true"` to include `RBAC` resources in the backup.
- A custom `serviceAccountName` with appropriate cluster-level RBAC permissions is required for accessing these resources.


Here is an example of `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
---
addon:
  name: workload-addon
  tasks:
    - name: manifest-restore
      params:
        includeRBACResources: "true"
        overrideResources: "true"
  jobTemplate:
    spec:
      serviceAccountName: cluster-resource-reader
```
Here,
- `includeRBACResources: "true"` restores the `RBAC` resources associated with the workloads.
- `overrideResources: "true"` ensures existing resources in the cluster are replaced with the restored versions.


> By default, both `includeRBACResources` and `overrideResources` are set to false. Enable them explicitly if needed for your use case.

### Improvements & Bug Fixes

#### Improved Pod Discovery for Deployment Backup/Restore
Previously, the system assumed that each Deployment would have only one associated ReplicaSet, which is not always true. We've updated the logic to identify the most recent ReplicaSet to accurately locate the active Pod during backup and restore operations.

#### Fixed Workload-Only Manifest Backup/Restore
Resolved an issue where backups and restores targeting only workload manifests were not functioning correctly. This fix ensures reliable handling of workload-only manifest operations.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

