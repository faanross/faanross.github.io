---
showTableOfContents: true
title: "The table Type: Key-Value Mappings"
type: "page"
---
## The Table Type: Zeek's Key-Value Mapping Powerhouse

In network security monitoring, we constantly need to track state - which IP addresses have we seen? How many times has port 80 been accessed? Which hosts are exhibiting suspicious behavior? Zeek's **table** type is the fundamental data structure for maintaining this kind of **associative memory**. Unlike arrays that use sequential numeric indices, tables map arbitrary **keys** to **values**, functioning like a dictionary in Python, a map in Golang, a hash map in Java, or an object in JavaScript.

Tables are not merely convenient - they're essential for stateful network analysis. Without them, Zeek would be limited to stateless pattern matching, unable to correlate events over time or track evolving attack patterns. Understanding tables is crucial for writing effective Zeek scripts, especially for detection logic that must maintain memory across thousands or millions of network events.


## The Core Concept: Associative Arrays

At its heart, a **table** is an **associative array**: a collection of key-value pairs where each unique key maps to exactly one value. The key can be any Zeek type (addresses, ports, strings, numbers), and the value can be any type as well. This flexibility makes tables extraordinarily versatile.

**Why tables matter for security monitoring:**

- **Stateful tracking**: Remember which hosts you've seen before without rescanning the entire dataset
- **Aggregation**: Count occurrences (failed logins, port scans, DNS queries) indexed by source
- **Enrichment**: Map IP addresses to threat intelligence labels or geographic locations
- **Correlation**: Link related events across time by maintaining lookup structures
- **Memory efficiency**: Built-in expiration mechanisms prevent unbounded memory growth in long-running scripts

## Basic Table Declaration and Operations

### Simple Table Creation

Let's start with the fundamentals. A table declaration specifies both the **key type** (what you'll use to look things up) and the **value type** (what data you'll store):


```c
# Declare a table that maps port numbers to counts
# Key type: port (e.g., 80/tcp, 443/tcp)
# Value type: count (non-negative integer)
# table[key] of value

global port_counts: table[port] of count;
```

This creates an empty table in global scope. The `global` keyword means it persists across events and is accessible from any function or event handler.


### Initialization with Literal Values

You can initialize tables with data at declaration time, useful for **threat intelligence feeds** or **configuration lookups**:

```c
# Initialize table with known malicious infrastructure
# Each IP address (key) maps to a description string (value)
global known_malicious_ips: table[addr] of string = {
    [192.0.2.1] = "Known C2 server",          
    [198.51.100.1] = "Malware distribution",  
    [203.0.113.1] = "Phishing site",          
};
```

**Key syntax notes:**

- Keys are enclosed in square brackets: `[192.0.2.1]`
- The `=` assigns the value for that key
- Multiple entries are comma-separated
- This initialization happens once when the script loads

### Adding and Accessing Entries

Tables support intuitive assignment and lookup syntax similar to arrays:

```c
# Add entries dynamically
# HTTP connections
port_counts[80/tcp] = 100;   
# HTTPS connections
port_counts[443/tcp] = 250;  

# Access entries (retrieve value by key)
local http_count = port_counts[80/tcp];  
# http_count = 100

# Accessing non-existent keys causes runtime error!
# ERROR: key not found
```

**Critical safety consideration**: Accessing a key that doesn't exist triggers a **runtime error** that can crash your script. Always check for key existence first (see below) or use the `&default` attribute.


### Membership Testing: The `in` Operator

Before accessing a key, verify it exists using the **`in` operator**, which returns a boolean:

```c
# Safe access pattern: check before accessing
if ( 80/tcp in port_counts )
    print "Have count for HTTP";
else
    print "HTTP not tracked yet";

# Common idiom: check-then-increment
if ( 80/tcp in port_counts )
    ++port_counts[80/tcp];
else
    port_counts[80/tcp] = 1;  # Initialize if first occurrence
```

The `in` operator is **essential for defensive programming** in Zeek - it prevents the most common source of script crashes.


### Removing Entries

The **`delete` statement** removes a key-value pair from the table:

```c
# Remove the entry for port 80
delete port_counts[80/tcp];

# After deletion, key no longer exists
if ( 80/tcp in port_counts )
    print "Still there";  # This won't execute
else
    print "Deleted successfully";
```

Manual deletion is rarely needed due to Zeek's automatic expiration mechanisms (covered below), but it's useful for immediate cleanup or when maintaining precise control over memory.

## Compound Keys: Multi-Dimensional Indexing

One of Zeek's most powerful features is **compound keys** - the ability to use multiple values as a composite key. This enables multi-dimensional indexing without nested tables.



### Declaring Tables with Multiple Key Types


```c
# Table indexed by BOTH IP address AND port
# This tracks per-IP, per-port connection counts
global connections: table[addr, port] of count;
```

The key type is now `[addr, port]` - a **tuple of two types**. You can have any number of key components: `[addr, addr, port, string]` would be a four-element compound key.

### Using Compound Keys

When accessing or assigning with compound keys, provide all key components in square brackets:


```c
# Add entries with compound keys
# Host .100 on HTTP
connections[192.168.1.100, 80/tcp] = 5;    
# Same host on HTTPS
connections[192.168.1.100, 443/tcp] = 10;  
# Different host on HTTP
connections[192.168.1.200, 80/tcp] = 3;    

# Access with all key components
if ( [192.168.1.100, 80/tcp] in connections )
{
    local count = connections[192.168.1.100, 80/tcp];  
    # count = 5
    print fmt("Host has %d HTTP connections", count);
}
```

**Why compound keys are powerful:**

- **Natural representation**: Models network tuples (src_ip, dst_ip, port) directly
- **Avoids nested structures**: `table[addr, port]` is simpler than `table[addr] of table[port]`
- **Efficient lookup**: Single hash operation instead of two table lookups
- **Atomic operations**: Check or update multi-dimensional state in one step

## TO BE CONTINUED...

---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./intro.md" >}})
[|NEXT|]({{< ref "./set.md" >}})

