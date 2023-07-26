---
title: "The 3 Modes of Threat Hunting (Beginner Theory)"
date: 2023-07-26T02:01:58+05:30
description: "A short article covering a single foundational concept related to Threat Hunting."
tags: [threat_hunting, theory]
author: "faan | ross"
---

*** 

# PREFACE

This is a short article covering a single foundational concept related to threat hunting. I'm publishing it here primarily as a stand-alone post because I would like to reference it in some of my courses without needing to reproduce it in every such instance. Nevertheless, I hope, even here by itself, **it might provide some value to you**.

Threat hunting, by its literal definition, is predicated on a single presumption: `we presume a compromise has already occurred, and thus an adversary is established on our network`.

This tenet - *the presumption of compromise* - is a wonderfully useful abstraction, indeed serving as the main departure point for literally the entire discipline. I do however believe it useful to also layer on some nuance to help us distinguish between different situations within this overarching context. *Yes*, we presume a breach has taken place, but there is still a difference between the way we behave when we simply presume, suspect, and actually confirm a breach. 

So as a threat hunter we need to distinguish between different mental modes, meaning *mindsets*, that in turn drive our behaviour re: which exact tools, techniques, and procedures we decide to apply at any given moment. 

# MODE 1 - OPEN-MINDED EXPLORATION

{{< figure src="/img/openmind.gif" title="" class="custom-figure" >}}

In the beginning, we approach a system anew, presuming a breach has taken place, even though we do not yet possess any concrete indications to confirm that it is indeed so. In this mode, we operate with a 'beginner's mind' - we strive to stay objective, free of bias, and regard everything as potentially suspicious. Here, it's more about breadth than depth.  

Instead of honing in on any specific process, event, connection, or service, we look at our system as a whole. We examine major indicators - high-probability, representative samples - and scrutinize them with the goal of finding any sign that something might potentially be rotten in the state of Denmark. 

And then, once we do...

# MODE 2 - BUILDING A CASE

{{< figure src="/img/inspector.gif" title="" class="custom-figure" >}}

The second mode begins the moment we find something that triggers our 'Spidey sense' - perhaps an unusual parent-child process relationship or a sporadic connection to an unknown IP. Something is off, our interest is piqued, but it's not a slam dunk yet. The last thing we want to do at this point is get trigger happy and call in the DFIR cavalry only for them to immediately refute our claim. Oh hell naw. 

So our mindset switches - instead of viewing everything as potentially suspect, we now `seek to build a case` around what we've identified as the potential indicator of compromise. We thus employ a more focused methodology, collecting supporting evidence until we feel satisfied that our conviction stands on firm empirical grounds, i.e., there's a (very) low probability of a false positive

Once this threshold has been reached we then declare an incident and alert DFIR. 

# MODE 3 - SUPPORT + COMMUNICATION 

{{< figure src="/img/dontworry.gif" title="" class="custom-figure" >}}

The key thing to know here is that the moment DFIR is alerted and confirms the incident *they* become the ones calling the shots. We are no longer the lead, we are the support - meaning this mode can be quite variable in nature. We can be very involved with the proceedings, or we can be not involved in all. Again, the decision of our involvement is now outside our purview.

So this mode is less formulaic since it can manifest in so many different ways. We will likely receive strict and highly specific instructions on how to proceed from DFIR, which we will be obligated to adhere to. Our goal is thus to support them to serve the greater goal of minimzing Mean Time to Remediation (MTTR).

# FINAL THOUGHTS

I hope this conceptual model will be of some use to you. For me, it helps to guide the overall operational strategy, especially when it comes to distinguishing between Modes 1 and 2. As a simple example, I view log analysis as a poor choice for Mode 1 since we could be dealing with a vast amount of data to sort through for as-of-yet undefined signs of compromise. This is quite impractical. However, once we switch to Mode 2 and start searching for specific signs, which helps limit what logs are of interest, log analysis can become a very useful tool to build our case.

If you'd like to learn more consider reading up on the following different, yet related, models:
[Cyber Threat Intelligence Lifecycle](https://www.crowdstrike.com/cybersecurity-101/threat-intelligence/)
[The Diamond Model for Intrusion Analysis](https://securityboulevard.com/2023/03/diamond-model-of-intrusion-analysis-a-quick-guide/)

***
