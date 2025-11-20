---
showTableOfContents: true
title: "What Is Threat Hunting?"
type: "page"
---
## Defining Threat Hunting

Threat hunting is the proactive and iterative search through networks, endpoints, and datasets to detect malicious, suspicious, or risky activities that have evaded existing security solutions. As David Bianco, a leading authority in the field, describes it: threat hunting encompasses "any manual or machine-assisted process for finding security incidents that your automated detection systems missed."

This definition captures something essential. Unlike traditional security operations that sit back and wait for alerts, threat hunting operates on a less comfortable assumption: some threats will inevitably slip through the cracks. Your fancy SIEM didn't catch it. Your EDR didn't flag it. Your firewall waved it right through. The hunter's job is to find those threats anyway.

At its core, threat hunting represents a fundamental shift in defensive strategy. Traditional security models operate on a detection-response paradigm: deploy controls, wait for alerts, investigate alerts, respond to confirmed incidents.

Threat hunting inverts this model. You begin with a hypothesis or question and actively search for evidence of compromise before automated systems generate alerts - or in cases where automated systems may never generate alerts at all.

The SANS Institute emphasizes three critical characteristics: threat hunting must be **proactive** (hunters initiate investigations rather than waiting for alerts), **iterative** (hunting is a continuous process, not a one-time exercise), and **focused on evasion** (hunting targets threats specifically designed to bypass automated defences).

## Threat Hunting Drives Security Improvement

Here's the crucial insight that separates mature hunting programs from basic ones: **the ultimate goal of threat hunting is not to find more security incidents - it's to drive continuous improvement across your entire security program.**

Think of it like debugging. You don't debug to find bugs - you debug to build better software. Each bug you find teaches you something about your system, improves your testing, and makes future bugs less likely. Threat hunting works the same way.

As Bianco emphasizes, when a hunter figures out a new way to detect malicious behaviour, the goal should be to automate that detection. That way, the next time similar malicious activity occurs, it will be alerted on and responded to quickly. This creates a virtuous cycle where human creativity feeds automated defences.

Consider what happens during threat hunts. Hunters are constantly exploring parts of your network data that other people either don't examine often or analyze in different ways. This unique perspective means hunters are more likely to notice:

- **Visibility gaps**: Areas where logging is insufficient or missing entirely
- **Data collection issues**: Problems with log forwarding, parsing, or retention
- **Misconfigurations**: Systems not set up according to security best practices
- **Detection gaps**: Attack techniques that aren't covered by existing rules

By identifying and reporting these issues, the hunt team becomes a constant source of what organizations often call "opportunities to improve." Each hunt - whether it finds threats or not - produces insights that strengthen your overall security posture.

So don't think of hunting as a way to find more security incidents using expensive humans. Instead, think about threat hunting as a way to improve your entire security program over time. We're not merely hunting to find present bad - we're hunting to decrease the probability of future bad.



## Core Characteristics of Threat Hunting

### 1. Machine-Assisted but Human-Led

Here's the reality: due to the volume and velocity of security data coming into today's organizations, we require good automated detection to keep up. With this much data, human review isn't just expensive - it's impossible to do comprehensively. This is precisely why the machine-assisted nature of hunting is critical.

But while threat hunting leverages sophisticated tools and technologies extensively, it fundamentally depends on something that cannot be automated: human creativity, intuition, and contextual reasoning. Unlike automated detection products, which can only alert on what they've been programmed to find, humans excel at identifying patterns even in the face of incomplete or ambiguous data.

Consider what makes an effective threat hunter. They must think like an adversary, constantly asking "If I wanted to compromise this environment, how would I do it?" This adversarial mindset allows hunters to anticipate attack paths that haven't been explicitly documented or detected before. When investigating, they recognize subtle anomalies that automated systems miss because those systems can only alert on what they've been programmed to detect.

Context is perhaps the most critical element that humans bring to hunting. Automated systems lack the deep understanding of business operations, organizational politics, user relationships, and system architecture that experienced hunters accumulate.

When you see a database administrator running unusual queries at 2 AM, is that malicious or is it emergency maintenance for tomorrow's product launch? When you observe large data transfers to a cloud storage service, is that data exfiltration or is the marketing team uploading assets for the new campaign?

These questions require contextual knowledge that exists in human minds, not in rule engines.



### 2. Hypothesis-Driven Investigation

Threat hunting typically begins with a hypothesis - an educated guess about how an adversary might compromise or move through your environment. These hypotheses are informed by:

- Threat intelligence about active adversary campaigns
- Understanding of MITRE ATT&CK techniques relevant to your environment
- Knowledge of organizational vulnerabilities and high-value assets
- Previous incidents or near-misses
- Anomalies observed in baseline behaviour
- Behavioural analysis
- ML- or statistical model-based scoring

For example: "If an adversary compromised a user workstation, they might attempt to enumerate domain administrators using built-in Windows utilities like 'net group' or PowerShell cmdlets. Let me search our endpoint logs for processes executing these commands from non-administrative users."

Another example informed by behavioural insights: "The way this specific process is periodically communicating outbound over an unusual port to an unknown host is reminiscent of C2 communication. Let me look at the communication closer using RITA and Zeek to see what else I can learn about the nature of the communication."





### 3. Focused on Threats That Evade Detection

Your automated systems are good at their jobs. Your antivirus catches known malware signatures. Your firewall blocks suspicious IP addresses. Your SIEM alerts on predefined correlation rules.

But what about the zero-day exploit that has no signature yet? What about the adversary who uses PowerShell, WMI, and other legitimate administrative tools - techniques often called "living off the land" - to avoid deploying any malicious binaries that could be detected?

Intrusion prevention doesn't work every time. The stealthy techniques attackers use can often escape detection, and attackers are innovating at an alarming rate, resulting in a constant stream of new and updated attack techniques. In fact, threats can break into any network and avoid detection for up to 280 days on average.

The fundamental assumption underlying all threat hunting is both sobering and pragmatic: some threats will inevitably bypass automated defenses. As SANS instructor Rob Lee memorably puts it, signature-based detection looks for "deer," but adversaries can evade detection by simply adding a "moustache" - making trivial modifications to their tactics, techniques, and procedures.

Change a few bytes in malware, alter command-line parameters, or use a different file extension, and suddenly the threat your defenses were trained to recognize becomes unrecognizable. Automated systems can't broaden their detection criteria to catch these variations because doing so would generate unacceptable false positive rates and potentially degrade system performance. The cure would be worse than the disease.

This is where human threat hunters provide irreplaceable value. While machines fail to recognize the deer once the moustache is added, humans excel at pattern recognition within broader contextual frameworks. A skilled analyst can look at the evidence and think, "that's a deer wearing a moustache" - recognizing the fundamental nature of the threat despite superficial modifications.

No signature database is complete. No behavioral baseline is perfect. Adversaries continuously adapt their techniques specifically to evade detection. Human investigation becomes the essential complementary capability - bringing creativity, context, and the ability to recognize patterns that no rule anticipated, whether the deer has a moustache or not.


### 4. Iterative and Continuous

Threat hunting is not a project with a defined end date - it's an operational capability that runs continuously. Think of it as analogous to physical security: you don't patrol a facility once and declare it secure forever. The threat landscape evolves. Your environment changes. New attack techniques emerge constantly. Ongoing vigilance isn't optional.

What makes hunting truly powerful is that it creates a self-improving cycle. Each hunt - whether it finds threats or not - produces insights that make future hunts more effective. In a well-designed threat hunting system, even hunts that **find nothing produce organizational value**. They validate that your defences are functioning correctly. They improve hunters' mental models of the environment. They reveal areas where documentation or baselining could be improved. They build the expertise that will make future hunts more efficient.

A mature hunting program views these "null results" not as wasted effort but as important validation and learning opportunities. Finding nothing isn't failure - it's just a different kind of win.


## Types of Threat Hunting

Threat hunting isn't a one-size-fits-all approach. Security teams should plan their process based on available resources, threat landscape, and specific areas of concern.

Rather than thinking about vague categories like "structured" or "unstructured" hunting, Bianco's PEAK Threat Hunting Framework provides a more useful taxonomy. An acronym for "Prepare, Execute, and Act with Knowledge," the framework is vendor- and tool-agnostic and incorporates three distinct types of hunts:


### Hypothesis-Driven Hunts

In this classic approach, hunters form a hunch or hypothesis about potential threats and their activities that may be present on the organization's network. Hunters then use data and analysis to confirm or deny their suspicions.

This is the type of hunting most people envision when they think about threat hunting. It starts with an idea - often informed by threat intelligence, security research, or understanding of the MITRE ATT&CK framework - and systematically tests that idea against available data.




### Baseline Hunts

In this type of hunt, hunters establish a baseline of "normal" behaviour and then search for deviations that could signal malicious activity. This hunt type is sometimes also known as exploratory data analysis (EDA).

Baseline hunting is particularly valuable for discovering anomalies that you didn't know to look for. By understanding what's typical in your environment, you can identify outliers that warrant investigation. These outliers might be malicious activity, might be misconfigurations, or might simply be unusual but benign behaviour - but you won't know unless you investigate.

### Model-Assisted Threat Hunts (M-ATH)

You could describe M-ATH hunts as "Sherlock Holmes meets artificial intelligence" and you wouldn't be wrong. Hunters use machine learning techniques to create models of known good or known malicious behaviour, then look for activity that deviates from or aligns with these models.

This is almost a hybrid of the hypothesis-driven and baseline types, but with substantial automation from ML. The human hunter still drives the process - selecting what to model, interpreting the results, making judgment calls about what requires deeper investigation - but machine learning handles the pattern matching at scale.






## Stages of Every Hunt (PEAK Framework)

According to the PEAK framework, each hunt should follow a three-stage process: Prepare, Execute, and Act.

**In the Prepare phase**, hunters select topics, conduct research, and plan out their hunt. This includes formulating hypotheses, identifying relevant data sources, and developing the analytical approach.

**The Execute phase** involves diving deep into data and analysis. This is where hunters query logs, correlate events, pivot based on findings, and follow investigative leads wherever they go.

**The Act phase** focuses on documentation, automation, and communication. You capture what you learned. You create detection rules to automate future detection of similar threats. You communicate findings to relevant stakeholders.

Crucially, each phase integrates **Knowledge**. This knowledge comes in many forms: organizational or business expertise, threat intelligence and OSINT, prior experience of the hunters, and any findings from the current hunt. Knowledge isn't a separate phase - it's the foundation that informs every stage of the hunting process.



## What Threat Hunting Is Not

To fully understand threat hunting, it helps to clarify what it is not:

**Not Incident Response**: Threat hunting occurs before an incident is confirmed. Once a threat is validated and containment begins, you've transitioned to incident response. Hunting is the search. Incident response is the reaction.

**Not Vulnerability Management**: While hunters should be aware of vulnerabilities in their environment, hunting focuses on detecting active exploitation or presence of threats - not cataloging potential weaknesses.

**Not Penetration Testing or Red Teaming**: These activities simulate attacks to test defences. Threat hunting assumes defences have already been bypassed and searches for actual threats.

**Not Pure Automation**: Tools labeled as "automated threat hunting" are typically advanced anomaly detection or behavioural analytics systems. True hunting requires human hypothesis generation, creative investigation paths, and contextual judgment.

**Not Alert Triage**: Investigating alerts generated by SIEM, EDR, or IDS systems is standard SOC work. Hunting begins without an alert - driven by curiosity or hypothesis rather than system-generated notification.

## The Value Proposition

Why invest in threat hunting? Because it provides several critical capabilities that automated systems simply can't:

### Reduced Dwell Time

Industry research consistently shows that adversaries remain undetected in victim networks for weeks or months - the "dwell time."  Proactive hunting can dramatically reduce this window by discovering threats before they achieve their objectives, minimizing potential losses.

### Discovery of Evasive Threats

Advanced adversaries specifically design their tools, techniques, and operational patterns to avoid detection. Threat hunting is often the only way to discover these sophisticated threats that are invisible to automated systems.

### Improved Security Posture

Threat hunting helps organizations identify and mitigate weaknesses in their detection rules, platforms, and data collection. By actively searching for threats, security teams gain valuable insights that can be used to improve security controls and prevent future attacks.

### Validation of Security Controls

Even hunts that find no threats provide value by validating that existing security controls are functioning correctly and that the environment matches expected baselines. This verification is essential for maintaining confidence in your defensive capabilities.

### Enhanced Detection Capabilities

Insights from hunting feed back into automated detection systems. When hunters discover a novel technique, detection engineers can create rules to catch it automatically in the future - raising the organization's overall security posture. This continuous improvement cycle is one of hunting's most valuable contributions.

### Organizational Learning

Hunting develops deep expertise about the environment, threat landscape, and adversary behavior. This knowledge improves all security functions, from architecture decisions to incident response effectiveness.

### Compliance and Reporting

Threat hunting allows you to provide concrete evidence of proactive threat detection and mitigation efforts. Organizations can document their activities and findings, creating detailed reports that highlight their commitment to security. This transparency ensures compliance and builds trust with stakeholders, customers, partners, and regulators.

## Measuring Threat Hunting Effectiveness

A crucial aspect of any threat hunting program is the ability to measure its impact. According to the [2024 SANS Threat Hunting Survey](https://www.youtube.com/watch?v=3UEFlapr-_4), nearly two-thirds (65%) of organizations now measure their hunting efforts, compared with only 35% the previous year. This dramatic increase reflects growing maturity in the field.

The PEAK framework offers a comprehensive set of hunting metrics. The philosophy is straightforward: hunters must measure what they've done and measure the effects of what they've done. Some areas to measure include:

- **Number of detections created or updated**: How many new or improved detection rules resulted from hunting activities?
- **Number of incidents opened during/as a result of a hunt**: How many actual threats did hunting discover?
- **Numbers of gaps identified and gaps closed**: How many visibility or detection gaps did hunters find, and how many have been addressed?
- **Number of vulnerabilities & misconfigurations identified and the number closed**: What security weaknesses did hunting uncover?

These metrics tell you what you really need to know: is hunting making your security program better? Are you finding threats you would have missed? Are you closing gaps in visibility and detection? Are you automating the discoveries hunters make?


## Growing Trends in Threat Hunting

Interest in threat hunting has been steadily climbing. The SANS 2024 Threat Hunting Survey reveals several important trends:

**Formal threat hunting programs are truly on the rise**. In 2023, 35% of participating organizations had threat hunting programs. This year, threat hunting has reached a majority: 51% of organizations reported they have established true hunting programs.

**Measuring effectiveness is becoming standard practice**. Nearly two-thirds (65%) of organizations now measure their hunting efforts, demonstrating growing program maturity.

**Organizations are increasingly outsourcing this critical practice**. In 2024, 37% leverage external sources for threat hunting. While this makes it easier to establish and scale a program, it introduces potential risks around data control, data governance, and integration with existing security systems.

These statistics demonstrate that threat hunting is moving from cutting-edge to mainstream - a recognition that proactive defense is essential in today's threat landscape.




## Prerequisites for Effective Hunting

Not every organization is ready for threat hunting. Effective hunting requires certain foundational capabilities:

- **Adequate logging and visibility**: Can't hunt what you can't see
- **Centralized data access**: Hunters need efficient access to diverse data sources
- **Baseline understanding of normal**: Must know what's typical to identify what's anomalous
- **Skilled personnel**: Hunters need deep technical knowledge and analytical thinking
- **Time and resources**: Hunting is labor-intensive and can't be rushed
- **Management support**: Leadership must understand hunting may not immediately yield "wins"

Organizations lacking these prerequisites should focus on building fundamentals before launching hunting programs. Attempting to hunt without adequate visibility or skilled personnel will produce frustration rather than results.

## Conclusion: Putting Security on the Offense

Unlike traditional incident detection programs, which are purely reactive, threat hunting is a proactive approach to identifying threat actors on your network that you might not already be detecting well. Threat hunting helps you address threats before they cause significant damage - staying ahead of attackers who are constantly innovating their tactics.

But remember: the goal isn't just to find more incidents. The goal is to continuously improve your entire security program. Each hunt should make your automated defenses smarter. Each gap discovered should be closed. Each technique identified should become automated detection.

Don't think of hunting as an expensive way to find security incidents. Think of it as an investment in continuous security improvement - one that pays dividends across your entire defensive capability. By incorporating threat hunting into your organization's security practices, you harness the power of human-driven pattern recognition while ultimately bolstering your automated detection capabilities.

The threats are out there. The question is: will you wait for them to trigger an alert, or will you go hunting?


---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|NEXT|]({{< ref "./02_history.md" >}})

