---
showTableOfContents: true
title: "Decoupling Memory Permissions (Theory 9.1)"
type: "page"
---
## Overview

We've now built a reflective loader capable of downloading, decrypting, mapping, and executing a DLL payload entirely from memory. This is good!

But, you may have noticed that, despite the fact that we've continuously worked on and evolved our loader, the actual DLL logic is literally the exact same that we created in our very first lab. Sure, we layered on a bit of obfuscation, but at the heart of our exported `LaunchCalc` function within `calc_dll.dll` we are  using `VirtualAlloc` to create RWX memory, copy shellcode into it, and then run it. This is not good!

It's like we've been going to the gym for a year and sculpting giant guns, but totally ignored our legs. There's some serious asymmetry in our design, and so the time has come to address it.

In this module we'll look at methods to improve the way we're executing our shellcode within the confines of the same overall paradigm - using functions in our DLL to interact directly with the win32 API. In future modules in this course we'll being to break free of the paradigm itself and explore completely different ways to execute our shellcode.

It's gonna we a wild and awesome ride!

## Where We're At

In our current `ExecuteShellcode` function within `calc_dll.cpp`, our first major act is to allocate the memory we want to copy our shellcode into:

```c++
    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_EXECUTE_READWRITE); // <<< The problematic part
````

You can see on the last line that we are requesting memory that is `PAGE_EXECUTE_READWRITE` - to put it bluntly, this is the noobiest of noob moves. It's  makes it utterly clear to any AV/EDR that we are up to no good.

This is because, aside from extremely specific events, legitimate processes don't allocate new memory with these permissions. Legitimate applications typically request a memory page that is _either_ writable _or_ executable, but _not both_ at the same time. That's because code (instructions) resides in executable pages (RX), and data resides in writable pages (RW or R).

It's worth unpacking in a bit more detail as it gives some good insight into the nature of memory architecture and exploitation in general.


## The Problem with RWX Memory

If one was inclined to a more grandiose way of expressing the issue at hand here, we could say that assigning RWX memory **"violates the principle of of W^X (Write XOR Execute)"**. Modern operating systems and security postures strive to enforce Data Execution Prevention (DEP), which is just the idea that a memory page should generally be _either_ writable _or_ executable, but _not both_ at the same time.

Why should this be true as a general rule? Well when a process writes code (instructions) to a region it obviously has to be readable, but when it's also writeable it becomes vulnerable. Many exploitation techniques are based on the ability of an attacker to overwrite data in memory (like a buffer) and then trick the process into executing that overwritten data as code. So when instructions are written to memory, they should not be "editable" - hence they should not be writable.

Because of this  **AV/EDRs are always on the lookout for memory regions that are RWX**. As I already mentioned, it's very rare for legitimate processes to use, but not unheard. As a consequence of their design, Just-In-Time (JIT) Compilers like the Java Virtual Machine (JVM), .NET Common Language Runtime (CLR), and Python implementations with JIT capabilities (e.g., PyPy) will create RWX memory permissions.

Regardless, any time it's assigned it will be scrutinized, and unless its use case can be justified as in the examples above, it will trigger a security alert.

So how do we improve on it? There are many ways, but for the educational value I'm not just going to tell you the final, "best" way. Instead, we'll explore each incremental improvement starting with the most basic - decoupling read and write permissions.

## The Standard, Safer Pattern: RW -> RX

Instead of immediately assigning our memory region as RWX, we assign it as RW, inject our shellcode into it, and then change the permissions to RX prior to execution.

We'll once again allocate the memory region using `VirtualAlloc`, but now with `PAGE_READWRITE` permissions only.

```cpp
// Allocate as ReadWrite first
LPVOID buffer = VirtualAlloc(NULL, shellcodeSize, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
```


Next, we'll copy the shellcode into the allocated `PAGE_READWRITE` buffer using either `memcpy`, or `RtlCopyMemory`.
```cpp
// Copy shellcode into the RW buffer
memcpy(buffer, shellcodeBytes, shellcodeSize);
// Or RtlCopyMemory(buffer, shellcodeBytes, shellcodeSize);
```


Now, before we can actually execute we have to use `VirtualProtect` (or `VirtualProtectEx` for remote processes) to change the memory protection of the buffer from `PAGE_READWRITE` to `PAGE_EXECUTE_READ` (or just `PAGE_EXECUTE` if reading isn't strictly necessary, though RX is common).

```cpp
DWORD oldProtect; // Variable to receive the previous protection flags
BOOL success = VirtualProtect(buffer, shellcodeSize, PAGE_EXECUTE_READ, &oldProtect);
// Check if VirtualProtect succeeded
```


We can now execute the code by casting the buffer address to a function pointer and calling it (or using `CreateThread`).
```cpp
// Create function pointer and call
void (*shellcode_func)() = (void(*)())buffer;
shellcode_func();
```

There is one final optional, but recommended step - cleanup. After execution finishes (assuming it returns), we'll change the memory protection back to `PAGE_READWRITE` or even `PAGE_NOACCESS` using `VirtualProtect` before freeing the memory. This minimizes the time window where executable code resides in memory. Then, we free the memory using `VirtualFree`.

```cpp
// Optional: Change back protection before freeing
VirtualProtect(buffer, shellcodeSize, PAGE_READWRITE, &oldProtect); // Or PAGE_NOACCESS
VirtualFree(buffer, 0, MEM_RELEASE);
```


## Conclusion
While decouple RWX into RW -> RX does signify a minor improvement, in all honesty it's not much. It's like we want from 99.999% probability of getting caught to 99.99%. You see assigning something as RW, and then *immediately* changing it to RX, is itself a huge red flag. So then, what was the point?

Well, notice how I said "immediately changing it"? Decoupling the events allows us now to introduce other things in between the two states so that it's not quite as immediate, thereby improving of chances of not getting caught. What are the exact things we can do? That's what we'll explore in the next lesson.



---
