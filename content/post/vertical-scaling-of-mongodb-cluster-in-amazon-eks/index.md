---
title: Vertical Scaling of MongoDB Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-08-05"
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
- mongodb
- mongodb-cluster
- mongodb-database
- vertical-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Vertical Scaling of MongoDB Cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB Cluster
3) Read/Write Sample Data
4) Vertical Scaling of MongoDB Cluster

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the `license.txt` file. For this tutorial we will use KubeDB.

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
appscode/kubedb-ops-manager       	v0.33.0      	v0.33.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.6.4    	v2024.6.4  	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.46.0      	v0.46.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.22.0      	v0.22.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.7.4    	0.7.3      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.7.4    	v2024.7.4  	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.22.0      	v0.22.0    	KubeDB Webhook Server by AppsCode 

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.6.4 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-54f5df974d-b67jf       1/1     Running   0          114s
kubedb      kubedb-kubedb-ops-manager-6b85bb7b68-rlzrb      1/1     Running   0          114s
kubedb      kubedb-kubedb-provisioner-5945f8757-wm696       1/1     Running   0          114s
kubedb      kubedb-kubedb-webhook-server-77fcfc96d8-h7cs6   1/1     Running   0          114s
kubedb      kubedb-petset-operator-77b6b9897f-4kd5x         1/1     Running   0          114s
kubedb      kubedb-petset-webhook-server-5f7f9b5fdc-xqgvs   2/2     Running   0          114s
kubedb      kubedb-sidekick-c898cff4c-9t2bt                 1/1     Running   0          114s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-08-05T12:49:44Z
connectclusters.kafka.kubedb.com                   2024-08-05T12:50:16Z
connectors.kafka.kubedb.com                        2024-08-05T12:50:16Z
druidversions.catalog.kubedb.com                   2024-08-05T12:49:44Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-08-05T12:50:12Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-08-05T12:50:12Z
elasticsearches.kubedb.com                         2024-08-05T12:50:12Z
elasticsearchopsrequests.ops.kubedb.com            2024-08-05T12:50:12Z
elasticsearchversions.catalog.kubedb.com           2024-08-05T12:49:44Z
etcdversions.catalog.kubedb.com                    2024-08-05T12:49:44Z
ferretdbversions.catalog.kubedb.com                2024-08-05T12:49:45Z
kafkaautoscalers.autoscaling.kubedb.com            2024-08-05T12:50:16Z
kafkaconnectorversions.catalog.kubedb.com          2024-08-05T12:49:45Z
kafkaopsrequests.ops.kubedb.com                    2024-08-05T12:50:16Z
kafkas.kubedb.com                                  2024-08-05T12:50:15Z
kafkaversions.catalog.kubedb.com                   2024-08-05T12:49:45Z
mariadbarchivers.archiver.kubedb.com               2024-08-05T12:50:19Z
mariadbautoscalers.autoscaling.kubedb.com          2024-08-05T12:50:19Z
mariadbdatabases.schema.kubedb.com                 2024-08-05T12:50:19Z
mariadbopsrequests.ops.kubedb.com                  2024-08-05T12:50:19Z
mariadbs.kubedb.com                                2024-08-05T12:50:19Z
mariadbversions.catalog.kubedb.com                 2024-08-05T12:49:45Z
memcachedversions.catalog.kubedb.com               2024-08-05T12:49:45Z
mongodbarchivers.archiver.kubedb.com               2024-08-05T12:50:23Z
mongodbautoscalers.autoscaling.kubedb.com          2024-08-05T12:50:23Z
mongodbdatabases.schema.kubedb.com                 2024-08-05T12:50:23Z
mongodbopsrequests.ops.kubedb.com                  2024-08-05T12:50:23Z
mongodbs.kubedb.com                                2024-08-05T12:50:22Z
mongodbversions.catalog.kubedb.com                 2024-08-05T12:49:45Z
mssqlserverversions.catalog.kubedb.com             2024-08-05T12:49:45Z
mysqlarchivers.archiver.kubedb.com                 2024-08-05T12:50:26Z
mysqlautoscalers.autoscaling.kubedb.com            2024-08-05T12:50:26Z
mysqldatabases.schema.kubedb.com                   2024-08-05T12:50:26Z
mysqlopsrequests.ops.kubedb.com                    2024-08-05T12:50:26Z
mysqls.kubedb.com                                  2024-08-05T12:50:26Z
mysqlversions.catalog.kubedb.com                   2024-08-05T12:49:45Z
perconaxtradbversions.catalog.kubedb.com           2024-08-05T12:49:45Z
pgbouncerversions.catalog.kubedb.com               2024-08-05T12:49:45Z
pgpoolversions.catalog.kubedb.com                  2024-08-05T12:49:45Z
postgresarchivers.archiver.kubedb.com              2024-08-05T12:50:30Z
postgresautoscalers.autoscaling.kubedb.com         2024-08-05T12:50:30Z
postgresdatabases.schema.kubedb.com                2024-08-05T12:50:30Z
postgreses.kubedb.com                              2024-08-05T12:50:30Z
postgresopsrequests.ops.kubedb.com                 2024-08-05T12:50:30Z
postgresversions.catalog.kubedb.com                2024-08-05T12:49:45Z
proxysqlversions.catalog.kubedb.com                2024-08-05T12:49:45Z
publishers.postgres.kubedb.com                     2024-08-05T12:50:30Z
rabbitmqversions.catalog.kubedb.com                2024-08-05T12:49:45Z
redisautoscalers.autoscaling.kubedb.com            2024-08-05T12:50:33Z
redises.kubedb.com                                 2024-08-05T12:50:33Z
redisopsrequests.ops.kubedb.com                    2024-08-05T12:50:33Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-08-05T12:50:33Z
redissentinelopsrequests.ops.kubedb.com            2024-08-05T12:50:33Z
redissentinels.kubedb.com                          2024-08-05T12:50:33Z
redisversions.catalog.kubedb.com                   2024-08-05T12:49:45Z
schemaregistries.kafka.kubedb.com                  2024-08-05T12:50:16Z
schemaregistryversions.catalog.kubedb.com          2024-08-05T12:49:45Z
singlestoreversions.catalog.kubedb.com             2024-08-05T12:49:45Z
solrversions.catalog.kubedb.com                    2024-08-05T12:49:45Z
subscribers.postgres.kubedb.com                    2024-08-05T12:50:30Z
zookeeperversions.catalog.kubedb.com               2024-08-05T12:49:45Z
```

## Deploy MongoDB Cluster

We are going to Deploy MongoDB Cluster using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MongoDB CR we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb-rs
  namespace: demo
spec:
  version: "7.0.8"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "gp2"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `mongodb-rs.yaml` 
Then create the above MongoDB CR,

```bash
$ kubectl apply -f mongodb-rs.yaml
mongodb.kubedb.com/mongodb-rs created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `7.0.8`. You can list the KubeDB supported versions of MongoDB by running `$ kubectl get mongodbversions` command.
* `spec.replicas` denotes the number of members in `rs0` mongodb replicaset.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy).

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME               READY   STATUS    RESTARTS   AGE
pod/mongodb-rs-0   2/2     Running   0          2m44s
pod/mongodb-rs-1   2/2     Running   0          78s
pod/mongodb-rs-2   2/2     Running   0          53s

NAME                      TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
service/mongodb-rs        ClusterIP   10.96.14.44   <none>        27017/TCP   2m49s
service/mongodb-rs-pods   ClusterIP   None          <none>        27017/TCP   2m49s

NAME                          READY   AGE
statefulset.apps/mongodb-rs   3/3     2m44s

NAME                                            TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb-rs   kubedb.com/mongodb   7.0.8     2m39s

NAME                            VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb-rs   7.0.8     Ready    2m49s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb-rs
NAME         VERSION   STATUS   AGE
mongodb-rs   7.0.8     Ready    3m5s
```
> We have successfully deployed MongoDB cluster in Amazon EKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database mongodb that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE                       DATA   AGE
mongodb-rs-auth   kubernetes.io/basic-auth   2      3m22s
mongodb-rs-key    Opaque                     1      3m22s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
mongodb-rs        ClusterIP   10.96.14.44   <none>        27017/TCP   3m37s
mongodb-rs-pods   ClusterIP   None          <none>        27017/TCP   3m37s
```

Now, we are going to use `mongodb-rs-auth` to export credentials.

```bash
$ kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\password}' | base64 -d
8Z.hA0QpzPzRB3jb
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB pod and insert some sample data.

```bash
$ kubectl exec -it mongodb-rs-0 -n demo bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mongodb-rs-0:/$ mongosh admin -u root -p '8Z.hA0QpzPzRB3jb'

Using MongoDB:		7.0.8
Using Mongosh:		2.2.5

For mongosh info see: https://docs.mongodb.com/mongodb-shell/


rs0 [direct: primary] admin> show dbs
admin          172.00 KiB
config         176.00 KiB
kubedb-system   40.00 KiB
local          404.00 KiB

rs0 [direct: primary] admin> use musicdb
switched to db musicdb

rs0 [direct: primary] musicdb> db.songs.insert({"name":"Annie's Song"});
{
  acknowledged: true,
  insertedIds: { '0': ObjectId('66ba074675cd9977362202d8') }
}

rs0 [direct: primary] musicdb> db.songs.find().pretty()
[ { _id: ObjectId('66ba074675cd9977362202d8'), name: "Annie's Song" } ]

rs0 [direct: primary] musicdb> exit
```

> We've successfully inserted some sample data to our database. More information about Deploy & Manage MongoDB on Kubernetes can be found in [Kubernetes MongoDB](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)


## Vertical Scaling of MongoDB Cluster

### Vertical Scale Up

Here, we are going to scale up the resources of the MongoDB cluster. Before applying Vertical Scaling, let’s check the current resources,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "800m",
    "memory": "1Gi"
  }
}
{}
```

### Create MongoDBOpsRequest

In order to scale up, we have to create a `MongoDBOpsRequest`. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vertical-scale-up
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mongodb-rs
  verticalScaling:
    replicaSet:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "1"
        limits:
          memory: "1.2Gi"
          cpu: "1"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling` specifies the expected mongodb container resources after scaling.

Let’s save this yaml configuration into `vertical-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS       AGE
vertical-scale-up   VerticalScaling   Successful   3m25s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the current resources,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "1",
    "memory": "1288490188800m"
  }
}
{}
```

> From all the above outputs we can see that the resources of the cluster is now increased. That means we have successfully scaled up the resources of the MongoDB cluster.

### Vertical Scale Down

Now, we are going to scale down the resources of the cluster.


#### Create MongoDBOpsRequest

In order to scale down, again we need to create a new `MongoDBOpsRequest`. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vertical-scale-down
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mongodb-rs
  verticalScaling:
    replicaSet:
      resources:
        requests:
          memory: "1Gi"
          cpu: "0.8"
        limits:
          memory: "1Gi"
          cpu: "0.8"
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling` specifies the expected mongodb container resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   2m30s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the resources,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "800m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "800m",
    "memory": "1Gi"
  }
}
{}
```
> From all the above outputs we can see that the resources of the cluster is decreased. That means we have successfully scaled down the resources of the MongoDB cluster.

If you want to learn more about Production-Grade MongoDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=f0W-h9paKMX9042P&amp;list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MongoDB on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
