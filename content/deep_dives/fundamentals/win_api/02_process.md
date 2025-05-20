---
showTableOfContents: true
title: "What is a Process?"
type: "page"
---

## Overview

Before we continue discussing the various components of the Windows Architecture, it's important to get a grip on what's arguably the most important one of all - the **process**. The **process** is the fundamental unit of execution, and we'll deal with constantly, so it's important to have some insight into what exactly a process is, and what it consists of.

A process can be thought of consisting of 5 essential components - the Executable Image, Virtual Private Address Space, Threads, Private Handle Table, and Primary Access Token. So you can think of a process as these 5 things together forming a discrete unit that is "more than the sum of its parts".

A good way to define each and make sense of it is to depart from the moment we run an application. So imagine here we open command prompt and write `notepad.exe`, thus instructing our system to launch the application (a human-term) so we can use it.


## Loading the Executable Image

The very first thing that happens when we launch `notepad.exe` is our system takes the code from our disk, and loads it into our system's memory. Since we now want to run this application we want to use memory since it is of course much faster than our hard drive.

So Windows takes the code - which consists of "instructions" + "initial data" - and loads it into memory. This is now known as the "executable image".


## Private Virtual Address Space (VAS)

But the executable image does not just float around in some arbitrary region in our system's memory. Instead, our system "carves out" (not physically, but in an abstract, or "virtual" manner) a region of continuous memory. Our process is now "given" this region of memory, which is larger than the executable image because of course while its running it might want to save tmp files, or use specific functions, that will require additional memory.

There an interesting "illusion" that takes place here, almost a Truman Show-esque trick that's played on the process. You see,  the process is given some memory and its neither aware of any other processes, nor any memory beyond that which it receives. So from its POV, it's the only process, and all the memory there is belongs to it. This is why its referred to as being **private**.

Now there are two main benefits to this - it ensures processes don't compete with one another to write to specific memory locations, meaning it's inherently thread-safe. More importantly, it also serves as a security barrier. One process can't just peek into another's memory (at least, not without asking the Kernel, or finding a vulnerability).


## Threads (The Workers)
One of my favourite teachers, Pavel Yosifovich has a saying: "A process does not run, a process manages". But of course, an application has code that has to run - it has to be executed by the CPU. So what allows this, what allows a process to run? **Threads** do. You can think of a thread as something that allows the code to go from the **VAS** to the **CPU** to be **executed**.

So every process has to have at least one thread in order to be able to execute. In fact if the OS detects a "threadless process", it will clean it up. It's also worth nothing that a process can, and often does, have multiple threads - dozens, or even hundreds. But at any given moment in time only a few (at least one) will actually be executing. So a process is constantly managing the state of its threads, suspending some, resuming others etc.

## Private Handle Table
Whenever a process opens a file, registry key, or another process, it gets a "handle", which you can just think of as a reference to these kernel objects. Each process keeps something called a private handle table, just like you might have an address book "translating" between people's names and their telephone numbers, so the private handle table maps these kernel objects to their handles (references).

So when a process wants to access another Kernel object, it can look for the reference in its private handle table. If it's there, it will use it, and if not it will request it from the Kernel. If it's denied, well then that's that. But if its given the handle then it will presumably use it, but also save it in the handle table so that the next time it can be used immediately without bothering the Kernel again.

## Primary Access Token
As mentioned above, threads are what allow code to run. Threads "do things". But one thing to note is that a process won't police itself - it won't say: "I really want to do this one thing, but I know I'm not allowed to, so I won't even try."

No, a process will try to do anything, but that does not mean it will be allowed to do anything. So what enforces this? It is of course the Kernel - when a thread attempts execution, the Kernel will first determine whether it has the permission to do the things it's trying to do. How does it determine it? By looking at it's **Primary Access Token**.

This is the process's ID badge. It dictates the user account, groups, and privileges (like "Debug Programs" or "Load Driver"). In short, it identifies to the Kernel what rights our process has.

We'll into this much deeper later, but it's worth nothing that a process is not stuck with its default per se. It might for example be able to steal a high-privilege token from another process (e.g., a SYSTEM process) and use it to escalate privileges.


## Video
If you'd like to review a video version of this lesson, which also goes into a bit more depth, see [this video](https://www.youtube.com/watch?v=LAnWQFQmgvI).






---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "01_intro.md" >}})
[|NEXT|]({{< ref "03_components.md" >}})