---
showTableOfContents: true
title: "Knowledge Validation"
type: "page"
---


## Knowledge Validation

### Multiple Choice Questions

**Question 1:** 
What is the primary advantage of Elasticsearch's document-oriented model over relational databases for security logs?

A) Documents can be updated more efficiently  
B) Documents are self-contained, eliminating the need for joins  
C) Documents enforce referential integrity  
D) Documents require less storage space

**Answer:**
B - Documents are self-contained, eliminating the need for joins. This is critical for performance when searching billions of log entries, as joins are expensive operations that don't scale well. Security logs are typically immutable and self-describing, making the document model ideal.


---


**Question 2:** If you have a 64GB RAM server, what is the recommended JVM heap size for Elasticsearch?

A) 32GB  
B) 64GB  
C) 31GB  
D) 16GB



---




**Answer:** C - 31GB. This follows two rules: (1) Never exceed 31GB to maintain compressed pointers, and (2) Leave approximately 50% of RAM for the OS file system cache, which is critical for performance.

**Question 3:** What does "near real-time" search mean in Elasticsearch?

A) Search results are always up-to-the-millisecond accurate  
B) There is typically a 1-second delay before indexed documents become searchable  
C) Searches take near-zero time to execute  
D) Real-time searches are not possible in Elasticsearch

**Answer:** B - There is typically a 1-second delay before indexed documents become searchable. This is due to the refresh interval, where documents are written to memory buffers and then flushed to Lucene segments every second by default.



---


**Question 4:** In the CAP theorem, which two guarantees does Elasticsearch prioritize?

A) Consistency and Availability  
B) Consistency and Partition Tolerance  
C) Availability and Partition Tolerance  
D) Elasticsearch provides all three guarantees

**Answer:** C - Availability and Partition Tolerance. Elasticsearch accepts eventual consistency to ensure high availability and continued operation during network partitions. This makes sense for log aggregation where a 1-second consistency lag is acceptable.



---



**Question 5:** What is an inverted index?

A) A data structure that maps documents to their contents  
B) A data structure that maps terms to documents containing those terms  
C) An index stored in reverse chronological order  
D) A backup index for disaster recovery

**Answer:** B - A data structure that maps terms to documents containing those terms. This is the fundamental data structure that enables fast full-text search. Instead of scanning every document for a term (slow), Elasticsearch looks up the term and instantly retrieves matching document IDs (fast).




---




**Question 6:** What is the primary purpose of replica shards?

A) To split data across multiple nodes  
B) To provide fault tolerance and increase read throughput  
C) To store backup copies for disaster recovery  
D) To enable faster writes

**Answer:** B - To provide fault tolerance and increase read throughput. Replicas serve two purposes: if a node with a primary shard fails, a replica can be promoted; and searches can be distributed across both primary and replica shards to handle higher query loads.



---



**Question 7:** Why can't you change the number of primary shards for an existing index?

A) It would require too much disk space  
B) Elasticsearch uses a hash function based on primary shard count for document routing  
C) Primary shards are immutable by design  
D) It would break replica synchronization

**Answer:** B - Elasticsearch uses a hash function based on primary shard count for document routing. The formula `shard_number = hash(document_id) % number_of_primary_shards` determines which shard stores each document. Changing the denominator would break this mapping, making existing documents impossible to locate.





---

**Question 8:** What is the difference between query context and filter context?

A) Query context is faster but less accurate  
B) Filter context calculates relevance scores; query context does not  
C) Query context calculates relevance scores; filter context does boolean yes/no matching  
D) There is no difference; they are interchangeable

**Answer:** C - Query context calculates relevance scores; filter context does boolean yes/no matching. Filters are used for exact matches, ranges, and existence checks where you only care if a document matches (yes/no), not how well it matches. Filters are faster and cacheable.

---

**Question 9:** In a 3-node cluster with an index configured for 3 primary shards and 1 replica, how many total shards exist?

A) 3  
B) 4  
C) 6  
D) 9

**Answer:** C - 6. You have 3 primary shards plus 1 replica of each primary (3 replica shards), totaling 6 shards. These would typically be distributed across nodes to ensure no single node holds both a primary and its replica.

---

**Question 10:** What is the recommended primary shard size range?

A) 1-5GB  
B) 20-50GB  
C) 100-200GB  
D) 500GB+

**Answer:** B - 20-50GB. This is a rule of thumb that balances performance and manageability. Shards that are too small create overhead; shards that are too large become difficult to move/recover and may cause performance issues.

---

### True/False Questions

**Question 11:** Elasticsearch can only store structured data with predefined schemas.

**Answer:** False. Elasticsearch is schema-less by default, allowing documents with different fields to coexist in the same index. While you can define explicit mappings (which is recommended for production), Elasticsearch will automatically detect and map fields if you don't define them.

---





**Question 12:** The OS file system cache is just as important as JVM heap memory for Elasticsearch performance.

**Answer:** True. The file system cache allows the OS to keep frequently accessed Lucene segments in RAM, providing RAM-speed access without consuming JVM heap. This is why you should only allocate ~50% of system RAM to heap - the other 50% is critical for file system cache.

---

**Question 13:** When you index a document in Elasticsearch, it is immediately searchable.

**Answer:** False. Documents are searchable after the next refresh operation (default: every 1 second). Documents are first written to an in-memory buffer and transaction log, then made searchable during the refresh when the buffer is flushed to a Lucene segment.

---

**Question 14:** Increasing the number of replica shards will improve write throughput.

**Answer:** False. Increasing replicas actually decreases write throughput because every write must be replicated to all replica shards. However, it does improve read throughput (search performance) and fault tolerance.

---

**Question 15:** Index Lifecycle Management (ILM) can help reduce storage costs by automatically moving old data to cheaper storage tiers.

**Answer:** True. ILM automates the process of transitioning indices through phases (hot → warm → cold → delete), allowing you to optimize costs by using faster, expensive storage for recent data and slower, cheaper storage for older data.

---

### Short Answer Questions

**Question 16:** Explain in your own words why Elasticsearch uses a document-oriented model instead of a relational model for log data.

**Sample Answer:** Logs are naturally self-contained events that don't require the relationships and referential integrity that relational databases provide. Each log entry (document) contains all relevant information - who, what, when, where - without needing to join across multiple tables. This allows Elasticsearch to distribute documents independently across shards and search them in parallel without expensive join operations. Additionally, different log types can have different fields without requiring schema changes, providing the flexibility needed in heterogeneous environments.

---

**Question 17:** If you have 1TB of data and want optimal shard sizing, approximately how many primary shards should you create?

**Sample Answer:** Using the rule of thumb of 20-50GB per shard, 1TB should be split into approximately 20-50 primary shards. For example, 25 primary shards would yield ~40GB per shard, which is right in the middle of the recommended range. The exact number would depend on factors like query patterns, hardware, and how the data will grow over time.

---

**Question 18:** Describe the journey of a document from when it's indexed until it's searchable and durable on disk.

**Sample Answer:** When a document is indexed: (1) It's sent to a coordinating node, which routes it to the appropriate primary shard. (2) The primary shard writes it to an in-memory buffer and appends it to the transaction log on disk for durability. (3) The node responds to the client that the write succeeded. (4) The document is asynchronously replicated to replica shards. (5) During the next refresh (default: 1 second), the memory buffer is flushed to a new Lucene segment, making the document searchable. (6) Eventually (default: every 30 minutes or when transaction log is full), segments are fsynced to disk and the transaction log is cleared, making the document durable on disk.

---

**Question 19:** Why is the inverted index data structure critical to Elasticsearch's performance?

**Sample Answer:** The inverted index maps terms to documents (rather than documents to terms), enabling constant-time lookups regardless of dataset size. When searching for a term like "login," Elasticsearch can instantly retrieve the list of documents containing that term by looking it up in the index, rather than scanning every document. This structure also efficiently supports boolean operations (AND, OR, NOT), phrase queries (by storing term positions), and relevance scoring (by storing term frequencies and other statistics). Without the inverted index, searches would require full document scans and would become prohibitively slow as data volume grows.

---

**Question 20:** Explain the relationship between primary shards, replica shards, and cluster fault tolerance.

**Sample Answer:** Primary shards hold the authoritative data, while replica shards hold copies for redundancy. If you have 3 primary shards with 1 replica each, you have 6 total shards. Elasticsearch distributes these across nodes ensuring that a primary and its replica are never on the same node. If a node fails, any primary shards on that node can be immediately promoted from their replicas on other nodes, maintaining data availability. The cluster can tolerate the loss of nodes as long as at least one copy (primary or replica) of each shard survives. More replicas mean greater fault tolerance but also higher storage costs and write overhead.

---

### Scenario-Based Questions

**Question 21:** You have a 3-node Elasticsearch cluster. Each node has 32GB of RAM. You create an index with 3 primary shards and 1 replica. One of your nodes fails. What happens to the cluster health and why?

**Sample Answer:** The cluster health will turn yellow (not red). Here's why: With 3 primary shards and 1 replica, you have 6 shards total normally distributed across 3 nodes. When one node fails, its shards are lost. However, because replicas exist on the other nodes, all primary data remains accessible - either the primary shard survived on another node, or its replica gets promoted to primary. The cluster is yellow (not green) because some replicas are now missing - you have all primaries but not all replicas. The cluster is still fully functional for reads and writes, just with reduced redundancy until the failed node recovers or new replicas are allocated.

---

**Question 22:** You notice that searches are taking longer than expected. You check and find that your JVM heap is consistently at 95% utilization with frequent old generation garbage collections taking 5-10 seconds. What are two things you should investigate and potentially fix?

**Sample Answer:**

1. **Heap may be too small**: If queries legitimately need this much memory, you should increase heap size (if you haven't hit the 31GB limit) or add more nodes to distribute the query load.

2. **Memory-intensive queries or high cardinality aggregations**: Review recent queries to identify expensive operations. Common culprits include aggregations on high-cardinality fields (like unique IP addresses), large result sets, or poorly optimized queries. Solutions include query optimization (use filters instead of queries, limit result size), using doc values for aggregations, or adding more replicas to distribute search load.


Additionally, you should verify that heap is set correctly (50% of RAM, max 31GB) and that you haven't disabled swap or otherwise starved the OS of memory for file system cache.

---

**Question 23:** Your security team wants logs to be searchable immediately (within 100ms of indexing). You set the refresh interval to 100ms, but now you're experiencing performance degradation. Explain why this is happening and propose an alternative solution.

**Sample Answer:** Setting refresh to 100ms means Elasticsearch creates new searchable segments 10 times per second instead of once per second, causing:

- 10x more CPU usage for refresh operations
- Many small segments requiring frequent merging (I/O intensive)
- Higher overall system load reducing write throughput

**Alternative solutions:**

1. **Rethink the requirement**: True real-time (<100ms) search is rarely necessary for security operations. Even active alerts query continuously and will catch events within 1-2 seconds with default refresh.

2. **Use the get API for specific documents**: If you need to immediately verify that a specific document was indexed, use `GET /index/_doc/{id}` which doesn't require refresh.

3. **Refresh on-demand**: For critical events only, trigger manual refresh: `POST /index/_refresh` after indexing.

4. **Add specialized "real-time" index**: Create a small index with fast refresh for recent events only, with a separate regular index for historical data.


The best solution is usually #1 - challenge the requirement. The 1-second default exists because it's an excellent balance for nearly all use cases.

---

## Knowledge Check Checklist

Before moving to the next module, ensure you can confidently answer "yes" to each of these:

### Core Concepts

- [ ] I can explain what Elasticsearch is and how it differs from traditional databases
- [ ] I understand why Elasticsearch is built on Apache Lucene and what Lucene provides
- [ ] I can explain the advantages and trade-offs of document-oriented vs. relational data models
- [ ] I understand JSON structure and why it's Elasticsearch's native format
- [ ] I can describe what nodes, clusters, shards, and replicas are and how they relate

### Architecture

- [ ] I understand the different node roles (master, data, ingest, coordinating)
- [ ] I can explain how shards enable horizontal scaling
- [ ] I know why primary shard count can't be changed after index creation
- [ ] I understand how replica shards provide fault tolerance and performance
- [ ] I can calculate total shard count given primary and replica configuration

### API and Querying

- [ ] I understand what RESTful means and why Elasticsearch uses it
- [ ] I know the basic HTTP methods (GET, POST, PUT, DELETE) and their purposes
- [ ] I understand the difference between query context and filter context
- [ ] I can explain what Query DSL is and why it's used instead of simple strings
- [ ] I know when to use filters vs. queries for optimal performance

### Performance and Consistency

- [ ] I understand the CAP theorem and Elasticsearch's AP choice
- [ ] I can explain what "near real-time" search means and why the delay exists
- [ ] I know what the inverted index is and why it makes search fast
- [ ] I understand the role of JVM heap memory and file system cache
- [ ] I can apply the heap sizing rules (50% of RAM, max 31GB)

### Index Lifecycle

- [ ] I understand what Index Lifecycle Management is and why it's important
- [ ] I can describe the typical phases (hot, warm, cold, delete)
- [ ] I know why time-based indices are standard for security logs
- [ ] I understand the trade-offs between storage cost and search performance

### Data Flow

- [ ] I can trace a document's journey from indexing to being searchable
- [ ] I understand the roles of memory buffer, transaction log, and segments
- [ ] I know what refresh and flush operations do
- [ ] I understand how segment merging works and why it's necessary
- [ ] I can explain the complete flow of a search query across a distributed cluster

If you can confidently check all these boxes, you're ready to move forward to the practical exercises where we'll apply these concepts hands-on. If any concepts are still unclear, review those sections before proceeding - the practical work assumes a solid understanding of these fundamentals.

---

## Conclusion

Congratulations! You've completed the theoretical foundation of Elasticsearch. You now understand:

- **What** Elasticsearch is: A distributed search and analytics engine
- **Why** it works the way it does: Document-oriented model, inverted indices, distributed architecture
- **How** to think about designing solutions: Shards for scale, replicas for redundancy, filters for performance
- **When** to use Elasticsearch: Log aggregation, security analytics, full-text search

This foundational knowledge will guide every decision you make as you build your threat hunting infrastructure. In the practical modules ahead, you'll apply these concepts, getting your hands dirty with actual Elasticsearch installations, configurations, and operations.

The theory we've covered isn't academic - it's the difference between an Elasticsearch cluster that struggles under load and one that handles billions of events with ease. It's the difference between debugging problems blindly and understanding exactly what's happening under the hood.

In the next module (1.1 Practical), you'll install Elasticsearch, start indexing real data, and see these concepts come to life. You'll create indices, configure shards and replicas, execute queries, and begin to develop the muscle memory that turns theoretical knowledge into practical expertise.

**Remember**: Elasticsearch is a tool that rewards understanding. The analyst who knows why a filter is faster than a query will write better detection rules. The engineer who understands shard allocation will design better architectures. The hunter who groks the inverted index will craft more efficient searches.

You've laid the foundation. Now let's build something incredible on top of it.









---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./13_real.md" >}})
[|NEXT|]({{< ref "./14_mem.md" >}})

