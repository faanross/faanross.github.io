---
showTableOfContents: true
title: "Part 7 - Practical Exercises"
type: "page"
---


## **PART 7 - PRACTICAL EXERCISES**

Now that you understand the landscape, let's analyze existing tools hands-on:

### **Exercise 1: Framework Comparison Analysis**

**Objective**: Compare two major frameworks to understand design differences.

1. Metasploit Framework (Ruby/C)
2. Sliver (Go)

**Analysis Template**:

```
For each framework, document:

1. ARCHITECTURE
   • Language(s) used
   • Server/client architecture
   • Payload generation method
   • Communication protocols

2. IMPLANT CAPABILITIES
   • Process injection techniques available
   • Persistence mechanisms
   • Lateral movement features
   • Data exfiltration methods

3. OPERATIONAL SECURITY
   • Encryption (C2 traffic)
   • Jitter/sleep obfuscation
   • Traffic shaping capabilities
   • AV/EDR evasion features

4. EXTENSIBILITY
   • Plugin/module system
   • How easy to add new functionality
   • Community contributions

5. DETECTION PROFILE
   • How well-signatured?
   • Known IOCs (Indicators of Compromise)
   • Behavioral tells

6. USE CASE
   • Pentesting vs red teaming?
   • Strengths and weaknesses
```


**Documentation**

Create a comparison document:

```markdown
# Framework Comparison Analysis

## Metasploit Framework
**Architecture**: Ruby-based framework with C payloads
**Payload Size**: ~73KB (meterpreter/reverse_https)
**Strengths**:
- Extensive module library (2000+ exploits)
- Well-documented
- Active community

**Weaknesses**:
- Heavily signatured by AV/EDR
- Large payload sizes
- Obvious network traffic patterns

**Detection**: Windows Defender catches default payloads immediately

---

## Sliver
**Architecture**: Go-based C2 with compiled implants
**Payload Size**: ~12MB (default, can be reduced)
**Strengths**:
- Modern, actively developed
- Multiple C2 protocols
- Better evasion out-of-box

**Weaknesses**:
- Larger binary size (Go runtime)
- Younger project, smaller community

**Detection**: Lower detection rate than MSF, but still catchable

---

## Comparison Matrix
| Feature | Metasploit | Sliver | Cobalt Strike |
|---------|------------|--------|---------------|
| Payload Size | Small (~70KB) | Large (~12MB) | Medium (~200KB) |
| Languages | Ruby/C | Go | Java/C |
| Evasion | Poor | Good | Excellent (with tuning) |
| Cost | Free | Free | $5900/year |
| Use Case | Pentesting | Red Team | Red Team |
| Signatures | Many | Few | Growing |

## Conclusions
- Metasploit: Great for learning, poor for real operations
- Sliver: Solid choice for red teaming, good reference for this course
- Cobalt Strike: Industry standard, but expensive and increasingly signatured

This analysis demonstrates why custom tools are valuable - known
frameworks have known weaknesses.
```

### **Exercise 2: Malware Source Code Study (Educational)**

**Objective**: Study real malware source code to understand implementation techniques.

**IMPORTANT DISCLAIMER**:

- Study ONLY for educational purposes
- Never execute malware
- Use isolated VM environments
- Legal and ethical obligations apply

**Recommended Sources for Study**:

```
1. GITHUB REPOSITORIES (Public, educational malware samples)
   • theZoo (malware repository for researchers)
   • MalwareBazaar (malware samples database)
   • vx-underground (malware source code collection)

2. TECHNIQUES TO STUDY:
   • Process injection methods
   • API resolution (GetProcAddress alternatives)
   • String obfuscation
   • Persistence mechanisms
   • C2 communication protocols
```

**Sample Analysis Exercise**:

Find a simple backdoor written in C/C++ and answer:

```
ANALYSIS QUESTIONS:

1. API RESOLUTION
   • How does it locate Windows APIs?
   • Does it use GetProcAddress or manual resolution?
   • Are API names obfuscated?

2. PERSISTENCE
   • What persistence mechanism(s) used?
   • Registry keys? Scheduled tasks? Services?
   • How would you improve it?

3. COMMUNICATION
   • What protocol for C2? (HTTP, TCP, DNS?)
   • Is traffic encrypted?
   • What are the IOCs (Indicators of Compromise)?

4. EVASION
   • Any anti-debugging techniques?
   • VM detection?
   • String obfuscation?
   • How would you enhance evasion?

5. CODE QUALITY
   • Is the code well-written?
   • What would you do differently?
   • Any bugs or vulnerabilities in the malware itself?
```

**Example: Studying a Simple Backdoor**

```c
// Example snippet from educational malware sample
// DO NOT USE FOR MALICIOUS PURPOSES

// API Resolution via PEB walking
FARPROC GetProcAddressR(HMODULE hModule, LPCSTR lpProcName) {
    // Walk PEB to find kernel32.dll
    // Parse export table
    // Return function address
    // This technique evades static analysis
}

// Study Questions:
// 1. Why walk PEB instead of using GetModuleHandle?
//    Answer: Evades IAT analysis, more stealthy
//
// 2. What are the limitations of this approach?
//    Answer: Still detectable via behavioral analysis
//
// 3. How would you improve it?
//    Answer: Add API hashing, indirect calls, timing checks
```


### **Exercise 3: Tool Architecture Design**

**Objective**: Design your own offensive tool architecture before writing code.

**Scenario**: You're tasked with building a custom post-exploitation framework for long-term red team engagements. Design the architecture.

**Requirements**:
- Must evade modern EDR solutions
- Multi-operator support
- Encrypted C2 communication
- Modular capabilities
- Cross-platform (Windows primary, Linux secondary)

#### Custom C2 Framework Architecture Design

##### 1. COMPONENT OVERVIEW

###### Server Architecture
- Language: Go (cross-platform, good concurrency)
- Database: SQLite (embedded, simple)
- API: REST + WebSocket (real-time updates)
- Operators: Web UI (React/Vue) or CLI

###### Implant Architecture
- Language: Go (compiled, small, fast)
- Communication: HTTP/S + DNS (fallback)
- Encryption: ChaCha20 + AES
- Size Target: <5MB

##### 2. IMPLANT DESIGN

###### Core Capabilities
1. Process Injection (5 techniques)
2. Credential Harvesting
3. File Operations
4. Screenshot Capture
5. Keylogging
6. Network Scanning
7. Lateral Movement

###### Evasion Features
1. Direct syscalls (unhook NTDLL)
2. String obfuscation (XOR + RC4)
3. Sleep obfuscation (Ekko technique)
4. API hashing (custom algorithm)
5. Certificate pinning
6. Jitter (randomized callback intervals)

###### Communication Protocol
- Traffic Profile: HTTPS + DNS
- User-Agent: Randomized (Chrome/Firefox/Edge)
- TLS: Certificate pinning
- Timing: Jitter between 60-300 seconds
- Size: Variable (padding to avoid signatures)


###### Module System

Core Implant (minimal size)
├─ Module: Lateral Movement (load on demand)
├─ Module: Credential Dumping (load on demand)
├─ Module: Screenshot (load on demand)
└─ Module: Keylogger (load on demand)

Modules transmitted encrypted, loaded reflectively


##### 3. OPERATIONAL SECURITY

###### Server Infrastructure

[Operator] → [Login Server] → [Team Server] → [Redirector] → [Implant] (Authentication) (Control Logic) (Traffic Proxy)
- Redirectors: Disposable, easily replaced
- Team Server: Hidden, never directly exposed
- All traffic: Encrypted end-to-end


###### Persistence Strategy

- Primary: Scheduled Task (user context)
- Fallback: Registry Run Key
- Emergency: WMI Event Subscription
- All persistence: Encrypted payloads, environmental keying


##### 4. OPERATOR INTERFACE

###### Web UI Features
- Active implants dashboard
- Task queuing system
- File browser for compromised systems
- Credential manager
- Activity logs with search
- Collaboration features (chat, notes)

###### CLI Features
- Scriptable operations
- Automation support
- Quick commands for power users

##### 5. COMPARISON TO EXISTING TOOLS

| Feature | My Design | Metasploit | Sliver | Cobalt Strike |
|---------|-----------|------------|--------|---------------|
| Evasion | High | Low | Medium | High |
| Modularity | High | High | Medium | Medium |
| Multi-Op | Yes | Limited | Yes | Yes |
| Protocols | HTTP/S, DNS | Many | Many | HTTP/S, DNS, SMB |
| Cost | $0 (custom) | $0 | $0 | $5900/year |

##### 6. DEVELOPMENT PHASES

Phase 1 (Weeks 1-4): Core implant, basic C2
Phase 2 (Weeks 5-8): Evasion features, syscalls
Phase 3 (Weeks 9-12): Module system, operator UI
Phase 4 (Weeks 13-16): Testing, hardening, docs

##### 7. SUCCESS METRICS

✓ Evades Windows Defender (built-in)
✓ Evades Sophos/CrowdStrike (tested in lab)
✓ No crashes in 24-hour stress test
✓ Supports 100+ concurrent implants
✓ <1% packet loss in C2 communication
✓ Documentation complete for operators

##### LESSONS FROM THIS EXERCISE

This design exercise demonstrates:
• Architectural thinking before coding
• Trade-offs (simplicity vs features)
• Why certain design decisions matter
• How requirements drive implementation

Throughout this course, you'll build many components from this design.
The skills you learn enable you to turn designs like this into reality.


---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./trends.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})