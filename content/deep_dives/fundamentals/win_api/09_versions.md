---
showTableOfContents: true
title: "API Evolution and Windows Versions"
type: "page"
---
## Overview

The Windows API has evolved significantly over the decades, expanding to support new features while maintaining compatibility with existing applications. Understanding this evolution is crucial for malware development - it directly impacts how our malware is designed.


## Major Windows Versions and Key Security-Relevant Features: A Quick Overview for Offensive Tooling

Understanding the security landscape of different Windows versions will inform us to the security measures we'll likely encounter.

### Windows XP
- Limited User Account Control (primarily for Administrator vs. Limited User), basic Windows Firewall, Data Execution Prevention (DEP) introduced in SP2, Address Space Layout Randomization (ASLR) was rudimentary or non-existent in early versions.
- Often considered a softer target due to fewer built-in defenses. Exploits might not need to contend with robust UAC or advanced memory protections. However, DEP and later ASLR (if patched) still present hurdles.

### Windows Vista
- Introduced User Account Control (UAC) with varying integrity levels (Low, Medium, High, System), mandatory integrity control (MIC), Kernel Patch Protection (PatchGuard) for 64-bit systems, improved ASLR, Code Integrity (CI), and a more robust Windows Firewall.
- UAC became a significant obstacle, requiring bypass techniques. Integrity levels dictate what processes can interact with each other, impacting injection and manipulation strategies. PatchGuard made kernel-level rootkits much harder to implement reliably on 64-bit systems.

### Windows 7
- Refined UAC (more granular settings), AppLocker (application whitelisting), improved BitLocker encryption, and further enhancements to DEP and ASLR. Expanded support for enterprise security features.
- AppLocker, if configured, poses a significant challenge to executing unauthorized tools. UAC bypasses continued to be a focus. Understanding how an organization uses features like BitLocker can be crucial for data exfiltration strategies.

### Windows 8/8.1
- Introduced Secure Boot (requires UEFI), Early Launch Anti-Malware (ELAM), Windows Defender improvements (more integrated), Fast Boot (impacting some traditional boot-time malware persistence), and expanded WinRT APIs (with their own security model). Control Flow Guard (CFG) was introduced.
- Secure Boot and ELAM aimed to prevent bootkits and rootkits. CFG adds another layer of exploit mitigation. Malware developers needed to adapt to these protections, for instance, by finding ways to bypass Secure Boot or exploit vulnerabilities in ELAM drivers.


### Windows 10/11
- Credential Guard, Device Guard (Virtualization-Based Security - VBS), Windows Defender Advanced Threat Protection (ATP) / Microsoft Defender for Endpoint, enhanced Control Flow Guard (CFG), Arbitrary Code Guard (ACG), eXtended Flow Guard (XFG), Attack Surface Reduction (ASR) rules, Windows Hello for biometric authentication, and ongoing improvements to all existing security features.
- These versions present the most formidable defense landscape. Credential Guard isolates LSA secrets, making pass-the-hash more difficult. Device Guard enforces strict code integrity. Exploiting these systems often requires highly sophisticated techniques, zero-day vulnerabilities, or misconfigurations. Understanding VBS and its implications for accessing kernel memory or sensitive data is critical.




## Understanding Major, Minor, and Build Versions

Windows versions are typically identified by a **major version**, a **minor version**, and a **build number** (e.g., Windows 10, version 22H2, Build 19045).

- **Major Version:** Indicates a significant release with substantial changes (e.g., Windows 7, Windows 8, Windows 10, Windows 11).
- **Minor Version (or Service Pack/Feature Update):** Represents updates within a major version. Historically, these were Service Packs (e.g., Windows XP SP3). For Windows 10/11, these are feature updates often denoted by YYHX (e.g., 22H2), indicating the year and half of release. These can introduce new APIs and security features.
- **Build Number:** A more granular identifier that increments with compilations of the OS, including security patches and minor fixes. Specific builds can have slightly different API availability or behaviour, particularly concerning undocumented functions.

## Why This Matters

Knowing the target Windows version and its specific API landscape is critical for several reasons in malware development:

1. **Exploitability:** Vulnerabilities are often version-specific. An exploit for an unpatched Windows 7 system might be ineffective or even crash a Windows 10 machine.
2. **Feature Abuse:** Malware can leverage legitimate API functions for malicious purposes (Living Off the Land Binaries - LOLBins). The availability and behaviour of these functions can vary. For example, PowerShell versions and their capabilities differ significantly.
3. **Evasion:**
    - **Security Software Hooking:** Antivirus and EDR solutions often hook API calls to monitor behaviour. The specific APIs they target and the methods they use can change with OS versions. Knowing this can help in designing evasion techniques (e.g., using direct syscalls or less common APIs).
    - **Signature Detection:** Version-specific artifacts or API usage patterns might be part of static or behavioural signatures.
    - **Bypassing Defenses:** Techniques to bypass UAC, AppLocker, Credential Guard, etc., are highly dependent on the OS version and patch level. An old UAC bypass won't work on a fully patched Windows 11.
4. **Stealth and Persistence:**
    - The APIs available for achieving persistence (e.g., registry keys, scheduled tasks, WMI event subscriptions) might differ or have different levels of scrutiny by security products across versions.
    - Methods for hiding artifacts or injecting code (e.g., `CreateRemoteThread`, `SetWindowsHookEx`, APC injection) may be more or less effective or detectable based on OS-level mitigations.
5. **Functionality:** Core malware functionality, like file system manipulation or process enumeration, relies on the Windows API. Using functions that don't exist on the target will lead to failure. For instance, trying to use a Windows 10-specific API for process hollowing on a Windows 7 machine will not work.
6. **Stability:** Using an API incorrectly or one that is not available can crash the malware or even the compromised host, drawing unwanted attention.

This evolution creates both opportunities and challenges. New API features might offer novel ways to achieve malicious objectives with potentially less scrutiny initially. However, using these requires careful consideration of compatibility. Relying on deprecated or undocumented functions is risky, as they can be removed or altered without notice.


## Determining Host OS Version After Landing

Once your initial payload has landed on a host, determining the precise Windows version is a critical first step for situational awareness and tailoring subsequent actions.

Several methods can be used.

### Using RtlGetVersion

The `GetVersionEx` function has been deprecated because it can be subject to compatibility shims and may return an incorrect version if the application isn't manifested for newer Windows versions. `RtlGetVersion` (from `ntdll.dll`) is the preferred way to get the actual OS version.

```c
#include <windows.h>
#include <stdio.h>

// Define the RTL_OSVERSIONINFOEXW structure if not already available (e.g., older SDKs)
// Or include <winternl.h> but that pulls in a lot.
// For simplicity, we'll assume ntstatus.h and necessary structures are available or defined.
// In a real offensive tool, you'd likely use direct syscalls or already have ntdll functions mapped.

typedef LONG NTSTATUS;
#define STATUS_SUCCESS ((NTSTATUS)0x00000000L)

typedef struct _RTL_OSVERSIONINFOEXW {
    ULONG dwOSVersionInfoSize;
    ULONG dwMajorVersion;
    ULONG dwMinorVersion;
    ULONG dwBuildNumber;
    ULONG dwPlatformId;
    WCHAR szCSDVersion[128];
    USHORT wServicePackMajor;
    USHORT wServicePackMinor;
    USHORT wSuiteMask;
    UCHAR wProductType;
    UCHAR wReserved;
} RTL_OSVERSIONINFOEXW, *PRTL_OSVERSIONINFOEXW;

// Function pointer type for RtlGetVersion
typedef NTSTATUS (WINAPI *pRtlGetVersion)(PRTL_OSVERSIONINFOEXW);

int main() {
    HMODULE hNtdll = GetModuleHandle(L"ntdll.dll");
    if (!hNtdll) {
        fprintf(stderr, "Failed to get handle to ntdll.dll\n");
        return 1;
    }

    pRtlGetVersion RtlGetVersionFunc = (pRtlGetVersion)GetProcAddress(hNtdll, "RtlGetVersion");
    if (!RtlGetVersionFunc) {
        fprintf(stderr, "Failed to get address of RtlGetVersion\n");
        // Fallback or error
        return 1;
    }

    RTL_OSVERSIONINFOEXW osInfo = {0};
    osInfo.dwOSVersionInfoSize = sizeof(osInfo);

    if (RtlGetVersionFunc(&osInfo) == STATUS_SUCCESS) {
        printf("Windows Version: %lu.%lu Build %lu\n",
               osInfo.dwMajorVersion,
               osInfo.dwMinorVersion,
               osInfo.dwBuildNumber);

        if (osInfo.dwMajorVersion == 10 && osInfo.dwMinorVersion == 0) {
            printf("This is Windows 10 or Windows 11 (or Server 2016/2019/2022).\n");
            // To distinguish Windows 11 from Windows 10 build 22000+,
            // you need to check if dwBuildNumber >= 22000 for Windows 11.
            if (osInfo.dwBuildNumber >= 22000) {
                printf("Likely Windows 11 (Build %lu)\n", osInfo.dwBuildNumber);
            } else {
                 printf("Likely Windows 10 (Build %lu)\n", osInfo.dwBuildNumber);
            }
        } else if (osInfo.dwMajorVersion == 6) {
            if (osInfo.dwMinorVersion == 1) printf("This is Windows 7 or Server 2008 R2.\n");
            else if (osInfo.dwMinorVersion == 2) printf("This is Windows 8 or Server 2012.\n");
            else if (osInfo.dwMinorVersion == 3) printf("This is Windows 8.1 or Server 2012 R2.\n");
            else if (osInfo.dwMinorVersion == 0) printf("This is Windows Vista or Server 2008.\n");
        }
        // Add more checks for other versions as needed
    } else {
        fprintf(stderr, "RtlGetVersion failed.\n");
        // Fallback, perhaps try GetVersionEx if desperate, or check registry.
    }
    return 0;
}
```



### Checking the Registry
The registry stores detailed version information. This is often used by scripts (PowerShell, VBScript) or when direct API calls are less convenient.

- **Key:** `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion`
- **Relevant Values:**
    - `ProductName` (e.g., "Windows 10 Pro")
    - `DisplayVersion` (e.g., "22H2" - for newer Windows 10/11)
    - `CurrentMajorVersionNumber` (e.g., 10)
    - `CurrentMinorVersionNumber` (e.g., 0)
    - `CurrentBuildNumber` (e.g., "19045")
    - `UBR` (Update Build Revision) - indicates the patch level.

**Example (Conceptual - actual registry reading code would be needed):**

```powershell
Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" | Select-Object ProductName, DisplayVersion, CurrentMajorVersionNumber, CurrentMinorVersionNumber, CurrentBuildNumber, UBR
```



## API Deprecation


We must also be aware of API deprecation, where functions are marked as obsolete and eventually removed or their behaviour altered. Microsoft typically provides replacement APIs and deprecation notices. For example, many of the original security functions have been deprecated in favour of more secure alternatives, often with "Ex" suffixes (like `CreateProcessAsUserW` being preferred over older, more limited ways to create processes in other user contexts), or entirely new API sets. Relying on deprecated functions can lead to tools breaking on newer OS versions or updates.

Understanding this API evolution, the security features tied to OS versions, and how to accurately determine the host's version is fundamental for crafting offensive tools that are effective, evasive, and stable across the diverse Windows ecosystem. By considering both backwards compatibility and forward-looking adoption of (or defense against) new features, we can create malware that maximizes operational success while minimizing detection.






---
[|TOC|]({{< ref "moc.md" >}})
[|PREV|]({{< ref "08_security.md" >}})
[|NEXT|]({{< ref "10_extended.md" >}})