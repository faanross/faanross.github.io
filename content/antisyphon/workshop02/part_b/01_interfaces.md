---
showTableOfContents: true
title: "Project Structure and Interfaces"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson01_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson01_Done).

## Overview
Go's interfaces provide an awesome way for us to implement a generalized feature, while abstracting away specific implementations thereof. This is incredibly useful if a given feature either:
1) Has **multiple** types of implementations, or
2) The **specific** type of implementation might **change** in the future.

In "**general-speak**": this is a **modular** design that allows for both **maintainability** (change something) and **extensibility** (add something).

And in our case we have 2 different generalized features that would benefit from this - both our agent (client) and server. Since we want to allow these two components to communicate to one another using either DNS or HTTPS, this is a perfect application of an interface. Plus, as an added bonus, we can then easily add other protocols in the future without tinkering with our main application code,
they are seperated via a modular design.


## What We'll Create
- Agent interface (`internals/models/interfaces.go`)
- Server interface (`internals/models/interfaces.go`)
- Agent factory function (`internals/models/factories.go`)
- Server factory function (`internals/models/factories.go`)
- Config struct (`internals/config/config.go`)
- Agent's main entrypoint (`cmd/agent/main.go`)


## Interfaces

The first thing we'll create is both an interface for our `Agent`, as well as `Server`. An interface is just a contract, it's just a list of all methods a type has to implement to fulfill the contract.


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

Meaning the return value of the send function is the reception of the response. 


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


## Factory Function

Once our application is built, it will essentially function like this - we'll specify what type of agent/server we want in a config file (let's say DNS), and then when we start our application, we want it to automatically create the correct types. In this case of course that would be a DNS agent and server.

So we'll get to creating our config system where we specify what we want soon enough, but this mythical ability to create the types we want is called a factory function. 
And it is EXTREMELY simple - it's essentially just a switch statement wrapped in a function. We pass the function a config, it looks at what type of agent/server we want, and then using an internal switch statement it simply calls the correct constructors for the types we want. That's it.

### Agent Factory Function

So in `internals/models/factories.go`, let's first create our Agent factory function:

```go
// NewAgent creates a new communicator based on the protocol
func NewAgent(cfg *config.Config) (Agent, error) {
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

Note that of course at this point we have not yet created `config.Config`, so you'll IDE will likely put up a fuss when referencing it. Don't worry, we'll attend to it soon enough.

So here we can see the essential logic which is as simple as I alluded to. Note of course that since none of our actual constructors have been implemented, we're not calling them yet, but rather just returning an error in both cases. As we create our actual types and associated constructors we'll replace these lines.


### Server Factory Function

The Server's factory function is the exact same logic:

```go
// NewServer creates a new server based on the protocol
func NewServer(cfg *config.Config) (Server, error) {
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


## Config struct

Since a struct is a custom type that allows us to create a collection of different types it's almost always used to represent a config internally in a go application.

So let's go ahead and create our Config in `internals/config/config.go`

```go
// Config holds all application configuration
type Config struct {
	ClientAddr string       
	ServerAddr string       
	Timing     TimingConfig 
	Protocol   string        // this will be the starting protocol
	TlsKey     string       
	TlsCert    string       
}

type TimingConfig struct {
	Delay  time.Duration   // Base delay between cycles
	Jitter int            // Jitter percentage (0-100)}
}
```


As you can probably deduce, we're just using one config for both our agent and server. At the moment, and for the purpose of this workshop it works, but if you wished to a be a bit more efficient, explicit, or as your needs evolve based on project complexity you might want to implement separate configs for each of them.

Also note that I've created an embedded config called `TimingConfig` inside of `Config`. Now to be honest, in this situation this was kinda unnecessary - it probably would have made more sense to just have `Delay` and `Jitter` directly inside of `Config`.

So why did I do it? Really just to show you the pattern and make you aware that you can do it. This might seem trivial, but the ability to embed structs within structs (within structs...) is an incredibly powerful and flexible feature that forms part of a meta-feature called **composition**.

Without getting into too much detail here, I do want to mention a few things. First, composition is Go's idiomatic answer to **inheritance**, one of the core "features" of OOP. And second, if you want to become a "serious" Go developer, this is a muscle you'll absolutely 100% want to develop.

Here are two great references to get you started. First, [this video](https://www.youtube.com/watch?v=hxGOiiR9ZKg) will go over why composition is superior to inheritance, while [this video](https://www.youtube.com/watch?v=kgCYq3EGoyE) is a great introduction to composition in Go.

That's pretty much it for this first lesson, we've really laid a completely logical foundation to create the rest of our application in an idiomatic and efficient manner.

We don't really have much to test yet, that being said we can cobble together a contrived main entrypoint for either the agent or server (I'll just pick agent in this case), which I think will help us just get somewhat of a better sense of how the config, factory function, and (eventually) constructors will fit together.


## Agent's main

Let's create a new file in `./cmd/agent/main.go`

```go
func main() {
	agentCfg := config.Config{
		Protocol: "https",
	}

	_, err := models.NewAgent(&agentCfg)
	if err != nil {
		fmt.Println(err)
	}

}
```


Since we don't yet have a constructor, we'll just create a struct literal `agentCfg`, and then call our `NewAgent` factory function with this as the sole argument.

We don't yet have any use for the first argument (which will be the agent type once we implement the constructor), so we throw it away using `_`. And of course since we know that our current factory function will return an error regardless of the argument, we expect this line to execute - `fmt.Println(err)`.


## Test

So let's run our Agent and see what happens:

```shell
❯ go run ./cmd/agent
HTTPS not yet implemented
```

And indeed of course as we expect we can see the error for HTTPS was returned and then printed to console. So not too exciting, but I just wanted to give you a sense that:
- We are going to specify our desired protocol (`https` or `dns`) in our config
- The config is passed to the Agent's factory function
- Based on the specified protocol a switch statement will execute specific logic

That's it for now, in the next lesson we'll create a more user-friendly implementation of a config system using YAML.





___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "00_setup.md" >}})
[|NEXT|]({{< ref "02_yaml.md" >}})