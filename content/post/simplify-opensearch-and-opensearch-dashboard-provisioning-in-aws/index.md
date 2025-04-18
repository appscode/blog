---
title: Simplify OpenSearch and OpenSearch-Dashboards Provisioning on Amazon EKS using KubeDB
date: "2023-04-07"
weight: 12
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- database
- dbaas
- eks
- elasticsearch
- elasticstack
- kubedb
- kubernetes
- s3
- xpack
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are Elasticsearch, Kafka, MySQL, MongoDB, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). KubeDB provides support not only for the official [Elasticsearch](https://www.elastic.co/) by Elastic and [OpenSearch](https://opensearch.org/) by AWS, but also other open source distributions like [SearchGuard](https://search-guard.com/) and [OpenDistro](https://opendistro.github.io/for-elasticsearch/). **KubeDB provides all of these distribution's support under the Elasticsearch CR of KubeDB**.
In this tutorial we will Simplify OpenSearch and OpenSearch-Dashboards Provisioning on Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

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
appscode/kubedb                   	v2023.02.28  	v2023.02.28	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.17.0      	v0.17.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.02.28  	v2023.02.28	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.02.28  	v2023.02.28	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.8.0       	v0.8.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.02.28  	v2023.02.28	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.02.28  	v2023.02.28	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.02.28  	v2023.02.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.19.0      	v0.19.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.02.28  	v2023.02.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.32.0      	v0.32.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.8.0       	v0.8.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.8.0       	v0.8.0     	KubeDB Webhook Server by AppsCode   

# Install KubeDB Enterprise operator chart
helm install kubedb appscode/kubedb \
  --version v2023.02.28 \
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
kubedb      kubedb-kubedb-autoscaler-584955b566-9v2m5      1/1     Running   0          82s
kubedb      kubedb-kubedb-dashboard-6c488fc485-8wcdv       1/1     Running   0          82s
kubedb      kubedb-kubedb-ops-manager-56f468db6d-znlw9     1/1     Running   0          82s
kubedb      kubedb-kubedb-provisioner-6cdcd7ffc6-pfcph     1/1     Running   0          82s
kubedb      kubedb-kubedb-schema-manager-fb9dbd766-zg6js   1/1     Running   0          82s
kubedb      kubedb-kubedb-webhook-server-f779b9957-pxlsd   1/1     Running   0          82s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-04-06T07:06:11Z
elasticsearchdashboards.dashboard.kubedb.com      2023-04-06T07:06:11Z
elasticsearches.kubedb.com                        2023-04-06T07:06:11Z
elasticsearchopsrequests.ops.kubedb.com           2023-04-06T07:06:15Z
elasticsearchversions.catalog.kubedb.com          2023-04-06T07:03:22Z
etcds.kubedb.com                                  2023-04-06T07:06:15Z
etcdversions.catalog.kubedb.com                   2023-04-06T07:03:22Z
kafkas.kubedb.com                                 2023-04-06T07:06:28Z
kafkaversions.catalog.kubedb.com                  2023-04-06T07:03:23Z
mariadbautoscalers.autoscaling.kubedb.com         2023-04-06T07:06:11Z
mariadbdatabases.schema.kubedb.com                2023-04-06T07:06:17Z
mariadbopsrequests.ops.kubedb.com                 2023-04-06T07:06:37Z
mariadbs.kubedb.com                               2023-04-06T07:06:16Z
mariadbversions.catalog.kubedb.com                2023-04-06T07:03:23Z
memcacheds.kubedb.com                             2023-04-06T07:06:16Z
memcachedversions.catalog.kubedb.com              2023-04-06T07:03:23Z
mongodbautoscalers.autoscaling.kubedb.com         2023-04-06T07:06:11Z
mongodbdatabases.schema.kubedb.com                2023-04-06T07:06:14Z
mongodbopsrequests.ops.kubedb.com                 2023-04-06T07:06:19Z
mongodbs.kubedb.com                               2023-04-06T07:06:15Z
mongodbversions.catalog.kubedb.com                2023-04-06T07:03:24Z
mysqlautoscalers.autoscaling.kubedb.com           2023-04-06T07:06:11Z
mysqldatabases.schema.kubedb.com                  2023-04-06T07:06:13Z
mysqlopsrequests.ops.kubedb.com                   2023-04-06T07:06:33Z
mysqls.kubedb.com                                 2023-04-06T07:06:13Z
mysqlversions.catalog.kubedb.com                  2023-04-06T07:03:24Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-04-06T07:06:11Z
perconaxtradbopsrequests.ops.kubedb.com           2023-04-06T07:06:53Z
perconaxtradbs.kubedb.com                         2023-04-06T07:06:25Z
perconaxtradbversions.catalog.kubedb.com          2023-04-06T07:03:24Z
pgbouncers.kubedb.com                             2023-04-06T07:06:26Z
pgbouncerversions.catalog.kubedb.com              2023-04-06T07:03:25Z
postgresautoscalers.autoscaling.kubedb.com        2023-04-06T07:06:11Z
postgresdatabases.schema.kubedb.com               2023-04-06T07:06:16Z
postgreses.kubedb.com                             2023-04-06T07:06:16Z
postgresopsrequests.ops.kubedb.com                2023-04-06T07:06:45Z
postgresversions.catalog.kubedb.com               2023-04-06T07:03:25Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-04-06T07:06:11Z
proxysqlopsrequests.ops.kubedb.com                2023-04-06T07:06:49Z
proxysqls.kubedb.com                              2023-04-06T07:06:27Z
proxysqlversions.catalog.kubedb.com               2023-04-06T07:03:25Z
publishers.postgres.kubedb.com                    2023-04-06T07:07:03Z
redisautoscalers.autoscaling.kubedb.com           2023-04-06T07:06:11Z
redises.kubedb.com                                2023-04-06T07:06:28Z
redisopsrequests.ops.kubedb.com                   2023-04-06T07:06:40Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-04-06T07:06:11Z
redissentinelopsrequests.ops.kubedb.com           2023-04-06T07:06:57Z
redissentinels.kubedb.com                         2023-04-06T07:06:28Z
redisversions.catalog.kubedb.com                  2023-04-06T07:03:26Z
subscribers.postgres.kubedb.com                   2023-04-06T07:07:07Z
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
  version: opensearch-2.5.0 
  storageType: Durable
  topology:
    master:
      replicas: 2
      resources:
      storage:
        storageClassName: "gp2"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 2
      resources:
      storage:
        storageClassName: "gp2"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      resources:
      storage:
        storageClassName: "gp2"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  terminationPolicy: WipeOut
```

Here,

* `spec.version` - is the name of the ElasticsearchVersion CR. Here, we are using OpenSearch version `opensearch-2.5.0` of OpenSearch distribution.
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
pod/os-cluster-data-0     1/1     Running   0          2m27s
pod/os-cluster-data-1     1/1     Running   0          2m13s
pod/os-cluster-ingest-0   1/1     Running   0          2m28s
pod/os-cluster-ingest-1   1/1     Running   0          2m10s
pod/os-cluster-master-0   1/1     Running   0          2m28s
pod/os-cluster-master-1   1/1     Running   0          2m1s

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/os-cluster          ClusterIP   10.100.65.165   <none>        9200/TCP   2m36s
service/os-cluster-master   ClusterIP   None            <none>        9300/TCP   2m36s
service/os-cluster-pods     ClusterIP   None            <none>        9200/TCP   2m36s

NAME                                 READY   AGE
statefulset.apps/os-cluster-data     2/2     2m34s
statefulset.apps/os-cluster-ingest   2/2     2m35s
statefulset.apps/os-cluster-master   2/2     2m35s

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/os-cluster   kubedb.com/elasticsearch   2.5.0     2m38s

NAME                                  VERSION            STATUS   AGE
elasticsearch.kubedb.com/os-cluster   opensearch-2.5.0   Ready    3m1s
```
> We have successfully deployed OpenSearch cluster in AWS. 

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
os-cluster-dashboard   dashboard.kubedb.com/v1alpha1   os-cluster   Ready    76s
```
Here, OpenSearch Dashboard is in `Ready` state. 


## Connect with OpenSearch Dashboard

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our OpenSearch database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our OpenSearch database is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
os-cluster             ClusterIP   10.100.65.165    <none>        9200/TCP   5m57s
os-cluster-dashboard   ClusterIP   10.100.171.239   <none>        5601/TCP   2m12s
os-cluster-master      ClusterIP   None             <none>        9300/TCP   5m57s
os-cluster-pods        ClusterIP   None             <none>        9200/TCP   5m57s
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
os-cluster-admin-cert              kubernetes.io/tls                     3      7m19s
os-cluster-admin-cred              kubernetes.io/basic-auth              2      7m18s
os-cluster-ca-cert                 kubernetes.io/tls                     2      7m19s
os-cluster-client-cert             kubernetes.io/tls                     3      7m18s
os-cluster-config                  Opaque                                3      7m16s
os-cluster-dashboard-ca-cert       kubernetes.io/tls                     2      3m35s
os-cluster-dashboard-config        Opaque                                2      3m35s
os-cluster-dashboard-server-cert   kubernetes.io/tls                     3      3m35s
os-cluster-http-cert               kubernetes.io/tls                     3      7m18s
os-cluster-kibanaro-cred           kubernetes.io/basic-auth              2      7m18s
os-cluster-kibanaserver-cred       kubernetes.io/basic-auth              2      7m18s
os-cluster-logstash-cred           kubernetes.io/basic-auth              2      7m18s
os-cluster-readall-cred            kubernetes.io/basic-auth              2      7m18s
os-cluster-snapshotrestore-cred    kubernetes.io/basic-auth              2      7m18s
os-cluster-token-2xzrs             kubernetes.io/service-account-token   3      7m20s
os-cluster-transport-cert          kubernetes.io/tls                     3      7m19s
```
Now, we can connect to the database with `os-cluster-admin-cred` which contains the admin credentials to connect with the database.

### Accessing Database Through Dashboard

To access the database through Dashboard, we have to get the credentials. We can do that by following command,

```bash
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo os-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
PiX8O_FWDbLSrX(U
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

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Elasticsearch in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-elasticsearch-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
