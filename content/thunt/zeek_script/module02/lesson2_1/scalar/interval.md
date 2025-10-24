---
showTableOfContents: true
title: "The interval Type: Time Durations"
type: "page"
---




## The interval Type: Time Durations

The `interval` type represents durations - lengths or spans of time. While the `time` type answers "when did this happen?" (a point on the timeline), the `interval` type answers "how long did this last?" (a distance along the timeline). Understanding this distinction is fundamental: a timestamp locates an event in time, while an interval measures elapsed time between events or the duration of an event.

### **Why Durations Matter in Security**

Security analysis constantly deals with questions of duration: How long did that connection last? How much time elapsed between failed login attempts? How long has this host been silent? What's the timeout window for this detection? How frequently should we check for suspicious patterns?

Durations are critical for distinguishing normal from abnormal. A connection lasting three seconds might be a normal web request. A connection lasting three hours could be a persistent backdoor. Failed logins spaced 10 minutes apart might be a forgetful user. Failed logins spaced 2 seconds apart is a brute force attack. Understanding duration and timing patterns is essential to accurate detection.

### **Basic Usage and Declaration**

Zeek makes working with intervals intuitive by supporting natural time unit notation:

```c
# Declaring intervals
local five_minutes: interval = 5min;
local one_hour: interval = 1hr;
local thirty_seconds: interval = 30sec;
local one_day: interval = 1day;

# Multiple units
local complex_interval: interval = 1hr + 30min + 15sec;

# From calculations
local duration: interval = end_time - start_time;
```

Notice how readable this is. You don't write `300` and hope someone remembers that's seconds - you write `5min` and the meaning is immediately clear. This readability extends to maintenance: when you revisit code months later, `30sec` is instantly understandable while `30000` requires you to figure out whether that's milliseconds, microseconds, or something else.

**Creating intervals through arithmetic** is equally common. When you subtract one `time` from another, you get an `interval` - the duration between those two moments.


### **Available Time Units**

Zeek provides a comprehensive set of time units covering the full range from microseconds to days:

```c
# Available time units (can be mixed)
# Microsecond
1usec   

# Millisecond 
1msec   

# Second
1sec    

# Minute
1min   

# Hour
1hr   

# Day  
1day    

# Examples
local short: interval = 100msec;
local medium: interval = 5min;
local long: interval = 24hr;
local very_long: interval = 7day;
```

You can **mix units freely** to express complex durations: `1hr + 30min + 15sec` is perfectly valid and represents exactly one hour, thirty minutes, and fifteen seconds. Zeek handles all the unit conversion automatically.

These units cover every practical timescale for network security:

- **Microseconds/milliseconds:** Network latency, packet timing, sub-second response analysis
- **Seconds/minutes:** Connection durations, brute force attack windows, event clustering
- **Hours/days:** Long-running connections, behavioral baselines, retention periods


### **Interval Arithmetic and Comparisons**

The `interval` type supports intuitive arithmetic operations that let you build sophisticated temporal logic:

**Addition and subtraction** combine or reduce durations:

```c
# Addition
local total: interval = 5min + 30sec;  
# 5 minutes 30 seconds

# Subtraction
local difference: interval = 1hr - 15min;  
# 45 minutes
```

**Multiplication and division** by scalars let you scale durations:

```c
# Multiplication by scalar
local triple: interval = 10sec * 3;  
# 30 seconds

# Division
local half: interval = 1hr / 2;  
# 30 minutes
```

These operations are useful for calculating timeouts ("wait twice as long as last time"), adjusting thresholds ("reduce the detection window by half"), or building schedules ("check every N minutes where N changes based on load").

**Comparison operations** let you implement threshold-based detection:

```c
# Comparison
if ( interval1 > 5min )
    print "More than 5 minutes";
    
if ( interval2 < 1sec )
    print "Sub-second duration";
```

You can check if a duration is longer than, shorter than, or equal to a threshold. This is the foundation of time-based alerting: "alert if connection lasts longer than X" or "alert if events happen faster than Y."




### **Real-World Example: Connection Duration Analysis**

Connection duration is one of the most informative characteristics for detecting malicious activity. Let's examine how to use intervals to build duration-based detections:

```c
# Detect suspiciously long connections (persistent backdoor)
event connection_state_remove(c: connection)
{
    # Only check if we have timing info
    if ( !c?$duration )
        return;
    
    local duration = c$duration;
    local src = c$id$orig_h;
    local dst = c$id$resp_h;
    
    # Connections lasting over 1 hour are unusual
    if ( duration > 1hr )
    {
        print fmt("Long connection: %s -> %s lasted %s",
                  src, dst, duration);
    }
    
    # Very short connections with data transfer (scanning?)
    if ( duration < 1sec && c$orig_bytes + c$resp_bytes > 0 )
    {
        print fmt("Fast connection: %s -> %s completed in %s",
                  src, dst, duration);
    }
}
```

**Understanding this detection:** We're examining connections as they close (the `connection_state_remove` event). First, we check if duration information is available - not all connections have complete timing data.

**Long connections** (over one hour) are statistically unusual. Most legitimate traffic consists of relatively short connections: web requests complete in seconds, email retrieval takes seconds to minutes, file transfers rarely exceed 30 minutes unless they're very large. A connection lasting hours could indicate a persistent shell, a backdoor maintaining a connection, or data exfiltration over a slow channel to avoid detection.

**Very short connections with data** (under one second but transferring bytes) might indicate scanning or automated probing. Normal connections usually involve some handshaking and data exchange that takes at least a second or two. When you see connections completing in milliseconds but still transferring data, it might be a scanner that connects, sends a probe, gets a response, and immediately disconnects - all in a fraction of a second.

### **Scheduling and Periodic Checks**

Intervals are also essential for scheduling recurring tasks - checking for patterns periodically:

```c
# Detect beaconing based on connection regularity
global connection_schedule: table[addr] of interval;

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    
    # Schedule check for regular connections
    if ( src !in connection_schedule )
    {
        # Check every minute
        connection_schedule[src] = 1min;  
        schedule connection_schedule[src] {
            check_for_beaconing(src)
        };
    }
}
```

**Understanding scheduled checks:** We're using the `schedule` statement to arrange future execution of code. The `connection_schedule[src]` interval (one minute) tells Zeek to wait that long before executing the scheduled code block. This pattern is common for periodic analysis: "every minute, check if this host is beaconing" or "every five minutes, summarize activity and look for anomalies."

Scheduling with intervals lets you build detections that aggregate data over time windows, check for patterns that only emerge across multiple events, or implement rate limiting and throttling of alerts.

### **Practical Patterns with Intervals**

Here are common patterns you'll use repeatedly in Zeek scripts:

**Timeout windows:** Define how long to wait before considering something "timed out"

```c
local timeout: interval = 5min;
if ( network_time() - last_seen_time > timeout )
    print "Connection timed out";
```

**Rate limiting:** Ensure events don't happen too frequently

```c
local min_interval: interval = 1sec;
if ( current_time - last_event_time < min_interval )
    return;  
# Too soon, ignore this event
```

**Time windows:** Aggregate or analyze data within sliding windows

```c
local window: interval = 10min;
# Count events in the last 10 minutes
```

**Threshold checking:** Alert when durations exceed or fall below limits

```c
if ( connection_duration > 30min )
    alert("Long-running connection detected");
```

### **Why This Matters for Security**

The `interval` type transforms how you think about time in security analysis. Instead of dealing with raw timestamp differences or converting everything to seconds manually, you work with durations as first-class values. This makes your code more readable, more maintainable, and less error-prone.

Consider the difference:

```c
# Without interval type (error-prone)
if ( (end_time - start_time) > 300.0 )  
# What unit is 300?
    
# With interval type (self-documenting)
if ( (end_time - start_time) > 5min )   
# Obviously 5 minutes
```

The second version is instantly clear to anyone reading the code. The first requires mental conversion and assumes you know the time is in seconds.

More importantly, intervals let you build temporal logic that mirrors how security analysts think:
- "Alert if the connection lasts longer than an hour," 
- "Check for patterns every five minutes," 
- "Ignore events that happen within one second of each other," 
- "Baseline activity over 24-hour windows." 


These natural expressions of time-based rules translate directly into clean, understandable Zeek code.

As you develop advanced detections, you'll find yourself using intervals constantly - for rate limiting to reduce alert fatigue, for defining behavioral baselines ("normal connections last between 30 seconds and 5 minutes"), for implementing adaptive thresholds that change based on observed patterns, and for scheduling periodic analysis tasks. The `interval` type isn't just a convenience - it's a fundamental building block of sophisticated temporal analysis.




## Knowledge Check: interval Type

**Q1: What's the fundamental difference between the time type and the interval type in terms of what they represent?**

A: The time type represents absolute points on the timeline (specific moments - "when did this happen?"), while the interval type represents durations or spans of time (lengths along the timeline - "how long did this last?"). A timestamp locates an event in time; an interval measures elapsed time between events or the duration of an event.

**Q2: What are all the available time units for intervals, and why is it better to write "5min" instead of "300"?**

A: Available units: usec (microsecond), msec (millisecond), sec (second), min (minute), hr (hour), day (day). Writing "5min" is better because it's immediately readable and self-documenting - anyone reading the code knows exactly what duration is meant. Writing "300" requires figuring out what unit is intended (seconds? milliseconds?) and mental conversion, making code harder to maintain and more error-prone.

**Q3: Explain what operations are valid with intervals and what type each produces: interval + interval, interval * 3, time + interval, time - time.**

A: interval + interval = interval (combining durations), interval * 3 = interval (scaling duration by a number), time + interval = time (calculating a future/past moment), time - time = interval (duration between two moments). Note that interval + time also equals time (commutative), but interval - time is not valid (can't subtract a point from a duration).

**Q4: Why is connection duration analysis (using intervals) valuable for security detection? Give examples of suspicious short and suspicious long durations.**

A: Duration reveals behavioral patterns. Suspiciously long connections (>1 hour) might indicate persistent shells, backdoors, or slow data exfiltration designed to evade detection. Suspiciously short connections with data transfer (<1 second) might indicate scanning or automated probing where the attacker connects, sends a probe, gets a response, and immediately disconnects. Normal connections typically have intermediate durations reflecting legitimate data exchange and protocol handshaking.

**Q5: How do intervals enable "scheduled" or "periodic" analysis in Zeek scripts?**

A: Intervals define time delays for the `schedule` statement, allowing you to arrange future code execution. For example, `schedule 5min { check_for_patterns() }` waits 5 minutes then executes the scheduled code. This enables periodic analysis like "every minute, check if this host is beaconing" or "every five minutes, aggregate and analyze activity," which is essential for detections that require looking at patterns over time windows rather than individual events.


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./time.md" >}})
[|NEXT|]({{< ref "./string.md" >}})

