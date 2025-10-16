---
showTableOfContents: true
title: "Part 2D - The int Type: Signed Integers"
type: "page"
---


## The int Type: Signed Integers

The `int` type represents signed integers - whole numbers that can be positive, negative, or zero. While the `count` type is far more common in Zeek scripts, `int` fills an important niche: it's the type to reach for when negative values are not just possible but actually meaningful in your analysis.

### When Do You Need Signed Integers?

Most quantities in network security analysis are inherently non-negative. You can't observe negative three packets or have negative seven connections. This is why `count` dominates Zeek scripts. However, certain calculations and concepts naturally produce or require negative numbers, and that's where `int` becomes essential.

Think about **differences and deltas**. If you're comparing the current byte count of a connection to a previous measurement, the difference could be negative - perhaps due to retransmissions or measurement timing. When you're tracking **relative positions or offsets**, negative values indicate direction: -5 might mean "five positions before the current point." When you're working with **time differences in certain contexts**, a negative value might represent "in the past" versus positive for "in the future."

The key principle: use `int` when negative numbers carry semantic meaning in your logic, and use `count` when they don't.

### Basic Usage

Working with `int` is straightforward and similar to `count`, except you can freely work with negative values:

```c
local temperature: int = -40;
local delta: int = 100 - 150;  # Result: -50
local offset: int = -5;
```

All the arithmetic operations you'd expect work naturally: addition, subtraction, multiplication, division (integer division, rounding toward zero), and modulo. Comparisons work identically to `count`, letting you check if one integer is greater than, less than, or equal to another.

### Choosing Between int and count

One of the most important skills when writing Zeek scripts is knowing which numeric type to use. Let's look at concrete examples that clarify the distinction:

**Use count for quantities that cannot logically be negative:**

```c
# Use count for things that can't be negative
local packets_seen: count = 0;        # ✓ Correct
local connections: count = 100;       # ✓ Correct
local failed_logins: count = 0;       # ✓ Correct
```

These are all counting absolute quantities. There's no scenario where you'd have negative packets or negative connections - these concepts don't make physical sense.

**Use int when negative values are possible or carry meaning:**

```c
# Use int when negatives are possible or meaningful
local time_difference: int = -30;     # ✓ Correct (30 seconds ago)
local position_offset: int = -10;     # ✓ Correct (10 units before)
local byte_delta: int = current_bytes - previous_bytes;  # Could be negative
```

In these cases, a negative value carries information. A time difference of -30 tells you something happened 30 seconds in the past. A position offset of -10 indicates a location before your reference point.

**What not to do:**

```c
# Don't do this:
local packets_seen: int = -5;       
# ✗ Logically wrong (can't have -5 packets)
 
local connection_count: int = -10;  
# ✗ Nonsensical (negative connections?)
```

Declaring a packet count or connection count as `int` isn't a syntax error - Zeek will allow it - but it's a **logical error**. It suggests your code might produce or accept negative values for something that can't be negative, which will lead to bugs and confusion later.

### Practical Guidance

Here's the bottom line: **in practice, you'll use count about 90% of the time** in Zeek scripts. Network security analysis is fundamentally about counting things - packets, bytes, connections, events, alerts. The `count` type's non-negativity constraint actually helps you write more correct code by preventing logical errors.

Reserve `int` for those specific situations where negative values genuinely make sense in your domain. If you're unsure, start with `count`. If you later find yourself needing to represent negative values and the type system complains, that's your signal to switch to `int`. This approach - defaulting to `count` and using `int` only when necessary - will lead to clearer, more maintainable security scripts.

The type system is your friend here. By choosing the most semantically appropriate type, you make your code's intent obvious and let Zeek catch mistakes before they become runtime bugs.


```
┌──────────────────────────────────────────────────────────────┐
│              INT TYPE USE CASES                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Deltas and Differences                                      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  local byte_diff: int = current_bytes - previous_bytes;      │
│  # Could be negative if connection decreased (retransmit)    │
│                                                              │
│  Relative Positions                                          │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  local offset: int = -5;  # 5 positions before current       │
│                                                              │
│  Directional Values                                          │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  local direction: int = -1;  # Backwards                     │
│                                                              │
│  In practice: count is used 90% of the time in Zeek scripts  │
│  Use int only when you specifically need negative values     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


## Knowledge Check: int Type

**Q1: In what scenarios should you use int instead of count? Provide two specific examples.**

A: Use int when negative values carry semantic meaning. Examples: (1) Time differences where negative means "in the past" (e.g., -30 seconds ago), (2) Byte deltas in connections where a negative value might indicate retransmissions or measurement timing issues, (3) Position offsets where -10 means "10 units before the current point."

**Q2: The document states you'll use count "about 90% of the time" in Zeek scripts. Why is int used so rarely in network security analysis?**

A: Most quantities in network security analysis are inherently non-negative - you can't observe negative packets, negative connections, or negative bytes transferred. Since count is semantically appropriate for these common cases and provides additional safety through its non-negativity constraint, int is only needed for the specific cases where negative values genuinely make sense.

**Q3: Is it a syntax error to declare a packet count as `local packets: int = -5`? If not, what kind of problem is it?**

A: No, it's not a syntax error - Zeek will allow it. However, it's a logical error. Declaring a packet count as int suggests your code might produce or accept negative values for something that can't logically be negative, which will lead to bugs and confusion. It's semantically incorrect even though it's syntactically valid.

**Q4: What's the better default choice when you're unsure whether to use count or int, and why?**

A: Start with count. If you later find yourself needing to represent negative values and the type system complains, that's your signal to switch to int. This approach - defaulting to count and using int only when necessary - leads to clearer, more maintainable code because count's non-negativity constraint helps catch logical errors.

---

## Knowledge Check: addr Type

**Q1: Why does Zeek have a dedicated addr type instead of just representing IP addresses as strings?**

A: The addr type stores addresses in optimized binary format and provides built-in intelligence about network topology, address families, and comparison operations. Using strings would require manually parsing, converting to binary, applying subnet masks, and comparing results every time you need to check subnet membership or perform other network operations - tens of lines of error-prone code for conceptually simple questions.

**Q2: How does the addr type handle IPv4 and IPv6 addresses? Do you need different code or operations for each?**

A: The addr type handles both IPv4 and IPv6 transparently and seamlessly. All operations (comparison, subnet membership, string conversion) work identically on both address families. You don't need different code paths or conversion functions. This design choice dramatically simplifies script writing and maintenance.

**Q3: What is the most powerful operation the addr type supports, and why is it significant for security analysis?**

A: Subnet membership testing using the `in` operator (e.g., `if ( ip in 192.168.0.0/16 )`). This single line replaces dozens of lines of bit manipulation and mask logic you'd need in most languages. It's self-documenting, impossible to get wrong, and is fundamental to network boundary logic that appears in virtually every security detection.

**Q4: When would you use the `is_v4_addr()` or `is_v6_addr()` functions, given that addr handles both transparently?**

A: Use these functions when you need version-specific logic - for example, applying different subnet checks to IPv4 private ranges versus IPv6 unique local addresses, or when you need to handle the two address families differently for some reason. Most of the time you don't need them because operations work on both.


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./count.md" >}})
[|NEXT|]({{< ref "./addr.md" >}})

