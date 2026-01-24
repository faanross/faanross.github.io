---
layout: course01
title: "Lesson 1: Interfaces and Factory Functions"
---


## Solutions

- **Starting Code:** [lesson_01_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_01_begin)
- **Completed Code:** [lesson_01_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_01_end)

## Overview

Go's interfaces provide an awesome way for us to implement a generalized feature, while abstracting away specific implementations thereof. This is incredibly useful if a given feature either:

1. Has **multiple** types of implementations, or
2. The **specific** type of implementation might **change** in the future.

In "**general-speak**": this is a **modular** design that allows for both **maintainability** (change something) and **extensibility** (add something).

And in our case we have 2 different generalized features that would benefit from this - both our agent (client) and server. Since we want to allow these two components to communicate to one another using either DNS or HTTPS, this is a perfect application of an interface. Plus, as an added bonus, we can then easily add other protocols in the future without tinkering with our main application code, they are separated via a modular design.

## What We'll Create

- Agent interface in `internals/agent/models.go`
- Server interface in `internals/server/models.go`
- Agent factory function in `internals/agent/models.go`
- Server factory function in `internals/server/models.go`
- Config structs in `internals/config/config.go`
- Agent main entry point in `cmd/agent/main.go`

## Interfaces

The first thing we'll create is both an interface for our `Agent`, as well as `Server`. An interface is just a contract - it's a list of all methods a type has to implement to fulfill that contract.

### Agent Interface

In a new file `internals/agent/models.go` let's create the Agent interface:

```go
package agent

import "context"

// Agent defines the contract for agents
type Agent interface {
	// Send sends a message and waits for a response
	Send(ctx context.Context) ([]byte, error)
}
```

That's it, our Agent just needs one method for now - `Send()`. Now this might seem like we're missing an obvious antipode - `Receive()`. But in most languages, and almost all libraries, you don't really have the ability to send without the ability to receive - they are really baked in together.

So for example in the `net/http` library, we'll typically say something like the following (pseudo-code):

```go
response, error := Send()
```

Meaning the return value of the send function is the reception of the response.

### Server Interface

In a new file `internals/server/models.go` we add:

```go
package server

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

## Factory Functions

Once our application is built, it will essentially function like this - we'll specify what type of agent/server we want in a config (let's say DNS), and then when we start our application, we want it to automatically create the correct types. In this case of course that would be a DNS agent and server.

So we'll get to creating our config system where we specify what we want soon enough, but this mythical ability to create the types we want is called a **factory function**. And it is EXTREMELY simple - it's essentially just a switch statement wrapped in a function. We pass the function a config, it looks at what type of agent/server we want, and then using an internal switch statement it simply calls the correct constructors for the types we want. That's it.

### Agent Factory Function

In `internals/agent/models.go`, let's add our Agent factory function:

```go
import (
	"context"
	"fmt"

	"your-module/internals/config"
)

// NewAgent creates a new agent based on the protocol
func NewAgent(cfg *config.AgentConfig) (Agent, error) {
	switch cfg.Protocol {
	case "https":
		return nil, fmt.Errorf("HTTPS not yet implemented")
	case "dns":
		return nil, fmt.Errorf("DNS not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```

Note that at this point we have not yet created `config.AgentConfig`, so your IDE will likely complain when referencing it. Don't worry, we'll create it soon.

Here we can see the essential logic which is as simple as I alluded to. Note of course that since none of our actual constructors have been implemented, we're not calling them yet, but rather just returning an error in both cases. As we create our actual types and associated constructors we'll replace these lines.

### Server Factory Function

In `internals/server/models.go` we add:

```go
import (
	"fmt"

	"your-module/internals/config"
)

// NewServer creates a new server based on the protocol
func NewServer(cfg *config.ServerConfig) (Server, error) {
	switch cfg.Protocol {
	case "https":
		return nil, fmt.Errorf("HTTPS not yet implemented")
	case "dns":
		return nil, fmt.Errorf("DNS not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```

Now let's go ahead and create the actual config we'll pass to these factory functions so they know which constructor to call.

## Config Structs

Since a struct is a custom type that allows us to create a collection of different types, it's almost always used to represent a config internally in a Go application.

Let's create our configs in `internals/config/config.go`:

```go
package config

import "time"

// AgentConfig holds all configuration values for the agent
type AgentConfig struct {
	ServerIP   string
	ServerPort string
	Timing     TimingConfig
	Protocol   string // this will be the starting protocol
}

// ServerConfig holds all configuration values for the server
type ServerConfig struct {
	ListeningInterface string
	ListeningPort      string
	Protocol           string // this will be the starting protocol
	TlsKey             string
	TlsCert            string
}

// TimingConfig holds timing-related configuration
type TimingConfig struct {
	Delay  time.Duration // Base delay between cycles
	Jitter int           // Jitter percentage (0-100)
}
```

### A Note on Composition

I've created an embedded config called `TimingConfig` inside of `AgentConfig`. Now to be honest, in this situation this was kinda unnecessary - it probably would have made more sense to just have `Delay` and `Jitter` directly inside of `AgentConfig`.

So why did I do it? Really just to show you the pattern and make you aware that you can do it. This might seem trivial, but the ability to embed structs within structs (within structs...) is an incredibly powerful and flexible feature that forms part of a meta-feature called **composition**.

Without getting into too much detail here, I do want to mention a few things:

1. Composition is Go's idiomatic answer to **inheritance**, one of the core "features" of OOP
2. If you want to become a "serious" Go developer, this is a muscle you'll absolutely 100% want to develop

Here are two great references to get you started:
- [This video](https://www.youtube.com/watch?v=hxGOiiR9ZKg) explains why composition is superior to inheritance
- [This video](https://www.youtube.com/watch?v=kgCYq3EGoyE) is a great introduction to composition in Go

## Agent's main

We don't have much to test yet, but we can cobble together a contrived main entrypoint which will help us get a better sense of how the config, factory function, and (eventually) constructors will fit together.

Create a new file `cmd/agent/main.go`:

```go
package main

import (
	"fmt"

	"your-module/internals/agent"
	"your-module/internals/config"
)

func main() {
	agentCfg := config.AgentConfig{
		Protocol: "https",
	}

	_, err := agent.NewAgent(&agentCfg)
	if err != nil {
		fmt.Println(err)
	}
}
```

Since we don't yet have a constructor, we'll just create a struct literal `agentCfg`, and then call our `NewAgent` factory function with this as the sole argument.

We don't yet have any use for the first argument (which will be the agent type once we implement the constructor), so we throw it away using `_`. And of course since we know that our current factory function will return an error regardless of the argument, we expect this line to execute - `fmt.Println(err)`.

## Test

Let's run our Agent and see what happens:

```bash
go run ./cmd/agent
```

**Expected output:**

```bash
HTTPS not yet implemented
```

And indeed of course as we expect we can see the error for HTTPS was returned and then printed to console. So not too exciting, but I just wanted to give you a sense that:

- We are going to specify our desired protocol (`https` or `dns`) in our config
- The config is passed to the Agent's factory function
- Based on the specified protocol a switch statement will execute specific logic

## Conclusion

In this lesson, we've laid a completely logical foundation for building our C2:

- Created `Agent` interface defining the communication contract
- Created `Server` interface defining the server contract
- Implemented factory functions that will create the correct types based on config
- Created config structs to hold our configuration values
- Tested the basic flow from config to factory function

In the next lesson, we'll implement the HTTPS server - our first concrete type that fulfills the Server interface.

---

[Previous: What We'll Build](/courses/course01/what-we-build) | [Next: Lesson 2 - HTTPS Server](/courses/course01/lesson-02) | [Course Home](/courses/course01)
