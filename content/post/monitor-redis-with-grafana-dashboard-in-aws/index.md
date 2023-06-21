---
title: Monitor Redis with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-06-21"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- dashboard
- database
- eks
- grafana
- grafana-dashboard
- kubedb
- kubernetes
- panopticon
- redis
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor Redis with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy Redis Clustered Database
5) Monitor with Grafana Dashboard

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8c4498337-358b-4dc0-be52-14440f4e061e
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
appscode/kubedb-ops-manager       	v0.21.0      	v0.21.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.06.19  	v2023.06.19	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.34.0      	v0.34.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.10.0      	v0.10.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
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
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS       AGE
kubedb      kubedb-kubedb-autoscaler-6ffd97b8d6-ms5xs       1/1     Running   0              2m16s
kubedb      kubedb-kubedb-dashboard-7c65fbfbb8-2gwtf        1/1     Running   0              2m16s
kubedb      kubedb-kubedb-ops-manager-7bcf56cd68-rrbks      1/1     Running   0              2m16s
kubedb      kubedb-kubedb-provisioner-79f796b8c6-x5rdt      1/1     Running   0              2m16s
kubedb      kubedb-kubedb-schema-manager-57c75c7fdb-gqst2   1/1     Running   0              2m16s
kubedb      kubedb-kubedb-webhook-server-6d8cf7cb9c-lrcgx   1/1     Running   0              2m16s

```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-06-20T05:27:09Z
elasticsearchdashboards.dashboard.kubedb.com      2023-06-20T05:27:09Z
elasticsearches.kubedb.com                        2023-06-20T05:27:09Z
elasticsearchopsrequests.ops.kubedb.com           2023-06-20T05:27:10Z
elasticsearchversions.catalog.kubedb.com          2023-06-20T05:24:21Z
etcds.kubedb.com                                  2023-06-20T05:27:15Z
etcdversions.catalog.kubedb.com                   2023-06-20T05:24:22Z
kafkas.kubedb.com                                 2023-06-20T05:27:31Z
kafkaversions.catalog.kubedb.com                  2023-06-20T05:24:22Z
mariadbautoscalers.autoscaling.kubedb.com         2023-06-20T05:27:09Z
mariadbdatabases.schema.kubedb.com                2023-06-20T05:27:37Z
mariadbopsrequests.ops.kubedb.com                 2023-06-20T05:27:35Z
mariadbs.kubedb.com                               2023-06-20T05:27:16Z
mariadbversions.catalog.kubedb.com                2023-06-20T05:24:22Z
memcacheds.kubedb.com                             2023-06-20T05:27:16Z
memcachedversions.catalog.kubedb.com              2023-06-20T05:24:23Z
mongodbautoscalers.autoscaling.kubedb.com         2023-06-20T05:27:09Z
mongodbdatabases.schema.kubedb.com                2023-06-20T05:27:15Z
mongodbopsrequests.ops.kubedb.com                 2023-06-20T05:27:13Z
mongodbs.kubedb.com                               2023-06-20T05:27:14Z
mongodbversions.catalog.kubedb.com                2023-06-20T05:24:23Z
mysqlautoscalers.autoscaling.kubedb.com           2023-06-20T05:27:09Z
mysqldatabases.schema.kubedb.com                  2023-06-20T05:27:13Z
mysqlopsrequests.ops.kubedb.com                   2023-06-20T05:27:31Z
mysqls.kubedb.com                                 2023-06-20T05:27:14Z
mysqlversions.catalog.kubedb.com                  2023-06-20T05:24:23Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-06-20T05:27:10Z
perconaxtradbopsrequests.ops.kubedb.com           2023-06-20T05:27:50Z
perconaxtradbs.kubedb.com                         2023-06-20T05:27:26Z
perconaxtradbversions.catalog.kubedb.com          2023-06-20T05:24:24Z
pgbouncers.kubedb.com                             2023-06-20T05:27:22Z
pgbouncerversions.catalog.kubedb.com              2023-06-20T05:24:24Z
postgresautoscalers.autoscaling.kubedb.com        2023-06-20T05:27:10Z
postgresdatabases.schema.kubedb.com               2023-06-20T05:27:27Z
postgreses.kubedb.com                             2023-06-20T05:27:28Z
postgresopsrequests.ops.kubedb.com                2023-06-20T05:27:43Z
postgresversions.catalog.kubedb.com               2023-06-20T05:24:24Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-06-20T05:27:10Z
proxysqlopsrequests.ops.kubedb.com                2023-06-20T05:27:47Z
proxysqls.kubedb.com                              2023-06-20T05:27:29Z
proxysqlversions.catalog.kubedb.com               2023-06-20T05:24:25Z
publishers.postgres.kubedb.com                    2023-06-20T05:28:00Z
redisautoscalers.autoscaling.kubedb.com           2023-06-20T05:27:10Z
redises.kubedb.com                                2023-06-20T05:27:30Z
redisopsrequests.ops.kubedb.com                   2023-06-20T05:27:39Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-06-20T05:27:10Z
redissentinelopsrequests.ops.kubedb.com           2023-06-20T05:27:54Z
redissentinels.kubedb.com                         2023-06-20T05:27:30Z
redisversions.catalog.kubedb.com                  2023-06-20T05:24:25Z
subscribers.postgres.kubedb.com                   2023-06-20T05:28:04Z
```

### Install Prometheus Stack
Install Prometheus stack if you haven't done it already. You can use [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) which installs the necessary components required for the Redis Grafana dashboards.

### Install Panopticon
KubeDB Enterprise License works for Panopticon too. So, we will use the same license that we have already obtained.

```bash
$ helm install panopticon appscode/panopticon -n kubeops \
    --create-namespace \
    --set monitoring.enabled=true \
    --set monitoring.agent=prometheus.io/operator \
    --set monitoring.serviceMonitor.labels.release=prometheus \
    --set-file license=/path/to/license.txt
```
Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=panopticon"
NAMESPACE   NAME                          READY   STATUS    RESTARTS   AGE
kubeops     panopticon-5dcd55d79b-k9j67   1/1     Running   0          31s
```


## Deploy Redis Clustered Database

Now, we are going to Deploy Redis with monitoring enabled using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of the Redis CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 7.0.10
  mode: Cluster
  cluster:
    master: 3
    replicas: 1 
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "gp2"
    accessModes:
    - ReadWriteOnce
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `redis-cluster.yaml` 
Then create the above Redis CRO

```bash
$ kubectl apply -f redis-cluster.yaml
redis.kubedb.com/redis-cluster created
```

In this yaml,
* `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.0.10`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these checkout [Termination Policy](https://kubedb.com/docs/v2023.06.19/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=redis-cluster'
NAME                         READY   STATUS    RESTARTS   AGE
pod/redis-cluster-shard0-0   2/2     Running   0          69s
pod/redis-cluster-shard0-1   2/2     Running   0          43s
pod/redis-cluster-shard1-0   2/2     Running   0          68s
pod/redis-cluster-shard1-1   2/2     Running   0          45s
pod/redis-cluster-shard2-0   2/2     Running   0          67s
pod/redis-cluster-shard2-1   2/2     Running   0          40s

NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
service/redis-cluster         ClusterIP   10.100.224.207   <none>        6379/TCP    71s
service/redis-cluster-pods    ClusterIP   None             <none>        6379/TCP    72s
service/redis-cluster-stats   ClusterIP   10.100.107.15    <none>        56790/TCP   69s

NAME                                    READY   AGE
statefulset.apps/redis-cluster-shard0   2/2     75s
statefulset.apps/redis-cluster-shard1   2/2     74s
statefulset.apps/redis-cluster-shard2   2/2     73s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/redis-cluster   kubedb.com/redis   7.0.10    77s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo redis-cluster
NAME            VERSION   STATUS   AGE
redis-cluster   7.0.10    Ready    2m34s
```
> We have successfully deployed redis in AWS.


### Create DB Metrics Configurations

First, you have to create a `MetricsConfiguration` object for database. This `MetricsConfiguration` object is used by Panopticon to generate metrics for DB instances.
Install `kubedb-metrics` charts which will create the `MetricsConfiguration` object for DB:

```bash
$ helm search repo appscode/kubedb-metrics --version=v2023.06.19
$ helm install kubedb-metrics appscode/kubedb-metrics -n kubedb --version=v2023.06.19
```

### Import Grafana Dashboard
Here, we will port-forward the `prometheus-grafana` service to access Grafana Dashboard from UI.

```bash
$ kubectl get service -n default
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   9m24s
kubernetes                                ClusterIP   10.100.0.1       <none>        443/TCP                      37m
prometheus-grafana                        ClusterIP   10.100.149.171   <none>        80/TCP                       9m29s
prometheus-kube-prometheus-alertmanager   ClusterIP   10.100.82.140    <none>        9093/TCP                     9m29s
prometheus-kube-prometheus-operator       ClusterIP   10.100.172.75    <none>        443/TCP                      9m29s
prometheus-kube-prometheus-prometheus     ClusterIP   10.100.59.16     <none>        9090/TCP                     9m29s
prometheus-kube-state-metrics             ClusterIP   10.100.182.47    <none>        8080/TCP                     9m29s
prometheus-operated                       ClusterIP   None             <none>        9090/TCP                     9m24s
prometheus-prometheus-node-exporter       ClusterIP   10.100.188.82    <none>        9100/TCP                     9m29s
```
To access Grafana UI Let's port-forward `prometheus-grafana` service to 3063 

```bash
$ kubectl port-forward -n default service/prometheus-grafana 3063:80
Forwarding from 127.0.0.1:3063 -> 3000
Forwarding from [::1]:3063 -> 3000
Handling connection for 3063

```
Now, Go to http://localhost:3063/ you will see a login panel of the Grafana UI, use default credential `admin` as the `Username` and `
` as the `Password`.

![Grafana Login](grafana-login.png)

After logged in successfuly on Grafana UI, import the json files of dashboards given below according to your choice.

Select Import button from left bar of the Grafana UI

![Import Dashboard](import-dashboard.png)

Upload the json file or copy-paste the json codes to the panel json and hit the load button:

![Upload Json](upload-json.png)


For Redis Summary Dashboard use [Redis Summary Dashboard Json](https://github.com/appscode/grafana-dashboards/blob/master/redis/redis_summary_dashboard.json)

For Redis Pod use [Redis Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/redis/redis_pod_dashboard.json)

For Redis Shard use [Redis Shard Json](https://github.com/appscode/grafana-dashboards/blob/master/redis/redis_shards_dashboard.json)

If you followed above instruction properly you will see Redis Grafana Dashboards in your Grafana UI

Here are some screenshots of our Redis deployment. You can visualize every single component supported by Grafana, checkout [Grafana Dashboard](https://grafana.com/docs/grafana/latest/) for more information. 

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


We have made an in depth video on how to Deploy Sharded Redis Cluster in Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/J7QI4EzuOVY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
