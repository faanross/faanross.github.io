---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---



## Execution: Running the Payload

Once inside, adversaries need to execute code. This might be through command and scripting interpreters, user execution of malicious files, scheduled tasks, or exploitation for client execution.

**Threat Hunting Reality**: Highly huntable. Execution generates substantial telemetry across multiple data sources that threat hunters can leverage. Process creation events (Sysmon Event ID 1) provide visibility into what's executing on endpoints. These logs capture not just the process name, but critical context like full command line arguments, parent-child process relationships, user accounts, execution paths, and even hashes.

This granularity enables powerful hunting opportunities. We can detect unusual parent-child relationships, such as Microsoft Word (`WINWORD.EXE`) spawning PowerShell (`powershell.exe`) - a classic indicator of macro-based malware execution.

Script execution provides another rich hunting ground. PowerShell script block logging (Event ID 4104) captures the actual commands being executed, allowing us to hunt for suspicious patterns like base64-encoded commands, download cradles using `Invoke-WebRequest` or `Net.WebClient`, or attempts to bypass execution policy.

Command line arguments deserve special attention. Execution techniques often reveal themselves through their parameters: PowerShell with `-EncodedCommand`, `-ExecutionPolicy Bypass`, or `-WindowStyle Hidden` flags; legitimate system utilities like `regsvr32.exe` or `rundll32.exe` being called with unusual DLL paths or export functions; or scheduled tasks created via `schtasks.exe` with suspicious action paths or trigger times.

Modern endpoint detection platforms enhance this further by providing behavioural context - memory analysis, loaded modules, network connections initiated by processes, and file system modifications. This allows us to not just see that execution occurred, but understand the full context of what the executed code is doing within our environment.

**However, the challenge is volume**. Legitimate execution happens constantly. Effective hunting requires understanding normal execution patterns in your environment so you can spot anomalies.






---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./04_initial.md" >}})
[|NEXT|]({{< ref "./06_persistence.md" >}})

