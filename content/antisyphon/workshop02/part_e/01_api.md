---
showTableOfContents: true
title: "Implement API Switch"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson09_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson09_Done).



## Overview
In this lesson we want to create a new endpoint on a new port that, when we hit it, indicates that we want to transition from one protocol to another. It's extremely simple, we don't have to convey any information, just the mere fact that we hit the endpoint is interpreted to mean: "transition from the current protocol to the other protocol".


## What We'll Create
- Control API (`./internals/control/control_api.go`)




## Global Flag System

There are numerous ways for us to implement this system. Here I'll opt for what's probably the simplest - which is just a global Boolean flag. Just imagine conceptually that this flag is `false` by default, and if we hit a specific endpoint using a specific method it'll change to `true`.

Now, as we saw before, usually our server will just always respond with either `false` (for HTTPS) or `42.42.42.42` (for DNS). It's currently hardcoded, there's no consideration to follow an alternative avenue.

But imagine that instead of just responding with the given value it first checks this global flag. BTW, global in this context means "accessible from anywhere in the application", in other languages it's sometimes also termed as being `public`.

It checks the global flag, if the flag is `false` (i.e. we did not hit the endpoint) it will indeed just response with `false`/`42.42.42.42`.

But, if the flag is `true` (i.e. we did hit the endpoint signalling our desire to change protocol) then our server will instead response with `true` (for HTTPS) or `69.69.69.69` (for DNS).

In this lesson we're implementing a mechanism to allow us to trigger a new type of response from our server to the agent. Then, all we need to do in the remaining lessons is ensure that our agent can take variable actions based on that information.

One final thing, one additional layer of nuance, we need to consider is this: I just said if we hit an endpoint the flag changes from `false` to `true`, and if the server sees that the flag is `true` it will send the "change" response to the agent.

But when we hit the endpoint, we don't want it to continuously change back and forth, we want it to only change a single time. Then, perhaps later if we hit the endpoint again, we want it to change again. But we don't want it either be stuck in "don't change" versus "change each time" mode.

In other words, we want to hit the endpoint, we want the flag to change to `true`, but then if the server sees it's `true` and changes its response to the agent accordingly, the flag should of course reset to `false`. This is known as a "consume once" pattern, and is very simple: if the server detects the flag is `true`, change it's response to the agent AND reset the flag to `false`.


## Create our Control API

Let's start implementing all the logic we just discussed in a new file  `./internals/control/control_api.go`


First thing, let's create our global flag. Now this could just be a boolean, but a much better practice would be to create a struct so that we could pair a boolean with a mutex.

Let's define the `struct`, and instantiate a global instance of it by capitalizing the name.

```go
// TransitionManager handles the global transition state
type TransitionManager struct {
    mu             sync.RWMutex
    shouldTransition bool
}

// Global instance
var Manager = &TransitionManager{
    shouldTransition: false,
}
```



Next we'll create a method that will change the value from false to true. That's all it does, and of course this is the method that will be called by the handler when we hit our endpoint.


```go
// TriggerTransition sets the transition flag
func (tm *TransitionManager) TriggerTransition() {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tm.shouldTransition = true
    log.Printf("Transition triggered")
}
```


We now need our method that our server can call to:
1. Check the value of `shouldTransition`,
2. Reset the value to `false` if it is `true`.


```go
// CheckAndReset atomically checks if transition is needed and resets the flag
// This ensures the transition signal is consumed only once
func (tm *TransitionManager) CheckAndReset() bool {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    if tm.shouldTransition {
        tm.shouldTransition = false  // Reset immediately
        log.Printf("Transition signal consumed and reset")
        return true
    }
    
    return false
}
```


Let's create a simple HTTP server (we'll use port 8080) that will expose an endpoint for us to hit to call the `TriggerTransition()` method.


```go
// StartControlAPI starts the control API server on port 8080
func StartControlAPI() {
	http.HandleFunc("/switch", handleSwitch)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}
```

As you can see I've chosen the endpoint `/switch`. Further, we're not actually calling `TriggerTransition()` here, but as with all endpoints we're calling a handler `handleSwitch`, which will be tasked with calling the method in turn.


Last thing, let's implement our handler `handleSwitch`:



```go
func handleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	Manager.TriggerTransition()

	response := "Protocol transition triggered"

	json.NewEncoder(w).Encode(response)
}
```


I've arbitrarily choses to limit the endpoint to POST method requests only just to show you how to do this, but of course we could leave this out and allow the use of any method.

So our Control API Client is complete, but now of course we need to rewrite both our HTTPS and DNS handlers so that instead of being hardcoded to response `false`/`42.42.42.42`, they should use the `CheckAndReset()` method we created.




## DNS Handler Changes

Let's first change our DNS Handler. Right now we have this single line that just says - always respond with `42.42.42.42`:

```go
A: net.ParseIP("42.42.42.42"),
```


So we'll remove this, and then make a few changes to allow for conditional logic:
```go
// handleDNSRequest is our DNS Server's handler
func (s *DNSServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create response message
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	// Process each question
	for _, question := range r.Question {
		// We only handle A records for now
		if question.Qtype != dns.TypeA {
			continue
		}

		// Log the query
		log.Printf("DNS query for: %s", question.Name)

		// NEW LOGIC STARTS HERE
		shouldTransition := control.Manager.CheckAndReset()
		var responseIP string
		if shouldTransition {
			responseIP = "69.69.69.69"
			log.Printf("DNS: Sending transition signal (69.69.69.69)")
		} else {
			responseIP = "42.42.42.42"
			log.Printf("DNS: Normal response (42.42.42.42)")
		}

		// Create the response with the appropriate IP
		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP(responseIP), // <-- Using variable instead of hardcoded
		}
		m.Answer = append(m.Answer, rr)
	}

	// Send response
	w.WriteMsg(m)
}
```


We create `shouldTransition`, which receives the return value of our exposed method `CheckAndReset()`. We then use an `if`/`else` statement to assign, based on the value of `shouldTransition` either `69.69.69.69` (if `true`), or `42.42.42.42` (if `false`) to the variable `responseIP`.

And then when we get to A, we use this variable, instead of a hardcoded value. EZ PZ.


## HTTPS Handler Changes

The change to our HTTP Handler is even simpler since we can just directly assign the value of the global flag to the field in our JSON. This is of course because we are sending a boolean value, and it should be the same as the flag - if the flag value is `true`, our `Change` field is equal to `true`; and vice-versa.

So replace this:
```go
	// Create response with change set to false
	response := HTTPSResponse{
		Change: false,
	}

```

With this:
```go
	// Check if we should transition
	shouldChange := control.Manager.CheckAndReset()
	response := HTTPSResponse{
		Change: shouldChange,
	}
	if shouldChange {
		log.Printf("HTTPS: Sending transition signal (change=true)")
	} else {
		log.Printf("HTTPS: Normal response (change=false)")
	}
```


For posterity's sake I'll include the entire function here:

```go
func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	// Check if we should transition
	shouldChange := control.Manager.CheckAndReset()
	response := HTTPSResponse{
		Change: shouldChange,
	}
	if shouldChange {
		log.Printf("HTTPS: Sending transition signal (change=true)")
	} else {
		log.Printf("HTTPS: Normal response (change=false)")
	}
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
```



## Start Control API Server in Server's main

Finally, we just need to ensure we call the constructor to instantiate our API server in our Server's `main` function. Let's place it between loading our config, and instantiating our actual server.

```go
	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Load our control API
	control.StartControlAPI()

	// Create server using interface's factory function
	server, err := models.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
```



## Test - HTTPS

Just a reminder that we're not yet expecting our actual agent to transition, all we want to see is that hitting our client will indeed change the global flag, which will lead to our server sending `true`/`69.69.69.69`, and that our agent will now display this value.


Let's first test HTTPS - so make sure the `protocol` field in `config.yaml` is set to `https`.

Once both are running, hit our control API endpoint with the following command:

```shell
curl -X POST http://localhost:8080/switch
```

Which will then trigger the output:

```shell
❯ curl -X POST http://localhost:8080/switch
"Protocol transition triggered"
```


Now let's have a look at the output on the server:
```shell
❯ go run ./cmd/server
2025/08/24 09:12:44 Starting Control API on :8080
2025/08/24 09:12:44 Starting https server on 127.0.0.1:8443
2025/08/24 09:13:10 Endpoint / has been hit by agent
2025/08/24 09:13:10 HTTPS: Normal response (change=false)
2025/08/24 09:13:11 Transition triggered
2025/08/24 09:13:14 Endpoint / has been hit by agent
2025/08/24 09:13:14 Transition signal consumed and reset
2025/08/24 09:13:14 HTTPS: Sending transition signal (change=true)

```


We can see that:
- We did indeed start our Control API on 8080
- Our server was initially hit by the agent, we responded with `false`
- The Control API endpoint was hit initiation the transition (`Transition triggered`)
- Thereafter, when the agent hit the endpoint, it "`consumed the signal and reset`" (`CheckAndReset()`)
- The server then sent `change=true`


We can confirm this on the agent's end:
```shell
❯ go run ./cmd/agent
2025/08/24 09:13:10 Starting https client run loop
2025/08/24 09:13:10 Delay: 5s, Jitter: 50%
2025/08/24 09:13:10 Received response: change=false
2025/08/24 09:13:10 Sleeping for 4.796768665s
2025/08/24 09:13:14 Received response: change=true

```


Great, let's do the same thing for DNS, so just change the protocol field value to `dns` in `config.yaml`


## Test - DNS

Do the exact same thing - start the server, start the agent, then hit our Control API endpoint:

```shell
❯ curl -X POST http://localhost:8080/switch
"Protocol transition triggered"

```


We see the same essential logic play out on the server:
```shell
❯ go run ./cmd/server
2025/08/24 09:18:34 Starting Control API on :8080
2025/08/24 09:18:34 Starting dns server on 127.0.0.1:8443
2025/08/24 09:18:42 DNS query for: www.thisdoesnotexist.com.
2025/08/24 09:18:42 DNS: Normal response (42.42.42.42)
2025/08/24 09:18:43 Transition triggered
2025/08/24 09:18:50 DNS query for: www.thisdoesnotexist.com.
2025/08/24 09:18:50 Transition signal consumed and reset
2025/08/24 09:18:50 DNS: Sending transition signal (69.69.69.69)

```


And on the agent
```shell
❯ go run ./cmd/agent
2025/08/24 09:18:42 Starting dns client run loop
2025/08/24 09:18:42 Delay: 5s, Jitter: 50%
2025/08/24 09:18:42 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/24 09:18:42 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
2025/08/24 09:18:42 Received response: IP=42.42.42.42
2025/08/24 09:18:42 Sleeping for 7.213587574s
2025/08/24 09:18:50 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/24 09:18:50 Received DNS response: www.thisdoesnotexist.com. -> 69.69.69.69
2025/08/24 09:18:50 Received response: IP=69.69.69.69
2025/08/24 09:18:50 Sleeping for 6.293953329s

```






___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_d/03_dns_loop.md" >}})
[|NEXT|]({{< ref "02_dual.md" >}})