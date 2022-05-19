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

For many years, logs have been an essential part of troubleshooting application and infrastucture performance. In Kubernetes, logging mechanism becomes more crucial to manage and monitor services and infrastructure.

In this post, we are going to give you a full setup guide about how you can setup Grafana Loki for collecting logs from KubeDB pods and how you can generate alert based on those logs.

Here is the outline of this post:

* Loki
* Install Loki in Kubernetes
* Promtail
* Install Promtail in Kubernetes
* Explore logs in Grafana
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
helm upgrade -i <release-name> grafana/promtail -n loki \
    --set config.lokiAddress=http://<loki-distributor-service-name>.<namespace>.svc:3100/loki/api/v1/push
```

Example:
```bash
~ $ kubectl get svc -n loki
NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
loki-loki-distributed-distributor         ClusterIP   10.96.247.189   <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-gateway             ClusterIP   10.96.146.44    <none>        80/TCP                       44m
loki-loki-distributed-ingester            ClusterIP   10.96.74.194    <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-ingester-headless   ClusterIP   None            <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-memberlist          ClusterIP   None            <none>        7946/TCP                     44m
loki-loki-distributed-querier             ClusterIP   10.96.165.151   <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-querier-headless    ClusterIP   None            <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-query-frontend      ClusterIP   None            <none>        3100/TCP,9095/TCP,9096/TCP   44m
```
In this case, `loki-loki-distributed-distributor` is the required service to write the logs.

Note: In loki mircoservice mode, distributor component takes log write request and sends it to available ingestor components. Then ingestor components actually write the log data in the configured storage. To learn more about loki components, please visit [here](https://grafana.com/docs/loki/latest/fundamentals/architecture/components/).


## Explore logs in Grafana

To explore the logs in Grafana, from Datasource section we have to add loki datasource like below:

![loki-datasource](./static/loki-add-ds.png)

Here, we have to add loki query component service address in url section.

Example:
```bash
~ $ kubectl get svc -n loki
NAME                                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
loki-loki-distributed-distributor         ClusterIP   10.96.247.189   <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-gateway             ClusterIP   10.96.146.44    <none>        80/TCP                       44m
loki-loki-distributed-ingester            ClusterIP   10.96.74.194    <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-ingester-headless   ClusterIP   None            <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-memberlist          ClusterIP   None            <none>        7946/TCP                     44m
loki-loki-distributed-querier             ClusterIP   10.96.165.151   <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-querier-headless    ClusterIP   None            <none>        3100/TCP,9095/TCP            44m
loki-loki-distributed-query-frontend      ClusterIP   None            <none>        3100/TCP,9095/TCP,9096/TCP   44m
```
In this case, `loki-loki-distributed-querier` is the required service to query the logs.


Now from Grafana `Explore` section, logs can be explored like below:

![loki-log-explore-sample](./static/sample-loki-logs.png)
