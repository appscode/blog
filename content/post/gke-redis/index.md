---
title: Run & Manage Redis in Google Kubernetes Engine (GKE) Using KubeDB
date: "2022-03-01"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- gcs
- gke
- google-cloud
- google-kubernetes-engine
- kubedb
- kubernetes
- redis
---

## Overview

The databases that KubeDB supports are Redis, MariaDB, MySQL, Elasticsearch, MongoDB, PostgreSQL, Percona XtraDB, ProxySQL, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy Redis database. We will cover the following steps:

1) Install KubeDB
2) Deploy Standalone Redis Database
3) Install Stash
4) Backup Redis Using Stash
5) Recover Redis Using Stash

## Install KubeDB

We will follow the steps to install KubeDB.

### Step 1: Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d 
```

### Step 2: Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![License Server](AppscodeLicense.png)

### Step 3: Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2022.02.22  	v2022.02.22	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.10.0      	v0.10.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.02.22  	v2022.02.22	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.02.22  	v2022.02.22	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.1.0       	v0.1.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.02.22  	v2022.02.22	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.02.22  	v2022.02.22	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.12.0      	v0.12.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.02.22  	v2022.02.22	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.25.0      	v0.25.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.1.0       	v0.1.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.1.0       	v0.1.0     	KubeDB Webhook Server by AppsCode

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.02.22 \
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
kubedb      kubedb-kubedb-autoscaler-7d47dc6fd7-vh4ws       1/1     Running   0          4m42s
kubedb      kubedb-kubedb-dashboard-68947d8b45-xhhdc        1/1     Running   0          4m43s
kubedb      kubedb-kubedb-ops-manager-79cbf68459-jnqfd      1/1     Running   0          4m42s
kubedb      kubedb-kubedb-provisioner-7856dcc597-qg5s7      1/1     Running   0          4m43s
kubedb      kubedb-kubedb-schema-manager-5cc5fdd9c6-6tbs2   1/1     Running   0          4m43s
kubedb      kubedb-kubedb-webhook-server-64f544ff6-5qhs4    1/1     Running   0          4m43s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-02-28T05:08:33Z
elasticsearchdashboards.dashboard.kubedb.com      2022-02-28T05:08:37Z
elasticsearches.kubedb.com                        2022-02-28T05:08:38Z
elasticsearchopsrequests.ops.kubedb.com           2022-02-28T05:09:27Z
elasticsearchversions.catalog.kubedb.com          2022-02-28T04:47:11Z
etcds.kubedb.com                                  2022-02-28T05:08:52Z
etcdversions.catalog.kubedb.com                   2022-02-28T04:47:12Z
mariadbautoscalers.autoscaling.kubedb.com         2022-02-28T05:08:36Z
mariadbdatabases.schema.kubedb.com                2022-02-28T05:08:36Z
mariadbopsrequests.ops.kubedb.com                 2022-02-28T05:09:07Z
mariadbs.kubedb.com                               2022-02-28T05:08:37Z
mariadbversions.catalog.kubedb.com                2022-02-28T04:47:12Z
memcacheds.kubedb.com                             2022-02-28T05:08:54Z
memcachedversions.catalog.kubedb.com              2022-02-28T04:47:13Z
mongodbautoscalers.autoscaling.kubedb.com         2022-02-28T05:08:29Z
mongodbdatabases.schema.kubedb.com                2022-02-28T05:08:31Z
mongodbopsrequests.ops.kubedb.com                 2022-02-28T05:08:40Z
mongodbs.kubedb.com                               2022-02-28T05:08:32Z
mongodbversions.catalog.kubedb.com                2022-02-28T04:47:13Z
mysqldatabases.schema.kubedb.com                  2022-02-28T05:08:29Z
mysqlopsrequests.ops.kubedb.com                   2022-02-28T05:09:01Z
mysqls.kubedb.com                                 2022-02-28T05:08:30Z
mysqlversions.catalog.kubedb.com                  2022-02-28T04:47:13Z
perconaxtradbs.kubedb.com                         2022-02-28T05:09:01Z
perconaxtradbversions.catalog.kubedb.com          2022-02-28T04:47:14Z
pgbouncers.kubedb.com                             2022-02-28T05:08:52Z
pgbouncerversions.catalog.kubedb.com              2022-02-28T04:47:14Z
postgresdatabases.schema.kubedb.com               2022-02-28T05:08:34Z
postgreses.kubedb.com                             2022-02-28T05:08:35Z
postgresopsrequests.ops.kubedb.com                2022-02-28T05:09:20Z
postgresversions.catalog.kubedb.com               2022-02-28T04:47:14Z
proxysqls.kubedb.com                              2022-02-28T05:09:04Z
proxysqlversions.catalog.kubedb.com               2022-02-28T04:47:15Z
redises.kubedb.com                                2022-02-28T05:09:04Z
redisopsrequests.ops.kubedb.com                   2022-02-28T05:09:11Z
redissentinels.kubedb.com                         2022-02-28T05:09:05Z
redisversions.catalog.kubedb.com                  2022-02-28T04:47:16Z
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
  version: "6.2.5"
  storageType: Durable
  storage:
    storageClassName: "standard"
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

* In this yaml we can see in the `spec.version` field specifies the version of Redis. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/v2022.02.22/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/sample-redis-0   1/1     Running   0          2m11s

NAME                        TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/sample-redis        ClusterIP   10.8.1.16    <none>        6379/TCP   2m11s
service/sample-redis-pods   ClusterIP   None         <none>        6379/TCP   2m11s

NAME                            READY   AGE
statefulset.apps/sample-redis   1/1     2m12s

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/sample-redis   kubedb.com/redis   6.2.5     2m14s

NAME                            VERSION   STATUS   AGE
redis.kubedb.com/sample-redis   6.2.5     Ready    2m22s
```
Let’s check if the database is ready to use,

```bash
kubectl get redis -n demo
NAME           VERSION   STATUS   AGE
sample-redis   6.2.5     Ready    4m
```
> We have successfully deployed Redis in GKE. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `sample-redis` that we have deployed. Let’s check them by following command,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                  TYPE                       DATA   AGE
sample-redis-auth     kubernetes.io/basic-auth   2      7m49s
sample-redis-config   Opaque                     1      7m49s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
sample-redis        ClusterIP   10.8.1.16    <none>        6379/TCP   7m52s
sample-redis-pods   ClusterIP   None         <none>        6379/TCP   7m52s

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

> This was just an example of our Redis database deployment. The other databases that KubeDB supports are Elasticsearch, MongoDB, MySQL, PostgreSQL, MariaDB, Percona XtraDB, ProxySQL, Memcached and PgBouncer. The tutorials on how to deploy these into the cluster can be found [HERE](https://kubedb.com/)

## Backup Redis Using Stash

Here, we are going to use Stash to backup the database we deployed before.

### Step 1: Install Stash

Go to [Appscode License Server](https://license-issuer.appscode.com/) again to get the Stash Enterprise license.
Here, we will use the Stash Enterprise license that we obtained.

```bash
$ helm install stash appscode/stash             \
  --version v2022.02.22                         \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise --watch
NAMESPACE     NAME                                      READY   STATUS    RESTARTS   AGE
kube-system   stash-stash-enterprise-5fd6c4bf8c-j9bpd   2/2     Running   0          5m43s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-02-28T11:05:11Z
backupblueprints.stash.appscode.com       2022-02-28T11:05:11Z
backupconfigurations.stash.appscode.com   2022-02-28T11:05:09Z
backupsessions.stash.appscode.com         2022-02-28T11:05:09Z
functions.stash.appscode.com              2022-02-28T11:02:54Z
repositories.stash.appscode.com           2022-02-28T05:08:41Z
restorebatches.stash.appscode.com         2022-02-28T11:05:11Z
restoresessions.stash.appscode.com        2022-02-28T05:08:43Z
tasks.stash.appscode.com                  2022-02-28T11:02:55Z
```


### Step 2: Prepare Backend

Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes persistent volume like HostPath, PersistentVolumeClaim, NFS etc.

For this tutorial we are going to use gcs-bucket. You can find other setups [here](https://stash.run/docs/v2022.02.22/guides/backends/overview/).

 ![My Empty GCS bucket](GoogleCloudEmpty.png)

At first we need to create a secret so that we can access the gcs bucket. We can do that by the following code:

```bash
$ echo -n 'YOURPASSWORD' > RESTIC_PASSWORD
$ echo -n 'YOURPROJECTNAME' > GOOGLE_PROJECT_ID
$ cat /PATH/TO/JSONKEY.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
        --from-file=./RESTIC_PASSWORD              \
        --from-file=./GOOGLE_PROJECT_ID            \
        --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
 ```

### Step 3: Create Repository

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-backup-dipta
      prefix: /sample-redis
    storageSecretName: gcs-secret
```

This repository CRD specifies the `gcs-secret` we created before and stores the name and path to the gcs-bucket. It also specifies the location in the bucket where we want to backup our database.
> Here, My bucket name is `stash-backup-dipta`. Don't forget to change `spec.backend.gcs.bucket` to your bucket name.

### Step 4: Create BackupConfiguration

Now, we need to create a `BackupConfiguration` file that specifies what to backup, where to backup and when to backup.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: gcs-repo
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

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the secret we created before called `gcs-secret`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/v2022.02.22/guides/redis/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
sample-redis-backup-1646052901   BackupConfiguration   sample-redis-backup   Succeeded   14s        85s

$ kubectl get repository -n demo
NAME       INTEGRITY   SIZE    SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo   true        317 B   1                115s                     9m
```

Now if we check our GCS bucket we can see that the backup has been successful.

![gcsSuccess](GoogleCloudSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up Redis using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20GKE).

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
  name: sample-redis-restore
  namespace: demo
spec:
  repository:
    name: gcs-repo
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
$ kubectl create -f sample-redis-restore.yaml
restoresession.stash.appscode.com/sample-redis-restore created
```

This `RestoreSession` specifies where the data will be restored.
Once this is applied, a `RestoreSession` will be created. Once it has succeeded, the database has been successfully recovered as you can see below:

```bash
$ kubectl get restoresession -n demo
NAME                   REPOSITORY   PHASE       DURATION   AGE
sample-redis-restore   gcs-repo     Succeeded   8s         10m
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

> You can see the data has been restored. The recovery of Redis Database has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20GKE).

We have made an in depth video on how to Run & Manage production-grade Redis in Kubernetes cluster using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/z-Kg_aX828I" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
