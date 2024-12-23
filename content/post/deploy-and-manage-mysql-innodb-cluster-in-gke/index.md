---
title: Deploy and Manage MySQL InnoDB Cluster in Google Kubernetes Engine (GKE)
date: "2023-08-30"
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
- innodb
- innodb-cluster
- kubedb
- kubernetes
- mysql
- mysql-database
- mysql-innodb
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Kafka, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will Deploy and Manage MySQL InnoDB Cluster in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL InnoDB Cluster
3) Read/Write through InnoDB
4) Horizontal Scaling of InnoDB Cluster

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
appscode/kubedb                   	v2023.08.18  	v2023.08.18	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.20.0      	v0.20.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.08.18  	v2023.08.18	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.08.18  	v2023.08.18	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.11.0      	v0.11.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.08.18  	v2023.08.18	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.08.18  	v2023.08.18	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.08.18  	v2023.08.18	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.22.0      	v0.22.3    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.08.18  	v2023.08.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.35.0      	v0.35.2    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.11.0      	v0.11.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.4.1      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.11.0      	v0.11.0    	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.08.18 \
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
NAMESPACE   NAME                                            READY   STATUS    RESTARTS     AGE
kubedb      kubedb-kubedb-autoscaler-6d45d7d4df-cchnj       1/1     Running   0            94s
kubedb      kubedb-kubedb-dashboard-74879c5f86-rbrln        1/1     Running   0            94s
kubedb      kubedb-kubedb-ops-manager-c7cc578c5-l9jds       1/1     Running   0            94s
kubedb      kubedb-kubedb-provisioner-5ffc8fdb9f-7cglg      1/1     Running   0            94s
kubedb      kubedb-kubedb-schema-manager-68f845bbb6-4jwnf   1/1     Running   0            94s
kubedb      kubedb-kubedb-webhook-server-6fdd5b4857-v8tn5   1/1     Running   0            93s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-08-30T04:44:17Z
elasticsearchdashboards.dashboard.kubedb.com      2023-08-30T04:44:18Z
elasticsearches.kubedb.com                        2023-08-30T04:44:19Z
elasticsearchopsrequests.ops.kubedb.com           2023-08-30T04:44:25Z
elasticsearchversions.catalog.kubedb.com          2023-08-30T04:41:41Z
etcds.kubedb.com                                  2023-08-30T04:44:25Z
etcdversions.catalog.kubedb.com                   2023-08-30T04:41:41Z
kafkas.kubedb.com                                 2023-08-30T04:44:52Z
kafkaversions.catalog.kubedb.com                  2023-08-30T04:41:42Z
mariadbautoscalers.autoscaling.kubedb.com         2023-08-30T04:44:18Z
mariadbdatabases.schema.kubedb.com                2023-08-30T04:44:26Z
mariadbopsrequests.ops.kubedb.com                 2023-08-30T04:45:01Z
mariadbs.kubedb.com                               2023-08-30T04:44:26Z
mariadbversions.catalog.kubedb.com                2023-08-30T04:41:42Z
memcacheds.kubedb.com                             2023-08-30T04:44:26Z
memcachedversions.catalog.kubedb.com              2023-08-30T04:41:42Z
mongodbautoscalers.autoscaling.kubedb.com         2023-08-30T04:44:18Z
mongodbdatabases.schema.kubedb.com                2023-08-30T04:44:20Z
mongodbopsrequests.ops.kubedb.com                 2023-08-30T04:44:29Z
mongodbs.kubedb.com                               2023-08-30T04:44:21Z
mongodbversions.catalog.kubedb.com                2023-08-30T04:41:42Z
mysqlautoscalers.autoscaling.kubedb.com           2023-08-30T04:44:19Z
mysqldatabases.schema.kubedb.com                  2023-08-30T04:44:17Z
mysqlopsrequests.ops.kubedb.com                   2023-08-30T04:44:58Z
mysqls.kubedb.com                                 2023-08-30T04:44:18Z
mysqlversions.catalog.kubedb.com                  2023-08-30T04:41:43Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-08-30T04:44:19Z
perconaxtradbopsrequests.ops.kubedb.com           2023-08-30T04:45:15Z
perconaxtradbs.kubedb.com                         2023-08-30T04:44:46Z
perconaxtradbversions.catalog.kubedb.com          2023-08-30T04:41:43Z
pgbouncers.kubedb.com                             2023-08-30T04:44:47Z
pgbouncerversions.catalog.kubedb.com              2023-08-30T04:41:44Z
postgresautoscalers.autoscaling.kubedb.com        2023-08-30T04:44:19Z
postgresdatabases.schema.kubedb.com               2023-08-30T04:44:24Z
postgreses.kubedb.com                             2023-08-30T04:44:25Z
postgresopsrequests.ops.kubedb.com                2023-08-30T04:45:09Z
postgresversions.catalog.kubedb.com               2023-08-30T04:41:44Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-08-30T04:44:20Z
proxysqlopsrequests.ops.kubedb.com                2023-08-30T04:45:12Z
proxysqls.kubedb.com                              2023-08-30T04:44:50Z
proxysqlversions.catalog.kubedb.com               2023-08-30T04:41:44Z
publishers.postgres.kubedb.com                    2023-08-30T04:45:26Z
redisautoscalers.autoscaling.kubedb.com           2023-08-30T04:44:21Z
redises.kubedb.com                                2023-08-30T04:44:50Z
redisopsrequests.ops.kubedb.com                   2023-08-30T04:45:05Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-08-30T04:44:21Z
redissentinelopsrequests.ops.kubedb.com           2023-08-30T04:45:19Z
redissentinels.kubedb.com                         2023-08-30T04:44:51Z
redisversions.catalog.kubedb.com                  2023-08-30T04:41:45Z
subscribers.postgres.kubedb.com                   2023-08-30T04:45:29Z
```

## Deploy MySQL InnoDB Cluster

We are going to Deploy MySQL InnoDB Cluster using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: innodb-cluster
  namespace: demo
spec:
  version: "8.0.31-innodb"
  replicas: 3
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `innodb-cluster.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl apply -f innodb-cluster.yaml
mysql.kubedb.com/innodb-cluster created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of MySQL. Here, we are using MySQL `8.0.31-innodb`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.topology` represents the clustering configuration for MySQL.
* `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `InnoDBCluster` to define the operator that we want to deploy a MySQL Innodb Cluster.
* `spec.topology.innoDBCluster` contains the `InnodbCluster` information. Innodb cluster comes with a router as a load balancer.
* `spec.topology.Router.replicas` is for the number of replica of innodb cluster router.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/) .

Once these are handled correctly and the MySQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                          READY   STATUS    RESTARTS   AGE
pod/innodb-cluster-0          2/2     Running   0          2m47s
pod/innodb-cluster-1          2/2     Running   0          2m20s
pod/innodb-cluster-2          2/2     Running   0          113s
pod/innodb-cluster-router-0   1/1     Running   0          2m46s

NAME                             TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/innodb-cluster           ClusterIP   10.32.8.39    <none>        3306/TCP   2m49s
service/innodb-cluster-pods      ClusterIP   None          <none>        3306/TCP   2m49s
service/innodb-cluster-standby   ClusterIP   10.32.7.233   <none>        3306/TCP   2m49s

NAME                                     READY   AGE
statefulset.apps/innodb-cluster          3/3     2m49s
statefulset.apps/innodb-cluster-router   1/1     2m48s

NAME                                                TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/innodb-cluster   kubedb.com/mysql   8.0.31    2m50s

NAME                              VERSION         STATUS   AGE
mysql.kubedb.com/innodb-cluster   8.0.31-innodb   Ready    2m55s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo innodb-cluster
NAME             VERSION         STATUS   AGE
innodb-cluster   8.0.31-innodb   Ready    3m27s
```
> We have successfully deployed MySQL InnoDB cluster in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database innodb-cluster that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=innodb-cluster
NAME                           TYPE                       DATA   AGE
innodb-cluster-auth            kubernetes.io/basic-auth   2      3m48s
innodb-cluster-router-config   Opaque                     1      3m46s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=innodb-cluster
NAME                     TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
innodb-cluster           ClusterIP   10.32.8.39    <none>        3306/TCP   4m5s
innodb-cluster-pods      ClusterIP   None          <none>        3306/TCP   4m5s
innodb-cluster-standby   ClusterIP   10.32.7.233   <none>        3306/TCP   4m5s
```

Now, we are going to use innodb-cluster-auth to get the credentials.

```
$ kubectl get secrets -n demo innodb-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo innodb-cluster-auth -o jsonpath='{.data.password}' | base64 -d
aF!SlcQtd73MP)t5
```
#### Check InnoDB Cluster Status

Now, we will exec into one of the pod to see the cluster status and list the cluster routers. The main advantage of Innodb cluster is its comes with an admin shell from where you are able to call the MySQL Admin API and configure cluster and it provide some functionality wokring with the cluster.

```bash
$ kubectl exec -it innodb-cluster-0 -n demo -c mysql -- bash

bash-4.4# mysqlsh --user=root --password='aF!SlcQtd73MP)t5'
MySQL Shell 8.0.31

Your MySQL connection id is 13069 (X protocol)
Server version: 8.0.31 MySQL Community Server - GPL
No default schema selected; type \use <schema> to set one.
 MySQL  localhost:33060+ ssl  JS > cluster=dba.getCluster();
<Cluster:innodb_cluster>
 MySQL  localhost:33060+ ssl  JS > cluster.status();
{
    "clusterName": "innodb_cluster", 
    "defaultReplicaSet": {
        "name": "default", 
        "primary": "innodb-cluster-0.innodb-cluster-pods.demo.svc:3306", 
        "ssl": "REQUIRED", 
        "status": "OK", 
        "statusText": "Cluster is ONLINE and can tolerate up to ONE failure.", 
        "topology": {
            "innodb-cluster-0.innodb-cluster-pods.demo.svc:3306": {
                "address": "innodb-cluster-0.innodb-cluster-pods.demo.svc:3306", 
                "memberRole": "PRIMARY", 
                "mode": "R/W", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }, 
            "innodb-cluster-1.innodb-cluster-pods.demo.svc:3306": {
                "address": "innodb-cluster-1.innodb-cluster-pods.demo.svc:3306", 
                "memberRole": "SECONDARY", 
                "mode": "R/O", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }, 
            "innodb-cluster-2.innodb-cluster-pods.demo.svc:3306": {
                "address": "innodb-cluster-2.innodb-cluster-pods.demo.svc:3306", 
                "memberRole": "SECONDARY", 
                "mode": "R/O", 
                "readReplicas": {}, 
                "replicationLag": "applier_queue_applied", 
                "role": "HA", 
                "status": "ONLINE", 
                "version": "8.0.31"
            }
        }, 
        "topologyMode": "Single-Primary"
    }, 
    "groupInformationSourceMember": "innodb-cluster-0.innodb-cluster-pods.demo.svc:3306"
}
 MySQL  localhost:33060+ ssl  JS > cluster.listRouters();
{
    "clusterName": "innodb_cluster", 
    "routers": {
        "innodb-cluster-router-0::": {
            "hostname": "innodb-cluster-router-0", 
            "lastCheckIn": "2023-08-30 04:58:58", 
            "roPort": "6447", 
            "roXPort": "6449", 
            "rwPort": "6446", 
            "rwXPort": "6448", 
            "version": "8.0.31"
        }
    }
}

```
To gather more knowledge about extra funtionalities of InnoDB cluster checkout [MySQL Shell API](https://dev.mysql.com/doc/dev/mysqlsh-api-javascript/8.0/classmysqlsh_1_1dba_1_1_dba.html)


#### Insert Sample Data

In this section, we are going to login into our MySQL database pod and insert some sample data.

```bash
$ kubectl exec -it innodb-cluster-0 -n demo -c mysql -- bash
bash-4.4# mysql --user=root --password='aF!SlcQtd73MP)t5'
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 2354

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE Music;
Query OK, 1 row affected (0.01 sec)

mysql> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected, 1 warning (0.02 sec)

mysql>  INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Annie's Song");
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

> We've successfully inserted some sample data to our database. More information about Run & Manage MySQL on Kubernetes can be found in [Kubernetes MySQL](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)


## Horizontal Scaling of MySQL InnoDB Cluster

### Horizontal Scale Up

Here, we are going to scale up the replicas of the InnoDB cluster replicaset to meet the desired number of replicas after scaling.
Before applying Horizontal Scaling, let’s check the current number of replicas,

```bash
$ kubectl get mysql -n demo innodb-cluster -o json | jq '.spec.replicas'
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
    name: innodb-cluster
  horizontalScaling:
    member: 5
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `innodb-cluster` database.
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
horizontal-scale-up   HorizontalScaling   Successful   2m19s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mysql -n demo innodb-cluster -o json | jq '.spec.replicas'
5
```

> From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the InnoDB cluster.

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
    name: innodb-cluster
  horizontalScaling:
    member: 3
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `innoDB-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired number of replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml 
mysqlopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `MySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ watch  kubectl get mysqlopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   2m28s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mysql -n demo innodb-cluster -o json | jq '.spec.replicas'
3
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the InnoDB cluster.

We have made an in depth tutorial on Deploying Resilient MySQL Cluster Using KubeDB on Kubernetes. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/qkX_SWmRhEc" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MySQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
