---
title: Announcing KubeDB v2023.12.21
date: "2023-12-21"
weight: 14
authors:
- Arnob kumar saha
tags:
- autoscaler
- database
- day-2-operations
- elasticsearch
- kafka
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- nodepool
- nodetopology
- percona-xtradb
- postgresql
- prometheus
- proxysql
- redis
- scheduler
- vertical-scaling
---

We are pleased to announce the release of [KubeDB v2023.12.21](https://kubedb.com/docs/v2023.12.21/setup/). This release was mainly focused on improving the kubedb-autoscaler feature. We also made changes to our grafana dashboards. This post lists all the changes done in this release since the last release. Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.12.21/README.md). Let’s see the changes done in this release.

## Improving KubeDB Autoscaler

Here is an overall workflow of the kubedb-autoscaler to better understand the problem, we solved in this release.
- The autoscaler operator watches the usages of compute resources (cpu, memory) & storage resources, And generates OpsRequest CR to automatically change the resources.
- The ops-manager operator then watches the created VerticalOpsRequest for compute resources, update db's statefulsets & evict the db pods.
- k8s scheduler see the updated resource requests in those pods, & find an appropriate node for scheduling.
- If k8s scheduler doesn't find appropriate node, cloud provider's cluster autoscaler (if enabled) scales one of the nodepool to make spaces for that pod.

This procedure works fine while up-scaling the compute resources. Some nodes from bigger nodepools will be automatically created by the cluster autoscaler whenever some scheduling issues occur.
But this procedure becomes very resource-intensive while down-scaling the compute resources. As the k8s scheduler sees some big nodes are already available for scheduling, & we are not forcing to choose a smaller node where these down-scaled pods could have been easily running.



























So to solve this issue, we need a way so that we can forcefully schedule those smaller pods into smaller nodepools.  We have introduced a new CRD to achieve it, called `NodeTopology`.
Here is an example NodeTopology CR :

```yaml
apiVersion: node.k8s.appscode.com/v1alpha1
kind: NodeTopology
metadata:
  name: gke-pools
spec:
  nodeSelectionPolicy: Taint
  topologyKey: "nodepool_type"
  nodeGroups:
    - topologyValue: tiny
      capacity:
        cpu: 4
        memory: 15Gi
    - topologyValue: small
      capacity:
        cpu: 8
        memory: 30Gi
    - topologyValue: medium
      capacity:
        cpu: 16
        memory: 60Gi
    - topologyValue: mid-large
      capacity:
        cpu: 32
        memory: 120Gi
    - topologyValue: large
      capacity:
        cpu: 64
        memory: 240Gi
```

It is a cluster-scoped resource. It supports two types of nodeSelectionPolicy : `LabelSelector`, `Taint`. Here is the general rule to choose between these two.

If you want to run the database pods in some dedicated nodes, and don't want to allow any other pods to be scheduled there, the `Taint` policy is appropriate for you. For other general cases, use `LabelSelector`.



It is also possible to schedule different types of db pods into different nodepools. Here is an example `MongoDB` CR yaml :
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-database
  namespace: demo
spec:
  version: "4.4.26"
  terminationPolicy: WipeOut
  replicas: 3
  replicaSet:
    name: "rs"
  podTemplate:
    spec:
      nodeSelector:
        app: kubedb
        instance: mongodb
        component: mg-database
      tolerations:
        - effect: NoSchedule
          key: app
          operator: Equal
          value: kubedb
        - effect: NoSchedule
          key: instance
          operator: Equal
          value: mongodb
        - effect: NoSchedule
          key: component
          operator: Equal
          value: mg-database
        - key: nodepool_type
          value: tiny
          effect: NoSchedule
      resources:
        requests:
          "cpu": 2100m
          "memory": 8Gi
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
```
























Lastly, for autoscaling, all we need is to specify the name of the nodeTopology in the autoscaler yaml.


IMPORTANT : The node pool sizes, the starting resource requests, and the auto scaler configuration must be carefully choreographed for optimal behavior.

- The node pool sizes should be 4x bigger than the previous size.
- The database's initial requested resources should be slightly larger than 1/2 the intended node's capacity.
- The minimum allowed size by the Autoscaler resource should be 1/2 of the smallest node's capacity.
- 
```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: compute-as
  namespace: demo
spec:
  databaseRef:
    name: mg-database
  opsRequestOptions:
    timeout: 20m
    apply: IfReady
  compute:
    replicaSet:
      trigger: "On"
      podLifeTimeThreshold: 30m
      resourceDiffPercentage: 200
      minAllowed:
        cpu: 2
        memory: 7.5Gi
      maxAllowed:
        cpu: 64
        memory: 240Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
    nodeTopologyRef:
      name: gke-pools
```


Now, kubedb-autoscaler operator will decide what is the minimum node-configuration for the scaled (up or down) pods to be scheduled. And create the `VerticalScale` opsRequest specifying the tolerations so that other nodepools don't tolerate these newly created pods.



```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-database
  verticalScaling:
    replicaSet:
      resources:
        requests:
          memory: "16Gi"
          cpu: "4200m"
        limits:
          memory: "16Gi"
      nodeSelectionPolicy: Taint
      topology:
        key: nodepool_type
        value: small
```









## Alerts in Grafana Dashboards
In older releases, users had to manually import the grafana-dashboard panels in the grafana UI for visualization. In this release, we make this whole procedure automatic with helm charts. We have also added option to integrate alert panels on those dashboards.  The older approach is also available [here](https://github.com/appscode/grafana-dashboards/).

```bash
helm repo update
helm install <some-name> appscode/kubedb-grafana-dashboards -f overwrite.yaml
```

Here is an example `overwrite.yaml` file for mongodb, with name `simple` in `demo` namespace, with alerts enabled.

```yaml
resources:
  - mongodb

dashboard:
  folderID: 0
  overwrite: true
  alerts: true  # you can import the dashboards without alerts also
  replacements: {}
#    job=\"kube-state-metrics\": job=\"kubernetes-service-endpoints\"
#    job=\"kubelet\": job=\"kubernetes-nodes-cadvisor\"
#    job=\"$app-stats\": job=\"kubedb-databases\"

grafana:
  version: 8.0.7
  url: "<>" # example: http://grafana.monitoring.svc:80
  apikey: "<grafana-api-key>" # write permission needed

app:
  name: "simple"  # db name
  namespace: "demo" # db namespace
```

The `dashboard.replacements` section is only useful when you are using the builtin prometheus as datasource, not the well-known `kube-prometheus-stack` chart. In that case, you need to comment out this `replacements` part.


## Deprecating older DB versions

We have deprecated all the older patch versions for `MySQL`, `MariaDB` & `Postgres` in this release.

Here is the list of available versions now :

`MySQL` :  "8.2.0", "8.1.0", "8.0.35", "8.0.31-innodb", "5.7.44"

`MariaDB`: "11.2.2", "11.1.3", "11.0.4", "10.11.6", "10.10.7", "10.6.16", "10.5.23", "10.4.32"

`Postgres`: "16.1-bookworm", "16.1", "15.5-bookworm", "15.5", "14.10-bookworm", "14.10", "timescaledb-2.5.0-pg14.1", "14.1-bullseye-postgis", "13.13-bookworm", "13.13", "13.5-bullseye-postgis", "timescaledb-2.1.0-pg13", "12.17-bookworm", "12.17", "12.9-bullseye-postgis", "timescaledb-2.1.0-pg12", "11.22-bookworm", "11.22", "11.14-bullseye-postgis", "timescaledb-2.1.0-pg11", "10.23-bullseye", "10.23"



## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.12.21/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.12.21/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
