---
showTableOfContents: true
title: "The PE File Format: Simple PE Parser Lab"
type: "page"
---


## Objective

Build a Go application that programmatically extracts all the critical PE values we manually found in Lab 2. This solidifies your understanding by implementing the parsing logic yourself, showing exactly how each field connects to your conceptual model of the PE format.

## What You'll Build

A command-line PE parser that:
- Reads any PE file
- Extracts all offensive-relevant values
- Displays them in a formatted table
- Validates PE structure
- Assesses security posture

**Output preview:**
```
PE Analysis Report: simple.exe
================================================================================

[DOS HEADER]
  Magic (MZ):                    0x5A4D ✓
  PE Header Offset (e_lfanew):   0x00F0

[FILE HEADER]
  Architecture:                  x64 (0x8664)
  Number of Sections:            6
  Timestamp:                     2024-01-15 14:23:10
  Characteristics:               0x0022
    ✓ Executable Image
    ✗ DLL
    ✓ Has Relocations (not stripped)

[OPTIONAL HEADER]
  Magic:                         0x020B (PE32+)
  Entry Point (RVA):             0x00012A40
  Image Base:                    0x0000000140000000
  Section Alignment:             0x00001000 (4096 bytes)
  File Alignment:                0x00000200 (512 bytes)
  Size of Image:                 0x000B8000
  Size of Headers:               0x00000400
  
[SECURITY FEATURES]
  ASLR (DYNAMIC_BASE):           ✓ ENABLED
  DEP (NX_COMPAT):               ✓ ENABLED
  
[DATA DIRECTORIES]
  Import Directory:              RVA 0x000A2000, Size 240 bytes
  Base Relocation:               RVA 0x000B0000, Size 5248 bytes
  Resource Directory:            RVA 0x000A8000, Size 1024 bytes
  TLS Directory:                 Not present
  
[SECTIONS]
  .text    RVA 0x00001000  Raw 0x00000400  Perms: R-X  Size: 688KB
  .rdata   RVA 0x000AB000  Raw 0x000AA600  Perms: R--  Size: 80KB
  .data    RVA 0x000BF000  Raw 0x000B7800  Perms: RW-  Size: 12KB

[IMPORT ANALYSIS]
  Imported DLLs: 3
    KERNEL32.dll (28 functions)
      ⚠️  VirtualAlloc
      ⚠️  LoadLibraryA
    ADVAPI32.dll (5 functions)
    WS2_32.dll (12 functions)

[ASSESSMENT]
  Can Rebase:                    YES ✓
  Process Hollowing Viable:      YES ✓
  Suspicious Imports:            YES ⚠️
  Overall Security:              MODERATE
```

---

## Code


```go
//go:build windows
// +build windows

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// These structures mirror the PE format we've been studying

// IMAGE_DOS_HEADER - The legacy DOS header (64 bytes)
type IMAGE_DOS_HEADER struct {
	E_magic    uint16     // Magic number "MZ" (0x5A4D)
	E_cblp     uint16     // Bytes on last page of file
	E_cp       uint16     // Pages in file
	E_crlc     uint16     // Relocations
	E_cparhdr  uint16     // Size of header in paragraphs
	E_minalloc uint16     // Minimum extra paragraphs needed
	E_maxalloc uint16     // Maximum extra paragraphs needed
	E_ss       uint16     // Initial (relative) SS value
	E_sp       uint16     // Initial SP value
	E_csum     uint16     // Checksum
	E_ip       uint16     // Initial IP value
	E_cs       uint16     // Initial (relative) CS value
	E_lfarlc   uint16     // File address of relocation table
	E_ovno     uint16     // Overlay number
	E_res      [4]uint16  // Reserved words
	E_oemid    uint16     // OEM identifier
	E_oeminfo  uint16     // OEM information
	E_res2     [10]uint16 // Reserved words
	E_lfanew   int32      // File address of new exe header (THE KEY FIELD!)
}

// IMAGE_FILE_HEADER - Core PE metadata (20 bytes)
type IMAGE_FILE_HEADER struct {
	Machine              uint16 // Architecture (0x8664 = x64, 0x014C = x86)
	NumberOfSections     uint16 // Count of sections
	TimeDateStamp        uint32 // Unix timestamp of compilation
	PointerToSymbolTable uint32 // Deprecated
	NumberOfSymbols      uint32 // Deprecated
	SizeOfOptionalHeader uint16 // Size of optional header that follows
	Characteristics      uint16 // File flags (DLL, relocations, etc.)
}

// IMAGE_OPTIONAL_HEADER64 - The "optional" (but required) header for x64
type IMAGE_OPTIONAL_HEADER64 struct {
	Magic                       uint16 // 0x020B for PE32+ (64-bit)
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32 // RVA where execution begins
	BaseOfCode                  uint32 // RVA of code section start
	ImageBase                   uint64 // Preferred load address
	SectionAlignment            uint32 // Section alignment in memory
	FileAlignment               uint32 // Section alignment on disk
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32 // Total size in memory
	SizeOfHeaders               uint32 // Size of all headers
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16 // Security flags (ASLR, DEP, etc.)
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]IMAGE_DATA_DIRECTORY // The roadmap!
}

// IMAGE_DATA_DIRECTORY - Points to specialized structures
type IMAGE_DATA_DIRECTORY struct {
	VirtualAddress uint32 // RVA to the data
	Size           uint32 // Size in bytes
}

// IMAGE_SECTION_HEADER - Describes one section (40 bytes)
type IMAGE_SECTION_HEADER struct {
	Name                 [8]byte // Section name (not always null-terminated!)
	VirtualSize          uint32  // Size in memory
	VirtualAddress       uint32  // RVA where section loads
	SizeOfRawData        uint32  // Size on disk
	PointerToRawData     uint32  // File offset where section starts
	PointerToRelocations uint32  // Obsolete
	PointerToLinenumbers uint32  // Obsolete
	NumberOfRelocations  uint16  // Obsolete
	NumberOfLinenumbers  uint16  // Obsolete
	Characteristics      uint32  // Section flags (permissions)
}

// IMAGE_IMPORT_DESCRIPTOR - Describes imports from one DLL
type IMAGE_IMPORT_DESCRIPTOR struct {
	OriginalFirstThunk uint32 // RVA to Import Name Table
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32 // RVA to DLL name string
	FirstThunk         uint32 // RVA to Import Address Table
}

// Constants for PE parsing
const (
	IMAGE_DOS_SIGNATURE = 0x5A4D     // "MZ"
	IMAGE_NT_SIGNATURE  = 0x00004550 // "PE\0\0"

	IMAGE_FILE_MACHINE_I386  = 0x014C
	IMAGE_FILE_MACHINE_AMD64 = 0x8664

	IMAGE_FILE_DLL                 = 0x2000
	IMAGE_FILE_EXECUTABLE_IMAGE    = 0x0002
	IMAGE_FILE_RELOCS_STRIPPED     = 0x0001
	IMAGE_FILE_LARGE_ADDRESS_AWARE = 0x0020

	IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE = 0x0040 // ASLR
	IMAGE_DLLCHARACTERISTICS_NX_COMPAT    = 0x0100 // DEP
	IMAGE_DLLCHARACTERISTICS_NO_SEH       = 0x0400

	IMAGE_SCN_MEM_EXECUTE = 0x20000000
	IMAGE_SCN_MEM_READ    = 0x40000000
	IMAGE_SCN_MEM_WRITE   = 0x80000000

	// Data Directory indices
	IMAGE_DIRECTORY_ENTRY_EXPORT    = 0
	IMAGE_DIRECTORY_ENTRY_IMPORT    = 1
	IMAGE_DIRECTORY_ENTRY_RESOURCE  = 2
	IMAGE_DIRECTORY_ENTRY_BASERELOC = 5
	IMAGE_DIRECTORY_ENTRY_DEBUG     = 6
	IMAGE_DIRECTORY_ENTRY_TLS       = 9
	IMAGE_DIRECTORY_ENTRY_IAT       = 12
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pe-parser <executable>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read entire PE file into memory
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nPE Analysis Report: %s\n", filename)
	fmt.Println("=" + repeat("=", 79))

	// Parse each layer
	parseDOSHeader(data)
}

func parseDOSHeader(data []byte) {
	fmt.Println("\n[DOS HEADER]")

	// Ensure we have enough bytes for DOS header (64 bytes)
	if len(data) < 64 {
		fmt.Println("  ✗ File too small to be a valid PE")
		os.Exit(1)
	}

	// Read the first 2 bytes - this is e_magic ("MZ")
	magic := binary.LittleEndian.Uint16(data[0:2])

	// Validate DOS signature
	if magic != IMAGE_DOS_SIGNATURE {
		fmt.Printf("  ✗ Invalid DOS signature: 0x%04X (expected 0x5A4D)\n", magic)
		os.Exit(1)
	}

	fmt.Printf("  Magic (MZ):                    0x%04X ✓\n", magic)

	// Read e_lfanew at offset 0x3C (60 bytes in)
	// This tells us where the PE headers start
	e_lfanew := binary.LittleEndian.Uint32(data[60:64])

	fmt.Printf("  PE Header Offset (e_lfanew):   0x%04X\n", e_lfanew)

	// Verify PE signature at e_lfanew location
	if int(e_lfanew)+4 > len(data) {
		fmt.Println("  ✗ Invalid e_lfanew - points outside file")
		os.Exit(1)
	}

	peSignature := binary.LittleEndian.Uint32(data[e_lfanew : e_lfanew+4])
	if peSignature != IMAGE_NT_SIGNATURE {
		fmt.Printf("  ✗ Invalid PE signature: 0x%08X (expected 0x00004550)\n", peSignature)
		os.Exit(1)
	}

	// Continue parsing from this offset
	parseFileHeader(data, e_lfanew+4)
}

func parseFileHeader(data []byte, offset uint32) {
	fmt.Println("\n[FILE HEADER]")

	// Read the 20-byte FILE_HEADER structure
	// We're at: e_lfanew + 4 (after PE signature)

	if int(offset)+20 > len(data) {
		fmt.Println("  ✗ File too small for FILE_HEADER")
		os.Exit(1)
	}

	// Parse each field
	machine := binary.LittleEndian.Uint16(data[offset : offset+2])
	numberOfSections := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
	timeDateStamp := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
	sizeOfOptionalHeader := binary.LittleEndian.Uint16(data[offset+16 : offset+18])
	characteristics := binary.LittleEndian.Uint16(data[offset+18 : offset+20])

	// Decode architecture
	archName := "Unknown"
	switch machine {
	case IMAGE_FILE_MACHINE_I386:
		archName = "x86 (32-bit)"
	case IMAGE_FILE_MACHINE_AMD64:
		archName = "x64 (64-bit)"
	}

	fmt.Printf("  Architecture:                  %s (0x%04X)\n", archName, machine)
	fmt.Printf("  Number of Sections:            %d\n", numberOfSections)

	// Decode timestamp
	if timeDateStamp == 0 {
		fmt.Println("  Timestamp:                     0 (likely packed/manipulated)")
	} else {
		compileTime := time.Unix(int64(timeDateStamp), 0)
		fmt.Printf("  Timestamp:                     %s\n", compileTime.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("  Characteristics:               0x%04X\n", characteristics)

	// Decode characteristics flags
	if characteristics&IMAGE_FILE_EXECUTABLE_IMAGE != 0 {
		fmt.Println("    ✓ Executable Image")
	}
	if characteristics&IMAGE_FILE_DLL != 0 {
		fmt.Println("    ✓ DLL")
	} else {
		fmt.Println("    ✗ DLL")
	}
	if characteristics&IMAGE_FILE_RELOCS_STRIPPED != 0 {
		fmt.Println("    ✗ Relocations Stripped (cannot rebase!)")
	} else {
		fmt.Println("    ✓ Has Relocations (not stripped)")
	}
	if characteristics&IMAGE_FILE_LARGE_ADDRESS_AWARE != 0 {
		fmt.Println("    ✓ Large Address Aware")
	}

	// Continue to Optional Header
	optionalHeaderOffset := offset + 20
	parseOptionalHeader(data, optionalHeaderOffset, machine, numberOfSections, sizeOfOptionalHeader)
}

func parseOptionalHeader(data []byte, offset uint32, machine uint16, numSections uint16, optHeaderSize uint16) {
	fmt.Println("\n[OPTIONAL HEADER]")

	if int(offset)+int(optHeaderSize) > len(data) {
		fmt.Println("  ✗ File too small for OPTIONAL_HEADER")
		os.Exit(1)
	}

	// Read magic to determine 32-bit vs 64-bit
	magic := binary.LittleEndian.Uint16(data[offset : offset+2])

	is64Bit := magic == 0x020B

	if is64Bit {
		fmt.Println("  Magic:                         0x020B (PE32+ / 64-bit)")
	} else {
		fmt.Println("  Magic:                         0x010B (PE32 / 32-bit)")
	}

	// Parse common fields (same layout for 32 and 64 bit up to BaseOfCode)
	entryPoint := binary.LittleEndian.Uint32(data[offset+16 : offset+20])
	_ = binary.LittleEndian.Uint32(data[offset+20 : offset+24]) // baseOfCode - read but not displayed

	fmt.Printf("  Entry Point (RVA):             0x%08X\n", entryPoint)

	// ImageBase differs in offset between 32/64 bit
	var imageBase uint64
	var sectionAlignment, fileAlignment uint32
	var sizeOfImage, sizeOfHeaders uint32
	var dllCharacteristics uint16
	var dataDirectoryOffset uint32

	if is64Bit {
		// 64-bit layout
		imageBase = binary.LittleEndian.Uint64(data[offset+24 : offset+32])
		sectionAlignment = binary.LittleEndian.Uint32(data[offset+32 : offset+36])
		fileAlignment = binary.LittleEndian.Uint32(data[offset+36 : offset+40])
		sizeOfImage = binary.LittleEndian.Uint32(data[offset+56 : offset+60])
		sizeOfHeaders = binary.LittleEndian.Uint32(data[offset+60 : offset+64])
		dllCharacteristics = binary.LittleEndian.Uint16(data[offset+70 : offset+72])
		dataDirectoryOffset = offset + 112 // Data directories start here in 64-bit
	} else {
		// 32-bit layout
		imageBase = uint64(binary.LittleEndian.Uint32(data[offset+28 : offset+32]))
		sectionAlignment = binary.LittleEndian.Uint32(data[offset+32 : offset+36])
		fileAlignment = binary.LittleEndian.Uint32(data[offset+36 : offset+40])
		sizeOfImage = binary.LittleEndian.Uint32(data[offset+56 : offset+60])
		sizeOfHeaders = binary.LittleEndian.Uint32(data[offset+60 : offset+64])
		dllCharacteristics = binary.LittleEndian.Uint16(data[offset+70 : offset+72])
		dataDirectoryOffset = offset + 96 // Data directories start here in 32-bit
	}

	fmt.Printf("  Image Base:                    0x%016X\n", imageBase)
	fmt.Printf("  Section Alignment:             0x%08X (%d bytes)\n", sectionAlignment, sectionAlignment)
	fmt.Printf("  File Alignment:                0x%08X (%d bytes)\n", fileAlignment, fileAlignment)
	fmt.Printf("  Size of Image:                 0x%08X\n", sizeOfImage)
	fmt.Printf("  Size of Headers:               0x%08X\n", sizeOfHeaders)

	// Parse security features
	fmt.Println("\n[SECURITY FEATURES]")

	aslrEnabled := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE != 0
	depEnabled := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_NX_COMPAT != 0
	noSEH := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_NO_SEH != 0

	if aslrEnabled {
		fmt.Println("  ASLR (DYNAMIC_BASE):           ✓ ENABLED")
	} else {
		fmt.Println("  ASLR (DYNAMIC_BASE):           ✗ DISABLED")
	}

	if depEnabled {
		fmt.Println("  DEP (NX_COMPAT):               ✓ ENABLED")
	} else {
		fmt.Println("  DEP (NX_COMPAT):               ✗ DISABLED")
	}

	if noSEH {
		fmt.Println("  SEH (Exception Handlers):      ✗ DISABLED (NO_SEH)")
	}

	// Parse Data Directories
	parseDataDirectories(data, dataDirectoryOffset)

	// Calculate where sections start
	sectionTableOffset := offset + uint32(optHeaderSize)
	parseSections(data, sectionTableOffset, numSections, fileAlignment, sectionAlignment)
}

func parseDataDirectories(data []byte, offset uint32) {
	fmt.Println("\n[DATA DIRECTORIES]")

	// Each directory is 8 bytes (4 byte RVA + 4 byte Size)
	directories := []struct {
		index int
		name  string
	}{
		{IMAGE_DIRECTORY_ENTRY_EXPORT, "Export"},
		{IMAGE_DIRECTORY_ENTRY_IMPORT, "Import"},
		{IMAGE_DIRECTORY_ENTRY_RESOURCE, "Resource"},
		{IMAGE_DIRECTORY_ENTRY_BASERELOC, "Base Relocation"},
		{IMAGE_DIRECTORY_ENTRY_DEBUG, "Debug"},
		{IMAGE_DIRECTORY_ENTRY_TLS, "TLS"},
		{IMAGE_DIRECTORY_ENTRY_IAT, "IAT"},
	}

	for _, dir := range directories {
		dirOffset := offset + uint32(dir.index*8)

		if int(dirOffset)+8 > len(data) {
			continue
		}

		rva := binary.LittleEndian.Uint32(data[dirOffset : dirOffset+4])
		size := binary.LittleEndian.Uint32(data[dirOffset+4 : dirOffset+8])

		if rva == 0 {
			fmt.Printf("  %-20s Not present\n", dir.name+":")
		} else {
			fmt.Printf("  %-20s RVA 0x%08X, Size %d bytes\n", dir.name+":", rva, size)
		}
	}
}

func parseSections(data []byte, offset uint32, numSections uint16, fileAlign uint32, sectionAlign uint32) {
	fmt.Println("\n[SECTIONS]")

	// Each section header is 40 bytes
	for i := uint16(0); i < numSections; i++ {
		sectionOffset := offset + uint32(i)*40

		if int(sectionOffset)+40 > len(data) {
			break
		}

		// Parse section header
		var name [8]byte
		copy(name[:], data[sectionOffset:sectionOffset+8])

		virtualSize := binary.LittleEndian.Uint32(data[sectionOffset+8 : sectionOffset+12])
		virtualAddress := binary.LittleEndian.Uint32(data[sectionOffset+12 : sectionOffset+16])
		_ = binary.LittleEndian.Uint32(data[sectionOffset+16 : sectionOffset+20]) // rawSize - read but not displayed separately
		rawAddress := binary.LittleEndian.Uint32(data[sectionOffset+20 : sectionOffset+24])
		characteristics := binary.LittleEndian.Uint32(data[sectionOffset+36 : sectionOffset+40])

		// Extract name (may not be null-terminated)
		sectionName := ""
		for _, b := range name {
			if b == 0 {
				break
			}
			sectionName += string(b)
		}

		// Decode permissions
		perms := ""
		if characteristics&IMAGE_SCN_MEM_READ != 0 {
			perms += "R"
		} else {
			perms += "-"
		}
		if characteristics&IMAGE_SCN_MEM_WRITE != 0 {
			perms += "W"
		} else {
			perms += "-"
		}
		if characteristics&IMAGE_SCN_MEM_EXECUTE != 0 {
			perms += "X"
		} else {
			perms += "-"
		}

		// Display section info
		fmt.Printf("  %-8s RVA 0x%08X  Raw 0x%08X  Perms: %s  Size: %dKB\n",
			sectionName,
			virtualAddress,
			rawAddress,
			perms,
			virtualSize/1024)

		// Security warnings
		if perms == "RWX" {
			fmt.Printf("           ⚠️  WARNING: RWX section (self-modifying code possible)\n")
		}

		// Check for custom section names
		standardSections := []string{".text", ".rdata", ".data", ".rsrc", ".reloc", ".pdata"}
		isStandard := false
		for _, std := range standardSections {
			if sectionName == std {
				isStandard = true
				break
			}
		}
		if !isStandard && sectionName != "" {
			fmt.Printf("           ℹ️  Custom section name (possible packer)\n")
		}
	}

	// Now parse imports using section table for RVA conversion
	parseImports(data, offset, numSections)
}

func parseImports(data []byte, sectionTableOffset uint32, numSections uint16) {
	fmt.Println("\n[IMPORT ANALYSIS]")

	// First, we need to find the Import Directory RVA
	// Go back and read it from data directories
	// For simplicity, we'll recalculate the offset

	// Read DOS header e_lfanew
	e_lfanew := binary.LittleEndian.Uint32(data[60:64])

	// Calculate optional header offset
	optHeaderOffset := e_lfanew + 4 + 20 // PE sig + file header

	// Read magic to know 32 vs 64 bit
	magic := binary.LittleEndian.Uint16(data[optHeaderOffset : optHeaderOffset+2])
	is64Bit := magic == 0x020B

	// Calculate data directory offset
	var dataDirectoryOffset uint32
	if is64Bit {
		dataDirectoryOffset = optHeaderOffset + 112
	} else {
		dataDirectoryOffset = optHeaderOffset + 96
	}

	// Read Import Directory entry (index 1)
	importDirOffset := dataDirectoryOffset + uint32(IMAGE_DIRECTORY_ENTRY_IMPORT*8)
	importRVA := binary.LittleEndian.Uint32(data[importDirOffset : importDirOffset+4])
	_ = binary.LittleEndian.Uint32(data[importDirOffset+4 : importDirOffset+8]) // importSize - read but not used

	if importRVA == 0 {
		fmt.Println("  No imports (statically linked or suspicious)")
		return
	}

	// Convert Import Directory RVA to file offset
	importFileOffset := rvaToFileOffset(data, sectionTableOffset, numSections, importRVA)
	if importFileOffset == 0 {
		fmt.Println("  Could not locate import directory in file")
		return
	}

	// Parse import descriptors
	dllCount := 0
	suspiciousFunctions := make(map[string][]string)

	offset := importFileOffset
	for {
		// Each import descriptor is 20 bytes
		if int(offset)+20 > len(data) {
			break
		}

		nameRVA := binary.LittleEndian.Uint32(data[offset+12 : offset+16])

		// Null descriptor marks end
		if nameRVA == 0 {
			break
		}

		// Convert DLL name RVA to file offset
		nameOffset := rvaToFileOffset(data, sectionTableOffset, numSections, nameRVA)
		if nameOffset == 0 {
			break
		}

		// Read DLL name
		dllName := readNullTerminatedString(data[nameOffset:])
		dllCount++

		// Parse function imports from this DLL
		firstThunkRVA := binary.LittleEndian.Uint32(data[offset : offset+4]) // OriginalFirstThunk
		if firstThunkRVA == 0 {
			firstThunkRVA = binary.LittleEndian.Uint32(data[offset+16 : offset+20]) // FirstThunk
		}

		thunkOffset := rvaToFileOffset(data, sectionTableOffset, numSections, firstThunkRVA)
		if thunkOffset != 0 {
			suspicious := parseFunctionImports(data, sectionTableOffset, numSections, thunkOffset, is64Bit)
			if len(suspicious) > 0 {
				suspiciousFunctions[dllName] = suspicious
			}
		}

		offset += 20
	}

	fmt.Printf("  Imported DLLs: %d\n", dllCount)

	if len(suspiciousFunctions) > 0 {
		fmt.Println("\n  ⚠️  Suspicious imports detected:")
		for dll, funcs := range suspiciousFunctions {
			fmt.Printf("    %s:\n", dll)
			for _, fn := range funcs {
				fmt.Printf("      - %s\n", fn)
			}
		}
	}
}

// Convert RVA to file offset using section table
func rvaToFileOffset(data []byte, sectionTableOffset uint32, numSections uint16, rva uint32) uint32 {
	// Iterate through sections to find which contains this RVA
	for i := uint16(0); i < numSections; i++ {
		sectionOffset := sectionTableOffset + uint32(i)*40

		if int(sectionOffset)+40 > len(data) {
			break
		}

		virtualAddress := binary.LittleEndian.Uint32(data[sectionOffset+12 : sectionOffset+16])
		virtualSize := binary.LittleEndian.Uint32(data[sectionOffset+8 : sectionOffset+12])
		rawAddress := binary.LittleEndian.Uint32(data[sectionOffset+20 : sectionOffset+24])

		// Check if RVA falls within this section
		if rva >= virtualAddress && rva < virtualAddress+virtualSize {
			// Calculate offset within section
			offsetInSection := rva - virtualAddress
			// Return file offset
			return rawAddress + offsetInSection
		}
	}

	return 0 // RVA not found in any section
}

func parseFunctionImports(data []byte, sectionTableOffset uint32, numSections uint16, thunkOffset uint32, is64Bit bool) []string {
	var suspicious []string

	// Suspicious API list
	suspiciousAPIs := map[string]bool{
		"VirtualAlloc": true, "VirtualAllocEx": true, "VirtualProtect": true,
		"WriteProcessMemory": true, "ReadProcessMemory": true,
		"CreateRemoteThread": true, "NtCreateThreadEx": true,
		"OpenProcess": true, "LoadLibrary": true, "LoadLibraryA": true,
		"LoadLibraryW": true, "GetProcAddress": true,
		"WinExec": true, "ShellExecute": true, "CreateProcess": true,
	}

	thunkSize := uint32(4)
	if is64Bit {
		thunkSize = 8
	}

	offset := thunkOffset
	for {
		if int(offset)+int(thunkSize) > len(data) {
			break
		}

		var thunkData uint64
		if is64Bit {
			thunkData = binary.LittleEndian.Uint64(data[offset : offset+8])
		} else {
			thunkData = uint64(binary.LittleEndian.Uint32(data[offset : offset+4]))
		}

		// Null thunk marks end
		if thunkData == 0 {
			break
		}

		// Check if import by ordinal (high bit set)
		ordinalFlag := uint64(0x8000000000000000)
		if !is64Bit {
			ordinalFlag = 0x80000000
		}

		if thunkData&ordinalFlag == 0 {
			// Import by name
			nameRVA := uint32(thunkData & 0x7FFFFFFF)
			nameOffset := rvaToFileOffset(data, sectionTableOffset, numSections, nameRVA)

			if nameOffset != 0 && int(nameOffset)+2 < len(data) {
				// Skip hint (2 bytes), read function name
				funcName := readNullTerminatedString(data[nameOffset+2:])

				if suspiciousAPIs[funcName] {
					suspicious = append(suspicious, funcName)
				}
			}
		}

		offset += thunkSize
	}

	return suspicious
}

func readNullTerminatedString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

```
____



## Code Explanation


### Part 1: Build Configuration and Imports

#### Build Tags

```go
//go:build windows
// +build windows
```

These build tags ensure the code only compiles on Windows. The `//go:build` syntax is the modern format (Go 1.17+), while `// +build` provides backwards compatibility. This is important because our PE parser uses Windows-specific concepts and is intended for analyzing Windows executables.

#### Required Imports

```go
import (
    "encoding/binary"
    "fmt"
    "os"
    "time"
)
```

- **encoding/binary**: Converts between byte slices and Go types. PE files store multi-byte values in little-endian format, which this package handles natively.
- **os**: Provides file I/O operations. We read the entire PE file into memory as a byte slice for parsing.
- **time**: Converts the Unix timestamp in the PE header to a human-readable compilation date.

### Part 2: PE Structure Definitions

Before parsing, we define Go structures that mirror the PE format. These aren't used directly for binary reading (we use manual offset-based parsing for educational clarity), but they document the expected layout.

#### IMAGE_DOS_HEADER (64 bytes)

Every PE file begins with this legacy DOS header, a relic from the era when executables needed to run on both DOS and Windows:

```go
type IMAGE_DOS_HEADER struct {
    E_magic    uint16     // Magic number "MZ" (0x5A4D)
    E_cblp     uint16     // Bytes on last page of file
    E_cp       uint16     // Pages in file
    E_crlc     uint16     // Relocations
    E_cparhdr  uint16     // Size of header in paragraphs
    E_minalloc uint16     // Minimum extra paragraphs needed
    E_maxalloc uint16     // Maximum extra paragraphs needed
    E_ss       uint16     // Initial (relative) SS value
    E_sp       uint16     // Initial SP value
    E_csum     uint16     // Checksum
    E_ip       uint16     // Initial IP value
    E_cs       uint16     // Initial (relative) CS value
    E_lfarlc   uint16     // File address of relocation table
    E_ovno     uint16     // Overlay number
    E_res      [4]uint16  // Reserved words
    E_oemid    uint16     // OEM identifier
    E_oeminfo  uint16     // OEM information
    E_res2     [10]uint16 // Reserved words
    E_lfanew   int32      // File address of new exe header
}
```

**Key Insight:** Only two fields matter for modern PE parsing: `E_magic` (validates this is a DOS/PE file) and `E_lfanew` (pointer to the actual PE headers). The rest are DOS-era artifacts.

#### IMAGE_FILE_HEADER (20 bytes)

The "COFF" header containing core PE metadata:

```go
type IMAGE_FILE_HEADER struct {
    Machine              uint16 // Architecture (0x8664=x64, 0x014C=x86)
    NumberOfSections     uint16 // Count of sections
    TimeDateStamp        uint32 // Unix timestamp of compilation
    PointerToSymbolTable uint32 // Deprecated
    NumberOfSymbols      uint32 // Deprecated
    SizeOfOptionalHeader uint16 // Size of optional header
    Characteristics      uint16 // File flags (DLL, relocations)
}
```

**Security Relevance:** The `Machine` field determines shellcode compatibility. The `Characteristics` bitfield reveals if relocations are stripped (breaking ASLR) or if it's a DLL.

#### IMAGE_OPTIONAL_HEADER64 (240 bytes for 64-bit)

Despite its name, this header is required. It contains critical execution parameters:

```go
type IMAGE_OPTIONAL_HEADER64 struct {
    Magic                       uint16 // 0x020B for PE32+ (64-bit)
    MajorLinkerVersion          uint8
    MinorLinkerVersion          uint8
    SizeOfCode                  uint32
    SizeOfInitializedData       uint32
    SizeOfUninitializedData     uint32
    AddressOfEntryPoint         uint32 // RVA where execution begins
    BaseOfCode                  uint32 // RVA of code section start
    ImageBase                   uint64 // Preferred load address
    SectionAlignment            uint32 // Section alignment in memory
    FileAlignment               uint32 // Section alignment on disk
    // ... version fields ...
    SizeOfImage                 uint32 // Total size in memory
    SizeOfHeaders               uint32 // Size of all headers
    CheckSum                    uint32
    Subsystem                   uint16
    DllCharacteristics          uint16 // Security flags (ASLR, DEP)
    // ... stack/heap sizes ...
    NumberOfRvaAndSizes         uint32
    DataDirectory               [16]IMAGE_DATA_DIRECTORY
}
```

**Critical Fields:** `AddressOfEntryPoint` is where execution begins (often targeted by packers). `ImageBase` is the preferred load address. `DllCharacteristics` contains security flags like ASLR and DEP.

#### IMAGE_SECTION_HEADER (40 bytes each)

Each section has a header describing its location and permissions:

```go
type IMAGE_SECTION_HEADER struct {
    Name                 [8]byte // Section name (not null-terminated!)
    VirtualSize          uint32  // Size in memory
    VirtualAddress       uint32  // RVA where section loads
    SizeOfRawData        uint32  // Size on disk
    PointerToRawData     uint32  // File offset of section
    PointerToRelocations uint32  // Obsolete
    PointerToLinenumbers uint32  // Obsolete
    NumberOfRelocations  uint16  // Obsolete
    NumberOfLinenumbers  uint16  // Obsolete
    Characteristics      uint32  // Section flags (R/W/X)
}
```

⚠️ **Malware Indicator:** Sections with RWX (Read-Write-Execute) permissions are highly suspicious as they allow self-modifying code. Legitimate software rarely needs this.

### Part 3: PE Constants

These constants define magic values and flags used throughout PE parsing:

```go
const (
    IMAGE_DOS_SIGNATURE = 0x5A4D     // "MZ"
    IMAGE_NT_SIGNATURE  = 0x00004550 // "PE\0\0"
 
    IMAGE_FILE_MACHINE_I386  = 0x014C
    IMAGE_FILE_MACHINE_AMD64 = 0x8664
 
    IMAGE_FILE_DLL                 = 0x2000
    IMAGE_FILE_EXECUTABLE_IMAGE    = 0x0002
    IMAGE_FILE_RELOCS_STRIPPED     = 0x0001
    IMAGE_FILE_LARGE_ADDRESS_AWARE = 0x0020
 
    IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE = 0x0040 // ASLR
    IMAGE_DLLCHARACTERISTICS_NX_COMPAT    = 0x0100 // DEP
    IMAGE_DLLCHARACTERISTICS_NO_SEH       = 0x0400
 
    IMAGE_SCN_MEM_EXECUTE = 0x20000000
    IMAGE_SCN_MEM_READ    = 0x40000000
    IMAGE_SCN_MEM_WRITE   = 0x80000000
 
    // Data Directory indices
    IMAGE_DIRECTORY_ENTRY_EXPORT    = 0
    IMAGE_DIRECTORY_ENTRY_IMPORT    = 1
    IMAGE_DIRECTORY_ENTRY_RESOURCE  = 2
    IMAGE_DIRECTORY_ENTRY_BASERELOC = 5
    IMAGE_DIRECTORY_ENTRY_DEBUG     = 6
    IMAGE_DIRECTORY_ENTRY_TLS       = 9
    IMAGE_DIRECTORY_ENTRY_IAT       = 12
)
```

**Understanding the Constants:** The "MZ" signature (0x5A4D) comes from Mark Zbikowski, a DOS architect. Machine constants identify CPU architecture. `DllCharacteristics` flags control modern security features like ASLR (Address Space Layout Randomization) and DEP (Data Execution Prevention).

### Part 4: Main Entry Point

```go
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: pe-parser <executable>")
        os.Exit(1)
    }
 
    filename := os.Args[1]
 
    // Read entire PE file into memory
    data, err := os.ReadFile(filename)
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        os.Exit(1)
    }
 
    fmt.Printf("\nPE Analysis Report: %s\n", filename)
    fmt.Println("=" + repeat("=", 79))
 
    // Parse each layer
    parseDOSHeader(data)
}
```

**Design Decision:** We read the entire file into a byte slice rather than using streaming I/O. This simplifies offset calculations and allows random access to any part of the file. For most PE files (typically < 100MB), this is efficient.

### Part 5: Parsing the DOS Header

The DOS header is our entry point into the PE format:

```go
func parseDOSHeader(data []byte) {
    fmt.Println("\n[DOS HEADER]")
 
    // Ensure we have enough bytes for DOS header (64 bytes)
    if len(data) < 64 {
        fmt.Println("  ✗ File too small to be a valid PE")
        os.Exit(1)
    }
 
    // Read the first 2 bytes - this is e_magic ("MZ")
    magic := binary.LittleEndian.Uint16(data[0:2])
 
    // Validate DOS signature
    if magic != IMAGE_DOS_SIGNATURE {
        fmt.Printf("  ✗ Invalid DOS signature: 0x%04X\n", magic)
        os.Exit(1)
    }
 
    fmt.Printf("  Magic (MZ): 0x%04X ✓\n", magic)
 
    // Read e_lfanew at offset 0x3C (60 bytes in)
    e_lfanew := binary.LittleEndian.Uint32(data[60:64])
    fmt.Printf("  PE Header Offset (e_lfanew): 0x%04X\n", e_lfanew)
 
    // Verify PE signature at e_lfanew location
    if int(e_lfanew)+4 > len(data) {
        fmt.Println("  ✗ Invalid e_lfanew - points outside file")
        os.Exit(1)
    }
 
    peSignature := binary.LittleEndian.Uint32(data[e_lfanew : e_lfanew+4])
    if peSignature != IMAGE_NT_SIGNATURE {
        fmt.Printf("  ✗ Invalid PE signature: 0x%08X\n", peSignature)
        os.Exit(1)
    }
 
    // Continue parsing from this offset
    parseFileHeader(data, e_lfanew+4)
}
```

**Conceptual Connection:** The DOS header acts as a "bridge" to the PE headers. The `e_lfanew` field at offset 0x3C is crucial - it's a 4-byte pointer to where the actual PE signature ("PE\0\0") begins. Without validating both signatures, we could be fooled by non-PE files.

### Part 6: Parsing the File Header

The File Header (COFF header) contains essential metadata about the executable:

```go
func parseFileHeader(data []byte, offset uint32) {
    fmt.Println("\n[FILE HEADER]")
 
    if int(offset)+20 > len(data) {
        fmt.Println("  ✗ File too small for FILE_HEADER")
        os.Exit(1)
    }
 
    // Parse each field at specific offsets
    machine := binary.LittleEndian.Uint16(data[offset : offset+2])
    numberOfSections := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
    timeDateStamp := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
    sizeOfOptionalHeader := binary.LittleEndian.Uint16(data[offset+16 : offset+18])
    characteristics := binary.LittleEndian.Uint16(data[offset+18 : offset+20])
 
    // Decode architecture
    archName := "Unknown"
    switch machine {
    case IMAGE_FILE_MACHINE_I386:
        archName = "x86 (32-bit)"
    case IMAGE_FILE_MACHINE_AMD64:
        archName = "x64 (64-bit)"
    }
 
    fmt.Printf("  Architecture: %s (0x%04X)\n", archName, machine)
    fmt.Printf("  Number of Sections: %d\n", numberOfSections)
 
    // Decode timestamp
    if timeDateStamp == 0 {
        fmt.Println("  Timestamp: 0 (likely packed/manipulated)")
    } else {
        compileTime := time.Unix(int64(timeDateStamp), 0)
        fmt.Printf("  Timestamp: %s\n", 
            compileTime.Format("2006-01-02 15:04:05"))
    }
 
    // Decode characteristics flags
    fmt.Printf("  Characteristics: 0x%04X\n", characteristics)
    if characteristics&IMAGE_FILE_EXECUTABLE_IMAGE != 0 {
        fmt.Println("    ✓ Executable Image")
    }
    if characteristics&IMAGE_FILE_DLL != 0 {
        fmt.Println("    ✓ DLL")
    }
    if characteristics&IMAGE_FILE_RELOCS_STRIPPED != 0 {
        fmt.Println("    ✗ Relocations Stripped (cannot rebase!)")
    }
 
    // Continue to Optional Header
    optionalHeaderOffset := offset + 20
    parseOptionalHeader(data, optionalHeaderOffset, machine,
                        numberOfSections, sizeOfOptionalHeader)
}
```

**Conceptual Connection:** The File Header immediately follows the PE signature. The `Machine` field tells us x86 vs x64 (critical for shellcode compatibility). `Characteristics` is a bitfield - we decode it bit by bit using AND operations. The `RELOCS_STRIPPED` flag determines if process hollowing is viable.

⚠️ **Timestamp Analysis:** A zeroed timestamp often indicates the file was packed or deliberately manipulated. Malware authors frequently zero this to hide compilation dates. Future timestamps or timestamps from before Windows existed are also red flags.

### Part 7: Parsing the Optional Header

The "Optional" Header (required despite its name) contains critical execution parameters:

```go
func parseOptionalHeader(data []byte, offset uint32, machine uint16,
                         numSections uint16, optHeaderSize uint16) {
    fmt.Println("\n[OPTIONAL HEADER]")
 
    if int(offset)+int(optHeaderSize) > len(data) {
        fmt.Println("  ✗ File too small for OPTIONAL_HEADER")
        os.Exit(1)
    }
 
    // Read magic to determine 32-bit vs 64-bit
    magic := binary.LittleEndian.Uint16(data[offset : offset+2])
    is64Bit := magic == 0x020B
 
    if is64Bit {
        fmt.Println("  Magic: 0x020B (PE32+ / 64-bit)")
    } else {
        fmt.Println("  Magic: 0x010B (PE32 / 32-bit)")
    }
 
    // Parse common fields
    entryPoint := binary.LittleEndian.Uint32(data[offset+16 : offset+20])
    fmt.Printf("  Entry Point (RVA): 0x%08X\n", entryPoint)
 
    // Field offsets differ between 32/64 bit
    var imageBase uint64
    var sectionAlignment, fileAlignment uint32
    var sizeOfImage, sizeOfHeaders uint32
    var dllCharacteristics uint16
    var dataDirectoryOffset uint32
 
    if is64Bit {
        imageBase = binary.LittleEndian.Uint64(data[offset+24 : offset+32])
        sectionAlignment = binary.LittleEndian.Uint32(data[offset+32 : offset+36])
        fileAlignment = binary.LittleEndian.Uint32(data[offset+36 : offset+40])
        sizeOfImage = binary.LittleEndian.Uint32(data[offset+56 : offset+60])
        sizeOfHeaders = binary.LittleEndian.Uint32(data[offset+60 : offset+64])
        dllCharacteristics = binary.LittleEndian.Uint16(data[offset+70 : offset+72])
        dataDirectoryOffset = offset + 112
    } else {
        imageBase = uint64(binary.LittleEndian.Uint32(data[offset+28 : offset+32]))
        // ... similar for 32-bit with adjusted offsets
        dataDirectoryOffset = offset + 96
    }
 
    fmt.Printf("  Image Base: 0x%016X\n", imageBase)
    fmt.Printf("  Section Alignment: 0x%08X (%d bytes)\n",
              sectionAlignment, sectionAlignment)
    fmt.Printf("  File Alignment: 0x%08X (%d bytes)\n",
              fileAlignment, fileAlignment)
}
```

**32-bit vs 64-bit Differences:** The Optional Header has different sizes and field offsets for PE32 (32-bit) vs PE32+ (64-bit). The `Magic` field (0x010B vs 0x020B) tells us which format we're dealing with. Key differences: `ImageBase` is 4 bytes in PE32 but 8 bytes in PE32+, shifting all subsequent field offsets.

### Part 8: Security Feature Analysis

The `DllCharacteristics` field contains critical security flags:

```go
// Parse security features from DllCharacteristics
fmt.Println("\n[SECURITY FEATURES]")
 
aslrEnabled := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_DYNAMIC_BASE != 0
depEnabled := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_NX_COMPAT != 0
noSEH := dllCharacteristics&IMAGE_DLLCHARACTERISTICS_NO_SEH != 0
 
if aslrEnabled {
    fmt.Println("  ASLR (DYNAMIC_BASE): ✓ ENABLED")
} else {
    fmt.Println("  ASLR (DYNAMIC_BASE): ✗ DISABLED")
}
 
if depEnabled {
    fmt.Println("  DEP (NX_COMPAT): ✓ ENABLED")
} else {
    fmt.Println("  DEP (NX_COMPAT): ✗ DISABLED")
}
 
if noSEH {
    fmt.Println("  SEH: ✗ DISABLED (NO_SEH)")
}
```

**Security Flags Explained:**

- **ASLR (DYNAMIC_BASE)**: Randomizes the load address each execution. Without ASLR, attackers can predict memory addresses for exploits.
- **DEP (NX_COMPAT)**: Marks data regions as non-executable. Prevents classic buffer overflow exploits that execute shellcode on the stack.
- **NO_SEH**: Indicates no Structured Exception Handlers. SEH overwrite attacks are a classic exploit technique.

⚠️ **Missing Security Flags:** Modern legitimate software should have both ASLR and DEP enabled. Missing security flags could indicate:
1. Legacy software compiled with old tools
2. Deliberately weakened security for exploitation
3. Malware designed to be easier to exploit or inject into

### Part 9: Data Directories

The Data Directory array is a "table of contents" pointing to specialized structures:

```go
func parseDataDirectories(data []byte, offset uint32) {
    fmt.Println("\n[DATA DIRECTORIES]")
 
    // Each directory is 8 bytes (4 byte RVA + 4 byte Size)
    directories := []struct {
        index int
        name  string
    }{
        {IMAGE_DIRECTORY_ENTRY_EXPORT, "Export"},
        {IMAGE_DIRECTORY_ENTRY_IMPORT, "Import"},
        {IMAGE_DIRECTORY_ENTRY_RESOURCE, "Resource"},
        {IMAGE_DIRECTORY_ENTRY_BASERELOC, "Base Relocation"},
        {IMAGE_DIRECTORY_ENTRY_DEBUG, "Debug"},
        {IMAGE_DIRECTORY_ENTRY_TLS, "TLS"},
        {IMAGE_DIRECTORY_ENTRY_IAT, "IAT"},
    }
 
    for _, dir := range directories {
        dirOffset := offset + uint32(dir.index*8)
 
        if int(dirOffset)+8 > len(data) {
            continue
        }
 
        rva := binary.LittleEndian.Uint32(data[dirOffset : dirOffset+4])
        size := binary.LittleEndian.Uint32(data[dirOffset+4 : dirOffset+8])
 
        if rva == 0 {
            fmt.Printf("  %-20s Not present\n", dir.name+":")
        } else {
            fmt.Printf("  %-20s RVA 0x%08X, Size %d bytes\n",
                      dir.name+":", rva, size)
        }
    }
}
```

**Key Data Directories for Analysis:**

- **Import Directory**: Lists DLLs and functions the executable needs. Reveals capabilities and suspicious API usage.
- **Export Directory**: Functions exposed by DLLs. Useful for understanding DLL capabilities.
- **Base Relocation**: Required for ASLR. If missing, the executable cannot be rebased.
- **TLS Directory**: Thread Local Storage callbacks execute before main(). Malware uses these for anti-debugging.
- **Debug Directory**: May contain PDB paths revealing original build environment and developer names.

### Part 10: Section Table Parsing

Sections divide the executable into logical regions with different purposes and permissions:

```go
func parseSections(data []byte, offset uint32, numSections uint16,
                   fileAlign uint32, sectionAlign uint32) {
    fmt.Println("\n[SECTIONS]")
 
    // Each section header is 40 bytes
    for i := uint16(0); i < numSections; i++ {
        sectionOffset := offset + uint32(i)*40
 
        if int(sectionOffset)+40 > len(data) {
            break
        }
 
        // Parse section header
        var name [8]byte
        copy(name[:], data[sectionOffset:sectionOffset+8])
 
        virtualSize := binary.LittleEndian.Uint32(
            data[sectionOffset+8 : sectionOffset+12])
        virtualAddress := binary.LittleEndian.Uint32(
            data[sectionOffset+12 : sectionOffset+16])
        rawAddress := binary.LittleEndian.Uint32(
            data[sectionOffset+20 : sectionOffset+24])
        characteristics := binary.LittleEndian.Uint32(
            data[sectionOffset+36 : sectionOffset+40])
 
        // Extract name (may not be null-terminated)
        sectionName := ""
        for _, b := range name {
            if b == 0 { break }
            sectionName += string(b)
        }
 
        // Decode permissions
        perms := ""
        if characteristics&IMAGE_SCN_MEM_READ != 0 { perms += "R" }
        else { perms += "-" }
        if characteristics&IMAGE_SCN_MEM_WRITE != 0 { perms += "W" }
        else { perms += "-" }
        if characteristics&IMAGE_SCN_MEM_EXECUTE != 0 { perms += "X" }
        else { perms += "-" }
 
        fmt.Printf("  %-8s RVA 0x%08X  Raw 0x%08X  Perms: %s\n",
                  sectionName, virtualAddress, rawAddress, perms)
 
        // Security warnings
        if perms == "RWX" {
            fmt.Println("    ⚠️ WARNING: RWX section")
        }
    }
}
```

**Common Section Names:**

- `.text`: Executable code. Should be R-X (read + execute, no write).
- `.rdata`: Read-only data (constants, import tables). Should be R-- only.
- `.data`: Initialized global variables. RW- (read + write, no execute).
- `.rsrc`: Resources (icons, dialogs, version info). R-- only.
- `.reloc`: Relocation information for ASLR. R-- only.

⚠️ **Suspicious Section Names:** Names like `UPX0`, `.aspack`, `.themida` indicate packers. Random-looking names (e.g., `.xyz123`) suggest custom packers or obfuscation. Section names are stored in an 8-byte array that's NOT guaranteed to be null-terminated!

### Part 11: RVA to File Offset Conversion

Understanding RVA (Relative Virtual Address) conversion is crucial for PE analysis:

```go
// Convert RVA to file offset using section table
func rvaToFileOffset(data []byte, sectionTableOffset uint32,
                     numSections uint16, rva uint32) uint32 {
    // Iterate through sections to find which contains this RVA
    for i := uint16(0); i < numSections; i++ {
        sectionOffset := sectionTableOffset + uint32(i)*40
 
        if int(sectionOffset)+40 > len(data) {
            break
        }
 
        virtualAddress := binary.LittleEndian.Uint32(
            data[sectionOffset+12 : sectionOffset+16])
        virtualSize := binary.LittleEndian.Uint32(
            data[sectionOffset+8 : sectionOffset+12])
        rawAddress := binary.LittleEndian.Uint32(
            data[sectionOffset+20 : sectionOffset+24])
 
        // Check if RVA falls within this section
        if rva >= virtualAddress && rva < virtualAddress+virtualSize {
            // Calculate offset within section
            offsetInSection := rva - virtualAddress
            // Return file offset
            return rawAddress + offsetInSection
        }
    }
 
    return 0 // RVA not found in any section
}
```

**RVA Conversion Explained:** When Windows loads a PE file, sections are mapped to memory at different addresses than their file offsets. An RVA is an offset from the image base in memory. To find data in the file, we:

1. Find which section contains the RVA by checking `VirtualAddress` ranges
2. Calculate the offset within that section (`RVA - VirtualAddress`)
3. Add this to `PointerToRawData` to get the file offset

**RVA Conversion Formula:**

```
FileOffset = PointerToRawData + (RVA - VirtualAddress)
```

### Part 12: Import Analysis

Imports reveal what capabilities an executable has and can identify suspicious behavior:

```go
func parseImports(data []byte, sectionTableOffset uint32,
                  numSections uint16) {
    fmt.Println("\n[IMPORT ANALYSIS]")
 
    // Locate Import Directory from data directories
    e_lfanew := binary.LittleEndian.Uint32(data[60:64])
    optHeaderOffset := e_lfanew + 4 + 20
    magic := binary.LittleEndian.Uint16(data[optHeaderOffset : optHeaderOffset+2])
    is64Bit := magic == 0x020B
 
    var dataDirectoryOffset uint32
    if is64Bit {
        dataDirectoryOffset = optHeaderOffset + 112
    } else {
        dataDirectoryOffset = optHeaderOffset + 96
    }
 
    // Read Import Directory entry (index 1)
    importDirOffset := dataDirectoryOffset + 
                       uint32(IMAGE_DIRECTORY_ENTRY_IMPORT*8)
    importRVA := binary.LittleEndian.Uint32(
                   data[importDirOffset : importDirOffset+4])
 
    if importRVA == 0 {
        fmt.Println("  No imports (statically linked or suspicious)")
        return
    }
 
    // Convert RVA to file offset
    importFileOffset := rvaToFileOffset(data, sectionTableOffset,
                                        numSections, importRVA)
 
    // Parse import descriptors
    offset := importFileOffset
    for {
        // Each import descriptor is 20 bytes
        nameRVA := binary.LittleEndian.Uint32(
                     data[offset+12 : offset+16])
 
        // Null descriptor marks end
        if nameRVA == 0 { break }
 
        // Convert DLL name RVA to file offset
        nameOffset := rvaToFileOffset(data, sectionTableOffset,
                                      numSections, nameRVA)
        dllName := readNullTerminatedString(data[nameOffset:])
 
        // Parse function imports from this DLL
        // ... (see full implementation)
 
        offset += 20
    }
}
```

**Suspicious API Categories:**

- **Memory Manipulation**: `VirtualAlloc`, `VirtualProtect`, `WriteProcessMemory` - used for code injection
- **Process Manipulation**: `OpenProcess`, `CreateRemoteThread`, `NtCreateThreadEx` - process hollowing, thread injection
- **Dynamic Loading**: `LoadLibrary`, `GetProcAddress` - evading static import analysis
- **Execution**: `WinExec`, `ShellExecute`, `CreateProcess` - spawning child processes

**Import Descriptor Structure (20 bytes):**

```go
type IMAGE_IMPORT_DESCRIPTOR struct {
    OriginalFirstThunk uint32 // RVA to Import Name Table (INT)
    TimeDateStamp      uint32 // 0 unless bound
    ForwarderChain     uint32 // For forwarded imports
    Name               uint32 // RVA to DLL name string
    FirstThunk         uint32 // RVA to Import Address Table (IAT)
}
```

**INT vs IAT:** The Import Name Table (`OriginalFirstThunk`) is a pristine copy of function references. The Import Address Table (`FirstThunk`) is overwritten by the loader with actual function addresses. Both initially contain the same RVAs pointing to function names.

### Part 13: Function Import Parsing

Each DLL's imports are stored as an array of thunks pointing to function names:

```go
func parseFunctionImports(data []byte, sectionTableOffset uint32,
                          numSections uint16, thunkOffset uint32,
                          is64Bit bool) []string {
    var suspicious []string
 
    // Suspicious API list
    suspiciousAPIs := map[string]bool{
        "VirtualAlloc": true, "VirtualAllocEx": true,
        "VirtualProtect": true, "WriteProcessMemory": true,
        "ReadProcessMemory": true, "CreateRemoteThread": true,
        "OpenProcess": true, "LoadLibraryA": true,
        "GetProcAddress": true, "WinExec": true,
    }
 
    thunkSize := uint32(4)
    if is64Bit { thunkSize = 8 }
 
    offset := thunkOffset
    for {
        var thunkData uint64
        if is64Bit {
            thunkData = binary.LittleEndian.Uint64(
                          data[offset : offset+8])
        } else {
            thunkData = uint64(binary.LittleEndian.Uint32(
                          data[offset : offset+4]))
        }
 
        // Null thunk marks end
        if thunkData == 0 { break }
 
        // Check if import by ordinal (high bit set)
        ordinalFlag := uint64(0x8000000000000000)
        if !is64Bit { ordinalFlag = 0x80000000 }
 
        if thunkData&ordinalFlag == 0 {
            // Import by name
            nameRVA := uint32(thunkData & 0x7FFFFFFF)
            nameOffset := rvaToFileOffset(data, sectionTableOffset,
                                          numSections, nameRVA)
            if nameOffset != 0 {
                // Skip hint (2 bytes), read function name
                funcName := readNullTerminatedString(
                              data[nameOffset+2:])
                if suspiciousAPIs[funcName] {
                    suspicious = append(suspicious, funcName)
                }
            }
        }
 
        offset += thunkSize
    }
 
    return suspicious
}
```

**Import by Ordinal vs Name:** Functions can be imported by name (a string) or by ordinal (a number). The high bit of the thunk indicates which: if set, the lower bits contain an ordinal number; if clear, the value is an RVA to an `IMAGE_IMPORT_BY_NAME` structure containing a 2-byte hint followed by the function name string.

⚠️ **No Imports = Suspicious:** An executable with no imports is highly unusual. It typically indicates:
1. Statically linked code (rare on Windows)
2. A packer that resolves imports dynamically at runtime
3. Shellcode or malicious payloads designed to minimize static analysis footprint

### Summary: PE Analysis Checklist

When analyzing a PE file for potential malicious behavior, examine these key indicators:

| Category | Red Flags |
|----------|-----------|
| **Header** | Zero or future timestamp, unusual machine type |
| **Security** | Missing ASLR/DEP, NO_SEH flag |
| **Sections** | RWX permissions, unusual names (UPX0, random strings), entropy >7.0 |
| **Imports** | No imports, suspicious APIs (VirtualAlloc, CreateRemoteThread), dynamic loading |
| **Exports** | Unusual exports for file type |
| **Resources** | Embedded executables, large resources with high entropy |
| **Entry Point** | Entry point in non-.text section, unusual entry point |

**Remember:** No single indicator definitively identifies malware. Legitimate software may trigger some flags (e.g., packers for copy protection). Always consider the full context and combine static analysis with dynamic analysis for thorough investigation.




---

## Build and Test

### Complete the Parser

Build the project (if you are on Windows):

```bash
go build -o pe-parser.exe
```

If you are building on Mac or Linux you'll have to include build tags to the top of your file

```go
//go:build windows  
// +build windows
```


And use this command to build

```
 GOOS=windows GOARCH=amd64 go build -o pe-parser.exe
```



### Run Against Your Test Executable

```bash
./pe-parser.exe beacon.exe
```





### Compare with Results From Previous Lab

Open your analysis table from the previous table and compare:

| Value | Lab 2 (PEBear) | Lab 3 (Your Parser) | Match? |
|-------|----------------|---------------------|--------|
| DOS Magic | ________ | ________ | Y/N |
| e_lfanew | ________ | ________ | Y/N |
| Machine | ________ | ________ | Y/N |
| Entry Point RVA | ________ | ________ | Y/N |
| ImageBase | ________ | ________ | Y/N |
| ASLR Enabled | ________ | ________ | Y/N |
| Import Directory RVA | ________ | ________ | Y/N |
| Section Count | ________ | ________ | Y/N |

**If all match: ✓ Your parser works correctly!**


___

## For Reference - Sliver beacon output

```bash
 .\pe-parser.exe .\beacon.exe

PE Analysis Report: .\beacon.exe
================================================================================

[DOS HEADER]
  Magic (MZ):                    0x5A4D ✓
  PE Header Offset (e_lfanew):   0x0080

[FILE HEADER]
  Architecture:                  x64 (64-bit) (0x8664)
  Number of Sections:            6
  Timestamp:                     0 (likely packed/manipulated)
  Characteristics:               0x0222
    ✓ Executable Image
    ✗ DLL
    ✓ Has Relocations (not stripped)
    ✓ Large Address Aware

[OPTIONAL HEADER]
  Magic:                         0x020B (PE32+ / 64-bit)
  Entry Point (RVA):             0x0005D0E0
  Image Base:                    0x0000000000400000
  Section Alignment:             0x00001000 (4096 bytes)
  File Alignment:                0x00000200 (512 bytes)
  Size of Image:                 0x01103000
  Size of Headers:               0x00000600

[SECURITY FEATURES]
  ASLR (DYNAMIC_BASE):           ✓ ENABLED
  DEP (NX_COMPAT):               ✓ ENABLED

[DATA DIRECTORIES]
  Export:              Not present
  Import:              RVA 0x010D5000, Size 1168 bytes
  Resource:            Not present
  Base Relocation:     RVA 0x010D6000, Size 177256 bytes
  Debug:               Not present
  TLS:                 Not present
  IAT:                 RVA 0x01024040, Size 328 bytes

[SECTIONS]
  .text    RVA 0x00001000  Raw 0x00000600  Perms: R-X  Size: 10368KB
  .rdata   RVA 0x00A22000  Raw 0x00A20800  Perms: R--  Size: 6148KB
  .data    RVA 0x01024000  Raw 0x01021C00  Perms: RW-  Size: 705KB
  .idata   RVA 0x010D5000  Raw 0x01063E00  Perms: RW-  Size: 1KB
           ℹ️  Custom section name (possible packer)
  .reloc   RVA 0x010D6000  Raw 0x01064400  Perms: R--  Size: 173KB
  .symtab  RVA 0x01102000  Raw 0x0108FA00  Perms: R--  Size: 0KB
           ℹ️  Custom section name (possible packer)

[IMPORT ANALYSIS]
  Imported DLLs: 1

  ⚠️  Suspicious imports detected:
    kernel32.dll:
      - VirtualAlloc
      - LoadLibraryA
      - LoadLibraryW
      - GetProcAddress
```




---

## Understanding What You Built

### The Parsing Flow

```
1. Read DOS Header (offset 0)
   ├─ Validate MZ signature
   └─ Get e_lfanew → PE header location

2. Read PE Signature (offset e_lfanew)
   └─ Validate PE\0\0

3. Read File Header (offset e_lfanew + 4)
   ├─ Get architecture, section count
   └─ Get characteristics (relocations, DLL flag)

4. Read Optional Header (follows File Header)
   ├─ Get entry point, ImageBase, alignments
   ├─ Get security flags (ASLR, DEP)
   └─ Parse Data Directories (roadmap)

5. Read Section Table (follows Optional Header)
   ├─ For each section: RVA, size, permissions
   └─ Build RVA→FileOffset mapping

6. Parse Import Directory
   ├─ Convert RVA to file offset (using sections)
   ├─ For each DLL: enumerate functions
   └─ Detect suspicious APIs
```

### Key Insights

**RVA to File Offset Conversion:**
This function is the heart of PE parsing:
```
For each section:
  If RVA >= section.VirtualAddress AND
     RVA < section.VirtualAddress + section.VirtualSize:
    
    offset_in_section = RVA - section.VirtualAddress
    file_offset = section.PointerToRawData + offset_in_section
    return file_offset
```

You do this conversion hundreds of times when parsing imports, exports, resources, etc.

**Why Section Alignment Matters:**
```
Section on disk:  RawAddress=0x400, RawSize=0x800
Section in memory: VirtualAddress=0x1000, VirtualSize=0x750

RVA 0x1200 → File offset calculation:
  offset_in_section = 0x1200 - 0x1000 = 0x200
  file_offset = 0x400 + 0x200 = 0x600

Without understanding this, you'd look at wrong file location!
```

---

## Enhancement Exercises

Want to extend your parser? Try adding:

### Exercise 1: Export Directory Parser
Parse DataDirectory[0] to show what functions this DLL exports (if it's a DLL)

### Exercise 2: Relocation Parser
Parse DataDirectory[5] to show base relocation blocks and count total relocations

### Exercise 3: Resource Parser
Parse DataDirectory[2] to list resources (icons, strings, etc.)

### Exercise 4: Security Scoring
Add a security score (0-100) based on:
- ASLR enabled: +25
- DEP enabled: +25
- Has relocations: +20
- No suspicious imports: +20
- No RWX sections: +10

### Exercise 5: JSON Output
Add a `-json` flag to output results in JSON format for automation

---

## Key Takeaways

### What You've Learned

1. **PE structure is sequential** - each layer builds on the previous
2. **RVA conversion is fundamental** - you can't parse without it
3. **Section table is your map** - it lets you translate RVAs to file offsets
4. **Import parsing is multilevel** - descriptor → DLL → functions
5. **Security features are encoded in flags** - bitwise operations reveal them

### How This Connects to Offensive Operations

**Process Hollowing Implementation:**
```
Your parser finds:
  - Entry Point RVA: 0x12A40
  - ImageBase: 0x140000000
  - SizeOfImage: 0xB8000
  - Sections: where code/data goes
  - Import RVAs: what needs resolving

Process hollowing uses these exact values to:
  - Allocate memory (SizeOfImage)
  - Copy sections (VirtualAddress → destination)
  - Resolve imports (Import Directory)
  - Fix relocations (if ImageBase differs)
  - Set entry point (Entry Point RVA)
```

**Reflective DLL Injection:**
Your parser logic IS the loader logic - just instead of executing from disk, you:
1. Load PE into memory buffer
2. Allocate memory for image
3. Copy sections to correct RVAs
4. Apply relocations
5. Resolve imports
6. Fix permissions
7. Call entry point

**Import Obfuscation:**
Now you understand why dynamic API resolution works:
- Static tools parse import directory (like your parser)
- If import directory is empty/minimal, they see nothing
- Runtime `GetProcAddress` bypasses this entirely

---

## Next Steps

You've now:
- ✓ Manually analyzed PEs
- ✓ Programmatically parsed PEs
- ✓ Connected theory to implementation
- ✓ Built a foundation for offensive PE manipulation


The PE format is no longer a black box - you understand every byte and why it matters!




---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./peB.md" >}})
[|NEXT|]({{< ref "./peD.md" >}})