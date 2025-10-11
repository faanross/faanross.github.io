---
showTableOfContents: true
title: "Part 3 - Virtual Memory Management"
type: "page"
---



## **The Virtual Memory Abstraction**

One of the most elegant and fundamental concepts in modern operating systems is **virtual memory** - an abstraction that creates a powerful illusion for every running process. Each process believes it has its own **private** and complete **address space**: a full **4GB on 32-bit systems** or an enormous **128TB on 64-bit systems**. This is a carefully maintained fiction, because in reality, **physical RAM is shared** among all processes, with the operating system and hardware conspiring to make each process believe it owns the entire memory space.

### Memory Layout Diagram

```
VIRTUAL ADDRESS SPACE (What Processes See):

Process A:                          Process B:
┌─────────────────────────────┐    ┌─────────────────────────────┐
│ 0x00400000 - A.exe Code     │    │ 0x00400000 - B.exe Code     │
│ 0x10000000 - Heap (private) │    │ 0x10000000 - Heap (private) │
│ 0x20000000 - Data (paged)   │◄─┐ │ 0x20000000 - Data (paged)   │◄─┐
│ 0x76D00000 - kernel32.dll   │  │ │ 0x76D00000 - kernel32.dll   │  │
│              (shared!)      │──┼─┼─┐(same DLL, same address!)  │  │
│ 0x7FF00000 - Stack          │  │ │ │0x7FF00000 - Stack         │  │
└─────────────────────────────┘  │ └─┼───────────────────────────┘  │
                                 │   │                              │
        ┌────────────────────────┘   │    ┌─────────────────────────┘
        │                            │    │
        ▼                            ▼    ▼
┌────────────────────────────────────────────────────────────────────┐
│                    PHYSICAL RAM (Actual Memory)                    │
├────────────────────────────────────────────────────────────────────┤
│ 0x00100000 - Process A Code (A.exe)                                │
│ 0x00500000 - Process B Code (B.exe)                                │
│ 0x00800000 - kernel32.dll (SHARED - mapped to both processes!)     │
│ 0x01000000 - Process A Heap                                        │
│ 0x02000000 - Process B Heap                                        │
│ 0x03000000 - Process A Stack                                       │
│ 0x04000000 - Process B Stack                                       │
│ ...                                                                │
└────────────────────────────────────────────────────────────────────┘
                                  ▲
                                  │
                    ┌─────────────┴──────────────┐
                    │   CPU Memory Management    │
                    │   Unit (MMU) + Page Tables │
                    └─────────────┬──────────────┘
                                  │
                                  ▼
┌────────────────────────────────────────────────────────────────────┐
│                    DISK (Page File / Swap Space)                   │
├────────────────────────────────────────────────────────────────────┤
│ Process A Data (0x20000000) - Paged out to disk                    │
│ Process B Data (0x20000000) - Paged out to disk                    │
│ Inactive pages from various processes...                           │
└────────────────────────────────────────────────────────────────────┘

Legend:
  ─────► : Page Table mapping (virtual → physical)
  Private pages: Each process has its own physical copy
  Shared pages: One physical copy mapped to multiple processes
  Paged out: Not in RAM; stored on disk, loaded on demand
```




#### The Illusion: What Each Process Sees

Consider two processes running simultaneously - perhaps two instances of Notepad, or Chrome and Word running side by side. From the perspective of **Process A**, the world looks simple and orderly: its **code** resides at **virtual address 0x00400000**, its **heap** (where dynamic allocations live) starts at **0x10000000**, and its **stack** sits near the top of user space at **0x7FF00000**. These addresses feel concrete and permanent to the process - when the program dereferences a pointer to `0x00400000`, it gets its code, every single time.

Now consider **Process B**, running concurrently on the same machine. Remarkably, it also sees its **code at 0x00400000**, its **heap at 0x10000000**, and its **stack at 0x7FF00000** - the **same addresses** as Process A! This seems impossible: how can two different programs occupy the same memory locations simultaneously? The answer is that these are **virtual addresses**, not real locations in **physical RAM**. Each process operates within its own private virtual address space, completely isolated and unaware of the others, seeing a pristine and exclusive view of memory that exists only from its perspective.

#### The Reality: Physical Memory Layout

Behind the scenes, the truth is quite different. **Physical RAM** - the actual silicon memory chips in your computer - is a single, shared resource that all processes must use cooperatively. When we peer behind the virtual memory curtain, we see that **Process A's code** might actually reside at **physical address 0x00100000**, while **Process B's code** is located at an entirely different location: **physical address 0x00500000**. Similarly, **Process A's heap** might occupy **physical address 0x01000000**, while **Process B's heap** is off at **0x02000000**. The physical addresses are scattered throughout RAM based on where the operating system's memory manager found available space, bearing no resemblance to the tidy virtual layout each process perceives.

This separation means that when Process A writes to its virtual address `0x10000000` (its heap), it's actually modifying physical memory at `0x01000000`, while Process B writing to its virtual address `0x10000000` is modifying a completely different region at physical address `0x02000000`. The two processes can use identical virtual addresses without conflict because those addresses don't directly reference physical memory - they're simply numbers that will be translated before reaching actual RAM.

#### The Magic: Hardware-Assisted Translation

The bridge between these two worlds - between the virtual addresses processes use and the physical addresses where data actually lives - is the **CPU's Memory Management Unit**, or **MMU**. This specialized hardware component sits between the CPU's execution units and the memory bus, intercepting every memory access. Its job is to **translate virtual addresses to physical addresses** in real-time, transparently and at tremendous speed, making the virtual memory illusion seamless.

The MMU doesn't perform this translation by guessing or through some complex algorithm running in software. Instead, it consults **page tables** - data structures maintained by the operating system that serve as lookup tables mapping virtual addresses to physical addresses. When Process A tries to access virtual address `0x00400000`, the MMU consults Process A's page tables, discovers that this virtual address maps to physical address `0x00100000`, and directs the memory request there. Moments later, when Process B accesses its virtual address `0x00400000`, the MMU consults Process B's completely separate page tables, finds that the same virtual address maps to physical address `0x00500000`, and routes that access to an entirely different location in RAM.

This translation happens for every single memory access - billions of times per second - yet modern CPUs perform it so efficiently through specialized caches (Translation Lookaside Buffers, or TLBs) that the overhead is nearly imperceptible. The result is a system where every process enjoys the simplicity and security of having its own private address space, while the operating system efficiently manages the shared physical memory underneath, allocating and reclaiming RAM as processes come and go, completely invisible to the applications themselves.


### Three Additional Key Concepts

#### 1. **Shared Memory: The DLL Efficiency Trick**

System DLLs like **kernel32.dll** are used by nearly every Windows process. Without sharing, if 50 processes each needed their own copy of `kernel32.dll` (which is ~1MB), that would waste 50MB of RAM holding identical copies.

**How Windows solves this:**

- `kernel32.dll` is loaded **once** into physical memory (e.g., at physical address `0x00800000`)
- Both Process A and Process B's **page tables** map their virtual address `0x76D00000` to the **same physical location**
- Result: One copy in RAM, shared by all processes - massive memory savings


#### 2. **Paging: When RAM Isn't Enough**

Your computer might have 16GB of RAM, but processes can allocate far more virtual memory than that. How? Not all virtual memory needs to be in physical RAM simultaneously.

**The paging mechanism:**

When physical RAM fills up, Windows uses the **page file** (`pagefile.sys` on disk) as overflow storage:

1. **Page out**: The OS identifies inactive memory pages (e.g., Process A's data at virtual address `0x20000000`)
2. **Write to disk**: Contents are written to the page file and the physical RAM is freed
3. **Update page table**: The page table entry is marked "not present" and records the disk location
4. **Later access**: When Process A tries to access `0x20000000`, the MMU triggers a **page fault**
5. **Page in**: The OS loads the data back from disk into RAM and updates the page table
6. **Execution continues**: The process never knows this happened - it's transparent

**Paging trade-offs:**

- ✅ **Advantage**: Processes can use more memory than physically available
- ✅ **Advantage**: Inactive data can be swapped out, freeing RAM for active processes
- ❌ **Disadvantage**: Disk is ~1000× slower than RAM; excessive paging ("thrashing") kills performance
- ❌ **Disadvantage**: Page faults cause delays while data is retrieved from disk

#### 3. **Private vs. Shared vs. Paged: The Full Picture**

Not all memory is created equal. Here's how different types of memory mappings work:

**Private Memory** (Process-specific data)

- Each process has its own physical pages
- Examples: heap allocations, stack, writable global variables
- Cannot be shared with other processes
- Can be paged to disk if inactive

**Shared Memory** (Read-only or explicitly shared)

- One physical page mapped to multiple virtual addresses across processes
- Examples: system DLLs, memory-mapped files opened with sharing
- Massive memory savings for common resources
- Code pages typically stay in RAM (not paged) due to frequent access

**Paged Memory** (Overflow to disk)

- Virtual pages not currently in physical RAM
- Stored in page file on disk
- Transparently loaded on access (page fault → page in)
- Can be either private or shared pages that were evicted



### The MMU's Job: Orchestrating It All

Every memory access goes through this flow:

```
1. Process accesses virtual address 0x76D00000
                ↓
2. MMU consults process's page table
                ↓
3. Page table lookup reveals:
   ├─► Present in RAM? → Direct to physical address (fast path)
   ├─► Not present (paged out)? → Page fault → OS loads from disk
   └─► Protection violation? → Access denied → Exception
                ↓
4. Physical memory access completes
                ↓
5. Result returned to process (which never knew about translation)
```

**Performance optimizations:**

- **TLB (Translation Lookaside Buffer)**: Caches recent virtual→physical translations
- **Working Set Management**: OS tracks which pages each process actively uses
- **Prefetching**: OS predicts which pages will be needed and loads them proactively




## Memory Regions and Protections: Security Through Hardware

Virtual memory isn't just about isolation - it's also about **security**. Every page of memory has **protection flags** that control what operations are allowed: reading data, writing data, or executing code. These protections, enforced by the CPU's MMU, form a critical defense against bugs and malicious code.

### Memory Protection Flags: The Access Control System

Windows provides a fine-grained system of **memory protection flags** that determine exactly what a process can do with each page of memory. These flags are set when memory is allocated and can be changed later (with the right permissions).

#### Base Protection Flags

```
┌──────────────────────────────────────────────────────────────┐
│                   MEMORY PROTECTION FLAGS                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  FLAG                      VALUE    DESCRIPTION              │
│  ────────────────────────────────────────────────────────────│
│  PAGE_NOACCESS            0x01     No access                 │
│  PAGE_READONLY            0x02     Read only                 │
│  PAGE_READWRITE           0x04     Read + Write              │
│  PAGE_WRITECOPY           0x08     Copy on write             │
│  PAGE_EXECUTE             0x10     Execute only              │
│  PAGE_EXECUTE_READ        0x20     Execute + Read            │
│  PAGE_EXECUTE_READWRITE   0x40     Execute + Read + Write    │
│  PAGE_EXECUTE_WRITECOPY   0x80     Execute + Copy on write   │
│                                                              │
│  MODIFIERS:                                                  │
│  PAGE_GUARD               0x100    Guard page (exception)    │
│  PAGE_NOCACHE             0x200    Disable caching           │
│  PAGE_WRITECOMBINE        0x400    Write combining           │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


|Flag|Value|What It Allows|Common Use|
|---|---|---|---|
|**PAGE_NOACCESS**|0x01|Nothing - any access triggers exception|Guard pages, reserved address space|
|**PAGE_READONLY**|0x02|Read only|String constants, import tables|
|**PAGE_READWRITE**|0x04|Read + Write|Heap, stack, writable globals|
|**PAGE_WRITECOPY**|0x08|Read + Copy-on-write|Shared DLLs with potential modifications|
|**PAGE_EXECUTE**|0x10|Execute only (rare)|Rarely used in practice|
|**PAGE_EXECUTE_READ**|0x20|Execute + Read|Code sections (.text)|
|**PAGE_EXECUTE_READWRITE**|0x40|Execute + Read + Write|⚠️ Dangerous - code that modifies itself|
|**PAGE_EXECUTE_WRITECOPY**|0x80|Execute + Copy-on-write|Shared code with potential modifications|

#### Protection Modifiers

Beyond the basic read/write/execute permissions, Windows offers special modifiers that change page behaviour:

**PAGE_GUARD (0x100)** - Guard Page Exception Trigger

- First access to this page causes a one-time exception
- After the exception, the guard flag is automatically removed
- Used for stack growth detection and memory debugging
- Example: The page just below your stack has `PAGE_GUARD` to catch stack overflows

**PAGE_NOCACHE (0x200)** - Disable CPU Caching

- Prevents the CPU from caching this memory in L1/L2/L3 caches
- Used for memory-mapped hardware registers where caching would cause stale data
- Critical for device driver memory that must always reflect current hardware state

**PAGE_WRITECOMBINE (0x400)** - Write Combining

- Multiple writes are batched together before sending to RAM
- Dramatically improves performance for video memory and framebuffers
- Trades consistency for speed - not suitable for normal program data



### Typical Memory Layout: Protection in Practice

Different regions of a process's memory require different protection levels based on their purpose. Here's how a typical process organizes its memory protections:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PROCESS MEMORY LAYOUT                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  0x00400000  ┌──────────────────────────────────┐                   │
│              │    .text (Code Section)          │                   │
│              │    Protection: PAGE_EXECUTE_READ │                   │
│              │    Machine instructions          │                   │
│              │    ✓ Can read  ✓ Can execute     │                   │
│              │    ✗ Cannot write (immutable)    │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
│  0x00410000  ┌──────────────────────────────────┐                   │
│              │    .data (Initialized Data)      │                   │
│              │    Protection: PAGE_READWRITE    │                   │
│              │    Global variables with values  │                   │
│              │    ✓ Can read  ✓ Can write       │                   │
│              │    ✗ Cannot execute              │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
│  0x00420000  ┌──────────────────────────────────┐                   │
│              │    .rdata (Read-Only Data)       │                   │
│              │    Protection: PAGE_READONLY     │                   │
│              │    String literals, import table │                   │
│              │    ✓ Can read                    │                   │
│              │    ✗ Cannot write  ✗ Cannot exec │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
│  0x10000000  ┌──────────────────────────────────┐                   │
│              │    Heap Allocations              │                   │
│              │    Protection: PAGE_READWRITE    │                   │
│              │    malloc(), new, HeapAlloc()    │                   │
│              │    ✓ Can read  ✓ Can write       │                   │
│              │    ✗ Cannot execute (DEP!)       │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
│  0x7FEFF000  ┌──────────────────────────────────┐                   │
│              │    Guard Page                    │                   │
│              │    PAGE_READWRITE + PAGE_GUARD   │                   │
│              │    Stack overflow detector       │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
│  0x7FF00000  ┌──────────────────────────────────┐                   │
│              │    Stack                         │                   │
│              │    Protection: PAGE_READWRITE    │                   │
│              │    Local variables, return addrs │                   │
│              │    ✓ Can read  ✓ Can write       │                   │
│              │    ✗ Cannot execute (DEP!)       │                   │
│              └──────────────────────────────────┘                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### Why These Protections Matter

**Code sections (.text)**: Marked **PAGE_EXECUTE_READ** to allow the CPU to run instructions but prevent modification. This stops attackers from overwriting your program's code with malicious instructions.

**Data sections (.data)**: Use **PAGE_READWRITE** because programs need to modify global variables, but these pages are never executable - even if an attacker corrupts data, they can't make the CPU jump there and run it as code.

**Read-only sections (.rdata)**: Protected as **PAGE_READONLY** for constants and import tables. Attempting to write here triggers an access violation, catching bugs where code accidentally tries to modify string literals.

**Heap and stack**: Both **PAGE_READWRITE** by default, but critically, **not executable**. This is enforced by **Data Execution Prevention (DEP)**, preventing the classic exploit technique of injecting shellcode into a buffer and jumping to it.

**Guard pages**: Placed at stack boundaries with **PAGE_READWRITE + PAGE_GUARD**. If the stack grows too large and hits the guard page, an exception fires before memory corruption occurs, catching stack overflow bugs.



### The Code Injection Problem: A Practical Example

Memory protections create a significant challenge for legitimate use cases like debugging tools, game trainers, and security research - and for attackers trying to inject malicious code. Understanding this problem illuminates why these protections exist.

#### The Classic Mistake

Imagine you're writing a tool that needs to inject code into another process. Here's what happens if you ignore memory protections:


##### Step 1: Allocate memory in target process
```c
   hRemoteMem = VirtualAllocEx(hProcess, NULL, size, MEM_COMMIT, PAGE_READWRITE);
   // Memory is readable and writable, but NOT executable
```

##### Step 2: Write your shellcode
```c
   WriteProcessMemory(hProcess, hRemoteMem, shellcode, size, NULL);
   // Successfully writes bytes to the allocated memory
```


##### Step 3: Try to execute it
```c
   CreateRemoteThread(hProcess, NULL, 0, hRemoteMem, NULL, 0, NULL);
   // 💥 CRASH! Access violation!
```


##### Problem
The memory was allocated with `PAGE_READWRITE`, which means the CPU can read from it and write to it, but when the thread tries to execute instructions from that address, the MMU blocks it - the page isn't marked executable!


#### A Feasible-but-Flawed Solution


##### Step 1: Allocate executable+writable memory

```c
hRemoteMem = VirtualAllocEx(hProcess, NULL, size, MEM_COMMIT, PAGE_EXECUTE_READWRITE);
   // RWX: Read, Write, AND Execute all at once
```

##### Step 2: Write shellcode
```c
   WriteProcessMemory(hProcess, hRemoteMem, shellcode, size, NULL);
```

##### Step 3: Execute

```c
   CreateRemoteThread(hProcess, NULL, 0, hRemoteMem, NULL, 0, NULL);
   // ✓ Works!
```


##### Problem
Since the memory is now marked as `PAGE_EXECUTE_READWRITE`, the MMU no longer has any issue with the request to execute the code, so it won't get in the way any longer. However, `PAGE_EXECUTE_READWRITE `pages are extremely rare in legitimate programs - they're a hallmark of exploits: **inject code, then run it**. So while the MMU won't have any issues, any **modern EDR will immediately flag RWX allocations**.



#### The Better Approach

#####  Step 1: Allocate writable memory

```c
   hRemoteMem = VirtualAllocEx(hProcess, NULL, size, MEM_COMMIT, PAGE_READWRITE);
```


##### Step 2: Write shellcode while memory is writable

```c
   WriteProcessMemory(hProcess, hRemoteMem, shellcode, size, NULL);
```

##### Step 3: Change protection to executable (but not writable)

```c
   DWORD oldProtect;
   VirtualProtectEx(hProcess, hRemoteMem, size, PAGE_EXECUTE_READ, &oldProtect);
   // Now the page is executable and readable, but not writable
```

##### Step 4: Execute the code

```c
   CreateRemoteThread(hProcess, NULL, 0, hRemoteMem, NULL, 0, NULL);
   // ✓ Works! The page is marked executable
```

##### **Why this is better:**
We separate the act of writing to memory from the act of executing from it - **code is never both writable and executable at the same time**.

This is, as a general rule, less suspicious to security software monitoring for RWX (read-write-execute) pages. But, marking memory as RW, copying into it, then immediately changing it to RX, can itself be flagged. So what to do? That's for a later lesson... ;)



### Data Execution Prevention (DEP): The Last Line of Defense

**Data Execution Prevention (DEP)**, also called **NX (No-Execute)** or **XD (Execute Disable)**, is a hardware-enforced security feature that fundamentally changes the memory protection landscape.

#### How DEP Works

**Hardware Level:**

- Modern CPUs have an **NX bit** in each page table entry
- When set, the page is marked non-executable regardless of its data
- The MMU enforces this at hardware speed - no software can bypass it

**Operating System Level:**

- Windows enables DEP by default for all processes (since Vista/7)
- The OS marks data pages (heap, stack) with the NX bit set
- Code pages (.text sections) have NX bit clear, allowing execution

**The Protection:**

```
Without DEP:                    With DEP:
───────────────                 ──────────────
Stack: PAGE_READWRITE           Stack: PAGE_READWRITE + NX bit
→ Attacker writes shellcode     → Attacker writes shellcode
→ Jumps to stack                → Jumps to stack
→ Code executes ✗ (exploit!)    → CPU blocks execution ✓ (safe!)
```

#### DEP Bypass Techniques

Despite DEP's strength, attackers have developed sophisticated bypass techniques:

|Technique|How It Works|Defense|
|---|---|---|
|**Allocate with EXECUTE**|Use VirtualAlloc with PAGE_EXECUTE_* flags|Easily detected by security tools monitoring for unusual allocations|
|**VirtualProtect**|Change existing memory to executable|Requires EXECUTE permission; monitored by security software|
|**Return-Oriented Programming (ROP)**|Chain together existing executable code snippets ("gadgets") to perform malicious actions without injecting new code|Address Space Layout Randomization (ASLR) makes gadgets harder to find; Control Flow Guard (CFG) blocks unexpected returns|
|**Code Caves**|Find unused space in existing executable sections and write code there|Limited space available; modern compilers leave less unused space; CFG prevents jumping to arbitrary locations|



### The Modern Security Stack

DEP doesn't work alone - it's part of a layered defense:

```
┌────────────────────────────────────────────────────┐
│              MODERN SECURITY LAYERS                │
├────────────────────────────────────────────────────┤
│                                                    │
│  DEP/NX              → Prevents data execution     │
│         +                                          │
│  ASLR                → Randomizes addresses        │
│         +                                          │
│  Control Flow Guard  → Validates indirect calls    │
│         +                                          │
│  Stack Cookies       → Detects buffer overflows    │
│         +                                          │
│  Heap Isolation      → Separates allocations       │
│                                                    │
│  = Defense in Depth                                │
│                                                    │
└────────────────────────────────────────────────────┘
```

Together, these protections make memory exploitation extraordinarily difficult, turning what used to be simple attacks into complex research projects requiring chains of multiple vulnerabilities.



## Virtual Address Descriptors (VADs): The Kernel's Memory Map

When a process allocates memory, opens a file mapping, or loads a DLL, the operating system needs to track what memory regions exist, where they are, what permissions they have, and what they represent. The kernel accomplishes this through **Virtual Address Descriptors (VADs)** - a sophisticated data structure that maintains a complete map of every memory region in a process's address space.

### What Are VADs?

**VADs** are kernel data structures that describe memory regions. Each VAD node represents a contiguous range of virtual addresses with consistent properties: the same protection flags, the same type (heap vs. file-backed), and the same backing store. The kernel organizes these VADs into a **balanced binary tree** for efficient lookup, allowing it to quickly answer questions like "what's at address 0x00400000?" or "is this memory region valid?"

Think of the VAD tree as the kernel's authoritative registry of a process's memory landscape - every allocated region, every mapped file, every loaded DLL has a corresponding VAD entry. Without VADs, the kernel would have no way to track what memory belongs to a process or enforce access controls.




### VAD Tree Structure: Organizing Memory Regions

The kernel maintains one **VAD tree per process**, rooted in the `EPROCESS` structure. This tree organizes memory regions hierarchically for fast searching.

```
                    VAD TREE FOR A PROCESS
                    
              Root VAD (0x10000000-0x20000000)
              Protection: PAGE_READWRITE
              Type: Private (Heap)
                         /              \
                        /                \
                       /                  \
         (0x00400000-0x00500000)    (0x30000000-0x40000000)
         Protection: PAGE_EXECUTE_READ   Protection: PAGE_READWRITE
         Type: Image (notepad.exe)       Type: Mapped (data file)
              /            \                    /              \
             /              \                  /                \
    (0x00100000-      (0x76D00000-    (0x20000000-       (0x50000000-
     0x00200000)       0x76E00000)     0x20010000)        0x50020000)
     Stack             ntdll.dll       Shared Memory      Private Heap
     RW-, Private      R-X, Image      RW-, Shareable     RW-, Private
```



### How the Tree Works

**Binary Search Efficiency:**

- The tree is organized by address ranges - left children have lower addresses, right children have higher
- Looking up "what's at address `0x76D50000`?" requires only `log(N`) comparisons
- Without the tree, the kernel would need to scan every memory region linearly

**Each VAD Node Contains:**

| Field                          | Purpose                                                |
| ------------------------------ | ------------------------------------------------------ |
| **Start Address**              | Beginning of this memory region (e.g., `0x00400000`)   |
| **End Address**                | End of this memory region (e.g., `0x00500000`)         |
| **Protection Flags**           | `PAGE_EXECUTE_READ`, `PAGE_READWRITE`, etc.            |
| **VAD Type**                   | Private, Image, Mapped, Shareable                      |
| **File Object Pointer**        | For mapped files/DLLs: points to the underlying file   |
| **Commit Charge**              | How much physical memory or page file this region uses |
| **Parent/Left/Right Pointers** | Tree structure navigation                              |

**Dynamic Updates:**

- When `VirtualAlloc()` is called, the kernel creates a new VAD node and inserts it into the tree
- When `VirtualFree()` is called, the corresponding VAD is removed
- When `VirtualProtect()` changes permissions, the VAD's protection flags are updated




### Enumerating VADs from User Mode: Practical Reconnaissance

While the kernel maintains the VAD tree internally, we can query memory regions from userland using the **VirtualQuery()** API. This allows an application to enumerate its own memory layout - or for offensive tools, to map a target process's address space.

#### Memory Region Scanner

Here's a practical Go implementation that walks through a process's entire address space, querying each region to build a memory map:

```go
// Query memory regions (VAD entries visible from user mode)
package main

import (
    "fmt"
    "syscall"
    "unsafe"
)

// MEMORY_BASIC_INFORMATION: Structure returned by VirtualQuery
// Describes a contiguous region of memory with uniform properties
type MEMORY_BASIC_INFORMATION struct {
    BaseAddress       uintptr  // Starting address of region
    AllocationBase    uintptr  // Base address of allocation that contains this region
    AllocationProtect uint32   // Protection when region was originally allocated
    RegionSize        uintptr  // Size of region in bytes
    State             uint32   // MEM_COMMIT, MEM_RESERVE, or MEM_FREE
    Protect           uint32   // Current protection flags
    Type              uint32   // MEM_PRIVATE, MEM_MAPPED, or MEM_IMAGE
}

// Memory state constants
const (
    MEM_COMMIT    = 0x1000   // Memory is committed (has physical/page file backing)
    MEM_RESERVE   = 0x2000   // Memory is reserved (address space reserved but not backed)
    MEM_FREE      = 0x10000  // Memory is free (not allocated)
    
    MEM_PRIVATE   = 0x20000    // Private memory (heap, stack)
    MEM_MAPPED    = 0x40000    // Mapped file
    MEM_IMAGE     = 0x1000000  // Executable image (PE file)
)

var (
    kernel32              = syscall.NewLazyDLL("kernel32.dll")
    procVirtualQuery      = kernel32.NewProc("VirtualQuery")
)

// VirtualQuery: Query information about a memory address
func VirtualQuery(address uintptr) (*MEMORY_BASIC_INFORMATION, error) {
    var mbi MEMORY_BASIC_INFORMATION
    
    // Call VirtualQuery API
    ret, _, err := procVirtualQuery.Call(
        address,                           // Address to query
        uintptr(unsafe.Pointer(&mbi)),    // Buffer to receive info
        unsafe.Sizeof(mbi),               // Size of buffer
    )
    
    if ret == 0 {
        return nil, err  // Query failed
    }
    return &mbi, nil
}

// EnumerateMemory: Walk through entire address space
func EnumerateMemory() {
    var address uintptr = 0
    
    // Scan from 0 to max user-mode address on x64
    for address < 0x7FFFFFFF0000 {
        mbi, err := VirtualQuery(address)
        if err != nil {
            break  // End of accessible memory
        }
        
        // Only show committed memory (ignore reserved/free)
        if mbi.State == MEM_COMMIT {
            protection := getProtectionString(mbi.Protect)
            memType := getTypeString(mbi.Type)
            
            fmt.Printf("0x%016X - 0x%016X  %s  %s\n",
                mbi.BaseAddress,
                mbi.BaseAddress + mbi.RegionSize,
                protection,
                memType)
        }
        
        // Jump to next region
        address = mbi.BaseAddress + mbi.RegionSize
    }
}

// Helper: Convert protection flags to readable string
func getProtectionString(protect uint32) string {
    switch protect & 0xFF {
    case 0x01: return "---"   // PAGE_NOACCESS
    case 0x02: return "R--"   // PAGE_READONLY
    case 0x04: return "RW-"   // PAGE_READWRITE
    case 0x20: return "R-X"   // PAGE_EXECUTE_READ
    case 0x40: return "RWX"   // PAGE_EXECUTE_READWRITE
    default: return "???"
    }
}

// Helper: Convert type flags to readable string
func getTypeString(memType uint32) string {
    switch memType {
    case MEM_PRIVATE: return "Private"
    case MEM_MAPPED:  return "Mapped"
    case MEM_IMAGE:   return "Image"
    default: return "Unknown"
    }
}

func main() {
    fmt.Println("Memory Map of Current Process:")
    fmt.Println("Start Address       - End Address         Prot  Type")
    fmt.Println("─────────────────────────────────────────────────────")
    EnumerateMemory()
}
```


Once again let's compile using `go build`, then run it on the target machine using admin privs.

You should get the following output, but note that due to ASLR your addresses will differ from mine - in fact they will differ each time you run your application!

```shell
Memory Map of Current Process:
Start Address       - End Address         Prot  Type
─────────────────────────────────────────────────────
0x0000000000010000 - 0x0000000000011000  RW-  Mapped
0x0000000000020000 - 0x0000000000030000  RW-  Mapped
0x0000000000030000 - 0x0000000000050000  R--  Mapped
0x0000000000050000 - 0x0000000000054000  R--  Mapped
0x0000000000060000 - 0x0000000000062000  RW-  Private
0x0000000000070000 - 0x0000000000081000  R--  Mapped
0x0000000000090000 - 0x00000000000A1000  R--  Mapped
0x00000000000B0000 - 0x00000000000B3000  R--  Mapped
0x00000000000C0000 - 0x00000000000C7000  R--  Mapped
0x00000000000D0000 - 0x00000000000D7000  R--  Mapped
0x00000000000E0000 - 0x00000000000E2000  R--  Mapped
0x00000000000F0000 - 0x00000000000F2000  R--  Mapped
0x0000000000100000 - 0x0000000000102000  RW-  Private
0x0000000000140000 - 0x0000000000143000  R--  Mapped
0x0000000000150000 - 0x0000000000161000  R--  Mapped
0x0000000000170000 - 0x0000000000181000  R--  Mapped
0x0000000000190000 - 0x0000000000191000  RW-  Private
0x00000000001A0000 - 0x00000000001E0000  RW-  Private
0x00000000001E0000 - 0x0000000000200000  RW-  Private
0x00000000003DD000 - 0x00000000003E8000  RW-  Private
0x00000000005FA000 - 0x00000000005FD000  RW-  Private
0x00000000005FD000 - 0x0000000000600000  RW-  Private
0x0000000000600000 - 0x00000000006D3000  R--  Mapped
0x00000000006E0000 - 0x00000000006F0000  RW-  Private
0x00000000006F0000 - 0x0000000000700000  RW-  Private
0x0000000000700000 - 0x0000000000740000  RW-  Private
0x00000000007B0000 - 0x00000000007CA000  RW-  Private
0x0000000000930000 - 0x0000000000931000  RW-  Private
0x0000000000DB6000 - 0x0000000000DB7000  RW-  Private
0x00000000031E0000 - 0x00000000031E1000  RW-  Private
0x0000000015330000 - 0x0000000015331000  RW-  Private
0x0000000035330000 - 0x0000000035331000  RW-  Private
0x00000000451B0000 - 0x00000000459B0000  RW-  Private
0x00000000459B0000 - 0x0000000045AB0000  RW-  Private
0x0000000045CAB000 - 0x0000000045CAE000  RW-  Private
0x0000000045CAE000 - 0x0000000045CB0000  RW-  Private
0x0000000045EAC000 - 0x0000000045EAF000  RW-  Private
0x0000000045EAF000 - 0x0000000045EB0000  RW-  Private
0x00000000460AC000 - 0x00000000460AF000  RW-  Private
0x00000000460AF000 - 0x00000000460B0000  RW-  Private
0x00000000462AC000 - 0x00000000462AF000  RW-  Private
0x00000000462AF000 - 0x00000000462B0000  RW-  Private
0x000000007FFE0000 - 0x000000007FFE1000  R--  Private
0x000000007FFEE000 - 0x000000007FFEF000  R--  Private
0x000000C000000000 - 0x000000C00006E000  RW-  Private
0x00007FF4FDEC0000 - 0x00007FF4FDEC5000  R--  Mapped
0x00007FF5FFFE0000 - 0x00007FF5FFFE1000  RW-  Private
0x00007FF5FFFF0000 - 0x00007FF5FFFF1000  R--  Mapped
0x00007FF702630000 - 0x00007FF702631000  R--  Image
0x00007FF702631000 - 0x00007FF7026D1000  R-X  Image
0x00007FF7026D1000 - 0x00007FF7027A1000  R--  Image
0x00007FF7027A1000 - 0x00007FF7027A3000  RW-  Image
0x00007FF7027A3000 - 0x00007FF7027A6000  ???  Image
0x00007FF7027A6000 - 0x00007FF7027AB000  RW-  Image
0x00007FF7027AB000 - 0x00007FF7027AD000  ???  Image
0x00007FF7027AD000 - 0x00007FF7027B1000  RW-  Image
0x00007FF7027B1000 - 0x00007FF7027B5000  ???  Image
0x00007FF7027B5000 - 0x00007FF7027B6000  RW-  Image
0x00007FF7027B6000 - 0x00007FF7027BD000  ???  Image
0x00007FF7027BD000 - 0x00007FF7027BE000  RW-  Image
0x00007FF7027BE000 - 0x00007FF7027C5000  ???  Image
0x00007FF7027C5000 - 0x00007FF7027CD000  RW-  Image
0x00007FF7027CD000 - 0x00007FF7027F3000  ???  Image
0x00007FF7027F3000 - 0x00007FF7027F8000  RW-  Image
0x00007FF7027F8000 - 0x00007FF70289F000  R--  Image
0x00007FF70289F000 - 0x00007FF7028A0000  ???  Image
0x00007FF7028A0000 - 0x00007FF7028BF000  R--  Image
0x00007FFD5B4B0000 - 0x00007FFD5B4B1000  R--  Image
0x00007FFD5B4B1000 - 0x00007FFD5B50A000  R-X  Image
0x00007FFD5B50A000 - 0x00007FFD5B530000  R--  Image
0x00007FFD5B530000 - 0x00007FFD5B532000  RW-  Image
0x00007FFD5B532000 - 0x00007FFD5B54E000  R--  Image
0x00007FFD5B54E000 - 0x00007FFD5B54F000  R-X  Image
0x00007FFD5E1B0000 - 0x00007FFD5E1B1000  R--  Image
0x00007FFD5E1B1000 - 0x00007FFD5E1BB000  R-X  Image
0x00007FFD5E1BB000 - 0x00007FFD5E1BF000  R--  Image
0x00007FFD5E1BF000 - 0x00007FFD5E1C0000  RW-  Image
0x00007FFD5E1C0000 - 0x00007FFD5E1C4000  R--  Image
0x00007FFD5E1C4000 - 0x00007FFD5E1C5000  R-X  Image
0x00007FFD5E1D0000 - 0x00007FFD5E1D1000  R--  Image
0x00007FFD5E1D1000 - 0x00007FFD5E1E5000  R-X  Image
0x00007FFD5E1E5000 - 0x00007FFD5E1F0000  R--  Image
0x00007FFD5E1F0000 - 0x00007FFD5E1F1000  RW-  Image
0x00007FFD5E1F1000 - 0x00007FFD5E22E000  R--  Image
0x00007FFD5E22E000 - 0x00007FFD5E22F000  R-X  Image
0x00007FFD5E350000 - 0x00007FFD5E351000  R--  Image
0x00007FFD5E351000 - 0x00007FFD5E3C7000  R-X  Image
0x00007FFD5E3C7000 - 0x00007FFD5E3E1000  R--  Image
0x00007FFD5E3E1000 - 0x00007FFD5E3E2000  RW-  Image
0x00007FFD5E3E2000 - 0x00007FFD5E3E9000  R--  Image
0x00007FFD5E3E9000 - 0x00007FFD5E3EA000  R-X  Image
0x00007FFD5E840000 - 0x00007FFD5E841000  R--  Image
0x00007FFD5E841000 - 0x00007FFD5E9ED000  R-X  Image
0x00007FFD5E9ED000 - 0x00007FFD5EBDC000  R--  Image
0x00007FFD5EBDC000 - 0x00007FFD5EBE4000  RW-  Image
0x00007FFD5EBE4000 - 0x00007FFD5EBE6000  ???  Image
0x00007FFD5EBE6000 - 0x00007FFD5EC33000  R--  Image
0x00007FFD5EC33000 - 0x00007FFD5EC34000  R-X  Image
0x00007FFD5ECD0000 - 0x00007FFD5ECD1000  R--  Image
0x00007FFD5ECD1000 - 0x00007FFD5EDC8000  R-X  Image
0x00007FFD5EDC8000 - 0x00007FFD5EE07000  R--  Image
0x00007FFD5EE07000 - 0x00007FFD5EE0A000  RW-  Image
0x00007FFD5EE0A000 - 0x00007FFD5EE1B000  R--  Image
0x00007FFD5EE1B000 - 0x00007FFD5EE1C000  R-X  Image
0x00007FFD5F7C0000 - 0x00007FFD5F7C1000  R--  Image
0x00007FFD5F7C1000 - 0x00007FFD5F89D000  R-X  Image
0x00007FFD5F89D000 - 0x00007FFD5F8C3000  R--  Image
0x00007FFD5F8C3000 - 0x00007FFD5F8C5000  RW-  Image
0x00007FFD5F8C5000 - 0x00007FFD5F8D8000  R--  Image
0x00007FFD5F8D8000 - 0x00007FFD5F8D9000  R-X  Image
0x00007FFD60470000 - 0x00007FFD60471000  R--  Image
0x00007FFD60471000 - 0x00007FFD604F7000  R-X  Image
0x00007FFD604F7000 - 0x00007FFD6052F000  R--  Image
0x00007FFD6052F000 - 0x00007FFD60531000  RW-  Image
0x00007FFD60531000 - 0x00007FFD60539000  R--  Image
0x00007FFD60539000 - 0x00007FFD6053A000  R-X  Image
0x00007FFD611C0000 - 0x00007FFD611C1000  R--  Image
0x00007FFD611C1000 - 0x00007FFD61335000  R-X  Image
0x00007FFD61335000 - 0x00007FFD6138E000  R--  Image
0x00007FFD6138E000 - 0x00007FFD61398000  RW-  Image
0x00007FFD61398000 - 0x00007FFD61429000  R--  Image
0x00007FFD61429000 - 0x00007FFD6142A000  R-X  Image
```



**Reading the output:**

Each line represents a **VAD entry** - a contiguous memory region with uniform properties:

- **Address Range**: Where this memory lives in the virtual address space
- **Protection (Prot)**: Current access permissions
    - `R--` = Read only
    - `RW-` = Read + Write
    - `R-X` = Read + Execute (typical for code)
    - `RWX` = Read + Write + Execute (⚠️ suspicious!)
- **Type**: What backs this memory
    - `Image` = Loaded from a PE file (.exe or .dll)
    - `Private` = Process-private allocation (heap, stack)
    - `Mapped` = File-backed mapping




#### Memory Region Scanner with Annotations

This is cool, but we can make our code a little more useful by interpreting key regions and telling us what we're looking at. This can help us identify specific targeted regions in more complex applications.

```go
//go:build windows
// +build windows

// Enhanced memory scanner with PE parsing and module enumeration
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// MEMORY_BASIC_INFORMATION: Structure returned by VirtualQuery
type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

// IMAGE_DOS_HEADER: Beginning of every PE file ("MZ" signature)
type IMAGE_DOS_HEADER struct {
	E_magic    uint16
	E_cblp     uint16
	E_cp       uint16
	E_crlc     uint16
	E_cparhdr  uint16
	E_minalloc uint16
	E_maxalloc uint16
	E_ss       uint16
	E_sp       uint16
	E_csum     uint16
	E_ip       uint16
	E_cs       uint16
	E_lfarlc   uint16
	E_ovno     uint16
	E_res      [4]uint16
	E_oemid    uint16
	E_oeminfo  uint16
	E_res2     [10]uint16
	E_lfanew   int32 // Offset to PE header
}

// IMAGE_NT_HEADERS64: Main PE header
type IMAGE_NT_HEADERS64 struct {
	Signature      uint32
	FileHeader     IMAGE_FILE_HEADER
	OptionalHeader IMAGE_OPTIONAL_HEADER64
}

type IMAGE_FILE_HEADER struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

type IMAGE_OPTIONAL_HEADER64 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	ImageBase                   uint64
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	// ... rest of fields truncated - not required
}

// IMAGE_SECTION_HEADER: Describes a section (.text, .data, etc.)
type IMAGE_SECTION_HEADER struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

// MODULEINFO: Information about a loaded module
type MODULEINFO struct {
	BaseOfDll   uintptr
	SizeOfImage uint32
	EntryPoint  uintptr
}

// Memory state and type constants
const (
	MEM_COMMIT  = 0x1000
	MEM_RESERVE = 0x2000
	MEM_FREE    = 0x10000

	MEM_PRIVATE = 0x20000
	MEM_MAPPED  = 0x40000
	MEM_IMAGE   = 0x1000000
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	psapi                    = syscall.NewLazyDLL("psapi.dll")
	procVirtualQuery         = kernel32.NewProc("VirtualQuery")
	procEnumProcessModules   = psapi.NewProc("EnumProcessModules")
	procGetModuleInformation = psapi.NewProc("GetModuleInformation")
	procGetModuleBaseNameW   = psapi.NewProc("GetModuleBaseNameW")
	procGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
)

// SectionInfo: Stores information about a PE section
type SectionInfo struct {
	Name  string
	Start uintptr
	End   uintptr
}

// ModuleInfo: Stores information about loaded modules (DLLs)
type ModuleInfo struct {
	BaseAddress uintptr
	Size        uint32
	Name        string
	Sections    []SectionInfo // List of sections with ranges
}

var loadedModules []ModuleInfo

// EnumerateModules: Build a list of all loaded DLLs and their sections
func EnumerateModules() {
	hProcess, _, _ := procGetCurrentProcess.Call()

	var modules [1024]syscall.Handle
	var needed uint32

	// Get all loaded module handles
	ret, _, _ := procEnumProcessModules.Call(
		hProcess,
		uintptr(unsafe.Pointer(&modules[0])),
		uintptr(len(modules)*int(unsafe.Sizeof(modules[0]))),
		uintptr(unsafe.Pointer(&needed)),
	)

	if ret == 0 {
		return
	}

	moduleCount := int(needed) / int(unsafe.Sizeof(modules[0]))

	// For each module, get its info and parse PE sections
	for i := 0; i < moduleCount; i++ {
		var modInfo MODULEINFO

		// Get module base address and size
		procGetModuleInformation.Call(
			hProcess,
			uintptr(modules[i]),
			uintptr(unsafe.Pointer(&modInfo)),
			unsafe.Sizeof(modInfo),
		)

		// Get module name
		var nameBuffer [260]uint16
		procGetModuleBaseNameW.Call(
			hProcess,
			uintptr(modules[i]),
			uintptr(unsafe.Pointer(&nameBuffer[0])),
			uintptr(len(nameBuffer)),
		)

		moduleName := syscall.UTF16ToString(nameBuffer[:])

		// Parse PE sections for this module
		sections := parsePESections(modInfo.BaseOfDll)

		loadedModules = append(loadedModules, ModuleInfo{
			BaseAddress: modInfo.BaseOfDll,
			Size:        modInfo.SizeOfImage,
			Name:        moduleName,
			Sections:    sections,
		})
	}
}

// parsePESections: Extract section names and ranges from PE headers
func parsePESections(baseAddress uintptr) []SectionInfo {
	var sections []SectionInfo

	defer func() {
		// Catch any access violations from reading invalid memory
		recover()
	}()

	// Read DOS header
	dosHeader := (*IMAGE_DOS_HEADER)(unsafe.Pointer(baseAddress))
	if dosHeader.E_magic != 0x5A4D { // "MZ"
		return sections
	}

	// Read NT headers (just signature and file header first)
	peHeaderOffset := baseAddress + uintptr(dosHeader.E_lfanew)
	signature := (*uint32)(unsafe.Pointer(peHeaderOffset))
	if *signature != 0x00004550 { // "PE\0\0"
		return sections
	}

	// Read file header (comes after 4-byte signature)
	fileHeader := (*IMAGE_FILE_HEADER)(unsafe.Pointer(peHeaderOffset + 4))

	// Section table comes after: Signature (4) + FileHeader (20) + OptionalHeader (variable)
	// Use SizeOfOptionalHeader from FileHeader for correct offset
	sectionTableOffset := peHeaderOffset + 4 +
		unsafe.Sizeof(IMAGE_FILE_HEADER{}) +
		uintptr(fileHeader.SizeOfOptionalHeader)

	numSections := int(fileHeader.NumberOfSections)

	for i := 0; i < numSections; i++ {
		sectionPtr := (*IMAGE_SECTION_HEADER)(unsafe.Pointer(
			sectionTableOffset + uintptr(i)*unsafe.Sizeof(IMAGE_SECTION_HEADER{}),
		))

		// Extract and validate section name
		name := cleanSectionName(sectionPtr.Name[:])
		if name == "" {
			continue // Skip invalid sections
		}

		sectionStart := baseAddress + uintptr(sectionPtr.VirtualAddress)
		sectionEnd := sectionStart + uintptr(sectionPtr.VirtualSize)

		sections = append(sections, SectionInfo{
			Name:  name,
			Start: sectionStart,
			End:   sectionEnd,
		})
	}

	return sections
}

// cleanSectionName: Extract and validate section name from PE header
func cleanSectionName(nameBytes []byte) string {
	// Find null terminator or use all 8 bytes
	length := 0
	for length < len(nameBytes) && nameBytes[length] != 0 {
		length++
	}

	if length == 0 {
		return ""
	}

	// Validate all characters are printable ASCII
	for i := 0; i < length; i++ {
		if nameBytes[i] < 0x20 || nameBytes[i] > 0x7E {
			return "" // Contains non-printable characters
		}
	}

	return string(nameBytes[:length])
}

// identifyRegion: Determine what a memory region represents
func identifyRegion(mbi *MEMORY_BASIC_INFORMATION) string {
	addr := mbi.BaseAddress

	// Check for special fixed addresses
	if addr == 0x7FFE0000 || addr == 0x7FFEE000 {
		return "PEB (Process Environment Block)"
	}
	if addr >= 0x7FFE0000 && addr < 0x7FFF0000 {
		return "Shared User Data / PEB"
	}

	// For Image type, match against loaded modules
	if mbi.Type == MEM_IMAGE {
		for _, mod := range loadedModules {
			if addr >= mod.BaseAddress && addr < mod.BaseAddress+uintptr(mod.Size) {
				// Check if this address falls within any section
				for _, section := range mod.Sections {
					if addr >= section.Start && addr < section.End {
						return fmt.Sprintf("%s (%s)", mod.Name, section.Name)
					}
				}
				// If no section match, just return module name
				return mod.Name
			}
		}
		return "Unknown Image"
	}

	// Private memory - could be heap or stack
	if mbi.Type == MEM_PRIVATE {
		// Large RW- regions in specific ranges are often heap
		if mbi.RegionSize > 0x10000 && (mbi.Protect&0xFF == 0x04) {
			return "Heap"
		}
		// Smaller regions might be stack or thread-local storage
		if mbi.RegionSize <= 0x10000 {
			return "Stack / TLS"
		}
		return "Private"
	}

	// Mapped memory (files)
	if mbi.Type == MEM_MAPPED {
		return "Memory-Mapped File"
	}

	return "Unknown"
}

// VirtualQuery: Query information about a memory address
func VirtualQuery(address uintptr) (*MEMORY_BASIC_INFORMATION, error) {
	var mbi MEMORY_BASIC_INFORMATION

	ret, _, err := procVirtualQuery.Call(
		address,
		uintptr(unsafe.Pointer(&mbi)),
		unsafe.Sizeof(mbi),
	)

	if ret == 0 {
		return nil, err
	}
	return &mbi, nil
}

// EnumerateMemory: Walk through entire address space
func EnumerateMemory() {
	var address uintptr = 0

	for address < 0x7FFFFFFF0000 {
		mbi, err := VirtualQuery(address)
		if err != nil {
			break
		}

		if mbi.State == MEM_COMMIT {
			protection := getProtectionString(mbi.Protect)
			memType := getTypeString(mbi.Type)
			interpretation := identifyRegion(mbi)

			fmt.Printf("0x%016X - 0x%016X  %s  %-7s  %s\n",
				mbi.BaseAddress,
				mbi.BaseAddress+mbi.RegionSize,
				protection,
				memType,
				interpretation)
		}

		address = mbi.BaseAddress + mbi.RegionSize
	}
}

// getProtectionString: Convert protection flags to readable string
func getProtectionString(protect uint32) string {
	switch protect & 0xFF {
	case 0x01:
		return "---"
	case 0x02:
		return "R--"
	case 0x04:
		return "RW-"
	case 0x20:
		return "R-X"
	case 0x40:
		return "RWX"
	default:
		return "???"
	}
}

// getTypeString: Convert type flags to readable string
func getTypeString(memType uint32) string {
	switch memType {
	case MEM_PRIVATE:
		return "Private"
	case MEM_MAPPED:
		return "Mapped"
	case MEM_IMAGE:
		return "Image"
	default:
		return "Unknown"
	}
}

func main() {
	fmt.Println("Enhanced Memory Map of Current Process:")
	fmt.Println("Start Address       - End Address         Prot  Type     Interpretation")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")

	// First, enumerate all loaded modules and parse their PE sections
	EnumerateModules()

	// Then scan memory and identify regions
	EnumerateMemory()
}

```


And now when we run our application we'll have a fifth column telling us, where possible, what this region represents.

```shell
PS C:\Users\tresa\OneDrive\Desktop> .\memscanner_annotated.exe
Enhanced Memory Map of Current Process:
Start Address       - End Address         Prot  Type     Interpretation
────────────────────────────────────────────────────────────────────────────────
0x0000000000010000 - 0x0000000000011000  RW-  Mapped   Memory-Mapped File
0x0000000000020000 - 0x0000000000030000  RW-  Mapped   Memory-Mapped File
0x0000000000030000 - 0x0000000000050000  R--  Mapped   Memory-Mapped File
0x0000000000050000 - 0x0000000000054000  R--  Mapped   Memory-Mapped File
0x0000000000060000 - 0x0000000000062000  RW-  Private  Stack / TLS
0x0000000000070000 - 0x0000000000081000  R--  Mapped   Memory-Mapped File
0x0000000000090000 - 0x00000000000A1000  R--  Mapped   Memory-Mapped File
0x00000000000B0000 - 0x00000000000B3000  R--  Mapped   Memory-Mapped File
0x00000000000C0000 - 0x00000000000C7000  R--  Mapped   Memory-Mapped File
0x00000000000D0000 - 0x00000000000D7000  R--  Mapped   Memory-Mapped File
0x00000000000E0000 - 0x00000000000F4000  RW-  Private  Heap
0x00000000001E0000 - 0x00000000001E2000  R--  Mapped   Memory-Mapped File
0x00000000001F0000 - 0x00000000001F2000  R--  Mapped   Memory-Mapped File
0x0000000000220000 - 0x000000000022B000  RW-  Private  Stack / TLS
0x00000000005FA000 - 0x00000000005FD000  RW-  Private  Stack / TLS
0x00000000005FD000 - 0x0000000000600000  RW-  Private  Stack / TLS
0x0000000000600000 - 0x0000000000602000  RW-  Private  Stack / TLS
0x0000000000640000 - 0x0000000000643000  R--  Mapped   Memory-Mapped File
0x0000000000650000 - 0x0000000000723000  R--  Mapped   Memory-Mapped File
0x0000000000730000 - 0x0000000000741000  R--  Mapped   Memory-Mapped File
0x0000000000750000 - 0x0000000000761000  R--  Mapped   Memory-Mapped File
0x0000000000770000 - 0x0000000000771000  RW-  Private  Stack / TLS
0x0000000000780000 - 0x00000000007C0000  RW-  Private  Heap
0x00000000007C0000 - 0x00000000007E0000  RW-  Private  Heap
0x0000000000860000 - 0x0000000000861000  RW-  Private  Stack / TLS
0x0000000000CE6000 - 0x0000000000CE7000  RW-  Private  Stack / TLS
0x0000000003110000 - 0x0000000003111000  RW-  Private  Stack / TLS
0x0000000015260000 - 0x0000000015261000  RW-  Private  Stack / TLS
0x0000000035260000 - 0x0000000035261000  RW-  Private  Stack / TLS
0x00000000450E0000 - 0x00000000458E0000  RW-  Private  Heap
0x00000000458E0000 - 0x00000000459E0000  RW-  Private  Heap
0x00000000459E0000 - 0x00000000459F0000  RW-  Private  Stack / TLS
0x00000000459F0000 - 0x0000000045A00000  RW-  Private  Stack / TLS
0x0000000045BFB000 - 0x0000000045BFE000  RW-  Private  Stack / TLS
0x0000000045BFE000 - 0x0000000045C00000  RW-  Private  Stack / TLS
0x0000000045DFC000 - 0x0000000045DFF000  RW-  Private  Stack / TLS
0x0000000045DFF000 - 0x0000000045E00000  RW-  Private  Stack / TLS
0x0000000045E00000 - 0x0000000045E40000  RW-  Private  Heap
0x000000004603B000 - 0x000000004603E000  RW-  Private  Stack / TLS
0x000000004603E000 - 0x0000000046040000  RW-  Private  Stack / TLS
0x000000004623C000 - 0x000000004623F000  RW-  Private  Stack / TLS
0x000000004623F000 - 0x0000000046240000  RW-  Private  Stack / TLS
0x000000007FFE0000 - 0x000000007FFE1000  R--  Private  PEB (Process Environment Block)
0x000000007FFEE000 - 0x000000007FFEF000  R--  Private  PEB (Process Environment Block)
0x000000C000000000 - 0x000000C000082000  RW-  Private  Heap
0x000000C000100000 - 0x000000C00010E000  RW-  Private  Stack / TLS
0x000000C00010E000 - 0x000000C000110000  RW-  Private  Stack / TLS
0x00007FF4FDEC0000 - 0x00007FF4FDEC5000  R--  Mapped   Memory-Mapped File
0x00007FF5FFFE0000 - 0x00007FF5FFFE1000  RW-  Private  Stack / TLS
0x00007FF5FFFF0000 - 0x00007FF5FFFF1000  R--  Mapped   Memory-Mapped File
0x00007FF6E41E0000 - 0x00007FF6E41E1000  R--  Image    memscanner_annotated.exe
0x00007FF6E41E1000 - 0x00007FF6E4282000  R-X  Image    memscanner_annotated.exe (.text)
0x00007FF6E4282000 - 0x00007FF6E4353000  R--  Image    memscanner_annotated.exe (.rdata)
0x00007FF6E4353000 - 0x00007FF6E4355000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E4355000 - 0x00007FF6E4358000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E4358000 - 0x00007FF6E435D000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E435D000 - 0x00007FF6E435F000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E435F000 - 0x00007FF6E4363000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E4363000 - 0x00007FF6E4367000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E4367000 - 0x00007FF6E4368000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E4368000 - 0x00007FF6E4370000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E4370000 - 0x00007FF6E4371000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E4371000 - 0x00007FF6E4377000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E4377000 - 0x00007FF6E437F000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E437F000 - 0x00007FF6E43A5000  ???  Image    memscanner_annotated.exe (.data)
0x00007FF6E43A5000 - 0x00007FF6E43AA000  RW-  Image    memscanner_annotated.exe (.data)
0x00007FF6E43AA000 - 0x00007FF6E4452000  R--  Image    memscanner_annotated.exe (.pdata)
0x00007FF6E4452000 - 0x00007FF6E4453000  ???  Image    memscanner_annotated.exe (.idata)
0x00007FF6E4453000 - 0x00007FF6E4472000  R--  Image    memscanner_annotated.exe (.reloc)
0x00007FFD5E1B0000 - 0x00007FFD5E1B1000  R--  Image    UMPDC.dll
0x00007FFD5E1B1000 - 0x00007FFD5E1BB000  R-X  Image    UMPDC.dll (.text)
0x00007FFD5E1BB000 - 0x00007FFD5E1BF000  R--  Image    UMPDC.dll (.rdata)
0x00007FFD5E1BF000 - 0x00007FFD5E1C0000  RW-  Image    UMPDC.dll (.data)
0x00007FFD5E1C0000 - 0x00007FFD5E1C4000  R--  Image    UMPDC.dll (.pdata)
0x00007FFD5E1C4000 - 0x00007FFD5E1C5000  R-X  Image    Unknown Image
0x00007FFD5E1D0000 - 0x00007FFD5E1D1000  R--  Image    powrprof.dll
0x00007FFD5E1D1000 - 0x00007FFD5E1E5000  R-X  Image    powrprof.dll (.text)
0x00007FFD5E1E5000 - 0x00007FFD5E1F0000  R--  Image    powrprof.dll (.rdata)
0x00007FFD5E1F0000 - 0x00007FFD5E1F1000  RW-  Image    powrprof.dll (.data)
0x00007FFD5E1F1000 - 0x00007FFD5E22E000  R--  Image    powrprof.dll (.pdata)
0x00007FFD5E22E000 - 0x00007FFD5E22F000  R-X  Image    Unknown Image
0x00007FFD5E350000 - 0x00007FFD5E351000  R--  Image    bcryptprimitives.dll
0x00007FFD5E351000 - 0x00007FFD5E3C7000  R-X  Image    bcryptprimitives.dll (.text)
0x00007FFD5E3C7000 - 0x00007FFD5E3E1000  R--  Image    bcryptprimitives.dll (.rdata)
0x00007FFD5E3E1000 - 0x00007FFD5E3E2000  RW-  Image    bcryptprimitives.dll (.data)
0x00007FFD5E3E2000 - 0x00007FFD5E3E9000  R--  Image    bcryptprimitives.dll (.pdata)
0x00007FFD5E3E9000 - 0x00007FFD5E3EA000  R-X  Image    Unknown Image
0x00007FFD5E840000 - 0x00007FFD5E841000  R--  Image    KERNELBASE.dll
0x00007FFD5E841000 - 0x00007FFD5E9ED000  R-X  Image    KERNELBASE.dll (.text)
0x00007FFD5E9ED000 - 0x00007FFD5EBDC000  R--  Image    KERNELBASE.dll (.rdata)
0x00007FFD5EBDC000 - 0x00007FFD5EBE4000  RW-  Image    KERNELBASE.dll (.data)
0x00007FFD5EBE4000 - 0x00007FFD5EBE6000  ???  Image    KERNELBASE.dll (.data)
0x00007FFD5EBE6000 - 0x00007FFD5EC33000  R--  Image    KERNELBASE.dll (.pdata)
0x00007FFD5EC33000 - 0x00007FFD5EC34000  R-X  Image    Unknown Image
0x00007FFD5ECD0000 - 0x00007FFD5ECD1000  R--  Image    ucrtbase.dll
0x00007FFD5ECD1000 - 0x00007FFD5EDC8000  R-X  Image    ucrtbase.dll (.text)
0x00007FFD5EDC8000 - 0x00007FFD5EE07000  R--  Image    ucrtbase.dll (.rdata)
0x00007FFD5EE07000 - 0x00007FFD5EE0A000  RW-  Image    ucrtbase.dll (.data)
0x00007FFD5EE0A000 - 0x00007FFD5EE1B000  R--  Image    ucrtbase.dll (.pdata)
0x00007FFD5EE1B000 - 0x00007FFD5EE1C000  R-X  Image    Unknown Image
0x00007FFD5F7C0000 - 0x00007FFD5F7C1000  R--  Image    RPCRT4.dll
0x00007FFD5F7C1000 - 0x00007FFD5F89D000  R-X  Image    RPCRT4.dll (.text)
0x00007FFD5F89D000 - 0x00007FFD5F8C3000  R--  Image    RPCRT4.dll (.rdata)
0x00007FFD5F8C3000 - 0x00007FFD5F8C5000  RW-  Image    RPCRT4.dll (.data)
0x00007FFD5F8C5000 - 0x00007FFD5F8D8000  R--  Image    RPCRT4.dll (.pdata)
0x00007FFD5F8D8000 - 0x00007FFD5F8D9000  R-X  Image    Unknown Image
0x00007FFD60470000 - 0x00007FFD60471000  R--  Image    KERNEL32.DLL
0x00007FFD60471000 - 0x00007FFD604F7000  R-X  Image    KERNEL32.DLL (.text)
0x00007FFD604F7000 - 0x00007FFD6052F000  R--  Image    KERNEL32.DLL (.rdata)
0x00007FFD6052F000 - 0x00007FFD60531000  RW-  Image    KERNEL32.DLL (.data)
0x00007FFD60531000 - 0x00007FFD60539000  R--  Image    KERNEL32.DLL (.pdata)
0x00007FFD60539000 - 0x00007FFD6053A000  R-X  Image    Unknown Image
0x00007FFD606C0000 - 0x00007FFD606C1000  R--  Image    psapi.dll
0x00007FFD606C1000 - 0x00007FFD606C2000  R-X  Image    psapi.dll (.text)
0x00007FFD606C2000 - 0x00007FFD606C4000  R--  Image    psapi.dll (.rdata)
0x00007FFD606C4000 - 0x00007FFD606C5000  RW-  Image    psapi.dll (.data)
0x00007FFD606C5000 - 0x00007FFD606C8000  R--  Image    psapi.dll (.pdata)
0x00007FFD611C0000 - 0x00007FFD611C1000  R--  Image    ntdll.dll
0x00007FFD611C1000 - 0x00007FFD61335000  R-X  Image    ntdll.dll (.text)
0x00007FFD61335000 - 0x00007FFD6138E000  R--  Image    ntdll.dll (.rdata)
0x00007FFD6138E000 - 0x00007FFD61398000  RW-  Image    ntdll.dll (.data)
0x00007FFD61398000 - 0x00007FFD61429000  R--  Image    ntdll.dll (.pdata)
0x00007FFD61429000 - 0x00007FFD6142A000  R-X  Image    Unknown Image
```


We can now also see proper section names in the fifth column, including:
- `.text` - Executable code (R-X protection)
- `.rdata` - Read-only data like string constants (R-- protection)
- `.data` - Initialized global/static variables (RW- protection)
- `.pdata` - Exception handling metadata (R-- protection)
- `.idata` - Import table (DLL imports)
- `.reloc` - Relocation information


#### Extra Credit - Understanding New Memory Scanner

Though we will cover interpretation of memory regions in greater depth later, it might be worth taking the time to review the code above to understand how we were able to interpret different regions. But consider this "extra credit" - it's not expected that you understand everything at this point. But I think that familiarizing yourself with some of the key concepts, early on, even if only in a broad sense, will serve you in the long run.

##### Adding Module Identification

The enhancement uses Windows APIs to enumerate all loaded modules in the process. `EnumProcessModules` returns handles to every DLL and executable, then `GetModuleBaseNameW` and `GetModuleInformation` provide their names, base addresses, and sizes. By storing this information, we can match any Image memory region to its source module by checking if the region's address falls within a module's address range.

##### Adding PE Section Parsing

To identify specific sections like `.text` (code) or `.data` (variables), we parse the PE (Portable Executable) file format directly from memory (more on this in the next section, Part 4). Each module starts with a DOS header containing the "MZ" signature, which points to the PE header. The PE header includes a section table listing all sections with their names, virtual addresses, and sizes.

We read these structures from memory using unsafe pointers, extract the section information, and store each section's address range. The key is using `FileHeader.SizeOfOptionalHeader` to correctly locate the section table, and validating that section names contain only printable ASCII characters.

##### How Interpretation Works

```
Program Start
    ↓
EnumerateModules()  ← Builds module list with PE parsing
    ↓
EnumerateMemory()   ← Scans address space
    ↓
For each region → identifyRegion() → Lookup in module list
    ↓
Print with interpretation column
```

The result is a memory map that shows not just anonymous addresses, but meaningful context: your executable's code section, ntdll's data section, kernel structures, heaps, and stacks - all clearly labeled despite ASLR randomization.





### Why VAD Enumeration Matters: Offensive and Defensive Perspectives

Understanding and enumerating VADs is crucial for both attackers and defenders. The VAD tree reveals the entire memory architecture of a process, exposing both opportunities and threats.

#### Offensive Use Cases

**1. Finding Injection Targets**

When injecting code into a remote process, you need executable memory. VAD enumeration reveals:

```
Looking for: R-X or RWX regions in target process
Found: 0x76D00000 - 0x76E00000  R-X  Image (ntdll.dll)
```

**2. Detecting Anti-Analysis**

Security tools often inject DLLs or allocate private executable memory:

```
Suspicious pattern detected:
0x30000000 - 0x30010000  RWX  Private  ← Security monitoring code?
0x40000000 - 0x40005000  R-X  Image    ← Unknown DLL injection?
Action: Identify and potentially bypass or manipulate these regions
```

**3. Code Cave Discovery**

PE files often contain unused space in executable sections:

```
Scan for: Null bytes or padding in R-X Image regions
0x00420000 - 0x00430000  R-X  Image
  → Contains 0x2000 bytes of 0x00 at offset +0x8000
  → Potential hiding spot for shellcode
```

**4. Bypassing DEP/ASLR**

VAD enumeration reveals executable memory locations even with ASLR:

```
ASLR randomizes base addresses, but VAD scan finds actual locations:
ntdll.dll base: 0x76D00000 (random)
kernel32.dll base: 0x77000000 (random)
→ Use these for ROP gadget hunting
```

#### Defensive Use Cases

**1. Detecting Suspicious Allocations**

EDR and antivirus tools monitor VAD changes for malicious patterns:

```
⚠️ ALERT: New private RWX allocation detected
Process: notepad.exe
Region: 0x50000000 - 0x50001000  RWX  Private
Action: Likely code injection → Flag for investigation
```

**2. Memory Forensics**

Incident responders use VAD enumeration to find injected code:

```
Expected: System DLLs (ntdll, kernel32, kernelbase)
Unexpected: 0x30000000 - 0x30005000  R-X  Private
Analysis: Injected code, not backed by legitimate PE file
```

**3. Integrity Checking**

Security software can baseline normal VAD layouts:

```
Baseline (clean process):
  - .exe at 0x00400000
  - ntdll, kernel32, standard DLLs
  - Private heap/stack regions

Current (suspicious):
  - .exe at 0x00400000 ✓
  - ntdll, kernel32 ✓
  - Unknown Image at 0x10000000 ✗ ← Injected DLL!
```

**4. Behavioral Analysis**

Tracking VAD changes over time reveals process behavior:

```
Time T0: 50 VAD entries (normal startup)
Time T1: 52 VAD entries (loaded plugin DLL)
Time T2: 75 VAD entries (massive allocation spike)
         → 20 new Private RW- regions
         → Potential unpacking or shellcode allocation
```














---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./thread.md" >}})
[|NEXT|]({{< ref "./security.md" >}})