---
showTableOfContents: true
title: "Basic Thread Obfuscation Concepts (Theory 9.4)"
type: "page"
---
## Overview

In our previous lessons, we improved our shellcode execution primitive within our `ExecuteShellcode` function. We decoupled our RWX memory permissions and added decoy and delay functions in Lab 9.1, and in Lab 9.2 we added runtime decryption to hide the static shellcode signature within the DLL. Now, at least our shellcode only exists in its raw, executable form transiently in memory with the correct (RX) permissions.

## One Major Remaining Blemish
However, even with these improvements, the way we *initiate* the execution can still attract unwanted attention.

Currently, our `ExecuteShellcode` function culminates in this line:

```c++
    void (*shellcode_func)() = (void(*)())exec_memory;
    shellcode_func();
````

Here we are directly casting the memory address (`exec_memory`) where our shellcode resides to a function pointer, and then we immediately call it. This direct execution start point is a significant indicator, especially once we start shifting towards remote process injection.


## Decoupling Execution with `CreateThread`

One common alternative to decouple the shellcode's execution from the main program flow is to launch it in a new thread using the `CreateThread` API.

So our relative simple code above becomes something like:

```cpp
    // New Declarations inside of ExecuteShellcode()
    HANDLE hThread = NULL; // Handle for the new thread
    DWORD threadId = 0;   // Variable to receive the thread ID

	// Other code is the same as before, until we get to void (*shellcode_func)() = (void(*)())exec_memory;
	// This is replaced with 3 major steps

	// --- STEP 1: Execute via CreateThread ---
    hThread = CreateThread(
                    NULL,                   // Default security attributes
                    0,                      // Default stack size
                    (LPTHREAD_START_ROUTINE)exec_memory, // Thread start address = shellcode buffer!
                    NULL,                   // No parameter to pass
                    0,                      // Run immediately
                    &threadId);             // Receive thread ID

    if (hThread == NULL) {
        // Attempt cleanup even on thread creation failure
        DWORD dummyProtect;
        VirtualProtect(exec_memory, sizeof(calc_shellcode), oldProtect, &dummyProtect);
        VirtualFree(exec_memory, 0, MEM_RELEASE);
        return FALSE;
    }

    // --- Change 2: Optionally Wait for Thread ---
    // For this example (launching calc), we wait so we see it pop before cleaning up.
    // In a real implant, you might skip waiting and let it run in the background.
    WaitForSingleObject(hThread, INFINITE); // Wait indefinitely for the thread to terminate

    // --- Execution presumed complete ---
    success = TRUE;

    // Change 3: Close the handle (important!)
    CloseHandle(hThread); 

```


## Pros and Cons of Using `CreateThread`

The issue here is that, while there might be some benefits in using `CreateThread`, it's potentially offset by some disadvantages.

In terms of benefits, the introduction of some basic concurrency may make the preparing function's behaviour look more normal. Our main function calls `CreateThread` and potentially returns _immediately_, while the shellcode runs independently in the background thread.

It is also creates a degree of logical separation - the shellcode execution is now contained within a separate thread context, which might slightly complicate analysis for tools focusing only on the original thread's call stack _after_ the `CreateThread` call returns.


But... There's are also some serious risks involved in using `CreateThread`. It is among the most heavily monitored API calls - and EDR will undoubtedly scrutinize new thread creation events.

But perhaps an even bigger red flag here is the `lpStartAddress`. EDRs expect threads to start execution at addresses within known, loaded modules (DLLs or the main EXE). Seeing a thread start execution directly within a region of dynamically allocated memory (`VirtualAlloc`-ed pages) that isn't mapped to a file on disk is **highly suspicious** and a strong heuristic for process injection or in-memory shellcode execution. In other words... We're stuck with the exact same problem as before.

This being the case, I wanted you to be aware of `CreateThread`, but since it's moot whether changing our current code to use it instead will have any improvement at all, we won't implement it in a lab.

It is clear however that the real problem here is related to the fact that we immediately jump to executing code that is now within a known, loaded module. So for the remainder of this lesson let's try to better understand exactly why this is the case, and then explore some techniques that may indeed have practical benefit in overcoming this hurdle.



## Thread Monitoring & Start Address Analysis

Modern EDR solutions don't just look at memory permissions or static file signatures; they heavily monitor system behaviour, including thread creation and execution patterns. When a new thread is created (either in the current process via `CreateThread` or in a remote process via `CreateRemoteThread`), EDRs often inspect several key characteristics:

### Thread Start Address
This is the memory address where the new thread begins execution. EDRs maintain knowledge about legitimate module memory ranges (like `kernel32.dll`, `ntdll.dll`, the main executable itself). A thread starting execution at an address that _doesn't_ map back to a known, loaded module on disk (i.e., it starts in dynamically allocated or "anonymous" memory, or potentially a modified section of a legitimate module) is immediately suspicious. Our current approach, where execution starts directly at `exec_memory` (which was allocated by `VirtualAlloc`), falls squarely into this suspicious category.

### Call Stack Analysis
EDRs can inspect the call stack of a thread. The call stack shows the sequence of function calls that led to the current execution point. A thread starting directly in anonymous memory will have a very shallow or unusual call stack compared to threads started via legitimate OS mechanisms or application entry points, which usually have deeper stacks reflecting the normal program flow (e.g., `main` -> `SomeFunction` -> `CreateThread`).

### Memory Region Characteristics
The memory region pointed to by the start address is also examined. Does it reside in recently allocated memory? Does it have unusual permissions (like RWX)? Does it contain code patterns associated with shellcode?


Simply put, creating a thread that immediately starts executing in a freshly allocated (or even RW->RX protected) memory region containing shellcode is a classic indicator that EDRs are specifically trained to detect.



## Hiding the Start Address

One basic approach to mitigate start address detection is to avoid starting the thread _directly_ at the beginning of our shellcode buffer. Instead, we might try to make the initial execution point look more legitimate. Some conceptual techniques include:

### ROP (Return-Oriented Programming) Chains
Start the thread at a ROP gadget (a small sequence of legitimate instructions ending in `ret`, found in loaded DLLs like `ntdll.dll`) that eventually pivots the execution flow to your shellcode. The start address is now within `ntdll.dll`, which looks less suspicious initially.

### Calling a "Benign" API First
Start the thread by calling a legitimate Windows API function. This function might be chosen such that, through careful argument manipulation or hooking its return value (more advanced techniques), it eventually redirects execution to your shellcode. The initial start address and call stack look more normal.

### Thread Start Address Spoofing
More advanced techniques attempt to manipulate OS structures or use specific API flags to make it _appear_ as if the thread started at a legitimate address, even though it quickly jumps to the malicious code.

Whatever exact technique is used, the core idea is essentially to make the initial instruction pointer (`RIP`/`EIP`) for the new thread point somewhere seemingly innocuous (like inside a system DLL) rather than directly into your dynamically allocated payload buffer. So you can think of it was integrating a detour - we're still planning on landing on our shellcode to execute it, but we just don't jump into it directly.


## Related Concept: Sleep Masking

Another area related to thread behaviour and evasion is **sleep masking**. C2 implants often need to "sleep" for periods to avoid constant activity that could be detected. However, a thread containing decrypted malicious code or configuration in memory that sleeps is more vulnerable to memory scanning.

With sleep masking techniques we aim to obfuscate the agent's state _during_ these sleep periods.

Some common approaches involve:
- Encrypting sensitive memory regions (like the agent's code or configuration) before sleeping.
- Using alternative sleep mechanisms (like waiting on timers or synchronization objects (`WaitForSingleObject`) instead of the simple `Sleep()` API, which is easily hooked).
- Restoring memory permissions and decrypting data only upon waking up, often using techniques like asynchronous procedure calls (APCs) or timer callbacks to trigger the wake-up logic.

We will delve much deeper into sleep masking techniques later in the curriculum, but it's useful to understand it as another facet of obfuscating not just _where_ code executes, but _when_ and _how_ it pauses and resumes.

## Conclusion
While we've improved our memory allocation pattern and hidden the static shellcode, the _initiation_ of execution via a 
direct call to our allocated buffer remains a potential detection point, especially when considering remote threads. EDRs closely monitor thread start addresses and call stacks.

Basic thread obfuscation concepts revolve around making this starting point appear more legitimate, perhaps by pointing it initially to existing code (like API functions or ROP gadgets) that subsequently redirects to our payload. Sleep masking decreases the probability of our malicious code being scanned once it's injected into memory during inevitable C2 sleep periods.

Since these techniques involve a significant jump in complexity, as well as a reliance on foundational knowledge 
we've not yet explored, we're not really equipped to do a real practical implementation of them. At this point it will create more questions than answers.

That being the case there won't be any lab to accompany this theoretical section, at least now. We'll cover the
discussed techniques in future modules once we're equipped to do so, for now I just wanted point this issue out
in an effort to illuminate the territory we're in, and the dilemmas we're facing.

This being the case, we've in some sense reached the limit of what this specific paradigm has to offer, and so we're
now ready to enter the exciting domain of external process injection. 


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "encrypt_lab.md" >}})
[|NEXT|]({{< ref "../module10/process.md" >}})