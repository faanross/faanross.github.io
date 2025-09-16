---
showTableOfContents: true
title: "DNS Loop"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson08_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson08_Done).

## Overview

We'll now make a small change to our current `RunLoop()` so that it's able to discriminate between HTTPS and DNS. This way we can use the same function for both instead of needing to each protocol to have its own distinct `RunLoop()` function.





## Key Differences from HTTPS Run Loop

Let's have a quick look at our current `RunLoop()` implementation:

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


If we review this carefully, the only "non-agnostic" code, i.e. the only logic that is HTTPS-specific, is the following:

```go
	// Parse and display response
	var httpsResp https.HTTPSResponse
	if err := json.Unmarshal(response, &httpsResp); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	log.Printf("Received response: change=%v", httpsResp.Change)
```


If we were to implement this same basic idea using DNS, it would be:
```go
	// Parse DNS response (IP address)
	ipAddr := string(response)
	log.Printf("Received response: IP=%v", ipAddr)
```



## Updated RunLoop

And so all we really need to do is implement conditional logic. In this case, even though we could get away with an `if/else`, I'll opt for a `switch` for 2 reasons:
- First it makes both cases explicit,
- Second it allows us to add more protocols in the future with resorting to multiple `if-else's`


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


And so we can see very simply that we switch on the `cfg.Protocol`, and then execute the logic pertinent to that protocol.

## Updating our Agent's main

Let's undo the changes we made at the end of the last lab so that we once again use `RunLoop()`:
1. Delete the single `Send()`
2. Uncomment the commented out code

```go

func main() {
	// Command line flag for config file path
	configPath := flag.String("config", pathToYAML, "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	comm, err := models.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create communicator: %v", err)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start run loop in goroutine
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



## Test

Let's first run our server with `go run ./cmd/server`,  then we can run our agent and confirm that it periodically checks in with our server and process the response.


```go
❯ go run ./cmd/agent
2025/08/12 12:12:34 Starting dns client run loop
2025/08/12 12:12:34 Delay: 5s, Jitter: 50%
2025/08/12 12:12:34 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/12 12:12:34 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
2025/08/12 12:12:34 Received response: IP=42.42.42.42
2025/08/12 12:12:34 Sleeping for 4.163428726s
2025/08/12 12:12:39 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/12 12:12:39 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
2025/08/12 12:12:39 Received response: IP=42.42.42.42
2025/08/12 12:12:39 Sleeping for 3.64610957s
2025/08/12 12:12:42 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/12 12:12:42 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
2025/08/12 12:12:42 Received response: IP=42.42.42.42
2025/08/12 12:12:42 Sleeping for 5.117270683s
^C2025/08/12 12:12:45 Shutting down client...

```



## Conclusion
Awesome. We have all our foundational code related to both of our core communication protocols HTTPS and DNS.

We're now ready to start the final phase of our workshop where we'll:
- Create an API trigger to indicate that we'd like to transition from one protocol to another,
- Create logic on our server so that it changes the response (from `false` to `true` for HTTPS, from `42.42.42.42` to `69.69.69.69` for DNS),
- Create logic on our agent to parse the response,
- Conditional logic on our agent to either continue using the same protocol, or transition to the opposite protocol, based on the parsed value.




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "02_dns_agent.md" >}})
[|NEXT|]({{< ref "../part_e/01_api.md" >}})