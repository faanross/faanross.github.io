---
title: "Section 7: Post-Mortem Forensics - Traffic Analysis"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---


`|` [Return to Course Overview](https://www.faanross.com/posts/course01/) `|` [Proceed to Section 8](https://www.faanross.com/course01/08_review/) `|`

***

&nbsp;  

# 7.1. Introduction

In many respects, the realm of network packets is the ultimate domain for threat hunting. It is the only place where malware cannot hide, especially if it intends to communicate. Thus, even if malicious traffic is buried under an avalanche of legitimate traffic, one thing is for sure: the malware's communication is always present, somewhere.

Traffic analysis is an absolutely integral part of threat hunting, playing a major role in nearly every aspect—whether you are searching for initial evidence or seeking to build a case. Accessing packets directly using tools like WireShark/Tshark, or employing specialized software such as Zeek/RITA, provides incredible opportunities for threat hunters.

In this course, however, we are only going to touch on it lightly. The reason for this approach is straightforward: we have simulated a very specific phase of being compromised. We emulated a stager reaching out to establish a C2 connection, and even though we briefly touched on some other actions, we severed the connection shortly after it was created.

In other words, we actually performed the initial exploitation (i.e., creating the connection), but we largely skipped the 'post-exploitation' phase. Beyond all the details, the major difference between these two phases often relates to duration: while initial exploitation is typically brief, post-exploitation can last weeks, months, or even years.

So here's the thing: traffic analysis is fundamentally about discerning patterns. But meaningful patterns generally emerge over time. For example, let’s say a C2 beacon reaches back to the C2 server once an hour. If you only had a one-hour packet capture, you would expect to see only a single callback, which is obviously not a pattern. Conversely, if you had a one-week packet capture, you could expect to see close to 150 callback packets, likely forming a discernible trend in terms of packet size and duration between sends.

All this to say: although traffic analysis is incredibly important for threat hunting, due to the specific nature of the attack we emulated here, it isn't an ideal match in this context. Nonetheless, I wanted to introduce it in a rudimentary sense in this course so that you have some exposure to what can be expected regarding an initial exploitation, even if it's minimal. Rest assured that in a future course, we will delve much deeper into traffic analysis, particularly to help identify unwanted persistent connections.

# 7.2. Analysis

**So let's have a quick look at what's going on in the packet capture.** Open your Ubuntu VM, open WireShark, and then open the packet capture we transferred over in Section `5.1`. 

{{< figure src="/img/image097.png" title="" class="custom-figure" >}}

We can see that in the brief amount of time we ran the capture for a total of 584 packets were captured. In case you are completely new to this: we can expect *a lot* of these to be completly unrelated to our attack. Even if you are not even interacting with your system it typically generates a lot of packets via ordinary backend operations.

So, our next step would now be to find which packets are related to the emulated attack. 

Scrolling down, in my capture we can see around packet 58 there is a DNS request for `raw.githubusercontent.com`.

{{< figure src="/img/image098.png" title="" class="custom-figure" >}}


first things of  interest seem to be 58 +59 - DNS query for the web server
we can look into second one and we can see that the ip for the URL was 185.199.108.133

then we see a whole series of convos between our IP and that IP, making connection, checking certs (TLS) etc

then 116 we can see ARP asking for IP of attacker, clearly now scipt has been injected and malware seeking to make conneciton back 

117 we can see response

then from 118 on we can see long convo between the two - victim and attacker 

let's follow convo see what's intersrting

- immediately what do we see? PE header - magic bytes + DOS stub
- then about 1/3 of way in we see what looks like a series of runtime errors


https://www.first.org/resources/papers/conference2010/cummings-slides.pdf

we see some strings, google it above 
can we save it and search it with YARA rule?? 

no, no positive hits with YARA

for now, let's abandon this since sidetrack

we can see it in course, find interesting, but say outside of scopt




we create a new folder, this is where output will go
we navigate to folder
we run the command
[full path to zeek] -r [full path to pcap]

analyst@analyst:~/Desktop/zeeklogs$ /opt/zeek/bin/zeek -r ~/Desktop/new_capture.pcapng 

when we do this it generates 6 logs


















+++++++++++++++++++++++++++++++++++




First, build a case mode UII


Second, Pareto Principle logging.












- mention this one usualyl more realm of SOC/SIEM and not Forensics, which usually more focus of threat hutning.
- Likely one of the thoughts underpinning this attitude is that logs are grunt-work, mountains of nothing that needs to be sifted through, mountains so huge its completly beyond the scope of humams, and so SIEMs not only CAN do it, but are better than humans in it. 

- but that is true for the general appraco to logs. But what we are speaking of here is a much specific way of looking at logs - limitiing the type of logs we look at. Additioarnlly, logs depending on the PHASE. See below - logs might not be a great place to start in Phase 1, but for example can be perfect for Phase 2. Sicne you already more or less know what you are looking for, makes the volume manageable, esp considering we're likely only interested in Sysmon, Powershell, and a highly select WEL IDs. 


ULTIMATELY, Phase 1 should focus on where they cannot hide - meaning memory and packets. Every where else, disk, logs, etc they can hide. but they can never hide from memory/packets = memory when they are at rest/use, packets when they are in transit. 


Ok so now as we go ahead, remember we have no idea of the attack, we don't know a DLL injected attack has happened, since that was "evil ash" doing it. Meaning we are in Phase 1, and then thios I might not mention but for own sanity - at end of live analysis II (PE Hacker), we can then switch over to Phase II.




Will be using same VM here, victim, in practice we would bnever do this,.


&nbsp;  

***

`|` [Return to Course Overview](https://www.faanross.com/posts/course01/) `|` [Proceed to Section 8](https://www.faanross.com/course01/08_review/) `|`