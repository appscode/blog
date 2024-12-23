---
title: Monitor MariaDB with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-04-26"
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
- mariadb
- panopticon
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor MariaDB with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy MariaDB Clustered Database
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
appscode/kubedb-autoscaler        	v0.18.0      	v0.18.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.04.10  	v2023.04.10	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.04.10  	v2023.04.10	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.9.0       	v0.9.0     	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.04.10  	v2023.04.10	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.04.10  	v2023.04.10	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.04.10  	v2023.04.10	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.20.0      	v0.20.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.04.10  	v2023.04.10	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.33.0      	v0.33.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.9.0       	v0.9.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.03.23  	0.3.28     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.9.0       	v0.9.0     	KubeDB Webhook Server by AppsCode  

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
kubedb      kubedb-kubedb-autoscaler-65b7f46d94-gxdhl       1/1     Running   0          73s
kubedb      kubedb-kubedb-dashboard-685ddf696f-5rxvm        1/1     Running   0          73s
kubedb      kubedb-kubedb-ops-manager-8bb4c785-9mjh4        1/1     Running   0          73s
kubedb      kubedb-kubedb-provisioner-76f96d8c66-z9wg5      1/1     Running   0          73s
kubedb      kubedb-kubedb-schema-manager-6d46c6c6c7-kvlnl   1/1     Running   0          73s
kubedb      kubedb-kubedb-webhook-server-5d57f85758-lhxxx   1/1     Running   0          73s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-04-25T10:20:06Z
elasticsearchdashboards.dashboard.kubedb.com      2023-04-25T10:20:05Z
elasticsearches.kubedb.com                        2023-04-25T10:20:05Z
elasticsearchopsrequests.ops.kubedb.com           2023-04-25T10:20:11Z
elasticsearchversions.catalog.kubedb.com          2023-04-25T10:17:25Z
etcds.kubedb.com                                  2023-04-25T10:20:08Z
etcdversions.catalog.kubedb.com                   2023-04-25T10:17:25Z
kafkas.kubedb.com                                 2023-04-25T10:20:21Z
kafkaversions.catalog.kubedb.com                  2023-04-25T10:17:25Z
mariadbautoscalers.autoscaling.kubedb.com         2023-04-25T10:20:06Z
mariadbdatabases.schema.kubedb.com                2023-04-25T10:20:10Z
mariadbopsrequests.ops.kubedb.com                 2023-04-25T10:20:33Z
mariadbs.kubedb.com                               2023-04-25T10:20:09Z
mariadbversions.catalog.kubedb.com                2023-04-25T10:17:26Z
memcacheds.kubedb.com                             2023-04-25T10:20:09Z
memcachedversions.catalog.kubedb.com              2023-04-25T10:17:26Z
mongodbautoscalers.autoscaling.kubedb.com         2023-04-25T10:20:07Z
mongodbdatabases.schema.kubedb.com                2023-04-25T10:20:07Z
mongodbopsrequests.ops.kubedb.com                 2023-04-25T10:20:16Z
mongodbs.kubedb.com                               2023-04-25T10:20:07Z
mongodbversions.catalog.kubedb.com                2023-04-25T10:17:26Z
mysqlautoscalers.autoscaling.kubedb.com           2023-04-25T10:20:07Z
mysqldatabases.schema.kubedb.com                  2023-04-25T10:20:06Z
mysqlopsrequests.ops.kubedb.com                   2023-04-25T10:20:29Z
mysqls.kubedb.com                                 2023-04-25T10:20:06Z
mysqlversions.catalog.kubedb.com                  2023-04-25T10:17:27Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-04-25T10:20:07Z
perconaxtradbopsrequests.ops.kubedb.com           2023-04-25T10:20:46Z
perconaxtradbs.kubedb.com                         2023-04-25T10:20:18Z
perconaxtradbversions.catalog.kubedb.com          2023-04-25T10:17:27Z
pgbouncers.kubedb.com                             2023-04-25T10:20:18Z
pgbouncerversions.catalog.kubedb.com              2023-04-25T10:17:27Z
postgresautoscalers.autoscaling.kubedb.com        2023-04-25T10:20:07Z
postgresdatabases.schema.kubedb.com               2023-04-25T10:20:09Z
postgreses.kubedb.com                             2023-04-25T10:20:10Z
postgresopsrequests.ops.kubedb.com                2023-04-25T10:20:40Z
postgresversions.catalog.kubedb.com               2023-04-25T10:17:28Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-04-25T10:20:08Z
proxysqlopsrequests.ops.kubedb.com                2023-04-25T10:20:43Z
proxysqls.kubedb.com                              2023-04-25T10:20:20Z
proxysqlversions.catalog.kubedb.com               2023-04-25T10:17:28Z
publishers.postgres.kubedb.com                    2023-04-25T10:20:57Z
redisautoscalers.autoscaling.kubedb.com           2023-04-25T10:20:08Z
redises.kubedb.com                                2023-04-25T10:20:20Z
redisopsrequests.ops.kubedb.com                   2023-04-25T10:20:36Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-04-25T10:20:08Z
redissentinelopsrequests.ops.kubedb.com           2023-04-25T10:20:50Z
redissentinels.kubedb.com                         2023-04-25T10:20:21Z
redisversions.catalog.kubedb.com                  2023-04-25T10:17:28Z
subscribers.postgres.kubedb.com                   2023-04-25T10:21:01Z
```

### Install Prometheus Stack
Install Prometheus stack if you haven't done it already. You can use [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) which installs the necessary components required for the MariaDB Grafana dashboards.

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
kubeops     panopticon-57c7f6cc7d-h5k98   1/1     Running   0          71s
```


## Deploy MariaDB Clustered Database

Now, we are going to Deploy MariaDB with monitoring enabled using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of the MariaDB CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: mariadb-cluster
  namespace: demo
spec:
  version: "10.11.2"
  replicas: 3
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

Let's save this yaml configuration into `mariadb-cluster.yaml` 
Then create the above MariaDB CRO

```bash
$ kubectl apply -f mariadb-cluster.yaml
mariadb.kubedb.com/mariadb-cluster created
```

* In this yaml we can see in the `spec.version` field specifies the version of MariaDB. Here, we are using MariaDB `version 10.11.2`. You can list the KubeDB supported versions of MariaDB by running `$ kubectl get mariadbversion` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mariadb/concepts/mariadb/#specterminationpolicy).

Once these are handled correctly and the MariaDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=mariadb-cluster'
NAME                    READY   STATUS    RESTARTS   AGE
pod/mariadb-cluster-0   3/3     Running   0          2m3s
pod/mariadb-cluster-1   3/3     Running   0          2m3s
pod/mariadb-cluster-2   3/3     Running   0          2m3s

NAME                            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
service/mariadb-cluster         ClusterIP   10.100.6.127     <none>        3306/TCP    2m7s
service/mariadb-cluster-pods    ClusterIP   None             <none>        3306/TCP    2m7s
service/mariadb-cluster-stats   ClusterIP   10.100.165.232   <none>        56790/TCP   2m6s

NAME                               READY   AGE
statefulset.apps/mariadb-cluster   3/3     2m10s

NAME                                                 TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mariadb-cluster   kubedb.com/mariadb   10.11.2   2m13s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mariadb -n demo mariadb-cluster
NAME              VERSION   STATUS   AGE
mariadb-cluster   10.11.2   Ready    3m21s
```
> We have successfully deployed MariaDB in AWS.


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
alertmanager-operated                     ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   14m
kubernetes                                ClusterIP   10.100.0.1       <none>        443/TCP                      28m
prometheus-grafana                        ClusterIP   10.100.193.200   <none>        80/TCP                       14m
prometheus-kube-prometheus-alertmanager   ClusterIP   10.100.33.204    <none>        9093/TCP                     14m
prometheus-kube-prometheus-operator       ClusterIP   10.100.116.187   <none>        443/TCP                      14m
prometheus-kube-prometheus-prometheus     ClusterIP   10.100.92.58     <none>        9090/TCP                     14m
prometheus-kube-state-metrics             ClusterIP   10.100.27.215    <none>        8080/TCP                     14m
prometheus-operated                       ClusterIP   None             <none>        9090/TCP                     14m
prometheus-prometheus-node-exporter       ClusterIP   10.100.133.145   <none>        9100/TCP                     14m
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

Upload the json file or copy-paste the json codes to the panel json and hit the load button:

![Upload Json](upload-json.png)


For MariaDB Database Dashboard use [MariaDB Database Dashboard Json](https://github.com/appscode/grafana-dashboards/blob/master/mariadb/mariadb-database.json)

For MariaDB Group Replication use [MariaDB Galera Cluster Json](https://github.com/appscode/grafana-dashboards/blob/master/mariadb/mariadb-galera.json)

For MariaDB Summary use [MariaDB Summary Json](https://github.com/appscode/grafana-dashboards/blob/master/mariadb/mariadb-summary.json)

For MariaDB Pod use [MariaDB Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/mariadb/mariadb-pod.json)

If you followed above instruction properly you will see MariaDB Grafana Dashboards in your Grafana UI

Here are some screenshots of our MariaDB deployment. You can visualize every single component supported by Grafana, checkout [Grafana Dashboard](https://grafana.com/docs/grafana/latest/) for more information. 

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


We have made an in depth tutorial on MariaDB and Percona XtraDB Support for KubeDB ProxySQL, ProxySQL Dashboard & Alerts. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/lQ6EdG7q_CY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MariaDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mariadb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
