---
showTableOfContents: true
title: "Part 2 - Scalar Types: Single Values"
type: "page"
---



## **The count Type: Non-Negative Integers**

The `count` type is one of Zeek's most frequently used types, representing non-negative integers - that is, zero and any positive whole number. It's specifically named "count" because its primary purpose is exactly what the name suggests: counting things. Whether you're tracking how many packets passed through your network, how many bytes were transferred in a connection, or how many failed login attempts came from a particular IP address, the `count` type is your tool of choice.

### **Why Have a Special Type for Counting?**

You might wonder why Zeek bothers with a dedicated `count` type instead of just using a generic integer. The answer is **safety and intent**. In network security analysis, many values are logically non-negative - you can't have negative five packets or negative three connections. By using the `count` type, Zeek enforces this constraint at the language level, catching errors before they become bugs. If you accidentally write code that would produce a negative count, Zeek will flag this as a type error, preventing logical mistakes from creeping into your security analysis.

Additionally, using `count` makes your code's intent clearer. When someone reads your script and sees a variable declared as `count`, they immediately understand it's tracking a quantity that increases from zero upward.

### **Basic Usage and Operations**

Working with `count` variables is straightforward and intuitive. Let's look at common operations:

```c  
# Declaring count variables  
local packet_count: count = 0;  
local max_connections: count = 1000;  
local threshold: count = 100;  
  
# Count arithmetic  
packet_count = packet_count + 1;  # Increment  
packet_count += 1;                # Shorthand increment  
packet_count = packet_count * 2;  # Multiplication  
```  

Notice the shorthand `+=` operator - this is particularly useful for incrementing counters, which you'll do constantly in security scripts. You can use all standard arithmetic operations: addition, subtraction, multiplication, division, and modulo (remainder). Just remember that subtraction must not produce a negative result, or Zeek will raise an error.

### **Understanding count's Range and Behavior**

The `count` type is implemented as a 64-bit unsigned integer, giving you an enormous range to work with. The minimum value is always zero, and the maximum is 18,446,744,073,709,551,615 (that's over 18 quintillion). For practical network security work, you'll virtually never hit this limit - even counting every packet on a very busy network for years would struggle to reach it.

Because `count` cannot be negative, this constraint is **enforced at compile time** - Zeek's script interpreter checks your code before running it. This means you'll catch mistakes like accidentally subtracting a larger number from a smaller one during development, not in production when it could cause silent failures or incorrect security decisions.

When you perform **division** with counts, remember that Zeek uses integer division, which rounds down. For example, `7 / 2` equals `3`, not `3.5`. If you need fractional results, you'll need to convert to the `double` type (we'll cover type casting later).

**Comparison operations** work exactly as you'd expect: you can check if one count is equal to, greater than, less than, or not equal to another. These comparisons are essential for implementing thresholds and triggering alerts.

### **Practical Example: Tracking Failed Connection Attempts**

Let's see `count` in action with a realistic security use case - monitoring failed connection attempts per IP address to detect potential brute force or scanning activity:

```c  
# Track failed connection attempts per IP  
global failed_attempts: table[addr] of count;  
  
event connection_rejected(c: connection)  
{  
    local src = c$id$orig_h;        # Initialize if first time seeing this IP  
    if ( src !in failed_attempts )        failed_attempts[src] = 0;        # Increment count  
    failed_attempts[src] += 1;        # Check threshold  
    if ( failed_attempts[src] >= 10 )    {        print fmt("%s has %d failed connections", src, failed_attempts[src]);    }}  
```  

**Walking through this example:** We're maintaining a table that maps IP addresses to their count of failed connection attempts. Each time a connection is rejected, we check if we've seen this source IP before. If not, we initialize its count to zero. Then we increment the count and check if it's reached our threshold of 10 failed attempts. This simple pattern - initialize, increment, compare - is fundamental to countless security detection scripts.

Notice how the `count` type makes this code clean and safe. We don't need to worry about accidentally storing negative numbers or handling type conversions. The type system ensures our counter behaves correctly.

### **Common Uses in Network Security**

The `count` type appears throughout security analysis scripts. Here are the most common scenarios where you'll reach for it:

**Packet and byte counting:** Track how many packets or bytes have been sent or received in a connection. This is essential for detecting data exfiltration, unusual upload/download patterns, or bandwidth abuse.

**Connection tracking:** Count how many connections each host has initiated or received. Sudden spikes might indicate scanning behavior or a compromised system reaching out to multiple targets.

**Threshold-based detection:** Set maximum allowed counts for various behaviors - maximum failed login attempts, maximum DNS queries per minute, maximum connections to unique destinations. When these thresholds are exceeded, you trigger alerts.

**Event frequency analysis:** Count how often specific events occur within time windows. For example, counting how many times a particular DNS query appears, how many HTTP requests hit a specific endpoint, or how many times a signature fires.

**Statistical aggregation:** When building baselines or performing behavioral analysis, you often need to count occurrences across different dimensions - counts per host, per subnet, per protocol, per time period.

The simplicity of the `count` type belies its power. Most sophisticated security detections ultimately rely on counting something and comparing it to expected values. Mastering when and how to use `count` effectively is foundational to writing robust Zeek scripts.


### Summary: Count Characteristics


```
┌──────────────────────────────────────────────────────────────┐
│                    COUNT TYPE PROPERTIES                     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Range: 0 to 2^64-1 (64-bit unsigned integer)                │
│  └─ Min: 0                                                   │
│  └─ Max: 18,446,744,073,709,551,615                          │
│                                                              │
│  Cannot be negative                                          │
│  └─ This is enforced at compile time                         │
│  └─ Prevents logical errors (can't have -5 packets)          │
│                                                              │
│  Arithmetic operations                                       │
│  ✓ Addition: a + b                                           │
│  ✓ Subtraction: a - b (but result must be >= 0)              │
│  ✓ Multiplication: a * b                                     │
│  ✓ Division: a / b (integer division, rounds down)           │
│  ✓ Modulo: a % b (remainder)                                 │
│                                                              │
│  Comparison operations                                       │
│  ✓ Equal: a == b                                             │
│  ✓ Not equal: a != b                                         │
│  ✓ Greater: a > b, a >= b                                    │
│  ✓ Lesser: a < b, a <= b                                     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

___



## **The int Type: Signed Integers**

The `int` type represents signed integers - whole numbers that can be positive, negative, or zero. While the `count` type is far more common in Zeek scripts, `int` fills an important niche: it's the type to reach for when negative values are not just possible but actually meaningful in your analysis.

### **When Do You Need Signed Integers?**

Most quantities in network security analysis are inherently non-negative. You can't observe negative three packets or have negative seven connections. This is why `count` dominates Zeek scripts. However, certain calculations and concepts naturally produce or require negative numbers, and that's where `int` becomes essential.

Think about **differences and deltas**. If you're comparing the current byte count of a connection to a previous measurement, the difference could be negative - perhaps due to retransmissions or measurement timing. When you're tracking **relative positions or offsets**, negative values indicate direction: -5 might mean "five positions before the current point." When you're working with **time differences in certain contexts**, a negative value might represent "in the past" versus positive for "in the future."

The key principle: use `int` when negative numbers carry semantic meaning in your logic, and use `count` when they don't.

### **Basic Usage**

Working with `int` is straightforward and similar to `count`, except you can freely work with negative values:

```c
local temperature: int = -40;
local delta: int = 100 - 150;  # Result: -50
local offset: int = -5;
```

All the arithmetic operations you'd expect work naturally: addition, subtraction, multiplication, division (integer division, rounding toward zero), and modulo. Comparisons work identically to `count`, letting you check if one integer is greater than, less than, or equal to another.

### **Choosing Between int and count**

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

Declaring a packet count or connection count as `int` isn't a syntax error - Zeek will allow it - but it's a **logical error**. It suggests your code might produce or accept negative values for something that can't be negative, which will lead to bugs and confusion later.

### **Practical Guidance**

Here's the bottom line: **in practice, you'll use count about 90% of the time** in Zeek scripts. Network security analysis is fundamentally about counting things - packets, bytes, connections, events, alerts. The `count` type's non-negativity constraint actually helps you write more correct code by preventing logical errors.

Reserve `int` for those specific situations where negative values genuinely make sense in your domain. If you're unsure, start with `count`. If you later find yourself needing to represent negative values and the type system complains, that's your signal to switch to `int`. This approach - defaulting to `count` and using `int` only when necessary - will lead to clearer, more maintainable security scripts.

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


___





---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./type.md" >}})
[|NEXT|]({{< ref "./complex.md" >}})

