---
showTableOfContents: true
title: "Part 5A - The PE File Format: Understanding Address Translation + Lab"
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


**You can see here in my case the values are:**
- Import Directory RVA: `0x0028C000`
- Import Directory Size: `0x53E` (1342 bytes)


Why do we care about this again? This is where the Import Directory exists **in memory** when the PE is loaded.

Note please your values are going to be different, write them down and USE YOUR VALUES in the steps below, NOT MINE.


### Try to Find It Using RVA as File Offset

1. Open `simple.exe` in **HxD** (or your hex editor)
2. Go to address `0x0028C000` (Ctrl+G → enter your RVA)

![lab 1 hex rva fail](../img/lab1_02.png)

As you can see, we get the error that the offset does not even exist... Which is of course the whole point of this exercise.

We were expecting this to fail. The error "invalid digit" is because `0x0028C000` is **2,670,592 bytes** in decimal. Let's look at the actual size of our binary:

![simple.exe size](../img/lab1_03.png)



We can see our file is **2,473,984 bytes** (about 2.35 MB). And since were trying to jump to `0x0028C000` which is **2,670,592 bytes** - that's **196,608 bytes PAST the end of the file**!

This is why HxD gave us the error - you literally asked it to go to a position that doesn't exist in the file.

Note that in the case that the file was larger than our offset we would not get this error, we would actually expect to be taken to a position in the file. But we'll encounter random-looking bytes, definitely not an import directory structure.

So to review - The RVA `0x0028C000` only makes sense **in memory** when Windows loads the executable with proper 4KB alignment. On disk, that same data is stored at a much earlier position (probably around offset `0x0028A000` or so, but we'll calculate the exact position shortly).


### Find the Correct File Offset

Now let's do it properly using section headers.

**In PEBear:**

1. Click on the **Section Hdrs** tab.
2. You should see the following.


![section headers size](../img/lab1_04.png)



We can see here for all the sections, we have both the virtual address, as well as virtual size which we can use to determine the range for the section.

As an example, for `.text` we have a VA of `0x1000` and a size of `0xAAAF1`, this means the range for `.text `is `0x1000` to `0xABAF1`.

Now we ask ourselves -  **which section contains your Import Directory RVA** of `0x28C000`.

Well, since `0x0028C000` > `0xABAF1` we know it's not `.text`. So now we go through each section, determine it's range in a similar fashion to determine where the Import  Directory RVA is.


In this case it's actually pretty easy since when we get down to `.idata` we can see that its Virtual Address `0x28C000` **exactly matches** our Import Directory RVA. This makes perfect sense because `.idata` (import data) is specifically the section that contains "import data".




### Now Calculate the File Offset

Using the formula from our instructions:

```
RVA of Import Directory:     0x0028C000
Section Virtual Address:     0x0028C000  (.idata)
Offset within section:       0x0028C000 - 0x0028C000 = 0x00000000

Section Raw Address (file):  0x0023C800  (.idata Raw Addr)
Offset within section:       0x00000000
File Offset:                 0x0023C800 + 0x00000000 = 0x0023C800
```


**So Our Import Directory File Offset is: `0x0023C800`**

### Verify in Hex Editor

Now we can proceed to verify this in our hex editr.

1. Open **HxD**
2. Press **Ctrl+G** and go to address `0x23C800` (in `HxD` you need to drop the `0x` - just use `23C800`)
3. This time it should work since this offset is well within your file size of 2,473,984 bytes

You should see the following:

![hxd results](../img/lab1_05.png)

### Verify with PEBear's Import View

1. In **PEBear**, click **Imports** in the left panel
2. You should see the list of imported DLLs (kernel32.dll, etc.)
3. PEBear automatically did the RVA→File Offset conversion for you

![pebear imports](../img/lab1_06.png)


**Key insight:** PEBear shows you the imports, but it had to perform the exact same conversion we just did manually to find them in the file.




## Analyzing the PE in Memory (Dynamic Analysis)

Now let's see how addresses work when the PE is actually loaded and running.

### Load in x64dbg

1. Launch **x64dbg**
2. **File → Open** and select `simple.exe`
3. The debugger will break at the entry point
4. **Do not press F9 (run) yet** - we want to examine memory first

### Find the ImageBase

1. On the top, click on the **Memory Map** tab
2. In the Info column, find `simple.exe`
3. Note the address here - this is our `ImageBase`
4. In my case it is - `0x00007FF716BB0000`

**NOTE:** This address is randomized by ASLR. This means not only will mine be different from yours, but yours will change every time you run it!



### Calculate the Import Directory Virtual Address

Before we calculate our Import Directory RVA: `0x0028C000`

So now let's calculate the Virtual Address (VA) in memory:
```
VA = ImageBase + RVA
VA = 0x00007FF716BB0000 + 0x0028C000
VA = 0x00007FF716E3C000
```

So in this case my Import Directory VA in memory: `0x00007FF716E3C000`

Let's confirm this back in x64dbg.

### Examine Memory at the Virtual Address

Now let's verify this is correct:

1. In `x64dbg`, click on `Dump 1` tab, then hit press **Ctrl+G** (Go to expression)
2. Enter: `0x00007FF716E3C000` (or just `7FF716E3C000`)
3. Press Enter

This should take you to the memory location where the Import Directory is loaded.

![x64dbg results](../img/lab1_07.png)



- Now look what we see in memory at `0x00007FF716E3C000`
- We once again see the module name (`kernel32`) and our functions names for example `WriteConsole`
- In other words, exactly what we saw in **HxD**.


### The Key Insight - Same Data, Different Addresses

| Location                | Tool   | Address              | Data                                                  |
| ----------------------- | ------ | -------------------- | ----------------------------------------------------- |
| **On Disk (File)**      | HxD    | `0x0023C800`         | Import directory with kernel32.dll and function names |
| **In Memory (Running)** | x64dbg | `0x00007FF716E3C000` | **Same exact data!**                                  |


**They're different because:**

1. **On disk**: Data stored sequentially in the .exe file (file offset)
2. **In memory**: Windows loader maps sections to virtual addresses (RVA + ImageBase)
3. **Section alignment**: 512 bytes on disk vs 4KB pages in memory



## Key Takeaways

### What We Just Learned

1. **RVAs in headers are NOT file offsets** - using them directly in a hex editor gives garbage
2. **You must convert:** RVA → Section → Calculate offset → File offset
3. **In memory, RVAs work differently:** ImageBase + RVA = Virtual Address
4. **Tools like PEBear hide this complexity** - but now you know what they're doing behind the scenes
5. **When manually loading PEs** (process hollowing, reflective DLL injection), you must handle both:
   - Reading from file offsets (disk)
   - Writing to virtual addresses (memory)

### Common Mistakes (That You Now Avoid)

❌ Using an RVA as a file offset in a hex editor
❌ Forgetting to add ImageBase when calculating memory addresses
❌ Ignoring section alignment differences between disk and memory
❌ Assuming all sections start at the same relative positions on disk and in memory

### Why This Matters for Offensive Security

When you implement:
- **Process Hollowing:** You copy sections from disk offsets to memory VAs
- **Reflective DLL Injection:** You must apply relocations using RVAs
- **Import Resolution:** You read import names from file offsets, write addresses to IAT using VAs
- **Manual PE Loading:** Every step requires converting between these address types







---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./securityD.md" >}})
[|NEXT|]({{< ref "./peB.md" >}})