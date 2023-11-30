---
title: Announcing Voyager Gateway v0.6.1
date: "2023-11-30"
weight: 14
authors:
- Tasdidur Rahman
tags:
- cloud-native
- gateway
- kubernetes
- voyager-gateway
---

We are pleased to announce the release of **Voyager Gateway v0.6.1**. This post lists all the major changes done in this release since the last release.

### Custom Database Routes

Voyager Gateway is built on the Envoyproxy gateway project which uses Envoyproxy under the hood. Envoyproxy supports protocol specific filters for databases like MySQL, PostgreSQL, MongoDB, Kafka and Redis. But k8s gateway api does not have a way to specify the database filters. We can expose databases via TCP connections but this restricts full potential of the database filters in Envoyproxy.
In Voyager Gateway we introduced custom routes for database backends. When a database server is exposed through custom database routes, it patches the filters, specifically developed for that particular protocol. With this we can generate stats, establish TLS secure connections and use the full potential of Envoyproxy.

```yaml
apiVersion: gateway.voyagermesh.com/v1alpha1
kind: MySQLRoute
metadata:
  name: my-route
  namespace: my-ns
spec:
  parentRefs:
  - name: my-gw
    sectionName: my-lis
  rules:
  - backendRefs:
    - name: mysql
      port: 3306
```

### TLS Termination on Gateway
In this release, we have added support to the database filters to terminate TLS on the gateway. This includes MySQL, PostgreSQL and MongoDB. We can leverage this feature to include different TLS certificates on the client-to-gateway and gateway-to-upstream. This enhances the security and decouples dependencies for cluster administrators and developers.


### Implement BackendTLSPolicy

In this release, we have implemented BackendTLSPolicy for supporting TLS secure connections for the upstream. BackendTLSPolicy is being proposed by the official Gateway-API project recently and Voyager Gateway has adapted the necessary changes to add support for it. We can now configure MySQL, PostgreSQL, MongoDB and HTTPS upstreams for TLS secured connections from the gateway. 

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: BackendTLSPolicy
metadata:
  name: sample-backend-tls
  namespace: default
spec:
  targetRef: 
    group: ''
    kind: Service
    name: sample-backend
    namespace: default
  tls: 
    caCertRefs:
    - name: backend-certs
      group: ''
      kind: Secret
    hostname: example.com
```

### Configurable NodePort

With Voyager Gateway, we can specify port mapping for NodePort type services. We have introduced specific annotation for this. The annotation is kept on the gateway object so the cluster administrator can configure it.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  …  …  … 
  annotations: 
    'port-mapping.gateway.voyagermesh.com/10002' : '30005'
spec:
  gatewayClassName: eg
  listeners:
  - name: sample-tls-lis 
    port: 10002
    … … … 
```


## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To watch tutorials of various Production-Grade Kubernetes Tools Subscribe our [YouTube](https://www.youtube.com/c/AppsCodeInc/) channel.