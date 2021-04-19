---
title: KubeDB v2021.04.16- Improved Features and Bug Fixes
date: 2021-04-18
weight: 19
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

We are pleased to announce [KubeDB v2021.04.16](https://kubedb.com/docs/v2021.04.16/setup/). This post lists all the major changes done in this release since `v2021.03.17`.  This release offers the latest images for MySQL and MariaDB official databases. A new phrase `Pending` is addded to *MongoDb*, *MySQL* and *Elasticsearch* databases to specify that the `opsrequest` of that database has not been started. **Elasticsearch** now supports Custom User(UID). Various Bugs have been fixed in **Postgres**, **Elasticsearch**, **MongoDB**, **MariaDB** and **MySQL**.


## **Postgres**



*   Added custom **User** and **Group** support for postgres
*   Fixed permission denied issue for** Mount Root** directory


## **MongoDB**



*   `ipv6` is disabled automatically if the cluster doesn't support `ipv6`.
*   Prevent the operator pod from getting panicked from misconfiguration.
*   Now, the database `BackupConfiguration` is paused before the `MongoDBOpsRequest` starts processing and resumes the `BackupConfiguration` after the ops request is `Successful`.
*   Now, `MongoDBOpsRequest` waits for the `BackupSession` and `RestoreSession` running for this database to finish before it starts processing.
*   A new phase `Pending` is added to `MongoDBOpsRequest`. This phase specifies that the `MongoDBOpsRequest` hasn’t started processing yet.
*   A new field `timeoutSeconds` is added to MongoDBOpsRequest. If the ops request doesn't complete within the specified timeout, the ops request will fail.


# **Elasticsearch**



*   Added support for Custom User (UID).
*   Prevent the operator pod from getting panicked for misconfiguration.
*   Added support for ElasticsearchVersion: "xpack-7.12.0", "searchguard-7.10.2".
*   ElasticsearchVersion CRs use official docker registries instead of “kubedb”. For example: “kubedb/elasticsearch:7.0.1-xpack” is replaced with “elasticsearch:7.0.1”, “kubedb/elasticsearch:7.10.2-searchguard” is replaced with “floragunncom/sg-elasticsearch:7.10.2-oss-49.0.0”.
*   Now, the database `BackupConfiguration` is paused before the `ElasticsearchOpsRequest` starts processing and resumes the `BackupConfiguration` after the ops request is `Successful`.
*   Now, `ElasticsearchOpsRequest` waits for the `BackupSession` and `RestoreSession` running for this database to finish before it starts processing.
*   A new phase `Pending` is added to `ElasticsearchOpsRequest`. This phase specifies that the `ElasticsearchOpsRequest` hasn’t started processing yet.


## **MariaDB**



*   Updated custom mariadb image with mariadb official image.
*   Added graceful shutdown on mariadb pod restart.
*   Added operator-pod’s panic resistance for misconfiguration.


## **MySQL**



*   Added official mysql image support without any wrapper.
*   Along with DNS, Added support of IPv4 and IPv6 for mysql group replication’s internal configuration.
*   Now, the database `BackupConfiguration` is paused before the `MySQLOpsRequest` starts processing and resumes the `BackupConfiguration` after the ops request is `Successful`.
*   Now, `MySQLOpsRequest` waits for the `BackupSession` and `RestoreSession` running for this database to finish before it starts processing.
*   A new phase `Pending` is added to `MySQLOpsRequest`. This phase specifies that the `MySQLOpsRequest` hasn’t started processing yet.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.03.17/setup).
- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.03.17/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
