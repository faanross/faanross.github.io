---
showTableOfContents: true
title: "Manual DLL Mapping in Go (Lab 3.1)"
type: "page"
---

## Goal
Building upon our PE parser from Lab 2.2, we'll now build out the core mapping process of reflective loading.

Specifically, we'll give our application the ability to:
1. Allocate executable memory using `VirtualAlloc`, attempting to use the DLL's preferred `ImageBase`.
2. Copy the DLL's headers from the file buffer into the beginning of the allocated memory.
3. Copy each section's raw data from the file buffer into its correct relative virtual address within the allocated memory. This exercise simulates the work the Windows loader does when mapping a DLL, but performed manually by our Go code.

We'll use `calc_dll.dll` as our target, and reuse much of the logic we created in our PE Header Parser. You can either start with a copy of your `pe_parser.go` from Lab 2.1 or create a new file `manual_mapper.go`. If you are starting with a new file ensure that you have the necessary PE struct definitions (`IMAGE_DOS_HEADER`, `IMAGE_FILE_HEADER`, `IMAGE_OPTIONAL_HEADER64`, `IMAGE_SECTION_HEADER`) and constants (`IMAGE_DOS_SIGNATURE`, `IMAGE_NT_SIGNATURE`) from Lab 2.2.


## Code
```go
//go:build windows
// +build windows

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe" // Needed for pointer conversions with syscall/windows

	"golang.org/x/sys/windows"
)

// --- PE Structures (Ensure these are present from Lab 2.1) ---

type IMAGE_DOS_HEADER struct {
	Magic    uint16     // Magic number (MZ)
	Cblp     uint16     // Bytes on last page of file
	Cp       uint16     // Pages in file
	Crlc     uint16     // Relocations
	Cparhdr  uint16     // Size of header in paragraphs
	MinAlloc uint16     // Minimum extra paragraphs needed
	MaxAlloc uint16     // Maximum extra paragraphs needed
	Ss       uint16     // Initial (relative) SS value
	Sp       uint16     // Initial SP value
	Csum     uint16     // Checksum
	Ip       uint16     // Initial IP value
	Cs       uint16     // Initial (relative) CS value
	Lfarlc   uint16     // File address of relocation table
	Ovno     uint16     // Overlay number
	Res      [4]uint16  // Reserved words
	Oemid    uint16     // OEM identifier (for e_oeminfo)
	Oeminfo  uint16     // OEM information; e_oemid specific
	Res2     [10]uint16 // Reserved words
	Lfanew   int32      // File address of new exe header (PE header offset)
}

type IMAGE_FILE_HEADER struct {
	Machine              uint16 // Architecture type
	NumberOfSections     uint16 // Number of sections
	TimeDateStamp        uint32 // Time and date stamp
	PointerToSymbolTable uint32 // Pointer to symbol table
	NumberOfSymbols      uint32 // Number of symbols
	SizeOfOptionalHeader uint16 // Size of optional header
	Characteristics      uint16 // File characteristics
}

type IMAGE_DATA_DIRECTORY struct {
	VirtualAddress uint32 // RVA of the directory
	Size           uint32 // Size of the directory
}

// Note: This is the 64-bit version
type IMAGE_OPTIONAL_HEADER64 struct {
	Magic                       uint16 // Magic number (0x20b for PE32+)
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32 // RVA of the entry point
	BaseOfCode                  uint32
	ImageBase                   uint64 // Preferred base address
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32 // Total size of the image in memory
	SizeOfHeaders               uint32 // Size of headers (DOS + PE + Section Headers)
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]IMAGE_DATA_DIRECTORY // Array of data directories
}

// Note: This is the 64-bit version
// We don't define IMAGE_NT_HEADERS64 directly as we read Signature, FileHeader,
// and OptionalHeader sequentially after seeking.

type IMAGE_SECTION_HEADER struct {
	Name                 [8]byte // Section name (null-padded)
	VirtualSize          uint32  // Actual size used in memory
	VirtualAddress       uint32  // RVA of the section
	SizeOfRawData        uint32  // Size of section data on disk
	PointerToRawData     uint32  // File offset of section data
	PointerToRelocations uint32  // File offset of relocations
	PointerToLinenumbers uint32  // File offset of line numbers
	NumberOfRelocations  uint16  // Number of relocations
	NumberOfLinenumbers  uint16  // Number of line numbers
	Characteristics      uint32  // Section characteristics (flags like executable, readable, writable)
}

// --- Constants ---
const (
	IMAGE_DOS_SIGNATURE = 0x5A4D     // "MZ"
	IMAGE_NT_SIGNATURE  = 0x00004550 // "PE\0\0"
)

// Helper function to convert null-padded byte array to string
func sectionNameToString(nameBytes [8]byte) string {
	n := bytes.IndexByte(nameBytes[:], 0)
	if n == -1 {
		n = 8
	}
	return string(nameBytes[:n])
}

func main() {
	fmt.Println("[+] Starting Manual DLL Mapper...")

	// --- Step 1: Read DLL and Parse Headers (similar to Lab 2.2) ---
	if len(os.Args) < 2 {
		log.Fatalf("[-] Usage: %s <path_to_dll>\n", os.Args[0])
	}
	dllPath := os.Args[1]
	fmt.Printf("[+] Reading file: %s\n", dllPath)
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read file '%s': %v\n", dllPath, err)
	}

	reader := bytes.NewReader(dllBytes)
	var dosHeader IMAGE_DOS_HEADER
	if err := binary.Read(reader, binary.LittleEndian, &dosHeader); err != nil {
		log.Fatalf("[-] Failed to read DOS header: %v\n", err)
	}
	if dosHeader.Magic != IMAGE_DOS_SIGNATURE {
		log.Fatalf("[-] Invalid DOS signature")
	}
	if _, err := reader.Seek(int64(dosHeader.Lfanew), 0); err != nil {
		log.Fatalf("[-] Failed to seek to NT Headers: %v\n", err)
	}
	var peSignature uint32
	if err := binary.Read(reader, binary.LittleEndian, &peSignature); err != nil {
		log.Fatalf("[-] Failed to read PE signature: %v\n", err)
	}
	if peSignature != IMAGE_NT_SIGNATURE {
		log.Fatalf("[-] Invalid PE signature")
	}
	var fileHeader IMAGE_FILE_HEADER
	if err := binary.Read(reader, binary.LittleEndian, &fileHeader); err != nil {
		log.Fatalf("[-] Failed to read File Header: %v\n", err)
	}
	var optionalHeader IMAGE_OPTIONAL_HEADER64 // Assuming 64-bit
	if err := binary.Read(reader, binary.LittleEndian, &optionalHeader); err != nil {
		log.Fatalf("[-] Failed to read Optional Header: %v\n", err)
	}
	if optionalHeader.Magic != 0x20b {
		log.Printf("[!] Warning: Optional Header Magic is not PE32+ (0x20b).")
	}

	fmt.Println("[+] Parsed PE Headers successfully.")
	fmt.Printf("[+] Target ImageBase: 0x%X\n", optionalHeader.ImageBase)
	fmt.Printf("[+] Target SizeOfImage: 0x%X (%d bytes)\n", optionalHeader.SizeOfImage, optionalHeader.SizeOfImage)

	// --- Step 2: Allocate Memory ---
	fmt.Printf("[+] Allocating 0x%X bytes of memory...\n", optionalHeader.SizeOfImage)
	allocSize := uintptr(optionalHeader.SizeOfImage)
	preferredBase := uintptr(optionalHeader.ImageBase)

	// Try allocating at the preferred base address first
	allocBase, err := windows.VirtualAlloc(preferredBase, allocSize,
		windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)

	if err != nil {
		// If preferred base failed, let the OS choose the address (lpAddress = 0)
		fmt.Printf("[*] Failed to allocate at preferred base 0x%X: %v. Trying arbitrary address...\n", preferredBase, err)
		allocBase, err = windows.VirtualAlloc(0, allocSize,
			windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
		if err != nil {
			log.Fatalf("[-] Failed to allocate memory at arbitrary address: %v\n", err)
		}
	}
	fmt.Printf("[+] Memory allocated successfully at base address: 0x%X\n", allocBase)

	// Ensure the allocated memory is freed when main exits
	defer func() {
		fmt.Printf("[+] Attempting to free allocated memory at 0x%X...\n", allocBase)
		err := windows.VirtualFree(allocBase, 0, windows.MEM_RELEASE)
		if err != nil {
			log.Printf("[!] Warning: Failed to free allocated memory: %v\n", err)
		} else {
			fmt.Println("[+] Allocated memory freed successfully.")
		}
	}()

	// --- Step 3: Copy Headers into Allocated Memory ---
	fmt.Printf("[+] Copying PE headers (%d bytes) to allocated memory...\n", optionalHeader.SizeOfHeaders)
	headerSize := uintptr(optionalHeader.SizeOfHeaders)
	dllBytesPtr := uintptr(unsafe.Pointer(&dllBytes[0])) // Pointer to start of our DLL byte slice

	// Use WriteProcessMemory for potentially safer copy, especially across protection boundaries (though not strictly needed here as we own the memory)
	var bytesWritten uintptr
	err = windows.WriteProcessMemory(windows.CurrentProcess(), allocBase, (*byte)(unsafe.Pointer(dllBytesPtr)), headerSize, &bytesWritten)
	if err != nil || bytesWritten != headerSize {
		log.Fatalf("[-] Failed to copy PE headers to allocated memory: %v (Bytes written: %d)", err, bytesWritten)
	}
	fmt.Printf("[+] Copied %d bytes of headers successfully.\n", bytesWritten)

	// --- Step 4: Copy Sections into Allocated Memory ---
	fmt.Println("[+] Copying sections...")
	// Calculate offset to the first section header: after DOS header + NT signature + file header + optional header
	// Alternatively, it's at dosHeader.Lfanew + 4 (PE Sig) + unsafe.Sizeof(fileHeader) + uintptr(fileHeader.SizeOfOptionalHeader)
	// Or, more simply, the reader is already positioned correctly if we didn't seek after optional header read in the parser part
	// For clarity, let's calculate it:
	firstSectionHeaderOffset := int64(dosHeader.Lfanew) + 4 + int64(unsafe.Sizeof(fileHeader)) + int64(fileHeader.SizeOfOptionalHeader)
	_, err = reader.Seek(firstSectionHeaderOffset, 0) // Position reader at the first section header
	if err != nil {
		log.Fatalf("[-] Failed to seek to first section header at offset 0x%X: %v\n", firstSectionHeaderOffset, err)
	}

	for i := uint16(0); i < fileHeader.NumberOfSections; i++ {
		var sectionHeader IMAGE_SECTION_HEADER
		// Read section header directly from the reader (positioned correctly)
		err = binary.Read(reader, binary.LittleEndian, &sectionHeader)
		if err != nil {
			log.Fatalf("[-] Failed to read Section Header %d: %v\n", i, err)
		}

		sectionName := sectionNameToString(sectionHeader.Name)
		fmt.Printf("  [*] Processing Section %d: '%s'\n", i, sectionName)

		// Skip sections with no raw data (like .bss)
		if sectionHeader.SizeOfRawData == 0 {
			fmt.Printf("    [*] Skipping section '%s' (SizeOfRawData is 0).\n", sectionName)
			continue
		}

		// Calculate source address in the DLL byte slice
		sourceAddr := dllBytesPtr + uintptr(sectionHeader.PointerToRawData)

		// Calculate destination address in the allocated memory block
		destAddr := allocBase + uintptr(sectionHeader.VirtualAddress)

		// Size of data to copy for this section
		sizeToCopy := uintptr(sectionHeader.SizeOfRawData)

		fmt.Printf("    [*] Copying %d bytes from file offset 0x%X to VA 0x%X\n",
			sizeToCopy, sectionHeader.PointerToRawData, destAddr)

		// Copy the section data
		err = windows.WriteProcessMemory(windows.CurrentProcess(), destAddr, (*byte)(unsafe.Pointer(sourceAddr)), sizeToCopy, &bytesWritten)
		if err != nil || bytesWritten != sizeToCopy {
			log.Fatalf("    [-] Failed to copy section '%s': %v (Bytes written: %d)", sectionName, err, bytesWritten)
		}
		fmt.Printf("    [+] Copied section '%s' successfully (%d bytes).\n", sectionName, bytesWritten)
	}

	fmt.Println("[+] All sections copied.")

	// --- Step 5: Self-Check (Basic) ---
	fmt.Println("[+] Manual mapping process complete (Headers and Sections copied).")
	fmt.Println("[+] Self-Check Suggestion: Use a debugger (like Delve for Go or x64dbg for the process)")
	fmt.Println("    to inspect the memory at the allocated base address (0x%X).", allocBase)
	fmt.Println("    Verify that the 'MZ' and 'PE' signatures are present at the start")
	fmt.Println("    and that data corresponding to sections appears at the correct RVAs.")

	// OPTIONAL
	// Keep the program alive briefly if needed for external debugging - UNCOMMENT BELOW IF DESIRED
	// fmt.Println("Press Enter to exit and free memory...")
	// fmt.Scanln()

	fmt.Println("[+] Mapper finished.")
}

// Helper functions (machineTypeToString, magicTypeToString) should be included as in Lab 2.1
func machineTypeToString(machine uint16) string {
	switch machine {
	case 0x0:
		return "Unknown"
	case 0x14c:
		return "x86 (I386)"
	case 0x8664:
		return "x64 (AMD64)"
	case 0xaa64:
		return "ARM64"
	case 0x1c0:
		return "ARM"
	default:
		return "Other"
	}
}

func magicTypeToString(magic uint16) string {
	switch magic {
	case 0x10b:
		return "PE32 (32-bit)"
	case 0x20b:
		return "PE32+ (64-bit)"
	default:
		return "Unknown/Invalid"
	}
}

```


## Code Breakdown

Note I am only explaining new logic that was not already discussed in our PE Parser, obvs if you wanted to understand all the code
here you'd also need to refer back to it. 

### New Imports
- `unsafe`: This package is introduced because manual mapping requires direct memory manipulation that bypasses Go's standard type safety. It's specifically used here to obtain raw pointers (`unsafe.Pointer`) to the beginning of the `dllBytes` slice and to convert between Go pointers and the `uintptr` type required by some Windows API functions.
- `golang.org/x/sys/windows`: This package provides direct access to Windows API calls from Go. It's essential for interacting with the Windows memory management system (`VirtualAlloc`, `VirtualFree`, `WriteProcessMemory`) and for getting a handle to the current process (`CurrentProcess`).

### `main` Function Logic 

#### Step 1: Read DLL and Parse Headers
This step remains largely the same as in `peparser.go`, extracting necessary information like `optionalHeader.SizeOfImage`, `optionalHeader.ImageBase`, `optionalHeader.SizeOfHeaders`, and `fileHeader.NumberOfSections`.

#### Step 2: Allocate Memory
The goal is to reserve a block of virtual memory within the current process large enough to hold the entire DLL image, as specified by `optionalHeader.SizeOfImage`.

The program first attempts to allocate this memory at the DLL's preferred base address, `optionalHeader.ImageBase`. This is done using `windows.VirtualAlloc`. 
- `windows.VirtualAlloc` Parameters:
  - `lpAddress`: The desired starting address. Initially set to `uintptr(optionalHeader.ImageBase)`.
  - `dwSize`: The amount of memory to allocate, `uintptr(optionalHeader.SizeOfImage)`.
  - `flAllocationType`: Flags specifying the type of allocation. `windows.MEM_RESERVE | windows.MEM_COMMIT` reserves the virtual address space *and* allocates physical memory backing for it in one step.
  - `flProtect`: Memory protection constants. `windows.PAGE_EXECUTE_READWRITE` marks the memory region as readable, writable, and executable – necessary for the mapped code and data but generally overly permissive for production scenarios (permissions should ideally be set per section).
        
If allocation at the preferred base fails (e.g., address space already in use), the code attempts `windows.VirtualAlloc` again, but this time passes `0` as the `lpAddress`. This tells the operating system to choose a suitable available address for the allocation.

The actual base address returned by the successful `VirtualAlloc` call is stored in the `allocBase` variable (type `uintptr`).

A `defer` statement is used with `windows.VirtualFree` to ensure the allocated memory block is released back to the system when the `main` function exits, preventing memory leaks. `windows.MEM_RELEASE` is used to decommit and release the entire region.

#### Step 3: Copy Headers
The PE headers (DOS, NT, Section Headers) must be the first thing present in the newly allocated memory block, mimicking how the Windows loader would map them.

The total size of the headers is taken from `optionalHeader.SizeOfHeaders`.

A raw pointer (`uintptr`) to the start of the original DLL data (`dllBytes[0]`) is obtained using `unsafe.Pointer`. This is needed for the `WriteProcessMemory` call.

`windows.WriteProcessMemory` is used to copy the header data:
- `hProcess`: A handle to the target process. `windows.CurrentProcess()` provides a pseudo-handle to the mapper's own process.
- `lpBaseAddress`: The destination address in the target process, which is our `allocBase`.
- `lpBuffer`: The source address of the data to copy. This requires casting the `uintptr` obtained from `dllBytes` back to a `(*byte)(unsafe.Pointer(...))` representation.
- `nSize`: The number of bytes to copy (`headerSize`).
- `lpNumberOfBytesWritten`: A pointer to a variable that receives the actual number of bytes successfully copied.

The number of bytes written is checked to ensure the entire header section was copied correctly.

#### Step 4: Copy Sections
This is the core mapping step where the actual code and data from the DLL file are placed into the allocated memory at their correct relative virtual addresses (RVAs).

The code calculates the file offset where the first `IMAGE_SECTION_HEADER` begins (`firstSectionHeaderOffset`) and uses `reader.Seek` to position the `bytes.Reader` there.

It then loops `fileHeader.NumberOfSections` times. In each iteration:
- It reads the `IMAGE_SECTION_HEADER` struct for the current section from the `reader`.
- It checks if `sectionHeader.SizeOfRawData` is zero. Sections like `.bss` (uninitialized data) often have a `VirtualSize` but no raw data on disk, so they don't need to be copied; the memory is already zeroed by `VirtualAlloc`.
- **Source Address Calculation:** The file offset of the section's data (`sectionHeader.PointerToRawData`) is added to the base pointer of the DLL file data (`dllBytesPtr`) to get the `sourceAddr` in the `dllBytes` slice.
- **Destination Address Calculation:** The section's intended RVA (`sectionHeader.VirtualAddress`) is added to the base address of our allocated memory block (`allocBase`) to get the correct `destAddr` in the process's virtual memory.
- The number of bytes to copy is `sectionHeader.SizeOfRawData`.
- `windows.WriteProcessMemory` is called again, this time copying `sizeToCopy` bytes from `sourceAddr` (in `dllBytes`) to `destAddr` (in the allocated memory block).
- Errors and the number of bytes written are checked for each section.



## Instructions

- Compile the mapper
```shell
GOOS=windows GOARCH=amd64 go build
```

- Then copy it over to target system and invoke from command-line, providing as argument the dll you'd like to analyze, for example

```bash
".\man_mapper.exe .\calc_dll.dll"
```



## Results
```shell
[+] Starting Manual DLL Mapper...
[+] Reading file: .\calc_dll.dll
[+] Parsed PE Headers successfully.
[+] Target ImageBase: 0x26A5B0000
[+] Target SizeOfImage: 0x22000 (139264 bytes)
[+] Allocating 0x22000 bytes of memory...
[+] Memory allocated successfully at base address: 0x26A5B0000
[+] Copying PE headers (1536 bytes) to allocated memory...
[+] Copied 1536 bytes of headers successfully.
[+] Copying sections...
  [*] Processing Section 0: '.text'
    [*] Copying 6144 bytes from file offset 0x600 to VA 0x26A5B1000
    [+] Copied section '.text' successfully (6144 bytes).
  [*] Processing Section 1: '.data'
    [*] Copying 512 bytes from file offset 0x1E00 to VA 0x26A5B3000
    [+] Copied section '.data' successfully (512 bytes).
  [*] Processing Section 2: '.rdata'
    [*] Copying 1536 bytes from file offset 0x2000 to VA 0x26A5B4000
    [+] Copied section '.rdata' successfully (1536 bytes).
  [*] Processing Section 3: '.pdata'
    [*] Copying 1024 bytes from file offset 0x2600 to VA 0x26A5B5000
    [+] Copied section '.pdata' successfully (1024 bytes).
  [*] Processing Section 4: '.xdata'
    [*] Copying 512 bytes from file offset 0x2A00 to VA 0x26A5B6000
    [+] Copied section '.xdata' successfully (512 bytes).
  [*] Processing Section 5: '.bss'
    [*] Skipping section '.bss' (SizeOfRawData is 0).
  [*] Processing Section 6: '.edata'
    [*] Copying 512 bytes from file offset 0x2C00 to VA 0x26A5B8000
    [+] Copied section '.edata' successfully (512 bytes).
  [*] Processing Section 7: '.idata'
    [*] Copying 2048 bytes from file offset 0x2E00 to VA 0x26A5B9000
    [+] Copied section '.idata' successfully (2048 bytes).
  [*] Processing Section 8: '.tls'
    [*] Copying 512 bytes from file offset 0x3600 to VA 0x26A5BA000
    [+] Copied section '.tls' successfully (512 bytes).
  [*] Processing Section 9: '.reloc'
    [*] Copying 512 bytes from file offset 0x3800 to VA 0x26A5BB000
    [+] Copied section '.reloc' successfully (512 bytes).
  [*] Processing Section 10: '/4'
    [*] Copying 1024 bytes from file offset 0x3A00 to VA 0x26A5BC000
    [+] Copied section '/4' successfully (1024 bytes).
  [*] Processing Section 11: '/19'
    [*] Copying 37888 bytes from file offset 0x3E00 to VA 0x26A5BD000
    [+] Copied section '/19' successfully (37888 bytes).
  [*] Processing Section 12: '/31'
    [*] Copying 6656 bytes from file offset 0xD200 to VA 0x26A5C7000
    [+] Copied section '/31' successfully (6656 bytes).
  [*] Processing Section 13: '/45'
    [*] Copying 6656 bytes from file offset 0xEC00 to VA 0x26A5C9000
    [+] Copied section '/45' successfully (6656 bytes).
  [*] Processing Section 14: '/57'
    [*] Copying 2560 bytes from file offset 0x10600 to VA 0x26A5CB000
    [+] Copied section '/57' successfully (2560 bytes).
  [*] Processing Section 15: '/70'
    [*] Copying 512 bytes from file offset 0x11000 to VA 0x26A5CC000
    [+] Copied section '/70' successfully (512 bytes).
  [*] Processing Section 16: '/81'
    [*] Copying 6144 bytes from file offset 0x11200 to VA 0x26A5CD000
    [+] Copied section '/81' successfully (6144 bytes).
  [*] Processing Section 17: '/97'
    [*] Copying 5120 bytes from file offset 0x12A00 to VA 0x26A5CF000
    [+] Copied section '/97' successfully (5120 bytes).
  [*] Processing Section 18: '/113'
    [*] Copying 512 bytes from file offset 0x13E00 to VA 0x26A5D1000
    [+] Copied section '/113' successfully (512 bytes).
[+] All sections copied.
[+] Manual mapping process complete (Headers and Sections copied).
[+] Self-Check Suggestion: Use a debugger (like Delve for Go or x64dbg for the process)
    to inspect the memory at the allocated base address (0x%X). 10374283264
    Verify that the 'MZ' and 'PE' signatures are present at the start
    and that data corresponding to sections appears at the correct RVAs.
[+] Mapper finished.
[+] Attempting to free allocated memory at 0x26A5B0000...
[+] Allocated memory freed successfully.
```

## Discussion
- Note that once again we found the same address for ImageBase as in Labs 2.1 and 2.2 - `0x26A5B0000`.
- Based on SizeOfImage (`0x22000`) we then allocate that amount of memory. 
- In this case we were indeed able to allocate at the preferred address `0x26A5B0000`, meaning that no relocations would need to take place.
- We then copy all the headers, and then iterate through each section, doing the same.

## Conclusion
Great, we've covered considerable ground in these first three modules - we can now parse a PE file, manually allocate memory,
and copy required sections into that memory. But, as mentioned, we're not quite done ready to run shellcode yet. 

In the next module we're going to cover two important steps:
- First, in the case that are not able to allocate memory at our preferred base address, we need to perform relocations.
- Next, it's more often than not the case that our DLL in turn uses function from other system DLLs, meaning we need to resolve those imports.

Let's roll on.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "mapping.md" >}})
[|NEXT|]({{< ref "../module04/reloc.md" >}})