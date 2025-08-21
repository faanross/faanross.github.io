---
showTableOfContents: true
title: "HTTPS Server"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson03_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson03_Done).

## Overview
In the next 3 lessons we'll create our complete HTTPS logic, including the server (this lesson), agent (Lesson 4), and run loop (lesson 5). Thereafter we'll do the exact same thing for DNS.

For our HTTPS server: we'll obviously use a struct to represent an instance of a HTTPS server. We'll then also create an accompanying constructor to instanstandtiate it - it's of course this exact constructor that we'll then call from the factory function.

Additionally, if we can recall from our first lesson - we created a Server interface (aka "contract"), that had two methods - `Start()` and `Stop()`. So of course, in addition to our Server struct and constructor, we'll need to implements these, as well as a handler so that our server can actually "do something" once our agent connects to it.

That's about it, let's get cracking.


## What We'll Create
- HTTPS server (`internals/https/server_https.go`)
- Server's main entrypoint (`cmd/server/main.go`)






___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_b/02_yaml.md" >}})
[|NEXT|]({{< ref "02_https_agent.md" >}})