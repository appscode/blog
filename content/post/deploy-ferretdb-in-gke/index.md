---
title: Deploy FerretDB in Google Kubernetes Engine (GKE) Using KubeDB
date: "2024-06-10"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- ferretdb
- gcp
- gke
- kubedb
- kubernetes
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy FerretDB in Google Kubernetes Engine (GKE) Using KubeDB. We will cover the following steps:

1) Install KubeDB
2) Deploy FerretDB
3) Connect with FerretDB 
4) Read/Write Sample Data

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
appscode/kubedb-ui                	v2024.6.3    	0.6.8      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.22.0      	v0.22.0    	KubeDB Webhook Server by AppsCode


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.6.4 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.FerretDB=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS      AGE
kubedb      kubedb-kubedb-autoscaler-75cf766448-qjlhc       1/1     Running   0             2m47s
kubedb      kubedb-kubedb-ops-manager-7c4b676b49-qxgp8      1/1     Running   0             2m47s
kubedb      kubedb-kubedb-provisioner-75475bd9f4-vz9t9      1/1     Running   0             2m47s
kubedb      kubedb-kubedb-webhook-server-589bfbb687-xdrz9   1/1     Running   0             2m47s
kubedb      kubedb-petset-operator-5d94b4ddb8-tpck5         1/1     Running   0             2m47s
kubedb      kubedb-petset-webhook-server-8647cc4564-fvsg2   2/2     Running   0             2m47s
kubedb      kubedb-sidekick-5d9947bd9-wgnnd                 1/1     Running   0             2m47s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-06-10T10:36:23Z
connectclusters.kafka.kubedb.com                   2024-06-10T10:38:11Z
connectors.kafka.kubedb.com                        2024-06-10T10:38:11Z
druidversions.catalog.kubedb.com                   2024-06-10T10:36:23Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-06-10T10:38:05Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-06-10T10:38:05Z
elasticsearches.kubedb.com                         2024-06-10T10:38:05Z
elasticsearchopsrequests.ops.kubedb.com            2024-06-10T10:38:05Z
elasticsearchversions.catalog.kubedb.com           2024-06-10T10:36:23Z
etcdversions.catalog.kubedb.com                    2024-06-10T10:36:23Z
ferretdbs.kubedb.com                               2024-06-10T10:38:08Z
ferretdbversions.catalog.kubedb.com                2024-06-10T10:36:23Z
kafkaautoscalers.autoscaling.kubedb.com            2024-06-10T10:38:11Z
kafkaconnectorversions.catalog.kubedb.com          2024-06-10T10:36:23Z
kafkaopsrequests.ops.kubedb.com                    2024-06-10T10:38:11Z
kafkas.kubedb.com                                  2024-06-10T10:38:11Z
kafkaversions.catalog.kubedb.com                   2024-06-10T10:36:23Z
mariadbarchivers.archiver.kubedb.com               2024-06-10T10:38:14Z
mariadbautoscalers.autoscaling.kubedb.com          2024-06-10T10:38:14Z
mariadbdatabases.schema.kubedb.com                 2024-06-10T10:38:14Z
mariadbopsrequests.ops.kubedb.com                  2024-06-10T10:38:14Z
mariadbs.kubedb.com                                2024-06-10T10:38:14Z
mariadbversions.catalog.kubedb.com                 2024-06-10T10:36:23Z
memcachedversions.catalog.kubedb.com               2024-06-10T10:36:23Z
mongodbarchivers.archiver.kubedb.com               2024-06-10T10:38:18Z
mongodbautoscalers.autoscaling.kubedb.com          2024-06-10T10:38:18Z
mongodbdatabases.schema.kubedb.com                 2024-06-10T10:38:18Z
mongodbopsrequests.ops.kubedb.com                  2024-06-10T10:38:18Z
mongodbs.kubedb.com                                2024-06-10T10:38:18Z
mongodbversions.catalog.kubedb.com                 2024-06-10T10:36:23Z
mssqlserverversions.catalog.kubedb.com             2024-06-10T10:36:23Z
mysqlarchivers.archiver.kubedb.com                 2024-06-10T10:38:21Z
mysqlautoscalers.autoscaling.kubedb.com            2024-06-10T10:38:21Z
mysqldatabases.schema.kubedb.com                   2024-06-10T10:38:21Z
mysqlopsrequests.ops.kubedb.com                    2024-06-10T10:38:21Z
mysqls.kubedb.com                                  2024-06-10T10:38:21Z
mysqlversions.catalog.kubedb.com                   2024-06-10T10:36:23Z
perconaxtradbversions.catalog.kubedb.com           2024-06-10T10:36:23Z
pgbouncerversions.catalog.kubedb.com               2024-06-10T10:36:23Z
pgpoolversions.catalog.kubedb.com                  2024-06-10T10:36:23Z
postgresarchivers.archiver.kubedb.com              2024-06-10T10:38:25Z
postgresautoscalers.autoscaling.kubedb.com         2024-06-10T10:38:25Z
postgresdatabases.schema.kubedb.com                2024-06-10T10:38:25Z
postgreses.kubedb.com                              2024-06-10T10:38:08Z
postgresopsrequests.ops.kubedb.com                 2024-06-10T10:38:25Z
postgresversions.catalog.kubedb.com                2024-06-10T10:36:23Z
proxysqlversions.catalog.kubedb.com                2024-06-10T10:36:23Z
publishers.postgres.kubedb.com                     2024-06-10T10:38:25Z
rabbitmqversions.catalog.kubedb.com                2024-06-10T10:36:23Z
redisautoscalers.autoscaling.kubedb.com            2024-06-10T10:38:28Z
redises.kubedb.com                                 2024-06-10T10:38:28Z
redisopsrequests.ops.kubedb.com                    2024-06-10T10:38:28Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-06-10T10:38:28Z
redissentinelopsrequests.ops.kubedb.com            2024-06-10T10:38:28Z
redissentinels.kubedb.com                          2024-06-10T10:38:28Z
redisversions.catalog.kubedb.com                   2024-06-10T10:36:23Z
schemaregistries.kafka.kubedb.com                  2024-06-10T10:38:11Z
schemaregistryversions.catalog.kubedb.com          2024-06-10T10:36:23Z
singlestoreversions.catalog.kubedb.com             2024-06-10T10:36:23Z
solrversions.catalog.kubedb.com                    2024-06-10T10:36:23Z
subscribers.postgres.kubedb.com                    2024-06-10T10:38:25Z
zookeeperversions.catalog.kubedb.com               2024-06-10T10:36:23Z
```

## Deploy FerretDB with KubeDB Managed PostgreSQL

We are going to deploy FerretDB using KubeDB. First, let's create a namespace in which we will deploy FerretDB.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the FerretDB CR we are going to use:

```yaml                                                                      
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: demo
spec:
  version: "1.18.0"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  backend:
    externallyManaged: false
  deletionPolicy: WipeOut
```
Let's save this yaml configuration into `ferret.yaml` 
Then create the above FerretDB CR

```bash
$ kubectl apply -f ferret.yaml
ferretdb.kubedb.com/ferret created
```
In this yaml,
* `spec.version` field specifies the version of FerretDB. Here, we are using FerretDB `1.18.0`. You can list the KubeDB supported versions of FerretDB by running `$ kubectl get ferretdbversions` command.
* `spec.storageType` specifies the type of storage that will be used for FerretDB. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`.
* `spec.backend` denotes the backend database information for FerretDB instance.
* `spec.terminationPolicy` field is *Wipeout* means it will be deleted without restrictions.

Once these are handled correctly and the FerretDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                              READY   STATUS    RESTARTS   AGE
pod/ferret-0                      1/1     Running   0          23s
pod/ferret-pg-backend-0           2/2     Running   0          91s
pod/ferret-pg-backend-1           2/2     Running   0          47s
pod/ferret-pg-backend-arbiter-0   1/1     Running   0          34s

NAME                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/ferret                      ClusterIP   10.96.39.167    <none>        27017/TCP                    95s
service/ferret-pg-backend           ClusterIP   10.96.119.129   <none>        5432/TCP,2379/TCP            95s
service/ferret-pg-backend-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   95s
service/ferret-pg-backend-standby   ClusterIP   10.96.140.224   <none>        5432/TCP                     95s

NAME                                         READY   AGE
statefulset.apps/ferret-pg-backend           2/2     91s
statefulset.apps/ferret-pg-backend-arbiter   1/1     34s

NAME                                                   TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/ferret              kubedb.com/ferretdb   1.18.0    23s
appbinding.appcatalog.appscode.com/ferret-pg-backend   kubedb.com/postgres   13.13     91s

NAME                                    VERSION   STATUS   AGE
postgres.kubedb.com/ferret-pg-backend   13.13     Ready    95s
```
Let’s check if the `ferret` is ready to use,

```bash
$ kubectl get ferretdb -n demo ferret
NAME     NAMESPACE   VERSION   STATUS   AGE
ferret   demo        1.18.0    Ready    117s
```
> We have successfully deployed FerretDB in GKE.

### Connect with FerretDB

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with FerretDB.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo | grep ferret
ferret                      ClusterIP   10.96.39.167    <none>        27017/TCP                    2m31s
ferret-pg-backend           ClusterIP   10.96.119.129   <none>        5432/TCP,2379/TCP            2m31s
ferret-pg-backend-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   2m31s
ferret-pg-backend-standby   ClusterIP   10.96.140.224   <none>        5432/TCP                     2m31s
```
Here, we are going to use Service named `ferret`. Now, let’s port-forward the `ferret` Service to the local machine's port `27017`.

```bash
$ kubectl port-forward -n demo svc/ferret 27017
Forwarding from 127.0.0.1:27017 -> 27017
Forwarding from [::1]:27017 -> 27017
```
#### Access the Credentials

KubeDB also create Secret for the `ferret` instance. Let’s check by following command,

```bash
$ kubectl get secret -n demo | grep ferret
ferret-pg-backend-auth   kubernetes.io/basic-auth   2      2m53s
```

Now, we are going to use `ferret-pg-backend-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo ferret-pg-backend-auth -o jsonpath='{.data.username}' | base64 -d
postgres

$ kubectl get secrets -n demo ferret-pg-backend-auth -o jsonpath='{.data.password}' | base64 -d
6QWXX*d8FcuGf95S
```

### Insert Sample Data

In this section, we will to log in via [MongoDB Shell](https://www.mongodb.com/try/download/shell) and insert some sample data.

```bash
$ mongosh 'mongodb://postgres:6QWXX*d8FcuGf95S@localhost:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID:	666ae3d956fa285b80a26a12
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.2.6
Using MongoDB:		7.0.42
Using Mongosh:		2.2.6

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2024-06-10T12:19:37.648Z: Powered by FerretDB v1.18.0 and PostgreSQL 13.13 on x86_64-pc-linux-musl, compiled by gcc.
   2024-06-10T12:19:37.648Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-06-10T12:19:37.648Z: The telemetry state is undecided.
   2024-06-10T12:19:37.648Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb> show dbs
kubedb_system  80.00 KiB


ferretdb> use musicdb
switched to db musicdb

musicdb> db.music.insertOne({"Avicii": "The Nights"})
{
  acknowledged: true,
  insertedId: ObjectId('666ae4b156fa285b80a26a15')
}

musicdb> db.music.insertOne({"Bon Jovi": "It's My Life"})
{
  acknowledged: true,
  insertedId: ObjectId('666ae4b956fa285b80a26a16')
}

musicdb> db.music.find()
[
  { _id: ObjectId('666ae4b156fa285b80a26a15'), Avicii: 'The Nights' },
  {
    _id: ObjectId('666ae4b956fa285b80a26a16'),
    'Bon Jovi': "It's My Life"
  }
]

musicdb> show dbs
kubedb_system  80.00 KiB
musicdb        80.00 KiB

musicdb> exit
```
> Here, we've stored some sample data in our `ferret-pg-backend` PostgreSQL using `mongosh`.

### Verify Data in PostgreSQL Backend Engine

Now, We are going to exec into the PostgreSQL pod to verify if the data has been stored successfully.

```bash
$ kubectl exec -it -n demo ferret-pg-backend-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ferret-pg-backend-0:/$ psql
psql (13.13)
Type "help" for help.

postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 ferretdb      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(5 rows)

postgres=# \c ferretdb
You are now connected to database "ferretdb" as user "postgres".

ferretdb=# \dn
     List of schemas
     Name      |  Owner   
---------------+----------
 kubedb_system | postgres
 musicdb       | postgres
 public        | postgres
(3 rows)

ferretdb=# SET search_path TO musicdb;
SET

ferretdb=# \dt
                    List of relations
 Schema  |            Name             | Type  |  Owner   
---------+-----------------------------+-------+----------
 musicdb | _ferretdb_database_metadata | table | postgres
 musicdb | music_9f9c4fd4              | table | postgres
(2 rows)

ferretdb=# SELECT * FROM music_9f9c4fd4;
                                                                              _jsonb                                                                              
------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "Avicii": {"t": "string"}}, "$k": ["_id", "Avicii"]}, "_id": "666ae4b156fa285b80a26a15", "Avicii": "The Nights"}
 {"$s": {"p": {"_id": {"t": "objectId"}, "Bon Jovi": {"t": "string"}}, "$k": ["_id", "Bon Jovi"]}, "_id": "666ae4b956fa285b80a26a16", "Bon Jovi": "It's My Life"}
(2 rows)

ferretdb=# exit
```

## Deploy FerretDB with Externally Managed PostgreSQL

In this blog post, we demonstrated how to deploy FerretDB with KubeDB Managed PostgreSQL. However, if you prefer to use your own PostgreSQL as the backend engine, you have the flexibility to do so. Below, we provide the yaml configuration for integrating FerretDB with an externally managed PostgreSQL instance.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb-external
  namespace: demo
spec:
  version: "1.18.0"
  authSecret:
    externallyManaged: true
    name: ha-postgres-auth
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi  
  backend:
    externallyManaged: true
    postgres:
      service:
        name: ha-postgres
        namespace: demo
        pgPort: 5432
  deletionPolicy: WipeOut
```
Here,
* `spec.postgres.serivce` is service information of users external postgres exist in the cluster.
* `spec.authSecret.name` is the name of the authentication secret of users external postgres database.

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://youtube.com/@appscode) channel.

More about [FerretDB on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-ferretdb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).