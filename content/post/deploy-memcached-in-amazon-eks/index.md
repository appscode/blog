---
title: Deploy Memcached in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-10-07"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- memcached
- memcached-server
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Solr, Microsoft SQL Server, Druid, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Memcached in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Memcached
3) Read/Write Sample Data

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB.

![License Server](AppscodeLicense.png)

### Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2024.9.30   	v2024.9.30 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.33.0      	v0.33.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.9.30   	v2024.9.30 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.3.0       	v0.3.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.9.30   	v2024.9.30 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.24.0      	v0.24.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.9.30   	v2024.9.30 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.9.30   	v2024.9.30 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.9.30   	v2024.9.30 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.35.0      	v0.35.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.9.30   	v2024.9.30 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.9.30   	v0.10.0    	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.9.30   	v0.10.0    	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.9.30   	v0.10.0    	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.48.0      	v0.48.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.24.0      	v0.24.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.9.30   	0.7.6      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.9.30   	v2024.9.30 	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.24.0      	v0.24.0    	KubeDB Webhook Server by AppsCode 

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.9.30 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.Memcached=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-6984957978-pq5bn       1/1     Running   0          2m17s
kubedb      kubedb-kubedb-ops-manager-68895444cb-b54kh      1/1     Running   0          2m17s
kubedb      kubedb-kubedb-provisioner-ff88d97cf-7kxwh       1/1     Running   0          2m17s
kubedb      kubedb-kubedb-webhook-server-6564f8b66f-tnwnr   1/1     Running   0          2m17s
kubedb      kubedb-petset-operator-6fb9b649cd-qjm4m         1/1     Running   0          2m17s
kubedb      kubedb-petset-webhook-server-7b59bdbdb5-b2g5x   2/2     Running   0          2m17s
kubedb      kubedb-sidekick-ff65dbd76-7ddlg                 1/1     Running   0          2m17s
```


We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
cassandraversions.catalog.kubedb.com               2024-10-07T12:34:10Z
clickhouseversions.catalog.kubedb.com              2024-10-07T12:34:10Z
connectclusters.kafka.kubedb.com                   2024-10-07T12:34:46Z
connectors.kafka.kubedb.com                        2024-10-07T12:34:46Z
druidversions.catalog.kubedb.com                   2024-10-07T12:34:10Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-10-07T12:34:43Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-10-07T12:34:43Z
elasticsearches.kubedb.com                         2024-10-07T12:34:42Z
elasticsearchopsrequests.ops.kubedb.com            2024-10-07T12:34:42Z
elasticsearchversions.catalog.kubedb.com           2024-10-07T12:34:10Z
etcdversions.catalog.kubedb.com                    2024-10-07T12:34:10Z
ferretdbversions.catalog.kubedb.com                2024-10-07T12:34:10Z
kafkaautoscalers.autoscaling.kubedb.com            2024-10-07T12:34:46Z
kafkaconnectorversions.catalog.kubedb.com          2024-10-07T12:34:10Z
kafkaopsrequests.ops.kubedb.com                    2024-10-07T12:34:46Z
kafkas.kubedb.com                                  2024-10-07T12:34:46Z
kafkaversions.catalog.kubedb.com                   2024-10-07T12:34:10Z
mariadbarchivers.archiver.kubedb.com               2024-10-07T12:34:49Z
mariadbautoscalers.autoscaling.kubedb.com          2024-10-07T12:34:49Z
mariadbdatabases.schema.kubedb.com                 2024-10-07T12:34:49Z
mariadbopsrequests.ops.kubedb.com                  2024-10-07T12:34:49Z
mariadbs.kubedb.com                                2024-10-07T12:34:49Z
mariadbversions.catalog.kubedb.com                 2024-10-07T12:34:10Z
memcachedautoscalers.autoscaling.kubedb.com        2024-10-07T12:34:53Z
memcachedopsrequests.ops.kubedb.com                2024-10-07T12:34:53Z
memcacheds.kubedb.com                              2024-10-07T12:34:53Z
memcachedversions.catalog.kubedb.com               2024-10-07T12:34:10Z
mongodbarchivers.archiver.kubedb.com               2024-10-07T12:34:56Z
mongodbautoscalers.autoscaling.kubedb.com          2024-10-07T12:34:56Z
mongodbdatabases.schema.kubedb.com                 2024-10-07T12:34:56Z
mongodbopsrequests.ops.kubedb.com                  2024-10-07T12:34:56Z
mongodbs.kubedb.com                                2024-10-07T12:34:56Z
mongodbversions.catalog.kubedb.com                 2024-10-07T12:34:10Z
mssqlserverversions.catalog.kubedb.com             2024-10-07T12:34:10Z
mysqlarchivers.archiver.kubedb.com                 2024-10-07T12:35:00Z
mysqlautoscalers.autoscaling.kubedb.com            2024-10-07T12:35:00Z
mysqldatabases.schema.kubedb.com                   2024-10-07T12:35:00Z
mysqlopsrequests.ops.kubedb.com                    2024-10-07T12:35:00Z
mysqls.kubedb.com                                  2024-10-07T12:35:00Z
mysqlversions.catalog.kubedb.com                   2024-10-07T12:34:10Z
perconaxtradbversions.catalog.kubedb.com           2024-10-07T12:34:10Z
pgbouncerversions.catalog.kubedb.com               2024-10-07T12:34:10Z
pgpoolversions.catalog.kubedb.com                  2024-10-07T12:34:10Z
postgresarchivers.archiver.kubedb.com              2024-10-07T12:35:03Z
postgresautoscalers.autoscaling.kubedb.com         2024-10-07T12:35:03Z
postgresdatabases.schema.kubedb.com                2024-10-07T12:35:03Z
postgreses.kubedb.com                              2024-10-07T12:35:03Z
postgresopsrequests.ops.kubedb.com                 2024-10-07T12:35:03Z
postgresversions.catalog.kubedb.com                2024-10-07T12:34:10Z
proxysqlversions.catalog.kubedb.com                2024-10-07T12:34:10Z
publishers.postgres.kubedb.com                     2024-10-07T12:35:03Z
rabbitmqversions.catalog.kubedb.com                2024-10-07T12:34:10Z
redisautoscalers.autoscaling.kubedb.com            2024-10-07T12:35:07Z
redises.kubedb.com                                 2024-10-07T12:35:07Z
redisopsrequests.ops.kubedb.com                    2024-10-07T12:35:07Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-10-07T12:35:07Z
redissentinelopsrequests.ops.kubedb.com            2024-10-07T12:35:07Z
redissentinels.kubedb.com                          2024-10-07T12:35:07Z
redisversions.catalog.kubedb.com                   2024-10-07T12:34:10Z
restproxies.kafka.kubedb.com                       2024-10-07T12:34:46Z
schemaregistries.kafka.kubedb.com                  2024-10-07T12:34:46Z
schemaregistryversions.catalog.kubedb.com          2024-10-07T12:34:10Z
singlestoreversions.catalog.kubedb.com             2024-10-07T12:34:10Z
solrversions.catalog.kubedb.com                    2024-10-07T12:34:10Z
subscribers.postgres.kubedb.com                    2024-10-07T12:35:03Z
zookeeperversions.catalog.kubedb.com               2024-10-07T12:34:10Z
```

## Deploy Memcached

Now, we are going to Deploy Memcached using KubeDB.
First, let's create a Namespace in which we will deploy the Memcached.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the Memcached CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: sample-memcached
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
  deletionPolicy: "WipeOut"
```

Let's save this yaml configuration into `sample-memcached.yaml` 
Then create the above Memcached CRO,

```bash
$ kubectl apply -f sample-memcached.yaml 
memcached.kubedb.com/sample-memcached created
```
In this yaml,
* `spec.version` field specifies the version of Memcached Here, we are using Memcached `1.6.22`. You can list the KubeDB supported versions of Memcached by running `$ kubectl get memcachedversions` command.
* `spec.replicas` is an optional field that specifies the number of desired Instances/Replicas of Memcached server.
* `spec.podTemplate` KubeDB allows providing a template for database pod. KubeDB operator will pass the information provided in `spec.podTemplate` to the Petset created for Memcached server.
* `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Once these are handled correctly and the Memcached object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/sample-memcached-0   1/1     Running   0          64s

NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/sample-memcached        ClusterIP   10.96.179.230   <none>        11211/TCP   64s
service/sample-memcached-pods   ClusterIP   None            <none>        11211/TCP   64s

NAME                                                  TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/sample-memcached   kubedb.com/memcached   1.6.22    64s

NAME                                    VERSION   STATUS   AGE
memcached.kubedb.com/sample-memcached   1.6.22    Ready    65s
```
Let’s check if the Memcached is ready to use,

```bash
$ kubectl get memcached -n demo sample-memcached
NAME               VERSION   STATUS   AGE
sample-memcached   1.6.22    Ready    89s
```
> We have successfully deployed Memcached in Amazon EKS.


## Connect with Memcached

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with Memcached. Here, we will use the `sample-memcached-0` pod. First, we need to port-forward the `sample-memcached-0` pod to the port `11211` on the local machine:

```bash
$ kubectl port-forward -n demo sample-memcached-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```
Once port forwarding is established, you can connect to Memcached locally using `telnet`. Start a `telnet` session to connect to the Memcached.

### Insert Sample Data

```bash
$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
```
With the connection active, you can insert some sample data into Memcached. Use the `set` command to store a key-value pair:

```bash
set AppsCode 0 2592000 6
KubeDB
STORED
```
Here, `AppsCode` is the key, `0` is the flags field, `2592000` is the expiration time in seconds, `6` is the length of the data, and `KubeDB` is the value being stored. To verify that the data was successfully stored, retrieve it using the `get` command:

```bash
get AppsCode
VALUE AppsCode 0 6
KubeDB
END

telnet> quit
```

> Congratulations! You've successfully connected to Memcached and inserted sample data. More information about Deploy & Manage Production-Grade Memcached on Kubernetes can be found in [Memcached Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-memcached-on-kubernetes/)


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Memcached on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-memcached-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
