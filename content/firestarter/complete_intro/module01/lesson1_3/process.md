---
showTableOfContents: true
title: "Part 1 - Process Architecture"
type: "page"
---

# **LESSON 1.3: WINDOWS INTERNALS REVIEW FOR OFFENSIVE OPERATIONS**

---

## **Understanding Your Battlefield**

You've chosen Go as your weapon. You understand the offensive tooling landscape and the language trade-offs. Now you must master your battlefield: **Windows**.

Every offensive technique you'll learn - process injection, privilege escalation, evasion, persistence - requires deep understanding of Windows internals. You can't manipulate what you don't understand. You can't evade detection if you don't know what defenders monitor. You can't exploit a system whose architecture is a mystery.

This lesson isn't about memorizing facts. It's about building a **mental model** of how Windows actually works under the hood - the architecture that Microsoft's documentation glosses over, the internal structures that offensive developers exploit, the mechanisms that both enable and constrain your operations.

By the end of this lesson, you will:

- **Understand process architecture** at a level that enables injection techniques
- **Navigate memory management** including virtual memory, VAD trees, and protections
- **Comprehend the security model** - tokens, privileges, integrity levels, and their exploitation
- **Parse PE file structures** to manipulate executables in memory
- **Abuse PEB/TEB structures** for evasion and information gathering
- **Recognize what defenders monitor** and how to operate beneath their sensors

This is foundational knowledge. Every subsequent module builds on these concepts. Let's begin.

---

## **PART 1:  PROCESS ARCHITECTURE**

### **The Process: Windows' Fundamental Execution Container**

A Windows process is fundamentally a container - a protective boundary that holds resources, memory, and security context. While threads perform the actual execution, the process provides the environment in which they operate, isolating each application from others and managing the resources they can access.

For a visual overview see this [video](https://www.youtube.com/watch?v=LAnWQFQmgvI).

```  
┌──────────────────────────────────────────────────────────────┐  
│                    WINDOWS PROCESS ANATOMY                   │  
├──────────────────────────────────────────────────────────────┤  
│                                                              │  
│  PROCESS COMPONENTS:                                         │  
│                                                              │  
│  1. EXECUTIVE PROCESS (EPROCESS)                             │  
│     • Kernel-mode structure                                  │  
│     • Process ID (PID)                                       │  
│     • Parent process ID (PPID)                               │  
│     • Token (security context)                               │  
│     • Handle table                                           │  
│     • VAD tree (memory mappings)                             │  
│                                                              │  
│  2. VIRTUAL ADDRESS SPACE                                    │  
│     • 0x00000000 - 0x7FFFFFFF: User space (2GB/4GB*)         │  
│     • 0x80000000 - 0xFFFFFFFF: Kernel space (2GB/4GB*)       │  
│     • *On 64-bit: User = 0x000 - 0x7FF..., much larger       │  
│     • Private, isolated per process                          │  
│                                                              │  
│  3. PRIMARY TOKEN                                            │  
│     • User SID (Security Identifier)                         │  
│     • Group memberships                                      │  
│     • Privileges (SeDebugPrivilege, etc.)                    │  
│     • Integrity level (Low/Medium/High/System)               │  
│                                                              │  
│  4. HANDLE TABLE                                             │  
│     • References to kernel objects                           │  
│     • Files, registry keys, processes, threads               │  
│     • Each handle has access rights                          │  
│                                                              │  
│  5. PEB (Process Environment Block)                          │  
│     • User-mode structure (visible to process)               │  
│     • Module list (loaded DLLs)                              │  
│     • Command line parameters                                │  
│     • Environment variables                                  │  
│                                                              │  
│  6. THREADS                                                  │  
│     • At least one (primary thread)                          │  
│     • Each has own stack and TEB                             │  
│     • Share process address space                            │  
│                                                              │  
└──────────────────────────────────────────────────────────────┘  
```  




#### 1. Executive Process (EPROCESS)

The **EPROCESS** is the kernel's master record for a process - a large data structure maintained in kernel memory that contains everything the operating system needs to manage and track the process.

- **Kernel-mode structure**: This exists in protected kernel memory, invisible and inaccessible to the process itself - only the OS can read or modify it.
- **Process ID (PID)**: A unique numerical identifier that distinguishes this process from all others currently running on the system.
- **Parent process ID (PPID)**: Records which process spawned this one, creating a family tree of processes useful for tracking relationships and inheritance.
- **Token (security context)**: A pointer to the security token that determines what this process is allowed to do - what files it can access, what privileges it holds.
- **Handle table**: A process-private table that maps handle values (like file handles) to actual kernel objects, allowing the process to reference system resources.
- **VAD tree (Virtual Address Descriptor tree)**: A data structure tracking all memory regions allocated to the process - which addresses are valid, what protections they have, and what they're mapped to.

#### 2. Virtual Address Space

Every process receives its own private virtual address space - an illusion of having the entire memory range to itself, even though physical RAM is shared among all processes. Remember, parts of the VAS that are not actively being used can also be mapped to disk (SWAP).

- **User space (0x00000000 - 0x7FFFFFFF on 32-bit)**: This is where the process's code, data, heap, and stacks live; the process can freely access this region.
- **Kernel space (0x80000000 - 0xFFFFFFFF on 32-bit)**: Reserved for the operating system kernel; attempting to access these addresses from user mode triggers an access violation.
- **64-bit expansion**: On 64-bit systems, user space extends to 128TB (`0x00000000` - `0x00007FFF'FFFFFFFF`), providing vastly more virtual memory for large applications.
- **Private and isolated**: Each process's address space is separate; a pointer to address `0x00400000` in one process refers to completely different physical memory than the same address in another process.


#### 3. Primary Token
The **primary token** is the process's security badge - it defines the security identity under which the process runs and what actions it's authorized to perform.

- **User SID (Security Identifier)**: Identifies which user account owns this process, forming the basis of access control decisions throughout Windows.
- **Group memberships**: Lists all security groups the user belongs to (Administrators, Users, etc.), which collectively determine permissions.
- **Privileges**: Special rights that override normal security checks, like `SeDebugPrivilege` (attach to any process) or `SeBackupPrivilege` (bypass file security for backups).
- **Integrity level**: A mandatory access control layer where Low-integrity processes (like sandboxed browsers) cannot modify resources owned by Medium or High-integrity processes, preventing privilege escalation.

#### 4. Handle Table
The **handle table** is the process's directory of system resources - a mapping between small integer handles and actual kernel objects that the process can use.

- **References to kernel objects**: Handles are indirect references; instead of raw pointers, processes use handles which the kernel translates to actual object addresses.
- **Resource variety**: Handles can refer to diverse objects - open files, registry keys, synchronization primitives (mutexes, events), other processes or threads, and more.
- **Access rights per handle**: Each handle carries its own permission mask; a process might have read-only access to one file handle and read-write access to another.

#### 5. PEB (Process Environment Block)
The **PEB** is a user-mode data structure that lives in the process's own address space, providing the process with information about itself and its environment.

- **User-mode accessibility**: Unlike `EPROCESS`, the PEB resides in user space where the process can directly read it without kernel transitions.
- **Module list (loaded DLLs)**: Contains linked lists of all loaded modules (EXEs and DLLs), their base addresses, and names - essential for dynamic linking and introspection.
- **Command line parameters**: Stores the full command line that launched the process, accessible via standard APIs like `GetCommandLine()`.
- **Environment variables**: A block of null-terminated strings containing environment variables (`PATH`, `TEMP`, `USERNAME`) inherited from the parent process or set at creation.

#### 6. Threads
While a process owns resources, **threads** are what actually execute code - they're the workers operating within the process's environment.

- **At least one (primary thread)**: Every process begins with one thread created automatically at process startup; the process lives as long as at least one thread remains.
- **Each has own stack and TEB**: Threads need private stacks for function calls and Thread Environment Blocks for thread-specific data, but these are allocated within the process's shared address space.
- **Share process address space**: All threads in a process see the same memory - they can access the same global variables, heap allocations, and code, which enables easy communication but requires careful synchronization.


### Process Memory Organization Layout

```  
USER MODE (Ring 3)  
┌─────────────────────────────────────────────────────────┐  
│                                                         │  
│  0x00400000   PE Image (notepad.exe)                    │  
│               ├─ .text (code)                           │  
│               ├─ .data (initialized data)               │  
│               └─ .rdata (read-only)                     │  
│                                                         │  
│  0x76D00000   ntdll.dll                                 │  
│  0x77000000   kernel32.dll                              │  
│  0x75000000   kernelbase.dll                            │  
│               ... other DLLs ...                        │  
│                                                         │  
│  0x00200000   Heap (dynamic allocations)                │  
│  0x00100000   Stack (thread 1)                          │  
│  0x00110000   Stack (thread 2)                          │  
│                                                         │  
│  0x7FFE0000   Shared User Data (KUSER_SHARED_DATA)      │  
│  0x7FFD0000   PEB (Process Environment Block)           │  
│                                                         │  
├─────────────────────────────────────────────────────────┤  
│               0x7FFFFFFF (User/Kernel boundary)         │  
├─────────────────────────────────────────────────────────┤  
│                                                         │  
│                 KERNEL MODE (Ring 0)                    │  
│                                                         │  
│  0x80000000   System code, drivers, kernel              │  
│               (Not directly accessible from user mode)  │  
│                                                         │  
└─────────────────────────────────────────────────────────┘  
  
```  

**Note on 64-bit systems**: This 32-bit layout shows a 4GB address space split between user and kernel. On 64-bit Windows, user space extends to 128TB (addresses 0x000000000000 to 0x00007FFFFFFFFFFF), providing vastly more room for large applications, while kernel space occupies the upper half starting at 0xFFFF800000000000 - both regions scaled dramatically to take advantage of 64-bit addressing.

When you launch an application, say `notepad.exe`, Windows creates a virtual address space - a private, isolated memory environment where the process lives. This diagram above shows how that address space is organized, from the executable code through to system DLLs and dynamic memory, all the way to the boundary where kernel space begins.

#### USER MODE (Ring 3)

User mode is where application code executes with restricted privileges, unable to directly access hardware or critical system resources. Ring 3 refers to the CPU's protection level - the least privileged ring where most code runs, protected from accidentally or maliciously damaging the system.

##### PE Image (0x00400000 - notepad.exe)

This is where the executable file itself gets loaded into memory - the starting point of the application. The base address 0x00400000 is the traditional default load address for Windows executables, though modern systems often randomize this for security (ASLR - Address Space Layout Randomization).

**• .text (code section)**: Contains the actual machine code instructions that the CPU executes - the compiled logic of your program marked as read-only and executable.

**• .data (initialized data)**: Holds global and static variables that have initial values defined in the executable, such as string constants or pre-configured settings; this section is readable and writable.

**• .rdata (read-only data)**: Stores constant data that should never change during execution, like string literals, import tables, and other immutable program data protected from accidental modification.


##### System DLLs

These are Windows system libraries that provide essential functionality to applications - they're loaded into every process's address space and contain thousands of functions that programs rely on.

- **ntdll.dll (0x76D00000)**: The lowest-level user-mode library that contains the actual system call stubs - every interaction with the kernel ultimately goes through ntdll, making it the bridge between user mode and kernel mode.
- **kernel32.dll (0x77000000)**: Provides the classic Win32 API functions for file operations, process management, memory allocation, and more; it's a wrapper around the lower-level kernelbase and ntdll functions.
- **kernelbase.dll (0x75000000)**: Contains the core implementation of many kernel32 functions, separated out in modern Windows versions to reduce code duplication and improve modularity.
- **Other DLLs**: Applications load additional libraries as needed - graphics libraries, networking components, UI frameworks - each mapped into the process's address space at their own base addresses.

##### Heap (0x00200000)

The **heap** is a region of memory used for dynamic allocations - whenever your code calls `malloc()`, `new`, or `HeapAlloc()`, memory comes from here. It grows and shrinks as the process allocates and frees memory, managed by the heap allocator which tracks free blocks and handles fragmentation.

##### Thread Stacks

Each thread in the process needs its own **stack** - a private region of memory used for function call frames, local variables, and return addresses.

- **Stack (thread 1) at 0x00100000**: The primary thread's stack, typically reserved as 1MB of virtual address space though only a small portion is initially committed to physical memory.
- **Stack (thread 2) at 0x00110000**: Additional threads get their own stacks at different addresses, allowing them to make function calls independently without interfering with each other's call chains.

##### System Data Structures

Near the top of user space, Windows places some special structures that provide process-level information and shared data.

- **KUSER_SHARED_DATA (0x7FFE0000)**: A special read-only page shared across all processes that contains frequently accessed system information like the current time, system version, and CPU features; mapping it into every process avoids expensive system calls for simple queries.
- **PEB - Process Environment Block (0x7FFD0000)**: The process's self-description structure containing the command line, environment variables, loaded module list, and other metadata that the process can query about itself without entering kernel mode.

#### The User/Kernel Boundary (0x7FFFFFFF)

This address marks the **dividing line** between user space and kernel space - a hard security barrier enforced by the CPU's memory management unit. Any attempt by user-mode code to access addresses above this boundary triggers an access violation, protecting the kernel from malicious or buggy application code.

#### KERNEL MODE (Ring 0)

Beyond the boundary lies **kernel space**, where the Windows kernel, device drivers, and system code execute with full hardware privileges. Ring 0 is the CPU's most privileged protection level, able to execute any instruction and access any memory.

- **System code (0x80000000 and above)**: This region contains the kernel itself (`ntoskrnl.exe`), device drivers, the HAL (Hardware Abstraction Layer), and kernel-mode system services. User-mode processes cannot directly read or write this memory; they must use system calls to request kernel services, and the kernel carefully validates all requests before executing them in this privileged space.

---




### **Process Creation: What Actually Happens**

When you execute a program, Windows performs complex orchestration:

```  
PROCESS CREATION STAGES:  
  
Stage 1: USER SPACE INITIALIZATION (kernel32!CreateProcess)  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  
1. Validate parameters (image path, command line)  
2. Open executable file  
3. Create initial process object (suspended)  
4. Create section object (memory-mapped file)  
  
Stage 2: KERNEL INITIALIZATION (ntdll!NtCreateUserProcess)  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  
5. Create EPROCESS structure  
6. Assign Process ID (PID)  
7. Create initial token (inherit or new)  
8. Initialize virtual address space  
9. Map ntdll.dll into new process  
10. Create PEB  
  
Stage 3: IMAGE LOADING  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  
11. Map PE image at preferred base address  
12. Resolve imports (load required DLLs)  
13. Apply relocations if needed  
14. Execute TLS callbacks  
15. Set up exception handlers  
  
Stage 4: THREAD CREATION  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  
16. Create primary thread  
17. Initialize thread stack  
18. Create TEB (Thread Environment Block)  
19. Set entry point to ntdll!LdrInitializeThunk  
  
Stage 5: EXECUTION  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  
20. LdrInitializeThunk runs  
21. Loader resolves all DLL dependencies  
22. Calls DllMain for each DLL  
23. Transfers control to entry point (main/WinMain)  
24. User code executes!  
```  


When you double-click an executable or launch a program from the command line, Windows embarks on a remarkably complex orchestration to transform a static file on disk into a living, executing process. This journey involves coordination between user-mode libraries, the kernel, the memory manager, and the loader - each playing a critical role in bringing your application to life.

#### Stage 1: User Space Initialization

1. The process begins in user space when an application calls **CreateProcess** (or one of its variants like `CreateProcessW`). This Win32 API function, residing in **kernel32.dll**, serves as the high-level entry point that most applications use to spawn new processes. Before handing control to the kernel, `CreateProcess` performs essential preliminary work: it **validates parameters** such as the **image path** (ensuring the executable exists and is accessible) and the **command line** arguments that will be passed to the new process.
2. The function then **opens the executable file** to verify it's a valid PE (Portable Executable) format and obtains a file handle.
3. At this point, `CreateProcess` **creates an initial process object** in a **suspended** state - the process exists but isn't yet running.
4. It then goes and creates a **section object**, which is Windows' abstraction for a **memory-mapped file** that will allow the executable to be efficiently loaded into memory without reading the entire file at once.

#### Stage 2: Kernel Initialization

5. The real heavy lifting begins when the request transitions into kernel mode through **NtCreateUserProcess**, the native system call interface in **ntdll.dll**. This is where the operating system kernel takes control and begins constructing the fundamental data structures that define a process. The kernel first **creates the EPROCESS structure**, the master control block that contains everything the OS needs to manage this process - its memory mappings, security context, handle table, and scheduling information.
6. The kernel then **assigns a Process ID (PID)**, a unique identifier that distinguishes this process from all others in the system.
7. Next comes security: the kernel **creates the initial token**, either by **inheriting** the security context from the parent process or creating a **new** one with different privileges if requested.
8. The kernel then **initializes the virtual address space**, setting up the page tables that will translate virtual addresses to physical memory and establishing the boundary between user and kernel space.
9. One of the first mappings in this new address space is critical: the kernel **maps ntdll.dll into the new process**, ensuring the lowest-level user-mode library is present for subsequent initialization.
10. Finally, the kernel **creates the PEB** (Process Environment Block), populating it with the command line, environment variables, and other process metadata.

#### Stage 3: Image Loading

11. With the process skeleton in place, Windows must now load the actual executable code and its dependencies. The kernel **maps the PE image** (the executable file itself) into the process's address space, ideally at its **preferred base address** - typically 0x00400000 for 32-bit executables, though Address Space Layout Randomization (ASLR) may choose a different location for security.
12. The PE file format contains an import table listing all the functions the executable needs from system DLLs, so the loader must **resolve imports** by **loading the required DLLs** (like kernel32.dll, user32.dll) into memory and fixing up the import address table to point to the actual function locations.
13. If the executable couldn't be loaded at its preferred base address - perhaps because another DLL already occupies that space - the loader must **apply relocations**, modifying hard-coded addresses throughout the code to account for the new base address.
14. Before the main code runs, the loader **executes TLS callbacks** (Thread Local Storage initialization routines that some executables register to run before the entry point).
15. Finally, the loader **sets up exception handlers**, establishing the chain of structured exception handling that will catch crashes and errors during execution.

#### Stage 4: Thread Creation

16. A process without threads is like a stage without actors - it holds resources but performs no work. Windows now **creates the primary thread**, the initial thread of execution that will bootstrap the process.
17. The system **initializes the thread stack**, reserving virtual address space (typically 1MB) and committing the initial pages needed for function calls.
18. Each thread needs its own metadata, so the kernel **creates the TEB** (Thread Environment Block), a user-mode structure containing thread-specific information like the thread ID, exception handling chain, and pointers to thread-local storage.
19. Interestingly, the thread's **entry point** isn't set directly to your program's `main()` function - instead, it's set to `ntdll!LdrInitializeThunk`, a special initialization function that will perform critical loader operations before your code ever runs.

#### Stage 5: Execution

20. The moment of truth arrives as the new thread begins executing. **LdrInitializeThunk runs** first, serving as the process's true starting point. This function in ntdll acts as the loader's coordinator, taking control before any application code executes.
21. The **loader resolves all DLL dependencies**, walking the import tables of the main executable and every DLL, recursively loading any additional libraries needed (if user32.dll needs gdi32.dll, it loads that too, and so on).
22. For each loaded library, the loader **calls DllMain**, the initialization function that every DLL can optionally implement to perform setup when it's loaded into a process - this is where DLLs allocate resources, initialize global state, and prepare for use.
23. Finally, with all dependencies satisfied and all DLLs initialized, the loader **transfers control to the entry point** - this is typically **main() for console applications or WinMain() for GUI applications**, the function you wrote that defines what your program actually does.
24. At last, **user code executes** - your program springs to life, completely unaware of the intricate dance that just occurred to bring it into existence.



#### **Why This Matters for Offensive Operations:**

Each stage is an **opportunity for manipulation**, for example:
- **Stage 1-2**: Process hollowing creates process then replaces image
- **Stage 3**: DLL injection hijacks import resolution
- **Stage 4**: Thread hijacking modifies entry point
- **Stage 5**: Reflective loading bypasses normal loader



### **The EPROCESS Structure**

As mentioned above, `EPROCESS` is the kernel's representation of a process, which is represented as a struct (simplified below) :

```c  
// Partial EPROCESS structure (varies by Windows version)
typedef struct _EPROCESS {
    KPROCESS         Pcb;                    // Process Control Block
    PVOID            UniqueProcessId;        // PID
    LIST_ENTRY       ActiveProcessLinks;     // Linked list of processes
    PVOID            Token;                  // Security token pointer
    PVOID            ObjectTable;            // Handle table
    PVOID            SectionObject;          // PE image section
    PVOID            VadRoot;                // Virtual Address Descriptor tree
    ULONG            SessionId;              // Terminal Services session
    CHAR             ImageFileName[16];      // Process name (e.g., "notepad.exe")
    // ... many more fields
} EPROCESS, *PEPROCESS;
```  


Like many of the other keystone data structures, we need to slowly develop familiarity with this struct over time as it contains a trove of offensive opportunities waiting to be exploited.

```  
EPROCESS Field          | Offensive Use  
─────────────────────────────────────────────────────────────  
Token                   | Token theft/impersonation  
ActiveProcessLinks      | Process hiding (unlink from list)  
VadRoot                 | Finding injected code in memory  
ImageFileName           | Process masquerading detection  
ObjectTable             | Handle duplication attacks  
UniqueProcessId         | Target process selection  
```  



#### **Accessing EPROCESS from User Mode:**

As mentioned above, we can't directly access `EPROCESS` from userland since it resides in kernel memory. We can however access the same process info as follows using Go.

**NOTE**: If you are developing a Windows-based application on MacOS/Linux using an IDE like Goland, you'll have to add "Windows build tags" to the top of your file to repress errors.

```go  
//go:build windows  
```  

If you're developing directly on a Windows system, you can ignore this.

```go  
//go:build windows

// Get process information via Windows API
package main

import (
    "flag"
    "fmt"
    "os"
    "syscall"
)

var (
    kernel32              = syscall.NewLazyDLL("kernel32.dll")
    procOpenProcess       = kernel32.NewProc("OpenProcess")
    procGetProcessId      = kernel32.NewProc("GetProcessId")
    procGetCurrentProcess = kernel32.NewProc("GetCurrentProcess")
)

// OpenProcess opens a handle to an existing process.
func OpenProcess(desiredAccess uint32, inheritHandle bool, processId uint32) (syscall.Handle, error) {
    inherit := 0
    if inheritHandle {
        inherit = 1
    }

    handle, _, err := procOpenProcess.Call(
        uintptr(desiredAccess),
        uintptr(inherit),
        uintptr(processId),
    )

    // A zero handle indicates failure.
    if handle == 0 {
        return 0, err
    }
    return syscall.Handle(handle), nil
}

func main() {
    // 1. Define an integer flag '-pid' to accept the target process ID.
    targetPid := flag.Int("pid", 0, "The Process ID of the target process.")
    flag.Parse()

    // 2. Validate that a PID was provided.
    if *targetPid == 0 {
        fmt.Println("Error: A target Process ID must be provided with the -pid flag.")
        flag.Usage() // Prints the default usage message.
        os.Exit(1)
    }

    // Get handle to the current running process
    currentProc, _, _ := procGetCurrentProcess.Call()

    // Get the PID of our own process
    pid, _, _ := procGetProcessId.Call(currentProc)

    fmt.Printf("Current Process ID: %d\n", pid)
    fmt.Printf("Attempting to open process with PID: %d\n", *targetPid)

    // 3. Use the PID from the flag in the OpenProcess call.
    // PROCESS_QUERY_INFORMATION (0x0400) allows querying information about the process.
    handle, err := OpenProcess(0x0400, false, uint32(*targetPid))
    if err != nil {
        // The error will often be "Access is denied." if you don't have sufficient privileges.
        fmt.Printf("Failed to open process %d: %v\n", *targetPid, err)
        return
    }
    defer syscall.CloseHandle(handle)

    fmt.Printf("Successfully opened handle for process %d: 0x%X\n", *targetPid, handle)
}

```  

- Compile the program and copy it over to the target (Windows) host.
- Open some process, for example `notepad.exe`, and use System Informer or Task Manager to find the PID of the target process, in my case it's `10980`.
- Open `ps.exe` or `cmd.exe` as Admin and run the following:

```shell  
.\eprocess.exe -pid 10980
	Current Process ID: 11292
	Attempting to open process with PID: 10980
	Successfully opened handle for process 10980: 0x160
```  


**We can see our application is capable of:**
- Determining and reporting its own PID  (`11292` in this case)
- Given a target process PID, getting a handle to the process. In this case we obtain a handle to `notepad.exe`, which can then be used as an argument for numerous other functions we'll learn about in future lessons.

Though both of these acts seem quite trivial, they'll be involved in many different more advanced techniques, so it's worth taking 1-2 minutes to review the code to ensure you understand how it works. Consider these type of actions "maldev table stakes" - in most circumstances it's the bare minimum you'll have to do to do anything else.






---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_2/conclusion.md" >}})
[|NEXT|]({{< ref "./thread.md" >}})