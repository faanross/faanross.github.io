---
showTableOfContents: true
title: "Agent Retrieves Command From Queue (Lab 11)"
type: "page"
---

## Overview

Our command is now in the queue on our server, ready to be retrieved by our agent. By how exactly does our agent do that? Earlier we created a single endpoint for our agent, every time it awakens from its slumber, it connects (if not yet connected), and then it sends a request to the root endpoint.

At present, we just added some simple placeholder functionality - when the agent hits the endpoint, we get a message on both the server and the agent side just indicating that it was able to do so. This obvs server little purpose other than indicating that our agent is able to do so.

So now, we'll build on that. Instead of just printing something to console, when our agent hits the endpoint each round the associated handler will call another function which:
- Looks in the queue to see if there is a command waiting.
- If no -> It returns false + no command
- If yes -> It removes the command from the queue and returns the command + true

So that's how we go from having the command in the queue on the server, to having the command on the side of the agent. So let's now go and build that out in this lab.


## internal/agent/agent/agent.go

Back in `runLoop()`, if you have not yet removed the called to `CommandsTest()` at the bottom you can do so now, remember to also remove the actual function. This was of course only included to test and make sure this functionality works, but we no longer need it.

Let's review what `runLoop()` looks like at present.
```go
func (a *Agent) runLoop() {
	for {
		select {
		case <-a.stopChan:
			return
		default:
		sleepTime := a.CalculateSleepWithJitter()

			// Connect
			err := a.Connect()
			if err != nil {
				log.Printf("Error connecting to agent: %v\n", err)
				time.Sleep(sleepTime)
				continue
			}

			// Send request
			resp, err := a.SendRequest(a.Config.Endpoint)
			if err != nil {
				log.Printf("Error sending request: %v\n", err)
				time.Sleep(sleepTime)
				continue
			}

			// Process response
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Printf("Error reading response body: %v\n", err)
			} else {
				log.Printf("Response body: %s\n", string(body))
			}

			// Sleep
			time.Sleep(sleepTime)


		}
	}
}
```

At the top we have our call to the `Connect()` function - this remains unchanged.

Then we have our call to `SendRequest()`. Here the logic stays the same, but instead of reaching out to the root endpoint, I'd like to change it to something a bit more explicit like `/command`.

However, we can see of course that we do not define it here, but in our config. So back in `config.go` file, inside of our constructor, let's change the value of `Endpoint`.


**SO THIS CHANGES IN CONFIG.GO**
```go
func NewConfig() *Config {
	return &Config{
		TargetHost:     "127.0.0.1",
		TargetPort:     "7777",
		RequestTimeout: 60 * time.Second,
		Sleep:          5 * time.Second,
		Jitter:         50.00,
		AgentUUID:      "",
		Endpoint:       "/command",
	}
}

```

Now back in runLoop(), we want to change the processing of our response considerably, however we'll wait till the next lab to do that. So for now just change the code from this.

```go
			// Process response
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Printf("Error reading response body: %v\n", err)
			} else {
				log.Printf("Response body: %s\n", string(body))
			}
```

To this.

```go
			// Process commands response
			resp.Body.Close()
```

Now this does not really make sense to do this, but we have to use our `resp` variable, since as I've mentioned in Go we always have to use it. So consider this just a "dangling thread" that we can leave like this for now since it won't affect our immediate goal.

Great, so we have this new route `/command`, but of course we have not created it yet, so let's do so now.

## internal/router/routes.go

So back in routes.go, simply change the endpoint and handler name of our existing route to reflect its updated functionality.

```go
func SetupRoutes(r chi.Router) {
	r.Use(middleware.UUIDMiddleware)

	r.Get("/command", CommandHandler)
}

```


So now we can go and create our new handler `CommandHandler`.



## internal/router/handlers.go

We'll also repurpose the existing RootHandler - change it's name, and then remove the last line we used to write to the response stream ("I'm Mister Derp!").

```go
func CommandHandler(w http.ResponseWriter, r *http.Request) {

	agentUUID, _ := r.Context().Value(middleware.AgentUUIDKey).(string)

	log.Printf("Endpoint %s has been hit by agent %s: \n", r.URL.Path, agentUUID)


}
```

And now we'll add some new logic. First thing, remember I said in the beginning of the lab that our agent hits an endpoint, which then calls a handler, which in turn calls a function that will look if there is a command in the lab and return it if so.


```go
func CommandHandler(w http.ResponseWriter, r *http.Request) {

	agentUUID, _ := r.Context().Value(middleware.AgentUUIDKey).(string)

	log.Printf("Endpoint %s has been hit by agent %s: \n", r.URL.Path, agentUUID)

	// Check if we have a command
	cmd, exists := websocket.AgentCommands.GetCommand()

	// MORE CODE WILL COME

}
```

So we're calling a new function (`GetCommand()`) which will return 2 things - the command, and a boolean indicating whether or not there is a command. Note in the case there is not, the cmd will just be an empty string.

Now we have more logic we'll add there to actually send the response back to the agent, but before we write that let's go and implement our `GetCommand()` function.

## internal/websocket/command_handler.go

Let's now create the function tasked with extracting the command from our queue.

```go
func (cq *CommandQueue) GetCommand() (string, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.PendingCommands) == 0 {
		return "", false
	}

	// Get the first command in the queue
	cmd := cq.PendingCommands[0]

	// Remove it from the queue
	cq.PendingCommands = cq.PendingCommands[1:]

	log.Printf("Command retrieved: %s\n", cmd)

	return cmd, true
}
```

Right at the top we'll once again use a mutex lock to ensure thread safety.

The first thing we then do is check if the queue is empty (`==0`), if it is we return an empty string and `false` - meaning we do not have anything in the queue.

If we get past this check then of course it means we do have a command in the queue, in which we case we:
1. Assign `cmd` equal to the first value in our queue (`PendingCommands[0]`), and
2. Remove that command from our queue (`cq.PendingCommands[1:]`)

We then just print to console to once again help us follow the flow of execution in the console, and return our `cmd` and `true` (as in we do have a command).

Let's now return back to our handler from where we called this function.


## internal/router/handler.go

Now that we've constructed `GetCommand()`, we can complete the rest of our function.


```go
func CommandHandler(w http.ResponseWriter, r *http.Request) {

	agentUUID, _ := r.Context().Value(middleware.AgentUUIDKey).(string)

	log.Printf("Command endpoint hit by agent %s\n", agentUUID)

	// Check if we have a command
	cmd, exists := websocket.AgentCommands.GetCommand()

	// Prepare response struct
	response := struct {
		Command    string `json:"command,omitempty"`
		HasCommand bool   `json:"hasCommand"`
	}{
		HasCommand: exists,
	}
	
	// add cmd to response struct if it exists
	if exists {
		response.Command = cmd
		log.Printf("Found command: %s\n", cmd)
	} else {
		log.Printf("No commands available\n")
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
```

After returning, we'll use the values from `GetCommand()` to construct a `response` struct. We'll then serialize and send it back to our agent using `json.NewEncoder(w).Encode(response)`.

That's it for this part, our response should now be received by the agent. However since we eviscerated our response processing logic at the start of this lab we won't be able to do anything it with it yet.

For now let's run a test to ensure our GetCommand() function is capable of retrieving the command prior to sending it back to our agent.


## Test

So run the client, server, the agent. And now select any command from the frontend, this should now show us both "Command retrieved" (from `GetCommand()`) and "Found command" (from `CommandHandler()`).

![lab11](../img/lab11a.png)





___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab10.md" >}})
[|NEXT|]({{< ref "lab12.md" >}})