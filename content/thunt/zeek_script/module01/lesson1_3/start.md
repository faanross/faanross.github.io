---
showTableOfContents: true
title: "Part 6 - Starting Zeek and Capturing Your First Traffic"
type: "page"
---



## **PART 6: STARTING ZEEK AND CAPTURING YOUR FIRST TRAFFIC**

### A Note on ZeekControl
We'll now use `zeekctl` to start using Zeek. Note that ZeekControl is as of writing still the recommended solution for multi-system clusters and those needing rich management capabilities [Management Framework — Book of Zeek (v8.0.1)](https://docs.zeek.org/en/current/frameworks/management.html).

However, Zeek has been developing an alternative called the **Management Framework** (with the `zeek-client`command-line tool). The Management Framework currently targets single-instance deployments, where traffic monitoring happens on a single system.

While it technically supports clusters spanning multiple monitoring systems, much of the infrastructure, such as the ability to deploy Zeek scripts and additional configuration, is not yet available in the Management Framework.

We will thus use zeekctl since it gives us the "full experience", but keep an eye on this development, and it might well be that in the near future you'd use `zeekctl` exclusively for multi-system clusters and `zeek-client` for single-instance deployments.



### **Deploying Zeek with ZeekControl**


Now comes the exciting moment - starting Zeek and capturing real network traffic! We'll use ZeekControl to manage Zeek:

```bash
# Start ZeekControl
sudo zeekctl
```

Note, if you are having trouble invoking `zeekctl` using `sudo`, create the following symlink:

```bash
sudo ln -s /opt/zeek/bin/zeekctl /usr/local/bin/zeekctl
```


You'll see the ZeekControl prompt:

```
Welcome to ZeekControl 2.6.0-28

Type "help" for help.

[ZeekControl] >
```

**Deploy Zeek for the first time:**

```
[ZeekControl] > install
[ZeekControl] > deploy
```


The output should look as follows:

```
[ZeekControl] > install
creating policy directories ...
installing site policies ...
generating standalone-layout.zeek ...
generating local-networks.zeek ...
generating zeekctl-config.zeek ...
generating zeekctl-config.sh ...
[ZeekControl] > deploy
checking configurations ...
installing ...
removing old policies in /opt/zeek/spool/installed-scripts-do-not-touch/site ...
removing old policies in /opt/zeek/spool/installed-scripts-do-not-touch/auto ...
creating policy directories ...
installing site policies ...
generating standalone-layout.zeek ...
generating local-networks.zeek ...
generating zeekctl-config.zeek ...
generating zeekctl-config.sh ...
stopping ...
stopping zeek ...
starting ...
starting zeek ...
```


**Understanding these commands:**

```
┌──────────────────────────────────────────────────────────────┐
│              ZEEKCONTROL COMMANDS                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  install                                                     │
│  └─ Install configuration changes                            │
│     Run this after modifying config files                    │
│                                                              │
│  deploy                                                      │
│  └─ Stop (if running), install config, and start Zeek        │
│     Essentially a "restart with new config" command          │
│                                                              │
│  start                                                       │
│  └─ Start Zeek instances                                     │
│                                                              │
│  stop                                                        │
│  └─ Stop Zeek instances gracefully                           │
│                                                              │
│  restart                                                     │
│  └─ Stop and start (without reinstalling config)             │
│                                                              │
│  status                                                      │
│  └─ Show status of all nodes                                 │
│                                                              │
│  check                                                       │
│  └─ Verify configuration without making changes              │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```



**Check Zeek's status:**

```
[ZeekControl] > status
```

You should see:

```
[ZeekControl] > status
Name         Type       Host          Status    Pid    Started
zeek         standalone localhost     running   51924  09 Oct 15:22:23
```

If Status shows "running", congratulations! Zeek is live and capturing traffic.



**If Status shows "crashed" or "stopped":**

Don't panic. Check the output from deploy - it usually tells you what went wrong. Common issues:

```
┌──────────────────────────────────────────────────────────────┐
│           COMMON STARTUP PROBLEMS                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ERROR: Permission denied on interface                       │
│  └─ Solution: Run zeekctl with sudo                          │
│                                                              │
│  ERROR: Interface eth0 not found                             │
│  └─ Solution: Check interface name with 'ip link show'       │
│     Update node.cfg with correct interface                   │
│                                                              │
│  ERROR: Syntax error in local.zeek                           │
│  └─ Solution: Check local.zeek for typos                     │
│     Zeek is sensitive to syntax errors                       │
│                                                              │
│  ERROR: Address already in use                               │
│  └─ Solution: Another Zeek instance might be running         │
│     Run 'zeekctl stop' then 'zeekctl start'                  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


**Exit ZeekControl:**

```
[ZeekControl] > exit
```



### **Understanding What Zeek is Doing**

Now that Zeek is running, what is it actually doing? Let's trace the activity:

**1. Packet Capture:** Zeek is reading packets from `eth0` using `AF_PACKET` (or `libpcap` if you didn't configure `AF_PACKET`). Every packet that flows through that interface is being captured.

**2. Protocol Analysis:** The event engine is parsing protocols, extracting metadata, and tracking connection state as we discussed in Lesson 1.2.

**3. Script Execution:** The scripts loaded in local.zeek are processing events and generating logs.

**4. Log Generation:** Zeek is writing structured logs to `/opt/zeek/logs/current/`.

Let's look at those logs:

```bash
ls -lh /opt/zeek/logs/current/
```



You'll see files like:

```
zeek@zeek-sensor:/usr/local/bin$ ls -lh /opt/zeek/logs/current/
total 304K
-rw-r--r-- 1 root zeek  250 Oct  9 15:23 capture_loss.log
-rw-r--r-- 1 root zeek  31K Oct  9 15:35 conn.log
-rw-r--r-- 1 root zeek  34K Oct  9 15:22 loaded_scripts.log
-rw-r--r-- 1 root zeek  753 Oct  9 15:23 notice.log
-rw-r--r-- 1 root zeek  251 Oct  9 15:22 packet_filter.log
-rw-r--r-- 1 root zeek  666 Oct  9 15:22 reporter.log
-rw-r--r-- 1 root zeek 3.3K Oct  9 15:35 ssh.log
-rw-r--r-- 1 root zeek  914 Oct  9 15:32 stats.log
-rw-r--r-- 1 root zeek   19 Oct  9 15:22 stderr.log
-rw-r--r-- 1 root zeek  204 Oct  9 15:22 stdout.log
-rw-r--r-- 1 root zeek 189K Oct  9 15:35 telemetry.log
-rw-r--r-- 1 root zeek 5.6K Oct  9 15:35 weird.log
```

The exact files present depend on what protocols are being used on your network right now.


### **Examining Your First Logs**

NOTE: We will do a thorough exploration of each default log in future lessons, here really I just want to give you a quick introduction and preview, and enough commands so that you can feel emboldened to go an explore the data in the meantime.




### **Generating Test Traffic**

If you're using a brand new VM  you probably don't have much traffic yet. Let's generate some so we can see Zeek in action:

**Generate web traffic:**

```bash
# Make some HTTP/S requests
curl http://example.com
curl https://www.google.com
curl https://github.com

# Perform DNS lookups
dig google.com
dig github.com
dig stackoverflow.com

```


### Reviewing Traffic (Preview)

**Connection log (conn.log):**

If we simply cat out the contents it might be overwhelming for a first time introduction to a `zeek` log.

```bash  
cat /opt/zeek/logs/current/conn.log
```  

As you can see it's tab-separated with many columns, and depending on your screen resolution and terminal sizing many won't even alight properly.


Let's use `zeek-cut`, a tool specifically designed for parsing Zeek logs. For now let's just look at the last 10 entries, and only look at the 4-tuple (source IP + port, destination IP + port) columns


```bash  
# Show just the source/destination IPs and ports  
zeek-cut id.orig_h id.orig_p id.resp_h id.resp_p < /opt/zeek/logs/current/conn.log | tail -10  
```  


**Output looks like:**

```  
136.177.84.33	42034	67.207.67.3	53
136.177.84.33	48005	67.207.67.3	53
136.177.84.33	49682	67.207.67.3	53
136.177.84.33	56837	67.207.67.3	53
135.237.126.218	59007	136.177.84.33	21
193.46.255.83	40043	136.177.84.33	443
79.124.56.6	50562	136.177.84.33	57864
78.128.114.22	57754	136.177.84.33	16967
196.251.80.30	40478	136.177.84.33	22
79.124.40.118	40797	136.177.84.33	58953
```  

You can see it always involves some connection that includes `136.177.84.33`, which is of course our VM's public IP, which we defined as our local IP in `network.cfg`.



**Understanding conn.log fields:**

The conn.log has many fields. Here are some of the most important ones:

| Field        | Meaning               | Example            |
| ------------ | --------------------- | ------------------ |
| `ts`         | Timestamp             | 1696348338.423     |
| `uid`        | Unique connection ID  | ChFs3N108QUiR6TQI6 |
| `id.orig_h`  | Source IP             | 192.168.1.100      |
| `id.orig_p`  | Source port           | 52847              |
| `id.resp_h`  | Destination IP        | 93.184.216.34      |
| `id.resp_p`  | Destination port      | 443                |
| `proto`      | Protocol              | tcp                |
| `service`    | Identified service    | ssl                |
| `duration`   | Connection duration   | 127.456            |
| `orig_bytes` | Bytes from originator | 4826               |
| `resp_bytes` | Bytes from responder  | 128472             |
| `conn_state` | Connection state      | SF                 |

**Connection states explained:**

The `conn_state` field is particularly important:
```  
┌──────────────────────────────────────────────────────────────┐  
│              CONNECTION STATE FLAGS                          │  
├──────────────────────────────────────────────────────────────┤  
│                                                              │  
│  SF    Normal establishment and termination                  │  
│        (SYN, SYN-ACK, ACK, data, FIN)                        │  
│                                                              │  
│  S0    Connection attempt seen, no reply                     │  
│        (Possible scan or firewall block)                     │  
│                                                              │  
│  S1    Connection established, not terminated                │  
│        (Normal for long connections)                         │  
│                                                              │  
│  REJ   Connection attempt rejected                           │  
│        (Service not available)                               │  
│                                                              │  
│  S2    Connection established, close attempt by originator   │  
│        (No reply from responder)                             │  
│                                                              │  
│  S3    Connection established, close attempt by responder    │  
│        (No reply from originator)                            │  
│                                                              │  
│  RSTO  Connection established, originator aborted (RST)      │  
│                                                              │  
│  RSTR  Responder sent RST                                    │  
│                                                              │  
│  RSTOS0 Originator sent SYN followed by RST                  │  
│         (Never saw SYN-ACK from responder)                   │  
│                                                              │  
│  RSTRH  Responder sent SYN-ACK followed by RST               │  
│         (Never saw SYN from originator)                      │  
│                                                              │  
│  SH     Originator sent SYN followed by FIN                  │  
│         (Half-open connection, no SYN-ACK seen)              │  
│                                                              │  
│  SHR    Responder sent SYN-ACK followed by FIN               │  
│         (Never saw SYN from originator)                      │  
│                                                              │  
│  OTH    No SYN seen, just midstream traffic                  │  
│         (Partial connection, could indicate malformed)       │  
│                                                              │  
└──────────────────────────────────────────────────────────────┘  
```




**Look at DNS queries in dns.log:**

```bash
zeek-cut ts query answers < /opt/zeek/logs/current/dns.log | head -10
```

You'll see domain names being resolved:

```
1696348338.123  www.google.com  172.217.164.164    
1696348339.456  api.github.com  140.82.113.5        
```

If you're on a VM you'll also often see reverse-lookups:

```
1760047200.967744	118.40.124.79.in-addr.arpa	ip-40-118.4vendeta.com
1760047200.999240	24.190.125.185.in-addr.arpa	esm-content-cache-2.ps5.canonical.com
1760047201.008168	110.114.128.78.in-addr.arpa	ip-114-110.superbithost.com
```

Other times we might also see for example **CNAME records** (canonical name / alias records) before finally resolving to an IP address:

```
1760050614.865125	www.cradleoffilth.com	cdn1.wixdns.net,td-ccm-neg-87-45.wixdns.net,34.149.87.45
```

This is just to indicate that Zeek will display the entire resolution chain.

There are of course many other interesting insights to gleam from dns.log, but we'll explore that in its own dedicated lesson in the future, this just served to give a preview.


**Examine HTTP requests:**

```bash
zeek-cut ts host method uri < /opt/zeek/logs/current/http.log | head -10
```

This shows web browsing activity:

```
1696348340.789  www.example.com  GET  /index.html
1696348341.123  cdn.example.com  GET  /assets/style.css
```



### Log Rotation

You might recall in `/opt/zeek/etc/zeekctl.cfg` we had this line: `LogRotationInterval = 3600`.

Essentially what this means is that, every 3600 seconds - i.e. 1 hour - Zeek should start brand new logs. So far we've been looking at logs in `current`, meaning these are the current "live" logs - the ones being written to. But once this hour passes, Zeek will create brand new logs in this folder, and the old ones will be "archived".

Where exactly? Once you've been running for at least one hour, have a look at `/opt/zeek/logs/`


```shell
zeek@zeek-sensor:/usr/local/bin$ ls -lh /opt/zeek/logs/
total 4.0K
drwxr-sr-x 2 root zeek 4.0K Oct  9 19:00 2025-10-09
lrwxrwxrwx 1 root zeek   20 Oct  9 15:22 current -> /opt/zeek/spool/zeek
```

We can see our current directory, but now we also see a directory with the current date, let's see what's inside of that:

```shell
zeek@zeek-sensor:/usr/local/bin$ ls -lh /opt/zeek/logs/2025-10-09
total 568K
-rw-r--r-- 1 root zeek  242 Oct  9 16:00 capture_loss.15:23:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  255 Oct  9 17:00 capture_loss.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  254 Oct  9 18:00 capture_loss.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  255 Oct  9 19:00 capture_loss.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek 1.7K Oct  9 16:00 conn-summary.15:22:37-16:00:00.log.gz
-rw-r--r-- 1 root zeek 1.8K Oct  9 17:00 conn-summary.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek 1.7K Oct  9 18:00 conn-summary.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek 1.7K Oct  9 19:00 conn-summary.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  24K Oct  9 16:00 conn.15:22:37-16:00:00.log.gz
-rw-r--r-- 1 root zeek  42K Oct  9 17:00 conn.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  30K Oct  9 18:00 conn.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  33K Oct  9 19:00 conn.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  507 Oct  9 16:00 dns.15:38:11-16:00:00.log.gz
-rw-r--r-- 1 root zeek 6.2K Oct  9 17:00 dns.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek 4.4K Oct  9 18:00 dns.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek 3.5K Oct  9 19:00 dns.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  491 Oct  9 17:00 files.16:22:17-17:00:00.log.gz
-rw-r--r-- 1 root zeek  447 Oct  9 17:00 http.16:22:17-17:00:00.log.gz
-rw-r--r-- 1 root zeek  518 Oct  9 18:00 http.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  211 Oct  9 17:00 known_services.16:41:56-17:00:00.log.gz
-rw-r--r-- 1 root zeek  345 Oct  9 18:00 ldap_search.17:18:39-18:00:00.log.gz
-rw-r--r-- 1 root zeek 3.6K Oct  9 16:00 loaded_scripts.15:22:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  457 Oct  9 16:00 notice.15:23:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  468 Oct  9 17:00 notice.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  464 Oct  9 18:00 notice.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  468 Oct  9 19:00 notice.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  374 Oct  9 16:00 ntp.15:37:00-16:00:00.log.gz
-rw-r--r-- 1 root zeek  445 Oct  9 17:00 ntp.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  447 Oct  9 18:00 ntp.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  370 Oct  9 19:00 ntp.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  211 Oct  9 16:00 packet_filter.15:22:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  351 Oct  9 19:00 radius.18:13:08-19:00:00.log.gz
-rw-r--r-- 1 root zeek  452 Oct  9 16:00 reporter.15:22:34-16:00:00.log.gz
-rw-r--r-- 1 root zeek  473 Oct  9 16:00 sip.15:49:09-16:00:00.log.gz
-rw-r--r-- 1 root zeek  702 Oct  9 17:00 sip.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  498 Oct  9 18:00 sip.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  546 Oct  9 19:00 sip.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  324 Oct  9 18:00 snmp.17:04:45-18:00:00.log.gz
-rw-r--r-- 1 root zeek  376 Oct  9 19:00 snmp.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek 2.1K Oct  9 16:00 ssh.15:22:53-16:00:00.log.gz
-rw-r--r-- 1 root zeek 3.1K Oct  9 17:00 ssh.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  730 Oct  9 18:00 ssh.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek 1.1K Oct  9 19:00 ssh.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  731 Oct  9 17:00 ssl.16:24:29-17:00:00.log.gz
-rw-r--r-- 1 root zeek  414 Oct  9 18:00 ssl.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  587 Oct  9 16:00 stats.15:22:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  748 Oct  9 17:00 stats.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  714 Oct  9 18:00 stats.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  713 Oct  9 19:00 stats.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek  29K Oct  9 16:00 telemetry.15:22:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek  57K Oct  9 17:00 telemetry.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek  63K Oct  9 18:00 telemetry.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek  66K Oct  9 19:00 telemetry.18:00:00-19:00:00.log.gz
-rw-r--r-- 1 root zeek 2.9K Oct  9 16:00 weird.15:22:24-16:00:00.log.gz
-rw-r--r-- 1 root zeek 6.2K Oct  9 17:00 weird.16:00:00-17:00:00.log.gz
-rw-r--r-- 1 root zeek 1.8K Oct  9 18:00 weird.17:00:00-18:00:00.log.gz
-rw-r--r-- 1 root zeek 2.0K Oct  9 19:00 weird.18:00:00-19:00:00.log.gz
```

We can see here that we likely started our capture at `15:22`, and since then, on the hour, Zeek takes the logs from current and stores them here with a timestamp. Note that note all the logs appear for each hour - it of course depends on the type of traffic that was observed. Finally, also note of course that once the day passes and we reach `2025-10-09`, Zeek will create a new directory and start writing to it.







### **Real-Time Log Monitoring**

To watch logs in real-time:

```bash
# Monitor all new log entries
tail -f /opt/zeek/logs/current/conn.log

# Or use zeek-cut for formatted output
tail -f /opt/zeek/logs/current/conn.log | zeek-cut id.orig_h id.resp_h service
```

Open another terminal and generate traffic with curl commands. Watch the logs update in real-time!


```shell
curl https://www.cradleoffilth.com
curl http://example.com
curl https://www.google.com
curl https://github.com
```











---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./configure.md" >}})
[|NEXT|]({{< ref "./management.md" >}})

