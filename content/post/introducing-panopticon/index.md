---
title: Introducing Panopticon, A Generic Kubernetes State Metrics Exporter
date: 2021-08-23
weight: 22
authors:
  - Pulak Kanti Bhowmick
tags:
  - cloud-native
  - database
  - elasticsearch
  - grafana
  - kubedb
  - kubernetes
  - kubernetes-exporter
  - mariadb
  - memcached
  - metrics
  - mongodb
  - monitoring
  - mysql
  - panopticon
  - postgresql
  - prometheus
  - prometheus-exporter
  - redis
---

We are excited to introduce `Panopticon`, a generic Kubernetes resource state metrics exporter. It comes with a lot of features and customization options.

## What is Panopticon?

Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in `/metrics` path for the Prometheus server to scrape.

## Background

We wanted to collect state metrics from our various products (eg, KubeDB, Stash and other). But we didn't find any existing tool that would accomplish our needs. Kubernetes has a project called [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) but it does not support collecting metrics from Kubernetes custom resources. Moreover, the metrics for Kubernetes native resources were predefined and there was hardly any customization options.

So, we decided to build our own generic resource metrics exporter, named `Panopticon`, which can collect metrics from any kind of Kubernetes resources. Like the real [Panopticon](https://en.wikipedia.org/wiki/Panopticon), `Panopticon` is a Kubernetes controller that watches Kubernetes resources passively and exports Prometheus metrics.

## How Panopticon works

Panopticon introduces a custom resource called `MetricsConfiguration` which holds the necessary configuration for generating metrics. This custom resource consists of mainly two parts. The first one is `targetRef` which defines the targeted Kubernetes resource for metrics collection. The other is `metrics` which holds desired metrics that we want to generate from that targeted resources. We'll discuss the custom resource briefly in the later section.

When a new `MetricsConfiguration` object is created, Panopticon gets the event and generates defined metrics for the targeted resources. It stores those metrics in an in-memory metrics store. When the `MetricsConfiguration` object is updated/deleted or any new instance of the targeted resource kind is created/updated/deleted, Panopticon syncs the new changes with its metrics store. So metrics store always holds the updated information according to the `MetricsConfiguration` object.

When the `/metrics` path is scraped, Panopticon serves the metrics from metrics store that is already generated. In this way, Panopticon efficiently serves metrics with low latency.

Let's know about `metrics` fields briefly.

| Field  | Required?  | Has SubFields? | Description  |
|---|---|---|---|
| name | yes  | no |`name` defines the metrics name. `name` should be in snake case. Example: `name: kube_deployment_spec_replicas`  |
| help | yes  | no |`help` is used to describe the metrics. For `kube_deployment_spec_replicas`, the `help` string can be "Number of desired pods for a deployment.  |
| type | yes  |  no |`type` defines the Prometheus type of the metrics. For Kubernetes based objects, `type` can only be "gauge" |
| field | no | yes | `field` contains the information of the field for which metric is collected. It has two sub-fields: `path` and `type`. `path` defines the json path of the object. Example: For deployment spec replica count, the path will be `.spec.replicas`. `type` defines the type of the value in the given `path`. `type` can be "Integer" for integer value like `.spec.replicas`, "DateTime" for time stamp value like `.metadata.creationTimestamp`. "Array" for array field like `.spec.containers`. "String" for string field like .statue.phase (for pod status). When some labels are collected with metric value 1 and the values are not from an array then `field` can be skipped. Otherwise, `field` must be specified. |
| labels | no  | yes | `labels` contains the information of a metric label. Given labels are always added in the metrics along with resource name and namespace. Resource's name and namespace are always added to the labels by default. No configuration is needed for name and namespace labels. It has three subfields. They are `key`, `value`, `valuePath`. `key` defines the label key. `value` defines the hardcoded label value. `valuePath` defines the label value path. Either `value` or `valuePath` must be specified for a Label.  If both are specified, `valuePath` is ignored. Note that, if `key` is not specified for a label and the given `valuePath` is invalid or doesn't exist for the resource, the label will be ignored. |
| params | no  | yes  | `params` is the list of parameters configuration used in expression evaluation. The parameter should contain a user-defined `key` and corresponding `value` or `valuePath`. Either `value` or `valuePath` must be specified. If both are specified, `valuePath` will be ignored.|
| states | conditionally required | yes | `states` contains the configuration for generating all the time series of a metric with label cardinality is greater than 1. `states` specify the possible states for a label and their corresponding MetricValue configuration. `metrics` must contain either `states` or `metricValue`. If both are specified, `metricsValue` will be ignored. It contains `labelKey` and `values`. `values` contain the list of state values. The size of the list is always equal to the cardinality of that label.|
| metricValue | conditionally required | yes | `metricValue` defines the configuration to obtain the metric value. `metricValue` contains only one of following fields: `value`, `valueFromPath`, and `valueFromExpression`. If multiple fields are assigned then only one field is considered and other fields are ignored. The priority rule is: "Value > ValueFromPath > ValueFromExpression". `value` contains the metric value. It is defined as "1" when some information of the object is collected as labels but there is no specific metric value. `valueFromPath` contains the field path of the manifest file of an object. `valueFromPath` is used when the metric value is coming from any specific json path of the object. Example: For metrics "kube_deployment_spec_replicas", the `metricValue` is coming from a specific path `.spec.replicas`. In this case, valueFromPath is defined as `valueFromPath: .spec.replicas`. `valueFromExpression` contains an expression function to evaluate the metric value. `params` is used to evaluate the expression.|

Available expression evaluation functions are:

| Function Definition | Description |
|---|---|
| int(expression) | Returns 1 if the expression is true otherwise 0. Example: int(phase == 'Running'), here `phase` is an argument which holds the `phase` of a Kubernetes resource|
| percentage(percent, total, roundUp) | `percent` can represent a percent(%) value or can be an Integer value. In the case of the percent(%) value, it will return the value of `(percent * total%)` and for the Integer value, it will simply return `percent` without any modification. `roundUp` is optional and contains a boolean value. By default its value is `false`. If the `roundUp` is `true`, the resultant value will be rounded up otherwise not.  Example: To get the maximum number of unavailable replicas of a deployment at the time of rolling update, we can use `percentage(maxUnavailable, replicas, false)` or `percentage(maxUnavailable, replicas)`. Here, the value of `maxUnavaiable` will be obtained from `.spec.strategy.rollingUpdate.maxUnavailable` path of the deployment and `replicas` represents the number of spec replica count |
| cpu_cores(arg) | Returns the CPU value in core. Let, cpuVal=500m then cpu_cores(cpuVal) will return 0.5. |
| bytes(arg) | Returns the memory value in byte. Let, memVal=1 Ki then bytes(memVal) will return 1024. |
| unix (arg) | Converts the DateTime string into unix and returns it. |
| resource_replicas(obj) | Takes Kubernetes object as input and returns it's replica count. |
| resource_mode(obj) | Takes Kubernetes object as input and returns it's mode. To get the MongoDB's mode(Standalone/ReplicaSet/Sharded), use: `resource_mode(MongoDB resource object)` |
| total_resource_limits(obj, resourceType) | Takes Kubernetes object as input and returns it's resource limits according to `resourceType`. `resourceType` can be `cpu`, `memory`, and `storage`. To get the MongoDB memory limit, use: `total_resource_limits(MongoDB resource object, "memory")`. |
| total_resource_requests(obj, resourceType) | Takes Kubernetes object as input and returns it's resource requests according to `resourceType`. `resourceType` can be `cpu`, `memory`, and `storage`. To get the MongoDB cpu request, use: `total_resource_limits(MongoDB resource object, "cpu")`. |
| app_resource_limits(obj, resourceType) | Takes Kubernetes object as input and returns the main application containers (excluding supporting sidecars like Prometheus exporters, etc.) resource limits according to `resourceType`. `resourceType` can be `cpu`, `memory`, and `storage`. To get the MongoDB database memory limit, use: `app_resource_limits(MongoDB resource object, "memory")`. |
| app_resource_requests(obj, resourceType) | Takes Kubernetes object as input and returns the main application containers (excluding supporting sidecars like Prometheus exporters, etc.) resource requests according to `resourceType`. `resourceType` can be `cpu`, `memory`, and `storage`. To get the MongoDB database cpu request, use: `app_resource_limits(MongoDB resource object, "cpu")`. |

Note: To know about CRD definition and evaluation functions in details, please visit [here](https://github.com/kmodules/custom-resources/blob/master/apis/metrics/v1alpha1/metricsconfiguration_types.go).

## How to install Panopticon

At first, we need to deploy the Panopticon helm chart which will be found [here](https://github.com/kubeops/installer). You will need to get a license key that can be found [here](https://license-issuer.appscode.com/?p=panopticon-enterprise).

```bash
helm repo add appscode https://charts.appscode.com/stable/
helm repo update

helm install panopticon appscode/panopticon \
    -n kubeops --create-namespace \
    --set-file license=/path/to/license.txt
```

## How to generate metrics using Panopticon

Now, let's see a sample `MetricsConfiguration` object for our MongoDB custom resource.

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
    - name: kubedb_mongodb_created
      help: "MongoDB creation timestamp in unix"
      type: gauge
      field:
        type: DateTime
        path: .metadata.creationTimestamp
      metricValue:
        valueFromPath: .metadata.creationTimestamp

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

    - name: kubedb_mongodb_replicas
      help: "Number of available replicas for MongoDB"
      type: gauge
      params:
        - key: obj
          valuePath: .
      metricValue:
        valueFromExpression: resource_replicas(obj)

    - name: kubedb_mongodb_resource_request_cpu
      help: "Requested CPU by MongoDB in core"
      type: gauge
      labels:
        - key: unit
          value: core
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: cpu
      metricValue:
        valueFromExpression: total_resource_requests(obj, resourceType)

    - name: kubedb_mongodb_resource_request_memory
      help: "Requested memory by MongoDB in byte"
      type: gauge
      labels:
        - key: unit
          value: byte
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: memory
      metricValue:
        valueFromExpression: total_resource_requests(obj, resourceType)

    - name: kubedb_mongodb_resource_request_storage
      help: "Requested storage by MongoDB in byte"
      type: gauge
      labels:
        - key: unit
          value: byte
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: storage
      metricValue:
        valueFromExpression: total_resource_requests(obj, resourceType)

    - name: kubedb_mongodb_resource_limit_cpu
      help: "CPU limit for MongoDB in core"
      type: gauge
      labels:
        - key: unit
          value: core
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: cpu
      metricValue:
        valueFromExpression: total_resource_limits(obj, resourceType)

    - name: kubedb_mongodb_resource_limit_memory
      help: "Memory limit for MongoDB in byte"
      type: gauge
      labels:
        - key: unit
          value: byte
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: memory
      metricValue:
        valueFromExpression: total_resource_limits(obj, resourceType)

    - name: kubedb_mongodb_resource_limit_storage
      help: "Storage limit for MongoDB in byte"
      type: gauge
      labels:
        - key: unit
          value: byte
      params:
        - key: obj
          valuePath: .
        - key: resourceType
          value: storage
      metricValue:
        valueFromExpression: total_resource_limits(obj, resourceType)
```

Like other Kubernetes native resources, `MetricsConfiguration` has `TypeMeta`, `ObjectMeta`, and `Spec` sections. However, it doesn't have a `Status` section. It is a cluster scoped resource and we recommend naming the object with the `{targetGroup}-{targetResourceSingular}`. Let's focus on the `spec` section. In `spec.targetRef`, we specified the `apiVersion` and `kind` of our targeted resource `MongoDB` from which we want to generate our metrics. The `spec.metrics` section specifies the list of metrics we want to collect.

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

After deploying, let's get the yaml using below command:

```bash
kubectl get mongodb mongodb-demo -n demo -o yaml
```

You'll find something like below. Some irrelevant fields are not shown here.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  creationTimestamp: "2021-08-16T11:37:13Z"
  finalizers:
  - kubedb.com
  generation: 2
  ...
  ...
  name: mongodb-demo
  namespace: demo
  resourceVersion: "169783"
  uid: 1c7f3eaa-c038-40a9-8745-07d2b5f6aaf2
spec:
  authSecret:
    name: mongodb-demo-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      ...
      ...
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mongodb-demo
  replicas: 1
  sslMode: disabled
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageEngine: wiredTiger
  storageType: Durable
  terminationPolicy: WipeOut
  version: 4.2.3
status:
  conditions:
  - lastTransitionTime: "2021-08-16T11:37:13Z"
    message: "The KubeDB operator has started the provisioning of MongoDB: demo/mongodb-demo"
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2021-08-16T11:37:26Z"
    message: "All desired replicas are ready."
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2021-08-16T11:37:57Z"
    message: "The MongoDB: demo/mongodb-demo is accepting client requests."
    observedGeneration: 2
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2021-08-16T11:37:57Z"
    message: "The MongoDB: demo/mongodb-demo is ready."
    observedGeneration: 2
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2021-08-16T11:37:57Z"
    message: "The MongoDB: demo/mongodb-demo is successfully provisioned."
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready
```

From the above MongoDB instance, the first metrics `kubedb_mongodb_created` in MetricsConfiguration will collect the MongoDB resource creation time in unix. the second one, `kubedb_mongodb_info` in MetricsConfiguration will collect some basic information and will set them as labels. As Prometheus metrics must contain a metrics value, we set the value as 1 here.

Next metrics `kubedb_mongodb_status_phase` is more interesting. This metrics will represent the MongoDB instance's current phase. The interesting part here, MongoDB instance can have six different phases called 'Ready', 'Critical', 'NotReady' etc. So, to understand the MongoDB instance's current phase properly, we need metrics for all of those phases.

To handle this type of scenario, there is one field in MetricsConfiguration called `states` which holds the label key and all possible label values. It also contains the corresponding configuration to find the value of the metrics. Here if the actual phase of an resource matches with the given phase, the `int` function will return 1 otherwise 0. Finally we will have six different metrics similar to below:

```
kubedb_mongodb_status_phase { ..., phase="Ready"}          1
kubedb_mongodb_status_phase { ..., phase="Halted"}         0 
kubedb_mongodb_status_phase { ..., phase="Provisioning"}   0
kubedb_mongodb_status_phase { ..., phase="Critical"}       0
kubedb_mongodb_status_phase { ..., phase="NotReady"}       0
kubedb_mongodb_status_phase { ..., phase="DataRestoring"}  0
```

Note: Here, we assume MongoDB instance's phase as "Ready".

The next metrics `kubedb_mongodb_replica` represents MongoDB replica count. It will calculate the number of replicas according to the given MongoDB object. Here, we send the full MongoDB object in the `params` and `resource_replicas` function to calculate the total number of replicas according to its mode(Standalone/Replicaset/Sharded).

The next metrics `kubedb_mongodb_resource_request_cpu` represents the requested CPU value by MongoDB in core. `total_resource_requests` function will get the full MongoDB object and `resourceType` from `params`. Then it will calculate the requested amount of that resource accordingly.

The next metrics `kubedb_mongodb_resource_request_memory` represents the requested memory value by MongoDB in byte. In this case, we also use `total_resource_requests` function but this time in `params` resourceType is specified as 'memory".

The next metrics `kubedb_mongodb_resource_request_storage` is similar to the previous two metrics. In this case, we have to specify `resourceType` as "storage"

To calculate the next three metrics `kubedb_mongodb_resource_limit_cpu`, `kubedb_mongodb_resource_limit_memory`, and `kubedb_mongodb_resource_limit_storage`, we use `total_resource_limits` function. This function takes the full MongoDB object and resource type in `params` and calculates the resource limit according to the resource type.

Now let's see a sample `MetricsConfiguration` object for Kubernetes native resource `Deployment`. All metrics for Deployment collected by "kube-state-metrics" are collected below using `Panopticon`. You can see "kube-state-metrics" project's configuration for deployment [here](https://github.com/kubernetes/kube-state-metrics/blob/master/internal/store/deployment.go).

```yaml
apiVersion: metrics.appscode.com/v1alpha1
kind: MetricsConfiguration
metadata:
  name: apps-deployment
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
  metrics:
    - name: kube_deployment_created
      help: "Unix creation timestamp"
      type: gauge
      field:
        path: .metadata.creationTimestamp
        type: DateTime
      metricValue:
        valueFromPath: .metadata.creationTimestamp

    - name: kube_deployment_status_replicas
      help: "The number of replicas per deployment"
      type: gauge
      field:
        path: .status.replicas
        type: Integer
      metricValue:
        valueFromPath: .status.replicas

    - name: kube_deployment_status_replicas_ready
      help: "The number of available replicas per deployment."
      type: gauge
      field: 
        path: .status.readyReplicas
        type: Integer
      metricValue: 
        valueFromPath: .status.readyReplicas

    - name: kube_deployment_status_replicas_available
      help: "The number of available replicas per deployment."
      type: gauge
      field: 
        path: .status.availableReplicas
        type: Integer
      metricValue:
        valueFromPath: .status.availableReplicas

    - name: kube_deployment_status_replicas_updated
      help: "The number of updated replicas per deployment."
      type: gauge
      field: 
        path: .status.updatedReplicas
        type: Integer
      metricValue:
        valueFromPath: .status.updatedReplicas

    - name: kube_deployment_status_observed_generation
      help: "The generation observed by the deployment controller."
      type: gauge
      field:
        path: .status.observedGeneration
        type: Integer
      metricValue:
        valueFromPath: .status.observedGeneration

    - name: kube_deployment_status_condition
      help: "The current status conditions of a deployment."
      type: gauge
      field:
        path: .status.conditions
        type: Array
      labels:
        - key: type
          valuePath: .status.conditions[*].type
        - key: status
          valuePath: .status.conditions[*].status
      metricValue:
        value: 1

    - name: kube_deployment_spec_replicas
      help: "Number of desired pods for a deployment."
      type: gauge
      field:
        path: .spec.replicas
        type: Integer
      metricValue:
        valueFromPath: .spec.replicas

    - name: kube_deployment_spec_paused
      help: "Whether the deployment is paused and will not be processed by the deployment controller."
      type: gauge
      params:
        - key: paused
          valuePath: .spec.paused
      metricValue:
        valueFromExpression: int(paused == 'true')

    - name: kube_deployment_spec_strategy_rollingupdate_max_unavailable
      help: "Maximum number of unavailable replicas during a rolling update of a deployment."
      type: gauge
      params: 
        - key: replicas
          valuePath: .spec.replicas
        - key: maxUnavailable
          valuePath: .spec.strategy.rollingUpdate.maxUnavailable
      metricValue:
        valueFromExpression: percentage(maxUnavailable, replicas, false)

    - name: kube_deployment_spec_strategy_rollingupdate_max_surge
      help: "Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a deployment."
      type: gauge
      params: 
        - key: replicas
          valuePath: .spec.replicas
        - key: maxSurge
          valuePath: .spec.strategy.rollingUpdate.maxSurge
      metricValue:
        valueFromExpression: percentage(maxSurge, replicas, true)

    - name: kube_deployment_metadata_generation
      help: "Sequence number representing a specific generation of the desired state."
      type: gauge
      field:
        path: .metadata.generation
        type: Integer
      metricValue: 
        valueFromPath: .metadata.generation
```

Similarly, we can collect various kinds of metrics not only from our custom resources but also from any Kubernetes native resources with just a MetricsConfiguration object.

## Webinar

We are delighted to announce a webinar on 26 August 2021. In this webinar, our experts of AppsCode will talk on “Panopticon: A Generic Kubernetes State Metrics Exporter” and demonstrate how to generate Prometheus metrics from Kubernetes native and custom resources.

Check here for details: https://appscode.com/webinar and don’t forget to register!


## Support

To speak with us, please leave a message on our [website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/AppsCodeHQ).
