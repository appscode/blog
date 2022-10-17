---
title: Elasticsearch Hot-Warm-Cold Architecture Management with Kibana in Kubernetes
date: "2022-03-02"
weight: 20
authors:
- Kamol Hasan
- Raihan Khan
tags:
- community
- elasticsearch
- elasticsearchdashboard
- enterprise
- filebeat
- hot-warm-cold
- index-lifecycle-management
- k8s-log-monitoring
- kibana
- kubedb
- logstash
---

## Summary

On 2nd March 2022, Appscode held a webinar on **Elasticsearch Hot-Warm-Cold Architecture Management with Kibana in Kubernetes**. Key contents of the webinars are:

-  Deploy Hot-Warm-Cold Elasticsearch Cluster
-  Deploy Kibana Instance with ElasticsearchDashboard CR
-  Deploy Logstash and Filebeat
-  Monitor Kubernetes Logs with ELK Stack
-  Configure Index Lifecycle Management for Hot-Warm-Cold Nodes



## Description of the Webinar

Earlier in this webinar, they discussed in detail about `Hot-Warm-Cold Architecture` for `Elasticsearch`. They created an elasticsearch instance with hot-warm-cold architecture.

Later in this webinar, They discussed `Kibana` support for `Dashboard`. Where they created an `ElasticsearchDashboard` instance to deploy `Kibana`. They showed lots of customizations for deploying `ElasticsearchDashboard`. They also created an `Index Lifecycle Management` policy through kibana for Elasticsearch Hot-Warm-Cold Architecture. An `Index Template` was created to configure elasticsearch indices. They deployed `Logstash` and `Filebeat` for propagating logs to elasticsearch. After that, they showed a demo of monitoring kubernetes container logs from kibana. Finally, they performed some manual testing to ensure the Elasticsearch Hot-Warm-Cold architecture is implemented, and it is functioning as the `ILM` policy is set to be.


  Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://youtube.com/embed/R-eYc2cUXQY" title="Elasticsearch Hot-Warm-Cold Architecture Management with Kibana in Kubernetes" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.12.21/welcome/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
