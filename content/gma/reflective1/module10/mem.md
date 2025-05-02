---
showTableOfContents: true
title: "Remote Memory Operations (WinAPI) (Theory 10.2)"
type: "page"
---
## Overview

In the previous lesson and accompanying lab, we learned how to enumerate running processes and obtain a handle (`hProcess`) to a target process with specific access rights. This handle is our key to interacting with the target process's virtual address space, i.e. allowing us to inject our shellcode into its memory.

Now that we have a handle, the next steps in performing process injection involve manipulating the target process's memory. Specifically, we need to:

1.  **Allocate** a region of memory within the target process to hold our  shellcode.
2.  **Inject** our shellcode into that allocated remote memory region.
3.  **(Often) Change** the memory protection flags of that region to make it executable just before we trigger it.

## The Need for Remote Operations

This process should seem very familiar of course since it's the same conceptual steps we perform when injecting into our own memory. The major difference will be the specific functions we use to allow us to do these things to an external process. Also, just to be clear: external here does not mean "on another host", it just means another process.

Fortunately, the standard Windows API provides all the functions we need to perform cross-process memory operations. They are often recognizable by an `Ex` suffix (for "Extended" or "External") compared to their counterparts that operate within the current process. For example instead of using `VirtualAlloc`, we'll use `VirtualAllocEx`.

As you might expect, in almost each case we'll need to pass our handle (`hProcess`) as an argument to call an `*Ex` function, which tells the Windows kernel which process's memory map to modify.

## Allocating Memory Remotely: `VirtualAllocEx`

This function is the cross-process equivalent of `VirtualAlloc`. It reserves, commits, or changes the state of a region of memory within the virtual address space of a *specified process*.

```c++
LPVOID VirtualAllocEx(
  HANDLE hProcess,         // Handle to the target process
  LPVOID lpAddress,        // Desired starting address (or NULL)
  SIZE_T dwSize,           // Size of the region to allocate
  DWORD  flAllocationType, // Allocation type (e.g., MEM_COMMIT | MEM_RESERVE)
  DWORD  flProtect         // Memory protection (e.g., PAGE_READWRITE)
);
````

**Key Differences from `VirtualAlloc`:**

- **`hProcess`**: This is the crucial first parameter. We must provide the valid handle to the target process that you obtained using `OpenProcess`. This handle must have been opened with the `PROCESS_VM_OPERATION` access right.
- **`lpAddress`**: Specifies the desired starting address _within the target process's_ address space. Passing `NULL` lets the system choose an address within the target process.
- **`dwSize`**: The size of the memory to allocate for our payload.
- **`flAllocationType`**: Typically `MEM_COMMIT | MEM_RESERVE`.
- **`flProtect`**: The initial memory protection. We should once again allocate this as **`PAGE_READWRITE`** initially, intending to change it later.

**Return Value:** If successful, `VirtualAllocEx` returns the base address of the allocated region _within the target process's address space_. This address is only meaningful within the context of that target process. If it fails, it returns `NULL`.

So just to recap: We started off by obtaining a handle to the target process, now hopefully using `VirtualAllocEx` we can obtain the specific starting address inside the process memory where we intend to write our shellcode to.

## Writing to Remote Memory: `WriteProcessMemory`

Once we've allocated memory in the target process (at the address returned by `VirtualAllocEx`), we need to copy our payload (e.g., shellcode) from our loader process into that remote buffer. `WriteProcessMemory` achieves this.

```cpp
BOOL WriteProcessMemory(
  HANDLE  hProcess,            // Handle to the target process
  LPVOID  lpBaseAddress,       // Base address to write TO (in target process)
  LPCVOID lpBuffer,            // Pointer to the data to write FROM (in current process)
  SIZE_T  nSize,               // Number of bytes to write
  SIZE_T  *lpNumberOfBytesWritten // Optional: Pointer to receive bytes actually written
);
```

**Parameters:**

- **`hProcess`**: The handle to the target process, which must have `PROCESS_VM_WRITE` and `PROCESS_VM_OPERATION` access rights.
- **`lpBaseAddress`**: The starting address _within the target process_ where writing should begin - this is the address returned by our  `VirtualAllocEx` call.
- **`lpBuffer`**: A pointer to the buffer _in the current (loader) process_ that contains the data to be written (e.g. our shellcode byte array).
- **`nSize`**: The number of bytes to write from `lpBuffer`.
- **`lpNumberOfBytesWritten`**: A pointer to a variable that will receive the number of bytes successfully written. Can be `NULL` if not needed, but checking it is good practice.

**Return Value:** Returns non-zero (`TRUE`) if successful, zero (`FALSE`) on failure.

## Changing Remote Memory Protections: `VirtualProtectEx`

After writing the shellcode into the remote process's `PAGE_READWRITE` memory region, we need to change its protection to make it executable, following the RW -> RX pattern. `VirtualProtectEx` is the function for modifying memory protections in another process.


```cpp
BOOL VirtualProtectEx(
  HANDLE hProcess,            // Handle to the target process
  LPVOID lpAddress,           // Base address of the region to change (in target process)
  SIZE_T dwSize,              // Size of the region
  DWORD  flNewProtect,        // The NEW desired memory protection (e.g., PAGE_EXECUTE_READ)
  PDWORD lpflOldProtect       // Pointer to receive the OLD protection flags
);
```

**Key Differences from `VirtualProtect`:**

- **`hProcess`**: The handle to the target process, requiring `PROCESS_VM_OPERATION` access right.
- **`lpAddress`**: The starting address _within the target process_ whose protection needs changing (again, typically the address from `VirtualAllocEx`).
- **`dwSize`**: The size of the memory region.
- **`flNewProtect`**: The new protection constant, usually **`PAGE_EXECUTE_READ`** (RX) for executing shellcode.
- **`lpflOldProtect`**: Pointer to a `DWORD` that will receive the previous protection flags.

**Return Value:** Returns non-zero (`TRUE`) if successful, zero (`FALSE`) on failure.

## Putting It Together (Conceptual Workflow)

The entire sequence, including obtaining our initial handle, now looks like this:

1. `OpenProcess` -> Get `hProcess` with required rights (`VM_OPERATION`, `VM_WRITE`, `CREATE_THREAD`, etc.).
2. `VirtualAllocEx(hProcess, NULL, payloadSize, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE)` -> Get `remoteBufferAddress`.
3. `WriteProcessMemory(hProcess, remoteBufferAddress, localPayloadBuffer, payloadSize, NULL)` -> Copy payload to target.
4. `VirtualProtectEx(hProcess, remoteBufferAddress, payloadSize, PAGE_EXECUTE_READ, &oldProtect)` -> Make remote buffer executable.

   So, we're almost there, but not quite yet, in order to execute we'll also still perform the following steps, which we'll cover in Theory 10.3.
5. _(Next Step)_ Use `remoteBufferAddress` as the start address for `CreateRemoteThread`.
6. _(Cleanup)_ Eventually call `VirtualFreeEx` on `remoteBufferAddress` and `CloseHandle` on `hProcess`.

## Conclusion

The Windows API provides specific `Ex` suffixed functions (`VirtualAllocEx`, `VirtualProtectEx`) and `WriteProcessMemory` to allow a process with appropriate permissions (obtained via `OpenProcess`) to allocate, write to, and change the protections of memory within another process's address space.

In the next lab, we will use these functions in Go to perform these remote memory operations.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "find_lab.md" >}})
[|NEXT|]({{< ref "mem_lab.md" >}})