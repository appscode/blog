---
title: Challenges of Autoscaling Databases in Kubernetes
date: "2022-05-18"
weight: 20
authors:
- Md Kamol Hasan
- Md Fahim Abrar
tags:
- auto-scaling
- elasticsearch
- horizontal-scaling
- kubedb
- mongodb
- scaling
- vertical-scaling
---

## Summary

On 18th May 2022, AppsCode held a webinar on **Challenges of Autoscaling Databases in Kubernetes**. The key contents of the webinar are:


- What is a stateful workload?
- Kubernetes StatefulSet and its features?
- Why auto-scaling? Advantages?
- Problem with database horizontal scaling
- How does KubeDB implement managed horizontal scaling?
- Challenges of auto-scaling databases using VPA
- How does KubeDB implement auto-scaling? Auto-scaling architecture.
- Live demonstration of MongoDB sharded cluster auto-scaling


## Description of the Webinar

Earlier in the webinar, we discussed the stateful workloads. We also discussed, how can we deploy and manage stateful workloads in the Kubernetes cluster using the StatefulSet object. How auto-scaling on stateful workload benefits users, was also discussed during the session.

Scaling has become one of the very important features of stateful workloads. To meet the demand, KubeDB supports both horizontal and vertical scaling of databases. Users can trigger the scaling using an ops-request CRD. KubeDB also supports auto-scaling (only vertical for now) of the databases based on the recommendation from VPA. We talked about the architecture of the KubeDB auto-scaler and how it overcomes the challenges of automating the scaling process. Finally, a live demonstration of auto-scaling of KubeDB managed MongoDB sharded cluster.

  Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/y_qsmaFe4QI" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.03.28/welcome/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
