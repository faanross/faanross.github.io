---
title: "Section 7: Post-Mortem Forensics - Traffic Analysis"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---


`|` [Return to Course Overview](https://www.faanross.com/posts/course01/) `|` [Proceed to Section 8](https://www.faanross.com/course01/08_review/) `|`

***

&nbsp;  





# 7. Post-Mortem Forensics: Traffic Analysis
# 7.1. Introduction


Hunting in the realm of traffic can be extremely challenging given the scale, hwoever it's also the one domain where, however difficult, the answer is always sure to be found.

Why? Well as once again Chris Benson likes to say: it's the one place malware cannot hide. Malware can find incredivbly sophistiaceted ways to obscure its presence in memory, it can find creative ways to avoid/delete logging, but it has to generate packets. And if it does not generate packets, it means it is not communicating. 


 is in some respect 

... truth - 

As Chris Benson like 



talk about why abbrevuiated





# TRAFFIC ANALYSIS
# Introduction

traffic analyssius one of most powerful ways to do threat hunting
but like every tool has strenghts and weaknesses

our specific investaition here, analyzing ane vent that was basicalkly only the initiiaal foothold, its weakness. 


For Traffic Analysis cIntroductyion
LIMITATION of traffic in this scenarion
- mention here that it's strength not really as much as others in deteceting intiail actions. Traffic is not great for finding individual actions, it's great for finding emergent patterns (time, session size etc), usually the longer period the better.

Here we only simulated an initial comprmoise, we did not really maintain a long perdio (1 day +) etc, communciating with server, sharing data etc. So


- first do the Threat Hunting Level 1 course
- then do the Chris Benton traffic analysis



- Let's first rerun attack (remember to drop cmd etc)
- redo pcap with just that
- then let's do threat hunting level 1, other vid courses teaching about c2 in traffic logs etc.



# FOR NOW REDO PCAP SO CLEANER


Ok our bew pcap has 584
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