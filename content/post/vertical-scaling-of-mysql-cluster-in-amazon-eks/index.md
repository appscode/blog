---
title: Vertical Scaling of MySQL Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-06-05"
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
- mysql-cluster
- mysql-database
- vertical-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Vertical scaling of MySQL cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1. Install KubeDB
2. Deploy MySQL Cluster
3. Read/Write Sample Data
4. Vertical Scaling of MySQL Cluster

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
appscode/kubedb-ui                	v2024.6.3    	0.6.8      	A Helm chart for Kubernetes                       
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
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-7dc46d4d87-pqjwf       1/1     Running   0          4m4s
kubedb      kubedb-kubedb-ops-manager-7fff98c6dd-27jk9      1/1     Running   0          4m4s
kubedb      kubedb-kubedb-provisioner-5865799fcd-2x7wd      1/1     Running   0          4m4s
kubedb      kubedb-kubedb-webhook-server-7fd769f98d-h2gcs   1/1     Running   0          4m4s
kubedb      kubedb-petset-operator-5d94b4ddb8-cgzx8         1/1     Running   0          4m4s
kubedb      kubedb-petset-webhook-server-7558f99b56-j6z2f   2/2     Running   0          4m4s
kubedb      kubedb-sidekick-5d9947bd9-wgqd6                 1/1     Running   0          4m4s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-06-05T05:10:43Z
connectclusters.kafka.kubedb.com                   2024-06-05T05:11:26Z
connectors.kafka.kubedb.com                        2024-06-05T05:11:26Z
druidversions.catalog.kubedb.com                   2024-06-05T05:10:43Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-06-05T05:11:23Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-06-05T05:11:23Z
elasticsearches.kubedb.com                         2024-06-05T05:11:23Z
elasticsearchopsrequests.ops.kubedb.com            2024-06-05T05:11:23Z
elasticsearchversions.catalog.kubedb.com           2024-06-05T05:10:43Z
etcdversions.catalog.kubedb.com                    2024-06-05T05:10:43Z
ferretdbversions.catalog.kubedb.com                2024-06-05T05:10:43Z
kafkaautoscalers.autoscaling.kubedb.com            2024-06-05T05:11:26Z
kafkaconnectorversions.catalog.kubedb.com          2024-06-05T05:10:43Z
kafkaopsrequests.ops.kubedb.com                    2024-06-05T05:11:26Z
kafkas.kubedb.com                                  2024-06-05T05:11:26Z
kafkaversions.catalog.kubedb.com                   2024-06-05T05:10:43Z
mariadbarchivers.archiver.kubedb.com               2024-06-05T05:11:30Z
mariadbautoscalers.autoscaling.kubedb.com          2024-06-05T05:11:30Z
mariadbdatabases.schema.kubedb.com                 2024-06-05T05:11:30Z
mariadbopsrequests.ops.kubedb.com                  2024-06-05T05:11:30Z
mariadbs.kubedb.com                                2024-06-05T05:11:30Z
mariadbversions.catalog.kubedb.com                 2024-06-05T05:10:43Z
memcachedversions.catalog.kubedb.com               2024-06-05T05:10:43Z
mongodbarchivers.archiver.kubedb.com               2024-06-05T05:11:33Z
mongodbautoscalers.autoscaling.kubedb.com          2024-06-05T05:11:33Z
mongodbdatabases.schema.kubedb.com                 2024-06-05T05:11:33Z
mongodbopsrequests.ops.kubedb.com                  2024-06-05T05:11:33Z
mongodbs.kubedb.com                                2024-06-05T05:11:33Z
mongodbversions.catalog.kubedb.com                 2024-06-05T05:10:43Z
mssqlserverversions.catalog.kubedb.com             2024-06-05T05:10:43Z
mysqlarchivers.archiver.kubedb.com                 2024-06-05T05:11:36Z
mysqlautoscalers.autoscaling.kubedb.com            2024-06-05T05:11:36Z
mysqldatabases.schema.kubedb.com                   2024-06-05T05:11:37Z
mysqlopsrequests.ops.kubedb.com                    2024-06-05T05:11:36Z
mysqls.kubedb.com                                  2024-06-05T05:11:36Z
mysqlversions.catalog.kubedb.com                   2024-06-05T05:10:43Z
perconaxtradbversions.catalog.kubedb.com           2024-06-05T05:10:43Z
pgbouncerversions.catalog.kubedb.com               2024-06-05T05:10:43Z
pgpoolversions.catalog.kubedb.com                  2024-06-05T05:10:43Z
postgresarchivers.archiver.kubedb.com              2024-06-05T05:11:40Z
postgresautoscalers.autoscaling.kubedb.com         2024-06-05T05:11:40Z
postgresdatabases.schema.kubedb.com                2024-06-05T05:11:40Z
postgreses.kubedb.com                              2024-06-05T05:11:40Z
postgresopsrequests.ops.kubedb.com                 2024-06-05T05:11:40Z
postgresversions.catalog.kubedb.com                2024-06-05T05:10:43Z
proxysqlversions.catalog.kubedb.com                2024-06-05T05:10:43Z
publishers.postgres.kubedb.com                     2024-06-05T05:11:40Z
rabbitmqversions.catalog.kubedb.com                2024-06-05T05:10:44Z
redisautoscalers.autoscaling.kubedb.com            2024-06-05T05:11:43Z
redises.kubedb.com                                 2024-06-05T05:11:43Z
redisopsrequests.ops.kubedb.com                    2024-06-05T05:11:43Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-06-05T05:11:43Z
redissentinelopsrequests.ops.kubedb.com            2024-06-05T05:11:43Z
redissentinels.kubedb.com                          2024-06-05T05:11:43Z
redisversions.catalog.kubedb.com                   2024-06-05T05:10:44Z
schemaregistries.kafka.kubedb.com                  2024-06-05T05:11:26Z
schemaregistryversions.catalog.kubedb.com          2024-06-05T05:10:44Z
singlestoreversions.catalog.kubedb.com             2024-06-05T05:10:44Z
solrversions.catalog.kubedb.com                    2024-06-05T05:10:44Z
subscribers.postgres.kubedb.com                    2024-06-05T05:11:40Z
zookeeperversions.catalog.kubedb.com               2024-06-05T05:10:44Z
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
    storageClassName: "gp2"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `mysql-cluster.yaml`
Then create the above MySQL CR,

```bash
$ kubectl apply -f mysql-cluster.yaml
mysql.kubedb.com/mysql-cluster created
```

In this yaml,

- In this yaml we can see in the `spec.version` field specifies the version of MySQL. Here, we are using MySQL `8.2.0`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
- `spec.topology` represents the clustering configuration for MySQL.
- `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `GroupReplication`.
- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs.
- `spec.terminationPolicy` field is _Wipeout_ means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy) .

Once these are handled correctly and the MySQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                  READY   STATUS    RESTARTS   AGE
pod/mysql-cluster-0   2/2     Running   0          2m15s
pod/mysql-cluster-1   2/2     Running   0          70s
pod/mysql-cluster-2   2/2     Running   0          59s

NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mysql-cluster           ClusterIP   10.96.144.149   <none>        3306/TCP   2m18s
service/mysql-cluster-pods      ClusterIP   None            <none>        3306/TCP   2m18s
service/mysql-cluster-standby   ClusterIP   10.96.208.47    <none>        3306/TCP   2m18s

NAME                             READY   AGE
statefulset.apps/mysql-cluster   3/3     2m15s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-cluster   kubedb.com/mysql   8.2.0     2m15s

NAME                             VERSION   STATUS   AGE
mysql.kubedb.com/mysql-cluster   8.2.0     Ready    2m18s
```

Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo mysql-cluster
NAME            VERSION   STATUS   AGE
mysql-cluster   8.2.0     Ready    2m36s
```

> We have successfully deployed MySQL cluster in Amazon EKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mysql-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                 TYPE                       DATA   AGE
mysql-cluster-auth   kubernetes.io/basic-auth   2      2m52s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mysql-cluster
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
mysql-cluster           ClusterIP   10.96.144.149   <none>        3306/TCP   3m6s
mysql-cluster-pods      ClusterIP   None            <none>        3306/TCP   3m6s
mysql-cluster-standby   ClusterIP   10.96.208.47    <none>        3306/TCP   3m6s
```

Now, we are going to use `mysql-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-cluster-auth -o jsonpath='{.data.password}' | base64 -d
Y*xW6nlN6Dst978l
```

#### Insert Sample Data

In this section, we are going to login into our MySQL database pod and insert some sample data.

```bash
$ kubectl exec -it mysql-cluster-0 -n demo -c mysql -- bash
bash-4.4$ mysql --user=root --password='Y*xW6nlN6Dst978l'

Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 195
Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE Music;
Query OK, 1 row affected (0.01 sec)

mysql> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected, 1 warning (0.01 sec)

mysql> INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Take Me Home, Country Roads");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM Music.Artist;
+----+-------------+-----------------------------+
| id | Name        | Song                        |
+----+-------------+-----------------------------+
|  1 | John Denver | Take Me Home, Country Roads |
+----+-------------+-----------------------------+
1 row in set (0.01 sec)

mysql> exit
Bye
```

> We've successfully inserted some sample data to our database. More information about Deploy & Manage MySQL on Kubernetes can be found in [Kubernetes MySQL](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

## Vertical Scaling of MySQL Cluster

### Vertical Scale Up

We are going to scale up the resources of the MySQL cluster. Before applying Vertical Scaling, let’s check the current resources.

```bash
$ kubectl get pod -n demo mysql-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
{}
```

### Create MySQLOpsRequest

In order to scale up, we have to create a `MySQLOpsRequest`, Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: vertical-scale-up
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mysql-cluster
  verticalScaling:
    mysql:
      resources:
        requests:
          memory: "1200Mi"
          cpu: "0.7"
        limits:
          memory: "1200Mi"
          cpu: "0.7"
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mysql-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mysql` specifies the expected mysql container resources after scaling.

Let’s save this yaml configuration into `vertical-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
mysqlopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `MySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ watch kubectl get mysqlopsrequest -n demo
NAME                TYPE              STATUS       AGE
vertical-scale-up   VerticalScaling   Successful   4m43s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the current resources,

```bash
$ kubectl get pod -n demo mysql-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "1200Mi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1200Mi"
  }
}
{}
```

> From all the above outputs we can see that the resources of the cluster is now increased. That means we have successfully scaled up the resources of the MySQL cluster.

### Vertical Scale Down

Now, we are going to scale down the resources of the cluster.

#### Create MySQLOpsRequest

In order to scale down, again we need to create a new `MySQLOpsRequest`. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: vertical-scale-down
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mysql-cluster
  verticalScaling:
    mysql:
      resources:
        requests:
          memory: "1Gi"
          cpu: "0.5"
        limits:
          memory: "1Gi"
          cpu: "0.5"
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mysql-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mysql` specifies the expected mysql container resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
mysqlopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `MySQLOpsRequest` `STATUS` to be Successful. Run the following command to watch `MySQLOpsRequest` CR,

```bash
$ watch kubectl get mysqlopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   5m7s
```

From the above output we can see that the `MySQLOpsRequest` has succeeded. Now, we are going to verify the resources,

```bash
$ kubectl get pod -n demo mysql-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
{}
```

> From all the above outputs we can see that the resources of the cluster is decreased. That means we have successfully scaled down the resources of the MySQL cluster.

If you want to learn more about Production-Grade MySQL on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=ThdfNCRulTAsqfnz&amp;list=PLoiT1Gv2KR1gNPaHZtfdBZb6G4wLx6Iks" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MySQL on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
