---
title: Monitor Percona XtraDB with Grafana Dashboard in Google Kubernetes Engine (GKE)
date: "2023-03-03"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- dashboard
- database
- gcp
- gcs
- gke
- grafana
- grafana-dashboard
- kubedb
- kubernetes
- panopticon
- percona-xtradb
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases [here](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor Percona XtraDB with Grafana Dashboard in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy Percona XtraDB Clustered Database
5) Monitor with Grafana Dashboard

### Get Cluster ID

We need the cluster ID to get the KubeDB License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
debacab3-y89q-4168-ba24-e97a553dcfa4
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
appscode/kubedb-ops-manager       	v0.19.0      	v0.19.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.02.28  	v2023.02.28	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.32.0      	v0.32.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.8.0       	v0.8.0     	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2022.06.14  	0.3.26     	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.8.0       	v0.8.0     	KubeDB Webhook Server by AppsCode  

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
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
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS      AGE
kubedb      kubedb-kubedb-autoscaler-664f65f787-ppztb       1/1     Running   0             4m
kubedb      kubedb-kubedb-dashboard-b666d767b-r6wn7         1/1     Running   0             4m 
kubedb      kubedb-kubedb-ops-manager-b7d8bb57c-qmc7g       1/1     Running   0             4m  
kubedb      kubedb-kubedb-provisioner-657f465467-8cvg5      1/1     Running   0             4m
kubedb      kubedb-kubedb-schema-manager-bd856b6d7-d452n    1/1     Running   0             4m
kubedb      kubedb-kubedb-webhook-server-7754958d57-s8xg9   1/1     Running   0             4m

```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-03-03T06:32:26Z
elasticsearchdashboards.dashboard.kubedb.com      2023-03-03T06:33:50Z
elasticsearches.kubedb.com                        2023-03-03T06:33:01Z
elasticsearchopsrequests.ops.kubedb.com           2023-03-03T06:33:20Z
elasticsearchversions.catalog.kubedb.com          2023-03-03T06:31:46Z
etcds.kubedb.com                                  2023-03-03T06:33:01Z
etcdversions.catalog.kubedb.com                   2023-03-03T06:31:46Z
kafkas.kubedb.com                                 2023-03-03T06:33:02Z
kafkaversions.catalog.kubedb.com                  2023-03-03T06:31:46Z
mariadbautoscalers.autoscaling.kubedb.com         2023-03-03T06:32:26Z
mariadbdatabases.schema.kubedb.com                2023-03-03T06:33:38Z
mariadbopsrequests.ops.kubedb.com                 2023-03-03T06:33:33Z
mariadbs.kubedb.com                               2023-03-03T06:33:01Z
mariadbversions.catalog.kubedb.com                2023-03-03T06:31:46Z
memcacheds.kubedb.com                             2023-03-03T06:33:01Z
memcachedversions.catalog.kubedb.com              2023-03-03T06:31:46Z
mongodbautoscalers.autoscaling.kubedb.com         2023-03-03T06:32:26Z
mongodbdatabases.schema.kubedb.com                2023-03-03T06:33:38Z
mongodbopsrequests.ops.kubedb.com                 2023-03-03T06:33:23Z
mongodbs.kubedb.com                               2023-03-03T06:33:01Z
mongodbversions.catalog.kubedb.com                2023-03-03T06:31:46Z
mysqlautoscalers.autoscaling.kubedb.com           2023-03-03T06:32:26Z
mysqldatabases.schema.kubedb.com                  2023-03-03T06:33:37Z
mysqlopsrequests.ops.kubedb.com                   2023-03-03T06:33:30Z
mysqls.kubedb.com                                 2023-03-03T06:33:01Z
mysqlversions.catalog.kubedb.com                  2023-03-03T06:31:46Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-03-03T06:32:26Z
perconaxtradbopsrequests.ops.kubedb.com           2023-03-03T06:33:46Z
perconaxtradbs.kubedb.com                         2023-03-03T06:33:01Z
perconaxtradbversions.catalog.kubedb.com          2023-03-03T06:31:46Z
pgbouncers.kubedb.com                             2023-03-03T06:33:01Z
pgbouncerversions.catalog.kubedb.com              2023-03-03T06:31:46Z
postgresautoscalers.autoscaling.kubedb.com        2023-03-03T06:32:26Z
postgresdatabases.schema.kubedb.com               2023-03-03T06:33:38Z
postgreses.kubedb.com                             2023-03-03T06:33:01Z
postgresopsrequests.ops.kubedb.com                2023-03-03T06:33:39Z
postgresversions.catalog.kubedb.com               2023-03-03T06:31:46Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-03-03T06:32:26Z
proxysqlopsrequests.ops.kubedb.com                2023-03-03T06:33:43Z
proxysqls.kubedb.com                              2023-03-03T06:33:01Z
proxysqlversions.catalog.kubedb.com               2023-03-03T06:31:46Z
publishers.postgres.kubedb.com                    2023-03-03T06:33:55Z
redisautoscalers.autoscaling.kubedb.com           2023-03-03T06:32:26Z
redises.kubedb.com                                2023-03-03T06:33:01Z
redisopsrequests.ops.kubedb.com                   2023-03-03T06:33:36Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-03-03T06:32:26Z
redissentinelopsrequests.ops.kubedb.com           2023-03-03T06:33:49Z
redissentinels.kubedb.com                         2023-03-03T06:33:02Z
redisversions.catalog.kubedb.com                  2023-03-03T06:31:46Z
subscribers.postgres.kubedb.com                   2023-03-03T06:33:58Z
```

### Install Prometheus Stack
Install Prometheus stack if you haven't done it already. You can use [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) which installs the necessary components required for the PerconaXtraDB Grafana dashboards.

### Install Panopticon
KubeDB Enterprise License works for Panopticon too. So, we will use the same license that we have already obtained.

```bash
$ helm install panopticon appscode/panopticon -n kubeops \
    --create-namespace \
    --set monitoring.enabled=true \
    --set monitoring.agent=prometheus.io/operator \
    --set monitoring.serviceMonitor.labels.release=latest \
    --set-file license=/path/to/license.txt
```
Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=panopticon"
NAMESPACE   NAME                          READY   STATUS    RESTARTS      AGE
kubeops     panopticon-5bd695c8f4-g48r2   1/1     Running   0             5m
```


## Deploy Percona XtraDB Clustered Database

Now, we are going to Deploy Percona XtraDB with monitoring enabled using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of the Percona XtraDB CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: percona-cluster
  namespace: demo
spec:
  version: "8.0.28"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: latest
        interval: 10s
  terminationPolicy: WipeOut
```

Let's save this yaml configuration into `percona-cluster.yaml` 
Then create the above Percona XtraDB CRO

```bash
$ kubectl apply -f percona-cluster.yaml 
perconaxtradb.kubedb.com/percona-cluster created
```

* In this yaml we can see in the `spec.version` field specifies the version of Percona XtraDB. Here, we are using Percona XtraDB `version 8.0.28`. You can list the KubeDB supported versions of Percona XtraDB by running `$ kubectl get perconaxtradbversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/percona-xtradb/concepts/percona-xtradb/#specterminationpolicy).

Once these are handled correctly and the Percona XtraDB object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=percona-cluster'
NAME                    READY   STATUS    RESTARTS      AGE
pod/percona-cluster-0   3/3     Running   0             5m
pod/percona-cluster-1   3/3     Running   0             5m
pod/percona-cluster-2   3/3     Running   0             5m

NAME                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/percona-cluster         ClusterIP   10.96.240.62   <none>        3306/TCP    5m
service/percona-cluster-pods    ClusterIP   None           <none>        3306/TCP    5m
service/percona-cluster-stats   ClusterIP   10.96.169.48   <none>        56790/TCP   5m

NAME                               READY   AGE
statefulset.apps/percona-cluster   3/3     5m

NAME                                                 TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/percona-cluster   kubedb.com/perconaxtradb   8.0.28    5m
```
Let’s check if the database is ready to use,

```bash
$ kubectl get perconaxtradb -n demo percona-cluster
NAME              VERSION   STATUS   AGE
percona-cluster   8.0.28    Ready    5m
```
> We have successfully deployed Percona XtraDB in GKE.


### Create DB Metrics Configurations

First, you have to create a `MetricsConfiguration` object for database. This `MetricsConfiguration` object is used by Panopticon to generate metrics for DB instances.
Install `kubedb-metrics` charts which will create the `MetricsConfiguration` object for DB:

```bash
$ helm search repo appscode/kubedb-metrics --version=v2023.02.28
$ helm install kubedb-metrics appscode/kubedb-metrics -n kubedb --version=v2023.02.28
```

### Import Grafana Dashboard
Here, we will port-forward the `latest-grafana` service to access Grafana Dashboard from UI.

```bash
$ kubectl get service -n default
NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None            <none>        9093/TCP,9094/TCP,9094/UDP   4h56m
kubernetes                                ClusterIP   10.96.0.1       <none>        443/TCP                      4h57m
latest-grafana                            ClusterIP   10.96.194.250   <none>        80/TCP                       4h56m
latest-kube-prometheus-sta-alertmanager   ClusterIP   10.96.44.231    <none>        9093/TCP                     4h56m
latest-kube-prometheus-sta-operator       ClusterIP   10.96.168.58    <none>        443/TCP                      4h56m
latest-kube-prometheus-sta-prometheus     ClusterIP   10.96.153.36    <none>        9090/TCP                     4h56m
latest-kube-state-metrics                 ClusterIP   10.96.240.13    <none>        8080/TCP                     4h56m
latest-prometheus-node-exporter           ClusterIP   10.96.9.50      <none>        9100/TCP                     4h56m
prometheus-operated                       ClusterIP   None            <none>        9090/TCP                     4h56m
```
To access Grafana UI Let's port-forward `latest-grafana` service to 3063 

```bash
$ kubectl port-forward -n default service/latest-grafana 3063:80
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


For PerconaXtraDB Cluster use [PerconaXtraDB Cluster Json](https://github.com/appscode/grafana-dashboards/blob/master/perconaxtradb/perconaxtradb-database.json)

For PerconaXtraDB Galera Cluster use [PerconaXtraDB Galera Cluster Json](https://github.com/appscode/grafana-dashboards/blob/master/perconaxtradb/perconaxtradb-galera.json)

For PerconaXtraDB Summary use [PerconaXtraDB Summary Json](https://github.com/appscode/grafana-dashboards/blob/master/perconaxtradb/perconaxtradb-summary.json)

For PerconaXtraDB Pod use [PerconaXtraDB Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/perconaxtradb/perconaxtradb-pod.json)

If you followed above instruction properly you will see PerconaXtraDB Grafana Dashboards in your Grafana UI

Here are some screenshots of our PerconaXtraDB deployment. You can visualize every single component supported by Grafana, checkout here for more about [Grafana Dashboard](https://grafana.com/docs/grafana/latest/)

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


We have made an in depth tutorial on Managing Percona XtraDB Cluster Day-2 Operations by using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/PsMbpDHg_oU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Percona XtraDB in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-percona-xtradb-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
