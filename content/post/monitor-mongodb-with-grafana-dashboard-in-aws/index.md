---
title: Monitor MongoDB with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-12-19"
weight: 14
authors:
- Dipta Roy
tags:
- aws
- cloud-native
- dashboard
- database
- eks
- grafana
- grafana-dashboard
- kubedb
- kubernetes
- mongodb
- panopticon
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor MongoDB with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy MongoDB Cluster
5) Monitor with Grafana Dashboard

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

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
NAME                              	CHART VERSION	APP VERSION  	DESCRIPTION                                       
appscode/kubedb                   	v2023.12.11  	v2023.12.11  	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.23.0      	v0.23.0      	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.12.11  	v2023.12.11  	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2      	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.12.11  	v2023.12.11  	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.14.0      	v0.14.0      	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2      	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.12.11  	v2023.12.11  	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2023.12.11  	v2023.12.11  	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2023.12.11  	v2023.12.11  	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.11  	v2023.12.11  	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.25.0      	v0.25.0      	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.12.11  	v2023.12.11  	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2023.12.11  	v0.0.2       	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2023.12.11  	v0.0.2       	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2023.12.11  	v0.0.2       	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.38.0      	v0.38.1      	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.14.0      	v0.14.0      	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.12.5   	0.6.1-alpha.2	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21  	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.14.0      	v0.14.0      	KubeDB Webhook Server by AppsCode  

$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2023.12.21 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS       AGE
kubedb      kubedb-kubedb-autoscaler-798b79c658-zd24v       1/1     Running   0              3m9s
kubedb      kubedb-kubedb-dashboard-7fdc974d5d-kwmdw        1/1     Running   0              3m9s
kubedb      kubedb-kubedb-ops-manager-64db46dcd4-s775f      1/1     Running   0              3m9s
kubedb      kubedb-kubedb-provisioner-f48fc5869-kml7h       1/1     Running   0              3m9s
kubedb      kubedb-kubedb-webhook-server-669d4d467d-mfsrg   1/1     Running   0              3m9s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-12-19T09:12:07Z
elasticsearchdashboards.dashboard.kubedb.com      2023-12-19T09:12:05Z
elasticsearches.kubedb.com                        2023-12-19T09:12:05Z
elasticsearchopsrequests.ops.kubedb.com           2023-12-19T09:12:17Z
elasticsearchversions.catalog.kubedb.com          2023-12-19T09:09:58Z
etcds.kubedb.com                                  2023-12-19T09:12:20Z
etcdversions.catalog.kubedb.com                   2023-12-19T09:09:59Z
kafkaopsrequests.ops.kubedb.com                   2023-12-19T09:13:10Z
kafkas.kubedb.com                                 2023-12-19T09:12:29Z
kafkaversions.catalog.kubedb.com                  2023-12-19T09:09:59Z
mariadbautoscalers.autoscaling.kubedb.com         2023-12-19T09:12:07Z
mariadbopsrequests.ops.kubedb.com                 2023-12-19T09:12:46Z
mariadbs.kubedb.com                               2023-12-19T09:12:21Z
mariadbversions.catalog.kubedb.com                2023-12-19T09:09:59Z
memcacheds.kubedb.com                             2023-12-19T09:12:21Z
memcachedversions.catalog.kubedb.com              2023-12-19T09:10:00Z
mongodbarchivers.archiver.kubedb.com              2023-12-19T09:12:33Z
mongodbautoscalers.autoscaling.kubedb.com         2023-12-19T09:12:07Z
mongodbopsrequests.ops.kubedb.com                 2023-12-19T09:12:21Z
mongodbs.kubedb.com                               2023-12-19T09:12:22Z
mongodbversions.catalog.kubedb.com                2023-12-19T09:10:00Z
mysqlautoscalers.autoscaling.kubedb.com           2023-12-19T09:12:07Z
mysqlopsrequests.ops.kubedb.com                   2023-12-19T09:12:42Z
mysqls.kubedb.com                                 2023-12-19T09:12:24Z
mysqlversions.catalog.kubedb.com                  2023-12-19T09:10:00Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-12-19T09:12:07Z
perconaxtradbopsrequests.ops.kubedb.com           2023-12-19T09:13:03Z
perconaxtradbs.kubedb.com                         2023-12-19T09:12:24Z
perconaxtradbversions.catalog.kubedb.com          2023-12-19T09:10:01Z
pgbouncers.kubedb.com                             2023-12-19T09:12:26Z
pgbouncerversions.catalog.kubedb.com              2023-12-19T09:10:01Z
postgresarchivers.archiver.kubedb.com             2023-12-19T09:12:36Z
postgresautoscalers.autoscaling.kubedb.com        2023-12-19T09:12:08Z
postgreses.kubedb.com                             2023-12-19T09:12:27Z
postgresopsrequests.ops.kubedb.com                2023-12-19T09:12:55Z
postgresversions.catalog.kubedb.com               2023-12-19T09:10:01Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-12-19T09:12:08Z
proxysqlopsrequests.ops.kubedb.com                2023-12-19T09:12:59Z
proxysqls.kubedb.com                              2023-12-19T09:12:27Z
proxysqlversions.catalog.kubedb.com               2023-12-19T09:10:01Z
publishers.postgres.kubedb.com                    2023-12-19T09:13:14Z
redisautoscalers.autoscaling.kubedb.com           2023-12-19T09:12:08Z
redises.kubedb.com                                2023-12-19T09:12:28Z
redisopsrequests.ops.kubedb.com                   2023-12-19T09:12:49Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-12-19T09:12:09Z
redissentinelopsrequests.ops.kubedb.com           2023-12-19T09:13:06Z
redissentinels.kubedb.com                         2023-12-19T09:12:29Z
redisversions.catalog.kubedb.com                  2023-12-19T09:10:02Z
subscribers.postgres.kubedb.com                   2023-12-19T09:13:18Z
```

### Install Prometheus Stack
Install Prometheus stack which installs the necessary components required for the MongoDB Grafana dashboards. You can use following commands,

```bash
$ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
$ helm install prometheus prometheus-community/kube-prometheus-stack
```
or visit [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) for more detailed information.

### Install Panopticon
KubeDB License works for Panopticon too. So, we will use the same license that we have already obtained.

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
kubeops     panopticon-5fd64785c7-f89fp   1/1     Running   0          86s
```


## Deploy MongoDB Cluster

Now, we are going to Deploy MongoDB with monitoring enabled using KubeDB.
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
  name: mongodb-cluster
  namespace: demo
spec:
  version: "6.0.12"
  replicas: 3
  replicaSet:
    name: rs
  storage:
    storageClassName: "gp2"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `mongodb-cluster.yaml` 
Then create the above MongoDB CRO

```bash
$ kubectl apply -f mongodb-cluster.yaml 
mongodb.kubedb.com/mongodb-cluster created
```

In this yaml,
* `spec.version` field specifies the version of MongoDB. Here, we are using MongoDB `version 6.0.12`. You can list the KubeDB supported versions of MongoDB by running `$ kubectl get mongodbversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mongodb/concepts/mongodb/#specterminationpolicy).

Once these are handled correctly and the MongoDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=mongodb-cluster'
NAME                    READY   STATUS    RESTARTS   AGE
pod/mongodb-cluster-0   3/3     Running   0          3m51s
pod/mongodb-cluster-1   3/3     Running   0          2m44s
pod/mongodb-cluster-2   3/3     Running   0          98s

NAME                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mongodb-cluster         ClusterIP   10.76.5.149    <none>        27017/TCP   3m53s
service/mongodb-cluster-pods    ClusterIP   None           <none>        27017/TCP   3m54s
service/mongodb-cluster-stats   ClusterIP   10.76.12.154   <none>        56790/TCP   3m50s

NAME                               READY   AGE
statefulset.apps/mongodb-cluster   3/3     3m54s

NAME                                                 TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mongodb-cluster   kubedb.com/mongodb   6.0.12    3m53s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo mongodb-cluster
NAME              VERSION   STATUS   AGE
mongodb-cluster   6.0.12    Ready    4m28s
```
> We have successfully deployed MongoDB in AWS.


### Create DB Metrics Configurations

First, you have to create a `MetricsConfiguration` object for database. This `MetricsConfiguration` object is used by Panopticon to generate metrics for DB instances.
Install `kubedb-metrics` charts which will create the `MetricsConfiguration` object for DB:

```bash
$ helm search repo appscode/kubedb-metrics --version=v2023.12.21
$ helm install kubedb-metrics appscode/kubedb-metrics -n kubedb --version=v2023.12.21
```

### Import Grafana Dashboard
Here, we will port-forward the `prometheus-grafana` service to access Grafana Dashboard from UI.

```bash
$ kubectl get service -n default
NAME                                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None           <none>        9093/TCP,9094/TCP,9094/UDP   7m
kubernetes                                ClusterIP   10.76.0.1      <none>        443/TCP                      13m
prometheus-grafana                        ClusterIP   10.76.12.110   <none>        80/TCP                       7m
prometheus-kube-prometheus-alertmanager   ClusterIP   10.76.10.203   <none>        9093/TCP,8080/TCP            7m
prometheus-kube-prometheus-operator       ClusterIP   10.76.12.243   <none>        443/TCP                      7m
prometheus-kube-prometheus-prometheus     ClusterIP   10.76.14.73    <none>        9090/TCP,8080/TCP            7m
prometheus-kube-state-metrics             ClusterIP   10.76.10.77    <none>        8080/TCP                     7m
prometheus-operated                       ClusterIP   None           <none>        9090/TCP                     7m
prometheus-prometheus-node-exporter       ClusterIP   10.76.14.3     <none>        9100/TCP                     7m
```
To access Grafana UI Let's port-forward `prometheus-grafana` service to 3063 

```bash
$ kubectl port-forward -n default service/prometheus-grafana 3063:80
Forwarding from 127.0.0.1:3063 -> 3000
Forwarding from [::1]:3063 -> 3000
Handling connection for 3063

```
Now, Go to http://localhost:3063/ you will see a login panel of the Grafana UI, use default credential `admin` as the `Username` and `prom-operator` as the `Password`.

![Grafana Login](grafana-login.png)

After logged in successfuly on Grafana UI, import the json files of dashboards given below according to your choice.

Select Import button from left bar of the Grafana UI

![Import Dashboard](import-dashboard.png)

Upload the json file or copy and paste the desired MongoDB dashboard json file from [HERE](https://github.com/appscode/grafana-dashboards/tree/master/mongodb) and paste it to the panel json and hit the load button:

![Upload Json](upload-json.png)


For MongoDB Replicaset Database use [MongoDB Replicaset Database Json](https://github.com/appscode/grafana-dashboards/blob/master/mongodb/mongodb-database-replset-dashboard.json)

For MongoDB Pod use [MongoDB Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/mongodb/mongodb-pod-dashboard.json)

For MongoDB Summary Dashboard use [MongoDB Summary Dashboard Json](https://github.com/appscode/grafana-dashboards/blob/master/mongodb/mongodb-summary-dashboard.json)

If you followed above instruction properly you will see MongoDB Grafana Dashboards in your Grafana UI

Here are some screenshots of our MongoDB deployment. You can visualize every single component supported by Grafana, checkout [Grafana Dashboard](https://grafana.com/docs/grafana/latest/) for more information. 

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


If you want to learn more about Production-Grade MongoDB on Kubernetes you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1jZmdzRaQW28eX4zR9lvUqf" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MongoDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mongodb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
