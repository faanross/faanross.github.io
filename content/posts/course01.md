---
title: "Threat Hunting for Beginners: Hunting Standard Dll-Injected C2 Implants (Practical Course)"
date: 2023-08-12T02:01:58+05:30
description: "In this beginner-friendly practical course we'll learn how to threat hunt standard DLL-injected C2 implants. We'll set up our own virtual environment, perform the attack, perform our threat hunting analysis, as well as write a report on our findings."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan|ross"
draft: false
type: post
---

*** 
# NOTE THIS IS CURRENTLY STILL WIP, THE REASON IT'S A PUBLIC DRAFT IS LONG AND CONVOLUTED SO JUST TRUST ME. ANYHOO - DO AS YOU WISH. 

***
# Hello friend, so glad you could make it.

{{< figure src="/img/poe.gif" title="" class="custom-figure-3" >}}

`This is the first in an ongoing + always-evolving series on threat hunting.`

<!-- [NOTE: FOR THE VIDEO VERSION OF THIS COURSE CLICK HERE]() -->

The main thing I want you to know about this course is that ***we will learn by doing***. 

`Set up`
We'll start off by creating + configuring our own virtual network, including VMs for the victim, attacker, and analyst. 

`Attack`
Then, instead of using prepackaged data we'll generate data by performing the attack ourselves. We'll use *Metasploit* and *Powersploit* to perform a standard DLL-injection attack. Once we have C2 established we'll simulate a few rudimentary actions such as data exfiltration.

`Live Analysis`
We'll then perform the actual threat hunt. We'll initially perform two rounds of live analysis - first using only Windows native tools to *check the vitals*, and then using *Process Hacker* we'll dig deeper into the memory. 

`Post-mortem Analysis`
In the post-mortem analysis we'll look at the memory dump(*Volatility3*) and perform log analysis (*Sysmon* + *PowerShell ScriptBlock*), before wrapping things up with an abbreviated traffic analysis (*WireShark*). 

`Review`
Finally we'll crystallize all our insights so we can both reinforce what we've learned, as well as learn how to effectively communicate our findings to the greater cybersecurity ecosystem. 

`Theory + References`
I will interject with theory when and where necessary, as well as provide references. If something is unclear I encourage you to take a sojourn in the spirit of returning with an improved understanding of our topic at hand. This is after all a journey that need not be linear - the goal is to learn, and have as much fun as possible. `Act accordingly`. 

{{< figure src="/img/brent.gif" title="" class="custom-figure" >}}

# Course Outline

- [0. Pre-Course Banter](https://www.faanross.com/course01/prebanter/)
- [1. Setting Up Our Virtual Environment](https://www.faanross.com/course01/01_settingup/)
- [2. Performing the Attack](https://www.faanross.com/course01/02_attack/)
- [3. Live Analysis - Native Windows Tools](https://www.faanross.com/course01/03_live_native/)
- [4. Live Analysis - Process Hacker](https://www.faanross.com/course01/04_live_hacker/)
- [5. Post-Mortem Forensics - Memory](https://www.faanross.com/course01/05_post_memory/)
- [6. Post-Mortem Forensics - Log Analysis](https://www.faanross.com/course01/06_post_logs/)
- [7. Post-Mortem Forensics - Traffic Analysis](https://www.faanross.com/course01/07_post_traffic/)
- [8. Course Review](https://www.faanross.com/course01/08_review/)


If you'd like to see a detailed overview of the the entire course [click here](https://www.faanross.com/course01/outline/)

***
















