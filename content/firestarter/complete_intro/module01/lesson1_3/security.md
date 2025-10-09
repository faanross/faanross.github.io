---
showTableOfContents: true
title: "Part 4 - Windows Security Model"
type: "page"
---


## Access Tokens: Your Digital Security Badge

In the Windows security model, every process carries a **digital identity card** called an **access token**. This token is the process's security context - the definitive answer to "who are you and what are you allowed to do?" When your process tries to open a file, read the registry, or access another process, Windows doesn't trust what your code claims; it consults your **primary token** to make authorization decisions. Understanding tokens is fundamental to grasping Windows security, from basic file permissions to sophisticated privilege escalation techniques.

## What Is an Access Token?

An **access token** is a kernel object that encapsulates everything about a security principal - the user account, group memberships, special privileges, and trust level. When you log into Windows, the Local Security Authority (LSA) authenticates your credentials and creates a **primary token** representing your identity. Every process you launch inherits a copy of this token, carrying your security context into the application.

Think of the token as a comprehensive security badge that answers:

- **Who you are** (User SID)
- **What groups you belong to** (Group SIDs)
- **What special powers you have** (Privileges)
- **How much the system trusts you** (Integrity Level)
- **Which session you're in** (Session ID)

The kernel checks this badge on every security-sensitive operation, making the token the cornerstone of Windows access control.


## Token Anatomy: The Five Core Components

Let's dissect what's inside an access token, examining each component and its role in the security ecosystem.

### 1. User SID: Your Unique Identity

The **Security Identifier (SID)** is a unique, immutable identifier for your user account. Unlike usernames (which can be renamed), SIDs never change and are mathematically guaranteed to be unique.

**Structure of a SID:**

```
S-1-5-21-3623811015-3361044348-30300820-1013
│ │ │  │                                  │
│ │ │  │                                  └─ RID (Relative ID): User-specific number
│ │ │  └──────────────────────────────────── Domain/Computer identifier (unique)
│ │ └───────────────────────────────────── Security Authority (5 = NT Authority)
│ └─────────────────────────────────────── Revision (always 1)
└───────────────────────────────────────── Prefix identifier
```

**Common well-known SIDs:**

|SID|Identity|Description|
|---|---|---|
|S-1-5-18|SYSTEM|The all-powerful local system account|
|S-1-5-19|LOCAL SERVICE|Limited service account|
|S-1-5-20|NETWORK SERVICE|Service account with network access|
|S-1-5-21-...-500|Administrator|Built-in administrator (RID 500)|
|S-1-5-21-...-1001+|Regular users|Standard user accounts|

**Why SIDs matter:**

- Permissions are granted to SIDs, not usernames
- Renaming "Bob" to "Robert" doesn't change his SID or access rights
- Deleting and recreating a user generates a new SID - they lose all previous permissions
- Security tools track activity by SID to prevent username spoofing



### 2. Group SIDs: Your Memberships

Beyond your individual identity, your token contains a list of **group SIDs** representing every security group you belong to. Group membership is how Windows implements role-based access control - instead of granting permissions to individual users, administrators grant them to groups.

**Common built-in groups:**

```
┌─────────────────────────────────────────────────────────────────┐
│                      GROUP MEMBERSHIPS                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  BUILTIN\Administrators (S-1-5-32-544)                          │
│  └─ Full control over the system                                │
│     Members can: Install software, modify system files,         │
│                  change security settings, access all data      │
│                                                                 │
│  BUILTIN\Users (S-1-5-32-545)                                   │
│  └─ Standard user group                                         │
│     Members can: Run applications, access own files,            │
│                  modify own profile                             │
│                                                                 │
│  BUILTIN\Power Users (S-1-5-32-547)                             │
│  └─ Legacy compatibility group (deprecated)                     │
│                                                                 │
│  Everyone (S-1-1-0)                                             │
│  └─ Universal group containing all users                        │
│     Often used for public resources                             │
│                                                                 │
│  NT AUTHORITY\Authenticated Users (S-1-5-11)                    │
│  └─ Any logged-in user (excludes Guest)                         │
│                                                                 │
│  NT AUTHORITY\INTERACTIVE (S-1-5-4)                             │
│  └─ Users logged in locally (not network/service)               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```


**How group membership works:**

When Windows checks if you can access a resource, it examines **all** SIDs in your token - your user SID plus every group SID. If **any** of them grant access, you're allowed in:

```
Your Token Contains:
  - User: Alice (S-1-5-21-xxx-1001)
  - Groups: BUILTIN\Users, BUILTIN\Administrators, Everyone

File ACL Says:
  - BUILTIN\Administrators: Full Control
  - BUILTIN\Users: Read
  
Result: You get Full Control (because Administrators group grants it)
```

**Deny vs. Allow:**

- Deny entries always win over Allow entries
- If any SID in your token matches a Deny entry, access is blocked regardless of Allow entries

### 3. Privileges: Special Powers Beyond Permissions

**Privileges** are special rights that override normal security checks. While file permissions control access to specific objects, privileges grant broad, system-wide capabilities that transcend typical access control.

**Critical privileges and their powers:**

|Privilege|Constant|What It Does|Risk Level|
|---|---|---|---|
|**SeDebugPrivilege**|20|Attach debugger to any process, even SYSTEM|🔴 Critical|
|**SeBackupPrivilege**|17|Read any file, bypassing ACLs (for backup)|🔴 Critical|
|**SeRestorePrivilege**|18|Write any file, bypassing ACLs (for restore)|🔴 Critical|
|**SeTakeOwnershipPrivilege**|9|Take ownership of any file/registry key|🔴 Critical|
|**SeLoadDriverPrivilege**|10|Load kernel drivers (kernel code execution)|🔴 Critical|
|**SeSystemEnvironmentPrivilege**|22|Modify firmware environment variables (UEFI)|🟠 High|
|**SeShutdownPrivilege**|19|Shut down the system|🟡 Medium|
|**SeChangeNotifyPrivilege**|23|Bypass traverse checking (traverse folders)|🟢 Low|

**Privilege states:**

Privileges exist in three states within a token:

1. **Not present**: You don't have this privilege at all
2. **Present but disabled**: You have it, but it's not currently active (default for most privileges)
3. **Present and enabled**: Actively in effect, granting the special power

```
Example Token Privileges:

SeChangeNotifyPrivilege     : ENABLED  (always on for traversal)
SeShutdownPrivilege         : DISABLED (you have it, but not active)
SeDebugPrivilege            : DISABLED (you have it, but not active)
SeBackupPrivilege           : NOT PRESENT (you don't have this)
```

**Why the enabled/disabled distinction?**

For safety, most powerful privileges are disabled by default. Applications must explicitly enable them when needed:

```
// Before: SeDebugPrivilege is disabled
OpenProcess(PROCESS_ALL_ACCESS, ...) → Fails on protected processes

// Enable SeDebugPrivilege
AdjustTokenPrivileges(hToken, SeDebugPrivilege, ENABLED)

// After: Now we can debug anything
OpenProcess(PROCESS_ALL_ACCESS, ...) → Success!
```

This design prevents accidental misuse - programs must consciously activate dangerous privileges, making suspicious behavior more detectable.


### 4. Integrity Level: The Trust Hierarchy

Starting with Windows Vista, Microsoft added **Mandatory Integrity Control (MIC)** - a mandatory access control layer that operates independently of traditional permissions. Every process and object has an **integrity level** that represents how much the system trusts it.

**The four integrity tiers:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    INTEGRITY LEVEL HIERARCHY                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  System (0x4000) - HIGHEST TRUST                                │
│  ───────────────────────────────────────────────────────────────│
│  • NT AUTHORITY\SYSTEM processes                                │
│  • Windows services                                             │
│  • Kernel-mode drivers                                          │
│  • Can access/modify anything below                             │
│                                                                 │
│  High (0x3000) - ELEVATED TRUST                                 │
│  ───────────────────────────────────────────────────────────────│
│  • Administrator processes running "elevated" (UAC)             │
│  • Installers, system configuration tools                       │
│  • Can access Medium and Low                                    │
│  • Cannot modify System-level objects                           │
│                                                                 │
│  Medium (0x2000) - STANDARD TRUST (DEFAULT)                     │
│  ───────────────────────────────────────────────────────────────│
│  • Normal user applications                                     │
│  • Most processes run here                                      │
│  • Can access own objects and Low                               │
│  • Cannot modify High or System                                 │
│                                                                 │
│  Low (0x1000) - SANDBOXED/RESTRICTED                            │
│  ───────────────────────────────────────────────────────────────│
│  • Internet Explorer Protected Mode                             │
│  • Sandboxed browsers                                           │
│  • Untrusted content handlers                                   │
│  • Extremely limited access                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```


Note that a user cannot directly log in as **`NT AUTHORITY\SYSTEM`**. It's a special built-in account used by the operating system and its services, not an interactive account for a person to use.

However, a user with administrative privileges **can launch processes** that run with SYSTEM-level permissions, which can have serious implications when it comes to malware.

If malware compromises an administrator account, it can use that access to elevate itself to run as `SYSTEM`. Once it achieves this, it has total control over the machine, including the ability to shut down security tools, achieve persistence via service/driver embedding, as well as the ability to access (RW) just about anything from disk or memory.



**The integrity rule: No-Write-Up**

The fundamental principle is simple: **A process cannot modify objects at a higher integrity level**, even if traditional permissions would allow it.

```
Scenario: Medium-integrity Chrome trying to modify a High-integrity registry key

Traditional ACL Check:
  Chrome's user (Alice) → Administrators group → Full Control ✓

Mandatory Integrity Check:
  Chrome's Integrity: Medium (0x2000)
  Registry Key Integrity: High (0x3000)
  Medium < High → WRITE DENIED ✗

Result: Access denied despite having Full Control in the ACL
```

**Why integrity levels exist:**

Before MIC, malware running as an administrator had unlimited power. With integrity levels:

- UAC keeps normal processes at Medium, even for admin accounts
- Elevated processes run at High, creating a meaningful barrier
- Even if malware tricks you into running it, starting at Medium limits the damage
- System processes remain untouchable at System level


This might seem to contradict what we just mentioned regarding the ability of malware to run with SYSTEM-level permissions having full control, but the distinction lies in the difference between **direct interference** and **authorized creation**.

The integrity level barrier is very real: it prevents a malicious process running at a High integrity level from directly tampering with or hijacking an _existing_ process that is already running at the untouchable System level.

However, a compromised administrator account still holds the _authority_ to make legitimate requests to the operating system. This authority allows the malware to ask core OS components, such as the Service Control Manager or Task Scheduler, to launch a _new_ malicious process that starts with full System-level privileges.



### 5. Session ID: Isolation Between Users

The **Session ID** identifies which Terminal Services session the token belongs to. This matters in multi-user environments:

**Session isolation:**

```
Session 0: System services (non-interactive)
  └─ services.exe, svchost.exe, etc.
  └─ No user interaction, isolated for security

Session 1: Alice's desktop
  └─ explorer.exe, chrome.exe, notepad.exe
  └─ Cannot interact with Session 2

Session 2: Bob's desktop (if using Remote Desktop)
  └─ explorer.exe, word.exe
  └─ Cannot interact with Session 1
```

**Security implications:**

- Processes in different sessions cannot send messages to each other's windows
- Session 0 isolation prevents services from displaying UI that could be exploited
- Session IDs prevent one user from accessing another user's GUI processes


## The Access Check Process: Token Meets ACL

When your process attempts to access a secured object (file, registry key, process, etc.), Windows performs a comprehensive **access check** that evaluates your token against the object's security descriptor.

### Step-by-Step Access Check

Let's walk through a real-world example: Notepad trying to open a sensitive system file.

```
┌─────────────────────────────────────────────────────────────────┐
│                    ACCESS CHECK WALKTHROUGH                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  STEP 1: Process Attempts Access                                │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  Process: notepad.exe                                           │
│  Action: Open file for reading                                  │
│  Target: C:\Windows\System32\config\SAM                         │
│           (Security Accounts Manager database)                  │
│                                                                 │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  STEP 2: Kernel Examines Process Token                          │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  Token Contents:                                                │
│    User SID:      S-1-5-21-xxx-1001 (Bob)                       │
│    Group SIDs:    BUILTIN\Users (S-1-5-32-545)                  │
│                   Everyone (S-1-1-0)                            │
│                   NT AUTHORITY\Authenticated Users              │
│    Integrity:     Medium (0x2000)                               │
│    Privileges:    SeShutdownPrivilege (disabled)                │
│                   SeChangeNotifyPrivilege (enabled)             │
│                                                                 │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  STEP 3: Kernel Retrieves Object's DACL                         │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  File: C:\Windows\System32\config\SAM                           │
│  Owner: NT AUTHORITY\SYSTEM                                     │
│                                                                 │
│  DACL (Discretionary Access Control List):                      │
│    1. NT AUTHORITY\SYSTEM        → Full Control (Allow)         │
│    2. BUILTIN\Administrators     → Read (Allow)                 │
│    3. (No other entries)                                        │
│                                                                 │
│  Integrity Label: System (0x4000)                               │
│                                                                 │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  STEP 4: Mandatory Integrity Check (MIC)                        │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  Process Integrity: Medium (0x2000)                             │
│  Object Integrity:  System (0x4000)                             │
│                                                                 │
│  Medium < System → READ ALLOWED, WRITE DENIED                   │
│  (Read access not blocked by integrity)                         │
│                                                                 │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  STEP 5: Discretionary Access Check (DACL)                      │
│  ────────────────────────────────────────────────────────────── │
│                                                                 │
│  Requested Access: FILE_READ_DATA                               │
│                                                                 │
│  Checking token SIDs against DACL:                              │
│    ✗ Bob (S-1-5-21-xxx-1001)              → No match in DACL    │
│    ✗ BUILTIN\Users                        → No match in DACL    │
│    ✗ Everyone                             → No match in DACL    │
│    ✗ NT AUTHORITY\Authenticated Users     → No match in DACL    │
│                                                                 │
│  None of Bob's SIDs match any DACL entries                      │
│                                                                 │
│  ─────────────────────────────────────────────────────────────  │
│                                                                 │
│  RESULT: ACCESS DENIED                                          │
│  ─────────────────────────────────────────────────────────────  │
│                                                                 │
│  Bob's token doesn't contain SYSTEM or Administrators SIDs      │
│  Therefore, no DACL entry grants read access                    │
│  The SAM file remains protected                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### What If Bob Were an Administrator?

Let's replay the scenario with Bob running an elevated process:

```
Modified Scenario:
  Process: notepad.exe (Run as Administrator)
  Token: High integrity, includes BUILTIN\Administrators

STEP 4: Mandatory Integrity Check
  Process Integrity: High (0x3000)
  Object Integrity: System (0x4000)
  High < System → READ ALLOWED ✓ (but WRITE still denied)

STEP 5: DACL Check
  Checking token SIDs:
    ✗ Bob → No match
    ✓ BUILTIN\Administrators → Matches entry #2 (Read access)
    
RESULT: ACCESS GRANTED (Read Only)
  Bob's elevated token contains Administrators group
  DACL grants Administrators read access
  Access succeeds, but only for reading (not writing/modifying)
```




## Real-World Security Scenarios

### Scenario 1: The UAC Split Token

When you log in as an administrator, Windows actually creates **two tokens**:

**Filtered Token (Medium Integrity):**

- Used for normal applications
- Administrators group marked as "deny-only" (disabled)
- Most privileges disabled
- Provides protection against accidental/malicious elevation

**Elevated Token (High Integrity):**

- Created when you click "Yes" on UAC prompt
- Administrators group fully enabled
- Powerful privileges available (SeDebugPrivilege, etc.)
- Limited to explicitly elevated processes

```
You (Admin) launch Chrome:
  → Uses filtered token
  → Medium integrity
  → Administrators group disabled
  → Can't modify system files despite being admin

You launch Regedit (elevated):
  → UAC prompt appears
  → Uses elevated token
  → High integrity
  → Administrators group enabled
  → Can modify system registry
```

This split-token design means being an administrator doesn't automatically mean running with full privileges - you must consciously elevate, making exploitation harder.










---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./memory.md" >}})
[|NEXT|]({{< ref "./pe.md" >}})