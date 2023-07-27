---
title: Deploy OpenSearch with OpenSearch-Dashboards on Google Kubernetes Engine (GKE) using KubeDB
date: "2023-07-28"
weight: 12
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- elasticsearch
- gcp
- gcs
- gke
- kubedb
- kubernetes
- opensearch
- opensearch-dashboard
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Elasticsearch, Kafka, MySQL, MongoDB, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic and [OpenSearch](https://opensearch.org/) by AWS, but also other open source distributions like [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will Deploy OpenSearch with OpenSearch-Dashboards on Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy OpenSearch Topology Cluster
3) Deploy OpenSearch-Dashboard
4) Read/Write Data through Dashboard


### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID, we can run the following command:

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
appscode/kubedb                   	v2023.06.19  	v2023.06.19	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.19.0      	v0.19.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.06.19  	v2023.06.19	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.06.19  	v2023.06.19	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.10.0      	v0.10.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.06.19  	v2023.06.19	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.06.19  	v2023.06.19	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.06.19  	v2023.06.19	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.21.0      	v0.21.3    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.06.19  	v2023.06.19	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.34.0      	v0.34.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.10.0      	v0.10.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.33-rc.2	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.10.0      	v0.10.0    	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.06.19 \
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
NAMESPACE   NAME                                            READY   STATUS    RESTARTS      AGE
kubedb      kubedb-kubedb-autoscaler-7d6d794847-z5gdt       1/1     Running   0             2m26s
kubedb      kubedb-kubedb-dashboard-6f866749cf-fmg8n        1/1     Running   0             2m26s
kubedb      kubedb-kubedb-ops-manager-b679d5df7-zqx57       1/1     Running   1             2m26s
kubedb      kubedb-kubedb-provisioner-bc8d7d45c-sv2lx       1/1     Running   0             2m26s
kubedb      kubedb-kubedb-schema-manager-8579f67b58-5s2ld   1/1     Running   0             2m25s
kubedb      kubedb-kubedb-webhook-server-7b6957848b-7cbl6   1/1     Running   0             2m26s

```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-07-27T05:55:58Z
elasticsearchdashboards.dashboard.kubedb.com      2023-07-27T05:55:56Z
elasticsearches.kubedb.com                        2023-07-27T05:55:57Z
elasticsearchopsrequests.ops.kubedb.com           2023-07-27T05:56:15Z
elasticsearchversions.catalog.kubedb.com          2023-07-27T05:53:20Z
etcds.kubedb.com                                  2023-07-27T05:56:10Z
etcdversions.catalog.kubedb.com                   2023-07-27T05:53:20Z
kafkas.kubedb.com                                 2023-07-27T05:56:33Z
kafkaversions.catalog.kubedb.com                  2023-07-27T05:53:21Z
mariadbautoscalers.autoscaling.kubedb.com         2023-07-27T05:55:58Z
mariadbdatabases.schema.kubedb.com                2023-07-27T05:56:15Z
mariadbopsrequests.ops.kubedb.com                 2023-07-27T05:56:46Z
mariadbs.kubedb.com                               2023-07-27T05:56:15Z
mariadbversions.catalog.kubedb.com                2023-07-27T05:53:21Z
memcacheds.kubedb.com                             2023-07-27T05:56:15Z
memcachedversions.catalog.kubedb.com              2023-07-27T05:53:21Z
mongodbautoscalers.autoscaling.kubedb.com         2023-07-27T05:55:58Z
mongodbdatabases.schema.kubedb.com                2023-07-27T05:56:02Z
mongodbopsrequests.ops.kubedb.com                 2023-07-27T05:56:20Z
mongodbs.kubedb.com                               2023-07-27T05:56:04Z
mongodbversions.catalog.kubedb.com                2023-07-27T05:53:22Z
mysqlautoscalers.autoscaling.kubedb.com           2023-07-27T05:55:59Z
mysqldatabases.schema.kubedb.com                  2023-07-27T05:55:58Z
mysqlopsrequests.ops.kubedb.com                   2023-07-27T05:56:42Z
mysqls.kubedb.com                                 2023-07-27T05:55:59Z
mysqlversions.catalog.kubedb.com                  2023-07-27T05:53:22Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-07-27T05:55:59Z
perconaxtradbopsrequests.ops.kubedb.com           2023-07-27T05:57:01Z
perconaxtradbs.kubedb.com                         2023-07-27T05:56:29Z
perconaxtradbversions.catalog.kubedb.com          2023-07-27T05:53:23Z
pgbouncers.kubedb.com                             2023-07-27T05:56:29Z
pgbouncerversions.catalog.kubedb.com              2023-07-27T05:53:23Z
postgresautoscalers.autoscaling.kubedb.com        2023-07-27T05:56:00Z
postgresdatabases.schema.kubedb.com               2023-07-27T05:56:09Z
postgreses.kubedb.com                             2023-07-27T05:56:15Z
postgresopsrequests.ops.kubedb.com                2023-07-27T05:56:54Z
postgresversions.catalog.kubedb.com               2023-07-27T05:53:23Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-07-27T05:56:00Z
proxysqlopsrequests.ops.kubedb.com                2023-07-27T05:56:57Z
proxysqls.kubedb.com                              2023-07-27T05:56:31Z
proxysqlversions.catalog.kubedb.com               2023-07-27T05:53:23Z
publishers.postgres.kubedb.com                    2023-07-27T05:57:13Z
redisautoscalers.autoscaling.kubedb.com           2023-07-27T05:56:01Z
redises.kubedb.com                                2023-07-27T05:56:32Z
redisopsrequests.ops.kubedb.com                   2023-07-27T05:56:49Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-07-27T05:56:02Z
redissentinelopsrequests.ops.kubedb.com           2023-07-27T05:57:04Z
redissentinels.kubedb.com                         2023-07-27T05:56:32Z
redisversions.catalog.kubedb.com                  2023-07-27T05:53:24Z
subscribers.postgres.kubedb.com                   2023-07-27T05:57:16Z
```

## Deploy OpenSearch Topology Cluster

We are going to use the KubeDB-provided Custom Resource object OpenSearch for deployment. The object will be deployed in demo namespace. So, let’s create the namespace first.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of OpenSearch we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: os-cluster
  namespace: demo
spec:
  enableSSL: true 
  version: opensearch-2.8.0 
  storageType: Durable
  topology:
    master:
      replicas: 2
      resources:
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 2
      resources:
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      resources:
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  terminationPolicy: WipeOut
```

Here,

* `spec.version` - is the name of the ElasticsearchVersion CR. Here, we are using OpenSearch version `opensearch-2.8.0` of OpenSearch distribution.
* `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
* `spec.storageType` - specifies the type of storage that will be used for OpenSearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the OpenSearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
* `spec.topology` - specifies the node-specific properties for the OpenSearch cluster.
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/elasticsearch/concepts/elasticsearch/#specterminationpolicy).

Let's deploy the above yaml by the following command:

```bash
$ kubectl apply -f os-cluster.yaml
elasticsearch.kubedb.com/os-cluster created
```
However, KubeDB also provides dedicated node support for other node roles like `data_hot`, `data_warm`, `data_cold`, `data_frozen`, `transform`, `coordinating`, `data_content` and `ml` for [Topology clustering](https://kubedb.com/docs/latest/guides/elasticsearch/clustering/topology-cluster/hot-warm-cold-cluster/).

Once these are handled correctly and the OpenSearch object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME                      READY   STATUS    RESTARTS   AGE
pod/os-cluster-data-0     1/1     Running   0          2m
pod/os-cluster-data-1     1/1     Running   0          2m
pod/os-cluster-ingest-0   1/1     Running   0          2m
pod/os-cluster-ingest-1   1/1     Running   0          2m
pod/os-cluster-master-0   1/1     Running   0          2m
pod/os-cluster-master-1   1/1     Running   0          2m

NAME                        TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/os-cluster          ClusterIP   10.8.12.24   <none>        9200/TCP   2m
service/os-cluster-master   ClusterIP   None         <none>        9300/TCP   2m
service/os-cluster-pods     ClusterIP   None         <none>        9200/TCP   2m

NAME                                 READY   AGE
statefulset.apps/os-cluster-data     2/2     2m
statefulset.apps/os-cluster-ingest   2/2     2m
statefulset.apps/os-cluster-master   2/2     2m

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/os-cluster   kubedb.com/elasticsearch   2.8.0     2m

NAME                                  VERSION            STATUS   AGE
elasticsearch.kubedb.com/os-cluster   opensearch-2.8.0   Ready    2m
```
> We have successfully deployed OpenSearch cluster in GKE. 

## Deploy ElasticsearchDashboard

```yaml
apiVersion: dashboard.kubedb.com/v1alpha1
kind: ElasticsearchDashboard
metadata:
  name: os-cluster-dashboard
  namespace: demo
spec:
  enableSSL: true
  databaseRef:
    name: os-cluster
  terminationPolicy: WipeOut
```
> Note: OpenSearch Database and OpenSearch dashboard should have to be deployed in the same namespace. In this tutorial, we use `demo` namespace for both cases.

- `spec.enableSSL` specifies whether the HTTP layer is secured with certificates or not.
- `spec.databaseRef.name` refers to the OpenSearch database name.
- `spec.terminationPolicy` refers to the strategy to follow during dashboard deletion. `Wipeout` means that the database will be deleted without restrictions. It can also be `DoNotTerminate` which will cause a restriction to delete the dashboard. Learn More about these [Termination Policy](https://kubedb.com/docs/latest/guides/elasticsearch/concepts/elasticsearch/#specterminationpolicy).

Let's deploy the above yaml by the following command:

```bash
$ kubectl apply -f os-cluster-dashboard.yaml
elasticsearchdashboard.dashboard.kubedb.com/os-cluster-dashboard created
```

KubeDB will create the necessary resources to deploy the OpenSearch dashboard according to the above specification. Let’s wait until the dashboard to be ready to use,

```bash
$ watch kubectl get elasticsearchdashboard -n demo
NAME                   TYPE                            DATABASE     STATUS   AGE
os-cluster-dashboard   dashboard.kubedb.com/v1alpha1   os-cluster   Ready    2m50s
```
Here, OpenSearch Dashboard is in `Ready` state. 


## Connect with OpenSearch Dashboard

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our OpenSearch database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our OpenSearch database is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                   TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
os-cluster             ClusterIP   10.8.12.24   <none>        9200/TCP   7m
os-cluster-dashboard   ClusterIP   10.8.9.220   <none>        5601/TCP   3m23s
os-cluster-master      ClusterIP   None         <none>        9300/TCP   7m
os-cluster-pods        ClusterIP   None         <none>        9200/TCP   7m
```
Here, we are going to use `os-cluster-dashboard` Service to connect with the database. Now, let’s port-forward the `os-cluster` Service to the port `5601` to local machine:

```bash
$ kubectl port-forward -n demo service/os-cluster-dashboard 5601
Forwarding from 127.0.0.1:5601 -> 5601
Forwarding from [::1]:5601 -> 5601
```
Now, our OpenSearch cluster dashboard is accessible at `https://localhost:5601`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `os-cluster`.

```bash
$ kubectl get secret -n demo | grep os-cluster
os-cluster-admin-cert              kubernetes.io/tls          3      9m
os-cluster-admin-cred              kubernetes.io/basic-auth   2      9m
os-cluster-ca-cert                 kubernetes.io/tls          2      9m
os-cluster-client-cert             kubernetes.io/tls          3      9m
os-cluster-config                  Opaque                     3      9m
os-cluster-dashboard-ca-cert       kubernetes.io/tls          2      5m43s
os-cluster-dashboard-config        Opaque                     2      5m43s
os-cluster-dashboard-server-cert   kubernetes.io/tls          3      5m43s
os-cluster-http-cert               kubernetes.io/tls          3      9m
os-cluster-kibanaro-cred           kubernetes.io/basic-auth   2      9m
os-cluster-kibanaserver-cred       kubernetes.io/basic-auth   2      9m
os-cluster-logstash-cred           kubernetes.io/basic-auth   2      9m
os-cluster-readall-cred            kubernetes.io/basic-auth   2      9m
os-cluster-snapshotrestore-cred    kubernetes.io/basic-auth   2      9m
os-cluster-transport-cert          kubernetes.io/tls          3      9m
```
Now, we can connect to the database with `os-cluster-admin-cred` which contains the admin credentials to connect with the database.

### Accessing Database Through Dashboard

To access the database through Dashboard, we have to get the credentials. We can do that by following command,

```bash
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
RwEOgPwu.n3T~WWS
```

Now, let's go to `https://localhost:5601` from our browser and login by using those credentials.

![Login Page](LoginPage.png)

After login successfully, we will see OpenSearch Dashboard UI. Now, We are going to `Dev tools` for running some queries into our OpenSearch database.

![Dashboard UI](DashboardUI.png)

Here, in `Dev tools` we will use `Console` section for running some queries. Let's run `GET /` query to check node informations.
```bash
GET /
```
![Get Query](GetQuery.png)

Now, we are going to insert some sample data to our OpenSearch cluster index `music/_doc/1` by using `PUT` query.
```bash
PUT music/_doc/1
{
    "Playlist": {
      "Song": "Take Me Home Country Roads",
      "Artist": "John Denver",
      "Album": "Poems, Prayers & Promises"
    }
}
```
![Sample Data](SampleData.png)

Let's check that sample data in the index `music/_doc/1` by using `GET` query.
```bash
GET music/_doc/1
```
![Get Data](GetData.png)

Now, we are going to update sample data in the index `music/_doc/1` by using `POST` query.
```bash
POST music/_doc/1
{
    "Playlist": {
      "Song": "Take Me Home Country Roads",
      "Artist": "John Denver",
      "Album": "Poems, Prayers & Promises",
      "Released": "April 6, 1971"
    }
}
```
![Post Data](PostData.png)

Let's verify the index `music/_doc/1` again to see whether the data is updated or not.
```bash
GET music/_doc/1
```
![Get Updated Data](GetUpdatedData.png)


We have made an in depth tutorial on OpenSearch OpsRequests - Day 2 Lifecycle Management for OpenSearch Cluster Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/gSoWaVV4iQo" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [OpenSearch in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
