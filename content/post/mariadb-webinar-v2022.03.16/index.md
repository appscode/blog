---
title: MariaDB Alerting and Multi-Tenancy Support by KubeDB
date: 2022-03-16
weight: 20
authors:
  - Md. Alif Biswas
  - Tasdidur Rahman
tags:
  - kubedb
  - mariadb
  - multi-tenancy
  - schema-manager
  - alerting
  - prometheus
  - galera
  - kubernetes
  - stash
  - ops-manager
---

## Summary

On 16th March 2022, AppsCode held a webinar on **"MariaDB Alerting and Multi-Tenancy Support by KubeDB"**. Speakers talked about the following contents:

**MariaDB Alerts**
- Database Server Alerts
- Galera Clustering Alerts
- Custom Resource Alerts (MariaDB, MariaDBOpsRequest)
- Stash Alerts

**MariaDB Schema Manager**
- Create Database Schema
- User Management using vault
- Initialize database from volume source
- Restore snapshot using stash


## MariaDB Alerts

The webinar is started with showing the workstation where a three node `MariaDB Cluster` and a `Stash` backup configuration were running. Next, Speaker talked about a values.yaml file which was responsible for deploying different group of MariaDB alerts. Different database related alerts like `MySQLInstanceDown`, `MySQLServiceDown` were deployed and fired by manually shutting down the MariaDB servers running on cluster. 

Afterwards, Speaker talked about the galera replication latency alert which was triggered by running a script that created a load on `MariaDB Clutser`. The alert status was initially updated to `PENDING` and after a minute the alert was `FIRING`.

Later, KubeDB Custom Resource alerts were deployed to illustrate alerts that cause under different failure scenarios of KubeDB CRs. To give an idea, A OpsReq was created with expected failing configuration which activated the Custom Resource Alerts.

In conclusion of first part, MariaDB Stash alerts were deployed and tested by creating a backup session with bad secret data. The alert fired succesfully upon the failure of backup session.


## MariaDB Multi-Tenancy 

In the second part of the webinar speaker discussed about multi tenancy support for `Mariadb`, where several databases schemas were creted using kubeDB’s upcoming feature Schema Manager. Creating database, altering them, initializing with script and restoring snapshot using Schemas were demonstrated. User management and security concerns regarding multi tenancy implementation were also discussed in a brief.


Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://youtube.com/embed/P8l2v6-yCHU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.02.22/setup/).
* Find the sample yamls from webinar [here](https://github.com/kubedb/project/tree/master/demo/mariadb/webinar-2022.03.16).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
