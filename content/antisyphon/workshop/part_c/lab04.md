---
showTableOfContents: true
title: "Agent UUID System and Server Middleware (Lab 04)"
type: "page"
---
## Overview

In our previous lab we created the foundational logic for our agent. We'll now furnish with a UUID, which stands for
**Universal Unique IDentifier**. A UUID is just a standardized 128-biy alphanumeric string that can be used to identify information.

In our case we want to ensure each of our agents is compiled with a UUID since it's crucial for us to be able to
keep track of each individual agent running on target systems. We can't use a unique connection ID, since losing and 
regaining a connection is expected. And if we decide to later employ "true" beaconing, well then we are intentionally
disconnecting and reconnencting between every round of request + response.

So you might think we could then just the target host IP to keep track of our agent, but even that can, and often does change.
It's something that can change that's not within our control. And so we have to base our tracking (and thus ability to manage)
on something which is within our control - hence statically compiling each agent with a UUID.

## But, We Have an Issue
First, I want to show just how easy it is to generate a UUID using the `uuid` library. We literally just run the
following command.

```
agentUUID := uuid.New().String()
```

And so that's awesome, you might think great, we just add that line to our agent's `main.go` entrypoint and ensure
we assign this before we call the Agent constructor, and we're a gtg. And that seems like a fine idea, but...

The issue is that this function is going to run, and generate a new UUID **every time** our program starts. And so assuming
we create some form of persistence (as we of course should), and at some point the host system reboots and thus restarts our 
process, well then we'll get a brand new UUID assigned to the agent. This defeats the entire purpose. 

So what we really want to do instead is create a Go build system so that:
- This function runs by itself generating a UUID.
- That UUID is then embedded in the agent source code.
- The agent is then statically compiled with the UUID that was generated and embedded in the previous steps.

Such a system is not complex, but it's actually a surprising amount of coding involved. And so it was, to cut to the chase,
another piece of fat I had to trim to get this workshop down to 4 hours. I will include that original lecture in the
Director's Cut, but for now we are going to simply use the function to generate UUID on the fly.

Now in this developmental phase it will work fine, since we are not really running an agent for extended periods on 
a target machine. At least not yet. Doing it this way will server our current purpose just fine, but, I wanted you
to at least be aware of this shortcoming so that you can address it when it becomes important. 

## cmd/agent/main.go

First thing, let's jump into our terminal and add the required library:
```shell
go get github.com/google/uuid
```

And now in the agent's main.go we'll add this line right after we've called the Config's constructor, but before
we pass that instantiated struct as the argument to the Agent's constructor.

```go
func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	agentConfig := config.NewConfig()

	agentConfig.AgentUUID = uuid.New().String()
	
	fmt.Printf("Agent UUID: %s\n", agentConfig.AgentUUID) // Also add this just to confirm it works

	c2Agent := agent.NewAgent(agentConfig)
```

Note I also added a print statement so we can confirm the UUID on the agent's side matches up to that on the server
side.

## internal/agent/agent/agent.go

So we're now generating a UUID, but of course the entire value of having it is being able to communicate it to the server.
So we need to now also include it when we communicate with the server. There are a few ways we could do this: we could
add it as a URL parameter to our GET request (`/?UUID=XXXXX`), we could include it in a POST body, or in this case we can
simply add it as a custom HTTP Header.

Since we want to include it in our request, let's locate the `SendRequest()` function in `internal/agent/agent/agent.go`.

```go
// SendRequest sends a request to the C2 server and returns the response
func (a *Agent) SendRequest(endpoint string) (*http.Response, error) {

	// OTHER CODE HERE...
	
	// Add basic headers
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// Add the Agent UUID as a custom header
	req.Header.Set("X-Agent-ID", a.Config.AgentUUID) // ADD THIS

```

And as you can see we've already set our User-Agent header before, I included this to show you just how simple it is
to set HTTP headers in Go. And now as you can see on the final line in the block above, we'll now also set a new header
we're naming `X-Agent-ID` equal to the UUID we generated. 

And so we are now both generating a UUID and sending it to the server. So now we need to ensure our server expects it,
and knows what to do with it.

## internal/middleware/middleware.go

To do this we are going to create a new file here - `internal/middleware/middleware.go`. If you're new to middleware -
it a component that is capable of intercepting a request before it hits the endpoint. It can then do any number of things,
like inspect certain values, parse values, change values etc. It has many uses, including of course **authentication**,
something we include in the Director's Cut. 

For now however we just want the give our middleware the ability to parse the UUID, save it as a contextkey, attach it to
the original (unaltered) request, and then send it on its way again.

Let's do that.

```go
package middleware

// Key for storing agent UUID in request context
type contextKey string
const AgentUUIDKey contextKey = "agentUUID"

// UUIDMiddleware extracts the agent UUID from headers
func UUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract the UUID from the header
		agentUUID := r.Header.Get("X-Agent-ID")

		// Add the UUID to the request context
		ctx := context.WithValue(r.Context(), AgentUUIDKey, agentUUID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
```

I've added explanatory comments above so you know what happens at every step. The important to know is that once this 
function is done, the request is back on its way to its original destination (target endpoint), but now this time it will
be sent with the `agentUUID` as context.


## internal/router/routes.go

We have middleware that contains the logic to extract the UUID, but we've not yet instructed our server to use it. We'll
do so right here in `routes.go`.

With `Chi` we have amazing control, we can configure middleware to attach to a specific endpoint, a group of specific
endpoints, or to all the endpoints. Since we always want a request to be tied to a specific agent, we'll apply
this middleware to all endpoints.

```go
func SetupRoutes(r chi.Router) {
    // Apply the middleware to all routes
    r.Use(middleware.UUIDMiddleware)
    
    // Register routes
    r.Get("/", RootHandler)
}
```

Easy as that. Now final thing - let's give our handler the ability to print the UUID to console, we'll then be able
to visually confirm that the server is correctly identifying the agent communicating with it in our test.

## internal/router/handlers.go

```go
func RootHandler(w http.ResponseWriter, r *http.Request) {

	agentUUID, _ := r.Context().Value(middleware.AgentUUIDKey).(string)

	log.Printf("Endpoint %s hit by agent %s\n", r.URL.Path, agentUUID)// EDIT THIS LINE!

	w.Write([]byte("I'm Mister Derp!"))
}
```

First, we call a function to return the `agentUUID` from the context we generated in our middleware. And then
we've also modified our `log.Printf` statement so that it now also prints the `agentUUID` to console.


## Test
We can run our agent and confirm it's generating + displaying a UUID to console.

![lab04](../img/lab04a.png)

And then once we run our server, we can see that when our agent hits the endpoint its able to parse and display the 
correct UUID. Big success!

![lab04](../img/lab04b.png)

___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab03.md" >}})
[|NEXT|]({{< ref "../part_d/lab05.md" >}})