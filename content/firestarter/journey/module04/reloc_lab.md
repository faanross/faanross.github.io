---
showTableOfContents: true
title: "Intentional Base Relocation (Lab 4.1)"
type: "page"
---

## Goal

Let's modify our Go program from Lab 3.1 (`manual_mapper.go`) to handle base relocations. If the DLL could not be loaded at its preferred `ImageBase`, this lab adds the logic to parse the DLL's relocation table and patch all necessary addresses within the mapped image based on the actual load address.

Now, even with ASLR, it's often the case, and especially for DLLs loaded reflectively into a process that hasn't had its address space heavily populated yet, that the preferred `ImageBase` is often available, meaning the `delta` is zero and the relocation code path doesn't run. So in this lab we're going to artificially "force" relocations here, otherwise we'll implement logic which won't even be tested.

So in addition to introducing our relocation logic we'll  modify the start of the allocation process to deliberately occupy the DLL's preferred `ImageBase` _before_ attempting to allocate the main block for the DLL. This will force the main allocation to use the fallback mechanism (letting the OS choose an address), resulting in a non-zero delta and ensuring your relocation logic gets tested

Note we'll once again load `call_dll.dll`.


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
	"runtime" // Import runtime package
	"unsafe"  // Needed for pointer conversions with syscall/windows

	"golang.org/x/sys/windows"
)

// --- Existing PE Structures ---
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

// --- NEW: Relocation Structures/Constants  ---
type IMAGE_BASE_RELOCATION struct { //nolint:revive // Windows struct
	VirtualAddress uint32 // RVA of the page this block applies to
	SizeOfBlock    uint32 // Total size of this relocation block (including header)
}

const (
	IMAGE_DOS_SIGNATURE             = 0x5A4D     // "MZ"
	IMAGE_NT_SIGNATURE              = 0x00004550 // "PE\0\0"
	IMAGE_DIRECTORY_ENTRY_BASERELOC = 5          // Base Relocation Table index in DataDirectory
	IMAGE_REL_BASED_DIR64           = 10         // Relocation type for 64-bit addresses (x64)
	IMAGE_REL_BASED_ABSOLUTE        = 0          // Padding/nop relocation type
	// Adding Memory constants if not already implicitly available via windows package
	MEM_COMMIT             = 0x00001000
	MEM_RESERVE            = 0x00002000
	MEM_RELEASE            = 0x8000
	PAGE_READWRITE         = 0x04
	PAGE_EXECUTE_READWRITE = 0x40
)

// --- Existing Helper Functions ---
func sectionNameToString(nameBytes [8]byte) string {
	n := bytes.IndexByte(nameBytes[:], 0)
	if n == -1 {
		n = 8
	}
	return string(nameBytes[:n])
}

// --- Main Function ---
func main() {
	// Ensure running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("[-] This program must be run on Windows.")
	}
	fmt.Println("[+] Starting Manual DLL Mapper (with Forced Relocation)...")

	// --- Step 1: Read DLL and Parse Headers ---
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
	var optionalHeader IMAGE_OPTIONAL_HEADER64
	if err := binary.Read(reader, binary.LittleEndian, &optionalHeader); err != nil {
		log.Fatalf("[-] Failed to read Optional Header: %v\n", err)
	}
	if optionalHeader.Magic != 0x20b {
		log.Printf("[!] Warning: Optional Header Magic is not PE32+ (0x20b).")
	}
	fmt.Println("[+] Parsed PE Headers successfully.")
	fmt.Printf("[+] Target ImageBase: 0x%X\n", optionalHeader.ImageBase)
	fmt.Printf("[+] Target SizeOfImage: 0x%X (%d bytes)\n", optionalHeader.SizeOfImage, optionalHeader.SizeOfImage)
	// --- End Step 1 ---

	// --- *** HERE WE ARE FORCING RELOCATION BY OCCUPYING IMAGEBASE, WE'LL REMOVE THIS AFTER THIS LAB *** ---
	preferredImageBase := uintptr(optionalHeader.ImageBase) // Get preferred base from parsed header
	fmt.Printf("[+] Attempting to reserve preferred ImageBase (0x%X) to force relocation...\n", preferredImageBase)
	blockingAllocSize := uintptr(4096) // Allocate just one page
	blockingAllocBase, errBlock := windows.VirtualAlloc(preferredImageBase, blockingAllocSize, MEM_COMMIT|MEM_RESERVE, PAGE_READWRITE)
	if errBlock != nil {
		log.Printf("[!] Warning: Could not allocate blocking memory at preferred base 0x%X: %v. Relocation might not be forced.", preferredImageBase, errBlock)
	} else {
		fmt.Printf("[+] Successfully allocated blocking memory at 0x%X. Relocation should be forced.\n", blockingAllocBase)
		// Defer release of the blocking allocation
		defer func() {
			fmt.Println("[+] Freeing blocking allocation...")
			errFree := windows.VirtualFree(blockingAllocBase, 0, MEM_RELEASE)
			if errFree != nil {
				log.Printf("[!] Warning: Failed to free blocking allocation: %v", errFree)
			} else {
				fmt.Println("[+] Blocking allocation freed.")
			}
		}()
	}
	// --- *** End  *** ---

	// --- Step 2: Allocate Memory for DLL ---
	fmt.Printf("[+] Allocating 0x%X bytes of memory for DLL...\n", optionalHeader.SizeOfImage)
	allocSize := uintptr(optionalHeader.SizeOfImage)
	// preferredBase already defined above
	allocBase, err := windows.VirtualAlloc(preferredImageBase, allocSize, windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
	if err != nil {
		// If preferred base failed (as expected), let the OS choose the address
		fmt.Printf("[*] Failed to allocate at preferred base 0x%X (EXPECTED): %v. Trying arbitrary address...\n", preferredImageBase, err)
		allocBase, err = windows.VirtualAlloc(0, allocSize, windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
		if err != nil {
			log.Fatalf("[-] Failed to allocate memory at arbitrary address: %v\n", err)
		}
	} else {
		// This case means our blocking allocation failed AND the preferred base was free
		fmt.Println("[*] Allocated at preferred base unexpectedly (blocking allocation might have failed).")
	}
	fmt.Printf("[+] DLL memory allocated successfully at actual base address: 0x%X\n", allocBase)
	// Defer release of the main DLL allocation
	defer func() {
		fmt.Printf("[+] Attempting to free main DLL allocation at 0x%X...\n", allocBase)
		err := windows.VirtualFree(allocBase, 0, windows.MEM_RELEASE)
		if err != nil {
			log.Printf("[!] Warning: Failed to free main DLL memory: %v\n", err)
		} else {
			fmt.Println("[+] Main DLL memory freed successfully.")
		}
	}()
	// --- End Step 2 ---

	// --- Step 3: Copy Headers into Allocated Memory ---
	fmt.Printf("[+] Copying PE headers (%d bytes) to allocated memory...\n", optionalHeader.SizeOfHeaders)
	headerSize := uintptr(optionalHeader.SizeOfHeaders)
	dllBytesPtr := uintptr(unsafe.Pointer(&dllBytes[0]))
	var bytesWritten uintptr
	err = windows.WriteProcessMemory(windows.CurrentProcess(), allocBase, (*byte)(unsafe.Pointer(dllBytesPtr)), headerSize, &bytesWritten)
	if err != nil || bytesWritten != headerSize {
		log.Fatalf("[-] Failed to copy PE headers to allocated memory: %v (Bytes written: %d)", err, bytesWritten)
	}
	fmt.Printf("[+] Copied %d bytes of headers successfully.\n", bytesWritten)
	// --- End Step 3 ---

	// --- Step 4: Copy Sections into Allocated Memory ---
	// NoteL Code adjusted slightly from Lab 3.1 to read section headers from allocBase
	fmt.Println("[+] Copying sections...")
	firstSectionHeaderOffset := uintptr(dosHeader.Lfanew) + 4 + unsafe.Sizeof(fileHeader) + uintptr(fileHeader.SizeOfOptionalHeader)
	sectionHeaderSize := unsafe.Sizeof(IMAGE_SECTION_HEADER{}) // Get size of struct for iteration

	for i := uint16(0); i < fileHeader.NumberOfSections; i++ {
		// *** Read section header from the *mapped* headers in allocBase ***
		currentSectionHeaderAddr := allocBase + firstSectionHeaderOffset + uintptr(i)*sectionHeaderSize
		sectionHeader := (*IMAGE_SECTION_HEADER)(unsafe.Pointer(currentSectionHeaderAddr))
		// *** End modification for reading section header ***

		sectionName := sectionNameToString(sectionHeader.Name)
		fmt.Printf("  [*] Processing Section %d: '%s'\n", i, sectionName)
		if sectionHeader.SizeOfRawData == 0 {
			fmt.Printf("    [*] Skipping section '%s' (SizeOfRawData is 0).\n", sectionName)
			continue
		}
		sourceAddr := dllBytesPtr + uintptr(sectionHeader.PointerToRawData)
		destAddr := allocBase + uintptr(sectionHeader.VirtualAddress)
		sizeToCopy := uintptr(sectionHeader.SizeOfRawData)
		fmt.Printf("    [*] Copying %d bytes from file offset 0x%X to VA 0x%X\n", sizeToCopy, sectionHeader.PointerToRawData, destAddr)
		err = windows.WriteProcessMemory(windows.CurrentProcess(), destAddr, (*byte)(unsafe.Pointer(sourceAddr)), sizeToCopy, &bytesWritten)
		if err != nil || bytesWritten != sizeToCopy {
			log.Fatalf("    [-] Failed to copy section '%s': %v (Bytes written: %d)", sectionName, err, bytesWritten)
		}
		fmt.Printf("    [+] Copied section '%s' successfully (%d bytes).\n", sectionName, bytesWritten)
	}
	fmt.Println("[+] All sections copied.")
	// --- End Step 4 ---

	// --- *** NEW: Step 5: Process Base Relocations *** ---
	fmt.Println("[+] Checking if base relocations are needed...")
	delta := int64(allocBase) - int64(optionalHeader.ImageBase) // Use the parsed preferred base

	if delta == 0 {
		// This should NOT happen if the blocking allocation worked
		fmt.Println("[!] Image loaded at preferred base unexpectedly. Relocations not tested.")
	} else {
		fmt.Printf("[+] Image loaded at non-preferred base (Delta: 0x%X). Processing relocations...\n", delta)
		// Find the Base Relocation Directory entry using the parsed optionalHeader
		relocDirEntry := optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_BASERELOC]
		relocDirRVA := relocDirEntry.VirtualAddress
		relocDirSize := relocDirEntry.Size

		if relocDirRVA == 0 || relocDirSize == 0 {
			fmt.Println("[!] Warning: Image rebased, but no relocation directory found or empty.")
		} else {
			fmt.Printf("[+] Relocation Directory found at RVA 0x%X, Size 0x%X\n", relocDirRVA, relocDirSize)
			relocTableBase := allocBase + uintptr(relocDirRVA) // VA of the start of the .reloc section in allocBase
			relocTableEnd := relocTableBase + uintptr(relocDirSize)
			currentBlockAddr := relocTableBase
			totalFixups := 0

			// Iterate through IMAGE_BASE_RELOCATION blocks
			for currentBlockAddr < relocTableEnd {
				// Read block header directly from allocBase memory
				blockHeader := (*IMAGE_BASE_RELOCATION)(unsafe.Pointer(currentBlockAddr))
				// Check for end marker or invalid size
				if blockHeader.VirtualAddress == 0 || blockHeader.SizeOfBlock <= uint32(unsafe.Sizeof(IMAGE_BASE_RELOCATION{})) {
					break // Stop processing
				}

				numEntries := (blockHeader.SizeOfBlock - uint32(unsafe.Sizeof(IMAGE_BASE_RELOCATION{}))) / 2
				entryPtr := currentBlockAddr + unsafe.Sizeof(IMAGE_BASE_RELOCATION{}) // Pointer to first entry

				// Iterate through the 16-bit entries in this block
				for i := uint32(0); i < numEntries; i++ {
					// Read entry directly from allocBase memory
					entry := *(*uint16)(unsafe.Pointer(entryPtr + uintptr(i*2)))
					relocType := entry >> 12
					offset := entry & 0xFFF

					if relocType == IMAGE_REL_BASED_DIR64 {
						// Calculate the absolute VA within allocBase where the patch needs to be applied
						patchAddr := allocBase + uintptr(blockHeader.VirtualAddress) + uintptr(offset)
						// Read the original 64-bit value directly from allocBase memory
						originalValuePtr := (*uint64)(unsafe.Pointer(patchAddr))
						originalValue := *originalValuePtr
						// Apply the delta
						newValue := uint64(int64(originalValue) + delta)
						// Write the new value back directly into allocBase memory
						*originalValuePtr = newValue
						totalFixups++
					} else if relocType != IMAGE_REL_BASED_ABSOLUTE {
						fmt.Printf("        [!] Warning: Skipping unhandled relocation type %d at offset 0x%X\n", relocType, offset)
					}
				}
				// Move to the next block header
				currentBlockAddr += uintptr(blockHeader.SizeOfBlock)
			}
			fmt.Printf("[+] Relocation processing complete. Total fixups applied: %d\n", totalFixups)
		}
	}
	// --- *** End Step 5 *** ---

	// --- Step 6: Self-Check -
	fmt.Println("[+] Manual mapping process complete (Headers, Sections copied, Relocations potentially applied).")
	fmt.Println("[+] Self-Check Suggestion: Use a debugger (like Delve for Go or x64dbg for the process)")
	fmt.Println("    to inspect the memory at the allocated base address (0x%X).", allocBase)
	fmt.Println("    Verify that the 'MZ' and 'PE' signatures are present at the start")
	fmt.Println("    and that data corresponding to sections appears at the correct RVAs.")
	fmt.Println("    If relocations occurred, check known absolute addresses (if any) were patched.")

	fmt.Println("\n[+] Press Enter to free memory and exit.")
	fmt.Scanln()

	fmt.Println("[+] Mapper finished.")
	// Deferred VirtualFree calls will execute now
}

// Helper functions
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
Note I am only explaining the logic that was added or significantly changed compared to our Manual Mapper.

### New Structs/Constants
#### Structs
**`IMAGE_BASE_RELOCATION` struct**: Added to define the layout of relocation block headers found in the `.reloc` section. Contains `VirtualAddress` (RVA of the page) and `SizeOfBlock`.

#### Constants
**`IMAGE_DIRECTORY_ENTRY_BASERELOC` (5)**: Index for the relocation table in the Data Directory.

**`IMAGE_REL_BASED_DIR64` (10)**: Type indicating a 64-bit address requires patching by adding the delta.

**`IMAGE_REL_BASED_ABSOLUTE` (0)**: Type indicating a padding entry to be skipped.

**Memory constants (`MEM_COMMIT`, `MEM_RESERVE`, `MEM_RELEASE`, `PAGE_READWRITE`, `PAGE_EXECUTE_READWRITE`)**: Added explicitly for clarity, though they are available in the `windows` package.

#### Imports
`errors`, `runtime`, `syscall` packages were added/ensured for error handling, OS checks, and potential future syscall use.


### Force Relocation by Occupying Preferred Base
- This is where we now force relocation by intentionally occupying the `ImageBase`.
- We do this by retrieving  `preferredImageBase` from the parsed `optionalHeader`.
- We then call `windows.VirtualAlloc` specifically requesting a single page *at* that `preferredImageBase`.

### Step 2: Allocate Memory for DLL
- We're still going to attempt  `windows.VirtualAlloc` at `preferredImageBase`.
- But of course now since we've occupied a page at the location, we expect it to fail
- Our existing fallback logic (`windows.VirtualAlloc` with 0 address) remains the same as before, but it should now execute, resulting in `allocBase` being different from `preferredImageBase`.

###  Step 5: Process Base Relocations
This is where all the important base relocation logic we learned about in Theory 4.1 is now implemented:
- **Calculate Delta:** Calculates `delta = int64(allocBase) - int64(preferredImageBase)`. This difference drives the patching process.
- **Check Delta:** Proceeds only if `delta` is non-zero.
- **Locate Relocation Directory:** Accesses `optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_BASERELOC]` to get the RVA and Size of the relocation table. Handles cases where the directory doesn't exist or is empty.
- **Calculate Table Addresses:** Determines the start (`relocTableBase`) and end (`relocTableEnd`) VAs of the relocation table within the `allocBase` memory region.
- **Iterate Blocks:** Enters a `for` loop that processes the table block by block (`IMAGE_BASE_RELOCATION`).
    * Reads the `blockHeader` directly from memory at `currentBlockAddr`.
    * Checks for a zero-sized block to terminate.
    * Calculates the number of 16-bit entries (`numEntries`) in the current block.
    * Calculates the starting address (`entryPtr`) of the entries following the header.
- **Iterate Entries:** Enters an inner `for` loop iterating `numEntries` times.
    * Reads the 16-bit `entry` value directly from memory.
    * Extracts the `relocType` (top 4 bits) and `offset` (bottom 12 bits).
- **Apply Fixup:** Checks if `relocType == IMAGE_REL_BASED_DIR64`.
    * Calculates the absolute VA (`patchAddr`) within `allocBase` that needs patching (`allocBase + blockRVA + offset`).
    * Uses `unsafe.Pointer` to cast `patchAddr` to `*uint64` to read the `originalValue` directly from memory.
    * Calculates the `newValue` by adding the `delta`.
    * Writes the `newValue` back to the memory location using the `*uint64` pointer.
    * Increments `totalFixups`.
    * Skips `IMAGE_REL_BASED_ABSOLUTE` and warns about other types.
- **Advance:** Moves `currentBlockAddr` to the next block using `blockHeader.SizeOfBlock`.

## Instructions

- Compile the base relocator.

```shell
GOOS=windows GOARCH=amd64 go build
```

- Then copy it over to target system and invoke from command-line, providing as argument the dll you’d like to analyze, for example

```bash
".\base_reloc.exe .\calc_dll.dll"
```


## Results

```shell
PS C:\Users\vuilhond\Desktop> .\base_reloc.exe .\calc_dll.dll
[+] Starting Manual DLL Mapper (with Forced Relocation)...
[+] Reading file: .\calc_dll.dll
[+] Parsed PE Headers successfully.
[+] Target ImageBase: 0x26A5B0000
[+] Target SizeOfImage: 0x22000 (139264 bytes)
[+] Attempting to reserve preferred ImageBase (0x26A5B0000) to force relocation...
[+] Successfully allocated blocking memory at 0x26A5B0000. Relocation should be forced.
[+] Allocating 0x22000 bytes of memory for DLL...
[*] Failed to allocate at preferred base 0x26A5B0000 (EXPECTED): Attempt to access invalid address.. Trying arbitrary address...
[+] DLL memory allocated successfully at actual base address: 0x1D74D540000
[+] Copying PE headers (1536 bytes) to allocated memory...
[+] Copied 1536 bytes of headers successfully.
[+] Copying sections...
  [*] Processing Section 0: '.text'
    [*] Copying 6144 bytes from file offset 0x600 to VA 0x1D74D541000
    [+] Copied section '.text' successfully (6144 bytes).
  [*] Processing Section 1: '.data'
    [*] Copying 512 bytes from file offset 0x1E00 to VA 0x1D74D543000
    [+] Copied section '.data' successfully (512 bytes).
  [*] Processing Section 2: '.rdata'
    [*] Copying 1536 bytes from file offset 0x2000 to VA 0x1D74D544000
    [+] Copied section '.rdata' successfully (1536 bytes).
  [*] Processing Section 3: '.pdata'
    [*] Copying 1024 bytes from file offset 0x2600 to VA 0x1D74D545000
    [+] Copied section '.pdata' successfully (1024 bytes).
  [*] Processing Section 4: '.xdata'
    [*] Copying 512 bytes from file offset 0x2A00 to VA 0x1D74D546000
    [+] Copied section '.xdata' successfully (512 bytes).
  [*] Processing Section 5: '.bss'
    [*] Skipping section '.bss' (SizeOfRawData is 0).
  [*] Processing Section 6: '.edata'
    [*] Copying 512 bytes from file offset 0x2C00 to VA 0x1D74D548000
    [+] Copied section '.edata' successfully (512 bytes).
  [*] Processing Section 7: '.idata'
    [*] Copying 2048 bytes from file offset 0x2E00 to VA 0x1D74D549000
    [+] Copied section '.idata' successfully (2048 bytes).
  [*] Processing Section 8: '.tls'
    [*] Copying 512 bytes from file offset 0x3600 to VA 0x1D74D54A000
    [+] Copied section '.tls' successfully (512 bytes).
  [*] Processing Section 9: '.reloc'
    [*] Copying 512 bytes from file offset 0x3800 to VA 0x1D74D54B000
    [+] Copied section '.reloc' successfully (512 bytes).
  [*] Processing Section 10: '/4'
    [*] Copying 1024 bytes from file offset 0x3A00 to VA 0x1D74D54C000
    [+] Copied section '/4' successfully (1024 bytes).
  [*] Processing Section 11: '/19'
    [*] Copying 37888 bytes from file offset 0x3E00 to VA 0x1D74D54D000
    [+] Copied section '/19' successfully (37888 bytes).
  [*] Processing Section 12: '/31'
    [*] Copying 6656 bytes from file offset 0xD200 to VA 0x1D74D557000
    [+] Copied section '/31' successfully (6656 bytes).
  [*] Processing Section 13: '/45'
    [*] Copying 6656 bytes from file offset 0xEC00 to VA 0x1D74D559000
    [+] Copied section '/45' successfully (6656 bytes).
  [*] Processing Section 14: '/57'
    [*] Copying 2560 bytes from file offset 0x10600 to VA 0x1D74D55B000
    [+] Copied section '/57' successfully (2560 bytes).
  [*] Processing Section 15: '/70'
    [*] Copying 512 bytes from file offset 0x11000 to VA 0x1D74D55C000
    [+] Copied section '/70' successfully (512 bytes).
  [*] Processing Section 16: '/81'
    [*] Copying 6144 bytes from file offset 0x11200 to VA 0x1D74D55D000
    [+] Copied section '/81' successfully (6144 bytes).
  [*] Processing Section 17: '/97'
    [*] Copying 5120 bytes from file offset 0x12A00 to VA 0x1D74D55F000
    [+] Copied section '/97' successfully (5120 bytes).
  [*] Processing Section 18: '/113'
    [*] Copying 512 bytes from file offset 0x13E00 to VA 0x1D74D561000
    [+] Copied section '/113' successfully (512 bytes).
[+] All sections copied.
[+] Checking if base relocations are needed...
[+] Image loaded at non-preferred base (Delta: 0x1D4E2F90000). Processing relocations...
[+] Relocation Directory found at RVA 0xB000, Size 0x64
[+] Relocation processing complete. Total fixups applied: 41
[+] Manual mapping process complete (Headers, Sections copied, Relocations potentially applied).
[+] Self-Check Suggestion: Use a debugger (like Delve for Go or x64dbg for the process)
    to inspect the memory at the allocated base address (0x%X). 2024226947072
    Verify that the 'MZ' and 'PE' signatures are present at the start
    and that data corresponding to sections appears at the correct RVAs.
    If relocations occurred, check known absolute addresses (if any) were patched.

[+] Press Enter to free memory and exit.

[+] Mapper finished.
[+] Attempting to free main DLL allocation at 0x1D74D540000...
[+] Main DLL memory freed successfully.
[+] Freeing blocking allocation...
[+] Blocking allocation freed.
```



## Discussion
Given that we've intentionally occupied ImageBase, the output is exactly what we expect:
- ** `Successfully allocated blocking memory at 0x26A5B0000. Relocation should be forced.` - This confirms the preferred base was occupied.
- `Failed to allocate at preferred base 0x26A5B0000 (EXPECTED): ... Trying arbitrary address...` followed by `DLL memory allocated successfully at actual base address: 0x1D74D540000` - This shows the loader correctly fell back to an OS-chosen address.
- `Image loaded at non-preferred base (Delta: 0x1D4E2F90000). Processing relocations...` - The non-zero delta correctly triggered the relocation code path.
- `Relocation processing complete. Total fixups applied: 41` - The code successfully parsed the `.reloc` section and applied the necessary patches.

## Conclusion
Most excellent. Let's move on ahead and get cracking with our IAT resolution.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "iat.md" >}})
[|NEXT|]({{< ref "iat_lab.md" >}})