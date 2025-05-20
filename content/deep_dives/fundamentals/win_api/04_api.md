---
showTableOfContents: true
title: "The Layered + Ring Models"
type: "page"
---
## The Layered Model

Understanding the Windows Architecture as a layered model with 3 essential components - Windows API, Native API, and Kernel - is a crucial concept for us as malware developers. Not only do different permissions exist around this, but a lot of security systems, like API hooking, are conceptually intertwined with this notion.

### Windows API
At the highest level is the main **Windows API**, which is also referred to as the **Win32 API**. Note that this is for historical reasons, today it also refers to 64-bit systems.  This is the documented interface - the one you are "supposed to" use, at least officially, according to Microsoft.

So when developers create applications they typically call functions in the Win32 API directly. At least traditionally that was the case, today, it's probably more common to use the .NET framework - see below.

Now depending on the type of operation the function performs, either one of two things will happen:
- If it's a **non-privileged** operation like performing calculations, or manipulating data structures in memory, then the operation is typically performed by code within the Win32 DLLs themselves.
- If it's a **privileged** operation, then the request has to be "forwarded" to the kernel via `ntdll.dll`.

### Native API
Below the Windows API is the **Native API**, implemented by `ntdll.dll`. As I just mentioned, if the Windows API wants to perform a privileged operation - let's say create a new network socket, or delete a file from disk - it *has* to go through ntdll.dll in order to "request permission" from the kernel.

This might seem odd, why do we need this bridge? I mean why did the Windows designers create it, why not just allow Win32 API function to make the request directly to the kernel? Aside from a few functional/security benefits, please recall that earlier I shared that initially there were 3 Subsystem APIs - the Windows, POSIX, and OS/2.

This design thus made much more sense in terms of historical context, since it would have been a poor design choice to let each of the 3 subsystems interface with the kernel directly, `ntdll.dll` was created as the sort of "universal bridge" to the kernel. Though the other 2 subsystems eventually went by the wayside, the overall design was maintained because it still has benefits (mostly related to security), and changing it would break backwards compatibility.

So whereas the Windows API function signatures are documented, those in the Native API are undocumented. At least officially - Microsoft does not share them since they want to retain the right to be able to change them in a way that can break backward compatibility. So essentially - you *can* use them, but if they stop working Microsoft's stance is that you cannot complain since you were not "supposed to" be using them.

So if they are undocumented, how do we know how to use them? Meaning, how do we know what arguments they take, what their return values and types are etc? Well, as I said they are officially undocumented by Microsoft, but you can find unofficial documentation quite easily online. There are other ways too, but that's a discussion for another time and place.

### System Service Dispatch Table (SSDT)
Before we get to the kernel, let's discuss the **System Service Dispatch Table (SSDT)**. While not a software layer itself, the SSDT is a critical internal kernel structure. The Native API in `ntdll.dll` relies on this mechanism for its calls to reach the correct kernel services. When a Native API function prepares to call its "kernel equivalent," it doesn't identify the service by name for the kernel.

Instead, it uses a specific **System Service Number (SSN)**. When the system call (syscall) instruction is executed, the SSN provided by `ntdll.dll` is then used by the kernel's system call dispatcher. This dispatcher uses the SSN as an index into the SSDT, which essentially maps these numbers to the actual memory addresses of the specific kernel functions. This lookup directs the execution flow to the correct service routine within kernel mode.

### Kernel
Once the system call is made and the System Service Number (SSN) has been used to identify the target kernel routine via the SSDT, control officially transfers from user mode to **kernel mode**. At this point, the code that's executing is no longer part of your application or even `ntdll.dll`; it's core operating system code, typically running within the main kernel image (`ntoskrnl.exe`) or sometimes other kernel-mode drivers if the request pertains to specific hardware or a loadable kernel module.

It's important to note that for privileged operations initiated via system calls, it is the **kernel itself that executes the core, sensitive part of the task directly in kernel mode.** It has to. User-mode code lacks the necessary privileges to, for example, directly manipulate hardware controller registers, modify protected system memory, or interact with the internal structures of the file system drivers.

If the request was to create a file, kernel-mode code will engage with the file system drivers, perform security access checks based on your process's token, allocate necessary kernel resources, and update internal system data structures. The kernel performs these actions because it is the trusted part of the operating system with the authority and capability to manage system-wide resources and enforce security boundaries.

After the kernel has completed the requested operation (like actually creating the file object and returning a handle, or modifying a system setting), it prepares any results or status codes. This information is then passed back to the `ntdll.dll` function that initiated the system call, and the processor transitions from kernel mode back to user mode.

### .NET framework
Though it's not traditionally part of the layered model per se, the **.NET Framework** is an increasingly popular, and common, way to interface with the Windows OS both for "regular" and malware developers. So I thought it is worth mentioning where it sits conceptually in relationship to the 3 core layers.

The **.NET Framework**  essentially adds another layer of abstraction on top of the Win32 API. Many of its functions that interact with the operating system—for instance, for file input/output or network communication—are essentially wrappers around the underlying Win32 API functions. So, when a .NET application performs such an operation, the call often flows from the .NET Framework's libraries to the appropriate Win32 API, which then follows the exact same path we just outlined. So it just adds another step before the Win32 API call.

But why does it exist - why another layer of abstraction? The Win32 API can, for lack of a better term, be a real pain in the ass to work with. It has a certain archaic feel to it, and so the .NET Framework was created, in part, to provide a more modern and productive way for developers to build applications for Windows.





## Let's Trace a Privileged Function Call

Let's quickly review a concrete example of how a privileged function call takes, specifically using the popular function `OpenProcess()`.


When `OpenProcess()` is called, `kernel32.dll` first processes the request in user mode. It may validate the parameters passed by the application, such as the process ID of the target process and the desired access rights.

Once these have been validated, `OpenProcess` then calls the corresponding function within the **Native API** and is typically named with an "Nt" prefix, in this case, **`NtOpenProcess`**.


Inside `ntdll.dll`, the `NtOpenProcess` function prepares to transition to kernel mode. It places a specific value, the **System Service Number (SSN)** for `NtOpenProcess`, into a designated processor register.  Then, `NtOpenProcess` executes a special CPU instruction (like `syscall` on modern systems) to trigger the **system call**.

Note that they exact SSN for `NtOpenProcess` (and other Native API functions) is not static; it often changes between different Windows versions and even updates. This variability helps maintain OS compatibility as functions evolve and also complicates attempts by software to directly use system calls with hardcoded SSNs.

Once the `syscall` instruction executes control is transferred to the kernel's system service dispatcher (a routine within `ntoskrnl.exe`). The dispatcher uses the SSN to locate and execute the actual kernel implementation of `NtOpenProcess`. This kernel function performs the necessary security checks based on the calling process's privileges and the requested access, and if permitted, it creates a handle to the target process.

After the kernel-mode operation completes, the processor switches back to user mode, returning control and the result to the `NtOpenProcess` function within `ntdll.dll`. `NtOpenProcess` in `ntdll.dll` then passes this result back to its caller, the `OpenProcess` function in `kernel32.dll`.

Finally, `kernel32.dll`'s `OpenProcess` function returns the handle (or indicates an error, often by returning `NULL` and setting an error code retrievable via `GetLastError()`) to the original application code.

**So, in a simplified manner, the journey for a function like OpenProcess that performs a privileged action is:**
Application Call (Win32 API `OpenProcess` in `kernel32.dll`) **→**
Native API (`NtOpenProcess` in `ntdll.dll`) **→**
System Call (using an SSN) **→**
Kernel (`ntoskrnl.exe`) **→**
Return to Native API (`ntdll.dll`) **→**
Return to Win32 API (`kernel32.dll`) **→**


## The Ring Model

![ring_model](../img/002.png)

In addition to the "layered model", you'll also sometimes hear the Windows API being described as a "ring model", specifically referring to "rings of privilege" - see image above. The thing about this image is that it's a more general, or universal, representation of conceptual privilege layers for OS, but Windows only has **Ring 0** (kernel mode) , and **Ring 3** (user mode).

I'm including a reference to this model for posterity since you're sure to encounter it, but to be frank I don't think it has too much value in malware development, other than communicating that yes - kernel has highest privilege, while user mode has the lowest. For our purposes, the layered model provides, in my opinion, much more conceptual value.




---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "03_components.md" >}})
[|NEXT|]({{< ref "05_types.md" >}})