---
title: Configure MongoDB Hidden-Node on Kubernetes using KubeDB
date: 2022-09-22
weight: 20
authors:
  - Arnob kumar saha
tags:
  - kubedb
  - MongoDB
  - Hidden
  - Arbiter
  - kubernetes
  - replicaset
  - shard
  - cert-manager
  - inMemory
  - opsRequest
  - Vertical scaling
  - Horizontal scaling
  - Reconfigure
  - Volume expansion
  - enterprise
  - community
---

## Summary

On 21st September 2022, Appscode held a webinar on **Configure MongoDB Hidden-Node on Kubernetes using KubeDB**. Key contents of the webinars are:
- Introducing the concept of MongoDB Hidden-Node
- How Arbiter works on Replica-set database & Sharded cluster.
- Automatic failover & high availability
- Live Demonstration
- Q & A Session.



## Description of the Webinar

Earlier in this webinar, We discussed the `Hidden-Node` feature for `MongoDB`. Where we showed how to use kubeDB’s newly introduced feature MongoDB hidden node.
We discussed the core concept behind replicaSet & shared-cluster architectures. We also talked about the primary-Election mechanism for automatic failover.


Later in this webinar, We created a hidden-node enabled replicaset database & a `verticalScale` opsRequest to show how to scale the resources for each pod.
We then applied a `reconfigure` opsRequest to change the db-specific internal configurations on the fly.
We also applied a hidden-node enabled sharded cluster, and `volume expansion` & `horizontal scaling` opsRequest on it, to show the scaling of volumes & pod counts respectively.



Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/4sVig7wJzug" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.09.09/welcome/).

* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2022.07.09/welcome/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
