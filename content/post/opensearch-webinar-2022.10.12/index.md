---
title: Day 2 Lifecycle Management for OpenSearch Cluster Using KubeDB
date: "2022-10-13"
weight: 20
authors:
- Raihan Khan
tags:
- kubedb
- kubernetes
- opensearch cluster
- ops requests
- provisioner
- reconfigure
- scaling
- upgrading
---

## Summary

AppsCode held a webinar on **“OpenSearch OpsRequests - Day 2 Lifecycle Management for OpenSearch Cluster Using KubeDB”** on 12th October 2022. The contents discussed on the webinar:
- OpenSearch OpsRequest and how it works
- Supported OpsRequests for OpenSearch
- Supported OpenSearch clustering methods
- Deploy OpenSearch Cluster and OpenSearch-Dashboards Using KubeDB
- Horizontal Scaling
- Vertical Scaling
- Version upgrade
- Restart your cluster

## Description of the Webinar

It is required to install the followings to get started:
- KubeDB Provisioner 
- KubeDB Ops-Manager
- KubeDB Dashboards

Live demo of the webinar is started with provisioning a TLS secured `OpenSearch Cluster` of 4 nodes using `KubeDB Provisioner` Operator and a TLS secured single node `OpenSearch-Dashboards` using `KubeDB Dashboards` operator referring to the `OpenSearch` database.

When the cluster and it's dashboard got ready, Horizontal scaling is demonstrated by horizontally upscaling of the cluster using KubeDB `HorizontalScaling` OpsRequest. Then the OpenSearch cluster is scaled vertically with KubeDB `VerticalScaling` OpsRequest. Memory and CPU of the database containers were vertically scaled using this OpsRequest. Some basic CRUD operations using the `OpenSearch-Dashboards` was also shown in the process. 

After that, the initially provisioned `OpenSearch-1.2.2` was upgraded to `OpenSearch-1.3.2`  using KubeDB `Upgrade` OpsRequest. The `OpenSearch-Dashboards` also got upgraded to compatible version autonomously triggered by the `KubeDB Dashboards` operator after the `OpenSearch` cluster got upgraded. 

Lastly, It’s demonstrated how to restart the `OpenSearch Cluster` using KubeDB `Restart` OpsRequest.

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/gSoWaVV4iQo" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs).
* Check out the detailed documentations for OpenSearch OpsRequest [here](https://kubedb.com/docs/v2022.08.08/guides/elasticsearch/concepts/elasticsearch-ops-request/)
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/OpenSearch/webinar-2022.10.12).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
