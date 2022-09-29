---
title: KubeVault v1alpha1 to v1alpha2 Migration Procedure
date: 2022-09-30
weight: 20
authors:
  - Sakib Alamin
tags:
  - KubeVault
  - ops-request
  - v1alpha1
  - v1alpha2
  - TLS
  - cert-manager
  - CLI
  - KubeVault CLI
  - kubernetes
  - secret-management
  - security
  - vault
  - hashicorp
  - enterprise
  - community
---

# KubeVault Operator upgrade procedure for Officebrands

# Current setup of officebrands

```bash
# CLI version they're using
$ curl -o kubectl-vault.tar.gz -fsSL https://github.com/kubevault/cli/releases/download/v0.6.0-alpha.0/kubectl-vault-linux-amd64.tar.gz \
  && tar zxvf kubectl-vault.tar.gz \
  && chmod +x kubectl-vault-linux-amd64 \
  && sudo mv kubectl-vault-linux-amd64 /usr/local/bin/kubectl-vault \
  && rm kubectl-vault.tar.gz LICENSE.md


# The operator version they're using: v2021.10.11
# validating webhook is false as they're using 1 unseal key
# We've updated the validating webhook, so now unseal key can be set to 1 (prev. min value was 2)
$ helm install kubevault appscode/kubevault \
    --version v2021.10.11 \
    --namespace kubevault --create-namespace \
    --set kubevault-operator.apiserver.enableValidatingWebhook=false \
    --set-file global.license=/home/sakib/Documents/webinar/deploy/officebrands-upgrade/vault-license.txt

# Current root-token & unseal-keys naming convention
vault-root-token: M01HVTkyMFFWVW4weFE5WUNMZ3drYjh3
vault-unseal-key-0: OTY5YmU2ZmQyMWVkYTg2NDI5MTk5ZWUyZjM3NTZlMzJlZTRlZmMzNjViN2RhOThhYmIxYjY4NWRmY2UxMDYyNw==

# Get the root-token using the old CLI (no command for unseal-key)
$ kubectl vault get-root-token vaultserver vault

```

# Check their KV values before doing anything (we must ensure those remain intact)

```bash
$ kubectl exec -it -n demo vault-0 vault -- /bin/sh
$ export VAULT_TOKEN=<token>
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_SKIP_VERIFY=true

# lists all the secret engines
$ vault secrets list

$ vault kv get <path>
```

# Uninstall the older version & Install the latest Operator & CLI

```bash
# uninstall kubevault (is any cleanup issue, delete them manually using above commands)
$ helm uninstall kubevault --namespace kubevault

# for older operator use kubectl delete <resource> -l app.kubernetes.io/name=kubevault-operator
$ kubectl delete mutatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-operator
$ kubectl delete validatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-operator
$ kubectl delete apiservices -l app.kubernetes.io/name=kubevault-operator

# for latest operator use kubectl delete <resource> -l app.kubernetes.io/name=kubevault-webhook-server
$ kubectl delete mutatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-webhook-server
$ kubectl delete validatingwebhookconfiguration -l app.kubernetes.io/name=kubevault-webhook-server
$ kubectl delete apiservices -l app.kubernetes.io/name=kubevault-webhook-server

# Latest KubeVault CLI install
$ curl -o kubectl-vault.tar.gz -fsSL https://github.com/kubevault/cli/releases/download/v0.10.0/kubectl-vault-linux-amd64.tar.gz \
  && tar zxvf kubectl-vault.tar.gz \
  && chmod +x kubectl-vault-linux-amd64 \
  && sudo mv kubectl-vault-linux-amd64 /usr/local/bin/kubectl-vault \
  && rm kubectl-vault.tar.gz LICENSE.md

# Install the latest Vault Operator(no problem with the ValidatingWebhook)
$ helm install kubevault appscode/kubevault \
    --version v2022.09.22 \
    --namespace kubevault --create-namespace \
    --set-file global.license=/home/sakib/Documents/webinar/deploy/officebrands-upgrade/vault-license.txt


# Sync the root-token & unseal-keys (note that: ignore namespace for officebrands or use default)
# check the kms & ensure that root-token & unseal-keys are updated according to new nameing convention
# old root-token & unseal-keys are also there for safety purpose
$ kubectl vault root-token sync vaultserver -n demo vault
$ kubectl vault unseal-key sync vaultserver -n demo vault

# **Now we update the certs & restarts the VaultServer** 

# Check out the unsealer log to ensure that correct root-token & unseal-keys naming convention is being used
$ kubectl logs -f -n demo demo vault-0 vault-unsealer

# Once confirmed that the new naming convention is used, we can safely delete the old root-token & unseal-keys

```

## Useful commands

```bash
# Get the cluster UID
$ kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'

# delete all the crds (note that: this is valid for old operator) & we don't need this!
$ kubectl delete crds -l app.kubernetes.io/name=kubevault

```

# Install Cert-manager

```bash
$ helm repo add jetstack https://charts.jetstack.io

$ helm repo update

$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.crds.yaml

$ helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.9.1

$ kubectl get pods --n cert-manager
```

# Yaml files

## rotate tls yaml

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: new-vault-issuer
    rotateCertificates: true
```

## Issuer yaml

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: new-vault-issuer
 namespace: demo
spec:
 ca:
   secretName: vault-ca-certs
```

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
