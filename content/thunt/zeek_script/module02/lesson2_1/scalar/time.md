---
showTableOfContents: true
title: "The time Type: Timestamps"
type: "page"
---



## **The time Type: Timestamps**

The `time` type represents absolute points in time - specific moments on the timeline. If you think of time as a number line stretching from the past into the future, a `time` value is a single point on that line. Zeek uses the `time` type extensively because network security analysis is fundamentally about understanding **when** things happen: when a connection started, when a packet arrived, when an alert fired, when suspicious behavior began.

### **Why Timestamps Matter in Security**

Time is one of the most critical dimensions in security analysis. Attacks unfold over time. Patterns emerge when you look at sequences of events. A single connection might seem innocent, but fifty connections spaced exactly 60 seconds apart suggests beaconing - a hallmark of command-and-control traffic.

Consider what you can detect with accurate timestamps: **brute force attacks** (many attempts in a short window), **beaconing malware** (periodic connections with regular intervals), **data exfiltration** (sustained transfers over time), **reconnaissance** (rapid connections to many targets), **time-based evasion** (attacks timed to avoid monitoring periods), and **coordinated attacks** (simultaneous activity across multiple hosts).

Without precise time tracking, you're flying blind. The `time` type gives you the foundation to build these temporal detections.




### **Basic Usage: Getting Time Values**

In Zeek scripts, time values typically come from the network events you're analyzing, but you can also get the current time when needed:

```c
# Current time
# Current time from packet timestamps
local now: time = network_time(); 
# Actual current time 
local current: time = current_time();  

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

For almost all security detection logic, **you should use `network_time()`**. Here's why: When analyzing live traffic, `network_time()` gives you precise timing from the packets themselves, accounting for any processing delays. More importantly, when analyzing saved packet captures (PCAPs) offline, `network_time()` works correctly - it uses the timestamps from when the traffic was originally captured. If you used `current_time()` in offline analysis, all your timing logic would be wrong because you'd be comparing 2024 packet timestamps to 2025 processing timestamps.

Think of `network_time()` as "when did this happen on the network?" and `current_time()` as "what time is it right now in the real world?" For security analysis, you almost always care about the former.

### **Working with Time Values**

The `time` type supports several essential operations that let you build temporal logic:

**Time arithmetic** with intervals lets you calculate future or past moments:

```c
# Time arithmetic
local start: time = network_time();
local duration: interval = 5min;
local end: time = start + duration;  
# time + interval = time
```

Adding an interval (a duration) to a time produces a new time. This is useful for calculating expiration times, timeout windows, or future scheduled events. The type system enforces correctness - you can only add intervals to times, not arbitrary numbers.

**Time comparison** tells you the ordering of events:

```c
# Time comparison
if ( end > start )
    print "End is after start";  
```

You can check if one event happened before, after, or at the same time as another. This is fundamental for detecting sequences ("Did the login happen before the file access?") or temporal proximity ("Did these two events happen within seconds of each other?").

**Time differences** produce intervals:

```c
# Time difference (produces interval)
local elapsed: interval = end - start;  
# time - time = interval
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


## Knowledge Check: time Type

**Q1: What's the critical difference between network_time() and current_time(), and which should you use for security detection logic?**

A: `network_time()` returns the timestamp from the packet currently being processed (when the event happened on the network), while `current_time()` returns the actual wall clock time right now. You should almost always use `network_time()` for detection logic because it uses packet timestamps for accurate timing and works correctly when analyzing saved PCAPs offline. Using `current_time()` in offline analysis would compare old packet timestamps to current processing time, breaking all timing logic.

**Q2: How is the time type represented internally, and what precision does it provide?**

A: Time is represented as a double-precision floating-point number storing seconds since Unix epoch (January 1, 1970). For example, 1696348338.423 means 1,696,348,338 seconds and 423 milliseconds since the epoch. It provides microsecond precision (6 decimal places), which is more than sufficient for network security analysis where most events are separated by milliseconds or more.

**Q3: What are the three main operations you can perform with time values, and what type does each operation produce?**

A: (1) Time + interval = time (calculating future or past moments), (2) Time - time = interval (measuring duration between events), (3) Time comparison (>, <, etc.) = bool (determining order of events). Note you cannot add two times together (that wouldn't make semantic sense), and subtracting times produces an interval, not another time.

**Q4: Why is temporal analysis (using time values) so important for security detection? Name three attack characteristics you can only detect with accurate timing.**

A: Time is critical because attacks unfold over time and patterns emerge in sequences of events. Three examples: (1) Brute force attacks - many attempts in a short window, (2) Beaconing malware - periodic connections with regular intervals indicating C2 communication, (3) Coordinated attacks - simultaneous activity across multiple hosts suggesting orchestrated behavior. Without precise timestamps, these patterns are invisible.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./port.md" >}})
[|NEXT|]({{< ref "./interval.md" >}})

