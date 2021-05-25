---
title: How to Manage Redis in Openshift Using KubeDB
date: 2021-04-23
weight: 22
authors:
  - Shohag Rana
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - mariadb
  - memcached
  - mongodb
  - mysql
  - postgresql
  - redis
  - kubedb
---

## Overview

The databases that KubeDB support are MongoDB, Elasticsearch, MySQL, MariaDB, PostgreSQL, Memcached and Redis. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy Redis database. We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Cluster
3) See the Automatic Failover feature

## Step 1: Installing KubeDB

We will follow the following sub-steps to install KubeDB.

### Step 1.1: Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ oc get ns kube-system -o=jsonpath='{.metadata.uid}'
08b1259c-5d51-4948-a2de-e2af8e6835a4 
```

### Step 1.2: Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![License Server](licenseserver.png)

### Step 1.3 Install KubeDB

We will use helm to install KubeDB.Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                        CHART VERSION APP VERSION DESCRIPTION
appscode/kubedb             v2021.04.16   v2021.04.16 KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler  v0.3.0        v0.3.0      KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog     v0.18.0       v0.18.0     KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community   v0.18.0       v0.18.0     KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds        v0.18.0       v0.18.0     KubeDB Custom Resource Definitions
appscode/kubedb-enterprise  v0.5.0        v0.5.0      KubeDB Enterprise by AppsCode - Enterprise feat...

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
    --version v2021.04.16 \
    --namespace kube-system \
    --set-file global.license=/path/to/the/license.txt \
    --set kubedb-enterprise.enabled=true \
    --set kubedb-autoscaler.enabled=true
```

Let's verify the installation:

```bash
$ watch oc get pods --all-namespaces -l "app.kubernetes
Every 2.0s: oc get pods --all-namespaces -l app.kubernetes.io/instance=kubedb                                                                                                      Shohag: Wed Apr 21 10:08:54 2021

NAMESPACE     NAME                                        READY   STATUS    RESTARTS   AGE
kube-system   kubedb-kubedb-autoscaler-569f66dbbc-qqmmb   1/1     Running   0          3m28s
kube-system   kubedb-kubedb-community-b6469fb9c-4hwbh     1/1     Running   0          3m28s
kube-system   kubedb-kubedb-enterprise-b658c95fc-kwqt6    1/1     Running   0          3m28s

```

We can see the CRD Groups that have been registered by the operator by running the following command:

```bash
$ oc get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2021-04-21T04:05:40Z
elasticsearches.kubedb.com                        2021-04-21T04:05:37Z
elasticsearchopsrequests.ops.kubedb.com           2021-04-21T04:05:37Z
elasticsearchversions.catalog.kubedb.com          2021-04-21T04:02:43Z
etcds.kubedb.com                                  2021-04-21T04:05:38Z
etcdversions.catalog.kubedb.com                   2021-04-21T04:02:44Z
mariadbs.kubedb.com                               2021-04-21T04:05:38Z
mariadbversions.catalog.kubedb.com                2021-04-21T04:02:44Z
memcacheds.kubedb.com                             2021-04-21T04:05:38Z
memcachedversions.catalog.kubedb.com              2021-04-21T04:02:45Z
mongodbautoscalers.autoscaling.kubedb.com         2021-04-21T04:05:37Z
mongodbopsrequests.ops.kubedb.com                 2021-04-21T04:05:40Z
mongodbs.kubedb.com                               2021-04-21T04:05:38Z
mongodbversions.catalog.kubedb.com                2021-04-21T04:02:46Z
mysqlopsrequests.ops.kubedb.com                   2021-04-21T04:05:48Z
mysqls.kubedb.com                                 2021-04-21T04:05:38Z
mysqlversions.catalog.kubedb.com                  2021-04-21T04:02:46Z
perconaxtradbs.kubedb.com                         2021-04-21T04:05:38Z
perconaxtradbversions.catalog.kubedb.com          2021-04-21T04:02:47Z
pgbouncers.kubedb.com                             2021-04-21T04:05:39Z
pgbouncerversions.catalog.kubedb.com              2021-04-21T04:02:47Z
postgreses.kubedb.com                             2021-04-21T04:05:39Z
postgresversions.catalog.kubedb.com               2021-04-21T04:02:48Z
proxysqls.kubedb.com                              2021-04-21T04:05:39Z
proxysqlversions.catalog.kubedb.com               2021-04-21T04:02:49Z
redises.kubedb.com                                2021-04-21T04:05:39Z
redisopsrequests.ops.kubedb.com                   2021-04-21T04:05:54Z
redisversions.catalog.kubedb.com                  2021-04-21T04:02:49Z
```

## Step 2: Deploying Database

Now we are going to Install Redis with the help of KubeDB.
At first, let's create a Namespace in which we will deploy the database.

```bash
$ oc create ns demo
```

Now, before deploying the Redis CRD let's perform some checks to ensure that it is deployed correctly.

### Check 1: StorageClass check

Let's check the availabe storage classes:

```bash
$ oc get storageclass
NAME         PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION
local-path   rancher.io/local-path   Delete          WaitForFirstConsumer   false    
```

Here, you can see that I have a storageclass named `local-path`. If you dont have a storage class you can run the following command:

```bash
$ oc apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

This will create the storage-class named local-path.

### Check 2: Correct Permissions

We need to ensure that the service account has correct permissions. To ensure correct permissions we should run:

```bash
$ oc adm policy add-scc-to-user privileged system:serviceaccount:local-path-storage:local-path-provisioner-service-account
```

This command will give the required permissions. </br>
Here is the yaml of the Redis CRD we are going to use:

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
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 200m
          memory: 300Mi
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into redis.yaml. Then apply using the command
`oc apply -f redis.yaml`

This yaml uses Redis CRD.

* In this yaml we can see in the `spec.version` field the version of Redis. You can change and get updated version by running `oc get redisversions` command.
* Another field to notice is the `spec.storagetype` field. This can be Durable or Ephemeral depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/v2021.04.16/guides/redis/concepts/redis/#specterminationpolicy).

### Deploy Redis CRD

Once these are handled correctly and the Redis CRD is deployed you will see that the following are created:

```bash
~ $ oc get all -n demo
NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   1/1     Running   0          27h
pod/redis-cluster-shard0-1   1/1     Running   0          27h
pod/redis-cluster-shard1-0   1/1     Running   0          27h
pod/redis-cluster-shard1-1   1/1     Running   0          27h
pod/redis-cluster-shard2-0   1/1     Running   0          27h
pod/redis-cluster-shard2-1   1/1     Running   0          27h

NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/redis-cluster        ClusterIP   10.217.5.200   <none>        6379/TCP   27h
service/redis-cluster-pods   ClusterIP   None           <none>        6379/TCP   27h

NAME                                    READY   AGE
statefulset.apps/redis-cluster-shard0   2/2     27h
statefulset.apps/redis-cluster-shard1   2/2     27h
statefulset.apps/redis-cluster-shard2   2/2     27h

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/redis-cluster   kubedb.com/redis   6.0.6     27h

NAME                             VERSION   STATUS   AGE
redis.kubedb.com/redis-cluster   6.0.6     Ready    27h
```

> We have successfully deployed Redis in OpenShift. Now we can exec into the container to use the database.

## Accessing Database Through CLI

To access the database through CLI we have to connect to any redis node. Then we

 ```bash
 # This command shows all the IP's of the redis pods
$ oc get pods --all-namespaces -o jsonpath='{range.items[*]}{.metadata.name} ---------- {.status.podIP}:6379{"\\n"}{end}' | grep redis
redis-cluster-shard0-0 ---------- 10.217.0.9:6379
redis-cluster-shard0-1 ---------- 10.217.0.28:6379
redis-cluster-shard1-0 ---------- 10.217.0.33:6379
redis-cluster-shard1-1 ---------- 10.217.0.43:6379
redis-cluster-shard2-0 ---------- 10.217.0.29:6379
redis-cluster-shard2-1 ---------- 10.217.0.21:6379


# This command shows the roles of each of the pods of Redis:
/data $ redis-cli -c cluster nodes
bb690f0802a203d8106139397febfa586c707b77 10.217.0.21:6379@16379 slave c9f383c2176a9da2fbda64bab379d0680a10d972 0 1621833286000 25 connected
a3bf9b4dbdf3b6877b81c462aa9ec3c0a651aa52 10.217.0.28:6379@16379 slave dcead4ade01632fe376466274345b0d9e846cfcf 0 1621833286000 23 connected
dcead4ade01632fe376466274345b0d9e846cfcf 10.217.0.9:6379@16379 myself,master - 0 1621833286000 23 connected 0-5460
e3f2085ee716bacf17a298081bfa29a1454a8a87 10.217.0.33:6379@16379 master - 0 1621833286000 21 connected 5461-10922
c9f383c2176a9da2fbda64bab379d0680a10d972 10.217.0.29:6379@16379 master - 0 1621833286912 25 connected 10923-16383
d9410e5a6f9d85b32aa8d4bfd73d7007be61b3c8 10.217.0.43:6379@16379 slave e3f2085ee716bacf17a298081bfa29a1454a8a87 0 1621833285007 21 connected

# connect to any node
~ $ oc exec -it redis-cluster-shard0-0 -n demo -c redis -- sh

# connect to any master pod
/data $ redis-cli -c -h 10.217.0.108

# set key 'hello' to value 'world'
10.217.0.108:6379> set hello world
OK
10.217.0.108:6379> get hello
"world"
10.217.0.108:6379> exit
 ```

Now we have entered into the Redis CLI and we can create and delete as we want.
redis stores data as key value pair. In the above commands, we set hello to "world".

> This was just one example of database deployment. The other databases that KubeDB suport are MySQL, Postgres, Elasticsearch, MongoDB and MariaDB. The tutorials on how to deploy these into the cluster can be found [HERE](https://kubedb.com/)

## Redis Clustering Features

There are 2 main features of Clustering which are `Data Availability` and `Automatic Failover`. These are shown below:

### Data Availability

In this section, we will see whether we can get the data from any other node (any master or replica) or not.
We can notice the replication of data among the other pods of Redis:

```bash
# switch the connection to the replica of the current master and get the data
/data $ redis-cli -c -h 10.217.0.28
10.217.0.28:6379> get hello
-> Redirected to slot [866] located at 10.217.0.9:6379
"world"
10.217.0.9:6379> exit

# switch the connection to any other node
# get the data
/data $ redis-cli -c -h 10.217.0.43
10.217.0.43:6379> get hello
-> Redirected to slot [866] located at 10.217.0.9:6379
"world"
```

### Automatic Failover

To test automatic failover, we will force a master node to restart. Since the master node (`pod`) becomes unavailable, the rest of the members will elect a replica (one of its replica in case of more than one replica under this master) of this master node as the new master. When the old master comes back, it will join the cluster as the new replica of the new master.

```bash
# connect to any node and get the master nodes info
$ oc exec -it redis-cluster-shard0-0 -n demo -c redis -- sh
/data $ redis-cli -c cluster nodes | grep master
dcead4ade01632fe376466274345b0d9e846cfcf 10.217.0.9:6379@16379 myself,master - 0 1621916192000 23 connected 0-5460
e3f2085ee716bacf17a298081bfa29a1454a8a87 10.217.0.33:6379@16379 master - 0 1621916193064 21 connected 5461-10922
c9f383c2176a9da2fbda64bab379d0680a10d972 10.217.0.29:6379@16379 master - 0 1621916193566 25 connected 10923-16383

# let's crash node 10.217.0.9 with the `DEBUG SEGFAULT` command
/data $ redis-cli -h 10.217.0.9 debug segfault
Error: Server closed the connection
command terminated with exit code 137

# now again connect to a node and get the master nodes info
$ oc exec -it redis-cluster-shard0-0 -n demo -c redis -- sh
/data $ redis-cli -c cluster nodes | grep master
e3f2085ee716bacf17a298081bfa29a1454a8a87 10.217.0.33:6379@16379 master - 0 1621931346000 21 connected 5461-10922
dcead4ade01632fe376466274345b0d9e846cfcf 10.217.0.28:6379@16379 master - 0 1621931347000 27 connected 0-5460
c9f383c2176a9da2fbda64bab379d0680a10d972 10.217.0.29:6379@16379 master - 0 1621931347000 25 connected 10923-16383
```

Notice that 10.217.0.28 is the new master and 10.217.0.9 is the replica of 10.217.0.28. This means that the replica has noe become the master node since the previous master node crashed. Here, we notice that there has been a successful recovery from failover.

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
