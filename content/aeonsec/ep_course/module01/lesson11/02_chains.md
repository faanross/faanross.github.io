---
showTableOfContents: true
title: "Phishing Attack Chains: The Complete Picture"
type: "page"
---

## Understanding the Attack Chain

A phishing attack is not a single action - it's a carefully orchestrated sequence of events. Let's break down the typical chain:

```
[Reconnaissance] → [Weaponization] → [Delivery] → [Exploitation] → [Installation] → [C2]
```

**Detailed Breakdown:**

1. **Reconnaissance Phase** 
    
    - Attacker identifies targets, gathers email addresses, understands organizational structure
- Researches technologies in use, employee roles, current events (OSINT)
2. **Weaponization Phase** (Our Focus)

    - Creating malicious documents or files
    - Embedding exploits or macros
    - Obfuscating payloads
    - Testing against antivirus
3. **Delivery Phase** (Our Focus)

    - Email delivery with social engineering
    - Web-based delivery (watering holes, malicious ads)
    - Physical media (USB drops)
    - Supply chain compromise
4. **Exploitation Phase**

    - User opens document
    - Macro/exploit executes
    - Code runs in context of application
5. **Installation Phase**

    - Payload downloads additional stages
    - Establishes persistence
    - Begins reconnaissance
6. **Command & Control (C2) Phase**

    - Beacon reaches out to attacker infrastructure
    - Attacker gains interactive access
    - Mission objectives begin

## The Psychology of Phishing

Effective phishing exploits human psychology more than technical vulnerabilities. Attackers leverage:

**Cognitive Biases:**

- **Authority Bias**: Posing as executives or IT departments
- **Urgency**: Creating time pressure to bypass critical thinking
- **Familiarity**: Mimicking known vendors or services
- **Curiosity**: Using intriguing subject lines or attachments

**Common Pretexts:**

- "Urgent: Your password will expire in 24 hours"
- "Invoice for your recent purchase" (with attachment)
- "HR: Update your benefits by end of day"
- "IT Security: Click here to verify your account"
- "You've received a secure file" (file-sharing service spoofs)

**Target Selection Strategy:** Attackers often operate on different tiers:

- **Spray and Pray**: Broad campaigns to thousands, hoping for any foothold
- **Targeted Spear Phishing**: Researched attacks against specific individuals
- **Whaling**: Highly crafted attacks against executives/high-value targets




---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

