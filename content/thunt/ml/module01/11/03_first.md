---
showTableOfContents: true
title: "The First AI Winter (1974-1980)"
type: "page"
---

## The Promise Collapses

By the early 1970s, the euphoria was fading. Despite early successes, AI systems consistently failed to scale beyond toy problems. Why?

### **Problem 1: Combinatorial Explosion**

Early AI relied on search - exploring possible solutions until finding the right one. But most interesting problems have astronomical solution spaces.

Consider chess: after each player makes just 4 moves, there are over 288 billion possible positions. Early computers couldn't search effectively at this scale. The "frame problem" in robotics was even worse: representing all the things that _don't_ change when you act is impossibly complex.

### **Problem 2: The Symbol Grounding Problem**

Symbolic AI manipulates symbols like "cat" or "threat," but the symbols have no intrinsic meaning to the computer - they're just tokens. A human knows what a cat _is_ through sensory experience. The computer just shuffles strings. This works for formal domains (chess, logic) but fails for real-world understanding.

### **Problem 3: The Perceptron Scandal**

In 1958, Frank Rosenblatt introduced the **[Perceptron](https://en.wikipedia.org/wiki/Perceptron)**, an algorithm inspired by biological neurons that could learn to classify inputs. It caused tremendous excitement - the New York Times proclaimed it "the embryo of an electronic computer that the Navy expects will be able to walk, talk, see, write, reproduce itself, and be conscious of its existence."

But in **1969**, Marvin Minsky and Seymour Papert published [_Perceptrons_](https://rodsmith.nz/wp-content/uploads/Minsky-and-Papert-Perceptrons.pdf), mathematically proving that single-layer perceptrons couldn't learn certain simple functions like XOR (exclusive OR). While they acknowledged that multi-layer networks might overcome this, they were skeptical about training them.

The impact was devastating. Funding for neural network research dried up almost overnight. The field entered what's called the **First AI Winter** - a period of disillusionment, reduced funding, and scarce progress.

## What We Learned

This wasn't wasted time. The failure of symbolic AI revealed deep truths:

- **Intelligence isn't just logic:** Real-world intelligence requires handling uncertainty, ambiguity, and partial information - not just manipulating logical symbols.
- **Knowledge representation is hard:** Capturing common sense in formal rules proved nearly impossible. We know that wet things shouldn't be put in paper bags, but formalizing such knowledge is extraordinarily difficult.
- **Scale matters:** Approaches that work on toy problems often collapse under realistic complexity.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./02_dawn.md" >}})
[|NEXT|]({{< ref "./04_expert.md" >}})

