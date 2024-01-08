---
title: Update Version of MariaDB Database in Azure Kubernetes Service (AKS)
date: "2023-12-28"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- mariadb
- microsoft-azure
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will update version of MariaDB Database in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MariaDB Cluster
3) Insert Sample Data
4) Update MariaDB Database Version


### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial, we will use KubeDB.

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

NAMESPACE   NAME                                            READY   STATUS    RESTARTS      AGE
kubedb      kubedb-kubedb-autoscaler-7b77bf574d-tr5f4       1/1     Running   0             4m
kubedb      kubedb-kubedb-dashboard-7557b6575c-kj7kb        1/1     Running   0             4m
kubedb      kubedb-kubedb-ops-manager-848d9797d-j8bbg       1/1     Running   0             4m
kubedb      kubedb-kubedb-provisioner-789645569f-jhcwg      1/1     Running   0             4m
kubedb      kubedb-kubedb-webhook-server-5f86864dcf-868f2   1/1     Running   0             4m
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-12-28T06:05:49Z
elasticsearchdashboards.dashboard.kubedb.com      2023-12-28T06:05:47Z
elasticsearches.kubedb.com                        2023-12-28T06:05:48Z
elasticsearchopsrequests.ops.kubedb.com           2023-12-28T06:05:56Z
elasticsearchversions.catalog.kubedb.com          2023-12-28T06:04:01Z
etcds.kubedb.com                                  2023-12-28T06:05:56Z
etcdversions.catalog.kubedb.com                   2023-12-28T06:04:01Z
kafkaopsrequests.ops.kubedb.com                   2023-12-28T06:06:58Z
kafkas.kubedb.com                                 2023-12-28T06:06:12Z
kafkaversions.catalog.kubedb.com                  2023-12-28T06:04:01Z
mariadbautoscalers.autoscaling.kubedb.com         2023-12-28T06:05:50Z
mariadbopsrequests.ops.kubedb.com                 2023-12-28T06:06:34Z
mariadbs.kubedb.com                               2023-12-28T06:05:57Z
mariadbversions.catalog.kubedb.com                2023-12-28T06:04:02Z
memcacheds.kubedb.com                             2023-12-28T06:05:58Z
memcachedversions.catalog.kubedb.com              2023-12-28T06:04:02Z
mongodbarchivers.archiver.kubedb.com              2023-12-28T06:06:15Z
mongodbautoscalers.autoscaling.kubedb.com         2023-12-28T06:05:50Z
mongodbopsrequests.ops.kubedb.com                 2023-12-28T06:06:01Z
mongodbs.kubedb.com                               2023-12-28T06:06:00Z
mongodbversions.catalog.kubedb.com                2023-12-28T06:04:02Z
mysqlarchivers.archiver.kubedb.com                2023-12-28T06:06:19Z
mysqlautoscalers.autoscaling.kubedb.com           2023-12-28T06:05:51Z
mysqlopsrequests.ops.kubedb.com                   2023-12-28T06:06:26Z
mysqls.kubedb.com                                 2023-12-28T06:06:04Z
mysqlversions.catalog.kubedb.com                  2023-12-28T06:04:03Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-12-28T06:05:51Z
perconaxtradbopsrequests.ops.kubedb.com           2023-12-28T06:06:51Z
perconaxtradbs.kubedb.com                         2023-12-28T06:06:06Z
perconaxtradbversions.catalog.kubedb.com          2023-12-28T06:04:03Z
pgbouncers.kubedb.com                             2023-12-28T06:06:07Z
pgbouncerversions.catalog.kubedb.com              2023-12-28T06:04:04Z
postgresarchivers.archiver.kubedb.com             2023-12-28T06:06:22Z
postgresautoscalers.autoscaling.kubedb.com        2023-12-28T06:05:52Z
postgreses.kubedb.com                             2023-12-28T06:06:09Z
postgresopsrequests.ops.kubedb.com                2023-12-28T06:06:43Z
postgresversions.catalog.kubedb.com               2023-12-28T06:04:04Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-12-28T06:05:52Z
proxysqlopsrequests.ops.kubedb.com                2023-12-28T06:06:46Z
proxysqls.kubedb.com                              2023-12-28T06:06:09Z
proxysqlversions.catalog.kubedb.com               2023-12-28T06:04:04Z
publishers.postgres.kubedb.com                    2023-12-28T06:07:02Z
redisautoscalers.autoscaling.kubedb.com           2023-12-28T06:05:53Z
redises.kubedb.com                                2023-12-28T06:06:10Z
redisopsrequests.ops.kubedb.com                   2023-12-28T06:06:37Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-12-28T06:05:53Z
redissentinelopsrequests.ops.kubedb.com           2023-12-28T06:06:54Z
redissentinels.kubedb.com                         2023-12-28T06:06:11Z
redisversions.catalog.kubedb.com                  2023-12-28T06:04:05Z
subscribers.postgres.kubedb.com                   2023-12-28T06:07:06Z
```

## Deploy MariaDB Cluster

Now we are going to deploy MariaDB cluster using KubeDB. First, let’s create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MariaDB we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: mariadb-cluster
  namespace: demo
spec:
  version: "10.4.32"
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

Let's save this yaml configuration into `mariadb-cluster.yaml` 
Then create the above MariaDB CR

```bash
$ kubectl apply -f mariadb-cluster.yaml
mariadb.kubedb.com/mariadb-cluster created
```

In this yaml,
* `spec.version` field specifies the version of MariaDB. Here, we are using MariaDB `version 10.4.32`. You can list the KubeDB supported versions of MariaDB by running `$ kubectl get mariadbversion` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/mariadb/concepts/mariadb/).

Once these are handled correctly and the MariaDB object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                    READY   STATUS    RESTARTS   AGE
pod/mariadb-cluster-0   2/2     Running   0          3m37s
pod/mariadb-cluster-1   2/2     Running   0          3m37s
pod/mariadb-cluster-2   2/2     Running   0          3m36s

NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/mariadb-cluster        ClusterIP   10.76.3.111   <none>        3306/TCP   3m39s
service/mariadb-cluster-pods   ClusterIP   None          <none>        3306/TCP   3m39s

NAME                               READY   AGE
statefulset.apps/mariadb-cluster   3/3     3m39s

NAME                                                 TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mariadb-cluster   kubedb.com/mariadb   10.4.32   3m40s

NAME                                 VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-cluster   10.4.32   Ready    3m42s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo mariadb-cluster
NAME              VERSION   STATUS   AGE
mariadb-cluster   10.4.32   Ready    4m21s
```
> We have successfully deployed MariaDB in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mariadb-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mariadb-cluster
NAME                   TYPE                       DATA   AGE
mariadb-cluster-auth   kubernetes.io/basic-auth   2      5m3s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mariadb-cluster
NAME                   TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
mariadb-cluster        ClusterIP   10.76.3.111   <none>        3306/TCP   5m26s
mariadb-cluster-pods   ClusterIP   None          <none>        3306/TCP   5m26s
```
Now, we are going to use `mariadb-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mariadb-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mariadb-cluster-auth -o jsonpath='{.data.password}' | base64 -d
DQpZV~nO(0XMq2Vk
```

#### Insert Sample Data

In this section, we are going to login into our MariaDB pod and insert some sample data.

```bash
$ kubectl exec -it mariadb-cluster-0 -n demo -c mariadb -- bash

mysql@mariadb-cluster-0:/$ mysql --user=root --password='DQpZV~nO(0XMq2Vk'
Welcome to the MariaDB monitor.  Commands end with ; or \g.

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.008 sec)

MariaDB [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
+--------------------+
5 rows in set (0.001 sec)

MariaDB [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.023 sec)

MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("Avicii", "The Nights");
Query OK, 1 row affected (0.005 sec)

MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+--------+------------+
| id | Name   | Song       |
+----+--------+------------+
|  1 | Avicii | The Nights |
+----+--------+------------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

> We’ve successfully inserted some sample data to our database. More information about Run & Manage MariaDB on Kubernetes can be found in [MariaDB Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

## Update MariaDB Database Version

In this section, we will update our MariaDB version from `10.4.32` to the latest version `10.6.16`. Let's check the current version,

```bash
$ kubectl get mariadb -n demo mariadb-cluster -o=jsonpath='{.spec.version}{"\n"}'
10.4.32
```

### Create MariaDBOpsRequest

In order to update the version of MariaDB cluster, we have to create a `MariaDBOpsRequest` CR with your desired version that is supported by KubeDB. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mariadb-cluster
  updateVersion:
    targetVersion: "10.6.16"
```

Let's save this yaml configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
mariadbopsrequest.ops.kubedb.com/update-version created
```

In this yaml,
* `spec.databaseRef.name` specifies that we are performing operation on `mariadb-cluster` MariaDB database.
* `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
* `spec.updateVersion.targetVersion` specifies the expected version of the database `10.6.16`.

### Verify the Updated MariaDB Version

`KubeDB` operator will update the image of MariaDB and related `StatefulSets` and `Pods`.
Let’s wait for `MariaDBOpsRequest` to be Successful. Run the following command to check `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   4m12s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded.
Now, we are going to verify whether the MariaDB and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get mariadb -n demo mariadb-cluster -o=jsonpath='{.spec.version}{"\n"}'
10.6.16
```

> You can see from above, our MariaDB database has been updated with the new version `10.6.16`. So, the database update process is successfully completed.


If you want to learn more about Production-Grade MariaDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=gCYMnO60Bb_HgcfL&amp;list=PLoiT1Gv2KR1gbxvKO8IYz5_MK4FhZp4R-" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MariaDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).