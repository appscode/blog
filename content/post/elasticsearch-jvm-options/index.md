---
title: Enable G1 Garbage Collector for Elasticsearch to Avoid Circuit Breaker Exceptions
date: 2021-07-10
weight: 26
authors:
  - Md Kamol Hasan
tags:
  - cloud-native
  - kubernetes
  - database
  - elasticsearch
  - garbage-collector
  - jvm-options
  - G1GC
  - CMS
  - kubedb
---


By default, the Opendistro of Elasticsearch cluster starts Concurrent Mark Sweep (`CMS`) garbage collector. In this blog post, we will see how the default `jvm.options` may lead to the circuit breaker exceptions and how can we avoid it by enabling the Garbage-First(`G1`) garbage collector.

## Elasticsearch Garbage Collectors

Elasticsearch mainly uses two different garbage collectors of Java: Concurrent Mark Sweep (CMS) and Garbage-First(G1).

[jvm.options](https://github.com/elastic/elasticsearch/blob/v7.13.3/distribution/src/config/jvm.options):

```options
## GC configuration

8-13:-XX:+UseConcMarkSweepGC
8-13:-XX:CMSInitiatingOccupancyFraction=75
8-13:-XX:+UseCMSInitiatingOccupancyOnly

## G1GC Configuration
# to use G1GC, uncomment the next two lines and update the version on the
# following three lines to your version of the JDK

# 8-13:-XX:-UseConcMarkSweepGC
# 8-13:-XX:-UseCMSInitiatingOccupancyOnly
14-:-XX:+UseG1GC
```

The CMS uses multiple garbage collector treads for garbage collection. It is designed for applications that prefer shorter garbage collection pauses. Overheads occur when the collector needs to promote young objects to the old generations, but didn't have enough time to clear the space.

```log
[2021-07-15T08:23:27,847][WARN ][o.e.m.j.JvmGcMonitorService] [elasticsearch-client-1] [gc][400] overhead, spent [856ms] collecting in the last [1.5s] 
```