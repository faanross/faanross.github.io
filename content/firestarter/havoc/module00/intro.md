---
showTableOfContents: true
title: "Course Introduction"
type: "page"
---


## **Course Focus: Reimplementing Demon Techniques in Go**

This course will pivot from a broad architectural overview to a deep technical dive into the _agent-side_ tradecraft implemented within the Havoc Demon. While the Demon itself is written in C/C++, our objective is to understand the underlying _techniques_ and explore how to achieve similar results using the Go programming language. We will dissect specific features such as:

- Advanced sleep obfuscation methods (Ekko, Zilean, Foliage).
- Call stack and return address spoofing.
- Indirect syscall execution.
- AMSI/ETW patching using hardware breakpoints.
- Process injection and migration strategies.
- Token manipulation and the concept of a token vault.
- Executing .NET assemblies and PowerShell from an unmanaged host.
- Loading and executing Beacon Object Files (BOFs).
- Interacting with Active Directory components.

Through theoretical lessons and practical Go-based labs, you will gain the knowledge and skills to implement these sophisticated offensive techniques, enhancing your capabilities in tool development and red team operations.



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "../module01/intro.md" >}})