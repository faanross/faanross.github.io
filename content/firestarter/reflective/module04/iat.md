---
showTableOfContents: true
title: "IAT Resolution (Theory 4.2)"
type: "page"
---
## Why is IAT Resolution Necessary?

DLLs rarely exist in isolation; they almost always utilize functions provided by _other_ DLLs. For instance, nearly any Windows DLL will need functions from `kernel32.dll` (for core OS services like memory management, process/thread control) and potentially others like `user32.dll` (for UI elements) or `ntdll.dll` (for lower-level system calls).

When a DLL calls a function from another DLL, it doesn't usually call that function's absolute address directly (as that address can change). Instead, the compiler generates an indirect call through a pointer stored within the calling DLL's memory space. The collection of these pointers are known as the **Import Address Table (IAT)**. Each entry in the IAT holds the actual memory address of an imported function.

## How LoadLibrary Creates the IAT

When the standard Windows loader loads a DLL via `LoadLibrary`, one of its crucial jobs is to:

1. Identify all the external DLLs the new DLL depends on.
2. Load those dependency DLLs into memory (if they aren't already loaded).
3. Find the addresses of all the required functions within those dependency DLLs.
4. Write these resolved addresses into the new DLL's IAT.

This "patches" the IAT so that when the new DLL's code calls an imported function indirectly through an IAT entry, it jumps to the correct location in the dependency DLL.

In reflective loading, since we bypass the OS loader, _we_ become responsible for performing this IAT resolution process manually. Without this step, any attempt by our reflectively loaded DLL to call an external function would fail, likely causing a crash, because the IAT entries would still contain placeholder information instead of valid function addresses.

## How the PE Format Supports Imports

The PE format provides the necessary information for resolving imports in the **Import Directory**. The `IMAGE_OPTIONAL_HEADER`'s `DataDirectory` array entry at index `IMAGE_DIRECTORY_ENTRY_IMPORT` (index 1) points to the start of this directory.

The Import Directory is essentially an array of `IMAGE_IMPORT_DESCRIPTOR` structures. Each structure corresponds to a single DLL that our manually loaded DLL depends on (e.g., one descriptor for `kernel32.dll`, another for `user32.dll`, etc.). The array is terminated by an `IMAGE_IMPORT_DESCRIPTOR` structure filled with nulls.

The key fields within an `IMAGE_IMPORT_DESCRIPTOR` are:

### Name
An RVA pointing to a null-terminated string that holds the name of the required dependency DLL (e.g., "kernel32.dll").

### OriginalFirstThunk`(OFT)
An RVA pointing to an array of `IMAGE_THUNK_DATA` entries (essentially pointer-sized values). This array, often called the Import Lookup Table (ILT), lists the specific functions to be imported from this dependency DLL. Each entry either encodes an ordinal number or points (via RVA) to an `IMAGE_IMPORT_BY_NAME` structure.

The `IMAGE_IMPORT_BY_NAME` structure (referenced by ILT entries when importing by name) simply contains a 16-bit "Hint" (an index suggestion for faster lookups in the exporting DLL) followed by the null-terminated string name of the function to import.

### FirstThunk (FT)
An RVA pointing to _another_ array of `IMAGE_THUNK_DATA` entries. This array is the **Import Address Table (IAT)** itself. Initially (before loading), the IAT often mirrors the ILT. The loader's job is to overwrite each entry in the IAT with the actual resolved address of the corresponding imported function.



## The IAT Resolution Process

A reflective loader must iterate through the Import Directory and resolve the imports for each required DLL.

The process generally follows these steps.

### Step 1: Locate the Import Directory
Find the RVA of the Import Directory from the `DataDirectory` (index 1) and calculate its VA (`ImportDirVA = ActualAllocatedBase + ImportDirRVA`).

### Step 2: Iterate Through Descriptors
Starting at `ImportDirVA`, process the array of `IMAGE_IMPORT_DESCRIPTOR` structures one by one until a null descriptor is encountered.

### Step 3: Then, for Each Descriptor
- **Get Dependency DLL Name:** Read the `Name` RVA, calculate the VA (`DllNameVA = ActualAllocatedBase + NameRVA`), and read the null-terminated string name of the required DLL.
- **Load Dependency:** Crucially, use the _standard_ Windows API function `LoadLibrary` (e.g., `windows.LoadLibrary` in Go) to load this dependency DLL into the current process's address space. This is necessary because we need the dependency DLL mapped correctly by the OS itself to find its exported functions. Keep the returned `HMODULE` handle.
- **Find ILT and IAT:** Calculate the base VAs of the Import Lookup Table (`ILT_VA = ActualAllocatedBase + OriginalFirstThunk`) and the Import Address Table (`IAT_VA = ActualAllocatedBase + FirstThunk`). (Note: If `OriginalFirstThunk` is zero, the ILT and IAT are the same, pointed to by `FirstThunk`).
- **Iterate Through Imports:** Loop through the entries in the ILT and IAT arrays in parallel (they correspond one-to-one). The loop terminates when an entry in the ILT is null (zero).
  - For each entry pair (`i`):
  - Read the value from the ILT entry (`ILT_VA + i * sizeof(uintptr)`). Let's call this `lookupValue`.
  - Determine if importing by **Ordinal** or **Name**:
  - **By Ordinal:** If the most significant bit of `lookupValue` is set (`IMAGE_ORDINAL_FLAG64` for 64-bit), the lower 16 bits represent the ordinal number of the function to import.
  - **By Name:** Otherwise, `lookupValue` is an RVA to an `IMAGE_IMPORT_BY_NAME` structure. Calculate its VA (`HintNameVA = ActualAllocatedBase + lookupValue`), read the function name string starting 2 bytes after `HintNameVA`.
  - **Get Function Address:** Use the _standard_ Windows API function `GetProcAddress` (e.g., `windows.GetProcAddress` or a syscall equivalent in Go), passing the `HMODULE` of the loaded dependency DLL and either the function name string or the ordinal number. This returns the actual VA of the required function within the loaded dependency DLL. Handle errors if `GetProcAddress` fails (this is usually fatal for the loader).
  - **Patch the IAT:** Write the function address obtained from `GetProcAddress` directly into the _corresponding entry_ in the IAT (`IAT_VA + i * sizeof(uintptr)`) within our reflectively mapped DLL's memory.
- **Repeat:** Continue processing descriptors until the null terminator is found.


After this process completes, the Import Address Table within our reflectively mapped DLL has been fully patched. All entries now point to the correct memory locations of the functions imported from external DLLs. The DLL's code can now successfully call these external functions via the patched IAT pointers.

## Conclusion

With both internal address references (relocations) and external dependencies (imports) resolved, the DLL is almost ready for execution. The final steps involve potentially calling its entry point (`DllMain`) and then invoking any specific exported function we need.

We'll cover this in Module 5, for now let's dip into some practical labs where we'll ensure our reflective loader is capable of relocations, and constructing an IAT table.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "reloc.md" >}})
[|NEXT|]({{< ref "reloc_lab.md" >}})