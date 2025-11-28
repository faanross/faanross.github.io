---
showTableOfContents: true
title: "Part 7 - Knowledge Validation"
type: "page"
---


## **KNOWLEDGE VALIDATION**

**Question 1: Memory Protections**

You allocate memory with `VirtualAlloc(PAGE_READWRITE)`, write shellcode, then call `CreateRemoteThread()`pointing to it. What happens?

A) Shellcode executes successfully  
B) Access violation - memory not executable  
C) Windows Defender blocks it  
D) Process crashes with stack overflow

**Question 2: Process Architecture**

Which is TRUE about Windows processes?

A) Processes execute code; threads are just containers  
B) Each process has isolated virtual address space  
C) All processes share the same kernel-mode memory  
D) Both B and C

**Question 3: PE Format**

You're doing process hollowing. After writing your PE to the target process, what MUST you do if it loads at a different base address than preferred?

A) Nothing - Windows handles it automatically  
B) Apply relocations manually  
C) Restart and try again  
D) Use only position-independent code

**Question 4: PEB/TEB**

What PEB field is commonly used for anti-debugging?

A) ImageBaseAddress  
B) BeingDebugged  
C) ProcessHeap  
D) Ldr

**Question 5: Privileges**

You want to inject into a SYSTEM process. Which privilege is ESSENTIAL?

A) SeImpersonatePrivilege  
B) SeBackupPrivilege  
C) SeDebugPrivilege  
D) SeTakeOwnershipPrivilege

---

**ANSWERS:**

**Q1: B** - Access violation. Memory allocated as RW is not executable. You must call `VirtualProtectEx(PAGE_EXECUTE_READ)` before executing. DEP/NX will crash the thread if you try to execute non-executable memory.

**Q2: D** - Both B and C are true. Each process has isolated _user-mode_ address space (0x0-0x7FF...), but all processes share the _same_ kernel-mode address space (0x800...-0xFFF...). This is why kernel exploits are so powerful.

**Q3: B** - Apply relocations manually. If the PE can't load at its preferred ImageBase, all absolute addresses in the code need to be adjusted by the delta. Process hollowing requires you to be the PE loader, including relocation handling.

**Q4: B** - PEB.BeingDebugged (offset +0x02) is set to 1 when a debugger is attached. Simple anti-debug check, though easily defeated. More advanced checks use PEB.NtGlobalFlag, heap flags, etc.

**Q5: C** - SeDebugPrivilege. This privilege allows opening ANY process with PROCESS_ALL_ACCESS, including SYSTEM processes. Without it, you're limited by normal access checks. Administrators have it but must enable it.

---









---

[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./peb_teb.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})