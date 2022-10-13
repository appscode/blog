---
title: Introducing KubeVault v2022.01.11
date: 2022-01-11
weight: 25
authors:
- Sakib Alamin
tags:
- kubevault
- cert-manager
- cli
- kubevault cli
- kubernetes
- secret-management
- security
- vault
- hashicorp
- enterprise
- community
---

We are very excited to announce the release of KubeVault v2022.01.11 Edition. The KubeVault `v2022.01.11` contains major improvements of the `KubeVault CLI` for better user experiences, `cert-manager` integration for managing `TLS`, clean-up of the `unseal-keys` and `root-token` along with the `VaultServer` when the `TerminationPolicy` is set to `WipeOut`, newly added `Expired` phase for  `SecretAccessRequest` based on `TTL` or admin side revocation, as well as few bug fixes. We're going to demonstrate some of these improvements below.

- [Install KubeVault](https://kubevault.com/docs/v2022.01.11/setup/)

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.01.11/README.md).

## What's new in this release?

- **Improved KubeVault CLI**
  i. `KubeVault CLI` now can be used to generate [SecretProviderClass](https://secrets-store-csi-driver.sigs.k8s.io/concepts.html#secretproviderclass) to ease the process of injecting `vault` secrets into `K8s` resources.

  ```bash
  # Generate secretproviderclass with name <name1> and namespace <ns1>
  # secretrolebinding with namespace <ns2> and name <name2>
  # vaultrole kind MongoDBRole and name <name3>
  # keys to mount <secretKey> and it's mapping name <objectName> 

  $ kubectl vault generate secretproviderclass <name1> -n <ns1> \
    --secretrolebinding=<ns2>/<name2> \
    --vaultrole=MongoDBRole/<name3> \
    --keys <secretKey>=<objectName> -o yaml 
    
  # command to generate secretproviderclass for the MongoDB username and password

  $ kubectl vault generate secretproviderclass mongo-secret-provider -n test \
    --secretrolebinding=dev/secret-r-binding \
    --vaultrole=MongoDBRole/mongo-role \
    --keys username=mongo-user --keys password=mongo-pass -o yaml

  ```

  ii. A new phase `Expired` has been added for `SecretAccessRequest` to show to expiration status by `TTL` or admin side revocation. To enable that, `revoke` command has been added so that the admin user can revoke a `SecretAccessRequest` using the `KubeVault CLI`. 

  ```bash
  # command to revoke SecretAccessRequest
  $ kubectl vault revoke secretaccessrequest <name> -n <ns>
    
  ```

  iii. You can now get the `decrypted` value of `vault-root-token` simply using the `get-root-token` command by `KubeVault CLI` instead of going through the tedious process of manually retrieving and decrypting from the major cloud providers storages e.g, `GCS`, `AWS`, `Azure` or even from `K8s Secret`.
  
  ```bash
  # command to get vault root-token 
  $ kubectl vault get-root-token vaultserver -n <ns> <name>

  # command to get vault root-token for vaultserver with name vault in demo namespace
  $ kubectl vault get-root-token vaultserver -n demo vault 
  k8s.kubevault.com.demo.vault-root-token: s.9WeTJ2YgcM1fLD44tKPK9rdO

  # you can pass flag --value-only to get only the value of the root-token
  $ kubectl vault get-root-token vaultserver -n demo vault --value-only
  s.9WeTJ2YgcM1fLD44tKPK9rdO
  
  ```
  

- **Cert-manager managed TLS**
  Now, `VaultServer` TLS can be managed with `cert-manager` along with existing `self-signed` certificate generation process. 
  Sample yamls shows how `VaultServer.spec` may look like with it's `TLS` enabled using `cert-manager`: 
  

  ```yaml
    apiVersion: kubevault.com/v1alpha1
    kind: VaultServer
    metadata:
      name: vault
      namespace: demo
    spec:
      tls:
        issuerRef:
          apiGroup: "cert-manager.io"
          kind: Issuer
          name: vault-issuer
      ---
  ```

- **Clean-up vault root-token, unseal-keys**
  `vault-root-token` and `vault-unseal-keys` will now be cleaned-up if the `TerminationPolicy` of the `VaultServer` is set to `WipeOut`.


- **Multi-cluster usage**
  `vault-unseal-keys` conflict issue for multi-cluster usage has been fixed in this release. You can provide `--clusterName` flag during `helm install` to set the cluster-name. Default value will be the `cluster UID` in case cluster-name is not provided.


- **GCS ServiceAccount token cleanup**
  `GCS Service Account` token cleanup issue on delete & expiration has been fixed in this release. 


## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.01.11/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
