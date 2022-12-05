---
title: Filter Panopticon Metrics by Namespaces
date: "2022-12-01"
weight: 25
authors:
- Pulak Kanti Bhowmick
tags:
- filtering
- kubedb
- kubernetes
- metrics
- monitoring
- namespace
- panopticon
- prometheus
---

Monitoring is an essential part of applications deployed in Kubernetes. We have a tool called `Panopticon` to collect metrics from Kubernetes native and custom resources. In this blog post, we'll briefly describe how you can configure `Panopticon` or `Prometheus Service Monitor` to filter Panopticon metrics by namespaces. 

Here is the outline of this post:

* Panopticon
* Install Panopticon in Kubernetes
* Filter metrics using Panopticon
* Filter metrics using Prometheus Service Monitor

## Panopticon

Panopticon is a generic state metrics exporter for Kubernetes resources. It can generate Prometheus metrics from both Kubernetes native and custom resources. Generated metrics are exposed in /metrics path for the Prometheus server to scrape. If you want to know more about Panopticon, you can check out [this blog post](https://blog.byte.builders/post/introducing-panopticon/).

## Install Panopticon in Kubernetes

Panopticon is an enterprise product. It needs a license to run it. To grab a trial license, please visit [here](https://license-issuer.appscode.com/?p=panopticon-enterprise).

If you already have an enterprise license for KubeDB or Stash, you do not need to issue a new license for Panopticon. Your existing KubeDB or Stash license will work with Panopticon. 

To install Panopticon in Kubernetes using helm please follow the below commands:

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm install panopticon appscode/panopticon -n kubeops \
    --create-namespace \
    --version v2022.06.14 \
    --set apiserver.enableValidatingWebhook=false \
    --set monitoring.enabled=true \
    --set monitoring.agent=prometheus.io/operator \
    --set monitoring.serviceMonitor.labels.release=<prometheus-service-monitor-selector-label> \
    --set-file license=/path/to/license-file.txt
```

## Filter metrics using Panopticon 

Panopticon provides a custom resource definition called `MetricsConfiguration`. It holds the target resource's group, version, kind, and a list of metrics to collect from the target resources. Be default, Panopticon collects metrics from resources from all namespaces. It was not possible before to restrict Panopticon to expose metrics for a subset of Kubernetes namespaces. 

Recently we have added namespace selector support in Panopticon. Now you can restrict Panopticon to expose metrics for a subset of Kubernetes namespaces. To do so, you have to set `namespaceSelector` value while installing Panopticon. `namespaceSelector` accepts any valid [Kubernetes label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) string. This feature can be used to expose only **Rancher project specific metrics** to the project Prometheus server.


```bash
$ helm install panopticon appscode/panopticon -n kubeops \
    --create-namespace \
    --version v2022.06.14 \
    --set monitoring.enabled=true \
    --set monitoring.agent=prometheus.io/operator \
    --set monitoring.serviceMonitor.labels.release=<prometheus-service-monitor-selector-label> \
    --set namespaceSelector='environment in (production)' \
    --set-file license=kubedb-license.txt
```

Pros: 
- Panopticon will only generate metrics for the resources in the selected namespaces. Other resource metrics are entirely skipped from Panopticon.
- Prometheus can't scrape other namespaces metrics as they are not generated.

Cons:
- As Panopticon only collects metrics from resources in the selected namespaces, resource metrics from other namespace is skipped.
- Users can not collect those metrics without changing the configuration. To collect metrics from other namespaces, users have to add a label to the namespace object according to `namespaceSelector`, upgrade helm installation with updated `namespaceSelector` or deploy Panopticon separetely using a different `namespaceSelector`.

## Filter metrics using Prometheus Service Monitor

If you are running Prometheus in Kubernetes for scraping metrics, you can filter metrics by namespaces using `metricsRelabelings` configuration in Service Monitor. In this method, you don't need to provide any namespaceSelector flag while installing Panopticon. So, Panopticon will collect metrics from all the namespaces.

Panopticon helm chart comes with a Service Monitor configuration out of the box.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    release: prometheus
  name: panopticon
  namespace: kubeops
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 10s
    port: api
    relabelings:
    - action: labeldrop
      regex: (pod|service|endpoint|namespace)
    scheme: https
    tlsConfig:
      ca:
        secret:
          key: tls.crt
          name: panopticon-apiserver-cert
      serverName: panopticon.kubeops.svc
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 10s
    port: telemetry
    scheme: http
  namespaceSelector:
    matchNames:
    - kubeops
  selector:
    matchLabels:
      app.kubernetes.io/instance: panopticon
      app.kubernetes.io/name: panopticon
```

You can create a new ServiceMonitor by copying the above example and add `metricsRelabelings` config in api endpoint to filter metrics by namespaces.

```yaml
metricRelabelings:
- action: keep
    regex: (demo|kubeops)
    sourceLabels:
    - namespace
```

The above `metricRelabelings` configuration will keep Panopticon metrics from the demo and kubeops namespace and drop metrics from other namespaces. You can modify or add such `metricsRelabelings` according to your need. Please visit [here](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs) to know more about metrics relabel configs.

The Service Monitor object yaml after adding the above `metricRelabelings` configuration looks like the below:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    release: prometheus
  name: panopticon
  namespace: kubeops
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 10s
    port: api
    metricRelabelings:  # added metrics relabeling configuration
    - action: keep
      regex: (demo|kubeops)
      sourceLabels:
      - namespace
    relabelings:
    - action: labeldrop
      regex: (pod|service|endpoint|namespace)
    scheme: https
    tlsConfig:
      ca:
        secret:
          key: tls.crt
          name: panopticon-apiserver-cert
      serverName: panopticon.kubeops.svc
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    interval: 10s
    port: telemetry
    scheme: http
  namespaceSelector:
    matchNames:
    - kubeops
  selector:
    matchLabels:
      app.kubernetes.io/instance: panopticon
      app.kubernetes.io/name: panopticon
```

Pros: 
- Panopticon collects metrics from the target resources from all the namespaces. Users can isolate metrics by namespaces at the Prometheus level.
- If users run multiple Prometheus for a single Panopticon instance and want to use different service monitor objects for each Prometheus then it will become possible to isolate Panopticon metrics over namespaces to different Prometheus.

Cons:
- Metrics for target resources from all the namespaces are available to Panopticon. It might be possible for Prometheus to scrape all the metrics for misconfiguration.
- If users want to collect metrics from a newly created namespace or don't want to collect metrics from a namespace anymore, users need to update the existing Service Monitor relabeling config every time.

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
