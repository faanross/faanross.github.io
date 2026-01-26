---
layout: course01
title: "Lesson 6: DNS Agent"
---


## Solutions

- **Starting Code:** [lesson_06_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_06_begin)
- **Completed Code:** [lesson_06_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_06_end)

## Overview

Now that we've created our DNS Server we can create our own DNS Agent, also leveraging `miekg/dns`, to communicate with it.

## What We'll Create

- DNS Agent (`internals/agent/agent_dns.go`)

## DNS Agent

Let's create a new file in `internals/agent/agent_dns.go`. We'll first create our DNS Agent struct and associated constructor.

```go
// DNSAgent implements the Agent interface for DNS
type DNSAgent struct {
	serverAddr string
	client     *dns.Client
}

// NewDNSAgent creates a new DNS client
func NewDNSAgent(serverIP string, serverPort string) *DNSAgent {
	return &DNSAgent{
		serverAddr: fmt.Sprintf("%s:%s", serverIP, serverPort),
		client:     new(dns.Client),
	}
}
```

## Send()

Then, to satisfy our Agent interface, we need to implement our Send() method.

```go
// Send implements Agent.Send for DNS
func (c *DNSAgent) Send(ctx context.Context) ([]byte, error) {
	// Create DNS query message
	m := new(dns.Msg)

	// For now, we'll query for a fixed domain
	domain := "www.thisdoesnotexist.com."
	m.SetQuestion(domain, dns.TypeA)
	log.Printf("Sending DNS query for: %s", domain)

	// Send query
	r, _, err := c.client.Exchange(m, c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("DNS exchange failed: %w", err)
	}

	// Check if we got an answer
	if len(r.Answer) == 0 {
		return nil, fmt.Errorf("no answer received")
	}

	// Extract the first A record
	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			// Return the IP address as string
			ipStr := a.A.String()
			log.Printf("Received DNS response: %s -> %s", domain, ipStr)
			return []byte(ipStr), nil
		}
	}

	return nil, fmt.Errorf("no A record in response")
}
```

We can see this time we start the same as with the DNS Server - by creating a `dns.Msg`. Note that when we create it it's a request by default. That's why with our server we had to manually set it to `reply`, whereas here we don't have to set anything.

We set all the required values, including domain hardcoded here to `"www.thisdoesnotexist.com."`. As already stated, this can be anything for now - our server has no conditional logic predicated on the actual value.

We then send the request with `Exchange()`, and process the response contained in `r.Answer`.

## Update Agent Factory Function

Now we can make the final update to our factory function in `internals/agent/models.go`:

```go
// NewAgent creates a new agent based on the protocol
func NewAgent(cfg *config.AgentConfig) (Agent, error) {
	switch cfg.Protocol {
	case "https":
		return NewHTTPSAgent(cfg.ServerIP, cfg.ServerPort), nil
	case "dns":
		return NewDNSAgent(cfg.ServerIP, cfg.ServerPort), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```

## Temp Change to Agent's main

Since our Agent's main currently leverages the runloop, which cannot yet handle DNS responses, we need to temporarily modify the code to do a once-off send with `comm.Send(ctx)` to test our DNS agent.

Update your `cmd/agent/main.go` to change the protocol to `dns` and temporarily test:

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"c2framework/internals/agent"
	"c2framework/internals/config"
)

func main() {
	// Create agent config - temporarily set to dns for testing
	cfg := &config.AgentConfig{
		Protocol:   "dns",
		ServerIP:   "127.0.0.1",
		ServerPort: "8443",
		Timing: config.TimingConfig{
			Delay:  5 * time.Second,
			Jitter: 50,
		},
	}

	// Call our factory function
	comm, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Temporary test: single send (runloop doesn't handle DNS yet)
	comm.Send(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("Shutting down client...")
	cancel()
}
```

Don't worry about the fact that we're not handling the return error from `Send()` - this is just a temporary test.

## Test

First start the server with `go run ./cmd/server` (make sure it's set to `dns` protocol), then run our agent and it will send a single request and process the response.

```shell
go run ./cmd/agent
2025/08/11 17:14:55 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/11 17:14:55 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
^C2025/08/11 17:15:01 Shutting down client...
```

Here we can indeed see `42.42.42.42` printed to terminal - the indication to not change the underlying communication protocol.

## Conclusion

Now we have both DNS agent and server. In the next lesson, we'll update our RunLoop to handle both HTTPS and DNS responses.

---

[Previous: Lesson 5 - DNS Server](/courses/course01/lesson-05) | [Next: Lesson 7 - Add DNS to RunLoop](/courses/course01/lesson-07) | [Course Home](/courses/course01)
