//go:build windows
// +build windows

package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"log"
	"syscall"
	"unsafe" // Required for pointer conversions
)

// osVersionInfoExW is the structure for GetVersionExW using the syscall package.
// Note the fixed size array for szCSDVersion.
type osVersionInfoExW struct {
	dwOSVersionInfoSize uint32
	dwMajorVersion      uint32
	dwMinorVersion      uint32
	dwBuildNumber       uint32
	dwPlatformId        uint32
	szCSDVersion        [128]uint16 // WCHAR szCSDVersion[128]
	wServicePackMajor   uint16
	wServicePackMinor   uint16
	wSuiteMask          uint16
	wProductType        byte
	wReserved           byte
}

func main() {
	fmt.Println("--- Using syscall package ---")

	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		log.Fatalf("Failed to load kernel32.dll: %v", err)
	}
	defer kernel32.Release()

	ntdll, err := syscall.LoadDLL("ntdll.dll")
	if err != nil {
		log.Printf("Warning: Failed to load ntdll.dll: %v. Some functions might not be found if only in ntdll.", err)
	}
	if ntdll != nil {
		defer ntdll.Release()
	}

	// --- GetCurrentProcessId (via GetCurrentProcess + GetProcessId) ---
	getCurrentProcess, err := kernel32.FindProc("GetCurrentProcess")
	if err != nil {
		log.Fatalf("Failed to find GetCurrentProcess: %v", err)
	}
	hProcessVal, _, callErrnoGCP := getCurrentProcess.Call()
	// Check the syscall error code. ERROR_SUCCESS (0) means no error from GetLastError.
	if callErrnoGCP != syscall.ERROR_SUCCESS {
		log.Fatalf("GetCurrentProcess syscall failed: %s", callErrnoGCP.Error())
	}
	// For GetCurrentProcess, hProcessVal itself is a pseudo-handle and not typically 0 on failure.
	// The primary error indication is callErrnoGCP.

	getProcessId, err := kernel32.FindProc("GetProcessId")
	if err != nil {
		log.Fatalf("Failed to find GetProcessId: %v", err)
	}
	pidUintptr, _, callErrnoGPID := getProcessId.Call(hProcessVal)
	if callErrnoGPID != syscall.ERROR_SUCCESS {
		// Log the error from GetLastError
		log.Printf("GetProcessId syscall returned error: %s", callErrnoGPID.Error())
	}
	// Additionally, the GetProcessId API function returns 0 on failure.
	if pidUintptr == 0 {
		// This means the API call failed, even if callErrnoGPID was ERROR_SUCCESS (which would be unusual).
		// If callErrnoGPID was already logged, this is supplemental.
		if callErrnoGPID == syscall.ERROR_SUCCESS { // Only log this if no syscall error was reported
			log.Printf("GetProcessId call returned PID 0 (API failure), but syscall error was ERROR_SUCCESS.")
		} else {
			log.Printf("GetProcessId call returned PID 0 (API failure).") // Syscall error already logged
		}
	}
	pid := uint32(pidUintptr)
	fmt.Printf("syscall.GetProcessId(GetCurrentProcess()): %d\n", pid)

	// --- GetCurrentThreadId ---
	getCurrentThreadId, err := kernel32.FindProc("GetCurrentThreadId")
	if err != nil {
		log.Fatalf("Failed to find GetCurrentThreadId: %v", err)
	}
	tidUintptr, _, callErrnoGTID := getCurrentThreadId.Call()
	if callErrnoGTID != syscall.ERROR_SUCCESS {
		log.Printf("GetCurrentThreadId syscall returned error: %s", callErrnoGTID.Error())
	}
	// GetCurrentThreadId API itself does not typically fail by returning 0.
	tid := uint32(tidUintptr)
	fmt.Printf("syscall.GetCurrentThreadId(): %d\n", tid)

	// --- GetVersionExW ---
	var procGetVersionExW *syscall.Proc
	var findProcErr error

	procGetVersionExW, findProcErr = kernel32.FindProc("GetVersionExW")
	if findProcErr != nil {
		log.Printf("GetVersionExW not found in kernel32.dll (error: %v).", findProcErr)
		if ntdll != nil {
			log.Println("Attempting to find GetVersionExW in ntdll.dll...")
			procGetVersionExW, findProcErr = ntdll.FindProc("GetVersionExW")
			if findProcErr != nil {
				log.Printf("Failed to find GetVersionExW in ntdll.dll as well: %v", findProcErr)
			}
		} else {
			log.Println("ntdll.dll was not loaded, cannot attempt fallback for GetVersionExW there.")
		}
	}

	if findProcErr == nil && procGetVersionExW != nil {
		var osInfo osVersionInfoExW
		osInfo.dwOSVersionInfoSize = uint32(unsafe.Sizeof(osInfo))

		retGVE, _, callErrnoGVE := procGetVersionExW.Call(uintptr(unsafe.Pointer(&osInfo)))
		// GetVersionExW API returns 0 on failure.
		if retGVE == 0 {
			log.Printf("GetVersionExW call failed (API returned 0).")
			if callErrnoGVE != syscall.ERROR_SUCCESS {
				log.Printf("  GetLastError reported: %s", callErrnoGVE.Error())
			} else {
				log.Printf("  GetLastError reported ERROR_SUCCESS (this is unusual for GetVersionExW failure).")
			}
		} else { // Success (retGVE is non-zero)
			// Optionally check callErrnoGVE even on success, though it should be ERROR_SUCCESS.
			if callErrnoGVE != syscall.ERROR_SUCCESS {
				log.Printf("GetVersionExW call succeeded (API returned non-zero) but GetLastError reported: %s (unusual).", callErrnoGVE.Error())
			}
			fmt.Printf("syscall.GetVersionExW:\n")
			fmt.Printf("  Major Version: %d\n", osInfo.dwMajorVersion)
			fmt.Printf("  Minor Version: %d\n", osInfo.dwMinorVersion)
			fmt.Printf("  Build Number: %d\n", osInfo.dwBuildNumber)
			csdString := syscall.UTF16ToString(osInfo.szCSDVersion[:])
			fmt.Printf("  Service Pack: %s\n", csdString)
			fmt.Printf("  Product Type: %d\n", osInfo.wProductType)
		}
	} else {
		log.Println("Skipping syscall.GetVersionExW due to errors finding the procedure.")
	}

	fmt.Println("\n--- Using golang.org/x/sys/windows ---")
	xSysWindows()
}

func xSysWindows() {
	// --- GetCurrentProcessId ---
	currentProcessHandle := windows.CurrentProcess()
	pid, err := windows.GetProcessId(currentProcessHandle)
	// The windows.GetProcessId wrapper returns an error value.
	// The underlying API GetProcessId returns 0 on failure.
	if err != nil {
		log.Printf("windows.GetProcessId failed with error: %v", err)
		// If err is not nil, pid might be 0 anyway.
	} else if pid == 0 {
		// This case implies the API call returned 0, but the Go wrapper didn't return an error.
		// This would be specific to the wrapper's implementation.
		log.Println("windows.GetProcessId returned PID 0 without an explicit error from the wrapper (implies API failure).")
	}

	if pid != 0 { // Only print if we believe it's a valid PID
		fmt.Printf("windows.GetProcessId(windows.CurrentProcess()): %d\n", pid)
	} else {
		fmt.Println("windows.GetProcessId result was 0, not printing.")
	}

	// --- GetCurrentThreadId ---
	tid := windows.GetCurrentThreadId()
	fmt.Printf("windows.GetCurrentThreadId(): %d\n", tid)

	// --- GetVersionExW / RtlGetVersion ---
	var osInfo windows.OsVersionInfoEx
	osInfo.OsVersionInfoSize = uint32(unsafe.Sizeof(osInfo))

	var versionError error

	statusRtlGV := windows.RtlGetVersion(&osInfo)
	if statusRtlGV != windows.STATUS_SUCCESS {
		versionError = fmt.Errorf("RtlGetVersion failed with NTSTATUS: 0x%X", statusRtlGV)
		log.Printf("windows.RtlGetVersion failed: status 0x%X. Attempting fallback with windows.GetVersionEx...", statusRtlGV)

		errGVEX := windows.GetVersionEx(&osInfo) // This is from golang.org/x/sys/windows
		if errGVEX != nil {
			log.Printf("Fallback windows.GetVersionEx also failed: %v", errGVEX)
			versionError = fmt.Errorf("RtlGetVersion failed (status 0x%X) and GetVersionEx also failed (%w)", statusRtlGV, errGVEX)
		} else {
			log.Println("Fallback windows.GetVersionEx succeeded.")
			versionError = nil
		}
	}

	if versionError == nil {
		fmt.Printf("windows.RtlGetVersion/GetVersionEx:\n")
		fmt.Printf("  Major Version: %d\n", osInfo.MajorVersion)
		fmt.Printf("  Minor Version: %d\n", osInfo.MinorVersion)
		fmt.Printf("  Build Number: %d\n", osInfo.BuildNumber)
		csdString := windows.UTF16ToString(osInfo.CsdVersion[:])
		fmt.Printf("  Service Pack: %s\n", csdString)
		fmt.Printf("  Product Type: %d\n", osInfo.ProductType)
	} else {
		fmt.Printf("  Failed to get OS Version information via x/sys/windows: %v\n", versionError)
	}
}
