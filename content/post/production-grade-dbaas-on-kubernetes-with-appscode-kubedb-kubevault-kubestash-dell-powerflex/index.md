---
title: Production Grade DBaaS on Kubernetes with AppsCode KubeDB, KubeVault, KubeStash, Dell PowerFlex, Dell ECS and EKS Anywhere
date: "2024-03-28"
weight: 14
authors:
- Ambar Hassani
tags:
- cloud-native
- data-on-kubernetes
- dbaas
- dell
- eks-anywhere
- kubedb
---

## Overview

In this blog we will look to build a production grade DBaaS platform using AppsCode and Dell Technologies leading platforms simplifying the entire endeavor and allowing a true enterprise experience. Needless to say, the cost and operational implications of building such powerful DBaaS platforms can be immense if done correctly.

> As a collaborative effort, I had the pleasure of building this end-to-end deployment scenario with [Tamal Saha](https://www.linkedin.com/in/tamalsaha/), founder and CEO of AppsCode along with this engineering team. Special thanks to their partnership and helping deliver this experiential blog.


The below visual overviews the environmental setup used in this example, wherein the technology spectrum covers rendering a Microservices platform (sock-shop) with persistent databases. The use case of Sock-shop in a primitive sense is already covered in a different blog of mine located here.

AppsCode products ([KubeDB](https://kubedb.com/), [KubeVault](https://kubevault.com/) and [KubeStash](https://kubestash.com/)) are used in conjunction with Dell Storage platforms (PowerFlex and ECS) to serve the persistence layer and backup requirements. The entire exhibit runs on top of EKS Anywhere, however can be equally rendered on any other compliant Kubernetes platform.

![Infra](infra.png)

The comprehensive use-cases within the scope of this blog extend much beyond a typical “hello-world” type of demonstration. Examples being import of existing databases, using a distributed vault client-server architecture, auto-scaling, integration with observability tools, etc.

Being an end-to-end scenario, there is a wide range of technical comprehension and deep dives captured in the series of YouTube videos that demonstrate actual implementation.

*Let’s begin with a brief perspective and a great dialogue with Tamal Saha (Founder and CEO of AppsCode)*

<iframe width="560" height="315" src="https://www.youtube.com/embed/wERbfMWdC90?si=ITeZMnepVym3D544" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Next, we will spare some time to understand the demo context and environmental setup.

<iframe width="560" height="315" src="https://www.youtube.com/embed/WWPOchbVz-g?si=SXVBTHAitdqDVhtC" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Following the contextual understand, we will deploy the two EKS Anywhere clusters with the pre-requisites around Load-balancing, Ingress controller, CSI drivers, etc.

<iframe width="560" height="315" src="https://www.youtube.com/embed/oO62gPJl3Hg?si=PFJpCCzG8zZZIqGN" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Next, we will setup Secrets as a service with KubeVault. KubeVault is further paired up with HashiCorp’s Vault Secrets Operator to consume the secrets in the workload cluster.

<iframe width="560" height="315" src="https://www.youtube.com/embed/uYD7n_qdpTY?si=2mp7RAa-1OuHqm37" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Now that we have the EKS Anywhere clusters and KubeVault setup, the next video will focus on understanding KubeDB in greater detail with the CRD comprehensions. We will also deploy the operator and understand the operational details.

<iframe width="560" height="315" src="https://www.youtube.com/embed/VWPvbiiHWvU?si=4smO8VBmMPdyFXCE" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

With the KubeDB operator deployed and operational, we will setup the persistent databases in a high-availability cluster configuration using MySQL, MongoDB, Redis and RabbitMQ CRD. Along the way we will also understand how easily databases can be seeded with intial data (maybe a migration from some other DBaaS). In a tandem execution, we will also deploy the sock-shop microservices and also observe the leverage of database secrets sourced from KubeVault.

<iframe width="560" height="315" src="https://www.youtube.com/embed/5YEp3J_C8mc?si=rK41Z5ihO2yn3Stm" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

We will then follow up the deployment to setup the quintessential components of DBaaS backup and recovery using KubeStash and Dell ECS, which is a high performant S3 compatible storage solution.

<iframe width="560" height="315" src="https://www.youtube.com/embed/OWvSYsrGUeY?si=Ps92MYQpjhRoEUi5" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

The concluding act will demonstrate integration of DBaaS with Prometheus and Grafana for observability needs.

<iframe width="560" height="315" src="https://www.youtube.com/embed/kkPM25jM9Zs?si=6p4fxHO_wdx30uFs" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

Hopefully you had the patience to actually view all the videos and enjoyed the practicality of the use-cases. Also hoping that one could comprehend the benefits of KubeDB and how it can simplify adoption of DBaaS on Kubernetes.

*Lastly a big shout to Tamal Saha (Founder and CEO of AppsCode) and his entire team., first for building such awesome products and secondly partnering in this blog development!*

Cheers,
Ambar@thecloudgarage
#iwork4dell

>PS: This blog was initially published on [Medium](https://ambar-thecloudgarage.medium.com/production-grade-dbaas-on-kubernetes-with-appscode-kubedb-kubevault-kubestash-dell-powerflex-379b393ad98b)


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).