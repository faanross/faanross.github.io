---
showTableOfContents: true
title: "LESSON 1.3 - DNS Packet Structure"
type: "page"
---


## Overall Message Structure

DNS messages - both queries and responses - share a common structure defined in RFC 1035.

It consists of five sections:

```
+---------------------+
|      Header         |  12 bytes (always present)
+---------------------+
|      Question       |  Variable length (queries being asked)
+---------------------+
|      Answer         |  Variable length (RRs answering the question)
+---------------------+
|     Authority       |  Variable length (RRs pointing toward authority)
+---------------------+
|     Additional      |  Variable length (RRs with helpful extra info)
+---------------------+
```

Only the header is fixed-size. The remaining sections contain a variable number of entries, with counts specified in the header. In a typical query, only the header and question sections are populated. The response adds answer, authority, and additional records.

## The Header: 12 Bytes of Control Information

The header is a masterclass in bit-packing efficiency:

```
     0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      ID                       |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |QR|   Opcode  |AA|TC|RD|RA|   Z   |   RCODE    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    QDCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ANCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    NSCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ARCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

### Transaction ID (16 bits)

A random identifier chosen by the client to match responses to queries. Since DNS originally ran over UDP without connection state, this ID is the only way to correlate a response with its query. The 16-bit space (65,536 values) is generally sufficient for most resolvers, though high-volume servers must be careful about ID reuse while queries are outstanding. This field is security-relevant - DNS cache poisoning attacks attempt to guess valid IDs to inject forged responses.


Also note that though this field is quite limited (16 bits), since it is under full control of the C2 agent, it could be used as a carrier of data and/or signals.

### QR - Query/Response (1 bit)

The simplest field: `0` for query, `1` for response. This single bit distinguishes whether the message is a question or an answer.

### Opcode (4 bits)

Specifies the query type. Standard queries (`QUERY`, value `0`) are by far the most common - "what are the records for this name?" Other opcodes include `IQUERY` (inverse query, now obsolete), `STATUS` (server status request, rarely used), and `NOTIFY` (RFC 1996, for zone change notification). In practice, you'll almost exclusively see opcode `0`.

### AA - Authoritative Answer (1 bit)

Set to `1` in responses from authoritative nameservers, indicating "I'm responsible for this zone and this is the definitive answer." Recursive resolvers never set this bit - they're returning cached or forwarded data, not authoritative information. This distinction matters: an authoritative answer carries more weight than a cached response that might be stale.

### TC - Truncation (1 bit)

Set to `1` when the response exceeds 512 bytes (the traditional UDP limit) and has been truncated. This signals the client to retry over TCP, which doesn't have the packet size constraint. With EDNS0 (RFC 2671), clients can advertise larger UDP buffer sizes, reducing truncation frequency. But TC remains important - if you see it set, the response is incomplete.

### RD - Recursion Desired (1 bit)

Set by the client to request recursive resolution. When `RD=1`, the client is saying "please do the full lookup for me." Stub resolvers always set this. When querying authoritative servers directly (as recursive resolvers do), `RD` is typically cleared - the querier wants an iterative response (answer or referral), not for the authoritative server to recurse on their behalf.

### RA - Recursion Available (1 bit)

Set by the server in responses to indicate "I support recursive queries." Authoritative-only servers clear this bit. Recursive resolvers set it. If a client sets `RD` but receives `RA=0`, the server won't recurse - useful for identifying server capabilities.

### Z - Reserved (3 bits)

According to RFC 1035, this field is "reserved for future use". Additionally, according to RFC it "must be 0", yet any value between 0 and 7 can effectively be set.

This field is co-opted by DNS Sandwich technique, which repurposes the Z flag (reserved bit) to signal when the last message in a stream has been received, and uses the qclass (question class) field to number messages for ordering purposes when splitting data across multiple DNS packets.

Note that with DNSSEC (RFC 4035) the Z value is shortened to a single bit in order to introduce two other values: Authentic Data (1 bit) and Checking Disabled (1 bit)


**Authentic Data (AD) (1 bit)**. When set by a validating recursive resolver, it indicates "I've validated the DNSSEC signatures on this data and it's authentic." Only meaningful in DNSSEC-aware environments.

**Checking Disabled (CD) (1 bit)**. When set in a query, it tells a validating resolver "don't perform DNSSEC validation, just give me the data." Useful for debugging or when the client wants to perform validation itself. When set in a response, indicates validation was disabled.

**SO, if DNSSEC is used:**
- Bit 1 of original Z: **AD flag**
- Bit 2 of original Z: **CD flag**
- Bit 3 of original Z: Still reserved

### RCODE - Response Code (4 bits)

The status code for the response. Critical values:

- `0` (`NOERROR`): Success, answer is in the response
- `1` (`FORMERR`): Format error, server couldn't parse the query
- `2` (`SERVFAIL`): Server failure, can't process due to internal issues
- `3` (`NXDOMAIN`): Non-existent domain, the queried name doesn't exist
- `4` (`NOTIMP`): Not implemented, server doesn't support this query type
- `5` (`REFUSED`): Server refuses to answer (policy reasons)

`NXDOMAIN` vs `NOERROR` with zero answers is important: `NXDOMAIN` means the name doesn't exist; NOERROR with no answers means the name exists but has no records of the requested type. EDNS0 extended RCODE to 12 bits for additional codes. The full 12-bit extended RCODE is formed by combining  8 upper bits from the OPT record with the original 4 lower bits from the header, allowing for more error codes beyond the basic 16 values.


### QDCOUNT (16 bits)

Number of entries in the question section. Almost always `1` in practice - DNS queries typically ask one question. The protocol theoretically supports multiple questions, but implementations don't generally handle it well, so it's avoided.

### ANCOUNT (16 bits)

Number of resource records in the answer section. `0` in queries. In responses, indicates how many RRs directly answer the question. For `www.example.com`, you might get one or more A records here.

### NSCOUNT (16 bits)

Number of RRs in the authority section. These are nameserver records pointing toward authoritative sources. In referrals, this section contains NS records for the next zone to query. In authoritative answers, it might contain SOA records for negative answers.

### ARCOUNT (16 bits)

Number of RRs in the additional section. These are "helpful" records - glue records for nameservers mentioned in the authority section, for example. If NSCOUNT points to `ns1.example.com`, ARCOUNT might include its A record to avoid an additional lookup.


## The Question Section

Each question entry has this structure:

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                     QNAME                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QTYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QCLASS                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

### QNAME (Variable Length)

The domain name being queried, encoded in a specific format. Domain names are split into labels (the parts between dots), and each label is length-prefixed:

```
www.example.com → [3]www[7]example[3]com[0]
```

Each label starts with a byte indicating its length (1-63 bytes, since length uses 6 bits), followed by that many ASCII bytes. A zero-length label terminates the name. This encoding allows parsing without knowing the string length in advance.

There's a clever compression scheme: if a name suffix appears elsewhere in the message, it can be replaced with a 2-byte pointer (indicated by the two high-order bits being `11`). For example, if `www.example.com` and `mail.example.com`both appear, `example.com` might be encoded once and referenced by pointer in the second occurrence. This saves space, especially in responses with multiple RRs from the same domain.

Maximum domain name length is 255 octets (including length bytes), though individual labels max out at 63 octets. These limits are hardcoded in DNS implementations.

### QTYPE (16 bits)

The resource record type being requested. Common values:

- `1` (A): IPv4 address
- `2` (NS): Nameserver
- `5` (CNAME): Canonical name (alias)
- `6` (SOA): Start of authority
- `15` (MX): Mail exchange
- `16` (TXT): Text record
- `28` (AAAA): IPv6 address
- `255` (ANY): All records (deprecated, often blocked)

QTYPE can also be `*` meta-queries like `252` (AXFR, zone transfer request) or `251` (IXFR, incremental zone transfer). The 16-bit space allows for extensibility - new record types can be defined as needed.

### QCLASS (16 bits)

The protocol class. Almost always `1` (IN for Internet). Other classes like `3` (CH for Chaos) exist but are rarely used in practice. The class system was designed to allow DNS to support multiple protocol families, but the Internet class dominates so completely that QCLASS is essentially vestigial. Still, it must be present and set correctly.

However, like the Z field, while QCLASS should be `1` in legitimate queries, it can technically be set to any value between 0 and 65,535 without most resolvers rejecting the packet outright. This flexibility makes it attractive for covert signalling - the DNS Sandwich technique for example exploits this by repurposing QCLASS as a sequence number to track message ordering across multiple DNS packets, demonstrating how theoretically "fixed" protocol fields can be co-opted when enforcement is lax in practice.


## The Answer Section

Answer, authority, and additional sections all contain resource records in the same format:

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                     NAME                      /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     TYPE                      |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     CLASS                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TTL                      |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                   RDLENGTH                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
    /                     RDATA                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

### NAME (Variable Length)

The domain name this record applies to, using the same encoding as QNAME. Compression is common here since multiple records often share the same name.

### TYPE (16 bits)

The RR type - same values as QTYPE, but without the meta-query types. Tells you what kind of data is in RDATA.

### CLASS (16 bits)

The RR class, matching QCLASS semantics. Again, almost always `1` (IN), but can be altered for alternative uses.

### TTL (32 bits)

Time-to-live in seconds. How long this record may be cached before it should be discarded and re-queried. Zero means "don't cache." Values range from seconds (for highly dynamic records) to days or weeks (for stable infrastructure). This 32-bit unsigned integer maxes out at ~136 years, though values beyond a few days are unusual.

The TTL creates eventual consistency: after changing a record, old values remain cached until TTLs expire. Setting aggressive TTLs (low values) allows faster propagation but increases query load on authoritative servers. Setting conservative TTLs (high values) reduces load but delays change propagation.


### RDLENGTH (16 bits)

Length of the RDATA field in octets. Since RDATA format varies by TYPE, this length field is necessary to parse the message correctly. Maximum RDATA length is 65,535 bytes, though practical DNS implementations often have lower limits.

### RDATA (Variable Length)

The actual resource record data, formatted according to TYPE:

- **A record**: 4 octets (IPv4 address)
- **AAAA record**: 16 octets (IPv6 address)
- **NS record**: Domain name (variable length)
- **CNAME record**: Domain name (variable length)
- **MX record**: 2-byte preference value + domain name
- **TXT record**: One or more length-prefixed character strings
- **SOA record**: Multiple fields including primary NS, responsible mailbox, serial number, timers

The RDATA format is type-specific, which is why RDLENGTH is necessary - parsers need to know how much data to read before moving to the next RR.

## The Authority Section

Contains RRs pointing toward authoritative nameservers. In referrals (non-authoritative answers), this section holds NS records for the next zone to query:

```
Question: www.example.com A?
Authority Section:
  example.com.  172800  IN  NS  ns1.example.com.
  example.com.  172800  IN  NS  ns2.example.com.
```

In negative answers (NXDOMAIN or NODATA), the authority section typically contains the SOA record for the zone, establishing which authority is asserting the name doesn't exist and providing TTL information for negative caching.



---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../../moc.md" >}})
[|NEXT|]({{< ref "./lesson1_2.md" >}})