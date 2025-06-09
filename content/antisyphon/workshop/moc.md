---
showTableOfContents: false
# tags: ["",]
title: "Let’s Build a Simple C2 in Golang and Vue.js"
type: "page"
---

## Overview
Below are the lecture notes for my AntiSyphon workshop presented on April 25th 2025. Though the notes are in 
general more descriptive than the actual lectures, they are not expanded, meaning the content from the lectures roughly map
1:1 onto these notes. This means some shortcuts were taken as I had to ensure the entire course could fit into the 
allotted 4 hours.


## Solutions
All the final solutions are available on GitHub [here](https://www.github.com/faanross/workshop_antisyphon_25042025).

## Part A: Introduction
- [Overview of What We'll Create]({{< ref "part_a/overview.md" >}})
- [Project Setup (Lab 00)]({{< ref "part_a/setup.md" >}})

## Part B: Server Setup
- [Basic Listener, Handler, Router (Lab 01)]({{< ref "part_b/lab01.md" >}})
- [Brief Introduction to Goroutines (Lab 02)]({{< ref "part_b/lab02.md" >}})

## Part C: Agent Setup
- [Basic Agent Setup and Config System (Lab 03)]({{< ref "part_c/lab03.md" >}})
- [Agent UUID System and Server Middleware (Lab 04)]({{< ref "part_c/lab04.md" >}})

## Part D: Client Setup
- [Minimal Client UI Setup with Vue.js (Lab 05)]({{< ref "part_d/lab05.md" >}})
- [Adding Websocket to Server (Lab 06)]({{< ref "part_d/lab06.md" >}})
- [Adding Websocket to Client (Lab 07)]({{< ref "part_d/lab07.md" >}})

## Part E: Weaving It All Together
- [Agent Command Execution (Lab 08)]({{< ref "part_e/lab08.md" >}})
- [Client UI Command + Results Handling (Lab 09)]({{< ref "part_e/lab09.md" >}})
- [Server Receives + Queues Commands (Lab 10)]({{< ref "part_e/lab10.md" >}})
- [Agent Retrieves Command From Queue (Lab 11)]({{< ref "part_e/lab11.md" >}})
- [Agent Executes, Returns Result, Server Result Processing (Lab 12)]({{< ref "part_e/lab12.md" >}})

## Part F: Conclusion
- [Workshop Review]({{< ref "part_f/review.md" >}})
- [Final Thoughts]({{< ref "part_f/final.md" >}})

___

