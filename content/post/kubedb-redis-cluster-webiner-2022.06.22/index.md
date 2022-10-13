---
title: Deploy Sharded Redis Cluster on Kubernetes using KubeDB
date: "2022-06-23"
weight: 20
authors:
- Abdullah Al Shaad
tags:
- cert-manager
- cluster
- community
- enterprise
- horizontalscaling
- kubedb
- kubernetes
- opsmanager
- opsrequest
- reconfigure
- redis
- shard
- versionupgrade
- verticalscaling
---

## Summary

On 22nd June 2022, Appscode held a webinar on ”Deploy Sharded Redis Cluster On Kubernetes using KubeDB”. The essential contents of the webinars are:
- Introducing the concept of Redis Shard
- Challenges of running Redis on Kubernetes
- What KubeDB offers to face those challenges
- Live Demonstration
- Q & A Session

## Description of the Webinar

Earlier in this webinar, We discussed the `Cluster` mode in `Redis`. In cluster mode, we can
divide our data into different shards. In each shard, there is one master and one or more replicas.

Later in this webinar, we explored how `KubeDB Redis` handles different failover scenarios. Then we 
discussed how we can scale up database horizontally and vertically. We also discussed how to reconfigure 
the database with redis specific internal configurations and how to upgrade version to any latest versions.
In the last demonstration, we explored how to make the database connections TLS secured using `cert-manager`.

Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/J7QI4EzuOVY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.05.24/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.06.16/setup/).

* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2022.06.21/setup/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).

