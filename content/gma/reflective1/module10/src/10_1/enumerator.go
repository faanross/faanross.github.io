//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe" // Import unsafe for Sizeof

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
