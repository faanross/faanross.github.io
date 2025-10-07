---
showTableOfContents: true
title: "Part 1 - Process & Thread Architecture"
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
## **PART 1: PROCESS AND THREAD ARCHITECTURE**

### **The Process: Windows' Fundamental Execution Container**

A process is more than just "a running program." It's a complex container of resources, security context, and execution state.

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


**Visual Process Layout:**

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

Note: 64-bit layout different, much larger address space
```



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_2/conclusion.md" >}})
[|NEXT|]({{< ref "../../moc.md" >}})