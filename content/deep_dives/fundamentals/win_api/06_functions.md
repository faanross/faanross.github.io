---
showTableOfContents: true
title: "Working with Windows API Functions"
type: "page"
---
## Function Documentation and Discovery

The thing you really need to get good at it is not remembering all the functions and how to use them - there's too many for that - but how to find them in the documentation, and how to correctly interpret their use based on this.

You can browse it [online](https://learn.microsoft.com/en-us/windows/apps/api-reference/), you can download them as PDFs
for offline viewing, or some IDEs like Visual Studio will link to them directly while working in the editor. In any case,
finding the function you want to use is not hard, if all else fails just Google for example "VirtualAlloc MSDN".


The trickier part is interpreting the function entry and deducing how to use it correctly, though with time this also becomes fairly straight-forward. We'll look at an actual example below, but it's worth mentioning some of the most important things we'll be on the lookout for here:


1. **Function signature**: This defines the function's name, return type, and parameters, providing a concise summary of its interface.
2. **Parameter descriptions**: Each parameter is described in detail, including its purpose, data type, and any special considerations. The [in], [out], and [in, out] annotations indicate whether parameters provide input to the function, receive output from it, or both.
3. **Return value**: This section explains what the function returns on success and failure, including any special return values that indicate specific conditions.
4. **Remarks**: Often the most valuable section, this provides important details about the function's behaviour, limitations, and interactions with other API functions. Some entries will even link to examples, which are invaluable for deductive purposes.
5. **Requirements**: This section lists the minimum supported Windows versions and the header and library files needed to use the function.


There are other sources which can guide the use of the Windows API, and it's worth taking note of them:
- Header files contain function declarations and related constants, providing insights into available functionality
- Sample code from similar applications found via for example GitHub can demonstrate practical usage of API functions in context
- API monitoring tools can also reveal which functions other applications use to accomplish specific tasks


## Working with API Functions: A Practical Approach

It's worth keeping in mind that the actual function call forms part of a bigger process you need to undertake in order to correctly use the Windows API. This will become second nature after some time, but when starting out it might be worthy using some formal checklist. Though somewhat contrived, it can help serve as "training wheels", and you can of course naturally abandon them once you've got the hang of things.


1. **Locate the appropriate function**: Identify the Windows API function that best addresses your goal, consulting documentation to understand its behavior and requirements.
2. **Prepare parameters**: Declare and initialize the variables needed for function parameters, paying careful attention to data types and [in]/[out] annotations.
3. **Perform the call**: Invoke the function with the prepared parameters, capturing the return value for error checking. The most common errors typically relate to using the wrong types for arguments, especially once ANSI/Unicode wires get crossed.
4. **Check for errors**: Verify the function's success using its return value and, if necessary, `GetLastError()` for additional information.
5. **Process results**: Use any output parameters or return values to accomplish your goal, implementing appropriate error handling if the function failed.
6. **Clean up resources**: Release any allocated resources when they're no longer needed, such as closing handles or freeing memory.


OK, enough abstraction. Let's look at an actual example.



## Calling VirtualAlloc()

Note I'm not going to reproduce the entire entry here, so I 100% encourage you to take a moment and read it from
top to bottom [here](https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualalloc). 


The first thing we want to understand is the function's "basic shape" – what arguments does it take + what does it return?

```cpp
LPVOID VirtualAlloc(
  [in, optional] LPVOID lpAddress,
  [in]           SIZE_T dwSize,
  [in]           DWORD  flAllocationType,
  [in]           DWORD  flProtect
);
```

Okay, it returns an `LPVOID`, which is essentially a generic pointer (`void*`). This will likely be the pointer to our allocated memory. What happens if it _can't_ allocate the memory? We jump down to the `Return value` section. It clearly states: "If the function succeeds, the return value is the base address of the allocated region of pages. If the function fails, the return value is 1 `NULL`."

This tells us our primary error check: after calling `VirtualAlloc`, we _must_ check if the returned pointer is `NULL`. If it is, something went wrong. The docs also mention calling `GetLastError` in case of failure to get more specific error information.

Now, let's take a look at the parameters.

1. `lpAddress` (`LPVOID`): The documentation describes this as "The starting address of the region to allocate." But then it offers a very convenient option: "If this parameter is `NULL`, the system determines where to allocate the region." Since we don't particularly care _where_ the memory is located for this example, passing `NULL` sounds like the easiest approach. Let's go with that.
2. `dwSize` (`SIZE_T`): This seems straightforward – "The size of the region, in bytes." We need to decide how much memory we want. Let's pick a common size, like 4096 bytes (4KB). The documentation mentions some details about rounding up to page boundaries if `lpAddress` is `NULL`, which is fine for us. We just need to provide the desired number of bytes as a `SIZE_T`.
3. `flAllocationType` (`DWORD`): This parameter dictates the _type_ of memory allocation. The documentation lists several flags. We want memory that's ready to use immediately. Scanning the options, `MEM_COMMIT` sounds relevant ("Allocates memory charges... the contents will be zero"). `MEM_RESERVE` also seems useful ("Reserves a range... without allocating any actual physical storage"). Can we do both? Yes! The documentation explicitly guides us: "To reserve and commit pages in one step, call VirtualAlloc with `MEM_COMMIT | MEM_RESERVE`." This is exactly what we need. We'll combine these two flags using the bitwise OR operator (`|`).
4. `flProtect` (`DWORD`): This controls the memory protection – what are we allowed to do with this memory? The docs say we can specify "any one of the memory protection constants." Since our goal is simple read/write memory, the constant `PAGE_READWRITE` seems like the obvious choice.

So, we've dissected the parameters based on our goal and the documentation's guidance. We'll use `NULL` for the address, `4096` for the size, `MEM_COMMIT | MEM_RESERVE` for the allocation type, and `PAGE_READWRITE` for the protection.

Our call will look something like: `LPVOID ptr = VirtualAlloc(NULL, 4096, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);`

If `ptr` comes back as non-`NULL`, we've successfully obtained our memory. The documentation summary even helpfully informs us that "Memory allocated by this function is automatically initialized to zero," which is good to know. We can then cast that `LPVOID` to a usable pointer type (like `char*` or `BYTE*`) and start working with it.

But we're not quite done. We allocated a system resource (memory). Good practice dictates we need to release it when we're finished. Does the documentation mention how? Scrolling down to the `Remarks` section, we find a crucial piece of information: "The `VirtualFree` function can decommit a committed page... or it can simultaneously decommit and release...".

That's our cleanup function. To completely release the memory block we allocated (which was both reserved and committed), we'll need to call `VirtualFree`. A quick look at `VirtualFree` (or common patterns) tells us we'll need to pass it the pointer we got from `VirtualAlloc`, specify a size of `0`, and use the `MEM_RELEASE` flag.

Let's see how this translates into actual code, including the error checking and cleanup we identified as necessary from the documentation.


```cpp
#include <windows.h>
#include <iostream> // For basic output
#include <string.h> // For strcpy_s

int main() {
    SIZE_T allocationSize = 4096; // Our chosen size
    LPVOID pMemory = nullptr;     // Variable to hold the returned pointer

    std::cout << "Attempting to allocate " << allocationSize << " bytes of memory..." << std::endl;

    // Making the call based on our interpretation
    pMemory = VirtualAlloc(
        NULL,                   // Let system choose address
        allocationSize,         // How much memory to allocate
        MEM_COMMIT | MEM_RESERVE, // Reserve and commit in one step
        PAGE_READWRITE          // Make it readable and writable
    );

    // Checking the return value, as the documentation instructed
    if (pMemory == NULL) {
        DWORD dwError = GetLastError(); // Get specific error code if NULL was returned
        std::cerr << "VirtualAlloc failed! Error code: " << dwError << std::endl;
        // In a real app, you might use FormatMessage to get a descriptive string
        return 1; // Exit indicating failure
    }

    // If we get here, allocation succeeded!
    std::cout << "VirtualAlloc succeeded! Memory allocated at address: " << pMemory << std::endl;
    std::cout << "Documentation says this memory should be zeroed." << std::endl;

    // Now we can use the memory (process results)
    char* charPtr = static_cast<char*>(pMemory);
    const char* message = "Testing write to allocated memory!";

    std::cout << "Writing message: '" << message << "'" << std::endl;
    strcpy_s(charPtr, allocationSize, message); // Use safe string copy

    std::cout << "Reading back from memory: '" << charPtr << "'" << std::endl;

    // Don't forget cleanup! Using VirtualFree as indicated by docs.
    std::cout << "Attempting to free the allocated memory..." << std::endl;
    // Size must be 0 when using MEM_RELEASE for blocks allocated this way
    BOOL freeResult = VirtualFree(
        pMemory,       // Pointer to the memory block
        0,             // Size (must be 0 with MEM_RELEASE)
        MEM_RELEASE    // Operation type: release the allocation
    );

    if (!freeResult) {
        DWORD dwError = GetLastError();
        // Log or handle cleanup failure appropriately
        std::cerr << "Warning: VirtualFree failed! Error code: " << dwError << std::endl;
        // Might still be considered overall success if main task worked
    } else {
        std::cout << "Memory successfully freed." << std::endl;
    }

    return 0; // Exit indicating success
}
```

Hopefully, this walkthrough gives you a better feel for how to approach API documentation – looking for the signature, understanding the return value (especially error conditions), dissecting each parameter's meaning and options, and finding related functions for tasks like cleanup.





---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "05_types.md" >}})
[|NEXT|]({{< ref "07_error.md" >}})