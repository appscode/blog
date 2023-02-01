---
title: Leveraging KubeVault to Manage the Redis Secret Engine in Amazon EKS - A Step-by-Step Guide
date: "2023-01-27"
weight: 14
authors:
- Dipta Roy
tags:
- amazon
- aws
- cloud-native
- database
- dbaas
- eks
- kubedb
- kubernetes
- kubevault
- redis
- s3
- secrets
---

## Overview

KubeVault is a Git-Ops ready, production-grade solution for deploying and configuring Hashicorp's Vault on Kubernetes. KubeVault provides various ways to configure your Vault deployment. You can pick your preferred Storage Backend, Unsealer Mode, TLS Mode, Secret Engines that you want to allow to attach with this VaultServer, Termination Policy to prevent accidental deletion or clean-up Vault deployment in a systematic way, Monitoring, etc. You can find the guides and detailed information in [KubeVault Documentation](https://kubevault.com/docs/latest/welcome/).
In this tutorial, we will show how to Manage Redis Secret Engine using KubeVault in Amazon Elastic Kubernetes Service (Amazon EKS). We will cover the following steps:

1) Install KubeDB
2) Install KubeVault
3) Deploy Vault Server
3) Deploy Redis Standalone Database
5) Manage User Privileges

### Get Cluster ID

We need the cluster ID to get the KubeDB and KubeVault License.
To get cluster ID we can run the following command:

```bash
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
8c4498337-358b-4dc0-be52-14440f4e061e
```

### Get License

Go to [Appscode License Server](https://license-issuer.appscode.com/) to get the license.txt file. For this tutorial we will use KubeDB and KubeVault Enterprise Edition.

![License Server](AppscodeLicense.png)

### Install KubeDB

We will use helm to install KubeDB. Please install helm [here](https://helm.sh/docs/intro/install/) if it is not already installed.
Now, let's install `KubeDB`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

# Install KubeDB Enterprise operator chart
$ helm install kubedb appscode/kubedb \
  --version v2023.01.17 \
  --namespace kubedb --create-namespace \
  --set kubedb-provisioner.enabled=true \
  --set kubedb-ops-manager.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-dashboard.enabled=true \
  --set kubedb-schema-manager.enabled=true \
  --set-file global.license=/path/to/the/license.txt
```

Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-54fcfbcdb7-28wc7       1/1     Running   0          2m22s
kubedb      kubedb-kubedb-dashboard-599fdd8d8f-tnj6l        1/1     Running   0          2m22s
kubedb      kubedb-kubedb-ops-manager-868d684c84-7v9dz      1/1     Running   0          2m22s
kubedb      kubedb-kubedb-provisioner-75b544fdf4-w6z8l      1/1     Running   0          2m22s
kubedb      kubedb-kubedb-schema-manager-66d686b986-kskfj   1/1     Running   0          2m22s
kubedb      kubedb-kubedb-webhook-server-677b88d9bb-29qln   1/1     Running   0          2m22s
```

## Install KubeVault Enterprise Operator

```bash
$ helm install kubevault appscode/kubevault \
  --version v2022.12.28 \
  --namespace kubevault --create-namespace \
  --set-file global.license=/path/to/the/license.txt
```
Let's verify the installation:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubevault"

NAMESPACE   NAME                                                  READY   STATUS    RESTARTS   AGE
kubevault   kubevault-kubevault-operator-86b8c7f688-6ln65         1/1     Running   0          22s
kubevault   kubevault-kubevault-webhook-server-8554c7cd7f-j9w6q   1/1     Running   0          22s
```
### Install KubeVault CLI

KubeVault provides a `kubectl` plugin to interact with KubeVault resources. Let's install [KubeVault CLI](https://kubevault.com/docs/latest/setup/install/kubectl_plugin/)

### Install Secret-store CSI Driver

```bash
$ helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
$ helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system
```
### Install Vault specific CSI Provider

```bash
$ helm repo add hashicorp https://helm.releases.hashicorp.com
$ helm install vault hashicorp/vault \
    --set "server.enabled=false" \
    --set "injector.enabled=false" \
    --set "csi.enabled=true"
```

### Create Namespace and Secret

To keep everything isolated, we are going to use a separate namespace `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created
```

We need to create a storage secret for our backend. here, we are using Amazon EKS.

```bash
$ echo -n '<your-secret-key-id-here>' > secret_key
$ echo -n '<your-secret-access-key-here>' > access_key
$ kubectl create secret generic -n demo aws-secret \
    --from-file=./secret_key \
    --from-file=./access_key
secret/aws-secret created
```
Also, We have created an `S3` bucket in Amazon to use this as our Vault Backend. We have created a bucket named `vault-demo-1` in `us-east-1` region. KubeVault supports various storage backends, you can find the details in [Storage Backend](https://kubevault.com/docs/latest/concepts/vault-server-crds/storage/overview/) 

## Deploy VaultServer

Now, we are going to deploy the `VaultServer`.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 1
  version: 1.12.1
  allowedSecretEngines:
    namespaces:
      from: All
    secretEngines:
      - redis
  backend:
    s3:
      bucket: "vault-demo-1"
      region: "us-east-1"
      endpoint: s3.amazonaws.com
      credentialSecretRef:
        name: aws-secret
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```
In this yaml, 
- `spec.replicas` specifies the number of Vault nodes to deploy. It has to be a positive number. Note: Amazon EKS does not support HA for Vault. As we using Amazon EKS as our backend it has to be 1.
- `spec.version` specifies the name of the `VaultServerVersion` CRD. This CRD holds the image name and version of the Vault, Unsealer, and Exporter.
- `spec.allowedSecretEngines` defines the Secret Engine informations which to be granted in this Vault Server.
- `spec.backend` is a required field that contains the Vault backend storage configuration.
- `spec.unsealer` specifies `Unsealer` configuration. `Unsealer` handles automatic initializing and unsealing of Vault.
- `spec.terminationPolicy` field is *Wipeout* means that vault will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubevault.com/docs/latest/concepts/vault-server-crds/vaultserver/#specterminationpolicy).

Let’s save this yaml configuration into `vault.yaml` and deploy it,

```bash
$ kubectl apply -f vault.yaml
vaultserver.kubevault.com/vault created
```

Once all of the above things are handled correctly then you will see that the following objects are created:

```bash
$ kubectl get all -n demo
NAME          READY   STATUS    RESTARTS   AGE
pod/vault-0   2/2     Running   0          64s

NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
service/vault            ClusterIP   10.100.41.92   <none>        8200/TCP,8201/TCP   67s
service/vault-internal   ClusterIP   None           <none>        8200/TCP,8201/TCP   67s

NAME                     READY   AGE
statefulset.apps/vault   1/1     72s

NAME                                       TYPE   VERSION   AGE
appbinding.appcatalog.appscode.com/vault                    75s

NAME                              REPLICAS   VERSION   STATUS   AGE
vaultserver.kubevault.com/vault   1          1.12.1    Ready    91s

NAME                                                                   STATUS    AGE
vaultpolicybinding.policy.kubevault.com/vault-auth-method-controller   Success   61s

NAME                                                            STATUS    AGE
vaultpolicy.policy.kubevault.com/vault-auth-method-controller   Success   64s
```

### Use KubeVault CLI

We will connect to the Vault by using `KubeVault CLI`. Therefore, we need to export the necessary environment varibles and port-forward the service. 

```bash
$ export VAULT_ADDR=http://127.0.0.1:8200
$ export VAULT_TOKEN=(kubectl vault root-token get vaultserver vault -n demo --value-only)
```
Now, Let's port-forward the service and interact via CLI,

```bash
$ kubectl port-forward -n demo service/vault 8200
Forwarding from 127.0.0.1:8200 -> 8200
Forwarding from [::1]:8200 -> 8200

##Check Vault Status
$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    5
Threshold       3
Version         1.12.1
Build Date      2022-10-27T12:32:05Z
Storage Type    s3
Cluster Name    vault-cluster-3bd6b372
Cluster ID      15df69fb-e717-9af1-8d00-0f4cc9df97d4
HA Enabled      false
```

## Deploy Redis Standalone Database

Now, we are going to Install Redis with the help of KubeDB.
Here is the yaml of the Redis CRD we are going to use:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-standalone       
  namespace: demo
spec:
  version: 7.0.6
  storageType: Durable
  storage:
    storageClassName: "gp2"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```
Let's save this yaml configuration into `redis-standalone.yaml` 
Then create the above Redis CRD

```bash
$ kubectl apply -f redis-standalone.yaml 
redis.kubedb.com/redis-standalone created
```
In this yaml,
* `spec.version` field specifies the version of Redis. Here, we are using Redis `version 7.0.6`. You can list the KubeDB supported versions of Redis by running `$ kubectl get redisversions` command.
* Another field to notice is the `spec.storageType` field. This can be `Durable` or `Ephemeral` depending on the requirements of the database to be persistent or not.
* `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any `StorageClass` available in your cluster with appropriate resource requests.
* Lastly, the `spec.terminationPolicy` field is *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/redis/concepts/redis/#specterminationpolicy).

Once these are handled correctly and the Redis object is deployed, you will see that the following are created for `redis-standalone`:

```bash
$ kubectl get all -n demo -l 'app.kubernetes.io/instance=redis-standalone'
NAME                     READY   STATUS    RESTARTS   AGE
pod/redis-standalone-0   1/1     Running   0          2m19s

NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/redis-standalone        ClusterIP   10.100.27.224   <none>        6379/TCP   2m21s
service/redis-standalone-pods   ClusterIP   None            <none>        6379/TCP   2m21s

NAME                                READY   AGE
statefulset.apps/redis-standalone   1/1     2m25s

NAME                                                  TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/redis-standalone   kubedb.com/redis   7.0.6     2m32s

```
Let’s check if the database is ready to use,

```bash
$ kubectl get redis -n demo
NAME               VERSION   STATUS   AGE
redis-standalone   7.0.6     Ready    5m54s
```

## Create Redis SecretEngine

Here, we are going to create a `Redis SecretEngine`. Secret engines are components that store, generate, or encrypt data. Secret engines are provided some set of data, they take some action on that data, and they return a result.


```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: redis-secret-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
    namespace: demo
  redis:
    databaseRef:
      name: redis-standalone
      namespace: demo
    pluginName: "redis-database-plugin"
```
In this yaml,
- `spec.vaultRef` is a required field that specifies an `AppBinding` reference which is used to connect with a Vault server.
- `spec.redis` specifies the configuration required to configure Redis database secret engine.

Let’s save this yaml configuration into `redis-secret-engine.yaml` and deploy it,

```bash
$ kubectl apply -f redis-secret-engine.yaml
secretengine.engine.kubevault.com/redis-secret-engine created
```
Let's check the `Secrets` list,

```bash
$ vault secrets list
Path                                                                        Type         Accessor              Description
----                                                                        ----         --------              -----------
cubbyhole/                                                                  cubbyhole    cubbyhole_a8440898    per-token private secret storage
identity/                                                                   identity     identity_e2462dcc     identity store
k8s.2903de3f-2693-44e3-b50d-aad10b403c1e.kv.demo.vault-health/              kv           kv_bab72ce4           n/a
k8s.2903de3f-2693-44e3-b50d-aad10b403c1e.redis.demo.redis-secret-engine/    database     database_c8d1044e     n/a
sys/                                                                        system       system_abe38f09       system endpoints used for control, policy and debugging
```

### Create Database Roles

Now, we are going to create a Redis database secret engine role to specify permissions to the user.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: RedisRole
metadata:
  name: write-read-role
  namespace: demo
spec:
  secretEngineRef:
    name: redis-secret-engine
  creationStatements:
    - '["~*", "+@read","+@write"]'
  defaultTTL: 1h
  maxTTL: 24h
```
In this yaml,
- `spec.secretEngineRef` is a required field that specifies the name of a `SecretEngine`.
- `spec.creationStatements` is a required field that specifies a list of database statements executed to create and configure a user.
- `spec.defaultTTL` is an optional field that specifies the TTL for the leases associated with this role. Accepts time suffixed strings (“1h”) or an integer number of seconds. Defaults to system/engine default TTL time.
- `spec.maxTTL` is an optional field that specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings (“1h”) or an integer number of seconds. Defaults to system/engine default TTL time.

Let’s save this yaml configuration into `redis-secret-engine.yaml` and apply it,

```bash
$ kubectl apply -f write-read-role.yaml 
redisrole.engine.kubevault.com/write-read-role created
```
Let's verify the `redisrole` status,

```bash
$ kubectl get redisrole -n demo
NAME              STATUS    AGE
write-read-role   Success   2m
```

### Create Service Account

To create a `SecretAccessRequest`, we need to provide a `ServiceAccount` for that. Let's create a `ServiceAccount`,

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: write-read-svc-a
  namespace: demo
```

Let’s save this yaml configuration into `write-read-svc-a.yaml` and apply it,

```bash
$ kubectl apply -f write-read-svc-a.yaml
serviceaccount/write-read-svc-a created
```

### Create SecretAccessRequest

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: write-read-secret-access
  namespace: demo
spec:
  roleRef:
    kind: RedisRole
    name: write-read-role
  subjects:
    - kind: ServiceAccount
      name: write-read-svc-a
      namespace: demo
```
In this yaml,
- `spec.roleRef` is a required field that specifies the role against which credentials will be issued.
- `spec.subjects` is a required field that contains a list of references to the object or user identities on whose behalf this request is made. These object or user identities will have read access to the k8s credential secret. This can either hold a direct API object reference or a value for non-objects such as user and group names.

Let’s save this yaml configuration into `write-read-secret-access.yaml` and apply it,

```bash
$ kubectl apply -f write-read-secret-access.yaml
secretaccessrequest.engine.kubevault.com/write-read-secret-access created
```

Let's check the `secretaccessrequest` status,

```bash
$ kubectl get secretaccessrequest -n demo
NAME                       STATUS               AGE
read-secret-access         WaitingForApproval   2m
write-read-secret-access   WaitingForApproval   2m
```
> Note: A `SecretAccessRequest` is a Kubernetes CustomResourceDefinition (CRD) which allows a user to request a Vault server for credentials in a Kubernetes native way. A `SecretAccessRequest` can be created under various `roleRef` e.g: `AWSRole`, `GCPRole`, `ElasticsearchRole`, `MongoDBRole`, etc. A `SecretAccessRequest` has three different phases e.g: `WaitingForApproval`, `Approved`, `Denied`. If database admin approve the `SecretAccessRequest` then the KubeVault operator will issue credentials and create Kubernetes secret containing credentials. 

Let's approve the `SecretAccessRequest`,

```bash
$ kubectl vault approve secretaccessrequest write-read-secret-access -n demo
secretaccessrequests write-read-secret-access approved
```

Let's verify the `secretaccessrequest` status again,

```bash
$ kubectl get secretaccessrequest -n demo
NAME                       STATUS     AGE
write-read-secret-access   Approved   5m
```
Now, let's get the username and password which is generated for `write-read-secret-access`,

```bash
$ kubectl get secret -n demo
NAME                                   TYPE                                  DATA   AGE
aws-secret                             Opaque                                2      5h
default-token-nrwj8                    kubernetes.io/service-account-token   3      5h
redis-standalone-auth                  kubernetes.io/basic-auth              2      4h11m
redis-standalone-config                Opaque                                1      4h11m
redis-standalone-token-lxj66           kubernetes.io/service-account-token   3      4h11m
vault-k8s-token-reviewer-token-flfzg   kubernetes.io/service-account-token   3      4h59m
vault-keys                             Opaque                                6      4h58m
vault-token-wwszh                      kubernetes.io/service-account-token   3      4h59m
vault-vault-config                     Opaque                                1      4h59m
write-read-secret-access-vhixwg        Opaque                                2      3m53s
write-read-svc-a-token-bbwlt           kubernetes.io/service-account-token   3      68m
```
> You can seperate the `SecretAccessRequest` generated secret by finding secrets named with `SecretAccessRequest-name-{suffix}`, in our case `write-read-secret-access-vhixwg` is the `SecretAccessRequest` generated secret.

```bash
$ kubectl view-secret -n demo write-read-secret-access-vhixwg -a
password=u-A2uMus21q34jtl9xW8
username=V_KUBERNETES-DEMO-VAULT_K8S.2903DE3F-2693-44E3-B50D-AAD10B403C1E.DEMO.WRITE-READ-ROLE_NBECFPMIUC5EU1
```

## Accessing Database Through CLI using Credentials

In this section, we are going to login into our Redis database pod using the credentials and read-write some sample data.

```bash
$ kubectl exec -it -n demo redis-standalone-0 -- bash
root@redis-standalone-0:/data# redis-cli
127.0.0.1:6379> auth V_KUBERNETES-DEMO-VAULT_K8S.2903DE3F-2693-44E3-B50D-AAD10B403C1E.DEMO.WRITE-READ-ROLE_NBECFPMIUC5EU1 u-A2uMus21q34jtl9xW8
OK
127.0.0.1:6379> SET Product KubeVault
OK
127.0.0.1:6379> get Product
"KubeVault"
```
Now, we will try to run some DB admin command,

```bash
$ kubectl exec -it -n demo redis-standalone-0 -- bash
root@redis-standalone-0:/data# redis-cli
127.0.0.1:6379> auth V_KUBERNETES-DEMO-VAULT_K8S.2903DE3F-2693-44E3-B50D-AAD10B403C1E.DEMO.WRITE-READ-ROLE_NBECFPMIUC5EU1 u-A2uMus21q34jtl9xW8
OK
127.0.0.1:6379> acl list
(error) NOPERM this user has no permissions to run the 'acl|list' command
127.0.0.1:6379> exit
```
As we can see this user does not have the permission to run that command.

### Revoke the SecretAccessRequest

Now, we are going to revoke our `SecretAccessRequest` so that no one can access the database with the previously generated credentials. This option is often used to prevent any kind of data breach.

```bash
$ kubectl vault revoke secretaccessrequest write-read-secret-access -n demo
secretaccessrequests write-read-secret-access revoked
```

```bash
$ kubectl get secretaccessrequest -n demo
NAME                       STATUS    AGE
write-read-secret-access   Expired   45m
```


We have made an in depth video on Manage Redis Secrets using KubeVault along with KubeDB. You can have a look into the video below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/tCcb1ZyXhxc" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements of KubeVault, follow us on [Twitter](https://twitter.com/KubeVault).

To receive product announcements of KubeDB, follow us on [Twitter](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.

More about [Redis in Kubernetes](https://kubedb.com/kubernetes/databases/run-and-manage-redis-on-kubernetes/)

Go through the concepts of [KubeVault](https://kubevault.com/)

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).