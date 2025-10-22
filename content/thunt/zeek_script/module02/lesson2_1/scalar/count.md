---
showTableOfContents: true
title: "The Count Type: Non-Negative Integers"
type: "page"
---


## The count Type: Non-Negative Integers

The `count` type is one of Zeek's most frequently used types, representing non-negative integers - that is, zero and any positive whole number. It's specifically named "count" because its primary purpose is exactly what the name suggests: counting things. Whether you're tracking how many packets passed through your network, how many bytes were transferred in a connection, or how many failed login attempts came from a particular IP address, the `count` type is your tool of choice.

### Why Have a Special Type for Counting?

You might wonder why Zeek bothers with a dedicated `count` type instead of just using a generic integer. The answer is **safety and intent**. In network security analysis, many values are logically non-negative - you can't have negative five packets or negative three connections. 

By using the `count` type, Zeek enforces this constraint at the language level, catching errors before they become bugs. If you accidentally write code that would produce a negative count, Zeek will flag this as a type error, preventing logical mistakes from creeping into your security analysis.

Additionally, using `count` makes your code's intent clearer. When someone reads your script and sees a variable declared as `count`, they immediately understand it's tracking a quantity that increases from zero upward.

### Basic Usage and Operations

Working with `count` variables is straightforward and intuitive. Let's look at common operations:

```c  
# Declaring count variables  
local packet_count: count = 0;  
local max_connections: count = 1000;  
local threshold: count = 100;  
  
# Count arithmetic
# Increment    
packet_count = packet_count + 1;  
# Shorthand increment  
packet_count += 1;                
# Multiplication  
packet_count = packet_count * 2;  
```  

Notice the shorthand `+=` operator - this is particularly useful for incrementing counters, which you'll do constantly in security scripts. You can use all standard arithmetic operations: addition, subtraction, multiplication, division, and modulo (remainder). Just remember that subtraction must not produce a negative result, or Zeek will raise an error.

### Understanding count's Range and Behavior

The `count` type is implemented as a 64-bit unsigned integer, giving you an enormous range to work with. The minimum value is always zero, and the maximum is 18,446,744,073,709,551,615 (that's over 18 quintillion). For practical network security work, you'll virtually never hit this limit - even counting every packet on a very busy network for years would struggle to reach it.

Because `count` cannot be negative, this constraint is **enforced at compile time** - Zeek's script interpreter checks your code before running it. This means you'll catch mistakes like accidentally subtracting a larger number from a smaller one during development, not in production when it could cause silent failures or incorrect security decisions.

When you perform **division** with counts, remember that Zeek uses integer division, which rounds down. For example, `7 / 2` equals `3`, not `3.5`. If you need fractional results, you'll need to convert to the `double` type (we'll cover type casting later).

**Comparison operations** work exactly as you'd expect: you can check if one count is equal to, greater than, less than, or not equal to another. These comparisons are essential for implementing thresholds and triggering alerts.

### Practical Example: Tracking Failed Connection Attempts

Let's see `count` in action with a realistic security use case - monitoring failed connection attempts per IP address to detect potential brute force or scanning activity:


```c
# Track failed connection attempts per IP
global failed_attempts: table[addr] of count;

event connection_rejected(c: connection)
    {
    # Initialize if first time seeing this IP
    local src = c$id$orig_h;
    
    if ( src !in failed_attempts )
        failed_attempts[src] = 0;
    
    # Increment count
    failed_attempts[src] += 1;
    
    # Check threshold
    if ( failed_attempts[src] >= 10 )
        {
        print fmt("%s has %d failed connections", src, failed_attempts[src]);
        }
    }
```


**Walking through this example:** 

We're maintaining a table that maps IP addresses to their count of failed connection attempts. Each time a connection is rejected, we check if we've seen this source IP before. 

If not, we initialize its count to zero. Then we increment the count and check if it's reached our threshold of 10 failed attempts. This simple pattern - **initialize, increment, compare** - is fundamental to countless security detection scripts.

Notice how the `count` type makes this code clean and safe. We don't need to worry about accidentally storing negative numbers or handling type conversions. The type system ensures our counter behaves correctly.

### Common Uses in Network Security

The `count` type appears throughout security analysis scripts. Here are the most common scenarios where you'll reach for it:

**Packet and byte counting:** Track how many packets or bytes have been sent or received in a connection. This is essential for detecting data exfiltration, unusual upload/download patterns, or bandwidth abuse.

**Connection tracking:** Count how many connections each host has initiated or received. Sudden spikes might indicate scanning behavior or a compromised system reaching out to multiple targets.

**Threshold-based detection:** Set maximum allowed counts for various behaviors - maximum failed login attempts, maximum DNS queries per minute, maximum connections to unique destinations. When these thresholds are exceeded, you trigger alerts.

**Event frequency analysis:** Count how often specific events occur within time windows. For example, counting how many times a particular DNS query appears, how many HTTP requests hit a specific endpoint, or how many times a signature fires.

**Statistical aggregation:** When building baselines or performing behavioral analysis, you often need to count occurrences across different dimensions - counts per host, per subnet, per protocol, per time period.

The simplicity of the `count` type belies its power. Most sophisticated security detections ultimately rely on counting something and comparing it to expected values. Mastering when and how to use `count` effectively is foundational to writing robust Zeek scripts.


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


## Knowledge Check

**Q1: What is the valid range of values for the count type, and why can't it represent negative numbers?**

A: The count type ranges from 0 to 18,446,744,073,709,551,615 (2^64-1). It cannot represent negative numbers because it's specifically designed for counting quantities that logically cannot be negative - you can't have -5 packets or -3 connections. This constraint is enforced at compile time to prevent logical errors in security analysis code.

**Q2: What happens if you write code that would produce a negative count value? When is this error detected?**

A: Zeek will flag this as a type error at compile time (before the script runs), not at runtime. For example, if you try to subtract a larger count from a smaller one directly, Zeek's script interpreter will catch this during the compilation phase, preventing the error from occurring in production.

**Q3: When performing division with count values (e.g., 7 / 2), what type of result do you get and what is its value?**

A: You get integer division that rounds down, so 7 / 2 equals 3 (not 3.5). The result is still a count type. If you need fractional results, you must convert to the double type first.

**Q4: Why is using the count type considered safer and more expressive than using a generic integer type for tracking network quantities?**

A: Count makes code safer by enforcing non-negativity at the language level, catching errors before they become bugs. It also makes code more expressive and self-documenting - when someone sees a variable declared as count, they immediately understand it's tracking a quantity that increases from zero upward, making the code's intent clearer.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./intro.md" >}})
[|NEXT|]({{< ref "./int.md" >}})

