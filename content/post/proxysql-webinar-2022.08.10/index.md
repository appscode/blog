---
title: ProxySQL Declarative Provisioning, Reconfiguration and Horizontal Scaling using KubeDB
date: 2022-08-12
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
- custom-configuration
- reconfigure
---

## Summary

On AUG 10, 2022, AppsCode held a webinar on **"ProxySQL Decalarative Provisioning, Reconfiguration and Horizontal Scaling using KubeDB"**. <br>
The essential contents fo this webinar are: <br>
* Declarative configuration for ProxySQL using KubeDB
* ProxySQL user synchronization with MySQL backend
* Reconfigure ProxySQL configuration using KubeDB Ops-request
* Horizontal scaling on KubeDB ProxySQL cluster
* Q&A session


## Description of the Webinar

The speaker started the webinar with a basic introduction of `Declarative Configuration` in KubeDB ProxySQL. Then he explains how KubeDB operator eases the initialization process of KubeDB ProxySQL with this feature. The concept of `SyncUser` was explained right after that. Then the theoretical discussion ended with `Ops-request` in KubeDB ProxySQL.

Later in the demonstration a KubeDB ProxySQL yaml with declarative configuration was explained and using that yaml a ProxySQL cluster was deployed in the kind cluster. After the ProxySQL server was ready all the configurations were checked to verify the process. A short demo of syncUser feature was given right after.

Reconfigure ops-requests were applied one after other and all the changes were observed both in the `Ultimate Configuration Secret` and in the actual ProxySQL server. 

At last, up-scaling and down-scaling with horizontal scaling feature was shown in the webinar and the speaker moved to the QnA session. 

Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/fT_cQDxfU9o" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/setup/).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/proxysql/webinar-2022.08.10).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
