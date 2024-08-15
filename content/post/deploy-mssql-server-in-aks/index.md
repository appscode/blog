---
title: Deploy Microsoft SQL Server (MSSQL) in Azure Kubernetes Service (AKS)
date: "2024-08-14"
weight: 14
authors:
- Dipta Roy
tags:
- aks
- azure
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- microsoft-azure
- microsoft-sql
- mssql
- mssql-cluster
- mssql-database
- mssql-server
- sqlserver
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Solr, Microsoft SQL Server, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Microsoft SQL Server (MSSQL) in Azure Kubernetes Service (AKS). We will cover the following steps:

1. Install KubeDB
2. Deploy MSSQL Server Availability Group Cluster
3. Read/Write Sample Data

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
NAME                              	CHART VERSION	APP VERSION	DESCRIPTION                                       
appscode/kubedb                   	v2024.6.4    	v2024.6.4  	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.31.0      	v0.31.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.6.4    	v2024.6.4  	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.1.0       	v0.1.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.6.4    	v2024.6.4  	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.22.0      	v0.22.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.6.4    	v2024.6.4  	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.6.4    	v2024.6.4  	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.6.4    	v2024.6.4  	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.33.0      	v0.33.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.6.4    	v2024.6.4  	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.6.4    	v0.8.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.46.0      	v0.46.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.22.0      	v0.22.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.7.4    	0.7.3      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.7.4    	v2024.7.4  	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.22.0      	v0.22.0    	KubeDB Webhook Server by AppsCode  


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.6.4 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.MSSQLServer=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-7cb959fccf-xsqjs       1/1     Running   0          102s
kubedb      kubedb-kubedb-ops-manager-854649d6df-cd7f4      1/1     Running   0          102s
kubedb      kubedb-kubedb-provisioner-5685bc4c99-sddrc      1/1     Running   0          103s
kubedb      kubedb-kubedb-webhook-server-64fc6759dc-gkw4l   1/1     Running   0          103s
kubedb      kubedb-petset-operator-77b6b9897f-8pzpg         1/1     Running   0          103s
kubedb      kubedb-petset-webhook-server-7988778876-shvmz   2/2     Running   0          103s
kubedb      kubedb-sidekick-c898cff4c-mcljb                 1/1     Running   0          103s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-08-14T05:32:17Z
connectclusters.kafka.kubedb.com                   2024-08-14T05:33:07Z
connectors.kafka.kubedb.com                        2024-08-14T05:33:07Z
druidversions.catalog.kubedb.com                   2024-08-14T05:32:17Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-08-14T05:33:03Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-08-14T05:33:03Z
elasticsearches.kubedb.com                         2024-08-14T05:33:03Z
elasticsearchopsrequests.ops.kubedb.com            2024-08-14T05:33:03Z
elasticsearchversions.catalog.kubedb.com           2024-08-14T05:32:17Z
etcdversions.catalog.kubedb.com                    2024-08-14T05:32:17Z
ferretdbversions.catalog.kubedb.com                2024-08-14T05:32:17Z
kafkaautoscalers.autoscaling.kubedb.com            2024-08-14T05:33:07Z
kafkaconnectorversions.catalog.kubedb.com          2024-08-14T05:32:17Z
kafkaopsrequests.ops.kubedb.com                    2024-08-14T05:33:07Z
kafkas.kubedb.com                                  2024-08-14T05:33:06Z
kafkaversions.catalog.kubedb.com                   2024-08-14T05:32:17Z
mariadbarchivers.archiver.kubedb.com               2024-08-14T05:33:11Z
mariadbautoscalers.autoscaling.kubedb.com          2024-08-14T05:33:11Z
mariadbdatabases.schema.kubedb.com                 2024-08-14T05:33:11Z
mariadbopsrequests.ops.kubedb.com                  2024-08-14T05:33:11Z
mariadbs.kubedb.com                                2024-08-14T05:33:10Z
mariadbversions.catalog.kubedb.com                 2024-08-14T05:32:17Z
memcachedversions.catalog.kubedb.com               2024-08-14T05:32:18Z
mongodbarchivers.archiver.kubedb.com               2024-08-14T05:33:16Z
mongodbautoscalers.autoscaling.kubedb.com          2024-08-14T05:33:15Z
mongodbdatabases.schema.kubedb.com                 2024-08-14T05:33:16Z
mongodbopsrequests.ops.kubedb.com                  2024-08-14T05:33:15Z
mongodbs.kubedb.com                                2024-08-14T05:33:15Z
mongodbversions.catalog.kubedb.com                 2024-08-14T05:32:18Z
mssqlservers.kubedb.com                            2024-08-14T05:33:19Z
mssqlserverversions.catalog.kubedb.com             2024-08-14T05:32:18Z
mysqlarchivers.archiver.kubedb.com                 2024-08-14T05:33:23Z
mysqlautoscalers.autoscaling.kubedb.com            2024-08-14T05:33:23Z
mysqldatabases.schema.kubedb.com                   2024-08-14T05:33:23Z
mysqlopsrequests.ops.kubedb.com                    2024-08-14T05:33:23Z
mysqls.kubedb.com                                  2024-08-14T05:33:23Z
mysqlversions.catalog.kubedb.com                   2024-08-14T05:32:18Z
perconaxtradbversions.catalog.kubedb.com           2024-08-14T05:32:18Z
pgbouncerversions.catalog.kubedb.com               2024-08-14T05:32:18Z
pgpoolversions.catalog.kubedb.com                  2024-08-14T05:32:18Z
postgresarchivers.archiver.kubedb.com              2024-08-14T05:33:28Z
postgresautoscalers.autoscaling.kubedb.com         2024-08-14T05:33:27Z
postgresdatabases.schema.kubedb.com                2024-08-14T05:33:28Z
postgreses.kubedb.com                              2024-08-14T05:33:27Z
postgresopsrequests.ops.kubedb.com                 2024-08-14T05:33:27Z
postgresversions.catalog.kubedb.com                2024-08-14T05:32:18Z
proxysqlversions.catalog.kubedb.com                2024-08-14T05:32:18Z
publishers.postgres.kubedb.com                     2024-08-14T05:33:28Z
rabbitmqversions.catalog.kubedb.com                2024-08-14T05:32:18Z
redisautoscalers.autoscaling.kubedb.com            2024-08-14T05:33:31Z
redises.kubedb.com                                 2024-08-14T05:33:31Z
redisopsrequests.ops.kubedb.com                    2024-08-14T05:33:31Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-08-14T05:33:32Z
redissentinelopsrequests.ops.kubedb.com            2024-08-14T05:33:32Z
redissentinels.kubedb.com                          2024-08-14T05:33:31Z
redisversions.catalog.kubedb.com                   2024-08-14T05:32:18Z
schemaregistries.kafka.kubedb.com                  2024-08-14T05:33:07Z
schemaregistryversions.catalog.kubedb.com          2024-08-14T05:32:18Z
singlestoreversions.catalog.kubedb.com             2024-08-14T05:32:18Z
solrversions.catalog.kubedb.com                    2024-08-14T05:32:18Z
subscribers.postgres.kubedb.com                    2024-08-14T05:33:28Z
zookeeperversions.catalog.kubedb.com               2024-08-14T05:32:18Z
```


### Create a Namespace

To keep resources isolated, we'll use a separate namespace called `demo` throughout this tutorial.
Run the following command to create the namespace:

```bash
$ kubectl create namespace demo
namespace/demo created
```


### Install Cert Manager

To manage TLS certificates within Kubernetes, we need to install [Cert Manager](https://cert-manager.io/). Follow these steps to install Cert Manager using the YAML manifest:

```bash
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.0/cert-manager.yaml
namespace/cert-manager created
...
```

### Create Issuer

Next, we need to create an Issuer, which will be used to generate certificates for TLS settings and internal endpoint authentication of availability group replicas.

Start by generating CA certificates using OpenSSL:
```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=KubeDB"
Generating a RSA private key
..........................................................................+++++
..+++++
writing new private key to './ca.key'
-----
```

Create a Kubernetes Secret to store the CA certificate:
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo
secret/mssqlserver-ca created
```

Create an Issuer using the CA certificate stored in the `mssqlserver-ca` Secret. Below is the YAML definition for the Issuer:
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssqlserver-issuer
  namespace: demo
spec:
  ca:
    secretName: mssqlserver-ca
```


Below is the YAML definition for the Issuer:
```bash
$ kubectl apply -f issuer.yaml
issuer.cert-manager.io/mssqlserver-issuer created
```

## Deploy MSSQL Server Availability Group Cluster

Now, we can deploy an MSSQL Server Availability Group Cluster using the following YAML definition:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - music
  internalAuth:
    endpointCert:
      issuerRef:
        apiGroup: cert-manager.io
        name: mssqlserver-issuer
        kind: Issuer
  tls:
    issuerRef:
      name: mssqlserver-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  storageType: Durable
  storage:
    storageClassName: "default"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's save this yaml configuration into `mssqlserver-ag-cluster.yaml` 
Then apply the above MSSQL Server yaml,

```bash
$ kubectl apply -f mssqlserver-ag-cluster.yaml 
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

In this yaml,
- `spec.version` is the name of the MSSQLServerVersion CR. Here, a MSSQLServer of `version 2022-cu12` will be created.
- `spec.topology.availabilityGroup.databases` specifies the list of databases that must be added to the `Availability Group` cluster.
- `spec.tls` specifies the TLS/SSL configurations. The KubeDB operator supports TLS management by using the [cert-manager](https://cert-manager.io/). Here `tls.clientTLS: false` means tls will not be enabled for SQL Server but the Issuer will be used to configure tls enabled wal-g proxy-server which is required for SQL Server backup operation.
- `spec.storageType` specifies the type of storage that will be used for MSSQLServer database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the MSSQLServer database using `EmptyDir` volume. In this case, you don’t have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the Petset created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don’t specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.deletionPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”.

Once these are handled correctly and the MSSQLServer object is deployed, you will see that the following resources are created:

```bash
$ kubectl get all -n demo
NAME                           READY   STATUS    RESTARTS   AGE
pod/mssqlserver-ag-cluster-0   2/2     Running   0          6m35s
pod/mssqlserver-ag-cluster-1   2/2     Running   0          5m30s
pod/mssqlserver-ag-cluster-2   2/2     Running   0          4m7s

NAME                                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/mssqlserver-ag-cluster             ClusterIP   10.128.113.192   <none>        1433/TCP   7m41s
service/mssqlserver-ag-cluster-pods        ClusterIP   None             <none>        1433/TCP   7m41s
service/mssqlserver-ag-cluster-secondary   ClusterIP   10.128.213.243   <none>        1433/TCP   7m41s

NAME                                                        TYPE                     VERSION   AGE
appbinding.appcatalog.appscode.com/mssqlserver-ag-cluster   kubedb.com/mssqlserver   2022      6m37s

NAME                                            VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssqlserver-ag-cluster   2022-cu12   Ready    7m41s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mssqlserver -n demo mssqlserver-ag-cluster
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    8m16s
```
> We have successfully deployed MSSQL Server in AKS. Now we can exec into the container to use the database.


### Accessing Database Through CLI

To access your database through the CLI, you first need the credentials for the database. KubeDB will create several Kubernetes Secrets and Services for your MSSQL Server instance. To view them, use the following commands:

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mssqlserver-ag-cluster 
NAME                                   TYPE                       DATA   AGE
mssqlserver-ag-cluster-auth            kubernetes.io/basic-auth   2      9m36s
mssqlserver-ag-cluster-client-cert     kubernetes.io/tls          3      9m36s
mssqlserver-ag-cluster-dbm-login       kubernetes.io/basic-auth   1      9m36s
mssqlserver-ag-cluster-endpoint-cert   kubernetes.io/tls          3      9m36s
mssqlserver-ag-cluster-master-key      kubernetes.io/basic-auth   1      9m36s
mssqlserver-ag-cluster-server-cert     kubernetes.io/tls          3      9m36s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mssqlserver-ag-cluster 
NAME                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
mssqlserver-ag-cluster             ClusterIP   10.128.113.192   <none>        1433/TCP   9m43s
mssqlserver-ag-cluster-pods        ClusterIP   None             <none>        1433/TCP   9m43s
mssqlserver-ag-cluster-secondary   ClusterIP   10.128.213.243   <none>        1433/TCP   9m43s
```

From the above list, the `mssqlserver-ag-cluster-auth` Secret contains the admin-level credentials needed to connect to the database. Use the following commands to obtain the username and password:

```bash
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth  -o jsonpath='{.data.username}' | base64 -d
sa
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth  -o jsonpath='{.data.password}' | base64 -d
pT42K58uKEe3pote
```


### Insert Sample Data

In this section, we will insert sample data into our MSSQL Server deployed on Kubernetes. Before we can insert data, we need to identify the primary node, as data writes are only permitted on the primary node.

To determine which pod is the primary node, run the following command to list the pods along with their roles:

```bash
$ kubectl get pods -n demo --selector=app.kubernetes.io/instance=mssqlserver-ag-cluster -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\n"}{end}'

mssqlserver-ag-cluster-0	primary
mssqlserver-ag-cluster-1	secondary
mssqlserver-ag-cluster-2	secondary
```

From the output above, we can see that `mssqlserver-ag-cluster-0` is the primary node. To insert data, log into the primary MSSQL Server pod. Use the following command,

```bash
$ kubectl exec -it mssqlserver-ag-cluster-0 -n demo bash
Defaulted container "mssql" out of: mssql, mssql-coordinator, mssql-init (init)
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "pT42K58uKEe3pote"

1> SELECT name FROM sys.databases
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
music                                                                                                                           
kubedb_system                                                                                                                   

(6 rows affected)

1> SELECT database_name
2> FROM sys.availability_databases_cluster
3> GO
database_name                                                                                                                   
--------------------------------------------------------------------------------------------------------------------------------
music                                                                                                                           

(1 rows affected)

1> USE music
2> GO
Changed database context to 'music'.

1> CREATE TABLE Playlist (Artist NVARCHAR(255), Song NVARCHAR(255));
2> GO
1> INSERT INTO Playlist(Artist, Song) VALUES ('John Denver', 'Country Roads');
2> GO

(1 rows affected)
1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
John Denver                                                                     Country Roads

(1 rows affected)
1> exit

...

# Confirm that the data inserted into the primary node has been replicated to the secondary nodes.
# Access the secondary node (Node 2) to verify that the data is present.
$ kubectl exec -it mssqlserver-ag-cluster-1 -n demo bash
Defaulted container "mssql" out of: mssql, mssql-coordinator, mssql-init (init)
mssql@mssqlserver-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "pT42K58uKEe3pote"

1> USE music
2> GO
Changed database context to 'music'.

1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
John Denver                                                                     Country Roads

(1 rows affected)
1> exit

...

# Access the secondary node (Node 3) to verify that the data is present.
$ kubectl exec -it mssqlserver-ag-cluster-2 -n demo bash
Defaulted container "mssql" out of: mssql, mssql-coordinator, mssql-init (init)
mssql@mssqlserver-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "pT42K58uKEe3pote"

1> USE music
2> GO
Changed database context to 'music'.

1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
John Denver                                                                     Country Roads

(1 rows affected)
1> exit
```

> We’ve successfully inserted some sample data to our database. Also, access it from every node of Availability Group Cluster. More information about Deploy & Manage MSSQL Server on Kubernetes can be found in [Kubernetes MSSQL Server](https://kubedb.com/kubernetes/databases/run-and-manage-mssqlserver-on-kubernetes/)

We have made a in depth tutorial on Seamlessly Provision and Manage Microsoft SQL Server Instances on Kubernetes Using KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/Hb_fhCVuVhQ?si=mkP-GWyWBS4DR4Y9" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [MSSQL Server on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mssqlserver-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
