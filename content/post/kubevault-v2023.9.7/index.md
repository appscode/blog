---
title: Introducing KubeVault v2023.9.7
date: 2023.09.07
weight: 25
authors:
- Abdullah Al Shaad
tags:
- hashicorp
- kubernetes
- kubevault
- secret-management
- security
- vault
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management,
encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes.
It also supports various secret engines' management, policy management in the Kubernetes native way.

We are very excited to announce the release of [KubeVault v2023.9.7](https://kubevault.com/docs/v2023.9.7/setup/) Edition.

You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2023.9.7/README.md).


## New Version Support

KubeVault now supports the latest vault version 1.13.3. To deploy a VaultServer with the latest release, apply the following manifest.

````yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 3
  version: 1.13.3
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
````

## Cross Namespace SecretAccessRequest
KubeVault now supports creating SecretAccessRequest from different namespace than SecretEngine and Role objects.

## Bug Fixes and Performance Improvements
We have fixed some minor bugs and improved performance in this release.

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2023.03.03/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
