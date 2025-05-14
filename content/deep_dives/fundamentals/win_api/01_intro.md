---
showTableOfContents: true
title: "Introduction to the Windows Architecture"
type: "page"
---
## The Layered Game: Why Windows Isn't Just One Big Target
Forget thinking of Windows as a single, massive blob of code. Underneath the pretty icons and windows, it's a meticulously organized, multi-layered cake. And for us, knowing how those layers are stacked is key to knowing how to exploit it.

This **layered architecture** is our first key concept:

- Each layer has its own job, its own set of rules.
- A layer talks to the one directly below it to get things done and provides services to the one directly above it.
- This creates a hierarchy, a chain of command, from the deepest parts of the OS right up to the applications you see.

But why should we care about these layers? Because this structure dictates everything from a security and exploitation standpoint:

- **Interaction Points:** It shows us how different parts of the OS connect. Where does user input go? How does a command to save a file travel through the system? These are all potential points for interception or manipulation.
- **Information Flow:** If we want to steal data, we need to know how it moves between these layers.
- **Critical Functions:** Understanding which layer handles process creation, network communication, or file system access tells us where to focus our efforts to control or subvert these functions.
- **Evasion Strategies:** Security software also operates at different layers (user-mode AV scanners, kernel-mode firewalls, hypervisor-level introspection). Knowing the layers helps us understand where these defenses are and how we might operate above, below, or between them.

Think of it this way: approaching Windows as a layered system turns an overwhelming complex beast into a series of interconnected, more manageable zones. We're building our mental map of the Windows internals.

## The Great Divide: User Mode vs. Kernel Mode

Beyond just general layers, there’s one **fundamental, hardware-enforced split** in Windows that you absolutely _must_ internalize: the **User Mode vs. Kernel Mode divide**. This isn't just some contrived suggestion by Microsoft; it's a hard boundary policed by the CPU itself. Every single piece of code running on that Windows box – from Notepad to the most sophisticated payload – lives in one of these two realms.

## **Kernel Mode: The Inner Sanctum

Kernel Mode is the operating system's core, its command center. Code running here has the keys to the kingdom:

- **Unrestricted Hardware Access:** Talk directly to any piece of hardware.
- **Total Memory Control:** Read or write any part of physical memory.
- **Highest Privilege (Ring 0):** On x86/x64 CPUs, this is Ring 0 – the most privileged execution level. Nothing can tell Kernel Mode code "no."

This is where the most fundamental OS components reside, including:

- The **Windows Executive:** The managers of the OS – handling memory, processes, threads, I/O, security, and more.
- The **Kernel (ntoskrnl.exe and friends):** The real low-level stuff – deciding which thread runs when, handling hardware interrupts.
- The **Hardware Abstraction Layer (HAL):** The part that smooths over differences between various motherboards and chipsets so the Kernel doesn't have to care.
- **Device Drivers:** Those critical pieces of software that let the OS talk to your graphics card, network adapter, disk drives, etc. This is a _prime_ target area for malware that wants persistence and power.

Operating in Kernel Mode means you _are_ the OS, for all intents and purposes. But with great power comes great instability if you screw up. A mistake, an unhandled exception, a bad pointer in Kernel Mode code doesn't just crash an app; it usually takes the entire system down with it – cue the infamous **Blue Screen of Death (BSOD)**. A BSOD is noisy and tends to make users suspicious, which is generally bad if you're trying to be stealthy.

## **User Mode: The Application Playground**

This is where all our everyday applications run: the web browser, Microsoft Word, games, and typically, the initial execution point of our malware.

Compared to Kernel Mode, User Mode is a walled garden:

- **Restricted Access:** No direct hardware access. Can't just peek into another process's memory or mess with the OS's critical data structures.
- **Private Address Spaces:** Each User Mode process thinks it has the entire address space to itself (an illusion maintained by the Kernel Mode memory manager).

If a User Mode application wants to do anything "interesting" – like read a file, send a packet over the network, create a new process, or even just get the current time – it can't do it directly. It has to first get permission from the Kernel.

## **Why This Divide?**

Microsoft didn't create this split just to make our lives harder, rather, it serves critical purposes, which, from our perspective, are both obstacles to overcome and features to understand for exploitation:

1. **System Stability (Obstacle/Feature):** If your user-mode payload crashes, it usually just takes itself out. The OS and other apps keep chugging along. This is good for the user, and frankly, good for a long-running, stealthy implant – we don't want your malware BSODing the box every five minutes.
2. **Security and Protection (The Wall We Climb):** This boundary is _the_ primary defense against malware. It stops one dodgy application from trashing the OS or sniffing data from your online banking session in another process. Our job is often to find ways _around_ or _through_ these protections. Understanding the rules of User Mode helps us understand how to bend or break them, or how to leverage Kernel Mode vulnerabilities to bypass them entirely.
3. **Controlled Hardware Access (The Gatekeepers):** Want to talk to the webcam? Control the keyboard? We can't just send commands to the hardware from User Mode. We have to go through Kernel Mode drivers. This means driver exploits, or loading our _own_ (malicious) driver, become attractive pathways if we need that level of control.

## **Crossing the Chasm: System Calls – The Gateway to Kernel Power**

So, if User Mode apps can't do much on their own, how does anything get done? They make system calls.

When User Mode code calls a standard Windows API function – say, `CreateFileW` to open a file – it's not `CreateFileW` itself that does the deep disk magic. Deep inside that `CreateFileW` function (often in a lower-level DLL like `ntdll.dll`), a special instruction (`syscall` on x64, or older mechanisms like `sysenter` or int 0x2E on 32-bit) is executed.

This instruction is a formal request to the Kernel:

1. The CPU switches from User Mode to Kernel Mode.
2. Control is transferred to a specific Kernel routine (the system call dispatcher).
3. The Kernel validates the request (Are we allowed to do this? Are the parameters sane?).
4. If all checks out, the Kernel performs the operation on our behalf (e.g., accesses the file system driver to open the file).
5. The Kernel prepares the result.
6. The CPU switches back to User Mode, and the User Mode code gets the result.

# **Why System Calls are Critical for Malware Devs**

Why should we care about syscalls as malware developers? Because it's this transition point, this meticulously controlled gateway between User Mode and Kernel Mode, where the action truly happens. It's the OS's primary chokepoint, and for us, that means it's a goldmine for intelligence, interception, and manipulation. Every significant request a User Mode application makes to the operating system core – whether it's opening a file, allocating memory, or sending a network packet – _must_ pass through this narrow channel. This makes the system call interface one of our primary targets.

Consider API hooking, a cornerstone of many malware techniques. When we're aiming to intercept or alter what another program is doing, we're often targeting functions in User Mode DLLs like `kernel32.dll` or, for a stealthier approach, the lower-level `ntdll.dll`. Our hooks catch the call _before_ it makes that leap into the kernel via a syscall. On the other side of the divide, for those aiming for deeper control, kernel-mode hooking techniques (like the now heavily guarded SSDT hooking) are designed to intercept these requests as they arrive _inside_ the kernel, right at the handlers for those system calls.

Beyond just interception, understanding the flow of system calls is like having an X-ray into a program's soul. Even if an application's high-level code is a tangled mess of obfuscation, the underlying system calls it makes tell a clearer story of its _actual_ intentions. Is it trying to write to unexpected files? Enumerate running processes? Connect to suspicious IP addresses? The sequence and parameters of its system calls betray its true purpose.

And, of course, we'll be using this gateway constantly, even if indirectly. Every time our payload uses a standard Windows API function to allocate memory for shellcode, create a new thread for C2 communication, write stolen data to a hidden file, or reach out across the network, we are implicitly relying on system calls to get the kernel to perform these privileged operations on our behalf. We're knocking on that same front door, making requests. The more we understand about how that door works, who answers, and what questions they ask, the better we'll be at getting what we want while not being noticed.



---
[|TOC|]({{< ref "moc.md" >}})
[|NEXT|]({{< ref "02_components.md" >}})