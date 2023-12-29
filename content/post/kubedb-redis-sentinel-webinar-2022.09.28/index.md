---
title: Redis Sentinel Ops Requests - Day 2 Lifecycle Management for Redis Sentinel Using KubeDB
date: "2022-09-29"
weight: 20
authors:
- Abdullah Al Shaad
tags:
- cert-manager
- high-availability
- horizontal-scaling
- kubedb
- kubernetes
- ops-manager
- redis
- sentinel
---

## Summary

On 28th September 2022, Appscode held a webinar on ”Redis Sentinel Ops Requests - Day 2 Lifecycle Management for Redis Sentinel Using KubeDB”. The essential contents of the webinars are:
- Introducing the concept of Redis Sentinel
- Deploy and Manage a Cluster in Redis Sentinel Mode
- Live Demonstration
- Q & A Session

## Description of the Webinar

Earlier in this webinar, We discussed the `Sentinel` mode in `Redis`. In Sentinel mode there are 
two types of cluster: Redis database cluster and Sentinel monitoring cluster. Sentinel Cluster
monitor and act as source of truth for the Redis Cluster. Sentinel Cluster
initiates automatic failover when master is in unreachable state and promotes a healthy
replica to master. Other replicas configures themselves with the new master.

Later in this webinar, we explored how to do different ops-requests using `KubeDB`. Then we
discussed how we can scale up Redis database replicas as well as Sentinels horizontally . We also discussed how to 
replace Sentinel Cluster of a Redis Cluster using RedisOpsRequest.
In the last demonstration, we explored how to make the database connections TLS secured using `cert-manager`.

Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/LToGVt1-D50" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.09.22/setup/).

* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2022.09.29/setup/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
