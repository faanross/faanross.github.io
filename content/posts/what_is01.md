---
title: "what is threat hunting part I - different strokes for different folks"
date: 2023-10-03T02:01:58+05:30
description: "this is the first in a series of episodes exploring the what + why of threat hunting using a conceptual framework."
tags: [threat_hunting, theory]
author: "faan|ross"
type: post
---

*** 

{{< youtube RwgS4l4Mx7Y >}}

# preface
Today I want to explore one of several frameworks I like to employ to help me understand exactly what threat hunting is at a high level. I have a few more related frameworks I'd like to explore in the future, so I guess we can go ahead and call this `what is threat hunting part i - different strokes for different folks`. 

Let's begin by exploring what I like to think of as the central problem of organizational cyber defense. 

# the central problem of organizational cyber defense

I want to be clear that I'm not referring to the central goal of organizational cyber defense, which is something more along the lines of finding and eliminating threats to maintain the informational integrity of the network. Rather, the central problem is something that must be solved to achieve our goal. 

But even before we get to that, I want to get perhaps a wee bit philosophical for just a second and I want us to consider what the act of security is. Well, to me the act of security is, at its most elementary level, a **relationship** between two elements: **something to secure** and **someone that secures**.

In the context of organizational cyber defense, our **something to secure** is, of course, the network. But it's not really the physical network, that's obviously more the domain of the physical security guard. Rather, it's more along the lines of the informational integrity of the network.

The **someone that secures** that something in this case is our security operator. At this point, I don't want us to become entrenched in specific paradigms or strategies regarding organizational cyber defense. For now, let's keep things generic and high-level, so we can simply think of this person as a security operator, nothing more.  

But there's something missing...

For the security operator to secure the network, they need a **view** of it. Now, to be clear, by **view** I do include the actual physical (visual) view, but it's not limited to that. Rather, you can think of the **view** as the sum of all abilities to obtain informational feedback in a periodic manner so as to be able to monitor the status of the network and ultimately detect any deviation indicative of a compromise.

{{< figure src="/img/gif/eye.gif" title="" class="no-border" >}}

In the case of a physical security guard, for example, they might be sitting in a security room where they have live feeds to different cameras, but they might also have a motion sensor, contact sensors, seismic sensors, etc. Taken together, all these currents of information can be considered the **view** in this context. 

But there's one more thing in our model that's not quite right, and that is, of course, what the **view** is of. As I said before, the security operator is not sitting there looking at the physical computers, routers, cables etc to see if a compromise has taken place. Rather, the security operator monitors the data produced by the network.

But it's also not all the data - it's not the user-generated or application data, it's not the presentations, invoices, emails etc. Rather, it's what we can term the security data, which is ultimately a form of meta-data - it's data about the network. It's everything from packets captures to the whole variety of logs produced on the host as well as network. And, if we were to include threat hunting, we can even make a reasonable argument that it should include memory dumps as well.

*But there's an issue...*

If the security operator had to try to process all security data generated by the network in a meaningful manner, they would be completely and utterly overwhelmed. And this is because at any given moment, much more data is being generated than can be processed.

{{< figure src="/img/gif/sweep.gif" title="" class="custom-figure" >}}

# data capacity incompatibility

This leads us to what I propose as the central problems of organizational cyber defense. Another way to frame this is as a `data capacity incompatibility` problem. What I mean to say here in plain English is that the security operator cannot process all security data in a meaningful manner to effectively find threats.

So then one useful frame we can use to think about what an organizational cyber defense strategy is, is how this problem is solved. In other words, how do you go about making an inaccessible amount of data accessible so that the security operator can effectively find threats.

So then, with this framework in mind let's explore how this problem is solved with today's conventional approach.

# The SOC-SIEM paradigm

Note that I'm going to be calling this approach the SOC-SIEM paradigm, knowing full well that modern organizational cyber defense strategies, of course, contain many more layers and complexity than this. But I think for the moment one can make a reasonable argument that it's these two elements that lie at the logical core of most organizational approaches to cyber defense. 

So we have our **something to be secured** - that is the informational integrity of the network. We have our **someone that secures** - in this case, the SOC analyst. And then, of course, the SOC analyst has their **view** - meaning here, at this point in time, all the security data being produced by the network.

But now the SOC analyst is faced with the problem of `data capacity incompatibility`. So then, how is this problem solved in the SOC-SIEM paradigm? Well, it's solved by an external filter. And this filter, when you get down to it, is a piece of code. Perhaps you prefer to think of it as an algorithm or a piece of software, but really the key point is that it is an artifact - something that exists outside of the SOC analyst themselves. 

So this external filter will receive **all** the security data being produced by the network, and then it will decide, based on its own rule set, which tiny subset of the data to present to the SOC analyst. And of course we call this tiny subset that's presented to the SOC analyst `alerts`.

So, what a SOC analyst is really doing in this scenario is responding to alerts.

{{< figure src="/img/whatis01/001.png" title="" class="custom-figure-6" >}}

One final thing I think is really worth paying attention to is that in the SOC-SIEM paradigm the view is reduced. Before, we thought of the view as the total field of security data under scrutiny by the security operator, which by default is all security data being produced by the network. 

But in the SOC-SIEM paradigm this no longer holds true. Though, yes, of course, the SOC analyst still retains the ability to consult all security data (via, for example, SIEM searches or other indexing/parsing methods), they no longer personally consider all security data when looking for threats. Thus the "view" has been reduced to subset determined not by the SOC analyst, but by the code. 

I want to mention this because it is this reduction in view which leads to the introduction of major blind spots (pink shaded areas in image below) into this approach to cybersecurity. And it's really these blind spots that present much of the potential for threat actors to move about undetected. 

{{< figure src="/img/whatis01/002.png" title="" class="custom-figure-6" >}}

Ok, so that's the SOC-SIEM approach; now let's have a look at this exact same scenario, but how one might go about approaching it using threat hunting.

# The Threat Hunting approach

So our setup is the exact same. First, we have our network to be secured. Then, we have our security analyst - in this case, of course, it's a threat hunter. Again, she has a view of the network to be able to monitor its status via the security data. And of course she is dealing with the exact same problem: the problem of `data capacity incompatibility`.

So how does she solve this problem? Well, as was the case with the SOC analyst, the threat hunter uses a filter. Crucially, however, the threat hunter does not use an outside-in filter, but an inside-out filter, that is to say the threat hunter applies an internal filter to reduce the data set. 

I'm going to term this filter a **kills-based** filter, really just because I could not find a single label to encompass everything I mean. All this to say, a threat hunter filters the total subset of data down in any given threat hunt based on their skills, experience, knowledge, available tool set, methodologies, resources, and perhaps even what they might be hunting for in any given moment based on threat intel.

The key takeaway here is that the decision on how to reduce the data so it is accessible to find a threat is done by the security operator themselves, not by an external entity. And because it is dynamic, that is contingent in any one instance upon all the variables mentioned above, the exact subset under investigation can change and assume many different forms. 

So whereas the SOC analyst was inherently reactive, that is reacting to alerts produced by the code-mediated filter, the threat hunter is inherently proactive, that is, they go to seek out the threat *in situ*, guided by an internal locus. 

{{< figure src="/img/whatis01/003.png" title="" class="custom-figure-6" >}}

I also want to point out that, *yes*, while in any specific instance of threat hunting, the view has been reduced to the subset of data under scrutiny, the total view is retained. So while the threat hunter might not be investigating all the security data at any given moment, they retain the ability to investigate any subset of the total security data in any given instance. In a sense, the "field of potential" has been retained - there is no inherent reduction of the total view as a result of employing threat hunting.

Let's quickly compare the SOC-SIEM and threat hunting approaches.

# SOC-SIEM paradigm vs Threat Hunting

{{< figure src="/img/whatis01/004.png" title="" class="custom-figure-6" >}}

With the SOC-SIEM approach the total set of security data is code-mediated, or externally-filtered. It is an inherently reactive approach - that is, the SOC analyst is reacting to the output of the code-mediated filter. And, with the SOC-SIEM approach, the total data set is inherently reduced to a tiny subset, thereby introducing blind spots.

In comparison, the threat hunting approach is skill-mediated, or internally filtered. It is a proactive approach - the first step forward to find the threat *in situ* is a product of the threat hunter's decision. Finally, the full scope of data visibility is retained, meaning that all security data could potentially be investigated. 

# Threat Hunting >>> SOC-SIEM paradigm?

Ok so then based on the way I've presented it here, is it just really the case that threat hunting is a new and improved version of the SOC-SIEM paradigm? Should we just replace all SOC analysts with threat hunters?

{{< figure src="/img/gif/howaboutno.gif" title="" class="custom-figure-6" >}}

Well of course not.

Rather, it's as the wise Jimi Hendrix said:

{{< figure src="/img/gif/jimi.gif" title="" class="custom-figure-6" >}}

Applied contextually here, what I mean is that the SOC-SIEM and threat hunting paradigms are two different solutions for two different problems. But to understand what I mean by this we have to introduce one additional layer of nuance, specifically as it related to the quality of threats.

# low- vs high-value attacks

So far we've implicitly assumed that all threats are the exact same, as if there's this homogeneous slop of threats out there, all of equal value. But of course we know this is not the case.

There are of course many ways in which we can categorize threats, one common way for example is according to a quality cline. Many of us have encountered such a categorization based on a quality cline (as shown below) when we studied introductory cyber security.

{{< figure src="/img/whatis01/005.png" title="" class="custom-figure-6" >}}

Now I want to do something similar, that is also categorize threats based on a quality cline, but something even simpler. Let's simply bifurcate all threats into two categories - which I'm going to call `low value` and `high value` attacks.

{{< figure src="/img/whatis01/006.png" title="" class="custom-figure-6" >}}

When it comes to `low value` attacks:
- Make up the **vast majority** of attacks, or at least attempted attacks.
- Are largely **automated**, meaning that beyond the initial push given by a human operator they don't require human input.
- Because of this, that is because of the ease of replication and low human input, they are **cheap**.
- Finally, we can also say that they operate according to a "shotgun", or "spray and pray" paradigm. Every individual attack has a very small likelihood of succeeding, however since the total amount of attacks are cheap to scale it ultimately still leads to many successful attacks.

`High value` attacks on the other hand:
- Are typically targeted, that is it usually involves a specific group of attackers targeting a specific network.
- Requires a high level of human input, and these individuals are also typically skilled and competent individuals with specialized knowledge.
- Because of this these attacks are expensive.
- As a consequence there are relatively fewer of them compared to low value attacks, however the probability of a high value attack succeeding is much higher.


So then how does this connect to how different problems are optimally solved by the SOC-SIEM versus threat hunting approaches? Well, whereas the SOC-SIEM paradigm is the right tool to address low value attacks, threat hunting is the right tool to address high value attacks.

That's because the SOC-SIEM paradigm is a mostly automated solution for a mostly automated problem, whereas threat hunting is a mostly human-driven solution for a mostly human-driven problem.

So then which approach is best? The answer is, of course, that it depends on each organization and the threats they are facing. **The best approach is figuring out the right proportion of each of these components to optimally address the type of threats you are predominantly dealing with**.

***
















