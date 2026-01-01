---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---






## Impact: The Final Act

Impact represents adversaries' final objectives - destroying data, encrypting files for ransom, manipulating data, disrupting services, or defacing systems.

**Threat Hunting Reality**: Highly visible but often reactive rather than proactive. Impact activities are typically detected through their effects, though we can hunt for precursor activities or early stages before full impact occurs. File system monitoring (Sysmon Event ID 11, 23), process creation (Sysmon Event ID 1), command line telemetry, and network connections (Sysmon Event ID 3) provide hunting opportunities for detecting impact before it completes.

Ransomware detection focuses on identifying encryption before it spreads throughout your environment. Hunt for processes accessing large numbers of files in rapid succession through file system monitoring - Sysmon Event ID 11 (file creation) showing hundreds of file modifications in minutes. Look for file extension changes as ransomware systematically renames files with unusual extensions (`.encrypted`, `.locked`, random strings). Mass file creation events with identical filenames across multiple directories - particularly ransom notes like `READ_ME.txt`, `HOW_TO_DECRYPT.html`, or `RECOVERY_INSTRUCTIONS.txt` - are clear indicators.

Command line telemetry reveals destructive intent. Hunt for Volume Shadow Copy deletion through Sysmon Event ID 1 capturing commands like `vssadmin delete shadows /all`, `wmic shadowcopy delete`, or PowerShell's `Get-WmiObject Win32_ShadowCopy | Remove-WmiObject`. Backup deletion commands targeting `wbadmin delete catalog`, database backups, or network backup repositories indicate adversaries preventing recovery. Event ID 524 (backup operation attempted) combined with service stop events (Event ID 7036) for backup services reveals backup tampering.

Process behavior provides early warning. Ransomware often executes from unusual locations - temporary directories, user AppData folders, or recently created paths. Hunt for processes with suspicious names (random strings, attempts to mimic system processes) accessing file system resources rapidly. Sysmon Event ID 1 captures parent-child relationships showing ransomware spawned by phishing documents, malicious scripts, or lateral movement tools.

Network activity precedes encryption in many ransomware variants. Zeek's `conn.log` and `http.log` capture pre-encryption C2 communication as ransomware beacons for encryption keys or exfiltrates data before encrypting (double extortion). Hunt for unusual external connections immediately before file system activity spikes. Some ransomware performs network scanning (visible in NetFlow or Zeek's `conn.log`) to identify additional targets before encryption.

Data destruction beyond ransomware includes wipers and defacement. Hunt for mass file deletion events (Sysmon Event ID 23), especially targeting critical system files, databases, or configuration files. MBR (Master Boot Record) or disk wiping shows through low-level disk access by unusual processes. Web defacement appears through unexpected modifications to web server directories - Sysmon Event ID 11 showing file creation or modification in `C:\inetpub\wwwroot\` or `/var/www/html/` by unauthorized processes.

Service disruption reveals itself through service stop events (Event ID 7036, 7040) for critical services, process termination of business-critical applications, or resource exhaustion. Hunt for commands stopping multiple services rapidly, particularly targeting security tools, databases, or email servers. DDoS from compromised internal systems shows through NetFlow as unusual outbound connection volumes to single destinations.

Data manipulation is subtler but detectable through integrity monitoring. Hunt for unexpected modifications to critical files, database records, or configuration files outside normal maintenance windows. Changes to financial records, customer data, or system configurations by unauthorized accounts or processes warrant investigation.

Precursor detection is crucial since impact rarely occurs in isolation. Hunt for the attack chain leading to impact: initial access followed by credential theft, lateral movement to critical systems, and collection or staging activities. Detecting these earlier stages prevents impact from occurring. Systems showing signs of reconnaissance, persistence establishment, or privilege escalation should trigger heightened monitoring for potential impact activities.

**However, the challenge is speed**. Impact activities often execute rapidly, leaving limited detection windows. Effective hunting requires automated alerting on high-risk indicators (shadow copy deletion, mass file encryption), strong baselines to quickly identify anomalies, and incident response readiness to contain impact before it spreads across your environment.




---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./14_exfil.md" >}})
[|NEXT|]({{< ref "./16_conclusion.md" >}})

