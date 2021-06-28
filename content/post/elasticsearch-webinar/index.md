---
title: Managing Production Grade Elasticsearch in Kubernetes Using KubeDB - Webinar
date: 2021-06-25
weight: 25
authors:
  - Shohag Rana
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - mariadb
  - memcached
  - mongodb
  - mysql
  - postgresql
  - redis
  - kubedb
---
## Summary

AppsCode held a webinar on "Managing Production Grade Elasticsearch in Kubernetes Using KubeDB". This took place 24th June 2021. The contents of what took place at the webinar is shown below:

1) What makes an Elasticsearch Cluster production-grade?
2) Why KubeDB Managed Elasticsearch?
3) Demo
    * Deploy TLS secure Elasticsearch
    * Version Upgrade
    * Horizontal Scale up and down
    * Delete and Restore from Backup
4) Q & A Session

## Description of the Webinar Demo

From this demo we get an in depth view of how the KubeDB Elasticsearch operator works. Firstly, we can see the TLS enabled deployment of Elasticsearch. Secondly, we can see the smart upgrade operation. By smart we mean that:

* It will disable the ongoing shard allocation, so that no data interrupted.
* It will upgrade one pod at a time and will wait for it to join the cluster before moving to the next one.
* It will restart pods in order. First, it will restart the ingest nodes, then the data nodes, and finally the master nodes.

Thirdly, we can see the horizontal scale up and down operations done flawlessly. It is described in the video about how the operator will scale up one pod at a time and it will wait for the pods to join the cluster. Once the pod has joined the cluster, it will move to the next one.
Again, Scaling down is comparatively easy. For the ingest node  the operator just removes the node.
While scaling down the master, the operator makes sure that it is removed from leader election.
Lastly, while scaling down the data node, the operator makes sure that all the shards safely move to other data nodes before removing the node.
In the last part of the video we can see how to backup and restore the elasticsearch cluster using Stash. All in all, it was an effective webinar.

Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/0mqPs6odwKk" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.06.23/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.06.23/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
