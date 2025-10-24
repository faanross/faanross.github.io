---
showTableOfContents: true
title: "Introduction: Understanding Scalar Types in Zeek"
type: "page"
---

## Introduction: Understanding Scalar Types in Zeek

Before we dive into the individual types, let's establish what scalar types are and why they're foundational to everything you'll do in Zeek scripting.

### What Are Scalar Types?

Scalar types represent **single, indivisible values** - they're the atomic building blocks of data in Zeek. Unlike complex types (which we'll cover in Part 3) that group multiple values together into structures like tables, sets, or records, scalar types represent one thing: one number, one IP address, one moment in time, one piece of text, one true/false value.

Think of scalar types as the individual LEGO bricks. You need them before you can build anything larger. A connection record (complex type) contains scalar fields: source IP address, destination port, start time, byte count. A table (complex type) maps scalar keys to scalar values. Everything in Zeek ultimately breaks down into these fundamental scalar types.


### Why Scalar Types Matter in Network Security

Network security analysis is fundamentally about working with concrete, specific values:

- **Counting things**: How many packets? How many failed attempts? How many connections?
- **Identifying entities**: Which IP address? Which port? Which host?
- **Tracking when**: When did it start? How long did it last? What's the time between events?
- **Examining content**: What's the hostname? What's the User-Agent? What's the query string?
- **Making decisions**: Is this suspicious? Did it exceed the threshold? Should we alert?

Each of these questions relies on scalar types. You can't analyze network traffic without precise ways to represent numbers, addresses, durations, text, and truth values. Zeek's scalar types aren't just convenience wrappers around primitive values - they're **semantically rich types** designed specifically for network security work.

For example, Zeek doesn't just have "numbers" - it has `count` (non-negative integers for counting things that can't be negative) and `int` (signed integers for deltas and differences). It doesn't just store "text" - it has a `string` type with powerful pattern-matching capabilities. It doesn't represent IP addresses as strings - it has an `addr` type that understands network topology and can test subnet membership in a single operation.

This semantic richness prevents bugs and makes your code more maintainable. When you see a variable declared as `count`, you immediately know it's tracking a quantity that starts at zero and increases. When you see `addr`, you know you can use all the network-aware operations without manual bit manipulation. The type system guides you toward correct code.

### The Scalar Types You'll Master

In this section, we'll cover eight fundamental scalar types that appear in virtually every Zeek script:

**Numeric Types:**

- **count**: Non-negative integers for counting things (0, 1, 2, ...)
- **int**: Signed integers for differences and relative values (..., -2, -1, 0, 1, 2, ...)

**Network Types:**

- **addr**: IP addresses (both IPv4 and IPv6) with network-aware operations
- **subnet**: Network ranges in CIDR notation with membership testing
- **port**: Network ports bound to their transport protocol (TCP/UDP)

**Temporal Types:**

- **time**: Absolute timestamps marking specific moments
- **interval**: Durations measuring lengths of time

**Other Essential Types:**

- **string**: Text data from protocols, logs, and user input
- **bool**: Boolean truth values (`T` or `F`) for logic and decisions

Each type gets its own deep dive where we'll explore its characteristics, operations, practical security use cases, and common patterns. By the end of this section, you'll understand not just how to use these types, but **when** to use each one and **why** Zeek designed them this way.

### How to Approach This Section

As you work through each type:

1. **Understand the semantic purpose**: Why does this type exist? What real-world concept does it represent?
2. **Learn the operations**: What can you do with values of this type? How do you compare, manipulate, and combine them?
3. **Study the security examples**: How do real detections use this type? What patterns appear repeatedly?
4. **Practice in your mind**: As you read, think about how you'd use each type for detections you want to build.

Don't rush through this section. These scalar types are the vocabulary of Zeek - every script you write will use them extensively. Mastering them now will make everything that follows easier and more intuitive.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "../type.md" >}})
[|NEXT|]({{< ref "./count.md" >}})

