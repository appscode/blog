---
title: JWT/OIDC Auth method & Automation with KubeVault CLI
date: 2022-07-20
weight: 20
authors:
  - Sakib Alamin
tags:
  - kubevault
  - jwt
  - oidc
  - authentication
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

## Summary

AppsCode held a webinar on **"JWT/OIDC Auth method & Automation with KubeVault CLI"**. This took place on 20th July 2022. The contents of what took place at the webinar are given below:

- `VaultServer` deployment with `JWT/OIDC` authentication method 
- Discussion on `JWT/OIDC` authentication method 
- Automation using `KubeVault CLI`

## Description of the Webinar

This webinar was focused on the `JWT/OIDC` authentication method with `KubeVault`, the ins and outs of the `KubeVault CLI`. 
Initially the `Vault` was deployed with `JWT/OIDC` authentication method enabled & `OIDC provider` configurations were discussed. It was also shown, how `VaultRole` can be created to `Login` to `Vault UI` using this `JWT/OIDC` authentication method.

Later on the webinar, various features of `KubeVault CLI` were discussed, e.g: CRUD operation on Vault `root-token` & `unseal-keys`, `Approve/Deny/Revoke` `SecretAccessRequest` for managing DB user privileges, `generate` `SecretProviderClass` for the secrets-store CSI driver, etc. Overall, it was shown how `KubeVault CLI` can be very handy while using `KubeVault` & its CRDs, also automating tedious tasks in general. 

> A step-by-step guide & the manifest files used during the webinar can be found [here](https://github.com/kubevault/demo). 

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/2bm5D8phdJQ" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
