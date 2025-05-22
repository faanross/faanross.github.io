---
showTableOfContents: true
title: "Let's Build a Reflective Loader in Golang"
type: "page"
---
<br>

![firestarter](../img/keif.gif)

<br>

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

<br>

![dancing_ghost](../img/max.gif)

___
