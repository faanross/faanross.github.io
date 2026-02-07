---
layout: course01
title: "Lesson 14: Command Validation and Processing"
---


## Solutions

- **Starting Code:** [lesson_14_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_14_begin)
- **Completed Code:** [lesson_14_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_14_end)

## Overview

Now that we have a command endpoint, we need to make it robust. Right now our server accepts any command string without checking if it's actually valid, and even valid commands aren't checked for proper arguments.

We want to follow the **"fail fast" principle** - validate immediately before doing any processing. This means we need to:

1. Check if the command keyword exists in our registry
2. Validate that the command has the required arguments
3. Process arguments into the format the agent expects

## What We'll Create

- Command registry with validators and processors in `command_api.go`
- `CommandValidator` and `CommandProcessor` function types
- `validateShellcodeCommand` and `processShellcodeCommand` functions
- `ShellcodeArgsAgent` type for agent-bound data
- Complete validation and processing pipeline in `commandHandler`

## Part 1: Command Registry and Existence Check

### Why Use a Map?

We need somewhere to store all our valid commands. The most efficient way to check if a command exists is to use a **map as a "set"**.

**Why a map and not a list?**

- **If you used a list (slice):** `[]string{"shellcode", "download", "upload"}`
    - To check if a command exists, you'd have to loop through the entire list
    - This is slow, especially with many commands (O(n) operation)
- **By using a map:** `map[string]struct{}{...}`
    - Checking if a key exists is almost instant: `if _, ok := myMap["shellcode"]`
    - Performance doesn't degrade as you add more commands (O(1) operation)

### Understanding Functions as Types in Go

Before we build the registry, let's understand an important Go concept: **functions can be types**.

In Go, you can define a function signature as a type:

```go
// Define a function type
type Adder func(int, int) int

// Any function matching this signature can be assigned to this type
func sum(a, b int) int {
    return a + b
}

// Use it
var myAdder Adder = sum
result := myAdder(5, 3) // returns 8
```

This allows us to store functions in maps, pass them as parameters, and create consistent interfaces for different implementations.

### Create the Complete Registry

In `internal/control/command_api.go`:

```go
// CommandValidator validates command-specific arguments
type CommandValidator func(json.RawMessage) error

// CommandProcessor processes command-specific arguments
type CommandProcessor func(json.RawMessage) (json.RawMessage, error)

// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
	Validator CommandValidator
	Processor CommandProcessor
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
		Processor: processShellcodeCommand,
	},
}
```

**Understanding the signatures:**

- `CommandValidator` takes raw JSON arguments and returns an error (nil if valid)
- `CommandProcessor` takes raw JSON and returns processed JSON (plus potential error)

## Part 2: Argument Validation

### Implement validateShellcodeCommand

Create `internal/control/shellcode.go`:

```go
// validateShellcodeCommand validates "shellcode" command arguments from client
func validateShellcodeCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("shellcode command requires arguments")
	}

	var args ShellcodeArgsClient

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("invalid argument format: %w", err)
	}

	if args.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}

	if args.ExportName == "" {
		return fmt.Errorf("export_name is required")
	}

	// Check if file exists
	if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", args.FilePath)
	}

	log.Printf("Validation passed: file_path=%s, export_name=%s", args.FilePath, args.ExportName)

	return nil
}
```

**Breaking it down:**

1. **Check arguments exist** - The shellcode command must have arguments
2. **Parse the JSON** - Unmarshal into our typed struct
3. **Validate FilePath** - Must not be empty
4. **Validate ExportName** - Must not be empty
5. **Check file exists** - Use `os.Stat` to verify the file is accessible

## Part 3: Argument Processing

### Why Process Arguments?

The client sends a file path, but we can't send a file path to the agent (the agent is on a different machine with a different filesystem). We need to:

1. Read the DLL file from disk
2. Convert the binary data to base64
3. Send the base64 string to the agent

### Create the Agent Arguments Type

Add this to `models/types.go`:

```go
// ShellcodeArgsAgent - what we send to the agent
type ShellcodeArgsAgent struct {
	ShellcodeBase64 string `json:"shellcode_base64"`
	ExportName      string `json:"export_name"`
}
```

Notice the transformation:
- `FilePath` becomes `ShellcodeBase64`
- `ExportName` stays the same

### Implement processShellcodeCommand

Add this to `internal/control/shellcode.go`:

```go
// processShellcodeCommand reads the DLL file and converts to base64
func processShellcodeCommand(rawArgs json.RawMessage) (json.RawMessage, error) {

	var clientArgs ShellcodeArgsClient

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
	agentArgs := ShellcodeArgsAgent{
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

## Part 4: Complete Command Handler

Now let's wire everything together in `commandHandler`:

```go
func commandHandler(w http.ResponseWriter, r *http.Request) {

	var cmdClient CommandClient

	if err := json.NewDecoder(r.Body).Decode(&cmdClient); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}

	cmdClient.Command = strings.ToLower(cmdClient.Command)
	var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
	log.Printf(commandReceived)

	// STEP 1: Check if command exists
	cmdConfig, exists := validCommands[cmdClient.Command]
	if !exists {
		var commandInvalid = fmt.Sprintf("ERROR: Unknown command: %s", cmdClient.Command)
		log.Printf(commandInvalid)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(commandInvalid)
		return
	}

	// STEP 2: Validate arguments (if validator exists)
	if cmdConfig.Validator != nil {
		if err := cmdConfig.Validator(cmdClient.Arguments); err != nil {
			var commandInvalid = fmt.Sprintf("ERROR: Validation failed for '%s': %v", cmdClient.Command, err)
			log.Printf(commandInvalid)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(commandInvalid)
			return
		}
	}

	// STEP 3: Process arguments (if processor exists)
	if cmdConfig.Processor != nil {
		processedArgs, err := cmdConfig.Processor(cmdClient.Arguments)
		if err != nil {
			var commandInvalid = fmt.Sprintf("ERROR: Processing failed for '%s': %v", cmdClient.Command, err)
			log.Printf(commandInvalid)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(commandInvalid)
			return
		}
		cmdClient.Arguments = processedArgs
		log.Printf("Processed command arguments: %s", cmdClient.Command)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)
}
```

## Test

**Start the server:**

```bash
go run ./cmd/server
```

**Test 1: Invalid command**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "derp"}'
```

**Expected:** `"ERROR: Unknown command: derp"`

**Test 2: Missing arguments**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "shellcode"}'
```

**Expected:** `"ERROR: Validation failed for 'shellcode': shellcode command requires arguments"`

**Test 3: Valid command with valid arguments**

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

**Expected server output:**

```bash
2025/11/06 14:42:14 Received command: shellcode
2025/11/06 14:42:14 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/06 14:42:14 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/06 14:42:14 Processed command arguments: shellcode
```

## Conclusion

In this lesson, we built the complete command validation and processing pipeline:

- Created function types for validators and processors
- Built a command registry with O(1) lookups
- Implemented shellcode-specific validation
- Implemented shellcode-specific processing (file reading + base64 encoding)
- Wired everything together in the command handler

The command is now validated and processed, ready to be queued for the agent. In the next lesson, we'll implement the command queue.

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/course01/lesson-13">← Previous: Lesson 13</a></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/lesson-15">Next: Lesson 15 →</a></div>
</div>