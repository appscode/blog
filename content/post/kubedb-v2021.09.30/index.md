---
title: Announcing KubeDB v2021.09.30
date: "2021-09-30"
weight: 25
authors:
- Tamal Saha
tags:
- cloud-native
- database
- elasticsearch
- kubedb
- kubernetes
- mariadb
- memcached
- mongodb
- mysql
- postgresql
- redis
---

We are pleased to announce the release of KubeDB v2021.09.30. This post lists all the major changes done in this release since `v2021.09.09`. The headline features of this release are Redis Sentinel mode support and Offline volume expansion support for MongoDB. The detailed commit by commit changelog can be found [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2021.09.30/README.md).

## Redis

* Added support for provisioning Redis sentinel mode instances with sentinel monitoring
* Added TLS support for Sentinel Monitoring Cluster

## MongoDB

In this release, we have added support for Offline volume expansion of MongoDB nodes. So, KubeDB now supports both Online and Offline volume expansion. In offline volume expansion, it is required to delete a pod before expanding the volume of that pod. We've implemented the Offline volume expansion in such a way that only one pod at a time goes down, so your database cluster will continue to serve properly without downtime.

To support both Online and Offline volume expansion, we've added a new field `mode` under the `volumeExpansion` section of `MongoDBOpsRequest`. This field accepts two values `Online` and `Offline`. The default volume expansion mode is `Online`. If your `storageclass` doesn't have support for online volume expansion you can set the `mode` as `Offline` while applying your volume expansion ops request. An example of offline volume expansion of a MongoDB replicaSet cluster is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vol-exp-offline-rs
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-rs
  volumeExpansion:
    mode: Offline
    replicaSet: 20Gi
```

## MariaDB

* Added support for `mysqld` process restart on mariadb server failure.
* Fixed repetitive patches on KubeDB operators log for MariaDB.
* Fixed MariaDB Reconfigure TLS OpsReq early Successful status.

## MySQL

* Added support for MySQL 5.7.35 and 8.0.26 in catalog

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
