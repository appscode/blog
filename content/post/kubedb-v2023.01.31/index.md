---
title: Announcing KubeDB v2023.01.31
date: "2023-01-31"
weight: 25
authors:
- Mehedi Hasan
tags:
- cloud-native
- dashboard
- database
- elasticsearch
- kafka
- kibana
- kubedb
- kubernetes
- mariadb
- mongodb
- mysql
- opensearch
- percona
- percona-xtradb
- postgresql
- redis
---

We are pleased to announce the release of [KubeDB v2023.01.31](https://kubedb.com/docs/v2023.01.31/setup/). This post lists all the major changes done in this release since the last release.

The release was mainly focus on some hot fix regarding `Private Registry ImagePull`, Postgres `pg-coordinator leader switch` issue and running Kibana and OpenSearch dashboard as `non root` user.

You can find the detailed changelogs [here](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.01.31/README.md).

## Private Registry ImagePull

We have faced issues related to pulling image from `insecure registry`.Now you can specify the image-pull secret for the private registry and register the insecure registry while installing / upgrading kubeDB.
Operator will use that plus any image pull secret in the `db.spec.podTemplate` to get the docker image digest.

```bash
kubectl create secret docker-registry regcred \
  --namespace kubedb \
  --docker-server=<your-registry-server> \
  --docker-username=<your-name> \
  --docker-password=<your-pword>
```

```bash
  helm install kubedb appscode/kubedb \
  --version v2023.01.31 \
  --namespace kubedb --create-namespace \
  --set kubedb-provisioner.enabled=true \
  --set kubedb-ops-manager.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-dashboard.enabled=true \
  --set kubedb-schema-manager.enabled=true \
  --set-file global.license=/path/to/the/license.txt \
  --set global.imagePullSecrets[0].name=regcred \
  --set global.insecureRegistries[0]=harbor.example.com
```

## Pg-coordinator Leader Switch

Pg-coordinator leader switch issue is addressed in this release.From this release KubeDB managed High Availability PostgreSQL will be stable more than before. Try out [kubedb manged PostgreSQL](https://kubedb.com/docs/v2023.01.31/guides/proxysql/concepts/proxysql/)

## Kibana and Opensearch Dashboard As non-root user

Kibana and Opensearch-Dashboards init container docker images will be using **Non-root** user. Running the Docker containers as a non-root user eliminates potential vulnerabilities in the daemon and the container runtime.
This is because if a user manages to break out of the application running as root in the container, he may gain root user access on the host. In addition, configuring a container to user unprivileged is the best way to prevent privilege escalation attacks. You can try out [Deploy Kibana With ElasticsearchDashboard
](https://kubedb.com/docs/v2023.01.31/guides/elasticsearch/elasticsearch-dashboard/kibana/)

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2023.01.31/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2023.01.31/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
