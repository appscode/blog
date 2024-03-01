---
title: Update Version of Elasticsearch Database in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-02-28"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- elastic
- elasticsearch
- kubedb
- kubernetes
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic and [OpenSearch](https://opensearch.org/) by AWS, but also other open source distributions like [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will update version of Elasticsearch database in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1. Install KubeDB
2. Deploy Elasticsearch Cluster
3. Insert Sample Data
4. Update Elasticsearch Database Version

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
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-5b9fbf7468-m7jhd       1/1     Running   0          2m51s
kubedb      kubedb-kubedb-ops-manager-74d65767c6-wxtlr      1/1     Running   0          2m51s
kubedb      kubedb-kubedb-provisioner-7b97fb9fdd-t4fxt      1/1     Running   0          2m51s
kubedb      kubedb-kubedb-webhook-server-86dd6bf6cb-spn5k   1/1     Running   0          2m51s
kubedb      kubedb-sidekick-5dc87959b7-ftrwd                1/1     Running   0          2m51s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-02-28T11:51:49Z
connectors.kafka.kubedb.com                        2024-02-28T11:51:49Z
druidversions.catalog.kubedb.com                   2024-02-28T11:51:06Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-02-28T11:51:45Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-02-28T11:51:45Z
elasticsearches.kubedb.com                         2024-02-28T11:51:45Z
elasticsearchopsrequests.ops.kubedb.com            2024-02-28T11:51:45Z
elasticsearchversions.catalog.kubedb.com           2024-02-28T11:51:06Z
etcdversions.catalog.kubedb.com                    2024-02-28T11:51:06Z
ferretdbversions.catalog.kubedb.com                2024-02-28T11:51:06Z
kafkaconnectorversions.catalog.kubedb.com          2024-02-28T11:51:06Z
kafkaopsrequests.ops.kubedb.com                    2024-02-28T11:51:49Z
kafkas.kubedb.com                                  2024-02-28T11:51:49Z
kafkaversions.catalog.kubedb.com                   2024-02-28T11:51:06Z
mariadbautoscalers.autoscaling.kubedb.com          2024-02-28T11:51:52Z
mariadbdatabases.schema.kubedb.com                 2024-02-28T11:51:52Z
mariadbopsrequests.ops.kubedb.com                  2024-02-28T11:51:52Z
mariadbs.kubedb.com                                2024-02-28T11:51:52Z
mariadbversions.catalog.kubedb.com                 2024-02-28T11:51:06Z
memcachedversions.catalog.kubedb.com               2024-02-28T11:51:06Z
mongodbarchivers.archiver.kubedb.com               2024-02-28T11:51:55Z
mongodbautoscalers.autoscaling.kubedb.com          2024-02-28T11:51:55Z
mongodbdatabases.schema.kubedb.com                 2024-02-28T11:51:56Z
mongodbopsrequests.ops.kubedb.com                  2024-02-28T11:51:55Z
mongodbs.kubedb.com                                2024-02-28T11:51:55Z
mongodbversions.catalog.kubedb.com                 2024-02-28T11:51:06Z
mysqlarchivers.archiver.kubedb.com                 2024-02-28T11:51:59Z
mysqlautoscalers.autoscaling.kubedb.com            2024-02-28T11:51:59Z
mysqldatabases.schema.kubedb.com                   2024-02-28T11:51:59Z
mysqlopsrequests.ops.kubedb.com                    2024-02-28T11:51:59Z
mysqls.kubedb.com                                  2024-02-28T11:51:59Z
mysqlversions.catalog.kubedb.com                   2024-02-28T11:51:06Z
perconaxtradbversions.catalog.kubedb.com           2024-02-28T11:51:06Z
pgbouncerversions.catalog.kubedb.com               2024-02-28T11:51:06Z
pgpoolversions.catalog.kubedb.com                  2024-02-28T11:51:06Z
postgresarchivers.archiver.kubedb.com              2024-02-28T11:52:02Z
postgresautoscalers.autoscaling.kubedb.com         2024-02-28T11:52:02Z
postgresdatabases.schema.kubedb.com                2024-02-28T11:52:03Z
postgreses.kubedb.com                              2024-02-28T11:52:02Z
postgresopsrequests.ops.kubedb.com                 2024-02-28T11:52:02Z
postgresversions.catalog.kubedb.com                2024-02-28T11:51:06Z
proxysqlversions.catalog.kubedb.com                2024-02-28T11:51:06Z
publishers.postgres.kubedb.com                     2024-02-28T11:52:03Z
rabbitmqversions.catalog.kubedb.com                2024-02-28T11:51:06Z
redisautoscalers.autoscaling.kubedb.com            2024-02-28T11:52:06Z
redises.kubedb.com                                 2024-02-28T11:52:06Z
redisopsrequests.ops.kubedb.com                    2024-02-28T11:52:06Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-02-28T11:52:06Z
redissentinelopsrequests.ops.kubedb.com            2024-02-28T11:52:06Z
redissentinels.kubedb.com                          2024-02-28T11:52:06Z
redisversions.catalog.kubedb.com                   2024-02-28T11:51:06Z
singlestoreversions.catalog.kubedb.com             2024-02-28T11:51:06Z
solrversions.catalog.kubedb.com                    2024-02-28T11:51:06Z
subscribers.postgres.kubedb.com                    2024-02-28T11:52:03Z
zookeeperversions.catalog.kubedb.com               2024-02-28T11:51:06Z
```

## Deploy Elasticsearch Cluster

Now, We are going to use the KubeDB-provided Custom Resource object `Elasticsearch` for deployment. First, let’s create a Namespace in which we will deploy the cluster.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
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

Let's save this yaml configuration into `es-cluster.yaml`
Then apply the above Elasticsearch yaml,

```bash
$ kubectl apply -f es-cluster.yaml
elasticsearch.kubedb.com/es-cluster created
```

In this yaml,

- `spec.version` field specifies the version of Elasticsearch. Here, we are using `xpack-8.2.3` version. You can list the KubeDB supported versions of Elasticsearch CR by running `$ kubectl get elasticsearchversions` command.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. You can get all the available `storageclass` in your cluster by running `$ kubectl get storageclass` command.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.terminationPolicy` field is Wipeout means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about [Termination Policy](https://kubedb.com/docs/latest/guides/elasticsearch/concepts/elasticsearch/#specterminationpolicy) .

Once these are handled correctly and the Elasticsearch object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME               READY   STATUS    RESTARTS   AGE
pod/es-cluster-0   1/1     Running   0          2m14s
pod/es-cluster-1   1/1     Running   0          2m6s
pod/es-cluster-2   1/1     Running   0          119s

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/es-cluster          ClusterIP   10.96.183.145   <none>        9200/TCP   2m19s
service/es-cluster-master   ClusterIP   None            <none>        9300/TCP   2m19s
service/es-cluster-pods     ClusterIP   None            <none>        9200/TCP   2m19s

NAME                          READY   AGE
statefulset.apps/es-cluster   3/3     2m14s

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-cluster   kubedb.com/elasticsearch   8.2.3     2m14s

NAME                                  VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-cluster   xpack-8.2.3   Ready    2m19s
```

> We have successfully deployed Elasticsearch in EKS.

## Connect with Elasticsearch Database

We will use [port-forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our Elasticsearch Database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Elasticsearch cluster is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
es-cluster          ClusterIP   10.96.183.145   <none>        9200/TCP   3m22s
es-cluster-master   ClusterIP   None            <none>        9300/TCP   3m22s
es-cluster-pods     ClusterIP   None            <none>        9200/TCP   3m22s
```

Here, we are going to use `es-cluster` Service to connect with the database. Now, let’s port-forward the `es-cluster` Service to the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/es-cluster 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `es-cluster`.

```bash
$ kubectl get secret -n demo | grep es-cluster
es-cluster-apm-system-cred               kubernetes.io/basic-auth   2      4m44s
es-cluster-beats-system-cred             kubernetes.io/basic-auth   2      4m44s
es-cluster-ca-cert                       kubernetes.io/tls          2      4m49s
es-cluster-client-cert                   kubernetes.io/tls          3      4m48s
es-cluster-config                        Opaque                     1      4m48s
es-cluster-elastic-cred                  kubernetes.io/basic-auth   2      4m48s
es-cluster-http-cert                     kubernetes.io/tls          3      4m48s
es-cluster-kibana-system-cred            kubernetes.io/basic-auth   2      4m44s
es-cluster-logstash-system-cred          kubernetes.io/basic-auth   2      4m44s
es-cluster-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      4m44s
es-cluster-transport-cert                kubernetes.io/tls          3      4m49s
```

Now, we can connect to the database with `es-cluster-elastic-cred` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.password}' | base64 -d
INl7kmwCy~;cLNVy
```

Now, let's check the health of our Elasticsearch cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k -u 'elastic:INl7kmwCy~;cLNVy' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 3,
  "active_shards" : 6,
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

In this section, we are going to create few indexes in Elasticsearch. You can use `curl` for post some sample data into Elasticsearch. Use the `-k` flag to disable attempts to verify self-signed certificates for testing purposes.:

```bash
$ curl -XPOST -k --user 'elastic:INl7kmwCy~;cLNVy' "https://localhost:9200/music/_doc?pretty" -H 'Content-Type: application/json' -d'
                                    {
                                        "Artist": "Bobby Bare",
                                        "Song": "Five Hundred Miles"
                                    }
                                    '
{
  "_index" : "music",
  "_id" : "1hXn9I0B-9Bs8zqEkq_C",
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
$ curl -XGET -k --user 'elastic:INl7kmwCy~;cLNVy' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index         uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   kubedb-system mKYXqboWS-SS2GUJr50rFw   1   1          1            3    401.9kb        192.4kb
green  open   music         JyDapu8KRDSNiWwe2VoZNw   1   1          1            0      9.4kb          4.7kb
```

Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'elastic:INl7kmwCy~;cLNVy' "https://localhost:9200/music/_search?pretty"
{
  "took" : 23,
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
        "_id" : "1hXn9I0B-9Bs8zqEkq_C",
        "_score" : 1.0,
        "_source" : {
          "Artist" : "Bobby Bare",
          "Song" : "Five Hundred Miles"
        }
      }
    ]
  }
}

```

> We've successfully inserted some sample data to our Elasticsearch database. More information about Deploy & Manage Production-Grade Elasticsearch Database on Kubernetes can be found in [Elasticsearch Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

## Update Elasticsearch Database Version

In this section, we will update our Elasticsearch version from `xpack-8.2.3` to the latest version `xpack-8.11.1`. Let's check the current version,

```bash
$ kubectl get es -n demo es-cluster -o=jsonpath='{.spec.version}{"\n"}'
xpack-8.2.3
```

### Create ElasticsearchOpsRequest

In order to update the version of Elasticsearch cluster, we have to create a `ElasticsearchOpsRequest` CR with your desired version that is supported by KubeDB. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: "xpack-8.11.1"
  databaseRef:
    name: es-cluster
```

Let's save this yaml configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
elasticsearchopsrequest.ops.kubedb.com/update-version created
```

In this yaml,

- `spec.databaseRef.name` specifies that we are performing operation on `es-cluster` Elasticsearch database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `xpack-8.11.1`.

### Verify the Updated Elasticsearch Version

`KubeDB` operator will update the image of Elasticsearch object and related `StatefulSets` and `Pods`.
Let’s wait for `ElasticsearchOpsRequest` to be Successful. Run the following command to check `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get ElasticsearchOpsRequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   5m37s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded.
Now, we are going to verify whether the Elasticsearch and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get es -n demo es-cluster -o=jsonpath='{.spec.version}{"\n"}'
xpack-8.11.1
```

> You can see from above, our Elasticsearch database has been updated with the new version `xpack-8.11.1`. So, the database update process is successfully completed.

We have made a tutorial on Provision Elasticsearch Multi-node Combined cluster and Topology Cluster using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/O42Pvf2NuCo?si=X41X0YRSIA1P6_WP" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Elasticsearch on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
