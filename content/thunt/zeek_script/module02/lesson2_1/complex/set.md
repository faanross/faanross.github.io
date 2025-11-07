---
showTableOfContents: true
title: "The set Type: Unique Collections"
type: "page"
---

# The `set` Type: Unique Collections in Zeek

In network security monitoring, one of the most common questions we need to answer is: "Have I seen this before?" Whether tracking IP addresses, domain names, or connection patterns, we constantly need to maintain collections of unique items and test for membership. Zeek's **`set` type** is specifically designed for this purpose - it provides a high-performance data structure that stores unique values with lightning-fast membership testing, making it ideal for deduplication, tracking state, and detecting anomalies based on uniqueness.

Unlike **tables** which map keys to values, sets simply maintain a collection of unique elements with no associated data. Think of a set as a specialized hash table where we only care about presence or absence, not about storing additional information. This simplicity makes sets both memory-efficient and extremely fast for their intended use cases.


## Fundamental Set Characteristics

Sets in Zeek have several defining properties that make them distinct from other collection types:

- **Uniqueness guarantee**: A set can contain each value only once; attempting to add a duplicate has no effect
- **Unordered storage**: Elements have no defined sequence or indexing; iteration order is implementation-dependent
- **O(1) membership testing**: Checking if an element exists in a set is a constant-time operation regardless of set size
- **Homogeneous typing**: All elements must be of the same type or compatible types (e.g., all `addr`, all `string`)
- **Memory efficient**: Sets only store the values themselves, not key-value pairs like tables

These characteristics make sets the optimal choice whenever you need a "seen it" tracker, a whitelist/blacklist, or any scenario where you're answering binary membership questions.



## Basic Set Operations

### Declaration and Initialization

Creating and populating sets follows Zeek's standard variable declaration syntax, with optional immediate initialization:

```c
# Declare an empty set of IP addresses
global seen_ips: set[addr];

# Declare and initialize a set with literal values
global local_subnets: set[subnet] = {
    10.0.0.0/8,        # Class A private range
    172.16.0.0/12,     # Class B private range
    192.168.0.0/16,    # Class C private range
};

# Set of suspicious ports to monitor
global monitored_ports: set[port] = {
    22/tcp, 23/tcp, 3389/tcp, 5900/tcp
};

# Set of known-good domains (whitelist)
global trusted_domains: set[string] = {
    "google.com",
    "microsoft.com",
    "cloudflare.com"
};
```

**Key initialization details:**

|Aspect|Behavior|Notes|
|---|---|---|
|**Empty declaration**|Creates set with zero elements|Must specify type in brackets|
|**Literal initialization**|Uses curly brace syntax `{ ... }`|Comma-separated values|
|**Type enforcement**|All elements must match declared type|Compile-time type checking|
|**Global vs local**|Can declare at global or local scope|Global sets persist across events|



### Adding and Removing Elements

The **`add`** statement inserts elements into a set, while **`delete`** removes them. Both operations are idempotent - adding an existing element or deleting a non-existent one has no effect and raises no error:

```c
# Add elements to the set
add seen_ips[192.168.1.100];
add seen_ips[192.168.1.101];
add seen_ips[192.168.1.102];

# Adding a duplicate has no effect (silently ignored)
add seen_ips[192.168.1.100];  # Set still contains only 3 elements

# Remove an element
delete seen_ips[192.168.1.100];  # Set now contains 2 elements

# Deleting non-existent element is safe (no error)
delete seen_ips[10.0.0.1];  # No-op if not present

# Get the current size of the set
local count = |seen_ips|;  # Returns 2
print fmt("Set contains %d unique IPs", count);
```

**Important behavioral notes:**

- **Idempotency**: Both `add` and `delete` can be called multiple times safely without conditional checks
- **No exceptions**: Unlike some languages, Zeek doesn't throw errors for duplicate adds or invalid deletes
- **Atomic operations**: Set modifications are atomic from the perspective of the Zeek event engine
- **Size operator**: The `|set|` operator returns the cardinality (number of unique elements) in O(1) time

### Membership Testing: The Core Use Case

The primary reason to use sets is their **extremely fast membership testing** using the **`in`** operator. This operation is O(1) constant time - checking membership in a set of 10 elements takes the same time as checking a set of 10 million elements:

```c
# Check if an IP has been seen before (very fast!)
if ( 192.168.1.100 in seen_ips )
{
    print "Already seen this IP - possible repeat visitor";
}
else
{
    print "First time seeing this IP";
    add seen_ips[192.168.1.100];  # Add it for next time
}

# Check if subnet is in private ranges
if ( 192.168.50.0/24 in local_subnets )
{
    print "This is a private subnet";
}

# Negative membership check
if ( c$id$orig_h !in seen_ips )
{
    print "New IP detected!";
    add seen_ips[c$id$orig_h];
}
```



**Performance characteristics:**

```
┌─────────────────────────────────────────────────────────────┐
│            SET MEMBERSHIP TESTING PERFORMANCE               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Operation: if ( x in my_set )                              │
│                                                             │
│  Time Complexity:    O(1) - constant time                   │
│  Set size:           Doesn't matter!                        │
│                                                             │
│  10 elements:        ~100 nanoseconds                       │
│  1,000 elements:     ~100 nanoseconds                       │
│  1,000,000 elements: ~100 nanoseconds                       │
│                                                             │
│  WHY SO FAST?                                               │
│  Sets use hash tables internally - direct address           │
│  computation means no searching or iteration needed         │
│                                                             │
│  COMPARE TO ALTERNATIVES:                                   │
│  Array search:       O(n) - must check every element        │
│  Sorted array:       O(log n) - binary search               │
│  Set lookup:         O(1) - hash table magic!               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

This performance characteristic is why sets are the standard choice for whitelists, blacklists, and deduplication in high-throughput network monitoring.

## Set Operations: Union, Intersection, and Difference

Zeek provides mathematical set operations that allow you to combine, compare, and analyze sets using intuitive operator syntax borrowed from mathematics:

```c
# Define two sets with some overlapping elements
local set_a: set[string] = { "apple", "banana", "cherry" };
local set_b: set[string] = { "banana", "date", "elderberry" };

# UNION: Combine both sets (all unique elements from A and B)
local union_set = set_a | set_b;  
# Result: { "apple", "banana", "cherry", "date", "elderberry" }
# Use case: Merge two IP blacklists into one master list

# INTERSECTION: Find common elements (only items in both A and B)
local intersect_set = set_a & set_b;  
# Result: { "banana" }
# Use case: Find IPs that appear in both "suspicious" and "verified threat" lists

# DIFFERENCE: Elements in A but not in B (A minus B)
local diff_set = set_a - set_b;  
# Result: { "apple", "cherry" }
# Use case: Remove whitelisted IPs from a detected threat list

# SYMMETRIC DIFFERENCE: Elements in either A or B but not both
local sym_diff = (set_a | set_b) - (set_a & set_b);
# Result: { "apple", "cherry", "date", "elderberry" }
# Use case: Find IPs that changed state between two monitoring periods
```



### Set Comparison Operators

Beyond basic operations, Zeek supports **subset and superset** testing for hierarchical relationships between sets:

```c
local small_set: set[string] = { "apple", "banana" };
local large_set: set[string] = { "apple", "banana", "cherry", "date" };

# Subset check: Is small_set contained within large_set?
if ( small_set <= large_set )
    print "small_set is a subset of large_set";  # TRUE

# Proper subset: subset but not equal
if ( small_set < large_set )
    print "small_set is a proper subset";  # TRUE

# Superset check: Does large_set contain all of small_set?
if ( large_set >= small_set )
    print "large_set is a superset of small_set";  # TRUE

# Equality check
if ( set_a == set_b )
    print "Sets contain identical elements";
```

**Practical security monitoring example:**

```c
# Define expected internal servers
global expected_dns_servers: set[addr] = {
    192.168.1.10, 192.168.1.11
};

# Track actual DNS servers observed
global observed_dns_servers: set[addr];

event dns_request(c: connection, msg: dns_msg, query: string, 
                  qtype: count, qclass: count)
{
    add observed_dns_servers[c$id$resp_h];
}

# Periodically check for rogue DNS servers
event zeek_done()
{
    # Find unauthorized DNS servers (in observed but not expected)
    local rogue_servers = observed_dns_servers - expected_dns_servers;
    
    if ( |rogue_servers| > 0 )
    {
        print "WARNING: Rogue DNS servers detected:";
        for ( server in rogue_servers )
            print fmt("  Unauthorized DNS: %s", server);
    }
}
```


## Set Iteration: Traversing Unique Elements

While sets are unordered, you can iterate over their elements using the **`for-in`** loop syntax. This is useful for bulk processing, reporting, or applying operations to each unique element:

```c
# Iterate over all seen IPs
for ( ip in seen_ips )
{
    print fmt("Previously seen IP: %s", ip);
    
    # Could perform lookups, generate reports, etc.
    # Note: Modifying the set during iteration is undefined behavior!
}

# Count elements matching a criterion
local private_count = 0;
for ( ip in seen_ips )
{
    if ( is_private_addr(ip) )
        ++private_count;
}
print fmt("Found %d private IPs out of %d total", 
          private_count, |seen_ips|);
```

**Critical iteration warnings:**

```
⚠️  ITERATION ORDER IS NOT GUARANTEED
    - Sets are hash-based and unordered
    - Iteration order may vary between runs
    - Never rely on any specific ordering
    
⚠️  DO NOT MODIFY SET DURING ITERATION
    - Adding/removing elements while iterating = undefined behavior
    - May cause elements to be skipped or processed twice
    - Build a separate list of changes, apply after iteration
```



**Safe pattern for conditional removal:**

```c
# WRONG: Modifying during iteration
for ( ip in seen_ips )
    if ( should_remove(ip) )
        delete seen_ips[ip];  # DANGEROUS!

# CORRECT: Collect changes, apply after
local to_remove: set[addr];
for ( ip in seen_ips )
    if ( should_remove(ip) )
        add to_remove[ip];

# Now safe to remove
for ( ip in to_remove )
    delete seen_ips[ip];
```

## Real-World Example: Domain Generation Algorithm (DGA) Detection

One of the most powerful applications of sets in security monitoring is detecting **Domain Generation Algorithms (DGA)** - malware that generates large numbers of pseudo-random domain names to evade blacklists. By tracking the **diversity of domains** queried by each host, we can identify this suspicious behavior:

```c
# Track unique domains queried by each source IP
# Expires entries after 1 hour of inactivity
global domains_by_ip: table[addr] of set[string]
    &create_expire = 1hr;

event dns_request(c: connection, msg: dns_msg, query: string, 
                  qtype: count, qclass: count)
{
    local src = c$id$orig_h;
    
    # Initialize empty set for this IP if first DNS query
    if ( src !in domains_by_ip )
        domains_by_ip[src] = set();
    
    # Add domain to this IP's unique domain set
    # Set automatically handles duplicates - no conditional needed!
    add domains_by_ip[src][query];
    
    # Check for suspiciously high domain diversity
    # Legitimate clients query 5-20 domains typically
    # DGA malware queries hundreds to thousands
    if ( |domains_by_ip[src]| > 100 )
    {
        print fmt("🚨 DGA ALERT: %s queried %d unique domains in 1 hour",
                  src, |domains_by_ip[src]|);
        print "Possible malware using domain generation algorithm";
        
        # Could trigger additional analysis:
        # - Log all queried domains for pattern analysis
        # - Check if domains follow DGA patterns (random strings)
        # - Correlate with other malware indicators
    }
}
```

**Why sets are perfect for this use case:**

| Requirement              | How Sets Provide It                                |
| ------------------------ | -------------------------------------------------- |
| **Count unique domains** | Set size                                           |
| **Handle duplicates**    | Set inherently deduplicates; no manual checking    |
| **Fast insertion**       | O(1) add operation doesn't slow down as list grows |
| **Memory efficient**     | Only stores domain strings once, not per query     |
| **Simple logic**         | No complex conditional checking needed             |



**Detection threshold tuning:**

```
┌──────────────────────────────────────────────────────────────┐
│              DGA DETECTION THRESHOLD ANALYSIS                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Unique Domains/Hour     Typical Behavior                    │
│  ────────────────────────────────────────────────────────────│
│   1-10                   Normal user (email, web browsing)   │
│  10-30                   Power user (many sites/apps)        │
│  30-50                   Workstation with many services      │
│  50-100                  Border case - investigate           │
│  100-500                 ⚠️  HIGH CONFIDENCE DGA              │
│  500+                    ⚠️  VERY HIGH CONFIDENCE DGA         │
│                                                              │
│  FALSE POSITIVE SOURCES:                                     │
│  • CDNs with many subdomains                                 │
│  • Aggressive software updaters                              │
│  • Load balancers with DNS round-robin                       │
│  • Development/QA environments                               │
│                                                              │
│  TUNING RECOMMENDATION:                                      │
│  Start with threshold of 100, adjust based on your           │
│  environment's baseline normal behavior                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## Sets with Automatic Expiration

Zeek's **attribute system** allows sets to automatically remove stale entries, preventing unbounded memory growth in long-running monitoring. The **`&create_expire`** attribute specifies how long entries remain in the set without being re-added:

```c
# Set with 5-minute expiration window
global recent_connections: set[addr]
    &create_expire = 5min;  # IPs expire after 5 minutes of inactivity

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    
    # Check if we've seen this IP recently (within 5 minutes)
    if ( src in recent_connections )
    {
        print fmt("Repeat connection from %s (seen in last 5min)", src);
        # Could track connection frequency, detect scanning, etc.
    }
    else
    {
        print fmt("First connection from %s in 5+ minutes", src);
    }
    
    # Add/refresh the IP in the set
    # If already present, this resets its 5-minute timer
    add recent_connections[src];
}
```

**How expiration works:**

```
Timeline of IP 192.168.1.100 in recent_connections set:

T=0:00  → add recent_connections[192.168.1.100]
          Entry created, expires at T=5:00

T=2:30  → add recent_connections[192.168.1.100]  
          Entry refreshed, now expires at T=7:30

T=7:00  → if ( 192.168.1.100 in recent_connections )
          Returns TRUE (still present)

T=7:31  → if ( 192.168.1.100 in recent_connections )
          Returns FALSE (expired at 7:30, automatically removed)
```

**Expiration attributes available:**

|Attribute|Behavior|Use Case|
|---|---|---|
|**`&create_expire`**|Expire after time since creation/last add|Sliding window tracking|
|**`&read_expire`**|Expire after time since last read|Activity-based retention|
|**`&expire_func`**|Custom function decides when to expire|Complex expiration logic|

**Preventing memory exhaustion example:**

```c
# Without expiration - DANGEROUS!
global all_ips_ever: set[addr];  # Grows forever, will eventually exhaust memory

# With expiration - SAFE for long-term monitoring
global recent_ips: set[addr] &create_expire = 24hr;  # Auto-cleanup after 1 day

event new_connection(c: connection)
{
    add recent_ips[c$id$orig_h];
    # Set will never exceed ~24 hours worth of unique IPs
    # Memory usage bounded by network activity rate
}
```

## Sets vs Tables: Choosing the Right Collection Type

A common point of confusion for new Zeek developers is deciding between `set[T]` and `table[T] of any`. Understanding when to use each is crucial for writing efficient, maintainable scripts:

```
┌──────────────────────────────────────────────────────────────┐
│              SETS VS TABLES: DECISION FRAMEWORK              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ USE SET WHEN:                                            │
│  ────────────────────────────────────────────────────────────│
│  • You only care about membership (is X in the collection?)  │
│  • You don't need to associate values with keys              │
│  • You're building whitelists/blacklists                     │
│  • You're deduplicating data                                 │
│  • Memory efficiency is critical                             │
│                                                              │
│  Example: global malicious_ips: set[addr];                   │
│           if ( src_ip in malicious_ips )                     │
│               # Block this connection                        │
│                                                              │
│  ✅ USE TABLE WHEN:                                          │
│  ────────────────────────────────────────────────────────────│
│  • You need to map keys to associated values                 │
│  • You're counting, tracking, or storing per-item data       │
│  • You need to store metadata about each element             │
│  • You're building state machines or complex tracking        │
│                                                              │
│  Example: global conn_count: table[addr] of count;           │
│           ++conn_count[src_ip];  # Track counts per IP       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


### Memory and Performance Trade-offs

```
┌──────────────────────────────────────────────────────────────┐
│           MEMORY OVERHEAD: SETS VS TABLES                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  set[addr] storing 1,000,000 IPs:                            │
│    ~16 bytes per IP (just the address + hash overhead)       │
│    Total: ~16 MB                                             │
│                                                              │
│  table[addr] of count storing 1,000,000 IPs:                 │
│    ~24 bytes per IP (address + count + hash overhead)        │
│    Total: ~24 MB (50% more memory)                           │
│                                                              │
│  table[addr] of record (with 5 fields):                      │
│    ~64+ bytes per IP                                         │
│    Total: ~64+ MB (4x more memory)                           │
│                                                              │
│  RECOMMENDATION:                                             │
│  If you don't need the associated data, use sets to          │
│  conserve memory. For large-scale monitoring with            │
│  millions of tracked items, this difference matters.         │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


## Advanced Patterns and Best Practices

### Pattern 1: Nested Sets in Tables

For complex tracking scenarios, you can combine tables and sets - using a table where each value is itself a set:

```c
# Track which ports each IP has connected to
global ip_to_ports: table[addr] of set[port] &create_expire = 1hr;

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    local dst_port = c$id$resp_p;
    
    # Initialize set if this is the first connection from this IP
    if ( src !in ip_to_ports )
        ip_to_ports[src] = set();
    
    # Add port to this IP's port set
    add ip_to_ports[src][dst_port];
    
    # Detect port scanning (connecting to many different ports)
    if ( |ip_to_ports[src]| > 20 )
    {
        print fmt("🚨 PORT SCAN: %s connected to %d different ports",
                  src, |ip_to_ports[src]|);
    }
}
```

### Pattern 2: Set-Based Rate Limiting

Use sets with expiration to implement sliding window rate limiting:

```c
# Allow max 10 login attempts per IP per 5 minutes
global recent_login_attempts: table[addr] of set[time]
    &create_expire = 5min;

event login_attempt(ip: addr, username: string)
{
    if ( ip !in recent_login_attempts )
        recent_login_attempts[ip] = set();
    
    # Add timestamp of this attempt
    add recent_login_attempts[ip][network_time()];
    
    # Check if exceeded rate limit
    if ( |recent_login_attempts[ip]| > 10 )
    {
        print fmt("⚠️  RATE LIMIT: %s exceeded 10 logins in 5min", ip);
        # Could trigger automatic blocking, alert, etc.
    }
}
```

### Pattern 3: Set Comparison for Change Detection

Track state changes by comparing current and previous sets:

```c
global active_services: set[port];
global previous_active_services: set[port];

event check_service_changes()
{
    # Find newly started services
    local new_services = active_services - previous_active_services;
    if ( |new_services| > 0 )
    {
        for ( svc in new_services )
            print fmt("✅ New service started on port %s", svc);
    }
    
    # Find stopped services
    local stopped_services = previous_active_services - active_services;
    if ( |stopped_services| > 0 )
    {
        for ( svc in stopped_services )
            print fmt("❌ Service stopped on port %s", svc);
    }
    
    # Update for next comparison
    previous_active_services = active_services;
}
```

### Best Practices Summary

**✅ DO:**

- Use sets for membership testing and deduplication
- Apply `&create_expire` to prevent unbounded growth
- Leverage set operations (union, intersection, difference) for analysis
- Choose sets over tables when you don't need associated values

**❌ DON'T:**

- Rely on iteration order (sets are unordered)
- Modify sets during iteration (undefined behavior)
- Use sets when you need to count or store metadata (use tables instead)
- Forget to initialize sets in tables before adding elements



### Conclusion

Sets are one of Zeek's most powerful and efficient data structures for network security monitoring. Their combination of uniqueness guarantees, O(1) membership testing, automatic deduplication, and support for mathematical set operations makes them indispensable for tracking state, detecting anomalies, and implementing efficient whitelists and blacklists. By understanding when to use sets versus tables and leveraging expiration attributes for memory management, you can build robust, scalable network monitoring solutions that handle millions of events with minimal overhead.




### Knowledge Check: set Type

**Q1: Why does Zeek provide a dedicated `set` type when you could just use `table[key_type] of bool` to track membership?**

A: While `table[key_type] of bool` can technically track presence/absence, `set` is **semantically correct, more efficient, and self-documenting**. Sets use less memory (no value storage needed, just keys), make intent clearer (membership testing vs. key-value mapping), prevent logical errors (can't accidentally store meaningful data in the "value" that gets ignored), and provide mathematical set operations (union, intersection, difference) that would require manual implementation with tables. The type system should express your intent - if you only care about "is X present?", use a set, not a table with meaningless boolean values.



**Q2: What makes set membership testing O(1) constant time, and why does this matter for high-volume network monitoring?**

A: Sets use **hash tables** internally, which compute a hash of the element and directly index to its storage location - no searching or iteration required. Checking if an IP is in a 10-element set takes the same time as checking a 10-million-element set. In network monitoring processing millions of packets per second, the difference between O(1) and O(n) lookup can be thousands of times - turning a microsecond operation into milliseconds means dropping packets. This is why sets are the standard choice for whitelists, blacklists, and any "have I seen this?" check in production scripts.



**Q3: How do set operations (union, intersection, difference) enable security analysis patterns that would be tedious with manual iteration?**

A: Set operations provide **declarative, one-line solutions** to common security questions: "Which IPs are in both threat feeds?" (intersection), "Combine all known malicious IPs" (union), "Which detected IPs aren't whitelisted?" (difference). Without set operations, you'd write nested loops iterating through collections, manually checking membership, and building new collections - dozens of lines of error-prone code for each analysis. Set mathematics is both **more concise and more correct** because it expresses intent directly rather than through implementation details, and it's optimized by Zeek's runtime instead of running interpreted script loops.


**Q4: When should you choose a set over a table, and when does a table become necessary?**

A: Choose a **set** when you only need to know "Is X present?" and don't need to associate any additional data with X - use cases include deduplication, whitelists/blacklists, uniqueness tracking (e.g., "unique domains queried"), and existence checks. Choose a **table** when you need to map keys to values - counters (how many times did we see X?), timestamps (when did we first/last see X?), or any per-entity metadata. If you find yourself wanting to know *how many* or *when* or *what attributes* for each element, you need a table. If you only care about the binary "seen/not seen" distinction, use a set.


**Q5: Why can't you rely on iteration order when processing set elements, and when does this become a problem?**

A: Sets are **unordered by design** - their internal hash table structure organizes elements by hash value, not insertion order. Iteration order is implementation-dependent and can even vary between Zeek versions or runs. This becomes a problem when you need **temporal relationships** (processing events in the order they occurred), **sequence detection** (looking for specific orderings), or **deterministic output** (debugging or testing where order variation causes confusion). If order matters, use a **vector** instead of a set. Never write security logic that depends on set iteration order - it's non-deterministic and will break in subtle ways.

---





---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./table.md" >}})
[|NEXT|]({{< ref "./record.md" >}})

