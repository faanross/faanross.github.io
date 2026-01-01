---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---

## Resource Development: Building the Arsenal

During resource development, adversaries establish the infrastructure and resources they'll need for their operation. They're registering domains, setting up command and control servers, compromising infrastructure to use as pivots, and developing or acquiring malware.

**Threat Hunting Reality**: Like reconnaissance, this occurs outside our environment. We can't hunt for what happens on adversary-controlled infrastructure. However, threat intelligence about these resources (malicious domains, IP addresses, tool signatures) becomes valuable when we do hunt within our network.

The value here is in consuming intelligence about developed resources and hunting for any interaction our environment has had with them.

---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./02_recon.md" >}})
[|NEXT|]({{< ref "./04_initial.md" >}})

