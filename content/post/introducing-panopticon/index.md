---
title: Introducing Panopticon, a new kubernetes generic state metrics exporter
date: 2021-08-16
weight: 22
authors:
  - Pulak Kanti Bhowmick
tags:
  - monitoring
  - prometheus
  - grafana
  - kubernetes
  - kubedb
  - metrics
  - prometheus-exporter
  - kubernetes-exporter
---

We are highly excited to introduce `Panopticon`, a generic kubernetes resource state metrics exporter. It comes with a lot of exciting features and customization options.

## What is Panopticon?
Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate metrics from both kubernetes native and custom resources. Generated metrics are exposed  in `/metrics` path according to prometheus metrics format.

## Background
When we wanted to collect state metrics from our product's (KubeDB, Stash and so many) custom resource, we didn't find any extising tool that would accomplish our needs. Kubernetes has a project called `kube-state-metrics` but we didn't find any features for collecting metrics from kubernetes custom resources then. Moreover it's metrics for kubernetes native resources were predefined and there was hardly any customization option that couldn't fulfill our demands.

Then we decided to build our own generic resource metrics exporter and named it `Panopticon` which can collect metrics from any kind of kubernetes resources. There is an interesting story about the name `Panopticon`. You can learn about that from this wikipedia [page](https://en.wikipedia.org/wiki/Panopticon).

## How Panopticon works
Panopticon watches a custom resource called `MetricsConfiguration` which holds the necessary configuration for generating our desired metrics. This custom resource consists of mainly two parts. One is `targetRef` which defines the targeted kubernetes resource for metrics collection. Another is `metrics` which holds our desired metrics that we want to generate from that targeted resources. We'll discuss about the custom resource breifly in the later section.

So, when a new `MetricsConfiguration` object is created, Panopticon gets the event and generates defined metrics for the targeted resources. It stores those metrics in an in-memory store say `metrics store`. When the `MetricsConfiguration` object is updated/deleted or any new targeted resource is created/updated/deleted, Panopticon syncs the new changes with it's metrics store. So metrics store always holds the updated information according to `MetricsConfiguration` object.

So, when `/metrics` path is hit, Panopticon serves the metrics from metrics store that is  already generated. In this way, Panopticon can serve metrics in an efficient way with low latency.

## How to generate metrics using Panopticon
Now let's see how can we generate metrics using Panopticon. At first, we need to deploy Panopticon helm chart which will be found [here](https://github.com/kubeops/installer).

After that, let's see a simple `MetricsConfiguration` object for our MongoDB custom resource.

```yaml
apiVersion: metrics.appscode.com/v1alpha1
kind: MetricsConfiguration
metadata:
  name: kubedb-com-mongodb
spec:
  targetRef:
    apiVersion: kubedb.com/v1alpha2
    kind: MongoDB
  metrics:
    - name: kubedb_mongodb_info
      help: "Kubedb mongodb instance info"
      type: gauge
      labels:
        - key: sslMode
          valuePath: .spec.sslMode
        - key: storageType
          valuePath: .spec.storageType
        - key: terminationPolicy
          valuePath: .spec.terminationPolicy
        - key: version
          valuePath: .spec.version
      metricValue:
        value: 1
    - name: kubedb_mongodb_status_phase
      help: "Mongodb instance current phase"
      type: gauge
      field:
        path: .status.phase
        type: String
      params:
        - key: phase
          valuePath: .status.phase
      states:
        labelKey: phase
        values:
          - labelValue: Ready
            metricValue:
              valueFromExpression: "int(phase == 'Ready')"
          - labelValue: Halted
            metricValue:
              valueFromExpression: "int(phase == 'Halted')"
          - labelValue: Provisioning
            metricValue:
              valueFromExpression: "int(phase == 'Provisioning')"
          - labelValue: Critical
            metricValue:
              valueFromExpression: "int(phase == 'Critical')"
          - labelValue: NotReady
            metricValue:
              valueFromExpression: "int(phase == 'NotReady')"
          - labelValue: DataRestoring
            metricValue:
              valueFromExpression: "int(phase == 'DataRestoring')"
```

Here apiVersion, kind and metadata is specified as usual. Let's focus on spec section. In `targetRef`, we specified the apiVersion and kind of our targeted resource `MongoDB` from which we want to generate our metrics. After that, there is a list of metrics. 

Here each metrics contains three mandatory fields. They are `name`, `help` and `type`. Here `name` defines the metrics name, `help` holds a short description about the metrics and `type` denotes the Prometheus type of the metrics. You'll find the details custom resource defination [here](https://github.com/kmodules/custom-resources/blob/master/apis/metrics/v1alpha1/metricsconfiguration_types.go).  

Let's see a MongoDB manifest file for better understanding.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongodb-demo
  namespace: demo
spec:
  version: "4.2.3"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

First metrics `kubedb_mongodb_info` in MetricsConfiguration, will collect some basic information from above MongoDB instance and will set them as labels. As prometheus metrics must contain a metrics value, we set the value as 1 here.

Next metrics `kubedb_mongodb_status_phase` is more interesting. This metrics will represent MongoDB instance current phase. The interesting part here, MongoDB instance can have six different phase called 'Ready', 'Critical', 'NotReady' etc. So, to understand the MongoDB instance current phase properly, we need individual metrics for all of those phases. 

To handle same kind of scenario, there is one field in MetricsConfiguration called 'states' which holds the label key and all possible label values. It also contains the corresponding configuration to find the metrics value. `metricsValue` can have three different fields. One is `value` which denotes the direct value of that metrics like first metrics. Another one is `valueFromPath` which denotes the json path of resource instance. Final one is `valueFromExpression` which is used in this metrics. It contains a predefined function which will take the necessary parameters from the `params` field and will calculate the metrics value accordingly. Here if phase matches with the given phase, 'int' function will return 1 otherwise 0. You'll get expression functions and their uses [here](https://github.com/kmodules/custom-resources/blob/master/apis/metrics/v1alpha1/metricsconfiguration_types.go). Finally we will have six different metrics similar like below:

```
kubedb_mongodb_status_phase { ..., phase="Ready"}          1
kubedb_mongodb_status_phase { ..., phase="Halted"}         0 
kubedb_mongodb_status_phase { ..., phase="Provisioning"}   0
kubedb_mongodb_status_phase { ..., phase="Critical"}       0
kubedb_mongodb_status_phase { ..., phase="NotReady"}       0
kubedb_mongodb_status_phase { ..., phase="DataRestoring"}  0
```
Note: Here, we assume MongoDB instanse's phase as "Ready".

In similar way, we can collect various kind of metrics not only from our custom resources but also from any kubernetes native resources with just a MetricsConfiguration object.

## Support
To speak with us, please leave a message on our [website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/AppsCodeHQ).
