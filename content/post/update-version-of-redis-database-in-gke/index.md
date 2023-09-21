---
title: Update Version of Redis Database in Google Kubernetes Engine (GKE)
date: "2023-09-20"
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
- redis
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Redis, PostgreSQL, MySQL, MongoDB, MariaDB, Elasticsearch, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will update version of Redis database in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Cluster
3) Insert Sample Data
4) Update Redis Database Version


### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
e5b4a1a0-5a67-4657-b390-db7200108bae
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
appscode/kubedb-ops-manager       	v0.22.0      	v0.22.8    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.08.18  	v2023.08.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.35.0      	v0.35.6    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.11.0      	v0.11.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.4.3      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.11.0      	v0.11.1    	KubeDB Webhook Server by AppsCode   

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
kubedb      kubedb-kubedb-autoscaler-5fcbf8f78-hslcv        1/1     Running   0          3m3s
kubedb      kubedb-kubedb-dashboard-6d8dc7bffc-nwgrw        1/1     Running   0          3m3s
kubedb      kubedb-kubedb-ops-manager-fd5c796bc-w8llt       1/1     Running   0          3m3s
kubedb      kubedb-kubedb-provisioner-7fc4796bf9-l8kvc      1/1     Running   0          3m3s
kubedb      kubedb-kubedb-schema-manager-95bbcf7b6-t6fgb    1/1     Running   0          3m3s
kubedb      kubedb-kubedb-webhook-server-656788f5bc-2fs7d   1/1     Running   0          3m3s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-09-20T09:05:06Z
elasticsearchdashboards.dashboard.kubedb.com      2023-09-20T09:04:53Z
elasticsearches.kubedb.com                        2023-09-20T09:04:10Z
elasticsearchopsrequests.ops.kubedb.com           2023-09-20T09:04:10Z
elasticsearchversions.catalog.kubedb.com          2023-09-20T09:03:25Z
etcds.kubedb.com                                  2023-09-20T09:04:23Z
etcdversions.catalog.kubedb.com                   2023-09-20T09:03:25Z
kafkas.kubedb.com                                 2023-09-20T09:04:25Z
kafkaversions.catalog.kubedb.com                  2023-09-20T09:03:25Z
mariadbautoscalers.autoscaling.kubedb.com         2023-09-20T09:05:06Z
mariadbdatabases.schema.kubedb.com                2023-09-20T09:04:34Z
mariadbopsrequests.ops.kubedb.com                 2023-09-20T09:04:23Z
mariadbs.kubedb.com                               2023-09-20T09:04:23Z
mariadbversions.catalog.kubedb.com                2023-09-20T09:03:25Z
memcacheds.kubedb.com                             2023-09-20T09:04:23Z
memcachedversions.catalog.kubedb.com              2023-09-20T09:03:25Z
mongodbautoscalers.autoscaling.kubedb.com         2023-09-20T09:05:06Z
mongodbdatabases.schema.kubedb.com                2023-09-20T09:04:33Z
mongodbopsrequests.ops.kubedb.com                 2023-09-20T09:04:13Z
mongodbs.kubedb.com                               2023-09-20T09:04:13Z
mongodbversions.catalog.kubedb.com                2023-09-20T09:03:25Z
mysqlautoscalers.autoscaling.kubedb.com           2023-09-20T09:05:06Z
mysqldatabases.schema.kubedb.com                  2023-09-20T09:04:32Z
mysqlopsrequests.ops.kubedb.com                   2023-09-20T09:04:20Z
mysqls.kubedb.com                                 2023-09-20T09:04:20Z
mysqlversions.catalog.kubedb.com                  2023-09-20T09:03:25Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-09-20T09:05:06Z
perconaxtradbopsrequests.ops.kubedb.com           2023-09-20T09:04:36Z
perconaxtradbs.kubedb.com                         2023-09-20T09:04:24Z
perconaxtradbversions.catalog.kubedb.com          2023-09-20T09:03:25Z
pgbouncers.kubedb.com                             2023-09-20T09:04:17Z
pgbouncerversions.catalog.kubedb.com              2023-09-20T09:03:25Z
postgresautoscalers.autoscaling.kubedb.com        2023-09-20T09:05:06Z
postgresdatabases.schema.kubedb.com               2023-09-20T09:04:33Z
postgreses.kubedb.com                             2023-09-20T09:04:25Z
postgresopsrequests.ops.kubedb.com                2023-09-20T09:04:30Z
postgresversions.catalog.kubedb.com               2023-09-20T09:03:25Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-09-20T09:05:06Z
proxysqlopsrequests.ops.kubedb.com                2023-09-20T09:04:33Z
proxysqls.kubedb.com                              2023-09-20T09:04:25Z
proxysqlversions.catalog.kubedb.com               2023-09-20T09:03:25Z
publishers.postgres.kubedb.com                    2023-09-20T09:04:46Z
redisautoscalers.autoscaling.kubedb.com           2023-09-20T09:05:06Z
redises.kubedb.com                                2023-09-20T09:04:25Z
redisopsrequests.ops.kubedb.com                   2023-09-20T09:04:26Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-09-20T09:05:07Z
redissentinelopsrequests.ops.kubedb.com           2023-09-20T09:04:39Z
redissentinels.kubedb.com                         2023-09-20T09:04:25Z
redisversions.catalog.kubedb.com                  2023-09-20T09:03:25Z
subscribers.postgres.kubedb.com                   2023-09-20T09:04:49Z
```

## Deploy Redis Cluster

Now we are going to deploy Redis cluster using KubeDB. First, let’s create a Namespace in which we will deploy the database.

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
  version: 6.0.6
  mode: Cluster
  cluster:
    master: 3
    replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `redis-cluster.yaml` 
Then create the above Redis CRD

```bash
$ kubectl create -f redis-cluster.yaml
redis.kubedb.com/sample-redis created
```

In this yaml,
* `spec.version` field specifies the version of Redis. Here, we are using Redis `version 6.0.6`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   1/1     Running   0          2m9s
pod/redis-cluster-shard0-1   1/1     Running   0          104s
pod/redis-cluster-shard1-0   1/1     Running   0          2m7s
pod/redis-cluster-shard1-1   1/1     Running   0          103s
pod/redis-cluster-shard2-0   1/1     Running   0          2m5s
pod/redis-cluster-shard2-1   1/1     Running   0          101s

NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/redis-cluster        ClusterIP   10.96.157.40   <none>        6379/TCP   2m11s
service/redis-cluster-pods   ClusterIP   None           <none>        6379/TCP   2m11s

NAME                                    READY   AGE
statefulset.apps/redis-cluster-shard0   2/2     2m9s
statefulset.apps/redis-cluster-shard1   2/2     2m7s
statefulset.apps/redis-cluster-shard2   2/2     2m5s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/redis-cluster   kubedb.com/redis   6.0.6     2m5s

NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   6.0.6     Ready    2m12s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME            VERSION   STATUS   AGE
redis-cluster   6.0.6     Ready    2m30s
```
> We have successfully deployed Redis in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `redis-cluster` that we have deployed. Let’s check them by following command,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=redis-cluster
NAME                   TYPE                       DATA   AGE
redis-cluster-auth     kubernetes.io/basic-auth   2      3m5s
redis-cluster-config   Opaque                     1      3m5s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=redis-cluster
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
redis-cluster        ClusterIP   10.96.157.40   <none>        6379/TCP   3m31s
redis-cluster-pods   ClusterIP   None           <none>        6379/TCP   3m31s
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
Defaulted container "redis" out of: redis, redis-init (init)

127.0.0.1:6379> set Product1 KubeDB
-> Redirected to slot [15299] located at 10.244.0.16:6379
OK

10.244.0.16:6379> set Product2 Stash
-> Redirected to slot [2976] located at 10.244.0.14:6379

OK
10.244.0.14:6379> get Product1
-> Redirected to slot [15299] located at 10.244.0.16:6379
"KubeDB"

10.244.0.16:6379> get Product2
-> Redirected to slot [2976] located at 10.244.0.14:6379
"Stash"
10.244.0.14:6379> exit
```

> We’ve successfully inserted some sample data to our database. More information about Production-Grade Redis on Kubernetes can be found [Redis Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

## Update Redis Database Version

In this section, we will update our Redis version from `6.0.6` to the latest version `7.0.9`. Let's check the current version,

```bash
$ kubectl get redis -n demo redis-cluster -o=jsonpath='{.spec.version}{"\n"}'
6.0.6
```

### Create RedisOpsRequest

In order to update the version of Redis cluster, we have to create a `RedisOpsRequest` CR with your desired version that is supported by KubeDB. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: redis-cluster
  updateVersion:
    targetVersion: 7.0.9
```

Let's save this yaml configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
redisopsrequest.ops.kubedb.com/update-version created
```

In this yaml,
* `spec.databaseRef.name` specifies that we are performing operation on `redis-cluster` Redis database.
* `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
* `spec.updateVersion.targetVersion` specifies the expected version of the database `7.0.9`.

### Verify the Updated Redis Version

`KubeDB` Enterprise operator will update the image of Redis object and related `StatefulSets` and `Pods`.
Let’s wait for `RedisOpsRequest` to be Successful. Run the following command to check `RedisOpsRequest` CR,

```bash
$ kubectl get redisopsrequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   3m
```

We can see from the above output that the `RedisOpsRequest` has succeeded.
Now, we are going to verify whether the Redis and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get redis -n demo redis-cluster -o=jsonpath='{.spec.version}{"\n"}'
7.0.9
```

> You can see from above, our Redis database has been updated with the new version `7.0.9`. So, the database update process is successfully completed.


If you want to learn more about Production-Grade Redis on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=wGXEPz8vSK1Qc12B&amp;list=PLoiT1Gv2KR1iSuQq_iyypzqvHW9u_un04" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).