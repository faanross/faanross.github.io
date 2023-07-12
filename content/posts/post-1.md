---
title: "Threat Hunting Dll-injected C2 beacons"
date: 2023-07-12T02:01:58+05:30
description: "In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 fundamental approaches: memory forensics, log analysis + UEBA, and traffic analysis."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan ross"
---

*** 

# Preview 

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

{{< figure src="/img/randy01.gif" title="" >}}







OK and finally before we get going on setting up our virtual environment let's jump into the only theory for this course, namely 
- what are DLLs?
- what is a classical DLL-injection attack?
- what are Command and Control (C2) frameworks?
- What is a C2 stager/beacon?

what exactly a DLL-injection attack is!

If you're already familiar with it and just wanna jump right in head to this time marker. 





























## The First Step: Triage

Whether you’re just starting your day, or you’re in the middle of the chaos and just need to find some sanity … the first step is to get into triage mode.

Triage, as you probably know, is sorting through the chaos to prioritize: what needs to be done now, what needs to be done today, what needs to be done this week, and what can wait? You’re looking at urgency, but also what’s meaningful and important.

Here’s what you might do:

* Pick out the things that need to be done today. Start a Short List for things you’re going to do today. That might be important tasks for big projects, urgent tasks that could result in damage if you don’t act, smaller admin tasks that you really should take care of today, and responding to important messages. I would recommend being ruthless and cutting out as much as you can, having just 5 things on your plate if that’s at all possible. Not everything needs to be done today, and not every email needs to be responded to.
* Push some things to tomorrow and the rest of the week. If you have deadlines that can be pushed back (or renegotiated), do that. Spread the work out over the week, even into next week. What needs to be done tomorrow? What can wait a day or two longer?
* Eliminate what you can. That might mean just not replying to some messages that aren’t that important and don’t really require a reply. It might mean telling some people that you can’t take on this project after all, or that you need to get out of the commitment that you said you’d do. Yes, this is uncomfortable. For now, just put them on a list called, “To Not Do,” and plan to figure out how to get out of them later.

OK, you have some breathing room and a manageable list now! Let’s shrink that down even further and just pick one thing.

## Next: Focus on One Thing

With a lot on your plate, it’s hard to pick one thing to focus on. But that’s exactly what I’m going to ask you to do.

Pick one thing, and give it your focus. Yes, there are a lot of other things you can focus on. Yes, they’re stressing you out and making it hard to focus. But think about it this way: if you allow it all to be in your head all the time, that will always be your mode of being. You’ll always be thinking about everything, stressing out about it all, with a frazzled mind … unless you start shifting.

The shift:

* Pick something to focus on. Look at the triaged list from the first section … if you have 5-6 things on this Short List, you can assess whether there’s any super urgent, time-sensitive things you need to take care of. If there are, pick one of them. If not, pick the most important one — probably the one you have been putting off doing.
* Clear everything else away. Just for a little bit. Close all browser tabs, turn off notifications, close open applications, put your phone away.
* Put that one task before you, and allow yourself to be with it completely. Pour yourself into it. Think of it as a practice, of letting go (of everything else), of focus, of radical simplicity.

When you’re done (or after 15-20 minutes have gone by at least), you can switch to something else. But don’t allow yourself to switch until then.

By closing off all exits, by choosing one thing, by giving yourself completely to that thing … you’re now in a different mode that isn’t so stressful or spread thin. You’ve started a shift that will lead to focus and sanity.

## Third: Schedule Time to Simplify

Remember the To Not Do list above? Schedule some time this week to start reducing your projects, saying no to people, getting out of commitments, crossing stuff off your task list … so that you can have some sanity back.

There are lots of little things that you’ve said “yes” to that you probably shouldn’t have. That’s why you’re overloaded. Protect your more important work, and your time off, and your peace of mind, by saying “no” to things that aren’t as important.

Schedule the time to simplify — you don’t have to do it today, but sometime soon — and you can then not have to worry about the things on your To Not Do list until then.

## Fourth: Practice Mindful Focus

Go through the rest of the day with an attitude of “mindful focus.” That means that you are doing one thing at a time, being as present as you can, switching as little as you can.

Think of it as a settling of the mind. A new mode of being. A mindfulness practice (which means you won’t be perfect at it).

As you practice mindful focus, you’ll learn to practice doing things with an open heart, with curiosity and gratitude, and even joy. Try these one at a time as you get to do each task on your Short List.

You’ll find that you’re not so overloaded, but that each task is just perfect for that moment. And that’s a completely new relationship with the work that you do, and a new relationship with life.
