---
showTableOfContents: true
title: "Part 4B - Lab: SeDebugPrivilege"
type: "page"
---


## Goal
Let's answer one simple question: What can SeDebugPrivilege do, and where does it stop working?

You can find the source code for this lab [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part04/sedebug.go).

For a technical companion to the source code see [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part04/sedebug_guide.md).

## The Three Scenarios

We'll look at 3 different types of processes to determine whether or not we are able to access it.

So let's get the PIDs for these processes.

```powershell
# Run as Administrator
# Launch calculator
calc.exe

# Get all PIDs
Get-Process CalculatorApp | Select-Object Id    # Your process
Get-Process spoolsv | Select-Object Id          # System process
Get-Process smss | Select-Object Id             # Special system process (kernel-protected)
```


In my case the PIDs are as follows, yours will be different.

| Process | PID  |
| ------- | ---- |
| calc    | 4572 |
| spoolsv | 2460 |
| smss    | 432  |


## Execute

Let's run our experiment. Note that to run it you need to provide 2 flags:

```powershell
# Test WITHOUT SeDebugPrivilege
.\sedebug.exe -pid 1123 -sedebug=false

# Test WITH SeDebugPrivilege
.\sedebug.exe -pid 1123 -sedebug=true
```



### Scenario 1 - Your Own Process (Standard User)

Open **regular** (non-admin) PowerShell and let's run it without SeDebugPrivilege.


```powershell
 .\sedebug.exe -pid 4572 -sedebug=true
═══════════════════════════════════════════════════
    SeDebugPrivilege Lab: What Can It Do?
═══════════════════════════════════════════════════

[*] Running as Administrator: false
[*] Enabling SeDebugPrivilege...
✅ SeDebugPrivilege enabled!
[*] SeDebugPrivilege enabled: false

[*] Opening process 4572...
✅ Handle obtained: 0x198

[*] Testing memory access...
✅ SUCCESS: Can read process memory

═══════════════════════════════════════════════════
RESULT: Full access granted
═══════════════════════════════════════════════════

You can:
  • Read memory
  • Write memory
  • Inject code
  • Terminate process
```

As predicted, since this is our own process, and since it's not a protected process, we have full access without elevated privileges or `SeDebugPrivilege` set.

**Lesson:** You always have full access to your own processes.

---

### Scenario 2 - System Process (Need Admin + SeDebugPrivilege)

Let's first see what happens if we run it in a **regular** (non-admin) PowerShell with `SeDebugPrivilege` set to true:

```powershell
.\sedebug.exe -pid 2460 -sedebug=true
═══════════════════════════════════════════════════
    SeDebugPrivilege Lab: What Can It Do?
═══════════════════════════════════════════════════

[*] Running as Administrator: false
[*] Enabling SeDebugPrivilege...
✅ SeDebugPrivilege enabled!
[*] SeDebugPrivilege enabled: false

[*] Opening process 2460...
❌ FAILED: access denied

═══════════════════════════════════════════════════
RESULT: Access denied
═══════════════════════════════════════════════════
```

As we predicted, it fails. Despite having `SeDebugPrivilege` set to true, we also need elevated (Admin) privileges.

Let's now open an **Admin** PowerShell and try our luck with `SeDebugPrivilege` set to false:

```powershell
.\sedebug.exe -pid 2460 -sedebug=false
═══════════════════════════════════════════════════
    SeDebugPrivilege Lab: What Can It Do?
═══════════════════════════════════════════════════

[*] Running as Administrator: true
[*] Disabling SeDebugPrivilege...
✅ SeDebugPrivilege disabled!
[*] SeDebugPrivilege enabled: false

[*] Opening process 2460...
❌ FAILED: access denied

═══════════════════════════════════════════════════
RESULT: Access denied
═══════════════════════════════════════════════════
```


We fail again since of course we need BOTH `SeDebugPrivilege` set to true, and elevated (Admin) privileges.

```powershell
.\sedebug.exe -pid 2460 -sedebug=true
═══════════════════════════════════════════════════
    SeDebugPrivilege Lab: What Can It Do?
═══════════════════════════════════════════════════

[*] Running as Administrator: true
[*] Enabling SeDebugPrivilege...
✅ SeDebugPrivilege enabled!
[*] SeDebugPrivilege enabled: true

[*] Opening process 2460...
✅ Handle obtained: 0x1A0

[*] Testing memory access...
✅ SUCCESS: Can read process memory

═══════════════════════════════════════════════════
RESULT: Full access granted
═══════════════════════════════════════════════════

You can:
  • Read memory
  • Write memory
  • Inject code
  • Terminate process
```


Now it succeeds.


**Lesson:** To access SYSTEM processes we need both elevated privileges and SeDebugPrivilege

---

### Scenario 3 - Special System Process (SeDebugPrivilege Not Enough)

Still in **Admin** PowerShell:


```powershell
.\sedebug.exe -pid 432 -sedebug=true
═══════════════════════════════════════════════════
    SeDebugPrivilege Lab: What Can It Do?
═══════════════════════════════════════════════════

[*] Running as Administrator: true
[*] Enabling SeDebugPrivilege...
✅ SeDebugPrivilege enabled!
[*] SeDebugPrivilege enabled: true

[*] Opening process 432...
❌ FAILED: access denied

═══════════════════════════════════════════════════
RESULT: Access denied
═══════════════════════════════════════════════════
```


**Lesson:** Some processes have additional kernel-level protection. SeDebugPrivilege has limits.

---


## Summary

| Process Type               | Example            | User Level    | SeDebugPrivilege | Result        | Notes               |
| -------------------------- | ------------------ | ------------- | ---------------- | ------------- | ------------------- |
| **Your Own Process**       | calc.exe (4572)    | Regular User  | ❌ False          | ✅ **Success** | You own the process |
| **System Process**         | spoolsv.exe (2460) | Regular User  | ✅ True           | ❌ **Failed**  | Need elevation      |
| **System Process**         | spoolsv.exe (2460) | Administrator | ❌ False          | ❌ **Failed**  | Need privilege      |
| **System Process**         | spoolsv.exe (2460) | Administrator | ✅ True           | ✅ **Success** | Both required       |
| **Special System Process** | smss.exe (432)     | Administrator | ✅ True           | ❌ **Failed**  | Kernel protection   |



## The Big Picture: Understanding SeDebugPrivilege

### What This Lab Teaches

**SeDebugPrivilege is powerful, but it's not magic.** Here we saw what it offers, and what it's limits are by looking at 3 types of processes.

1. **Owner-based security:** You always have full control over your own processes, regardless of privileges. This is fundamental to how Windows works.
2. **Privileged access:** To interact with processes owned by other users or SYSTEM, you need **both** administrative rights **and** SeDebugPrivilege enabled. This is why SeDebugPrivilege is considered a critical security boundary - enabling it essentially allows an administrator to break into any normal process on the system.
3. **Kernel-level protection (Tier 3):** Even administrator + SeDebugPrivilege can't access everything. Processes like Session Manager (smss.exe) are created directly by the kernel and have additional protections. This is Windows' last line of defense for critical system components.




This is why SeDebugPrivilege is a high-value target, it gives us the ability to do things being an Administrator alone does not. We'll explore exactly what we can do from an offensive POV more in a future lab (hint: it involves `lsass`), for now I just wanted us to get a sense of what it has to offer, and what its limits are.













---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./securityA.md" >}})
[|NEXT|]({{< ref "./securityC.md" >}})