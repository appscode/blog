---
title: Update Version of OpenSearch Database in Google Kubernetes Engine (GKE)
date: "2024-01-19"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- gcs
- gke
- kubedb
- kubernetes
- opensearch
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic and [OpenSearch](https://opensearch.org/) by AWS, but also other open source distributions like [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will update version of OpenSearch database in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy OpenSearch Cluster
3) Insert Sample Data
4) Update OpenSearch Database Version


### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID, we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8e336615-0dbb-4ae8-b72f-2e7ec34c399d
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB.

![License Server](AppscodeLicense.png)

### Install KubeDB

We will use helm to install KubeDB. Please install [helm](https://helm.sh/docs/intro/install/), if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm search repo appscode/kubedb
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2023.12.28  	v2023.12.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.25.0      	v0.25.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.12.28  	v2023.12.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.12.28  	v2023.12.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.16.0      	v0.16.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.12.28  	v2023.12.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2023.12.28  	v2023.12.28	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2023.12.28  	v2023.12.28	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.27.0      	v0.27.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.12.28  	v2023.12.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2023.12.28  	v0.2.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.40.0      	v0.40.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.16.0      	v0.16.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.12.20  	0.6.1      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.16.0      	v0.16.0    	KubeDB Webhook Server by AppsCode   

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2023.12.28 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                           READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-7dbf9844b4-jzw4r      1/1     Running   0          2m33s
kubedb      kubedb-kubedb-dashboard-759c885467-p78fr       1/1     Running   0          2m33s
kubedb      kubedb-kubedb-ops-manager-84c6dd4b4f-dmw86     1/1     Running   0          2m33s
kubedb      kubedb-kubedb-provisioner-6788588dd4-j8f48     1/1     Running   0          2m33s
kubedb      kubedb-kubedb-webhook-server-d8c665cc8-6gdgg   1/1     Running   0          2m33s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2024-01-19T06:07:12Z
elasticsearchdashboards.dashboard.kubedb.com      2024-01-19T06:07:12Z
elasticsearches.kubedb.com                        2024-01-19T06:07:12Z
elasticsearchopsrequests.ops.kubedb.com           2024-01-19T06:07:15Z
elasticsearchversions.catalog.kubedb.com          2024-01-19T06:05:12Z
etcds.kubedb.com                                  2024-01-19T06:07:15Z
etcdversions.catalog.kubedb.com                   2024-01-19T06:05:12Z
kafkaopsrequests.ops.kubedb.com                   2024-01-19T06:07:54Z
kafkas.kubedb.com                                 2024-01-19T06:07:19Z
kafkaversions.catalog.kubedb.com                  2024-01-19T06:05:12Z
mariadbautoscalers.autoscaling.kubedb.com         2024-01-19T06:07:13Z
mariadbopsrequests.ops.kubedb.com                 2024-01-19T06:07:33Z
mariadbs.kubedb.com                               2024-01-19T06:07:15Z
mariadbversions.catalog.kubedb.com                2024-01-19T06:05:12Z
memcacheds.kubedb.com                             2024-01-19T06:07:16Z
memcachedversions.catalog.kubedb.com              2024-01-19T06:05:13Z
mongodbarchivers.archiver.kubedb.com              2024-01-19T06:07:20Z
mongodbautoscalers.autoscaling.kubedb.com         2024-01-19T06:07:13Z
mongodbopsrequests.ops.kubedb.com                 2024-01-19T06:07:19Z
mongodbs.kubedb.com                               2024-01-19T06:07:16Z
mongodbversions.catalog.kubedb.com                2024-01-19T06:05:13Z
mysqlarchivers.archiver.kubedb.com                2024-01-19T06:07:21Z
mysqlautoscalers.autoscaling.kubedb.com           2024-01-19T06:07:13Z
mysqlopsrequests.ops.kubedb.com                   2024-01-19T06:07:30Z
mysqls.kubedb.com                                 2024-01-19T06:07:17Z
mysqlversions.catalog.kubedb.com                  2024-01-19T06:05:13Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2024-01-19T06:07:13Z
perconaxtradbopsrequests.ops.kubedb.com           2024-01-19T06:07:47Z
perconaxtradbs.kubedb.com                         2024-01-19T06:07:17Z
perconaxtradbversions.catalog.kubedb.com          2024-01-19T06:05:14Z
pgbouncers.kubedb.com                             2024-01-19T06:07:17Z
pgbouncerversions.catalog.kubedb.com              2024-01-19T06:05:14Z
postgresarchivers.archiver.kubedb.com             2024-01-19T06:07:22Z
postgresautoscalers.autoscaling.kubedb.com        2024-01-19T06:07:13Z
postgreses.kubedb.com                             2024-01-19T06:07:18Z
postgresopsrequests.ops.kubedb.com                2024-01-19T06:07:40Z
postgresversions.catalog.kubedb.com               2024-01-19T06:05:14Z
proxysqlautoscalers.autoscaling.kubedb.com        2024-01-19T06:07:13Z
proxysqlopsrequests.ops.kubedb.com                2024-01-19T06:07:43Z
proxysqls.kubedb.com                              2024-01-19T06:07:18Z
proxysqlversions.catalog.kubedb.com               2024-01-19T06:05:15Z
publishers.postgres.kubedb.com                    2024-01-19T06:07:57Z
redisautoscalers.autoscaling.kubedb.com           2024-01-19T06:07:13Z
redises.kubedb.com                                2024-01-19T06:07:18Z
redisopsrequests.ops.kubedb.com                   2024-01-19T06:07:36Z
redissentinelautoscalers.autoscaling.kubedb.com   2024-01-19T06:07:13Z
redissentinelopsrequests.ops.kubedb.com           2024-01-19T06:07:50Z
redissentinels.kubedb.com                         2024-01-19T06:07:19Z
redisversions.catalog.kubedb.com                  2024-01-19T06:05:15Z
subscribers.postgres.kubedb.com                   2024-01-19T06:08:00Z
```

## Deploy OpenSearch Cluster

Now, We are going to use the KubeDB-provided Custom Resource object `OpenSearch` for deployment. First, let’s create a Namespace in which we will deploy the cluster.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: os-cluster
  namespace: demo
spec:
  version: opensearch-2.0.1
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `os-cluster.yaml` 
Then apply the above OpenSearch yaml,

```bash
$ kubectl apply -f os-cluster.yaml
elasticsearch.kubedb.com/os-cluster created
```

In this yaml,
* `spec.version` field specifies the version of OpenSearch. Here, we are using `opensearch-2.0.1` version. You can list the KubeDB supported versions of Elasticsearch CR by running `$ kubectl get elasticsearchversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. You can get all the available `storageclass` in your cluster by running `$ kubectl get storageclass` command.
* `spec.storageType` - specifies the type of storage that will be used for OpenSearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the OpenSearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
* `spec.terminationPolicy` field is Wipeout means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/elasticsearch/concepts/elasticsearch/#specterminationpolicy) .

Once these are handled correctly and the OpenSearch object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME               READY   STATUS    RESTARTS   AGE
pod/os-cluster-0   1/1     Running   0          13m
pod/os-cluster-1   1/1     Running   0          13m
pod/os-cluster-2   1/1     Running   0          12m

NAME                        TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/os-cluster          ClusterIP   10.76.4.24   <none>        9200/TCP   13m
service/os-cluster-master   ClusterIP   None         <none>        9300/TCP   13m
service/os-cluster-pods     ClusterIP   None         <none>        9200/TCP   13m

NAME                          READY   AGE
statefulset.apps/os-cluster   3/3     13m

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/os-cluster   kubedb.com/elasticsearch   2.0.1     13m

NAME                                  VERSION            STATUS   AGE
elasticsearch.kubedb.com/os-cluster   opensearch-2.0.1   Ready    13m
```

> We have successfully deployed OpenSearch in Google Kubernetes Engine (GKE).


## Connect with OpenSearch Database

We will use [port-forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our OpenSearch Database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our OpenSearch cluster is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
os-cluster          ClusterIP   10.76.4.24   <none>        9200/TCP   17m
os-cluster-master   ClusterIP   None         <none>        9300/TCP   17m
os-cluster-pods     ClusterIP   None         <none>        9200/TCP   17m
```
Here, we are going to use `os-cluster` Service to connect with the database. Now, let’s port-forward the `os-cluster` Service to the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/os-cluster 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```
Now, our OpenSearch cluster is accessible at `localhost:9200`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `os-cluster`.

```bash
$ kubectl get secret -n demo | grep os-cluster
os-cluster-admin-cert             kubernetes.io/tls          3      19m
os-cluster-admin-cred             kubernetes.io/basic-auth   2      19m
os-cluster-ca-cert                kubernetes.io/tls          2      19m
os-cluster-client-cert            kubernetes.io/tls          3      19m
os-cluster-config                 Opaque                     3      19m
os-cluster-http-cert              kubernetes.io/tls          3      19m
os-cluster-kibanaro-cred          kubernetes.io/basic-auth   2      19m
os-cluster-kibanaserver-cred      kubernetes.io/basic-auth   2      19m
os-cluster-logstash-cred          kubernetes.io/basic-auth   2      19m
os-cluster-readall-cred           kubernetes.io/basic-auth   2      19m
os-cluster-snapshotrestore-cred   kubernetes.io/basic-auth   2      19m
os-cluster-transport-cert         kubernetes.io/tls          3      19m
```
Now, we can connect to the database with `os-cluster-admin-cred` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
Y;V)2KqoOIENO9~8
```

Now, let's check the health of our OpenSearch cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k -u 'admin:Y;V)2KqoOIENO9~8' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "os-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "discovered_master" : true,
  "discovered_cluster_manager" : true,
  "active_primary_shards" : 3,
  "active_shards" : 7,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}

```

### Insert Sample Data

In this section, we are going to create few indexes in OpenSearch. You can use `curl` for post some sample data into OpenSearch. Use the `-k` flag to disable attempts to verify self-signed certificates for testing purposes.:

```bash
$ curl -XPOST -k --user 'admin:Y;V)2KqoOIENO9~8' "https://localhost:9200/music/_doc?pretty" -H 'Content-Type: application/json' -d'
                           {
                               "Artist": "John Denver",
                               "Song": "Country Roads"
                           }
                           '
{
  "_index" : "music",
  "_id" : "aUd9RI0Bj5RvCUUenOzq",
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
$ curl -XGET -k --user 'admin:Y;V)2KqoOIENO9~8' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index                        uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   .opendistro_security         1p7keaVzSjmI9jfWSfVcfg   1   2          9            0    160.8kb         62.4kb
green  open   kubedb-system                O-alezDiSQuAdUVFVd2ixA   1   1          1           47    588.9kb        153.1kb
green  open   music                        cVDM2WLRTECrHAomFSraoA   1   1          1            0      9.3kb          4.6kb
green  open   security-auditlog-2024.01.26 G42vMzl_QKyVx6Xnt08JIQ   1   1         11            0    541.7kb        270.8kb
```
Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'admin:Y;V)2KqoOIENO9~8' "https://localhost:9200/music/_search?pretty"
{
  "took" : 104,
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
        "_id" : "aUd9RI0Bj5RvCUUenOzq",
        "_score" : 1.0,
        "_source" : {
          "Artist" : "John Denver",
          "Song" : "Country Roads"
        }
      }
    ]
  }
}

```
> We've successfully inserted some sample data to our OpenSearch database. More information about Deploy & Manage Production-Grade OpenSearch Database on Kubernetes can be found in [OpenSearch Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-opensearch-on-kubernetes/)

## Update OpenSearch Database Version

In this section, we will update our OpenSearch version from `opensearch-2.0.1` to the latest version `opensearch-2.11.1`. Let's check the current version,

```bash
$ kubectl get es -n demo os-cluster -o=jsonpath='{.spec.version}{"\n"}'
opensearch-2.0.1
```

### Create ElasticsearchOpsRequest

In order to update the version of OpenSearch cluster, we have to create a `ElasticsearchOpsRequest` CR with your desired version that is supported by KubeDB. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: "opensearch-2.11.1"
  databaseRef:
    name: os-cluster
```

Let's save this yaml configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
elasticsearchopsrequest.ops.kubedb.com/update-version created
```

In this yaml,
* `spec.databaseRef.name` specifies that we are performing operation on `os-cluster` OpenSearch database.
* `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
* `spec.updateVersion.targetVersion` specifies the expected version of the database `opensearch-2.11.1`.

### Verify the Updated OpenSearch Version

`KubeDB` operator will update the image of OpenSearch object and related `StatefulSets` and `Pods`.
Let’s wait for `ElasticsearchOpsRequest` to be Successful. Run the following command to check `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get ElasticsearchOpsRequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   4m12s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded.
Now, we are going to verify whether the OpenSearch and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get es -n demo os-cluster -o=jsonpath='{.spec.version}{"\n"}'
opensearch-2.11.1
```

> You can see from above, our OpenSearch database has been updated with the new version `opensearch-2.11.1`. So, the database update process is successfully completed.


If you want to learn more about Production-Grade OpenSearch on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=NX1Xoaq7ceJW5SnQ&amp;list=PLoiT1Gv2KR1gvo2SbHjg4KMajG8Q59ecu" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [OpenSearch on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-opensearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
