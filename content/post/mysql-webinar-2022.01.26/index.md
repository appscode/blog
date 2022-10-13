---
title: MySQL Read Replica and Multi tenancy Support by KubeDB
date: 2022-01-27
weight: 20
authors:
- Mehedi Hasan
- Tasdidur Rahman
tags:
- kubedb
- mysql
- read replica
- re-configure tls
- multi tanancy
- schema manager
- kubevault
- cert-manager
- stash
- kubernetes
- secret-management
- security
- vault
- hashicorp
- enterprise
- community
---

## Summary

On 26th Jan 2020 Appscode held a webinar on **"MySQL Read Replica and Multi tanancy Support by KubeDB"**. Key contents of the webinars are:

- MySQL Read Replica
- Creating and Managing Read Replicas using kubeDB
- Multi Tenancy
- Creating and managing Multi Tenant Database Schemas using KubeDB Schema Manager.


## Description of the Webinar

Earlier in the webinar they discuss mysql read replica and  configure a KubeDB managed `MySQL` Instance to Allow Read Replica. Moving with that they create a MySQL `Read Replica` Using KubeDB. Then they Secure the Source and Replication Using kubeDB OpsRequest `ReConfigureTLS`.

Later in the Webinar they discuss about multi tenancy.Where they create several databases schemas using kubeDB's newly introduced features Schema Manager and showed creating , altering , initializing Database using Schema.They discussed about user management and security concerns regarding multi tenancy. 

Lastly they talked about their Upcoming features `Semi-Synchronous` Replication support for `KubeDB`



  Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/egzPGc6Yk_A" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.12.21/welcome/).

* If you want to install **KubeVault**, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.01.11/setup/).
 
* If you want to install **Stash**, please follow the installation instruction from [here](https://stash.run/docs/v2021.11.24/setup/).



## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
