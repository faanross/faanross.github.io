---
showTableOfContents: true
title: "Finding and Opening Target Processes (Lab 10.1)"
type: "page"
---
## Goal
In this lab we'll apply what we discussed in the previous lesson to create a standalone "process enumerator" in Go.

Specifically, our program will:
1.  Enumerate currently running processes using the `Toolhelp` Snapshot API (`CreateToolhelp32Snapshot`, `Process32FirstW`, `Process32NextW`).
2.  Print the Process ID (PID) and executable name for each process found.
3.  Take a target process name as a command-line argument.
4.  Find the PID of the first process matching the target name.
5.  Attempts to open the target process using `OpenProcess` with access rights suitable for common injection techniques (`PROCESS_VM_OPERATION`, `PROCESS_VM_WRITE`, `PROCESS_VM_READ`, `PROCESS_CREATE_THREAD`, `PROCESS_QUERY_INFORMATION`).
6.  Report whether opening the process handle was successful and prints the handle value or the error encountered.

## Code
Note: This is a standalone application, we'll integrate this logic into our overall project later.

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

	"golang.org/x/sys/windows"
)

// findProcessPID uses Toolhelp snapshot to find the PID of the first process matching targetName.
func findProcessPID(targetName string) (uint32, error) {
	fmt.Printf("[*] Searching for process: %s\n", targetName)

	// Create a snapshot of current processes
	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	// Ensure snapshot handle is closed eventually
	defer windows.CloseHandle(handle)

	// Initialize PROCESSENTRY32W struct. dwSize MUST be set.
	var entry windows.ProcessEntry32
	// Use unsafe.Sizeof for struct size
	entry.Size = uint32(unsafe.Sizeof(entry)) // <--- FIX 1: Use unsafe.Sizeof

	// Get the first process
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return 0, fmt.Errorf("Process32First failed: %w", err)
	}

	// Loop through processes
	for {
		// Convert process name (WCHAR array) to Go string
		processName := windows.UTF16ToString(entry.ExeFile[:])
		// fmt.Printf("  PID: %d, Name: %s\n", entry.ProcessID, processName) // Optional: Print all processes

		// Case-insensitive comparison
		if strings.EqualFold(processName, targetName) {
			fmt.Printf("[+] Found target process '%s' with PID: %d\n", targetName, entry.ProcessID)
			return entry.ProcessID, nil // Return the found PID
		}

		// Get the next process
		err = windows.Process32Next(handle, &entry)
		if err != nil {
			// ERROR_NO_MORE_FILES is expected when the loop finishes
			if err == windows.ERROR_NO_MORE_FILES {
				break // End of process list
			}
			// Otherwise, it's an unexpected error
			return 0, fmt.Errorf("Process32Next failed: %w", err)
		}
	}

	// If loop finishes without finding the process
	return 0, fmt.Errorf("process '%s' not found", targetName)
}

func main() {
	fmt.Println("[+] Process Enumeration and Handle Acquisition Tool")

	// --- Argument Check ---
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <target_process_name.exe>\n", os.Args[0])
		fmt.Println("Example: .\\process_finder.exe notepad.exe")
		return
	}
	targetProcessName := os.Args[1]

	// --- Find Target PID ---
	targetPID, err := findProcessPID(targetProcessName)
	if err != nil {
		log.Fatalf("[-] Failed to find PID: %v", err)
	}
	if targetPID == 0 {
		// Should be caught by the error above, but double-check
		log.Fatalf("[-] Process '%s' not found.", targetProcessName)
	}

	// --- Define Desired Access Rights ---
	// Combine flags needed for typical injection
	desiredAccess := uint32(windows.PROCESS_CREATE_THREAD |
		windows.PROCESS_QUERY_INFORMATION |
		windows.PROCESS_VM_OPERATION |
		windows.PROCESS_VM_WRITE |
		windows.PROCESS_VM_READ)

	fmt.Printf("[*] Attempting to open process PID %d with access rights: 0x%X\n", targetPID, desiredAccess)

	// --- Open Target Process ---
	// windows.OpenProcess wraps the OpenProcess API call
	hProcess, err := windows.OpenProcess(desiredAccess, false, targetPID)
	// bInheritHandle = false

	if err != nil {
		// Check for specific common errors using constants from the 'windows' package
		// We compare the error directly with known windows error values
		if err == windows.ERROR_ACCESS_DENIED { // <--- FIX 2a: Use windows.ERROR_ACCESS_DENIED
			log.Printf("[-] OpenProcess failed: Access Denied (Error 5). Insufficient privileges?")
		} else if err == windows.ERROR_INVALID_PARAMETER { // <--- FIX 2b: Use windows.ERROR_INVALID_PARAMETER
			log.Printf("[-] OpenProcess failed: Invalid Parameter (Error 87). PID %d might no longer exist.", targetPID)
		} else {
			log.Printf("[-] OpenProcess failed: %v", err)
		}
		return // Exit if opening failed
	}

	// --- Success ---
	fmt.Printf("[+] Successfully obtained handle to process PID %d.\n", targetPID)
	fmt.Printf("[+] Handle Value: 0x%X\n", hProcess)

	// --- Cleanup ---
	// IMPORTANT: Always close the handle when done
	defer func() {
		fmt.Println("[*] Closing process handle...")
		errClose := windows.CloseHandle(hProcess)
		if errClose != nil {
			log.Printf("[!] Warning: Failed to close process handle: %v", errClose)
		} else {
			fmt.Println("[+] Process handle closed.")
		}
	}()

	// TODO: In future labs, use hProcess for injection steps here...
	fmt.Println("[*] Handle obtained. (Injection steps would follow here)")
	// Keep alive briefly to observe handle, etc.
	fmt.Println("Press Enter to close handle and exit...")
	fmt.Scanln()
}

````

## Code Breakdown
Let's break down the primary logic responsible for the 6 steps as we've defined them in our `Goal` section above.
- **Process Enumeration (Toolhelp Snapshot):** Our `findProcessPID` function calls `windows.CreateToolhelp32Snapshot` to get a snapshot of running processes, then iterates through them using `windows.Process32First` and `windows.Process32Next`.
- **Print Process Details:** Inside the loop within `findProcessPID`, process information (`entry.ProcessID` and `entry.ExeFile` converted via `windows.UTF16ToString`) is accessed. (Note: Printing _all_ processes is currently commented out, feel free to uncomment this if you'd like to enumerate all running processes).
- **Command-Line Argument:** The `main` function checks `os.Args` to retrieve the target process name provided by the us.
- **Find Target PID:** The `findProcessPID` function compares process names (case-insensitively using `strings.EqualFold`) against the command-line argument within its loop, returning the `entry.ProcessID` upon finding a match.
- **Open Target Process:** Back in `main`, after obtaining the `targetPID`, our code defines `desiredAccess` flags and attempts to get a handle using `windows.OpenProcess`.
- **Report Outcome:** The `main` function checks the error returned by `windows.OpenProcess`. It prints the obtained `hProcess` value on success or logs specific errors (like access denied or invalid parameter) on failure. The `defer` block ensures `windows.CloseHandle` is called eventually


## Instructions

Compile the code using `go build`.
```shell
GOOS=windows GOARCH=amd64 go build -o process_finder.exe process_finder.go
```

Then transfer it to target system.

Before executing you should know which process you are going to search for, in my example I will first open `notepad.exe`, and we'll look for that.

```shell
.\process_finder.exe notepad.exe
```

Let's also attempt to enumerate a process we won't have sufficient privileges for (`lsass.exe`), as well as a process that does not exist (`nosuchprocess.exe`).

```shell
.\process_finder.exe lsass.exe
```


```shell
.\process_finder.exe nosuchprocess.exe
```


## Results
Enumerating notepad.exe should yield both a PID, as well as a handle to the process.

```shell
PS C:\Users\vuilhond\Desktop> .\process_finder.exe notepad.exe
[+] Process Enumeration and Handle Acquisition Tool
[*] Searching for process: notepad.exe
[+] Found target process 'notepad.exe' with PID: 4536
[*] Attempting to open process PID 4536 with access rights: 0x43A
[+] Successfully obtained handle to process PID 4536.
[+] Handle Value: 0x164
[*] Handle obtained. (Injection steps would follow here)
```

We can use a tool like [System Informer](https://systeminformer.sourceforge.io/downloads) to confirm our results. Under the list of active processes, we can confirm that the PID of `notepad.exe` is indeed `4536`.

![find_lab](../img/find_lab_a.png)

Let's also confirm the handle is correct. In the list of processes, find `process_finder.exe` (or whatever you named your enumerating application), and double-click on it. Find the `Handles` tab, and then at the top right-click on the column header, select `Choose columns...`. On the LHS select `Handle`, then click on `Show >`, this should transfer it over to the RHS, you can then click OK.

![find_lab](../img/find_lab_a.png)

You should be able to find the handle to `notepad.exe` in your list, assuming you did not hit enter in the console to close handle yet. Here we can now confirm the value corresponds to that displayed by our application, in my case `0x164`.

Further, attempting to enumerate `lsass.exe` should give us an `Access Denied (Error 5)`.

```shell
PS C:\Users\vuilhond\Desktop> .\process_finder.exe lsass.exe
[+] Process Enumeration and Handle Acquisition Tool
[*] Searching for process: lsass.exe
[+] Found target process 'lsass.exe' with PID: 852
[*] Attempting to open process PID 852 with access rights: 0x43A
2025/05/01 10:13:06 [-] OpenProcess failed: Access Denied (Error 5). Insufficient privileges?
```

And, attempting to enumerate a non-existent process will also fail.

```shell
PS C:\Users\vuilhond\Desktop> .\process_finder.exe nosuchprocess.exe
[+] Process Enumeration and Handle Acquisition Tool
[*] Searching for process: nosuchprocess.exe
2025/05/01 10:13:33 [-] Failed to find PID: process 'nosuchprocess.exe' not found
```



## Discussion

This lab demonstrates the fundamental steps of locating a target process by name using the `Toolhelp` Snapshot API and attempting to acquire a handle with specific permissions using `OpenProcess`.

The outcome of `OpenProcess` is critical. Successfully obtaining a handle (like `hProcess` in the code) confirms that our current process has the requested permissions to interact with the target process. This handle is essential for all subsequent WinAPI injection steps (`VirtualAllocEx`, `WriteProcessMemory`, `CreateRemoteThread`, etc.).

Failures, particularly "Access Denied," highlight the importance of process privileges and security boundaries in Windows. We generally cannot open highly privileged system processes with full access rights unless our loader process is also running with sufficient privileges (e.g., as Administrator or SYSTEM, often requiring `SeDebugPrivilege` to be enabled). Choosing an appropriate target process that matches the privilege level of our loader (or one we can elevate to match) is a key consideration for successful injection.

## Conclusion

We now have a functional Go tool to enumerate processes and obtain a handle to a specific target process with the necessary rights for injection. This handle is the key prerequisite for the next steps: allocating and writing memory within the target process's address space.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "process.md" >}})
[|NEXT|]({{< ref "mem.md" >}})