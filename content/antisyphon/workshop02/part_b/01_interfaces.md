---
showTableOfContents: true
title: "Project Structure and Interfaces"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson01_Begin).
The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson01_Done).

## Overview
As we discussed in the previous lecture, Go's interfaces provide an awesome way for us to implement a generalized feature, while abstracting away specific implementations thereof. This is incredibly useful if a given feature either:
1) Has multiple types of implementations, or
2) The specific type of implementation might change in the future.

In "general-speak": this is a modular design that allows for both maintainability (change something) and extensibility (add something).

And in our case we have 2 different generalized features that would benefit from this - both our agent (client) and server. Since we want to allow these two components to communicate to one another using either DNS or HTTPS, this is a perfect application of an interface. Plus, as an added bonus, we can then easily add other protocols in the future without tinkering with our main application code.


## What We'll Create
- Agent interface (`internals/models/interfaces.go`)
- Server interface (`internals/models/interfaces.go`)
- Config struct (`internals/config/config.go`)
- Agent factory function (`internals/models/factories.go`)
- Server factory function (`internals/models/factories.go`)
- Agent's main entrypoint (`cmd/agent/main.go`)



___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_a/06_golang.md" >}})
[|NEXT|]({{< ref "02_yaml.md" >}})