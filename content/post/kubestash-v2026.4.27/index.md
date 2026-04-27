---
title: Introducing KubeStash v2026.4.27
date: "2026-04-27"
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

We are pleased to announce the release of [KubeStash v2026.4.27](https://kubestash.com/docs/v2026.4.27/setup/). This release delivers secure and seamless backup & restore across the KubeStash ecosystem, with major improvements for credential-less operation on AWS and Azure. You can check out the full changelog [HERE](https://github.com/kubestash/CHANGELOG/blob/master/releases/v2026.4.27/README.md). In this post, we'll highlight the major changes.

---

### Quick highlights
- Extended AWS credential-less mode with dynamically created IAM roles to overcome the trust policy character quota limit on production clusters.
- Added job suspension via webhook to prevent backup jobs from running before AWS IAM permissions are fully propagated.
- Introduced Azure credential-less backup & restore on AKS using Azure's Identity Binding feature to overcome the 20 federated credential limit.
- Manually fetch and inject AWS credentials in IRSA-enabled clusters to ensure restic can access S3 reliably.
- Fixed `BackupSession` status remaining stuck in `Running` state on retention job failure.

---

### What's New

#### Overcoming AWS IAM Trust Policy Quota with Seed Roles

KubeStash's AWS credential-less mode uses IRSA (IAM Roles for Service Accounts) to grant backup jobs access to S3 without credentials. However, AWS IAM Role trust policies have a [quota of 2048 characters](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html#reference_iam-quotas-entities), which allows only roughly ~20 service accounts per trust policy. In a production cluster with hundreds of databases—each with its own `BackupConfiguration`—this limit is quickly exhausted, since every scheduled backup job creates a distinct service account.

To overcome this constraint, KubeStash now creates IAM roles dynamically using a **seed role** as a template. The seed role carries the cluster's OIDC provider configuration and permission set; each derived role gets a numeric suffix (e.g., `kubestash-selfhost-i`) and holds a small batch of service accounts in its trust policy.

**Seed role trust policy example**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::748392615204:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringLike": {
          "oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E:sub": [
            "system:serviceaccount:kubestash:kubestash-operator-operator"
          ],
          "oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
```

**Derived role trust policy example** (the `-i` suffix indicates the i-th dynamically created role from the seed)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::748392615204:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringLike": {
          "oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E:sub": [
            "system:serviceaccount:demo:pg-backup-1772006820",
            "system:serviceaccount:demo:mongo-backup-177300719"
          ],
          "oidc.eks.us-east-1.amazonaws.com/id/9F3A7C1B5D8E42A6C91E0F2B4A7D6C8E:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
```

**How it works**

1. The KubeStash operator annotates each backup job's service account with `go.klusters.dev/seed-role-name`, pointing to the seed role ARN (e.g., `go.klusters.dev/seed-role-name: arn:aws:iam::748392615204:role/kubestash-selfhost`).
2. The credential manager checks whether any existing derived role has capacity to accommodate the new service account.
3. If none does, it creates a new derived role cloned from the seed role—same OIDC provider and permissions—with a numeric suffix (e.g., `kubestash-selfhost-7`).
4. The new service account is added to that derived role's trust policy and annotated with the assigned role ARN (e.g., `eks.amazonaws.com/role-arn: arn:aws:iam::748392615204:role/kubestash-selfhost-7`).
5. By default, at most `100` derived roles can be created from a single seed; this ceiling is configurable via the `roleCreationLimit` flag.

This strategy scales credential-less backup to production clusters with hundreds of `BackupConfiguration` resources without hitting the trust policy character quota.

---

#### Job Suspension

The AWS credential manager watches all jobs created by KubeStash (identified by the `"app.kubernetes.io/managed-by": "kubestash.com"` label). If a job's service account does not yet have the required IAM permissions at the time the job is created, the job is **suspended** via a webhook. Once the service account is granted the proper permissions, the job is automatically **unsuspended** and proceeds normally.

This prevents a race condition where a backup job starts executing before its IAM role has been fully provisioned.

---

#### Azure Credential-less Mode

KubeStash now supports credential-less backup & restore on AKS clusters using Azure's [Identity Binding](https://learn.microsoft.com/en-us/azure/aks/identity-bindings) feature—an extension of Azure Workload Identity introduced specifically to address the [20 federated credential limit](https://learn.microsoft.com/en-us/azure/aks/identity-bindings-concepts).

With this mode, no explicit secret needs to be referenced in the `BackupStorage`:

**Sample `BackupStorage`**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: azure-storage
  namespace: default
spec:
  storage:
    provider: azure
    azure:
      storageAccount: nipunstorageaccount7620
      container: container7620
      prefix: nipun-identity-binding
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: WipeOut
```

Notice that no `secretName` is specified—KubeStash resolves Azure Blob Storage credentials through the managed identity.

**How it works**

For any job that needs to access Azure Blob Storage, KubeStash adds the following annotations to the job and its service account:

```yaml
klusters.dev/azure-mi-name: "<MI_NAME>"
klusters.dev/azure-rg-name: "<RESOURCE_GROUP>"
klusters.dev/azure-subscription-id: "<SUBSCRIPTION_ID>"
```

The `azure-credential-manager` then:

1. Watches service accounts that carry the above annotations.
2. Annotates the service account with the managed identity's tenant ID and client ID:

```yaml
azure.workload.identity/client-id: <MI_CLIENT_ID>
azure.workload.identity/tenant-id: <MI_TENANT_ID>
```

3. Adds the required labels and annotations to service accounts and job pod templates for identity binding:

```yaml
labels:
  azure.workload.identity/use: "true"
annotations:
  azure.workload.identity/use-identity-binding: "true"
```

4. Configures RBAC to grant the service account permission to use the managed identity:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: use-mi-<MI_CLIENT_ID>
rules:
  - verbs: ["use-managed-identity"]
    apiGroups: ["cid.wi.aks.azure.com"]
    resources: ["<MI_CLIENT_ID>"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: use-mi-<MI_CLIENT_ID>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: use-mi-<MI_CLIENT_ID>
subjects:
  - kind: ServiceAccount
    name: pg-backup-1772006820
    namespace: demo
```

The backup pod will have the following Azure environment variables and projected token volume mounts injected automatically:

```bash
$ kubectl get pod <kubestash-operator-pod> -n <kubestash-namespace> -o yaml | grep AZURE
      - name: AZURE_CLIENT_ID
      - name: AZURE_TENANT_ID
      - name: AZURE_FEDERATED_TOKEN_FILE
      - name: AZURE_AUTHORITY_HOST
      - name: AZURE_KUBERNETES_TOKEN_PROXY
      - name: AZURE_KUBERNETES_SNI_NAME
      - name: AZURE_KUBERNETES_CA_FILE
```

---

## Bug Fix and Improvements

### Manually Set AWS Credentials in IRSA-enabled Clusters**

In IRSA-enabled clusters, restic may fail to fetch credentials from the service account token, causing errors like:
`s3.getCredentials: no credentials found`.

Since restic relies on `minio-go-client` instead of the official AWS SDK, fixing this at the source is complex. As a workaround, we now manually retrieve and inject AWS credentials (with retries) so restic can reliably access S3.

PR Link: https://github.com/kubestash/apimachinery/commits/master/

---

### Fix BackupSession Status Running

`BackupSession` could previously remain stuck in the `Running` state if the retention policy job failed.
Now, on retention job failure, a condition is set with an appropriate message, allowing the `BackupSession` controller to correctly mark the status as `Failed`.

PR Link: https://github.com/kubestash/kubestash/pull/353

---

## What Next?
Please try the latest release and give us your valuable feedback.

- If you want to install KubeStash in a clean cluster, please follow the installation instruction from [HERE](https://kubestash.com/docs/v2026.4.27/setup/install/kubestash/).
- If you want to upgrade KubeStash from a previous version, please follow the upgrade instruction from [HERE](https://kubestash.com/docs/v2026.4.27/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).
