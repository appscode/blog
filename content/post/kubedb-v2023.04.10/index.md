---
title: Announcing KubeDB v2023.04.10
date: "2023-04-12"
weight: 14
authors:
- Dipta Roy
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

We are pleased to announce the release of [KubeDB v2023.04.10](https://kubedb.com/docs/v2023.04.10/setup/). This post lists all the major changes done in this release since the last release.
The release includes new changes like `One chart to Install KubeDB and Stash`, `Migration to GitHub Container Registry`, `Kafka Monitoring using Prometheus and Grafana`. Also, new verison support for `Kafka 3.3.2 , 3.4.0`, `MariaDB 10.11.2`, `MongoDB 6.0.5`, `MongoDB 5.0.15`, `Redis 7.0.10`, `Percona XtraDB 8.0.31` and bug fixes for `MySQL` and `Redis`.

You can find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.04.10/README.md).

## Migration to GitHub Container Registry
In this release we have migrated all docker images published by AppsCode to the GitHub Container Registry (ghcr.io) from the Docker Hub. Going forward, we are going to publish docker images exclusively to ghcr.io . This will resolve any issues related to rate limiting by Docker Hub. You can find the images by visiting the [LINK](https://github.com/orgs/kubedb/packages).

## One chart to Install KubeDB and Stash

We have created a new kubedb-one chart that includes both KubeDB and Stash. You can find the details [HERE](https://github.com/kubedb/installer/blob/master/charts/kubedb-one/Chart.yaml#L14-L49)
To use the chart, you can run a command like below:
```bash
helm upgrade -i kubedb appscode/kubedb-one \
  --version v2023.04.10 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt
```

## Kafka
### Monitoring using Prometheus & Grafana

KubeDB now offers support for monitoring Kafka clusters using Prometheus and Grafana dashboards. With the in-built JMX exporter, Kafka metrics can be easily exported to Prometheus via a service monitor, and Grafana dashboards can be created to visualize those metrics. KubeDB also provides a pre-built Grafana dashboard for Kafka metrics. For example, here's a sample YAML for deploying a TLS-secured Kafka cluster with built-in monitoring support.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka
  namespace: demo
spec:
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: kafka-ca-issuer
      kind: Issuer
  replicas: 3
  version: 3.4.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9091
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  storageType: Durable
  terminationPolicy: WipeOut
```
Here’s the KubeDB built [Grafana Dashboard](https://github.com/appscode/grafana-dashboards/tree/master/kafka)
Also, We have added support for Kafka version `3.3.2` & `3.4.0`. From this release, Kafka docker images will be using OpenJDK-based Java version 11 instead of Java 8 as it has been deprecated since `3.0.0`. 



## MongoDB
We have added the MongoDB version `6.0.5` and `5.0.15` in this release. To deploy a MongoDB replica-set instance with version `6.0.5`, you can apply this yaml:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mg
  namespace: demo
spec:
  version: 6.0.5
   replicaSet:
     name: "rs1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```



## MariaDB

We have added the latest MariaDB version `10.11.2` in this release. To deploy a MariaDB Standalone instance with version `MariaDB 10.11.2`, you can apply this yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: mariadb-standalone
  namespace: demo
spec:
  version: "10.11.2"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
**Change Default WSREP_SST_METHOD**: From this release, the default value for `spec.wsrepSSTmethod` will be `rsync`. Previously it was set to `mariabackup`. 



## Percona XtraDB
We have added the latest Percona XtraDB version `8.0.31` in this release. To deploy a Percona XtraDB Galera cluster with version `Percona XtraDB 8.0.31`, you can apply this yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: xtradb-galera
  namespace: demo
spec:
  version: "8.0.31"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```


## Redis
We have added the latest Redis version `7.0.10` in this release. To deploy a Redis Standalone instance with version `Redis 7.0.10`, you can apply this yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: 7.0.10
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
``` 
**Bug Fix**: When the database is created with the same name in multiple namespaces, multiple redis instances was trying to own a single `clusterrole`, this bug is fixed in this release. Now to grant redis instances access to sentinel instances, only one `clusterrole` is created without owner reference.


## MySQL

**Bug Fix**: Earlier MySQL instance was used to create a cluster role for reading the `mysqlversion` from the `mysql-coordinator`. When the database is created with the same name in multiple namespaces, multiple MySQL instances try to own a single cluster role, this bug is fixed in this release. Now to grant MySQL coordinator access to `mysqlversions`, only one cluster role `mysql-version-reader` is created without owner reference.




## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.04.10/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.04.10/setup/upgrade/).




## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
