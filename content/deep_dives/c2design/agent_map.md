---
showTableOfContents: true
title: "A Map of C2 Agent Behavior on the Endpoint"
type: "page"
---

## Introduction

The modern cyber-attack lifecycle is a multi-stage campaign, but it is within the post-exploitation phase that an adversary's true objectives are pursued and realized. Once initial access is achieved, the C2 agent, a malicious process (or set of processes) operating on the compromised host, becomes the **primary instrument of the intrusion**.

This transition is a shift from a static payload to an active, dynamic entity tasked with establishing a durable presence, understanding its environment, escalating privileges, expanding its foothold, and ultimately, achieving the attacker's strategic goals. 
This report provides an **exhaustive, multi-level conceptual map of the on-host behaviours exhibited by a C2 agent**, detailing the intricate web of actions that unfold after the initial point of compromise.

---

## MITRE ATT&CK

To infuse some sense of order into what could be perceived as an overly complex and chaotic conceptual terrain, 
I've decided to leverage the [MITRE ATT&CK](https://attack.mitre.org) framework as its foundational taxonomy.
The ATT&CK framework provides a globally recognized, behaviour-centric lexicon that **categorizes adversary actions into tactics (the "why") and techniques (the "how")**, based on real-world observations of cyber incidents.

By focusing on the post-compromise tactics, from [Execution (TA0002)](https://attack.mitre.org/tactics/TA0002/) through 
[Exfiltration (TA0010)](https://attack.mitre.org/tactics/TA0010/), this analysis offers what I hope may serve as a 
doctrinal blueprint of adversary operations on the endpoint. 
But, it's crucial to understand that these tactics are not a linear progression. An adversary does not simply move from left to right across the ATT&CK matrix.

Instead, they operate in a fluid, iterative cycle, often returning to previous tactics as they deepen their understanding of the target environment and acquire new capabilities.


This map, therefore, is not a simple checklist but a **conceptual map to the complex, interconnected, and often cyclical landscape of post-exploitation behaviour**.


---

## Part I: Establishing and Fortifying the Beachhead











---
[|TOC|]({{< ref "../../malware/_index.md" >}})
[|PREV|]({{< ref "../../malware/_index.md" >}})

