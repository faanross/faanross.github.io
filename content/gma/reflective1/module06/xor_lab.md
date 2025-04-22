---
showTableOfContents: true
title: "XOR Functions in Go (Lab 6.1)"
type: "page"
---

## Goal
Let's write a Go function to perform simple XOR obfuscation/deobfuscation on byte slices using a repeating key. This will help us get a feel for the symmetric nature of the XOR operation.

## Code
- Feel free to replace both the `key` and `plaintext` at the top of `main()` with your own values.

```Go
package main

import (
	"bytes" 
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

```



## Code Breakdown
### `xorEncryptDecrypt` function
- Takes the input `data` and the `key` (both byte slices) as arguments.
- Handles edge cases where data or key might be empty. If the key is empty, it returns the original data as XORing with nothing changes nothing (though a warning is printed).
- Creates a `result` byte slice of the same length as the input `data`.
- It iterates through each byte of the `data` using index `i`.
- Inside the loop, `keyIndex := i % len(key)` calculates which byte of the key to use for the current data byte. The modulo operator (`%`) ensures the key wraps around and repeats if it's shorter than the data.
- `result[i] = data[i] ^ key[keyIndex]` performs the core bitwise XOR operation between the current data byte and the corresponding key byte, storing the result.
- Finally, it returns the `result` slice containing the XORed data.


### `main` function
- Sets up sample `plaintext` data and a sample `key`.
- Calls `xorEncryptDecrypt` once to produce the `ciphertext`.
- Prints the original data, key, and ciphertext (using `%X` format verb for hex representation, which is often clearer for arbitrary byte values).
- Calls `xorEncryptDecrypt` a _second time_, passing the `ciphertext` and the _exact same key_. This demonstrates the decryption process.
- Prints the `decryptedText`.
- Uses `bytes.Equal` to perform a byte-by-byte comparison between the original `plaintext` and the final `decryptedText` to formally verify that the decryption correctly reversed the encryption.



## Instructions
- You can either use `go build` to compile, then run your application, or for simplicity use `go run`.
- Navigate to the directory containing your go file (make sure it's the only one in this directory) and enter `go run .`.


## Results

```
❯ go run .
Original Plaintext : Hi! I'm Mister Derp! DERP!
Original Plaintext (Hex): 4869212049276D204D6973746572204465727021204445525021
Key (Hex)          : DEADBEEF

[+] Encrypting using XOR...
Ciphertext (Hex)   : 96C49FCF978AD3CF93C4CD9BBBDF9EABBBDFCECEFEE9FBBD8E8C

[+] Decrypting using the SAME XOR function and key...
Decrypted Text     : Hi! I'm Mister Derp! DERP!
Decrypted Text (Hex): 4869212049276D204D6973746572204465727021204445525021

[+] Verifying decryption matches original plaintext...
[+] Success! Decrypted text matches original plaintext.

```


## Discussion
We can clearly see that our `xorEncryptDecrypt` function successfully transformed the original plaintext into ciphertext, and also perfectly the original plaintext, confirming the symmetric property of XOR. This forms the basis for simple payload obfuscation.


## Conclusion
Let's now integrate this concept into our shellcode-containing DLL.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "simple.md" >}})
[|NEXT|]({{< ref "load_lab.md" >}})