---
showTableOfContents: true
title: "Lesson 1.1: External Reconnaissance (Attacker Perspective)"
type: "page"
---


## Introduction: The Reconnaissance Phase

External reconnaissance is the first stage of any targeted cyber attack. Before an adversary can breach your network, they must first gather intelligence about it. This phase occurs **entirely outside your network perimeter**, using publicly available information and tools that generate no logs on your systems. The attacker is invisible at this stage - operating in the shadows of the internet, piecing together a profile of your organization.

Understanding reconnaissance from the attacker's perspective is crucial for threat hunters because:

1. **Attribution and IOCs**: The tools and techniques used leave traces in various public systems
2. **Defensive Posture**: Knowing what information is exposed helps reduce your attack surface
3. **Early Warning**: Detecting reconnaissance attempts can provide advance notice of targeted attacks
4. **Threat Modeling**: Understanding what attackers can learn informs your security priorities

This lesson examines the methodologies, tools, and techniques adversaries use to build a comprehensive target profile - all before they ever touch your network.

---



## The OSINT Methodology

### What is OSINT?

**Open Source Intelligence (OSINT)** refers to intelligence collected from publicly available sources. In the context of cyber reconnaissance, this includes:

- Public websites and social media
- Search engines and cached content
- DNS records and domain registration data
- Public code repositories
- Certificate transparency logs
- Job postings and employee LinkedIn profiles
- News articles and press releases
- Public file metadata

### The OSINT Collection Cycle

Professional reconnaissance follows a structured methodology:

```
┌─────────────────┐
│  1. Planning    │ ← Define objectives, targets, constraints
└────────┬────────┘
         ↓
┌─────────────────┐
│  2. Collection  │ ← Gather raw data from sources
└────────┬────────┘
         ↓
┌─────────────────┐
│  3. Processing  │ ← Filter, organize, deduplicate
└────────┬────────┘
         ↓
┌─────────────────┐
│  4. Analysis    │ ← Identify patterns, relationships
└────────┬────────┘
         ↓
┌─────────────────┐
│  5. Reporting   │ ← Document findings for exploitation
└─────────────────┘
```

### Planning Phase Considerations

Before beginning reconnaissance, attackers define:

- **Target scope**: Which domains, IP ranges, and subsidiaries to investigate
- **Objectives**: What information is needed (emails for phishing, technologies for exploit selection, org structure for social engineering)
- **Time constraints**: How long can they remain undetected (passive vs active reconnaissance)
- **Legal boundaries**: Nation-state actors may ignore these; ethical red teams must respect them

---



## OSINT Tools and Techniques

### 1. theHarvester: Email and Subdomain Discovery

**theHarvester** is a Python-based OSINT tool that aggregates information from multiple public sources. It's particularly effective for discovering:

- Email addresses associated with a domain
- Subdomains
- Virtual hosts
- Employee names
- Open ports and banners

**How theHarvester Works:**

The repo can be found [here](https://github.com/laramies/theHarvester), please refer to it for more info on installing and using it. It also contains a comprehensive + up-to-date list on all the sources it uses.


**Practical Example:**

This is just meant to give some idea of what query results may look like - this is by no means supposed to serve as a guide.

```bash
theHarvester -d corp.local -b all -l 500

# Output interpretation:
[*] Target: corp.local

[*] Searching Google...
    john.smith@corp.local
    sarah.johnson@corp.local
    it-support@corp.local

[*] Searching LinkedIn...
    John Smith - IT Administrator
    Sarah Johnson - Security Analyst
    Michael Brown - Help Desk Technician

[*] Hosts found:
    mail.corp.local
    vpn.corp.local
    portal.corp.local
    dev.corp.local
```

**What Attackers Learn:**

- **Email format**: firstname.lastname@corp.local (useful for credential stuffing)
- **Key personnel**: IT staff are high-value targets
- **Infrastructure**: Mail server, VPN endpoint, development environment
- **Technology stack**: Can infer technologies from subdomain names



### 2. Shodan: The Search Engine for Internet-Connected Devices

**Shodan** continuously scans the entire IPv4 space, cataloging every internet-facing device and service. Unlike Google (which indexes web content), Shodan indexes **banners** - the metadata services expose when contacted.

**What Shodan Reveals:**

- Open ports and running services
- Software versions (often vulnerable)
- Default credentials and misconfigurations
- Industrial control systems (SCADA, ICS)
- IoT devices (cameras, routers, printers)
- Geographic locations of assets

**Shodan Query Syntax:**

```bash
# Search by organization name
org:"Target Corporation"

# Search by domain
hostname:corp.local

# Search by IP range
net:203.0.113.0/24

# Search by service
port:3389 country:US

# Combine filters
hostname:corp.local port:445 os:Windows

# Find specific vulnerabilities
vuln:CVE-2017-0144  # EternalBlue
```

**Practical Example:**

```
Query: hostname:corp.local

Results:
─────────────────────────────────────────────────
IP: 203.0.113.45
Hostnames: vpn.corp.local
Port: 443/tcp
Service: https
  SSL Certificate:
    Subject: CN=vpn.corp.local
    Issuer: Let's Encrypt
    Valid until: 2024-06-15
  HTTP Headers:
    Server: Apache/2.4.41 (Ubuntu)
    X-Powered-By: PHP/7.4.3

IP: 203.0.113.67
Hostnames: mail.corp.local
Port: 25/tcp
Service: smtp
  Banner: 220 mail.corp.local ESMTP Postfix (Ubuntu)
  
Port: 587/tcp
Service: smtp
  
Port: 993/tcp
Service: imaps
```

**What Attackers Learn:**

- **VPN server** running Apache on Ubuntu with PHP (potential exploit targets)
- **Mail server** running Postfix on Ubuntu (SMTP relay misconfiguration?)
- **Open ports** that might be vulnerable
- **Software versions** to search for known exploits










---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

