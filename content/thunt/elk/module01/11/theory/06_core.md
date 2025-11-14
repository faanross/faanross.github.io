---
showTableOfContents: true
title: "Core Concepts: The Architecture of Distributed Search"
type: "page"
---

Understanding Elasticsearch's architecture is essential for both performance optimization and troubleshooting. Let's build this understanding from the ground up.

### Nodes: The Building Blocks

A **node** is a single instance of Elasticsearch running on a server. Think of it as one worker in a team.

**Node roles** (a node can have multiple roles):

1. **Master-eligible node**: Can be elected as the master node
    - **Master node**: Manages cluster-wide operations (index creation, node tracking, shard allocation)
    - Only one master at a time; others are on standby
2. **Data node**: Stores data and executes search queries
    - The workhorses of your cluster
    - Need substantial disk space and RAM
3. **Ingest node**: Preprocesses documents before indexing
    - Applies transformations, enrichment
    - Like a mini-Logstash inside Elasticsearch
4. **Coordinating node**: Routes requests and aggregates results
    - Every node can coordinate, but you can have dedicated coordinators
    - Load balancer for your cluster

**Why multiple roles matter:**

In a small lab, one node might have all roles. In production, you separate concerns:

- Master nodes handle cluster management (low resource needs, high availability priority)
- Data nodes handle storage and search (high disk, RAM, CPU needs)
- Coordinating nodes handle client requests (protect data nodes from query load)

### Clusters: Coordinated Groups of Nodes

A **cluster** is a collection of nodes working together, sharing the same cluster name. All nodes in a cluster know about each other and can forward requests to the appropriate node.

**Key cluster concepts:**

- **Cluster name**: Identifier (e.g., "security-production-elk")
- **Cluster state**: Metadata about indices, shards, nodes
    - Managed by the master node
    - Broadcast to all nodes
    - Critical for coordinated operations

**Cluster health states:**

- 🟢 **Green**: All primary and replica shards are allocated
- 🟡 **Yellow**: All primary shards allocated, but some replicas missing (data safe, redundancy reduced)
- 🔴 **Red**: Some primary shards unallocated (data loss or unavailability)



### Indices: Logical Containers for Documents

An **index** is a collection of documents with similar characteristics. Think of it as a database in the relational world, but optimized for searching.

**Index naming conventions for security logs:**

Time-based indices are standard:

```
security-logs-2023.11.04
security-logs-2023.11.05
firewall-logs-2023.11.04
sysmon-logs-2023.11.04
```

**Why time-based indices?**

- **Efficient retention management**: Delete old indices easily
- **Performance optimization**: Search only relevant time ranges
- **Index lifecycle management**: Move old indices to cheaper storage
- **Parallelization**: Spread searches across indices

**Index naming best practices:**

- Use lowercase (Elasticsearch enforces this)
- Use hyphens for readability
- Include date in sortable format (YYYY.MM.DD)
- Be descriptive: `windows-security-events-2023.11.04` not `logs-2023.11.04`

### Shards: Horizontal Scaling Units

A **shard** is a self-contained Lucene index - a fully functional subset of your data. Elasticsearch divides each index into multiple shards to distribute data and parallelize operations.

**Two types of shards:**

1. **Primary shards**: Original, authoritative copies of your data
    - Number set at index creation
    - Cannot be changed later (without reindexing)
2. **Replica shards**: Copies of primary shards
    - Provide redundancy (survive node failures)
    - Increase search throughput (more copies = more parallel searches)
    - Number can be changed dynamically

**Example configuration:**

```json
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1
  }
}
```

This creates:

- 3 primary shards (data split into 3 pieces)
- 3 replica shards (1 copy of each primary)
- Total: 6 shards across the cluster

**Shard distribution example (3-node cluster):**

```
Node 1: [Primary-0] [Replica-1] [Replica-2]
Node 2: [Primary-1] [Replica-0] [Replica-2]
Node 3: [Primary-2] [Replica-0] [Replica-1]
```

Notice: Each node has a mix of primaries and replicas. If Node 1 fails, Replicas on Nodes 2 and 3 get promoted to primaries. Data remains accessible.

**Shard sizing principles:**

- **Too few shards**: Can't distribute load effectively, limited parallelism
- **Too many shards**: Overhead from managing many small indices
- **Rule of thumb**: Aim for shards between 20-50GB
    - For 100GB index: 2-5 primary shards
    - For 1TB index: 20-50 primary shards

**Why you can't change primary shard count:**

Elasticsearch uses the following formula to route documents to shards:

```
shard_number = hash(document_id) % number_of_primary_shards
```

If you change the number of primary shards, this formula changes, and Elasticsearch won't know where to find existing documents. The only solution is to reindex - create a new index with the desired shard count and copy all data.

### Replicas: High Availability and Performance

**Replica shards** serve two crucial purposes:

1. **Fault tolerance**: If a node fails, replicas on other nodes ensure no data loss
2. **Read throughput**: Searches can be executed against replicas, distributing load

**Replica configuration trade-offs:**

```
0 replicas: Maximum write performance, zero fault tolerance
1 replica:  Good balance (industry standard)
2+ replicas: Maximum fault tolerance, higher storage cost
```

For security logs where data ingestion rate is high but storage is manageable, **1 replica** is typically the sweet spot.

**Dynamic replica adjustment:**

Unlike primary shards, you can change replica count anytime:

```json
PUT /security-logs-2023.11.04/_settings
{
  "number_of_replicas": 2
}
```

This is useful for:

- Temporarily increasing replicas during heavy search periods
- Reducing replicas before deleting an index (saves replication overhead)



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./05_json.md" >}})
[|NEXT|]({{< ref "./07_index.md" >}})

