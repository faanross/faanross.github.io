---
showTableOfContents: true
title: "C2 over DNS: Deep Dive"
type: "page"
---


## Introduction
DO LATER


## From Theory to Threat

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

## Why DNS?
Now having some understanding of where and when the concept of using DNS as a covert channel evolved let's explore why it has not only persisted for more than 2 decades, but indeed can be considered one of the "big three" C2 protocols (along with HTTP and HTTPS).

<br>

### Ubiquity and Trust
DNS is a foundational protocol of the internet, and nearly every network-aware application relies on it to function. Consequently, outbound traffic on UDP and TCP port 53 is almost universally permitted through firewalls with little question. Blocking or heavily restricting DNS would break most legitimate network functionality, making it an "all-weather" protocol for attackers.

Further, because DNS traffic is often allowed and not deeply inspected, it can be used to bypass network policies and even some proxy or firewall rules that might otherwise block other outbound communication channels (like HTTP/HTTPS or custom protocols). This is especially true in environments where egress filtering is less stringent for DNS.

<br>

### Easy to Blend in with Legitimate Traffic
C2 communication over DNS is hard to spot because it can mimic normal DNS traffic. Attackers often encode commands or exfiltrate data by embedding it in **subdomains** of a seemingly legitimate domain. They can also use various **DNS record types (like A, TXT, or CNAME records)** within the **responses** from their C2 server to send instructions back to compromised systems. This clever camouflage helps malicious activity hide within the massive amount of daily DNS traffic, making it tough for standard security tools to tell the good from the bad.

<br>

### Recursive Resolution as a Proxy
The hierarchical nature of DNS means that a C2 agent located inside a protected network does not need a direct, routable connection to the attacker's C2 server. The implant simply sends a query to its local, trusted DNS server, which then does the work of traversing the global DNS infrastructure to deliver the query to the attacker's authoritative server. This effectively turns the entire internet's DNS system into a massive, distributed, and unwitting proxy network for the attacker.

<br>

### Decentralized and Resilient
DNS is a distributed system with numerous redundant servers. This decentralized nature provides a resilient infrastructure for C2. Even if some C2 servers are taken down, C2 agents can be programmed to try other domains or IP addresses, maintaining persistence and control. Dynamic DNS (DDNS) and Domain Generation Algorithms (DGAs) further enhance this resilience by constantly changing the C2 communication endpoints.

<br>


## A Covert Channel Primer

To fully grasp the mechanics of DNS tunnelling, a foundational understanding of the specific protocol components and processes that are abused is essential. The technique's effectiveness stems from exploiting the intended functionality of DNS in unintended ways.



### The Resolution Path

The core mechanism that makes DNS C2 possible from within a restricted network is the standard DNS resolution path. One major benefit of this approach vs all other protocols is that it leverages the trusted, hierarchical nature of DNS.

Note that below I'll define the most standard and common expression of C2 over DNS, however there are variations I'll get to later.

1. **Attacker Setup:** The attacker first registers a domain (e.g., `legit-server.com`) and configures an **authoritative name server** under their control to be responsible for this domain. This server also runs the C2 server.
2. **Implant Query:** A compromised host housing the C2 agent desires sending data to the C2 server. It does this by constructing a DNS query for a specially crafted hostname, such as `[encoded-data].legit-server.com`. Notice that the encoded data is used as the subdomain.
3. **Local Resolver:** The implant sends this query to its locally configured DNS server (the "recursive resolver"). In most corporate networks, for security reasons, this will be a trusted server within the corporate network itself, like a MS AD domain controller. Meaning that all traffic from the C2 agent is typically sent to a local host, which typically attracts less scrutiny.
4. **Recursive Lookup:** The corporate resolver is not authoritative for `legit-server.com`, so it begins the recursive DNS lookup process. It queries the internet's root DNS servers, which direct it to the TLD servers for the `.com` zone. The `.com` TLD servers then inform the resolver that the authoritative name server for `legit-server.com` is the attacker's C2 server.
5. **Delivery:** The corporate resolver forwards the original query, containing the encoded data as subdomain, directly to the attacker's C2 server. The attacker has now successfully received data from an internal host that may have no direct internet access, using the organization's own DNS infrastructure as a delivery mechanism.







---
[|TOC|]({{< ref "../../guides/_index.md" >}})

