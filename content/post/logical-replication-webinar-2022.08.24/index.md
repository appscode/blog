---
title: "PostgreSQL Logical Replication using KubeDB in Kubernetes"
date: 2022-08-29
weight: 26
authors:
- Rakibul Hossain
tags:
- kubedb
- postgres
- postgresql
- logical replication
- kubernetes
- publisher
- subscriber
---

## Summary
On August 24, 2022, AppsCode held a webinar on **PostgreSQL Logical Replication using KubeDB in Kubernetes**. The key contents of the webinars are:
1) PostgreSQL Logical Replication in KubeDB
2) Features
3) Demo
4) Q & A Session

## Description of the Webinar
At first, we gave an overview of `PostgreSQL Logical Replication`.  To implement it in `KubeDB` we introduced two `CRD` `Publisher` & `Subscriber`, which will be controlled by `Logical Replication Operator`. `Publisher` will create and maintain `Publications` inside `KubeDB PostgreSQL Server`, and `Subscriber` will create and maintain `Subscription` inside `KubeDB PostgreSQL Server`.

We described the features such as:
 * Different Postgres Versions
 * TableCreationPolicy
 * Pre Conflict detection
 * Double Opt-in

Then we started the demonstration, at first we create a postgres server `publisher-db` in the `demo` namespace and a database `pub` with a table `table_1` inside the database. Then create a publisher `publisher-sample` in the same namespace, which will create and maintain a publication `my_pub` for table `table_1`. 

After that we created another postgres server `subscriber-db` in the test namespace and a database `sub`. Then we create a subscriber `subscriber-one` in the same namespace, which will create and maintain a subscription `my_sub`  for the publication `my_pub` managed by the Publisher CRD. It also create the table `table_1` as we specified the `tableCreationPolicy: IfNotPresent` in the subscriber's `Spec`.
Then we inserted some data into the table `table_1` inside the publisher database `pub` and checked if it replicates into the table inside the subscriber database `sub`.

Later we create a table `table_2` inside the publisher database `pub` and create a publication `your_pub` for that table. After that we create another subscriber `subscriber-two`, which create and maintain a subscription `my_sub` for publication `you_pub`.
Then we inserted some data into the table `table_2` inside the publisher database `pub` and checked if it replicates into the table inside the subscriber database `sub`.

Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/7ruM4yRfnw0" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


You can find the yamls used in the webinar [here](https://github.com/kubedb/project/tree/master/demo/postgres-logical-replication)

Please try the latest release and give us your valuable feedback.

- If you want to install **KubeDB**, please follow the installation instruction from [here](https://kubedb.com/docs/v2022.08.08/welcome/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/kubedb).

If you have found a bug with KubeDB or want to request new features, please [file an issue](https://github.com/kubedb/project/issues/new).