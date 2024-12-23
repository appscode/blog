---
title: Deploy MySQL Remote Replica Across Cluster
date: "2023-10-10"
weight: 14
authors:
- Dipta Roy
tags:
- cloud-native
- database
- high-availability
- kubedb
- kubernetes
- mysql
- mysql-database
- mysql-remote-replica
- remote-replica
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases that KubeDB supports are MongoDB, Kafka, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, ProxySQL, Percona XtraDB, Memcached and PgBouncer. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/). In this tutorial we will show how to deploy MySQL Remote Replica across cluster. Remote Replica allows you to replicate data from an KubeDB managed MySQL server to a read-only MySQL server. The whole process uses MySQL asynchronous replication to keep up-to-date the replica with source server. It’s useful to use Remote Replica to scale of read-intensive workloads, can be a workaround for your BI and analytical workloads and can be geo-replicated.
We will cover the following steps:

1) Install KubeDB
2) Deploy MySQL with TLS/SSL
3) Insert Sample Data
4) Deploy MySQL in a Different Region
5) Validate Remote Replica

### Get Cluster ID

We need the cluster ID to get the KubeDB License. To get cluster ID, we can run the following command:

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
appscode/kubedb                   	v2023.10.9   	v2023.10.9 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.21.0      	v0.21.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2023.10.9   	v2023.10.9 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crds              	v2023.10.9   	v2023.10.9 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.12.0      	v0.12.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2023.10.9   	v2023.10.9 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics           	v2023.10.9   	v2023.10.9 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.10.9   	v2023.10.9 	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.23.0      	v0.23.1    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2023.10.9   	v2023.10.9 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provisioner       	v0.36.0      	v0.36.1    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.12.0      	v0.12.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2023.10.1   	0.4.5      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.12.0      	v0.12.0    	KubeDB Webhook Server by AppsCode 

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.10.9 \
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
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-57d9f76c78-vwpkw       1/1     Running   0          3m49s
kubedb      kubedb-kubedb-dashboard-7bfb4db4cf-rbprn        1/1     Running   0          3m49s
kubedb      kubedb-kubedb-ops-manager-7459d67fd4-gjsqc      1/1     Running   0          3m49s
kubedb      kubedb-kubedb-provisioner-5dddf7dcc5-pmc8k      1/1     Running   0          3m49s
kubedb      kubedb-kubedb-schema-manager-6b99f7846b-r84wr   1/1     Running   0          3m49s
kubedb      kubedb-kubedb-webhook-server-7f56c85dc5-rr6ps   1/1     Running   0          3m49s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-10-10T05:56:23Z
elasticsearchdashboards.dashboard.kubedb.com      2023-10-10T05:55:14Z
elasticsearches.kubedb.com                        2023-10-10T05:55:14Z
elasticsearchopsrequests.ops.kubedb.com           2023-10-10T05:55:55Z
elasticsearchversions.catalog.kubedb.com          2023-10-10T05:52:52Z
etcds.kubedb.com                                  2023-10-10T05:55:33Z
etcdversions.catalog.kubedb.com                   2023-10-10T05:52:52Z
kafkaopsrequests.ops.kubedb.com                   2023-10-10T05:56:28Z
kafkas.kubedb.com                                 2023-10-10T05:55:35Z
kafkaversions.catalog.kubedb.com                  2023-10-10T05:52:52Z
mariadbautoscalers.autoscaling.kubedb.com         2023-10-10T05:56:23Z
mariadbdatabases.schema.kubedb.com                2023-10-10T05:55:04Z
mariadbopsrequests.ops.kubedb.com                 2023-10-10T05:56:09Z
mariadbs.kubedb.com                               2023-10-10T05:55:04Z
mariadbversions.catalog.kubedb.com                2023-10-10T05:52:52Z
memcacheds.kubedb.com                             2023-10-10T05:55:33Z
memcachedversions.catalog.kubedb.com              2023-10-10T05:52:52Z
mongodbautoscalers.autoscaling.kubedb.com         2023-10-10T05:56:23Z
mongodbdatabases.schema.kubedb.com                2023-10-10T05:55:04Z
mongodbopsrequests.ops.kubedb.com                 2023-10-10T05:55:59Z
mongodbs.kubedb.com                               2023-10-10T05:55:04Z
mongodbversions.catalog.kubedb.com                2023-10-10T05:52:52Z
mysqlautoscalers.autoscaling.kubedb.com           2023-10-10T05:56:23Z
mysqldatabases.schema.kubedb.com                  2023-10-10T05:55:04Z
mysqlopsrequests.ops.kubedb.com                   2023-10-10T05:56:06Z
mysqls.kubedb.com                                 2023-10-10T05:55:04Z
mysqlversions.catalog.kubedb.com                  2023-10-10T05:52:52Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-10-10T05:56:23Z
perconaxtradbopsrequests.ops.kubedb.com           2023-10-10T05:56:21Z
perconaxtradbs.kubedb.com                         2023-10-10T05:55:34Z
perconaxtradbversions.catalog.kubedb.com          2023-10-10T05:52:52Z
pgbouncers.kubedb.com                             2023-10-10T05:55:34Z
pgbouncerversions.catalog.kubedb.com              2023-10-10T05:52:52Z
postgresautoscalers.autoscaling.kubedb.com        2023-10-10T05:56:23Z
postgresdatabases.schema.kubedb.com               2023-10-10T05:55:04Z
postgreses.kubedb.com                             2023-10-10T05:55:04Z
postgresopsrequests.ops.kubedb.com                2023-10-10T05:56:15Z
postgresversions.catalog.kubedb.com               2023-10-10T05:52:52Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-10-10T05:56:23Z
proxysqlopsrequests.ops.kubedb.com                2023-10-10T05:56:18Z
proxysqls.kubedb.com                              2023-10-10T05:55:34Z
proxysqlversions.catalog.kubedb.com               2023-10-10T05:52:52Z
publishers.postgres.kubedb.com                    2023-10-10T05:56:31Z
redisautoscalers.autoscaling.kubedb.com           2023-10-10T05:56:23Z
redises.kubedb.com                                2023-10-10T05:55:35Z
redisopsrequests.ops.kubedb.com                   2023-10-10T05:56:12Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-10-10T05:56:23Z
redissentinelopsrequests.ops.kubedb.com           2023-10-10T05:56:25Z
redissentinels.kubedb.com                         2023-10-10T05:55:35Z
redisversions.catalog.kubedb.com                  2023-10-10T05:52:52Z
subscribers.postgres.kubedb.com                   2023-10-10T05:56:34Z
```

## Deploy MySQL Server

We are going to Deploy MySQL server using KubeDB.
First, let's create a Namespace in which we will deploy the database.

```bash
$ kubectl create namespace demo
namespace/demo created
```

### Create Issuer

We will create a TLS secured instance since were planning to replicate across cluster. Lets start with creating a secret to access database and we will deploy a TLS secured instance. So, we will to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired `Issuer`,


```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mysql/O=kubedb"

$ kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/my-ca created
```
Now, we are going to create an `Issuer` using the `my-ca` secret that holds the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mysql-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```

Let’s create the `Issuer` we have shown above,

```bash
$ kubectl apply -f mysql-issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

### Create Auth Secret

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cm9vdA==
kind: Secret
metadata:
  name: mysql-singapore-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

Let’s create the Auth Secret we have shown above,

```bash
$ kubectl apply -f mysql-singapore-auth.yaml
secret/mysql-singapore-auth created
```

## Deploy MySQL with TLS/SSL configuration

Here is the yaml of the MySQL CRO we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-singapore
  namespace: demo
spec:
  authSecret:
    name: mysql-singapore-auth
  version: "8.0.31"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "linode-block-storage"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `mysql-singapore.yaml` 
Then create the above MySQL CRO

```bash
$ kubectl apply -f mysql-singapore.yaml
mysql.kubedb.com/mysql-singapore created
```
In this yaml,
* `spec.version` field specifies the version of MySQL. Here, we are using MySQL `8.0.31`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.topology` represents the clustering configuration for MySQL.
* `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `GroupReplication` to define the operator that we want to deploy MySQL Group Replication.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/).

Let’s check if the database is ready to use,

```bash
$ kubectl get mysql -n demo
NAME              VERSION   STATUS   AGE
mysql-singapore   8.0.31    Ready    22h
```
> We have successfully deployed MySQL. Now we can exec into the container to use the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. KubeDB will create Secret and Service for the database that we have deployed. Now, we are going to use `mysql-singapore-auth` to get the credentials.

```bash
$ kubectl get secrets -n demo mysql-singapore-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-singapore-auth -o jsonpath='{.data.\password}' | base64 -d
pass
```

#### Insert Sample Data

In this section, we are going to login into our MySQL database pod and insert some sample data.

```bash
# create a database on primary
$ kubectl exec -it -n demo mysql-singapore-0 -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo mysql-singapore-0 -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo mysql-singapore-0 -c mysql -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo mysql-singapore-0 -c mysql -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

```

> We've successfully inserted some sample data to our database. More information about Run & Manage MySQL on Kubernetes can be found in [MySQL Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)


## Expose MySQL to Outside

Here, we will expose our MySQL with ingress to outside,

```bash
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx  \
                                      --namespace demo --create-namespace \
                                      --set tcp.3306="demo/mysql-singapore:3306"
```

Let’s apply the ingress YAML thats refers to `mysql-singpore` service

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mysql-singapore
  namespace: demo  
spec:
  ingressClassName: nginx
  rules:
  - host: mysql-singapore.something.org
    http:
      paths:
      - backend:
          service:
            name: mysql-singapore
            port:
              number: 3306
        path: /
        pathType: Prefix
```

Save this yaml configuration into `mysql-singapore.yaml` and apply it,

```bash
$ kubectl apply -f mysql-singapore.yaml
ingress.networking.k8s.io/mysql-singapore created
```
Let's check the ingress,

```bash
$ kubectl get ingress -n demo
NAME              CLASS   HOSTS                           ADDRESS          PORTS   AGE
mysql-singapore   nginx   mysql-singapore.something.org   172.104.37.147   80      22h
```
Now will be able to communicate from another cluster to our source database.

### Prepare for Remote Replica

We wil use the [KubeDB Plugin](https://kubedb.com/docs/latest/setup/install/kubectl_plugin/) to generate YAML configuration for Remote Replica. It will create the [AppBinding](https://kubedb.com/docs/latest/guides/mysql/concepts/appbinding/) and and necessary secrets to connect with the source server.

```bash
$ kubectl dba remote-config mysql -n demo mysql-singapore -uremote -ppass -d 172.104.37.147 -y
home/test/yamls/mysql/mysql-singapore-remote-config.yaml
```
### Prepare for Remote Replica

We have prepared another cluster like above but now for London region for replicating across cluster.

#### Create sourceRef

We will apply the generated YAML config from kubeDB plugin to create the sourceRefs and secrets for it.

```bash
$ kubectl apply -f  home/test/yamls/mysql/mysql-singapore-remote-config.yaml

secret/mysql-singapore-remote-replica-auth created
secret/mysql-singapore-client-cert-remote created
appbinding.appcatalog.appscode.com/mysql-singapore created

$ kubectl get appbinding -n  demo
NAME              TYPE               VERSION   AGE
mysql-singapore   kubedb.com/mysql   8.0.31    4m17s
```

#### Create Remote Replica Auth

Here, we will need to use the same Auth secrets for Remote Replicas since operations like clone also replicated the auth-secrets from the source server.

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cm9vdA==
kind: Secret
metadata:
  name: mysql-london-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

Let’s save this yaml configuration into `mysql-london-auth.yaml` and apply it,

```bash
$ kubectl apply -f mysql-london-auth.yaml
secret/mysql-london-auth created
```

## Deploy MySQL in a Different Region

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-london
  namespace: demo
spec:
  authSecret:
    name: mysql-london-auth
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
    disableWriteCheck: true
  version: "8.0.31"
  replicas: 1
  topology:
    mode: RemoteReplica
    remoteReplica:
      sourceRef:
        name: mysql-singapore
        namespace: demo
  storageType: Durable
  storage:
    storageClassName: "linode-block-storage"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  terminationPolicy: WipeOut
```

In this yaml,
* `spec.version` field specifies the version of MySQL. Here, we are using MySQL `8.0.31`. You can list the KubeDB supported versions of MySQL by running `$ kubectl get mysqlversions` command.
* `spec.topology` represents the clustering configuration for MySQL.
* `spec.topology.mode` defining the server will be working a Remote Replica.
* `spec.topology.remoteReplica.sourceref` referring to source to read, the MySQL instance we previously created.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/mysql/concepts/database/).

Let's save this yaml configuration into `mysql-london.yaml` 
Then apply the above MySQL CRO

```bash
$ kubectl apply -f mysql-london.yaml
mysql.kubedb.com/mysql-london created
```
Now, KubeDB will provision a Remote Replica from the source MySQL instance. KubeDB operator sets the `status.phase` to Ready once the database is successfully created. Run the following command to see the modified MySQL object:

```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-london   8.0.31    Ready    7m17s
```

## Validate Remote Replica

Since both source and replica database are in the `ready` state, now we can validate Remote Replica if that is working properly by checking the replication status,

```bash
$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass' --host=mysql-london-0.mysql-london-pods.demo -e "show slave status\G" 
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: 172.104.37.147
                  Master_User: remote
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000001
          Read_Master_Log_Pos: 4698131
               Relay_Log_File: mysql-london-0-relay-bin.000007
                Relay_Log_Pos: 1415154
        Relay_Master_Log_File: binlog.000001
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
            ....   
```
### Read Data

Previously, we have inserted some sample data into the primary pod. Now, we'll read from secondary pods to determine whether the data has been successfully copied to the secondary pods or not.

```bash

$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass'  --host=mysql-london-0.mysql-london-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

> So, we've successfully accessed the sample data from different region via Remote Replica.

### Failover Remote Replica
In case you need to rsync with the primary cluster or any other secondary if available with mysql `clone plugin`

```bash
$ kubectl exec -it -ndemo  mysql-london-0 -- bash
    $ mysql -uroot $MYSQL_ROOT_PASSWORD
      
      mysql> SET GLOBAL clone_valid_donor_list='172.104.37.147:3306'
      mysql> CLONE INSTANCE FROM 'root'@'172.104.37.147':3306 IDENTIFIED BY 'pass';
```

If you want to learn more about Production-Grade MySQL you can have a look into that playlist below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/videoseries?si=dmtaSpYlKplDUA_G&amp;list=PLoiT1Gv2KR1gNPaHZtfdBZb6G4wLx6Iks" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [MySQL on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-mysql-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
