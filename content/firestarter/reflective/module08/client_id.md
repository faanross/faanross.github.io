---
showTableOfContents: true
title: "Environmental Keying + Client ID (Theory 8.3)"
type: "page"
---
## Client ID Generation

In the previous lesson we spoke about the use of a `clientID` as one component of the environmental information we'll use to derive our key. This identifier may also serve to distinguish different client machines connecting to the server, but the question is: how does the agent generate this ID?

## Goal

It's useful to first define the goal of identifier before designing it. So, our aim is to generate an identifier that is:
1. **Reproducible:** Running the agent multiple times on the _same_ machine should ideally produce the _same_ `clientID`. This is important for consistency, potentially for tracking or if the server uses the ID for longer-term associations.
2. **Somewhat Unique:** Different machines running the agent should ideally produce _different_ `clientID`s. This helps distinguish between different compromised hosts or legitimate clients.


This process is sometimes referred to as **environmental keying**, where characteristics of the host environment are used to generate an identifier.



## Our Implementation

We'll create a specific function, `getEnvironmentalID`, to achieve this. It will combine two pieces of information readily available on a Windows system:

1. **Volume Serial Number:** It calls the `windows.GetVolumeInformation` API function, specifically targeting the C: drive (`"C:\\"`). One of the pieces of information returned by this function is the volume serial number (`volumeSerial`). This is a unique number assigned to a disk volume when it's formatted and tends to be relatively stable for a given operating system installation on specific hardware.
2. **Hostname:** It calls `os.Hostname()` to retrieve the computer's network name. This is often configurable by the user or administrator but provides another piece of system-specific information.



## Combining the Information

Our `getEnvironmentalID` function will combine these two pieces of information to create the final `clientID` string. We'll take the first 5 characters of the hostname, and then append the volume serial number formatted as a hexadecimal string (e.g., "`ABCDE-a1b2c3d4`").

This approach generates an ID based on relatively stable system characteristics (volume serial number is less likely to change than hostname, but both are fairly constant for a given machine). It provides a reasonable balance between reproducibility and uniqueness for the purpose of deriving distinct session keys per client in the context of our project's design. Other environmental factors could also be incorporated (e.g., MAC address, CPU information, specific registry keys, username) to create more complex or robust fingerprints if needed - again I encourage you to explore alternatives.

This concludes the theory sections for Module 8, we're now ready to integrate client/server communication,  key derivation, and obfuscation to our existing reflective loading logic.




---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "cs_lab.md" >}})