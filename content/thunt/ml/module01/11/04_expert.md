---
showTableOfContents: true
title: "Expert Systems and the Second Wave (1980-1987)"
type: "page"
---


## A Narrower Ambition

In the 1980s, AI researchers took a more pragmatic approach: instead of general intelligence, build **expert systems** - programs that captured the knowledge of human experts in narrow domains.

The architecture was straightforward:

```
Expert System Architecture:
┌─────────────────┐
│  Knowledge Base │ ← Rules from domain experts
│   (IF-THEN)     │    "IF temperature > 100°F AND cough
└────────┬────────┘     THEN consider pneumonia"
         │
         ↓
┌─────────────────┐
│ Inference Engine│ ← Applies rules to facts
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ User Interface  │ ← Questions and explanations
└─────────────────┘
```

## Notable Successes

**1. MYCIN (1974-1979):** Diagnosed bacterial infections and recommended antibiotics. It performed at the level of expert physicians, sometimes better. The catch? Building it required interviewing infectious disease specialists for years to extract and formalize their knowledge.

**2. XCON (1980s):** Configured VAX computer systems for Digital Equipment Corporation. This wasn't diagnosis but configuration - a perfect fit for rule-based systems. It saved the company millions.

**3. DENDRAL (1965-1970):** Inferred molecular structures from mass spectrometry data. It helped chemists discover new chemical compounds.

## The Fatal Flaw

Expert systems were brittle. They excelled in their narrow domains but couldn't generalize or adapt. Adding new knowledge required laboriously encoding more rules. Worse, they couldn't learn from experience - every rule had to be hand-crafted by human experts.

By the mid-1980s, companies had invested billions in expert systems. When these systems failed to deliver on their promise (the infamous "maintenance nightmare"), disillusionment set in again.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./03_first.md" >}})
[|NEXT|]({{< ref "./05_second.md" >}})

