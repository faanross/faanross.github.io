---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---






## Addendum B: Threat Hunting Telemetry Reference Matrix


|**Tactic**|**Sysmon Events**|**Windows Event Logs**|**Other Endpoint**|**Zeek Logs**|**NetFlow/IPFIX**|**Other Network**|
|---|---|---|---|---|---|---|
|**Initial Access**|ID 1 (Process creation from Office apps)<br>ID 3 (Network connections post-email)<br>ID 11 (File downloads)|ID 4624 (VPN/remote logons)<br>ID 4625 (Failed auth attempts)|Web server logs (IIS, Apache)<br>Email gateway logs|ssh.log (SSH attempts)<br>rdp.log (RDP attempts)<br>http.log (exploitation)|Connections from unusual geolocations<br>Inbound to public-facing services|Firewall logs<br>IDS/IPS alerts|
|**Execution**|ID 1 (Process creation, command lines, parent-child)<br>ID 11 (Script file creation)|ID 4688 (Process creation)<br>ID 4104 (PowerShell script block)<br>ID 4103 (PowerShell module logging)|Application logs<br>Script execution logs|-|-|-|
|**Persistence**|ID 1 (Process creating persistence)<br>ID 11 (File creation in startup)<br>ID 12, 13, 14 (Registry modifications)<br>ID 20 (Scheduled tasks)|ID 4698 (Scheduled task creation)<br>ID 4697, 7045 (Service installation)<br>ID 4720 (Account creation)<br>ID 4732, 4728 (Group additions)|Application installation logs<br>Startup folder monitoring|-|-|-|
|**Privilege Escalation**|ID 10 (Process access to lsass, winlogon)<br>ID 1 (Exploitation attempts)|ID 4672 (Special privileges assigned)<br>ID 4624 (Logon with new privileges)<br>ID 1000, 1001 (Application crashes)|Memory analysis<br>Crash dumps<br>Unsigned driver loads|-|-|-|
|**Defense Evasion**|ID 1 (CREATE_SUSPENDED processes)<br>ID 8 (CreateRemoteThread)<br>ID 12, 13 (Defender registry changes)<br>ID 3 (Unusual process connections)|ID 1102, 104 (Log clearing)<br>ID 7036, 7040 (Service stop/start)|Security tool logs<br>AV logs<br>Process memory analysis|-|-|Unusual processes making external connections|
|**Credential Access**|ID 10 (LSASS process access)<br>ID 1 (Credential dumping tools)<br>ID 11 (Credential file staging)<br>ID 12 (Registry SAM access)|ID 4625 (Failed logons)<br>ID 4624 (Successful logons)<br>ID 4648 (Explicit credentials)<br>ID 4663 (Credential file access)|Screenshot capture<br>Keylogger detection<br>Clipboard monitoring|kerberos.log (Kerberos auth)<br>ntlm.log (NTLM auth)|Authentication traffic patterns<br>Geographic anomalies|RADIUS logs<br>LDAP query logs|
|**Discovery**|ID 1 (Enumeration commands: net, PowerShell Get-*)<br>ID 22 (DNS queries)|ID 4798, 4799 (Group enumeration)<br>ID 5140, 5145 (Share access)|-|dns.log (DNS enumeration)<br>smb_mapping.log (Share enumeration)|Port scanning patterns<br>Rapid multi-host connections|-|
|**Lateral Movement**|ID 1 (wmic, PsExec commands, remote processes)<br>ID 3 (WMI/RDP network connections)<br>ID 19, 20, 21 (WMI events)|ID 4624 (Network logons Type 3, 10)<br>ID 4648 (Explicit credentials)<br>ID 4778, 4779 (RDP sessions)<br>ID 4697, 7045 (Remote service creation)<br>ID 5140, 5145 (Admin share access)|Process created by wmiprvse.exe<br>Remote execution artifacts|dce_rpc.log (WMI/RPC)<br>rdp.log (RDP sessions)<br>smb_files.log (File writes to shares)<br>smb_mapping.log (Share mappings)|Port 135 (RPC) connections<br>Port 3389 (RDP) connections<br>Port 445 (SMB) traffic<br>Workstation-to-workstation patterns|Firewall logs showing lateral patterns|
|**Collection**|ID 1 (Collection scripts, archive tools)<br>ID 11 (File staging, compression)<br>ID 3 (Connections from collection tools)|ID 4663 (Mass file access)<br>ID 5145 (Detailed share access)|Office 365 audit logs (MailItemsAccessed)<br>Exchange logs (mailbox access)|smb_files.log (Mass file reads)<br>http.log (Webmail access)|High inbound traffic from file servers to workstations|Email server logs|
|**Command & Control**|ID 1 (C2 tool execution)<br>ID 3 (C2 connections, beaconing)<br>ID 22 (DNS C2 queries)|-|Process network behavior<br>Beaconing detection|dns.log (DNS tunneling, high entropy)<br>http.log (Beaconing, unusual URIs)<br>ssl.log (Suspicious certificates, SNI)<br>conn.log (Connection patterns, durations)|Regular connection intervals<br>Rare destinations<br>Long-duration connections<br>Traffic to unusual ports|Proxy logs<br>TLS inspection<br>IDS/IPS signatures|
|**Exfiltration**|ID 3 (Large upload connections)<br>ID 1 (Exfil tool processes)<br>ID 11 (Data staging before exfil)|-|Cloud sync client logs<br>Data transfer logs|ssl.log (Cloud storage SNI)<br>http.log (Large POSTs)<br>dns.log (DNS exfil)<br>ftp.log (FTP uploads)<br>smtp.log (Email exfil)<br>conn.log (Upload patterns)|High outbound volumes<br>Sustained uploads<br>Skewed upload/download ratios<br>Off-hours transfers|DLP alerts<br>Proxy logs<br>Firewall bandwidth monitoring|
|**Impact**|ID 1 (Ransomware, wipers, vssadmin)<br>ID 11 (Mass file modifications, ransom notes)<br>ID 23 (Mass file deletions)<br>ID 3 (Pre-encryption C2)|ID 524 (Backup operations)<br>ID 7036 (Service stops)<br>ID 1102 (Log clearing)|File integrity monitoring<br>Backup system logs<br>Database logs|conn.log (Pre-encryption beaconing)<br>http.log (C2 before impact)|Network scanning before encryption<br>Unusual connection spikes|Web server logs (defacement)<br>Application performance monitoring|

### Legend & Usage Notes

**Sysmon Event IDs:**

- **ID 1**: Process creation (captures executables, command lines, parent-child relationships)
- **ID 3**: Network connection (shows which processes connect where)
- **ID 8**: CreateRemoteThread (process injection technique)
- **ID 10**: Process access (critical for detecting LSASS access)
- **ID 11**: File creation (tracks file writes, staging)
- **ID 12, 13, 14**: Registry events (create, set value, rename)
- **ID 19, 20, 21**: WMI events (consumer, filter, binding)
- **ID 20**: WmiEvent (scheduled tasks via WMI)
- **ID 22**: DNS query
- **ID 23**: File delete

**Windows Event Log IDs:**

- **4624**: Successful logon (various types: 2=interactive, 3=network, 10=RDP)
- **4625**: Failed logon attempt
- **4648**: Logon using explicit credentials
- **4663**: Object access (file/registry)
- **4672**: Special privileges assigned to new logon
- **4688**: Process creation (alternative to Sysmon ID 1)
- **4697/7045**: Service installation
- **4698**: Scheduled task creation
- **4720**: User account creation
- **4732/4728**: User added to privileged group
- **4778/4779**: RDP session connect/disconnect
- **4798/4799**: Group enumeration
- **5140/5145**: Network share access
- **1102/104**: Security/System log cleared

**Empty cells (-) indicate limited or no relevant telemetry for that data source and tactic combination.**




---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./17_add_a.md" >}})


