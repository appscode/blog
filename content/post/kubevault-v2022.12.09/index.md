---
title: Introducing KubeVault v2022.12.09
date: "2022-12-09"
weight: 25
authors:
- Sakib Alamin
tags:
- backup
- hashicorp
- kubernetes
- kubevault
- restore
- security
- stash
- vault
---

[KubeVault](https://kubevault.com) is a Git-Ops ready, production-grade solution for deploying and configuring [HashiCorp Vault](https://www.vaultproject.io/) on Kubernetes.

We are excited to announce the release of KubeVault `v2022.12.09`. You can now take `Backup` & `Restore` your `Vault` cluster managed by `KubeVault` or even deployed with ` Vault Helm Charts` using `Stash`.

`Stash` by `AppsCode` is a complete Kubernetes native disaster recovery solution for backup and restore your volumes and databases in Kubernetes on any public and private clouds.
Read more about `Stash` [here](https://stash.run/). `Stash` add-on for [Vault](https://github.com/stashed/vault) lets you take `Backup` snapshot & `Restore` it whenever required for a Vault cluster backed with `Raft` maintaining the `SOP` provided by Vault. 

Upon deploying `Vault` managed by `KubeVault`, an `AppBinding` will be created, which contains the necessary information needed for `Backup` & `Restore` process.
So, if you're deploying `Vault` using `Helm Charts`, you'll need to create the `AppBinding` yourself accordingly. 

An `AppBinding` is a Kubernetes `CustomResourceDefinition(CRD)` which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://blog.byte.builders/post/appbinding/).

Here's a sample `AppBinding` yaml created by the `KubeVault` operator on `VaultServer` deployment:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/instance: vault
    app.kubernetes.io/managed-by: kubevault.com
    app.kubernetes.io/name: vaultservers.kubevault.com
  name: vault
  namespace: demo
  ownerReferences:
  - apiVersion: kubevault.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: VaultServer
    name: vault
spec:
  appRef:
    apiGroup: kubevault.com
    kind: VaultServer
    name: vault
    namespace: demo
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: http
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    stash:
      addon:
        backupTask:
          name: vault-backup-1.10.3
          params:
          - name: keyPrefix
            value: k8s.kubevault.com.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
          - name: keyPrefix
            value: k8s.kubevault.com.demo.vault
    unsealer:
      mode:
        googleKmsGcs:
          bucket: vault-testing-keys
          credentialSecretRef:
            name: gcp-cred
          kmsCryptoKey: vault-testing-key
          kmsKeyRing: vault-testing
          kmsLocation: global
          kmsProject: appscode-testing
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller

```

So, in simple terms, all the information regarding your `Vault` deployment must be passed through the `AppBinding` itself.

For example:
- `spec.appRef` contains the application information for which this `AppBinding` is created. For `Vault Helm Charts` deployment, such field can be ignored. 
- `spec.parameters.backend` indicates the type of `Storage Backend` your `Vault` deployment is using.
- `spec.parameters.unsealer` contains the information about your unseal mode & secrets associated with it.

For, `Vault Helm Charts` deployment, you'll also need to create a `token` that has the necessary permission to take the `Backup` snapshot & `Restore` it.
You'll then need create a `Secret` with this `token` & provide its reference in the `spec.parameters.backupTokenSecretRef`. 

A sample policy may look like this:

```hcl
path "sys/storage/raft/snapshot" {
        capabilities = ["read"]
}

path "sys/storage/raft/snapshot-force" {
        capabilities = ["read"]
}

```

`KubeVault` takes care of all of these stuff itself, as it creates the `AppBinding` during the `VaultServer` deployment based on the provided configurations, making it hassle-free to manage your `Vault` life-cycle & do `Day-2` operations like `Backup` & `Restore`. 

The detailed commit by commit changelog can be found [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.12.09/README.md).

> We held a webinar on Vault `Backup` & `Restore`, take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/TOxufXiyVok" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.12.09/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
