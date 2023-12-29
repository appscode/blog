---
title: Announcing KubeDB v2023.02.28
date: "2023-03-06"
weight: 25
authors:
- Mehedi Hasan
tags:
- cloud-native
- dashboard
- database
- elasticsearch
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
- postgresql
- proxysql
- redis
---

We are pleased to announce the release of [KubeDB v2023.02.28](https://kubedb.com/docs/v2023.02.28/setup/). This post lists all the major changes done in this release since the last release.
The release include new features `Combined PEM Certifiacte` `Postgres StandAlone to High Availibility`, `Acme Protocol based certificate support for ProxySQL & PgBouncer ` and  new verisons for `MySQL:8.0.32 , 5.7.41` ,
`OpenSearch 2.0.1 , 2.5.0` , `Redis: 6.0.18,6.2.11,7.0.9`. 

You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.02.28/README.md).

## Combined PEM Certificate

From this release, client certificates secret for TLS enabled databases now contain combined PEM certificates. 
The combined PEM certificate is stored under the key `tls-combined.pem` in the client certificate secret.
To enable this, the cert-manager needs to be deployed with feature-gate `AdditionalCertificateOutputFormats=true`. 
To install cert-manager with this flag using helm you can use the following command.
```bash
 helm install \
   cert-manager jetstack/cert-manager \
   --namespace cert-manager \
   --create-namespace \
   --set installCRDs=true \
   --set featureGates=AdditionalCertificateOutputFormats=true
```
## Acme Protocol Certificate

### ProxySQL & PgBouncer
In this release we  introduced TLS support using ACME Protocol based Certificates with Let’s Encrypt. 
Now you can provision KubeDB PgBouncer with TLS secured connection via both CA and ACME issuer using cert-manager. Here's an example using acme issuer with ProxySQL

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: le-issuer
  namespace: demo
spec:
  acme:
    # server: https://acme-v02.api.letsencrypt.org/directory
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: <your_email>
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: <secret_ref_name>
    # ACME DNS-01 provider configurations
    solvers:
    # An empty 'selector' means that this solver matches all domains
    - selector: {}
      dns01:
        cloudflare:
          email: <your_email>
          apiTokenSecretRef:
            name: <secret_ref_name>
            key: api-token

---
apiVersion: v1
kind: Secret
metadata:
  name: <secret_name>
  namespace: demo
type: Opaque
stringData:
  api-token: "sometoken"
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 1
  mode: GroupReplication
  backend:
    name: mysql-server
  syncUsers: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: le-issuer
    certificates:
    - alias: server
      dnsNames:
      - <challenge_resolver_dns>
  terminationPolicy: Delete
  healthChecker:
    failureThreshold: 3
```

# PostgreSQL

Now you will be able to migrate from StandAlone to High Availability Cluster for  PostgreSQL. The migration process will copy the data for new replicas first and thus preventing possible data loss during the migration process.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: move-to-high-availibility
  namespace: demo
spec:
  type: HorizontalScaling
  horizontalScaling:
    replicas: 3
    streamingMode: Synchronous
    standbyMode: Hot
  databaseRef:
    name: demo-pg
```

## Elasticsearch
Latest versions for  OpenSearch  `2.0.1` and `2.5.0` added in this release. Now, you can provision and manage `OpenSearch V2` with `Elasticsearch` CRD using KubeDB.
Check out all the currently supported Opensearch Versions using the following command.
```bash
$ kubectl get esversion | grep opensearch
opensearch-1.1.0            1.1.0     OpenSearch     opensearchproject/opensearch:1.1.0             29h
opensearch-1.2.2            1.2.2     OpenSearch     opensearchproject/opensearch:1.2.2             29h
opensearch-1.3.2            1.3.2     OpenSearch     opensearchproject/opensearch:1.3.2             29h
opensearch-2.0.1            2.0.1     OpenSearch     opensearchproject/opensearch:2.0.1             29h
opensearch-2.5.0            2.5.0     OpenSearch     opensearchproject/opensearch:2.5.0             29h
```

If you are using any of the Opensearch 1.x.x versions, you can very easily upgrade to `OpenSearch 2.x.x` with data persistence using the following `ElasticsearchOpsRequest` CRD.


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: os-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: os-cluster       # refer the name of you opensearch instance
  upgrade:
    targetVersion: opensearch-2.0.1    # refer the target version to be upgraded 
```
We recommend upgrading to OpenSearch `2.0.1` before upgrading to any other V2 minor versions.

## Elasticsearch Dashboard

Along with support for Opensearch V2, KubeDB also brings support for Opensearch Dashboards version `2.0.1` & `2.5.0` in this release. Visualize Opensearch data easily and conveniently by provisioning Opensearch-Dashboards in your cluster using KubeDB. Use the following YAML to deploy TLS secured Opensearch-Dashboards with `ElasticsearchDashboard` CRD.
```yaml
apiVersion: dashboard.kubedb.com/v1alpha1
kind: ElasticsearchDashboard
metadata:
  name: os-cluster-dashboard
  namespace: demo
spec:
  enableSSL: true
  databaseRef:
    name: os-cluster      # refer the name of you opensearch instance
  terminationPolicy: WipeOut
```
On upgrading the Opensearch version using `ElasticsearchVersion` CRD, KubeDB ops-manager autonomously upgrades Opensearch-Dashboards to their compatible versions.

## MySQL 
Latest version for MySQL `8.0.31` and `5.7.41` also added in this release. Here's a example instance that usage MySQL Group Replication
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-group
  namespace: demo
spec:   
  version: "8.0.32"
  replicas: 3
  topology:
   mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  terminationPolicy: Delete
```

## Redis
Latest versions for Redis  `6.0.18`, `6.2.11` and `7.0.9` also added  in this release. Example of a  Redis Standalone instance with version `Redis 6.0.18`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: 6.0.18
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

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2023.02.28/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2023.02.28/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
