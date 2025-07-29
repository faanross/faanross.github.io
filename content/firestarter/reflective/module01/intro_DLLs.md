---
showTableOfContents: true
title: "Introduction to DLLs (Theory 1.1)"
type: "page"
---

Before diving into the development of an actual C2 implant/loader, we must first build a solid understanding of two core components involved: Dynamic Link Libraries (DLLs) and shellcode.

## What is a Dynamic Link Library (DLL)?

Imagine you are building several different software applications, and each one needs the ability to perform a common set of tasks, such as reading configuration files, encrypting data, or drawing graphical elements on the screen. One approach would be to copy and paste the code for these tasks into each application. However, this is inefficient: it increases the size of each application, makes updates difficult (you'd have to update the code in every application separately), and wastes memory if multiple applications running the same duplicated code are active simultaneously.

A much more elegant solution is provided by **Dynamic Link Libraries (DLLs)**. A DLL is essentially a library containing code and data that can be used by multiple programs _at the same time_. Instead of linking the common code directly into each application when it's compiled (static linking), the applications link to the DLL _dynamically_ at runtime.

When an application needs to use a function contained within a DLL:

1. The operating system checks if the required DLL is already loaded into memory by another process.
2. If it is, the system maps the existing DLL's code section into the calling application's virtual address space, allowing the application to use the code without loading another copy. This saves memory.
3. If the DLL is not already loaded, the operating system loads it into memory and then maps it into the application's address space.
4. The application can then call the functions within the DLL as if they were part of its own code.

## Why Use DLLs?

The use of DLLs provides several significant advantages:

1. **Modularity:** Software can be broken down into smaller, manageable components. Different teams can work on different DLLs independently.
2. **Code Reusability:** As discussed, common functions can be placed in a DLL and shared by many applications, avoiding code duplication.
3. **Memory Savings:** The operating system can share a single copy of a DLL's code in physical memory among multiple applications that use it.
4. **Easier Updates:** If a bug is fixed or a feature is added in a DLL, only the DLL needs to be updated. All applications using that DLL will automatically benefit from the change the next time they run, without needing to be recompiled themselves (assuming the DLL's interface remains compatible).
5. **Platform Extensibility:** Windows itself uses DLLs extensively. Core operating system functionality (like file operations, window management, networking) resides in system DLLs (e.g., `kernel32.dll`, `user32.dll`, `ws2_32.dll`). This allows the OS to be updated and extended more easily.

## Exporting Functions from DLLs

For an application to use a function residing within a DLL, the DLL must make that function available. This process is called **exporting**. When we compile a DLL, we have to explicitly specify which functions should be made available to external programs. These exported functions then form the DLL's public interface.

Any process that wants to use a DLL has to know 2 key pieces of information:
1. The **name** of the DLL file.
2. The **name** (or sometimes an **ordinal number**, which is like a numeric index) of the specific function it wants to call within that DLL.

Think of it like a library building (the DLL) containing many books (functions). To borrow a specific book, you need to know the library's name and the title of the book you want.

## Conclusion
Though this is really an extremely brief introduction to DLLs, it's enough to get us going for the purpose of this course. If however you wanted to know more, please see Deep Dive on DLLs (TODO: link).

## References
[CppCon 2017: James McNellis “Everything You Ever Wanted to Know about DLLs”](https://www.youtube.com/watch?v=JPQWQfDhICA)
___

[|TOC|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "intro_shellcode.md" >}})