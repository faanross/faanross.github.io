---
showTableOfContents: true
title: "PE Header Parser in Go (Lab 2.2)"
type: "page"
---


## Goal

In our previous lab we used PE-Bear, which essentially just parses a PE file, interprets the information, and presents it to us in a clear and logical GUI. We're now going to peel away one layer of abstraction and do the parsing ourselves, which will help us develop a clearer understanding of where the value reside in the actual DLL, and how we can calculate/interpret them directly.

Note that I'll once again use the `calc_dll.dll` file we created in Lab 1.1, but of course feel free to explore other files (64-bit only) if you'd like.


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
)

// --- PE Structures ---

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
	fmt.Println("[+] Starting PE Header Parser...")

	// Check for command line argument
	if len(os.Args) < 2 {
		log.Fatalf("[-] Usage: %s <path_to_dll>\n", os.Args[0])
	}
	dllPath := os.Args[1]

	fmt.Printf("[+] Reading file: %s\n", dllPath)

	// Read the entire DLL file into memory
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read file '%s': %v\n", dllPath, err)
	}

	// Create a reader for easier parsing with encoding/binary
	reader := bytes.NewReader(dllBytes)

	// Parse IMAGE_DOS_HEADER
	var dosHeader IMAGE_DOS_HEADER
	err = binary.Read(reader, binary.LittleEndian, &dosHeader)
	if err != nil {
		log.Fatalf("[-] Failed to read DOS header: %v\n", err)
	}

	// Validate DOS signature ("MZ")
	if dosHeader.Magic != IMAGE_DOS_SIGNATURE {
		log.Fatalf("[-] Invalid DOS signature (Expected 0x%X, Got 0x%X)\n", IMAGE_DOS_SIGNATURE, dosHeader.Magic)
	}
	fmt.Printf("[+] DOS Signature: MZ (0x%X)\n", dosHeader.Magic)
	fmt.Printf("[+] Offset to NT Headers (e_lfanew): 0x%X (%d)\n", dosHeader.Lfanew, dosHeader.Lfanew)

	// Seek to the NT Headers offset specified in the DOS header
	_, err = reader.Seek(int64(dosHeader.Lfanew), 0) // 0 means relative to start of file
	if err != nil {
		log.Fatalf("[-] Failed to seek to NT Headers offset (0x%X): %v\n", dosHeader.Lfanew, err)
	}

	// Read and validate PE signature ("PE\0\0")
	var peSignature uint32
	err = binary.Read(reader, binary.LittleEndian, &peSignature)
	if err != nil {
		log.Fatalf("[-] Failed to read PE signature: %v\n", err)
	}
	if peSignature != IMAGE_NT_SIGNATURE {
		log.Fatalf("[-] Invalid PE signature (Expected 0x%X, Got 0x%X)\n", IMAGE_NT_SIGNATURE, peSignature)
	}
	fmt.Printf("[+] PE Signature: PE\\0\\0 (0x%X)\n", peSignature)

	// Read IMAGE_FILE_HEADER
	var fileHeader IMAGE_FILE_HEADER
	err = binary.Read(reader, binary.LittleEndian, &fileHeader)
	if err != nil {
		log.Fatalf("[-] Failed to read File Header: %v\n", err)
	}

	fmt.Printf("--- File Header ---\n")
	fmt.Printf("  Machine: 0x%X (%s)\n", fileHeader.Machine, machineTypeToString(fileHeader.Machine))
	fmt.Printf("  NumberOfSections: %d\n", fileHeader.NumberOfSections)
	fmt.Printf("  SizeOfOptionalHeader: %d bytes\n", fileHeader.SizeOfOptionalHeader)
	fmt.Printf("  Characteristics: 0x%X\n", fileHeader.Characteristics)

	// Read IMAGE_OPTIONAL_HEADER64 (Assuming 64-bit DLL for this lab)
	if fileHeader.SizeOfOptionalHeader == 0 {
		log.Fatalf("[-] Optional Header size is zero, cannot proceed.")
	}

	var optionalHeader IMAGE_OPTIONAL_HEADER64
	err = binary.Read(reader, binary.LittleEndian, &optionalHeader)
	if err != nil {
		// If the optional header read fails, it might be because the size doesn't match IMAGE_OPTIONAL_HEADER64
		// Check if SizeOfOptionalHeader indicates a different size (e.g., 32-bit)
		log.Printf("[-] Failed to read Optional Header (tried 64-bit): %v. Expected size %d bytes.\n", err, binary.Size(optionalHeader))
		// You might want to add logic here to try parsing IMAGE_OPTIONAL_HEADER32 if needed.
		log.Fatalf("Stopping execution.")
	}

	// Basic check for 64-bit magic number
	if optionalHeader.Magic != 0x20b {
		log.Printf("[!] Warning: Optional Header Magic (0x%X) is not 0x20b (PE32+), parsing may be incorrect if not 64-bit.\n", optionalHeader.Magic)
		// Consider adding a check for 0x10b (PE32) here.
	}

	fmt.Printf("--- Optional Header (64-bit) ---\n")
	fmt.Printf("  Magic: 0x%X (%s)\n", optionalHeader.Magic, magicTypeToString(optionalHeader.Magic))
	fmt.Printf("  AddressOfEntryPoint (RVA): 0x%X\n", optionalHeader.AddressOfEntryPoint)
	fmt.Printf("  ImageBase: 0x%X\n", optionalHeader.ImageBase)
	fmt.Printf("  SizeOfImage: 0x%X (%d bytes)\n", optionalHeader.SizeOfImage, optionalHeader.SizeOfImage)
	fmt.Printf("  SizeOfHeaders: 0x%X (%d bytes)\n", optionalHeader.SizeOfHeaders, optionalHeader.SizeOfHeaders)
	fmt.Printf("  NumberOfRvaAndSizes: %d\n", optionalHeader.NumberOfRvaAndSizes)

	// --- Section Headers ---
	fmt.Printf("--- Section Headers (%d) ---\n", fileHeader.NumberOfSections)
	// Section headers immediately follow the optional header.
	// The reader is already positioned correctly after reading the optional header.
	for i := uint16(0); i < fileHeader.NumberOfSections; i++ {
		var sectionHeader IMAGE_SECTION_HEADER
		err = binary.Read(reader, binary.LittleEndian, &sectionHeader)
		if err != nil {
			log.Fatalf("[-] Failed to read Section Header %d: %v\n", i, err)
		}

		sectionName := sectionNameToString(sectionHeader.Name)
		fmt.Printf("  Section %d: '%s'\n", i, sectionName)
		fmt.Printf("    VirtualAddress (RVA): 0x%X\n", sectionHeader.VirtualAddress)
		fmt.Printf("    SizeOfRawData: 0x%X (%d bytes)\n", sectionHeader.SizeOfRawData, sectionHeader.SizeOfRawData)
		fmt.Printf("    PointerToRawData: 0x%X (%d)\n", sectionHeader.PointerToRawData, sectionHeader.PointerToRawData)
		fmt.Printf("    Characteristics: 0x%X\n", sectionHeader.Characteristics)
	}

	fmt.Println("[+] PE Header Parser finished.")
}

// Helper functions for printing descriptive names
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
	// Add other common types if needed
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


- **Imports:** The program imports several standard Go packages. `os` is needed for interacting with the operating system, specifically to read the file specified by the user and to access command-line arguments. `fmt` and `log` are used for printing output: `fmt` for standard informational messages and `log` for formatted error messages, often followed by program termination (`log.Fatalf`). `encoding/binary` is essential for reading the binary data directly from the PE file and mapping it onto Go struct layouts, handling byte order (endianness). `bytes` provides the `bytes.Reader`, a tool that allows reading from a byte slice as if it were a file, supporting sequential reads and seeking to specific positions.

- **Struct Definitions:** The code defines several Go structs (`IMAGE_DOS_HEADER`, `IMAGE_FILE_HEADER`, `IMAGE_DATA_DIRECTORY`, `IMAGE_OPTIONAL_HEADER64`, `IMAGE_SECTION_HEADER`). These structures are designed to precisely mirror the C structures found in Windows development headers (like `winnt.h`) which define the Portable Executable (PE) file format. By matching these structures, `encoding/binary` can automatically populate the Go struct fields from the file's byte stream. The example specifically uses `IMAGE_OPTIONAL_HEADER64`, assuming the target PE file is 64-bit, which dictates the size and layout of the optional header.

- **`sectionNameToString` Helper:** This is a utility function created to handle the fixed-size, null-padded character arrays (`[8]byte`) used for section names within the PE format. It takes the 8-byte array, finds the position of the first null byte (which signifies the end of the string), and returns a standard Go string containing only the characters before the null byte. If no null byte is found, it assumes the name uses all 8 bytes.

`main` Function Logic:**

- **Argument Handling:** The program first checks `os.Args` (a slice containing command-line arguments) to ensure the user provided at least one argument after the program name, which should be the path to the PE file (e.g., a DLL or EXE). If not, it prints a usage message via `log.Fatalf` and exits.

- **File Reading:** It attempts to read the entire content of the file specified by the path argument into a byte slice named `dllBytes` using `os.ReadFile`. Errors during reading (e.g., file not found, permissions issues) are fatal.

- **Reader Creation:** A `bytes.Reader` is initialized with the `dllBytes`. This provides an `io.Reader` and `io.Seeker` interface over the in-memory byte slice, necessary for `encoding/binary` and for jumping to different offsets within the file data.

- **DOS Header Parsing:** `binary.Read` is called to read bytes from the beginning of the `bytes.Reader` directly into the `dosHeader` struct. It uses `binary.LittleEndian` because the PE format specifies little-endian byte ordering for its fields. Error checking is performed.

- **DOS Signature Check:** The `Magic` field of the parsed `dosHeader` is compared against the constant `IMAGE_DOS_SIGNATURE` (0x5A4D, representing the characters "MZ") to confirm it's likely a valid PE file. An invalid signature results in a fatal error.

- **Seeking to NT Headers:** The crucial `Lfanew` field from the `dosHeader` contains the file offset (from the beginning of the file) where the main PE headers (NT Headers) start. The program uses `reader.Seek` to move the current reading position within the `bytes.Reader` to this specific offset.

- **PE Signature Check:** Immediately after seeking, the program reads the next 4 bytes, expecting them to match the `IMAGE_NT_SIGNATURE` constant (0x00004550, representing "PE\0\0"). This confirms the start of the NT Headers.

- **File & Optional Header Parsing:** Following the PE signature, `binary.Read` is used again to parse the `IMAGE_FILE_HEADER` structure, followed immediately by the `IMAGE_OPTIONAL_HEADER64` structure (assuming 64-bit). Error checking occurs after each read.

- **Information Printing:** Key information extracted from the parsed headers (like the target machine type from `fileHeader.Machine`, the number of sections `fileHeader.NumberOfSections`, the entry point RVA `optionalHeader.AddressOfEntryPoint`, the preferred load address `optionalHeader.ImageBase`, the total memory size `optionalHeader.SizeOfImage`, and the size of all headers `optionalHeader.SizeOfHeaders`) is printed to the console using `fmt.Printf`. Hexadecimal formatting (`0x%X`) is commonly used for addresses and flags. The helper functions `machineTypeToString` and `magicTypeToString` are called to convert raw numeric values (like machine code or optional header magic number) into human-readable strings.

- **Section Header Loop:** The section headers are known to reside in the file immediately following the Optional Header. The code enters a loop that iterates `fileHeader.NumberOfSections` times. In each iteration, it uses `binary.Read` to parse one `IMAGE_SECTION_HEADER` struct. It then prints the section's name (using `sectionNameToString`), its relative virtual address (`VirtualAddress`), its size on disk (`SizeOfRawData`), and its file offset (`PointerToRawData`).

## Instructions

- Compile the parser
```shell
GOOS=windows GOARCH=amd64 go build
```

- Then copy it over to target system and invoke from command-line, providing as argument the dll you'd like to analyze, for example

```shell
C:\Users\vuilhond\Desktop> .\peparser.exe .\calc_dll.dll
```

## Output

```Shell
[+] Starting PE Header Parser...
[+] Reading file: .\calc_dll.dll
[+] DOS Signature: MZ (0x5A4D)
[+] Offset to NT Headers (e_lfanew): 0x80 (128)
[+] PE Signature: PE\0\0 (0x4550)
--- File Header ---
  Machine: 0x8664 (x64 (AMD64))
  NumberOfSections: 19
  SizeOfOptionalHeader: 240 bytes
  Characteristics: 0x2026
--- Optional Header (64-bit) ---
  Magic: 0x20B (PE32+ (64-bit))
  AddressOfEntryPoint (RVA): 0x1330
  ImageBase: 0x26A5B0000
  SizeOfImage: 0x22000 (139264 bytes)
  SizeOfHeaders: 0x600 (1536 bytes)
  NumberOfRvaAndSizes: 16
--- Section Headers (19) ---
  Section 0: '.text'
    VirtualAddress (RVA): 0x1000
    SizeOfRawData: 0x1800 (6144 bytes)
    PointerToRawData: 0x600 (1536)
    Characteristics: 0x60000020
  Section 1: '.data'
    VirtualAddress (RVA): 0x3000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x1E00 (7680)
    Characteristics: 0xC0000040
  Section 2: '.rdata'
    VirtualAddress (RVA): 0x4000
    SizeOfRawData: 0x600 (1536 bytes)
    PointerToRawData: 0x2000 (8192)
    Characteristics: 0x40000040
  Section 3: '.pdata'
    VirtualAddress (RVA): 0x5000
    SizeOfRawData: 0x400 (1024 bytes)
    PointerToRawData: 0x2600 (9728)
    Characteristics: 0x40000040
  Section 4: '.xdata'
    VirtualAddress (RVA): 0x6000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x2A00 (10752)
    Characteristics: 0x40000040
  Section 5: '.bss'
    VirtualAddress (RVA): 0x7000
    SizeOfRawData: 0x0 (0 bytes)
    PointerToRawData: 0x0 (0)
    Characteristics: 0xC0000080
  Section 6: '.edata'
    VirtualAddress (RVA): 0x8000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x2C00 (11264)
    Characteristics: 0x40000040
  Section 7: '.idata'
    VirtualAddress (RVA): 0x9000
    SizeOfRawData: 0x800 (2048 bytes)
    PointerToRawData: 0x2E00 (11776)
    Characteristics: 0x40000040
  Section 8: '.tls'
    VirtualAddress (RVA): 0xA000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x3600 (13824)
    Characteristics: 0xC0000040
  Section 9: '.reloc'
    VirtualAddress (RVA): 0xB000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x3800 (14336)
    Characteristics: 0x42000040
  Section 10: '/4'
    VirtualAddress (RVA): 0xC000
    SizeOfRawData: 0x400 (1024 bytes)
    PointerToRawData: 0x3A00 (14848)
    Characteristics: 0x42000040
  Section 11: '/19'
    VirtualAddress (RVA): 0xD000
    SizeOfRawData: 0x9400 (37888 bytes)
    PointerToRawData: 0x3E00 (15872)
    Characteristics: 0x42000040
  Section 12: '/31'
    VirtualAddress (RVA): 0x17000
    SizeOfRawData: 0x1A00 (6656 bytes)
    PointerToRawData: 0xD200 (53760)
    Characteristics: 0x42000040
  Section 13: '/45'
    VirtualAddress (RVA): 0x19000
    SizeOfRawData: 0x1A00 (6656 bytes)
    PointerToRawData: 0xEC00 (60416)
    Characteristics: 0x42000040
  Section 14: '/57'
    VirtualAddress (RVA): 0x1B000
    SizeOfRawData: 0xA00 (2560 bytes)
    PointerToRawData: 0x10600 (67072)
    Characteristics: 0x42000040
  Section 15: '/70'
    VirtualAddress (RVA): 0x1C000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x11000 (69632)
    Characteristics: 0x42000040
  Section 16: '/81'
    VirtualAddress (RVA): 0x1D000
    SizeOfRawData: 0x1800 (6144 bytes)
    PointerToRawData: 0x11200 (70144)
    Characteristics: 0x42000040
  Section 17: '/97'
    VirtualAddress (RVA): 0x1F000
    SizeOfRawData: 0x1400 (5120 bytes)
    PointerToRawData: 0x12A00 (76288)
    Characteristics: 0x42000040
  Section 18: '/113'
    VirtualAddress (RVA): 0x21000
    SizeOfRawData: 0x200 (512 bytes)
    PointerToRawData: 0x13E00 (81408)
    Characteristics: 0x42000040
[+] PE Header Parser finished.
```

## Results

Executing the Go PE parser script (`peparser.exe`) on `calc_dll.dll` successfully yielded results that directly correspond to the values observed using PE-Bear in our previous lab.

* **DOS Header:** The script correctly identified the `MZ` magic number (`0x5A4D`) and the crucial `e_lfanew` offset (`0x80`).
* **NT Headers Signature:** The script confirmed the `PE\0\0` signature (`0x4550`) at the expected offset `0x80`.
* **File Header:** Key values matched PE-Bear's findings:
    * `Machine`: `0x8664` (x64)
    * `NumberOfSections`: `19`
    * `SizeOfOptionalHeader`: `240` bytes (0xF0)
    * `Characteristics`: `0x2026`, indicating a DLL among other flags.
* **Optional Header:** The script's output aligned with the values noted in PE-Bear:
    * `Magic`: `0x20B` (PE32+)
    * `AddressOfEntryPoint` (RVA): `0x1330`
    * `ImageBase`: `0x26A5B0000`
    * `SizeOfImage`: `0x22000`
    * `SizeOfHeaders`: `0x600`
    * `NumberOfRvaAndSizes`: `16`
* **`.text` Section Header:** The script successfully parsed the primary code section's header, matching PE-Bear:
    * `VirtualAddress` (RVA): `0x1000`
    * `PointerToRawData` (Raw Addr): `0x600`
    * `SizeOfRawData` (Raw size): `0x1800`
    * `Characteristics`: `0x60000020` (Read/Execute permissions)

## Conclusion
Our application successfully interprets the fundamental structures of a PE file, 
extracting the information necessary to understand how to map the file into memory. This logic will be essential
to our reflective loader, which we'll now begin discussing in Module 03. 



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "pebear.md" >}})
[|NEXT|]({{< ref "../module03/intro.md" >}})