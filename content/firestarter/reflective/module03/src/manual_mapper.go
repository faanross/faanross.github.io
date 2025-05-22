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

	// --- Step 1: Read DLL and Parse Headers (similar to Lab 2.1) ---
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

	// Keep the program alive briefly if needed for external debugging
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
