---
title: Introducing KubeStash v2025.3.24
date: "2025-03-24"
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

We are pleased to announce the release of [KubeStash v2025.3.24](https://kubestash.com/docs/v2025.3.24/setup/), packed with new features and important bug fixes. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2025.3.24/README.md).

### New Features

Here, we are going to highlight the new features that have been introduced in this release.

#### Custom Restic Cache Volume Configuration

We've introduced functionality to pass any existing volume or volumeclaimTemplate in `backupConfiguration` or `restoreSession`. This volume will be attached with the backup/restore pod and used as a `Restic` cache volume.   

You can specify a `Restic` cache volume using either an existing `PVC` name or a `volumeClaimTemplate`:

**Use Existing PVCs**

Reference pre-provisioned `PVC` directly in your `BackupConfiguration or `RestoreSession`:
```yaml
addonVolumes:
  - name: ${RESTIC_CACHE_VOLUME}
    source:
      persistentVolumeClaim:
        claimName: my-cache-pvc  # Existing PVC****
```

**Dynamic Volume Provisioning**

Define a `volumeClaimTemplate` to let KubeStash automatically provision cache volumes:

```yaml
addonVolumes:
  - name: ${RESTIC_CACHE_VOLUME}
    source:
      volumeClaimTemplate:
        spec:
          accessModes: [ReadWriteOnce]
          resources:
            requests:
              storage: 1Gi
```

#### Automated AWS IRSA Annotation for Backup/Restore Jobs

Eliminate manual credential management with automatic IAM Role for Service Accounts (IRSA) propagation:

**How It Works**
Annotate `KubeStash Operator's` Service Account with AWS role ARN. `KubeStash` automatically injects these annotations into `backup/restore` Service Account, enabling secure access to S3 buckets without static credentials.

**Setup Guide**
- [Credentials with IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [How to create OIDC Provider](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)
- [Assigning IAM Role](https://docs.aws.amazon.com/eks/latest/userguide/associate-service-account-role.html)
- [Using IRSA with KubeStash on Amazon EK](https://kubestash.com/docs/v2025.3.24/guides/platforms/eks-irsa/)


### Improvements & Bug Fixes

#### KubeStash CLI Fix
Resolved an issue where `kubestash download` commands failed for `Google Cloud Storage` backends due to environment variable handling.

#### Multi-Region S3 Support
Fixed region detection for `IRSA-authenticated` S3 buckets. Previously only the `default region` worked; now all AWS regions are supported.

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2025.2.10/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

