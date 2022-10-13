---
title: Kubernetes Backup and Restore - A Complete Solution
date: "2022-06-21"
weight: 14
authors:
- Dipta Roy
tags:
- backup and restore volumes and databases in kubernetes
- backup databases in k8s
- backup volumes and databases in kubernetes
- cloud-native platform
- database
- kubernetes
- kubernetes database
- kubernetes native
- kubernetes native data platform
- kubernetes volumes and databases recovery
- recover volumes and databases in kubernetes
- restore volumes and databases in kubernetes
- stash
- stash kubernetes
---

Over the last few years, we've progressed through the initial excitement and adoption stages of Kubernetes. As businesses adopt Kubernetes in production and the number of Kubernetes clusters and applications expands, it becomes necessary to have a data protection solution for their business applications and microservices.
As the deployment of Kubernetes in the production environment creates new organizational and operational challenges, the conventional backup and restore solutions won't be effective here. A single Kubernetes application in production may include hundreds of components, including pods, volumes, ConfigMaps, secrets, and certificates running on multiple machines holding different configurations. In order to ensure a reliable recovery, the Kubernetes backup solution must collect data and application configuration at a detailed level. Also, you need to backup all namespaces, all Kubernetes control plane data stored in etcd, and all Persistent Volumes to provide a complete Kubernetes cluster backup.

Despite all of these, there are few typical solutions available if you want to backup & restore your volumes and databases on Kubernetes. You can go with the cloud provider's backup and restore solution or you have to write custom backup and restore scripts for the different infrastructure levels.
If you choose a cloud service, they will have their own backup and replication systems. However, the platform just manages some part of it, so you'll need to set up data backups and other features yourself. Moreover, the data in the volume is all that exists. The cluster state is not included in this backup process. You can create custom scripts to manage backup solutions at various levels, including components and the state both inside and outside the cluster. However, because the data and state are scattered across multiple layers and platforms, these scripts can quickly become highly complex, and the scripts are often tied up to the underlying platform where the data is stored. The same goes for the restore process. As a result, you may end up with a large number of complex self-managed scripts that are difficult to maintain.

Apart from all of the typical non-efficient solutions above, you can go with the Kubernetes native operator-based approach where you can simplify and automate the whole backup and restore process on Kubernetes.

[Stash](https://stash.run/) is a complete Kubernetes native production-grade disaster recovery solution for backup and restore your Kubernetes workloads, volumes, and databases on various public and private clouds. Stash is a Kubernetes operator that uses restic and Kubernetes CSI Driver VolumeSnapshotter functionality to address these issues. Using Stash, you can backup Kubernetes volumes mounted in workloads, stand-alone volumes and databases. Users may even extend Stash via addons for any custom workload. Stash offers so many features like Logical Backup, Policy based backup, Kubernetes Volume Backup, Database Backup, Declarative API, Deduplication, Data Encryption, Monitoring, Volume Snapshot and so on. Also, Stash supports multiple storages where you can store your backed up data. Stash supports AWS S3, Minio, Rook, Google Cloud Storage, Azure Blob Storage, OpenStack Swift, DigitalOcean Spaces, Blackbaze B2 and REST server as backup storage. You can also use Kubernetes persistent volumes as backend. Stash supports backup of various popular databases like [Elasticsearch](https://stash.run/addons/databases/backup-and-restore-elasticsearch-on-kubernetes/), [MariaDB](https://stash.run/addons/databases/backup-and-restore-mariadb-on-kubernetes/), [Redis](https://stash.run/addons/databases/backup-and-restore-redis-on-kubernetes/), [PostgreSQL](https://stash.run/addons/databases/backup-and-restore-postgres-on-kubernetes/), [MySQL](https://stash.run/addons/databases/backup-and-restore-mysql-on-kubernetes/), [MongoDB](https://stash.run/addons/databases/backup-and-restore-mongodb-on-kubernetes/), [Percona XtraDB](https://stash.run/addons/databases/backup-and-restore-percona-xtradb-on-kubernetes/) and [Etcd](). Also, you can backup [NATS](https://stash.run/addons/message-queue/backup-and-restore-nats-on-kubernetes/) Server or any [Kubernetes resources](https://stash.run/addons/kubernetes/backup-kubernetes-resources/) using Stash. Stash is an open core project used by thousands of engineers around the world.

[Stash](https://stash.run/) offers an Enterprise edition for a production-grade environment and also provides 24/7 support. Stash community edition is FREE to use on any supported Kubernetes engines. You can Backup & Restore your volumes and databases on Kubernetes using Stash. There is no up-front investment required. Stash offers a 30 days license FREE of cost to try the Stash Enterprise edition. You can simply start with Stash by installing it from here: [stash.run](https://stash.run/).
Click here for [documentation](https://stash.run/docs/latest/welcome/).

We have made an in depth video about **A Better Cloud-Native Backup Experience Using Stash**. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/MREdcm9S8Xg" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/). 

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://kubernetes.slack.com/messages/C8NCX6N23/) channel `#stash`.

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Managing Databases in Kubernetes](https://kubedb.com/) & also [Backup and Restore in Kubernetes](https://stash.run/)

