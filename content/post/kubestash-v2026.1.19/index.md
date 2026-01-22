---
title: Introducing KubeStash v2026.1.19
date: "2026-01-19"
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

We are pleased to announce the release of [KubeStash v2026.1.19](https://kubestash.com/docs/v2026.1.19/setup/), packed with improvements across the KubeStash ecosystem that packed with improvements across the KubeStash ecosystem that increase compatibility with newer Kubernetes releases, improve reliability of backups and restores, and advance our certified chart and image publishing work. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.1.19/README.md). In this post, we’ll highlight the changes done in this release.

---

### Quick highlights
- aws-credential-manager: added a mutating webhook to validate bucket access on `credentialless` (IRSA) EKS setups.
- Kubernetes client libraries upgraded to Kubernetes 1.34 in many components for better forward compatibility.
- Image references moved to fully-qualified docker image strings where code expects them.
- Documentation improved with clarifications manifest-based cluster resource backup & restore.
- Installer and charts: generated certified charts, stricter semver for certified charts, many `CVE` report updates, chart tests and better test logging.
- Improved compatibility and packaging for `Red Hat` certification (published images in several repos).

--- 

### What's New

#### AWS Credential Manager

We added a mutating admission webhook in aws-credential-manager that injects an `init-container` to the Jobs for checking access to S3 buckets for `credentialless` (IRSA) EKS setups.

The workflow:

- The KubeStash operator creates Jobs (Backup, RetentionPolicy, Restore) and the corresponding ServiceAccounts with required annotations.
- The operator's ServiceAccount must include the annotation:
  `eks.amazonaws.com/role-arn: arn:aws:iam::<ACCOUNT_ID>:role/<role-name>` for IRSA (credentialless) access. This annotation is set by the cluster administrator.
- When the role ARN annotation is present, KubeStash propagates a second annotation to Job ServiceAccounts listing required buckets:
  `go.klusters.dev/bucket-names: <bucket1,bucket2,...>`. Bucket names are derived from all `BackupStorage` resources referenced by the `BackupConfiguration`.
- The AWS Credential Manager updates the IAM role trust policy to allow the annotated ServiceAccount to assume the role.
- When a Job ServiceAccount carries the `go.klusters.dev/bucket-names` annotation, the webhook injects an init-container that verifies access to each bucket before the Job's main containers run. The init-container retries at a configured interval until access succeeds or a total timeout is reached. On success the Job proceeds; on failure the init-container exits non‑zero and the Job is prevented from running, surfacing misconfigured credentials or permissions early. 

     This mechanism was added to address a race condition: the AWS Credential Manager updates the IAM role trust policy after Jobs and ServiceAccounts are created, and `cloud latency` can cause pods to start before the ServiceAccount has permission to assume the role, leading to authentication errors. The injected init-container blocks the main containers until bucket access succeeds (or the timeout elapses), preventing premature start and noisy failures.

Tunable flags (defaults shown):
- `--aws-max-interval-seconds` (default: 5) — retry interval between access attempts.
- `--aws-max-wait-seconds` (default: 300) — overall timeout for the access test.

This behavior reduces failed backups by detecting credential, permission, or cloud-latency issues before workload containers start.


### Documentation update
-  Updated the cluster-resources guide with a manifest-based ***Full Cluster Backup & Restore***, including a concise ***Keep in mind*** note to clarify the distinction between database backup/restore and cluster resources. See the details [here](https://kubestash.com/docs/v2026.1.19/guides/cluster-resources/full-cluster-backup-and-restore/#keep-in-mind).

---

### Improvements and Bug fixes

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

