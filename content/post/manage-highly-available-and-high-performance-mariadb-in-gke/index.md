---
title: Manage Highly Available and High-Performance MariaDB in Google Kubernetes Engine (GKE) Using KubeDB
date: "2023-03-10"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- database
- eks
- high-availability
- high-performance
- horizontal-scaling
- kubedb
- kubernetes
- mariadb
- s3
- vertical-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will Manage Highly Available and High-Performance MariaDB in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy MariaDB Clustered Database
3) Horizontal Scaling of MariaDB Database
4) Vertical Scaling of MariaDB Database

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
debacab3-y89q-4168-ba24-e97a553dcfa4
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
appscode/kubedb                   	v2023.02.28  	v2023.02.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.17.0      	v0.17.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.02.28  	v2023.02.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.02.28  	v2023.02.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.8.0       	v0.8.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.02.28  	v2023.02.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.02.28  	v2023.02.28	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.19.0      	v0.19.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.02.28  	v2023.02.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.32.0      	v0.32.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.8.0       	v0.8.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.26     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.8.0       	v0.8.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.02.28 \
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
kubedb      kubedb-kubedb-autoscaler-55f8544c4b-5fqlq       1/1     Running   0          2m15s
kubedb      kubedb-kubedb-dashboard-79b68c8985-nsbwb        1/1     Running   0          2m15s
kubedb      kubedb-kubedb-ops-manager-6f8fdb47cb-45cwg      1/1     Running   0          2m15s
kubedb      kubedb-kubedb-provisioner-79998586f9-q5ht2      1/1     Running   0          2m15s
kubedb      kubedb-kubedb-schema-manager-585b456768-qmsw6   1/1     Running   0          2m15s
kubedb      kubedb-kubedb-webhook-server-69ddfcd656-pmbzl   1/1     Running   0          2m15s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-03-10T04:42:02Z
elasticsearchdashboards.dashboard.kubedb.com      2023-03-10T04:42:14Z
elasticsearches.kubedb.com                        2023-03-10T04:41:49Z
elasticsearchopsrequests.ops.kubedb.com           2023-03-10T04:42:45Z
elasticsearchversions.catalog.kubedb.com          2023-03-10T04:41:19Z
etcds.kubedb.com                                  2023-03-10T04:41:49Z
etcdversions.catalog.kubedb.com                   2023-03-10T04:41:19Z
kafkas.kubedb.com                                 2023-03-10T04:41:50Z
kafkaversions.catalog.kubedb.com                  2023-03-10T04:41:19Z
mariadbautoscalers.autoscaling.kubedb.com         2023-03-10T04:42:02Z
mariadbdatabases.schema.kubedb.com                2023-03-10T04:42:28Z
mariadbopsrequests.ops.kubedb.com                 2023-03-10T04:42:59Z
mariadbs.kubedb.com                               2023-03-10T04:41:49Z
mariadbversions.catalog.kubedb.com                2023-03-10T04:41:19Z
memcacheds.kubedb.com                             2023-03-10T04:41:49Z
memcachedversions.catalog.kubedb.com              2023-03-10T04:41:19Z
mongodbautoscalers.autoscaling.kubedb.com         2023-03-10T04:42:02Z
mongodbdatabases.schema.kubedb.com                2023-03-10T04:42:27Z
mongodbopsrequests.ops.kubedb.com                 2023-03-10T04:42:48Z
mongodbs.kubedb.com                               2023-03-10T04:41:50Z
mongodbversions.catalog.kubedb.com                2023-03-10T04:41:19Z
mysqlautoscalers.autoscaling.kubedb.com           2023-03-10T04:42:02Z
mysqldatabases.schema.kubedb.com                  2023-03-10T04:42:27Z
mysqlopsrequests.ops.kubedb.com                   2023-03-10T04:42:56Z
mysqls.kubedb.com                                 2023-03-10T04:41:50Z
mysqlversions.catalog.kubedb.com                  2023-03-10T04:41:19Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-03-10T04:42:02Z
perconaxtradbopsrequests.ops.kubedb.com           2023-03-10T04:43:11Z
perconaxtradbs.kubedb.com                         2023-03-10T04:41:50Z
perconaxtradbversions.catalog.kubedb.com          2023-03-10T04:41:19Z
pgbouncers.kubedb.com                             2023-03-10T04:41:50Z
pgbouncerversions.catalog.kubedb.com              2023-03-10T04:41:19Z
postgresautoscalers.autoscaling.kubedb.com        2023-03-10T04:42:02Z
postgresdatabases.schema.kubedb.com               2023-03-10T04:42:28Z
postgreses.kubedb.com                             2023-03-10T04:41:50Z
postgresopsrequests.ops.kubedb.com                2023-03-10T04:43:05Z
postgresversions.catalog.kubedb.com               2023-03-10T04:41:19Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-03-10T04:42:02Z
proxysqlopsrequests.ops.kubedb.com                2023-03-10T04:43:08Z
proxysqls.kubedb.com                              2023-03-10T04:41:50Z
proxysqlversions.catalog.kubedb.com               2023-03-10T04:41:19Z
publishers.postgres.kubedb.com                    2023-03-10T04:43:21Z
redisautoscalers.autoscaling.kubedb.com           2023-03-10T04:42:02Z
redises.kubedb.com                                2023-03-10T04:41:50Z
redisopsrequests.ops.kubedb.com                   2023-03-10T04:43:02Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-03-10T04:42:02Z
redissentinelopsrequests.ops.kubedb.com           2023-03-10T04:43:15Z
redissentinels.kubedb.com                         2023-03-10T04:41:50Z
redisversions.catalog.kubedb.com                  2023-03-10T04:41:19Z
subscribers.postgres.kubedb.com                   2023-03-10T04:43:24Z
```

## Deploy MariaDB Clustered Database

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
  name: mariadb-cluster
  namespace: demo
spec:
  version: "10.10.2"
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

Let's save this yaml configuration into `mariadb-cluster.yaml` 
Then create the above MariaDB CRO

```bash
$ kubectl apply -f mariadb-cluster.yaml
mariadb.kubedb.com/mariadb-cluster created
```

* In this yaml we can see in the `spec.version` field specifies the version of MariaDB Here, we are using MariaDB `version 10.10.2`. You can list the KubeDB supported versions of MariaDB by running `$ kubectl get mariadbversion` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mariadb/concepts/mariadb/#specterminationpolicy).

Once these are handled correctly and the MariaDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                    READY   STATUS    RESTARTS   AGE
pod/mariadb-cluster-0   2/2     Running   0          2m51s
pod/mariadb-cluster-1   2/2     Running   0          2m51s
pod/mariadb-cluster-2   2/2     Running   0          2m51s

NAME                           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/mariadb-cluster        ClusterIP   10.96.47.157   <none>        3306/TCP   2m57s
service/mariadb-cluster-pods   ClusterIP   None           <none>        3306/TCP   2m57s

NAME                               READY   AGE
statefulset.apps/mariadb-cluster   3/3     2m51s

NAME                                                 TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mariadb-cluster   kubedb.com/mariadb   10.10.2   2m51s

NAME                                 VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-cluster   10.10.2   Ready    2m57s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo mariadb-cluster
NAME              VERSION   STATUS   AGE
mariadb-cluster   10.10.2   Ready    3m33s
```
> We have successfully deployed MariaDB in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create `Secret` and `Service` for the database `mariadb-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mariadb-cluster
NAME                   TYPE                       DATA   AGE
mariadb-cluster-auth   kubernetes.io/basic-auth   2      4m21s


$ kubectl get service -n demo -l=app.kubernetes.io/instance=mariadb-cluster
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mariadb-cluster        ClusterIP   10.96.47.157   <none>        3306/TCP   5m39s
mariadb-cluster-pods   ClusterIP   None           <none>        3306/TCP   5m39s
```
Now, we are going to use `mariadb-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mariadb-cluster-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo mariadb-cluster-auth -o jsonpath='{.data.password}' | base64 -d
p9Gy*uja*PiThJuY

```

#### Insert Sample Data

In this section, we are going to login into our MariaDB database pod and insert some sample data. 

```bash
$ kubectl exec -it mariadb-cluster-0 -n demo -c mariadb -- bash
root@mariadb-cluster-0:/# mysql --user=root --password='p9Gy*uja*PiThJuY'
Welcome to the MariaDB monitor.  Commands end with ; or \g.

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE Music;
Query OK, 1 row affected (0.004 sec)

MariaDB [(none)]> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(50));
Query OK, 0 rows affected (0.011 sec)

MariaDB [(none)]> INSERT INTO Music.Artist (Name, Song) VALUES ("John Denver", "Take Me Home Country Roads");
Query OK, 1 row affected (0.003 sec)

MariaDB [(none)]> SELECT * FROM Music.Artist;
+----+-------------+----------------------------+
| id | Name        | Song                       |
+----+-------------+----------------------------+
|  1 | John Denver | Take Me Home Country Roads |
+----+-------------+----------------------------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye

```

> We've successfully inserted some sample data to our database. More information about Run & Manage Production-Grade MariaDB Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

## Horizontal Scaling of MariaDB Cluster

### Scale Up Replicas

Here, we are going to scale up the replicas of the MariaDB cluster to meet the desired number of replicas after scaling.

Before applying Horizontal Scaling, let's check the current number of replicas,

```bash
$ kubectl get mariadb -n demo mariadb-cluster -o json | jq '.spec.replicas'
3
```

Let’s connect to a MariaDB instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it mariadb-cluster-0 -n demo -c mariadb -- bash
root@mariadb-cluster-0:/# mysql --user=root --password='p9Gy*uja*PiThJuY'
Welcome to the MariaDB monitor.  Commands end with ; or \g.

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW STATUS LIKE 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```


### Create MariaDBOpsRequest

In order to scale up the replicas of the cluster of the database, we have to create a `MariaDBOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mariadb-cluster
  horizontalScaling:
    member : 5
```
Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mariadb-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
mariadbopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `MariaDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   4m25s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mariadb -n demo mariadb-cluster -o json | jq '.spec.replicas'
5
```

Let’s connect to a MariaDB instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it mariadb-cluster-0 -n demo -c mariadb -- bash
root@mariadb-cluster-0:/# mysql --user=root --password='p9Gy*uja*PiThJuY'
Welcome to the MariaDB monitor.  Commands end with ; or \g.

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW STATUS LIKE 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 5     |
+--------------------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```

> From all the above outputs we can see that the replicas of the cluster is now increased to 5. That means we have successfully scaled up the replicas of the MariaDB cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the cluster to meet the desired number of replicas after scaling.

#### Create MariaDBOpsRequest

In order to scale down the cluster of the database, we need to create a `MariaDBOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mariadb-cluster
  horizontalScaling:
    member : 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mariadb-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.member` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
mariadbopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `MariaDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   4m22s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get mariadb -n demo mariadb-cluster -o json | jq '.spec.replicas'
3
```

Let’s connect to a MariaDB instance and run this command to check the number of replicas,

```bash
$ kubectl exec -it mariadb-cluster-0 -n demo -c mariadb -- bash
root@mariadb-cluster-0:/# mysql --user=root --password='p9Gy*uja*PiThJuY'
Welcome to the MariaDB monitor.  Commands end with ; or \g.

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW STATUS LIKE 'wsrep_cluster_size';
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
```
> From all the above outputs we can see that the replicas of the cluster is decreased to 3. That means we have successfully scaled down the replicas of the MariaDB cluster.


## Vetical Scaling of MariaDB Cluster

Here, we are going to scale up the current cpu resource of the MariaDB cluster by applying Vertical Scaling.
Before applying it, let's check the current resources,

```bash
$ kubectl get pod -n demo mariadb-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

#### Create MariaDBOpsRequest

In order to update the resources of the database, we have to create a `MariaDBOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: vertical-scale-up
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mariadb-cluster
  verticalScaling:
    mariadb:
      requests:
        memory: "1.5Gi"
        cpu: "0.7"
      limits:
        memory: "1.5Gi"
        cpu: "0.7"
```
Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mariadb-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mariadb` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
mariadbopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `MariaDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
NAME                    TYPE                STATUS       AGE
vertical-scale-up       VerticalScaling     Successful   4m9s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mariadb-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1536Mi"
  }
}
```
> The above output verifies that we have successfully scaled up the resources of the MariaDB database.


### Vertical Scale Down

#### Create MariaDBOpsRequest

In order to scale down the resources of the database, we have to create a `MariaDBOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: vertical-scale-down
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mariadb-cluster
  verticalScaling:
    mariadb:
      requests:
        memory: "1Gi"
        cpu: "0.5"
      limits:
        memory: "1Gi"
        cpu: "0.5"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mariadb-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.mariadb` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
mariadbopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `MariaDBOpsRequest` `STATUS` to be Successful. Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   3m
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo mariadb-cluster-0 -o json | jq '.spec.containers[].resources' 
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```
> The above output verifies that we have successfully scaled down the resources of the MariaDB cluster.

We have made an in depth tutorial on MariaDB Alerting and Multi-Tenancy Support by using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/P8l2v6-yCHU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MariaDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
