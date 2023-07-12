---
title: "Threat Hunting Dll-injected C2 beacons"
date: 2023-07-12T02:01:58+05:30
description: "In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 fundamental approaches: memory forensics, log analysis + UEBA, and traffic analysis."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan ross"
---

*** 

# Introduction

In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 approaches: memory forensics, log analysis + UEBA, and traffic analysis. The entire course is practically-oriented, meaning that we'll learn by doing. I'll sprinkle in a tiny bit of theory just so we are on the same page re: C2 frameworks and DLL-injection attacks; and in case you wanted to dig in deeper I provide extensive references throughout this document. 

So here's a brief overview of what we'll be getting upto...
- In PART 1 we're going to set up the virtualized environment,
- we'll create a windows 10 VM which will server as our victim,
- we'll also set up a kali linux box which will be our attacker, 
- as well as an ubuntu box which we'll use to run some post-mortem analysis on.

.
- In PART 2 we'll run the actual attack ourselves,
- for the classical dll-injection we'll use metasploit to generate both the stager and meterpreter handler,
- once we've transferred the stager to the victim we'll run it from memory using powersploit,
- for the reflective dll-injection we'll perform the entire process using metasploit.

.
- In PART 3 we'll cover memory forensics,
- first we'll do a basic live read using Process Hacker,
- we'll then dump the memory with winpmem,
- finally we'll have a look at the it with Volatility.

.
- IN PART 4 we'll get into some logs,
- along with standard Windows Event Logs, we'll also use other (cough, far superior, cough), logs we setup in the first part: namely sysmon and powershell logging,
- we'll briefly jump into the raw logs just to look at some very high-level indicators and then,
- we'll process them using the awesome UEBA framework DeepBlueCLIv3.

.
- IN PART 5 we'll look at traffic analysis,
- we'll run our PCAPS through Zeek,
- and get some insights from the threat hunting framework RITA.

In the end we'll recap and formulate some key takeaways to serve you on your journey as you venture forth into the world and become a bada$$ hunter.

But first, *le sigh*, it's required we just dip our toes into a wee bit of theory. But I promise once we're done here - 10 to 15 mins tops - it'll be applied learning until the end of our journey. 

Sounds good? Let's get it.


{{< figure src="/img/randy01.gif" title="" class="custom-figure" >}}



***

# Theory
# what is a DLL?
Succinctly as possible, a DLL is a communal library containing code. They are not a program or an executable in and of themselves, but they are in essence a collection of functions and data that can be used by other programs. 

So think of a DLL as a communal resource: let's say you have 4 programs running and they all want to use a common function - let's say for the sake of simplicity the ability to minimize the gui window. Now instead of each of those programs having their own personal copy of the function that allows that, they'll instead access a DLL that contains the function to minimize gui windows instead. So when you click on the minimize icon and that program needs the code to know how to behave, it does not get instructions from its own program code, rather it pulls it from the appropriate DLL with some help from the Windows API. 

Thus any program you run will constantly call on different DLLs to get access to a wide-variety of common (and often critical) functions and data.

# what is a classical DLL-injection?
So keeping what I just mentioned in mind - that any running program is accessing a variety of code from various DLLs at any time - what then is a DLL-injection attack? Well in a normal environment we have legit programs accessing code from legit DLLs. 

With a DLL-injection attack we enter into the population of legit DLLs a malicious one, that is a DLL that contains the code the attacker wants executed. The attacker then injects it into the memory space of a legitimate process. Using a Windows API function (commonly LoadLibrary or CreateRemoteThread), the attacker manipulates the legitimate process into loading and executing the malicious DLL. This effectively allows the malicious code within the DLL to run, often with the same permissions as the hijacked process.

Threat actors love DLL-injection attacks because since they are executed within the context of a legitimate process they run with the same privileges as that of the process (ie potentially elevated), but even more so it makes them much harder to detect. No longer can we look on the process-level for malware, instead we have to peer beneath them at a arguably convoluted level of abstraction. 

Even though classical DLL-injection attacks are less noisy for this exact reason, they still have a design flaw that makes our lives as threat hunters easier - they leave their fingerprints all over the disc. When the malicious DLL is initially transferred to the victim's system, it's written to disc, allowing us a potential breadcrumb for discovery. 

And thus the inevitable next iteration in this branch of digital evolution is...

# what is a reflective DLL-injection?
At a *high-level*  classical and reflective DLLs are identical save for one difference: whereas the former is written to disc then injected into memory space, the latter is injected into memory space directly. This makes them conventionally even harder to catch since we can't rely on any disc forensics to reveal its presence. However, as we'll learn in this course, in another way it makes it for those who know what to look for perhaps a bit easier. 

How come?

Well, on a pattern-level we can observe that the very fact that a DLL, meaning ANY DLL, is in memory without a disc counterpart is very unusual. Perhaps not immediate incident alert level unusual, but at the very least more than unusual enough to warrant further prodding with piqued interest. 

As a bridge to the closing part of our theory section let's zoom out a bit. Here we have been speaking about a specific mechanism of how malware (that is bad code) gets a victim's system to execute it. There are obviously many other such mechanisms, and equally bviously there are many different types of malware that use specifically DLL-injection attacks as the means to their desired ends (ie getting executed). 

In this specific course however we'll be focussing on a very specific type of malware, actually it would be even more accurate to say we'll focus on a specific component of a specific type of malware... 

# what is a Command and Control (C2) framework, stager, and beacon?

Let's start by sketching a scenario of how many typical attacks play out these days.

{{< figure src="/img/hackers01.gif" title="" class="custom-figure" >}}

An attacker sends a spear-phishing email to an employee at a company. The employee, perhaps tired and not paying full attention, opens the "uregent invoice" attached to the email. Opening this attachment executes a tiny program called a stager.

A stager, though not inherently malicious, "sets the stage" by performing a specific task: it reaches out to a designated address (owned by the hacker) to download another piece of code, then executes it.

The downloaded code establishes the attacker's presence on the victim's system. It acts as a "gateway," allowing the attacker to execute commands on the victim's system from their own.

So the system that the attacker uses to execute these commands is called the Command and Control (C2) server.

The code downloaded by the stager is a type of C2 implant known as a beacon, an approach popularized by Cobalt Strike. Unlike traditional C2 implants that maintain a continuous, persistent network connection (which can raise suspicion), a beacon does not. 

Instead, it periodically "calls home" to the C2 server, asking whether there are any new commands. If there are no commands, the connection is immediately terminated. If there are commands, the beacon retrieves them and then terminates the connection, lying dormant until the next scheduled "check-in". This sporadic communication helps the beacon blend into normal network traffic, making it more difficult to detect.

GREAT, and that's it for the theory, it's time to get going! But in case you are feeling inspired here are a selection of incredible resources that helped me.

{{< youtube borfuQGrB8g >}}

.
{{< youtube lz2ARbZ_5tE >}}

.
{{< youtube ihElrBBJQo8 >}}

*** 

# PART 1: Setting up our virtualized environment
# Overview

For this course I'll be using [VMWare Workstation](https://store-us.vmware.com/workstation_buy_dual) which as of writing costs around $200. However you could also do it with either [VMWare Player](https://www.vmware.com/ca/products/workstation-player.html), or [Oracle Virtualbox](https://www.virtualbox.org/wiki/Downloads), both of which are free. 

Note that some of the details of the setup might be slightly different if you choose to use either of the lastmentioned options and if that occurs then it'll be upto you to figure that out. Google, ChatGPT, StackExchange etc.

One final thing before we get setting up, you'll need the following three iso's (all free of course):
* for the victim we'll use [Windows 10 Enterprise Evaluation](https://info.microsoft.com/ww-landing-windows-10-enterprise.html)
* for the attacker we'll use [Kali Linux](https://www.kali.org/get-kali/#kali-installer-images)
* for post-mortem analysis we'll be using [Ubuntu Linux 20.04 Focal Fossa](https://releases.ubuntu.com/focal/). Just note here the actual edition 20.04 is important since we'll run RITA on it, which, as of writing, runs best on Focal Fossa.

Ok so at this point if you have your hosted hypervisor and all three iso's we are ready to proceed.

# VM 1: Windows 10 aka "The Victim" 

{{< figure src="/img/screamdrew.gif" title="" class="custom-figure" >}}
 
First we'll install the OS using the iso, following that we'll make a bunch of configurations including: 
- deep disable MS Defender
- deep disable Windows updates
- install sysmon
- enable powershell logging
- install Process Hacker
- install winpmem
- install wireshark

.
In VMWare Workstation goto `File` -> New Virtual Machine. Choose `Typical (recommended)`, then click `Next`. Then select `I will install the operating system later` and hit `Next`.

{{< figure src="/img/image001.png" title="" class="custom-figure" >}}

 and that do it.

 ok how about that then do it.
















