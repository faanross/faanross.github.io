//go:build windows
// +build windows

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// --- Existing PE Structures ---
type IMAGE_DOS_HEADER struct {
	Magic  uint16
	_      [58]byte
	Lfanew int32
} //nolint:revive
type IMAGE_FILE_HEADER struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}                                                               //nolint:revive
type IMAGE_DATA_DIRECTORY struct{ VirtualAddress, Size uint32 } //nolint:revive
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
} //nolint:revive
type IMAGE_SECTION_HEADER struct {
	Name                                                                                                     [8]byte
	VirtualSize, VirtualAddress, SizeOfRawData, PointerToRawData, PointerToRelocations, PointerToLinenumbers uint32
	NumberOfRelocations, NumberOfLinenumbers                                                                 uint16
	Characteristics                                                                                          uint32
}                                                                                                                 //nolint:revive
type IMAGE_BASE_RELOCATION struct{ VirtualAddress, SizeOfBlock uint32 }                                           //nolint:revive
type IMAGE_IMPORT_DESCRIPTOR struct{ OriginalFirstThunk, TimeDateStamp, ForwarderChain, Name, FirstThunk uint32 } //nolint:revive

// --- Constants ---
const (
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

// --- Main Function ---
func main() {
	// Ensure running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("[-] This program must be run on Windows.")
	}
	fmt.Println("[+] Starting Manual DLL Mapper (with IAT Resolution)...")

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

	// --- Step 2: Allocate Memory for DLL ---
	fmt.Printf("[+] Allocating 0x%X bytes of memory for DLL...\n", optionalHeader.SizeOfImage)
	allocSize := uintptr(optionalHeader.SizeOfImage)
	preferredBase := uintptr(optionalHeader.ImageBase)
	allocBase, err := windows.VirtualAlloc(preferredBase, allocSize, windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
	if err != nil {
		fmt.Printf("[*] Failed to allocate at preferred base 0x%X: %v. Trying arbitrary address...\n", preferredBase, err)
		allocBase, err = windows.VirtualAlloc(0, allocSize, windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
		if err != nil {
			log.Fatalf("[-] Failed to allocate memory at arbitrary address: %v\n", err)
		}
	}
	fmt.Printf("[+] DLL memory allocated successfully at actual base address: 0x%X\n", allocBase)
	defer func() {
		fmt.Printf("[+] Attempting to free main DLL allocation at 0x%X...\n", allocBase)
		err := windows.VirtualFree(allocBase, 0, windows.MEM_RELEASE)
		if err != nil {
			log.Printf("[!] Warning: Failed to free main DLL memory: %v\n", err)
		} else {
			fmt.Println("[+] Main DLL memory freed successfully.")
		}
	}()

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

	// --- Step 4: Copy Sections into Allocated Memory ---
	fmt.Println("[+] Copying sections...")
	firstSectionHeaderAddr := allocBase + uintptr(dosHeader.Lfanew) + 4 + unsafe.Sizeof(fileHeader) + uintptr(fileHeader.SizeOfOptionalHeader) // Address of first section header IN allocBase
	sectionHeaderSize := unsafe.Sizeof(IMAGE_SECTION_HEADER{})
	numSections := fileHeader.NumberOfSections
	for i := uint16(0); i < numSections; i++ {
		currentSectionHeaderAddr := firstSectionHeaderAddr + uintptr(i)*sectionHeaderSize
		sectionHeader := (*IMAGE_SECTION_HEADER)(unsafe.Pointer(currentSectionHeaderAddr))
		// sectionName := sectionNameToString(sectionHeader.Name) // Less verbose logging
		if sectionHeader.SizeOfRawData == 0 {
			continue
		}
		if uintptr(sectionHeader.PointerToRawData)+uintptr(sectionHeader.SizeOfRawData) > uintptr(len(dllBytes)) {
			log.Printf("[!] Warning: Section %d ('%s') raw data exceeds file size. Skipping copy.", i, sectionNameToString(sectionHeader.Name))
			continue
		}
		sourceAddr := dllBytesPtr + uintptr(sectionHeader.PointerToRawData)
		if uintptr(sectionHeader.VirtualAddress)+uintptr(sectionHeader.SizeOfRawData) > allocSize {
			log.Printf("[!] Warning: Section %d ('%s') virtual address/size exceeds allocated size. Skipping copy.", i, sectionNameToString(sectionHeader.Name))
			continue
		}
		destAddr := allocBase + uintptr(sectionHeader.VirtualAddress)
		sizeToCopy := uintptr(sectionHeader.SizeOfRawData)
		err = windows.WriteProcessMemory(windows.CurrentProcess(), destAddr, (*byte)(unsafe.Pointer(sourceAddr)), sizeToCopy, &bytesWritten)
		if err != nil || bytesWritten != sizeToCopy {
			log.Fatalf("    [-] Failed to copy section '%s': %v (Bytes written: %d)", sectionNameToString(sectionHeader.Name), err, bytesWritten)
		}
	}
	fmt.Println("[+] All sections copied.")

	// --- Step 5: Process Base Relocations ---
	fmt.Println("[+] Checking if base relocations are needed...")
	delta := int64(allocBase) - int64(optionalHeader.ImageBase)
	if delta == 0 {
		fmt.Println("[+] Image loaded at preferred base. No relocations needed.")
	} else {
		fmt.Printf("[+] Image loaded at non-preferred base (Delta: 0x%X). Processing relocations...\n", delta)
		relocDirEntry := optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_BASERELOC]
		relocDirRVA := relocDirEntry.VirtualAddress
		relocDirSize := relocDirEntry.Size
		if relocDirRVA == 0 || relocDirSize == 0 {
			fmt.Println("[!] Warning: Image rebased, but no relocation directory found or empty.")
		} else {
			fmt.Printf("[+] Relocation Directory found at RVA 0x%X, Size 0x%X\n", relocDirRVA, relocDirSize)
			relocTableBase := allocBase + uintptr(relocDirRVA)
			relocTableEnd := relocTableBase + uintptr(relocDirSize)
			currentBlockAddr := relocTableBase
			totalFixups := 0
			for currentBlockAddr < relocTableEnd {
				if currentBlockAddr < allocBase || currentBlockAddr+unsafe.Sizeof(IMAGE_BASE_RELOCATION{}) > allocBase+allocSize {
					log.Printf("[!] Error: Relocation block address 0x%X is outside allocated range. Stopping relocations.", currentBlockAddr)
					break
				}
				blockHeader := (*IMAGE_BASE_RELOCATION)(unsafe.Pointer(currentBlockAddr))
				if blockHeader.VirtualAddress == 0 || blockHeader.SizeOfBlock <= uint32(unsafe.Sizeof(IMAGE_BASE_RELOCATION{})) {
					break
				}
				if currentBlockAddr+uintptr(blockHeader.SizeOfBlock) > relocTableEnd {
					log.Printf("[!] Error: Relocation block size (%d) at 0x%X exceeds directory bounds. Stopping relocations.", blockHeader.SizeOfBlock, currentBlockAddr)
					break
				}
				numEntries := (blockHeader.SizeOfBlock - uint32(unsafe.Sizeof(IMAGE_BASE_RELOCATION{}))) / 2
				entryPtr := currentBlockAddr + unsafe.Sizeof(IMAGE_BASE_RELOCATION{})
				for i := uint32(0); i < numEntries; i++ {
					entryAddr := entryPtr + uintptr(i*2)
					if entryAddr < allocBase || entryAddr+2 > allocBase+allocSize {
						log.Printf("    [!] Error: Relocation entry address 0x%X is outside allocated range. Skipping entry.", entryAddr)
						continue
					}
					entry := *(*uint16)(unsafe.Pointer(entryAddr))
					relocType := entry >> 12
					offset := entry & 0xFFF
					if relocType == IMAGE_REL_BASED_DIR64 {
						patchAddr := allocBase + uintptr(blockHeader.VirtualAddress) + uintptr(offset)
						if patchAddr < allocBase || patchAddr+8 > allocBase+allocSize {
							log.Printf("        [!] Error: Relocation patch address 0x%X is outside allocated range. Skipping fixup.", patchAddr)
							continue
						}
						originalValuePtr := (*uint64)(unsafe.Pointer(patchAddr))
						*originalValuePtr = uint64(int64(*originalValuePtr) + delta)
						totalFixups++
					} else if relocType != IMAGE_REL_BASED_ABSOLUTE {
						fmt.Printf("        [!] Warning: Skipping unhandled relocation type %d at offset 0x%X\n", relocType, offset)
					}
				}
				currentBlockAddr += uintptr(blockHeader.SizeOfBlock)
			}
			fmt.Printf("[+] Relocation processing complete. Total fixups applied: %d\n", totalFixups)
		}
	}
	// --- End Step 5 ---

	// --- Step 6: Process Import Address Table (IAT) ---
	fmt.Println("[+] Processing Import Address Table (IAT)...")
	importDirEntry := optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_IMPORT]
	importDirRVA := importDirEntry.VirtualAddress
	if importDirRVA == 0 {
		fmt.Println("[*] No Import Directory found. Skipping IAT processing.")
	} else {
		fmt.Printf("[+] Import Directory found at RVA 0x%X\n", importDirRVA)
		importDescSize := unsafe.Sizeof(IMAGE_IMPORT_DESCRIPTOR{})
		importDescBase := allocBase + uintptr(importDirRVA)
		importCount := 0
		// fmt.Printf("    DEBUG: Import Directory VA: 0x%X\n", importDescBase)

		// // DEBUG: Read and print the first descriptor BEFORE the loop
		// if importDescBase < allocBase || importDescBase+importDescSize > allocBase+allocSize {
		// 	log.Printf("    [-] Error: Calculated Import Directory VA 0x%X is outside allocated range [0x%X - 0x%X]. Cannot read first descriptor.",
		// 		importDescBase, allocBase, allocBase+allocSize-1)
		// } else {
		// 	firstDesc := (*IMAGE_IMPORT_DESCRIPTOR)(unsafe.Pointer(importDescBase))
		// 	fmt.Printf("    DEBUG: First Descriptor Raw Values: OFT=0x%X, TS=0x%X, FC=0x%X, NameRVA=0x%X, FT=0x%X\n",
		// 		firstDesc.OriginalFirstThunk, firstDesc.TimeDateStamp, firstDesc.ForwarderChain, firstDesc.Name, firstDesc.FirstThunk)
		// }

		// Iterate through IMAGE_IMPORT_DESCRIPTOR array (null terminated)
		for i := 0; ; i++ {
			currentDescAddr := importDescBase + uintptr(i)*importDescSize
			if currentDescAddr < allocBase || currentDescAddr+importDescSize > allocBase+allocSize {
				log.Printf("    [!] Error: Calculated descriptor address 0x%X is outside allocated range. Stopping IAT processing.", currentDescAddr)
				break
			}
			// fmt.Printf("\n    DEBUG: Reading descriptor %d at address 0x%X\n", i, currentDescAddr) // Keep DEBUG optional
			importDesc := (*IMAGE_IMPORT_DESCRIPTOR)(unsafe.Pointer(currentDescAddr))
			// fmt.Printf("        DEBUG: Desc %d: OFT=0x%X, TS=0x%X, FC=0x%X, NameRVA=0x%X, FT=0x%X\n", i, importDesc.OriginalFirstThunk, importDesc.TimeDateStamp, importDesc.ForwarderChain, importDesc.Name, importDesc.FirstThunk) // Keep DEBUG optional

			if importDesc.OriginalFirstThunk == 0 && importDesc.FirstThunk == 0 { /* fmt.Printf("    DEBUG: Null descriptor found at index %d. Stopping.\n", i); */
				break
			}
			importCount++

			dllNameRVA := importDesc.Name
			if dllNameRVA == 0 {
				log.Printf("    [!] Warning: Descriptor %d has null Name RVA. Skipping.", i)
				continue
			}
			dllNamePtrAddr := allocBase + uintptr(dllNameRVA)
			// fmt.Printf("        DEBUG: DLL Name String RVA=0x%X, VA=0x%X\n", dllNameRVA, dllNamePtrAddr)
			if dllNamePtrAddr < allocBase || dllNamePtrAddr >= allocBase+allocSize {
				log.Printf("    [!] Error: Calculated DLL Name VA 0x%X is outside allocated range. Skipping descriptor %d.", dllNamePtrAddr, i)
				continue
			}
			dllNamePtr := (*byte)(unsafe.Pointer(dllNamePtrAddr))
			dllName := windows.BytePtrToString(dllNamePtr)
			fmt.Printf("    [->] Processing imports for: %s\n", dllName)

			hModule, err := windows.LoadLibrary(dllName)
			if err != nil {
				log.Fatalf("    [-] FATAL: Failed to load dependency library '%s': %v\n", dllName, err)
			}
			// fmt.Printf("        [+] Loaded '%s' successfully. Handle: 0x%X\n", dllName, hModule) // Less verbose

			iltRVA := importDesc.OriginalFirstThunk
			if iltRVA == 0 {
				iltRVA = importDesc.FirstThunk
			}
			iatRVA := importDesc.FirstThunk
			if iltRVA == 0 || iatRVA == 0 {
				log.Printf("    [!] Warning: Descriptor %d for '%s' has null ILT/IAT RVA. Skipping.", i, dllName)
				continue
			}
			iltBase := allocBase + uintptr(iltRVA)
			iatBase := allocBase + uintptr(iatRVA)
			entrySize := unsafe.Sizeof(uintptr(0))
			// fmt.Printf("        DEBUG: ILT VA=0x%X, IAT VA=0x%X\n", iltBase, iatBase)

			for j := uintptr(0); ; j++ {
				iltEntryAddr := iltBase + (j * entrySize)
				iatEntryAddr := iatBase + (j * entrySize)
				if iltEntryAddr < allocBase || iltEntryAddr >= allocBase+allocSize {
					log.Printf("    [!] Error: Calculated ILT Entry VA 0x%X is outside allocated range. Stopping imports for %s.", iltEntryAddr, dllName)
					break
				}
				iltEntry := *(*uintptr)(unsafe.Pointer(iltEntryAddr))
				// fmt.Printf("            DEBUG: Reading ILT Entry %d at 0x%X, Value=0x%X\n", j, iltEntryAddr, iltEntry)

				if iltEntry == 0 {
					break
				}

				var funcAddr uintptr
				var procErr error
				importNameStr := ""

				if iltEntry&IMAGE_ORDINAL_FLAG64 != 0 {
					ordinal := uint16(iltEntry & 0xFFFF)
					importNameStr = fmt.Sprintf("Ordinal %d", ordinal)
					// fmt.Printf("            DEBUG: Importing by %s\n", importNameStr)
					ret, _, callErr := procGetProcAddress.Call(uintptr(hModule), uintptr(ordinal))
					if ret == 0 {
						errMsg := fmt.Sprintf("GetProcAddress by ordinal %d returned NULL", ordinal)
						if callErr != nil && callErr != windows.ERROR_SUCCESS {
							procErr = fmt.Errorf("%s - syscall error: %w", errMsg, callErr)
						} else {
							procErr = errors.New(errMsg)
						}
					} else if callErr != nil && callErr != windows.ERROR_SUCCESS {
						procErr = fmt.Errorf("GetProcAddress by ordinal %d syscall failed: %w", ordinal, callErr)
					}
					funcAddr = ret
				} else {
					hintNameRVA := uint32(iltEntry)
					hintNameAddr := allocBase + uintptr(hintNameRVA)
					// fmt.Printf("            DEBUG: Importing by Name. Hint/Name RVA=0x%X, VA=0x%X\n", hintNameRVA, hintNameAddr)
					if hintNameAddr < allocBase || hintNameAddr+2 >= allocBase+allocSize {
						log.Printf("        [!] Error: Calculated Hint/Name VA 0x%X is outside allocated range. Skipping import.", hintNameAddr)
						continue
					}
					funcNamePtr := unsafe.Pointer(hintNameAddr + 2)
					funcName := windows.BytePtrToString((*byte)(funcNamePtr))
					importNameStr = fmt.Sprintf("Function '%s'", funcName)
					// fmt.Printf("            DEBUG: Importing %s\n", importNameStr)
					funcAddr, procErr = windows.GetProcAddress(hModule, funcName)
					if procErr != nil && funcAddr == 0 {
						procErr = fmt.Errorf("GetProcAddress failed for %s: %w", funcName, procErr)
					} else if procErr == nil && funcAddr == 0 {
						procErr = fmt.Errorf("GetProcAddress returned NULL for %s", funcName)
					}
				}

				if procErr != nil || funcAddr == 0 {
					log.Fatalf("        [-] FATAL: Failed to resolve import %s from %s: %v (Addr: 0x%X)\n", importNameStr, dllName, procErr, funcAddr)
				}

				if iatEntryAddr < allocBase || iatEntryAddr >= allocBase+allocSize {
					log.Printf("        [!] Error: Calculated IAT Entry VA 0x%X is outside allocated range. Skipping patch for %s.", iatEntryAddr, importNameStr)
					continue
				}
				iatEntryPtr := (*uintptr)(unsafe.Pointer(iatEntryAddr))
				*iatEntryPtr = funcAddr
				// fmt.Printf("            [+] Resolved %s -> 0x%X. Patched IAT at 0x%X.\n", importNameStr, funcAddr, iatEntryAddr) // Less verbose
			} // End inner loop
			fmt.Printf("    [+] Finished imports for '%s'.\n", dllName)
		} // End outer loop
		fmt.Printf("[+] Import processing complete (%d DLLs).\n", importCount)
	}
	// --- *** End Step 6 *** ---

	// --- Step 7: Self-Check ---
	fmt.Println("[+] Manual mapping process complete (Headers, Sections copied, Relocations potentially applied, IAT resolved).")
	fmt.Println("[+] Self-Check Suggestion: Use a debugger...")
	fmt.Printf("    to inspect the memory at the allocated base address (0x%X).\n", allocBase)
	fmt.Println("    Verify that the 'MZ' and 'PE' signatures are present at the start")
	fmt.Println("    and that data corresponding to sections appears at the correct RVAs.")
	fmt.Println("    If relocations occurred, check known absolute addresses (if any) were patched.")
	fmt.Println("    Inspect the IAT section: pointers should now point to actual function addresses in loaded modules.")

	fmt.Println("\n[+] Press Enter to free memory and exit.")
	fmt.Scanln()

	fmt.Println("[+] Mapper finished.")
}
