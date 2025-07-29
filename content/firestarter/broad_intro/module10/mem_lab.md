---
showTableOfContents: true
title: "Performing Remote Memory Operations (Lab 10.2)"
type: "page"
---

## Goal

Having learned how to obtain a handle to a target process in our previous lab, we'll now interact with that process's memory space. Based on the theory from our previous lesson, this involves using specific WinAPI functions designed for cross-process memory manipulation.

Specifically, in this lab we'll build on our solution from Lab 10.1, so that by the end our application it will be capable of:
1.  Finding a target process by name and obtain a handle (`hProcess`) with necessary permissions (including `PROCESS_VM_READ` this time). (**From Lab 10.1**)
2.  Allocate ReadWrite (RW) memory in the target process using `VirtualAllocEx`.
3.  Write a sample string into the allocated remote memory using `WriteProcessMemory`.
4.  Read the string back from the remote memory using `ReadProcessMemory` to verify the write.
5.  Change the remote memory's protection to ReadOnly (`PAGE_READONLY`) using `VirtualProtectEx`.
6.  *Attempt* to write to the memory region again (now ReadOnly) using `WriteProcessMemory`, demonstrating the protection change (expecting failure).
7.  Clean up by freeing the allocated remote memory using `VirtualFreeEx`.


## Code
### A Quick Note on Using 3rd Party Libraries
In developing the code for this lab, I encountered a couple of hurdles, which I thought would have some value in sharing here instead of just showing you the final product.

When I originally wrote the code I assumed that `VirtualAllocEx` and `VirtualFreeEx`, like almost all the other common win32 API functions, would be found in `golang.org/x/sys/windows`. However it soon became clear from compilation errors, and the lack of effect that dependency tidying and cache invalidation had, that something was amiss. (NOTE: Since I work on Mac OS but do these applications with Windows build tags, my IDE suppresses any errors. Just mentioning that in case you are curious why I had to wait for the compiler to point the issue out to me. )

And so I went to the official package docs - `https://pkg.go.dev/golang.org/x/sys/windows` - and searched for these 2 functions, only to come up empty handed. In other words, `VirtualAllocEx` and `VirtualFreeEx` are not in the `golang.org/x/sys/windows` library. Now I have no idea whether they were there before but later removed, or if they were never there in the first place, I had just assumed, since most common win32 API functions reside there that they would too.

In any case, we do need them, so that leaves us with 2 options - use CGO (which would allow us to use any Windows function, but is considerably more complex), or find another library candidate. There are pros and cons to both approaches, but for now I would definitely prefer to find another library candidate before resorting to CGO.

To find which package has a specific function you are looking for you can use the search bar right at the top of `pkg.go.dev` that says `Search Packages or Symbols`, just search for `VirtualAllocEx`. Note that you'll immediately come up empty-handed since it defaults to looking for packages, and well there is no package with this name. So change the selected tab to `Symbols`, and then you should see a number of results.

And I actually made another error of poor judgement (or just complacency) here, which I'll also share. The results of the search are ranked in order of total imports - i.e. how popular a library is **over it's lifetime**.  So I opted for `github.com/AllenDang/w3`, since that ranked at the top, with no further consideration. One can find the function signatures on the package page, so you can then go and adapt it in your code.

I thought this was all to it, and then when I compiled this error came up:

```shell
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildvcs=false

# github.com/AllenDang/w32

/go/pkg/mod/github.com/!allen!dang/w32@v0.0.0-20180428130237-ad0a36d80adc/user32.go:1040:3: cannot use flag (variable of type uint32) as uintptr value in argument to procRedrawWindow.Call
```

Now this is a compilation error of a completely different nature since it's not an issue with my code - there's an issue with the package. And how do we know this? The error `user32.go:1040:3: cannot use flag (variable of type uint32) as uintptr value...` is happening _inside the source code of the `github.com/AllenDang/w32` library itself_, specifically in the `user32.go` file.

This means that it's the package code that is breaking. I then went back to the original search results and saw that the package has not been updated since April 18, 2018.

![mem_lab](../img/mem_lab_a.png)

This is a big deal, since I am currently using the latest version of Go (1.23), and Go's internal `syscall` mechanisms and type requirements have evolved significantly since then (especially around Go 1.17 and later). Code written for Go versions pre-2018 might have relied on implicit conversions that are no longer allowed or where function signatures have changed slightly in the underlying Go runtime/syscall implementation.

All to say, this package code is likely no longer compatible with more modern versions of Go. So back to the search results, it's clear we should not just rely on import total, but also when last it's been updated. Looking at the next two results:
- The one from `TheTitanrain` has 45 imports, but has not been updated in 5 years, so it's probably been abandoned.
- The next one from `zzl` has slightly less imports (35), but has been updated last year, so is most likely still actively being maintained.

So, in this case I opted for it, and it worked.

Do note however that if we were creating code that's critical and we wanted to be assured that we're not relying on a package that the maintainer might lose interest in in a few years, we'd probably have to opt for CGO. Doing the hard work up front, but future-proofing our code. This is however not one of those cases, and so I'm happy to take a bit of a risk. We will however definitely cover CGO in the future since it not only ensures we don't rely on library code, but sometimes the function we want to use might not exists in a library at all, in which case we have to use CGO.

OK, with all that - let's get back to the actual code.

### Updated Code

```go
//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe" 

	"github.com/zzl/go-win32api/v2/win32"

	"golang.org/x/sys/windows"
)

// This is our function from Lab 10.1
func findProcessPID(targetName string) (uint32, error) {
	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(handle, &entry)
	if err != nil {
		return 0, err
	}

	for {
		processName := windows.UTF16ToString(entry.ExeFile[:])
		if strings.EqualFold(processName, targetName) {
			fmt.Printf("[+] Found target process '%s' with PID: %d\n", targetName, entry.ProcessID)
			return entry.ProcessID, nil
		}
		err = windows.Process32Next(handle, &entry)
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return 0, err
		}
	}
	return 0, fmt.Errorf("process '%s' not found", targetName)
}

func main() {
	fmt.Println("[+] Remote Memory Operations Tool")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <target_process_name.exe>\n", os.Args[0])
		return
	}
	targetProcessName := os.Args[1]

	targetPID, err := findProcessPID(targetProcessName)
	if err != nil {
		log.Fatalf("[-] Failed to find PID: %v", err)
	}

	// --- Define Access Rights (using windows package constants) ---
	desiredAccess := uint32(windows.PROCESS_CREATE_THREAD |
		windows.PROCESS_QUERY_INFORMATION |
		windows.PROCESS_VM_OPERATION |
		windows.PROCESS_VM_WRITE |
		windows.PROCESS_VM_READ)

	fmt.Printf("[*] Attempting to open process PID %d with access rights: 0x%X\n", targetPID, desiredAccess)

	// --- Open Target Process (using windows package) ---
	hProcess, err := windows.OpenProcess(desiredAccess, false, targetPID)
	if err != nil {
		log.Fatalf("[-] OpenProcess failed: %v", err)
	}
	fmt.Printf("[+] Successfully obtained handle: 0x%X\n", hProcess)
	// Ensure handle is closed eventually (using windows package)
	defer windows.CloseHandle(hProcess)

	// --- Allocate Memory Remotely (using zzl/go-win32api) ---
	const memSize = 1024 // Allocate 1KB for example
	fmt.Printf("[*] Allocating %d bytes in target process (RW)...\n", memSize)

	// Use win32.VirtualAllocEx from zzl/go-win32api
	remoteAddrPtr, errCode := win32.VirtualAllocEx(
		win32.HANDLE(hProcess),             // Cast windows.Handle to win32.HANDLE
		nil,                                // Let system choose address (pass nil for unsafe.Pointer)
		uintptr(memSize),                   // Size
		win32.MEM_COMMIT|win32.MEM_RESERVE, // Allocation type from win32 pkg
		win32.PAGE_READWRITE,               // Protection flags from win32 pkg
	)

	// Error check: uses returned WIN32_ERROR code
	if errCode != win32.NO_ERROR {
		log.Fatalf("[-] win32.VirtualAllocEx failed (Error code: %d)", errCode)
	}
	if remoteAddrPtr == nil { // Safety check
		log.Fatalf("[-] win32.VirtualAllocEx succeeded (NO_ERROR) but returned nil address.")
	}

	// Convert unsafe.Pointer to uintptr for general use (printing, passing to windows funcs)
	remoteAddrUintptr := uintptr(remoteAddrPtr)

	fmt.Printf("[+] Allocated %d bytes at remote address: 0x%X\n", remoteAddrUintptr, remoteAddrUintptr) // Use uintptr for printing

	// Ensure memory is freed eventually (using zzl/go-win32api in defer)
	defer func(addrToFreeUintptr uintptr) { // Pass uintptr address to defer
		if addrToFreeUintptr == 0 {
			return // Don't try to free if allocation failed
		}
		fmt.Println("[*] Freeing remote memory...")

		// Use win32.VirtualFreeEx from zzl/go-win32api
		// Convert uintptr address back to unsafe.Pointer for the call
		_, errCode := win32.VirtualFreeEx(
			win32.HANDLE(hProcess),            // Cast handle
			unsafe.Pointer(addrToFreeUintptr), // Address to free
			0,                                 // dwSize (must be 0 for MEM_RELEASE)
			win32.MEM_RELEASE,                 // Free type from win32 pkg
		)

		// Error check: uses returned WIN32_ERROR code
		if errCode != win32.NO_ERROR {
			log.Printf("[!] Warning: win32.VirtualFreeEx failed (Error code: %d)", errCode)
		} else {
			fmt.Println("[+] Remote memory freed.")
		}
	}(remoteAddrUintptr) // Pass the uintptr address to the deferred function

	// --- Write to Remote Memory (Keep using windows package) ---
	dataToWrite := []byte("Hello from remote process! \\o/\000")
	var bytesWritten uintptr
	fmt.Printf("[*] Writing %d bytes to remote address 0x%X...\n", len(dataToWrite), remoteAddrUintptr)                      // Use uintptr address
	err = windows.WriteProcessMemory(hProcess, remoteAddrUintptr, &dataToWrite[0], uintptr(len(dataToWrite)), &bytesWritten) // Use uintptr address
	if err != nil {
		log.Fatalf("[-] WriteProcessMemory failed: %v", err)
	}
	if bytesWritten != uintptr(len(dataToWrite)) {
		log.Fatalf("[-] WriteProcessMemory: incomplete write (%d/%d bytes)", bytesWritten, len(dataToWrite))
	}
	fmt.Printf("[+] Successfully wrote %d bytes.\n", bytesWritten)

	// --- Read Back from Remote Memory (Keep using windows package) ---
	readBuffer := make([]byte, len(dataToWrite))
	var bytesRead uintptr
	fmt.Printf("[*] Reading %d bytes back from remote address 0x%X...\n", len(readBuffer), remoteAddrUintptr)          // Use uintptr address
	err = windows.ReadProcessMemory(hProcess, remoteAddrUintptr, &readBuffer[0], uintptr(len(readBuffer)), &bytesRead) // Use uintptr address
	if err != nil {
		log.Fatalf("[-] ReadProcessMemory failed: %v", err)
	}
	if bytesRead != uintptr(len(readBuffer)) {
		log.Fatalf("[-] ReadProcessMemory: incomplete read (%d/%d bytes)", bytesRead, len(readBuffer))
	}
	fmt.Printf("[+] Successfully read %d bytes: \"%s\"\n", bytesRead, string(readBuffer))
	// Verify content
	if string(readBuffer) != string(dataToWrite) {
		log.Println("[!] Warning: Read data does not match written data!")
	} else {
		fmt.Println("[+] Read data verification successful.")
	}

	// --- Change Remote Memory Protection using windows package ---
	var oldProtect uint32
	// Use constant from windows package as input to windows.VirtualProtectEx
	newProtect := uint32(windows.PAGE_READONLY)
	fmt.Printf("[*] Changing protection of remote address 0x%X to PAGE_READONLY (0x%X)...\n", remoteAddrUintptr, newProtect) // Use uintptr address
	// Use windows package for VirtualProtectEx
	err = windows.VirtualProtectEx(hProcess, remoteAddrUintptr, uintptr(memSize), newProtect, &oldProtect) // Use uintptr address
	if err != nil {
		log.Fatalf("[-] VirtualProtectEx failed: %v", err)
	}
	fmt.Printf("[+] Protection changed successfully. Old protection was: 0x%X\n", oldProtect)

	// --- Attempt Second Write using windows package ---
	secondData := []byte("Attempting second write...\000")
	fmt.Printf("[*] Attempting to write again to remote address 0x%X (should fail)...\n", remoteAddrUintptr) // Use uintptr address
	// Use windows package for WriteProcessMemory
	err = windows.WriteProcessMemory(hProcess, remoteAddrUintptr, &secondData[0], uintptr(len(secondData)), &bytesWritten) // Use uintptr address
	if err != nil {
		fmt.Printf("[+] WriteProcessMemory failed as expected after changing protection: %v\n", err)
	} else {
		log.Printf("[!] Warning: WriteProcessMemory succeeded unexpectedly after setting PAGE_READONLY!")
	}

	// --- Final Cleanup (handled by defers) ---
	fmt.Println("[+] Lab complete.")
}

```



## Code Breakdown
**`findProcessPID` Function:**
- This is our function from Lab 10.1, it uses functions from `golang.org/x/sys/windows` (`CreateToolhelp32Snapshot`, `Process32First`, `Process32Next`, `CloseHandle`) to find and return the target process PID.


**Access Rights:**
- Defines `desiredAccess` using constants from `golang.org/x/sys/windows` (`PROCESS_VM_READ`, `VM_OPERATION`, `VM_WRITE`, etc.).

**`OpenProcess`:**
- Opens the target process using `windows.OpenProcess` with the specified rights, obtaining `hProcess` (a `windows.Handle`). A `defer windows.CloseHandle(hProcess)` ensures the handle is closed on exit.

**`VirtualAllocEx` (using `zzl/go-win32api`):**
- Calls `win32.VirtualAllocEx` from the `github.com/zzl/go-win32api/v2/win32` package.
- Passes the process handle cast to `win32.HANDLE(hProcess)`, `nil` for the address (`unsafe.Pointer`), the `memSize` (`uintptr`), allocation type constants from the `win32` package (`win32.MEM_COMMIT | win32.MEM_RESERVE`), and initial protection flags from the `win32` package (`win32.PAGE_READWRITE`).
- Receives the remote base address as `remoteAddrPtr` (`unsafe.Pointer`) and an error code `errCode` (`win32.WIN32_ERROR`).
- Checks for allocation failure by comparing `errCode` with `win32.NO_ERROR`.
- Converts the `unsafe.Pointer` address to `uintptr` (`remoteAddrUintptr`) for easier use with other functions and printing.
- Adds a `defer` statement containing `win32.VirtualFreeEx` (see next point) to ensure remote memory is freed.

**`VirtualFreeEx` (within `defer`, using `zzl/go-win32api`):**
- The deferred function calls `win32.VirtualFreeEx`.
- Passes `win32.HANDLE(hProcess)`, the address cast back to `unsafe.Pointer(remoteAddrUintptr)`, `0` for size, and the free type constant `win32.MEM_RELEASE`.
- Checks the returned `errCode` against `win32.NO_ERROR` to determine success or failure.

**`WriteProcessMemory` (using `golang.org/x/sys/windows`):**
- Defines sample `dataToWrite`.
- Calls `windows.WriteProcessMemory` using `hProcess`, the `remoteAddrUintptr`, a pointer to the data, data length, and a pointer to `bytesWritten`.
- Performs error checking based on the returned `error` and verifies the number of bytes written.


**`ReadProcessMemory` (using `golang.org/x/sys/windows`):**
- Creates local `readBuffer`.
- Calls `windows.ReadProcessMemory` using `hProcess`, `remoteAddrUintptr`, a pointer to the buffer, buffer length, and a pointer to `bytesRead`.
- Performs error checking and verifies bytes read, comparing read data to original data.


**`VirtualProtectEx` (using `golang.org/x/sys/windows`):**
- Calls `windows.VirtualProtectEx` using `hProcess`, `remoteAddrUintptr`, `memSize` (cast to `uintptr`), the new protection constant `windows.PAGE_READONLY`, and a pointer to `oldProtect`.
- Checks the returned `error`.

**Second Write Attempt (using `golang.org/x/sys/windows`):**
- Calls `windows.WriteProcessMemory` again to the same `remoteAddrUintptr`.
- Expects this call to fail due to the memory now being read-only. Error handling confirms that an error (`err != nil`) _is_ expected.



## Instructions

Remember to add the third library manually using `go get`(see below), or `go mod tidy` should handle it automatically.

```shell
go get github.com/zzl/go-win32api/v2/win32
```


Once again compile it for Windows, since I'm working on Mac OS I will use:

```shell
GOOS=windows GOARCH=amd64 go build -buildvcs=false
```


I'll once again use notepad.exe, so I'll open that first, and then run:
```shell
.\proc_mem_write.exe notepad.exe
```


## Results

```go
PS C:\Users\vuilhond\Desktop> .\proc_mem_write.exe notepad.exe
[+] Remote Memory Operations Tool
[+] Found target process 'notepad.exe' with PID: 9412
[*] Attempting to open process PID 9412 with access rights: 0x43A
[+] Successfully obtained handle: 0x16C
[*] Allocating 1024 bytes in target process (RW)...
[+] Allocated 3115692654592 bytes at remote address: 0x2D56DC10000
[*] Writing 31 bytes to remote address 0x2D56DC10000...
[+] Successfully wrote 31 bytes.
[*] Reading 31 bytes back from remote address 0x2D56DC10000...
[+] Successfully read 31 bytes: "Hello from remote process! \o/"
[+] Read data verification successful.
[*] Changing protection of remote address 0x2D56DC10000 to PAGE_READONLY (0x2)...
[+] Protection changed successfully. Old protection was: 0x4
[*] Attempting to write again to remote address 0x2D56DC10000 (should fail)...
[+] WriteProcessMemory failed as expected after changing protection: Invalid access to memory location.
[+] Lab complete.
[*] Freeing remote memory...
[+] Remote memory freed.
```

- We can see that is with the previous lab, a handle to `notepad.exe` is successfully obtained.
- Memory is allocated in `notepad.exe`'s address space.
- Our string is then successfully written and read back correctly.
- `VirtualProtectEx` successfully changes the protection to `PAGE_READONLY` (0x2) from the original `PAGE_READWRITE` (0x4).
- The second `WriteProcessMemory` call **fails**, likely with an "Access is denied" error, confirming the memory is no longer writable.
- Cleanup occurs successfully.

## Discussion

This lab successfully demonstrates the use of `VirtualAllocEx`, `WriteProcessMemory`, `ReadProcessMemory`, and `VirtualProtectEx` to manipulate the memory of another process. We allocated memory, wrote data to it, verified the write by reading it back, and successfully changed its permissions to prevent further writes.

This sequence forms the core preparation phase for many remote process injection techniques. Instead of writing a simple string, we would write our shellcode to the `remoteAddr`. Instead of changing the protection to `PAGE_READONLY`, we would change it to `PAGE_EXECUTE_READ` (RX).

## Conclusion

We have now mastered the WinAPI functions required to allocate, write, read, and modify the protection of memory in a remote process using Go. This foundational capability allows us to place our payload precisely where we need it in the target's address space and set the stage for its execution.

The next logical step is now to actually trigger the execution of the code residing at `remoteAddr` within the target process.


---
