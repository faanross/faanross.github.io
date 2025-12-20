---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---




## Privilege Escalation: Climbing the Access Ladder

Adversaries seek higher-level permissions to access protected resources, execute with elevated rights, and expand their control. They exploit vulnerabilities, abuse misconfigurations, steal privileged credentials, or leverage access tokens.

**Threat Hunting Reality**: Highly huntable with strong endpoint focus. Privilege escalation attempts generate distinctive telemetry through Windows Security. Event ID 4672 (special privileges assigned to new logon) combined with Event ID 4624 (successful logon) can reveal token manipulation when the logon type and privileges don't align with expected behaviour. Sysmon Event ID 10 (process access) captures attempts to access privileged processes like `lsass.exe` or `winlogon.exe`, which often precede privilege escalation.

Exploitation attempts generate crash dumps, application errors (Event ID 1000, 1001), and unusual memory access patterns. Hunt for processes with unexpected memory permissions (RWX regions), services or applications crashing repeatedly on specific systems, or processes attempting to load unsigned drivers or kernel modules.

Abuse of privileged credentials reveals itself through authentication patterns. Look for standard user accounts suddenly authenticating with administrative privileges (Event ID 4672), especially when combined with unusual parent processes or network logons from unexpected systems. Hunt for cached credential usage (logon type 11) from accounts that shouldn't have cached credentials on particular systems.


**However, the challenge is distinguishing malicious from legitimate**. Many enterprise applications legitimately use token manipulation, and administrative activities generate similar telemetry to attacks. Effective hunting requires understanding normal privileged operations in your environment and the typical escalation patterns used by your administrative tools.






---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./06_persistence.md" >}})
[|NEXT|]({{< ref "./08_evade.md" >}})

