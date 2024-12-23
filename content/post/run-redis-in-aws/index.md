---
title: Run Redis in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB
date: "2022-10-31"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- redis
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Redis, PostgreSQL, MySQL, MongoDB, MariaDB, Elasticsearch, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy Redis in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Redis Standalone Database
3) Install Stash
4) Backup Redis Database Using Stash
5) Recover Redis Database Using Stash

## Install KubeDB

We will follow the steps to install KubeDB.

### Step 1: Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
e5b4a1a0-5a67-4657-b390-db7200108bae
```

### Step 2: Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial, we will use KubeDB Enterprise Edition.

![License Server](AppscodeLicense.png)

### Step 3: Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2022.10.18  	v2022.10.18	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.14.0      	v0.14.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.10.18  	v2022.10.18	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.10.18  	v2022.10.18	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.5.0       	v0.5.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.10.18  	v2022.10.18	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.10.18  	v2022.10.18	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.16.0      	v0.16.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.10.18  	v2022.10.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.29.0      	v0.29.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.5.0       	v0.5.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.9      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.5.0       	v0.5.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.10.18 \
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
kubedb      kubedb-kubedb-autoscaler-8449d656f9-wmfrl       1/1     Running   0          2m
kubedb      kubedb-kubedb-dashboard-d48976bc4-v2gvr         1/1     Running   0          2m
kubedb      kubedb-kubedb-ops-manager-7965b65d68-fndmq      1/1     Running   0          2m
kubedb      kubedb-kubedb-provisioner-6b8fbd5979-7zkld      1/1     Running   0          2m
kubedb      kubedb-kubedb-schema-manager-68895588ff-m97p5   1/1     Running   0          2m
kubedb      kubedb-kubedb-webhook-server-7667cdc5d4-k6b8f   1/1     Running   0          2m
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-10-28T06:39:01Z
elasticsearchdashboards.dashboard.kubedb.com      2022-10-28T06:38:59Z
elasticsearches.kubedb.com                        2022-10-28T06:38:59Z
elasticsearchopsrequests.ops.kubedb.com           2022-10-28T06:39:06Z
elasticsearchversions.catalog.kubedb.com          2022-10-28T06:31:58Z
etcds.kubedb.com                                  2022-10-28T06:39:09Z
etcdversions.catalog.kubedb.com                   2022-10-28T06:31:59Z
mariadbautoscalers.autoscaling.kubedb.com         2022-10-28T06:39:02Z
mariadbdatabases.schema.kubedb.com                2022-10-28T06:39:02Z
mariadbopsrequests.ops.kubedb.com                 2022-10-28T06:39:24Z
mariadbs.kubedb.com                               2022-10-28T06:39:03Z
mariadbversions.catalog.kubedb.com                2022-10-28T06:32:00Z
memcacheds.kubedb.com                             2022-10-28T06:39:10Z
memcachedversions.catalog.kubedb.com              2022-10-28T06:32:01Z
mongodbautoscalers.autoscaling.kubedb.com         2022-10-28T06:39:02Z
mongodbdatabases.schema.kubedb.com                2022-10-28T06:39:00Z
mongodbopsrequests.ops.kubedb.com                 2022-10-28T06:39:10Z
mongodbs.kubedb.com                               2022-10-28T06:39:00Z
mongodbversions.catalog.kubedb.com                2022-10-28T06:32:02Z
mysqlautoscalers.autoscaling.kubedb.com           2022-10-28T06:39:02Z
mysqldatabases.schema.kubedb.com                  2022-10-28T06:38:59Z
mysqlopsrequests.ops.kubedb.com                   2022-10-28T06:39:21Z
mysqls.kubedb.com                                 2022-10-28T06:38:59Z
mysqlversions.catalog.kubedb.com                  2022-10-28T06:32:02Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2022-10-28T06:39:02Z
perconaxtradbopsrequests.ops.kubedb.com           2022-10-28T06:39:39Z
perconaxtradbs.kubedb.com                         2022-10-28T06:39:14Z
perconaxtradbversions.catalog.kubedb.com          2022-10-28T06:32:03Z
pgbouncers.kubedb.com                             2022-10-28T06:39:14Z
pgbouncerversions.catalog.kubedb.com              2022-10-28T06:32:04Z
postgresautoscalers.autoscaling.kubedb.com        2022-10-28T06:39:02Z
postgresdatabases.schema.kubedb.com               2022-10-28T06:39:01Z
postgreses.kubedb.com                             2022-10-28T06:39:02Z
postgresopsrequests.ops.kubedb.com                2022-10-28T06:39:32Z
postgresversions.catalog.kubedb.com               2022-10-28T06:32:05Z
proxysqlopsrequests.ops.kubedb.com                2022-10-28T06:39:35Z
proxysqls.kubedb.com                              2022-10-28T06:39:15Z
proxysqlversions.catalog.kubedb.com               2022-10-28T06:32:06Z
publishers.postgres.kubedb.com                    2022-10-28T06:39:46Z
redisautoscalers.autoscaling.kubedb.com           2022-10-28T06:39:03Z
redises.kubedb.com                                2022-10-28T06:39:15Z
redisopsrequests.ops.kubedb.com                   2022-10-28T06:39:28Z
redissentinelautoscalers.autoscaling.kubedb.com   2022-10-28T06:39:03Z
redissentinelopsrequests.ops.kubedb.com           2022-10-28T06:39:43Z
redissentinels.kubedb.com                         2022-10-28T06:39:16Z
redisversions.catalog.kubedb.com                  2022-10-28T06:32:07Z
subscribers.postgres.kubedb.com                   2022-10-28T06:39:49Z
```

## Deploy Standalone Redis Database

Now we are going to Install Redis with the help of KubeDB.
At first, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create ns demo
namespace/demo created
```

Here is the yaml of the Redis CRD we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: "7.0.5"
  storageType: Durable
  storage:
    storageClassName: "gp2"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's save this yaml configuration into `sample-redis.yaml` 
Then create the above Redis CRD

```bash
$ kubectl create -f sample-redis.yaml
redis.kubedb.com/sample-redis created
```

* In this yaml we can see in the `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.0.5`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/sample-redis-0   1/1     Running   0          17s

NAME                        TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/sample-redis        ClusterIP   10.100.116.125   <none>        6379/TCP   24s
service/sample-redis-pods   ClusterIP   None             <none>        6379/TCP   25s

NAME                            READY   AGE
statefulset.apps/sample-redis   1/1     27s

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/sample-redis   kubedb.com/redis   7.0.5     36s

NAME                            VERSION   STATUS   AGE
redis.kubedb.com/sample-redis   7.0.5     Ready    70s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME           VERSION   STATUS   AGE
sample-redis   7.0.5     Ready    100s
```
> We have successfully deployed Redis in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `sample-redis` that we have deployed. Let’s check them by following command,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                  TYPE                       DATA   AGE
sample-redis-auth     kubernetes.io/basic-auth   2      2m17s
sample-redis-config   Opaque                     1      2m17s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
sample-redis        ClusterIP   10.100.116.125   <none>        6379/TCP   2m36s
sample-redis-pods   ClusterIP   None             <none>        6379/TCP   2m37s
```
Now, we are going to use `PASSWORD` to authenticate and insert some sample data.
At first, let’s export the `PASSWORD` as environment variables to make further commands re-usable.

```bash
$ export PASSWORD=$(kubectl get secrets -n demo sample-redis-auth -o jsonpath='{.data.\password}' | base64 -d)
```

#### Insert Sample Data

In this section, we are going to login into our Redis database pod and insert some sample data. 

```bash
$ kubectl exec -it -n demo sample-redis-0 -- redis-cli -a $PASSWORD
127.0.0.1:6379> set Product1 KubeDB
OK
127.0.0.1:6379> set Product2 Stash
OK
127.0.0.1:6379> get Product1
"KubeDB"
127.0.0.1:6379> get Product2
"Stash"
127.0.0.1:6379> exit

```

> We've successfully inserted some sample data to our database. And this was just an example of our Redis database deployment. More information about Run & Manage Production-Grade Redis Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)


## Backup Redis Using Stash

Here, we are going to use Stash to backup the database we deployed before.

### Step 1: Install Stash

Kubedb Enterprise License works for Stash too.
So, we will use the Enterprise license that we have already obtained.

```bash
$ helm install stash appscode/stash             \
  --version v2022.09.29                  \
  --namespace stash --create-namespace \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise
NAMESPACE   NAME                                      READY   STATUS    RESTARTS   AGE
stash       stash-stash-enterprise-59b84bb9b5-7n8t8   2/2     Running   0          62s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-10-28T08:59:58Z
backupblueprints.stash.appscode.com       2022-10-28T08:59:58Z
backupconfigurations.stash.appscode.com   2022-10-28T08:59:57Z
backupsessions.stash.appscode.com         2022-10-28T08:59:57Z
functions.stash.appscode.com              2022-10-28T08:55:45Z
repositories.stash.appscode.com           2022-10-28T06:39:05Z
restorebatches.stash.appscode.com         2022-10-28T08:59:59Z
restoresessions.stash.appscode.com        2022-10-28T06:39:05Z
tasks.stash.appscode.com                  2022-10-28T08:55:47Z
```


### Prepare Backend

Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes native resources like HostPath, PersistentVolumeClaim etc. or NFS.

For this tutorial we are going to use AWS S3 storage. You can find other setups [here](https://stash.run/docs/latest/guides/backends/overview/).

 ![My Empty AWS storage](AWSStorageEmpty.png)

At first we need to create a secret so that we can access the AWS S3 storage bucket. We can do that by the following code:

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

### Create Repository

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: s3-repo
  namespace: demo
spec:
  backend:
    s3:
      endpoint: s3.amazonaws.com
      bucket: stash-qa
      region: us-east-1
      prefix: /redis-backup
    storageSecretName: s3-secret
```

This repository CRO specifies the `s3-secret` we created before and stores the name and path to the AWS storage bucket. It also specifies the location to the container where we want to backup our database.
> Here, My bucket name is `stash-qa`. Don't forget to change `spec.backend.s3.bucket` to your bucket name and For `S3`, use `s3.amazonaws.com` as endpoint.

Lets create this repository,

```bash
$ kubectl apply -f s3-repo.yaml
repository.stash.appscode.com/s3-repo created
```

### Create BackupConfiguration

Now, we need to create a `BackupConfiguration` file that specifies what to backup, where to backup and when to backup.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: redis-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-redis
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```
Create this `BackupConfiguration` by following command,

```bash
$ kubectl apply -f redis-backup.yaml
backupconfiguration.stash.appscode.com/redis-backup created
```

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the secret we created before called `s3-secret`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup.
* `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
* `spec.retentionPolicy` specifies the policy to follow for cleaning old snapshots. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/latest/guides/redis/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                      INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
redis-backup-1666948445   BackupConfiguration   redis-backup   Succeeded   15s        27s

$ kubectl get repository -n demo
NAME      INTEGRITY   SIZE    SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
s3-repo   true        511 B   1                50s                      5m3s
```

Now if we check our Amazon S3 bucket, we can see that the backup has been successful.

![AWSSuccess](AWSStorageSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up Redis Database using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20AWS).

## Recover Redis Using Stash

Let's think of a scenario in which the database has been accidentally deleted or there was an error in the database causing it to crash.

#### Temporarily pause backup

At first, let’s stop taking any further backup of the database so that no backup runs after we delete the sample data. We are going to pause the `BackupConfiguration` object. Stash will stop taking any further backup when the `BackupConfiguration` is paused.

```bash
$ kubectl patch backupconfiguration -n demo sample-redis-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-redis-backup patched
```

Now, we are going to delete those data to simulate accidental database deletion.

```bash
$ kubectl exec -it -n demo sample-redis-0 -- redis-cli -a $PASSWORD
127.0.0.1:6379> get Product1
"KubeDB"
127.0.0.1:6379> get Product2
"Stash"
127.0.0.1:6379> del Product1
(integer) 1
127.0.0.1:6379> del Product2
(integer) 1
127.0.0.1:6379> get Product1
(nil)
127.0.0.1:6379> get Product2
(nil)
127.0.0.1:6379> exit
```

### Step 1: Create a RestoreSession

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: redis-restore
  namespace: demo
spec:
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-redis
  rules:
    - snapshots: [latest]
```

Now, let's create `RestoreSession` that will initiate restoring from the cloud.

```bash
$ kubectl create -f redis-restore.yaml
restoresession.stash.appscode.com/redis-restore created
```

This `RestoreSession` specifies where the data will be restored.
Once this is applied, a `RestoreSession` will be created. Once it has succeeded, the database has been successfully recovered as you can see below:

```bash
$ kubectl get restoresession -n demo
NAME            REPOSITORY   PHASE       DURATION   AGE
redis-restore   s3-repo      Succeeded   5s         21s
```

Now, let's check whether the data has been correctly restored:

```bash
$ kubectl exec -it -n demo sample-redis-0 -- redis-cli -a $PASSWORD
127.0.0.1:6379> get Product1
"KubeDB"
127.0.0.1:6379> get Product2
"Stash"
127.0.0.1:6379> exit
```

> You can see the data has been restored. The recovery of Redis Database has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20AWS).

We have made an in depth video on Redis Sentinel Ops Requests - Day 2 Lifecycle Management for Redis Sentinel Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/LToGVt1-D50" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).