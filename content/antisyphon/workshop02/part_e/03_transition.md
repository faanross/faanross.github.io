---
showTableOfContents: true
title: "Agent Parsing + Protocol Transition"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson11_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson11_Done).

## Overview

We've made it! Our final lesson.

To tie everything together, we need our agent to take a specific action based on the response it receives. Right now it does not do anything, whatever value is received, it will just print it to terminal and continue with business as usual.

We need to add the logic for it to transition to the opposite protocol if the right signal was detected. And if not, then just continue with business as usual.

This will all take place in our `RunLoop()`.


## Current State of RunLoop()

Let's just have a quick peek at the current state of RunLoop() to orient ourselves:
```go

func RunLoop(ctx context.Context, comm models.Agent, cfg *config.Config) error {

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
            log.Println("Run loop cancelled")
            return nil
		default:
		}

		response, err := comm.Send(ctx)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			// Don't exit - just sleep and try again
			time.Sleep(cfg.Timing.Delay)
            continue // Skip to next iteration
		}

		// BASED ON PROTOCOL, HANDLE PARSING DIFFERENTLY

		switch cfg.Protocol {
		case "https":
			// Parse and display response
			var httpsResp https.HTTPSResponse
			if err := json.Unmarshal(response, &httpsResp); err != nil {
				log.Fatalf("Failed to parse response: %v", err)
			}
			log.Printf("Received response: change=%v", httpsResp.Change)
		case "dns":
			ipAddr := string(response)
			log.Printf("Received response: IP=%v", ipAddr)

		}

		// Calculate sleep duration with jitter
		sleepDuration := CalculateSleepDuration(time.Duration(cfg.Timing.Delay), cfg.Timing.Jitter)
		log.Printf("Sleeping for %v", sleepDuration)

		// Sleep with cancellation support
		select {
		case <-time.After(sleepDuration):
			// Continue to next iteration
		case <-ctx.Done():
            log.Println("Run loop cancelled")
            return nil
		}
	}
}
```


As I just said above, in our switch statement, based on the protocol, it will "capture" the response, and then print it to terminal.






## Gameplan

There's a handful of changes to fully implement our desired logic, so I'll do it step-by-step. This is also arguably the most complex change we've performed, and so I think it's worth just quickly reviewing what we'll do before just jumping straight into it.


**As we just saw above, currently our `RunLoop()`:**
1. Uses the same `comm` agent for the entire loop
2. Just logs responses without checking them
3. Has no mechanism to switch protocols

**What we need to add in order to achieve our goals are:**
1. Variables to track current state
2. Logic to detect and handle transitions
3. Ability to create and swap to a new agent


## Improving our RunLoop()

First, as I said we need to be able to track the current state. Right now we just have "the state" (`comm`), but we'd like to know both - what is the current state (https/dns), and then by extension what is the opposite state. We can't know what to transition to, if we don't know what the current state is.


So right at the top of the function, first thing even before the `for` loop, add these two variables we'll use to achieve this:
```go
    // ADD THESE TWO LINES:
    currentProtocol := cfg.Protocol  // Track which protocol we're using
    currentAgent := comm              // Track current agent (can change!)
```



Now, instead of referencing `comm` as the current agent, we should use `currentAgent`. Remember, `comm` is hardcoded, it's the value we find initially in our config. Though `currentAgent` starts of being equal to it, it'll soon get the ability to change based on the response we receive from the server.

Find this line:

```go
response, err := comm.Send(ctx)
```

And change it to:

```go
response, err := currentAgent.Send(ctx)
```


And now, all we really need to do is add the transition detection after receiving the response.

So we'll replace our entire current switch block from:

```go
// BASED ON PROTOCOL, HANDLE PARSING DIFFERENTLY
switch cfg.Protocol {
case "https":
    // ... current logging code
case "dns":
    // ... current logging code
}
```


To the following:

```go
// Check if this is a transition signal
if detectTransition(currentProtocol, response) {
    log.Printf("TRANSITION SIGNAL DETECTED! Switching protocols...")
    
    // Figure out what protocol to switch TO
    newProtocol := "dns"
    if currentProtocol == "dns" {
        newProtocol = "https"
    }
    
    // Create config for new protocol
    tempConfig := *cfg  // Copy the config
    tempConfig.Protocol = newProtocol
    
    // Try to create new agent
    newAgent, err := models.NewAgent(&tempConfig)
    if err != nil {
        log.Printf("Failed to create %s agent: %v", newProtocol, err)
        // Don't switch if we can't create agent
    } else {
        // Update our tracking variables
        log.Printf("Successfully switched from %s to %s", currentProtocol, newProtocol)
        currentProtocol = newProtocol
        currentAgent = newAgent
    }
} else {
    // Normal response - parse and log as before
    switch currentProtocol {  // Note: use currentProtocol, not cfg.Protocol
    case "https":
        var httpsResp https.HTTPSResponse
        json.Unmarshal(response, &httpsResp)
        log.Printf("Received response: change=%v", httpsResp.Change)
    case "dns":
        ipAddr := string(response)
        log.Printf("Received response: IP=%v", ipAddr)
    }
}
```


Now this will error out because we're referencing a new helper function `detectTransition()` that we have not yet created. We'll create it soon, but you can already guess what it does - it returns a `bool`, `true` if the value is `true` (HTTPS) or `69.69.69.69` (DNS), and if not `false`.


Let's quickly review what the code does. So if the detection transition was detected (`if detectTransition()`), then:
- We create `newProtocol`, which will assume the opposite value of `currentProtocol`
- We then create a new config by copying the current `config`, and assign it's `protocol` value equal to `newProtocol`
- We'll then create our new agent according to our updated protocol value
- If this succeeds `currentProtocol` becomes `newProtocol`, and `currentAgent` becomes `newAgent`.

And if `detectTransition()` returns `false`, represented by the `else` case, then we're simply using our old logic - we're just parsing and printing the value to console.



## detectTransition()

```go
// detectTransition checks if the response indicates we should switch protocols
func detectTransition(protocol string, response []byte) bool {
    switch protocol {
    case "https":
        var httpsResp https.HTTPSResponse
        if err := json.Unmarshal(response, &httpsResp); err != nil {
            return false
        }
        return httpsResp.Change
        
    case "dns":
        ipAddr := string(response)
        return ipAddr == "69.69.69.69"
    }
    
    return false
}
```

For HTTPS: We return the actual value of the Change field.
For DNS: We are asking: **is ipAddr == "69.69.69.69"**? If yes, it returns `true` and vice-versa.


## Quick Recap
Let's just quickly recap what we did:
1. **State Tracking**: We now track which protocol and agent we're currently using
2. **Dynamic Switching**: When transition detected, we create a new agent
3. **Simple Swap**: We just update the variables - next loop iteration uses new agent
4. **Graceful Degradation**: If new agent creation fails, keep using current one




### The Flow

```
Loop iteration 1: HTTPS agent → server → "false" → normal log
Loop iteration 2: HTTPS agent → server → "false" → normal log
[API HIT]
Loop iteration 3: HTTPS agent → server → "true" → DETECT! → create DNS agent → switch
Loop iteration 4: DNS agent → server → "42.42.42.42" → normal log
Loop iteration 5: DNS agent → server → "42.42.42.42" → normal log
```

That's it! The beauty is in its simplicity - just track state, detect signals, and swap references.


## Test

Let's start up our server, and our agent. We can start with any protocol, in this specific case I'll start with `https`.


Once it's running, let's hit our endpoint with:
```bash
curl -X POST http://localhost:8080/switch
```


And then wait a few moments, and do it again to confirm we can switch back. Of course feel free to repeat this as many times as you'd like.

Let's have a look at our server-side output:
```shell
❯ go run ./cmd/server
2025/08/24 10:43:05 Starting Control API on :8080
2025/08/24 10:43:05 Starting both protocol servers on 127.0.0.1:8443
2025/08/24 10:43:05 Starting HTTPS server on 127.0.0.1:8443 (TCP)
2025/08/24 10:43:05 Starting DNS server on 127.0.0.1:8443 (UDP)
2025/08/24 10:43:28 Endpoint / has been hit by agent
2025/08/24 10:43:28 HTTPS: Normal response (change=false)
2025/08/24 10:43:29 Transition triggered
2025/08/24 10:43:32 Endpoint / has been hit by agent
2025/08/24 10:43:32 Transition signal consumed and reset
2025/08/24 10:43:32 HTTPS: Sending transition signal (change=true)
2025/08/24 10:43:36 DNS query for: www.thisdoesnotexist.com.
2025/08/24 10:43:36 DNS: Normal response (42.42.42.42)
2025/08/24 10:43:38 Transition triggered
2025/08/24 10:43:40 DNS query for: www.thisdoesnotexist.com.
2025/08/24 10:43:40 Transition signal consumed and reset
2025/08/24 10:43:40 DNS: Sending transition signal (69.69.69.69)
2025/08/24 10:43:45 Endpoint / has been hit by agent
2025/08/24 10:43:45 HTTPS: Normal response (change=false)

```


We can confirm that:
- We started on HTTPs
- We endpoint has hit and we transitioned to sending a DNS response
- The endpoint was hit again and we transitioned back to a HTTPS response



And if we take a peek at our agent-side output we'll see this same pattern play out on the coin's other side:
```shell
❯ go run ./cmd/agent
2025/08/24 10:43:28 Starting https client run loop
2025/08/24 10:43:28 Delay: 5s, Jitter: 50%
2025/08/24 10:43:28 Received response: change=false
2025/08/24 10:43:28 Sleeping for 3.570247958s
2025/08/24 10:43:32 TRANSITION SIGNAL DETECTED! Switching protocols...
2025/08/24 10:43:32 Successfully switched from https to dns
2025/08/24 10:43:32 Sleeping for 4.767393586s
2025/08/24 10:43:36 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/24 10:43:36 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
2025/08/24 10:43:36 Received response: IP=42.42.42.42
2025/08/24 10:43:36 Sleeping for 3.761613577s
2025/08/24 10:43:40 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/24 10:43:40 Received DNS response: www.thisdoesnotexist.com. -> 69.69.69.69
2025/08/24 10:43:40 TRANSITION SIGNAL DETECTED! Switching protocols...
2025/08/24 10:43:40 Successfully switched from dns to https
2025/08/24 10:43:40 Sleeping for 5.035840383s
2025/08/24 10:43:45 Received response: change=false
2025/08/24 10:43:45 Sleeping for 4.166015597s

```




## Conclusion

And that's it, we now have a simple, but potent foundation for a covert channel that can transition between two different protocols on-demand.


___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "02_dual.md" >}})
