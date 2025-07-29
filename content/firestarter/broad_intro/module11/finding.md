---
showTableOfContents: true
title: "Finding Native API Functions (Theory 11.2)"
type: "page"
---
## Overview

In the previous lesson, we established that `ntdll.dll` serves as the primary interface between user-mode applications and the Windows kernel, exporting the low-level Native API functions. We also discussed the motivation for calling these functions directly – primarily to bypass user-mode hooks placed on higher-level WinAPI functions in libraries like `kernel32.dll`.

We then also briefly discussed where/how to find info on their use, and then I gave a high-level overview on how to use them, including the second step - finding the function address. In this lesson we'll explore exactly how to do this in more detail.

Unlike our own reflectively loaded DLL where we manually parsed the export table, for system DLLs like `ntdll.dll` that are already loaded into our process, we can leverage standard Windows API functions to find function addresses.

## `ntdll.dll`: Always Present

A key fact simplifies our task: `ntdll.dll` is a fundamental component of the Windows user-mode environment. It is **loaded into the address space of every user-mode process** during its initialization by the Windows loader, long before our own code starts executing. This means we don't need to manually load `ntdll.dll`; we can safely assume it's already present in our process's memory map.

## Getting a Handle: `GetModuleHandleW`

Since `ntdll.dll` is already loaded, we can obtain a handle to it using the `GetModuleHandleW` function (the `W` denotes the Unicode character version).

```c++
HMODULE GetModuleHandleW(
  LPCWSTR lpModuleName
);
````

- **`lpModuleName`**: A pointer to a null-terminated string specifying the module name (e.g., `"ntdll.dll"`). If this parameter is `NULL`, it returns a handle to the _calling process's_ executable file itself.

So let's look at the shape of actually using the function:
```cpp
HMODULE hNtdll = GetModuleHandleW(L"ntdll.dll");
if (hNtdll == NULL) {
    // This should almost never happen for ntdll.dll in a running process
    // Handle error - indicates a serious problem
} else {
    // hNtdll now holds the base address where ntdll.dll is loaded
    // in the current process's virtual address space.
}
```

Calling `GetModuleHandleW(L"ntdll.dll")` returns the **base address** where `ntdll.dll` is mapped in the current process's virtual memory. This base address also serves as the module handle (`HMODULE`) required by `GetProcAddress`.

## Finding Functions: `GetProcAddress`

Once we have the handle (base address) of `ntdll.dll`, we can find the address of any function _exported_ by it using the familiar `GetProcAddress` function.


```c++
FARPROC GetProcAddress(
  HMODULE hModule,    // Handle to the DLL module (from GetModuleHandleW)
  LPCSTR  lpProcName  // Function name (ANSI string)
);
```

- **`hModule`**: The handle to the DLL module where the function resides (in our case, the handle to `ntdll.dll` obtained from `GetModuleHandleW`).
- **`lpProcName`**: A null-terminated **ANSI** string containing the name of the function to find (e.g., `"NtAllocateVirtualMemory"`). Note that `GetProcAddress` typically uses ANSI strings for the function name, even when using `GetModuleHandleW`.


```c++
// Assume hNtdll was obtained successfully via GetModuleHandleW

FARPROC pNtAllocateVirtualMemory = GetProcAddress(hNtdll, "NtAllocateVirtualMemory");
if (pNtAllocateVirtualMemory == NULL) {
    // Function not found or error occurred. Check GetLastError().
} else {
    // pNtAllocateVirtualMemory now holds the VA of the NtAllocateVirtualMemory function
    // We need to cast it to the correct function pointer type before calling.
}
```

By calling `GetProcAddress` with the `ntdll` handle and the name of the desired Native API function, we have now obtained its absolute virtual address within our process.

## Conclusion

Finding the addresses of Native API functions within the already-loaded `ntdll.dll` is relatively simple using the standard `GetModuleHandleW` and `GetProcAddress` WinAPI functions. This of course assumes however we've already done the relatively harder part of determining the exact function signature (parameters, return type, calling convention) as we outlined in the previous lesson.

Once we have determined the address we can then proceed to call `ntdll` functions, bypassing potential hooks at the `kernel32.dll` layer. In the next lesson, we will practice calling these functions from Go.




---
