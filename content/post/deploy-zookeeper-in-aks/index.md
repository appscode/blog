---
title: Deploy ZooKeeper in Azure Kubernetes Service (AKS) Using KubeDB
date: "2024-07-10"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- microsoft-azure
- zookeeper
- zookeeper-ensemble
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy ZooKeeper in Azure Kubernetes Service (AKS) Using KubeDB. We will cover the following steps:

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
appscode/kubedb                   	v2024.6.4    	v2024.6.4  	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.31.0      	v0.31.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.6.4    	v2024.6.4  	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.1.0       	v0.1.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.6.4    	v2024.6.4  	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.22.0      	v0.22.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.6.4    	v2024.6.4  	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.6.4    	v2024.6.4  	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.6.4    	v2024.6.4  	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.33.0      	v0.33.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.6.4    	v2024.6.4  	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.46.0      	v0.46.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.22.0      	v0.22.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.7.4    	0.7.2      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.7.4    	v2024.7.4  	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.22.0      	v0.22.0    	KubeDB Webhook Server by AppsCode 

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.6.4 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.ZooKeeper=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-8455db7d97-8v22z       1/1     Running   0          4m17s
kubedb      kubedb-kubedb-ops-manager-5fc4c4b6f4-kbn5f      1/1     Running   0          4m17s
kubedb      kubedb-kubedb-provisioner-fdff7f57b-25xx5       1/1     Running   0          4m17s
kubedb      kubedb-kubedb-webhook-server-59688bf84c-x8sst   1/1     Running   0          4m17s
kubedb      kubedb-petset-operator-54877fd499-xrqgt         1/1     Running   0          4m17s
kubedb      kubedb-petset-webhook-server-55b4747ffd-dnqtv   2/2     Running   0          4m17s
kubedb      kubedb-sidekick-5d9947bd9-p6wc7                 1/1     Running   0          4m17s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-07-10T05:36:02Z
connectclusters.kafka.kubedb.com                   2024-07-10T05:36:49Z
connectors.kafka.kubedb.com                        2024-07-10T05:36:49Z
druidversions.catalog.kubedb.com                   2024-07-10T05:36:02Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-07-10T05:36:46Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-07-10T05:36:46Z
elasticsearches.kubedb.com                         2024-07-10T05:36:46Z
elasticsearchopsrequests.ops.kubedb.com            2024-07-10T05:36:46Z
elasticsearchversions.catalog.kubedb.com           2024-07-10T05:36:02Z
etcdversions.catalog.kubedb.com                    2024-07-10T05:36:02Z
ferretdbversions.catalog.kubedb.com                2024-07-10T05:36:02Z
kafkaautoscalers.autoscaling.kubedb.com            2024-07-10T05:36:49Z
kafkaconnectorversions.catalog.kubedb.com          2024-07-10T05:36:02Z
kafkaopsrequests.ops.kubedb.com                    2024-07-10T05:36:49Z
kafkas.kubedb.com                                  2024-07-10T05:36:49Z
kafkaversions.catalog.kubedb.com                   2024-07-10T05:36:03Z
mariadbarchivers.archiver.kubedb.com               2024-07-10T05:36:52Z
mariadbautoscalers.autoscaling.kubedb.com          2024-07-10T05:36:52Z
mariadbdatabases.schema.kubedb.com                 2024-07-10T05:36:52Z
mariadbopsrequests.ops.kubedb.com                  2024-07-10T05:36:52Z
mariadbs.kubedb.com                                2024-07-10T05:36:52Z
mariadbversions.catalog.kubedb.com                 2024-07-10T05:36:03Z
memcachedversions.catalog.kubedb.com               2024-07-10T05:36:03Z
mongodbarchivers.archiver.kubedb.com               2024-07-10T05:36:56Z
mongodbautoscalers.autoscaling.kubedb.com          2024-07-10T05:36:55Z
mongodbdatabases.schema.kubedb.com                 2024-07-10T05:36:56Z
mongodbopsrequests.ops.kubedb.com                  2024-07-10T05:36:55Z
mongodbs.kubedb.com                                2024-07-10T05:36:55Z
mongodbversions.catalog.kubedb.com                 2024-07-10T05:36:03Z
mssqlserverversions.catalog.kubedb.com             2024-07-10T05:36:03Z
mysqlarchivers.archiver.kubedb.com                 2024-07-10T05:36:59Z
mysqlautoscalers.autoscaling.kubedb.com            2024-07-10T05:36:59Z
mysqldatabases.schema.kubedb.com                   2024-07-10T05:36:59Z
mysqlopsrequests.ops.kubedb.com                    2024-07-10T05:36:59Z
mysqls.kubedb.com                                  2024-07-10T05:36:59Z
mysqlversions.catalog.kubedb.com                   2024-07-10T05:36:03Z
perconaxtradbversions.catalog.kubedb.com           2024-07-10T05:36:03Z
pgbouncerversions.catalog.kubedb.com               2024-07-10T05:36:03Z
pgpoolversions.catalog.kubedb.com                  2024-07-10T05:36:03Z
postgresarchivers.archiver.kubedb.com              2024-07-10T05:37:02Z
postgresautoscalers.autoscaling.kubedb.com         2024-07-10T05:37:02Z
postgresdatabases.schema.kubedb.com                2024-07-10T05:37:02Z
postgreses.kubedb.com                              2024-07-10T05:37:02Z
postgresopsrequests.ops.kubedb.com                 2024-07-10T05:37:02Z
postgresversions.catalog.kubedb.com                2024-07-10T05:36:03Z
proxysqlversions.catalog.kubedb.com                2024-07-10T05:36:03Z
publishers.postgres.kubedb.com                     2024-07-10T05:37:02Z
rabbitmqversions.catalog.kubedb.com                2024-07-10T05:36:03Z
redisautoscalers.autoscaling.kubedb.com            2024-07-10T05:37:06Z
redises.kubedb.com                                 2024-07-10T05:37:05Z
redisopsrequests.ops.kubedb.com                    2024-07-10T05:37:06Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-07-10T05:37:06Z
redissentinelopsrequests.ops.kubedb.com            2024-07-10T05:37:06Z
redissentinels.kubedb.com                          2024-07-10T05:37:06Z
redisversions.catalog.kubedb.com                   2024-07-10T05:36:03Z
schemaregistries.kafka.kubedb.com                  2024-07-10T05:36:49Z
schemaregistryversions.catalog.kubedb.com          2024-07-10T05:36:03Z
singlestoreversions.catalog.kubedb.com             2024-07-10T05:36:03Z
solrversions.catalog.kubedb.com                    2024-07-10T05:36:03Z
subscribers.postgres.kubedb.com                    2024-07-10T05:37:02Z
zookeepers.kubedb.com                              2024-07-10T05:37:09Z
zookeeperversions.catalog.kubedb.com               2024-07-10T05:36:03Z
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
    storageClassName: "default"
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
pod/zookeeper-0   1/1     Running   0          118s
pod/zookeeper-1   1/1     Running   0          74s
pod/zookeeper-2   1/1     Running   0          64s

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/zookeeper                ClusterIP   10.96.229.242   <none>        2181/TCP                     2m3s
service/zookeeper-admin-server   ClusterIP   10.96.131.106   <none>        8080/TCP                     2m3s
service/zookeeper-pods           ClusterIP   None            <none>        2181/TCP,2888/TCP,3888/TCP   2m3s

NAME                                           TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/zookeeper   kubedb.com/zookeeper   3.9.1     2m3s
```
Let’s check if the ZooKeeper is ready to use,

```bash
$ kubectl get zookeeper -n demo zookeeper
NAME        TYPE                  VERSION   STATUS   AGE
zookeeper   kubedb.com/v1alpha2   3.9.1     Ready    2m28s
```
> We have successfully deployed ZooKeeper in AKS. Now we can exec into the container to use the database.

### Create Sample Node

In this section, we are going to exec into our ZooKeeper pod and create a sample node.

```bash
$ kubectl exec -it -n demo zookeeper-0 -- sh

$ echo ruok | nc localhost 2181
imok

$ zkCli.sh create /product kubedb
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
kubedb
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
