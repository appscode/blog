---
title: "Stash: Introducing a Better Cloud-Native Backup Experience"
date: 2022-03-09
weight: 15
authors:
- Hossain Mahmud
tags:
- cloud-native
- kubernetes
- stash
- backup
- restore
- cross-namespace
- cli
---

## Summary
On March 9, 2022, Appscode held a webinar on **Stash: Introducing a Better Cloud-Native Backup Experience**. Key contents of the webinars are:
1) Cross-namespaced Backup and Recovery
2) Stash Kubectl Plugin
    * Triggering instant backup
    * Pause, Resume backup
    * Download snapshot locally
    * Debugging helper
3) Sending Backup notification to Slack
4) Q & A Session

## Description of the Webinar
Initially, they gave an overview of Stash. They discussed the features, supported applications, and supported platforms by Stash.
Later, they introduced the new cross-namespaced backup and recovery support. They deployed a `MySQL` instance to show the cross-namespaced backup and recovery. They discussed how the `usagePolicy` section in the `Repository` can be used to restrict its usage from different namespaces. After that, they introduced the enhanced Stash `CLI` and showed its usage. Finally, they introduced another new feature of sending backup notifications to `Slack`. They discussed how to configure a Slack incoming webhook with Stash and showed a demo for failed and successful backup and restore notifications to Slack.


  Take a deep dive into the full webinar below:

<iframe width="560" height="315" src="https://www.youtube.com/embed/MREdcm9S8Xg" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## What Next?

Please try the latest release and give us your valuable feedback.

- If you want to install Stash in a clean cluster, please follow the installation instruction from [here](https://stash.run/docs/latest/setup/).
- If you want to upgrade Stash from a previous version, please follow the upgrade instruction from [here](https://stash.run/docs/latest/setup/upgrade/).

### Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the Stash community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#stash`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeStash).

If you have found a bug with Stash or want to request new features, please [file an issue](https://github.com/stashed/project/issues/new).
