---
title: Panopticon, A Generic Kubernetes State Metrics Exporter - Webinar
date: "2021-08-30"
weight: 20
authors:
- Shohag Rana
tags:
- cloud-native
- database
- elasticsearch
- kubedb
- kubernetes
- mariadb
- memcached
- metrics
- mongodb
- mysql
- panopticon
- postgresql
- prometheus
- redis
---

## Summary

AppsCode held a webinar on **"Panopticon: A Generic Kubernetes State Metrics Exporter"**. This took place on 26th August 2021. The contents of what took place at the webinar are shown below:

1) What is Panopticon?
2) Background story of Panopticon?
3) Key features of Panopticon.
4) Comparison between Panopticon and kube-state-metrics
5) Demo
    * Generate metrics for KubeDB MongoDB custom resource using Panopticon
    * Generate metrics for Deployment using Panopticon
6) Q & A Session

## Description of the Webinar

At present, there is `no available state metrics exporter` for Kubernetes custom resources. Also, Exporter(kube-state-metrics) is available for Kubernetes native resources but `users have no control over the metrics`. Besides, `No available generic exporter that is highly configurable` for any kind of Kubernetes resources. This is where **Panopticon** comes in. `Panopticon` is a Kubernetes Controller that watches Kubernetes resources passively and exports Prometheus metrics. In this webinar, the speaker continues to describe the `architecture` of the Panopticon and how it works. After that, he states the `key features` of Panopticon and how it is `different from kube-state-metrics`. Ater this the `Demo` portion of the webinar starts.

From this demo, we get an in-depth view of how the Panopticon works. At first, it is shown how to `install Panopticon`. After that, the speaker shows how to generate metrics for `KubeDB MongoDB custom resource` using Panopticon. The demo portion ends with showing how to generate metrics for `Deployment` using Panopticon. The webinar ends with a Q&A session.
All in all, it was an effective webinar.

Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/xDvna1MNBuc" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **Panopticon**, please follow the installation instruction from [here](https://byte.builders/blog/post/introducing-panopticon/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with Panopticon or want to request new features, please [file an issue](https://github.com/kubeops/installer).
