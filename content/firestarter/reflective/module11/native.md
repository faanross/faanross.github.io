---
showTableOfContents: true
title: "API-Hooking and the  Native API (Theory 11.1)"
type: "page"
---
## Overview

In the previous module we explored process injection using standard, documented functions from the Windows API (WinAPI), primarily those found in libraries like `kernel32.dll` (e.g., `OpenProcess`, `VirtualAllocEx`, `WriteProcessMemory`, `VirtualProtectEx`, `CreateRemoteThread`, `LoadLibraryW`). These functions provide a relatively stable and well-documented interface for application developers to interact with the operating system, and for some time they were being reliable used by attackers to server their ends.

However like all good things, the free lunch abruptly ended with the introduction of API hooking. In the the mid-to-late 2010s, EDRs began to widely incorporate API hooking as a core detection mechanism. Before we create some greater context to see why native API calls came into the picture, we should probably explore what exactly API hooking is.

Please note that to make sense of what we're going to discuss you should have some basic understanding of the Windows Architecture. If you're unfamiliar with it, I recommend you take the time and work through this introductory guide I've created on this [exact topic]({{< ref "../../../deep_dives/fundamentals/win_api/moc.md" >}}).



## API Hooking
Modern EDRs use a technique called API hooking to monitor and protect against malware. When a program, potentially malware, tries to execute a Windows API function, the EDR intercepts this call using hooks. This allows the EDR to analyze it and determine whether it's doing possibly something suspicious.

This interception often happens as requests travel from high-level functions (like Win32 APIs, which are commonly used by applications) down to lower-level Native APIs (which interface more directly with the operating system kernel), and ultimately towards kernel operations.

By "hooking" functions at various points, the EDR gains visibility into program behaviour. So the EDR is kind of like a security guard at a checkpoint. It briefly stops the call, gives it a once-over, and then, based on a whole lot of complicated criteria we don't need to get into right now, decides if the call seems safe.

- If everything looks good, the EDR just lets the call proceed as normal.
- If it looks totally out of line, the EDR might block the call completely and could even decide to shut down the program that tried to make it.
- Or, it might be a bit of a grey area. The EDR could provisionally let the call go through but flag the process as suspicious. If that happens, it'll start keeping much closer tabs on everything that program does from then on.

So, when you boil it down, hooking is really just about inserting a point of inspection into the usual path that function calls take as they move through the Windows system. This point then allows function calls to be stopped, and calling processes to be terminated, if the behaviour is deemed suspicious.

## `ntdll.dll` and the Native API

It's important to emphasize that hooks can be placed at different junctions, most commonly between (1) the calling process and the Win32 API, as well as (2) the Win32 API and Native API. In evolutionary terms, those at (1) can be considered "original" EDR hooks. When they first came on the scene they were a big deal - they largely removed attacker's ability to interact with the win32 API directly without any risk.

Attackers then responded by simply skipping the win32 API altogether. So instead of our original process calling the Win32 API function to in turn call its Native API equivalent, the process simply calls the Native API function directly.

So it sounds easy enough conceptually, but the challenge here is that Native API functions are undocumented. Microsoft does not want you to use them, they want you to go through the Win32 API (which is documented), since this is the layer at which they can guarantee backward compatibility. The reason they don't want direct interaction with the Native API, and thus officially don't support it, is that it gives them the freedom to make changes and break backward compatibility.

So if the functions aren't documented, how can we know how to use them?


## Info on Native Functions

Even though Microsoft is mostly unwilling to release any info on how to use Native functions, there are a number of ways to find this info.


### Community Resources
First and foremost, there are a number of excellent resources that have been made available by some badass researchers.

Here are a few solid ones:
- [Geoff Chappell](https://www.geoffchappell.com/studies/windows/win32/ntdll/api/native.htm)
- [NTInternals](http://undocumented.ntinternals.net)
- [NtDoc](https://ntdoc.m417z.com)
- [Windows Native API Programming by Pavel Yosifovich](https://leanpub.com/windowsnativeapiprogramming)
- [Vergilius Project](https://www.vergiliusproject.com)



### Reverse Engineering
You can use IDA Pro, Ghidra, or even just a debugger to peek inside `ntdll.dll` and see how it works. If you're competent enough you can even observe how regular Win32 API functions (in places like `kernel32.dll`) make their calls to `ntdll.dll` functions. This allows you to deduce exactly what parameters the functions need, and what they do.

### Windows Driver Kit + Public Symbols
Despite what I've said before, some of the Native API functions _are_ documented, but for developers writing kernel-mode drivers (using the Windows Driver Kit, or WDK). The user-mode versions in `ntdll.dll` often look and behave very similarly to their kernel-mode cousins.

Further, Microsoft often releases "symbol files" (PDB files) for `ntdll.dll`, which help debuggers understand the code. These can give away function names and sometimes basic info about their parameters, even if full-blown documentation is missing.



OK, let's say you've perused some of the excellent community resources and you know the signature of the function you want to call. How do you actually go ahead and call it in Golang?


## Calling Native Function in Golang

### Get a Handle to `ntdll.dll
The Go program first needs to tell Windows it wants to use `ntdll.dll`. This is usually done with functions like `windows.LoadLibrary("ntdll.dll")` or `windows.GetModuleHandle("ntdll.dll")` from Go's `golang.org/x/sys/windows` package.


### Find the Function's Address
With `ntdll.dll` "loaded," we can then use `windows.GetProcAddress(ntdllHandle, "NameOfTheNtFunction")` to pinpoint the exact memory address where the Native API function lives (e.g., for `NtAllocateVirtualMemory` or whatever under-the-hood function we're targeting).


### Define the Function's Shape
This is a really important step - we have to create a Go function type that perfectly mirrors the C-style signature of the Native API function. This means getting the number of arguments, their data types (like integers, pointers, etc.), and the return type spot on.


### Make the Call Using `syscall` or `unsafe`
- The `golang.org/x/sys/windows` package is pretty comprehensive and actually provides ready-to-use wrappers for many common (and some not-so-common) `ntdll` functions. If the function we need is already in there, that's the simplest and safest route.
- If it's not pre-wrapped, we can also possibly use the lower-level `syscall` package, with functions like `syscall.SyscallN` (where `N` is the number of arguments: `Syscall`, `Syscall6`, `Syscall9`, etc.).
- This almost always involves using `unsafe.Pointer` to cast Go data types into shapes that the C-based Windows API understands. It's called "unsafe" for a good reason – we're essentially telling the Go compiler, "Don't worry bruh, I got this one," so if we get it wrong, things can crash or behave weirdly.

### Recreate C Structures
If the Native API function expects data to be passed in C-style structures, those structures have to be meticulously defined in Go, ensuring their memory layout is an exact match to the C version.

In general working with the Native API is more involved and riskier than calling standard, documented functions. Because these APIs aren't officially supported for this kind of use, any Go code that relies on them could break if Microsoft decides to change how these undocumented functions work in a future Windows update.



## Native API Conventions

Before we conclude this introduction section on the Native API I want to provide a few references you can use to help you understand some of the main Native API conventions.

* **`NTSTATUS` Return Values:** Many `Nt*`/`Zw*` functions return an `NTSTATUS` code (a `LONG` or `int32`) instead of a `BOOL` with `GetLastError`. `STATUS_SUCCESS` (0) indicates success. Non-zero values are error codes defined in headers like `ntstatus.h`. We need to check for `STATUS_SUCCESS` rather than just non-`NULL`/non-zero.
* **`UNICODE_STRING` Structure:** Strings are often passed using a `UNICODE_STRING` structure, which contains the length (in bytes), maximum length, and a pointer (`PWSTR`) to the actual wide character buffer. We often need to manually initialize these structures.
* **`OBJECT_ATTRIBUTES` Structure:** Functions that operate on kernel objects (files, sections, processes, threads, etc.) often take a pointer to an `OBJECT_ATTRIBUTES` structure. This structure defines object attributes like the object name (`OBJECT_NAME_INFORMATION` using a `UNICODE_STRING`), security descriptor, and flags (e.g., `OBJ_CASE_INSENSITIVE`). Proper initialization is crucial.


## Conclusion

The Native API provided by `ntdll.dll` represents a lower level of interaction with the Windows operating system compared to the standard WinAPI found in libraries like `kernel32.dll`. While more complex and often less documented, calling Native API functions directly can bypass user-mode hooks placed on their higher-level counterparts, offering a potential step up in evasion. Understanding this layering is crucial as we progress towards even lower-level techniques like direct system calls. In the next lesson, we'll look at how to dynamically find the addresses of these Native API functions within `ntdll.dll` from our Go code.







---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module10/remote_lab.md" >}})
[|NEXT|]({{< ref "finding.md" >}})