---
showTableOfContents: true
title: "Lesson 3: Validate Command Arguments"
type: "page"
---
## Solutions

The starting solution can be found here.

The final solution can be found here.

## Overview

Now that we can validate that a command exists, we need to validate that it has the correct arguments. Each command will have its own specific requirements:

- **shellcode** needs a file path and export name
- **download** (future) might need source and destination paths
- **upload** (future) might need different parameters

In this lesson, we'll:

1. Create a function type for command validators
2. Map each command to its validator function
3. Implement the validator for the shellcode command
4. Integrate validation into our command handler

## What We'll Create

- `CommandValidator` function type in `command_api.go`
- `validateShellcodeCommand` function in `shellcode.go`
- Updated command registry with validators
- Validation logic in `commandHandler`

## Understanding Functions as Types in Go

Before we dive in, let's understand an important Go concept: **functions can be types**.

In Go, you can define a function signature as a type, just like you define struct types. This is incredibly powerful for creating flexible, extensible systems.

**Example:**

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

This allows us to store functions in maps, pass them as parameters, and create consistent interfaces for different implementations. We'll use this pattern extensively for our command system.

## Create CommandValidator Type

Let's define a function type that all command validators must follow. Add this to `command_api.go`:

```go
// CommandValidator validates command-specific arguments
type CommandValidator func(json.RawMessage) error
```

**Understanding this signature:**

- **Input:** `json.RawMessage` - the raw JSON bytes of the command arguments
- **Output:** `error` - returns `nil` if valid, or an error describing what's wrong

This type defines the "contract" that all validator functions must follow. Any function that:

- Takes a `json.RawMessage` as input
- Returns an `error`

...can be used as a `CommandValidator`.

## Update validCommands Registry

Now we can enhance our command registry to include validators. Update `command_api.go`:

```go
// Registry of valid commands with their validators and processors
var validCommands = map[string]struct {
	Validator CommandValidator
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
	},
}
```

**What changed:**

- The empty struct `struct{}{}` now has a field: `Validator CommandValidator`
- For the "shellcode" command, we map it to the `validateShellcodeCommand` function
- We haven't created `validateShellcodeCommand` yet, so this will error until we do

**Understanding the structure:**

```go
map[string]struct {
    Validator CommandValidator
}
```

This is a map where:

- **Key:** string (command name like "shellcode")
- **Value:** a struct containing a `Validator` field of type `CommandValidator`


## Create validateShellcodeCommand

Since shellcode-specific logic deserves its own file, let's create `internal/control/shellcode.go`:

```go
// validateShellcodeCommand validates "shellcode" command arguments from client
func validateShellcodeCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("shellcode command requires arguments")
	}

	var args models.ShellcodeArgsClient

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


### Check if arguments exist


```go
if len(rawArgs) == 0 {
    return fmt.Errorf("shellcode command requires arguments")
}
```


The shellcode command must have arguments - it can't be empty.


### Parse the arguments

```go
var args models.ShellcodeArgsClient

if err := json.Unmarshal(rawArgs, &args); err != nil {
     return fmt.Errorf("invalid argument format: %w", err)
}
```

Try to unmarshal the raw JSON into our `ShellcodeArgsClient` struct. If this fails, the JSON structure is wrong.


### Validate FilePath
```go
if args.FilePath == "" {
    return fmt.Errorf("file_path is required")
}
```

The file path must not be empty.


### Validate ExportName

```go
if args.ExportName == "" {
    return fmt.Errorf("export_name is required")
}
```


The export name must not be empty.

### Check if file exists

```go
if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
    return fmt.Errorf("file does not exist: %s", args.FilePath)
}
```

Use `os.Stat` to verify the file actually exists on the server's filesystem. If not, we can't load it later.



### Log success

```go
log.Printf("Validation passed: file_path=%s, export_name=%s", args.FilePath, args.ExportName)
return nil
```

If everything passes, log the validated parameters and return `nil` (no error).

**Understanding the arguments:**

- **FilePath:** The path on the **server** to the DLL containing the shellcode
- **ExportName:** The name of the exported function in the DLL we want to call

This design allows multiple exported functions per DLL, giving us greater flexibility.

## Update commandHandler

Now we need to actually call the validator in our command handler. Update `commandHandler` in `control_api.go`:

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
	cmdConfig, exists := validCommands[cmdClient.Command] // Changed: capture the command config
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

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)
}
```


Let's go through the new code/changes:

### Capture the command configuration

```go
cmdConfig, exists := validCommands[cmdClient.Command]
```

Previously we used `_` because we only needed the existence boolean. Now we need the actual struct containing the validator, so we capture it in `cmdConfig`.


### Call the validator

```go
if err := cmdConfig.Validator(cmdClient.Arguments); err != nil {
```

We call the validator function stored in `cmdConfig.Validator`, passing it the raw arguments from the command. Remember:
- `cmdConfig.Validator` is a function of type `CommandValidator`
- For "shellcode", this is `validateShellcodeCommand`
- We pass `cmdClient.Arguments` (the raw JSON)


### Handle validation failures

```go
var commandInvalid = fmt.Sprintf("ERROR: Validation failed for '%s': %v", cmdClient.Command, err)
log.Printf(commandInvalid)
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(commandInvalid)
return
```


If validation fails, we log the specific error, send a 400 status, send the error message to the client, and return early.


## Test

Let's test with both invalid and valid arguments!

**Start the server:**

```bash
go run ./cmd/server
```

**Test 1: File doesn't exist**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "shellcode",
    "data": {
      "file_path": "./payloads/derp.dll",
      "export_name": "LaunchCalc"
    }
  }'
```

**Expected client-side response:**

```bash
"ERROR: Validation failed for 'shellcode': file does not exist: ./payloads/derp.dll"
```

**Expected server-side output:**

```bash
2025/11/06 14:07:45 Received command: shellcode
2025/11/06 14:07:45 ERROR: Validation failed for 'shellcode': file does not exist: ./payloads/derp.dll
```

**Test 2: Valid arguments** (assuming you have `./payloads/calc.dll`)

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
2025/11/06 14:07:55 Received command: shellcode
2025/11/06 14:07:55 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
```

Perfect! Our server now validates not just that commands exist, but that they have correct, valid arguments.

## Why This Design?

This validation pattern gives us several advantages:

1. **Fail fast** - Bad arguments are caught immediately, before any processing
2. **Command-specific** - Each command can have its own validation logic
3. **Clear errors** - Users get specific error messages about what's wrong
4. **Server-side safety** - We verify files exist before attempting to load them
5. **Extensible** - Easy to add new commands with their own validators

The validation happens on the server side, but it's good practice to also validate on the agent side (which we'll do later) as a defense-in-depth measure - the arguments could potentially be corrupted over the wire.

## Conclusion

In this lesson, we've implemented command argument validation:

- Created a function type for validators (`CommandValidator`)
- Enhanced our command registry to store validator functions
- Implemented shellcode-specific validation
- Integrated validation into the command handler
- Tested with both invalid and valid arguments

In the next lesson, we'll add command argument **processing** - transforming arguments into the format needed by the agent.



___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./02_validate_command.md" >}})
[|NEXT|]({{< ref "./04_process_arguments.md" >}})