---
showTableOfContents: true
title: "Finding Native API Function Addresses (Lab 11.1)"
type: "page"
---

## Goal

We now know that `ntdll.dll` is readily available in our process space and that we can use standard WinAPI functions (`GetModuleHandleW`, `GetProcAddress`) to locate the exported Native API functions within it.

In this lab, we will write a Go program to:
1.  Obtain a handle to the `ntdll.dll` module using `GetModuleHandleW`.
2.  Define a list of target Native API function names.
3.  Use `GetProcAddress` in a loop to find the virtual memory address for each of these functions within the loaded `ntdll.dll`.
4.  Print the names and resolved addresses of the found functions.

## Code

```go
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
		"NtWriteVirtualMemory", 
		"NtCreateThreadEx",
		"NtOpenProcess", 
		"EtwEventWrite", 
	}

	fmt.Println("[*] Getting handle to ntdll.dll...")

	// Declare ntdllHandle as windows.Handle
	var ntdllHandle windows.Handle

	// Module name to find
	moduleName := "ntdll.dll"

	// Convert the module name to a UTF-16 pointer.
	moduleNamePtr := windows.StringToUTF16Ptr(moduleName)
	// Note: In a production scenario where moduleName might come from untrusted input,
	// we might want to validate it for NUL characters before this call, or use a safer conversion.

	// Flags: windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT ensures the module's reference count is not incremented.
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
		procAddr, errGetProc := windows.GetProcAddress(ntdllHandle, funcName) 
		if errGetProc != nil {
			// Log error but continue trying other functions
			log.Printf("[!] Warning: GetProcAddress failed for '%s': %v", funcName, errGetProc)
			funcAddrs[funcName] = 0 
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

	// The 'funcAddrs' map now holds the addresses (or 0 if not found)
	fmt.Println("[*] Example: NtAllocateVirtualMemory address:", funcAddrs["NtAllocateVirtualMemory"])
}

```



## Code Breakdown
**`targetFunctions` Slice**
- Defines a string slice containing the names of the Native API functions we are interested in locating. We've included the key ones for process injection and an ETW function for future reference.



**`GetModuleHandleEx`**
- Declares a `windows.Handle` variable `ntdllHandle` to store the module handle.
- Defines the `moduleName` string as "ntdll.dll".
- Calls `windows.StringToUTF16Ptr(moduleName)` to convert the module name into a UTF-16 encoded pointer, which is required by `GetModuleHandleExW` (the underlying Windows API function that `windows.GetModuleHandleEx` calls when a `*uint16` is passed for the module name).
    - Calls `windows.GetModuleHandleEx` with:
        - `windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT`: This flag ensures that the module's reference count is not incremented, mimicking the behavior of the older `GetModuleHandle` function.
        - `moduleNamePtr`: The pointer to the UTF-16 string for "ntdll.dll".
        - `&ntdllHandle`: A pointer to the `ntdllHandle` variable where the retrieved handle will be stored.
    - Includes error checking to ensure the call to `GetModuleHandleEx` was successful and that a valid, non-zero handle is returned.



**`funcAddrs` Map:**
- Initializes an empty map to store the function names (string) as keys and their resolved addresses (`uintptr`) as values.
- Using a map provides a convenient way to store and later retrieve the addresses by name.



**Finding Loop:**
- Iterates through each `funcName` in the `targetFunctions` slice.
- **`GetProcAddress`:** Inside the loop, `windows.GetProcAddress` is called with the `ntdllHandle` (which is a `windows.Handle`, compatible with `HMODULE`) and the current `funcName`. Note that `GetProcAddress` takes an ANSI string (`LPCSTR`) for the function name, and the Go wrapper handles this conversion from Go's native string type.
- **Error Handling:** Checks for errors returned by `GetProcAddress`. If an error occurs (e.g., function genuinely doesn't exist by that name), it logs a warning but continues to the next function. It also specifically checks if `procAddr` is `0` even if no error was returned, logging a warning in that case too.
- **Store Address:** If `GetProcAddress` succeeds and returns a non-zero address, it prints the found address (in hex) and stores the `funcName` and `procAddr` in the `funcAddrs` map. If it fails, `0` is stored to indicate failure for that function.



## Instructions

Compile using `go build`.

```shell
GOOS=windows GOARCH=amd64 go build -buildvcs=false
```


In case it’s required, transfer the binary over to the target system, and run it.


```shell
.\native_func.exe
```



## Results


```shell
PS C:\Users\vuilhond\Desktop> .\native_func.exe
[+] Native API Function Finder
[*] Getting handle to ntdll.dll...
[+] Got ntdll.dll handle: 0x7FF99F190000
[*] Finding function addresses...
  [+] Found 'NtAllocateVirtualMemory' at address: 0x7FF99F22D7E0
  [+] Found 'NtProtectVirtualMemory' at address: 0x7FF99F22DEE0
  [+] Found 'NtWriteVirtualMemory' at address: 0x7FF99F22DC20
  [+] Found 'NtCreateThreadEx' at address: 0x7FF99F22ED10
  [+] Found 'NtOpenProcess' at address: 0x7FF99F22D9A0
  [+] Found 'EtwEventWrite' at address: 0x7FF99F1E0300
[+] Function address lookup complete.
[*] Example: NtAllocateVirtualMemory address: 140710093445088

```

## Discussion
This lab successfully demonstrates how to we can obtain a handle to `ntdll.dll` and resolve the addresses of its exported Native API functions using standard WinAPI calls (`GetModuleHandleEx`, `GetProcAddress`) via their Go wrappers in the `golang.org/x/sys/windows` package.

We now have the _addresses_ of numerous lower-level functions. Storing these addresses (e.g., in our `funcAddrs` map) allows us to use them later when we implement calls using mechanisms like `syscall.SyscallN` or assembly stubs, thereby bypassing hooks on the `kernel32.dll` equivalents.

## Conclusion
We've successfully located the necessary Native API functions within `ntdll.dll`. This prepares us for the next step: actually calling these functions from Go, which involves understanding their specific signatures and using appropriate methods like the `syscall` package.


---
