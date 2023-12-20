---
title: Deploy Kafka Cluster in Google Kubernetes Engine (GKE)
date: "2023-12-11"
weight: 14
authors:
- Dipta Roy
tags:
- apache-kafka
- cloud-native
- gcp
- gcs
- gke
- kafka
- kubedb
- kubernetes
- streaming-platform
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer, and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will deploy Kafka Cluster in Google Kubernetes Engine (GKE). We will cover the following steps:

1) Install KubeDB
2) Deploy Kafka Cluster
3) Publish & Consume Messages with Kafka

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
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

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

# Install KubeDB operator chart
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2023.12.11 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-86b776dc7-cndqr        1/1     Running   0          5m7s
kubedb      kubedb-kubedb-dashboard-6c7fb4d7c5-6l4s9        1/1     Running   0          5m7s
kubedb      kubedb-kubedb-ops-manager-c4ffbc66f-m5g6k       1/1     Running   0          5m7s
kubedb      kubedb-kubedb-provisioner-7bfb4bdb7b-mmts2      1/1     Running   0          5m7s
kubedb      kubedb-kubedb-webhook-server-7c78b9f747-pqf65   1/1     Running   0          5m7s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                              CREATED AT
elasticsearchautoscalers.autoscaling.kubedb.com   2023-12-11T09:22:54Z
elasticsearchdashboards.dashboard.kubedb.com      2023-12-11T09:22:54Z
elasticsearches.kubedb.com                        2023-12-11T09:22:54Z
elasticsearchopsrequests.ops.kubedb.com           2023-12-11T09:23:02Z
elasticsearchversions.catalog.kubedb.com          2023-12-11T09:20:43Z
etcds.kubedb.com                                  2023-12-11T09:23:01Z
etcdversions.catalog.kubedb.com                   2023-12-11T09:20:44Z
kafkaopsrequests.ops.kubedb.com                   2023-12-11T09:23:54Z
kafkas.kubedb.com                                 2023-12-11T09:23:11Z
kafkaversions.catalog.kubedb.com                  2023-12-11T09:20:44Z
mariadbautoscalers.autoscaling.kubedb.com         2023-12-11T09:22:54Z
mariadbopsrequests.ops.kubedb.com                 2023-12-11T09:23:34Z
mariadbs.kubedb.com                               2023-12-11T09:23:02Z
mariadbversions.catalog.kubedb.com                2023-12-11T09:20:44Z
memcacheds.kubedb.com                             2023-12-11T09:23:02Z
memcachedversions.catalog.kubedb.com              2023-12-11T09:20:45Z
mongodbarchivers.archiver.kubedb.com              2023-12-11T09:23:12Z
mongodbautoscalers.autoscaling.kubedb.com         2023-12-11T09:22:55Z
mongodbopsrequests.ops.kubedb.com                 2023-12-11T09:23:06Z
mongodbs.kubedb.com                               2023-12-11T09:23:04Z
mongodbversions.catalog.kubedb.com                2023-12-11T09:20:45Z
mysqlautoscalers.autoscaling.kubedb.com           2023-12-11T09:22:55Z
mysqlopsrequests.ops.kubedb.com                   2023-12-11T09:23:30Z
mysqls.kubedb.com                                 2023-12-11T09:23:05Z
mysqlversions.catalog.kubedb.com                  2023-12-11T09:20:45Z
perconaxtradbautoscalers.autoscaling.kubedb.com   2023-12-11T09:22:55Z
perconaxtradbopsrequests.ops.kubedb.com           2023-12-11T09:23:47Z
perconaxtradbs.kubedb.com                         2023-12-11T09:23:06Z
perconaxtradbversions.catalog.kubedb.com          2023-12-11T09:20:46Z
pgbouncers.kubedb.com                             2023-12-11T09:23:07Z
pgbouncerversions.catalog.kubedb.com              2023-12-11T09:20:46Z
postgresarchivers.archiver.kubedb.com             2023-12-11T09:23:15Z
postgresautoscalers.autoscaling.kubedb.com        2023-12-11T09:22:55Z
postgreses.kubedb.com                             2023-12-11T09:23:07Z
postgresopsrequests.ops.kubedb.com                2023-12-11T09:23:41Z
postgresversions.catalog.kubedb.com               2023-12-11T09:20:46Z
proxysqlautoscalers.autoscaling.kubedb.com        2023-12-11T09:22:56Z
proxysqlopsrequests.ops.kubedb.com                2023-12-11T09:23:44Z
proxysqls.kubedb.com                              2023-12-11T09:23:08Z
proxysqlversions.catalog.kubedb.com               2023-12-11T09:20:47Z
publishers.postgres.kubedb.com                    2023-12-11T09:23:57Z
redisautoscalers.autoscaling.kubedb.com           2023-12-11T09:22:56Z
redises.kubedb.com                                2023-12-11T09:23:09Z
redisopsrequests.ops.kubedb.com                   2023-12-11T09:23:37Z
redissentinelautoscalers.autoscaling.kubedb.com   2023-12-11T09:22:56Z
redissentinelopsrequests.ops.kubedb.com           2023-12-11T09:23:51Z
redissentinels.kubedb.com                         2023-12-11T09:23:10Z
redisversions.catalog.kubedb.com                  2023-12-11T09:20:47Z
subscribers.postgres.kubedb.com                   2023-12-11T09:24:01Z
```

## Deploy Kafka Cluster

We are going to Deploy Kafka Cluster by using KubeDB.
First, let's create a Namespace in which we will deploy Kafka.

```bash
$ kubectl create namespace demo
namespace/demo created
```

Here is the yaml of the Kafka CR we are going to use:

```yaml                                                                      
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-cluster
  namespace: demo
spec:
  replicas: 3
  version: 3.6.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `kafka-cluster.yaml` 
Then create the above Kafka CR

```bash
$ kubectl apply -f kafka-cluster.yaml
kafka.kubedb.com/kafka-cluster created
```
In this yaml,
* `spec.version` field specifies the version of Kafka. Here, we are using Kafka `3.6.0`. You can list the KubeDB supported versions of Kafka by running `$ kubectl get kafkaversions` command.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means it will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/kafka/concepts/kafka/#specterminationpolicy) .

Once these are handled correctly and the Kafka object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=kafka-cluster'
NAME                  READY   STATUS    RESTARTS   AGE
pod/kafka-cluster-0   1/1     Running   0          5m22s
pod/kafka-cluster-1   1/1     Running   0          4m50s
pod/kafka-cluster-2   1/1     Running   0          4m11s

NAME                         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-cluster-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   5m24s

NAME                             READY   AGE
statefulset.apps/kafka-cluster   3/3     5m25s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-cluster   kubedb.com/kafka   3.6.0     5m25s

NAME                              TYPE                       DATA   AGE
secret/kafka-cluster-admin-cred   kubernetes.io/basic-auth   2      5m31s
secret/kafka-cluster-config       Opaque                     2      5m31s
```
Let’s check if the `kafka-cluster` is ready to use,

```bash
$ kubectl get kafka -n demo kafka-cluster
NAME            TYPE                  VERSION   STATUS   AGE
kafka-cluster   kubedb.com/v1alpha2   3.6.0     Ready    3m48s
```
> We have successfully deployed Kafka cluster in GKE.

## Publish & Consume Messages with Kafka

### Accessing Kafka Through CLI

In this section, we will now exec into one of the kafka brokers in interactive mode and then describe the broker metadata for the quorum.

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash
kafka@kafka-cluster-0:~$ kafka-metadata-quorum.sh --command-config $HOME/config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11ee-a481-4ab0c9c3652w
LeaderId:               0
LeaderEpoch:            7
HighWatermark:          1436
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   0
CurrentVoters:          [0,1,2]
CurrentObservers:       []
```

We can see the important metadata information like clusterID, current leader ID, broker IDs which are participating in leader election voting and IDs of those brokers who are observers. It is important to mention that each broker is assigned a numeric ID which is called its broker ID. The ID is assigned sequentially with respect to the host pod name.

### Create a Topic

Let’s create a topic named `music` with 3 partitions and a replication factor of 3. Describe the topic once it’s created. You will see the leader ID for each partition and their replica IDs along with in-sync-replicas(ISR).

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash

kafka@kafka-cluster-0:~$ kafka-topics.sh --create  --bootstrap-server localhost:9092 --command-config $HOME/config/clientauth.properties --topic music --partitions 3 --replication-factor 3

Created topic music.

kafka@kafka-cluster-0:~$ kafka-topics.sh --describe --topic music --bootstrap-server localhost:9092 --command-config $HOME/config/clientauth.properties

Topic: music	TopicId: 4QTWweCFSxaQmd1L-HERKA	PartitionCount: 3	ReplicationFactor: 3	Configs: segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: music	Partition: 0	Leader: 2	Replicas: 2,0,1	Isr: 2,0,1
	Topic: music	Partition: 1	Leader: 0	Replicas: 0,1,2	Isr: 0,1,2
	Topic: music	Partition: 2	Leader: 1	Replicas: 1,2,0	Isr: 1,2,0

```
Now, we are going to start a producer and a consumer for topic `music`. Let’s use this current terminal for producing messages and open a new terminal for consuming messages. From the topic description we can see that the leader partition for partition 1 is 0 (the broker that we are on). If we produce messages to `kafka-cluster-0` broker(brokerID=0) it will store those messages in partition 1. Let’s produce messages in the producer terminal and consume them from the consumer terminal.

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash
kafka@kafka-cluster-0:~$ kafka-console-producer.sh  --topic music --request-required-acks all --bootstrap-server localhost:9092 --producer.config $HOME/config/clientauth.properties

>Five Hundred Miles
>Annie's Song
>The Nights
```

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash
kafka@kafka-cluster-0:~$ kafka-console-consumer.sh --topic music --from-beginning --bootstrap-server localhost:9092 --consumer.config $HOME/config/clientauth.properties

Five Hundred Miles
Annie's Song
The Nights
```
> Here we can see messages are coming to the consumer as you continue sending messages via producer. So, we have created a Kafka topic and used Kafka console producer and consumer for publishing and consuming messages successfully. More information about Run & Manage Kafka on Kubernetes can be found in [Kafka Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-kafka-on-kubernetes/)

If you want to learn more about Kafka Ops Requests - Day 2 Lifecycle Management Using KubeDB you can have a look into that video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/EYGZaKfbqVE?si=UxxA2uXY6X000Vdj" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn more about [Kafka in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-kafka-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
