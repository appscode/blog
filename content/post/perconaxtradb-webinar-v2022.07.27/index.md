---
title: Managing Production Grade Percona XtraDB Cluster in Kubernetes using KubeDB
date: "2022-07-29"
weight: 20
authors:
- Md Alif Biswas
tags:
- galera
- kubedb
- kubernetes
- percona
- percona-xtradb
- provisioner
---

## Summary

AppsCode held a webinar on **“Managing Production Grade Percona XtraDB Cluster in Kubernetes using KubeDB”** on 27th July 2022. The contents discussed on the webinar:
- TLS Secured Percona XtraDB Cluster
- Cluster Failure Recovery
- Monitor Percona XtraDB Metrics
- Custom Configuration
- Live Demonstration
- Q&A Session


## Description of the Webinar

It is required to install the followings to get started:
- KubeDB Provisioner 
- KubeDB Ops-Manager
- Kube Prometheus Stack helm chart
- Cert Manager

Live demo of the webinar is started with provisioning a TLS secured `Percona XtraDB Cluster` of 3 nodes using `KubeDB Provisioner` Operator.
When the cluster got ready, Cluster failure recovery was simulated by deleting multiple nodes from the cluster and observing its automatic recovery without data loss. 

After that, Some metrics scraped from Percona XtraDB Cluster were shown from Prometheus UI and discussed on how to configure the PerconaXtraDB CR of KubeDB enable monitoring.

Lastly, It’s demonstrated how to deploy Percona XtraDB Cluster with custom configuration using KubeDB in simple and fast way.

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/YNC7CgIwje8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/perconaxtradb/webinar-2022.07.27).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
