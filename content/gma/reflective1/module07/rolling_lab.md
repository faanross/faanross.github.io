---
showTableOfContents: true
title: "Implementing Rolling XOR (Lab 7.1)"
type: "page"
---
## Goal
In this lab we'll replace the simple XOR function from Lab 6.1 with a rolling XOR implementation. This will once again be a standalone application to just illustrate its functionality, we'll integrate it into our main application in the next lab.


## Code
Once again feel free to replace `plaintext` and `key` with your own values.

```go
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

```

## Code Breakdown
### `rollingXor` function
- `baseKeyByte := keyBytes[i % keyLen]` gets the standard repeating key byte.
- `positionByte := byte(i & 0xFF)` gets a byte value based on the index `i`.
- `rollingKeyByte := baseKeyByte ^ positionByte` combines these two to create the effective key for this position.
- `result[i] = data[i] ^ rollingKeyByte` performs the XOR using this position-dependent key byte.

### `main` function
 - This is structured identically to the `main` function in Lab 6.1.
 - It defines plaintext and a base key.
 - It calls `rollingXor` for encryption.
 - It calls the _same_ `rollingXor` function for decryption, demonstrating symmetry.
 - It verifies the result using `bytes.Equal`.

## Instructions
- You can either use `go build` to compile, then run your application, or for simplicity use `go run`.
- Navigate to the directory containing your go file (make sure it’s the only one in this directory) and enter `go run .`


## Results

```shell
❯ go run .
[+] Rolling XOR Obfuscation/Deobfuscation Demo
Original Plaintext : I am Lorde, YA YA YA!
Original Plaintext (Hex): 4920616D204C6F7264652C20594120594120594121
Key (Hex)          : DEADBEEF

[+] Encrypting using ROLLING XOR...
Ciphertext (Hex)   : 978CDD81FAE4D79AB2C198C48BE190B98F9CF5BDEB

[+] Decrypting using the SAME ROLLING XOR function and key...
Decrypted Text     : I am Lorde, YA YA YA!
Decrypted Text (Hex): 4920616D204C6F7264652C20594120594120594121

[+] Verifying decryption matches original plaintext...
[+] Success! Decrypted text matches original plaintext.

```


## Discussion
Everything is pretty self-explanatory in our output, bit we can confirm of course that are pre-encrypted plaintext, and decrypted plaintext, match.


## Conclusion
So that was not terribly exciting, but at least we confirmed that rolling XOR offered the same self-inverting ability as simple XOR did. Now let's go and integrate this and key derivation logic into our reflective loader.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "key.md" >}})
[|NEXT|]({{< ref "key_lab.md" >}})