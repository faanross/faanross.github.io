---
showTableOfContents: true
title: "Part 1 - The Single-Instance Processing Pipeline"
type: "page"
---

# **LESSON 1.2: ARCHITECTURE DEEP DIVE**


## **Welcome to the Inner Workings of Zeek**

In Lesson 1.1, you learned about Zeek's philosophical approach and its place in the network security ecosystem. You understand that Zeek performs behavioural analysis through an event-driven architecture, but we haven't yet explored what that really means at a technical level. In this lesson, we're going to open the hood and examine exactly how Zeek works internally.

This isn't just academic knowledge. Understanding Zeek's architecture will directly impact your success in deploying and using the system. When you're planning your deployment, architectural knowledge will help you determine how many sensors you need and where to place them. When you're writing detection scripts, understanding the processing pipeline will help you write efficient code that doesn't degrade performance. When something goes wrong - and at some point, something always goes wrong - architectural knowledge will help you troubleshoot effectively.

We're going to explore three major aspects of Zeek's architecture. First, we'll examine the single-instance processing pipeline: how packets flow from the network interface through protocol analysis to script execution. Second, we'll dive deep into cluster architecture, understanding how Zeek scales to handle high-bandwidth networks by distributing work across multiple machines. Finally, we'll discuss memory management and performance considerations that will inform both your deployment decisions and your scripting practices.



---

## **PART 1: THE SINGLE-INSTANCE PROCESSING PIPELINE**

### **Understanding How Zeek Processes Network Traffic**

Let's start by understanding how a single instance of Zeek processes network traffic. Even if you eventually deploy Zeek in a distributed cluster (and most production deployments do), understanding the single-instance architecture is crucial because every node in a cluster runs this same processing pipeline.

Here's a high-level view of the pipeline:


```
┌─────────────────────────────────────────────────────────────┐
│                    NETWORK INTERFACE                        │
│         (Packets arriving from monitored network)           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  PACKET ACQUISITION                         │
│   (Libpcap, AF_PACKET, PF_RING, or other capture method)    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   PACKET FILTER                             │
│        (BPF filter to exclude irrelevant traffic)           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  EVENT ENGINE                               │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Protocol Analysis & Parsing                  │   │
│  │  (TCP/IP stack, HTTP, DNS, SSL, FTP, SSH, etc.)      │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │          Connection State Management                 │   │
│  │    (Tracking ongoing connections and their state)    │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │            Event Generation                          │   │
│  │    (Creating events from protocol activities)        │   │
│  └──────────────────┬───────────────────────────────────┘   │
└─────────────────────┼───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    SCRIPT LAYER                             │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Event Handlers                             │   │
│  │    (Your detection scripts respond to events)        │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Logging Framework                          │   │
│  │      (Generate structured log files)                 │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                       │
│                     ▼                                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │          Notice Framework                            │   │
│  │    (Generate alerts for suspicious activity)         │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      OUTPUT                                 │
│       (Log files, notices, exported data)                   │
└─────────────────────────────────────────────────────────────┘
```




This diagram shows the complete path that network traffic takes through Zeek. Let's examine each stage in detail, understanding not just what happens but why it's designed this way and what implications it has for your deployment and scripting.

### **Stage 1: Packet Acquisition - Getting Traffic Into Zeek**

The first challenge Zeek must solve is getting packets from the network into the application for analysis. This might seem straightforward - after all, operating systems provide mechanisms for applications to capture network traffic - but packet acquisition is actually one of the most critical performance considerations in network monitoring. If Zeek can't acquire packets fast enough to keep up with your network's traffic volume, it will drop packets, resulting in incomplete visibility.

Zeek supports several different packet acquisition methods, each with different performance characteristics and use cases. Understanding these options will help you choose the right approach for your deployment.



#### **Libpcap: The Default and Most Portable Option**

The default packet acquisition mechanism is `libpcap` (or WinPcap on Windows, though Zeek is rarely deployed on Windows). Libpcap is a portable C library that provides a consistent API for packet capture across different operating systems. When you install Zeek without any special configuration, it uses `libpcap` to capture packets from your network interface.

Here's how libpcap works at a high level:

```
Network Interface → Kernel Network Stack → Packet Filter (BPF) → Userspace Buffer → Zeek
```

Packets arrive at your network interface and are processed by the kernel's network stack. A packet filter written in BPF (Berkeley Packet Filter) bytecode runs in kernel space, allowing you to discard irrelevant traffic before it's copied to userspace. The packets that pass the filter are copied to a buffer in userspace where Zeek can access them.

**Advantages of libpcap:**

- **Portability**: Works on Linux, BSD, macOS, and other Unix-like systems with minimal differences
- **Maturity**: Extremely well-tested and reliable
- **Simplicity**: Easy to set up with no special configuration required
- **Flexibility**: Works with any network interface your OS supports

**Limitations of libpcap:**

- **Performance ceiling**: On very high-bandwidth networks (10+ Gbps), libpcap may struggle to keep up
- **Single-threaded**: Libpcap delivers packets to a single processing thread, creating a bottleneck
- **Kernel copying overhead**: Every packet must be copied from kernel space to userspace

For many deployments, especially those monitoring networks below 1 Gbps, libpcap performs perfectly well. It's the right choice when you're starting out or when you don't need maximum performance.




#### **AF_PACKET: Linux's Native High-Performance Alternative**

On Linux systems, Zeek can use AF_PACKET instead of libpcap for significantly better performance. AF_PACKET is a Linux-specific interface that allows applications to capture packets with lower overhead than libpcap.

The key advantage of AF_PACKET is its use of memory-mapped buffers. Instead of copying each packet from kernel space to userspace, AF_PACKET sets up a shared memory region that both the kernel and Zeek can access. Packets are written directly into this shared memory, eliminating the copy overhead.

**AF_PACKET packet flow:**

```
Network Interface → Kernel NIC Driver → Memory-Mapped Ring Buffer → Zeek
                                         (Shared between kernel and userspace)
```

**Advantages of AF_PACKET:**

- **Better performance**: Lower CPU overhead and higher packet rates compared to libpcap
- **Fewer dropped packets**: More efficient handling reduces the risk of packet loss
- **Native to Linux**: No additional software required on modern Linux systems

**Considerations:**

- **Linux-only**: Not portable to other operating systems
- **Requires proper configuration**: You need to specify fanout modes and buffer sizes appropriately
- **More complex setup**: Not as simple as the default libpcap option

For production deployments on Linux, especially those monitoring networks above 1 Gbps, AF_PACKET is generally the better choice. We'll configure it when we install Zeek in Lesson 1.3.


#### **PF_RING: The High-Performance Commercial Option**

For the most demanding environments-networks running at 10 Gbps or higher-there's PF_RING, a high-performance packet capture framework developed by ntop. PF_RING provides even better performance than AF_PACKET through several optimizations.

**Key PF_RING features:**

- **Kernel bypass capabilities**: Can deliver packets directly from the network card to userspace, bypassing much of the kernel network stack
- **Hardware acceleration**: Supports offloading filtering and load balancing to compatible network cards
- **Better distribution**: More sophisticated algorithms for distributing packets across multiple processing threads

**The PF_RING trade-offs:**

```
┌─────────────────────────────────────────────────────────────┐
│                      PERFORMANCE                            │
│                                                             │
│  Libpcap     <     AF_PACKET     <     PF_RING              │
│  (Baseline)        (2-3x better)       (3-5x better)        │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                     COMPLEXITY                              │
│                                                             │
│  Libpcap     <     AF_PACKET     <     PF_RING              │
│  (Simple)          (Moderate)          (Complex)            │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                        COST                                 │
│                                                             │
│  Libpcap     <     AF_PACKET     <     PF_RING              │
│  (Free)            (Free)              (License required)   │
└─────────────────────────────────────────────────────────────┘
```

For most learning and even production deployments, PF_RING isn't necessary. We'll use AF_PACKET in this course, which provides excellent performance for networks up to several gigabits per second without the complexity or cost of PF_RING.


### **Stage 2: Packet Filtering - Reducing Unnecessary Load**

Before Zeek processes packets, they pass through a packet filter implemented using BPF (Berkeley Packet Filter). This filter runs in kernel space and can discard packets before they're delivered to Zeek, dramatically reducing the processing load.

You might wonder why you'd want to filter packets when the goal is comprehensive network visibility. The answer is that not all network traffic is relevant to security monitoring, and processing irrelevant traffic wastes resources that could be used for analyzing important traffic.

**Common filtering scenarios:**

|Traffic Type|Should Monitor?|Reasoning|
|---|---|---|
|Internet-bound traffic|Yes|Critical for detecting C2, data exfiltration|
|Internal server traffic|Yes|Important for lateral movement detection|
|Printer to file server|Probably not|High volume, low security relevance|
|Backup traffic|Probably not|High volume, legitimate file transfers|
|Network management (SNMP)|Maybe|Depends on your monitoring goals|

BPF filters are specified using a simple syntax. Here are some examples:

```
# Monitor only TCP traffic
tcp

# Monitor web traffic (HTTP and HTTPS)
port 80 or port 443

# Monitor all traffic except to/from backup server
not host 192.168.1.50

# Monitor only outbound traffic from internal network
src net 192.168.0.0/16 and dst net not 192.168.0.0/16

# Complex: Monitor HTTP(S) and DNS, but exclude internal DNS server
(port 80 or port 443 or port 53) and not (host 192.168.1.10 and port 53)
```

The key principle is to be thoughtful about what you filter out. Filter aggressively enough to reduce unnecessary load, but not so aggressively that you lose visibility into important security events. When in doubt, err on the side of capturing more rather than less-you can always adjust your filters based on experience.

In case you wanted to alter your BPF filter, this can typically be done directly **within Zeek's own configuration files**.

Think of it as giving Zeek a set of instructions on what traffic to even look at in the first place.

#### Where and How to Configure BPF Filters

The most common place to set your BPF filters is in the `local.zeek` file. This is the primary file for site-specific Zeek configurations. If you're running a Zeek cluster, you'll typically find this file in `<zeek_install_dir>/share/zeek/site/`.

You'll use the `restrict_filters` variable to define your BPF rules. Let's take a look at how you'd implement the examples from above.

```
# Redefine the restrict_filters to add our custom BPF rules.
redef restrict_filters += {
    # Monitor only TCP traffic
    ["only-tcp"] = "tcp",

    # Monitor web traffic (HTTP and HTTPS)
    ["web-traffic"] = "port 80 or port 443",

    # Monitor all traffic except to/from backup server
    ["exclude-backup-server"] = "not host 192.168.1.50",

    # Monitor only outbound traffic from internal network
    ["outbound-only"] = "src net 192.168.0.0/16 and not dst net 192.168.0.0/16",

    # Complex: Monitor HTTP(S) and DNS, but exclude internal DNS server
    ["web-and-dns-filtered"] = "(port 80 or port 443 or port 53) and not (host 192.168.1.10 and port 53)"
};
```


**Key things to note:**

- **`redef restrict_filters +=`**: This is important. The `+=` appends your filters to any existing ones. If you use a single `=`, you'll overwrite the default filters.
- **`["filter-name"]`**: This is just a descriptive name for your filter. It's good practice to give each filter a unique and meaningful name.
- **`"bpf-syntax"`**: This is where you place your actual BPF filter, just like in your course examples.

#### Another Method: Using `PacketFilter::exclude`

You can also use the `PacketFilter::exclude` function within the `zeek_init()` event in a seperate Zeek script. This can be useful for more dynamic or complex filtering scenarios.

Here's an example of how you might use this method in a custom script:


```
event zeek_init() {
    # Exclude traffic to and from the backup server
    PacketFilter::exclude("exclude-backup-server", "host 192.168.1.50");
}
```

This achieves the same goal as the `restrict_filters` example. For most common use cases, sticking with `restrict_filters` is perfectly fine and often easier to manage.

After you've made changes to your configuration, you'll need to restart Zeek for the new BPF filters to take effect. You can verify that your filters have been applied by using the command `zeekctl diag`.


### **Stage 3: The Event Engine - The Heart of Zeek**

Now we reach the most complex and interesting part of Zeek's architecture: the event engine. This is where Zeek transforms raw packets into the high-level events that your scripts respond to. The event engine is what makes Zeek fundamentally different from packet inspection tools.

The event engine has three main responsibilities: protocol analysis, connection state management, and event generation. Let's examine each.

#### **Protocol Analysis: Understanding What's Being Communicated**

When packets arrive at the event engine, Zeek doesn't just look at their raw bytes. Instead, it parses the protocols being used, extracting meaningful information from protocol headers and payloads. Zeek includes built-in analyzers for dozens of protocols, from low-level protocols like IP and TCP up through application protocols like HTTP, DNS, SSH, SSL, and many others.

**Zeek's protocol analysis stack:**

```
┌─────────────────────────────────────────────────────────────┐
│                   APPLICATION PROTOCOLS                     │
│   HTTP, DNS, SSL/TLS, SSH, FTP, SMTP, SMB, RDP, etc.        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ├─→ Extract: URLs, headers, status codes
                     ├─→ Extract: Query names, response IPs
                     ├─→ Extract: Certificates, cipher suites
                     └─→ Extract: Commands, file transfers
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  TRANSPORT PROTOCOLS                        │
│                    TCP, UDP, ICMP                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ├─→ Track: Sequence numbers, ACKs
                     ├─→ Track: Connection states
                     └─→ Reassemble: Fragmented packets
                     │
┌────────────────────▼────────────────────────────────────────┐
│                   NETWORK PROTOCOLS                         │
│                      IP, IPv6                               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     └─→ Extract: Addresses, TTL, flags
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    LINK LAYER                               │
│                 Ethernet, VLAN Tags                         │
└─────────────────────────────────────────────────────────────┘
```

Let's walk through an example to make this concrete. Imagine someone on your network visits a website. Here's what Zeek's protocol analysis does:




**Step 1: Link Layer Processing**

```
Ethernet Frame → Extract source/destination MAC addresses
              → Identify protocol (IPv4)
              → Strip Ethernet header
```

**Step 2: Network Layer Processing**

```
IP Packet → Extract source IP: 192.168.1.100
         → Extract destination IP: 93.184.216.34
         → Identify protocol (TCP)
         → Strip IP header
```

**Step 3: Transport Layer Processing**

```
TCP Segment → Extract source port: 52847
           → Extract destination port: 443
           → Track sequence numbers
           → Handle TCP flags (SYN, ACK, etc.)
           → Reassemble packets into data stream
           → Identify this as an HTTPS connection
```

**Step 4: Application Layer Processing**

```
TLS/SSL → Parse ClientHello
        → Extract TLS version: 1.3
        → Extract cipher suites offered
        → Parse ServerHello
        → Extract chosen cipher suite
        → Parse certificate
        → Extract subject: www.example.com
        → Extract issuer: DigiCert Inc.
        → Validate certificate chain
```

All of this analysis happens automatically. By the time your detection scripts see this connection, Zeek has already extracted every meaningful piece of information from it. Your scripts work with high-level abstractions like "HTTP request" or "SSL certificate" rather than raw packet bytes.

This multi-layer protocol analysis is what enables behavioural detection. You can ask questions like "Show me HTTPS connections where the certificate subject doesn't match the SNI hostname" or "Find HTTP POST requests with unusually large payloads sent to recently-registered domains." These questions would be nearly impossible to answer working at the packet level, but they're straightforward with Zeek's protocol analysis.




#### **Connection State Management: Tracking the Full Picture**

The second major responsibility of the event engine is maintaining state about network connections. Unlike stateless packet inspection, which examines each packet independently, Zeek tracks connections from establishment through termination, building a complete picture of each communication session.

For every connection, Zeek maintains a connection record that includes:

**Core Connection Information:**

```
┌──────────────────────────────────────────────────────┐
│ Connection UID: Unique identifier (C9x4Cz3...)       │
├──────────────────────────────────────────────────────┤
│ Source IP: 192.168.1.100                             │
│ Source Port: 52847                                   │
│ Destination IP: 93.184.216.34                        │
│ Destination Port: 443                                │
│ Protocol: TCP                                        │
├──────────────────────────────────────────────────────┤
│ Start Time: 2025-10-03 14:32:18.423                  │
│ Duration: 127.456 seconds                            │
│ Last Activity: 2025-10-03 14:34:25.879               │
├──────────────────────────────────────────────────────┤
│ Originator Bytes: 4,826                              │
│ Responder Bytes: 128,472                             │
│ Originator Packets: 89                               │
│ Responder Packets: 143                               │
├──────────────────────────────────────────────────────┤
│ Connection State: Established → Data Transfer →      │
│                   Closing → Closed                   │
│ State Flags: SF (normal establishment and close)     │
├──────────────────────────────────────────────────────┤
│ Service: SSL                                         │
│ Service Confidence: High                             │
└──────────────────────────────────────────────────────┘
```

But Zeek doesn't just track basic connection metadata. For protocols it understands, it maintains protocol-specific state:


**HTTP Connection State:**

```
┌──────────────────────────────────────────────────────┐
│ HTTP Transaction History:                            │
│                                                      │
│ Request 1:                                           │
│   Method: GET                                        │
│   URI: /index.html                                   │
│   User-Agent: Mozilla/5.0...                         │
│   Response Code: 200                                 │
│   Response Size: 4,826 bytes                         │
│                                                      │
│ Request 2:                                           │
│   Method: GET                                        │
│   URI: /style.css                                    │
│   User-Agent: Mozilla/5.0...                         │
│   Response Code: 200                                 │
│   Response Size: 12,340 bytes                        │
│                                                      │
│ [... additional requests ...]                        │
└──────────────────────────────────────────────────────┘
```






This stateful tracking enables sophisticated analysis. You can detect:

- Connections that established but never transferred data (potential scanning)
- Connections with unusual byte ratios (potential C2 beaconing)
- HTTP sessions where user agents change mid-session (potential session hijacking)
- Long-lived connections that periodically exchange small amounts of data (potential backdoors)

The state management also handles complex real-world scenarios. Zeek correctly handles:

- **TCP retransmissions**: Recognizing and deduplicating retransmitted packets
- **Out-of-order packets**: Reassembling them into the correct sequence
- **Fragmented packets**: Reassembling IP fragments and TCP segments
- **Connection reuse**: Tracking multiple HTTP transactions over a persistent connection
- **Timeouts**: Expiring state for connections that go idle


#### **Event Generation: Translating Protocol Activity Into Events**

The final responsibility of the event engine is generating events based on protocol activity. An event is a signal that something happened on the network - a connection was established, an HTTP request was made, a DNS query was issued, a file transfer began. Events are the interface between Zeek's core protocol analysis and your detection scripts.

Zeek generates events at multiple levels of granularity:

**Low-Level Network Events:**

```
new_connection()        - TCP connection established
connection_attempt()    - SYN packet seen (connection attempt)
connection_rejected()   - Connection refused (RST received)
connection_state_remove() - Connection finished
```

**Protocol-Specific Events:**

```
http_request()          - HTTP request observed
http_reply()            - HTTP response observed
dns_request()           - DNS query seen
dns_reply()             - DNS response received
ssl_established()       - SSL/TLS handshake completed
x509_certificate()      - X.509 certificate observed
smtp_request()          - SMTP command sent
smtp_reply()            - SMTP response received
```

**File Transfer Events:**

```
file_new()              - New file transfer detected
file_over_new_connection() - File transferred over new connection
file_hash()             - File hash calculated
```

**Connection Analysis Events:**

```
conn_stats()            - Periodic connection statistics
weird()                 - Protocol violation detected
```

When an event fires, it carries rich contextual information. For example, when the `http_request()` event fires, your script receives:

```
event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
```

Where:

- `c` is the full connection record with all state information
- `method` is the HTTP method (GET, POST, etc.)
- `original_URI` is the requested URI exactly as seen
- `unescaped_URI` is the URI with URL encoding resolved
- `version` is the HTTP version (1.0, 1.1, 2.0)

Your detection script can access all this information plus the full connection context. You might check if the URI matches a known-malicious pattern, if the requesting IP is on a watchlist, if the User-Agent is suspicious, if this is an unusual destination for this source, and more - all with the context Zeek has already gathered and structured for you.



### **Stage 4: The Script Layer - Where Your Detection Logic Lives**

After the event engine has parsed protocols, tracked state, and generated events, control passes to the script layer. This is where your detection logic executes. The script layer is implemented in Zeek's scripting language, which we'll explore in detail in Module 2, but it's important to understand its role in the architecture now.

#### **Event Handlers: Responding to Network Activity**

The primary way your scripts interact with Zeek is by defining event handlers - functions that are called when specific events occur. When you write a detection script, you're essentially telling Zeek "When this type of event happens, execute this code."

Here's a simple conceptual example (don't worry about the exact syntax yet):

```zeek
# When any HTTP request is observed...
event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
{
    # Check if the request is a POST to a suspicious path
    if ( method == "POST" && /\/api\/beacon/ in original_URI )
    {
        # This looks like potential C2 traffic
        # Generate an alert
        print fmt("Suspicious POST to %s from %s", original_URI, c$id$orig_h);
    }
}
```



When Zeek processes HTTP traffic, the event engine generates `http_request()` events. Each event triggers your handler function, passing in all the relevant information. Your code executes, performs whatever analysis you've programmed, and returns. The event engine continues processing traffic, generating more events, and calling your handlers.


#### **Execution Model: Single-Threaded Event Processing**

It's crucial to understand that Zeek's script layer is single-threaded within each Zeek process. Events are processed sequentially, one at a time. When an event handler is executing, no other events are processed until it completes.

```
Packet Arrives → Event Generated → Handler Executes → Handler Completes → Next Event
                                         ↓
                              Script execution blocks
                              other event processing
```

This has important implications:

**Performance Impact:** If your event handler is slow, it will slow down all event processing. Imagine you write a handler that does an expensive operation - perhaps querying a remote database - on every HTTP request. If each query takes 100 milliseconds and you're processing 1000 HTTP requests per second, you'd need 100 seconds to process one second of traffic. Zeek would fall hopelessly behind, dropping packets and missing events.

**Best Practices:**

- Keep event handlers fast and efficient
- Avoid blocking operations (network calls, disk I/O) in event handlers
- Use Zeek's built-in mechanisms for expensive operations (like the Intel Framework for lookups)
- Aggregate data and perform expensive analysis periodically rather than on every event

**State Sharing:** The single-threaded model does have an advantage: you don't need to worry about concurrency. All your event handlers share the same state, and you can modify shared data structures without locking or synchronization. When you track connection patterns across multiple events, you know your data structures won't be modified by another thread while you're using them.



#### **Frameworks: Structured Functionality for Common Tasks**

Zeek includes several frameworks - structured collections of scripts that provide sophisticated functionality for common security monitoring tasks. These frameworks execute in the script layer, but they're substantial enough that they feel like core Zeek features.

**Key Frameworks:**

|Framework|Purpose|Common Use Cases|
|---|---|---|
|**Intel Framework**|Match network activity against threat intelligence|Detect connections to known-malicious IPs, domains, URLs|
|**Notice Framework**|Generate and manage alerts|Create actionable alerts with context, suppression, and escalation|
|**Logging Framework**|Create custom log streams|Log specific detected behaviors in structured format|
|**File Analysis Framework**|Extract and analyze files|Extract executables, compute hashes, submit to sandboxes|
|**Input Framework**|Read data from external sources|Load configuration, read indicator lists, import data|
|**Cluster Framework**|Coordinate distributed deployment|Manage communication between cluster nodes|
|**SumStats Framework**|Perform statistical analysis|Count events, track thresholds, detect volumetric anomalies|

We'll work with these frameworks extensively in later modules. For now, understand that they provide powerful capabilities that your detection scripts can leverage. Rather than implementing statistical analysis from scratch, you can use the SumStats Framework. Rather than building threat intelligence matching logic, you can use the Intel Framework.




#### **The Logging System: Structured Output Generation**

The final stage in the script layer is logging. Based on the events processed and the analysis performed, Zeek generates structured log files that document network activity and detected threats.

Zeek's logs are tab-separated value (TSV) files with a specific structure:

```
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	conn
#open	2025-10-03-14-32-18
#fields	ts	uid	id.orig_h	id.orig_p	id.resp_h	id.resp_p	proto	service	duration	orig_bytes	resp_bytes	conn_state	local_orig	local_resp	missed_bytes	history	orig_pkts	orig_ip_bytes	resp_pkts	resp_ip_bytes	tunnel_parents
#types	time	string	addr	port	addr	port	enum	string	interval	count	count	string	bool	bool	count	string	count	count	count	count	set[string]
1696348338.423000	C9x4Cz3...	192.168.1.100	52847	93.184.216.34	443	tcp	ssl	127.456	4826	128472	SF	-	-	0	ShADadFf	89	8744	143	135294	(empty)
```

**Standard Log Files:**

Zeek generates many log types by default:

|Log File|Contents|
|---|---|
|`conn.log`|All network connections with metadata|
|`http.log`|HTTP requests and responses|
|`dns.log`|DNS queries and responses|
|`ssl.log`|SSL/TLS connection parameters|
|`x509.log`|Certificates observed|
|`files.log`|File transfers across protocols|
|`smtp.log`|Email transactions|
|`ssh.log`|SSH connections|
|`weird.log`|Protocol violations and anomalies|
|`notice.log`|Alerts generated by detection scripts|

Your detection scripts can also create custom log files to record specific behaviors you're tracking. We'll learn how to do this in Module 3.

The logs are designed to be both human-readable and machine-parseable. You can examine them with standard Unix tools (`less`, `grep`, `awk`), but they're also structured enough to be easily ingested by log analysis platforms, SIEMs, and custom analysis scripts.

---



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_1/validation.md" >}})
[|NEXT|]({{< ref "./cluster.md" >}})

