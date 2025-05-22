---
showTableOfContents: true
title: "Havoc C2 Overview & Demon Architecture (Theory 1.1)"
type: "page"
---

## Enter Havoc
Havoc is a modern C2 framework created by [C5pider](https://github.com/Cracked5pider). Compared to established commercial frameworks like Cobalt Strike, Havoc offers an open-source alternative with a strong focus on modern evasion techniques implemented directly within its C/C++ agent.

Havoc differentiates itself primarilly by integrating advanced in-memory evasion techniques, such as sophisticated sleep obfuscation (Ekko, Foliage), indirect syscalls, and hardware breakpoint utilization.

Compared to other popular open-source Go-based frameworks like Sliver or Merlin, Havoc's choice of C/C++ for the Demon allows for more direct and potentially lower-level interaction with the Windows OS and memory, facilitating the implementation of these advanced tradecraft features.

## Architecture


The Havoc framework is comprised of three primary components:

1. **Teamserver:** Implemented in Go, the teamserver is tasked with managing incoming agent connections via configurable listeners (e.g., HTTP/S, SMB). It orchestrates tasking assigned by operators, receives results from agents, maintains state information, logs events, and provides the interface for operator clients.
2. **Client:** Havoc's GUI is developed in C++ and the Qt framework. Once it connects to the Teamserver, it allows operators to view connected agents, issue tasks, manage listeners, examine gathered loot, and collaborate with other team members. _While we won't delve deeply into the Client's implementation in this course, understanding its role as the operator's console is crucial._
3. **Demon:** This is the agent or implant component that executes on the compromised target systems. Written primarily in C and C++, the Demon is responsible for establishing communication back to the Teamserver (via a listener), receiving and executing tasks, and sending back results. It embodies the core post-exploitation capabilities and evasion techniques that are the central focus of this course.

## Core Features
### Demon Agent Capabilities
- Sleep Obfuscation: Techniques like Ekko, Ziliean, or FOLIAGE are used.
- Evasion Techniques: Includes x64 return address spoofing, indirect syscalls for Nt* APIs, and patching Amsi/Etw via hardware breakpoints.
- Functionality: Supports SMB, a token vault for managing stolen tokens, and a variety of built-in post-exploitation commands.
- Indirect Syscalls: Demon can perform indirect syscalls for many Nt* APIs by masquerading the instruction pointer to point within `ntdll.dll`, potentially evading EDR solutions. Syscall stubs are dynamically crafted.

### Extensibility
According to the developer's note: "The Havoc Framework hasn't been developed to be evasive. Rather it has been designed to be as malleable & modular as possible. Giving the operator the capability to add custom features or modules that evades their targets detection system."

- **External C2:** Allows for integration with other C2 channels.
- **Custom Agent Support:** Operators can develop and integrate their own agents, with "Talon" being an example of a third-party agent.
- **Python API:** A Python API (`havoc-py`) facilitates interaction with the teamserver and the development of custom tools and scripts.
- **Modules:** Functionality can be expanded through modules, with official examples including "Powerpick" (for unmanaged PowerShell) and "InvokeAssembly" (for executing .NET assemblies in a separate process).

### Teamserver Customization:
- **Profiles:** The Teamserver uses profiles in `yaotl` format (built on HCL) for configuration, allowing detailed setup of the teamserver, operators, Demon agent defaults, and listeners.
- **Listeners:** Supports HTTP/HTTPS listeners with extensive customization options for hosts, ports, methods, user agents, URIs, and headers.

### *Client Interface:
- Provides views for listeners, session tables, a session graph, and interaction consoles.
- Allows for payload generation and management of agents.

## Docs
The goal of this course is NOT to teach you how to use Havoc, but to analyze some of its key features, learn the
theory behind it, and then implement it ourselves using Golang. If you wanted to learn how to use it, or just for 
more background reading as a supplement to this course, I highly recommend you consult the [official docs](https://havocframework.com/docs/welcome).






---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "internals.md" >}})