---
showTableOfContents: true
title: "The Threat Hunting Core Loop"
type: "page"
---



## The Cyclical Nature of Hunting

If you've ever tried to debug a complex software problem, you've experienced something similar to threat hunting. You start with a hypothesis about what might be wrong, you test that hypothesis by examining code, your findings lead to new questions, and you iterate until you either find the bug or validate that your initial hypothesis was incorrect. Then you take what you've learned and apply it to the next investigation.

But here's the critical insight: you don't debug just to find bugs - you debug to build better software. Each bug you find teaches you something about your system, improves your testing processes, and makes future bugs less likely. Threat hunting works the same way.

This reveals a counterintuitive truth about threat hunting: **the ultimate goal isn't to find more security incidents - it's to drive continuous improvement across your entire security program.** We're not merely hunting to find present bad, we're hunting to decrease the probability of future bad.

Threat hunting follows this same iterative, cyclical pattern. Unlike incident response - which typically has a clear beginning (incident detected) and end (threat remediated) - threat hunting is a continuous process without definitive completion. Each hunt feeds into the next, creating a self-improving cycle where organizational knowledge, detection capabilities, and investigative skills continuously advance.

This cyclical nature is so fundamental to threat hunting that virtually every framework and methodology describes some version of a "hunting loop." While specific frameworks differ in their phases and terminology, they all share a common core workflow. Understanding this core workflow provides the conceptual foundation for all hunting activities, regardless of which specific framework you eventually adopt.



## The Core Workflow: Four Fundamental Phases

At its essence, threat hunting follows a four-phase cycle that repeats continuously:

```
  ┌─────────────────────────────┐         ┌─────────────────────────────┐
  │  Phase 1: DETERMINING       │────────>│  Phase 2: INVESTIGATION     │
  │  YOUR FOCUS                 │         │                             │
  │                             │         │  • Query data sources       │
  │  • Hypothesis-Driven        │         │  • Filter and refine        │
  │  • Statistical/Behavioral   │         │  • Follow evidence trails   │
  │  • ML/DL Models             │         │  • Document findings        │
  │  • Threat Intelligence      │         │                             │
  └─────────────────────────────┘         └─────────────────────────────┘
            ↑                                           |
            |                                           |
            |                                           ↓
  ┌─────────────────────────────┐         ┌─────────────────────────────┐
  │  Phase 4: KNOWLEDGE         │<────────│  Phase 3: DISCOVERY AND     │
  │  REFINEMENT & IMPROVEMENT   │         │  ANALYSIS                   │
  │                             │         │                             │
  │  • Create detection rules   │         │  • Threats → IR             │
  │  • Document baselines       │         │  • Nothing → Validation     │
  │  • Share knowledge          │         │  • Gaps → Improvement       │
  │  • Generate new focus areas │         │                             │
  └─────────────────────────────┘         └─────────────────────────────┘
```

Let's explore each phase in depth to understand not just what happens, but why it happens and how it contributes to the overall cycle of continuous security improvement.



## Phase 1: Determining Your Focus

One of the biggest challenges in threat hunting is answering the fundamental question: **where do we begin?** You can't just randomly poke around in logs hoping to find something interesting.

![logs](../img/logs.png)




Given the needle-in-haystack nature of the problem - searching for potentially malicious activity hidden within vast amounts of normal network traffic, system events, and user behaviour - hunters need a systematic way to focus their efforts and determine what to investigate first.

This is what fundamentally distinguishes threat hunting from alert triage - hunters must proactively choose where to direct their investigative efforts, rather than waiting for automated systems to generate alerts. But this choice of focus can be achieved through multiple approaches, each with its own strengths and appropriate use cases.



### Approaches for Determining Your Focus

The good news: there are several effective methods for deciding where to begin your hunt, and you don't have to pick just one - most mature programs blend all these approaches over time.



#### 1. Hypothesis-Driven Hunting

The most widely discussed approach involves generating specific hypotheses about how adversaries might compromise or operate within your environment. You start with an educated guess about a particular attack technique or threat, perhaps informed by the latest threat intelligence, then design an investigation to test whether evidence of that technique exists in your environment.

For example: "I bet adversaries are using scheduled tasks for persistence on our web servers" or "Are any systems beaconing to external infrastructure on regular intervals, suggesting command-and-control communication?"

This approach provides clear direction and focus, making it easy to determine when an investigation is complete and to measure coverage across different attack techniques.



#### 2. Statistical and Behavioural Analysis

Tools like RITA (Real Intelligence Threat Analytics) and AC-Hunter flip the script entirely - instead of starting with a hypothesis, they use network-based behavioural analysis grounded in statistical methods to automatically score and rank potentially suspicious activity. These tools analyze patterns in network traffic - looking at connection frequencies, data volumes, timing patterns, and other behavioural characteristics - to identify statistical outliers that warrant investigation.

Rather than starting with a specific hypothesis, this approach lets the data guide you. The tools surface the most anomalous connections or behaviours, essentially saying "hey, this is unusual, maybe investigate?" - giving you a prioritized list of leads to investigate. You then apply your analytical skills to determine whether these statistically unusual patterns represent genuine threats or benign edge cases.

This approach is particularly effective for discovering unknown threats or attack patterns you hadn't thought to hunt for, as it's driven by what's actually unusual in your environment rather than preconceived notions of what threats might be present.




#### 3. Machine Learning and Deep Learning Models

Machine learning and deep learning models can be applied to network and system data to identify suspicious patterns and suggest specific connections or events that warrant deeper investigation. These models are trained to recognize characteristics associated with malicious activity and can highlight specific Zeek logs, network flows, or system events that deviate from learned baselines.

This approach combines elements of both hypothesis-driven and statistical analysis - the models are often trained with specific threat scenarios in mind, but they also discover novel patterns through their analysis of your actual data. They provide hunters with specific starting points: "investigate this connection" or "examine this process execution," backed by a probability score or confidence level.



#### 4. Baseline-Anomaly Hunting

Baseline hunting flips the question entirely - instead of looking for specific threats, you map what "normal" looks like in your environment, then investigate anything that doesn't fit the pattern. This involves establishing comprehensive baselines of normal behaviour and investigating deviations from those baselines. The initial focus is often simply "something unusual is happening" - the investigation determines whether it's benign or malicious.



#### 5. Other Focusing Methods

Beyond these primary approaches, hunters may determine their focus through:

- **Threat Intelligence Alerts**: External intelligence reporting a specific campaign or technique targeting organizations like yours
- **Vulnerability-Driven Hunting**: Focusing on systems affected by newly disclosed vulnerabilities
- **Asset-Based Prioritization**: Beginning with your most critical assets and investigating activity around them
- **Temporal Triggers**: Hunting in response to specific events like mergers, incident anniversaries, or geopolitical developments



### Choosing Your Approach

Most mature hunting programs don't rely exclusively on one method. Instead, they blend approaches based on available resources, current threat landscape, and organizational priorities. You might use statistical analysis tools to surface interesting leads, then develop hypotheses about specific patterns you observe. Or you might start with hypothesis-driven hunting of a specific technique, then use ML models to find variants or related activity.

The key is understanding that the first phase isn't about rigidly following one methodology - it's about having systematic methods for answering "where do we begin?" Each approach provides a different lens for examining your environment, and using multiple approaches over time ensures comprehensive coverage.



## Phase 2: Investigation

With your focus determined, you begin the investigation phase. This is where hunters spend most of their time, querying data sources, analyzing results, and following leads wherever they point. This is where you actually do the detective work.



### The Art of Querying

Investigation typically begins with queries against your available data sources. Though the specific tools and query languages vary, the intellectual process remains similar.

Initial queries are often broad, casting a wide net based on your starting point. If hunting for scheduled task persistence, you might start by querying all scheduled task creation events over the past 30 days. If investigating a high-scoring connection from RITA, you might pull all related flows and contextual data. This initial query will likely return far more results than you can investigate individually, which leads to the next step: filtering and refinement.

You progressively narrow results by filtering out known-good activity. Perhaps you exclude tasks created by your patch management system, or filter out tasks that run standard Windows maintenance jobs. This filtering requires environment knowledge - you need to understand what's normal to identify what's anomalous.

As you analyze results, you begin to notice patterns. Maybe most scheduled tasks run at predictable times for legitimate purposes, but a few run at odd hours. Maybe most tasks execute standard system binaries, but a few execute PowerShell with encoded commands. These patterns guide further investigation.



### Following the Evidence

Investigation is iterative and rarely linear. Rarely does a single query definitively prove or disprove your initial focus. Instead, investigation proceeds iteratively, with each query informing the next. You find something interesting in scheduled tasks, which leads you to investigate the user account that created those tasks, which leads you to examine authentication logs for that account, which leads you to investigate the source IP addresses where those authentications originated... you get it.

This is where hunting becomes intellectually challenging and rewarding. You're following a trail of evidence, not knowing where it leads. You must decide which threads to pursue and which to abandon. You must recognize when you're chasing false positives versus genuine threats. You must balance thoroughness with efficiency - investigating deeply enough to reach conclusions, but not so deeply that investigations never conclude.

The "art" is knowing which threads to pull and which to drop. Experienced hunters develop intuition about which patterns deserve deeper investigation. Something might be technically anomalous but contextually explainable - a database administrator running unusual queries at 2 AM might be responding to a critical production issue. Distinguishing between genuinely suspicious activity and benign anomalies requires both technical skills and organizational context.




### Documenting Your Investigation

As you investigate, document your process. Note which queries you ran, what results you found, why you followed certain leads, and what you learned.

**This serves several purposes:**

If you find a threat, your documentation becomes the initial incident report for response teams. If you find nothing, your documentation helps other hunters avoid duplicating your work. If you get interrupted and must resume investigation later, your notes help you pick up where you left off. When sharing knowledge with team members, your documented process helps them learn investigative techniques.

Documentation need not be perfect or elaborate during active investigation. Quick notes, saved queries, and screenshots often suffice. You can polish documentation later if needed. The key is capturing enough information to make your investigation reproducible and your reasoning understandable.


## Phase 3: Discovery and Analysis

Eventually the investigation reaches a conclusion. You've found evidence that supports your initial focus, found nothing of concern, or determined that you can't conclusively answer your question with available data. Here's the critical mindset shift: **approached with this perspective, any outcome is a win - they can all make you more secure.** Each outcome has value and leads to different next steps.


### Finding Threats: The Obvious Win

When investigation reveals actual threats, this represents the most visible form of hunting success. You've discovered adversaries that automated systems missed. You've potentially shortened dwell time significantly. You've validated that hunting provides value.

But finding threats is just the beginning of a new process. At this point, hunting typically transitions to incident response. You document your findings, preserve evidence, and escalate to the incident response team - in which case awesome, hand it to incident response. They'll take over containment, eradication, and recovery while you return to hunting activities.

However, before completely handing off, hunters often conduct "scoping" activities - searching for additional indicators related to the discovered threat. If you found one system compromised with specific malware, are other systems similarly compromised? If you discovered credential theft, what other accounts might be compromised? This scoping helps IR teams understand the full extent of compromise.

The threat discovery also generates immediate feedback for your detection engineering. Whatever you found through manual hunting should ideally be detectable automatically in the future. You work with detection engineers to create SIEM rules, EDR detections, or other automated alerts that would catch similar threats going forward.





But maybe, most probably in fact, you found nothing...



### Finding Nothing: The Hidden Win



Many hunts conclude without finding active threats. Novice hunters sometimes view this as failure, but finding nothing isn't failure - it's validation, or discovery of a different kind. When a thorough hunt finds no evidence of specific attack techniques, you've validated several things:

You've confirmed that your detection coverage for those techniques is working (win!), or that those techniques aren't currently being used against you. You've improved your understanding of what normal looks like in your environment, making future anomaly detection more effective.

Or it might reveal that a detection you thought existed doesn't actually exist, or that it's misconfigured and wouldn't fire even if the attack happened (also a win - now you know what to fix). You've identified gaps in logging or visibility that should be addressed to enable better hunting in the future.

Perhaps most importantly, you've exercised and improved your hunting skills. Each investigation makes you more proficient with your tools, more familiar with your environment, and more skilled at distinguishing signal from noise.

Organizations that punish or devalue hunts that find nothing create perverse incentives. Hunters may avoid thorough investigation, cherry-pick easy hypotheses likely to find something, or even exaggerate the significance of minor findings. Mature hunting programs celebrate thorough investigations regardless of outcome, recognizing that validation of security controls has genuine value.




### Finding Gaps: The Improvement Opportunity

Sometimes investigation reveals neither active threats nor validation of security posture, but rather gaps in your capabilities. Perhaps you can't fully investigate your initial focus because you lack necessary logging. Maybe you discover that key data sources aren't being retained long enough. Perhaps you find that you can't correlate events across different systems effectively.

These gaps are valuable discoveries in themselves. They provide specific, evidence-based justification for security investments. Rather than abstractly arguing for better logging, you can point to specific hunting focuses you couldn't investigate and threats you couldn't detect without improved telemetry.

Documenting these gaps and advocating for their closure becomes part of the hunting program's value proposition. Each closed gap expands your hunting capabilities and overall detection coverage.



## Phase 4: Knowledge Refinement and Improvement

The final phase of each hunting loop involves capturing lessons learned and taking concrete improvement actions. **This is where threat hunting transforms from investigation into force multiplication.** This phase transforms hunting from simple investigation into a force multiplier that continuously improves organizational security. Everything you learned feeds back into your security program.



### Creating Detection Rules

When you discover new attack techniques or novel implementations of known techniques, the first improvement action is often creating automated detection rules. You shouldn't need to manually hunt for the same thing repeatedly - once discovered, it should be detected automatically going forward.

Working with detection engineers, you translate your hunting findings into SIEM correlation rules, EDR detection logic, or other automated alerts. This requires balancing sensitivity and specificity: the rules should catch genuine threats without generating excessive false positives that overwhelm SOC analysts.

Good detection rules often require iteration. Initial rules might be too sensitive (too many false positives) or too specific (missing variants of the technique). Over time, rules are refined based on operational experience, creating increasingly effective automated detection.



### Documenting Techniques and Baselines

Each hunt improves your understanding of your environment. You learn what's normal: which scheduled tasks typically run, what authentication patterns look like, how network traffic flows, which processes commonly execute on different system types.

Documenting these baselines about normal behavior in your environment has multiple benefits. Future hunters can reference them when investigating anomalies. New team members can learn about the environment more quickly. Detection engineers can use them when tuning rules. Incident responders can use them to assess whether observed activity is anomalous.

Some organizations maintain formal "normal behavior" documentation, while others rely on tribal knowledge and informal notes. Either approach works, but documented knowledge scales better as teams grow and personnel change.



### Sharing Knowledge

Hunting knowledge should be shared both within your organization and, where appropriate, with the broader security community. Internally, regular threat hunting briefings keep SOC teams, incident responders, and security leadership informed about hunting activities, findings, and capabilities.

When hunters discover new techniques, novel indicators, or effective investigation approaches, sharing these with the security community (through blog posts, conference presentations, or threat intelligence sharing platforms) raises the bar for all defenders. Obviously, you must sanitize organizational details and avoid revealing information that could harm your security posture, but sharing TTPs and detection strategies helps everyone.

This knowledge sharing also builds your hunters' professional reputations and contributes to career development - important factors in retaining skilled personnel.


### Generating New Hypotheses

Each completed hunt naturally generates new hypotheses for future investigation (back to Step 1...). Perhaps you investigated one MITRE ATT&CK technique and now want to explore related techniques. Maybe your investigation revealed an area of your environment you didn't fully understand, prompting deeper exploration. Perhaps findings suggest new threat scenarios worth investigating.

These new hypotheses feed directly into the next iteration of the hunting loop. This is what makes hunting truly cyclical - each hunt's conclusion becomes the next hunt's beginning. Over time, this creates comprehensive coverage of your threat landscape, with hunters systematically working through potential attack vectors while continuously circling back to revisit previous areas with fresh perspectives and new intelligence.



### Measuring and Communicating Value

The refinement phase also includes measuring hunting program effectiveness and communicating value to stakeholders. While measuring proactive activities is challenging (you can't easily quantify prevented incidents), several metrics demonstrate hunting value:

- Number and severity of threats discovered
- Detection rules created based on hunting findings
- Dwell time reduction for discovered threats
- Coverage of MITRE ATT&CK techniques investigated
- Visibility gaps identified and closed
- "Negative validation" of security controls through thorough investigation

Regular reporting on these metrics helps justify hunting program investment and demonstrates continuous value even during periods when active threats aren't discovered.



## The Continuous Improvement Cycle: The Self-Improving Spiral

What makes the hunting loop powerful is its self-improving nature, its cumulative nature. Each iteration enhances organizational capabilities:

Your detection coverage expands as hunting findings become automated rules. Your environmental knowledge deepens as each hunt reveals more about normal behaviour. Your hunters' skills improve through repeated investigation practice. Your data sources and tooling improve as gaps are identified and addressed. Your organizational security posture strengthens through systematic exploration of attack vectors.

This improvement is cumulative and accelerating. Your first few hunts might feel slow and uncertain as hunters learn the environment and build baselines. But as knowledge accumulates and skills improve, hunts become more efficient and effective. Patterns become obvious. Investigation approaches become routine. Areas previously opaque become clear. Patterns previously hidden become obvious. Investigation approaches previously uncertain become routine.

Organizations that commit to sustained threat hunting programs often describe a maturity curve where the program's value increases significantly over the first 12-18 months as this knowledge and capability accumulation reaches critical mass.

Remember: you're not just looking for today's threats. You're building tomorrow's defenses. Every loop makes the next one better. That's the loop that keeps on giving.



## The Loop as Foundation

This chapter has presented the core hunting workflow at a conceptual level, showing the fundamental cycle that underlies all threat hunting activities. In the next chapter, we'll explore specific frameworks that provide more detailed methodologies, tools, and structure around this core loop. But regardless of which framework you adopt, you'll find it's built on these fundamental phases: determine focus, investigate, analyze findings, refine knowledge, and repeat.

Understanding this core loop provides the conceptual foundation for effective hunting. The specifics of tools, techniques, and procedures matter, but they're tactical implementations of this strategic workflow. Master the loop, and you'll be able to adapt to new tools, new threats, and new environments. Lose sight of the loop, and you'll find yourself doing activities that look like hunting but don't provide hunting's core value: continuous, self-improving discovery and organizational learning.





---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./04_other.md" >}})
[|NEXT|]({{< ref "./06_maturity.md" >}})

