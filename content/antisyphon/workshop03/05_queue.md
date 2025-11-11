---
showTableOfContents: true
title: "Lesson 5: Queue Commands"
type: "page"
---

## Solutions

The starting solution can be found here.

The final solution can be found here.

## Overview

Now that we can validate and process commands, we need somewhere to store them while waiting for the agent to check in. Remember, our agent operates on a periodic check-in pattern - it's not constantly connected to the server.

We need a **queue** where:

- Validated and processed commands wait for pickup
- Commands are retrieved in the order they were added (FIFO - First In, First Out)
- Multiple goroutines can safely access it (thread-safe) since we might want multiple agents in the future

In this lesson, we'll:

1. Create a CommandQueue struct with thread-safety
2. Implement methods to add and retrieve commands
3. Create a global queue instance
4. Integrate queuing into our command handler



## What We'll Create

- `CommandQueue` struct in `command_api.go`
- `addCommand()` method to add commands to the queue
- Global queue instance (`AgentCommands`)
- Queue integration in `commandHandler`

## Create CommandQueue Struct

Let's create our queue structure in `command_api.go`:

```go
// CommandQueue stores commands ready for agent pickup
type CommandQueue struct {
	PendingCommands []models.CommandClient
	mu              sync.Mutex
}
```



**Understanding the fields:**

1. **PendingCommands** - A slice of `CommandClient` structs
    - Slices in Go are perfect for queues
    - We can easily add to the end with `append()`
    - We can read from the front with `[0]`
    - We can remove from the front with slicing `[1:]`
2. **mu** - A mutex for thread safety
    - Multiple goroutines might try to access the queue simultaneously
    - The mutex ensures only one goroutine can modify the queue at a time
    - This prevents race conditions and data corruption

**Why use a slice for queuing?**

- Dynamic sizing (grows as needed)
- O(1) append operation (adding to end)
- Easy to access first element
- Built-in Go type (no external dependencies)





## Create Global Queue Instance

Since we only want a single queue for all commands, we'll create a global instance. Add this right below the `CommandQueue` struct definition:

```go
// AgentCommands is the global command queue
var AgentCommands = CommandQueue{
	PendingCommands: make([]models.CommandClient, 0),
}
```

**Understanding this declaration:**

- `var AgentCommands` - Creates a global variable named `AgentCommands`
- `CommandQueue{...}` - Initializes it as a `CommandQueue` struct
- `make([]models.CommandClient, 0)` - Creates an empty slice with initial capacity of 0

**Why initialize the slice?** In Go, an uninitialized slice is `nil`, which means you can't append to it. We use `make()` to create an actual, empty slice that's ready to use.



## Implement addCommand Method

Now let's create a method to add commands to the queue:

```go
// addCommand adds a validated command to the queue
func (cq *CommandQueue) addCommand(command models.CommandClient) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.PendingCommands = append(cq.PendingCommands, command)
	log.Printf("QUEUED: %s", command.Command)
}
```



### Method receiver

```go
func (cq *CommandQueue) addCommand(...)
```

This is a method on `CommandQueue`. The `cq` parameter is a pointer to the queue instance.




### Lock the mutex

```go
cq.mu.Lock()
defer cq.mu.Unlock()
```

- `Lock()` - Acquire exclusive access to the queue
- `defer Unlock()` - Ensure the lock is released when the function returns
- Using `defer` is a Go best practice - it guarantees the unlock happens even if there's a panic




### Add to queue

```go
cq.PendingCommands = append(cq.PendingCommands, command)
```


The `append()` function adds the command to the end of the slice. If the slice needs to grow, Go handles this automatically.

### Log the action

```go
log.Printf("QUEUED: %s", command.Command)
```

Provides visibility into what's being queued.




## Update commandHandler

Now let's queue the validated and processed command. Update the end of `commandHandler` in `control_api.go`:

```go
func commandHandler(w http.ResponseWriter, r *http.Request) {

	// Instantiate custom type to receive command from client
	var cmdClient models.CommandClient

	// The first thing we need to do is unmarshal the request body into the custom type
	if err := json.NewDecoder(r.Body).Decode(&cmdClient); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}

	// Normalize command to lowercase
	cmdClient.Command = strings.ToLower(cmdClient.Command)

	// Visually confirm we get the command we expected
	var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
	log.Printf(commandReceived)

	// Check if command exists
	cmdConfig, exists := validCommands[cmdClient.Command]
	if !exists {
		var commandInvalid = fmt.Sprintf("ERROR: Unknown command: %s", cmdClient.Command)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// Validate arguments
	if err := cmdConfig.Validator(cmdClient.Arguments); err != nil {
		var commandInvalid = fmt.Sprintf("ERROR: Validation failed for '%s': %v", cmdClient.Command, err)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// Process arguments (e.g., load file and convert to base64)
	processedArgs, err := cmdConfig.Processor(cmdClient.Arguments)
	if err != nil {
		var commandInvalid = fmt.Sprintf("ERROR: Processing failed for '%s': %v", cmdClient.Command, err)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// Update command with processed arguments
	cmdClient.Arguments = processedArgs
	log.Printf("Processed command arguments: %s", cmdClient.Command)

	// Queue the validated and processed command
	AgentCommands.addCommand(cmdClient)

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)
}
```

**What's new:**

```go
// Queue the validated and processed command
AgentCommands.addCommand(cmdClient)
```

This single line adds our validated and processed command to the global queue. Notice:

- `AgentCommands` - Our global queue instance
- `.addCommand(cmdClient)` - Call the method, passing the complete command struct
- The command now contains processed arguments (with base64 data, not file path)

## Test

Let's verify commands are being queued!

**Start the server:**

```bash
go run ./cmd/server
```

**Send a command:**

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

**Expected client-side response:**

```bash
"Received command: shellcode"
```

**Expected server-side output:**

```bash
2025/11/06 14:49:22 Received command: shellcode
2025/11/06 14:49:22 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/06 14:49:22 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/06 14:49:22 Processed command arguments: shellcode
2025/11/06 14:49:22 QUEUED: shellcode
```

**Analyzing the output:**

- Command received ✓
- Validation passed ✓
- File processed ✓
- **Command queued** ✓ ← NEW!

Perfect! The command is now sitting in the queue, waiting for an agent to pick it up.

## Understanding the Complete Flow So Far

Let's trace a command through the entire pipeline we've built:

1. **Client sends:** Command with arguments via curl
2. **Server receives:** Parses JSON into `CommandClient`
3. **Command validation:** Checks if command exists
4. **Argument validation:** Validates command-specific arguments
5. **Argument processing:** Transforms arguments (e.g., file → base64)
6. **Queuing:** Adds to queue for agent pickup ← **We are here**
7. **Agent pickup:** (Next lesson) Agent retrieves command from queue
8. **Agent execution:** (Future lessons) Agent executes command
9. **Result reporting:** (Future lessons) Agent sends results back

## What's Next?

In the next lesson, we'll implement the server logic to:

- Check if there's a command in the queue
- Retrieve and remove it (dequeue)
- Send it to the agent when they check in

## Conclusion

In this lesson, we've implemented command queuing:

- Created a thread-safe `CommandQueue` struct
- Implemented the `addCommand()` method
- Created a global queue instance
- Integrated queuing into the command handler
- Tested the complete flow from client to queue


In the next lesson, we'll implement the server-side logic to dequeue commands and send them to the agent when it checks in.





___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./04_process_arguments.md" >}})
[|NEXT|]({{< ref "./06_dequeue.md" >}})