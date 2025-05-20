---
showTableOfContents: true
title: "Key Components of the Windows Architecture"
type: "page"
---
## User vs Kernel Mode Components
The Windows operating system is composed of numerous interacting components, broadly divided into those that operate in User Mode and those in Kernel Mode.

![win_architecture](../img/001.png)


### User Mode Components

#### Processes (General Concept)
Central to understanding User Mode is the concept of the **process**.

A process is essentially an instance of a running program. When a user launches an application, such as Notepad, the Windows operating system creates a new process for that application (e.g., `notepad.exe`). The creation and initialization of this process involve several steps and the allocation of resources:

1. **Loading the Executable Image**: The operating system reads the application's executable file (e.g., `notepad.exe`) from disk and maps its contents—the program's machine code instructions and initial static data—into memory. This loaded representation is known as the executable image of the process.

2. **Allocating Core Process Structures**: Along with the image, the OS assigns several critical structures to the new process:
    - **Private Virtual Address Space**: Each process is provided with its own unique and isolated range of memory addresses. This private virtual address space ensures that one process cannot directly view or modify the memory contents of another, a crucial feature for system stability and security. The process code and data operate within this space.
    - **Threads**: A process must contain at least one thread of execution to perform any work. A thread is a sequence of instructions that the processor executes. When a process is created, an initial thread is typically started, which begins executing the program at its designated entry point. A process can create additional threads to perform tasks concurrently.
    - **Private Handle Table**: Processes interact with system resources (such as files, registry keys, synchronization objects, or even other processes and threads) through identifiers called handles. Each process maintains a private handle table, which is an internal list mapping these handles to the actual kernel objects they represent. This table is accessible only from within the process.
    - **Primary Access Token**: Every process has an access token associated with it. This token is a security object that defines the security context of the process, including the identity of the user account that launched it, the security groups to which the user belongs, and the specific privileges held by the user. The operating system uses the information in the access token to make authorization decisions whenever the process attempts to access securable objects or perform privileged operations.

If you'd like to understand processes in more depth see [this video](https://www.youtube.com/watch?v=LAnWQFQmgvI) I made on the topic.


#### User Processes

User processes represent instances of executing applications, such as `notepad.exe` or `explorer.exe`. Each user process runs in a private, isolated virtual address space. This isolation is a fundamental security boundary in User Mode, preventing one application from directly accessing or interfering with the memory of another, thereby enhancing system stability and security.


#### Subsystem DLLs
Subsystem Dynamic Link Libraries (DLLs) implement the Application Programming Interfaces (APIs) for various operating system environments, historically known as subsystems. Early versions of Windows NT were architected to support multiple subsystems—notably for OS/2 and POSIX applications—at a time when broader application compatibility was a strategic goal (MS had not yet cornered the market, so wanted to be more "accommodating").

Over time, as the Windows environment itself became predominant, these other subsystems were largely deprecated or removed. The Windows Subsystem is now the essential and primary environment. The architectural concept of distinct subsystems persists mainly for backward compatibility.


#### win32 API
This refers to the remaining subsystem DLL, i.e. the Windows Subsystem. But the term is an abstraction for a collection of different DLLs, including `kernel32.dll`, `user32.dll`, `gdi32.dll`, `advapi32.dll`, and `combase.dll`. These collectively constitute the documented Windows API (win32 API), the primary interface through which applications interact with the OS. From a security perspective, these DLLs are common points for API monitoring and hooking by security software and, conversely, by malware seeking to intercept or modify program behaviour.

#### NTDLL.DLL
Positioned at the lowest level of User Mode code, `ntdll.dll` implements the Windows native API and serves as the crucial gateway for transitions into Kernel Mode via system calls. Beyond managing this transition, NTDLL is responsible for core User Mode functionalities such as the Heap Manager, the Image Loader (for loading executables and DLLs), and aspects of the User Mode thread pool. While a significant portion of the native API exposed by NTDLL remains undocumented, understanding parts of it is vital for advanced system programming, reverse engineering, and malware development.


#### System Processes

System processes are specialized executables that perform essential operating system functions in the background, typically without direct user interaction. Many of these are critical for system stability; terminating them can lead to a system crash. System processes often run with high privileges within User Mode (e.g., under the `SYSTEM` account) and utilize the native API. Notable examples include the Session Manager Subsystem (`Smss.exe`), Local Security Authority Subsystem Service (`Lsass.exe` – a frequent target for credential theft attacks), `Winlogon.exe`, and the Service Control Manager (`Services.exe`). Their privileged nature, plus the fact that they automatically respawn following a system reboot, makes them prime targets when seeking to escalate privileges or persist on a system.


#### Subsystem Process (Csrss.exe)

The Client Server Runtime Subsystem (`Csrss.exe`) functions as an essential support process for the Windows environment. It manages console windows, process and thread creation/deletion tracking (in conjunction with the kernel), and other critical system functions. `Csrss.exe` is a critical process; its termination will result in a system crash. Typically, an instance of `Csrss.exe` runs for each active session, including session 0 (for system services) and user-specific interactive sessions.


### Kernel Mode Components

#### The Executive

The Executive forms the upper layer of `NtOskrnl.exe` (the core OS image often referred to simply as "the kernel") and contains the bulk of Kernel Mode operating system code. It is composed of several "managers," each responsible for a distinct aspect of the system. These include the Object Manager (manages resources like files and registry keys, and enforces access controls), Memory Manager, Process Support, I/O Manager (manages device communication), Plug & Play Manager, Power Manager, Configuration Manager (interfaces with the registry), and the Security Reference Monitor (SRM, which enforces security policies and performs access checks). The Executive provides the APIs and services used by other Kernel Mode components, including device drivers.



#### The Kernel (Layer)

Beneath the Executive within `NtOskrnl.exe` lies the Kernel layer. This layer is responsible for the most fundamental, hardware-proximate, and often time-critical functions of the operating system. Its duties include low-level thread scheduling and dispatching, interrupt and exception handling, and managing kernel synchronization primitives (e.g., mutexes, semaphores). To optimize performance for these critical tasks, portions of the Kernel layer may be implemented in assembly language specific to the CPU architecture. The integrity of this layer is paramount for overall system stability and security.

#### Device Drivers

Device drivers are loadable Kernel Mode modules (.sys files) that enable the operating system to interact with hardware devices. They operate with full Kernel Mode privileges, allowing direct hardware manipulation and access to kernel services. While many drivers facilitate hardware communication, others, known as software drivers or filter drivers, can provide system services, extend OS functionality, or intercept I/O requests for security monitoring (e.g., antivirus file system filter drivers) or other purposes. Due to their privileged execution and direct hardware access, third-party device drivers represent a significant attack surface and are a common source of system instability or security vulnerabilities if not carefully developed and vetted.

#### Win32k.sys

Win32k.sys is the Kernel Mode component of the Windows graphical subsystem. It manages windowing, user interface controls, and the Graphics Device Interface (GDI). Historically, due to its complexity and the breadth of functionality exposed to User Mode, win32k.sys has been a notable source of security vulnerabilities, particularly those leading to privilege escalation.

#### Hardware Abstraction Layer (HAL)

The Hardware Abstraction Layer (hal.dll) is a critical Kernel Mode component that isolates the Windows Executive, the Kernel layer, and device drivers from variations in underlying hardware platforms (motherboards, interrupt controllers, multiprocessor configurations). It provides a consistent interface for accessing hardware resources, simplifying driver development and enhancing OS portability across different hardware. The HAL's correct functioning is essential for system stability and fundamental operations.

#### Hyper-V Hypervisor and Virtualization Based Security (VBS)

On systems supporting hardware virtualization extensions, the Hyper-V hypervisor can underpin the operating system. When Virtualization Based Security (VBS) is enabled (available in certain Windows versions like Windows 10/11 Enterprise and Windows Server 2016 and later), the hypervisor creates a more secure operating environment. VBS uses the hypervisor to establish isolated regions of memory and execution known as Virtual Trust Levels (VTLs). The main operating system kernel runs in VTL0, while more sensitive security operations and data (like credential hashes protected by Credential Guard, or code integrity enforcement via Hypervisor-Enforced Code Integrity - HVCI) can be isolated in a secure kernel running in VTL1. This architecture provides protection for critical security assets even if the main OS kernel (VTL0) is compromised.




---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "02_process.md" >}})
[|NEXT|]({{< ref "04_api.md" >}})