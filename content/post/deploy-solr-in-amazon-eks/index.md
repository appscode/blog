---
title: Deploy Solr in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-09-13"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- solr
- solr-cluster
- solr-database
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Solr in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1. Install KubeDB
2. Deploy ZooKeeper
3. Deploy Solr Cluster
4. Read/Write Sample Data

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB.

![License Server](AppscodeLicense.png)

### Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2024.8.21   	v2024.8.21 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.32.0      	v0.32.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.8.21   	v2024.8.21 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.2.0       	v0.2.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.8.21   	v2024.8.21 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.23.0      	v0.23.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.8.21   	v2024.8.21 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.8.21   	v2024.8.21 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.8.21   	v2024.8.21 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.34.0      	v0.34.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.8.21   	v2024.8.21 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.47.0      	v0.47.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.23.0      	v0.23.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.8.21   	0.7.6      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.8.21   	v2024.8.21 	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.23.0      	v0.23.1    	KubeDB Webhook Server by AppsCode


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.8.21 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.Solr=true \
  --set global.featureGates.ZooKeeper=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-7bf9c48b5c-sk6wq       1/1     Running   0          2m15s
kubedb      kubedb-kubedb-ops-manager-56bbd9b584-9wrmh      1/1     Running   0          2m15s
kubedb      kubedb-kubedb-provisioner-595f6757cd-hmgvx      1/1     Running   0          2m15s
kubedb      kubedb-kubedb-webhook-server-574f8d5767-4gj6p   1/1     Running   0          2m15s
kubedb      kubedb-petset-operator-77b6b9897f-69g2n         1/1     Running   0          2m15s
kubedb      kubedb-petset-webhook-server-75b578785f-wc469   2/2     Running   0          2m15s
kubedb      kubedb-sidekick-c898cff4c-h99wd                 1/1     Running   0          2m15s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-09-19T07:14:58Z
connectclusters.kafka.kubedb.com                   2024-09-19T07:15:34Z
connectors.kafka.kubedb.com                        2024-09-19T07:15:34Z
druidversions.catalog.kubedb.com                   2024-09-19T07:14:58Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-09-19T07:15:30Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-09-19T07:15:30Z
elasticsearches.kubedb.com                         2024-09-19T07:15:30Z
elasticsearchopsrequests.ops.kubedb.com            2024-09-19T07:15:30Z
elasticsearchversions.catalog.kubedb.com           2024-09-19T07:14:58Z
etcdversions.catalog.kubedb.com                    2024-09-19T07:14:58Z
ferretdbversions.catalog.kubedb.com                2024-09-19T07:14:58Z
kafkaautoscalers.autoscaling.kubedb.com            2024-09-19T07:15:34Z
kafkaconnectorversions.catalog.kubedb.com          2024-09-19T07:14:58Z
kafkaopsrequests.ops.kubedb.com                    2024-09-19T07:15:34Z
kafkas.kubedb.com                                  2024-09-19T07:15:34Z
kafkaversions.catalog.kubedb.com                   2024-09-19T07:14:58Z
mariadbarchivers.archiver.kubedb.com               2024-09-19T07:15:37Z
mariadbautoscalers.autoscaling.kubedb.com          2024-09-19T07:15:37Z
mariadbdatabases.schema.kubedb.com                 2024-09-19T07:15:37Z
mariadbopsrequests.ops.kubedb.com                  2024-09-19T07:15:37Z
mariadbs.kubedb.com                                2024-09-19T07:15:37Z
mariadbversions.catalog.kubedb.com                 2024-09-19T07:14:58Z
memcachedversions.catalog.kubedb.com               2024-09-19T07:14:58Z
mongodbarchivers.archiver.kubedb.com               2024-09-19T07:15:41Z
mongodbautoscalers.autoscaling.kubedb.com          2024-09-19T07:15:41Z
mongodbdatabases.schema.kubedb.com                 2024-09-19T07:15:41Z
mongodbopsrequests.ops.kubedb.com                  2024-09-19T07:15:41Z
mongodbs.kubedb.com                                2024-09-19T07:15:41Z
mongodbversions.catalog.kubedb.com                 2024-09-19T07:14:58Z
mssqlserverversions.catalog.kubedb.com             2024-09-19T07:14:58Z
mysqlarchivers.archiver.kubedb.com                 2024-09-19T07:15:45Z
mysqlautoscalers.autoscaling.kubedb.com            2024-09-19T07:15:44Z
mysqldatabases.schema.kubedb.com                   2024-09-19T07:15:45Z
mysqlopsrequests.ops.kubedb.com                    2024-09-19T07:15:44Z
mysqls.kubedb.com                                  2024-09-19T07:15:44Z
mysqlversions.catalog.kubedb.com                   2024-09-19T07:14:58Z
perconaxtradbversions.catalog.kubedb.com           2024-09-19T07:14:58Z
pgbouncerversions.catalog.kubedb.com               2024-09-19T07:14:58Z
pgpoolversions.catalog.kubedb.com                  2024-09-19T07:14:58Z
postgresarchivers.archiver.kubedb.com              2024-09-19T07:15:48Z
postgresautoscalers.autoscaling.kubedb.com         2024-09-19T07:15:48Z
postgresdatabases.schema.kubedb.com                2024-09-19T07:15:48Z
postgreses.kubedb.com                              2024-09-19T07:15:48Z
postgresopsrequests.ops.kubedb.com                 2024-09-19T07:15:48Z
postgresversions.catalog.kubedb.com                2024-09-19T07:14:58Z
proxysqlversions.catalog.kubedb.com                2024-09-19T07:14:58Z
publishers.postgres.kubedb.com                     2024-09-19T07:15:48Z
rabbitmqversions.catalog.kubedb.com                2024-09-19T07:14:58Z
redisautoscalers.autoscaling.kubedb.com            2024-09-19T07:15:51Z
redises.kubedb.com                                 2024-09-19T07:15:51Z
redisopsrequests.ops.kubedb.com                    2024-09-19T07:15:51Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-09-19T07:15:51Z
redissentinelopsrequests.ops.kubedb.com            2024-09-19T07:15:51Z
redissentinels.kubedb.com                          2024-09-19T07:15:51Z
redisversions.catalog.kubedb.com                   2024-09-19T07:14:58Z
restproxies.kafka.kubedb.com                       2024-09-19T07:15:34Z
schemaregistries.kafka.kubedb.com                  2024-09-19T07:15:34Z
schemaregistryversions.catalog.kubedb.com          2024-09-19T07:14:58Z
singlestoreversions.catalog.kubedb.com             2024-09-19T07:14:59Z
solrautoscalers.autoscaling.kubedb.com             2024-09-19T07:15:55Z
solrs.kubedb.com                                   2024-09-19T07:15:55Z
solrversions.catalog.kubedb.com                    2024-09-19T07:14:59Z
subscribers.postgres.kubedb.com                    2024-09-19T07:15:48Z
zookeeperautoscalers.autoscaling.kubedb.com        2024-09-19T07:15:58Z
zookeepers.kubedb.com                              2024-09-19T07:15:58Z
zookeeperversions.catalog.kubedb.com               2024-09-19T07:14:59Z
```

### Create a Namespace

To keep resources isolated, we'll use a separate namespace called `demo` throughout this tutorial.
Run the following command to create the namespace:

```bash
$ kubectl create namespace demo
namespace/demo created
```

## Create ZooKeeper Instance

Since KubeDB Solr operates in `solrcloud` mode, it requires an external ZooKeeper to manage replica distribution and configuration.

In this tutorial, we will use KubeDB ZooKeeper. Below is the configuration for the ZooKeeper instance we'll create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zookeeper
  namespace: demo
spec:
  version: 3.9.1
  replicas: 3
  adminServerPort: 8080
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: gp2
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```
Let’s save this yaml configuration into `zookeeper.yaml` Then create the above ZooKeeper CRO,

```bash
$ kubectl apply -f zookeeper.yaml 
zookeeper.kubedb.com/zookeeper created
```

In this yaml,

- `spec.version` field specifies the version of ZooKeeper Here, we are using ZooKeeper `version 3.9.1`. You can list the KubeDB supported versions of ZooKeeper by running `$ kubectl get zookeeperversions` command.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”.

Once the ZooKeeper instance’s `STATUS` is `Ready`, we can proceed to deploy Solr in our cluster.

```bash
$ kubectl get zookeeper -n demo zookeeper
NAME        TYPE                  VERSION   STATUS   AGE
zookeeper   kubedb.com/v1alpha2   3.9.1     Ready    77s
```

## Deploy Solr Cluster

Here is the yaml of the Solr we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  replicas: 3
  zookeeperRef:
    name: zookeeper
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: gp2
  deletionPolicy: "WipeOut"
```

Let's save this yaml configuration into `solr-cluster.yaml` 
Then apply the above Solr yaml,

```bash
$ kubectl apply -f solr-cluster.yaml 
solr.kubedb.com/solr-cluster created
```

In this yaml,
- `spec.version` is the name of the SolrVersion CR. Here, a Solr of `version 9.4.1` will be created.
- `spec.replicas` specifies the number of Solr nodes.
- `spec.storageType` specifies the type of storage that will be used for Solr database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Solr database using `EmptyDir` volume. In this case, you don’t have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the Petset created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don’t specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”.

Once these are handled correctly and the Solr object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/solr-cluster-0   1/1     Running   0          2m56s
pod/solr-cluster-1   1/1     Running   0          52s
pod/solr-cluster-2   1/1     Running   0          44s
pod/zookeeper-0      1/1     Running   0          5m6s
pod/zookeeper-1      1/1     Running   0          4m37s
pod/zookeeper-2      1/1     Running   0          4m28s

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/solr-cluster             ClusterIP   10.96.247.137   <none>        8983/TCP                     2m58s
service/solr-cluster-pods        ClusterIP   None            <none>        8983/TCP                     2m58s
service/zookeeper                ClusterIP   10.96.179.181   <none>        2181/TCP                     5m10s
service/zookeeper-admin-server   ClusterIP   10.96.99.105    <none>        8080/TCP                     5m10s
service/zookeeper-pods           ClusterIP   None            <none>        2181/TCP,2888/TCP,3888/TCP   5m10s

NAME                                              TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/solr-cluster   kubedb.com/solr        9.4.1     2m58s
appbinding.appcatalog.appscode.com/zookeeper      kubedb.com/zookeeper   3.9.1     5m10s

NAME                           TYPE                  VERSION   STATUS   AGE
solr.kubedb.com/solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    2m58s

NAME                             TYPE                  VERSION   STATUS   AGE
zookeeper.kubedb.com/zookeeper   kubedb.com/v1alpha2   3.9.1     Ready    5m10s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get solr -n demo solr-cluster
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    3m21s
```
> We have successfully deployed Solr in Amazon EKS. Now we can exec into the container to use the database.


## Connect with Solr Database

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our Solr database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Solr database is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
solr-cluster             ClusterIP   10.96.247.137   <none>        8983/TCP                     3m53s
solr-cluster-pods        ClusterIP   None            <none>        8983/TCP                     3m53s
zookeeper                ClusterIP   10.96.179.181   <none>        2181/TCP                     6m5s
zookeeper-admin-server   ClusterIP   10.96.99.105    <none>        8080/TCP                     6m5s
zookeeper-pods           ClusterIP   None            <none>        2181/TCP,2888/TCP,3888/TCP   6m5s
```
To connect to the Solr database, we will use the `solr-cluster` service. First, we need to port-forward the `solr-cluster` service to port `8983` on the local machine:

```bash
$ kubectl port-forward -n demo svc/solr-cluster 8983
Forwarding from 127.0.0.1:8983 -> 8983
Forwarding from [::1]:8983 -> 8983
```
Now, the Solr cluster is accessible at `localhost:8983`.

#### Export the Credentials

KubeDB creates several Secrets for managing the database. To view the Secrets created for `solr-cluster`, run the following command:

```bash
$ kubectl get secret -n demo | grep solr-cluster
solr-cluster-admin-cred           kubernetes.io/basic-auth   2      4m27s
solr-cluster-auth-config          Opaque                     1      4m27s
solr-cluster-config               Opaque                     1      4m27s
solr-cluster-zk-digest            kubernetes.io/basic-auth   2      4m27s
solr-cluster-zk-digest-readonly   kubernetes.io/basic-auth   2      4m27s
```
From the above list, the `solr-cluster-admin-cred` Secret contains the admin-level credentials needed to connect to the database.

### Accessing Database Through CLI

To access the database via the CLI, you first need to retrieve the credentials. Use the following commands to obtain the username and password:

```bash
$ kubectl get secret -n demo solr-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo solr-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
3cYYLNQ0!))Ksly3
```

Now, let's check the health of our Solr cluster.

```bash
# curl -XGET -k -u 'username:password' "http://localhost:8983/solr/admin/collections?action=CLUSTERSTATUS"
$ curl -XGET -k -u 'admin:3cYYLNQ0!))Ksly3' "http://localhost:8983/solr/admin/collections?action=CLUSTERSTATUS"

{
  "responseHeader":{
    "status":0,
    "QTime":1
  },
  "cluster":{
    "collections":{
      "kubedb-system":{
        "pullReplicas":"0",
        "configName":"kubedb-system.AUTOCREATED",
        "replicationFactor":1,
        "router":{
          "name":"compositeId"
        },
        "nrtReplicas":1,
        "tlogReplicas":"0",
        "shards":{
          "shard1":{
            "range":"80000000-7fffffff",
            "state":"active",
            "replicas":{
              "core_node2":{
                "core":"kubedb-system_shard1_replica_n1",
                "node_name":"solr-cluster-1.solr-cluster-pods.demo:8983_solr",
                "type":"NRT",
                "state":"active",
                "leader":"true",
                "force_set_state":"false",
                "base_url":"http://solr-cluster-1.solr-cluster-pods.demo:8983/solr"
              }
            },
            "health":"GREEN"
          }
        },
        "health":"GREEN",
        "znodeVersion":4
      }
    },
    "live_nodes":["solr-cluster-1.solr-cluster-pods.demo:8983_solr","solr-cluster-2.solr-cluster-pods.demo:8983_solr","solr-cluster-0.solr-cluster-pods.demo:8983_solr"]
  }
}
```


### Insert Sample Data

In this section, we'll create a collection in Solr and insert some sample data using `curl`. To disable certificate verification (useful for testing with self-signed certificates), use the `-k` flag.

Execute the following command to create a collection named `music` in Solr:

```bash
$ curl -XPOST -k -u 'admin:3cYYLNQ0!))Ksly3' "http://localhost:8983/solr/admin/collections?action=CREATE&name=music&numShards=2&replicationFactor=2&wt=xml"


<?xml version="1.0" encoding="UTF-8"?>
<response>

<lst name="responseHeader">
  <int name="status">0</int>
  <int name="QTime">3712</int>
</lst>
<lst name="success">
  <lst name="solr-cluster-1.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">2428</int>
    </lst>
    <str name="core">music_shard1_replica_n2</str>
  </lst>
  <lst name="solr-cluster-2.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">2634</int>
    </lst>
    <str name="core">music_shard2_replica_n1</str>
  </lst>
  <lst name="solr-cluster-2.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">2869</int>
    </lst>
    <str name="core">music_shard1_replica_n6</str>
  </lst>
  <lst name="solr-cluster-0.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">3031</int>
    </lst>
    <str name="core">music_shard2_replica_n4</str>
  </lst>
</lst>
</response>



$ curl -X POST -u 'admin:3cYYLNQ0!))Ksly3' -H 'Content-Type: application/json' "http://localhost:8983/solr/music/update" --data-binary '[{ "Artist": "Avicii","Song": "The Nights"}]'

{
  "responseHeader":{
    "rf":2,
    "status":0,
    "QTime":527
  }
}
```

To verify that the collection has been created successfully, run the following command:

```bash
$ curl -X GET -u 'admin:3cYYLNQ0!))Ksly3' 'http://localhost:8983/solr/admin/collections?action=LIST&wt=json'
{
  "responseHeader":{
    "status":0,
    "QTime":0
  },
  "collections":["kubedb-system","music"]
}
```
To check the sample data in the `music` collection, use the following command:

```bash
$ curl -X GET -u 'admin:3cYYLNQ0!))Ksly3' "http://localhost:8983/solr/music/select" -H 'Content-Type: application/json' -d '{"query": "*:*"}'
{
  "responseHeader":{
    "zkConnected":true,
    "status":0,
    "QTime":87,
    "params":{
      "json":"{\"query\": \"*:*\"}"
    }
  },
  "response":{
    "numFound":1,
    "start":0,
    "maxScore":1.0,
    "numFoundExact":true,
    "docs":[{
      "Artist":["Avicii"],
      "Song":["The Nights"],
      "id":"798b62b5-adcf-4ed3-b83e-af79efe019f6",
      "_version_":1810618668053168128
    }]
  }
}
```

> We've successfully inserted some sample data to our Solr database. More information about Deploy & Manage Production-Grade Solr Database on Kubernetes can be found in [Solr Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-solr-on-kubernetes/)

We have made a in depth tutorial on Provision and Manage Solr on Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/PH4VgF35ryo?si=u-Ro-DmCy84K-3Ya" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Solr on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-solr-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
