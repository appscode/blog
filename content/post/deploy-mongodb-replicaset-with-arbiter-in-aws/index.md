---
title: Deploy MongoDB ReplicaSet with Arbiter in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-09-15"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- eks
- kubedb
- kubernetes
- mongodb
- mongodb-arbiter
- mongodb-database
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Elasticsearch, MySQL, MariaDB, Kafka, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy MongoDB ReplicaSet with Arbiter in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy MongoDB ReplicaSet with Arbiter
3) Connect MongoDB with Arbiter
4) Read/Write with Arbiter

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
appscode/kubedb                   	v2023.08.18  	v2023.08.18	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.20.0      	v0.20.1    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.08.18  	v2023.08.18	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.08.18  	v2023.08.18	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.11.0      	v0.11.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.08.18  	v2023.08.18	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.08.18  	v2023.08.18	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.08.18  	v2023.08.18	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.22.0      	v0.22.8    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.08.18  	v2023.08.18	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.35.0      	v0.35.6    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.11.0      	v0.11.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.4.3      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.11.0      	v0.11.1    	KubeDB Webhook Server by AppsCode 



# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.08.18 \
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
kubedb      kubedb-kubedb-autoscaler-ccd798cf6-m5stn        1/1     Running   0          2m30s
kubedb      kubedb-kubedb-dashboard-b68496968-st22m         1/1     Running   0          2m30s
kubedb      kubedb-kubedb-ops-manager-65b6bf54b4-xnc2h      1/1     Running   0          2m30s
kubedb      kubedb-kubedb-provisioner-8bdccf856-b6bnr       1/1     Running   0          2m30s
kubedb      kubedb-kubedb-schema-manager-78cb7dbd47-c4mgk   1/1     Running   0          2m30s
kubedb      kubedb-kubedb-webhook-server-75566f45d5-4mg59   1/1     Running   0          2m30s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-09-18T07:11:37Z
elasticsearchdashboards.dashboard.kubedb.com      2023-09-18T07:10:59Z
elasticsearches.kubedb.com                        2023-09-18T07:10:59Z
elasticsearchopsrequests.ops.kubedb.com           2023-09-18T07:11:15Z
elasticsearchversions.catalog.kubedb.com          2023-09-18T07:10:37Z
etcds.kubedb.com                                  2023-09-18T07:12:03Z
etcdversions.catalog.kubedb.com                   2023-09-18T07:10:37Z
kafkas.kubedb.com                                 2023-09-18T07:11:46Z
kafkaversions.catalog.kubedb.com                  2023-09-18T07:10:37Z
mariadbautoscalers.autoscaling.kubedb.com         2023-09-18T07:11:37Z
mariadbdatabases.schema.kubedb.com                2023-09-18T07:11:48Z
mariadbopsrequests.ops.kubedb.com                 2023-09-18T07:11:28Z
mariadbs.kubedb.com                               2023-09-18T07:11:28Z
mariadbversions.catalog.kubedb.com                2023-09-18T07:10:37Z
memcacheds.kubedb.com                             2023-09-18T07:12:03Z
memcachedversions.catalog.kubedb.com              2023-09-18T07:10:37Z
mongodbautoscalers.autoscaling.kubedb.com         2023-09-18T07:11:37Z
mongodbdatabases.schema.kubedb.com                2023-09-18T07:11:48Z
mongodbopsrequests.ops.kubedb.com                 2023-09-18T07:11:18Z
mongodbs.kubedb.com                               2023-09-18T07:11:18Z
mongodbversions.catalog.kubedb.com                2023-09-18T07:10:37Z
mysqlautoscalers.autoscaling.kubedb.com           2023-09-18T07:11:37Z
mysqldatabases.schema.kubedb.com                  2023-09-18T07:11:47Z
mysqlopsrequests.ops.kubedb.com                   2023-09-18T07:11:25Z
mysqls.kubedb.com                                 2023-09-18T07:11:25Z
mysqlversions.catalog.kubedb.com                  2023-09-18T07:10:37Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-09-18T07:11:37Z
perconaxtradbopsrequests.ops.kubedb.com           2023-09-18T07:11:40Z
perconaxtradbs.kubedb.com                         2023-09-18T07:11:40Z
perconaxtradbversions.catalog.kubedb.com          2023-09-18T07:10:37Z
pgbouncers.kubedb.com                             2023-09-18T07:11:21Z
pgbouncerversions.catalog.kubedb.com              2023-09-18T07:10:37Z
postgresautoscalers.autoscaling.kubedb.com        2023-09-18T07:11:37Z
postgresdatabases.schema.kubedb.com               2023-09-18T07:11:48Z
postgreses.kubedb.com                             2023-09-18T07:11:34Z
postgresopsrequests.ops.kubedb.com                2023-09-18T07:11:34Z
postgresversions.catalog.kubedb.com               2023-09-18T07:10:37Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-09-18T07:11:37Z
proxysqlopsrequests.ops.kubedb.com                2023-09-18T07:11:37Z
proxysqls.kubedb.com                              2023-09-18T07:11:37Z
proxysqlversions.catalog.kubedb.com               2023-09-18T07:10:37Z
publishers.postgres.kubedb.com                    2023-09-18T07:11:50Z
redisautoscalers.autoscaling.kubedb.com           2023-09-18T07:11:37Z
redises.kubedb.com                                2023-09-18T07:11:31Z
redisopsrequests.ops.kubedb.com                   2023-09-18T07:11:31Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-09-18T07:11:37Z
redissentinelopsrequests.ops.kubedb.com           2023-09-18T07:11:43Z
redissentinels.kubedb.com                         2023-09-18T07:11:43Z
redisversions.catalog.kubedb.com                  2023-09-18T07:10:37Z
subscribers.postgres.kubedb.com                   2023-09-18T07:11:53Z
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
  version: "6.0.5"
  replicaSet:
    name: "rs0"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "gp2"
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
* `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `version 6.0.5`. You can list the KubeDB supported versions of MongoDB by running `$ kubectl get mongodbversions` command.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs.
* `spec.arbiter` denotes arbiter spec of the deployed MongoDB CRD. There are two fields under it, configSecret & podTemplate. `spec.arbiter.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used. `spec.arbiter.podTemplate` holds the arbiter-podSpec. `null` value of it, instructs kubedb operator to use the default arbiter podTemplate. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy) .

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                    READY   STATUS    RESTARTS   AGE
pod/mongodb-0           2/2     Running   0          3m46s
pod/mongodb-1           2/2     Running   0          2m32s
pod/mongodb-arbiter-0   1/1     Running   0          108s

NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mongodb        ClusterIP   10.96.249.83   <none>        27017/TCP   3m51s
service/mongodb-pods   ClusterIP   None           <none>        27017/TCP   3m51s

NAME                               READY   AGE
statefulset.apps/mongodb           2/2     3m46s
statefulset.apps/mongodb-arbiter   1/1     108s

NAME                                         TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb   kubedb.com/mongodb   6.0.5     98s

NAME                         VERSION   STATUS   AGE
mongodb.kubedb.com/mongodb   6.0.5     Ready    3m51s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb
NAME      VERSION   STATUS   AGE
mongodb   6.0.5     Ready    4m16s
```
> We have successfully deployed MongoDB ReplicaSet with Arbiter in AWS. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database `mongodb` that we have deployed. Let’s check them using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mongodb
NAME           TYPE                       DATA   AGE
mongodb-auth   kubernetes.io/basic-auth   2      4m37s
mongodb-key    Opaque                     1      4m37s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mongodb
NAME           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mongodb        ClusterIP   10.96.249.83   <none>        27017/TCP   4m50s
mongodb-pods   ClusterIP   None           <none>        27017/TCP   4m50s
```
Now, we are going to use `mongodb-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mongodb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mongodb-auth -o jsonpath='{.data.\password}' | base64 -d
lNk7SIpOW~jdav6L
```

#### Insert Sample Data

In this section, we are going to login into our MongoDB and insert some sample data. 

```bash
$ kubectl exec -it mongodb-0 -n demo bash
root@mongodb-0:/# mongosh admin -u root -p 'lNk7SIpOW~jdav6L'
Using MongoDB:		6.0.5
Using Mongosh:		1.8.2

rs0 [direct: primary] admin> rs.status()
{
  set: 'rs0',
  date: ISODate("2023-09-18T07:28:48.328Z"),
  myState: 1,
  term: Long("1"),
  syncSourceHost: '',
  syncSourceId: -1,
  heartbeatIntervalMillis: Long("2000"),
  majorityVoteCount: 2,
  writeMajorityCount: 2,
  votingMembersCount: 3,
  writableVotingMembersCount: 2,
  optimes: {
    lastCommittedOpTime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
    lastCommittedWallTime: ISODate("2023-09-18T07:28:39.527Z"),
    readConcernMajorityOpTime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
    appliedOpTime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
    durableOpTime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
    lastAppliedWallTime: ISODate("2023-09-18T07:28:39.527Z"),
    lastDurableWallTime: ISODate("2023-09-18T07:28:39.527Z")
  },
  lastStableRecoveryTimestamp: Timestamp({ t: 1695022069, i: 1 }),
  electionCandidateMetrics: {
    lastElectionReason: 'electionTimeout',
    lastElectionDate: ISODate("2023-09-18T07:22:49.491Z"),
    electionTerm: Long("1"),
    lastCommittedOpTimeAtElection: { ts: Timestamp({ t: 1695021769, i: 1 }), t: Long("-1") },
    lastSeenOpTimeAtElection: { ts: Timestamp({ t: 1695021769, i: 1 }), t: Long("-1") },
    numVotesNeeded: 1,
    priorityAtElection: 1,
    electionTimeoutMillis: Long("10000"),
    newTermStartDate: ISODate("2023-09-18T07:22:49.508Z"),
    wMajorityWriteAvailabilityDate: ISODate("2023-09-18T07:22:49.517Z")
  },
  members: [
    {
      _id: 0,
      name: 'mongodb-0.mongodb-pods.demo.svc.cluster.local:27017',
      health: 1,
      state: 1,
      stateStr: 'PRIMARY',
      uptime: 372,
      optime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
      optimeDate: ISODate("2023-09-18T07:28:39.000Z"),
      lastAppliedWallTime: ISODate("2023-09-18T07:28:39.527Z"),
      lastDurableWallTime: ISODate("2023-09-18T07:28:39.527Z"),
      syncSourceHost: '',
      syncSourceId: -1,
      infoMessage: '',
      electionTime: Timestamp({ t: 1695021769, i: 2 }),
      electionDate: ISODate("2023-09-18T07:22:49.000Z"),
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
      uptime: 315,
      optime: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
      optimeDurable: { ts: Timestamp({ t: 1695022119, i: 1 }), t: Long("1") },
      optimeDate: ISODate("2023-09-18T07:28:39.000Z"),
      optimeDurableDate: ISODate("2023-09-18T07:28:39.000Z"),
      lastAppliedWallTime: ISODate("2023-09-18T07:28:39.527Z"),
      lastDurableWallTime: ISODate("2023-09-18T07:28:39.527Z"),
      lastHeartbeat: ISODate("2023-09-18T07:28:46.871Z"),
      lastHeartbeatRecv: ISODate("2023-09-18T07:28:46.871Z"),
      pingMs: Long("0"),
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
      uptime: 281,
      lastHeartbeat: ISODate("2023-09-18T07:28:46.872Z"),
      lastHeartbeatRecv: ISODate("2023-09-18T07:28:47.000Z"),
      pingMs: Long("0"),
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
    clusterTime: Timestamp({ t: 1695022119, i: 1 }),
    signature: {
      hash: Binary(Buffer.from("fe08f082ee8ad4de7dd57e57b293095d15e7fced", "hex"), 0),
      keyId: Long("7280063063863066631")
    }
  },
  operationTime: Timestamp({ t: 1695022119, i: 1 })
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
    userId: new UUID("c5387bf3-5ad5-43f4-b7dc-73dea0f810b9"),
    user: 'root',
    db: 'admin',
    roles: [ { role: 'root', db: 'admin' } ],
    mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
  }
]
rs0 [direct: primary] admin> use musicdb
switched to db musicdb
rs0 [direct: primary] musicdb> db.songs.insert({"name":"Annie's Song"});
DeprecationWarning: Collection.insert() is deprecated. Use insertOne, insertMany, or bulkWrite.
{
  acknowledged: true,
  insertedIds: { '0': ObjectId("6507fcee87f0f74bc51d6ddd") }
}
rs0 [direct: primary] musicdb> exit
```

> We've successfully inserted some sample data to our MongoDB database. More information about Run & Manage MongoDB on Kubernetes can be found [Kubernetes MongoDB](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

Now, To check the data availability to the secondary members we'll now exec into `mongodb-1` pod (which is secondary member right now).


```bash
$ kubectl exec -it mongodb-1 -n demo bash
root@mongodb-1:/# mongosh admin -u root -p 'lNk7SIpOW~jdav6L'
Using MongoDB:		6.0.5
Using Mongosh:		1.8.2


rs0 [direct: secondary] admin> rs.secondaryOk()
DeprecationWarning: .setSecondaryOk() is deprecated. Use .setReadPref("primaryPreferred") instead
Setting read preference from "primary" to "primaryPreferred"

rs0 [direct: secondary] admin> show users
[
  {
    _id: 'admin.root',
    userId: new UUID("c5387bf3-5ad5-43f4-b7dc-73dea0f810b9"),
    user: 'root',
    db: 'admin',
    roles: [ { role: 'root', db: 'admin' } ],
    mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
  }
]
rs0 [direct: secondary] admin> use musicdb
switched to db musicdb
rs0 [direct: secondary] musicdb> db.songs.find().pretty()
[ { _id: ObjectId("6507fcee87f0f74bc51d6ddd"), name: "Annie's Song" } ]
rs0 [direct: secondary] musicdb> exit
```
> we've successfully access the data in `mongodb-1`. So, the data is available to the secondary members.

If you want to learn more about Production-Grade MongoDB you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MongoDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
