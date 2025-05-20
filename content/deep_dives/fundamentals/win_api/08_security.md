---
showTableOfContents: true
title: "Security Implications"
type: "page"
---
## Windows Architecture and Security Models

The Windows architecture deeply impacts its security model. The very nature of the design establishes boundaries between different security contexts and controlling access to system resources, so a deep understanding of the architecture is crucial to gain insight into the defenses we'll come up against as malware developers.

As we explored right at the start of this guide, the Windows architecture is, at the most fundamental level, bifurcates into the user mode/kernel mode division. And so quite obviously then, the security model builds on this bifurcation. This boundary serves as a primary security perimeter, preventing user-mode applications from directly accessing hardware or manipulating critical system structures. When applications need privileged operations, they must request these services through controlled interfaces (the Windows API), allowing the system to validate the request and enforce security policies.

So this division is the main "security checkpoint", but by no means the only one, Windows implements several additional security mechanisms:
1. **Access Control Lists (ACLs)**: These structures define which users and groups can access specific resources and what operations they can perform. The Windows Object Manager applies these permissions to all named objects in the system, including files, registry keys, and synchronization objects. So in a sense it build on top of the user/kernel division to create more "access nuance", or fine-grained control.
2. **Security Identifiers (SIDs)**: Every security principal (user, group, or service) has a unique SID that identifies it within the system. These SIDs form the basis for access control decisions, appearing in both security tokens and ACLs.
3. **Security Tokens**: Each process receives a security token that encapsulates its security context, including the user's SID, group memberships, privileges, and integrity level. The system consults this token when the process attempts to access secured resources.
4. **Mandatory Integrity Control (MIC)**: Introduced in Windows Vista, this mechanism assigns integrity levels to processes and objects, preventing lower-integrity processes from modifying higher-integrity resources regardless of ACL permissions.
5. **Virtualization Based Security (VBS)**: Modern Windows versions can use hardware virtualization to create isolated security environments, protecting critical system components even if the main operating system is compromised.


Understanding these components will become critical once we dive deeper into EDRs, understanding their workings better so that we can effectively bypass them.


## Security Implications for System Programmers

In some sense, developing malware for Windows systems is a form of system programming. Here however, I want to explore the typical "best practices" employed by people developing legitimate Windows software, as it obviously provides a lot of valuable insights when attempting to exploit them + the environment in which they operate.

Several key principles guide secure Windows API programming:

1. **Principle of Least Privilege**: Applications should request and use only the minimum privileges necessary to accomplish their tasks. This applies to both the overall process token (which might include administrative privileges) and specific resource access requests. For example, opening a file for reading when write access isn't needed:

```c
HANDLE hFile = CreateFile(
    fileName,
    GENERIC_READ,  // Only request read access
    FILE_SHARE_READ,
    NULL,
    OPEN_EXISTING,
    FILE_ATTRIBUTE_NORMAL,
    NULL
);
```

2. **Input Validation**: All input, especially from external sources, should be validated before use. Windows API functions often have specific parameter requirements, and passing invalid values can lead to unexpected behaviour or security vulnerabilities. Proper validation (also referred to as sanitation) includes checking ranges, formats, and sizes of input data.
3. **Secure Resource Management**: Handles, memory, and other resources should be properly acquired, protected, and released to prevent leaks or unauthorized access. Security-sensitive handles should be protected with appropriate security descriptors, and memory containing sensitive information should be securely allocated and cleared when no longer needed.
4. **Error Handling**: Comprehensive error detection and handling improves both reliability and security. Unhandled errors can leave applications in inconsistent states, potentially exposing security vulnerabilities. Proper error handling includes checking return values from all security-relevant API calls and implementing appropriate recovery or failure mechanisms.
5. **Race Condition Awareness**: Windows API programming often involves asynchronous operations and shared resources, creating potential for race conditions. Secure code must account for these possibilities, using proper synchronization and avoiding time-of-check-to-time-of-use (TOCTOU) vulnerabilities.
6. **Defensive Coding**: Anticipate potential misuse and failure scenarios, implementing defensive measures to detect and prevent security issues. This includes bounds checking, null pointer validation, and defensive resource management, even when the API documentation doesn't explicitly require these checks.

In addition, when working on security-sensitive applications, systems programmers might leverage additional security mechanisms provided by Windows:

- **Advanced ACL Usage**: Creating and applying custom security descriptors to protect application resources.
- **Impersonation**: Temporarily adopting a different security context for specific operations.
- **Windows Security Features**: Leveraging Address Space Layout Randomization (ASLR), Data Execution Prevention (DEP), and Control Flow Guard (CFG).
- **Secure Boot and Code Signing**: Ensuring system integrity through verified boot processes and signed code.



## Security Implications for Defensive Security Practitioners

In much the same way as understanding the best practices of Windows systems programmers provides valuable insight to us as malware developers, so too does understanding the best practices of defensive security practitioners.

In defensive security contexts, particularly in Endpoint Detection and Response (EDR) development, practitioners leverage Windows API knowledge to:

1. **Monitor System Activity**: By hooking key API functions or leveraging the Windows Event Tracing framework, EDR solutions can observe application behaviour, file system activity, network connections, and process creation/termination.
2. **Detect Anomalous Behavior**: Understanding normal API usage patterns helps identify suspicious activities, such as unusual privilege escalation requests, attempts to disable security features, or exploitation of known vulnerabilities.
3. **Implement Preventive Measures**: Through API hooking, policy enforcement, or direct kernel-mode intervention, security tools can block potentially malicious activities before they cause harm.
4. **Perform Forensic Analysis**: Analyzing API call sequences and parameters helps reconstruct security incidents and understand attack methodologies.


## Security Implications for Offensive Security Practitioners

In offensive security contexts, including malware development but also for example exploit development, foundational knowledge of the Windows API can be leveraged to:

1. **Identify Potential Vulnerabilities**: Understanding API behaviour and limitations helps us to identify potential security weaknesses, such as parameter validation issues, race conditions, or privilege escalation opportunities.
2. **Develop Exploitation Techniques**: Knowledge of the Windows architecture guides the development of exploitation methods that bypass security controls or escalate privileges.
3. **Create Stealthy Tools**: Understanding how defensive tools monitor API usage helps in developing techniques that evade detection, such as direct system calls, API unhooking, or alternative execution methods. This one in particular is apt for malware/implant development.
4. **Reverse Engineering**: Familiarity with the Windows API aids in analyzing and understanding compiled code, identifying function purposes, and reconstructing program logic.




---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "07_error.md" >}})
[|NEXT|]({{< ref "09_versions.md" >}})