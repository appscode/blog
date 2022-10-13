---
title: KubeVault OpsRequest - Day 2 Life-cycle Management for Vault on Kubernetes
date: 2022-09-01
weight: 20
authors:
  - Sakib Alamin
tags:
  - kubevault
  - ops-request
  - tls
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

AppsCode held a webinar on **"KubeVault OpsRequest - Day 2 Life-cycle Management for Vault on Kubernetes"**. This took place on 1st September 2022. The contents of what took place at the webinar are given below:

- `VaultServer` deployment 
- Discussion on `Pod Disruption Budget` for Vault Cluster
- Discussion on updated `Vault Health Checker`
- KubeVault `Ops-requests`
- `Generate` & `Rotate` Vault root-token with `KubeVault CLI`

## Description of the Webinar

This webinar was focused on the Vault `Ops-requests`, `PDB` support for Vault Cluster & updated Vault `Health Checker`. Initially, there was a discussion on inclusion of `PDB` for KubeVault & the updated user configurable Vault `Health Checker`. 

`VaultServer` was deployed to show the creation of `PDB` with it, as well as various `Vault Ops-requests` were made to reconfigure the Vault `TLS`.

Later on the webinar, some upcoming features like, `Generate Vault root-token`, `Rotate Vault root-token` were also demonstrated using the `KubeVault CLI`.

> A step-by-step guide & the manifest files used during the webinar can be found [here](https://github.com/kubevault/demo). 

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/A0n80pnwTpY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
