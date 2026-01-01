---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---






## Addendum A: Comparable Overview Table

Here's a comparative overview of all fourteen MITRE ATT&CK tactics from a threat hunter's perspective:

|**Tactic**|**Huntability**|**Primary Challenges**|**Endpoint Telemetry**|**Network Telemetry**|
|---|---|---|---|---|
|**Reconnaissance**|Very Low|Occurs outside our environment; minimal internal telemetry|None - external activity only|None - external activity only|
|**Resource Development**|Very Low|Adversary infrastructure beyond our visibility|None - external activity only|None - external activity only|
|**Initial Access**|Moderate|Often hunting consequences rather than access itself; ephemeral evidence|Process creation from Office apps (Sysmon ID 1); file downloads; execution from temp directories|VPN auth anomalies (Event ID 4624); web server exploitation attempts; connections from unusual geolocations|
|**Execution**|High|High volume of legitimate execution; requires baseline understanding|Process creation (Sysmon ID 1); command line arguments; parent-child relationships; script block logging (Event ID 4104)|Limited - primarily endpoint-focused|
|**Persistence**|High|Distinguishing malicious from legitimate software installations|Registry modifications (Sysmon ID 12-14); scheduled tasks (Event ID 4698); service creation (Event ID 4697); new accounts (Event ID 4720)|Limited - primarily endpoint-focused|
|**Privilege Escalation**|High|Differentiating attacks from legitimate admin operations|Process access to privileged processes (Sysmon ID 10); special privileges assigned (Event ID 4672); token manipulation; exploitation attempts|Limited - primarily endpoint-focused|
|**Defense Evasion**|Moderate-High|Paradoxically creates telemetry while trying to hide; context crucial|Log clearing (Event ID 1102, 104); service stops (Event ID 7036); process injection (Sysmon ID 8); processes with CREATE_SUSPENDED flag|Unusual processes making network connections (notepad.exe, calc.exe)|
|**Credential Access**|Very High|High volume of legitimate authentication; baseline critical|LSASS access (Sysmon ID 10); registry SAM access; failed logins (Event ID 4625); credential dumping tools|Brute force patterns; authentication from unusual sources; geographic anomalies|
|**Discovery**|Moderate|High noise from legitimate admin activity; requires operational understanding|Enumeration commands (net, PowerShell Get-* cmdlets); group/user queries; process enumeration|SMB share enumeration (Event ID 5140, 5145); port scanning; rapid sequential connections|
|**Lateral Movement**|Very High|Legitimate IT operations use same tools/protocols|Remote process creation; WMI activity (wmiprvse.exe children); service creation on remote systems; authentication events (Event ID 4624)|RPC traffic (port 135); RDP connections (port 3389); SMB to admin shares (Zeek smb.log); rapid multi-system connections|
|**Collection**|Moderate|Massive legitimate file access volume; baseline critical|Mass file access (Event ID 5145); staging to temp directories (Sysmon ID 11); archive creation; PowerShell recursive searches|High-volume SMB reads (Zeek smb_files.log); unusual inbound traffic from file servers; mailbox access patterns|
|**Command & Control**|Very High|Encrypted traffic hides content; legitimate cloud services used maliciously|Processes making unusual external connections (Sysmon ID 3); beaconing intervals; unexpected network activity|DNS anomalies (Zeek dns.log); HTTP/S beaconing (Zeek http.log, ssl.log); regular connection intervals; rare destinations; protocol violations|
|**Exfiltration**|High|Encrypted channels; cloud service abuse; distinguishing from legitimate uploads|Processes generating large uploads; unusual applications making external connections|Large upload volumes (NetFlow); cloud storage connections (Zeek ssl.log SNI); sustained outbound traffic; skewed upload/download ratios; off-hours transfers|
|**Impact**|High (but reactive)|Speed of execution; often detected post-impact rather than during|Mass file modifications (Sysmon ID 11); shadow copy deletion (vssadmin); rapid file encryption; ransom note creation; service stops|Pre-encryption C2 beaconing; network scanning before ransomware spread; DDoS traffic patterns|

### Key Insights

**Endpoint-Centric Tactics**: Execution, Persistence, Privilege Escalation, and Defense Evasion are almost entirely dependent on endpoint telemetry. Success here requires robust endpoint logging (Sysmon, Windows Event Logs) and strong baselines of normal process behavior.

**Network-Centric Tactics**: Command & Control is predominantly network-focused, requiring deep packet inspection, protocol analysis (Zeek), and behavioral analytics. Network telemetry provides the primary detection surface.

**Balanced Tactics**: Lateral Movement, Credential Access, Collection, and Exfiltration benefit from both endpoint and network visibility. Correlation between endpoint activity and network behavior provides the strongest detection.

**External Tactics**: Reconnaissance and Resource Development occur outside our environment and offer virtually no hunting opportunities through internal telemetry. These rely on external threat intelligence rather than internal hunting.

**Volume vs. Visibility Trade-off**: The most huntable tactics (Credential Access, Lateral Movement, C2) often come with the highest volume challenges. Effective hunting requires sophisticated baselining, anomaly detection, and contextual analysis to separate signal from noise.



---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./16_conclusion.md" >}})
[|NEXT|]({{< ref "./18_add_b.md" >}})

