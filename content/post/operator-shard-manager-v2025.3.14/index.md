---
title: Announcing Operator-Shard-Manager v2025.03.14
date: "2025-03-27"
weight: 14
authors:
- Saurov Chandra Biswas
tags:
- alert
- archiver
- autoscaler
- backup
- cassandra
- clickhouse
- cloud-native
- dashboard
- database
- druid
- grafana
- kafka
- kubedb
- kubernetes
- kubestash
- memcached
- mongodb
- mssqlserver
- network
- networkpolicy
- pgbouncer
- pgpool
- postgres
- postgresql
- prometheus
- rabbitmq
- redis
- restore
- s3
- secret
- security
- singlestore
- solr
- tls
- virtual-secret
- zookeeper
---

# **Operator Shard Manager** (Introduced in `2025.3.14`)

## **Use Case**

The **Operator Shard Manager** is designed using **Consistent Hashing with bounded loads algorithm** to efficiently distribute workloads among multiple controller instances. **This workload distribution must be stateless.**

Lets try to understand how this works.

### **Installing KubeDB**

Let's assume you have installed `kubedb-v2025.3.24` in your Kubernetes cluster. If you haven't installed it yet, follow the instructions [here](https://kubedb.com/docs/v2025.3.24/setup/install/kubedb/).

```bash
➤ helm ls -n kubedb
NAME  	NAMESPACE	REVISION	UPDATED                                	STATUS	CHART            	APP VERSION
kubedb	kubedb   	1       	2025-03-25 10:25:05.598350871 +0600 +06	Succeeded	kubedb-v2025.3.24	v2025.3.24 
```
This confirms that `kubedb-v2025.3.24` is installed on the cluster.

### **Checking Installed StatefulSets**

```bash
➤ kubectl get sts -n kubedb
NAME                        READY   AGE
kubedb-kubedb-autoscaler    1/1     2d1h
kubedb-kubedb-ops-manager   1/1     2d1h
kubedb-kubedb-provisioner   1/1     2d1h
```
After installation, three StatefulSets are created along with few deployments and few other resources.

### **Checking Created Pods**

```bash
➤ kubectl get pods -n kubedb
kubedb-kubedb-autoscaler-0                      1/1     Running   1 (41h ago)   2d1h
kubedb-kubedb-ops-manager-0                     1/1     Running   2 (41h ago)   2d1h
kubedb-kubedb-provisioner-0                     1/1     Running   1 (41h ago)   41h
kubedb-kubedb-webhook-server-79cf5d496f-c29f7   1/1     Running   1 (41h ago)   2d1h
kubedb-petset-7f4474897b-w5hqg                  1/1     Running   1 (41h ago)   2d1h
kubedb-sidekick-78fbd9947c-bw6b2                1/1     Running   1 (41h ago)   2d1h
```

We can see only single `kubedb-kubedb-provisioner-0` pod has been created.

---

## **Creating a PostgreSQL Custom Resource for understanding sharding**

With KubeDB installed, you can now create a **Postgres** custom resource to provision a PostgreSQL cluster.

### **Postgres Custom Resource (CR)**

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  storageType: Durable
  podTemplate:
    spec:
      restartPolicy: Always
      containers:
      - name: postgres
        resources:
          requests:
            cpu: "300m"
            memory: "356Mi"
      - name: pg-coordinator
        resources:
          requests:
            cpu: "150m"
            memory: "100Mi"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 7Gi
  version: "17.4"
```

### **Applying the Custom Resource**

```bash
➤ kubectl apply -f ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

### **Checking Created Resources**

```bash
➤ kubectl get pods,pg,svc,secrets,pvc -n demo
NAME                READY   STATUS    RESTARTS   AGE
pod/ha-postgres-0   2/2     Running   0          13m
pod/ha-postgres-1   2/2     Running   0          12m
pod/ha-postgres-2   2/2     Running   0          12m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   17.4      Ready    13m

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/ha-postgres           ClusterIP   10.43.93.78     <none>        5432/TCP,2379/TCP            13m
service/ha-postgres-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   13m
service/ha-postgres-standby   ClusterIP   10.43.225.204   <none>        5432/TCP                     13m

NAME                                   TYPE                       DATA   AGE
secret/ha-postgres-auth                kubernetes.io/basic-auth   2      45h

NAME                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-ha-postgres-0   Bound    pvc-95135495-3a9e-469a-9a50-f60d7726c60e   7Gi        RWO            longhorn       <unset>                 13m
persistentvolumeclaim/data-ha-postgres-1   Bound    pvc-c9c14a21-af97-45da-8ca4-14de70c705c5   7Gi        RWO            longhorn       <unset>                 12m
persistentvolumeclaim/data-ha-postgres-2   Bound    pvc-a28e95a4-ebb2-4186-b992-c3b157b61074   7Gi        RWO            longhorn       <unset>                 12m
```

As you can see, KubeDB has created multiple resources such as **services, PVCs, secrets, and pods**.

---

## **Why Do We Need Operator Sharding?**

Everything works fine when you have a **reasonable number of custom resources** in your cluster. Like 5 or 10 `Postgres` CR.

### **But what happens when you have thousands of custom resources?**

This is where **Operator Sharding** becomes essential! Because managing these much resources might be troublesome for a 
single operator. Also it makes debugging harder if you have this much resources processed by a single operator.


### **How Operator Sharding Works**

Instead of a single provisioner handling all custom resources, **Operator Shard Manager** distributes the workload among multiple controller pods.

For example, if you scale up the `kubedb-kubedb-provisioner` StatefulSet:

```bash
kubectl scale sts -n kubedb kubedb-kubedb-provisioner --replicas=3
```
You will get:

- `kubedb-kubedb-provisioner-0`
- `kubedb-kubedb-provisioner-1`
- `kubedb-kubedb-provisioner-2`

Each of these pods can **independently** handle different custom resources if you use **operator shard manager**.

### **Why Scaling Alone without shard-manager is Not Enough?**

If you **only** scale up the StatefulSet, all the provisioner pods will process the same custom resources **in parallel**, which can lead to conflicts, inconsistent state, and unnecessary resource consumption.

### **How Operator Shard Manager Helps**

With **Operator Shard Manager**, each **custom resource is assigned to a specific controller pod**. This ensures:

- **No duplicate processing**  
- **Efficient workload distribution**  
- **Better scalability and reliability**

Now, instead of all provisioner pods competing for the same resources, **each custom resource is handled by a single assigned controller**, ensuring smooth and conflict-free execution.

Enough of talks, lets have some hands on. For that we will need to upgrade the kubedb like below:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2025.3.24 \
  --namespace kubedb --create-namespace \
  --set operator-shard-manager.enabled=true \
  --set kubedb-provisioner.replicaCount=3 \
  --set-file global.license=license.txt \
  --wait --burst-limit=10000 --debug
```
> make sure you change the license.txt with correct license.

- `--set operator-shard-manager.enabled=true` will install the `Operator Shard Manager` in your cluster.

- `--set kubedb-provisioner.replicaCount=3` will upgrade your cluster with 3 provisioner pods(`kubedb-kubedb-provisioner-0`, `kubedb-kubedb-provisioner-1`, `kubedb-kubedb-provisioner-2`).

Once you are done with the upgrade, you should also get a `ShardConfiguration` object in your cluster.

```bash
➤ kubectl get shardconfiguration kubedb-provisioner -oyaml
```
```yaml
apiVersion: operator.k8s.appscode.com/v1alpha1
kind: ShardConfiguration
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2025-03-25T04:26:48Z"
  generation: 1
  labels:
    app.kubernetes.io/managed-by: Helm
  name: kubedb-provisioner
  resourceVersion: "502805"
  uid: 0334416e-ea7d-4ed1-b25e-570023ca04a1
spec:
  controllers:
  - apiGroup: apps
    kind: StatefulSet
    name: kubedb-kubedb-provisioner
    namespace: kubedb
  resources:
  - apiGroup: kubedb.com
  - apiGroup: elasticsearch.kubedb.com
  - apiGroup: kafka.kubedb.com
  - apiGroup: postgres.kubedb.com
status:
  controllers:
  - apiGroup: apps
    kind: StatefulSet
    name: kubedb-kubedb-provisioner
    namespace: kubedb
    pods:
    - kubedb-kubedb-provisioner-0
    - kubedb-kubedb-provisioner-1
    - kubedb-kubedb-provisioner-2

```

### How Sharding Works
As the installation done, lets figure out how this works.

The `ShardConfiguration` has two parts in the specification:
- **Controllers**: `spec.controllers[*]` defines which StatefulSets/Deployments/DaemonSets will be managing/processing resources. We will then get the pods that were created by this StatefulSets/Deployments/DaemonSets and update the `status.controllers`.
- **Resources**: `.spec.resources` defines which API group (like `kubedb.com`) objects are sharded.

Example for Postgres (API group: `kubedb.com`) and `kubedb-kubedb-provisioner` StatefulSet:
```yaml
spec:
  controllers:
  - apiGroup: apps
    kind: StatefulSet
    name: kubedb-kubedb-provisioner
    namespace: kubedb
  resources:
  - apiGroup: kubedb.com
```

The status shows which pods are handling resources:
```yaml
status:
  controllers:
  - pods:
    - kubedb-kubedb-provisioner-0 # Index number 0
    - kubedb-kubedb-provisioner-1 # Index number 1
    - kubedb-kubedb-provisioner-2 # Index number 2
```

### Sharding in Action
Each resource in the `kubedb.com` apiGroup gets a label like:  
`shard.operator.k8s.appscode.com/kubedb-provisioner: "X"`  
where `X` is the **index of the pod handling it**.

#### Example 1: `ha-postgres`

```bash
> kubectl get pg -n demo ha-postgres -oyaml
```

```yaml
...
metadata:
  labels:
    shard.operator.k8s.appscode.com/kubedb-provisioner: "0" ## Index number
...
```
→ Managed by `kubedb-kubedb-provisioner-0` because it has index 0 in `.status.controllers[0].pods[0]` <- `kubedb-kubedb-provisioner-0`.

#### Example 2: `tls` Postgres

Lets create another **Postgres** object.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: tls
  namespace: demo
spec:
  version: "17.4"
  replicas: 3
  clientAuthMode: "cert"
  sslMode: "verify-full"
  standbyMode: Hot
  storageType: Durable
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: pg-issuer
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
> Please follow [this](https://kubedb.com/docs/v2025.3.24/guides/postgres/tls/configure/) tutorial to setup a cluster for using tls. Though this is not necessary for this tutorial. Only problem is your `tls` db will be in provisioning state.

```bash
> kubectl apply -f tls.yaml
postgres.kubedb.com/tls created
```

Now get the yaml:

```bash
> kubectl get pg -n demo tls -oyaml
```

```yaml
...
metadata:
  labels:
    shard.operator.k8s.appscode.com/kubedb-provisioner: "1"
...
```
→ Managed by `kubedb-kubedb-provisioner-1` because it has index 1 in `.status.controllers[0].pods[1]` <- `kubedb-kubedb-provisioner-1`.

**Note: we are talking about index from here**
```yaml
status:
  controllers:
  - pods:
    - kubedb-kubedb-provisioner-0 # Index number 0
    - kubedb-kubedb-provisioner-1 # Index number 1
    - kubedb-kubedb-provisioner-2 # Index number 2
```


Lets verify this by checking the logs of `kubedb-kubedb-provisioner-0` & `kubedb-kubedb-provisioner-1` whether they are processing their respective keys.
```bash
➤ kubectl logs -f -n kubedb kubedb-kubedb-provisioner-0
...
I0327 08:30:44.768773       1 postgres.go:392] "validating" logger="postgres-resource" name="ha-postgres" # processing ha-postgres postgres resource of kubedb.com apiGroup

➤ kubectl logs -f -n kubedb kubedb-kubedb-provisioner-1
...
I0327 08:31:46.479722       1 postgres.go:392] "validating" logger="postgres-resource" name="tls" # processing tls postgres resource of kubedb.com apiGroup
I0327 08:31:46.632462       1 reconciler.go:93] Required secrets for postgres: demo/tls are not ready yet
```

Here you can see, `kubedb-kubedb-provisioner-0` is processing `ha-postgres` key and `kubedb-kubedb-provisioner-1` is processing `tls` key.


# **Important Note on Sharding and Scaling**

**Caution:** Scaling the sharded StatefulSet/Deployments/Daemonsets (up or down) may trigger **resharding**, causing resources (like `ha-postgres`, `tls`) to be reassigned to different pods.

### **When to Use Sharding?**
**Stateless workloads** (safe for resharding):
- Example: `kubedb-kubedb-provisioner` (no persistent state).
- If a resource moves to another pod, it continues working without issues.

**Stateful workloads** (avoid sharding):
- Example: `kubedb-kubedb-ops-manager` (must maintain state until tasks complete).
- Resharding could disrupt operations if a resource moves mid-execution.

### **Best Practice**
- Use sharding only for **stateless controllers**.
- Avoid sharding for **stateful controllers** that require stable pod assignments.

Need more details? Let me know at `sourav.cse4.bu@gmail.com`.