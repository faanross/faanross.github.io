---
showTableOfContents: true
title: "Part 5D - The PE File Format: Process Memory Mapping Lab"
type: "page"
---






## Lab: PE Parsing and Process Memory Mapping

### Learning Objectives
By completing this lab, you will:
- Understand the Windows PE (Portable Executable) file format structure
- Learn how to enumerate loaded modules (DLLs) in a process
- Parse PE headers to extract section information (.text, .data, .rdata, etc.)
- Map the complete memory layout of a Windows process
- Identify different memory region types (Image, Private, Mapped)
- Understand memory protection flags and their security implications
- Gain practical experience with Windows memory forensics techniques

### What You'll Learn
- How PE files are structured in memory
- The difference between DOS headers, NT headers, and section tables
- How to safely parse PE structures without crashing
- Memory region identification and classification
- The relationship between virtual memory and loaded modules


## Part 1: The Failure - Basic Memory Enumeration Without Context





In this part, you'll create a basic memory scanner that simply lists memory regions without any context. You'll see raw addresses and flags, but won't understand what they represent.

### Create Basic Memory Scanner

Create `memscan_v1.go`, source code is [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v1/memscan_v1.go).

### Compile and Run

```powershell
# compile on Windows with
go build -o memscan_v1.exe memscan_v1.go

# compile on Linux/MacOS with
GOOS=windows GOARCH=amd64 go build -o memscan_v1.exe memscan_v1.go
```


Once the target system simply run:

```powershell
.\memscan_v1.exe
```

### Results


You can see the complete results [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v1/results.md).

Here is a truncated overview for discussion purposes:

```powershell
...
0x0000000000050000 - 0x0000000000054000  R--  Mapped
0x0000000000060000 - 0x0000000000062000  RW-  Private
...
0x00007FFD5E841000 - 0x00007FFD5E9ED000  R-X  Image
0x00007FFD5E9ED000 - 0x00007FFD5EBDC000  R--  Image
0x00007FFD5EBDC000 - 0x00007FFD5EBE4000  RW-  Image
0x00007FFD5EBE4000 - 0x00007FFD5EBE6000  ???  Image
...
```


**What we're seeing:**
- Raw memory addresses with protection flags (`R-X`, `RW`-, etc.)
- Generic types (`Image`, `Private`, `Mapped`)
- No identification of which DLL or section each region belongs to

**Why this is insufficient:**
1. **No attribution**: Can't tell if `R-X` regions are from kernel32.dll, ntdll.dll, or from our executable
2. **No section info**: Can't distinguish between `.text` (code) and `.rdata` (read-only data)
3. **Limited value**: We need to know WHAT code is WHERE

**The fundamental problem:** Windows loads modules (EXEs and DLLs) as PE files, but our scanner doesn't understand PE structure, so it sees memory regions without understanding their purpose.


## Part 2: Understanding PE Structure - Adding Module Enumeration


Now we'll enhance the scanner to enumerate loaded modules and identify which memory regions belong to which DLLs.

### Reviewing PE File Format

Before we code, let's review what we're parsing:

```
PE File Structure in Memory:
┌─────────────────────────────────────────┐
│ DOS Header (IMAGE_DOS_HEADER)           │  ← Always starts with "MZ"
│  - e_magic: 0x5A4D ("MZ")               │
│  - e_lfanew: offset to PE header        │
├─────────────────────────────────────────┤
│ DOS Stub (old DOS program)              │
├─────────────────────────────────────────┤
│ PE Signature: "PE\0\0" (0x00004550)     │
├─────────────────────────────────────────┤
│ IMAGE_FILE_HEADER                       │
│  - NumberOfSections                     │
│  - SizeOfOptionalHeader                 │
├─────────────────────────────────────────┤
│ IMAGE_OPTIONAL_HEADER64                 │
│  - ImageBase, SizeOfImage               │
│  - AddressOfEntryPoint                  │
├─────────────────────────────────────────┤
│ Section Table (array of sections)       │
│  [0] .text  (executable code)           │
│  [1] .rdata (read-only data)            │
│  [2] .data  (initialized data)          │
│  [3] .pdata (exception info)            │
│  ...                                    │
└─────────────────────────────────────────┘
```

### Create Enhanced Scanner with Module Enumeration

Create `memscan_v2.go`, source code is [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v2/memscan_v2.go).


### Compile and Run

```powershell
# compile on Windows with
go build -o memscan_v2.exe memscan_v2.go

# compile on Linux/MacOS with
GOOS=windows GOARCH=amd64 go build -o memscan_v2.exe memscan_v2.go
```


Once the target system simply run:

```powershell
.\memscan_v2.exe
```

### Resullts

You can see the complete results [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v2/results.md).

Here is a truncated overview for discussion purposes:

```powershell
╔════════════════════════════════════════════════════════════╗
║  ENHANCED MEMORY SCANNER v2.0 - With Module Enumeration    ║
╚════════════════════════════════════════════════════════════╝

[*] Step 1: Enumerating loaded modules...
[✓] Found 11 loaded modules

    [  1] memscan_v2.exe                  Base: 0x00007FF66FEA0000  Size: 0x00293000
    [  2] ntdll.dll                       Base: 0x00007FFD611C0000  Size: 0x00269000
...
    [ 10] UMPDC.dll                       Base: 0x00007FFD5E1B0000  Size: 0x00014000
    [ 11] psapi.dll                       Base: 0x00007FFD606C0000  Size: 0x00008000

═════════════════════════════════════════════════════════════════
[*] Step 2: Scanning memory with module attribution...
Start Address       - End Address         Prot  Type     Module
────────────────────────────────────────────────────────────────────────────────
0x0000000000010000 - 0x0000000000011000  RW-  Mapped   Memory-Mapped File
0x0000000000020000 - 0x0000000000030000  RW-  Mapped   Memory-Mapped File
...
0x000000000025B000 - 0x0000000000268000  RW-  Private  Heap/Stack/Private
0x00000000005FA000 - 0x00000000005FD000  RW-  Private  Heap/Stack/Private
...
0x00007FF66FEA1000 - 0x00007FF66FF42000  R-X  Image    memscan_v2.exe
0x00007FF66FF42000 - 0x00007FF670014000  R--  Image    memscan_v2.exe
...
0x00007FFD5EBE6000 - 0x00007FFD5EC33000  R--  Image    KERNELBASE.dll
0x00007FFD5EC33000 - 0x00007FFD5EC34000  R-X  Image    Unknown Image
0x00007FFD5ECD0000 - 0x00007FFD5ECD1000  R--  Image    ucrtbase.dll
...
[✓] Scan complete with module attribution

[⚠️] PARTIAL SUCCESS: We know which DLL, but not which section!
    - Can identify: memory belongs to 'kernel32.dll'
    - Cannot identify: whether it's .text, .data, or .rdata
    - Next step: Parse PE sections for complete forensics
```


**What improved:**
- Now we can see which DLL each memory region belongs to
- We identified that multiple consecutive regions can be part of the same module
- We distinguished between application memory (i.e. the exe) and system libraries

**What's still missing:**
- We see three different regions in `memscan_v2.exe` with different protections (`R-X`, `R--`, `RW-`)
- These are likely `.text`, `.rdata`, and `.data` sections, but we can't confirm
- We need this granular detail to locate specific code or data

**Why sections matter:**
- `.text`: Contains executable code (we often want to inject code here)
- `.data`: Contains writable data (potential for data exfiltration)
- `.rdata`: Contains read-only data like import tables (important for anti-analysis efforts)

---


## Part 3: Complete Solution - PE Section Parsing

Now we'll implement full PE parsing to extract section names and ranges.

### Create Complete Memory Forensics Tool

Create `memscan_v3.go`, source code is [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v3/memscan_v3.go).





### Compile and Run


```powershell
# compile on Windows with
go build -o memscan_v3.exe memscan_v3.go

# compile on Linux/MacOS with
GOOS=windows GOARCH=amd64 go build -o memscan_v3.exe memscan_v3.go
```


Once the target system simply run:

```powershell
.\memscan_v3.exe
```


### Results

You can see the complete results [HERE](https://github.com/faanross/complete_go_course/blob/main/lesson_01_3/part05/lab_memscan/memscan_v3/results.md), please refer there.


### Complete Success Indicators

#### **PE Section Parsing is Working** 

Look at Phase 1 - it's now parsing ALL sections from each DLL:

```powershell
Module [  1]: memscan_v3.exe                 @ 0x00007FF6178C0000
    [✓] memscan_v3.exe: PE validated, parsing 16 sections
        [1] .text     0x00007FF6178C1000 - 0x00007FF617962EB1  (size: 0xA1EB1)
        [2] .rdata    0x00007FF617963000 - 0x00007FF617A34FA8  (size: 0xD1FA8)
        [3] .data     0x00007FF617A35000 - 0x00007FF617A8BB28  (size: 0x56B28)
        [4] .pdata    0x00007FF617A8C000 - 0x00007FF617A90AA0  (size: 0x4AA0)
```



#### **Section Attribution in Memory Scan** 

Now look at Phase 2 - the memory regions are showing **which section** they belong to:

```
0x00007FF6178C0000 - 0x00007FF6178C1000  R--  Image    memscan_v3.exe (PE Headers)
0x00007FF6178C1000 - 0x00007FF617963000  R-X  Image    memscan_v3.exe (.text)      ← CODE!
0x00007FF617963000 - 0x00007FF617A35000  R--  Image    memscan_v3.exe (.rdata)     ← READ-ONLY DATA!
0x00007FF617A35000 - 0x00007FF617A37000  RW-  Image    memscan_v3.exe (.data)      ← WRITABLE DATA!
0x00007FF617A8C000 - 0x00007FF617B34000  R--  Image    memscan_v3.exe (.pdata)     ← EXCEPTION INFO!
```


####  Attribution for ALL DLLs

Same level of detail for system DLLs:
```
0x00007FFD60471000 - 0x00007FFD604F7000  R-X  Image    KERNEL32.DLL (.text)
0x00007FFD604F7000 - 0x00007FFD6052F000  R--  Image    KERNEL32.DLL (.rdata)
0x00007FFD6052F000 - 0x00007FFD60531000  RW-  Image    KERNEL32.DLL (.data)
0x00007FFD60531000 - 0x00007FFD60539000  R--  Image    KERNEL32.DLL (.pdata)
```



#### Interesting Observations in Our Output

##### Go Compiler Peculiarities:

`memscan_v3.exe` has unusual section names:

```
[6] /4        0x00007FF617A92000 - 0x00007FF617A92154
[7] /19       0x00007FF617A93000 - 0x00007FF617AB9B1F
[8] /32       0x00007FF617ABA000 - 0x00007FF617AC15D2
```

These `/4`, `/19`, `/32` sections are **Go runtime-specific sections** - not standard PE sections! This is actually really cool to see.

##### Memory Protection Oddities:

```
0x00007FF617A37000 - 0x00007FF617A3A000  ???  Image    memscan_v3.exe (.data)
```

Some regions show `???` protection - these have non-standard protection flags. This is normal for Go binaries which use custom memory layouts.

##### PEB Detection Working:

```
0x000000007FFE0000 - 0x000000007FFE1000  R--  Private  PEB (Process Environment Block)
0x000000007FFEE000 - 0x000000007FFEF000  R--  Private  PEB (Process Environment Block)
```

Our special address detection is working.

##### Heap vs Stack Classification:

```
0x0000000000060000 - 0x0000000000062000  RW-  Private  Stack / TLS (Thread-Local)
0x00000000001A0000 - 0x00000000001E0000  RW-  Private  Heap (Dynamic Allocation)
```

The heuristics for identifying heap vs stack are working.

##### **Statistics Look Good** 

```
Statistics:
  • Image regions (DLLs/EXE):     73   ← All DLL sections
  • Private regions (Heap/Stack): 35   ← Memory allocations
  • Mapped regions (Files):       17   ← Memory-mapped files
  • Total committed regions:      125
```





## Summary


### What We've Achieved

#### **Compare the three versions:**

|Feature|v1|v2|v3|
|---|---|---|---|
|Shows memory addresses|✓|✓|✓|
|Shows protection flags|✓|✓|✓|
|Identifies which DLL|✗|✓|✓|
|**Identifies which section**|✗|✗|**✓**|
|**Parses PE headers**|✗|✗|**✓**|
|**Full forensic attribution**|✗|✗|**✓**|

### Key Concepts Learned

✓ **PE File Structure**

- DOS Header → PE Signature → File Header → Optional Header → Sections
- Every Windows executable follows this format
- Sections define code (`.text`), data (`.data`), and constants (`.rdata`)

✓ **Memory Region Types**

- **MEM_IMAGE**: Loaded PE files (EXEs, DLLs)
- **MEM_PRIVATE**: Heap, stack, thread-local storage
- **MEM_MAPPED**: Memory-mapped files

✓ **Protection Flags**

- **R-X**: Read-Execute (code sections)
- **R--**: Read-only (constants, import tables)
- **RW-**: Read-Write (variables, heap, stack)
- **RWX**: Read-Write-Execute (SUSPICIOUS!)

✓ **Attribution**

- Module enumeration identifies loaded DLLs
- PE parsing extracts section boundaries
- Memory scanning maps virtual address space
- Combined: complete memory layout



**Congratulations!** You've built a complete Windows memory forensics tool from scratch. You now understand how PE files are structured, how to safely parse them, and how to map an entire process's memory space. These skills are fundamental for malware analysis, exploit development, and security research.






---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./peC.md" >}})
[|NEXT|]({{< ref "./peb_teb.md" >}})