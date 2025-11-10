---
showTableOfContents: true
title: "Lesson 2: Validate Command Exists"
type: "page"
---

## Solutions

The starting solution can be found here.

The final solution can be found here.

## Overview

Right now our server accepts any command string without checking if it's actually valid. We want to follow the **"fail fast" principle** - validate immediately before doing any processing.

In this lesson, we'll:

1. Create a registry of valid commands
2. Check incoming commands against this registry
3. Reject invalid commands immediately
4. Only proceed with valid commands

This ensures we don't waste resources processing commands that don't exist, and gives clear feedback to the operator when they make a mistake.

## What We'll Create

- Command registry using a map in `command_api.go`
- Validation logic in `commandHandler`
- Proper error responses for invalid commands

## Why Use a Map?

We need somewhere to store all our valid commands. The most efficient way to check if a command exists is to use a **map as a "set"**.

**Why a map and not a list?**

- **If you used a list (slice):** `[]string{"shellcode", "download", "upload"}`
    - To check if a command exists, you'd have to loop through the entire list
    - This is slow, especially with many commands (O(n) operation)
- **By using a map:** `map[string]struct{}{...}`
    - Checking if a key exists is almost instant: `if _, ok := myMap["shellcode"]`
    - Performance doesn't degrade as you add more commands (O(1) operation)

**Why use `struct{}` as the value?**

The most memory-efficient and idiomatic way in Go to create a set is to use an **empty struct (`struct{}`)** as the value, because it takes up **zero memory**.

Later, we'll actually use this map for more than just lookups - we'll add validator and processor functions to the struct. But for now, it's empty.

## Create Command Registry

Let's create a new file: `internal/control/command_api.go`

Add the following:

```go
package control

// Registry of valid commands with their validators and processors
var validCommands = map[string]struct{}{
	"shellcode": {},
}
```

**Understanding this code:**

- The **key** is a string: the command name
- The **value** is an empty struct: `struct{}{}`
- This is the idiomatic Go way for "lookups" - using an empty struct as explained above
- We'll expand this struct later to include validator and processor functions

For now, we only have one valid command: `shellcode`. As we add more commands in the future, we'll simply add more entries to this map.



## Update commandHandler

Now let's modify our command handler in `control_api.go` to validate against this registry.

First, let's **normalize** the command to lowercase to ensure consistency (so "Shellcode", "SHELLCODE", "shellcode" etc are all accepted):

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
	_, exists := validCommands[cmdClient.Command]
	if !exists {
		var commandInvalid = fmt.Sprintf("ERROR: Unknown command: %s", cmdClient.Command)
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


Let's look at our new code...


### Normalize the command
```go
cmdClient.Command = strings.ToLower(cmdClient.Command)
```


### Check if command exists in registry

```go
_, exists := validCommands[cmdClient.Command]
```

**Understanding Go map lookups:** When you access a map in Go, you get two values:
- The first value is what's stored at that key (we don't need it, so we use `_`)
- The second value is a boolean indicating whether the key exists


So `_, exists := validCommands[cmdClient.Command]` means:
- "Look up this command in the map"
- "I don't care about the value (empty struct), just tell me if it exists"
- "Store the existence boolean in the `exists` variable"


### Reject invalid commands

```go
    if !exists {
        var commandInvalid = fmt.Sprintf("ERROR: Unknown command: %s", cmdClient.Command)
        log.Printf(commandInvalid)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(commandInvalid)
        return
    }
```


**If the command doesn't exist, we:**
- Log the error on the server
- Send a 400 Bad Request status
- Send a clear error message to the client
- Return early (stop processing)



## Test

Let's test both valid and invalid commands!

**Start the server:**

```bash
go run ./cmd/server
```

**Test 1: Invalid command**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "derp"}'
```

**Expected client-side response:**

```bash
"ERROR: Unknown command: derp"
```

**Expected server-side output:**

```bash
2025/11/04 14:44:46 Received command: derp
2025/11/04 14:44:46 ERROR: Unknown command: derp
```

**Test 2: Valid command**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "shellcode"}'
```

**Expected client-side response:**

```bash
"Received command: shellcode"
```

**Expected server-side output:**

```bash
2025/11/04 14:44:50 Received command: shellcode
```

Perfect! Our server now validates commands and rejects invalid ones immediately.

## Understanding the Benefits

This validation pattern gives us several advantages:

1. **Fast rejection** - Invalid commands are caught immediately before any processing
2. **Clear feedback** - Operators know immediately when they've typed a wrong command
3. **Efficient checking** - O(1) lookup time regardless of how many commands we add
4. **Extensible** - We can easily add new commands by adding entries to the map
5. **Type-safe** - The map will later hold command-specific functions for validation and processing

## Conclusion

In this lesson, we've implemented command validation:

- Created a command registry using a map
- Added validation logic to check incoming commands
- Provided clear error messages for invalid commands
- Tested both valid and invalid commands

In the next lesson, we'll add command-specific argument validation to ensure that each command receives the correct parameters.



___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./01_endpoint.md" >}})
[|NEXT|]({{< ref "./03_validate_argument.md" >}})