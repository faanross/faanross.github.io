---
showTableOfContents: true
title: "Introduction to Shellcode (Theory 1.2)"
type: "page"
---

## Overview

While DLLs are standard components for legitimate software development, the term **shellcode** is more related to the world of
exploitation and malware development.

**Shellcode** refers to a small piece of self-contained machine code (native CPU instructions) typically designed to be injected 
and executed in the memory space of a target process. The name originates from its historical use: one of the earliest and 
most common goals after exploiting a vulnerability on a Unix-like system was to gain a command shell (like `/bin/sh`) on the 
target machine. Thus, the payload code was called _shellcode_ - code that lets you "pop a shell". 

## Shellcode Uses

Today, the term is used more broadly to describe any payload injected and executed in this manner, regardless of whether 
its goal is to actually launch a command shell. Its purpose might be to:
- Download and execute a larger piece of malware (to go from stager to full implant).
- Download and execute a specific tool (for ex [Mimikatz](https://attack.mitre.org/software/S0002/) for credential dumping).
- Create a reverse connection back to a server (C2 connection).
- Modify system settings (for persistence etc).
- Simply demonstrate successful control over the instruction pointer (e.g., by launching a harmless application like `calc.exe`).

## Key Characteristics of Shellcode

- **Position-Independent:** Often, shellcode must be designed to run correctly regardless of where it ends up in memory. It cannot rely on absolute memory addresses for its internal jumps or data access, instead using relative addressing.
- **Self-Contained:** It usually cannot rely on standard library functions being readily available in the way normally compiled code can. It often needs to manually find necessary system functions (like those in `kernel32.dll`) by parsing OS structures in memory or by using direct system calls.
- **Compact:** Especially in exploit scenarios, the available space for injected code might be very limited, demanding efficiency.
- **Avoid Null Bytes (Often):** In certain exploit types (like string buffer overflows), null bytes (`0x00`) can terminate the string prematurely, truncating the injected shellcode. Therefore, shellcode often needs to be crafted using only instructions whose machine code representation does not contain null bytes.

## Example: Launching `calc.exe`

A very common "proof-of-concept" payload for demonstrating successful code execution is shellcode that simply launches the Windows 
Calculator (`calc.exe`), which is exactly what we'll be doing in this course. 
While harmless, it provides visual confirmation that we are able to execute arbitrary code on the target system.
So essentially we're stating - since we can arbitrarily execute `calc.exe` via shellcode, we've proven to ourselves that we can 
arbitrarily execute any shellcode, whatever its purpose. 

## How to Package + Execute Shellcode

There are many different forms/patterns one can use to execute shellcode, here are some of the main ones:

### Executable Wrapper/Loader
In this pattern, the shellcode is typically embedded, often in an encrypted or encoded form, 
directly within a standard executable file (.exe). When this executable is launched, 
its primary code executes steps to prepare and run the hidden shellcode. 
This is an "all-in-one" approach - it  uses the .exe both as a self-contained delivery and execution mechanism for the shellcode.

### Script-Based Loader
Here we embed the shellcode, commonly encoded as a large string or byte array, 
within a script file such as PowerShell, VBScript, JScript, or Python. 
When the corresponding interpreter runs the script, the script leverages operating system API calls 
accessible from the scripting environment. 

### Document Macro Loader
Similar in principle to script-based loaders, this method hides the encoded shellcode and the necessary execution 
logic within macros of documents, typically Microsoft Office files like Word or Excel. 
Execution begins when a user opens the document and explicitly enables macros. The embedded VBA (Visual Basic for Applications) 
code then runs within the host Office application's process space. Because it presents itself as an innocuous document it is 
of course one of the popular methods for phishing attacks, with the intention of executing a stager which is then tasked
to call back, download, inject into memory, and execute the full C2 implant. 

### Direct Exploit Injection (Classic Vulnerability Exploitation)
Here, the shellcode is included as part of the data payload sent to a target application to trigger a software flaw, 
like a buffer overflow or use-after-free. The exploit carefully crafts this data not only to trigger the bug but also to overwrite 
crucial process control data, such as a function's return address on the stack or a function pointer stored in memory. 
The shellcode itself is placed somewhere within the process's memory by the exploit payload. When the corrupted control data 
is subsequently used by the program, instead of continuing normal execution, the process's instruction pointer is redirected to the 
location of the injected shellcode, causing it to run directly within the compromised process context.

### Library (DLL) Loader
This popular pattern technique has many variations, all which typically involve the combination of two core components:
a loader process and a Dynamic Link Library (.dll). The DLL primarily serves as a container, which might hold the raw shellcode payload 
itself, the functional code (exported DLL functions) required to execute the shellcode, or often both. Please note: this is a 
simplified explanation, there are many nuanced variations and exceptions that can be derived from this technique. For now
let's just focus on the "traditional pattern", in due time we'll cover all variations. 

The loader process (which could be an executable, script, etc.) on the other hand is responsible for initiating the action. 
It causes the DLL to be loaded into a target memory space – either the loader's own process or another target process via 
injection techniques.

The core execution logic – the steps responsible for taking the raw shellcode bytes, placing them into executable memory 
(e.g., using VirtualAlloc), and then transferring execution control to them (e.g., using CreateThread or a function pointer call) 
– can reside in one of two primary places:

- Inside the DLL: An exported function within the DLL contains this execution logic. The loader's main job after loading the DLL 
is simply to find and call this specific function.

- Inside the Loader: The loader itself contains this execution logic. After loading the DLL, the loader extracts the raw shellcode 
data from the DLL's mapped memory space (e.g., from its resources) and then performs the allocation, copying, and execution steps itself. 
In this variation, the DLL acts more like a passive data file once loaded.

## References

[Shellcode Execution with GoLang - Joff Thyer](https://www.youtube.com/watch?v=gH9qyHVc9-M)

[pwn.college - Shellcode Injection - Introduction](https://www.youtube.com/watch?v=715v_-YnpT8)

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "intro_DLLs.md" >}})
[|NEXT|]({{< ref "dll_loading.md" >}})