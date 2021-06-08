---
title: How to Manage Memcached in Openshift Using KubeDB
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
In this tutorial we will deploy Memcached database. We will cover the following steps:

1) Install KubeDB
2) Deploy Memcached Cluster

## Install KubeDB

We will follow the following sub-steps to install KubeDB.

### Step 1: Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ oc get ns kube-system -o=jsonpath='{.metadata.uid}'
08b1259c-5d51-4948-a2de-e2af8e6835a4 
```

### Step 2: Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![License Server](licenseserver.png)

### Step 3 Install KubeDB

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
$ watch oc get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
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

## Deploy Memcached Cluster

Now we are going to Install Memcached with the help of KubeDB.
At first, let's create a Namespace in which we will deploy the database.

```bash
$ oc create ns demo
```

Now, before deploying the Memcached CRD let's perform some checks to ensure that it is deployed correctly.

### Check 1: StorageClass Check

Let's check the availabe storage classes:

```bash
$ oc get storageclass
NAME         PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION
local-path   rancher.io/local-path   Delete          WaitForFirstConsumer   false    
```

Here, we can see that I have a storageclass named `local-path`. If you do not have a storage class you can run the following command:

```bash
$ oc apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

This will create the storage-class named local-path.

### Check 2: Correct Permissions

We can ensure that the service account has correct permissions by running the following command:

```bash
$ oc adm policy add-scc-to-user privileged system:serviceaccount:local-path-storage:local-path-provisioner-service-account
```

OpenShift has Security Context Constraints for which the MariaDB CRD is restricted to be deployed. The above command will give the required permissions. </br>

### Deploy Memcached CRD

Now, let's have a look into the yaml of the Memcached CRD we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "1.5.4-v1"
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into memcached.yaml. Then apply using the command
`oc apply -f memcached.yaml`

* In this yaml we can see in the `spec.version` field the version of Memcached. You can change and get updated version by running `oc get memcachedversions` command.
* Another field to notice is the `spec.storagetype` field. This can be Durable or Ephemeral depending on the requirements of the database to be persistent or not.
* `spec.storage.storageClassName` contains the name of the storage class we obtained before named "local-path".
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/v2021.04.16/guides/memcached/concepts/memcached/#specterminationpolicy).

Once these are handled correctly and the Memcached CRD is deployed you will see that the following are created:

```bash
$ oc get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/memcd-quickstart-0   1/1     Running   0          70s
pod/memcd-quickstart-1   1/1     Running   0          66s
pod/memcd-quickstart-2   1/1     Running   0          63s

NAME                            TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
service/memcd-quickstart        ClusterIP   10.217.4.2   <none>        11211/TCP   71s
service/memcd-quickstart-pods   ClusterIP   None         <none>        11211/TCP   71s

NAME                                READY   AGE
statefulset.apps/memcd-quickstart   3/3     73s

NAME                                                  TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/memcd-quickstart   kubedb.com/memcached   1.5.4     66s

NAME                                    VERSION    STATUS   AGE
memcached.kubedb.com/memcd-quickstart   1.5.4-v1   Ready    82s
```

> We have successfully deployed Memcached in OpenShift. Now we can exec into the container to use the database.

### Accessing Database Through CLI

Now, you can connect to this Memcached cluster using telnet. Here, we will connect to Memcached server from local-machine through port-forwarding.

 ```bash
$ oc port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
Handling connection for 11211
 ```

Now in a new terminal lets connect to the pod with telnet:

```bash
$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
set my_key 0 2592000 1
2
STORED

# Meaning:
# 0       => no flags
# 2592000 => TTL (Time-To-Live) in [s]
# 1       => size in byte
# 2       => value

# Now let's view the stored data
get my_key
VALUE my_key 0 1
2
END

quit
Connection closed by foreign host.
```

> This was just one example of database deployment. The other databases that KubeDB support are MySQL, Postgres, Elasticsearch, MongoDB and MariaDB. The tutorials on how to deploy these into the cluster can be found [HERE](https://kubedb.com/)

## Memcached Clustering Features

There are 2 main features of Clustering which are `Data Availability` and `Automatic Failover`. These are shown in the following sections.

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

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
