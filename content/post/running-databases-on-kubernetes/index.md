---
title: Running Production-Grade Databases on Kubernetes - Challenges and Solutions
date: "2022-05-26"
weight: 14
authors:
- Tamal Saha
- Dipta Roy
tags:
- cloud-native
- database
- dbaas
- kubedb
- kubernetes
- kubernetes-database
- kubernetes-native
- stash
---

As containers are taking over the world of software development, Kubernetes has emerged as the platform that lets developers seamlessly deploy, scale, run applications, and manage their life cycles. Kubernetes is a DevOps game-changer since it allows teams to focus on applications and deployment rather than worrying about the underlying infrastructure. Given the multi-cloud environment in which DevOps teams perform, Kubernetes abstracts the cloud provider and enables enterprises to build cloud-native applications that can run anywhere.

But managing stateful workloads in a containerized environment has always been a challenge. However, as Kubernetes developed, the whole community worked hard to bring stateful workloads to meet the needs of their enterprise users. As a result, Kubernetes introduced StatefulSets which supports stateful workloads since Kubernetes version 1.9. Users of Kubernetes now can use stateful applications like databases, AI workloads, and big data. Kubernetes support for stateful workloads comes in the form of StatefulSets. 

Despite all of these, the data layer is still demanding because it requires a vast knowledge to run data applications or stateful applications on Kubernetes. And as of today, if you want to run your stateful applications on Kubernetes and you need a database, you have three solutions or options to implement it, one is to go with the managed database service of a cloud provider like AWS, Azure, GCP. Or, you can go with the hosted database service that is usually provided by the database vendors themselves like Elastic Cloud, MongoDB Atlas, etc. And the third option is you run the Kubernetes cluster and run your databases inside Kubernetes. 

If you go to the cloud provider solution, the primary drawback is that cloud providers aren't really your DBA. If you have a slow query, they are not going to tell you how to fix that and you have to figure that out on your own. And your operating procedures will become vendor locked. There is limited option for customization. You can’t just use that database extension you need. If you have data sovereignty or air-gapped requirements, you are out of luck. The usage based pricing model of cloud services can be very expensive at scale. And the same kind of issue goes with the Database Vendor Solutions too. Here, you will face an additional issue which is the choice of more than one database engine. Database vendors only offer one database which they are working on.

So, you are left with the option of running databases inside Kubernetes. Kubernetes is much more stable and mature today than it was at its inception in 2014. But using databases in Kubernetes native way requires some preparation. You should have a good knowledge of Kubernetes, Helm Chart, YAML,  some database engine and then provision those things according to your needs. Even after initial provisioning, you will face complications with routine database tasks such as Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup and Recovery, Failure detection, and Security. There will be a lot of things you have to do manually to keep your database for production workloads in Kubernetes.

This is where a Kubernetes operator based approach can really shine. Kubernetes operators are applications that run inside a Kubernetes cluster and extend Kubernetes for a specific application. Like a human operator or DBA, a Kubernetes operator for a database can simplify and automate routine database tasks such as Provisioning, Monitoring, Upgrading, Patching, Scaling, Volume Expansion, Backup and Recovery, Failure detection, and Security configuration.

Unlike human DBAs, you can shift left database management to app developers. Developers can adopt cloud native DevOps practices uniformly across any number of Kubernetes clusters using GitOps. You are no longer limited to the cloud. You can bring those same practices to on-premises where the majority of workloads run today. You don’t need to worry about all of the knowledge of Kubernetes concept, Helm, and database configuration. A Kubernetes operator simplifies the whole process for you using Custom Resources. You just have to use standard Kubernetes CLI and API to provision your databases. You can upgrade your database to any major or minor update of the database version. You can use Prometheus metrics to scale your database both vertically and horizontally with the rest of your application. You can expand your storage capacity of the database in Kubernetes. A Kubernetes operator can secure your database using TLS and automate certificate rotation. Kubernetes operators can also protect your database from disaster by periodically taking backups to object stores like S3 etc. In short, a well implemented Kubernetes operator based database management solution can offer you the best of both worlds - you get the benefits of a cloud provider managed solution without any of its limitations.

[KubeDB](https://kubedb.com/) by [AppsCode](https://appscode.com/) is an open-core production-grade cloud-native database management solution for Kubernetes. KubeDB simplifies and automates routine database tasks such as provisioning, patching, backup, recovery, scaling & tls management for various popular databases on private and public clouds. It frees you to focus on your applications so you can give them the fast performance, high availability, security and compatibility they need. KubeDB offers many familiar database engines to choose from, including PostgreSQL, MySQL, MongoDB, Elasticsearch, Redis and Memcached. And the list is growing. KubeDB’s native integration with Kubernetes makes it a unique solution compared to competitive solutions from cloud providers and database vendors. KubeDB has been under development since 2017 and is powering critical infrastructure for many enterprises in Telecom, Healthcare, Gaming, EPR, and Financial industries.


Visit: 
[KubeDB.com](https://kubedb.com/)
[Stash.run](https://stash.run/)
[AppsCode.com](https://appscode.com/)


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

More about [Managing Databases in Kubernetes](https://kubedb.com/) & also [Backup and Restore in Kubernetes](https://stash.run/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
