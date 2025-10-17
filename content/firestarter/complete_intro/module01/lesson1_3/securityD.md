---
showTableOfContents: true
title: "Part 4D - Windows Process Protection Framework"
type: "page"
---

## Windows Process Protection Framework

Before we move on to the next part of Lesson 1.3 I wanted to share the following framework. Note that this is more of a reference, and not a lesson you need to read and imbibe top to bottom per se. That being said, going through it won't hurt, and it's not all that long.

We've learned of so many different types of defenses that Windows uses, but what often happens is you'll find distinct, common "categories" (for lack of a better term) of processes based on specific combinations of defense types they possess.

So here I just wanted to share this roadmap. It's not precise, but it will give you a sense of clusters of processes, and what's usually required in order to gain control, or manipulate them in some way.

Please also note that this is not an official Microsoft classification system. This framework is a conceptual model designed to help understand the various protection mechanisms Windows employs to secure processes. Real-world security is complex and these categories can overlap.



## Protection Mechanisms Overview

Before we discuss the process categories, lets' quickly review at a birds-eye level the different tools Windows uses to secure applications and accounts:

### Access Control Mechanisms

- **Security Descriptors/ACLs** - Standard Windows access control
- **User Account Separation** - Different users can't access each other's processes
- **Integrity Levels** - Low/Medium/High/System integrity separation
- **Privileges** - Special rights like SeDebugPrivilege, SeLoadDriverPrivilege
- **Session Isolation** - Session 0 (services) vs User sessions

### Kernel-Level Protection

- **Protected Process (PP)** - DRM-level protection, Microsoft signatures only
- **Protected Process Light (PPL)** - Antimalware protection, third-party signatures allowed
- **Critical Process Flag** - Termination causes BSOD
- **Process Mitigation Policies** - DEP, ASLR, CFG, etc.

### Hardware & Virtualization-Based Security

- **Virtualization-Based Security (VBS)** - Hypervisor isolation
- **Credential Guard** - Credential isolation in VTL 1
- **HVCI (Memory Integrity)** - Hypervisor-enforced code integrity
- **Secure Boot** - Boot chain integrity
- **TPM** - Hardware root of trust

---

## Process Categories

### Category A: User-Owned Processes

**Protection Model:** Basic user-level isolation

**What Protects Them:**

- Security descriptor (process owner)
- Standard Windows ACLs
- Integrity level (usually Medium)

**What Can Access Them:**

- ✅ The user who owns the process (full control)
- ✅ Administrators (with appropriate privileges)
- ❌ Other standard users
- ❌ Lower integrity processes (can't write up)

**Typical Examples:**

- Applications launched by the user
- `notepad.exe` (your own)
- `chrome.exe` (your own)
- Calculator, Office apps, etc.

**Attack Surface:**

- Same-user malware has full access
- DLL injection from same user
- Memory manipulation from same user
- No protection against owner

**Key Concept:** You always have full control over your own processes. This is by design - user convenience vs. security tradeoff.

---

### Category B: Other Users' Processes

**Protection Model:** User account separation

**What Protects Them:**

- Security descriptor (different owner)
- ACL blocking other users
- Session isolation (sometimes)
- Integrity level separation

**What Can Access Them:**

- ❌ Other standard users (ACCESS_DENIED)
- ✅ SYSTEM account
- ✅ Administrators (often, depending on configuration)
- ✅ Administrators with SeDebugPrivilege (definitely)

**Typical Examples:**

- `notepad.exe` launched by User A, accessed by User B
- Another user's applications on a shared machine

**Attack Surface:**

- Protected from lateral movement between users
- Admin accounts can typically access
- Privilege escalation can bypass

**Key Concept:** Windows' multi-user design. Each user's processes are isolated from others, but not from admins.

---

### Category C: SYSTEM Processes (Regular Services)

**Protection Model:** Privilege-based access control

**What Protects Them:**

- Runs as NT AUTHORITY\SYSTEM
- Different security context from users
- Requires SeDebugPrivilege to access
- Often has service-specific ACLs

**What Can Access Them:**

- ❌ Standard users (ACCESS_DENIED)
- ❌ Administrators without SeDebugPrivilege (ACCESS_DENIED)
- ✅ Administrators WITH SeDebugPrivilege enabled (full access)
- ✅ Kernel mode (always)

**Typical Examples:**

- `spoolsv.exe` (Print Spooler)
- Most `svchost.exe` instances
- `services.exe` (without PP)
- Regular Windows services

**Attack Surface:**

- Protected from standard users
- Vulnerable to admin-level malware
- SeDebugPrivilege is the key
- Kernel exploits can access

**Key Concept:** This is where SeDebugPrivilege matters. It's the "break glass in case of debugging" privilege that lets admins access system processes.

---

### Category D: Critical System Processes

**Protection Model:** Stability protection (not security-focused)

**What Protects Them:**

- Runs as NT AUTHORITY\SYSTEM
- Critical process flag set (`RtlSetProcessIsCritical`)
- Often in Session 0
- May have additional kernel protections

**What Can Access Them:**

- ❌ Standard users (ACCESS_DENIED)
- ✅ Administrators with SeDebugPrivilege (CAN access memory)
- ✅ Kernel mode (always)
- ⚠️ Termination causes BSOD (Bug Check 0xF4)

**Special Characteristics:**

- **CAN be read/written with SeDebugPrivilege**
- **CANNOT be terminated without crashing Windows**
- **May be partially accessible or completely blocked depending on implementation**

**Typical Examples:**

- `csrss.exe` (Client/Server Runtime Subsystem)
- `wininit.exe` (Windows Initialization)
- `smss.exe` (Session Manager) - also has extra kernel protections

**Attack Surface:**

- Memory accessible with SeDebugPrivilege
- Code injection possible (but risky)
- Termination = instant BSOD
- Some have additional kernel-level access restrictions

**Key Concept:** Protected for system stability, not security. The goal is preventing accidental crashes, not preventing malicious access. Killing these processes is like pulling the foundation out from under Windows.

**Why smss.exe is Special:** `smss.exe` is created directly by the kernel (`ntoskrnl.exe`) during early boot, runs in Session 0, and has additional kernel-level protections beyond just the critical flag. It often resists even OpenProcess with SeDebugPrivilege.

---

### Category E: Protected Process Light (PPL)

**Protection Model:** Kernel-enforced code signing requirements

**What Protects Them:**

- Protected Process Light flag
- Kernel validates all access attempts
- Code signing certificate requirements
- Memory access blocked at kernel level

**What Can Access Them:**

- ❌ Standard users (ACCESS_DENIED)
- ❌ Administrators (ACCESS_DENIED)
- ❌ Administrators WITH SeDebugPrivilege (STILL ACCESS_DENIED)
- ✅ Kernel drivers with appropriate signatures (Antimalware, WinTcb, LSA)
- ✅ Kernel exploits
- ✅ DMA attacks (physical access)

**The "Tricky Handle" Behavior:**

```
OpenProcess() → Returns valid handle (compatibility)
ReadProcessMemory() → ACCESS_DENIED (protection enforced)
```

Windows allows handle creation but blocks dangerous operations.

**Signing Requirements:**

- Antimalware (registered AV vendors)
- Windows (Microsoft)
- WinTcb (Trusted Computing Base)
- LSA (Local Security Authority)

**Typical Examples:**

- `lsass.exe` (when Credential Guard not enabled, on some systems)
- `MsMpEng.exe` (Windows Defender)
- `audiodg.exe` (Audio Device Graph Isolation)
- Antivirus processes

**Attack Surface:**

- Kernel driver required (must be signed)
- Driver signature enforcement must be bypassed
- Kernel exploits can bypass
- Physical DMA attacks (Thunderbolt, PCI Express)
- Vulnerable drivers (bring your own vulnerable driver - BYOVD)

**Key Concept:** SeDebugPrivilege stops being useful here. PPL is kernel-enforced protection designed to protect security-critical processes from ALL user-mode access, even with maximum privileges.

**Historical Note:** PPL was introduced in Windows 8.1 specifically to protect `lsass.exe` from credential theft attacks like Mimikatz.

---

### Category F: Protected Process (PP)

**Protection Model:** Full DRM-level protection

**What Protects Them:**

- Protected Process flag (stronger than PPL)
- Strictest kernel enforcement
- Microsoft-only code signing
- Isolated from ALL non-PP processes

**What Can Access Them:**

- ❌ Everything in user mode (complete block)
- ❌ PPL processes (PP > PPL in hierarchy)
- ❌ Kernel drivers with third-party signatures
- ✅ Kernel drivers with Microsoft Windows signatures ONLY
- ✅ Kernel exploits
- ✅ Hardware-based attacks

**Signing Requirements:**

- ONLY Microsoft Windows component signatures
- Categories: Windows, WinTcb
- Third-party signatures explicitly NOT accepted

**Typical Examples:**

- DRM media processes (protected audio/video paths)
- `services.exe` (on modern Windows with certain features enabled)
- Windows Update components (some)
- Credential Guard components (`lsaiso.exe` - though this runs in VTL 1)

**Attack Surface:**

- Requires Microsoft-signed kernel driver
- Kernel exploits
- Hardware vulnerabilities
- Firmware-level attacks
- Supply chain compromise

**Key Concept:** Designed for DRM (Digital Rights Management) and highest-security Windows components. The goal is protecting content and credentials from ALL software attacks, including sophisticated malware.

**PP vs PPL Hierarchy:**

```
Protected Process (PP)
    ↓ can access
Protected Process Light (PPL)
    ↓ can access
Regular Processes
```

---

### Category G: Kernel Mode (Ring 0)

**Protection Model:** Privileged processor mode

**What Protects Them:**

- Runs in Ring 0 (privileged CPU mode)
- PatchGuard (Kernel Patch Protection) - detects kernel modifications
- Driver Signature Enforcement (DSE) - blocks unsigned drivers
- HVCI (optional) - hypervisor validates code execution
- Secure Boot - validates boot chain integrity

**What Can Access Them:**

- ❌ All user-mode code (architectural CPU protection)
- ❌ Unsigned drivers (blocked by DSE)
- ✅ Signed kernel drivers (with WHQL certification)
- ✅ Hypervisor (Ring -1)
- ✅ Boot-time exploits (before protections load)
- ✅ Hardware attacks

**PatchGuard (KPP):**

- Monitors critical kernel structures (IDT, GDT, SSDT, etc.)
- Detects unauthorized modifications
- Triggers BSOD (CRITICAL_STRUCTURE_CORRUPTION) if tampering detected
- Active on x64 Windows only
- Checks periodically and randomly

**Driver Signature Enforcement:**

- All drivers must be digitally signed
- Signature validated against Microsoft's certificate chain
- Cannot be disabled on modern Windows (without boot options)
- Test mode for development

**HVCI (Memory Integrity):**

- Uses hardware virtualization (VT-x/AMD-V)
- Hypervisor validates all executable code pages
- Prevents code injection even with kernel access
- Blocks unsigned code from executing
- Optional on Windows 10, default on Windows 11 (with capable hardware)

**Capabilities:**

- Read/write ANY process memory (including PP/PPL)
- Hook system calls
- Modify kernel structures
- Install rootkits
- Disable security features (if PatchGuard bypassed)
- Complete system control

**Limitations (with HVCI):**

- Cannot execute unsigned code pages
- Cannot modify read-only memory marked by hypervisor
- Cannot bypass hypervisor-enforced policies

**Typical Examples:**

- `ntoskrnl.exe` (Windows kernel)
- `hal.dll` (Hardware Abstraction Layer)
- `win32k.sys` (Win32 subsystem kernel component)
- All `.sys` drivers (graphics, network, storage)
- Antivirus kernel drivers
- Hypervisor drivers (Hyper-V)

**Attack Surface:**

- Vulnerable signed drivers (BYOVD attacks)
- Kernel exploits (use-after-free, buffer overflow, etc.)
- Zero-day vulnerabilities
- Supply chain attacks (compromised signed drivers)
- Physical access to disable Secure Boot
- DMA attacks

**Key Concept:** Kernel mode is the highest privilege in traditional OS architecture. Gaining kernel access means you effectively own the system - you can bypass ALL user-mode protections (PPL, PP, everything).

---

### Category H: Hypervisor / Secure Kernel (Ring -1)

**Protection Model:** Hardware virtualization-based isolation

**What Protects Them:**

- Runs at Ring -1 (higher privilege than kernel)
- Hardware virtualization (VT-x/AMD-V required)
- Isolated execution environment
- VTL 1 (Virtual Trust Level 1) - separate from normal Windows
- SLAT (Second Level Address Translation)

**What Can Access Them:**

- ❌ User mode (multiple privilege levels below)
- ❌ Kernel mode (one privilege level below)
- ❌ Signed drivers (still in VTL 0)
- ✅ Hypervisor exploits (extremely rare)
- ✅ CPU vulnerabilities (Spectre, Meltdown class)
- ✅ Firmware-level attacks
- ✅ Physical hardware attacks

**What It Protects:**

- Virtualization-Based Security (VBS) components
- Secure Kernel (`securekernel.exe`)
- Credential Guard (`lsaiso.exe` - isolated lsass)
- Device Guard / Application Control policies
- Kernel memory integrity (HVCI enforcement)
- Virtual Secure Mode (VSM) processes

**Architecture:**

```
VTL 1 (Virtual Trust Level 1 - Secure World)
  ├─ Secure Kernel
  ├─ lsaiso.exe (Credential Guard)
  └─ Security enforcement
       ↓ enforces policies on
VTL 0 (Virtual Trust Level 0 - Normal World)
  ├─ Windows Kernel
  ├─ Drivers
  └─ All user processes
```

**Typical Examples:**

- Hyper-V Hypervisor (`hvix64.exe`, `hvax64.exe`)
- Secure Kernel (`securekernel.exe`)
- `lsaiso.exe` (Credential Guard - isolated lsass)
- VBS components

**Attack Surface:**

- Hypervisor vulnerabilities (very rare, high-value targets)
- CPU vulnerabilities (microarchitectural)
- Firmware exploits (UEFI vulnerabilities)
- Supply chain attacks on hardware/firmware
- Side-channel attacks

**Key Concept:** This is "security by isolation." Even if an attacker gets kernel access, they're still in VTL 0 and cannot access VTL 1 components. The hypervisor enforces this separation using hardware features.

**Important:** VBS/HVCI is NOT enabled on all systems:

- Requires compatible hardware (VT-x/AMD-V, SLAT)
- Optional on Windows 10
- Default on Windows 11 (with compatible hardware)
- Enterprise environments may enable via Group Policy

---

## Attack Progression & Privilege Escalation

Understanding these categories helps visualize attack progression:

### Typical Attack Path:

```
1. Initial Access (Standard User)
   ↓ exploit vulnerability
2. Code Execution (User Context)
   ↓ privilege escalation
3. Administrator Access
   ↓ enable SeDebugPrivilege
4. SYSTEM Process Access (Category C)
   ↓ kernel exploit
5. Kernel Mode Access (Category G)
   ↓ bypass/disable protections
6. PPL/PP Access (Categories E/F)
   ↓ hypervisor exploit (very rare)
7. Hypervisor Access (Category H)
```

### Defense in Depth:

Each category represents a layer of defense:

- **Early layers** (A-C): Stop casual attackers and basic malware
- **Middle layers** (D-F): Stop sophisticated malware and APTs
- **Deep layers** (G-H): Stop nation-state attackers and zero-days

No single layer is perfect - defense requires multiple layers.

---

## Security Mechanism Effectiveness by Category

Understanding which security mechanisms are effective at each protection level:

|Category|User ACLs|Admin Rights|Privileges (SeDebug, etc.)|Signed Drivers|Kernel Exploits|Hardware/Firmware|
|---|---|---|---|---|---|---|
|A - Your processes|❌|✅|✅|✅|✅|✅|
|B - Other users|✅|✅|✅|✅|✅|✅|
|C - SYSTEM processes|✅|⚠️|✅|✅|✅|✅|
|D - Critical processes|✅|⚠️|✅|✅|✅|✅|
|E - PPL|✅|✅|✅|⚠️|✅|✅|
|F - PP|✅|✅|✅|⚠️|✅|✅|
|G - Kernel|✅|✅|✅|✅|⚠️|✅|
|H - Hypervisor|✅|✅|✅|✅|✅|⚠️|

**Legend:**

- ✅ Blocks this attack vector effectively
- ⚠️ Partially effective or bypasses this protection
- ❌ Not effective

**Key Insight:** Each category requires progressively more sophisticated attacks. No single security mechanism protects everything - defense requires layering multiple mechanisms.

---

## Practical Implications for Offensive Security

- User → Admin: Easiest (many paths)
- Admin → SYSTEM: Easy (with SeDebugPrivilege)
- SYSTEM → Kernel: Moderate (need exploit)
- Kernel → Hypervisor: Very difficult (rare exploits)

## Common Misconceptions

### ❌ "Administrator has full control over everything"

**Reality:** Admins cannot access PPL/PP processes without kernel access.

### ❌ "SeDebugPrivilege lets you debug anything"

**Reality:** Stops at kernel-enforced protection (PPL/PP).

### ❌ "Protected Process is just a flag"

**Reality:** It's kernel-enforced with hardware backing (when VBS enabled).

### ❌ "Kernel access means total control"

**Reality:** With VBS/HVCI, even kernel is monitored by hypervisor.

### ❌ "These 'levels' are official Microsoft categories"

**Reality:** This is a conceptual framework. Microsoft uses different terminology.

## Microsoft's Official Terminology

For completeness, here's how Microsoft officially describes these:

- **Protected Process** - Official term, documented
- **Protected Process Light** - Official term, documented
- **Virtualization-Based Security** - Official term
- **Virtual Secure Mode** - Official term
- **Credential Guard** - Official product name
- **HVCI / Memory Integrity** - Official terms

The "category" framework presented here is for educational purposes to help understand how these pieces fit together.

## Conclusion

Windows process protection is a layered, complex system. No single protection mechanism provides complete security. Instead, Windows combines:

1. **Access control** (ACLs, privileges)
2. **Isolation** (users, sessions, integrity levels)
3. **Kernel enforcement** (PPL, PP)
4. **Hardware features** (VBS, HVCI)
5. **Process mitigations** (DEP, ASLR, CFG)

Understanding these categories helps you:

- Design better security architectures
- Understand attack surfaces
- Make informed security decisions
- Debug and research effectively

**Remember:** Security is about making attacks more expensive, not impossible. Each layer increases the cost for attackers - in time, money, and expertise required.


_This guide represents a conceptual understanding as of 2025. Windows security continues to evolve with each release._



---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./securityC.md" >}})
[|NEXT|]({{< ref "./pe.md" >}})