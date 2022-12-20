---
title: Deploy and Manage PgBouncer in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB
date: "2022-12-19"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- conncection-pooling
- database
- eks
- kubedb
- kubernetes
- pgbouncer
- postgresql
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy and manage PgBouncer in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

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
appscode/kubedb                   	v2022.10.18  	v2022.10.18	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.14.0      	v0.14.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.10.18  	v2022.10.18	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.10.18  	v2022.10.18	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.5.0       	v0.5.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.10.18  	v2022.10.18	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.10.18  	v2022.10.18	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.16.0      	v0.16.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.10.18  	v2022.10.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.29.0      	v0.29.2    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.5.0       	v0.5.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.22     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.5.0       	v0.5.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.10.18 \
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
kubedb      kubedb-kubedb-autoscaler-c44c66449-6l9vb        1/1     Running   0          92s
kubedb      kubedb-kubedb-dashboard-666897b7b8-jmvsl        1/1     Running   0          92s
kubedb      kubedb-kubedb-ops-manager-bc85d9fb9-fdc88       1/1     Running   0          92s
kubedb      kubedb-kubedb-provisioner-6bf689b479-zzptr      1/1     Running   0          92s
kubedb      kubedb-kubedb-schema-manager-d4bb5999-xpfpr     1/1     Running   0          92s
kubedb      kubedb-kubedb-webhook-server-6cd9d766d7-fn8xt   1/1     Running   0          92s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-12-12T07:10:56Z
elasticsearchdashboards.dashboard.kubedb.com      2022-12-12T07:10:55Z
elasticsearches.kubedb.com                        2022-12-12T07:10:56Z
elasticsearchopsrequests.ops.kubedb.com           2022-12-12T07:11:00Z
elasticsearchversions.catalog.kubedb.com          2022-12-12T06:59:21Z
etcds.kubedb.com                                  2022-12-12T07:11:00Z
etcdversions.catalog.kubedb.com                   2022-12-12T06:59:22Z
mariadbautoscalers.autoscaling.kubedb.com         2022-12-12T07:10:56Z
mariadbdatabases.schema.kubedb.com                2022-12-12T07:11:03Z
mariadbopsrequests.ops.kubedb.com                 2022-12-12T07:11:23Z
mariadbs.kubedb.com                               2022-12-12T07:11:01Z
mariadbversions.catalog.kubedb.com                2022-12-12T06:59:23Z
memcacheds.kubedb.com                             2022-12-12T07:11:01Z
memcachedversions.catalog.kubedb.com              2022-12-12T06:59:24Z
mongodbautoscalers.autoscaling.kubedb.com         2022-12-12T07:10:56Z
mongodbdatabases.schema.kubedb.com                2022-12-12T07:10:59Z
mongodbopsrequests.ops.kubedb.com                 2022-12-12T07:11:04Z
mongodbs.kubedb.com                               2022-12-12T07:11:01Z
mongodbversions.catalog.kubedb.com                2022-12-12T06:59:29Z
mysqlautoscalers.autoscaling.kubedb.com           2022-12-12T07:10:56Z
mysqldatabases.schema.kubedb.com                  2022-12-12T07:10:58Z
mysqlopsrequests.ops.kubedb.com                   2022-12-12T07:11:20Z
mysqls.kubedb.com                                 2022-12-12T07:10:59Z
mysqlversions.catalog.kubedb.com                  2022-12-12T06:59:35Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2022-12-12T07:10:57Z
perconaxtradbopsrequests.ops.kubedb.com           2022-12-12T07:11:39Z
perconaxtradbs.kubedb.com                         2022-12-12T07:11:12Z
perconaxtradbversions.catalog.kubedb.com          2022-12-12T06:59:36Z
pgbouncers.kubedb.com                             2022-12-12T07:11:12Z
pgbouncerversions.catalog.kubedb.com              2022-12-12T06:59:37Z
postgresautoscalers.autoscaling.kubedb.com        2022-12-12T07:10:57Z
postgresdatabases.schema.kubedb.com               2022-12-12T07:11:02Z
postgreses.kubedb.com                             2022-12-12T07:11:03Z
postgresopsrequests.ops.kubedb.com                2022-12-12T07:11:31Z
postgresversions.catalog.kubedb.com               2022-12-12T06:59:38Z
proxysqlopsrequests.ops.kubedb.com                2022-12-12T07:11:35Z
proxysqls.kubedb.com                              2022-12-12T07:11:13Z
proxysqlversions.catalog.kubedb.com               2022-12-12T06:59:39Z
publishers.postgres.kubedb.com                    2022-12-12T07:11:46Z
redisautoscalers.autoscaling.kubedb.com           2022-12-12T07:10:57Z
redises.kubedb.com                                2022-12-12T07:11:14Z
redisopsrequests.ops.kubedb.com                   2022-12-12T07:11:27Z
redissentinelautoscalers.autoscaling.kubedb.com   2022-12-12T07:10:57Z
redissentinelopsrequests.ops.kubedb.com           2022-12-12T07:11:43Z
redissentinels.kubedb.com                         2022-12-12T07:11:14Z
redisversions.catalog.kubedb.com                  2022-12-12T06:59:41Z
subscribers.postgres.kubedb.com                   2022-12-12T07:11:50Z
```

## Deploy PostgreSQL Clustered Database

Now, we are going to Deploy PostgreSQL Clustered Database using KubeDB.
First, let's create a Namespace in which we will deploy the server.

```bash
$ kubectl create ns demo
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
  version: "13.2"
  replicas: 3 
  standbyMode: Hot 
  storageType: Durable 
  storage:
    storageClassName: "gp2" 
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
* `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 13.2`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversions` command.
* `spec.standby` is an optional field that specifies the standby mode `hot` or `warm` to use for standby replicas. In `hot` standby mode, standby replicas can accept connection and run read-only queries. In `warm` standby mode, standby replicas can’t accept connection and only used for replication purpose.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy).

Let’s check if the server is ready to use,

```bash
$ kubectl get postgres -n demo postgres
NAME       VERSION   STATUS   AGE
postgres   13.2      Ready    2m17s
```

### Create Database, User & Grant Privileges

Here, we are going to create a database with a couple of users and grant them all privileges to the database. 

```bash
$ kubectl exec -it postgres-0 -n demo -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)

$ psql -c "create database test" && psql -c "create role roy with login password '12345'" && psql -c "grant all privileges on database test to roy" && psql -c "create role rakib with login password '67890'" && psql -c "grant all privileges on database test to rakib"
CREATE DATABASE
CREATE ROLE
GRANT
CREATE ROLE
GRANT
```

#### Create Secret

Now, we'll create a secret with a userlist that includes the `User` and `Password` that we created above.

```bash
#userlist
"roy" "12345"
"rakib" "67890"
```

```bash
$ kubectl create secret generic -n pg db-user-pass --from-file=userlist
secret/db-user-pass created
```

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
  version: 1.17.0
  replicas: 3
  databases:
  - alias: "testdb"
    databaseName: "test"
    databaseRef:
      name: "postgres"
      namespace: demo 
  connectionPool:
    port: 5432
    poolMode: session
    authType: md5
    adminUsers:
    - roy
    authUser: roy
  userListSecretRef:
    name: db-user-pass
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `pgbouncer.yaml` 
Then create the above PgBouncer CRO

```bash
$ kubectl create -f pgbouncer.yaml
pgbouncer.kubedb.com/pgbouncer created
```
In this yaml,
* `spec.version` field specifies the version of PgBouncer. Here, we are using PgBouncer `1.17.0`. You can list the KubeDB supported versions of PgBouncer by running `$ kubectl get pgbouncerversions` command.
* `spec.databaseName` contains the name of PostgreSQL database which is `test` in this case.
* `spec.databaseRef.name` contains the name of backend server which is `postgres` in this case.
* `spec.databaseRef.namespace` contains the namespace information of backend server.
* `spec.connectionPool.poolMode` specifies when a server connection can be reused by other clients. `session` defines Server is released back to pool after client disconnects.
* `spec.connectionPool.authType` specifies client authentication type.
* `spec.userListSecretRef.name` specifes the name of the secret which contains a userlist.  
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Let’s check if the server is ready to use,

```bash
$ kubectl get pgbouncer -n demo pgbouncer
NAME        VERSION   STATUS   AGE
pgbouncer   1.17.0    Ready    4m12s
```

Once all of the above things are handled correctly then you will see that the following objects are created:

```bash
NAME              READY   STATUS    RESTARTS   AGE
pod/pgbouncer-0   1/1     Running   0          4m42s
pod/pgbouncer-1   1/1     Running   0          4m39s
pod/pgbouncer-2   1/1     Running   0          4m36s
pod/postgres-0    2/2     Running   0          26m
pod/postgres-1    2/2     Running   0          26m
pod/postgres-2    2/2     Running   0          26m

NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/pgbouncer          ClusterIP   10.100.125.190   <none>        5432/TCP                     5m
service/pgbouncer-pods     ClusterIP   None             <none>        5432/TCP                     5m
service/postgres           ClusterIP   10.100.174.89    <none>        5432/TCP,2379/TCP            27m
service/postgres-pods      ClusterIP   None             <none>        5432/TCP,2380/TCP,2379/TCP   27m
service/postgres-standby   ClusterIP   10.100.225.217   <none>        5432/TCP                     27m

NAME                         READY   AGE
statefulset.apps/pgbouncer   3/3     5m42s
statefulset.apps/postgres    3/3     27m

NAME                                           TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/pgbouncer   kubedb.com/pgbouncer   1.17.0    5m51s
appbinding.appcatalog.appscode.com/postgres    kubedb.com/postgres    13.2      28m

NAME                           VERSION   STATUS   AGE
postgres.kubedb.com/postgres   13.2      Ready    28m

NAME                             VERSION   STATUS   AGE
pgbouncer.kubedb.com/pgbouncer   1.17.0    Ready    6m29s

```
> We have successfully deployed PgBouncer in Amazon EKS. Now, we can exec into the container to use the database.




#### Insert Sample Data

Now, let’s exec to the PgBouncer Pod to enter into PostgreSQL server using previously created user credentials to write and read some sample data to the database,

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- sh
Defaulted container "pgbouncer" out of: pgbouncer, pgbouncer-init-container (init)

$ psql -d "host=localhost user=roy password=12345 dbname=testdb"
psql (14.2, server 13.2)
Type "help" for help.

testdb=> create table Music(id int, artist varchar, name varchar);
CREATE TABLE

testdb=> insert into Music values(1, 'Bon Jovi', 'Its My Life');
INSERT 0 1

testdb=> select * from music;
 id |  artist  |    name     
----+----------+-------------
  1 | Bon Jovi | Its My Life
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
psql (13.2)
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
               |          |          |            |            | roy=CTc/postgres     +
               |          |          |            |            | rakib=CTc/postgres
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
 id |  artist  |    name     
----+----------+-------------
  1 | Bon Jovi | Its My Life
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

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [PgBouncer in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-pgbouncer-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
