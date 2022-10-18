---
title: Run PostgreSQL in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB
date: "2022-10-18"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- s3
- cloud-native
- database
- eks
- kubedb
- kubernetes
- postgresql
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are PostgreSQL, MySQL, MongoDB, MariaDB, Elasticsearch, Redis, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy PostgreSQL database in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy PostgreSQL Clustered Database
3) Install Stash
4) Backup PostgreSQL Database Using Stash
5) Recover PostgreSQL Database Using Stash

## Install KubeDB

We will follow the steps to install KubeDB.

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
60b010fb-9ad6-4ac6-89f4-7321e697f469
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![License Server](AppscodeLicense.png)

### Install KubeDB

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
appscode/kubedb-ops-manager       	v0.16.0      	v0.16.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.10.18  	v2022.10.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.29.0      	v0.29.0    	KubeDB Provisioner by AppsCode - Community feat...
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
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-fbc6bc99b-dq4lw        1/1     Running   0          3m32s
kubedb      kubedb-kubedb-dashboard-5df5d847d5-7qzms        1/1     Running   0          3m32s
kubedb      kubedb-kubedb-ops-manager-55748d5bc4-d6jnw      1/1     Running   0          3m32s
kubedb      kubedb-kubedb-provisioner-69d5857b87-zhsjj      1/1     Running   0          3m32s
kubedb      kubedb-kubedb-schema-manager-5bff669b5f-vqvhl   1/1     Running   1          3m32s
kubedb      kubedb-kubedb-webhook-server-5bd8d49d84-jqzfj   1/1     Running   0          3m32s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-10-17T08:53:53Z
elasticsearchdashboards.dashboard.kubedb.com      2022-10-17T08:53:52Z
elasticsearches.kubedb.com                        2022-10-17T08:53:52Z
elasticsearchopsrequests.ops.kubedb.com           2022-10-17T08:53:57Z
elasticsearchversions.catalog.kubedb.com          2022-10-17T08:42:56Z
etcds.kubedb.com                                  2022-10-17T08:53:55Z
etcdversions.catalog.kubedb.com                   2022-10-17T08:42:57Z
mariadbautoscalers.autoscaling.kubedb.com         2022-10-17T08:53:53Z
mariadbdatabases.schema.kubedb.com                2022-10-17T08:54:08Z
mariadbopsrequests.ops.kubedb.com                 2022-10-17T08:54:20Z
mariadbs.kubedb.com                               2022-10-17T08:53:55Z
mariadbversions.catalog.kubedb.com                2022-10-17T08:43:09Z
memcacheds.kubedb.com                             2022-10-17T08:53:55Z
memcachedversions.catalog.kubedb.com              2022-10-17T08:43:10Z
mongodbautoscalers.autoscaling.kubedb.com         2022-10-17T08:53:53Z
mongodbdatabases.schema.kubedb.com                2022-10-17T08:53:56Z
mongodbopsrequests.ops.kubedb.com                 2022-10-17T08:54:01Z
mongodbs.kubedb.com                               2022-10-17T08:53:57Z
mongodbversions.catalog.kubedb.com                2022-10-17T08:43:11Z
mysqldatabases.schema.kubedb.com                  2022-10-17T08:53:54Z
mysqlopsrequests.ops.kubedb.com                   2022-10-17T08:54:17Z
mysqls.kubedb.com                                 2022-10-17T08:53:54Z
mysqlversions.catalog.kubedb.com                  2022-10-17T08:43:17Z
perconaxtradbopsrequests.ops.kubedb.com           2022-10-17T08:54:36Z
perconaxtradbs.kubedb.com                         2022-10-17T08:54:00Z
perconaxtradbversions.catalog.kubedb.com          2022-10-17T08:43:22Z
pgbouncers.kubedb.com                             2022-10-17T08:54:00Z
pgbouncerversions.catalog.kubedb.com              2022-10-17T08:43:23Z
postgresdatabases.schema.kubedb.com               2022-10-17T08:54:07Z
postgreses.kubedb.com                             2022-10-17T08:54:00Z
postgresopsrequests.ops.kubedb.com                2022-10-17T08:54:29Z
postgresversions.catalog.kubedb.com               2022-10-17T08:43:24Z
proxysqlopsrequests.ops.kubedb.com                2022-10-17T08:54:32Z
proxysqls.kubedb.com                              2022-10-17T08:54:01Z
proxysqlversions.catalog.kubedb.com               2022-10-17T08:43:25Z
redises.kubedb.com                                2022-10-17T08:54:01Z
redisopsrequests.ops.kubedb.com                   2022-10-17T08:54:24Z
redissentinels.kubedb.com                         2022-10-17T08:54:02Z
redisversions.catalog.kubedb.com                  2022-10-17T08:43:31Z
```

## Deploy PostgreSQL Clustered Database

Now, we are going to Deploy PostgreSQL with the help of KubeDB.
At first, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create ns demo
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
  version: "14.2"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "gp2"
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
$ kubectl create -f postgres-cluster.yaml
postgres.kubedb.com/postgres-cluster created
```

* In this yaml we can see in the `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 14.2`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy).

Once these are handled correctly and the PostgreSQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/postgres-cluster-0   2/2     Running   0          4m14s
pod/postgres-cluster-1   2/2     Running   0          3m51s
pod/postgres-cluster-2   2/2     Running   0          3m26s

NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/postgres-cluster           ClusterIP   10.100.94.172   <none>        5432/TCP,2379/TCP            4m17s
service/postgres-cluster-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   4m17s
service/postgres-cluster-standby   ClusterIP   10.100.37.12    <none>        5432/TCP                     4m17s

NAME                                READY   AGE
statefulset.apps/postgres-cluster   3/3     4m20s

NAME                                                  TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/postgres-cluster   kubedb.com/postgres   14.2      4m23s

NAME                                   VERSION   STATUS   AGE
postgres.kubedb.com/postgres-cluster   14.2      Ready    4m39s

```
Let’s check if the database is ready to use,

```bash
$ kubectl get postgres -n demo postgres-cluster
NAME               VERSION   STATUS   AGE
postgres-cluster   14.2      Ready    5m33s
```
> We have successfully deployed PostgreSQL in EKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create Secret and Service for the database `postgres-cluster` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                    TYPE                       DATA   AGE
postgres-cluster-auth   kubernetes.io/basic-auth   2      5m57s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=postgres-cluster
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
postgres-cluster           ClusterIP   10.100.94.172   <none>        5432/TCP,2379/TCP            6m20s
postgres-cluster-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   6m20s
postgres-cluster-standby   ClusterIP   10.100.37.12    <none>        5432/TCP                     6m20s

```
Now, we are going to use `postgres-cluster-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.username}' | base64 -d
postgres

$ kubectl get secrets -n demo postgres-cluster-auth -o jsonpath='{.data.password}' | base64 -d
mn~9wEF8;p*Ri(id

```

#### Insert Sample Data

In this section, we are going to login into our PostgreSQL database pod and insert some sample data. 

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -c postgres -- bash
bash-5.1$ psql -d "user=postgres password=mn~9wEF8;p*Ri(id"
psql (14.2)
Type "help" for help.

postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(4 rows)

postgres=# CREATE DATABASE music;
CREATE DATABASE
postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 music         | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(5 rows)

postgres=# \c music
You are now connected to database "music" as user "postgres".

music=# CREATE TABLE artist (name VARCHAR(50) NOT NULL, song VARCHAR(50) NOT NULL);
CREATE TABLE

music=# INSERT INTO artist (name, song) VALUES('Bon Jovi', 'Its My Life');
INSERT 0 1

music=# SELECT * FROM artist;
   name   |    song     
----------+-------------
 Bon Jovi | Its My Life
(1 row)

music=# \q

bash-5.1$ exit
exit
```

> We've successfully inserted some sample data to our database. And this was just an example of our PostgreSQL Clustered database deployment. More information about Run & Manage Production-Grade PostgreSQL Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

## Backup PostgreSQL Database Using Stash

Here, we are going to use Stash to backup the PostgreSQL database that we have just deployed.

### Install Stash

Kubedb Enterprise License works for Stash too.
So, we will use the Enterprise license that we have already obtained.

```bash
$ helm install stash appscode/stash             \
  --version v2022.09.29                  \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise
NAMESPACE     NAME                                      READY   STATUS    RESTARTS   AGE
kube-system   stash-stash-enterprise-76767bcb95-52qfh   2/2     Running   0          32s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-10-17T12:30:26Z
backupblueprints.stash.appscode.com       2022-10-17T12:30:26Z
backupconfigurations.stash.appscode.com   2022-10-17T12:30:24Z
backupsessions.stash.appscode.com         2022-10-17T12:30:24Z
functions.stash.appscode.com              2022-10-17T12:25:14Z
repositories.stash.appscode.com           2022-10-17T08:54:10Z
restorebatches.stash.appscode.com         2022-10-17T12:30:26Z
restoresessions.stash.appscode.com        2022-10-17T08:54:10Z
tasks.stash.appscode.com                  2022-10-17T12:25:16Z
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
      prefix: /postgres-backup 
    storageSecretName: s3-secret
```

This repository CRO specifies the `s3-secret` we created before and stores the name and path to the AWS storage bucket. It also specifies the location to the container where we want to backup our database.
> Here, My bucket name is `stash-qa`. Don't forget to change `spec.backend.s3.bucket` to your bucket name and For `S3`, use `s3.amazonaws.com` as endpoint.

Lets create this repository,

```bash
$ kubectl create -f s3-repo.yaml 
repository.stash.appscode.com/s3-repo created
```

### Create BackupConfiguration

Now, we need to create a `BackupConfiguration` file that specifies what to backup, where to backup and when to backup.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: postgres-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: postgres-cluster
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```
Create this `BackupConfiguration` by following command,

```bash
$ kubectl apply -f postgres-backup.yaml
backupconfiguration.stash.appscode.com/postgres-backup created
```

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the repository name that we have created before called `s3-repo`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup.
* `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
* `spec.retentionPolicy` specifies the policy to follow for cleaning old snapshots. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/latest/guides/postgres/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                         INVOKER-TYPE          INVOKER-NAME      PHASE       DURATION   AGE
postgres-backup-1666010403   BackupConfiguration   postgres-backup   Succeeded   23s        43s

$ kubectl get repository -n demo
NAME      INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
s3-repo   true        5.166 KiB   1                60s                      2m53s
```

Now if we check our Amazon S3 bucket, we can see that the backup has been successful.

![AWSSuccess](AWSStorageSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up PostgreSQL Database using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20AWS).

## Recover PostgreSQL Database Using Stash

Let's think of a scenario in which the database has been accidentally deleted or there was an error in the database causing it to crash.

#### Temporarily pause backup

At first, let’s stop taking any further backup of the database so that no backup runs after we delete the sample data. We are going to pause the `BackupConfiguration` object. Stash will stop taking any further backup when the `BackupConfiguration` is paused.

```bash
$ kubectl patch backupconfiguration -n demo postgres-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/postgres-backup patched
```

Now, we are going to delete database to simulate accidental database deletion.

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -c postgres -- bash
bash-5.1$ psql -d "user=postgres password=mn~9wEF8;p*Ri(id"
psql (14.2)
Type "help" for help.

postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 music         | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(5 rows)

postgres=# DROP DATABASE music;
DROP DATABASE
postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(4 rows)

postgres=# \q
bash-5.1$ exit
exit

```

### Create a RestoreSession

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: postgres-restore
  namespace: demo
spec:
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: postgres-cluster
  rules:
    - snapshots: [latest]
```

Now, let's create `RestoreSession` that will initiate restoring from the cloud.

```bash
$ kubectl apply -f postgres-restore.yaml 
restoresession.stash.appscode.com/postgres-restore created
```

This `RestoreSession` specifies where the data will be restored.
Once this is applied, a `RestoreSession` will be created. Once it has succeeded, the database has been successfully recovered as you can see below:

```bash
$ kubectl get restoresession -n demo
NAME               REPOSITORY   PHASE       DURATION   AGE
postgres-restore   s3-repo      Succeeded   4s         23s
```

Now, let's check whether the database has been correctly restored:

```bash
$ kubectl exec -it postgres-cluster-0 -n demo -c postgres -- bash
bash-5.1$ psql -d "user=postgres password=mn~9wEF8;p*Ri(id"
psql (14.2)
Type "help" for help.

postgres=# \l
                                   List of databases
     Name      |  Owner   | Encoding |  Collate   |   Ctype    |   Access privileges   
---------------+----------+----------+------------+------------+-----------------------
 kubedb_system | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 music         | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 postgres      | postgres | UTF8     | en_US.utf8 | en_US.utf8 | 
 template0     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
 template1     | postgres | UTF8     | en_US.utf8 | en_US.utf8 | =c/postgres          +
               |          |          |            |            | postgres=CTc/postgres
(5 rows)

postgres=# \c music
You are now connected to database "music" as user "postgres".

music=# SELECT * FROM artist;
   name   |    song     
----------+-------------
 Bon Jovi | Its My Life
(1 row)

music=# \q
bash-5.1$ exit
exit

```

> You can see the database has been restored. The recovery of PostgreSQL Database has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20AWS).

We have made an in depth video on PostgreSQL Backup with Wal G and point in time Recover with KubeDB managed PostgreSQL Database. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/gR5UdN6Y99c" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [PostgreSQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
