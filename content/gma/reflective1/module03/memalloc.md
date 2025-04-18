---
showTableOfContents: true
title: "Memory Allocation (Theory 3.2)"
type: "page"
---
## Overview

Now that we understand the concept of reflective loading — loading a DLL from memory without `LoadLibrary` — we need to address the first practical challenge: where do we put the DLL's contents in the process's virtual address space? Unlike standard loading where the OS handles memory management, **a reflective loader must explicitly allocate a suitable block of memory**. The primary Windows API function for this task is `VirtualAlloc`.

## Windows API: `VirtualAlloc`

`VirtualAlloc` (residing in `kernel32.dll`) allows a process to reserve, commit, or change the state of a region of memory pages in its own virtual address space.

- **Reserving:** This marks a range of virtual addresses as set aside for future use but doesn't actually allocate physical memory or page file space yet. It prevents other allocation functions from using that address range.
- **Committing:** This allocates physical storage (either RAM or page file space) for the specified region of previously reserved or newly reserved pages. Only committed pages can actually store data.

For reflective loading, we typically need to both reserve and commit memory in a single step to create a usable memory block for our DLL image.

## Purpose in Reflective Loading

Before we can copy the PE headers and various sections (`.text`, `.data`, etc.) from our in-memory DLL buffer, we need a contiguous block of virtual memory within the current process large enough to hold the entire mapped image. The required size is specified by the `SizeOfImage` field in the DLL's `IMAGE_OPTIONAL_HEADER`.
So, once we've determined that value, we can use `VirtualAlloc` to request this block from the operating system's memory manager.


## Key `VirtualAlloc` Parameters

The `VirtualAlloc` function (or its Go wrapper `windows.VirtualAlloc`) takes several important parameters that dictate how the memory is allocated:

### `lpAddress` (Desired Base Address)
- This parameter allows you to _suggest_ a starting virtual address for the allocation.
- **For Reflective Loading:** You typically first attempt to allocate memory starting at the DLL's _preferred_ `ImageBase` (obtained from the `IMAGE_OPTIONAL_HEADER`). If the DLL can be loaded at its preferred base, it simplifies the process as base relocations (discussed later) won't be necessary.
- **If `NULL` (or 0 in Go):** If you pass `NULL` or `0`, you are telling the operating system to choose a suitable base address for the allocation itself. This is the fallback option if allocation at the preferred `ImageBase` fails (e.g., because that address range is already in use). Using `NULL` guarantees finding _some_ available block (if memory is available) but likely means you _will_ need to process base relocations later.
- **Return Value:** `VirtualAlloc` returns the actual base address of the allocated block, which might be the requested `lpAddress` or a different address chosen by the OS if `lpAddress` was `NULL` or unavailable.

### `dwSize` (Size)
- This specifies the total size, in bytes, of the memory region to allocate.
- **For Reflective Loading:** This value should be the `SizeOfImage` obtained from the DLL's `IMAGE_OPTIONAL_HEADER`. This ensures the allocated block is large enough to accommodate the entire mapped image, including all headers and sections, according to the alignment specified in the PE file.

### `flAllocationType` (Allocation Type)

- This determines the type of memory operation. For our purposes, the most common combination is:
    - `MEM_COMMIT` (0x1000): Allocates physical storage for the specified pages. The pages are initialized to zero.
    - `MEM_RESERVE` (0x2000): Reserves the range of virtual addresses without allocating physical storage. This prevents other allocations from using this range.
- **For Reflective Loading:** You almost always use `MEM_COMMIT | MEM_RESERVE` together. This reserves the address range _and_ commits physical memory to it simultaneously, making it ready to receive the DLL's data.

### `flProtect` (Memory Protection)

- This parameter sets the memory protection attributes for the allocated pages. It controls whether the memory can be read, written, or executed. Common flags include:
    - `PAGE_READONLY` (0x02): Read access only.
    - `PAGE_READWRITE` (0x04): Read and write access.
    - `PAGE_EXECUTE` (0x10): Execute access only.
    - `PAGE_EXECUTE_READ` (0x20): Execute and read access.
    - `PAGE_EXECUTE_READWRITE` (0x40): Execute, read, and write access.
- **For Reflective Loading:** A common, simpler approach is to initially allocate the entire region with `PAGE_EXECUTE_READWRITE`. This allows the loader to easily write the PE headers and sections into the allocated memory and also allows the code sections to be executed later.
- **Security Consideration:** While `PAGE_EXECUTE_READWRITE` is convenient during the loading process, it's poor security practice to leave code sections writable or data sections executable long-term. A more robust loader would use `VirtualAlloc` with `PAGE_READWRITE` initially, copy all sections, and _then_ use a separate function (`VirtualProtect`) to adjust the permissions of individual sections based on their characteristics (e.g., setting `.text` to `PAGE_EXECUTE_READ`, `.rdata` to `PAGE_READONLY`, `.data` to `PAGE_READWRITE`) before executing any code from the loaded DLL. 

## Conclusion
- The key takeaway here is that successfully calling `VirtualAlloc` provides the loader with a base address in the process's virtual memory space, pointing to a newly allocated block of the correct size (`SizeOfImage`) with appropriate initial permissions (`PAGE_EXECUTE_READWRITE`).
- This newly allocated block is the foundation upon which our reflective loader will build the DLL's structure by copying the headers and sections we learned about in Module 2.

___



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "intro.md" >}})
[|NEXT|]({{< ref "mapping.md" >}})