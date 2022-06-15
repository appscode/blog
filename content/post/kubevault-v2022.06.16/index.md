---
title: Introducing KubeVault v2022.06.16
date: 2022-06-16
weight: 25
authors:
  - Sakib Alamin
tags:
  - kubevault
  - CLI
  - kubevault CLI
  - kubernetes
  - secret-management
  - security
  - vault
  - hashicorp
  - enterprise
  - community
---

We are very excited to announce the release of KubeVault v2022.06.16 Edition. The KubeVault `v2022.06.16` contains VaultServer latest api version `v1alpha2`, update to authentication method with addition of `JWT/OIDC` auth method. A new `SecretEngine` for `MariaDB` has been added, `KubeVault CLI` has been updated along with various fixes on KubeVault resource sync. We're going to discuss some of them in details below.

- [Install KubeVault](https://kubevault.com/docs/v2022.06.16/setup/)

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.06.16/README.md).

## What's new in this release?
- **JWT/OIDC Authentication Method**

    The `JWT` auth method can be used to authenticate with Vault using `OIDC` or by providing a `JWT`. The `OIDC` method allows authentication via a configured `OIDC Provider` using the user's web browser. This method may be initiated from the `Vault UI` or the command line. Alternatively, a `JWT` can be provided directly.

  While deploying the `VaultServer` it's possible to define the list of auth methods users want to enable with it.

    A `VaultServer` yaml may look like this:

  ```yaml
    apiVersion: kubevault.com/v1alpha2
    kind: VaultServer
    metadata:
      name: vault
      namespace: demo
    spec:
      version: 1.10.3
      replicas: 2
      allowedSecretEngines:
        namespaces:
          from: All
        secretEngines:
          - gcp
      authMethods:
        - type: jwt
          path: jwt
          jwtConfig:
            defaultLeaseTTL: 1h
            defaultRole: k8s.kubevault.com.demo.reader-writer-role
            oidcClientID: aFSrk3w06WsQqyjA30HvhbbJIR1VBidU
            oidcDiscoveryURL: https://dev-tob49v6v.us.auth0.com/
            credentialSecretRef:
              name: jwt-cred
      backend:
        raft:
          storage:
            storageClassName: "standard"
            resources:
              requests:
                storage: 1Gi
      unsealer:
        secretShares: 3
        secretThreshold: 2
        mode:
          kubernetesSecret:
            secretName: vault-keys
      monitor:
        agent: prometheus.io
        prometheus:
          exporter:
            resources: {}
      terminationPolicy: WipeOut
      
  ```

  * `.spec.authMethods.type` is a required field, the type of authentication method we want to enable.
  * `.spec.authMethods.path` is a required field, the path where we want to enable this authentication method.
  * `.spec.authMethods.jwtConfig / .spec.authMethods.oidcConfig` contains various configuration for this authentication method. Some of the `parameters` are: `defaultLeaseTTL`, `maxLeaseTTL`, `pluginName`, `credentialSecretRef`, `tlsSecretRef`, `oidcDiscoveryURL`, `oidcClientID`, `oidcResponseMode`, `defaultRole`, `providerConfig`, etc. Check out [this](https://kubevault.com/docs/v2022.06.16/concepts/vault-server-crds/auth-methods/jwt-oidc/) for more details.
    After an authentication method is successfully enabled, `KubeVault` operator will configure it with the provided configuration.

    After successfully enabling & configuring authentication methods, a VaultServer `.status.authMethodStatus` may look like this:
    ```yaml
    status:
      authMethodStatus:
      - path: jwt
        status: EnableSucceeded
        type: jwt
      - path: kubernetes
        status: EnableSucceeded
        type: kubernetes
    
    ```

    We can verify it using the `Vault CLI`:

    ```bash
    $ vault auth list
    
    Path           Type          Accessor                    Description
    ----           ----          --------                    -----------
    jwt/           jwt           auth_jwt_ba23cc30           n/a
    kubernetes/    kubernetes    auth_kubernetes_40fd86fd    n/a
    token/         token         auth_token_950c8b80         token based credentials
    ```
- **MariaDB SecretEngine**

  Now, `MariaDB` SecretEngine can be enabled, configured & `MariaDBRole` can also be created with `KubeVault`.
  Here's a sample yaml for MariDB `SecretEngine` & `MariaDBRole`:

  ```yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: SecretEngine
  metadata:
    name: mariadb-engine
    namespace: demo
  spec:
    vaultRef:
      name: vault
      namespace: demo
    mariadb:
      databaseRef:
        name: mariadb
        namespace: db
      pluginName: "mysql-database-plugin"
  
  ```
  ```yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: MariaDBRole
  metadata:
    name: mariadb-role
    namespace: dev
  spec:
    secretEngineRef:
      name: mariadb-engine
    creationStatements:
      - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
      - "GRANT CREATE, SELECT, INSERT ON *.* TO '{{name}}'@'%';"
    revocationStatements:
      - "DROP USER '{{name}}'@'%';" 
    defaultTTL: 3h
    maxTTL: 24h
  ```
- **Merge secrets using KubeVault CLI**
  
  Now, you can merge two `Kubernetes` `Secret`s using `KubeVault CLI`.
  ```bash
  # merge two secret name1 & name2 from ns1 & ns2 namespaces respectively
  $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2>
  
  # --overwrite-keys flag will overwrite keys in destination if set to true.
  $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2> --overwrite-keys=true
  
  $ kubectl vault merge-secrets --src=demo/src-secret --dst=demo/dest-cred
  ```
  
## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
