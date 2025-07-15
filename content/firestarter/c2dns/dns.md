---
showTableOfContents: true
title: "C2 over DNS: Deep Dive"
type: "page"
---





## A Covert Channel Primer

To fully grasp the mechanics of DNS tunnelling, a foundational understanding of the specific protocol components and processes that are abused is essential. The technique's effectiveness stems from exploiting the intended functionality of DNS in unintended ways.

<br>

### The Resolution Path

The core mechanism that makes DNS C2 possible from within a restricted network is the standard DNS resolution path. One major benefit of this approach vs all other protocols is that it leverages the trusted, hierarchical nature of DNS.

Note that below I'll define the most standard and common expression of C2 over DNS, however there are variations I'll get to later.

1. **Attacker Setup:** The attacker first registers a domain (e.g., `legit-server.com`) and configures an **authoritative name server** under their control to be responsible for this domain. This server also runs the C2 server.
2. **Implant Query:** A compromised host housing the C2 agent desires sending data to the C2 server. It does this by constructing a DNS query for a specially crafted hostname, such as `[encoded-data].legit-server.com`. Notice that the encoded data is used as the subdomain.
3. **Local Resolver:** The implant sends this query to its locally configured DNS server (the "recursive resolver"). In most corporate networks, for security reasons, this will be a trusted server within the corporate network itself, like a MS AD domain controller. Meaning that all traffic from the C2 agent is typically sent to a local host, which typically attracts less scrutiny.
4. **Recursive Lookup:** The corporate resolver is not authoritative for `legit-server.com`, so it begins the recursive DNS lookup process. It queries the internet's root DNS servers, which direct it to the TLD servers for the `.com` zone. The `.com` TLD servers then inform the resolver that the authoritative name server for `legit-server.com` is the attacker's C2 server.
5. **Delivery:** The corporate resolver forwards the original query, containing the encoded data as subdomain, directly to the attacker's C2 server. The attacker has now successfully received data from an internal host that may have no direct internet access, using the organization's own DNS infrastructure as a delivery mechanism.



### The Vanilla Tunnel

The most basic and earliest form of DNS tunnelling uses a simple, stateless request-response loop. This method is straightforward to implement but also the easiest to detect, for reasons we'll discuss in the next section.

1. **Upstream Transmission (Agent to Server):** The implant on the compromised host takes a piece of data to be exfiltrated (e.g., `password123`). It breaks this data into manageable chunks, encodes each chunk using a DNS-safe character set (typically hexadecimal or Base64), and prepends the result as a subdomain to the attacker's primary domain. For example, the data `password123` might be sent via a query for `cGFzc3dvcmQxMjM.c2-server.com`. To send larger files, this process is repeated for each chunk, generating a high volume of unique subdomain queries.
2. **Server-Side Reception:** The attacker's authoritative nameserver is configured not to resolve these queries in a traditional sense, but simply to log the QNAME from the incoming request. The C2 server software then parses these logs, extracts the encoded subdomains, and reassembles the original data.
3. **Downstream Transmission (Server to Agent):** To send a command back to the implant, the C2 server responds to one of the implant's queries. It crafts a DNS response, typically using a `TXT` or `CNAME` record, and places the encoded command within the record's data field (RDATA).
4. **Client-Side Reception:** The implant receives the DNS response, extracts and decodes the payload from the RDATA field, and executes the command. The entire cycle then repeats for the next command or data exfiltration task.

<br>

<br>

### The Anatomy of a DNS Packet

Now that we have a basic handle on the process, let's deconstruct the actual DNS packets in both directions to better understand how it works. The main thing to be aware of is that a DNS package provides distinct fields that can be repurposed to carry covert data.

<br>

#### Upstream (Client to Server): The Query Name (QNAME)
The primary vehicle for sending data from the implant to the C2 server is the `QNAME` field, which contains the domain name being requested. Attackers encode their data directly into this field, typically as the subdomains.

Bandwidth is constrained in this regard by protocol definition:
- A Fully Qualified Domain Name (FQDN) cannot exceed 255 bytes in total length.
- Each individual label (the text between the dots) is limited to 63 bytes.
- Allowed characters are typically alphanumeric (`a-z`, `0-9`) and the hyphen (`-`), though the hyphen cannot be at the beginning or end of a label.

<br>

#### Downstream (Server to Client): The Resource Data (RDATA)
To send commands or data back to the implant, the  C2 server sends a DNS response. The payload is typically placed in the Answer Section of the response, specifically within the RDATA (Resource Data) field of a given resource record. The capacity of this downstream channel is determined by the type of record used - see the following section.

<br>


### The Data Containers: Abusing Resource Records

The choice of DNS record type is critical, representing a trade-off between bandwidth, stealth, and reliability. From the defender's POV the record type used can often reveal the attacker's intent.

<br>

#### A and AAAA Records
These records are used to resolve a domain name to an IPv4 or IPv6 address, respectively. They offer extremely low bandwidth (4 bytes for an A record, 16 for an AAAA record). However, they are the most common query types on any network, making them very difficult to detect as anomalous. Their primary use in C2 is for "heartbeat" beacons, or conveying deterministic commands (i.e. pre-programmed), as famously demonstrated by the SUNBURST malware.

<br>

#### TXT Records
Officially designed to hold arbitrary human-readable text, TXT records are a popular choice for C2 because they are flexible and have a relatively high capacity. A single TXT record can contain multiple character-strings, each up to 255 bytes in length. While legitimate services like SPF and DKIM use TXT records, a high volume of queries for TXT records containing non-ASCII or base64-encoded data is highly unusual.


<br>


#### CNAME Records
A Canonical Name record maps an alias domain to a true domain. The target domain name field can be up to 255 bytes long and can be abused to carry encoded data back to the implant. Its use is less common than TXT but remains a viable and reasonably high-bandwidth downstream channel.

<br>

#### MX Records
Mail Exchanger records, which specify mail servers for a domain, also contain a hostname field that can be repurposed to carry a data payload. This is an uncommon but possible vector, it will however draw attention if relied on too heavily.

<br>


#### NULL Records
This record type is technically obsolete (as per RFC 1035) but is still supported by many DNS servers. Its key feature is its ability to contain up to 65,535 bytes of arbitrary binary data. This makes it by far the highest-bandwidth option and the preferred 
record type for tools like [Iodine](https://github.com/yarrick/iodine), which are designed to tunnel full IP traffic and require maximum throughput. However, because it is obsolete and has no common legitimate use, many modern firewalls and security-aware DNS resolvers will block or drop queries for`NULL` records, making it a less reliable choice in hardened environments.

<br>


### Pushing the Limits with EDNS0

The standard DNS protocol specifies that DNS messages transported over UDP should not exceed 512 bytes. If a response is larger than this, the server sets a "truncated" flag, signaling the client to retry the query over TCP. Since outbound TCP on port 53 is often blocked or more heavily inspected, this 512-byte limit is a significant bottleneck for DNS tunneling, forcing data to be fragmented across many small packets.

**Extension Mechanisms for DNS (EDNS0)**, defined in RFC 6891, provides a solution. It allows a DNS client to use pseudo-records in its query to signal to the server that it can handle UDP packets larger than the 512-byte default. A sophisticated C2 implant can leverage EDNS0 to advertise a large buffer size (e.g., 4096 bytes).

If the C2 server and all intermediate DNS resolvers support EDNS0, the server can send a much larger payload in a single UDP response. This significantly increases the downstream bandwidth, reduces the number of packets needed for a transfer, and can make the communication stealthier by reducing the overall volume of queries. The use of EDNS0 to request unusually large UDP packet sizes can itself be an indicator of compromise.

<br>

### Pushing the Limits with EDNS0

The standard DNS protocol specifies that DNS messages transported over UDP should not exceed 512 bytes. If a response is larger than this, the server sets a "truncated" flag, signaling the client to retry the query over TCP. Since outbound TCP on port 53 is often blocked or more heavily inspected, this 512-byte limit is a significant bottleneck for DNS tunneling, forcing data to be fragmented across many small packets.

**Extension Mechanisms for DNS (EDNS0)**, defined in [RFC 6891](https://datatracker.ietf.org/doc/html/rfc6891), provides a solution. It allows a DNS client to use pseudo-records in its query to signal to the server that it can handle UDP packets larger than the 512-byte default. A sophisticated C2 implant can leverage EDNS0 to advertise a large buffer size (e.g., 4096 bytes).

If the C2 server and all intermediate DNS resolvers support EDNS0, the server can send a much larger payload in a single UDP response. This significantly increases the downstream bandwidth, reduces the number of packets needed for a transfer, and can make the communication stealthier by reducing the overall volume of queries. The use of EDNS0 to request unusually large UDP packet sizes can itself be an indicator of compromise.


<br>

___


## Mechanics of DNS Command and Control

Let's discuss "vanilla" DNS tunnelling in a bit more detail now that we have a good grasp on the entire process and fundamental building blocks. Thereafter, let's discuss a few variations and nuanced applications that might not strictly conform to this mold.

<br>




### The DNS Sandwich

This novel technique, [detailed](https://blog.gigamon.com/2021/01/20/dns-c2-sandwich-a-novel-approach/) by security researcher Spencer Walden, represents a direct attempt to evade security monitoring and logging systems that are configured to only parse and record the most common and expected fields of a DNS packet. Instead of placing session data in the highly visible and frequently logged subdomain, it repurposes other, less-scrutinized fields within the DNS header and query structure itself.





<br>







---
[|TOC|]({{< ref "../../../guides/_index.md" >}})

