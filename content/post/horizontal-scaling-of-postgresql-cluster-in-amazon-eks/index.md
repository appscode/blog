---
title: Horizontal Scaling of PostgreSQL Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-07-05"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- high-availability
- kubedb
- kubernetes
- postgresql
- postgresql-cluster
- postgresql-database
- postgresql-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Horizontal scaling of PostgreSQL cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy PostgreSQL Cluster
3) Read/Write Sample Data
4) Horizontal Scaling of PostgreSQL Cluster

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
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.6.4 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug

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
appscode/kubedb-ui                	v2024.6.18   	0.6.9      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.6.18   	v2024.6.18 	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.22.0      	v0.22.0    	KubeDB Webhook Server by AppsCode
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-645b49bc48-w6k9h       1/1     Running   0          3m25s
kubedb      kubedb-kubedb-ops-manager-5b67db5b95-gzzlm      1/1     Running   0          3m26s
kubedb      kubedb-kubedb-provisioner-bb69f7dbf-nvtjh       1/1     Running   0          3m26s
kubedb      kubedb-kubedb-webhook-server-85766fd79d-gkd6x   1/1     Running   0          3m26s
kubedb      kubedb-petset-operator-54877fd499-527zm         1/1     Running   0          3m26s
kubedb      kubedb-petset-webhook-server-7d5589d84-crnm4    2/2     Running   0          3m25s
kubedb      kubedb-sidekick-5d9947bd9-qnchg                 1/1     Running   0          3m26s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-07-05T12:36:49Z
connectclusters.kafka.kubedb.com                   2024-07-05T12:37:32Z
connectors.kafka.kubedb.com                        2024-07-05T12:37:32Z
druidversions.catalog.kubedb.com                   2024-07-05T12:36:49Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-07-05T12:37:29Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-07-05T12:37:29Z
elasticsearches.kubedb.com                         2024-07-05T12:37:29Z
elasticsearchopsrequests.ops.kubedb.com            2024-07-05T12:37:29Z
elasticsearchversions.catalog.kubedb.com           2024-07-05T12:36:49Z
etcdversions.catalog.kubedb.com                    2024-07-05T12:36:49Z
ferretdbversions.catalog.kubedb.com                2024-07-05T12:36:49Z
kafkaautoscalers.autoscaling.kubedb.com            2024-07-05T12:37:32Z
kafkaconnectorversions.catalog.kubedb.com          2024-07-05T12:36:49Z
kafkaopsrequests.ops.kubedb.com                    2024-07-05T12:37:32Z
kafkas.kubedb.com                                  2024-07-05T12:37:32Z
kafkaversions.catalog.kubedb.com                   2024-07-05T12:36:49Z
mariadbarchivers.archiver.kubedb.com               2024-07-05T12:37:35Z
mariadbautoscalers.autoscaling.kubedb.com          2024-07-05T12:37:35Z
mariadbdatabases.schema.kubedb.com                 2024-07-05T12:37:35Z
mariadbopsrequests.ops.kubedb.com                  2024-07-05T12:37:35Z
mariadbs.kubedb.com                                2024-07-05T12:37:35Z
mariadbversions.catalog.kubedb.com                 2024-07-05T12:36:49Z
memcachedversions.catalog.kubedb.com               2024-07-05T12:36:49Z
mongodbarchivers.archiver.kubedb.com               2024-07-05T12:37:39Z
mongodbautoscalers.autoscaling.kubedb.com          2024-07-05T12:37:39Z
mongodbdatabases.schema.kubedb.com                 2024-07-05T12:37:39Z
mongodbopsrequests.ops.kubedb.com                  2024-07-05T12:37:39Z
mongodbs.kubedb.com                                2024-07-05T12:37:39Z
mongodbversions.catalog.kubedb.com                 2024-07-05T12:36:49Z
mssqlserverversions.catalog.kubedb.com             2024-07-05T12:36:49Z
mysqlarchivers.archiver.kubedb.com                 2024-07-05T12:37:43Z
mysqlautoscalers.autoscaling.kubedb.com            2024-07-05T12:37:42Z
mysqldatabases.schema.kubedb.com                   2024-07-05T12:37:43Z
mysqlopsrequests.ops.kubedb.com                    2024-07-05T12:37:42Z
mysqls.kubedb.com                                  2024-07-05T12:37:42Z
mysqlversions.catalog.kubedb.com                   2024-07-05T12:36:49Z
perconaxtradbversions.catalog.kubedb.com           2024-07-05T12:36:49Z
pgbouncerversions.catalog.kubedb.com               2024-07-05T12:36:49Z
pgpoolversions.catalog.kubedb.com                  2024-07-05T12:36:49Z
postgresarchivers.archiver.kubedb.com              2024-07-05T12:37:46Z
postgresautoscalers.autoscaling.kubedb.com         2024-07-05T12:37:46Z
postgresdatabases.schema.kubedb.com                2024-07-05T12:37:46Z
postgreses.kubedb.com                              2024-07-05T12:37:46Z
postgresopsrequests.ops.kubedb.com                 2024-07-05T12:37:46Z
postgresversions.catalog.kubedb.com                2024-07-05T12:36:49Z
proxysqlversions.catalog.kubedb.com                2024-07-05T12:36:49Z
publishers.postgres.kubedb.com                     2024-07-05T12:37:46Z
rabbitmqversions.catalog.kubedb.com                2024-07-05T12:36:49Z
redisautoscalers.autoscaling.kubedb.com            2024-07-05T12:37:49Z
redises.kubedb.com                                 2024-07-05T12:37:49Z
redisopsrequests.ops.kubedb.com                    2024-07-05T12:37:49Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-07-05T12:37:49Z
redissentinelopsrequests.ops.kubedb.com            2024-07-05T12:37:49Z
redissentinels.kubedb.com                          2024-07-05T12:37:49Z
redisversions.catalog.kubedb.com                   2024-07-05T12:36:49Z
schemaregistries.kafka.kubedb.com                  2024-07-05T12:37:32Z
schemaregistryversions.catalog.kubedb.com          2024-07-05T12:36:49Z
singlestoreversions.catalog.kubedb.com             2024-07-05T12:36:49Z
solrversions.catalog.kubedb.com                    2024-07-05T12:36:49Z
subscribers.postgres.kubedb.com                    2024-07-05T12:37:46Z
zookeeperversions.catalog.kubedb.com               2024-07-05T12:36:49Z
```

## Deploy PostgreSQL Cluster

We are going to Deploy PostgreSQL Cluster using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the PostgreSQL CR we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-cluster
  namespace: demo
spec:
  version: "16.1"
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
Let's save this yaml configuration into `postgres-cluster.yaml` 
Then create the above PostgreSQL CR

```bash
$ kubectl apply -f postgres-cluster.yaml
postgres.kubedb.com/postgres-cluster created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `16.1`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversions` command.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy) .

Once these are handled correctly and the PostgreSQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/postgres-cluster-0   2/2     Running   0          3m37s
pod/postgres-cluster-1   2/2     Running   0          43s
pod/postgres-cluster-2   2/2     Running   0          34s

NAME                               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/postgres-cluster           ClusterIP   10.96.117.56   <none>        5432/TCP,2379/TCP            4m22s
service/postgres-cluster-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   4m22s
service/postgres-cluster-standby   ClusterIP   10.96.131.2    <none>        5432/TCP                     4m22s

NAME                                READY   AGE
statefulset.apps/postgres-cluster   3/3     3m37s

NAME                                                  TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/postgres-cluster   kubedb.com/postgres   16.1      3m37s

NAME                                   VERSION   STATUS   AGE
postgres.kubedb.com/postgres-cluster   16.1      Ready    4m22s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get postgres -n demo postgres-cluster
NAME               VERSION   STATUS   AGE
postgres-cluster   16.1      Ready    4m59s
```
> We have successfully deployed PostgreSQL cluster in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `postgres-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                    TYPE                       DATA   AGE
postgres-cluster-auth   kubernetes.io/basic-auth   2      7m25s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
postgres-cluster           ClusterIP   10.96.117.56   <none>        5432/TCP,2379/TCP            7m40s
postgres-cluster-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   7m40s
postgres-cluster-standby   ClusterIP   10.96.131.2    <none>        5432/TCP                     7m40s
```

Now, we are going to use `postgres-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.username}' | base64 -d
postgres

$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.password}' | base64 -d
LxZOBQbhB;k7PB2R
```

#### Insert Sample Data

In this section, we are going to login into our PostgreSQL database pod and insert some sample data.

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -c postgres -- bash
postgres-cluster-0:/$ psql -d "user=postgres password=LxZOBQbhB;k7PB2R"
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
(4 rows)

postgres=# CREATE DATABASE music;
CREATE DATABASE
postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 music         | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(5 rows)

postgres=# CREATE TABLE artist (name VARCHAR(50) NOT NULL, song VARCHAR(50) NOT NULL);
CREATE TABLE
postgres=# INSERT INTO artist (name, song) VALUES('BOBBY BARE', 'FIVE HUNDRED MILES');
INSERT 0 1
postgres=# SELECT * FROM artist;
    name    |        song        
------------+--------------------
 BOBBY BARE | FIVE HUNDRED MILES
(1 row)

postgres=# \q
postgres-cluster-0:/$ exit
exit
```

> We've successfully inserted some sample data to our database. More information about Run & Manage PostgreSQL on Kubernetes can be found in [Kubernetes PostgreSQL](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)


## Horizontal Scaling of PostgreSQL Cluster

### Horizontal Scale Up

Here, we are going to scale up the replicas of the PostgreSQL cluster replicaset to meet the desired number of replicas after scaling.
Before applying Horizontal Scaling, let’s check the current number of replicas,

```bash
$ kubectl get postgres -n demo postgres-cluster -o json | jq '.spec.replicas'
3
```

### Create PostgresOpsRequest

In order to scale up, we have to create a `PostgresOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: postgres-cluster
  horizontalScaling:
    replicas: 5
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `postgres-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
postgresopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `PostgresOpsRequest` `STATUS` to be Successful. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ watch kubectl get postgresopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   97s
```

From the above output we can see that the `PostgresOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get postgres -n demo postgres-cluster -o json | jq '.spec.replicas'
5
```

> From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the PostgreSQL cluster.

### Horizontal Scale Down

Now, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.


#### Create PostgresOpsRequest

In order to scale down, again we need to create a `PostgresOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: postgres-cluster
  horizontalScaling:
    replicas: 3
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `postgres-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
postgresopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `PostgresOpsRequest` `STATUS` to be Successful. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ watch kubectl get postgresopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   2m33s
```

From the above output we can see that the `PostgresOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get postgres -n demo postgres-cluster -o json | jq '.spec.replicas'
3
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the PostgreSQL cluster.

If you want to learn more about Production-Grade PostgreSQL on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=RIbDYu9MQqW2n8lK&amp;list=PLoiT1Gv2KR1imqnrYFhUNTLHdBNFXPKr_" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [PostgreSQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
