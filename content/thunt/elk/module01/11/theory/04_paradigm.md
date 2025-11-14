---
showTableOfContents: true
title: "Document-Oriented vs. Relational Databases: A Paradigm Shift"
type: "page"
---

### The Relational Model: Tables, Rows, and Joins

Traditional relational databases organize data into **tables** with fixed **schemas** (column definitions). Related data lives in separate tables, connected through **foreign keys** and **joins**.

**Example: A traditional RDBMS design for security logs**

```
Table: authentication_events
+----+---------+------------+--------+---------+
| id | user_id | timestamp  | result | host_id |
+----+---------+------------+--------+---------+
| 1  | 42      | 1699123456 | FAIL   | 15      |
| 2  | 43      | 1699123457 | PASS   | 16      |
+----+---------+------------+--------+---------+

Table: users
+----+---------+------------+
| id | username| department |
+----+---------+------------+
| 42 | jsmith  | IT         |
| 43 | bjones  | Finance    |
+----+---------+------------+

Table: hosts
+----+----------------+----------+
| id | hostname       | location |
+----+----------------+----------+
| 15 | web-server-01  | DMZ      |
| 16 | db-server-01   | Internal |
+----+----------------+----------+
```

To get meaningful information, you'd need to **JOIN** these tables:

```sql
SELECT users.username, authentication_events.result, hosts.hostname
FROM authentication_events
JOIN users ON authentication_events.user_id = users.id
JOIN hosts ON authentication_events.host_id = hosts.id
WHERE authentication_events.result = 'FAIL';
```

**Problems with this approach for security logs:**

- **Joins are expensive**: Especially across billions of rows
- **Schema rigidity**: Adding a new field requires altering table structure
- **Scaling challenges**: Distributing joins across multiple servers is complex
- **Not natural for logs**: Each log entry is self-contained; why split it up?

### The Document-Oriented Model: Self-Contained JSON Objects

Elasticsearch takes a fundamentally different approach. Each event is a complete, self-contained **document** stored as JSON.

**Example: The same authentication event in Elasticsearch**

```json
{
  "@timestamp": "2023-11-04T10:17:36.000Z",
  "event": {
    "category": "authentication",
    "outcome": "failure"
  },
  "user": {
    "name": "jsmith",
    "department": "IT",
    "id": "42"
  },
  "host": {
    "name": "web-server-01",
    "location": "DMZ",
    "id": "15"
  },
  "source": {
    "ip": "192.168.1.105",
    "geo": {
      "country": "US",
      "city": "New York"
    }
  },
  "message": "Failed authentication attempt for user jsmith from 192.168.1.105"
}
```

**Advantages of this model:**

- **No joins needed**: All data is in one place
- **Schema flexibility**: Different documents can have different fields
- **Naturally denormalized**: Optimized for read performance
- **Self-descriptive**: Each document tells its complete story
- **Easy to scale**: Documents can be distributed independently

**Trade-offs:**

- **Data duplication**: User information is repeated in every authentication event (this is intentional!)
- **Update complexity**: Changing user information doesn't retroactively update old logs (usually fine for time-series data)
- **Storage overhead**: More storage used due to duplication (but storage is cheap, speed is valuable)

### When to Use Each Model

**Use relational databases for:**

- Transactional systems requiring ACID guarantees
- Data with complex relationships requiring consistency
- Frequently updated records
- Scenarios where data normalization is crucial

**Use Elasticsearch for:**

- Log aggregation and analysis
- Full-text search
- Time-series data
- Read-heavy workloads
- Scenarios requiring horizontal scalability
- Analytics and metrics

For security operations, Elasticsearch's document model is superior because **logs are immutable**. Once an authentication event happens, it happened. We don't update it; we just search, analyze, and correlate it with other events.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./03_apache.md" >}})
[|NEXT|]({{< ref "./05_json.md" >}})

