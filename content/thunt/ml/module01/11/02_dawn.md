---
showTableOfContents: true
title: "The Dawn of Artificial Intelligence (1943-1956)"
type: "page"
---



## The Conceptual Foundation

The story begins not with computers, but with neurons. In **1943**, neurophysiologist Warren McCulloch and mathematician Walter Pitts published a groundbreaking paper: "[A Logical Calculus of the Ideas Immanent in Nervous Activity](https://www.cs.cmu.edu/~epxing/Class/10715/reading/McCulloch.and.Pitts.pdf)." They proposed that neural activity could be modeled using mathematical logic - that the brain's neurons were essentially performing logical operations.

**Why this mattered:** This was the first formal suggestion that thinking could be computational. If neurons are logic gates, then perhaps intelligence is just very complex computation.

Their model was simple: a neuron receives inputs (either on or off), sums them with weights, and fires (outputs 1) if the sum exceeds a threshold, otherwise it doesn't fire (outputs 0). Sound familiar? This is the direct ancestor of the artificial neurons we use today.

```
Simple McCulloch-Pitts Neuron Logic:
Input 1 (x₁) ──→ [weight w₁]
Input 2 (x₂) ──→ [weight w₂]  ──→ Σ(wᵢxᵢ) ──→ Threshold ──→ Output (0 or 1)
Input 3 (x₃) ──→ [weight w₃]
```

## The Birth of "Artificial Intelligence"

Fast forward to **1956**: [The Dartmouth Summer Research Project on Artificial Intelligence](https://en.wikipedia.org/wiki/Dartmouth_workshop). This two-month workshop is considered the official birth of AI as a field. The proposal, written by John McCarthy, Marvin Minsky, Nathaniel Rochester, and Claude Shannon, made a bold claim:

> "We propose that a 2-month, 10-man study of artificial intelligence be carried out... The study is to proceed on the basis of the conjecture that every aspect of learning or any other feature of intelligence can in principle be so precisely described that a machine can be made to simulate it."

**The optimism was intoxicating.** They believed that within a generation, machines would match human intelligence. This wasn't unreasonable hubris - early successes seemed to validate their confidence.



## Early Triumphs

The 1950s and early 1960s saw remarkable achievements that seemed to prove AI's inevitability:

**1. Logic Theorist (1956):** Created by Allen Newell and Herbert Simon, this program proved mathematical theorems from Russell and Whitehead's _Principia Mathematica_. It even found a more elegant proof than the original for one theorem. This was intelligence, wasn't it?

**2. General Problem Solver (1957):** Also by Newell and Simon, this program attempted to solve any problem that could be expressed as a series of logical rules. The ambition was breathtaking: one algorithm to rule them all.

**3. ELIZA (1966):** Joseph Weizenbaum's program simulated a Rogerian psychotherapist through pattern matching and substitution. While Weizenbaum intended it as a demonstration of superficiality, many users became emotionally attached, believing the machine understood them.

These systems operated on **symbolic AI**: the idea that intelligence involves manipulating symbols according to rules. If you can write down the rules of chess, you can program a chess-playing intelligence. If you can formalize medical diagnosis, you can build a diagnostic system.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./01_intro.md" >}})
[|NEXT|]({{< ref "./03_first.md" >}})

