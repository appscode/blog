---
title: Using EKS IRSA and Kube2iam with Stash
date: "2022-07-14"
weight: 15
authors:
- Hossain Mahmud
tags:
- backup
- cloud-native
- eks
- irsa
- kube2iam
- kubernetes
- restore
- stash
---

## Summary

On July 6, 2022, Appscode held a webinar on **Using EKS IRSA and Kube2iam with Stash**. Key contents of the webinar are:

1) Setup IRSA
2) Backup and Restore a database using IRSA
3) Setup Kub2iam
4) Backup and Restore a database using Kub2iam
4) Q & A Session

## Description of the Webinar

Initially, they demonstrated the IAM roles for service accounts (IRSA) setup on EKS and deployed a `MariaDB` instance to show the credential-less backup and recovery using IRSA. Later, they demonstrated the Kub2iam setup on EKS and showed another backup and recovery using Kub2iam. 

Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/sxQynhH45VE" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
