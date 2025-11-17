---
showTableOfContents: true
title: "CAP Theorem and Elasticsearch"
type: "page"
---


### The CAP Theorem Explained

The **CAP theorem** states that a distributed system can provide only **two out of three** guarantees:

- **C**onsistency: All nodes see the same data at the same time
- **A**vailability: Every request receives a response (success or failure)
- **P**artition tolerance: System continues operating despite network failures

In practical terms: when a network partition occurs (nodes can't communicate), you must choose between consistency and availability.

### Elasticsearch's CAP Trade-off

Elasticsearch **prioritizes Availability and Partition Tolerance (AP)**, accepting **eventual consistency**.

**What this means:**

1. **Writes are asynchronous to replicas**:

  - Document indexed to primary shard immediately
  - Replicated to replica shards in the background
  - Brief window where primary and replicas differ
2. **Search results may reflect slightly stale data**:

  - Called "near real-time" (NRT) search
  - Default refresh interval: 1 second
  - Recent writes might not appear in searches for ~1 second
3. **During network partitions**:

  - Nodes continue accepting writes
  - Risk of conflicting updates (handled by versioning)
  - Consistency restored when partition heals

### Why This Choice Makes Sense for Security Logs

**Benefits of AP choice:**

- **High availability**: Log ingestion never stops (critical for security monitoring)
- **Performance**: No waiting for replicas before confirming writes
- **Scalability**: Easier to scale horizontally

**Acceptable trade-offs:**

- **1-second search lag**: Acceptable for most security use cases (not stock trading)
- **Eventual consistency**: Logs are immutable (not being updated), so conflicts are rare

**When to be aware of this:**

If you index a document and immediately search for it, it might not appear for ~1 second. In practice, for security operations, this delay is negligible. You're typically searching historical data (minutes to days old), not trying to catch events in real-time milliseconds.




---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./09_query.md" >}})
[|NEXT|]({{< ref "./11_read.md" >}})

