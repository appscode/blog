---
title: Run Redis in Azure Kubernetes Service (AKS) Using KubeDB
date: "2022-08-03"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- azure-container
- azure-storage
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- redis
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are PostgreSQL, MySQL, MongoDB, MariaDB, Elasticsearch, Redis, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy Redis database in Azure Kubernetes Service (AKS). We will cover the following steps:

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
appscode/kubedb                   	v2022.05.24  	v2022.05.24	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.12.0      	v0.12.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.05.24  	v2022.05.24	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.05.24  	v2022.05.24	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.3.0       	v0.3.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.05.24  	v2022.05.24	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.05.24  	v2022.05.24	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.14.0      	v0.14.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.05.24  	v2022.05.24	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.27.0      	v0.27.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.3.0       	v0.3.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.9      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.3.0       	v0.3.0     	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.05.24 \
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
kubedb      kubedb-kubedb-autoscaler-57d5855f5f-zbqwl       1/1     Running   0          4m27s
kubedb      kubedb-kubedb-dashboard-85b89769-xbhf9          1/1     Running   0          4m27s
kubedb      kubedb-kubedb-ops-manager-9459b5dc4-5jvd4       1/1     Running   0          4m27s
kubedb      kubedb-kubedb-provisioner-f6cc7b44-c2vw2        1/1     Running   0          4m27s
kubedb      kubedb-kubedb-schema-manager-8584985c68-rrd8c   1/1     Running   0          4m27s
kubedb      kubedb-kubedb-webhook-server-6d86b94687-tvkwh   1/1     Running   0          4m27s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-07-22T12:33:30Z
elasticsearchdashboards.dashboard.kubedb.com      2022-07-22T12:33:18Z
elasticsearches.kubedb.com                        2022-07-22T12:33:12Z
elasticsearchopsrequests.ops.kubedb.com           2022-07-22T12:33:15Z
elasticsearchversions.catalog.kubedb.com          2022-07-22T12:30:28Z
etcds.kubedb.com                                  2022-07-22T12:33:12Z
etcdversions.catalog.kubedb.com                   2022-07-22T12:30:28Z
mariadbautoscalers.autoscaling.kubedb.com         2022-07-22T12:33:33Z
mariadbdatabases.schema.kubedb.com                2022-07-22T12:33:19Z
mariadbopsrequests.ops.kubedb.com                 2022-07-22T12:33:30Z
mariadbs.kubedb.com                               2022-07-22T12:33:12Z
mariadbversions.catalog.kubedb.com                2022-07-22T12:30:28Z
memcacheds.kubedb.com                             2022-07-22T12:33:12Z
memcachedversions.catalog.kubedb.com              2022-07-22T12:30:29Z
mongodbautoscalers.autoscaling.kubedb.com         2022-07-22T12:33:27Z
mongodbdatabases.schema.kubedb.com                2022-07-22T12:33:17Z
mongodbopsrequests.ops.kubedb.com                 2022-07-22T12:33:19Z
mongodbs.kubedb.com                               2022-07-22T12:33:13Z
mongodbversions.catalog.kubedb.com                2022-07-22T12:30:29Z
mysqldatabases.schema.kubedb.com                  2022-07-22T12:33:17Z
mysqlopsrequests.ops.kubedb.com                   2022-07-22T12:33:27Z
mysqls.kubedb.com                                 2022-07-22T12:33:13Z
mysqlversions.catalog.kubedb.com                  2022-07-22T12:30:29Z
perconaxtradbs.kubedb.com                         2022-07-22T12:33:13Z
perconaxtradbversions.catalog.kubedb.com          2022-07-22T12:30:29Z
pgbouncers.kubedb.com                             2022-07-22T12:33:13Z
pgbouncerversions.catalog.kubedb.com              2022-07-22T12:30:30Z
postgresdatabases.schema.kubedb.com               2022-07-22T12:33:18Z
postgreses.kubedb.com                             2022-07-22T12:33:13Z
postgresopsrequests.ops.kubedb.com                2022-07-22T12:33:37Z
postgresversions.catalog.kubedb.com               2022-07-22T12:30:30Z
proxysqlopsrequests.ops.kubedb.com                2022-07-22T12:33:40Z
proxysqls.kubedb.com                              2022-07-22T12:33:13Z
proxysqlversions.catalog.kubedb.com               2022-07-22T12:30:30Z
redises.kubedb.com                                2022-07-22T12:33:13Z
redisopsrequests.ops.kubedb.com                   2022-07-22T12:33:33Z
redissentinels.kubedb.com                         2022-07-22T12:33:14Z
redisversions.catalog.kubedb.com                  2022-07-22T12:30:31Z
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
    storageClassName: "default"
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
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/sample-redis-0   1/1     Running   0          26s

NAME                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/sample-redis        ClusterIP   10.0.16.209   <none>        6379/TCP   27s
service/sample-redis-pods   ClusterIP   None          <none>        6379/TCP   27s

NAME                            READY   AGE
statefulset.apps/sample-redis   1/1     28s

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/sample-redis   kubedb.com/redis   6.2.5     28s

NAME                            VERSION   STATUS   AGE
redis.kubedb.com/sample-redis   6.2.5     Ready    34s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME           VERSION   STATUS   AGE
sample-redis   6.2.5     Ready    79s
```
> We have successfully deployed Redis in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

#### Export the Credentials

KubeDB will create Secret and Service for the database `sample-redis` that we have deployed. Let’s check them by following command,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                  TYPE                       DATA   AGE
sample-redis-auth     kubernetes.io/basic-auth   2      2m10s
sample-redis-config   Opaque                     1      2m10s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
sample-redis        ClusterIP   10.0.16.209   <none>        6379/TCP   2m36s
sample-redis-pods   ClusterIP   None          <none>        6379/TCP   2m36s
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

> We've successfully inserted some sample data to our database. And this was just an example of our Redis Clustered database deployment. More information about Run & Manage Production-Grade Redis Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)


## Backup Redis Using Stash

Here, we are going to use Stash to backup the database we deployed before.

### Step 1: Install Stash

Kubedb Enterprise License works for Stash too.
So, we will use the Enterprise license that we have already obtained.

```bash
$ helm install stash appscode/stash             \
  --version v2022.07.09                  \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise --watch
NAMESPACE     NAME                                      READY   STATUS    RESTARTS   AGE
kube-system   stash-stash-enterprise-7f7d44ff7c-gftc8   2/2     Running   0          25s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-07-22T13:08:32Z
backupblueprints.stash.appscode.com       2022-07-22T13:08:32Z
backupconfigurations.stash.appscode.com   2022-07-22T13:08:31Z
backupsessions.stash.appscode.com         2022-07-22T13:08:31Z
functions.stash.appscode.com              2022-07-22T13:06:48Z
repositories.stash.appscode.com           2022-07-22T12:33:19Z
restorebatches.stash.appscode.com         2022-07-22T13:08:33Z
restoresessions.stash.appscode.com        2022-07-22T12:33:20Z
tasks.stash.appscode.com                  2022-07-22T13:06:49Z
```


### Step 2: Prepare Backend

Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes persistent volume like HostPath, PersistentVolumeClaim, NFS etc.

For this tutorial we are going to use Azure storage. You can find other setups [here](https://stash.run/docs/latest/guides/backends/overview/).

 ![My Empty azure storage](AzureStorageEmpty.png)

At first we need to create a secret so that we can access the Azure storage container. We can do that by the following code:

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-azure-storage-account-name>' > AZURE_ACCOUNT_NAME
$ echo -n '<your-azure-storage-account-key>' > AZURE_ACCOUNT_KEY
$ kubectl create secret generic -n demo azure-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./AZURE_ACCOUNT_NAME \
    --from-file=./AZURE_ACCOUNT_KEY
secret/azure-secret created
 ```

### Create Repository

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: azure-repo
  namespace: demo
spec:
  backend:
    azure:
      container: stash-backup
      prefix: /sample-redis
    storageSecretName: azure-secret
```

This repository CRO specifies the `azure-secret` we created before and stores the name and path to the azure storage container. It also specifies the location to the container where we want to backup our database.
> Here, My container name is `stash-backup`. Don't forget to change `spec.backend.azure.container` to your container name.

Lets create this repository,

```bash
$ kubectl apply -f azure-repo.yaml 
repository.stash.appscode.com/azure-repo created
```

### Create BackupConfiguration

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
    name: azure-repo
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
$ kubectl apply -f sample-redis-backup.yaml
backupconfiguration.stash.appscode.com/sample-redis-backup created
```

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the secret we created before called `azure-secret`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup.
* `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
* `spec.retentionPolicy` specifies the policy to follow for cleaning old snapshots. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/latest/guides/redis/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
sample-redis-backup-1658496001   BackupConfiguration   sample-redis-backup   Succeeded   10s        26s

$ kubectl get repository -n demo
NAME         INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
azure-repo   true        3.740 KiB   2                52s                      7m46s
```

Now if we check our azure storage container, we can see that the backup has been successful.

![AzureSuccess](AzureStorageSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up Redis Database using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20AKS).

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
    name: azure-repo
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
sample-redis-restore   azure-repo   Succeeded   4s         13s
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

> You can see the data has been restored. The recovery of Redis Database has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20AKS).

We have made an in depth video on How to Deploy Sharded Redis Cluster in Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/J7QI4EzuOVY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [PostgreSQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).