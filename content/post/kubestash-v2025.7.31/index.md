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

We are pleased to announce the release of [KubeStash v2025.7.31](https://kubestash.com/docs/v2025.7.31/setup/), packed with new features. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.7.31/README.md).


---

## What's New in v2025.7.31


## New CLI Features

In **v2025.6.30**, we introduced `kubedump-restore` for manifest-based **selective resource restoration**.
In **v2025.7.31**, we’ve taken it to the next level with:

- **New CLI commands** for controlled restores
- **Dry-run restore mode** for safe validation before applying
- **Improved dependency resolution** with **automatic owner reference updates** during restore
- **Support for group resources** in filtering flags, alongside individual resources
---

### **View Snapshot Contents — `manifest-view`**
You can now inspect the contents of a snapshot before restoring.

```bash
kubectl kubestash manifest-view \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --include-namespaces="*" \
  --include-resources="*" \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --exclude-resources="endpointslices.discovery.k8s.io,endpoints" \
  --v=5
```

### Example output: 
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

### **Dry-Run before Restore — `manifest-restore` with `--dry-run-dir`**
Preview the restore process without impacting your cluster.

```bash
kubectl kubestash manifest-restore \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --include-namespaces="demo-b,demo-a" \
  --include-resources="*" \
  --exclude-resources="pods,nodes.metrics.k8s.io,nodes,pods.metrics.k8s.io,metrics.k8s.io,endpointslices.discovery.k8s.io" \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --dry-run-dir="/home/nipun/Downloads" \
  --v=5 \
  --max-iterations=5
```

## Benefits

- **Downloads manifests locally** – Safely downloads resource manifests on your local machine without interacting with the live cluster.
- **No changes to the cluster** – Review and validate manifests before applying them, ensuring zero risk during the verification phase.
- **Perfect for verifying large-scale restores** – Ideal for testing full-cluster restore scenarios without impacting production workloads.

---

### **Full Restore — Apply Manifests to Cluster**
Once you’ve verified via dry-run, you can perform the actual restore:

```bash
kubectl kubestash manifest-restore \
  --snapshot=azure-repo-cluster-resources-backup-1754997440 \
  --namespace=demo \
  --include-namespaces="demo-b,demo-a" \
  --include-resources="*" \
  --exclude-resources="pods,nodes.metrics.k8s.io,nodes,pods.metrics.k8s.io,metrics.k8s.io,endpointslices.discovery.k8s.io" \
  --include-cluster-resources=true \
  --and-label-selectors="app" \
  --v=5 \
  --max-iterations=5
```

## Restore Behavior:

- Restores **CRDs** first to ensure all custom resources can be created successfully.
- Automatically creates **namespaces** if they do not already exist.
- Uses a **multi-iteration restore process** to prevent resource dependency issues from blocking restoration.
- In early iterations, resources are created as **orphans** (without owner references) to avoid dependency loops.
- After all iterations, **owner references** are updated for all resources to match the live cluster state.
---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.7.31/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.7.31/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

