---
title: Announcing KubeDB v2022.03.28
date: 2022-03-28
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
  - elasticsearch-dashboard
  - kibana
  - schema-manager
---

We are pleased to announce the release of [KubeDB v2022.03.28](https://kubedb.com/docs/v2022.03.28/setup/). This release is a bug fix release for v2022.02.22 . In this release we have fixed a memory leak in Postgres sidecar (known as `pg-coordinator`) which will cause the postgres pod to restart due to OOMKill by Kubernetes. Our regularly scheduled feature release is planned to be out in 2 weeks. If you are not affected by this particular issue, you can ignore this release. You can find the detailed change logs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2022.03.28/README.md).

## Postgres side-car Memory Leak

KubeDB injects a sidecar image to Postgres containers for doing leader election and failover. This is known as pg-coordinator. In release v2022.02.22 various improvements were made to this sidecar image. One of those improvements used `kubectl exec` to check the status of `postgres` server process. We have received reports from multiple clients that this will result in memory leak (actually, Go routine leak) due to the connection not properly closed in Kubernetes version 1.19 and 1.20 . You can check whether you are experiencing this issue by running `kubectl get pods` command for your database pods and looking at the restart count. If the restart counts are high and keep increasing with time (in 1-2 hours), then you are probably experiencing this issue.

In today's release, we have modified this logic to address this memory leak issue. To fix, please upgrade the KubeDB operator to this release from v2022.02.22 following the instructions [here](https://kubedb.com/docs/v2022.03.28/setup/upgrade/). Once the operator is upgraded, please make some changes to the postgres custom resource objects (eg, add an annotation, etc.) so the operator reconciles the StatefulSet. To confirm that the reconciliation has happened, you can run a command like `kubectl get sts <pg_sts_name> -o yaml | grep image` and confirm that it is using `kubedb/pg-coordinator:v0.10.0` image and not the buggy `kubedb/pg-coordinator:v0.9.0` image. After that restart the database pods one by one so that the pods start using this image.

## GO 1.18

In this release, all the binaries are built using the GO 1.18 compiler. We don't expect any user facing impact of this.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.03.28/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.03.28/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
