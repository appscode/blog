---
title: Deploy and Manage Redis in Sentinel Mode in Azure Kubernetes Service (AKS)
date: "2023-09-06"
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
- microsoft-azure
- redis
- sentinel
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Kafka, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). In this tutorial we will deploy and manage Redis in Sentinel Mode in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Sentinel 
3) Deploy Redis Cluster
4) Horizontal Scaling of Redis Sentinel

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
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
appscode/kubedb-ops-manager       	v0.22.0      	v0.22.4    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.08.18  	v2023.08.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.35.0      	v0.35.3    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.11.0      	v0.11.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.4.1      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.11.0      	v0.11.0    	KubeDB Webhook Server by AppsCode 


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

NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-65c5b6bf76-cxr5w       1/1     Running   0          2m42s
kubedb      kubedb-kubedb-dashboard-59f847486d-xkfm9        1/1     Running   0          2m42s
kubedb      kubedb-kubedb-ops-manager-69f58475bf-kgj8l      1/1     Running   0          2m42s
kubedb      kubedb-kubedb-provisioner-7764ccc56c-sj77v      1/1     Running   0          2m41s
kubedb      kubedb-kubedb-schema-manager-76ccf4f9db-56db7   1/1     Running   0          2m42s
kubedb      kubedb-kubedb-webhook-server-7cb496867-h4522    1/1     Running   0          2m42s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-09-06T13:31:19Z
elasticsearchdashboards.dashboard.kubedb.com      2023-09-06T13:29:40Z
elasticsearches.kubedb.com                        2023-09-06T13:29:40Z
elasticsearchopsrequests.ops.kubedb.com           2023-09-06T13:30:36Z
elasticsearchversions.catalog.kubedb.com          2023-09-06T13:29:19Z
etcds.kubedb.com                                  2023-09-06T13:31:02Z
etcdversions.catalog.kubedb.com                   2023-09-06T13:29:19Z
kafkas.kubedb.com                                 2023-09-06T13:31:04Z
kafkaversions.catalog.kubedb.com                  2023-09-06T13:29:19Z
mariadbautoscalers.autoscaling.kubedb.com         2023-09-06T13:31:19Z
mariadbdatabases.schema.kubedb.com                2023-09-06T13:31:46Z
mariadbopsrequests.ops.kubedb.com                 2023-09-06T13:30:49Z
mariadbs.kubedb.com                               2023-09-06T13:30:49Z
mariadbversions.catalog.kubedb.com                2023-09-06T13:29:19Z
memcacheds.kubedb.com                             2023-09-06T13:31:02Z
memcachedversions.catalog.kubedb.com              2023-09-06T13:29:19Z
mongodbautoscalers.autoscaling.kubedb.com         2023-09-06T13:31:19Z
mongodbdatabases.schema.kubedb.com                2023-09-06T13:31:46Z
mongodbopsrequests.ops.kubedb.com                 2023-09-06T13:30:39Z
mongodbs.kubedb.com                               2023-09-06T13:30:39Z
mongodbversions.catalog.kubedb.com                2023-09-06T13:29:19Z
mysqlautoscalers.autoscaling.kubedb.com           2023-09-06T13:31:19Z
mysqldatabases.schema.kubedb.com                  2023-09-06T13:31:45Z
mysqlopsrequests.ops.kubedb.com                   2023-09-06T13:30:45Z
mysqls.kubedb.com                                 2023-09-06T13:30:45Z
mysqlversions.catalog.kubedb.com                  2023-09-06T13:29:19Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-09-06T13:31:19Z
perconaxtradbopsrequests.ops.kubedb.com           2023-09-06T13:31:01Z
perconaxtradbs.kubedb.com                         2023-09-06T13:31:01Z
perconaxtradbversions.catalog.kubedb.com          2023-09-06T13:29:19Z
pgbouncers.kubedb.com                             2023-09-06T13:30:42Z
pgbouncerversions.catalog.kubedb.com              2023-09-06T13:29:19Z
postgresautoscalers.autoscaling.kubedb.com        2023-09-06T13:31:19Z
postgresdatabases.schema.kubedb.com               2023-09-06T13:31:46Z
postgreses.kubedb.com                             2023-09-06T13:30:55Z
postgresopsrequests.ops.kubedb.com                2023-09-06T13:30:55Z
postgresversions.catalog.kubedb.com               2023-09-06T13:29:19Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-09-06T13:31:19Z
proxysqlopsrequests.ops.kubedb.com                2023-09-06T13:30:58Z
proxysqls.kubedb.com                              2023-09-06T13:30:58Z
proxysqlversions.catalog.kubedb.com               2023-09-06T13:29:19Z
publishers.postgres.kubedb.com                    2023-09-06T13:31:11Z
redisautoscalers.autoscaling.kubedb.com           2023-09-06T13:31:19Z
redises.kubedb.com                                2023-09-06T13:30:52Z
redisopsrequests.ops.kubedb.com                   2023-09-06T13:30:52Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-09-06T13:31:19Z
redissentinelopsrequests.ops.kubedb.com           2023-09-06T13:31:04Z
redissentinels.kubedb.com                         2023-09-06T13:31:03Z
redisversions.catalog.kubedb.com                  2023-09-06T13:29:19Z
subscribers.postgres.kubedb.com                   2023-09-06T13:31:14Z
```

## Deploy Redis Sentinel

Now we are going to deploy Redis sentinel using KubeDB. First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the Redis sentinel we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: sentinel
  namespace: demo
spec:
  version: 7.0.5
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "default"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `sentinel.yaml` 
Then create the above Redis sentinel

```bash
$ kubectl create -f sentinel.yaml
redissentinel.kubedb.com/sentinel created
```
In this yaml,
* Here, we can see in the `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.0.5`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Let's check the status of Redis sentinel,

```bash
$ kubectl get redissentinel -n demo
NAME       VERSION   STATUS   AGE
sentinel   7.0.5     Ready    90s
```

## Deploy Redis Cluster

Now, we are going to deploy Redis cluster using KubeDB. Here is the yaml we are going to use,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis
  namespace: demo
spec:
  version: 7.0.5
  replicas: 3
  sentinelRef: 
    name: sentinel
    namespace: demo
  mode: Sentinel
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "default"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `redis.yaml` 
Then create the above Redis CRD

```bash
$ kubectl apply -f redis.yaml
redis.kubedb.com/redis created
```
In this yaml,
* Here, we can see in the `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.0.5`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* `spec.sentinelRef.name` and `spec.sentinelRef.namespace` specifies the sentinel instance which will monitor this Redis database.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [Termination Policy](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME             READY   STATUS    RESTARTS   AGE
pod/redis-0      2/2     Running   0          99s
pod/redis-1      2/2     Running   0          77s
pod/redis-2      2/2     Running   0          71s
pod/sentinel-0   1/1     Running   0          4m23s
pod/sentinel-1   1/1     Running   0          3m57s
pod/sentinel-2   1/1     Running   0          3m50s

NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/redis           ClusterIP   10.96.44.195    <none>        6379/TCP    103s
service/redis-pods      ClusterIP   None            <none>        6379/TCP    103s
service/redis-standby   ClusterIP   10.96.224.72    <none>        6379/TCP    103s
service/sentinel        ClusterIP   10.96.157.150   <none>        26379/TCP   4m27s
service/sentinel-pods   ClusterIP   None            <none>        26379/TCP   4m27s

NAME                        READY   AGE
statefulset.apps/redis      3/3     99s
statefulset.apps/sentinel   3/3     4m23s

NAME                                          TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/redis      kubedb.com/redis           7.0.5     99s
appbinding.appcatalog.appscode.com/sentinel   kubedb.com/redissentinel   7.0.5     4m23s

NAME                     VERSION   STATUS   AGE
redis.kubedb.com/redis   7.0.5     Ready    103s

NAME                                VERSION   STATUS   AGE
redissentinel.kubedb.com/sentinel   7.0.5     Ready    4m27s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME    VERSION   STATUS   AGE
redis   7.0.5     Ready    2m39s
```
> We have successfully deployed Redis sentinel in AKS.

### Accessing Sentinel Through CLI

In this section, We will exec into the sentinel pod and show you that it is continuously monitoring the Redis database,

```bash
$ kubectl exec -it -n demo sentinel-0 -- bash
Defaulted container "redissentinel" out of: redissentinel, sentinel-init (init)
root@sentinel-0:/data# redis-cli -p 26379
127.0.0.1:26379> sentinel masters
1)  1) "name"
    2) "demo/redis"
    3) "ip"
    4) "redis-0.redis-pods.demo.svc"
    5) "port"
    6) "6379"
    7) "runid"
    8) "43080fc19bf23e81bb894cc4e4c32208c9f7caac"
    9) "flags"
   10) "master"
   11) "link-pending-commands"
   12) "0"
   13) "link-refcount"
   14) "1"
   15) "last-ping-sent"
   16) "0"
   17) "last-ok-ping-reply"
   18) "358"
   19) "last-ping-reply"
   20) "358"
   21) "down-after-milliseconds"
   22) "5000"
   23) "info-refresh"
   24) "7570"
   25) "role-reported"
   26) "master"
   27) "role-reported-time"
   28) "100382"
   29) "config-epoch"
   30) "0"
   31) "num-slaves"
   32) "2"
   33) "num-other-sentinels"
   34) "2"
   35) "quorum"
   36) "2"
   37) "failover-timeout"
   38) "5000"
   39) "parallel-syncs"
   40) "1"
127.0.0.1:26379> exit
```

## Horizontal Scaling of Redis Sentinel

### Scale Up Replicas

Here, we are going to scale up the replicas of the Redis sentinel to meet the desired number of replicas after scaling.

Before applying Horizontal Scaling, let's check the current number of replicas,

```bash
$ kubectl get redissentinel -n demo sentinel -o json | jq '.spec.replicas'
3
```

### Create RedisSentinelOpsRequest

In order to scale up the replicas, we have to create a `RedisSentinelOpsRequest` CR with our desired replicas. Let’s create it using this following yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sentinel
  horizontalScaling:
    replicas: 5
```
Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `sentinel`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-up.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-up.yaml
redissentinelopsrequest.ops.kubedb.com/horizontal-scale-up created
```

Let’s wait for `RedisSentinelOpsRequest` `STATUS` to be Successful. Run the following command to watch `RedisSentinelOpsRequest` CR,

```bash
$ watch kubectl get redissentinelopsrequest -n demo
NAME                  TYPE                STATUS       AGE
horizontal-scale-up   HorizontalScaling   Successful   78s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get redissentinel -n demo sentinel -o json | jq '.spec.replicas'
5
```
> From all the above outputs we can see that the replicas of the sentinel is now increased to 5. That means we have successfully scaled up the Redis Sentinel.

### Scale Down Replicas

Here, we are going to scale down the replicas of the Redis sentinel to meet the desired number of replicas after scaling.

#### Create RedisSentinelOpsRequest

In order to scale down the replicas, we need to create a `RedisSentinelOpsRequest` CR with our desired replicas. Let’s create it using this yaml,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sentinel
  horizontalScaling:
    replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `sentinel`.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let’s save this yaml configuration into `horizontal-scale-down.yaml` and apply it,

```bash
$ kubectl apply -f horizontal-scale-down.yaml
redissentinelopsrequest.ops.kubedb.com/horizontal-scale-down created
```

Let’s wait for `RedisSentinelOpsRequest` `STATUS` to be Successful. Run the following command to watch `RedisSentinelOpsRequest` CR,

```bash
$ watch kubectl get RedisSentinelOpsRequest -n demo
NAME                    TYPE                STATUS       AGE
horizontal-scale-down   HorizontalScaling   Successful   2m11s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded. Now, we are going to verify the number of replicas,

```bash
$ kubectl get redissentinel -n demo sentinel -o json | jq '.spec.replicas'
3
```

> From all the above outputs we can see that the replicas of the Redis sentinel is decreased to 3. That means we have successfully scaled down the Redis sentinel.

We have made an in depth video on Redis Sentinel Ops Requests - Day 2 Lifecycle Management for Redis Sentinel Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/LToGVt1-D50" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).