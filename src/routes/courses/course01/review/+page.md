---
layout: course01
title: "Course Review"
---

# Course Review

## What We Built

Congratulations! You've built a fully functional Command and Control framework from scratch. Let's review everything we've accomplished.

```
                    +---------------------+
                    |   CONTROL API       |
                    |   (Port 8080)       |
                    |  /command, /switch  |
                    +---------+-----------+
                              |
         +--------------------+--------------------+
         |                    |                    |
         v                    v                    v
+-----------------+  +-----------------+  +-----------------+
|  HTTPS Server   |  |   DNS Server    |  |  Command Queue  |
|  (Port 8443)    |  |  (Port 8443)    |  |                 |
+--------+--------+  +--------+--------+  +--------+--------+
         |                    |                    |
         +--------------------+--------------------+
                              |
                    +---------v---------+
                    |       AGENT       |
                    |  +-------------+  |
                    |  |   RunLoop   |  |
                    |  +------+------+  |
                    |         |         |
                    |  +------v------+  |
                    |  | ExecuteTask |  |
                    |  +------+------+  |
                    |         |         |
                    |  +------v------+  |
                    |  |Orchestrators|  |
                    |  +------+------+  |
                    |         |         |
                    |  +------v------+  |
                    |  |   Doers     |  |
                    |  +-------------+  |
                    +-------------------+
```

## Part 1: Foundation (Lesson 1)

We established the core design patterns:

- **Interfaces** - `Server` and `Agent` interfaces for protocol abstraction
- **Factory functions** - `NewServer()` and `NewAgent()` for dynamic creation
- **Polymorphism** - Different protocols implementing the same interface

```go
type Server interface {
    Start() error
    GetCommand() json.RawMessage
}

type Agent interface {
    Beacon() (json.RawMessage, error)
    SendResult(resultData []byte) error
}
```

## Part 2: HTTPS Communication (Lessons 2-4)

We built the primary communication channel:

- **HTTPS Server** - TLS-secured server with Chi router
- **HTTPS Agent** - Client with custom TLS configuration
- **Run Loop** - Periodic check-ins with jitter for evasion

```go
func (agent *HTTPSAgent) RunLoop() {
    for {
        response, _ := agent.Beacon()
        if hasJob(response) {
            agent.ExecuteTask(response)
        }
        time.Sleep(calculateJitter(agent.sleep, agent.jitter))
    }
}
```

## Part 3: DNS Communication (Lessons 5-7)

We added a covert channel:

- **DNS Server** - Using miekg/dns library
- **DNS Agent** - A record queries for data exchange
- **Unified Run Loop** - Both protocols sharing the same loop

## Part 4: Protocol Switching (Lessons 8-10)

We enabled dynamic protocol transitions:

- **Control API** - `/switch` endpoint for operator commands
- **Dual Server Startup** - Both servers running simultaneously
- **Transition Logic** - Agent-side protocol switching

```go
switch activeProtocol {
case "https":
    agent = NewHTTPSAgent(config)
case "dns":
    agent = NewDNSAgent(config)
}
```

## Part 5: Security Layer (Lessons 11-12)

We secured all communications:

- **HMAC Authentication** - Message integrity verification
- **AES-GCM Encryption** - Payload confidentiality

```go
// HMAC for authentication
mac := hmac.New(sha256.New, key)
mac.Write(message)
signature := mac.Sum(nil)

// AES-GCM for encryption
block, _ := aes.NewCipher(key)
gcm, _ := cipher.NewGCM(block)
ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
```

## Part 6: Command Infrastructure (Lessons 13-16)

We built the server-side command system:

- **Command Endpoint** - `/command` for operator input
- **Validation & Processing** - Type-safe argument handling
- **Command Queue** - Thread-safe pending job storage
- **Dequeue & Send** - Delivering commands to agents

```go
var validCommands = map[string]struct{
    Validator CommandValidator
    Processor CommandProcessor
}{
    "shellcode": {validateShellcodeCommand, processShellcodeCommand},
    "download":  {validateDownloadCommand, processDownloadCommand},
    "persist":   {validatePersistCommand, processPersistCommand},
}
```

## Part 7: Agent Execution Framework (Lessons 17-21)

We built the agent-side execution system:

- **ExecuteTask** - Command dispatcher using method expressions
- **Orchestrators** - Command-specific coordination logic
- **Doer Interface** - Cross-platform with build tags
- **Windows Shellcode** - Reflective DLL loading
- **Result Handling** - Server receives and displays results

```go
type OrchestratorFunc func(*HTTPSAgent, *server.HTTPSResponse) AgentTaskResult

func registerCommands(agent *HTTPSAgent) {
    agent.commandOrchestrators["shellcode"] = (*HTTPSAgent).orchestrateShellcode
    agent.commandOrchestrators["download"] = (*HTTPSAgent).orchestrateDownload
    agent.commandOrchestrators["persist"] = (*HTTPSAgent).orchestratePersist
}
```

## Part 8: Extending the Framework (Lessons 22-23)

We demonstrated framework extensibility:

- **Download Command** - File exfiltration in ~100 lines
- **Persistence Command** - Windows Registry and Startup folder

The architecture makes adding new commands predictable:

1. Create argument types (client and agent)
2. Add validator and processor on server
3. Add orchestrator on agent
4. Add doer on agent (with build tags if OS-specific)
5. Register the command

## Key Go Patterns Mastered

| Pattern | Where Used |
|---------|------------|
| Interfaces & Polymorphism | Server/Agent protocol abstraction |
| Factory Functions | NewServer(), NewAgent() |
| Method Expressions | Command routing in ExecuteTask |
| Build Tags | OS-specific doer implementations |
| Goroutines & Channels | Concurrent servers, signal handling |
| json.RawMessage | Flexible command arguments |
| Context & Cancellation | Graceful shutdown |

## Commands Implemented

| Command | Description |
|---------|-------------|
| shellcode | Execute shellcode via reflective DLL loading |
| download | Exfiltrate files from target system |
| persist | Install Windows persistence (Registry/Startup) |

## The Complete Flow

```
1. Operator -> curl POST /command
2. Server validates, processes, queues
3. Agent beacons, receives command
4. ExecuteTask routes to orchestrator
5. Orchestrator validates, calls doer
6. Doer executes OS-specific logic
7. Agent sends result to server
8. Server displays result
```

You've built a real C2 framework with:

- Multiple communication protocols
- Dynamic protocol switching
- Encrypted and authenticated communications
- Extensible command architecture
- Cross-platform support via build tags
- Windows persistence capabilities

This is a significant accomplishment!

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/course01/lesson-23">← Previous: Lesson 23</a></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/where-to-from-here">Next: Where to From Here →</a></div>
</div>