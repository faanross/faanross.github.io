---
showTableOfContents: true
title: "PE File Structure Essentials (Theory 2.1)"
type: "page"
---
## Preface
Be aware that the PE file structure is a very deep rabbit hole. I've personally found that other courses on this topic can get bogged down in laborious technical explanations, often causing interest to wane. My goal with this section, therefore, is to cover the absolute bare minimum needed for us to progress. Plenty of resources exist to take you deep into that rabbit hole if you wish, but what's provided here are just enough "bread crumbs." This will give you a foundational understanding for the upcoming labs – enough to know what you're looking at and, more importantly, why we care about these specific elements.

## Anatomy of a Windows Executable - The PE Format

In the previous module we learned how the standard Windows functions `LoadLibrary` and `GetProcAddress` are used to load DLLs from disk. 
I also mentioned the drawbacks of this approach, namely the reliance on disk-based files and the monitoring capabilities of the operating 
system's loader. To overcome these limitations and perform reflective loading, we need to essentially replicate the core tasks of the 
Windows loader ourselves. To do that, we must first understand the blueprint of Windows executables and DLLs: the **Portable Executable (PE) file format**.


The PE format is the standard file format for all versions of 32-bit and 64-bit Windows executables (`.exe`), object code, and DLLs (`.dll`). It's derived from the older Common Object File Format (COFF) used in Unix systems. Think of it as a container holding all the necessary information in a predictable arrangement, which allows the Windows loader to map the file into memory and prepare it for execution.

## PE File Structure Essentials

A PE file isn't just a jumble of code and data. It's a highly structured file, organized into various headers and sections. This is a very deep rabbit hole to potentially go down, but for the purpose of understanding manual loading (so that we can then develop a reflective loader), we'll only focus on the key components that guide this process.

When you look at a PE file  as a sequence of bytes, the structure generally contains 4 major components: DOS Header, NT Headers, Section Headers, Sections (Raw Data).

### DOS Header (`IMAGE_DOS_HEADER`)
- The DOS Header is located at the very beginning of the file.
-  It exists primarily for backward compatibility with MS-DOS
- As malware developers we typically only care about two fields here: `e_magic` and `e_lfanew`

#### `e_magic` (WORD - 2 bytes)
- Contains the signature `0x5A4D`, which is ASCII for "MZ".
- This identifies the file as a potential PE file (or at least a DOS-compatible executable).
- All PE files start with these two bytes.
- We typically look for `e_magic` to confirm it's potentially a valid file.

#### `e_lfanew` (LONG - 4 bytes)
- This is the crucial field for modern loaders.
- It holds the **file offset** (the byte position relative to the start of the file) where the main `IMAGE_NT_HEADERS` structure begins.
- We can then use this to jump directly to the important NT Headers, skipping the rest of the DOS-related information.


### NT Headers (`IMAGE_NT_HEADERS`)
- Starts at the file offset specified by `e_lfanew` in the DOS Header.
- This is the most important structure for understanding the PE file's layout and characteristics for loading.
- It's composed of three parts: Signature, File Header, and Optional Header - see below.
- The NT Headers provide the roadmap:
    - is it 32/64-bit?
    - Where does execution start?
    - Where _should_ it be loaded in memory (`ImageBase`)?
    - How much memory does it need (`SizeOfImage`)?
    - How big are the headers (`SizeOfHeaders`)?
    - And critically, where are the tables for imports, exports, and relocations (`DataDirectory`)?

#### **Signature** (DWORD - 4 bytes)
- Contains the value `0x00004550`, which is ASCII for "PE\0\0".
- This validates the file as a PE format file.
- If this signature isn't present after the "MZ", it's not a valid PE file.

#### **File Header (`IMAGE_FILE_HEADER`)**
- Contains basic information about the file itself.
- `Machine` (WORD): Specifies the target CPU architecture. Common values are `0x014c` (IMAGE_FILE_MACHINE_I386 for x86) or `0x8664` (IMAGE_FILE_MACHINE_AMD64 for x64). This is vital for ensuring the loader isn't trying to load, for example, a 64-bit DLL into a 32-bit process.
- `NumberOfSections` (WORD): Indicates how many section headers (and therefore sections) follow the Optional Header.
- `SizeOfOptionalHeader` (WORD): Specifies the size of the next part, the Optional Header. Needed to know where the section headers begin.
- `Characteristics` (WORD): Flags describing the file (e.g., whether it's an executable or a DLL, `IMAGE_FILE_EXECUTABLE_IMAGE`, `IMAGE_FILE_DLL`).

#### Optional Header (`IMAGE_OPTIONAL_HEADER`)
-  Despite its name, this header is essential for executable files and DLLs.
- It contains the most critical information for the loader regarding how to map the file into memory.
- Its exact structure differs slightly between 32-bit (PE32) and 64-bit (PE32+) files, but the key fields are present in both
- `Magic` (WORD): Identifies if it's PE32 (`0x10b`) or PE32+ (`0x20b` for 64-bit).
- `AddressOfEntryPoint` (DWORD): The **RVA** (Relative Virtual Address - explained below) of the first instruction to execute when the module starts (typically the CRT startup code, which eventually calls `main` or `DllMain`).
- `ImageBase` (DWORD for PE32, ULONGLONG for PE32+): The **preferred** virtual address where the loader should map the start of the PE file in memory. The loader will _try_ to load at this address, but might have to choose another if this address range is already occupied (leading to relocations, discussed later)
- `SectionAlignment` / `FileAlignment` (DWORDs): Specify the alignment requirements for sections in memory and in the file, respectively.
- `SizeOfImage` (DWORD): The **total size** (in bytes) that the mapped module will occupy in virtual memory. This is crucial for allocating the correct amount of memory before mapping.
- `SizeOfHeaders` (DWORD): The combined size of the DOS Header, NT Headers, and all Section Headers. This tells the loader how much data at the beginning of the file constitutes the headers, which are typically mapped first.
- `Subsystem` (WORD): Indicates the target subsystem (e.g., Windows GUI, Windows Console).
- `NumberOfRvaAndSizes` (DWORD): Specifies the number of entries in the `DataDirectory` array that follows.
- `DataDirectory` (Array of `IMAGE_DATA_DIRECTORY`): This is an array (usually 16 entries) where each entry points to a significant table or directory within the PE file, such as the Export Table, Import Table, Resource Table, Base Relocation Table, etc. Each entry contains two DWORDs: `VirtualAddress` (the RVA of the table) and `Size` (the size of the table in bytes). These entries are fundamental for locating structures needed during loading (like finding imports and exports).

### Section Headers (`IMAGE_SECTION_HEADER` Array)
- Section Headers are located immediately following the Optional Header.
- There is one `IMAGE_SECTION_HEADER` structure for each section specified by `NumberOfSections` in the File Header.
- Each header describes a corresponding section in the file, telling the loader where to find the section's data in the file and where to map it in memory, along with its characteristics.
- `Name` ([8]byte): An 8-byte (often null-padded) ASCII name for the section (e.g., `.text`, `.data`, `.rdata`, `.reloc`). These names are mostly informational, though conventions exist.
- `VirtualSize` (DWORD): The actual size of the section's data in memory (can be larger than `SizeOfRawData` for sections like `.bss` that contain uninitialized data).
- `VirtualAddress` (DWORD): The **RVA** where the beginning of this section should be mapped in memory, relative to the `ImageBase`.
- `SizeOfRawData` (DWORD): The size of the section's data _on disk_ (in the PE file). Must be a multiple of `FileAlignment`.
- `PointerToRawData` (DWORD): The **file offset** where the section's data begins in the PE file.
- `Characteristics` (DWORD): Flags defining the section's attributes (e.g., contains code `IMAGE_SCN_CNT_CODE`, readable `IMAGE_SCN_MEM_READ`, writable `IMAGE_SCN_MEM_WRITE`, executable `IMAGE_SCN_MEM_EXECUTE`). These are crucial for setting memory permissions correctly after loading.

- Why do we care about section headers as malware developers? They provide the specific instructions for mapping each chunk of the file. The loader iterates through these headers, reads `SizeOfRawData` bytes from `PointerToRawData` in the file, and writes those bytes to the memory location `ImageBase + VirtualAddress`. The `Characteristics` guide setting memory permissions later.

### Sections (Raw Data)
- After all the headers we find the actual data here, meaning you can think of everything up until this point as the meta-data, whereas in this section we find the actual data, logic etc associated with the PE file.
- These actual blocks of data (code, initialized data, resources, etc.) were described above by the Section Headers.
- Their locations and sizes in the file are given by `PointerToRawData` and `SizeOfRawData` in their respective headers.

## Conclusion
Now that we have a basic overview of the structure of a PE file we need to delve a bit deeper into an important concept that will be foundational to the logic of our reflective loader - the ability to map file offsets to specific memory addresses.

## Reference - **Corkami**
While understanding the PE format through its defined structures and fields is essential, the complexity can sometimes be overwhelming. 
For an excellent visual perspective, I highly recommend exploring the work of Ange Albertini, available through the [Corkami project](https://github.com/corkami/pics/blob/master/binary/pe101/README.md).

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module01/create_loader.md" >}})
[|NEXT|]({{< ref "addresses.md" >}})