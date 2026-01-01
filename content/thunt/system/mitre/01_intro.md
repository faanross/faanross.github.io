---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---

## Introduction to Using MITRE ATT&CK for Threat Hunting
When you first encounter the MITRE ATT&CK framework, it can feel overwhelming. Hundreds of techniques, thousands of procedures, an ever-expanding matrix of adversary behaviours - it's easy to wonder whether this massive knowledge base is truly practical for day-to-day threat hunting, or just an abstract academic exercise.

Many security professionals initially struggle to see how ATT&CK translates from theory to action. Is it just another compliance checklist? A taxonomy for writing reports? A nice-to-have reference that sits unused while we tackle real security problems?

The reality is quite different. Once you understand how to apply it, the MITRE ATT&CK framework becomes one of the most valuable tools in a threat hunter's arsenal. It helps us understand our detection coverage - what we can see and what we're blind to. It provides a common language for receiving threat intelligence that's actually actionable rather than vague warnings about "advanced threats." It enables us to share findings with our teams and the broader security community in a way that's immediately understood and useful. And perhaps most importantly for hunters, it guides our hunts by giving us a structured approach to exploring adversary behaviours in our environments.

At its core, ATT&CK is organized around a simple but powerful structure: 
- **Tactics** represent the adversary's objectives - the "why" of what they're doing. Think of these as the stages of an attack: gaining initial access, establishing persistence, moving laterally, and so on. 
- **Techniques** describe how adversaries achieve those tactical objectives - the specific methods they employ. 
- **Sub-techniques** provide even more granular detail, breaking down techniques into specific variations and implementations.

Currently, the framework defines fourteen tactics that map the typical lifecycle of a cyberattack. But here's something crucial for threat hunters to understand: these fourteen tactics are not equally valuable for our work. Some tactics, like Reconnaissance and Resource Development, occur entirely outside our environment with virtually no telemetry for us to hunt through. Others, like Lateral Movement and Command and Control, generate rich, observable behaviors that provide excellent hunting opportunities.

Even among the "huntable" tactics, our approaches differ dramatically. Privilege Escalation hunts rely almost exclusively on endpoint telemetry. Command and Control detection requires deep network traffic analysis. Lateral Movement often demands both endpoint and network visibility working together.

In this article, we'll explore each of the fourteen tactics with a threat hunter's perspective - understanding what they mean, why they matter, and most critically, which ones deserve our focused attention. For those tactics that provide genuine hunting opportunities, we'll walk through concrete examples of how to detect them in real environments, moving beyond theoretical knowledge to practical application.


---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|NEXT|]({{< ref "./02_recon.md" >}})

