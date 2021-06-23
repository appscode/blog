---
title: KubeDB v2021.06.23 and Stash v2021.06.23 - Latest Kubernetes Support
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

We are pleased to announce a dual release of [KubeDB v2021.06.23](https://kubedb.com/docs/v2021.06.23/setup/) and [Stash v2021.06.23](https://stash.run/docs/v2021.06.23/guides/latest/backends/overview/). This post lists all the major changes done in this release since `v2021.04.16`.This release offers support for the latest Kubernetes version 1.21.1. Many network traffic redundancies have been handled in multiple databases and also stash. Besides resolving some minor bugs, health check issues and klog log level issues have been resolved in all the databases. Major features in the MariaDB Ops Request have been added.

## **Postgres**

* Now, Stash checks are skipped in Ops Request, if Stash is not installed.
* health check issues are now fixed.
* Added some exciting Ops Request features including Upgrade, Vertical Scaling, Horizontal Scaling, Reconfigure Custom Configuration, Reconfigure TLS and Volume expansion.
* fixed log level issue.
* fixed some issues with pg-coordinator.
* Only the CPU request is set (if missing), keeping the CPU limit empty.

## **MongoDB**

* Added support for **MongoDB 4.4.2**
* Fixed MongoDB Exporter TLS connection failure bug.
* Now, Stash checks are skipped in Ops Request, if Stash is not installed.
* Klog log level issues are fixed.
* MongoDB health check issues are fixed.
* Only the CPU request is set (if missing), keeping the CPU limit empty.

## **Elasticsearch**

* Supports **hot-warm-cold** clusters for Elasticsearch from ElasticStack.
* Supports **machine learning** nodes for Elasticsearch from ElasticStack.
* Also supports roles like: data_content, data_frozen, and transform.
* New ElasticsearchVersion: **xpack-7.13.2**, xpack-7.12.0-v1, xpack-7.9.1-v2, xpack-6.8.16.
* Deprecated ElasticsearchVersion: xpack-7.9.1, xpack-7.8.0, xpack-7.7.1, xpack-7.6.2, xpack-7.5.2, xpack-7.4.2, xpack-7.3.2, xpack-7.2.1, xpack-7.1.1, xpack-6.8.10.
* Add timeout for ElasticsearchOpsRequest.
* Now, Stash checks are skipped in Ops Request, if Stash is not installed.
* Fixed health check.
* Now the **heap size** of Elasticsearch node is **50% of Pod’s memory limit**.
* Only the CPU request is set (if missing), keeping the CPU limit empty.
* Various code improvements and bug fixes.
* Klog log level issues are fixed.

## **MariaDB**

* Added some exciting Ops Request features including Upgrade, Vertical Scaling, Horizontal Scaling, Reconfigure Custom Configuration and Volume expansion.
* Now, Custom configuration for MariaDB cluster is supported.
* MariaDB health check issues are now fixed.
* Klog log level issues are fixed.
* Now, Stash checks are skipped in Ops Request, if Stash is not installed.
* Only the CPU request is set (if missing), keeping the CPU limit empty.

## **MySQL**

* Fixed Klog log level issues
* Now, Stash checks are skipped in Ops Request, if Stash is not installed.
* MySQL health check issues are fixed.
* Only the CPU request is set (if missing), keeping the CPU limit empty.

## **Stash**

* This release offers the latest Kubernetes v1.21.1.
* ImagePullPolicy is now set to `IfNotPresent` instead of the previous `Always` which now saves valuable network traffic in the cluster.
* A bug that causes skipping backup due to name collision is now fixed. Details about the issue can be found [here](https://github.com/stashed/stash/issues/1341).
* It also fixes the PostgreSQL bug where backups were failing due to missing `sslmode` in the AppBinding. For more details, please refer to [here](https://github.com/stashed/postgres/pull/801).
* Stash now has an Audit event that collects metadata about Stash resources, enabling us to add usage-based billing for PAYG users. It also enables us to better understand how Stash is being used in different environments so that we can improve the product. This is an open-source feature. You can see how we collect the data and what we collect [HERE](https://github.com/bytebuilders/audit). You can disable the feature easily. For disabling see the following code:

For community version:

```bash
$ helm install stash appscode/stash           \
  --version v2021.06.23                       \
  --namespace kube-system                     \
  --set features.community=true               \
  --set stash-community.enableAnalytics=false \
  --set-file global.license=/path/to/the/license.txt
```

For enterprise version:

```bash
$ helm install stash appscode/stash             \
  --version v2021.06.23                         \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set stash-enterprise.enableAnalytics=false  \
  --set-file global.license=/path/to/the/license.txt
```

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.06.23/setup).
* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.06.23/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
