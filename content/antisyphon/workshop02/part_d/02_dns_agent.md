---
showTableOfContents: true
title: "DNS Agent"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson07_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson07_Done).


## Overview
Now that we've created our DNS Server we can create our own DNS Agent, also leveraging `miekg/dns`, to communicate with it.


## What We'll Create
- DNS Agent  (`internals/dns/agent_dns.go`)



## DNS Agent

Let's create a new file in `internals/dns/agent_dns.go`. We'll first create our DNS Agent struct and associated constructor.

```go
// DNSAgent implements the Agent interface for DNS
type DNSAgent struct {
	serverAddr string
	client     *dns.Client
}

// NewDNSAgent creates a new DNS client
func NewDNSAgent(serverAddr string) *DNSAgent {
	return &DNSAgent{
		serverAddr: serverAddr,
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

We set all the required values, including domain hardcoded here to  `"www.thisdoesnotexist.com."`. As already stated, this can be anything for now - our server has no conditional logic predicated on the actual value.

We then send the request with `Exchange()`, and process the response contained in `r.Answer`.







## Update Agent Factory Function

Now we can make the final update to our factory function:
```go
// NewAgent creates a new communicator based on the protocol
func NewAgent(cfg *config.Config) (Agent, error) {
	switch cfg.Protocol {
	case "https":
		return https.NewHTTPSAgent(cfg.ServerAddr), nil
	case "dns":
		return dns.NewDNSAgent(cfg.ServerAddr), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```





## Temp Change to Agent's main

Since our Agent's main currently leverages the runloop, which cannot yet handle DNS responses, we need to temporarily comment out the code associated with the runloop and add temporary code for an once-off send with `comm.Send(ctx)` in order to test our DNS agent:


```go
// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	comm.Send(ctx)

	// Start run loop in goroutine
	//go func() {
	//	log.Printf("Starting %s client run loop", cfg.Protocol)
	//	log.Printf("Delay: %v, Jitter: %d%%", cfg.Timing.Delay.Duration, cfg.Timing.Jitter)
	//
	//	if err := runloop.RunLoop(ctx, comm, cfg.Timing.Delay.Duration, cfg.Timing.Jitter); err != nil {
	//		log.Printf("Run loop error: %v", err)
	//	}
	//}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
```


Don't worry about the fact that we're not handling the return error from `Send()` - this is just a temporary test.


## Test

First start the server with go run `./cmd/server`, we can then run our agent and it will send a single request, and process the response.


```shell
❯ go run ./cmd/agent
2025/08/11 17:14:55 Sending DNS query for: www.thisdoesnotexist.com.
2025/08/11 17:14:55 Received DNS response: www.thisdoesnotexist.com. -> 42.42.42.42
^C2025/08/11 17:15:01 Shutting down client...

```


Here we can indeed see `42.42.42.42` printed to terminal - the indication to not change the underlying communication protocol.



___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "01_dns_server.md" >}})
[|NEXT|]({{< ref "03_dns_loop.md" >}})