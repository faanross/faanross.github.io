---
showTableOfContents: true
title: "Part 8 - Practical Exercises"
type: "page"
---



## **Exercise 1: Installation Verification and Troubleshooting**

Let's verify your installation is working correctly through a series of tests:

### **Task 1: Verify All Core Components**

Create a checklist and verify each item:

```bash
# 1. Zeek binary exists and runs
zeek --version
```


```bash
# 2. ZeekControl is functional
sudo zeekctl status
```


```bash
# 3. Configuration files are present
ls -la /opt/zeek/etc/
```


```bash
# 4. Log directory exists and is writable
ls -la /opt/zeek/logs/
```


```bash
# 5. Zeek can see network interfaces
zeek -i eth0 --version  # Should not error about interface
```


```bash
# 6. Basic scripts load without errors
zeek -C -r /opt/zeek/share/zeek/test-pcap.pcap
```



### **Task 2: Performance Baseline**

Establish a performance baseline for your system:

```bash
# Check current packet drop rate
sudo zeekctl netstats

# Generate sustained traffic and monitor
# In one terminal:
while true; do curl -s http://example.com > /dev/null; sleep 1; done

# In another terminal, monitor:
watch "tail /opt/zeek/logs/current/capture_loss.log"
```

Document:

- Peak memory usage
- Packet capture rate
- Any dropped packets
- CPU usage (use `top` or `htop`)

### **Task 3: Log Analysis Verification**

Generate specific traffic and verify it's logged correctly:

```bash
# 1. Make HTTP request
curl http://neverssl.com

# 2. Verify it appears in logs (wait 10 seconds)
grep neverssl /opt/zeek/logs/current/http.log

# 3. Make DNS query
dig example.com

# 4. Verify DNS logging
grep example.com /opt/zeek/logs/current/dns.log

# 5. Make HTTPS request
curl https://github.com

# 6. Verify SSL/TLS logging
grep github /opt/zeek/logs/current/ssl.log
```

**Deliverable:**

Create a document with:

- Screenshot or copy-paste of each verification step
- Any errors encountered and how you fixed them
- Performance baseline metrics
- Confirmation that all log types are being generated

## **Exercise 2: Configuration Practice**

Practice modifying Zeek's configuration:

### **Task 1: Customize networks.cfg**

Edit `/opt/zeek/etc/networks.cfg` to accurately reflect your environment:

```bash
sudo nano /opt/zeek/etc/networks.cfg
```

**Add:**
- Your droplet's specific IP and subnet
- Any other networks you'll be monitoring
- Appropriate descriptions

### **Task 2: Modify local.zeek**

Add additional scripts to your configuration:

```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```

For now we'll just keep it simple, in a future lesson we'll unpack local.zeek in much more detail. Uncomment the following line which should be present, if not simply add it.

```zeek
@load policy/protocols/ssl/heartbleed
```

### **Task 3: Test Configuration Changes**

```bash
# Check for errors
sudo zeekctl check
# should say: zeek scripts are ok.

# If no errors, deploy
sudo zeekctl deploy
# final line should read: starting zeek ...

# Verify Zeek started successfully
sudo zeekctl status
# Status should read: running

# Check logs for any startup errors
tail /opt/zeek/logs/current/stderr.log
# Should only report: listening on eth0
```

### **Task 4: Tune Performance Settings**

If your system has multiple cores, modify node.cfg to use AF_PACKET with multiple processes:

```bash
sudo nano /opt/zeek/etc/node.cfg
```

Experiment with different settings:

- Change `lb_procs` based on available CPU cores
- Adjust `af_packet_buffer_size`
- Monitor impact on performance

**Deliverable:**

Document:

- Your customized networks.cfg (sanitize any sensitive IPs)
- Scripts you added to local.zeek and why
- AF_PACKET configuration and observed performance impact
- Any interesting events detected after enabling additional scripts

## **Exercise 3: Generate and Analyze Realistic Traffic**

Create realistic network traffic scenarios and observe how Zeek logs them:

### **Scenario 1: Web Browsing Simulation**

```bash
# Simulate normal web browsing
curl http://www.example.com
curl http://www.github.com
curl http://stackoverflow.com
curl http://reddit.com

# Wait for logging
sleep 10

# Analyze the captured HTTP traffic
zeek-cut ts host method uri status_code < /opt/zeek/logs/current/http.log | tail -20
```

### **Scenario 2: DNS Activity**

```bash
# Simulate various DNS queries
for domain in google.com facebook.com twitter.com amazon.com; do
    dig $domain
    dig @8.8.8.8 $domain
done

# Analyze DNS logs
zeek-cut ts query qtype_name answers < /opt/zeek/logs/current/dns.log | tail -20
```

### **Scenario 3: Multiple Protocols**

```bash
# FTP attempt (will likely fail, but generates traffic)
ftp ftp.gnu.org <<EOF
quit
EOF

# SSH attempt
ssh -o ConnectTimeout=5 test@example.com

# Web traffic
curl https://api.github.com/users/zeek

# Check what services Zeek identified
zeek-cut id.orig_h id.resp_h service < /opt/zeek/logs/current/conn.log | tail -20
```

### **Scenario 4: Simulate Suspicious Behavior**

```bash
# Port scan simulation (from your droplet to itself)
for port in 22 23 80 443 3389 8080; do
    nc -zv 127.0.0.1 $port 2>&1
done

# Check if Zeek detected scanning behavior
cat /opt/zeek/logs/current/notice.log
grep -i scan /opt/zeek/logs/current/weird.log
```

**Analysis Tasks:**

For each scenario, document:

1. What logs were generated (which .log files)
2. What information Zeek extracted
3. How accurate was service identification
4. Any notices or weird events generated

**Deliverable:**

Create a report with:

- Commands you ran for each scenario
- Interesting log excerpts showing what Zeek captured
- Analysis of Zeek's detection accuracy
- Any unexpected behaviors or surprises

## **Exercise 4: Troubleshooting Practice**

Deliberately break your Zeek installation in controlled ways, then fix it. This builds troubleshooting skills:

### **Break 1: Invalid Interface**

```bash
sudo nano /opt/zeek/etc/node.cfg
# Change interface to 'eth999' (doesn't exist)

sudo zeekctl deploy
# It will fail - read the error message carefully

# Fix it
# Change interface back to correct value

sudo zeekctl deploy
# Should succeed now
```

### **Break 2: Syntax Error in Script**

```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
# Add a line with intentional syntax error:
# this is broken syntax!!!

sudo zeekctl check
# Read the error message

# Fix the error
sudo zeekctl deploy
```

### **Break 3: Permission Issues**

```bash
# Make log directory unwritable
sudo chmod 000 /opt/zeek/logs/current

# Try to start Zeek
sudo zeekctl restart

# Check logs for errors
sudo cat /opt/zeek/logs/current/stderr.log

# Fix permissions
sudo chmod 755 /opt/zeek/logs/current
```

**Deliverable:**

Document each break:

- What you broke
- What error message appeared
- How you diagnosed the problem
- How you fixed it
- What you learned

This exercise builds confidence in troubleshooting real problems.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./management.md" >}})
[|NEXT|]({{< ref "./validation.md" >}})

