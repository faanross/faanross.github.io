---
layout: course01
title: "Lesson 11: HMAC Authentication"
---


## Solutions

- **Starting Code:** [lesson_11_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_11_begin)
- **Completed Code:** [lesson_11_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_11_end)

## Overview

Right now, any application that knows our server address can connect to it. There's no way for the server to verify that the agent checking in is actually *our* agent versus a security researcher probing the infrastructure, a curious sysadmin, or even a competing red team.

This is a fundamental problem in C2 architecture: **how do you establish trust between the server and agent?**

In this lesson, we'll implement HMAC-based authentication - a simple but effective way to verify that requests come from agents that possess a shared secret. This is the same principle used by AWS request signing, API authentication systems, and countless other security-critical applications.

## What is Authentication?

Before we write code, let's understand what we're trying to achieve.

**Authentication** answers the question: *"Are you who you claim to be?"*

There are several ways to prove identity:

1. **Something you know** - Passwords, PINs, shared secrets
2. **Something you have** - Certificates, hardware tokens
3. **Something you are** - Biometrics

For our C2, we'll use a **shared secret** - a key that only the server and legitimate agents know. If an agent can prove it knows this secret (without revealing it), we can trust it's legitimate.

## Authentication Options for C2

Before diving into implementation, let's survey the main approaches used to authenticate agents in C2 frameworks:

### 1. HMAC (Hash-based Message Authentication Code)

A shared secret is used to generate a signature for each request.

**Pros:**
- Simple to implement
- Low overhead - just adds headers to existing requests
- No additional infrastructure required
- Well-understood cryptographic primitive

**Cons:**
- Single shared secret across all agents (unless you implement per-agent keys)
- Secret must be embedded in the agent binary
- No forward secrecy - if secret is compromised, all past/future traffic is at risk

### 2. Mutual TLS (mTLS)

Both server and agent present certificates to authenticate each other.

**Pros:**
- Strong authentication with unique per-agent certificates
- Built into TLS - no additional protocol work
- Certificate revocation allows disabling compromised agents
- Provides encryption and authentication in one

**Cons:**
- Complex certificate management (CA, issuance, distribution)
- Certificates embedded in agent can be extracted
- More infrastructure to maintain
- Certificate expiration adds operational complexity

### 3. Pre-Shared Keys with Challenge-Response

Server sends a challenge, agent proves it knows the key by responding correctly.

**Pros:**
- Prevents replay attacks inherently
- Can implement mutual authentication

**Cons:**
- Requires additional round-trip
- More complex protocol design
- Still has shared secret distribution problem

### 4. Token-Based (JWT/OAuth)

Agent authenticates once, receives a token for subsequent requests.

**Pros:**
- Tokens can expire quickly, limiting damage from compromise
- Can embed claims/permissions in token
- Stateless verification on server

**Cons:**
- Initial authentication still needs another method
- Token theft allows impersonation until expiry
- Overkill for most C2 scenarios

### Why We're Using HMAC

For this course, we'll implement **HMAC-based authentication** for several reasons:

1. **Simplicity** - It's straightforward to understand and implement, making it ideal for learning
2. **No infrastructure** - No certificate authority or token service to set up
3. **Practical** - Many real-world systems (AWS request signing, API authentication) use this exact approach
4. **Foundation** - The concepts transfer directly to more complex schemes

**Realistic limitations to acknowledge:**
- A single shared secret means if one agent is compromised, all agents are potentially compromised
- The secret is embedded in the binary and could be extracted through reverse engineering
- For production red team operations, you'd likely want per-agent keys or mTLS

That said, HMAC provides meaningful protection against casual probing, request forgery, and replay attacks - a significant improvement over no authentication at all.

## What is HMAC?

**HMAC (Hash-based Message Authentication Code)** is a cryptographic construction that combines:

- A cryptographic hash function (like SHA-256)
- A secret key

The result is a **signature** that proves two things:

1. **The sender knows the secret key** - Only someone with the key could produce this signature
2. **The message hasn't been tampered with** - Any change to the message invalidates the signature

**How it works conceptually:**

```
HMAC = Hash(key + message)  // Simplified - actual construction is more complex
```

If you have the same key and the same message, you get the same HMAC. If either differs, you get a completely different result.

**Why HMAC instead of just hashing?**

Plain hashes are vulnerable to **length extension attacks**. HMAC's construction prevents this. Always use HMAC for authentication, never raw hashes.

## What We'll Create

- Shared secret configuration for both server and agent
- HMAC signature generation on the agent side
- HMAC signature verification on the server side
- Replay protection using timestamps

**Note:** This HMAC implementation is specifically for HTTPS communication. DNS has no HTTP headers to carry the signature and timestamp, so the DNS protocol remains unauthenticated. In a production framework, you'd implement a different authentication scheme for DNS (such as embedding signatures in the query data itself), but that's beyond our scope here.

## The Authentication Flow

```
1. Agent prepares request
   |-- Gets current timestamp
   |-- Computes: HMAC-SHA256(secret, timestamp + body)
   |-- Adds headers: X-Auth-Timestamp, X-Auth-Signature

2. Agent sends request
   |-- Headers contain timestamp and signature
   |-- Body contains normal data

3. Server receives request
   |-- Extracts timestamp from header
   |-- Checks timestamp is within tolerance (e.g., +/-5 minutes)
   |-- Recomputes HMAC with same inputs
   |-- Compares signatures

4. Server decision
   |-- Signatures match + timestamp valid -> Process request
   |-- Anything else -> Reject with 401 Unauthorized
```

## Part 1: Configure the Shared Secret

Both the server and agent need access to the same secret. We'll add it as a field to each config struct, then assign the value when we instantiate the config in each `main.go`.

### Add to Config Structs

In `internals/config/config.go`, add a `SharedSecret` field to both `AgentConfig` and `ServerConfig`:

```go
// AgentConfig holds all configuration values for the agent
type AgentConfig struct {
	ServerIP     string
	ServerPort   string
	Timing       TimingConfig
	Protocol     string // this will be the starting protocol
	SharedSecret string // HMAC authentication key
}

// ServerConfig holds all configuration values for the server
type ServerConfig struct {
	ListeningInterface string
	ListeningPort      string
	Protocol           string // this will be the starting protocol
	TlsKey             string
	TlsCert            string
	SharedSecret       string // HMAC authentication key
}
```

### Assign in Agent Main

In `cmd/agent/main.go`, add `SharedSecret` when instantiating the config:

```go
cfg := &config.AgentConfig{
	Protocol:     "https",
	ServerIP:     "127.0.0.1",
	ServerPort:   "8443",
	SharedSecret: "your-super-secret-key-change-in-production",
	Timing: config.TimingConfig{
		Delay:  5 * time.Second,
		Jitter: 50,
	},
}
```

### Assign in Server Main

In `cmd/server/main.go`, add the same `SharedSecret` value:

```go
cfg := &config.ServerConfig{
	Protocol:           "https",
	ListeningInterface: "127.0.0.1",
	ListeningPort:      "8443",
	TlsCert:            "./certs/server.crt",
	TlsKey:             "./certs/server.key",
	SharedSecret:       "your-super-secret-key-change-in-production",
}
```

**Important:** In production, use a cryptographically random key (at least 32 bytes), and never commit it to source control. Consider embedding it at compile time using build flags.

## Part 2: Agent-Side Signature Generation

Create a new file `internals/agent/auth.go`:

```go
package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"your-module/internals/config"
)

// SignRequest adds HMAC authentication headers to an HTTP request
func SignRequest(req *http.Request, body []byte, secret string) {
	// Get current timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Create the message to sign: timestamp + body
	message := timestamp + string(body)

	// Compute HMAC-SHA256
	signature := computeHMAC(message, secret)

	// Add headers
	req.Header.Set("X-Auth-Timestamp", timestamp)
	req.Header.Set("X-Auth-Signature", signature)
}

// computeHMAC calculates HMAC-SHA256 and returns hex-encoded result
func computeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
```

**Understanding the code:**

`SignRequest` is the public-facing function that orchestrates the signing process. It grabs the current Unix timestamp, concatenates it with the request body to form a single message, then passes that message along with the shared secret to `computeHMAC`. The resulting signature and timestamp are attached as custom HTTP headers (`X-Auth-Timestamp` and `X-Auth-Signature`), so the server can later extract them and verify the request. The timestamp is included in the signed message so that the signature is unique to this moment in time - even identical request bodies produce different signatures at different times, which is what gives us replay protection.

`computeHMAC` is where the actual cryptographic work happens. It creates an HMAC instance keyed with our shared secret and configured to use SHA-256 as the underlying hash function. The message bytes are fed into this HMAC via `Write`, and `Sum(nil)` finalizes the computation and returns the raw binary digest. That binary output is then hex-encoded into a string so it can be safely transmitted in an HTTP header. The key point is that only someone who possesses the same secret can produce the same signature for a given message - that's the entire basis of HMAC authentication.

### Update the Agent's Send Method

Modify the `Send()` method in `agent/agent_https.go` to sign requests:

```go
func (agent *HTTPSAgent) Send(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("https://%s/", agent.serverAddr)

	// For GET requests, body is empty
	var body []byte = nil

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Sign the request
	SignRequest(req, body, c.sharedSecret)

	resp, err := agent.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for authentication failure
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed - check shared secret")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, body)
	}

	return io.ReadAll(resp.Body)
}
```

## Part 3: Server-Side Signature Verification

Create a new file `internals/server/auth.go`:

```go
package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const TimestampTolerance = 300 // 5 minutes

// VerifyRequest checks HMAC signature and timestamp validity
func VerifyRequest(r *http.Request, secret string) error {
	// Extract headers
	timestamp := r.Header.Get("X-Auth-Timestamp")
	signature := r.Header.Get("X-Auth-Signature")

	if timestamp == "" || signature == "" {
		return fmt.Errorf("missing authentication headers")
	}

	// Verify timestamp is within tolerance
	if err := verifyTimestamp(timestamp); err != nil {
		return err
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	// Recompute the signature
	message := timestamp + string(body)
	expectedSignature := serverComputeHMAC(message, secret)

	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// verifyTimestamp checks if timestamp is within acceptable range
func verifyTimestamp(timestampStr string) error {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	now := time.Now().Unix()
	diff := now - timestamp

	// Check if timestamp is too old or too far in the future
	if diff < -TimestampTolerance || diff > TimestampTolerance {
		return fmt.Errorf("timestamp outside acceptable range")
	}

	return nil
}

// serverComputeHMAC calculates HMAC-SHA256 (same as agent)
func serverComputeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
```

**Critical security detail:**

```go
if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
```

We use `hmac.Equal` instead of `==` for comparison. This performs **constant-time comparison** to prevent timing attacks. With `==`, an attacker could measure response times to guess the signature byte-by-byte.

### Why Timestamp Tolerance?

The timestamp serves two purposes:

1. **Replay protection** - An attacker who captures a valid request can't replay it hours later
2. **Clock skew handling** - Systems may have slightly different clocks; 5 minutes tolerance handles this

## Part 4: Add Middleware to Server

Create authentication middleware in `internals/server/middleware.go`:

```go
package server

import (
	"log"
	"net/http"
)

// AuthMiddleware returns a middleware that wraps a handler with HMAC authentication
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := VerifyRequest(r, secret); err != nil {
				log.Printf("Authentication failed: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

### Apply Middleware to Routes

Update the `Start()` method in `internals/server/server_https.go`:

```go
func (s *HTTPSServer) Start() error {
	r := chi.NewRouter()

	// Apply authentication middleware to agent routes
	r.With(AuthMiddleware).Get("/", RootHandler)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: r,
	}

	return s.server.ListenAndServeTLS(s.tlsCert, s.tlsKey)
}
```

## Test

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Expected agent output (successful auth):**

```bash
2025/11/10 14:29:05 Starting Agent Run Loop
2025/11/10 14:29:05 Response from server: {"job": false}
```

**Test with wrong secret (modify agent's secret temporarily):**

```bash
2025/11/10 14:29:05 Error sending request: authentication failed - check shared secret
```

**Expected server output (failed auth):**

```bash
2025/11/10 14:29:05 Authentication failed: invalid signature
```

## Security Considerations

### What This Protects Against

- **Random probing** - Scanners won't have valid signatures
- **Request forgery** - Can't create valid requests without the secret
- **Replay attacks** - Old signatures expire due to timestamp checks
- **Tampering** - Modified requests have invalid signatures

### What This Doesn't Protect Against

- **Secret compromise** - If the secret leaks, all bets are off
- **Traffic analysis** - Attacker can still see that communication is happening
- **Endpoint discovery** - URLs are not hidden

### Production Improvements

1. **Unique secrets per agent** - Different key for each deployed agent
2. **Key rotation** - Periodically change secrets
3. **Nonce tracking** - Prevent replay within the timestamp window
4. **Rate limiting** - Slow down brute force attempts

## Conclusion

In this lesson, we implemented HMAC-based authentication:

- Created a shared secret configuration
- Implemented signature generation on the agent
- Implemented signature verification on the server
- Added timestamp-based replay protection
- Applied authentication as middleware

Our C2 now verifies that connecting agents possess the shared secret. In the next lesson, we'll add encryption to protect the actual content of our communications.

---

<div style="display: flex; justify-content: space-between; margin-top: 2rem;">
<div><a href="/courses/course01/lesson-10">← Previous: Lesson 10</a></div>
<div><a href="/courses/course01">↑ Table of Contents</a></div>
<div><a href="/courses/course01/lesson-12">Next: Lesson 12 →</a></div>
</div>