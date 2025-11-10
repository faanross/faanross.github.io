---
showTableOfContents: true
title: "Conceptual Overview - What We'll Build"
type: "page"
---


## Overview

In this lesson, we'll get a bird's-eye view of what we're building. Understanding the complete architecture before diving into code will help you see how each lesson contributes to the final system.

We're transforming our basic communication framework into a full command and control (C2) system capable of:

- Receiving commands from operators
- Validating and processing those commands
- Queuing commands for agent pickup
- Executing commands on remote systems
- Reporting results back to operators

In this workshop specifically we'll implement one command - a shellcode loader. But, using the pattern you'll learn as we implement this will allow you to essentially implement any command you desire.


Let's walk through how this system will work in a bit more detail.


## The Complete System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    FINAL C2 ARCHITECTURE                        │
└─────────────────────────────────────────────────────────────────┘

     CLIENT                     SERVER                      AGENT
    ========                    ======                      =====

       │                                                      │
       │  1. POST /command (:8080)                            │
       │     Send Command + Arguments                         │
       │                                                      │
       │                                                      │
       ├─────────────────────────►                            │
       │                         │                            │
       │                    Validate Command                  │
       │                         │                            │
       │                    Validate Arguments                │
       │                         │                            │
       │                    Process Arguments                 │
       │                    (file → base64)                   │
       │                         │                            │
       │                    Queue Command                     │
       │                         │                            │
       │  2. "Command queued"    │                            │
       │  ◄──────────────────────┤                            │
       │                         │                            │
       │                         │    3. GET /  (:8443)       │
       │                         │    (periodic check-in)     │
       │                         │  ◄─────────────────────────┤
       │                         │                            │
       │                    Check Queue                       │
       │                         │                            │
       │                    Dequeue Command                   │
       │                         │                            │
       │                    Generate Job ID                   │
       │                         │                            │
       │                         │  4. ServerResponse         │
       │                         │     Command, Args etc      │
       │                         │                            │
       │                         │                            │
       │                         │                            │
       │                         ├─────────────────────────►  │
       │                         │                            │
       │                         │                       ExecuteTask
       │                         │                            │
       │                         │                       Orchestrator
       │                         │                            │
       │                         │                       Doer (OS-specific)
       │                         │                            │
       │                         │                       Build Result
       │                         │                            │
       │                         │  5. POST /results (:8443)  │
       │                         │     AgentTaskResult        │
       │                         │  ◄─────────────────────────┤
       │                         │                            │
       │                    Display Result                    │
       │                         │                            │
       │                         │                            │
```




## The Journey: What Changes Each Lesson

Let's see how the system evolves lesson by lesson.

### Current State (Starting Code)

**What we have:**

- Agent checks in periodically
- Server responds with a simple string
- No command functionality

```
Agent → GET / → Server
Agent ← "You hit the endpoint" ← Server
```

### Lessons 1-2: Command Reception and Validation

**What we add:**

- `/command` endpoint for operators
- Command registry to validate command keywords
- Immediate rejection of invalid commands

```
Operator → POST /command {"command": "shellcode"} → Server
         → Validate: Does "shellcode" exist?
         → YES: Continue | NO: Reject
```


### Lessons 3-4: Argument Validation and Processing

**What we add:**

- Every type of command has its own unique arguments
- For example:
    - shell command: run `whoami`
    - download: download this specific file
    - etc
- These command-specific arguments need to be validated - are they the right/expected arguments? Are the valid?
- Then, in some cases, the arguments need to be processed before they are ready for the agent
- For example for our shellcode loader, we will provide as an argument the path to the DLL containing the shellcode, but we don't want to send this to the agent, we want to send the actual encoded data from the DLL!


```
Operator → POST /command 
           {"command": "shellcode",
            "data": {"file_path": "./payloads/calc.dll", "export_name": "LaunchCalc"}}
         → Validate: calc.dll exists? export_name present?
         → Process: Read file → Convert to base64
```



### Lesson 5: Command Queuing

**What we add:**

- We will implement a command queue
- Why? We just received an instruction from the operator, but might be some time until the agent checks in again
- The queue is a thread-safe "container" where it will reside until the agent connects to server


```
Client → Command validated & processed
         → Add to queue
         → Response: "Command queued"

Queue: [Command1, Command2, Command3] (FIFO)
```




### Lesson 6: Sending Commands to Agent

**What we add:**

- When an agent checks in we need to see if there is something in queue
- If there is, remove it from queue (dequeue), add additional fields like Job ID, and send to agent


```
Agent → GET / → Server checks queue
              → If command: Dequeue, generate job ID, send
              → If empty: Send {"job": false}
```






### Lesson 7: Agent Execution Framework

**What we add:**

- Once agent receives command needs a dedicated function to route command to correct function `ExecuteTask()`
- We also need to implement a command orchestrators registration system - which commands are available on the agent?
- This same function - `ExecuteTask()` - will also send back the final result using a new function called `SendResult()`



### Lessons 8-9: Shellcode Orchestration

**What we add:**

- `orchestrateShellcode()` - Prepares arguments for execution
- Calls the doer function via the shellcode interface, contains OS-specific implementation of the command

```
ExecuteTask
  ↓
orchestrateShellcode
  ↓ Unmarshal arguments
  ↓ Validate (agent-side)
  ↓ Decode base64 → raw bytes
  ↓
DoShellcode (interface)
  ↓
OS-specific implementation
```



### Lessons 10-11: Windows Shellcode Loader

**What we add:**

- This is where the magic happens - allows us to actually execute the shellcode in memory and run an arbitrary process
- Includes:
    - PE header parsing
    - Memory allocation and section mapping
    - Import resolution (IAT patching)
    - Base relocations
    - DllMain and export calling

```
Raw DLL bytes
  ↓ Parse PE headers
  ↓ Allocate memory
  ↓ Map sections (.text, .data, etc.)
  ↓ Process relocations (fix addresses)
  ↓ Resolve imports (patch IAT)
  ↓ Call DllMain
  ↓ Find and call exported function
  ↓ CALC.EXE LAUNCHES! 🎉
```


### After Lesson 12: Results Display

**What we add:**

- Need to add another endpoint on server to receive and display the final result from agent
- `/results` endpoint on server


```
Agent executes command
  ↓ Build AgentTaskResult
  ↓ POST /results
  ↓
Server receives
  ↓ Parse result
  ↓ Extract message
  ↓ Display: "Job (ID: job_123456) succeeded"
```

**Complete loop:**

```
Operator → Command → Server → Queue → Agent → Execute → Result → Server → Display
```


## Conclusion

That's it for our preview. I don't want to get lost in conceptual previews, so now the time is perfect for us to jump in and build on our starting code.



___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./00B_starting.md" >}})
[|NEXT|]({{< ref "./01_endpoint.md" >}})