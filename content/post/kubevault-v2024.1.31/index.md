---
title: Introducing KubeVault v2024.1.31
date: "2024-02-05"
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

We are pleased to announce the release of KubeVault `v2024.1.31`.
This release contains a few bug fixes in KubeVault `SecretEngine`s for GCP, AWS and Azure.

Also, we have updated Kubernetes client library dependencies to `1.29`.
The detailed commit by commit changelog can be found [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2024.1.31/README.md).


## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2024.1.31/setup/install/kubevault/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
