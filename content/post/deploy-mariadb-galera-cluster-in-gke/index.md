---
title: Deploy MariaDB Galera Cluster in Google Kubernetes Engine (GKE)
date: "2024-03-20"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- galera-cluster
- gcp
- gke
- kubedb
- kubernetes
- mariadb
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). MariaDB Galera Cluster is a [virtually synchronous](https://mariadb.com/kb/en/about-galera-replication/#synchronous-vs-asynchronous-replication) multi-master cluster for MariaDB. The Server replicates a transaction at commit time by broadcasting the write set associated with the transaction to every node in the cluster. The client connects directly to the DBMS and experiences behavior that is similar to native MariaDB in most cases. The wsrep API (write set replication API) defines the interface between Galera replication and MariaDB.
In this tutorial we will deploy MariaDB Galera Cluster in Google Kubernetes Engine (GKE). We will cover the following steps:

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
appscode/kubedb                   	v2024.3.16   	v2024.3.16 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.29.0      	v0.29.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.3.16   	v2024.3.16 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.8       	v0.0.8     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.3.16   	v2024.3.16 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.20.0      	v0.20.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.3.16   	v2024.3.16 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.3.16   	v2024.3.16 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.3.16   	v2024.3.16 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.31.0      	v0.31.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.3.16   	v2024.3.16 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.44.0      	v0.44.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.20.0      	v0.20.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.2.13   	0.6.4      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.20.0      	v0.20.0    	KubeDB Webhook Server by AppsCode  

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.3.16 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-575f984478-wx4d7       1/1     Running   0          2m41s
kubedb      kubedb-kubedb-ops-manager-6fc8fd6b6c-hhw2q      1/1     Running   0          2m41s
kubedb      kubedb-kubedb-provisioner-7cdfdfb55f-6dx22      1/1     Running   0          2m41s
kubedb      kubedb-kubedb-webhook-server-59965c78c4-brhqh   1/1     Running   0          2m41s
kubedb      kubedb-petset-operator-5d94b4ddb8-ts67d         1/1     Running   0          2m41s
kubedb      kubedb-petset-webhook-server-d49599c45-zhvvc    2/2     Running   0          2m41s
kubedb      kubedb-sidekick-5dc87959b7-j44lm                1/1     Running   0          2m41s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-03-20T10:23:47Z
connectors.kafka.kubedb.com                        2024-03-20T10:23:47Z
druidversions.catalog.kubedb.com                   2024-03-20T10:23:03Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-03-20T10:23:44Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-03-20T10:23:44Z
elasticsearches.kubedb.com                         2024-03-20T10:23:44Z
elasticsearchopsrequests.ops.kubedb.com            2024-03-20T10:23:44Z
elasticsearchversions.catalog.kubedb.com           2024-03-20T10:23:03Z
etcdversions.catalog.kubedb.com                    2024-03-20T10:23:03Z
ferretdbversions.catalog.kubedb.com                2024-03-20T10:23:03Z
kafkaautoscalers.autoscaling.kubedb.com            2024-03-20T10:23:47Z
kafkaconnectorversions.catalog.kubedb.com          2024-03-20T10:23:03Z
kafkaopsrequests.ops.kubedb.com                    2024-03-20T10:23:47Z
kafkas.kubedb.com                                  2024-03-20T10:23:47Z
kafkaversions.catalog.kubedb.com                   2024-03-20T10:23:03Z
mariadbarchivers.archiver.kubedb.com               2024-03-20T10:23:51Z
mariadbautoscalers.autoscaling.kubedb.com          2024-03-20T10:23:50Z
mariadbdatabases.schema.kubedb.com                 2024-03-20T10:23:50Z
mariadbopsrequests.ops.kubedb.com                  2024-03-20T10:23:50Z
mariadbs.kubedb.com                                2024-03-20T10:23:50Z
mariadbversions.catalog.kubedb.com                 2024-03-20T10:23:03Z
memcachedversions.catalog.kubedb.com               2024-03-20T10:23:03Z
mongodbarchivers.archiver.kubedb.com               2024-03-20T10:23:54Z
mongodbautoscalers.autoscaling.kubedb.com          2024-03-20T10:23:54Z
mongodbdatabases.schema.kubedb.com                 2024-03-20T10:23:54Z
mongodbopsrequests.ops.kubedb.com                  2024-03-20T10:23:54Z
mongodbs.kubedb.com                                2024-03-20T10:23:54Z
mongodbversions.catalog.kubedb.com                 2024-03-20T10:23:03Z
mysqlarchivers.archiver.kubedb.com                 2024-03-20T10:23:58Z
mysqlautoscalers.autoscaling.kubedb.com            2024-03-20T10:23:58Z
mysqldatabases.schema.kubedb.com                   2024-03-20T10:23:58Z
mysqlopsrequests.ops.kubedb.com                    2024-03-20T10:23:58Z
mysqls.kubedb.com                                  2024-03-20T10:23:58Z
mysqlversions.catalog.kubedb.com                   2024-03-20T10:23:03Z
perconaxtradbversions.catalog.kubedb.com           2024-03-20T10:23:03Z
pgbouncerversions.catalog.kubedb.com               2024-03-20T10:23:03Z
pgpoolversions.catalog.kubedb.com                  2024-03-20T10:23:03Z
postgresarchivers.archiver.kubedb.com              2024-03-20T10:24:01Z
postgresautoscalers.autoscaling.kubedb.com         2024-03-20T10:24:01Z
postgresdatabases.schema.kubedb.com                2024-03-20T10:24:01Z
postgreses.kubedb.com                              2024-03-20T10:24:01Z
postgresopsrequests.ops.kubedb.com                 2024-03-20T10:24:01Z
postgresversions.catalog.kubedb.com                2024-03-20T10:23:03Z
proxysqlversions.catalog.kubedb.com                2024-03-20T10:23:03Z
publishers.postgres.kubedb.com                     2024-03-20T10:24:02Z
rabbitmqversions.catalog.kubedb.com                2024-03-20T10:23:03Z
redisautoscalers.autoscaling.kubedb.com            2024-03-20T10:24:05Z
redises.kubedb.com                                 2024-03-20T10:24:05Z
redisopsrequests.ops.kubedb.com                    2024-03-20T10:24:05Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-03-20T10:24:05Z
redissentinelopsrequests.ops.kubedb.com            2024-03-20T10:24:05Z
redissentinels.kubedb.com                          2024-03-20T10:24:05Z
redisversions.catalog.kubedb.com                   2024-03-20T10:23:03Z
singlestoreversions.catalog.kubedb.com             2024-03-20T10:23:03Z
solrversions.catalog.kubedb.com                    2024-03-20T10:23:03Z
subscribers.postgres.kubedb.com                    2024-03-20T10:24:02Z
zookeeperversions.catalog.kubedb.com               2024-03-20T10:23:03Z
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
    storageClassName: "standard"
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
pod/galera-cluster-0   2/2     Running   0          2m15s
pod/galera-cluster-1   2/2     Running   0          2m15s
pod/galera-cluster-2   2/2     Running   0          2m15s

NAME                          TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/galera-cluster        ClusterIP   10.96.35.48   <none>        3306/TCP   2m18s
service/galera-cluster-pods   ClusterIP   None          <none>        3306/TCP   2m18s

NAME                              READY   AGE
statefulset.apps/galera-cluster   3/3     2m15s

NAME                                                TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/galera-cluster   kubedb.com/mariadb   11.2.2    2m15s

NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/galera-cluster   11.2.2    Ready    2m18s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo galera-cluster
NAME             VERSION   STATUS   AGE
galera-cluster   11.2.2    Ready    2m39s
```
> We have successfully deployed MariaDB Galera Cluster in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create `Secret` and `Service` for the database `galera-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE                       DATA   AGE
galera-cluster-auth   kubernetes.io/basic-auth   2      2m58s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=galera-cluster
NAME                  TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
galera-cluster        ClusterIP   10.96.35.48   <none>        3306/TCP   3m10s
galera-cluster-pods   ClusterIP   None          <none>        3306/TCP   3m10s
```
Now, we are going to use `galera-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo galera-cluster-auth -o jsonpath='{.data.password}' | base64 -d
RJSzcs7XsFeAMHTt
```

### Data Availability

In a MariaDB Galera Cluster, Each member can read and write. Now, we will insert data from any of the available nodes, and we will see whether we can get the data from every other members.

### Insert Sample Data

In this section, we are going to login into our MariaDB database pod and insert some sample data. 

```bash
$ kubectl exec -it galera-cluster-0 -n demo -c mariadb -- bash

mysql@galera-cluster-0:/$ mariadb --user=root --password='RJSzcs7XsFeAMHTt'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.013 sec)

MariaDB [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.011 sec)

MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Bobby Bare", "Five Hundred Miles");
Query OK, 1 row affected (0.006 sec)

MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+------------+--------------------+
| id | Name       | Song               |
+----+------------+--------------------+
|  1 | Bobby Bare | Five Hundred Miles |
+----+------------+--------------------+
1 row in set (0.002 sec)

MariaDB [(none)]> exit
Bye
...

$ kubectl exec -it galera-cluster-1 -n demo -c mariadb -- bash
mysql@galera-cluster-1:/$ mariadb --user=root --password='RJSzcs7XsFeAMHTt'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Read data from Node 2
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+------------+--------------------+
| id | Name       | Song               |
+----+------------+--------------------+
|  1 | Bobby Bare | Five Hundred Miles |
+----+------------+--------------------+
1 row in set (0.003 sec)

MariaDB [(none)]> exit
Bye
...

$ kubectl exec -it galera-cluster-2 -n demo -c mariadb -- bash
root@galera-cluster-2:/# mariadb --user=root --password='RJSzcs7XsFeAMHTt'

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.


# Insert data from Node 3
MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Bon Jovi", "It's My Life");
Query OK, 1 row affected (0.004 sec)

# Read data from Node 3
MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+--------------------+
| id | Name        | Song               |
+----+-------------+--------------------+
|  1 | Bobby Bare  | Five Hundred Miles |
|  2 | Bon Jovi    | It's My Life       |
+----+-------------+--------------------+
2 rows in set (0.002 sec)

MariaDB [(none)]> exit
Bye
```

> We've successfully inserted some sample data to our database. Also, access it from every node of galera cluster. More information about Deploy & Manage Production-Grade MariaDB Database on Kubernetes can be found in [MariaDB Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)


If you want to learn more about Production-Grade MariaDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=CW-3F1P1FKlAnfZ9&amp;list=PLoiT1Gv2KR1gbxvKO8IYz5_MK4FhZp4R-" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MariaDB on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
