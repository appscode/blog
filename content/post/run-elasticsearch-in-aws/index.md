---
title: Run Elasticsearch in Amazon Elastic Kubernetes Service (Amazon EKS) Using KubeDB
date: "2022-08-29"
weight: 12
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- database
- dbaas
- elasticsearch
- elasticstack
- kubedb
- kubernetes
- s3
- xpack
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Elasticsearch, MySQL, MongoDB, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic, but also other open source distributions like, [OpenSearch](https://opensearch.org/), [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will deploy Elasticsearch database in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Elasticsearch Topology Cluster
3) Install Stash
4) Backup Elasticsearch Using Stash
5) Recover Elasticsearch Using Stash

## Install KubeDB

We will follow the steps to install KubeDB.

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID, we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
6c08dcb8-8440-4388-849f-1f2b590b731e
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
appscode/kubedb                   	v2022.08.08  	v2022.08.08	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.13.0      	v0.13.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2022.08.08  	v2022.08.08	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2022.08.08  	v2022.08.08	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.4.0       	v0.4.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2022.08.08  	v2022.08.08	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2022.08.08  	v2022.08.08	KubeDB State Metrics                              
appscode/kubedb-ops-manager       	v0.15.0      	v0.15.4    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2022.08.08  	v2022.08.08	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.28.0      	v0.28.4    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.4.0       	v0.4.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.9      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.4.0       	v0.4.4     	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2022.08.08 \
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
NAMESPACE   NAME                                           READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-67dbb56797-dtmm5      1/1     Running   0          2m19s
kubedb      kubedb-kubedb-dashboard-5dbddbfbc9-h9c4x       1/1     Running   0          2m19s
kubedb      kubedb-kubedb-ops-manager-66f6dc4774-rxbkm     1/1     Running   0          2m19s
kubedb      kubedb-kubedb-provisioner-6f67b5948f-6pqql     1/1     Running   0          2m19s
kubedb      kubedb-kubedb-schema-manager-f5c79bf8d-nx454   1/1     Running   0          2m19s
kubedb      kubedb-kubedb-webhook-server-574b775c7-t7gp7   1/1     Running   0          2m19s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2022-08-26T06:00:17Z
elasticsearchdashboards.dashboard.kubedb.com      2022-08-26T06:00:16Z
elasticsearches.kubedb.com                        2022-08-26T06:00:16Z
elasticsearchopsrequests.ops.kubedb.com           2022-08-26T06:00:19Z
elasticsearchversions.catalog.kubedb.com          2022-08-26T05:53:35Z
etcds.kubedb.com                                  2022-08-26T06:00:19Z
etcdversions.catalog.kubedb.com                   2022-08-26T05:53:36Z
mariadbautoscalers.autoscaling.kubedb.com         2022-08-26T06:00:17Z
mariadbdatabases.schema.kubedb.com                2022-08-26T06:00:22Z
mariadbopsrequests.ops.kubedb.com                 2022-08-26T06:00:37Z
mariadbs.kubedb.com                               2022-08-26T06:00:20Z
mariadbversions.catalog.kubedb.com                2022-08-26T05:53:37Z
memcacheds.kubedb.com                             2022-08-26T06:00:20Z
memcachedversions.catalog.kubedb.com              2022-08-26T05:53:37Z
mongodbautoscalers.autoscaling.kubedb.com         2022-08-26T06:00:17Z
mongodbdatabases.schema.kubedb.com                2022-08-26T06:00:17Z
mongodbopsrequests.ops.kubedb.com                 2022-08-26T06:00:23Z
mongodbs.kubedb.com                               2022-08-26T06:00:18Z
mongodbversions.catalog.kubedb.com                2022-08-26T05:53:38Z
mysqldatabases.schema.kubedb.com                  2022-08-26T06:00:16Z
mysqlopsrequests.ops.kubedb.com                   2022-08-26T06:00:34Z
mysqls.kubedb.com                                 2022-08-26T06:00:16Z
mysqlversions.catalog.kubedb.com                  2022-08-26T05:53:39Z
perconaxtradbopsrequests.ops.kubedb.com           2022-08-26T06:00:51Z
perconaxtradbs.kubedb.com                         2022-08-26T06:00:27Z
perconaxtradbversions.catalog.kubedb.com          2022-08-26T05:53:40Z
pgbouncers.kubedb.com                             2022-08-26T06:00:27Z
pgbouncerversions.catalog.kubedb.com              2022-08-26T05:53:41Z
postgresdatabases.schema.kubedb.com               2022-08-26T06:00:20Z
postgreses.kubedb.com                             2022-08-26T06:00:21Z
postgresopsrequests.ops.kubedb.com                2022-08-26T06:00:44Z
postgresversions.catalog.kubedb.com               2022-08-26T05:53:42Z
proxysqlopsrequests.ops.kubedb.com                2022-08-26T06:00:48Z
proxysqls.kubedb.com                              2022-08-26T06:00:29Z
proxysqlversions.catalog.kubedb.com               2022-08-26T05:53:43Z
redises.kubedb.com                                2022-08-26T06:00:29Z
redisopsrequests.ops.kubedb.com                   2022-08-26T06:00:40Z
redissentinels.kubedb.com                         2022-08-26T06:00:29Z
redisversions.catalog.kubedb.com                  2022-08-26T05:53:44Z
```

## Deploy Elasticsearch Topology Cluster

Now, We are going to use the KubeDB-provided Custom Resource object `Elasticsearch` for deployment. The object will be deployed in demo namespace. So, let's create the namespace first.

```bash
$ kubectl create ns demo
namespace/demo created
```

Here is the yaml of the Elasticsearch we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-topology-cluster
  namespace: demo
spec:
  enableSSL: true 
  version: xpack-8.2.0
  storageType: Durable
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "default"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 2
      storage:
        storageClassName: "default"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      storage:
        storageClassName: "default"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's save this yaml configuration into `es-topology-cluster.yaml` 
Then create the above Elasticsearch yaml

```bash
$ kubectl create -f es-topology-cluster.yaml
elasticsearch.kubedb.com/es-topology-cluster created
```

* In this yaml we can see in the `spec.version` field specifies the version of Elasticsearch. Here, we are using Elasticsearch version `xpack-8.2.0` which is used to provision `Elasticsearch-8.2.0` with xpack auth plugin. You can list the KubeDB supported versions of Elasticsearch CR by running `kubectl get elasticsearchversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
* `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
* `spec.topology` - specifies the node-specific properties for the Elasticsearch cluster.
  - `topology.master` - specifies the properties of master nodes.
    - `master.replicas` - specifies the number of master nodes.
    - `master.storage` - specifies the master node storage information that passed to the StatefulSet.
  - `topology.data` - specifies the properties of data nodes.
    - `data.replicas` - specifies the number of data nodes.
    - `data.storage` - specifies the data node storage information that passed to the StatefulSet.
  - `topology.ingest` - specifies the properties of ingest nodes.
    - `ingest.replicas` - specifies the number of ingest nodes.
    - `ingest.storage` - specifies the ingest node storage information that passed to the StatefulSet.

However, KubeDB also provides dedicated node support for other node roles like `data_hot`, `data_warm`, `data_cold`, `data_frozen`, `transform`, `coordinating`, `data_content` and `ml` for Topology clustering.

Once these are handled correctly and the Elasticsearch object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME                               READY   STATUS    RESTARTS   AGE
pod/es-topology-cluster-data-0     1/1     Running   0          118s
pod/es-topology-cluster-data-1     1/1     Running   0          79s
pod/es-topology-cluster-ingest-0   1/1     Running   0          118s
pod/es-topology-cluster-ingest-1   1/1     Running   0          81s
pod/es-topology-cluster-master-0   1/1     Running   0          118s
pod/es-topology-cluster-master-1   1/1     Running   0          74s

NAME                                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/es-topology-cluster          ClusterIP   10.100.102.136   <none>        9200/TCP   2m2s
service/es-topology-cluster-master   ClusterIP   None             <none>        9300/TCP   2m2s
service/es-topology-cluster-pods     ClusterIP   None             <none>        9200/TCP   2m2s

NAME                                          READY   AGE
statefulset.apps/es-topology-cluster-data     2/2     2m4s
statefulset.apps/es-topology-cluster-ingest   2/2     2m4s
statefulset.apps/es-topology-cluster-master   2/2     2m4s

NAME                                                     TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-topology-cluster   kubedb.com/elasticsearch   8.2.0     2m8s

NAME                                           VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-topology-cluster   xpack-8.2.0   Ready    2m30s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get elasticsearch -n demo es-topology-cluster
NAME                  VERSION       STATUS   AGE
es-topology-cluster   xpack-8.2.0   Ready    3m1s
```
> We have successfully deployed Elasticsearch in Amazon EKS. Now we can exec into the container to use the database.

### Insert Sample Data

In this section, we are going to create few indexes in Elasticsearch. On the deployment of Elasticsearch yaml, the operator creates a governing service that is named after the Elasticsearch object name itself. We are going to use this service to port-forward and connect with the database from our local machine. Then, we are going to insert some data into the Elasticsearch.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s see the Services created by KubeDB for our Elasticsearch,

```bash
$ kubectl get service -n demo -l=app.kubernetes.io/instance=es-topology-cluster
NAME                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
es-topology-cluster          ClusterIP   10.100.102.136   <none>        9200/TCP   3m36s
es-topology-cluster-master   ClusterIP   None             <none>        9300/TCP   3m36s
es-topology-cluster-pods     ClusterIP   None             <none>        9200/TCP   3m36s
```
Here, we are going to use the `es-topology-cluster` Service to connect with the database. Now, let’s port-forward the `es-topology-cluster` Service.

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-topology-cluster 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

#### Export the Credentials

KubeDB will create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `es-topology-cluster`.

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=es-topology-cluster
NAME                                              TYPE                       DATA   AGE
es-topology-cluster-apm-system-cred               kubernetes.io/basic-auth   2      10m
es-topology-cluster-beats-system-cred             kubernetes.io/basic-auth   2      10m
es-topology-cluster-ca-cert                       kubernetes.io/tls          2      10m
es-topology-cluster-client-cert                   kubernetes.io/tls          3      10m
es-topology-cluster-config                        Opaque                     1      10m
es-topology-cluster-elastic-cred                  kubernetes.io/basic-auth   2      10m
es-topology-cluster-http-cert                     kubernetes.io/tls          3      10m
es-topology-cluster-kibana-system-cred            kubernetes.io/basic-auth   2      10m
es-topology-cluster-logstash-system-cred          kubernetes.io/basic-auth   2      10m
es-topology-cluster-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      10m
es-topology-cluster-transport-cert                kubernetes.io/tls          3      10m
```
Now, we can connect to the database with any of these secret that have the prefix `cred`. Here, we are using `es-topology-cluster-elastic-cred` which contains the admin level credentials to connect with the database.


### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo es-topology-cluster-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-topology-cluster-elastic-cred -o jsonpath='{.data.password}' | base64 -d
Gdrowr9je1g9ZvhF
```

Then login and insert some data into Elasticsearch:

```bash
$ curl -XPOST -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/music/_doc?pretty" -H 'Content-Type: application/json' -d'
         {
             "Artist": "John Denver",
             "Song": "Take Me Home Country Roads"
         }
         '
{
  "_index" : "music",
  "_id" : "E8L22IIBNQpxAeKAr75D",
  "_version" : 1,
  "result" : "created",
  "_shards" : {
    "total" : 2,
    "successful" : 2,
    "failed" : 0
  },
  "_seq_no" : 0,
  "_primary_term" : 1
}

```

Now, let’s verify that the index has been created successfully.

```bash
$ curl -XGET -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index         uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   music         lCR6u-VsSFSKYWwdBHTpaQ   1   1          1            0      9.3kb          4.6kb
green  open   kubedb-system P8V6uX3OQvCRmGX7hjbmSQ   1   1          1            3    621.1kb        276.3kb
```
Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/music/_search?pretty"
{
  "took" : 6,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "music",
        "_id" : "E8L22IIBNQpxAeKAr75D",
        "_score" : 1.0,
        "_source" : {
          "Artist" : "John Denver",
          "Song" : "Take Me Home Country Roads"
        }
      }
    ]
  }
}

```
> We've successfully inserted some sample data to our Elasticsearch. And this was just an example of our Elasticsearch topology cluster deployment. More information about Run & Manage Production-Grade Elasticsearch Database on Kubernetes can be found [HERE](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

## Backup Elasticsearch Database Using Stash

Here, we are going to use Stash to backup the Elasticsearch database that we have just deployed.

### Install Stash

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
$ watch kubectl get pods --all-namespaces -l app.kubernetes.io/name=stash-enterprise
NAMESPACE     NAME                                    READY   STATUS    RESTARTS   AGE
kube-system   stash-stash-enterprise-bf8d6cc7-n28jk   2/2     Running   0          5m21s
```

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=stash
NAME                                      CREATED AT
backupbatches.stash.appscode.com          2022-08-26T07:11:47Z
backupblueprints.stash.appscode.com       2022-08-26T07:11:48Z
backupconfigurations.stash.appscode.com   2022-08-26T07:11:46Z
backupsessions.stash.appscode.com         2022-08-26T07:11:46Z
functions.stash.appscode.com              2022-08-26T07:07:48Z
repositories.stash.appscode.com           2022-08-26T06:00:26Z
restorebatches.stash.appscode.com         2022-08-26T07:11:48Z
restoresessions.stash.appscode.com        2022-08-26T06:00:26Z
tasks.stash.appscode.com                  2022-08-26T07:07:50Z
```


### Prepare Backend

Stash supports various backends for storing data snapshots. It can be a cloud storage like GCS bucket, AWS S3, Azure Blob Storage etc. or a Kubernetes persistent volume like HostPath, PersistentVolumeClaim, NFS etc.

For this tutorial we are going to use Amazon S3 storage. You can find other setups [here](https://stash.run/docs/v2022.05.18/guides/backends/overview/).

 ![My Empty S3 Storage](AWSStorageEmpty.png)

At first we need to create a secret so that we can access the Amazon S3 bucket. We can do that by the following code:

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
      prefix: /es-topology-cluster
    storageSecretName: s3-secret
```

This repository CRO specifies the `s3-secret` we created before and stores the name and path to the S3 storage bucket. It also specifies the location to the container where we want to backup our database.
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
  name: es-topology-cluster-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: es-topology-cluster
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```
Create this `BackupConfiguration` by following command,

```bash
$ kubectl create -f es-topology-cluster-backup.yaml 
backupconfiguration.stash.appscode.com/es-topology-cluster-backup created
```

* `BackupConfiguration` creates a cronjob that backs up the specified database (`spec.target`) every 5 minutes.
* `spec.repository` contains the repository name that we have created before called `s3-repo`.
* `spec.target.ref` contains the reference to the appbinding that we want to backup.
* `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
* `spec.retentionPolicy` specifies the policy to follow for cleaning old snapshots. 
* To learn more about `AppBinding`, click here [AppBinding](https://kubedb.com/docs/v2022.05.24/guides/elasticsearch/concepts/appbinding/). 
So, after 5 minutes we can see the following status:

```bash
$ kubectl get backupsession -n demo
NAME                                    INVOKER-TYPE          INVOKER-NAME                 PHASE       DURATION   AGE
es-topology-cluster-backup-1661499303   BackupConfiguration   es-topology-cluster-backup   Succeeded   32s        3m48s

$ kubectl get repository -n demo
NAME      INTEGRITY   SIZE         SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
s3-repo   true        14.223 KiB   1                4m10s                    9m48s
```

Now if we check our Amazon S3 bucket, we can see that the backup has been successful.

![AWSSuccess](AWSStorageSuccess.png)

> **If you have reached here, CONGRATULATIONS!! :confetti_ball: :confetti_ball: :confetti_ball: You have successfully backed up Elasticsearch Database using Stash.** If you had any problem during the backup process, you can reach out to us via [EMAIL](mailto:support@appscode.com?subject=Stash%20Backup%20Failed%20in%20AWS).

## Recover Elasticsearch Database Using Stash

Let's think of a scenario in which the database has been accidentally deleted or there was an error in the database causing it to crash.

#### Temporarily pause backup

At first, let’s stop taking any further backup of the database so that no backup runs after we delete the sample data. We are going to pause the `BackupConfiguration` object. Stash will stop taking any further backup when the `BackupConfiguration` is paused.

```bash
$ kubectl patch backupconfiguration -n demo es-topology-cluster-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/es-topology-cluster-backup patched

```
Verify that the `BackupConfiguration` has been paused,

```bash
$ kubectl get backupconfiguration -n demo es-topology-cluster-backup
NAME                         TASK   SCHEDULE      PAUSED   PHASE   AGE
es-topology-cluster-backup          */5 * * * *   true     Ready   8m27s
```
Notice the `PAUSED` column. Value `true` for this field means that the `BackupConfiguration` has been paused.
Stash will also suspend the respective CronJob.

```bash
$ kubectl get cronjob -n demo
NAME                                       SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
stash-trigger-es-topology-cluster-backup   */5 * * * *   True      0        6m23s           9m35s
```

At first, let's simulate an accidental database deletion. Here, we are going to delete the `music` index that we have created earlier.

```bash
$ curl -XDELETE -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/music?pretty"
{
  "acknowledged" : true
}
```
Now, let’s verify that the indexes have been deleted from the database,

```bash
$ curl -XGET -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index         uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   kubedb-system P8V6uX3OQvCRmGX7hjbmSQ   1   1          1          537      1.3mb        640.2kb
```

### Create a RestoreSession

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: es-topology-cluster-restore
  namespace: demo
spec:
  repository:
    name: s3-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: es-topology-cluster 
  rules:
    - snapshots: [latest]
```

Now, let's create `RestoreSession` that will initiate restoring from the cloud.

```bash
$ kubectl create -f es-topology-cluster-restore.yaml
restoresession.stash.appscode.com/es-toplogy-cluster-restore created
```

This `RestoreSession` specifies where the data will be restored.
Once this is applied, a `RestoreSession` will be created. Once it has succeeded, the database has been successfully recovered as you can see below:

```bash
$ kubectl get restoresession -n demo
NAME                          REPOSITORY   PHASE       DURATION   AGE
es-topology-cluster-restore   s3-repo      Succeeded   8s         28s
```

Now, let's check whether the database has been correctly restored:

```bash
$ curl -XGET -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index         uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   kubedb-system P8V6uX3OQvCRmGX7hjbmSQ   1   1          1          537      1.5mb          731kb
green  open   music         -BrQ0V6TQfyPGKAMLBoq-Q   1   1          1            0      9.3kb          4.6kb
```
Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'elastic:Gdrowr9je1g9ZvhF' "https://localhost:9200/music/_search?pretty"
{
  "took" : 4,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "music",
        "_id" : "FMLh2YIBNQpxAeKAY74_",
        "_score" : 1.0,
        "_source" : {
          "Artist" : "John Denver",
          "Song" : "Take Me Home Country Roads"
        }
      }
    ]
  }
}

```


> You can see the database has been restored. The recovery of Elasticsearch has been successful. If you faced any difficulties in the recovery process, you can reach out to us through [EMAIL](mailto:support@appscode.com?subject=Stash%20Recovery%20Failed%20in%20AWS).

We have made an in depth tutorial on Elasticsearch Hot-Warm-Cold Architecture Management with Kibana in Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/R-eYc2cUXQY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Elasticsearch in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
