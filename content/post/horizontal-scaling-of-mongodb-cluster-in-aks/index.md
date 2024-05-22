---
title: Horizontal Scaling of MongoDB Cluster in Azure Kubernetes Service (AKS)
date: "2024-05-15"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- dbaas
- high-availability
- kubedb
- kubernetes
- microsoft-azure
- mongodb
- mongodb-cluster
- mongodb-database
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Horizontal scaling of MongoDB cluster in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB Cluster
3) Read/Write Sample Data
4) Horizontal Scaling of MongoDB Cluster

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
appscode/kubedb                   	v2024.4.27   	v2024.4.27 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.30.0      	v0.30.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.4.27   	v2024.4.27 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.9       	v0.0.9     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.4.27   	v2024.4.27 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.21.0      	v0.21.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.4.27   	v2024.4.27 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.4.27   	v2024.4.27 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.4.27   	v2024.4.27 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.32.0      	v0.32.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.4.27   	v2024.4.27 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.45.0      	v0.45.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.21.0      	v0.21.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.5.3    	0.6.6      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.21.0      	v0.21.1    	KubeDB Webhook Server by AppsCode 

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.4.27 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-9f945c8dc-rmrkg        1/1     Running   0          3m34s
kubedb      kubedb-kubedb-ops-manager-7dcc797544-2lv8s      1/1     Running   0          3m34s
kubedb      kubedb-kubedb-provisioner-ddb87568d-sw7pn       1/1     Running   0          3m34s
kubedb      kubedb-kubedb-webhook-server-66fd46fb57-4w5f9   1/1     Running   0          3m34s
kubedb      kubedb-petset-operator-5d94b4ddb8-kg2ql         1/1     Running   0          3m34s
kubedb      kubedb-petset-webhook-server-d7887fd67-hctkb    2/2     Running   0          3m34s
kubedb      kubedb-sidekick-5d9947bd9-q9lkj                 1/1     Running   0          3m34s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-05-15T06:45:16Z
connectors.kafka.kubedb.com                        2024-05-15T06:45:16Z
druidversions.catalog.kubedb.com                   2024-05-15T06:40:34Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-05-15T06:45:13Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-05-15T06:45:13Z
elasticsearches.kubedb.com                         2024-05-15T06:45:13Z
elasticsearchopsrequests.ops.kubedb.com            2024-05-15T06:45:13Z
elasticsearchversions.catalog.kubedb.com           2024-05-15T06:40:34Z
etcdversions.catalog.kubedb.com                    2024-05-15T06:40:35Z
ferretdbversions.catalog.kubedb.com                2024-05-15T06:40:35Z
kafkaautoscalers.autoscaling.kubedb.com            2024-05-15T06:45:16Z
kafkaconnectorversions.catalog.kubedb.com          2024-05-15T06:40:35Z
kafkaopsrequests.ops.kubedb.com                    2024-05-15T06:45:16Z
kafkas.kubedb.com                                  2024-05-15T06:45:16Z
kafkaversions.catalog.kubedb.com                   2024-05-15T06:40:35Z
mariadbarchivers.archiver.kubedb.com               2024-05-15T06:45:19Z
mariadbautoscalers.autoscaling.kubedb.com          2024-05-15T06:45:19Z
mariadbdatabases.schema.kubedb.com                 2024-05-15T06:45:19Z
mariadbopsrequests.ops.kubedb.com                  2024-05-15T06:45:19Z
mariadbs.kubedb.com                                2024-05-15T06:45:19Z
mariadbversions.catalog.kubedb.com                 2024-05-15T06:40:35Z
memcachedversions.catalog.kubedb.com               2024-05-15T06:40:35Z
mongodbarchivers.archiver.kubedb.com               2024-05-15T06:45:23Z
mongodbautoscalers.autoscaling.kubedb.com          2024-05-15T06:45:23Z
mongodbdatabases.schema.kubedb.com                 2024-05-15T06:45:23Z
mongodbopsrequests.ops.kubedb.com                  2024-05-15T06:45:23Z
mongodbs.kubedb.com                                2024-05-15T06:45:22Z
mongodbversions.catalog.kubedb.com                 2024-05-15T06:40:35Z
mssqlserverversions.catalog.kubedb.com             2024-05-15T06:40:35Z
mysqlarchivers.archiver.kubedb.com                 2024-05-15T06:45:26Z
mysqlautoscalers.autoscaling.kubedb.com            2024-05-15T06:45:26Z
mysqldatabases.schema.kubedb.com                   2024-05-15T06:45:26Z
mysqlopsrequests.ops.kubedb.com                    2024-05-15T06:45:26Z
mysqls.kubedb.com                                  2024-05-15T06:45:26Z
mysqlversions.catalog.kubedb.com                   2024-05-15T06:40:35Z
perconaxtradbversions.catalog.kubedb.com           2024-05-15T06:40:35Z
pgbouncerversions.catalog.kubedb.com               2024-05-15T06:40:35Z
pgpoolversions.catalog.kubedb.com                  2024-05-15T06:40:35Z
postgresarchivers.archiver.kubedb.com              2024-05-15T06:45:29Z
postgresautoscalers.autoscaling.kubedb.com         2024-05-15T06:45:29Z
postgresdatabases.schema.kubedb.com                2024-05-15T06:45:29Z
postgreses.kubedb.com                              2024-05-15T06:45:29Z
postgresopsrequests.ops.kubedb.com                 2024-05-15T06:45:29Z
postgresversions.catalog.kubedb.com                2024-05-15T06:40:35Z
proxysqlversions.catalog.kubedb.com                2024-05-15T06:40:35Z
publishers.postgres.kubedb.com                     2024-05-15T06:45:30Z
rabbitmqversions.catalog.kubedb.com                2024-05-15T06:40:35Z
redisautoscalers.autoscaling.kubedb.com            2024-05-15T06:45:33Z
redises.kubedb.com                                 2024-05-15T06:45:33Z
redisopsrequests.ops.kubedb.com                    2024-05-15T06:45:33Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-05-15T06:45:33Z
redissentinelopsrequests.ops.kubedb.com            2024-05-15T06:45:33Z
redissentinels.kubedb.com                          2024-05-15T06:45:33Z
redisversions.catalog.kubedb.com                   2024-05-15T06:40:35Z
singlestoreversions.catalog.kubedb.com             2024-05-15T06:40:35Z
solrversions.catalog.kubedb.com                    2024-05-15T06:40:35Z
subscribers.postgres.kubedb.com                    2024-05-15T06:45:30Z
zookeeperversions.catalog.kubedb.com               2024-05-15T06:40:35Z
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
    storageClassName: "default"
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
pod/mongodb-rs-0   2/2     Running   0          7m31s
pod/mongodb-rs-1   2/2     Running   0          3m3s
pod/mongodb-rs-2   2/2     Running   0          2m35s

NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mongodb-rs        ClusterIP   10.96.151.68   <none>        27017/TCP   7m45s
service/mongodb-rs-pods   ClusterIP   None           <none>        27017/TCP   7m45s

NAME                          READY   AGE
statefulset.apps/mongodb-rs   3/3     7m31s

NAME                                            TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb-rs   kubedb.com/mongodb   7.0.8     7m17s

NAME                            VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb-rs   7.0.8     Ready    7m45s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb-rs
NAME         VERSION   STATUS   AGE
mongodb-rs   7.0.8     Ready    8m3s
```
> We have successfully deployed MongoDB cluster in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database mongodb that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE                       DATA   AGE
mongodb-rs-auth   kubernetes.io/basic-auth   2      9m42s
mongodb-rs-key    Opaque                     1      9m42s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb-rs
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mongodb-rs        ClusterIP   10.96.151.68   <none>        27017/TCP   9m56s
mongodb-rs-pods   ClusterIP   None           <none>        27017/TCP   9m56s
```

Now, we are going to use `mongodb-rs-auth` to export credentials.

```bash
$ kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongodb-rs-auth -o jsonpath='{.data.\password}' | base64 -d
LRhf)ac41Y_LxE;d
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB pod and insert some sample data.

```bash
$ kubectl exec -it mongodb-rs-0 -n demo bash
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mongodb-rs-0:/$ mongosh admin -u root -p 'LRhf)ac41Y_LxE;d'

Using MongoDB:		7.0.8
Using Mongosh:		2.2.5

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

rs0 [direct: primary] admin> show dbs
admin          172.00 KiB
config         344.00 KiB
kubedb-system   40.00 KiB
local          448.00 KiB

rs0 [direct: primary] admin> use musicdb
switched to db musicdb

rs0 [direct: primary] musicdb> db.songs.insert({"name":"The Nights"});
{
  acknowledged: true,
  insertedIds: { '0': ObjectId('66472c4464d2c14f2c2202d8') }
}

rs0 [direct: primary] musicdb> db.songs.find().pretty()
[ { _id: ObjectId('66472c4464d2c14f2c2202d8'), name: 'The Nights' } ]

rs0 [direct: primary] musicdb> exit
mongodb@mongodb-rs-0:/$ exit
exit
```

> We've successfully inserted some sample data to our database. More information about Deploy & Manage MongoDB on Kubernetes can be found in [Kubernetes MongoDB](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)


## Horizontal Scaling of MongoDB Cluster

### Horizontal Scale Up

Here, we are going to scale up the replicas of the MongoDB cluster replicas to meet the desired number of replicas after scaling.
Before applying Horizontal Scaling, let’s check the current number of replicas,

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
- `spec.horizontalScaling.member` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
mongodbopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   2m58s

```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mongodb -n demo mongodb-rs -o json | jq '.spec.replicas'
5
```

> From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the MongoDB cluster.

### Horizontal Scale Down

Now, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.


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
- `spec.horizontalScaling.member` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
mongodbopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `MongoDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   117s
```

From the above output we can see that the `MongoDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mongodb -n demo mongodb-rs -o json | jq '.spec.replicas'
3
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the MongoDB cluster.

If you want to learn more about Production-Grade MongoDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=f0W-h9paKMX9042P&amp;list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MongoDB on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
