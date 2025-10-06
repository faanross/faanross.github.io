---
showTableOfContents: true
title: "Part 2 - Cluster Architecture"
type: "page"
---

## **PART 2: CLUSTER ARCHITECTURE - SCALING TO HIGH-BANDWIDTH NETWORKS**

### **Why Clusters Are Necessary**

The single-instance architecture we've explored works well for many networks, but it has inherent scalability limits. A single Zeek process running on one machine can typically handle somewhere between 500 Mbps and 2 Gbps of traffic, depending on the hardware, the traffic mix, and the complexity of your detection scripts. What happens when you need to monitor a 10 Gbps network? Or a 100 Gbps data center?

The answer is Zeek's cluster architecture, which distributes the analysis workload across multiple machines or processors. Rather than trying to build a bigger, faster single system (vertical scaling), Zeek scales horizontally by adding more systems that work together.

**Scaling approaches compared:**

```
┌─────────────────────────────────────────────────────────────┐
│                    VERTICAL SCALING                         │
│              (Single Faster Machine)                        │
│                                                             │
│   Traffic → │████████████│ → Analysis                       │
│             │One Powerful│                                  │
│             │   Zeek     │                                  │
│                                                             │
│   Limits: Hardware ceiling, cost increases exponentially    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   HORIZONTAL SCALING                        │
│                (Multiple Machines)                          │
│                                                             │
│                ┌─────────┐                                  │
│   Traffic →    │Worker 1 │ → Analysis                       │
│          ↓     └─────────┘                                  │
│          →     ┌─────────┐                                  │
│          ↓     │Worker 2 │ → Analysis                       │
│          →     └─────────┘                                  │
│          ↓     ┌─────────┐                                  │
│          →     │Worker 3 │ → Analysis                       │
│                └─────────┘                                  │
│                                                             │
│   Limits: Coordination overhead, but scales much further    │
└─────────────────────────────────────────────────────────────┘
```



### **Cluster Roles: Division of Labor**

A Zeek cluster consists of multiple nodes, each playing a specific role. Understanding these roles and how they interact is essential for designing effective deployments. There are four different types of nodes.


#### **1. Workers: The Packet Processing Engines**

Workers are the nodes that actually capture and analyze network traffic. In a cluster, traffic is distributed across multiple workers, with each worker handling a portion of the total traffic volume. Workers perform all the packet acquisition, protocol analysis, and event generation we discussed in the single-instance architecture.

**Worker characteristics:**

- Directly capture packets from network interfaces or load balancers
- Run the full event engine and script layer
- Generate logs from their portion of the traffic
- Send notices and summaries to the manager
- Require significant CPU and memory resources

**Typical worker deployment:**

```
Network → Load Balancer → Worker 1 (handles 25% of traffic)
                      ├─→ Worker 2 (handles 25% of traffic)
                      ├─→ Worker 3 (handles 25% of traffic)
                      └─→ Worker 4 (handles 25% of traffic)
```

The number of workers you need depends on your traffic volume and the complexity of your analysis. A good rule of thumb is that each worker can handle 500 Mbps to 2 Gbps, so a 10 Gbps network might need 5-10 workers depending on configuration and hardware.

#### **2. Manager: The Coordination Hub**

The manager node coordinates the cluster, receives notices from workers, and makes cluster-wide decisions. Unlike workers, the manager doesn't directly analyze network traffic. Instead, it performs administrative and coordination tasks.

**Manager responsibilities:**

```
┌─────────────────────────────────────────────────────────────┐
│                   MANAGER NODE DUTIES                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ► Receive notices from all workers                         │
│    (Workers send alerts to manager for deduplication)       │
│                                                             │
│  ► Aggregate and deduplicate notices                        │
│    (If multiple workers detect same threat, merge alerts)   │
│                                                             │
│  ► Make cluster-wide policy decisions                       │
│    (Determine if behavior is significant across all workers)│
│                                                             │
│  ► Distribute updated intelligence/configuration            │
│    (Push new indicators to all workers)                     │
│                                                             │
│  ► Generate summary statistics                              │
│    (Aggregate metrics from all workers)                     │
│                                                             │
│  ► Coordinate response actions                              │
│    (Orchestrate cluster-wide responses to threats)          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

The manager is typically a less powerful machine than the workers since it's not processing high-volume traffic. However, it requires good network connectivity to all workers and sufficient resources to handle the volume of notices and coordination messages.

#### **3. Proxy: The Optional Load Distribution Layer**

Proxies are optional intermediary nodes that sit between workers and the manager. They aggregate data from multiple workers before forwarding it to the manager, reducing the connection and processing load on the manager.

**When proxies are valuable:**

|Cluster Size|Proxy Recommendation|Reasoning|
|---|---|---|
|1-5 workers|Not needed|Manager can handle direct connections|
|6-20 workers|One proxy|Reduces manager load significantly|
|20+ workers|Multiple proxies|Essential for scalability|

**Proxy communication flow:**

```
Worker 1  ┐
Worker 2  ├─→ Proxy 1 ─┐
Worker 3  ┘            ├─→ Manager
                       │
Worker 4  ┐            │
Worker 5  ├─→ Proxy 2 ─┘
Worker 6  ┘
```

Proxies aggregate and deduplicate data before sending it to the manager. For example, if three workers behind a proxy all detect connections to the same malicious IP, the proxy can aggregate these into a single notice with a count, rather than sending three separate notices to the manager.

#### **4. Logger: The Centralized Log Collector**

The logger node receives log data from all workers and writes consolidated log files. Rather than having each worker write its own log files (which would result in fragmented logs across multiple machines), the logger centralizes logging.

**Logger architecture:**

```
Worker 1 ─┐
Worker 2 ─┼─→ Logger ─→ Consolidated Logs
Worker 3 ─┘              (conn.log, http.log, etc.)
```

**Advantages of centralized logging:**

- Single location for all log data (easier analysis)
- Consistent log file rotation and management
- Reduced storage requirements (no duplication)
- Simplified log forwarding to SIEM

**Storage considerations:** The logger needs substantial storage capacity. As a rough estimate:

- 1 Gbps network: ~50-100 GB per day
- 10 Gbps network: ~500 GB - 1 TB per day
- Storage needs vary greatly based on traffic composition and retention requirements


### **Cluster Communication: How Nodes Coordinate**

The nodes in a Zeek cluster need to communicate with each other to coordinate analysis, share state, and aggregate results. This communication is handled by Zeek's Broker library, a high-performance pub/sub messaging system designed specifically for Zeek.

**Broker communication model:**

```
┌──────────────────────────────────────────────────────────────┐
│                    ZEEK CLUSTER COMMUNICATION                │
│                                                              │
│                        Manager                               │
│                           ▲                                  │
│                           │                                  │
│            ┌──────────────┼──────────────┐                   │
│            │              │              │                   │
│            ▼              ▼              ▼                   │
│         Proxy 1        Proxy 2        Proxy 3                │
│            ▲              ▲              ▲                   │
│       ┌────┼────┐    ┌────┼────┐    ┌────┼────┐              │
│       │    │    │    │    │    │    │    │    │              │
│       ▼    ▼    ▼    ▼    ▼    ▼    ▼    ▼    ▼              │
│      W1   W2   W3   W4   W5   W6   W7   W8   W9              │
│                                                              │
│  Communication Types:                                        │
│  ════════════════════                                        │
│  W→P: Logs, Notices, Metrics (high volume)                   │
│  P→M: Aggregated data (moderate volume)                      │
│  M→P: Policy updates, Intel feeds (low volume)               │
│  P→W: Distributed policy (low volume)                        │
└──────────────────────────────────────────────────────────────┘
```

**What gets communicated:**

|Data Type|Direction|Purpose|Volume|
|---|---|---|---|
|Log entries|Workers → Logger|Centralized logging|Very High|
|Notices|Workers → Manager|Alerts and detections|Medium|
|Metrics|Workers → Manager|Performance monitoring|Low|
|Intel updates|Manager → Workers|Threat intelligence|Low|
|State synchronization|Bidirectional|Shared analysis state|Variable|

### **Load Balancing: Distributing Traffic Across Workers**

One of the most critical aspects of cluster deployment is how you distribute network traffic across workers. If the distribution is uneven, some workers will be overwhelmed while others sit idle. If it doesn't maintain connection affinity, workers won't have complete visibility into connections.

**Load balancing requirements:**

```
┌──────────────────────────────────────────────────────────────┐
│          REQUIREMENTS FOR EFFECTIVE LOAD BALANCING           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  1. BALANCE                                                  │
│     Each worker should receive approximately equal traffic   │
│     volume to prevent overload                               │
│                                                              │
│  2. AFFINITY                                                 │
│     All packets from the same connection must go to the      │
│     same worker so it can maintain complete state            │
│                                                              │
│  3. STABILITY                                                │
│     Load distribution should be consistent - same connection │
│     should always go to same worker                          │
│                                                              │
│  4. SCALABILITY                                              │
│     Should support adding/removing workers without           │
│     completely redistributing all traffic                    │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**Common load balancing methods:**

#### **Method 1: Hardware Load Balancers**

Expensive network switches and load balancers can distribute traffic based on connection 5-tuples (source IP, source port, destination IP, destination port, protocol). They use hashing to ensure connection affinity while distributing load.

**Advantages:**

- Purpose-built hardware with line-rate performance
- No overhead on Zeek workers
- Supports very high bandwidth (40 Gbps+)

**Disadvantages:**

- Expensive
- Requires specialized knowledge
- Single point of failure (unless redundant)


#### **Method 2: AF_PACKET Fanout**

On Linux, `AF_PACKET` supports "fanout" mode, which distributes packets across multiple processes on the same host. Each worker process gets a portion of the traffic from a shared network interface.

```
┌────────────────────────────────────────────────────────────┐
│               Single Host with Multiple Workers            │
│                                                            │
│   Network Interface (10 Gbps)                              │
│         │                                                  │
│         ├──────────┬──────────┬──────────┬──────────       │
│         │          │          │          │                 │
│         ▼          ▼          ▼          ▼                 │
│     Worker 1   Worker 2   Worker 3   Worker 4              │
│     (2.5G)     (2.5G)     (2.5G)     (2.5G)                │
│                                                            │
│   AF_PACKET fanout distributes packets with connection     │
│   affinity maintained through hash of 5-tuple              │
└────────────────────────────────────────────────────────────┘
```

**Configuration:**

```
# In node.cfg
[worker-1]
type=worker
host=sensor-01
interface=af_packet::eth0
af_packet_fanout_id=23
af_packet_fanout_mode=PACKET_FANOUT_HASH
```

**Advantages:**

- No additional hardware required
- Excellent performance
- Built into Linux kernel
- Free

**Disadvantages:**

- Limited to single host (can't distribute across multiple machines)
- Requires modern Linux kernel


#### **Method 3: PF_RING Clustering**

`PF_RING` includes sophisticated clustering capabilities that can distribute traffic across multiple processes or even multiple hosts.

**Advantages:**

- Best performance for very high bandwidth
- Can distribute across multiple physical hosts
- Supports hardware offload with compatible NICs

**Disadvantages:**

- Requires PF_RING license
- More complex setup
- Additional cost

### **Cluster Deployment Patterns**

Let's look at some common cluster deployment patterns for different scales and requirements.

#### **Small Cluster (1-5 Gbps)**

```
┌─────────────────────────────────────────────────────────────┐
│              SMALL CLUSTER ARCHITECTURE                     │
│                                                             │
│                   ┌───────────┐                             │
│                   │ Manager   │                             │
│                   │ + Logger  │ (Combined on one host)      │
│                   └─────┬─────┘                             │
│                         │                                   │
│             ┌───────────┼───────────┐                       │
│             │           │           │                       │
│             ▼           ▼           ▼                       │
│        ┌────────┐  ┌────────┐  ┌────────┐                   │
│        │Worker 1│  │Worker 2│  │Worker 3│                   │
│        └────────┘  └────────┘  └────────┘                   │
│                                                             │
│  Hardware:                                                  │
│  - Manager/Logger: 4 cores, 16GB RAM, 1TB storage           │
│  - Each Worker: 8 cores, 32GB RAM, modest storage           │
│                                                             │
│  Best for: Small to medium enterprises, branch offices      │
└─────────────────────────────────────────────────────────────┘
```

#### **Medium Cluster (5-20 Gbps)**

```
┌─────────────────────────────────────────────────────────────┐
│              MEDIUM CLUSTER ARCHITECTURE                    │
│                                                             │
│               ┌──────────┐        ┌──────────┐              │
│               │ Manager  │        │  Logger  │              │
│               └────┬─────┘        └────┬─────┘              │
│                    │                   │                    │
│                    └─────────┬─────────┘                    │
│                              │                              │
│                    ┌─────────┴─────────┐                    │
│                    │                   │                    │
│                    ▼                   ▼                    │
│               ┌────────┐          ┌────────┐                │
│               │Proxy 1 │          │Proxy 2 │                │
│               └───┬────┘          └───┬────┘                │
│                   │                   │                     │
│        ┌──────────┼──────┐   ┌────────┼──────────┐          │
│        │          │      │   │        │          │          │
│        ▼          ▼      ▼   ▼        ▼          ▼          │
│    ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐        │
│    │Work 1 │ │Work 2 │ │Work 3 │ │Work 4 │ │Work 5 │        │
│    └───────┘ └───────┘ └───────┘ └───────┘ └───────┘        │
│        ...                           ...                    │
│                                                             │
│  Hardware:                                                  │
│  - Manager: 8 cores, 32GB RAM                               │
│  - Logger: 4 cores, 16GB RAM, 5TB+ storage                  │
│  - Proxies: 4 cores, 16GB RAM each                          │
│  - Workers: 12+ cores, 64GB RAM each                        │
│                                                             │
│  Best for: Large enterprises, service providers             │
└─────────────────────────────────────────────────────────────┘
```

#### **Large Cluster (20+ Gbps)**

```
┌──────────────────────────────────────────────────────────────┐
│               LARGE CLUSTER ARCHITECTURE                     │
│                                                              │
│                      ┌──────────┐                            │
│                      │ Manager  │                            │
│                      └────┬─────┘                            │
│                           │                                  │
│        ┌──────────────────┼──────────────────┐               │
│        │                  │                  │               │
│        ▼                  ▼                  ▼               │
│   ┌────────┐        ┌────────┐        ┌────────┐             │
│   │Proxy 1 │        │Proxy 2 │        │Proxy 3 │             │
│   └───┬────┘        └───┬────┘        └───┬────┘             │
│       │                 │                 │                  │
│    [Workers]         [Workers]         [Workers]             │
│    Rack 1            Rack 2            Rack 3                │
│    W1-W8             W9-W16            W17-W24               │
│                                                              │
│              Separate Logging Infrastructure:                │
│                                                              │
│                      ┌──────────┐                            │
│                      │ Logger   │                            │
│                      │ Frontend │                            │
│                      └────┬─────┘                            │
│                           │                                  │
│              ┌────────────┼────────────┐                     │
│              │            │            │                     │
│              ▼            ▼            ▼                     │
│         ┌────────┐   ┌────────┐   ┌────────┐                 │
│         │Storage │   │Storage │   │Storage │                 │
│         │ Node 1 │   │ Node 2 │   │ Node 3 │                 │
│         └────────┘   └────────┘   └────────┘                 │
│                                                              │
│  Best for: Data centers, ISPs, critical infrastructure       │
└──────────────────────────────────────────────────────────────┘
```





---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./pipeline.md" >}})
[|NEXT|]({{< ref "./optimize.md" >}})

