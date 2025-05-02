---
showTableOfContents: true
title: "Agent Executes, Returns Result, Server Result Processing (Lab 12)"
type: "page"
---


## Overview
In the previous lab we created a new endpoint (`/command`) and associated handler, which allows our agent to check-in and retrieve any commands, if they are available. We ended the lab by sending the retrieved command back to the agent, where we'll now pick it up again.

So we'll reintegrate some logic allowing our agent to receive and process the command, execute it, and return the result to the server (via a POST request to a new endpoint `/result`). Thereafter the server will process the result, and send it back to our client, which will then display the result to us in the browser.

## internal/agent/agent/agent.go

So let's head back to our agent's `runLoop()`, if you can recall at the start of Lab 11 we stripped most of the the response processing logic, so let's first address that. To ensure we're all on the same page, this is what the function currently looks like:


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

			// CONTINUE HERE!
			// Process response
			resp.Body.Close()

			// Sleep
			time.Sleep(sleepTime)
		}
	}
}
```


And now we add the following logic to the area indicated by the comment above (`// CONTINUE HERE!`).

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

			// NEW LOGIC STARTS HERE -->
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading command response body: %v\n", err)
				continue
			}
			
			var cmdResp struct {
				Command    string `json:"command"`
				HasCommand bool   `json:"hasCommand"`
			}
			
			if err = json.Unmarshal(body, &cmdResp); err != nil {
				log.Printf("Error parsing command response: %v\n", err)
				continue
			}
			
			log.Printf("Command check response: hasCommand=%v, command=%s\n",
				cmdResp.HasCommand,
				cmdResp.Command)
			
			if cmdResp.HasCommand {
				a.executeCommand(cmdResp.Command)
			}

			// TILL HERE -->
			
			time.Sleep(sleepTime)
		}
	}
}
```


We'll use `io.ReadAll` to read the response and assign it to `body`. We create a new struct called `cmdResp`, after which we immediately deserialize `body` into using `json.Unmarshal`.

We then once again print the contents of this struct to the console to help us track the flow of our application, and then it all really culminates with call to `executeCommand` if the command does exist.

Now this will be a very important function that does all the high-level orchestration regarding executing the command, processing the response, and sending it back to the server. So let's start creating it in this same file.



## executeCommand()

So once again, in the same file (agent.go) let's start building out the `executeCommand()` function.

```go
// executeCommand handles command execution and sending results
func (a *Agent) executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)

	// Execute the command
	output, err := commands.Execute(command)

	// MORE TO COME
}
```


The cool thing here is of course that we've already implemented `Execute()` - it's the "command router" function that uses a switch statement to decide which specific function to call. And if you recall, it will ultimately return with the output (i.e. result) of executing the function, as well as a potential error.

So since all of that's done we can immediately continue building our function.

```go
// executeCommand handles command execution and sending results
func (a *Agent) executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)

	// Execute the command
	output, err := commands.Execute(command)

	// Prepare result
	if err != nil {
		output = err.Error()
	}

	// Create result JSON
	result := struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Output  string `json:"output"`
	}{
		Type:    "response",
		Command: command,
		Output:  output,
	}

	// Convert to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v\n", err)
		return
	}

	// Prepare the result before sending back to server as request
	reader := bytes.NewReader(resultJSON)
	
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/result", a.GetTargetAddress()), reader)
	
	if err != nil {
		log.Printf("Error creating result request: %v\n", err)
		return
	}

	// MORE LOGIC TO COME
	
}
```


After we've received the `output` we prepare a new `result` struct containing the results, we serialize it using `json.Marshal()`, and we then prepare it using `bytes.NewReader()` since our data has to be an `io.Reader`, which allows the data to be streamed (read sequentially).

We then *prepare* (but not send) our `POST` request which will contain the result. Note that we are sending it to a new endpoint `/result`. As indicated by my comment, we're not quite done here yet, but I'd like to go implement this new endpoint and associated handler, after which we can circle back here and wrap everything up.

## internal/router/routes.go


So back in our router package we can add our new endpoint.

```go
func SetupRoutes(r chi.Router) {
	r.Use(middleware.UUIDMiddleware)

	r.Get("/command", CommandHandler)
	r.Post("/result", ResultHandler)
}

```

And of course, we're calling a new handler here that does not exist, so let's go and implement it.



## internal/router/handlers.go


So we'll create our `ResultHandler`. I think before we write it, it's useful just to remind ourselves what's going on - we're back in our server, we just received the POST request from the agent containing the `io.Reader` + `serialized` result from running the command. So we want to process it, and then send it to our client by calling another function.


```go
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	// Get the agent UUID from the request context
	agentUUID, _ := r.Context().Value(middleware.AgentUUIDKey).(string)

	log.Printf("Result endpoint hit by agent: %s\n", agentUUID)

	// Parse the incoming result - use Message type directly
	var result websocket.Message
	
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received result from agent %s:\n  Command: %s\n  Status: %s\n  Output: %s\n",
		agentUUID,
		result.Command,
		result.Output)

	// Forward the result to the WebSocket clients
	if websocket.GlobalWSServer != nil {
		// Update the Message with required fields
		result.Type = websocket.ResponseMessage
		result.AgentUUID = agentUUID


		// Broadcast this message directly
		websocket.GlobalWSServer.Broadcast(result)

		log.Printf("Broadcasting result from agent %s to all clients\n", agentUUID)
	}

	// Send acknowledgment
	w.WriteHeader(http.StatusOK)
}

```

The first thing to note is this time we're not creating a struct to deserialize into, but a `websocket.Message`, since of course we'll seen be sending it to our client over websocket.

Right after that we then deserialize the JSON, and print the results to screen.

Then we run a check to ask - is there a global Singleton instance of our Websocket Server? If you recall in Lab 06, we created GlobalWSServer and then assigned it a value when `StartWebSocketServer()` is called. So the conditional is contingent upon our websocket server running.

Inside of it we assign two of the fields (`Type` and `AgentUUID`), and then we call the `Broadcast()` function to send our result to our client. Now this function does not exist yet, so we'll get to it shortly.

Before we do we can also send an acknowledgement to our agent that we successfully received the result with `w.WriteHeader`. We're not doing anything with in this case, but we could of course create a conditional on the agent side, that informs us that the server successfully processed the response.

Let's move onto the `Broadcast()` function.

## internal/websocket/wss.go

I just want you to be aware that you will ultimately want to type of functions to send data to client(s) - a broadcast function which sends to all connected clients, and a "send message" function, which will only respond to one specific client. Why do you need both? Well obviously sometimes there are events you want all clients to be aware of - like a new agent that was deployed.

But then at other times, perhaps a specific client enquired as to a command history in the database, then in that case you want to send the result of that specific query back to only the client that asked for it.

In any case, at present, since we only have one client they'd be doing the same thing, I just opted for broadcast because it's a relatively simpler function to write.

So let's go ahead and create, but before we do, we have to address an issue we have. Let's look at our current websocket server struct.


```go
// WebSocketServer represents a simple WebSocket server
type WebSocketServer struct {
	port     int
}
```


Right now we only have one field - the port. So we have no instance of any connected client, meaning we're not able to send anything to it. So what we need to do is add a new field capable of "holding" an instance of our client.


```go
// WebSocketServer represents a simple WebSocket server
type WebSocketServer struct {
	port     int
	client   *websocket.Conn
	clientMx sync.Mutex
}
```


So we'll add `client`, which is a pointer to a websocket connection instance. Note that if you want this struct to be able to keep instances of multiple clients, you should use a slice of pointers here.

And then once again we'll add a mutex, not really required since we're only allowing a single client to be saved at present, but just some shoes to grow into.

And along with adding these fields to the actual struct definition, we now of of course also need to initialize the field in our constructor.

```go
func NewWebSocketServer(port int) *WebSocketServer {
	return &WebSocketServer{
		port:    port,
		client: nil,
	}
}
```

Great, so now we're capable of storing an instance of the websocket connection in our websocket server struct, but this of course does not mean it will just happen automatically.

So let's add this logic inside of our handleWebSocket function. Note I'm not going to reproduce the entire function here, but only up until we log the message to console, it currently looks like this:

```go
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrader failed to upgrade: %v", err)
		return
	}
	defer conn.Close()

	fmt.Println("WebSocket connection established")

	// Rest of function here

}
```

And we'll now add new logic above the print statement, and above that we're going to change the current defer `conn.Close()` by wrapping it in an anonymous function.

```go
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrader failed to upgrade: %v", err)
		return
	}
	defer func() {
		conn.Close()

		s.clientMx.Lock()
		s.client = nil
		s.clientMx.Unlock()
	}()

	s.clientMx.Lock()
	s.client = conn
	s.clientMx.Unlock()

	fmt.Println("WebSocket connection established")

	// Rest of function here

}
```


You can see that on the bottom we're adding the newly connected client with `s.client = conn`, and then in our anonymous function we make sure to remove any clients from the `client` field using `s.client = nil` once we disconnect.


## Broadcast()

So now that we have access to a connected client via the client field in our `WebSocketServer` struct, let's create our `Broadcast() `function. We'll also create this in `internal/websocket/wss.go`.


```go
func (s *WebSocketServer) Broadcast(msg Message) {
	s.clientMx.Lock()
	defer s.clientMx.Unlock()

	if s.client != nil {
		err := s.client.WriteJSON(msg)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			return
		}
	}
}
```

So it's simple enough - we use a mutex lock once again, and then if there is a client (`s.client != nil`), then we simply serialize and send `msg`, which is the argument passed to the function.


That's it, if you recall this function was called from our `ResultHandler()` function, which in turn is of course called when the `/results `endpoint is hit. So now everything on the server is essentially, we just need to quickly head back to our agent to finalize our `executeCommand()` function.


## internal/agent/agent/agent.go

Our function current looks like this:

```go
// executeCommand handles command execution and sending results
func (a *Agent) executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)

	// Execute the command
	output, err := commands.Execute(command)

	// Prepare result
	if err != nil {
		output = err.Error()
	}

	// Create result JSON
	result := struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Output  string `json:"output"`
	}{
		Type:    "response",
		Command: command,
		Output:  output,
	}

	// Convert to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v\n", err)
		return
	}

	// Prepare the result before sending back to server as request
	reader := bytes.NewReader(resultJSON)
	
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/result", a.GetTargetAddress()), reader)
	
	if err != nil {
		log.Printf("Error creating result request: %v\n", err)
		return
	}

	// MORE LOGIC TO COME
	
}
```


Where the comment indicates, we can now add the following:

```go
// executeCommand handles command execution and sending results
func (a *Agent) executeCommand(command string) {
	log.Printf("Executing command: %s\n", command)

	// Execute the command
	output, err := commands.Execute(command)

	// Prepare result
	if err != nil {
		output = err.Error()
	}

	// Create result JSON
	result := struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Output  string `json:"output"`
	}{
		Type:    "response",
		Command: command,
		Output:  output,
	}

	// Convert to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v\n", err)
		return
	}

	// Prepare the result before sending back to server as request
	reader := bytes.NewReader(resultJSON)
	
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/result", a.GetTargetAddress()), reader)
	
	if err != nil {
		log.Printf("Error creating result request: %v\n", err)
		return
	}

	// MORE LOGIC FOLLOWS
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-ID", a.Config.AgentUUID)

	// Send the request
	client := &http.Client{Timeout: a.Config.RequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending result: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// At the end of executeCommand
	log.Println("Command execution complete!")
	
}
```

So we add our headers, including our AgentUUID, we then send the actual POST request using `client.Do`, and then at the bottom we add a final confirmation statement to indicate that we're done.

That is it, we're all connected now, so let's go ahead and test.


## Test
So run the client, run the server, and run your agent. You should now be able to run any of the commands and then, following a short wait (agent sleep), the result will appear on your client.

![lab12](../img/lab12a.png)

Congrats! Though the actual functionality is rudimentary, the pattern is there. The groundwork has been laid and there is a lot of potential to build on this, which I'll discuss in our concluding section next.







___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab11.md" >}})
[|NEXT|]({{< ref "../part_f/review.md" >}})