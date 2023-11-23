---
title: "basic c2 defense evasion by creating a custom dropper (red team series 002)"
date: 2023-11-15T02:01:58+05:30
description: "we create a simple dropper that perform basic defense evasion and remote execution of our msfvenom payload."
tags: [metasploit, red_team, c2]
author: "faan|ross"
type: post
---

*** 

{{< youtube ztUySXIRVKc >}}

# description
this is the second episode in a series where we'll be learning the basics of command and control (C2) by using metasploit. 

in this (optional) video we'll elaborate on what we did in the previous video by creating a custom dropper. 

this custom dropper will:
- detect whether or not ms defender is running
- abort the script if it is
- continue the script if it is not
- it will then download a c2 payload from a web server
- and finally it will execute to establish a c2 connection with meterpreter

though as mentioned this lesson is optional - meaning you can continue the course without doing it - it will increase your understanding of how an more realistic attack will look like on a pattern level. it's also cool af, obvs... so do it!

&nbsp; 
***
















