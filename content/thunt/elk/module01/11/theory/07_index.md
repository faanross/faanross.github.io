---
showTableOfContents: true
title: "Index Lifecycle Management Philosophy"
type: "page"
---




### The Problem: Infinite Growth

Without management, indices grow forever. This leads to:

- Exploding storage costs
- Degraded performance (searching old data you rarely need)
- Compliance violations (retaining data longer than legally allowed)

### The Solution: Automated Lifecycle Management

**Index Lifecycle Management (ILM)** automates the process of managing indices as they age:

**Typical lifecycle phases:**

1. **Hot**: Active indexing and frequent searching
    - Latest data
    - High-performance storage (SSD)
    - All queries hit this phase
2. **Warm**: No new writes, less frequent searches
    - Last 30-90 days
    - Can use slower storage
    - Optimize for storage efficiency
3. **Cold**: Rarely searched, archival
    - Historical data (90 days to several years)
    - Cheapest storage
    - Searchable snapshots on object storage
4. **Delete**: Retention period expired
    - Automatic deletion
    - Compliance requirement

**Example policy (in concept, actual implementation in practical section):**

```
Security logs policy:
- Hot phase: 7 days (active investigations)
- Warm phase: 90 days (forensic lookback)
- Cold phase: 1 year (compliance requirement)
- Delete: After 1 year
```

**Why this matters for threat hunting:**

Recent data (hot phase) gets searched constantly - investigating alerts, active hunts. You need blazing speed. Historical data (cold phase) gets searched rarely - maybe during incident response or compliance audits. Slower access is acceptable if it saves 90% on storage costs.

ILM lets you optimize for both without manual intervention. Elasticsearch automatically moves indices through phases based on age or size criteria.





---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./06_core.md" >}})
[|NEXT|]({{< ref "./08_rest.md" >}})

