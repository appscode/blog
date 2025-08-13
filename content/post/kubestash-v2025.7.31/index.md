---
title: Introducing KubeStash v2025.7.31
date: "2025-07-31"
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

We are pleased to announce the release of [KubeStash v2025.7.31](https://kubestash.com/docs/v2025.7.31/setup/), packed with major improvements. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.7.31/README.md). In this post, we’ll highlight the changes done in this release.

---

### Introduced manifest restore, view commands to CLI


Previously, we've introduced `kubedump-restore` for manifest-based **selective resource restoration**.

Now in this release, we’ve brought these capabilities directly into the CLI with `manifest-restore` and `manifest-view` commands to make restores more **accessible, safer, and easier to manage**, providing:

- Give users **better control** over which resources and namespaces are restored.
- **Dry-run validation** before applying changes to live clusters

---

#### `manifest-view`
You can now inspect the contents of a snapshot before restoring.

```bash
kubectl kubestash manifest-view \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --exclude-resources="endpointslices.discovery.k8s.io,endpoints"
```

#### Example output:
```bash 
┌─ .
└─ 
   └─ kubestash-tmp
      └─ manifest
         ├─ deployments.apps
         │  └─ namespaces
         │     └─ demo-a
         │        └─ my-deployment-a.yaml
         ├─ controllerrevisions.apps
         │  └─ namespaces
         │     └─ demo-b
         │        └─ my-statefulset-7bc9c486fc.yaml
         ├─ services
         │  └─ namespaces
         │     ├─ demo-a
         │     │  └─ my-service-a.yaml
         │     └─ demo-b
         │        └─ my-service-b.yaml
         ...
         ├─ serviceaccounts
         │  └─ namespaces
         │     ├─ demo-a
         │     │  └─ my-serviceaccount-a.yaml
         │     └─ demo-b
         │        └─ my-serviceaccount-b.yaml
         └─ configmaps
            └─ namespaces
               ├─ demo-a
               │  └─ my-config-a.yaml
               └─ demo-b
                  └─ my-config-b.yaml
# (output truncated)
```
---

#### `manifest-restore` with `--dry-run-dir`
Preview the restore process without impacting your cluster.

```bash
kubectl kubestash manifest-restore \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --exclude-resources="pods,nodes.metrics.k8s.io,nodes,pods.metrics.k8s.io,metrics.k8s.io,endpointslices.discovery.k8s.io" \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --dry-run-dir="/home/nipun/Downloads" \
  --max-iterations=5
```

#### Benefits

- **Downloads manifests locally** – Safely downloads resource manifests on your local machine without interacting with the live cluster.
- **No real changes to the cluster** – Review and validate manifests before applying them, ensuring zero risk during the verification phase.
- **Perfect for verifying large-scale restores** – Ideal for testing full-cluster restore scenarios without impacting production workloads.

---

#### `manifest-restore` without `--dry-run-dir`
Once you’ve verified via dry-run, you can perform the actual restore:

```bash
kubectl kubestash manifest-restore \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --exclude-resources="nodes.metrics.k8s.io,nodes,pods.metrics.k8s.io,metrics.k8s.io" \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --max-iterations=5
```

#### Restore Behavior:

- Restores **CRDs** first to ensure all custom resources can be created successfully.
- Automatically creates **namespaces** if they do not already exist.
- Uses a **multi-iteration restore process** to prevent resource dependency issues from blocking restoration.
- In early iterations, resources are created as **orphans** (without owner references) to avoid dependency loops.
- After all iterations, **owner references** are updated for all resources to match the live cluster state.

---

### Automatic `Restic` Unlocking — No More Manual Hassles

We’ve added a thoughtful little feature in this release that quietly takes care of something annoying: **locked Restic repositories** after a backup pod vanishes.

Sometimes, In Kubernetes cluster, a node might crash, resources could become scarce, or the cluster autoscaler might decide to reschedule workloads.
In such cases, Kubernetes may terminate a backup pod while it's still running, even if the backup hasn’t finished. This can leave the backup incomplete and, in the case of `Restic`, the repository locked, blocking future backups until someone manually unlocks it. Not ideal.

But now, **KubeStash will automatically unlock the `Restic` repo** if it detects that the `BackupSession` failed because the pod disappeared. No more manual commands. No more wondering why your next backup won't start.

#### What’s better now:

1. **Auto-Unlock Magic** — If the `Restic` repo get locked, KubeStash will notice and unlock the repo for you.
2. **Smoother Experience** — Less manual cleanup, less friction. Backups just keep working.
3. **Less Downtime** — No waiting around or debugging why your backups are stuck.

---

### Improvements and Bug Fixes

- **Automatic owner reference updates** in kubedump to simplify dependency handling.
- **Multiple-iteration restore process** in kubedump to avoid resource creation being blocked due to dependencies.
- **Support for group resources** in filtering flags, alongside individual resources in kubedump.
  - **Example:**
    ```bash
    --include-resources="deployments,clusterroles.rbac.authorization.k8s.io"
    --exclude-resources="endpointslices,nodes.metrics.k8s.io,nodes"
    ```
- **Updated `label-selectors` flag** to support filtering with `key`.
  - **Previous format:** `"key1:value1,key2:value2"`
  - **New format:** `"key1:value1,key2:value2,key3"`
  - **Example:**
    ```bash
    --and-label-selectors="app:my-app,app:my-sts,app"
    ```

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.7.31/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.7.31/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

