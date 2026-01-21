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
- Documentation improved with clarifications around database backup & restore and a new manifest-based cluster resource workflow.
- Installer and charts: generated certified charts, stricter semver for certified charts, many `CVE` report updates, chart tests and better test logging.
- Improved compatibility and packaging for `Red Hat` certification (published images in several repos).

--- 

### What's New

#### AWS Credential Manager

We added a mutating admission webhook in aws-credential-manager that validates S3/S3-compatible bucket access for `credentialless` (IRSA) EKS setups. When KubeStash creates backup or restore Jobs, the webhook injects an init-container that runs a bucket-access test before the Job’s main containers start. If the test fails, the Job is prevented from proceeding, surfacing misconfigured credentials or permission issues early and reducing failed backups.

##### Two tunable flags control the init-container behavior

- `--aws-max-interval-seconds` (default: 5) — retry interval between access attempts.
- `--aws-max-wait-seconds` (default: 300) — overall timeout for the access test.

The init-container retries at the configured interval until access succeeds or the total wait is exceeded—adjust these to balance retry aggressiveness and overall timeout for high-latency or eventually-consistent storage backends.

### Documentation update
- Updated the cluster-resources guide with a manifest-based "Full Cluster Backup & Restore" workflow and a concise "Keep in mind" note clarifying backup tasks. See the details [here](https://kubestash.com/docs/v2026.1.19/guides/cluster-resources/full-cluster-backup-and-restore/#keep-in-mind).

---

### Improvements and Bug fixes

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

