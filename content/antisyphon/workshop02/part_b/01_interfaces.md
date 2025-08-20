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
1) Has **multiple** types of implementations, or
2) The **specific** type of implementation might **change** in the future.

In "**general-speak**": this is a **modular** design that allows for both **maintainability** (change something) and **extensibility** (add something).

And in our case we have 2 different generalized features that would benefit from this - both our agent (client) and server. Since we want to allow these two components to communicate to one another using either DNS or HTTPS, this is a perfect application of an interface. Plus, as an added bonus, we can then easily add other protocols in the future without tinkering with our main application code,
they are seperated via a modular design.


## What We'll Create
- Agent interface (`internals/models/interfaces.go`)
- Server interface (`internals/models/interfaces.go`)
- Config struct (`internals/config/config.go`)
- Agent factory function (`internals/models/factories.go`)
- Server factory function (`internals/models/factories.go`)
- Agent's main entrypoint (`cmd/agent/main.go`)


## Interfaces

The first thing we'll create is both an interface for our `Agent`, as well as `Server`. And of course, as we've now extensively covered, an interface is just a contract, it's just a list of all methods a type has to implement to fulfill the contract.


### Agent interface

So in a new file `internals/models/interfaces.go` let's create the Agent interface:

```go
// Agent defines the contract for agents
type Agent interface {
	// Send sends a message and waits for a response
	Send(ctx context.Context) ([]byte, error)
}
```

That's it, our Agent just needs one method - `Send()`. Now this might seem like we're missing an obvious antipode - `Receive()`. But in most languages, and almost all libraries, you don't really have the ability to send without the ability to receive - they are really baked in together.

So for example in the `net/http` library, we'll typically say something like the following (pseudo-code):

```go
response, error := Send()
```

Meaning the return value of the send function is the reception of the response. All to say, they're really baked in together, you can't really have one without the other.


### Server interface

In the same file we'll add our Server interface:

```go
// Server defines the contract for servers
type Server interface {
	// Start begins listening for requests
	Start() error

	// Stop gracefully shuts down the server
	Stop() error
}
```

Once again it's pretty simple - we have the ability to `Start()` and `Stop()` our server.


Now that we have our two interfaces we can move towards implementing them - meaning we can create both a HTTPS-, and a DNS-implementation of the methods. However, before we can do that we'll need a custom type for each since what separates a method (i.e. receiver function) from a "normal function" is that it is attached to a type.

So in the upcoming lessons we want to work towards creating our **4 types** - a HTTPS agent and server, and a DNS agent and server.

But before we get to that we want to design our system that will give us the correct type based on what we specify in a config.



___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_a/06_golang.md" >}})
[|NEXT|]({{< ref "02_yaml.md" >}})