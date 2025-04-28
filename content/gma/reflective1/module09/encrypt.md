---
showTableOfContents: true
title: "Shellcode Encryption & Decryption In-Memory (Theory 9.3)"
type: "page"
---
## Overview
The title of this lesson might throw you off since we've clearly already implemented encryption/decryption in Modules 6 and 7, no? Well, of course we did, but there is still a gap.


## Breakdown
Let's break this down, starting with our current implementation:
1. When we run our agent/loader, it connects to our server.
2. The server derives a shared key using parameters and disguised constants.
3. The server uses this key to encrypt the **entire** DLL payload (shellcode + execution function), and sends it over to our agent.
4. Our agent then decrypts it, resulting in the raw `dllBytes`.
5. This raw `dllBytes` (i.e. the complete, deobfuscated DLL) is _then_ passed to the reflective loading logic (parsing PE headers, allocating memory, mapping sections, resolving imports, etc.).


## The Gap
The issue is here at Steps 4 + 5 - after the entire payload has been deobfuscated it is reflectively loaded into memory. Then the exported function `LaunchCalc` within our DLL gets called. Inside `LaunchCalc`, it calls our `ExecuteShellcode` function.

So, the logic inside of our DLL is something like:

```C++
unsigned char calc_shellcode[] = {
    0x50, 0x51, 0x52, /* ... more bytes ... */ 0x58, 0xC3
};

BOOL ExecuteShellcode() {
    // ...
    // Copies the raw calc_shellcode bytes
    memcpy(exec_memory, calc_shellcode, sizeof(calc_shellcode));
    // ...
    // Decrypts nothing, because calc_shellcode is plaintext
    // ...
    // VirtualProtect 
}
```


So our issue is this: though we have obfuscated our DLL as a whole during download and initial loading, our actual `calc_shellcode` byte array _itself_, residing within the C++ DLL's data section, is still stored in **plaintext**. In other words, there are two levels which we want to obfuscate - the entire DLL (which we've done), but then within the DLL itself we want to also obfuscate the shellcode.


## Why Obfuscate the Shellcode Itself?

If some detection software captures `dllBytes` from our Go agent's memory _after_ deobfuscation but before it's fully loaded/executed, it could analyze this plaintext DLL and easily extract the raw `calc_shellcode`. So by not obfuscating the shellcode itself within an obfuscated DLL we are just increasing its exposure to potential memory scanning, which is just unnecessary.


## Solution

In this case we'll just implement simple XOR encryption to illustrate the point.  For the purpose of the course I don't want to repeat the same lessons, but you are now of course equipped with the knowledge on how to improve simple XOR, so please by all means go ahead and do so for "extra credit".

Before we go ahead and implement it in our next lab, I do want to provide a brief overview of the process:
1. **Encrypt Offline:** XOR the `calc_shellcode` byte array with a key (we'll use `0xDEADC0DE`) and store this _encrypted_ version in your C++ DLL code.
2. **Runtime Decryption:** In `ExecuteShellcode`, after allocating RW memory and copying the _encrypted_ bytes into it, perform an in-place XOR decryption using the same key.
3. **Proceed:** Continue with the rest of the steps (optional delays, `VirtualProtect` to RX, `CreateThread`).


## Conclusion
So, we'll create a new separate file to encrypt our shellcode, copy this new shellcode into our DLL source file, and also add a decryption function in the payload which will run prior to changing our memory permissions + execution. So let's get to it.



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "decouple_lab.md" >}})
[|NEXT|]({{< ref "encrypt_lab.md" >}})