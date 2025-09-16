---
showTableOfContents: true
title: "HTTPS Loop"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson05_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson05_Done).

## Overview
At this point we have a HTTPS agent and server with the ability to connect to one another, but as we just saw it's a kinda "one and done" deal.

As we know, when it comes to C2s, we want them to periodically check-in. We want the agent, after some amount of time (delay + jitter) to send a request to the server.

So that's what we'll create in this section, the runloop - the agent-side logic that will ensure it periodically connects to the server, sends a request, and processes the response.





## What We'll Create
- Agent runloop (`internals/runloop/runloop.go`)







## Helper function

Before we get to the actual `RunLoop()` function, let's add a helper function that will determine the exact amount of time to sleep between each round. If you can recall, we already specified `delay`, and `jitter` in our config. This function thus uses those values and each time calculates a new random value within the range determined by the chosen delay and jitter.

In a new file `./internals/runloop/runloop.go` let's create the following:
```go
// CalculateSleepDuration calculates the actual sleep time with jitter
func CalculateSleepDuration(baseDelay time.Duration, jitterPercent int) time.Duration {
	if jitterPercent == 0 {
		return baseDelay
	}

	// Calculate jitter range
	jitterRange := float64(baseDelay) * float64(jitterPercent) / 100.0

	// Random value between -jitterRange and +jitterRange
	jitter := (rand.Float64()*2 - 1) * jitterRange

	// Calculate final duration
	finalDuration := float64(baseDelay) + jitter

	// Ensure we don't go negative
	if finalDuration < 0 {
		finalDuration = 0
	}

	return time.Duration(finalDuration)
}

```


At the end we'll return a duration of type `time.Duration`, which can be used directly by our `RunLoop()` function.


## RunLoop()

Speaking of, in the same file we can now create our `RunLoop()`. 

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

		// Parse and display response
		var httpsResp https.HTTPSResponse
		if err := json.Unmarshal(response, &httpsResp); err != nil {
			log.Fatalf("Failed to parse response: %v", err)
		}

		log.Printf("Received response: change=%v", httpsResp.Change)

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


The logical heart of this function is repurposed directly from the previous lesson's `main()` function - we use `Send()` to send a GET request to the server and return the raw response body, whereafter we unmarshall it here and print it to terminal.

Also note of course that everything is wrapped inside an infinite `for{}` loop - this is what allows this function to repeat indefinitely. But, of course, we need some kind of "out" - that's exactly what the select statement at the top is for. In the case of `ctx.Done()`, it will return and thus break out of the for `loop`. We'll soon see how that condition is triggered from our `main()` function.


Then note at the bottom we used another `select` statement. Now, we could have just said `time.After(sleepDuration)`, that would have worked too. The issue with that is, let's say we're using a long sleep duration of for example 5 minutes, and 1 minute in we want to intentionally kill the Agent process. In that case it'll take 4 minutes before being able to trigger the top `select` statement and do so.

Thus by doing this we ensure that we're able to exit our `RunLoop()` function, even when `Sleep()` is executing.

So with that implemented, we can now update our Agent's main function to use this instead.

## Agent's main

Back inside of `./cmd/agent/main.go` we can now remove the previous "once off" logic and utilize our `RunLoop()`, and add some signal handling to allow for graceful shutdown.

```go
const pathToYAML = "./configs/config.yaml"

func main() {
	// Command line flag for config file path
	configPath := flag.String("config", pathToYAML, "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Call our factory function
	comm, err := models.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create communicator: %v", err)
	}

	// ALL THIS DOWN HERE IS THE NEW CODE
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Start run loop in goroutine
	go func() {
		log.Printf("Starting %s client run loop", cfg.Protocol)
		log.Printf("Delay: %v, Jitter: %d%%", cfg.Timing.Delay, cfg.Timing.Jitter)

		if err := runloop.RunLoop(ctx, comm, cfg); err != nil {
			log.Printf("Run loop error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("Shutting down client...")
	cancel() // This will cause the run loop to exit
}

```

Note that, as was the case with the server, we'll call `RunLoop` in a separate goroutine since it's blocking. Then we'll implement the same system of signal handling using `sigChan`, which blocks at `<-sigChan`.

Once we hit Ctrl+C we'll pass beyond that point, which will call `cancel()`. If `cancel()` is called, it then leads to the `ctx.Done()` being closed, so that select case is met in our `RunLoop()`, allowing it to exit out of the infinite `for`.




## Test

Let's run our server, and now if we run our agent we expect there to be periodical "check-ins" - we hit the endpoint, sleep for some time, and repeat this until we hit Ctrl + C.

```shell
❯ go run ./cmd/agent
2025/08/11 14:21:35 Starting https client run loop
2025/08/11 14:21:35 Delay: 5s, Jitter: 50%
2025/08/11 14:21:35 Received response: change=false
2025/08/11 14:21:35 Sleeping for 6.785310531s
2025/08/11 14:21:42 Received response: change=false
2025/08/11 14:21:42 Sleeping for 5.524003837s
^C2025/08/11 14:21:46 Shutting down client...

```


And that's exactly what we get.


## Conclusion

We have all our core HTTPS-logic. We'll now head over and do the same essential thing for DNS - we'll create a server and agent. And then, instead of creating it's own `RunLoop()`, we'll just repurpose this `RunLoop()` to work for both protocols.




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "02_https_agent.md" >}})
[|NEXT|]({{< ref "../part_d/01_dns_server.md" >}})