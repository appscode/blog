---
title: Introducing KubeVault v2023.05.05
date: "2023-05-05"
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

We are very excited to announce the release of [KubeVault v2023.05.05](https://kubevault.com/docs/v2023.05.05/setup/) Edition. In this release, we have fixed a bug and did some code improvements

You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2023.05.05/README.md).


## Bug Fixes

Fixed a bug which occurs due to the change in creation of secret based Service Account in Kubernetes version 1.24.
[#108309](https://github.com/kubernetes/kubernetes/pull/108309)


Now the token secret is explicitly created when it is not automatically created by Kubernetes in versions greater than 1.23. [#96](https://github.com/kubevault/operator/pull/96)

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2023.03.03/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
