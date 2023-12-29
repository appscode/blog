---
title: PostgreSQL Backup with Wal-G and point in time Recover with KubeDB managed Postgres Database
date: "2022-09-07"
weight: 26
authors:
- Md Habibur Rahaman Emon
tags:
- backup
- kubedb
- kubernetes
- pitr
- postgresql
- wal-g
---

## Summary
On Septempber 07, 2022, AppsCode held a webinar on **PostgreSQL Backup with Wal-G and point in time Recover with KubeDB managed Postgres Database**. The key contents of the webinars are:
1) PostgreSQL Backup and Point in time Recovery
2) Features of Backup and Recover with KubeDB
3) Demo
4) Q & A Session

## Description of the Webinar
At first, we talked about `PostgreSQL Backup and Point in time Recovery` in general. To implement it in `KubeDB` we introduced `Wal-g` & `volumesnapshotter`, which will be controlled by `Kubedb and Stash operator`. 

We described the features such as:
 * continuous Backup 
 * Point in time Recovery
 * Volume snapshot wit CSI snapshotter

Then we started the demonstration, at first we create a postgres cluster `demo-pg` in the `demo` namespace where we have mentioned the `archiver` spec for the `backup` configuration. Then we create a database `test`, also created a table `employee` inside it.
Then we have simulated a disaster scenario where we have dropped the table from database.
After that we created another postgres server `restore-pg` in the demo namespace where we have configured `ArchiveRecovery` spec to restore at a point in time before the disaster happened.

Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/gR5UdN6Y99c" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


You can find the yamls used in the webinar [here](https://github.com/kubedb/project/tree/master/demo/postgresql/webinar-2022.09.07)

Please try the latest release and give us your valuable feedback.

- If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/welcome/).


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2022.08.08/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
