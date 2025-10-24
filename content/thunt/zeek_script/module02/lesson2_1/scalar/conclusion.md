---
showTableOfContents: true
title: "Conclusion"
type: "page"
---

## What You've Learned

You've now worked through all eight of Zeek's fundamental scalar types. Let's step back and see how they fit together and where we're headed next.

**You've gained a complete toolkit of atomic data types for network security analysis:**



The **numeric types** (`count` and `int`) give you precise ways to track quantities, calculate differences, and implement threshold-based detection. You understand that `count` is your default for quantities - it can't be negative, making it semantically correct for counting packets, connections, attempts, and occurrences. You reach for `int` only when negative values carry meaning: deltas, offsets, and directional values.

The **network types** (`addr`, `subnet`, and `port`) transform network analysis from tedious bit manipulation into elegant, expressive code. You work with IP addresses as first-class values that understand topology. You test subnet membership in a single line. You bind ports to their protocols, preventing an entire class of bugs. These types embody deep knowledge about how networks actually work.

The **temporal types** (`time` and `interval`) unlock an entire dimension of security analysis. You track when events happen, measure how long things last, detect periodic patterns, implement timeouts and windows, and correlate activity across time. Time-based detection is what separates simple signature matching from sophisticated behavioural analysis - and you now have the tools to build both.

The **text and logic types** (`string` and `bool`) complete your foundation. Strings let you analyze application-layer protocols, detect exploits in textual content, match patterns with regular expressions, and work with the human-readable data that fills modern protocols. Booleans give you the logic to combine conditions, track state, implement flags, and express the decision-making at the heart of every detection.

## How These Types Work Together

Scalar types rarely work in isolation. Real security detections combine multiple types to answer complex questions:

```
"Has this IP address (addr) made more than 10 (count) failed login 
attempts in the last 5 minutes (interval), and is it coming from 
outside our network (subnet membership), and is the User-Agent string 
suspicious (string pattern matching), and have we not already alerted 
on this host (bool flag)?"
```

This single sophisticated detection uses six different scalar types working in concert. Each type contributes its specialized capabilities:

- `addr` identifies the host
- `count` tracks the attempts
- `interval` defines the time window
- `subnet` determines network boundaries
- `string` analyzes protocol content
- `bool` prevents duplicate alerts

**Mastering scalar types means mastering the building blocks that combine into powerful detection logic.** You don't just understand `count` - you understand how to use counts with time intervals to detect rate anomalies. You don't just understand `addr` - you understand how to use addresses with subnet membership and boolean flags to implement sophisticated allow/deny logic.

## Type Safety as a Security Tool

Throughout this section, you've seen how Zeek's type system isn't just about preventing syntax errors - it's about **preventing logical errors that could compromise your security posture**.

Using `count` instead of `int` for packet counts means you can't accidentally produce negative packets. Using `port`instead of just numbers means you can't accidentally compare TCP port 80 to UDP port 80. Using `time` and `interval`as distinct types means you can't accidentally add two timestamps when you meant to calculate a duration.

The type system is your first line of defense against bugs. Let it guide you. When Zeek complains about a type mismatch, it's usually trying to prevent you from expressing something that doesn't make semantic sense. Listen to those errors - they're teaching you to think more precisely about your data.


## Patterns You'll Use Constantly

As you move forward, you'll find yourself returning to these patterns again and again:

**Threshold Detection:** Compare counts to limits, durations to maximums, intervals to minimums

**Network Boundary Logic:** Test addresses against subnets, check port ranges, determine directionality

**Temporal Analysis:** Calculate durations, detect periodicity, implement time windows, correlate sequences

**String Analysis:** Match patterns, extract components, sanitize untrusted input, normalize case

**State Tracking:** Use booleans to remember what you've seen, prevent duplicate processing, track flags

**Type Conversion:** Move between related types when necessary (port to count, time to string, etc.)

These patterns form the vocabulary of Zeek scripting. The examples in this section aren't just illustrations - they're templates you'll adapt for your own detections.

## What's Next: Complex Types

Scalar types represent individual values. But network security analysis requires working with **collections** of values, **structured** data, and **relationships** between entities:

- How do you track failed login attempts **per IP address**?
- How do you maintain a **set** of known malicious domains?
- How do you store **multiple attributes** about a connection in a single structure?
- How do you build **mappings** from source IPs to the list of destinations they've contacted?

These questions require **complex types** - data structures that organize and relate multiple scalar values.

In Part 3, we'll explore these complex types in depth. You'll learn how to build sophisticated data structures that scale to real-world traffic, how to efficiently query and update them, and how to combine them into comprehensive security monitoring systems.




---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./bool.md" >}})
[|NEXT|]({{< ref "../complex/intro.md" >}})

