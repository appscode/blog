---
title: Deploy OpenSearch Hot-Warm-Cold Cluster in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-04-30"
weight: 12
authors:
- Dipta Roy
tags:
- amazon-eks
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- opensearch
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic and [OpenSearch](https://opensearch.org/) by AWS, but also other open source distributions like [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will deploy OpenSearch Hot-Warm-Cold Cluster in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy OpenSearch Hot-Warm-Cold Cluster
3) Verify Node Role
4) Read/Write Sample Data


# OpenSearch Hot-Warm-Cold Cluster

Hot-warm-cold architectures are common for time series data such as logging or metrics and it also has various use cases too. For example, assume OpenSearch is being used to aggregate log files from multiple systems. Logs from today are actively being indexed and this week's logs are the most heavily searched (hot). Last week's logs may be searched but not as much as the current week's logs (warm). Last month's logs may or may not be searched often, but are good to keep around just in case (cold).

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
appscode/kubedb                   	v2024.4.27   	v2024.4.27 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.30.0      	v0.30.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.4.27   	v2024.4.27 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.9       	v0.0.9     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.4.27   	v2024.4.27 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.21.0      	v0.21.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.4.27   	v2024.4.27 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.4.27   	v2024.4.27 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.4.27   	v2024.4.27 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.32.0      	v0.32.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.4.27   	v2024.4.27 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.4.27   	v0.7.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.45.0      	v0.45.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.21.0      	v0.21.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.4.2    	0.6.5      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.21.0      	v0.21.0    	KubeDB Webhook Server by AppsCode  


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.4.27 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-74c6b4b8b9-z9t2q       1/1     Running   0          2m3s
kubedb      kubedb-kubedb-ops-manager-59575b7569-8fqp2      1/1     Running   0          2m3s
kubedb      kubedb-kubedb-provisioner-5cb6654845-wrclj      1/1     Running   0          2m3s
kubedb      kubedb-kubedb-webhook-server-65dc7b4789-qprln   1/1     Running   0          2m3s
kubedb      kubedb-petset-operator-5d94b4ddb8-jn4mp         1/1     Running   0          2m3s
kubedb      kubedb-petset-webhook-server-7d9ddc9585-v7w75   2/2     Running   0          2m3s
kubedb      kubedb-sidekick-5dc87959b7-tmd8r                1/1     Running   0          2m3s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-04-30T06:48:37Z
connectors.kafka.kubedb.com                        2024-04-30T06:48:37Z
druidversions.catalog.kubedb.com                   2024-04-30T06:47:56Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-04-30T06:48:34Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-04-30T06:48:34Z
elasticsearches.kubedb.com                         2024-04-30T06:48:33Z
elasticsearchopsrequests.ops.kubedb.com            2024-04-30T06:48:34Z
elasticsearchversions.catalog.kubedb.com           2024-04-30T06:47:56Z
etcdversions.catalog.kubedb.com                    2024-04-30T06:47:56Z
ferretdbversions.catalog.kubedb.com                2024-04-30T06:47:56Z
kafkaautoscalers.autoscaling.kubedb.com            2024-04-30T06:48:37Z
kafkaconnectorversions.catalog.kubedb.com          2024-04-30T06:47:56Z
kafkaopsrequests.ops.kubedb.com                    2024-04-30T06:48:37Z
kafkas.kubedb.com                                  2024-04-30T06:48:37Z
kafkaversions.catalog.kubedb.com                   2024-04-30T06:47:56Z
mariadbarchivers.archiver.kubedb.com               2024-04-30T06:48:40Z
mariadbautoscalers.autoscaling.kubedb.com          2024-04-30T06:48:40Z
mariadbdatabases.schema.kubedb.com                 2024-04-30T06:48:40Z
mariadbopsrequests.ops.kubedb.com                  2024-04-30T06:48:40Z
mariadbs.kubedb.com                                2024-04-30T06:48:40Z
mariadbversions.catalog.kubedb.com                 2024-04-30T06:47:56Z
memcachedversions.catalog.kubedb.com               2024-04-30T06:47:56Z
mongodbarchivers.archiver.kubedb.com               2024-04-30T06:48:44Z
mongodbautoscalers.autoscaling.kubedb.com          2024-04-30T06:48:44Z
mongodbdatabases.schema.kubedb.com                 2024-04-30T06:48:44Z
mongodbopsrequests.ops.kubedb.com                  2024-04-30T06:48:44Z
mongodbs.kubedb.com                                2024-04-30T06:48:44Z
mongodbversions.catalog.kubedb.com                 2024-04-30T06:47:56Z
mssqlserverversions.catalog.kubedb.com             2024-04-30T06:47:56Z
mysqlarchivers.archiver.kubedb.com                 2024-04-30T06:48:47Z
mysqlautoscalers.autoscaling.kubedb.com            2024-04-30T06:48:47Z
mysqldatabases.schema.kubedb.com                   2024-04-30T06:48:47Z
mysqlopsrequests.ops.kubedb.com                    2024-04-30T06:48:47Z
mysqls.kubedb.com                                  2024-04-30T06:48:47Z
mysqlversions.catalog.kubedb.com                   2024-04-30T06:47:56Z
perconaxtradbversions.catalog.kubedb.com           2024-04-30T06:47:56Z
pgbouncerversions.catalog.kubedb.com               2024-04-30T06:47:56Z
pgpoolversions.catalog.kubedb.com                  2024-04-30T06:47:56Z
postgresarchivers.archiver.kubedb.com              2024-04-30T06:48:51Z
postgresautoscalers.autoscaling.kubedb.com         2024-04-30T06:48:51Z
postgresdatabases.schema.kubedb.com                2024-04-30T06:48:51Z
postgreses.kubedb.com                              2024-04-30T06:48:51Z
postgresopsrequests.ops.kubedb.com                 2024-04-30T06:48:51Z
postgresversions.catalog.kubedb.com                2024-04-30T06:47:56Z
proxysqlversions.catalog.kubedb.com                2024-04-30T06:47:56Z
publishers.postgres.kubedb.com                     2024-04-30T06:48:51Z
rabbitmqversions.catalog.kubedb.com                2024-04-30T06:47:56Z
redisautoscalers.autoscaling.kubedb.com            2024-04-30T06:48:54Z
redises.kubedb.com                                 2024-04-30T06:48:54Z
redisopsrequests.ops.kubedb.com                    2024-04-30T06:48:54Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-04-30T06:48:54Z
redissentinelopsrequests.ops.kubedb.com            2024-04-30T06:48:54Z
redissentinels.kubedb.com                          2024-04-30T06:48:54Z
redisversions.catalog.kubedb.com                   2024-04-30T06:47:57Z
singlestoreversions.catalog.kubedb.com             2024-04-30T06:47:57Z
solrversions.catalog.kubedb.com                    2024-04-30T06:47:57Z
subscribers.postgres.kubedb.com                    2024-04-30T06:48:51Z
zookeeperversions.catalog.kubedb.com               2024-04-30T06:47:57Z
```

## Deploy OpenSearch Hot-Warm-Cold Cluster

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
  name: opensearch-hwc
  namespace: demo
spec:
  version: opensearch-2.8.0
  enableSSL: true
  storageType: Durable
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "standard" 
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    dataHot:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        requests:
          cpu: 1.5
          memory: 2Gi
        limits:
          cpu: 2
          memory: 3Gi
    dataWarm:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        limits:
          cpu: 1
          memory: 2Gi
    dataCold:
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
      resources:
        requests:
          cpu: .5
          memory: 1Gi
        limits:
          cpu: .5
          memory: 1.5Gi
```

Let's save this yaml configuration into `opensearch-hwc.yaml` 
Then apply the above OpenSearch yaml,

```bash
$ kubectl apply -f opensearch-hwc.yaml
elasticsearch.kubedb.com/opensearch-hwc created
```

In this yaml,
* `spec.version` field specifies the version of OpenSearch. Here, we are using OpenSearch version `opensearch-2.8.0`. You can list the KubeDB supported versions of OpenSearch by running `$ kubectl get elasticsearchversions | grep opensearch` command. If you want to get other distributions, use `grep` command accordingly.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. You can get all the available `storageclass` in your cluster by running `$ kubectl get storageclass` command.
* `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
* `spec.storageType` - specifies the type of storage that will be used for OpenSearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the OpenSearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
* `spec.topology` - specifies the node-specific properties for the OpenSearch cluster.
  - `topology.master` - specifies the properties of master nodes.
    - `master.replicas` - specifies the number of master nodes.
    - `master.storage` - specifies the master node storage information that passed to the StatefulSet.
  - `topology.data` - specifies the properties of data nodes.
    - `data.replicas` - specifies the number of data nodes.
    - `data.storage` - specifies the data node storage information that passed to the StatefulSet.
  - `topology.ingest` - specifies the properties of ingest nodes.
    - `ingest.replicas` - specifies the number of ingest nodes.
    - `ingest.storage` - specifies the ingest node storage information that passed to the StatefulSet.

You can see the detailed yaml specifications in the [Kubernetes OpenSearch](https://kubedb.com/docs/latest/guides/elasticsearch/quickstart/overview/opensearch/) documentation.

Once these are handled correctly and the OpenSearch object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME                             READY   STATUS    RESTARTS      AGE
pod/opensearch-hwc-data-cold-0   1/1     Running   0             4m9s
pod/opensearch-hwc-data-hot-0    1/1     Running   0             4m18s
pod/opensearch-hwc-data-hot-1    1/1     Running   0             4m2s
pod/opensearch-hwc-data-hot-2    1/1     Running   0             4m2s
pod/opensearch-hwc-data-warm-0   1/1     Running   0             4m18s
pod/opensearch-hwc-data-warm-1   1/1     Running   0             4m2s
pod/opensearch-hwc-ingest-0      1/1     Running   0             4m18s
pod/opensearch-hwc-ingest-1      1/1     Running   0             4m2s
pod/opensearch-hwc-master-0      1/1     Running   0             4m18s
pod/opensearch-hwc-master-1      1/1     Running   0             4m2s

NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/opensearch-hwc          ClusterIP   10.96.144.238   <none>        9200/TCP   4m18s
service/opensearch-hwc-master   ClusterIP   None            <none>        9300/TCP   4m18s
service/opensearch-hwc-pods     ClusterIP   None            <none>        9200/TCP   4m18s

NAME                                        READY   AGE
statefulset.apps/opensearch-hwc-data-cold   1/1     4m9s
statefulset.apps/opensearch-hwc-data-hot    3/3     4m18s
statefulset.apps/opensearch-hwc-data-warm   2/2     4m18s
statefulset.apps/opensearch-hwc-ingest      2/2     4m18s
statefulset.apps/opensearch-hwc-master      2/2     4m18s

NAME                                                TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/opensearch-hwc   kubedb.com/elasticsearch   2.8.0     4m9s

NAME                                      VERSION            STATUS   AGE
elasticsearch.kubedb.com/opensearch-hwc   opensearch-2.8.0   Ready    4m18s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get elasticsearch -n demo opensearch-hwc
NAME             VERSION            STATUS   AGE
opensearch-hwc   opensearch-2.8.0   Ready    5m11s
```
> We have successfully deployed OpenSearch in Amazon EKS. Now we can exec into the container to use the database.


## Connect with OpenSearch Database

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our OpenSearch database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our OpenSearch database is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
opensearch-hwc          ClusterIP   10.96.144.238   <none>        9200/TCP   6m32s
opensearch-hwc-master   ClusterIP   None            <none>        9300/TCP   6m32s
opensearch-hwc-pods     ClusterIP   None            <none>        9200/TCP   6m32s
```
Here, we are going to use `opensearch-hwc` Service to connect with the database. Now, let’s port-forward the `opensearch-hwc` Service to the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/opensearch-hwc 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```
Now, our OpenSearch cluster is accessible at `localhost:9200`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `opensearch-hwc`.

```bash
$ kubectl get secret -n demo | grep opensearch-hwc
opensearch-hwc-admin-cert             kubernetes.io/tls          3      7m24s
opensearch-hwc-admin-cred             kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-ca-cert                kubernetes.io/tls          2      7m24s
opensearch-hwc-client-cert            kubernetes.io/tls          3      7m24s
opensearch-hwc-config                 Opaque                     3      7m24s
opensearch-hwc-http-cert              kubernetes.io/tls          3      7m24s
opensearch-hwc-kibanaro-cred          kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-kibanaserver-cred      kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-logstash-cred          kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-readall-cred           kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-snapshotrestore-cred   kubernetes.io/basic-auth   2      7m24s
opensearch-hwc-transport-cert         kubernetes.io/tls          3      7m24s
```
Now, we can connect to the database with `opensearch-hwc-admin-cred` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo opensearch-hwc-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo opensearch-hwc-admin-cred -o jsonpath='{.data.password}' | base64 -d
INl7kmwCy~;cLNVy
```

Now, let's check the health of our OpenSearch cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k -u 'admin:INl7kmwCy~;cLNVy' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "opensearch-hwc",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 10,
  "number_of_data_nodes" : 6,
  "discovered_master" : true,
  "discovered_cluster_manager" : true,
  "active_primary_shards" : 4,
  "active_shards" : 13,
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

### Verify Node Role

As we have assigned a dedicated role to each type of node, let's verify them by following command,

```bash
$ curl -XGET -k -u 'admin:INl7kmwCy~;cLNVy' "https://localhost:9200/_cat/nodes?v"
ip          heap.percent ram.percent cpu load_1m load_5m load_15m node.role node.roles                   cluster_manager name
10.244.0.7            43          68   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-hot-0
10.244.0.10           55          67   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-warm-1
10.244.0.12           51          66   0    1.04    6.23     5.90 m         cluster_manager              *               opensearch-hwc-master-1
10.244.0.19           46          68   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-hot-2
10.244.0.9            52          68   0    1.04    6.23     5.90 ir        ingest,remote_cluster_client -               opensearch-hwc-ingest-1
10.244.0.6            43          69   0    1.04    6.23     5.90 ir        ingest,remote_cluster_client -               opensearch-hwc-ingest-0
10.244.0.17           14          74   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-cold-0
10.244.0.8            46          66   0    1.04    6.23     5.90 m         cluster_manager              -               opensearch-hwc-master-0
10.244.0.18           12          65   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-hot-1
10.244.0.11           14          75   0    1.04    6.23     5.90 d         data                         -               opensearch-hwc-data-warm-0
```

- `node.role` field specifies the dedicated role that we have assigned for each type of node. Where `d` refers to the data node, `ir` refers to the ingest node, `m` refers to the master node.
- `master` field specifies the active master node. Here, we can see a `*` in the `opensearch-hwc-master-1` which shows that it is the active master node now.


### Insert Sample Data

In this section, we are going to create few indexes in OpenSearch. You can use `curl` for post some sample data into OpenSearch. Use the `-k` flag to disable attempts to verify self-signed certificates for testing purposes.

```bash
$ curl -XPOST -k --user 'admin:INl7kmwCy~;cLNVy' "https://localhost:9200/music/_doc?pretty" -H 'Content-Type: application/json' -d'
                           {
                               "Artist": "Bobby Bare",
                               "Song": "Five Hundred Miles"
                           }
                           '
{
  "_index" : "music",
  "_id" : "ODcaHo4Bvi0hOBvmUCCs",
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
$ curl -XGET -k --user 'admin:INl7kmwCy~;cLNVy' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index                        uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   .opendistro_security         w0PT_AHsRw2HS3bhWD_uYw   1   5         10            0    418.7kb         75.4kb
green  open   .opensearch-observability    3SOLAZuJSUucTV3gR4W7Zg   1   2          0            0       624b           208b
green  open   kubedb-system                b59odv8ZTmmK_VzOJEd0Lg   1   1          1          133      1.2mb        668.8kb
green  open   music                        HbCoMR1pTXiRAw7Ane_8ZA   1   1          1            0      9.2kb          4.6kb
green  open   security-auditlog-2024.03.08 RP2oa_G-SWepIK71oan6GQ   1   1         12            0    421.8kb        210.9kb
```
Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'admin:INl7kmwCy~;cLNVy' "https://localhost:9200/music/_search?pretty"
{
  "took" : 93,
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
        "_id" : "ODcaHo4Bvi0hOBvmUCCs",
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
> We've successfully inserted some sample data to our OpenSearch database. More information about Deploy & Manage Production-Grade OpenSearch Database on Kubernetes can be found in [OpenSearch Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-opensearch-on-kubernetes/)

If you want to learn more about Production-Grade OpenSearch on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=kSHEAxPiq_pfSsj4&amp;list=PLoiT1Gv2KR1gvo2SbHjg4KMajG8Q59ecu" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [OpenSearch on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-opensearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
