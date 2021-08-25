---
title: Announcing KubeDB v2021.08.23
date: 2021-08-24
weight: 25
authors:
  - Tamal Saha
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

We are pleased to announce the release of KubeDB v2021.08.23. This post lists all the major changes done in this release since `v2021.06.23`. This release offers support for the latest `Kubernetes version 1.22`. The `KubeDB CLI` now has exciting new features. `MongoDB` now uses the official docker images. `Elasticsearch` supports the latest **xpack and opendistro versions** and provides pre-built Docker images with snapshot plugins. KubeDB managed `Redis` now provides Password Authentication for the default user. KubeDB v2021.08.23 brings further changes to the Community Edition and deprecates prior releases of KubeDB operators. The detailed commit by commit changelog can be found [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2021.08.23/README.md).

## **Kubernetes 1.22**

As you may know, [Kubernetes 1.22 removed several deprecated beta APIs](https://kubernetes.io/blog/2021/07/14/upcoming-changes-in-kubernetes-1-22/) that were used by KubeDB. So, we have updated KubeDB project to use the corresponding GA version of those APIs. With this change, KubeDB v2021.08.23 supports Kubernetes 1.16 - 1.22 .

## **KubeDB CLI**

In this release, we’ve added some exciting features in KubeDB CLI. The CLI now has some new commands to make your database administration easier. The commands are listed below:

* **connect**: The connect command is used to connect to the shell of a database, where you can run your database commands.
* **exec**: Using the exec command, you can execute a script file or run a database command directly via flag without connecting to the database shell. For example,  you can run a `javascript` file in `MongoDB` database or a `sql` file in `MySQL` or `Postgres` database using the exec command.
* **show-credentials**: This command is used to print the credentials i.e. the root username and password to connect to the database.
* **pause**: The pause command is used to pause a database so that the KubeDB operators don’t process any changes made to the database `CRO` (Custom Resource Object).
* **resume**: You can use the resume command when you want to resume the database from a paused state. After resuming a database, the KubeDB operators will start processing the database again. The command also waits for the database to sync properly before exiting.

## **Elasticsearch**

* Elasticsearch versions support: `xpack-7.14.0`, `opendistro-1.13.2`
* KubeDB managed Elasticsearch now provides Elasticsearch docker images with pre-installed snapshot plugins: repository-s3, repository-azure, repository-hdfs, and repository-gcs. ElasticsearchVersion with snapshot plugins:  `kubedb-xpack-7.14.0`, `kubedb-xpack-7.13.2`, `kubedb-xpack-7.12.0`, and `kubedb-xpack-7.9.1`. You can find a detailed tutorial [here](https://kubedb.com/docs/v2021.08.23/guides/elasticsearch/plugins-backup/overview/).
* While using plugins to take snapshots, users need to provide secure settings. KubeDB allows you to provide secure settings through a k8s secret. Now, users can also provide `KEYSTORE_PASSWORD` to secure the `elasticsearch.keystore`.

```bash
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: sample-es
spec:
  secureConfigSecret:
     name: k8s-secret-name-with-settings
```

* KubeDB supports hot-warm clustering for Opendistro of Elasticsearch.
* Includes various feature improvements and bug fixes.

## **Redis**

KubeDB managed Redis now provides `Password Authentication` for Default user.

## **MongoDB**

* Previously, the MongoDB database docker image was maintained by KubeDB as we had to insert some scripts inside the official MongoDB image. But to give our users a more reliable experience, we decided to use the MongoDB official images without any modifications. From this release, we are using `MongoDB official docker images` for the provisioning of the MongoDB database using KubeDB.
* Previously, KubeDB enterprise operator printed the intermediate logs of a `MongoDBOpsRequest` which flooded the KubeDB enterprise operator logs with unnecessary logs. Now, we’ve changed the log verbosity of the intermediate logs. So, the intermediate logs of the `MongoDBOpsRequest` don’t get printed anymore. If you want to see the intermediate logs, you can change the log level of KubeDB enterprise operator to `--v=4`.

## Changes to Community Edition

Kubernetes v2021.08.23 Community Edition will manage database custom resources in Kubernetes `demo` namespace. Community Edition is feature limited and with this change we are making it clear to end users that Community Edition is primarily targeted for demo use-cases and the Enterprise Edition is targeted for production usage. To manage databases in any namespace, please try the Enterprise Edition.

## Deprecating Previous KubeDB Releases

KubeDB as a product has evolved quite a bit since [our decision to adopt an open/core model](https://blog.byte.builders/post/relicensing/) for the project last year and provide a sustainable future for the project. With this release, we are announcing the deprecation of all prior KubeDB releases. The previous versions of KubeDB operator will become unavailable by Dec 31, 2021. So, we encourage users to upgrade to the latest version of KubeDB.

## KubeDB Operator uses `kubedb` namespace

In previous releases, the KubeDB documentation will show KubeDB operator deployed in `kube-system` namespace. From this release, we have updated the documentation to use the `kubedb` namespace for installing the operator. This is generally recommended to deploy operators in their own namespace instead of `kube-system` namespace. But you can still deploy KubeDB is the `kube-system` namespace or any other namespace you like using Helm.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
