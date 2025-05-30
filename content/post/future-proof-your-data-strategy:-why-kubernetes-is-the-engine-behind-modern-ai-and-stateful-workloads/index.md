---
title: Future-Proof Your Data Strategy - Why Kubernetes is the Engine Behind Stateful Workloads & Modern AI
date: "2025-05-15"
weight: 14
authors:
- Neaj Morshad
tags:
- ai
- appscode
- cloud-native
- databases
- dbaas
- gcp
- gke
- kubedb
- kubernetes
---

# Why do Enterprises Choose Kubernetes for Stateful Apps?
In today’s rapidly evolving tech landscape, managing stateful applications—databases, caches, queues—at scale, efficiently is a top priority for enterprises.    
At Google Cloud Next 25, titled "Data on Kubernetes: Run stateful apps and AI workloads on GKE", highlighted the growing adoption of Kubernetes for databases, AI, and machine learning (ML) workloads, showcasing its cost-efficiency, scalability, and performance benefits. Drawing from insights shared in that session, this blog explores why Kubernetes—paired with solutions like KubeDB from AppsCode—is the ideal platform for running stateful applications and future-proofing AI workloads.


# The Growing Adoption of Databases on Kubernetes

The Google Cloud Next session kicked off with a compelling statistic from a survey conducted by the Data on Kubernetes organization. In a poll of 150 IT leaders of organization size 1000–50,000+, **databases emerged as one of the most common workloads running on Kubernetes.**

![Diagram of the survey conducted by the Data on Kubernetes organization, 2024 Survey of 150 Technology Leaders of Org Size 1,0000-50,0000+](databases-on-kubernetes.png)

Database workloads remain the most prevalent use case for the third consecutive year, with **over 80% of IT leaders run databases on Kubernetes in production by 2024**.

This trend reflects a significant shift from Kubernetes being primarily used for stateless applications to now **confidently handling stateful workloads like databases.** Advancements in storage solutions and orchestration capabilities have made it a robust choice for stateful workloads. **The adoption of Kubernetes Operators** has played a crucial role in this transition by automating tasks such as backups, scaling, and updates, thereby simplifying the management of database instances on Kubernetes.

The session’s audience vibe check further reinforced this trend. When asked how many attendees were running databases on Kubernetes, a notable portion raised their hands. These real-world use cases demonstrate that **Kubernetes is no longer just a theoretical fit for stateful apps—it’s a proven solution adopted by organizations worldwide.**

## Why the shift?
**Unified Infrastructure:** Managing stateless and stateful applications on the same platform simplifies operations, observability, and tooling.   
**Flexibility & Control:** Kubernetes offers fine-grained control over deployment, scaling, networking, and storage configuration that often surpasses managed service limitations.   
**Avoiding Lock in:** Running databases within Kubernetes provides portability across cloud providers and on-premises environments.   
**Cost Optimization:** As highlighted in the talk, Kubernetes (with tools like Hyperdisk, Storage Pools, and Data Cache) enables sophisticated cost savings through resource right-sizing and efficient utilization, challenging the assumption that managed services are inherently cheaper at scale.   

At AppsCode, we’ve seen this trend firsthand. Even Google’s customers, who have access to managed database services on Google Cloud, are increasingly turning to Kubernetes to run stateful applications. Why? Kubernetes offers **unparalleled flexibility, cost savings, high availability, and scalability.** With **KubeDB for running production-grade databases on Kubernetes, organizations can simplify database management** while leveraging the same orchestration power that makes Kubernetes so effective for stateless workloads.

# Kubernetes: The Key to Cost Savings and Performance

One of the standout themes from the Google Cloud Next session was the emphasis on cost savings and performance optimization for stateful workloads on Google Kubernetes Engine (GKE). Google’s product manager, Brian Kaufman, highlighted several innovative storage solutions that make Kubernetes an attractive choice:

**Hyperdisk for Tunable Performance:** Unlike traditional persistent disks, Hyperdisk allows users to tune IOPS and throughput independently of capacity. This means you only pay for the performance you need, avoiding overprovisioning and reducing costs.

**Storage Pools for Efficiency:** By pooling multiple disks (e.g., boot disks and persistent storage), organizations can share capacity and performance across workloads. This approach achieves up to 80% capacity utilization, significantly lowering costs compared to traditional storage provisioning, where unused capacity is common.

**GKE Data Cache for Low Latency:** For read-heavy workloads like vector databases, GKE’s data cache uses local SSDs to provide sub-millisecond latency at a fraction of the cost of memory ($0.08/GB vs. $3/GB for memory). Tests with PostgreSQL showed an 80% reduction in latency and a 480% increase in transactions per second.

These solutions underscore Kubernetes’ ability to deliver high performance without breaking the bank. With KubeDB, AppsCode takes this a step further by providing a unified platform to manage 20+ databases like PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch on Kubernetes, ensuring seamless integration with these advanced storage options.

# Future-Proofing AI Workloads with Kubernetes

The session also delved into the unique demands of AI and ML workloads, which are **increasingly being deployed on Kubernetes**. AI workloads are broadly categorized into training and inference, each with distinct requirements:

**Training Workloads:** These involve large-scale data reads from object storage and frequent checkpointing to ensure fault tolerance. Google’s solutions, like multi-tiered checkpointing and Cloud Storage FUSE with caching, optimize these processes by minimizing disruptions and accelerating data access.

**Inference Workloads:** Inference requires rapid startup times to handle fluctuating demand. Large models, such as Llama 3.1 (130 GB), can take up to 20 minutes to download from object storage. Google’s container preloading and Hyperdisk ML solutions reduce startup times by up to 29x, enabling real-time scaling for inference-heavy applications.

These advancements make Kubernetes a future-proof platform for AI workloads. As AI models grow in size and complexity, Kubernetes’ ability to scale dynamically, manage diverse storage needs, and integrate with accelerators like GPUs positions it as the go-to solution. **KubeDB powerfully complements this by simplifying the management of databases that often serve as the backbone for AI data pipelines, ensuring reliability and performance at scale.**

## AI-Optimized Data Stores in KubeDB

### SingleStore
SingleStore’s native vector data type and indexed ANN (HNSW, IVF) deliver high-throughput vector search alongside real-time analytics and SQL compatibility.

### Apache Druid
Druid is a real-time OLAP datastore providing sub-second analytical queries and plug-in support for ML-driven dashboards and anomaly detection.

### Milvus
Milvus is an open-source vector database built for GenAI applications. Designed for high-performance approximate nearest neighbor (ANN) search over massive, high-dimensional embeddings.

### Elasticsearch & OpenSearch
Both Elasticsearch and OpenSearch provide `dense_vector` fields and k-NN search (e.g., HNSW graphs) for fast semantic and hybrid keyword/vector queries.

### Redis (with Redis Vector)
Redis includes a `vector` data type and modules that let you perform in-memory vector similarity searches, making it ideal for feature stores and low-latency inference caches.

### Apache Kafka
While not a database, Kafka is the de facto platform for real-time streaming pipelines in AI/ML workflows—ingesting, buffering, and distributing data and feature updates at scale.

### PostgreSQL + pgvector
The `pgvector` extension adds embedding storage and nearest-neighbor search capabilities to PostgreSQL, combining ACID transactions with vector similarity.


# Real-World Success Stories: Qdrant and Codeway

The session featured two Google customers—Qdrant and Codeway—who showcased how they leverage Kubernetes to power their AI and database workloads:

**Qdrant:** A vector database company, Qdrant uses GKE to deliver high-speed search and recommendation capabilities for clients like Johnson & Johnson and Disney. By combining GKE’s persistent disks with data caching, Qdrant achieves 10x faster search speeds compared to standard disks, even in worst-case scenarios without RAM caching. This performance is critical for handling billions of vectors at scale.    
**Codeway:** Codeway’s AI-driven talking head platform, integrated into their language learning app `Learner`, relies on GKE for real-time video generation. Using Gaussian splitting and optimized GPU workloads, Codeway delivers lifelike avatars with synchronized lip movements. Their in-house AI model development hub, built on GKE, abstracts infrastructure complexity, allowing researchers to focus on innovation.

These success stories highlight Kubernetes’ versatility in supporting diverse workloads, from vector databases to real-time AI applications. With KubeDB, AppsCode enables similar outcomes by providing a robust framework for managing stateful workloads, making it easier for organizations to replicate Qdrant and Codeway’s successes.

# Why Adopt Kubernetes with KubeDB?

Adopting Kubernetes for your data is a smart, future-proof decision—but running stateful workloads like databases, caches, and message queues **still demands expertise for Day 2 operations** (backup, recovery, monitoring, patching, scaling). That’s exactly where **KubeDB by AppsCode excels.**

KubeDB automates the full lifecycle of production-grade databases—PostgreSQL, MySQL, MongoDB, Elasticsearch, Redis, and more—on any Kubernetes cluster. With KubeDB, your teams get:

**Proven Adoption:** Stateful and AI workloads are now mainstream on Kubernetes, backed by survey data and real-world feedback.   
**Cost Efficiency:** Solutions like Hyperdisk, storage pools, and GKE data cache minimize costs while maximizing performance, making Kubernetes a financially savvy choice.   
**Scalability for AI:** Kubernetes’ ability to handle large-scale AI training and inference workloads ensures your infrastructure can grow with your ambitions.   
**Simplified Management with KubeDB:** KubeDB streamlines database operations on Kubernetes, offering:    
 - **Simplified Provisioning: Effortlessly deploy 20+ databases with production-ready configurations.**    
 - **Automated Day 2 Operations:** Critical tasks like **Backups, recovery, upgrades, scaling, monitoring, and alerting are expertly streamlined.**   
 - **Database Governance:** Enforce security and compliance via Kubernetes **RBAC, network policies, Secret Management, and GitOps workflows**—no new APIs or tooling required.
 - **Cloud-Native Experience:** Manage your databases using familiar Kubernetes tools and workflows (kubectl, GitOps).   
   **Ultimately, KubeDB lets your team focus on innovation, not operations.**  

As AI and data-driven applications continue to shape the future, Kubernetes offers a scalable, cost-effective, and resilient foundation. AppsCode’s KubeDB enhances this foundation by making database management effortless, empowering organizations to run production-grade databases with confidence.

# Conclusion: Embrace Kubernetes with AppsCode

> **“Kubernetes is a more than appropriate compute fabric to run your stateful applications.”**  
> — `Brian Kaufman`, Senior Product Manager, AI Infrastructure and Data @ Google Cloud

The insights from Google Cloud Next underscore a critical truth: Kubernetes is the future for stateful applications and AI workloads. With statistics showing widespread adoption of databases and AI on Kubernetes, now is the time to embrace this powerful platform. If you’re running **stateful applications like databases, Kubernetes—paired with KubeDB from AppsCode—offers the tools and flexibility to succeed.**

**Ready to take the next step?**   
Explore **KubeDB** and see how AppsCode can help you unlock the full potential of Kubernetes for your stateful and AI workloads. We’re passionate about empowering businesses to harness the full potential of Kubernetes for these critical workloads. **Visit `AppsCode.com` to learn more and start your journey today.**

# Reference
<iframe width="560" height="315" 
    src="https://www.youtube.com/embed/udQYLxsMGeU" 
    title="Data on Kubernetes: Run stateful apps and AI workloads on GKE" 
    frameborder="0" 
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" 
    allowfullscreen>
</iframe>

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To receive product announcements, follow us on [X](https://twitter.com/KubeDB).

To watch tutorials of various Production-Grade Kubernetes Tools, Subscribe to our [YouTube](https://youtube.com/@appscode) channel.