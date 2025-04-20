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
