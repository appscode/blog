---
title: Announcing KubeDB v2021.06.23 and Stash v2021.06.23
date: 2021-06-23
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

We are pleased to announce a dual release of [KubeDB v2021.06.23](https://kubedb.com/docs/v2021.06.23/setup/) and [Stash v2021.06.23](https://stash.run/docs/v2021.06.23/guides/latest/backends/overview/). This post lists all the major changes done in this release since `v2021.04.16`. This release offers support for the latest Kubernetes version v1.21.1, MongoDB 4.4.2, Elasticsearch 7.13.2. This release adds support for automated Day 2 operations for MariaDB and PostgreSQL databases. There has been various bug fixes across the board that improves the fault tolerance of the KubeDB operator. You can find the detailed change logs here: https://github.com/kubedb/CHANGELOG/blob/master/releases/v2021.06.23/README.md

## **Elasticsearch**

* Supports **hot-warm-cold** clusters for Elasticsearch from ElasticStack.
* Supports **machine learning** nodes for Elasticsearch from ElasticStack.
* Also supports roles like: data_content, data_frozen, and transform.
* New ElasticsearchVersion: **xpack-7.13.2**, xpack-7.12.0-v1, xpack-7.9.1-v2, xpack-6.8.16.
* Deprecated ElasticsearchVersion: xpack-7.9.1, xpack-7.8.0, xpack-7.7.1, xpack-7.6.2, xpack-7.5.2, xpack-7.4.2, xpack-7.3.2, xpack-7.2.1, xpack-7.1.1, xpack-6.8.10.
* Add timeout for ElasticsearchOpsRequest.
* Stash checks are skipped in Ops Request, if Stash is not installed.
* Fixed health check even with bas Elasticsearch deployments.
* Now the **heap size** of Elasticsearch node is **50% of Pod’s memory limit**.
* KubeDB no longer sets default cpu limits. Only cpu requests and memory limits are set by default.
* Log level issues are fixed.

## **MongoDB**

* Added support for **MongoDB 4.4.2**
* MongoDB Exporter has been updated and a TLS connection failure bug has been fixed.
* Stash checks are skipped in Ops Request, if Stash is not installed.
* Log level issues are fixed.
* MongoDB health check issues are fixed.
* KubeDB no longer sets default cpu limits. Only cpu requests and memory limits are set by default.

## **PostgreSQL**

* KubeDB now provisions PostgreSQL databases in  OpenShift clusters without privileged permission using fsGroup.
* Stash checks are skipped in Ops Request, if Stash is not installed.
* Health check issues are now fixed.
* Added some exciting Ops Request features including Upgrade, Vertical Scaling, Horizontal Scaling, Reconfigure Custom Configuration, Reconfigure TLS and Volume expansion.
* Log level issues are fixed.
* pg-coordinator can now recover in case a PostgreSQL cluster loses quorum.
* KubeDB no longer sets default cpu limits. Only cpu requests and memory limits are set by default.

## **MariaDB**

* Added some exciting Ops Request features including Upgrade, Vertical Scaling, Horizontal Scaling, Reconfigure Custom Configuration and Volume expansion.
* Custom configuration for MariaDB cluster is supported.
* MariaDB health check issues are fixed.
* Log level issues are fixed.
* Stash checks are skipped in Ops Request, if Stash is not installed.
* KubeDB no longer sets default cpu limits. Only cpu requests and memory limits are set by default.

## **MySQL**

* Log level issues are fixed.
* Stash checks are skipped in Ops Request, if Stash is not installed.
* MySQL health check issues are fixed.
* KubeDB no longer sets default cpu limits. Only cpu requests and memory limits are set by default.

## **Stash**

* This release offers the latest Kubernetes v1.21.1.
* ImagePullPolicy is now set to `IfNotPresent` instead of the previous `Always` which now saves valuable network traffic in the cluster.
* A bug that causes skipping backup due to name collision is now fixed. Details about the issue can be found [here](https://github.com/stashed/stash/issues/1341).
* It also fixes a PostgreSQL addon bug where backups were failing due to missing `sslmode` in the AppBinding. For more details, please refer to [here](https://github.com/stashed/postgres/pull/801).

## Non user facing Changes

In this release, we have updated KubeDB and Stash codebase to use Kubernetes v1.21.1 client libraries. This sets us up for removing support for deprecated api versions in upcoming Kubernetes 1.22 release. In this release, we have also introduced an built-in auditor that collects analytics data for billing purposes. This will be uses in a future release to prepare usage based billing reports for our PAYG customers. This is an open-source feature. You can see how we collect the data and what we collect [HERE](https://github.com/bytebuilders/audit).

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.06.23/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.06.23/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
