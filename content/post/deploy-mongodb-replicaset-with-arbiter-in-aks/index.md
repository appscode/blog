---
title: Deploy MongoDB ReplicaSet with Arbiter in Azure Kubernetes Service (AKS)
date: "2024-02-14"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- kubedb
- kubernetes
- microsoft-azure
- mongodb
- mongodb-arbiter
- mongodb-database
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy MongoDB ReplicaSet with Arbiter in Azure Kubernetes Service (AKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB ReplicaSet with Arbiter
3) Connect MongoDB with Arbiter
4) Read/Write with Arbiter

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2024.2.14   	v2024.2.14 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.27.0      	v0.27.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.2.14   	v2024.2.14 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.5       	v0.0.5     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.2.14   	v2024.2.14 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.18.0      	v0.18.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.2.14   	v2024.2.14 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.2.14   	v2024.2.14 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.2.14   	v2024.2.14 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.29.0      	v0.29.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.2.14   	v2024.2.14 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.2.14   	v0.4.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.42.0      	v0.42.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.18.0      	v0.18.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.2.13   	0.6.4      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.18.0      	v0.18.0    	KubeDB Webhook Server by AppsCode  


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.2.14 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                           READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-c8fc67b98-5jwdm       1/1     Running   0          7m47s
kubedb      kubedb-kubedb-ops-manager-54655f8c55-pq5fg     1/1     Running   0          7m47s
kubedb      kubedb-kubedb-provisioner-59898fdfcd-9jcgb     1/1     Running   0          7m47s
kubedb      kubedb-kubedb-webhook-server-b659c84ff-9m2w6   1/1     Running   0          7m47s
kubedb      kubedb-sidekick-5dc87959b7-hjplr               1/1     Running   0          7m47s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-02-14T09:06:01Z
connectors.kafka.kubedb.com                        2024-02-14T09:06:01Z
druidversions.catalog.kubedb.com                   2024-02-14T09:05:19Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-02-14T09:05:57Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-02-14T09:05:58Z
elasticsearches.kubedb.com                         2024-02-14T09:05:57Z
elasticsearchopsrequests.ops.kubedb.com            2024-02-14T09:05:57Z
elasticsearchversions.catalog.kubedb.com           2024-02-14T09:05:19Z
etcdversions.catalog.kubedb.com                    2024-02-14T09:05:19Z
ferretdbversions.catalog.kubedb.com                2024-02-14T09:05:19Z
kafkaconnectorversions.catalog.kubedb.com          2024-02-14T09:05:19Z
kafkaopsrequests.ops.kubedb.com                    2024-02-14T09:06:01Z
kafkas.kubedb.com                                  2024-02-14T09:06:01Z
kafkaversions.catalog.kubedb.com                   2024-02-14T09:05:19Z
mariadbautoscalers.autoscaling.kubedb.com          2024-02-14T09:06:04Z
mariadbdatabases.schema.kubedb.com                 2024-02-14T09:06:04Z
mariadbopsrequests.ops.kubedb.com                  2024-02-14T09:06:04Z
mariadbs.kubedb.com                                2024-02-14T09:06:04Z
mariadbversions.catalog.kubedb.com                 2024-02-14T09:05:19Z
memcachedversions.catalog.kubedb.com               2024-02-14T09:05:19Z
mongodbarchivers.archiver.kubedb.com               2024-02-14T09:06:08Z
mongodbautoscalers.autoscaling.kubedb.com          2024-02-14T09:06:08Z
mongodbdatabases.schema.kubedb.com                 2024-02-14T09:06:08Z
mongodbopsrequests.ops.kubedb.com                  2024-02-14T09:06:08Z
mongodbs.kubedb.com                                2024-02-14T09:06:07Z
mongodbversions.catalog.kubedb.com                 2024-02-14T09:05:19Z
mysqlarchivers.archiver.kubedb.com                 2024-02-14T09:06:11Z
mysqlautoscalers.autoscaling.kubedb.com            2024-02-14T09:06:11Z
mysqldatabases.schema.kubedb.com                   2024-02-14T09:06:11Z
mysqlopsrequests.ops.kubedb.com                    2024-02-14T09:06:11Z
mysqls.kubedb.com                                  2024-02-14T09:06:11Z
mysqlversions.catalog.kubedb.com                   2024-02-14T09:05:19Z
perconaxtradbversions.catalog.kubedb.com           2024-02-14T09:05:19Z
pgbouncerversions.catalog.kubedb.com               2024-02-14T09:05:19Z
pgpoolversions.catalog.kubedb.com                  2024-02-14T09:05:19Z
postgresarchivers.archiver.kubedb.com              2024-02-14T09:06:15Z
postgresautoscalers.autoscaling.kubedb.com         2024-02-14T09:06:15Z
postgresdatabases.schema.kubedb.com                2024-02-14T09:06:15Z
postgreses.kubedb.com                              2024-02-14T09:06:15Z
postgresopsrequests.ops.kubedb.com                 2024-02-14T09:06:15Z
postgresversions.catalog.kubedb.com                2024-02-14T09:05:19Z
proxysqlversions.catalog.kubedb.com                2024-02-14T09:05:19Z
publishers.postgres.kubedb.com                     2024-02-14T09:06:15Z
rabbitmqversions.catalog.kubedb.com                2024-02-14T09:05:19Z
redisautoscalers.autoscaling.kubedb.com            2024-02-14T09:06:18Z
redises.kubedb.com                                 2024-02-14T09:06:18Z
redisopsrequests.ops.kubedb.com                    2024-02-14T09:06:18Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-02-14T09:06:18Z
redissentinelopsrequests.ops.kubedb.com            2024-02-14T09:06:18Z
redissentinels.kubedb.com                          2024-02-14T09:06:18Z
redisversions.catalog.kubedb.com                   2024-02-14T09:05:19Z
singlestoreversions.catalog.kubedb.com             2024-02-14T09:05:19Z
solrversions.catalog.kubedb.com                    2024-02-14T09:05:19Z
subscribers.postgres.kubedb.com                    2024-02-14T09:06:15Z
zookeeperversions.catalog.kubedb.com               2024-02-14T09:05:19Z
```

## Deploy MongoDB Cluster with Arbiter

We are going to Deploy MongoDB Cluster with Arbiter by using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the MongoDB CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb
  namespace: demo
spec:
  version: "6.0.12"
  replicaSet:
    name: "rs0"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "default"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 512Mi
  arbiter:
    podTemplate: {}
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `mongodb.yaml` 
Then create the above MongoDB CRO

```bash
$ kubectl apply -f mongodb.yaml
mongodb.kubedb.com/mongodb created
```
In this yaml,
* `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `version 6.0.12`. You can list the KubeDB supported versions of MongoDB by running `$ kubectl get mongodbversions` command.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs.
* `spec.arbiter` denotes arbiter spec of the deployed MongoDB CRD. There are two fields under it, configSecret & podTemplate. `spec.arbiter.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used. `spec.arbiter.podTemplate` holds the arbiter-podSpec. `null` value of it, instructs kubedb operator to use the default arbiter podTemplate. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy) .

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                    READY   STATUS    RESTARTS   AGE
pod/mongodb-0           2/2     Running   0          5m26s
pod/mongodb-1           2/2     Running   0          2m52s
pod/mongodb-arbiter-0   1/1     Running   0          2m8s

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/mongodb        ClusterIP   10.96.168.152   <none>        27017/TCP   5m32s
service/mongodb-pods   ClusterIP   None            <none>        27017/TCP   5m32s

NAME                               READY   AGE
statefulset.apps/mongodb           2/2     5m26s
statefulset.apps/mongodb-arbiter   1/1     2m8s

NAME                                         TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb   kubedb.com/mongodb   6.0.12    117s

NAME                         VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb   6.0.12    Ready    5m32s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb
NAME      VERSION   STATUS   AGE
mongodb   6.0.12    Ready    5m48s
```
> We have successfully deployed MongoDB ReplicaSet with Arbiter in AKS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mongodb` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb
NAME           TYPE                       DATA   AGE
mongodb-auth   kubernetes.io/basic-auth   2      6m5s
mongodb-key    Opaque                     1      6m5s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb
NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mongodb        ClusterIP   10.96.168.152   <none>        27017/TCP   6m20s
mongodb-pods   ClusterIP   None            <none>        27017/TCP   6m20s
```
Now, we are going to use `mongodb-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mongodb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongodb-auth -o jsonpath='{.data.\password}' | base64 -d
9mOvP5rd4JZv18ul
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB and insert some sample data. 

```bash
$ kubectl exec -it mongodb-0 -n demo bash
mongodb@mongodb-0:/$ mongosh admin -u root -p '9mOvP5rd4JZv18ul'
Using MongoDB:		6.0.12
Using Mongosh:		2.1.1


rs0 [direct: primary] admin> rs.status()
{
  set: 'rs0',
  date: ISODate('2024-02-14T10:47:56.287Z'),
  myState: 1,
  term: Long('1'),
  syncSourceHost: '',
  syncSourceId: -1,
  heartbeatIntervalMillis: Long('2000'),
  majorityVoteCount: 2,
  writeMajorityCount: 2,
  votingMembersCount: 3,
  writableVotingMembersCount: 2,
  optimes: {
    lastCommittedOpTime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
    lastCommittedWallTime: ISODate('2024-02-14T10:47:48.432Z'),
    readConcernMajorityOpTime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
    appliedOpTime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
    durableOpTime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
    lastAppliedWallTime: ISODate('2024-02-14T10:47:48.432Z'),
    lastDurableWallTime: ISODate('2024-02-14T10:47:48.432Z')
  },
  lastStableRecoveryTimestamp: Timestamp({ t: 1708426058, i: 1 }),
  electionCandidateMetrics: {
    lastElectionReason: 'electionTimeout',
    lastElectionDate: ISODate('2024-02-14T10:42:38.399Z'),
    electionTerm: Long('1'),
    lastCommittedOpTimeAtElection: { ts: Timestamp({ t: 1708425758, i: 1 }), t: Long('-1') },
    lastSeenOpTimeAtElection: { ts: Timestamp({ t: 1708425758, i: 1 }), t: Long('-1') },
    numVotesNeeded: 1,
    priorityAtElection: 1,
    electionTimeoutMillis: Long('10000'),
    newTermStartDate: ISODate('2024-02-14T10:42:38.416Z'),
    wMajorityWriteAvailabilityDate: ISODate('2024-02-14T10:42:38.425Z')
  },
  members: [
    {
      _id: 0,
      name: 'mongodb-0.mongodb-pods.demo.svc.cluster.local:27017',
      health: 1,
      state: 1,
      stateStr: 'PRIMARY',
      uptime: 331,
      optime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
      optimeDate: ISODate('2024-02-14T10:47:48.000Z'),
      lastAppliedWallTime: ISODate('2024-02-14T10:47:48.432Z'),
      lastDurableWallTime: ISODate('2024-02-14T10:47:48.432Z'),
      syncSourceHost: '',
      syncSourceId: -1,
      infoMessage: '',
      electionTime: Timestamp({ t: 1708425758, i: 2 }),
      electionDate: ISODate('2024-02-14T10:42:38.000Z'),
      configVersion: 4,
      configTerm: 1,
      self: true,
      lastHeartbeatMessage: ''
    },
    {
      _id: 1,
      name: 'mongodb-1.mongodb-pods.demo.svc.cluster.local:27017',
      health: 1,
      state: 2,
      stateStr: 'SECONDARY',
      uptime: 273,
      optime: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
      optimeDurable: { ts: Timestamp({ t: 1708426068, i: 1 }), t: Long('1') },
      optimeDate: ISODate('2024-02-14T10:47:48.000Z'),
      optimeDurableDate: ISODate('2024-02-14T10:47:48.000Z'),
      lastAppliedWallTime: ISODate('2024-02-14T10:47:48.432Z'),
      lastDurableWallTime: ISODate('2024-02-14T10:47:48.432Z'),
      lastHeartbeat: ISODate('2024-02-14T10:47:54.508Z'),
      lastHeartbeatRecv: ISODate('2024-02-14T10:47:54.504Z'),
      pingMs: Long('0'),
      lastHeartbeatMessage: '',
      syncSourceHost: 'mongodb-0.mongodb-pods.demo.svc.cluster.local:27017',
      syncSourceId: 0,
      infoMessage: '',
      configVersion: 4,
      configTerm: 1
    },
    {
      _id: 2,
      name: 'mongodb-arbiter-0.mongodb-pods.demo.svc.cluster.local:27017',
      health: 1,
      state: 7,
      stateStr: 'ARBITER',
      uptime: 233,
      lastHeartbeat: ISODate('2024-02-14T10:47:54.507Z'),
      lastHeartbeatRecv: ISODate('2024-02-14T10:47:54.610Z'),
      pingMs: Long('0'),
      lastHeartbeatMessage: '',
      syncSourceHost: '',
      syncSourceId: -1,
      infoMessage: '',
      configVersion: 4,
      configTerm: 1
    }
  ],
  ok: 1,
  '$clusterTime': {
    clusterTime: Timestamp({ t: 1708426068, i: 1 }),
    signature: {
      hash: Binary.createFromBase64('nzsWSYtpLygWzfxA1SmmnBsARc4=', 0),
      keyId: Long('7337632758254010375')
    }
  },
  operationTime: Timestamp({ t: 1708426068, i: 1 })
}
```
Here you can see the arbiter pod in the members list of `rs.status()` output.

```bash
rs0 [direct: primary] admin> rs.isMaster().primary
mongodb-0.mongodb-pods.demo.svc.cluster.local:27017
rs0 [direct: primary] admin> show dbs
admin          172.00 KiB
config         288.00 KiB
kubedb-system   40.00 KiB
local          428.00 KiB
rs0 [direct: primary] admin> show users
[
  {
    _id: 'admin.root',
    userId: UUID('de2ca38f-ef66-4ab4-9000-1531b14b306c'),
    user: 'root',
    db: 'admin',
    roles: [ { role: 'root', db: 'admin' } ],
    mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
  }
]
rs0 [direct: primary] admin> use musicdb
switched to db musicdb
rs0 [direct: primary] musicdb> db.songs.insert({"name":"Take Me Home Country Roads"});
DeprecationWarning: Collection.insert() is deprecated. Use insertOne, insertMany, or bulkWrite.
{
  acknowledged: true,
  insertedIds: { '0': ObjectId('65d4842d73373080f0d7511f') }
}
rs0 [direct: primary] musicdb> exit
```

> We've successfully inserted some sample data to our MongoDB database. More information about Deploy & Manage MongoDB on Kubernetes can be found [Kubernetes MongoDB](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

Now, To check the data availability to the secondary members we'll now exec into `mongodb-1` pod (which is secondary member right now).

```bash
$ kubectl exec -it mongodb-1 -n demo bash
mongodb@mongodb-1:/$ mongosh admin -u root -p '9mOvP5rd4JZv18ul'
Using MongoDB:		6.0.12
Using Mongosh:		2.1.1

rs0 [direct: secondary] admin> rs.secondaryOk()

rs0 [direct: secondary] admin> show users
[
  {
    _id: 'admin.root',
    userId: UUID('de2ca38f-ef66-4ab4-9000-1531b14b306c'),
    user: 'root',
    db: 'admin',
    roles: [ { role: 'root', db: 'admin' } ],
    mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
  }
]
rs0 [direct: secondary] admin> use musicdb
switched to db musicdb
rs0 [direct: secondary] musicdb> db.songs.find().pretty()
[
  {
    _id: ObjectId('65d4842d73373080f0d7511f'),
    name: 'Take Me Home Country Roads'
  }
]
rs0 [direct: secondary] musicdb> exit
```
> we've successfully access the data in `mongodb-1`. So, the data is available to the secondary members.

If you want to learn more about Production-Grade MongoDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MongoDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
