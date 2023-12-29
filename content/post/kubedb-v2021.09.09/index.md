---
title: Announcing KubeDB v2021.09.09
date: "2021-09-13"
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

We are pleased to announce the release of KubeDB v2021.09.09. This post lists all the major changes done in this release since `v2021.08.23`. This release is primarily a bug fix release for v2021.08.23. We have also added support for MongoDB 5.0.2. The detailed commit by commit changelog can be found [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2021.09.09/README.md).

## KubeDB API

Various KubeDB supported databases currently uses a coordinator sidecar for failover and recovery of clustered databases. This includes PostgreSQL, MongoDB, MySQL, MariaDB and Redis. In this release we have added a `spec.coordinator` field so that you can customize the resources and security context of the coordinator container.

## KubeDB CLI

- --namespace flag stopped working in the last release. This regression has been fixed.
- `pause/resume` commands will not show error when Stash is not installed.

## Elasticsearch

- Customizing internal users will incorrectly cause a validation error in previous releases. This has been fixed.
- If an unknown monitoring agent type is passed in the `spec.monitor`, operator will not panic.

## Postgres

- Fix Memory leak issue from Coordinator. The database connections are properly closed to fix this issue.
- Add support to configure Coordinator resources.

## MongoDB

In this release, we are happy to announce that we have added support for the latest MongoDB version, `5.0.2`. You can try it by installing/upgrading to KubeDB `v2021.09.09` and deploying the MongoDB database with version `5.0.2`. One thing to note here, in this version, MongoDB sharding database backup isn't working as expected. This is a known bug for `5.0.2`. A fix has been committed in the  upstream MongoDB 5.0.3 version. We are going to release an update once MongoDB 5.0.3 is released.

## MariaDB

- Added support for full or partial crash recovery of MariaDB Cluster.
- Add, update, delete & rotate TLS configurations are supported.
- MariaDB version 10.6.4 support has been added.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
