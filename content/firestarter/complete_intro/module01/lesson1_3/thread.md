---
showTableOfContents: true
title: "Part 2 - Thread Architecture"
type: "page"
---

## Thread Anatomy

**Processes don't execute - threads execute.** A process is merely a container that holds resources like memory, file handles, and security context. The actual execution - the running of instructions - happens within threads, which are the fundamental units of CPU scheduling.

```
┌──────────────────────────────────────────────────────────────┐
│                         THREAD ANATOMY                       │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  KERNEL THREAD (ETHREAD/KTHREAD)                             │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  • Thread ID (TID)                                           │
│  • Thread state (Running, Waiting, etc.)                     │
│  • Priority and scheduling information                       │
│  • Kernel stack pointer                                      │
│  • Context (register values: RAX, RIP, RSP, etc.)            │
│                                                              │
│  USER THREAD COMPONENTS                                      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  • TEB (Thread Environment Block) - user-mode structure      │
│  • User stack (typically 1MB reserved)                       │
│  • TLS (Thread Local Storage) - per-thread data              │
│                                                              │
│  THREAD CONTEXT (CPU STATE)                                  │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  Saved when thread switches:                                 │
│  • RIP/EIP: Instruction pointer (where thread executes)      │
│  • RSP/ESP: Stack pointer                                    │
│  • RAX, RBX, RCX, etc.: General purpose registers            │
│  • RFLAGS: CPU flags                                         │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Kernel Thread (ETHREAD/KTHREAD)

This is the kernel's representation of a thread - the operating system's internal data structure that tracks everything the OS needs to manage and schedule the thread.

- **Thread ID (TID)**: A unique numerical identifier assigned by the OS that distinguishes this thread from all others in the system.
- **Thread state**: The current lifecycle status of the thread - whether it's actively running on a CPU, waiting for I/O, ready to run, or terminated.
- **Priority and scheduling information**: Determines when this thread gets CPU time relative to other threads; higher priority threads are scheduled more frequently by the OS.
- **Kernel stack pointer**: Points to a separate stack used when the thread executes kernel-mode code (like system calls); this is distinct from the user-mode stack.
- **Context (register values)**: The saved CPU register state that allows the OS to pause a thread and resume it later exactly where it left off.

### User Thread Components

These are the user-mode structures that exist in the process's address space and support thread execution outside the kernel.

- **TEB (Thread Environment Block)**: A per-thread data structure in user space that contains thread metadata like the thread ID, exception handling chain, and pointers to TLS. It's accessible from user mode without kernel calls.
- **User stack**: Private memory allocated for this thread's function call chain, local variables, and return addresses; each thread gets its own stack to prevent interference.
- **TLS (Thread Local Storage)**: A mechanism for storing variables that are unique to each thread, allowing global-like variables that don't conflict across multiple threads.

### Thread Context (CPU State)

When the OS switches between threads (context switching), it must save and restore the CPU's entire state so each thread can resume exactly where it stopped.

**Saved when thread switches:**

- **RIP/EIP (Instruction Pointer)**: Holds the memory address of the next instruction to execute; saving this tells the CPU where to continue when the thread resumes.
- **RSP/ESP (Stack Pointer)**: Points to the current top of the thread's stack; essential for maintaining the function call chain and local variable access.
- **RAX, RBX, RCX, etc. (General Purpose Registers)**: Store intermediate computation values, function arguments, and return values that must be preserved across context switches.
- **RFLAGS (CPU Flags)**: Contains status bits like zero flag, carry flag, and interrupt enable that reflect the result of the last operation and control CPU behaviour.


## Thread States

```
Thread Lifecycle:
                                                    
  Created → Ready → Running → Waiting → Terminated
     │       ↑  ↓      │  ↑       │
     │       └─────────┘  └───────┘
     │       (Preempted)   (Wait complete)
     │
     └──────────────────────────────────────→ Terminated
                    (Never scheduled)

States:
• Ready:      Waiting for CPU time
• Running:    Executing on CPU
• Waiting:    Blocked on I/O, synchronization, etc.
• Terminated: Finished execution
```


A thread's journey through its lifetime is managed by the operating system's scheduler, which orchestrates when and how threads gain access to the CPU. Understanding this lifecycle is fundamental to grasping how multitasking and concurrency work at the system level.

### The Journey Begins: Creation

When a new thread is created - whether at process startup or spawned by an existing thread - it enters the **Created** state. At this point, the OS has allocated the necessary data structures (kernel thread object, stack space, TEB) but hasn't yet made the thread eligible for execution. From here, the thread moves to the Ready state to join the queue of threads waiting for CPU time, or in rare cases, it might be terminated immediately if the creating process exits before the thread ever runs.

### Ready: Waiting in the Wings

In the **Ready** state, the thread is fully prepared to execute - all initialization is complete, and it's simply waiting for the scheduler to assign it to a CPU core. This is the waiting room where threads compete for processor time based on their priority levels. Threads can return to this state from Running when they're preempted (forcibly removed from the CPU to give other threads a turn) or from Waiting when whatever blocked them has completed.

### Running: The Spotlight

The **Running** state is where the action happens - the thread is actively executing instructions on a physical CPU core. Only as many threads can be in this state simultaneously as there are CPU cores available. A running thread doesn't stay running forever; it will transition out when it voluntarily waits for something (I/O, a lock, a timer), gets preempted by the scheduler to enforce fair sharing, or completes its work and terminates.

### Waiting: Blocked and Patient

When a thread enters the **Waiting** state, it has voluntarily relinquished the CPU because it needs something that isn't immediately available - perhaps it's waiting for disk I/O to complete, for another thread to release a mutex, for a network packet to arrive, or for a timer to expire. The thread remains in this blocked state, consuming no CPU cycles, until the awaited event occurs. Once the wait condition is satisfied, the OS moves the thread back to the Ready state where it competes again for CPU time.

### The Preemption Cycle

The arrow between Running and Ready (marked "Preempted") represents one of the scheduler's most important responsibilities: enforcing fairness and responsiveness. Even if a thread is happily executing and hasn't requested anything, the scheduler will periodically interrupt it - typically after a time slice of 10-30 milliseconds - and move it back to Ready, giving other threads a chance to run. This preemptive multitasking prevents any single thread from monopolizing the CPU.

### Terminated: The End

Eventually, every thread reaches the **Terminated** state, either by completing its main function, being explicitly killed, or when its parent process exits. Once terminated, the thread no longer exists as an executable entity, though some cleanup and bookkeeping may still occur before the OS fully reclaims its resources. This is a one-way transition - terminated threads cannot be resurrected.

### **States Summary:**

- **Created**: The OS has allocated the necessary data structures for the thread to exist, but hasn't yet made the thread eligible for execution.
- **Ready**: The thread is runnable and waiting in the scheduler's queue for its turn on a CPU core; it has everything it needs except processor time.
- **Running**: The thread is currently executing instructions on a physical CPU core; this is the only state where actual work happens.
- **Waiting**: The thread is blocked, unable to proceed until some external event occurs (I/O completion, lock acquisition, signal arrival); it consumes no CPU resources in this state.
- **Terminated**: The thread has finished execution and its lifecycle is complete; the OS will reclaim its stack and control structures.




## Offensive Thread Manipulation

**NOTE**: In this case I won't provide functioning code, just an outline with comments. Providing the complete logic would require multiple leaps over knowledge gaps, and I'm afraid that act will cause more confusion than anything else at this point. But don't fret - we will return to this application in Module 5 and develop it in all its glory. For now, just read the code (you won't be able to compile), and get a sense for the overall structure and logical flow.

```go
// Thread creation for injection
package main

import (
    "syscall"
    "unsafe"
)

var (
    kernel32              = syscall.NewLazyDLL("kernel32.dll")
    procCreateRemoteThread = kernel32.NewProc("CreateRemoteThread")
)

// Create thread in remote process
func CreateRemoteThread(
    hProcess syscall.Handle,
    lpStartAddress uintptr,
    lpParameter uintptr,
) (syscall.Handle, error) {
    
    handle, _, err := procCreateRemoteThread.Call(
        uintptr(hProcess),
        0, // Default security
        0, // Default stack size
        lpStartAddress,
        lpParameter,
        0, // Run immediately
        0, // Don't need thread ID
    )
    
    if handle == 0 {
        return 0, err
    }
    return syscall.Handle(handle), nil
}

// Example usage (simplified injection)
func InjectDLL(processHandle syscall.Handle, dllPath string) error {
    // 1. Allocate memory in target process
    // 2. Write DLL path to allocated memory
    // 3. Get LoadLibraryA address
    // 4. Create remote thread at LoadLibraryA with DLL path as parameter
    
    // This is the foundation of DLL injection (Module 5)
    return nil
}
```



## **The TEB (Thread Environment Block)**

Similar to what we see in the previous section on Processes, the `ETHREAD/KTHREAD` contains important information related to the thread, but since it resides in kernel memory we cannot access it directly from userland. We can however again access another user-mode structure, the **TEB**, which makes this information available to us.

User-mode structure accessible to each thread:

```c
// Simplified TEB structure (ntdll!_TEB)
typedef struct _TEB {
    NT_TIB          NtTib;                  // Thread Information Block
    PVOID           EnvironmentPointer;     // Environment variables
    CLIENT_ID       ClientId;               // Process ID + Thread ID
    PVOID           ActiveRpcHandle;        
    PVOID           ThreadLocalStoragePointer; // TLS array
    PPEB            ProcessEnvironmentBlock;   // Pointer to PEB
    ULONG           LastErrorValue;         // GetLastError() value
    // ... many more fields
} TEB, *PTEB;
```





---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./process.md" >}})
[|NEXT|]({{< ref "../../moc.md" >}})