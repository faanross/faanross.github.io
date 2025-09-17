---
showTableOfContents: false
# tags: ["",]
title: "Let's Build a Multi-Modal C2 Covert Channel in Golang"
type: "page"
---

## Overview
Below are the lecture notes for my AntiSyphon workshop presented **September 18, 2025**. Though the notes are in
general more descriptive than the actual lectures, they are not expanded, meaning the content from the lectures roughly map
1:1 onto these notes.


## Solutions
All the final solutions are available on GitHub [here](https://github.com/faanross/workshop_antisyphon_18092025). You'll find a copy of the complete slides, as well as the
lectures available here in markdown format. 


[//]: # (## Part A: Welcome + Theory)

[//]: # (- [Welcome To The Workshop]&#40;{{< ref "part_a/01_welcome.md" >}}&#41;)

[//]: # (- [The C2 Agent Communication Loop]&#40;{{< ref "part_a/02_loop.md" >}}&#41;)

[//]: # (- [C2 Over DNS + The Multi-Modal Advantage]&#40;{{< ref "part_a/03_dns.md" >}}&#41;)

[//]: # (- [Multi-Modal Case Studies]&#40;{{< ref "part_a/04_multi.md" >}}&#41;)

[//]: # (- [What We'll Be Creating]&#40;{{< ref "part_a/05_creation.md" >}}&#41;)

[//]: # (- [Interfaces + Composition in Golang]&#40;{{< ref "part_a/06_golang.md" >}}&#41;)




## Foundation
- [Setup Guide]({{< ref "part_b/00_setup.md" >}})
- [Project Structure and Interfaces]({{< ref "part_b/01_interfaces.md" >}})
- [YAML-based Configuration Management System]({{< ref "part_b/02_yaml.md" >}})

## HTTPS Implementation
- [HTTPS Server]({{< ref "part_c/01_https_server.md" >}})
- [HTTPS Agent]({{< ref "part_c/02_https_agent.md" >}})
- [HTTPS Loop]({{< ref "part_c/03_https_loop.md" >}})

## DNS Implementation
- [DNS Server]({{< ref "part_d/01_dns_server.md" >}})
- [DNS Agent]({{< ref "part_d/02_dns_agent.md" >}})
- [DNS Loop]({{< ref "part_d/03_dns_loop.md" >}})

## Transition Using API Switch
- [Implement API Switch]({{< ref "part_e/01_api.md" >}})
- [Dual-server Startup]({{< ref "part_e/02_dual.md" >}})
- [Agent Parsing + Protocol Transition]({{< ref "part_e/03_transition.md" >}})

[//]: # (## Part F: Wrap-up)

[//]: # (- [Where To From Here?]&#40;{{< ref "part_f/01_where_to.md" >}}&#41;)

[//]: # (- [Conclusion]&#40;{{< ref "part_f/02_conclusion.md" >}}&#41;)


___

