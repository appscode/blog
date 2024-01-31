---
title: Deploy MariaDB Galera Cluster in Azure Kubernetes Service (AKS)
date: "2024-01-26"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- dbaas
- galera-cluster
- kubedb
- kubernetes
- mariadb
- microsoft-azure
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, Kafka, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). MariaDB Galera Cluster is a [virtually synchronous](https://mariadb.com/kb/en/about-galera-replication/#synchronous-vs-asynchronous-replication) multi-master cluster for MariaDB. The Server replicates a transaction at commit time by broadcasting the write set associated with the transaction to every node in the cluster. The client connects directly to the DBMS and experiences behavior that is similar to native MariaDB in most cases. The wsrep API (write set replication API) defines the interface between Galera replication and MariaDB.
In this tutorial we will deploy MariaDB Galera Cluster in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MariaDB Galera Cluster
3) Read/Write Sample Data

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

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
appscode/kubedb                   	v2023.12.28  	v2023.12.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.25.0      	v0.25.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.12.28  	v2023.12.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.12.28  	v2023.12.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.16.0      	v0.16.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.12.28  	v2023.12.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2023.12.28  	v2023.12.28	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2023.12.28  	v2023.12.28	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.27.0      	v0.27.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.12.28  	v2023.12.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.40.0      	v0.40.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.16.0      	v0.16.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.12.20  	0.6.1      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.16.0      	v0.16.0    	KubeDB Webhook Server by AppsCode   

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2023.12.28 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS        AGE
kubedb      kubedb-kubedb-autoscaler-58d9c78d74-2h9ml       1/1     Running   0               3m54s
kubedb      kubedb-kubedb-dashboard-9d45fdc88-6fbnm         1/1     Running   0               3m54s
kubedb      kubedb-kubedb-ops-manager-59b75cb5b-p8rcv       1/1     Running   0               3m54s
kubedb      kubedb-kubedb-provisioner-54568fcf99-lhch8      1/1     Running   0               3m54s
kubedb      kubedb-kubedb-webhook-server-7c869d7b9f-bvbkd   1/1     Running   0               3m54s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2024-01-26T05:40:44Z
elasticsearchdashboards.dashboard.kubedb.com      2024-01-26T05:40:44Z
elasticsearches.kubedb.com                        2024-01-26T05:40:44Z
elasticsearchopsrequests.ops.kubedb.com           2024-01-26T05:40:57Z
elasticsearchversions.catalog.kubedb.com          2024-01-26T05:38:34Z
etcds.kubedb.com                                  2024-01-26T05:40:57Z
etcdversions.catalog.kubedb.com                   2024-01-26T05:38:35Z
kafkaopsrequests.ops.kubedb.com                   2024-01-26T05:41:57Z
kafkas.kubedb.com                                 2024-01-26T05:41:10Z
kafkaversions.catalog.kubedb.com                  2024-01-26T05:38:35Z
mariadbautoscalers.autoscaling.kubedb.com         2024-01-26T05:40:44Z
mariadbopsrequests.ops.kubedb.com                 2024-01-26T05:41:34Z
mariadbs.kubedb.com                               2024-01-26T05:40:58Z
mariadbversions.catalog.kubedb.com                2024-01-26T05:38:35Z
memcacheds.kubedb.com                             2024-01-26T05:40:59Z
memcachedversions.catalog.kubedb.com              2024-01-26T05:38:36Z
mongodbarchivers.archiver.kubedb.com              2024-01-26T05:41:13Z
mongodbautoscalers.autoscaling.kubedb.com         2024-01-26T05:40:45Z
mongodbopsrequests.ops.kubedb.com                 2024-01-26T05:41:02Z
mongodbs.kubedb.com                               2024-01-26T05:41:00Z
mongodbversions.catalog.kubedb.com                2024-01-26T05:38:36Z
mysqlarchivers.archiver.kubedb.com                2024-01-26T05:41:15Z
mysqlautoscalers.autoscaling.kubedb.com           2024-01-26T05:40:46Z
mysqlopsrequests.ops.kubedb.com                   2024-01-26T05:41:31Z
mysqls.kubedb.com                                 2024-01-26T05:41:03Z
mysqlversions.catalog.kubedb.com                  2024-01-26T05:38:36Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2024-01-26T05:40:46Z
perconaxtradbopsrequests.ops.kubedb.com           2024-01-26T05:41:49Z
perconaxtradbs.kubedb.com                         2024-01-26T05:41:04Z
perconaxtradbversions.catalog.kubedb.com          2024-01-26T05:38:36Z
pgbouncers.kubedb.com                             2024-01-26T05:41:05Z
pgbouncerversions.catalog.kubedb.com              2024-01-26T05:38:37Z
postgresarchivers.archiver.kubedb.com             2024-01-26T05:41:17Z
postgresautoscalers.autoscaling.kubedb.com        2024-01-26T05:40:46Z
postgreses.kubedb.com                             2024-01-26T05:41:06Z
postgresopsrequests.ops.kubedb.com                2024-01-26T05:41:42Z
postgresversions.catalog.kubedb.com               2024-01-26T05:38:37Z
proxysqlautoscalers.autoscaling.kubedb.com        2024-01-26T05:40:47Z
proxysqlopsrequests.ops.kubedb.com                2024-01-26T05:41:45Z
proxysqls.kubedb.com                              2024-01-26T05:41:07Z
proxysqlversions.catalog.kubedb.com               2024-01-26T05:38:38Z
publishers.postgres.kubedb.com                    2024-01-26T05:42:00Z
redisautoscalers.autoscaling.kubedb.com           2024-01-26T05:40:48Z
redises.kubedb.com                                2024-01-26T05:41:08Z
redisopsrequests.ops.kubedb.com                   2024-01-26T05:41:37Z
redissentinelautoscalers.autoscaling.kubedb.com   2024-01-26T05:40:48Z
redissentinelopsrequests.ops.kubedb.com           2024-01-26T05:41:52Z
redissentinels.kubedb.com                         2024-01-26T05:41:09Z
redisversions.catalog.kubedb.com                  2024-01-26T05:38:38Z
subscribers.postgres.kubedb.com                   2024-01-26T05:42:03Z
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
  version: "11.2.2"
  replicas: 3
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

Let's save this yaml configuration into `galera-cluster.yaml` 
Then create the above MariaDB CRO

```bash
$ kubectl apply -f galera-cluster.yaml
mariadb.kubedb.com/galera-cluster created
```
In this yaml,
* `spec.version` field specifies the version of MariaDB Here, we are using MariaDB `version 11.2.2`. You can list the KubeDB supported versions of MariaDB by running `$ kubectl get mariadbversion` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/mariadb/concepts/mariadb/#specterminationpolicy).

Once these are handled correctly and the MariaDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                   READY   STATUS    RESTARTS   AGE
pod/galera-cluster-0   2/2     Running   0          2m56s
pod/galera-cluster-1   2/2     Running   0          2m56s
pod/galera-cluster-2   2/2     Running   0          2m55s

NAME                          TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/galera-cluster        ClusterIP   10.76.5.4    <none>        3306/TCP   2m57s
service/galera-cluster-pods   ClusterIP   None         <none>        3306/TCP   2m57s

NAME                              READY   AGE
statefulset.apps/galera-cluster   3/3     2m57s

NAME                                                TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/galera-cluster   kubedb.com/mariadb   11.2.2    2m58s

NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/galera-cluster   11.2.2    Ready    3m1s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo galera-cluster
NAME             VERSION   STATUS   AGE
galera-cluster   11.2.2    Ready    3m22s
```
> We have successfully deployed MariaDB Galera Cluster in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create `Secret` and `Service` for the database `galera-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE                       DATA   AGE
galera-cluster-auth   kubernetes.io/basic-auth   2      3m58s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
galera-cluster        ClusterIP   10.76.5.4    <none>        3306/TCP   4m14s
galera-cluster-pods   ClusterIP   None         <none>        3306/TCP   4m14s
```
Now, we are going to use `galera-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.password}' | base64 -d
qMTNlxhj*wM~Q1n7
```

### Data Availability

In a MariaDB Galera Cluster, Each member can read and write. Now, we will insert data from any of the available nodes, and we will see whether we can get the data from every other members.

### Insert Sample Data

In this section, we are going to login into our MariaDB database pod and insert some sample data. 

```bash
$ kubectl exec -it galera-cluster-0 -n demo -c mariadb -- bash

mysql@galera-cluster-0:/$ mariadb --user=root --password='qMTNlxhj*wM~Q1n7'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.015 sec)

MariaDB [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.014 sec)

MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Country Roads");
Query OK, 1 row affected (0.004 sec)

MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+---------------+
| id | Name        | Song          |
+----+-------------+---------------+
|  1 | John Denver | Country Roads |
+----+-------------+---------------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
...

$ kubectl exec -it galera-cluster-1 -n demo -c mariadb -- bash
mysql@galera-cluster-1:/$ mariadb --user=root --password='qMTNlxhj*wM~Q1n7'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Read data from Node 2
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+---------------+
| id | Name        | Song          |
+----+-------------+---------------+
|  1 | John Denver | Country Roads |
+----+-------------+---------------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
...

$ kubectl exec -it galera-cluster-2 -n demo -c mariadb -- bash
root@galera-cluster-2:/# mariadb --user=root --password='qMTNlxhj*wM~Q1n7'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.


# Insert data from Node 3
MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Avicii", "The Nights");
Query OK, 1 row affected (0.004 sec)

# Read data from Node 3
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+---------------+
| id | Name        | Song          |
+----+-------------+---------------+
|  1 | John Denver | Country Roads |
|  2 | Avicii      | The Nights    |
+----+-------------+---------------+
2 rows in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

> We've successfully inserted some sample data to our database. Also, access it from every node of galera cluster. More information about Run & Manage Production-Grade MariaDB Database on Kubernetes can be found in [MariaDB Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)


If you want to learn more about Production-Grade MariaDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=CW-3F1P1FKlAnfZ9&amp;list=PLoiT1Gv2KR1gbxvKO8IYz5_MK4FhZp4R-" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MariaDB on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
