---
title: "Section 8: Course Review"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---


`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 6](https://www.faanross.com/course01/07_post_traffic/) `|`

***

&nbsp;  

{{< figure src="/img/gif/manhattan.png" title="" class="custom-figure" >}}

# Section 8: Course Review

If you made it this far **CONGRATS**! You have learned a *tremendous* amount - creating our own virtualized setup, emulating a dll-injection attack, and then performing memory, log, and traffic analysis. It's also my sincere hope that you had fun doing this. Threat hunting is an incredibly exciting and I'm glad we got to learn together - may it be the first of many.

One final thing before we high-five and part ways, I though it useful to recap, in its entirety, what exactly happened in terms of our attack and how each stage of our threat hunt contributed to an increased understanding thereof. I'll make this brief, but I do recommend one final review so we can really crystallize our insights.

# Attack

During our attack phase, we simulated a DLL-injection attack, incorporating a few shortcuts for efficiency. I will first outline how we executed the attack and then compare it to how this attack would transpire in a real-world scenario. Recognizing these differences is crucial, as it informs us about what we should and shouldn't expect during an actual threat hunt.

Here’s how we emulated the attack:
- We crafted a malicious DLL on our system.
- We transferred this DLL over to the victim’s system.
- We opened a meterpreter handler on our system.
- On the victim’s system we then downloaded a powershell script from a web server (Github content), and injected it into the victim's memory.
- We opened a legitimate program (rufus.exe).
- We then ran the script we downloaded above, causing the malicious dll to be injected into the memory space of rufus.exe.
- The injected DLL is executed, calling back to the meterpreter handler we created, thereby establishing our backdoor connection.
- We exfiltrated some data using our meterpreter shell.
- We used our meterpreter shell to spawn a command prompt shell.
- We ran a simple command in the new shell (*whoami*).
- We closed the connection.

OK. Now let’s review what an actual attack might have looked like:
- An attacker does some recon/OSINT, discovering info that allows them to craft a very personalized email to a company’s head of sales as part of a spearphishing attack.
- The attacker included in this email a word document labelled “urgent invoice”, and by using some masterful social engineering techniques they convince the head of sales to immediately open the document to pay it.
- Once the head of sales opens the invoice it runs an embedded VBA macro, which contains the adversary’s malicious code.
- This code can then do many, and even all, of the things we did manually:
    - It can download the malicious DLL.
    - It can download and then inject the script responsible for performing the attack into memory.
    - It can also execute the actual script.
- The connection would then be established, allowing the attacker to take further action. 

So the two major differences would be:
1. A lot of actions that would automatically take place once the malicious Word document was opened we performed manually. The same events would (more or less) however take place, thus from a diagnostic point of view many of the same IOCs would stay true.
2. In our simulation we chose a program (rufus.exe) and even opened it ourselves. In an actual attack this highly improbable since it represents unnecessary risk. Rather, the attacker would select a process that is already running to inject into, which could even lead to elevated privileges. So we would not expect to see any IOCs related to this event in an actual threat hunt. 

# Live Analysis: Windows Tools
At this point we, as the threat hunter, are not aware a compromise has taken place. Using a variety of simple Windows native tools we thus take a high-level view of some of the most important indicators. 

Here we discovered:
- An unusual process (rundll32.exe) being involved in a network connection to an unknown external IP.
- This process han an unusual parent-child relationship with rufus.exe.
- The need for the parent relationship to ultimately create a network connection is unusual.
- We would expect rundll32.exe to be run with command-line arguments but it was not, which was unusual. 


# Live Analysis: Process Hacker
Using Process Hacker our suspicion surrounding this process was further reinforced: 
- rundll32.exe itself spawned cmd.exe - very suspicious.
- rundll32.exe was ran from the desktop - unusual.
- The process also had RWX memory space permissions, which is a big red flag.
- We saw that the memory content of the RWX memory space contained a PE file - again, red flag.

# Post-Mortem Forensics - Memory
We did not learn any new information here, however performing this post-mortem analysis allowed us to see how we could derive many of the same conclusions from the two sections above with a memory dump. This would be valuable if, for whatever reason, we could not perform live analysis. Further, even if we did perform the live analysis, it might still be useful to validate the findings on a non-compromised system.

# Post-Mortem Forensics - Sysmon
This was the first of two types of log analysis we performed. At this point we essentially only had three critical pieces of info - the name of the suspicious process (rundll32.exe), the name of the parent process that spawned it (rufus.exe), and the ip address it connected to (ie potentially the ip of the attacker, C2 server). 

Sysmon log analysis then showed us that: 
- The URL, IP, and hostname of the web server the stager reached out to download the injection script.
- The malware manipulated the DisableAntiSpyware registry keys.
- The malware accessed lsass.exe, indicating some credentials were potentially compromised.
- The malware launched raserver.exe with the /offerraupdate flag, creating another potential backdoor.

# Post-Mortem Forensics - Powershell ScriptBlock

We then further learned with Powershell ScriptBlock analysis:
- The actual command that was used by the “stager” to download the script from the web server and inject it into memory.
- Crucially, we learned the actual contents of the dll-injection script.
- The command was actually used to inject the script into rufus.exe, from here we will also learn the id/location of the malicious dll.

Additionally, the logs from both sections provided us with exact timestamps for many major events, which can be very useful in the incident response process.

# Post-Mortem Forensics - Traffic

In our abbreviated traffic analysis we:
- We confirmed the ip + FQDN of the server where initial script was downloaded from.
- We confirmed the ip of C2 server - what our victim system connected to. 
- We also saw the encrypted contents of the conversation between the victim and C2 server which contained some clear text, which we could potentially leverage to learn more about the malware, even potentially its identity. 

# Next Course

That's it friends. Please feel free to reach out to me if you have any questions or comments - I'd love to hear from you.

In the next course (already in the works) we'll create our own little script to simulate beaconing and then leave it running for an extended period (24 hours). We'll then use some of my favorite tools - Zeek, ACHunter, and RITA - to see how we can pick up on this simulated C2 activity.

It's gonna be awesome... 

{{< figure src="/img/gif/obviously.png" title="" class="custom-figure" >}}


&nbsp;  

***

`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 6](https://www.faanross.com/course01/07_post_traffic/) `|`