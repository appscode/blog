---
author: Tamal Saha
date: 2018-02-02
linktitle: The case for AppBinding
title: The case for Test post
weight: 10
authorAvatar: tamal.png
tags:
  - kubernetes
  - crd
  - service-catalog
---

Kubernetes has become the de-facto orchestrator for the cloud native world. Kubernetes upholds the philosophy that the core should be small and allow developers to write their own extensions. One way to introduce new resource types is using `CustomResourceDefintions (CRD)` (originally known as `ThirdPartyResources`). Using CRDs anyone can define a new resource type that behaves like standard Kubernetes resources. This allows anyone to write a controller for custom resources and capture operational knowledge in a software form. CoreOS popularized the term "[operators](https://coreos.com/blog/introducing-operators.html)" as a name for this pattern.

At [AppsCode](https://twitter.com/AppsCodeHQ) , we have used this model to build various Kubernetes native applications. For example, we have been working on a project called [KubeDB](https://twitter.com/KubeDB) that automates the management of databases on Kubernetes. This is kind of like AWS RDS but using containers running on Kubernetes. For example, to deploy a PostgreSQL database you can use a yaml definition like below:

{{< highlight yaml >}}
apiVersion: kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
spec:
  version: "10.2-v1"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate
{{< /highlight >}}

## Connecting Operators

As time goes on, more and more functionality of running the day to day operations on Kubernetes is captured via CRDs and their associated operators. For example, our users have been [asking](https://github.com/kubedb/project/issues/141) us to support user management for KubeDB managed databases using [HashiCorp Vault](https://www.vaultproject.io/). Vault is a secret management server for various types of secrets including Database secrets. We have been working on a project called [KubeVault](https://twitter.com/KubeVault) to bring a Kubernetes native experience for secret management via Vault. Now, one of the challenges we face is how to connect CRDs from KubeDB project with CRDs from KubeVault project. Here are the requirements:

- Users should not be forced to use both projects. They may choose to use one of the projects for their intended purpose. But they should be able to leverage them both for additional functionality.
- Users should be able to bring in databases managed outside a Kubernetes cluster and use the KubeVault operator to issue secrets for such databases. For example, a user may be using an AWS RDS managed database. They want to use the KubeVault project to manage secrets for such databases. Or, users may want to provision databases using Helm charts but use KubeVault for user management.

These requirements mean that we can’t tightly couple CRDs from different projects. Rather we need an intermediate resource that can connect services managed by different tools.

## Connection Configuration

The key piece of information needed to connect different services are usually an URL for the application, some secret credential and/or some configuration parameters. So, we considered using a standard Kubernetes `Service` object. But that will mean that we have to use annotations to point to additional secrets and configmaps for such a Service. We would like to avoid using annotations for specification. Also, if such a service is hosted outside a Kubernetes cluster, a Kubernetes `Service` object seems like the wrong place to store connection information.

## Service Catalog

We looked around to see if this problem has been addressed by others in the community. The closed project we found is [Service Catalog](https://github.com/kubernetes-incubator/service-catalog). Service Catalog lets anyone provision cloud services directly from the comfort of native Kubernetes tooling. This project brings integration with service brokers to the Kubernetes ecosystem via the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker). Service catalog project has a resource type called [`ServiceBinding`](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/v1/api.md#servicebinding-resource) that can be used to represent connection information for a resource managed via service broker. But it has 2 major issues:

- Service catalog expects resources to be managed by a service broker. But in our case, the resource may be already present and not managed by any broker. Going back to our prior example, a user may already have an AWS RDS managed PostgreSQL database. They just want to manage its secrets using KubeVault.
- Service catalog project is implemented as a Kubernetes extended apiserver. This means users of a service catalog has to deploy and manage an additional etcd cluster. This introduced significant operational complexity for users. The friction of depending on an extended apiserver is unacceptable for an operator like KubeVault.

## AppBinding
Hence, we came up with the concept of `AppBinding`. AppBinding points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if hosted inside a Kubernetes cluster), some optional parameters and a credential secret. AppBinding is pretty similar to ServiceBinding concept in Service catalog with the following key differences:

- AppBinding is a CRD. So, any project that wants to depend on the AppBinding concept can register the CRD type and start using it. There is zero operational overload of using an AppBindign resource.
- An AppBinding can refer to an application via its URL or service reference (if hosted inside a Kubernetes cluster). There is no need to come up with service classes, service plans, etc. like in the case of a service catalog.

Here is an example of an AppBinding that points to a PostgreSQL instance.

{{< highlight yaml >}}
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: postgres-app
  namespace: demo
spec:
  secret:
    name: postgres-user-cred # secret
  clientConfig:
    service:
      name: postgres
      scheme: postgresql
      port: 5432
      path: "postgres"
      query: "sslmode=disable"
    insecureSkipTLSVerify: true
  parameters:
    # names of the allowed roles to use this connection config in Vault
    allowedRoles: "*"
{{< /highlight >}}

Now, KubeVault operator can refer to this database object from it’s own CRD and issue secrets accordingly.

You can find the type definition and auto-generated Go client for AppBinding on [GitHub](https://github.com/kmodules/custom-resources/blob/master/apis/appcatalog/v1alpha1/appbinding_types.go).

If you have read all the way to the end, I want to thank you. If you have any questions and want to know more, you can reach me via [Twitter](https://twitter.com/tsaha) or [Email](mailto:tamal@appscode.com).
