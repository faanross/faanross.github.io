---
showTableOfContents: true
title: "Base Relocations (Theory 4.1)"
type: "page"
---
## Overview

In the previous module, we successfully allocated memory and copied the DLL's headers and sections into it, creating a mapped image. However, this image might contain hardcoded assumptions about its location in memory. Therefore, before our DLL's code can run reliably, we need to address these assumptions, starting with base relocations.


## Why are Base Relocations Needed?

When a compiler and linker build a DLL or executable, they often embed absolute virtual addresses directly into the code or data sections. For example, a `call` instruction might target a specific function address, or a pointer in the `.data` section might point to a global string. These absolute addresses are calculated based on the assumption that the DLL will be loaded into memory starting at its **preferred `ImageBase`**, a value specified in the `IMAGE_OPTIONAL_HEADER`.

However, as we saw in Lab 3.1, our call to `VirtualAlloc` might not succeed in allocating memory at that preferred `ImageBase` if the address range is already occupied. In such cases, `VirtualAlloc` gives us a _different_ base address (`ActualAllocatedBase`). If the DLL is loaded at an address other than its preferred `ImageBase`, all the hardcoded absolute addresses within its mapped image will be incorrect, pointing to the wrong locations in memory. Executing code that relies on these incorrect addresses would lead to crashes or unpredictable behaviour.

**Base relocations** are the mechanism defined by the PE format to fix these hardcoded addresses after the module has been loaded at an actual base address that differs from its preferred one.



## How the PE Format Supports Relocations

The PE file contains specific information telling the loader exactly which locations within the mapped image need to be adjusted if the base address changes. This information is stored in the **base relocation table**, typically found in a section named `.reloc`.

The `IMAGE_OPTIONAL_HEADER`'s `DataDirectory` array points to this table. Specifically, the entry at index `IMAGE_DIRECTORY_ENTRY_BASERELOC` (index 5) contains the RVA and size of the base relocation table.

This table is structured as a series of **relocation blocks**. Each block starts with an `IMAGE_BASE_RELOCATION` structure:


```C
typedef struct _IMAGE_BASE_RELOCATION {
    DWORD   VirtualAddress; // RVA of the page this block applies to
    DWORD   SizeOfBlock;    // Total size of this block, including this header and all entries
} IMAGE_BASE_RELOCATION;
```

- `VirtualAddress`: The base RVA for all the relocations described in this block. Usually, this corresponds to the start of a memory page (e.g., `0x1000`, `0x2000`).
- `SizeOfBlock`: The total size of this `IMAGE_BASE_RELOCATION` header _plus_ all the 16-bit relocation entries that follow it. This tells the loader how many entries are in the current block and where the next block begins.

Immediately following each `IMAGE_BASE_RELOCATION` header is a series of 16-bit (WORD) entries. Each entry describes a single location within the page specified by `VirtualAddress` that needs patching. These 16-bit entries are structured as follows:
- **Relocation Type (Top 4 bits):** Specifies how the relocation should be applied. For modern reflective loaders targeting 64-bit Windows, the most important type is `IMAGE_REL_BASED_DIR64` (value `10`). This indicates that the relocation applies to a full 64-bit address. Other types exist (like `IMAGE_REL_BASED_HIGHLOW` for 32-bit, `IMAGE_REL_BASED_ABSOLUTE` which is padding and skipped), but `DIR64` is key for x64.
- **Offset (Bottom 12 bits):** An offset relative to the `VirtualAddress` specified in the block header. Adding this offset to the block's `VirtualAddress` gives the RVA within the DLL image where the address needs to be fixed.

## The Relocation Process

If the loader determines that the DLL was loaded at an `ActualAllocatedBase` different from the preferred `ImageBase`, it must perform the following steps:

1. Calculate the Delta: Compute the difference between the actual load address and the preferred base address:

   **delta = ActualAllocatedBase - PreferredImageBase**

   This delta is the value that needs to be added to each hardcoded address within the DLL that requires relocation. Note that this needs to be calculated using pointer-sized integers (e.g., int64 or uintptr arithmetic in Go) to handle potential address differences correctly. 

2. **Locate the Relocation Table:** Find the RVA and Size of the base relocation directory from the `DataDirectory` (index 5). Calculate the starting virtual address of the table: `RelocTableVA = ActualAllocatedBase + RelocTableRVA`.

3. **Iterate Through Blocks:** Starting at `RelocTableVA`, process the table block by block:
    - Read the `IMAGE_BASE_RELOCATION` header for the current block.
    - If `SizeOfBlock` is zero, stop processing (end of table).
    - Calculate the number of 16-bit entries following this header: `numEntries = (SizeOfBlock - sizeof(IMAGE_BASE_RELOCATION)) / 2`.
    - Get a pointer to the first entry (immediately following the header).

4. **Iterate Through Entries:** For each of the `numEntries` in the current block:
    - Read the 16-bit entry.
    - Extract the `RelocationType` (top 4 bits) and `Offset` (bottom 12 bits).
    - If `RelocationType` is `IMAGE_REL_BASED_DIR64` (for 64-bit):
        - Calculate the **VA to patch**: `PatchVA = ActualAllocatedBase + BlockVirtualAddress + Offset`.
        - Read the 64-bit value currently stored at `PatchVA`.
        - Add the `delta` calculated in Step 1 to this value.
        - Write the new, adjusted 64-bit value back to `PatchVA`.
    - If `RelocationType` is `IMAGE_REL_BASED_ABSOLUTE` (0), do nothing (it's just padding).
    - Handle or ignore other relocation types as needed (though `DIR64` is primary for x64).

5. **Advance to Next Block:** Move the processing pointer forward by `SizeOfBlock` bytes to get to the header of the next relocation block and repeat from Step 3.


After successfully processing all relocation blocks, any hardcoded absolute addresses within the DLL's mapped image that depended on the preferred `ImageBase` have now been adjusted to be correct relative to the `ActualAllocatedBase`. This makes the code and data consistent with the DLL's actual location in memory.


## Conclusion
With relocations handled, the DLL's internal addressing should now be self-consistent. However, it likely still depends on functions from _other_ DLLs, so our next critical step is resolving these external dependencies by processing the Import Address Table (IAT).




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module03/maplab.md" >}})
[|NEXT|]({{< ref "iat.md" >}})