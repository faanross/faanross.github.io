---
showTableOfContents: true
title: "Traditional SIEM vs. ELK Stack Philosophy"
type: "page"
---





### 3.1 Architectural Philosophy Differences

**Traditional SIEM Approach: Pre-Built Intelligence**

Traditional SIEMs are designed around the concept of "out-of-the-box" correlation rules and built-in threat intelligence:

```
[Log Sources] → [Proprietary Collector] → [Normalization Engine] 
    → [Pre-built Correlation Rules] → [Alert] → [Pre-built Dashboard]
```

**Characteristics:**

- Heavy emphasis on preconfigured content
- "Turn it on and get value immediately" mentality
- Rule packages for compliance (PCI DSS, HIPAA, etc.)
- Vendor maintains the detection logic
- Updates come through vendor channels

**Strength:** Faster initial deployment, especially for organizations without mature security teams

**Weakness:** Less flexibility, difficult to customize, expensive to modify

**ELK Approach: Data Lake with Flexible Analysis**

ELK philosophy is fundamentally different - it's a platform for you to build on:

```
[Log Sources] → [You Configure Collection] → [You Parse/Normalize] 
    → [You Build Searches/Detections] → [You Create Dashboards]
```

**Characteristics:**

- Maximum flexibility - you control everything
- "Give me all the data, I'll decide what matters" mentality
- You build custom detection logic for your environment
- Community shares patterns, but you implement them
- Rapid iteration based on your needs

**Strength:** Ultimate flexibility, adaptable to any use case, scales your way

**Weakness:** Steeper learning curve, requires skilled staff




### 3.2 Data Model Philosophy

**Traditional SIEM: Normalized Schema**

Traditional SIEMs use a Common Event Format (CEF) or similar:

```
timestamp | source_ip | dest_ip   | source_port | dest_port | protocol | action
----------|-----------|-----------|-------------|-----------|----------|--------
2025-11...| 10.0.1.5  | 10.0.2.10 | 54321       | 443.      | TCP      | ALLOW
```

All logs are forced into predefined fields. If your log source has unique fields, they might go into "custom" fields or be lost.

**ELK: Flexible JSON Documents**

ELK stores each event as a complete JSON document:

```json
{
  "@timestamp": "2025-11-10T14:30:00Z",
  "source": {
    "ip": "10.0.1.5",
    "port": 54321,
    "bytes": 1024,
    "geo": {
      "city": "New York",
      "country": "US"
    }
  },
  "destination": {
    "ip": "10.0.2.10",
    "port": 443
  },
  "network": {
    "protocol": "tcp",
    "transport": "tcp"
  },
  "event": {
    "action": "allowed",
    "category": "network",
    "type": "connection"
  },
  "custom_field_your_siem_doesnt_support": "value",
  "nested": {
    "data": {
      "as": {
        "deep": {
          "as": {
            "you": "need"
          }
        }
      }
    }
  }
}
```

Every field is preserved, nested structures are supported, and you can add anything you want.



### 3.3 Search and Query Philosophy

**Traditional SIEM: GUI-Driven Queries**

Most traditional SIEMs provide GUI query builders:

- Click field names from dropdowns
- Select operators (equals, contains, greater than)
- Add filters one at a time
- Limited to what the GUI exposes

**Example:**

```
[Source IP] [equals] [192.168.1.100]
AND [Event Type] [equals] [Login]
AND [Result] [equals] [Failure]
AND [Time] [in last] [24 hours]
```

**ELK: Code-Like Query Language**

ELK offers multiple query languages, but the most powerful is Query DSL (Domain Specific Language):

```json
{
  "query": {
    "bool": {
      "must": [
        { "match": { "event.type": "authentication" } },
        { "match": { "event.outcome": "failure" } }
      ],
      "filter": [
        { "term": { "source.ip": "192.168.1.100" } },
        { "range": { "@timestamp": { "gte": "now-24h" } } }
      ]
    }
  }
}
```

This might appear more complex, but it offers:

- Programmatic query building
- Complex boolean logic
- Nested queries
- Aggregations and analytics within queries
- Full text search with relevance scoring
- Export and version control of queries

There are other languages that can also be used, for example **Kibana Query Language (KQL)** provides a simpler syntax for common queries:

```
source.ip: 192.168.1.100 AND event.outcome: failure AND @timestamp >= now-24h
```


### 3.4 Cost Model Philosophy

**Traditional SIEM: Pay for Consumption**

Pricing models typically include:

- **Events Per Second (EPS)**: Licensed for X,000 events/second
- **Storage-based**: Pay per GB ingested/stored
- **Hybrid**: Combination of both

**Example Scenario:** You're licensed for 10,000 EPS but suddenly need to ingest application logs (adding 5,000 EPS). Options:

1. Pay for license upgrade (expensive, immediate)
2. Sample/filter data (lose visibility)
3. Don't ingest it (blind spot)

**ELK: Pay for Infrastructure**

You pay only for:

- Servers/cloud instances
- Storage (disk space)
- Network bandwidth
- Staff time

**Same Scenario with ELK:** Need 5,000 more EPS? Add more hardware/cloud resources. No licensing negotiation, no vendor approval. The cost scales linearly with infrastructure, not artificial metrics.











---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./02_why.md" >}})
[|NEXT|]({{< ref "./04_ecosystem.md" >}})

