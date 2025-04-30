---
showTableOfContents: true
title: "An Epic Introductory Journey in Malware Development"
type: "page"
---

**Last Updated: 29 April 2025**

## Preface (still to do)

- [Course Introduction]({{< ref "module00/hello.md" >}})
- [Course Preview + Curriculum]
- [Recommended Lab Setup]


## Module 1: DLLs and Basic Loading
- [Introduction to DLLs (Theory 1.1)]({{< ref "module01/intro_DLLs.md" >}})
- [Introduction to Shellcode (Theory 1.2)]({{< ref "module01/intro_shellcode.md" >}})
- [Standard DLL Loading in Windows (Theory 1.3)]({{< ref "module01/dll_loading.md" >}})
- [Create a Basic DLL (Lab 1.1)]({{< ref "module01/create_dll.md" >}})
- [Create a Basic Loader in Go (Lab 1.2)]({{< ref "module01/create_loader.md" >}})


## Module 2: PE Format for Loaders
- [PE File Structure Essentials (Theory 2.1)]({{< ref "module02/structure.md" >}})
- [Addressing in PE Files (Theory 2.2)]({{< ref "module02/addresses.md" >}})
- [PE Header Inspection with PE-Bear (Lab 2.1)]({{< ref "module02/pebear.md" >}})
- [PE Header Parser in Go (Lab 2.2)]({{< ref "module02/peparser.md" >}})


## Module 3: Reflective DLL Loading Core Logic
- [Intro to Reflective DLL Loading (Theory 3.1)]({{< ref "module03/intro.md" >}})
- [Memory Allocation (Theory 3.2)]({{< ref "module03/memalloc.md" >}})
- [Mapping the DLL Image (Theory 3.3)]({{< ref "module03/mapping.md" >}})
- [Manual DLL Mapping in Go (Lab 3.1)]({{< ref "module03/maplab.md" >}})

## Module 4: Handling Relocations and Imports
- [Base Relocations (Theory 4.1)]({{< ref "module04/reloc.md" >}})
- [IAT Resolution (Theory 4.2)]({{< ref "module04/iat.md" >}})
- [Intentional Base Relocation (Lab 4.1)]({{< ref "module04/reloc_lab.md" >}})
- [IAT Processing (Lab 4.2)]({{< ref "module04/iat_lab.md" >}})


## Module 5: Execution and Exports
- [The DLL Entry Point (Theory 5.1)]({{< ref "module05/entry.md" >}})
- [Exported Functions (Theory 5.2)]({{< ref "module05/export.md" >}})
- [Call DllMain (Lab 5.1)]({{< ref "module05/entry_lab.md" >}})
- [Call Exported Function (Lab 5.2)]({{< ref "module05/export_lab.md" >}})


## Module 6: Basic Obfuscation - XOR
- [Introduction to Obfuscation (Theory 6.1)]({{< ref "module06/intro.md" >}})
- [Simple XOR (Theory 6.2)]({{< ref "module06/simple.md" >}})
- [XOR Functions in Go (Lab 6.1)]({{< ref "module06/xor_lab.md" >}})
- [Obfuscated Loading (Lab 6.2)]({{< ref "module06/load_lab.md" >}})

## Module 7: Rolling XOR & Key Derivation
- [Rolling XOR (Theory 7.1)]({{< ref "module07/rolling.md" >}})
- [Key Derivation Logic (Theory 7.2)]({{< ref "module07/key.md" >}})
- [Implementing Rolling XOR (Lab 7.1)]({{< ref "module07/rolling_lab.md" >}})
- [Implementing Key Derivation (Lab 7.2)]({{< ref "module07/key_lab.md" >}})

## Module 8: Network Delivery & Client/Server
- [Client + Server Communication (Theory 8.1)]({{< ref "module08/client_server.md" >}})
- [Communication Protocol Design (Theory 8.2)]({{< ref "module08/protocol.md" >}})
- [Environmental Keying + Client ID (Theory 8.3)]({{< ref "module08/client_id.md" >}})
- [Client + Server Logic (Lab 8.1)]({{< ref "module08/cs_lab.md" >}})
- [Implement Client ID and Key Derivation (Lab 8.2)]({{< ref "module08/key_lab.md" >}})

## Module 9: Refining In-Process Execution
- [Decoupling Memory Permissions (Theory 9.1)]({{< ref "module09/decouple.md" >}})
- [Introducing Basic Delays and Misdirection (Theory 9.2)]({{< ref "module09/delay.md" >}})
- [Decoupling, Delays, and Misdirections (Lab 9.1)]({{< ref "module09/decouple_lab.md" >}})
- [Shellcode Encryption & Decryption In-Memory (Theory 9.3)]({{< ref "module09/encrypt.md" >}})
- [Implementing Runtime Shellcode Decryption (Lab 9.2)]({{< ref "module09/encrypt_lab.md" >}})
- [Basic Thread Obfuscation Concepts (Theory 9.4)]({{< ref "module09/thread.md" >}})


## Module 10: Process Injection Fundamentals (WinAPI)
- [Process Injection Introduction & Target Selection (Theory 10.1)]({{< ref "module10/process.md" >}})
- [Finding and Opening Target Processes (Lab 10.1)]({{< ref "module10/find_lab.md" >}})
- [Remote Memory Operations (WinAPI) (Theory 10.2)]({{< ref "module10/mem.md" >}})
- [Performing Remote Memory Operations (Lab 10.2)]({{< ref "module10/mem_lab.md" >}})
- [Remote Thread Execution (WinAPI) (Theory 10.3)]({{< ref "module10/remote.md" >}})
- [Executing Code via CreateRemoteThread (Lab 10.3)]({{< ref "module10/remote_lab.md" >}})
- [Classic DLL Injection (WinAPI) (Theory 10.4)]({{< ref "module10/classic.md" >}})
- [Implementing Classic DLL Injection (Lab 10.4)]({{< ref "module10/classic_lab.md" >}})





## MORE TO COME... WIP

___
