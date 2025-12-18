---
showTableOfContents: true
title: "Understanding the Attacker: Frameworks for Threat Analysis"
type: "page"
---



To hunt effectively, you must think like an attacker. Two frameworks help structure this understanding: the Cyber Kill Chain and MITRE ATT&CK.

## The Cyber Kill Chain

Developed by Lockheed Martin in 2011, the Kill Chain describes seven stages of a cyber attack:

**Stage 1: Reconnaissance**

```
Attacker Activity: Research target organization
Methods:
- OSINT (Open Source Intelligence)
- Social media scraping
- DNS enumeration
- Employee identification
- Technology stack discovery

Defender Goal: Minimize information leakage
```

**Stage 2: Weaponization**

```
Attacker Activity: Create malicious payload
Methods:
- Exploit + backdoor combination
- Malicious document creation
- Trojanized legitimate software

Defender Goal: Threat intelligence about TTPs
```

**Stage 3: Delivery**

```
Attacker Activity: Transmit weapon to target
Methods:
- Phishing emails
- Compromised websites (watering holes)
- USB drops
- Supply chain compromise
- Public exploit 
- Brute force exposed service ports

Defender Goal: Block at perimeter
```

**Stage 4: Exploitation**

```
Attacker Activity: Trigger exploit
Methods:
- User enables macro
- Browser exploit executes
- Vulnerability triggered

Defender Goal: Endpoint protection, patching
```

**Stage 5: Installation**

```
Attacker Activity: Install persistence mechanism
Methods:
- Registry run keys
- Scheduled tasks
- Service creation
- DLL hijacking

Defender Goal: Detect installation artifacts
```

**Stage 6: Command & Control (C2)**

```
Attacker Activity: Establish communication channel
Methods:
- HTTP/HTTPS beaconing
- DNS tunneling
- Cloud service abuse

Defender Goal: Network monitoring, anomaly detection, network threat hunting
```

**Stage 7: Actions on Objectives**

```
Attacker Activity: Achieve mission goals
Methods:
- Data exfiltration
- Ransomware deployment
- Destruction
- Espionage

Defender Goal: Rapid detection and containment
```

**Kill Chain Philosophy:**

The key insight is that defenders only need to break ONE link to stop the attack. Attackers must succeed at ALL stages.

**Limitations of the Kill Chain:**

- **Linear Assumption**: Real attacks are iterative, not linear
- **Perimeter-Focused**: Less relevant for insider threats or compromised credentials
- **Missing Lateral Movement**: No explicit phase for moving within the network
- **Oversimplified**: Modern attacks are more complex than 7 stages








---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

