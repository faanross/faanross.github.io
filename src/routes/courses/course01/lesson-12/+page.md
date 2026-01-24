---
layout: course01
title: "Lesson 12: Payload Encryption"
---


## Solutions

- **Starting Code:** [lesson_12_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_12_begin)
- **Completed Code:** [lesson_12_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/main/lesson_12_end)

## Overview

In the previous lesson, we added authentication - the server now knows requests come from our agent. But the *content* of those requests is still visible to anyone monitoring the network.

Yes, we're using HTTPS, which provides transport encryption. But consider:

- TLS terminates at the server - someone with access to the server sees plaintext
- Corporate proxies may perform TLS inspection
- Logging systems might capture request bodies
- Defense in depth means not relying on a single layer

In this lesson, we'll add **application-layer encryption** using AES-GCM. This provides an additional layer of confidentiality and integrity beyond what TLS offers.

## Encoding vs Encryption

Before we proceed, let's clear up a common confusion.

**Encoding** (like Base64) is **not encryption**:
- Base64 is a *representation* - it transforms binary to text
- Anyone can decode it - there's no secret
- It provides zero confidentiality

**Encryption** requires a key:
- Data is transformed using a secret key
- Only someone with the key can decrypt
- It provides confidentiality

We already use Base64 to encode our shellcode for transmission. Now we'll encrypt the entire payload for confidentiality.

## What is AES-GCM?

**AES** (Advanced Encryption Standard) is a symmetric encryption algorithm - the same key encrypts and decrypts.

**GCM** (Galois/Counter Mode) is an operating mode that provides:

1. **Confidentiality** - Data is encrypted
2. **Integrity** - Tampering is detected
3. **Authentication** - Proves the sender knew the key

GCM is an **AEAD** cipher (Authenticated Encryption with Associated Data) - it's the modern standard for symmetric encryption.

**Key concepts:**

- **Key** - The shared secret (must be 16, 24, or 32 bytes for AES-128/192/256)
- **Nonce** - A unique value for each encryption (never reuse with same key!)
- **Ciphertext** - The encrypted data
- **Tag** - Authentication code appended to ciphertext

## What We'll Create

- Encryption configuration (derived from shared secret)
- Encryption function for outbound data
- Decryption function for inbound data
- Integration with agent and server communication

## The Encryption Flow

```
AGENT SENDING:
1. Prepare plaintext payload (JSON)
2. Generate random 12-byte nonce
3. Encrypt: ciphertext = AES-GCM(key, nonce, plaintext)
4. Prepend nonce: message = nonce + ciphertext
5. Base64 encode for HTTP transmission
6. Send in request body

SERVER RECEIVING:
1. Base64 decode the body
2. Extract nonce (first 12 bytes)
3. Extract ciphertext (remaining bytes)
4. Decrypt: plaintext = AES-GCM(key, nonce, ciphertext)
5. Parse JSON from plaintext
```

## Part 1: Derive Encryption Key

We'll derive our encryption key from the shared secret using a key derivation function. This is better than using the secret directly.

Create `internals/crypto/crypto.go`:

```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// DeriveKey derives a 32-byte AES-256 key from the shared secret
func DeriveKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// NonceSize is the size of the GCM nonce
const NonceSize = 12
```

**Why derive the key?**

- SHA-256 always produces exactly 32 bytes (perfect for AES-256)
- Makes key size consistent regardless of secret length
- Provides some protection if secret has low entropy

In production, you'd use a proper KDF like HKDF or Argon2.

## Part 2: Encryption Function

Add to `internals/crypto/crypto.go`:

```go
// Encrypt encrypts plaintext using AES-GCM and returns base64-encoded result
func Encrypt(plaintext []byte, secret string) (string, error) {
	key := DeriveKey(secret)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	// Encrypt and append tag
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	result := append(nonce, ciphertext...)

	// Base64 encode for transmission
	return base64.StdEncoding.EncodeToString(result), nil
}
```

**Understanding the code:**

1. **Derive key** from secret
2. **Create AES cipher** with the key
3. **Create GCM mode** wrapper around AES
4. **Generate random nonce** - CRITICAL: must be unique for each encryption
5. **Seal** encrypts and appends authentication tag
6. **Prepend nonce** so receiver can extract it
7. **Base64 encode** for safe HTTP transmission

## Part 3: Decryption Function

Add to `internals/crypto/crypto.go`:

```go
// Decrypt decrypts base64-encoded ciphertext using AES-GCM
func Decrypt(encoded string, secret string) ([]byte, error) {
	key := DeriveKey(secret)

	// Base64 decode
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	// Check minimum length (nonce + at least some ciphertext)
	if len(data) < NonceSize+1 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	// Decrypt and verify tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
```

**Understanding the code:**

1. **Base64 decode** the received data
2. **Extract nonce** (first 12 bytes)
3. **Extract ciphertext** (remaining bytes)
4. **Create cipher and GCM** (same as encryption)
5. **Open** decrypts and verifies the authentication tag

**If decryption fails:**

- The key was wrong
- The nonce was modified
- The ciphertext was tampered with
- The authentication tag was invalid

Any of these returns an error - you can't tell which failed (by design).

## Part 4: Update Agent Communication

Modify the agent's `Send()` method to encrypt outbound data:

```go
func (agent *HTTPSAgent) Send(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("https://%s/", agent.serverAddr)

	// Prepare check-in data (could include agent ID, status, etc.)
	checkInData := map[string]interface{}{
		"status": "active",
	}

	plaintext, _ := json.Marshal(checkInData)

	// Encrypt the payload
	encryptedBody, err := crypto.Encrypt(plaintext, config.SharedSecret)
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
	SignRequest(req, []byte(encryptedBody))

	resp, err := agent.client.Do(req)
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
	decrypted, err := crypto.Decrypt(string(encryptedResponse), config.SharedSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypting response: %w", err)
	}

	return decrypted, nil
}
```

## Part 5: Update Server Handler

Modify the server's `RootHandler` to decrypt incoming data and encrypt responses:

```go
func RootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	// Read encrypted body
	encryptedBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Decrypt the payload
	plaintext, err := crypto.Decrypt(string(encryptedBody), config.SharedSecret)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		http.Error(w, "Decryption failed", http.StatusBadRequest)
		return
	}

	log.Printf("Decrypted payload: %s", string(plaintext))

	// Process the decrypted data...
	// (existing logic here)

	// Prepare response
	response := HTTPSResponse{
		Job:    false,
		Change: false,
	}

	responsePlaintext, _ := json.Marshal(response)

	// Encrypt response
	encryptedResponse, err := crypto.Encrypt(responsePlaintext, config.SharedSecret)
	if err != nil {
		log.Printf("Response encryption failed: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte(encryptedResponse))
}
```

## The Nonce Problem

**CRITICAL:** Never reuse a nonce with the same key.

If you encrypt two messages with the same nonce and key:
- An attacker can XOR the ciphertexts
- The result is the XOR of the plaintexts
- This completely breaks confidentiality

Our implementation uses `crypto/rand` to generate random nonces. With 12 bytes (96 bits), the probability of collision is negligible for typical usage.

**For extremely high-volume systems:**

Consider using a counter-based nonce instead:
- Start at 0, increment for each message
- Never reset, persist across restarts
- Guaranteed unique if properly managed

## Test

**Start the server:**

```bash
go run ./cmd/server
```

**Start the agent:**

```bash
go run ./cmd/agent
```

**Expected server output:**

```bash
2025/11/10 14:29:05 Endpoint / has been hit by agent
2025/11/10 14:29:05 Decrypted payload: {"status":"active"}
```

**Test with wrong key (modify agent's secret temporarily):**

```bash
2025/11/10 14:29:05 Decryption failed: cipher: message authentication failed
```

The authentication tag verification fails - we can't decrypt with the wrong key.

## Security Analysis

### What We've Achieved

```
Layer 1: TLS (HTTPS)
|-- Encrypts transport
|-- Server authentication via certificate
|-- Protects against network eavesdropping

Layer 2: HMAC Authentication
|-- Verifies agent has shared secret
|-- Prevents request forgery
|-- Includes replay protection

Layer 3: AES-GCM Encryption
|-- Application-layer confidentiality
|-- Integrity verification
|-- Protection even if TLS is compromised
```

### Remaining Considerations

- **Key management** - Secrets need secure distribution
- **Perfect forward secrecy** - If key is compromised, past traffic is readable
- **Metadata** - Request timing and sizes are still visible

For production systems, consider using established protocols like Noise or TLS 1.3 mutual authentication.

## Conclusion

In this lesson, we implemented application-layer encryption:

- Created AES-GCM encryption and decryption functions
- Derived encryption keys from the shared secret
- Updated agent to encrypt outbound data and decrypt responses
- Updated server to decrypt incoming data and encrypt responses
- Understood the critical importance of nonce uniqueness

Our C2 now has three layers of protection: TLS, HMAC authentication, and payload encryption. In the next lesson, we'll implement the command endpoint.

---

[Previous: Lesson 11 - HMAC Authentication](/courses/course01/lesson-11) | [Next: Lesson 13 - Command Endpoint](/courses/course01/lesson-13) | [Course Home](/courses/course01)
