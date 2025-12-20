---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---



## Defense Evasion: Staying Under the Radar

Adversaries employ numerous techniques to avoid detection - disabling security tools, obfuscating code, masquerading as legitimate processes, clearing logs, and hiding artifacts.

**Threat Hunting Reality**: Paradoxically huntable. While the goal of defense evasion is to hide, the techniques used to evade detection often create their own telemetry. Disabling a security tool generates service stop events (Event ID 7036, 7040). Log clearing generates Event ID 1102 (Security log cleared) or 104 (System log cleared). Process injection and hollowing create observable anomalies in process behavior.

Process hollowing and injection techniques reveal themselves through unusual process characteristics. Sysmon Event ID 1 (process creation) captures processes created with the `CREATE_SUSPENDED` flag in the command line - while some legitimate software uses suspended creation, it's uncommon. Hunt for processes created this way, especially by unusual parents like Office applications or browsers. Sysmon Event ID 8 (CreateRemoteThread) detects when one process creates a thread in another process's memory space, a common injection technique.

Behavioral anomalies often expose hollowed processes. Network connection logs (Sysmon Event ID 3) showing `notepad.exe` or `calc.exe` making external connections are immediate red flags - these processes have no legitimate reason to communicate over the network. Look for signed system binaries executing from unusual paths or with unexpected parent processes.

Security tool tampering generates clear signals. Service stop/start events (Event IDs 7036, 7040) for antivirus or monitoring tools, especially initiated by non-administrative processes or at unusual times, warrant investigation. Hunt for processes modifying Windows Defender settings through registry changes (Sysmon Event IDs 12, 13) to `HKLM\SOFTWARE\Microsoft\Windows Defender\` or executing commands like `Set-MpPreference -DisableRealtimeMonitoring $true`.

Log clearing is self-documenting through Event ID 1102 and 104. While legitimate administrators clear logs, context matters - clearing logs outside maintenance windows, from workstations rather than administrative systems, or immediately following other suspicious activities suggests evasion.

File and registry masquerading shows up through naming anomalies. Hunt for processes with names similar to legitimate system processes but with slight variations (`svch0st.exe` instead of `svchost.exe`), or legitimate process names executing from wrong locations (`svchost.exe` running from `C:\Users\` instead of `C:\Windows\System32\`).

**However, the challenge is context**. Many evasion techniques use legitimate Windows functionality. Effective hunting requires correlating multiple indicators and understanding timing - evasion techniques rarely occur in isolation but as part of a broader attack chain.




---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./07_priv_esc.md" >}})
[|NEXT|]({{< ref "./09_cred.md" >}})

