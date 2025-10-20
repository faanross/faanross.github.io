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




---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../../moc.md" >}})
[|NEXT|]({{< ref "./lesson1_2.md" >}})