---
showTableOfContents: true
title: "PE Header Inspection with PE-Bear (Lab 2.1)"
type: "page"
---
## Goal

Following our discussion of the essential PE file structures in Theory 2.1, this lab provides a hands-on opportunity 
to visually inspect these structures within the `calc_dll.dll` file created in Lab 1.1. Using the popular PE analysis tool, 
PE-Bear, we'll manually locate and examine the key header fields that are critical for understanding how a PE file is 
organized and eventually loaded into memory.

You can download PE-Bear [here](https://github.com/hasherezade/pe-bear/releases/tag/v0.7.0.4).

## Instructions

To get going: download PE-Bear, start the application, and open the dll we created in Lab 1.1 (`calc_dll.dll`).

## DOS Header
- Select `DOS Header` from the file structure tree on the LHS. 
- The first bytes in the `DOS Header` should be the `e_magic` field, We can confirm its value is `0x5A4D`, which corresponds to the ASCII characters "MZ".

![MZ](../img/mz.png)

- Next, we want to locate the `e_lfanew`, which is typically located at offset `3C`.
- Note that PE-Bear and many other tools might give it a slightly more descriptive label.
- Indeed we can see at `3C` a field called  `File address of new exe header` with a value of 80 (hex).
- This tells us that the NT Headers structure starts at the $80$th byte (offset 0x80) from the beginning of the file.

![3C](../img/3c.png)

## NT Headers

- Using the offset value from `e_lfanew`, we could manually find the NT Header in a hex editor, but PE-Bear makes it easier. It's still crucial to understand how to do it manually since that's how we'll be able to parse it ourselves in the next lab, as well as in our reflective loader.
- For now however let's allow PE-Bear to do the heavy lifting for us. 
- Locate the `NT HEADERS` entry in the structure view. PE-Bear usually groups the Signature, File Header, and Optional Header under this.

### Signature
- Verify the `Signature` field immediately follows the DOS Header (at the offset `e_lfanew`) is `0x00004550`, representing "PE\0\0".

![signature](../img/pe.png)

-  Note that the difference between the expected signature value `0x00004550` and the bytes we see (`50 45 00 00`) is due to **little-endian** byte ordering.
- The PE file format (like most things on Windows/Intel architecture) uses little-endian. This means that for multi-byte values, the _least significant byte_ is stored first (at the lowest memory address).
- Also notice the location of the signature - offset 0x80 - is exactly where `e_lfanew` indicated it would be located. 


### File Header
- Select the File Header tab within NT Headers. Locate and **note down the values** for the following key fields:
    - `Machine`: Identify the target architecture (e.g., `0x8664` for IMAGE_FILE_MACHINE_AMD64).
    - `Sections Count`: See how many sections the DLL contains.
    - `Size of OptionalHeader`: Note the size of the structure that follows.
    - `Characteristics`: Check the flags. Look for `IMAGE_FILE_DLL` (0x2000) to confirm it's a DLL.

![file header](../img/fileh.png)


### Optional Header
- Select the Optional Header tab, locate and **note down the values** for these fields:
  - `Magic`: Confirm it's `0x20B` for NT64.
  - `Entry Point`: Note this **RVA** (Relative Virtual Address). This is where execution conceptually begins.
  - `ImageBase`: Note the **preferred** starting virtual address for this DLL in memory.`
  - `SizeOfImage`: Note the **total virtual memory size** required for the loaded DLL.
  - `SizeOfHeaders`: Note the combined size of all headers at the beginning of the file.
  - `NumberOfRvaAndSizes`: Note the number of entries defined for the Data Directory.
  - `DataDirectory`: Observe the array of `IMAGE_DATA_DIRECTORY` structures at the end of the Optional Header. You don't need to analyze each entry now, but recognize that this array points to crucial tables like Imports, Exports, and Relocations using RVAs.

![optional header](../img/optheader.png)


### Section Headers (`IMAGE_SECTION_HEADER` Array)
- Select the Section Headers tab and locate the .text section:

![text section](../img/text.png)

- Locate and note down the values for:
  - **`Name`**: `.text`
  - **`Raw Addr`**:  Also known as `PointerToRawData`, this is the offset _from the beginning of the file_ where the section's raw data starts. The value is **`600`** (hex), below it we can note `1E00`, which indicates the file offset **immediately following** the `.text` section's raw data (600 + 1800 = 1E00) - i.e. `.data`.
  - **`Raw size`**: Also known as `SizeOfRawData`,  this is the size of the section's data _in the file_ on disk, we can see the value is `1800` (hex).
  - **`Virtual Addr`**: OK, this is kind of confusing but this is actually not VA, but RVA... But it's not PE-Bear that's at fault here - it's a legacy naming convention. The address is "virtual" because it relates to the virtual memory map (as opposed to a raw file offset), but the value stored is "relative" to allow for image relocation. Ideally, the field might have been named `RelativeVirtualAddress` from the start, but it wasn't.

    In any case we see our RVA value is 0x1000, so combining this with our ImageBase from the Optional Header gives us **0x26A5B0000 + 0x1000 = 0x26A5B1000**, which of course represents the **Virtual Address** within the process's private virtual address space where the `.text` section intends to begin. But a reminder: if that address is occupied or ASLE is enabled,  the actual `ImageBase` will differ, but the _offset_ (`0x1000`) from whatever that actual `ImageBase` turns out to be will remain constant.

    Below it we see 3000, which is again where the next section, it `.data`, begins (1000 + 2000 mapped virtual size)
  - **`Virtual Size`**: This is the _total aligned virtual address space_ consumed by the `.text` section. It starts at `1000` and the next section starts at `3000`, so the total span allocated in the virtual address space, including any padding for alignment, is `3000 - 1000 = 2000` bytes. Even though the actual data (`Virtual Size`) is `16F0`, it reserves `2000` bytes of address space.
  - **`Characteristics`**: The value is **`60000020`**(hex), which represents bitmasks, fortunately PEBear makes our lives easier by decoding the flags as `r-x`, meaning of course **Read + Execute**.

- If you'd like some more practice, feel free to repeat this same analysis for `.rdata` for read-only data or `.data` for initialized read/write data) to observe how their `VirtualAddress`, `PointerToRawData`, and `Characteristics` differ.
- Since these values are crucial, it's worth repeating: the `Raw Addr` tells the loader _where to read from in the file_, `Raw size` tells _how much to read_, and `Virtual Addr` (RVA) tells _where to place it in memory_ (relative to the `ImageBase`). The `Characteristics` dictate the memory permissions.

## Conclusion
- We have now visually navigated a PE file's structure using a common analysis tool, directly observing the header fields discussed theoretically.
- But of course, it does not stop here - now that you know how to navigate the tool and what's important, it's now up to you to get into the habit of analyzing PE files of interest to really solidify this knowledge.
- This is table-stakes: having a clear, deep, and intuitive grasp of these values, what they mean, and why thy are important is foundational knowledge for malware development, there's no shortcut here.
- Finally note that [PEExplorerV2](https://github.com/zodiacon/AllTools), a tool by Pavel Yosifovich, is also excellent and will give you the same information as PE-Bear, so if for any reason you'd like to explore an alternative feel free to do so.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "addresses.md" >}})
[|NEXT|]({{< ref "peparser.md" >}})