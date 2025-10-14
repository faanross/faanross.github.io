---
showTableOfContents: true
title: "Part 2 - Scalar Types: Single Values"
type: "page"
---



## The count Type: Non-Negative Integers

The `count` type is one of Zeek's most frequently used types, representing non-negative integers - that is, zero and any positive whole number. It's specifically named "count" because its primary purpose is exactly what the name suggests: counting things. Whether you're tracking how many packets passed through your network, how many bytes were transferred in a connection, or how many failed login attempts came from a particular IP address, the `count` type is your tool of choice.

### Why Have a Special Type for Counting?

You might wonder why Zeek bothers with a dedicated `count` type instead of just using a generic integer. The answer is **safety and intent**. In network security analysis, many values are logically non-negative - you can't have negative five packets or negative three connections. By using the `count` type, Zeek enforces this constraint at the language level, catching errors before they become bugs. If you accidentally write code that would produce a negative count, Zeek will flag this as a type error, preventing logical mistakes from creeping into your security analysis.

Additionally, using `count` makes your code's intent clearer. When someone reads your script and sees a variable declared as `count`, they immediately understand it's tracking a quantity that increases from zero upward.

### Basic Usage and Operations

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

### Understanding count's Range and Behavior

The `count` type is implemented as a 64-bit unsigned integer, giving you an enormous range to work with. The minimum value is always zero, and the maximum is 18,446,744,073,709,551,615 (that's over 18 quintillion). For practical network security work, you'll virtually never hit this limit - even counting every packet on a very busy network for years would struggle to reach it.

Because `count` cannot be negative, this constraint is **enforced at compile time** - Zeek's script interpreter checks your code before running it. This means you'll catch mistakes like accidentally subtracting a larger number from a smaller one during development, not in production when it could cause silent failures or incorrect security decisions.

When you perform **division** with counts, remember that Zeek uses integer division, which rounds down. For example, `7 / 2` equals `3`, not `3.5`. If you need fractional results, you'll need to convert to the `double` type (we'll cover type casting later).

**Comparison operations** work exactly as you'd expect: you can check if one count is equal to, greater than, less than, or not equal to another. These comparisons are essential for implementing thresholds and triggering alerts.

### Practical Example: Tracking Failed Connection Attempts

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

### Common Uses in Network Security

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



## The int Type: Signed Integers

The `int` type represents signed integers - whole numbers that can be positive, negative, or zero. While the `count` type is far more common in Zeek scripts, `int` fills an important niche: it's the type to reach for when negative values are not just possible but actually meaningful in your analysis.

### When Do You Need Signed Integers?

Most quantities in network security analysis are inherently non-negative. You can't observe negative three packets or have negative seven connections. This is why `count` dominates Zeek scripts. However, certain calculations and concepts naturally produce or require negative numbers, and that's where `int` becomes essential.

Think about **differences and deltas**. If you're comparing the current byte count of a connection to a previous measurement, the difference could be negative - perhaps due to retransmissions or measurement timing. When you're tracking **relative positions or offsets**, negative values indicate direction: -5 might mean "five positions before the current point." When you're working with **time differences in certain contexts**, a negative value might represent "in the past" versus positive for "in the future."

The key principle: use `int` when negative numbers carry semantic meaning in your logic, and use `count` when they don't.

### Basic Usage

Working with `int` is straightforward and similar to `count`, except you can freely work with negative values:

```c
local temperature: int = -40;
local delta: int = 100 - 150;  # Result: -50
local offset: int = -5;
```

All the arithmetic operations you'd expect work naturally: addition, subtraction, multiplication, division (integer division, rounding toward zero), and modulo. Comparisons work identically to `count`, letting you check if one integer is greater than, less than, or equal to another.

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

Declaring a packet count or connection count as `int` isn't a syntax error - Zeek will allow it - but it's a **logical error**. It suggests your code might produce or accept negative values for something that can't be negative, which will lead to bugs and confusion later.

### Practical Guidance

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



## The addr Type: IP Addresses

The `addr` type is arguably the most fundamental type in Zeek's arsenal. Since network security analysis revolves around tracking who's communicating with whom, IP addresses are the core identifiers you'll work with constantly. The `addr`type isn't just a container for address strings - it's a sophisticated type with built-in intelligence about network topology, address families, and comparison operations that make security analysis intuitive and powerful.

### Why addr Is Special

In most programming languages, you'd represent an IP address as a string like `"192.168.1.100"`. This approach is simple but creates constant friction. Every time you need to check if an address belongs to a particular subnet, you'd need to parse the string, convert it to binary, apply subnet masks, and compare results - tens of lines of error-prone code for a conceptually simple question.

Zeek's `addr` type handles all of this complexity internally. It stores addresses in an optimized binary format, supports both IPv4 and IPv6 seamlessly, and provides natural operations like subnet membership testing and address comparison. This isn't just convenience - it's a fundamental design choice that makes security logic readable and correct.

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
local attacker_ip: addr = c$id$orig_h;  # From connection record
```

Notice how IPv4 and IPv6 addresses use their standard notation - no quotes, no special syntax. Zeek recognizes the format and handles the rest. When extracting addresses from connection records (like `c$id$orig_h` for the originating host), you're getting proper `addr` types that carry all the built-in functionality.

### What You Can Do with Addresses

The `addr` type supports a rich set of operations that mirror how you think about IP addresses in security work:

**Equality and inequality comparisons** let you check if an address matches a specific value or differs from another address:

```c
# Comparison
if ( ip1 == 192.168.1.100 )
    print "Matched specific IP";
    
if ( ip1 != ip2 )
    print "Different IPs";
```

This is essential for allow lists, deny lists, and detecting known malicious or trusted hosts.

**Lexicographic ordering** allows you to sort addresses and perform range comparisons. While this might seem abstract, it's useful when building ordered data structures or implementing IP-based indexes:

```c
# Lexicographic comparison (useful for sorting)
if ( ip1 < ip2 )
    print fmt("%s comes before %s lexicographically", ip1, ip2);
```

**String conversion** is straightforward when you need to log or display addresses in human-readable format:

```c
# String conversion
local ip_string = fmt("%s", ip1);  # "192.168.1.100"
```

**Subnet membership testing** is where `addr` truly shines. Checking if an address belongs to a subnet is a one-liner:

```c
# Subnet membership testing (covered more in subnet type)
if ( ip1 in 192.168.0.0/16 )
    print "IP is in private network";
```

This single line replaces dozens of lines of bit manipulation and mask logic you'd need in most languages. It's not just shorter - it's self-documenting and impossible to get wrong.

### Real-World Security Example: Detecting Lateral Movement

Let's see the `addr` type in action with a practical security use case - detecting potential lateral movement by identifying hosts that connect to many different internal targets:

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

**Understanding this example:** We're maintaining a table that maps each source IP address to a set of destination addresses it has contacted. Each time we see a new connection, we add the destination to that source's set. When any source IP has contacted 10 or more distinct destinations, we trigger an alert - this pattern often indicates an attacker moving laterally through a network after initial compromise.

Notice how naturally the `addr` type works here. We use addresses as table keys, store them in sets, and compare them for uniqueness - all without any special handling. The type system just works, letting us focus on the security logic rather than data wrangling.

### IPv4 and IPv6: Transparent Handling

One of the most elegant aspects of the `addr` type is its transparent handling of both IPv4 and IPv6 addresses. Modern networks use both addressing schemes, and your security scripts need to handle both without special cases or conditional logic.

```c
# Zeek handles both transparently
local ipv4: addr = 192.168.1.1;
local ipv6: addr = 2001:db8::1;

# Same operations work on both
if ( ipv4 == 192.168.1.1 )  # Works
    print "IPv4 matched";
    
if ( ipv6 == 2001:db8::1 )  # Works
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

The `addr` type's design reflects a deep understanding of network security work. Most security detections ultimately answer questions like "Who connected to whom?", "Is this address in our network?", "Have we seen this IP before?", and "How many different hosts has this IP contacted?"

By making IP addresses first-class citizens with rich built-in operations, Zeek lets you express these questions naturally and correctly. You spend less time fighting with data types and more time building effective detections. The type system guides you toward correct code - you can't accidentally compare an IP address to a string or forget to handle IPv6.

As you work through more complex Zeek scripts, you'll find the `addr` type appearing everywhere: as table keys for tracking host behavior, as set members for allow/deny lists, in subnet comparisons for network segmentation enforcement, and in logging for forensic analysis. Mastering its capabilities and idioms is essential to effective Zeek scripting.

___


## The subnet Type: Network Ranges

The `subnet` type represents network ranges using CIDR (Classless Inter-Domain Routing) notation, and it's essential for practical network security analysis. Almost every meaningful security detection needs to distinguish between "our network" and "the outside world," or to group addresses by their network membership. The `subnet` type makes these operations natural and efficient.

### Why Subnets Matter in Security

Think about the questions you ask constantly in security work: "Is this connection coming from inside our network or outside?" "Which subnet is generating the most DNS queries?" "Is this IP in our DMZ or our internal corporate network?" All of these questions fundamentally rely on understanding network boundaries and address groupings.

Without a proper subnet type, you'd need to manually implement subnet mask logic every time - converting addresses to binary, applying masks, comparing network portions. It's tedious, error-prone, and obscures your actual security logic. Zeek's `subnet` type handles all the complexity, giving you a clean, declarative way to work with network ranges.

### Basic Usage and Declaration

Defining subnets in Zeek uses standard CIDR notation - the same notation you use in routing tables, firewall rules, and network documentation:

```c
# CIDR notation
local private_net: subnet = 192.168.0.0/16;
local corporate_net: subnet = 10.10.0.0/16;
local single_host: subnet = 192.168.1.100/32;  # Single IP as subnet

# IPv6 subnets
local ipv6_net: subnet = 2001:db8::/32;
```

The notation is intuitive: the address followed by a slash and the prefix length (number of network bits). A `/16` subnet contains about 65,000 addresses, a `/24` contains 256 addresses, and `/32` represents exactly one address (a single host treated as a subnet).

Just like the `addr` type, `subnet` handles both IPv4 and IPv6 transparently. The same operations and logic work for both address families.

### Subnet Membership: The Core Operation

The most powerful and frequently used operation with subnets is **membership testing** - checking if an IP address falls within a network range. This is the foundation of most network boundary logic:

```c
# Membership testing - THE most common use
local ip: addr = 192.168.1.100;

if ( ip in 192.168.0.0/16 )
    print "IP is in this subnet";
    
if ( ip !in 10.0.0.0/8 )
    print "IP is NOT in this subnet";
```

Notice the elegant `in` operator. This single expression - `ip in subnet` - performs all the binary mathematics behind the scenes: converting the address to binary, applying the subnet mask, and comparing the network portions. The result is code that reads almost like English: "if this IP is in this subnet, then..."

**Checking multiple subnets** is just as straightforward using logical operators:

```c
# Multiple subnet checks
if ( ip in 192.168.0.0/16 || ip in 10.0.0.0/8 )
    print "IP is in one of our private networks";
```

You can also **compare subnets directly** to check if they represent the same network range:

```c
# Subnet comparison
if ( 192.168.1.0/24 == 192.168.1.0/24 )
    print "Same subnet";
```

This is useful when normalizing or deduplicating network definitions in configuration.




### Real-World Example: Defining Your Network Boundary

One of the first things you'll do in any Zeek deployment is define which networks belong to your organization. This lets you distinguish internal traffic from external traffic, a fundamental categorization for almost all security detections. Here's how to do it properly:

```c
# Define your organization's networks
global local_networks: set[subnet] = {
    10.0.0.0/8,        # RFC 1918 private network
    172.16.0.0/12,     # RFC 1918 private network
    192.168.0.0/16,    # RFC 1918 private network
    203.0.113.0/24,    # Your public IP space (example)
};

# Function to check if IP is local
function is_local(ip: addr): bool
{
    for ( net in local_networks )
    {
        if ( ip in net )
            return T;
    }
    return F;
}
```

This pattern - defining a set of subnets and creating a helper function to check membership - appears in virtually every Zeek deployment. It centralizes your network definition, making updates easy and ensuring consistency across all your scripts.

**Using this in detection logic:**

```c
event connection_established(c: connection)
{
    local src = c$id$orig_h;
    local dst = c$id$resp_h;
    
    # Detect connections from external to internal
    if ( !is_local(src) && is_local(dst) )
    {
        print fmt("Inbound connection: %s -> %s", src, dst);
    }
    
    # Detect connections from internal to external
    if ( is_local(src) && !is_local(dst) )
    {
        print fmt("Outbound connection: %s -> %s", src, dst);
    }
}
```

**Understanding this example:** For every established connection, we check the directionality. If the source is external but the destination is internal, it's an inbound connection - someone from the internet reaching into your network. If the source is internal but the destination is external, it's an outbound connection - someone inside reaching out to the internet.

This distinction is fundamental to security. Many detections only care about one direction. For example, you might alert on certain outbound connections (potential data exfiltration or C2 traffic) but not the same activity inbound. Or you might track inbound connections to internal services to detect scanning or exploitation attempts.





## **Subnet Masking and Address Aggregation**

Sometimes you need to extract just the network portion of an IP address - essentially converting an address into its parent subnet. This is useful for **aggregating statistics by network block** or **grouping related addresses**.

Zeek provides the `mask_addr()` function for this:

```c
# Get network address from IP
local ip: addr = 192.168.1.100;
local net: subnet = 192.168.1.0/24;

# Extract network portion
local network_addr = mask_addr(ip, 24);  # Returns 192.168.1.0
```

The second parameter to `mask_addr()` is the prefix length - how many bits of the address to keep. A `/24` keeps the first three octets and zeros out the last octet, giving you the network address.

**Practical use case - grouping connections by /24 subnet:**

```c
# Useful for aggregating by /24 blocks
global connections_per_subnet: table[addr] of count;

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    local subnet_addr = mask_addr(src, 24);  # Group by /24
    
    if ( subnet_addr !in connections_per_subnet )
        connections_per_subnet[subnet_addr] = 0;
    
    connections_per_subnet[subnet_addr] += 1;
}
```

**Why this is useful:** Instead of tracking connections per individual IP address (which could be millions of entries), you're grouping by /24 subnets (about 16 million possible values for IPv4, but typically far fewer in practice). This aggregation helps identify which network blocks are most active, which subnets might be compromised, or which ranges are generating suspicious patterns.

This technique is especially valuable for detecting scanning activity. If you see hundreds of connections from different IPs within the same /24, it might indicate a distributed scan from a botnet or compromised subnet rather than isolated individual hosts.

### **Why This Matters for Security Analysis**

The `subnet` type embodies a fundamental truth about network security: **context is everything**. An IP address means different things depending on whether it's internal or external, whether it's in your DMZ or your corporate network, whether it's a single host or part of a larger cloud provider range.

By making subnets first-class citizens in the type system, Zeek lets you express network context naturally and correctly. You don't fight with bit manipulation or mask arithmetic - you write clear, declarative logic about network boundaries and membership.

As you build more sophisticated detections, you'll use subnets for increasingly nuanced purposes: defining trusted networks that skip certain detections, identifying networks with different security postures (guest WiFi vs employee networks vs servers), tracking activity by provider networks to detect cloud-based threats, and aggregating statistics at the subnet level to find patterns invisible at the per-IP level.

Mastering the `subnet` type means mastering one of the most fundamental abstractions in network security. Every professional Zeek deployment relies heavily on proper subnet definitions and membership testing - it's the foundation on which almost everything else builds.



## **The port Type: Network Ports**

The `port` type represents network ports - those numbers between 0 and 65535 that identify specific services or applications on a host. But Zeek's `port` type does something crucial that simple integers can't: it **binds the port number to its transport protocol** (TCP or UDP). This protocol awareness is essential because port 80 over TCP (HTTP web traffic) is completely different from port 80 over UDP (which might be something else entirely).

### **Why Protocol Context Matters**

In the real world of network security, port numbers without protocol context are almost meaningless. When you see traffic on port 53, you need to know: is it 53/tcp or 53/udp? DNS primarily uses UDP, so 53/udp is normal DNS traffic. But 53/tcp is typically DNS zone transfers or responses too large for UDP - much rarer and potentially interesting from a security perspective.

Similarly, 80/tcp is standard HTTP web traffic you'd expect everywhere. But 80/udp? That's unusual and worth investigating. By making protocol an intrinsic part of the `port` type, Zeek ensures you're always working with the complete picture. You can't accidentally compare TCP ports to UDP ports or forget to check which protocol you're dealing with - the type system enforces correctness.

### **Basic Usage and Declaration**

Declaring ports in Zeek uses an intuitive notation: the port number followed by a slash and the protocol:

```c
# Port with protocol
local http_port: port = 80/tcp;
local dns_port: port = 53/udp;
local https_port: port = 443/tcp;
```

The syntax reads naturally: "port 80 over TCP," "port 53 over UDP." This is exactly how network engineers and security analysts think and talk about ports in practice.

**Critically, the same number with different protocols creates different values:**

```c
# Same number, different protocols are DIFFERENT
local port1: port = 80/tcp;
local port2: port = 80/udp;

# port1 != port2  (they're different!)
```

This distinction prevents an entire class of bugs. If you're checking for HTTP traffic on port 80, you're checking for `80/tcp`, and you won't accidentally match unrelated UDP traffic that happens to use the same port number.

### **Working with Port Values**

The `port` type supports several useful operations for security analysis:

**Comparison** is straightforward - you can check if a port matches a specific value:

```c
# Comparison
if ( http_port == 80/tcp )
    print "Standard HTTP port";
```

**Extracting the numeric portion** when you need to do arithmetic or range checks:

```c
# Extract port number
local port_num = port_to_count(443/tcp);  # Returns 443 as count
```

The `port_to_count()` function gives you the raw port number as a `count` type, which you can then use in numeric comparisons or calculations. This is useful when you need to categorize ports by range (well-known, registered, ephemeral) or perform other numeric operations.

**Protocol information** is intrinsic to the port type itself. While you can't extract the protocol as a separate string value directly, you test against protocol-specific port values to determine what you're dealing with.

**Range checking** helps categorize ports by their IANA designation:

```c
# Check if port is in range
if ( port_num >= 1024 )
    print "Ephemeral or registered port";
```

Ports 0-1023 are "well-known" ports requiring root privileges on Unix systems. Ports 1024-49151 are "registered" ports for specific services. Ports 49152-65535 are "dynamic/ephemeral" ports used by clients for outbound connections.



### **Real-World Example: Detecting Protocol Mismatches**

One common evasion technique attackers use is running services on non-standard ports to avoid detection. Here's how to detect HTTP traffic on unusual ports:

```c
# Detect HTTP on non-standard ports (potential evasion)
event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
{
    local resp_port = c$id$resp_p;
    
    # HTTP detected on non-80, non-443 port
    if ( resp_port != 80/tcp && resp_port != 443/tcp && resp_port != 8080/tcp )
    {
        print fmt("ALERT: HTTP on unusual port %s from %s", 
                  resp_port, c$id$orig_h);
    }
}
```

**Understanding this detection:** Zeek's HTTP analyzer has identified HTTP protocol traffic based on the actual content of the packets - it's truly HTTP regardless of port. But the server is listening on something other than the standard ports 80, 443, or 8080. This could be legitimate (maybe it's a development server on port 8000), or it could be an attacker trying to evade simple port-based filtering.

**Detecting service/port mismatches more generally:**

```c
# Detect services on unexpected ports
global expected_services: table[port] of string = {
    [22/tcp] = "ssh",
    [80/tcp] = "http",
    [443/tcp] = "https",
    [53/udp] = "dns",
};

event connection_established(c: connection)
{
    local dst_port = c$id$resp_p;
    
    if ( dst_port in expected_services )
    {
        # We expect certain services on these ports
        # Check if actual service matches
        if ( c?$service && c$service != expected_services[dst_port] )
        {
            print fmt("Port/Service mismatch: %s service on port %s",
                      c$service, dst_port);
        }
    }
}
```

**Why this matters:** We've defined a table mapping ports to the services we expect on them. When a connection uses one of these ports, we check if Zeek's protocol detection identifies it as the expected service. If SSH traffic appears on port 80, or HTTP appears on port 22, something unusual is happening - either misconfiguration or intentional misdirection.

This detection leverages Zeek's deep packet inspection. Unlike simple port-based filtering that just looks at the port number, Zeek analyzes the actual protocol content. The `port` type lets us express the expected relationship between ports and protocols clearly.

### **Practical Helper Functions**

When working with ports in security analysis, you often need to categorize or filter them. Here are common patterns:

**Categorizing ports by range:**

```c
# Function to categorize ports
function categorize_port(p: port): string
{
    local port_num = port_to_count(p);
    
    if ( port_num < 1024 )
        return "well-known";
    else if ( port_num < 49152 )
        return "registered";
    else
        return "ephemeral";
}
```



This function tells you whether a port falls in the well-known range (typically servers), registered range (also servers, but less privileged), or ephemeral range (typically client-side ports for outbound connections). Knowing this helps filter noise - you probably don't care about ephemeral ports as much as well-known ones.

**Identifying commonly targeted ports:**

```c
# Function to check if port is commonly attacked
function is_interesting_port(p: port): bool
{
    return p in set(
        21/tcp,    # FTP
        22/tcp,    # SSH
        23/tcp,    # Telnet
        80/tcp,    # HTTP
        443/tcp,   # HTTPS
        445/tcp,   # SMB
        3389/tcp,  # RDP
        1433/tcp,  # MSSQL
        3306/tcp   # MySQL
    );
}
```

This function checks if a port is in your "interesting" set - ports that are frequently targeted by attackers or represent critical services. You might use this to prioritize alerts, focusing on connection attempts to these ports over connections to random high-numbered ports.

Notice how we're using complete `port` values (with protocols) in the set, not just numbers. This precision matters - 22/tcp is SSH and interesting, but if there were a 22/udp service (uncommon), you might treat it differently.

### **Why This Matters for Security**

The `port` type reflects a fundamental reality of network security: **services are defined by both port and protocol**. Treating ports as just numbers loses critical context and leads to imprecise detections.

Consider scanning detection. If you just count "connections to port 22," you're mixing TCP (SSH) with any UDP traffic that might coincidentally use port 22. Your counts become meaningless. But if you count connections to `22/tcp`specifically, you're tracking actual SSH access attempts - much more valuable.

Or consider allow/deny lists. If you want to permit HTTP but block everything else, you need to specify `80/tcp` and `443/tcp`. Just blocking "port 80" without protocol context could inadvertently block legitimate UDP traffic.

By making protocol an integral part of the `port` type, Zeek ensures your security logic is precise and correct. You're forced to think about ports the way they actually work on networks - as protocol-specific endpoints - rather than as abstract numbers. This design choice prevents countless subtle bugs and makes your detections more accurate.

As you build more sophisticated Zeek scripts, you'll use ports extensively: detecting services on non-standard ports (evasion), tracking which services are most accessed (baseline), identifying port scans (reconnaissance), correlating port usage with protocol detection (validation), and building service-specific detections. The `port` type's protocol awareness makes all of these tasks cleaner and more reliable.




## **The time Type: Timestamps**

The `time` type represents absolute points in time - specific moments on the timeline. If you think of time as a number line stretching from the past into the future, a `time` value is a single point on that line. Zeek uses the `time` type extensively because network security analysis is fundamentally about understanding **when** things happen: when a connection started, when a packet arrived, when an alert fired, when suspicious behavior began.

### **Why Timestamps Matter in Security**

Time is one of the most critical dimensions in security analysis. Attacks unfold over time. Patterns emerge when you look at sequences of events. A single connection might seem innocent, but fifty connections spaced exactly 60 seconds apart suggests beaconing - a hallmark of command-and-control traffic.

Consider what you can detect with accurate timestamps: **brute force attacks** (many attempts in a short window), **beaconing malware** (periodic connections with regular intervals), **data exfiltration** (sustained transfers over time), reconnaissance (rapid connections to many targets), **time-based evasion** (attacks timed to avoid monitoring periods), and **coordinated attacks** (simultaneous activity across multiple hosts).

Without precise time tracking, you're flying blind. The `time` type gives you the foundation to build these temporal detections.




### **Basic Usage: Getting Time Values**

In Zeek scripts, time values typically come from the network events you're analyzing, but you can also get the current time when needed:

```c
# Current time
local now: time = network_time();  # Current time from packet timestamps
local current: time = current_time();  # Actual current time

# Time from events (most common)
event connection_established(c: connection)
{
    local conn_start: time = c$start_time;
    print fmt("Connection started at %s", conn_start);
}

# Specific time (rarely used, usually comes from events)
# Time is represented as seconds since Unix epoch (1970-01-01 00:00:00 UTC)
```

**Understanding network_time() vs current_time():** This distinction is crucial. `network_time()` returns the timestamp from the packet currently being processed - it's the time according to the network traffic itself. `current_time()` returns the actual wall clock time right now.

For almost all security detection logic, **you should use network_time()**. Here's why: When analyzing live traffic, `network_time()` gives you precise timing from the packets themselves, accounting for any processing delays. More importantly, when analyzing saved packet captures (PCAPs) offline, `network_time()` works correctly - it uses the timestamps from when the traffic was originally captured. If you used `current_time()` in offline analysis, all your timing logic would be wrong because you'd be comparing 2024 packet timestamps to 2025 processing timestamps.

Think of `network_time()` as "when did this happen on the network?" and `current_time()` as "what time is it right now in the real world?" For security analysis, you almost always care about the former.

### **Working with Time Values**

The `time` type supports several essential operations that let you build temporal logic:

**Time arithmetic** with intervals lets you calculate future or past moments:

```c
# Time arithmetic
local start: time = network_time();
local duration: interval = 5min;
local end: time = start + duration;  # time + interval = time
```

Adding an interval (a duration) to a time produces a new time. This is useful for calculating expiration times, timeout windows, or future scheduled events. The type system enforces correctness - you can only add intervals to times, not arbitrary numbers.

**Time comparison** tells you the ordering of events:

```c
# Time comparison
if ( end > start )
    print "End is after start";  
    # Obviously true
```

You can check if one event happened before, after, or at the same time as another. This is fundamental for detecting sequences ("Did the login happen before the file access?") or temporal proximity ("Did these two events happen within seconds of each other?").

**Time differences** produce intervals:

```c
# Time difference (produces interval)
local elapsed: interval = end - start;  # time - time = interval
```

Subtracting one time from another gives you the duration between them as an `interval` type. This is how you measure how long something took or how much time elapsed between events.

**Converting time to human-readable strings** for logging or display:

```c
# Convert time to readable string
local time_str = strftime("%Y-%m-%d %H:%M:%S", start);
```

The `strftime()` function formats timestamps using standard format codes - the same ones used in C, Python, and many other languages.


### **Understanding Time's Representation and Precision**

```
┌──────────────────────────────────────────────────────────────┐
│                 TIME TYPE CHARACTERISTICS                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Representation: Double-precision floating point             │
│  └─ Seconds since Unix epoch (Jan 1, 1970)                   │
│  └─ Example: 1696348338.423                                  │
│                                                              │
│  Precision: Microseconds (6 decimal places)                  │
│  └─ Sufficient for sub-millisecond timing analysis           │
│                                                              │
│  Range: ~1970 to ~2106                                       │
│  └─ Sufficient for security monitoring purposes              │
│                                                              │
│  Time vs Current Time:                                       │
│  • network_time(): Time from packet being processed          │
│  • current_time(): Actual clock time now                     │
│                                                              │
│  Almost always use network_time() in detection logic!        │
│  └─ Uses packet timestamps for accurate timing               │
│  └─ Works correctly when analyzing PCAPs offline             │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

Internally, Zeek represents time as a **double-precision floating-point number** storing seconds since the Unix epoch (January 1, 1970, 00:00:00 UTC). A value like `1696348338.423` means "1,696,348,338 seconds and 423 milliseconds since the epoch."

The **precision** is microseconds - six decimal places. This means Zeek can distinguish events that occur within millionths of a second of each other. For network security, this is more than sufficient. Most network events are separated by milliseconds or more, and even microsecond precision is overkill for detecting most attacks. However, having this precision ensures you never lose timing information due to rounding.

The **range** extends from approximately 1970 to 2106 - far beyond any practical security monitoring timeframe. You'll never hit range limits in normal operation.

**Key principle:** Always prefer `network_time()` over `current_time()` in detection logic. It works correctly both in real-time monitoring and offline PCAP analysis, and it gives you the actual time events occurred on the network rather than when Zeek processed them.

### **Real-World Example: Detecting Beaconing Behaviour**

One of the most powerful applications of time analysis is detecting beaconing - when malware periodically "phones home" to a command-and-control server. Beaconing creates a distinctive pattern: connections with regular, predictable intervals. Here's one way we can detect it:

```c
# Track connection timing to detect beaconing C2
global last_connection_time: table[addr] of time;
global connection_intervals: table[addr] of vector of interval;

event connection_established(c: connection)
{
    local src = c$id$orig_h;
    local now = network_time();
    
    # If we've seen this IP before
    if ( src in last_connection_time )
    {
        # Calculate interval since last connection
        local interval_between = now - last_connection_time[src];
        
        # Store interval
        if ( src !in connection_intervals )
            connection_intervals[src] = vector();
        
        connection_intervals[src][|connection_intervals[src]|] = interval_between;
        
        # If we have enough samples, check for regularity
        if ( |connection_intervals[src]| >= 10 )
        {
            # Analyze intervals for consistency (beaconing indicator)
            local intervals = connection_intervals[src];
            local sum: interval = 0sec;
            local i: count;
            
            for ( i in intervals )
                sum += intervals[i];
            
            local avg = sum / |intervals|;
            
            # Check if intervals are consistent (low jitter = beaconing)
            # This is simplified; real detection would calculate std deviation
            print fmt("Average interval for %s: %s", src, avg);
        }
    }
    
    # Update last seen time
    last_connection_time[src] = now;
}
```

**Understanding this detection:** We're tracking when each source IP makes connections. For each IP, we store the time of its last connection and build a history of the intervals between consecutive connections. Once we have at least 10 samples, we calculate the average interval.

The key insight is this: **legitimate user traffic has irregular timing**, but **automated malware beaconing has regular timing**. A human browsing the web might connect at intervals like 3s, 45s, 2s, 120s - highly variable. But malware configured to beacon every 60 seconds will produce intervals like 60.1s, 59.8s, 60.2s, 60.0s - remarkably consistent.

In a production version, you'd calculate standard deviation to measure consistency mathematically. Low standard deviation relative to the mean indicates regular beaconing. High standard deviation indicates normal irregular human behavior.

This example showcases the power of the `time` type: we're doing precise timestamp arithmetic (subtracting times to get intervals), aggregating temporal data over multiple observations, and using statistical analysis to detect patterns invisible in individual events.

### **Why This Matters for Security**

The `time` type isn't just about knowing when things happened - it's about understanding the **temporal dimension of threats**. Attacks have timing characteristics:

- **Speed:** How quickly did the attacker move from reconnaissance to exploitation?
- **Persistence:** How long has this suspicious activity been going on?
- **Periodicity:** Does this behavior repeat on a schedule?
- **Sequence:** Did event A happen before event B, establishing causality?
- **Duration:** How long did this connection or session last?
- **Clustering:** Are events happening in suspicious bursts or patterns?

All of these questions require precise time tracking and the ability to perform temporal arithmetic and comparisons. Zeek's `time` type, combined with the `interval` type (which we'll cover next), gives you the tools to build detections that understand time as more than just a log field - time becomes a dimension you can analyze, correlate, and use to distinguish attacks from normal activity.

As you develop more sophisticated Zeek scripts, temporal analysis will become one of your most powerful techniques. Master the `time` type, and you unlock an entire category of detections impossible with simple signature-based approaches.





---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./type.md" >}})
[|NEXT|]({{< ref "./complex.md" >}})

