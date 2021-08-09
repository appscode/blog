# Introducing `Panopticon`, a new kubernetes generic state metrics exporter in town

We are highly excited to introduce `Panopticon`, a generic kubernetes resource state metrics exporter. It comes with some exciting features and customization options.

## What is Panopticon
Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate metrics from both kubernetes native and custom resources. Generated metrics are exposed  in `/metrics` path according to prometheus metrics format.

## Background
When we wanted to collect state metrics from our product's (KubeDB, Stash and so many) custom resource, we didn't find any extising tool that would accomplish our needs. Kubernetes has a project called `kube-state-metrics` but we didn't find any features for collecting metrics from kubernetes custom resources then. Moreover it's metrics for kubernetes native resources were predefined and there was hardly any customization option that couldn't fulfill our demands.

Then we decided to build our own generic resource metrics exporter and named it `Panopticon` which can collect metrics from any kind of kubernetes resources. There is an interesting story about the name `Panopticon`. You can learn about that from this wikipedia <a href="https://en.wikipedia.org/wiki/Panopticon" target="_blank">url</a>.

## How Panopticon works
Panopticon watches a custom resource called `MetricsConfiguration` which holds the necessary configuration for generating our desired metrics. This custom resource consists of mainly two parts. One is `targetRef` which defines the targeted kubernetes resource for metrics collection. Another is `metrics` which holds our desired metrics that we want to generate from that targeted resources. We'll discuss about the custom resource breifly in the later section.

So, when a new `MetricsConfiguration` object is created, Panopticon gets the event and generates defined metrics for the targeted resources. It stores those metrics in an in-memory store say `metrics store`. When the `MetricsConfiguration` object is updated/deleted or any new targeted resource is created/updated/deleted, Panopticon syncs the new changes with it's metrics store. So metrics store always holds the updated information according to `MetricsConfiguration` object.

So, when `/metrics` path is hit, Panopticon serves the metrics from metrics store that is  already generated. In this way, Panopticon can serve metrics in an efficient way with low latency.

## How to generate metrics using Panopticon
Now let's see how can we generate metrics using Panopticon. At first, we need to deploy Panopticon helm chart which will be found <a href="https://github.com/kubeops/installer" target ="_blank">here</a>.

## What's next

## Support
