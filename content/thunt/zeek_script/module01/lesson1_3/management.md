---
showTableOfContents: true
title: "Part 7 - Basic Operations + Management"
type: "page"
---



## **Essential ZeekControl Commands**

Let's practice the basic operational commands you'll use daily:

**Check Zeek status:**

```bash
sudo zeekctl status
```

```bash
Name         Type       Host          Status    Pid    Started
zeek         standalone localhost     running   51924  09 Oct 15:22:23
```


This shows if Zeek is running and provides the process ID and start time.

**Stop Zeek:**

```bash
sudo zeekctl stop
```

Zeek will gracefully shut down, closing any open log files cleanly.

**Start Zeek:**

```bash
sudo zeekctl start
```

**Restart Zeek (stop then start):**

```bash
sudo zeekctl restart
```

**Deploy new configuration:**

When you modify configuration files, deploy installs the new configuration and restarts Zeek:

```bash
sudo zeekctl deploy
```

**Check configuration without deploying:**

```bash
sudo zeekctl check
```

```bash
sudo zeekctl check
zeek scripts are ok.
```


This validates your configuration files without actually deploying changes. Use this to catch errors before restarting Zeek.

## **Monitoring Zeek Health**

Zeek generates its own performance and health statistics. Let's examine them:

### Checking for dropped packets

**Check for packet drops:**

```bash
# View capture_loss.log
cat /opt/zeek/logs/current/capture_loss.log
```


```bash
zeek@zeek-sensor:/usr/local/bin$ cat /opt/zeek/logs/current/capture_loss.log
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	capture_loss
#open	2025-10-09-19-19-31
#fields	ts	ts_delta	peer	gaps	acks	percent_lost
#types	time	interval	string	count	count	double
1760051971.131300	60.001036	zeek	0	0	0.0
```

**Note the last 3 columns:**
- `gaps`: 0 (number of packet gaps detected)
- `acks`: 0 (acknowledgments - related to packet capture)
- `percent_lost`: **0.0** ← This is the key metric!

This of course means our Zeek instance is handling **100% of the traffic** without dropping any packets. This is ideal and means:

- Zeek is keeping up with the traffic volume
- You have sufficient CPU/memory resources
- Your network interface isn't being overwhelmed



### **Check Zeek's resource usage**

```bash
# View stats.log
zeek-cut ts mem proc < /opt/zeek/logs/current/stats.log | tail
```

This shows memory usage and process statistics over time. Watch for:

- Memory usage trending upward (potential leak)
- Event queue depth increasing (scripts can't keep up)

In our case below we can see memory usage was holding stable:

```
1760040144.969685	248
1760040444.969934	248
1760040744.970859	248
1760041044.971061	248
1760041344.976762	248
1760041644.977121	249
1760041944.977447	249
1760042244.978320	249
1760042544.978550	249
1760042844.979198	249
1760043144.980137	249
1760043444.981138	249
```


### **Check for weird/unusual events:**

```bash
# View weird.log - protocol violations and unusual behavior
cat /opt/zeek/logs/current/weird.log
```

Zeek logs "weird" events when it sees something that doesn't conform to protocol specifications. A few weirds are normal, but many could indicate:

- Network scanning
- Malformed attack traffic
- Misconfigured applications


Here are a few examples - again we'll cover weird.log in its own lesson:

```bash
zeek-cut name < /opt/zeek/logs/current/weird.log | sort | uniq -c | sort -rn
      5 bad_TCP_checksum
      4 active_connection_reuse
      3 inappropriate_FIN
      3 data_before_established
      2 above_hole_data_without_any_acks
      1 truncated_tcp_payload
      1 bad_UDP_checksum
```


- **`bad_TCP_checksum`**: The packet's integrity check failed, meaning the data may be corrupted.
- **`active_connection_reuse`**: A new connection attempt (SYN packet) was seen for a connection that Zeek already considers to be active. This can sometimes be normal with aggressive NAT devices but could also indicate certain types of network scans.
- **`inappropriate_FIN`**: A TCP packet meant to close a connection (a FIN packet) was received at a time when the connection wasn't fully established or was in an unexpected state. This is often seen during TCP port scanning where the scanner abruptly terminates the connection attempt.
- **`data_before_established`**: Data was sent from a client before the TCP three-way handshake was fully completed. This is highly unusual for normal traffic.
- **`above_hole_data_without_any_acks`**: The system received a chunk of data that is far ahead of where it should be in the sequence, implying a large amount of data in between was lost. This indicates significant packet loss or reordering on the network.
- **`truncated_tcp_payload`**: The packet header claimed the payload was a certain size, but the actual data received was smaller. This suggests the packet was cut off or damaged somewhere in transit.
- **`bad_UDP_checksum`**: Similar to its TCP counterpart, this indicates the integrity check for a UDP packet failed. While checksums are optional for UDP on IPv4, if a checksum is present and it's wrong, it points to data corruption.














---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./start.md" >}})
[|NEXT|]({{< ref "./exercises.md" >}})

