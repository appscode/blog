---
title: Deploy Highly Available MongoDB Cluster in Google Kubernetes Engine (GKE) using KubeDB
date: "2023-03-31"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- gcp
- gcs
- gke
- high-availability
- kubedb
- kubernetes
- mongo
- mongodb
- mongodb-database
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Highly Available MongoDB Cluster in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB Cluster
3) Horizontal Scaling of MongoDB Cluster
4) Vertical Scaling of MongoDB Cluster

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
appscode/kubedb                   	v2023.02.28  	v2023.02.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.17.0      	v0.17.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.02.28  	v2023.02.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.02.28  	v2023.02.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.8.0       	v0.8.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.02.28  	v2023.02.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.02.28  	v2023.02.28	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.19.0      	v0.19.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.02.28  	v2023.02.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.32.0      	v0.32.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.8.0       	v0.8.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.8.0       	v0.8.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.02.28 \
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
kubedb      kubedb-kubedb-autoscaler-5d4d9db7b-7tvjx        1/1     Running   0          94s
kubedb      kubedb-kubedb-dashboard-7b848bb7c-lk2s5         1/1     Running   0          94s
kubedb      kubedb-kubedb-ops-manager-86786bfd4f-bcccw      1/1     Running   0          94s
kubedb      kubedb-kubedb-provisioner-7657456794-dmdcs      1/1     Running   0          94s
kubedb      kubedb-kubedb-schema-manager-bb9f7b758-7pz5m    1/1     Running   0          94s
kubedb      kubedb-kubedb-webhook-server-7f5cbfb695-jndmv   1/1     Running   0          94s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-03-31T05:43:55Z
elasticsearchdashboards.dashboard.kubedb.com      2023-03-31T05:43:35Z
elasticsearches.kubedb.com                        2023-03-31T05:43:35Z
elasticsearchopsrequests.ops.kubedb.com           2023-03-31T05:43:40Z
elasticsearchversions.catalog.kubedb.com          2023-03-31T05:40:33Z
etcds.kubedb.com                                  2023-03-31T05:43:47Z
etcdversions.catalog.kubedb.com                   2023-03-31T05:40:33Z
kafkas.kubedb.com                                 2023-03-31T05:44:14Z
kafkaversions.catalog.kubedb.com                  2023-03-31T05:40:34Z
mariadbautoscalers.autoscaling.kubedb.com         2023-03-31T05:43:55Z
mariadbdatabases.schema.kubedb.com                2023-03-31T05:44:02Z
mariadbopsrequests.ops.kubedb.com                 2023-03-31T05:44:13Z
mariadbs.kubedb.com                               2023-03-31T05:43:48Z
mariadbversions.catalog.kubedb.com                2023-03-31T05:40:34Z
memcacheds.kubedb.com                             2023-03-31T05:43:49Z
memcachedversions.catalog.kubedb.com              2023-03-31T05:40:35Z
mongodbautoscalers.autoscaling.kubedb.com         2023-03-31T05:43:55Z
mongodbdatabases.schema.kubedb.com                2023-03-31T05:43:51Z
mongodbopsrequests.ops.kubedb.com                 2023-03-31T05:43:43Z
mongodbs.kubedb.com                               2023-03-31T05:43:43Z
mongodbversions.catalog.kubedb.com                2023-03-31T05:40:35Z
mysqlautoscalers.autoscaling.kubedb.com           2023-03-31T05:43:56Z
mysqldatabases.schema.kubedb.com                  2023-03-31T05:43:49Z
mysqlopsrequests.ops.kubedb.com                   2023-03-31T05:44:01Z
mysqls.kubedb.com                                 2023-03-31T05:43:49Z
mysqlversions.catalog.kubedb.com                  2023-03-31T05:40:36Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-03-31T05:43:56Z
perconaxtradbopsrequests.ops.kubedb.com           2023-03-31T05:44:49Z
perconaxtradbs.kubedb.com                         2023-03-31T05:44:00Z
perconaxtradbversions.catalog.kubedb.com          2023-03-31T05:40:36Z
pgbouncers.kubedb.com                             2023-03-31T05:43:49Z
pgbouncerversions.catalog.kubedb.com              2023-03-31T05:40:36Z
postgresautoscalers.autoscaling.kubedb.com        2023-03-31T05:43:57Z
postgresdatabases.schema.kubedb.com               2023-03-31T05:44:00Z
postgreses.kubedb.com                             2023-03-31T05:44:00Z
postgresopsrequests.ops.kubedb.com                2023-03-31T05:44:37Z
postgresversions.catalog.kubedb.com               2023-03-31T05:40:37Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-03-31T05:43:57Z
proxysqlopsrequests.ops.kubedb.com                2023-03-31T05:44:45Z
proxysqls.kubedb.com                              2023-03-31T05:44:12Z
proxysqlversions.catalog.kubedb.com               2023-03-31T05:40:37Z
publishers.postgres.kubedb.com                    2023-03-31T05:45:15Z
redisautoscalers.autoscaling.kubedb.com           2023-03-31T05:43:58Z
redises.kubedb.com                                2023-03-31T05:44:13Z
redisopsrequests.ops.kubedb.com                   2023-03-31T05:44:18Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-03-31T05:43:58Z
redissentinelopsrequests.ops.kubedb.com           2023-03-31T05:45:07Z
redissentinels.kubedb.com                         2023-03-31T05:44:13Z
redisversions.catalog.kubedb.com                  2023-03-31T05:40:38Z
subscribers.postgres.kubedb.com                   2023-03-31T05:45:19Z
```

## Deploy MongoDB Cluster

We are going to Deploy MongoDB Cluster by using KubeDB.
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
  name: mongodb-rs
  namespace: demo
spec:
  version: "5.0.3"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `mongodb-rs.yaml` 
Then create the above MongoDB CRO

```bash
$ kubectl apply -f mongodb-rs.yaml
mongodb.kubedb.com/mongodb-rs created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `version 5.0.3`. You can list the KubeDB supported versions of MongoDB by running `kubectl get mongodbversions` command.
* `spec.replicas` denotes the number of members in `rs0` mongodb replicaset.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy) .

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME               READY   STATUS    RESTARTS   AGE
pod/mongodb-rs-0   2/2     Running   0          3m40s
pod/mongodb-rs-1   2/2     Running   0          2m50s
pod/mongodb-rs-2   2/2     Running   0          119s

NAME                      TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
service/mongodb-rs        ClusterIP   10.8.1.145   <none>        27017/TCP   3m44s
service/mongodb-rs-pods   ClusterIP   None         <none>        27017/TCP   3m44s

NAME                          READY   AGE
statefulset.apps/mongodb-rs   3/3     3m45s

NAME                                            TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb-rs   kubedb.com/mongodb   5.0.3     72s

NAME                            VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb-rs   5.0.3     Ready    4m10s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb-rs
NAME         VERSION   STATUS   AGE
mongodb-rs   5.0.3     Ready    5m31s
```
> We have successfully deployed MongoDB cluster in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `mongodb-rs` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE                       DATA   AGE
mongodb-rs-auth   kubernetes.io/basic-auth   2      6m20s
mongodb-rs-key    Opaque                     1      6m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
mongodb-rs        ClusterIP   10.8.1.145   <none>        27017/TCP   6m39s
mongodb-rs-pods   ClusterIP   None         <none>        27017/TCP   6m39s
```
Now, we are going to use `mongodb-rs-auth` to export credentials.
Let’s export the `USER` and `PASSWORD` as environment variables to make further commands re-usable.

```bash
$ export USER=$(kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\password}' | base64 -d)
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB pod and insert some sample data. 

```bash
$ kubectl exec -it -n demo mongodb-rs-0 -- mongo admin -u $USER -p $PASSWORD
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
MongoDB shell version v5.0.3
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("489f58e0-ad19-4b25-8933-d6e49eb1e21c") }
MongoDB server version: 5.0.3

rs0:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs0:PRIMARY> use musicdb
switched to db musicdb

rs0:PRIMARY> db.songs.insert({"name":"Take Me Home Country Roads"});
WriteResult({ "nInserted" : 1 })

rs0:PRIMARY> db.songs.find().pretty()
{
	"_id" : ObjectId("6426c44cdf79c82c76cd3e44"),
	"name" : "Take Me Home Country Roads"
}

rs0:PRIMARY> exit
bye
```

> We've successfully inserted some sample data to our database. More information about Run & Manage MongoDB on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)


## Horizontal Scaling of MongoDB Cluster

### Horizontal Scale Up

Here, we are going to increase the number of MongoDB replicas to meet the desired number of replicas.
Before applying Horizontal Scaling, let's check the current number of MongoDB replicas,

```bash
$ kubectl get mongodb -n demo mongodb-rs -o json | jq '.spec.replicas'
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
    name: mongodb-rs
  horizontalScaling:
      replicas: 5
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   2m26s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mongodb -n demo mongodb-rs -o json | jq '.spec.replicas'
5
```

From all the above outputs we can see that the number of replicas is now increased to 5. That means we have successfully scaled up the number of MongoDB replicas.

### Horizontal Scale Down

Now, we are going to scale down the number of MongoDB replicas to meet the desired number of replicas.

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
    name: mongodb-rs
  horizontalScaling:
      replicas: 3
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

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

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of MongoDB replicas,

```bash
$ kubectl get mongodb -n demo mongodb-rs -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the number of replicas is now decreased to 3. That means we have successfully scaled down the number of MongoDB replicas.


## Vetical Scaling of MongoDB Cluster

We are going to scale up the current cpu resource of the MongoDB cluster by applying Vertical Scaling.
Before applying it, let's check the current resources,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
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
    name: mongodb-rs
  verticalScaling:
    replicaSet:
      requests:
        memory: "1100Mi"
        cpu: "0.55"
      limits:
        memory: "1100Mi"
        cpu: "0.55"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.replicaSet` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS       AGE
vertical-scale-up   VerticalScaling   Successful   4m49s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
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
{}
```
> The above output verifies that we have successfully scaled up the resources of the MongoDB cluster.

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
    name: mongodb-rs
  verticalScaling:
    replicaSet:
      requests:
        memory: "1Gi"
        cpu: "0.5"
      limits:
        memory: "1Gi"
        cpu: "0.5"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mongodb-rs` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.replicaSet` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
mongodbopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   2m
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mongodb-rs-0 -o json | jq '.spec.containers[].resources'
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
> The above output verifies that we have successfully scaled down the resources of the MongoDB cluster.

If you want to learn more about Production-Grade MongoDB you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MongoDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
