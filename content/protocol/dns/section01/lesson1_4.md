---
showTableOfContents: true
title: "LESSON 1.4 - Record Types Deep Dive"
type: "page"
---


## DNS Record Types: A Deep Dive for Offensive Security

DNS record types form the vocabulary of the protocol - each defines a specific data structure for a particular purpose. While DNS was designed for legitimate name resolution, attackers have long recognized that DNS traffic is ubiquitous, rarely blocked, and capable of carrying arbitrary data.

The fundamental asymmetry that makes DNS attractive for command-and-control (C2) is this: **answers flow from server to agent**. An infected machine queries a DNS name, and the attacker-controlled authoritative nameserver responds with data encoded in resource records. The agent extracts instructions, data, or configuration from these answers. Some record types are better suited for this than others.

Let's examine each operationally relevant record type.


## A Records (Type 1, 0x01)

### Technical Description

A records map domain names to 32-bit IPv4 addresses. This is the most fundamental DNS record type - when you type a URL into your browser, an A record lookup happens.

**RDATA Format:**
```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                   IPv4 Address                |
|                   (4 octets)                  |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

The RDATA is simply four bytes representing the IPv4 address in network byte order (big-endian). For `192.0.2.53`, the bytes are `0xC0 0x00 0x02 0x35`.

**Resolution Example:**
```
Client Query:
  QNAME: www.example.com
  QTYPE: 1 (A)
  QCLASS: 1 (IN)

Authoritative Response:
  Answer Section:
    www.example.com.  300  IN  A  192.0.2.10
    
  RDATA: 0xC0 0x00 0x02 0x0A (192.0.2.10)
```

Multiple A records for the same name provide simple load balancing - clients typically use the first record or round-robin through them.

### Covert Channel Misuse

A records provide **4 bytes of data per response**. While this seems minimal, it's reliable, fast, and completely normal-looking in network traffic. Attackers use A records for compact command delivery.


### **Command Flow**




```
Agent queries: cmd1.attacker.com
Attacker NS responds: 
  cmd1.attacker.com.  60  IN  A  10.0.0.15
                                    ^^^^^^^^
                                    4 bytes = command/data

Decoding:
  Byte 1 (0x0A = 10):  Command opcode
  Byte 2 (0x00 = 0):   Sub-command
  Byte 3 (0x00 = 0):   Parameter 1
  Byte 4 (0x0F = 15):  Parameter 2
```



### **Advantages**
- Universal support - every DNS resolver and firewall expects A records
- Fast, low-overhead responses
- Can return multiple A records for multi-packet responses 
- Blends perfectly with legitimate traffic

### **Disadvantages**
- Only 4 bytes per record (low bandwidth)
- Data must fit in IPv4 address space (limits encoding schemes)
- Some networks filter private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) in public DNS responses

### **Examples**





### **Detection Considerations**
- High query volume to single domains with varying subdomains
- Queries to unusual TLDs or newly-registered domains
- A records returning addresses that are never actually contacted
- Entropy analysis of subdomain labels (DGA detection)

---


## AAAA Records (Type 28, 0x1C)

### Technical Description

AAAA records (pronounced "quad-A") map domain names to 128-bit IPv6 addresses. The format is identical to A records, just with 16 bytes instead of 4.

**RDATA Format:**
```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
|              IPv6 Address                     |
|              (16 octets)                      |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

For `2001:db8::1`, the full 16-byte representation is `20 01 0d b8 00 00 00 00 00 00 00 00 00 00 00 01`.

**Resolution Example:**
```
Client Query:
  QNAME: www.example.com
  QTYPE: 28 (AAAA)
  QCLASS: 1 (IN)

Authoritative Response:
  Answer Section:
    www.example.com.  300  IN  AAAA  2001:db8::1
    
  RDATA: 20 01 0d b8 00 00 00 00 00 00 00 00 00 00 00 01
```

### Security Implications & C2 Usage

AAAA records provide **16 bytes per response** - four times the capacity of A records. This makes them significantly more attractive for data transfer.

**Data Exfiltration/Command Flow:**
```
Agent queries: cmd.attacker.com
Attacker NS responds:
  cmd.attacker.com.  60  IN  AAAA  2001:db8:1234:5678:9abc:def0:1234:5678
                                    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                    16 bytes = substantial data

Decoding:
  16 bytes can encode:
    - 16 single-byte commands
    - 4 32-bit integers
    - 2 64-bit timestamps/IDs
    - Small JSON/binary payloads
```

**Advantages:**
- 16 bytes per record (4x more than A records)
- Still normal-looking traffic
- Can return multiple AAAA records for even more bandwidth
- Less filtering than A records (many security tools focus on IPv4)

**Disadvantages:**
- Not universally queried - if the target network is IPv4-only, AAAA queries might stand out
- Some networks don't route IPv6, making AAAA queries unusual
- Slightly larger packets (matters for UDP size limits)

**Real-World Usage:**

**AAAA records are increasingly popular in modern C2 frameworks** precisely because security tools often focus on A/TXT records. The larger payload capacity without moving to TXT (which attracts more scrutiny) makes AAAA attractive.

**Cobalt Strike** (commercial penetration testing tool, widely abused) supports DNS beaconing over AAAA records. The default configuration uses A records, but AAAA provides better throughput.

**Detection Considerations:**
- AAAA queries from IPv4-only networks (why ask for IPv6 addresses you can't use?)
- High volume of AAAA queries to single domains
- AAAA records returning addresses in non-routable IPv6 ranges (like IPv4 private ranges, but for IPv6)
- Dual-stack environments where AAAA queries spike without corresponding IPv6 traffic

---

## TXT Records (Type 16, 0x10)

### Technical Description

TXT records store arbitrary text strings, originally intended for human-readable information like SPF records, domain verification, or descriptive text. They're the most flexible record type.

**RDATA Format:**
```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|         Length        |                       |
+--+--+--+--+--+--+--+--+                       +
/               Character String                /
/                (up to 255 bytes)              /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

TXT RDATA consists of one or more length-prefixed character strings. Each string is prefixed by a single byte indicating its length (0-255), followed by that many bytes of data. A single TXT record can contain multiple strings:

```
"string1" "string2" "string3"
[0x07]string1[0x07]string2[0x07]string3
```

**Total RDATA length is limited by DNS message size** (traditionally 512 bytes for UDP, up to 4096 with EDNS0), not by the record format itself. In practice, TXT records with 1-2KB of data are common and unremarkable.

**Resolution Example:**
```
Client Query:
  QNAME: _spf.example.com
  QTYPE: 16 (TXT)
  QCLASS: 1 (IN)

Authoritative Response:
  Answer Section:
    _spf.example.com.  300  IN  TXT  "v=spf1 include:_spf.google.com ~all"
    
  RDATA: 
    [0x2B] "v=spf1 include:_spf.google.com ~all"
```

### Security Implications & C2 Usage

TXT records are the **gold standard for DNS-based data exfiltration and C2**. They provide massive bandwidth compared to A/AAAA records, support arbitrary data (not just addresses), and have legitimate use cases that make them common in enterprise networks.

**Data Exfiltration/Command Flow:**
```
Agent queries: cmd.attacker.com
Attacker NS responds:
  cmd.attacker.com.  60  IN  TXT  "483AA74B98129DDE728ABC920110005BA5D..."
                                   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                   Hundreds of bytes of encoded data

Agent parses TXT, decodes, extracts command, executes.

```

**Advantages:**
- **Massive bandwidth**: 255 bytes per string, multiple strings per record, multiple records per response
- **Arbitrary data**: Can encode Base64, hex, JSON, even compressed binary
- **Legitimate use**: SPF, DKIM, domain verification, ACME challenges - TXT records are everywhere
- **Minimal parsing**: Data comes as strings, easy to work with

**Disadvantages:**
- **High visibility**: Security teams know TXT records are abused, so monitoring is common
- **Size attracts attention**: Large TXT records or high query volumes stand out
- **Encoding overhead**: Base64 adds ~33% overhead, reduces effective bandwidth

**Real-World Usage:**

TXT records have been used by countless malware families and C2 frameworks. Here are notable examples:

**DNSMessenger** (2017, targeted financial sector): Used TXT records to download and execute PowerShell commands. Each stage of the attack was delivered via TXT records queried in sequence. Notably stealthy because it used Word documents with DDE exploits to initiate the DNS C2 channel.

**Joker (Android malware)**: A sophisticated ad-fraud malware that used TXT records to receive C2 commands and URLs for subscription fraud. Queried specific subdomains, extracted instructions from TXT responses, performed actions, then repeated.

**ScreenMate (2023)**: Korean APT campaign using legitimate-looking screensaver applications as droppers. The malware queried TXT records from attacker-controlled domains to receive multi-stage payloads encoded in Base64. Each TXT record contained portions of a larger payload, which the malware reassembled.

**Cobalt Strike**: DNS beacon mode supports TXT records for task delivery. The beacon queries the teamserver's domain, receives tasks in TXT records (often Base64-encoded), executes them, and exfiltrates results via subdomains in subsequent queries.

**OilRig (APT34)**: Iranian threat actor used TXT records in their BONDUPDATER malware. TXT records delivered commands and second-stage payloads. The use of DNS helped evade network monitoring in targeted Middle Eastern organizations.

**DNSpionage** (2018): Targeted Middle Eastern government agencies. After initial compromise, used TXT records to deliver HTTP injector code. The malware queried specific subdomains, received code in TXT responses, and injected it into browsers to steal credentials.

**SUNBURST** (SolarWinds supply chain attack, 2020): SUNBURST used DNS for initial C2. The malware generated unique subdomains containing victim information and queried them. Attacker infrastructure could respond with TXT records containing CNAME redirections to actual C2 servers, or with `A` records encoding continue/stop commands.


**Detection Considerations:**
- Query patterns: repeated queries to unusual subdomains
- TXT record size: records >512 bytes or multiple concatenated strings
- Frequency: normal TXT queries are infrequent; C2 is regular
- Historical analysis: newly-registered domains returning TXT records

**Defense:**

DNS-layer security products specifically watch for TXT record abuse. Machine learning models flag:
- Statistical anomalies in subdomain structure
- Unusual TXT record sizes
- Query timing patterns inconsistent with human behaviour
- Queries to domains with poor reputation or DGA characteristics

---


## CNAME Records (Type 5, 0x05)

### Technical Description

CNAME (Canonical Name) records create aliases. They point one domain name to another, allowing multiple names to resolve to the same destination. CNAMEs simplify DNS management - change the target once, and all aliases follow.

**RDATA Format:**
```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                 CNAME (domain name)           /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

The RDATA is a domain name in standard DNS encoding (length-prefixed labels).

**Resolution Example:**
```
Client Query:
  QNAME: www.example.com
  QTYPE: 1 (A)
  QCLASS: 1 (IN)

Authoritative Response:
  Answer Section:
    www.example.com.  300  IN  CNAME  cdn.example.net.
    cdn.example.net.  300  IN  A      192.0.2.10
```

When a resolver encounters a CNAME, it must re-query for the canonical name (unless the answer is already in the response). This allows indirection.

**CNAME Chaining:**
```
www.example.com → cdn.provider.com → lb01.provider.com → 192.0.2.10
```

DNS allows multiple CNAMEs in a chain, though excessive chaining is discouraged (performance penalty, potential loops).


### Security Implications & C2 Usage
CNAMEs can encode commands/data in the target domain name, functioning as a data channel for server-to-agent communication. While less efficient than TXT records, they attract less attention from security monitoring tools that have increasingly focused on TXT-based DNS tunnelling.

**CNAME as Data Channel (Server → Agent):**
```
Agent queries: beacon-abc123.c2.malware.com

Attacker NS responds with encoded command in CNAME target:
  beacon-abc123.c2.malware.com.  60  IN  CNAME  cmVib290LXN5c3RlbQ.cmd.attacker.com
                                                 ^^^^^^^^^^^^^^^^^^
                                                 Base64: "reboot-system"
  cmd.attacker.com.  60  IN  A  192.0.2.15

Agent parses CNAME from DNS response, extracts and decodes command
```

The malware intercepts the CNAME record from the DNS response to extract encoded data from the target domain. While resolvers follow CNAMEs automatically (which can be resolved using a wildcard), the full resolution chain is visible in the response, allowing the agent to parse intermediate records.


**Advantages:**
- **Lower detection rates**: TXT records are heavily monitored for tunnelling; CNAME usage is less scrutinized
- **Legitimate appearance**: CNAMEs are commonly used for infrastructure management, blending with normal traffic
- **Flexible infrastructure**: Can still use CNAME for legitimate redirection purposes alongside data encoding
- **Works with standard resolvers**: Doesn't require special DNS setup beyond authoritative control

**Disadvantages:**
- **Limited capacity**: ~63 bytes per label, ~253 bytes total per domain name vs ~255 bytes of arbitrary data in TXT records
- **Less efficient**: Requires more queries to transfer equivalent data compared to TXT records
- **Still detectable**: Unusual CNAME patterns (random-looking subdomains, high query volume, short TTLs) can trigger alerts


**Detection Considerations:**
- High-entropy CNAME target domains (random-looking strings)
- Unusual CNAME chain depth or frequency
- Short TTLs combined with frequently changing targets
- CNAMEs pointing to recently registered domains or known bulletproof hosting







---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./lesson1_3.md" >}})
[|NEXT|]({{< ref "./lesson1_5.md" >}})