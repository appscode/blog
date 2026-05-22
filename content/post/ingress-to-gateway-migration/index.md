---
title: Zero-Downtime Migration from Ingress to Kubernetes Gateway API in ACE
date: "2026-05-22"
weight: 15
authors:
- Arnob Kumar Saha
tags:
- ace
- envoy-gateway
- gateway-api
- ingress
- kubedb-platform
- kubernetes
- migration
---

The Kubernetes [Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway/) is the designated successor to Ingress. It offers richer routing semantics, native TCP/UDP support, explicit timeout configuration, upstream TLS policies, and a cleaner separation between infrastructure owners and application teams — all without relying on controller-specific annotations. For kubedb-platform, migrating to Gateway API was a clear necessity. The hard part was doing it without breaking anything for our existing self-hosted users.

This post explains the three-phase strategy we used to achieve a fully transparent, zero-downtime migration. In the full demonstration, Lets assume "dbaas.kubedb.cloud" is the domain user is using.

---

## Phase 1: The Starting Point — Ingress Only

In the initial state, all HTTP(S) traffic is routed through nginx Ingress using the `nginx-ace` IngressClass. There are four Ingress objects in the `ace` namespace:

| Ingress | Purpose |
|---------|---------|
| `ace` | Main routing — 7 path-prefix rules across all ACE services |
| `ace-home` | Redirect `/` to `/accounts/selfhost-home` |
| `ace-nats` | Proxy WebSocket/HTTP traffic to NATS (regex path, upstream TLS) |
| `service-backend` | Route `/bind/` to the service backend |

The main `ace` Ingress maps URL paths to backend services:

```yaml
rules:
- http:
    paths:
    - path: /accounts   → ace-platform-api:80
    - path: /api        → ace-platform-api:80
    - path: /console    → ace-cluster-ui:80
    - path: /db         → ace-kubedb-ui:80
    - path: /id         → ace-platform-ui:80
    - path: /grafana    → ace-grafana:80
    - path: /prometheus → ace-trickster:4000
```

The `ace-home` Ingress uses `nginx.ingress.kubernetes.io/rewrite-target: /accounts/selfhost-home` to redirect root traffic. The `ace-nats` Ingress relies on two nginx-specific annotations — `backend-protocol: HTTPS` for upstream TLS and `use-regex: "true"` with a capture group in the path for stripping the `/nats` prefix before forwarding to the NATS service.

DNS is managed by an `ExternalDNS` resource named `ace-ingress`, which watches Ingress objects and registers:

```
dbaas.kubedb.cloud → <nginx LB IP>
```

This is the URL users know and use.

---

## Phase 2: Dual-Stack — Gateway API Running Alongside Ingress

Before cutting over, we deploy the full Gateway API configuration alongside the existing Ingress objects. Both serve live traffic from their own load balancers during this phase.

### The Gateway Resource

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ace
  namespace: ace
spec:
  gatewayClassName: ace   # Envoy Gateway
  listeners:
  - name: http
    port: 80
    protocol: HTTP
  - name: https
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - name: ace-gw-cert
  - name: nats-tcp
    port: 4222
    protocol: TCP
  - name: s3proxy-tcp
    port: 4224
    protocol: TCP
```

The `https` listener terminates TLS using a certificate provisioned specifically for the Gateway. The two TCP listeners handle NATS and S3 proxy traffic at the transport layer — something Ingress simply cannot do.

### HTTPRoutes Mirror the Ingress Rules

Each Ingress object has a corresponding HTTPRoute. The translation makes the implicit explicit.

**Main routing (`ace` HTTPRoute):**
```yaml
rules:
- matches:
  - path: {type: PathPrefix, value: /api}
  backendRefs:
  - name: ace-platform-api
    port: 80
  timeouts:
    request: 2m
    backendRequest: 2m
# ... same pattern for /accounts, /console, /db, /id, /grafana, /prometheus
```

Explicit timeouts per route — not possible with standard Ingress.

**Root redirect (`ace-home` HTTPRoute):**
```yaml
rules:
- matches:
  - path: {type: PathPrefix, value: /}
  filters:
  - type: URLRewrite
    urlRewrite:
      path:
        type: ReplaceFullPath
        replaceFullPath: /accounts/selfhost-home
  backendRefs:
  - name: ace-platform-api
    port: 80
```

The `nginx.ingress.kubernetes.io/rewrite-target` annotation is replaced by a standard `URLRewrite` filter.

**HTTP→HTTPS redirect (`ace-http-redirect` HTTPRoute):**
```yaml
rules:
- matches:
  - path: {type: PathPrefix, value: /}
  filters:
  - type: RequestRedirect
    requestRedirect:
      scheme: https
      statusCode: 301
```

No annotation needed — this is a first-class Gateway API primitive.

**NATS routing (`ace-nats` HTTPRoute):**
```yaml
rules:
- matches:
  - path: {type: PathPrefix, value: /nats}
  filters:
  - type: URLRewrite
    urlRewrite:
      path:
        type: ReplacePrefixMatch
        replacePrefixMatch: /
  backendRefs:
  - name: ace-nats
    port: 443
```

The regex path + capture group from Ingress is replaced by `ReplacePrefixMatch`. Upstream TLS validation (previously `backend-protocol: HTTPS`) is expressed as a separate `BackendTLSPolicy` resource that references the CA certificate and validates the upstream hostname — a proper policy object rather than an opaque annotation.

### Two DNS Entries Coexist

A second `ExternalDNS` resource (`ace-gw`) watches the Envoy proxy Service (filtered by label) and registers:

```
ace.ace.dbaas.kubedb.cloud → <Envoy LB IP>
```

During Phase 2, both domains are active:

| Domain | Points To | Managed By |
|--------|-----------|------------|
| `dbaas.kubedb.cloud` | nginx LB | `ace-ingress` ExternalDNS |
| `ace.ace.dbaas.kubedb.cloud` | Envoy LB | `ace-gw` ExternalDNS |

Users can validate the new Gateway URL independently before any cutover.

---

## Phase 3: The Switching Moment — Zero-Downtime Cutover

When the user is ready, they disable the ingress feature in their selfhost installer values and trigger an ACE upgrade or reconfigure:

```yaml
# Before
ingress-dns:
  enabled: true

# After
ingress-dns:
  enabled: false
```

On the next upgrade, the ACE upgrader job automates the cutover in three steps.

### Step 1: Helm Upgrade Deletes Ingress Resources

`upgrader.UpdateACEInstaller()` runs a helm upgrade with the new values. Since `ingress-dns.enabled` is now false, helm removes the `ace-ingress` ExternalDNS resource and all four Ingress objects from the cluster. The nginx LB still exists momentarily, but DNS no longer points to it.

### Step 2: Upgrader Detects the Transition

The upgrader checks the installer values:

```go
gatewayDNSEnabled := getNestedBool(installerValues, ..., "gateway-dns", "enabled")
ingressDNSEnabled := getNestedBool(installerValues, ..., "ingress-dns", "enabled")

if gatewayDNSEnabled && !ingressDNSEnabled {
    err = syncExternalDNS(kc)
}
```

### Step 3: `syncExternalDNS` Performs the DNS Handoff

This is the critical piece. It cannot simply create the new DNS record immediately — the old `ace-ingress` ExternalDNS was just deleted by helm, but the deletion may not be complete yet. If both ExternalDNS resources try to own the same DNS record simultaneously, they will conflict.

The function waits for the deletion to complete before proceeding:

```go
func waitForIngressExternalDNSToBeDeleted(kc client.Client) error {
    return wait.PollUntilContextTimeout(ctx, 3*time.Second, 20*time.Minute, true,
        func(ctx context.Context) (bool, error) {
            var ingDNS extdnsv1a1.ExternalDNS
            if err := kc.Get(ctx, ..., &ingDNS); errors.IsNotFound(err) {
                return true, nil   // gone, proceed
            }
            if ingDNS.Labels[meta_util.ManagedByLabelKey] == "Helm" {
                return false, nil  // still exists, keep waiting
            }
            return true, nil
        },
    )
}
```

Once deleted, `copyExternalDNS` creates a new `ace-ingress` ExternalDNS modeled on the existing `ace-gw` spec, but targeting the base domain instead of the `ace.ace.*` subdomain. It strips the `ace.ace.` prefix from the gateway's DNS record to derive the base FQDN:

```
ace.ace.dbaas.kubedb.cloud  →  dbaas.kubedb.cloud
```

The new ExternalDNS watches the same Envoy proxy Service, so:

```
dbaas.kubedb.cloud → <Envoy LB IP>
```

The user's existing URL now routes to Gateway — without any change on their end.

### Cutover Sequence

```
Helm upgrade
  └─ delete Ingress objects (ace, ace-home, ace-nats, service-backend)
  └─ delete ace-ingress ExternalDNS

Upgrader polls: waiting for ace-ingress ExternalDNS to disappear...

  └─ create new ace-ingress ExternalDNS (source: Envoy Service)
       └─ DNS: dbaas.kubedb.cloud → Envoy LB IP  ✓

Users continue using dbaas.kubedb.cloud — now served by Gateway
```

---

## Ingress vs Gateway API: What Changed

| Capability | Ingress | Gateway API |
|------------|---------|-------------|
| Path rewrite | `nginx.ingress.kubernetes.io/rewrite-target` annotation | `URLRewrite` filter on HTTPRoute |
| Upstream TLS | `nginx.ingress.kubernetes.io/backend-protocol: HTTPS` annotation | `BackendTLSPolicy` resource |
| HTTP→HTTPS redirect | Controller-specific annotation | `RequestRedirect` filter on HTTPRoute |
| Per-route timeouts | Not supported | Native `timeouts` field on HTTPRoute rule |
| TCP/UDP routing | Not supported | `TCPRoute` / `UDPRoute` + TCP/UDP listener |
| Regex path matching | `ImplementationSpecific` + `use-regex` annotation | Not needed — `ReplacePrefixMatch` covers the common case |

---

## Summary

The migration involved three stages: running ingress-only, running both in parallel, and then atomically handing off DNS at upgrade time. The dual-stack phase gave us a safe window to validate Gateway behavior before committing. The upgrader job's `syncExternalDNS` function handled the DNS handoff automatically — waiting for the old record owner to be gone before claiming it, ensuring no conflict and no interruption for users.

Gateway API is now the default for all new ACE installations, and existing users migrate transparently on their next upgrade.
