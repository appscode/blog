---
title: Introducing Kubeform Module
date: 2022-02-10
weight: 25
authors:
  - Sahadat Hossain 
tags:
  - cloud-native
  - kubernetes
  - on-prem
  - kubeform
  - terraform
  - module
  - cli
  - gcp
  - linode
  - aws
  - azure
---


On 10th Feb 2022, AppsCode held a webinar on **"Introducing Kubeform Module"**. Key contents of the webinar are:

* Terraform Module supports in Kubeform by Kubeform Module
* Generating Kubeform Module Definitions from existing terraform module
* Kubeform Module Object to use generated Kubeform Module Definition as a reference of module
* Supports both public and private git repo of terraform module
* Demo:
  * Create, Update and Deletion of Kubeform Module resource

## Description of the Webinar

It is required to install the followings to get started:
  - Kubeform Module Operator
  - Kubeform CLI

Earlier in the webinar the speaker discussed about `Kubeform`, it's features and the cloud providers that Kubeform supports.

After that discussed the `Kubeform Module`, the workflow of Kubeform Module in details with showing how to generate Kubeform Module Definitions from existing `terraform modules` using `Kubeform CLI`. Also showed yaml manifests of `Kubeform Module Definitions`, Kubeform Module Object and described those manifests, showed the comparison between the typical terraform configuration `vs` Kubeform Module configuration to use terraform module.

Finally, there was a comprehensive hands-on demo on `Kubeform Module` where speaker showed full life cycle of Kubeform Module, basically there speaker showed how to create a Kubeform Module Object which used generated Kubeform Module Definition to use any terraform module.


Take a deep dive into the full webinar below:

<iframe width="800" height="500" src="https://www.youtube.com/embed/FiSuJWDR_FY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

* If you want to install Kubeform, please follow the installation instruction from [here](http://www.kubeform.com/docs/latest/setup).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Kubeform community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#Kubeform`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/Kubeform).

If you have found a bug with Kubeform or want to request for new features, please [file an issue](https://github.com/Kubeform/Kubeform/issues/new).
