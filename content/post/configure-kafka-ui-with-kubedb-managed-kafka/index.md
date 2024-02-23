---
title: Configure Kafka-UI with KubeDB Managed Kafka and Kafka Connect Cluster
date: "2024-02-26"
weight: 14
authors:
- M. Obaydullah
tags:
- apache-kafka
- kafka-ui
- cloud-native
- acl
- kafka
- kubedb
- kubernetes
- streaming-platform
---

## Overview

Apache Kafka is a powerful distributed event streaming platform, Kafka Connect Cluster is a framework for building connectors to integrate Kafka with other systems.

There are a few open source Kafka UI. We are using `UI for Apache Kafka` by Provectus. Provectus Kafka UI simplifies the management and monitoring of multiple Kafka clusters. Together, they provide a robust ecosystem for real-time data streaming and processing.

## Workflow

In this tutorial, We will cover the following steps:

1) Configure Kafka ACL and Kafka User
2) Install UI for Apache Kafka
3) Kafka User Management in Kafka UI
