---
title: Secure Secrets - A Cloud-Native Approach made simple with KubeVault
date: 2022-01-13
weight: 20
authors:
  - Sakib Alamin
tags:
  - kubevault
  - cert-manager
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

## Summary

AppsCode held a webinar on **"Secure Secrets: A Cloud-Native Approach made simple with KubeVault"**. This took place on 12th Jan 2022. The contents of what took place at the webinar are shown below:
- Deploy TLS Secured VaultServer 
- Enable SecretEngine
- Create Database Roles
- Manage User Privileges
- KubeVault CLI in Action
- Q & A Session

## Description of the Webinar

It is required to install the followings to get started:
  - KubeDB Enterprise Operator
  - KubeVault Enterprise Operator
  - Secrets Store CSI Driver
  - Vault Specific CSI Provider

The speaker starts by deploying TLS secured `VaultServer` (TLS managed by `cert-manager`) & `MySQL` Database by `KubeDB`. Speaker shows how easy it is to get the decrypted `vault-root-token` from GCS bucket using KubeVault CLI. Followed by, enabling `SecretEngine` & creating some Database `Roles`. 

After that, it's shown how to manage user privileges using two different ways. Firstly, using the `SecretAccessRequest`, which is more human interaction friendly, that can be `Approved` or `Denied` using the KubeVault CLI. Secondly, using the `SecretRoleBinding` which is a more machine friendly way, that binds some roles to a `K8s ServiceAccount`.

Then, it's demonstrated how microservices can be made more secure using the `Dynamic Secrets` by Vault, where a microservice is deployed that reads mounted credentials, logs into the DB and makes queries. DB secrets are mounted on directories with the help of Secrets store CSI Driver & Vault CSI Provider. 

Lastly, it's shown how `KubeVault CLI` can `Revoke` a user privileges by using a simple command.

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/dLW4ZX3vcJI" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.01.11/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
