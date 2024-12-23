---
title: Update Version of Kafka in Amazon Elastic Kubernetes Service (Amazon EKS)
date: "2024-03-29"
weight: 14
authors:
- Dipta Roy
tags:
- amazon-eks
- apache-kafka
- aws
- cloud-native
- eks
- kafka
- kubedb
- kubernetes
- streaming-platform
---

## Overview

KubeDB is the Kubernetes Native Database Management Solution which simplifies and automates routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup, Recovery, Failure detection, and Repair for various popular databases on private and public clouds. The databases supported by KubeDB include MongoDB, Elasticsearch, MySQL, MariaDB, Redis, PostgreSQL, FerretDB, Percona XtraDB, and Memcached. Additionally, KubeDB also supports ProxySQL, PgBouncer and the streaming platform Kafka. You can find the guides to all the supported databases in [KubeDB](https://kubedb.com/).
In this tutorial we will update version of Kafka in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Deploy Kafka Cluster
3) Publish & Consume Messages with Kafka
4) Update Kafka Version

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
appscode/kubedb                   	v2024.3.16   	v2024.3.16 	KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler        	v0.29.0      	v0.29.0    	KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog           	v2024.3.16   	v2024.3.16 	KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-community         	v0.24.2      	v0.24.2    	KubeDB Community by AppsCode - Community featur...
appscode/kubedb-crd-manager       	v0.0.8       	v0.0.8     	KubeDB CRD Manager by AppsCode                    
appscode/kubedb-crds              	v2024.3.16   	v2024.3.16 	KubeDB Custom Resource Definitions                
appscode/kubedb-dashboard         	v0.20.0      	v0.20.0    	KubeDB Dashboard by AppsCode                      
appscode/kubedb-enterprise        	v0.11.2      	v0.11.2    	KubeDB Enterprise by AppsCode - Enterprise feat...
appscode/kubedb-grafana-dashboards	v2024.3.16   	v2024.3.16 	A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-kubestash-catalog 	v2024.3.16   	v2024.3.16 	KubeStash Catalog by AppsCode - Catalog of Kube...
appscode/kubedb-metrics           	v2024.3.16   	v2024.3.16 	KubeDB State Metrics                              
appscode/kubedb-one               	v2023.12.28  	v2023.12.28	KubeDB and Stash by AppsCode - Production ready...
appscode/kubedb-ops-manager       	v0.31.0      	v0.31.0    	KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter         	v2024.3.16   	v2024.3.16 	KubeDB Opscenter by AppsCode                      
appscode/kubedb-provider-aws      	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB AWS Provider for Crossp...
appscode/kubedb-provider-azure    	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB Azure Provider for Cros...
appscode/kubedb-provider-gcp      	v2024.3.16   	v0.6.0     	A Helm chart for KubeDB GCP Provider for Crossp...
appscode/kubedb-provisioner       	v0.44.0      	v0.44.0    	KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager    	v0.20.0      	v0.20.0    	KubeDB Schema Manager by AppsCode                 
appscode/kubedb-ui                	v2024.2.13   	0.6.4      	A Helm chart for Kubernetes                       
appscode/kubedb-ui-server         	v2021.12.21  	v2021.12.21	A Helm chart for kubedb-ui-server by AppsCode     
appscode/kubedb-webhook-server    	v0.20.0      	v0.20.0    	KubeDB Webhook Server by AppsCode  


$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2024.3.16 \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

Let's verify the installation:

```bash
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-65bb97ff8f-h2ct5       1/1     Running   0          4m1s
kubedb      kubedb-kubedb-ops-manager-57cd64669f-s4hqw      1/1     Running   0          4m1s
kubedb      kubedb-kubedb-provisioner-66f54589cc-q7phb      1/1     Running   0          4m1s
kubedb      kubedb-kubedb-webhook-server-5596cd9b99-mhklw   1/1     Running   0          4m1s
kubedb      kubedb-petset-operator-5d94b4ddb8-pgnxd         1/1     Running   0          4m1s
kubedb      kubedb-petset-webhook-server-75c7fd9c54-gdv4b   2/2     Running   0          4m1s
kubedb      kubedb-sidekick-5dc87959b7-8bnlv                1/1     Running   0          4m1s
```

We can list the CRD Groups that have been registered by the operator by running the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
NAME                                               CREATED AT
connectclusters.kafka.kubedb.com                   2024-03-29T05:13:05Z
connectors.kafka.kubedb.com                        2024-03-29T05:13:05Z
druidversions.catalog.kubedb.com                   2024-03-29T05:12:30Z
elasticsearchautoscalers.autoscaling.kubedb.com    2024-03-29T05:13:02Z
elasticsearchdashboards.elasticsearch.kubedb.com   2024-03-29T05:13:02Z
elasticsearches.kubedb.com                         2024-03-29T05:13:02Z
elasticsearchopsrequests.ops.kubedb.com            2024-03-29T05:13:02Z
elasticsearchversions.catalog.kubedb.com           2024-03-29T05:12:30Z
etcdversions.catalog.kubedb.com                    2024-03-29T05:12:30Z
ferretdbversions.catalog.kubedb.com                2024-03-29T05:12:30Z
kafkaautoscalers.autoscaling.kubedb.com            2024-03-29T05:13:05Z
kafkaconnectorversions.catalog.kubedb.com          2024-03-29T05:12:30Z
kafkaopsrequests.ops.kubedb.com                    2024-03-29T05:13:05Z
kafkas.kubedb.com                                  2024-03-29T05:13:05Z
kafkaversions.catalog.kubedb.com                   2024-03-29T05:12:31Z
mariadbarchivers.archiver.kubedb.com               2024-03-29T05:13:09Z
mariadbautoscalers.autoscaling.kubedb.com          2024-03-29T05:13:08Z
mariadbdatabases.schema.kubedb.com                 2024-03-29T05:13:08Z
mariadbopsrequests.ops.kubedb.com                  2024-03-29T05:13:08Z
mariadbs.kubedb.com                                2024-03-29T05:13:08Z
mariadbversions.catalog.kubedb.com                 2024-03-29T05:12:31Z
memcachedversions.catalog.kubedb.com               2024-03-29T05:12:31Z
mongodbarchivers.archiver.kubedb.com               2024-03-29T05:13:12Z
mongodbautoscalers.autoscaling.kubedb.com          2024-03-29T05:13:12Z
mongodbdatabases.schema.kubedb.com                 2024-03-29T05:13:12Z
mongodbopsrequests.ops.kubedb.com                  2024-03-29T05:13:12Z
mongodbs.kubedb.com                                2024-03-29T05:13:12Z
mongodbversions.catalog.kubedb.com                 2024-03-29T05:12:31Z
mysqlarchivers.archiver.kubedb.com                 2024-03-29T05:13:16Z
mysqlautoscalers.autoscaling.kubedb.com            2024-03-29T05:13:16Z
mysqldatabases.schema.kubedb.com                   2024-03-29T05:13:16Z
mysqlopsrequests.ops.kubedb.com                    2024-03-29T05:13:16Z
mysqls.kubedb.com                                  2024-03-29T05:13:16Z
mysqlversions.catalog.kubedb.com                   2024-03-29T05:12:31Z
perconaxtradbversions.catalog.kubedb.com           2024-03-29T05:12:31Z
pgbouncerversions.catalog.kubedb.com               2024-03-29T05:12:31Z
pgpoolversions.catalog.kubedb.com                  2024-03-29T05:12:31Z
postgresarchivers.archiver.kubedb.com              2024-03-29T05:13:19Z
postgresautoscalers.autoscaling.kubedb.com         2024-03-29T05:13:19Z
postgresdatabases.schema.kubedb.com                2024-03-29T05:13:19Z
postgreses.kubedb.com                              2024-03-29T05:13:19Z
postgresopsrequests.ops.kubedb.com                 2024-03-29T05:13:19Z
postgresversions.catalog.kubedb.com                2024-03-29T05:12:31Z
proxysqlversions.catalog.kubedb.com                2024-03-29T05:12:31Z
publishers.postgres.kubedb.com                     2024-03-29T05:13:19Z
rabbitmqversions.catalog.kubedb.com                2024-03-29T05:12:31Z
redisautoscalers.autoscaling.kubedb.com            2024-03-29T05:13:23Z
redises.kubedb.com                                 2024-03-29T05:13:23Z
redisopsrequests.ops.kubedb.com                    2024-03-29T05:13:23Z
redissentinelautoscalers.autoscaling.kubedb.com    2024-03-29T05:13:23Z
redissentinelopsrequests.ops.kubedb.com            2024-03-29T05:13:23Z
redissentinels.kubedb.com                          2024-03-29T05:13:23Z
redisversions.catalog.kubedb.com                   2024-03-29T05:12:31Z
singlestoreversions.catalog.kubedb.com             2024-03-29T05:12:31Z
solrversions.catalog.kubedb.com                    2024-03-29T05:12:31Z
subscribers.postgres.kubedb.com                    2024-03-29T05:13:19Z
zookeeperversions.catalog.kubedb.com               2024-03-29T05:12:31Z
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
  version: 3.3.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: gp2
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
* `spec.version` field specifies the version of Kafka. Here, we are using Kafka `3.3.2`. You can list the KubeDB supported versions of Kafka by running `$ kubectl get kafkaversions` command.
* `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. 
* `spec.terminationPolicy` field is *Wipeout* means it will be deleted without restrictions. It can also be “Halt”, “Delete” and “DoNotTerminate”. Learn More about these checkout [Termination Policy](https://kubedb.com/docs/latest/guides/kafka/concepts/kafka/#specterminationpolicy) .

Once these are handled correctly and the Kafka object is deployed, you will see that the following objects are created:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=kafka-cluster'
NAME                  READY   STATUS    RESTARTS   AGE
pod/kafka-cluster-0   1/1     Running   0          2m32s
pod/kafka-cluster-1   1/1     Running   0          87s
pod/kafka-cluster-2   1/1     Running   0          81s

NAME                         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-cluster-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   2m33s

NAME                             READY   AGE
statefulset.apps/kafka-cluster   3/3     2m32s

NAME                                               TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-cluster   kubedb.com/kafka   3.3.2     2m32s

NAME                              TYPE                       DATA   AGE
secret/kafka-cluster-admin-cred   kubernetes.io/basic-auth   2      2m33s
secret/kafka-cluster-config       Opaque                     2      2m33s
```
Let’s check if the `kafka-cluster` is ready to use,

```bash
$ kubectl get kafka -n demo kafka-cluster
NAME            TYPE                  VERSION   STATUS   AGE
kafka-cluster   kubedb.com/v1alpha2   3.3.2     Ready    2m50s
```
> We have successfully deployed Kafka cluster in Amazon EKS.

## Publish & Consume Messages with Kafka

### Accessing Kafka Through CLI

In this section, we will now exec into one of the kafka brokers in interactive mode and then describe the broker metadata for the quorum.

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash
kafka@kafka-cluster-0:~$ kafka-metadata-quorum.sh --command-config $HOME/config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11ee-a586-62737322a62w
LeaderId:               1
LeaderEpoch:            23
HighWatermark:          331
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   350
CurrentVoters:          [0,1,2]
CurrentObservers:       []
```

We can see the important metadata information like clusterID, current leader ID, node IDs which are participating in leader election voting and IDs of those brokers who are observers. It is important to mention that each broker is assigned a numeric ID which is called its broker ID. The ID is assigned sequentially with respect to the host pod name.

### Create a Topic

Let’s create a topic named `music` with 3 partitions and a replication factor of 3. Describe the topic once it’s created. You will see the leader ID for each partition and their replica IDs along with in-sync-replicas(ISR).

```bash
$ kubectl exec -it -n demo  kafka-cluster-0 -- bash

kafka@kafka-cluster-0:~$ kafka-topics.sh --create  --bootstrap-server localhost:9092 --command-config $HOME/config/clientauth.properties --topic music --partitions 3 --replication-factor 3

Created topic music.

kafka@kafka-cluster-0:~$ kafka-topics.sh --describe --topic music --bootstrap-server localhost:9092 --command-config $HOME/config/clientauth.properties

Topic: music	TopicId: 2iWEOLHyQDc2iIdRBJez2f	PartitionCount: 3	ReplicationFactor: 3	Configs: segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: music	Partition: 0	Leader: 1	Replicas: 1,2,0	Isr: 1,2,0
	Topic: music	Partition: 1	Leader: 2	Replicas: 2,0,1	Isr: 2,0,1
	Topic: music	Partition: 2	Leader: 0	Replicas: 0,1,2	Isr: 0,1,2

```
Now, we are going to start a producer and a consumer for topic `music`. Let’s use this current terminal for producing messages and open a new terminal for consuming messages. From the topic description we can see that the leader partition for partition 0 is 1 (the broker that we are on). If we produce messages to `kafka-cluster-1` broker(brokerID=1) it will store those messages in partition 0 and  `--request-required-acks all` ensures that the message is durably stored on all replicas before the producer considers the message sent. Let’s produce messages in the producer terminal and consume them from the consumer terminal.

```bash
$ kubectl exec -it -n demo  kafka-cluster-1 -- bash
kafka@kafka-cluster-1:~$ kafka-console-producer.sh  --topic music --request-required-acks all --bootstrap-server localhost:9092 --producer.config $HOME/config/clientauth.properties

>The Nights
>Annie's Song
>Five Hundred Miles
```

```bash
$ kubectl exec -it -n demo  kafka-cluster-1 -- bash
kafka@kafka-cluster-1:~$ kafka-console-consumer.sh --topic music --from-beginning --bootstrap-server localhost:9092 --consumer.config $HOME/config/clientauth.properties

The Nights
Annie's Song
Five Hundred Miles
```
> Here we can see messages are coming to the consumer as you continue sending messages via producer. So, we have created a Kafka topic and used Kafka console producer and consumer for publishing and consuming messages successfully. More information about Deploy & Manage Kafka on Kubernetes can be found in [Kafka Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-kafka-on-kubernetes/)


## Update Kafka Version

In this section, we will update our Kafka version from `3.3.2` to the latest version `3.6.1`. Let’s check the current version,

```bash
$ kubectl get kafka -n demo kafka-cluster -o=jsonpath='{.spec.version}{"\n"}'
3.3.2
```

### Create KafkaOpsRequest

In order to update the version of Kafka, we have to create a KafkaOpsRequest CR with your desired version that is supported by KubeDB. Below is the YAML of the KafkaOpsRequest CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: update-version 
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: kafka-cluster
  updateVersion:
    targetVersion: 3.6.1
```

Let’s save this YAML configuration into `update-version.yaml` and apply it,

```bash
$ kubectl apply -f update-version.yaml
kafkaopsrequest.ops.kubedb.com/update-version created
```

In this yaml,
* `spec.databaseRef.name` specifies that we are performing operation on Kafka-cluster.
* `spec.type` specifies that we are going to perform UpdateVersion on our database.
* `spec.updateVersion.targetVersion` specifies the expected version of Kafka to `3.6.1`.

### Verify the Updated Kafka Version

KubeDB operator will update the image of Kafka and related `StatefulSets` and `Pods`. Let’s wait for KafkaOpsRequest to be Successful. Run the following command to check KafkaOpsRequest CR,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME             TYPE            STATUS       AGE
update-version   UpdateVersion   Successful   3m50s
```

We can see from the above output that the KafkaOpsRequest has succeeded. Now, we are going to verify whether the Kafka and the related `StatefulSets` their `Pods` have the new version image. Let’s verify it by following command,

```bash
$ kubectl get kafka -n demo kafka-cluster -o=jsonpath='{.spec.version}{"\n"}'
3.6.1
```

> You can see from above, Kafka has been updated with the new version`3.6.1`. So, the update process is successfully completed.

If you want to learn about Kafka Ops Requests - Day 2 Lifecycle Management Using KubeDB you can have a look into that video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/EYGZaKfbqVE?si=UxxA2uXY6X000Vdj" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://youtube.com/@appscode) channel.

Learn more about [Kafka on Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-kafka-on-kubernetes/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
