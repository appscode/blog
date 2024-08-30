---
title: Deploy and Manage Pgpool in Azure Kubernetes Service (AKS)
date: "2024-06-14"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- conncection-pooling
- database
- kubedb
- kubernetes
- microsoft-azure
- pgpool
- postgresql
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy and manage Pgpool in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy PostgreSQL Cluster
3) Deploy Pgpool
4) Read/Write through Pgpool

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
  --set global.featureGates.Pgpool=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-6559f975b8-fz7sk       1/1     Running   0          2m37s
kubedb      kubedb-kubedb-ops-manager-dcdf65968-vwfgn       1/1     Running   0          2m37s
kubedb      kubedb-kubedb-provisioner-6f5c9ff8db-7zwjh      1/1     Running   0          2m37s
kubedb      kubedb-kubedb-webhook-server-655ff746d4-vzcnf   1/1     Running   0          2m37s
kubedb      kubedb-petset-operator-5d94b4ddb8-w2hwr         1/1     Running   0          2m37s
kubedb      kubedb-petset-webhook-server-58c94d5bbd-tpwmb   2/2     Running   0          2m37s
kubedb      kubedb-sidekick-5d9947bd9-g5fwt                 1/1     Running   0          2m37s

```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-06-14T08:24:25Z
connectclusters.kafka.kubedb.com                   2024-06-14T08:25:00Z
connectors.kafka.kubedb.com                        2024-06-14T08:25:00Z
druidversions.catalog.kubedb.com                   2024-06-14T08:24:25Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-06-14T08:24:57Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-06-14T08:24:57Z
elasticsearches.kubedb.com                         2024-06-14T08:24:57Z
elasticsearchopsrequests.ops.kubedb.com            2024-06-14T08:24:57Z
elasticsearchversions.catalog.kubedb.com           2024-06-14T08:24:26Z
etcdversions.catalog.kubedb.com                    2024-06-14T08:24:26Z
ferretdbversions.catalog.kubedb.com                2024-06-14T08:24:26Z
kafkaautoscalers.autoscaling.kubedb.com            2024-06-14T08:25:00Z
kafkaconnectorversions.catalog.kubedb.com          2024-06-14T08:24:26Z
kafkaopsrequests.ops.kubedb.com                    2024-06-14T08:25:00Z
kafkas.kubedb.com                                  2024-06-14T08:25:00Z
kafkaversions.catalog.kubedb.com                   2024-06-14T08:24:26Z
mariadbarchivers.archiver.kubedb.com               2024-06-14T08:25:03Z
mariadbautoscalers.autoscaling.kubedb.com          2024-06-14T08:25:03Z
mariadbdatabases.schema.kubedb.com                 2024-06-14T08:25:03Z
mariadbopsrequests.ops.kubedb.com                  2024-06-14T08:25:03Z
mariadbs.kubedb.com                                2024-06-14T08:25:03Z
mariadbversions.catalog.kubedb.com                 2024-06-14T08:24:26Z
memcachedversions.catalog.kubedb.com               2024-06-14T08:24:26Z
mongodbarchivers.archiver.kubedb.com               2024-06-14T08:25:07Z
mongodbautoscalers.autoscaling.kubedb.com          2024-06-14T08:25:07Z
mongodbdatabases.schema.kubedb.com                 2024-06-14T08:25:07Z
mongodbopsrequests.ops.kubedb.com                  2024-06-14T08:25:06Z
mongodbs.kubedb.com                                2024-06-14T08:25:06Z
mongodbversions.catalog.kubedb.com                 2024-06-14T08:24:26Z
mssqlserverversions.catalog.kubedb.com             2024-06-14T08:24:26Z
mysqlarchivers.archiver.kubedb.com                 2024-06-14T08:25:10Z
mysqlautoscalers.autoscaling.kubedb.com            2024-06-14T08:25:10Z
mysqldatabases.schema.kubedb.com                   2024-06-14T08:25:10Z
mysqlopsrequests.ops.kubedb.com                    2024-06-14T08:25:10Z
mysqls.kubedb.com                                  2024-06-14T08:25:10Z
mysqlversions.catalog.kubedb.com                   2024-06-14T08:24:26Z
perconaxtradbversions.catalog.kubedb.com           2024-06-14T08:24:26Z
pgbouncerversions.catalog.kubedb.com               2024-06-14T08:24:26Z
pgpoolautoscalers.autoscaling.kubedb.com           2024-06-14T08:25:13Z
pgpoolopsrequests.ops.kubedb.com                   2024-06-14T08:25:13Z
pgpools.kubedb.com                                 2024-06-14T08:25:13Z
pgpoolversions.catalog.kubedb.com                  2024-06-14T08:24:26Z
postgresarchivers.archiver.kubedb.com              2024-06-14T08:25:16Z
postgresautoscalers.autoscaling.kubedb.com         2024-06-14T08:25:16Z
postgresdatabases.schema.kubedb.com                2024-06-14T08:25:17Z
postgreses.kubedb.com                              2024-06-14T08:25:13Z
postgresopsrequests.ops.kubedb.com                 2024-06-14T08:25:16Z
postgresversions.catalog.kubedb.com                2024-06-14T08:24:26Z
proxysqlversions.catalog.kubedb.com                2024-06-14T08:24:26Z
publishers.postgres.kubedb.com                     2024-06-14T08:25:17Z
rabbitmqversions.catalog.kubedb.com                2024-06-14T08:24:26Z
redisautoscalers.autoscaling.kubedb.com            2024-06-14T08:25:20Z
redises.kubedb.com                                 2024-06-14T08:25:20Z
redisopsrequests.ops.kubedb.com                    2024-06-14T08:25:20Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-06-14T08:25:20Z
redissentinelopsrequests.ops.kubedb.com            2024-06-14T08:25:20Z
redissentinels.kubedb.com                          2024-06-14T08:25:20Z
redisversions.catalog.kubedb.com                   2024-06-14T08:24:26Z
schemaregistries.kafka.kubedb.com                  2024-06-14T08:25:00Z
schemaregistryversions.catalog.kubedb.com          2024-06-14T08:24:26Z
singlestoreversions.catalog.kubedb.com             2024-06-14T08:24:26Z
solrversions.catalog.kubedb.com                    2024-06-14T08:24:26Z
subscribers.postgres.kubedb.com                    2024-06-14T08:25:17Z
zookeeperversions.catalog.kubedb.com               2024-06-14T08:24:26Z
```

## Deploy PostgreSQL Cluster

Now, we are going to Deploy PostgreSQL Cluster using KubeDB.
First, let's create a Namespace in which we will deploy the server.

```bash
$ kubectl create namespace demo
namespace/demo created
```

PostgreSQL is readily available in KubeDB as CRD and can easily be deployed. But by default this will create a PostgreSQL server with `max_connections=100`, but we need more than 100 connections for our Pgpool to work as expected.

Pgpool requires at least `2*num_init_children*max_pool*spec.replicas` connections in PostgreSQL server. So we can use [Custom Configuration File](https://kubedb.com/docs/latest/guides/postgres/configuration/using-config-file/) to create a PostgreSQL server with custom `max_connections`.

Now, create a Secret using this configuration file.

#### Create Secret with Custom Configuration

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-configuration
  namespace: demo
stringData:
  user.conf: max_connections=400
```
Let's save this yaml configuration into `pg-configuration.yaml` 
Then create the above Secret,

```bash
kubectl apply -f pg-configuration.yaml
secret/pg-configuration created
```

Here, is the yaml of the PostgreSQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-cluster
  namespace: demo
spec:
  replicas: 3
  version: "16.1"
  configSecret:
    name: pg-configuration
  storageType: Durable
  storage:
    storageClassName: "default"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `postgres-cluster.yaml` 
Then create the above PostgreSQL CRO,

```bash
$ kubectl apply -f postgres-cluster.yaml
postgres.kubedb.com/postgres-cluster created
```
In this yaml,
* `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 16.1`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.configSecret` is an optional field that allows users to provide custom configuration for PostgreSQL.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these in [Termination Policy](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy).

Let’s check if the server is ready to use,

```bash
$ kubectl get postgres -n demo postgres-cluster
NAME               VERSION   STATUS   AGE
postgres-cluster   16.1      Ready    2m49s
```

### Create Database, User & Grant Privileges

Here, we are going to create a database with a new user and grant all privileges to the database. 

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

postgres-cluster-0:/$ psql -c "create database test"
CREATE DATABASE

postgres-cluster-0:/$ psql -c "create role roy with login password '12345'"
CREATE ROLE

postgres-cluster-0:/$ psql -c "grant all privileges on database test to roy"
GRANT

postgres-cluster-0:/$ psql test
psql (16.1)
Type "help" for help.

test=# GRANT ALL ON SCHEMA public TO roy;
GRANT

test=# exit

postgres-cluster-0:/$ exit
exit
```

#### Create Secret

Now, we'll create a secret that includes the `User` and `Password` with values from newly created role and password above. The secret must have two labels, one is `app.kubernetes.io/name: postgreses.kubedb.com` and another is `app.kubernetes.io/instance: <appbinding name>`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-user-pass
  namespace: demo
  labels:
    app.kubernetes.io/instance: postgres-cluster
    app.kubernetes.io/name: postgreses.kubedb.com
stringData:
  password: "12345"
  username: roy
```
Let's save this yaml configuration into `db-user-pass.yaml` 
Then create the above Secret,

```bash
$ kubectl apply -f db-user-pass.yaml 
secret/db-user-pass created
```

## Deploy Pgpool

We are going to Deploy Pgpool using KubeDB.
Here, is the yaml of the Pgpool CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: postgres-cluster
    namespace: demo
  syncUsers: true
  deletionPolicy: WipeOut
```

Let's save this yaml configuration into `pgpool.yaml` 
Then create the above Pgpool CRO,

```bash
$ kubectl apply -f pgpool.yaml
pgpool.kubedb.com/pgpool created
```
In this yaml,
* `spec.version` field specifies the version of Pgpool. Here, we are using Pgpool `4.5.0`. You can list the KubeDB supported versions of Pgpool by running `$ kubectl get pgpoolversions` command.
* `spec.postgresRef` specifies the name and the namespace of the appbinding that points to the PostgreSQL server.
* `spec.syncUsers` specifies whether user want to sync additional users to Pgpool.
* And the `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Let’s check if the server is ready to use,

```bash
$ kubectl get pgpool -n demo pgpool
NAME     TYPE                  VERSION   STATUS   AGE
pgpool   kubedb.com/v1alpha2   4.5.0     Ready    67s
```

Once all of the above things are handled correctly then you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/pgpool-0             1/1     Running   0          72s
pod/postgres-cluster-0   2/2     Running   0          5m34s
pod/postgres-cluster-1   2/2     Running   0          5m2s
pod/postgres-cluster-2   2/2     Running   0          5m2s

NAME                               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/pgpool                     ClusterIP   10.96.81.203   <none>        9999/TCP                     72s
service/pgpool-pods                ClusterIP   None           <none>        9999/TCP                     72s
service/postgres-cluster           ClusterIP   10.96.252.10   <none>        5432/TCP,2379/TCP            5m34s
service/postgres-cluster-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   5m34s
service/postgres-cluster-standby   ClusterIP   10.96.203.14   <none>        5432/TCP                     5m34s

NAME                                READY   AGE
statefulset.apps/postgres-cluster   3/3     5m34s

NAME                                                  TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/postgres-cluster   kubedb.com/postgres   16.1      5m34s

NAME                       TYPE                  VERSION   STATUS   AGE
pgpool.kubedb.com/pgpool   kubedb.com/v1alpha2   4.5.0     Ready    72s

NAME                                   VERSION   STATUS   AGE
postgres.kubedb.com/postgres-cluster   16.1      Ready    5m34s
```
> We have successfully deployed Pgpool in AKS. Now, we can exec into the container to use the database.


### Connect via Pgpool

To connect via Pgpool we have to expose its service to localhost. KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
pgpool                     ClusterIP   10.96.81.203   <none>        9999/TCP                     103s
pgpool-pods                ClusterIP   None           <none>        9999/TCP                     103s
postgres-cluster           ClusterIP   10.96.252.10   <none>        5432/TCP,2379/TCP            6m12s
postgres-cluster-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   6m12s
postgres-cluster-standby   ClusterIP   10.96.203.14   <none>        5432/TCP                     6m12s
```
Here, we are going to use `pgpool` Service to connect. Now, let’s port-forward the `pgpool` Service to the port `9999` to local machine:

```bash
$ kubectl port-forward -n demo svc/pgpool 9999
Forwarding from 127.0.0.1:9999 -> 9999
```

#### Insert Sample Data

Let’s read and write some sample data to the database via Pgpool,

```bash
$ psql --host=localhost --port=9999 --username=roy test
psql (12.18 (Ubuntu 12.18-0ubuntu0.20.04.1), server 16.1)
Type "help" for help.

test=> CREATE TABLE music(id int, artist varchar, name varchar);
CREATE TABLE

test=> INSERT INTO music VALUES(1, 'Avicii', 'The Nights');
INSERT 0 1

test=# SELECT * FROM music;
 id | artist |  name      
----+--------+------------
  1 | Avicii | The Nights
(1 row)

test=> exit
```

#### Verify Data in PostgreSQL

Here, we are going to exec into PostgreSQL pod to verify the inserted data through Pgpool.

```bash
$ kubectl exec -it -n demo postgres-cluster-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
postgres-cluster-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 test          | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =Tc/postgres         +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres+
               |          |          |                 |            |            |            |           | roy=CTc/postgres
(5 rows)

postgres=# \c test
You are now connected to database "test" as user "postgres".

test=# \dt
       List of relations
 Schema | Name  | Type  | Owner 
--------+-------+-------+-------
 public | music | table | roy
(1 row)

test=# SELECT * FROM music;
 id | artist |  name      
----+--------+------------
  1 | Avicii | The Nights
(1 row)

test=# exit
postgres-cluster-0:/$ exit
exit
```

> We've successfully access our PostgreSQL database through Pgpool. Click [Kubernetes Pgpool Documentation](https://kubedb.com/docs/latest/guides/pgpool/) for more detailed information.


We have made an in depth tutorial on Seamlessly Provision and Manage Pgpool on Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/PLEyCstG3X4?si=rAlSf3qWWiEUpLKC" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Pgpool on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-pgpool-on-kubernetes/))

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
