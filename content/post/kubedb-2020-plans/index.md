---
title: KubeDB 2020 H1 Plans
date: 2020-01-27
weight: 13
authors:
  - Tamal Saha
tags:
  - kubernetes
  - kubedb
  - database
---

<h2></h2>


<h2>Day 2 Features</h2>


<h3>Upgrade and Scaling **<code>[ENTERPRISE] [AFN]</code></strong></h3>


KubeDB will provide operator managed human-in-the-loop upgrade, downgrade and scaling features via its new DBA crds called [XYZModificationRequest](https://github.com/kubedb/apimachinery/tree/master/apis/dba/v1alpha1). The CRD and documentation will be open source but the operation will be closed source and only available to paid users. This feature will be added to the following order to supported databases:



*   MongoDB
*   MySQL
*   Elasticsearch
*   PostgreSQL
*   Redis
*   PerconaXtraDB
*   PgBouncer
*   ProxySQL

<h3>Rework Postgres Operator</h3>


We need to re-wrok Postgres operator to address [https://github.com/kubedb/issues/issues/385](https://github.com/kubedb/issues/issues/385) . This is a prerequisite for adding upgrade support for Postgres.

<h3>Multi-zone support**<code> [AFN]</code></strong></h3>


KubeDB currently supports affinity and anti-affinity attributes for database crds. But this does not work properly for sharded databases. This will be reworked to simplify how zones and regions are provided.

```


```
spec:
  preferredZone: ["us-east-1a", "us-east-1b", "us-east-1c"]
```


```

<h3>Remove Snapshot CRD</h3>


In 2019, we integrated KubeDB with [Stash](https://stash.run) to use as a backup and recovery tool for Databases. Snapshot CRD in KubeDB is now considered deprecated.

ToDos:



*   Delete Snapshot CRD
*   Delete any Snapshot CRD related code from KubeDB operators
*   If TerminationPolicy: WipeOut, set KubeDB CRD as the owner for 
*   BackupConfiguration
*   Repository

<h3>Remove DormantDatabase CRD</h3>


In 2017, when we started the KubeDB project, there was no way to prevent accidental deletion of a database CRD and the associated. So, we invented the DormantDatabase CRD. DormantDatabase can be confusing. DormantDatabase CRD is a union of all the database CRDs. This is starting to cross the limit of a CRD definition with the introduction of OpenAPIV3 schema. Now that ValidatingWebhook are a standard feature of Kubernetes 1.9+, we are going to remove the DormantDatabase and merge its functionality into the database specific CRD. \


ToDos:



*   Introduce `spec.halt` in KubeDB database CRDs as a replacement for the `TerminationPolicy: Pause` feature. If this field is set to true, all Kubernetes resources for a running database except PVC and the database CRD itself will be deleted.
*   Delete DormantDatabase CRD and enable description fields in 
*   Delete DormantDatabase CRD and  related code from KubeDB operators

<h2>Monitoring Databases</h2>


<h3>Changes in Prometheus operator Integration**<code> [AFN]</code></strong></h3>


KubeDB operator currently integrates with both Prometheus operator and Prometheus built-in annotations. But [Prometheus upstream no longer recommends using annotations](https://github.com/kubernetes-sigs/kubebuilder/pull/1030). On the other hand, users are asking for [additional config options to configure ServiceMonitor](https://github.com/kmodules/monitoring-agent-api/issues/25). We don't want KubeDB crds to become an uber CRD encompassing all the offshot K8s resources.

ToDos:



*   KubeDB (and AppsCode projects in general) will just add Prometheus sidecar and create an in-cluster “stats” service to expose the Prometheus metrics.
*   Users will be in charge of creating ServiceMonitors with their preferred configuration.
*   We have [restructured the YAML spec](https://github.com/kmodules/monitoring-agent-api/pull/27/files#diff-0a46dcdef0a1eaa34a80c2e2be20b485R110-R145) to clarify what part of the configuration applies to exporter sidecar. This will require code changes and documentation updates in KubeDb operators.

<h3>Grafana Dashboard**<code> [AFN]</code></strong></h3>


We are looking to make it possible to create a Grafana Dashboard using CRD. Currently there is no Grafana operator and CRD. We are looking to invent one [https://github.com/searchlight/grafana-operator](https://github.com/searchlight/grafana-operator) . 

ToDos:



*   Design and Implement Grafana operator that supports Dashboard and DashboardTemplates CRDs.
*   Introduce official KubeDB dashboards for supported databases.

<h2>UX</h2>


<h3>Chart / Kustomize Bundle</h3>


Users have been asking for a chart per database type so that it is easy to deploy various configurations with one command. Relevant issues:



*   [https://github.com/kubedb/issues/issues/439](https://github.com/kubedb/issues/issues/439)
*   [https://github.com/kubedb/issues/issues/679](https://github.com/kubedb/issues/issues/679)
*   [https://github.com/stashed/stash/issues/988](https://github.com/stashed/stash/issues/988)
*   [https://github.com/kmodules/monitoring-agent-api/issues/25](https://github.com/kmodules/monitoring-agent-api/issues/25)

It is not clear whether we should create a Chart or Kustomize bundle. Kustomize bundle will make it easy to customize but Charts are also very popular and more used at this time. This bundle will potentially include the following CRDs:


<table>
  <tr>
   <td><strong>Functionality</strong>
   </td>
   <td><strong>Name of CRD</strong>
   </td>
  </tr>
  <tr>
   <td>Database CRD
   </td>
   <td><code>Postgres / Elasticsearch</code>
   </td>
  </tr>
  <tr>
   <td>Backup using Stash
   </td>
   <td><code>BackupConfiguration</code>
<p>
<code>Repository</code>
   </td>
  </tr>
  <tr>
   <td>Monitoring via Prometheus 
<p>
+
<p>
Dashboards using Grafana
   </td>
   <td><code>ServiceMonitor (prometheus-operator)</code>
<p>
<code>PrometheusRule (prometheus-operator)</code>
<p>
<code>DashboardTemplate (grafana-operator)</code>
<p>
<code>Dashboard (grafana-operator)</code>
   </td>
  </tr>
  <tr>
   <td>User Management using KubeVault
   </td>
   <td><code>VaultServer</code>
<p>
<code>PostgresRole</code>
<p>
<code>MongoDBRole</code>
<p>
<code>MySQRole</code>
   </td>
  </tr>
  <tr>
   <td>Certificate Management using jetstack/cert-manager
   </td>
   <td><code>Issuer</code>
<p>
<code>ClusterIssuer</code>
   </td>
  </tr>
  <tr>
   <td>Security
   </td>
   <td><code>NetworkPolicy</code>
   </td>
  </tr>
  <tr>
   <td>Ingress
   </td>
   <td><code>Ingress</code>
   </td>
  </tr>
  <tr>
   <td>Policy Management using 
<p>
<a href="https://github.com/open-policy-agent/gatekeeper">Open Policy Agent</a>
<p>
(needs further research)
   </td>
   <td>
   </td>
  </tr>
</table>


<h3>Web Dashboard **<code>[ENTERPRISE]</code></strong></h3>


We are working on a web based management console for KubeDB and Kubernetes in general. It is going to be offered a SaaS service on [https://byte.builders/](https://byte.builders/) and an enterprise version that users can run on their own servers.

<h3>Kubedb cli</h3>


In 2017, kubectl did not handle CRDs well. So, we started with our own [cli](https://github.com/kubedb/cli). Since then kubectl has improved its support for CRD and introduced a plugin mechanism.

ToDos:



*   Convert kubedb/cli into a kubectl plugin and update the docs accordingly.
*   [Provide commands to easily exec into a running database.](https://github.com/kubedb/cli/pull/447)
*   Provide commands to easily create logical databases.
*   Provide commands to deploy pre-configured databases using Charts/Kustomize bundle
*   Provide commands to approve/reject Upgrade requests.
*   Publish kubedb plugin to [krew](https://github.com/kubernetes-sigs/krew).

<h2>Security</h2>


<h3>Cert-manager Integration **<code>[ENTERPRISE] [AFN]</code></strong></h3>


KubeDB operator currently provisions certificates for MongoDB and Elasticsearch. It does not support certificate management for other database types. Even for MongoDB and Elasticsearch, it leaves renewing certificates, issues client certificates etc up to the user. We are looking to use Jetstack’s cert-manager project to support these features.

ToDos:



*   Add [`spec.tls`](https://github.com/kubedb/apimachinery/blob/17bd64fb2de2c27220bed0dceeaa822dc45f7ae7/apis/kubedb/v1alpha1/types.go#L135-L160) to all the database crds.
*   Support providing certificated directly using Kubernetes secret
*   Add docs explaining how to use the feature and issue client certificates.
*   Update Prometheus Exporters to support tls enabled database servers.

<h3>User Management via KubeVault</h3>


KubeDB supports user management using Hashicorp Vault via KubeVault project.

ToDos:



*   Implement Elasticsearch credential management using Vault
*   Database root password rotation using Vault
*   Add documentation of kubedb.com on how to use KubeVault. Currently the documentation lives on kubevault.com website.

<h2>Performance & Benchmarking</h2>


<h3>Publish Benchmark Numbers</h3>


Test benchmark tests against large databases and publish white papers.

<h2>Additional Features</h2>


<h3>In-memory MongoDB Storage**<code> [AFN]</code></strong></h3>


Support in-memory MongoDB storage.

There is a list of issues open on kubedb/issues repo with the label `feature/rds-parity`. Those are under consideration: [https://github.com/kubedb/issues/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3Afeature%2Frds-parity](https://github.com/kubedb/issues/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3Afeature%2Frds-parity)

<h3>Feature Priority Ranking</h3>


1. Upgrade (from v? -> v?) and Scaling

2. Changes in Prometheus operator Integration

3. Grafana Dashboard

4. Cert-manager Integration

5. Multi-zone Support


<!-- Docs to Markdown version 1.0β17 -->


<a style="background-color:black;color:white;text-decoration:none;padding:4px 6px;font-family:-apple-system, BlinkMacSystemFont, &quot;San Francisco&quot;, &quot;Helvetica Neue&quot;, Helvetica, Ubuntu, Roboto, Noto, &quot;Segoe UI&quot;, Arial, sans-serif;font-size:12px;font-weight:bold;line-height:1.2;display:inline-block;border-radius:3px" href="https://unsplash.com/@tavi004?utm_medium=referral&amp;utm_campaign=photographer-credit&amp;utm_content=creditBadge" target="_blank" rel="noopener noreferrer" title="Download free do whatever you want high-resolution photos from Octavian Rosca"><span style="display:inline-block;padding:2px 3px"><svg xmlns="http://www.w3.org/2000/svg" style="height:12px;width:auto;position:relative;vertical-align:middle;top:-2px;fill:white" viewBox="0 0 32 32"><title>unsplash-logo</title><path d="M10 9V0h12v9H10zm12 5h10v18H0V14h10v9h12v-9z"></path></svg></span><span style="display:inline-block;padding:2px 3px">Octavian Rosca</span></a>