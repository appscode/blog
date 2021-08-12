---
title: Introducing Panopticon, A Generic Kubernetes State Metrics Exporter
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

We are highly excited to introduce `Panopticon`, a generic Kubernetes resource state metrics exporter. It comes with a lot of exciting features and customization options.

## What is Panopticon?
Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape them.

## Background
When we wanted to collect state metrics from our product's (KubeDB, Stash and so many) custom resources, we didn't find any existing tool that would accomplish our needs. Kubernetes has a project called `kube-state-metrics` but it does not support collecting metrics from Kubernetes custom resources. Moreover, the metrics for Kubernetes native resources were predefined and there was hardly any customization option.

So, we decided to build our own generic resource metrics exporter and named `Panopticon` which can collect metrics from any kind of Kubernetes resources. There is an interesting story about the name `Panopticon`. You can learn about that from this Wikipedia [page](https://en.wikipedia.org/wiki/Panopticon).

## How Panopticon works
Panopticon watches a custom resource called `MetricsConfiguration` which holds the necessary configuration for generating our desired metrics. This custom resource consists of mainly two parts. The first one is `targetRef` which defines the targeted Kubernetes resource for metrics collection. The other is `metrics` which holds desired metrics that we want to generate from that targeted resources. We'll discuss the custom resource briefly in the later section.

When a new `MetricsConfiguration` object is created, Panopticon gets the event and generates defined metrics for the targeted resources. It stores those metrics in an in-memory store say `metrics store`. When the `MetricsConfiguration` object is updated/deleted or any new instance of the targeted resource is created/updated/deleted, Panopticon syncs the new changes with its metrics store. So metrics store always holds the updated information according to the `MetricsConfiguration` object.

When the `/metrics` path is hit, Panopticon serves the metrics from metrics store that is already generated. In this way, Panopticon efficiently serves metrics with low latency.

## How to generate metrics using Panopticon
Now let's see how can we generate metrics using Panopticon. At first, we need to deploy the Panopticon helm chart which will be found [here](https://github.com/kubeops/installer).

After that, let's see a sample `MetricsConfiguration` object for our MongoDB custom resource.

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

Like other Kubernetes native resources, it has `TypeMeta`, `ObjectMeta`, and `Spec` sections. However, it doesn't have a `Status` section. Let's focus on the `spec` section. In `targetRef`, we specified the `apiVersion` and `kind` of our targeted resource `MongoDB` from which we want to generate our metrics. The `metrics` section specifies the list of metrics we want to collect.

// TODO: field list

Here each metrics contains three mandatory fields. They are `name`, `help`, and `type`. Here `name` defines the metrics name, `help` holds a short description about the metrics and `type` denotes the Prometheus type of the metrics. You'll find the details of custom resource definition [here](https://github.com/kmodules/custom-resources/blob/master/apis/metrics/v1alpha1/metricsconfiguration_types.go).  

Let's see a sample MongoDB manifest file for better understanding.

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

From the above MongoDB instance, the first metrics `kubedb_mongodb_info` in MetricsConfiguration will collect some basic information and will set them as labels. As Prometheus metrics must contain a metrics value, we set the value as 1 here.

Next metrics `kubedb_mongodb_status_phase` is more interesting. This metrics will represent the MongoDB instance's current phase. The interesting part here, MongoDB instance can have six different phases called 'Ready', 'Critical', 'NotReady' etc. So, to understand the MongoDB instance's current phase properly, we need metrics for all of those phases.

To handle this type of scenario, there is one field in MetricsConfiguration called 'states' which holds the label key and all possible label values. It also contains the corresponding configuration to find the value of the metrics. `metricsValue` can have three different fields. One is `value` which denotes the direct value of that metrics like the first metrics. Another one is `valueFromPath` which denotes the json path of the resource instance. The final one is `valueFromExpression` which is used in this metrics. It contains a predefined function that will take the necessary parameters from the `params` field and will calculate the value of the metrics accordingly. Here if the phase matches with the given phase, the 'int' function will return 1 otherwise 0. You'll get expression functions and their uses [here](https://github.com/kmodules/custom-resources/blob/master/apis/metrics/v1alpha1/metricsconfiguration_types.go). Finally we will have six different metrics similar to below:

```
kubedb_mongodb_status_phase { ..., phase="Ready"}          1
kubedb_mongodb_status_phase { ..., phase="Halted"}         0 
kubedb_mongodb_status_phase { ..., phase="Provisioning"}   0
kubedb_mongodb_status_phase { ..., phase="Critical"}       0
kubedb_mongodb_status_phase { ..., phase="NotReady"}       0
kubedb_mongodb_status_phase { ..., phase="DataRestoring"}  0
```
Note: Here, we assume MongoDB instance's phase as "Ready".

Similarly, we can collect various kinds of metrics not only from our custom resources but also from any Kubernetes native resources with just a MetricsConfiguration object.

## What's next
// TODO

## Support
To speak with us, please leave a message on our [website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/AppsCodeHQ).
