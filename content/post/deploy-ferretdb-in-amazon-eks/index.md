---
title: Deploy FerretDB in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB
date: "2024-03-11"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- eks
- ferretdb
- kubedb
- kubernetes
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy FerretDB in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB. We will cover the following steps:

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
appscode/kubedb                   	v2024.2.14   	v2024.2.14 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.27.0      	v0.27.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.2.14   	v2024.2.14 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.7       	v0.0.7     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.2.14   	v2024.2.14 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.18.0      	v0.18.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.2.14   	v2024.2.14 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.2.14   	v2024.2.14 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.2.14   	v2024.2.14 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.29.0      	v0.29.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.2.14   	v2024.2.14 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.42.0      	v0.42.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.18.0      	v0.18.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.2.13   	0.6.4      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.18.0      	v0.18.0    	KubeDB Webhook Server by AppsCode 


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.2.14 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.FerretDB=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-77bc658b47-8rsw5       1/1     Running   0          2m6s
kubedb      kubedb-kubedb-ops-manager-564cd7ddd5-59c4h      1/1     Running   0          2m6s
kubedb      kubedb-kubedb-provisioner-5c4687f696-zn2tc      1/1     Running   0          2m6s
kubedb      kubedb-kubedb-webhook-server-7ccfd65f9d-ljqxp   1/1     Running   0          2m6s
kubedb      kubedb-sidekick-8684467889-2ck7q                1/1     Running   0          2m6s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-03-11T09:46:57Z
connectors.kafka.kubedb.com                        2024-03-11T09:46:57Z
druidversions.catalog.kubedb.com                   2024-03-11T09:46:14Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-03-11T09:46:54Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-03-11T09:46:54Z
elasticsearches.kubedb.com                         2024-03-11T09:46:54Z
elasticsearchopsrequests.ops.kubedb.com            2024-03-11T09:46:54Z
elasticsearchversions.catalog.kubedb.com           2024-03-11T09:46:14Z
etcdversions.catalog.kubedb.com                    2024-03-11T09:46:14Z
ferretdbs.kubedb.com                               2024-03-11T09:46:14Z
ferretdbversions.catalog.kubedb.com                2024-03-11T09:46:14Z
kafkaconnectorversions.catalog.kubedb.com          2024-03-11T09:46:14Z
kafkaopsrequests.ops.kubedb.com                    2024-03-11T09:46:57Z
kafkas.kubedb.com                                  2024-03-11T09:46:57Z
kafkaversions.catalog.kubedb.com                   2024-03-11T09:46:14Z
mariadbautoscalers.autoscaling.kubedb.com          2024-03-11T09:47:00Z
mariadbdatabases.schema.kubedb.com                 2024-03-11T09:47:00Z
mariadbopsrequests.ops.kubedb.com                  2024-03-11T09:47:00Z
mariadbs.kubedb.com                                2024-03-11T09:47:00Z
mariadbversions.catalog.kubedb.com                 2024-03-11T09:46:14Z
memcachedversions.catalog.kubedb.com               2024-03-11T09:46:14Z
mongodbarchivers.archiver.kubedb.com               2024-03-11T09:47:03Z
mongodbautoscalers.autoscaling.kubedb.com          2024-03-11T09:47:03Z
mongodbdatabases.schema.kubedb.com                 2024-03-11T09:47:04Z
mongodbopsrequests.ops.kubedb.com                  2024-03-11T09:47:03Z
mongodbs.kubedb.com                                2024-03-11T09:47:03Z
mongodbversions.catalog.kubedb.com                 2024-03-11T09:46:14Z
mysqlarchivers.archiver.kubedb.com                 2024-03-11T09:47:07Z
mysqlautoscalers.autoscaling.kubedb.com            2024-03-11T09:47:07Z
mysqldatabases.schema.kubedb.com                   2024-03-11T09:47:07Z
mysqlopsrequests.ops.kubedb.com                    2024-03-11T09:47:07Z
mysqls.kubedb.com                                  2024-03-11T09:47:07Z
mysqlversions.catalog.kubedb.com                   2024-03-11T09:46:14Z
perconaxtradbversions.catalog.kubedb.com           2024-03-11T09:46:14Z
pgbouncerversions.catalog.kubedb.com               2024-03-11T09:46:14Z
pgpoolversions.catalog.kubedb.com                  2024-03-11T09:46:14Z
postgresarchivers.archiver.kubedb.com              2024-03-11T09:47:10Z
postgresautoscalers.autoscaling.kubedb.com         2024-03-11T09:47:10Z
postgresdatabases.schema.kubedb.com                2024-03-11T09:47:10Z
postgreses.kubedb.com                              2024-03-11T09:47:10Z
postgresopsrequests.ops.kubedb.com                 2024-03-11T09:47:10Z
postgresversions.catalog.kubedb.com                2024-03-11T09:46:14Z
proxysqlversions.catalog.kubedb.com                2024-03-11T09:46:14Z
publishers.postgres.kubedb.com                     2024-03-11T09:47:10Z
rabbitmqversions.catalog.kubedb.com                2024-03-11T09:46:14Z
redisautoscalers.autoscaling.kubedb.com            2024-03-11T09:47:14Z
redises.kubedb.com                                 2024-03-11T09:47:14Z
redisopsrequests.ops.kubedb.com                    2024-03-11T09:47:14Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-03-11T09:47:14Z
redissentinelopsrequests.ops.kubedb.com            2024-03-11T09:47:14Z
redissentinels.kubedb.com                          2024-03-11T09:47:14Z
redisversions.catalog.kubedb.com                   2024-03-11T09:46:14Z
singlestoreversions.catalog.kubedb.com             2024-03-11T09:46:14Z
solrversions.catalog.kubedb.com                    2024-03-11T09:46:14Z
subscribers.postgres.kubedb.com                    2024-03-11T09:47:10Z
zookeeperversions.catalog.kubedb.com               2024-03-11T09:46:14Z
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
  terminationPolicy: WipeOut
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
pod/ferret-0                      1/1     Running   0          2m1s
pod/ferret-pg-backend-0           2/2     Running   0          3m12s
pod/ferret-pg-backend-1           2/2     Running   0          2m25s
pod/ferret-pg-backend-arbiter-0   1/1     Running   0          2m15s

NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/ferret                      ClusterIP   10.96.42.133   <none>        27017/TCP                    3m16s
service/ferret-pg-backend           ClusterIP   10.96.85.117   <none>        5432/TCP,2379/TCP            3m16s
service/ferret-pg-backend-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   3m16s
service/ferret-pg-backend-standby   ClusterIP   10.96.81.188   <none>        5432/TCP                     3m16s

NAME                                         READY   AGE
statefulset.apps/ferret                      1/1     2m1s
statefulset.apps/ferret-pg-backend           2/2     3m12s
statefulset.apps/ferret-pg-backend-arbiter   1/1     2m15s

NAME                                                   TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/ferret              kubedb.com/ferretdb   1.18.0    2m1s
appbinding.appcatalog.appscode.com/ferret-pg-backend   kubedb.com/postgres   13.13     2m15s

NAME                                    VERSION   STATUS   AGE
postgres.kubedb.com/ferret-pg-backend   13.13     Ready    3m16s
```
Let’s check if the `ferret` is ready to use,

```bash
$ kubectl get ferretdb -n demo ferret
NAME     NAMESPACE   VERSION   STATUS   AGE
ferret   demo        1.18.0    Ready    4m7s
```
> We have successfully deployed FerretDB in Amazon EKS.

### Connect with FerretDB

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with FerretDB.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo | grep ferret
ferret                      ClusterIP   10.96.42.133   <none>        27017/TCP                    12m
ferret-pg-backend           ClusterIP   10.96.85.117   <none>        5432/TCP,2379/TCP            12m
ferret-pg-backend-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   12m
ferret-pg-backend-standby   ClusterIP   10.96.81.188   <none>        5432/TCP                     12m
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
ferret-pg-backend-auth   kubernetes.io/basic-auth   2      12m
```

Now, we are going to use `ferret-pg-backend-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo ferret-pg-backend-auth -o jsonpath='{.data.username}' | base64 -d
postgres

$ kubectl get secrets -n demo ferret-pg-backend-auth -o jsonpath='{.data.password}' | base64 -d
Wkh_dX_.w!M_iumW
```

### Insert Sample Data

In this section, we will to log in via [MongoDB Shell](https://www.mongodb.com/try/download/shell) and insert some sample data.

```bash
$ mongosh 'mongodb://postgres:Wkh_dX_.w!M_iumW@localhost:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID:	65f1423ed66d87075d82b09a
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.1.5
Using MongoDB:		7.0.42
Using Mongosh:		2.1.5

------
   The server generated these startup warnings when booting
   2024-03-11T06:22:32.145Z: Powered by FerretDB v1.18.0 and PostgreSQL 13.13 on x86_64-pc-linux-musl, compiled by gcc.
   2024-03-11T06:22:32.145Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-03-11T06:22:32.145Z: The telemetry state is undecided.
   2024-03-11T06:22:32.145Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb> show dbs
kubedb_system  80.00 KiB

ferretdb> use musicdb
switched to db musicdb

musicdb> db.music.insertOne({"John Denver": "Country Roads"})
{
  acknowledged: true,
  insertedId: ObjectId('65f14363fcadc47154a84b3a')
}

musicdb> db.music.insertOne({"Bobby Bare": "Five Hundred Miles"})
{
  acknowledged: true,
  insertedId: ObjectId('65f14678ad212002eb83d7e1')
}

musicdb> db.music.find()
[
  {
    _id: ObjectId('65f14363fcadc47154a84b3a'),
    'John Denver': 'Country Roads'
  },
  {
    _id: ObjectId('65f14678ad212002eb83d7e1'),
    'Bobby Bare': 'Five Hundred Miles'
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
------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "John Denver": {"t": "string"}}, "$k": ["_id", "John Denver"]}, "_id": "65f14363fcadc47154a84b3a", "John Denver": "Country Roads"}
 {"$s": {"p": {"_id": {"t": "objectId"}, "Bobby Bare": {"t": "string"}}, "$k": ["_id", "Bobby Bare"]}, "_id": "65f14678ad212002eb83d7e1", "Bobby Bare": "Five Hundred Miles"}
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
  terminationPolicy: WipeOut
```
Here,
* `spec.postgres.serivce` is service information of users external postgres exist in the cluster.
* `spec.authSecret.name` is the name of the authentication secret of users external postgres database.

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).