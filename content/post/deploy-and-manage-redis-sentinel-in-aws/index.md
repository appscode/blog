---
title: Deploy and Manage Redis in Sentinel Mode in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-08-16"
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
- redis
- sentinel
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Kafka, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). In this tutorial we will deploy and manage Redis in Sentinel Mode in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Sentinel 
3) Deploy Redis Cluster
4) Horizontal Scaling of Redis Sentinel

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
fc435a61-c74b-9243-83a5-f1110ef2462c
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
appscode/kubedb                   	v2023.06.19  	v2023.06.19	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.19.0      	v0.19.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.06.19  	v2023.06.19	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.06.19  	v2023.06.19	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.10.0      	v0.10.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.06.19  	v2023.06.19	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.06.19  	v2023.06.19	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.06.19  	v2023.06.19	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.21.0      	v0.21.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.06.19  	v2023.06.19	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.34.0      	v0.34.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.10.0      	v0.10.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.33-rc.2	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.10.0      	v0.10.0    	KubeDB Webhook Server by AppsCode  


# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.06.19 \
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
kubedb      kubedb-kubedb-autoscaler-5d6cfbbdc6-mn7xn       1/1     Running   0          2m28s
kubedb      kubedb-kubedb-dashboard-7f6d5c646b-kzmj4        1/1     Running   0          2m26s
kubedb      kubedb-kubedb-ops-manager-57db88cc8-gmpb5       1/1     Running   0          2m28s
kubedb      kubedb-kubedb-provisioner-f88f9d4f6-gq8nh       1/1     Running   0          2m28s
kubedb      kubedb-kubedb-schema-manager-74447b985c-5qqhp   1/1     Running   0          2m27s
kubedb      kubedb-kubedb-webhook-server-75dd594dcf-54lnw   1/1     Running   0          2m28s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-08-16T10:58:44Z
elasticsearchdashboards.dashboard.kubedb.com      2023-08-16T10:58:45Z
elasticsearches.kubedb.com                        2023-08-16T10:58:47Z
elasticsearchopsrequests.ops.kubedb.com           2023-08-16T10:59:00Z
elasticsearchversions.catalog.kubedb.com          2023-08-16T10:56:21Z
etcds.kubedb.com                                  2023-08-16T10:58:59Z
etcdversions.catalog.kubedb.com                   2023-08-16T10:56:21Z
kafkas.kubedb.com                                 2023-08-16T10:59:21Z
kafkaversions.catalog.kubedb.com                  2023-08-16T10:56:21Z
mariadbautoscalers.autoscaling.kubedb.com         2023-08-16T10:58:44Z
mariadbdatabases.schema.kubedb.com                2023-08-16T10:58:57Z
mariadbopsrequests.ops.kubedb.com                 2023-08-16T10:59:29Z
mariadbs.kubedb.com                               2023-08-16T10:58:59Z
mariadbversions.catalog.kubedb.com                2023-08-16T10:56:22Z
memcacheds.kubedb.com                             2023-08-16T10:59:03Z
memcachedversions.catalog.kubedb.com              2023-08-16T10:56:22Z
mongodbautoscalers.autoscaling.kubedb.com         2023-08-16T10:58:44Z
mongodbdatabases.schema.kubedb.com                2023-08-16T10:58:48Z
mongodbopsrequests.ops.kubedb.com                 2023-08-16T10:59:05Z
mongodbs.kubedb.com                               2023-08-16T10:58:51Z
mongodbversions.catalog.kubedb.com                2023-08-16T10:56:23Z
mysqlautoscalers.autoscaling.kubedb.com           2023-08-16T10:58:44Z
mysqldatabases.schema.kubedb.com                  2023-08-16T10:58:44Z
mysqlopsrequests.ops.kubedb.com                   2023-08-16T10:59:26Z
mysqls.kubedb.com                                 2023-08-16T10:58:45Z
mysqlversions.catalog.kubedb.com                  2023-08-16T10:56:23Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-08-16T10:58:44Z
perconaxtradbopsrequests.ops.kubedb.com           2023-08-16T10:59:44Z
perconaxtradbs.kubedb.com                         2023-08-16T10:59:16Z
perconaxtradbversions.catalog.kubedb.com          2023-08-16T10:56:23Z
pgbouncers.kubedb.com                             2023-08-16T10:59:16Z
pgbouncerversions.catalog.kubedb.com              2023-08-16T10:56:23Z
postgresautoscalers.autoscaling.kubedb.com        2023-08-16T10:58:44Z
postgresdatabases.schema.kubedb.com               2023-08-16T10:58:54Z
postgreses.kubedb.com                             2023-08-16T10:58:56Z
postgresopsrequests.ops.kubedb.com                2023-08-16T10:59:37Z
postgresversions.catalog.kubedb.com               2023-08-16T10:56:24Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-08-16T10:58:45Z
proxysqlopsrequests.ops.kubedb.com                2023-08-16T10:59:40Z
proxysqls.kubedb.com                              2023-08-16T10:59:19Z
proxysqlversions.catalog.kubedb.com               2023-08-16T10:56:24Z
publishers.postgres.kubedb.com                    2023-08-16T10:59:55Z
redisautoscalers.autoscaling.kubedb.com           2023-08-16T10:58:45Z
redises.kubedb.com                                2023-08-16T10:59:19Z
redisopsrequests.ops.kubedb.com                   2023-08-16T10:59:33Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-08-16T10:58:46Z
redissentinelopsrequests.ops.kubedb.com           2023-08-16T10:59:48Z
redissentinels.kubedb.com                         2023-08-16T10:59:20Z
redisversions.catalog.kubedb.com                  2023-08-16T10:56:24Z
subscribers.postgres.kubedb.com                   2023-08-16T10:59:58Z
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
    storageClassName: "gp2"
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
NAME                                VERSION   STATUS   AGE
redissentinel.kubedb.com/sentinel   7.0.5     Ready    3m9s
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
    storageClassName: "gp2"
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
pod/redis-0      2/2     Running   0          3m20s
pod/redis-1      2/2     Running   0          2m54s
pod/redis-2      2/2     Running   0          2m11s
pod/sentinel-0   1/1     Running   0          11m
pod/sentinel-1   1/1     Running   0          11m
pod/sentinel-2   1/1     Running   0          11m

NAME                    TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
service/redis           ClusterIP   10.8.2.133   <none>        6379/TCP    3m14s
service/redis-pods      ClusterIP   None         <none>        6379/TCP    3m29s
service/redis-standby   ClusterIP   10.8.9.240   <none>        6379/TCP    3m18s
service/sentinel        ClusterIP   10.8.6.95    <none>        26379/TCP   11m
service/sentinel-pods   ClusterIP   None         <none>        26379/TCP   11m

NAME                        READY   AGE
statefulset.apps/redis      3/3     4m26s
statefulset.apps/sentinel   3/3     19m

NAME                                          TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/redis      kubedb.com/redis           7.0.5     3m37s
appbinding.appcatalog.appscode.com/sentinel   kubedb.com/redissentinel   7.0.5     11m

NAME                                VERSION   STATUS   AGE
redissentinel.kubedb.com/sentinel   7.0.5     Ready    12m

NAME                     VERSION   STATUS   AGE
redis.kubedb.com/redis   7.0.5     Ready    3m42s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME    VERSION   STATUS   AGE
redis   7.0.5     Ready    4m41s
```
> We have successfully deployed Redis sentinel in AWS.

### Accessing Sentinel Through CLI

In this section, We will exec into the sentinel pod and show you that it is continuously monitoring the Redis database,

```bash
kc exec -it -n demo sentinel-0 -- bash
Defaulted container "redissentinel" out of: redissentinel, sentinel-init (init)
root@sentinel-0:/data# redis-cli -p 26379
127.0.0.1:26379> 
127.0.0.1:26379> sentinel masters
1)  1) "name"
    2) "demo/redis"
    3) "ip"
    4) "redis-0.redis-pods.demo.svc"
    5) "port"
    6) "6379"
    7) "runid"
    8) "d927e06b07b8bf7140cff0ceb82b77d092b82a45"
    9) "flags"
   10) "master"
   11) "link-pending-commands"
   12) "0"
   13) "link-refcount"
   14) "1"
   15) "last-ping-sent"
   16) "0"
   17) "last-ok-ping-reply"
   18) "513"
   19) "last-ping-reply"
   20) "513"
   21) "down-after-milliseconds"
   22) "5000"
   23) "info-refresh"
   24) "4395"
   25) "role-reported"
   26) "master"
   27) "role-reported-time"
   28) "936584"
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
horizontal-scale-up   HorizontalScaling   Successful   59s
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
horizontal-scale-down   HorizontalScaling   Successful   2m33s
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

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).