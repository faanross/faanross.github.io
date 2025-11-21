---
showTableOfContents: true
title: "Memory vs. Disk Usage Patterns"
type: "page"
---



## The Memory Hierarchy

Elasticsearch performance depends on understanding its memory usage:

```
Speed:    [CPU Cache] > [RAM] > [SSD] > [HDD]
Size:     [Tiny]       [GB]    [TB]    [TB]
Usage:    [Lucene]     [Heap]  [Data]  [Archive]
```

## JVM Heap Memory

Elasticsearch runs on the Java Virtual Machine (JVM), which manages memory in a "heap."

**What uses heap memory:**

- Query execution (sorting, aggregations)
- Field data cache (for sorting/aggregating)
- Node query cache (cached filter results)
- Request caches
- Indexing buffers

**Heap size rules of thumb:**

- Allocate 50% of system RAM to heap (up to 31GB)
- Never exceed 31GB heap (compressed pointers break above this)
- Leave other 50% for OS file system cache (critical for performance)

**Example for a 64GB RAM server:**

- JVM heap: 31GB
- OS file system cache: 33GB (remaining RAM)

## OS File System Cache

This is **the secret weapon** for Elasticsearch performance. The operating system caches frequently accessed file segments in RAM.

**How it works:**

1. Elasticsearch reads segment from disk
2. OS loads segment into RAM (file system cache)
3. Subsequent reads come from RAM (fast!)
4. OS automatically evicts old cached data when RAM fills

**Why this matters:**

- Hot data (recently accessed) stays in RAM
- Cold data (rarely accessed) stays on disk
- No configuration needed (OS handles it)
- Effectively gives you RAM-speed access to frequently searched data

**This is why "heap = 50% of RAM" rule exists**: You need that other 50% for file system cache. A 64GB server with 62GB heap and 2GB for OS will be slower than 31GB heap with 33GB for OS cache.

## Disk Usage

**Storage needs:**

1. **Source data**: Your original JSON documents
2. **Inverted index**: The term → document mappings
3. **Doc values**: Column-oriented storage for sorting/aggregations
4. **Stored fields**: For fast retrieval
5. **Replicas**: Multiply everything by (1 + number_of_replicas)

**Rough estimate:** Original data size × 1.5 to 2.0 = Total disk usage (including 1 replica)

Example: 100GB/day of raw logs → 150-200GB/day stored in Elasticsearch

**Compression:** Elasticsearch compresses data on disk. Text compresses well (5:1 or better ratio). But the inverted index and doc values add overhead, roughly balancing out compression gains.

## I/O Patterns

**Write pattern**: Sequential (append to segments)

- SSD is helpful but not critical
- Network-attached storage (NAS) acceptable for cold data

**Read pattern**: Random access (searching across segments)

- SSD makes a huge difference
- Local disk strongly preferred over network

**Recommendation for security operations:**

- **Hot nodes**: Local SSDs for recent data (high query rate)
- **Warm nodes**: Local HDDs acceptable (lower query rate)
- **Cold nodes**: S3/object storage (searchable snapshots)



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./13_real.md" >}})
[|NEXT|]({{< ref "./15_jvm.md" >}})

