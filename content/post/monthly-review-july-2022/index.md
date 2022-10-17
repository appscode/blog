---
title: Monthly Review - July, 2022
date: "2022-08-01"
weight: 14
authors:
- Dipta Roy
tags:
- appscode
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- kubernetes-database
- kubernetes-native
- stash
---

# Releases


#### Introducing Stash v2022.07.09

We are announcing Stash v2022.07.09 which introduces a new feature and a few improvements. We are going to highlight the major changes in this post.

Here is the new feature of this release: 

Add Hook ExecutionPolicy: We have added executionPolicy to our PostBackupHook and PostRestoreHook. You can specify any one of these three executionPolicy - ‘Always’, ‘OnFailure’, ‘OnSuccess’ along with the Hooks. The default executionPolicy is ‘Always’. If you use ‘OnSuccess’ executionPolicy, the Hooks will be executed only after a successful backup or restore. On the other hand, ‘onFailure’ will execute the Hooks after failed backup or restore only. 

We have also made some improvements. Here are the notable changes, Ensured that the backupsession will be marked as completed after executing all the steps #176, Added RBAC permissions for finalizer #1458, Custom labels will be passed to Volume Snapshotter Job properly #1461.

Link: https://blog.byte.builders/post/stash-v2022.07.09/



# Webinars


#### Managing Production Grade Percona XtraDB Cluster in Kubernetes using KubeDB

AppsCode held a webinar on “Managing Production Grade Percona XtraDB Cluster in Kubernetes using KubeDB” on 27th July 2022. The contents discussed on the webinar:

- TLS Secured Percona XtraDB Cluster
- Cluster Failure Recovery
- Monitor Percona XtraDB Metrics
- Custom Configuration
- Live Demonstration
- Q&A Session

Link: https://blog.byte.builders/post/perconaxtradb-webinar-v2022.07.27/

#### JWT/OIDC Auth method & Automation with KubeVault CLI

AppsCode held a webinar on “JWT/OIDC Auth method & Automation with KubeVault CLI”. This took place on 20th July 2022. The contents of what took place at the webinar are given below:

- VaultServer deployment with JWT/OIDC authentication method
- Discussion on JWT/OIDC authentication method
- Automation using KubeVault CLI

Link: https://blog.byte.builders/post/kubevault-webinar-2022.07.20/

#### Using EKS IRSA and Kube2iam with Stash

On July 6, 2022, Appscode held a webinar on Using EKS IRSA and Kube2iam with Stash. Key contents of the webinar are:

- Setup IRSA
- Backup and Restore a database using IRSA
- Setup Kub2iam
- Backup and Restore a database using Kub2iam
- Q&A Session

Link: https://blog.byte.builders/post/stash-webinar-2022-07-06/



# Blogs Published


#### Run PostgreSQL in Azure Kubernetes Service (AKS) Using KubeDB

KubeDB simplifies Provision, Upgrade, Scaling, Volume Expansion, Monitor, Backup, Restore for various Databases in Kubernetes on any Public & Private Cloud. Here is how to Run & Manage PostgreSQL in Azure Kubernetes Service (AKS) Using KubeDB.

Link: https://blog.byte.builders/post/run-postgresql-in-aks/



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

Follow our Linkedin for more [AppsCode Inc](https://www.linkedin.com/company/appscode/)

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).
