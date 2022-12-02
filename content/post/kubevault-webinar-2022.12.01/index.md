---
title: Backup & Restore Vault cluster with Stash
date: "2022-12-01"
weight: 20
authors:
- Sakib Alamin
tags:
- backup
- hashicorp
- kubernetes
- kubevault
- restore
- secret-management
- security
- stash
- vault
---

## Summary

AppsCode held a webinar on **"Backup & Restore Vault cluster with Stash"**. This took place on 1st December 2022. The contents of what took place at the webinar are given below:

- Install `KubeVault`
- Install `Stash`
- Deploy Vault using `KubeVault`
- Backup Vault data with `Stash`
- Restore Vault data with `Stash`

## Description of the Webinar

This webinar was focused on the Vault `Backup` & `Restore` solution using `Stash` based on the `KubeVault` release `2022.11.30`.

`KubeVault` supports a number of different `Storage Backend` types. One of the most recommended `Storage Backend` types is the Integrated Storage `Raft`. `Stash` add-on for Vault lets you take Backup snapshot & Restore it whenever required for a Vault cluster backed with Raft.

In this webinar, initially we talked about various `CRDs` associated with `KubeVault` & `Stash` e.g: `VaultServer`, `Function`, `Task`, `BackupConfiguration`, `RestoreSession`, etc. We also discussed the workflow of the backup & restoration process in different scenarios. Later, we demonstrated the full flow of the `Backup` & `Restore` process using `Stash`.


> A step-by-step guide to re-create the process & the manifest files used during the webinar can be found [here](https://github.com/kubevault/demo/tree/master/webinar-12-01-22). 

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/TOxufXiyVok" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.11.30/setup/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
