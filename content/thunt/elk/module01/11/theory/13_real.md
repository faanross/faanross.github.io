---
showTableOfContents: true
title: "How Elasticsearch Achieves Near Real-Time Search"
type: "page"
---



## The Refresh Process

When you index a document, it doesn't immediately appear in search results. Here's why:

**The indexing pipeline:**

1. **Index request arrives**: Document sent to Elasticsearch
2. **Written to memory buffer**: Stored in RAM (fast, but not durable)
3. **Written to transaction log**: Appended to log on disk (durable)
4. **Refresh** (every 1 second by default): Memory buffer written to a new Lucene segment
5. **Document now searchable**: But still only in memory
6. **Flush** (every 30 minutes or when transaction log full): Segments synced to disk (durable and searchable)

## The "Near" in Near Real-Time

**Why the 1-second delay?**

Creating a Lucene segment is expensive (building inverted index, sorting, etc.). Doing it for every document would be too slow. By batching documents and refreshing every second, Elasticsearch balances:

- Search latency (1-second staleness)
- Write throughput (bulk processing)
- Resource efficiency (fewer, larger segments)

**Can you make it faster?**

Yes, at a cost:

```json
PUT /security-logs-2023.11.04/_settings
{
  "refresh_interval": "100ms"
}
```

Now searches see documents within 100ms. But:

- More CPU usage (10x more refresh operations)
- More segments created (require merging)
- Lower write throughput

**For security operations**, the default 1 second is almost always right:

- Real-time alerting still works (alerts query continuously)
- 1 second is imperceptible for human analysts
- Optimized for sustained ingestion rate

## Segment Merging

Over time, many small segments accumulate. Elasticsearch automatically merges them:

```
Initial: [seg1:10docs] [seg2:15docs] [seg3:8docs] [seg4:12docs]
After merge: [seg5:45docs]
```

**Why merge?**

- Fewer segments = faster searches (fewer files to open)
- Delete tombstones are purged (documents aren't truly deleted until merge)
- Compression improves (larger segments compress better)

**Trade-off:** Merging is CPU and I/O intensive. Elasticsearch tunes this automatically, but you can tune merge policies for specific workloads (covered in advanced modules).



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./12_inverted.md" >}})
[|NEXT|]({{< ref "./14_mem.md" >}})

