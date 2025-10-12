---
showTableOfContents: true
title: "Part 9 - Knowledge Validation"
type: "page"
---


## **Question 1: Installation Methods**

You're deploying Zeek in three different scenarios:

- A. Learning environment on a laptop
- B. Production 5 Gbps network monitoring
- C. Development/testing with frequent Zeek version changes

For each scenario, which installation method would you choose (package, source, or container) and why?

### Answer

**A. Learning environment on a laptop**

- **Choose: Container (Docker)**
- **Why:** Quick setup, isolated environment, easy to reset/remove, minimal system impact, works across different OS platforms

**B. Production 5 Gbps network monitoring**

- **Choose: Source compilation**
- **Why:** Optimized performance for your specific hardware, can enable PF_RING or AF_PACKET optimizations, full control over compile-time features, better performance for high-throughput environments

**C. Development/testing with frequent version changes**

- **Choose: Package or Container**
- **Why:** Package managers allow quick version switching without recompilation. Containers offer even better isolation between versions, allowing multiple versions to coexist for regression testing






**Question 2: Configuration Hierarchy**

Explain the relationship between these configuration locations:

- `/opt/zeek/etc/node.cfg`
- `/opt/zeek/etc/networks.cfg`
- `/opt/zeek/share/zeek/site/local.zeek`

What does each control, and in what order would you configure them when setting up a new sensor?

### Answer
**Relationship & Purpose:**

- **`/opt/zeek/etc/node.cfg`**: Defines cluster topology - which nodes run where, worker assignments, CPU pinning, load balancing settings
- **`/opt/zeek/etc/networks.cfg`**: Defines internal/local networks for Zeek to distinguish internal vs external traffic
- **`/opt/zeek/share/zeek/site/local.zeek`**: Script-level configuration - loads scripts, sets protocol analyzers, defines custom logging, intelligence feeds

**Configuration Order:**

1. **First: `node.cfg`** - Establishes the deployment architecture (standalone vs cluster, interface assignments)
2. **Second: `networks.cfg`** - Defines what traffic is "local" for proper directional analysis
3. **Third: `local.zeek`** - Loads policy scripts and customizations that operate on the architecture you've defined



**Question 3: AF_PACKET Configuration**

Your droplet has 8 vCPUs. You configure node.cfg with:

```ini
lb_procs=8
pin_cpus=0,1,2,3,4,5,6,7
```

Is this optimal? Why or why not? What would you recommend instead?

### Answer
**This is NOT optimal.**

**Problem:** Using all 8 vCPUs for Zeek workers leaves no CPU resources for:

- The operating system
- The manager process
- The proxy process (in cluster mode)
- Other system processes

**Recommended configuration:**

```ini
lb_procs=6
pin_cpus=2,3,4,5,6,7
```

**Rationale:** Reserve CPUs 0-1 for system processes and Zeek management components. Use 6 workers on the remaining cores to maintain system stability and avoid resource contention. This prevents CPU starvation and ensures the system remains responsive.


**Question 4: Log Analysis**

You run `zeek-cut id.orig_h id.resp_h conn_state < conn.log` and see many entries with state "S0". What does this likely indicate? Should you be concerned?

### Answer
**S0 Connection State Meaning:** Connection attempt seen, no reply

**What it indicates:**

- Port scans (most common)
- Connections to closed ports
- Filtered/dropped traffic
- Failed connection attempts
- Potential reconnaissance activity

**Should you be concerned?**

- **Volume matters:** A few S0 entries are normal network behavior
- **Many S0s to many destinations:** Likely port scanning - investigate the source
- **Pattern analysis:** Check if originator is internal (compromised host scanning) or external (reconnaissance)
- **Review related logs:** Check `notice.log` and `weird.log` for correlated anomalies



**Question 5: Troubleshooting**

After deploying a configuration change, `zeekctl status` shows your Zeek instance as "crashed". What are your first three troubleshooting steps?

### Answer

**First three steps:**

1. **Check stderr.log and stdout.log**

```bash
   tail -50 /opt/zeek/logs/current/stderr.log
   tail -50 /opt/zeek/logs/current/stdout.log
```

Look for error messages, script syntax errors, or crash reasons

2. **Verify configuration syntax**

```bash
zeekctl check
```

This validates scripts and configuration files before attempting restart

3. **Check system resources and reporter.log**


```bash
   cat /opt/zeek/logs/current/reporter.log
   df -h  # disk space
   free -m  # memory
```

Look for resource exhaustion, permission issues, or Zeek-specific warnings



**Question 6: Performance Monitoring**

Your `capture_loss.log` shows 3% packet drops during peak hours. What might be causing this, and what are three possible solutions?

Take time to think through these questions. They test practical knowledge you'll use constantly.

### Answer

**Likely causes:**

- Insufficient CPU resources for traffic volume
- Buffer sizes too small
- Worker imbalance in load distribution
- Disk I/O bottlenecks (slow logging)
- NIC ring buffer exhaustion

**Three possible solutions:**

1. **Increase worker processes** (if CPU available)
    - Add more `lb_procs` to distribute packet processing load
    - Ensure proper CPU pinning to avoid context switching
2. **Tune AF_PACKET buffer sizes**


```ini
   af_packet_buffer_size=128*1024*1024  # Increase from default
```

Larger buffers provide more headroom during traffic bursts

3. **Optimize logging strategy**
    - Disable unnecessary logs or reduce retention
    - Use remote logging to faster storage
    - Implement log filtering to reduce volume
    - Consider JSON logging for better write performance

**Additional consideration:** Verify NIC offloading settings (checksum, segmentation) are properly configured for packet capture.



---

## **PREPARING FOR MODULE 2**

Congratulations! You've successfully completed Module 1. You now have a working Zeek sensor that's capturing and analyzing network traffic. Let's summarize what you've accomplished and preview what's ahead.

**What You've Mastered:**

You understand Zeek's history, philosophy, and architecture. You can explain why behavioural analysis differs from signature-based detection and articulate when each approach is appropriate. You understand the event-driven processing pipeline and can describe how packets flow from the network interface through protocol analysis to log generation.

You've explored three different installation methods and understand the trade-offs of each. You can compile Zeek from source, install from packages, and deploy in containers. You understand when each method is appropriate for different scenarios.

You've configured Zeek for optimal performance on your system, including AF_PACKET setup for improved packet capture. You understand the role of each configuration file and can customize Zeek's behaviour for your specific needs.

You can operate Zeek confidently using ZeekControl, managing startup and shutdown, deploying configuration changes, and monitoring system health. You understand log rotation, backup strategies, and basic troubleshooting procedures.

Most importantly, you've seen Zeek in action, generating real logs from real traffic. You understand what Zeek logs contain, how to parse them with zeek-cut, and how to interpret the information Zeek extracts from network traffic.

**What's Coming in Module 2:**

Module 2 is where we transition from operating Zeek to programming it. You're going to learn Zeek's scripting language from the ground up, starting with basic syntax and progressing to writing your first detection scripts.

We'll cover data types, variables, functions, and control flow. You'll learn how to respond to events, maintain state across multiple events, and implement detection logic. By the end of Module 2, you'll write scripts that detect port scans, SSH brute-force attacks, unusual DNS patterns, and basic SQL injection attempts.

Module 2 is intensive - there's a lot of programming to learn. But you have a solid foundation now. Your working Zeek sensor provides the perfect laboratory for experimenting with script development. Every concept we cover, you'll immediately test on your sensor, seeing how your code affects what Zeek detects.

**Before Module 2 Begins:**

Spend some time exploring your Zeek installation:

- Generate various types of traffic and watch how Zeek logs it
- Experiment with enabling different scripts in local.zeek
- Browse `/opt/zeek/share/zeek/base/` to see existing Zeek scripts
- Try parsing logs in different ways with zeek-cut

Review basic programming concepts if you're new to scripting:

- Variables and data types
- Conditional statements (if/else)
- Loops (for, while)
- Functions

You don't need to be an expert programmer - Zeek's language is learnable for beginners. But familiarity with basic programming concepts will accelerate your learning.

**Final Thoughts on Module 1:**

Installation and configuration might seem like dry material, but it's the essential foundation for everything that follows. You can't write effective detection scripts if you don't understand how Zeek processes traffic. You can't troubleshoot performance problems if you don't understand the architecture. You can't deploy Zeek in production if you don't know how to configure it properly.

The time you've invested in thoroughly understanding Zeek's foundation will pay enormous dividends. When you're writing sophisticated detection logic in later modules, you'll understand why certain approaches work and others don't. When you're hunting threats in production logs, you'll know how to optimize queries and what information is available. When you're deploying sensors across your organization, you'll design them correctly the first time.

Take pride in what you've accomplished. You've built a working network security monitoring system from scratch. You understand how it works from the packet level to the log level. You're ready to start writing detection logic that will identify real threats.

**A Note on Pace:**

Module 1 was dense with information. If you feel overwhelmed, that's normal. You don't need to have everything memorized - this course document is your reference. What matters is that you understand the concepts and know where to look when you need details.

Module 2 will be different. Instead of absorbing concepts, you'll be writing code. The learning will be more active, more hands-on, and for many students, more engaging. Theory becomes practice. Abstract becomes concrete.

Take a break if you need one. When you're ready, Module 2 awaits - where we transform you from a Zeek operator into a Zeek programmer.



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./exercises.md" >}})
[|NEXT|]({{< ref "../../module02/lesson2_1/type.md" >}})

