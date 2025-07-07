---
title: Expose Redis Clusters outside Kubernetes using Horizon DNS
date: "2025-07-01"
weight: 25
authors:
- Hiranmoy Das Chowdhury
tags:
- announce
- cluster
- database
- horizondns
- kubedb
- kubernetes
- redis
---

> New to KubeDB? Please start [here](https://kubedb.com/).

# Expose Redis Clusters outside Kubernetes using Horizon DNS

Redis Announce is a feature in Redis that enables external connections to Redis cluster deployed within Kubernetes. It allows applications or clients outside the Kubernetes cluster to connect to different shards of redis cluster by mapping internal Kubernetes DNS names to externally accessible hostnames or IP addresses. This is useful for scenarios where external access is needed, such as hybrid deployments or connecting from outside the cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](https://kubedb.com/docs/v2025.6.30/setup/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/master/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prerequisites

Ensure the following components are installed before proceeding:

### Install Voyager Gateway

Install voyager gateway using the following command:
```bash
helm install ace oci://ghcr.io/appscode-charts/voyager-gateway \
  --version v2025.6.30 \
  -n ace-gw --create-namespace \
  --set gateway-converter.enabled=false \
  --wait --burst-limit=10000 --debug
```

### Create EnvoyProxy and GatewayClass

Deploy `EnvoyProxy` with the following configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ace
  namespace: ace-gw
spec:
  logging:
    level:
      default: warn
  mergeGateways: true
  provider:
    kubernetes:
      envoyDeployment:
        container:
          image: ghcr.io/voyagermesh/envoy:v1.34.1-ac
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            privileged: false
            runAsNonRoot: true
            runAsUser: 65534
            seccompProfile:
              type: RuntimeDefault
        patch:
          value:
            spec:
              template:
                spec:
                  containers:
                    - name: shutdown-manager
                      securityContext:
                        allowPrivilegeEscalation: false
                        capabilities:
                          drop:
                            - ALL
                        privileged: false
                        runAsNonRoot: true
                        runAsUser: 65534
                        seccompProfile:
                          type: RuntimeDefault
      envoyService:
        externalTrafficPolicy: Cluster
        type: LoadBalancer
    type: Kubernetes
```


> If you want to use `NodePort` service. Update `.spec.provider.kubernetes.envoyService.type` to `NodePort` in the above YAML.

Apply the configuration:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/envoyproxy.yaml
envoyproxy.gateway.envoyproxy.io/ace created
```

Deploy the `GatewayClass`:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  annotations:
    catalog.appscode.com/gateway-config: |-
      service:
        externalTrafficPolicy: Cluster
        nodeportRange: 30000-32767
        portRange: 10000-12767
        seedBackendPort: 8080
        type: LoadBalancer
      vaultServer:
        name: vault
        namespace: ace
    catalog.appscode.com/is-default-gatewayclass: "true"
  name: ace
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  description: Default Service GatewayClass
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: ace
    namespace: ace-gw
```
Apply it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/gatewayclass.yaml
gatewayclass.gateway.networking.k8s.io/ace created
```

Check the `GatewayClass` status `True`.
```bash
$ kubectl get gatewayclass 
NAME   CONTROLLER                                      ACCEPTED   AGE
ace    gateway.envoyproxy.io/gatewayclass-controller   True       16s
```

### Install `FluxCD` in your cluster
Install `FluxCD` in your cluster using the following command:
```bash
helm upgrade -i flux2 \
  oci://ghcr.io/appscode-charts/flux2 \
  --version 2.15.0 \
  --namespace flux-system --create-namespace \
  --wait --debug --burst-limit=10000
```

###  Install Keda

Install `Keda` in your cluster using the following command:
```bash
$ kubectl create ns kubeops
namespace/kubeops created

$ kubectl apply -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/helmrepo.yaml
helmrepository.source.toolkit.fluxcd.io/appscode-charts-oci created

$ kubectl apply -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/keda.yaml
helmrelease.helm.toolkit.fluxcd.io/keda created
helmrelease.helm.toolkit.fluxcd.io/keda-add-ons-http created
```

### Install `Catalog Manager`

Install `Catalog Manager` in your cluster using the following command:
```bash
helm install catalog-manager oci://ghcr.io/appscode-charts/catalog-manager \
  --version=v2025.6.30 \
  -n ace --create-namespace \
  --set helmrepo.name=appscode-charts-oci \
  --set helmrepo.namespace=kubeops \
  --wait --burst-limit=10000 --debug
```

## Overview

KubeDB uses following crd fields to enable Redis Announce:

- `spec:`
    - `announce:`
        - `type`
        - `shards`
            - `endpoints`

Read about the fields in details in [redis concept](https://kubedb.com/docs/v2025.6.30/guides/redis/concepts/redis/)

## Redis Cluster with Announce

### Create DNS Records
You need to create `A` or `CNAME` records for each `Redis` pod that will be externally accessible. For example, with `3` shards and `2` replicas per shard, the DNS entries would be:

Example:
- `A/CNAME Record` for each Redis replicas with exposed Envoy Gateway `LoadBalancer/NodePort` IP/Host:
    - "rd0-0.kubedb.appscode"
    - "rd0-1.kubedb.appscode"
    - "rd1-0.kubedb.appscode"
    - "rd1-1.kubedb.appscode"
    - "rd2-0.kubedb.appscode"
    - "rd2-1.kubedb.appscode"


Below is the YAML for Redis Announce.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-announce
  namespace: demo
spec:
  version: 7.4.0
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
    announce:
      type: hostname
      shards:
        - endpoints:
            - "rd0-0.kubedb.appscode"
            - "rd0-1.kubedb.appscode"
        - endpoints:
            - "rd1-0.kubedb.appscode"
            - "rd1-1.kubedb.appscode"
        - endpoints:
            - "rd2-0.kubedb.appscode"
            - "rd2-1.kubedb.appscode"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Here,
- `.spec.cluster.announce.type` specifies preferred dns type. It can be hostname or ip.
- `.spec.cluster.announce.shards` specifies the DNS names for each shards in the replica set.
- `.spec.cluster.announce.shards.endpoints`  specifies the DNS names for each pod in the specific shard.

### Deploy Redis Cluster Announce

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/redis.yaml
redis.kubedb.com/redis-announce created
```

Now, wait until `redis-announce` has status `Ready`. i.e,

```bash
$ watch kubectl get rd -n demo
Every 2.0s: kubectl get rd -n demo
NAME            VERSION   STATUS   AGE
redis-announce   7.4.0     Ready    6m56s
```


Now, create `RedisBinding` object to configure the whole process.
```yaml                                                                                                                           
apiVersion: catalog.appscode.com/v1alpha1
kind: RedisBinding
metadata:
  name: redis-bind
  namespace: demo
spec:
  sourceRef:
    name: redis-announce
    namespace: demo                                                                                                               
```                                                                                                                               

```bash                                                                                                                           
$ kubectl create -f https://github.com/kubedb/docs/raw/master/docs/examples/redis/announce/binding.yaml 
redisbinding.catalog.appscode.com/redis-bind created                                                                          
```                                                                                                                               
Now, check the status of `redisbinding` objects and ops requests.

```bash
$ kubectl get redisbinding,rdopsrequest -n demo
NAME                                               SRC_NS   SRC_NAME           STATUS   AGE
redisbinding.catalog.appscode.com/redis-bind       demo     redis-announce      Current  3m28s

NAME                                                       TYPE       STATUS       AGE
redisopsrequest.ops.kubedb.com/redis-announce-jddiql        Announce   Successful   2m58s
```

### Connect to Redis as Cluster

To connect to the Redis replica set, you can use the following command:

Collect the announces from the `redis` object:
```bash
$ kubectl get redis -n demo redis-announce -ojson | jq .spec.cluster.announce
{
  "shards": [
    {
      "endpoints": [
        "rd0-0.kubedb.appscode:10050@10056",
        "rd0-1.kubedb.appscode:10051@10057"
      ]
    },
    {
      "endpoints": [
        "rd1-0.kubedb.appscode:10052@10058",
        "rd1-1.kubedb.appscode:10053@10059"
      ]
    },
    {
      "endpoints": [
        "rd2-0.kubedb.appscode:10054@10060",
        "rd2-1.kubedb.appscode:10055@10061"
      ]
    }
  ],
}
```

Connect with the database:

```bash
$ redis-cli -h rd0-0.kubedb.appscode -p 10050 -a <password> -c ping
PONG
```

Set data in different shards:

```bash
$ redis-cli -h rd1-0.kubedb.appscode -p 10051 -a <password> -c set batman appscode
-> Redirected to slot [13947] located at rd0-0.kubedb.appscode:10050
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete redisbinding -n demo redis-bind
kubectl delete rd -n demo redis-announce

kubectl delete gatewayclass ace
kubectl delete -n ace-gw envoyproxy ace

helm uninstall -n ace catalog-manager
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](https://kubedb.com/docs/v2025.6.30/setup/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe to our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).