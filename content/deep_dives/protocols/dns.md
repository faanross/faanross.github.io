---
showTableOfContents: true
title: "C2 over DNS: Deep Dive"
type: "page"
---


## Introduction
DO LATER


## Background + History

### From Theory to Threat

The conceptual underpinning of DNS tunneling is the [covert channel](https://en.wikipedia.org/wiki/Covert_channel), 
which simply implies a communication pathway that leverages a system or protocol in a manner for which it was not originally intended. This idea was first formally discussed in academic computer science as early as 1973 by Butler Lampson, who defined such channels as "those not intended for information transfer at all". For decades, this remained a largely theoretical concept confined to operating system security.

The first known public discussion of DNS as a specific vector for tunnelling data appeared in an April 1998 post by 
Oskar Pearson on the influential [Bugtraq](https://seclists.org/bugtraq/) security mailing list. 
This marked the technique's transition from a theoretical possibility to a practical idea within the nascent cybersecurity community.

<br>

![bug](../img/dns/bugtraq.webp)

<br>

The next major milestone in the technique's evolution was Dan Kaminsky's [presentation](https://www.youtube.com/watch?v=Feu6gcUf7NA) on DNS Tunnelling at Black Hat in 2004
Kaminsky demonstrated a practical implementation of tunneling IP-over-DNS, which allowed for protocols like SSH to be run entirely over DNS queries and responses. His work was pivotal because it not only provided a working proof-of-concept but also articulated the core architectural vulnerability: the ability to abuse the hierarchical and recursive nature of DNS to bypass firewalls and communicate with internal networks. (**RIP Dan Kaminsky**)

<br>

![dan](../img/dns/dan.jpg)

<br>




---
[|TOC|]({{< ref "../../guides/_index.md" >}})

