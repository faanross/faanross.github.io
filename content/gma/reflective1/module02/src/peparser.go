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
