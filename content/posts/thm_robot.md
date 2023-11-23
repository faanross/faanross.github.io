---
title: "tryhackme CTF walkthrough - mr. robot"
date: 2023-11-19T02:01:58+05:30
description: "a CTF walkthrough for the beginner-friendly THM challenge, 'Mr. Robot'."
tags: [ctf, thm, pentesting]
author: "faan|ross"
type: post
---

*** 

{{< youtube YpJwIPP8lII >}}

# description
In this beginner-friendly CTF walkthrough from TryHackMe we will:
- use nmap to enumerate ports/services
- use gobuster to discover hidden directories and files on the web server
- discover encoded login credentials on a hidden page
- use this to log into a wordpress portal
- use editor privileges to get execute a php rev shell script to get onto the box
- decode credentials to elevate privileges
- discover a SUID bit set on nmap binary
- use a simple 2-step process to get root using the SUID bit

&nbsp; 
***
















