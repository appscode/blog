---
title: Manage Percona XtraDB Cluster Day-2 Operations using KubeDB
date: 2022-10-10
weight: 20
authors:
- Md. Alif Biswas
tags:
- kubedb
- percona xtradb cluster
- kubernetes
- ops requests
- provisioner
- scaling
- upgrading
- reconfigure
---

## Summary

AppsCode held a webinar on **“Manage Percona XtraDB Cluster Day-2 Operations using KubeDB”** on 6th October 2022. The contents discussed on the webinar:
- Deploy Percona XtraDB Cluster Using KubeDB
- Horizontal Scaling
- Vertical Scaling
- Reconfigure Percona XtraDB Cluster
- Reconfigure TLS
- Upgrading
- Volume Expansion


## Description of the Webinar

It is required to install the followings to get started:
- KubeDB Provisioner 
- KubeDB Ops-Manager
- Cert Manager

Live demo of the webinar is started with provisioning a TLS secured `Percona XtraDB Cluster` of 3 nodes using `KubeDB Provisioner` Operator.
When the cluster got ready, Horizontal scaling is demonstrated by horizontally upscaling and downscaling of the cluster using KubeDB `HorizontalScaling` OpsRequest. Then the Percona 
XtraDB Cluster is scaled vertically with KubeDB `VerticalScaling` OpsRequest. Memory and CPU of the database containers were vertically scaled using this OpsRequest. 

After that, Reconfiguration of Percona XtraDB Cluster is demonstrated by updating mysql variable `max_connections` using KubeDB `Reconfigure` OpsRequest. Also, Existing TLS Certificate
configurations of Percona XtraDB Cluster were updated using KubeDB `ReconfigureTLS` OpsRequest. 

Later, The Percona XtraDB Cluster is updated from version `8.0.26` to version `8.0.28` using KubeDB `Upgrade` OpsRequest.

Lastly, It’s demonstrated how to expand the volume of existing Percona XtraDB Cluster using KubeDB `VolumeExpansion` OpsRequest.

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/PsMbpDHg_oU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/perconaxtradb/webinar-2022.10.06).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
