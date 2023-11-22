---
title: Deploy MariaDB Galera Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-11-20"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- galera-cluster
- kubedb
- kubernetes
- mariadb
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, Kafka, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). MariaDB Galera Cluster is a [virtually synchronous](https://mariadb.com/kb/en/about-galera-replication/#synchronous-vs-asynchronous-replication) multi-master cluster for MariaDB. The Server replicates a transaction at commit time by broadcasting the write set associated with the transaction to every node in the cluster. The client connects directly to the DBMS and experiences behavior that is similar to native MariaDB in most cases. The wsrep API (write set replication API) defines the interface between Galera replication and MariaDB.
In this tutorial we will deploy MariaDB Galera Cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MariaDB Galera Cluster
3) Read/Write Sample Data

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
appscode/kubedb                   	v2023.11.2   	v2023.11.2 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.22.0      	v0.22.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.11.2   	v2023.11.2 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.11.2   	v2023.11.2 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.13.0      	v0.13.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.11.2   	v2023.11.2 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.11.2   	v2023.11.2 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.11.2   	v2023.11.2 	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.24.0      	v0.24.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.11.2   	v2023.11.2 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v0.0.1       	v0.0.1     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v0.0.1       	v0.0.1     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v0.0.1       	v0.0.1     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.37.0      	v0.37.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.13.0      	v0.13.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.11.14  	0.5.0      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.13.0      	v0.13.0    	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.11.2 \
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
kubedb      kubedb-kubedb-autoscaler-5dbd6f5fdf-dsbzm       1/1     Running   0          4m
kubedb      kubedb-kubedb-dashboard-6f5597f98-6nn4t         1/1     Running   0          4m
kubedb      kubedb-kubedb-ops-manager-555cb7f784-x8f79      1/1     Running   0          4m
kubedb      kubedb-kubedb-provisioner-5d5f6d899-5jmnr       1/1     Running   0          4m
kubedb      kubedb-kubedb-schema-manager-5b8fd86697-m5kc6   1/1     Running   0          4m
kubedb      kubedb-kubedb-webhook-server-6d44647846-fgmdg   1/1     Running   0          4m
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-11-20T06:34:38Z
elasticsearchdashboards.dashboard.kubedb.com      2023-11-20T06:34:47Z
elasticsearches.kubedb.com                        2023-11-20T06:34:47Z
elasticsearchopsrequests.ops.kubedb.com           2023-11-20T06:35:09Z
elasticsearchversions.catalog.kubedb.com          2023-11-20T06:34:11Z
etcds.kubedb.com                                  2023-11-20T06:35:34Z
etcdversions.catalog.kubedb.com                   2023-11-20T06:34:12Z
kafkaopsrequests.ops.kubedb.com                   2023-11-20T06:35:43Z
kafkas.kubedb.com                                 2023-11-20T06:35:35Z
kafkaversions.catalog.kubedb.com                  2023-11-20T06:34:12Z
mariadbautoscalers.autoscaling.kubedb.com         2023-11-20T06:34:38Z
mariadbdatabases.schema.kubedb.com                2023-11-20T06:35:20Z
mariadbopsrequests.ops.kubedb.com                 2023-11-20T06:35:22Z
mariadbs.kubedb.com                               2023-11-20T06:35:20Z
mariadbversions.catalog.kubedb.com                2023-11-20T06:34:12Z
memcacheds.kubedb.com                             2023-11-20T06:35:34Z
memcachedversions.catalog.kubedb.com              2023-11-20T06:34:12Z
mongodbautoscalers.autoscaling.kubedb.com         2023-11-20T06:34:38Z
mongodbdatabases.schema.kubedb.com                2023-11-20T06:35:19Z
mongodbopsrequests.ops.kubedb.com                 2023-11-20T06:35:13Z
mongodbs.kubedb.com                               2023-11-20T06:35:13Z
mongodbversions.catalog.kubedb.com                2023-11-20T06:34:12Z
mysqlautoscalers.autoscaling.kubedb.com           2023-11-20T06:34:38Z
mysqldatabases.schema.kubedb.com                  2023-11-20T06:35:19Z
mysqlopsrequests.ops.kubedb.com                   2023-11-20T06:35:19Z
mysqls.kubedb.com                                 2023-11-20T06:35:19Z
mysqlversions.catalog.kubedb.com                  2023-11-20T06:34:12Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-11-20T06:34:38Z
perconaxtradbopsrequests.ops.kubedb.com           2023-11-20T06:35:36Z
perconaxtradbs.kubedb.com                         2023-11-20T06:35:35Z
perconaxtradbversions.catalog.kubedb.com          2023-11-20T06:34:12Z
pgbouncers.kubedb.com                             2023-11-20T06:35:16Z
pgbouncerversions.catalog.kubedb.com              2023-11-20T06:34:12Z
postgresautoscalers.autoscaling.kubedb.com        2023-11-20T06:34:38Z
postgresdatabases.schema.kubedb.com               2023-11-20T06:35:20Z
postgreses.kubedb.com                             2023-11-20T06:35:20Z
postgresopsrequests.ops.kubedb.com                2023-11-20T06:35:29Z
postgresversions.catalog.kubedb.com               2023-11-20T06:34:12Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-11-20T06:34:39Z
proxysqlopsrequests.ops.kubedb.com                2023-11-20T06:35:33Z
proxysqls.kubedb.com                              2023-11-20T06:35:33Z
proxysqlversions.catalog.kubedb.com               2023-11-20T06:34:12Z
publishers.postgres.kubedb.com                    2023-11-20T06:35:46Z
redisautoscalers.autoscaling.kubedb.com           2023-11-20T06:34:39Z
redises.kubedb.com                                2023-11-20T06:35:25Z
redisopsrequests.ops.kubedb.com                   2023-11-20T06:35:25Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-11-20T06:34:39Z
redissentinelopsrequests.ops.kubedb.com           2023-11-20T06:35:39Z
redissentinels.kubedb.com                         2023-11-20T06:35:35Z
redisversions.catalog.kubedb.com                  2023-11-20T06:34:12Z
subscribers.postgres.kubedb.com                   2023-11-20T06:35:49Z
```

## Deploy MariaDB Galera Cluster

Now, we are going to Deploy MariaDB using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MariaDB CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: galera-cluster
  namespace: demo
spec:
  version: "10.11.2"
  replicas: 3
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

Let's save this yaml configuration into `galera-cluster.yaml` 
Then create the above MariaDB CRO

```bash
$ kubectl apply -f galera-cluster.yaml
mariadb.kubedb.com/galera-cluster created
```
In this yaml,
* `spec.version` field specifies the version of MariaDB Here, we are using MariaDB `version 10.11.2`. You can list the KubeDB supported versions of MariaDB by running `$ kubectl get mariadbversion` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/mariadb/concepts/mariadb/#specterminationpolicy).

Once these are handled correctly and the MariaDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                   READY   STATUS    RESTARTS   AGE
pod/galera-cluster-0   2/2     Running   0          2m18s
pod/galera-cluster-1   2/2     Running   0          2m18s
pod/galera-cluster-2   2/2     Running   0          2m18s

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/galera-cluster        ClusterIP   10.96.109.158   <none>        3306/TCP   2m22s
service/galera-cluster-pods   ClusterIP   None            <none>        3306/TCP   2m22s

NAME                              READY   AGE
statefulset.apps/galera-cluster   3/3     2m18s

NAME                                                TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/galera-cluster   kubedb.com/mariadb   10.11.2   2m18s

NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/galera-cluster   10.11.2   Ready    2m22s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo galera-cluster
NAME             VERSION   STATUS   AGE
galera-cluster   10.11.2   Ready    2m48s
```
> We have successfully deployed MariaDB Galera Cluster in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create `Secret` and `Service` for the database `galera-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE                       DATA   AGE
galera-cluster-auth   kubernetes.io/basic-auth   2      3m26s


$ kubectl get service -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
galera-cluster        ClusterIP   10.96.109.158   <none>        3306/TCP   4m7s
galera-cluster-pods   ClusterIP   None            <none>        3306/TCP   4m7s
```
Now, we are going to use `galera-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.password}' | base64 -d
ndhijJGM)cvy26gr
```

### Data Availability

In a MariaDB Galera Cluster, Each member can read and write. Now, we will insert data from any of the available nodes, and we will see whether we can get the data from every other members.

### Insert Sample Data

In this section, we are going to login into our MariaDB database pod and insert some sample data. 

```bash
$ kubectl exec -it galera-cluster-0 -n demo -c mariadb -- bash
root@galera-cluster-0:/# mysql --user=root --password='ndhijJGM)cvy26gr'

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.005 sec)

MariaDB [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.008 sec)

MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Bobby Bare", "Five Hundred Miles");
Query OK, 1 row affected (0.001 sec)

MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+------------+--------------------+
| id | Name       | Song               |
+----+------------+--------------------+
|  1 | Bobby Bare | Five Hundred Miles |
+----+------------+--------------------+
1 row in set (0.000 sec)

MariaDB [(none)]> exit
Bye

...

$ kubectl exec -it galera-cluster-1 -n demo -c mariadb -- bash
root@galera-cluster-1:/# mysql --user=root --password='ndhijJGM)cvy26gr'

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Read data from Node 2
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+------------+--------------------+
| id | Name       | Song               |
+----+------------+--------------------+
|  1 | Bobby Bare | Five Hundred Miles |
+----+------------+--------------------+
1 row in set (0.000 sec)

MariaDB [(none)]> exit
Bye

...

$ kubectl exec -it galera-cluster-2 -n demo -c mariadb -- bash
root@galera-cluster-2:/# mysql --user=root --password='ndhijJGM)cvy26gr'

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Insert data from Node 3
MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Annie's Song");
Query OK, 1 row affected (0.001 sec)

# Read data from Node 3
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+--------------------+
| id | Name        | Song               |
+----+-------------+--------------------+
|  1 | Bobby Bare  | Five Hundred Miles |
|  3 | John Denver | Annie's Song       |
+----+-------------+--------------------+
2 rows in set (0.000 sec)

MariaDB [(none)]> exit
Bye
```

> We've successfully inserted some sample data to our database. Also, access it from every node of galera cluster. More information about Run & Manage Production-Grade MariaDB Database on Kubernetes can be found in [MariaDB Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)


We have made an in depth tutorial on MariaDB Alerting and Multi-Tenancy Support by using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/P8l2v6-yCHU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MariaDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
