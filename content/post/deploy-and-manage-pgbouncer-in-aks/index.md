---
title: Deploy and Manage PgBouncer in Azure Kubernetes Service (AKS) Using KubeDB
date: "2023-07-05"
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
- pgbouncer
- postgresql
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Kafka, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy and manage PgBouncer in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy PostgreSQL Clustered Database
3) Deploy PgBouncer Cluster
4) Read/Write through PgBouncer

## Install KubeDB

We will follow the steps to install KubeDB.

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d 
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
appscode/kubedb                   	v2023.06.19  	v2023.06.19	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.19.0      	v0.19.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.06.19  	v2023.06.19	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.06.19  	v2023.06.19	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.10.0      	v0.10.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.06.19  	v2023.06.19	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.06.19  	v2023.06.19	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.06.19  	v2023.06.19	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.21.0      	v0.21.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.06.19  	v2023.06.19	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.34.0      	v0.34.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.10.0      	v0.10.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.10.0      	v0.10.0    	KubeDB Webhook Server by AppsCode  

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.06.19 \
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
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-5bfdd8bfdc-7mldj       1/1     Running   0          73s
kubedb      kubedb-kubedb-dashboard-84ddfdb557-dfnht        1/1     Running   0          73s
kubedb      kubedb-kubedb-ops-manager-bd6b6479f-pp7n9       1/1     Running   0          73s
kubedb      kubedb-kubedb-provisioner-59cdd74486-vpnzr      1/1     Running   0          73s
kubedb      kubedb-kubedb-schema-manager-5f4cfb8475-7dt7n   1/1     Running   0          73s
kubedb      kubedb-kubedb-webhook-server-7b9cb7bf96-ms89v   1/1     Running   0          73s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-07-05T04:50:49Z
elasticsearchdashboards.dashboard.kubedb.com      2023-07-05T04:50:46Z
elasticsearches.kubedb.com                        2023-07-05T04:50:39Z
elasticsearchopsrequests.ops.kubedb.com           2023-07-05T04:50:44Z
elasticsearchversions.catalog.kubedb.com          2023-07-05T04:48:09Z
etcds.kubedb.com                                  2023-07-05T04:50:39Z
etcdversions.catalog.kubedb.com                   2023-07-05T04:48:09Z
kafkas.kubedb.com                                 2023-07-05T04:50:40Z
kafkaversions.catalog.kubedb.com                  2023-07-05T04:48:09Z
mariadbautoscalers.autoscaling.kubedb.com         2023-07-05T04:50:49Z
mariadbdatabases.schema.kubedb.com                2023-07-05T04:50:57Z
mariadbopsrequests.ops.kubedb.com                 2023-07-05T04:50:59Z
mariadbs.kubedb.com                               2023-07-05T04:50:39Z
mariadbversions.catalog.kubedb.com                2023-07-05T04:48:10Z
memcacheds.kubedb.com                             2023-07-05T04:50:39Z
memcachedversions.catalog.kubedb.com              2023-07-05T04:48:10Z
mongodbautoscalers.autoscaling.kubedb.com         2023-07-05T04:50:49Z
mongodbdatabases.schema.kubedb.com                2023-07-05T04:50:56Z
mongodbopsrequests.ops.kubedb.com                 2023-07-05T04:50:48Z
mongodbs.kubedb.com                               2023-07-05T04:50:39Z
mongodbversions.catalog.kubedb.com                2023-07-05T04:48:10Z
mysqlautoscalers.autoscaling.kubedb.com           2023-07-05T04:50:49Z
mysqldatabases.schema.kubedb.com                  2023-07-05T04:50:55Z
mysqlopsrequests.ops.kubedb.com                   2023-07-05T04:50:56Z
mysqls.kubedb.com                                 2023-07-05T04:50:39Z
mysqlversions.catalog.kubedb.com                  2023-07-05T04:48:11Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-07-05T04:50:49Z
perconaxtradbopsrequests.ops.kubedb.com           2023-07-05T04:51:13Z
perconaxtradbs.kubedb.com                         2023-07-05T04:50:39Z
perconaxtradbversions.catalog.kubedb.com          2023-07-05T04:48:11Z
pgbouncers.kubedb.com                             2023-07-05T04:50:40Z
pgbouncerversions.catalog.kubedb.com              2023-07-05T04:48:11Z
postgresautoscalers.autoscaling.kubedb.com        2023-07-05T04:50:49Z
postgresdatabases.schema.kubedb.com               2023-07-05T04:50:57Z
postgreses.kubedb.com                             2023-07-05T04:50:40Z
postgresopsrequests.ops.kubedb.com                2023-07-05T04:51:06Z
postgresversions.catalog.kubedb.com               2023-07-05T04:48:12Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-07-05T04:50:49Z
proxysqlopsrequests.ops.kubedb.com                2023-07-05T04:51:09Z
proxysqls.kubedb.com                              2023-07-05T04:50:40Z
proxysqlversions.catalog.kubedb.com               2023-07-05T04:48:12Z
publishers.postgres.kubedb.com                    2023-07-05T04:51:23Z
redisautoscalers.autoscaling.kubedb.com           2023-07-05T04:50:49Z
redises.kubedb.com                                2023-07-05T04:50:40Z
redisopsrequests.ops.kubedb.com                   2023-07-05T04:51:02Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-07-05T04:50:49Z
redissentinelopsrequests.ops.kubedb.com           2023-07-05T04:51:16Z
redissentinels.kubedb.com                         2023-07-05T04:50:40Z
redisversions.catalog.kubedb.com                  2023-07-05T04:48:12Z
subscribers.postgres.kubedb.com                   2023-07-05T04:51:26Z
```

## Deploy PostgreSQL Clustered Database

Now, we are going to Deploy PostgreSQL Clustered Database using KubeDB.
First, let's create a Namespace in which we will deploy the server.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here, is the yaml of the PostgreSQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2 
kind: Postgres
metadata:
  name: postgres
  namespace: demo
spec:
  version: "15.1"
  replicas: 3 
  standbyMode: Hot 
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

Let's save this yaml configuration into `postgres.yaml` 
Then create the above PostgreSQL CRO

```bash
$ kubectl create -f postgres.yaml
postgres.kubedb.com/postgres created
```
In this yaml,
* `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 15.1`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversions` command.
* `spec.standby` is an optional field that specifies the standby mode `hot` or `warm` to use for standby replicas. In `hot` standby mode, standby replicas can accept connection and run read-only queries. In `warm` standby mode, standby replicas can’t accept connection and only used for replication purpose.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these in [Termination Policy](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy).

Let’s check if the server is ready to use,

```bash
$ kubectl get postgres -n demo postgres
NAME       VERSION   STATUS   AGE
postgres   15.1      Ready    2m50s
```

### Create Database, User & Grant Privileges

Here, we are going to create a database with a couple of users and grant them all privileges to the database. 


```bash
$ kubectl exec -it postgres-0 -n demo -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

postgres-0:/$ psql -c "create database test"
CREATE DATABASE

postgres-0:/$ psql -c "create role roy with login password '12345'"
CREATE ROLE

postgres-0:/$ psql -c "grant all privileges on database test to roy"
GRANT

postgres-0:/$ psql test
psql (15.1)
Type "help" for help.

test=# GRANT ALL ON SCHEMA public TO roy;
GRANT

test=# exit

postgres-0:/$ exit
exit
```

#### Create Secret

Now, we'll create a secret that includes the `User` and `Password` that we created as Postgres roles above.

```bash
apiVersion: v1
stringData:
  password: "12345"
  username: roy
kind: Secret
metadata:
  name: db-user-pass
  namespace: demo
type: kubernetes.io/basic-auth
```
Let's save this yaml configuration into `db-user-pass.yaml` 
Then create the above Secret

```bash
$ kubectl apply -f db-user-pass.yaml 
secret/db-user-pass created
```


#### Create AppBinding

Now, we are going to create a AppBinding which we will connect as a database reference to PgBouncer,

```bash
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: pg-appbinding
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: pg-appbinding
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Postgres
    name: postgres
    namespace: demo
  clientConfig:
    service:
      name: postgres
      path: /
      port: 5432
      query: sslmode=disable
      scheme: postgresql
  parameters:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: StashAddon
    stash:
      addon:
        backupTask:
          name: postgres-backup-15.1
        restoreTask:
          name: postgres-restore-15.1
  secret:
    name: db-user-pass
  type: kubedb.com/postgres
  version: "15.1"
```
Let's save this yaml configuration into `pg-appbinding.yaml` 
Then create the above AppBinding

```bash
$ kubectl apply -f pg-appbinding.yaml 
appbinding.appcatalog.appscode.com/pg-appbinding created
```
In this yaml,
* `spec.appRef` refers to the underlying application. It contains the information of 4 fields named `apiGroup`, `kind`, `name` & `namespace`.
* `spec.clientConfig.service` If you are running the database inside the Kubernetes cluster, you can use Kubernetes service to connect with the database. You have to specify the following fields like `name`, `scheme`, `port` in `spec.clientConfig.service` section if you manually create an `AppBinding` object.


## Deploy PgBouncer Cluster

We are going to Deploy PgBouncer cluster using KubeDB.
Here, is the yaml of the PgBouncer CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pgbouncer
  namespace: demo
spec:
  version: "1.18.0"
  replicas: 3
  databases:
  - alias: "testdb"
    databaseName: "test"
    databaseRef:
      name: "pg-appbinding"
      namespace: demo
  connectionPool:
    port: 5432
    poolMode: session
    authType: md5
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `pgbouncer.yaml` 
Then create the above PgBouncer CRO

```bash
$ kubectl apply -f pgbouncer.yaml
pgbouncer.kubedb.com/pgbouncer created
```
In this yaml,
* `spec.version` field specifies the version of PgBouncer. Here, we are using PgBouncer `1.18.0`. You can list the KubeDB supported versions of PgBouncer by running `$ kubectl get pgbouncerversions` command.
* `spec.databaseName` contains the name of PostgreSQL database which is `test` in this case.
* `spec.databaseRef.name` contains the name of the appbinding which is `pg-appbinding` in this case.
* `spec.databaseRef.namespace` contains the namespace information of backend server.
* `spec.connectionPool.poolMode` specifies when a server connection can be reused by other clients. `session` defines Server is released back to pool after client disconnects.
* `spec.connectionPool.authType` specifies client authentication type.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Let’s check if the server is ready to use,

```bash
$ kubectl get pgbouncer -n demo pgbouncer
NAME        VERSION   STATUS   AGE
pgbouncer   1.18.0    Ready    109s
```

Once all of the above things are handled correctly then you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME              READY   STATUS    RESTARTS   AGE
pod/pgbouncer-0   1/1     Running   0          113s
pod/pgbouncer-1   1/1     Running   0          111s
pod/pgbouncer-2   1/1     Running   0          105s
pod/postgres-0    2/2     Running   0          7m29s
pod/postgres-1    2/2     Running   0          8m11s
pod/postgres-2    2/2     Running   0          8m59s

NAME                       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                      AGE
service/pgbouncer          ClusterIP   10.8.6.63    <none>        5432/TCP                     3m7s
service/pgbouncer-pods     ClusterIP   None         <none>        5432/TCP                     3m7s
service/postgres           ClusterIP   10.8.6.135   <none>        5432/TCP,2379/TCP            8m32s
service/postgres-pods      ClusterIP   None         <none>        5432/TCP,2380/TCP,2379/TCP   8m32s
service/postgres-standby   ClusterIP   10.8.0.37    <none>        5432/TCP                     8m32s

NAME                         READY   AGE
statefulset.apps/pgbouncer   3/3     2m3s
statefulset.apps/postgres    3/3     9m34s

NAME                                               TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/pg-appbinding   kubedb.com/postgres    15.1      3m52s
appbinding.appcatalog.appscode.com/pgbouncer       kubedb.com/pgbouncer   1.18.0    3m7s
appbinding.appcatalog.appscode.com/postgres        kubedb.com/postgres    15.1      8m58s

NAME                           VERSION   STATUS   AGE
postgres.kubedb.com/postgres   15.1      Ready    9m

NAME                             VERSION   STATUS   AGE
pgbouncer.kubedb.com/pgbouncer   1.18.0    Ready    3m27s
```
> We have successfully deployed PgBouncer in AKS. Now, we can exec into the container to use the database.



#### Insert Sample Data

Now, let’s exec to the PgBouncer Pod to enter into PostgreSQL server using previously created user credentials to write and read some sample data to the database,

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- sh
Defaulted container "pgbouncer" out of: pgbouncer, pgbouncer-init-container (init)

$ psql -d "host=localhost user=roy password=12345 dbname=testdb"
psql (15.1)
Type "help" for help.

testdb=> create table music(id int, artist varchar, name varchar);
CREATE TABLE

testdb=> insert into music values(1, 'Bobby Bare', 'Five Hundred Miles');
INSERT 0 1

testdb=> select * from music;
 id |   artist    |  name        
----+-------------+-------------------
  1 |  Bobby Bare | Five Hundred Miles
(1 row)

testdb=> exit
$ exit
```


#### Verify Data in PostgreSQL

Here, we are going to exec into PostgreSQL pod to verify the inserted data through PgBouncer.

```bash
$ kubectl exec -it -n demo postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

$ psql
psql (15.1)
Type "help" for help.

postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 test          | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =Tc/postgres         +
               |          |          |            |            | postgres=CTc/postgres+
               |          |          |            |            | roy=CTc/postgres
(5 rows)

postgres=# \c test
You are now connected to database "test" as user "postgres".

test=# \dt
       List of relations
 Schema | Name  | Type  | Owner 
--------+-------+-------+-------
 public | music | table | roy
(1 row)

test=# select * from music;
 id |   artist    |  name        
----+-------------+-------------------
  1 |  Bobby Bare | Five Hundred Miles
(1 row)

test=# exit
$ exit
```

> We've successfully access our PostgreSQL database through PgBouncer. Click [Run & Manage Production-Grade PgBouncer on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-pgbouncer-on-kubernetes/) for more detailed information.



We have made an in depth tutorial on PostgreSQL Connection Pooling In Kubernetes Using KubeDB PGBouncer. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/am4tabT2lXU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [PgBouncer in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-pgbouncer-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
