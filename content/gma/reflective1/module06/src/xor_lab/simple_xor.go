package main

import (
	"bytes" // Used for comparing byte slices
	"fmt"
	"log"
)

// xorEncryptDecrypt performs XOR operation between data and key.
// Due to XOR's symmetry, this function works for both encryption and decryption.
// The key is repeated cyclically if it's shorter than the data.
func xorEncryptDecrypt(data []byte, key []byte) []byte {
	// Handle edge cases: empty data or empty key
	if len(data) == 0 {
		return []byte{} // Return empty if data is empty
	}
	if len(key) == 0 {
		log.Println("[!] Warning: XOR key is empty. Returning original data.")
		// Return a copy to avoid modifying the original slice buffer if caller expects it
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}

	// Allocate result buffer
	result := make([]byte, len(data))
	keyLen := len(key)

	// Loop through each byte of the data
	for i := 0; i < len(data); i++ {
		// Calculate the index for the key byte (repeating the key)
		keyIndex := i % keyLen
		// Perform XOR operation
		result[i] = data[i] ^ key[keyIndex]
	}

	return result
}

func main() {
	// --- Define Sample Data and Key ---
	// Sample plaintext data (e.g., a simple string)
	plaintext := []byte("Hi! I'm Mister Derp! DERP!")

	// Sample key (can be any sequence of bytes)
	key := []byte{0xDE, 0xAD, 0xBE, 0xEF} // A simple 4-byte key

	fmt.Printf("Original Plaintext : %s\n", string(plaintext))
	fmt.Printf("Original Plaintext (Hex): %X\n", plaintext)
	fmt.Printf("Key (Hex)          : %X\n", key)

	// --- Encrypt (Obfuscate) ---
	fmt.Println("\n[+] Encrypting using XOR...")
	ciphertext := xorEncryptDecrypt(plaintext, key)
	fmt.Printf("Ciphertext (Hex)   : %X\n", ciphertext)
	// Attempting to print ciphertext as string will likely produce garbage
	// fmt.Printf("Ciphertext (String): %s\n", string(ciphertext))

	// --- Decrypt (Deobfuscate) ---
	fmt.Println("\n[+] Decrypting using the SAME XOR function and key...")
	decryptedText := xorEncryptDecrypt(ciphertext, key)
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
