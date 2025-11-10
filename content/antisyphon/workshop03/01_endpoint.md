---
showTableOfContents: true
title: "Lesson 1: Implement Command Endpoint"
type: "page"
---
# Lesson 1: Implement Command Endpoint

## Solutions

The starting solution can be found here.

The final solution can be found here

## Overview

Right now we have a dummy endpoint that our client can hit via `curl`. It doesn't really do anything - nothing happens. The server just prints a message and returns a simple response.

In this lesson, we'll transform this into a real command endpoint that will eventually receive commands from an operator.

We'll:
1. Change the endpoint from `/dummy` to `/command`
2. Switch from GET to POST (since we'll send command data in the request body)
3. Create custom types to represent commands
4. Parse incoming command JSON and extract the command keyword

This is the first step in building our command and control infrastructure. By the end of this lesson, we'll be able to send commands via curl and see the server successfully parse and display them.

## What We'll Create

- Command endpoint (`/command`) in `control_api.go`
- Custom types to represent commands in `models/types.go`
- Command handler that parses incoming JSON

## Review Current Code

Let's look at what we're starting with in `control_api.go`:

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

This contains two functions:

- `StartControlAPI()` is called from our server's main. It starts the control API listener on port 8080 and makes one endpoint available: `/dummy`
- `dummyHandler` is the function that gets called when the endpoint is hit



## Update StartControlAPI()

First, let's change the endpoint name to something more appropriate like `/command`. We'll also change from GET to POST since we'll be sending command data in the request body:

```go
func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	// Define the POST endpoint
	r.Post("/command", commandHandler) // CHANGED: endpoint and method

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}
```

**What changed:**

- Endpoint path: `/dummy` → `/command`
- HTTP method: `Get` → `Post`
- Handler function: `dummyHandler` → `commandHandler`

## Create Custom Types

Before we implement our new `commandHandler`, we need a custom type to represent the commands we'll receive from the client.

We'll have many custom types throughout this project, so it's good practice to group them together. Let's create a new file: `internal/models/types.go`

Add the following:

```go
package models

import "encoding/json"

// CommandClient represents a command with its arguments as sent by Client
type CommandClient struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}
```

**Understanding the structure:**

- `Command` - The command keyword (e.g., "shellcode", "download", "upload")
- `Arguments` - Command-specific arguments stored as raw JSON

The `Arguments` field is `json.RawMessage`, which is a special type in Go. It allows us to defer parsing of JSON data. Why? Because different commands will have different argument structures:

- A shellcode loader might need a file path and export name
- A download command might need a source and destination path
- An upload command might need different parameters entirely

By using `json.RawMessage`, we can store the arguments as raw JSON bytes, then parse them later based on which command we're processing.


## Add Shellcode-Specific Arguments

While we're here, let's also add the specific argument struct for our shellcode loader command:

```go
// ShellcodeArgsClient contains the command-specific arguments for Shellcode Loader as sent by Client
type ShellcodeArgsClient struct {
	FilePath   string `json:"file_path"`
	ExportName string `json:"export_name"`
}
```

**Understanding the fields:**

- `FilePath` - The path to the DLL containing the shellcode (on the server)
- `ExportName` - The name of the exported function in the DLL that should be called

We'll discuss these in more detail when we implement validation. For now, just be aware that this is how the client will specify which DLL to load and which function to call.

**Note on naming:** This type is specifically called `ShellcodeArgsClient` (not just `ShellcodeArgs`) because the arguments as received from the client won't be exactly the same when we send them to the agent. We'll need another type for that later - more on this in future lessons.

## Implement commandHandler

Now we can implement our command handler back in `control_api.go`:

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

	// Visually confirm we get the command we expected
	var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
	log.Printf(commandReceived)

	// Confirm on the client side command was received
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commandReceived)
}
```


### First, we instantiate our custom type

- This will hold the parsed command

```go
var cmdClient models.CommandClient
```


### Unmarshall JSON
-  From the request body into our struct:
```go
	if err := json.NewDecoder(r.Body).Decode(&cmdClient); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}
```

If decoding fails (invalid JSON, wrong structure, etc.), we log the error, return a 400 Bad Request status, and send an error message to the client.


### For now let's simply log the command on server side

- This just helps us to confirm things are working as they should at this point.

```go
	var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
	log.Printf(commandReceived)
```



### Similarly, we send confirmation to the client

```go
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(commandReceived)
```




## Test

Let's test our new command endpoint!

**Start the server:**

```bash
go run ./cmd/server
```

You should see:

```bash
2025/11/04 13:44:38 Starting Control API on :8080
2025/11/04 13:44:38 Starting server on 127.0.0.1:8443
```

**Send a command using curl:**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "shellcode"}'
```

**Expected client-side response:**

```bash
"Received command: shellcode"
```

**Expected server-side output:**

```bash
2025/11/04 13:44:56 Received command: shellcode
```

Perfect! We can now successfully send commands to our server and have them parsed correctly.

## Conclusion

In this lesson, we've laid the groundwork for our command infrastructure:

- Created a proper `/command` endpoint using POST
- Defined custom types to represent commands and their arguments
- Implemented a handler that can parse incoming command JSON
- Tested the entire flow with `curl`

In the next lesson, we'll add command validation to ensure that only valid commands are accepted by our server.





___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./00B_starting.md" >}})
[|NEXT|]({{< ref "./02_validate_command.md" >}})