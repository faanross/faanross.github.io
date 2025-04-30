---
showTableOfContents: true
title: "Process Injection Introduction & Target Selection (Theory 10.1)"
type: "page"
---
## Overview
In Modules 1 - 8 we build a reflective loader capable of executing remote obfuscated shellcode. In Module 9 we then focussed on refining the execution of our payload *within the context of our own loader process*. We improved memory permissions (RW -> RX), added delays + decoy function, and obfuscated the static shellcode using runtime decryption.

While these are valuable steps, it also revolved around injecting our payload into the memory space of our actual loader process itself. This is generally not preferred since it not only comes with additional risk, but it also does not capitalize on potential benefits of other techniques.

One common alternative then is the use of **process injection**: instead of injecting then payload into our own (i.e. the calling process) memory space, we inject it into the memory space of another process.

So in this Module 10 then we'll explore the fundamentals of process injection using the standard, documented Windows API (WinAPI).

## Why Inject Code into Another Process?

Injecting code into a different process offers several potential advantages compared to running it in your initial loader.

### Evasion & Stealth
Running malicious code within the context of a legitimate, trusted process (e.g., `explorer.exe`, `svchost.exe`, or even a common browser) can make it less conspicuous. Security tools might apply less scrutiny to threads and memory allocations within these known-good processes compared to an unknown or unsigned loader executable. It helps the malicious code "blend in" with normal system activity.

This should make sense - consider that for example we're often creating an outbound connection to a C2 server using this process, if it's just some random, unknown process, or notepad.exe connecting back for days/weeks on end, well that is much more likely to draw scrutiny vs a process that "makes sense" like `chrome.exe` or `svchost.exe`.

### Leveraging Trust & Permissions
Some processes run with higher privileges or integrity levels, or might already be allowed through host-based firewalls. Injecting into such a process could potentially grant the payload inherited permissions or network access that the initial loader process didn't have. (Note: Injecting into a higher-privilege process typically requires the injector process to already have sufficient privileges, like `SeDebugPrivilege`).

### Persistence
While not covered in basic injection, some persistence techniques involve injecting code into long-running system processes so the payload survives user logoffs or reboots (though more robust persistence usually involves other methods).

### Decoupling from Loader (Resilience)
If the initial loader process is terminated (e.g., by AV/EDR or the user), the injected code running in the separate target process can continue to execute independently.


## Finding Target Processes: Toolhelp Snapshots

Before we can inject into a process, you need to find which actual process we want to inject into. This is usually done by looking at all the processes running on a system + their Process ID (PID). While we could inject into a process we created ourselves, it's often more desirable to inject into an existing, legitimate process.


The Windows API provides a mechanism for enumerating running processes called **Toolhelp Snapshots**, which involves the use of a few key functions.

### `CreateToolhelp32Snapshot`
- This function takes a snapshot of specified system information, including running processes and threads.
- We pass flags indicating what we want to include in the snapshot.
- For processes, the flag is `TH32CS_SNAPPROCESS`.
- It returns a handle to the snapshot object.
```c++
    // Example Snippet (Conceptual C++)
    #include <tlhelp32.h>
    // ...
    HANDLE hSnap = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if (hSnap == INVALID_HANDLE_VALUE) { /* Handle error */ }

```


### `Process32FirstW` / `Process32First`
- After creating the snapshot, we call this function (use the `W` version for Unicode compatibility) with the snapshot handle and a pointer to a `PROCESSENTRY32W` structure.
- It populates the structure with information about the *first* process found in the snapshot.
- We need to initialize the `dwSize` member of the structure before calling.

    ```c++
    // Example Snippet (Conceptual C++)
    PROCESSENTRY32W pe32;
    pe32.dwSize = sizeof(PROCESSENTRY32W); // IMPORTANT: Initialize size
    if (!Process32FirstW(hSnap, &pe32)) { /* Handle error, maybe no processes? CloseHandle(hSnap); */ }
    ```


### `Process32NextW` / `Process32Next`

-  We can then call this function repeatedly in a loop, passing the same snapshot handle and `PROCESSENTRY32W` structure pointer.
- Each successful call populates the structure with information about the *next* process in the snapshot.
- The loop continues until `Process32NextW` returns `FALSE`, indicating the end of the process list.

    ```c++
    // Example Snippet (Conceptual C++)
    do {
        // Process information in pe32 structure
        // pe32.th32ProcessID holds the PID
        // pe32.szExeFile holds the executable name (WCHAR array)
        printf("PID: %lu, Name: %ws\n", pe32.th32ProcessID, pe32.szExeFile);

        // Find target process by comparing pe32.szExeFile to desired name
        // if (wcscmp(pe32.szExeFile, L"notepad.exe") == 0) {
        //     targetPID = pe32.th32ProcessID;
        //     break; // Found it
        // }

    } while (Process32NextW(hSnap, &pe32)); // Get next process
    ```

### `CloseHandle`
- Once finished, we must close the snapshot handle using `CloseHandle(hSnap)`.

Inside the loop, the `PROCESSENTRY32W` structure gives us the `th32ProcessID` (the PID) and `szExeFile` (the executable name) for each process, allowing us to identify your target process by name and retrieve its PID.

## Requesting Permissions: Process Access Rights

But, as is often the case, just knowing the PID isn't enough. To interact with another process's memory and threads, we need to obtain a **handle** to that process with the appropriate **access rights** (permissions). When we try to get this handle, the system performs security checks based on the privileges of our current process and the security descriptor of the target process.

The function we use to get this handle is `OpenProcess`. It takes the desired access rights as one of its main arguments. For typical code injection, we'll need a combination of rights, including:

* `PROCESS_VM_OPERATION`: Required for `VirtualProtectEx` and often for other memory operations.
* `PROCESS_VM_READ`: Required for `ReadProcessMemory` (if you need to read from the target).
* `PROCESS_VM_WRITE`: Required for `WriteProcessMemory` (to write your shellcode).
* `PROCESS_CREATE_THREAD`: Required for `CreateRemoteThread`.
* `PROCESS_QUERY_INFORMATION` / `PROCESS_QUERY_LIMITED_INFORMATION`: Often needed to query basic information about the target process (like its architecture).

We typically combine these desired rights using the bitwise OR operator (`|`). A common combination for injection is:
`PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ`

Attempting to open a process with rights we aren't privy to (e.g., trying to open a highly privileged system process from a low-privilege user process) will cause `OpenProcess` to fail.

## Opening the Target: `OpenProcess`

Once you have the target `PID` and have decided on the necessary `dwDesiredAccess` rights, we can call `OpenProcess`:

```c++
// Example Snippet (Conceptual C++)
#include <windows.h>
// ...
DWORD targetPID = 1234; // PID obtained from Toolhelp snapshot
DWORD dwDesiredAccess = PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ;

HANDLE hProcess = OpenProcess(dwDesiredAccess, FALSE, targetPID);
// bInheritHandle is typically FALSE

if (hProcess == NULL) {
    // Failed to open process. Check GetLastError().
    // Common reasons: Access Denied (5), Invalid Parameter (87 - PID doesn't exist?)
} else {
    // Success! hProcess is now a valid handle to the target process
    // Proceed with VirtualAllocEx, WriteProcessMemory, etc.
    // ...
    CloseHandle(hProcess); // IMPORTANT: Close the handle when done!
}
````

- `dwDesiredAccess`: The access rights we are requesting (bitwise OR flags).
- `bInheritHandle`: Usually `FALSE`, indicating whether processes created by the current process should inherit the handle.
- `dwProcessId`: The PID of the target process.

If successful, `OpenProcess` returns a valid `HANDLE` that can be used in subsequent API calls like `VirtualAllocEx`, `WriteProcessMemory`, `VirtualProtectEx`, and `CreateRemoteThread`. If it fails, it returns `NULL`, and `GetLastError()` can provide more details on the failure reason. Remember we should always close the obtained handle using `CloseHandle` when we are finished with it.

## Conclusion
So to recap: In order for us to inject into another process we first need to enumerate processes on a target host, allowing us to decide on a PID. Once we have a PID and we've constructed our desired access rights we can then hopefully obtain a handle, which can is the gateway to performing the actual injection steps, which we'll cover after our following lab.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module09/thread.md" >}})
[|NEXT|]({{< ref "find_lab.md" >}})