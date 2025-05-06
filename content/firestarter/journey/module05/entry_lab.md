---
showTableOfContents: true
title: "Call DllMain (Lab 5.1)"
type: "page"
---


## Goal
Let's continue with our loader application, now adding a step to call the DLL's entry point, `DllMain`, since we know it exists.


## Code
Note that since our application is becoming quite large, and most of it is staying the same, I'm going to give instructions on how to modify the code from Lab 4.2 instead of providing the entire solution. If however you want to see the complete final solution, refer to the Github repo here.

First, add the following import:
```go
import (
        "syscall" // Needed for SyscallN
    )
```


Then under our constants declarations add this entry:
```go
    const (
        DLL_PROCESS_ATTACH = 1 // Reason for calling DllMain on load
    )
```


We'll now add our logic to call DllMain directly inside of our main() function. Paste the following logic _after_ the IAT processing loop completes successfully, but _before_ the final Self-Check messages. Note that this now becomes Step 7, and our Self-Check is renamed to Step 8.

```go
// --- Call DLL Entry Point (DllMain) ---
        fmt.Println("[+] Locating and calling DLL Entry Point (DllMain)...")
        dllEntryRVA := optionalHeader.AddressOfEntryPoint
    
        if dllEntryRVA == 0 {
            fmt.Println("[*] DLL has no entry point (AddressOfEntryPoint is 0). Skipping DllMain call.")
        } else {
            entryPointAddr := allocBase + uintptr(dllEntryRVA)
            fmt.Printf("[+] Entry Point found at RVA 0x%X (VA 0x%X).\n", dllEntryRVA, entryPointAddr)
            fmt.Printf("[+] Calling DllMain(0x%X, DLL_PROCESS_ATTACH, 0)...\n", allocBase)
    
            // Call DllMain: BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved);
            // Arguments:
            //   hinstDLL = base address of DLL (allocBase)
            //   fdwReason = DLL_PROCESS_ATTACH (1)
            //   lpvReserved = 0 (standard for dynamic loads)
            ret, _, callErr := syscall.SyscallN(entryPointAddr, allocBase, DLL_PROCESS_ATTACH, 0)
    
            // Check for errors during the system call itself
            // Note: '0' corresponds to ERROR_SUCCESS for the syscall status
            if callErr != 0 {
                log.Fatalf("    [-] Syscall error during DllMain call: %v\n", callErr)
                // Consider cleanup before fatal exit if needed
            }
    
            // Check the boolean return value from DllMain itself
            // DllMain returns TRUE (non-zero) on success, FALSE (zero) on failure for attach.
            if ret != 0 { // Non-zero means TRUE
                fmt.Printf("    [+] DllMain executed successfully (returned TRUE).\n")
            } else { // Zero means FALSE
                // Failure during DLL_PROCESS_ATTACH usually means the DLL cannot initialize
                log.Fatalf("    [-] DllMain reported initialization failure (returned FALSE). Aborting.\n")
                // Consider cleanup before fatal exit if needed
            }
        }
    
        // --- Existing Step (Self-Check) should follow ---
```
``

Finally, let's replace our Self-Check step to focus on reporting results related to this new logic.

```go
// --- Step 8: Self-Check (Basic) --- (Renumbered)
	fmt.Println("[+] Manual mapping process complete (Headers, Sections copied, Relocations potentially applied, IAT resolved, DllMain called).") // Updated message
	fmt.Println("[+] Self-Check Suggestion:")
	fmt.Printf("    - Verify console output shows DllMain call attempt and success/failure.\n")                                                               // New check
	fmt.Printf("    - Use debugger: Set breakpoint at VA 0x%X before syscall to step into DllMain.\n", allocBase+uintptr(optionalHeader.AddressOfEntryPoint)) // New check
	fmt.Printf("    - Inspect memory at allocated base address (0x%X).\n", allocBase)
	fmt.Println("    - Verify PE signatures and section data.")
	fmt.Println("    - Verify IAT pointers are resolved.")

	fmt.Println("\n[+] Press Enter to free memory and exit.")
	fmt.Scanln()

	fmt.Println("[+] Loader finished.")
```



## Code Breakdown

Note: This explains only the logic added or significantly changed compared to the IAT Resolving Mapper code from Lab 4.2.

### New Constants

* **`DLL_PROCESS_ATTACH` (value 1):** Added constant representing the reason code passed to `DllMain` when a DLL is first loaded into a process.

### `main` Function Logic

#### NEW Call DLL Entry Point (Step 7)
- **Get Entry Point RVA:** Retrieves the `AddressOfEntryPoint` RVA directly from the `optionalHeader` struct (which was parsed in Step 1).
- **Check for Entry Point:** An `if dllEntryRVA == 0` check determines if the DLL actually has an entry point. If not, it logs a message and skips the rest of this step.
- **Calculate Entry Point VA:** If `dllEntryRVA` is non-zero, it calculates the absolute virtual address (`entryPointAddr`) by adding the RVA to the `allocBase` where the DLL was mapped.
-  **Boundary Check:** A safety check is added to ensure the calculated `entryPointAddr` falls within the memory region allocated for the DLL (`allocBase` to `allocBase + allocSize`) before attempting to call it.
- **Log Call Attempt:** Prints messages indicating the VA found and that `DllMain` is about to be called with the specific arguments (`allocBase` for `hinstDLL`, `DLL_PROCESS_ATTACH` for `fdwReason`, `0` for `lpvReserved`).

- **Execute `DllMain`:** Uses `syscall.SyscallN` to make the call:
    * `entryPointAddr`: The address of the function to call.
    * `allocBase`: The first argument passed to `DllMain`.
    * `DLL_PROCESS_ATTACH`: The second argument passed to `DllMain`.
    * `0`: The third argument passed to `DllMain`.

- **Check Syscall Error:** Checks the third return value from `syscall.SyscallN` (`callErr`). If it's non-zero, it indicates an error during the syscall attempt itself (e.g., access violation if memory wasn't executable, invalid address) and logs a fatal error.

- **Check `DllMain` Return Value:** Checks the first return value from `syscall.SyscallN` (`ret`), which corresponds to the `BOOL` return value from `DllMain`. If `ret` is `0` (FALSE), it indicates the DLL's initialization failed, and the program logs a fatal error. If `ret` is non-zero (TRUE), it logs success.



## Instructions

- Compile the new application.

```shell
GOOS=windows GOARCH=amd64 go build
```

- Then copy it over to target system and invoke from command-line, providing as argument the dll you’d like to analyze, for example:

```bash
".\entry_call.exe .\calc_dll.dll"
```



## Results
```shell
PS C:\Users\vuilhond\Desktop> .\entry_call.exe .\calc_dll.dll
[+] Starting Manual DLL Mapper (with IAT Resolution)...
[+] Reading file: .\calc_dll.dll
[+] Parsed PE Headers successfully.
[+] Target ImageBase: 0x26A5B0000
[+] Target SizeOfImage: 0x22000 (139264 bytes)
[+] Allocating 0x22000 bytes of memory for DLL...
[+] DLL memory allocated successfully at actual base address: 0x26A5B0000
[+] Copying PE headers (1536 bytes) to allocated memory...
[+] Copied 1536 bytes of headers successfully.
[+] Copying sections...
[+] All sections copied.
[+] Checking if base relocations are needed...
[+] Image loaded at preferred base. No relocations needed.
[+] Processing Import Address Table (IAT)...
[+] Import Directory found at RVA 0x9000
    [->] Processing imports for: KERNEL32.dll
    [+] Finished imports for 'KERNEL32.dll'.
    [->] Processing imports for: api-ms-win-crt-environment-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-environment-l1-1-0.dll'.
    [->] Processing imports for: api-ms-win-crt-heap-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-heap-l1-1-0.dll'.
    [->] Processing imports for: api-ms-win-crt-runtime-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-runtime-l1-1-0.dll'.
    [->] Processing imports for: api-ms-win-crt-stdio-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-stdio-l1-1-0.dll'.
    [->] Processing imports for: api-ms-win-crt-string-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-string-l1-1-0.dll'.
    [->] Processing imports for: api-ms-win-crt-time-l1-1-0.dll
    [+] Finished imports for 'api-ms-win-crt-time-l1-1-0.dll'.
[+] Import processing complete (7 DLLs).
[+] Locating and calling DLL Entry Point (DllMain)...
[+] Entry Point found at RVA 0x1330 (VA 0x26A5B1330).
[+] Calling DllMain(0x26A5B0000, DLL_PROCESS_ATTACH, 0)...
    [+] DllMain executed successfully (returned TRUE).
[+] Manual mapping process complete (Headers, Sections copied, Relocations potentially applied, IAT resolved, DllMain called).
[+] Self-Check Suggestion:
    - Verify console output shows DllMain call attempt and success/failure.
    - Use debugger: Set breakpoint at VA 0x26A5B1330 before syscall to step into DllMain.
    - Inspect memory at allocated base address (0x26A5B0000).
    - Verify PE signatures and section data.
    - Verify IAT pointers are resolved.

[+] Press Enter to free memory and exit.

[+] Loader finished.
[+] Attempting to free main DLL allocation at 0x26A5B0000...
[+] Main DLL memory freed successfully.
```


## Discussion
We can see are result was a success - the loader correctly found and invoked the DLL's entry point, and the DLL signalled successful initialization. The mapped DLL is now fully prepared and initialized in memory.

- **`Entry Point found at RVA 0x1330 (VA 0x26A5B1330).`** - This shows the loader successfully read the `AddressOfEntryPoint` from the Optional Header and calculated the correct virtual address for `DllMain`.
- **`Calling DllMain(0x26A5B0000, DLL_PROCESS_ATTACH, 0)...`** - Confirms the program is preparing to call the entry point via `syscall.SyscallN` with the correct arguments (`hinstDLL`, `fdwReason`, `lpvReserved`).
- **`DllMain executed successfully (returned TRUE).`** - This is the key result for this lab: the `syscall.SyscallN` executed without error, _and_ the `DllMain` function itself returned a non-zero value (TRUE), indicating that the DLL's internal initialization routine completed successfully.

## Conclusion
We'll now proceed with the final lab for this first part of our curriculum: finding and calling the exported function within our mapped DLL. 
This should bring everything together for a successful execution. Hell Yes!!


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "export.md" >}})
[|NEXT|]({{< ref "export_lab.md" >}})