package main

import (
	"bytes"
	"fmt"
	"log"
)

// Here we XOR data with a key where the key byte is modified based on position.
// It is symmetric: calling it twice with the same key restores the original data.
func rollingXor(data []byte, key []byte) []byte {
	keyBytes := []byte(key) // Ensure key is treated as bytes
	keyLen := len(keyBytes)
	result := make([]byte, len(data))

	// Handle edge cases
	if len(data) == 0 {
		return []byte{}
	}
	if keyLen == 0 {
		log.Println("[!] Warning: Rolling XOR key is empty. Returning original data.")
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}

	for i := 0; i < len(data); i++ {
		// --- Rolling Key Calculation ---
		// 1. Get the base key byte for this position (repeating key)
		baseKeyByte := keyBytes[i%keyLen]
		// 2. Get a modifier based on the current position (lower 8 bits of index)
		positionByte := byte(i & 0xFF)
		// 3. Calculate the effective key byte for this position
		rollingKeyByte := baseKeyByte ^ positionByte
		// --- End Rolling Key Calculation ---

		// XOR the data byte with the *rolling* key byte
		result[i] = data[i] ^ rollingKeyByte
	}
	return result
}

func main() {
	fmt.Println("[+] Rolling XOR Obfuscation/Deobfuscation Demo")

	// --- Define Sample Data and Key ---
	plaintext := []byte("I am Lorde, YA YA YA!")

	key := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	fmt.Printf("Original Plaintext : %s\n", string(plaintext))
	fmt.Printf("Original Plaintext (Hex): %X\n", plaintext)
	fmt.Printf("Key (Hex)          : %X\n", key)

	// --- Encrypt (Obfuscate) using Rolling XOR ---
	fmt.Println("\n[+] Encrypting using ROLLING XOR...")
	ciphertext := rollingXor(plaintext, key)
	fmt.Printf("Ciphertext (Hex)   : %X\n", ciphertext)

	// --- Decrypt (Deobfuscate) using Rolling XOR ---
	fmt.Println("\n[+] Decrypting using the SAME ROLLING XOR function and key...")
	decryptedText := rollingXor(ciphertext, key)
	fmt.Printf("Decrypted Text     : %s\n", string(decryptedText))
	fmt.Printf("Decrypted Text (Hex): %X\n", decryptedText)

	// --- Verify Symmetry ---
	fmt.Println("\n[+] Verifying decryption matches original plaintext...")
	if bytes.Equal(plaintext, decryptedText) {
		fmt.Println("[+] Success! Decrypted text matches original plaintext.")
	} else {
		fmt.Println("[-] Failure! Decrypted text does NOT match original plaintext.")
	}
}
