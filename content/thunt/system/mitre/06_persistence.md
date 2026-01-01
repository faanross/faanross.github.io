---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---


## Persistence: Maintaining the Foothold

Adversaries establish persistence to survive system reboots, credential changes, and other disruptions. They modify registry run keys, create scheduled tasks, add entries to startup folders, create new user accounts, install services, or hijack legitimate processes through DLL search order manipulation.

**Threat Hunting Reality**: Excellent hunting opportunities. Persistence mechanisms must modify system state in observable ways, leaving artifacts that generate telemetry we can hunt through. Registry modifications (Sysmon Event ID 12, 13, 14), scheduled task creation (Windows Security Event ID 4698, Sysmon Event ID 20), service installation (Event ID 4697, 7045), and new account creation (Event ID 4720) all provide detection points.

Registry-based persistence is particularly prevalent. Adversaries commonly target autorun registry keys like `HKLM\Software\Microsoft\Windows\CurrentVersion\Run` and `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. We can hunt for registry modifications in these locations, paying special attention to values pointing to unusual file paths - temporary directories, user AppData folders, or executables with suspicious names. Look for registry modifications made by unexpected processes: why is `powershell.exe` spawned from a web browser creating a Run key?

Scheduled tasks provide another rich detection surface. Windows Security Event ID 4698 logs task creation with full details including the execution path, trigger conditions, and the creating account. Hunt for tasks pointing to temporary directories, configured to run with SYSTEM privileges but created by standard users, or with unusual trigger patterns like "every six hours." Examine the parent process that created the task - Office applications spawning `schtasks.exe` is highly suspicious.

Service creation generates clear telemetry through Event IDs 4697 and 7045. Hunt for new services with binary paths outside of `C:\Windows\System32\`, services configured to start automatically but installed from unexpected systems, or services loading DLLs from unusual locations.

Account creation (Event ID 4720) combined with privileged group additions can indicate persistence through credential creation. Look for new local accounts created outside normal provisioning processes, especially those added to Administrators or Remote Desktop Users groups.

**However, the challenge is baseline knowledge**. Legitimate software installs services, creates scheduled tasks, and modifies registry keys constantly. Effective persistence hunting requires understanding your organization's normal software deployment processes and legitimate administrative activities so you can distinguish malicious persistence from expected system changes.






---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./05_execution.md" >}})
[|NEXT|]({{< ref "./07_priv_esc.md" >}})

