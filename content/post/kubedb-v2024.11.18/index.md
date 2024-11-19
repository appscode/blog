---
title: Announcing KubeDB v2024.11.18
date: "2024-11-18"
weight: 14
authors:
- Tapajit Chandra Paul
tags:
- alert
- archiver
- autoscaler
- backup
- cassandra
- clickhouse
- cloud-native
- dashboard
- database
- druid
- grafana
- kafka
- kubedb
- kubernetes
- kubestash
- memcached
- mongodb
- mssqlserver
- network
- networkpolicy
- pgbouncer
- pgpool
- postgres
- postgresql
- prometheus
- rabbitmq
- redis
- restore
- s3
- security
- singlestore
- solr
- tls
- zookeeper
---

We are thrilled to announce the release of **KubeDB v2024.11.18**. This release introduces several key features, including:

- **OpsRequest Support**: Enhanced operational request capabilities for Druid, Microsoft SQL Server, PgBouncer, Solr, and ZooKeeper, providing greater management flexibility.

- **Autoscaling**: Added autoscaling support for FerretDB and Microsoft SQL Server to automatically adjust resources based on workload demands.

- **Network Policies**: We have added support for [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) in the release. Now Users can pass `--set global.networkPolicy.enabled=true` while installing KubeDB. The required network policies for operators will be created as part of the release process. And the required network policies for DB will be created by the operator when some database gets provisioned.

- **Backup & Restore**: Comprehensive disaster recovery support for Druid clusters using Kubestash (Stash 2.0), and manifest backup support for SingleStore.

- **New Version Support**: Added support for Druid version `30.0.1`.

For detailed changelogs, please refer to the [CHANGELOG](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2024.11.18/README.md). You can now explore the detailed features and updates included in this release.

## Cassandra

## Druid

In this release, we are introducing **TLS support for Apache `Druid`**. By implementing TLS support, Druid enhances the security of client-to-server communication within the cluster environment.

With TLS enabled, client applications can securely connect to the Druid cluster, ensuring that data transmitted between clients and servers remains encrypted and protected from unauthorized access or tampering. This encryption adds an extra layer of security, particularly important for sensitive data environments where confidentiality and integrity are paramount.

In addition to securing client-to-server communication, **internal communication** between Druid nodes is also encrypted. Furthermore, **connections to external dependencies**, such as metadata storage and deep storage systems, are secured.

To configure TLS/SSL in Druid, KubeDB utilizes cert-manager to issue certificates. Before proceeding with TLS configuration in Druid, ensure that cert-manager is installed in your cluster. You can follow the steps provided [here](https://cert-manager.io/docs/installation/kubectl/) to install cert-manager in your cluster.

To issue a certificate, cert-manager employs the following Custom Resource (CR):
Issuer/ClusterIssuer: Issuers and ClusterIssuers represent certificate authorities (CAs) capable of generating signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to fulfill the request. Further details can be found [here](https://cert-manager.io/docs/concepts/issuer/).

Certificate: cert-manager introduces the concept of Certificates, which define the desired x509 certificate to be renewed and maintained up to date. More details on Certificates can be found [here](https://cert-manager.io/docs/usage/certificate/).

Druid `CRD` Specifications:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster-tls
  namespace: demo
spec:
  version: 30.0.1
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: druid-ca-issuer
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

### Druid Ops-Requests:

We are introducing four new Ops-Requests for `Druid` i.e. Horizontal Scaling, Reconfigure, Update Version, Reconfigre TLS. You can find the example manifests files to perform these operations on a druid cluster below.

**Horizontal Scaling**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
    name: druid-hscale-up
    namespace: demo
spec:
    type: HorizontalScaling
    databaseRef:
        name: druid-cluster
    horizontalScaling:
        topology:
            coordinators: 2
            historicals: 2
```

**Update Version**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
    name: druid-update-version
    namespace: demo
spec:
    type: UpdateVersion
    databaseRef:
        name: druid-cluster
    updateVersion:
        targetVersion: 30.0.1
    timeout: 5m
    apply: IfReady
```

**Reconfigure**

```
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: reconfigure-drops
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: druid-cluster
  configuration:
    configSecret:
      name: new-config
```
Here, `new-config` is the name of the new custom configuration secret.

**Reconfigure-TLS**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-cluster
  tls:
    issuerRef:
      name: druid-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - druid
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```
This is an example showing how to add TLS to an existing druid cluster. Reconfigure-TLS also supports features like **Removing TLS**, **Rotating Certificates** or **Changing Issuer**.

### New Version Support

Support for Druid Version `30.0.1` has been added in this release and `30.0.0` is marked as deprecated.


## Elasticsearch

## FerretDB

## Kafka

## Memcached

## Microsoft SQL Server

## MongoDB

## PgBouncer

## Postgres

## SingleStore

## Solr

## ZooKeeper


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).