---
title: Introducing KubeVault v2021.08.02
date: "2021-08-02"
weight: 25
authors:
- Md Kamol Hasan
tags:
- elasticsearch
- hashicorp
- kubernetes
- kubevault
- raft
- secret-management
- security
- vault
---

We are very excited to announce KubeVault Enterprise Edition with the release `v2021.08.02`. The KubeVault `v2021.08.02` contains various feature improvements and bug fixes for a better user experience. It also comes with a KubeVault Community Edition which is **free of cost** but only limited to the `default` namespace.

- [Install KubeVault](https://kubevault.com/docs/v2021.08.02/setup/)

[KubeVault](https://kubevault.com) is a Kubernetes operator for [HashiCorp Vault](https://www.vaultproject.io/). The Vault is a tool for secrets management, encryption as a service, and privileged access management. The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It also supports various secret engines management, policy management in the Kubernetes native way.

In this post, we are going to highlight the major changes. You can find the complete changelog [here](https://github.com/kubevault/CHANGELOG).

## What's new in this release?

- **Use Statefulset for Vault Server Deployment**
  
  The operator uses k8s StatefulSet object instead of k8s Deployment object to deploy Vault Server.

- **Support Kubernetes Secrets Store CSI Driver**
  
  The Secrets Store CSI driver [secrets-store.csi.k8s.io](https://github.com/kubernetes-sigs/secrets-store-csi-driver) allows Kubernetes to mount multiple secrets, keys, and certs stored in Vault into their pods as a volume. Once the Volume is attached, the data in it is mounted into the container's file system.

- **Support Integrated Storage Backend (Raft)**

  Now the Vault Server can be created with the integrated storage backend (Raft). The k8s PVCs are used to persist the data over time. The highly available Vault Cluster deployment has never been this easy before.

- **Support for Elasticsearch Secret Engine**

  The support for Elasticsearch secret engine management is added to this release. It also includes enabling secret engine, role creation, and access-key-request in the Kubernetes native way (i.e. by CRDs).  

- **Accidental Deletion Prevention**
  The support for the `Termination Policy` has been added in this release, so that users can now clean up their resources accordingly or even prevent any accidental deletion of resources.

  - DoNotTerminate
  - Halt
  - Delete (Default)
  - WipeOut

- **Dynamic Phase Reflections**

  The Vault Server Phase is now dynamic and it reflects the current status of the Vault Server. The status field of the VaultServer CRD has a condition array  with various `states` of the cluster with the `lastTransitionTime`.

  Possible values of `status.phase`:
  - `Initializing`
  - `Sealed`
  - `Unsealing`
  - `Critical`
  - `Ready`
  - `NotReady`

  ```yaml
  status:
    phase: Ready
    conditions:
      - lastTransitionTime: "2021-07-09T12:02:27Z"
        message: VaultServer Initializing process is completed
        reason: VaultServerInitializingCompleted
        status: "False"
        type: Initializing
      - lastTransitionTime: "2021-07-09T12:03:34Z"
        message: All desired replicas are ready.
        reason: AllReplicasReady
        status: "True"
        type: AllReplicasReady
      - lastTransitionTime: "2021-07-09T12:02:27Z"
        message: VaultServer is initialized and accepting connection
        reason: VaultServerAcceptingConnection
        status: "True"
        type: AcceptingConnection
      - lastTransitionTime: "2021-07-09T12:02:27Z"
        message: VaultServer is already Initialized
        reason: VaultServerInitialized
        status: "True"
        type: Initialized
      - lastTransitionTime: "2021-07-09T12:02:27Z"
        message: VaultServer unsealing is completed
        reason: VaultServerUnsealingCompleted
        status: "False"
        type: Unsealing
      - lastTransitionTime: "2021-07-09T12:02:27Z"
        message: VaultServer is initialized and unsealed
        reason: VaultServerUnsealed
        status: "True"
        type: Unsealed
  ```

- **Disable TLS**

  Now the Vault Server can be deployed with both TLS enabled (ie. `https://` ) or disabled (ie. `http://` ) modes. The TLS disabled mode is handy when comes to testing and it is recommended not to use it in the production environment.

## Why Enterprise Edition?

AppsCode started as a commercial entity to accelerate the adoption of Kubernetes and containers in the Enterprise. We launched a number of open source products like [Voyager](https://voyagermesh.com), [Stash](https://stash.run), [KubeDB](https://kubedb.com), [KubeVault](https://kubevault.com), [Kubeform](https://kubeform.com), etc. Consequently, we began receiving many feature requests, bug reports, and general support questions via our GitHub repositories and public slack account. We are very much thankful to the existing users of our open source projects. As a commercial entity, now we are focusing on building a sustainable business for the future. Supporting open-source projects is not sustainable without any kind of revenue stream. Last year we offered support packages for our various products; but that did not help. Often, we get queries in the meetings with prospective customers regarding the difference between the free open source version and the paid version. In the process, we have learned that support is not enough to convert those users into customers. *For businesses, it is not rational to pay for something that they can get for free.* As a result, since the end of last year, we started developing closed source features for the “Enterprise” version of our products.

## Webinar

We are overjoyed to announce a webinar on 12 August, 2021. In this webinar, our experts of AppsCode will talk on and demonstrate "Manage Vault in Kubernetes native way using KubeVault".

Check here for details: https://appscode.com/webinar and don't forget to register!

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install KubeVault, please follow the installation instruction from [here](https://kubevault.com/docs/v2021.08.02/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeVault).

If you have found a bug with KubeVault or want to request new features, please [file an issue](https://github.com/kubevault/project/issues/new).
  