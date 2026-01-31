---
layout: course01
title: "Welcome to Building a C2 Framework in Go"
---

Welcome to this comprehensive 16-hour course where you'll build a complete Command and Control (C2) framework from scratch using Go.

## What You'll Learn

In this course, you'll develop a fully functional C2 framework with:

- **Multiple Communication Protocols** - HTTPS and DNS-based communication channels
- **Protocol Switching** - Dynamic transition between protocols based on operational needs
- **Secure Communications** - HMAC authentication and payload encryption
- **Command Infrastructure** - Full command queuing, validation, and execution pipeline
- **Agent Execution Framework** - Modular architecture for running commands on target systems
- **Shellcode Execution** - Reflective DLL loading on Windows targets
- **Persistence Mechanisms** - Registry-based persistence for agent survival

## Course Philosophy

This course follows a **"build to understand"** approach. Rather than using existing C2 frameworks, you'll construct every component yourself. This gives you:

1. **Deep Understanding** - Know exactly how C2 frameworks work under the hood
2. **Customization Skills** - Ability to modify and extend any C2 tool
3. **Detection Knowledge** - Understanding attack patterns helps with defense
4. **Go Proficiency** - Practical experience with a language ideal for security tooling

## Prerequisites

- Basic Go programming knowledge (functions, structs, interfaces)
- Understanding of HTTP/HTTPS concepts
- Familiarity with DNS basics
- Access to a Windows VM for testing (shellcode execution)
- A Linux or macOS development environment

## Course Structure

The course is divided into logical sections:

1. **Foundation (Lessons 1-4)** - Core interfaces, HTTPS server/agent, and run loop
2. **Multi-Protocol (Lessons 5-7)** - DNS communication channel
3. **Protocol Management (Lessons 8-10)** - Switching and dual-server architecture
4. **Security (Lessons 11-12)** - Authentication and encryption
5. **Command System (Lessons 13-16)** - Full command pipeline
6. **Execution Framework (Lessons 17-20)** - Agent execution and shellcode loading
7. **Completion (Lessons 21-23)** - Results handling, file operations, persistence

## Ready to Begin?

Head to the [Setup](/courses/course01/setup) page to configure your development environment, then proceed to [What We'll Build](/courses/course01/what-we-build) for an architectural overview.

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/setup">Next: Setup →</a></div>
</div>
