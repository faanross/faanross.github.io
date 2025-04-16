---
showTableOfContents: true
title: "Addressing in PE Files (Theory 2.2)"
type: "page"
---
## Overview

Let's now turn our attention to a fundamental concept: how locations within the file are addressed and how these relate to memory 
addresses once the file is loaded. We'll explore the distinction between **Relative Virtual Addresses (RVAs)** 
and **Virtual Addresses (VAs)**. 

Understanding this difference is absolutely critical for developing a reflective loader. 
Why? Because a reflective loader must _manually_ replicate the process of mapping the PE file from its raw, disk-based format into a 
usable, executable image in the process's virtual memory. This involves interpreting the addresses stored within the PE headers 
(which are predominantly RVAs) and correctly calculating the final memory locations (VAs) where code and data should reside. 
Without understanding RVA-to-VA translation, accurately reconstructing the module in memory – the core task of any loader, 
including a reflective one – is impossible.

## Understanding PE Addressing
### RVA
As we examined the various PE headers (like the Optional Header and Section Headers), we noted numerous fields containing 
address information, such as `AddressOfEntryPoint`, the `VirtualAddress` in `IMAGE_SECTION_HEADER` entries, and the `VirtualAddress` 
fields within the `DataDirectory` array. A crucial characteristic of the PE format is that most of these addresses are not 
absolute memory locations. Instead, they are specified as **Relative Virtual Addresses (RVAs)**, the keyword here being **relative**.

An RVA is an offset, measured in bytes, relative to the starting memory address where the PE file is loaded. 
This starting address is known as the **ImageBase**. Think of the loaded module as a contiguous block of memory; 
the RVA tells you how far into that block a particular piece of data or code is located, starting from the very first byte 
(the `ImageBase`).

The primary reason for using RVAs is to achieve **relocatability**. PE files (especially DLLs, but also EXEs) are typically compiled with a _preferred_ `ImageBase` address specified in the Optional Header. This is the ideal memory location where the module would like to be loaded. However, when the Windows loader (or our manual reflective loader) attempts to load the module, this preferred address range might already be occupied by the main executable, another DLL, or a different memory allocation. In such cases, the loader must place, or _relocate_, the module at a different base address in the process's virtual address space.

If the addresses stored within the PE file were absolute, they would all point to incorrect locations if the 
module were loaded at any address other than its preferred `ImageBase`. By using RVAs, 
the internal references within the PE file remain valid regardless of where the module is actually loaded in memory. 
**The RVA represents a constant offset from whatever base address is ultimately chosen.**

### VA
In contrast to an RVA, a **Virtual Address (VA)**, often referred to as an absolute address, represents the final, 
actual memory address within the process's virtual address space where a specific piece of code or data resides after the 
module has been loaded.

The relationship between these addresses is straightforward and essential to grasp:

`VA = ActualImageBase + RVA`

Where:

- `VA` is the final Virtual Address in memory.
- `ActualImageBase` is the actual base memory address where the loader mapped the beginning (the first byte) of the PE file. This might be the preferred `ImageBase`, or it might be a different address assigned by the loader due to relocation.
- `RVA` is the Relative Virtual Address as stored within the PE file structure.

Consider an example: A DLL has a preferred `ImageBase` of `0x10000000` and its `.text` section header specifies a `VirtualAddress` (which is an RVA) of `0x1000`.

- If the loader successfully maps the DLL at its preferred `ImageBase` of `0x10000000`, the `.text` section will begin at the VA `0x10000000 + 0x1000 = 0x10001000`.
- However, if that address space is occupied and the loader must relocate the DLL, perhaps mapping it at `0x11500000` instead, the `.text` section will then begin at the VA `0x11500000 + 0x1000 = 0x11501000`.

Notice that the RVA (`0x1000`) stored within the PE file's section header remains constant; only the `ActualImageBase` changes, resulting in a different final VA for the section's start. (The potential need to adjust internal code references due to such base address changes is handled by base relocations, a topic we will cover later).

## Mapping File Content to Memory

The `IMAGE_SECTION_HEADER` structures provide the crucial link needed to map the raw data from the PE file on disk into the correct locations in virtual memory. Each section header contains the necessary information:

- `PointerToRawData`: This field specifies the **file offset** – the starting position (byte index from the beginning of the file) of the section's content within the PE file on disk.
- `SizeOfRawData`: This indicates the size, in bytes, of this section's data as it exists in the file.
- `VirtualAddress`: As discussed, this is the **RVA** specifying where the beginning of this section should be placed in memory, relative to the `ActualImageBase`.

Therefore, for each section defined in the PE file's section table, the loader performs the following conceptual steps:

1. It reads the `PointerToRawData`, `SizeOfRawData`, and `VirtualAddress` (RVA) values from the section's `IMAGE_SECTION_HEADER`.
2. It calculates the target destination **VA** in memory using the formula: `DestinationVA = ActualImageBase + VirtualAddress` (RVA).
3. It reads `SizeOfRawData` bytes directly from the PE file on disk, starting at the file offset specified by `PointerToRawData`.
4. It writes these bytes into the process's allocated virtual memory, starting at the calculated `DestinationVA`.

This process is repeated for every section, effectively copying the relevant parts of the PE file from disk and arranging them correctly in memory according to the layout defined by the headers.

## Conclusion
With this foundational understanding of PE structure and addressing mechanisms, let's now do a lab in which we can explore these
values directly using an application called PEBear (Lab 2.1), after which we'll create our very own PE Header parser (Lab 2.2),
which is going to be a core part of the logic of our final reflective loader.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "structure.md" >}})
[|NEXT|]({{< ref "pebear.md" >}})