---
title: Deploy Microsoft SQL Server (MSSQL) in Google Kubernetes Engine (GKE)
date: "2024-09-09"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- gcp
- gke
- kubedb
- kubernetes
- microsoft-sql
- mssql
- mssql-cluster
- mssql-database
- mssql-server
- sqlserver
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Solr, Microsoft SQL Server, FerretDB, SingleStore, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, Pgpool, ZooKeeper and the streaming platform Kafka, RabbitMQ. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Microsoft SQL Server (MSSQL) in Google Kubernetes Engine (GKE). We will cover the following steps:

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
appscode/kubedb                   	v2024.8.21   	v2024.8.21 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.32.0      	v0.32.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.8.21   	v2024.8.21 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.2.0       	v0.2.0     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.8.21   	v2024.8.21 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.23.0      	v0.23.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.8.21   	v2024.8.21 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.8.21   	v2024.8.21 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.8.21   	v2024.8.21 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.34.0      	v0.34.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.8.21   	v2024.8.21 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.8.21   	v0.9.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.47.0      	v0.47.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.23.0      	v0.23.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.8.21   	0.7.5      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-presets        	v2024.8.21   	v2024.8.21 	KubeDB UI Presets                                 
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.23.0      	v0.23.0    	KubeDB Webhook Server by AppsCode   


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.8.21 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.MSSQLServer=true \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-76f47cb964-ckz7c       1/1     Running   0          3m56s
kubedb      kubedb-kubedb-ops-manager-69b5bfdc4d-lm7jv      1/1     Running   0          3m56s
kubedb      kubedb-kubedb-provisioner-5dd9c86655-j7ph7      1/1     Running   0          3m56s
kubedb      kubedb-kubedb-webhook-server-867b8cf8c4-nmgpc   1/1     Running   0          3m56s
kubedb      kubedb-petset-operator-77b6b9897f-cjgtc         1/1     Running   0          3m56s
kubedb      kubedb-petset-webhook-server-556b48c68b-sc6fw   2/2     Running   0          3m56s
kubedb      kubedb-sidekick-c898cff4c-4q58x                 1/1     Running   0          3m56s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
clickhouseversions.catalog.kubedb.com              2024-09-09T12:13:09Z
connectclusters.kafka.kubedb.com                   2024-09-09T12:14:04Z
connectors.kafka.kubedb.com                        2024-09-09T12:14:04Z
druidversions.catalog.kubedb.com                   2024-09-09T12:13:09Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-09-09T12:14:00Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-09-09T12:14:00Z
elasticsearches.kubedb.com                         2024-09-09T12:14:00Z
elasticsearchopsrequests.ops.kubedb.com            2024-09-09T12:14:00Z
elasticsearchversions.catalog.kubedb.com           2024-09-09T12:13:09Z
etcdversions.catalog.kubedb.com                    2024-09-09T12:13:09Z
ferretdbversions.catalog.kubedb.com                2024-09-09T12:13:09Z
kafkaautoscalers.autoscaling.kubedb.com            2024-09-09T12:14:04Z
kafkaconnectorversions.catalog.kubedb.com          2024-09-09T12:13:09Z
kafkaopsrequests.ops.kubedb.com                    2024-09-09T12:14:04Z
kafkas.kubedb.com                                  2024-09-09T12:14:04Z
kafkaversions.catalog.kubedb.com                   2024-09-09T12:13:09Z
mariadbarchivers.archiver.kubedb.com               2024-09-09T12:14:07Z
mariadbautoscalers.autoscaling.kubedb.com          2024-09-09T12:14:07Z
mariadbdatabases.schema.kubedb.com                 2024-09-09T12:14:07Z
mariadbopsrequests.ops.kubedb.com                  2024-09-09T12:14:07Z
mariadbs.kubedb.com                                2024-09-09T12:14:07Z
mariadbversions.catalog.kubedb.com                 2024-09-09T12:13:09Z
memcachedversions.catalog.kubedb.com               2024-09-09T12:13:09Z
mongodbarchivers.archiver.kubedb.com               2024-09-09T12:14:11Z
mongodbautoscalers.autoscaling.kubedb.com          2024-09-09T12:14:11Z
mongodbdatabases.schema.kubedb.com                 2024-09-09T12:14:11Z
mongodbopsrequests.ops.kubedb.com                  2024-09-09T12:14:11Z
mongodbs.kubedb.com                                2024-09-09T12:14:11Z
mongodbversions.catalog.kubedb.com                 2024-09-09T12:13:09Z
mssqlserverarchivers.archiver.kubedb.com           2024-09-09T12:14:15Z
mssqlserverautoscalers.autoscaling.kubedb.com      2024-09-09T12:14:15Z
mssqlservers.kubedb.com                            2024-09-09T12:14:14Z
mssqlserverversions.catalog.kubedb.com             2024-09-09T12:13:09Z
mysqlarchivers.archiver.kubedb.com                 2024-09-09T12:14:18Z
mysqlautoscalers.autoscaling.kubedb.com            2024-09-09T12:14:18Z
mysqldatabases.schema.kubedb.com                   2024-09-09T12:14:18Z
mysqlopsrequests.ops.kubedb.com                    2024-09-09T12:14:18Z
mysqls.kubedb.com                                  2024-09-09T12:14:18Z
mysqlversions.catalog.kubedb.com                   2024-09-09T12:13:09Z
perconaxtradbversions.catalog.kubedb.com           2024-09-09T12:13:10Z
pgbouncerversions.catalog.kubedb.com               2024-09-09T12:13:10Z
pgpoolversions.catalog.kubedb.com                  2024-09-09T12:13:10Z
postgresarchivers.archiver.kubedb.com              2024-09-09T12:14:22Z
postgresautoscalers.autoscaling.kubedb.com         2024-09-09T12:14:22Z
postgresdatabases.schema.kubedb.com                2024-09-09T12:14:22Z
postgreses.kubedb.com                              2024-09-09T12:14:21Z
postgresopsrequests.ops.kubedb.com                 2024-09-09T12:14:22Z
postgresversions.catalog.kubedb.com                2024-09-09T12:13:10Z
proxysqlversions.catalog.kubedb.com                2024-09-09T12:13:10Z
publishers.postgres.kubedb.com                     2024-09-09T12:14:22Z
rabbitmqversions.catalog.kubedb.com                2024-09-09T12:13:10Z
redisautoscalers.autoscaling.kubedb.com            2024-09-09T12:14:25Z
redises.kubedb.com                                 2024-09-09T12:14:25Z
redisopsrequests.ops.kubedb.com                    2024-09-09T12:14:25Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-09-09T12:14:25Z
redissentinelopsrequests.ops.kubedb.com            2024-09-09T12:14:25Z
redissentinels.kubedb.com                          2024-09-09T12:14:25Z
redisversions.catalog.kubedb.com                   2024-09-09T12:13:10Z
restproxies.kafka.kubedb.com                       2024-09-09T12:14:04Z
schemaregistries.kafka.kubedb.com                  2024-09-09T12:14:04Z
schemaregistryversions.catalog.kubedb.com          2024-09-09T12:13:10Z
singlestoreversions.catalog.kubedb.com             2024-09-09T12:13:10Z
solrversions.catalog.kubedb.com                    2024-09-09T12:13:10Z
subscribers.postgres.kubedb.com                    2024-09-09T12:14:22Z
zookeeperversions.catalog.kubedb.com               2024-09-09T12:13:10Z
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
    storageClassName: "standard"
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
pod/mssqlserver-ag-cluster-0   2/2     Running   0          5m29s
pod/mssqlserver-ag-cluster-1   2/2     Running   0          4m24s
pod/mssqlserver-ag-cluster-2   2/2     Running   0          3m1s

NAME                                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/mssqlserver-ag-cluster             ClusterIP   10.128.113.192   <none>        1433/TCP   6m35s
service/mssqlserver-ag-cluster-pods        ClusterIP   None             <none>        1433/TCP   6m35s
service/mssqlserver-ag-cluster-secondary   ClusterIP   10.128.213.243   <none>        1433/TCP   6m35s

NAME                                                        TYPE                     VERSION   AGE
appbinding.appcatalog.appscode.com/mssqlserver-ag-cluster   kubedb.com/mssqlserver   2022      5m31s

NAME                                            VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssqlserver-ag-cluster   2022-cu12   Ready    6m35s
```
Let’s check if the database is ready to use,

```bash
$ kubectl get mssqlserver -n demo mssqlserver-ag-cluster
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    7m11s
```
> We have successfully deployed MSSQL Server in GKE. Now we can exec into the container to use the database.


### Accessing Database Through CLI

To access your database through the CLI, you first need the credentials for the database. KubeDB will create several Kubernetes Secrets and Services for your MSSQL Server instance. To view them, use the following commands:

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mssqlserver-ag-cluster 
NAME                                   TYPE                       DATA   AGE
mssqlserver-ag-cluster-auth            kubernetes.io/basic-auth   2      8m31s
mssqlserver-ag-cluster-client-cert     kubernetes.io/tls          3      8m31s
mssqlserver-ag-cluster-dbm-login       kubernetes.io/basic-auth   1      8m31s
mssqlserver-ag-cluster-endpoint-cert   kubernetes.io/tls          3      8m31s
mssqlserver-ag-cluster-master-key      kubernetes.io/basic-auth   1      8m31s
mssqlserver-ag-cluster-server-cert     kubernetes.io/tls          3      8m31s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mssqlserver-ag-cluster 
NAME                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
mssqlserver-ag-cluster             ClusterIP   10.128.113.192   <none>        1433/TCP   8m37s
mssqlserver-ag-cluster-pods        ClusterIP   None             <none>        1433/TCP   8m37s
mssqlserver-ag-cluster-secondary   ClusterIP   10.128.213.243   <none>        1433/TCP   8m37s
```

From the above list, the `mssqlserver-ag-cluster-auth` Secret contains the admin-level credentials needed to connect to the database. Use the following commands to obtain the username and password:

```bash
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth  -o jsonpath='{.data.username}' | base64 -d
sa
$ kubectl get secret -n demo mssqlserver-ag-cluster-auth  -o jsonpath='{.data.password}' | base64 -d
dS57E93oLDi6wezv
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
mssql@mssqlserver-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "dS57E93oLDi6wezv"

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

# Verify that the database 'music' has been created and added to the `availability group cluster`. Then, we will insert data into this `music` database.
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
1> INSERT INTO Playlist(Artist, Song) VALUES ('Bobby Bare', 'Five Hundred Miles');
2> GO

(1 rows affected)
1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
Bobby Bare                                                                      Five Hundred Miles

(1 rows affected)
1> exit

...

# Confirm that the data inserted into the primary node has been replicated to the secondary nodes.
# Access the secondary node (Node 2) to verify that the data is present.
$ kubectl exec -it mssqlserver-ag-cluster-1 -n demo bash
Defaulted container "mssql" out of: mssql, mssql-coordinator, mssql-init (init)
mssql@mssqlserver-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "dS57E93oLDi6wezv"

1> USE music
2> GO
Changed database context to 'music'.

1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
Bobby Bare                                                                      Five Hundred Miles

(1 rows affected)
1> exit

...

# Access the secondary node (Node 3) to verify that the data is present.
$ kubectl exec -it mssqlserver-ag-cluster-2 -n demo bash
Defaulted container "mssql" out of: mssql, mssql-coordinator, mssql-init (init)
mssql@mssqlserver-ag-cluster-2:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "dS57E93oLDi6wezv"

1> USE music
2> GO
Changed database context to 'music'.

1> SELECT * FROM Playlist
2> GO

Artist                                                                          Song
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------
Bobby Bare                                                                      Five Hundred Miles

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
