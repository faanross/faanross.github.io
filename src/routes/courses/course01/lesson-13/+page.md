---
layout: course01
title: "Lesson 13: Command Endpoint"
---


## Solutions

- **Starting Code:** [lesson_13_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_13_begin)
- **Completed Code:** [lesson_13_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_13_end)

## Overview

We want to add a new endpoint for our client API. Right now we have a `/switch` endpoint for protocol transitions. Now we need a `/command` endpoint where operators can submit commands to be executed by agents.

First, let's switch from the standard `net/http` library to Chi for our Control API (same as our HTTPS server):

```go
// StartControlAPI starts the control API server on port 8080
func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	r.Post("/switch", handleSwitch)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}

func handleSwitch(w http.ResponseWriter, r *http.Request) {

	Manager.TriggerTransition()

	response := "Protocol transition triggered"

	json.NewEncoder(w).Encode(response)
}
```

Now let's add the command endpoint:

```go
// StartControlAPI starts the control API server on port 8080
func StartControlAPI() {
	// Create Chi router
	r := chi.NewRouter()

	r.Post("/switch", handleSwitch)

	r.Post("/command", commandHandler)

	log.Println("Starting Control API on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Printf("Control API error: %v", err)
		}
	}()
}
```

## Create Command Types

Before implementing the handler, we need types to represent commands. Create `internal/control/models.go`:

```go
package control

import "encoding/json"

// CommandClient represents a command with its arguments as sent by Client
type CommandClient struct {
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"data,omitempty"`
}
```

**Understanding the structure:**

- `Command` - The command keyword (e.g., "shellcode", "download", "upload")
- `Arguments` - Command-specific arguments stored as raw JSON

The `Arguments` field is `json.RawMessage`, which is a special type in Go. It allows us to defer parsing of JSON data. Why? Because different commands will have different argument structures:

- A shellcode loader might need a file path and export name
- A download command might need a source and destination path
- An upload command might need different parameters entirely

By using `json.RawMessage`, we can store the arguments as raw JSON bytes, then parse them later based on which command we're processing.

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

- `FilePath` - The path to the DLL containing the shellcode (on the server)
- `ExportName` - The name of the exported function in the DLL that should be called

**Note on naming:** This type is specifically called `ShellcodeArgsClient` (not just `ShellcodeArgs`) because the arguments as received from the client won't be exactly the same when we send them to the agent. We'll need another type for that later.

## Implement commandHandler

Now we can implement our command handler in `control_api.go`:

```go
func commandHandler(w http.ResponseWriter, r *http.Request) {

	// Instantiate custom type to receive command from client
	var cmdClient CommandClient

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

### Breaking it down:

**First, we instantiate our custom type:**

```go
var cmdClient CommandClient
```

**Unmarshal JSON from the request body:**

```go
if err := json.NewDecoder(r.Body).Decode(&cmdClient); err != nil {
    log.Printf("ERROR: Failed to decode JSON: %v", err)
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode("error decoding JSON")
    return
}
```

If decoding fails (invalid JSON, wrong structure, etc.), we log the error, return a 400 Bad Request status, and send an error message to the client.

**Log the command on server side:**

```go
var commandReceived = fmt.Sprintf("Received command: %s", cmdClient.Command)
log.Printf(commandReceived)
```

**Send confirmation to the client:**

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

- Created a proper `/command` endpoint using POST
- Defined custom types to represent commands and their arguments
- Implemented a handler that can parse incoming command JSON
- Tested the entire flow with `curl`

In the next lesson, we'll add command validation to ensure that only valid commands are accepted by our server.

---

[Previous: Lesson 12 - Payload Encryption](/courses/course01/lesson-12) | [Next: Lesson 14 - Command Validation and Processing](/courses/course01/lesson-14) | [Course Home](/courses/course01)
