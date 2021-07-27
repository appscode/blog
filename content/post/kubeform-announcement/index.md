---
title: Introducing Kubeform v2021.07.13
date: 2021-07-27
weight: 25
authors:
  - Shohag Rana
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - mariadb
  - memcached
  - mongodb
  - mysql
  - postgresql
  - redis
  - kubedb
---

We are happy to announce Kubeform v2021.07.13. This post lists all the changes and features this release brings to you.

## Kubeform Enterprise

In this release, we are announcing the Kubeform Enterprise edition. Currently, In the Kubeform community edition, you can do everything the enterprise edition does, but you will be limited to only the `default` namespace. We plan to bring some exciting new features in the enterprise edition that will not be available to the community edition in the future release. Please see the What’s Next section to get an idea of the upcoming features.

## Re-designed the Architecture of Kubeform

In this release, we have re-designed the Kubeform architecture. Kubeform controller is now divided into 5 different controllers, one controller for each cloud provider. Kubeform currently supports 5 major cloud providers, Google, AWS, Azure, Linode & DigitalOcean.

## No Dependency on Terraform CLI

We’ve removed dependency from the Terraform CLI. Previously, we’ve used the terraform CLI in the Kubeform controller to provision the cloud resources. But, from this release, we are not using the Terraform CLI anymore, instead, we are using the respective resource APIs to manage the resources via the Kubeform controllers.

## Accidental Deletion Protection

This release adds TerminationPolicy to protect the resource against accidental deletion. You can provide `DoNotTerminate` as the `TerminationPolicy` of your resource. When you delete the resource that has DoNotTerminate set as its TerminationPolicy, you’ll get an error message from the validation webhook saying that, the resource can’t be terminated when TerminationPolicy is set to DoNotTerminate. So, this protects you from the accidental deletion of the resource. If you want to terminate the resource, you can update the TerminationPolicy as Delete. Then, the resource will be terminated successfully without any error.

## Update Policy

We’ve added UpdatePolicy to ensure that, the resource doesn’t get deleted without your approval while updating the resource. If the user sets the UpdatePolicy as DoNotDestroy, the resource won’t get deleted in the process of updating. The kubeform resource will be in the Failed state. To recover from this, the user will have to change the UpdatePolicy to Destroy or the field that will cause the resource to be deleted and then re-created again. We plan to add this in the validation webhook so that the user gets the error while applying the resource.

## Sensitive Secret Watcher

We’ve added a sensitive secret watcher. So, whenever there is a change in the sensitive secret, the kubeform resource will be reconciled by the controller and the corresponding cloud resource will be updated.

## New Status and Conditions

Kubeform resources now have 4 status phases. These are:

* InProgress: It means the resource is now reconciling.
* Current: It means the resource is reconciled successfully and the corresponding cloud resource is updated.
* Terminating: It means the resource is currently in the process of deleting.
* Failed: It means the resource has encountered some error while reconciling.

Also, we’ve added conditions in the kubeform resources.

## Dropped support for Terraform Module

In this release, We’ve dropped support for the Terraform module.

## Upcoming Features

* Remote backend for the resources. The resource state can be maintained remotely, such as google bucket, amazon s3 bucket, etc.
* A CLI command to generate Terraform `.tf` files from the kubeform resources.
* A CLI command to check the execution plan of the kubeform resources. Using this plan, you will know the configuration of the resource that will be created by kubeform,  the changes in the resource when updating the resource configuration, etc.
* Halt and Resume resources. Using this feature, you can keep the resource in the Kubernetes, but terminate the actual cloud resource. Then, when you need the resource again, you can just resume the resource and the same resource will be created again!
* Seamless integration with KubeDB. KubeDB is a product by AppsCode that simplifies and automates routine database tasks such as provisioning, patching, backup, recovery, failure detection, and repair for various popular databases on private and public clouds.

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install KubeDB, please follow the installation instruction from [here](https://kubedb.com/docs/v2021.06.23/setup).

* If you want to upgrade KubeDB from a previous version, please follow the upgrade instruction from [here](https://kubedb.com/docs/v2021.06.23/setup/upgrade/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
