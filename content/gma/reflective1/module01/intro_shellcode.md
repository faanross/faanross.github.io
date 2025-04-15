---
showTableOfContents: true
title: "Introduction to Shellcode"
type: "page"
---

## Introduction to Shellcode
### Overview

While DLLs are standard components for legitimate software development, the term **shellcode** is more related to the world of
exploitation and malware development.

**Shellcode** refers to a small piece of self-contained machine code (native CPU instructions) typically designed to be injected 
and executed in the memory space of a target process. The name originates from its historical use: one of the earliest and 
most common goals after exploiting a vulnerability on a Unix-like system was to gain a command shell (like `/bin/sh`) on the 
target machine. Thus, the payload code was called _shellcode_ - code that lets you "pop a shell". 

Today, the term is used more broadly to describe any payload injected and executed in this manner, regardless of whether 
its goal is to actually launch a command shell. Its purpose might be to:
- Download and execute a larger piece of malware (to go from stager to full implant).
- Download and execute a specific tool (for ex [Mimikatz](https://attack.mitre.org/software/S0002/) for credential dumping).
- Create a reverse connection back to a server (create C2 connection).
- Modify system settings (for persistence etc).
- Simply demonstrate successful control over the instruction pointer (e.g., by launching a harmless application like `calc.exe`).

### Key Characteristics of Shellcode

- **Position-Independent:** Often, shellcode must be designed to run correctly regardless of where it ends up in memory. It cannot rely on absolute memory addresses for its internal jumps or data access, instead using relative addressing.
- **Self-Contained:** It usually cannot rely on standard library functions being readily available in the way normally compiled code can. It often needs to manually find necessary system functions (like those in `kernel32.dll`) by parsing OS structures in memory or by using direct system calls.
- **Compact:** Especially in exploit scenarios, the available space for injected code might be very limited, demanding efficiency.
- **Avoid Null Bytes (Often):** In certain exploit types (like string buffer overflows), null bytes (`0x00`) can terminate the string prematurely, truncating the injected shellcode. Therefore, shellcode often needs to be crafted using only instructions whose machine code representation does not contain null bytes.

### Example: Launching `calc.exe`

A very common "proof-of-concept" payload for demonstrating successful code execution is shellcode that simply launches the Windows Calculator (`calc.exe`), which is exactly what we'll be doing in this course. While harmless, it provides visual confirmation that the attacker (or developer testing an injection technique) was able to execute arbitrary code on the target system.

So we can create a cpp source file that we will compile into a DLL which contains a byte array with our shellcode. These bytes represent the actual x64 machine instructions that, when executed directly by the CPU, will perform the necessary steps using Windows API functions (found dynamically) to launch the calculator process.

Aside from the dll containing the actual byte array, we'll also need a specific function that "does the work" of executing it. There are many ways to do this to varying degrees of sophistication (rooted in the ever present cat-and-mouse game to avoid AV/EDR), but in the simplest possible iteration it involves allocating executable memory, copying the shellcode bytes into it, and then jumping to that memory, causing the CPU to execute the shellcode instructions.

Understanding shellcode is crucial because the ultimate payload delivered by a loader is often shellcode, or other "more advanced forms" (mainly BOFs and COFFs), which we will cover elsewhere.

---