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



## **Log Rotation and Management**

As we saw in the previous section, Zeek automatically rotates logs based on your configuration. As we saw there we have current and archived logs.

**Current vs. archived logs:**

```
/opt/zeek/logs/
├── current/              # Active log files being written
│   ├── conn.log
│   ├── dns.log
│   └── ...
└── 2025-10-03/          # Rotated logs from Oct 3
    ├── conn.12:00:00-13:00:00.log.gz
    ├── conn.13:00:00-14:00:00.log.gz
    └── ...
```





### Reading Compressed Logs

As we can see, our logs are compressed (`*.gz`) once they are moved to an archived directory.

Because of this, we can no longer directly cat/read the files:

```bash
zeek@zeek-sensor:/opt/zeek/logs/2025-10-10$ cat tunnel.21:00:00-22:00:00.log.gz
 ��hU��J�0��'O!�F��$��ۭʂ��"���n�MB�����v��p�s8
                                             3ԣm\�����%�C�fpK(6��O�}׋�!t�s�F��&!�ZcO���Ar�F�ϒ"�X�.}�è:P3N��Jv&�ޞ�����:�,B�e4���������s`�
                                                                                                                                        �z�AD�q����Y��i��v��8����DQU��e��2e"-������L�	K�)�/�l6���������ж7��-#�'�_���
```


That's because the data is now binary, we can confirm this by running the output through hexdump:

```bash
zeek@zeek-sensor:/opt/zeek/logs/2025-10-10$ cat tunnel.21:00:00-22:00:00.log.gz | hexdump -C
00000000  1f 8b 08 00 20 ba e9 68  00 03 55 8f dd 4a c4 30  |.... ..h..U..J.0|
00000010  10 85 af 27 4f 21 e4 46  c1 86 24 fd df db ad ca  |...'O!.F..$.....|
00000020  82 a0 e8 22 5e 08 a5 b4  e3 6e a0 4d 42 92 8a eb  |..."^....n.MB...|
00000030  d3 db 76 a5 ac 70 18 be  73 38 0c 33 d4 a3 6d 5c  |..v..p..s8.3..m\|
00000040  13 8c bb fa f8 e6 25 a1  1e 43 bd 66 70 4b 28 0e  |......%..C.fpK(.|
00000050  36 9c ea 4f 85 7d 07 d7  8b b9 21 74 d4 73 ef 1c  |6..O.}....!t.s..|
00000060  46 84 da 26 1c 21 8c 5a  63 4f a8 b1 a8 41 72 99  |F..&.!.ZcO...Ar.|
00000070  46 82 cf 92 22 9a 58 16  84 2e 7d 0f c1 c3 a8 3a  |F...".X...}....:|
00000080  50 1d 33 4e 1d ea e3 4a  76 26 87 de 9e b3 85 ec  |P.3N...Jv&......|
00000090  df de 3a 9c 2c 42 d3 06  65 34 a1 b3 99 16 a9 01  |..:.,B..e4......|
000000a0  c1 07 a7 f4 01 9a ae 73  60 8d 0b 17 84 7a 1c 96  |.......s`....z..|
000000b0  41 44 9e 71 91 a4 a5 cc  59 96 c7 69 92 c3 76 ff  |AD.q....Y..i..v.|
000000c0  f0 38 bc df 7f b5 e3 8f  44 51 55 02 b2 84 65 92  |.8......DQU...e.|
000000d0  89 32 65 22 2d 80 83 88  8b c9 e5 4c c4 09 4b e2  |.2e"-......L..K.|
000000e0  29 d8 2f 87 6c 36 bb e7  15 ab dd eb f6 e9 ed ee  |)./.l6..........|
000000f0  85 d0 b6 37 1e ff bd 2d  23 ce 27 91 5f 9a ea d2  |...7...-#.'._...|
00000100  0c 63 01 00 00                                    |.c...|
00000105
```



OK, but so how can we then actually read the logs?

We have a few options...


#### **Option 1: zcat (easiest)**

bash

```bash
zcat tunnel.21:00:00-22:00:00.log.gz
```

```bash
zeek@zeek-sensor:/opt/zeek/logs/2025-10-10$ zcat tunnel.21:00:00-22:00:00.log.gz
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	tunnel
#open	2025-10-10-21-25-28
#fields	ts	uid	id.orig_h	id.orig_p	id.resp_h	id.resp_p	tunnel_type	action
#types	time	string	addr	port	addr	port	enum	enum
1760145927.673547	CTGLmXFvcuz2e1DD1	64.62.195.158	0	138.197.134.43	0	Tunnel::IP	Tunnel::DISCOVER
#close	2025-10-10-22-00-00
```


#### **Option 2: gunzip with pipe**

```bash
gunzip -c tunnel.21:00:00-22:00:00.log.gz
```

```bash
zeek@zeek-sensor:/opt/zeek/logs/2025-10-10$ gunzip -c tunnel.21:00:00-22:00:00.log.gz
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	tunnel
#open	2025-10-10-21-25-28
#fields	ts	uid	id.orig_h	id.orig_p	id.resp_h	id.resp_p	tunnel_type	action
#types	time	string	addr	port	addr	port	enum	enum
1760145927.673547	CTGLmXFvcuz2e1DD1	64.62.195.158	0	138.197.134.43	0	Tunnel::IP	Tunnel::DISCOVER
#close	2025-10-10-22-00-00
```


#### **Option 3: zless (for viewing)**


```bash
zless tunnel.21:00:00-22:00:00.log.gz
```

```bash
#separator \x09
#set_separator  ,
#empty_field    (empty)
#unset_field    -
#path   tunnel
#open   2025-10-10-21-25-28
#fields ts      uid     id.orig_h       id.orig_p       id.resp_h       id.resp_p       tunnel_type     action
#types  time    string  addr    port    addr    port    enum    enum
1760145927.673547       CTGLmXFvcuz2e1DD1       64.62.195.158   0       138.197.134.43  0       Tunnel::IP      Tunnel::DISCOVER
#close  2025-10-10-22-00-00
tunnel.21:00:00-22:00:00.log.gz (END)
```





## **Backup Strategy**

Since Zeek logs are valuable sources of forensic data and often required for compliance, a solid backup strategy is essential. The key is balancing storage costs, retention requirements, and operational needs.

### **Planning Your Backup Approach**

**Retention periods** vary widely based on organizational needs:

- **Typical ranges**: 30-90 days for routine operations, 6-12 months for compliance-heavy industries
- **Legal/regulatory requirements**: Healthcare (HIPAA), finance (SOX, PCI-DSS), and government sectors often mandate 1-7 years
- **Practical considerations**: Balance retention against storage costs - Zeek can generate 10-100+ GB daily depending on network size

**Storage destinations** to consider:

- **Local backup disks**: Fast, simple, but limited by physical capacity
- **Network storage (NFS/SMB)**: Centralized, easier to manage multiple Zeek instances
- **Cloud storage (S3, Azure Blob)**: Scalable and cost-effective for long-term retention, though retrieval may be slower
- **Tape archives**: Still used in some enterprises for long-term compliance storage

**Timing and disk management**:

- **Schedule backups during low-activity periods** (typically 2-4 AM) to minimize performance impact
- **Keep 15-20% free disk space** on active log partitions to prevent write failures during traffic spikes
- **Monitor backup destination capacity** - ensure it can hold your full retention period plus some buffer

### **Basic Backup Script**

Here's a straightforward backup approach that archives logs after they've aged a bit:

```bash
#!/bin/bash
# backup-zeek-logs.sh

# Configuration
ZEEK_LOGS="/opt/zeek/logs"
BACKUP_DIR="/backup/zeek"
RETENTION_DAYS=90
MIN_FREE_PERCENT=15

# Create backup directory structure
mkdir -p $BACKUP_DIR

# Check available space on backup destination
BACKUP_FREE=$(df -h $BACKUP_DIR | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $BACKUP_FREE -gt $((100 - MIN_FREE_PERCENT)) ]; then
    echo "$(date): WARNING - Backup destination has less than ${MIN_FREE_PERCENT}% free space" >> /var/log/zeek-backup.log
fi

# Sync logs older than 2 days to backup (gives time for log rotation to complete)
find $ZEEK_LOGS -type d -name "20*" -mtime +2 -exec rsync -av {} $BACKUP_DIR/ \;

# Delete backups older than retention period
find $BACKUP_DIR -type d -name "20*" -mtime +$RETENTION_DAYS -exec rm -rf {} \;

# Log completion with basic stats
BACKUP_SIZE=$(du -sh $BACKUP_DIR | awk '{print $1}')
echo "$(date): Backup completed - Total backup size: $BACKUP_SIZE" >> /var/log/zeek-backup.log
```

Save this as `/usr/local/bin/backup-zeek-logs.sh`, make it executable:

```bash
sudo chmod +x /usr/local/bin/backup-zeek-logs.sh

# Add to daily cron (runs at 2 AM)
sudo crontab -e
# Add:
0 2 * * * /usr/local/bin/backup-zeek-logs.sh
```



### **Production Considerations**

For production environments, enhance your backup strategy with:

- **Remote storage integration**: Sync to S3 (`aws s3 sync`), cloud storage, or network shares to protect against local hardware failures
- **Compression**: Use `gzip` or `tar` to reduce storage footprint - Zeek logs compress well (often 10:1 ratios)
- **Encryption**: Encrypt backups in transit and at rest, especially if storing off-site or in cloud
- **Automated monitoring**: Set up alerts for backup failures, capacity thresholds, or unusual log volumes
- **Regular restore testing**: Schedule quarterly restore drills to verify backup integrity - backups are only valuable if you can actually restore them

**Quick tip**: Many organizations keep "hot" logs (recent 7-30 days) on fast local storage for active investigations, then move older logs to cheaper backup storage for compliance retention.













---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./start.md" >}})
[|NEXT|]({{< ref "./exercises.md" >}})

