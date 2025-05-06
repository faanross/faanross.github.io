---
showTableOfContents: true
title: "Create a Basic Loader in Go (Lab 1.2)"
type: "page"
---
## Goal
In this lab we'll write a simple Windows application written in Go that will serve as a loader for our dll.
It uses standard Windows API functions, accessed via Go's `windows` and `syscall` packages, 
to dynamically load the `calc_dll.dll` (created in Lab 1.1) and execute its exported `LaunchCalc` function. 
This helps us come to grips with the most vanilla method of interacting with DLLs before we move on to more advanced techniques.

## Code
Note that I also provide the code with a hefty helping of explanatory comments at the bottom.
Also note the build tags at the top are only required if you are developing in Darwin/Linux, if you're working
directly inside of Windows these can be omitted. 
```go
//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"syscall"
	_ "unsafe"

	"golang.org/x/sys/windows"
)

func main() {
	fmt.Println("[+] Starting basic Go DLL loader...")

	dllPath := "calc_dll.dll"
	fmt.Printf("[+] Attempting to load DLL: %s\n", dllPath)

	dllHandle, err := windows.LoadLibrary(dllPath)
	if err != nil {
		log.Fatalf("[-] Failed to load DLL '%s': %v\n", dllPath, err)
	}

	defer func() {
		fmt.Println("[+] Attempting to free DLL handle...")
		err := windows.FreeLibrary(dllHandle)
		if err != nil {
			log.Printf("[!] Warning: Failed to free DLL handle: %v\n", err)
		} else {
			fmt.Println("[+] DLL handle freed successfully.")
		}
	}()

	fmt.Printf("[+] DLL loaded successfully. Handle: 0x%X\n", dllHandle)

	funcName := "LaunchCalc"
	fmt.Printf("[+] Attempting to get address of function: %s\n", funcName)

	funcAddr, err := windows.GetProcAddress(dllHandle, funcName)
	if err != nil {
		log.Fatalf("[-] Failed to find function '%s' in DLL: %v\n", funcName, err)
	}

	fmt.Printf("[+] Function '%s' found at address: 0x%X\n", funcName, funcAddr)

	fmt.Printf("[+] Calling function '%s'...\n", funcName)

	ret, _, callErr := syscall.SyscallN(funcAddr, 0, 0, 0, 0)

	if callErr != 0 {
		log.Fatalf("[-] Error occurred during syscall to '%s': %v\n", funcName, callErr)
	}

	if ret != 0 {
		fmt.Printf("[+] Function '%s' executed successfully (returned TRUE).\n", funcName)
	} else {
		fmt.Printf("[-] Function '%s' execution reported failure (returned FALSE).\n", funcName)
	}

	fmt.Println("[+] Loader finished.")
}
```

## Code Breakdown
Imports:
- `fmt`, `log`: Standard packages for printing output and handling errors.
- `syscall`: Used specifically for the `SyscallN` function, which allows calling arbitrary function pointers (like the one returned by `GetProcAddress`).
- `golang.org/x/sys/windows`: The simplest package for interacting with the Windows API. It provides Go-style wrappers like `LoadLibrary` and `GetProcAddress`.

`dllPath := "calc_dll.dll"`:
- Defines the name of the DLL file.
- Since no full path is given, `LoadLibrary` will search for it in standard locations, including the directory where `loader.exe` is launched from.

`windows.LoadLibrary(dllPath):`
- Calls the `LoadLibraryW` Windows API function to load the specified DLL into the current process's memory.
- It returns a handle (`HMODULE`) to the loaded DLL or an error if it fails.

`defer windows.FreeLibrary(dllHandle): `
- Idiomatic Go way of cleaning up.
- The defer statement ensures that `FreeLibrary` is called after the `main` function finishes (either normally or due to a panic).
- `FreeLibrary` decrements the DLL's reference count; the OS unloads the DLL from memory when its reference count drops to zero.

`windows.GetProcAddress(dllHandle, funcName):`
- Calls the `GetProcAddress` Windows API function.
- It takes the handle of the loaded DLL and the name of the exported function ("`LaunchCalc`") and returns the memory address where that function resides, or an error if the function isn't found in the DLL's export table.

`syscall.SyscallN(funcAddr, 0, 0, 0, 0): `
- This is the core execution step.
- `funcAddr`: The memory address of `LaunchCalc` obtained from `GetProcAddress`.
- `0`: The number of arguments our `LaunchCalc` function takes, in this case, none.
- The subsequent `0`s are placeholders for the arguments themselves, since we have `0` arguments, these are just padding.
- It returns `ret` (the function's return value, cast to `uintptr`), a reserved value (usually ignored), and `callErr` (an error object representing the `syscall`'s success/failure status).

Error/Return Value Checks:
- The code checks both `callErr` (did the `syscall` itself fail?) and `ret`
- What did `LaunchCalc` return?
    - `TRUE`/`1` for success,
    - `FALSE`/`0` for failure.

## Instructions
Compile source code into an amd64 *.exe binary for Windows
```shell
GOOS=windows GOARCH=amd64 go build
```

Then simply run executable on target machine, in same directory as *.dll produced in Lab 1.1.

## Expected Outcome
Upon executing the loader `calc.exe` should launch along with the following output printed to terminal:

![expected outcome for lab 1.2](../img/results.png)
___
## Code with Comments
```go
//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"syscall"
	_ "unsafe" // Required only if directly manipulating pointers in complex ways, but good practice to know syscall uses them.

	// Use the preferred windows package for API calls
	"golang.org/x/sys/windows"
)

func main() {
	fmt.Println("[+] Starting basic Go DLL loader...")

	// Define the path to the DLL.
	// LoadLibrary will search in standard locations, including the current directory.
	dllPath := "calc_dll.dll"
	fmt.Printf("[+] Attempting to load DLL: %s\n", dllPath)

	// Load the DLL using LoadLibraryW (the Unicode version)
	// windows.LoadLibrary is a wrapper around the LoadLibraryW Windows API call.
	dllHandle, err := windows.LoadLibrary(dllPath)
	if err != nil {
		// If LoadLibrary fails, err will be non-nil.
		log.Fatalf("[-] Failed to load DLL '%s': %v\n", dllPath, err)
	}
	// If LoadLibrary succeeds, ensure FreeLibrary is called when main exits.
	// This decrements the DLL's reference count.
	defer func() {
		fmt.Println("[+] Attempting to free DLL handle...")
		err := windows.FreeLibrary(dllHandle)
		if err != nil {
			log.Printf("[!] Warning: Failed to free DLL handle: %v\n", err)
		} else {
			fmt.Println("[+] DLL handle freed successfully.")
		}
	}()

	fmt.Printf("[+] DLL loaded successfully. Handle: 0x%X\n", dllHandle)

	// Define the name of the function we want to call
	funcName := "LaunchCalc"
	fmt.Printf("[+] Attempting to get address of function: %s\n", funcName)

	// Get the address of the exported function using GetProcAddress
	// windows.GetProcAddress wraps the GetProcAddress Windows API call.
	// It requires the DLL handle and the function name (as a null-terminated string).
	funcAddr, err := windows.GetProcAddress(dllHandle, funcName)
	if err != nil {
		// If GetProcAddress fails (e.g., function not found), err will be non-nil.
		log.Fatalf("[-] Failed to find function '%s' in DLL: %v\n", funcName, err)
	}

	fmt.Printf("[+] Function '%s' found at address: 0x%X\n", funcName, funcAddr)

	// Call the function using syscall.SyscallN
	// SyscallN is used to call a function pointer when the number of arguments is known at compile time.
	// For LaunchCalc(), which takes no arguments (BOOL LaunchCalc()), we call it like this:
	// SyscallN(functionAddress, argCount, arg1, arg2, ...)
	// Here, argCount is 0. We pass uintptr(0) for unused arguments.
	fmt.Printf("[+] Calling function '%s'...\n", funcName)
	// The first return value 'ret' holds the function's return value (BOOL as uintptr: 1 for TRUE, 0 for FALSE).
	// The second return value is reserved (usually 0 on success).
	// The third return value 'callErr' holds any error from the syscall itself (e.g., access violation).
	ret, _, callErr := syscall.SyscallN(funcAddr, 0, 0, 0, 0)

	// Check the error returned by the syscall mechanism itself.
	// callErr != 0 indicates a problem during the call setup or execution (like invalid address).
	// Note: '0' corresponds to ERROR_SUCCESS in Windows syscalls.
	if callErr != 0 {
		log.Fatalf("[-] Error occurred during syscall to '%s': %v\n", funcName, callErr)
	}

	// Check the actual return value of the LaunchCalc function.
	// Our DLL function returns TRUE (1) on success, FALSE (0) on failure.
	if ret != 0 { // Corresponds to TRUE
		fmt.Printf("[+] Function '%s' executed successfully (returned TRUE).\n", funcName)
	} else { // Corresponds to FALSE
		// This might happen if VirtualAlloc failed inside the DLL, for example.
		fmt.Printf("[-] Function '%s' execution reported failure (returned FALSE).\n", funcName)
	}

	fmt.Println("[+] Loader finished.")
}
```

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "create_dll.md" >}})
[|NEXT|]({{< ref "../module02/structure.md" >}})