---
showTableOfContents: true
title: "HTTPS Agent"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson04_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson04_Done).

## Overview
We'll now create our HTTPS agent, the yin to our server's yang. We'll once again use a struct to represent an instance of
an agent, a constructor to instantiate it (which will be wired into our `NewAgent()` factory function), and then to satisfy
the `Agent` interface we'll need to create the associated `Send()` method. Let's do this.


## What We'll Create
- HTTPS agent (`./internals/https/agent_https.go`)





## Agent struct

Just as we did on the server side, we'll create a struct to represent an instance of our HTTPS agent. We'll house all our HTTPS agent logic in `./internals/https/agent_https.go`.

```go
// HTTPSAgent implements the Communicator interface for HTTPS
type HTTPSAgent struct {
	serverAddr string
	client     *http.Client
}
```


We'll once again use the standard `net/http` library, this time for our `client` field.


## Constructor

Let's add our constructor to instantiate the `HTTPSAgent` struct.


```go

// NewHTTPSAgent creates a new HTTPS agent
func NewHTTPSAgent(serverAddr string) *HTTPSAgent {
	// Create TLS config that accepts self-signed certificates
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Create HTTP client with custom TLS config
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &HTTPSAgent{
		serverAddr: serverAddr,
		client:     client,
	}
}
```

This is all pretty straight-forward, the one thing to perhaps point out of course is that we're explicitly allowing the use of self-signed cert - we're telling the agent,  no need to verify whether the server's cert is self-signed or not, just accept whatever is presented.


## Send()

On our agent side we now require the `Send()` method to satisfy the interface.

```go

// Send implements Communicator.Send for HTTPS
func (c *HTTPSAgent) Send(ctx context.Context) ([]byte, error) {
	// Construct the URL
	url := fmt.Sprintf("https://%s/", c.serverAddr)

	// Create GET request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, body)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}


	// Return the raw JSON as message data
	return body, nil
}

```

As we saw on the server-side, we just need to send a GET request to the root endpoint `/` to call the handler. So we construct the target URL, create our GET request, and we send it.

We'll read the response body using `io.ReadAll()`, we're of course expecting this to be the JSON with the `change` field. Right now all we'll do is return this value to the caller, in a later lesson we'll also parse this to see whether it's `true` or `false`, and implement conditional logic based on that.


## Update factory

Now that we have our Agent type, constructor, and method to satisfy the interface we can go ahead and wire this into our `NewAgent` factory function.


```go
// NewAgent creates a new communicator based on the protocol
func NewAgent(cfg *config.Config) (Agent, error) {
	switch cfg.Protocol {
	case "https":
		return https.NewHTTPSAgent(cfg.ServerAddr), nil
	case "dns":
		return nil, fmt.Errorf("DNS not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```




## Update our agent's main entrypoint

We can now update our Agent's `main` function so that it passes the config as an argument to the `NewAgent` factory function, whereafter we can call `Send()` on our `Agent` instance.


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
	
		// TEMPORARY CODE JUST TO TEST!
		// Send a test message
	
		log.Printf("Sending request to %s server...", cfg.Protocol)
		response, err := comm.Send(context.Background())
		if err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
	
		// Parse and display response
		var httpsResp https.HTTPSResponse
		if err := json.Unmarshal(response, &httpsResp); err != nil {
			log.Fatalf("Failed to parse response: %v", err)
		}
	
		log.Printf("Received response: change=%v", httpsResp.Change)

}

```

Note the temporary code as indicated by the comment. In the next lesson we'll create our Agent loop - that is the logic so it repetitively calls `Send()` in a loop. For now we'll just call `Send()` once, and since that function just returns the response body without printing it, we'll unmarshall and print it here.


## test

Let's run our server:
```shell
❯ go run cmd/server/main.go
2025/08/11 12:02:40 Starting https server on 127.0.0.1:8443
```



Then let's run our agent:
```shell
❯ go run ./cmd/agent/
2025/08/11 12:03:05 Sending request to https server...
2025/08/11 12:03:05 Received response: change=false
```

We can see it received the correct `false` value for the `change` field.

And looking back at the server output we'll once again see confirmation that it was hit.

```shell
❯ go run cmd/server/main.go
2025/08/11 12:02:40 Starting https server on 127.0.0.1:8443
2025/08/11 12:03:05 Endpoint / has been hit by agent
```




## Conclusion

We can now add a loop to our HTTPS logic so that the agent will periodically connect to the server, which of course serves as our "heartbeat".





___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "01_https_server.md" >}})
[|NEXT|]({{< ref "03_https_loop.md" >}})