---
showTableOfContents: true
title: "The DLL Entry Point (Theory 5.1)"
type: "page"
---
## Overview

After successfully mapping the DLL image, fixing base relocations, and resolving all imported function addresses by patching the IAT, our manually loaded DLL is finally prepared structurally and contextually to run code. Just as we have a `main()` function entrypoint in a Go applications (and most languages for that matter), so too a DLL has a standard entry point function known as `DllMain`.

Note that unlike a `main()` function in a typical application, `DllMain` is **optional**. If present, the Windows loader calls this function automatically at specific times to notify the DLL about four key events (listed below).


## Structure

For reference here is `Dllmain` from our very first lab... (What a long way we've come!)
```C
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


**`BOOL WINAPI`**: This specifies the function's return type and calling convention.
- **`BOOL`**: Success (1) or Failure (0)
- **`WINAPI`**: This is a macro that defines the calling convention for the function


Inside of the function, inside of the switch statement, we have 4 conditions which may be triggered:
- **Process Attach (`DLL_PROCESS_ATTACH`):** The DLL is being loaded into (attaching to) a process's address space for the first time. This is the most common place for a DLL to perform one-time initializations.
- **Process Detach (`DLL_PROCESS_DETACH`):** The DLL is being unloaded from (detaching from) a process (e.g., due to `FreeLibrary` being called when the reference count hits zero, or because the process is terminating). This is where DLLs can perform cleanup tasks.
- **Thread Attach (`DLL_THREAD_ATTACH`):** A new thread is being created within the process _after_ the DLL has already been loaded.
- **Thread Detach (`DLL_THREAD_DETACH`):** A thread within the process is exiting cleanly _while_ the DLL is still loaded.


As mentioned above, a DLL does not require a `DllMain`. If a DLL doesn't require any specific initialization or cleanup tied to these events, it can omit the function entirely. In such cases, the `AddressOfEntryPoint` field in the PE Optional Header will be zero. Our `calc_dll.dll` includes a `DllMain`, but it doesn't perform any actions within its `switch` statement, effectively making it a placeholder.

## Signature

If a DLL implements `DllMain`, the function must adhere to a specific signature defined by the Windows API:

```C
BOOL WINAPI DllMain(
    HINSTANCE hinstDLL,     // Handle to DLL module (actually the base address)
    DWORD     fdwReason,    // Reason for calling function
    LPVOID    lpvReserved   // Reserved
);
```

- **`hinstDLL` (HINSTANCE):** For a DLL loaded reflectively, this parameter should be the actual base address where the DLL was mapped in memory (the `ActualAllocatedBase` we obtained from `VirtualAlloc`). This allows code inside `DllMain` to calculate absolute addresses relative to its own loaded position if needed (e.g., for accessing resources).
- **`fdwReason` (DWORD):** This value indicates _why_ `DllMain` is being called. It will be one of the constants mentioned earlier (`DLL_PROCESS_ATTACH`, `DLL_PROCESS_DETACH`, `DLL_THREAD_ATTACH`, `DLL_THREAD_DETACH`). For the initial call after loading, we use `DLL_PROCESS_ATTACH`.
- **`lpvReserved` (LPVOID):** This parameter provides additional context. It's typically `NULL` for dynamic loads (like reflective loading or calls via `LoadLibrary`). It can be non-NULL during static loading or process termination under certain circumstances, but for our reflective call, passing `0` is appropriate.

## Reflective Call to DllMain

After the IAT has been successfully patched (as we just did in Lab 4.2), the reflective loader can attempt to call `DllMain` to allow the DLL to initialize itself, mimicking the behavior of the standard Windows loader.

The process is:
1. **Find Entry Point RVA:** Get the `AddressOfEntryPoint` value from the `IMAGE_OPTIONAL_HEADER` of the mapped DLL (which resides at `ActualAllocatedBase`).
2. **Check if Entry Point Exists:** If `AddressOfEntryPoint` is zero, the DLL does not have a `DllMain`, so skip the call and proceed to the next step (like calling a specific exported function).
3. **Calculate Entry Point VA:** If the RVA is non-zero, calculate the absolute virtual address of `DllMain`: `DllMainVA = ActualAllocatedBase + AddressOfEntryPoint`
4. **Call `DllMain`:** Use a mechanism like Go's `syscall.SyscallN` to execute the code at `DllMainVA`. Pass the required arguments according to the signature:
    - Argument 1 (`hinstDLL`): `ActualAllocatedBase` (cast to `uintptr`).
    - Argument 2 (`fdwReason`): `DLL_PROCESS_ATTACH` (constant value 1, cast to `uintptr`).
    - Argument 3 (`lpvReserved`): `0` (cast to `uintptr`).
5. **Check Return Value:** `DllMain` returns a `BOOL` (non-zero for TRUE, zero for FALSE). When called with `DLL_PROCESS_ATTACH`, returning `FALSE` signals that the DLL failed to initialize. A well-behaved loader (standard or reflective) should typically treat a `FALSE` return during `DLL_PROCESS_ATTACH` as a fatal error, abort the loading process, and potentially unload the DLL or terminate. If `DllMain` returns `TRUE`, initialization succeeded, and the loader can proceed.

Calling `DllMain` correctly allows the reflectively loaded DLL to perform any necessary setup before its exported functions are used, ensuring behavior consistent with standard loading practices.

## Conclusion

Once `DllMain` (if present) has been called successfully, the DLL is fully initialized and ready. The next step is to locate and call a specific _exported_ function within the DLL to trigger its main payload or functionality, in our case this is of course `LaunchCalc()`.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module04/iat_lab.md" >}})
[|NEXT|]({{< ref "export.md" >}})