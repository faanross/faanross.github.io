---
showTableOfContents: true
title: "LESSON 1.1 - History and Background of DNS"
type: "page"
---


## The Purpose of DNS

The Domain Name System is fundamentally a **distributed, hierarchical database** that translates human-readable domain names into machine-usable IP addresses. While this seems simple on the surface, DNS's design decisions - made in the 1980s for a vastly different internet - create the perfect storm of properties that make it exploitable for covert communications.

Think of DNS as a global phone book that anyone can query, where looking up a number doesn't require authentication, happens constantly in the background, and leaves minimal forensic traces. This is precisely why it's so valuable to attackers.


## Background

Before DNS, the internet had a namespace problem. In the ARPANET days of the early 1970s, hostname-to-address mapping was maintained in a single file - HOSTS.TXT - managed by the Stanford Research Institute's Network Information Center. Every site that wanted current mappings had to periodically FTP this file from SRI-NIC. As the network grew from dozens to hundreds to thousands of hosts, this approach collapsed under its own weight. The file became unwieldy, update conflicts were inevitable, and the centralized distribution model simply couldn't scale.

## The Problem Space

By the early 1980s, the limitations were glaring. HOSTS.TXT offered no hierarchical structure, making namespace collisions increasingly common. The update latency meant different hosts had inconsistent views of the namespace. Perhaps most critically, the administrative burden of managing every hostname in a single flat file was unsustainable. The internet needed a distributed, hierarchical system that could delegate authority and scale horizontally.

## Enter Paul Mockapetris

In 1983, Paul Mockapetris, then at USC's Information Sciences Institute, designed the Domain Name System. His insight was elegant: **create a hierarchical tree structure for names**, **distribute authority across the hierarchy**, and **cache aggressively to improve performance**. Rather than one organization maintaining one file, administrative control would be delegated - each domain owner would manage their own subtree.

The architecture Mockapetris proposed was fundamentally different from its predecessor. DNS would be a distributed database, with queries resolved through a recursive lookup process starting from authoritative root servers and proceeding down through the hierarchy. This design naturally accommodated growth and eliminated single points of administrative failure.

## The Core RFCs

Mockapetris formalized DNS in two foundational documents published in November 1987 (though initial versions appeared in 1983):

**RFC 1034** defines the concepts and facilities - the domain name space structure, the notion of resource records (RRs), and the resolver/server architecture. It establishes the tree structure, the concept of zones as units of delegation, and the query resolution algorithm.

**RFC 1035** provides the implementation specification - the actual wire format, message structure, and protocol mechanics. It defines how queries and responses are encoded, the structure of resource records (including the critical A, NS, CNAME, MX, and SOA types), and the behavior of both recursive and iterative resolution.

Together, these RFCs created a system of remarkable durability. While subsequent RFCs have extended DNS - adding security (DNSSEC in RFC 4033-4035), IPv6 addresses (AAAA records in RFC 3596), dynamic updates (RFC 2136), and numerous other features - the core architecture from 1987 remains intact.


## Why It Worked

DNS succeeded because it solved the right problems with the right tradeoffs. The hierarchical delegation model distributed administrative load naturally. Caching provided performance while accepting eventual consistency - a pragmatic choice that acknowledged that namespace data doesn't change rapidly enough to require strong consistency. The protocol's simplicity (UDP for queries, TCP for zone transfers and large responses) made implementation straightforward.

The resource record abstraction was also prescient. Rather than hardcoding specific data types, DNS used a flexible RR format that could accommodate new record types as needs evolved. This extensibility has proven invaluable over four decades.

## Legacy and Evolution

DNS deployment began in 1984-85, coexisting with HOSTS.TXT before fully replacing it by the late 1980s. Today, DNS is fundamental internet infrastructure, handling hundreds of billions of queries daily. While modern extensions like DNSSEC address security concerns (the original DNS had no authentication), and DoH/DoT encrypt queries for privacy, the core protocol remains Mockapetris's 1987 design.

That a protocol designed for a network of thousands still serves a network of billions speaks to its architectural soundness. DNS exemplifies successful internet protocol design: simple enough to implement widely, flexible enough to extend, and built on solid distributed systems principles.



---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../../moc.md" >}})
[|NEXT|]({{< ref "./lesson1_2.md" >}})