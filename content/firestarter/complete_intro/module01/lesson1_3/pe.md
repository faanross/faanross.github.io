---
showTableOfContents: true
title: "Part 5A - The PE File Format: Intro + Understanding Address Translation"
type: "page"
---

## The Portable Executable (PE) Format: An Offensive Security Primer

Every Windows executable - whether it's a `.exe`, `.dll`, or `.sys` file - uses the Portable Executable (PE) format. If you want to inject code, hide malicious behaviour, bypass security controls, or understand how Windows loads and executes programs, you must understand PE structure.

This isn't academic. When you perform process hollowing, you're manipulating PE headers. When you implement reflective DLL injection, you're reimplementing the Windows PE loader. When you hide API calls from static analysis, you're exploiting the Import Address Table structure. The PE format is the foundation beneath every offensive technique on Windows.



## The Big Picture: A Layered Architecture

Think of a PE file as an onion with distinct layers, each serving a specific purpose in the executable's lifecycle:

```
┌──────────────────────────────────────────────────────────────┐
│                    PE FILE STRUCTURE                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  DOS HEADER (IMAGE_DOS_HEADER)                   [Legacy]    │
│  ┌────────────────────────────────────────┐                  │
│  │ e_magic: "MZ" (0x5A4D)                 │  ← Signature     │
│  │ e_lfanew: Offset to PE header          │  ← Bridge        │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  DOS STUB                                        [Legacy]    │
│  ┌────────────────────────────────────────┐                  │
│  │ "This program cannot be run in DOS     │                  │
│  │  mode." + 16-bit stub code             │                  │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  PE SIGNATURE ("PE\0\0" = 0x00004550)           [Validation] │
│                                                              │
│  FILE HEADER (IMAGE_FILE_HEADER)                [Metadata]   │
│  ┌────────────────────────────────────────┐                  │
│  │ Machine: 0x8664 (x64) / 0x14C (x86)    │  ← Architecture  │
│  │ NumberOfSections: Count of sections    │                  │
│  │ TimeDateStamp: Compile time            │  ← Forensics     │
│  │ Characteristics: Flags                 │  ← File type     │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  OPTIONAL HEADER (IMAGE_OPTIONAL_HEADER)        [Critical]   │
│  ┌────────────────────────────────────────┐                  │
│  │ Magic: 0x10B (PE32) / 0x20B (PE32+)    │  ← Format        │
│  │ AddressOfEntryPoint: Start address     │  ← Execution     │
│  │ ImageBase: Preferred load address      │  ← ASLR base     │
│  │ SectionAlignment: Memory alignment     │  ← 4KB typical   │
│  │ FileAlignment: Disk alignment          │  ← 512B typical  │
│  │ SizeOfImage: Total image size          │  ← VirtualAlloc  │
│  │ DataDirectory[16]: Important tables    │  ← See below     │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  DATA DIRECTORIES (16 entries)                  [Pointers]   │
│  ┌────────────────────────────────────────┐                  │
│  │ [0] Export Directory                   │  ← DLL exports   │
│  │ [1] Import Directory (IAT)             │  ← API imports ★ │
│  │ [2] Resource Directory                 │  ← Icons, etc    │
│  │ [3] Exception Directory                │  ← SEH data      │
│  │ [5] Relocation Directory               │  ← ASLR fixes ★  │
│  │ [6] Debug Directory                    │  ← PDB info      │
│  │ [9] TLS Directory                      │  ← Thread data   │
│  │ ... others ...                         │                  │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  SECTION TABLE (Array of headers)               [Descriptors]│
│  ┌────────────────────────────────────────┐                  │
│  │ .text  → Code section                  │  ← RX perms      │
│  │ .data  → Initialized global data       │  ← RW perms      │
│  │ .rdata → Read-only data + imports      │  ← R only        │
│  │ .rsrc  → Resources (icons, strings)    │  ← R only        │
│  │ .reloc → Relocation data               │  ← R only        │
│  │ ... custom sections possible ...       │                  │
│  └────────────────────────────────────────┘                  │
│                                                              │
│  SECTION DATA (Actual content)                  [Payloads]   │
│  ┌────────────────────────────────────────┐                  │
│  │ Machine code, variables, imports,      │                  │
│  │ resources, and other section contents  │                  │
│  └────────────────────────────────────────┘                  │
│                                                              │
└──────────────────────────────────────────────────────────────┘

  Key for Offensive Security:
  ★ = Critical for injection techniques
  [Legacy] = Historical compatibility
  [Validation] = Loader checks
  [Metadata] = File properties  
  [Critical] = Essential for loading
  [Pointers] = Directory of important data
  [Descriptors] = Section metadata
  [Payloads] = Actual content
```

## Understanding Address Translation: RVA vs File Offset

This is **the most fundamental concept** you need to internalize before analyzing or manipulating PE files. Get this wrong, and everything breaks.

### The Core Problem

**Addresses in PE headers are not file positions**. When you open a PE file in a hex editor, you're looking at disk offsets. But every address referenced in PE headers - entry points, import tables, section locations - uses **Relative Virtual Addresses (RVAs)**: offsets from where the image will load in memory, not where data sits in the file.

| Address Type | Context | Example | Notes |
|--------------|---------|---------|-------|
| **File Offset** | On disk | Byte 0x1200 in the .exe file | Direct position in file |
| **RVA (Relative Virtual Address)** | In memory | 0x1000 from image base | Relative to where loaded |
| **VA (Virtual Address)** | In memory | ImageBase + RVA (e.g., 0x140001000) | Absolute memory address |

### Why This Will Trip You Up Constantly

**Scenario:** You want to find the Import Address Table to see what APIs an executable uses.

1. You read the Optional Header and find `DataDirectory[1].VirtualAddress = 0x6000`
2. You jump to byte 0x6000 in your hex editor
3. **You see garbage** - wrong data entirely
4. **Why?** That 0x6000 is an RVA (memory address), not a file offset

**What you actually need to do:**
1. Read the Import Directory RVA: `0x6000`
2. Look up which section contains RVA `0x6000` (iterate through section headers)
3. Find that `.rdata` section has `VirtualAddress = 0x6000`, `PointerToRawData = 0x4800`
4. Calculate: File Offset = `0x4800 + (0x6000 - 0x6000)` = `0x4800`
5. Jump to byte `0x4800` in your hex editor
6. **Now you see the Import Directory**

This conversion - from RVA to file offset - is something you'll do dozens of times when analyzing a single PE. Every interesting structure (imports, exports, relocations, resources) is referenced by RVA in the headers.


### The Alignment Mismatch: Why RVA ≠ File Offset

Sections align differently on disk versus memory, which is why you can't just use RVAs as file offsets:

```
DISK ALIGNMENT (FileAlignment = 0x200 = 512 bytes):
┌─────────────────────────────────────────────┐
│ Headers        │ 0x000 - 0x1FF (512 bytes)  │
│ .text section  │ 0x200 - 0x7FF (1536 bytes) │ ← Padded to 0x800
│ .data section  │ 0x800 - 0x9FF (512 bytes)  │
│ .rdata section │ 0xA00 - 0xBFF (512 bytes)  │
└─────────────────────────────────────────────┘

MEMORY ALIGNMENT (SectionAlignment = 0x1000 = 4KB):
┌─────────────────────────────────────────────┐
│ Headers        │ RVA 0x0000 - 0x0FFF        │
│ .text section  │ RVA 0x1000 - 0x1FFF        │ ← Padded to 0x2000
│ .data section  │ RVA 0x2000 - 0x2FFF        │
│ .rdata section │ RVA 0x3000 - 0x3FFF        │
└─────────────────────────────────────────────┘
```

Notice: `.rdata` section starts at file offset `0xA00` but at memory RVA `0x3000`. If a header says "Import Directory is at RVA 0x3010", you need to:
1. Find the section containing RVA `0x3010` (`.rdata`: RVA `0x3000-0x3FFF`)
2. Calculate offset within section: `0x3010 - 0x3000 = 0x10`
3. Add to section's file position: `0xA00 + 0x10 = 0xA10`
4. Read from file offset `0xA10`

**Why different alignments?**
- **Disk:** 512-byte alignment minimizes file size (historical sector size)
- **Memory:** 4KB alignment matches page size for efficient memory management and protection

### The Conversion Algorithm You'll Use Constantly

When you need to find data referenced by an RVA:

```
INPUT: RVA you want to locate (e.g., 0x3010)

STEP 1: Iterate through section headers
  For each section:
    Check if: RVA >= section.VirtualAddress 
          AND RVA < (section.VirtualAddress + section.VirtualSize)
    If true: RVA is in this section, continue to step 2

STEP 2: Calculate offset within the section
  offset_in_section = RVA - section.VirtualAddress

STEP 3: Calculate file offset
  file_offset = section.PointerToRawData + offset_in_section

STEP 4: Read data from file at file_offset
```

**Real example:** You want to read the entry point code.
- Optional Header says: `AddressOfEntryPoint = 0x1420` (this is an RVA)
- Find section: `.text` has `VirtualAddress = 0x1000`, `PointerToRawData = 0x400`
- Calculate: `0x1420 - 0x1000 = 0x420` (offset in section)
- File offset: `0x400 + 0x420 = 0x820`
- Jump to byte `0x820` in the file to see the entry point instructions


### When You're Manually Loading a PE

When you're doing for example process hollowing or reflective DLL injection, you're copying sections from disk into memory. You must use **both** address types:

- **File offsets** to read section data from disk: `section.PointerToRawData`
- **RVAs** to know where to write in memory: `ImageBase + section.VirtualAddress`
- **Size adjustments**: Use `section.SizeOfRawData` for disk, `section.VirtualSize` for memory (they differ!)

Mess up this conversion, and your injected code crashes immediately because instructions reference wrong addresses.

### The Bottom Line

Every time you see an address in a PE header, ask yourself: "Is this an RVA or a file offset?"
- If you're reading the file from disk: Convert RVA → file offset
- If you're loading into memory: Use RVA directly (after adding ImageBase for absolute VA)
- Section headers are your Rosetta Stone - they contain both RVAs and file offsets, letting you translate between them

Master this conversion, and PE analysis becomes straightforward. Skip it, and you'll spend hours wondering why nothing works.

If this is your first time encountering this, perhaps it seems a bit confusing, it certainly tripped me up the first time I encountered it. So let's do a quick lab where you can encounter this first hand and do the calculations yourself, I promise that by the end you'll see just how elementary this actually is. And again, I cannot stress this enough - if you can't do this correctly, everything else will break. It's tablestakes.



## Lab: Understanding RVA to File Offset Conversion

## Objective

See firsthand why RVA ≠ File Offset by examining a real PE file both on disk and in memory. You'll locate the Import Address Table using both methods and prove they require different addresses.

## Lab Setup

### Prerequisites

We'll first build a simple Go application, thereafter we'll analyze it using:
1. **PEBear** - PE analysis tool ([download](https://github.com/hasherezade/pe-bear/releases))
2. **x64dbg** - Debugger ([download](https://x64dbg.com/))
3. **HxD** or any hex editor ([download](https://mh-nexus.de/en/hxd/))


### Build the Test Executable

Let's create a simple Go program that we can analyze. Save this as `simple.go`:

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Simple PE Analysis Test Program")
    fmt.Println("================================")
    fmt.Println("This program imports several Windows APIs")
    fmt.Println("Press Enter to exit...")
    
    // Wait for input so the process stays alive for debugging
    fmt.Scanln()
}
```

**Build the executable:**
```bash
# compiling on Windows
go build -o simple.exe simple.go

# compiling on Linux/Darwin
GOOS=windows GOARCH=amd64 go build -o simple.exe simple.go
```


This creates a standard Windows PE executable with imports from kernel32.dll and other system DLLs.

If you did not build the application on your target machine, transfer it over using your preferred method.

## Analyzing the PE on Disk (Static Analysis)

### Open in PEBear

1. Launch **PEBear**
2. **File → Load PEBear** and select `simple.exe`
3. You should see the PE structure in the left panel

### Locate the Import Directory RVA

1. In the left panel, click **Optional Header**
2. Scroll down to **Data Directories**
3. Find **Import Directory**
4. Note two values:
    - Import Directory Address (this is the RVA)
    - Import Directory Size

![lab 1 data directory](../img/lab1_01.png)





---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./securityD.md" >}})
[|NEXT|]({{< ref "./peB.md" >}})