---
showTableOfContents: true
title: "Windows API Data Types"
type: "page"
---

## The Windows Type System

The Windows API has its own specialized type system that extends beyond the basic data types found in standard C/C++. These Windows-specific types serve several purposes:
- They abstract hardware-specific details,
- Provide clearer semantic meaning, and
- Ensure consistency across the API surface.

But beyond these benefits, they are in all honestly a bit of a bummer to learn. At least in the beginning - there's now a whole new type system you need to get used to just for the purposes of interacting with the Windows API. But, mastering them is essential for effective Windows system programming, and so we do as we must.

## Conventions

The Windows type system uses uppercase naming conventions (such as `DWORD` and `HANDLE`) to distinguish its types from standard C types. Many of these types are defined as aliases or typedefs of more basic types, with the specific definitions sometimes varying based on the target platform (32-bit vs. 64-bit). This approach allows Microsoft to adapt the API for different architectures while maintaining source compatibility.

Understanding this type system is not merely a matter of syntax—it provides valuable semantic information about how the API expects data to be used. For example, when encountering a parameter of type `HANDLE`, a developer immediately knows they're dealing with a reference to a system resource, not a direct pointer or simple integer value. This semantic clarity helps clarify the intended use of API functions.

## Core Types

**DWORD** (Double Word) is a 32-bit unsigned integer, regardless of whether the platform is x86 or x64. It can represent values between 0 and approximately 4.3 billion, making it suitable for many counters, flags, and identifiers within the Windows API. For example, `DWORD dwFlags = 0x00000001;` might be used to specify options for an API function.

**SIZE_T** is an unsigned integer whose size depends on the platform architecture—32 bits on x86 systems and 64 bits on x64 systems. This type is typically used for representing memory sizes and counts, ensuring compatibility across different architectures. The `sizeof()` operator returns results of this type, as in `SIZE_T bufferSize = sizeof(DWORD);`.

**VOID** indicates the absence of a data type, often used when declaring functions that don't return values or don't accept parameters. More powerfully, a void pointer (`void*`) can point to data of any type, making it useful for generic functions that handle different data types. This flexibility comes with responsibility, as the programmer must keep track of the actual data type being referenced.

**PVOID** (Pointer to VOID) explicitly indicates a pointer to any data type, equivalent to `void*` in C. This type is commonly used for generic memory operations and functions that need to handle different data types without knowing their specific structure. For example, `PVOID pMemory = malloc(1024);` creates a generic pointer to allocated memory.

**HANDLE** represents an opaque reference to a Windows resource such as a file, process, thread, or event. Unlike pointers, handles don't directly expose memory addresses—they're abstractions managed by the operating system to reference resources. Applications use handles to refer to these resources in subsequent API calls, as in `HANDLE hFile = CreateFile(...);`.

**HMODULE** is a specialized handle representing a loaded module (DLL or executable) in memory. In most implementations, this handle corresponds to the base address of the module in memory, though applications should treat it as an opaque identifier rather than assuming this detail. For example, `HMODULE hModule = GetModuleHandle(NULL);` retrieves a handle to the current executable.

**String Types** include various representations for character strings:

- **LPCSTR/PCSTR**: Pointers to read-only ANSI (8-bit) strings, equivalent to `const char*` in C
- **LPSTR/PSTR**: Pointers to writable ANSI strings, equivalent to `char*` in C
- **LPCWSTR/PCWSTR**: Pointers to read-only Unicode (16-bit) strings, equivalent to `const wchar_t*` in C
- **LPWSTR/PWSTR**: Pointers to writable Unicode strings, equivalent to `wchar_t*` in C

**ULONG_PTR** is an unsigned integer large enough to hold a pointer value on the current platform—32 bits on x86 systems and 64 bits on x64 systems. This type is often used for pointer arithmetic and ensures compatibility across different architectures. For example, `ULONG_PTR address = (ULONG_PTR)pSomePointer + offset;` performs arithmetic on a pointer value safely.

My suggestions would be to not sit and try to memorize each, but get used to them as you work with each. Do note however the universal pattern - each type communicates not just a size and format but also semantic meaning about how the data should be used within the API.

## Pointer Types and Naming Conventions

The Windows API follows specific naming conventions for pointer types, adding another layer of semantic information to the type system.

**When working with the Windows API, pointer types are typically indicated in one of two ways:**
1. Types beginning with 'P' (such as `PDWORD` or `PHANDLE`) represent pointers to the base type
2. Types containing "LP" (such as `LPDWORD`) also represent pointers, with the "L" prefix being a historical artifact from "long pointer" in 16-bit Windows

These naming patterns create symmetry in the type system, where each base type typically has corresponding pointer types. For example:

|Pointer Type|Equivalent C Type|
|---|---|
|PHANDLE|HANDLE*|
|PSIZE_T|SIZE_T*|
|PDWORD|DWORD*|
|LPSTR|char*|
|LPCSTR|const char*|

The distinction between these pointer types and their base types becomes particularly important when understanding function parameters. For example, a parameter of type `HANDLE` expects a handle value directly, while a parameter of type `PHANDLE` expects the address of a variable containing a handle value, allowing the function to modify that handle.

This passing of pointers to allow functions to modify values leads to another important aspect of Windows API programming: the use of "In" and "Out" parameter annotations.


## In and Out Parameters

The Windows API also provides these conventions that indicate whether parameters provide input to a function, receive output from it, or both. These conventions are documented using the annotations [in], [out], and [in, out], which provide important semantic information about how functions interact with their parameters.

An **[in] parameter** is passed to a function as input, providing information that the function needs to perform its task. The function reads from these parameters but does not modify them. For example, a file path provided to a file-opening function would be an [in] parameter.

An **[out] parameter** is used for returning data back from the function to the caller. The function writes to these parameters, typically through pointers that allow it to modify variables in the caller's scope. For example, a handle that will receive a newly created resource would be an [out] parameter.

Note: This seems strange because you might think that the return type already does this, why do we need an argument to receive data back? Many Windows API functions actually return the result - i.e. did the function succeed or not - so that if it has actual data it also wants to return, you essentially initialize an empty variable, pass it as the [out] parameter, whereafter its written to it.

An **[in, out] parameter** both provides input to and receives output from a function. The function both reads from and writes to these parameters. For example, a buffer that initially contains a request and will be overwritten with the response would be an [in, out] parameter.

Consider this example function that demonstrates an [out] parameter:

```c
BOOL HackThePlanet(OUT int* num) {
    // Setting the value of num to 123
    *num = 123;
    
    // Returning a boolean value
    return TRUE;
}

int main() {
    int a = 0;

    // 'HackThePlanet' will return true
    HackThePlanet(&a);
    // 'a' will contain the value 123 after the call
}
```


While these annotations don't affect the actual code execution — they're documentation rather than directives — they provide valuable information for developers using the API.



## ANSI vs. Unicode and Character Encoding

The Windows API system supports both ANSI and Unicode representations, again to ensure backwards compatibility.  You'llencounter this dual support throughout the API since most string-handling functions have two variants: one ending with 'A' for ANSI and another ending with 'W' for Wide (Unicode) characters.

For example, the MessageBox function exists as both `MessageBoxA` and `MessageBoxW`. The choice between these variants determines both the expected parameter types and the internal processing of string data:

- `MessageBoxA` accepts ANSI strings (char*/LPSTR) and processes text as 8-bit characters
- `MessageBoxW` accepts Unicode strings (wchar_t*/LPWSTR) and processes text as 16-bit characters

Unless you are writing software for legacy systems you will typically default to using Unicode. Even in these case you should still consider the fact that your application might be interacted with an ANSI-environment, and thus you need to integrate compatibility measures.

Note of course that the choice does not just affect the specific function you'll use, but the string literal syntax and memory requirements. Unicode strings in C/C++ use the L prefix for string literals (e.g., L"Hello World") and require twice as much memory per character compared to ANSI strings. For example:

```c
char ansiString[] = "Example";      // 8 bytes (including null terminator)
wchar_t wideString[] = L"Example";  // 16 bytes (including null terminator)
```

Modern Windows applications typically use Unicode internally, with conversion to ANSI occurring only when interacting with legacy components that don't support Unicode. The Windows API provides functions like `MultiByteToWideChar` and `WideCharToMultiByte` to facilitate these conversions when necessary.



---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "04_api.md" >}})
[|NEXT|]({{< ref "06_functions.md" >}})