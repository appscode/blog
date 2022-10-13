---
title: KubeDB Health Checker
date: 2022-08-19
weight: 25
authors:
- Md Fahim Abrar
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
- percona
- percona-xtradb
---

## KubeDB Health Checker

In [KubeDB v2022.08.08](https://kubedb.com/docs/v2022.08.08/setup/), we've improved KubeDB database health checks and users can now control the behavior of the health checks performed by KubeDB. We've added a new field called `healthChecker` under spec. It controls the behavior of health checks. It has the following fields:

- `spec.healthChecker.periodSeconds` specifies the interval between each health check iteration.
- `spec.healthChecker.timeoutSeconds` specifies the timeout for each health check iteration.
- `spec.healthChecker.failureThreshold` specifies the number of consecutive failures to mark the database as NotReady.
- `spec.healthChecker.disableWriteCheck` specifies if you want to disable database write check, by default KubeDB performs write check.

For example, the `healthChecker` section of a KubeDB database yaml looks like this:
```yaml
spec:
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 3
    disableWriteCheck: true
```

If you provide the above health checker configuration in your KubeDB database yaml, health check will be performed like the following:

- A health check will be performed every `15 seconds` for your database.
- If a health check takes more than `10 seconds`, that health check will be timed out and will be considered as failed.
- If a particular health check, say creating a database client, fails for `3 times` only then the database will be considered as `NotReady`. But if the number of failures is less than 3 then the database won't be considered as `NotReady`.
- Write check won't be performed for the database.

## How does KubeDB Health Checker Work?

Now, let's see how the KubeDB health check is performed by the operator. We can describe the health check in the following way. First, we can see how the phase is calculated using 2 status conditions of the database object. Then we can see how these conditions are calculated.

### How DB Phase is calculated?

The flowchart below shows how KubeDB calculates the database Phase.

<figure align="center">
 <img alt="kubedb phase calculation" src="kubedb-phase-calculation.png">
 <figcaption align="center">Fig: Flowchart of Phase calculation</figcaption>
</figure>

The flowchart shows that KubeDB checks the status conditions of the database objects.

- First, it checks the `ReplicaReady` condition.
- Then, it checks the `AcceptingConnection` condition.
- If both of the conditions are true, KubeDB sets the database phase as `Ready`.
- If `ReplicaReady` is false, but `AcceptingConnection` is true, the database phase is set to `Critical`.
- If `AcceptingConnection` is false, irrespective of the `ReplicaReady` conditions the database is set to `NotReady`.

The below table shows the currently supported KubeDB phases and what each phase means:

| Phases        | Reason                                                                                             |     DB Usable      |
|:--------------|----------------------------------------------------------------------------------------------------|:------------------:|
| Provisioning  | Databases are currently provisioning                                                               |        :x:         |
| DataRestoring | Databases for which data is currently restoring                                                    | :white_check_mark: |
| Ready         | Databases that are currently `ReplicaReady`, `AcceptingConnection`                                 | :white_check_mark: |
| Critical      | Clients can connect to database but one or more replicas are not ready (via watching statefulsets) | :white_check_mark: |
| NotReady      | Databases that can't connect (either connect, ping, write or other checks failed)                  |        :x:         |
| Halted        | Databases that are halted (Pods deleted, PVCs exist)                                               |        :x:         |
| Unknown       | Health checker has been disabled (happens during horizontal scaling)                               |     :question:     |


### How ReplicaReady condition is determined?

The flowchart below shows how KubeDB determines the `ReplicaReady` condition.

<figure align="center">
 <img alt="kubedb determining replicaReady condition" src="kubedb-determining-replica-ready-condition.png">
 <figcaption align="center">Fig: Flowchart of determining ReplicaReady condition</figcaption>
</figure>

We can see from the flowchart that, When a database StatefulSets have any changes, the operator checks if all the Pods of that StatefulSet are in Ready state.

- If all the Pods are in Ready state, the `ReplicaReady` condition is set to True.
- Otherwise, the `ReplicaReady` condition is set to False.

### How AcceptingConnection condition is determined?

The flowchart below shows how KubeDB determines the `AcceptingConnection` condition.

<figure align="center">
 <img alt="kubedb determining acceptingConnection condition" src="kubedb-determining-accepting-connections-condition.png">
 <figcaption align="center">Fig: Flowchart of determining AcceptingConnection condition</figcaption>
</figure>

From the flowchart we can see that, on each health check, KubeDB performs the following checks:

- Can a database client be created?
- Can the database servers be pinged using the created client?
- If `disableHealthCheck` is not true, can the database be written to?
- If the database is in cluster mode with primary and secondary nodes, Is there only one Primary node?

If all the answers are affirmative, KubeDB sets the `AcceptingConnection` condition as `True`.

But, if any of the checks failed, KubeDB checks how many times that particular check failed before. If the number of failures crosses the `failureThreshold` provided by the user, KubeDB sets the `AcceptingConnection` condition as `False`.

So, using the `AccptingConnection` and `ReplicaReady` conditions, KubeDB determines the database phase and the phase reflects the current health of the Database.

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
