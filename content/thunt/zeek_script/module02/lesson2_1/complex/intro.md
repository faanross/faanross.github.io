---
showTableOfContents: true
title: "Introduction: Complex Types"
type: "page"
---

## Introduction: From Atoms to Structures

You've mastered Zeek's scalar types - the atomic building blocks that represent individual values: single IP addresses, individual timestamps, discrete port numbers, isolated counts. These primitives are essential, but network security analysis rarely operates on isolated data points. Real-world monitoring requires tracking **relationships**, maintaining **collections**, and organizing **hierarchies** of information. This is where complex types transform Zeek from a language that handles individual values into a platform for sophisticated stateful analysis.

## The Limitations of Scalar Types Alone

Consider a realistic security scenario: detecting SSH brute-force attacks. With only scalar types, you might attempt:

```zeek
global last_failed_ip: addr;
global last_failed_count: count;
global last_failed_time: time;
```

This approach immediately breaks down:

- **What about multiple attackers?** You can only track one IP at a time
- **How do you clean up old data?** No mechanism to expire stale tracking information
- **What if you need more attributes?** Adding fields means creating dozens of parallel global variables
- **How do you correlate related data?** No way to link the IP, count, and timestamp as a unified entity

With scalar types alone, you're forced to choose between tracking one attacker perfectly or many attackers poorly. Neither is acceptable for production security monitoring.


## The Power of Complex Types

Complex types solve these problems by providing **structured data organization** at scale:

- **tables**: Map keys to values, enabling stateful tracking of thousands or millions of entities simultaneously
- **sets**: Maintain unique collections with lightning-fast membership testing for deduplication and whitelists
- **records**: Group related fields into coherent structures with type safety and semantic clarity
- **vectors**: Preserve ordered sequences for time-series analysis and pattern detection

These are the **foundational mechanisms** that enable Zeek to perform stateful network analysis at scale. Every sophisticated detection pattern, from behavioural analytics to multi-stage attack correlation, relies on complex types to maintain context across millions of network events.


## How Complex Types Work Together

Real security detections rarely use a single complex type in isolation. The power emerges when you **combine** them:

```c
# A sophisticated threat tracking system uses ALL four types:

# RECORD: structured data (similar to struct)
type ThreatProfile: record {          
    ip: addr;
    failed_logins: count;
    scan_attempts: count;
    # SET: unique services
    services_contacted: set[port];
    # VECTOR: ordered timestamps    
    connection_times: vector of time; 
};

# TABLE: per-IP lookup
global threat_profiles: table[addr] of ThreatProfile  
    &create_expire = 24hr;
```

This single structure demonstrates:

- **table** for per-IP stateful tracking
- **record** to organize related threat indicators
- **set** to track unique services without duplication
- **vector** to maintain chronological connection history

## Critical Concepts: Memory Management

Unlike scalar types which represent single values, complex types can **grow unbounded** as they track more entities. A table tracking per-IP connection counts could grow to millions of entries if left unchecked. This makes **memory management** a central concern:

**Every production-grade complex type must have an expiration strategy.**

Throughout this section, you'll learn:

- How `&create_expire` and `&read_expire` prevent unbounded memory growth
- When to use `&max_size` to hard-limit collection sizes
- How to implement custom expiration functions for intelligent cleanup
- The performance implications of different expiration strategies
- Common patterns for balancing detection windows with memory constraints

**This isn't optional knowledge** - improper memory management will cause your Zeek deployment to run out of RAM and crash during high-volume traffic. The memory management patterns you'll learn here are as important as the data structures themselves.

## Type Safety as a Development Tool

Zeek's type system is **strongly typed** - the compiler catches type mismatches before your script ever runs. This is particularly powerful with complex types:

```c
global ip_counts: table[addr] of count;
# COMPILER ERROR: key is string, not addr
ip_counts["192.168.1.1"] = 5;  
# COMPILER ERROR: value is string, not count
ip_counts[192.168.1.1] = "5";  
```

The type system prevents entire categories of bugs:

- Can't accidentally use a string as a table key when you declared `table[addr]`
- Can't store the wrong type of value in a record field
- Can't mix incompatible types in set operations
- Can't access vector elements as if they were table keys

Let the compiler guide you. When it reports a type error, it's preventing a runtime crash or logical error. Understanding these constraints will make you a better Zeek scripter.

## How to Approach This Section

Each complex type gets a comprehensive deep dive covering:

1. **Core semantics**: What the type represents and why it exists
2. **Declaration and initialization**: How to create and populate instances
3. **Operations and patterns**: What you can do with the type and common idioms
4. **Memory management**: Expiration, size limits, and resource control
5. **Real-world examples**: Production-grade security monitoring patterns
6. **Integration patterns**: How the type combines with others

**Study the examples carefully.** The code patterns you see here aren't just illustrations - they're templates you'll adapt for your own detections. The examples progressively build in sophistication, from basic syntax to complete monitoring solutions.

**Practice the mental model.** As you read, continuously ask: "Which complex type would I use for this problem? Why? How would I combine types to solve this completely?" This active engagement transforms reading into mastery.

**Don't rush.** Complex types are where Zeek's power lives. Investing time here pays dividends in every script you write. When you finish this section, you'll understand how to build scalable, memory-safe, production-grade network security monitoring systems.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "../scalar/conclusion.md" >}})
[|NEXT|]({{< ref "./table.md" >}})

