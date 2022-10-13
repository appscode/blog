---
title: Introducing KubeVault v2022.09.09
date: 2022-09-09
weight: 25
authors:
- Sakib Alamin
tags:
- kubevault
- ops-requests
- recommendation
- cli
- kubevault cli
- kubernetes
- secret-management
- security
- vault
- hashicorp
- enterprise
- community
---

We are very excited to announce the release of KubeVault v2022.09.09 Edition. 
The KubeVault `v2022.09.09` contains numerous improvements from `KubeVault Operator` & `CLI` end. 
It includes features for managing Day-2 lifecycle of Vault including Vault `Ops-request`, `Recommendation` generation for managing Vault TLS, updated `Health Checker`, `KubeVault CLI`, support for `Pod Disruption Budget`, etc. for Vault Cluster. 

- [Install KubeVault](https://kubevault.com/docs/v2022.09.09/setup/)

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

In this post, we are going to highlight the major changes. You can find the complete commit by commit changelog [here](https://github.com/kubevault/CHANGELOG/blob/master/releases/v2022.09.09/README.md).

## Vault Ops-requests

Managing Day-2 lifecycle of `Vault` is now even easier with newly added Vault `Ops-requests`. This is convenient for managing/reconfiguring TLS for your Vault deployment. Now, it's very easy to add, remove, reconfigure, rotate, etc. Vault TLS at your will. 

Here's a simple example of type `ReconfigureTLS` that reconfigures `server` certificates with the provided fields.

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    certificates:
    - alias: server
      subject:
        organizations:
        - appscode:kubevault
      emailAddresses:
      - "sakibalamin@appscode.com"
```

## KubeVault Recommendation Engine

Recommendation Engine generates recommendation to automate the day-2 life cycle of Kubernetes objects. 
KubeVault Recommendation Engine is a part of KubeVault operator which will run inside the KubeVault operator pod. 
It watches the Vault Server custom resources and generates recommendation based on the Vault Server state. 
Currently, Recommendation engine generates Rotate TLS recommendation for TLS secured Vault Server.

In this release, we have introduced the KubeVault Recommendation Engine to generate recommendations for vault resources.
Currently, it will generate **Rotate TLS** ops-request recommendations depending on the TLS certificate expiry date. We have already introduced Supervisor to execute the recommendation in a user-defined maintenance window. To install the Supervisor helm chart, please visit [here](https://github.com/kubeops/installer/tree/master/charts/supervisor).

To see the generated recommendations:

```bash
$ kubectl get recommendations.supervisor.appscode.com -A
```

### Recommendation: Rotate TLS

By default, the recommendation engine will generate a Rotate TLS recommendation for the TLS secured vault before one month of the TLS certificate expiry date. But if the TLS certificate lifespan is less than one month, then it will generate the recommendation before half of the TLS certificate lifespan. Also while installing KubeVault, users can specify custom flags to configure the recommendation-engine to create Rotate TLS recommendations before a specific time of TLS certificate expiry date.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  labels:
    app.kubernetes.io/instance: vault
    app.kubernetes.io/managed-by: kubevault.com
    rotate-tls: rotate-tls
  name: vault-x-vaultserver-x-rotate-tls-tthnu9
  namespace: demo
spec:
  backoffLimit: 5
  description: TLS Certificate is going to be expire on 2022-12-04 05:34:37 +0000 UTC
  operation:
    apiVersion: ops.kubevault.com/v1alpha1
    kind: VaultOpsRequest
    metadata:
      name: rotate-tls-otxghu
      namespace: demo
    spec:
      tls:
        rotateCertificates: true
      type: ReconfigureTLS
      vaultRef:
        name: vault
    status: {}
  recommender:
    name: vault-operator
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubevault.com/v1alpha2
    kind: VaultServer
    name: vault
```

```bash
$ helm install kubevault appscode/kubevault \
    --version v2022.09.09 \
    --namespace kubevault --create-namespace \
    --set kubevault-ops-manager.recommendationEngine.genRotateTLSRecommendationBeforeExpiryMonth=2 \
    --set-file global.license=/path/to/the/license.txt
    
```

With the above installation, the recommendation engine will generate recommendations before two months of the TLS certificate expiry date. To know more about the recommendation engine flags please visit [here](https://github.com/kubevault/installer/tree/master/charts/kubevault-operator).

## Vault Health Checker

In this release, we've improved KubeVault health checks. We've added a new field called `healthChecker` under spec. It controls the behavior of health checks. It has the following fields:

- `spec.healthChecker.periodSeconds` : Specifies the interval between each health check iteration.
- `spec.healthChecker.timeoutSeconds` : Specifies the timeout for each health check iteration.
- `spec.healthChecker.failureThreshold` : Specifies the number of consecutive failures to mark the Vault as NotReady.
- `spec.healthChecker.disableWriteCheck` : KubeVault does write check by default, but if you want to disable it, you can use this field.

Example YAML:

```yaml
spec:
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
    disableWriteCheck: true
```

## Pod Disruption Budget

A Pod Disruption Budget (PDB) allows you to limit the voluntary disruption to your application when its pods need to be rescheduled for some reason such as upgrades or routine maintenance work on the Kubernetes nodes.

Now, a `PDB` will also be created along with your `Vault` deployment & it'll set a `maxUnavailable` pods value, so an eviction will be allowed if at most `maxUnavailable` pods selected by `selector` are unavailable after the eviction.

## Rotate, Generate Vault root-token

`KubeVault CLI` always focuses on improving user experience by automating tedious tasks associated with `Vault`. This release is no exception. 

Now, you can `Generate` & `Rotate` the Vault `root-token` using the `KubeVault CLI`. 

`kubectl vault generate` command will generate a completely new `root-token` using the `unseal-keys`. At least the threshold number of `unseal-keys` must be present for this operation to be successful.

```bash
# this will generate a new root-token using the available unseal-keys
# at least the threshold number of keys must be present for this to be successful
$ kubectl vault root-token generate vaultserver -n demo vault

root-token generation successful
generated root-token: hvs.oxEoWs355PYgfs1mv73cxOsc
```

`kubectl vault rotate` command will rotate the `root-token` using the `unseal-keys`. The `root-token` & at least the threshold number of `unseal-keys` must be present for this operation to be successful.

```bash
# this will rotate the root-token using the available unseal-keys
# old root-token privileges will be revoked after successfully rotating the token
$ kubectl vault root-token rotate vaultserver -n demo vault

root-token generation successful
root-token rotation successful
```

## What's Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2022.09.09/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
