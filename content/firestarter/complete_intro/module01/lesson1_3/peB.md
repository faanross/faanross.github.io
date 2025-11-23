---
showTableOfContents: true
title: "The PE File Format: The 7 Layers"
type: "page"
---



## Layer 1: DOS Header and Stub - The Backward Compatibility Gateway

### The DOS Header: Your First Checkpoint

Every PE starts with a 64-byte DOS header containing two things we care about:

1. **Magic number `MZ` (0x5A4D)**: The first validation checkpoint. Without it, Windows won't even try to load the file. Every PE parser - legitimate or malicious - checks this signature first.

2. **`e_lfanew` field** (at offset 0x3C): Contains the file offset to the actual PE headers. This is your bridge from legacy DOS structures to modern PE format.

### The DOS Stub: Exploitable Dead Space

Between the DOS header and PE headers sits a small DOS program that displays "This program cannot be run in DOS mode" if executed in MS-DOS.

**Why we care:** This is writable space between critical headers that many parsers skip. Some packers hide data here. Some malware uses custom stubs to detect analysis environments. It's also a convenient place to hide small payloads that static scanners might ignore.

## Layer 2: PE Signature and File Header - Essential Characteristics

### The PE Signature

After following `e_lfanew`, you find the 4-byte sequence `PE\0\0`. This is validation checkpoint #2.

### The File Header: What You're Dealing With

This 20-byte structure tells you fundamental facts about the executable:

**Machine** (2 bytes): Target CPU architecture
- `0x014C` = x86 (32-bit Intel)
- `0x8664` = x64 (64-bit AMD/Intel)
- `0xAA64` = ARM64

**Why we care:** Your shellcode must match this architecture. Mismatches cause immediate crashes.

**NumberOfSections** (2 bytes): How many sections follow (.text, .data, etc.)

**Why we care:** When adding sections to hide code, you increment this. When parsing, you use it to iterate through sections.

**TimeDateStamp** (4 bytes): Unix timestamp of compilation

**Why we care:** Forensic indicator of compromise (IOC). Zero often indicates packing or timestamp manipulation. Can be forged to appear legitimate.

**Characteristics** (2 bytes): Bitfield describing file properties

Key flags we care about:
- `IMAGE_FILE_EXECUTABLE_IMAGE` (0x0002): Valid executable
- `IMAGE_FILE_DLL` (0x2000): This is a DLL, not an EXE
- `IMAGE_FILE_RELOCS_STRIPPED` (0x0001): **Critical** - no relocations present, cannot rebase

**Why we care:** The `RELOCS_STRIPPED` flag means the PE can only load at its preferred address. If that address is taken (common with ASLR), loading fails. This breaks process hollowing and reflective loading if you can't allocate at the exact ImageBase.


## Layer 3: Optional Header - The Loading Blueprint

Despite its name, this header is **required** for all executables. It contains the most critical information for loading and executing the image. Size differs between 32-bit (224 bytes) and 64-bit (240 bytes).

### Fields That Matter for Offensive Operations

**Magic** (2 bytes): Format identifier
- `0x010B` = PE32 (32-bit)
- `0x020B` = PE32+ (64-bit)

**AddressOfEntryPoint** (4 bytes, RVA): Where execution begins

**Why we care:** In process hollowing, you modify this to redirect execution to your payload. In backdooring, you point this to your code first, then jump to the original entry point. This single field controls where the process starts executing.

**ImageBase** (8 bytes on x64): Preferred load address

**Why we care:** With ASLR, Windows rarely loads here. When manually loading PEs (reflective DLL injection), you try to allocate at ImageBase but must handle failure. If you load elsewhere and the PE has relocations, you must apply fixups. If no relocations exist, you're stuck - must allocate at exactly this address or fail.

**SectionAlignment** (4 bytes): Memory alignment (typically 0x1000 = 4KB)
**FileAlignment** (4 bytes): Disk alignment (typically 0x200 = 512 bytes)

**Why we care:** When manually loading, you must honor section alignment in memory. When parsing, you use these to convert between file offsets and RVAs.

**SizeOfImage** (4 bytes): Total size when loaded in memory

**Why we care:** This is what you pass to `VirtualAlloc` when manually loading a PE. Too small and you'll corrupt memory; the loader needs this exact value.

**DllCharacteristics** (2 bytes): Security features bitfield

Key flags:
- `IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE` (0x0040): ASLR supported
- `IMAGE_DLLCHARACTERISTICS_NX_COMPAT` (0x0100): DEP compatible
- `IMAGE_DLLCHARACTERISTICS_NO_SEH` (0x0400): No exception handlers

**Why we care:** These flags tell you what security features the binary supports. No ASLR means predictable memory layout (easier exploitation). No DEP means code can execute from data pages. When crafting malicious PEs, you might deliberately disable these for operational convenience.

**DataDirectory** (16 entries): Table of contents for specialized structures

Each entry contains an RVA and size pointing to critical tables. This is perhaps the most important field in the entire optional header.

## Layer 4: Data Directories - Your Roadmap

The DataDirectory array is a 16-element index pointing to specialized structures. Think of it as a table of contents for everything important in the PE.

### The Entries That Matter

**[0] Export Directory**: Lists functions this DLL exports

**Why we care:** When manually resolving APIs (GetProcAddress implementation), you parse this. When implementing API hashing or direct syscalls, you enumerate exports to find target functions.

**[1] Import Directory**: Lists DLLs and functions this PE imports

**Why we care:** **This is critical**. Static analysis tools scan imports to identify suspicious behavior. If your "calculator.exe" imports `CreateRemoteThread`, `WriteProcessMemory`, and `VirtualAllocEx`, that's a red flag. We bypass this through dynamic API resolution, API hashing, or direct syscalls - all of which require understanding how imports work.

**[2] Resource Directory**: Icons, strings, dialogs, embedded files

**Why we care:** Common hiding spot for payloads. You can embed encrypted shellcode as fake icons or version info. Resource compilers make this easy, and some AVs scan resources less thoroughly than code sections.

**[5] Base Relocation Directory**: Fixup data for ASLR

**Why we care:** **Essential for process hollowing and reflective loading**. When your PE can't load at ImageBase (common), you must apply these relocations to fix hard-coded addresses. If this directory is missing (`RELOCS_STRIPPED`), you can't rebase - major problem for injection techniques.

**[6] Debug Directory**: Path to PDB file, debug info

**Why we care:** Can leak developer paths and build environment details. When stripping binaries, you might remove this. When analyzing malware, it sometimes contains useful forensic information.

**[9] TLS (Thread Local Storage) Directory**: Thread-local data and callbacks

**Why we care:** **TLS callbacks execute before the entry point** - before debuggers typically break. Packers use this to run anti-debugging checks before you can attach. When analyzing suspicious PEs, always check for TLS callbacks.

**[12] Import Address Table (IAT)**: Runtime function pointers

**Why we care:** At runtime, this contains actual memory addresses of imported functions. EDR products hook entries here to intercept API calls. IAT hooking is a fundamental offensive and defensive technique. Understanding IAT structure is prerequisite to bypassing these hooks.

## Layer 5: Section Headers - The Memory Map

After the optional header comes an array of section headers, one per section. Each header describes a logical division of the executable.

### Standard Sections and What They Mean

**.text**: Compiled machine code

- **Permissions**: Read + Execute (RX)
- **Why we care**: This is where legitimate code lives. EDR heavily monitors execution from non-.text sections. When injecting shellcode, getting it into .text makes it less suspicious. Self-modifying code requires changing permissions from RX to RWX (detectable).

**.data**: Initialized global/static variables with values

- **Permissions**: Read + Write (RW)
- **Why we care**: Writable, so good for storing decryption keys or small payloads at runtime. Some packers encrypt this section and decrypt it on load.

**.rdata**: Read-only data, string literals, **and the IAT**

- **Permissions**: Read only (R)
- **Why we care**: The Import Address Table lives here. To hook IAT entries, you must change page permissions from R to RW first (detectable). String literals here can leak information about functionality.

**.bss**: Uninitialized global/static variables

- **Permissions**: Read + Write (RW)
- **Special property**: Takes zero space on disk, allocated only in memory
- **Why we care**: Often ignored by static analysis tools. Good for runtime allocation of malicious structures.

**.rsrc**: Resources (icons, dialogs, strings, embedded files)

- **Permissions**: Read only (R)
- **Why we care**: Common payload hiding spot. Embed encrypted shellcode as resources. Some AVs scan less aggressively here.

**.reloc**: Base relocation table for ASLR

- **Permissions**: Read only, marked discardable
- **Why we care**: Essential for process hollowing and reflective loading. Without this, PE cannot rebase. Loader discards this section after applying relocations to save memory.

### Section Characteristics: The Security Story

Each section header contains a **Characteristics** field that tells Windows how to protect it in memory:

- `IMAGE_SCN_MEM_EXECUTE` (0x20000000): Executable
- `IMAGE_SCN_MEM_READ` (0x40000000): Readable
- `IMAGE_SCN_MEM_WRITE` (0x80000000): Writable

**Why we care about permissions:**

**RWX sections** (Read + Write + Execute) are major red flags. They allow self-modifying code and bypass DEP naturally. Legitimate software rarely needs RWX sections. Malware uses them for runtime packing/unpacking and code obfuscation.

**Custom section names** (not .text, .data, etc.) often indicate packing. Tools like UPX, Themida, and custom packers create sections with distinctive names (.upx, .themida, .packed).

**Section manipulation** is a core offensive technique:
- Add new sections to hide malicious code
- Change section permissions to enable code injection
- Encrypt sections and decrypt at runtime
- Hide payloads in unused section space


## Layer 6: Import Address Table - The API Connection

The IAT is how Windows executables call functions in external DLLs. Understanding it is **absolutely critical** because:

1. Static analysis scans IAT to profile behavior
2. EDR hooks IAT entries to intercept suspicious calls
3. Dynamic API resolution bypasses IAT entirely (key evasion)
4. IAT manipulation is fundamental to both offense and defense

### How Imports Work: The Loading Process

```
THE COMPLETE IMPORT RESOLUTION PROCESS:

1. LOADER READS DATA DIRECTORY[1] (Import Directory)
   ┌─────────────────────────────────────────────┐
   │ Points to array of IMAGE_IMPORT_DESCRIPTOR  │
   │ One descriptor per imported DLL             │
   └─────────────────────────────────────────────┘

2. FOR EACH IMPORT DESCRIPTOR:
   ┌─────────────────────────────────────────────┐
   │ Read DLL name: "KERNEL32.dll"               │
   │ Load DLL into process: LoadLibrary()        │
   │   - If DLL not loaded, load it now          │
   │   - Increment reference count if present    │
   └─────────────────────────────────────────────┘

3. FOR EACH FUNCTION FROM THIS DLL:
   ┌─────────────────────────────────────────────┐
   │ Read from Import Name Table (INT)           │
   │   - Either: Function name "CreateFileW"     │
   │   - Or: Ordinal number (e.g., #42)          │
   │                                             │
   │ Resolve address: GetProcAddress()           │
   │   - Search DLL's export table               │
   │   - Get actual memory address of function   │
   │                                             │
   │ Write address to Import Address Table (IAT) │
   │   - IAT entry now contains real function ptr│
   └─────────────────────────────────────────────┘

4. EXECUTION TIME:
   ┌─────────────────────────────────────────────┐
   │ Your code: CALL [IAT entry]                 │
   │   ↓                                         │
   │ Indirect jump through IAT                   │
   │   ↓                                         │
   │ Execute actual DLL function                 │
   └─────────────────────────────────────────────┘

BEFORE RESOLUTION (on disk):          AFTER RESOLUTION (in memory):
┌──────────────┐                      ┌──────────────┐
│ IAT Entry 1  │ → 0x00000000         │ IAT Entry 1  │ → 0x7FFE0001234
│ IAT Entry 2  │ → 0x00000000         │ IAT Entry 2  │ → 0x7FFE0005678
│ IAT Entry 3  │ → 0x00000000         │ IAT Entry 3  │ → 0x7FFE000ABCD
└──────────────┘                      └──────────────┘
   (Empty pointers)                      (Real addresses)
```

### Why Imports Reveal Your Intentions

Static analysis tools simply enumerate your IAT to see what APIs you use:

- `CreateRemoteThread`, `WriteProcessMemory`, `VirtualAllocEx` = Process injection
- `NtQuerySystemInformation`, `NtReadVirtualMemory` = Process enumeration
- `RegSetValueEx`, `CreateService` = Persistence mechanisms
- `socket`, `connect`, `send` = Network communication (C2)
- `GetProcAddress`, `LoadLibrary` = Dynamic API resolution (evasion)

**The problem:** If your "document_viewer.exe" imports these functions, it's obviously suspicious.

**The solutions** (which we'll implement later):
- **Dynamic API resolution**: Resolve functions at runtime with `GetProcAddress` - only benign imports visible statically
- **API hashing**: Resolve functions by hash of their name - no strings visible
- **Direct syscalls**: Bypass Win32 APIs entirely, invoke system calls directly - no imports from ntdll.dll
- **Delay-loaded imports**: Mark imports as delay-loaded so they don't appear in standard IAT enumeration

Understanding IAT structure is the prerequisite for all these evasion techniques.

## Layer 7: Relocations - Making Position-Independent Code

Modern Windows uses ASLR for security - loading executables at random addresses rather than their preferred ImageBase. When a PE can't load at its preferred address, Windows must apply **relocations** - fixups to hard-coded addresses embedded in the code.

### Why Relocations Matter

Consider this assembly:
```
MOV RAX, [0x140001000]  ; Load from fixed address
```

If PE loads at preferred base `0x140000000`, address `0x140001000` is correct.

If ASLR loads it at `0x7FF800000000`, address `0x140001000` is wrong - causes crash or reads garbage.

Relocations tell the loader: "These addresses need adjustment if you load me elsewhere."

### The Relocation Process

```
RELOCATION ALGORITHM:

1. LOADER ATTEMPTS TO LOAD AT ImageBase (preferred address)
   ┌─────────────────────────────────────────┐
   │ Try: VirtualAlloc(ImageBase, ...)       │
   └─────────────────────────────────────────┘

2. IF ADDRESS IS AVAILABLE:
   ┌─────────────────────────────────────────┐
   │ Load at ImageBase                       │
   │ Delta = 0                               │
   │ No relocations needed ✓                 │
   └─────────────────────────────────────────┘

3. IF ADDRESS IS TAKEN (ASLR, DLL collision):
   ┌─────────────────────────────────────────┐
   │ Choose new base: NewBase                │
   │ Calculate: Delta = NewBase - ImageBase  │
   │ Must apply relocations →                │
   └─────────────────────────────────────────┘

4. CHECK FOR RELOCATION DIRECTORY:
   ┌─────────────────────────────────────────┐
   │ Read DataDirectory[5] (Base Relocation) │
   │ If RVA == 0: NO RELOCATIONS ✗           │
   │   → Cannot rebase, loading fails        │
   │ If RVA != 0: Process relocations →      │
   └─────────────────────────────────────────┘

5. FOR EACH RELOCATION BLOCK:
   ┌─────────────────────────────────────────┐
   │ Read block header (RVA, Size)           │
   │ For each relocation entry in block:     │
   │   - Read type and offset                │
   │   - Calculate address to fix            │
   │   - Read original value at address      │
   │   - Add delta to value                  │
   │   - Write back corrected value          │
   └─────────────────────────────────────────┘

CONCRETE EXAMPLE:

Preferred ImageBase: 0x0000000140000000
Actual load address: 0x00007FF800000000
Delta = 0x00007FF800000000 - 0x0000000140000000 = 0x00007FF6C0000000

Original code at RVA 0x1050:
  MOV RAX, [0x0000000140003000]  ; Hard-coded absolute address

Relocation entry says: "Fix address at RVA 0x1052" (the operand)

Fixup process:
  1. Read 8 bytes at RVA 0x1052: 0x0000000140003000
  2. Add delta: 0x0000000140003000 + 0x00007FF6C0000000 
     = 0x00007FF800003000
  3. Write back: 0x00007FF800003000

Fixed code:
  MOV RAX, [0x00007FF800003000]  ; Now points to correct address
```

### Relocation Directory Structure

The relocation directory consists of variable-sized **blocks**, one per 4KB page:

```
Relocation Directory:
┌────────────────────────────────────────────────┐
│ Block 1: Page at RVA 0x1000                    │
│   Header: VirtualAddress=0x1000, Size=0x14     │
│   Entries:                                     │
│     Type=DIR64, Offset=0x028 → Fix at 0x1028   │
│     Type=DIR64, Offset=0x130 → Fix at 0x1130   │
│     Type=DIR64, Offset=0x248 → Fix at 0x1248   │
│     Type=ABSOLUTE, Offset=0x000 (padding)      │
├────────────────────────────────────────────────┤
│ Block 2: Page at RVA 0x2000                    │
│   Header: VirtualAddress=0x2000, Size=0x0C     │
│   Entries:                                     │
│     Type=DIR64, Offset=0x010 → Fix at 0x2010   │
│     Type=DIR64, Offset=0xA48 → Fix at 0x2A48   │
├────────────────────────────────────────────────┤
│ ...more blocks...                              │
└────────────────────────────────────────────────┘

Each block covers one 4KB page for memory locality
```

Each relocation entry is a 16-bit value:
- **Upper 4 bits**: Type (most common: DIR64 for x64, HIGHLOW for x86)
- **Lower 12 bits**: Offset within the page (0-4095)

### When Relocations Matter for Offensive Operations

**Process Hollowing:**
1. Create suspended legitimate process
2. Unmap its image with `NtUnmapViewOfSection`
3. Allocate memory for your malicious PE
4. **If allocated address ≠ ImageBase**: Must apply relocations
5. If PE has no relocations: Must allocate at exact ImageBase (may fail)

**Reflective DLL Injection:**
When manually loading a DLL from memory, you reimplement the Windows loader:
1. Allocate memory (prefer ImageBase, accept anything)
2. Copy headers and sections
3. **Apply relocations if loaded at different address**
4. Resolve imports
5. Fix section permissions
6. Call DllMain

**Without relocations**, both techniques become much harder - you're forced to allocate at a specific address that might not be available.

### ASLR Detection

Binaries compiled without ASLR support lack relocations and always load at the same address:

- `IMAGE_FILE_RELOCS_STRIPPED` flag set = No relocations present
- `IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE` not set = ASLR disabled

**Why this matters:** Predictable memory layout makes exploitation easier. When analyzing targets, you look for these flags to identify vulnerable binaries.

## Summary: What We Really Care About

From an offensive security perspective, these are the critical PE components:

| Component | Why It Matters | Offensive Use Cases |
|-----------|---------------|---------------------|
| **Entry Point** | Where execution starts | Process hollowing, backdooring, control flow hijacking |
| **ImageBase** | Preferred load address | Manual PE loading, relocation handling, ASLR analysis |
| **SizeOfImage** | Total memory footprint | Allocating space for manual loading |
| **Import Directory** | API dependencies visible statically | Hiding intentions via dynamic resolution, identifying targets for hooking |
| **IAT** | Runtime function addresses | API hooking, understanding how to bypass hooks |
| **Relocation Directory** | Position-independence support | Essential for process hollowing and reflective loading |
| **Section Headers** | Memory layout and permissions | Code injection, finding writable/executable regions |
| **DllCharacteristics** | Security features (ASLR, DEP) | Identifying vulnerable targets, understanding protections |
| **TLS Directory** | Pre-entry point callbacks | Packer anti-debugging, early execution for evasion |
| **Resources** | Embedded data | Payload hiding, masquerading as legitimate files |




## What's Next

In upcoming modules, we'll leverage this knowledge to implement actual offensive capabilities:

- **Manual PE loading from memory** (reflective DLL injection)
- **Import obfuscation** (dynamic resolution, API hashing, syscalls)
- **Process hollowing** with proper relocation handling
- **Section manipulation** for code injection
- **IAT hooking** for interception and persistence
- **Resource embedding** for payload concealment

Understanding PE structure is the foundation - every advanced injection technique, every evasion method, and every in-memory attack builds on this knowledge. You can't manipulate what you don't understand, and you can't hide from defenses if you don't know what they're looking for.

The PE format isn't just a file format specification - it's the instruction manual for Windows executables, and mastering it is prerequisite to offensive Windows development.




___


## Lab 2: Manual PE Analysis - Finding Critical Values

### Objective

Now that you understand PE structure, let's systematically locate every critical field we've discussed. You'll use PEBear to find the values that matter for offensive security operations, understanding exactly where each piece of information lives and why it's important.

### Setup

- **Tool:** PEBear
- **Target:** `simple.exe` from Lab 1 (or any PE executable)
- **Duration:** ~20 minutes
- **Deliverable:** Complete the analysis table at the end

---

### Analysis Checklist

We'll go through the PE structure layer by layer, finding only the values that matter for offensive operations.

---

### Layer 1: DOS Header - The Gateway

#### What We're Looking For
- DOS signature (validation checkpoint #1)
- `e_lfanew` (bridge to PE headers)

#### Steps

1. Open `simple.exe` in PEBear
2. In the middle panel, click the **DOS Hdr** tab.

**Find These Values:**

| Field        | Location    | Value           | Why It Matters                             |
| ------------ | ----------- | --------------- | ------------------------------------------ |
| **e_magic**  | Offset 0x00 | `0x5A4D` ("MZ") | First validation - no MZ, no load          |
| **e_lfanew** | Offset 0x3C | `________`      | Points to PE headers - your bridge forward |

**Verification:**
- `e_magic` should always be `0x5A4D` (MZ in ASCII)
- `e_lfanew` is typically `0x80` to `0x100` for modern PEs


**My Result:**

![dos header](../img/dos_header.png)

- Right at the top you can see at Offset `0` we have `Magic number` with a predicted value of `5A4D` - what we also expect it to be.
- Then right at the bottom we have, at Offset `3C`, `File address of the new exe header`, which is `e_lfanew.` In my case the Value is `80`.

**Offensive Relevance:**
- Every PE parser checks `e_magic` first
- `e_lfanew` tells you where the real headers begin
- DOS stub between header and PE can hide small payloads

---

## Layer 2: File Header - Core Characteristics

### What We're Looking For
- Architecture (x86/x64)
- Number of sections
- Timestamp (forensic IOC)
- Characteristics (file type, relocation status)

### Steps

1. Click **File Hdr** tab in the middle panel

**Find These Values:**

| Field                      | Your Value | Interpretation                         |
| -------------------------- | ---------- | -------------------------------------- |
| **Machine**                | `________` | `0x014C` = x86, `0x8664` = x64         |
| **Sections Count**         | `________` | How many sections (.text, .data, etc.) |
| **Time Date Stamp**        | `________` | Compile time (Unix timestamp)          |
| **Size of OptionalHeader** | `________` | Should be 224 (x86) or 240 (x64)       |
| **Characteristics**        | `________` | Bitfield - see decode below            |

**My Results:**
![file header](../img/file_header.png)


| Field                      | My Value                              |
| -------------------------- | ------------------------------------- |
| **Machine**                | `8664`                                |
| **Sections Count**         | `8`                                   |
| **Time Date Stamp**        | `0`                                   |
| **Size of OptionalHeader** | `240`                                 |
| **Characteristics**        | `2 - File is executable`              |
|                            | `20 - App can handle > 2gb addresses` |

**Time Date Stamp:**


We can see in Time Date Stamp we do not see the Unix Timestamp for compile time, but 0 - why? This is actually **expected behavior** for Go binaries.

Starting with Go 1.10 (and refined in later versions), the Go toolchain implements **reproducible builds** by default. This means that compiling the same source code multiple times produces bit-for-bit identical binaries. To achieve this, Go intentionally sets the PE timestamp to 0 instead of embedding the actual compilation time.





**Decode Characteristics :**

For the Characteristics, there are 4 flags we expect to potentially encounter:

| Flag Value | Flag Name           | Meaning                           |
| ---------- | ------------------- | --------------------------------- |
| `0x0002`   | EXECUTABLE_IMAGE    | Valid executable file             |
| `0x2000`   | DLL                 | This is a DLL (not EXE)           |
| `0x0001`   | RELOCS_STRIPPED     | ⚠️ NO relocations - can't rebase! |
| `0x0020`   | LARGE_ADDRESS_AWARE | Can handle >2GB addresses         |

You can see in my case above, and most likely in your case as well, the value is `0x0022`:
- `0x0022` & `0x0002` = `0x0002` ✓ EXECUTABLE_IMAGE is SET
- `0x0022` & `0x2000` = `0x0000` ✗ DLL is NOT SET
- `0x0022` & `0x0001` = `0x0000` ✗ RELOCS_STRIPPED is NOT SET (good!)
- `0x0022` & `0x0020` = `0x0020` ✓ LARGE_ADDRESS_AWARE is SET

Is `RELOCS_STRIPPED` set?
- **YES:** 🔴 Cannot rebase - process hollowing will fail if can't allocate at exact ImageBase
- **NO:** ✓ Has relocations - can load anywhere

Is this a DLL or EXE?
- **DLL:** Different loading considerations, exports instead of entry point focus
- **EXE:** Standard executable, focus on entry point

---





## Layer 3: Optional Header - The Loading Blueprint

### What We're Looking For
- Magic number (PE32 vs PE32+)
- Entry point (where execution starts)
- ImageBase (preferred load address)
- Size of image (memory allocation size)
- Section alignment values
- Security flags (ASLR, DEP)

### Steps

1. Click the **Optional Hdr** tab in the centre panel:

**Find These Values:**

| Field                   | Your Value | Why It Matters                                              |
| ----------------------- | ---------- | ----------------------------------------------------------- |
| **Magic**               | `________` | `0x010B` = PE32 (32-bit), `0x020B` = PE32+ (64-bit)         |
| **Entry Point**         | `________` | RVA where execution begins - target for backdooring         |
| **Image Base**          | `________` | Preferred load address - needed for relocation calculations |
| **Section Alignment**   | `________` | Memory alignment (typically `0x1000` = 4KB)                 |
| **File Alignment**      | `________` | Disk alignment (typically `0x200` = 512 bytes)              |
| **Size of Image**       | `________` | Total memory size - what you pass to VirtualAlloc           |
| **Size of Headers**     | `________` | Size of all headers combined                                |
| **Dll Characteristics** | `________` | Security features - decode below                            |



**My Results:**  
![optional header](../img/optional_header.png)



| Field                   | My Value    | Remarks                                                      |
| ----------------------- | ----------- | ------------------------------------------------------------ |
| **Magic**               | `20B`       | Indicates 64-bit                                             |
| **Entry Point**         | `7C620`     | RVA where execution begins                                   |
| **Image Base**          | `140000000` | Preferred load address - needed for relocation calculations  |
| **Section Alignment**   | `1000`      | Memory alignment is 4KB as expected                          |
| **File Alignment**      | `200`       | Disk alignment is 512 bytes as expected                      |
| **Size of Image**       | `84b000`    | Total memory size - what we pass to VirtualAlloc             |
| **Size of Headers**     | `600`       | Size of all headers combined                                 |
| **Dll Characteristics** | `8160`      |                                                              |
|                         | `20`        | Image can handle a high entropy 64-bit virtual address space |
|                         | `40`        | DLL can move (means ASLR is enabled)                         |
|                         | `1000`      | Image is NX compatible                                       |
|                         | `8000`      | TerminalServer aware                                         |





**Decode DllCharacteristics:**

| Flag Value | Flag Name             | Set? (Y/N) | Security Implication                       |
| ---------- | --------------------- | ---------- | ------------------------------------------ |
| `0x0020`   | HIGH_ENTROPY_VA       | `Y`        | High entropy ASLR for 64-bit address space |
| `0x0040`   | DYNAMIC_BASE          | `Y`        | ASLR enabled - loads at random address     |
| `0x0100`   | NX_COMPAT             | `Y`        | DEP enabled - data pages not executable    |
| `0x0400`   | NO_SEH                | `N`        | No exception handlers                      |
| `0x8000`   | TERMINAL_SERVER_AWARE | `Y`        | Terminal Server/RDS aware                  |




**Security Assessment:**

```
ASLR: ENABLED
- If disabled: Predictable memory layout (easier exploitation)

DEP: ENABLED
- If disabled: Can execute code from data pages

Overall: MODERN SECURITY
```



📝 **Entry Point RVA:** 7C620
- To find in memory: ImageBase + EntryPoint RVA
- Process hollowing: Change this to redirect execution
- Backdooring: Point to your code first, then jump to original

📝 **ImageBase:** `140000000`
- If ASLR enabled: Won't actually load here at runtime
- Manual PE loading: Try to allocate here, handle failure
- If relocations stripped AND can't allocate here: Loading fails

---






---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./pe.md" >}})
[|NEXT|]({{< ref "./peC.md" >}})