---
showTableOfContents: true
title: "Lesson 4: Process Command Arguments"
type: "page"
---
## Solutions

The starting solution can be found here.

The final solution can be found here.



## Overview

Just like we validated command-specific arguments, sometimes we also need to **process** them - prepare them before they can be received by the agent. Note the word *sometimes* - for some commands we'll be able to send the same arguments we receive from client directly to the agent. In other cases however, like with our shellcode loader, we have to do some transformation first.

For our shellcode command, the client sends us a file path, but we can't send a file path to the agent (remember the agent is on a different machine with a different filesystem). Instead, we need to:

1. Read the DLL file from disk
2. Convert the binary data to base64
3. Send the base64 string to the agent

In this lesson then we'll:

1. Create a function type for command processors
2. Create a new argument type for agent-bound data
3. Implement the processor for shellcode
4. Integrate processing into our command handler

## What We'll Create

- `CommandProcessor` function type in `command_api.go`
- `ShellcodeArgsAgent` type in `models/types.go`
- `processShellcodeCommand` function in `shellcode.go`
- Updated command registry with processors
- Processing logic in `commandHandler`


___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./03_validate_argument.md" >}})
[|NEXT|]({{< ref "./05_queue.md" >}})