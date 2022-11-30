---
title: Introducing KubeVault v2022.11.30
date: "2022-11-30"
weight: 25
authors:
- Sakib Alamin
tags:
- backup
- hashicorp
- kubernetes
- kubevault
- restore
- security
- stash
- vault
---

[KubeVault](https://kubevault.com) is a Git-Ops ready, production-grade solution for deploying and configuring [HashiCorp Vault](https://www.vaultproject.io/) on Kubernetes.

We are excited to announce the release of KubeVault `v2022.11.30`. This release takes us one step closer to Vault Backup & Restore process. 

With this release, `KubeVault` is integrated with `Stash`! You can now take `Backup` & `Restore` your `Vault` cluster using `Stash`. 
`Stash` by `AppsCode` is a complete Kubernetes native disaster recovery solution for backup and restore your volumes and databases in Kubernetes on any public and private clouds.
Read more about `Stash` [here](https://stash.run/).

`KubeVault` supports a number of different Storage Backend types. One of the most recommended Storage Backend types is the `Integrated Storage` [Raft](https://developer.hashicorp.com/vault/docs/configuration/storage/raft).
`Stash` add-on for [Vault](https://github.com/stashed/vault) lets you take `Backup` snapshot & `Restore` it whenever required for a Vault cluster backed with `Raft`. 

Vault provides a set of standard operating procedures (SOP) for backing up a Vault cluster. `Stash` add-on for `Vault` protects your Vault cluster against data corruption or sabotage keeping those SOP in mind.

`Stash` will take the snapshot using a consistent mode that forwards the request to the cluster leader, and the leader will verify it is still in power before taking the snapshot. It'll also store the Vault `Unseal keys` & `root token` with the snapshot. Even if you lose our Vault cluster, `Unseal keys` & `root token`, you'll be able to restore it using this snapshot. Thus, we'll have a complete backup & restore solution! 


The detailed commit by commit changelog can be found [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.11.30/README.md).

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.11.30/setup).

> We are going to demo these features in our upcoming webinar on Dec 01, 2022 1:00 PM CEST. Please register [here](https://appscode.com/webinar/) to attend!

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
