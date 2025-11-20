---
showTableOfContents: true
title: "Pulling It All Together: The Big Picture"
type: "page"
---



Let's synthesize everything we've learned into a coherent mental model of Elasticsearch:

## The Complete Indexing Flow
1. Application sends JSON document
2. Coordinating node receives request
3. Routes to primary shard (based on hash(document_id))
4. Primary shard processes:  
   a. Analyze text fields (tokenize, lowercase, stem)   
   b. Build inverted index entries   
   c. Write to transaction log (durability)   
   d. Store in memory buffer
5. Respond to application (acknowledge)
6. Replicate to replica shards (async)
7. Refresh (every 1 second): Memory → Lucene segment
8. Document now searchable
9. Flush (every 30 min): Segment → Disk, clear transaction log




## The Complete Search Flow

1. Application sends search query
2. Coordinating node receives request
3. Broadcast query to all shards (primary or replica)
4. Each shard searches:  
   a. Parse Query DSL   
   b. Look up terms in inverted index   
   c. Score matching documents   
   d. Return top N results to coordinator
5. Coordinating node aggregates results:  
   a. Merge results from all shards   
   b. Re-score and re-sort globally   
   c. Return top N overall
6. Application receives results

## Mental Model for Performance

**Fast operations** (use inverted index):

- Term queries: `user.name.keyword: "jsmith"`
- Wildcard prefix: `domain: "evil.*"`
- Range queries: `timestamp: [now-1h TO now]`
- Filters: `status: 401`

**Slow operations** (require scoring or aggregation):

- Full-text search with scoring: `match: "suspicious activity"`
- Aggregations on high-cardinality fields: `terms: unique_ip_addresses`
- Sorting large result sets: `sort: score`

**Very slow operations** (use lots of memory):

- Nested queries over nested objects
- Parent-child relationships
- Scripted fields (computed at query time)

**Golden rule**: Use filters wherever possible (exact matches). Use queries only when you need full-text search or scoring.

## Mental Model for Scaling

**Scaling for writes** (indexing throughput):

- Add more data nodes
- Increase primary shard count (for new indices)
- Tune refresh interval (trade-off with search latency)
- Use bulk API (batch documents)

**Scaling for reads** (search performance):

- Add more replica shards (more parallel searches)
- Add more coordinating nodes (distribute query load)
- Optimize queries (use filters, limit fields returned)
- Use time-based indices (search only relevant time ranges)

**Scaling for storage**:

- Add more data nodes (distribute shards)
- Implement ILM (move old data to cheaper storage)
- Adjust retention policies (delete old data)
- Use searchable snapshots (cold tier on object storage)






---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./13_real.md" >}})
[|NEXT|]({{< ref "./14_mem.md" >}})

