---
showTableOfContents: true
title: "Integration Philosophy: ELK as Central Nervous System"
type: "page"
---

### 8.1 The Hub-and-Spoke Model

Modern security architecture treats the SIEM (in our case, ELK) as the central nervous system:

```
                    ┌──────────────┐
                    │     ELK      │
                    │  (Central    │
                    │   Hub)       │
                    └───────┬──────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼──────┐  ┌─────────▼────────┐  ┌──────▼─────┐
│   Endpoint   │  │    Network       │  │   Cloud    │
│   Detection  │  │    Monitoring    │  │   Logs     │
│   (EDR)      │  │  (Zeek/Suricata) │  │  (AWS/GCP) │
└──────────────┘  └──────────────────┘  └────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                    ┌───────▼──────┐
                    │Threat Intel  │
                    │  Platform    │
                    └──────────────┘
                            │
                    ┌───────▼──────┐
                    │    SOAR      │
                    │ (Automated   │
                    │  Response)   │
                    └──────────────┘
```

**Why This Architecture?**

1. **Single Source of Truth**: All security data flows through one system
2. **Correlation**: Events from different sources can be connected
3. **Context**: Each alert gets enriched with data from other systems
4. **Response Orchestration**: SOAR can pull context from ELK to inform automated responses

### 8.2 Integration Patterns

**Pattern 1: Log Forwarding**

- Security tools send logs directly to ELK
- Most common: syslog, HTTP/HTTPS API, file-based

**Pattern 2: API Polling**

- ELK (via Logstash or custom tools) polls security tool APIs
- Used when tools don't push logs (e.g., cloud services)

**Pattern 3: Webhook/Callback**

- Security tools call ELK webhooks when events occur
- Real-time, event-driven integration

**Pattern 4: Bidirectional**

- ELK receives data AND sends enrichment back
- Example: ELK queries threat intel platform, gets verdict, adds to event

### 8.3 The Detection Hierarchy

Not all detections happen in ELK. Modern security uses a layered approach:

**Layer 1: Endpoint (EDR)**

- Detects: Malware execution, suspicious process behavior
- Speed: Immediate (milliseconds)
- Scope: Single endpoint
- Response: Can isolate endpoint immediately

**Layer 2: Network (IDS/IPS)**

- Detects: Network-based attacks, C2 communication
- Speed: Very fast (seconds)
- Scope: Network segment
- Response: Can block connections

**Layer 3: SIEM (ELK)**

- Detects: Multi-stage attacks, lateral movement, anomalies
- Speed: Fast (minutes)
- Scope: Entire environment
- Response: Orchestrate response across tools

**Layer 4: Threat Hunting (Manual in ELK)**

- Detects: Novel techniques, APT campaigns, unknown threats
- Speed: Variable (hours to days)
- Scope: Comprehensive
- Response: Custom, case-by-case

**ELK's role is layers 3 and 4**, while also aggregating alerts from layers 1 and 2.

### 8.4 Real-World Integration Example

**Scenario: Ransomware Attack**

**Timeline of Events:**

```
T+0:00 - User clicks phishing link
T+0:01 - EDR detects suspicious PowerShell, generates alert → sent to ELK
T+0:05 - Zeek detects C2 beaconing pattern → sent to ELK
T+0:10 - ELK correlation rule fires: "PowerShell + external connection from same host"
T+0:11 - ELK sends alert to SOAR platform
T+0:12 - SOAR queries ELK for all events from affected host (last 24h)
T+0:13 - SOAR isolates host via EDR API
T+0:14 - SOAR creates ticket in ITSM system
T+0:15 - Analyst reviews enriched case in SOAR (all context from ELK)
```

**Without Integration:**

- EDR alert sits in EDR console
- Network alert sits in IDS console
- Analyst must manually correlate
- Response takes 30+ minutes

**With ELK Integration:**

- All context in one place immediately
- Automated response in <15 seconds
- Comprehensive timeline for investigation




---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./07_role.md" >}})
[|NEXT|]({{< ref "./09_summary.md" >}})

