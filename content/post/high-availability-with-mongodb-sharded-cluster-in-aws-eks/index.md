---
title: High Availability with MongoDB Sharded Cluster in Amazon Elastic Kubernetes Service (Amazon EKS) using KubeDB
date: "2023-02-15"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- database
- high-availability
- kubedb
- kubernetes
- mongo
- mongodb
- mongodb-database
- mongodb-shard
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy and manage MongoDB sharded cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB Sharded Cluster
3) Horizontal Scaling of MongoDB Sharded Cluster
4) Vertical Scaling of MongoDB Sharded Cluster

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
fc435a61-c74b-9243-83a5-f1110ef2462c
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![License Server](AppscodeLicense.png)

### Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2023.01.31  	v2023.01.31	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.16.0      	v0.16.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.01.31  	v2023.01.31	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.01.31  	v2023.01.31	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.7.0       	v0.7.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.01.31  	v2023.01.31	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.01.31  	v2023.01.31	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.18.0      	v0.18.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.01.31  	v2023.01.31	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.31.0      	v0.31.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.7.0       	v0.7.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.26     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.7.0       	v0.7.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.01.31 \
  --namespace kubedb --create-namespace \
  --set kubedb-provisioner.enabled=true \
  --set kubedb-ops-manager.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-dashboard.enabled=true \
  --set kubedb-schema-manager.enabled=true \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-578b597fd9-4696c       1/1     Running   0          5m48s
kubedb      kubedb-kubedb-dashboard-54cc8997c9-26tzk        1/1     Running   0          5m48s
kubedb      kubedb-kubedb-ops-manager-7f497bd5bb-2qrnr      1/1     Running   0          5m48s
kubedb      kubedb-kubedb-provisioner-85875fc459-wmldn      1/1     Running   0          5m48s
kubedb      kubedb-kubedb-schema-manager-69c7d849d4-86jnk   1/1     Running   0          5m48s
kubedb      kubedb-kubedb-webhook-server-6988b8ccf7-7gdlh   1/1     Running   0          5m48s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-02-14T08:56:33Z
elasticsearchdashboards.dashboard.kubedb.com      2023-02-14T08:56:32Z
elasticsearches.kubedb.com                        2023-02-14T08:56:32Z
elasticsearchopsrequests.ops.kubedb.com           2023-02-14T08:56:41Z
elasticsearchversions.catalog.kubedb.com          2023-02-14T08:50:00Z
etcds.kubedb.com                                  2023-02-14T08:56:41Z
etcdversions.catalog.kubedb.com                   2023-02-14T08:50:00Z
kafkas.kubedb.com                                 2023-02-14T08:56:51Z
kafkaversions.catalog.kubedb.com                  2023-02-14T08:50:01Z
mariadbautoscalers.autoscaling.kubedb.com         2023-02-14T08:56:33Z
mariadbdatabases.schema.kubedb.com                2023-02-14T08:56:37Z
mariadbopsrequests.ops.kubedb.com                 2023-02-14T08:57:01Z
mariadbs.kubedb.com                               2023-02-14T08:56:37Z
mariadbversions.catalog.kubedb.com                2023-02-14T08:50:02Z
memcacheds.kubedb.com                             2023-02-14T08:56:43Z
memcachedversions.catalog.kubedb.com              2023-02-14T08:50:03Z
mongodbautoscalers.autoscaling.kubedb.com         2023-02-14T08:56:34Z
mongodbdatabases.schema.kubedb.com                2023-02-14T08:56:34Z
mongodbopsrequests.ops.kubedb.com                 2023-02-14T08:56:45Z
mongodbs.kubedb.com                               2023-02-14T08:56:35Z
mongodbversions.catalog.kubedb.com                2023-02-14T08:50:04Z
mysqlautoscalers.autoscaling.kubedb.com           2023-02-14T08:56:34Z
mysqldatabases.schema.kubedb.com                  2023-02-14T08:56:33Z
mysqlopsrequests.ops.kubedb.com                   2023-02-14T08:56:57Z
mysqls.kubedb.com                                 2023-02-14T08:56:34Z
mysqlversions.catalog.kubedb.com                  2023-02-14T08:50:05Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-02-14T08:56:34Z
perconaxtradbopsrequests.ops.kubedb.com           2023-02-14T08:57:16Z
perconaxtradbs.kubedb.com                         2023-02-14T08:56:49Z
perconaxtradbversions.catalog.kubedb.com          2023-02-14T08:50:06Z
pgbouncers.kubedb.com                             2023-02-14T08:56:49Z
pgbouncerversions.catalog.kubedb.com              2023-02-14T08:50:07Z
postgresautoscalers.autoscaling.kubedb.com        2023-02-14T08:56:34Z
postgresdatabases.schema.kubedb.com               2023-02-14T08:56:36Z
postgreses.kubedb.com                             2023-02-14T08:56:36Z
postgresopsrequests.ops.kubedb.com                2023-02-14T08:57:08Z
postgresversions.catalog.kubedb.com               2023-02-14T08:50:08Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-02-14T08:56:34Z
proxysqlopsrequests.ops.kubedb.com                2023-02-14T08:57:12Z
proxysqls.kubedb.com                              2023-02-14T08:56:50Z
proxysqlversions.catalog.kubedb.com               2023-02-14T08:50:09Z
publishers.postgres.kubedb.com                    2023-02-14T08:57:26Z
redisautoscalers.autoscaling.kubedb.com           2023-02-14T08:56:34Z
redises.kubedb.com                                2023-02-14T08:56:50Z
redisopsrequests.ops.kubedb.com                   2023-02-14T08:57:04Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-02-14T08:56:34Z
redissentinelopsrequests.ops.kubedb.com           2023-02-14T08:57:19Z
redissentinels.kubedb.com                         2023-02-14T08:56:51Z
redisversions.catalog.kubedb.com                  2023-02-14T08:50:10Z
subscribers.postgres.kubedb.com                   2023-02-14T08:57:30Z
```

## Deploy MongoDB Sharded Cluster

We are going to Deploy MongoDB Sharded Cluster by using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MongoDB CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb-shard
  namespace: demo
spec:
  version: 5.0.3
  shardTopology:
    configServer:
      replicas: 3
      storage:
        resources:
          requests:
            storage: 512Mi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 3
      shards: 2
      storage:
        resources:
          requests:
            storage: 512Mi
        storageClassName: standard
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `mongodb-shard.yaml` 
Then create the above MongoDB CRO

```bash
$ kubectl apply -f mongodb-shard.yaml
mongodb.kubedb.com/mongodb-shard created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `version 5.0.3`. You can list the KubeDB supported versions of MongoDB by running `$ kubectl get mongodbversions` command.
* `spec.shardTopology` represents the topology configuration for sharding.
* `spec.shardTopology.configServer` defines configuration for ConfigServer component of mongodb.
* `spec.shardTopology.configServer.replicas` represents number of replicas for configServer replicaset.
* `spec.shardTopology.mongos` defines configuration for Mongos component of mongodb. Mongos instances run as stateless components (deployment).
* `spec.shardTopology.mongos.replicas` specifies number of replicas of Mongos instance. Here, Mongos is not deployed as replicaset.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy) .

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                            READY   STATUS    RESTARTS   AGE
pod/mongodb-shard-configsvr-0   1/1     Running   0          6m
pod/mongodb-shard-configsvr-1   1/1     Running   0          6m
pod/mongodb-shard-configsvr-2   1/1     Running   0          6m
pod/mongodb-shard-mongos-0      1/1     Running   0          6m
pod/mongodb-shard-mongos-1      1/1     Running   0          6m
pod/mongodb-shard-shard0-0      1/1     Running   0          6m
pod/mongodb-shard-shard0-1      1/1     Running   0          6m
pod/mongodb-shard-shard0-2      1/1     Running   0          6m
pod/mongodb-shard-shard1-0      1/1     Running   0          6m
pod/mongodb-shard-shard1-1      1/1     Running   0          6m
pod/mongodb-shard-shard1-2      1/1     Running   0          6m

NAME                                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/mongodb-shard                  ClusterIP   10.96.245.105   <none>        27017/TCP   6m
service/mongodb-shard-configsvr-pods   ClusterIP   None            <none>        27017/TCP   6m
service/mongodb-shard-mongos-pods      ClusterIP   None            <none>        27017/TCP   6m
service/mongodb-shard-shard0-pods      ClusterIP   None            <none>        27017/TCP   6m
service/mongodb-shard-shard1-pods      ClusterIP   None            <none>        27017/TCP   6m

NAME                                       READY   AGE
statefulset.apps/mongodb-shard-configsvr   3/3     6m
statefulset.apps/mongodb-shard-mongos      2/2     6m
statefulset.apps/mongodb-shard-shard0      3/3     6m
statefulset.apps/mongodb-shard-shard1      3/3     6m

NAME                                               TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb-shard   kubedb.com/mongodb   5.0.3     6m

NAME                               VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb-shard   5.0.3     Ready    6m
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb-shard
NAME            VERSION   STATUS   AGE
mongodb-shard   5.0.3     Ready    6m
```
> We have successfully deployed MongoDB shard in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `mongodb-shard` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb-shard
NAME                 TYPE                       DATA   AGE
mongodb-shard-auth   kubernetes.io/basic-auth   2      7m
mongodb-shard-key    Opaque                     1      7m

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb-shard
NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mongodb-shard                  ClusterIP   10.96.245.105   <none>        27017/TCP   7m
mongodb-shard-configsvr-pods   ClusterIP   None            <none>        27017/TCP   7m
mongodb-shard-mongos-pods      ClusterIP   None            <none>        27017/TCP   7m
mongodb-shard-shard0-pods      ClusterIP   None            <none>        27017/TCP   7m
mongodb-shard-shard1-pods      ClusterIP   None            <none>        27017/TCP   7m
```
Now, we are going to use `mongodb-shard-auth` to export credentials.
Let’s export the `USER` and `PASSWORD` as environment variables to make further commands re-usable.

```bash
$ export USER=$(kubectl get secrets -n demo mongodb-shard-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo mongodb-shard-auth -o jsonpath='{.data.\password}' | base64 -d)
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB shard pod and insert some sample data. 

```bash
$  kubectl exec -it -n demo mongodb-shard-shard0-1 -- mongo admin -u $USER -p $PASSWORD
Defaulted container "mongodb" out of: mongodb, copy-config (init)
MongoDB shell version v5.0.3
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("2b11867d-9d5e-4aea-86c7-94f1fce63a16") }
MongoDB server version: 5.0.3

shard0:PRIMARY> show dbs
admin          0.000GB
config         0.001GB
kubedb-system  0.000GB
local          0.001GB

shard0:PRIMARY> use musicdb
switched to db musicdb

shard0:PRIMARY> db.songs.insert({"name":"Five Hundred Miles"});
WriteResult({ "nInserted" : 1 })

shard0:PRIMARY> db.songs.find().pretty()
{
	"_id" : ObjectId("63ec741ae6d320dafd14e938"),
	"name" : "Five Hundred Miles"
}

shard0:PRIMARY> exit
bye

```

> We've successfully inserted some sample data to our database. More information about Run & Manage MongoDB on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)


## Horizontal Scaling of MongoDB Sharded Cluster

### Horizontal Scale Up

Here, we are going to scale up the number of MongoDB shard and also their replicas to meet the desired number of replicas.
Before applying Horizontal Scaling, let's check the current number of MongoDB shard and their replicas,

```bash
$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.shards'
2

$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.replicas'
3
```

### Create MongoDBOpsRequest

In order to scale up, we have to create a `MongoDBOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mongodb-shard
  horizontalScaling:
    shard: 
      shards: 3
      replicas: 4
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mongodb-shard` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.shard.shards` specifies the desired number of shards after scaling.
- `spec.horizontalScaling.shard.replicas` specifies the desired number of shard replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   2m52s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of shard and their replicas,

```bash
$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.shards'
3

$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.replicas'
4
```

From all the above outputs we can see that the number of shards is now increased to 3 and also, their replicas increased to 4. That means we have successfully scaled up the number of shards and their replicas.

### Horizontal Scale Down

Now, we are going to scale down the number of MongoDB shard and also their replicas to meet the desired number of replicas.

#### Create MongoDBOpsRequest

In order to scale down, again we need to create a `MongoDBOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mongodb-shard
  horizontalScaling:
    shard: 
      shards: 2
      replicas: 3
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mongodb-shard` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.shard.shards` specifies the desired number of shards after scaling.
- `spec.horizontalScaling.shard.replicas` specifies the desired number of shard replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
mongodbopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   2m52s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of shard and their replicas,

```bash
$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.shards'
2

$ kubectl get mongodb -n demo mongodb-shard -o json | jq '.spec.shardTopology.shard.replicas'
3
```
From all the above outputs we can see that the number of shards is now decreased to 2 and also, their replicas decreased to 3. That means we have successfully scaled down the number of shards and their replicas.


## Vetical Scaling of MongoDB Sharded Cluster

We are going to scale up the current cpu resource of the MongoDB sharded cluster by applying Vertical Scaling.
Before applying it, let's check the current resources,

```bash
$ kubectl get pod -n demo mongodb-shard-shard0-0 -o json | jq '.spec.containers[].resources' 
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```
### Vertical Scale Up

#### Create MongoDBOpsRequest

In order to update the resources of the cluster, we have to create a `MongoDBOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vertical-scale-up
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mongodb-shard
  verticalScaling:
    shard:
      requests:
        memory: "1100Mi"
        cpu: "0.55"
      limits:
        memory: "1100Mi"
        cpu: "0.55"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-shard` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.shard` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS       AGE
vertical-scale-up   VerticalScaling   Successful   4m33s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mongodb-shard-shard0-0 -o json | jq '.spec.containers[].resources' 
{
  "limits": {
    "cpu": "550m",
    "memory": "1100Mi"
  },
  "requests": {
    "cpu": "550m",
    "memory": "1100Mi"
  }
}
```
> The above output verifies that we have successfully scaled up the resources of the MongoDB sharded cluster.

### Vertical Scale Down

#### Create MongoDBOpsRequest

In order to update the resources of the database, we have to create a `MongoDBOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vertical-scale-down
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mongodb-shard
  verticalScaling:
    shard:
      requests:
        memory: "1Gi"
        cpu: "0.5"
      limits:
        memory: "1Gi"
        cpu: "0.5"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-shard` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.shard` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   3m
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mongodb-shard-shard0-0 -o json | jq '.spec.containers[].resources' 
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```
> The above output verifies that we have successfully scaled down the resources of the MongoDB sharded cluster.

We have made an in depth tutorial on Configure MongoDB Hidden Node on Kubernetes using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/4sVig7wJzug" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MongoDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
