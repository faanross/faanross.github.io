---
showTableOfContents: true
title: "RESTful API Architecture: The Universal Interface"
type: "page"
---



### What is REST?

REST (Representational State Transfer) is an architectural style using standard HTTP methods. Elasticsearch's entire API is RESTful, meaning every operation is an HTTP request.

**HTTP methods Elasticsearch uses:**

- **GET**: Retrieve data (search, get document)
- **POST**: Create or update (index document, complex searches)
- **PUT**: Create or update (create index, update mapping)
- **DELETE**: Delete (remove document, delete index)
- **HEAD**: Check existence (does index exist?)

### Why RESTful Matters

**Universal accessibility**: Any tool that speaks HTTP can talk to Elasticsearch:

- Command line: `curl`
- Programming languages: Python, Go, Java, JavaScript, etc.
- GUI tools: Kibana, Postman
- Other applications: Logstash, Beats, custom integrations

**Self-documenting**: URLs describe resources:

```
GET /security-logs-2023.11.04/_doc/abc123
     │                        │      │
     index name               action document ID
```

**Stateless**: Each request contains all necessary information. No sessions to manage.

### Common API Patterns

**1. Index a document (create/update):**

```
POST /security-logs-2023.11.04/_doc
{
  "@timestamp": "2023-11-04T10:17:36.000Z",
  "user": "jsmith",
  "action": "login"
}
```

**2. Get a specific document:**

```
GET /security-logs-2023.11.04/_doc/abc123
```

**3. Search for documents:**

```
GET /security-logs-2023.11.04/_search
{
  "query": {
    "match": {
      "user": "jsmith"
    }
  }
}
```

**4. Delete a document:**

```
DELETE /security-logs-2023.11.04/_doc/abc123
```

**5. Cluster health:**

```
GET /_cluster/health
```

**The power of REST**: You can learn Elasticsearch in `curl`, then apply that knowledge in any language or tool. The concepts remain the same.








---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./07_index.md" >}})
[|NEXT|]({{< ref "./09_query.md" >}})

