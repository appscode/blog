---
title: Update Version of MySQL Database in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-09-29"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- mysql
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Redis, PostgreSQL, Kafka, MySQL, MongoDB, MariaDB, Elasticsearch, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will update version of MySQL Database in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL Cluster
3) Insert Sample Data
4) Update MySQL Database Version


### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
6c08dcb8-8440-4388-849f-1f2b590b731e
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial, we will use KubeDB Enterprise Edition.

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
appscode/kubedb-autoscaler        	v0.20.0      	v0.20.1    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.08.18  	v2023.08.18	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.08.18  	v2023.08.18	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.11.0      	v0.11.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.08.18  	v2023.08.18	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.08.18  	v2023.08.18	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.08.18  	v2023.08.18	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.22.0      	v0.22.8    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.08.18  	v2023.08.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.35.0      	v0.35.6    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.11.0      	v0.11.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.4.3      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.11.0      	v0.11.1    	KubeDB Webhook Server by AppsCode   

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
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS              RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-dc66f95d5-dg2zv        1/1     Running             0          2m44s
kubedb      kubedb-kubedb-dashboard-9d94bb465-tgw4k         1/1     Running             0          2m44s
kubedb      kubedb-kubedb-ops-manager-979b5f4b-kdwsx        1/1     Running             0          2m44s
kubedb      kubedb-kubedb-provisioner-5c9f759d5-8khw6       1/1     Running             0          2m44s
kubedb      kubedb-kubedb-schema-manager-6c54bf49cb-rvdnc   1/1     Running             0          2m44s
kubedb      kubedb-kubedb-webhook-server-d65c58877-826kf    1/1     Running             0          2m44s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-09-29T04:44:17Z
elasticsearchdashboards.dashboard.kubedb.com      2023-09-29T04:44:18Z
elasticsearches.kubedb.com                        2023-09-29T04:44:19Z
elasticsearchopsrequests.ops.kubedb.com           2023-09-29T04:44:25Z
elasticsearchversions.catalog.kubedb.com          2023-09-29T04:41:41Z
etcds.kubedb.com                                  2023-09-29T04:44:25Z
etcdversions.catalog.kubedb.com                   2023-09-29T04:41:41Z
kafkas.kubedb.com                                 2023-09-29T04:44:52Z
kafkaversions.catalog.kubedb.com                  2023-09-29T04:41:42Z
mariadbautoscalers.autoscaling.kubedb.com         2023-09-29T04:44:18Z
mariadbdatabases.schema.kubedb.com                2023-09-29T04:44:26Z
mariadbopsrequests.ops.kubedb.com                 2023-09-29T04:45:01Z
mariadbs.kubedb.com                               2023-09-29T04:44:26Z
mariadbversions.catalog.kubedb.com                2023-09-29T04:41:42Z
memcacheds.kubedb.com                             2023-09-29T04:44:26Z
memcachedversions.catalog.kubedb.com              2023-09-29T04:41:42Z
mongodbautoscalers.autoscaling.kubedb.com         2023-09-29T04:44:18Z
mongodbdatabases.schema.kubedb.com                2023-09-29T04:44:20Z
mongodbopsrequests.ops.kubedb.com                 2023-09-29T04:44:29Z
mongodbs.kubedb.com                               2023-09-29T04:44:21Z
mongodbversions.catalog.kubedb.com                2023-09-29T04:41:42Z
mysqlautoscalers.autoscaling.kubedb.com           2023-09-29T04:44:19Z
mysqldatabases.schema.kubedb.com                  2023-09-29T04:44:17Z
mysqlopsrequests.ops.kubedb.com                   2023-09-29T04:44:58Z
mysqls.kubedb.com                                 2023-09-29T04:44:18Z
mysqlversions.catalog.kubedb.com                  2023-09-29T04:41:43Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-09-29T04:44:19Z
perconaxtradbopsrequests.ops.kubedb.com           2023-09-29T04:45:15Z
perconaxtradbs.kubedb.com                         2023-09-29T04:44:46Z
perconaxtradbversions.catalog.kubedb.com          2023-09-29T04:41:43Z
pgbouncers.kubedb.com                             2023-09-29T04:44:47Z
pgbouncerversions.catalog.kubedb.com              2023-09-29T04:41:44Z
postgresautoscalers.autoscaling.kubedb.com        2023-09-29T04:44:19Z
postgresdatabases.schema.kubedb.com               2023-09-29T04:44:24Z
postgreses.kubedb.com                             2023-09-29T04:44:25Z
postgresopsrequests.ops.kubedb.com                2023-09-29T04:45:09Z
postgresversions.catalog.kubedb.com               2023-09-29T04:41:44Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-09-29T04:44:20Z
proxysqlopsrequests.ops.kubedb.com                2023-09-29T04:45:12Z
proxysqls.kubedb.com                              2023-09-29T04:44:50Z
proxysqlversions.catalog.kubedb.com               2023-09-29T04:41:44Z
publishers.postgres.kubedb.com                    2023-09-29T04:45:26Z
redisautoscalers.autoscaling.kubedb.com           2023-09-29T04:44:21Z
redises.kubedb.com                                2023-09-29T04:44:50Z
redisopsrequests.ops.kubedb.com                   2023-09-29T04:45:05Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-09-29T04:44:21Z
redissentinelopsrequests.ops.kubedb.com           2023-09-29T04:45:19Z
redissentinels.kubedb.com                         2023-09-29T04:44:51Z
redisversions.catalog.kubedb.com                  2023-09-29T04:41:45Z
subscribers.postgres.kubedb.com                   2023-09-29T04:45:29Z
```

## Deploy MySQL Cluster

Now we are going to deploy MySQL cluster using KubeDB. First, let’s create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MySQL we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-cluster
  namespace: demo
spec:
  version: "5.7.41"
  replicas: 3
  topology:
    mode: GroupReplication
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

Let's save this yaml configuration into `mysql-cluster.yaml` 
Then create the above MySQL CRD

```bash
$ kubectl apply -f mysql-cluster.yaml
mysql.kubedb.com/mysql-cluster created
```

In this yaml,
* `spec.version` field specifies the version of MySQL. Here, we are using MySQL `version 5.7.41`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).

Once these are handled correctly and the MySQL object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                  READY   STATUS    RESTARTS   AGE
pod/mysql-cluster-0   2/2     Running   0          2m11s
pod/mysql-cluster-1   2/2     Running   0          101s
pod/mysql-cluster-2   2/2     Running   0          93s

NAME                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/mysql-cluster           ClusterIP   10.96.11.213   <none>        3306/TCP   2m15s
service/mysql-cluster-pods      ClusterIP   None           <none>        3306/TCP   2m15s
service/mysql-cluster-standby   ClusterIP   10.96.179.52   <none>        3306/TCP   2m15s

NAME                             READY   AGE
statefulset.apps/mysql-cluster   3/3     2m11s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-cluster   kubedb.com/mysql   5.7.41    2m11s

NAME                             VERSION   STATUS   AGE
mysql.kubedb.com/mysql-cluster   5.7.41    Ready    2m15s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo mysql-cluster
NAME            VERSION   STATUS   AGE
mysql-cluster   5.7.41    Ready    2m34s
```
> We have successfully deployed MySQL in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mysql-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                 TYPE                       DATA   AGE
mysql-cluster-auth   kubernetes.io/basic-auth   2      3m58s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mysql-cluster           ClusterIP   10.96.121.31   <none>        3306/TCP   4m25s
mysql-cluster-pods      ClusterIP   None           <none>        3306/TCP   4m25s
mysql-cluster-standby   ClusterIP   10.96.82.183   <none>        3306/TCP   4m25s
```
Now, we are going to use `mysql-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.password}' | base64 -d
fno!sZHPYzBLOwf1
```

#### Insert Sample Data

In this section, we are going to login into our MySQL pod and insert some sample data.

```bash
$ kubectl exec -it mysql-cluster-0 -n demo -c mysql -- bash

root@sample-mysql-0:/# mysql --user=root --password='fno!sZHPYzBLOwf1'
Welcome to the MySQL monitor.  Commands end with ; or \g.

mysql> CREATE DATABASE Music;
Query OK, 1 row affected (0.02 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.05 sec)

mysql> INSERT INTO Music.Artist (Name, Song) VALUES ("Bon Jovi", "It's My Life");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM Music.Artist;
+----------+--------------+
| Name     | Song         |
+----------+--------------+
| Bon Jovi | It's My Life |
+----------+--------------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

> We’ve successfully inserted some sample data to our database. More information about Run & Manage MySQL on Kubernetes can be found in [MySQL Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

## Update MySQL Database Version

In this section, we will update our MySQL version from `5.7.41` to the latest version `8.0.32`. Let's check the current version,

```bash
$ kubectl get mysql -n demo mysql-cluster -o=jsonpath='{.spec.version}{"\n"}'
5.7.41
```

### Create MySQLOpsRequest

In order to update the version of MySQL cluster, we have to create a `MySQLOpsRequest` CR with your desired version that is supported by KubeDB. Below is the YAML of the `MySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mysql-cluster
  updateVersion:
    targetVersion: "8.0.32"
```

Let's save this yaml configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
mysqlopsrequest.ops.kubedb.com/update-version created
```

In this yaml,
* `spec.databaseRef.name` specifies that we are performing operation on `mysql-cluster` MySQL database.
* `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
* `spec.updateVersion.targetVersion` specifies the expected version of the database `8.0.32`.

### Verify the Updated MySQL Version

`KubeDB` Enterprise operator will update the image of MySQL object and related `StatefulSets` and `Pods`.
Let’s wait for `MySQLOpsRequest` to be Successful. Run the following command to check `MySQLOpsRequest` CR,

```bash
$ kubectl get mysqlopsrequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   5m12s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded.
Now, we are going to verify whether the MySQL and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get mysql -n demo mysql-cluster -o=jsonpath='{.spec.version}{"\n"}'
8.0.32
```

> You can see from above, our MySQL database has been updated with the new version `8.0.32`. So, the database update process is successfully completed.


If you want to learn more about Production-Grade MySQL you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=jWPGnPZx0zPYoAQ-&amp;list=PLoiT1Gv2KR1gNPaHZtfdBZb6G4wLx6Iks" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MySQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).