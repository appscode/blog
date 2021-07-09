---
title: Enable G1 Garbage Collector for Elasticsearch to Avoid Circuit Breaker Exceptions
date: 2021-07-10
weight: 26
authors:
  - Md Kamol Hasan
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - garbage-collector
  - jvm-options
  - G1GC
  - CMS
  - kubedb
---


By default, the Opendistro of Elasticsearch cluster starts Concurrent Mark Sweep (`CMS`) garbage collector. In this blog post, we will see how the default `jvm.options` may lead to the circuit breaker exceptions and how can we avoid it by enabling the Garbage-First(`G1`) garbage collector.

## Elasticsearch Garbage Collectors

Elasticsearch mainly uses two different garbage collectors of Java: Concurrent Mark Sweep (CMS) and Garbage-First(G1).

[jvm.options](https://github.com/elastic/elasticsearch/blob/v7.13.3/distribution/src/config/jvm.options):

```options
## GC configuration

8-13:-XX:+UseConcMarkSweepGC
8-13:-XX:CMSInitiatingOccupancyFraction=75
8-13:-XX:+UseCMSInitiatingOccupancyOnly

## G1GC Configuration
# to use G1GC, uncomment the next two lines and update the version on the
# following three lines to your version of the JDK

# 8-13:-XX:-UseConcMarkSweepGC
# 8-13:-XX:-UseCMSInitiatingOccupancyOnly
14-:-XX:+UseG1GC
```

The CMS uses multiple garbage collector treads for garbage collection. It is designed for applications that prefer shorter garbage collection pauses. Overheads occur when the collector needs to promote young objects to the old generations, but didn't have enough time to clear the space.

```log
[WARN ][o.e.m.j.JvmGcMonitorService] [elasticsearch-client-1] [gc][400] overhead, spent [856ms] collecting in the last [1.5s] 
```

```log
[INFO ][o.e.i.b.HierarchyCircuitBreakerService] [elasticsearch-ingest-0] GC did not bring memory usage down, before [494656216], after [498360048], allocations [37], duration [13]

```

If more than 98% of the total time is spent in garbage collection and less than 2% of the heap is recovered, an `OutOfMemoryError` will be thrown. This feature is designed to prevent applications from running for an extended period while making little or no progress because the heap is too small. If necessary, this feature can be disabled by adding the option `-XX:-UseGCOverheadLimit` to the command line.

```log
[ERROR][o.e.b.ElasticsearchUncaughtExceptionHandler] [elasticsearch-ingest-0] fatal error in thread [Thread-7], exiting
java.lang.OutOfMemoryError: Java heap space
fatal error in thread [Thread-7], exiting
java.lang.OutOfMemoryError: Java heap space
```

```log
[ERROR][o.e.b.ElasticsearchUncaughtExceptionHandler] [elasticsearch-ingest-1] fatal error in thread [Thread-7], exiting
java.lang.OutOfMemoryError: Java heap space
fatal error in thread [Thread-7], exiting
java.lang.OutOfMemoryError: Java heap space

```

## Environment

Kubernetes cluster version:

```bash
$ kubectl version --short
Client Version: v1.21.1
Server Version: v1.21.1
```

Helm charts:

```bash
$ helm list --all --all-namespaces
NAME                 	NAMESPACE              	REVISION	UPDATED                                	STATUS  	CHART                            APP VERSION
kube-prometheus-stack	monitoring             	1       	2021-07-06 18:54:20.421959653 +0600 +06	deployed	kube-prometheus-stack-16.12.1    0.48.1     
kubedb              	kube-system            	1       	2021-07-09 11:22:17.527521498 +0600 +06	deployed	kubedb-v2021.06.23               v2021.06.23                      
rancher              	cattle-system          	1       	2021-07-07 17:35:21.293544984 +0600 +06	deployed	rancher-2.5.8                    v2.5.8     
```

- Install KubeDB operator from [here](https://kubedb.com/docs/v2021.06.23/setup/).
- Install Rancher from [here](https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/).
- Install kube-prometheus-stack from [here](https://artifacthub.io/packages/helm/prometheus-worawutchan/kube-prometheus-stack).

Let's create a new namespace for the demonstrations:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Elasticsearch with Default JVM Configurations

Since we have the KubeDB operator installed, let's deploy the Elasticsearch cluster with 3 master nodes, 2 data nodes , and 2 ingest nodes. We will use the Elasticsearch images provided by the Opendistro of Elasticsearch.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: elasticsearch
  namespace: demo
spec:
  enableSSL: false 
  version: opendistro-1.12.0
  storageType: Durable
  terminationPolicy: WipeOut
  # enable prometheus monitoring
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: kube-prometheus-stack
        interval: 10s
  topology:
    master:
      suffix: master
      replicas: 3
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      suffix: data
      replicas: 2 
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
    ingest:
      suffix: ingest
      replicas: 2
      storage:
        storageClassName: "linode-block-storage"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Deploy Elasticsearch:

```bash
$ kubectl apply -f elasticsearch.yaml
elasticsearch.kubedb.com/elasticsearch created
```

Wait for the Elasticsearch to become ready:

```bash
$ kubectl get elasticsearch -n demo -w
NAME            VERSION             STATUS         AGE
elasticsearch   opendistro-1.12.0   Provisioning   52s
elasticsearch   opendistro-1.12.0   Provisioning   84s
elasticsearch   opendistro-1.12.0   Provisioning   110s
elasticsearch   opendistro-1.12.0   Ready          2m21s
```

```bash
$ kubectl get pods -n demo
NAME                     READY   STATUS    RESTARTS   AGE
elasticsearch-data-0     1/1     Running   0          4m
elasticsearch-data-1     1/1     Running   0          3m
elasticsearch-ingest-0   2/2     Running   0          4m
elasticsearch-ingest-1   2/2     Running   0          3m
elasticsearch-master-0   1/1     Running   0          4m
elasticsearch-master-1   1/1     Running   0          3m
elasticsearch-master-2   1/1     Running   0          2m
```

Now, we will generate load to our Elasticsearch cluster. We will use rancher UI to write cluster logs to our Elasticsearch cluster.

```bash
# port-forward rancher UI:
$ kubectl port-forward -n cattle-system svc/rancher 8443:443
Forwarding from 127.0.0.1:8443 -> 444
```

Visit: https://localhost:8433

```
```

## References

- [JVM Garbage Collectors](https://www.baeldung.com/jvm-garbage-collectors)
- [OutOfMemoryError - GC overhead limit exceeded](https://www.petefreitag.com/item/746.cfm)
- [Garbage Collectors – Serial vs. Parallel vs. CMS vs. G1 (and what’s new in Java 8)](https://www.overops.com/blog/garbage-collectors-serial-vs-parallel-vs-cms-vs-the-g1-and-whats-new-in-java-8/)
- [Grafana Dashboard](https://github.com/prometheus-community/elasticsearch_exporter/blob/master/examples/grafana/dashboard.json)