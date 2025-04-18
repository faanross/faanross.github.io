---
showTableOfContents: true
title: "Mapping the DLL Image (Theory 3.3)"
type: "page"
---
## Overview

With a suitable block of virtual memory allocated (Theory 3.2), the next critical phase is to populate this block with the contents of the DLL, effectively reconstructing the DLL's layout as it would appear in memory if loaded conventionally. This process involves two main copying steps: **transferring the PE headers** and then **mapping each individual section**.


## Allocate Memory (Recap)

As covered in the preceding section, the foundational step is allocating a contiguous block of virtual memory using `VirtualAlloc`. The size requested is `SizeOfImage` from the DLL's Optional Header, and initial permissions are  set to allow subsequent writing and eventual execution. Let's call the base address returned by `VirtualAlloc` the `AllocatedBase`.


## Copy PE Headers

The first thing we'll do is copy the parts of the PE Headers that contain all the essential metadata: **the DOS Header**, **NT Headers** (including the File Header and Optional Header), and t**he array of Section Headers**. It's critical that its copied directly to the beginning of our newly allocated memory block.

In order to do this we'll need to use the `SizeOfHeaders` value obtained from the `IMAGE_OPTIONAL_HEADER` since it tells us exactly how many bytes, starting from the beginning of our DLL memory buffer, constitute the complete set of headers.

So we copy `SizeOfHeaders` bytes from the start of the source DLL buffer to the `AllocatedBase` address. After this, the first `SizeOfHeaders` bytes of our allocated block will identical to the first `SizeOfHeaders` bytes of the original DLL file buffer.


## Copy Sections

Our PE Headers provide the map, but the actual code and data reside in the various sections (`.text`, `.data`, `.rdata`, etc.) located later in the DLL file buffer. Our goal now is to copy this raw data, for each section, from its position in the buffer to its correct **virtual position** within the allocated memory block.

We can do this by using the array of `IMAGE_SECTION_HEADER` structures which we just copied in the previous step. The `NumberOfSections` field from the `IMAGE_FILE_HEADER` tells us how many section headers to process.

Once completed, we then  iterate through each section header:
- **Identify Source:** Get the `PointerToRawData` and `SizeOfRawData` from the current section header. `PointerToRawData` is the file offset within our source DLL buffer where this section's data begins. `SizeOfRawData` is how many bytes to copy from that offset.
- **Identify Destination:** Get the `VirtualAddress` (an RVA) from the current section header. Calculate the absolute destination address in our allocated block: `DestinationVA = AllocatedBase + VirtualAddress`.
- **Action:** Copy `SizeOfRawData` bytes from `(SourceDllBufferBase + PointerToRawData)` to `DestinationVA`.
- **Handle Empty Sections:** If a section header has `SizeOfRawData` equal to zero, it means the section occupies space in memory but has no corresponding data in the file (this is typical for uninitialized data sections like `.bss`). In this case, no copying is needed for this section, as `VirtualAlloc` already initialized the committed memory pages to zero.
- **Iteration:** Repeat this process for all sections specified by `NumberOfSections`.


I know this is quite a lot, but we'll "map" each action here to its specific code and outcome in the next lab, that should help you form a much more concrete picture of what's happening here.

## Result of Mapping

After completing this final step the memory block starting at `AllocatedBase` will now contain a complete, mapped image of the DLL. The headers are at the beginning, and each section's data has been copied from its file offset in the source buffer to its correct relative virtual address within the allocated block.

This mapped image _structurally_ resembles how the DLL would look in memory if loaded by `LoadLibrary`, *but* it's not quite ready for execution yet. The addresses within the code and data might still be pointing to locations relative to the DLL's _preferred_ `ImageBase`, not the `AllocatedBase` where we actually loaded it.
Furthermore, the DLL likely depends on functions from other system DLLs, and the pointers for these imports haven't been resolved yet. These are the crucial tasks of **fixing relocations** and **resolving imports**, which we will cover in Module 4.

## Checklist
Before we proceed with our lab, I thought it would be useful to distill what we've discussed here in a numbered checklist of sorts, which can then be referenced server as a roadmap to guide us.

1. Allocate a contiguous block of virtual memory (result is `AllocatedBase`).
2. Obtain the `SizeOfHeaders` value from the PE Optional Header.
3. Copy `SizeOfHeaders` bytes from the start of the source DLL buffer to `AllocatedBase`.
4. Obtain the `NumberOfSections` value from the PE File Header.
5. Calculate the starting address of the first `IMAGE_SECTION_HEADER` within the mapped headers (at `AllocatedBase`).
6. Start a loop to iterate through each section (from 0 to `NumberOfSections - 1`).
7. Read the current `IMAGE_SECTION_HEADER` structure.
8. Get the `PointerToRawData` value from the current section header.
9. Get the `SizeOfRawData` value from the current section header.
10. Get the `VirtualAddress` (RVA) value from the current section header.
11. Check if `SizeOfRawData` is zero.
12. If `SizeOfRawData` is _not_ zero, proceed to the next step; otherwise, skip to step 16 (continue loop).
13. Calculate the source address for copying (`SourceDllBufferBase + PointerToRawData`).
14. Calculate the destination address for copying (`AllocatedBase + VirtualAddress`).
15. Copy `SizeOfRawData` bytes from the calculated source address to the calculated destination address.
16. Continue the loop for the next section until all sections are processed.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "memalloc.md" >}})
[|NEXT|]({{< ref "maplab.md" >}})