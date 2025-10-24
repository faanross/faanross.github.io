---
showTableOfContents: true
title: "The port Type: Network Ports"
type: "page"
---







## The port Type: Network Ports

The `port` type represents network ports - those numbers between 0 and 65535 that identify specific services or applications on a host. But Zeek's `port` type does something crucial that simple integers can't: it **binds the port number to its transport protocol** (TCP or UDP). This protocol awareness is essential because port 80 over TCP (HTTP web traffic) is completely different from port 80 over UDP (which might be something else entirely).

### **Why Protocol Context Matters**

In the real world of network security, port numbers without protocol context are almost meaningless. When you see traffic on port 53, you need to know: is it 53/tcp or 53/udp? DNS primarily uses UDP, so 53/udp is normal DNS traffic. But 53/tcp is typically DNS zone transfers or responses too large for UDP - much rarer and potentially interesting from a security perspective.

Similarly, 80/tcp is standard HTTP web traffic you'd expect everywhere. But 80/udp? That's unusual and worth investigating. By making protocol an intrinsic part of the `port` type, Zeek ensures you're always working with the complete picture. You can't accidentally compare TCP ports to UDP ports or forget to check which protocol you're dealing with - the type system enforces correctness.

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

This distinction prevents an entire class of bugs. If you're checking for HTTP traffic on port 80, you're checking for `80/tcp`, and you won't accidentally match unrelated UDP traffic that happens to use the same port number.

### **Working with Port Values**

The `port` type supports several useful operations for security analysis:

**Comparison** is straightforward - you can check if a port matches a specific value:

```c
# Comparison
if ( http_port == 80/tcp )
    print "Standard HTTP port";
```

**Extracting the numeric portion** when you need to do arithmetic or range checks:

```c
# Extract port number
local port_num = port_to_count(443/tcp);  
# Returns 443 as count
```

The `port_to_count()` function gives you the raw port number as a `count` type, which you can then use in numeric comparisons or calculations. This is useful when you need to categorize ports by range (well-known, registered, ephemeral) or perform other numeric operations.

**Protocol information** is intrinsic to the port type itself. While you can't extract the protocol as a separate string value directly, you test against protocol-specific port values to determine what you're dealing with.

**Range checking** helps categorize ports by their IANA designation:

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

**Understanding this detection:** Zeek's HTTP analyzer has identified HTTP protocol traffic based on the actual content of the packets - it's truly HTTP regardless of port. But the server is listening on something other than the standard ports 80, 443, or 8080. This could be legitimate (maybe it's a development server on port 8000), or it could be an attacker trying to evade simple port-based filtering.

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

**Why this matters:** We've defined a table mapping ports to the services we expect on them. When a connection uses one of these ports, we check if Zeek's protocol detection identifies it as the expected service. If SSH traffic appears on port 80, or HTTP appears on port 22, something unusual is happening - either misconfiguration or intentional misdirection.

This detection leverages Zeek's deep packet inspection. Unlike simple port-based filtering that just looks at the port number, Zeek analyzes the actual protocol content. The `port` type lets us express the expected relationship between ports and protocols clearly.

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

Notice how we're using complete `port` values (with protocols) in the set, not just numbers. This precision matters - 22/tcp is SSH and interesting, but if there were a 22/udp service (uncommon), you might treat it differently.

### **Why This Matters for Security**

The `port` type reflects a fundamental reality of network security: **services are defined by both port and protocol**. 
Treating ports as just numbers loses critical context and leads to imprecise detections.

Consider scanning detection. If you just count "connections to port 22," you're mixing TCP (SSH) with any UDP traffic that 
might coincidentally use port 22. Your counts become meaningless. But if you count connections to `22/tcp` specifically, you're tracking actual SSH access attempts - much more valuable.

Or consider allow/deny lists. If you want to permit HTTP but block everything else, you need to specify `80/tcp` and `443/tcp`. Just blocking "port 80" without protocol context could inadvertently block legitimate UDP traffic.

By making protocol an integral part of the `port` type, Zeek ensures your security logic is precise and correct. You're forced to think about ports the way they actually work on networks - as protocol-specific endpoints - rather than as abstract numbers. This design choice prevents countless subtle bugs and makes your detections more accurate.

As you build more sophisticated Zeek scripts, you'll use ports extensively: 
- Detecting services on non-standard ports (**evasion**), 
- Tracking which services are most accessed (**baseline**), 
- Identifying port scans (**reconnaissance**), 
- Correlating port usage with protocol detection (**validation**), and 
- Building service-specific detections. 


The `port` type's protocol awareness makes all of these tasks cleaner and more reliable.


## Knowledge Check: port Type

**Q1: Why does Zeek's port type bind the port number to its transport protocol rather than just storing a number?**

A: Because port numbers without protocol context are almost meaningless in network security. Port 53/udp (normal DNS) is completely different from 53/tcp (DNS zone transfers). Port 80/tcp (HTTP) is different from 80/udp (uncommon, potentially suspicious). By making protocol intrinsic to the port type, Zeek ensures you're always working with the complete picture and prevents accidentally comparing TCP ports to UDP ports.

**Q2: Are `80/tcp` and `80/udp` considered the same value or different values in Zeek? Why does this matter?**

A: They are different values. This distinction prevents an entire class of bugs. If you're checking for HTTP traffic on port 80, you're checking for `80/tcp` specifically, and you won't accidentally match unrelated UDP traffic that happens to use the same port number. This precision is essential for accurate detection.

**Q3: What does the `port_to_count()` function do, and when would you use it?**

A: It extracts the numeric portion of a port value and returns it as a count type. You use this when you need to perform numeric operations like range checking (is the port in the ephemeral range 49152-65535?) or categorization (well-known ports < 1024, registered ports 1024-49151, dynamic ports >= 49152).

**Q4: Describe what "protocol/service mismatch detection" means and why the port type's design makes it possible.**

A: It means detecting when a service is running on an unexpected port - like HTTP traffic on port 22 or SSH traffic on port 80. Zeek's deep packet inspection identifies the actual protocol, while the port type tells you what port it's using. By comparing expected service-to-port mappings against what Zeek actually detects, you can identify potential evasion techniques or misconfigurations. The port type's protocol awareness is crucial for expressing these expected relationships clearly.


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./subnet.md" >}})
[|NEXT|]({{< ref "./time.md" >}})

