---
title: Introducing KubeVault v2022.02.22
date: "2022-02-22"
weight: 25
authors:
- Sakib Alamin
tags:
- cli
- hashicorp
- kubernetes
- kubevault
- kubevault-cli
- secret-management
- security
- vault
---

We are very excited to announce the release of KubeVault v2022.02.22 Edition. The KubeVault `v2022.02.22` contains major improvements of the `KubeVault CLI` for better user experiences. Now, using `KubeVault CLI` you can `get`, `set`, `delete`, `list` and `sync` vault `unseal-keys` and `root-token`.

- [Install KubeVault](https://kubevault.com/docs/v2022.02.22/setup/)

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.02.22/README.md).

## What's new in this release?

- **Improved KubeVault CLI**

  i. You can now get, set, delete, sync and list the value of `vault-root-token` simply using `KubeVault CLI` instead of going through the tedious process of manually retrieving and decrypting from the major cloud providers storages e.g, `GCS`, `AWS`, `Azure` or even from `K8s Secret`.

  ```bash
  # GET root-token
  # get the decrypted root-token of a vaultserver with name vault in demo namespace
  $ kubectl vault root-token get vaultserver vault -n demo
  
  # pass the --value-only flag to get only the decrypted value
  $ kubectl vault root-token get vaultserver vault -n demo --value-only
  
  # pass the --token-name flag to get only the decrypted root-token value with a specific token name
  $ kubectl vault root-token get vaultserver vault -n demo --token-name <token-name> --value-only 

  ```

  ```bash
  # SET root-token
  # set the root-token with name --token-name flag & value --token-value flag
  $ kubectl vault root-token set vaultserver vault -n demo --token-name <name> --token-value <value>

  # default name for root-token will be used if --token-name flag is not provided
  # default root-token naming format: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-root-token
  $ kubectl vault root-token set vaultserver vault -n demo --token-value <value>

  ```

  ```bash
  # DELETE root-token
  # delete the root-token with name set by --token-name flag
  $ kubectl vault root-token delete vaultserver vault -n demo --token-name <name>
  
  # default name for root-token will be used if --token-name flag is not provided
  # default root-token naming format: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-root-token
  $ kubectl vault root-token delete vaultserver vault -n demo 
  
  ```

  ii. You can also get, delete, set & list the value of `vault-unseal-key` simply using `KubeVault CLI` instead of going through the tedious process of manually retrieving and decrypting from the major cloud providers storages e.g, `GCS`, `AWS`, `Azure` or even from `K8s Secret`.

  ```bash
  # GET unseal-key
  # get the decrypted unseal-key of a vaultserver with name vault in demo namespace with --key-id flag
  # default unseal-key format: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-unseal-key-{id}
  $ kubectl vault unseal-key get vaultserver vault -n demo --key-id <id>
  
  # pass the --key-name flag to get only the decrypted unseal-key value with a specific key name
  $ kubectl vault unseal-key get vaultserver vault -n demo --key-name <name>  
  
  ```

  ```bash
  # SET unseal-key
  # set the unseal-key with name --key-name flag & value --key-value flag
  $ kubectl vault unseal-key set vaultserver vault -n demo --key-name <name> --key-value <value>
  
  # pass the --key-id flag to set the default unseal-key with given <id>
  $ kubectl vault unseal-key set vaultserver vault -n demo --key-id <id> --key-value <value>
  
  # default name for unseal-key will be used if --key-name flag is not provided
  # default unseal-key naming format: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-unseal-key-{id}
  $ kubectl vault unseal-key set vaultserver vault -n demo --key-id <id> --key-value <value>
  
  ```

  ```bash
  # DELETE unseal-key
  # delete the unseal-key with name set by --key-name flag
  $ kubectl vault unseal-key delete vaultserver vault -n demo --key-name <name>
  
  # delete the unseal-key with name set by --key-id flag
  $ kubectl vault unseal-key delete vaultserver vault -n demo --key-id <id>
  
  ```

  ```bash
  # LIST unseal-key
  # list the vault unseal-keys
  $ kubectl vault unseal-key list vaultserver vault -n demo
  
  ```

  iii. You can use the **sync** command to update the naming format of your vaultserver `root-token` & `unseal-keys`.

  ```bash
  # SYNC
  # sync the vaultserver root-token & unseal-keys
  # old naming conventions: vault-root-token
  # new naming convention for root-token: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-root-token
  # example: kubectl vault root-token sync vaultserver <vault-name> -n <vault-namespace>
  $ kubectl vault root-token sync vaultserver vault -n demo
  
  # old naming conventions: vault-unseal-key-0, vault-unseal-key-1, etc.
  # new naming convention for unseal-key: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-unseal-key-{id}
  # example: kubectl vault unseal-key sync vaultserver <vault-name> -n <vault-namespace>
  $ kubectl vault unseal-key sync vaultserver vault -n demo

  ```

> Note: It's suggested that you use the `sync` command to update your vault `root-token` & `unseal-key` naming formats.

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.02.22/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
