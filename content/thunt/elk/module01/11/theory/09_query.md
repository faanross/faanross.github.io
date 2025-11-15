---
showTableOfContents: true
title: "Query DSL: The Language of Search"
type: "page"
---


### What is Query DSL?

**Query DSL (Domain Specific Language)** is Elasticsearch's rich, JSON-based language for expressing searches. It's flexible enough to express everything from simple keyword searches to complex boolean logic with filters, aggregations, and scoring customization.

### Two Contexts: Query vs. Filter

**Query context**: "How well does this document match?"

- Calculates relevance scores
- Results ranked by score
- Use for full-text search

**Filter context**: "Does this document match? Yes or no."

- No scoring (just true/false)
- Cached for performance
- Use for exact matches, ranges, existence checks

**Example illustrating both:**

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "message": "authentication failure"
          }
        }
      ],
      "filter": [
        {
          "term": {
            "event.outcome": "failure"
          }
        },
        {
          "range": {
            "@timestamp": {
              "gte": "2023-11-04T00:00:00",
              "lt": "2023-11-05T00:00:00"
            }
          }
        }
      ]
    }
  }
}
```

**Breaking this down:**

- **must** (query context): Find documents where "authentication failure" appears in message field
    - Documents are scored by relevance
    - Maybe "authentication" and "failure" appear in different places - still matches, lower score
- **filter** (filter context): But only include documents where:
    - `event.outcome` is exactly "failure" (yes/no check)
    - `@timestamp` is within Nov 4, 2023 (yes/no check)
    - These filters don't affect scoring, just inclusion

**When to use each:**

- **Query context** for: Searching text, relevance ranking matters
- **Filter context** for: Exact matches, ranges, existence, everything where relevance doesn't matter

**Performance tip**: Filters are faster and cacheable. Use them whenever you don't need scoring.

### Basic Query Types

**1. Match query** (full-text search):

```json
{
  "query": {
    "match": {
      "message": "failed login"
    }
  }
}
```

Finds documents where "failed" or "login" appear in message (analyzed, fuzzy).

**2. Term query** (exact match):

```json
{
  "query": {
    "term": {
      "user.name.keyword": "jsmith"
    }
  }
}
```

Finds documents where user.name.keyword is exactly "jsmith" (not analyzed, precise).

**3. Range query**:

```json
{
  "query": {
    "range": {
      "login_count": {
        "gte": 5,
        "lte": 10
      }
    }
  }
}
```

Finds documents where login_count is between 5 and 10 (inclusive).

**4. Bool query** (boolean logic):

```json
{
  "query": {
    "bool": {
      "must": [/* all must match */],
      "should": [/* at least one should match */],
      "must_not": [/* none must match */],
      "filter": [/* all must match, no scoring */]
    }
  }
}
```

### Why DSL vs. Simple Query Strings?

You might wonder: why this complex JSON structure instead of simple strings like `"user:jsmith AND status:failed"`?

**Advantages of Query DSL:**

- **Programmatic**: Easy to generate from code
- **Unambiguous**: No parsing ambiguities
- **Expressive**: Can represent complex logic clearly
- **Composable**: Build queries from reusable parts
- **Type-safe**: Field types are respected

That said, Elasticsearch _does_ support query string syntax for ad-hoc searches in Kibana. But Query DSL is what you'll use for detection rules, automation, and complex hunting queries.







---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./08_rest.md" >}})
[|NEXT|]({{< ref "./10_cap.md" >}})

