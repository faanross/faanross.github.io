---
showTableOfContents: true
title: "DNS Server"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson06_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson06_Done).


## Overview
Though some of the details will differ, we'll now do the same essential thing for DNS as we just did for HTTPS. In this lesson we'll create the server, then we'll create the DNS agent (lesson 7), and finally we'll adjust our existing runloop to make it compatible with both HTTPS and DNS.

When we're done with that we'll have all our foundational communication logic in place, which sets us up perfectly for the development of a trigger signal, parsing, and transition logic in our final chapters.


## What We'll Create
- DNS Server (`internals/dns/server_dns.go`)


## Import Library

We'll use another external library for DNS:

```bash
go get github.com/miekg/dns
```


There are a few DNS libraries in Go, but imo this one reigns supreme. It's not only simple and straight-forward to use for cases where you want to keep things high-level (and thus "outsource" a lot of the low level logic to the library), but it allows you near complete control of all aspects of DNS objects and packets.

For example in crafting DNS requests, the library will literally allow you to set every single field of the packet header except for the Z-value.

We won't jump in that deep in this workshop, but I want you to get exposure to this library since in a number of my "more advanced" DNS tools (for [example](https://github.com/faanross/spinnekop), and [here](https://github.com/faanross/dns-packet-analyzer)), as well as other workshops/courses I have planned, having such complete control over DNS packet crafting allows for tremendous opportunities in creating novel and hard-to-detect DNS covert channel communication techniques.





___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_c/03_https_loop.md" >}})
[|NEXT|]({{< ref "02_dns_agent.md" >}})