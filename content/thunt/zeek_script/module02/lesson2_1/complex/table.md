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


### Practical Example: Tracking Per-Host, Per-Service Connections



```c
# Monitor which services each host accesses
global service_usage: table[addr, port] of count;

# In a connection event handler
event connection_established(c: connection)
{
	# Source IP
    local src = c$id$orig_h;
    # Destination port   
    local dst_port = c$id$resp_p;  
    
    # Increment or initialize count for this host/service pair
    if ( [src, dst_port] !in service_usage )
        service_usage[src, dst_port] = 0;
    
    ++service_usage[src, dst_port];
    
    # Detect port scanning: one host touching many services
    local ports_accessed = 0;
    for ( [ip, port] in service_usage )
    {
        if ( ip == src )
            ++ports_accessed;
    }
    
    if ( ports_accessed > 20 )
        print fmt("ALERT: %s scanned %d ports", src, ports_accessed);
}
```


## Iterating Over Tables

Tables support **iteration** over their keys, allowing you to process all entries:

### Simple Table Iteration


```c
# Iterate over single-key table
for ( port in port_counts )
{
    # 'port' takes on each key value in turn
    print fmt("Port %s: %d connections", port, port_counts[port]);
}
```

**Output example:**

```bash
Port 80/tcp: 100 connections
Port 443/tcp: 250 connections
Port 22/tcp: 50 connections
```

**Important characteristics:**

- Iteration order is **undefined** (tables are unordered collections)
- The loop variable (`port`) receives each key
- Access the value using the key: `port_counts[port]`


### Compound Key Iteration with Destructuring

When iterating over compound keys, you can **destructure** the key tuple into separate variables:


```c
# Iterate over table with compound keys
for ( [ip, port] in connections )
{
    # 'ip' and 'port' are automatically unpacked from the compound key
    print fmt("%s:%s has %d connections", 
              ip, port, connections[ip, port]);
}
```

**Output example:**

```bash
192.168.1.100:80/tcp has 5 connections
192.168.1.100:443/tcp has 10 connections
192.168.1.200:80/tcp has 3 connections
```

The syntax `[ip, port]` in the loop header **destructures** the compound key into its components, making the code cleaner than manual tuple handling.


## Automatic Expiration: Memory Management for Long-Running Analysis

In production network monitoring, tables can grow unbounded - a busy network might see millions of unique IP addresses in a day. **Automatic expiration** is Zeek's solution: entries can be configured to **delete themselves** after a time period, preventing memory exhaustion.

### Basic Time-Based Expiration

The **`&create_expire`** attribute sets a **sliding timeout window** - each entry is deleted if not accessed within the specified interval:


```c
# Entries automatically deleted after 1 hour of inactivity
global recent_scanners: table[addr] of count 
	# Time interval: 1 hour
    &create_expire = 1hr;  
```

**How expiration works:**

- When you add an entry, Zeek starts a timer for `create_expire` duration
- **Each access** to the entry (read or write) **resets the timer**
- If the timer expires without access, the entry is automatically deleted
- This is a **sliding window** - active entries never expire, idle ones do


**Time interval syntax:**

- `1sec`, `30sec` (seconds)
- `5min`, `30min` (minutes)
- `1hr`, `24hr` (hours)
- `1day`, `7days` (days)



### Why Expiration is Critical


```c
# Without expiration: memory grows forever (BAD)
global all_ips_ever: table[addr] of count;  

# With expiration: only recent activity tracked (GOOD)
global recent_ips: table[addr] of count 
    &create_expire = 1hr;  
```

On a busy network seeing 10,000 unique IPs per hour, the first table would grow to millions of entries within days. The second never exceeds ~10,000 entries because old IPs automatically expire.

### Custom Expiration Logic

For advanced cases, **`&expire_func`** lets you dynamically adjust expiration time per entry:


```c
global suspicious_ips: table[addr] of count
	# Default expiration
    &create_expire = 30min  
    &expire_func = function(t: table[addr] of count, idx: addr): interval
    {
        # Function called when entry is about to expire
        # Can extend expiration based on entry's value
        
        if ( t[idx] > 100 )
            return 2hr;   # High activity: keep longer
        else
            return 30min; # Low activity: expire sooner
    };
```

**Expiration function parameters:**

|Parameter|Type|Purpose|
|---|---|---|
|`t`|`table[addr] of count`|Reference to the entire table|
|`idx`|`addr`|The specific key being evaluated for expiration|
|**Returns**|`interval`|New expiration duration (or `0sec` to delete immediately)|

**Use cases for custom expiration:**

- **Adaptive tracking**: Keep high-severity threats in memory longer
- **Rate limiting**: Expire entries early if table is growing too large
- **Graduated response**: Short timeout for low-risk indicators, long timeout for confirmed threats


## Default Values: Simplifying Initialization

The **`&default`** attribute provides a fallback value for missing keys, eliminating the need for explicit existence checks:

```c
# Table with default value of 0
global ssh_failed_attempts: table[addr] of count 
	# Non-existent keys return 0
    &default = 0;  

event ssh_auth_failed(c: connection, user: string)
{
    local src = c$id$orig_h;
    
    # No need to check if key exists!
    # Accessing missing key returns default (0), then we increment
    # Safe even on first access
    ++ssh_failed_attempts[src];  
    
    if ( ssh_failed_attempts[src] >= 5 )
    {
        print fmt("ALERT: %s has %d failed SSH attempts",
                  src, ssh_failed_attempts[src]);
    }
}
```

**Without `&default`**, you'd need:


```c
if ( src !in ssh_failed_attempts )
    ssh_failed_attempts[src] = 0;
++ssh_failed_attempts[src];  
```

**With `&default`**, this boilerplate vanishes - the table automatically initializes missing keys with the default value.




## Real-World Example: Brute Force Detection

Here's a complete, production-ready pattern combining expiration and defaults:


```c
# Track failed SSH login attempts with automatic expiration
global ssh_failed_attempts: table[addr] of count 
	# Reset count after 1 hour of inactivity
    &create_expire = 1hr  
    # Missing keys default to 0 
    &default = 0;          

event ssh_auth_failed(c: connection, user: string)
{
	# Attacker's IP
    local src = c$id$orig_h;  
    
    # Increment failure count (default=0 makes this safe)
    ++ssh_failed_attempts[src];
    
    # Check threshold
    if ( ssh_failed_attempts[src] >= 5 )
    {
        print fmt("ALERT: %s has %d failed SSH attempts in last hour",
                  src, ssh_failed_attempts[src]);
        
        # Could trigger IDS alert, firewall block, etc.
    }
}
```

**Why this pattern is effective:**

- **Automatic cleanup**: After 1 hour of no activity, the IP's count resets (expiration)
- **No initialization code**: First failed login automatically creates entry with count=1
- **Memory bounded**: Table only contains IPs with failed logins in the last hour
- **Simple logic**: No manual reset or cleanup required


## Table Size Management and Limits

### Querying Table Size

Use the **size operator `| |`** to get the number of entries:

```c
local size = |port_counts|;  # Returns count of entries
print fmt("Tracking %d ports", size);
```

This is useful for monitoring memory usage or detecting abnormal growth.


### Clearing All Entries

The **`clear_table()`** built-in function removes all entries:


```c
# Remove everything from the table
clear_table(port_counts);

# After clearing
# Prints 0
print |port_counts|;  
```

**Use case**: Periodic resets, like daily statistics that clear at midnight.



### Enforcing Size Limits

For critical memory control, **`&max_size`** caps the maximum number of entries:

```c
global large_table: table[addr] of count
    &create_expire = 1hr
    &max_size = 10000   
    &on_size_limit = function(t: table[addr] of count)
    {
        # Called when table reaches max_size
        print "Table size limit reached, clearing old entries";
        
        # Could implement LRU eviction, selective pruning, etc.
        clear_table(t);  # Simple approach: clear everything
    };
```

**Size limit behavior:**

|Scenario|Behavior|
|---|---|
|Table has < `max_size` entries|New entries added normally|
|Table reaches `max_size`|`&on_size_limit` function is called|
|`&on_size_limit` not specified|**Runtime error** if size exceeded|
|`&on_size_limit` specified|Script handles overflow (clear, evict, alert, etc.)|

**Critical for production**: Always set `&max_size` for tables that could grow unbounded (IP tracking, connection state, etc.) to prevent out-of-memory crashes.




## Advanced Table Patterns

### Nested Tables: Tables of Tables

Sometimes you need hierarchical structures - for example, tracking connections per service, per IP:


```c
# Outer table: service name → inner table
# Inner table: IP address → connection count
global connections_per_service: table[string] of table[addr] of count;

# Helper function to safely add entries
function track_connection(service: string, ip: addr)
{
    # Check if service entry exists
    if ( service !in connections_per_service )
	    # Create inner table
        connections_per_service[service] = table();  
    
    # Check if IP entry exists in inner table
    if ( ip !in connections_per_service[service] )
        connections_per_service[service][ip] = 0;
    
    # Increment count
    ++connections_per_service[service][ip];
}

# Usage
track_connection("http", 192.168.1.100);
track_connection("http", 192.168.1.100);
track_connection("ssh", 192.168.1.200);

# Access: connections_per_service["http"][192.168.1.100] == 2
```

**When to use nested tables vs. compound keys:**

|Pattern|Best For|Example|
|---|---|---|
|**Compound keys** `table[addr, port]`|Fixed relationships, simple queries|Per-host, per-service tracking|
|**Nested tables** `table[string] of table[addr]`|Dynamic outer keys, need to iterate outer level|Per-service aggregation|

**Trade-off**: Nested tables require more initialization code but allow independent expiration/sizing of inner tables.




### Tables of Complex Values

Tables can store structured data using **record types**:


```c
# Define a record type for connection metadata
type ConnectionInfo: record {
    first_seen: time;    
    last_seen: time;     
    total_bytes: count;  
};

# Table mapping IP to connection info
global conn_tracking: table[addr] of ConnectionInfo;

# Add an entry
conn_tracking[192.168.1.100] = ConnectionInfo(
    $first_seen = network_time(),
    $last_seen = network_time(),
    $total_bytes = 0
);

# Update an existing entry
if ( 192.168.1.100 in conn_tracking )
{
    conn_tracking[192.168.1.100]$last_seen = network_time();
    conn_tracking[192.168.1.100]$total_bytes += 1024;
}
```

**Advantages of record-valued tables:**

- **Semantic clarity**: Named fields are more readable than parallel tables
- **Atomic updates**: All related data stored together
- **Type safety**: Compiler ensures field types are correct
- **Extensibility**: Easy to add new fields to the record definition





## Performance and Best Practices


### Memory Management Guidelines

**DO:**

- ✓ Always use `&create_expire` for tables tracking network state
- ✓ Set `&default` when values have natural initialization (counts, booleans)
- ✓ Use `&max_size` for tables that could grow to millions of entries
- ✓ Prefer compound keys over nested tables when possible

**DON'T:**

- ✗ Create tables without expiration for per-packet or per-connection state
- ✗ Access keys without checking existence (unless `&default` is set)
- ✗ Store entire connection records - store connection UIDs and look up in conn.log
- ✗ Use tables for small, fixed lookups (use sets or hardcoded if-else)

### Security Considerations

**Offensive perspective (attackers exploiting Zeek):**

- **Memory exhaustion attacks**: Send millions of unique IPs to overflow tables without expiration
- **Hash collision attacks**: Craft keys that hash to same bucket, degrading lookup to O(n)
- **Expiration timing attacks**: Trigger activity just before expiration to extend tracking indefinitely

**Defensive perspective (protecting Zeek):**

- **Always set `&max_size`**: Prevents unbounded growth from preventing monitoring
- **Use appropriate expiration**: Balance detection window vs. memory constraints
- **Monitor table sizes**: Alert if tables grow unexpectedly large
- **Rate limit table additions**: In `&on_size_limit`, could drop new entries rather than clearing




## Summary: Tables as the Foundation of Stateful Analysis

Tables are not just a data structure - they're the **fundamental mechanism** by which Zeek maintains state across network events. Every sophisticated detection pattern relies on tables:

- **Anomaly detection**: Track baselines per host, service, or application
- **Threat correlation**: Link related indicators (IP → domain → hash) across time
- **Behavioral analysis**: Model normal activity patterns to detect deviations
- **Attack tracking**: Maintain state machines for multi-stage attacks

Understanding tables deeply - their syntax, expiration semantics, compound keys, and performance characteristics - is essential for writing production-quality Zeek scripts that are both powerful and safe. With proper expiration and size limits, tables provide unbounded analytical capabilities within bounded memory constraints, making them the workhorse of network security monitoring.



## Knowledge Check: table Type

**Q1: Why is checking key existence with the `in` operator critical before accessing table values, even though some other languages allow direct access that returns null/undefined?**

A: Accessing a non-existent key in a Zeek table causes a **runtime error that crashes your script**. Unlike languages that return null/undefined, Zeek enforces explicit existence checking to prevent logic bugs where you inadvertently process missing data as if it were present. This design choice trades convenience for reliability - forcing you to consciously handle the "key doesn't exist" case prevents an entire class of subtle bugs in security monitoring where silently treating missing data as present could cause false negatives.



**Q2: What is the fundamental difference between `&create_expire` and `&read_expire`, and when would you choose each for a production deployment?**

A: `&create_expire` starts the expiration timer when an entry is **first created** and never resets it - entries are deleted after a fixed time regardless of whether they're accessed. Use this for **time-windowed detection** (e.g., "failed logins in the last hour") where you want to count events within a specific timeframe.

`&read_expire` resets the timer **every time the entry is accessed** - active entries stay alive indefinitely. Use this for **activity-based tracking** (e.g., "maintain state for IPs we're currently seeing") where ongoing activity should extend tracking. Choosing wrong can mean either premature expiration of active threats or unbounded memory growth for inactive entries.


**Q3: How do compound keys in Zeek tables differ from nested tables, and why is `table[addr, port]` preferable to `table[addr] of table[port]`?**

A: Compound keys (`table[addr, port]`) create a **single-level table** indexed by a tuple, requiring one  operation for lookup. Nested tables (`table[addr] of table[port]`) create a **two-level structure** requiring two  operations and explicit initialization of inner tables. Compound keys are simpler (no inner table management), faster (one lookup instead of two), more memory-efficient (one hash table instead of potentially thousands of tiny inner tables), and eliminate the entire class of bugs related to forgetting to initialize inner tables. Use compound keys whenever possible for multi-dimensional indexing.




**Q4: Why must production tables tracking network state always include expiration and/or size limits, and what happens if you don't?**

A: Tables without expiration **grow unbounded** as they accumulate entries for every unique key encountered. In a high-volume network, a table tracking per-IP state could easily reach millions of entries consuming gigabytes of RAM, eventually causing Zeek to exhaust memory and crash with an OOM error.

This stops all monitoring - a **complete security blind spot**. Expiration (`&create_expire`, `&read_expire`) provides time-based cleanup; `&max_size` provides a hard cap as a safety net. Together, they ensure your monitoring system remains stable under all traffic conditions, including potential memory exhaustion attacks.

**Q5: When would you use `&default` on a table, and how does it change the safety requirements for accessing table values?**

A: Use `&default` when table values have a **natural zero/empty state** that's meaningful for your logic - particularly for `count` types that start at zero, `bool` flags that default to false, or empty collections. With `&default` set, accessing a non-existent key returns the default value instead of crashing, eliminating the need for existence checking in initialization-then-increment patterns. However, you lose the distinction between "key never seen" and "key seen but has default value" - if that distinction matters for your detection logic, don't use `&default` and check existence explicitly.








---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./intro.md" >}})
[|NEXT|]({{< ref "./set.md" >}})

