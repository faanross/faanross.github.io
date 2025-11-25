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




## Layer 4: Data Directories - Table of Contents

### What We're Looking For
- Import Directory (what APIs this PE uses)
- Export Directory (what APIs this PE provides)
- Resource Directory (icons, strings, potential payload hiding)
- Relocation Directory (can this PE rebase?)
- TLS Directory (pre-entry point execution)

### Steps

1. Still in **Optional Hdr**, scroll down to **Data Directories**


**Find These Values:**

| Index | Directory Name | RVA | Size | Present? | Purpose |
|-------|---------------|-----|------|----------|---------|
| **[0]** | Export | `________` | `________` | Y/N | Functions this DLL exports |
| **[1]** | Import | `________` | `________` | Y/N | 🔴 APIs this PE imports (critical!) |
| **[2]** | Resource | `________` | `________` | Y/N | Icons, strings, payload hiding spot |
| **[5]** | Base Relocation | `________` | `________` | Y/N | 🔴 Required for process hollowing |
| **[6]** | Debug | `________` | `________` | Y/N | PDB path, build info |
| **[9]** | TLS | `________` | `________` | Y/N | Pre-entry point callbacks |
| **[12]** | IAT | `________` | `________` | Y/N | Import Address Table location |


**My Results:**

![data directory](../img/data_directory.png)

| Index    | Directory Name  | RVA      | Size    | Present? | Remarks                                                       |
| -------- | --------------- | -------- | ------- | -------- | ------------------------------------------------------------- |
| **[0]**  | Export          | `0`      | `0`     | N        | No exported functions - this is an executable, not a library  |
| **[1]**  | Import          | `836000` | `53E`   | Y        | Imports Windows APIs - check these to understand PE behavior  |
| **[2]**  | Resource        | `0`      | `0`     | N        | No embedded resources (icons, strings, etc.)                  |
| **[5]**  | Base Relocation | `837000` | `12868` | Y        | Contains relocation data - supports ASLR functionality        |
| **[6]**  | Debug           | `0`      | `0`     | N        | No debug info - stripped for release build                    |
| **[9]**  | TLS             | `0`      | `0`     | N        | No Thread Local Storage callbacks                             |
| **[12]** | IAT             | `75E560` | `178`   | Y        | Import Address Table present - relatively small (few imports) |


**Analysis Questions:**

**Import Directory:**

- RVA: `836000` (if 0, no imports - suspicious!)
- Size: `53E`
- ✓ Present - this PE imports external functions

**Base Relocation:**

- RVA: `837000` (if 0, cannot rebase!)
- Size: `12868`
- Analysis: `CAN REBASE`
- Impact: `ASLR WORKS`

**TLS Directory:**

- RVA: `0` (if 0, no TLS callbacks)
- If present: ⚠️ TLS callbacks execute BEFORE entry point
- Offensive use: Packers use this for anti-debugging

**Resource Directory:**

- RVA: `0`
- Size: `0`
- Potential for payload hiding: `NO`

---



## Layer 5: Sections - The Memory Map

### What We're Looking For
- Standard sections (.text, .data, .rdata, etc.)
- Section permissions (RWX analysis)
- Section sizes (disk vs memory)
- Custom sections (packer indicators)

### Steps

1. Click **Section Hdrs** tab in the middle panel

**Map Your Sections:**

For each section, fill in this table:

| Section Name | VirtualAddr (RVA) | VirtualSize | RawAddr (File) | RawSize | Permissions |
|--------------|-------------------|-------------|----------------|---------|-------------|
| `.________` | `________` | `________` | `________` | `________` | `R__ / RW_ / R_X / RWX` |
| `.________` | `________` | `________` | `________` | `________` | `R__ / RW_ / R_X / RWX` |
| `.________` | `________` | `________` | `________` | `________` | `R__ / RW_ / R_X / RWX` |
| `.________` | `________` | `________` | `________` | `________` | `R__ / RW_ / R_X / RWX` |


**My Results:**

![section](../img/section.png)


| Section Name | VirtualAddr (RVA) | VirtualSize | RawAddr (File) | RawSize  | Permissions | Notes                |
| ------------ | ----------------- | ----------- | -------------- | -------- | ----------- | -------------------- |
| `.text`      | `1000`            | `3938F1`    | `600`          | `393A00` | `R_X`       | Code section         |
| `.rdata`     | `395000`          | `3C89C0`    | `394000`       | `3C8A00` | `R__`       | IAT lives here       |
| `.data`      | `75E000`          | `C00D0`     | `75CA00`       | `69800`  | `RW_`       | Global variables     |
| `.pdata`     | `81F000`          | `15474`     | `7C6200`       | `15600`  | `R__`       | Function unwind data |
| `.reloc`     | `837000`          | `12868`     | `7DC000`       | `12A00`  | `R__`       | Relocations          |

**Notes on permissions:**

- `.text` = R_X (Read + Execute) - Contains executable code
- `.rdata` = R__ (Read only) - Contains read-only data (constants, imports)
- `.data` = RW_ (Read + Write) - Contains initialized writable data
- `.pdata` = R__ (Read only) - Contains exception handling data (x64 specific)
- `.reloc` = R__ (Read only) - Contains base relocation table for ASLR



Note: Permissions are derive from the `Characteristics` column:

| Flag | Value | Meaning |
|------|-------|---------|
| IMAGE_SCN_MEM_READ | `0x40000000` | Readable (R) |
| IMAGE_SCN_MEM_WRITE | `0x80000000` | Writable (W) |
| IMAGE_SCN_MEM_EXECUTE | `0x20000000` | Executable (X) |


**Security Red Flags:**

Check for these suspicious characteristics:

🔴 **RWX Sections** (all three permissions):
- Section name: **None found**
- Why suspicious: Self-modifying code, runtime unpacking
- Common in: Malware, packers, obfuscators

⚠️ **Custom Section Names** (not standard .text, .data, etc.):
- Found: **NO**
- Names: **All standard (.text, .rdata, .data, .pdata, .reloc)**
- Why suspicious: Often indicates packing (UPX, Themida, etc.)

⚠️ **Unusually Large Sections**:
- Section: **.rdata**
- Size: **3C89C0 (≈3.9 MB virtual size)**
- Why suspicious: May contain hidden payload (though large .rdata is normal for Go binaries with embedded metadata)

**Calculate Section Alignment Differences:**

Pick one section (e.g., .text) and calculate:
```
Section: .text
Virtual Address (RVA):     0x1000
Raw Address (File Offset): 0x600
Difference:                0xA00 (2560 bytes)

Why they differ: Different alignment requirements
- Memory: aligned to SectionAlignment (typically 4KB)
- Disk: aligned to FileAlignment (typically 512 bytes)
```


---


## Layer 6: Imports - The API Dependencies

### What We're Looking For
- Which DLLs this PE imports from
- Which functions it imports (behavioral indicators)
- Suspicious API combinations

### Steps

1. Click **Imports** in the left panel
2. PEBear displays all imported DLLs and their functions

**Map Your Imports:**

List each imported DLL and note suspicious functions.


### DLL 1: `________________`

| Function Name | Suspicion Level | Why? |
|---------------|----------------|------|
| `________` | 🔴 HIGH / 🟡 MED / 🟢 LOW | ________ |
| `________` | 🔴 HIGH / 🟡 MED / 🟢 LOW | ________ |
| `________` | 🔴 HIGH / 🟡 MED / 🟢 LOW | ________ |

NOTE: Perform for each DLL and see if any of these functions are present:

| API Function | Present? | Category | Implication |
|--------------|----------|----------|-------------|
| VirtualAlloc | Y/N | Memory | Allocates executable memory |
| VirtualAllocEx | Y/N | Injection | Remote process memory allocation |
| VirtualProtect | Y/N | Memory | Changes page permissions (RX→RWX) |
| WriteProcessMemory | Y/N | Injection | Writes to remote process |
| CreateRemoteThread | Y/N | Injection | Creates thread in remote process |
| OpenProcess | Y/N | Injection | Opens handle to target process |
| LoadLibrary | Y/N | Evasion | Dynamic DLL loading |
| GetProcAddress | Y/N | Evasion | Dynamic API resolution |
| NtQuerySystemInformation | Y/N | Recon | System/process enumeration |
| RegSetValueEx | Y/N | Persistence | Registry modification |
| CreateService | Y/N | Persistence | Service creation |
| socket, connect, send | Y/N | Network | C2 communication |

**Behavioral Analysis:**

Based on imports, this executable likely:

```
Primary Purpose: [BENIGN / SUSPICIOUS / MALICIOUS]

Capabilities Detected:
□ Process Injection (VirtualAllocEx, WriteProcessMemory, CreateRemoteThread)
□ Dynamic API Resolution (GetProcAddress, LoadLibrary)
□ Memory Manipulation (VirtualAlloc, VirtualProtect)
□ Network Communication (ws2_32.dll functions)
□ Persistence Mechanisms (Registry, Services)
□ Anti-Analysis (IsDebuggerPresent, CheckRemoteDebuggerPresent)
□ System Enumeration (NtQuerySystemInformation)

Overall Assessment: ____________________________________
```




**My Results:**

![imports](../img/imports.png)


NOTE: To make this section a little more interesting I imported a HTTPS beacon I generated with Sliver.

### DLL 1: `KERNEL32.DLL` (40 entries)

|Function Name|Suspicion Level|Why?|
|---|---|---|
|`LoadLibraryA`|🔴 HIGH|Enables dynamic loading of any DLL at runtime - core evasion technique|
|`LoadLibraryW`|🔴 HIGH|Unicode version of LoadLibrary - loads DLLs dynamically|
|`GetProcAddress`|🔴 HIGH|Resolves function addresses at runtime - bypasses static import analysis|
|`VirtualAlloc`|🔴 HIGH|Allocates memory with executable permissions - shellcode staging|
|`VirtualFree`|🟡 MED|Memory cleanup - often paired with VirtualAlloc|
|`VirtualQuery`|🟡 MED|Queries memory region info - reconnaissance/validation|
|`CreateThread`|🔴 HIGH|Creates new execution thread - payload execution or injection setup|
|`ResumeThread`|🟡 MED|Resumes suspended thread - often used in process hollowing|
|`SuspendThread`|🟡 MED|Suspends thread execution - manipulation technique|
|`SetThreadContext`|🔴 HIGH|Modifies thread context - process injection indicator|
|`GetThreadContext`|🔴 HIGH|Retrieves thread context - used in advanced injection|
|`SwitchToThread`|🟢 LOW|Thread scheduling - can be legitimate|
|`GetSystemInfo`|🟡 MED|System reconnaissance - gathering environment info|
|`GetSystemDirectoryA`|🟡 MED|Locates system directory - often for DLL loading paths|
|`GetEnvironmentStringsW`|🟢 LOW|Accesses environment variables - common in Go runtime|
|`FreeEnvironmentStringsW`|🟢 LOW|Cleanup function for environment strings|
|`ExitProcess`|🟢 LOW|Normal process termination|
|`CreateFileA`|🟡 MED|File operations - could be legitimate or for dropping payloads|
|`WriteFile`|🟡 MED|Writes data to file - payload dropping or logging|
|`WriteConsoleW`|🟢 LOW|Console output - debugging or user interaction|
|`CloseHandle`|🟢 LOW|Resource cleanup - standard practice|
|`DuplicateHandle`|🟡 MED|Handle duplication - can be used for privilege manipulation|
|`CreateEventA`|🟢 LOW|Synchronization primitive - normal threading behavior|
|`SetEvent`|🟢 LOW|Event signaling - synchronization|
|`WaitForSingleObject`|🟢 LOW|Thread synchronization - common pattern|
|`WaitForMultipleObjects`|🟢 LOW|Multi-object synchronization|
|`CreateWaitableTimerExW`|🟢 LOW|Timer creation - scheduling tasks|
|`SetWaitableTimer`|🟢 LOW|Timer configuration|
|`CreateIoCompletionPort`|🟡 MED|Async I/O - often used in network communication frameworks|
|`GetQueuedCompletionStatusEx`|🟡 MED|Async I/O completion - network operations|
|`PostQueuedCompletionStatus`|🟡 MED|I/O completion posting - async networking|
|`GetStdHandle`|🟢 LOW|Standard I/O handles - console interaction|
|`GetConsoleMode`|🟢 LOW|Console properties - normal behavior|
|`SetConsoleCtrlHandler`|🟡 MED|Console event handling - can catch Ctrl+C for persistence|
|`SetErrorMode`|🟡 MED|Error handling configuration - can suppress error dialogs|
|`GetProcessAffinityMask`|🟡 MED|CPU affinity info - system reconnaissance|
|`SetProcessPriorityBoost`|🟢 LOW|Process scheduling - performance tuning|
|`SetUnhandledExceptionFilter`|🟡 MED|Exception handling - can be anti-debugging technique|
|`AddVectoredExceptionHandler`|🟡 MED|Exception handling - anti-debugging or error recovery|
|`TlsAlloc`|🟢 LOW|Thread-local storage - normal Go runtime behavior|


**Suspicious API Checklist:**

|API Function|Present?|Category|Implication|
|---|---|---|---|
|VirtualAlloc|**Y**|Memory|Allocates executable memory for shellcode/payloads|
|VirtualAllocEx|**N**|Injection|Remote process memory allocation (loaded dynamically)|
|VirtualProtect|**N**|Memory|Changes page permissions (loaded dynamically)|
|WriteProcessMemory|**N**|Injection|Writes to remote process (loaded dynamically)|
|CreateRemoteThread|**N**|Injection|Creates thread in remote process (loaded dynamically)|
|OpenProcess|**N**|Injection|Opens handle to target process (loaded dynamically)|
|LoadLibrary|**Y**|Evasion|**CRITICAL: Dynamic DLL loading capability**|
|GetProcAddress|**Y**|Evasion|**CRITICAL: Dynamic API resolution**|
|NtQuerySystemInformation|**N**|Recon|System/process enumeration (loaded dynamically)|
|RegSetValueEx|**N**|Persistence|Registry modification (loaded dynamically)|
|CreateService|**N**|Persistence|Service creation (loaded dynamically)|
|socket, connect, send|**N**|Network|C2 communication (ws2_32.dll loaded dynamically)|


**Behavioral Analysis:**

Based on imports, this executable likely:

```
Primary Purpose: [MALICIOUS - C2 Beacon]

Capabilities Detected:
☑ Dynamic API Resolution (LoadLibrary + GetProcAddress) ← PRIMARY RED FLAG
☑ Memory Manipulation (VirtualAlloc for shellcode staging)
☑ Thread Context Manipulation (SetThreadContext, GetThreadContext)
☑ Thread Control (CreateThread, SuspendThread, ResumeThread)
☑ Asynchronous I/O Framework (CreateIoCompletionPort, GQCS) - Network Comms
☑ Exception Handling (Anti-debugging potential)
☑ System Reconnaissance (GetSystemInfo, GetProcessAffinityMask)
□ Network Communication (ws2_32.dll loaded at RUNTIME - not in imports)
□ Process Injection APIs (VirtualAllocEx, WriteProcessMemory loaded at RUNTIME)
□ Persistence Mechanisms (Registry, Services loaded on-demand)

Overall Assessment: 
This is a SOPHISTICATED MALWARE employing ADVANCED IMPORT OBFUSCATION. 
The binary shows classic C2 beacon characteristics with deliberate 
evasion techniques. The combination of LoadLibrary + GetProcAddress 
with thread manipulation and memory allocation APIs indicates a 
payload that will dynamically resolve additional capabilities at runtime.

THREAT LEVEL: HIGH
```


**REMARKS: Why Only KERNEL32.DLL?**

This is NOT a limitation - it's a deliberate evasion technique.

**Reasons for Minimal Import Table:**

1. **Go Language Static Compilation**

    - Sliver is written in Go, which statically compiles most of its runtime
    - Go's runtime handles many OS interactions through minimal syscalls
    - Standard Go binaries naturally have smaller import tables than C/C++ equivalents
2. **Dynamic API Resolution Strategy**

    - The presence of `LoadLibraryA/W` + `GetProcAddress` is the **KEY INDICATOR**
    - These two functions act as "master keys" to access ANY Windows API at runtime
    - The malware can load any DLL (ws2_32.dll, advapi32.dll, ntdll.dll) dynamically
    - Functions are resolved by name at runtime, never appearing in the import table
3. **Direct/Indirect Syscalls**

    - Advanced malware often bypasses high-level APIs entirely
    - Uses direct syscalls to ntdll.dll functions (NtAllocateVirtualMemory, etc.)
    - Syscall numbers are resolved dynamically, leaving no import traces
4. **Evasion Benefits**

    - **Static Analysis Evasion**: Antivirus/EDR scanning imports will see minimal indicators
    - **Behavioral Signature Evasion**: Suspicious API combinations (VirtualAllocEx + WriteProcessMemory + CreateRemoteThread) aren't visible
    - **YARA Rule Evasion**: Many detection rules rely on import table patterns
    - **Analyst Confusion**: Makes static analysis incomplete/misleading

**What's Hidden:**

The beacon likely performs these actions at runtime (not visible in imports):

- Network communication via `ws2_32.dll` (socket, connect, send, recv)
- Process injection via `kernel32.dll` extended APIs (VirtualAllocEx, WriteProcessMemory, CreateRemoteThread)
- Persistence via `advapi32.dll` (RegSetValueEx, CreateService)
- Privilege escalation attempts
- Anti-analysis checks (debugger detection, VM detection)


**This PE demonstrates that a "clean" import table ≠ benign software.**

---




## Layer 7: Relocations - ASLR Support

### What We're Looking For
- Presence of base relocation table
- Number of relocation blocks
- Which sections need fixing

### Steps

1. Click **BaseReloc** tab in the middle panel
2. If present, PEBear shows relocation blocks

**Relocation Analysis:**

```
Base Relocation Directory Present: [YES / NO]

If YES:
  Number of Relocation Blocks: ________
  Total Relocations: ________
  
  Sections with relocations:
  □ .text (code section - expected)
  □ .data (data section - expected)
  □ .rdata (read-only data - expected)
  □ Other: ________

If NO:
  ⚠️ WARNING: Cannot rebase!
  - PE must load at exact ImageBase: 0x________
  - If address taken: Loading fails
  - ASLR ineffective
  - Process hollowing risk: [HIGH / MEDIUM / LOW]
```

**Offensive Implications:**

```
Can this PE be used in process hollowing? [YES / NO / MAYBE]

Reasoning:
- Has relocations: [YES / NO]
- If YES: Can load at any address ✓
- If NO: Must allocate at ImageBase 0x________ or fail ✗

Manual loading complexity: [LOW / MEDIUM / HIGH]
```


**My Results:**

![relocations](../img/relocations.png)


**Relocation Analysis:**

```
Base Relocation Directory Present: YES ✓

Total Relocations: 139+ entries (shown as "139 entries" in Relocation Block section)

Sections with relocations:
☑ .text (code section - expected) - Multiple blocks (B00000-B09000 ranges indicate .text)
☑ .data (data section - expected) - Blocks in AFE000-AFF000 range
☑ .rdata (read-only data - expected) - Present in various blocks
□ Other: N/A

Page RVAs with Relocations:
- AFE000 (192 entries, Block Size: 188)
- AFF000 (113 entries, Block Size: EA)
- B00000 (54 entries, Block Size: 74)
- B01000 (63 entries, Block Size: 86)
- B02000 (64 entries, Block Size: 88)
- B03000 (65 entries, Block Size: 8A)
- B04000 (138 entries, Block Size: 11C)
- B05000 (179 entries, Block Size: 16E)
- B06000 (189 entries, Block Size: 182)
- B07000 (189 entries, Block Size: 182)
- B08000 (195 entries, Block Size: 18E)
- B09000+ (additional blocks continue...)

Relocation Type: 64-bit field (IMAGE_REL_BASED_DIR64)
- Indicates 64-bit absolute addresses that need adjustment
- Standard for x64 PE files
```


**Security Analysis:**

```
ASLR Support: FULLY ENABLED ✓

This binary FULLY SUPPORTS Address Space Layout Randomization:
✓ Base relocation table present and populated
✓ Extensive relocations across all sections (139+ entries)
✓ Can load at any memory address
✓ Windows loader can randomize base address on each execution
✓ Significantly increases exploitation difficulty

Defense Benefits:
- Memory corruption exploits must bypass ASLR
- ROP chains cannot use hardcoded addresses
- Return addresses randomized per execution
- Exploitation requires information leaks
```


**Offensive Implications:**

```
Can this PE be used in process hollowing? YES ✓

Reasoning:
- Has relocations: YES ✓
- If YES: Can load at any address ✓ CONFIRMED
- Relocation table is comprehensive (139+ entries)
- All necessary sections have relocation entries

Manual loading complexity: MEDIUM

Why MEDIUM complexity:
✓ PRO: Relocation table present - can rebase to any address
✓ PRO: Standard PE format - well-documented relocation process
✗ CON: Manual relocation required when hollowing (must walk relocation blocks)
✗ CON: Must calculate delta between original ImageBase and injection target
✗ CON: Must apply fixups to all 139+ relocation entries
✗ CON: 64-bit relocations require proper pointer arithmetic

Process Hollowing Workflow:
1. Create suspended process (target)
2. Unmap target's image
3. Allocate memory at preferred address (or any available address)
4. Copy beacon sections to target process
5. **APPLY RELOCATIONS**: Walk relocation table, calculate delta, fix addresses
6. Update PEB ImageBase pointer
7. Set thread context (EntryPoint)
8. Resume thread

The presence of relocations makes this PE SUITABLE for injection techniques
but requires the attacker to manually process the relocation table during
the injection process.
```


**Additional Technical Details:**

```
Relocation Format: IMAGE_BASE_RELOCATION structure
- Each block covers one 4KB page (Page RVA)
- Block Size includes header (8 bytes) + relocation entries
- Each entry: 2 bytes (4-bit type + 12-bit offset)
- Type 10 (0xA) = IMAGE_REL_BASED_DIR64 (64-bit absolute address)

Example Block Analysis:
Page RVA: B00000 (RVA in .text section)
Block Size: 74 (116 bytes)
Entries: 36 relocations
Formula: (116 - 8) / 2 = 54 relocations ✓ (matches "36" in hex = 54 decimal)

This indicates the .text section has hardcoded addresses that need
adjustment when the PE loads at a different base address.
```



**Comparison Note:**

```
🔍 INTERESTING OBSERVATION:

Many shellcode loaders and malware samples deliberately STRIP relocations
to make analysis harder and avoid certain detection signatures.

This Sliver beacon KEEPS relocations, which indicates:
✓ Professional development (maintains PE compatibility)
✓ Flexibility for various injection scenarios
✓ Proper ASLR support for defense evasion
✓ Can be used as standalone executable OR injected payload

Some malware families remove .reloc section entirely to:
- Reduce file size
- Complicate memory forensics
- Force loading at fixed address (predictable for attacker)

Sliver's approach suggests a mature C2 framework designed for
operational flexibility rather than pure size optimization.
```


**Final Assessment:**

```
VERDICT: This PE is RELOCATION-AWARE and ASLR-COMPATIBLE

✓ Supports dynamic base address loading
✓ Compatible with modern Windows security features
✓ Suitable for process injection/hollowing (with manual relocation)
✓ Professional PE structure maintained
✓ Harder to exploit due to ASLR, but also harder to detect static addresses

This is characteristic of SOPHISTICATED MALWARE that maintains
compatibility with legitimate PE loading mechanisms while supporting
advanced injection techniques.
```



---

## Next Steps

You've now manually located every critical field in a PE file. In new next lab we'll automate this process by building our own PE parser that extracts all these values programmatically.

**What you've learned:**
- ✓ How to systematically analyze any PE file
- ✓ Which values matter for offensive operations
- ✓ How to identify security weaknesses
- ✓ How to assess behavioral intent from imports
- ✓ The relationship between PE structure and loading process

Save your completed analysis - you'll compare these values to your parser's output in the next lab!

___






---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./pe.md" >}})
[|NEXT|]({{< ref "./peC.md" >}})