---
title: Monthly Review - June, 2022
date: 2022-07-01
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native platform
- kubernetes native data platform
- kubernetes native
- kubernetes
- database
- kubernetes database
- run production-grade database
- run & manage database in k8s
- run & manage database in kubernetes
- manage database in kubernetes
- kubedb
- stash
- appscode
- backup kubernetes statefulset
- running production-grade databases on kubernetes
---

# Releases


#### Introducing Stash v2022.06.27

 We are announcing Stash v2022.06.27 which include a critical bug fix for PostgreSQL and MongoDB addon. This release fixes a bug that was causing the license checker to fail in the backup and restore jobs for PostgreSQL (#1077) and MongoDB (#1593) databases. This bug was introduced in Stash v2022.06.21 release.

 Link: https://blog.byte.builders/post/stash-v2022.06.27/

#### Introducing Stash v2022.06.21

We are very excited to announce Stash v2022.06.21. This release adds support for Kubernetes v1.24.x. We have also introduces a few new features and improvements. We have squashed a few bugs as well.

Link: https://blog.byte.builders/post/stash-v2022.06.21/

#### Announcing Voyager v2022.06.20

We are pleased to announce the release of Voyager v2022.06.20. In this release, we have released operator and HAProxy images to fix CVE-2022-1586 & CVE-2022-1587.

Link: https://blog.byte.builders/post/voyager-v2022.06.20/

#### Introducing KubeVault v2022.06.16

We are very excited to announce the release of KubeVault v2022.06.16 Edition. The KubeVault v2022.06.16 contains VaultServer latest api version v1alpha2, update to authentication method with addition of JWT/OIDC auth method. A new SecretEngine for MariaDB has been added, KubeVault CLI has been updated along with various fixes on KubeVault resource sync.

Link: https://blog.byte.builders/post/kubevault-v2022.06.16/



# Webinars


#### Using GKE Workload Identity with Stash

On June 29, 2022, Appscode held a webinar on Using GKE Workload Identity with Stash. 

Key contents of the webinar:
- Stash overview
- Setup GKE workload identity
- Backup and Restore a database using workload identity

Link: https://blog.byte.builders/post/stash-webinar-2022-06-29/

#### Deploy Sharded Redis Cluster on Kubernetes using KubeDB

On 22nd June 2022, Appscode held a webinar on ”Deploy Sharded Redis Cluster On Kubernetes using KubeDB”. 

The essential contents of the webinar:
- Introducing the concept of Redis Shard
- Challenges of running Redis on Kubernetes
- What KubeDB offers to face those challenges
- Live Demonstration

Link: https://blog.byte.builders/post/kubedb-redis-cluster-webiner-2022.06.22/

#### Deploy TLS secured ProxySQL Cluster for KubeDB provisioned MySQL Group Replication in Kubernetes

On JUN 8, 2022, AppsCode held a webinar on “Deploy TLS secured ProxySQL Cluster for KubeDB provisioned MySQL Group Replication in Kubernetes”.

The essential contents fo this webinar:
- ProxySQL initial setup made easy with KubeDB
- TLS secured frontend and backend connections for ProxySQL
- Load balance with ProxySQL cluster
- Failover recovery of ProxySQL cluster node
- Custom configuration in a declarative way

Link: https://blog.byte.builders/post/proxysql-webinar-2022.06.08/

#### Managing MongoDB with Arbiter on Kubernetes using KubeDB

On 1st June 2022, Appscode held a webinar on Managing MongoDB with Arbiter on Kubernetes using KubeDB. 

Key contents of the webinar:
- Introducing the concept of MongoDB Arbiter
- Importance of voting in MongoDB
- How Arbiter works on Replica-set database & Sharded cluster.
- Live Demonstration

Link: https://blog.byte.builders/post/mongodb-arbiter-2022.06.01/



# Blogs Published


#### A workaround of adding custom container to KubeDB managed Databases

Let’s assume you have a KubeDB managed Database deployed in your Kubernetes environment. Now, You want to inject a sidecar container in the database StatefulSet in order to extend and enhance the functionality of existing containers. Currently, KubeDB doesn’t have support for custom container insertion yet. We will discuss a workaround to run a custom container along with the managed-containers.

Link: https://blog.byte.builders/post/add-custom-container-to-kubedb-databases/

#### Kubernetes Backup and Restore - A Complete Solution

Link: https://blog.byte.builders/post/kubernetes-backup-and-restore/

#### Load Balance MySQL Group Replication with TLS secured ProxySQL Cluster

ProxySQL is an open source high performance, high availability, database protocol aware proxy for MySQL. 
From the KubeDB release v2022.05.24 we have added ProxySQL support for KubeDB MySQL. Now you can provision a ProxySQL server or cluster of ProxySQL servers with declarative yamls using KubeDB operator.

Link: https://blog.byte.builders/post/proxysql-one-2022.06.01/

#### Run Elasticsearch with SearchGuard Plugin in Azure Kubernetes Service (AKS) Using KubeDB

KubeDB simplifies Provision, Upgrade, Scaling, Volume Expansion, Monitor, Backup, Restore for various Databases in Kubernetes on any Public & Private Cloud. Here is how to Run & Manage Elasticsearch with SearchGuard Plugin in Azure Kubernetes Service (AKS) Using KubeDB.

Link: https://blog.byte.builders/post/run-elasticsearch-in-aks/



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

Follow our Linkedin for more [AppsCode Inc](https://www.linkedin.com/company/appscode/)

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).
