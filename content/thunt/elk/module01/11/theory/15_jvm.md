---
showTableOfContents: true
title: "JVM Heap Sizing Philosophy"
type: "page"
---


## Why Heap Size Matters

The JVM heap is where Elasticsearch does its work in memory. Too small, and you'll see OutOfMemory errors. Too large, and garbage collection pauses will cripple performance.

## The Golden Rules

**Rule 1: Never exceed 31GB heap**

Above ~31GB, the JVM loses "compressed pointers":

- Below 31GB: Object pointers are 4 bytes
- Above 31GB: Object pointers are 8 bytes
- Result: You can actually address less memory with a 40GB heap than a 31GB heap!

**Rule 2: Allocate 50% of system RAM (up to 31GB)**

Examples:

- 8GB system → 4GB heap
- 32GB system → 16GB heap
- 64GB system → 31GB heap (not 32GB!)
- 128GB system → 31GB heap (not 64GB!)

**Rule 3: Leave the rest for OS file system cache**

This is not "wasted" memory - it's where the magic happens:

- OS caches Lucene segments
- Recent searches hit cache (RAM speed)
- Old searches hit disk (still acceptably fast with SSD)

## Monitoring Heap Usage

**Healthy heap usage pattern:**

```
[---Used: 8GB---][---Free: 8GB---]
     (50%)              (50%)
```

**Unhealthy pattern:**

```
[----------Used: 15GB----------][Free: 1GB]
              (94%)              (6%)
```

If you're consistently using >75% of heap, you need:

- More RAM (to increase heap)
- More nodes (to distribute load)
- Query optimization (to reduce memory usage)

## Garbage Collection

The JVM periodically reclaims unused memory ("garbage collection"):

**Types:**

1. **Young generation GC**: Fast (milliseconds), frequent
2. **Old generation GC**: Slow (seconds), infrequent

**Healthy GC pattern:**

- Young GC: Every few seconds, <100ms
- Old GC: Every few hours, <1 second

**Unhealthy GC pattern:**

- Old GC: Every few minutes
- Long pauses: >5 seconds
- "GC thrashing": Spending >50% of time in GC

**Causes of GC problems:**

- Heap too small (increase it)
- Memory-intensive queries (optimize them)
- Field data cache overload (use doc values, limit cardinality)

## Heap Size Configuration

Set heap size in `jvm.options`:

```
# Good: Explicit min and max (equal values)
-Xms16g
-Xmx16g

# Bad: Different min and max (causes fragmentation)
-Xms8g
-Xmx16g
```

**Why set min = max?**

- JVM pre-allocates full heap at startup
- Prevents resize operations (disruptive)
- OS knows exactly how much RAM is committed






---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./13_real.md" >}})
[|NEXT|]({{< ref "./14_mem.md" >}})

