---
showTableOfContents: true
title: "Basic Agent Setup and Config System (Lab 03)"
type: "page"
---

## Overview
In this Lab we'll set up the foundation of our agent. Specifically, we'll create 3 components
- **Config system:** where we can easily set different config parameters like target host, port, sleep, jitter etc.
- **Main agent logic**: where our all-important runLoop() function, the "heart" of the agent, will live.
- **Main agent entrypoint**: Where the high-level "orchestrating" of our agent occurs. 

With that, let's get to it.

## internal/agent/config/config.go
We'll create these new directors + file where we'll place our config system. 

The first thing we want to create is our `Config` struct, which is our definition of our custom data type we're going
to use to represent the configuration of any agent.

```go
type Config struct {
	// Server connection details
	TargetHost string
	TargetPort string

	// Connection behavior
	RequestTimeout    time.Duration

	// Operational behavior
	Sleep  time.Duration
	Jitter float64 // As a percentage (0-100)

	// Identity
	AgentUUID string

	//Check-in Endpoint
	Endpoint string
}

```

Now a struct by itself is really just a blueprint. Much like how in OOP we need to instantiate a class to have an 
object we can actually work with, so we also need to instantiate a struct.

There are a few ways to do this, and we'll cover all of them through the course of today's workshop, but the first 
way I want to show you how to create an instance is probably the most idiomatic way - using a constructor function.
Note that it is both similar in some regards, and different in others, that a constructor method in OOP.

So I'll just show you what it looks like, then we can discuss it right afterward.

```go

func NewConfig() *Config {
	return &Config{
		TargetHost: "127.0.0.1",
		TargetPort: "7777",

		RequestTimeout:    60 * time.Second,

		Sleep:  10 * time.Second,
		Jitter: 50.00,
		
		AgentUUID: "",

		Endpoint: "/",
	}

}

```

First thing, it's conventional to name a constructor as "New" + whatever it instantiates. So, here we are instantiating
the `Config` struct, hence `NewConfig()`.

We can then see it's returning a pointer to our instantiated struct, and then inside the function itself it's pretty
straightforward - we give every field a suitable value. Note that for now UUID is blank, we'll generate a value for it
in the following lab.

Also one thing to note is think of these as all being "suitable default" values - our constructor ensures that, at minimum,
our Config struct instance has all the info it needs to run. But we are always to, at a later point, based on for example
user input via the UI, to override any of these values.

Now that we have our config in place we'll set up our agent's core logic.

## internal/agent/agent.go
### Struct and Constructor
The first thing we want to do is, in similar fashion to our config, define what an instance of an Agent will look like
by defining it's struct.

```go
type Agent struct {
	Config *config.Config

	client *http.Client

	stopChan  chan struct{}
	running   bool
	connected bool
}
```

The interesting, and also highly idiomatic, thing to note here is that we are of course embedding our config struct
inside of Agent struct. At its core, embedding a struct in Go is the language's idiomatic approach to 
**"composition over inheritance"**. This is a very interesting design consideration which we don't have time for here,
but I encourage you look it up.


We'll also similarly use a constructor to instantiate our Agent struct.

```go
func NewAgent(config *config.Config) *Agent {

	return &Agent{
		Config: config,
		client: &http.Client{
			Timeout: config.RequestTimeout},
		stopChan:  make(chan struct{}),
		running:   false,
		connected: false,
	}
}
```

The `client` is, in similar fashion to our listener/`server`, an instance we get from the `net/http` standard library.
We can see here we only need a single argument to call it - `RequestTimeout` - which in this case we already defined in our
config. 


One this is done we'll give our agent the two most basic commands - the ability to `Start()`, and to `Stop()`.

### Stop() and Start()

```go
func (a *Agent) Start() error {
	if a.running {
		return fmt.Errorf("agent already running")
	}

	a.running = true

	go a.runLoop()

	return nil
}
```


With `Start()` we'll first check to ensure it's not running yet, if so we'll exit. If not, we'll set it as running, and 
then call another function, a receiver function called `runLoop()`, in its own goroutine. 

Our `Stop()` function will be similar in some regards
```go
func (a *Agent) Stop() error {
	if !a.running {
		return fmt.Errorf("agent not running")
	}

	close(a.stopChan)
	a.running = false

	fmt.Println("Agent stopped")
	return nil
}
```

Now we check it's already not running, and exit if indeed so. If not we call a built-in function called `close()`, and
pass it as argument. This will essentially "trigger" our stopChan, which will serve as a signal for the agent to stop.
How exactly? You'll see that when we build our `runLoop()` function next.


### runLoop()

As I mentioned, this is really where the heart of our agent is, an infinite `for` loop that is going to send requests, 
receive and process responses, sleep, and do it over and over again until it's signalled to stop (or crashes). 


```go

func (a *Agent) runLoop() {
	for {
		select {
		case <-a.stopChan:
			return
		default:
			sleepTime := a.CalculateSleepWithJitter()

			err := a.Connect()
			if err != nil {
				fmt.Printf("Connection error: %v\n", err)
				time.Sleep(sleepTime)
				continue
			}

			resp, err := a.SendRequest(a.Config.Endpoint)
			if err != nil {
				fmt.Printf("Request error: %v\n", err)
				time.Sleep(sleepTime)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Printf("Error reading response body: %v\n", err)
			} else {
				log.Printf("Response: %s\n", string(body))
			}

			time.Sleep(sleepTime)
		}
	}
}
```
The first thing to note is we have a `select` statement, and unlike the one we used in Lab 2, this time it we've introduced
some logic into the fray. You can think of a `select` statement structurally like a `switch` statement, but instead of making 
decision based on input/argument values, we do so based on channel signals. 

And the one here is a very specific form where we only have one real (non-default) case - `stopChan`. Recall that in `Stop()` I said that
when we call `close()` on our stopChan, it will essentially "trigger" it? Well, here we are saying: in case `stopChan` is
ever triggered, do this thing. And what do we do - we break out of the `for{}` loop, thus causing the agent to stop.

And then with this pattern, all the code we want to run continuously when `stopChan` has not yet been called we simply 
place in the "default" case. In other words, all the main operational logic of our agent goes here.

First, we'll calculate `sleepTime` as the return value from a helper method we've not created yet called `CalculateSleepWithJitter()`.
We'll discuss that later.

We'll then connet by calling a Connect(), which also does not exist yet. Ditto for sending a request using SendRequest().

For our response we use the library function `io.ReadAll()`, which can take body of the HTTP response (`resp`), and
functionally convert it from an open stream into a byte slice (`[]byte`) we can work with in Go. We then also call
Close() on it to terminate the stream and release all resources associated with it. 

Once we're done with that we'll sleep, before repeating the whole loop again.

Let's now go build out all the functions we referenced here, but which don't yet exist.

### Connect() and GetTargetAddress()

```go
func (a *Agent) Connect() error {
	url := fmt.Sprintf("http://%s/", a.GetTargetAddress())

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	a.connected = true
	return nil
}
```

We construct the `url` using another helper function `GetTargetAddress()`, we've also yet to create. Here we simply use a 
lightweight HEAD request as a means to force the underlying network stack to establish a TCP connection (if needed).


Let's also take care of our new helper function, which aims to simply make our lives easier by combining our target's IP and
port. But why bother? We'll instead of having to use fmt.Sprintf numerous times each time we need to combine them, we can
just create this function, and call on it to abstract all that away. It's just a bit of tidying, but nothing essential
obviously. 

```go
func (a *Agent) GetTargetAddress() string {
	return fmt.Sprintf("%s:%s", a.Config.TargetHost, a.Config.TargetPort)
}
```

### SendRequest

After we've ensured we're connected we'll send a request:

```go
func (a *Agent) SendRequest(endpoint string) (*http.Response, error) {
	// Check if we're connected
	if !a.connected {
		return nil, fmt.Errorf("not connected to server")
	}

	// Create the full URL
	url := fmt.Sprintf("http://%s%s", a.GetTargetAddress(), endpoint)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add basic headers
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// Send the request
	resp, err := a.client.Do(req)
	if err != nil {
		a.connected = false 
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return resp, nil
}
```

Everything should be well explained by the accompanying comments, just note as well our ability to manually
set our HTTP Headers. In this case we've set `User-Agent`, just to illustrate, but this ability will come in use
in our next lab.

### CalculateSleepWithJitter()

We now get to our final helper function, which will calculate our sleep time as a product of sleep and jitter each time
it's called. 

```go
func (a *Agent) CalculateSleepWithJitter() time.Duration {
	// Apply jitter as a percentage of the base sleep time
	jitterFactor := 1.0 + (rand.Float64() * a.Config.Jitter / 100.0)
	return time.Duration(float64(a.Config.Sleep) * jitterFactor)
}
```

We now have our agent's config and operational logic in place, the only thing left to do now is
create our agent's main entrypoint to orchestrate its execution.


## cmd/agent/main.go

```go
package main

import (
		// imports here 
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize configuration with defaults
	agentConfig := config.NewConfig()

	// Create agent instance
	c2Agent := agent.NewAgent(agentConfig)

	// Start agent
	err := c2Agent.Start()
	if err != nil {
		fmt.Printf("Failed to start agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent started!")
	fmt.Printf("Connected to: %s\n", c2Agent.GetTargetAddress())

	// Wait for termination signal
	<-sigChan

	// Gracefully stop the agent
	fmt.Println("Shutting down agent...")
	c2Agent.Stop()

}
```

Right at the top we'll set up signal handling to allow for graceful shutdown. Now of course in actual practice we 
would not need this since an agent is running on someone else's system to which we don't have access to. 
This is really just to help us now while developing to give us an ability to stop our agent gracefully.

We'll then initialize our `Config` struct by calling the constructor. This allows us to then create our actual agent
by calling its constructor, passing the config as an argument.

We then `Start()` our agent, after which we wait for the termination signal (`SIGTERM`). Once called it will unblock
the main thread, causing the `Stop()` function to be called on the agent.

## Test 

Let's first start our server (`go run ./cmd/server`), and then we can start our agent (`go run ./cmd/agent`).

We'll immediately see that our agent will automatically hit our server's endpoint every 5 - 7.5 seconds, and we'll
periodically see the expected output on both the server and agent console. 

![lab03](../img/lab03a.png)

![lab03](../img/lab03b.png)




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_b/lab02.md" >}})
[|NEXT|]({{< ref "lab04.md" >}})