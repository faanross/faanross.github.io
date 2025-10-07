---
showTableOfContents: true
title: "Part 6 - Knowledge Validation"
type: "page"
---


## **KNOWLEDGE VALIDATION**

Test your understanding:

**Question 1: Binary Size**

Your implant is 8MB. The client complains it's too large. What's the BEST combination of techniques to reduce size while maintaining functionality?

A) UPX compression only  
B) Strip symbols + UPX + remove unused imports  
C) Strip symbols + trimpath + avoid large packages  
D) Rewrite in C

**Question 2: Cross-Compilation**

You're on macOS and need to build for Windows with CGO (C integration). What's required?

A) Just `GOOS=windows go build`  
B) Install MinGW cross-compiler and set CC variable  
C) Impossible, need Windows machine  
D) Use Docker with Windows container

**Question 3: Runtime Behavior**

An EDR detects your Go implant by monitoring periodic memory patterns. What's the MOST likely cause?

A) String literals in binary  
B) Garbage collection behavior  
C) Import table  
D) File size

**Question 4: Language Selection**

You need a 50KB implant with direct syscalls and manual memory control. Which language?

A) Go  
B) Python  
C) C/C++  
D) C#

**Question 5: Evasion**

Which Go runtime feature is MOST helpful for evasion?

A) Garbage collection  
B) Type reflection  
C) Static compilation (single binary)  
D) Large standard library

---

**ANSWERS:**

**Q1: C** - Strip symbols + trimpath + avoid large packages is best. UPX can trigger AV. Rewriting in C is overkill and time-consuming. Proper Go optimization can get you to 2-3MB which is acceptable.

**Q2: B** - CGO cross-compilation requires the target platform's C compiler. Install MinGW-w64, set `CC=x86_64-w64-mingw32-gcc`, enable `CGO_ENABLED=1`, then build.

**Q3: B** - Periodic GC pauses create timing patterns EDR can detect. Mitigate by disabling GC during sensitive operations or using `debug.SetGCPercent()`.

**Q4: C** - For 50KB size with direct syscalls and manual memory control, C/C++ is the only realistic choice. Go can't achieve 50KB, Python isn't compiled, C# requires runtime.

**Q5: C** - Static compilation (single binary with no dependencies) is most helpful. Makes deployment easier, reduces forensic artifacts. Other features are neutral or negative for evasion.

---





---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./setup.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})