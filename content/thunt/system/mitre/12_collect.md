---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---


## Collection: Gathering the Prize

Collection covers techniques adversaries use to gather data of interest - staging files, accessing email, capturing screenshots, recording audio/video, or harvesting data from local systems and network shares.

**Threat Hunting Reality**: Moderately huntable, though detection difficulty varies by technique. File system monitoring through Event ID 4663 (object access) and Sysmon Event ID 11 (file creation) reveals staging activities. Email access appears in Exchange logs or Office 365 audit logs. Command line telemetry (Sysmon Event ID 1) captures collection scripts. Network analysis through Zeek's `smb_files.log` and NetFlow shows data aggregation patterns.

Automated collection from network shares generates distinctive patterns. Windows Security Event ID 5145 logs detailed file share access - hunt for high-volume sequential access where users access hundreds or thousands of files in minutes rather than typical dozens per session. Build baselines of normal file access counts per user and timeframe, then identify deviations. Look for sequential alphabetical access, systematic access to specific extensions (`.xlsx`, `.pdf`, `.docx`), or methodical folder hierarchy traversal indicating scripted behavior. Users typically work with consistent file types relevant to their role - a user suddenly accessing hundreds of PDFs or database backups when they normally use Office documents suggests collection activity.

Command line telemetry reveals collection tools through Sysmon Event ID 1. Hunt for PowerShell commands with recursive directory traversal (`Get-ChildItem -Recurse`), file filtering by extension or content (`Select-String -Pattern "password"`), and bulk copy operations (`Copy-Item`, `Robocopy`). Archive creation commands (`Compress-Archive`, `7z.exe`, `rar.exe`) consolidating multiple files indicate adversaries preparing data for exfiltration.

File staging activities appear through Sysmon Event ID 11 showing large numbers of files copied to unusual locations - temporary directories (`C:\Users\Public\`, `C:\Temp\`), web server directories (`C:\inetpub\wwwroot\`), or newly created folders with generic names. Event ID 4663 captures access to sensitive file types across multiple shares by single accounts.

Email collection shows through mailbox access patterns in Exchange logs and Office 365 unified audit logs. Hunt for accounts accessing unusual numbers of mailboxes, searching across multiple folders, or exporting mailbox contents. The `MailItemsAccessed` operation reveals mass access - look for activity from unusual clients, IP addresses, or outside business hours.

Network telemetry reveals collection through SMB traffic analysis. Zeek's `smb_files.log` shows file read operations - hunt for single sources reading massive numbers of files from multiple shares. NetFlow data reveals unusually high inbound traffic to workstations from file servers, indicating bulk downloading that differs from normal access patterns.

**However, the challenge is volume and legitimacy**. File shares exist to be accessed, and normal business involves significant file operations. Effective hunting requires strong baselines of user behavior, understanding job role requirements, and focusing on anomalies - unusual timing, unexpected file types, or massive volume deviations from established patterns.







---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./11_lat.md" >}})
[|NEXT|]({{< ref "./13_c2.md" >}})

