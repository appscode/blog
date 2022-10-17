---
title: Run MySQL in Azure Kubernetes Service (AKS) Using KubeDB
date: "2022-05-16"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- azure-storage
- azure-container
- cloud-native
- database
- kubedb
- kubernetes
- kubernetes-mysql
- managed-dbaas
- microsoft-azure
- mysql
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/).
In this tutorial we will deploy MySQL database in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL Clustered Database
3) Install Stash
4) Backup MySQL Database Using Stash
5) Recover MySQL Database Using Stash

## Install KubeDB

We will follow the steps to install KubeDB.

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
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
appscode/kubedb                   	v2022.03.28  	v2022.03.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.11.0      	v0.11.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.03.28  	v2022.03.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.03.28  	v2022.03.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.2.0       	v0.2.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.03.28  	v2022.03.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.03.28  	v2022.03.28	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.13.0      	v0.13.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.03.28  	v2022.03.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.26.0      	v0.26.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.2.0       	v0.2.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.2.0       	v0.2.0     	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.03.28 \
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
kubedb      kubedb-kubedb-autoscaler-79d6bcfbbb-xhdts       1/1     Running   0          2m
kubedb      kubedb-kubedb-dashboard-7f4d7bb9bc-5cxbj        1/1     Running   0          2m
kubedb      kubedb-kubedb-ops-manager-74f8d99865-x9d65      1/1     Running   0          2m
kubedb      kubedb-kubedb-provisioner-5fcbf5b584-26hms      1/1     Running   0          2m
kubedb      kubedb-kubedb-schema-manager-7859949584-nssfc   1/1     Running   0          2m
kubedb      kubedb-kubedb-webhook-server-6b97496659-zx4f6   1/1     Running   0          2m
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-05-12T09:18:17Z
elasticsearchdashboards.dashboard.kubedb.com      2022-05-12T09:18:43Z
elasticsearches.kubedb.com                        2022-05-12T09:17:59Z
elasticsearchopsrequests.ops.kubedb.com           2022-05-12T09:19:05Z
elasticsearchversions.catalog.kubedb.com          2022-05-12T09:17:27Z
etcds.kubedb.com                                  2022-05-12T09:17:59Z
etcdversions.catalog.kubedb.com                   2022-05-12T09:17:27Z
mariadbautoscalers.autoscaling.kubedb.com         2022-05-12T09:18:20Z
mariadbdatabases.schema.kubedb.com                2022-05-12T09:18:30Z
mariadbopsrequests.ops.kubedb.com                 2022-05-12T09:19:19Z
mariadbs.kubedb.com                               2022-05-12T09:17:59Z
mariadbversions.catalog.kubedb.com                2022-05-12T09:17:27Z
memcacheds.kubedb.com                             2022-05-12T09:17:59Z
memcachedversions.catalog.kubedb.com              2022-05-12T09:17:27Z
mongodbautoscalers.autoscaling.kubedb.com         2022-05-12T09:18:14Z
mongodbdatabases.schema.kubedb.com                2022-05-12T09:18:29Z
mongodbopsrequests.ops.kubedb.com                 2022-05-12T09:19:08Z
mongodbs.kubedb.com                               2022-05-12T09:18:00Z
mongodbversions.catalog.kubedb.com                2022-05-12T09:17:27Z
mysqldatabases.schema.kubedb.com                  2022-05-12T09:18:29Z
mysqlopsrequests.ops.kubedb.com                   2022-05-12T09:19:15Z
mysqls.kubedb.com                                 2022-05-12T09:18:00Z
mysqlversions.catalog.kubedb.com                  2022-05-12T09:17:27Z
perconaxtradbs.kubedb.com                         2022-05-12T09:18:00Z
perconaxtradbversions.catalog.kubedb.com          2022-05-12T09:17:27Z
pgbouncers.kubedb.com                             2022-05-12T09:18:00Z
pgbouncerversions.catalog.kubedb.com              2022-05-12T09:17:27Z
postgresdatabases.schema.kubedb.com               2022-05-12T09:18:30Z
postgreses.kubedb.com                             2022-05-12T09:18:00Z
postgresopsrequests.ops.kubedb.com                2022-05-12T09:19:25Z
postgresversions.catalog.kubedb.com               2022-05-12T09:17:27Z
proxysqls.kubedb.com                              2022-05-12T09:18:01Z
proxysqlversions.catalog.kubedb.com               2022-05-12T09:17:27Z
redises.kubedb.com                                2022-05-12T09:18:01Z
redisopsrequests.ops.kubedb.com                   2022-05-12T09:19:22Z
redissentinels.kubedb.com                         2022-05-12T09:18:01Z
redisversions.catalog.kubedb.com                  2022-05-12T09:17:27Z
```

## Deploy MySQL Clustered Database

Now, we are going to Deploy MySQL with the help of KubeDB.
At first, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create ns demo
namespace/demo created
```

Here is the yaml of the MySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.27"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "default"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `sample-mysql.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl create -f sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```

* In this yaml we can see in the `spec.version` field specifies the version of MySQL. Here, we are using MySQL `version 8.0.27`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/v2022.03.28/guides/mysql/concepts/database/#specterminationpolicy).

Once these are handled correctly and the MySQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/sample-mysql-0   1/1     Running   0          6m21s
pod/sample-mysql-1   1/1     Running   0          5m53s
pod/sample-mysql-2   1/1     Running   0          5m40s

NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/sample-mysql        ClusterIP   10.0.249.213   <none>        3306/TCP   6m22s
service/sample-mysql-pods   ClusterIP   None           <none>        3306/TCP   6m22s

NAME                            READY   AGE
statefulset.apps/sample-mysql   3/3     6m23s

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/sample-mysql   kubedb.com/mysql   8.0.27    6m24s

NAME                            VERSION   STATUS   AGE
mysql.kubedb.com/sample-mysql   8.0.27    Ready    6m28s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo sample-mysql
NAME           VERSION   STATUS   AGE
sample-mysql   8.0.27    Ready    7m15s
```
> We have successfully deployed MySQL in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access.
KubeDB will create Secret and Service for the database `sample-mysql` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-mysql
NAME                TYPE                       DATA   AGE
sample-mysql-auth   kubernetes.io/basic-auth   2      8m

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mysql
NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
sample-mysql        ClusterIP   10.0.249.213   <none>        3306/TCP   8m34s
sample-mysql-pods   ClusterIP   None           <none>        3306/TCP   8m34s
```
Now, we are going to use `sample-mysql-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo sample-mysql-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mysql-auth -o jsonpath='{.data.password}' | base64 -d
yAl1~7YcDaTzO_mm

$ kubectl exec -it sample-mysql-0 -n demo -c mysql -- bash
```

#### Insert Sample Data

In this section, we are going to login into our MySQL database pod and insert some sample data. 

```bash
root@sample-mysql-0:/# mysql --user=root --password='yAl1~7YcDaTzO_mm'
Welcome to the MySQL monitor.  Commands end with ; or \g.

mysql> CREATE DATABASE Music;
Query OK, 1 row affected (0.02 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE Music.Artist (id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(50), Song VARCHAR(25));
Query OK, 0 rows affected (0.05 sec)

mysql> INSERT INTO Music.Artist (Name, Song) VALUES ("Bon Jovi", "It's My Life");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM Music.Artist;
+----------+--------------+
| Name     | Song         |
+----------+--------------+
| Bon Jovi | It's My Life |
+----------+--------------+
1 row in set (0.00 sec)

mysql> exit
Bye

```

> We've successfully inserted some sample data to our database. And this was just an example of our MySQL Clustered database deployment. More information about Run & Manage Production-Grade MySQL Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

## Backup MySQL Database Using Stash

Here, we are going to use Stash to backup the MySQL database that we have just deployed.

### Install Stash

Kubedb Enterprise License works for Stash too.
So, we will use the Enterprise license that we have already obtained.

```bash
$ helm install stash appscode/stash             \
  --version v2022.02.22                         \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise
NAMESPACE     NAME                                      READY   STATUS    RESTARTS   AGE
kube-system   stash-stash-enterprise-66dd58ccf7-b696r   2/2     Running   0          2m54s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-05-12T11:09:26Z
backupblueprints.stash.appscode.com       2022-05-12T11:09:27Z
backupconfigurations.stash.appscode.com   2022-05-12T11:09:25Z
backupsessions.stash.appscode.com         2022-05-12T11:09:25Z
functions.stash.appscode.com              2022-05-12T11:07:16Z
repositories.stash.appscode.com           2022-05-12T10:41:10Z
restorebatches.stash.appscode.com         2022-05-12T11:09:27Z
restoresessions.stash.appscode.com        2022-05-12T10:41:10Z
tasks.stash.appscode.com                  2022-05-12T11:07:17Z

```


### Prepare Backend

Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes persistent volume like HostPath, PersistentVolumeClaim, NFS etc.

For this tutorial we are going to use azure storage. You can find other setups [here](https://stash.run/docs/v2022.05.12/guides/backends/overview/).

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
      prefix: /sample-mysql
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
  name: sample-mysql-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: azure-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```
Create this `BackupConfiguration` by following command,

```bash
$ kubectl apply -f sample-mysql-backup.yaml
backupconfiguration.stash.appscode.com/sample-mysql-backup created
```

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the secret we created before called `azure-secret`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup.
* `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
* `spec.retentionPolicy` specifies the policy to follow for cleaning old snapshots. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/v2022.03.28/guides/mysql/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
sample-mysql-backup-1652358900   BackupConfiguration   sample-mysql-backup   Succeeded   8s         7m

$ kubectl get repository -n demo
NAME         INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
azure-repo   true        8.878 MiB   5                31s                      7m
```

Now if we check our azure storage container, we can see that the backup has been successful.

![AzureSuccess](AzureStorageSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up MySQL Database using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20GKE).

## Recover MySQL Database Using Stash

Let's think of a scenario in which the database has been accidentally deleted or there was an error in the database causing it to crash.

#### Temporarily pause backup

At first, let’s stop taking any further backup of the database so that no backup runs after we delete the sample data. We are going to pause the `BackupConfiguration` object. Stash will stop taking any further backup when the `BackupConfiguration` is paused.

```bash
$ kubectl patch backupconfiguration -n demo sample-mysql-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mysql-backup patched
```

Now, we are going to delete database to simulate accidental database deletion.

```bash
$ kubectl exec -it sample-mysql-0 -n demo -c mysql -- bash
root@sample-mysql-0:/# mysql --user=root --password='yAl1~7YcDaTzO_mm'
Welcome to the MySQL monitor.  Commands end with ; or \g.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> DROP DATABASE Music;
Query OK, 1 row affected (0.04 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
4 rows in set (0.00 sec)

mysql> exit
Bye

```

### Create a RestoreSession

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mysql-restore
  namespace: demo
spec:
  repository:
    name: azure-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql
  rules:
    - snapshots: [latest]
```

Now, let's create `RestoreSession` that will initiate restoring from the cloud.

```bash
$ kubectl create -f sample-mysql-restore.yaml
restoresession.stash.appscode.com/sample-mysql-restore created
```

This `RestoreSession` specifies where the data will be restored.
Once this is applied, a `RestoreSession` will be created. Once it has succeeded, the database has been successfully recovered as you can see below:

```bash
$ kubectl get restoresession -n demo
NAME                   REPOSITORY   PHASE       DURATION   AGE
sample-mysql-restore   azure-repo   Succeeded   7s         38s
```

Now, let's check whether the database has been correctly restored:

```bash
$ kubectl exec -it sample-mysql-0 -n demo -c mysql -- bash
root@sample-mysql-0:/# mysql --user=root --password='yAl1~7YcDaTzO_mm'
Welcome to the MySQL monitor.  Commands end with ; or \g.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| Music              |
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> SELECT * FROM Music.Artist;
+----+----------+--------------+
| id | Name     | Song         |
+----+----------+--------------+
|  1 | Bon Jovi | It's My Life |
+----+----------+--------------+
1 row in set (0.00 sec)

mysql> exit
Bye


```

> You can see the database has been restored. The recovery of MySQL Database has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20GKE).

We have made an in depth video on Managing Semi-synchronous MySQL Cluster Using KubeDB in Kubernetes. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/y9VgofqwGyE" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MySQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
