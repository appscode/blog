---
title: Monitor ProxySQL with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2023-07-19"
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
- proxysql
- s3
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MySQL, MongoDB, MariaDB, Kafka, Elasticsearch, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). And Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.
In this tutorial we will Monitor ProxySQL with Grafana Dashboard in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install Prometheus Stack
3) Install Panopticon
4) Deploy MySQL
5) Deploy ProxySQL Cluster
6) Monitor with Grafana Dashboard

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
appscode/kubedb-ops-manager       	v0.21.0      	v0.21.2    	KubeDB Ops Manager by AppsCode - Enterprise fea...
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
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-5b8b948645-n5jhk       1/1     Running   0          81s
kubedb      kubedb-kubedb-dashboard-54b55ffcf6-7sbsf        1/1     Running   0          81s
kubedb      kubedb-kubedb-ops-manager-8449f79bb8-27szq      1/1     Running   0          81s
kubedb      kubedb-kubedb-provisioner-fdf4c96f7-qmtt2       1/1     Running   0          81s
kubedb      kubedb-kubedb-schema-manager-695c94f9c7-fxnkz   1/1     Running   0          81s
kubedb      kubedb-kubedb-webhook-server-8647d646d9-jj4x7   1/1     Running   0          81s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-07-19T06:23:51Z
elasticsearchdashboards.dashboard.kubedb.com      2023-07-19T06:23:53Z
elasticsearches.kubedb.com                        2023-07-19T06:23:50Z
elasticsearchopsrequests.ops.kubedb.com           2023-07-19T06:23:50Z
elasticsearchversions.catalog.kubedb.com          2023-07-19T06:21:00Z
etcds.kubedb.com                                  2023-07-19T06:23:56Z
etcdversions.catalog.kubedb.com                   2023-07-19T06:21:01Z
kafkas.kubedb.com                                 2023-07-19T06:24:08Z
kafkaversions.catalog.kubedb.com                  2023-07-19T06:21:01Z
mariadbautoscalers.autoscaling.kubedb.com         2023-07-19T06:23:51Z
mariadbdatabases.schema.kubedb.com                2023-07-19T06:24:02Z
mariadbopsrequests.ops.kubedb.com                 2023-07-19T06:24:11Z
mariadbs.kubedb.com                               2023-07-19T06:23:57Z
mariadbversions.catalog.kubedb.com                2023-07-19T06:21:01Z
memcacheds.kubedb.com                             2023-07-19T06:23:57Z
memcachedversions.catalog.kubedb.com              2023-07-19T06:21:02Z
mongodbautoscalers.autoscaling.kubedb.com         2023-07-19T06:23:51Z
mongodbdatabases.schema.kubedb.com                2023-07-19T06:23:55Z
mongodbopsrequests.ops.kubedb.com                 2023-07-19T06:23:53Z
mongodbs.kubedb.com                               2023-07-19T06:23:54Z
mongodbversions.catalog.kubedb.com                2023-07-19T06:21:02Z
mysqlautoscalers.autoscaling.kubedb.com           2023-07-19T06:23:51Z
mysqldatabases.schema.kubedb.com                  2023-07-19T06:23:54Z
mysqlopsrequests.ops.kubedb.com                   2023-07-19T06:24:07Z
mysqls.kubedb.com                                 2023-07-19T06:23:54Z
mysqlversions.catalog.kubedb.com                  2023-07-19T06:21:02Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-07-19T06:23:52Z
perconaxtradbopsrequests.ops.kubedb.com           2023-07-19T06:24:25Z
perconaxtradbs.kubedb.com                         2023-07-19T06:24:04Z
perconaxtradbversions.catalog.kubedb.com          2023-07-19T06:21:02Z
pgbouncers.kubedb.com                             2023-07-19T06:24:00Z
pgbouncerversions.catalog.kubedb.com              2023-07-19T06:21:03Z
postgresautoscalers.autoscaling.kubedb.com        2023-07-19T06:23:52Z
postgresdatabases.schema.kubedb.com               2023-07-19T06:24:00Z
postgreses.kubedb.com                             2023-07-19T06:24:01Z
postgresopsrequests.ops.kubedb.com                2023-07-19T06:24:18Z
postgresversions.catalog.kubedb.com               2023-07-19T06:21:03Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-07-19T06:23:52Z
proxysqlopsrequests.ops.kubedb.com                2023-07-19T06:24:21Z
proxysqls.kubedb.com                              2023-07-19T06:24:06Z
proxysqlversions.catalog.kubedb.com               2023-07-19T06:21:03Z
publishers.postgres.kubedb.com                    2023-07-19T06:24:35Z
redisautoscalers.autoscaling.kubedb.com           2023-07-19T06:23:52Z
redises.kubedb.com                                2023-07-19T06:24:07Z
redisopsrequests.ops.kubedb.com                   2023-07-19T06:24:14Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-07-19T06:23:52Z
redissentinelopsrequests.ops.kubedb.com           2023-07-19T06:24:28Z
redissentinels.kubedb.com                         2023-07-19T06:24:07Z
redisversions.catalog.kubedb.com                  2023-07-19T06:21:04Z
subscribers.postgres.kubedb.com                   2023-07-19T06:24:38Z
```

### Install Prometheus Stack
Install Prometheus stack which installs the necessary components required for the ProxySQL Grafana dashboards. You can use following commands,

```bash
$ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
$ helm install prometheus prometheus-community/kube-prometheus-stack
```
or visit [kube-prometheus-stack](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack) for more detailed information.


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
NAMESPACE   NAME                         READY   STATUS    RESTARTS   AGE
kubeops     panopticon-6f98b7c9f-5mzv5   1/1     Running   0          46s
```

## Deploy MySQL Group Replication

Now, we are going to Deploy MySQL Group Replication using KubeDB. First, let’s create a Namespace in which we will deploy the server.

```bash
$ kubectl create namespace demo
namespace/demo created
```
Here is the yaml of the MySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.0.32"
  replicas: 3
  topology:
    mode: GroupReplication
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

Let's save this yaml configuration into `mysql-server.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl apply -f mysql-server.yaml
mysql.kubedb.com/mysql-server created
```

* In this yaml we can see in the `spec.version` field specifies the version of MySQL. Here, we are using MySQL `version 8.0.32`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
* `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
* `spec.monitor.prometheus.serviceMonitor.labels` specifies the release name that KubeDB should use in `ServiceMonitor`.
* `spec.monitor.prometheus.interval` defines that the Prometheus server should scrape metrics from this database with 10 seconds interval.
* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy)

Let’s check if the server is ready to use,
```bash
$ kubectl get mysql -n demo mysql-server
NAME           VERSION   STATUS   AGE
mysql-server   8.0.32    Ready    2m46s
```
## Deploy ProxySQL Cluster

Now, we are going to Deploy ProxySQL cluster using KubeDB. Here is the yaml of the ProxySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 3
  mode: GroupReplication
  backend:
      name: mysql-server
  syncUsers: true
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `proxysql-server.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl apply -f proxysql-server.yaml
mysql.kubedb.com/proxysql-server created
```

In this yaml,

* `spec.version` field specifies the version of ProxySQL. Here, we are using ProxySQL `2.4.4-debian`. You can list the KubeDB supported versions of ProxySQL by running `$ kubectl get proxysqlversions` command.
* `spec.backend.name` contains the name of MySQL server backend which is `mysql-server` in this case.
* `spec.syncUsers` confirms that the ProxySQL will sync it’s user list with MySQL server or not.

* And the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”.


Let’s check if the server is ready to use,
```bash
$ kubectl get proxysql -n demo proxy-server
NAME           VERSION        STATUS   AGE
proxy-server   2.4.4-debian   Ready    2m34s
```

Once these are handled correctly, you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME                 READY   STATUS    RESTARTS   AGE
pod/mysql-server-0   2/2     Running   0          8m25s
pod/mysql-server-1   2/2     Running   0          7m57s
pod/mysql-server-2   2/2     Running   0          7m24s
pod/proxy-server-0   1/1     Running   0          2m58s
pod/proxy-server-1   1/1     Running   0          2m46s
pod/proxy-server-2   1/1     Running   0          2m35s

NAME                           TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
service/mysql-server           ClusterIP   10.100.129.133   <none>        3306/TCP            8m27s
service/mysql-server-pods      ClusterIP   None             <none>        3306/TCP            8m27s
service/mysql-server-standby   ClusterIP   10.100.240.71    <none>        3306/TCP            8m27s
service/proxy-server           ClusterIP   10.100.212.21    <none>        6033/TCP            2m59s
service/proxy-server-pods      ClusterIP   None             <none>        6032/TCP,6033/TCP   2m59s
service/proxy-server-stats     ClusterIP   10.100.179.96    <none>        6070/TCP            2m26s

NAME                            READY   AGE
statefulset.apps/mysql-server   3/3     8m27s
statefulset.apps/proxy-server   3/3     3m

NAME                                              TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/mysql-server   kubedb.com/mysql   8.0.32    8m28s

NAME                            VERSION   STATUS   AGE
mysql.kubedb.com/mysql-server   8.0.32    Ready    8m34s

NAME                               VERSION        STATUS   AGE
proxysql.kubedb.com/proxy-server   2.4.4-debian   Ready    3m9s
```
> We have successfully deployed ProxySQL in AWS.


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
NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None            <none>        9093/TCP,9094/TCP,9094/UDP   17m
kubernetes                                ClusterIP   10.100.0.1      <none>        443/TCP                      45m
prometheus-grafana                        ClusterIP   10.100.20.62    <none>        80/TCP                       17m
prometheus-kube-prometheus-alertmanager   ClusterIP   10.100.0.56     <none>        9093/TCP,8080/TCP            17m
prometheus-kube-prometheus-operator       ClusterIP   10.100.152.40   <none>        443/TCP                      17m
prometheus-kube-prometheus-prometheus     ClusterIP   10.100.99.210   <none>        9090/TCP,8080/TCP            17m
prometheus-kube-state-metrics             ClusterIP   10.100.32.203   <none>        8080/TCP                     17m
prometheus-operated                       ClusterIP   None            <none>        9090/TCP                     17m
prometheus-prometheus-node-exporter       ClusterIP   10.100.85.153   <none>        9100/TCP                     17m
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


For ProxySQL Database Dashboard use [ProxySQL Database Dashboard](https://github.com/appscode/grafana-dashboards/blob/master/proxysql/proxysql-database.json)

For ProxySQL Pod use [ProxySQL Pod Json](https://github.com/appscode/grafana-dashboards/blob/master/proxysql/proxysql-pod.json)

For ProxySQL Summary use [ProxySQL Summary Json](https://github.com/appscode/grafana-dashboards/blob/master/proxysql/proxysql-summary.json)



If you followed above instruction properly you will see ProxySQL Grafana Dashboards in your Grafana UI

Here are some screenshots of our ProxySQL deployment. You can visualize every single component supported by Grafana, checkout here for more about [Grafana Dashboard](https://grafana.com/docs/grafana/latest/)

![Sample UI 1](sample-ui-1.png)

![Sample UI 2](sample-ui-2.png)

![Sample UI 3](sample-ui-3.png)


We have made an in depth tutorial on ProxySQL Declarative Provisioning, Reconfiguration and Horizontal Scaling using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/fT_cQDxfU9o" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [ProxySQL in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-proxysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
