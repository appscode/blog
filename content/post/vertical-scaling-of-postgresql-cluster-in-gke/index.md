---
title: Vertical Scaling of PostgreSQL Cluster in Google Kubernetes Engine (GKE)
date: "2024-04-22"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- gcp
- gke
- kubedb
- kubernetes
- postgresql
- postgresql-cluster
- postgresql-database
- vertical-scaling
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will show Vertical scaling of PostgreSQL cluster in Google Kubernetes Engine (GKE). We will cover the following steps:

1. Install KubeDB
2. Deploy PostgreSQL Cluster
3. Read/Write Sample Data
4. Vertical Scaling of PostgreSQL Cluster

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
appscode/kubedb                   	v2024.3.16   	v2024.3.16 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.29.0      	v0.29.1    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
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
appscode/kubedb-provisioner       	v0.44.0      	v0.44.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.20.0      	v0.20.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.4.2    	0.6.5      	A Helm chart for Kubernetes                       
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
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-777bc6b47d-66f9f       1/1     Running   0          4m52s
kubedb      kubedb-kubedb-ops-manager-8655d7fd94-cvwgm      1/1     Running   0          4m52s
kubedb      kubedb-kubedb-provisioner-6868bbff85-wlbjf      1/1     Running   0          4m52s
kubedb      kubedb-kubedb-webhook-server-848d6fb7b7-hr59l   1/1     Running   0          4m52s
kubedb      kubedb-petset-operator-5d94b4ddb8-9gwf9         1/1     Running   0          4m52s
kubedb      kubedb-petset-webhook-server-78d78bc87-5hx2f    2/2     Running   0          4m52s
kubedb      kubedb-sidekick-5dc87959b7-khh5r                1/1     Running   0          4m52s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-04-22T03:45:17Z
connectors.kafka.kubedb.com                        2024-04-22T03:45:17Z
druidversions.catalog.kubedb.com                   2024-04-22T03:44:34Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-04-22T03:45:14Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-04-22T03:45:14Z
elasticsearches.kubedb.com                         2024-04-22T03:45:14Z
elasticsearchopsrequests.ops.kubedb.com            2024-04-22T03:45:14Z
elasticsearchversions.catalog.kubedb.com           2024-04-22T03:44:34Z
etcdversions.catalog.kubedb.com                    2024-04-22T03:44:34Z
ferretdbversions.catalog.kubedb.com                2024-04-22T03:44:34Z
kafkaautoscalers.autoscaling.kubedb.com            2024-04-22T03:45:17Z
kafkaconnectorversions.catalog.kubedb.com          2024-04-22T03:44:34Z
kafkaopsrequests.ops.kubedb.com                    2024-04-22T03:45:17Z
kafkas.kubedb.com                                  2024-04-22T03:45:17Z
kafkaversions.catalog.kubedb.com                   2024-04-22T03:44:34Z
mariadbarchivers.archiver.kubedb.com               2024-04-22T03:45:21Z
mariadbautoscalers.autoscaling.kubedb.com          2024-04-22T03:45:21Z
mariadbdatabases.schema.kubedb.com                 2024-04-22T03:45:21Z
mariadbopsrequests.ops.kubedb.com                  2024-04-22T03:45:21Z
mariadbs.kubedb.com                                2024-04-22T03:45:21Z
mariadbversions.catalog.kubedb.com                 2024-04-22T03:44:34Z
memcachedversions.catalog.kubedb.com               2024-04-22T03:44:34Z
mongodbarchivers.archiver.kubedb.com               2024-04-22T03:45:25Z
mongodbautoscalers.autoscaling.kubedb.com          2024-04-22T03:45:24Z
mongodbdatabases.schema.kubedb.com                 2024-04-22T03:45:25Z
mongodbopsrequests.ops.kubedb.com                  2024-04-22T03:45:24Z
mongodbs.kubedb.com                                2024-04-22T03:45:24Z
mongodbversions.catalog.kubedb.com                 2024-04-22T03:44:34Z
mysqlarchivers.archiver.kubedb.com                 2024-04-22T03:45:28Z
mysqlautoscalers.autoscaling.kubedb.com            2024-04-22T03:45:28Z
mysqldatabases.schema.kubedb.com                   2024-04-22T03:45:28Z
mysqlopsrequests.ops.kubedb.com                    2024-04-22T03:45:28Z
mysqls.kubedb.com                                  2024-04-22T03:45:28Z
mysqlversions.catalog.kubedb.com                   2024-04-22T03:44:34Z
perconaxtradbversions.catalog.kubedb.com           2024-04-22T03:44:34Z
pgbouncerversions.catalog.kubedb.com               2024-04-22T03:44:34Z
pgpoolversions.catalog.kubedb.com                  2024-04-22T03:44:34Z
postgresarchivers.archiver.kubedb.com              2024-04-22T03:45:32Z
postgresautoscalers.autoscaling.kubedb.com         2024-04-22T03:45:32Z
postgresdatabases.schema.kubedb.com                2024-04-22T03:45:32Z
postgreses.kubedb.com                              2024-04-22T03:45:32Z
postgresopsrequests.ops.kubedb.com                 2024-04-22T03:45:32Z
postgresversions.catalog.kubedb.com                2024-04-22T03:44:34Z
proxysqlversions.catalog.kubedb.com                2024-04-22T03:44:34Z
publishers.postgres.kubedb.com                     2024-04-22T03:45:32Z
rabbitmqversions.catalog.kubedb.com                2024-04-22T03:44:34Z
redisautoscalers.autoscaling.kubedb.com            2024-04-22T03:45:35Z
redises.kubedb.com                                 2024-04-22T03:45:35Z
redisopsrequests.ops.kubedb.com                    2024-04-22T03:45:35Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-04-22T03:45:35Z
redissentinelopsrequests.ops.kubedb.com            2024-04-22T03:45:35Z
redissentinels.kubedb.com                          2024-04-22T03:45:35Z
redisversions.catalog.kubedb.com                   2024-04-22T03:44:34Z
singlestoreversions.catalog.kubedb.com             2024-04-22T03:44:34Z
solrversions.catalog.kubedb.com                    2024-04-22T03:44:34Z
subscribers.postgres.kubedb.com                    2024-04-22T03:45:32Z
zookeeperversions.catalog.kubedb.com               2024-04-22T03:44:34Z
```

## Deploy PostgreSQL Cluster

We are going to Deploy PostgreSQL Cluster by using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the PostgreSQL CRO we are going to use:

```yaml                                                                      
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-cluster
  namespace: demo
spec:
  version: "16.1"
  replicas: 3
  standbyMode: Hot
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
Let's save this yaml configuration into `postgres-cluster.yaml` 
Then create the above PostgreSQL CRO

```bash
$ kubectl apply -f postgres-cluster.yaml
postgres.kubedb.com/postgres-cluster created
```
In this yaml,
* In this yaml we can see in the `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 16.1`. You can list the KubeDB supported versions of PostgreSQL by running `kubectl get postgresversions` command.
* `spec.standby` is an optional field that specifies the standby mode `hot` or `warm` to use for standby replicas. In `hot` standby mode, standby replicas can accept connection and run read-only queries. In `warm` standby mode, standby replicas can’t accept connection and only used for replication purpose.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy) .

Once these are handled correctly and the PostgreSQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/postgres-cluster-0   2/2     Running   0          2m15s
pod/postgres-cluster-1   2/2     Running   0          90s
pod/postgres-cluster-2   2/2     Running   0          83s

NAME                               TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                      AGE
service/postgres-cluster           ClusterIP   10.96.10.43   <none>        5432/TCP,2379/TCP            2m19s
service/postgres-cluster-pods      ClusterIP   None          <none>        5432/TCP,2380/TCP,2379/TCP   2m19s
service/postgres-cluster-standby   ClusterIP   10.96.84.21   <none>        5432/TCP                     2m19s

NAME                                READY   AGE
statefulset.apps/postgres-cluster   3/3     2m15s

NAME                                                  TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/postgres-cluster   kubedb.com/postgres   16.1      2m15s

NAME                                   VERSION   STATUS   AGE
postgres.kubedb.com/postgres-cluster   16.1      Ready    2m19s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get pg -n demo postgres-cluster
NAME               VERSION   STATUS   AGE
postgres-cluster   16.1      Ready    2m39s
```
> We have successfully deployed PostgreSQL cluster in GKE. Now we can exec into the container to use the database.


#### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `postgres-cluster` that we have deployed. Let’s check them using the following commands,

```bash 
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                    TYPE                       DATA   AGE
postgres-cluster-auth   kubernetes.io/basic-auth   2      2m56s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                       TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                      AGE
postgres-cluster           ClusterIP   10.96.10.43   <none>        5432/TCP,2379/TCP            3m15s
postgres-cluster-pods      ClusterIP   None          <none>        5432/TCP,2380/TCP,2379/TCP   3m15s
postgres-cluster-standby   ClusterIP   10.96.84.21   <none>        5432/TCP                     3m15s
```
Now, we are going to use `postgres-cluster-auth` to get the credentials.
```bash
$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.username}' | base64 -d
postgres

$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.password}' | base64 -d
9E6YsUhCwYldZT8t
```

#### Insert Sample Data

In this section, we are going to login into our PostgreSQL pod and insert some sample data. 

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -c postgres -- bash
postgres-cluster-0:/$ psql -d "user=postgres password=9E6YsUhCwYldZT8t"
psql (16.1)
Type "help" for help.

postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(4 rows)

postgres=# CREATE DATABASE music;
CREATE DATABASE

postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 music         | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(5 rows)

postgres=# CREATE TABLE artist (name VARCHAR(50) NOT NULL, song VARCHAR(50) NOT NULL);
CREATE TABLE

postgres=# INSERT INTO artist (name, song) VALUES('John Denver', 'Country Roads');
INSERT 0 1

postgres=# SELECT * FROM artist;
    name     |     song      
-------------+---------------
 John Denver | Country Roads
(1 row)

postgres=# \q

postgres-cluster-0:/$ exit
exit
```

> We've successfully inserted some sample data to our database. More information about Run & Manage PostgreSQL on Kubernetes can be found in [PostgreSQL Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

## Vertical Scaling of PostgreSQL Cluster

We are going to scale up the current cpu resource of the PostgreSQL cluster by applying Vertical Scaling.
Before applying it, let's check the current resources,

```bash
$ kubectl get pod -n demo postgres-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
{
  "limits": {
    "memory": "256Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "256Mi"
  }
}
```
### Vertical Scale Up

#### Create PostgresOpsRequest

In order to update the resources of the cluster, we have to create a `PostgresOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: vertical-scale-up
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: postgres-cluster
  verticalScaling:
    postgres:
      resources:
        requests:
          memory: "1100Mi"
          cpu: "0.55"
        limits:
          memory: "1100Mi"
          cpu: "0.55"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `postgres-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.resources` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-up.yaml
postgresopsrequest.ops.kubedb.com/vertical-scale-up created
```

Let’s wait for `PostgresOpsRequest` `STATUS` to be Successful. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ watch kubectl get postgresopsrequest -n demo
NAME                TYPE              STATUS       AGE
vertical-scale-up   VerticalScaling   Successful   115s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo postgres-cluster-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "550m",
    "memory": "1100Mi"
  },
  "requests": {
    "cpu": "550m",
    "memory": "1100Mi"
  }
}
{
  "limits": {
    "memory": "256Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "256Mi"
  }
}
```
> The above output verifies that we have successfully scaled up the resources of the PostgreSQL cluster.

### Vertical Scale Down

#### Create PostgresOpsRequest

In order to update the resources of the database, we have to create a `PostgresOpsRequest` CR with our desired resources. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: vertical-scale-down
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: postgres-cluster
  verticalScaling:
    postgres:
      resources:
        requests:
          memory: "1Gi"
          cpu: "0.5"
        limits:
          memory: "1Gi"
          cpu: "0.5"
```
In this yaml,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `postgres-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.resources` specifies the desired resources after scaling.

Let’s save this yaml configuration into `vertical-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f vertical-scale-down.yaml
postgresopsrequest.ops.kubedb.com/vertical-scale-down created
```

Let’s wait for `PostgresOpsRequest` `STATUS` to be Successful. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ watch kubectl get postgresopsrequest -n demo
NAME                  TYPE              STATUS       AGE
vertical-scale-down   VerticalScaling   Successful   2m5s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state. Let’s check with the following command,

```bash
$ kubectl get pod -n demo postgres-cluster-0 -o json | jq '.spec.containers[].resources'
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
{
  "limits": {
    "memory": "256Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "256Mi"
  }
}

```
> The above output verifies that we have successfully scaled down the resources of the PostgreSQL cluster.

If you want to learn more about Production-Grade PostgreSQL you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1imqnrYFhUNTLHdBNFXPKr_" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://youtube.com/@appscode) channel.

Learn more about [PostgreSQL on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).

