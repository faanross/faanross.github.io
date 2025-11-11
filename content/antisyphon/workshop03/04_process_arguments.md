---
showTableOfContents: true
title: "Lesson 4: Process Command Arguments"
type: "page"
---
## Solutions

The starting solution can be found here.

The final solution can be found here.



## Overview

Just like we validated command-specific arguments, sometimes we also need to **process** them - prepare them before they can be received by the agent. Note the word *sometimes* - for some commands we'll be able to send the same arguments we receive from client directly to the agent. In other cases however, like with our shellcode loader, we have to do some transformation first.

For our shellcode command, the client sends us a file path, but we can't send a file path to the agent (remember the agent is on a different machine with a different filesystem). Instead, we need to:

1. Read the DLL file from disk
2. Convert the binary data to base64
3. Send the base64 string to the agent

In this lesson then we'll:

1. Create a function type for command processors
2. Create a new argument type for agent-bound data
3. Implement the processor for shellcode
4. Integrate processing into our command handler

## What We'll Create

- `CommandProcessor` function type in `command_api.go`
- `ShellcodeArgsAgent` type in `models/types.go`
- `processShellcodeCommand` function in `shellcode.go`
- Updated command registry with processors
- Processing logic in `commandHandler`


## Create CommandProcessor Type

We'll follow the same pattern we used for validators. Add this to `command_api.go`:

```go
// CommandProcessor processes command-specific arguments
type CommandProcessor func(json.RawMessage) (json.RawMessage, error)
```

**Understanding this signature:**

- **Input:** `json.RawMessage` - the raw JSON bytes of validated arguments
- **Output:** `json.RawMessage, error` - returns processed arguments as JSON, or an error

Notice the similarity to `CommandValidator`:

```go
type CommandValidator func(json.RawMessage) error
type CommandProcessor func(json.RawMessage) (json.RawMessage, error)
```

Both take raw JSON as input, but the processor also returns processed JSON.

## Update validCommands Registry

Let's add the processor to our command registry. Update `command_api.go`:

```go
// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
	Validator CommandValidator
	Processor CommandProcessor  // NEW
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
		Processor: processShellcodeCommand,  // NEW
	},
}
```

**What changed:**

- Added `Processor CommandProcessor` field to the struct
- Mapped "shellcode" to `processShellcodeCommand` (which we'll create next)



## Create ShellcodeArgsAgent Type

Recall that in `models/types.go` we have `ShellcodeArgsClient` for arguments from the client:

```go
type ShellcodeArgsClient struct {
	FilePath   string `json:"file_path"`
	ExportName string `json:"export_name"`
}
```

But the agent won't receive a file path - it will receive the actual DLL data. Let's add a new type for arguments sent to the agent in `models/types.go`:

```go
type ShellcodeArgsAgent struct {
	ShellcodeBase64 string `json:"shellcode_base64"`
	ExportName      string `json:"export_name"`
}
```

**What changed:**

- `FilePath` → `ShellcodeBase64`
- `ExportName` stays the same

The processing step transforms the client arguments into agent arguments by:

- Reading the file at `FilePath`
- Converting its contents to base64
- Storing in `ShellcodeBase64`

## Implement processShellcodeCommand

Now let's implement the processor. Add this to `internal/control/shellcode.go`:

```go
// processShellcodeCommand reads the DLL file and converts to base64 to create arguments sent to agent
func processShellcodeCommand(rawArgs json.RawMessage) (json.RawMessage, error) {

	var clientArgs models.ShellcodeArgsClient

	if err := json.Unmarshal(rawArgs, &clientArgs); err != nil {
		return nil, fmt.Errorf("unmarshaling args: %w", err)
	}

	// Read the DLL file
	file, err := os.Open(clientArgs.FilePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Convert to base64
	shellcodeB64 := base64.StdEncoding.EncodeToString(fileBytes)

	// Create the arguments that will be sent to the agent
	agentArgs := models.ShellcodeArgsAgent{
		ShellcodeBase64: shellcodeB64,
		ExportName:      clientArgs.ExportName,
	}

	// Marshal arguments ready to be sent to agent
	processedJSON, err := json.Marshal(agentArgs)
	if err != nil {
		return nil, fmt.Errorf("marshaling processed args: %w", err)
	}

	log.Printf("Processed file: %s (%d bytes) -> base64 (%d chars)",
		clientArgs.FilePath, len(fileBytes), len(shellcodeB64))

	return processedJSON, nil
}
```


### Parse client arguments

```go
var clientArgs models.ShellcodeArgsClient
if err := json.Unmarshal(rawArgs, &clientArgs); err != nil {
    return nil, fmt.Errorf("unmarshaling args: %w", err)
}
```

Unmarshal the raw JSON into the client argument struct so we can access the fields.


### Open the file

```go
file, err := os.Open(clientArgs.FilePath)
if err != nil {
    return nil, fmt.Errorf("opening file: %w", err)
}
defer file.Close()
```

Open the DLL file at the specified path. The `defer` ensures the file is closed when the function returns.



### Read the file contents

```go
fileBytes, err := io.ReadAll(file)
if err != nil {
    return nil, fmt.Errorf("reading file: %w", err)
}
```

Read all bytes from the file into memory.


### Convert to base64

```go
shellcodeB64 := base64.StdEncoding.EncodeToString(fileBytes)
```

Convert the binary data to a base64 string. This allows us to safely transmit binary data over JSON/HTTP.


### Create agent arguments

```go
agentArgs := models.ShellcodeArgsAgent{
    ShellcodeBase64: shellcodeB64,
    ExportName:      clientArgs.ExportName,
}
```

Create a new struct with the transformed data. Notice:
- `FilePath` is replaced with `ShellcodeBase64`
- `ExportName` is copied directly


### Marshal to JSON

```go
processedJSON, err := json.Marshal(agentArgs)
if err != nil {
    return nil, fmt.Errorf("marshaling processed args: %w", err)
}
```

Convert the agent arguments struct back to JSON bytes.



### Log and return

```go
log.Printf("Processed file: %s (%d bytes) -> base64 (%d chars)",
    clientArgs.FilePath, len(fileBytes), len(shellcodeB64))
return processedJSON, nil
```

Log useful metrics and return the processed JSON.




## Update commandHandler

Now let's integrate processing into our command handler. Update `commandHandler` in `control_api.go`:

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

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)
}
```


Let's break down the new code...





### Call the processor

```go
processedArgs, err := cmdConfig.Processor(cmdClient.Arguments)
```

Call the processor function, passing the validated arguments. This returns either:
- Processed arguments as JSON bytes
- An error if processing fails



### Handle processing failures

```go
if err != nil {
    var commandInvalid = fmt.Sprintf("ERROR: Processing failed for '%s': %v", cmdClient.Command, err)
    log.Printf(commandInvalid)
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(commandInvalid)
    return
}
```

If processing fails (file can't be read, etc.), log the error, send a 400 status, and return early.


### Update the command
```go
cmdClient.Arguments = processedArgs
log.Printf("Processed command arguments: %s", cmdClient.Command)
```


Replace the original arguments (with file path) with the processed arguments (with base64 data). This is what will eventually be sent to the agent.



## Test

Let's test the complete flow!

**Start the server:**

```bash
go run ./cmd/server
```

**Send a valid command:**

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
2025/11/06 14:42:14 Received command: shellcode
2025/11/06 14:42:14 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/06 14:42:14 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/06 14:42:14 Processed command arguments: shellcode
```

**Analyzing the output:**

- Command was received and validated ✓
- File was read: 111,493 bytes
- Converted to base64: 148,660 characters (base64 is about 33% larger than binary)
- Arguments successfully processed ✓

## Understanding the Complete Flow

Now let's trace a command through the entire pipeline:

1. **Client sends:** `{"command": "shellcode", "data": {"file_path": "./payloads/calc.dll", "export_name": "LaunchCalc"}}`
2. **Server receives:** Parses JSON into `CommandClient` struct
3. **Validation:** Checks file exists, export name provided
4. **Processing:** Reads DLL, converts to base64
5. **Result:** `cmdClient.Arguments` now contains `{"shellcode_base64": "TVqQAAMAAAAEAAAA...", "export_name": "LaunchCalc"}`

The command is now ready to be queued for the agent!

## Why This Design?

This processing pattern gives us several advantages:

1. **Agent flexibility** - Agent doesn't need filesystem access to the same files as the server
2. **Network safety** - Base64 encoding ensures binary data survives JSON serialization
3. **Security** - File reading happens on the server where we can control access
4. **Validation** - We confirm the file exists and is readable before queuing
5. **Clean separation** - Client args vs. agent args are distinct types with clear purposes

## Conclusion

In this lesson, we've implemented command argument processing:

- Created a function type for processors (`CommandProcessor`)
- Created a separate argument type for agent-bound data (`ShellcodeArgsAgent`)
- Implemented shellcode-specific processing (file reading and base64 encoding)
- Integrated processing into the command handler
- Tested the complete validation → processing pipeline

We've now done everything on the server side after receiving a command:

- ✓ Verified the command exists
- ✓ Validated command-specific arguments
- ✓ Processed arguments into agent-ready format

In the next lesson, we'll implement the command queue so validated and processed commands can wait for the agent to check in.






___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./03_validate_argument.md" >}})
[|NEXT|]({{< ref "./05_queue.md" >}})