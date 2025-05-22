//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows"
)

func main() {
	fmt.Println("[+] Native API Function Finder")

	// Target Native API functions we want to find
	targetFunctions := []string{
		"NtAllocateVirtualMemory",
		"NtProtectVirtualMemory",
		"NtWriteVirtualMemory", // Added for completeness in injection sequence
		"NtCreateThreadEx",
		"NtOpenProcess", // Added for completeness
		"EtwEventWrite", // Example ETW function often targeted for patching
	}

	fmt.Println("[*] Getting handle to ntdll.dll...")

	// Declare ntdllHandle as windows.Handle
	var ntdllHandle windows.Handle

	// Module name to find
	moduleName := "ntdll.dll"

	// Convert the module name to a UTF-16 pointer.
	// windows.StringToUTF16Ptr panics on error (e.g., if moduleName contains a NUL character),
	// so we don't expect an error return value here.
	moduleNamePtr := windows.StringToUTF16Ptr(moduleName)
	// Note: In a production scenario where moduleName might come from untrusted input,
	// you might want to validate it for NUL characters before this call, or use a safer conversion.

	// Call GetModuleHandleEx.
	// Flags: windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT ensures the module's reference count is not incremented.
	// This is similar to the behavior of the classic GetModuleHandle.
	// The function expects a pointer to a Handle to store the result.
	err := windows.GetModuleHandleEx(windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT, moduleNamePtr, &ntdllHandle)
	if err != nil {
		log.Fatalf("[-] Failed to get handle for ntdll.dll using GetModuleHandleEx: %v", err)
	}

	// Check if the handle is NULL (0), which indicates failure.
	if ntdllHandle == 0 {
		log.Fatalf("[-] GetModuleHandleEx for ntdll.dll returned a NULL handle.")
	}
	fmt.Printf("[+] Got ntdll.dll handle: 0x%X\n", ntdllHandle)

	fmt.Println("[*] Finding function addresses...")

	// Map to store function names and their addresses
	funcAddrs := make(map[string]uintptr)

	// Loop through the target function names
	for _, funcName := range targetFunctions {
		// GetProcAddress expects HMODULE (which ntdllHandle is, as windows.Handle is a HMODULE)
		// and LPCSTR (ANSI string for the function name).
		procAddr, errGetProc := windows.GetProcAddress(ntdllHandle, funcName) // Renamed err to errGetProc to avoid conflict
		if errGetProc != nil {
			// Log error but continue trying other functions
			log.Printf("[!] Warning: GetProcAddress failed for '%s': %v", funcName, errGetProc)
			funcAddrs[funcName] = 0 // Indicate failure
		} else {
			if procAddr == 0 {
				log.Printf("[!] Warning: GetProcAddress returned NULL for '%s'. Function might not exist or name is wrong.", funcName)
				funcAddrs[funcName] = 0
			} else {
				fmt.Printf("  [+] Found '%s' at address: 0x%X\n", funcName, procAddr)
				funcAddrs[funcName] = procAddr // Store the address
			}
		}
	}

	fmt.Println("[+] Function address lookup complete.")

	// Optional: Verify we found addresses for key functions needed later
	if funcAddrs["NtAllocateVirtualMemory"] == 0 || funcAddrs["NtCreateThreadEx"] == 0 {
		log.Println("[-] Critical function address(es) not found. Subsequent labs might fail.")
	}

	// The 'funcAddrs' map now holds the addresses (or 0 if not found)
	// In a real tool, these addresses would be used for syscalls.
	fmt.Println("[*] Example: NtAllocateVirtualMemory address:", funcAddrs["NtAllocateVirtualMemory"])
}
