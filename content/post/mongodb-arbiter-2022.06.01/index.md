---
title: Managing MongoDB with Arbiter on Kubernetes using KubeDB
date: 2022-06-01
weight: 20
authors:
- Arnob kumar saha
  tags:
- kubeDB
- MongoDB
- Arbiter
- kubernetes
- replicaset
- shard
- cert-manager
- cost reducing
- opsRequest
- verticalScaling
- Reconfigure
- enterprise
- community
---

## Summary

On 1st June 2022, Appscode held a webinar on **Managing MongoDB with Arbiter on Kubernetes using KubeDB**. Key contents of the webinars are:
- Introducing the concept of MongoDB Arbiter
- Importance of voting in MongoDB
- How Arbiter works on Replica-set database & Sharded cluster.
- Live Demonstration
- Q & A Session.



## Description of the Webinar

Earlier in this webinar, The speaker discussed about the `Arbiter` feature for `MongoDB`. Where He showed how to use kubeDB’s newly introduced features MongoDB-Arbiter. 
He discussed the core concept behind replicaSet & shared-cluster architectures. He also talked about the primary-Election mechanism for automatic failover.


Later in this webinar, He created an arbiter-enabled replicaset database & a verticalScale opsRequest to show how to scale the resources for each pod & to show how the election makes impact.
He also applied an arbiter-enabled sharded cluster & Reconfigure opsRequest to show how to change the db-specific internal configurations on the fly.



Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/QIDlhiEOvEg" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.12.21/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.01.11/setup/).

* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2021.11.24/setup/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
