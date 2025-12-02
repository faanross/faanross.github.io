---
showTableOfContents: true
title: "Review of Starting Code"
type: "page"
---

The starting code can be found [here](https://github.com/faanross/workshop_antisyphon_23012026/tree/main/lesson_00_starting_code). 

# Lesson A: Review of Starting Code

## Overview

Before we begin building our command and control system, let's review the code we're starting with. Understanding this foundation will help you see how we'll extend and enhance it throughout the workshop.

Our starting code provides a basic HTTPS server and agent that can communicate with each other. The agent periodically checks in with the server, and the server responds. It's minimal but functional - exactly what we need as a foundation.

## What We're Starting With

```
workshop3_dev/
├── cmd/
│   ├── agent/
│   │   └── main.go          # Agent entry point
│   └── server/
│       └── main.go          # Server entry point
├── internals/
│   ├── agent/
│   │   ├── agent.go         # Agent communication logic
│   │   └── runloop.go       # Periodic check-in loop
│   ├── control/
│   │   └── control_api.go   # Dummy control endpoint
│   └── server/
│       └── server.go        # HTTPS server
├── certs/
│   ├── server.crt           # Self-signed certificate
│   └── server.key           # Private key
└── go.mod                   # Go module dependencies
```



## Server Components

### Server Main (`cmd/server/main.go`)

```go
func main() {

	serverInterface := "0.0.0.0:8443"

	// Load our control API
	control.StartControlAPI()

	newServer := server.NewServer(serverInterface)

	// Start server in goroutine
	go func() {
		log.Printf("Starting  server on %s", serverInterface)
		if err := newServer.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down server...")

	if err := newServer.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

}
```

**What it does:**

1. **Starts two servers:**
    - Main HTTPS server on port 8443 (for agent communication)
    - Control API on port 8080 (for operator commands)
2. **Signal handling:** Listens for Ctrl+C to shut down gracefully
3. **Concurrent execution:** Runs servers in goroutines so they don't block

**Key concepts:**

- `go func() { }()` - Starts a goroutine (concurrent execution)
- `signal.Notify()` - Registers for OS signals
- `<-sigChan` - Blocks until signal received



### HTTPS Server (`internals/server/server.go`)

```go
type Server struct {
	addr    string
	server  *http.Server
	tlsCert string
	tlsKey  string
}

func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		tlsCert: "./certs/server.crt",
		tlsKey:  "./certs/server.key",
	}
}

func (server *Server) Start() error {
	// Create Chi router
	r := chi.NewRouter()

	// Define our GET endpoint
	r.Get("/", RootHandler)

	// Create the HTTP server
	server.server = &http.Server{
		Addr:    server.addr,
		Handler: r,
	}

	// Start the server
	return server.server.ListenAndServeTLS(server.tlsCert, server.tlsKey)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	response := "You have hit the server's root endpoint"

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (server *Server) Stop() error {
	// If there's no server, nothing to stop
	if server.server == nil {
		return nil
	}

	// Give the server 5 seconds to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.server.Shutdown(ctx)
}
```

**What it does:**

1. **Struct-based design:** Encapsulates server configuration
2. **Constructor pattern:** `NewServer()` initializes the server
3. **HTTPS with self-signed certs:** Uses TLS for encrypted communication
4. **Simple handler:** Returns a basic JSON response when agent checks in
5. **Graceful shutdown:** 5-second timeout to finish handling requests

**Key concepts:**

- **Method receivers:** `(server *Server)` makes functions methods on the struct
- **Chi router:** Cleaner than standard library routing
- **Context with timeout:** Prevents hanging during shutdown
- **JSON encoding:** Automatic marshaling to JSON



### Control API (`internals/control/control_api.go`)

```go
func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	// Define the POST endpoint
	r.Get("/dummy", dummyHandler)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("dummyHandler called")

	response := "Dummy endpoint triggered"

	json.NewEncoder(w).Encode(response)
}
```

**What it does:**

1. **Separate API:** Port 8080 for operator commands (distinct from agent communication on 8443)
2. **Dummy endpoint:** `/dummy` placeholder that we'll replace with later on
3. **Non-blocking:** Runs in a goroutine so it doesn't block server startup


## Agent Components

### Agent Main (`cmd/agent/main.go`)

```go
func main() {

	serverAddr := "0.0.0.0:8443"
	delay := 5 * time.Second
	jitter := 50

	// Create our Agent instance
	newAgent := agent.NewAgent(serverAddr)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start run loop in goroutine
	go func() {
		log.Printf("Starting Agent Run Loop")
		log.Printf("Delay: %v, Jitter: %d%%", delay, jitter)

		if err := agent.RunLoop(newAgent, ctx, delay, jitter); err != nil {
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

**What it does:**

1. **Configuration:** Server address, delay (5s), jitter (50%)
2. **Context for cancellation:** Allows clean shutdown
3. **Run loop in goroutine:** Continuous operation without blocking
4. **Signal handling:** Ctrl+C triggers shutdown

**Key concepts:**

- **Context:** Go's standard way to handle cancellation and timeouts
- `context.WithCancel()` - Creates cancellable context
- `cancel()` - Triggers cancellation
- `defer cancel()` - Ensures context is cancelled on exit




### Agent Communication (`internals/agent/agent.go`)

```go
// Agent implements the Communicator interface for HTTPS
type Agent struct {
	serverAddr string
	client     *http.Client
}

// NewAgent creates a new HTTPS agent
func NewAgent(serverAddr string) *Agent {
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

	return &Agent{
		serverAddr: serverAddr,
		client:     client,
	}
}

// Send implements Communicator.Send for HTTPS
func (agent *Agent) Send(ctx context.Context) ([]byte, error) {
	// Construct the URL
	url := fmt.Sprintf("https://%s/", agent.serverAddr)

	// Create GET request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Send request
	resp, err := agent.client.Do(req)
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

**What it does:**

1. **TLS configuration:** `InsecureSkipVerify: true` accepts self-signed certs
2. **HTTP client:** Reusable client with custom TLS config
3. **Context-aware requests:** Can be cancelled mid-flight

**Key concepts:**
- **InsecureSkipVerify:** Required for self-signed certs
- **Context in requests:** `NewRequestWithContext()` allows cancellation
- **Error wrapping:** `fmt.Errorf(..., %w, err)` preserves error chain
- **defer resp.Body.Close():** Prevents resource leaks



### Agent Run Loop (`internals/agent/runloop.go`)

```go
func RunLoop(agent *Agent, ctx context.Context, delay time.Duration, jitter int) error {

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			log.Println("Run loop cancelled")
			return nil
		default:
		}

		response, err := agent.Send(ctx)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			// Don't exit - just sleep and try again
			time.Sleep(delay)
			continue // Skip to next iteration
		}

		log.Printf("Response from server: %s", response)

		// Calculate sleep duration with jitter
		sleepDuration := CalculateSleepDuration(delay, jitter)

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

**What it does:**

1. **Infinite loop:** Runs until cancelled
2. **Context checking:** Two places where cancellation is checked
3. **Send request:** Contact server, get response
4. **Error handling:** Don't crash on errors, just log and retry
5. **Jitter calculation:** Randomizes sleep time to avoid patterns

**Understanding jitter:**

```
Base delay: 5 seconds
Jitter: 50%
Jitter range: ±2.5 seconds
Actual sleep: Random between 2.5s and 7.5s

Why? Prevents pattern recognition:
- Without jitter: 10:00:00, 10:00:05, 10:00:10 (predictable)
- With jitter:    10:00:00, 10:00:06, 10:00:09 (unpredictable)
```

**Key concepts:**

- **Select statement:** Non-blocking channel operations
- `case <-ctx.Done()` - Triggered when context cancelled
- `case <-time.After()` - Triggered after duration
- `default:` - Runs if no case is ready
- **Continue:** Skip to next loop iteration



## Running the Starting Code


You should now have a good overall sense of the major parts of our starting code logic and what it does. Let's quickly run it, just to get a first-hand experience of what it does.

From a terminal navigate to the root folder of the starting code.

### Start the Server

```bash
go run ./cmd/server
```

You should see both the Control API and Server running

```bash
❯ go run ./cmd/server
2025/11/10 14:26:25 Starting Control API on :8080
2025/11/10 14:26:25 Starting  server on 0.0.0.0:8443
```


We can of course confirm this using `lsof`:
```bash
❯ lsof -i :8443,8080
COMMAND   PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
server  73657 faanross    5u  IPv6 0x6c6595677cc65d97      0t0  TCP *:pcsync-https (LISTEN)
server  73657 faanross    6u  IPv6 0x22656dd0ffd4bfe2      0t0  TCP *:http-alt (LISTEN)
```



### Test the Control API

In another terminal:

```bash
curl http://localhost:8080/dummy
```


We should get the following feedback immediately after running the command.

```bash
❯ curl http://localhost:8080/dummy
"Dummy endpoint triggered"
```


And in server logs we should see:

```bash
2025/11/10 14:27:53 dummyHandler called
```



### Start the Agent

In another terminal:

```bash
go run ./cmd/agent
```

We'll see the following periodical output on the agent side.

```bash
❯ go run ./cmd/agent
2025/11/10 14:29:05 Starting Agent Run Loop
2025/11/10 14:29:05 Delay: 5s, Jitter: 50%
2025/11/10 14:29:05 Response from server: "You have hit the server's root endpoint"
2025/11/10 14:29:05 Sleeping for 2.934748748s
```


And on the server side we'll see our endpoint is being hit.
```bash
2025/11/10 14:29:05 Endpoint / has been hit by agent
```


Perfect! The agent is checking in, and the server is responding.


## Key Takeaways

### What Works

✓ **HTTPS communication** with self-signed certificates  
✓ **Periodic check-ins** with jitter  
✓ **Graceful shutdown** via context cancellation  
✓ **Concurrent execution** with goroutines  
✓ **Basic routing** with Chi  
✓ **Error handling** with wrapped errors

### What's Missing

✗ **No command system** - Just a dummy endpoint  
✗ **No command validation** - No checking of inputs  
✗ **No command processing** - No argument handling  
✗ **No command queue** - No storage while waiting for agent  
✗ **No command execution** - Agent doesn't do anything with responses  
✗ **No results reporting** - No feedback from agent to server


## Conclusion

This is a good "vanilla" foundation for us to build upon, most of the basic housekeeping is done and we can now implement most of the features listed above under "What's Missing". Let's review exactly what we'll implement in the next lesson.






___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./00A_setup.md" >}})
[|NEXT|]({{< ref "./00C_overview.md" >}})