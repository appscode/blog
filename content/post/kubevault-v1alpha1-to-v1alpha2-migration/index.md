---
title: KubeVault v1alpha1 to v1alpha2 Migration
date: "2022-10-07"
weight: 20
authors:
- Sakib Alamin
tags:
- cert-manager
- cli
- hashicorp
- kubernetes
- kubevault
- kubevault-cli
- migration
- operator
- ops-manager
- secret-management
- security
- tls
- vault
---

[KubeVault](https://kubevault.com) is a Git-Ops ready, production-grade solution for deploying and configuring [HashiCorp Vault](https://www.vaultproject.io/) on Kubernetes.

# Background

This blog is based on our experience upgrading one of our `KubeVault Enterprise` users from KubeVault `v1alpha1` to KubeVault `v1alpha2`.
The same steps can be followed carefully to get your KubeVault deployment upgraded to the latest version.

## Why the Upgrade?

We wanted to ship all the new features & fixes to our client that were added in KubeVault `Operator`, v1alpha2 `APIs` & KubeVault `CLI`.
Also, the `TLS` certificates that were issued were about to expire after a year. So, in order to upgrade the `Operator`, `CLI` & `TLS` certificates, the upgrade was necessary.

Some KubeVault Operator v1alpha2 features are `Vault Ops-request`, `Cert-manager` support, `Recommendation` generation for managing Vault `TLS`, updated `Health Checker`, support for `Pod Disruption Budget`, updated root-token & unseal-keys naming convention for `multi-cluster` usage, etc.
And, from KubeVault `CLI` end you can now `get`, `set`, `delete`, `list`, and `sync` vault `unseal-keys` and `root-token` stored in various KMS stores.

If you want to learn more about KubeVault features, please check [this](https://kubevault.com) out.

# Upgrade Procedure

Prior to the upgrade, the KubeVault CLI that was in use was in version `v0.6.0-alpha.0` & the Operator version was `v2021.10.11`.

Here's the commands for installing the respective versions:

```bash
# to install KubeVault CLI v0.6.0-alpha.0

$ curl -o kubectl-vault.tar.gz -fsSL https://github.com/kubevault/cli/releases/download/v0.6.0-alpha.0/kubectl-vault-linux-amd64.tar.gz \
  && tar zxvf kubectl-vault.tar.gz \
  && chmod +x kubectl-vault-linux-amd64 \
  && sudo mv kubectl-vault-linux-amd64 /usr/local/bin/kubectl-vault \
  && rm kubectl-vault.tar.gz LICENSE.md

# to install KubeVault Operator v2021.10.11

$ helm install kubevault appscode/kubevault \
    --version v2021.10.11 \
    --namespace kubevault --create-namespace \
    --set kubevault-operator.apiserver.enableValidatingWebhook=false \
    --set-file global.license=/path/to/vault/license.txt
```

The `Unsealer` mode used by them was `AWS KMS` & `Storage` used as `AWS S3`. The `VaultServer` configuration used is similar to the one given below.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: default
spec:
  replicas: 1
  tls:
    certificates:
    - alias: ca
    - alias: server
  monitor:
    agent: prometheus.io
  authMethods:
  - path: kubernetes
    type: kubernetes
  backend:
    s3:
      bucket: <bucket-name>
      region: <bucket-region>
  unsealer:
    mode:
      awsKmsSsm:
        kmsKeyID: <kms-key-id>
        region: <kms-region>
    secretShares: 1
    secretThreshold: 1
  version: 0.11.5
```

Using the KubeVault `CLI`, we got the decrypted `root-token` from `AWS KMS`, exported it in the environment variable & checked out the secrets stored in `KV` Secret Engine.

```bash
# previous root-token & unseal-keys naming convention

vault-root-token: <root-token>
vault-unseal-key-0: <unseal-key-0>

# get the root-token using KubeVault CLI

$ kubectl vault get-root-token vaultserver vault

# Commands to interact with Vault

$ kubectl exec -it vault-0 vault -- /bin/sh
$ export VAULT_TOKEN=<token>
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_SKIP_VERIFY=true

# Check out the secrets to ensure that they remain intact after the upgrade procedure

$ vault secrets list
$ vault kv get <secret-path>

```

## Uninstall the Older Operator

```bash
# uninstall kubevault older operator

$ helm uninstall kubevault --namespace kubevault

# in case any of these resources remain after helm uninstall, use kubectl delete <resource> -l app.kubernetes.io/name=kubevault-operator

$ kubectl delete mutatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-operator
$ kubectl delete validatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-operator
$ kubectl delete apiservices -l app.kubernetes.io/name=kubevault-operator

```

## Install the Latest Operator & CLI

```bash
# Install KubeVault CLI version v0.10.0

$ curl -o kubectl-vault.tar.gz -fsSL https://github.com/kubevault/cli/releases/download/v0.10.0/kubectl-vault-linux-amd64.tar.gz \
  && tar zxvf kubectl-vault.tar.gz \
  && chmod +x kubectl-vault-linux-amd64 \
  && sudo mv kubectl-vault-linux-amd64 /usr/local/bin/kubectl-vault \
  && rm kubectl-vault.tar.gz LICENSE.md

# Install the KubeVault Operator version v2022.09.22 

$ helm install kubevault appscode/kubevault \
    --version v2022.09.22 \
    --namespace kubevault --create-namespace \
    --set-file global.license=/path/to/vault/license.txt
```

Now, if the Vault `Pod` is restarted, it'd expect the `root-token` & `unseal-keys` in the new naming convention which is: `k8s.<kubevault.com or cluster uid>.namespace.name`.
So, before doing that we must update the root token & unseal keys in the new naming convention. We used the KubeVault CLI sync command to do that.

```bash
# Sync the root-token & unseal-keys 
# check the kms & ensure that root-token & unseal-keys are updated according to new naming convention
# old root-token & unseal-keys are also there for safety purpose

$ kubectl vault root-token sync vaultserver vault
$ kubectl vault unseal-key sync vaultserver vault
```

## Update the TLS Certificates

We decided to go with `cert-manager` managed certificates this time. So, we first installed Cert-manager in our client cluster.
Using the existing `CA` certs we created an `Issuer` & updated the `VaultServer` spec along with proper duration. To install Cert-manager visit [here](https://cert-manager.io/docs/installation/helm/).

```yaml
# Vault Issuer created using the existing vault-ca-certs

apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: vault-issuer
 namespace: default
spec:
 ca:
   secretName: vault-ca-certs

```

```yaml
# VaultServer spec with Issuer Ref & updated certificates duration

tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: vault-issuer
    certificates:
    - alias: ca
      secretName: vault-ca-certs
    - alias: server
      duration: "8952h"
      secretName: vault-server-certs
    - alias: client
      duration: "8952h"
      secretName: vault-client-certs

```

Now, all that was required was to run a Vault `Ops-request` to `rotate` the VaultServer `TLS`. Once, the `Ops-request` was successful
Vault Pod came up with updated TLS certificates. Rotate TLS `Ops-request` looks similar to the one given below:

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: renew-tls-2022
  namespace: default
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: vault-issuer
    rotateCertificates: true

```

Once the upgrade was done, we went through the similar approach described above to verify the secrets in SecretEngine.

## What Next?

Please try the latest release and give us your valuable feedback. If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
