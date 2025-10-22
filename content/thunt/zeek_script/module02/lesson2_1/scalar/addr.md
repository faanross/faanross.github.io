---
showTableOfContents: true
title: "The addr Type: IP Addresses"
type: "page"
---






## The addr Type: IP Addresses

The `addr` type is arguably the most fundamental type in Zeek's arsenal. 
Since network security analysis revolves around tracking who's communicating with whom, 
IP addresses are the core identifiers you'll work with constantly. The `addr` type isn't just a container for address strings - it's a sophisticated type with built-in intelligence about network topology, address families, and comparison operations that make security analysis intuitive and powerful.

### Why addr Is Special

In most programming languages, you'd represent an IP address as a string like `"192.168.1.100"`. This approach is simple but creates constant friction. Every time you need to check if an address belongs to a particular subnet, you'd need to parse the string, convert it to binary, apply subnet masks, and compare results - tens of lines of error-prone code for a conceptually simple question.

Zeek's `addr` type handles all of this complexity internally. It stores addresses in an optimized binary format, supports both IPv4 and IPv6 seamlessly, and provides natural operations like subnet membership testing and address comparison. This isn't just convenience - it's a fundamental design choice that makes security logic readable and correct.

### Basic Usage and Declaration

Declaring and working with IP addresses in Zeek feels natural. You simply write the address in its standard notation:

```c
# IPv4 addresses
local ip1: addr = 192.168.1.100;
local ip2: addr = 10.0.0.5;

# IPv6 addresses
local ip6: addr = 2001:db8::1;
local ip6_local: addr = fe80::1;

# Addresses from variables
local attacker_ip: addr = c$id$orig_h;  
# From connection record
```

Notice how IPv4 and IPv6 addresses use their standard notation - no quotes, no special syntax. Zeek recognizes the format and handles the rest. When extracting addresses from connection records (like `c$id$orig_h` for the originating host), you're getting proper `addr` types that carry all the built-in functionality.

### What You Can Do with Addresses

The `addr` type supports a rich set of operations that mirror how you think about IP addresses in security work:

#### Equality and Inequality Comparisons 

Lets you check if an address matches a specific value or differs from another address:

```c
# Comparison
if ( ip1 == 192.168.1.100 )
    print "Matched specific IP";
    
if ( ip1 != ip2 )
    print "Different IPs";
```

This is essential for allow lists, deny lists, and detecting known malicious or trusted hosts.

#### Lexicographic Ordering
Allows you to sort addresses and perform range comparisons. While this might seem abstract, it's useful when building ordered data structures or implementing IP-based indexes:

```c
# Lexicographic comparison (useful for sorting)
if ( ip1 < ip2 )
    print fmt("%s comes before %s lexicographically", ip1, ip2);
```

#### String Conversion
Is straightforward when you need to log or display addresses in human-readable format:

```c
# String conversion
local ip_string = fmt("%s", ip1);  
# Prints to terminal: "192.168.1.100"
```

#### Subnet membership testing
Is where `addr` truly shines. Checking if an address belongs to a subnet is a one-liner:

```c
# Subnet membership testing (covered more in subnet type)
if ( ip1 in 192.168.0.0/16 )
    print "IP is in private network";
```

This single line replaces dozens of lines of bit manipulation and mask logic you'd need in most languages. It's not just shorter - it's self-documenting and impossible to get wrong.

### Real-World Security Example: Detecting Lateral Movement

Let's see the `addr` type in action with a practical security use case - detecting potential lateral movement by identifying hosts that connect to many different internal targets:

```c
# Track IPs that connect to multiple internal hosts (lateral movement)
global connections_by_ip: table[addr] of set[addr];

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    local dst = c$id$resp_h;
    
    # Initialize set if first time seeing this source
    if ( src !in connections_by_ip )
        connections_by_ip[src] = set();
    
    # Add destination to set of targets
    add connections_by_ip[src][dst];
    
    # Check if connecting to many hosts (potential lateral movement)
    if ( |connections_by_ip[src]| >= 10 )
    {
        print fmt("ALERT: %s connected to %d different hosts", 
                  src, |connections_by_ip[src]|);
    }
}
```

**Understanding this example:** We're maintaining a table that maps each source IP address to a set of destination addresses it has contacted. Each time we see a new connection, we add the destination to that source's set. When any source IP has contacted 10 or more distinct destinations, we trigger an alert - this pattern often indicates an attacker moving laterally through a network after initial compromise.

Notice how naturally the `addr` type works here. We use addresses as table keys, store them in sets, and compare them for uniqueness - all without any special handling. The type system just works, letting us focus on the security logic rather than data wrangling.

### IPv4 and IPv6: Transparent Handling

One of the most elegant aspects of the `addr` type is its transparent handling of both IPv4 and IPv6 addresses. Modern networks use both addressing schemes, and your security scripts need to handle both without special cases or conditional logic.

```c
# Zeek handles both transparently
local ipv4: addr = 192.168.1.1;
local ipv6: addr = 2001:db8::1;

# Same operations work on both
if ( ipv4 == 192.168.1.1 )  
    print "IPv4 matched";
    
if ( ipv6 == 2001:db8::1 )  
    print "IPv6 matched";
```

All the operations we've discussed - comparison, subnet membership, string conversion - work identically on both IPv4 and IPv6 addresses. You don't need different code paths or conversion functions. This design choice dramatically simplifies script writing and maintenance.

**When you do need to distinguish between address families**, Zeek provides helper functions:

```c
# Check IP version
if ( is_v4_addr(ipv4) )
    print "This is IPv4";
    
if ( is_v6_addr(ipv6) )
    print "This is IPv6";
```

These functions are useful when you need version-specific logic - for example, applying different subnet checks to IPv4 private ranges versus IPv6 unique local addresses.

### Why This Matters for Security Analysis

The `addr` type's design reflects a deep understanding of network security work. Most security detections ultimately answer questions like "Who connected to whom?", "Is this address in our network?", "Have we seen this IP before?", and "How many different hosts has this IP contacted?"

By making IP addresses first-class citizens with rich built-in operations, Zeek lets you express these questions naturally and correctly. You spend less time fighting with data types and more time building effective detections. The type system guides you toward correct code - you can't accidentally compare an IP address to a string or forget to handle IPv6.

As you work through more complex Zeek scripts, you'll find the `addr` type appearing everywhere: as table keys for tracking host behavior, as set members for allow/deny lists, in subnet comparisons for network segmentation enforcement, and in logging for forensic analysis. Mastering its capabilities and idioms is essential to effective Zeek scripting.


---

## Knowledge Check: addr Type

**Q1: Why does Zeek have a dedicated addr type instead of just representing IP addresses as strings?**

A: The addr type stores addresses in optimized binary format and provides built-in intelligence about network topology, address families, and comparison operations. Using strings would require manually parsing, converting to binary, applying subnet masks, and comparing results every time you need to check subnet membership or perform other network operations - tens of lines of error-prone code for conceptually simple questions.

**Q2: How does the addr type handle IPv4 and IPv6 addresses? Do you need different code or operations for each?**

A: The addr type handles both IPv4 and IPv6 transparently and seamlessly. All operations (comparison, subnet membership, string conversion) work identically on both address families. You don't need different code paths or conversion functions. This design choice dramatically simplifies script writing and maintenance.

**Q3: What is the most powerful operation the addr type supports, and why is it significant for security analysis?**

A: Subnet membership testing using the `in` operator (e.g., `if ( ip in 192.168.0.0/16 )`). This single line replaces dozens of lines of bit manipulation and mask logic you'd need in most languages. It's self-documenting, impossible to get wrong, and is fundamental to network boundary logic that appears in virtually every security detection.

**Q4: When would you use the `is_v4_addr()` or `is_v6_addr()` functions, given that addr handles both transparently?**

A: Use these functions when you need version-specific logic - for example, applying different subnet checks to IPv4 private ranges versus IPv6 unique local addresses, or when you need to handle the two address families differently for some reason. Most of the time you don't need them because operations work on both.


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./int.md" >}})
[|NEXT|]({{< ref "./subnet.md" >}})

