---
title: Manage HashiCorp Vault in Kubernetes Native Way Using KubeVault - Webinar
date: 2021-08-12
weight: 25
authors:
  - Shohag Rana
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - mariadb
  - memcached
  - mongodb
  - mysql
  - postgresql
  - redis
  - kubedb
---
## Summary

AppsCode held a webinar on **"Manage HashiCorp Vault in Kubernetes Native Way Using KubeVault"**. This took place on 12th August 2021. The contents of what took place at the webinar are shown below:

1) What is a secret?
2) Managing secrets in Kubernetes.
3) Consuming external secrets in Kubernetes
4) Managing Vault in Kubernetes (Kubernetes native way)
5) Operator over Helm charts
6) KubeVault Introduction & Features
7) Demo
    * Deploy VaultServer using KubeVault Operator
    * Enable & Configure Database SecretEngine
    * Mount Dynamically generated credentials in a Pod using CSI Driver
    * High Availability & Disaster Recovery
8) Q & A Session

## Description of the Webinar

The webinar starts with describing how to manage secrets in Kubernetes and the lackings that Kubernetes has in doing so. Then, it is described why operators are preferred over Helm/YAML. After that, the description of what KubeVault is and its features are shown. The features are:

* Auto Initialization & Unsealing of Vault
* Dynamic Phase Reflections
* Accidental Deletion Prevention
* Vault Policy Control
* Multiple Authentication Method (TLS, Userpass, Token etc.)
* Multiple Storage Backends Support
* Multiple Secret Engines Support

After showing the different features of KubeVault the `demo` portion of the webinar started. In the demo, at first, it was shown how to `install KubeVault, Secrets Store CSI Driver and Vault Specific CSI Provider`. An `Elasticsearch database` using `KubeDB` operator by AppsCode was used.

After that, it was shown how `VaultServer` can be deployed using `Raft Storage Backend`. `GCP bucket` was used to store the `Vault root-token & the unseal-keys`. Besides this, Enabling & Configuring SecretEngine using KubeVault was shown. Finally, in the demo, it was shown how to generate Dynamic Elasticsearch credentials & Mounted them in a Pod using Secrets Store CSI drive. During the Demo, different CRD of KubeVault were also discussed.

At the last part of the demo, different scenarios to show the `High Availability & Disaster Recovery` capability of KubeVault were simulated. Finally, the `Q&A session` was held and the webinar was finished. All in all, it was an effective webinar which showed the importance and contribution of KubeVault and how we can use it effectively.

Take a deep dive into the full webinar below:

<iframe style="height: 500px; width: 800px" src="https://www.youtube.com/embed/T8Be6iKonxE" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/latest/setup/).

* Step by step guide & the manifest files used in the demo can be found [here](https://github.com/kubevault/demo/tree/master/webinar-08-12-21).

* If you want to upgrade KubeVault from a previous version, please follow the upgrade instruction from [here](https://kubevault.com/docs/latest/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeVault community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
