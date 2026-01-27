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

We are pleased to announce the release of [KubeStash v2026.1.19](https://kubestash.com/docs/v2026.1.19/setup/), packed with improvements across the KubeStash ecosystem that increase compatibility with newer Kubernetes releases, improve reliability of backups and restores, and advance our certified chart and image publishing work. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.1.19/README.md). In this post, we’ll highlight the changes done in this release.

---

### Quick highlights
- aws-credential-manager: added a mutating webhook to validate bucket access on `credentialless` (IRSA) EKS setups.
- Kubernetes client libraries upgraded to Kubernetes 1.34 in many components for better forward compatibility.
- Image references moved to fully-qualified docker image strings where code expects them.
- Documentation improved with clarifications on manifest-based cluster resource backup & restore.
- Installer and charts: generated certified charts, stricter semver for certified charts, many `CVE` report updates, chart tests and better test logging.
- Improved compatibility and packaging for `Red Hat` certification (published images in several repos).

--- 

### What's New

#### Mutating Webhook for AWS IRSA

This release introduces a mutating admission webhook in `AWS Credential Manager` that injects an init-container into Jobs that access S3 buckets using IRSA authentication.
The init-container ensures bucket access is available before the Job starts, preventing failures caused by IAM propagation delays.

##### Responsibilities and Workflow

###### KubeStash [PR](https://github.com/kubestash/apimachinery/pull/196)

* KubeStash creates:
     * Backup/restore Jobs for accessing cloud buckets
     * Corresponding ServiceAccounts used by those Jobs
* The ServiceAccount must be pre-configured by the cluster administrator with the IRSA annotation:
    ```bash
    eks.amazonaws.com/role-arn: arn:aws:iam::<ACCOUNT_ID>:role/<role-name>
     ```
* When this role ARN annotation is present:
    * KubeStash derives bucket names from all referenced BackupStorage resources.
    * KubeStash propagates the following annotation to the Job ServiceAccount:
       ```bash
      go.klusters.dev/bucket-names: <bucket1,bucket2,...>
       ```

###### AWS Credential Manager

 * Watches for ServiceAccounts annotated with:
   ```bash
   go.klusters.dev/bucket-names
   ```

* Updates the `IAM role trust policy` to allow the annotated `ServiceAccount` to assume the IRSA role.

* Provides a `mutating admission webhook` that:
   * Detects Jobs whose ServiceAccount includes the bucket annotation
   * Injects an `init-container` into those Jobs
* Init-Container Behavior: Injected by AWS Credential Manager
    * Runs before the Job’s main containers
    * Verifies access to each annotated `S3 bucket`
* Race Condition Fix
    * There is an inherent race condition in IRSA-based workflows:
        * Jobs and ServiceAccounts may be created `before` The IAM role trust policy update fully propagates in AWS

Due to `cloud propagation latency`, pods may start before the `ServiceAccount` is allowed to assume the role, resulting in authentication failures. The injected init-container blocks `Job execution` until bucket access is confirmed.         

This behavior reduces failed backups by detecting credential, permission, or cloud-latency issues before workload containers start.

### Introduce Concurrent Worker Pool Pattern [PR](https://github.com/kubestash/apimachinery/commit/98bb7d6a)
In Kubestash the `Backupstorage` with `WipeOut` during the `Deletion` operation we introduced a `concurrent pool pattern` to handle the deletion of the backup data from the cloud.

* `MaxConnections`:
  * In  `BackupStorage` each cloud provider has field `MaxConnections` which is used to limit the number of concurrent connections to the cloud provider.
  * We use this field to delete the backup data from the cloud.
  * Previously we used to use only one worker to delete now, by default `10` workers will use to delete the backup data from the cloud.
  * User can configure this field using `MaxConnections` field in `BackupStorage` CRD under the `spec.provider.s3` section (Here, I used s3 as an example).

### Documentation Update
*  Updated the cluster-resources guide with a manifest-based `Full Cluster Backup & Restore`, including a concise `Keep in mind` note to clarify the distinction between database backup/restore and cluster resources. See the details [here](https://kubestash.com/docs/v2026.1.19/guides/cluster-resources/full-cluster-backup-and-restore/#keep-in-mind).

---

### Improvements and Bug Fixes

#### BackupSession/SnapShot Phase Stuck Due to Pod Eviction [PR](https://github.com/kubestash/kubestash/pull/310)

* Previously, if Job's (`Backup`/`RetentionPolicy`/`Restore`) pods were evicted or deleted (due to preemption, node-pressure), BackupSessions or SnapShots were stuck on `Running` phase as those wouldn't be marked `incomplete` by the Job controller. 
* Now the controller detects evicted Job pods and sets the appropriate incomplete  conditions ensuring `BackupSessions`, `SnapShots`, `RestoreSessions` reflects actual `failure` and allowing next scheduled Backup to trigger. 


#### Remove non-exclusive locks also while ensureNoExclusive locks [PR](https://github.com/kubestash/apimachinery/commit/e6109991)
  * Removing all stale locks (restic determines which locks are stale)
  * Checking if any exclusive locks remain (if they do, they're active)
  * Waiting for active exclusive locks to be released

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.1.19/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

