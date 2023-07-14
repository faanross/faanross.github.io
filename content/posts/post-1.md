---
title: "Threat Hunting Dll-injected C2 beacons (Practical Course)"
date: 2023-07-12T02:01:58+05:30
description: "In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 approaches: memory forensics, log analysis + UEBA, and traffic analysis."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan ross"
---

*** 

# Introduction

In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 approaches: memory forensics, log analysis + UEBA, and traffic analysis. The entire course is practically-oriented, meaning that we'll learn by doing. I'll sprinkle in a tiny bit of theory just so we are on the same page re: C2 frameworks and DLL-injection attacks; and in case you wanted to dig in deeper I provide extensive references throughout this document. 

In case you're new and a little trepidated...
I'm literally going to hold your hand from point A to Z so even if you are a beginner and most of this seems foreign **fret not**! WRITE SOMETHING and add a GIF




Here's a brief overview of what we'll be getting upto...
- In PART 1 we're going to set up the virtualized environment,
- we'll create a windows 10 VM which will server as our victim,
- we'll also set up a kali linux box which will be our attacker, 
- as well as an ubuntu box which we'll use to run some post-mortem analysis on.

.
- In PART 2 we'll run the actual attack ourselves,
- for the classical dll-injection we'll use metasploit to generate both the stager and meterpreter handler,
- once we've transferred the stager to the victim we'll run it from memory using powersploit,
- for the reflective dll-injection we'll perform the entire process using metasploit.

.
- In PART 3 we'll cover memory forensics,
- first we'll do a basic live read using Process Hacker,
- we'll then dump the memory with winpmem,
- finally we'll have a look at the it with Volatility.

.
- IN PART 4 we'll get into some logs,
- along with standard Windows Event Logs, we'll also use other (cough, far superior, cough), logs we setup in the first part: namely sysmon and powershell logging,
- we'll briefly jump into the raw logs just to look at some very high-level indicators and then,
- we'll process them using the awesome UEBA framework DeepBlueCLIv3.

.
- IN PART 5 we'll look at traffic analysis,
- we'll run our PCAPS through Zeek,
- and get some insights from the threat hunting framework RITA.

In the end we'll recap and formulate some key takeaways to serve you on your journey as you venture forth into the world and become a bada$$ hunter.

But first, *le sigh*, it's required we just dip our toes into a wee bit of theory. But I promise once we're done here - 10 to 15 mins tops - it'll be applied learning until the end of our journey. 

Sounds good? Let's get it.


{{< figure src="/img/randy01.gif" title="" class="custom-figure" >}}



***

# Theory
# what is a DLL?
Succinctly as possible, a DLL is a communal library containing code. They are not a program or an executable in and of themselves, but they are in essence a collection of functions and data that can be used by other programs. 

So think of a DLL as a communal resource: let's say you have 4 programs running and they all want to use a common function - let's say for the sake of simplicity the ability to minimize the gui window. Now instead of each of those programs having their own personal copy of the function that allows that, they'll instead access a DLL that contains the function to minimize gui windows instead. So when you click on the minimize icon and that program needs the code to know how to behave, it does not get instructions from its own program code, rather it pulls it from the appropriate DLL with some help from the Windows API. 

Thus any program you run will constantly call on different DLLs to get access to a wide-variety of common (and often critical) functions and data.

# what is a classical DLL-injection?
So keeping what I just mentioned in mind - that any running program is accessing a variety of code from various DLLs at any time - what then is a DLL-injection attack? Well in a normal environment we have legit programs accessing code from legit DLLs. 

With a DLL-injection attack we enter into the population of legit DLLs a malicious one, that is a DLL that contains the code the attacker wants executed. The attacker then injects it into the memory space of a legitimate process. Using a Windows API function (commonly LoadLibrary or CreateRemoteThread), the attacker manipulates the legitimate process into loading and executing the malicious DLL. This effectively allows the malicious code within the DLL to run, often with the same permissions as the hijacked process.

Threat actors love DLL-injection attacks because since they are executed within the context of a legitimate process they run with the same privileges as that of the process (ie potentially elevated), but even more so it makes them much harder to detect. No longer can we look on the process-level for malware, instead we have to peer beneath them at a arguably convoluted level of abstraction. 

Even though classical DLL-injection attacks are less noisy for this exact reason, they still have a design flaw that makes our lives as threat hunters easier - they leave their fingerprints all over the disc. When the malicious DLL is initially transferred to the victim's system, it's written to disc, allowing us a potential breadcrumb for discovery. 

And thus the inevitable next iteration in this branch of digital evolution is...

# what is a reflective DLL-injection?
At a *high-level*  classical and reflective DLLs are identical save for one difference: whereas the former is written to disc then injected into memory space, the latter is injected into memory space directly. This makes them conventionally even harder to catch since we can't rely on any disc forensics to reveal its presence. However, as we'll learn in this course, in another way it makes it for those who know what to look for perhaps a bit easier. 

How come?

Well, on a pattern-level we can observe that the very fact that a DLL, meaning ANY DLL, is in memory without a disc counterpart is very unusual. Perhaps not immediate incident alert level unusual, but at the very least more than unusual enough to warrant further prodding with piqued interest. 

As a bridge to the closing part of our theory section let's zoom out a bit. Here we have been speaking about a specific mechanism of how malware (that is bad code) gets a victim's system to execute it. There are obviously many other such mechanisms, and equally bviously there are many different types of malware that use specifically DLL-injection attacks as the means to their desired ends (ie getting executed). 

In this specific course however we'll be focussing on a very specific type of malware, actually it would be even more accurate to say we'll focus on a specific component of a specific type of malware... 

# what is a Command and Control (C2) framework, stager, and beacon?

Let's start by sketching a scenario of how many typical attacks play out these days.

{{< figure src="/img/hackers01.gif" title="" class="custom-figure" >}}

An attacker sends a spear-phishing email to an employee at a company. The employee, perhaps tired and not paying full attention, opens the "uregent invoice" attached to the email. Opening this attachment executes a tiny program called a stager.

A stager, though not inherently malicious, "sets the stage" by performing a specific task: it reaches out to a designated address (owned by the hacker) to download another piece of code, then executes it.

The downloaded code establishes the attacker's presence on the victim's system. It acts as a "gateway," allowing the attacker to execute commands on the victim's system from their own.

So the system that the attacker uses to execute these commands is called the Command and Control (C2) server.

The code downloaded by the stager is a type of C2 implant known as a beacon, an approach popularized by Cobalt Strike. Unlike traditional C2 implants that maintain a continuous, persistent network connection (which can raise suspicion), a beacon does not. 

Instead, it periodically "calls home" to the C2 server, asking whether there are any new commands. If there are no commands, the connection is immediately terminated. If there are commands, the beacon retrieves them and then terminates the connection, lying dormant until the next scheduled "check-in". This sporadic communication helps the beacon blend into normal network traffic, making it more difficult to detect.

GREAT, and that's it for the theory, it's time to get going! But in case you are feeling inspired here are a selection of incredible resources that helped me.

{{< youtube borfuQGrB8g >}}

.
{{< youtube lz2ARbZ_5tE >}}

.
{{< youtube ihElrBBJQo8 >}}

*** 

# PART 1: Setting up our virtualized environment
# Overview

For this course I'll be using [VMWare Workstation](https://store-us.vmware.com/workstation_buy_dual) which as of writing costs around $200. However you could also do it with either [VMWare Player](https://www.vmware.com/ca/products/workstation-player.html), or [Oracle Virtualbox](https://www.virtualbox.org/wiki/Downloads), both of which are free. 

Note that some of the details of the setup might be slightly different if you choose to use either of the lastmentioned options and if that occurs then it'll be upto you to figure that out. Google, ChatGPT, StackExchange etc.

One final thing before we get setting up, you'll need the following three iso's (all free of course):
* for the victim we'll use [Windows 10 Enterprise Evaluation](https://info.microsoft.com/ww-landing-windows-10-enterprise.html)
* for the attacker we'll use [Kali Linux](https://www.kali.org/get-kali/#kali-installer-images)
* for post-mortem analysis we'll be using [Ubuntu Linux 20.04 Focal Fossa](https://releases.ubuntu.com/focal/). Just note here the actual edition 20.04 is important since we'll run RITA on it, which, as of writing, runs best on Focal Fossa.

Ok so at this point if you have your hosted hypervisor and all three iso's we are ready to proceed.

# VM 1: Windows 10 aka "The Victim" 

{{< figure src="/img/screamdrew.gif" title="" class="custom-figure" >}}
 
First we'll install the OS using the iso, following that we'll make a bunch of configurations including: 
- deep disable MS Defender + Windows updates
- install sysmon
- enable powershell logging
- install Process Hacker
- install winpmem
- install wireshark
- turn our VM into a template so we can clone copies in the future

# Installation

1. In VMWare Workstation goto `File` -> New Virtual Machine. 
2. Choose `Typical (recommended)`, then click `Next`. 
3. Then select `I will install the operating system later` and hit `Next`.

{{< figure src="/img/image001.png" title="" class="custom-figure" >}}

4. Select `Microsoft Windows`, and under Version select `Windows 10`. 
5. Here you are free to call the machine whatever you'd like, in my case I am calling it `Victim`. 
6. Select 60 GB and `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.

{{< figure src="/img/image002.png" title="" class="custom-figure" >}}

8. Under `Memory` (see left hand column) I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Windows 10 iso file.

{{< figure src="/img/image003.png" title="" class="custom-figure" >}}

11. Now finally select `Network Adapter`. Note we'll change this later to `Host-Only` (to minimize noise), but for now we'll need an internet connection to finish the installation so you can select either `NAT` or `Bridged`. Click `Close` then `Finish`.

You should now see your VM in your Library (left hand column), select it and then click on Power on this virtual machine.

{{< figure src="/img/image004.png" title="" class="custom-figure" >}}

Wait a short while and then you should see a Windows Setup window. Choose your desired language etc, select Next and then click on Install Now. Select ‘I accept the license terms’ and click Next. Next select ‘Custom: Install Windows only (advanced)’, and then select your virtual HD and click Next.

{{< figure src="/img/image005.png" title="" class="custom-figure" >}}

Once its done installing we’ll get to the setup, select your region, preferred keyboard layout etc. Accept the License Agreement (if you dare!). Now once you reach the Sign in page don’t fill anything in, rather select ‘Domain join instead’ in the bottom left corner.

{{< figure src="/img/image006.png" title="" class="custom-figure" >}}

Choose any username and password, in my case it'll be the highly original choice of `User` and `password`, feel free to choose something else. Then choose 3 securty questions, since this is a "burner" system used for the express purpose of this course don't overthink it. Turn off all the privacy settings (below), and for Cortana select `Not Now`.

{{< figure src="/img/image007.png" title="" class="custom-figure" >}}

Windows will now finalize installation + configuration, this could take a few minutes, whereafter you will see your Desktop.

# VMWare Tools
Next we'll install VMWare Tools which will ensure our VMs screen resolution assumes that of our actual monitor, but more importantly it also gives us the ability to copy and paste between the host and the VM. So this is optional, if you're oldskool and prefer writing all commands out by hand then feel free to skip this. 

So just to be sure, at this point you should be starting at a Windows desktop. Now in the VMWare windoes click `VM` and then `Install VMWare Tools`. If you open explorer (in the VM) you should now see a D drive. 

{{< figure src="/img/image008.png" title="" class="custom-figure" >}}

Double-click the drive, hit `Yes` when asked if we want this app to make changes to the device. Hit `Next`, select `Typical` and hit `Next`. Finally hit `Install` and then once done `Finish`. You'll need to restart your system for the changes to take effect, but we'll shut it down since we need to change a setting. So hit the Windows icon, Power icon, and then select `Shut Down`.

Right-click on your VM and select `Settings`. In the list on the LHS select `Display`, which should be right at the bottom. On the bottom - deselect `Automatically adjust user interface size in the virtual machine`, as well as `Strech mode`, it should now look like this:

{{< figure src="/img/image009.png" title="" class="custom-figure" >}}

Go ahead and start-up the VM once again, we'll now get to configuring our VM.

# Configuration
# Deep disable MS Defender + Windows updates

I call this 'deep disable' because simply toggling off the switches in `Settings` won't actually fully disable Defender and Updates. Windows looks at you like a little brother - it feels the need to protect you a bit, most of the time without you even knowing. (Unlike Linux of course which will allow you to basically nuke your entire OS if you so desired.) 

And just so you know why it is we're doing this:
- We are disabling Defender so that the AV won't interfere with us attacking the system. Now you might think well this represents an unrealistic situation since in real-life we'll always have our AV running. Thing is, this is a simulation - we are simulating an actual attack. Yes the AV might pick up on our mischevious escapades here since we are using very well-known and widely-used malware (Metasploit mainly). But, if you are being attacked by and actual threat actor worth their salt they likely won't be using something so familiar as default Metasploit modules - they will likely be capcable of using analogous technologies that your AV will not pick up on.
- As for updates, we disable this because sometimes we can spend all this time configuring and setting things up and then one day we boot our VM up, Windows does it's automatic update schpiel, and suddenly things are broken. This is thus simply to support the stability of our long-term use of this VM. 

1. **Disable Tamper Protection**
    1. Hit the `Start` icon, then select the `Settings` icon.
    2. Selet **`Update & Security `**.
    3. In LHS column, select `Windows Security`, then click `Open Windows Security`.
    4. A new window will pop up. Click on `Virus & threat protection`.
    5. Scroll down to the heading that says `Virus & threat protection settings` and click on `Manage settings`.
    6. There should be 4 toggles in total, we are really interested in disabling `Real-time protection`, however since we are here just go ahead and disable all of them. 
    7. Note that Windows will warn you and ask if you want to allow this app to make changes to the device, hit `Yes`.
    8. All 4 toggle settings should now be disabled.

{{< figure src="/img/image010.png" title="" class="custom-figure" >}}
    
2. **Disable the Windows Update service**
    1. Open the Run dialog box by pressing Win+R.
    2. Type **`services.msc`** and press Enter.
    3. In the Services list, find **`Windows Update`**, and double-click it.
    4. In the Windows Update Properties (Local Computer) window, under the **`General`** tab, in the **`Startup type:`** dropdown menu, select **`Disabled`** - see image below.
    5. Click **`Apply`** and then **`OK`**.
    
 {{< figure src="/img/image011.png" title="" class="custom-figure" >}}

3. **Disable Defender via Group Policy Editor**
    1. Open the Run dialog box by pressing Win+R.
    2. Type `gpedit.msc` and hit enter. The `Local Group Policy Editor` should have popped up.
    3. In the tree on the LHS navigate to the following: `Computer Configuration` > `Administrative Templates` > `Windows Components` > `Microsoft Defender Antivirus`.
    4. In the RHS double-click on `Turn off Microsoft Defender Antivirus`.
    5. In the new window on the top left select `Enabled` - see image below. 
    6. First hit `Apply` then `OK`.

 {{< figure src="/img/image012.png" title="" class="custom-figure" >}}

4. **Disable Updates via Group Policy Editor**
    1. Still in `Local Group Policy Editor`, navigate to: `Computer Configuration` > `Administrative Templates` > `Windows Components` > `Windows Update`.
    2. In the RHS double-click on `Configure Automatic Updates`.
    3. Select `Disabled`, then click `Apply` and `OK`.

{{< figure src="/img/image013.png" title="" class="custom-figure" >}}

5. **Disable Defender via Registry**
    1. In the search bar on the bottom type `cmd`.
    2. On the top left, right under `Best match` you should see `Command Prompt`.
    3. Right-click and select `Run as administrator`, hit `Yes`.
    4. Copy and paste the following command below into your command prompt and hit enter.
    ```
    REG ADD "hklm\software\policies\microsoft\windows defender" /v DisableAntiSpyware /t REG_DWORD /d 1 /f
    ```
    
Almost there! We just need to boot into Safe Mode to make some final adjustments to the registry and then we are good to go.

6. **Reboot system in Safe Mode**
    1. Open the Run dialog box by pressing Win+R.
    2. Write `msconfig` and hit enter.
    3. Select the `Boot` tab.
    4. Under `Boot options` select `Safe boot`, ensure `Minimal` is selected - see image below. 
    5. Hit `Apply` first, the `OK`.
    6. Select `Restart`.
    
{{< figure src="/img/image014.png" title="" class="custom-figure" >}}

7. **Disable Defender via Registry**
    1. Open the Run dialog box by pressing Win+R.
    2. Write `regedit` and hit enter, this should bring up the `Registry Editor`.
    3. Below you will see a list of 6 keys. For each of these keys you will follow the same process: once the key is selected find the `Start` value in the RHS, double-click, change the value to `4` and hit `OK` - see image below.
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `Sense`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdBoot`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WinDefend`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdNisDrv`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdNisSvc`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdFilter`

{{< figure src="/img/image015.png" title="" class="custom-figure" >}}

8. **Disable Updates via Registry**
    1. Still in `Registry Editor` let's navigate to the following:
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SOFTWARE` > `Microsoft` > `Windows` > `CurrentVersion` > `WindowsUpdate` > `Auto Update`
    2. Right-click the `Auto Update` key, select `New`, and then click `DWORD (32-bit) Value`.
    3. Name the new key **`AUOptions`** and press Enter.
    4. Double-click the new **`AUOptions`** key and change its value to **`2`**. Click **`OK`** - see image below.
    5. Close Registry Editor.

{{< figure src="/img/image016.png" title="" class="custom-figure" >}}

9. **Leave Safe Mode**
    1. All that's left to do is get back into our regular Windows environment.
    2. Open the Run dialog box by pressing Win+R.
    3. Write `msconfig` and hit enter.
    4. Select `Boot` tab.
    5. Deselect `Safe boot`, hit `Apply`, hit `OK`.
    6. Hit `Restart`.

And that, I can promise you, is by far the most boring part of this entire course. But I did it on purpose - this is very important if you are going to start simulating attacks and threat hunting on your own system. And the cool thing is now that we've done it we'll also learn how to create templates + clones, meaning you would conceivably have to do it again, but simply clone the VM we've created. But before that, let's setup all the awesome tools we'll be using in this course. 

LFG!

IMAGE HERE

# SYSMON 
 
Ok so now you should be back in the normal Windows environment looking at your Desktop. We'll now setup Sysmon, for right now all you need to know is that Sysmon is a simple, free, Microsoft-owned program that will DRAMATICALLY improve our logging ability. 

Thing is the standard Windows Event Logs (hence forth referred to simply as WEL) were not designed by somebody with security in mind. In fact, ask most security professionals what they think of WEL and you'll probably get either a sarcastic chuckle or a couple of expletives.

But Sysmon, created by the legend Mark Russinovich, takes about 5 minutes to setup and will DRAMATICALLY improve logging, specifically as it relates to security events. In case you wanted to learn more about Sysmon's ins and outs [see this talk](https://www.youtube.com/watch?v=6W6pXp6EojY). And if you really wanted to get in deep, which at some point I recommend you do, see [this playlist](https://www.youtube.com/playlist?list=PLk-dPXV5k8SG26OTeiiF3EIEoK4ignai7) from TrustedSec. Finally here is another great talk by one of my favourite SANS instructors (Eric Conrad) on [using Sysmon for  Threat Hunting](https://www.youtube.com/watch?v=7dEfKn70HCI).

Before we get installing Sysmon there's just one thing you need to know - in addition to download the actual Sysmon install we also need a config file. Of late  I have heard a few trusted sources, included Eric Conrad (mentioned above) prefer [this version from Neo23x0](https://github.com/bakedmuffinman/Neo23x0-sysmon-config) whose authors included another blue team giant, Florian Roth. 

So first download the config file (which is in xml format), the [go here to download Sysmon](https://download.sysinternals.com/files/Sysmon.zip). You should now have two zip files - the config you download on Github, as well as the Sysmon zip file. Extract the Sysmon file, the contents should look as follows:

{{< figure src="/img/image017.png" title="" class="custom-figure" >}}

Now also extract the zip file containing the config. Inside of the folder rename `sysmonconfig-export.xml` to `sysmonconfig.xml`. Now simply cut (or copy) the file and paste it in the folder containing sysmon. 

Great everything is setup so now we can install it with a simple command. Open command prompt as administrator and navigate to the folder containing sysmon and the config file - in my case it is `c:\Users\User\Downloads\Sysmon`. Run the following command:

```
Sysmon.exe -accepteula -i
```

This is what a successful installation will look like:

{{< figure src="/img/image018.png" title="" class="custom-figure" >}}

Now let's just validate that it's running. First type `powershell` so we change over into a PS shell, and rrun the command `Get-Service sysmon`. In the image below we can see it is running - we are good to go!

{{< figure src="/img/image019.png" title="" class="custom-figure" >}}

That's it for Sysmon, not let's enable PowerShell logging. 

# PowerShell Logging

Unlike Sysmon we don't have to install anything here, Windows comes pre-configured with PS logging, but it's turned off by default. So we just need to flip this switch which again, as is the case with Sysmon, will dramatically improve logging as it relates to security.

Why?

Hackers often exploit PowerShell due to its powerful capabilities and direct access to the Windows API and other lower-level operations. This tactic, known as "living off the land," involves using the system's own tools against it. By enabling PowerShell logging, you record all command-line activities, which can be invaluable when conducting threat hunting. By analyzing the logs, you can identify the exact commands executed by an attacker, and therefore gain a clearer understanding of what they did and how they did it. 

Enable PowerShell logging:
1. Hit **`Win + R`** keys together to open the Run dialog box.
2. Type **`gpedit.msc`** and press Enter. This will open the **`Local Group Policy Editor`**.
3. In the left-hand panel, navigate to **`Computer Configuration`** > **`Administrative Templates`** > **`Windows Components`** > **`Windows PowerShell`**.
4. On the right-hand side, you will see a policy setting named **`Turn on PowerShell Script Block Logging`**, double-click.
5. In the properties window, select the **`Enabled`** - see image below.
6. Hit **`Apply`** and then **`OK`**.

{{< figure src="/img/image020.png" title="" class="custom-figure" >}}

# Install Software

And now finally we'll install three programs:
- We'll use **Process Hacker** for live memory forensics 
- We'll use **winpmem** to create a memory dump for post-mortem memory forensics 
- We'll use **Wireshark** to generate a pcap for egress analysis

.
You can download [Process Hacker here](https://processhacker.sourceforge.io/downloads.php). Once downloaded go ahead and install.

You can download the latest release of [winpmem here](https://github.com/Velocidex/WinPmem/releases). Since its a portable executable there is no installation required, just download the `.exe` file and place it on the desktop. 

And finally the Wireshark setup file can be [downloaded from here](https://2.na.dl.wireshark.org/win32/Wireshark-win32-3.6.15.exe). Once downloaded run Setup, just keep all options per default, nothing fancy required. 

That's it friend. We are done with BY FAR the heaviest lifting in terms of VM setup - the next two will be a breeze. But before we get to that there's one very simple thing we can do that will make our lives much easier in the future - turning this VM into a template for cloning.

# Creating a Template 

So why do we want to do this. Well by turning this VM we just created into a template we are in essence creating an archetype (blueprint). Then, whenever we want this same "victim" system for any project or course we can simply clone it. Thus instead of repeating this entire, rather cumbersome process we can click a few buttons and have it ready to go in under a minute. This is also useful if we ever "mess up" the VM, we can just come back to this starting point where the machine is fresh, but all our configurations and software are as required. 

1. First shut down the VM.
2. In VMWare you should see the library pane on the LHS listing our VM. If you don't, hit `F9`, or go to `View` > `Customize` > `Library`.
3. Right-click on our VM (`Victim`), select `Snapshot` > `Take Snapshot`.
4. Name it anything you'd like, I will be calling it `Genesis`. Hit `Take Snapshot`.
5. Again right-click the VM and select `Settings`. 
6. On the top left we can see two tabs - `Hardware` and `Options`, select `Options`.
7. Go down to the bottom and select `Advanced`.
8. Select `Enable Template mode (to be used for cloning)`, hit `OK`.

{{< figure src="/img/image021.png" title="" class="custom-figure" >}}

9. Note you might want to rename this VM to something like `Victim Template`, so we are aware this is the template that we should not be using, but rather use for cloning. You can do this under `Settings` > `Options` > `General`.

{{< figure src="/img/image022.png" title="" class="custom-figure" >}}

10. Now let's create our first clone which we will actually be using in the course. Right-click on `Victim Template`, select `Manage` > `Clone`. Hit `Next`.
11. We'll select the snapshot we created and hit `Next`. 
12. Keep selection as `Create a linked clone` and hit `Next`. 
13. Give your clone a name, I will be calling it `Victim01`. Choose a location and hit `Next`.

That's it! You should now see both `Victim Template` and `Victim01` in your library.

The bad news - we still have two VMs to install. The good news - they will require minimal-to-no configuration, so at this point we're about 80% done with our VM setup. So let's get it done.

# Kali Linux Installation

We'll be using Kali Linux for attack, that is it'll effectively serve as our C2 server. The great thing about Kali Linux is that everything we'll need is already installed, so we just have to install the actual operating system. 

1. In VMWare hit `File` > `New Virtual Machine...`
2. `Typical (recommended)` and hit `Next`. 
3. `I will install the operating system later` and hit `Next`.
4. Select `Linux`, and under Version select `Debian 11.x 64-bit`. (Note: Kali Linux is built on top of Debian Linux).

{{< figure src="/img/image023.png" title="" class="custom-figure" >}}

5. Again call the machine whatever you'd like, in my case I am calling it `Hacker`. 
6. Increase the Maximum disk size to 60 GB and select `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.
8. Under `Memory` I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Kali Linux iso image.
11. And again for `Network Adapter` we'll keep it as either `NAT` or `Bridged` for now. Click `Close` then `Finish`.

So now let's get to actually installing it:
1. Right-click on the VM and select `Power` > `Start Up Guest`.
2. Select `Graphical Install`.
3. Select language, country etc.
4. Choose any `Hostname`, leave `Domain name` blank, for Full name and username I chose `hacker`.
5. Create a password, again though OBVIOUSLY not a suggested real-world practice, in these simulations I tend to simply use `password` since it minimizes any administrative friction. 
6. Choose a timezone.
7. Next select `Guided - use entire disk` and hit `Continue`.
8. The only disk should be selected, hit `Continue`.
9. Keep `All files in one partition (recommended for new users)`, hit `Continue`.
10. Keep `Finish partinioning and write changes to disk`, hit `Continue`.
11. Select `Yes` and `Continue`.
12. In `Software selection` keep the default selection and hit `Continue`. Kali will now start installing, just be aware this can take a few minutes, probably around 5 to 10. 

{{< figure src="/img/image024.png" title="" class="custom-figure" >}}

13. Next it'll ask you about installing a GRUB boot loader, keep it selected as `Yes` and hit `Continue`. 
14. Select `/dev/sda` and hit `Continue`. More installing... 

{{< figure src="/img/image025.png" title="" class="custom-figure" >}}

15. Finally it will inform us it's complete, we can hit `Continue` causing the system to reboot into Kali Linux. Enter your username and password and hit `Log In`.
16. Let's shut down the VM, then right-click on it in the library and select `Settings`. Under `Display` deselect `Stretch mode` and hit `OK`.

{{< figure src="/img/image026.png" title="" class="custom-figure" >}}

And that's it for our attacker machine - feel free to repeat the Template-Cloning process we performed for our Windows 10 VM if you so desire.

# Ubuntu Linux 20.04 Installation

And now finally we'll set up our Ubuntu VM, afterwards we'll install RITA, Zeek, and Volatility. 

1. In VMWare hit `File` > `New Virtual Machine...`
2. `Typical (recommended)` and hit `Next`. 
3. `I will install the operating system later` and hit `Next`.
4. Select `Linux`, and under Version select `Ubuntu 64-bit`.
5. Again call the machine whatever you'd like, in my case I am calling it `Analyst`. 
6. Increase the Maximum disk size to 60 GB and select `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.
8. Under `Memory` I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
- NOTE: Keep in mind that you will never run more than 2 VMs at the same time (Victim + Hacker), this VM will always only run by itself after the simulated attack. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Ubuntu Linux 20.04 iso image.
11. And again for `Network Adapter` we'll keep it as either `NAT` or `Bridged` for now. Click `Close` then `Finish`.

So now let's get to actually installing it:
1. Right-click on the VM and select `Power` > `Start Up Guest`.



Remmeber once all three are installed change Network adapter to host only. 