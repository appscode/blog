---
title: Connect AWS EKS Clusters with Submariner
date: "2023-02-06"
weight: 26
authors:
- Pulak Kanti Bhowmick
tags:
- aws
- clusters
- eks
- kubernetes
- networking
- service
- submariner
---

## Submariner

[Submariner](https://github.com/submariner-io/submariner) is a CNCF sandbox project which can enable direct networking between Pods and Services in different Kubernetes clusters, either on-premises or in the cloud.

In this blog post, I am going to show how can you connect two AWS EKS Kubernetes clusters with default vpc cni using Submariner.

## Installation & Verify

1. At first, we have to create two Kubernetes clusters. To do so, we'll use [eksctl](https://eksctl.io/).

submariner-1.yaml:
```yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: submariner-1
  region: us-east-1
vpc:
  cidr: 192.168.0.0/16
nodeGroups:
  - name: ng-1
    instanceType: t3.medium
    desiredCapacity: 3
```

To create:

```bash
eksctl create cluster -f submariner1.yaml
```

submariner-2.yaml
```yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: submariner-2
  region: us-east-2
vpc:
  cidr: 10.0.0.0/16
nodeGroups:
  - name: ng-1
    instanceType: t3.medium
    desiredCapacity: 3
```

To create:

```
eksctl create cluster -f submariner2.yaml
```

Note: According to Submariner [prerequisites](https://submariner.io/getting-started/#prerequisites): `Worker node IPs on all connected clusters must be outside of the Pod/Service CIDR ranges.`. To make sure of this, we are using two different vpc cidr blocks for the clusters.

2. Now, we are going to install Submariner broker in Submariner-1 cluster. Here we are using [subctl](https://submariner.io/operations/deployment/subctl/) CLI tool.

```bash
$ subctl deploy-broker --kubeconfig /path/to/submariner-1/kubeconfig
 ✓ Setting up broker RBAC 
 ✓ Deploying the Submariner operator 
 ✓ Created operator namespace: submariner-operator
 ✓ Created operator service account and role
 ✓ Created submariner service account and role
 ✓ Created lighthouse service account and role
 ✓ Deployed the operator successfully
 ✓ Deploying the broker
 ✓ Saving broker info to file "broker-info.subm"
```

3. Now, both Submariner-1 & 2 clusters are needed to be join in the broker cluster(Submariner-1)

```bash
$ subctl join broker-info.subm --clusterid cluster-a --kubeconfig /path/to/submariner-1/kubeconfig
```

```bash
$ subctl join broker-info.subm --clusterid cluster-b --kubeconfig /path/to/submariner-2/kubeconfig
```
Let's check network details of both clusters:

Submariner-1:
```bash
~ $ subctl show networks
Cluster "submariner-1.us-east-1.eksctl.io"
 ✓ Showing Network details
    Discovered network details via Submariner:
        Network plugin:  generic
        Service CIDRs:   [10.100.0.0/16]
        Cluster CIDRs:   [192.168.0.0/16]
```

Submariner-2:
```bash
~ $ subctl show networks
Cluster "submariner-2.us-east-2.eksctl.io"
 ✓ Showing Network details
    Discovered network details via Submariner:
        Network plugin:  generic
        Service CIDRs:   [172.20.0.0/16]
        Cluster CIDRs:   [10.0.0.0/16]
```

4. Let's prepare AWS Clusters for Submariner. Submariner Gateway nodes need to be able to accept traffic over UDP ports (4500 and 4490 by default). Submariner also uses UDP port 4800 to encapsulate traffic from the worker and master nodes to the Gateway nodes, and TCP port 8080 to retrieve metrics from the Gateway nodes. Ref: [Submariner doc](https://submariner.io/getting-started/quickstart/openshift/aws/#prepare-aws-clusters-for-submariner)

So, we need to open those ports from the associate security group. 

Then, we can verify connectivity with `subctl`.

From Submariner-1 cluster:
```bash
~ $ subctl show connections
Cluster "submariner-1.us-east-1.eksctl.io"
 ✓ Showing Connections 
GATEWAY                          CLUSTER     REMOTE IP       NAT   CABLE DRIVER   SUBNETS                      STATUS      RTT avg.      
ip-10-0-22-156.us-east-2.compu   cluster-b   <remote-ip>     yes   libreswan      172.20.0.0/16, 10.0.0.0/16   connected   11.424726ms 
```

From Submariner-2 cluster:
```
~ $ subctl show connections
Cluster "submariner-2.us-east-2.eksctl.io"
 ✓ Showing Connections 
GATEWAY                         CLUSTER     REMOTE IP      NAT   CABLE DRIVER   SUBNETS                         STATUS      RTT avg.     
ip-192-168-31-89.ec2.internal   cluster-a   <remote-ip>    yes   libreswan      10.100.0.0/16, 192.168.0.0/16   connected   11.40141ms 
```

Both cluster connection statuses are showing `connected`. If not, then you can find out the reason by running `subctl diagnose all` command.

5. Let's deploy an nginx in Submariner-1 cluster and export via `subctl`.
```bash
~ $ kubectl create ns demo
namespace/demo created
~ $ kubectl create deploy nginx --image=nginx --replicas=1 -n demo
deployment.apps/nginx created
~ $ kubectl expose deploy nginx --port=80 -n demo
service/nginx exposed
~ $ kubectl get svc -n demo
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
nginx   ClusterIP   10.100.39.195   <none>        80/TCP    10s
~ $ subctl export service nginx -n demo
 ✓ Service exported successfully
```
Note: By default, in a cluster created by `eksctl`, Pods & Nodes are sharing IPs from a single IP CIDR range shown above as `Cluster CIDRs`. AWS ENI attached with the Nodes has a source/destination check enabled which can block traffic from any non-GW node to a GW node in any cluster. To allow such traffic, we need to disable the source/destination check from the attached ENI(Not recommended in production). To be more specific, the nginx service from the submariner-1 cluster has an IP from the 10.100.0.0/16 CIDR range. But Pods in the submariner-2 cluster have IP from the 10.0.0.0/16 CIDR range. As packets are traveling from non-GW node to GW with svc IP from 10.100.0.0/16 range, ENI drops those packets.

Let's verify from submariner-2 cluster:

```bash
~ $ kubectl run tmp-shell --rm -i --tty --image quay.io/submariner/nettest  -- /bin/bash

If you don't see a command prompt, try pressing enter.
bash-5.0# curl cluster-a.nginx.demo.svc.clusterset.local
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
bash-5.0# dig +short cluster-a.nginx.demo.svc.clusterset.local
10.100.39.195
```

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).


