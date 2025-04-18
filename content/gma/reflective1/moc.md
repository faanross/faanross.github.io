---
description: "GMA course on reflective loader part 1"
showTableOfContents: true
title: "Let's Build a Reflective Loader - Part 1"
type: "page"
---

## Module 1 - DLLs and Basic Loading
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
- [IAT Resolution Implementation (Lab 4.2)]({{< ref "module04/iat_lab.md" >}})

