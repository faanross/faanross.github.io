---
showTableOfContents: true
title: "The subnet Type: Network Ranges"
type: "page"
---




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





### Subnet Masking and Address Aggregation

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



## Knowledge Check: subnet Type

**Q1: What does the "/24" mean in the subnet notation "192.168.1.0/24"? How many addresses does this subnet contain?**

A: The "/24" is the prefix length - the number of network bits. It means the first 24 bits (first three octets) define the network portion. A /24 subnet contains 256 addresses (2^8, since 32-24=8 bits remain for hosts). More generally, a /n prefix leaves (32-n) bits for host addresses in IPv4.

**Q2: Explain why "ip in subnet" is such a powerful and frequently used operation in security detection. Provide a concrete example.**

A: The `in` operator performs all the binary mathematics (converting address to binary, applying subnet mask, comparing network portions) behind the scenes in a single, readable expression. For example, `if ( ip in 192.168.0.0/16 )` instantly tells you if an IP is in your private network, which is fundamental for distinguishing "our network" from "the outside world" - a categorization needed in almost all security detections.

**Q3: What is the purpose of the `mask_addr()` function, and when would you use it?**

A: `mask_addr()` extracts the network portion of an IP address, converting it into its parent subnet address. For example, `mask_addr(192.168.1.100, 24)` returns `192.168.1.0`. This is useful for aggregating statistics by network block (tracking connections per /24 instead of per IP) or grouping related addresses, which helps identify patterns like subnet-wide scanning or compromised network ranges.

**Q4: Why is defining a set of subnets representing "local networks" typically one of the first things you do in a Zeek deployment?**

A: Distinguishing internal traffic from external traffic is a fundamental categorization for almost all security detections. Many detections only care about one direction (e.g., outbound connections to certain ports might indicate data exfiltration, while inbound might be different concerns). Having a centralized definition of local networks ensures consistency across all scripts and makes updates easy when network topology changes.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./addr.md" >}})
[|NEXT|]({{< ref "./port.md" >}})

