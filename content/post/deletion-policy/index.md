---
title: Protecting KubeDB managed Databases from Accidental Deletion
date: "2025-07-02"
weight: 26
authors:
- Bonusree Datta
tags:
- Kubernetes
- KubeDB
- Druid
- Kafka 
- Microsoft SQL Server
- OpenSearch
- Pgpool
- RabbitMQ
- Solr
- Elasticsearch 
- MariaDB
- MongoDB
- PerconaXtraDB
- PostgreSQL
- Redis
- Zookeeper
- FerretDB
- Memcached 
- MySql
- PgBouncer
- ProxySQL
- SingleStore

---

> New to KubeDB? Please start [here](https://kubedb.com/docs/v2025.5.30/welcome/).

> 💡 **Note:** The deletion policy is the same for all kubedb supported databases. Here, an example for PostgreSQL is included to show how users can use this feature.

# Using Database Deletion Policy

KubeDB supports setting a deletion policy for all supported databases. This guide will help you choose the appropriate deletion policy to manage your database workloads safely, while aligning with your organization’s data retention and disaster recovery requirements.

## Prerequisite

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](https://kubedb.com/docs/v2025.5.30/setup/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.
  ```shell
    $ kubectl create ns demo
    namespace/demo created
  ```
- **Create a PostgreSQL database (or your chosen database)**

  Below is the Postgres object created in this tutorial.

    ```yaml
    apiVersion: kubedb.com/v1
    kind: Postgres
    metadata:
      name: quick-postgres
      namespace: demo
    spec:
      version: "13.13"
      storageType: Durable
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
      deletionPolicy: Delete
    ```
    
  Create above Postgres object with following command

    ```bash
    $ kubectl create -f https://github.com/kubedb/docs/raw/v2025.5.30/docs/examples/postgres/quickstart/quick-postgres-v1.yaml
    postgres.kubedb.com/quick-postgres created
    ```
  Once you’ve created the Postgres object, use the following command to check the resources:
  
  ```shell
  $ kubectl get petsets,pods,svc,pvc,secrets -n demo -l 'app.kubernetes.io/instance=quick-postgres'
  ```
    
 You'll see:

   ```bash
   NAME                   READY   STATUS    RESTARTS   AGE
    pod/quick-postgres-0   1/1     Running   0          8m50s
    
    NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
    service/quick-postgres        ClusterIP   10.96.59.115   <none>        5432/TCP,2379/TCP            8m53s
    service/quick-postgres-pods   ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   8m53s
    
    NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    persistentvolumeclaim/data-quick-postgres-0   Bound    pvc-1f94b385-59ce-4b0f-ac25-c4c8d16aa21f   1Gi        RWO            standard       8m50s
    
    NAME                         TYPE                       DATA   AGE
    secret/quick-postgres-auth   kubernetes.io/basic-auth   2      8m53s
    
   ```
Let's see what KubeDB operator has created for additional RBAC permission


## What is DeletionPolicy
The deletionPolicy (previously known as TerminationPolicy) in KubeDB defines what happens when you delete a PostgreSQL custom resource (CR). It gives you control over whether deletion should be rejected entirely, or which associated resources KubeDB should keep or remove.

KubeDB supports four types of termination policies:
1. DoNotTerminate
2. Halt
3. Delete (Default)
4. WipeOut

➡️  The following table shows what KubeDB does when users delete any KubeDB CRD with different termination policies:

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |



The use cases for each policy are described below.

#### DoNotTerminate

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB uses Kubernetes’ `ValidatingWebhook` feature (available in Kubernetes 1.9.0 and later) to prevent deletion of the PostgreSQL custom resource.

If admission webhooks are enabled in your cluster, any attempt to delete the database will be blocked as long as `spec.deletionPolicy` is set to `DoNotTerminate`.

**How to set it**

First, edit your PostgreSQL resource to set `spec.deletionPolicy` to `DoNotTerminate`:

```shell
$ kubectl edit pg -n demo quick-postgres
```

In the editor, update it to:

```yaml
spec:
  deletionPolicy: DoNotTerminate
```

Or patch it directly with:
```bash
$ kubectl patch -n demo pg/quick-postgres -p '{"spec":{"deletionPolicy":"DoNotTerminate"}}' --type="merge"
```

**Simulate Deletion**

```bash
$ kubectl delete postgres.kubedb.com/quick-postgres -n demo
```

You'll see:
```bash
Error from server (Forbidden): admission webhook "postgreswebhook.validators.kubedb.com" denied the request: postgres "demo/quick-postgres" can't be terminated. To delete, change spec.deletionPolicy
```
Check resources:
You can now see the same resources as at the initial stage.
```bash
$ kubectl get petsets,pods,svc,pvc,secrets -n demo -l 'app.kubernetes.io/instance=quick-postgres'
```
You'll see:
```bash

NAME                   READY   STATUS    RESTARTS   AGE
pod/quick-postgres-0   1/1     Running   0          8m50s

NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/quick-postgres        ClusterIP   10.96.59.115   <none>        5432/TCP,2379/TCP            8m53s
service/quick-postgres-pods   ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   8m53s

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-quick-postgres-0   Bound    pvc-1f94b385-59ce-4b0f-ac25-c4c8d16aa21f   1Gi        RWO            standard       8m50s

NAME                         TYPE                       DATA   AGE
secret/quick-postgres-auth   kubernetes.io/basic-auth   2      8m53s


```
#### Halt

Suppose you want to reuse your PostgreSQL data volumes and credentials to redeploy the database in the future with the same configuration. But right now, you want to delete the database while keeping the data volumes and credentials intact. In this scenario, you should set the PostgreSQL object's `deletionPolicy` to `Halt`.

When the `deletionPolicy` is set to `Halt` and the PostgreSQL object is deleted, the KubeDB operator will remove the Petset and its pods but will keep the PersistentVolumeClaims (PVCs), Secrets, and any database backup data (snapshots) intact.

**How to set it**

First, edit your PostgreSQL resource to set `spec.deletionPolicy` to `Halt`:
```bash
$ kubectl edit pg -n demo quick-postgres
```
In the editor, update it to:
```yaml
spec:
  deletionPolicy: Halt
```

Or patch it directly with:
```bash
$ kubectl patch -n demo pg/quick-postgres -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
```

**Simulate Delete**

Now, if you delete the Postgres object, the KubeDB operator will delete every resource created for this Postgres CR, but leaves the auth secrets, snapshots and PVCs.

```bash
$ kubectl delete postgres.kubedb.com/quick-postgres -n demo
postgres.kubedb.com "quick-postgres" deleted
```

Check resources:
```bash
$ kubectl get petsets,pods,svc,pvc,secrets -n demo -l 'app.kubernetes.io/instance=quick-postgres'
```
You'll see:
```bash
NAME                         TYPE                       DATA   AGE
secret/quick-postgres-auth   kubernetes.io/basic-auth   2      85m

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-quick-postgres-0   Bound    pvc-4328e3f4-74dd-4edb-9f44-486dab86f05b   1Gi        RWO            standard       85m

```

From the above output, you can see that all Postgres resources(PetSet, Service, etc.) are deleted except PVC and Secret. You can recreate your postgres again using these resources.


#### Delete (default)

If you want to delete an existing PostgreSQL database along with its data volumes, but still plan to restore it later using previously taken snapshots and Secrets, you should set the PostgreSQL object's `deletionPolicy` to `Delete`.

With this setting, when the PostgreSQL object is deleted, the KubeDB operator will remove the Petset, its pods, and the associated PersistentVolumeClaims (PVCs). However, it will retain the Secrets and any database backup data (snapshots), allowing you to restore the database later if needed.

> 💡 **Note:** If you don't set `deletionPolicy` explicitly, it will default to `Delete`.

**How to set it**

First, edit your PostgreSQL resource to set `spec.deletionPolicy` to `Delete`:
```bash
$ kubectl edit pg -n demo quick-postgres
```
In the editor, update it to:
```yaml
spec:
  deletionPolicy: Delete
```

Or patch it directly with:
```bash
$ kubectl patch -n demo pg/quick-postgres -p '{"spec":{"deletionPolicy":"Delete"}}' --type="merge"
postgres.kubedb.com/quick-postgres patched
```

**Simulate Deletion**

Now, if you delete the Postgres object, the KubeDB operator will delete every resource created for this Postgres CR, but leaves the auth secrets and snapshots.

```bash
$ kubectl delete postgres.kubedb.com/quick-postgres -n demo
postgres.kubedb.com "quick-postgres" deleted
```

Check resources:
```bash
$ kubectl get petsets,pods,svc,pvc,secrets -n demo -l 'app.kubernetes.io/instance=quick-postgres'
```
You'll see:
```bash

NAME                         TYPE                       DATA   AGE
secret/quick-postgres-auth   kubernetes.io/basic-auth   2      143m

```
From the above output, you can see that all postgres resources(PetSet, Service, PVCs etc.) are deleted except Secret. You can initialize your postgres using snapshots(if previously taken) and secret.

#### WipeOut
You can completely remove the PostgreSQL database and all related resources without leaving any trace by setting the `deletionPolicy` to `WipeOut`.

When `deletionPolicy` is set to `WipeOut`, the KubeDB operator will delete all associated resources for the PostgreSQL database, including the Petset, PersistentVolumeClaims (PVCs), Secrets, and any backup data (snapshots).

**How to set it**

First, edit your PostgreSQL resource to set `spec.deletionPolicy` to `WipeOut`:

```bash
$ kubectl edit pg -n demo quick-postgres
```

In the editor, update it to:
```yaml
spec:
  deletionPolicy: WipeOut
```

Or patch it directly with:
```bash
$ kubectl patch -n demo pg/quick-postgres -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
postgres.kubedb.com/quick-postgres patched

```

**Simulate Deletion**

Now, if you delete the Postgres object, the KubeDB operator will delete every resource created for this Postgres CR, but leaves the auth secrets and snapshots.

```bash
$ kubectl delete postgres.kubedb.com/quick-postgres -n demo
postgres.kubedb.com "quick-postgres" deleted

```

Check resources:
```bash
$ kubectl get petsets,pods,svc,pvc,secrets -n demo -l 'app.kubernetes.io/instance=quick-postgres'
```
You'll see:
```bash
No resources found in demo namespace.
```

From the above output, you can see that all postgres resources are deleted. there is no option to recreate/reinitialize your database if `deletionPolicy` is set to `WipeOut`.


## Support
- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases?pli=1) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
