---
showTableOfContents: true
title: "Part 4C - Integrity Levels and UAC"
type: "page"
---


## The Mandatory Trust Boundary

Beyond traditional discretionary access control (where file owners decide who can access their files), Windows implements a second, mandatory layer of security called **Mandatory Integrity Control (MIC)**. This system enforces a trust hierarchy where every process and every object is assigned an **integrity level** - a measure of how much the operating system trusts it. This trust-based architecture, combined with **User Account Control (UAC)**, fundamentally changed Windows security by ensuring that even administrator accounts don't automatically run with full power.

## Mandatory Integrity Control: The Trust Hierarchy

**Mandatory Integrity Control (MIC)** operates on a simple but powerful principle: processes and objects exist in a hierarchy of trust, and lower-trust entities cannot contaminate higher-trust entities. This is enforced automatically by the kernel - hence "mandatory" - regardless of what traditional permissions say.

### The Five Integrity Levels

Every process, file, registry key, and object in Windows has one of five integrity levels:

```
┌─────────────────────────────────────────────────────────────────┐
│                  INTEGRITY LEVEL HIERARCHY                      │
│                                                                 │
│   System (0x4000) ──────────────────────────────────────────────│
│        ↑                                                        │
│        │ Cannot write/modify higher levels                      │
│        │                                                        │
│   High (0x3000) ────────────────────────────────────────────────│
│        ↑                                                        │
│        │                                                        │
│   Medium (0x2000) ──────────────────────────────────────────────│
│        ↑                                                        │
│        │                                                        │
│   Low (0x1000) ─────────────────────────────────────────────────│
│        ↑                                                        │
│        │                                                        │
│   Untrusted (0x0000) ───────────────────────────────────────────│
│                                                                 │
│   Direction of increasing trust ↑                               │
└─────────────────────────────────────────────────────────────────┘
```

Let's examine each level and what runs at it:

|Level|Hex Value|Who Uses It|Typical Processes|Trust Implications|
|---|---|---|---|---|
|**System**|0x4000|The OS itself|services.exe, lsass.exe, system drivers|Absolute trust; can modify anything below|
|**High**|0x3000|Elevated administrators|Applications launched via "Run as Administrator"|Trusted to modify system; can change most files/registry|
|**Medium**|0x2000|Standard operation|Most user applications, normal desktop processes|Default trust level; can only modify user's own data|
|**Low**|0x1000|Sandboxed processes|IE Protected Mode, Chrome renderer processes|Minimal trust; highly restricted access|
|**Untrusted**|0x0000|Rarely used|Temporary internet files, anonymous pipes|Almost no trust; extremely limited capabilities|

### The Fundamental Rule: No Write Up

The core protection mechanism is deceptively simple: **A process cannot write to or modify objects at a higher integrity level**, even if traditional ACL permissions would grant access.

```
THE NO-WRITE-UP RULE:

Allowed:
  High   →  Medium  ✓ (Can write down)
  Medium →  Low     ✓ (Can write down)
  Medium →  Medium  ✓ (Same level)

Blocked:
  Medium →  High    ✗ (Cannot write up)
  Low    →  Medium  ✗ (Cannot write up)
  Low    →  High    ✗ (Cannot write up)

Reading:
  Lower can typically read higher (unless ACL explicitly denies)
  Writing up is what's blocked
```

**Why this matters:**

Before integrity levels, if malware tricked you into running it, and you were logged in as an administrator, the malware instantly had full administrator rights. With MIC:

1. Normal applications run at **Medium** integrity (even for admin users)
2. System files and registry keys are marked **High** or **System** integrity
3. The malware, running at Medium, cannot modify High/System objects
4. Even with administrator ACL permissions, the write is blocked by MIC
5. Damage is contained to the user's profile - system files remain protected

### Real-World Integrity Check Example

Let's see how this plays out with a concrete example:

```
Scenario: Chrome (Medium integrity) tries to modify the Windows registry

Process:
  Name: chrome.exe
  Integrity Level: Medium (0x2000)
  User: Administrator (with full ACL permissions)
  
Target:
  Object: HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run
  Integrity Level: High (0x3000)
  ACL: Administrators have Full Control
  
Access Check Sequence:

1. ACL Check:
   Chrome's token contains Administrators group → Full Control ✓
   Traditional permission check: GRANTED
   
2. Mandatory Integrity Check (MIC):
   Process integrity: Medium (0x2000)
   Object integrity: High (0x3000)
   0x2000 < 0x3000 → Write operations DENIED ✗
   
Final Result: ACCESS DENIED

Despite having administrator ACL permissions, Chrome cannot modify
the registry key because mandatory integrity control blocks writes
from Medium to High integrity levels.
```

This is why browser exploits can't simply write to startup registry keys anymore - the integrity boundary prevents it.

## Mini-Lab Checking Your Process's Integrity Level
### Overview
Understanding what integrity level your process runs at is crucial for security purposes. The following application is capable of determining the security level of either the process itself, or another process.

You can find the complete source code [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part04/integrity.go).

You can find a technical overview and explanation of the source code [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part04/integrity_guide.md).




### Running the Code

Once you've compiled and transferred it to the target system we'll run it a few different ways.

Note that if we run the program without any argument it will check the integrity level of the actual process. Additionally, we can provide a PID as an argument to another process if wanted to check the its integrity level instead.

```shell
# Check current process
integrity.exe

# Check specific process (e.g., notepad)
integrity.exe 1234
```


First, run the application in a non-Admin shell:
```powershell
.\integrity.exe
================================
Process Name:    C:\Users\TestUser\Desktop\integrity.exe
Process ID:      9696
Integrity Level: Medium
================================

```


Let's open an Admin shell and run it again:
```powershell
.\integrity.exe
================================
Process Name:    C:\Users\TestUser\Desktop\integrity.exe
Process ID:      6528
Integrity Level: High
================================
```


Now let's get the PID of `lsass`, and then run the program to check it's integrity level:
```powershell
 Get-Process lsass | Select-Object Id
 
  Id
 --
816
```

```powershell
.\integrity.exe 816
================================
Process Name:    lsass.exe
Process ID:      816
Integrity Level: System
================================
```



### Conclusion

So nothing too surprising or revelatory, but we can see that our code has an effective mechanism for determining the integrity level. Consider this a "building block", which we can weave later on into more complex sequences.



## User Account Control (UAC): Integrity Levels in Action

**User Account Control (UAC)** is the UI manifestation of integrity levels. When you see that famous prompt asking "Do you want to allow this app to make changes to your device?", you're witnessing the transition from Medium to High integrity.

### The Split-Token Architecture

When you log in as an administrator, Windows creates two tokens:

```
┌─────────────────────────────────────────────────────────────────┐
│              ADMINISTRATOR LOGIN: SPLIT TOKENS                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  FILTERED TOKEN (Default)              FULL TOKEN (Elevated)    │
│  ─────────────────────────             ──────────────────────   │
│                                                                 │
│  Integrity: Medium (0x2000)            Integrity: High (0x3000) │
│                                                                 │
│  Groups:                               Groups:                  │
│    ✓ Users                               ✓ Users                │
│    ✗ Administrators (disabled)           ✓ Administrators       │
│    ✓ Everyone                            ✓ Everyone             │
│                                                                 │
│  Privileges:                           Privileges:              │
│    SeChangeNotifyPrivilege (enabled)     All admin privileges   │
│    SeShutdownPrivilege (disabled)        SeDebugPrivilege ✓     │
│    SeDebugPrivilege (disabled)           SeBackupPrivilege ✓    │
│                                          SeTakeOwnership ✓      │
│                                                                 │
│  Used for:                             Used for:                │
│    • All normal applications             • Explicitly elevated  │
│    • Chrome, Word, Notepad               • Regedit, cmd (admin) │
│    • Everything you launch               • System installers    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key insight:** Being an administrator doesn't mean running as an administrator. By default, even admin accounts use the filtered token at Medium integrity, with the Administrators group disabled. Only when you explicitly elevate does the full token activate.

### The UAC Elevation Flow

Here's what happens when you launch an application that requires elevation:

```
1. User double-clicks installer.exe
   ↓
2. Windows reads embedded manifest
   <requestedExecutionLevel level="requireAdministrator" />
   ↓
3. System recognizes elevation needed
   ↓
4. UAC prompt appears (Secure Desktop)
   "Do you want to allow this app to make changes?"
   ↓
5. User clicks "Yes"
   ↓
6. consent.exe validates user intent
   ↓
7. New process created with:
   • Full token (not filtered)
   • High integrity (0x3000)
   • Administrators group enabled
   • Powerful privileges available
   ↓
8. Installer can now modify system files
```

**The Secure Desktop:** Notice the screen dims when UAC prompts appear? That's the **Secure Desktop** - a separate, isolated desktop that normal processes can't interact with. This prevents malware from clicking "Yes" programmatically or covering the prompt with a fake UI.

## UAC Bypass: The Attacker's Perspective

While UAC provides significant protection, attackers have developed numerous techniques to bypass it. Understanding these methods is crucial for both offensive and defensive security.

**NOTE**: We'll just overview the bypasses in this lesson, we'll implement actual UAC bypass applications in a later module.

### The Challenge for Attackers

```
ATTACKER'S PROBLEM:

Starting Position:
  ✓ Code execution achieved (via phishing, exploit, etc.)
  ✓ Running as Administrator user account
  ✗ BUT: Process is at Medium integrity
  ✗ Cannot modify system files
  ✗ Cannot install persistence in system locations
  ✗ Cannot access other users' data
  
Goal: Elevate to High integrity WITHOUT triggering UAC prompt
```

### Bypass Strategy 1: Auto-Elevation Abuse

Windows maintains a whitelist of trusted executables that **auto-elevate** - they jump to High integrity without prompting. These are typically built-in Windows tools that users expect to have elevated privileges.

**How it works:**

```
Auto-Elevated Executables (examples):
  • eventvwr.exe (Event Viewer)
  • fodhelper.exe (Features on Demand Helper)
  • computerdefaults.exe (Set Program Defaults)
  • sdclt.exe (Backup and Restore)

These executables:
  1. Are signed by Microsoft
  2. Located in C:\Windows\System32\
  3. Have autoElevate="true" in their manifest
  4. Elevate silently to High integrity
```

**The vulnerability:**

Some auto-elevated executables perform actions that can be hijacked:

```
Example: fodhelper.exe UAC Bypass

1. fodhelper.exe auto-elevates to High integrity (no prompt)
2. It reads registry keys to determine what to launch
3. Attacker modifies their own registry (Medium can write to HKCU):
   
   HKCU\Software\Classes\ms-settings\shell\open\command
   (Default) = "C:\malware.exe"
   
4. Launch fodhelper.exe
5. fodhelper.exe (running High) reads attacker's registry key
6. Launches malware.exe with High integrity!
7. Attacker now has elevated access without UAC prompt
```

**Why this works:**

- Fodhelper is trusted and auto-elevates
- It reads user-controllable registry keys
- Medium integrity can write to HKCU (user's own registry)
- High integrity process launches attacker's code
- Result: Elevation without user consent

### Bypass Strategy 2: DLL Hijacking in Auto-Elevated Processes

Another approach targets how auto-elevated executables load DLLs:

```
DLL Hijacking UAC Bypass:

1. Find auto-elevated executable (e.g., computerdefaults.exe)
2. Identify missing DLL it attempts to load
3. Place malicious DLL in a user-writable location:
   
   User can write to: C:\Users\Bob\AppData\Local\Temp\
   
4. Launch the auto-elevated executable
5. It searches for DLL in load order:
   a. Application directory (protected)
   b. System directories (protected)
   c. Current directory (user-writable!) ← Hijack here
   
6. Loads attacker's DLL with High integrity
7. Malicious code executes elevated
```

**Defense:** Modern Windows uses **SafeDllSearchMode** and **LOAD_LIBRARY_SEARCH_SYSTEM32** to prefer system directories, but misconfigurations or legacy code can still be vulnerable.

### Bypass Strategy 3: Token Manipulation

If an attacker already has certain privileges, they can manipulate tokens directly:

```
Required Privilege: SeImpersonatePrivilege
Common on: IIS worker processes, SQL Server, service accounts

Attack Flow:
1. Trick a SYSTEM-level process into connecting to attacker
2. Use SeImpersonatePrivilege to impersonate that connection
3. Duplicate the SYSTEM token
4. Create new process with SYSTEM token
5. Skip High entirely - go straight to System integrity
```

This is the basis of attacks like **Potato** exploits (Hot Potato, Rotten Potato, Juicy Potato), which abuse Windows RPC and COM to coerce SYSTEM processes into authenticating to attacker-controlled endpoints.

### Bypass Strategy 4: COM Elevation Moniker Abuse

Windows COM (Component Object Model) has an elevation mechanism for specific objects:

```
COM Elevation Moniker Attack:

1. Certain COM objects are marked for auto-elevation
2. Create instance of elevated COM object (no UAC prompt)
3. COM object runs at High integrity
4. Call methods on this object to perform privileged operations
5. Example: Use ICMLuaUtil interface to launch elevated processes

Code Example:
  CoGetObject("Elevation:Administrator!new:{...CLSID...}", NULL, IID, &pObj)
  → Creates elevated COM object without UAC prompt
  → Call methods to execute elevated code
```

**Why this exists:** Certain system operations need to elevate without prompting (like Windows Update), so COM provides a mechanism. Attackers abuse these pre-approved elevation paths.

## Defensive Strategies

Understanding defense informs how we design bypasses:

|Defense Mechanism|How It Helps|Limitations|
|---|---|---|
|**UAC at highest setting**|Requires prompt for all elevation attempts|Users may dismiss prompts habitually|
|**Remove admin rights**|Standard users don't have filtered/full token split|Legitimate admin tasks become difficult|
|**Privilege reduction**|Run as standard user, elevate only when needed|Requires user discipline|
|**Application whitelisting**|Only approved executables can run|Management overhead|
|**EDR monitoring**|Detects suspicious registry changes, DLL loads|Can be evaded with advanced techniques|
|**Patch management**|Microsoft fixes known bypass techniques|Zero-days exist before patches|

## The Security Reality

UAC and integrity levels aren't perfect - they're **speed bumps**, not impenetrable walls. However, they dramatically raise the bar:

**Before MIC/UAC (Windows XP era):**

```
1. Phishing email → User opens attachment
2. Malware runs with administrator rights
3. Instant full system compromise
4. Game over
```

**After MIC/UAC (Vista+):**

```
1. Phishing email → User opens attachment
2. Malware runs at Medium integrity
3. Cannot modify system files
4. Cannot install drivers
5. Cannot persist in system locations
6. Attacker needs additional bypass technique
7. Each bypass is detectable by EDR
8. Multiple chances for detection/prevention
```

The goal isn't to make elevation impossible - it's to make it **visible, auditable, and require multiple steps**, giving defenders multiple opportunities to detect and respond. In modern Windows security, integrity levels and UAC are foundational layers in a defense-in-depth strategy, buying time and creating detection opportunities that simply didn't exist before.















---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./securityB.md" >}})
[|NEXT|]({{< ref "./securityD.md" >}})