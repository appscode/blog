---
title: Introducing KubeStash v2025.6.30
date: "2025-06-30"
weight: 10
authors:
- Not specified 
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


### Backup and Restore Manifests of All Kubernetes Resources

KubeStash now supports backing up and restoring **manifests of all cluster resources**, including fine-grained filtering capabilities.

> You can now filter resources more specifically during backup using include/exclude flags and label selectors.
> - By default, include flags are set to *, meaning all resources are considered.
> - Exclude flags and label selectors are empty by default, which means no filtering is applied. 

Here is an example of `BackupConfiguration`:

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

> You have now enhanced control over resource restore operations:
> - RestorePVs: Whether to restore PersistentVolumes.
> - OverrideResources: Whether to override existing resources.
> - StorageClassMappings: Specify mappings from old to new storage classes (e.g., old1=new1,old2=new2).

Here is an example of `RestoreSession`:

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


## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.6.30/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

