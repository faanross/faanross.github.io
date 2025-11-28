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




### Critical PEB Field #2: Ldr - The Module Database

The **Ldr** field (at PEB+0x18) points to a `PEB_LDR_DATA` structure that maintains **three separate linked lists** of all modules (DLLs and the main executable) loaded into the process address space. This is the process's internal directory of loaded code and is critical for understanding how dynamic linking works at runtime.

**The PEB_LDR_DATA structure:**

```go
type PEB_LDR_DATA struct {
    Length                          uint32      // Size of this structure
    Initialized                     uint32      // Boolean: loader initialized?
    SsHandle                        uintptr     // Reserved
    InLoadOrderModuleList           LIST_ENTRY  // ★ Modules ordered by load sequence
    InMemoryOrderModuleList         LIST_ENTRY  // ★ Modules ordered by base address
    InInitializationOrderModuleList LIST_ENTRY  // ★ Modules ordered by init sequence
    EntryInProgress                 uintptr     // Module currently being processed
    ShutdownInProgress              uint32      // Boolean: loader shutting down?
    ShutdownThreadId                uintptr     // Thread performing shutdown
}

type LIST_ENTRY struct {
    Flink uintptr  // Forward link (next entry)
    Blink uintptr  // Backward link (previous entry)
}
```

**Why three separate lists?** Different Windows subsystems need to walk modules in different orders:

| List Name | Order | Use Case |
|-----------|-------|----------|
| **InLoadOrderModuleList** | Load chronology | Debugging (shows initialization sequence) |
| **InMemoryOrderModuleList** | Memory address | Quick address-to-module lookups |
| **InInitializationOrderModuleList** | DllMain() call order | Dependency resolution during startup/shutdown |

**Each list contains LDR_DATA_TABLE_ENTRY structures:**

```go
type LDR_DATA_TABLE_ENTRY struct {
    InLoadOrderLinks           LIST_ENTRY      // Links for load-order list
    InMemoryOrderLinks         LIST_ENTRY      // Links for memory-order list
    InInitializationOrderLinks LIST_ENTRY      // Links for init-order list
    DllBase                    uintptr         // ★ Module base address in memory
    EntryPoint                 uintptr         // ★ Address of DllMain or entry point
    SizeOfImage                uint32          // ★ Total size of module in memory
    FullDllName                UNICODE_STRING  // Full path (C:\Windows\System32\...)
    BaseDllName                UNICODE_STRING  // ★ Filename only (kernel32.dll)
    Flags                      uint32          // Module flags
    LoadCount                  uint16          // Reference count
    TlsIndex                   uint16          // Thread Local Storage index
    HashLinks                  LIST_ENTRY      // Hash table links
    TimeDateStamp              uint32          // PE timestamp
    EntryPointActivationContext uintptr        // Activation context
    // ... additional fields in newer Windows versions
}

type UNICODE_STRING struct {
    Length        uint16   // Length in bytes (not including null terminator)
    MaximumLength uint16   // Allocated buffer size
    Buffer        uintptr  // Pointer to wide-character (UTF-16) string
}
```

**Walking the module list - complete implementation:**

```go
func EnumerateModulesViaPEB() {
    // Step 1: Get TEB address
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    
    // Step 2: Get PEB from TEB (x64 offset shown)
    peb := *(*uintptr)(unsafe.Pointer(teb + 0x60))
    
    // Step 3: Get Ldr from PEB
    // Ldr is at PEB+0x18
    ldr := *(*uintptr)(unsafe.Pointer(peb + 0x18))
    
    // Step 4: Get the head of InLoadOrderModuleList
    // InLoadOrderModuleList is at PEB_LDR_DATA+0x10
    moduleListHead := ldr + 0x10
    
    // Step 5: Get first entry (Flink of list head)
    currentEntry := *(*uintptr)(unsafe.Pointer(moduleListHead))
    
    fmt.Println("═══════════════════════════════════════════════════════════════")
    fmt.Println("                    LOADED MODULES (via PEB)")
    fmt.Println("═══════════════════════════════════════════════════════════════")
    fmt.Println("Base Address       Size       Entry Point    Name")
    fmt.Println("───────────────────────────────────────────────────────────────")
    
    // Step 6: Walk the circular linked list
    // Stop when we circle back to the head
    for currentEntry != moduleListHead {
        // Cast to LDR_DATA_TABLE_ENTRY
        // Note: currentEntry points to InLoadOrderLinks field, which is offset 0
        // in LDR_DATA_TABLE_ENTRY, so no adjustment needed
        entry := (*LDR_DATA_TABLE_ENTRY)(unsafe.Pointer(currentEntry))
        
        // Step 7: Extract module name from UNICODE_STRING
        nameBuffer := entry.BaseDllName.Buffer
        nameLength := entry.BaseDllName.Length / 2 // Divide by 2: UTF-16 to char count
        
        name := make([]uint16, nameLength)
        for i := uint16(0); i < nameLength; i++ {
            // Read each wide character (2 bytes)
            name[i] = *(*uint16)(unsafe.Pointer(nameBuffer + uintptr(i*2)))
        }
        
        // Step 8: Display module information
        fmt.Printf("0x%016X  0x%08X  0x%016X  %s\n",
            entry.DllBase,      // Where module loaded in memory
            entry.SizeOfImage,  // Total memory footprint
            entry.EntryPoint,   // DllMain address
            syscall.UTF16ToString(name))
        
        // Step 9: Move to next entry
        currentEntry = entry.InLoadOrderLinks.Flink
    }
    
    fmt.Println("═══════════════════════════════════════════════════════════════")
}
```

**Example output:**

```
═══════════════════════════════════════════════════════════════
                    LOADED MODULES (via PEB)
═══════════════════════════════════════════════════════════════
Base Address       Size       Entry Point    Name
───────────────────────────────────────────────────────────────
0x00007FF6C2A00000  0x00015000  0x00007FF6C2A05C30  malware.exe
0x00007FFDC3B00000  0x001F0000  0x0000000000000000  ntdll.dll
0x00007FFDC1A00000  0x000D3000  0x00007FFDC1A1D590  kernel32.dll
0x00007FFDC0F00000  0x0025B000  0x00007FFDC0F17B50  kernelbase.dll
0x00007FFDC2100000  0x00099000  0x00007FFDC2115E40  advapi32.dll
0x00007FFDC2D00000  0x00057000  0x00007FFDC2D10B20  msvcrt.dll
═══════════════════════════════════════════════════════════════
```

### Why Module Enumeration Matters: Offensive and Defensive Applications

**Offensive techniques using module enumeration:**

| Technique | Description | Why PEB Access? |
|-----------|-------------|-----------------|
| **Finding ntdll.dll base** | Required for direct syscalls to bypass EDR hooks | Avoids `GetModuleHandle()` API which may be hooked |
| **Manual GetProcAddress** | Locate exported functions without API calls | Stealthy - doesn't trigger module load events |
| **DLL hiding (unlinking)** | Remove module from PEB lists to hide presence | Rootkit technique - module disappears from tools |
| **Hook detection** | Compare memory with clean DLL on disk | Detect inline hooks placed by security products |
| **ASLR bypass** | Find randomized base addresses of system DLLs | Calculate gadget addresses for ROP chains |

**Example: Finding ntdll.dll for direct syscalls:**

```go
// Find ntdll.dll base address without calling GetModuleHandle()
func FindNtdllBase() uintptr {
    // ... [PEB walking code from above] ...
    
    for currentEntry != moduleListHead {
        entry := (*LDR_DATA_TABLE_ENTRY)(unsafe.Pointer(currentEntry))
        
        // Get module name
        nameBuffer := entry.BaseDllName.Buffer
        nameLength := entry.BaseDllName.Length / 2
        name := make([]uint16, nameLength)
        for i := uint16(0); i < nameLength; i++ {
            name[i] = *(*uint16)(unsafe.Pointer(nameBuffer + uintptr(i*2)))
        }
        moduleName := syscall.UTF16ToString(name)
        
        // Check if this is ntdll.dll (case-insensitive)
        if strings.EqualFold(moduleName, "ntdll.dll") {
            return entry.DllBase  // Return base address
        }
        
        currentEntry = entry.InLoadOrderLinks.Flink
    }
    
    return 0  // Not found
}

// Now use this base to resolve syscall numbers for direct invocation
// This bypasses userland hooks placed by EDR products
```

**Defensive perspective - detecting DLL unlinking:**

Security products can detect when malware tries to hide modules by unlinking them from PEB lists. While the DLL remains in memory and functional, it won't appear in the lists. Detection works by:

1. Walking the module lists and recording all entries
2. Scanning memory for PE headers to find hidden modules
3. Comparing the two sets to identify unlinked modules
4. Flagging discrepancies as suspicious behavior



### Critical PEB Field #3: ProcessParameters - Command Line and Environment

The **ProcessParameters** field at PEB+0x20 points to an `RTL_USER_PROCESS_PARAMETERS` structure containing process startup information. This is how Windows stores the command-line arguments, environment variables, current directory, and standard handles that were provided when the process was created.

**RTL_USER_PROCESS_PARAMETERS structure:**

```go
type RTL_USER_PROCESS_PARAMETERS struct {
    MaximumLength     uint32         // Allocated size of this structure
    Length            uint32         // Used size
    Flags             uint32         // Normalization flags
    DebugFlags        uint32         // Debug-related flags
    ConsoleHandle     uintptr        // Handle to console window
    ConsoleFlags      uint32         // Console mode flags
    StandardInput     uintptr        // ★ stdin handle
    StandardOutput    uintptr        // ★ stdout handle
    StandardError     uintptr        // ★ stderr handle
    CurrentDirectory  CURDIR         // Current working directory
    DllPath           UNICODE_STRING // DLL search path
    ImagePathName     UNICODE_STRING // ★ Full path to executable
    CommandLine       UNICODE_STRING // ★ Complete command line
    Environment       uintptr        // ★ Environment block pointer
    StartingX         uint32         // Window position X
    StartingY         uint32         // Window position Y
    CountX            uint32         // Window width
    CountY            uint32         // Window height
    CountCharsX       uint32         // Console buffer width
    CountCharsY       uint32         // Console buffer height
    FillAttribute     uint32         // Console text attributes
    WindowFlags       uint32         // Window style flags
    ShowWindowFlags   uint32         // SW_SHOW, SW_HIDE, etc.
    WindowTitle       UNICODE_STRING // Window title string
    DesktopInfo       UNICODE_STRING // Desktop name
    ShellInfo         UNICODE_STRING // Shell-specific data
    RuntimeData       UNICODE_STRING // Runtime environment data
    // ... more fields
}

type CURDIR struct {
    DosPath UNICODE_STRING  // Current directory path
    Handle  uintptr         // Handle to directory
}
```

**Extracting the command line directly from PEB:**

```go
func GetCommandLineFromPEB() string {
    // Get PEB address
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    peb := *(*uintptr)(unsafe.Pointer(teb + 0x60))
    
    // Get ProcessParameters pointer (PEB+0x20)
    processParams := *(*uintptr)(unsafe.Pointer(peb + 0x20))
    
    // CommandLine UNICODE_STRING offset in RTL_USER_PROCESS_PARAMETERS
    // On x64: typically at offset +0x70
    // On x86: typically at offset +0x40
    var cmdLineOffset uintptr
    if unsafe.Sizeof(uintptr(0)) == 8 {
        cmdLineOffset = 0x70  // x64
    } else {
        cmdLineOffset = 0x40  // x86
    }
    
    // Get pointer to CommandLine UNICODE_STRING
    cmdLine := (*UNICODE_STRING)(unsafe.Pointer(processParams + cmdLineOffset))
    
    // Extract the string data
    length := cmdLine.Length / 2  // Convert bytes to wide-char count
    buffer := make([]uint16, length)
    for i := uint16(0); i < length; i++ {
        buffer[i] = *(*uint16)(unsafe.Pointer(cmdLine.Buffer + uintptr(i*2)))
    }
    
    return syscall.UTF16ToString(buffer)
}

func main() {
    cmdLine := GetCommandLineFromPEB()
    fmt.Printf("Command line (from PEB): %s\n", cmdLine)
    
    // Compare with traditional API:
    // cmdLine := strings.Join(os.Args, " ")
    
    // OFFENSIVE USE CASES:
    // 1. Parse C2 server address from command-line argument
    //    Example: malware.exe --server=evil.com --key=abc123
    //    Stealth benefit: No config file on disk to discover
    
    // 2. Parse decryption key from parent process's command line
    //    Allows staged payloads without hardcoded keys
    
    // 3. Check if running with specific flags for anti-sandbox
    //    Example: Malware requires specific flag that sandbox won't provide
}
```

**Why we would want to use ProcessParameters instead of os.Args or GetCommandLine():**

1. **Stealth**: Avoids API calls that may be hooked or monitored by EDR
2. **Parent process inspection**: Can access parent's command line (with appropriate access)
3. **Rootkit techniques**: Can modify command line in memory to hide arguments
4. **Anti-forensics**: Understanding this helps hide command-line indicators

**Environment variable access from PEB:**

```go
// The Environment field points to a block of null-terminated strings
// Format: "VAR1=value1\0VAR2=value2\0\0"
func GetEnvironmentFromPEB() map[string]string {
    // ... [get processParams as above] ...
    
    // Environment is at offset +0x80 on x64
    envPtr := *(*uintptr)(unsafe.Pointer(processParams + 0x80))
    
    envVars := make(map[string]string)
    offset := uintptr(0)
    
    for {
        // Read wide-char string
        var wstr []uint16
        for {
            wchar := *(*uint16)(unsafe.Pointer(envPtr + offset))
            offset += 2
            if wchar == 0 {
                break  // End of this string
            }
            wstr = append(wstr, wchar)
        }
        
        if len(wstr) == 0 {
            break  // Double-null = end of environment block
        }
        
        envString := syscall.UTF16ToString(wstr)
        parts := strings.SplitN(envString, "=", 2)
        if len(parts) == 2 {
            envVars[parts[0]] = parts[1]
        }
    }
    
    return envVars
}
```

### The Thread Environment Block (TEB): Per-Thread Metadata

While the PEB is per-process, the **Thread Environment Block (TEB)** is **per-thread** - each thread in a process has its own TEB. The TEB contains thread-specific information like exception handlers, thread local storage, stack boundaries, and a pointer back to the owning process's PEB.

**TEB structure layout (simplified):**

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        TEB STRUCTURE (x64 Layout)                            │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  OFFSET  FIELD                         PURPOSE & SECURITY IMPLICATIONS       │
│  ──────────────────────────────────────────────────────────────────────────  │
│  +0x000  NtTib                         ★ NT_TIB (exception handling chain)   │
│  +0x018  EnvironmentPointer            Environment data pointer              │
│  +0x020  ClientId                      ★ CLIENT_ID (PID + TID)               │
│  +0x030  ActiveRpcHandle               Active RPC call handle                │
│  +0x038  ThreadLocalStoragePointer     ★ TLS array pointer                   │
│  +0x040  ProcessEnvironmentBlock       ★ Pointer to owning process's PEB     │
│  +0x048  LastErrorValue                ★ GetLastError() return value         │
│  +0x04C  CountOfOwnedCriticalSections  Number of owned critical sections     │
│  +0x050  CsrClientThread               CSRSS thread data                     │
│  +0x058  Win32ThreadInfo               Win32k thread information             │
│  +0x0C0  InstrumentationCallback       ★ Windows 10+ anti-debug callback     │
│  ...                                                                         │
│  +0x1788 LastStatusValue               ★ NTSTATUS from last Nt*() call       │
│  +0x1690 DeallocationStack             Stack deallocation address            │
│  +0x1698 ReservedForPerf               Performance monitoring data           │
│  ...                                                                         │
│                                                                              │
│  ★ = Security-relevant fields                                                │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**The NT_TIB: Structured Exception Handling (SEH) Chain**

The first field in the TEB is the **NT_TIB (NT Thread Information Block)**, which contains critical information about exception handling:

```go
type NT_TIB struct {
    ExceptionList       uintptr  // ★ SEH chain (linked list of exception handlers)
    StackBase           uintptr  // ★ Top of stack (high address)
    StackLimit          uintptr  // ★ Bottom of stack (low address)
    SubSystemTib        uintptr  // Subsystem-specific TIB
    FiberData           uintptr  // Fiber local storage or version
    ArbitraryUserPointer uintptr // User-defined pointer
    Self                uintptr  // Pointer to TEB itself
}
```

The **ExceptionList** field is particularly important - it's the head of a linked list of exception handlers that will be called if an exception occurs in this thread. This is a common target for exploits (SEH overwrite attacks) and anti-debugging tricks.

### Extracting Process and Thread IDs from TEB

One of the most useful TEB fields is **ClientId**, which contains both the process ID (PID) and thread ID (TID). Accessing these via TEB avoids API calls entirely:

```go
type CLIENT_ID struct {
    UniqueProcess uintptr  // Process ID (PID)
    UniqueThread  uintptr  // Thread ID (TID)
}

func GetPIDTIDFromTEB() (uint32, uint32) {
    // Get TEB address
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    
    // ClientId is at TEB+0x40 on x64, TEB+0x20 on x86
    var clientIdOffset uintptr
    if unsafe.Sizeof(uintptr(0)) == 8 {
        clientIdOffset = 0x40  // x64
    } else {
        clientIdOffset = 0x20  // x86
    }
    
    clientId := (*CLIENT_ID)(unsafe.Pointer(teb + clientIdOffset))
    
    return uint32(clientId.UniqueProcess), uint32(clientId.UniqueThread)
}

func main() {
    pid, tid := GetPIDTIDFromTEB()
    fmt.Printf("PID from TEB: %d\n", pid)
    fmt.Printf("TID from TEB: %d\n", tid)
    
    // Compare with traditional API:
    apiPid := syscall.Getpid()
    fmt.Printf("PID from API: %d\n", apiPid)
    
    // OFFENSIVE BENEFIT:
    // Getting PID/TID from TEB avoids calling GetCurrentProcessId()
    // or GetCurrentThreadId(), which may be:
    // 1. Hooked by security products
    // 2. Logged for behavioral analysis
    // 3. Used as detection heuristic (frequent API calls)
}
```

### Advanced TEB Technique: LastErrorValue Manipulation

The **LastErrorValue** field (TEB+0x68 on x64) stores the value returned by `GetLastError()`. Understanding this allows for interesting techniques:

```go
func SetLastErrorDirectly(errorCode uint32) {
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    
    // LastErrorValue at TEB+0x68 on x64
    lastErrorPtr := (*uint32)(unsafe.Pointer(teb + 0x68))
    *lastErrorPtr = errorCode
    
    // This is effectively what SetLastError() does internally
}

func GetLastErrorDirectly() uint32 {
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    
    lastErrorPtr := (*uint32)(unsafe.Pointer(teb + 0x68))
    return *lastErrorPtr
}

// OFFENSIVE USE:
// Malware can manipulate error codes to confuse analysts or bypass
// error-checking code without actually succeeding at the operation
```

### Windows 10+ Anti-Debug Enhancement: InstrumentationCallback

Modern Windows versions (10+) added the **InstrumentationCallback** field at TEB+0x1B8 (x64). This is a function pointer that, when set, gets called on every syscall transition. Security products and debuggers use this for low-level monitoring, but it can also be used for anti-debugging:

```go
// Check if instrumentation callback is set (might indicate debugging/monitoring)
func CheckInstrumentationCallback() bool {
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    procNtCurrentTeb := ntdll.NewProc("NtCurrentTeb")
    teb, _, _ := procNtCurrentTeb.Call()
    
    // InstrumentationCallback at TEB+0x1B8 on x64
    callbackPtr := *(*uintptr)(unsafe.Pointer(teb + 0x1B8))
    
    if callbackPtr != 0 {
        fmt.Println("⚠️  Instrumentation callback detected!")
        fmt.Printf("Callback address: 0x%X\n", callbackPtr)
        return true
    }
    
    return false
}
```

## PEB/TEB Offensive Techniques: Complete Reference

Here's a comprehensive table of techniques that leverage PEB and TEB for both offensive and defensive operations:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    PEB/TEB OFFENSIVE & DEFENSIVE TECHNIQUES                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  TECHNIQUE                  STRUCTURE   FIELD/OFFSET    SECURITY CONTEXT    │
│  ────────────────────────────────────────────────────────────────────────── │
│  Anti-Debug Detection       PEB         BeingDebugged   Detect debuggers    │
│  OS Version Fingerprint     PEB         OSMajorVersion  Targeted exploits   │
│  Find ntdll.dll Base        PEB         Ldr.Modules     Direct syscalls     │
│  Find kernel32.dll Base     PEB         Ldr.Modules     API resolution      │
│  Manual GetProcAddress      PEB         Ldr.Modules     Hook evasion        │
│  DLL Hiding (Unlinking)     PEB         Ldr.Modules     Rootkit technique   │
│  Hook Detection             PEB         Ldr.Modules     Security evasion    │
│  Extract Command Line       PEB         ProcessParams   C2 config parsing   │
│  Extract Environment Vars   PEB         ProcessParams   Credential theft    │
│  Get PID without API        TEB         ClientId        Stealth operation   │
│  Get TID without API        TEB         ClientId        Stealth operation   │
│  Custom GetLastError        TEB         LastErrorValue  Error manipulation  │
│  SEH Chain Walking          TEB         NtTib.ExceptionList  Exploit dev    │
│  Stack Bounds Check         TEB         NtTib.StackBase/Limit  Stack pivot  │
│  TLS Access                 TEB         ThreadLocalStoragePtr  Data hiding  │
│  Instrumentation Detection  TEB         InstrumentationCallback  Anti-debug │
│  NTSTATUS Reading           TEB         LastStatusValue Native API results  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Practical Example: Complete Stealth Module Enumeration

Combining multiple PEB/TEB techniques for a comprehensive stealth enumeration tool:

```go
package main

import (
    "fmt"
    "strings"
    "syscall"
    "unsafe"
)

func main() {
    fmt.Println("═══════════════════════════════════════════════════════════")
    fmt.Println("          STEALTH PROCESS INTROSPECTION TOOL")
    fmt.Println("          (No Windows APIs - Direct PEB/TEB Access)")
    fmt.Println("═══════════════════════════════════════════════════════════\n")
    
    // Get all info without API calls
    pid, tid := GetPIDTIDFromTEB()
    isDebugged := IsDebuggerPresent()
    cmdLine := GetCommandLineFromPEB()
    imageBase := GetImageBaseAddress()
    
    fmt.Printf("Process ID:        %d (from TEB)\n", pid)
    fmt.Printf("Thread ID:         %d (from TEB)\n", tid)
    fmt.Printf("Debugger Present:  %v (from PEB)\n", isDebugged)
    fmt.Printf("Image Base:        0x%X (from PEB)\n", imageBase)
    fmt.Printf("Command Line:      %s (from PEB)\n\n", cmdLine)
    
    // Enumerate modules
    fmt.Println("Loaded Modules:")
    EnumerateModulesViaPEB()
    
    fmt.Println("\n✓ All information gathered without calling Windows APIs")
    fmt.Println("✓ EDR/AV cannot hook what we never called")
}
```

This tool demonstrates the power of direct PEB/TEB access - entire process introspection without a single hooked API call, making it extremely difficult for security products to detect or intercept.

### Defensive Countermeasures and Detection

**For security researchers and defenders:**

- **Monitor PEB modification**: Tools that modify BeingDebugged or unlink DLLs are highly suspicious
- **Validate module lists**: Compare PEB module lists with memory scanning results to detect unlinking
- **Protect TEB/PEB pages**: Some security products mark these pages as read-only or monitored
- **Detect anomalous access**: Frequent direct TEB/PEB access from non-system code may indicate malware
- **Kernel callbacks**: Use kernel-mode PsSetLoadImageNotifyRoutine to get authoritative module loads

Understanding PEB and TEB is essential for both offensive security (malware development, exploit writing, EDR evasion) and defensive security (threat hunting, memory forensics, rootkit detection). These structures represent the bridge between user-mode convenience and kernel-mode authority, making them perpetual targets and tools in the security arms race.








---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./pe.md" >}})
[|NEXT|]({{< ref "./exercises.md" >}})