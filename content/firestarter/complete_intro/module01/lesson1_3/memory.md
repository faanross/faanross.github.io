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


### The Illusion: What Each Process Sees

Consider two processes running simultaneously - perhaps two instances of Notepad, or Chrome and Word running side by side. From the perspective of **Process A**, the world looks simple and orderly: its **code** resides at **virtual address 0x00400000**, its **heap** (where dynamic allocations live) starts at **0x10000000**, and its **stack** sits near the top of user space at **0x7FF00000**. These addresses feel concrete and permanent to the process - when the program dereferences a pointer to `0x00400000`, it gets its code, every single time.

Now consider **Process B**, running concurrently on the same machine. Remarkably, it also sees its **code at 0x00400000**, its **heap at 0x10000000**, and its **stack at 0x7FF00000** - the **same addresses** as Process A! This seems impossible: how can two different programs occupy the same memory locations simultaneously? The answer is that these are **virtual addresses**, not real locations in **physical RAM**. Each process operates within its own private virtual address space, completely isolated and unaware of the others, seeing a pristine and exclusive view of memory that exists only from its perspective.

### The Reality: Physical Memory Layout

Behind the scenes, the truth is quite different. **Physical RAM** - the actual silicon memory chips in your computer - is a single, shared resource that all processes must use cooperatively. When we peer behind the virtual memory curtain, we see that **Process A's code** might actually reside at **physical address 0x00100000**, while **Process B's code** is located at an entirely different location: **physical address 0x00500000**. Similarly, **Process A's heap** might occupy **physical address 0x01000000**, while **Process B's heap** is off at **0x02000000**. The physical addresses are scattered throughout RAM based on where the operating system's memory manager found available space, bearing no resemblance to the tidy virtual layout each process perceives.

This separation means that when Process A writes to its virtual address `0x10000000` (its heap), it's actually modifying physical memory at `0x01000000`, while Process B writing to its virtual address `0x10000000` is modifying a completely different region at physical address `0x02000000`. The two processes can use identical virtual addresses without conflict because those addresses don't directly reference physical memory - they're simply numbers that will be translated before reaching actual RAM.

### The Magic: Hardware-Assisted Translation

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















---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./thread.md" >}})
[|NEXT|]({{< ref "./securityA.md" >}})