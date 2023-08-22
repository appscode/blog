---
title: Announcing KubeDB v2023.08.18
date: "2023-08-19"
weight: 14
authors:
- Raihan Khan
tags:
- cloud-native
- dashboard
- database
- elasticsearch
- opensearch
- kafka
- cruise-control
- kubedb
- kubernetes
- security
- mariadb
- mongodb
- mysql
- percona
- percona-xtradb
- pgbouncer
- postgresql
- prometheus
- proxysql
- redis
- day-2-operations
---

We are pleased to announce the release of KubeDB v2023.08.18. This post lists all the major changes done in this release since the last release. The release includes -

- **Use of the restricted pod security label** ⇒  Pod security policies has been removed in k8s 1.25. In place, It brings `Pod Security Standards` into the picture. We are using the restricted mode (https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted) to all of the namespaces where our operators will be installed.   Thus, we achieve some good security standards, like running as non-root-user, privilege escalation will not happen, some destructive kernel capabilities will be dropped etc.

- **Uniform conditions across database opsRequests** ⇒ We utilize the status.Conditions section of the opsRequest CR for correctly maintaining the steps of an opsRequest like VersionUpdate, HorizontalScaling etc. These conditions have been made uniform in all of our supported databases now.

- **Reduce get/patch api calls** ⇒ #TODO

- **Fix generating VersionUpdate recommendation** ⇒ Previously, our recommender used to generate the same recommendation for updating db version, multiple times. In this release, we encountered this issue. Now, the same recommendation will not be generated multiple times.

- **Confirm the db has been paused before opsRequest reconciliation continues** ⇒ We are now using the  `DatabasePaused` opsRequest condition to ensure that KubeDB provisioner operator is in sync with the OpsManager operator while pausing the database.  To do it, We are setting the Paused condition to `Unknown` for ops-manager, & set this to `True` from Provisioner , confirming that it is in sync with the ops-manager. And the ops-manager operator will continue only if it finds the Paused condition to True.

Find the detailed changelogs [HERE](https://github.com/kubedb/CHANGELOG/blob/master/releases/v2023.08.18/README.md). Let’s see what are the database specific changes coming with this release.

## Kafka:
**Cruise Control for Apache Kafka:** Support for Cruise Control backend with it’s UI to be deployed along with Apache Kafka has been added in this release. Cruise Control support includes custom configuration where reconfiguring Cruise Control properties, Cruise-Control-UI setup, Broker Capacity and BrokerSets info, Cluster configurations properties are permissible. Here’s a sample yaml -

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: kafka-ca-issuer
      kind: Issuer
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
        storageClassName: standard
    controller:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
        storageClassName: standard
  cruiseControl:
    suffix: "cc"
    replicas: 1
    configSecret:
      name: kafka-custom-cc
    podTemplate:
      spec:
        resources:
          limits:
            cpu: 1.5
          requests:
            cpu: 800m
            memory: "1Gi"
  terminationPolicy: WipeOut
  storageType: Durable
  version: 3.5.1
```

**Kafka custom configuration:** With this release support for kafka custom configuration is coming to light. User provided configuration in kubernetes secret will be merged with the default configuration prioritizing the user provided one. Configurations can be provided for both kafka Combined mode and kafka Topology mode for dedicated brokers. Here’s how to configure Kafka with custom configuration secret.

- Create a k8s secret with required configuration file (`server.properties` for combined mode and `broker.properties`/`controller.properties` for dedicated mode).
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: custom-config
  namespace: demo
stringData:
  server.properties: |
    log.dirs=/var/custom-logdir
    metadata.log.dir=/var/custom-logdir/metadata
    min.insync.replicas=2
```
- Add the secret reference in Kafka spec.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  configSecret:
    name: custom-config
  replicas: 3
  version: 3.5.1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: DoNotTerminate
```

**New version:** support for kafka version `3.4.1` which have some major bugs fixes and `3.5.1` which is the latest one at the time of this release have been added.

## Elasticsearch/OpenSearch:

Previously, Pod Disruption Budget (PDB) was created only when Elasticsearch custom resource spec was provided with `.spec.maxUnavailable` (for combined cluster) or `.spec.topology.<node-type>.maxUnavailable` (for dedicated cluster). This release ensures that Elasticsearch/Opensearch clusters can have at most a single pod from that set that can be unavailable after the eviction. PDB with such configuration is created by default unless it’s a standalone cluster.

**Fixes:** `Internal User credential synchronization for Elasticsearch failure when security is disabled` issue got fixed with this release. `Vertical Scaling not scaling the pod resources` and `Horizontal Scaling Failure in ES v8` issue also have been resolved.

## MongoDB:

**New Version:** Support for MongoDB `4.2.24` have been introduced in this release. Apply the following YAML to try out this new version with KubeDB.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: restore
  namespace: demo
spec:
  version: "4.2.24"
  terminationPolicy: WipeOut
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

**Fixes:** Use --bind_ip to fix version `3.4.*` CrashLoopbackOff issue.
Fix for MongoDB HorizontalScale down for shared cluster also have been addressed in this release.

## MySQL:


#TODO


## Redis:

We have added the latest Redis version `7.2.0` in this release. To deploy a Redis Standalone instance with version Redis `7.2.0`, you can apply this yaml:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: 7.2.0
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

## KubeDB Cli:
We have added a new set of commands in KubeDB cli to help you insert, verify and drop random data in the KubeDB managed databases. Please install or update the `krew` plugin to use the new commands.

We have added **insert**, **verify** and **drop** sub commands for each database which can be run with data command.

```bash
kubectl dba data <sub-command> <kind>  -n <ns> <db-name> –rows <data-count>

# Examples : 
# To insert 1000 rows in a Postgres table
kubectl dba data insert postgres -n demo pg-sample - -rows 1000
# To insert 1000 documents in an Elasticsearch/OpenSearch index
kubectl dba data insert elasticsearch -n demo es-sample -r 1000
#To verify if a MongoDB database contains 500 rows
kubectl dba data verify mongodb -n demo mg-shard  - -rows 500
#To drop all the CLI inserted data from Redis database
kubectl dba data delete redis -n demo rd-sample
```


## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeDB, please follow the installation instruction from [KubeDB Setup](https://kubedb.com/docs/v2023.08.18/setup).

- If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [KubeDB Upgrade](https://kubedb.com/docs/v2023.08.18/setup/upgrade/).


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

Learn More about [Production-Grade Databases in Kubernetes](https://kubedb.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
