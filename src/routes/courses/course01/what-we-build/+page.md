---
layout: course01
title: "What We'll Build"
---

Before diving into code, let's understand the complete architecture of the C2 framework we're building.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                            OPERATOR                             │
│                        (You, the attacker)                      │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                       CONTROL API (:8080)                       │
│                  curl commands to queue tasks                   │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                            SERVER                               │
│   ┌──────────────────┐       ┌──────────────────┐               │
│   │   HTTPS Server   │       │    DNS Server    │               │
│   │     (:8443)      │       │     (:5353)      │               │
│   └──────────────────┘       └──────────────────┘               │
│               │                       │                         │
│               └───────────┬───────────┘                         │
│                           ▼                                     │
│                  ┌──────────────────┐                           │
│                  │  Command Queue   │                           │
│                  └──────────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
                                 │
               ┌─────────────────┼─────────────────┐
               ▼                 ▼                 ▼
        ┌───────────┐     ┌───────────┐     ┌───────────┐
        │   AGENT   │     │   AGENT   │     │   AGENT   │
        │ (Target 1)│     │ (Target 2)│     │ (Target 3)│
        └───────────┘     └───────────┘     └───────────┘
```

## Core Components

### 1. The Server

The server is the heart of our C2 infrastructure:

- **HTTPS Listener** - Primary communication channel on port 8443
- **DNS Listener** - Fallback channel using DNS queries on port 5353
- **Control API** - HTTP interface on port 8080 for operator commands
- **Command Queue** - Stores pending commands for agents

### 2. The Agent

The agent runs on target machines:

- **Multi-Protocol Support** - Can communicate via HTTPS or DNS
- **Protocol Switching** - Dynamically changes communication method
- **Run Loop** - Periodic check-in with the server
- **Execution Framework** - Runs commands received from server
- **Shellcode Loader** - Executes DLLs in memory (Windows)

### 3. The Control API

The operator interface:

- **Command Submission** - Queue commands for agents
- **Validation** - Verify command syntax and parameters
- **Processing** - Transform commands for agent consumption

## Communication Flow

```
1. Operator → POST /command → Server
   "Execute shellcode on agent"

2. Server validates and queues command

3. Agent → GET / → Server (periodic check-in)
   "Any commands for me?"

4. Server → Response → Agent
   {command: "shellcode", data: {...}}

5. Agent executes command

6. Agent → POST /results → Server
   "Command completed successfully"
```

## Command Execution Architecture

The agent uses a layered execution model:

```
┌──────────────────────────────────────────────────────────────┐
│                          RUN LOOP                            │
│               (Periodic server communication)                │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                        EXECUTE TASK                          │
│                    (Command dispatcher)                      │
└──────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
     ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
     │ ORCHESTRATOR│   │ ORCHESTRATOR│   │ ORCHESTRATOR│
     │  Shellcode  │   │  Download   │   │  Persist    │
     └─────────────┘   └─────────────┘   └─────────────┘
            │                 │                 │
            ▼                 ▼                 ▼
     ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
     │    DOER     │   │    DOER     │   │    DOER     │
     │ (Win/Mac/   │   │ (File I/O)  │   │ (Registry)  │
     │  Linux)     │   │             │   │             │
     └─────────────┘   └─────────────┘   └─────────────┘
```

Each command has:
- **Orchestrator** - Validates arguments and coordinates execution
- **Doer** - Performs the actual OS-specific operation

## Security Features

### HMAC Authentication

Every message between agent and server is authenticated:

```
Message + Secret Key → HMAC-SHA256 → Signature
```

The server verifies the signature before processing.

### Payload Encryption

Sensitive data is encrypted using AES-256-GCM:

```
Plaintext → AES-256-GCM + Nonce → Ciphertext
```

This protects command data in transit.

## What You'll Implement

By the end of this course, you'll have built:

| Component | Description |
|-----------|-------------|
| Server interfaces | Abstraction for multiple protocols |
| HTTPS server | TLS-encrypted web server |
| DNS server | DNS-based C2 channel |
| HTTPS agent | Client-side HTTPS communication |
| DNS agent | Client-side DNS communication |
| Protocol switching | Dynamic protocol selection |
| HMAC auth | Message authentication |
| AES encryption | Payload encryption |
| Command API | Operator interface |
| Command queue | Job management |
| Execution framework | Agent command execution |
| Shellcode loader | Reflective DLL loading |
| Download command | File exfiltration |
| Persistence | Registry-based survival |

## Ready to Code

Now that you understand what we're building, let's start with [Lesson 1: Interfaces and Factory Functions](/courses/course01/lesson-01).

---

[Previous: Setup](/courses/course01/setup) | [Next: Lesson 1 - Interfaces and Factory Functions](/courses/course01/lesson-01) | [Course Home](/courses/course01)
