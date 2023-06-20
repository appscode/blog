---
title: Announcing KubeDB v2023.06.19
date: "2023-06-19"
weight: 14
authors:
- Tamal Saha
tags:
- cloud-native
- dashboard
- database
- elasticsearch
- grafana
- kafka
- kibana
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- opensearch
- percona
- percona-xtradb
- pgbouncer
- postgresql
- prometheus
- proxysql
- redis
---

We are pleased to announce the release of [KubeDB v2023.06.19](https://kubedb.com/docs/v2023.06.19/setup/). In this release, we have primarly focused on bug fixes.

Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.06.19/README.md).

## Reducing load on kube-apiserver

In this release, KubeDB operator has been updated to cache access to Secrets. In case of sidecars, we have passed the secrets via environment variable. This will significantly reduce load on kube-apiserver.

## Add `enableServiceLinks` to db podSpec

Kubernetes exports an environment variable for every single service in the same namespace into every pod by default. This ends up being huge on large shared dev clusters. It appears to slow pod startup by a lot for things like elasticsearch, kafka, zookeeper. Now, you can disable injecting these environment variables by setting `enableServiceLinks: false` in the PodTemplate spec. 

## Memory leak fixes in Provisioner

A memory leak in the KubeDB Provisioner component was reported which will result to operator restart every 4-5 days. We have found a fixed a Go routine leak issue in the health checker. If the issue persists, please let us know.

## ProxySQL Improvements

In this release, we have changed the ProxySQL api such that spec.mode is not required any more. It will be auto-detected by the operator. Also, ProxySQL will run using non-root `proxysql` user (uid 999 for debian based images and 998 for centos based images) by default in this release.

## Postgres

We have added the PostgreSQL versions `15.3`, `14.8`, `13.11`, `12.15`, and `11.20` in this release.

## Elasticsearch

We have added the Elasticsearch versions `8.6.2` and `8.8.0` in this release.

## Auto acquire License in case of expired licenses

Previously KubeDB operator will only auto acquire license via license-proxyserver if no license file is provided. Now, KubeDB operator will also try to auto acquire license if the license file has expired.

## KubeDB Gateway

We have released KubeDB Gatway v0.0.1 based on [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/). This offers a Envoy based Gateway that can be to expose databases running on one cluster to another Kubernetes cluster. The helm chart can be found [here](https://github.com/voyagermesh/installer/tree/master/charts/gateway-helm). You can see a demo [here](https://youtu.be/l0UB7IZTZ44). Please reach out, if you are interested in trying it.

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.06.19/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.06.19/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
