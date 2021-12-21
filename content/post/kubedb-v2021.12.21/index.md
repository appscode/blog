---
title: Announcing KubeDB v2021.12.21 (Includes Log4j CVE Fixes)
date: 2021-12-21
weight: 25
authors:
  - Md Kamol Hasan
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

We are pleased to announce the release of KubeDB `v2021.12.21`. The headline feature of this release is that it has the support for **log4j CVEs fixed** images: `elasticsearch:7.16.2`,  `elasticsearch:6.8.22`, and `opensearch:1.2.2`. These docker images are using `Log4j 2.17.0`.

## ElasticsearchVersion

The corresponding ElasticsearchVersion CROs for the elasticsearch:7.16.2, elasticsearch:6.8.22, and opensearch:1.2.2 are:

```bash
$ kubectl get esversion 
NAME                   VERSION   DISTRIBUTION   DB_IMAGE                                          DEPRECATED   AGE
kubedb-xpack-7.16.2    7.16.2    KubeDB         kubedb/elasticsearch:7.16.2-xpack-v2021.12.24                  12s
opensearch-1.2.2       1.2.2     OpenSearch     opensearchproject/opensearch:1.2.2                             12s
xpack-6.8.22           6.8.22    ElasticStack   elasticsearch:6.8.22                                           12s
xpack-7.16.2           7.16.2    ElasticStack   elasticsearch:7.16.2                                           12s
```

**N.B.**: `kubedb-xpack-7.16.2` comes with the pre-installed snapshot plugins: `repository-s3`, `repository-azure`, `repository-hdfs`, and `repository-gcs`.

Sample Elasticsearch YAML:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: sample-es
  namespace: demo
spec:
  version: opensearch-1.2.2
  replicas: 3
  enableSSL: true 
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io
  terminationPolicy: "WipeOut"
```

## Upgrade from Older Version

Say, you are using Elasticsearch from OpenSearch distribution with version `opensearch-1.1.0`. To upgrade to the latest `opensearch-1.2.2`, use `ElasticsearchOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: sample-es-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: sample-es
  upgrade:
    targetVersion: opensearch-1.2.2
```

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/latest/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
