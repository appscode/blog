---
title: Deploy, Manage & Autoscaler inMemory MongoDB on Kubernetes using KubeDB
date: 2022-08-05
weight: 20
authors:
- Arnob kumar saha
tags:
- kubedb
- MongoDB
- inmemory
- kubernetes
- replicaset
- shard
- cert-manager
- availability
- opsRequest
- verticalScaling
- horizontalScaling
- vpa
- autoscaling
---

## Summary

On 3rd August 2022, Appscode held a webinar on **Deploy, Manage & Autoscaler inMemory MongoDB on Kubernetes using KubeDB**. Key contents of the webinars are:
- Introduction to InMemory MongoDB for replicaset & sharded Cluster
- Manual Failover & High Availability
- Vertical & Horizontal Scaling for InMemory databases
- AutoScaling with integrated VPA for InMemory Storage
- Live Demonstration
- Q & A Session.



## Description of the Webinar

Earlier in this webinar, We discussed about the `InMemory Storage Engine` feature for `MongoDB`. Where we showed how to use kubeDB’s feature inMemory Databases. 
We discussed the core concept behind replicaSet & shared-cluster architectures. We also talked about automatic failover & high Availability, And put some manual testing to show that.


On the live demo part, We created a replicaset database with StorageEngine inMemory & StorageType Ephemeral.
We created a `VerticalScaling` opsRequest to show how to scale the resources for each pod & to show how the election makes impact.
We also applied an inMemory sharded cluster & a `HorizontalScaling` opsRequest to show how to scale up the db horizontally.

Later in this webinar, We introduced our new VPA-integrated autoscaler-operator, & argued why one should use it. Then we have showed the autoscaler-operator in action by inserting a bunch of data using script.


Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/XOqR5GJ2mM4" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.05.24/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/welcome/).

* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2022.07.09/welcome/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
