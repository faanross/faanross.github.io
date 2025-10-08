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



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_2/conclusion.md" >}})
[|NEXT|]({{< ref "./thread.md" >}})