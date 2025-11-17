---
showTableOfContents: true
title: "Read/Write Consistency Models"
type: "page"
---



## Write Consistency

When you index a document, Elasticsearch must decide: how many shard copies should acknowledge the write before responding to the client?

**Consistency levels:**

1. **One** (default): Only primary shard must acknowledge

    - Fastest writes
    - Risk: Data loss if primary fails before replication
2. **All**: Primary and all replicas must acknowledge

    - Slowest writes
    - Maximum durability
3. **Quorum**: Majority of shard copies must acknowledge

    - Balance between speed and durability

**For security logs**, the default (one) is usually fine because:

- High write throughput is crucial
- Replicas are replicated quickly (milliseconds)
- Logs can be resent from source if truly lost (Beats, Logstash have buffers)

## Read Consistency

When you search, Elasticsearch must decide: which shard copies to query?

**Read preference options:**

1. **Primary only**: Always query primaries

    - Most consistent
    - Doesn't leverage replicas for performance
2. **Replica-preferred**: Prefer replicas if available

    - Spreads load
    - May see slightly older data
3. **Adaptive** (default): Automatically routes to best shard

    - Elasticsearch picks fastest responding shard
    - Balances load and performance

**For security searches**, adaptive works well:

- Automatically optimizes performance
- Slight staleness (remember that 1-second refresh) is acceptable
- Load is distributed for better response times




---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./10_cap.md" >}})
[|NEXT|]({{< ref "./12_inverted.md" >}})

