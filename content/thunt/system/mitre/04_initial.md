---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---




## Initial Access: Breaching the Perimeter

Initial access is where adversaries first get their foothold in your environment - through phishing emails, exploiting public-facing applications, compromising supply chains, or leveraging valid accounts.

**Threat Hunting Reality**: Moderately huntable, though often we're hunting for the _consequences_ of initial access rather than the access itself. Email gateway logs, web server logs, VPN authentication logs (Event ID 4624 with logon type 10), and endpoint telemetry provide detection opportunities, but the initial breach moment can be ephemeral.

Phishing-related initial access reveals itself through execution patterns following email receipt. Hunt for Office applications spawning unusual child processes - `OUTLOOK.EXE` or `WINWORD.EXE` launching `powershell.exe`, `cmd.exe`, or `wscript.exe` shortly after receiving external emails (Sysmon Event ID 1). Look for file downloads to temporary locations followed by execution, or macro-enabled documents opened from email attachments that immediately generate network connections (Sysmon Event ID 3) to external IPs.

Exploitation of public-facing applications appears in web server logs as unusual request patterns - excessive 404 errors suggesting path traversal attempts, requests with SQL injection patterns, or POST requests to administrative endpoints from unexpected sources. IIS logs, Apache access logs, and application logs capture failed authentication attempts (Event ID 4625) against web applications, especially when followed by successful authentication. Hunt for web server processes spawning unexpected child processes like `cmd.exe` or `powershell.exe` - indicating successful exploitation leading to code execution.

Valid account compromise shows through authentication anomalies. Hunt for VPN logins from unusual geolocations, impossible travel scenarios (logins from distant locations within impossible timeframes), or authentication from previously unseen IP addresses. Event ID 4624 with unusual logon hours, multiple failed attempts (Event ID 4625) followed by success, or accounts authenticating from both internal and external sources simultaneously warrant investigation.

External remote services abuse appears through RDP (port 3389) or SSH (port 22) connections from external IPs in firewall logs or NetFlow data. Look for successful authentications to internet-facing systems, especially outside business hours or from countries where your organization doesn't operate. Zeek's `ssh.log` and `rdp.log` can reveal brute force patterns or successful authentications from suspicious sources.

Supply chain compromise is harder to hunt proactively but may reveal itself through unexpected software behavior. Hunt for trusted applications making unusual network connections, digitally signed binaries executing from unexpected locations, or legitimate software update processes fetching from unfamiliar domains.

**However, the challenge is distinguishing initial access from normal operations**. Users click links, access web applications, and authenticate from various locations legitimately. Effective hunting requires understanding normal email patterns, typical authentication sources, and expected web application usage to spot genuine breaches among the noise.



---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./03_resource.md" >}})
[|NEXT|]({{< ref "./05_execution.md" >}})

