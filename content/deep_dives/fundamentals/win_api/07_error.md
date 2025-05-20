---
showTableOfContents: true
title: "Error Handling and Debugging"
type: "page"
---

## Windows API Error Handling

We already briefly covered error handling in the previous, or at least I showed one example of how to do it. The thing to note however is that Windows actually has a number of different ways how to handle errors. Unfortunately, strange seemingly non-sensical exceptions, and contradictions is just something you'll need to get used to working with the Windows API. It's obviously a multi-decade year old codebase, worked on by different generations of people, and there was often a tension between a need to improve things, but also maintaining backward compatibility. It's just the way it is.

**There are 3 main types of errors returned by Windows:**
- Many functions return a boolean value (`TRUE` for success, `FALSE` for failure)
- Some functions return `NULL` or `INVALID_HANDLE_VALUE` on failure
- Others return specific error codes directly

A cool thing about the last-mentioned error return type is that we can follow up and use

When a function indicates failure through its return value, additional error information is typically available through the `GetLastError` function, which returns a system error code identifying the specific reason for the failure. But we have to call the function immediately after the failed API call, any subsequent API call may reset the error code.

For example, proper error handling for the `CreateFile` function might look like this:
```c
HANDLE hFile = CreateFile(fileName, GENERIC_READ, ...);
if (hFile == INVALID_HANDLE_VALUE) {
    DWORD error = GetLastError();
    switch (error) {
        case ERROR_FILE_NOT_FOUND:
            printf("The specified file does not exist.\n");
            break;
        case ERROR_ACCESS_DENIED:
            printf("Access to the file was denied.\n");
            break;
        default:
            printf("Failed to open file. Error code: %lu\n", error);
            break;
    }
    // Handle the error appropriately
}
```

The Windows API also provides a few functions to help convert error codes into human-readable messages:

- `FormatMessage` can convert a system error code into a descriptive text message
- `GetLastError` retrieves the most recent error code for the calling thread
- `SetLastError` can be used by custom functions to set specific error codes

As with error handling, it's not really just about detecting errors, but creating logic that dictates how the program should continue should an error arise. It usually depends of course on the type of operation, probability of failure, as well as implications of failure. So we may want to retry the operation, attempt an alternative approach, simplify notify the user and continue undeterred, or gracefully terminate the application.

## Native API Error Handling

Error handling for the Native API (`NTAPI`) functions differs significantly from the standard Windows API approach. While Windows API functions use `GetLastError()` to provide error information, Native API functions return error codes directly through `NTSTATUS` values.

An `NTSTATUS` is a 32-bit value where zero (`STATUS_SUCCESS`) indicates successful execution, and non-zero values represent various error conditions. These values follow a structured format where different bits indicate the severity, customer code, facility, and specific error code. So the code itself communicates detailed information directly, meaning we no longer need to use a separate function like `GetLastError()`.

NTAPI error handling typically follows this pattern:
```c
NTSTATUS status = NtCreateFile(&fileHandle, ...);
if (!NT_SUCCESS(status)) {
    // Handle the error based on the specific status code
    printf("NtCreateFile failed with status: 0x%08X\n", status);
    // Take appropriate action based on the status code
}
```

The `NT_SUCCESS` macro simplifies checking for success, returning `TRUE` if the status code indicates success and FALSE otherwise. Additional macros like `NT_INFORMATION`, `NT_WARNING`, and `NT_ERROR` help categorize status codes by severity.

Microsoft provides documentation for common NTSTATUS values, though many values remain undocumented or are only documented indirectly. The NTSTATUS.H header file contains definitions for many common status codes, providing symbolic names that improve code readability compared to hexadecimal values. For example, instead of checking for `status == 0xC0000022`, code can use the more readable `status == STATUS_ACCESS_DENIED`.


## Debugging Techniques

**Beyond basic error checking, Windows provides several built-in debugging aids:**

1. **Debug Output**: The `OutputDebugString` function sends a string to the debugger for display, allowing applications to emit diagnostic information that doesn't interfere with normal operation.
2. **Debug Heap**: The Windows heap manager includes special debugging features that can be enabled to detect memory corruption and leaks. These can be activated through application manifest settings or environment variables.
3. **Windows Event Log**: Applications can write structured diagnostic information to the Windows Event Log, providing a persistent record of application behaviour and errors.

**Specialized debugging tools further enhance the debugging process:**

1. **Debuggers**: Tools like WinDbg and Visual Studio's debugger allow us to set breakpoints, inspect memory, and step through code execution to identify issues.
2. **API Monitors**: Tools like API Monitor and Process Monitor track API calls made by an application, showing parameters, return values, and timing information.
3. **Memory Analysis Tools**: Applications like VMMap and RAMMap help identify memory usage patterns and potential leaks.
4. **ETW (Event Tracing for Windows)**: This framework allows high-performance logging of system and application events, with tools like Windows Performance Analyzer (WPA) providing visualization and analysis capabilities.



---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "06_functions.md" >}})
[|NEXT|]({{< ref "08_security.md" >}})