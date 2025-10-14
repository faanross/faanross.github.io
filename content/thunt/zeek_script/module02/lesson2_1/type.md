---
showTableOfContents: true
title: "Part 1 - Understanding Zeek's Type System"
type: "page"
---


## **Welcome to Zeek Programming**

Welcome to the moment you've been building toward - writing your first Zeek scripts. In Module 1, you learned to operate Zeek as a tool. In Module 2, you'll learn to extend Zeek as a platform. The difference is profound. Operating Zeek gives you visibility into network activity; programming Zeek gives you the power to define exactly what that visibility means, how to interpret it, and what actions to take.

Zeek's scripting language is purpose-built for network security analysis. It's not a general-purpose language like Python or JavaScript - it's specifically designed to make network security tasks natural and efficient. The language includes native types for IP addresses, network ports, time intervals, and subnets. It understands network protocols inherently. It's built around the event-driven model we explored in Lesson 1.2, making it natural to write code that responds to network activity.

This lesson introduces you to Zeek's type system, which is the foundation of everything you'll write. We'll start with simple scalar types - individual values like numbers, IP addresses, and strings - and progress to complex types like tables, sets, and records that let you build sophisticated data structures. You'll learn about variable scoping, which determines where your variables are accessible. You'll understand type inference, which lets Zeek figure out types automatically, and type casting, which lets you convert between types when needed.

The theory in this lesson might feel dense - there are many types to learn and rules to understand. But every concept we cover has direct application to threat hunting. When you're tracking which IP addresses have made connections to your network, you'll use sets and tables. When you're analyzing the timing between packets to detect beaconing, you'll use time and interval types. When you're correlating data across multiple events, you'll use records to structure related information.

This is a foundational lesson. Take your time. Type out the examples - don't just read them. Experiment. Break things intentionally to see what error messages you get. The deeper your understanding of these fundamentals, the more powerful your detection scripts will become.

Let's begin by understanding what makes Zeek's type system special.

---

## **Why Zeek Has Its Own Type System**

When you first start writing Zeek scripts, you might wonder why you can't just use familiar types from languages like Python, C, or JavaScript. The answer lies in what makes network security analysis fundamentally different from general-purpose programming.

Think about it: when you're analyzing network traffic, you're not just shuffling generic data around. You're working with **domain-specific concepts** that have their own rules, relationships, and behaviors. An IP address isn't just a string of characters - it represents a network endpoint with inherent properties like subnet membership and routing logic. A port number isn't just an integer - it's intrinsically tied to a transport protocol and has meaningful ranges and classifications.

Zeek's type system was designed from the ground up to make these network security concepts **first-class citizens** in the language. This means you can work with IP addresses, subnets, ports, and time intervals using natural, intuitive syntax rather than fighting with string parsing and manual validation.

## **Network Security Has Domain-Specific Needs**

Let's explore the specific challenges that make network security analysis different from general-purpose programming, and why Zeek's specialized types exist.


```
┌───────────────────────────────────────────────────────────────┐
│         NETWORK SECURITY TYPE REQUIREMENTS                    │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  IP Addresses are NOT just numbers or strings                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━     │
│  • IPv4: 192.168.1.100                                        │
│  • IPv6: 2001:db8::1                                          │
│  • Need subnet matching: is 192.168.1.100 in 192.168.0.0/16?  │
│  • Need comparison operators: which IP is "larger"?           │
│  • String representation vs binary representation             │
│  • Automatic handling of both IPv4 and IPv6 formats           │
│                                                               │
│  Why it matters: Without a dedicated IP type, you'd write     │
│  dozens of lines of parsing and validation code for every     │
│  subnet check or address comparison.                          │
│                                                               │
│  Ports are NOT just integers                                  │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━     │
│  • Must associate with protocol: 80/tcp vs 80/udp different   │
│  • Range: 0-65535 (not arbitrary integers)                    │
│  • Well-known vs ephemeral ports have different meanings      │
│  • Need to correlate with service expectations                │
│                                                               │
│  Why it matters: When checking if traffic is on a standard    │
│  web port, you need protocol context - 80/tcp is HTTP, but    │
│  80/udp is something else entirely.                           │
│                                                               │
│  Time is CRITICAL and complex                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━     │
│  • Timestamps: absolute points in time (when did this occur?) │
│  • Intervals: durations between events (how long did it last?)│
│  • Need arithmetic: timestamp + interval = timestamp          │
│  • Sub-second precision essential (attacks happen fast!)      │
│  • Timezone and formatting considerations                     │
│                                                               │
│  Why it matters: Detecting fast-flux DNS or brute force       │
│  attacks requires precise time calculations. Generic time     │
│  libraries lack the network-specific operations you need.     │
│                                                               │
│  Network data has inherent structure                          │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━     │
│  • Connection = src_ip + src_port + dst_ip + dst_port + ...   │
│  • DNS query = query_name + query_type + answers + ...        │
│  • HTTP request = method + uri + headers + body + ...         │
│  • Need structured types (records) not just primitives        │
│                                                               │
│  Why it matters: Network protocols are structured data.       │
│  Zeek's record types mirror this structure, making your code  │
│  self-documenting and reducing bugs.                          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```


**IP Addresses Are More Than Strings or Numbers**

At first glance, an IP address like `192.168.1.100` might seem like it could be stored as a simple string. But in practice, IP addresses have rich semantics that strings can't capture. You need to be able to ask questions like "Does this address belong to my internal network?" or "Is this IPv6 address in the same subnet as another?"

Consider subnet matching - one of the most common operations in security analysis. Determining whether `192.168.1.100` falls within `192.168.0.0/16` requires converting the address to binary, applying a subnet mask, and comparing the results. Without a dedicated IP address type, you'd need to write (and maintain) dozens of lines of parsing, validation, and comparison logic every time you need to check subnet membership.

IP addresses also need to support both IPv4 and IPv6 formats seamlessly. Modern networks use both, and your scripts shouldn't break when encountering an IPv6 address like `2001:db8::1`. You also need meaningful comparison operations - determining which address is "larger" for sorting or range checking isn't straightforward with string representations.

Zeek's `addr` type handles all of this automatically, letting you write intuitive expressions like `if (ip_addr in 192.168.0.0/16)` without worrying about the underlying complexity.

**Ports Need Protocol Context**

Port numbers might look like simple integers between 0 and 65535, but they're actually much more nuanced. The same port number means completely different things depending on the transport protocol. Port 80 over TCP is typically HTTP web traffic, but port 80 over UDP is something else entirely - possibly DNS or a custom application protocol.

When you're analyzing network security, this context matters tremendously. You need to know not just that traffic is on port 80, but that it's on port 80/tcp versus 80/udp. You also need to distinguish between well-known ports (like 80 for HTTP or 443 for HTTPS) and ephemeral high-numbered ports that clients use for outbound connections.

Generic integer types can't capture these distinctions. Zeek's `port` type, however, inherently carries protocol information and enforces valid ranges, making your security logic clearer and less error-prone.

**Time Requires Precision and Special Operations**

Network security analysis is deeply concerned with time. When did an event occur? How long did a connection last? Did multiple failed login attempts happen within a suspicious time window? These questions require two distinct concepts: **timestamps** (absolute points in time) and **intervals** (durations or spans of time).

Generic programming languages typically provide basic time libraries, but they lack the specialized operations network security demands. You need to perform arithmetic like "timestamp plus interval equals new timestamp" or "timestamp minus timestamp equals interval." You need sub-second precision because attacks and anomalies can unfold in milliseconds. You need to handle timezone considerations and format timestamps in standard ways for logging.

Detecting threats like fast-flux DNS (where domain resolution changes rapidly) or brute force attacks (many attempts in a short window) requires precise, efficient time calculations. Zeek's `time` and `interval` types make these operations natural and performant, with built-in support for the arithmetic and comparisons you need.

**Network Protocols Are Inherently Structured**

Network traffic isn't just random bytes - it's highly structured data following specific protocols. A TCP connection is defined by a source IP, source port, destination IP, destination port, and protocol. A DNS query has a query name, query type, response code, and answer records. An HTTP request contains a method, URI, headers, and possibly a body.

Trying to represent these structures with primitive types (strings, integers, etc.) leads to code that's hard to read and maintain. You end up with variables like `conn_src_ip`, `conn_src_port`, `conn_dst_ip`, and so on - a flat soup of related values with no clear relationship.

Zeek's **record types** let you mirror the natural structure of network protocols directly in your code. A connection becomes a `conn_id` record with clearly defined fields. A DNS query becomes a structured object with all its components organized logically. This makes your scripts self-documenting - anyone reading your code can immediately understand what data you're working with. It also reduces bugs because the type system enforces that all required fields are present and correctly typed.



## **A Concrete Comparison**

To see the difference in practice, consider checking if an IP address belongs to your internal network:

**In Python/JavaScript (general-purpose approach):**

```python
ip_string = "192.168.1.100"
# Need to:
# 1. Parse the string into octets
# 2. Convert to binary representation
# 3. Parse the subnet CIDR notation
# 4. Apply subnet mask
# 5. Compare the results
# This is 15-20 lines of error-prone code
```

**In Zeek (domain-specific approach):**

```zeek
local ip_addr = 192.168.1.100;
if ( ip_addr in 192.168.0.0/16 )
    # Just works - natural and readable
```

The Zeek version isn't just shorter - it's **clearer, safer, and more maintainable**. The type system handles all the complexity behind the scenes, letting you focus on the security logic rather than low-level data manipulation.


## **What You'll Learn in This Chapter**

This chapter provides a comprehensive exploration of Zeek's type system and how to use it effectively in your security scripts. We'll cover:

- **Part 2: Scalar Types** - Single values like addresses, ports, counts, and booleans that form the building blocks of your scripts
- **Part 3: Complex Types** - How to organize and manage structured data using records, tables, sets, and vectors
- **Part 4: Variable Scoping and Namespaces** - Understanding where variables are accessible and how to organize your code effectively
- **Part 5: Type Inference and Type Casting** - How Zeek determines types automatically and when you need to convert between types explicitly
- **Part 6: Operators and Expressions** - The full range of operations you can perform on different types, from arithmetic to logical comparisons
- **Part 7: Working with Zeek's Built-in Data Structures** - Leveraging Zeek's powerful collections for tracking connections, aggregating data, and detecting patterns
- **Part 8: Practical Exercises** - Hands-on problems to reinforce your understanding and build real-world skills
- **Part 9: Knowledge Validation** - Testing your comprehension of the concepts covered throughout the chapter

By the end, you'll understand not just _what_ Zeek's types are, but _why_ they exist and how they make your security analysis code more powerful, expressive, and maintainable. You'll be comfortable working with everything from simple IP addresses to complex data structures that track sophisticated attack patterns. Let's dive in.

---


[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../../module01/lesson1_3/validation.md" >}})
[|NEXT|]({{< ref "./scalar/intro.md" >}})

