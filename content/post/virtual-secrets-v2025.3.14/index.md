---
title: Announcing Virtual Secrets v2024.03.14
date: "2025-03-25"
weight: 14
authors:
- Tapajit Chandra Paul
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

Kubernetes secrets are not really "secret". They are typically stored in plain-text inside etcd. Check out the [Kubernetes Docs](https://kubernetes.io/docs/concepts/configuration/secret/#:~:text=Kubernetes%20Secrets%20are,Kubernetes%20Secrets.) for reference.

This **Virtual Secrets** project offers a new `Secret` which is a Kubernetes Extended API resource that addresses these limitations of the core `Secret`. It supports the full api `Read` / `Write` / `Mount` operation for external/cloud secret management systems without using the k8s secrets rail.

For this initial version, the support for HashiCorp Vault has been added. In the coming release, support for AWS, Azure, and GCP Secret Manager will be integrated.

## Virtual Secrets Design

**Virtual Secrets** introduces a new `Secret` resource under the `virtual-secrets.dev` API group, functioning similarly to the core Kubernetes `Secret` resource from a user point of view. However, instead of storing secret data in `etcd`, it securely stores it in an external secret manager.

The Secret resource is divided into two parts:

- **Secret Data**: Stored in an external secret manager to enhance security.

- **Secret Metadata**: Stored within the cluster, as it is less sensitive and improves performance.

## How to use Virtual Secrets

Let's go ahead and see how to use the `Secret` offered by **Virtual Secrets**.

To keep things isolated, let's create a namespace called `demo` where all the resources will be deployed.

```bash
$ kubectl create namespace demo
namespace/demo created
```

### Install Virtual Secrets Server

First, install the virtual-secret-server which is a custom api server for the `secrets.virtual-secrets.dev` resource.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/virtual-secrets-server --version=v2025.3.14
$ helm upgrade -i virtual-secrets-server appscode/virtual-secrets-server \
    --version=v2025.3.14 -n kubevault --create-namespace 
```

### Deploy and Configure Vault Server

Before we create a custom `Secret`, we need to deploy a vault server where the secret data will be stored. Also, it needs to be configured to grant necessary permissions like create, update, read, list, delete and delete in a kv secret engine named `virtual-secrets.dev` to the virtual-secrets-server.

If you do not have a vault server running. You can use [KubeVault](https://kubevault.com) to deploy a vault server. Here is a [guide](https://kubevault.com/articles/how-to-use-hashicorp-vault-in-kubernetes-using-kubevault/#:~:text=cloud%2Dnative%20applications.-,Deploy%20Vault%20on%20Kubernetes,-Pre%2Drequisites) to do that.

Now let's configure the vault server with following commands:

```bash
# enable kv secret engine in the path virtual-secrets.dev
$ vault secrets enable -path=virtual-secrets.dev -version=2 kv
Success! Enabled the kv secrets engine at: virtual-secrets.dev/


# creates a policy with the permission to create, update, read, list and delete  
$ vault policy write virtual-secrets-policy - <<EOF
path "virtual-secrets.dev/*" {
capabilities = ["create", "update", "read", "list", "delete"]
}
EOF
Success! Uploaded policy: virtual-secrets-policy


# binds this policy with a service account of the virtual-secrets server
$ vault write auth/kubernetes/role/virtual-secrets-role \
    bound_service_account_names=virtual-secrets-server \
    bound_service_account_namespaces=kubevault \
    policies="virtual-secrets-policy"
Success! Data written to: auth/kubernetes/role/virtual-secrets-role
```

### Create SecretStore

We need to create another resource called `SecretStore` which will contain the connection information to the external secret manager where the secrets will be stored. 

```yaml
apiVersion: config.virtual-secrets.dev/v1alpha1
kind: SecretStore
metadata:
  name: vault
spec:
  vault:
    url: http://vault.demo:8200
    roleName: virtual-secrets-role
```

Here,
- `spec.vault` - section describes the connection information for vault.
- `spec.url` - contains the connection url to the vault server.
- `spec.roleName` - contains the role name we specified when binding the policy to the service account earlier.

> **Note**: `spec.aws`, `spec.azure` and `spec.gcp` can be used to specify the connection information of the corresponding secret manager.
 
### Create Virtual Secret

Now, we can create custom `Secret` resource using the YAML below:

```yaml
apiVersion: virtual-secrets.dev/v1alpha1
kind: Secret
metadata:
  name: virtual-secret
  namespace: demo
stringData:
  username: appscode
  password: virtual-secret
secretStoreName: vault
```

Here,
- `secretStoreName` - specifies the `SecretStore` we just created.
- Other than that, everything else is similar to a core Kubernetes `Secret`.

Let's go ahead and apply the `Secret`,

```bash
$ kubectl apply -f virtual-secrets/virtual-secret.yaml
secret.virtual-secrets.dev/virtual-secret created
```

Let's list the `Secrets` to see if it is created or not,

```bash
$ kubectl get secrets.virtual-secrets.dev -n demo
NAME             TYPE     DATA   AGE
virtual-secret   Opaque   2      1m6s
```

We can also get the whole definition of the `Secret`,

```bash
$ kubectl get secrets.virtual-secrets.dev -n demo virtual-secret -oyaml
apiVersion: virtual-secrets.dev/v1alpha1
data:
  password: dmlydHVhbC1zZWNyZXQ=
  username: YXBwc2NvZGU=
kind: Secret
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"virtual-secrets.dev/v1alpha1","kind":"Secret","metadata":{"annotations":{},"name":"virtual-secret","namespace":"demo"},"secretStoreName":"vault","stringData":{"password":"virtual-secret","username":"appscode"}}
  creationTimestamp: "2025-03-25T10:37:45Z"
  generation: 1
  name: virtual-secret
  namespace: demo
  resourceVersion: "332000"
  uid: e2fd622a-020a-4ca4-acbe-501a9147a089
secretStoreName: vault
type: Opaque
```

We can see that this `Secret` actually behaves identical of the core `Secret`. But the data is not stored in the `etcd` and it is way more secure than using the native k8s `Secret`.

Let's check our vault server to see if the secret data exists there or not,

```bash
$ vault kv get virtual-secrets.dev/demo/virtual-secret
================ Secret Path ================
virtual-secrets.dev/data/demo/virtual-secret

======= Metadata =======
Key                Value
---                -----
created_time       2025-03-25T10:37:45.851370497Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

====== Data ======
Key         Value
---         -----
password    virtual-secret
username    appscode
```

We can see that the secret data is stored in the `virtual-secrets.dev/demo/virtual-secret` path where,
- `virtual-secret.dev` is the secret engine name.
- `demo` is the namespace.
- `virtual-secret` is the name of the secret.

### Mount Virtual Secret

`Secrets` are not that useful if we can not mount them to pods. We can mount the `virtual secrets` using [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/).

Virtual Secrets comes with a custom provider of `Secrets Store CSI Driver`, named `secrets-store-csi-driver-provider-virtual-secrets` which leverages `virtual-secrets-server` to read secret data from `virtual secrets` and uses the `Secrets Store CSI Driver` to mount those into to the pods.

Let's go ahead and install `Secrets Store CSI Driver` and `secrets-store-csi-driver-provider-virtual-secrets` into our clusrer,

```bash
$ helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
$ helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system

$ helm search repo appscode/secrets-store-csi-driver-provider-virtual-secrets --version=v2025.3.14
$ helm upgrade -i secrets-store-csi-driver-provider-virtual-secrets appscode/secrets-store-csi-driver-provider-virtual-secrets -n kube-system --create-namespace --version=v2025.3.14
```

If both of them are deployed we should see two new pods in the `kube-system` namespace.

```bash
$ kubectl get pods -n kube-system
NAME                                                      READY   STATUS    RESTARTS      AGE
csi-secrets-store-secrets-store-csi-driver-rvpvm          3/3     Running   0             61s
secrets-store-csi-driver-provider-virtual-secrets-m78gv   1/1     Running   0             34s
```

The `Secrets Store CSI Driver` uses a custom resource named `SecretProviderClass` to mount the secret. Let's go ahead and create that,

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: virtual-secret
  namespace: demo
spec:
  provider: virtual-secrets
  parameters:
    secretName: virtual-secret
```

Here,
- `spec.provider` - specifies the provider for `Secrets Store CSI Driver` to communicate and use.
- `parameters.secretName` - specifies the name of the `virtual secret` we want to mount. 

> **Note**: 
>   - We can also call the `mount` subresource of the `virtual secret` to create the `SecretProviderClass` for us.
>   - The namespace and the name of `SecretProviderClass` should be same as the `Virtual Secret` it is being used for.
Let's create the `SecretProviderClass`,

```bash
$ kubectl apply -f virtual-secrets/secret-provider-class.yaml 
secretproviderclass.secrets-store.csi.x-k8s.io/virtual-secret created
```

With the custom secret created, the authentication configured and role created, the `provider-virtual-secrets` extension installed and the `SecretProviderClass` defined it is finally time to create a pod that mounts the desired secret.

```bash
kind: Pod
apiVersion: v1
metadata:
  name: webapp
  namespace: demo
spec:
  containers:
    - image: jweissig/app:0.0.1
      name: webapp
      volumeMounts:
        - name: virtual-secrets-store
          mountPath: "/mnt/virtual-secrets"
          readOnly: true
  volumes:
    - name: virtual-secrets-store
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "virtual-secret"
```

Here,
- In `spec.volumes[0]`, a volume with name `virtual-secrets-store` with necessary configs is specified.
- In `spec.containers[0].volumeMounts`, the volume is referred to be mounted in the `/mnt/virtual-secrets` path.

Let's create the pod,

```bash
$ kubectl apply -f virtual-secrets/app.yaml
pod/webapp created
```

If we get the pod we will see that it will get to the `Running` state after some period,

```bash
$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
webapp              1/1     Running   0          6m45s
```

Now, check the secret data written to the file system at /mnt/virtual-secrets on the webapp pod.

```bash
$ kubectl exec -n demo webapp -- cat /mnt/virtual-secrets/username
appscode⏎

$ kubectl exec -n demo webapp -- cat /mnt/virtual-secrets/password
virtual-secret⏎ 
```

The value displayed matches the username and password value for the custom secret named `virtual-secret` we created earlier.


### Use Virtual Secrets with KubeDB

**Virtual Secrets** is integrated with KubeDB from the `v2025.3.24` and it can be used to store KubeDB's database credential. Initially, the support has been added for `PostgreSQL`.

Make sure that the following things are configured in your cluster before trying out `Virtual Secrets` with `KubeDB`:
- **Virtual Secrets Server** is installed. You can find the guide to install it [HERE](#install-virtual-secrets-server).
- **Secrets Store CSI Driver** is installed. You can find the guide to install it [HERE](#mount-virtual-secret).
- **Provider Virtual Secret** is installed. You can find the guide to install it [HERE](#mount-virtual-secret).
- **[KubeDB](https://kubedb.com)** is installed. Make sure that the version is **v2025.3.24** or latest.
- Have a vault server deployed and configured for the virtual-secrets-server to use. You can find the guide to do it in [HERE](#Deploy-and-Configure-Vault-Server).
- Create a `SecretStore` from [HERE](#create-secretstore).


If all of these are installed, we can proceed with deploying a `postgres` which will use `virtual-secrets` to create custom secret for the database authentication credential.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-demo
  namespace: demo
spec:
  authSecret:
    apiGroup: "virtual-secrets.dev"
    secretStoreName: vault
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  deletionPolicy: Delete
  version: "13.13"
```

Here,
- `spec.authSecret.apiGroup` - specifies that we want to use virtual secrets instead of native k8s secret.
- `spec.authSecret.secretStoreName` - specifies the `SecretStore` resource that contains the connection information for external secret store to store the secret data.

> To know more about KubeDB postgres. Check out this [Quick Start](https://kubedb.com/docs/v2025.3.24/guides/postgres/quickstart/quickstart/#:~:text=quick%2Dpostgres%20created-,Here%2C,a%20user%20from%20deleting%20this%20object%20if%20admission%20webhook%20is%20enabled.,-Note%3A%20spec)

We can now apply the `postgres`,

```bash
$ kubectl apply -f virtual-secrets/postgres.yaml
postgres.kubedb.com/pg-demo created
```

Let's wait for some time for the postgres database to become `Ready`,

```bash
$ kubectl get pg -n demo -w
NAME      VERSION   STATUS         AGE
pg-demo   13.13     Provisioning   17s
pg-demo   13.13     Provisioning   24s
pg-demo   13.13     Provisioning   35s
pg-demo   13.13     Ready          41s
```

Now, lets go ahead and check what secret it is using,

```bash
$ kubectl get secrets.virtual-secrets.dev -n demo 
NAME             TYPE                       DATA   AGE
pg-demo-auth     kubernetes.io/basic-auth   2      1m53s
```

We can see that a `virtual-secret` named `pg-demo-auth` has been created by the KubeDB operator. Let's get the whole definition of the virtual secret,

```bash
$ kubectl get secrets.virtual-secrets.dev -n demo pg-demo-auth -oyaml
apiVersion: virtual-secrets.dev/v1alpha1
data:
  password: MEVUZnZfTikyUjA4dShFNg==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  annotations:
    kubedb.com/auth-active-from: "2025-03-25T12:27:59Z"
  creationTimestamp: "2025-03-25T12:27:59Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: pg-demo
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: pg-demo-auth
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: Postgres
    name: pg-demo
    uid: dba22641-75cd-40ba-acf8-f66894b9559e
  resourceVersion: "335961"
  uid: 40b432b3-b5af-4330-b8ae-4cc1a5b37f77
secretStoreName: vault
type: kubernetes.io/basic-auth
```

In our vault server, we can check if this data exists or not,

```bash
$ vault kv get virtual-secrets.dev/demo/pg-demo-auth
=============== Secret Path ===============
virtual-secrets.dev/data/demo/pg-demo-auth

======= Metadata =======
Key                Value
---                -----
created_time       2025-03-25T12:27:59.772594581Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            9

====== Data ======
Key         Value
---         -----
password    0ETfv_N)2R08u(E6
username    postgres
```

This is how you can use `virtual secret` to keep your sensitive data secure in an external secret manager. Also, use it with KubeDB seamlessly. 

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter/X](https://x.com/AppsCodeHQ).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://youtube.com/@appscode) channel.

If you have found a bug with Virtual Secrets or want to request for new features, please [file an issue](https://github.com/virtual-secrets/project/issues/new).