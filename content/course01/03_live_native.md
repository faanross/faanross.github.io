---
title: "Section 3: Live Analysis - Native Windows Tools"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---


`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 2](https://www.faanross.com/course01/02_attack/) `|` [Proceed to Section 4](https://www.faanross.com/course01/04_live_hacker/) `|`

***

&nbsp;  

{{< figure src="/img/gif/office_sp.gif" title="" class="custom-figure" >}}

# 3. Live Analysis: Native Windows Tools
# 3.1. Introduction
So our first analysis will be a quick review using standard (native) Windows tools. These tools are a quick way to get a finger on the pulse - they'll give us a broad overview of some important indicators while at the same time being limited in the depth of information.

So if we have at our disposal better tools, ie tools that can provide more information, why bother? I'm of the belief (inspired by one the greats, [John Strand](https://twitter.com/strandjs)), that even if there are better tools available you should *also* be able to do a basic analysis with the native Windows tools. 

Tools may change, they come and go, or, you might land in a situation where they are, for whatever reason, unavailable. Knowing how to get a basic read with Windows tools in any situation covers your bases. Think of it as learning how to survive in the outdoors - yes you can always make a fire using a lighter, but there's a good reason to also learn how to make it, however cumbersome, with what's always available - it might just save your butt in case your lighter fails. 

{{< figure src="/img/survivorman2.gif" title="" class="custom-figure" >}}

***

&nbsp;  

# 3.2. Theory
You will benefit from understanding the [following short theoretical framework on the '3 Modes of Threat Hunting'](https://www.faanross.com/posts/three_modes/). I leave the decision of whether or not to read it up to you, though it will be referenced throughout the remainder of the course. 

***

&nbsp;  

# 3.3. Analysis
There are a number of things we can look at when we do a live analyis using the native tools, including: connections, processes, shares, firewall settings, services, accounts, groups, registry keys, scheduled tasks etc.

For this course we will only focus on connections and processes. If you are keen to learn more about how to investigate the other factors I suggest you watch [this excellent talk by John Strand](https://www.youtube.com/watch?v=fEip9gl2MTA). 

{{< figure src="/img/speech.gif" title="" class="custom-figure" >}}

# 3.3.1. Connections
Let's run `netstat`, which will display active network connections and listening ports. After all, most malware serves merely as a way for the adversary to ultimately have a connection to the victim's machine to run commands and exfiltrate data.

So open a PowerShell admin terminal on our Windows 10 system and run the following command:
```
netstat -naob
```
Note in particular the inclusion of `o` and `b` in our command which will also show the PID, as well as name of executable, involved in each connection.

In the results we can immediately see a variety of connections, as well as ports our system is listening on. Let's especially pay attention to `ESTABLISHED` connections.

We scroll through the list and then as threat hunters something unusual should stick out to us:

{{< figure src="/img/image071.png" title="" class="custom-figure" >}}

What exactly is unusual about this? Well even though `rundll32.exe` is a completely legitimate Windows process, it's used to load DLLs. The question then beckons: why exactly is it involved in an outbound connection?

In this case we can see it's connected to another system on our local network, but remember that's only because of our VLAN setup. In an actual attack scenario this would not be the case, meaning we see `rundll32.exe`, a process not known to be involved in creating network connections, being responsible for establishing a connection to a system outside of our network. 

In a typical scenario we'd immediately want to know more about this IP. Is it known? Is there a business use case associated with it? Are other systems on the network also connecting to it? Because if the answer to all those questions are no - well then we definitely have something weird on our hands.

{{< figure src="/img/weirdal.gif" title="" class="custom-figure" >}}


So let's use our native Windows tools to learn more about this process. To do so  let's just note of our PID, as can be seen in the image above mine is `3948`, yours will be different. 

# 3.3.2. Processes

Let's learn more about this strange process, specifically: what command-line options were used to run it, what is it's parent process, and what DLLs are being used by the process.

**Let's have a look at the DLLs, staying in our PowerShell terminal we run:**
```
tasklist /m /fi "pid eq 3948"
```
{{< figure src="/img/image072.png" title="" class="custom-figure" >}}

On quick glance nothing seems unusual about this output - no DLL sticks out as being out of placed for `rundll32.exe`. So for now let's move on with the knowledge that we can always circle back and dig deeper if need be. 

**Next let's have a look at the Parent Process ID (PPID):**
```
wmic process where processid=3948 get parentprocessid
```
{{< figure src="/img/image073.png" title="" class="custom-figure" >}}

Great, we can see the PPID is `6944`, now let's figure out the name of the process it belongs to:
```
wmic process where processid=6944 get Name
```
{{< figure src="/img/image074.png" title="" class="custom-figure" >}}

We see thus that the name of the Parent Process, that is the name of the process that spawned `rundll32.exe` is `rufus.exe` - a program used to create bootable thumb drives. 

On quick glance this too seems unusual - why is this app needing to call `rundll32.exe`? However, since we're not an expert on this program's design, this could potentially be part of its normal operation - we'd have to jump in deeper to understand that.

{{< figure src="/img/sus2.gif" title="" class="custom-figure" >}}

Let's keep the bigger picture in mind again - we came upon `rundll32.exe` because it created a network connection to an external IP. So in that sense, yes this is very weird - why is a program used to create bootable thumb drives spawning `rundll32.exe` which then creates a network connection? 

One final thing here using our native tools, let's have a look at the command-line arguments:
```
wmic process where processid=3948 get commandline
```
{{< figure src="/img/image075.png" title="" class="custom-figure" >}}

We can see the command is nude - no arguments are provided. Well, since again the `rundll32.exe` command is typically used to execute a function in a specific DLL file, you would expect to see it accompanied by arguments specifying the DLL file and function it's supposed to execute. But here it's simply executed by itself, again reinforcing our suspicion that something is amiss. 

***

&nbsp;  

# 3.4. Closing Thoughts
So we started with an open mind, spotted an unusual process being involved in a network connection, and then using other native Windows tools learned more about this process. And the more we learned, the more our suspicion was confirmed:
- The parent-child relationship is unusual.
- The need for the parent relationship to ultimately create a network connection is unusual.
- The fact that the process was ran without command-line arguments was unusual. 

Now that our suspicion is well and truly aroused let's dig in deeper to build our case using `Process Hacker`.


&nbsp;  

***

`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 2](https://www.faanross.com/course01/02_attack/) `|` [Proceed to Section 4](https://www.faanross.com/course01/04_live_hacker/) 