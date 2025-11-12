---
showTableOfContents: true
title: "Lesson 9: Create Shellcode Doer Interface and Implementations"
type: "page"
---

## Solutions

The starting solution can be found here.

The final solution can be found here.


## Overview

In the previous lesson, we created the orchestrator that prepares arguments and calls a doer. Now we need to properly implement the doer system using interfaces and OS-specific implementations.

**Why do we need an interface?**

Different operating systems have completely different APIs for loading and executing code:

- **Windows:** Uses PE format, Windows API (VirtualAlloc, LoadLibrary, etc.)
- **Linux:** Uses ELF format, different system calls
- **macOS:** Uses Mach-O format, different APIs

By using an interface, we can:

1. Define a common contract that all implementations must follow
2. Write OS-specific implementations using build tags
3. Let Go's build system automatically choose the right implementation
4. Keep our orchestrator code clean and OS-agnostic

In this lesson, we'll:

1. Understand why interfaces are necessary for cross-platform code
2. Create the proper interface definition
3. Create a stub macOS implementation for development/testing
4. Prepare the structure for the Windows implementation (next lesson)

## What We'll Create

- Clean interface definition in `interface_shellcode.go`
- Stub macOS implementation in `doer_shellcode_mac.go`
- Structure for Windows implementation (code in next lesson)



## Understanding the Interface Problem

So using an interface allows us to create multiple OS-specific implementations of a command like the shellcode loader. But there is another practical reasons why someone like me that uses Mac OS as my base OS, meaning that I develop on, has to use it in this case.

See, when I implement our Windows-specific implementation of the shellcode loader doer, I have to use Windows build tags otherwise it will error out. But, if I do that, it leaves me in another binds since then that file is essentially invisible to the rest of my code.

So let's say I just called it directly from my orchestrator like so...


```go
// orchestrator.go
func (agent *Agent) orchestrateShellcode(...) {
    // Process arguments...
    
    // This won't work! DoShellcodeWindows doesn't exist on Mac
    result := DoShellcodeWindows(rawShellcode, exportName)
    
    // Return result...
}
```

The issue here is that since `DoShellcodeWindows()` has Windows build tags it won't be found. When developing on macOS (or Linux for that matter), the Windows file is completely invisible to the compiler. So when your orchestrator tries to call `DoShellcode()`, it has no idea that function will exist when compiled for Windows. This causes confusing compilation errors.

So instead, what we do is we:
- Create an interface
- Create a type that satisfies the interface
- Create OS-specific method implementations that we call on the type


```go
// interface_shellcode.go (no build tags - always compiled)
type CommandShellcode interface {
    DoShellcode(...) (models.ShellcodeResult, error)
}

// doer_shellcode_win.go
//go:build windows
type windowsShellcode struct{}
func (ws *windowsShellcode) DoShellcode(...) { /* Windows impl */ }
func New() CommandShellcode { return &windowsShellcode{} }

// doer_shellcode_mac.go
//go:build darwin
type macShellcode struct{}
func (ms *macShellcode) DoShellcode(...) { /* Mac impl */ }
func New() CommandShellcode { return &macShellcode{} }

// orchestrator.go
func orchestrateShellcode(...) {
    shellcode := shellcode.New()  // Returns CommandShellcode interface
    result := shellcode.DoShellcode(...)  // Calls OS-specific implementation
}
```


Then, instead of calling any OS-specific implementation from our orchestrator, we instead call the interface method on the type. This will works since the interface is always visible (no build tags), so the orchestrator knows about the method. The `New()` constructor exists in all OS files, returning the appropriate implementation.



## Create ShellcodeResult Type


First thing, we need a command-specific type for the results, which our doer will return to the orchestrator. So let's define the following in `models/types.go`:

```go
// ShellcodeResult represents the result of shellcode execution
type ShellcodeResult struct {
	Message string `json:"message"`
}
```

**Why so simple?**

Shellcode execution doesn't produce output like a shell command would. It either:

- Succeeds (shellcode runs)
- Fails (something went wrong)

The message field just provides context about what happened. For other commands (like downloading files), this struct might contain much more data.

We're now ready to create our actual interface.

## The Interface File

Create the following file `internal/shellcode/interface_shellcode.go` and add this interface:

```go
// CommandShellcode is the interface for shellcode execution
type CommandShellcode interface {
	DoShellcode(dllBytes []byte, exportName string) (models.ShellcodeResult, error)
}
```

**Key points:**

1. **No build tags** - This file is compiled on all platforms
2. **Defines the contract** - Any type with this method satisfies the interface
3. **Return types are consistent** - All implementations return the same types

**Understanding the signature:**

- **Input 1:** `dllBytes []byte` - The raw DLL binary data (already decoded from base64)
- **Input 2:** `exportName string` - The function to call within the DLL
- **Output 1:** `models.ShellcodeResult` - Contains status message
- **Output 2:** `error` - Error if execution failed, nil if successful




## The macOS Stub Implementation

We can now create our Mac OS-specific implementation of the interface. Note that if you are working on Linux, feel free to adapt this and create a Linux-specific implementation, since it's a stub there is no real OS-specific logic, this way at least we get to test it at the end of this lesson!


Create the following file `internal/shellcode/doer_shellcode_mac.go`:

```go
//go:build darwin

package shellcode

import (
	"fmt"
	"workshop3_dev/internals/models"
)

// macShellcode implements the CommandShellcode interface for Darwin/MacOS
type macShellcode struct{}

// New is the constructor for our Mac-specific Shellcode command
func New() CommandShellcode {
	return &macShellcode{}
}

// DoShellcode is the stub implementation for macOS
func (ms *macShellcode) DoShellcode(dllBytes []byte, exportName string) (models.ShellcodeResult, error) {
	fmt.Println("|❗ SHELLCODE DOER MACOS| This feature has not yet been implemented for MacOS.")

	result := models.ShellcodeResult{
		Message: "FAILURE: Not implemented on macOS",
	}
	return result, nil
}
```



### **Build constraint**

```go
//go:build darwin
```

This file is ONLY compiled when building for macOS (Darwin is the kernel name for macOS).



### Implementation struct

```go
type macShellcode struct{}
```

An empty struct that will satisfy the interface. It doesn't need any fields because the stub doesn't maintain state.


### Constructor
```go
func New() CommandShellcode {
    return &macShellcode{}
}
```

Returns a pointer to macShellcode. The return type is the interface, not the concrete type. This is important - it means callers work with the interface, not the specific implementation.


### Interface implementation

```go
func (ms *macShellcode) DoShellcode(dllBytes []byte, exportName string) (models.ShellcodeResult, error)
```

This method signature matches the interface exactly, so `macShellcode` satisfies the `CommandShellcode` interface.


### Stub behavior

```go
fmt.Println("|❗ SHELLCODE DOER MACOS| This feature has not yet been implemented for MacOS.")
    
result := models.ShellcodeResult{
     Message: "FAILURE: Not implemented on macOS",
}

return result, nil
```

Just prints a warning and returns a "not implemented" message. Notice we return `nil` for the error - this isn't an error in execution, it's just that the feature doesn't exist on this platform.






## Understanding Build Tags in Detail

Let's understand how Go's build system uses these tags:

**When compiling on macOS:**

```bash
go build ./cmd/agent
```

Go sees:

- ✓ `interface_shellcode.go` - NO build tags → Compiled
- ✓ `doer_shellcode_mac.go` - `//go:build darwin` → Compiled (we're on darwin)
- ✗ `doer_shellcode_win.go` - `//go:build windows` → NOT compiled (we're not on windows)

**When compiling for Windows (cross-compile from macOS):**

```bash
GOOS=windows GOARCH=amd64 go build ./cmd/agent
```

Go sees:

- ✓ `interface_shellcode.go` - NO build tags → Compiled
- ✗ `doer_shellcode_mac.go` - `//go:build darwin` → NOT compiled (target is windows)
- ✓ `doer_shellcode_win.go` - `//go:build windows` → Compiled (target is windows)

**The magic:** Both files define a `New()` function, but only one is ever compiled. The orchestrator calls `shellcode.New()`, and Go automatically uses whichever implementation is compiled for the target OS.

## Why Return `nil` for Error in the Stub?

You might wonder why the macOS stub returns `nil` for the error:

```go
return result, nil  // Why nil?
```

There are two philosophies we could follow:

**Option 1: Return an error (not implemented is an error)**

```go
return result, errors.New("not implemented on macOS")
```

This would cause the orchestrator to mark the task as failed.

**Option 2: Return nil (not implemented is a status, not an error)**

```go
return result, nil
```

This allows the task to "succeed" but with a message indicating it's not implemented.

We chose Option 2 because:

1. It's not an error in execution - the code ran fine
2. The message clearly indicates the feature isn't available
3. For testing, it's useful to see the full flow complete

In a production system, you might choose Option 1 to make it clear that the command didn't actually execute.




## Preparing for Windows Implementation

In the next lesson, we'll create `internal/shellcode/doer_shellcode_win.go` which will have the same structure:

```go
//go:build windows

package shellcode

import (
	// Windows-specific imports...
	"workshop3_dev/internals/models"
)

// windowsShellcode implements the CommandShellcode interface for Windows
type windowsShellcode struct{}

// New is the constructor for our Windows-specific Shellcode command
func New() CommandShellcode {
	return &windowsShellcode{}
}

// DoShellcode performs reflective DLL loading on Windows
func (ws *windowsShellcode) DoShellcode(dllBytes []byte, exportName string) (models.ShellcodeResult, error) {
	// COMPLEX WINDOWS IMPLEMENTATION HERE
	// - Parse PE headers
	// - Allocate memory
	// - Map sections
	// - Process relocations
	// - Resolve imports
	// - Call DllMain
	// - Call exported function
	
	return result, nil
}
```

The structure is identical to the macOS version:

- ✓ Build tag (`//go:build windows`)
- ✓ Implementation struct (`windowsShellcode`)
- ✓ Constructor returning interface (`New()`)
- ✓ Method implementing interface (`DoShellcode()`)

But the implementation will be much more complex (hundreds of lines of Windows PE loading code).

## Test Again

Let's verify that everything works, even if we're just calling a stub.

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Queue a command:**

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

**Expected agent output (on macOS):**

```bash
2025/11/07 08:44:15 Job received from Server
-> Command: shellcode
-> JobID: job_840709
2025/11/07 08:44:15 AGENT IS NOW PROCESSING COMMAND shellcode with ID job_840709
2025/11/07 08:44:15 |✅ SHELLCODE ORCHESTRATOR| Task ID: job_840709. Executing Shellcode, Export Function: LaunchCalc, ShellcodeLen(b64)=148660
|❗ SHELLCODE DOER MACOS| This feature has not yet been implemented for MacOS.
2025/11/07 08:44:15 |👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully for TaskID job_840709. Loader Message: FAILURE: Not implemented on macOS
2025/11/07 08:44:15 |AGENT TASK|-> Sending result for Task ID job_840709 (66 bytes)...
2025/11/07 08:44:15 |RETURN RESULTS|-> Sending 66 bytes of results via POST to https://0.0.0.0:8443/results
2025/11/07 08:44:15 💥 SUCCESSFULLY SENT FINAL RESULTS BACK TO SERVER.
2025/11/07 08:44:15 |AGENT TASK|-> Successfully sent result for Task ID job_840709.
```


**Analyzing the output:**

1. Job received ✓
2. ExecuteTask called ✓
3. Orchestrator unpacked and validated arguments ✓
4. Base64 decoded (148660 chars) ✓
5. Doer called (stub implementation) ✓
6. Result marshaled (66 bytes) ✓
7. Result sent to server ✓

Perfect! The entire flow is working. The result is being sent to the `/results` endpoint, but since that doesn't exist on the server yet, we don't see a response. We'll create it in a future lesson.



## Conclusion

In this lesson, we've created a robust cross-platform architecture:

- Understood why interfaces are necessary for OS-specific code
- Implemented our interface definition (platform-agnostic)
- Implemented our macOS stub implementation
- Understood build tags and how they work
- Learned why `New()` returns the interface type
- Tested the complete stub flow

Our system now has:

- ✓ Clean interface definition
- ✓ Platform-specific implementations using build tags
- ✓ Testable on any platform
- ✓ Ready for Windows implementation

In the next lesson, we'll implement the actual Windows shellcode loader - the complex part where we'll parse PE files, allocate memory, process relocations, resolve imports, and execute the shellcode!





___
[|TOC|]({{< ref "./moc.md" >}})
[|PREV|]({{< ref "./08_orchestrator.md" >}})
[|NEXT|]({{< ref "./10_doer.md" >}})