---
title: MongoDB and PostgreSQL Multi-Tenancy Support by KubeDB Schema Manager
date: 2022-02-03
weight: 20
authors:
  - Arnob kumar saha
  - Rakibul Hossain
tags:
  - kubeDB
  - MongoDB
  - PostgreSQL
  - Multi tenancy
  - Schema Manager
  - kubevault
  - cert-manager
  - Stash
  - kubernetes
  - secret-management
  - security
  - vault
  - hashicorp
  - shard
  - enterprise
  - community
---

## Summary

On 3rd Feb 2022, Appscode held a webinar on **MongoDB and PostgreSQL Multi-Tenancy Support by KubeDB Schema Manager**. Key contents of the webinars are:
-  Creating Schema manager for MongoDB & PostgreSQL
-  User & credential management using Vault
-  Initialize Databases with Scripts given in Volume sources
-  Restoring the BackUp data using Stash



## Description of the Webinar

Earlier in this webinar, they discussed `Multi-tenancy` support for `MongoDB`. Where they created several database-schemas using kubeDB’s newly introduced features Schema Manager. They showed creation , initialization & restoration for Database using Schema. 
They also discussed user management and security concerns regarding `multi-tenancy`.
Then they showed the behavior of Schema-manager for `Sharded` MongoDB cluster & `standalone` MongoDB instance.

Later in this webinar, they discussed Multi-tenancy support for `PostgreSQL`, Where they created several database-schemas using kubeDB Schema Manager. Here, they discussed about kubeDB `double-optIn` support in schema-manager. Then, They showed creation , initialization & restoration for Database using Schema. 



  Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/_rVS3oe1usA" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.12.21/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.01.11/setup/).
 
* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2021.11.24/setup/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
