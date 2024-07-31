---
title: Horizontal Scaling of MySQL Cluster in Azure Kubernetes Service (AKS)
date: "2024-07-18"
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
- mysql
- mysql-cluster
- mysql-database
- mysql-replication
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Horizontal scaling of MySQL cluster in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL Cluster
3) Read/Write Sample Data
4) Horizontal Scaling of MySQL Cluster

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
NAMESPACE   NAME                                            READY   STATUS        RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-66f6f47cdf-77pb5       1/1     Running       0          5m16s
kubedb      kubedb-kubedb-ops-manager-f79d74f97-g4nd2       0/1     Running       0          5m15s
kubedb      kubedb-kubedb-provisioner-77cf667bdb-tpt25      1/1     Running       0          5m16s
kubedb      kubedb-kubedb-webhook-server-769bd7bc9f-t8w6n   1/1     Running       0          5m15s
kubedb      kubedb-petset-operator-77b6b9897f-7rfsz         1/1     Running       0          5m16s
kubedb      kubedb-petset-webhook-server-85dbdf9c7b-djc8c   2/2     Running       0          5m16s
kubedb      kubedb-sidekick-c898cff4c-bjjqk                 1/1     Running       0          5m16s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-07-18T12:38:16Z
connectclusters.kafka.kubedb.com                   2024-07-18T12:40:07Z
connectors.kafka.kubedb.com                        2024-07-18T12:40:07Z
druidversions.catalog.kubedb.com                   2024-07-18T12:38:16Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-07-18T12:40:03Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-07-18T12:40:03Z
elasticsearches.kubedb.com                         2024-07-18T12:40:03Z
elasticsearchopsrequests.ops.kubedb.com            2024-07-18T12:40:03Z
elasticsearchversions.catalog.kubedb.com           2024-07-18T12:38:16Z
etcdversions.catalog.kubedb.com                    2024-07-18T12:38:16Z
ferretdbversions.catalog.kubedb.com                2024-07-18T12:38:16Z
kafkaautoscalers.autoscaling.kubedb.com            2024-07-18T12:40:07Z
kafkaconnectorversions.catalog.kubedb.com          2024-07-18T12:38:16Z
kafkaopsrequests.ops.kubedb.com                    2024-07-18T12:40:07Z
kafkas.kubedb.com                                  2024-07-18T12:40:07Z
kafkaversions.catalog.kubedb.com                   2024-07-18T12:38:16Z
mariadbarchivers.archiver.kubedb.com               2024-07-18T12:40:11Z
mariadbautoscalers.autoscaling.kubedb.com          2024-07-18T12:40:11Z
mariadbdatabases.schema.kubedb.com                 2024-07-18T12:40:11Z
mariadbopsrequests.ops.kubedb.com                  2024-07-18T12:40:10Z
mariadbs.kubedb.com                                2024-07-18T12:40:10Z
mariadbversions.catalog.kubedb.com                 2024-07-18T12:38:16Z
memcachedversions.catalog.kubedb.com               2024-07-18T12:38:16Z
mongodbarchivers.archiver.kubedb.com               2024-07-18T12:40:15Z
mongodbautoscalers.autoscaling.kubedb.com          2024-07-18T12:40:15Z
mongodbdatabases.schema.kubedb.com                 2024-07-18T12:40:16Z
mongodbopsrequests.ops.kubedb.com                  2024-07-18T12:40:15Z
mongodbs.kubedb.com                                2024-07-18T12:40:15Z
mongodbversions.catalog.kubedb.com                 2024-07-18T12:38:16Z
mssqlserverversions.catalog.kubedb.com             2024-07-18T12:38:16Z
mysqlarchivers.archiver.kubedb.com                 2024-07-18T12:40:19Z
mysqlautoscalers.autoscaling.kubedb.com            2024-07-18T12:40:19Z
mysqldatabases.schema.kubedb.com                   2024-07-18T12:40:20Z
mysqlopsrequests.ops.kubedb.com                    2024-07-18T12:40:19Z
mysqls.kubedb.com                                  2024-07-18T12:40:19Z
mysqlversions.catalog.kubedb.com                   2024-07-18T12:38:16Z
perconaxtradbversions.catalog.kubedb.com           2024-07-18T12:38:16Z
pgbouncerversions.catalog.kubedb.com               2024-07-18T12:38:16Z
pgpoolversions.catalog.kubedb.com                  2024-07-18T12:38:16Z
postgresarchivers.archiver.kubedb.com              2024-07-18T12:40:24Z
postgresautoscalers.autoscaling.kubedb.com         2024-07-18T12:40:23Z
postgresdatabases.schema.kubedb.com                2024-07-18T12:40:24Z
postgreses.kubedb.com                              2024-07-18T12:40:23Z
postgresopsrequests.ops.kubedb.com                 2024-07-18T12:40:23Z
postgresversions.catalog.kubedb.com                2024-07-18T12:38:16Z
proxysqlversions.catalog.kubedb.com                2024-07-18T12:38:17Z
publishers.postgres.kubedb.com                     2024-07-18T12:40:24Z
rabbitmqversions.catalog.kubedb.com                2024-07-18T12:38:17Z
redisautoscalers.autoscaling.kubedb.com            2024-07-18T12:40:28Z
redises.kubedb.com                                 2024-07-18T12:40:28Z
redisopsrequests.ops.kubedb.com                    2024-07-18T12:40:28Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-07-18T12:40:28Z
redissentinelopsrequests.ops.kubedb.com            2024-07-18T12:40:28Z
redissentinels.kubedb.com                          2024-07-18T12:40:28Z
redisversions.catalog.kubedb.com                   2024-07-18T12:38:17Z
schemaregistries.kafka.kubedb.com                  2024-07-18T12:40:07Z
schemaregistryversions.catalog.kubedb.com          2024-07-18T12:38:17Z
singlestoreversions.catalog.kubedb.com             2024-07-18T12:38:17Z
solrversions.catalog.kubedb.com                    2024-07-18T12:38:17Z
subscribers.postgres.kubedb.com                    2024-07-18T12:40:24Z
zookeeperversions.catalog.kubedb.com               2024-07-18T12:38:17Z
```

## Deploy MySQL Cluster

We are going to Deploy MySQL Cluster using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MySQL CR we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-cluster
  namespace: demo
spec:
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
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
Let's save this yaml configuration into `mysql-cluster.yaml` 
Then create the above MySQL CR

```bash
$ kubectl apply -f mysql-cluster.yaml
mysql.kubedb.com/mysql-cluster created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of MySQL. Here, we are using MySQL `8.2.0`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.topology` represents the clustering configuration for MySQL.
* `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `GroupReplication`.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).

Once these are handled correctly and the MySQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                  READY   STATUS    RESTARTS   AGE
pod/mysql-cluster-0   2/2     Running   0          3m48s
pod/mysql-cluster-1   2/2     Running   0          2m37s
pod/mysql-cluster-2   2/2     Running   0          116s

NAME                            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/mysql-cluster           ClusterIP   10.128.138.123   <none>        3306/TCP   3m52s
service/mysql-cluster-pods      ClusterIP   None             <none>        3306/TCP   3m52s
service/mysql-cluster-standby   ClusterIP   10.128.12.248    <none>        3306/TCP   3m52s

NAME                             READY   AGE
statefulset.apps/mysql-cluster   3/3     3m50s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-cluster   kubedb.com/mysql   8.2.0     3m49s

NAME                             VERSION   STATUS   AGE
mysql.kubedb.com/mysql-cluster   8.2.0     Ready    3m53s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo mysql-cluster
NAME            VERSION   STATUS   AGE
mysql-cluster   8.2.0     Ready    4m25s
```
> We have successfully deployed MySQL cluster in Azure. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mysql-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                 TYPE                       DATA   AGE
mysql-cluster-auth   kubernetes.io/basic-auth   2      4m49s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
mysql-cluster           ClusterIP   10.128.138.123   <none>        3306/TCP   5m7s
mysql-cluster-pods      ClusterIP   None             <none>        3306/TCP   5m7s
mysql-cluster-standby   ClusterIP   10.128.12.248    <none>        3306/TCP   5m7s
```

Now, we are going to use `mysql-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.password}' | base64 -d
kkDrVtHzg4)1ZO(0
```

#### Insert Sample Data

In this section, we are going to login into our MySQL database pod and insert some sample data.

```bash
$ kubectl exec -it mysql-cluster-0 -n demo -c mysql -- bash
bash-4.4$ mysql --user=root --password='kkDrVtHzg4)1ZO(0'

Welcome to the MySQL monitor.  Commands end with ; or \g.
Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE Music;
Query OK, 1 row affected (0.00 sec)

mysql> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected, 1 warning (0.01 sec)

mysql> INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Annie's Song");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM Music.Artist;
+----+-------------+--------------+
| id | Name        | Song         |
+----+-------------+--------------+
|  1 | John Denver | Annie's Song |
+----+-------------+--------------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

> We've successfully inserted some sample data to our database. More information about Deploy & Manage MySQL on Kubernetes can be found in [Kubernetes MySQL](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)


## Horizontal Scaling of MySQL Cluster

### Horizontal Scale Up

Here, we are going to scale up the replicas of the MySQL cluster replicaset to meet the desired number of replicas after scaling.
Before applying Horizontal Scaling, let’s check the current number of replicas,

```bash
$ kubectl get mysql -n demo mysql-cluster -o json | jq '.spec.replicas'
3
```

### Create MySQLOpsRequest

In order to scale up, we have to create a `MySQLOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: mysql-cluster
  horizontalScaling:
    member: 5
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mysql-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
mysqlopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `MySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ watch kubectl get mysqlopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   3m23s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mysql -n demo mysql-cluster -o json | jq '.spec.replicas'
5
```

> From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the MySQL cluster.

### Horizontal Scale Down

Now, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.


#### Create MySQLOpsRequest

In order to scale down, again we need to create a `MySQLOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling  
  databaseRef:
    name: mysql-cluster
  horizontalScaling:
    member: 3
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mysql-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
mysqlopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `MySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ watch kubectl get mysqlopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   3m27s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mysql -n demo mysql-cluster -o json | jq '.spec.replicas'
3
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the MySQL cluster.

If you want to learn more about Production-Grade MySQL on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=ThdfNCRulTAsqfnz&amp;list=PLoiT1Gv2KR1gNPaHZtfdBZb6G4wLx6Iks" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MySQL on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
