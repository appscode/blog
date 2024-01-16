---
title: Horizontal Scaling of Redis Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-01-05"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- horizontal-scaling
- kubedb
- kubernetes
- redis
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). In this tutorial we will show the horizontal scaling of Redis cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Cluster 
3) Read/Write Sample Data
4) Horizontal Scaling of Redis Cluster

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

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

NAMESPACE   NAME                                            READY   STATUS    RESTARTS        AGE
kubedb      kubedb-kubedb-autoscaler-6b4c89dbf9-9gwm4       1/1     Running   0               4m
kubedb      kubedb-kubedb-dashboard-5b6f5598d9-jf5tl        1/1     Running   0               4m
kubedb      kubedb-kubedb-ops-manager-747cf9796d-9dhvz      1/1     Running   0               4m
kubedb      kubedb-kubedb-provisioner-c9db4cc89-xfl8b       1/1     Running   0               4m
kubedb      kubedb-kubedb-webhook-server-748d599596-jf64t   1/1     Running   0               4m
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2024-01-05T05:21:05Z
elasticsearchdashboards.dashboard.kubedb.com      2024-01-05T05:21:03Z
elasticsearches.kubedb.com                        2024-01-05T05:21:03Z
elasticsearchopsrequests.ops.kubedb.com           2024-01-05T05:21:17Z
elasticsearchversions.catalog.kubedb.com          2024-01-05T05:19:17Z
etcds.kubedb.com                                  2024-01-05T05:21:06Z
etcdversions.catalog.kubedb.com                   2024-01-05T05:19:18Z
kafkaopsrequests.ops.kubedb.com                   2024-01-05T05:22:17Z
kafkas.kubedb.com                                 2024-01-05T05:21:22Z
kafkaversions.catalog.kubedb.com                  2024-01-05T05:19:18Z
mariadbautoscalers.autoscaling.kubedb.com         2024-01-05T05:21:05Z
mariadbopsrequests.ops.kubedb.com                 2024-01-05T05:21:52Z
mariadbs.kubedb.com                               2024-01-05T05:21:11Z
mariadbversions.catalog.kubedb.com                2024-01-05T05:19:18Z
memcacheds.kubedb.com                             2024-01-05T05:21:12Z
memcachedversions.catalog.kubedb.com              2024-01-05T05:19:19Z
mongodbarchivers.archiver.kubedb.com              2024-01-05T05:21:24Z
mongodbautoscalers.autoscaling.kubedb.com         2024-01-05T05:21:11Z
mongodbopsrequests.ops.kubedb.com                 2024-01-05T05:21:21Z
mongodbs.kubedb.com                               2024-01-05T05:21:13Z
mongodbversions.catalog.kubedb.com                2024-01-05T05:19:19Z
mysqlarchivers.archiver.kubedb.com                2024-01-05T05:21:28Z
mysqlautoscalers.autoscaling.kubedb.com           2024-01-05T05:21:11Z
mysqlopsrequests.ops.kubedb.com                   2024-01-05T05:21:47Z
mysqls.kubedb.com                                 2024-01-05T05:21:16Z
mysqlversions.catalog.kubedb.com                  2024-01-05T05:19:19Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2024-01-05T05:21:11Z
perconaxtradbopsrequests.ops.kubedb.com           2024-01-05T05:22:08Z
perconaxtradbs.kubedb.com                         2024-01-05T05:21:17Z
perconaxtradbversions.catalog.kubedb.com          2024-01-05T05:19:20Z
pgbouncers.kubedb.com                             2024-01-05T05:21:18Z
pgbouncerversions.catalog.kubedb.com              2024-01-05T05:19:20Z
postgresarchivers.archiver.kubedb.com             2024-01-05T05:21:31Z
postgresautoscalers.autoscaling.kubedb.com        2024-01-05T05:21:12Z
postgreses.kubedb.com                             2024-01-05T05:21:19Z
postgresopsrequests.ops.kubedb.com                2024-01-05T05:22:01Z
postgresversions.catalog.kubedb.com               2024-01-05T05:19:20Z
proxysqlautoscalers.autoscaling.kubedb.com        2024-01-05T05:21:12Z
proxysqlopsrequests.ops.kubedb.com                2024-01-05T05:22:04Z
proxysqls.kubedb.com                              2024-01-05T05:21:19Z
proxysqlversions.catalog.kubedb.com               2024-01-05T05:19:21Z
publishers.postgres.kubedb.com                    2024-01-05T05:22:20Z
redisautoscalers.autoscaling.kubedb.com           2024-01-05T05:21:13Z
redises.kubedb.com                                2024-01-05T05:21:20Z
redisopsrequests.ops.kubedb.com                   2024-01-05T05:21:55Z
redissentinelautoscalers.autoscaling.kubedb.com   2024-01-05T05:21:13Z
redissentinelopsrequests.ops.kubedb.com           2024-01-05T05:22:12Z
redissentinels.kubedb.com                         2024-01-05T05:21:21Z
redisversions.catalog.kubedb.com                  2024-01-05T05:19:21Z
subscribers.postgres.kubedb.com                   2024-01-05T05:22:24Z
```

## Deploy Redis Cluster

Now we are going to deploy Redis cluster using KubeDB. First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the Redis we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 7.2.3
  mode: Cluster
  cluster:
    master: 3
    replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "gp2"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `redis-cluster.yaml` 
Then create the above Redis cluster,

```bash
$ kubectl apply -f redis-cluster.yaml
redis.kubedb.com/redis-cluster created
```
In this yaml,
* we can see in the `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.2.3`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).


Once these are handled correctly you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   1/1     Running   0          8m57s
pod/redis-cluster-shard0-1   1/1     Running   0          8m35s
pod/redis-cluster-shard1-0   1/1     Running   0          8m57s
pod/redis-cluster-shard1-1   1/1     Running   0          8m32s
pod/redis-cluster-shard2-0   1/1     Running   0          8m56s
pod/redis-cluster-shard2-1   1/1     Running   0          8m37s

NAME                         TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/redis-cluster        ClusterIP   10.76.12.61   <none>        6379/TCP   8m58s
service/redis-cluster-pods   ClusterIP   None          <none>        6379/TCP   8m58s

NAME                                    READY   AGE
statefulset.apps/redis-cluster-shard0   2/2     8m59s
statefulset.apps/redis-cluster-shard1   2/2     8m58s
statefulset.apps/redis-cluster-shard2   2/2     8m57s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/redis-cluster   kubedb.com/redis   7.2.3     8m59s

NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   7.2.3     Ready    9m4s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME            VERSION   STATUS   AGE
redis-cluster   7.2.3     Ready    3m19s
```
> We have successfully deployed Redis cluster in AWS.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `redis-cluster` that we have deployed. Let’s check them by following command,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=redis-cluster
NAME                   TYPE                       DATA   AGE
redis-cluster-auth     kubernetes.io/basic-auth   2      9m46s
redis-cluster-config   Opaque                     1      9m46s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=redis-cluster
NAME                 TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
redis-cluster        ClusterIP   10.76.12.61   <none>        6379/TCP   10m
redis-cluster-pods   ClusterIP   None          <none>        6379/TCP   10m
```
Now, we are going to use `PASSWORD` to authenticate and insert some sample data.
At first, let’s export the `PASSWORD` as environment variables to make further commands re-usable.

```bash
$ export PASSWORD=$(kubectl get secrets -n demo redis-cluster-auth -o jsonpath='{.data.\password}' | base64 -d)
```

#### Insert Sample Data

In this section, we are going to login into our Redis database pod and insert some sample data. 

```bash
$ kubectl exec -it -n demo redis-cluster-shard0-0 -- redis-cli -c -a $PASSWORD

127.0.0.1:6379> set Product1 KubeDB
OK
10.72.1.7:6379> set Product2 KubeStash
OK
10.72.5.7:6379> get Product1
"KubeDB"
10.72.1.7:6379> get Product2
"KubeStash"
10.72.5.7:6379> exit
```
> We've successfully inserted some sample data to our database. More information about Deploy & Manage Production-Grade Redis Database on Kubernetes can be found in [Redis Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

## Horizontal Scaling of Redis Cluster

### Scale Up Replicas

Here, we are going to scale up the replicas of the Redis cluster to meet the desired number of master and replicas after scaling.

Before applying Horizontal Scaling, let's check the current number of master and replicas,

```bash
$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.cluster.master'
3

$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.replicas'
1
```

### Create RedisOpsRequest

In order to scale up the replicas, we have to create a `RedisOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: redis-cluster
  horizontalScaling:
    master: 4
    replicas: 2
```
Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `redis-cluster`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.master` specifies the desired master after scaling.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
redisopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `RedisOpsRequest` `STATUS` to be Successful. Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   3m57s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.cluster.master'
4

$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.replicas'
2
```
> From all the above outputs we can see that the master and the replicas is now increased to 4 and 2 successively. That means we have successfully scaled up the Redis cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the Redis cluster to meet the desired number of master and replicas after scaling.

#### Create RedisOpsRequest

In order to scale down the replicas, we need to create a `RedisOpsRequest` CR with our desired replicas. Let’s create it using this yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: redis-cluster
  horizontalScaling:
    master: 3
    replicas: 1
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `redis-cluster`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.master` specifies the desired master after scaling.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
redisopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `RedisOpsRequest` `STATUS` to be Successful. Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   5m21s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.cluster.master'
3

$ kubectl get redis -n demo redis-cluster -o json | jq '.spec.replicas'
1
```

> From all the above outputs we can see that the master and replicas of the Redis cluster is decreased to 3 and 1 successively. That means we have successfully scaled down the Redis cluster.



If you want to learn more about Production-Grade Redis on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=n07b2O7yiNYFtob0&amp;list=PLoiT1Gv2KR1iSuQq_iyypzqvHW9u_un04" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Redis on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).