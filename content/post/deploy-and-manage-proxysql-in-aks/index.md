---
title: Deploy and Manage ProxySQL in Azure Kubernetes Service (AKS) Using KubeDB
date: "2023-06-02"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- azure
- aks
- microsoft-azure
- horizontal-scaling
- kubedb
- kubernetes
- proxysql
- vertical-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy and manage ProxySQL in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL Group Replication
3) Deploy ProxySQL Cluster
4) Read/Write through ProxySQL
5) Horizontal Scaling of ProxySQL


### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

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
appscode/kubedb                   	v2023.04.10  	v2023.04.10	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.18.0      	v0.18.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.04.10  	v2023.04.10	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.04.10  	v2023.04.10	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.9.0       	v0.9.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.04.10  	v2023.04.10	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.04.10  	v2023.04.10	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.04.10  	v2023.04.10	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.20.0      	v0.20.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.04.10  	v2023.04.10	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.33.0      	v0.33.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.9.0       	v0.9.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.9.0       	v0.9.0     	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.04.10 \
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
kubedb      kubedb-kubedb-autoscaler-5b8b68c9d9-hqjlz       1/1     Running   0          59s
kubedb      kubedb-kubedb-dashboard-6bfc6fb797-lkl2b        1/1     Running   0          59s
kubedb      kubedb-kubedb-ops-manager-c67d67684-mrgfq       0/1     Running   0          59s
kubedb      kubedb-kubedb-provisioner-54898dd948-tfdd4      1/1     Running   0          59s
kubedb      kubedb-kubedb-schema-manager-fb749f8-ls9xl      1/1     Running   0          59s
kubedb      kubedb-kubedb-webhook-server-85b7fcffd6-vhnpp   1/1     Running   0          59s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-06-01T08:51:07Z
elasticsearchdashboards.dashboard.kubedb.com      2023-06-01T08:50:49Z
elasticsearches.kubedb.com                        2023-06-01T08:50:49Z
elasticsearchopsrequests.ops.kubedb.com           2023-06-01T08:51:00Z
elasticsearchversions.catalog.kubedb.com          2023-06-01T08:48:40Z
etcds.kubedb.com                                  2023-06-01T08:50:53Z
etcdversions.catalog.kubedb.com                   2023-06-01T08:48:40Z
kafkas.kubedb.com                                 2023-06-01T08:50:55Z
kafkaversions.catalog.kubedb.com                  2023-06-01T08:48:40Z
mariadbautoscalers.autoscaling.kubedb.com         2023-06-01T08:51:07Z
mariadbdatabases.schema.kubedb.com                2023-06-01T08:51:04Z
mariadbopsrequests.ops.kubedb.com                 2023-06-01T08:51:14Z
mariadbs.kubedb.com                               2023-06-01T08:50:53Z
mariadbversions.catalog.kubedb.com                2023-06-01T08:48:40Z
memcacheds.kubedb.com                             2023-06-01T08:50:53Z
memcachedversions.catalog.kubedb.com              2023-06-01T08:48:41Z
mongodbautoscalers.autoscaling.kubedb.com         2023-06-01T08:51:08Z
mongodbdatabases.schema.kubedb.com                2023-06-01T08:51:03Z
mongodbopsrequests.ops.kubedb.com                 2023-06-01T08:51:03Z
mongodbs.kubedb.com                               2023-06-01T08:50:53Z
mongodbversions.catalog.kubedb.com                2023-06-01T08:48:41Z
mysqlautoscalers.autoscaling.kubedb.com           2023-06-01T08:51:08Z
mysqldatabases.schema.kubedb.com                  2023-06-01T08:51:02Z
mysqlopsrequests.ops.kubedb.com                   2023-06-01T08:51:11Z
mysqls.kubedb.com                                 2023-06-01T08:50:54Z
mysqlversions.catalog.kubedb.com                  2023-06-01T08:48:41Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-06-01T08:51:08Z
perconaxtradbopsrequests.ops.kubedb.com           2023-06-01T08:51:28Z
perconaxtradbs.kubedb.com                         2023-06-01T08:50:54Z
perconaxtradbversions.catalog.kubedb.com          2023-06-01T08:48:42Z
pgbouncers.kubedb.com                             2023-06-01T08:50:54Z
pgbouncerversions.catalog.kubedb.com              2023-06-01T08:48:42Z
postgresautoscalers.autoscaling.kubedb.com        2023-06-01T08:51:08Z
postgresdatabases.schema.kubedb.com               2023-06-01T08:51:04Z
postgreses.kubedb.com                             2023-06-01T08:50:54Z
postgresopsrequests.ops.kubedb.com                2023-06-01T08:51:21Z
postgresversions.catalog.kubedb.com               2023-06-01T08:48:42Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-06-01T08:51:08Z
proxysqlopsrequests.ops.kubedb.com                2023-06-01T08:51:24Z
proxysqls.kubedb.com                              2023-06-01T08:50:54Z
proxysqlversions.catalog.kubedb.com               2023-06-01T08:48:43Z
publishers.postgres.kubedb.com                    2023-06-01T08:51:37Z
redisautoscalers.autoscaling.kubedb.com           2023-06-01T08:51:08Z
redises.kubedb.com                                2023-06-01T08:50:54Z
redisopsrequests.ops.kubedb.com                   2023-06-01T08:51:17Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-06-01T08:51:08Z
redissentinelopsrequests.ops.kubedb.com           2023-06-01T08:51:31Z
redissentinels.kubedb.com                         2023-06-01T08:50:54Z
redisversions.catalog.kubedb.com                  2023-06-01T08:48:43Z
subscribers.postgres.kubedb.com                   2023-06-01T08:51:41Z
```

## Deploy MySQL Group Replication

Now, we are going to Deploy MySQL Group Replication using KubeDB.
First, let's create a Namespace in which we will deploy the server.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here, is the yaml of the MySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.0.32"
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

Let's save this yaml configuration into `mysql-server.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl apply -f mysql-server.yaml 
mysql.kubedb.com/mysql-server created
```
In this yaml,
* `spec.version` field specifies the version of MySQL. Here, we are using MySQL `version 8.0.32`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.topology` contains the information of clustering configuration for MySQL.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).

Let’s check if the server is ready to use,

```bash
$ kubectl get mysql -n demo mysql-server
NAME           VERSION   STATUS   AGE
mysql-server   8.0.32    Ready    3m21s
```

## Deploy ProxySQL Cluster

We are going to Deploy ProxySQL cluster using KubeDB.
Here, is the yaml of the ProxySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 3
  mode: GroupReplication
  backend:
      name: mysql-server
  syncUsers: true
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `proxy-server.yaml` 
Then create the above ProxySQL CRO

```bash
$ kubectl apply -f proxy-server.yaml 
proxysql.kubedb.com/proxy-server created
```
In this yaml,
* `spec.version` field specifies the version of ProxySQL. Here, we are using ProxySQL `2.4.4-debian`. You can list the KubeDB supported versions of ProxySQL by running `$ kubectl get proxysqlversions` command.
* `spec.backend.name` contains the name of MySQL server backend which is `mysql-server` in this case.
* `spec.syncUsers` confirms that the ProxySQL will sync it's user list with MySQL server or not. 
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate".

Let’s check if the server is ready to use,

```bash
$ kubectl get proxysql -n demo proxy-server
NAME           VERSION        STATUS   AGE
proxy-server   2.4.4-debian   Ready    2m57s
```

Once all of the above things are handled correctly then you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/mysql-server-0   2/2     Running   0          9m13s
pod/mysql-server-1   2/2     Running   0          8m42s
pod/mysql-server-2   2/2     Running   0          6m3s
pod/proxy-server-0   1/1     Running   0          3m19s
pod/proxy-server-1   1/1     Running   0          3m10s
pod/proxy-server-2   1/1     Running   0          84s

NAME                           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
service/mysql-server           ClusterIP   10.0.47.68     <none>        3306/TCP            9m14s
service/mysql-server-pods      ClusterIP   None           <none>        3306/TCP            9m14s
service/mysql-server-standby   ClusterIP   10.0.111.109   <none>        3306/TCP            9m14s
service/proxy-server           ClusterIP   10.0.102.106   <none>        6033/TCP            3m19s
service/proxy-server-pods      ClusterIP   None           <none>        6032/TCP,6033/TCP   3m19s

NAME                            READY   AGE
statefulset.apps/mysql-server   3/3     9m15s
statefulset.apps/proxy-server   3/3     3m21s

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-server   kubedb.com/mysql   8.0.32    9m16s

NAME                            VERSION   STATUS   AGE
mysql.kubedb.com/mysql-server   8.0.32    Ready    9m20s

NAME                               VERSION        STATUS   AGE
proxysql.kubedb.com/proxy-server   2.4.4-debian   Ready    3m26s
```
> We have successfully deployed ProxySQL in Azure. Now, we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create `Secret` and `Service` for `mysql-server` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mysql-server
NAME                TYPE                       DATA   AGE
mysql-server-auth   kubernetes.io/basic-auth   2      10m


$ kubectl get service -n demo -l=app.kubernetes.io/instance=mysql-server
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mysql-server           ClusterIP   10.0.47.68     <none>        3306/TCP   11m
mysql-server-pods      ClusterIP   None           <none>        3306/TCP   11m
mysql-server-standby   ClusterIP   10.0.111.109   <none>        3306/TCP   11m
```
Now, we are going to use `mysql-server-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mysql-server-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-server-auth -o jsonpath='{.data.password}' | base64 -d
3i7ig3RQXKD!2ksc
```

#### Insert Sample Data

Now, let’s exec to the ProxySQL Pod to enter into MySQL server using MySQL user credentials to write and read some sample data to the database,

```bash
$ kubectl exec -it proxy-server-0 -n demo -- bash
root@proxy-server-0:/# mysql --user=root --password='3i7ig3RQXKD!2ksc' --host 127.0.0.1 --port=6033

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.035 sec)

MySQL [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.004 sec)

MySQL [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected, 1 warning (0.049 sec)

MySQL [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Bobby Bare", "Five Hundred Miles");
Query OK, 1 row affected (0.010 sec)

MySQL [(none)]> SELECT * FROM Music.Artist;
+----+----------+--------------+
| id | Name     | Song         |
+----+----------+--------------+
|  1 | Bon Jovi | It's My Life |
+----+----------+--------------+
1 row in set (0.014 sec)

MySQL [(none)]> exit
Bye
root@proxy-server-0:/# exit
exit
```

> We've successfully inserted some sample data to our database. Click [Run & Manage Production-Grade ProxySQL on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-proxysql-on-kubernetes/) for more detailed information.

## Horizontal Scaling of ProxySQL Cluster

### Scale Up Replicas

Here, we are going to scale up the replicas of the ProxySQL cluster replicaset to meet the desired number of replicas after scaling.

Before applying Horizontal Scaling, let's check the current number of replicas,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas'
3
```

Let’s connect to a ProxySQL instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it proxy-server-0 -n demo -- bash
root@proxy-server-0:/# mysql -uadmin -padmin --host 127.0.0.1 --port=6032

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT * FROM proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
3 rows in set (0.004 sec)

MySQL [(none)]> exit
Bye
```


### Create ProxySQLOpsRequest

In order to scale up the replicas of the replicaset of the database, we have to create a `ProxySQLOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 5
```
Here,

- `spec.proxyref.name` specifies that we are performing horizontal scaling operation on `proxy-server` .
- `spec.type` specifies that we are performing `HorizontalScaling` on our ProxySQL.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
proxysqlopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `ProxySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ watch kubectl get proxysqlopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   91s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas'
5
```

Let’s connect to a ProxySQL instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it proxy-server-0 -n demo -- bash
root@proxy-server-0:/# mysql -uadmin -padmin --host 127.0.0.1 --port=6032

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT * FROM proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-3.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-4.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
5 rows in set (0.005 sec)

MySQL [(none)]> exit
Bye
root@proxy-server-0:/# exit
exit
```

From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the ProxySQL cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create ProxySQLOpsRequest

In order to scale down the cluster of the database, we need to create a `ProxySQLOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling  
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `proxy-server`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our ProxySQL.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
proxysqlopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `ProxySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ watch kubectl get proxysqlopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   82s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get proxysql -n demo proxy-server -o json | jq '.spec.replicas'
3
```

Let’s connect to a ProxySQL instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it proxy-server-0 -n demo -- bash
root@proxy-server-0:/# mysql -uadmin -padmin --host 127.0.0.1 --port=6032

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> SELECT * FROM proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
3 rows in set (0.004 sec)

MySQL [(none)]> exit
Bye
root@proxy-server-0:/# exit
exit
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the ProxySQL cluster.


We have made an in depth tutorial on ProxySQL Declarative Provisioning, Reconfiguration and Horizontal Scaling using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/fT_cQDxfU9o" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [ProxySQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-proxysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
