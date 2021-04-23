---
title: How to Manage Database in Openshift Using KubeDB
date: 2021-04-23
weight: 20
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

1) Install KubeDB
2) Deploy Database
3) Install Stash
4) Backup Using Stash
5) Recover Using Stash

## Step 1: Installing KubeDB 
There are 3 steps in installing KubeDB.

### Step 1.1: Get Cluster ID
```bash
$ oc get ns kube-system -o=jsonpath='{.metadata.uid}'
08b1259c-5d51-4948-a2de-e2af8e6835a4 
```
###  Step 1.2: Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB Enterprise Edition.

![The KubeVault Overview](licenseserver.png)

### Step 1.3 Install KubeDB
We need helm to install KubeDB. It can be installed [here](https://helm.sh/docs/intro/install/) if it is not already installed.
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
```

# Step 2: Deploying Database

> Now we can Install a number of common databases with the help of KubeDB.

The databases that KubeDB support are MongoDB, Elasticsearch, MySQL, MariaDB, PostgreSQL and Redis. You can find the guides to all the supported databases [here](https://kubedb.com/).
## Deploying MySQL Database
Let's first create a Namespace in which we will deploy the database.
```bash
$ oc create ns demo
```
Now lets apply the following yaml file:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "8.0.23-v1"
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
This yaml uses MySQL CRD. In this yaml we can see in the *spec.version* field the version of MySQL. You can change and get updated version by running `oc get mysqlversions` command. Another field to notice is the *spec.storagetype* field. This can be Durable or Ephemeral depending on the requirements of the database to be persistent or not. Lastly, the *spec.terminationPolicy* field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/v2021.04.16/guides/mysql/concepts/database/#specterminationpolicy)
> NOTE: This might fail if correct permissions and storage class is not set. Let's make some checks so that the above yaml does not fail.
### Check 1: StorageClass check
Let's First check if storageclass is available:
```bash
$ oc get storageclass
NAME         PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION
local-path   rancher.io/local-path   Delete          WaitForFirstConsumer   false    
```
If you dont see the above output then you should run:
```bash
$ oc apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```
This will create the storage-class named local-path.

### Check 2: Correct Permissions

If you apply the above yaml and it is stuck in provisioning state then the pvc does not have required permissions. In such a case you should run:
```bash
$ oc adm policy add-scc-to-user privileged system:serviceaccount:local-path-storage:local-path-provisioner-service-account
```
This command will give the required permissions. </br>
### Deploy MySQL CRD
Once these are handled correctly and the MySQL CRD is deployed you will see that the following are created:
```bash
$ oc get all -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pod/mysql-quickstart-0   1/1     Running   0          31m

NAME                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/mysql-quickstart        ClusterIP   10.217.5.195   <none>        3306/TCP   31m
service/mysql-quickstart-pods   ClusterIP   None           <none>        3306/TCP   31m

NAME                                READY   AGE
statefulset.apps/mysql-quickstart   1/1     31m

NAME                                                  TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-quickstart   kubedb.com/mysql   8.0.23    31m

NAME                                VERSION     STATUS   AGE
mysql.kubedb.com/mysql-quickstart   8.0.23-v1   Ready    31m
```

> We have successfully deployed MySQL database in OpenShift. Now we can exec into the container to use the database.

## Accessing Database Through CLI

To access the database through CLI we have to exec into the container:
 ```bash
$ oc exec -it -n demo mysql-quickstart-0 -- bash
 ```
 Then to login into mysql:
 ```bash
mysql -uroot -p${MYSQL_ROOT_PASSWORD}
 ```
Now we have entered into the MySQL CLI and we can create and delete as we want.
let's create a database and create a table called MyGuests:
```bash
mysql> create database testdb;
mysql> show databases;

+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
| testdb             |
+--------------------+
5 rows in set (0.01 sec)
mysql> use testdb

Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

mysql> CREATE TABLE MyGuests (
    -> id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    -> firstname VARCHAR(30) NOT NULL,
    -> lastname VARCHAR(30) NOT NULL,
    -> email VARCHAR(50),
    -> reg_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    -> );
Query OK, 0 rows affected, 1 warning (0.02 sec)

mysql> show tables;
+------------------+
| Tables_in_testdb |
+------------------+
| MyGuests         |
+------------------+
1 row in set (0.02 sec)

```
> This was just one example of database deployment. The other databases that KubeDB suport are MongoDB, Elasticsearch, MariaDB, PostgreSQL and Redis. The tutorials on how to deploy these into the cluster can be found [HERE](https://kubedb.com/)

# Backup and Recover In OpenShift

## Backup

### Step 1: Install Stash
Here we will use the KubeDB license we obtained earlier.
```bash
$ helm install stash appscode/stash             \
  --version v2021.04.12                  \
  --namespace kube-system                       \
  --set features.enterprise=true                \
  --set-file global.license=/path/to/the/license.txt
```
Let's verify the installation:
```bash
$ oc get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise --watch
```
### Step 2: Prepare Backend
Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes persistent volume like HostPath, PersistentVolumeClaim, NFS etc.

For this tutorial we are going to use gcs-bucket. You can find other setups [here](https://stash.run/docs/v2021.04.12/guides/latest/backends/overview/).

 ![My GCS bucket](gcsEmptyBucket.png)
 **Create Secret:**
 ```bash
$ echo -n 'YOURPASSWORD' > RESTIC_PASSWORD
$ echo -n 'YOURPROJECTNAME' > GOOGLE_PROJECT_ID
$ cat /PATH/TO/JSONKEY.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ oc create secret generic -n demo gcs-secret \
        --from-file=./RESTIC_PASSWORD \
        --from-file=./GOOGLE_PROJECT_ID \
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
      bucket: YOURBUCKETNAME
      prefix: /demo/mysql/sample-mysql
    storageSecretName: gcs-secret
```
This repository specifies the gcs-secret we created before and connects to the gcs-bucket. It also specifies the location in the bucket where we want to backup our database.
> Don't forget to change `spec.backend.gcs.bucket` to your bucket name.

### Step 4: Create BackupConfiguration
```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mysql-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: mysql-quickstart
  runtimeSettings:
    container:
      securityContext:
        runAsUser: 1000610000
        runAsGroup: 1000610000
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```
This BackupConfiguration creates a cronjob that backs up the specified database every 5 minutes.</br>
So, after 5 minutes we can see the following status:

![](backup.png)

Now if we check our GCS bucket we can see that the backup has been successful.

![](gcsSuccess.png)

> **If you reached here CONGRATULATIONS!! :confetti_ball:  :partying_face: 		:confetti_ball: The backup has been successful**. If you didn't its okay. You can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20OpenShift).

## Recover
Let's think of a scenario in which the database has been accidentally deleted or there was an error in the database causing it to crash.</br>
In such a case, we have to pause the backupconfiguration so that the failed/damaged database does not get backed up into the cloud:
```bash
oc patch backupconfiguration -n demo sample-mysql-backup --type="merge" --patch='{"spec": {"paused": true}}'
```
In order to show that the recovery has been successful let's simulate accidental database deletion.

![Delete Database](deleteDatabase.png)

**Now let's start recovering the database.**
### Step 1: Create a RestoreSession
```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mysql-restore
  namespace: demo
spec:
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: mysql-quickstart
  runtimeSettings:
    container:
      securityContext:
        runAsUser: 1000610000
        runAsGroup: 1000610000
  rules:
    - snapshots: [latest]
```
Once this is applied, a RestoreSession will be created. Once it has succeeded, the database has been successfully recovered as you can see in the images below:

![Recovery Succeeded](recoverSucceed.png)

![Recovered Database](recoveredDB.png)

> **CONGRATULATIONS!! :confetti_ball:  :partying_face: 		:confetti_ball: The recovery has been successful**. If you faced any difficulties in the recovery process you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20OpenShift).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).