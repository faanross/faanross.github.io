---
showTableOfContents: true
title: "Lesson 8: Implement Shellcode Orchestrator"
type: "page"
---

## Solutions

The starting solution can be found here.

The final solution can be found here.


## Overview

We have the execution framework in place, but no actual command implementations. In this lesson, we'll create the orchestrator for the shellcode command.

The orchestrator's responsibilities:

1. Unpack the `ServerResponse` to extract command-specific arguments
2. Validate arguments on the agent side
3. Decode the base64 shellcode data back to raw bytes
4. Call the OS-specific "doer" to execute the shellcode
5. Handle results and errors
6. Return an `AgentTaskResult` to `ExecuteTask`

We'll implement the orchestrator in this lesson. The actual shellcode doer (the complex part) will come in the next lessons.

## What We'll Create

- `orchestrateShellcode()` method in `agent/orchestrator.go`
- Registration of the shellcode command

## Create the Orchestrator

Create a new file `agent/orchestrator.go`:

```go
// orchestrateShellcode is the orchestrator for the "shellcode" command
func (agent *Agent) orchestrateShellcode(job *models.ServerResponse) models.AgentTaskResult {

	// Create an instance of the shellcode args struct
	var shellcodeArgs models.ShellcodeArgsAgent

	// ServerResponse.Arguments contains the command-specific args, so now we unmarshal the field into the struct
	if err := json.Unmarshal(job.Arguments, &shellcodeArgs); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal ShellcodeArgs for Task ID %s: %v. ", job.JobID, err)
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("failed to unmarshal ShellcodeArgs"),
		}
	}
	log.Printf("|✅ SHELLCODE ORCHESTRATOR| Task ID: %s. Executing Shellcode, Export Function: %s, ShellcodeLen(b64)=%d\n",
		job.JobID, shellcodeArgs.ExportName, len(shellcodeArgs.ShellcodeBase64))

	// Some basic agent-side validation
	if shellcodeArgs.ShellcodeBase64 == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ShellcodeBase64 is empty.", job.JobID)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("ShellcodeBase64 cannot be empty"),
		}
	}

	if shellcodeArgs.ExportName == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ExportName is empty.", job.JobID)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("ExportName must be specified for DLL execution"),
		}
	}

	// Now let's decode our b64
	rawShellcode, err := base64.StdEncoding.DecodeString(shellcodeArgs.ShellcodeBase64)
	if err != nil {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: Failed to decode ShellcodeBase64: %v", job.JobID, err)
		return models.AgentTaskResult{
			JobID:   job.JobID,
			Success: false,
			Error:   errors.New("Failed to decode shellcode"),
		}
	}

	// Call the "doer" function
	commandShellcode := shellcode.New()
	shellcodeResult, err := commandShellcode.DoShellcode(rawShellcode, shellcodeArgs.ExportName)

	finalResult := models.AgentTaskResult{
		JobID: job.JobID,
		// Output will be set below after JSON encoding
	}

	outputJSON, _ := json.Marshal(string(shellcodeResult.Message))

	finalResult.CommandResult = outputJSON

	if err != nil {
		loaderError := fmt.Sprintf("|❗ERR SHELLCODE ORCHESTRATOR| Loader execution error for TaskID %s: %v. Loader Message: %s",
			job.JobID, err, shellcodeResult.Message)
		log.Printf(loaderError)
		finalResult.Error = errors.New(loaderError)
		finalResult.Success = false

	} else {
		log.Printf("|👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully for TaskID %s. Loader Message: %s",
			job.JobID, shellcodeResult.Message)
		finalResult.Success = true
	}

	return finalResult
}
```

That's quite a lot of new code, so let's break it down bit by bit


### Step 1: Unmarshal Arguments

```go
// Create an instance of the shellcode args struct
var shellcodeArgs models.ShellcodeArgsAgent

// ServerResponse.Arguments contains the command-specific args, so now we unmarshal the field into the struct
if err := json.Unmarshal(job.Arguments, &shellcodeArgs); err != nil {
	errMsg := fmt.Sprintf("Failed to unmarshal ShellcodeArgs for Task ID %s: %v. ", job.JobID, err)
	log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| %s", errMsg)
	return models.AgentTaskResult{
		JobID:   job.JobID,
		Success: false,
		Error:   errors.New("failed to unmarshal ShellcodeArgs"),
	}
}
```

**Create struct instance**

```go
 var shellcodeArgs models.ShellcodeArgsAgent
```

Here we prepare a struct to hold the parsed arguments.

**Unmarshal the arguments**

```go
json.Unmarshal(job.Arguments, &shellcodeArgs)
```

Remember, `job.Arguments` is `json.RawMessage` (raw JSON bytes). We unmarshal it into our typed struct so we can access the fields.


**Handle errors** 
- If unmarshaling fails (corrupted data, wrong structure, etc.), we immediately return a failure result with the job ID and error message.


**Log success**
```go
log.Printf("|✅ SHELLCODE ORCHESTRATOR| Task ID: %s. Executing Shellcode, Export Function: %s, ShellcodeLen(b64)=%d\n",
        job.JobID, shellcodeArgs.ExportName, len(shellcodeArgs.ShellcodeBase64))
```

Provide visibility into what we're about to execute.



### Step 2: Agent-Side Validation

```go
// Some basic agent-side validation
if shellcodeArgs.ShellcodeBase64 == "" {
	log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ShellcodeBase64 is empty.", job.JobID)
	return models.AgentTaskResult{
		JobID:   job.JobID,
		Success: false,
		Error:   errors.New("ShellcodeBase64 cannot be empty"),
	}
}

if shellcodeArgs.ExportName == "" {
	log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ExportName is empty.", job.JobID)
	return models.AgentTaskResult{
		JobID:   job.JobID,
		Success: false,
		Error:   errors.New("ExportName must be specified for DLL execution"),
	}
}
```

**Why validate again on the agent?**

We already validated on the server, but we validate again here as a **defense-in-depth** measure:

1. **Data corruption:** Arguments could be corrupted during transmission
2. **Direct agent access:** In some scenarios, an agent might be controlled directly (bypassing the server)
3. **Safety:** Better to fail fast than to execute with bad data
4. **Clear errors:** Agent-side validation provides specific error messages in the agent's context

This is good engineering practice - don't trust that validation happened elsewhere.




## Step 3: Decode Base64

```go
// Now let's decode our b64
rawShellcode, err := base64.StdEncoding.DecodeString(shellcodeArgs.ShellcodeBase64)
if err != nil {
	log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: Failed to decode ShellcodeBase64: %v", job.JobID, err)
	return models.AgentTaskResult{
		JobID:   job.JobID,
		Success: false,
		Error:   errors.New("Failed to decode shellcode"),
	}
}
```



**What's happening:**

Transform the base64 string back into raw bytes:

- **Input:** `"TVqQAAMAAAAEAAAA..."` (base64 string)
- **Output:** `[]byte{0x4D, 0x5A, 0x90, ...}` (raw DLL bytes)

The `base64.StdEncoding.DecodeString()` function:

- Takes a base64-encoded string
- Returns the original binary data as a byte slice
- Returns an error if the string is invalid base64

If decoding fails, it means the data was corrupted or wasn't actually base64.



## Step 4: Call the Doer

```go
// Call the "doer" function
commandShellcode := shellcode.New()
shellcodeResult, err := commandShellcode.DoShellcode(rawShellcode, shellcodeArgs.ExportName)
```


**Create doer instance**

```go
commandShellcode := shellcode.New()
```

Call the constructor for the shellcode doer. This returns an OS-specific implementation (we'll create this in the next lessons).


**Call the doer**

```go
shellcodeResult, err := commandShellcode.DoShellcode(rawShellcode, shellcodeArgs.ExportName)
```

Pass the raw DLL bytes and export name to the doer. Returns:
- `shellcodeResult` - Contains a message about what happened
- `err` - Error if execution failed

The doer is where the actual shellcode loading and execution happens.



## Step 5: Build Result

```go
finalResult := models.AgentTaskResult{
	JobID: job.JobID,
	// Output will be set below after JSON encoding
}

outputJSON, _ := json.Marshal(string(shellcodeResult.Message))

finalResult.CommandResult = outputJSON

if err != nil {
	loaderError := fmt.Sprintf("|❗ERR SHELLCODE ORCHESTRATOR| Loader execution error for TaskID %s: %v. Loader Message: %s",
		job.JobID, err, shellcodeResult.Message)
	log.Printf(loaderError)
	finalResult.Error = errors.New(loaderError)
	finalResult.Success = false

} else {
	log.Printf("|👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully for TaskID %s. Loader Message: %s",
		job.JobID, shellcodeResult.Message)
	finalResult.Success = true
}

return finalResult
```


**Create base result**
```go
    finalResult := models.AgentTaskResult{
        JobID: job.JobID,
    }
```
Start with the job ID (for correlation).



**Marshal the message**
```go
outputJSON, _ := json.Marshal(string(shellcodeResult.Message))
finalResult.CommandResult = outputJSON
```
Convert the doer's message into JSON and store in `CommandResult`.


**Handle error case**
```go
    if err != nil {
        loaderError := fmt.Sprintf("...")
        log.Printf(loaderError)
        finalResult.Error = errors.New(loaderError)
        finalResult.Success = false
    }
```

If the doer returned an error, log it and mark the result as failed.


**Handle success case**

```go
    else {
        log.Printf("|👊 SHELLCODE SUCCESS| ...")
        finalResult.Success = true
    }
```

If no error, log success and mark the result as successful.


**Return**

```go
return finalResult
```

Return the complete result to `ExecuteTask`, which will marshal and send it to the server.




## Register the Command

Now we can uncomment the registration line. Update `registerCommands()` in `agent/commands.go`:

```go
func registerCommands(agent *Agent) {
	agent.commandOrchestrators["shellcode"] = (*Agent).orchestrateShellcode
	// Register other commands here in the future
}
```

**What changed:** We uncommented the line that maps "shellcode" to our orchestrator method using method expression.

## Test

We'll once again not be able to test since we have some dangling threads. From the next lesson on this will no longer be the case :)


## Conclusion

In this lesson, we've implemented our shellcode orchestrator:

- Created `orchestrateShellcode()` with complete argument handling
- Implemented agent-side validation (defense in depth)
- Decoded base64 back to raw bytes
- Called the doer interface (not yet implemented)
- Built and returned proper `AgentTaskResult`
- Registered the shellcode command


In the next two lessons, we'll create the actual shellcode doer interface and stub implementations for different operating systems!




___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./07_execute_task.md" >}})
[|NEXT|]({{< ref "./09_interface.md" >}})