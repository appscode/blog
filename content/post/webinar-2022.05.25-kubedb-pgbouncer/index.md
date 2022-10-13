---
title: PostgreSQL Connection Pooling in Kubernetes using Kubedb PgBouncer
date: "2022-05-27"
weight: 20
authors:
- Rakibul Hossain
tags:
- community
- enterprise
- kubedb
- kubernetes
- pgbouncer
- postgres
- statefulsets
---

## Summary

On 25th May 2022, AppsCode held a webinar on **PostgreSQL Connection Pooling in Kubernetes using Kubedb PgBouncer**. The essential contents of this webinar are:

- PgBouncer Clustering.
- Update PgBouncer Configuration.
- TLS configuration for client & server.
- Q & A Session.



## Description of the Webinar

At first, in this webinar, the speaker talked about what database connection pooling is and why it is essential for PostgreSQL. Then we talked about what PgBouncer is and its architecture. Then we discussed KubeDB PgBouncer's new release & its features.

Later in the demo, we showed Pgbouncer Clustering. We deployed `KubeDB PgBouncer Server` with three replicas, referring to the `PostgreSQL` by `KubeDB` and a `Secret` with `userlist` authentication data created earlier. We showed how the `PgBouncer` hits the `Postgres server`.

Then we showed how the update on PgBouncer Configuration works. We added another admin `nazmul` and showed how it gets auto `RELOAD` inside `KubeDB PgBouncer Server`.

At last, we showed the TLS configuration for `KubeDB Pgbouncer`. We deployed `TLS` secured `KubeDB PgBouncer Server` (TLS managed by `cert-manager`), referring to the `PostgreSQL` by `KubeDB` and a `Secret` with `userlist` authentication data and the  `ClusterIssuer` created earlier.


Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/am4tabT2lXU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.05.24/welcome/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).
