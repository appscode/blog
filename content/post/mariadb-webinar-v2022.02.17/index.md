---
title: MariaDB Auto-Scaling and Reconfiguration in Kubernetes Native way using KubeDB
date: "2022-02-24"
weight: 20
authors:
- Md. Alif Biswas
tags:
- autoscaling
- kubedb
- kubernetes
- mariadb
- ops-manager
- provisioner
- reconfigure
- stash
- volume-expansion
---

## Summary

AppsCode held a webinar on **“MariaDB Auto-Scaling and Reconfiguration in Kubernetes Native way using KubeDB”**. This took place on 17th Feb 2022. The contents of what took place at the webinar are shown below:
- Deploy KubeDB Provisioned MariaDB Cluster
- Volume Expansion
- Reconfigure MariaDB
- Storage Autoscaling
- Compute Autoscaling
- Q&A Session


## Description of the Webinar

It is required to install the followings to get started:
- KubeDB Provisioner 
- KubeDB Ops-Manager
- KubeDB Autoscaler
- Prometheus
- Metrics Server
- Vertical Pod Autoscaler
- Topolvm Provisioner(Or any storageclass that allows volume expansion)


Live demo of the webinar is started with provisioning a `MariaDB Cluster` of 3 nodes using `KubeDB Provisioner` Operator. When the cluster got ready, `Volume Expansion` OpsRequest by `KubeDB Ops-Manager`  is performed to expand the volume of each node from 1Gi to 2Gi. 

After that, Another OpsRequest `Reconfigure` is performed on the MariaDB cluster to reconfigure the MariaDB server values like _max_connections_ and _read_buffer_size_.  `Applyconfig` option of Reconfigure OpsRequest is also demonstrated.

Then speaker talked about the new feature `Autoscaler`  that can automatically scale different database resources like storage, memory and cpu. Following, A `Storage Autoscaler` was deployed on the MariaDB cluster which was triggered by raising the volume usage to more than 20%.

Lastly, It’s shown how to deploy `Compute Autoscaler` with some configurations and how it takes recommendation from vertical pod autoscaler and creates Vertical Scaling OpsRequest.

  Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/wg1kJGkFXdg" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.02.22/setup/).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/mariadb/webinar-2022.02.17).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
