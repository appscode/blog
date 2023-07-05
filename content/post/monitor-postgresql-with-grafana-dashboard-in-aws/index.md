---
title: Monitor PostgreSQL with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-05-24"
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
- postgresql
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor PostgreSQL with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy PostgreSQL Clustered Database
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
appscode/kubedb                   	v2023.04.10  	v2023.04.10	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.18.0      	v0.18.1    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.04.10  	v2023.04.10	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.04.10  	v2023.04.10	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.9.0       	v0.9.1     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.04.10  	v2023.04.10	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.04.10  	v2023.04.10	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.04.10  	v2023.04.10	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.20.0      	v0.20.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.04.10  	v2023.04.10	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.33.0      	v0.33.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.9.0       	v0.9.1     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.9.0       	v0.9.1     	KubeDB Webhook Server by AppsCode  

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.04.10 \
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

NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-6f489c85cd-qr5kb       1/1     Running   0          2m20s
kubedb      kubedb-kubedb-dashboard-6f7f59c498-2jvql        1/1     Running   0          2m20s
kubedb      kubedb-kubedb-ops-manager-797c8fc648-vcsxx      1/1     Running   0          2m20s
kubedb      kubedb-kubedb-provisioner-688fcfd4bd-8mjts      1/1     Running   0          2m20s
kubedb      kubedb-kubedb-schema-manager-854c9dbf58-cpsp7   1/1     Running   0          2m20s
kubedb      kubedb-kubedb-webhook-server-647bf447f8-d4zx9   1/1     Running   0          2m20s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-05-23T06:16:51Z
elasticsearchdashboards.dashboard.kubedb.com      2023-05-23T06:16:52Z
elasticsearches.kubedb.com                        2023-05-23T06:16:52Z
elasticsearchopsrequests.ops.kubedb.com           2023-05-23T06:16:57Z
elasticsearchversions.catalog.kubedb.com          2023-05-23T06:14:22Z
etcds.kubedb.com                                  2023-05-23T06:16:58Z
etcdversions.catalog.kubedb.com                   2023-05-23T06:14:23Z
kafkas.kubedb.com                                 2023-05-23T06:17:08Z
kafkaversions.catalog.kubedb.com                  2023-05-23T06:14:23Z
mariadbautoscalers.autoscaling.kubedb.com         2023-05-23T06:16:51Z
mariadbdatabases.schema.kubedb.com                2023-05-23T06:17:00Z
mariadbopsrequests.ops.kubedb.com                 2023-05-23T06:17:16Z
mariadbs.kubedb.com                               2023-05-23T06:16:59Z
mariadbversions.catalog.kubedb.com                2023-05-23T06:14:23Z
memcacheds.kubedb.com                             2023-05-23T06:16:59Z
memcachedversions.catalog.kubedb.com              2023-05-23T06:14:23Z
mongodbautoscalers.autoscaling.kubedb.com         2023-05-23T06:16:51Z
mongodbdatabases.schema.kubedb.com                2023-05-23T06:16:58Z
mongodbopsrequests.ops.kubedb.com                 2023-05-23T06:17:01Z
mongodbs.kubedb.com                               2023-05-23T06:16:58Z
mongodbversions.catalog.kubedb.com                2023-05-23T06:14:24Z
mysqlautoscalers.autoscaling.kubedb.com           2023-05-23T06:16:51Z
mysqldatabases.schema.kubedb.com                  2023-05-23T06:16:57Z
mysqlopsrequests.ops.kubedb.com                   2023-05-23T06:17:12Z
mysqls.kubedb.com                                 2023-05-23T06:16:57Z
mysqlversions.catalog.kubedb.com                  2023-05-23T06:14:24Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-05-23T06:16:51Z
perconaxtradbopsrequests.ops.kubedb.com           2023-05-23T06:17:29Z
perconaxtradbs.kubedb.com                         2023-05-23T06:17:06Z
perconaxtradbversions.catalog.kubedb.com          2023-05-23T06:14:24Z
pgbouncers.kubedb.com                             2023-05-23T06:17:06Z
pgbouncerversions.catalog.kubedb.com              2023-05-23T06:14:25Z
postgresautoscalers.autoscaling.kubedb.com        2023-05-23T06:16:51Z
postgresdatabases.schema.kubedb.com               2023-05-23T06:16:59Z
postgreses.kubedb.com                             2023-05-23T06:16:59Z
postgresopsrequests.ops.kubedb.com                2023-05-23T06:17:23Z
postgresversions.catalog.kubedb.com               2023-05-23T06:14:25Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-05-23T06:16:51Z
proxysqlopsrequests.ops.kubedb.com                2023-05-23T06:17:26Z
proxysqls.kubedb.com                              2023-05-23T06:17:07Z
proxysqlversions.catalog.kubedb.com               2023-05-23T06:14:25Z
publishers.postgres.kubedb.com                    2023-05-23T06:17:39Z
redisautoscalers.autoscaling.kubedb.com           2023-05-23T06:16:51Z
redises.kubedb.com                                2023-05-23T06:17:07Z
redisopsrequests.ops.kubedb.com                   2023-05-23T06:17:19Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-05-23T06:16:51Z
redissentinelopsrequests.ops.kubedb.com           2023-05-23T06:17:33Z
redissentinels.kubedb.com                         2023-05-23T06:17:07Z
redisversions.catalog.kubedb.com                  2023-05-23T06:14:26Z
subscribers.postgres.kubedb.com                   2023-05-23T06:17:42Z
```

### Install Prometheus Stack
Install Prometheus stack if you haven't done it already. You can use [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) which installs the necessary components required for the PostgreSQL Grafana dashboards.

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
kubeops     panopticon-5cb5b56bdf-kpbhr   1/1     Running   0          99s
```


## Deploy PostgreSQL Clustered Database

Now, we are going to Deploy PostgreSQL with monitoring enabled using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of the PostgreSQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: postgres-cluster
  namespace: demo
spec:
  version: "15.1"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
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


Let's save this yaml configuration into `postgres-cluster.yaml` 
Then create the above PostgreSQL CRO

```bash
$ kubectl apply -f postgres-cluster.yaml
postgres.kubedb.com/postgres-cluster created
```

In this yaml,
* `spec.version` field specifies the version of PostgreSQL. Here, we are using PostgreSQL `version 15.1`. You can list the KubeDB supported versions of PostgreSQL by running `$ kubectl get postgresversion` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/postgres/concepts/postgres/#specterminationpolicy).

Once these are handled correctly and the PostgreSQL object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=postgres-cluster'
NAME                     READY   STATUS    RESTARTS   AGE
pod/postgres-cluster-0   3/3     Running   0          3m35s
pod/postgres-cluster-1   3/3     Running   0          3m14s
pod/postgres-cluster-2   3/3     Running   0          2m48s

NAME                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/postgres-cluster           ClusterIP   10.100.69.98     <none>        5432/TCP,2379/TCP            3m38s
service/postgres-cluster-pods      ClusterIP   None             <none>        5432/TCP,2380/TCP,2379/TCP   3m38s
service/postgres-cluster-standby   ClusterIP   10.100.248.107   <none>        5432/TCP                     3m38s
service/postgres-cluster-stats     ClusterIP   10.100.136.153   <none>        56790/TCP,23790/TCP          3m37s

NAME                                READY   AGE
statefulset.apps/postgres-cluster   3/3     3m41s

NAME                                                  TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/postgres-cluster   kubedb.com/postgres   15.1      3m45s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get postgres -n demo postgres-cluster
NAME               VERSION   STATUS   AGE
postgres-cluster   15.1      Ready    4m39s
```
> We have successfully deployed PostgreSQL in AWS.


### Create DB Metrics Configurations

First, you have to create a `MetricsConfiguration` object for database. This `MetricsConfiguration` object is used by Panopticon to generate metrics for DB instances.
Install `kubedb-metrics` charts which will create the `MetricsConfiguration` object for DB:

```bash
$ helm search repo appscode/kubedb-metrics --version=v2023.04.10
$ helm install kubedb-metrics appscode/kubedb-metrics -n kubedb --version=v2023.04.10
```

### Import Grafana Dashboard
Here, we will port-forward the `prometheus-grafana` service to access Grafana Dashboard from UI.

```bash
$ kubectl get service -n default
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   13m
kubernetes                                ClusterIP   10.100.0.1       <none>        443/TCP                      90m
prometheus-grafana                        ClusterIP   10.100.55.47     <none>        80/TCP                       13m
prometheus-kube-prometheus-alertmanager   ClusterIP   10.100.57.197    <none>        9093/TCP                     13m
prometheus-kube-prometheus-operator       ClusterIP   10.100.176.91    <none>        443/TCP                      13m
prometheus-kube-prometheus-prometheus     ClusterIP   10.100.91.105    <none>        9090/TCP                     13m
prometheus-kube-state-metrics             ClusterIP   10.100.143.45    <none>        8080/TCP                     13m
prometheus-operated                       ClusterIP   None             <none>        9090/TCP                     13m
prometheus-prometheus-node-exporter       ClusterIP   10.100.166.158   <none>        9100/TCP                     13m
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


For PostgreSQL Database Dashboard use [PostgreSQL Database Dashboard Json](https://github.com/appscode/grafana-dashboards/blob/master/postgres/postgres_databases_dashboard.json)

For PostgreSQL Summary use [PostgreSQL Summary Json](https://github.com/appscode/grafana-dashboards/blob/master/postgres/postgres_summary_dashboard.json)

For PostgreSQL Pod use [PostgreSQL Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/postgres/postgres_pods_dashboard.json)

If you followed above instruction properly you will see PostgreSQL Grafana Dashboards in your Grafana UI

Here are some screenshots of our PostgreSQL deployment. You can visualize every single component supported by Grafana, checkout [Grafana Dashboard](https://grafana.com/docs/grafana/latest/) for more information. 

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


If you want to learn more about Production-Grade PostgreSQL you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?list=PLoiT1Gv2KR1imqnrYFhUNTLHdBNFXPKr_" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [PostgreSQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-postgres-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
