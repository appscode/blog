---
title: Introducing KubeVault v2024.3.12
date: "2024-03-15"
weight: 25
authors:
- Abdullah Al Shaad
tags:
- cli
- hashicorp
- kubernetes
- kubevault
- kubevault-cli
- pki
- secret-management
- security
- vault
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines' management, policy management in the Kubernetes native way.

We are very excited to announce the release of [KubeVault v2024.3.12](https://kubevault.com/docs/v2024.3.12/setup/) Edition. In this release, the `PKI` `SecretEngine` has been added.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2024.3.12/README.md).

### PKI SecretEngine

The PKI secrets engine generates dynamic X.509 certificates. With this secret engine, services can get certificates without going through the usual manual process of generating a private key and CSR, submitting to a CA, and waiting for a verification and signing process to complete. Vault's built-in authentication and authorization mechanisms provide the verification functionality.
First, we need to enable a `SecretEngine` for `PKI` which works as a CA, and then we can create different roles with different set of rules using `PKIRole`. Then Vault will generate credentials for the role when user request for credentials.

### Root CA

Lets create a RootCA using `PKI` Secret Engine. The CA is is self signed.
```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: pki-root
  namespace: demo
spec:
  vaultRef:
    name: vault
    namespace: demo
  pki:
    isRootCA: true
    commonName: "kubevault.com"
    ttl: "87600h"
  maxLeaseTTL: "48000h"
```

### Intermediate CA

Now, we are going to create an intermediate CA. It is signed by the root CA we have created before.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: pki-inter
  namespace: demo
spec:
  vaultRef:
    name: vault
    namespace: demo
  pki:
    isRootCA: false
    parentCARef:
      name: pki-root
      namespace: demo
    commonName: "intermediate.kubevault.com"
    ttl: "87600h"
```

### PKI Role

Lets create a `PKIRole` which will define rules for certificate creation.
```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: PKIRole
metadata:
  name: pki-role
  namespace: demo
spec:
  secretEngineRef:
    name: pki-root
  allowedDomains:
  - "kubevault.com"
  - "kubedb.com"
  allowSubdomains: true
  maxTTL: "720h"
```

### Certificates

Now we can request certificates using `SecretAccessRequest`
```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: pki-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: PKIRole
    name: pki-role
    namespace: demo
  pki:
    commonName: "test.kubevault.com"
    ttl: "24h"
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Upon approval, the certificate will be available through a Kubernetes Secret in `demo` namespace.

### What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2024.3.12/setup).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
