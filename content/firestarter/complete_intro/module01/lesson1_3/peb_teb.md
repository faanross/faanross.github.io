---
showTableOfContents: true
title: "Part 6 - PEB and TEB"
type: "page"
---



## The Process Environment Block: A User-Mode Information Repository

Every Windows process maintains a rich data structure in user-mode memory called the **Process Environment Block (PEB)**. Unlike kernel structures like EPROCESS that require elevated privileges to access, the PEB is deliberately placed in user-accessible memory, making it a treasure trove of process metadata that any thread can query directly. This accessibility makes the PEB both incredibly useful for legitimate programming and a prime target for malware, debuggers, and security tools.

The PEB serves as the process's **self-awareness mechanism** - it contains information about what modules are loaded, what the command line arguments were, whether a debugger is attached, and dozens of other environmental details.

### Why the PEB Exists: Bridging Kernel and User Space

Windows maintains a strict separation between kernel mode (Ring 0) and user mode (Ring 3). While the kernel maintains authoritative information about processes in structures like EPROCESS, user-mode code needs frequent access to certain process information. Rather than forcing expensive kernel transitions for every query, Windows maintains a **synchronized shadow copy** of select information in user-mode memory - this is the PEB.

**Key characteristics of the PEB:**

- Located in user-mode address space (readable without privilege escalation)
- One PEB per process (shared by all threads in that process)
- Maintained by the Windows loader and kernel
- Contains both static information (image base address) and dynamic information (loaded modules)
- Accessible via the Thread Environment Block (TEB) of any thread in the process

### PEB Structure Layout: Critical Fields for Security Research

The PEB is a complex structure with over 100 fields, but certain offsets are particularly important for security work. Here's the detailed layout with offensive/defensive implications:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        PEB STRUCTURE (x64 Layout)                            │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  OFFSET  FIELD                         PURPOSE & SECURITY IMPLICATIONS       │
│  ──────────────────────────────────────────────────────────────────────────  │
│  +0x000  InheritedAddressSpace        Reserved/unused in modern Windows      │
│  +0x001  ReadImageFileExecOptions     Read .exe compatibility flags          │
│  +0x002  BeingDebugged                 ★ CRITICAL: Debugger detection flag   │
│  +0x003  BitField                      Flags: ImageUsesLargePages, etc.      │
│  +0x008  Mutant                        Process initialization mutex          │
│  +0x010  ImageBaseAddress              ★ Base address where .exe loaded      │
│  +0x018  Ldr                           ★ Pointer to PEB_LDR_DATA (modules)   │
│  +0x020  ProcessParameters             ★ RTL_USER_PROCESS_PARAMETERS pointer │
│  +0x028  SubSystemData                 Subsystem-specific data               │
│  +0x030  ProcessHeap                   ★ Default heap for this process       │
│  +0x038  FastPebLock                   Critical section for PEB access       │
│  +0x040  AtlThunkSListPtr              ATL thunk SList pointer               │
│  +0x048  IFEOKey                       Image File Execution Options key      │
│  +0x050  CrossProcessFlags             Various cross-process flags           │
│  ...                                                                         │
│  +0x0BC  OSMajorVersion                ★ Windows major version (e.g., 10)    │
│  +0x0C0  OSMinorVersion                ★ Windows minor version               │
│  +0x0C4  OSBuildNumber                 ★ Windows build number (e.g., 19044)  │
│  +0x0C8  OSCSDVersion                  Service pack version                  │
│  +0x0CC  OSPlatformId                  Platform ID (VER_PLATFORM_WIN32_NT)   │
│  +0x0D0  ImageSubsystem                Subsystem type (GUI/CUI/NATIVE)       │
│  +0x0D4  ImageSubsystemMajorVersion    Subsystem version info                │
│  +0x0D8  ImageSubsystemMinorVersion                                          │
│  ...                                                                         │
│  +0x0EC  SessionId                     Terminal Services session ID          │
│                                                                              │
│  ★ = Frequently targeted by malware and security tools                       │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Important note on architecture differences:** The offsets shown above are for x64 (64-bit) Windows. On x86 (32-bit) systems, most offsets are different due to pointer size differences. For example, the PEB pointer in TEB is at offset +0x30 on x86 but +0x60 on x64. Always account for architecture when writing portable PEB-accessing code.




### Critical PEB Field #1: BeingDebugged - The Classic Anti-Debug Check

The **BeingDebugged** flag at offset +0x002 is a single byte that the Windows kernel sets to `1` when a debugger attaches to the process. This is the simplest and most well-known anti-debugging technique - often the first check you want to implement in your malware.

**Just as a reminder, why do we want to avoid debuggers?**
- **Security researchers** use debuggers to step through code, understand functionality, and develop signatures/detections
- **Automated sandboxes** (like VirusTotal, Any.Run, Cuckoo) often run samples under instrumentation
- **Incident responders** attach debuggers to understand what happened during a breach

**Common evasion responses we can implement if this is detected:**

- Exit immediately  -  analyst gets nothing useful
- Execute benign/decoy behaviour  -  waste analyst time, evade sandbox verdicts
- Corrupt or delete itself  -  anti-forensics
- Sleep or stall  -  timeout sandbox analysis windows (often 2-5 minutes)
- Crash intentionally  -  make analysis frustrating

**The broader category is "environment detection":**

| Check | What it detects |
|-------|-----------------|
| PEB.BeingDebugged | Debugger attached |
| VM artifacts (registry, MAC address) | Sandbox/VM |
| Low RAM/CPU count | Analysis VM (often minimal resources) |
| No recent files/documents | Freshly spun-up analysis machine |
| Mouse movement patterns | Automated vs human interaction |
| Process names (wireshark.exe, x64dbg.exe) | Analyst tools running |

The goal is distinguishing "real victim endpoint" from "analyst lab machine".

Now back to BeingDebugged...


**How BeingDebugged gets set:**

1. When you launch a process under a debugger (e.g., `windbg.exe target.exe`), the kernel sets this flag during process creation
2. When you attach to a running process, the `DebugActiveProcess()` API sets this flag
3. The flag remains set until the debugger detaches or the process terminates

**Implementation in Go:**

```go
// Anti-debug check using PEB.BeingDebugged
package main

import (
    "fmt"
    "syscall"
    "unsafe"
)

func IsDebuggerPresent() bool {
    // Get TEB (Thread Environment Block) address
    // NtCurrentTeb() is an intrinsic that returns gs:[0x30] on x64 or fs:[0x18] on x86
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    
    teb, _, _ := procNtCurrentTeb.Call()
    
    // PEB is referenced from TEB at different offsets based on architecture
    var pebOffset uintptr
    if unsafe.Sizeof(uintptr(0)) == 8 {
        pebOffset = 0x60 // x64: TEB+0x60 contains PEB pointer
    } else {
        pebOffset = 0x30 // x86: TEB+0x30 contains PEB pointer
    }
    
    // Dereference to get PEB address
    peb := *(*uintptr)(unsafe.Pointer(teb + pebOffset))
    
    // BeingDebugged is a BOOLEAN (byte) at PEB+0x02
    beingDebugged := *(*byte)(unsafe.Pointer(peb + 0x02))
    
    return beingDebugged != 0
}

func main() {
    if IsDebuggerPresent() {
        fmt.Println("⚠️  Debugger detected via PEB.BeingDebugged!")
        // Real malware might:
        // - Exit gracefully to avoid analysis
        // - Execute fake/decoy code path
        // - Trigger anti-forensics (delete artifacts)
        // - Enter infinite loop to waste analyst time
    } else {
        fmt.Println("✓ No debugger detected")
        // Proceed with malicious payload
    }
}
```


**What the code does:**

1. Find the thread's metadata block (TEB)
2. From there, find the process's metadata block (PEB)
3. Read the `BeingDebugged` byte - as mentikoned Windows sets this to 1 when a debugger attaches
4. React accordingly (in real malware: hide, exit, or behave differently)

That's it we are just reading a byte from a known location in memory that the OS conveniently maintains for us.

**Defensive perspective - bypassing this check:**

- **Manual flag clearing:** Debuggers can write `0` to `PEB+0x02` after attaching
- **Breakpoint on access:** Set hardware breakpoint on PEB+0x02 and return false value
- **API hooking:** Hook `NtQueryInformationProcess` which can also reveal debugging
- **Kernel-mode debugging:** Kernel debuggers (WinDbg in kernel mode) don't set this flag since they operate below the user-mode detection layer

**Why this remains effective:** Despite being well-known, this check still appears in malware because it's trivial to implement and catches careless analysts who forget to bypass it.







---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./pe.md" >}})
[|NEXT|]({{< ref "./exercises.md" >}})