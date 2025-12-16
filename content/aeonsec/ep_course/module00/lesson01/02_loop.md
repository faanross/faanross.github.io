---
showTableOfContents: true
title: "The Threat Hunting Loop"
type: "page"
---




The Threat Hunting Loop is a continuous, iterative process that transforms raw data into actionable security improvements. It's not a one-time activity but a cycle that constantly refines your security posture.

## Phase 1: Hypothesis

**What is a Hypothesis?**

A hypothesis is an educated guess about how an attacker might be operating in your environment. It's based on:

- **Threat Intelligence**: Reports about adversary tactics
- **Known Vulnerabilities**: Weaknesses in your environment
- **Anomalous Patterns**: Unusual activity you've observed
- **Industry Trends**: Attacks targeting your sector
- **Attacker Motivation**: What would adversaries want from your organization?

**Example Hypotheses:**

**Hypothesis 1**: "Attackers may be using Office macros to establish initial access because our users frequently receive documents from external sources and macros are enabled by policy."

**Hypothesis 2**: "An insider threat may be exfiltrating customer data via cloud storage because we lack visibility into personal cloud service usage."

**Hypothesis 3**: "Attackers who compromised our VPN may be using Kerberoasting to escalate privileges because we have service accounts with weak passwords and SPNs registered."

**Characteristics of Good Hypotheses:**

- **Specific**: Not "attackers are in my network" but "attackers may be using WMI for lateral movement"
- **Testable**: You can design queries/searches to prove or disprove it
- **Based on Evidence**: Grounded in threat intelligence, vulnerabilities, or observations
- **Scoped**: Focused on specific techniques or attack vectors
- **Actionable**: If proven true, you know what to do next



## Phase 2: Investigation

**Data Collection:**

This phase involves gathering the telemetry needed to test your hypothesis:

- **Endpoint Telemetry**:

    - Sysmon logs (process creation, network connections, file operations)
    - Windows Event Logs (authentication, privilege use, service creation)
    - PowerShell logs (script block logging, module logging)
    - Application logs
- **Network Telemetry**:

    - Network connection logs
    - DNS queries
    - Proxy logs
    - Firewall logs
- **Contextual Data**:

    - Asset inventory (what systems exist, their purpose)
    - User behaviour baselines
    - Scheduled maintenance windows
    - Recent changes

**Investigation Techniques:**

Let's use **Hypothesis 1** (Office macros for initial access) as our example:

**Step 1: Define Search Parameters**

```
What to look for:
- Office applications (WINWORD.EXE, EXCEL.EXE, POWERPNT.EXE)
- Spawning child processes
- Especially: cmd.exe, powershell.exe, wscript.exe, cscript.exe
- Within specific timeframe (last 30 days)
```

**Step 2: Query Telemetry**

```
Using Sysmon Event ID 1 (Process Creation):
- Filter ParentImage = *\WINWORD.EXE OR *\EXCEL.EXE
- AND Image = *\powershell.exe OR *\cmd.exe
- Review CommandLine field for suspicious arguments
```

**Step 3: Analyze Results**

```
For each instance found:
1. Is this expected behavior? (Check with user/business context)
2. What was the command line? (Encoded? Downloading files?)
3. What happened next? (Follow the process tree)
4. What network connections occurred? (C2 callback?)
5. What files were created? (Malware dropped?)
```

**Investigation Anti-Patterns to Avoid:**

- **Query Overload**: Running too many broad queries that return millions of results
- **Analysis Paralysis**: Getting stuck examining every minor anomaly
- **Confirmation Bias**: Only looking for evidence that supports your hypothesis
- **Scope Creep**: Starting with one hypothesis but chasing every tangent




## Phase 3: Pattern Discovery

**What is a Pattern?**

A pattern is a repeatable indicator or sequence of events that signals malicious activity. Patterns emerge when you find something suspicious during investigation.

**Types of Patterns:**

**Behavioral Patterns:**

```
Example: Beaconing
- Malware calls home every 60 seconds
- Pattern: Periodic network connections to same external IP
- Detection: Statistical analysis of connection frequency
```

**Artifact Patterns:**

```
Example: Credential Dumping
- LSASS.exe accessed by non-system process
- Pattern: Suspicious process opening handle to LSASS with PROCESS_VM_READ
- Detection: Sysmon Event ID 10 with specific access rights
```

**Sequence Patterns:**

```
Example: Lateral Movement
- User logs on Type 3 (network logon) to Server A
- Seconds later, Server A connects to Server B with same user
- Pattern: Rapid cross-system authentication chain
- Detection: Timeline correlation of authentication events
```

**From Pattern to Detection Rule:**

Let's convert our Office macro finding into a detection rule:

**Discovery:**

```
Found: WINWORD.EXE spawned PowerShell.exe with encoded command
Command: powershell.exe -enc JABjAGwAaQBlAG4AdAAgAD0AIABOAGUAdwAt...
Decoded: Downloads payload from attacker infrastructure
```

**Pattern Identified:**

```
Office Application → PowerShell → Encoded Command → Network Download
```

**Detection Rule (Pseudocode):**

```
IF ParentProcess = (WINWORD.EXE OR EXCEL.EXE OR POWERPNT.EXE)
AND ChildProcess = powershell.exe
AND CommandLine CONTAINS ("-enc" OR "-encodedcommand" OR "downloadstring")
THEN Alert: "Suspicious Office Macro Behavior"
```

**Pattern Documentation:**

For each pattern discovered, document:

1. **Technical Details**: Exact sequence of events
2. **MITRE ATT&CK Mapping**: Which technique(s) does this represent?
3. **Detection Logic**: How to identify this pattern
4. **False Positive Potential**: What legitimate activity might trigger this?
5. **Response Actions**: What to do when detected

## Phase 4: Enrichment

**What is Enrichment?**

Enrichment is the process of taking your discoveries and using them to improve your overall security posture. This is where threat hunting provides return on investment.

**Enrichment Activities:**

**1. Create New Detections:**

```
Action: Implement the detection rule you developed
Tools: SIEM rules, EDR policies, custom scripts
Result: Automated alerting for future occurrences
```

**2. Update Threat Intelligence:**

```
Action: Document IOCs (IPs, domains, hashes, patterns)
Tools: Threat intelligence platform, IOC feeds
Result: Enriched threat intelligence for organization and community
```

**3. Improve Logging:**

```
Discovery: "We can't see X because we're not logging Y"
Action: Enable additional logging sources or increase verbosity
Result: Better visibility for future hunts
```

**4. Remediate Vulnerabilities:**

```
Discovery: "Attackers exploited misconfigured Z"
Action: Fix configuration across environment
Result: Attack surface reduced
```

**5. Train Security Team:**

```
Action: Share findings with SOC analysts
Tools: Documentation, training sessions, playbooks
Result: Improved analyst skills and faster response
```

**6. Inform Risk Management:**

```
Action: Report findings to leadership
Content: What was found, impact, remediation
Result: Risk-informed decision making
```

**Enrichment Example:**

Following our Office macro investigation:

```
Week 1: Hunt discovers macro-based initial access
  ↓
Week 2: Detection rule deployed to SIEM
  ↓
Week 3: SOC analysts trained on macro threat patterns
  ↓
Week 4: GPO updated to disable macros from internet sources
  ↓
Week 5: Email gateway updated to quarantine macro documents
  ↓
Week 6: User awareness training on macro risks
  ↓
Result: Multi-layered defense against this attack vector
```

## The Loop Continues

**Critical Point**: The Threat Hunting Loop never truly ends. Each hunt generates new questions, each detection rule may need tuning, and the threat landscape constantly evolves.

**Example of Continuous Improvement:**

```
Hunt 1: Find Office macro malware
  ↓
Detection Rule: Alert on Office→PowerShell
  ↓
Attacker Adapts: Uses WMI instead of PowerShell
  ↓
Hunt 2: Find Office→WMI patterns
  ↓
Detection Rule: Expanded to include WMI
  ↓
Attacker Adapts: Uses VBA to directly download via URLDownloadToFile
  ↓
Hunt 3: Monitor for Office process network connections
  ↓
... and so on
```






---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

