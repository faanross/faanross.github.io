---
showTableOfContents: true
title: "JSON Document Structure: The Language of Elasticsearch"
type: "page"
---

### Why JSON?

JSON (JavaScript Object Notation) is Elasticsearch's native data format. It's:

- **Human-readable**: Easy to understand and debug
- **Machine-parseable**: Efficient for computers to process
- **Ubiquitous**: Every programming language can handle JSON
- **Flexible**: Supports nested structures and arrays

### Anatomy of an Elasticsearch Document

Every document in Elasticsearch has:

1. **An index**: Where the document is stored (like a database name)
2. **A unique ID**: Identifies the document (auto-generated or specified)
3. **A _source field**: The actual JSON document you stored
4. **Metadata fields**: System fields prefixed with underscore

**Example of a complete document with metadata:**

```json
{
  "_index": "security-logs-2023.11.04",
  "_id": "abc123def456",
  "_version": 1,
  "_score": 1.0,
  "_source": {
    "@timestamp": "2023-11-04T10:17:36.000Z",
    "event": {
      "category": "authentication",
      "outcome": "failure"
    },
    "user": {
      "name": "jsmith"
    }
  }
}
```

- **_index**: The index containing this document
- **_id**: Unique identifier (used for retrieval, updates, deletes)
- **_version**: Increments with each update (for optimistic concurrency)
- **_score**: Relevance score (if from a search query)
- **_source**: Your actual data

### Field Types and Nested Structures

JSON supports multiple data types, and Elasticsearch maps these intelligently:

**Primitive types:**

```json
{
  "username": "jsmith",              // string
  "login_count": 42,                 // integer
  "success_rate": 0.95,              // float
  "is_admin": true,                  // boolean
  "last_login": "2023-11-04T10:17:36.000Z"  // date (ISO 8601 format)
}
```

**Nested objects:**

```json
{
  "user": {
    "name": "jsmith",
    "details": {
      "department": "IT",
      "clearance_level": "secret"
    }
  }
}
```

**Arrays:**

```json
{
  "tags": ["vpn", "remote-access", "suspicious"],
  "ip_addresses": ["192.168.1.10", "10.0.0.15"]
}
```

### Importance of Consistent Structure

While Elasticsearch allows schema flexibility, **consistency is crucial** for effective searching and aggregation. If authentication failures are sometimes stored as:

```json
{"result": "FAIL"}
```

and other times as:

```json
{"outcome": "failure"}
```

You'll need to search both fields, complicating queries and analysis.

This is why standards like **Elastic Common Schema (ECS)** exist - they provide a consistent field naming convention across all data sources. We'll cover ECS in detail later, but the principle is: **flexibility in what fields exist, consistency in how you name them.**


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./04_paradigm.md" >}})
[|NEXT|]({{< ref "./06_core.md" >}})

