package main

import (
	"fmt"
	"log"
	"os"
)

// --- Include the xorEncryptDecrypt function from Lab 6.1 ---
// xorEncryptDecrypt performs XOR operation between data and key.
// The key is repeated cyclically if it's shorter than the data.
func xorEncryptDecrypt(data []byte, key []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	if len(key) == 0 {
		log.Println("[!] Warning: XOR key is empty. Returning original data.")
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}
	result := make([]byte, len(data))
	keyLen := len(key)
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%keyLen]
	}
	return result
}

func main() {
	fmt.Println("[+] DLL Obfuscator Tool")

	// --- Configuration ---
	inputDllPath := "calc_dll.dll"   // Original DLL
	outputFilePath := "calc_dll.xor" // Obfuscated output file
	// IMPORTANT: Use the same key in the loader!
	xorKey := []byte{0x55, 0xAA, 0x55, 0xAA, 0x11, 0x22, 0x33, 0x44} // Example 8-byte key

	fmt.Printf("[+] Input DLL: %s\n", inputDllPath)
	fmt.Printf("[+] Output File: %s\n", outputFilePath)
	fmt.Printf("[+] XOR Key (Hex): %X\n", xorKey)

	// --- Read Input DLL ---
	fmt.Printf("[+] Reading input DLL...\n")
	plaintextBytes, err := os.ReadFile(inputDllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read input file '%s': %v\n", inputDllPath, err)
	}
	fmt.Printf("[+] Read %d bytes from input DLL.\n", len(plaintextBytes))

	// --- Obfuscate ---
	fmt.Println("[+] Obfuscating DLL bytes using XOR...")
	obfuscatedBytes := xorEncryptDecrypt(plaintextBytes, xorKey)
	fmt.Println("[+] Obfuscation complete.")

	// --- Write Output File ---
	fmt.Printf("[+] Writing %d obfuscated bytes to '%s'...\n", len(obfuscatedBytes), outputFilePath)
	err = os.WriteFile(outputFilePath, obfuscatedBytes, 0644) // Standard file permissions
	if err != nil {
		log.Fatalf("[-] Failed to write output file '%s': %v\n", outputFilePath, err)
	}
	fmt.Println("[+] Output file written successfully.")
	fmt.Println("[+] Obfuscator finished.")
}
