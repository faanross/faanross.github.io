---
showTableOfContents: true
title: "LESSON 1.2 - DNS Resolution"
type: "page"
---

## The Hierarchical Namespace

DNS organizes names as a tree, read right-to-left with dots as delimiters. The domain `mail.example.com.` represents a path from root (`.`) through `com`, through `example`, to the leaf `mail`. Each node can contain resource records and delegate authority to child nodes.


```
                            . (root)
                            |
        +-------------------+-------------------+
        |                   |                   |
       com                 org                 net
        |                   |                   |
    +---+---+           +---+---+           +---+
    |       |           |       |           |
 example  google    wikipedia  ietf      cdn
    |
 +--+--+
 |     |
www  mail
```

The root zone sits at the apex, currently served by 13 named root server systems (A through M, though anycast distribution means hundreds of physical servers). These roots delegate to Top-Level Domain (TLD) operators - generic TLDs like `.com`, `.org`, and country-code TLDs like `.uk`, `.de`. TLD operators delegate to second-level domains, which can further delegate subdomains.

Each delegation point represents a zone boundary. A zone is a contiguous portion of the namespace under single administrative control. `example.com` might be one zone, while `engineering.example.com` could be a separate delegated zone with its own authoritative servers.


## Hierarchy of DNS Servers

![dns hierarcy](../img/hierarchy.png)

**The DNS hierarchy typically displays a three-tier structure**. At the top sit the root DNS servers, which direct queries to the appropriate TLD (Top-Level Domain) servers like .com, .org, and .net. These TLD servers then point to the authoritative nameservers that hold the actual DNS records for specific domains. When you look up a website like example.com, your query travels down this hierarchy - from root to TLD to authoritative server - until it finds the IP address you need. This distributed system ensures the internet's naming infrastructure remains scalable and resilient.


## Types of DNS Servers

The DNS ecosystem has four major server types:


**Recursive resolvers** do the heavy lifting for clients. They accept queries from your device, perform whatever lookups are necessary by querying multiple authoritative servers in sequence, cache results, and return the final answer. Your ISP runs recursive resolvers, as do public services like Google's 8.8.8.8 or Cloudflare's 1.1.1.1.

**Root servers** are a special type of authoritative server that sits at the top of the DNS hierarchy. They don't know the IP addresses of websites, but they know which TLD servers to direct queries to. There are 13 root server addresses managed by various organizations worldwide.

**TLD servers** are authoritative servers that manage top-level domains like .com, .org, and .net. They respond with referrals pointing to the authoritative nameservers for specific domains within their TLD. For example, the .com TLD servers know which nameservers are authoritative for google.com.


**Authoritative nameservers** hold the actual resource records for zones they're responsible for. They answer queries with definitive data or referrals to other nameservers. They don't perform lookups on behalf of clients - they only answer questions about zones they manage.





## Resolution: Iterative vs Recursive

The distinction between iterative and recursive resolution is fundamental to understanding DNS behaviour.

**Recursive resolution** means the queried server takes full responsibility for answering. When your laptop queries your ISP's recursive resolver for `www.example.com`, you're making a recursive query. That resolver won't send you partial answers or referrals - it will do whatever work is necessary and return the final answer or an error.

**Iterative resolution** is how recursive resolvers actually obtain answers. When querying authoritative servers, the resolver explicitly requests iterative behaviour. Instead of the authoritative server doing further lookups, it returns either the answer (if it has it) or a referral pointing to the next servers to query.


**INSERT IMAGE HERE**

**The key insight**: your stub resolver makes one recursive query and waits. The recursive resolver makes multiple iterative queries, following referrals down the hierarchy until it reaches an authoritative answer.

## Why This Design?

The recursive/iterative split creates an elegant division of labor.

**Stub resolvers are simple** - they just need to know one recursive resolver and can ask it anything.

**Recursive resolvers are complex** - they implement the full resolution algorithm, handle caching, retry logic, and load balancing across multiple authoritative servers.

Authoritative servers only answer iterative queries, which keeps them simple and fast. They never need to query other servers or maintain state about ongoing lookups. This statelessness is crucial for scalability.






## Caching and TTLs

Every resource record includes a Time-To-Live (TTL) value, specified in seconds (which we'll cover in more detail in Lesson 1.3). When a recursive resolver obtains an answer, it caches both positive answers (the A record exists with this value) and negative answers (NXDOMAIN - the name doesn't exist) according to their TTLs.

Subsequent queries for the same name hit the cache, avoiding the full iterative lookup. This dramatically reduces query load on authoritative servers and improves response times. The tradeoff is eventual consistency - changes to DNS records aren't visible to all clients until cached entries expire.

### The Caching Problem for C2 Operations

From an attacker's perspective, caching is fundamentally problematic. Since our C2 server acts as the authoritative nameserver and our agent is the DNS client, any cached response effectively breaks real-time communication between our agent and server. If the recursive resolver returns a cached answer, our agent's query never reached its target, our C2 server.

We can however set the TTL when our C2 server returns DNS responses, but there's a critical caveat: while we can request any TTL, let's say something as low as 1 second (essentially disabling caching), intermediate caching servers are free to ignore our TTL request and enforce their own minimum values. Most recursive resolvers will ignore any TTL below 300 seconds (5 minutes) and default to this value as their enforced minimum. This behaviour is a deliberate anti-abuse measure to reduce query load on authoritative servers.

### Two Approaches to the Caching Problem

This caching behaviour forces us into one of two operational strategies:

**Option 1: Low-and-Slow Beaconing**

**We accept the caching constraint and simply beacon every 5+ minutes.**

There remains some risk that certain middleboxes might cache for longer than five minutes. When this occurs, the agent's queries initially return cached responses that never reach the C2 server. However, once the cache entry expires, the next query will reach the server as a cache miss.

At this point, the server can observe the gap pattern between successful queries and infer the actual cache duration in the path. For example, if the agent is beaconing every 5 minutes but the server only receives queries every 15 minutes, the server can deduce that caching is approximately 10-15 minutes. In the next successful response, the server can instruct the agent to increase its beacon interval to 15 minutes, ensuring future queries bypass the cache. 



**Option 2: Unique Subdomain Generation**

We bypass caching entirely by generating unique subdomains for each communication. Since each request is for a different name (e.g., `request-1a2b3c.attacker.com`, `request-4d5e6f.attacker.com`), cached responses are impossible by definition - there's no prior query for that specific name to cache. This is the default approach for agent-to-server exfiltration, where encoded data naturally creates unique subdomains.

However, this approach has a significant drawback: generating an excessive volume of unique subdomains becomes a major red flag. Modern DNS security tools specifically look for this pattern - high query volumes with low or zero repetition - as an indicator of data exfiltration or DGA (Domain Generation Algorithm) activity. Each unique query increases detection risk.

### Choosing Between Strategies

The choice between these approaches depends on operational requirements:

**Low-and-slow (Option 1)** is preferred when DNS serves as a sleeper channel for long-term persistence. The agent checks in infrequently, receives occasional commands, and maintains minimal network footprint. Detection risk is lower, but operational tempo is constrained by caching intervals.

**High-bandwidth unique subdomains (Option 2)** are preferred when DNS serves as an active, high-throughput exfiltration channel. The agent must move data quickly, accepting higher detection risk in exchange for bandwidth and responsiveness.

We will examine these divergent strategies - their specific implementations, operational considerations, detection profiles, and tactical tradeoffs - in much greater depth in Section 4.




## The Resolver Cache

Modern recursive resolvers maintain sophisticated caches. A single lookup for `www.example.com` populates the cache not just with the A record, but also with NS records for `example.com`, glue records for those nameservers, and potentially the `.com` TLD servers. Future queries for `mail.example.com` skip querying the root and `.com` entirely - the resolver already knows example.com's nameservers.

This hierarchical caching is why DNS scales. The root servers handle relatively few queries because nearly every recursive resolver has `.com`'s NS records cached. Similarly, TLD servers handle mostly cache misses for less-frequently-queried domains.

## Delegation and Glue Records

Zone delegation requires NS records pointing to nameservers for the child zone. But there's a bootstrapping problem: if `example.com`'s nameservers are `ns1.example.com` and `ns2.example.com`, how do you resolve those names?

Glue records solve this. The parent zone (`.com`) includes not just NS records for `example.com`, but also A records for the nameservers themselves. This breaks the circular dependency and allows the resolver to contact the child zone's nameservers directly.

The resolution algorithm elegantly handles delegation, referrals, caching, and the hierarchical authority model - a testament to Mockapetris's original design.



---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./lesson1_1.md" >}})
[|NEXT|]({{< ref "./lesson1_3.md" >}})