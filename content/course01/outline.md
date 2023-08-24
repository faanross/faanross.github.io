---
title: "Outline: Threat Hunting for Beginners: Hunting Standard Dll-Injected C2 Implants"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---

[Return to course page](https://www.faanross.com/posts/course01/)

| # | ***Topic*** |
|----------|----------|
| 0 | `Pre-Course Banter` | 
| 1 | `Setting Up Your Virtual Environment` | 
| 1.1 | Introduction |
| 1.2 | Requirements |
| 1.3 | Hosted Hypervisor |
| 1.4 | VM Images |
| 1.5 | VM 1: Windows 10 aka "The Victim" |
| 1.5.1 | Installation |
| 1.5.2 | VMWare Tools |
| 1.5.3 | Deep disable MS Defender + Updates |
| 1.5.4 | Sysmon |
| 1.5.5 | PowerShell ScriptBlock Logging |
| 1.5.6 | Install Software |
| 1.5.7 | Creating a Template |
| 1.6 | VM 2: Kali Linux aka "The Attacker" |
| 1.7 | VM 3: Ubuntu Linux 20.04 aka "The Analyst" |
| 1.7.1 | Installation |
| 1.7.2 | Install Software |
| 1.7.2.1 | Volatility3 |
| 1.7.2.2 | WireShark |
| 2. | `Performing the Attack` | 
| 2.1 | Introduction |
| 2.2 | Theory |
| 2.2.1 | What is DLL? |
| 2.2.2 | What is a DLL-Injection Attack? |
| 2.2.3 | What is a Command and Control (C2) Stager, Server, and Payload? |
| 2.2.4 | Further Reading |
| 2.3 | ATTACK! |
| 2.3.1 | Getting IPs |
| 2.3.2 | Generate + Transfer Stager |
| 2.3.3 | Hit The Record Button |
| 2.3.4 | Preparing Our Injection Script |
| 2.3.5 | Injecting Our Malicious DLL |
| 2.3.6 | Artifact Consolidation |
| 2.4 | Shenanigans! A (honest) review of our attack |
| 3. | `Live Analysis: Native Windows Tools` |
| 3.1 | Introduction |
| 3.2 | Theory |
| 3.3 | Analysis |
| 3.3.1 | Connections |
| 3.3.2 | Processes |
| 3.4 | Final Thoughts |
| 4. | `Live Analysis: Process Hacker` |
| 4.1 | Introduction |
| 4.2 | Theory |
| 4.3 | Analysis |
| 4.4 | Final Thoughts |
| 5 | `Post-Mortem Forensics: Memory` |
| 5.1 | Transferring the Artifacts |
| 5.2 | Introduction to Volatility |
| 5.3 | Analysis |
| 5.3.1 | pslist, pstree, and psinfo |
| 5.3.2 | handles |
| 5.3.3 | cmdline |
| 5.3.4 | netscan |
| 5.3.5 | malfind |
| 5.4 | Final Thoughts |
| 6 | `Post-Mortem Forensics: Log Analysis` |
| 6.1 | Introduction |
| 6.2 | A Quick Note |
| 6.3 | Sysmon |
| 6.3.1 | Theory |
| 6.3.2 | Analysis |
| 6.4 | PowerShell ScriptBlock |
| 6.4.1 | Analysis |
| 6.5 | Final Thoughts |
| 7 | `Post-Mortem Forensics: Traffic Analysis` |
| 7.1 | Introduction |
| 7.2 | Analysis |
| 8 | `Course Review` |



[Return to course page](https://www.faanross.com/posts/course01/)