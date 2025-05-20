---
showTableOfContents: true
title: "Calling Native API Functions (Theory 11.3)"
type: "page"
---
## Overview

In the previous lab we successfully used `GetModuleHandleEx` and `GetProcAddress` to obtain the memory addresses of key Native API functions exported by `ntdll.dll`. So we now have the function pointers, but of course simply having the address isn't enough.

We need a way to execute the code at that address, passing the correct arguments in the correct way, and interpreting the return value correctly. In this lesson we'll explore the mechanics of calling these lower-level functions from Go, highlighting the challenges and common approaches.

## The Calling Challenge: Signatures and Wrappers

The `golang.org/x/sys/windows` package, which we've used extensively, provides convenient Go wrappers for many *documented* Win32 API functions (like `VirtualAllocEx`, `CreateRemoteThread`, etc.). These wrappers handle the details of converting Go types (like `string` or `[]byte`) into the C-style pointers and types expected by the API, and they often translate error codes (`GetLastError`) into Go `error` values.

However, this package **does not** typically provide wrappers for the lower-level, often undocumented or partially documented, Native API functions in `ntdll.dll`. So, if we want to call a function like `NtAllocateVirtualMemory` whose address we found via `GetProcAddress`, we cannot rely on a pre-built wrapper in `golang.org/x/sys/windows`, we need a more direct method.

## Using Go's `syscall` Package

Go's built-in `syscall` package provides us with the necessary tools for making low-level operating system calls, including calling arbitrary function pointers. The key functions are `syscall.Syscall`, `syscall.Syscall6`, `syscall.Syscall9`, ..., `syscall.SyscallN`.

The most general form is `syscall.SyscallN`:

```go
func SyscallN(trap uintptr, args ...uintptr) (r1, r2 uintptr, err Errno)
````


### Arguments
- **`trap uintptr`**: This is the **address of the function** we want to call, meaning of course it's the the `procAddr` we obtained from `windows.GetProcAddress` in the previous lab, cast to `uintptr`.
- **`args ...uintptr`**: A variable number of arguments to pass to the function. **Crucially, ALL ARGUMENTS must be passed as `uintptr`**. This often requires using the `unsafe` package to convert Go pointers (like `*MyStruct` or `&myVariable`) to `uintptr`.
    - `uintptr(unsafe.Pointer(&myVariable))`
    - `uintptr(myPointer)`
    - Simple integer types often need casting: `uintptr(myIntValue)`, `uintptr(myHandle)`
    - Pointers to buffers: `uintptr(unsafe.Pointer(&myByteBuffer[0]))`
    - `NULL` pointers are passed as `uintptr(0)`.

### Return Values
- `r1`, `r2`: These hold the primary and secondary return values from the system call, respectively. For most Native API functions, `r1` contains the `NTSTATUS` result. `r2` is often unused or holds secondary OS-specific information.
- `err Errno`: This captures any error that occurred _during the syscall mechanism itself_ (e.g., invalid parameters passed to `SyscallN`, not necessarily the logical result of the Native API function). If the syscall mechanism succeeded, `err` will typically be 0 (which corresponds to `ERROR_SUCCESS`).



## Example

```go
// Assume we have:
// ntAllocateVirtualMemoryAddr uintptr // Address from GetProcAddress
// hProcess windows.Handle         // Target process handle (-1 for current)
// baseAddress uintptr             // Pointer to receive base address
// zeroBits uintptr                 // Usually 0
// regionSize uintptr              // Pointer to region size
// allocationType uint32           // MEM_COMMIT | MEM_RESERVE
// protect uint32                  // PAGE_READWRITE

// IMPORTANT: Argument count MUST match the function signature.
// NtAllocateVirtualMemory takes 6 arguments.
ntstatus, _, errno := syscall.SyscallN(ntAllocateVirtualMemoryAddr,
    uintptr(hProcess),                          // ProcessHandle
    uintptr(unsafe.Pointer(&baseAddress)),      // BaseAddress (output)
    uintptr(zeroBits),                          // ZeroBits
    uintptr(unsafe.Pointer(&regionSize)),       // RegionSize (input/output)
    uintptr(allocationType),                    // AllocationType
    uintptr(protect),                           // Protect
    // Add dummy 0s if the function takes more args than Syscall6 allows
)

// Check syscall mechanism error FIRST
if errno != 0 {
    log.Fatalf("SyscallN error: %v", errno)
}

// Check NTSTATUS logical result
if ntstatus != 0 { // 0 is STATUS_SUCCESS
    log.Fatalf("NtAllocateVirtualMemory failed with NTSTATUS: 0x%X", ntstatus)
}

// If we reach here, the call succeeded. baseAddress and regionSize contain results.
fmt.Printf("Successfully allocated memory at 0x%X, size %d\n", baseAddress, regionSize)

```

Using `syscall.SyscallN` requires meticulous attention to the function's signature: the exact number, order, and type of arguments are critical. All arguments must be correctly converted to `uintptr`.

## Defining Native Structures in Go

Many Native API functions require pointers to specific structures (like `OBJECT_ATTRIBUTES` or `UNICODE_STRING`). Since these aren't typically predefined in standard Go packages, we must **define corresponding Go structs** that exactly match the memory layout (field order, types, and alignment) of the C structures.


```go
// Example: Minimal UNICODE_STRING definition in Go
type UnicodeString struct {
    Length        uint16
    MaximumLength uint16
    Buffer        *uint16 // PWSTR - Pointer to wide char buffer
}

// Example: Minimal OBJECT_ATTRIBUTES definition in Go
type ObjectAttributes struct {
    Length                   uint32
    RootDirectory            windows.Handle
    ObjectName               *UnicodeString
    Attributes               uint32
    SecurityDescriptor       *byte // PVOID
    SecurityQualityOfService *byte // PVOID
}

// --- How to use ---
var objName UnicodeString
var objAttr ObjectAttributes

// Need to allocate buffer for objName.Buffer, copy string data, set lengths...
// (Manual initialization required, similar to C)

// Initialize objAttr
objAttr.Length = uint32(unsafe.Sizeof(objAttr))
objAttr.ObjectName = &objName
// ... set other fields as needed (often to 0/nil for simple cases)

// Pass pointer to the function via SyscallN
// ... syscall.SyscallN(..., uintptr(unsafe.Pointer(&objAttr)), ...)
```

Getting these structure definitions correct is vital and often requires consulting the "unofficial documentation" references I provided earlier. The `unsafe.Sizeof` function is useful for setting `Length` fields correctly.

## Handling `NTSTATUS` Return Values

As mentioned, Native API functions usually return an `NTSTATUS` code in `r1`. `0` (`STATUS_SUCCESS`) indicates success. Any non-zero value indicates an error. We should explicitly check `if ntstatus != 0` after verifying `errno == 0`. While comprehensive error handling involves mapping specific `NTSTATUS` codes to meaningful errors, for basic checks, simply ensuring the status is `0` is often sufficient during development.

## Alternative: Assembly Stubs

For complex functions, or when absolute control is needed (especially when preparing for direct syscalls later), we might write small **assembly language stubs**. These stubs can be linked with Go code using `cgo` or Go's internal assembler. The Go code calls a simple Go function prototype, which internally transfers control to the assembly stub. The stub then correctly arranges arguments in registers and on the stack according to the required calling convention, invokes the Native API function pointer, retrieves the result, and returns it to the Go caller. This abstracts the low-level calling convention details away from the main Go logic but requires knowledge of assembly language.

## Conclusion

Calling Native API functions from Go requires moving beyond the standard `golang.org/x/sys/windows` wrappers for documented WinAPI calls. The `syscall` package, particularly `syscall.SyscallN`, provides the mechanism to call arbitrary function pointers, but demands careful handling of argument types (converting everything to `uintptr`, often using `unsafe.Pointer`) and strict adherence to the target function's signature. Defining corresponding Go structs for required Native API structures is often necessary. Checking the `NTSTATUS` return value is crucial for determining the success or failure of the Native API call itself.

In the next lab, we'll practice using this `syscall.SyscallN` approach.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "finding_lab.md" >}})
[|NEXT|]({{< ref "calling_lab.md" >}})