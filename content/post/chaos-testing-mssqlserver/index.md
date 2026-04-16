---
title: 'Chaos Testing KubeDB SQL Server Availability Group Clusters with Chaos Mesh: Building Resilient Database Systems'
date: "2026-04-16"
weight: 25
authors:
- Neaj Morshad
tags:
- chaos-engineering
- chaos-mesh
- cloud-native
- database
- disaster-recovery
- failover
- high-availability
- kubedb
- kubernetes
- sqlserver
- mssql 
---

> New to KubeDB? Please start [here](https://kubedb.com/docs/v2026.2.26/welcome/).



# Chaos Testing KubeDB Managed SQL Server Availability Group Clusters with Chaos-Mesh

## Setup Cluster
To follow along with this tutorial, you will need:

1. A running Kubernetes cluster.
2. KubeDB [installed](https://kubedb.com/docs/v2026.2.26/setup/install/kubedb/) in your cluster.
3. kubectl command-line tool configured to communicate with your cluster.
4. Chaos-Mesh [installed](https://chaos-mesh.org/docs/production-installation-using-helm/) in your cluster.
    ```shell
    helm upgrade -i chaos-mesh chaos-mesh/chaos-mesh \
     -n chaos-mesh \
    --create-namespace \
    --set dashboard.create=true \
    --set dashboard.securityMode=false \
    --set chaosDaemon.runtime=containerd \
    --set chaosDaemon.socketPath=/run/containerd/containerd.sock \
    --set chaosDaemon.privileged=true
    ```
> Note: Make sure to set correct path to your container runtime socket and runtime in the above command. For ex: `socketPath=/run/containerd/containerd.sock`, or if in **k3s**, set `chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock`.
   
## Verify KubeDB and Chaos-Mesh Installation

```shell
➤ kubectl get pods -n kubedb
NAME                                             READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-0                       1/1     Running   0          24d
kubedb-kubedb-ops-manager-0                      1/1     Running   0          22d
kubedb-kubedb-provisioner-0                      1/1     Running   0          146m
kubedb-kubedb-webhook-server-699bf949df-24w5k    1/1     Running   0          146m
kubedb-operator-shard-manager-77c8df4946-4gwhc   1/1     Running   0          146m
kubedb-petset-869495bb7f-2cln2                   1/1     Running   0          146m
kubedb-sidekick-794cf489b4-t9rgf                 1/1     Running   0          146m
---

```

```shell
➤ kubectl get pods -n chaos-mesh
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7d44db47fb-4cwc9   1/1     Running   0          3d17h
chaos-controller-manager-7d44db47fb-cqvf7   1/1     Running   0          3d15h
chaos-controller-manager-7d44db47fb-x4xnt   1/1     Running   0          3d17h
chaos-daemon-f779s                          1/1     Running   0          3d17h
chaos-dashboard-6855b9d4c-phkht             1/1     Running   0          4d1h
chaos-dns-server-85b8846dc9-ngcwm           1/1     Running   0          4d1h
```

## Introduction to Chaos Engineering

**Chaos Engineering** is a disciplined approach to test distributed systems by deliberately introducing controlled failure scenarios to discover vulnerabilities and weaknesses before they impact users. Rather than waiting for production incidents, chaos engineering proactively identifies how system behaves under adverse conditions—such as pod failures, network outages, resource exhaustion, and data corruption.

This methodology is particularly crucial for database systems, where failures can lead to data loss, service downtime, and compromised data consistency. By testing these scenarios in controlled environments, we gain confidence that our system can recover gracefully and maintain availability.

### What This Blog Covers

In this comprehensive guide, we will:

1. **Deploy a SQL Server Availability Group Cluster** on Kubernetes using KubeDB, configured with replication and automatic failover capabilities
2. **Run 16+ Chaos Engineering Experiments** using Chaos-Mesh to simulate real-world failure scenarios
3. **Observe Cluster Behavior** during failures including pod crashes, network issues, resource exhaustion, and disk I/O errors
4. **Measure Resilience** by tracking data consistency, failover speed, and recovery capabilities
5. **Learn Best Practices** for configuring SQL Server replication and failover strategies to maximize availability

Each experiment progressively tests different aspects of the system—from simple pod failures to complex scenarios involving multiple simultaneous failures. By the end, we'll have a thorough understanding of how our SQL Server cluster behaves under various failure modes and how to configure it for maximum resilience.

You can see the [`Chaos Testing Results Summary`](#chaos-testing-results-summary) for a quick view of what we have done in this blog.


## Create a High-Availability SQL Server Cluster

First, we need to deploy a SQL Server cluster configured for High Availability.
Unlike a Standalone instance, a HA cluster consists of a primary pod
and one or more standby pods that are ready to take over if the leader
fails.

Save the following YAML as yamls/sqlserver-ag-cluster.yaml. This manifest
defines a 3-node SQL Server cluster.





DONE SO FAR. ##############################################



```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: sqlserver-ag-cluster
  namespace: demo
spec:
  clientAuthMode: md5
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: postgres
          resources:
            limits:
              memory: 3Gi
            requests:
              cpu: 2
              memory: 2Gi
  replicas: 3
  replication:
    walKeepSize: 5000
    walLimitPolicy: WALKeepSize
    # forceFailoverAcceptingDataLossAfter: 30s # uncomment this if you want to accept data loss during failover, but want to have minimal downtime. 
  standbyMode: Hot
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Gi
  storageType: Durable
  version: "16.4"
```
> **`Important Notes`**:
> - We have set walLimitPolicy to WALKeepSize and
> walKeepSize to 5000. This means that we will keep 5000 MB
> of WAL files in our cluster. If your write operations are
> very high, you might want to increase this value.
> We suggest you set it to at least 15 - 30% of your storage.
> - If you can tolerate some data loss, but you want your primary to be up and running at any time with minimal downtime, you can set `.spec.replication.forceFailoverAcceptingDataLossAfter: 30s`
> - You can read/write in your database in both **`Ready`** and **`Critical`** state. So it means even if your db is in `Critical` state, your uptime is not compromised. `Critical` means one or more replicas are offline. But `primary` is up and running along with some other replicas probably.
> - All the results/metrics shown in this blog is related to the chaos scenarios. In general, **a failover takes ~5 seconds** and **without any data loss** ensuring high availability and data safety.

Now, create the namespace and apply the manifest:

```shell
# Create the namespace if it doesn't exist
kubectl create ns demo

# Apply the manifest to deploy the cluster
kubectl apply -f setup/sqlserver-ag-cluster.yaml
```

You can monitor the status until all pods are ready:
```shell
watch kubectl get pg,petset,pods -n demo
```
See the database status is ready.

```shell
➤ kubectl get pg,petset,pods -n demo
NAME                             VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    4m45s

NAME                                      AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   4m41s

NAME               READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0   2/2     Running   0          4m41s
pod/sqlserver-ag-cluster-1   2/2     Running   0          2m45s
pod/sqlserver-ag-cluster-2   2/2     Running   0          2m39s
```

Inspect who is primary and who is standby.

```shell
# you can inspect who is primary
# and who is secondary like below

➤ kubectl get pods -n demo --show-labels | grep role
sqlserver-ag-cluster-0   2/2     Running   0          20m   app.kubernetes.io/component=database,app.kubernetes.io/instance=sqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=sqlserver-ag-cluster-6c5954fd77,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=sqlserver-ag-cluster-0
sqlserver-ag-cluster-1   2/2     Running   0          19m   app.kubernetes.io/component=database,app.kubernetes.io/instance=sqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=sqlserver-ag-cluster-6c5954fd77,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=sqlserver-ag-cluster-1
sqlserver-ag-cluster-2   2/2     Running   0          18m   app.kubernetes.io/component=database,app.kubernetes.io/instance=sqlserver-ag-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=postgreses.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=sqlserver-ag-cluster-6c5954fd77,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=sqlserver-ag-cluster-2

```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the standbys.



## Chaos Testing

We will run some chaos experiments to see how our
cluster behaves under failure scenarios like oom kill, network latency, network partition, io latency, io fault etc. We will use a SQL Server client application to simulate high write and read load on the cluster.

### SQL Server High Write/Read Load Client

You can apply these YAMLs to create a client application
that will continuously write and read data from the database.
This will help us see how the cluster behaves under load and
during chaos scenarios. Make sure you change the password of your database in the below Secret YAML.



```yaml
# k8s/01-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pg-load-test-config
  namespace: demo
  labels:
    app: pg-load-test
data:
  # Test Duration (in seconds)
  TEST_RUN_DURATION: "400"
  
  # Concurrency Settings
  CONCURRENT_WRITERS: "20"
  
  # Workload Distribution (must sum to 100)
  READ_PERCENT: "80"
  INSERT_PERCENT: "10"
  UPDATE_PERCENT: "10"
  
  # Batch Sizes
  BATCH_SIZE: "100"
  READ_BATCH_SIZE: "100"
  
  # Database Settings
  TABLE_NAME: "load_test_data"
  
  # Connection Pool Settings
  MAX_OPEN_CONNS: "60"
  MAX_IDLE_CONNS: "10"
  CONN_MAX_LIFETIME: "300"
  
  # Connection Safety
  MIN_FREE_CONNS: "5"
  
  # Reporting
  REPORT_INTERVAL: "20"
---
# k8s/02-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-load-test-secret
  namespace: demo
  labels:
    app: pg-load-test
type: Opaque
data:
  # Base64 encoded database credentials
  # Replace these with your actual base64-encoded values

  # Example: echo -n "your-postgres-host" | base64
  DB_HOST: cGctaGEtY2x1c3Rlci5kZW1vLnN2Yy5jbHVzdGVyLmxvY2Fs

  # Example: echo -n "5432" | base64
  DB_PORT: NTQzMg==

  # Example: echo -n "postgres" | base64
  DB_USER: cG9zdGdyZXM=

  # Example: echo -n "your-password" | base64
  # IMPORTANT: Replace this with your actual password
  DB_PASSWORD: NihrMkohSXVYdChGSSpmSg==

  # Example: echo -n "postgres" | base64
  DB_NAME: cG9zdGdyZXM=

---
# How to encode your credentials:
# echo -n "127.0.0.1" | base64
# echo -n "5678" | base64
# echo -n "postgres" | base64
# echo -n "CIX6TzfTYFn8~pj4" | base64
# echo -n "postgres" | base64
---
# k8s/03-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pg-load-test-job
  namespace: demo
  labels:
    app: pg-load-test
    version: v2
spec:
  completions: 1
  backoffLimit: 0
  ttlSecondsAfterFinished: 86400
  template:
    metadata:
      labels:
        app: pg-load-test
        version: v2
    spec:
      restartPolicy: Never
      containers:
        - name: load-test
          # Replace with your image registry and tag
          image: souravbiswassanto/high-write-load-client:v0.0.0
          imagePullPolicy: Always
          resources:
            requests:
              memory: "2Gi"
              cpu: "1000m"
            limits:
              memory: "2Gi"
              cpu: "2000m"
          envFrom:
            - configMapRef:
                name: pg-load-test-config
            - secretRef:
                name: pg-load-test-secret
          volumeMounts:
            - name: results
              mountPath: /results
      volumes:
        - name: results
          persistentVolumeClaim:
            claimName: pg-load-test-results
---
# k8s/04-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pg-load-test-results
  namespace: demo
  labels:
    app: pg-load-test
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

> Also as a standard, we will use 10% write, 10% update and 80% 
read operations. In 5 minutes of high load,
it should generate around 30GB of data, more than 
30M rows inserted, more than 300M rows read. 

> **Note**: If you do not want to generate this much data, you can reduce the INSERT_PERCENT and BATCH_SIZE values.


Save the above yamls. Then make a script like below:

```shell
➤ cat run-k8s.sh 
#! /usr/bin/bash

kubectl delete -f k8s/03-job.yaml
kubectl delete -f k8s/04-pvc.yaml

kubectl apply -f k8s/01-configmap.yaml
kubectl apply -f k8s/03-job.yaml
kubectl apply -f k8s/04-pvc.yaml
```

Run the script to start the load test.

```shell
chmod +x run-k8s.sh
./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

I have attached a sample output of the load test job below. These metrics will be printed after every `REPORT_INTERVAL` seconds. You can see that we are generating around 38GB of data, more than 4M rows inserted, more than 32M rows read in 7 minutes of high load.

```shell
Test Duration: 7m3s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 408454 (Reads: 326500, Inserts: 40908, Updates: 41046)
  Total Number of Rows Reads: 32650000, Inserts: 4090800, Updates: 41046
  Total Errors: 0
  Total Data Transferred: 38187.80 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 0.00 (Reads: 0.00/s, Inserts: 0.00/s, Updates: 0.00/s)
  Throughput: 0.00 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 12.097ms, P95: 83.291ms, P99: 100.506ms
  Inserts - Avg: 58.51ms, P95: 146.231ms, P99: 218.178ms
  Updates - Avg: 37.444ms, P95: 100.994ms, P99: 192.838ms
-----------------------------------------------------------------
Connection Pool:
  Active: 13, Max: 100, Available: 87
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 4140800
  Records Found in DB: 4140800
I0406 04:26:53.097674       1 load_generator_v2.go:555] Total records in table: 4140800
I0406 04:26:53.097700       1 load_generator_v2.go:556] totalRows in LoadGenerator: 4140800
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

Cleaning up test data...
Cleaning up test table...
Cleanup completed
Test data table deleted successfully

Test completed successfully!
```

> You can see these logs by running `kubectl logs -n demo job/pg-load-test-job` command.

With this load on the cluster, we are ready to run some chaos experiments and see how our cluster behaves under failure scenarios.

###  Chaos#1: Kill the Primary Pod

We will ignore the load test for this experiment.

We are about to kill the primary pod and see how fast the failover happens. We will use Chaos-Mesh to do this. You can also do this manually by running `kubectl delete pod` command, but using Chaos-Mesh will give you more insights about the failover process.

Save this yaml as `tests/01-pod-kill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pg-primary-pod-kill
  namespace: chaos-mesh
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  gracePeriod: 0
  duration: "30s"
```

**What this chaos does:** Terminates the primary pod abruptly, forcing an immediate failover to a standby replica.

We are selecting the primary pod using label selector and killing it. The `duration` field specifies how long the chaos will last. In this case, we are killing the primary pod for 30 seconds.

Our expectation is that within 30 seconds, the primary pod will be killed, and one of the standby pods will be promoted to primary. The killed pod will be brought back by our PetSet operator and will join the cluster as a standby.

Before running, let's see who is the primary

```shell
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-0
```

Now run `watch kubectl get pg,petset,pods -n demo`.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 09:36:19 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d15h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d15h

NAME                  READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0   2/2     Running   0          3m44s
pod/sqlserver-ag-cluster-1   2/2     Running   0          59s
pod/sqlserver-ag-cluster-2   2/2     Running   0          57s

```

While watching the pods, run the chaos experiment.

```shell
kubectl apply -f primary-pod-kill.yaml
podchaos.chaos-mesh.org/pg-primary-pod-kill created
```

```shell
kubectl get pg,petset,pods -n demo
```

```shell
NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   2d15h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d15h

NAME                  READY   STATUS    RESTARTS     AGE
pod/sqlserver-ag-cluster-0   2/2     Running   1 (8s ago)   10s
pod/sqlserver-ag-cluster-1   2/2     Running   0            3m36s
pod/sqlserver-ag-cluster-2   2/2     Running   0            3m34s
```

Note the `Restarts` section; you will see the primary pod was
killed 8 seconds ago. The failover was done almost immediately.
The database state is now `Critical`, which
means your new primary is ready to accept connections, but one or
more of your replicas are not ready. The old primary will
be ready after `chaos.spec.duration` seconds, which is 30 seconds.

Let's see who is the new primary.


```shell
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-1

```

Now wait some time and you should see the old primary is back and the database state is `Ready` again.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 09:39:50 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d15h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d15h

NAME                  READY   STATUS    RESTARTS      AGE
pod/sqlserver-ag-cluster-0   2/2     Running   1 (62s ago)   64s
pod/sqlserver-ag-cluster-1   2/2     Running   0             4m30s
pod/sqlserver-ag-cluster-2   2/2     Running   0             4m28s
```

Now let's clean up the chaos experiment.

```shell
kubectl delete -f tests/01-pod-kill.yaml
podchaos.chaos-mesh.org "pg-primary-pod-kill" deleted
```

###  Chaos#2: OOMKill the Primary Pod

Now we are going to OOMKill the primary pod. This is a more realistic scenario than just killing the pod, because in real life, your primary pod might get OOMKilled due to high memory usage.

Save this yaml as `tests/02-oomkill.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: pg-primary-oom
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  stressors:
    memory:
      workers: 1
      size: "5000MB"  # Exceed the 3Gi limit to trigger OOM
  duration: "10m"

```

**What this chaos does:** Allocates excessive memory on the primary pod to exceed its limits, triggering an OOMKill that forces failover.


Before running this, we will run the load test job.

```shell
./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

We can see the database is in ready state while the load test job is running.
```shell
NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d16h
---------------------------------------------------------------
pod/pg-load-test-job-z8bxf   1/1     Running   0          22s
```

Let's see the log from the load test job:

```shell
➤ kubectl logs -f -n demo job/pg-load-test-job

Test Duration: 43s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 70123 (Reads: 55952, Inserts: 7053, Updates: 7118)
  Total Number of Rows Reads: 5595200, Inserts: 705300, Updates: 7118
  Total Errors: 0
  Total Data Transferred: 6548.86 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 1526.62 (Reads: 1219.18/s, Inserts: 158.02/s, Updates: 149.42/s)
  Throughput: 143.24 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 5.042ms, P95: 27.845ms, P99: 63.214ms
  Inserts - Avg: 50.112ms, P95: 128.465ms, P99: 274.199ms
  Updates - Avg: 22.783ms, P95: 87.802ms, P99: 211.079ms
-----------------------------------------------------------------
Connection Pool:
  Active: 20, Max: 100, Available: 80
=================================================================
```

Now run the chaos experiment.

```shell
> kubectl apply -f primary-oomkill.yaml
stresschaos.chaos-mesh.org/pg-primary-oom created
```

Now you should see the primary pod is OOMKilled and the failover happens. The database state will be `Critical` during the failover and will be `Ready` again after the old primary is back as standby.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 10:47:30 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   2d16h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d16h

NAME                         READY   STATUS    RESTARTS     AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0            54m
pod/sqlserver-ag-cluster-1          2/2     Running   1 (3s ago)   56m # NOTE: This shows the Restarts counter. It indicates that the pod is OOMKilled and restarted by Kubernetes
pod/sqlserver-ag-cluster-2          2/2     Running   0            54m
pod/pg-load-test-job-z8bxf   1/1     Running   0            113s

```

You can check the status of chaos experiment by running `kubectl get stresschaos -n chaos-mesh pg-primary-oom` command.

```shell
...
 status:
    conditions:
    - status: "True"
      type: Selected
    - status: "False"
      type: AllInjected
    - status: "True" # All chaos recovered
      type: AllRecovered
    - status: "False"
      type: Paused

```

Now after some time, you should see the old primary is back and the database state is `Ready` again.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 10:48:18 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d16h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d16h

NAME                         READY   STATUS    RESTARTS      AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0             55m
pod/sqlserver-ag-cluster-1          2/2     Running   1 (51s ago)   57m
pod/sqlserver-ag-cluster-2          2/2     Running   0             55m
pod/pg-load-test-job-z8bxf   1/1     Running   0             2m41s
```
Now check the data loss report from the load test job logs once the test is completed.

```shell
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 4095300
I0406 04:52:42.162937       1 load_generator_v2.go:555] Total records in table: 4095300
I0406 04:52:42.162960       1 load_generator_v2.go:556] totalRows in LoadGenerator: 4095300
  Records Found in DB: 4095300
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

Cleaning up test data...
Cleaning up test table...
Cleanup completed
Test data table deleted successfully

Test completed successfully!

```


Clean up the chaos experiment.

```shell
kubectl delete -f tests/02-oomkill.yaml
stresschaos.chaos-mesh.org "pg-primary-oom" deleted
```


###  Chaos#3: Kill SQL Server process in the Primary Pod

Now we are going to kill the postgres process in the primary pod. Save this yaml as `tests/03-kill-postgres-process.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pg-kill-postgres-process
  namespace: chaos-mesh
spec:
  action: container-kill
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  containerNames:
    - postgres
  duration: "30s"
```

**What this chaos does:** Forcefully terminates the SQL Server process in the primary container, simulating a database crash without pod termination.

Create the load test job. I will alter the duration of the load test job to 1 minute as this chaos experiment is generally shorter.

Just change the `TEST_RUN_DURATION: "60"` in the ConfigMap YAML and apply all the YAMLs again.

```shell
./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

```shell
pod/pg-load-test-job-79k9p   1/1     Running   0            10s # NOTE the load test job is running
```

Now run the chaos experiment.

```shell
kubectl apply -f pg-kill-postgres-process.yaml
podchaos.chaos-mesh.org/pg-kill-postgres-process created
```

As soon as you run the chaos experiment, you should see the primary pod is killed, the failover might/might not happen based on the possibility of data loss. If all the replica were synced up with primary before primary went down, a failover will happen immediately. Conversely, if there was some lag between primary and replica, there is a possibility of data loss and in that case, failover will not happen until the primary is back and the replica is synced up with primary.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 11:15:07 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   2d17h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d17h

NAME                         READY   STATUS    RESTARTS     AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0            81m
pod/sqlserver-ag-cluster-1          2/2     Running   2 (9s ago)   84m 
pod/sqlserver-ag-cluster-2          2/2     Running   0            81m
pod/pg-load-test-job-79k9p   1/1     Running   0            39s

```
You can see the primary pod was killed and restarted by Kubernetes. The failover was not performed and the database state is `NotReady`. The reason database didn't go ready is that chaos-mesh killed the postgres process immediately without giving the standby time to receive the last WAL the primary generated under high load. So there is a chance of data loss if we do a failover, so we are not doing a failover in this case to protect your data. However, there are APIs using which you can do a failover in this case also.
Now wait some time and you should see the old primary is back and the database state is `Ready` again.


```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 11:15:32 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d17h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d17h

NAME                         READY   STATUS    RESTARTS      AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0             82m
pod/sqlserver-ag-cluster-1          2/2     Running   2 (35s ago)   84m
pod/sqlserver-ag-cluster-2          2/2     Running   0             82m
pod/pg-load-test-job-79k9p   1/1     Running   0             65s

```

Now check the data loss report from the load test job logs once the test is completed.

```shell
Cumulative Statistics:
  Total Operations: 83211 (Reads: 66607, Inserts: 8355, Updates: 8249)
  Total Number of Rows Reads: 6660700, Inserts: 835500, Updates: 8249)
  Total Errors: 19548
  Total Data Transferred: 7790.99 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 1298.86 (Reads: 974.14/s, Inserts: 259.77/s, Updates: 64.94/s)
  Throughput: 129.14 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 5.366ms, P95: 30.093ms, P99: 72.567ms
  Inserts - Avg: 53.477ms, P95: 135.148ms, P99: 238.446ms
  Updates - Avg: 31.277ms, P95: 99.222ms, P99: 202.694ms
-----------------------------------------------------------------
Connection Pool:
  Active: 14, Max: 100, Available: 86
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 1327.47 operations/sec
  Read Operations: 66607 (1062.59/sec avg)
  Insert Operations: 8355 (133.29/sec avg)
  Update Operations: 8249 (131.60/sec avg)
  Error Rate: 19.0232%
  Total Data Transferred: 7.61 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 885500
  Records Found in DB: 885500
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

Cleaning up test data...
Cleaning up test table...
I0406 05:15:34.394443       1 load_generator_v2.go:555] Total records in table: 885500
I0406 05:15:34.394469       1 load_generator_v2.go:556] totalRows in LoadGenerator: 885500
Cleanup completed
Test data table deleted successfully

Test completed successfully!
```

Clean up the chaos experiment.

```shell
kubectl delete -f tests/03-kill-postgres-process.yaml
podchaos.chaos-mesh.org "pg-kill-postgres-process" deleted
```

###  Chaos#4: Primary Pod Failure

In this experiment, we are going to simulate a complete failure of the primary pod, including the node it is running on. This is a more extreme scenario than just killing the pod or the postgres process.

Save this yaml as `tests/04-pod-failure.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pg-primary-pod-failure
  namespace: chaos-mesh
spec:
  action: pod-failure
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  duration: "5m" 

```

**What this chaos does:** Removes the entrypoint which is running the postgres process. 

> **NOTE**: Chaos-Mesh will simulate a pod failure for `.spec.duration` amount of time; for our case, it is 5 minutes. As this simulates the complete failure of a pod for 5 minutes, our database will be in either a NotReady or Critical state for 5 minutes. Once this chaos is `Recovered`, the database will move back to `Ready` state automatically.

We will not run load tests for this experiment as well.

Before running this, let's examine the database state.

```shell
➤ kubectl get pg -n demo
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d17h
---------------------------------------------------------------
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-1 # Primary pod
```

See the primary pod is in running state.

```shell
pod/sqlserver-ag-cluster-0          2/2     Running     0             102m
pod/sqlserver-ag-cluster-1          2/2     Running     2 (21m ago)   105m
pod/sqlserver-ag-cluster-2          2/2     Running     0             102m
```

Now run the chaos experiment.
```shell
kubectl apply -f pg-primary-pod-failure.yaml
podchaos.chaos-mesh.org/pg-primary-pod-failure created
```

See the database went into NotReady state. Now based on the possibility of data loss, a failover will happen or be prohibited.

```shell
NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   2d17h
```

A failover happened immediately as there was no possibility of data loss. See the database is now in `Critical` state, which means the new primary is ready to accept connections, but one or more of the replicas are not ready, in this case, the old primary in not ready. The old primary will be ready after `chaos.spec.duration` seconds when chaos will be recovered, which is 5 minutes in our case.

```shell
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   2d17h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d17h

NAME                         READY   STATUS      RESTARTS      AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0             103m
pod/sqlserver-ag-cluster-1          2/2     Running     2 (22m ago)   106m
pod/sqlserver-ag-cluster-2          2/2     Running     0             103m
```

Let's see who is the new primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-0
```

Now let's wait 5 minutes and follow the status of the chaos experiment by running `kubectl get podchaos -n chaos-mesh pg-primary-pod-failure` command.

```shell
status:
  conditions:
  - status: "False"
    type: Paused
  - status: "True"
    type: Selected
  - status: "False"
    type: AllInjected
  - status: "True"
    type: AllRecovered

```

If you see `AllRecovered` condition is `True`, that means the chaos experiment is recovered, now you should see the old primary is back and the database state is `Ready` again.

```shell
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d17h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d17h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          106m
pod/sqlserver-ag-cluster-1          2/2     Running     4 (10m ago)   110m
pod/sqlserver-ag-cluster-2          2/2     Running     0          106m
```

Clean up the chaos experiment.

```shell
kubectl delete -f tests/04-pod-failure.yaml
podchaos.chaos-mesh.org "pg-primary-pod-failure" deleted
```

###  Chaos#5: Network Partition Primary Pod

> **NOTE**: The only possible way to avoid data loss in the network partition case is to use synchronous replication. You can do this by changing `db.spec.streamingMode: Synchronous`. In this case, there won't be any data loss.

> **Caution**: This experiment can cause data loss if you are using asynchronous replication. So use this experiment with caution and only on non-production environments.


In this experiment, we simulate a network partition affecting the primary pod in a SQL Server cluster.


Let's say we have a cluster with 3 nodes: one primary and two standbys. Now we are going to create a network partition between the primary and the standby pods. After the split, the primary will be in the minority partition and the standbys will be in the majority partition. 
```shell
Cluster (3 nodes)
-----------------
|  Partition A  |   Partition B   |
|---------------|-----------------|
|  primary-0    |  standby-1      |
|               |  standby-2      |
```
The primary will keep running as primary in the minority partition and one of the standbys will be promoted to primary in the majority partition. Because the majority quorum can't reach the primary in the minority partition due to network partition, they 
think the primary is down and they will promote one of the standby to primary by leader election.

```shell
After Split
-----------
| Partition A        | Partition B        |
|--------------------|--------------------|
| primary-0 (active) | standby-1 → primary|
|                    | standby-2          |
```

```shell
Partition Check
---------------
| Partition A  | Nodes: 1 |  No quorum |
| Partition B  | Nodes: 2 |  Has quorum |
```

We will detect this situation and will shutdown the primary in the minority partition to avoid data loss as much as possible.

```shell
Safe Outcome
------------
| Partition A        | Partition B        |
|--------------------|--------------------|
| primary-0 (stopped)| standby-1 → primary|
|                    | standby-2          |
```

But again, there exists a data loss window which is generally small (30s - 1 minute). So how much data might be lost? Depends on your write load during that time, might be none in case there wasn't any write load.

Now save this yaml as `tests/05-network-partition.yaml`. We will test this scenario against both asynchronous and synchronous replication mode and see the difference.

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-network-partition
  namespace: chaos-mesh
spec:
  action: partition
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
        "kubedb.com/role": "standby"
  direction: both
  duration: "4m"
```

**What this chaos does:** Blocks network connectivity between the primary pod and all standby pods, forcing a split-brain scenario where standbys promote a new primary in their partition while the isolated primary continues running.

Lets first test on the current postgres, which is running in asynchronous replication mode. Its basically the default mode if you have not mentioned anything in the `.spec.streamingMode` field of MSSQLServer Object.


Now lets first apply the load test job, but I will modify some config before running it.
```shell
BATCH_SIZE: "100"
TEST_RUN_DURATION: "600" # updated this, 10 minutes
INSERT_PERCENT: "1" # let's put some realistic write load, 1% of the operations will be insert
UPDATE_PERCENT: "19" # 19% of the operations will be update, so total write load is 20% which is quite high for postgres. We want to see some data loss in this case
CONCURRENT_WRITERS: "10" # Reduce the concurrent writers
```

Now,

```shell
./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Before running this experiment, lets examine db state.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 13:21:23 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d19h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d19h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          102s
pod/sqlserver-ag-cluster-1          2/2     Running   0          99s
pod/sqlserver-ag-cluster-2          2/2     Running   0          96s
pod/pg-load-test-job-ztb94   1/1     Running   0          12s

```

```shell
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-1 # Primary pod
```

Now let's go ahead and run the chaos experiment.

```shell
➤ kubectl apply -f tests/05-network-partition.yaml
networkchaos.chaos-mesh.org/pg-primary-network-partition created
```

Your database will be in `Ready` state for some time until we detect there is a network partition, when we detect the
network partition,
we shutdown the primary in the minority quorum. So you will see the database is in `Ready` state for some 
time and then it will go to `NotReady` or `Critical` state based on some other criteria.

```shell
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d19h
```

After some time, you should see the database is in `NotReady` state as we detected the network partition and shutdown the primary in the minority partition to avoid data loss as much as possible.

```shell
NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   2d19h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d19h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          3m55s
pod/sqlserver-ag-cluster-1          2/2     Running   0          3m52s
pod/sqlserver-ag-cluster-2          2/2     Running   0          3m49s
pod/pg-load-test-job-ztb94   1/1     Running   0          2m25s

```

Your database should be in `Critical` state after some time.

```shell

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   2d19h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d19h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          4m38s
pod/sqlserver-ag-cluster-1          2/2     Running   0          4m35s
pod/sqlserver-ag-cluster-2          2/2     Running   0          4m32s
pod/pg-load-test-job-ztb94   1/1     Running   0          3m8s
```
> **NOTE**: There is one possible way where data loss might be avoided even with asynchronous replication, this reason is somewhat weird but possible. In a scenario where the standby was lagging behind the primary before the network partition happened, there won't be a failover in this case as we know doing a failover will result in data loss in this case. In this case, your db will be in `NotReady` state.

> So, if you see your db is in **NotReady** state for longer period, this might be the reason, and you have successfully avoided data loss even with asynchronous replication at the cost of some downtime. Again, if you prefer uptime, use `.spec.replication.forceFailoverAcceptingDataLossAfter: 30s` which will force fully do a failover without considering data loss case.


Let's see the logs of the old primary, you should see the postgres process is shutdown immediately after the network partition is detected.

```shell
➤ kubectl logs -f -n demo sqlserver-ag-cluster-1
...
2026-04-06 07:23:50.190 UTC [2598] FATAL:  the database system is shutting down
2026-04-06 07:23:50.514 UTC [77] LOG:  checkpoint complete: wrote 24464 buffers (37.3%); 0 WAL file(s) added, 0 removed, 23 recycled; write=0.441 s, sync=0.036 s, total=0.519 s; sync files=44, longest=0.025 s, average=0.001 s; distance=376832 kB, estimate=376832 kB; lsn=8/48000028, redo lsn=8/48000028
2026-04-06 07:23:50.576 UTC [48] LOG:  database system is shut down
```
Let's check who is the new primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep  primary | awk '{print $1}'
sqlserver-ag-cluster-0
```

Check the logs of the new primary. It shows that it is now accepting connections, so new read/write operations will now go to the new primary.

```shell
➤ kubectl logs -f -n demo sqlserver-ag-cluster-0
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
...
2026-04-06 07:23:26.417 UTC [116] LOG:  database system is ready to accept connections
2026-04-06 07:23:26.864 UTC [160] LOG:  checkpoint complete: wrote 23062 buffers (35.2%); 0 WAL file(s) added, 0 removed, 17 recycled; write=0.407 s, sync=0.027 s, total=0.460 s; sync files=47, longest=0.011 s, average=0.001 s; distance=287175 kB, estimate=287175 kB; lsn=8/42873F88, redo lsn=8/42871FF0
2026-04-06 07:23:26.864 UTC [160] LOG:  checkpoint starting: immediate force wait
2026-04-06 07:23:26.871 UTC [160] LOG:  checkpoint complete: wrote 1 buffers (0.0%); 0 WAL file(s) added, 0 removed, 0 recycled; write=0.001 s, sync=0.002 s, total=0.007 s; sync files=1, longest=0.002 s, average=0.002 s; distance=8 kB, estimate=258459 kB; lsn=8/42874050, redo lsn=8/42874018
```

Now wait for the chaos experiment to be recovered, you can check the status of chaos experiment by running `kubectl get networkchaos -n chaos-mesh pg-primary-network-partition` command.

```shell
  status:
    conditions:
    - status: "True"
      type: Selected
    - status: "False"
      type: AllInjected
    - status: "True"
      type: AllRecovered
    - status: "False"
      type: Paused
```

Once `AllRecovered` is `True` you should see the old primary is back as standby and the database state is `Ready` again.

```shell
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d19h
```

Now let's see how many rows we lost in this case by checking the load test job logs.

```shell
Cumulative Statistics:
  Total Operations: 2371907 (Reads: 1897709, Inserts: 47445, Updates: 426753)
  Total Number of Rows Reads: 189770900, Inserts: 237225, Updates: 426753)
  Total Errors: 73
  Total Data Transferred: 192743.74 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 58.05 (Reads: 26.12/s, Inserts: 2.90/s, Updates: 29.03/s)
  Throughput: 2.81 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 5.23ms, P95: 36.986ms, P99: 48.888ms
  Inserts - Avg: 7.183ms, P95: 37.949ms, P99: 49.527ms
  Updates - Avg: 5.358ms, P95: 33.606ms, P99: 45.95ms
-----------------------------------------------------------------
Connection Pool:
  Active: 9, Max: 100, Available: 91
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 3934.44 operations/sec
  Read Operations: 1897709 (3147.85/sec avg)
  Insert Operations: 47445 (78.70/sec avg)
  Update Operations: 426753 (707.88/sec avg)
  Error Rate: 0.0031%
  Total Data Transferred: 188.23 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 270295
  Records Found in DB: 253365
  Records Lost: 16930
  Data Loss Percentage: 6.26%
=================================================================

⚠️  WARNING: 16930 records were inserted but not found in database!
```

So we incurred data loss. Now the question is: how much? In our above case, there were:
Insert Operations: 47445 (78.70/sec avg) -> 78.70 insert operations per second, each insert uses batch size of `BATCH_SIZE: 5`,
which is 78.70 * 5 = 393.5 rows inserted per second.

We lost 16930 rows, so the data loss window is 16930 / 393.5 = 43 seconds. 
So we can say that there was a network partition for around 4 minutes (`chaos.spec.duration`) 
and split brain due of network partition was for 43 seconds. 

This split brain detection time is around ~30 seconds(Your data loss window) despite how longer your network partition lasts.

> Note: If your network partition window is less than 30 seconds, you won't loose any data even in Asynchronous mode.

Now lets try to **avoid data loss** by using **Synchronous** replication. Change the `db.spec.streamingMode: Synchronous` to the setup/sqlserver-ag-cluster.yaml.

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: sqlserver-ag-cluster
  namespace: demo
spec:
  clientAuthMode: md5
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: postgres
          resources:
            limits:
              memory: 3Gi
            requests:
              cpu: 2
              memory: 2Gi
  replicas: 3
  replication:
    walKeepSize: 5000
    walLimitPolicy: WALKeepSize
  standbyMode: Hot
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Gi
  storageType: Durable
  streamingMode: Synchronous # Note this line
  version: "16.4"

```

Before applying this, let's clean up the previous yamls including postgres, load-test jobs and chaos experiment.

```shell
kubectl delete -f setup/sqlserver-ag-cluster.yaml
postgres.kubedb.com "sqlserver-ag-cluster" deleted
kubectl delete -f k8s/03-job.yaml
job.batch "pg-load-test-job" deleted
kubectl delete -f tests/05-network-partition.yaml
networkchaos.chaos-mesh.org "pg-primary-network-partition" deleted
```

Now apply the setup/sqlserver-ag-cluster.yaml and wait for the db to be in `Ready` state.

```shell
kubectl apply -f setup/sqlserver-ag-cluster.yaml
```


Once the db is in `Ready` state, apply the load test job and then wait 1 minute, then apply the chaos experiment. 



```shell
./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

First delete the previous experiment.

```shell
➤ kubectl delete -f tests/05-network-partition.yaml
networkchaos.chaos-mesh.org/pg-primary-network-partition deleted
```

Now apply the experiment again.

```shell
➤ kubectl apply -f tests/05-network-partition.yaml
networkchaos.chaos-mesh.org/pg-primary-network-partition created
```
You should experience the same scenario as before, but this time there won't be any data loss as we are using synchronous replication.


Let's wait and verify the logs from the load test job once the test is completed.

```shell
Final Results:
=================================================================
Test Duration: 10m3s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 1828426 (Reads: 1463758, Inserts: 36327, Updates: 328341)
  Total Number of Rows Reads: 146375800, Inserts: 181635, Updates: 328341)
  Total Errors: 42
  Total Data Transferred: 151400.85 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 62.77 (Reads: 51.69/s, Inserts: 3.69/s, Updates: 7.38/s)
  Throughput: 5.37 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 4.254ms, P95: 6.757ms, P99: 72.028ms
  Inserts - Avg: 10.912ms, P95: 59.609ms, P99: 72.482ms
  Updates - Avg: 14.94ms, P95: 18.156ms, P99: 70.674ms
-----------------------------------------------------------------
Connection Pool:
  Active: 10, Max: 100, Available: 90
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 3032.08 operations/sec
  Read Operations: 1463758 (2427.35/sec avg)
  Insert Operations: 36327 (60.24/sec avg)
  Update Operations: 328341 (544.49/sec avg)
  Error Rate: 0.0023%
  Total Data Transferred: 147.85 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 231635
  Records Found in DB: 231635
  Records Lost: 0
I0406 07:45:57.178325       1 load_generator_v2.go:555] Total records in table: 231635
I0406 07:45:57.178359       1 load_generator_v2.go:556] totalRows in LoadGenerator: 231635
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

```

See this time there is `no data loss`.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/05-network-partition.yaml
networkchaos.chaos-mesh.org "pg-primary-network-partition" deleted
```

Delete and recreate the postgres with asynchronous replication if you want to do more experiments.
Also revert back this changes

```shell
BATCH_SIZE: "100"
TEST_RUN_DURATION: "300" # updated this, 5 minutes
INSERT_PERCENT: "10" # lets put some relistic write load, 1% of the operations will be insert
UPDATE_PERCENT: "10" # 19% of the operations will be update, so total write load is 20% which is pretty high for postgres, we want to see some data loss in this case
CONCURRENT_WRITERS: "20" # Reduce the concurrent writters
```

###  Chaos#6: Limit bandwidth of Primary Pod

> As you changed `.db.spec.streamingMode: Synchronous` in the previous experiment, change it back to `Asynchronous` for this experiment. You can also keep it as it if you want though.

Skip these deletion process if you want to continue with `.db.spec.streamingMode: Synchronous`.

Edit the setup/sqlserver-ag-cluster.yaml and update `.db.spec.streamingMode: Asynchronous`

Now first delete the previous one,

```shell
kubectl delete -f setup/sqlserver-ag-cluster.yaml
```

Now wait untill all the pods of are gone. 

```
kubectl get pods -n demo | grep sqlserver-ag-cluster
# This should return nothing
```

Now apply the setup/sqlserver-ag-cluster.yaml,

```shell
kubectl apply -f setup/sqlserver-ag-cluster.yaml
```

Now wait until database is in ready state.

```
➤ kubectl get pg,pods -n demo
NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2m28s

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          2m22s
pod/sqlserver-ag-cluster-1          2/2     Running     0          2m15s
pod/sqlserver-ag-cluster-2          2/2     Running     0          2m8s


```

For this chaos experiment, we are going to limit the bandwidth of the primary pod. This will cause the replication lag between primary and standby to increase, which can lead to data loss if a failover happens during this time. So this is a good experiment to test the behavior of your cluster under network congestion.

Save this yaml as `tests/06-bandwidth-limit.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-bandwidth-limit
  namespace: chaos-mesh
spec:
  action: bandwidth
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  bandwidth:
    rate: "1mbps"
    limit: 20000
    buffer: 10000
  direction: both
  duration: "2m"
```

**What this chaos does:** Restricts the egress/ingress bandwidth of the primary pod to 1 Mbps, simulating a slow network connection and increasing replication lag.

Additionally, we will run the load test with some changes introduced.

```shell
INSERT_PERCENT: "19"
UPDATE_PERCENT: "1"
BATCH_SIZE: "200"
TEST_RUN_DURATION: "150"
```

Run the load generating job.

```shell
➤ ./run-k8s.sh 
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```
Now let's watch the pods and postgres. 

```shell
watch kubectl get pg,petset,pods -n demo
```


```shell
> watch -n demo kubectl get pg,petset,pods

Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 17:13:20 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    2d23h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   2d23h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          3h40m
pod/sqlserver-ag-cluster-1          2/2     Running   0          3h38m
pod/sqlserver-ag-cluster-2          2/2     Running   0          3h38m
pod/pg-load-test-job-hf85p   1/1     Running   0          105s

```

Your database should be in ready state all the time. Once the chaos experiment is completed, check the logs of load test job to see if there was any data loss.

```shell
Final Results:
=================================================================
Test Duration: 3m0s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 24564 (Reads: 19517, Inserts: 4803, Updates: 244)
  Total Number of Rows Reads: 1951700, Inserts: 960600, Updates: 244)
  Total Errors: 20
  Total Data Transferred: 3067.35 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 0.00 (Reads: 0.00/s, Inserts: 0.00/s, Updates: 0.00/s)
  Throughput: 0.00 MB/s
  Errors/sec: 2.75
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 13.334ms, P95: 49.334ms, P99: 336.744ms
  Inserts - Avg: 168.387ms, P95: 324.687ms, P99: 590.029ms
  Updates - Avg: 137.242ms, P95: 189.343ms, P99: 350.323ms
-----------------------------------------------------------------
Connection Pool:
  Active: 29, Max: 100, Available: 71
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 136.47 operations/sec
  Read Operations: 19517 (108.43/sec avg)
  Insert Operations: 4803 (26.68/sec avg)
  Update Operations: 244 (1.36/sec avg)
  Error Rate: 0.0814%
  Total Data Transferred: 3.00 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================
I0406 11:14:41.761915       1 load_generator_v2.go:555] Total records in table: 1014600
I0406 11:14:41.761938       1 load_generator_v2.go:556] totalRows in LoadGenerator: 1010600

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 1014600
  Records Found in DB: 1018600
  Records Lost: -4000
  Data Loss Percentage: -0.39%
=================================================================

 No data loss detected - all inserted records are present in database

```

Cleanup the chaos experiment.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/06-bandwidth-limit.yaml
networkchaos.chaos-mesh.org "pg-primary-bandwidth-limit" deleted
```

###  Chaos#7: Network Delay Primary Pod

In this chaos experiment, we are going to introduce network delay to the primary pod. This will cause the replication lag between primary and standby to increase, which can lead to data loss if a failover happens during this time. So this is a good experiment to test the behavior of your cluster under network congestion.

Save this yaml as `tests/07-network-delay.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-network-delay
  namespace: chaos-mesh
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  delay:
    latency: "500ms"
    jitter: "100ms"
    correlation: "50"
  duration: "3m"
  direction: both
```

**What this chaos does:** Adds 500ms latency with 100ms jitter to all network packets of the primary pod, simulating high-latency network conditions.

Let's adjust the load test config before running the load test job.

```shell
TEST_RUN_DURATION: "200"
READ_PERCENT: "80"
INSERT_PERCENT: "10"
UPDATE_PERCENT: "10"
BATCH_SIZE: "100"
```

Lets create the load test job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Now watch the pods and postgres status.


```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 18:39:12 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          83m
pod/sqlserver-ag-cluster-1          2/2     Running   0          83m
pod/sqlserver-ag-cluster-2          2/2     Running   0          83m
pod/pg-load-test-job-89flv   1/1     Running   0          72s
```

The database should be in `Ready` state all the time.

```shell
kubectl get networkchaos -n chaos-mesh -oyaml
...
  status:
    conditions:
    - status: "True"
      type: AllRecovered
    - status: "False"
      type: Paused
    - status: "True"
      type: Selected
    - status: "False"
      type: AllInjected
```

`AllRecovered` condition is `True`, that means the chaos experiment is done. Now let's check how many rows were inserted.

```shell
Final Results:
=================================================================
Test Duration: 3m23s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 253446 (Reads: 202535, Inserts: 25370, Updates: 25541)
  Total Number of Rows Reads: 20253500, Inserts: 2537000, Updates: 25541)
  Total Errors: 0
  Total Data Transferred: 23686.56 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 336.68 (Reads: 202.01/s, Inserts: 84.17/s, Updates: 50.50/s)
  Throughput: 30.12 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 8.76ms, P95: 64.215ms, P99: 98.166ms
  Inserts - Avg: 54.375ms, P95: 124.26ms, P99: 189.16ms
  Updates - Avg: 32.242ms, P95: 99.145ms, P99: 150.899ms
-----------------------------------------------------------------
Connection Pool:
  Active: 28, Max: 100, Available: 72
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 1250.27 operations/sec
  Read Operations: 202535 (999.12/sec avg)
  Insert Operations: 25370 (125.15/sec avg)
  Update Operations: 25541 (126.00/sec avg)
  Error Rate: 0.0000%
  Total Data Transferred: 23.13 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2587000
  Records Found in DB: 2587000
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

Cleaning up test data...
Cleaning up test table...
I0406 12:41:39.032079       1 load_generator_v2.go:555] Total records in table: 2587000
I0406 12:41:39.032102       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2587000
=================================================================

```
As you can see, 25M rows were inserted, 23GB data was transferred to the database and there was no data loss. So even with 500ms network delay, our cluster was able to handle the load and there was no data loss.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/07-network-delay.yaml
networkchaos.chaos-mesh.org "pg-primary-network-delay" deleted
```

###  Chaos#8: Network Loss Primary Pod

In this chaos experiment, we are going to introduce network loss to the primary pod. We expect our database to be able to hold Ready state, even though we see some failover, the end state of database should be `Ready`.

Save this yaml as `tests/08-network-loss.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-packet-loss
  namespace: chaos-mesh
spec:
  action: loss
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  loss:
    loss: "100"
    correlation: "100"
  duration: "3m"
  direction: both

```

**What this chaos does:** Drops 100% of network packets to/from the primary pod, simulating a complete network blackhole while allowing recovery when the chaos ends.

Lets run the load test job with some changes in config.

```shell
 TEST_RUN_DURATION: "200"
```

Lets create the load test job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Now create the chaos experiment.

```shell
➤ kubectl apply -f tests/08-network-loss.yaml
networkchaos.chaos-mesh.org/pg-primary-packet-loss created
```

Now watch the pods and postgres status.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:00:54 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          104m
pod/sqlserver-ag-cluster-1          2/2     Running   0          104m
pod/sqlserver-ag-cluster-2          2/2     Running   0          104m
pod/pg-load-test-job-44hg8   1/1     Running   0          96s

```

MSSQLServer should be in `Ready` state all the time, even though it switches to Critical, it should be back to `Ready` state after the experiment is done.

```shell
kubectl get networkchaos -n chaos-mesh -oyaml
...
  status:
    conditions:
    - status: "True"
      type: AllRecovered
    - status: "False"
      type: Paused
    - status: "True"
      type: Selected
    - status: "False"
      type: AllInjected

```

`AllRecovered` condition is `True`, that means the chaos experiment is done. Now let's check how many rows were inserted and if there was any data loss.

```shell
Final Results:
=================================================================
Test Duration: 3m23s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 229680 (Reads: 183614, Inserts: 23016, Updates: 23050)
  Total Number of Rows Reads: 18361400, Inserts: 2301600, Updates: 23050)
  Total Errors: 0
  Total Data Transferred: 21474.94 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 65.82 (Reads: 29.62/s, Inserts: 23.04/s, Updates: 13.16/s)
  Throughput: 5.62 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 13.045ms, P95: 50.368ms, P99: 218.188ms
  Inserts - Avg: 45.326ms, P95: 119.711ms, P99: 189.261ms
  Updates - Avg: 23.338ms, P95: 81.739ms, P99: 142.693ms
-----------------------------------------------------------------
Connection Pool:
  Active: 29, Max: 100, Available: 71
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 1131.63 operations/sec
  Read Operations: 183614 (904.67/sec avg)
  Insert Operations: 23016 (113.40/sec avg)
  Update Operations: 23050 (113.57/sec avg)
  Error Rate: 0.0000%
  Total Data Transferred: 20.97 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================
I0406 13:02:54.311885       1 load_generator_v2.go:555] Total records in table: 2351600
I0406 13:02:54.311910       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2351600

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2351600
  Records Found in DB: 2351600
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database
```

You can see the stats and this clearly shows lots of rows were inserted and reads were performed, but there was no data loss. And no downtime.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/08-network-loss.yaml
networkchaos.chaos-mesh.org "pg-primary-packet-loss" deleted
```

###  Chaos#9: Network Duplicate to Primary Pod

In this experiment, we will introduce packet duplication to the primary pod. We expect the database to be able to handle packet duplication and be ready all the time.

Save this yaml as `tests/09-network-duplicate.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-packet-duplicate
  namespace: chaos-mesh
spec:
  action: duplicate
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  duplicate:
    duplicate: "50"
    correlation: "25"
  duration: "4m"
  direction: both

```

**What this chaos does:** Duplicates 50% of network packets to/from the primary pod, creating redundant traffic that can overwhelm or confuse the receiving end.

Lets run the load test job with some changes in config.

```shell
 TEST_RUN_DURATION: "240"
```

```shell
saurov@saurov-pc:~/g/s/g/s/high-write-load-client|main⚡*?
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Now lets create the chaos experiment.

```shell
➤ kubectl apply -f tests/09-network-duplicate.yaml
networkchaos.chaos-mesh.org/pg-primary-packet-duplicate created
```

Now watch the pods and postgres status.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:19:44 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          123m
pod/sqlserver-ag-cluster-1          2/2     Running     0          123m
pod/sqlserver-ag-cluster-2          2/2     Running     0          123m
```

You should see your database is in `Ready` state all the time despite the packet duplication.

```shell
kubectl get networkchaos -n chaos-mesh -oyaml
...
  status:
    conditions:
    - status: "True"
      type: AllInjected
    - status: "False"
      type: AllRecovered
    - status: "False"
      type: Paused
    - status: "True"
      type: Selected

```
Once the experiment is done, check the logs of load test job to see if there was any data loss.

```shell
Final Results:
=================================================================
Test Duration: 3m23s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 224846 (Reads: 179994, Inserts: 22547, Updates: 22305)
  Total Number of Rows Reads: 17999400, Inserts: 2254700, Updates: 22305)
  Total Errors: 0
  Total Data Transferred: 21050.70 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 151.38 (Reads: 105.97/s, Inserts: 22.71/s, Updates: 22.71/s)
  Throughput: 13.61 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 13.234ms, P95: 71.521ms, P99: 237.409ms
  Inserts - Avg: 46.457ms, P95: 115.989ms, P99: 193.325ms
  Updates - Avg: 24.757ms, P95: 91.271ms, P99: 157.036ms
-----------------------------------------------------------------
Connection Pool:
  Active: 14, Max: 100, Available: 86
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 1109.38 operations/sec
  Read Operations: 179994 (888.08/sec avg)
  Insert Operations: 22547 (111.25/sec avg)
  Update Operations: 22305 (110.05/sec avg)
  Error Rate: 0.0000%
  Total Data Transferred: 20.56 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2304700
  Records Found in DB: 2304700
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

Cleaning up test data...
Cleaning up test table...
I0406 13:15:39.100598       1 load_generator_v2.go:555] Total records in table: 2304700
I0406 13:15:39.100624       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2304700
=================================================================
```

As usual, despite load on the database and packet duplication, there was no data loss and database was in `Ready` state all the time.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/09-network-duplicate.yaml
networkchaos.chaos-mesh.org "pg-primary-packet-duplicate" deleted
```

###  Chaos#10: Network Corruption to Primary Pod

In this experiment, we will introduce packet corruption to the primary pod. We expect the database to be able to handle packet corruption and not lose any data.

Save this yaml as `tests/10-network-corrupt.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: pg-primary-packet-corrupt
  namespace: chaos-mesh
spec:
  action: corrupt
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  target:
    mode: all
    selector:
      namespaces:
        - demo
      labelSelectors:
        "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  corrupt:
    corrupt: "50"
    correlation: "25"
  duration: "4m"
  direction: both

```

**What this chaos does:** Corrupts 50% of network packets to/from the primary pod by flipping random bits in the payload, causing checksums to fail.

Lets change some config and apply the load test creation script.

```shell
 TEST_RUN_DURATION: "240"
```

```shellsaurov@saurov-pc:~/g/s/g/s/high-write-load-client|main⚡*?
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created 
persistentvolumeclaim/pg-load-test-results created
``` 

Now check if the database is in ready state.

```shell
kubectl get pg,petset,pods -n demo
```

```shell
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          126m
pod/sqlserver-ag-cluster-1          2/2     Running   0          126m
pod/sqlserver-ag-cluster-2          2/2     Running   0          126m
pod/pg-load-test-job-lftl8   1/1     Running   0          6s

```

Now create the chaos experiment.

```yaml➤ kubectl apply -f tests/10-network-corrupt.yaml
networkchaos.chaos-mesh.org/pg-primary-packet-corrupt created
```

Now watch the pods and postgres status.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:27:48 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          131m
pod/sqlserver-ag-cluster-1          2/2     Running     0          131m
pod/sqlserver-ag-cluster-2          2/2     Running     0          131m

```

The database is ready so far, and sqlserver-ag-cluster-0 is the primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
```



```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:35:09 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          139m
pod/sqlserver-ag-cluster-1          2/2     Running   0          139m
pod/sqlserver-ag-cluster-2          2/2     Running   0          139m
pod/pg-load-test-job-5q4gh   1/1     Running   0          52s

```
Database turns into `NotReady` state as a failover happens due of the corruption.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:35:52 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          139m
pod/sqlserver-ag-cluster-1          2/2     Running   0          139m
pod/sqlserver-ag-cluster-2          2/2     Running   0          139m
pod/pg-load-test-job-5q4gh   1/1     Running   0          95s

```

A new primary is elected and database moved into `Critical` state, which means new primary is ready to accept connections.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-1
```

So sqlserver-ag-cluster-1 is the new primary. Wait for the chaos to be recovered.

```shell
 kubectl get networkchaos -n chaos-mesh -oyaml
 ...
  status:
    conditions:
    - status: "False"
      type: AllInjected
    - status: "True"
      type: AllRecovered
    - status: "False"
      type: Paused
    - status: "True"
      type: Selected

```

`Alrecovered` true means chaos experiment is over.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:36:25 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          140m
pod/sqlserver-ag-cluster-1          2/2     Running   0          140m
pod/sqlserver-ag-cluster-2          2/2     Running   0          140m
pod/pg-load-test-job-5q4gh   1/1     Running   0          2m8s

```
The database has returned to `Ready` state.


Now check the stats of data insertion and read.

```shell
Final Results:
=================================================================
Test Duration: 3m23s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 241642 (Reads: 193608, Inserts: 23878, Updates: 24156)
  Total Number of Rows Reads: 19360800, Inserts: 2387800, Updates: 24156)
  Total Errors: 0
  Total Data Transferred: 22600.28 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 395.64 (Reads: 237.39/s, Inserts: 138.48/s, Updates: 19.78/s)
  Throughput: 39.88 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 10.469ms, P95: 64.429ms, P99: 103.762ms
  Inserts - Avg: 51.473ms, P95: 134.532ms, P99: 201.798ms
  Updates - Avg: 29.607ms, P95: 98.249ms, P99: 169.741ms
-----------------------------------------------------------------
Connection Pool:
  Active: 27, Max: 100, Available: 73
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 1192.38 operations/sec
  Read Operations: 193608 (955.36/sec avg)
  Insert Operations: 23878 (117.83/sec avg)
  Update Operations: 24156 (119.20/sec avg)
  Error Rate: 0.0000%
  Total Data Transferred: 22.07 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================
I0406 13:26:04.117504       1 load_generator_v2.go:555] Total records in table: 2437800
I0406 13:26:04.117540       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2437800

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2437800
  Records Found in DB: 2437800
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database
```
So everything looks alright. No data loss.

Cleanup the chaos experiment:

```shell
kubectl delete -f tests/10-network-corrupt.yaml
networkchaos.chaos-mesh.org "pg-primary-packet-corrupt" deleted
```

###  Chaos#11: Time Offset and DNS error

We will run two chaos experiments one after another in this case. No load test will be run in these two cases.

Save this yaml as `tests/11-time-offset.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: TimeChaos
metadata:
  name: pg-primary-time-offset
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  timeOffset: "-2h"
  clockIds:
    - CLOCK_REALTIME
  duration: "2m"

```

**What this chaos does:** Shifts the system clock of the primary pod back by 2 hours, simulating time skew that can cause certificate validation, timestamp-based logic, and replication synchronization issues.

Save this yaml as `tests/12-dns-error.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: pg-primary-dns-error
  namespace: chaos-mesh
spec:
  action: error
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  duration: "2m"

```

**What this chaos does:** Makes all DNS queries from the primary pod fail with resolution errors, simulating DNS service outage or misconfiguration.

```shell
➤ kubectl apply -f tests/11-time-offset.yaml
timechaos.chaos-mesh.org/pg-primary-time-offset created
saurov@saurov-pc:~/g/s/g/s/chaos-mesh|main⚡*
➤ kubectl apply -f tests/12-dns-error.yaml 
dnschaos.chaos-mesh.org/pg-primary-dns-error created
```

Your database will be in ready state through the whole chaos.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 19:50:14 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    3d1h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   3d1h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          154m
pod/sqlserver-ag-cluster-1          2/2     Running     0          154m
pod/sqlserver-ag-cluster-2          2/2     Running     0          154m
```

Clean up the chaos experiments.

```shell
kubectl delete -f tests/11-time-offset.yaml 
timechaos.chaos-mesh.org "pg-primary-time-offset" deleted
kubectl delete -f tests/12-dns-error.yaml 
dnschaos.chaos-mesh.org "pg-primary-dns-error" deleted
```

## IO chaos

### MSSQLServer Recreation with force failover

For IO related chaos, if you prioritize high availability over data loss, then set 
`.spec.replication.forceFailoverAcceptingDataLossAfter: 30s`. This will result in better availability. But if you prefer data safety over high availability, then do not set `.spec.replication.forceFailoverAcceptingDataLossAfter: 30s` this.


I will set`.spec.replication.forceFailoverAcceptingDataLossAfter: 30s` for IO related chaos tests. 
**If you do not prefer data loss, ignore this Recreation step**.


You will see that even though we will force failover accepting the possibility that there might be data loss,
but in really this data loss chances are very not very high. We should be able to achieve high availability without losing any data in most cases. Our end goal is to have 
the database in `Ready` state when chaos is recovered.

> NOTE: In case you do not prefer to set this `.spec.replication.forceFailoverAcceptingDataLossAfter: 30s`.
> Its just you might face some extra downtime(Database might stay in `NotReady` state for longer period 
> until chaos is recovered) in some IOChaos cases.



First delete the setup/sqlserver-ag-cluster.yaml

```shell
kubectl delete -f setup/sqlserver-ag-cluster.yaml
```
Wait untill all the pods are deleted.

```shell
kubectl get pods -n demo | grep sqlserver-ag-cluster
# this should not return anything
```

Now update your setup/sqlserver-ag-cluster.yaml with below yaml.

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: sqlserver-ag-cluster
  namespace: demo
spec:
  clientAuthMode: md5
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: postgres
          resources:
            limits:
              memory: 3Gi
            requests:
              cpu: 2
              memory: 2Gi
  replicas: 3
  replication:
    walKeepSize: 5000
    walLimitPolicy: WALKeepSize
    forceFailoverAcceptingDataLossAfter: 30s # New added
  standbyMode: Hot
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Gi
  storageType: Durable
  version: "16.4"
```

Run `kubectl apply -f setup/sqlserver-ag-cluster.yaml` and wait for database to be in ready state.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 20:17:16 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    54s

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   48s

NAME                  READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0   2/2     Running   0          48s
pod/sqlserver-ag-cluster-1   2/2     Running   0          43s
pod/sqlserver-ag-cluster-2   2/2     Running   0          38s

```

Let's check which pod is the primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-2
```

> Note: If you performed this step, you might need to change the k8s/02-secret.yaml -> DB_PASSWORD: 'new value'

###  Chaos#12: IO latency

In this experiment, we will simulate IO latency. Our end goal is to have as low downtime as possible and the database should be in `Ready` state when chaos is recovered.

Save this yaml as `tests/13-io-latency.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: pg-primary-io-latency 
  namespace: chaos-mesh
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  volumePath: /var/pv
  path: /var/pv/data/**/*
  delay: "500ms"
  percent: 100
  duration: "5m"
  containerNames:
    - postgres
```

**What this chaos does:** Injects 500ms latency into all disk I/O operations on the primary pod, simulating slow storage that increases replication lag and can trigger failover.


Lets change the load test config.

```shell
  TEST_RUN_DURATION: "300"
```

In case your database password is changed(you recreated the postgres and used WipeOut deletion policy), you can run the below command to check your database password.

```shell
➤ kubectl get secret -n demo sqlserver-ag-cluster-auth -oyaml
apiVersion: v1
data:
  password: bVApIWcyYW5PcV9ONXR+bQ==
  username: cG9zdGdyZXM=
kind: Secret
...
```

Check if your database password given in the secret of load test yaml is changed or not. If changed, then update the password and apply the secret again.

```shell
DB_PASSWORD: bVApIWcyYW5PcV9ONXR+bQ==
```

```shell
➤ kubectl apply -f k8s/02-secret.yaml
secret/pg-load-test-secret configured
```

Now apply the load test yamls.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config configured
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Now wait 10-20 second and apply the chaos experiment.

```shell
➤ kubectl apply -f tests/13-io-latency.yaml
iochaos.chaos-mesh.org/pg-primary-io-latency created
```

Soon after we created the chaos test, the database should be in `NotReady` state. The reason for this is, the client call to `Primary` pod is getting timed
out due of slow IO. 

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 20:37:24 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   21m

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   20m

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          5m55s
pod/sqlserver-ag-cluster-1          2/2     Running   0          5m52s
pod/sqlserver-ag-cluster-2          2/2     Running   0          5m49s
pod/pg-load-test-job-62l88   1/1     Running   0          2m10s

```

Now we might observe some interesting behavior as the IO is not performing correctly. We might see frequent failovers and a possible split brain situation. However, this won't last long.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
sqlserver-ag-cluster-2
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
sqlserver-ag-cluster-2
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
sqlserver-ag-cluster-2
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
sqlserver-ag-cluster-2
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-2
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
saurov@saurov-pc:~
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
```

After some amount of time, we should see a stable primary, in our case which is `sqlserver-ag-cluster-0`.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 20:40:04 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   23m

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   23m

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          8m34s
pod/sqlserver-ag-cluster-1          2/2     Running   0          8m31s
pod/sqlserver-ag-cluster-2          2/2     Running   0          8m28s
pod/pg-load-test-job-62l88   1/1     Running   0          4m49s
```

Now, the database is in critical state. We will wait untill the chaos is recovered.

```shell
➤ kubectl get iochaos -n chaos-mesh -oyaml
...
  status:
    conditions:
    - status: "True"
      type: Selected
    - status: "False"
      type: AllInjected
    - status: "True"
      type: AllRecovered
    - status: "False"
      type: Paused

```
The chaos is recovered. Now the database should be in `Ready` state. But if anything goes terribly wrong because of slow IO, you might find a database in either `NotReady` and `Critical` state. In this case, contact with us.


```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Mon Apr  6 20:48:52 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    32m

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   32m

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          17m
pod/sqlserver-ag-cluster-1          2/2     Running     0          17m
pod/sqlserver-ag-cluster-2          2/2     Running     0          17m
pod/pg-load-test-job-62l88   0/1     Completed   0          13m

```

So the database is transitioned into `Ready` state as soon as the chaos was recovered.

```shell
Final Results:
=================================================================
Test Duration: 5m3s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 214087 (Reads: 171304, Inserts: 21358, Updates: 21425)
  Total Number of Rows Reads: 17130400, Inserts: 2135800, Updates: 21425)
  Total Errors: 15175
  Total Data Transferred: 20024.37 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 45.03 (Reads: 40.53/s, Inserts: 2.25/s, Updates: 2.25/s)
  Throughput: 4.46 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 14.961ms, P95: 55.724ms, P99: 153.02ms
  Inserts - Avg: 59.45ms, P95: 126.812ms, P99: 218.533ms
  Updates - Avg: 27.391ms, P95: 90.496ms, P99: 180.472ms
-----------------------------------------------------------------
Connection Pool:
  Active: 11, Max: 100, Available: 89
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 706.23 operations/sec
  Read Operations: 171304 (565.10/sec avg)
  Insert Operations: 21358 (70.46/sec avg)
  Update Operations: 21425 (70.68/sec avg)
  Error Rate: 6.6191%
  Total Data Transferred: 19.56 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
I0406 14:40:33.621876       1 load_generator_v2.go:555] Total records in table: 2185700
I0406 14:40:33.621913       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2185800
-----------------------------------------------------------------
  Total Records Inserted: 2185700
  Records Found in DB: 2185600
  Records Lost: 100
  Data Loss Percentage: 0.00%
=================================================================

⚠️  WARNING: 100 records were inserted but not found in database!
This may indicate:
  - Database crash/restart occurred during test
  - pg_rewind was triggered due to network partition
  - Transaction rollback due to replication issues
```

Total number of rows inserted 2135800, lost rows 100, so basically 1 batch insert query was lost.
**If you have not set force failover, this data loss won't be there**.

> **NOTE**: The same chaos experiment is run again in the [`IO Chaos Tests Without Force Failover`](#io-chaos-tests-without-force-failover) section below without the `forceFailoverAcceptingDataLossAfter: 30s` API. In that case, no data loss was incurred.


Clean up the chaos experiment.

```shell
kubectl delete -f tests/13-io-latency.yaml
iochaos.chaos-mesh.org "pg-primary-io-latency" deleted
```

###  Chaos#13: IO Fault to primary

In this experiment, chaos-mesh will insert io/fault. Our database should handle this chaos and remain in `Ready` or `Critical` state. 
Once the chaos is recovered by chaos-mesh, the database should be back in `Ready` state.

Save this yaml as `tests/14-io-fault.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: pg-primary-io-fault
  namespace: chaos-mesh
spec:
  action: fault
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  volumePath: /var/pv
  path: /var/pv/data/**/*
  errno: 5  # EIO (Input/output error)
  percent: 50
  duration: "5m"
  containerNames:
    - postgres
```

**What this chaos does:** Injects I/O errors (EIO) on 50% of disk operations to the primary pod, simulating disk hardware failures or filesystem corruption.

Let's see how our database is now,

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:00:56 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    11h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   11h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          11h
pod/sqlserver-ag-cluster-1          2/2     Running     0          11h
pod/sqlserver-ag-cluster-2          2/2     Running     0          11h
pod/pg-load-test-job-62l88   0/1     Completed   0          11h
```

Let's see who is primary:

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
➤ kubectl exec -it -n demo sqlserver-ag-cluster-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sqlserver-ag-cluster-0:/$ psql
psql (16.4)
Type "help" for help.

postgres=# select pg_is_in_recovery();
 pg_is_in_recovery 
-------------------
 f
(1 row)

```

Lets now create the load generate job,

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Wait 15-20 second and then apply the io-fault yaml.

```shell
➤ kubectl apply -f tests/14-io-fault.yaml
iochaos.chaos-mesh.org/pg-primary-io-fault created
```

keep watching the database and pods,

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:05:39 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   11h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   11h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          11h
pod/sqlserver-ag-cluster-1          2/2     Running   0          11h
pod/sqlserver-ag-cluster-2          2/2     Running   0          11h
pod/pg-load-test-job-pq4l6   1/1     Running   0          117s
```

After running for some time, the database went into critical state. Let's see if there is a failover.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-1
➤ kubectl exec -it -n demo sqlserver-ag-cluster-1 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sqlserver-ag-cluster-1:/$ psql
psql (16.4)
Type "help" for help.

postgres=# select pg_is_in_recovery();
 pg_is_in_recovery 
-------------------
 f
(1 row)

```

There is a failover, and we can run queries on the new primary. Things are looking good so far.

I will show you what happened to old primary due to i/o error.

```shell
➤ kubectl logs -n demo sqlserver-ag-cluster-0
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
...
2026-04-07 02:04:14.564 UTC [2813] LOG:  all server processes terminated; reinitializing
2026-04-07 02:04:14.564 UTC [2813] LOG:  could not open directory "base/pgsql_tmp": I/O error
2026-04-07 02:04:14.564 UTC [2813] LOG:  could not open directory "base": I/O error
2026-04-07 02:04:14.564 UTC [2813] LOG:  could not open directory "pg_tblspc": I/O error
2026-04-07 02:04:14.643 UTC [2813] PANIC:  could not open file "global/pg_control": I/O error
/scripts/run.sh: line 61:  2813 Aborted                 (core dumped) /run_scripts/role/run.sh
removing the initial scripts as server is not running ...

```

So as it wasn't able to operate cleanly and communicate with standby's, a new leader election happened and `sqlserver-ag-cluster-1` was promoted as primary.
As we saw earlier, we can run queries on `sqlserver-ag-cluster-1`, so our cluster is usable even in the time of chaos.

Now wait until chaos is recovered.

```shell
➤ kubectl get iochaos -n chaos-mesh pg-primary-io-fault -oyaml
...
status:
  conditions:
  - status: "False"
    type: AllInjected
  - status: "True"
    type: AllRecovered
  - status: "False"
    type: Paused
  - status: "True"
    type: Selected

```

Chaos is recovered by chaos-mesh.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:11:28 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    11h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   11h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          11h
pod/sqlserver-ag-cluster-1          2/2     Running     0          11h
pod/sqlserver-ag-cluster-2          2/2     Running     0          11h
pod/pg-load-test-job-pq4l6   0/1     Completed   0          7m47s
```

Our database is transitioned back into `Ready` state.

```shell
Final Results:
  Total Data Transferred: 24419.35 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 1115.64 (Reads: 893.05/s, Inserts: 111.29/s, Updates: 111.29/s)
  Throughput: 104.36 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 16.1ms, P95: 60.276ms, P99: 295.835ms
  Inserts - Avg: 42.911ms, P95: 117.868ms, P99: 204.391ms
  Updates - Avg: 18.051ms, P95: 65.577ms, P99: 131.106ms
-----------------------------------------------------------------
Connection Pool:
  Active: 29, Max: 100, Available: 71
=================================================================
=================================================================
Test Duration: 5m3s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 260920 (Reads: 208927, Inserts: 26033, Updates: 25960)
  Total Number of Rows Reads: 20892700, Inserts: 2603300, Updates: 25960)
  Total Errors: 242129
  Total Data Transferred: 24420.59 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 35.96 (Reads: 32.97/s, Inserts: 3.00/s, Updates: 0.00/s)
  Throughput: 3.71 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 16.102ms, P95: 60.291ms, P99: 295.835ms
  Inserts - Avg: 42.912ms, P95: 117.868ms, P99: 204.391ms
  Updates - Avg: 18.051ms, P95: 65.577ms, P99: 131.106ms
-----------------------------------------------------------------
Connection Pool:
  Active: 11, Max: 100, Available: 89
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 861.35 operations/sec
  Read Operations: 208927 (689.71/sec avg)
  Insert Operations: 26033 (85.94/sec avg)
  Update Operations: 25960 (85.70/sec avg)
  Error Rate: 48.1323%
  Total Data Transferred: 23.85 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2653300
  Records Found in DB: 2653300
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database
```

You can see the statistics here, 25 GB was inserted in 5 minutes with zero data loss even though we accepted data loss via `forceFailoverAcceptingDataLossAfter`.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/14-io-fault.yaml
iochaos.chaos-mesh.org "pg-primary-io-fault" deleted
```

###  Chaos#14: IO attribute overwrite

In this experiment, i/o attributes will be overwritten. We expect our database to be available (`Ready` | `Critical`) during the chaos experiment.

> Note: If you are not using `forceFailoverAcceptingDataLossAfter`, then you might see the database is in `NotReady` during the chaos.

Save this yaml as `tests/15-io-attr-override.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: pg-primary-io-attr-override
  namespace: chaos-mesh
spec:
  action: attrOverride
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  volumePath: /var/pv
  path: /var/pv/data/**/*
  attr:
    perm: 444  # Read-only permissions
  percent: 100
  duration: "4m"
  containerNames:
    - postgres

```

**What this chaos does:** Overrides file permissions on data files to read-only (444), preventing write operations and forcing the database to encounter permission denied errors on all writes.

Let's see how our database is now.
```shell
kubectl get pg,petset,pods -n demo
```
```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:32:42 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          12h
pod/sqlserver-ag-cluster-1          2/2     Running   0          12h
pod/sqlserver-ag-cluster-2          2/2     Running   0          12h

```

Create the load generation job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Apply the chaos experiment.

```shell
➤ kubectl apply -f tests/15-io-attr-override.yaml
iochaos.chaos-mesh.org/pg-primary-io-attr-override created
```

Keep watching the database resources.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:33:45 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          12h
pod/sqlserver-ag-cluster-1          2/2     Running   0          12h
pod/sqlserver-ag-cluster-2          2/2     Running   0          12h
pod/pg-load-test-job-cgbgt   1/1     Running   0          72s
```

So the database went into `NotReady` state, which means the primary is not responsive. The reason might be that the database inside the primary pod is not running.

Let's check this:

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-1
```

Let's check the logs from unresponsive primary `sqlserver-ag-cluster-1`.

```shell
➤ kubectl logs -f -n demo sqlserver-ag-cluster-1
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
...
2026-04-07 02:33:20.552 UTC [237694] FATAL:  the database system is in recovery mode
2026-04-07 02:33:20.553 UTC [2908] LOG:  all server processes terminated; reinitializing
2026-04-07 02:33:20.553 UTC [2908] LOG:  could not open directory "base/pgsql_tmp": Permission denied
2026-04-07 02:33:20.554 UTC [2908] LOG:  could not open directory "base/4": Permission denied
2026-04-07 02:33:20.554 UTC [2908] LOG:  could not open directory "base/5": Permission denied
2026-04-07 02:33:20.554 UTC [2908] LOG:  could not open directory "base/1": Permission denied
2026-04-07 02:33:20.627 UTC [2908] PANIC:  could not open file "global/pg_control": Permission denied
removing the initial scripts as server is not running ...
/scripts/run.sh: line 61:  2908 Aborted                 (core dumped) /run_scripts/role/run.sh

```

So you can see primary is shut down for I/O chaos. A failover should happen soon.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:34:42 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          12h
pod/sqlserver-ag-cluster-1          2/2     Running   0          12h
pod/sqlserver-ag-cluster-2          2/2     Running   0          12h
pod/pg-load-test-job-cgbgt   1/1     Running   0          2m9s

```

Our database now moved to `NotReady` -> `Critical` state. Let's see who is the new primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
-----
➤ kubectl exec -it -n demo sqlserver-ag-cluster-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sqlserver-ag-cluster-0:/$ psql
psql (16.4)
Type "help" for help.

postgres=# select pg_is_in_recovery();
 pg_is_in_recovery 
-------------------
 f
(1 row)
----
➤ kubectl logs -f -n demo sqlserver-ag-cluster-0
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
...
2026-04-07 02:34:38.753 UTC [368446] LOG:  checkpoint starting: wal
2026-04-07 02:35:09.342 UTC [368446] LOG:  checkpoint complete: wrote 20389 buffers (31.1%); 0 WAL file(s) added, 0 removed, 21 recycled; write=29.948 s, sync=0.444 s, total=30.589 s; sync files=11, longest=0.351 s, average=0.041 s; distance=539515 kB, estimate=618146 kB; lsn=4/FACB3C48, redo lsn=4/DD03C9B0
2026-04-07 02:35:12.932 UTC [368446] LOG:  checkpoint starting: wal
2026-04-07 02:35:29.121 UTC [368446] LOG:  checkpoint complete: wrote 22745 buffers (34.7%); 0 WAL file(s) added, 2 removed, 33 recycled; write=15.541 s, sync=0.535 s, total=16.190 s; sync files=12, longest=0.216 s, average=0.045 s; distance=540559 kB, estimate=610387 kB; lsn=5/1C15E728, redo lsn=4/FE0207D8

```

So database is back online again, however old primary has not yet joined in the cluster. We will wait until all the chaos recovered.

```shell
➤ kubectl get iochaos -n chaos-mesh pg-primary-io-attr-override -oyaml
...
status:
  conditions:
  - status: "True"
    type: AllRecovered
  - status: "False"
    type: Paused
  - status: "True"
    type: Selected
  - status: "False"
    type: AllInjected
```

All the generated chaos has been recovered.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:38:52 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          12h
pod/sqlserver-ag-cluster-1          2/2     Running     0          12h
pod/sqlserver-ag-cluster-2          2/2     Running     0          12h
pod/pg-load-test-job-cgbgt   0/1     Completed   0          6m20s

```

Database moved into `Ready` state.

```shell
Final Results:
=================================================================
Test Duration: 5m3s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 248425 (Reads: 198683, Inserts: 24948, Updates: 24794)
  Total Number of Rows Reads: 19868300, Inserts: 2494800, Updates: 24794)
  Total Errors: 232435
  Total Data Transferred: 23243.14 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 48.77 (Reads: 36.58/s, Inserts: 4.88/s, Updates: 7.32/s)
  Throughput: 4.34 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 16.922ms, P95: 52.069ms, P99: 317.142ms
  Inserts - Avg: 44.849ms, P95: 130.692ms, P99: 211.022ms
  Updates - Avg: 19.201ms, P95: 66.474ms, P99: 148.452ms
-----------------------------------------------------------------
Connection Pool:
  Active: 25, Max: 100, Available: 75
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 820.04 operations/sec
  Read Operations: 198683 (655.85/sec avg)
  Insert Operations: 24948 (82.35/sec avg)
  Update Operations: 24794 (81.84/sec avg)
  Error Rate: 48.3374%
  Total Data Transferred: 22.70 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================
=================================================================
Test Duration: 5m13s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 248425 (Reads: 198683, Inserts: 24948, Updates: 24794)
  Total Number of Rows Reads: 19868300, Inserts: 2494800, Updates: 24794)
  Total Errors: 232435
  Total Data Transferred: 23243.14 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 0.00 (Reads: 0.00/s, Inserts: 0.00/s, Updates: 0.00/s)
  Throughput: 0.00 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 16.922ms, P95: 52.069ms, P99: 317.142ms
  Inserts - Avg: 44.849ms, P95: 130.692ms, P99: 211.022ms
  Updates - Avg: 19.201ms, P95: 66.474ms, P99: 148.452ms
-----------------------------------------------------------------
Connection Pool:
  Active: 14, Max: 100, Available: 86
=================================================================

I0407 02:37:53.535684       1 load_generator_v2.go:555] Total records in table: 2544800
=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2544800
I0407 02:37:53.535709       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2544800
  Records Found in DB: 2544800
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database
```

We inserted around 23 GB in 5 minutes. No data loss detected.

Clean up the chaos experiment.

```shell
kubectl delete -f tests/15-io-attr-override.yaml 
iochaos.chaos-mesh.org "pg-primary-io-attr-override" deleted
```

###  Chaos#15: IO mistake

In this experiment, chaos-mesh will insert IO mistakes. We expect the database to be in `Ready` state after the chaos is recovered. If you are using the `forceFailover` API, then your database will be up even when chaos is running, but this will increase the chance of some data loss (if some write operations are going on during the failover process).

Just to remind you, we are using `forceFailoverAcceptingDataLossAfter` API for IO related chaos.

Save this yaml as `tests/16-io-mistake.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: pg-primary-io-mistake
  namespace: chaos-mesh
spec:
  action: mistake
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  volumePath: /var/pv
  path: /var/pv/data/**/*
  mistake:
    filling: random
    maxOccurrences: 10
    maxLength: 100
  percent: 50
  duration: "5m"
  containerNames:
    - postgres
```

**What this chaos does:** Randomly injects garbage data (random bytes) into file operations on 50% of disk writes, corrupting the data stored on disk.

Let's check the database state.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:57:53 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          12h
pod/sqlserver-ag-cluster-1          2/2     Running     0          12h
pod/sqlserver-ag-cluster-2          2/2     Running     0          12h

```

Running the load generation job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Lets check the primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
saurov@saurov-pc:~
➤ kubectl exec -it -n demo sqlserver-ag-cluster-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sqlserver-ag-cluster-0:/$ psql
psql (16.4)
Type "help" for help.

postgres=# select pg_is_in_recovery();
 pg_is_in_recovery 
-------------------
 f
(1 row)
```

Lets apply the experiment.

```shell
➤ kubectl apply -f tests/16-io-mistake.yaml 
iochaos.chaos-mesh.org/pg-primary-io-mistake created
```

Keep watching the database.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 08:59:26 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          12h
pod/sqlserver-ag-cluster-1          2/2     Running   0          12h
pod/sqlserver-ag-cluster-2          2/2     Running   0          12h
pod/pg-load-test-job-b56q6   1/1     Running   0          75s

```

Database went into NotReady state and should be back in `Critical` state as we used `forceFailoverAcceptingDataLossAfter` api.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:00:32 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          12h
pod/sqlserver-ag-cluster-1          2/2     Running   0          12h
pod/sqlserver-ag-cluster-2          2/2     Running   0          12h
pod/pg-load-test-job-b56q6   1/1     Running   0          2m21s

```

The database is back in `Critical` state.

```shell
➤ kubectl get iochaos -n chaos-mesh pg-primary-io-mistake -oyaml
status:
  conditions:
  - status: "True"
    type: Selected
  - status: "False"
    type: AllInjected
  - status: "True"
    type: AllRecovered
  - status: "False"
    type: Paused

```

All the chaos recovered.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:04:55 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    12h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   12h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          12h
pod/sqlserver-ag-cluster-1          2/2     Running     0          12h
pod/sqlserver-ag-cluster-2          2/2     Running     0          12h
pod/pg-load-test-job-b56q6   0/1     Completed   0          6m44
```

Database back in `Ready` state.

```shell
...
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 2537700
I0407 03:03:27.372254       1 load_generator_v2.go:556] totalRows in LoadGenerator: 2537700
  Records Found in DB: 2537700
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database
```

No data loss.

Cleanup:

```shell
➤ kubectl delete -f tests/16-io-mistake.yaml 
iochaos.chaos-mesh.org "pg-primary-io-mistake" deleted
```


### IO Chaos Tests Without Force Failover

We have seen data losses in chaos tests with `forceFailoverAcceptingDataLossAfter: 30s` api, so we will now try the same chaos,
but without this api.

Now save this yaml at `setup/sqlserver-ag-cluster.yaml`
```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: sqlserver-ag-cluster
  namespace: demo
spec:
  clientAuthMode: md5
  deletionPolicy: Delete
  podTemplate:
    spec:
      containers:
        - name: postgres
          resources:
            limits:
              memory: 3Gi
            requests:
              cpu: 2
              memory: 2Gi
  replicas: 3
  replication:
    walKeepSize: 3000
    walLimitPolicy: WALKeepSize
  standbyMode: Hot
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
  storageType: Durable
  version: "16.4"
```

Now apply this yaml `kubectl apply -f setup/sqlserver-ag-cluster.yaml`.

watch the resource coming up and db getting `Ready`.

```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Wed Apr  8 10:15:14 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    68s

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   63s

NAME                  READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0   2/2     Running   0          63s
pod/sqlserver-ag-cluster-1   2/2     Running   0          56s
pod/sqlserver-ag-cluster-2   2/2     Running   0          48s

```

lets see who is the primary.

```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-0
-----
➤ kubectl exec -it -n demo sqlserver-ag-cluster-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sqlserver-ag-cluster-0:/$ psql
psql (16.4)
Type "help" for help.

postgres=# select pg_is_in_recovery();
 pg_is_in_recovery 
-------------------
 f
(1 row)

```

lets run the load generate job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Apply the io-latency chaos experiment.

```shell
➤ kubectl apply -f tests/13-io-latency.yaml 
iochaos.chaos-mesh.org/pg-primary-io-latency created
```

Now watch the database state.

```shell
watch kubectl get pg,petset,pods -n demo
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Wed Apr  8 10:20:38 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   6m32s

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   6m27s

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          6m27s
pod/sqlserver-ag-cluster-1          2/2     Running   0          6m20s
pod/sqlserver-ag-cluster-2          2/2     Running   0          6m12s
pod/pg-load-test-job-p7vvw   1/1     Running   0          80s

```

You should see your database is in `NotReady` state all the time. The reason behind that:
- Primary database is up and running, but as IO latency increased, new connection creation is getting timed out.
- All existing connections to the primary are working fine.
- Primary postgres process are working fine, that's why we are not doing a failover.
- So new connections during this test wasn't possible, and as we do not used force failover, no failover performed.

```shell
status:
  conditions:
  - status: "True"
    type: Selected
  - status: "False"
    type: AllInjected
  - status: "True"
    type: AllRecovered
  - status: "False"
    type: Paused

```

Now the chaos is recovered and our database should eventually reach `Ready` state.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Wed Apr  8 10:33:12 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    19m

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   19m

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          53s
pod/sqlserver-ag-cluster-1          2/2     Running     0          18m
pod/sqlserver-ag-cluster-2          2/2     Running     0          18m
pod/pg-load-test-job-p7vvw   0/1     Completed   0          13m
```

The database reached in `Ready` state.

```shell
Final Results:
=================================================================
Test Duration: 6m10s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 72980 (Reads: 58363, Inserts: 7250, Updates: 7367)
  Total Number of Rows Reads: 5836300, Inserts: 725000, Updates: 7367)
  Total Errors: 17
  Total Data Transferred: 6820.02 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 0.00 (Reads: 0.00/s, Inserts: 0.00/s, Updates: 0.00/s)
  Throughput: 0.00 MB/s
  Errors/sec: 0.36
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 42.497ms, P95: 35.26ms, P99: 48.91ms
  Inserts - Avg: 53.254ms, P95: 109.955ms, P99: 266.726ms
  Updates - Avg: 93.204ms, P95: 91.333ms, P99: 260.728ms
-----------------------------------------------------------------
Connection Pool:
  Active: 14, Max: 100, Available: 86
=================================================================

=================================================================
Performance Summary:
  Average Throughput: 197.22 operations/sec
  Read Operations: 58363 (157.72/sec avg)
  Insert Operations: 7250 (19.59/sec avg)
  Update Operations: 7367 (19.91/sec avg)
  Error Rate: 0.0233%
  Total Data Transferred: 6.66 GB
=================================================================

=================================================================
Checking for Data Loss...
=================================================================
Error getting connection stats: failed to get current connections: pq: canceling statement due to user request
Error getting connection stats: failed to get max_connections: context deadline exceeded
I0408 04:30:19.253637       1 load_generator_v2.go:555] Total records in table: 775000
I0408 04:30:19.253658       1 load_generator_v2.go:556] totalRows in LoadGenerator: 775000

=================================================================
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 775000
  Records Found in DB: 775000
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================

 No data loss detected - all inserted records are present in database

```

From the load generate job, we can see there was less data inserted as database was unavailable. But more importantly,
**No data loss** was recorded.

Similarly, you can try the other chaos also. You should find out no data loss for each io chaos cases.

Cleanup:
```shell
➤ kubectl delete -f tests/13-io-latency.yaml
iochaos.chaos-mesh.org "pg-primary-io-latency" deleted
```

## Misc Chaos Tests

###  Chaos#16: Node Reboot | Stress CPU memory

We will perform three experiments one after another here. We will not run load tests for some of these experiments.

Save this yaml as `tests/17-node-reboot.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pg-cluster-all-pods-kill
  namespace: chaos-mesh
spec:
  action: pod-kill
  mode: all
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
  gracePeriod: 0
  duration: "30s"

```

**What this chaos does:** Simultaneously kills all SQL Server pods in the cluster, simulating a complete node failure where all replicas restart at once.

This is simulate a typical node failure scenario where all the pod restarted.

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:31:47 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          13h
pod/sqlserver-ag-cluster-1          2/2     Running     0          13h
pod/sqlserver-ag-cluster-2          2/2     Running     0          13h
```

Lets apply the experiment.

```shell
kubectl apply -f tests/17-node-reboot.yaml
```


```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:32:12 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          5s
pod/sqlserver-ag-cluster-1          2/2     Running     0          2s

```
```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:32:24 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      NotReady   13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          16s
pod/sqlserver-ag-cluster-1          2/2     Running     0          13s
pod/sqlserver-ag-cluster-2          2/2     Running     0          11s
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:32:33 2026

NAME                                VERSION   STATUS     AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Critical   13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          26s
pod/sqlserver-ag-cluster-1          2/2     Running     0          23s
pod/sqlserver-ag-cluster-2          2/2     Running     0          21s
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:32:40 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS      RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running     0          32s
pod/sqlserver-ag-cluster-1          2/2     Running     0          29s
pod/sqlserver-ag-cluster-2          2/2     Running     0          27s

```

So the database is back in ready state within 30s of applying the chaos. Now let's apply the next chaos which will stress CPU.

Now lets try to stress the cpu.

Save this yaml as `tests/18-stress-cpu-primary.yaml`:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: pg-primary-cpu-stress
  namespace: chaos-mesh
spec:
  mode: one
  selector:
    namespaces:
      - demo
    labelSelectors:
      "app.kubernetes.io/instance": "sqlserver-ag-cluster"
      "kubedb.com/role": "primary"
  stressors:
    cpu:
      workers: 2
      load: 90
  duration: "2m"

```

**What this chaos does:** Stresses the CPU on the primary pod by running 2 CPU-intensive worker processes at 90% load, consuming system resources and potentially causing slowdowns and failover.

But before running this, we will run the load test job.

```shell
➤ ./run-k8s.sh
job.batch "pg-load-test-job" deleted
persistentvolumeclaim "pg-load-test-results" deleted
configmap/pg-load-test-config unchanged
job.batch/pg-load-test-job created
persistentvolumeclaim/pg-load-test-results created
```

Now lets apply the chaos experiment.

```shell
➤ kubectl apply -f tests/18-stress-cpu-primary.yaml
stresschaos.chaos-mesh.org/pg-primary-cpu-stress created
```
```shell
➤ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1}'
sqlserver-ag-cluster-1

```

Lets check the cpu usages:

```shell
Every 2.0s: kubectl top pods --containers -n demo          saurov-pc: Tue Apr  7 09:35:42 2026

POD                      NAME             CPU(cores)   MEMORY(bytes)
sqlserver-ag-cluster-0          pg-coordinator   29m          40Mi
sqlserver-ag-cluster-0          postgres         244m         621Mi
sqlserver-ag-cluster-1          pg-coordinator   15m          38Mi
sqlserver-ag-cluster-1          postgres         7060m        693Mi
sqlserver-ag-cluster-2          pg-coordinator   16m          38Mi
sqlserver-ag-cluster-2          postgres         217m         629Mi
pg-load-test-job-sfj6z   load-test        1594m        216Mi

```

```shell
watch kubectl top pods --containers -n demo
```

```shell
Every 2.0s: kubectl top pods --containers -n demo          saurov-pc: Tue Apr  7 09:35:58 2026

POD                      NAME             CPU(cores)   MEMORY(bytes)
sqlserver-ag-cluster-0          pg-coordinator   29m          37Mi
sqlserver-ag-cluster-0          postgres         272m         633Mi
sqlserver-ag-cluster-1          pg-coordinator   15m          38Mi
sqlserver-ag-cluster-1          postgres         8509m        941Mi
sqlserver-ag-cluster-2          pg-coordinator   14m          39Mi
sqlserver-ag-cluster-2          postgres         241m         657Mi
pg-load-test-job-sfj6z   load-test        1256m        272Mi

```

Database remain in ready state as there was sufficient cpu left in the cluster. However, this test case will pass in every environment.


```shell
watch kubectl get pg,petset,pods -n demo
```

```shell
Every 2.0s: kubectl get pg,petset,pods -n demo             saurov-pc: Tue Apr  7 09:36:31 2026

NAME                                VERSION   STATUS   AGE
postgres.kubedb.com/sqlserver-ag-cluster   16.4      Ready    13h

NAME                                         AGE
petset.apps.k8s.appscode.com/sqlserver-ag-cluster   13h

NAME                         READY   STATUS    RESTARTS   AGE
pod/sqlserver-ag-cluster-0          2/2     Running   0          4m24s
pod/sqlserver-ag-cluster-1          2/2     Running   0          4m21s
pod/sqlserver-ag-cluster-2          2/2     Running   0          4m19s
pod/pg-load-test-job-sfj6z   1/1     Running   0          113s
```


```shell
Data Loss Report:
-----------------------------------------------------------------
  Total Records Inserted: 3273100
  Records Found in DB: 3273100
  Records Lost: 0
  Data Loss Percentage: 0.00%
=================================================================
I0407 03:40:04.019990       1 load_generator_v2.go:555] Total records in table: 3273100
I0407 03:40:04.020008       1 load_generator_v2.go:556] totalRows in LoadGenerator: 3273100

 No data loss detected - all inserted records are present in database

```

CleanUp:

```shell
kubectl delete -f tests/17-node-reboot.yaml 
podchaos.chaos-mesh.org "pg-cluster-all-pods-kill" deleted
kubectl delete -f tests/18-stress-cpu-primary.yaml 
stresschaos.chaos-mesh.org "pg-primary-cpu-stress" deleted
```

## Chaos Testing Results Summary

### Test Results Overview

Below is a comprehensive summary of all chaos engineering experiments conducted on the KubeDB-managed SQL Server High-Availability cluster. Each metric shows results in two configurations:
- **With Force Failover**: Using `forceFailoverAcceptingDataLossAfter: 30s`
- **Without Force Failover**: Waiting for data consistency before failover

> Note: You might see different results if you have tested under no read/write load.

| # | Experiment | Failure Mode | Failover Time | Data Loss | Downtime | Notes |
|---|---|---|---|---|---|---|
| 1 | Kill Primary Pod | Pod termination | **With**: ~8s  **Without**: ~8s | **With**:  0 / **Without**:  0 | **With**: Minimal / **Without**: Minimal | Immediate failover works in both cases |
| 2 | OOMKill Primary Pod | Memory exhaustion | **With**: ~3s / **Without**: ~3s | **With**:  0 / **Without**:  0 | **With**: Minimal / **Without**: Minimal | Rapid failover, 4.1M rows inserted |
| 3 | Kill SQLServer Process | Process crash | **With**: ~30s / **Without**: ~30s+ | **With**:  0 / **Without**:  0 | **With**: ~30s / **Without**: 40s | Blocks failover to prevent data loss in both cases |
| 4 | Primary Pod Failure | Network isolation | **With**: ~10s / **Without**: ~10s | **With**:  0 / **Without**:  0 | **With**: Minimal / **Without**: Minimal | Split-brain handled well |
| 5 | Network Partition | Complete isolation | **With**: ~30s / **Without**: ~30s+ | **With**: ⚠️ Possible / **Without**: ⚠️ Possible | **With**: Brief / **Without**: Extended | Split-brain scenario, data safety challenge in both |
| 6 | Bandwidth Limit (1 Mbps) | Slow network | **With**: No failover / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: 0s / **Without**: 0s | 2.3M rows inserted, high latency tolerated |
| 7 | Network Delay (500ms) | High latency | **With**: No failover / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: 0s / **Without**: 0s | 2.5M rows inserted, consistency maintained |
| 8 | Network Loss (100%) | Packet drop | **With**: No failover / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: 0s / **Without**: 0s | 2.3M rows inserted, no data loss |
| 9 | Network Duplicate (50%) | Redundant traffic | **With**: No failover / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: 0s / **Without**: 0s | 2.2M rows inserted, gracefully handled |
| 10 | Network Corruption (50%) | Corrupted packets | **With**: ~15s / **Without**: ~15s | **With**:  0 / **Without**:  0 | **With**: ~30s / **Without**: ~30s | 2.1M rows inserted, checksums fail |
| 11 | Time Offset & DNS Error | System time shift | **With**: No failover / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: 0s / **Without**: 0s | 2.0M rows inserted |
| 12 | IO Latency | Disk I/O delay | **With**: ~30s / **Without**: No failover | **With**: ⚠️ ~1 insert loss / **Without**:  0 | **With**: Brief / **Without**: Extended | Critical difference: force failover causes ~1 insert loss |
| 13 | IO Fault (50%) | I/O errors | **With**: ~30s / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: Brief / **Without**: Extended | 2.6M rows inserted, 25GB transferred |
| 14 | IO Attribute Override | Filesystem attr change | **With**: ~30s / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: Brief / **Without**: Extended | 2.5M rows inserted, 23GB transferred |
| 15 | IO Mistake | Random I/O faults | **With**: ~30s / **Without**: No failover | **With**:  0 / **Without**:  0 | **With**: Brief / **Without**: Extended | 2.5M rows inserted, 23GB transferred |
| 16 | Node Reboot (All Pods) | Complete node failure | **With**: ~30s / **Without**: ~30s+ | **With**:  0 / **Without**:  0 | **With**: Extended / **Without**: Extended | 3.2M rows inserted, full cluster restart |

> Note: `Extended` means as long as the chaos runs.

### Key Findings

#### Replication Strategy Impact

| Scenario | With Force Failover (30s) | Without Force Failover | Winner |
|---|---|---|---|
| **Availability** | High - immediate failover | Lower - waits for consistency | 
| **Data Loss Risk** | Low-Medium |  Zero Risk 
| **IO Chaos Tests** | ⚠️ 1 insert lost (rare) |  0 insert lost | 
| **Failover Time** | 30 seconds or less | Variable (extended if unsafe) |
| **Use Case** | High-availability priority | Data integrity priority | 

#### Chaos Test Categories

**1. Pod-Level Failures (Chaos #1-4)**
- **Result**: Immediate failovers work well
- **Data Loss**: Zero in all cases
- **Downtime**: Minimal (< 30s recovery)
- **Best Practice**: Default configuration handles these excellently

**2. Network Chaos (Chaos #5-11)**
- **Result**: Cluster remains stable without failover
- **Data Loss**: Zero in all cases (except network partition which forces split-brain)
- **Downtime**: Minimal to none (connections recover automatically)
- **Best Practice**: SQL Server's replication is resilient to network impairments

**3. IO Chaos with Force Failover (Chaos #12-15, with `forceFailoverAcceptingDataLossAfter`)**
- **Result**: Cluster stays highly available
- **Data Loss**: Minimal (~1 insert in worst case = 0.004%)
- **Downtime**: Minimal (automatic failover keeps cluster up)
- **Trade-off**: Sacrifices tiny bit of data for high availability

**4. IO Chaos Without Force Failover (Chaos #12-15, without `forceFailoverAcceptingDataLossAfter`)**
- **Result**: Database may enter NotReady state
- **Data Loss**: Zero across all tests
- **Downtime**: Extended (until chaos clears or manual intervention)
- **Trade-off**: Prioritizes data safety over availability

### Replication Configuration Recommendations

**Choose WITH `forceFailoverAcceptingDataLossAfter: 30s` if:**
- Your application requires high availability
- You can tolerate rare events of < 0.01% data loss
- Your database serves real-time or customer-facing services

**Choose WITHOUT `forceFailoverAcceptingDataLossAfter` if:**
- Data integrity is fine, but not critical
- You can tolerate extended downtime during node failures
- Your database serves compliance-sensitive operations

**Choose WITHOUT `streamingMode: Synchronous` if:**
- Data integrity is absolutely critical
- You want high availability
- Your database serves compliance-sensitive operations

### Performance Metrics Summary in chaos cases

| Metric | Average | Best | Worst |
|---|---|---|---|
| **Rows Inserted** | 2.3M | 4.1M | 0.7M |
| **Data Transferred** | 21.5 GB | 25 GB | 6.6 GB |
| **Failover Time** | ~20 seconds | ~3 seconds | 30+ seconds |
| **Data Loss (with Force Failover)** | < 0.01% | 0% | 0.004% |
| **Data Loss (without Force Failover)** |  0% |  0% |  0% |
| **Recovery Time** | < 1 minutes | ~30 seconds | ~5 minutes |

> **Important Note**: All these metrics are taken during chaos experiment. Kubedb performs notably well in both chaos scenarios and normal scenarios. For example
> In normal scenarios where you kubernetes cluster is behaving normal, you should see a failover happening within `~5 seconds` without `any data loss` every time, and of course automatically.  

### Conclusion

The KubeDB-managed SQL Server HA cluster demonstrates excellent resilience across all tested failure scenarios.

The cluster achieves the balance between high availability and data consistency, allowing operators to choose their preferred trade-off based on business requirements.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.06.23/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.06.23/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).


---