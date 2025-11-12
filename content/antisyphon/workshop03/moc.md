---
showTableOfContents: false
# tags: ["",]
title: "Let's Build a Reflective Loader + C2 Channel in Golang"
type: "page"
---

## Overview
Below are the lecture notes for my AntiSyphon workshop presented **January 23, 2026**. Though the notes are in
general more descriptive than the actual lectures, they are not expanded, meaning the content from the lectures roughly map
1:1 onto these notes.


## Solutions
All the final solutions are available on GitHub here. You'll find a copy of the complete slides, as well as the
lectures available here in markdown format. 







## Foundation
- [Setup Guide]({{< ref "./00A_setup.md" >}})
- [Review of Starting Code]({{< ref "./00B_starting.md" >}})
- [Conceptual Overview - What We'll Build]({{< ref "./00C_overview.md" >}})

## Server Implementation
- [Lesson 1: Implement Command Endpoint]({{< ref "./01_endpoint.md" >}})
- [Lesson 2: Validate Command Exists]({{< ref "./02_validate_command.md" >}})
- [Lesson 3: Validate Command Arguments]({{< ref "./03_validate_argument.md" >}})
- [Lesson 4: Process Command Arguments]({{< ref "./04_process_arguments.md" >}})
- [Lesson 5: Queue Commands]({{< ref "./05_queue.md" >}})
- [Lesson 6: Dequeue and Send Commands to Agent]({{< ref "./06_dequeue.md" >}})


## Agent Implementation
- [Lesson 7: Create Agent Command Execution Framework]({{< ref "./07_execute_task.md" >}})
- [Lesson 8: Implement Shellcode Orchestrator]({{< ref "./08_orchestrator.md" >}})
- [Lesson 9: Create Shellcode Doer Interface and Implementations]({{< ref "./09_interface.md" >}})
- [Lesson 10: Implement Windows Shellcode Doer]({{< ref "./10_doer.md" >}})
- [Lesson 11: Server Receives and Displays Results]({{< ref "./11_result_ep.md" >}})


## Wrap-Up
- [Conclusion and Review]({{< ref "./12_conclusion.md" >}})


___

