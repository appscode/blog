---
title: Use Loki with KubeDB
date: 2022-05-23
weight: 40
authors:
  - Pulak Kanti Bhowmick
tags:
  - log
  - monitoring
  - alert
  - loki
  - alertmanager
  - promtail
  - kubedb
  - grafana
  - kubernetes
---

For many years, logs have been an essential part of troubleshooting application and infrastucture performance. For Kubernetes, logging mechanism becomes more crucial to manage and monitor services and infrastructure.

In this post, we are going to give you a full setup guide about how you can setup Grafana Loki for collecting logs from KubeDB pods and how you can generate alert based on those logs.

Here is the outline of this post:

* Loki
* Install Loki in Kubernetes
* Promtail
* Install Promtail in Kubernetes
* Setup Loki with Alertmanager
* Setup Loki with GCS or S3 bucket

## Loki

Loki by Grafana Labs is a log aggregation system inspired by Prometheus. It is designed to store and query logs from all your applications and infrastructure.

## Install Loki in Kubernetes

We install Loki in Kubernetes using official loki helm chart from [here](https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed). Here, we install loki-distributed helm chart in loki namespace which will run Grafana Loki in mircoservice mode.


```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade -i <release-name> grafana/loki-distributed -n loki --create-namespace
```

Note: You can run Loki as a single binary or as a simple scalable mode. But for Kubernetes, it is recommended to run Loki in mircoservice mode.

## Promtail

Promtail is an agent which ships the contents of local logs to a Loki instance. Promtail has Kubernetes service discovery out of the box. Kubernetes service discovery fetches required labels from the Kubernetes API server. To learn more about promtail, you can visit [here](https://grafana.com/docs/loki/latest/clients/promtail/).

Note: Loki supports a good number of official clients like Promtail for sending logs. You can learn more about them from [here](https://grafana.com/docs/loki/latest/clients/).

## Install Promtail in Kubernetes

We install Promtail in Kubernetes using official helm chart from [here](https://github.com/grafana/helm-charts/tree/main/charts/promtail). Promtail is deployed as a Kubernetes DaemonSet to every node for collecting the logs from that node.

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade -i <release-name> grafana/promtail -n loki 
    --set config.lokiAddress=http://<loki-distributor-service-name>.<namespace>.svc:3100/loki/api/v1/push
```

Note: Here we provide loki distributor service name as we are running loki in microservice mode. In loki mircoservice mode, distributor component takes log write request and sends it to available ingestor components. Then ingestor components actually write the log data in the configured storage. To learn more about loki components, please visit [here](https://grafana.com/docs/loki/latest/fundamentals/architecture/components/).


