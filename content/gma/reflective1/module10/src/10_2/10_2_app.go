//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe" // Keep for pointer conversions

	// ADD the new win32api package
	"github.com/zzl/go-win32api/v2/win32"

	// KEEP x/sys/windows for OpenProcess, Handle, Toolhelp, Write/Read/Protect Memory, UTF16ToString etc.
	// We will use its Handle type and functions other than Alloc/FreeEx.
	"golang.org/x/sys/windows"
)

// findProcessPID function remains the same (uses golang.org/x/sys/windows)
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

	// --- Change Remote Memory Protection (Keep using windows package) ---
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

	// --- Attempt Second Write (Keep using windows package) ---
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
