---
showTableOfContents: true
title: "Exported Functions (Theory 5.2)"
type: "page"
---


## Overview
After potentially calling `DllMain`, the final step before we execute our shellcode is to invoke the function that does not. Now remember, in our design we have both our shellcode AND the functions to allocate memory, inject shellcode into memory, and then execute the shellcode in the DLL. But, this does not have to be the case.

One can also use a different approach in which it only the shellcode resides in the DLL, with the functions contained in the loader's logic. As always each design has its own pros and cons, in general I prefer keeping the functions with the shellcode. The point of mentioning this is however not to contrast the merits of each approach, but just to highlights that exporting a function is not always required.


## Why Call Exported Functions?

While `DllMain` provides an initialization point, the main work or payload of a DLL is usually contained within one or more functions explicitly **exported** by the DLL developer. As we discussed in Theory 1.1, exporting makes functions available for external callers. In our working example, the DLL exports the `LaunchCalc` function specifically to trigger the calculator shellcode. Our reflective loader thus needs a way to find the address of this `LaunchCalc` function within the mapped DLL memory so it can call it.

Relying solely on `DllMain` is often insufficient because:

- `DllMain` has strict limitations on what API calls can be safely made within it to avoid deadlocks during process/thread attach/detach events.
- The main functionality might need to be triggered on demand or multiple times, requiring a dedicated exported function.
- `DllMain` might not even exist if the DLL requires no special initialization.

Therefore, after initializing the DLL via `DllMain` (if present), the loader typically proceeds to locate and call a specific target exported function.

## How the PE Format Supports Exports

Similar to imports, the PE format has a dedicated structure for defining exported functions: the **Export Directory**. The `IMAGE_OPTIONAL_HEADER`'s `DataDirectory` array entry at index `IMAGE_DIRECTORY_ENTRY_EXPORT` (index 0) points (via RVA) to an `IMAGE_EXPORT_DIRECTORY` structure.

This `IMAGE_EXPORT_DIRECTORY` structure contains fields pointing to three crucial tables - see below. Additionally, the `IMAGE_EXPORT_DIRECTORY` also contains `NumberOfFunctions` (total number of exports, size of EAT) and `NumberOfNames` (number of functions exported by name, size of ENPT and EOT).

### `AddressOfFunctions` (EAT - Export Address Table)
- An RVA pointing to an array of RVAs.
- Each entry in this array is the RVA of an exported function's starting code address within the DLL.
- This table lists _all_ exported functions, whether exported by name or just by ordinal number. The index into this array corresponds to the function's ordinal number (adjusted by the `Base` field in the `IMAGE_EXPORT_DIRECTORY`, though the base is often 0 or 1).


### `AddressOfNames` (ENPT - Export Name Pointer Table)
- An RVA pointing to an array of RVAs.
- Each entry in this array is an RVA pointing to a null-terminated ASCII string containing the name of an exported function.
- Importantly, this table is sorted alphabetically by function name, allowing for efficient binary searches (though a simple linear scan is often sufficient).
- This table only lists functions exported _by name_.


### `AddressOfNameOrdinals` (EOT - Export Ordinal Table)
- An RVA pointing to an array of 16-bit (WORD) values.
- This table acts as a bridge between the ENPT and the EAT. The index of an entry in the EOT corresponds to the index of a name RVA in the ENPT. The _value_ stored at that index in the EOT is the ordinal number (which serves as the correct index into the EAT) for that function name.
- For example, if the 5th entry in ENPT points to the string "LaunchCalc", the 5th entry in EOT will contain the ordinal number for "LaunchCalc". Let's say that ordinal is 2. Then, the RVA of the actual `LaunchCalc` function will be found in the EAT at index 2 (`EAT[2]`).


## The Process of Finding an Export by Name

To find the address of a specific exported function (like `"LaunchCalc"`) using these tables, the reflective loader performs the following steps:

1. **Locate Export Directory:** Get the RVA of the Export Directory from `DataDirectory[0]`. If zero, the DLL exports nothing. Calculate the VA (`ExportDirVA = ActualAllocatedBase + ExportDirRVA`). Read the `IMAGE_EXPORT_DIRECTORY` structure at `ExportDirVA`.
2. **Locate Tables:** Calculate the VAs of the EAT, ENPT, and EOT using the RVAs (`AddressOfFunctions`, `AddressOfNames`, `AddressOfNameOrdinals`) stored in the export directory structure and the `ActualAllocatedBase`.
3. **Search Names:** Iterate through the Export Name Pointer Table (ENPT) from index `i = 0` to `NumberOfNames - 1`.
    - For each index `i`:
        - Read the RVA of the name string from `ENPT[i]`.
        - Calculate the VA of the name string (`NameVA = ActualAllocatedBase + NameRVA`).
        - Read the null-terminated string at `NameVA`.
        - Compare this string to the target function name (e.g., `"LaunchCalc"`).
        - **If Match Found:**
            - Read the 16-bit ordinal from the Export Ordinal Table (EOT) at the **same index `i`**: `ordinal = EOT[i]`.
            - Use this `ordinal` value as the index into the Export Address Table (EAT). Read the function's RVA from `EAT[ordinal]`: `FunctionRVA = EAT[ordinal]`.
            - Calculate the final Virtual Address of the target function: `FunctionVA = ActualAllocatedBase + FunctionRVA`.
            - Store `FunctionVA` and stop searching.
4. **Handle Not Found:** If the loop completes without finding the target name, the function is not exported by name from this DLL.

## Calling the Function

Once the `FunctionVA` of the target exported function has been successfully determined:

1. **Use `syscall.SyscallN`:** Just like calling `DllMain`, we can use Go's `syscall.SyscallN` (or an equivalent mechanism for calling function pointers) to execute the code at `FunctionVA`.
2. **Pass Arguments:** Provide the correct number of arguments expected by the exported function. For our `LaunchCalc` function, which takes no arguments (`BOOL LaunchCalc()`), we pass `0` for the argument count and `0` for the subsequent argument placeholders in `syscall.SyscallN`.
3. **Check Return Value:** Handle any return value from the exported function as appropriate. `LaunchCalc` returns a `BOOL` indicating success/failure of executing the shellcode.

Calling the target exported function typically triggers the main intended action of the reflectively loaded DLL. In our example, successfully finding and calling `LaunchCalc` should finally result in the Windows Calculator appearing on the screen.

## Conclusion

This concludes the core theory behind finding and calling exported functions. We can now implement these last two crucial steps into our application, which will give us a complete and functional reflective loader.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "entry.md" >}})
[|NEXT|]({{< ref "entry_lab.md" >}})
