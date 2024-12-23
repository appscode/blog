---
title: Deploy ZooKeeper in Google Kubernetes Engine (GKE) Using KubeDB
date: "2024-09-18"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- gcp
- gke
- kubedb
- kubernetes
- zookeeper
- zookeeper-ensemble
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Solr, Microsoft SQL Server, Druid, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy ZooKeeper in Google Kubernetes Engine (GKE) Using KubeDB. We will cover the following steps:

1) Install KubeDB
2) Deploy ZooKeeper
3) Access/Create Sample Node

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
appscode/kubedb                   	v2024.8.21   	v2024.8.21 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.32.0      	v0.32.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.8.21   	v2024.8.21 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.2.0       	v0.2.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.8.21   	v2024.8.21 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.23.0      	v0.23.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.8.21   	v2024.8.21 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.8.21   	v2024.8.21 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.8.21   	v2024.8.21 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.34.0      	v0.34.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.8.21   	v2024.8.21 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.47.0      	v0.47.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.23.0      	v0.23.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.8.21   	0.7.6      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.8.21   	v2024.8.21 	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.23.0      	v0.23.1    	KubeDB Webhook Server by AppsCode 

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.8.21 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.ZooKeeper=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-59f9df467d-krnx6       1/1     Running   0          2m26s
kubedb      kubedb-kubedb-ops-manager-7c87bf5688-9592l      1/1     Running   0          2m26s
kubedb      kubedb-kubedb-provisioner-84c84f6f45-4bt4c      1/1     Running   0          2m26s
kubedb      kubedb-kubedb-webhook-server-6fb75fb77b-pb6jf   1/1     Running   0          2m26s
kubedb      kubedb-petset-operator-77b6b9897f-7pvsh         1/1     Running   0          2m26s
kubedb      kubedb-petset-webhook-server-7b756866b7-2j6s2   2/2     Running   0          2m26s
kubedb      kubedb-sidekick-c898cff4c-4kz5w                 1/1     Running   0          2m26s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-09-18T09:02:21Z
connectclusters.kafka.kubedb.com                   2024-09-18T09:02:55Z
connectors.kafka.kubedb.com                        2024-09-18T09:02:55Z
druidversions.catalog.kubedb.com                   2024-09-18T09:02:21Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-09-18T09:02:52Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-09-18T09:02:52Z
elasticsearches.kubedb.com                         2024-09-18T09:02:51Z
elasticsearchopsrequests.ops.kubedb.com            2024-09-18T09:02:52Z
elasticsearchversions.catalog.kubedb.com           2024-09-18T09:02:21Z
etcdversions.catalog.kubedb.com                    2024-09-18T09:02:22Z
ferretdbversions.catalog.kubedb.com                2024-09-18T09:02:22Z
kafkaautoscalers.autoscaling.kubedb.com            2024-09-18T09:02:55Z
kafkaconnectorversions.catalog.kubedb.com          2024-09-18T09:02:22Z
kafkaopsrequests.ops.kubedb.com                    2024-09-18T09:02:55Z
kafkas.kubedb.com                                  2024-09-18T09:02:55Z
kafkaversions.catalog.kubedb.com                   2024-09-18T09:02:22Z
mariadbarchivers.archiver.kubedb.com               2024-09-18T09:02:58Z
mariadbautoscalers.autoscaling.kubedb.com          2024-09-18T09:02:58Z
mariadbdatabases.schema.kubedb.com                 2024-09-18T09:02:58Z
mariadbopsrequests.ops.kubedb.com                  2024-09-18T09:02:58Z
mariadbs.kubedb.com                                2024-09-18T09:02:58Z
mariadbversions.catalog.kubedb.com                 2024-09-18T09:02:22Z
memcachedversions.catalog.kubedb.com               2024-09-18T09:02:22Z
mongodbarchivers.archiver.kubedb.com               2024-09-18T09:03:02Z
mongodbautoscalers.autoscaling.kubedb.com          2024-09-18T09:03:02Z
mongodbdatabases.schema.kubedb.com                 2024-09-18T09:03:02Z
mongodbopsrequests.ops.kubedb.com                  2024-09-18T09:03:02Z
mongodbs.kubedb.com                                2024-09-18T09:03:02Z
mongodbversions.catalog.kubedb.com                 2024-09-18T09:02:22Z
mssqlserverversions.catalog.kubedb.com             2024-09-18T09:02:22Z
mysqlarchivers.archiver.kubedb.com                 2024-09-18T09:03:06Z
mysqlautoscalers.autoscaling.kubedb.com            2024-09-18T09:03:06Z
mysqldatabases.schema.kubedb.com                   2024-09-18T09:03:06Z
mysqlopsrequests.ops.kubedb.com                    2024-09-18T09:03:06Z
mysqls.kubedb.com                                  2024-09-18T09:03:06Z
mysqlversions.catalog.kubedb.com                   2024-09-18T09:02:22Z
perconaxtradbversions.catalog.kubedb.com           2024-09-18T09:02:22Z
pgbouncerversions.catalog.kubedb.com               2024-09-18T09:02:22Z
pgpoolversions.catalog.kubedb.com                  2024-09-18T09:02:22Z
postgresarchivers.archiver.kubedb.com              2024-09-18T09:03:09Z
postgresautoscalers.autoscaling.kubedb.com         2024-09-18T09:03:09Z
postgresdatabases.schema.kubedb.com                2024-09-18T09:03:09Z
postgreses.kubedb.com                              2024-09-18T09:03:09Z
postgresopsrequests.ops.kubedb.com                 2024-09-18T09:03:09Z
postgresversions.catalog.kubedb.com                2024-09-18T09:02:22Z
proxysqlversions.catalog.kubedb.com                2024-09-18T09:02:22Z
publishers.postgres.kubedb.com                     2024-09-18T09:03:09Z
rabbitmqversions.catalog.kubedb.com                2024-09-18T09:02:22Z
redisautoscalers.autoscaling.kubedb.com            2024-09-18T09:03:12Z
redises.kubedb.com                                 2024-09-18T09:03:12Z
redisopsrequests.ops.kubedb.com                    2024-09-18T09:03:12Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-09-18T09:03:13Z
redissentinelopsrequests.ops.kubedb.com            2024-09-18T09:03:13Z
redissentinels.kubedb.com                          2024-09-18T09:03:12Z
redisversions.catalog.kubedb.com                   2024-09-18T09:02:22Z
restproxies.kafka.kubedb.com                       2024-09-18T09:02:55Z
schemaregistries.kafka.kubedb.com                  2024-09-18T09:02:55Z
schemaregistryversions.catalog.kubedb.com          2024-09-18T09:02:22Z
singlestoreversions.catalog.kubedb.com             2024-09-18T09:02:22Z
solrversions.catalog.kubedb.com                    2024-09-18T09:02:22Z
subscribers.postgres.kubedb.com                    2024-09-18T09:03:09Z
zookeeperautoscalers.autoscaling.kubedb.com        2024-09-18T09:03:16Z
zookeepers.kubedb.com                              2024-09-18T09:03:16Z
zookeeperversions.catalog.kubedb.com               2024-09-18T09:02:22Z
```

## Deploy ZooKeeper

Now, we are going to Deploy ZooKeeper using KubeDB.
First, let's create a Namespace in which we will deploy the ZooKeeper.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the ZooKeeper CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zookeeper
  namespace: demo
spec:
  version: "3.9.1"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Let's save this yaml configuration into `zookeeper.yaml` 
Then create the above ZooKeeper CRO,

```bash
$ kubectl apply -f zookeeper.yaml 
zookeeper.kubedb.com/zookeeper created
```
In this yaml,
* `spec.version` field specifies the version of ZooKeeper Here, we are using ZooKeeper `version 3.9.1`. You can list the KubeDB supported versions of ZooKeeper by running `$ kubectl get zookeeperversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Once these are handled correctly and the ZooKeeper object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME              READY   STATUS    RESTARTS   AGE
pod/zookeeper-0   1/1     Running   0          105s
pod/zookeeper-1   1/1     Running   0          70s
pod/zookeeper-2   1/1     Running   0          63s

NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/zookeeper                ClusterIP   10.96.76.214   <none>        2181/TCP                     109s
service/zookeeper-admin-server   ClusterIP   10.96.87.216   <none>        8080/TCP                     109s
service/zookeeper-pods           ClusterIP   None           <none>        2181/TCP,2888/TCP,3888/TCP   109s

NAME                                           TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/zookeeper   kubedb.com/zookeeper   3.9.1     109s

NAME                             TYPE                  VERSION   STATUS   AGE
zookeeper.kubedb.com/zookeeper   kubedb.com/v1alpha2   3.9.1     Ready    109s
```
Let’s check if the ZooKeeper is ready to use,

```bash
$ kubectl get zookeeper -n demo zookeeper
NAME        TYPE                  VERSION   STATUS   AGE
zookeeper   kubedb.com/v1alpha2   3.9.1     Ready    2m21s
```
> We have successfully deployed ZooKeeper in GKE.

### Create Sample Node

In this section, we are going to exec into our ZooKeeper pod and create a sample node.

```bash
$ kubectl exec -it -n demo zookeeper-0 -- sh

$ echo ruok | nc localhost 2181
imok

$ zkCli.sh create /product KubeDB
Connecting to localhost:2181
...
Connection Log Messeges
...
Created /product

$ zkCli.sh get /product
Connecting to localhost:2181
...
Connection Log Messeges
...
KubeDB
```

> We've successfully access ZooKeeper and create a sample ZooKeeper node. More information about Deploy & Manage Production-Grade ZooKeeper on Kubernetes can be found in [ZooKeeper Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-zookeeper-on-kubernetes/)


We have made an in depth tutorial on Provision and Manage ZooKeeper Ensemble on Kubernetes using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/9HEeLN0Vwb4?si=UJiKIJmTodaKpeZn" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [ZooKeeper on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-zookeeper-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
