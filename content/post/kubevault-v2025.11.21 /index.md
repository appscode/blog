---
title: Introducing KubeVault v2025.11.21
date: "2025-11-21"
weight: 25
authors:
- Rudro Debnath
tags:
- cli
- hashicorp
- kubernetes
- kubevault
- kubevault-cli
- openbao
- secret-management
- vault
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines' management, policy management in the Kubernetes native way.

We are very excited to announce the release of [KubeVault v2025.11.21](https://kubevault.com/docs/v2025.11.21/setup/) Edition. 

You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2025.11.21/README.md).

## New Version Support

KubeVault now supports the latest OpenBao version 2.4.3. To deploy a VaultServer with the latest release, apply the following manifest.

````yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 3
  version: openbao-2.4.3
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

After deployment:

- You can exec into the pod and use either the `vault` or `bao` CLI.
- All secret engines, auth methods, policies, tokens, and KubeVault workflows work as expected.

## Bug Fixes and Performance Improvements
We have fixed some minor bugs and improved performance in this release.


## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2025.11.21/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
