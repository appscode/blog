---
title: Deploy TLS secured ProxySQL Cluster for KubeDB provisioned MySQL Group Replication in Kubernetes
date: 2022-06-08
weight: 20
authors:
- Tasdidur Rahman
tags:
- kubedb
- proxysql
- mysql
- kubernetes
- provisioner
- ops-manager
---

## Summary

On JUN 8, 2022, AppsCode held a webinar on **"Deploy TLS secured ProxySQL Cluster for KubeDB provisioned MySQL Group Replication in Kubernetes"**. <br>
The essential contents fo this webinar are: <br>
* ProxySQL initial setup made easy with KubeDB
* TLS secured frontend and backend connections for ProxySQL
* Load balance with ProxySQL cluster
* Failover recovery of ProxySQL cluster node
* Custom configuration in a declarative way
* Q&A session


## Description of the Webinar

The speaker started the webinar with a basic introduction of ProxySQL. Then he explains how KubeDB operator eases the initialization process of ProxySQL in kubernetes.

Later in the demonstration a KubeDB ProxySQL yaml was explained and using that yaml a ProxySQL cluster was deployed in the cluster. After showing the successful connection establishment with the backend KubeDB MySQL a new user was introduced to the scenario and TLS secured connections were checked for that user on both ProxySQL frontend and backend.

Query loads were sent over the ProxySQL cluster and load distribution was observed. Then the demonstrator created a node failover scenario and showed how KubeDB ProxySQL recovered from this.

At last, the speaker talked about custom configuration and how KubeDB ProxySQL supports this in a declarative manner.

Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/oFi37vqCjcw" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.05.24/setup/).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/proxysql/webinar-2022.06.08).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
