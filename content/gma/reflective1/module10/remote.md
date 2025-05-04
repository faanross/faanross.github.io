---
showTableOfContents: true
title: "Remote Thread Execution (WinAPI) (Theory 10.3)"
type: "page"
---
## Overview

In the previous lessons of this module, we successfully identified a target process, obtained a handle to it, allocated memory within its address space (`VirtualAllocEx`), wrote our payload (a string in Lab 10.2, but conceptually our shellcode) into that memory (`WriteProcessMemory`), and finally, adjusted the memory protections (`VirtualProtectEx`) to make it executable (or ReadOnly in the lab example, but we'd use `PAGE_EXECUTE_READ` for shellcode).

The payload now sits prepared in the target process's memory, ready to run. The final step in standard WinAPI process injection is to actually trigger its execution. The most common way to achieve this is by creating a new thread *within the target process* that starts executing directly at the beginning of our payload buffer. The WinAPI function for this is `CreateRemoteThread`.

## Triggering Execution: `CreateRemoteThread`

`CreateRemoteThread` creates a thread that runs in the virtual address space of *another* process.

```c++
HANDLE CreateRemoteThread(
  HANDLE                 hProcess,           // Handle to the target process
  LPSECURITY_ATTRIBUTES  lpThreadAttributes, // Thread security attributes (usually NULL)
  SIZE_T                 dwStackSize,        // Initial stack size (0 for default)
  LPTHREAD_START_ROUTINE lpStartAddress,     // <<< START ADDRESS of the thread function
  LPVOID                 lpParameter,        // Argument to pass to the thread function
  DWORD                  dwCreationFlags,    // Creation flags (e.g., 0 to run immediately)
  LPDWORD                lpThreadId          // Optional: Pointer to receive thread ID
);
````

**Parameters:**

- **`hProcess`**: This is the handle to the target process obtained via `OpenProcess`. It _must_ have the `PROCESS_CREATE_THREAD` access right (along with `PROCESS_QUERY_INFORMATION`, `PROCESS_VM_OPERATION`, `PROCESS_VM_WRITE` usually needed for the preceding steps).
- **`lpThreadAttributes`**: Security attributes for the new thread. Typically `NULL` for default security.
- **`dwStackSize`**: The initial size of the stack for the new thread. Passing `0` uses the default stack size for the executable.
- **`lpStartAddress`**: This is the **critical parameter** for injection. It's a pointer to the application-defined function to be executed by the thread. **Crucially, this address must be valid _within the target process's_ address space.** For our injection workflow, this will be the `remoteBufferAddress` returned by `VirtualAllocEx` (after we've written our shellcode to it and changed its protection to `PAGE_EXECUTE_READ`). The function signature expected is `DWORD WINAPI ThreadProc(LPVOID lpParameter)`. Our raw shellcode typically doesn't conform perfectly, but execution will still begin at this address.
- **`lpParameter`**: A pointer to a variable to be passed as an argument to the thread function (`lpStartAddress`). If our shellcode doesn't require an argument (which is common), we pass `NULL`.
- **`dwCreationFlags`**: Flags that control thread creation. Passing `0` means the thread runs immediately after creation. `CREATE_SUSPENDED` (0x00000004) would create the thread but not run it until `ResumeThread` is called (useful for more complex setup scenarios - we'll explore this in a future lesson).
- **`lpThreadId`**: An optional pointer to a variable that receives the thread identifier if the function succeeds. Can be `NULL`.

**Return Value:** If successful, `CreateRemoteThread` returns a `HANDLE` to the newly created thread. If it fails, it returns `NULL`.

## The Injection Workflow Completion

Using `CreateRemoteThread` completes our standard injection sequence:

1. `OpenProcess` -> `hProcess`
2. `VirtualAllocEx` (RW) -> `remoteBufferAddress`
3. `WriteProcessMemory` (payload -> `remoteBufferAddress`)
4. `VirtualProtectEx` (`remoteBufferAddress` -> RX)
5. **`CreateRemoteThread(hProcess, ..., remoteBufferAddress, ...)`** -> `hRemoteThread`

At step 5, the operating system creates a new thread context within the target process specified by `hProcess`. It sets the instruction pointer (`RIP`/`EIP`) of this new thread to `remoteBufferAddress` and (if `dwCreationFlags` is 0) schedules the thread to run. When the thread gets CPU time, it starts executing the instructions (our shellcode) located at `remoteBufferAddress`.

## Caveats and Detection Points

While `CreateRemoteThread` is the standard function for this purpose, its use is a **major red flag** for security software for several reasons.

### **Cross-Process Thread Creation**
Legitimate applications rarely need to create threads directly in _other_ unrelated processes. This action is strongly associated with code injection and malware. EDRs heavily monitor calls to `CreateRemoteThread` and its Native API equivalent (`NtCreateThreadEx`).

### **Start Address Analysis:**
As discussed in Theory 9.3, the `lpStartAddress` provided to `CreateRemoteThread` is heavily scrutinized. If it points to dynamically allocated memory (`VirtualAllocEx` region) or memory not backed by a legitimate file on disk, it's highly suspicious.

### **Parent Process:**
The process calling `CreateRemoteThread` becomes the parent or creator of the remote thread, which might look anomalous (e.g., why did `MyLoader.exe` create a thread in `notepad.exe`?).

### **Open Handles:**
The act of opening a process handle (`OpenProcess`) with powerful rights like `PROCESS_CREATE_THREAD` and `PROCESS_VM_WRITE` is itself a monitored event.


Because of these detection points, more advanced techniques (often involving the Native API `NtCreateThreadEx` or other methods like APC injection, discussed later in our course) are employed to make our thread creation stealthier. However, understanding `CreateRemoteThread` is fundamental as it represents the baseline technique.

## Conclusion

`CreateRemoteThread` is the standard WinAPI function used to initiate execution of code that has been placed into a remote process's memory. By providing the handle to the target process and the starting address of our prepared payload buffer, we can create a new thread within that process to run our code. While effective, its direct use is a significant indicator for EDRs due to the inherent suspicion surrounding cross-process thread creation and the analysis of the thread's start address. In the next lab, we will use `CreateRemoteThread` in our Go program to execute shellcode injected into a target process.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "mem_lab.md" >}})
[|NEXT|]({{< ref "remote_lab.md" >}})