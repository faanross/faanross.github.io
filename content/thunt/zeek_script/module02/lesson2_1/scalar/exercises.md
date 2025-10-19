---
showTableOfContents: true
title: "Part 2K - Scalar Types Practical Exercises"
type: "page"
---

## Exercise 1: Port Scan Detection

### Objective

Detect when a single host connects to many different ports on another host within a short time window - the signature of a port scan.

### The Attack Scenario

Port scanning is reconnaissance - an attacker probing which services are running on a target. Tools like [Nmap](https://nmap.org) systematically try connections to many ports to map the attack surface. This is almost always the first phase of an attack.

### The Detection Logic

We'll track how many unique destination ports each source IP contacts on each destination IP. When this count exceeds a threshold (let's say 20 ports), we alert.

**Key types used:**

- `count`: Track number of ports contacted
- `addr`: Identify source and destination IPs

### Write the Script

Create the detection script:

```bash
sudo nano /opt/zeek/share/zeek/site/custom/scan-detection.zeek
```

```c
# scan-detection.zeek
# Detects port scanning by tracking unique ports contacted per src->dst pair
# Logs only when scanning threshold is reached

@load base/frameworks/notice

module Scanning;

export {
    redef enum Notice::Type += {
        ## Indicates a host is port scanning another host
        Port_Scan_Detected
    };
    
    # How many unique ports before we consider it scanning
    const scan_threshold: count = 20;
    
    # Define the structure of our custom log
    type Info: record {
        ts: time &log;              # Timestamp when threshold was reached
        src: addr &log;             # Source IP (scanner)
        dst: addr &log;             # Destination IP (target)
        port_count: count &log;     # Number of unique ports contacted
    };
    
    # Create the logging stream
    redef enum Log::ID += { LOG };
}

# Track: src_ip -> dst_ip -> set of ports contacted
global port_tracker: table[addr, addr] of set[port] &create_expire=5min;

event zeek_init()
{
    print "=== SCAN DETECTION SCRIPT LOADED ===";
    print fmt("Threshold: %d ports", scan_threshold);
    print "Tracking window: 5 minutes of inactivity";
    
    # Initialize the custom log file
    Log::create_stream(Scanning::LOG, [$columns=Info, $path="port_scan"]);
    
    print "Custom log 'port_scan.log' initialized";
    print "Watching for port scans...";
}

event new_connection(c: connection)
{
    local src = c$id$orig_h;
    local dst = c$id$resp_h;
    local dst_port = c$id$resp_p;
    
    # Initialize the set if this is first connection from src to dst
    if ( [src, dst] !in port_tracker )
        port_tracker[src, dst] = set();
    
    # Add this destination port to the set
    add port_tracker[src, dst][dst_port];
    
    # Count how many unique ports we've seen
    local port_count = |port_tracker[src, dst]|;
    
    # Debug output (comment out for production)
    print fmt("Connection attempt: %s -> %s:%s (unique ports: %d)", 
              src, dst, dst_port, port_count);
    
    # Check threshold
    if ( port_count == scan_threshold )
    {
        print fmt("*** SCAN DETECTED! %s hit %d ports on %s ***", 
                  src, port_count, dst);
        
        # Send notice (appears in notice.log)
        NOTICE([$note=Port_Scan_Detected,
                $src=src,
                $msg=fmt("%s scanned %d ports on %s", src, port_count, dst)]);
        
        # Log to our custom port_scan.log (only at threshold)
        Log::write(Scanning::LOG, [
            $ts=network_time(),
            $src=src,
            $dst=dst,
            $port_count=port_count
        ]);
        
        print "Alert written to notice.log and port_scan.log";
    }
}
```

### Detailed Code Breakdown


```c
@load base/frameworks/notice
```

- Loads Zeek's Notice framework, which provides the alerting infrastructure for generating security notices that appear in `notice.log`.
- Note in this script that we will also provide detailed logging to a custom log called `port_scan.log`
- These serve different purposes, our custom log is for detailed telemetry capture, `notice.log` sends a notice, and can be integrated with for example our SIEM, e-mail based alerts etc.
- It's obviously not a requirement to use both, but I'm showing both here for educational/reference purposes.


```c
module Scanning;
```

Creates the `Scanning` namespace to encapsulate all declarations and avoid naming conflicts with other Zeek scripts.

```c
export {
```

- Begins the export block.
- Everything declared here becomes part of the module's public interface, accessible to other scripts and configurable by users.

```c
    redef enum Notice::Type += {
        ## Indicates a host is port scanning another host
        Port_Scan_Detected
    };
```

- Extends Zeek's built-in `Notice::Type` enumeration by adding `Port_Scan_Detected` as a new alert type.
- The `redef`keyword modifies an existing type definition, and `+=` appends to the enum.
- The `##` comment becomes documentation that appears in Zeek's automatically-generated documentation.

```c
    const scan_threshold: count = 20;
```

- Declares a configurable constant that defines the detection threshold.
- Once a source contacts 20 UNIQUE ports on a destination, it's classified as scanning and written to log.
- Temporal constraint is imposed as a sliding window (later on in script), whenever a scan is detected, there is a 5 minute grace period before the count is reset. Each new scan will reset the count (hence "sliding"). **More details on this below.**
- The `count` type ensures non-negative integers. Being in the export block allows administrators to override this value in local configuration files.


```c
    type Info: record {
```

- Defines a record type named `Info` that serves as the schema for log entries.
- This structure determines what columns appear in `port_scan.log`.

```c
        ts: time &log;              
```

- The `ts` field stores the timestamp when the scanning threshold was reached.
- This is not perfect - we might want to know for example when first scan was encountered. There are different ways of configuring it with their owns pros and cons, for now we are using this relatively simpler organization for introductory purposes.
- The `time` type represents epoch seconds with microsecond precision. The `&log` attribute marks this field for inclusion in log output.

```c
        src: addr &log;            
```

- The `src` field stores the source IP address (the scanning host).
- The `addr` type handles both IPv4 and IPv6 addresses.

```c
        dst: addr &log;           
```

- The `dst` field stores the destination IP address (the scan target).

```c
        port_count: count &log;    
```

- The `port_count` field records how many unique ports were contacted.
- In this implementation, this will always be 20 (the threshold value) when logged, but the field provides clarity and allows for future threshold adjustments.



**Mental model of the log schema:**

```
+-------------------+-------------+-------------+------------+
| ts                | src         | dst         | port_count |
+-------------------+-------------+-------------+------------+
| 1634567910.123456 | 192.168.1.5 | 10.0.0.100  | 20         |
+-------------------+-------------+-------------+------------+
```



```c
    redef enum Log::ID += { LOG };
```

- Extends Zeek's `Log::ID` enumeration to register a new log stream identifier.
- This creates `Scanning::LOG` as a unique identifier that Zeek's logging framework uses to track this specific log file.
- Every log type in Zeek (`conn.log`, `dns.log`, `http.log`, etc.) has a corresponding `Log::ID` entry.


**THIS PART IS THE KEY TO THE LOGIC OF OUR DETECTION:**
```c
global port_tracker: table[addr, addr] of set[port] &create_expire=5min;
```

This is the core data structure that maintains scanning state:

- `global` = module-level scope, accessible throughout the script
- `port_tracker` = variable name
- `table[addr, addr]` = a hash table with composite keys of two IP addresses
    - First `addr` = source IP address
    - Second `addr` = destination IP address
- `of set[port]` = values are sets containing port numbers
    - `set` = unordered collection that **automatically handles duplicates**
    - `port` = Zeek's type for TCP/UDP port numbers (0-65535)
- `&create_expire=5min` = automatic expiration based on inactivity
    - Entries are deleted after **5 minutes of no new connections** from that source to that destination
    - The timer **resets every time a new connection is made** (sliding window)
    - Tracking can persist indefinitely as long as connections occur within 5-minute intervals
    - If the gap between connections exceeds 5 minutes, all tracking data is lost and the count resets to zero
    - **Critical implication:** Slow scans (e.g., one port every 10 minutes) will never trigger detection because tracking expires between connections
    - Prevents unbounded memory growth by cleaning up stale entries

**Mental model:**

```json
port_tracker = {
  (192.168.1.5, 10.0.0.100): {22, 80, 443, 3306, 5432, 8080, ...},
  (192.168.1.5, 10.0.0.101): {22, 23, 445},
  (10.0.0.50, 192.168.1.100): {3389, 5900, 5901}
}
```

- Each source→destination pair is tracked independently.


```c
event zeek_init()
```

As we saw in the previous section, the `zeek_init()` event fires exactly once when Zeek starts, before any network traffic is processed. This is where one-time setup tasks occur.

```c
    Log::create_stream(Scanning::LOG, [$columns=Info, $path="port_scan"]);
```

Registers the custom log stream with Zeek's logging framework:

- `Log::create_stream()` = Zeek's function to create a new log
- `Scanning::LOG` = the log identifier we registered earlier
- `[$columns=Info, $path="port_scan"]` = configuration record:
    - `$columns=Info` = use the `Info` record structure for column definitions
    - `$path="port_scan"` = output filename (becomes `port_scan.log`)

This creates `/usr/local/zeek/logs/current/port_scan.log` with columns matching the `Info` record structure.


```c
event new_connection(c: connection)
```

- This event handler triggers whenever a TCP connection successfully completes the 3-way handshake. The `c` parameter is a record containing all connection metadata (5-tuple, timing, state, etc.).
- **Important:** `new_connection` fires when Zeek first observes a connection - triggered by the initial SYN packet or first packet of any connection type. This catches port scans regardless of whether they complete the 3-way handshake, making it more comprehensive than `connection_established` (which only fires after a full handshake) and more reliable than `connection_attempt` (which may not fire consistently in all environments).
- This only fires for established connections, not failed connection attempts. SYN scans that don't complete handshakes won't trigger this event.
- I also just want to point out that we ahve now seen 3 types of events - when the script starts, ends, when a connection is started. Zeek has many built-in event types we can use, and of course we can even create our own events. There will be an entire module dedicated to this since of course it's a foundational aspect of Zeek scripting. That being the case won't add any details here, I just wanted you to be aware of this for now.



```c
    local src = c$id$orig_h;
    local dst = c$id$resp_h;
    local dst_port = c$id$resp_p;
```

Extracts the relevant connection information:

- `local` = function-scoped variables that exist only within this event handler
- `c$id` = the connection identifier structure containing the 5-tuple (source + dest IP and port, and the protocol)
- `c$id$orig_h` = originator host (source IP address)
- `c$id$resp_h` = responder host (destination IP address)
- `c$id$resp_p` = responder port (destination port number)

**Note**: Zeek uses "originator" and "responder" terminology rather than "source" and "destination" to be protocol-agnostic and more precise about connection directionality.


```c
    if ( [src, dst] !in port_tracker )
        port_tracker[src, dst] = set();
```

**"Lazy initialization" pattern:**
- `[src, dst]` = composite key (tuple of two IP addresses)
- `!in` = "not in" operator checking if the key exists in the table
- If this source has never contacted this destination before, create a new empty set

This approach only allocates memory when needed, rather than pre-creating entries for every possible IP pair.


```c
    add port_tracker[src, dst][dst_port];
```

Adds the contacted port to the tracking set:

- `add` = Zeek's built-in function for adding elements to sets
- `port_tracker[src, dst]` = retrieves the set for this source→destination pair
- `[dst_port]` = the port to add

Sets automatically handle duplicates. If port 80 is contacted multiple times, the set still contains only one instance of port 80. This ensures we count unique ports, not total connections.

**

```c
    # Count how many unique ports we've seen
    local port_count = |port_tracker[src, dst]|;
```

Determines how many unique ports have been contacted:

- `| |` = cardinality operator (returns the size of a set)
- `port_tracker[src, dst]` = the set of ports for this source→destination pair
- Result stored in `port_count`

Example: if the set is `{22, 80, 443, 3306, 5432}`, then `port_count = 5`.


```c
    if ( port_count == scan_threshold )
```


**Checks if the unique port count has reached the threshold:**
- Uses `==` (equality) rather than `>=` to trigger exactly once
- When `port_count` equals 20 (the default threshold), enter this block
- On subsequent connections (21st, 22nd port, etc.), this condition is false

This ensures each source→destination scan pair generates exactly one alert and one log entry.

```c
        NOTICE([$note=Port_Scan_Detected,
                $src=src,
                $msg=fmt("%s scanned %d ports on %s", src, port_count, dst)]);
```

Generates a high-priority security notice:

- `NOTICE()` = Zeek's alerting function
- `$note=Port_Scan_Detected` = alert type (from our export block)
- `$src=src` = identifies the source IP as the alert subject
- `$msg=...` = human-readable message using string formatting
- `fmt()` = Zeek's printf-style formatting function
    - `%s` = string/address placeholders
    - `%d` = decimal integer placeholder

Example output: `"192.168.1.5 scanned 20 ports on 10.0.0.100"`

This notice appears in `notice.log` and can trigger additional actions (email alerts, SIEM integration, etc.) based on Zeek's Notice framework configuration.

```c
        # Log to our custom port_scan.log (only at threshold)
        Log::write(Scanning::LOG, [
            $ts=network_time(),
            $src=src,
            $dst=dst,
            $port_count=port_count
        ]);
```

Writes a single entry to the custom log:

- `Log::write()` = Zeek's function for writing log entries
- `Scanning::LOG` = which log stream to write to
- `[...]` = record containing field values (must match `Info` structure)
- `$ts=network_time()` = current packet timestamp (not system time)
- `$src=src`, `$dst=dst`, `$port_count=port_count` = populate remaining fields

This creates one line in `port_scan.log` when the threshold is reached.


### **Expiration-Based Logging (Alternative Approach)**

Before we move on to using our script I just briefly want to mention another approach we could have used instead of threshold-based logging.

An alternative approach logs a summary when a tracking window expires. Let's say for example we count the total scans in a 60-minute windows, and when that time expires, it logs the total count of everything.

Note: 60 mins is of course completely adjustable, it can be increased/decreased, pushing it in either direction has pros and cons. Every decision in security is about optimizing trade-offs in numerous dimensions.


### Verify and Load our Script

**Verify our syntax:**
```bash
zeek -a /opt/zeek/share/zeek/site/custom/scan-detection.zeek
```

No news is good news.


**Load it in `local.zeek`:**
```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```

**Add at the bottom:**

```zeek
@load ./custom/scan-detection.zeek
```



### Let's Now Run a Single Instance of Zeek

Please note that when you run a single instance of zeek (i.e. using zeek instead of zeekctl), it will write the logs to the directory you are running from NOT `/opt/zeek/logs/current/`.

So just create a new directory for our test to run from

```bash
mkdir test_scan
cd ./test_scan

# assuming your interface is eth0, if not run `ip a` to find out what it is
sudo zeek -C -i eth0 local.zeek
```


You should immediately see our debug output:
```bash
zeek@zeek-sensor:~/test_scan$ sudo zeek -C -i eth0 local.zeek
listening on eth0

=== SCAN DETECTION SCRIPT LOADED ===
Threshold: 20 ports
Tracking window: 5 minutes of inactivity
Custom log 'port_scan.log' initialized
Watching for port scans...
```

If you don't see it, it probably means you forgot to load it in `local.zeek`

I'll be running the scan from another machine, if you want to do the same run `ip a` to get your public IP.



### Generate the Attack

We'll use **Nmap** to perform a real port scan:

```bash
# Install nmap if needed
sudo apt install nmap
```

Let's scan ports 1 to 100:
```bash
sudo nmap -p 1-100 <public IP of target machine>
```



### Detection

You should see the scans appear in the terminal where Zeek is running and eventually the alert too:

```
*** SCAN DETECTED! 90.29.239.33 hit 20 ports on 136.97.124.83 ***
Alert written to notice.log and port_scan.log
```


And now we can also confirm it was written to both notice.log and port_scan.log. Just please keep in mind these files will be in `pwd`, NOT the daemon's log directories in `/opt`.


```bash
zeek@zeek-sensor:~/test_scan$ cat notice.log
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	notice
#open	2025-10-19-08-14-22
#fields	ts	uid	id.orig_h	id.orig_p	id.resp_h	id.resp_p	fuid	file_mime_type	file_desc	proto	note	msg	sub	src	dst	p	n	peer_descr	actions	email_dest	suppress_for	remote_location.country_code	remote_location.region	remote_location.city	remote_location.latitude	remote_location.longitude
#types	time	string	addr	port	addr	port	string	string	string	enum	enum	string	string	addr	addr	port	count	string	set[enum]	set[string]	interval	string	string	string	double	double
1760876062.226688	-	-	-	-	-	-	-	-	-	Scanning::Port_Scan_Detected	90.29.239.33 scanned 20 ports on 136.97.124.83	-	90.29.239.33-	-	-	-	Notice::ACTION_LOG	(empty)	3600.000000	-	-	-	-	-
```

```bash
zeek@zeek-sensor:~/test_scan$ cat port_scan.log
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	port_scan
#open	2025-10-19-08-14-22
#fields	ts	src	dst	port_count
#types	time	addr	addr	count
1760876062.226688	70.27.237.44	138.197.134.43	20
```



---







---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./first.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})

