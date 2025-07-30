---
showTableOfContents: true
title: "Standard DLL Loading in Windows (Theory 1.3)"
type: "page"
---
## The Windows API Functions
The Windows API provides a vast collection of functions that allow us to interact with the operating system, 
manage resources, create user interfaces, and much more. We'll get to know many of these key functions as we go along, 
but for now we'll learn about two of the most important ones - `LoadLibrary` and `GetProcAddress`. 


In our [introduction to DLLs](https://www.faanross.com/firestarter/reflective/module01/intro_dlls/), we mentioned that for a process to use a DLL function it essentially needs to know two things - 
the name of the DLL, and the name of the exported function. And it's these two functions that we use along with that information 
to load our shellcode:
- `LoadLibrary` is given the name/path of DLL, which it then loads.
- `GetProcAddress` is given the name of the exported function residing in the loaded DLL, to which it then gets the specific memory address pointing to it.


## LoadLibrary (or LoadLibraryEx)
This function is responsible for loading a specified DLL file from disk into the virtual address space of the calling process.

When an application calls `LoadLibrary` with the name or path of a DLL file (e.g., `"mydll.dll"` or `"C:\folder\mydll.dll"`), 
the Windows loader performs several steps:
- It searches for the DLL file in a predefined sequence of locations (including the application's directory, system directories, etc.).
- If found, it checks if the same DLL (matching path and version) is already loaded in memory by another process. If so, it maps the existing code pages. If not, it reads the DLL file from disk.
- It allocates virtual memory within the calling process's address space.
- It maps the different sections of the DLL (code, data, resources) from the file into the allocated memory, performing necessary base relocations if the DLL couldn't be loaded at its preferred base address.
- It resolves the DLL's own dependencies by recursively loading any other DLLs it imports. 
- Crucially, it resolves the Import Address Table (IAT) of the newly loaded DLL, filling it with the addresses of functions imported from other libraries. Keep this step in mind, since it will become central to our ability to create a reflective loader.
- If the DLL has an entry point (`DllMain`), it calls `DllMain` with the `DLL_PROCESS_ATTACH` reason, allowing the DLL to perform any necessary initialization.

If successful, `LoadLibrary` returns a **handle** (often referred to as `HMODULE` or `HINSTANCE`) to the loaded DLL. 
This handle is essentially a unique identifier representing the base address where the DLL was loaded in the process's memory.
If it fails (e.g., file not found, invalid format), it returns `NULL`.

**NOTE:** If every step and action described above does not make complete sense yet - that's fine! We're going
to dig into each much deeper, and you'll gain direct experience of each. By the end, this will all make perfect sense. 

## GetProcAddress
Once a DLL has been successfully loaded using `LoadLibrary` and we have its handle, `GetProcAddress` is then used to find the memory 
address of a specific exported function _within_ that loaded DLL.

The application calls `GetProcAddress`, providing two key arguments:
- The handle (`HMODULE`) returned by the earlier `LoadLibrary` call.
- The name (a string, e.g., `"LaunchCalc"`) or the ordinal number (an integer) of the desired exported function.
- `GetProcAddress` then searches the export table of the specified loaded DLL for an entry matching the provided name or ordinal.

If the function is found, `GetProcAddress` returns the **virtual memory address** where the function's code begins within 
the process's address space. This address can then be cast to a function pointer and called directly by the application. 
If the function is not found in the DLL's export table, `GetProcAddress` returns `NULL`.


## Calling Windows API from Go

Go provides numerous mechanisms to interact with native C-style libraries, right now I want to introduce the two most elementary techniques:

1. **`syscall` Package:** This built-in package offers lower-level access. You can load DLLs (`syscall.LoadDLL`) and find procedures (`dll.FindProc`). You then invoke the procedure using methods like `proc.Call()` or `syscall.SyscallN()`, passing arguments as `uintptr` types and handling potential errors. This method gives fine-grained control but requires careful management of types and error checking based on Windows conventions.

2. **`golang.org/x/sys/windows` Package:** This is the more "user-friendly" package (i.e. abstracted), but comes at the expense of having less control. It provides Go-style wrappers around many common Windows API functions. For instance, it directly offers functions like `windows.LoadLibrary` and `windows.GetProcAddress`. These wrappers handle much of the type conversion and error checking boilerplate.


Regardless of the package used, the underlying principle is the same: Go code obtains a pointer to the native Windows API function (like `LoadLibrary`) and then invokes that function according to its documented C signature, marshaling Go types to their C equivalents (like Go strings to null-terminated C strings or pointers).

## Drawbacks of Standard Loading

While `LoadLibrary` and `GetProcAddress` are the standard, documented way to load DLLs, this mechanism has major drawbacks
for malware development:

1. **Requires DLL on Disk:** `LoadLibrary` fundamentally operates on files. The DLL must exist somewhere on the file system where the Windows loader can find it. This creates a significant forensic footprint. If malware drops a malicious DLL to disk, it can be easily found, analyzed, and signatured by antivirus software.
2. **OS Loader Mechanisms Can Be Monitored:** The actions performed by the Windows loader during a `LoadLibrary` call (file access, registry checks, memory mapping, import resolution) are well-defined and can be monitored by security tools. Many security products like EDRs hook into `LoadLibrary` or related lower-level OS functions specifically to inspect and potentially block the loading of suspicious DLLs. The OS also maintains records of loaded modules within each process (e.g., in the Process Environment Block or PEB), which are easily inspected.

These drawbacks are the primary motivation for developing alternative loading techniques, such as **reflective DLL loading**. The goal of reflective loading is to achieve the same end result — getting a DLL's code mapped and functional within a process's memory — but _without_ relying on `LoadLibrary` and thus avoiding the mandatory file-on-disk requirement and the standard, easily monitored OS loading procedures.

In the next chapter, we will begin exploring the PE file format in more detail, laying the groundwork necessary to understand how one might manually replicate the actions of the Windows loader, which is the core idea behind reflective loading.

For now however let's do two practical labs to better understand what we've been discussing thus far.

___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "intro_shellcode.md" >}})
[|NEXT|]({{< ref "create_dll.md" >}})