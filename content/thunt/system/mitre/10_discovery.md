---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---




## Discovery: Mapping the Environment

Once inside, adversaries need to understand their environment. They enumerate users, groups, systems, network shares, running processes, and security software. This reconnaissance helps them plan their next moves.

**Threat Hunting Reality**: Huntable with moderate difficulty. Discovery activities generate numerous events, but distinguishing malicious from legitimate enumeration requires understanding normal administrative behavior. Sysmon Event ID 1 (process creation) captures enumeration commands, while Windows Security Event IDs 5140 (network share access) and 5145 (shared object access check) reveal share enumeration. Event ID 4798 (user's local group membership enumerated) and 4799 (security-enabled local group membership enumerated) detect group discovery.

Network share enumeration is common during discovery phases. Hunt for `net view`, `net share`, or PowerShell's `Get-SmbShare` in command line logs (Sysmon Event ID 1). Look for rapid sequential access patterns in Event IDs 5140 and 5145 - an account accessing dozens of shares in minutes suggests automated enumeration rather than normal user behaviour. Pay special attention to enumeration of administrative shares (`C$`, `ADMIN$`, `IPC$`) across multiple systems, especially from workstations.

Account and group enumeration reveals itself through commands like `net user`, `net group`, `net localgroup administrators`, or PowerShell's `Get-ADUser` and `Get-LocalGroupMember`. These generate process creation events with distinctive command lines. Hunt for enumeration executed from unusual sources - why is a standard workstation running domain-wide user queries?

System and network discovery shows up through commands like `ipconfig`, `arp -a`, `nslookup`, `nltest /domain_trusts`, or `Get-NetIPConfiguration`. While individually common, rapid sequential execution of multiple discovery commands suggests reconnaissance. Look for `ping` sweeps across IP ranges, `nbtstat` queries against multiple hosts, or systematic DNS lookups.

Process and service enumeration appears through `tasklist`, `sc query`, `Get-Process`, or `Get-Service` commands. Hunt for these executed remotely via WMI or PsExec, or executed locally following other suspicious activities like initial access or credential theft.

Security software discovery involves queries for antivirus processes, firewall status, or Windows Defender configuration. Commands like `Get-MpComputerStatus`, `netsh advfirewall show`, or `tasklist` filtered for security product names indicate adversaries mapping defenses before acting.

**However, the challenge is baseline noise**. Administrators legitimately enumerate systems, shares, and users regularly. Effective hunting requires understanding your IT operations - knowing when, where, and how legitimate discovery occurs so you can spot adversaries performing the same actions from wrong contexts, wrong times, or wrong sources.






---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./09_cred.md" >}})
[|NEXT|]({{< ref "./11_lat.md" >}})

