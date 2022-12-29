---
title: Introducing KubeVault v2022.12.28
date: "2022-12-28"
weight: 25
authors:
- Abdullah Al Shaad
tags:
- cli
- hashicorp
- kubernetes
- kubevault
- kubevault-cli
- redis
- secret-management
- security
- vault
---

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

We are very excited to announce the release of [KubeVault v2022.12.28](https://kubevault.com/docs/v2022.12.28/setup/) Edition. In this release, the `SecretEngine` for `Redis` has been added, `KubeVault CLI` has been updated for generating `SecretProviderClass` for `Redis`.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.12.28/README.md).

- **Redis SecretEngine**

Redis Secret can be used to generate dynamic credentials for Redis Standalone database using Vault. First, we need to enable a `SecretEngine` for  `Redis`
and then we can create different roles with different set of permissions using `RedisRole`. Then Vault will generate credentials for the role when user request for credentials.
We can also mount the secret in a pod using `SecretProviderClass`

  Now, `Redis` SecretEngine can be enabled, configured & `RedisRole` can also be created with `KubeVault`.
  Here's a sample yaml for Redis `SecretEngine` & `RedisRole`:

  ```yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: SecretEngine
  metadata:
    name: redis-secret-engine
    namespace: demo
  spec:
    vaultRef:
      name: vault
      namespace: demo
    redis:
      databaseRef:
        name: redis
        namespace: db
      pluginName: "redis-database-plugin"
  ```


  ```yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: RedisRole
  metadata:
    name: write-read-role
    namespace: demo
  spec:
    secretEngineRef:
      name: redis-secret-engine
    creationStatements:
      - '["~*", "+@read","+@write"]'
    defaultTTL: 1h
    maxTTL: 24h
  ```
We can bind the role with a service account and mount the generated credentials using `SecretProviderClass`. After creating a `SecretRoleBinding` 
the following commands generates `SecretProviderClass` YAML for `RedisRole`
```bash
$ kubectl vault generate secretproviderclass vault-db-provider -n demo \
       --secretrolebinding=demo/secret-role-binding \
       --vaultrole=RedisRole/write-read-role \
       --keys username=redis-user --keys password=redis-pass -o yaml
```
 To learn more about how to mount Redis credentials in pod, head over [here](https://kubevault.com/docs/v2022.12.28/guides/secret-engines/redis/csi-driver/)

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.12.28/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
