---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---


## Credential Access: Stealing the Keys to the Kingdom

Credential access represents adversaries' efforts to steal usernames, passwords, hashes, tokens, and keys. They dump credentials from memory, capture keystrokes, brute force authentication, or search files for stored credentials.

**Threat Hunting Reality**: Extremely huntable and critically important. Credential theft is often the gateway to further compromise. Sysmon Event ID 10 (process access) captures attempts to access sensitive processes like `lsass.exe`, while Windows Security Event IDs 4625 (failed logon) and 4648 (explicit credential use) reveal authentication abuse patterns. File access monitoring through Event ID 4663 detects searches for credential files.

LSASS memory dumping is one of the most common techniques. Hunt by monitoring Sysmon Event ID 10 for process access to `lsass.exe` with specific permissions - particularly `PROCESS_VM_READ` combined with `PROCESS_QUERY_INFORMATION`. Focus on unusual processes accessing LSASS: legitimate Windows components have predictable names and paths (`C:\Windows\System32\`), while a process named `svchost.exe` running from `C:\Users\Public\` indicates masquerading. Look for known dumping utilities like `procdump.exe`, or `rundll32.exe`loading `comsvcs.dll` with MiniDump parameters.

Brute force attacks reveal themselves through authentication logs. Hunt for Event ID 4625 (failed logons) with high frequency against single accounts, especially with varying usernames against a single system (password spraying) or single usernames with many failures (credential stuffing). Look for failed attempts followed by successful logon (Event ID 4624) from the same source - indicating the attack succeeded. Network telemetry through Zeek's `kerberos.log` or `ntlm.log` can reveal authentication attempts from unusual sources or geographic locations.

Credential dumping from registry hives generates Sysmon Event ID 12 (registry object access) when adversaries access `HKLM\SAM`, `HKLM\SECURITY`, or `HKLM\SYSTEM` to extract password hashes. Commands like `reg save HKLM\SAM` create process execution telemetry (Sysmon Event ID 1) that's highly suspicious from non-administrative contexts.

File searches for credentials show up through command line telemetry. Hunt for PowerShell or `cmd.exe` executing searches with keywords like "password", "pwd", "credential" in filenames, or accessing files like `web.config`, `unattend.xml`, or `.kdbx` (KeePass databases). Sysmon Event ID 11 (file creation) can detect when found credentials are staged to temporary locations.

Keylogging and input capture are harder to detect without endpoint agents but may reveal themselves through suspicious process behavior. Hunt for processes loading keyboard hook libraries, accessing raw input devices, or creating files with keystroke-like patterns. Processes monitoring clipboard activity or taking frequent screenshots (rapid `.png` or `.jpg`creation via Sysmon Event ID 11) also indicate credential capture attempts.


**However, the challenge is volume and legitimacy**. LSASS is accessed by many Windows components, and failed logins occur regularly from forgotten passwords. Effective hunting requires filtering known-good processes, establishing authentication baselines, and focusing on context - unusual timing, unexpected sources, or correlation with other suspicious activities like lateral movement or privilege escalation following credential access.



---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./08_evade.md" >}})
[|NEXT|]({{< ref "./10_discovery.md" >}})

