---
showTableOfContents: true
title: "Agent UUID System and Server Middleware (Lab 04)"
type: "page"
---
## Overview

In our previous lab


Our agent is assigned a UUID (Universally Unique Identifier) to serve as its permanent, unique fingerprint. This is critical because temporary network details like connection IDs or IP addresses are unreliable for tracking; connections drop and IPs change beyond our control. By embedding a fixed UUID within the agent, we ensure a stable identifier that persists across restarts and network changes, allowing us to reliably track and manage this specific agent instance consistently over time.




## the issue
**VERY SIMPLE TO CREATE**

```
agentUUID := uuid.New().String()
```



**CREATES NEW UUID EACH TIME**

**NEED BUILD SYSTEM - RUN, EMBED, COMPILE**

**SIMPLE, BUT A LOT OF EFFORT TO SET THIS UP**


**SO FOR NOW DO THIS 'WRONG' WAY, RIGHT WAY IS IN EXPANDED VERSION**







___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab03.md" >}})
[|NEXT|]({{< ref "../part_d/lab05.md" >}})