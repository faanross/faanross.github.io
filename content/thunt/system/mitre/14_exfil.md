---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---




## Exfiltration: Stealing the Data

Exfiltration covers how adversaries transfer stolen data out of your network. They might use encrypted channels, web services, alternative protocols, DNS tunneling, or less common methods like ICMP or physical media.

**Threat Hunting Reality**: Highly huntable through network monitoring, though encrypted traffic and legitimate cloud services provide challenges. Zeek's protocol logs, NetFlow data, firewall logs, and endpoint telemetry (Sysmon Event ID 3 for network connections) provide detection opportunities. The key is identifying volume, timing, and behavioral anomalies.

Cloud storage exfiltration is increasingly common because services like Dropbox, Google Drive, and OneDrive blend with normal business traffic. Zeek's `http.log` and `ssl.log` capture connections to cloud storage domains - hunt for unusual upload volumes by analyzing connection sizes and durations. Baseline normal upload patterns per user and system, then identify deviations: sudden large uploads (multi-gigabyte transfers), sustained upload sessions outside business hours, or systems that never used cloud storage suddenly transferring data. Look for connections to unauthorized services - if your organization doesn't use Dropbox, any connection to Dropbox infrastructure is suspicious. Zeek's `ssl.log` can identify cloud storage providers through SNI (Server Name Indication) fields even when traffic is encrypted.

Protocol-based exfiltration uses various channels. DNS exfiltration appears in Zeek's `dns.log` through high query volumes, large TXT record responses, or unusual query patterns to suspicious domains. HTTP exfiltration shows in `http.log` through large POST requests, unusual upload endpoints, or connections with skewed upload/download ratios. Hunt for long-duration connections with continuous outbound data flow - normal web browsing has mixed bidirectional traffic, while exfiltration shows sustained uploads.

Alternative protocol abuse reveals itself through unexpected protocol usage. Zeek's `ftp.log`, `smtp.log`, and custom protocol parsers can detect exfiltration via FTP, email attachments, or unusual protocols. Hunt for FTP uploads to external servers from systems that don't normally use FTP, SMTP traffic from non-mail servers, or ICMP packets with unusual payload sizes (data hidden in ping packets). NetFlow analysis shows traffic to uncommon ports or protocols that stand out from baseline patterns.

Volume analysis across all protocols is critical. Identify top talkers - systems generating unusually high outbound traffic volumes. Calculate bytes-out ratios; normal systems have relatively balanced traffic, while exfiltration shows heavy outbound bias. Look for gradual sustained transfers over time (low-and-slow exfiltration) or sudden large bursts that spike above normal patterns.

Timing patterns help distinguish exfiltration from legitimate activity. Hunt for large outbound transfers during off-hours when networks are quiet and transfers are less likely to be noticed. Weekend or holiday exfiltration attempts stand out clearly against minimal legitimate business activity. Zeek's `conn.log` timestamps enable temporal analysis of connection patterns.

Destination analysis reveals suspicious targets. Hunt for connections to newly registered domains, hosting providers popular with adversaries, or countries where your organization doesn't operate. Rare destinations - external IPs contacted by only one or two internal systems - warrant investigation. Correlation with threat intelligence feeds identifying known malicious infrastructure enhances detection.

Endpoint correlation provides context. Sysmon Event ID 3 (network connection) shows which processes generate outbound traffic. Hunt for unusual processes making large transfers - `powershell.exe`, `cmd.exe`, or custom tools rather than legitimate browsers or sync clients. File creation events (Sysmon Event ID 11) showing recent staging activity followed by network transfers to external destinations confirms the attack chain from collection through exfiltration.

Encrypted traffic analysis requires metadata focus since payloads are hidden. Analyze packet sizes, inter-arrival times, connection durations, and TLS certificate characteristics. Hunt for self-signed certificates, unusual cipher suites, or TLS connections to suspicious domains. Traffic fingerprinting can sometimes identify exfiltration tools even within encrypted channels based on behavioral signatures.


**However, the challenge is volume and encryption**. Legitimate business operations generate massive data transfers daily - backups, file synchronization, cloud application usage, video conferencing. Distinguishing malicious exfiltration from normal high-volume transfers requires strong baselines of typical data flows per user, system, and time period.

Encryption hides the _content_ of exfiltration but doesn't prevent detection of the exfiltration itself. We can't see what's being stolen inside encrypted channels, but we can identify strong indicators that exfiltration is occurring through volume anomalies, timing patterns, destination analysis, and behavioral deviations.





---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./13_c2.md" >}})
[|NEXT|]({{< ref "./15_impact.md" >}})

